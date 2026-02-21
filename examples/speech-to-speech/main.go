// Example: Speech-to-Speech - Voice conversion
//
// This example demonstrates converting speech from one voice to another
// while preserving the content. Useful for voice anonymization, dubbing,
// or creating consistent voice personas.
//
// Usage:
//
//	export ELEVENLABS_API_KEY="your-api-key"
//	go run main.go <input-audio.mp3> <output-audio.mp3>
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	elevenlabs "github.com/agentplexus/go-elevenlabs"
	"github.com/grokify/mogo/log/slogutil"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <input-audio> <output-audio>")
		fmt.Println("")
		fmt.Println("Converts the input audio to a different voice.")
		fmt.Println("Example: go run main.go speaker_a.mp3 converted.mp3")
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

	// Create context with logger attached
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	ctx = slogutil.ContextWithLogger(ctx, slog.Default())

	// Create client
	client, err := elevenlabs.NewClient()
	if err != nil {
		logError(ctx, "Failed to create client", err)
		os.Exit(1)
	}

	// Get a voice to convert to
	voices, err := client.Voices().List(ctx)
	if err != nil {
		logError(ctx, "Failed to list voices", err)
		os.Exit(1)
	}
	if len(voices) == 0 {
		logError(ctx, "No voices available", nil)
		os.Exit(1)
	}

	// Pick a different voice (use second voice if available)
	targetVoice := voices[0]
	if len(voices) > 1 {
		targetVoice = voices[1]
	}
	logInfo(ctx, "Converting to voice", "name", targetVoice.Name, "id", targetVoice.VoiceID)

	// Open input file
	inputFile, err := os.Open(inputPath) //nolint:gosec // G703: Example CLI, user-specified path is expected
	if err != nil {
		logError(ctx, "Failed to open input file", err, "path", inputPath)
		os.Exit(1)
	}
	defer inputFile.Close()

	// Get file info for progress
	fileInfo, _ := inputFile.Stat()
	logInfo(ctx, "Input file", "path", inputPath, "bytes", fileInfo.Size())

	// Convert voice
	logInfo(ctx, "Converting...")
	resp, err := client.SpeechToSpeech().Convert(ctx, &elevenlabs.SpeechToSpeechRequest{
		VoiceID:       targetVoice.VoiceID,
		Audio:         inputFile,
		AudioFilename: inputPath,

		// Optional: use multilingual model for non-English
		ModelID: "eleven_english_sts_v2",

		// Optional: configure voice settings
		VoiceSettings: &elevenlabs.VoiceSettings{
			Stability:       0.5,
			SimilarityBoost: 0.8,
		},

		// Optional: clean up source audio
		RemoveBackgroundNoise: true,
	})
	if err != nil {
		logError(ctx, "Conversion failed", err)
		os.Exit(1)
	}

	// Save output
	outputFile, err := os.Create(outputPath) //nolint:gosec // G703: Example CLI, user-specified path is expected
	if err != nil {
		logError(ctx, "Failed to create output file", err, "path", outputPath)
		os.Exit(1)
	}
	defer outputFile.Close()

	written, err := io.Copy(outputFile, resp.Audio)
	if err != nil {
		logError(ctx, "Failed to write output", err)
		os.Exit(1)
	}

	logInfo(ctx, "Converted audio saved", "path", outputPath, "bytes", written)
}

// logInfo logs an info message using the logger from context.
func logInfo(ctx context.Context, msg string, args ...any) {
	slogutil.LoggerFromContext(ctx, slogutil.Null()).Info(msg, args...)
}

// logError logs an error message using the logger from context.
func logError(ctx context.Context, msg string, err error, args ...any) {
	logger := slogutil.LoggerFromContext(ctx, slogutil.Null())
	if err != nil {
		args = append([]any{"error", err}, args...)
	}
	logger.Error(msg, args...)
}

// Example: Simple one-liner conversion
//
//nolint:unused // Example function for documentation
func simpleConversion(ctx context.Context, client *elevenlabs.Client, voiceID string, audio io.Reader) {
	// Simple method for quick conversion
	output, err := client.SpeechToSpeech().Simple(ctx, voiceID, audio)
	if err != nil {
		logError(ctx, "Conversion failed", err)
		os.Exit(1)
	}

	// Save or process output
	outFile, err := os.Create("output.mp3")
	if err != nil {
		logError(ctx, "Failed to create output file", err)
		os.Exit(1)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, output); err != nil {
		logError(ctx, "Failed to write output", err)
		os.Exit(1)
	}
}

// Example: Streaming conversion for real-time use
//
//nolint:unused // Example function for documentation
func streamingConversion(ctx context.Context, client *elevenlabs.Client, voiceID string, audio io.Reader) {
	resp, err := client.SpeechToSpeech().ConvertStream(ctx, &elevenlabs.SpeechToSpeechRequest{
		VoiceID:      voiceID,
		Audio:        audio,
		OutputFormat: "pcm_22050", // PCM for real-time playback
	})
	if err != nil {
		logError(ctx, "Streaming conversion failed", err)
		os.Exit(1)
	}

	// Stream to audio player
	// audioPlayer.Write(resp.Audio)
	_ = resp
}

// Example: Using seed audio for consistent style
//
//nolint:unused // Example function for documentation
func seededConversion(ctx context.Context, client *elevenlabs.Client, voiceID string) {
	sourceFile, err := os.Open("source.mp3")
	if err != nil {
		logError(ctx, "Failed to open source file", err)
		os.Exit(1)
	}
	defer sourceFile.Close()

	seedFile, err := os.Open("seed_reference.mp3")
	if err != nil {
		logError(ctx, "Failed to open seed file", err)
		os.Exit(1)
	}
	defer seedFile.Close()

	resp, err := client.SpeechToSpeech().Convert(ctx, &elevenlabs.SpeechToSpeechRequest{
		VoiceID: voiceID,
		Audio:   sourceFile,

		// Seed audio influences the conversion style
		SeedAudio:         seedFile,
		SeedAudioFilename: "seed_reference.mp3",
	})
	if err != nil {
		logError(ctx, "Seeded conversion failed", err)
		os.Exit(1)
	}

	_ = resp
}
