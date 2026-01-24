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

// WebSocketSTTService handles real-time speech-to-text via WebSocket.
type WebSocketSTTService struct {
	client *Client
}

// WebSocketSTTOptions configures the WebSocket STT connection.
type WebSocketSTTOptions struct {
	// ModelID is the transcription model to use.
	// Default: "scribe_v2_realtime"
	ModelID string

	// AudioFormat specifies the audio encoding format.
	// Options: "pcm_8000", "pcm_16000", "pcm_22050", "pcm_24000", "pcm_44100", "pcm_48000", "ulaw_8000"
	// Default: "pcm_16000"
	AudioFormat string

	// LanguageCode is the expected language (e.g., "en", "es").
	// If not specified, language will be auto-detected.
	LanguageCode string

	// IncludeTimestamps enables word-level timing information.
	IncludeTimestamps bool

	// IncludeLanguageDetection includes detected language in responses.
	IncludeLanguageDetection bool

	// CommitStrategy determines how transcripts are committed.
	// Options: "manual" (default), "vad" (voice activity detection)
	CommitStrategy string

	// VAD settings (only used when CommitStrategy is "vad")

	// VADSilenceThresholdSecs is the silence duration to trigger commit in VAD mode.
	// Default: 1.5
	VADSilenceThresholdSecs float64

	// VADThreshold is the VAD sensitivity threshold.
	// Default: 0.4
	VADThreshold float64

	// MinSpeechDurationMs is the minimum speech duration in milliseconds.
	// Default: 100
	MinSpeechDurationMs int

	// MinSilenceDurationMs is the minimum silence duration in milliseconds.
	// Default: 100
	MinSilenceDurationMs int
}

// DefaultWebSocketSTTOptions returns default options for real-time STT.
func DefaultWebSocketSTTOptions() *WebSocketSTTOptions {
	return &WebSocketSTTOptions{
		ModelID:           "scribe_v2_realtime",
		AudioFormat:       "pcm_16000",
		CommitStrategy:    "manual",
		IncludeTimestamps: true,
	}
}

// WebSocketSTTConnection represents an active WebSocket STT connection.
type WebSocketSTTConnection struct {
	conn      *websocket.Conn
	options   *WebSocketSTTOptions
	sessionID string
	mu        sync.Mutex
	closed    bool

	// Channels for async operation
	transcriptOut chan *STTTranscript
	errChan       chan error
	closeChan     chan struct{}
	closeOnce     sync.Once
}

// STTTranscript represents a transcription result.
type STTTranscript struct {
	// Text is the transcribed text.
	Text string `json:"text"`

	// IsFinal indicates if this is a final (committed) result.
	IsFinal bool `json:"is_final"`

	// Words contains word-level timing if enabled.
	Words []STTWord `json:"words,omitempty"`

	// LanguageCode is the detected language.
	LanguageCode string `json:"language_code,omitempty"`
}

// STTWord represents a single word with timing.
type STTWord struct {
	Text      string  `json:"text"`
	Start     float64 `json:"start"`
	End       float64 `json:"end"`
	Type      string  `json:"type,omitempty"`       // "word" or "spacing"
	SpeakerID string  `json:"speaker_id,omitempty"` // Speaker identification
}

// sttInputAudioChunk is the audio data message for v2 API.
type sttInputAudioChunk struct {
	MessageType  string `json:"message_type"`
	AudioBase64  string `json:"audio_base_64"`
	Commit       bool   `json:"commit,omitempty"`
	SampleRate   int    `json:"sample_rate,omitempty"`
	PreviousText string `json:"previous_text,omitempty"`
}

// sttResponse is a generic response for parsing message_type.
type sttResponse struct {
	MessageType  string    `json:"message_type"`
	Text         string    `json:"text,omitempty"`
	LanguageCode string    `json:"language_code,omitempty"`
	Words        []STTWord `json:"words,omitempty"`
	Error        string    `json:"error,omitempty"`
	SessionID    string    `json:"session_id,omitempty"`
}

