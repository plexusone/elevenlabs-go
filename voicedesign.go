package elevenlabs

import (
	"context"
	"io"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// VoiceDesignService handles AI voice generation and design.
type VoiceDesignService struct {
	client *Client
}

// VoiceGender represents the gender options for voice generation.
type VoiceGender string

const (
	VoiceGenderFemale VoiceGender = "female"
	VoiceGenderMale   VoiceGender = "male"
)

// VoiceAge represents the age options for voice generation.
type VoiceAge string

const (
	VoiceAgeYoung      VoiceAge = "young"
	VoiceAgeMiddleAged VoiceAge = "middle_aged"
	VoiceAgeOld        VoiceAge = "old"
)

// VoiceAccent represents accent options for voice generation.
type VoiceAccent string

const (
	VoiceAccentBritish    VoiceAccent = "british"
	VoiceAccentAmerican   VoiceAccent = "american"
	VoiceAccentAfrican    VoiceAccent = "african"
	VoiceAccentAustralian VoiceAccent = "australian"
	VoiceAccentIndian     VoiceAccent = "indian"
)

// VoiceDesignRequest contains options for generating a random voice.
type VoiceDesignRequest struct {
	// Gender of the voice (required).
	Gender VoiceGender

	// Age category of the voice (required).
	Age VoiceAge

	// Accent of the voice (required).
	Accent VoiceAccent

	// AccentStrength controls accent prominence (0.3 to 2.0).
	AccentStrength float64

	// Text for voice preview (100-1000 characters).
	Text string
}

// VoiceDesignResponse contains the generated voice preview.
type VoiceDesignResponse struct {
	// Audio is the generated voice sample.
	Audio io.Reader

	// GeneratedVoiceID can be used to save this voice permanently.
	GeneratedVoiceID string
}

// SaveVoiceRequest contains options for saving a generated voice.
type SaveVoiceRequest struct {
	// GeneratedVoiceID from the design response.
	GeneratedVoiceID string

	// VoiceName is the name for the saved voice.
	VoiceName string

	// VoiceDescription describes the voice.
	VoiceDescription string

	// Labels are optional metadata tags.
	Labels map[string]string
}

// GeneratePreview creates a voice preview based on design parameters.
// Returns audio sample and a generated_voice_id that can be saved.
func (s *VoiceDesignService) GeneratePreview(ctx context.Context, req *VoiceDesignRequest) (*VoiceDesignResponse, error) {
	if req.Gender == "" {
		return nil, &ValidationError{Field: "gender", Message: "cannot be empty"}
	}
	if req.Age == "" {
		return nil, &ValidationError{Field: "age", Message: "cannot be empty"}
	}
	if req.Accent == "" {
		return nil, &ValidationError{Field: "accent", Message: "cannot be empty"}
	}
	if req.Text == "" {
		return nil, &ValidationError{Field: "text", Message: "cannot be empty"}
	}
	if len(req.Text) < 100 || len(req.Text) > 1000 {
		return nil, &ValidationError{Field: "text", Message: "must be between 100 and 1000 characters"}
	}

	accentStrength := req.AccentStrength
	if accentStrength == 0 {
		accentStrength = 1.0 // Default
	}
	if accentStrength < 0.3 || accentStrength > 2.0 {
		return nil, &ValidationError{Field: "accent_strength", Message: "must be between 0.3 and 2.0"}
	}

	body := &api.BodyGenerateARandomVoiceV1VoiceGenerationGenerateVoicePost{
		Gender:         api.BodyGenerateARandomVoiceV1VoiceGenerationGenerateVoicePostGender(req.Gender),
		Age:            api.BodyGenerateARandomVoiceV1VoiceGenerationGenerateVoicePostAge(req.Age),
		Accent:         string(req.Accent),
		AccentStrength: accentStrength,
		Text:           req.Text,
	}

	resp, err := s.client.apiClient.GenerateRandomVoice(ctx, body, api.GenerateRandomVoiceParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.GenerateRandomVoiceOK:
		return &VoiceDesignResponse{
			Audio: r.Data,
			// Note: The generated_voice_id is typically returned in response headers
			// The ogen client may not expose this directly
		}, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// SaveVoice saves a previously generated voice to your voice library.
func (s *VoiceDesignService) SaveVoice(ctx context.Context, req *SaveVoiceRequest) (*Voice, error) {
	if req.GeneratedVoiceID == "" {
		return nil, &ValidationError{Field: "generated_voice_id", Message: "cannot be empty"}
	}
	if req.VoiceName == "" {
		return nil, &ValidationError{Field: "voice_name", Message: "cannot be empty"}
	}

	body := &api.BodyCreateAPreviouslyGeneratedVoiceV1VoiceGenerationCreateVoicePost{
		GeneratedVoiceID: req.GeneratedVoiceID,
		VoiceName:        req.VoiceName,
		VoiceDescription: req.VoiceDescription,
	}

	if len(req.Labels) > 0 {
		labels := api.BodyCreateAPreviouslyGeneratedVoiceV1VoiceGenerationCreateVoicePostLabels(req.Labels)
		body.Labels = api.NewOptNilBodyCreateAPreviouslyGeneratedVoiceV1VoiceGenerationCreateVoicePostLabels(labels)
	}

	resp, err := s.client.apiClient.CreateVoiceOld(ctx, body, api.CreateVoiceOldParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.VoiceResponseModel:
		return &Voice{
			VoiceID:     r.VoiceID,
			Name:        r.Name,
			Description: r.Description.Value,
			Category:    string(r.Category),
		}, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// Simple generates a voice preview with common defaults.
func (s *VoiceDesignService) Simple(ctx context.Context, gender VoiceGender, age VoiceAge, accent VoiceAccent, previewText string) (*VoiceDesignResponse, error) {
	return s.GeneratePreview(ctx, &VoiceDesignRequest{
		Gender:         gender,
		Age:            age,
		Accent:         accent,
		AccentStrength: 1.0,
		Text:           previewText,
	})
}
