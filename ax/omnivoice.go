// Package ax provides Agent Experience (AX) helpers for ElevenLabs API integration.
//
// This file contains helper functions specifically designed for OmniVoice integration,
// bridging between ElevenLabs AX error codes and the omnivoice-core resilience package.
package ax

import (
	"github.com/plexusone/omnivoice-core/resilience"
)

// CategoryForCode returns the resilience.ErrorCategory for an AX error code.
// Returns CategoryUnknown if the code is not recognized.
//
// Example:
//
//	if code, ok := ax.ContainsErrorCode(err); ok {
//	    category := ax.CategoryForCode(code)
//	    if category == resilience.CategoryAuth {
//	        // Handle auth error
//	    }
//	}
func CategoryForCode(code string) resilience.ErrorCategory {
	info := GetErrorInfo(code)
	if info == nil {
		return resilience.CategoryUnknown
	}

	switch info.Category {
	case "auth":
		return resilience.CategoryAuth
	case "not_found":
		return resilience.CategoryNotFound
	case "validation":
		return resilience.CategoryValidation
	case "rate_limit":
		return resilience.CategoryRateLimit
	case "quota":
		return resilience.CategoryQuota
	case "server":
		return resilience.CategoryServer
	case "transient":
		return resilience.CategoryTransient
	default:
		return resilience.CategoryUnknown
	}
}

// IsRetryableCode returns true if the AX error code represents a retryable error.
// Returns false for unknown codes.
//
// Example:
//
//	if code, ok := ax.ContainsErrorCode(err); ok {
//	    if ax.IsRetryableCode(code) {
//	        // Safe to retry
//	    }
//	}
func IsRetryableCode(code string) bool {
	info := GetErrorInfo(code)
	if info == nil {
		return false
	}
	return info.Retryable
}

// SuggestionForCode returns a helpful suggestion for resolving an AX error code.
// Returns a generic suggestion for unknown codes.
//
// Example:
//
//	if code, ok := ax.ContainsErrorCode(err); ok {
//	    suggestion := ax.SuggestionForCode(code)
//	    log.Printf("Error: %s. Suggestion: %s", code, suggestion)
//	}
func SuggestionForCode(code string) string {
	switch code {
	case ErrDocumentNotFound:
		return "Verify the document ID exists and you have access to it"
	case ErrInvalidUID:
		return "Check that the provided ID is a valid format"
	case ErrMissingFeedback:
		return "Include required feedback in the request"
	case ErrNeedsAuthorization:
		return "This operation requires additional authorization; check your subscription tier"
	case ErrNotLoggedIn:
		return "Provide a valid API key in the request"
	case ErrNoEditChanges:
		return "Include at least one change in the edit request"
	case ErrUnprocessableEntity:
		return "Check the request body for validation errors"
	case ErrUserNotFound:
		return "Verify the user ID exists"
	case ErrWorkspaceNotFound:
		return "Verify the workspace ID exists and you have access to it"
	default:
		return "Check the error details and ElevenLabs API documentation"
	}
}

// Operation represents an ElevenLabs API operation.
type Operation string

// Known ElevenLabs API operations.
const (
	OpTextToSpeech     Operation = "text-to-speech"
	OpSpeechToText     Operation = "speech-to-text"
	OpVoiceClone       Operation = "voice-clone"
	OpVoiceDesign      Operation = "voice-design"
	OpSoundGeneration  Operation = "sound-generation"
	OpAudioIsolation   Operation = "audio-isolation"
	OpDubbing          Operation = "dubbing"
	OpProjects         Operation = "projects"
	OpPronunciation    Operation = "pronunciation"
	OpConversationalAI Operation = "conversational-ai"
)

// RequiredField represents a required field for an operation.
type RequiredField struct {
	Name        string
	Description string
	Example     string
}

