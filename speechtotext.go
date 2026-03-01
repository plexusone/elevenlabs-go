package elevenlabs

import (
	"context"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// SpeechToTextService handles speech-to-text transcription.
type SpeechToTextService struct {
	client *Client
}

// TranscriptionRequest contains options for transcription.
type TranscriptionRequest struct {
	// FileURL is the HTTPS URL of the file to transcribe.
	// Either FileURL or FileContent must be provided.
	FileURL string

	// FileContent is the base64-encoded file content.
	// Either FileURL or FileContent must be provided.
	FileContent string

	// LanguageCode is an ISO-639-1 or ISO-639-3 language code.
	// If not provided, language is auto-detected.
	LanguageCode string

	// Diarize enables speaker diarization (who said what).
	Diarize bool

	// NumSpeakers is the expected number of speakers (for diarization).
	NumSpeakers int

	// TagAudioEvents tags audio events like laughter, applause, etc.
	TagAudioEvents bool

	// ModelID is the transcription model to use (default: "scribe_v1").
	ModelID string
}

// TranscriptionResponse contains the transcription result.
type TranscriptionResponse struct {
	// Text is the full transcribed text.
	Text string

	// LanguageCode is the detected language.
	LanguageCode string

	// Words contains word-level details with timestamps.
	Words []TranscriptionWord

	// Utterances contains speaker-labeled segments (when diarization is enabled).
	Utterances []TranscriptionUtterance
}

// TranscriptionWord represents a single word with timing.
type TranscriptionWord struct {
	// Text is the word text.
	Text string

	// Start is the start time in seconds.
	Start float64

	// End is the end time in seconds.
	End float64

	// Confidence is the confidence score (0-1).
	Confidence float64

	// Speaker is the speaker ID (when diarization is enabled).
	Speaker string

	// Type is the word type (e.g., "word", "punctuation").
	Type string
}

// TranscriptionUtterance represents a speaker segment.
type TranscriptionUtterance struct {
	// Text is the utterance text.
	Text string

	// Start is the start time in seconds.
	Start float64

	// End is the end time in seconds.
	End float64

	// Speaker is the speaker ID.
	Speaker string
}

// Transcribe transcribes audio to text.
func (s *SpeechToTextService) Transcribe(ctx context.Context, req *TranscriptionRequest) (*TranscriptionResponse, error) {
	if req.FileURL == "" && req.FileContent == "" {
		return nil, &ValidationError{Field: "file", Message: "either file_url or file_content must be provided"}
	}

	body := &api.BodySpeechToTextV1SpeechToTextPostMultipart{}

	if req.FileURL != "" {
		body.CloudStorageURL = api.NewOptNilString(req.FileURL)
	}
	if req.FileContent != "" {
		body.File = api.NewOptNilString(req.FileContent)
	}
	if req.LanguageCode != "" {
		body.LanguageCode = api.NewOptNilString(req.LanguageCode)
	}
	if req.Diarize {
		body.Diarize = api.NewOptBool(true)
	}
	if req.NumSpeakers > 0 {
		body.NumSpeakers = api.NewOptNilInt(req.NumSpeakers)
	}
	if req.TagAudioEvents {
		body.TagAudioEvents = api.NewOptBool(true)
	}
	if req.ModelID != "" {
		body.ModelID = req.ModelID
	}

	resp, err := s.client.apiClient.SpeechToText(ctx, body, api.SpeechToTextParams{})
	if err != nil {
		return nil, err
	}

	switch r := resp.(type) {
	case *api.SpeechToTextOK:
		// SpeechToTextOK is a oneOf type, extract the chunk response
		if !r.IsSpeechToTextChunkResponseModel() {
			return nil, &APIError{Message: "unexpected response format"}
		}
		chunk := r.SpeechToTextChunkResponseModel

		result := &TranscriptionResponse{
			Text:         chunk.Text,
			LanguageCode: chunk.LanguageCode,
		}

		// Convert words
		for _, w := range chunk.Words {
			word := TranscriptionWord{
				Text: w.Text,
				Type: string(w.Type),
			}
			if w.Start.Set && !w.Start.Null {
				word.Start = w.Start.Value
			}
			if w.End.Set && !w.End.Null {
				word.End = w.End.Value
			}
			if w.SpeakerID.Set && !w.SpeakerID.Null {
				word.Speaker = w.SpeakerID.Value
			}
			result.Words = append(result.Words, word)
		}

		return result, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// TranscribeURL transcribes audio from a URL.
func (s *SpeechToTextService) TranscribeURL(ctx context.Context, url string) (*TranscriptionResponse, error) {
	return s.Transcribe(ctx, &TranscriptionRequest{FileURL: url})
}

// TranscribeWithDiarization transcribes audio with speaker identification.
func (s *SpeechToTextService) TranscribeWithDiarization(ctx context.Context, url string) (*TranscriptionResponse, error) {
	return s.Transcribe(ctx, &TranscriptionRequest{
		FileURL: url,
		Diarize: true,
	})
}
