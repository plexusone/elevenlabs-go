package ttsconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.ModelID != "eleven_v3" {
		t.Errorf("ModelID = %q, want %q", cfg.ModelID, "eleven_v3")
	}
	if cfg.OutputFormat != "mp3_44100_128" {
		t.Errorf("OutputFormat = %q, want %q", cfg.OutputFormat, "mp3_44100_128")
	}
	if cfg.VoiceSettings == nil {
		t.Fatal("VoiceSettings is nil")
	}
	if cfg.VoiceSettings.Stability == nil || *cfg.VoiceSettings.Stability != 0.5 {
		t.Errorf("Stability = %v, want 0.5", cfg.VoiceSettings.Stability)
	}
	if cfg.VoiceSettings.Speed == nil || *cfg.VoiceSettings.Speed != 1.0 {
		t.Errorf("Speed = %v, want 1.0", cfg.VoiceSettings.Speed)
	}
}

func TestMerge(t *testing.T) {
	dst := Default()
	stability := 0.3
	src := &Config{
		VoiceID: "test-voice",
		VoiceSettings: &VoiceSettings{
			Stability: &stability,
		},
	}

	dst.Merge(src)

	if dst.VoiceID != "test-voice" {
		t.Errorf("VoiceID = %q, want %q", dst.VoiceID, "test-voice")
	}
	if *dst.VoiceSettings.Stability != 0.3 {
		t.Errorf("Stability = %v, want 0.3", *dst.VoiceSettings.Stability)
	}
	// Speed should be unchanged
	if *dst.VoiceSettings.Speed != 1.0 {
		t.Errorf("Speed = %v, want 1.0", *dst.VoiceSettings.Speed)
	}
}

func TestMergeNil(t *testing.T) {
	dst := Default()
	original := *dst.VoiceSettings.Stability

	dst.Merge(nil)

	if *dst.VoiceSettings.Stability != original {
		t.Errorf("Stability changed after nil merge")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	original := Oratory()
	original.VoiceID = "test-voice-id"

	if err := Save(configPath, original); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.VoiceID != original.VoiceID {
		t.Errorf("VoiceID = %q, want %q", loaded.VoiceID, original.VoiceID)
	}
	if loaded.ModelID != original.ModelID {
		t.Errorf("ModelID = %q, want %q", loaded.ModelID, original.ModelID)
	}
	if loaded.OutputFormat != original.OutputFormat {
		t.Errorf("OutputFormat = %q, want %q", loaded.OutputFormat, original.OutputFormat)
	}
	if *loaded.VoiceSettings.Stability != *original.VoiceSettings.Stability {
		t.Errorf("Stability = %v, want %v", *loaded.VoiceSettings.Stability, *original.VoiceSettings.Stability)
	}
}

func TestLoadNonexistent(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Load() expected error for nonexistent file")
	}
}

func TestSaveContainsComments(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	cfg := Default()
	if err := Save(configPath, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)
	if !contains(content, "# ElevenLabs TTS Configuration") {
		t.Error("Config file missing header comment")
	}
	if !contains(content, "# Presets:") {
		t.Error("Config file missing presets comment")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestToElevenLabsSettings(t *testing.T) {
	stability := 0.4
	similarity := 0.8
	style := 0.3
	speed := 0.95

	vs := &VoiceSettings{
		Stability:       &stability,
		SimilarityBoost: &similarity,
		Style:           &style,
		Speed:           &speed,
	}

	settings := vs.ToElevenLabsSettings()

	if settings.Stability != 0.4 {
		t.Errorf("Stability = %v, want 0.4", settings.Stability)
	}
	if settings.SimilarityBoost != 0.8 {
		t.Errorf("SimilarityBoost = %v, want 0.8", settings.SimilarityBoost)
	}
	if settings.Style != 0.3 {
		t.Errorf("Style = %v, want 0.3", settings.Style)
	}
	if settings.Speed != 0.95 {
		t.Errorf("Speed = %v, want 0.95", settings.Speed)
	}
}

func TestToElevenLabsSettingsNil(t *testing.T) {
	var vs *VoiceSettings
	settings := vs.ToElevenLabsSettings()

	if settings == nil {
		t.Fatal("ToElevenLabsSettings() returned nil for nil VoiceSettings")
	}
	// Should return default settings
	if settings.Stability != 0.5 {
		t.Errorf("Default Stability = %v, want 0.5", settings.Stability)
	}
}
