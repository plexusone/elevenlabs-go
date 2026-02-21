package elevenlabs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// SpeechToSpeechService handles voice conversion operations.
type SpeechToSpeechService struct {
	client *Client
}

// SpeechToSpeechRequest is a request to convert speech to a different voice.
type SpeechToSpeechRequest struct {
	// VoiceID is the target voice to convert to.
	VoiceID string

	// Audio is the source audio data to convert.
	Audio io.Reader

	// AudioFilename is the filename for the audio (optional, helps with format detection).
	AudioFilename string

	// ModelID is the model to use. Defaults to "eleven_english_sts_v2".
	ModelID string

	// VoiceSettings configures the voice parameters.
	VoiceSettings *VoiceSettings

	// OutputFormat specifies the audio output format.
	// Examples: "mp3_44100_128", "pcm_16000", "pcm_22050"
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
		return ErrEmptyVoiceID
	}
	if r.Audio == nil {
		return &APIError{Message: "audio is required"}
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

// Convert converts speech from one voice to another.
func (s *SpeechToSpeechService) Convert(ctx context.Context, req *SpeechToSpeechRequest) (*SpeechToSpeechResponse, error) {
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
		return nil, fmt.Errorf("failed to create audio form field: %w", err)
	}
	if _, err := io.Copy(audioWriter, req.Audio); err != nil {
		return nil, fmt.Errorf("failed to write audio: %w", err)
	}

	// Add model ID
	modelID := req.ModelID
	if modelID == "" {
		modelID = "eleven_english_sts_v2"
	}
	if err := writer.WriteField("model_id", modelID); err != nil {
		return nil, fmt.Errorf("failed to write model_id: %w", err)
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
			return nil, fmt.Errorf("failed to create seed_audio form field: %w", err)
		}
		if _, err := io.Copy(seedWriter, req.SeedAudio); err != nil {
			return nil, fmt.Errorf("failed to write seed audio: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Build URL
	url := fmt.Sprintf("%s/v1/speech-to-speech/%s", s.client.baseURL, req.VoiceID)
	if req.OutputFormat != "" {
		url += "?output_format=" + req.OutputFormat
	}

	// Make request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("xi-api-key", s.client.apiKey)

	resp, err := http.DefaultClient.Do(httpReq) //nolint:gosec // G704: API client, URL is fixed ElevenLabs endpoint
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return &SpeechToSpeechResponse{Audio: resp.Body}, nil
}

// ConvertStream converts speech with streaming response.
func (s *SpeechToSpeechService) ConvertStream(ctx context.Context, req *SpeechToSpeechRequest) (*SpeechToSpeechResponse, error) {
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
		return nil, fmt.Errorf("failed to create audio form field: %w", err)
	}
	if _, err := io.Copy(audioWriter, req.Audio); err != nil {
		return nil, fmt.Errorf("failed to write audio: %w", err)
	}

	// Add model ID
	modelID := req.ModelID
	if modelID == "" {
		modelID = "eleven_english_sts_v2"
	}
	if err := writer.WriteField("model_id", modelID); err != nil {
		return nil, fmt.Errorf("failed to write model_id: %w", err)
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
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Build URL for streaming endpoint
	url := fmt.Sprintf("%s/v1/speech-to-speech/%s/stream", s.client.baseURL, req.VoiceID)
	if req.OutputFormat != "" {
		url += "?output_format=" + req.OutputFormat
	}

	// Make request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("xi-api-key", s.client.apiKey)

	resp, err := http.DefaultClient.Do(httpReq) //nolint:gosec // G704: API client, URL is fixed ElevenLabs endpoint
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return &SpeechToSpeechResponse{Audio: resp.Body}, nil
}

// Simple is a convenience method for basic voice conversion.
func (s *SpeechToSpeechService) Simple(ctx context.Context, voiceID string, audio io.Reader) (io.Reader, error) {
	resp, err := s.Convert(ctx, &SpeechToSpeechRequest{
		VoiceID:       voiceID,
		Audio:         audio,
		VoiceSettings: DefaultVoiceSettings(),
	})
	if err != nil {
		return nil, err
	}
	return resp.Audio, nil
}
