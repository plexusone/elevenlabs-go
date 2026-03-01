package elevenlabs

import (
	"context"
	"io"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// SoundEffectsService handles sound effect generation.
type SoundEffectsService struct {
	client *Client
}

// SoundEffectRequest contains options for generating a sound effect.
type SoundEffectRequest struct {
	// Text is the description of the sound effect to generate.
	// Examples: "car engine starting", "thunder and rain", "crowd cheering"
	Text string

	// DurationSeconds is the target duration (0.5 to 30 seconds).
	// If not set, the optimal duration will be guessed from the prompt.
	DurationSeconds float64

	// PromptInfluence controls how closely the generation follows the prompt (0.0 to 1.0).
	// Higher values = more faithful to prompt but less variation.
	// Default is 0.3.
	PromptInfluence float64

	// Loop creates a sound effect that loops smoothly.
	Loop bool

	// OutputFormat specifies the audio format (e.g., "mp3_44100_128").
	OutputFormat string
}

// Validate validates the sound effect request.
func (r *SoundEffectRequest) Validate() error {
	if r.Text == "" {
		return ErrEmptyText
	}
	if r.DurationSeconds != 0 && (r.DurationSeconds < 0.5 || r.DurationSeconds > 30) {
		return &ValidationError{Field: "duration_seconds", Message: "must be between 0.5 and 30"}
	}
	if r.PromptInfluence != 0 && (r.PromptInfluence < 0 || r.PromptInfluence > 1) {
		return &ValidationError{Field: "prompt_influence", Message: "must be between 0 and 1"}
	}
	return nil
}

// SoundEffectResponse contains the generated sound effect.
type SoundEffectResponse struct {
	// Audio is the generated sound effect data.
	Audio io.Reader
}

// Generate creates a sound effect from a text description.
func (s *SoundEffectsService) Generate(ctx context.Context, req *SoundEffectRequest) (*SoundEffectResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	body := &api.BodySoundGenerationV1SoundGenerationPost{
		Text: req.Text,
	}

	if req.DurationSeconds > 0 {
		body.DurationSeconds = api.NewOptNilFloat64(req.DurationSeconds)
	}
	if req.PromptInfluence > 0 {
		body.PromptInfluence = api.NewOptNilFloat64(req.PromptInfluence)
	}
	if req.Loop {
		body.Loop = api.NewOptBool(true)
	}

	params := api.SoundGenerationParams{}
	if req.OutputFormat != "" {
		params.OutputFormat = api.NewOptSoundGenerationOutputFormat(
			api.SoundGenerationOutputFormat(req.OutputFormat),
		)
	}

	resp, err := s.client.apiClient.SoundGeneration(ctx, body, params)
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.SoundGenerationOKHeaders:
		return &SoundEffectResponse{Audio: r.Response.Data}, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// Simple generates a sound effect with minimal configuration.
func (s *SoundEffectsService) Simple(ctx context.Context, description string) (io.Reader, error) {
	resp, err := s.Generate(ctx, &SoundEffectRequest{
		Text: description,
	})
	if err != nil {
		return nil, err
	}
	return resp.Audio, nil
}

// GenerateLoop generates a looping sound effect.
func (s *SoundEffectsService) GenerateLoop(ctx context.Context, description string, durationSeconds float64) (io.Reader, error) {
	resp, err := s.Generate(ctx, &SoundEffectRequest{
		Text:            description,
		DurationSeconds: durationSeconds,
		Loop:            true,
	})
	if err != nil {
		return nil, err
	}
	return resp.Audio, nil
}
