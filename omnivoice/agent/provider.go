// Package agent provides an OmniVoice Agent provider implementation using ElevenLabs.
package agent

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	elevenlabs "github.com/agentplexus/go-elevenlabs"
	"github.com/agentplexus/go-elevenlabs/omnivoice"
	"github.com/agentplexus/omnivoice/agent"
)

// Verify interface compliance at compile time.
var _ agent.Provider = (*Provider)(nil)

// Provider implements agent.Provider using ElevenLabs real-time services.
type Provider struct {
	client *elevenlabs.Client

	// Session management
	mu       sync.RWMutex
	sessions map[string]*Session
	counter  uint64
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

// New creates a new ElevenLabs Agent provider.
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
		client:   client,
		sessions: make(map[string]*Session),
	}, nil
}

// NewWithClient creates a Provider with an existing ElevenLabs client.
func NewWithClient(client *elevenlabs.Client) *Provider {
	return &Provider{
		client:   client,
		sessions: make(map[string]*Session),
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return omnivoice.ProviderName
}

// CreateSession creates a new voice session.
func (p *Provider) CreateSession(ctx context.Context, config agent.Config) (agent.Session, error) {
	// Generate session ID
	id := fmt.Sprintf("elevenlabs-%d-%d", time.Now().UnixNano(), atomic.AddUint64(&p.counter, 1))

	session := &Session{
		id:       id,
		config:   config,
		client:   p.client,
		events:   make(chan agent.Event, 100),
		audioIn:  make(chan []byte, 100),
		audioOut: make(chan []byte, 100),
		done:     make(chan struct{}),
		started:  time.Now(),
	}

	// Register session
	p.mu.Lock()
	p.sessions[id] = session
	p.mu.Unlock()

	return session, nil
}

// GetSession retrieves an existing session by ID.
func (p *Provider) GetSession(ctx context.Context, sessionID string) (agent.Session, error) {
	p.mu.RLock()
	session, ok := p.sessions[sessionID]
	p.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// ListSessions lists active session IDs.
func (p *Provider) ListSessions(ctx context.Context) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ids := make([]string, 0, len(p.sessions))
	for id := range p.sessions {
		ids = append(ids, id)
	}
	return ids, nil
}

// RemoveSession removes a session from the registry.
// Call this after a session is stopped to clean up resources.
func (p *Provider) RemoveSession(id string) {
	p.mu.Lock()
	delete(p.sessions, id)
	p.mu.Unlock()
}

// Client returns the underlying ElevenLabs client for advanced operations.
func (p *Provider) Client() *elevenlabs.Client {
	return p.client
}

// Session implements agent.Session using ElevenLabs WebSocket TTS and STT.
type Session struct {
	id     string
	config agent.Config
	client *elevenlabs.Client

	// Connections
	ttsConn *elevenlabs.WebSocketTTSConnection
	sttConn *elevenlabs.WebSocketSTTConnection

	// Channels
	events   chan agent.Event
	audioIn  chan []byte
	audioOut chan []byte
	done     chan struct{}

	// State
	mu         sync.RWMutex
	started    time.Time
	transcript []agent.Turn
	metrics    agent.Metrics
	running    bool
	closed     bool
}

// ID returns the unique session identifier.
func (s *Session) ID() string {
	return s.id
}

// Start begins the voice session.
func (s *Session) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("session already started")
	}
	s.running = true
	s.mu.Unlock()

	// Connect TTS
	ttsOpts := &elevenlabs.WebSocketTTSOptions{
		ModelID:                  "eleven_turbo_v2_5",
		OutputFormat:             "pcm_16000",
		OptimizeStreamingLatency: 3,
	}

	var err error
	s.ttsConn, err = s.client.WebSocketTTS().Connect(ctx, s.config.VoiceID, ttsOpts)
	if err != nil {
		return fmt.Errorf("failed to connect TTS: %w", err)
	}

	// Connect STT
	sttOpts := &elevenlabs.WebSocketSTTOptions{
		ModelID:           "scribe_v2_realtime",
		AudioFormat:       "pcm_16000",
		IncludeTimestamps: true,
		LanguageCode:      s.config.Language,
	}

	s.sttConn, err = s.client.WebSocketSTT().Connect(ctx, sttOpts)
	if err != nil {
		_ = s.ttsConn.Close() // Best effort cleanup
		return fmt.Errorf("failed to connect STT: %w", err)
	}

	// Start audio routing
	go s.routeAudio(ctx)

	// Send session started event
	s.events <- agent.Event{
		Type:      agent.EventSessionStarted,
		Timestamp: time.Now(),
	}

	return nil
}

