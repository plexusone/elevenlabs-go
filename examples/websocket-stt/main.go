// Example: WebSocket STT - Real-time speech-to-text streaming
//
// This example demonstrates real-time audio transcription via WebSocket,
// which provides partial results for responsive UIs and word-level timing.
//
// Usage:
//
//	export ELEVENLABS_API_KEY="your-api-key"
//	go run main.go <audio-file.wav>
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
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <audio-file>")
		fmt.Println("")
		fmt.Println("The audio file should be PCM WAV format, 16kHz sample rate.")
		fmt.Println("Example: go run main.go recording.wav")
		os.Exit(1)
	}

	audioPath := os.Args[1]

	// Create context with logger and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	ctx = slogutil.ContextWithLogger(ctx, slog.Default())

	// Create client
	client, err := elevenlabs.NewClient()
	if err != nil {
		logError(ctx, "Failed to create client", err)
		os.Exit(1)
	}

	// Connect to WebSocket STT with scribe_v2_realtime
	conn, err := client.WebSocketSTT().Connect(ctx, &elevenlabs.WebSocketSTTOptions{
		ModelID:           "scribe_v2_realtime",
		AudioFormat:       "pcm_16000", // 16kHz PCM audio
		IncludeTimestamps: true,        // Get word-level timing
		CommitStrategy:    "manual",    // Manual commit for control
	})
	if err != nil {
		logError(ctx, "Failed to connect WebSocket", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Open audio file
	audioFile, err := os.Open(audioPath) //nolint:gosec // G703: Example CLI, user-specified path is expected
	if err != nil {
		logError(ctx, "Failed to open audio file", err, "path", audioPath)
		os.Exit(1)
	}
	defer audioFile.Close()

	// Skip WAV header (44 bytes) if present
	header := make([]byte, 44)
	n, err := audioFile.Read(header)
	if err != nil && err != io.EOF {
		logError(ctx, "Failed to read file header", err)
		os.Exit(1)
	}
	if n >= 4 && string(header[0:4]) != "RIFF" {
		// Not a WAV file, seek back to start
		if _, err := audioFile.Seek(0, 0); err != nil {
			logError(ctx, "Failed to seek file", err)
			os.Exit(1)
		}
	}

	// Start receiving transcripts in background
	done := make(chan struct{})
	go func() {
		defer close(done)
		for transcript := range conn.Transcripts() {
			if transcript.IsFinal {
				fmt.Printf("\n[FINAL] %s\n", transcript.Text)
				if len(transcript.Words) > 0 {
					fmt.Println("  Word timing:")
					for _, word := range transcript.Words {
						fmt.Printf("    '%s': %.2fs - %.2fs\n",
							word.Text, word.Start, word.End)
					}
				}
				if transcript.LanguageCode != "" {
					fmt.Printf("  Language: %s\n", transcript.LanguageCode)
				}
			} else {
				// Partial result - update in place
				fmt.Printf("\r[...] %s", transcript.Text)
			}
		}
	}()

	// Monitor errors
	go func() {
		for err := range conn.Errors() {
			logError(ctx, "WebSocket error", err)
		}
	}()

	// Stream audio in chunks (simulating real-time capture)
	logInfo(ctx, "Streaming audio", "path", audioPath)
	fmt.Println("---")

	chunkSize := 3200 // 100ms of 16kHz 16-bit audio
	buffer := make([]byte, chunkSize)

	for {
		n, err := audioFile.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			logError(ctx, "Error reading audio", err)
			break
		}

		if err := conn.SendAudio(buffer[:n]); err != nil {
			logError(ctx, "Error sending audio", err)
			break
		}

		// Small delay to simulate real-time streaming
		time.Sleep(50 * time.Millisecond)
	}

	// Signal end of audio stream by committing the final transcript
	fmt.Println("\n---")
	logInfo(ctx, "End of audio, committing final transcript...")
	if err := conn.Commit(); err != nil {
		logError(ctx, "Error committing transcript", err)
	}

	// Wait for all transcripts
	<-done

	logInfo(ctx, "Transcription complete")
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

// Example using StreamAudio helper with channels
//
//nolint:unused // Example function for documentation
func streamAudioExample(ctx context.Context, client *elevenlabs.Client) {
	conn, err := client.WebSocketSTT().Connect(ctx, &elevenlabs.WebSocketSTTOptions{
		AudioFormat:       "pcm_16000",
		IncludeTimestamps: true,
	})
	if err != nil {
		logError(ctx, "Failed to connect", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Create audio input channel
	audioStream := make(chan []byte)

	// Use StreamAudio helper - handles Commit automatically
	transcriptOut, errOut := conn.StreamAudio(ctx, audioStream)

	// Simulate audio capture in goroutine
	go func() {
		defer close(audioStream)
		// In real use, this would read from microphone
		// For demo, we just send empty audio
		for i := 0; i < 10; i++ {
			audioStream <- make([]byte, 3200)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Process transcripts
	for transcript := range transcriptOut {
		if transcript.IsFinal {
			logInfo(ctx, "Final transcript", "text", transcript.Text)
		}
	}

	// Check for errors
	if err := <-errOut; err != nil {
		logError(ctx, "Stream error", err)
	}
}
