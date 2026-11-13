package tts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/plexusone/omnivoice-core/tts"
)

// TestSynthesizeStreamSSMLParsing checks the WebSocket handshake: SSML
// parsing is enabled only when the streamed text carries phoneme tags.
func TestSynthesizeStreamSSMLParsing(t *testing.T) {
	miura := []tts.Pronunciation{{Term: "Miura", IPA: "ˈmjura", Respelling: "mee-OO-rah"}}
	cases := []struct {
		name   string
		config tts.SynthesisConfig
		want   string
	}{
		{"phoneme tags", tts.SynthesisConfig{Model: "eleven_flash_v2", Pronunciations: miura}, "true"},
		{"respellings", tts.SynthesisConfig{Model: "eleven_multilingual_v2", Pronunciations: miura}, ""},
		{"phonemes turned off", WithPhonemes(tts.SynthesisConfig{Model: "eleven_flash_v2", Pronunciations: miura}, false), ""},
		{"phonemes turned on", WithPhonemes(tts.SynthesisConfig{Model: "eleven_multilingual_v2", Pronunciations: miura}, true), "true"},
		{"respelling only", tts.SynthesisConfig{Model: "eleven_flash_v2", Pronunciations: []tts.Pronunciation{{Term: "Miura", Respelling: "mee-OO-rah"}}}, ""},
		{"no pronunciations", tts.SynthesisConfig{Model: "eleven_flash_v2"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var mu sync.Mutex
			var got string
			var called bool
			// The server records the handshake and refuses the upgrade.
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				got, called = r.URL.Query().Get("enable_ssml_parsing"), true
				mu.Unlock()
				http.Error(w, "no upgrade", http.StatusBadRequest)
			}))
			defer srv.Close()

			p, err := New(WithAPIKey("test-key"), WithBaseURL(srv.URL))
			if err != nil {
				t.Fatal(err)
			}
			tc.config.VoiceID = "voice"
			if _, err := p.SynthesizeStream(context.Background(), "The Miura.", tc.config); err == nil {
				t.Fatal("SynthesizeStream connected to a server that refuses the upgrade")
			}
			mu.Lock()
			defer mu.Unlock()
			if !called {
				t.Fatal("no handshake reached the server")
			}
			if got != tc.want {
				t.Errorf("enable_ssml_parsing = %q, want %q", got, tc.want)
			}
		})
	}
}
