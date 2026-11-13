package tts

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// Service handles text-to-speech, speech-to-speech, and dialogue generation.
type Service struct {
	apiClient *api.Client
	apiKey    string
	baseURL   string
}

// New creates a new TTS service.
func New(apiClient *api.Client, apiKey, baseURL string) *Service {
	return &Service{
		apiClient: apiClient,
		apiKey:    apiKey,
		baseURL:   baseURL,
	}
}

// DefaultModelID is the recommended model for text-to-speech.
const DefaultModelID = "eleven_multilingual_v2"

// DefaultSTSModelID is the default model for speech-to-speech.
const DefaultSTSModelID = "eleven_english_sts_v2"

// VoiceSettings contains the voice configuration for generation.
type VoiceSettings struct {
	// Stability determines how stable the voice is (0.0 to 1.0).
	// Lower values introduce broader emotional range.
	Stability float64

	// SimilarityBoost determines how closely the AI should adhere to
	// the original voice (0.0 to 1.0).
	SimilarityBoost float64

	// Style determines the style exaggeration (0.0 to 1.0).
	// Higher values amplify the original speaker's style.
	Style float64

	// Speed adjusts the speed of the voice (0.25 to 4.0).
	// 1.0 is the default speed.
	Speed float64

	// UseSpeakerBoost boosts similarity to the original speaker.
	UseSpeakerBoost bool
}

// DefaultVoiceSettings returns sensible default voice settings.
func DefaultVoiceSettings() *VoiceSettings {
	return &VoiceSettings{
		Stability:       0.5,
		SimilarityBoost: 0.75,
		Style:           0.0,
		Speed:           1.0,
		UseSpeakerBoost: true,
	}
}

// Validate validates the voice settings.
func (vs *VoiceSettings) Validate() error {
	if vs.Stability < 0 || vs.Stability > 1 {
		return fmt.Errorf("tts: stability must be between 0.0 and 1.0")
	}
	if vs.SimilarityBoost < 0 || vs.SimilarityBoost > 1 {
		return fmt.Errorf("tts: similarity_boost must be between 0.0 and 1.0")
	}
	if vs.Style < 0 || vs.Style > 1 {
		return fmt.Errorf("tts: style must be between 0.0 and 1.0")
	}
	if vs.Speed != 0 && (vs.Speed < 0.25 || vs.Speed > 4.0) {
		return fmt.Errorf("tts: speed must be between 0.25 and 4.0")
	}
	return nil
}

// Voice settings presets for different platforms and use cases.
// These presets are tuned for specific content types and platforms.

// VoiceSettingsForUdemy returns settings tuned for Udemy courses.
// Neutral, clear, consistent, safe for long lectures.
func VoiceSettingsForUdemy() *VoiceSettings {
	return &VoiceSettings{
		Stability:       0.5,
		SimilarityBoost: 0.75,
		Style:           0.05,
		Speed:           1.0,
		UseSpeakerBoost: true,
	}
}

// VoiceSettingsForCoursera returns settings tuned for Coursera courses.
// Slightly expressive, engaging for mixed media content.
func VoiceSettingsForCoursera() *VoiceSettings {
	return &VoiceSettings{
		Stability:       0.7,
		SimilarityBoost: 0.85,
		Style:           0.2,
		Speed:           1.0,
		UseSpeakerBoost: true,
	}
}

// VoiceSettingsForEdX returns settings tuned for edX courses.
// Very stable, highly intelligible, slightly faster for dense academic content.
func VoiceSettingsForEdX() *VoiceSettings {
	return &VoiceSettings{
		Stability:       0.8,
		SimilarityBoost: 0.9,
		Style:           0.15,
		Speed:           1.05,
		UseSpeakerBoost: true,
	}
}

// VoiceSettingsForInstagram returns settings tuned for Instagram content.
// Energetic but polished, suitable for brand content.
func VoiceSettingsForInstagram() *VoiceSettings {
	return &VoiceSettings{
		Stability:       0.4,
		SimilarityBoost: 0.85,
		Style:           0.35,
		Speed:           1.1,
		UseSpeakerBoost: true,
	}
}

// VoiceSettingsForTikTok returns settings tuned for TikTok content.
// Designed for immediate engagement in the first 1-3 seconds.
func VoiceSettingsForTikTok() *VoiceSettings {
	return &VoiceSettings{
		Stability:       0.3,
		SimilarityBoost: 0.85,
		Style:           0.45,
		Speed:           1.15,
		UseSpeakerBoost: true,
	}
}

