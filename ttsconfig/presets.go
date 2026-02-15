package ttsconfig

// Oratory returns a config preset optimized for oratory/speech delivery.
// Lower stability (0.4) for more expressive delivery,
// moderate style (0.3) for dramatic emphasis,
// and slightly slower pace (0.95) for gravitas.
func Oratory() *Config {
	stability := 0.4
	similarity := 0.75
	style := 0.3
	speed := 0.95

	return &Config{
		ModelID:      "eleven_v3",
		OutputFormat: "pcm_48000",
		VoiceSettings: &VoiceSettings{
			Stability:       &stability,
			SimilarityBoost: &similarity,
			Style:           &style,
			Speed:           &speed,
		},
	}
}

// Podcast returns a config preset optimized for podcast narration.
// Balanced stability (0.5) for natural delivery,
// no style exaggeration (0.0) for conversational tone,
// and normal pace (1.0) for easy listening.
func Podcast() *Config {
	stability := 0.5
	similarity := 0.75
	style := 0.0
	speed := 1.0

	return &Config{
		ModelID:      "eleven_v3",
		OutputFormat: "mp3_44100_128",
		VoiceSettings: &VoiceSettings{
			Stability:       &stability,
			SimilarityBoost: &similarity,
			Style:           &style,
			Speed:           &speed,
		},
	}
}

// Audiobook returns a config preset optimized for audiobook narration.
// Higher stability (0.6) for consistent long-form delivery,
// subtle style (0.1) for character without exaggeration,
// and slightly slower pace (0.95) for comfortable listening.
func Audiobook() *Config {
	stability := 0.6
	similarity := 0.8
	style := 0.1
	speed := 0.95

	return &Config{
		ModelID:      "eleven_v3",
		OutputFormat: "pcm_48000",
		VoiceSettings: &VoiceSettings{
			Stability:       &stability,
			SimilarityBoost: &similarity,
			Style:           &style,
			Speed:           &speed,
		},
	}
}

// PresetNames returns the list of available preset names.
func PresetNames() []string {
	return []string{"oratory", "podcast", "audiobook"}
}

// GetPreset returns a preset config by name, or nil if not found.
func GetPreset(name string) *Config {
	switch name {
	case "oratory":
		return Oratory()
	case "podcast":
		return Podcast()
	case "audiobook":
		return Audiobook()
	default:
		return nil
	}
}