// OperationRequiredFields returns the required fields for an ElevenLabs API operation.
// This enables pre-flight validation before making API calls.
//
// Example:
//
//	fields := ax.OperationRequiredFields(ax.OpTextToSpeech)
//	for _, f := range fields {
//	    if isEmpty(request[f.Name]) {
//	        return fmt.Errorf("missing required field: %s", f.Name)
//	    }
//	}
func OperationRequiredFields(op Operation) []RequiredField {
	switch op {
	case OpTextToSpeech:
		return []RequiredField{
			{Name: "voice_id", Description: "The voice ID to use for synthesis", Example: "21m00Tcm4TlvDq8ikWAM"},
			{Name: "text", Description: "The text to convert to speech", Example: "Hello world"},
		}
	case OpSpeechToText:
		return []RequiredField{
			{Name: "audio", Description: "Audio data to transcribe", Example: "(binary audio data)"},
		}
	case OpVoiceClone:
		return []RequiredField{
			{Name: "name", Description: "Name for the cloned voice", Example: "My Voice"},
			{Name: "files", Description: "Audio samples for voice cloning", Example: "(audio files)"},
		}
	case OpVoiceDesign:
		return []RequiredField{
			{Name: "text", Description: "Text to generate voice preview", Example: "Hello, this is a test."},
			{Name: "gender", Description: "Voice gender", Example: "female"},
			{Name: "age", Description: "Voice age category", Example: "young"},
			{Name: "accent", Description: "Voice accent", Example: "american"},
		}
	case OpSoundGeneration:
		return []RequiredField{
			{Name: "text", Description: "Description of sound to generate", Example: "thunder rolling in the distance"},
		}
	case OpAudioIsolation:
		return []RequiredField{
			{Name: "audio", Description: "Audio data to isolate vocals from", Example: "(binary audio data)"},
		}
	case OpDubbing:
		return []RequiredField{
			{Name: "source_url", Description: "URL of video/audio to dub", Example: "https://example.com/video.mp4"},
			{Name: "target_lang", Description: "Target language code", Example: "es"},
		}
	case OpProjects:
		return []RequiredField{
			{Name: "name", Description: "Project name", Example: "My Audiobook"},
		}
	case OpPronunciation:
		return []RequiredField{
			{Name: "name", Description: "Pronunciation dictionary name", Example: "Technical Terms"},
		}
	case OpConversationalAI:
		return []RequiredField{
			{Name: "agent_id", Description: "Conversational AI agent ID", Example: "agent_abc123"},
		}
	default:
		return nil
	}
}

// ToErrorInfo converts an AX error code to a resilience.ErrorInfo.
// This is useful for wrapping ElevenLabs errors as ProviderErrors.
//
// Example:
//
//	if code, ok := ax.ContainsErrorCode(err); ok {
//	    info := ax.ToErrorInfo(code)
//	    return resilience.NewProviderError("elevenlabs", "Synthesize", err, info)
//	}
func ToErrorInfo(code string) resilience.ErrorInfo {
	axInfo := GetErrorInfo(code)
	if axInfo == nil {
		return resilience.ErrorInfo{
			Category:   resilience.CategoryUnknown,
			Retryable:  false,
			Code:       code,
			Message:    "Unknown error code",
			Suggestion: SuggestionForCode(code),
		}
	}

	return resilience.ErrorInfo{
		Category:   CategoryForCode(code),
		Retryable:  axInfo.Retryable,
		Code:       code,
		Message:    axInfo.Description,
		Suggestion: SuggestionForCode(code),
	}
}

// ClassifyHTTPStatus returns a resilience.ErrorInfo for an HTTP status code.
// This is useful when no AX error code is available in the response.
func ClassifyHTTPStatus(status int, message string) resilience.ErrorInfo {
	classifier := &resilience.HTTPStatusClassifier{}
	return classifier.ClassifyStatus(status, message)
}

// AllCategories returns all error categories used by ElevenLabs AX errors.
func AllCategories() []string {
	seen := make(map[string]bool)
	var categories []string

	for _, info := range ErrorCodeMetadata {
		if !seen[info.Category] {
			seen[info.Category] = true
			categories = append(categories, info.Category)
		}
	}

	return categories
}

// CodesByCategory returns all AX error codes for a given category.
//
// Example:
//
//	authCodes := ax.CodesByCategory("auth")
//	// Returns: ["NEEDS_AUTHORIZATION", "NOT_LOGGED_IN"]
func CodesByCategory(category string) []string {
	var codes []string
	for code, info := range ErrorCodeMetadata {
		if info.Category == category {
			codes = append(codes, code)
		}
	}
	return codes
}
