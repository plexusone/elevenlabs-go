package stt

import (
	"os"
	"testing"

	"github.com/agentplexus/omnivoice/stt/providertest"
)

// TestConformance runs the OmniVoice STT provider conformance tests.
//
// The ElevenLabs STT provider uses the scribe_v2_realtime WebSocket API
// with support for real-time streaming transcription.
func TestConformance(t *testing.T) {
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		t.Skip("ELEVENLABS_API_KEY not set, skipping conformance tests")
	}

	p, err := New(WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	providertest.RunAll(t, providertest.Config{
		Provider:          p,
		StreamingProvider: p,
		SkipIntegration:   false,
	})
}

// TestInterfaceConformance runs only interface tests.
func TestInterfaceConformance(t *testing.T) {
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		t.Skip("ELEVENLABS_API_KEY not set, skipping interface tests that require API")
	}

	p, err := New(WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	providertest.RunInterfaceTests(t, providertest.Config{
		Provider: p,
	})
}
