package elevenlabs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketTTSService handles real-time text-to-speech via WebSocket.
//
// # Stream Completion Behavior
//
// ElevenLabs WebSocket TTS does not send an explicit "end of stream" signal.
// After calling Flush(), the server generates remaining audio and then waits
// for more input. If no input arrives within the inactivity timeout (default
// 20 seconds), the server sends an "input_timeout_exceeded" error and closes
// the connection.
//
// For applications that need faster stream completion detection, set a shorter
// InactivityTimeout in WebSocketTTSOptions (e.g., 5 seconds) and treat the
// timeout as a successful completion if audio was received after flush.
type WebSocketTTSService struct {
	client *Client
}

// WebSocketTTSOptions configures the WebSocket TTS connection.
type WebSocketTTSOptions struct {
	// ModelID is the model to use. Defaults to "eleven_turbo_v2_5" for low latency.
	ModelID string

	// OutputFormat specifies the audio output format.
	// Recommended for real-time: "pcm_16000", "pcm_22050", "pcm_24000", "pcm_44100"
	// Also supports: "mp3_44100_64", "mp3_44100_96", "mp3_44100_128", "mp3_44100_192"
	OutputFormat string

	// VoiceSettings configures the voice parameters.
	VoiceSettings *VoiceSettings

	// OptimizeStreamingLatency reduces latency at the cost of quality (0-4).
	// 0 = no optimization, 4 = maximum optimization.
	OptimizeStreamingLatency int

	// EnableSSMLParsing enables SSML parsing for the input text.
	EnableSSMLParsing bool

	// LanguageCode is the ISO language code (e.g., "en", "es").
	LanguageCode string

	// ChunkLengthSchedule controls text chunking for audio generation.
	// Array of integers representing character counts before generating audio.
	ChunkLengthSchedule []int

	// InactivityTimeout is the context timeout in seconds (default 20).
	InactivityTimeout int

	// PronunciationDictionaryIDs is a list of pronunciation dictionary IDs to use.
	PronunciationDictionaryIDs []string
}

// DefaultWebSocketTTSOptions returns default options optimized for low latency.
func DefaultWebSocketTTSOptions() *WebSocketTTSOptions {
	return &WebSocketTTSOptions{
		ModelID:                  "eleven_turbo_v2_5",
		OutputFormat:             "pcm_16000",
		OptimizeStreamingLatency: 3,
	}
}

// WebSocketTTSConnection represents an active WebSocket TTS connection.
type WebSocketTTSConnection struct {
	conn    *websocket.Conn
	voiceID string
	options *WebSocketTTSOptions
	mu      sync.Mutex
	closed  bool
	flushed bool // tracks if Flush() has been called

	// Channels for async operation
	audioOut  chan []byte
	alignOut  chan *TTSAlignment
	errChan   chan error
	doneChan  chan struct{} // signals when all audio is received after flush
	closeChan chan struct{}
	closeOnce sync.Once
	doneOnce  sync.Once
}

// TTSAlignment contains word-level timing information.
type TTSAlignment struct {
	Characters     []string  `json:"characters"`
	CharacterStart []float64 `json:"character_start_times_seconds"`
	CharacterEnd   []float64 `json:"character_end_times_seconds"`
}

// ttsWSMessage is the WebSocket message format for TTS.
type ttsWSMessage struct {
	Text                       string           `json:"text,omitempty"`
	VoiceSettings              *wsVoiceSettings `json:"voice_settings,omitempty"`
	GenerationConfig           *wsGenConfig     `json:"generation_config,omitempty"`
	XIAPIKey                   string           `json:"xi_api_key,omitempty"`
	TryTriggerGeneration       bool             `json:"try_trigger_generation,omitempty"`
	Flush                      bool             `json:"flush,omitempty"`
	CloseConnection            bool             `json:"close_connection,omitempty"`
	ContextID                  string           `json:"context_id,omitempty"`
	PronunciationDictionaryIDs []string         `json:"pronunciation_dictionary_locators,omitempty"`
}

type wsVoiceSettings struct {
	Stability       float64 `json:"stability"`
	SimilarityBoost float64 `json:"similarity_boost"`
	Style           float64 `json:"style,omitempty"`
	UseSpeakerBoost bool    `json:"use_speaker_boost,omitempty"`
}

type wsGenConfig struct {
	ChunkLengthSchedule []int `json:"chunk_length_schedule,omitempty"`
}

// ttsWSResponse is the WebSocket response from TTS.
type ttsWSResponse struct {
	Audio               string        `json:"audio,omitempty"`
	IsFinal             bool          `json:"isFinal,omitempty"`
	NormalizedAlignment *TTSAlignment `json:"normalizedAlignment,omitempty"`
	Alignment           *TTSAlignment `json:"alignment,omitempty"`
	Error               string        `json:"error,omitempty"`
	Message             string        `json:"message,omitempty"`
	Code                int           `json:"code,omitempty"`
}

