package elevenlabs

import (
	"context"
	"io"
	"time"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// DubbingService handles dubbing operations.
type DubbingService struct {
	client *Client
}

// DubbingProject represents a dubbing project.
type DubbingProject struct {
	// DubbingID is the unique identifier.
	DubbingID string

	// Name is the project name.
	Name string

	// Status is the current status (dubbed, dubbing, failed, cloning).
	Status string

	// TargetLanguages are the target languages for dubbing.
	TargetLanguages []string

	// SourceLanguage is the source language.
	SourceLanguage string

	// Error contains any error message if the project failed.
	Error string

	// CreatedAt is when the project was created.
	CreatedAt time.Time
}

// DubbingResponse contains the result of creating a dubbing project.
type DubbingResponse struct {
	// DubbingID is the ID of the created project.
	DubbingID string

	// ExpectedDurationSeconds is the expected duration.
	ExpectedDurationSeconds float64
}

// DubbingRequest contains options for creating a dubbing project.
type DubbingRequest struct {
	// Name is the name of the dubbing project.
	Name string

	// SourceURL is the URL of the source media (alternative to file upload).
	SourceURL string

	// File is the source media file (alternative to SourceURL).
	File io.Reader

	// SourceLanguage is the source language code (ISO 639-1).
	SourceLanguage string

	// TargetLanguage is the target language code (ISO 639-1).
	TargetLanguage string

	// NumSpeakers is the number of speakers (0 for auto-detection).
	NumSpeakers int

	// Watermark enables watermark (for free tier).
	Watermark bool

	// StartTime is the start time in seconds for dubbing.
	StartTime int

	// EndTime is the end time in seconds for dubbing.
	EndTime int

	// HighestResolution requests highest resolution output.
	HighestResolution bool

	// DropBackgroundAudio removes background audio.
	DropBackgroundAudio bool
}

// CreateFromURL creates a dubbing project from a URL source.
func (s *DubbingService) CreateFromURL(ctx context.Context, req *DubbingRequest) (*DubbingResponse, error) {
	if req.SourceURL == "" {
		return nil, &ValidationError{Field: "source_url", Message: "cannot be empty"}
	}
	if req.TargetLanguage == "" {
		return nil, &ValidationError{Field: "target_language", Message: "cannot be empty"}
	}

	// Build request body
	body := api.BodyDubAVideoOrAnAudioFileV1DubbingPostMultipart{}
	body.SourceURL = api.NewOptNilString(req.SourceURL)
	body.TargetLang = api.NewOptNilString(req.TargetLanguage)

	if req.Name != "" {
		body.Name = api.NewOptNilString(req.Name)
	}
	if req.SourceLanguage != "" {
		body.SourceLang = api.NewOptString(req.SourceLanguage)
	}
	if req.NumSpeakers != 0 {
		body.NumSpeakers = api.NewOptInt(req.NumSpeakers)
	}
	if req.Watermark {
		body.Watermark = api.NewOptBool(true)
	}
	if req.StartTime > 0 {
		body.StartTime = api.NewOptNilInt(req.StartTime)
	}
	if req.EndTime > 0 {
		body.EndTime = api.NewOptNilInt(req.EndTime)
	}
	if req.HighestResolution {
		body.HighestResolution = api.NewOptBool(true)
	}
	if req.DropBackgroundAudio {
		body.DropBackgroundAudio = api.NewOptBool(true)
	}

	resp, err := s.client.apiClient.CreateDubbing(ctx, api.NewOptBodyDubAVideoOrAnAudioFileV1DubbingPostMultipart(body), api.CreateDubbingParams{})
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.DoDubbingResponseModel:
		return &DubbingResponse{
			DubbingID:               r.DubbingID,
			ExpectedDurationSeconds: r.ExpectedDurationSec,
		}, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// Get returns a dubbing project metadata by ID.
func (s *DubbingService) Get(ctx context.Context, dubbingID string) (*DubbingProject, error) {
	if dubbingID == "" {
		return nil, &ValidationError{Field: "dubbing_id", Message: "cannot be empty"}
	}

	resp, err := s.client.apiClient.GetDubbedMetadata(ctx, api.GetDubbedMetadataParams{
		DubbingID: dubbingID,
	})
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.DubbingMetadataResponse:
		project := &DubbingProject{
			DubbingID:       r.DubbingID,
			Name:            r.Name,
			Status:          r.Status,
			TargetLanguages: r.TargetLanguages,
			CreatedAt:       r.CreatedAt,
		}

		if r.Error.Set && !r.Error.Null {
			project.Error = r.Error.Value
		}

		return project, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// Delete deletes a dubbing project by ID.
func (s *DubbingService) Delete(ctx context.Context, dubbingID string) error {
	if dubbingID == "" {
		return &ValidationError{Field: "dubbing_id", Message: "cannot be empty"}
	}

	_, err := s.client.apiClient.DeleteDubbing(ctx, api.DeleteDubbingParams{
		DubbingID: dubbingID,
	})
	return err
}

// GetDubbedFile returns the dubbed audio/video file for a specific language.
func (s *DubbingService) GetDubbedFile(ctx context.Context, dubbingID, languageCode string) (io.Reader, error) {
	if dubbingID == "" {
		return nil, &ValidationError{Field: "dubbing_id", Message: "cannot be empty"}
	}
	if languageCode == "" {
		return nil, &ValidationError{Field: "language_code", Message: "cannot be empty"}
	}

	resp, err := s.client.apiClient.GetDubbedFile(ctx, api.GetDubbedFileParams{
		DubbingID:    dubbingID,
		LanguageCode: languageCode,
	})
	if err != nil {
		return nil, err
	}

	// Handle response type - can be audio or video
	switch r := resp.(type) {
	case *api.GetDubbedFileOKAudioMpeg:
		return r.Data, nil
	case *api.GetDubbedFileOKVideoMP4:
		return r.Data, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// IsComplete checks if a dubbing project is complete.
func (p *DubbingProject) IsComplete() bool {
	return p.Status == "dubbed"
}

// IsFailed checks if a dubbing project has failed.
func (p *DubbingProject) IsFailed() bool {
	return p.Status == "failed"
}

// IsProcessing checks if a dubbing project is still processing.
func (p *DubbingProject) IsProcessing() bool {
	return p.Status == "dubbing" || p.Status == "cloning"
}
