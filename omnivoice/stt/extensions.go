package stt

import (
	"github.com/agentplexus/omnivoice/stt"
)

// Extension keys for ElevenLabs-specific STT settings.
const (
	// ExtTagAudioEvents enables detection of audio events like music and laughter.
	ExtTagAudioEvents = "elevenlabs.tag_audio_events"

	// ExtNumSpeakers hints the expected number of speakers for diarization.
	ExtNumSpeakers = "elevenlabs.num_speakers"

	// ExtTimestampsGranularity sets timestamp granularity ("word" or "character").
	ExtTimestampsGranularity = "elevenlabs.timestamps_granularity"
)

// WithTagAudioEvents enables detection of audio events like music and laughter.
func WithTagAudioEvents(config stt.TranscriptionConfig, enabled bool) stt.TranscriptionConfig {
	ensureExtensions(&config)
	config.Extensions[ExtTagAudioEvents] = enabled
	return config
}

// GetTagAudioEvents extracts tag audio events setting from config extensions.
func GetTagAudioEvents(config stt.TranscriptionConfig) bool {
	if v, ok := config.Extensions[ExtTagAudioEvents].(bool); ok {
		return v
	}
	return false
}

// WithNumSpeakers hints the expected number of speakers for diarization.
func WithNumSpeakers(config stt.TranscriptionConfig, numSpeakers int) stt.TranscriptionConfig {
	ensureExtensions(&config)
	config.Extensions[ExtNumSpeakers] = numSpeakers
	return config
}

// GetNumSpeakers extracts expected number of speakers from config extensions.
func GetNumSpeakers(config stt.TranscriptionConfig) int {
	if v, ok := config.Extensions[ExtNumSpeakers].(int); ok {
		return v
	}
	return 0
}

// WithTimestampsGranularity sets timestamp granularity ("word" or "character").
func WithTimestampsGranularity(config stt.TranscriptionConfig, granularity string) stt.TranscriptionConfig {
	ensureExtensions(&config)
	config.Extensions[ExtTimestampsGranularity] = granularity
	return config
}

// GetTimestampsGranularity extracts timestamp granularity from config extensions.
func GetTimestampsGranularity(config stt.TranscriptionConfig) string {
	if v, ok := config.Extensions[ExtTimestampsGranularity].(string); ok {
		return v
	}
	return "word"
}

// ensureExtensions initializes the Extensions map if nil.
func ensureExtensions(config *stt.TranscriptionConfig) {
	if config.Extensions == nil {
		config.Extensions = make(map[string]any)
	}
}
