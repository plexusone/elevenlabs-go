// Package elevenlabs provides a Go client for the ElevenLabs API.
//
// The client wraps the ogen-generated API client with a higher-level
// interface that handles authentication and provides convenient methods
// for common operations.
package elevenlabs

import (
	"net/http"
	"os"
	"time"

	"github.com/agentplexus/go-elevenlabs/internal/api"
)

// Version is the SDK version.
const Version = "0.3.0"

// DefaultBaseURL is the default ElevenLabs API base URL.
const DefaultBaseURL = "https://api.elevenlabs.io"

// DefaultModelID is the recommended model for text-to-speech.
const DefaultModelID = "eleven_multilingual_v2"

// Client is the main ElevenLabs client for interacting with the API.
type Client struct {
	apiClient *api.Client
	apiKey    string
	baseURL   string

	// Service accessors
	tts             *TextToSpeechService
	voices          *VoicesService
	models          *ModelsService
	history         *HistoryService
	user            *UserService
	dubbing         *DubbingService
	soundEffects    *SoundEffectsService
	pronunciation   *PronunciationService
	projects        *ProjectsService
	speechToText    *SpeechToTextService
	forcedAlignment *ForcedAlignmentService
	audioIsolation  *AudioIsolationService
	textToDialogue  *TextToDialogueService
	voiceDesign     *VoiceDesignService
	music           *MusicService

	// Real-time services
	webSocketTTS   *WebSocketTTSService
	webSocketSTT   *WebSocketSTTService
	twilio         *TwilioService
	phoneNumbers   *PhoneNumberService
	speechToSpeech *SpeechToSpeechService
}

// NewClient creates a new ElevenLabs client with the given options.
func NewClient(opts ...Option) (*Client, error) {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Try environment variable if API key not set
	if options.apiKey == "" {
		options.apiKey = os.Getenv("ELEVENLABS_API_KEY")
	}

	// Create HTTP client with auth headers
	httpClient := options.httpClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: options.timeout,
		}
	}

	// Wrap with auth transport
	authClient := &authHTTPClient{
		client: httpClient,
		apiKey: options.apiKey,
	}

	// Create the ogen client
	apiClient, err := api.NewClient(
		options.baseURL,
		api.WithClient(authClient),
	)
	if err != nil {
		return nil, err
	}

	c := &Client{
		apiClient: apiClient,
		apiKey:    options.apiKey,
		baseURL:   options.baseURL,
	}

	// Initialize services
	c.tts = &TextToSpeechService{client: c}
	c.voices = &VoicesService{client: c}
	c.models = &ModelsService{client: c}
	c.history = &HistoryService{client: c}
	c.user = &UserService{client: c}
	c.dubbing = &DubbingService{client: c}
	c.soundEffects = &SoundEffectsService{client: c}
	c.pronunciation = &PronunciationService{client: c}
	c.projects = &ProjectsService{client: c}
	c.speechToText = &SpeechToTextService{client: c}
	c.forcedAlignment = &ForcedAlignmentService{client: c}
	c.audioIsolation = &AudioIsolationService{client: c}
	c.textToDialogue = &TextToDialogueService{client: c}
	c.voiceDesign = &VoiceDesignService{client: c}
	c.music = &MusicService{client: c}

	// Initialize real-time services
	c.webSocketTTS = &WebSocketTTSService{client: c}
	c.webSocketSTT = &WebSocketSTTService{client: c}
	c.twilio = &TwilioService{client: c}
	c.phoneNumbers = &PhoneNumberService{client: c}
	c.speechToSpeech = &SpeechToSpeechService{client: c}

	return c, nil
}

// authHTTPClient wraps an http.Client to add authentication headers.
type authHTTPClient struct {
	client *http.Client
	apiKey string
}