// VoiceSettingsForYouTube returns settings tuned for YouTube content.
// Designed to hold attention for 5-20 minutes without sounding robotic or theatrical.
func VoiceSettingsForYouTube() *VoiceSettings {
	return &VoiceSettings{
		Stability:       0.45,
		SimilarityBoost: 0.8,
		Style:           0.2,
		Speed:           1.05,
		UseSpeakerBoost: true,
	}
}

// VoiceSettingsForPodcast returns settings tuned for podcast content.
// Natural conversational tone for long-form audio content.
func VoiceSettingsForPodcast() *VoiceSettings {
	return &VoiceSettings{
		Stability:       0.55,
		SimilarityBoost: 0.75,
		Style:           0.15,
		Speed:           1.0,
		UseSpeakerBoost: true,
	}
}

// VoiceSettingsForAudiobook returns settings tuned for audiobook narration.
// Clear, consistent, easy to listen to for extended periods.
func VoiceSettingsForAudiobook() *VoiceSettings {
	return &VoiceSettings{
		Stability:       0.65,
		SimilarityBoost: 0.8,
		Style:           0.1,
		Speed:           0.95,
		UseSpeakerBoost: true,
	}
}

// --- Text-to-Speech ---

// Request is a request to generate speech from text.
type Request struct {
	// VoiceID is the voice to use for generation.
	VoiceID string

	// Text is the text to convert to speech.
	Text string

	// ModelID is the model to use. Defaults to DefaultModelID.
	ModelID string

	// VoiceSettings configures the voice parameters.
	// If nil, default settings will be used.
	VoiceSettings *VoiceSettings

	// OutputFormat specifies the audio output format.
	// Examples: "mp3_44100_128", "pcm_16000", "pcm_22050"
	OutputFormat string

	// LanguageCode is the ISO 639-1 language code for text normalization.
	LanguageCode string

	// PronunciationDictionaries are applied in order; ElevenLabs accepts up
	// to MaxPronunciationDictionaries.
	PronunciationDictionaries []PronunciationDictionaryLocator

	// PreviousText and NextText are the text spoken before and after this
	// request, which keep the delivery continuous when long text is split
	// across requests.
	PreviousText string
	NextText     string
}

// MaxPronunciationDictionaries is the most dictionaries one request takes.
const MaxPronunciationDictionaries = 3

// PronunciationDictionaryLocator names a pronunciation dictionary.
type PronunciationDictionaryLocator struct {
	// ID is the dictionary's ID.
	ID string
	// VersionID is the dictionary version; empty means the latest.
	VersionID string
}

// phonemeModels are the models that honour SSML phoneme tags and the
// phoneme rules of pronunciation dictionaries. Other models read the text
// inside a phoneme tag as written; alias rules work with every model.
var phonemeModels = map[string]bool{
	"eleven_flash_v2":       true,
	"eleven_turbo_v2":       true,
	"eleven_monolingual_v1": true,
}

// SupportsPhonemes reports whether a model honours phoneme tags. An empty
// model means DefaultModelID.
func SupportsPhonemes(modelID string) bool {
	if modelID == "" {
		modelID = DefaultModelID
	}
	return phonemeModels[modelID]
}

// ValidOutputFormats lists the valid audio output formats.
var ValidOutputFormats = map[string]bool{
	// MP3 formats (lossy, widely compatible)
	"mp3_22050_32":  true,
	"mp3_24000_48":  true,
	"mp3_44100_32":  true,
	"mp3_44100_64":  true,
	"mp3_44100_96":  true,
	"mp3_44100_128": true, // default
	"mp3_44100_192": true, // highest quality MP3
	// PCM formats (lossless raw audio, can be wrapped in WAV)
	"pcm_8000":  true,
	"pcm_16000": true,
	"pcm_22050": true,
	"pcm_24000": true,
	"pcm_32000": true,
	"pcm_44100": true, // CD quality
	"pcm_48000": true, // highest quality
	// Telephony formats
	"ulaw_8000": true,
	"alaw_8000": true,
	// Opus formats (efficient lossy codec)
	"opus_48000_32":  true,
	"opus_48000_64":  true,
	"opus_48000_96":  true,
	"opus_48000_128": true,
	"opus_48000_192": true,
}

