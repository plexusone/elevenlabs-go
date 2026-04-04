package omnivoice_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	elevenlabs "github.com/plexusone/elevenlabs-go"
	elevenlabstts "github.com/plexusone/elevenlabs-go/omnivoice/tts"
	"github.com/plexusone/omnivoice-core/resilience"
	"github.com/plexusone/omnivoice-core/tts"
)

// mockAPIServer creates a test server that simulates ElevenLabs API behavior.
type mockAPIServer struct {
	*httptest.Server
	requestCount atomic.Int32
	responses    []mockResponse
	currentIdx   atomic.Int32
}

type mockResponse struct {
	statusCode int
	body       interface{}
	delay      time.Duration
}

func newMockAPIServer(responses ...mockResponse) *mockAPIServer {
	m := &mockAPIServer{
		responses: responses,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.requestCount.Add(1)
		idx := int(m.currentIdx.Add(1)) - 1

		var resp mockResponse
		if idx < len(m.responses) {
			resp = m.responses[idx]
		} else if len(m.responses) > 0 {
			// Use last response for subsequent requests
			resp = m.responses[len(m.responses)-1]
		} else {
			resp = mockResponse{statusCode: 200, body: map[string]string{"status": "ok"}}
		}

		if resp.delay > 0 {
			time.Sleep(resp.delay)
		}

		// Set content type based on status code and body type
		if resp.statusCode == 200 {
			// Success responses for TTS are audio
			w.Header().Set("Content-Type", "audio/mpeg")
		} else {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(resp.statusCode)

		if resp.body != nil {
			switch v := resp.body.(type) {
			case []byte:
				_, _ = w.Write(v)
			case string:
				_, _ = w.Write([]byte(v))
			default:
				_ = json.NewEncoder(w).Encode(v)
			}
		}
	}))

	return m
}

func (m *mockAPIServer) getRequestCount() int {
	return int(m.requestCount.Load())
}

// TestRetry_RateLimitThenSuccess tests that rate limit errors are retried.
func TestRetry_RateLimitThenSuccess(t *testing.T) {
	// First request: 429 rate limit
	// Second request: 429 rate limit
	// Third request: 200 success
	server := newMockAPIServer(
		mockResponse{statusCode: 429, body: map[string]string{"detail": "rate limit exceeded"}},
		mockResponse{statusCode: 429, body: map[string]string{"detail": "rate limit exceeded"}},
		mockResponse{statusCode: 200, body: []byte("fake audio data")},
	)
	defer server.Close()

	client, err := elevenlabs.NewClient(
		elevenlabs.WithAPIKey("test-key"),
		elevenlabs.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	provider := elevenlabstts.NewWithClient(client)

	// Use fast retry config for testing
	provider.SetRetryConfig(resilience.RetryConfig{
		MaxAttempts: 5,
		Backoff:     &resilience.NoBackoff{},
		Classifier:  provider.Classifier(),
	})

	ctx := context.Background()
	_, err = provider.Synthesize(ctx, "Hello", ttsConfig())

	// Should succeed after retries
	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}

	// Should have made 3 requests
	if count := server.getRequestCount(); count != 3 {
		t.Errorf("Expected 3 requests, got %d", count)
	}
}

