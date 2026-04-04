package omnivoice

import (
	"errors"
	"testing"
	"time"

	"github.com/plexusone/elevenlabs-go/ax"
	"github.com/plexusone/omnivoice-core/resilience"
)

func TestClassifier_Classify(t *testing.T) {
	c := NewClassifier()

	tests := []struct {
		name          string
		err           error
		wantCategory  resilience.ErrorCategory
		wantRetryable bool
	}{
		{
			name:          "nil error",
			err:           nil,
			wantCategory:  resilience.CategoryUnknown,
			wantRetryable: false,
		},
		{
			name:          "regular error",
			err:           errors.New("something went wrong"),
			wantCategory:  resilience.CategoryUnknown,
			wantRetryable: false,
		},
		{
			name:          "rate limit in message",
			err:           errors.New("rate limit exceeded"),
			wantCategory:  resilience.CategoryRateLimit,
			wantRetryable: true,
		},
		{
			name:          "unauthorized in message",
			err:           errors.New("unauthorized access"),
			wantCategory:  resilience.CategoryAuth,
			wantRetryable: false,
		},
		{
			name:          "not found in message",
			err:           errors.New("resource not found"),
			wantCategory:  resilience.CategoryNotFound,
			wantRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := c.Classify(tt.err)
			if info.Category != tt.wantCategory {
				t.Errorf("Category = %v, want %v", info.Category, tt.wantCategory)
			}
			if info.Retryable != tt.wantRetryable {
				t.Errorf("Retryable = %v, want %v", info.Retryable, tt.wantRetryable)
			}
		})
	}
}

func TestClassifier_ClassifyProviderError(t *testing.T) {
	c := NewClassifier()

	pe := resilience.NewProviderError("test", "op", errors.New("test"), resilience.ErrorInfo{
		Category:  resilience.CategoryRateLimit,
		Retryable: true,
		Code:      "RATE_LIMITED",
	})

	info := c.Classify(pe)
	if info.Category != resilience.CategoryRateLimit {
		t.Errorf("Category = %v, want %v", info.Category, resilience.CategoryRateLimit)
	}
	if !info.Retryable {
		t.Error("Retryable should be true")
	}
}

func TestMapAXCategory(t *testing.T) {
	tests := []struct {
		input string
		want  resilience.ErrorCategory
	}{
		{"auth", resilience.CategoryAuth},
		{"not_found", resilience.CategoryNotFound},
		{"validation", resilience.CategoryValidation},
		{"rate_limit", resilience.CategoryRateLimit},
		{"quota", resilience.CategoryQuota},
		{"server", resilience.CategoryServer},
		{"transient", resilience.CategoryTransient},
		{"unknown", resilience.CategoryUnknown},
		{"", resilience.CategoryUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := mapAXCategory(tt.input)
			if got != tt.want {
				t.Errorf("mapAXCategory(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSuggestionForAXCode(t *testing.T) {
	// Test that all known AX codes have suggestions
	codes := []string{
		ax.ErrDocumentNotFound,
		ax.ErrInvalidUID,
		ax.ErrMissingFeedback,
		ax.ErrNeedsAuthorization,
		ax.ErrNotLoggedIn,
		ax.ErrNoEditChanges,
		ax.ErrUnprocessableEntity,
		ax.ErrUserNotFound,
		ax.ErrWorkspaceNotFound,
	}

	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			suggestion := suggestionForAXCode(code)
			if suggestion == "" {
				t.Errorf("suggestionForAXCode(%q) returned empty string", code)
			}
		})
	}

	// Test unknown code returns generic suggestion
	suggestion := suggestionForAXCode("UNKNOWN_CODE")
	if suggestion == "" {
		t.Error("suggestionForAXCode for unknown code returned empty string")
	}
}

func TestClassifier_WrapError(t *testing.T) {
	c := NewClassifier()

	// Wrap nil error
	if got := c.WrapError("test", nil); got != nil {
		t.Errorf("WrapError(nil) = %v, want nil", got)
	}

	// Wrap regular error
	originalErr := errors.New("test error")
	wrappedErr := c.WrapError("Synthesize", originalErr)
	if wrappedErr == nil {
		t.Fatal("WrapError should not return nil for non-nil error")
	}

	// Should be a ProviderError
	pe, ok := resilience.IsProviderError(wrappedErr)
	if !ok {
		t.Fatal("WrapError should return a ProviderError")
	}

	if pe.Provider != ProviderName {
		t.Errorf("Provider = %q, want %q", pe.Provider, ProviderName)
	}
	if pe.Op != "Synthesize" {
		t.Errorf("Op = %q, want %q", pe.Op, "Synthesize")
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("MaxAttempts = %d, want 3", config.MaxAttempts)
	}
	if config.Backoff == nil {
		t.Error("Backoff should not be nil")
	}
	if config.Classifier == nil {
		t.Error("Classifier should not be nil")
	}

	// Classifier should be our ElevenLabs classifier
	_, ok := config.Classifier.(*Classifier)
	if !ok {
		t.Error("Classifier should be *Classifier")
	}
}

func TestRetryConfigWithCallback(t *testing.T) {
	var called bool
	config := RetryConfigWithCallback(func(attempt int, err error, delay time.Duration) {
		called = true
	})

	if config.OnRetry == nil {
		t.Error("OnRetry should not be nil")
	}

	// Verify callback is set (we won't actually call it here)
	_ = called
}
