// Package omnivoice provides OmniVoice integration for elevenlabs-go.
package omnivoice

import (
	"time"

	elevenlabs "github.com/plexusone/elevenlabs-go"
	"github.com/plexusone/elevenlabs-go/ax"
	"github.com/plexusone/omnivoice-core/resilience"
)

// Classifier implements resilience.ErrorClassifier for ElevenLabs API errors.
// It uses AX error codes and HTTP status codes to provide accurate classification.
type Classifier struct {
	httpClassifier resilience.HTTPStatusClassifier
}

// NewClassifier creates a new ElevenLabs error classifier.
func NewClassifier() *Classifier {
	return &Classifier{}
}

// Classify analyzes an ElevenLabs error and returns its metadata.
// It prioritizes AX error codes when available, falling back to HTTP status classification.
func (c *Classifier) Classify(err error) resilience.ErrorInfo {
	if err == nil {
		return resilience.ErrorInfo{Category: resilience.CategoryUnknown}
	}

	// Check for wrapped resilience.ProviderError first
	if pe, ok := resilience.IsProviderError(err); ok {
		return pe.Info
	}

	// Try to parse as ElevenLabs APIError
	apiErr := elevenlabs.ParseAPIError(err)
	if apiErr != nil {
		return c.classifyAPIError(apiErr)
	}

	// Fall back to default classification
	return (&resilience.DefaultClassifier{}).Classify(err)
}

// classifyAPIError classifies an ElevenLabs APIError using AX codes and HTTP status.
func (c *Classifier) classifyAPIError(apiErr *elevenlabs.APIError) resilience.ErrorInfo {
	// First, try to extract AX error code for precise classification
	if code, ok := apiErr.AXErrorCode(); ok {
		return c.classifyAXCode(code, apiErr)
	}

	// Fall back to HTTP status code classification
	info := c.httpClassifier.ClassifyStatus(apiErr.StatusCode, apiErr.Message)

	// Add detail if available
	if apiErr.Detail != "" {
		info.Message = apiErr.Detail
	}

	return info
}

// classifyAXCode maps AX error codes to resilience.ErrorInfo.
func (c *Classifier) classifyAXCode(code string, apiErr *elevenlabs.APIError) resilience.ErrorInfo {
	// Get metadata from ax package
	axInfo := ax.GetErrorInfo(code)
	if axInfo == nil {
		// Unknown AX code, fall back to HTTP status
		return c.httpClassifier.ClassifyStatus(apiErr.StatusCode, apiErr.Message)
	}

	// Map ax.ErrorCodeInfo to resilience.ErrorInfo
	info := resilience.ErrorInfo{
		Category:   mapAXCategory(axInfo.Category),
		Retryable:  axInfo.Retryable,
		Code:       code,
		Message:    axInfo.Description,
		Suggestion: suggestionForAXCode(code),
	}

	// Override message with API detail if more specific
	if apiErr.Detail != "" {
		info.Message = apiErr.Detail
	}

	return info
}

// mapAXCategory maps AX category strings to resilience.ErrorCategory.
func mapAXCategory(category string) resilience.ErrorCategory {
	switch category {
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

// suggestionForAXCode returns a helpful suggestion for each AX error code.
func suggestionForAXCode(code string) string {
	switch code {
	case ax.ErrDocumentNotFound:
		return "Verify the document ID exists and you have access to it"
	case ax.ErrInvalidUID:
		return "Check that the provided ID is a valid format"
	case ax.ErrMissingFeedback:
		return "Include required feedback in the request"
	case ax.ErrNeedsAuthorization:
		return "This operation requires additional authorization; check your subscription tier"
	case ax.ErrNotLoggedIn:
		return "Provide a valid API key in the request"
	case ax.ErrNoEditChanges:
		return "Include at least one change in the edit request"
	case ax.ErrUnprocessableEntity:
		return "Check the request body for validation errors"
	case ax.ErrUserNotFound:
		return "Verify the user ID exists"
	case ax.ErrWorkspaceNotFound:
		return "Verify the workspace ID exists and you have access to it"
	default:
		return "Check the error details and ElevenLabs API documentation"
	}
}

// ClassifyWithRetryAfter classifies an error and extracts Retry-After hint if available.
func (c *Classifier) ClassifyWithRetryAfter(err error, retryAfter time.Duration) resilience.ErrorInfo {
	info := c.Classify(err)
	if retryAfter > 0 {
		info.RetryAfter = retryAfter
	}
	return info
}

// WrapError wraps an ElevenLabs error as a resilience.ProviderError with classification.
func (c *Classifier) WrapError(op string, err error) error {
	if err == nil {
		return nil
	}

	info := c.Classify(err)
	return resilience.NewProviderError(ProviderName, op, err, info)
}

// DefaultRetryConfig returns a RetryConfig suitable for ElevenLabs API calls.
// It uses:
//   - 3 max attempts (1 initial + 2 retries)
//   - Exponential backoff starting at 1s, max 30s
//   - ElevenLabs-specific error classification
func DefaultRetryConfig() resilience.RetryConfig {
	return resilience.RetryConfig{
		MaxAttempts: 3,
		Backoff:     resilience.DefaultBackoff(),
		Classifier:  NewClassifier(),
	}
}

// RetryConfigWithCallback returns a RetryConfig with a callback for monitoring retries.
func RetryConfigWithCallback(onRetry func(attempt int, err error, delay time.Duration)) resilience.RetryConfig {
	config := DefaultRetryConfig()
	config.OnRetry = onRetry
	return config
}
