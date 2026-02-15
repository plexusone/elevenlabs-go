package ttsconfig

import (
	"fmt"
	"strings"
)

// CreditEstimate contains the results of a credit estimation.
type CreditEstimate struct {
	// WordCount is the number of words in the text (excluding tags).
	WordCount int

	// CharCount is the total character count of the text.
	CharCount int

	// DurationSecs is the estimated audio duration in seconds.
	DurationSecs float64

	// Credits is the estimated number of ElevenLabs credits.
	Credits int

	// Speed is the speed multiplier used for estimation.
	Speed float64
}

// EstimateCredits estimates the number of ElevenLabs credits for text.
// ElevenLabs uses ~1,000 credits per minute of audio output.
// Returns estimated credits, duration in seconds, and word count.
func EstimateCredits(text string, speed float64) (credits int, durationSecs float64, wordCount int) {
	// Count words (split on whitespace)
	words := strings.Fields(text)
	wordCount = len(words)

	// Base speaking rate: ~150 words per minute for natural speech
	// ElevenLabs default is around 150-160 wpm
	baseWPM := 150.0

	// Adjust for speed setting (lower speed = longer duration)
	effectiveWPM := baseWPM * speed

	// Calculate duration in minutes, then convert to seconds
	durationMins := float64(wordCount) / effectiveWPM
	durationSecs = durationMins * 60

	// Credits: ~1,000 credits per minute of audio
	// 30,000 credits ≈ 30 min, 100,000 credits ≈ 100 min
	credits = int(durationMins * 1000)

	return credits, durationSecs, wordCount
}

// Estimate returns a full CreditEstimate for the given text and speed.
func Estimate(text string, speed float64) CreditEstimate {
	credits, durationSecs, wordCount := EstimateCredits(text, speed)
	return CreditEstimate{
		WordCount:    wordCount,
		CharCount:    len(text),
		DurationSecs: durationSecs,
		Credits:      credits,
		Speed:        speed,
	}
}

// FormatDuration formats seconds into a human-readable duration string.
func FormatDuration(seconds float64) string {
	mins := int(seconds) / 60
	secs := int(seconds) % 60
	if mins > 0 {
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

// Duration returns the formatted duration string.
func (e CreditEstimate) Duration() string {
	return FormatDuration(e.DurationSecs)
}

// StripMarkup removes SSML tags and emotion markers from text for word counting.
// This removes <break.../> tags and [emotion] markers.
func StripMarkup(text string) string {
	result := text

	// Remove <break.../> tags
	for strings.Contains(result, "<") {
		start := strings.Index(result, "<")
		end := strings.Index(result, ">")
		if start != -1 && end != -1 && end > start {
			result = result[:start] + " " + result[end+1:]
		} else {
			break
		}
	}

	// Remove [emotion] tags
	for strings.Contains(result, "[") {
		start := strings.Index(result, "[")
		end := strings.Index(result, "]")
		if start != -1 && end != -1 && end > start {
			result = result[:start] + " " + result[end+1:]
		} else {
			break
		}
	}

	return result
}
