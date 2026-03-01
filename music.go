package elevenlabs

import (
	"context"
	"io"

	ht "github.com/ogen-go/ogen/http"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// MusicService handles music composition and generation.
type MusicService struct {
	client *Client
}

// MusicRequest contains options for music generation.
type MusicRequest struct {
	// Prompt is a simple text description of the music to generate.
	// Cannot be used with CompositionPlan.
	Prompt string

	// DurationMs is the length of the song in milliseconds (3000-600000).
	// If not provided, the model will choose based on the prompt.
	DurationMs int

	// ForceInstrumental ensures the song has no vocals.
	ForceInstrumental bool

	// Seed for deterministic generation (optional).
	Seed int
}

// MusicResponse contains the music generation result.
type MusicResponse struct {
	// Audio is the generated music.
	Audio io.Reader

	// SongID is the unique identifier for this song.
	SongID string
}

// Generate creates music from a text prompt.
//
//nolint:dupl // Similar to GenerateStream but uses different ogen-generated types
func (s *MusicService) Generate(ctx context.Context, req *MusicRequest) (*MusicResponse, error) {
	if req.Prompt == "" {
		return nil, &ValidationError{Field: "prompt", Message: "cannot be empty"}
	}

	body := &api.BodyComposeMusicV1MusicPost{
		Prompt: api.NewOptNilString(req.Prompt),
	}

	if req.DurationMs > 0 {
		body.MusicLengthMs = api.NewOptNilInt(req.DurationMs)
	}
	if req.ForceInstrumental {
		body.ForceInstrumental = api.NewOptBool(true)
	}
	if req.Seed > 0 {
		body.Seed = api.NewOptNilInt(req.Seed)
	}

	resp, err := s.client.apiClient.Generate(ctx, api.NewOptBodyComposeMusicV1MusicPost(*body), api.GenerateParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.GenerateOKHeaders:
		return &MusicResponse{
			Audio:  r.Response.Data,
			SongID: r.SongID.Value,
		}, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// GenerateStream creates music with streaming output.
//
//nolint:dupl // Similar to Generate but uses different ogen-generated types
func (s *MusicService) GenerateStream(ctx context.Context, req *MusicRequest) (*MusicResponse, error) {
	if req.Prompt == "" {
		return nil, &ValidationError{Field: "prompt", Message: "cannot be empty"}
	}

	body := &api.BodyStreamComposedMusicV1MusicStreamPost{
		Prompt: api.NewOptNilString(req.Prompt),
	}

	if req.DurationMs > 0 {
		body.MusicLengthMs = api.NewOptNilInt(req.DurationMs)
	}
	if req.ForceInstrumental {
		body.ForceInstrumental = api.NewOptBool(true)
	}
	if req.Seed > 0 {
		body.Seed = api.NewOptNilInt(req.Seed)
	}

	resp, err := s.client.apiClient.StreamCompose(ctx, api.NewOptBodyStreamComposedMusicV1MusicStreamPost(*body), api.StreamComposeParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.StreamComposeOKHeaders:
		return &MusicResponse{
			Audio:  r.Response.Data,
			SongID: r.SongID.Value,
		}, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// Simple generates music from a prompt with default settings.
func (s *MusicService) Simple(ctx context.Context, prompt string) (io.Reader, error) {
	resp, err := s.Generate(ctx, &MusicRequest{Prompt: prompt})
	if err != nil {
		return nil, err
	}
	return resp.Audio, nil
}

// GenerateInstrumental generates instrumental music from a prompt.
func (s *MusicService) GenerateInstrumental(ctx context.Context, prompt string, durationMs int) (io.Reader, error) {
	resp, err := s.Generate(ctx, &MusicRequest{
		Prompt:            prompt,
		DurationMs:        durationMs,
		ForceInstrumental: true,
	})
	if err != nil {
		return nil, err
	}
	return resp.Audio, nil
}

// CompositionPlan represents a detailed music composition plan.
// This can be used with GenerateDetailed for fine-grained control over music generation.
type CompositionPlan struct {
	// PositiveGlobalStyles are styles that should be present throughout the song.
	PositiveGlobalStyles []string

	// NegativeGlobalStyles are styles that should NOT be present in the song.
	NegativeGlobalStyles []string

	// Sections defines the structure of the song with individual sections.
	Sections []SongSection
}

// SongSection represents a section of a song in a composition plan.
type SongSection struct {
	// SectionName is the name of the section (e.g., "intro", "verse", "chorus").
	SectionName string

	// DurationMs is the duration in milliseconds (3000-120000).
	DurationMs int

	// Lines are the lyrics for this section (max 200 chars per line).
	Lines []string

	// PositiveLocalStyles are styles for this specific section.
	PositiveLocalStyles []string

	// NegativeLocalStyles are styles to avoid in this section.
	NegativeLocalStyles []string
}

// CompositionPlanRequest contains options for generating a composition plan.
type CompositionPlanRequest struct {
	// Prompt is the text description of the music to plan.
	Prompt string

	// DurationMs is the target duration in milliseconds (3000-600000).
	DurationMs int

	// SourcePlan is an optional existing plan to use as a starting point.
	SourcePlan *CompositionPlan
}

// GeneratePlan creates a composition plan from a text prompt.
// The returned plan can be modified and used with GenerateDetailed.
//
// Example:
//
//	plan, err := client.Music().GeneratePlan(ctx, &MusicCompositionPlanRequest{
//	    Prompt:     "upbeat pop song about summer",
//	    DurationMs: 180000, // 3 minutes
//	})
//	// Modify the plan if needed
//	plan.Sections[0].Lines = []string{"Custom lyrics here"}
//	// Generate music from the plan
//	resp, err := client.Music().GenerateDetailed(ctx, &MusicDetailedRequest{
//	    CompositionPlan: plan,
//	})
func (s *MusicService) GeneratePlan(ctx context.Context, req *CompositionPlanRequest) (*CompositionPlan, error) {
	if req.Prompt == "" {
		return nil, &ValidationError{Field: "prompt", Message: "cannot be empty"}
	}

	body := &api.BodyGenerateCompositionPlanV1MusicPlanPost{
		Prompt: req.Prompt,
	}

	if req.DurationMs > 0 {
		body.MusicLengthMs = api.NewOptNilInt(req.DurationMs)
	}

	if req.SourcePlan != nil {
		apiPlan := compositionPlanToAPI(req.SourcePlan)
		body.SourceCompositionPlan = api.NewOptMusicPrompt(apiPlan)
	}

	resp, err := s.client.apiClient.ComposePlan(ctx, body, api.ComposePlanParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.MusicPrompt:
		return compositionPlanFromAPI(r), nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// MusicDetailedRequest contains options for detailed music generation.
type MusicDetailedRequest struct {
	// Prompt is a simple text description (cannot be used with CompositionPlan).
	Prompt string

	// CompositionPlan is a detailed plan (cannot be used with Prompt).
	CompositionPlan *CompositionPlan

	// DurationMs is the length in milliseconds (only used with Prompt).
	DurationMs int

	// ForceInstrumental ensures no vocals (only used with Prompt).
	ForceInstrumental bool

	// Seed for deterministic generation.
	Seed int

	// WithTimestamps returns word timestamps in the response.
	WithTimestamps bool
}

// MusicDetailedResponse contains the detailed music generation result.
type MusicDetailedResponse struct {
	// Audio is the generated music.
	Audio io.Reader

	// SongID is the unique identifier for this song.
	SongID string
}

// GenerateDetailed creates music with detailed options and metadata.
// Use either Prompt for simple generation or CompositionPlan for fine-grained control.
//
// Example with prompt:
//
//	resp, err := client.Music().GenerateDetailed(ctx, &MusicDetailedRequest{
//	    Prompt:     "epic orchestral music",
//	    DurationMs: 60000,
//	})
//
// Example with composition plan:
//
//	plan, _ := client.Music().GeneratePlan(ctx, &CompositionPlanRequest{Prompt: "pop song"})
//	resp, err := client.Music().GenerateDetailed(ctx, &MusicDetailedRequest{
//	    CompositionPlan: plan,
//	})
func (s *MusicService) GenerateDetailed(ctx context.Context, req *MusicDetailedRequest) (*MusicDetailedResponse, error) {
	if req.Prompt == "" && req.CompositionPlan == nil {
		return nil, &ValidationError{Field: "prompt", Message: "either prompt or composition_plan is required"}
	}
	if req.Prompt != "" && req.CompositionPlan != nil {
		return nil, &ValidationError{Field: "prompt", Message: "cannot use both prompt and composition_plan"}
	}

	body := &api.BodyComposeMusicWithADetailedResponseV1MusicDetailedPost{}

	if req.Prompt != "" {
		body.Prompt = api.NewOptNilString(req.Prompt)
		if req.DurationMs > 0 {
			body.MusicLengthMs = api.NewOptNilInt(req.DurationMs)
		}
		if req.ForceInstrumental {
			body.ForceInstrumental = api.NewOptBool(true)
		}
	}

	if req.CompositionPlan != nil {
		apiPlan := compositionPlanToAPI(req.CompositionPlan)
		body.CompositionPlan = api.NewOptMusicPrompt(apiPlan)
	}

	if req.Seed > 0 {
		body.Seed = api.NewOptNilInt(req.Seed)
	}
	if req.WithTimestamps {
		body.WithTimestamps = api.NewOptBool(true)
	}

	resp, err := s.client.apiClient.ComposeDetailed(ctx,
		api.NewOptBodyComposeMusicWithADetailedResponseV1MusicDetailedPost(*body),
		api.ComposeDetailedParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.ComposeDetailedOKHeaders:
		return &MusicDetailedResponse{
			Audio:  r.Response.Data,
			SongID: r.SongID.Value,
		}, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// StemSeparationRequest contains options for stem separation.
type StemSeparationRequest struct {
	// File is the audio file to separate.
	File io.Reader

	// Filename is the name of the file.
	Filename string

	// StemVariation specifies which stem variation to use.
	// Options: "two_stems_v1" (vocals + music), "six_stems_v1" (vocals, drums, bass, other - default)
	StemVariation string
}

// SeparateStems separates a song into individual stems (vocals, instruments, etc.).
//
// Example:
//
//	f, _ := os.Open("song.mp3")
//	stems, err := client.Music().SeparateStems(ctx, &StemSeparationRequest{
//	    File:     f,
//	    Filename: "song.mp3",
//	})
//	// Save the separated stems (returned as a zip file)
//	output, _ := os.Create("stems.zip")
//	io.Copy(output, stems)
func (s *MusicService) SeparateStems(ctx context.Context, req *StemSeparationRequest) (io.Reader, error) {
	if req.File == nil {
		return nil, &ValidationError{Field: "file", Message: "cannot be nil"}
	}
	if req.Filename == "" {
		return nil, &ValidationError{Field: "filename", Message: "cannot be empty"}
	}

	body := &api.BodyStemSeparationV1MusicStemSeparationPostMultipart{
		File: ht.MultipartFile{
			Name: req.Filename,
			File: req.File,
		},
	}

	if req.StemVariation != "" {
		body.StemVariationID = api.NewOptBodyStemSeparationV1MusicStemSeparationPostMultipartStemVariationID(
			api.BodyStemSeparationV1MusicStemSeparationPostMultipartStemVariationID(req.StemVariation))
	}

	resp, err := s.client.apiClient.SeparateSongStems(ctx, body, api.SeparateSongStemsParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.SeparateSongStemsOKHeaders:
		return r.Response.Data, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// SeparateStemsFile is a convenience method to separate stems from a file path.
func (s *MusicService) SeparateStemsFile(ctx context.Context, filePath string) (io.Reader, error) {
	// Note: This would require os.Open which we avoid in the SDK
	// Users should open the file themselves and use SeparateStems
	return nil, &ValidationError{Field: "file_path", Message: "use SeparateStems with an opened file instead"}
}

// Helper functions to convert between SDK types and API types

func compositionPlanToAPI(plan *CompositionPlan) api.MusicPrompt {
	apiPlan := api.MusicPrompt{
		PositiveGlobalStyles: plan.PositiveGlobalStyles,
		NegativeGlobalStyles: plan.NegativeGlobalStyles,
	}

	for _, section := range plan.Sections {
		apiSection := api.SongSection{
			SectionName:         section.SectionName,
			DurationMs:          section.DurationMs,
			Lines:               section.Lines,
			PositiveLocalStyles: section.PositiveLocalStyles,
			NegativeLocalStyles: section.NegativeLocalStyles,
		}
		apiPlan.Sections = append(apiPlan.Sections, apiSection)
	}

	return apiPlan
}

func compositionPlanFromAPI(apiPlan *api.MusicPrompt) *CompositionPlan {
	plan := &CompositionPlan{
		PositiveGlobalStyles: apiPlan.PositiveGlobalStyles,
		NegativeGlobalStyles: apiPlan.NegativeGlobalStyles,
	}

	for _, apiSection := range apiPlan.Sections {
		section := SongSection{
			SectionName:         apiSection.SectionName,
			DurationMs:          apiSection.DurationMs,
			Lines:               apiSection.Lines,
			PositiveLocalStyles: apiSection.PositiveLocalStyles,
			NegativeLocalStyles: apiSection.NegativeLocalStyles,
		}
		plan.Sections = append(plan.Sections, section)
	}

	return plan
}
