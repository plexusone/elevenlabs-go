//go:build ignore

// This test makes raw HTTP requests to the ElevenLabs API
// without using the client library, useful for debugging API issues.
//
// Run with: go run raw_http.go
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
)

func main() {
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	fmt.Printf("API Key length: %d\n", len(apiKey))

	req, _ := http.NewRequest("GET", "https://api.elevenlabs.io/v1/voices", nil)
	req.Header.Set("xi-api-key", apiKey)

	dump, _ := httputil.DumpRequestOut(req, false)
	fmt.Printf("=== Request ===\n%s\n", dump)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)
	body, _ := io.ReadAll(resp.Body)
	if len(body) > 200 {
		fmt.Printf("Body: %s...\n", body[:200])
	} else {
		fmt.Printf("Body: %s\n", body)
	}
}
