package ax

import (
	"errors"
	"testing"
)

func TestIsErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     string
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			code:     ErrDocumentNotFound,
			expected: false,
		},
		{
			name:     "matching error code",
			err:      errors.New("API error: DOCUMENT_NOT_FOUND - The document was not found"),
			code:     ErrDocumentNotFound,
			expected: true,
		},
		{
			name:     "non-matching error code",
			err:      errors.New("API error: DOCUMENT_NOT_FOUND"),
			code:     ErrUserNotFound,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("network timeout"),
			code:     ErrDocumentNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsErrorCode(tt.err, tt.code)
			if result != tt.expected {
				t.Errorf("IsErrorCode(%v, %q) = %v, want %v", tt.err, tt.code, result, tt.expected)
			}
		})
	}
}

func TestContainsErrorCode(t *testing.T) {
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
			name:         "contains DOCUMENT_NOT_FOUND",
			err:          errors.New("Error: DOCUMENT_NOT_FOUND"),
			expectedCode: ErrDocumentNotFound,
			expectedOk:   true,
		},
		{
			name:         "contains USER_NOT_FOUND",
			err:          errors.New("Error: USER_NOT_FOUND - user does not exist"),
			expectedCode: ErrUserNotFound,
			expectedOk:   true,
		},
		{
			name:         "no known error code",
			err:          errors.New("unknown error"),
			expectedCode: "",
			expectedOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := ContainsErrorCode(tt.err)
			if code != tt.expectedCode || ok != tt.expectedOk {
				t.Errorf("ContainsErrorCode(%v) = (%q, %v), want (%q, %v)",
					tt.err, code, ok, tt.expectedCode, tt.expectedOk)
			}
		})
	}
}

func TestGetErrorInfo(t *testing.T) {
	// Test known error code
	info := GetErrorInfo(ErrDocumentNotFound)
	if info == nil {
		t.Fatal("GetErrorInfo(ErrDocumentNotFound) returned nil")
	}
	if info.Category != "not_found" {
		t.Errorf("expected category 'not_found', got %q", info.Category)
	}

	// Test unknown error code
	info = GetErrorInfo("UNKNOWN_ERROR")
	if info != nil {
		t.Errorf("GetErrorInfo(UNKNOWN_ERROR) should return nil, got %+v", info)
	}
}

func TestErrorCategoryHelpers(t *testing.T) {
	tests := []struct {
		code         string
		isAuth       bool
		isNotFound   bool
		isValidation bool
	}{
		{ErrNotLoggedIn, true, false, false},
		{ErrNeedsAuthorization, true, false, false},
		{ErrDocumentNotFound, false, true, false},
		{ErrUserNotFound, false, true, false},
		{ErrInvalidUID, false, false, true},
		{ErrUnprocessableEntity, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if IsAuthError(tt.code) != tt.isAuth {
				t.Errorf("IsAuthError(%q) = %v, want %v", tt.code, !tt.isAuth, tt.isAuth)
			}
			if IsNotFoundError(tt.code) != tt.isNotFound {
				t.Errorf("IsNotFoundError(%q) = %v, want %v", tt.code, !tt.isNotFound, tt.isNotFound)
			}
			if IsValidationError(tt.code) != tt.isValidation {
				t.Errorf("IsValidationError(%q) = %v, want %v", tt.code, !tt.isValidation, tt.isValidation)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		operationID string
		expected    bool
	}{
		{"get_voices", true},
		{"get_models", true},
		{"get_user_info", true},
		{"text_to_speech_full", false},
		{"create_voice", false},
		{"delete_voice", false},
		{"unknown_operation", false},
	}

	for _, tt := range tests {
		t.Run(tt.operationID, func(t *testing.T) {
			result := IsRetryable(tt.operationID)
			if result != tt.expected {
				t.Errorf("IsRetryable(%q) = %v, want %v", tt.operationID, result, tt.expected)
			}
		})
	}
}

func TestGetRequiredFields(t *testing.T) {
	// Test operation with required fields
	fields := GetRequiredFields("text_to_speech_full")
	if len(fields) == 0 {
		t.Error("text_to_speech_full should have required fields")
	}
	found := false
	for _, f := range fields {
		if f == "text" {
			found = true
			break
		}
	}
	if !found {
		t.Error("text_to_speech_full should require 'text' field")
	}

	// Test operation without required fields (or unknown)
	fields = GetRequiredFields("get_voices")
	if len(fields) != 0 {
		t.Errorf("get_voices should have no required fields, got %v", fields)
	}
}

func TestMissingFields(t *testing.T) {
	present := map[string]bool{
		"text": true,
	}

	// All required fields present
	missing := MissingFields("text_to_speech_full", present)
	if len(missing) != 0 {
		t.Errorf("expected no missing fields, got %v", missing)
	}

	// Missing required field
	missing = MissingFields("create_batch_call", present)
	if len(missing) == 0 {
		t.Error("expected missing fields for create_batch_call")
	}
}

func TestValidateFields(t *testing.T) {
	present := map[string]bool{
		"text": true,
	}

	// Valid
	msg := ValidateFields("text_to_speech_full", present)
	if msg != "" {
		t.Errorf("expected empty validation message, got %q", msg)
	}

	// Invalid - missing fields
	msg = ValidateFields("create_batch_call", present)
	if msg == "" {
		t.Error("expected validation error message for create_batch_call")
	}
}

func TestCapabilities(t *testing.T) {
	// Test read operation
	caps := GetCapabilities("get_voices")
	if len(caps) == 0 {
		t.Fatal("get_voices should have capabilities")
	}
	if !HasCapability("get_voices", CapRead) {
		t.Error("get_voices should have read capability")
	}
	if HasCapability("get_voices", CapWrite) {
		t.Error("get_voices should not have write capability")
	}

	// Test write operation
	if !HasCapability("create_agent_route", CapWrite) {
		t.Error("create_agent_route should have write capability")
	}

	// Test delete operation
	if !HasCapability("delete_voice", CapDelete) {
		t.Error("delete_voice should have delete capability")
	}

	// Test admin operation
	if !RequiresAdmin("invite_user") {
		t.Error("invite_user should require admin")
	}
}

func TestIsReadOnly(t *testing.T) {
	if !IsReadOnly("get_voices") {
		t.Error("get_voices should be read-only")
	}
	if IsReadOnly("create_agent_route") {
		t.Error("create_agent_route should not be read-only")
	}
	if IsReadOnly("delete_voice") {
		t.Error("delete_voice should not be read-only")
	}
}
