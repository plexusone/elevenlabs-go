package elevenlabs

import (
	"errors"
	"testing"

	"github.com/plexusone/elevenlabs-go/ax"
)

func TestValidationError(t *testing.T) {
	err := &ValidationError{Field: "voice_id", Message: "cannot be empty"}
	expected := "elevenlabs: validation error for voice_id: cannot be empty"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %s, want %s", err.Error(), expected)
	}
}

func TestAPIError(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name:     "without detail",
			err:      &APIError{StatusCode: 401, Message: "Unauthorized"},
			expected: "elevenlabs: API error (status 401): Unauthorized",
		},
		{
			name:     "with detail",
			err:      &APIError{StatusCode: 400, Message: "Bad Request", Detail: "Invalid voice_id"},
			expected: "elevenlabs: API error (status 400): Bad Request - Invalid voice_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("APIError.Error() = %s, want %s", tt.err.Error(), tt.expected)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "404 error",
			err:      &APIError{StatusCode: 404, Message: "Not Found"},
			expected: true,
		},
		{
			name:     "401 error",
			err:      &APIError{StatusCode: 401, Message: "Unauthorized"},
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFoundError(tt.err); got != tt.expected {
				t.Errorf("IsNotFoundError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsUnauthorizedError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "401 error",
			err:      &APIError{StatusCode: 401, Message: "Unauthorized"},
			expected: true,
		},
		{
			name:     "404 error",
			err:      &APIError{StatusCode: 404, Message: "Not Found"},
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnauthorizedError(tt.err); got != tt.expected {
				t.Errorf("IsUnauthorizedError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "429 error",
			err:      &APIError{StatusCode: 429, Message: "Too Many Requests"},
			expected: true,
		},
		{
			name:     "401 error",
			err:      &APIError{StatusCode: 401, Message: "Unauthorized"},
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRateLimitError(tt.err); got != tt.expected {
				t.Errorf("IsRateLimitError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAPIError_AXErrorCode(t *testing.T) {
	tests := []struct {
		name         string
		err          *APIError
		expectedCode string
		expectedOk   bool
	}{
		{
			name:         "error code in message",
			err:          &APIError{StatusCode: 404, Message: "DOCUMENT_NOT_FOUND", Detail: "The document was not found"},
			expectedCode: ax.ErrDocumentNotFound,
			expectedOk:   true,
		},
		{
			name:         "error code in detail",
			err:          &APIError{StatusCode: 401, Message: "Unauthorized", Detail: "NOT_LOGGED_IN"},
			expectedCode: ax.ErrNotLoggedIn,
			expectedOk:   true,
		},
		{
			name:         "no error code",
			err:          &APIError{StatusCode: 500, Message: "Internal Server Error", Detail: "Something went wrong"},
			expectedCode: "",
			expectedOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := tt.err.AXErrorCode()
			if code != tt.expectedCode || ok != tt.expectedOk {
				t.Errorf("AXErrorCode() = (%q, %v), want (%q, %v)",
					code, ok, tt.expectedCode, tt.expectedOk)
			}
		})
	}
}

func TestAPIError_HasAXCode(t *testing.T) {
	err := &APIError{StatusCode: 404, Message: "DOCUMENT_NOT_FOUND", Detail: "The document was not found"}

	if !err.HasAXCode(ax.ErrDocumentNotFound) {
		t.Error("HasAXCode should return true for DOCUMENT_NOT_FOUND")
	}

	if err.HasAXCode(ax.ErrUserNotFound) {
		t.Error("HasAXCode should return false for USER_NOT_FOUND")
	}
}

func TestIsAXError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     string
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			code:     ax.ErrDocumentNotFound,
			expected: false,
		},
		{
			name:     "APIError with matching code",
			err:      &APIError{StatusCode: 404, Message: "DOCUMENT_NOT_FOUND"},
			code:     ax.ErrDocumentNotFound,
			expected: true,
		},
		{
			name:     "APIError with non-matching code",
			err:      &APIError{StatusCode: 404, Message: "DOCUMENT_NOT_FOUND"},
			code:     ax.ErrUserNotFound,
			expected: false,
		},
		{
			name:     "generic error with code in message",
			err:      errors.New("Error: USER_NOT_FOUND - user does not exist"),
			code:     ax.ErrUserNotFound,
			expected: true,
		},
		{
			name:     "generic error without code",
			err:      errors.New("network timeout"),
			code:     ax.ErrDocumentNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAXError(tt.err, tt.code); got != tt.expected {
				t.Errorf("IsAXError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetAXErrorCode(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode string
		expectedOk   bool
	}{
		{
			name:         "nil error",
			err:          nil,
			expectedCode: "",
			expectedOk:   false,
		},
		{
			name:         "APIError with code",
			err:          &APIError{StatusCode: 403, Message: "NEEDS_AUTHORIZATION"},
			expectedCode: ax.ErrNeedsAuthorization,
			expectedOk:   true,
		},
		{
			name:         "APIError without code",
			err:          &APIError{StatusCode: 500, Message: "Internal Server Error"},
			expectedCode: "",
			expectedOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := GetAXErrorCode(tt.err)
			if code != tt.expectedCode || ok != tt.expectedOk {
				t.Errorf("GetAXErrorCode() = (%q, %v), want (%q, %v)",
					code, ok, tt.expectedCode, tt.expectedOk)
			}
		})
	}
}
