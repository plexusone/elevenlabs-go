// Package stt provides an OmniVoice STT provider implementation using ElevenLabs.
package stt

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sync"

	elevenlabs "github.com/agentplexus/go-elevenlabs"
	"github.com/agentplexus/go-elevenlabs/omnivoice"
	"github.com/agentplexus/omnivoice/stt"
)

// Verify interface compliance at compile time.
var (
	_ stt.Provider          = (*Provider)(nil)
	_ stt.StreamingProvider = (*Provider)(nil)
)

// Provider implements stt.StreamingProvider using the ElevenLabs API.
type Provider struct {
	client *elevenlabs.Client
}

// Option configures the Provider.
type Option func(*options)

type options struct {
	apiKey  string
	baseURL string
}

// WithAPIKey sets the ElevenLabs API key.
func WithAPIKey(apiKey string) Option {
	return func(o *options) {
		o.apiKey = apiKey
	}
}

// WithBaseURL sets a custom API base URL.
func WithBaseURL(baseURL string) Option {
	return func(o *options) {
		o.baseURL = baseURL
	}
}

// New creates a new ElevenLabs STT provider.
func New(opts ...Option) (*Provider, error) {
	cfg := &options{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Build client options
	var clientOpts []elevenlabs.Option
	if cfg.apiKey != "" {
		clientOpts = append(clientOpts, elevenlabs.WithAPIKey(cfg.apiKey))
	}
	if cfg.baseURL != "" {
		clientOpts = append(clientOpts, elevenlabs.WithBaseURL(cfg.baseURL))
	}

	client, err := elevenlabs.NewClient(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create ElevenLabs client: %w", err)
	}

	return &Provider{
		client: client,
	}, nil
}

// NewWithClient creates a Provider with an existing ElevenLabs client.
func NewWithClient(client *elevenlabs.Client) *Provider {
	return &Provider{
		client: client,
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return omnivoice.ProviderName
}

// Transcribe converts audio bytes to text.
func (p *Provider) Transcribe(ctx context.Context, audio []byte, config stt.TranscriptionConfig) (*stt.TranscriptionResult, error) {
	// Base64 encode the audio
	encoded := base64.StdEncoding.EncodeToString(audio)

	req := &elevenlabs.TranscriptionRequest{
		FileContent:  encoded,
		LanguageCode: config.Language,
		Diarize:      config.EnableSpeakerDiarization,
		NumSpeakers:  config.MaxSpeakers,
		ModelID:      config.Model,
	}

	resp, err := p.client.SpeechToText().Transcribe(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs stt failed: %w", err)
	}

	return omnivoice.TranscriptionResultFromResponse(resp), nil
}

// TranscribeFile transcribes audio from a file path.
func (p *Provider) TranscribeFile(ctx context.Context, filePath string, config stt.TranscriptionConfig) (*stt.TranscriptionResult, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.Transcribe(ctx, data, config)
}

// TranscribeURL transcribes audio from a URL.
func (p *Provider) TranscribeURL(ctx context.Context, url string, config stt.TranscriptionConfig) (*stt.TranscriptionResult, error) {
	req := &elevenlabs.TranscriptionRequest{
		FileURL:      url,
		LanguageCode: config.Language,
		Diarize:      config.EnableSpeakerDiarization,
		NumSpeakers:  config.MaxSpeakers,
		ModelID:      config.Model,
	}

	resp, err := p.client.SpeechToText().Transcribe(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs stt failed: %w", err)
	}

	return omnivoice.TranscriptionResultFromResponse(resp), nil
}

// TranscribeStream starts a streaming transcription session.
func (p *Provider) TranscribeStream(ctx context.Context, config stt.TranscriptionConfig) (io.WriteCloser, <-chan stt.StreamEvent, error) {
	opts := omnivoice.ConfigToWebSocketSTTOptions(config)

	conn, err := p.client.WebSocketSTT().Connect(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect WebSocket STT: %w", err)
	}

	eventCh := make(chan stt.StreamEvent, 100)
	writer := &streamWriter{
		conn:    conn,
		eventCh: eventCh,
		ctx:     ctx,
	}

	// Start forwarding transcripts
	go func() {
		defer close(eventCh)

		for {
			select {
			case transcript, ok := <-conn.Transcripts():
				if !ok {
					return
				}
				event := omnivoice.TranscriptToStreamEvent(transcript)
				select {
				case eventCh <- event:
				case <-ctx.Done():
					eventCh <- stt.StreamEvent{Type: stt.EventError, Error: ctx.Err()}
					return
				}
			case err := <-conn.Errors():
				if err != nil {
					select {
					case eventCh <- stt.StreamEvent{Type: stt.EventError, Error: err}:
					case <-ctx.Done():
						return
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return writer, eventCh, nil
}

// streamWriter wraps the WebSocket connection to implement io.WriteCloser.
type streamWriter struct {
	conn    *elevenlabs.WebSocketSTTConnection
	eventCh chan stt.StreamEvent
	ctx     context.Context
	mu      sync.Mutex
	closed  bool
}

// Write sends audio data to the transcription service.
func (w *streamWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, io.ErrClosedPipe
	}

	if err := w.conn.SendAudio(p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close commits the final transcript and closes the connection.
func (w *streamWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}
	w.closed = true

	// Commit the final transcript
	if err := w.conn.Commit(); err != nil {
		// Log but don't fail - connection might already be closing
		_ = err
	}

	return w.conn.Close()
}

// Client returns the underlying ElevenLabs client for advanced operations.
func (p *Provider) Client() *elevenlabs.Client {
	return p.client
}
