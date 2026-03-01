// Example: WebSocket TTS - Real-time text-to-speech streaming
//
// This example demonstrates streaming text to speech via WebSocket,
// which is ideal for LLM integration where you want to play audio
// as the response is being generated.
//
// Usage:
//
//	export ELEVENLABS_API_KEY="your-api-key"
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/grokify/mogo/log/slogutil"
	elevenlabs "github.com/plexusone/elevenlabs-go"
)

func main() {
	// Create context with logger and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = slogutil.ContextWithLogger(ctx, slog.Default())

	// Create client
	client, err := elevenlabs.NewClient()
	if err != nil {
		logError(ctx, "Failed to create client", err)
		os.Exit(1)
	}

	// Get a voice to use
	voices, err := client.Voices().List(ctx)
	if err != nil {
		logError(ctx, "Failed to list voices", err)
		os.Exit(1)
	}
	if len(voices) == 0 {
		logError(ctx, "No voices available", nil)
		os.Exit(1)
	}
	voiceID := voices[0].VoiceID
	logInfo(ctx, "Using voice", "name", voices[0].Name, "id", voiceID)

	// Connect to WebSocket TTS with low-latency settings
	conn, err := client.WebSocketTTS().Connect(ctx, voiceID, &elevenlabs.WebSocketTTSOptions{
		ModelID:                  "eleven_turbo_v2_5", // Fast model for real-time
		OutputFormat:             "mp3_44100_128",     // MP3 for easy playback
		OptimizeStreamingLatency: 3,                   // Balance latency vs quality
	})
	if err != nil {
		logError(ctx, "Failed to connect WebSocket", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Create output file
	outFile, err := os.Create("websocket_output.mp3")
	if err != nil {
		logError(ctx, "Failed to create output file", err)
		os.Exit(1)
	}
	defer outFile.Close()

	// Simulate LLM streaming output - sending text in chunks
	textChunks := []string{
		"Hello! ",
		"This is a demonstration ",
		"of real-time text-to-speech streaming. ",
		"Each chunk of text is sent ",
		"as it becomes available, ",
		"and audio is generated immediately. ",
		"This is perfect for LLM integration!",
	}

	// Start receiving audio in background
	done := make(chan struct{})
	go func() {
		defer close(done)
		totalBytes := 0
		for audio := range conn.Audio() {
			n, err := outFile.Write(audio)
			if err != nil {
				logError(ctx, "Error writing audio", err)
				return
			}
			totalBytes += n
			fmt.Printf("\rReceived %d bytes of audio...", totalBytes)
		}
		fmt.Printf("\nTotal audio received: %d bytes\n", totalBytes)
	}()

	// Monitor errors
	go func() {
		for err := range conn.Errors() {
			logError(ctx, "WebSocket error", err)
		}
	}()

	// Send text chunks with small delays (simulating LLM token generation)
	logInfo(ctx, "Sending text chunks...")
	for i, chunk := range textChunks {
		fmt.Printf("Sending chunk %d: %q\n", i+1, chunk)
		if err := conn.SendText(chunk); err != nil {
			logError(ctx, "Error sending text", err, "chunk", i+1)
			break
		}
		time.Sleep(100 * time.Millisecond) // Simulate LLM delay
	}

	// Flush to signal end of input and get remaining audio
	logInfo(ctx, "Flushing...")
	if err := conn.Flush(); err != nil {
		logError(ctx, "Error flushing", err)
	}

	// Wait for all audio to be received
	<-done

	logInfo(ctx, "Audio saved", "path", "websocket_output.mp3")
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

// Example with StreamText helper for channel-based input
//
//nolint:unused // Example function for documentation
func streamTextExample(ctx context.Context, client *elevenlabs.Client, voiceID string) {
	conn, err := client.WebSocketTTS().Connect(ctx, voiceID, nil)
	if err != nil {
		logError(ctx, "Failed to connect", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Create a channel for text input
	textStream := make(chan string)

	// Use StreamText helper - it handles flushing automatically
	audioOut, errOut := conn.StreamText(ctx, textStream)

	// Send text in goroutine
	go func() {
		defer close(textStream)
		textStream <- "Hello from "
		textStream <- "the streaming API!"
	}()

	// Receive audio
	for audio := range audioOut {
		// Process audio chunks
		_ = audio
	}

	// Check for errors
	if err := <-errOut; err != nil {
		logError(ctx, "Stream error", err)
	}
}

// Example showing word alignment timestamps
//
//nolint:unused // Example function for documentation
func alignmentExample(ctx context.Context, client *elevenlabs.Client, voiceID string) {
	conn, err := client.WebSocketTTS().Connect(ctx, voiceID, nil)
	if err != nil {
		logError(ctx, "Failed to connect", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Receive alignments in background
	go func() {
		for align := range conn.Alignments() {
			fmt.Println("Word alignments:")
			for i, char := range align.Characters {
				fmt.Printf("  %s: %.3fs - %.3fs\n",
					char,
					align.CharacterStart[i],
					align.CharacterEnd[i])
			}
		}
	}()

	if err := conn.SendText("Hello world!"); err != nil {
		logError(ctx, "Failed to send text", err)
		os.Exit(1)
	}
	if err := conn.Flush(); err != nil {
		logError(ctx, "Failed to flush", err)
		os.Exit(1)
	}

	// Drain audio
	for range conn.Audio() {
	}
}
