package elevenlabs

import (
	"context"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// VoicesService handles voice operations.
type VoicesService struct {
	client *Client
}

// Voice represents an ElevenLabs voice.
type Voice struct {
	// VoiceID is the unique identifier for the voice.
	VoiceID string

	// Name is the display name of the voice.
	Name string

	// Category is the category of the voice (e.g., "premade", "cloned").
	Category string

	// Description is the description of the voice.
	Description string

	// PreviewURL is the URL to preview the voice.
	PreviewURL string

	// Labels contains additional metadata about the voice.
	Labels map[string]string
}

// List returns all available voices.
func (s *VoicesService) List(ctx context.Context) ([]*Voice, error) {
	resp, err := s.client.apiClient.GetVoices(ctx, api.GetVoicesParams{})
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.GetVoicesResponseModel:
		voices := make([]*Voice, 0, len(r.Voices))
		for _, v := range r.Voices {
			voice := &Voice{
				VoiceID:  v.VoiceID,
				Name:     v.Name,
				Category: string(v.Category),
				Labels:   make(map[string]string),
			}
			if v.Description.Set && !v.Description.Null {
				voice.Description = v.Description.Value
			}
			if v.PreviewURL.Set && !v.PreviewURL.Null {
				voice.PreviewURL = v.PreviewURL.Value
			}
			// Convert labels
			for k, val := range v.Labels {
				voice.Labels[k] = val
			}
			voices = append(voices, voice)
		}
		return voices, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// Get returns a voice by ID.
func (s *VoicesService) Get(ctx context.Context, voiceID string) (*Voice, error) {
	if voiceID == "" {
		return nil, ErrEmptyVoiceID
	}

	resp, err := s.client.apiClient.GetVoiceByID(ctx, api.GetVoiceByIDParams{
		VoiceID: voiceID,
	})
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.VoiceResponseModel:
		voice := &Voice{
			VoiceID:  r.VoiceID,
			Name:     r.Name,
			Category: string(r.Category),
			Labels:   make(map[string]string),
		}
		if r.Description.Set && !r.Description.Null {
			voice.Description = r.Description.Value
		}
		if r.PreviewURL.Set && !r.PreviewURL.Null {
			voice.PreviewURL = r.PreviewURL.Value
		}
		// Convert labels
		for k, val := range r.Labels {
			voice.Labels[k] = val
		}
		return voice, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// GetSettings returns the settings for a voice.
func (s *VoicesService) GetSettings(ctx context.Context, voiceID string) (*VoiceSettings, error) {
	if voiceID == "" {
		return nil, ErrEmptyVoiceID
	}

	resp, err := s.client.apiClient.GetVoiceSettings(ctx, api.GetVoiceSettingsParams{
		VoiceID: voiceID,
	})
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.VoiceSettingsResponseModel:
		settings := &VoiceSettings{}
		if r.Stability.Set && !r.Stability.Null {
			settings.Stability = r.Stability.Value
		}
		if r.SimilarityBoost.Set && !r.SimilarityBoost.Null {
			settings.SimilarityBoost = r.SimilarityBoost.Value
		}
		if r.Style.Set && !r.Style.Null {
			settings.Style = r.Style.Value
		}
		if r.Speed.Set && !r.Speed.Null {
			settings.Speed = r.Speed.Value
		}
		return settings, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// GetDefaultSettings returns the default voice settings.
func (s *VoicesService) GetDefaultSettings(ctx context.Context) (*VoiceSettings, error) {
	resp, err := s.client.apiClient.GetVoiceSettingsDefault(ctx)
	if err != nil {
		return nil, err
	}

	settings := &VoiceSettings{}
	if resp.Stability.Set && !resp.Stability.Null {
		settings.Stability = resp.Stability.Value
	}
	if resp.SimilarityBoost.Set && !resp.SimilarityBoost.Null {
		settings.SimilarityBoost = resp.SimilarityBoost.Value
	}
	if resp.Style.Set && !resp.Style.Null {
		settings.Style = resp.Style.Value
	}
	if resp.Speed.Set && !resp.Speed.Null {
		settings.Speed = resp.Speed.Value
	}
	return settings, nil
}

// Delete deletes a voice by ID.
func (s *VoicesService) Delete(ctx context.Context, voiceID string) error {
	if voiceID == "" {
		return ErrEmptyVoiceID
	}

	_, err := s.client.apiClient.DeleteVoice(ctx, api.DeleteVoiceParams{
		VoiceID: voiceID,
	})
	return err
}