// Connect establishes a WebSocket connection for real-time TTS.
func (s *WebSocketTTSService) Connect(ctx context.Context, voiceID string, opts *WebSocketTTSOptions) (*WebSocketTTSConnection, error) {
	if voiceID == "" {
		return nil, ErrEmptyVoiceID
	}

	if opts == nil {
		opts = DefaultWebSocketTTSOptions()
	}

	// Build WebSocket URL
	wsURL, err := s.buildWebSocketURL(voiceID, opts)
	if err != nil {
		return nil, err
	}

	// Create dialer with context
	dialer := websocket.Dialer{
		HandshakeTimeout: 0, // Use context timeout
	}

	// Add headers
	headers := http.Header{}
	headers.Set("xi-api-key", s.client.apiKey)

	// Connect
	conn, _, err := dialer.DialContext(ctx, wsURL, headers)
	if err != nil {
		return nil, fmt.Errorf("websocket dial failed: %w", err)
	}

	wsc := &WebSocketTTSConnection{
		conn:      conn,
		voiceID:   voiceID,
		options:   opts,
		audioOut:  make(chan []byte, 100),
		alignOut:  make(chan *TTSAlignment, 100),
		errChan:   make(chan error, 1),
		doneChan:  make(chan struct{}),
		closeChan: make(chan struct{}),
	}

	// Send initial configuration
	if err := wsc.sendInit(); err != nil {
		conn.Close()
		return nil, err
	}

	// Start reading responses
	go wsc.readLoop()

	return wsc, nil
}

func (s *WebSocketTTSService) buildWebSocketURL(voiceID string, opts *WebSocketTTSOptions) (string, error) {
	baseURL := s.client.baseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	// Convert HTTP URL to WebSocket URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	u.Path = fmt.Sprintf("/v1/text-to-speech/%s/stream-input", voiceID)

	// Add query parameters
	q := u.Query()
	if opts.ModelID != "" {
		q.Set("model_id", opts.ModelID)
	}
	if opts.OutputFormat != "" {
		q.Set("output_format", opts.OutputFormat)
	}
	if opts.OptimizeStreamingLatency > 0 {
		q.Set("optimize_streaming_latency", fmt.Sprintf("%d", opts.OptimizeStreamingLatency))
	}
	if opts.EnableSSMLParsing {
		q.Set("enable_ssml_parsing", "true")
	}
	if opts.LanguageCode != "" {
		q.Set("language_code", opts.LanguageCode)
	}
	if opts.InactivityTimeout > 0 {
		q.Set("inactivity_timeout", fmt.Sprintf("%d", opts.InactivityTimeout))
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (wsc *WebSocketTTSConnection) sendInit() error {
	msg := ttsWSMessage{
		Text: " ", // Initial empty text to establish connection
	}

	if wsc.options.VoiceSettings != nil {
		msg.VoiceSettings = &wsVoiceSettings{
			Stability:       wsc.options.VoiceSettings.Stability,
			SimilarityBoost: wsc.options.VoiceSettings.SimilarityBoost,
			Style:           wsc.options.VoiceSettings.Style,
			UseSpeakerBoost: wsc.options.VoiceSettings.UseSpeakerBoost,
		}
	}

	if len(wsc.options.ChunkLengthSchedule) > 0 {
		msg.GenerationConfig = &wsGenConfig{
			ChunkLengthSchedule: wsc.options.ChunkLengthSchedule,
		}
	}

	if len(wsc.options.PronunciationDictionaryIDs) > 0 {
		msg.PronunciationDictionaryIDs = wsc.options.PronunciationDictionaryIDs
	}

	return wsc.sendJSON(msg)
}

func (wsc *WebSocketTTSConnection) sendJSON(msg any) error {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()

	if wsc.closed {
		return fmt.Errorf("connection closed")
	}

	return wsc.conn.WriteJSON(msg)
}

func (wsc *WebSocketTTSConnection) readLoop() {
	defer wsc.closeChannels()

	for {
		select {
		case <-wsc.closeChan:
			return
		default:
		}

		_, message, err := wsc.conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				select {
				case wsc.errChan <- err:
				default:
				}
			}
			return
		}

		var resp ttsWSResponse
		if err := json.Unmarshal(message, &resp); err != nil {
			select {
			case wsc.errChan <- fmt.Errorf("failed to parse response: %w", err):
			default:
			}
			continue
		}

		// Check for errors
		if resp.Error != "" || resp.Message != "" {
			errMsg := resp.Error
			if errMsg == "" {
				errMsg = resp.Message
			}
			select {
			case wsc.errChan <- fmt.Errorf("server error: %s", errMsg):
			default:
			}
			continue
		}

		// Decode and send audio
		if resp.Audio != "" {
			audioBytes, err := base64.StdEncoding.DecodeString(resp.Audio)
			if err != nil {
				select {
				case wsc.errChan <- fmt.Errorf("failed to decode audio: %w", err):
				default:
				}
				continue
			}
			if len(audioBytes) > 0 {
				select {
				case wsc.audioOut <- audioBytes:
				case <-wsc.closeChan:
					return
				}
			}
		}

		// Send alignment if available
		if resp.NormalizedAlignment != nil {
			select {
			case wsc.alignOut <- resp.NormalizedAlignment:
			default:
			}
		} else if resp.Alignment != nil {
			select {
			case wsc.alignOut <- resp.Alignment:
			default:
			}
		}

		// Check if this is the final response after flush
		if resp.IsFinal {
			wsc.mu.Lock()
			flushed := wsc.flushed
			wsc.mu.Unlock()

			if flushed {
				// Signal that all audio has been received
				wsc.doneOnce.Do(func() {
					close(wsc.doneChan)
				})
			}
		}
	}
}

