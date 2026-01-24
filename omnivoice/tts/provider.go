// Package tts provides an OmniVoice TTS provider implementation using ElevenLabs.
package tts

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	elevenlabs "github.com/agentplexus/go-elevenlabs"
	"github.com/agentplexus/go-elevenlabs/omnivoice"
	"github.com/agentplexus/omnivoice/tts"
)

// Verify interface compliance at compile time.
var (
	_ tts.Provider          = (*Provider)(nil)
	_ tts.StreamingProvider = (*Provider)(nil)
)

// Provider implements tts.StreamingProvider using the ElevenLabs API.
type Provider struct {
	client *elevenlabs.Client

	// Cache for voices
	voicesMu    sync.RWMutex
	voicesCache []tts.Voice
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

// New creates a new ElevenLabs TTS provider.
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

// Synthesize converts text to speech and returns audio data.
func (p *Provider) Synthesize(ctx context.Context, text string, config tts.SynthesisConfig) (*tts.SynthesisResult, error) {
	req := omnivoice.ConfigToTTSRequest(text, config)

	resp, err := p.client.TextToSpeech().Generate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs tts failed: %w", err)
	}

	// Read all audio data
	audioData, err := io.ReadAll(resp.Audio)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio: %w", err)
	}

	result := &tts.SynthesisResult{
		Audio:          audioData,
		Format:         config.OutputFormat,
		SampleRate:     config.SampleRate,
		CharacterCount: len(text),
	}

	// Set defaults if not specified
	if result.Format == "" {
		result.Format = "mp3"
	}
	if result.SampleRate == 0 {
		result.SampleRate = 44100
	}

	return result, nil
}

// SynthesizeStream converts text to speech with streaming output.
func (p *Provider) SynthesizeStream(ctx context.Context, text string, config tts.SynthesisConfig) (<-chan tts.StreamChunk, error) {
	opts := omnivoice.ConfigToWebSocketTTSOptions(config)
	// Use shorter inactivity timeout - ElevenLabs closes connection on timeout
	// which signals end of audio stream
	if opts.InactivityTimeout == 0 {
		opts.InactivityTimeout = 5 // 5 seconds
	}

	conn, err := p.client.WebSocketTTS().Connect(ctx, config.VoiceID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect WebSocket TTS: %w", err)
	}

	out := make(chan tts.StreamChunk, 100)

	go func() {
		defer close(out)
		defer func() { _ = conn.Close() }()

		var receivedAudio bool

		// Send the text
		if err := conn.SendText(text); err != nil {
			out <- tts.StreamChunk{Error: fmt.Errorf("failed to send text: %w", err)}
			return
		}

		// Flush to get remaining audio
		if err := conn.Flush(); err != nil {
			out <- tts.StreamChunk{Error: fmt.Errorf("failed to flush: %w", err)}
			return
		}

		// Forward audio chunks until done
		for {
			select {
			case audio, ok := <-conn.Audio():
				if !ok {
					// Audio channel closed
					out <- tts.StreamChunk{IsFinal: true}
					return
				}
				receivedAudio = true
				out <- tts.StreamChunk{Audio: audio}
			case <-conn.Done():
				// All audio received after flush - drain any remaining audio
				for audio := range conn.Audio() {
					out <- tts.StreamChunk{Audio: audio}
				}
				out <- tts.StreamChunk{IsFinal: true}
				return
			case err := <-conn.Errors():
				// ElevenLabs signals end of stream via inactivity timeout
				// If we received audio and flushed, treat timeout as success
				if receivedAudio && isInactivityTimeout(err) {
					out <- tts.StreamChunk{IsFinal: true}
					return
				}
				if err != nil {
					out <- tts.StreamChunk{Error: err}
				}
				return
			case <-ctx.Done():
				out <- tts.StreamChunk{Error: ctx.Err()}
				return
			}
		}
	}()

	return out, nil
}

// isInactivityTimeout checks if the error is an ElevenLabs inactivity timeout.
func isInactivityTimeout(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "input_timeout_exceeded")
}

