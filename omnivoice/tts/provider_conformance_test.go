package tts

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/plexusone/omnivoice-core/tts/providertest"
)

// TestConformance runs the OmniVoice TTS provider conformance tests.
func TestConformance(t *testing.T) {
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		t.Skip("ELEVENLABS_API_KEY not set, skipping conformance tests")
	}

	p, err := New(WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	// Find a voice the user has access to (prefer user's own voices over library voices)
	testVoiceID := findAccessibleVoice(t, p)

	providertest.RunAll(t, providertest.Config{
		Provider:          p,
		StreamingProvider: p,
		SkipIntegration:   false,
		TestVoiceID:       testVoiceID,
		TestText:          "Hello, this is a conformance test.",
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

// findAccessibleVoice finds a voice the user can access for testing.
// It tries to get each voice and returns the first one that works.
func findAccessibleVoice(t *testing.T, p *Provider) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	voices, err := p.ListVoices(ctx)
	if err != nil {
		t.Fatalf("ListVoices() error: %v", err)
	}

	if len(voices) == 0 {
		t.Fatal("no voices available")
	}

	// Try to find a voice we can actually access
	for _, v := range voices {
		_, err := p.GetVoice(ctx, v.ID)
		if err == nil {
			t.Logf("Using voice: %s (%s)", v.ID, v.Name)
			return v.ID
		}
	}

	// Fallback to first voice and let the test fail with proper error
	t.Logf("Warning: Could not find accessible voice, using first voice: %s", voices[0].ID)
	return voices[0].ID
}
