package omnivoice_test

import (
	"errors"
	"testing"

	"github.com/plexusone/elevenlabs-go/ax"
	"github.com/plexusone/elevenlabs-go/omnivoice"
	"github.com/plexusone/omnivoice-core/resilience"
)

// BenchmarkClassifier_Classify benchmarks error classification.
// Target: <1ms per classification
func BenchmarkClassifier_Classify(b *testing.B) {
	classifier := omnivoice.NewClassifier()
	err := errors.New("rate limit exceeded")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = classifier.Classify(err)
	}
}

// BenchmarkClassifier_ClassifyProviderError benchmarks classifying a ProviderError.
func BenchmarkClassifier_ClassifyProviderError(b *testing.B) {
	classifier := omnivoice.NewClassifier()
	err := resilience.NewProviderError("elevenlabs", "Synthesize", errors.New("rate limit"), resilience.ErrorInfo{
		Category:  resilience.CategoryRateLimit,
		Retryable: true,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = classifier.Classify(err)
	}
}

// BenchmarkCategoryForCode benchmarks AX code to category mapping.
func BenchmarkCategoryForCode(b *testing.B) {
	code := ax.ErrNotLoggedIn

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ax.CategoryForCode(code)
	}
}

// BenchmarkIsRetryableCode benchmarks retryability check.
func BenchmarkIsRetryableCode(b *testing.B) {
	code := ax.ErrDocumentNotFound

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ax.IsRetryableCode(code)
	}
}

// BenchmarkToErrorInfo benchmarks full ErrorInfo conversion.
func BenchmarkToErrorInfo(b *testing.B) {
	code := ax.ErrNeedsAuthorization

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ax.ToErrorInfo(code)
	}
}

// BenchmarkShouldRetry benchmarks retry decision logic.
// Target: <0.1ms per decision
func BenchmarkShouldRetry(b *testing.B) {
	config := resilience.DefaultRetryConfig()

	b.Run("retryable_error", func(b *testing.B) {
		err := resilience.NewProviderError("test", "op", errors.New("rate limit"), resilience.ErrorInfo{
			Retryable: true,
		})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = resilience.IsRetryable(err)
		}
	})

	b.Run("non_retryable_error", func(b *testing.B) {
		err := resilience.NewProviderError("test", "op", errors.New("auth"), resilience.ErrorInfo{
			Retryable: false,
		})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = resilience.IsRetryable(err)
		}
	})

	b.Run("with_classifier", func(b *testing.B) {
		err := errors.New("rate limit exceeded")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			info := config.Classifier.Classify(err)
			_ = info.Retryable
		}
	})
}

// BenchmarkBackoff_NextDelay benchmarks backoff calculation.
func BenchmarkBackoff_NextDelay(b *testing.B) {
	b.Run("exponential", func(b *testing.B) {
		backoff := &resilience.ExponentialBackoff{
			Initial:    1000000000, // 1s in nanoseconds
			Max:        30000000000,
			Multiplier: 2.0,
			Jitter:     0.1,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = backoff.NextDelay(3)
		}
	})

	b.Run("constant", func(b *testing.B) {
		backoff := &resilience.ConstantBackoff{Delay: 1000000000}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = backoff.NextDelay(3)
		}
	})

	b.Run("linear", func(b *testing.B) {
		backoff := &resilience.LinearBackoff{
			Initial:   1000000000,
			Increment: 1000000000,
			Max:       30000000000,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = backoff.NextDelay(3)
		}
	})
}

// BenchmarkContainsErrorCode benchmarks AX error code extraction.
func BenchmarkContainsErrorCode(b *testing.B) {
	b.Run("contains_code", func(b *testing.B) {
		err := errors.New("error: NOT_LOGGED_IN - please authenticate")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ax.ContainsErrorCode(err)
		}
	})

	b.Run("no_code", func(b *testing.B) {
		err := errors.New("generic network error")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = ax.ContainsErrorCode(err)
		}
	})
}

// BenchmarkOperationRequiredFields benchmarks field lookup.
func BenchmarkOperationRequiredFields(b *testing.B) {
	op := ax.OpTextToSpeech

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ax.OperationRequiredFields(op)
	}
}