// SynthesizeFromReader reads text from a reader and streams audio output.
// This is useful for streaming LLM output directly to TTS.
func (p *Provider) SynthesizeFromReader(ctx context.Context, reader io.Reader, config tts.SynthesisConfig) (<-chan tts.StreamChunk, error) {
	opts := omnivoice.ConfigToWebSocketTTSOptions(config)
	// Use shorter inactivity timeout - ElevenLabs closes connection on timeout
	// which signals end of audio stream
	if opts.InactivityTimeout == 0 {
		opts.InactivityTimeout = 5 // 5 seconds
	}

	conn, err := p.client.WebSocketTTS().Connect(ctx, config.VoiceID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect WebSocket TTS: %w", err)
	}

	out := make(chan tts.StreamChunk, 100)

	go func() {
		defer close(out)
		defer func() { _ = conn.Close() }()

		var receivedAudio bool

		// Read text in chunks and send
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				if sendErr := conn.SendText(string(buf[:n])); sendErr != nil {
					out <- tts.StreamChunk{Error: fmt.Errorf("failed to send text: %w", sendErr)}
					return
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				out <- tts.StreamChunk{Error: fmt.Errorf("failed to read text: %w", err)}
				return
			}
		}

		// Flush to get remaining audio
		if err := conn.Flush(); err != nil {
			out <- tts.StreamChunk{Error: fmt.Errorf("failed to flush: %w", err)}
			return
		}

		// Forward audio chunks until done
		for {
			select {
			case audio, ok := <-conn.Audio():
				if !ok {
					// Audio channel closed
					out <- tts.StreamChunk{IsFinal: true}
					return
				}
				receivedAudio = true
				out <- tts.StreamChunk{Audio: audio}
			case <-conn.Done():
				// All audio received after flush - drain any remaining audio
				for audio := range conn.Audio() {
					out <- tts.StreamChunk{Audio: audio}
				}
				out <- tts.StreamChunk{IsFinal: true}
				return
			case err := <-conn.Errors():
				// ElevenLabs signals end of stream via inactivity timeout
				// If we received audio and flushed, treat timeout as success
				if receivedAudio && isInactivityTimeout(err) {
					out <- tts.StreamChunk{IsFinal: true}
					return
				}
				if err != nil {
					out <- tts.StreamChunk{Error: err}
				}
				return
			case <-ctx.Done():
				out <- tts.StreamChunk{Error: ctx.Err()}
				return
			}
		}
	}()

	return out, nil
}

// ListVoices returns available voices from ElevenLabs.
func (p *Provider) ListVoices(ctx context.Context) ([]tts.Voice, error) {
	// Check cache first
	p.voicesMu.RLock()
	if p.voicesCache != nil {
		cached := make([]tts.Voice, len(p.voicesCache))
		copy(cached, p.voicesCache)
		p.voicesMu.RUnlock()
		return cached, nil
	}
	p.voicesMu.RUnlock()

	// Fetch from API
	voices, err := p.client.Voices().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list voices: %w", err)
	}

	result := make([]tts.Voice, 0, len(voices))
	for _, v := range voices {
		result = append(result, omnivoice.VoiceToOmniVoice(v))
	}

	// Update cache
	p.voicesMu.Lock()
	p.voicesCache = result
	p.voicesMu.Unlock()

	return result, nil
}

// GetVoice returns a specific voice by ID.
func (p *Provider) GetVoice(ctx context.Context, voiceID string) (*tts.Voice, error) {
	voice, err := p.client.Voices().Get(ctx, voiceID)
	if err != nil {
		// Parse the error to get API-level details
		apiErr := elevenlabs.ParseAPIError(err)
		if apiErr != nil && (apiErr.StatusCode == 404 || apiErr.StatusCode == 400) {
			// ElevenLabs returns 400 or 404 for invalid voice IDs
			return nil, tts.ErrVoiceNotFound
		}
		return nil, fmt.Errorf("failed to get voice: %w", err)
	}

	result := omnivoice.VoiceToOmniVoice(voice)
	return &result, nil
}

// ClearVoiceCache clears the cached voices list.
func (p *Provider) ClearVoiceCache() {
	p.voicesMu.Lock()
	p.voicesCache = nil
	p.voicesMu.Unlock()
}

// Client returns the underlying ElevenLabs client for advanced operations.
func (p *Provider) Client() *elevenlabs.Client {
	return p.client
}