func (wsc *WebSocketTTSConnection) closeChannels() {
	wsc.closeOnce.Do(func() {
		close(wsc.closeChan)
		close(wsc.audioOut)
		close(wsc.alignOut)
	})
}

// SendText sends text to be converted to speech.
// The text can be sent in chunks as it becomes available (e.g., from an LLM stream).
func (wsc *WebSocketTTSConnection) SendText(text string) error {
	if text == "" {
		return nil
	}

	msg := ttsWSMessage{
		Text: text,
	}

	return wsc.sendJSON(msg)
}

// SendTextWithContext sends text with a specific context ID for multi-context sessions.
func (wsc *WebSocketTTSConnection) SendTextWithContext(text, contextID string) error {
	if text == "" {
		return nil
	}

	msg := ttsWSMessage{
		Text:      text,
		ContextID: contextID,
	}

	return wsc.sendJSON(msg)
}

// TriggerGeneration forces audio generation for buffered text.
func (wsc *WebSocketTTSConnection) TriggerGeneration() error {
	msg := ttsWSMessage{
		Text:                 " ",
		TryTriggerGeneration: true,
	}
	return wsc.sendJSON(msg)
}

// Flush signals that no more text will be sent and flushes remaining audio.
// This should be called when the text stream is complete.
// After calling Flush, use Done() to wait for all audio to be received.
func (wsc *WebSocketTTSConnection) Flush() error {
	wsc.mu.Lock()
	wsc.flushed = true
	wsc.mu.Unlock()

	msg := ttsWSMessage{
		Text:  "",
		Flush: true,
	}
	return wsc.sendJSON(msg)
}

// Audio returns a channel that receives audio chunks as they are generated.
func (wsc *WebSocketTTSConnection) Audio() <-chan []byte {
	return wsc.audioOut
}

// Alignments returns a channel that receives word alignment information.
func (wsc *WebSocketTTSConnection) Alignments() <-chan *TTSAlignment {
	return wsc.alignOut
}

// Errors returns a channel that receives errors from the connection.
func (wsc *WebSocketTTSConnection) Errors() <-chan error {
	return wsc.errChan
}

// Done returns a channel that is closed when all audio has been received after Flush().
// Use this to wait for completion before closing the connection.
func (wsc *WebSocketTTSConnection) Done() <-chan struct{} {
	return wsc.doneChan
}

// Close closes the WebSocket connection gracefully.
func (wsc *WebSocketTTSConnection) Close() error {
	wsc.mu.Lock()
	if wsc.closed {
		wsc.mu.Unlock()
		return nil
	}
	wsc.closed = true
	wsc.mu.Unlock()

	// Send close message
	msg := ttsWSMessage{
		CloseConnection: true,
	}
	_ = wsc.sendJSON(msg)

	// Close the connection
	wsc.closeChannels()
	return wsc.conn.Close()
}

// StreamText is a convenience method that sends all text from a channel and returns audio.
// It handles flushing automatically when the input channel closes.
func (wsc *WebSocketTTSConnection) StreamText(ctx context.Context, textStream <-chan string) (<-chan []byte, <-chan error) {
	audioOut := make(chan []byte, 100)
	errOut := make(chan error, 1)

	go func() {
		defer close(audioOut)
		defer close(errOut)

		// Forward audio from connection
		done := make(chan struct{})
		go func() {
			defer close(done)
			for audio := range wsc.Audio() {
				select {
				case audioOut <- audio:
				case <-ctx.Done():
					return
				}
			}
		}()

		// Send text as it arrives
		for {
			select {
			case text, ok := <-textStream:
				if !ok {
					// Input stream closed, flush and wait for remaining audio
					if err := wsc.Flush(); err != nil {
						errOut <- err
						return
					}
					<-done
					return
				}
				if err := wsc.SendText(text); err != nil {
					errOut <- err
					return
				}
			case err := <-wsc.Errors():
				errOut <- err
				return
			case <-ctx.Done():
				errOut <- ctx.Err()
				return
			}
		}
	}()

	return audioOut, errOut
}
