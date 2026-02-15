package ttsconfig

import "testing"

func TestOratory(t *testing.T) {
	cfg := Oratory()

	if cfg.ModelID != "eleven_v3" {
		t.Errorf("ModelID = %q, want %q", cfg.ModelID, "eleven_v3")
	}
	if cfg.OutputFormat != "pcm_48000" {
		t.Errorf("OutputFormat = %q, want %q", cfg.OutputFormat, "pcm_48000")
	}
	if cfg.VoiceSettings == nil {
		t.Fatal("VoiceSettings is nil")
	}
	if *cfg.VoiceSettings.Stability != 0.4 {
		t.Errorf("Stability = %v, want 0.4", *cfg.VoiceSettings.Stability)
	}
	if *cfg.VoiceSettings.Style != 0.3 {
		t.Errorf("Style = %v, want 0.3", *cfg.VoiceSettings.Style)
	}
	if *cfg.VoiceSettings.Speed != 0.95 {
		t.Errorf("Speed = %v, want 0.95", *cfg.VoiceSettings.Speed)
	}
}

func TestPodcast(t *testing.T) {
	cfg := Podcast()

	if cfg.OutputFormat != "mp3_44100_128" {
		t.Errorf("OutputFormat = %q, want %q", cfg.OutputFormat, "mp3_44100_128")
	}
	if *cfg.VoiceSettings.Stability != 0.5 {
		t.Errorf("Stability = %v, want 0.5", *cfg.VoiceSettings.Stability)
	}
	if *cfg.VoiceSettings.Style != 0.0 {
		t.Errorf("Style = %v, want 0.0", *cfg.VoiceSettings.Style)
	}
	if *cfg.VoiceSettings.Speed != 1.0 {
		t.Errorf("Speed = %v, want 1.0", *cfg.VoiceSettings.Speed)
	}
}

func TestAudiobook(t *testing.T) {
	cfg := Audiobook()

	if cfg.OutputFormat != "pcm_48000" {
		t.Errorf("OutputFormat = %q, want %q", cfg.OutputFormat, "pcm_48000")
	}
	if *cfg.VoiceSettings.Stability != 0.6 {
		t.Errorf("Stability = %v, want 0.6", *cfg.VoiceSettings.Stability)
	}
	if *cfg.VoiceSettings.Style != 0.1 {
		t.Errorf("Style = %v, want 0.1", *cfg.VoiceSettings.Style)
	}
	if *cfg.VoiceSettings.SimilarityBoost != 0.8 {
		t.Errorf("SimilarityBoost = %v, want 0.8", *cfg.VoiceSettings.SimilarityBoost)
	}
}

func TestPresetNames(t *testing.T) {
	names := PresetNames()

	if len(names) != 3 {
		t.Errorf("len(PresetNames()) = %d, want 3", len(names))
	}

	expected := map[string]bool{
		"oratory":   true,
		"podcast":   true,
		"audiobook": true,
	}

	for _, name := range names {
		if !expected[name] {
			t.Errorf("unexpected preset name: %q", name)
		}
	}
}

func TestGetPreset(t *testing.T) {
	tests := []struct {
		name    string
		wantNil bool
	}{
		{"oratory", false},
		{"podcast", false},
		{"audiobook", false},
		{"unknown", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GetPreset(tt.name)
			if (cfg == nil) != tt.wantNil {
				t.Errorf("GetPreset(%q) = %v, wantNil = %v", tt.name, cfg, tt.wantNil)
			}
		})
	}
}