// Stop ends the voice session gracefully.
func (s *Session) Stop(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.running = false
	s.mu.Unlock()

	close(s.done)

	// Close connections (best effort, log errors internally)
	var closeErr error
	if s.ttsConn != nil {
		if err := s.ttsConn.Close(); err != nil {
			closeErr = err
		}
	}
	if s.sttConn != nil {
		if err := s.sttConn.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	_ = closeErr // Connection close errors are non-fatal

	// Send session ended event
	s.events <- agent.Event{
		Type:      agent.EventSessionEnded,
		Timestamp: time.Now(),
	}

	close(s.events)
	close(s.audioOut)

	return nil
}

// SendAudio sends audio data to the agent.
func (s *Session) SendAudio(audio []byte) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return fmt.Errorf("session closed")
	}
	s.mu.RUnlock()

	select {
	case s.audioIn <- audio:
		return nil
	default:
		return fmt.Errorf("audio buffer full")
	}
}

// ReceiveAudio returns a channel for receiving agent audio.
func (s *Session) ReceiveAudio() <-chan []byte {
	return s.audioOut
}

// SendText sends text input to the agent (bypass STT).
func (s *Session) SendText(text string) error {
	s.mu.RLock()
	if s.closed || s.ttsConn == nil {
		s.mu.RUnlock()
		return fmt.Errorf("session not started or closed")
	}
	s.mu.RUnlock()

	// Record turn
	s.addTurn("user", text)

	// Send event
	s.events <- agent.Event{
		Type:      agent.EventUserTranscript,
		Timestamp: time.Now(),
		Data:      text,
	}

	return nil
}

// Events returns a channel for session events.
func (s *Session) Events() <-chan agent.Event {
	return s.events
}

// Transcript returns the conversation transcript so far.
func (s *Session) Transcript() []agent.Turn {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]agent.Turn, len(s.transcript))
	copy(result, s.transcript)
	return result
}

// Metrics returns session performance metrics.
func (s *Session) Metrics() agent.Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate session duration
	s.metrics.SessionDurationMs = int(time.Since(s.started).Milliseconds())
	s.metrics.TurnCount = len(s.transcript)

	return s.metrics
}

// routeAudio handles audio routing between STT, processing, and TTS.
func (s *Session) routeAudio(ctx context.Context) {
	// Forward user audio to STT
	go func() {
		for {
			select {
			case audio := <-s.audioIn:
				if s.sttConn != nil {
					if err := s.sttConn.SendAudio(audio); err != nil {
						s.events <- agent.Event{
							Type:      agent.EventError,
							Timestamp: time.Now(),
							Error:     err,
						}
					}
				}
			case <-s.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	// Forward STT transcripts to events
	go func() {
		for {
			select {
			case transcript, ok := <-s.sttConn.Transcripts():
				if !ok {
					return
				}
				if transcript.IsFinal {
					s.events <- agent.Event{
						Type:      agent.EventUserTranscript,
						Timestamp: time.Now(),
						Data:      transcript.Text,
					}
					s.addTurn("user", transcript.Text)
				}
			case <-s.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	// Forward TTS audio to output
	go func() {
		for {
			select {
			case audio, ok := <-s.ttsConn.Audio():
				if !ok {
					return
				}
				select {
				case s.audioOut <- audio:
				default:
					// Buffer full, skip
				}
			case <-s.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	// Forward TTS errors
	go func() {
		for {
			select {
			case err, ok := <-s.ttsConn.Errors():
				if !ok {
					return
				}
				if err != nil {
					s.events <- agent.Event{
						Type:      agent.EventError,
						Timestamp: time.Now(),
						Error:     err,
					}
				}
			case <-s.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	// Forward STT errors
	go func() {
		for {
			select {
			case err, ok := <-s.sttConn.Errors():
				if !ok {
					return
				}
				if err != nil {
					s.events <- agent.Event{
						Type:      agent.EventError,
						Timestamp: time.Now(),
						Error:     err,
					}
				}
			case <-s.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// addTurn adds a turn to the transcript.
func (s *Session) addTurn(role, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.transcript = append(s.transcript, agent.Turn{
		Role:      role,
		Text:      text,
		Timestamp: time.Now(),
	})
}

// SpeakText sends text to be spoken by the agent.
func (s *Session) SpeakText(text string) error {
	s.mu.RLock()
	if s.closed || s.ttsConn == nil {
		s.mu.RUnlock()
		return fmt.Errorf("session not started or closed")
	}
	s.mu.RUnlock()

	// Send event
	s.events <- agent.Event{
		Type:      agent.EventAgentSpeechStart,
		Timestamp: time.Now(),
	}

	// Send text to TTS
	if err := s.ttsConn.SendText(text); err != nil {
		return fmt.Errorf("failed to send text: %w", err)
	}

	// Flush to generate audio
	if err := s.ttsConn.Flush(); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

	// Record turn
	s.addTurn("agent", text)

	// Send transcript event
	s.events <- agent.Event{
		Type:      agent.EventAgentTranscript,
		Timestamp: time.Now(),
		Data:      text,
	}

	return nil
}
