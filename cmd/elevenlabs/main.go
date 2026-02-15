// Command elevenlabs is the CLI for ElevenLabs text-to-speech services.
package main

import (
	"os"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
