package omnivoice

import (
	"fmt"
	"time"

	elevenlabs "github.com/agentplexus/go-elevenlabs"
	"github.com/agentplexus/omnivoice/stt"
	"github.com/agentplexus/omnivoice/tts"
)

// Extension keys for ElevenLabs-specific TTS settings.
// These match the keys defined in the tts sub-package extensions.go.
const (
	extStyle        = "elevenlabs.style"
	extSpeakerBoost = "elevenlabs.speaker_boost"
)

// VoiceToOmniVoice converts an ElevenLabs Voice to an OmniVoice Voice.
func VoiceToOmniVoice(v *elevenlabs.Voice) tts.Voice {
	voice := tts.Voice{
		ID:       v.VoiceID,
		Name:     v.Name,
		Provider: ProviderName,
		Metadata: make(map[string]any),
	}

	// Extract labels as metadata
	for k, val := range v.Labels {
		voice.Metadata[k] = val
	}

	// Try to extract gender from labels
	if gender, ok := v.Labels["gender"]; ok {
		voice.Gender = gender
	}

	// Try to extract language from labels
	if lang, ok := v.Labels["language"]; ok {
		voice.Language = lang
	}

	// Add category to metadata
	voice.Metadata["category"] = v.Category
	voice.Metadata["description"] = v.Description
	voice.Metadata["preview_url"] = v.PreviewURL

	return voice
}

// ConfigToTTSRequest converts an OmniVoice SynthesisConfig to an ElevenLabs TTSRequest.
func ConfigToTTSRequest(text string, config tts.SynthesisConfig) *elevenlabs.TTSRequest {
	// Apply pitch adjustment via SSML if specified
	processedText := applyPitchSSML(text, config.Pitch)

	req := &elevenlabs.TTSRequest{
		VoiceID: config.VoiceID,
		Text:    processedText,
		ModelID: config.Model,
	}

	// Map output format
	if config.OutputFormat != "" {
		req.OutputFormat = mapOutputFormat(config.OutputFormat, config.SampleRate)
	}

	// Get extension values
	style := getStyleFromExtensions(config.Extensions)
	speakerBoost := getSpeakerBoostFromExtensions(config.Extensions)

	// Set voice settings if any are specified (core or extensions)
	if config.Stability > 0 || config.SimilarityBoost > 0 || config.Speed > 0 || style > 0 || speakerBoost {
		settings := elevenlabs.DefaultVoiceSettings()
		if config.Stability > 0 {
			settings.Stability = config.Stability
		}
		if config.SimilarityBoost > 0 {
			settings.SimilarityBoost = config.SimilarityBoost
		}
		if config.Speed > 0 {
			settings.Speed = config.Speed
		}
		// Apply ElevenLabs-specific extensions
		if style > 0 {
			settings.Style = style
		}
		if speakerBoost {
			settings.UseSpeakerBoost = true
		}
		req.VoiceSettings = settings
	}

	return req
}

// getStyleFromExtensions extracts style from extensions map.
func getStyleFromExtensions(extensions map[string]any) float64 {
	if extensions == nil {
		return 0.0
	}
	if v, ok := extensions[extStyle].(float64); ok {
		return v
	}
	return 0.0
}

// getSpeakerBoostFromExtensions extracts speaker boost setting from extensions map.
func getSpeakerBoostFromExtensions(extensions map[string]any) bool {
	if extensions == nil {
		return false
	}
	if v, ok := extensions[extSpeakerBoost].(bool); ok {
		return v
	}
	return false
}

// applyPitchSSML wraps text in SSML prosody tags if pitch is specified.
// OmniVoice pitch range: -1.0 to 1.0 (0 = normal)
// SSML pitch: percentage like "+20%" or "-30%"
func applyPitchSSML(text string, pitch float64) string {
	if pitch == 0 {
		return text
	}
	// Convert -1.0 to 1.0 range to percentage (-50% to +50%)
	// This provides a reasonable pitch adjustment range
	percentage := int(pitch * 50)
	if percentage > 0 {
		return fmt.Sprintf(`<speak><prosody pitch="+%d%%">%s</prosody></speak>`, percentage, text)
	}
	return fmt.Sprintf(`<speak><prosody pitch="%d%%">%s</prosody></speak>`, percentage, text)
}

// ConfigToWebSocketTTSOptions converts OmniVoice SynthesisConfig to ElevenLabs WebSocket options.
func ConfigToWebSocketTTSOptions(config tts.SynthesisConfig) *elevenlabs.WebSocketTTSOptions {
	opts := elevenlabs.DefaultWebSocketTTSOptions()

	if config.Model != "" {
		opts.ModelID = config.Model
	}

	if config.OutputFormat != "" {
		opts.OutputFormat = mapOutputFormat(config.OutputFormat, config.SampleRate)
	}

	// Get extension values
	style := getStyleFromExtensions(config.Extensions)
	speakerBoost := getSpeakerBoostFromExtensions(config.Extensions)

	// Set voice settings if any are specified (core or extensions)
	if config.Stability > 0 || config.SimilarityBoost > 0 || config.Speed > 0 || style > 0 || speakerBoost {
		settings := elevenlabs.DefaultVoiceSettings()
		if config.Stability > 0 {
			settings.Stability = config.Stability
		}
		if config.SimilarityBoost > 0 {
			settings.SimilarityBoost = config.SimilarityBoost
		}
		if config.Speed > 0 {
			settings.Speed = config.Speed
		}
		// Apply ElevenLabs-specific extensions
		if style > 0 {
			settings.Style = style
		}
		if speakerBoost {
			settings.UseSpeakerBoost = true
		}
		opts.VoiceSettings = settings
	}

	return opts
}

