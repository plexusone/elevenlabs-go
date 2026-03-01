package elevenlabs

import (
	"context"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// ModelsService handles model operations.
type ModelsService struct {
	client *Client
}

// Language represents a language supported by a model.
type Language struct {
	// LanguageID is the unique identifier (ISO code).
	LanguageID string

	// Name is the display name of the language.
	Name string
}

// Model represents an ElevenLabs model.
type Model struct {
	// ModelID is the unique identifier for the model.
	ModelID string

	// Name is the display name of the model.
	Name string

	// Description is the model description.
	Description string

	// CanDoTextToSpeech indicates if the model supports TTS.
	CanDoTextToSpeech bool

	// CanDoVoiceConversion indicates if the model supports voice conversion.
	CanDoVoiceConversion bool

	// CanBeFinetuned indicates if the model can be fine-tuned.
	CanBeFinetuned bool

	// CanUseStyle indicates if the model supports style settings.
	CanUseStyle bool

	// CanUseSpeakerBoost indicates if the model supports speaker boost.
	CanUseSpeakerBoost bool

	// Languages is the list of supported languages.
	Languages []*Language

	// MaxCharactersFreeUser is the max characters for free users.
	MaxCharactersFreeUser int

	// MaxCharactersSubscribedUser is the max characters for subscribed users.
	MaxCharactersSubscribedUser int

	// TokenCostFactor is the cost factor for the model.
	TokenCostFactor float64
}

// List returns all available models.
func (s *ModelsService) List(ctx context.Context) ([]*Model, error) {
	resp, err := s.client.apiClient.GetModels(ctx, api.GetModelsParams{})
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.GetModelsOKApplicationJSON:
		models := make([]*Model, 0, len(*r))
		for _, m := range *r {
			model := &Model{
				ModelID:                     m.ModelID,
				Name:                        m.Name,
				Description:                 m.Description,
				CanDoTextToSpeech:           m.CanDoTextToSpeech,
				CanDoVoiceConversion:        m.CanDoVoiceConversion,
				CanBeFinetuned:              m.CanBeFinetuned,
				CanUseStyle:                 m.CanUseStyle,
				CanUseSpeakerBoost:          m.CanUseSpeakerBoost,
				MaxCharactersFreeUser:       m.MaxCharactersRequestFreeUser,
				MaxCharactersSubscribedUser: m.MaxCharactersRequestSubscribedUser,
				TokenCostFactor:             m.TokenCostFactor,
				Languages:                   make([]*Language, 0, len(m.Languages)),
			}
			for _, lang := range m.Languages {
				model.Languages = append(model.Languages, &Language{
					LanguageID: lang.LanguageID,
					Name:       lang.Name,
				})
			}
			models = append(models, model)
		}
		return models, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// ListTTSModels returns only models that support text-to-speech.
func (s *ModelsService) ListTTSModels(ctx context.Context) ([]*Model, error) {
	models, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	ttsModels := make([]*Model, 0)
	for _, m := range models {
		if m.CanDoTextToSpeech {
			ttsModels = append(ttsModels, m)
		}
	}
	return ttsModels, nil
}