// TestRetry_AuthErrorNoRetry tests that auth errors are not retried.
func TestRetry_AuthErrorNoRetry(t *testing.T) {
	server := newMockAPIServer(
		mockResponse{statusCode: 401, body: map[string]string{"detail": "NOT_LOGGED_IN"}},
	)
	defer server.Close()

	client, err := elevenlabs.NewClient(
		elevenlabs.WithAPIKey("invalid-key"),
		elevenlabs.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	provider := elevenlabstts.NewWithClient(client)
	provider.SetRetryConfig(resilience.RetryConfig{
		MaxAttempts: 3,
		Backoff:     &resilience.NoBackoff{},
		Classifier:  provider.Classifier(),
	})

	ctx := context.Background()
	_, err = provider.Synthesize(ctx, "Hello", ttsConfig())

	// Should fail immediately without retry
	if err == nil {
		t.Error("Expected error for auth failure")
	}

	// Should have made only 1 request (no retries for auth errors)
	if count := server.getRequestCount(); count != 1 {
		t.Errorf("Expected 1 request (no retry for auth), got %d", count)
	}

	// Error should be classified as non-retryable
	if pe, ok := resilience.IsProviderError(err); ok {
		if pe.IsRetryable() {
			t.Error("Auth error should not be retryable")
		}
		if pe.GetCategory() != resilience.CategoryAuth {
			t.Errorf("Expected CategoryAuth, got %v", pe.GetCategory())
		}
	}
}

// TestRetry_ServerErrorThenSuccess tests that server errors are retried.
func TestRetry_ServerErrorThenSuccess(t *testing.T) {
	server := newMockAPIServer(
		mockResponse{statusCode: 500, body: map[string]string{"detail": "internal server error"}},
		mockResponse{statusCode: 200, body: []byte("audio data")},
	)
	defer server.Close()

	client, err := elevenlabs.NewClient(
		elevenlabs.WithAPIKey("test-key"),
		elevenlabs.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	provider := elevenlabstts.NewWithClient(client)
	provider.SetRetryConfig(resilience.RetryConfig{
		MaxAttempts: 3,
		Backoff:     &resilience.NoBackoff{},
		Classifier:  provider.Classifier(),
	})

	ctx := context.Background()
	_, err = provider.Synthesize(ctx, "Hello", ttsConfig())

	if err != nil {
		t.Errorf("Expected success after retry, got: %v", err)
	}

	if count := server.getRequestCount(); count != 2 {
		t.Errorf("Expected 2 requests, got %d", count)
	}
}

// TestRetry_ExhaustedRetries tests behavior when all retries are exhausted.
func TestRetry_ExhaustedRetries(t *testing.T) {
	server := newMockAPIServer(
		mockResponse{statusCode: 429, body: map[string]string{"detail": "rate limit"}},
		mockResponse{statusCode: 429, body: map[string]string{"detail": "rate limit"}},
		mockResponse{statusCode: 429, body: map[string]string{"detail": "rate limit"}},
	)
	defer server.Close()

	client, err := elevenlabs.NewClient(
		elevenlabs.WithAPIKey("test-key"),
		elevenlabs.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	provider := elevenlabstts.NewWithClient(client)
	provider.SetRetryConfig(resilience.RetryConfig{
		MaxAttempts: 3,
		Backoff:     &resilience.NoBackoff{},
		Classifier:  provider.Classifier(),
	})

	ctx := context.Background()
	_, err = provider.Synthesize(ctx, "Hello", ttsConfig())

	// Should fail after exhausting retries
	if err == nil {
		t.Error("Expected error after exhausting retries")
	}

	// Should be a RetryError
	var retryErr *resilience.RetryError
	if !isRetryError(err, &retryErr) {
		t.Errorf("Expected RetryError, got %T", err)
	} else if retryErr.Attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", retryErr.Attempts)
	}

	if count := server.getRequestCount(); count != 3 {
		t.Errorf("Expected 3 requests, got %d", count)
	}
}

// TestRetry_NotFoundNoRetry tests that 404 errors are not retried.
func TestRetry_NotFoundNoRetry(t *testing.T) {
	server := newMockAPIServer(
		mockResponse{statusCode: 404, body: map[string]string{"detail": "DOCUMENT_NOT_FOUND"}},
	)
	defer server.Close()

	client, err := elevenlabs.NewClient(
		elevenlabs.WithAPIKey("test-key"),
		elevenlabs.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	provider := elevenlabstts.NewWithClient(client)
	provider.SetRetryConfig(resilience.RetryConfig{
		MaxAttempts: 3,
		Backoff:     &resilience.NoBackoff{},
		Classifier:  provider.Classifier(),
	})

	ctx := context.Background()
	_, err = provider.Synthesize(ctx, "Hello", ttsConfig())

	if err == nil {
		t.Error("Expected error for not found")
	}

	// Should have made only 1 request
	if count := server.getRequestCount(); count != 1 {
		t.Errorf("Expected 1 request (no retry for 404), got %d", count)
	}
}

// TestRetry_ValidationErrorNoRetry tests that validation errors are not retried.
func TestRetry_ValidationErrorNoRetry(t *testing.T) {
	server := newMockAPIServer(
		mockResponse{statusCode: 422, body: map[string]string{"detail": "UNPROCESSABLE_ENTITY"}},
	)
	defer server.Close()

	client, err := elevenlabs.NewClient(
		elevenlabs.WithAPIKey("test-key"),
		elevenlabs.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	provider := elevenlabstts.NewWithClient(client)
	provider.SetRetryConfig(resilience.RetryConfig{
		MaxAttempts: 3,
		Backoff:     &resilience.NoBackoff{},
		Classifier:  provider.Classifier(),
	})

	ctx := context.Background()
	_, err = provider.Synthesize(ctx, "Hello", ttsConfig())

	if err == nil {
		t.Error("Expected error for validation failure")
	}

	if count := server.getRequestCount(); count != 1 {
		t.Errorf("Expected 1 request (no retry for validation), got %d", count)
	}
}

// TestRetry_OnRetryCallback tests that the OnRetry callback is invoked.
func TestRetry_OnRetryCallback(t *testing.T) {
	server := newMockAPIServer(
		mockResponse{statusCode: 429, body: map[string]string{"detail": "rate limit"}},
		mockResponse{statusCode: 200, body: []byte("audio")},
	)
	defer server.Close()

	client, err := elevenlabs.NewClient(
		elevenlabs.WithAPIKey("test-key"),
		elevenlabs.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	var callbackCalls []int
	provider := elevenlabstts.NewWithClient(client)
	provider.SetRetryConfig(resilience.RetryConfig{
		MaxAttempts: 3,
		Backoff:     &resilience.NoBackoff{},
		Classifier:  provider.Classifier(),
		OnRetry: func(attempt int, err error, delay time.Duration) {
			callbackCalls = append(callbackCalls, attempt)
		},
	})

	ctx := context.Background()
	_, _ = provider.Synthesize(ctx, "Hello", ttsConfig())

	// OnRetry should be called once (before the second attempt)
	if len(callbackCalls) != 1 {
		t.Errorf("Expected 1 OnRetry call, got %d", len(callbackCalls))
	}
	if len(callbackCalls) > 0 && callbackCalls[0] != 1 {
		t.Errorf("Expected attempt 1 in callback, got %d", callbackCalls[0])
	}
}

// TestRetry_ContextCancellation tests that retries respect context cancellation.
func TestRetry_ContextCancellation(t *testing.T) {
	server := newMockAPIServer(
		mockResponse{statusCode: 429, body: map[string]string{"detail": "rate limit"}, delay: 100 * time.Millisecond},
		mockResponse{statusCode: 429, body: map[string]string{"detail": "rate limit"}},
	)
	defer server.Close()

	client, err := elevenlabs.NewClient(
		elevenlabs.WithAPIKey("test-key"),
		elevenlabs.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	provider := elevenlabstts.NewWithClient(client)
	provider.SetRetryConfig(resilience.RetryConfig{
		MaxAttempts: 10,
		Backoff:     &resilience.ConstantBackoff{Delay: 50 * time.Millisecond},
		Classifier:  provider.Classifier(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err = provider.Synthesize(ctx, "Hello", ttsConfig())

	// Should fail due to context cancellation
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
}

// Helper to create TTS config for tests
func ttsConfig() tts.SynthesisConfig {
	return tts.SynthesisConfig{
		VoiceID: "21m00Tcm4TlvDq8ikWAM",
	}
}

// Helper to check for RetryError
func isRetryError(err error, target **resilience.RetryError) bool {
	if err == nil {
		return false
	}
	// Unwrap the error chain
	for e := err; e != nil; {
		if re, ok := e.(*resilience.RetryError); ok {
			*target = re
			return true
		}
		if unwrapper, ok := e.(interface{ Unwrap() error }); ok {
			e = unwrapper.Unwrap()
		} else {
			break
		}
	}
	return false
}