// Validate validates the TTS request.
func (r *Request) Validate() error {
	if r.VoiceID == "" {
		return fmt.Errorf("tts: voice ID is required")
	}
	if r.Text == "" {
		return fmt.Errorf("tts: text cannot be empty")
	}
	if r.VoiceSettings != nil {
		if err := r.VoiceSettings.Validate(); err != nil {
			return err
		}
	}
	if r.OutputFormat != "" && !ValidOutputFormats[r.OutputFormat] {
		return fmt.Errorf("tts: invalid output format %q", r.OutputFormat)
	}
	if len(r.PronunciationDictionaries) > MaxPronunciationDictionaries {
		return fmt.Errorf("tts: at most %d pronunciation dictionaries, got %d", MaxPronunciationDictionaries, len(r.PronunciationDictionaries))
	}
	for _, d := range r.PronunciationDictionaries {
		if d.ID == "" {
			return fmt.Errorf("tts: pronunciation dictionary ID is required")
		}
	}
	return nil
}

// Response contains the generated audio from text-to-speech.
type Response struct {
	// Audio is the generated audio data.
	Audio io.Reader
}

// Generate generates speech from text.
func (s *Service) Generate(ctx context.Context, req *Request) (*Response, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Build request body
	body := &api.BodyTextToSpeechFull{
		Text: req.Text,
	}

	// Set model ID
	modelID := req.ModelID
	if modelID == "" {
		modelID = DefaultModelID
	}
	body.ModelID = api.NewOptString(modelID)

	// Set voice settings if provided
	if req.VoiceSettings != nil {
		vs := api.VoiceSettingsResponseModel{
			Stability:       api.NewOptNilFloat64(req.VoiceSettings.Stability),
			SimilarityBoost: api.NewOptNilFloat64(req.VoiceSettings.SimilarityBoost),
			Style:           api.NewOptNilFloat64(req.VoiceSettings.Style),
		}
		if req.VoiceSettings.Speed != 0 {
			vs.Speed = api.NewOptNilFloat64(req.VoiceSettings.Speed)
		}
		body.VoiceSettings = api.NewOptVoiceSettingsResponseModel(vs)
	}

	// Set language code if provided
	if req.LanguageCode != "" {
		body.LanguageCode = api.NewOptNilString(req.LanguageCode)
	}
	if len(req.PronunciationDictionaries) > 0 {
		locators := make([]api.PronunciationDictionaryVersionLocatorRequestModel, len(req.PronunciationDictionaries))
		for i, d := range req.PronunciationDictionaries {
			locators[i] = api.PronunciationDictionaryVersionLocatorRequestModel{PronunciationDictionaryID: d.ID}
			if d.VersionID != "" {
				locators[i].VersionID = api.NewOptNilString(d.VersionID)
			}
		}
		body.PronunciationDictionaryLocators = api.NewOptNilPronunciationDictionaryVersionLocatorRequestModelArray(locators)
	}
	if req.PreviousText != "" {
		body.PreviousText = api.NewOptNilString(req.PreviousText)
	}
	if req.NextText != "" {
		body.NextText = api.NewOptNilString(req.NextText)
	}

	// Build params
	params := api.TextToSpeechFullParams{
		VoiceID: req.VoiceID,
	}

	// Set output format if provided
	if req.OutputFormat != "" {
		params.OutputFormat = api.NewOptTextToSpeechFullOutputFormat(
			api.TextToSpeechFullOutputFormat(req.OutputFormat),
		)
	}

	// Make the API call
	resp, err := s.apiClient.TextToSpeechFull(ctx, body, params)
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.TextToSpeechFullOK:
		return &Response{Audio: r.Data}, nil
	default:
		return nil, fmt.Errorf("tts: unexpected response type")
	}
}

// GenerateToWriter generates speech and writes it to a writer.
func (s *Service) GenerateToWriter(ctx context.Context, req *Request, w io.Writer) error {
	resp, err := s.Generate(ctx, req)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, resp.Audio)
	return err
}

// Simple is a convenience method that generates speech with minimal parameters.
func (s *Service) Simple(ctx context.Context, voiceID, text string) (io.Reader, error) {
	resp, err := s.Generate(ctx, &Request{
		VoiceID:       voiceID,
		Text:          text,
		VoiceSettings: DefaultVoiceSettings(),
	})
	if err != nil {
		return nil, err
	}
	return resp.Audio, nil
}

// --- Speech-to-Speech ---

