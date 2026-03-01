package elevenlabs

import (
	"context"
	"time"

	"github.com/plexusone/elevenlabs-go/internal/api"
)

// UserService handles user and subscription operations.
type UserService struct {
	client *Client
}

// User represents an ElevenLabs user.
type User struct {
	// UserID is the unique user identifier.
	UserID string

	// FirstName is the user's first name.
	FirstName string

	// Subscription contains the user's subscription details.
	Subscription *Subscription

	// CreatedAt is when the user was created.
	CreatedAt time.Time
}

// Subscription represents a user's subscription details.
type Subscription struct {
	// Tier is the subscription tier (e.g., "free", "starter", "creator").
	Tier string

	// Status is the subscription status.
	Status string

	// CharacterCount is the number of characters used.
	CharacterCount int

	// CharacterLimit is the maximum characters allowed.
	CharacterLimit int

	// VoiceLimit is the maximum number of voices allowed.
	VoiceLimit int

	// VoiceSlotsUsed is the number of voice slots used.
	VoiceSlotsUsed int

	// CanUseInstantVoiceCloning indicates if instant cloning is available.
	CanUseInstantVoiceCloning bool

	// CanUseProfessionalVoiceCloning indicates if pro cloning is available.
	CanUseProfessionalVoiceCloning bool

	// NextCharacterResetUnix is when characters reset (Unix timestamp).
	NextCharacterResetUnix int64
}

// CharactersRemaining returns the number of characters remaining.
func (s *Subscription) CharactersRemaining() int {
	remaining := s.CharacterLimit - s.CharacterCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetInfo returns the current user's information including subscription.
func (s *UserService) GetInfo(ctx context.Context) (*User, error) {
	resp, err := s.client.apiClient.GetUserInfo(ctx, api.GetUserInfoParams{})
	if err != nil {
		return nil, err
	}

	// Handle response type
	switch r := resp.(type) {
	case *api.UserResponseModel:
		user := &User{
			UserID:    r.UserID,
			CreatedAt: time.Unix(int64(r.CreatedAt), 0),
		}

		// Set first name if available
		if r.FirstName.Set && !r.FirstName.Null {
			user.FirstName = r.FirstName.Value
		}

		// Set subscription details
		sub := r.Subscription
		user.Subscription = &Subscription{
			Tier:                           sub.Tier,
			Status:                         string(sub.Status),
			CharacterCount:                 sub.CharacterCount,
			CharacterLimit:                 sub.CharacterLimit,
			VoiceLimit:                     sub.VoiceLimit,
			CanUseInstantVoiceCloning:      sub.CanUseInstantVoiceCloning,
			CanUseProfessionalVoiceCloning: sub.CanUseProfessionalVoiceCloning,
		}

		if sub.NextCharacterCountResetUnix.Set && !sub.NextCharacterCountResetUnix.Null {
			user.Subscription.NextCharacterResetUnix = int64(sub.NextCharacterCountResetUnix.Value)
		}

		return user, nil
	default:
		return nil, &APIError{Message: "unexpected response type"}
	}
}

// GetSubscription returns the current user's subscription details.
// This is a convenience method that calls GetInfo and returns just the subscription.
func (s *UserService) GetSubscription(ctx context.Context) (*Subscription, error) {
	user, err := s.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	return user.Subscription, nil
}

// GetCharactersRemaining returns the number of characters remaining in the current period.
func (s *UserService) GetCharactersRemaining(ctx context.Context) (int, error) {
	sub, err := s.GetSubscription(ctx)
	if err != nil {
		return 0, err
	}
	return sub.CharactersRemaining(), nil
}
