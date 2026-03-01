package elevenlabs

import (
	"context"
	"io"
	"time"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// HistoryService handles history operations.
type HistoryService struct {
	client *Client
}

// HistoryItem represents a speech generation history item.
type HistoryItem struct {
	// HistoryItemID is the unique identifier.
	HistoryItemID string

	// VoiceID is the ID of the voice used.
	VoiceID string

	// VoiceName is the name of the voice used.
	VoiceName string

	// VoiceCategory is the category of the voice.
	VoiceCategory string

	// ModelID is the ID of the model used.
	ModelID string

	// Text is the text that was converted to speech.
	Text string

	// State is the state of the history item.
	State string

	// Source is the source of the generation.
	Source string

	// ContentType is the content type of the audio.
	ContentType string

	// CharactersUsed is the number of characters used.
	CharactersUsed int

	// CreatedAt is when the item was created.
	CreatedAt time.Time
}

// HistoryListResponse contains the list of history items and pagination info.
type HistoryListResponse struct {
	// Items is the list of history items.
	Items []*HistoryItem

	// HasMore indicates if there are more items to fetch.
	HasMore bool

	// LastHistoryItemID is the ID of the last item (for pagination).
	LastHistoryItemID string
}

// HistoryListOptions contains options for listing history items.
type HistoryListOptions struct {
	// PageSize is the number of items per page.
	PageSize int

	// StartAfterHistoryItemID is for pagination (fetch items after this ID).
	StartAfterHistoryItemID string

	// VoiceID filters by voice ID.
	VoiceID string
}

// List returns a list of speech history items.
func (s *HistoryService) List(ctx context.Context, opts *HistoryListOptions) (*HistoryListResponse, error) {
	params := api.GetSpeechHistoryParams{}

	if opts != nil {
		if opts.PageSize > 0 {
			params.PageSize = api.NewOptInt(opts.PageSize)
		}
		if opts.StartAfterHistoryItemID != "" {
			params.StartAfterHistoryItemID = api.NewOptNilString(opts.StartAfterHistoryItemID)
		}
		if opts.VoiceID != "" {
			params.VoiceID = api.NewOptNilString(opts.VoiceID)
		}
	}

	resp, err := s.client.apiClient.GetSpeechHistory(ctx, params)
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.GetSpeechHistoryResponseModel:
		result := &HistoryListResponse{
			HasMore: r.HasMore,
			Items:   make([]*HistoryItem, 0, len(r.History)),
		}

		if r.LastHistoryItemID.Set && !r.LastHistoryItemID.Null {
			result.LastHistoryItemID = r.LastHistoryItemID.Value
		}

		for _, h := range r.History {
			item := &HistoryItem{
				HistoryItemID:  h.HistoryItemID,
				State:          string(h.State),
				ContentType:    h.ContentType,
				CharactersUsed: h.CharacterCountChangeTo - h.CharacterCountChangeFrom,
				CreatedAt:      time.Unix(int64(h.DateUnix), 0),
			}

			if h.VoiceID.Set && !h.VoiceID.Null {
				item.VoiceID = h.VoiceID.Value
			}
			if h.VoiceName.Set && !h.VoiceName.Null {
				item.VoiceName = h.VoiceName.Value
			}
			if h.VoiceCategory.Set && !h.VoiceCategory.Null {
				item.VoiceCategory = string(h.VoiceCategory.Value)
			}
			if h.ModelID.Set && !h.ModelID.Null {
				item.ModelID = h.ModelID.Value
			}
			if h.Text.Set && !h.Text.Null {
				item.Text = h.Text.Value
			}
			if h.Source.Set && !h.Source.Null {
				item.Source = string(h.Source.Value)
			}

			result.Items = append(result.Items, item)
		}

		return result, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// Get returns a specific history item by ID.
func (s *HistoryService) Get(ctx context.Context, historyItemID string) (*HistoryItem, error) {
	if historyItemID == "" {
		return nil, &ValidationError{Field: "history_item_id", Message: "cannot be empty"}
	}

	resp, err := s.client.apiClient.GetSpeechHistoryItemByID(ctx, api.GetSpeechHistoryItemByIDParams{
		HistoryItemID: historyItemID,
	})
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.SpeechHistoryItemResponseModel:
		item := &HistoryItem{
			HistoryItemID:  r.HistoryItemID,
			State:          string(r.State),
			ContentType:    r.ContentType,
			CharactersUsed: r.CharacterCountChangeTo - r.CharacterCountChangeFrom,
			CreatedAt:      time.Unix(int64(r.DateUnix), 0),
		}

		if r.VoiceID.Set && !r.VoiceID.Null {
			item.VoiceID = r.VoiceID.Value
		}
		if r.VoiceName.Set && !r.VoiceName.Null {
			item.VoiceName = r.VoiceName.Value
		}
		if r.VoiceCategory.Set && !r.VoiceCategory.Null {
			item.VoiceCategory = string(r.VoiceCategory.Value)
		}
		if r.ModelID.Set && !r.ModelID.Null {
			item.ModelID = r.ModelID.Value
		}
		if r.Text.Set && !r.Text.Null {
			item.Text = r.Text.Value
		}
		if r.Source.Set && !r.Source.Null {
			item.Source = string(r.Source.Value)
		}

		return item, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// GetAudio returns the audio for a history item.
func (s *HistoryService) GetAudio(ctx context.Context, historyItemID string) (io.Reader, error) {
	if historyItemID == "" {
		return nil, &ValidationError{Field: "history_item_id", Message: "cannot be empty"}
	}

	resp, err := s.client.apiClient.GetAudioFullFromSpeechHistoryItem(ctx, api.GetAudioFullFromSpeechHistoryItemParams{
		HistoryItemID: historyItemID,
	})
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.GetAudioFullFromSpeechHistoryItemOK:
		return r.Data, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// Delete deletes a history item by ID.
func (s *HistoryService) Delete(ctx context.Context, historyItemID string) error {
	if historyItemID == "" {
		return &ValidationError{Field: "history_item_id", Message: "cannot be empty"}
	}

	_, err := s.client.apiClient.DeleteSpeechHistoryItem(ctx, api.DeleteSpeechHistoryItemParams{
		HistoryItemID: historyItemID,
	})
	return err
}
