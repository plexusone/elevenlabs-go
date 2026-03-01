package elevenlabs

import (
	"context"
	"io"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// TextToDialogueService handles multi-voice dialogue generation.
type TextToDialogueService struct {
	client *Client
}

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

// Generate creates dialogue audio from multiple voice inputs.
// Returns an io.Reader containing the combined audio.
//
//nolint:dupl // Similar to GenerateStream but uses different ogen-generated types
func (s *TextToDialogueService) Generate(ctx context.Context, req *DialogueRequest) (io.Reader, error) {
	if len(req.Inputs) == 0 {
		return nil, &ValidationError{Field: "inputs", Message: "cannot be empty"}
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

	resp, err := s.client.apiClient.TextToDialogue(ctx, body, api.TextToDialogueParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.TextToDialogueOK:
		return r.Data, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// GenerateWithTimestamps creates dialogue audio with timing information.
func (s *TextToDialogueService) GenerateWithTimestamps(ctx context.Context, req *DialogueRequest) (*DialogueResponse, error) {
	if len(req.Inputs) == 0 {
		return nil, &ValidationError{Field: "inputs", Message: "cannot be empty"}
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

	resp, err := s.client.apiClient.TextToDialogueFullWithTimestamps(ctx, body, api.TextToDialogueFullWithTimestampsParams{})
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
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// GenerateStream creates dialogue audio with streaming output.
//
//nolint:dupl // Similar to Generate but uses different ogen-generated types
func (s *TextToDialogueService) GenerateStream(ctx context.Context, req *DialogueRequest) (io.Reader, error) {
	if len(req.Inputs) == 0 {
		return nil, &ValidationError{Field: "inputs", Message: "cannot be empty"}
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

	resp, err := s.client.apiClient.TextToDialogueStream(ctx, body, api.TextToDialogueStreamParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.TextToDialogueStreamOK:
		return r.Data, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// Simple generates dialogue audio from text-voice pairs with default settings.
func (s *TextToDialogueService) Simple(ctx context.Context, inputs []DialogueInput) (io.Reader, error) {
	return s.Generate(ctx, &DialogueRequest{Inputs: inputs})
}