// SpeechToSpeechRequest is a request to convert speech to a different voice.
type SpeechToSpeechRequest struct {
	// VoiceID is the target voice to convert to.
	VoiceID string

	// Audio is the source audio data to convert.
	Audio io.Reader

	// AudioFilename is the filename for the audio (optional, helps with format detection).
	AudioFilename string

	// ModelID is the model to use. Defaults to DefaultSTSModelID.
	ModelID string

	// VoiceSettings configures the voice parameters.
	VoiceSettings *VoiceSettings

	// OutputFormat specifies the audio output format.
	OutputFormat string

	// RemoveBackgroundNoise removes background noise from the source audio.
	RemoveBackgroundNoise bool

	// SeedAudio is optional seed audio to influence the conversion.
	SeedAudio io.Reader

	// SeedAudioFilename is the filename for the seed audio.
	SeedAudioFilename string
}

// Validate validates the speech-to-speech request.
func (r *SpeechToSpeechRequest) Validate() error {
	if r.VoiceID == "" {
		return fmt.Errorf("tts: voice ID is required")
	}
	if r.Audio == nil {
		return fmt.Errorf("tts: audio is required")
	}
	if r.VoiceSettings != nil {
		if err := r.VoiceSettings.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// SpeechToSpeechResponse contains the converted audio.
type SpeechToSpeechResponse struct {
	// Audio is the converted audio data.
	Audio io.Reader
}

// ConvertSpeech converts speech from one voice to another.
func (s *Service) ConvertSpeech(ctx context.Context, req *SpeechToSpeechRequest) (*SpeechToSpeechResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Build multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add audio file
	audioFilename := req.AudioFilename
	if audioFilename == "" {
		audioFilename = "audio.mp3"
	}
	audioWriter, err := writer.CreateFormFile("audio", audioFilename)
	if err != nil {
		return nil, fmt.Errorf("tts: failed to create audio form field: %w", err)
	}
	if _, err := io.Copy(audioWriter, req.Audio); err != nil {
		return nil, fmt.Errorf("tts: failed to write audio: %w", err)
	}

	// Add model ID
	modelID := req.ModelID
	if modelID == "" {
		modelID = DefaultSTSModelID
	}
	if err := writer.WriteField("model_id", modelID); err != nil {
		return nil, fmt.Errorf("tts: failed to write model_id: %w", err)
	}

	// Add voice settings if provided
	if req.VoiceSettings != nil {
		if err := writer.WriteField("stability", fmt.Sprintf("%.2f", req.VoiceSettings.Stability)); err != nil {
			return nil, err
		}
		if err := writer.WriteField("similarity_boost", fmt.Sprintf("%.2f", req.VoiceSettings.SimilarityBoost)); err != nil {
			return nil, err
		}
		if req.VoiceSettings.Style > 0 {
			if err := writer.WriteField("style", fmt.Sprintf("%.2f", req.VoiceSettings.Style)); err != nil {
				return nil, err
			}
		}
		if req.VoiceSettings.UseSpeakerBoost {
			if err := writer.WriteField("use_speaker_boost", "true"); err != nil {
				return nil, err
			}
		}
	}

	// Add remove background noise option
	if req.RemoveBackgroundNoise {
		if err := writer.WriteField("remove_background_noise", "true"); err != nil {
			return nil, err
		}
	}

	// Add seed audio if provided
	if req.SeedAudio != nil {
		seedFilename := req.SeedAudioFilename
		if seedFilename == "" {
			seedFilename = "seed.mp3"
		}
		seedWriter, err := writer.CreateFormFile("seed_audio", seedFilename)
		if err != nil {
			return nil, fmt.Errorf("tts: failed to create seed_audio form field: %w", err)
		}
		if _, err := io.Copy(seedWriter, req.SeedAudio); err != nil {
			return nil, fmt.Errorf("tts: failed to write seed audio: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("tts: failed to close multipart writer: %w", err)
	}

	// Build URL
	url := fmt.Sprintf("%s/v1/speech-to-speech/%s", s.baseURL, req.VoiceID)
	if req.OutputFormat != "" {
		url += "?output_format=" + req.OutputFormat
	}

	// Make request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("xi-api-key", s.apiKey)

	resp, err := http.DefaultClient.Do(httpReq) //nolint:gosec // G704: API client, URL is fixed ElevenLabs endpoint
	if err != nil {
		return nil, fmt.Errorf("tts: request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tts: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return &SpeechToSpeechResponse{Audio: resp.Body}, nil
}

// ConvertSpeechStream converts speech with streaming response.
func (s *Service) ConvertSpeechStream(ctx context.Context, req *SpeechToSpeechRequest) (*SpeechToSpeechResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Build multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add audio file
	audioFilename := req.AudioFilename
	if audioFilename == "" {
		audioFilename = "audio.mp3"
	}
	audioWriter, err := writer.CreateFormFile("audio", audioFilename)
	if err != nil {
		return nil, fmt.Errorf("tts: failed to create audio form field: %w", err)
	}
	if _, err := io.Copy(audioWriter, req.Audio); err != nil {
		return nil, fmt.Errorf("tts: failed to write audio: %w", err)
	}

	// Add model ID
	modelID := req.ModelID
	if modelID == "" {
		modelID = DefaultSTSModelID
	}
	if err := writer.WriteField("model_id", modelID); err != nil {
		return nil, fmt.Errorf("tts: failed to write model_id: %w", err)
	}

	// Add voice settings if provided
	if req.VoiceSettings != nil {
		if err := writer.WriteField("stability", fmt.Sprintf("%.2f", req.VoiceSettings.Stability)); err != nil {
			return nil, err
		}
		if err := writer.WriteField("similarity_boost", fmt.Sprintf("%.2f", req.VoiceSettings.SimilarityBoost)); err != nil {
			return nil, err
		}
		if req.VoiceSettings.Style > 0 {
			if err := writer.WriteField("style", fmt.Sprintf("%.2f", req.VoiceSettings.Style)); err != nil {
				return nil, err
			}
		}
		if req.VoiceSettings.UseSpeakerBoost {
			if err := writer.WriteField("use_speaker_boost", "true"); err != nil {
				return nil, err
			}
		}
	}

	// Add remove background noise option
	if req.RemoveBackgroundNoise {
		if err := writer.WriteField("remove_background_noise", "true"); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("tts: failed to close multipart writer: %w", err)
	}

	// Build URL for streaming endpoint
	url := fmt.Sprintf("%s/v1/speech-to-speech/%s/stream", s.baseURL, req.VoiceID)
	if req.OutputFormat != "" {
		url += "?output_format=" + req.OutputFormat
	}

	// Make request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("xi-api-key", s.apiKey)

	resp, err := http.DefaultClient.Do(httpReq) //nolint:gosec // G704: API client, URL is fixed ElevenLabs endpoint
	if err != nil {
		return nil, fmt.Errorf("tts: request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tts: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return &SpeechToSpeechResponse{Audio: resp.Body}, nil
}

// SimpleSpeechToSpeech is a convenience method for basic voice conversion.
func (s *Service) SimpleSpeechToSpeech(ctx context.Context, voiceID string, audio io.Reader) (io.Reader, error) {
	resp, err := s.ConvertSpeech(ctx, &SpeechToSpeechRequest{
		VoiceID:       voiceID,
		Audio:         audio,
		VoiceSettings: DefaultVoiceSettings(),
	})
	if err != nil {
		return nil, err
	}
	return resp.Audio, nil
}

// --- Dialogue ---

// DialogueInput represents a single dialogue turn with text and voice.
type DialogueInput struct {
	// Text is the text to be spoken.
	Text string

	// VoiceID is the ID of the voice to use.
	VoiceID string
}

// DialogueRequest contains options for dialogue generation.
type DialogueRequest struct {
	// Inputs is a list of dialogue turns with text and voice pairs.
	Inputs []DialogueInput

	// ModelID is the model to use (default: eleven_multilingual_v2).
	ModelID string

	// LanguageCode is the ISO 639-1 language code (e.g., "en").
	LanguageCode string

	// Seed for deterministic generation (0-4294967295).
	Seed int
}

// DialogueResponse contains the dialogue generation result with timestamps.
type DialogueResponse struct {
	// AudioBase64 is the base64-encoded audio data.
	AudioBase64 string

	// VoiceSegments contains timing info for each voice segment.
	VoiceSegments []VoiceSegment
}

// VoiceSegment represents a segment of audio for a specific voice.
type VoiceSegment struct {
	// VoiceID is the voice used for this segment.
	VoiceID string

	// StartTime is the start time in seconds.
	StartTime float64

	// EndTime is the end time in seconds.
	EndTime float64
}

// GenerateDialogue creates dialogue audio from multiple voice inputs.
// Returns an io.Reader containing the combined audio.
//
//nolint:dupl // Similar to GenerateDialogueStream but uses different ogen-generated types
func (s *Service) GenerateDialogue(ctx context.Context, req *DialogueRequest) (io.Reader, error) {
	if len(req.Inputs) == 0 {
		return nil, fmt.Errorf("tts: inputs cannot be empty")
	}

	// Convert inputs
	inputs := make([]api.DialogueInput, len(req.Inputs))
	for i, input := range req.Inputs {
		inputs[i] = api.DialogueInput{
			Text:    input.Text,
			VoiceID: input.VoiceID,
		}
	}

	body := &api.BodyTextToDialogueMultiVoiceV1TextToDialoguePost{
		Inputs: inputs,
	}

	if req.ModelID != "" {
		body.ModelID = api.NewOptString(req.ModelID)
	}
	if req.LanguageCode != "" {
		body.LanguageCode = api.NewOptNilString(req.LanguageCode)
	}
	if req.Seed > 0 {
		body.Seed = api.NewOptNilInt(req.Seed)
	}

	resp, err := s.apiClient.TextToDialogue(ctx, body, api.TextToDialogueParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.TextToDialogueOK:
		return r.Data, nil
	default:
		return nil, fmt.Errorf("tts: unexpected response type")
	}
}

// GenerateDialogueWithTimestamps creates dialogue audio with timing information.
func (s *Service) GenerateDialogueWithTimestamps(ctx context.Context, req *DialogueRequest) (*DialogueResponse, error) {
	if len(req.Inputs) == 0 {
		return nil, fmt.Errorf("tts: inputs cannot be empty")
	}

	// Convert inputs
	inputs := make([]api.DialogueInput, len(req.Inputs))
	for i, input := range req.Inputs {
		inputs[i] = api.DialogueInput{
			Text:    input.Text,
			VoiceID: input.VoiceID,
		}
	}

	body := &api.BodyTextToDialogueFullWithTimestamps{
		Inputs: inputs,
	}

	if req.ModelID != "" {
		body.ModelID = api.NewOptString(req.ModelID)
	}
	if req.LanguageCode != "" {
		body.LanguageCode = api.NewOptNilString(req.LanguageCode)
	}
	if req.Seed > 0 {
		body.Seed = api.NewOptNilInt(req.Seed)
	}

	resp, err := s.apiClient.TextToDialogueFullWithTimestamps(ctx, body, api.TextToDialogueFullWithTimestampsParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.AudioWithTimestampsAndVoiceSegmentsResponseModel:
		result := &DialogueResponse{
			AudioBase64: r.AudioBase64,
		}

		// Convert voice segments
		for _, seg := range r.VoiceSegments {
			result.VoiceSegments = append(result.VoiceSegments, VoiceSegment{
				VoiceID:   seg.VoiceID,
				StartTime: seg.StartTimeSeconds,
				EndTime:   seg.EndTimeSeconds,
			})
		}

		return result, nil
	default:
		return nil, fmt.Errorf("tts: unexpected response type")
	}
}

// GenerateDialogueStream creates dialogue audio with streaming output.
//
//nolint:dupl // Similar to GenerateDialogue but uses different ogen-generated types
func (s *Service) GenerateDialogueStream(ctx context.Context, req *DialogueRequest) (io.Reader, error) {
	if len(req.Inputs) == 0 {
		return nil, fmt.Errorf("tts: inputs cannot be empty")
	}

	// Convert inputs
	inputs := make([]api.DialogueInput, len(req.Inputs))
	for i, input := range req.Inputs {
		inputs[i] = api.DialogueInput{
			Text:    input.Text,
			VoiceID: input.VoiceID,
		}
	}

	body := &api.BodyTextToDialogueMultiVoiceStreamingV1TextToDialogueStreamPost{
		Inputs: inputs,
	}

	if req.ModelID != "" {
		body.ModelID = api.NewOptString(req.ModelID)
	}
	if req.LanguageCode != "" {
		body.LanguageCode = api.NewOptNilString(req.LanguageCode)
	}
	if req.Seed > 0 {
		body.Seed = api.NewOptNilInt(req.Seed)
	}

	resp, err := s.apiClient.TextToDialogueStream(ctx, body, api.TextToDialogueStreamParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.TextToDialogueStreamOK:
		return r.Data, nil
	default:
		return nil, fmt.Errorf("tts: unexpected response type")
	}
}

// SimpleDialogue generates dialogue audio from text-voice pairs with default settings.
func (s *Service) SimpleDialogue(ctx context.Context, inputs []DialogueInput) (io.Reader, error) {
	return s.GenerateDialogue(ctx, &DialogueRequest{Inputs: inputs})
}