// Connect establishes a WebSocket connection for real-time STT.
func (s *WebSocketSTTService) Connect(ctx context.Context, opts *WebSocketSTTOptions) (*WebSocketSTTConnection, error) {
	if opts == nil {
		opts = DefaultWebSocketSTTOptions()
	}

	// Build WebSocket URL with all configuration as query parameters
	wsURL, err := s.buildWebSocketURL(opts)
	if err != nil {
		return nil, err
	}

	// Create dialer with context
	dialer := websocket.Dialer{
		HandshakeTimeout: 0, // Use context timeout
	}

	// Add headers for authentication
	headers := http.Header{}
	headers.Set("xi-api-key", s.client.apiKey)

	// Connect
	conn, _, err := dialer.DialContext(ctx, wsURL, headers)
	if err != nil {
		return nil, fmt.Errorf("websocket dial failed: %w", err)
	}

	wsc := &WebSocketSTTConnection{
		conn:          conn,
		options:       opts,
		transcriptOut: make(chan *STTTranscript, 100),
		errChan:       make(chan error, 1),
		closeChan:     make(chan struct{}),
	}

	// Start reading responses
	go wsc.readLoop()

	return wsc, nil
}

func (s *WebSocketSTTService) buildWebSocketURL(opts *WebSocketSTTOptions) (string, error) {
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

	u.Path = "/v1/speech-to-text/realtime"

	// Add query parameters (v2 API uses query params for configuration)
	q := u.Query()

	if opts.ModelID != "" {
		q.Set("model_id", opts.ModelID)
	}
	if opts.AudioFormat != "" {
		q.Set("audio_format", opts.AudioFormat)
	}
	if opts.LanguageCode != "" {
		q.Set("language_code", opts.LanguageCode)
	}
	if opts.IncludeTimestamps {
		q.Set("include_timestamps", "true")
	}
	if opts.IncludeLanguageDetection {
		q.Set("include_language_detection", "true")
	}
	if opts.CommitStrategy != "" {
		q.Set("commit_strategy", opts.CommitStrategy)
	}

	// VAD settings (only relevant when using VAD commit strategy)
	if opts.CommitStrategy == "vad" {
		if opts.VADSilenceThresholdSecs > 0 {
			q.Set("vad_silence_threshold_secs", fmt.Sprintf("%.2f", opts.VADSilenceThresholdSecs))
		}
		if opts.VADThreshold > 0 {
			q.Set("vad_threshold", fmt.Sprintf("%.2f", opts.VADThreshold))
		}
		if opts.MinSpeechDurationMs > 0 {
			q.Set("min_speech_duration_ms", fmt.Sprintf("%d", opts.MinSpeechDurationMs))
		}
		if opts.MinSilenceDurationMs > 0 {
			q.Set("min_silence_duration_ms", fmt.Sprintf("%d", opts.MinSilenceDurationMs))
		}
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (wsc *WebSocketSTTConnection) sendJSON(msg any) error {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()

	if wsc.closed {
		return fmt.Errorf("connection closed")
	}

	return wsc.conn.WriteJSON(msg)
}

func (wsc *WebSocketSTTConnection) readLoop() {
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

		// Parse message to determine type
		var resp sttResponse
		if err := json.Unmarshal(message, &resp); err != nil {
			select {
			case wsc.errChan <- fmt.Errorf("failed to parse response: %w", err):
			default:
			}
			continue
		}

		switch resp.MessageType {
		case "session_started":
			// Session established, store session ID
			wsc.mu.Lock()
			wsc.sessionID = resp.SessionID
			wsc.mu.Unlock()

		case "partial_transcript":
			// Partial (interim) transcription result
			transcript := &STTTranscript{
				Text:    resp.Text,
				IsFinal: false,
			}
			select {
			case wsc.transcriptOut <- transcript:
			case <-wsc.closeChan:
				return
			}

		case "committed_transcript":
			// Final transcription result (without timestamps)
			transcript := &STTTranscript{
				Text:    resp.Text,
				IsFinal: true,
			}
			select {
			case wsc.transcriptOut <- transcript:
			case <-wsc.closeChan:
				return
			}

		case "committed_transcript_with_timestamps":
			// Final transcription result with word-level timestamps
			transcript := &STTTranscript{
				Text:         resp.Text,
				IsFinal:      true,
				LanguageCode: resp.LanguageCode,
				Words:        resp.Words,
			}
			select {
			case wsc.transcriptOut <- transcript:
			case <-wsc.closeChan:
				return
			}

		case "error", "auth_error", "quota_exceeded", "rate_limited",
			"input_error", "transcriber_error", "chunk_size_exceeded",
			"insufficient_audio_activity", "session_time_limit_exceeded",
			"resource_exhausted", "queue_overflow", "commit_throttled",
			"unaccepted_terms":
			// Error response
			errMsg := resp.Error
			if errMsg == "" {
				errMsg = resp.MessageType
			}
			select {
			case wsc.errChan <- fmt.Errorf("server error (%s): %s", resp.MessageType, errMsg):
			default:
			}

		default:
			// Unknown message type, ignore
		}
	}
}

func (wsc *WebSocketSTTConnection) closeChannels() {
	wsc.closeOnce.Do(func() {
		close(wsc.closeChan)
		close(wsc.transcriptOut)
	})
}

// SendAudio sends audio data for transcription.
// The audio should be in the format specified in WebSocketSTTOptions.AudioFormat.
func (wsc *WebSocketSTTConnection) SendAudio(audio []byte) error {
	if len(audio) == 0 {
		return nil
	}

	msg := sttInputAudioChunk{
		MessageType: "input_audio_chunk",
		AudioBase64: base64.StdEncoding.EncodeToString(audio),
	}

	return wsc.sendJSON(msg)
}

// SendAudioWithCommit sends audio data and optionally commits the transcript.
// When commit is true, the server will finalize the current transcript segment.
// This is useful for manual commit strategy.
func (wsc *WebSocketSTTConnection) SendAudioWithCommit(audio []byte, commit bool) error {
	msg := sttInputAudioChunk{
		MessageType: "input_audio_chunk",
		AudioBase64: base64.StdEncoding.EncodeToString(audio),
		Commit:      commit,
	}

	return wsc.sendJSON(msg)
}

// Commit forces a commit of the current transcript segment.
// This sends an empty audio chunk with commit=true.
func (wsc *WebSocketSTTConnection) Commit() error {
	msg := sttInputAudioChunk{
		MessageType: "input_audio_chunk",
		AudioBase64: "",
		Commit:      true,
	}

	return wsc.sendJSON(msg)
}

// Transcripts returns a channel that receives transcription results.
func (wsc *WebSocketSTTConnection) Transcripts() <-chan *STTTranscript {
	return wsc.transcriptOut
}

// Errors returns a channel that receives errors from the connection.
func (wsc *WebSocketSTTConnection) Errors() <-chan error {
	return wsc.errChan
}

// SessionID returns the session ID assigned by the server.
// This is available after the connection is established.
func (wsc *WebSocketSTTConnection) SessionID() string {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()
	return wsc.sessionID
}

// Close closes the WebSocket connection gracefully.
func (wsc *WebSocketSTTConnection) Close() error {
	wsc.mu.Lock()
	if wsc.closed {
		wsc.mu.Unlock()
		return nil
	}
	wsc.closed = true
	wsc.mu.Unlock()

	// Close the connection
	wsc.closeChannels()
	return wsc.conn.Close()
}

// StreamAudio is a convenience method that streams audio from a channel.
// It handles committing automatically when the input channel closes.
func (wsc *WebSocketSTTConnection) StreamAudio(ctx context.Context, audioStream <-chan []byte) (<-chan *STTTranscript, <-chan error) {
	transcriptOut := make(chan *STTTranscript, 100)
	errOut := make(chan error, 1)

	go func() {
		defer close(transcriptOut)
		defer close(errOut)

		// Forward transcripts from connection
		done := make(chan struct{})
		go func() {
			defer close(done)
			for transcript := range wsc.Transcripts() {
				select {
				case transcriptOut <- transcript:
				case <-ctx.Done():
					return
				}
			}
		}()

		// Send audio as it arrives
		for {
			select {
			case audio, ok := <-audioStream:
				if !ok {
					// Input stream closed, commit final transcript and wait
					if err := wsc.Commit(); err != nil {
						// Commit error is non-fatal, connection might be closing
					}
					<-done
					return
				}
				if err := wsc.SendAudio(audio); err != nil {
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

	return transcriptOut, errOut
}
