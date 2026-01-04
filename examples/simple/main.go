package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	elevenlabs "github.com/agentplexus/go-elevenlabs"
	"github.com/grokify/mogo/os/osutil"
	"github.com/grokify/mogo/text/stringcase"
)

const (
	ModelEleven3 = "eleven_v3"
)

func main() {
	// Create client (uses ELEVENLABS_API_KEY env var)
	client, err := elevenlabs.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// List available voices
	voices, err := client.Voices().List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found %d voices", len(voices))

	// Find a voice to use
	var voiceID string
	for _, voice := range voices {
		if voice.Name == "Roger" || voice.Name == "Rachel" {
			voiceID = voice.VoiceID
			fmt.Printf("Using voice: %s (ID: %s)\n", voice.Name, voiceID)
			break
		}
	}

	if voiceID == "" && len(voices) > 0 {
		// Use the first available voice
		voiceID = voices[0].VoiceID
		fmt.Printf("Using voice: %s (ID: %s)\n", voices[0].Name, voiceID)
	}

	if voiceID == "" {
		log.Fatal("No voices available")
	}
	voiceID = "CwhRBWXzGAHq8TQ4Fs17" // VoiceIDRoger

	quoteText := "When vibe coding works this well, make everything as-code."

	if 1 == 0 {
		// Generate speech with voice settings preset
		audio, err := client.TextToSpeech().Generate(ctx, &elevenlabs.TTSRequest{
			VoiceID:       voiceID,
			Text:          quoteText,
			VoiceSettings: elevenlabs.VoiceSettingsForYouTube(),
		})
		if err != nil {
			log.Fatal(err)
		}

		// Save to file
		f, err := os.Create("output.mp3")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		_, err = io.Copy(f, audio.Audio)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Valid output formats: mp3_22050_32, mp3_44100_128, pcm_16000, etc.
	// Default is mp3_44100_128 if not specified
	format := "mp3_44100_128"
	//format := "mp3_44100_192"

	model := "eleven_multilingual_v2"
	req := &elevenlabs.TTSRequest{
		VoiceID:       voiceID,
		Text:          quoteText,
		ModelID:       model,
		VoiceSettings: elevenlabs.VoiceSettingsForCoursera(),
		OutputFormat:  format,
	}

	resp, err := client.TextToSpeech().Generate(context.Background(), req)
	if err != nil {
		fmt.Println("ERR_HERE")
		if apiErr := elevenlabs.ParseAPIError(err); apiErr != nil {
			fmt.Printf("Status: %d\n", apiErr.StatusCode)
			fmt.Printf("Message: %s\n", apiErr.Message)
			fmt.Printf("Detail: %s\n", apiErr.Detail)
		}
		log.Fatal(err)
	}
	/*
		// Convert PCM to WAV
		wavData, err := elevenlabs.PCMToWAV(resp.Audio, 44100)
		if err != nil {
			log.Fatal(err)
		}

		// Save as .wav file
		filename := stringcase.ToKebabCase(quoteText) + ".wav"
		err = os.WriteFile(filename, wavData, 0644)
		logutil.FatalErr(err)
	*/

	filename := stringcase.ToKebabCase(quoteText) + ".mp3"
	err = osutil.WriteFileReader(filename, resp.Audio)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Saved to %s\n", filename)
}