// Do implements ht.Client interface.
func (c *authHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Add authentication header
	if c.apiKey != "" {
		req.Header.Set("xi-api-key", c.apiKey)
	}

	// Add SDK version headers
	req.Header.Set("X-ElevenLabs-SDK-Version", Version)
	req.Header.Set("X-ElevenLabs-SDK-Lang", "go")

	return c.client.Do(req) //nolint:gosec // G704: HTTP client library, URL is caller-controlled by design
}

// API returns the underlying ogen-generated API client for advanced usage.
// Use this when you need access to API endpoints not covered by the
// high-level wrapper methods.
func (c *Client) API() *api.Client {
	return c.apiClient
}

// TextToSpeech returns the text-to-speech service.
func (c *Client) TextToSpeech() *TextToSpeechService {
	return c.tts
}

// Voices returns the voices service.
func (c *Client) Voices() *VoicesService {
	return c.voices
}

// Models returns the models service.
func (c *Client) Models() *ModelsService {
	return c.models
}

// History returns the history service.
func (c *Client) History() *HistoryService {
	return c.history
}

// User returns the user service.
func (c *Client) User() *UserService {
	return c.user
}

// Dubbing returns the dubbing service.
func (c *Client) Dubbing() *DubbingService {
	return c.dubbing
}

// SoundEffects returns the sound effects service.
func (c *Client) SoundEffects() *SoundEffectsService {
	return c.soundEffects
}

// Pronunciation returns the pronunciation dictionary service.
func (c *Client) Pronunciation() *PronunciationService {
	return c.pronunciation
}

// Projects returns the projects (Studio) service.
func (c *Client) Projects() *ProjectsService {
	return c.projects
}

// SpeechToText returns the speech-to-text transcription service.
func (c *Client) SpeechToText() *SpeechToTextService {
	return c.speechToText
}

// ForcedAlignment returns the forced alignment service.
func (c *Client) ForcedAlignment() *ForcedAlignmentService {
	return c.forcedAlignment
}

// AudioIsolation returns the audio isolation service.
func (c *Client) AudioIsolation() *AudioIsolationService {
	return c.audioIsolation
}

// TextToDialogue returns the text-to-dialogue service.
func (c *Client) TextToDialogue() *TextToDialogueService {
	return c.textToDialogue
}

// VoiceDesign returns the voice design/generation service.
func (c *Client) VoiceDesign() *VoiceDesignService {
	return c.voiceDesign
}

// Music returns the music composition service.
func (c *Client) Music() *MusicService {
	return c.music
}

// WebSocketTTS returns the WebSocket text-to-speech service for real-time streaming.
func (c *Client) WebSocketTTS() *WebSocketTTSService {
	return c.webSocketTTS
}

// WebSocketSTT returns the WebSocket speech-to-text service for real-time transcription.
func (c *Client) WebSocketSTT() *WebSocketSTTService {
	return c.webSocketSTT
}

// Twilio returns the Twilio phone integration service.
func (c *Client) Twilio() *TwilioService {
	return c.twilio
}

// PhoneNumbers returns the phone number management service.
func (c *Client) PhoneNumbers() *PhoneNumberService {
	return c.phoneNumbers
}

// SpeechToSpeech returns the speech-to-speech voice conversion service.
func (c *Client) SpeechToSpeech() *SpeechToSpeechService {
	return c.speechToSpeech
}

// clientOptions holds the options for creating a Client.
type clientOptions struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

func defaultClientOptions() *clientOptions {
	return &clientOptions{
		baseURL: DefaultBaseURL,
		timeout: 120 * time.Second, // TTS can take a while
	}
}

// Option is a functional option for configuring the Client.
type Option func(*clientOptions)

// WithAPIKey sets the API key for authentication.
func WithAPIKey(apiKey string) Option {
	return func(o *clientOptions) {
		o.apiKey = apiKey
	}
}

// WithBaseURL sets the API base URL.
func WithBaseURL(baseURL string) Option {
	return func(o *clientOptions) {
		o.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(o *clientOptions) {
		o.httpClient = client
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}
