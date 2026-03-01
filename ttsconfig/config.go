// Package ttsconfig provides configuration types and utilities for ElevenLabs TTS.
package ttsconfig

import (
	"os"

	elevenlabs "github.com/plexusone/elevenlabs-go"
	"gopkg.in/yaml.v3"
)

// Config represents the configuration for text-to-speech generation.
type Config struct {
	// Voice settings
	VoiceID string `yaml:"voice_id,omitempty"`
	ModelID string `yaml:"model_id,omitempty"`

	// Output settings
	OutputFormat string `yaml:"output_format,omitempty"`

	// Voice parameters
	VoiceSettings *VoiceSettings `yaml:"voice_settings,omitempty"`
}

// VoiceSettings represents voice parameter configuration.
type VoiceSettings struct {
	// Stability determines how stable the voice is (0.0 to 1.0).
	// Lower values introduce broader emotional range.
	Stability *float64 `yaml:"stability,omitempty"`

	// SimilarityBoost determines how closely the AI should adhere to
	// the original voice (0.0 to 1.0).
	SimilarityBoost *float64 `yaml:"similarity_boost,omitempty"`

	// Style determines the style exaggeration (0.0 to 1.0).
	// Higher values amplify the original speaker's style.
	Style *float64 `yaml:"style,omitempty"`

	// Speed adjusts the speed of the voice (0.25 to 4.0).
	// 1.0 is the default speed.
	Speed *float64 `yaml:"speed,omitempty"`
}

// Load loads a TTS configuration from a YAML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save saves a TTS configuration to a YAML file with comments.
func Save(path string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	// Prepend helpful comments
	header := `# ElevenLabs TTS Configuration
#
# voice_id: Your ElevenLabs voice ID
# model_id: Model to use (eleven_v3, eleven_multilingual_v2, eleven_turbo_v2_5)
# output_format: Audio format (pcm_48000, mp3_44100_128, mp3_44100_192, etc.)
#
# voice_settings:
#   stability: 0.0-1.0 (lower = more expressive, higher = more consistent)
#   similarity_boost: 0.0-1.0 (higher = closer to original voice)
#   style: 0.0-1.0 (higher = more stylized delivery)
#   speed: 0.25-4.0 (1.0 = normal, 0.95 = slightly slower for gravitas)
#
# Presets:
#   Oratory:   stability=0.4, style=0.3, speed=0.95
#   Podcast:   stability=0.5, style=0.0, speed=1.0
#   Audiobook: stability=0.6, style=0.1, speed=0.95

`
	return os.WriteFile(path, []byte(header+string(data)), 0600)
}

// ToElevenLabsSettings converts VoiceSettings to elevenlabs.VoiceSettings.
func (v *VoiceSettings) ToElevenLabsSettings() *elevenlabs.VoiceSettings {
	if v == nil {
		return elevenlabs.DefaultVoiceSettings()
	}

	settings := elevenlabs.DefaultVoiceSettings()

	if v.Stability != nil {
		settings.Stability = *v.Stability
	}
	if v.SimilarityBoost != nil {
		settings.SimilarityBoost = *v.SimilarityBoost
	}
	if v.Style != nil {
		settings.Style = *v.Style
	}
	if v.Speed != nil {
		settings.Speed = *v.Speed
	}

	return settings
}

// Default returns a Config with default values.
func Default() *Config {
	return &Config{
		ModelID:      "eleven_v3",
		OutputFormat: "mp3_44100_128",
		VoiceSettings: &VoiceSettings{
			Stability:       ptrFloat64(0.5),
			SimilarityBoost: ptrFloat64(0.75),
			Style:           ptrFloat64(0.0),
			Speed:           ptrFloat64(1.0),
		},
	}
}

// Merge merges non-zero values from src into dst.
// This allows CLI flags to override config file values.
func (c *Config) Merge(src *Config) {
	if src == nil {
		return
	}
	if src.VoiceID != "" {
		c.VoiceID = src.VoiceID
	}
	if src.ModelID != "" {
		c.ModelID = src.ModelID
	}
	if src.OutputFormat != "" {
		c.OutputFormat = src.OutputFormat
	}
	if src.VoiceSettings != nil {
		if c.VoiceSettings == nil {
			c.VoiceSettings = &VoiceSettings{}
		}
		if src.VoiceSettings.Stability != nil {
			c.VoiceSettings.Stability = src.VoiceSettings.Stability
		}
		if src.VoiceSettings.SimilarityBoost != nil {
			c.VoiceSettings.SimilarityBoost = src.VoiceSettings.SimilarityBoost
		}
		if src.VoiceSettings.Style != nil {
			c.VoiceSettings.Style = src.VoiceSettings.Style
		}
		if src.VoiceSettings.Speed != nil {
			c.VoiceSettings.Speed = src.VoiceSettings.Speed
		}
	}
}

func ptrFloat64(v float64) *float64 {
	return &v
}
