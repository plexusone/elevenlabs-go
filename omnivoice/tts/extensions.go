package tts

import (
	"github.com/agentplexus/omnivoice/tts"
)

// Extension keys for ElevenLabs-specific TTS settings.
const (
	// ExtStyle is the style exaggeration setting (0.0-1.0).
	// Higher values amplify the original speaker's style.
	ExtStyle = "elevenlabs.style"

	// ExtSpeakerBoost enables speaker boost for clearer voice.
	ExtSpeakerBoost = "elevenlabs.speaker_boost"

	// ExtPronunciationDictionaryID specifies a pronunciation dictionary to use.
	ExtPronunciationDictionaryID = "elevenlabs.pronunciation_dictionary_id"

	// ExtPreviousText provides context from previous text for better prosody.
	ExtPreviousText = "elevenlabs.previous_text"

	// ExtNextText provides context from upcoming text for better prosody.
	ExtNextText = "elevenlabs.next_text"

	// ExtOptimizeStreamingLatency sets latency optimization level (0-4).
	// Higher values reduce latency but may affect quality.
	ExtOptimizeStreamingLatency = "elevenlabs.optimize_streaming_latency"
)

// WithStyle adds ElevenLabs style setting to config.
// Style controls style exaggeration (0.0-1.0, default 0.0).
// Higher values amplify the original speaker's style.
func WithStyle(config tts.SynthesisConfig, style float64) tts.SynthesisConfig {
	ensureExtensions(&config)
	config.Extensions[ExtStyle] = style
	return config
}

// GetStyle extracts style from config extensions.
// Returns 0.0 (default) if not set.
func GetStyle(config tts.SynthesisConfig) float64 {
	if v, ok := config.Extensions[ExtStyle].(float64); ok {
		return v
	}
	return 0.0
}

// WithSpeakerBoost enables speaker boost for clearer voice.
func WithSpeakerBoost(config tts.SynthesisConfig, enabled bool) tts.SynthesisConfig {
	ensureExtensions(&config)
	config.Extensions[ExtSpeakerBoost] = enabled
	return config
}

// GetSpeakerBoost extracts speaker boost setting from config extensions.
// Returns false (default) if not set.
func GetSpeakerBoost(config tts.SynthesisConfig) bool {
	if v, ok := config.Extensions[ExtSpeakerBoost].(bool); ok {
		return v
	}
	return false
}

// WithPronunciationDictionary sets the pronunciation dictionary ID to use.
func WithPronunciationDictionary(config tts.SynthesisConfig, dictionaryID string) tts.SynthesisConfig {
	ensureExtensions(&config)
	config.Extensions[ExtPronunciationDictionaryID] = dictionaryID
	return config
}

// GetPronunciationDictionary extracts pronunciation dictionary ID from config extensions.
// Returns empty string if not set.
func GetPronunciationDictionary(config tts.SynthesisConfig) string {
	if v, ok := config.Extensions[ExtPronunciationDictionaryID].(string); ok {
		return v
	}
	return ""
}

// WithPreviousText sets previous text context for better prosody.
func WithPreviousText(config tts.SynthesisConfig, text string) tts.SynthesisConfig {
	ensureExtensions(&config)
	config.Extensions[ExtPreviousText] = text
	return config
}

// GetPreviousText extracts previous text from config extensions.
func GetPreviousText(config tts.SynthesisConfig) string {
	if v, ok := config.Extensions[ExtPreviousText].(string); ok {
		return v
	}
	return ""
}

// WithNextText sets next text context for better prosody.
func WithNextText(config tts.SynthesisConfig, text string) tts.SynthesisConfig {
	ensureExtensions(&config)
	config.Extensions[ExtNextText] = text
	return config
}

// GetNextText extracts next text from config extensions.
func GetNextText(config tts.SynthesisConfig) string {
	if v, ok := config.Extensions[ExtNextText].(string); ok {
		return v
	}
	return ""
}

// WithOptimizeStreamingLatency sets the streaming latency optimization level.
// Level 0: Default (no optimization)
// Level 1-4: Increasing latency optimization (may affect quality)
func WithOptimizeStreamingLatency(config tts.SynthesisConfig, level int) tts.SynthesisConfig {
	ensureExtensions(&config)
	config.Extensions[ExtOptimizeStreamingLatency] = level
	return config
}

// GetOptimizeStreamingLatency extracts streaming latency optimization level.
// Returns 0 (default) if not set.
func GetOptimizeStreamingLatency(config tts.SynthesisConfig) int {
	if v, ok := config.Extensions[ExtOptimizeStreamingLatency].(int); ok {
		return v
	}
	return 0
}

// ensureExtensions initializes the Extensions map if nil.
func ensureExtensions(config *tts.SynthesisConfig) {
	if config.Extensions == nil {
		config.Extensions = make(map[string]any)
	}
}
