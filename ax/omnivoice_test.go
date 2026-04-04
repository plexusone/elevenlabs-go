package ax

import (
	"fmt"
	"testing"

	"github.com/plexusone/omnivoice-core/resilience"
)

func TestCategoryForCode(t *testing.T) {
	tests := []struct {
		code string
		want resilience.ErrorCategory
	}{
		{ErrDocumentNotFound, resilience.CategoryNotFound},
		{ErrInvalidUID, resilience.CategoryValidation},
		{ErrMissingFeedback, resilience.CategoryValidation},
		{ErrNeedsAuthorization, resilience.CategoryAuth},
		{ErrNotLoggedIn, resilience.CategoryAuth},
		{ErrNoEditChanges, resilience.CategoryValidation},
		{ErrUnprocessableEntity, resilience.CategoryValidation},
		{ErrUserNotFound, resilience.CategoryNotFound},
		{ErrWorkspaceNotFound, resilience.CategoryNotFound},
		{"UNKNOWN_CODE", resilience.CategoryUnknown},
		{"", resilience.CategoryUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := CategoryForCode(tt.code)
			if got != tt.want {
				t.Errorf("CategoryForCode(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsRetryableCode(t *testing.T) {
	// All current ElevenLabs AX codes are non-retryable
	for _, code := range AllErrorCodes {
		t.Run(code, func(t *testing.T) {
			got := IsRetryableCode(code)
			info := GetErrorInfo(code)
			if info == nil {
				t.Fatalf("GetErrorInfo(%q) returned nil", code)
			}
			if got != info.Retryable {
				t.Errorf("IsRetryableCode(%q) = %v, want %v", code, got, info.Retryable)
			}
		})
	}

	// Unknown codes should not be retryable
	if IsRetryableCode("UNKNOWN") {
		t.Error("IsRetryableCode(UNKNOWN) should be false")
	}
}

func TestSuggestionForCode(t *testing.T) {
	// All known codes should have non-empty suggestions
	for _, code := range AllErrorCodes {
		t.Run(code, func(t *testing.T) {
			suggestion := SuggestionForCode(code)
			if suggestion == "" {
				t.Errorf("SuggestionForCode(%q) returned empty string", code)
			}
		})
	}

	// Unknown codes should have a generic suggestion
	suggestion := SuggestionForCode("UNKNOWN")
	if suggestion == "" {
		t.Error("SuggestionForCode(UNKNOWN) should return generic suggestion")
	}
}

func TestOperationRequiredFields(t *testing.T) {
	tests := []struct {
		op        Operation
		wantCount int
		wantField string
	}{
		{OpTextToSpeech, 2, "voice_id"},
		{OpSpeechToText, 1, "audio"},
		{OpVoiceClone, 2, "name"},
		{OpVoiceDesign, 4, "text"},
		{OpSoundGeneration, 1, "text"},
		{OpAudioIsolation, 1, "audio"},
		{OpDubbing, 2, "source_url"},
		{OpProjects, 1, "name"},
		{OpPronunciation, 1, "name"},
		{OpConversationalAI, 1, "agent_id"},
	}

	for _, tt := range tests {
		t.Run(string(tt.op), func(t *testing.T) {
			fields := OperationRequiredFields(tt.op)
			if len(fields) != tt.wantCount {
				t.Errorf("OperationRequiredFields(%q) returned %d fields, want %d", tt.op, len(fields), tt.wantCount)
			}
			if len(fields) > 0 && fields[0].Name != tt.wantField {
				t.Errorf("First field = %q, want %q", fields[0].Name, tt.wantField)
			}
		})
	}

	// Unknown operation should return nil
	if fields := OperationRequiredFields("unknown"); fields != nil {
		t.Errorf("OperationRequiredFields(unknown) = %v, want nil", fields)
	}
}

func TestOperationRequiredFields_FieldsHaveAllInfo(t *testing.T) {
	ops := []Operation{
		OpTextToSpeech, OpSpeechToText, OpVoiceClone, OpVoiceDesign,
		OpSoundGeneration, OpAudioIsolation, OpDubbing, OpProjects,
		OpPronunciation, OpConversationalAI,
	}

	for _, op := range ops {
		fields := OperationRequiredFields(op)
		for _, f := range fields {
			if f.Name == "" {
				t.Errorf("%s: field has empty Name", op)
			}
			if f.Description == "" {
				t.Errorf("%s: field %q has empty Description", op, f.Name)
			}
			if f.Example == "" {
				t.Errorf("%s: field %q has empty Example", op, f.Name)
			}
		}
	}
}

func TestToErrorInfo(t *testing.T) {
	tests := []struct {
		code          string
		wantCategory  resilience.ErrorCategory
		wantRetryable bool
	}{
		{ErrNotLoggedIn, resilience.CategoryAuth, false},
		{ErrDocumentNotFound, resilience.CategoryNotFound, false},
		{ErrInvalidUID, resilience.CategoryValidation, false},
		{"UNKNOWN", resilience.CategoryUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			info := ToErrorInfo(tt.code)
			if info.Category != tt.wantCategory {
				t.Errorf("Category = %v, want %v", info.Category, tt.wantCategory)
			}
			if info.Retryable != tt.wantRetryable {
				t.Errorf("Retryable = %v, want %v", info.Retryable, tt.wantRetryable)
			}
			if info.Code != tt.code {
				t.Errorf("Code = %q, want %q", info.Code, tt.code)
			}
			if info.Suggestion == "" {
				t.Error("Suggestion should not be empty")
			}
		})
	}
}

func TestClassifyHTTPStatus(t *testing.T) {
	tests := []struct {
		status        int
		wantCategory  resilience.ErrorCategory
		wantRetryable bool
	}{
		{400, resilience.CategoryValidation, false},
		{401, resilience.CategoryAuth, false},
		{403, resilience.CategoryAuth, false},
		{404, resilience.CategoryNotFound, false},
		{429, resilience.CategoryRateLimit, true},
		{500, resilience.CategoryServer, true},
		{502, resilience.CategoryServer, true},
		{503, resilience.CategoryServer, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.status), func(t *testing.T) {
			info := ClassifyHTTPStatus(tt.status, "test message")
			if info.Category != tt.wantCategory {
				t.Errorf("ClassifyHTTPStatus(%d).Category = %v, want %v", tt.status, info.Category, tt.wantCategory)
			}
			if info.Retryable != tt.wantRetryable {
				t.Errorf("ClassifyHTTPStatus(%d).Retryable = %v, want %v", tt.status, info.Retryable, tt.wantRetryable)
			}
		})
	}
}

func TestAllCategories(t *testing.T) {
	categories := AllCategories()
	if len(categories) == 0 {
		t.Error("AllCategories() returned empty slice")
	}

	// Should contain known categories
	expectedCategories := map[string]bool{
		"auth":       false,
		"not_found":  false,
		"validation": false,
	}

	for _, cat := range categories {
		if _, ok := expectedCategories[cat]; ok {
			expectedCategories[cat] = true
		}
	}

	for cat, found := range expectedCategories {
		if !found {
			t.Errorf("Expected category %q not found in AllCategories()", cat)
		}
	}
}

func TestCodesByCategory(t *testing.T) {
	tests := []struct {
		category  string
		wantCodes []string
	}{
		{"auth", []string{ErrNeedsAuthorization, ErrNotLoggedIn}},
		{"not_found", []string{ErrDocumentNotFound, ErrUserNotFound, ErrWorkspaceNotFound}},
		{"validation", []string{ErrInvalidUID, ErrMissingFeedback, ErrNoEditChanges, ErrUnprocessableEntity}},
		{"unknown_category", nil},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			codes := CodesByCategory(tt.category)
			if tt.wantCodes == nil {
				if len(codes) != 0 {
					t.Errorf("CodesByCategory(%q) = %v, want empty", tt.category, codes)
				}
				return
			}

			// Check that all expected codes are present
			codeSet := make(map[string]bool)
			for _, c := range codes {
				codeSet[c] = true
			}
			for _, want := range tt.wantCodes {
				if !codeSet[want] {
					t.Errorf("CodesByCategory(%q) missing code %q", tt.category, want)
				}
			}
		})
	}
}
