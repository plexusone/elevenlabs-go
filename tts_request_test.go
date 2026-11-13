package elevenlabs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/plexusone/elevenlabs-go/tts"
)

func TestGenerateSendsDictionariesAndContext(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(data, &body); err != nil {
			t.Errorf("body %s: %v", data, err)
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		if _, err := w.Write([]byte("mp3")); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("test"), WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.TTS().Generate(context.Background(), &tts.Request{
		VoiceID: "v1",
		Text:    "The Miura.",
		PronunciationDictionaries: []tts.PronunciationDictionaryLocator{
			{ID: "d1", VersionID: "v7"},
			{ID: "d2"},
		},
		PreviousText: "Before.",
		NextText:     "After.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadAll(resp.Audio); err != nil {
		t.Fatal(err)
	}

	locs, ok := body["pronunciation_dictionary_locators"].([]any)
	if !ok || len(locs) != 2 {
		t.Fatalf("locators = %v", body["pronunciation_dictionary_locators"])
	}
	first, ok := locs[0].(map[string]any)
	if !ok || first["pronunciation_dictionary_id"] != "d1" || first["version_id"] != "v7" {
		t.Errorf("first locator = %v", locs[0])
	}
	second, ok := locs[1].(map[string]any)
	if !ok || second["pronunciation_dictionary_id"] != "d2" || second["version_id"] != nil {
		t.Errorf("second locator = %v", locs[1])
	}
	if body["previous_text"] != "Before." || body["next_text"] != "After." {
		t.Errorf("context = %v, %v", body["previous_text"], body["next_text"])
	}
}

func TestRequestValidatesDictionaries(t *testing.T) {
	r := &tts.Request{VoiceID: "v", Text: "t", PronunciationDictionaries: make([]tts.PronunciationDictionaryLocator, tts.MaxPronunciationDictionaries+1)}
	if err := r.Validate(); err == nil {
		t.Error("too many dictionaries were accepted")
	}
	r.PronunciationDictionaries = []tts.PronunciationDictionaryLocator{{}}
	if err := r.Validate(); err == nil {
		t.Error("a dictionary without an ID was accepted")
	}
}

func TestSupportsPhonemes(t *testing.T) {
	for model, want := range map[string]bool{
		"":                       false,
		"eleven_multilingual_v2": false,
		"eleven_flash_v2":        true,
		"eleven_flash_v2_5":      false,
		"eleven_turbo_v2":        true,
		"eleven_monolingual_v1":  true,
	} {
		if got := tts.SupportsPhonemes(model); got != want {
			t.Errorf("SupportsPhonemes(%q) = %v", model, got)
		}
	}
}
