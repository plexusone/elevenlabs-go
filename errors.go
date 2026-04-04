package elevenlabs

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/plexusone/elevenlabs-go/ax"
	"github.com/plexusone/ogen-tools/ogenerror"
)

// Common errors
var (
	// ErrNoAPIKey is returned when no API key is provided.
	ErrNoAPIKey = errors.New("elevenlabs: API key is required")

	// ErrEmptyText is returned when text is empty.
	ErrEmptyText = errors.New("elevenlabs: text cannot be empty")

	// ErrEmptyVoiceID is returned when voice ID is empty.
	ErrEmptyVoiceID = errors.New("elevenlabs: voice ID is required")

	// ErrInvalidStability is returned when stability is out of range.
	ErrInvalidStability = errors.New("elevenlabs: stability must be between 0.0 and 1.0")

	// ErrInvalidSimilarityBoost is returned when similarity_boost is out of range.
	ErrInvalidSimilarityBoost = errors.New("elevenlabs: similarity_boost must be between 0.0 and 1.0")

	// ErrInvalidStyle is returned when style is out of range.
	ErrInvalidStyle = errors.New("elevenlabs: style must be between 0.0 and 1.0")

	// ErrInvalidSpeed is returned when speed is out of range.
	ErrInvalidSpeed = errors.New("elevenlabs: speed must be between 0.25 and 4.0")
)

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("elevenlabs: validation error for %s: %s", e.Field, e.Message)
}

// APIError represents an error returned by the ElevenLabs API.
type APIError struct {
	StatusCode int
	Message    string
	Detail     string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("elevenlabs: API error (status %d): %s - %s", e.StatusCode, e.Message, e.Detail)
	}
	return fmt.Sprintf("elevenlabs: API error (status %d): %s", e.StatusCode, e.Message)
}

// AXErrorCode extracts the AX error code from the API error, if present.
// Returns the error code constant (e.g., ax.ErrDocumentNotFound) and true if found.
// Use this for machine-readable error handling:
//
//	if apiErr := ParseAPIError(err); apiErr != nil {
//	    if code, ok := apiErr.AXErrorCode(); ok {
//	        switch code {
//	        case ax.ErrDocumentNotFound:
//	            // Handle document not found
//	        case ax.ErrNeedsAuthorization:
//	            // Handle auth required
//	        }
//	    }
//	}
func (e *APIError) AXErrorCode() (string, bool) {
	// Check Message and Detail for known error codes
	for _, code := range ax.AllErrorCodes {
		if strings.Contains(e.Message, code) || strings.Contains(e.Detail, code) {
			return code, true
		}
	}
	return "", false
}

// HasAXCode checks if the API error contains a specific AX error code.
func (e *APIError) HasAXCode(code string) bool {
	axCode, ok := e.AXErrorCode()
	return ok && axCode == code
}

// IsNotFoundError returns true if the error is a 404 Not Found error.
func IsNotFoundError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsUnauthorizedError returns true if the error is a 401 Unauthorized error.
func IsUnauthorizedError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return false
}

// IsRateLimitError returns true if the error is a 429 Too Many Requests error.
func IsRateLimitError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 429
	}
	return false
}

// IsForbiddenError returns true if the error is a 403 Forbidden error.
func IsForbiddenError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 403
	}
	return false
}

// IsAXError checks if an error contains a specific AX error code.
// This works with any error type by first trying to parse it as an APIError.
//
// Usage:
//
//	if IsAXError(err, ax.ErrDocumentNotFound) {
//	    // Handle document not found
//	}
func IsAXError(err error, code string) bool {
	if err == nil {
		return false
	}

	// Try to get as APIError first for structured checking
	apiErr := ParseAPIError(err)
	if apiErr != nil {
		return apiErr.HasAXCode(code)
	}

	// Fall back to string matching
	return ax.IsErrorCode(err, code)
}

// GetAXErrorCode extracts the AX error code from any error.
// Returns the error code and true if found, empty string and false otherwise.
func GetAXErrorCode(err error) (string, bool) {
	if err == nil {
		return "", false
	}

	// Try structured extraction first
	apiErr := ParseAPIError(err)
	if apiErr != nil {
		return apiErr.AXErrorCode()
	}

	// Fall back to string matching
	return ax.ContainsErrorCode(err)
}

// ParseAPIError extracts API error details from an error returned by the SDK.
// It handles ogen's UnexpectedStatusCodeError and parses the response body
// to extract the ElevenLabs error message.
//
// Usage:
//
//	resp, err := client.TextToSpeech().Generate(ctx, req)
//	if err != nil {
//	    if apiErr := elevenlabs.ParseAPIError(err); apiErr != nil {
//	        fmt.Printf("Status: %d, Message: %s\n", apiErr.StatusCode, apiErr.Message)
//	    }
//	    log.Fatal(err)
//	}
func ParseAPIError(err error) *APIError {
	if err == nil {
		return nil
	}

	// Check if it's already an APIError
	var existing *APIError
	if errors.As(err, &existing) {
		return existing
	}

	// Use ogen-tools to extract status code and body
	status := ogenerror.Parse(err)
	if status == nil {
		return nil
	}

	apiErr := &APIError{
		StatusCode: status.StatusCode,
		Message:    fmt.Sprintf("HTTP %d", status.StatusCode),
	}

	// Parse ElevenLabs-specific error format
	if len(status.Body) > 0 {
		var errResp struct {
			Detail interface{} `json:"detail"`
		}
		if json.Unmarshal(status.Body, &errResp) == nil {
			switch d := errResp.Detail.(type) {
			case string:
				apiErr.Detail = d
			case map[string]interface{}:
				if msg, ok := d["message"].(string); ok {
					apiErr.Message = msg
				}
				if detail, ok := d["status"].(string); ok {
					apiErr.Detail = detail
				}
			}
		}
		// If parsing failed, use raw body as detail
		if apiErr.Detail == "" && apiErr.Message == fmt.Sprintf("HTTP %d", status.StatusCode) {
			apiErr.Detail = string(status.Body)
		}
	}

	return apiErr
}
