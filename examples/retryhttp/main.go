// Example demonstrating retry middleware with ElevenLabs client.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/grokify/mogo/net/http/retryhttp"
	elevenlabs "github.com/plexusone/elevenlabs-go"
	"github.com/plexusone/elevenlabs-go/voices"
)

func main() {
	// Create a retry transport with custom settings
	retryTransport := retryhttp.NewWithOptions(
		retryhttp.WithMaxRetries(3),
		retryhttp.WithInitialBackoff(1*time.Second),
		retryhttp.WithMaxBackoff(30*time.Second),
		retryhttp.WithBackoffMultiplier(2.0),
		retryhttp.WithJitter(0.1),
		retryhttp.WithLogger(slog.Default()), // Injectable logger for internal errors
		retryhttp.WithOnRetry(func(attempt int, req *http.Request, resp *http.Response, err error, backoff time.Duration) {
			if resp != nil {
				log.Printf("Retry attempt %d: status=%d, backoff=%v", attempt, resp.StatusCode, backoff)
			} else if err != nil {
				log.Printf("Retry attempt %d: error=%v, backoff=%v", attempt, err, backoff)
			}
		}),
	)

	// Create HTTP client with retry transport
	httpClient := retryTransport.Client()

	// Create ElevenLabs client with the retry-enabled HTTP client
	client, err := elevenlabs.NewClient(
		elevenlabs.WithHTTPClient(httpClient),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Make API calls - they will automatically retry on rate limits
	audio, err := client.TextToSpeech().Simple(ctx, voices.Rachel, "Hello world!")
	if err != nil {
		log.Fatalf("TTS failed: %v", err)
	}

	// Save the audio
	f, err := os.Create("output.mp3")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, audio); err != nil {
		log.Fatalf("Failed to write audio: %v", err)
	}

	fmt.Println("Audio saved to output.mp3")
}
