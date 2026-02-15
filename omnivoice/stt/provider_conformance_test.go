package stt

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentplexus/omnivoice/stt/providertest"
)

// testAudioURL is Deepgram's public test audio file (works with any STT provider).
// "Life moves pretty fast. If you don't stop and look around once in a while, you could miss it."
const testAudioURL = "https://static.deepgram.com/examples/Bueller-Life-moves-pretty-fast.wav"

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

	// Download test audio file for TranscribeFile test
	testAudioFile := downloadTestAudio(t)

	providertest.RunAll(t, providertest.Config{
		Provider:          p,
		StreamingProvider: p,
		SkipIntegration:   false,
		TestAudioFile:     testAudioFile,
		TestAudioURL:      testAudioURL,
		TestExpectedText:  "life moves pretty fast",
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

// downloadTestAudio downloads the test audio file to a temp directory.
func downloadTestAudio(t *testing.T) string {
	t.Helper()

	resp, err := http.Get(testAudioURL)
	if err != nil {
		t.Fatalf("failed to download test audio: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("failed to download test audio: status %d", resp.StatusCode)
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-audio.wav")

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		t.Fatalf("failed to write test audio: %v", err)
	}

	return filePath
}
