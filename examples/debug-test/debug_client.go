//go:build ignore

// This test uses the ElevenLabs client with a debug HTTP transport
// to show all HTTP request details.
//
// Run with: go run debug_client.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	elevenlabs "github.com/plexusone/elevenlabs-go"
)

type debugTransport struct {
	rt http.RoundTripper
}

func (d *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, _ := httputil.DumpRequestOut(req, false)
	fmt.Printf("=== Request ===\n%s\n", dump)
	return d.rt.RoundTrip(req)
}

func main() {
	fmt.Printf("API Key from env: %q (length: %d)\n", os.Getenv("ELEVENLABS_API_KEY")[:10]+"...", len(os.Getenv("ELEVENLABS_API_KEY")))

	httpClient := &http.Client{
		Transport: &debugTransport{rt: http.DefaultTransport},
	}

	client, err := elevenlabs.NewClient(elevenlabs.WithHTTPClient(httpClient))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	voices, err := client.Voices().List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found %d voices", len(voices))
}
