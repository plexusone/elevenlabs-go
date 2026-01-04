// Package omnivoice provides OmniVoice provider implementations using the ElevenLabs API.
//
// This package implements the OmniVoice interfaces (tts.Provider, stt.Provider,
// agent.Provider) using the go-elevenlabs SDK as the underlying client.
//
// # Quick Start
//
//	import (
//	    "github.com/agentplexus/omnivoice/tts"
//	    eleventts "github.com/agentplexus/go-elevenlabs/omnivoice/tts"
//	)
//
//	// Create provider
//	provider, err := eleventts.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use with OmniVoice client
//	client := tts.NewClient(provider)
//	result, err := client.Synthesize(ctx, "Hello world", tts.SynthesisConfig{
//	    VoiceID: "21m00Tcm4TlvDq8ikWAM",
//	})
//
// # Environment Variables
//
// The providers use the ELEVENLABS_API_KEY environment variable for authentication
// by default. You can also provide the API key explicitly using WithAPIKey.
package omnivoice

// Version is the OmniVoice integration version.
const Version = "0.1.0"

// ProviderName is the name used to identify this provider in OmniVoice.
const ProviderName = "elevenlabs"
