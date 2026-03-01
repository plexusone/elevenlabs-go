package elevenlabs

import (
	"context"
	"io"

	ht "github.com/ogen-go/ogen/http"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// AudioIsolationService handles audio isolation (vocal/speech extraction).
type AudioIsolationService struct {
	client *Client
}

// AudioIsolationRequest contains options for audio isolation.
type AudioIsolationRequest struct {
	// Audio is the audio file to process (required).
	Audio io.Reader

	// Filename is the name of the file (required).
	Filename string
}

// Isolate extracts vocals/speech from audio, removing background noise.
// Returns an io.Reader containing the isolated audio.
func (s *AudioIsolationService) Isolate(ctx context.Context, req *AudioIsolationRequest) (io.Reader, error) {
	if req.Audio == nil {
		return nil, &ValidationError{Field: "audio", Message: "cannot be nil"}
	}

	body := &api.BodyAudioIsolationV1AudioIsolationPostMultipart{
		Audio: ht.MultipartFile{
			Name: req.Filename,
			File: req.Audio,
		},
	}

	resp, err := s.client.apiClient.AudioIsolation(ctx, body, api.AudioIsolationParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.AudioIsolationOK:
		return r.Data, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// IsolateFile is a convenience method to isolate vocals from an audio file.
func (s *AudioIsolationService) IsolateFile(ctx context.Context, audio io.Reader, filename string) (io.Reader, error) {
	return s.Isolate(ctx, &AudioIsolationRequest{
		Audio:    audio,
		Filename: filename,
	})
}

// IsolateStream extracts vocals/speech from audio with streaming output.
// Returns an io.Reader for streaming the isolated audio.
func (s *AudioIsolationService) IsolateStream(ctx context.Context, req *AudioIsolationRequest) (io.Reader, error) {
	if req.Audio == nil {
		return nil, &ValidationError{Field: "audio", Message: "cannot be nil"}
	}

	body := &api.BodyAudioIsolationStreamV1AudioIsolationStreamPostMultipart{
		Audio: ht.MultipartFile{
			Name: req.Filename,
			File: req.Audio,
		},
	}

	resp, err := s.client.apiClient.AudioIsolationStream(ctx, body, api.AudioIsolationStreamParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.AudioIsolationStreamOK:
		return r.Data, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}
