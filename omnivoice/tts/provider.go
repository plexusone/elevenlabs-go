// Package tts provides an OmniVoice TTS provider implementation using ElevenLabs.
package tts

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	elevenlabs "github.com/plexusone/elevenlabs-go"
	"github.com/plexusone/elevenlabs-go/omnivoice"
	"github.com/plexusone/omnivoice-core/resilience"
	"github.com/plexusone/omnivoice-core/tts"
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

	// AX-aware resilience
	classifier  *omnivoice.Classifier
	retryConfig resilience.RetryConfig
}

// Option configures the Provider.
type Option func(*options)

type options struct {
	apiKey      string
	baseURL     string
	retryConfig *resilience.RetryConfig
	onRetry     func(attempt int, err error, delay time.Duration)
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

// WithRetryConfig sets a custom retry configuration.
// If not set, defaults to omnivoice.DefaultRetryConfig().
func WithRetryConfig(config resilience.RetryConfig) Option {
	return func(o *options) {
		o.retryConfig = &config
	}
}

// WithOnRetry sets a callback to be called before each retry attempt.
// Useful for logging or metrics collection.
func WithOnRetry(fn func(attempt int, err error, delay time.Duration)) Option {
	return func(o *options) {
		o.onRetry = fn
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

	// Set up retry configuration
	classifier := omnivoice.NewClassifier()
	retryConfig := omnivoice.DefaultRetryConfig()
	if cfg.retryConfig != nil {
		retryConfig = *cfg.retryConfig
	}
	if cfg.onRetry != nil {
		retryConfig.OnRetry = cfg.onRetry
	}
	// Ensure classifier is our AX-aware classifier
	retryConfig.Classifier = classifier

	return &Provider{
		client:      client,
		classifier:  classifier,
		retryConfig: retryConfig,
	}, nil
}

// NewWithClient creates a Provider with an existing ElevenLabs client.
func NewWithClient(client *elevenlabs.Client) *Provider {
	classifier := omnivoice.NewClassifier()
	retryConfig := omnivoice.DefaultRetryConfig()
	retryConfig.Classifier = classifier

	return &Provider{
		client:      client,
		classifier:  classifier,
		retryConfig: retryConfig,
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return omnivoice.ProviderName
}

// Synthesize converts text to speech and returns audio data.
// It automatically retries transient errors (rate limits, server errors) using
// exponential backoff with the AX-aware error classifier.
func (p *Provider) Synthesize(ctx context.Context, text string, config tts.SynthesisConfig) (*tts.SynthesisResult, error) {
	req := omnivoice.ConfigToTTSRequest(text, config)

	result, err := resilience.RetryWithResult(ctx, p.retryConfig, func() (*tts.SynthesisResult, error) {
		resp, err := p.client.TextToSpeech().Generate(ctx, req)
		if err != nil {
			return nil, p.classifier.WrapError("Synthesize", err)
		}

		// Read all audio data
		audioData, err := io.ReadAll(resp.Audio)
		if err != nil {
			return nil, fmt.Errorf("failed to read audio: %w", err)
		}

		r := &tts.SynthesisResult{
			Audio:          audioData,
			Format:         config.OutputFormat,
			SampleRate:     config.SampleRate,
			CharacterCount: len(text),
		}

		// Set defaults if not specified
		if r.Format == "" {
			r.Format = "mp3"
		}
		if r.SampleRate == 0 {
			r.SampleRate = 44100
		}

		return r, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// SynthesizeStream converts text to speech with streaming output.
// Connection errors are classified using the AX-aware error classifier.
func (p *Provider) SynthesizeStream(ctx context.Context, text string, config tts.SynthesisConfig) (<-chan tts.StreamChunk, error) {
	opts := omnivoice.ConfigToWebSocketTTSOptions(config)
	// Use shorter inactivity timeout - ElevenLabs closes connection on timeout
	// which signals end of audio stream
	if opts.InactivityTimeout == 0 {
		opts.InactivityTimeout = 5 // 5 seconds
	}

	conn, err := p.client.WebSocketTTS().Connect(ctx, config.VoiceID, opts)
	if err != nil {
		return nil, p.classifier.WrapError("SynthesizeStream", err)
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
// Connection errors are classified using the AX-aware error classifier.
func (p *Provider) SynthesizeFromReader(ctx context.Context, reader io.Reader, config tts.SynthesisConfig) (<-chan tts.StreamChunk, error) {
	opts := omnivoice.ConfigToWebSocketTTSOptions(config)
	// Use shorter inactivity timeout - ElevenLabs closes connection on timeout
	// which signals end of audio stream
	if opts.InactivityTimeout == 0 {
		opts.InactivityTimeout = 5 // 5 seconds
	}

	conn, err := p.client.WebSocketTTS().Connect(ctx, config.VoiceID, opts)
	if err != nil {
		return nil, p.classifier.WrapError("SynthesizeFromReader", err)
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
// It automatically retries transient errors using exponential backoff.
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

	// Fetch from API with retry
	result, err := resilience.RetryWithResult(ctx, p.retryConfig, func() ([]tts.Voice, error) {
		voices, err := p.client.Voices().List(ctx)
		if err != nil {
			return nil, p.classifier.WrapError("ListVoices", err)
		}

		r := make([]tts.Voice, 0, len(voices))
		for _, v := range voices {
			r = append(r, omnivoice.VoiceToOmniVoice(v))
		}
		return r, nil
	})

	if err != nil {
		return nil, err
	}

	// Update cache
	p.voicesMu.Lock()
	p.voicesCache = result
	p.voicesMu.Unlock()

	return result, nil
}

// GetVoice returns a specific voice by ID.
// It automatically retries transient errors using exponential backoff.
// Returns tts.ErrVoiceNotFound for invalid voice IDs.
func (p *Provider) GetVoice(ctx context.Context, voiceID string) (*tts.Voice, error) {
	result, err := resilience.RetryWithResult(ctx, p.retryConfig, func() (*tts.Voice, error) {
		voice, err := p.client.Voices().Get(ctx, voiceID)
		if err != nil {
			// Parse the error to get API-level details
			apiErr := elevenlabs.ParseAPIError(err)
			if apiErr != nil && (apiErr.StatusCode == 404 || apiErr.StatusCode == 400) {
				// ElevenLabs returns 400 or 404 for invalid voice IDs
				// Return as non-retryable error
				return nil, tts.ErrVoiceNotFound
			}
			return nil, p.classifier.WrapError("GetVoice", err)
		}

		r := omnivoice.VoiceToOmniVoice(voice)
		return &r, nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
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

// Classifier returns the AX-aware error classifier.
func (p *Provider) Classifier() *omnivoice.Classifier {
	return p.classifier
}

// RetryConfig returns the current retry configuration.
func (p *Provider) RetryConfig() resilience.RetryConfig {
	return p.retryConfig
}

// SetRetryConfig updates the retry configuration.
func (p *Provider) SetRetryConfig(config resilience.RetryConfig) {
	p.retryConfig = config
	// Ensure classifier is set
	if p.retryConfig.Classifier == nil {
		p.retryConfig.Classifier = p.classifier
	}
}
