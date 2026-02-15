package ttsconfig

import (
	"math"
	"testing"
)

func TestEstimateCredits(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		speed        float64
		wantWords    int
		wantCredits  int
		wantDuration float64
	}{
		{
			name:         "150 words at normal speed",
			text:         generateWords(150),
			speed:        1.0,
			wantWords:    150,
			wantCredits:  1000, // 1 minute = 1000 credits
			wantDuration: 60.0, // 1 minute
		},
		{
			name:         "150 words at half speed",
			text:         generateWords(150),
			speed:        0.5,
			wantWords:    150,
			wantCredits:  2000,  // 2 minutes = 2000 credits
			wantDuration: 120.0, // 2 minutes
		},
		{
			name:         "300 words at normal speed",
			text:         generateWords(300),
			speed:        1.0,
			wantWords:    300,
			wantCredits:  2000,  // 2 minutes = 2000 credits
			wantDuration: 120.0, // 2 minutes
		},
		{
			name:         "empty text",
			text:         "",
			speed:        1.0,
			wantWords:    0,
			wantCredits:  0,
			wantDuration: 0,
		},
		{
			name:         "oratory speed (0.95)",
			text:         generateWords(150),
			speed:        0.95,
			wantWords:    150,
			wantCredits:  1052,  // ~1.05 minutes
			wantDuration: 63.16, // slightly longer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credits, duration, words := EstimateCredits(tt.text, tt.speed)

			if words != tt.wantWords {
				t.Errorf("words = %d, want %d", words, tt.wantWords)
			}
			if credits != tt.wantCredits {
				t.Errorf("credits = %d, want %d", credits, tt.wantCredits)
			}
			if math.Abs(duration-tt.wantDuration) > 0.5 {
				t.Errorf("duration = %v, want ~%v", duration, tt.wantDuration)
			}
		})
	}
}

func TestEstimate(t *testing.T) {
	text := generateWords(150)
	est := Estimate(text, 1.0)

	if est.WordCount != 150 {
		t.Errorf("WordCount = %d, want 150", est.WordCount)
	}
	if est.Credits != 1000 {
		t.Errorf("Credits = %d, want 1000", est.Credits)
	}
	if est.Speed != 1.0 {
		t.Errorf("Speed = %v, want 1.0", est.Speed)
	}
	if est.CharCount == 0 {
		t.Error("CharCount should not be 0")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds float64
		want    string
	}{
		{0, "0s"},
		{30, "30s"},
		{59, "59s"},
		{60, "1m 0s"},
		{90, "1m 30s"},
		{125, "2m 5s"},
		{3600, "60m 0s"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatDuration(tt.seconds)
			if got != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.seconds, got, tt.want)
			}
		})
	}
}

func TestCreditEstimateDuration(t *testing.T) {
	est := CreditEstimate{DurationSecs: 90}
	if est.Duration() != "1m 30s" {
		t.Errorf("Duration() = %q, want %q", est.Duration(), "1m 30s")
	}
}

func TestStripMarkup(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no markup",
			input: "Hello world",
			want:  "Hello world",
		},
		{
			name:  "SSML break tag",
			input: "Hello <break time=\"1s\"/> world",
			want:  "Hello   world",
		},
		{
			name:  "emotion tag",
			input: "[calm] Hello world",
			want:  "  Hello world",
		},
		{
			name:  "multiple tags",
			input: "[excited] Hello <break time=\"0.5s\"/> world [firm]",
			want:  "  Hello   world  ",
		},
		{
			name:  "nested angle brackets",
			input: "Hello <tag attr=\"value\"> world",
			want:  "Hello   world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripMarkup(tt.input)
			if got != tt.want {
				t.Errorf("StripMarkup(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripMarkupWordCount(t *testing.T) {
	// Real-world example with markup
	text := `[calm] <break time="1s"/>
There are moments <break time="0.15s"/>
in history <break time="0.2s"/>
when humanity TRANSFORMS.`

	stripped := StripMarkup(text)
	words := len(extractWords(stripped))

	// Should count actual words, not tags
	if words != 8 {
		t.Errorf("word count = %d, want 8", words)
	}
}

// Helper to generate N words
func generateWords(n int) string {
	if n == 0 {
		return ""
	}
	result := "word"
	for i := 1; i < n; i++ {
		result += " word"
	}
	return result
}

// Helper to extract words (same as strings.Fields)
func extractWords(s string) []string {
	var words []string
	word := ""
	for _, c := range s {
		if c == ' ' || c == '\n' || c == '\t' || c == '\r' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(c)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}