// mapOutputFormat maps OmniVoice format names to ElevenLabs format strings.
func mapOutputFormat(format string, sampleRate int) string {
	// If already in ElevenLabs format, return as-is
	if len(format) > 4 && (format[:4] == "mp3_" || format[:4] == "pcm_" ||
		format[:5] == "ulaw_" || format[:5] == "alaw_" || format[:5] == "opus_") {
		return format
	}

	// Default sample rate
	if sampleRate == 0 {
		sampleRate = 44100
	}

	switch format {
	case "mp3":
		switch sampleRate {
		case 22050:
			return "mp3_22050_32"
		case 44100:
			return "mp3_44100_128"
		default:
			return "mp3_44100_128"
		}
	case "pcm":
		switch sampleRate {
		case 8000:
			return "pcm_8000"
		case 16000:
			return "pcm_16000"
		case 22050:
			return "pcm_22050"
		case 24000:
			return "pcm_24000"
		case 44100:
			return "pcm_44100"
		default:
			return "pcm_16000"
		}
	case "wav":
		// ElevenLabs uses PCM for raw audio
		return "pcm_44100"
	case "opus":
		// Fallback to mp3 as ElevenLabs doesn't support opus
		return "mp3_44100_128"
	// Telephony formats - critical for Twilio/PSTN integration
	// ElevenLabs supports these natively, no conversion needed!
	case "ulaw", "mulaw", "g711u":
		return "ulaw_8000"
	case "alaw", "g711a":
		return "alaw_8000"
	default:
		return "mp3_44100_128"
	}
}

// ConfigToWebSocketSTTOptions converts OmniVoice TranscriptionConfig to ElevenLabs WebSocket options.
func ConfigToWebSocketSTTOptions(config stt.TranscriptionConfig) *elevenlabs.WebSocketSTTOptions {
	opts := elevenlabs.DefaultWebSocketSTTOptions()

	if config.Model != "" {
		opts.ModelID = config.Model
	}

	if config.Language != "" {
		opts.LanguageCode = config.Language
	}

	// Map sample rate and encoding to AudioFormat
	if config.SampleRate > 0 || config.Encoding != "" {
		opts.AudioFormat = mapAudioFormat(config.Encoding, config.SampleRate)
	}

	// Enable word timestamps if requested
	opts.IncludeTimestamps = config.EnableWordTimestamps

	return opts
}

// mapAudioFormat maps OmniVoice encoding and sample rate to ElevenLabs audio_format.
func mapAudioFormat(encoding string, sampleRate int) string {
	// Default sample rate
	if sampleRate == 0 {
		sampleRate = 16000
	}

	// Handle mulaw encoding
	if encoding == "mulaw" || encoding == "pcm_mulaw" {
		return "ulaw_8000"
	}

	// PCM formats based on sample rate
	switch sampleRate {
	case 8000:
		return "pcm_8000"
	case 16000:
		return "pcm_16000"
	case 22050:
		return "pcm_22050"
	case 24000:
		return "pcm_24000"
	case 44100:
		return "pcm_44100"
	case 48000:
		return "pcm_48000"
	default:
		return "pcm_16000"
	}
}

// TranscriptToStreamEvent converts an ElevenLabs STT transcript to an OmniVoice stream event.
func TranscriptToStreamEvent(t *elevenlabs.STTTranscript) stt.StreamEvent {
	event := stt.StreamEvent{
		Transcript: t.Text,
		IsFinal:    t.IsFinal,
		Type:       stt.EventTranscript,
	}

	// Convert words if available
	if len(t.Words) > 0 {
		// Calculate start/end time from first and last word
		var startTime, endTime float64
		if len(t.Words) > 0 {
			startTime = t.Words[0].Start
			endTime = t.Words[len(t.Words)-1].End
		}

		segment := &stt.Segment{
			Text:      t.Text,
			StartTime: time.Duration(startTime * float64(time.Second)),
			EndTime:   time.Duration(endTime * float64(time.Second)),
			Language:  t.LanguageCode,
		}

		for _, w := range t.Words {
			segment.Words = append(segment.Words, stt.Word{
				Text:      w.Text,
				StartTime: time.Duration(w.Start * float64(time.Second)),
				EndTime:   time.Duration(w.End * float64(time.Second)),
			})
		}

		event.Segment = segment
	}

	return event
}

// TranscriptionResultFromResponse converts an ElevenLabs transcription response to OmniVoice format.
func TranscriptionResultFromResponse(resp *elevenlabs.TranscriptionResponse) *stt.TranscriptionResult {
	result := &stt.TranscriptionResult{
		Text:     resp.Text,
		Language: resp.LanguageCode,
	}

	// Convert words to segment
	if len(resp.Words) > 0 {
		segment := stt.Segment{
			Text: resp.Text,
		}

		for _, w := range resp.Words {
			word := stt.Word{
				Text:       w.Text,
				StartTime:  time.Duration(w.Start * float64(time.Second)),
				EndTime:    time.Duration(w.End * float64(time.Second)),
				Confidence: w.Confidence,
				Speaker:    w.Speaker,
			}
			segment.Words = append(segment.Words, word)
		}

		// Set segment timing from first and last word
		if len(segment.Words) > 0 {
			segment.StartTime = segment.Words[0].StartTime
			segment.EndTime = segment.Words[len(segment.Words)-1].EndTime
		}

		result.Segments = append(result.Segments, segment)
	}

	// Convert utterances (speaker segments)
	for _, u := range resp.Utterances {
		segment := stt.Segment{
			Text:      u.Text,
			StartTime: time.Duration(u.Start * float64(time.Second)),
			EndTime:   time.Duration(u.End * float64(time.Second)),
			Speaker:   u.Speaker,
		}
		result.Segments = append(result.Segments, segment)
	}

	return result
}
