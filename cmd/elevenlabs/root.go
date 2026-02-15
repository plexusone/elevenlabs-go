package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "elevenlabs",
	Short: "ElevenLabs text-to-speech CLI",
	Long: `A command-line interface for ElevenLabs text-to-speech services.

Environment:
  ELEVENLABS_API_KEY    Required API key for ElevenLabs

Examples:
  # Generate speech from a text file
  elevenlabs tts -voice <voice-id> input.txt

  # Generate speech from a JSON script
  elevenlabs ttsscript -lang en script.json`,
	Version: "0.1.0",
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("elevenlabs version %s\n", rootCmd.Version))
}
