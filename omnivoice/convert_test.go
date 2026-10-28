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
