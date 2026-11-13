package omnivoice

import (
	"testing"

	"github.com/plexusone/omnivoice-core/tts"
)

func TestOutputSampleRate(t *testing.T) {
	tests := []struct {
		format string
		rate   int
		want   int
	}{
		{"", 0, 44100},
		{"", 8000, 44100}, // no format: ElevenLabs' mp3 default
		{"mp3", 0, 44100},
		{"mp3", 22050, 22050},
		{"pcm", 16000, 16000},
		{"pcm", 0, 44100},
		{"mulaw", 0, 8000},
		{"ulaw", 44100, 8000},
		{"alaw", 0, 8000},
		{"ulaw_8000", 0, 8000},
		{"pcm_24000", 0, 24000},
		{"mp3_22050_32", 44100, 22050},
		{"opus_48000_64", 0, 48000},
	}
	for _, tt := range tests {
		got := OutputSampleRate(tts.SynthesisConfig{OutputFormat: tt.format, SampleRate: tt.rate})
		if got != tt.want {
			t.Errorf("OutputSampleRate(%q, %d) = %d, want %d", tt.format, tt.rate, got, tt.want)
		}
	}
}

func TestConfigToTTSRequestPronunciations(t *testing.T) {
	ps := []tts.Pronunciation{{Term: "Miura", IPA: "ˈmjura", Respelling: "mee oora"}}
	cases := []struct {
		name   string
		config tts.SynthesisConfig
		want   string
	}{
		{"default model reads respellings", tts.SynthesisConfig{Pronunciations: ps}, "The mee oora."},
		{"flash v2 takes phonemes", tts.SynthesisConfig{Model: "eleven_flash_v2", Pronunciations: ps}, `The <phoneme alphabet="ipa" ph="ˈmjura">Miura</phoneme>.`},
		{"forced phonemes", tts.SynthesisConfig{Pronunciations: ps, Extensions: map[string]any{extPhonemes: true}}, `The <phoneme alphabet="ipa" ph="ˈmjura">Miura</phoneme>.`},
		{"forced respellings", tts.SynthesisConfig{Model: "eleven_flash_v2", Pronunciations: ps, Extensions: map[string]any{extPhonemes: false}}, "The mee oora."},
	}
	for _, tc := range cases {
		if got := ConfigToTTSRequest("The Miura.", tc.config).Text; got != tc.want {
			t.Errorf("%s: %s", tc.name, got)
		}
	}
}

func TestConfigToTTSRequestExtensions(t *testing.T) {
	req := ConfigToTTSRequest("x", tts.SynthesisConfig{Extensions: map[string]any{
		extPronunciationDictionaryID:   "d1",
		extPronunciationDictionaryVers: "v2",
		extPreviousText:                "before",
		extNextText:                    "after",
	}})
	if len(req.PronunciationDictionaries) != 1 || req.PronunciationDictionaries[0].ID != "d1" || req.PronunciationDictionaries[0].VersionID != "v2" {
		t.Errorf("dictionaries = %+v", req.PronunciationDictionaries)
	}
	if req.PreviousText != "before" || req.NextText != "after" {
		t.Errorf("context = %q, %q", req.PreviousText, req.NextText)
	}
}
