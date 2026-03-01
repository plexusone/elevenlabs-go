package elevenlabs

import (
	"context"
	"io"

	ht "github.com/ogen-go/ogen/http"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// ForcedAlignmentService handles forced alignment between audio and text.
type ForcedAlignmentService struct {
	client *Client
}

// ForcedAlignmentRequest contains options for forced alignment.
type ForcedAlignmentRequest struct {
	// File is the audio file to align (required).
	File io.Reader

	// Filename is the name of the file (required).
	Filename string

	// Text is the text to align with the audio (required).
	Text string
}

// ForcedAlignmentResponse contains the alignment result.
type ForcedAlignmentResponse struct {
	// Words contains word-level timing information.
	Words []AlignmentWord

	// Characters contains character-level timing information.
	Characters []AlignmentCharacter

	// Loss is the average alignment confidence score.
	Loss float64
}

// AlignmentWord represents a word with timing information.
type AlignmentWord struct {
	// Text is the word text.
	Text string

	// Start is the start time in seconds.
	Start float64

	// End is the end time in seconds.
	End float64

	// Loss is the confidence score for this word.
	Loss float64
}

// AlignmentCharacter represents a character with timing information.
type AlignmentCharacter struct {
	// Text is the character text.
	Text string

	// Start is the start time in seconds.
	Start float64

	// End is the end time in seconds.
	End float64
}

// Align performs forced alignment between audio and text.
// This is useful for generating word-level timestamps for captions and subtitles.
func (s *ForcedAlignmentService) Align(ctx context.Context, req *ForcedAlignmentRequest) (*ForcedAlignmentResponse, error) {
	if req.File == nil {
		return nil, &ValidationError{Field: "file", Message: "cannot be nil"}
	}
	if req.Text == "" {
		return nil, &ValidationError{Field: "text", Message: "cannot be empty"}
	}

	body := &api.BodyCreateForcedAlignmentV1ForcedAlignmentPostMultipart{
		File: ht.MultipartFile{
			Name: req.Filename,
			File: req.File,
		},
		Text: req.Text,
	}

	resp, err := s.client.apiClient.ForcedAlignment(ctx, body, api.ForcedAlignmentParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.ForcedAlignmentResponseModel:
		result := &ForcedAlignmentResponse{
			Loss: r.Loss,
		}

		// Convert words
		for _, w := range r.Words {
			result.Words = append(result.Words, AlignmentWord{
				Text:  w.Text,
				Start: w.Start,
				End:   w.End,
				Loss:  w.Loss,
			})
		}

		// Convert characters
		for _, c := range r.Characters {
			result.Characters = append(result.Characters, AlignmentCharacter{
				Text:  c.Text,
				Start: c.Start,
				End:   c.End,
			})
		}

		return result, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// AlignFile is a convenience method to align audio from a file reader with text.
func (s *ForcedAlignmentService) AlignFile(ctx context.Context, file io.Reader, filename, text string) (*ForcedAlignmentResponse, error) {
	return s.Align(ctx, &ForcedAlignmentRequest{
		File:     file,
		Filename: filename,
		Text:     text,
	})
}
