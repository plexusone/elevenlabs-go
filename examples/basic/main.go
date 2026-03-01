// Example basic shows how to use the ElevenLabs SDK for common operations.
//
// Run with:
//
//	ELEVENLABS_API_KEY=your-api-key go run main.go
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	elevenlabs "github.com/plexusone/elevenlabs-go"
)

func main() {
	// Create client - API key will be read from ELEVENLABS_API_KEY environment variable
	client, err := elevenlabs.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// List available voices
	fmt.Println("=== Available Voices ===")
	voices, err := client.Voices().List(ctx)
	if err != nil {
		log.Fatalf("Failed to list voices: %v", err)
	}
	for _, v := range voices {
		fmt.Printf("  %s: %s (%s)\n", v.VoiceID, v.Name, v.Category)
	}

	// List available models
	fmt.Println("\n=== Available Models ===")
	models, err := client.Models().ListTTSModels(ctx)
	if err != nil {
		log.Fatalf("Failed to list models: %v", err)
	}
	for _, m := range models {
		fmt.Printf("  %s: %s\n", m.ModelID, m.Name)
	}

	// Check subscription
	fmt.Println("\n=== Subscription ===")
	sub, err := client.User().GetSubscription(ctx)
	if err != nil {
		log.Fatalf("Failed to get subscription: %v", err)
	}
	fmt.Printf("  Tier: %s\n", sub.Tier)
	fmt.Printf("  Characters: %d / %d (remaining: %d)\n",
		sub.CharacterCount, sub.CharacterLimit, sub.CharactersRemaining())

	// List Studio Projects
	fmt.Println("\n=== Studio Projects ===")
	projects, err := client.Projects().List(ctx)
	if err != nil {
		log.Printf("  Failed to list projects: %v", err)
	} else {
		fmt.Printf("  Found %d projects\n", len(projects))
		for _, p := range projects {
			fmt.Printf("    %s: %s\n", p.ProjectID, p.Name)
		}
	}

	// List Pronunciation Dictionaries
	fmt.Println("\n=== Pronunciation Dictionaries ===")
	dictResp, err := client.Pronunciation().List(ctx, nil)
	if err != nil {
		log.Printf("  Failed to list dictionaries: %v", err)
	} else {
		fmt.Printf("  Found %d dictionaries\n", len(dictResp.Dictionaries))
		for _, d := range dictResp.Dictionaries {
			fmt.Printf("    %s: %s (%d rules)\n", d.ID, d.Name, d.RulesCount)
		}
	}

	// Generate speech (only if we have voices and characters remaining)
	if len(voices) > 0 && sub.CharactersRemaining() > 50 {
		fmt.Println("\n=== Generating Speech ===")
		voiceID := voices[0].VoiceID

		audio, err := client.TextToSpeech().Simple(ctx, voiceID, "Hello, welcome to ElevenLabs!")
		if err != nil {
			log.Fatalf("Failed to generate speech: %v", err)
		}

		// Save to file
		outFile, err := os.Create("output.mp3")
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer outFile.Close()

		n, err := io.Copy(outFile, audio)
		if err != nil {
			log.Fatalf("Failed to save audio: %v", err)
		}
		fmt.Printf("  Saved %d bytes to output.mp3\n", n)
	}

	// Generate sound effect (example - commented out to avoid unnecessary API usage)
	/*
		fmt.Println("\n=== Generating Sound Effect ===")
		sfxAudio, err := client.SoundEffects().Simple(ctx, "short beep notification")
		if err != nil {
			log.Fatalf("Failed to generate sound effect: %v", err)
		}

		sfxFile, err := os.Create("sound_effect.mp3")
		if err != nil {
			log.Fatalf("Failed to create sound effect file: %v", err)
		}
		defer sfxFile.Close()

		n, err := io.Copy(sfxFile, sfxAudio)
		if err != nil {
			log.Fatalf("Failed to save sound effect: %v", err)
		}
		fmt.Printf("  Saved %d bytes to sound_effect.mp3\n", n)
	*/

	fmt.Println("\nDone!")
}
