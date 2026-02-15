package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	elevenlabs "github.com/agentplexus/go-elevenlabs"
	"github.com/agentplexus/go-elevenlabs/ttsconfig"
	"github.com/spf13/cobra"
)

var (
	ttsVoiceID         string
	ttsModelID         string
	ttsOutputFile      string
	ttsOutputFormat    string
	ttsStability       float64
	ttsSimilarityBoost float64
	ttsStyle           float64
	ttsSpeed           float64
	ttsConfigFile      string
	ttsSaveConfig      string
	ttsPreset          string
	ttsEstimate        bool
)

var ttsCmd = &cobra.Command{
	Use:   "tts <text-file>",
	Short: "Generate speech from a text file",
	Long: `Generate speech from a text file using ElevenLabs TTS.

The text file can contain plain text or ElevenLabs-formatted text with
SSML break tags and audio emotion tags for the v3 model.

Configuration can be loaded from a YAML file with --config, or use
built-in presets with --preset (oratory, podcast, audiobook).

CLI flags override config file settings.

Examples:
  # Basic usage with voice ID
  elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko speech.txt

  # Use a config file
  elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko --config tts.yaml speech.txt

  # Use a preset (oratory, podcast, audiobook)
  elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko --preset oratory speech.txt

  # Save current settings to a config file
  elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko --preset oratory --save-config my-config.yaml speech.txt

  # High-quality PCM output (48kHz, 16-bit, mono)
  elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko -f pcm_48000 speech.txt

  # Oratory style with manual settings
  elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko --speed 0.95 --style 0.3 --stability 0.4 speech.txt`,
	Args: cobra.ExactArgs(1),
	RunE: runTTS,
}

func init() {
	// Core flags
	ttsCmd.Flags().StringVarP(&ttsVoiceID, "voice", "v", "", "Voice ID (required unless in config)")
	ttsCmd.Flags().StringVarP(&ttsModelID, "model", "m", "", "Model ID (default: eleven_v3)")
	ttsCmd.Flags().StringVarP(&ttsOutputFile, "output", "o", "", "Output file (default: input basename + .mp3/.wav)")
	ttsCmd.Flags().StringVarP(&ttsOutputFormat, "format", "f", "", "Output format (mp3_44100_128, pcm_48000, etc.)")

	// Voice settings flags
	ttsCmd.Flags().Float64Var(&ttsStability, "stability", -1, "Voice stability 0.0-1.0 (lower = more expressive)")
	ttsCmd.Flags().Float64Var(&ttsSimilarityBoost, "similarity", -1, "Similarity boost 0.0-1.0 (higher = closer to original voice)")
	ttsCmd.Flags().Float64Var(&ttsStyle, "style", -1, "Style exaggeration 0.0-1.0 (higher = more stylized)")
	ttsCmd.Flags().Float64Var(&ttsSpeed, "speed", -1, "Speech speed 0.25-4.0 (0.95 = slightly slower for gravitas)")

	// Config flags
	ttsCmd.Flags().StringVarP(&ttsConfigFile, "config", "c", "", "Load settings from YAML config file")
	ttsCmd.Flags().StringVar(&ttsSaveConfig, "save-config", "", "Save current settings to YAML config file")
	ttsCmd.Flags().StringVarP(&ttsPreset, "preset", "p", "", "Use preset: oratory, podcast, audiobook")

	// Utility flags
	ttsCmd.Flags().BoolVar(&ttsEstimate, "estimate", false, "Estimate credits without calling API")

	rootCmd.AddCommand(ttsCmd)
}

func runTTS(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	logger := slog.Default()

	// Start with defaults
	config := ttsconfig.Default()

	// Apply preset if specified
	if ttsPreset != "" {
		preset := ttsconfig.GetPreset(strings.ToLower(ttsPreset))
		if preset == nil {
			return fmt.Errorf("unknown preset: %s (valid: %v)", ttsPreset, ttsconfig.PresetNames())
		}
		config = preset
		logger.Info("using preset", "preset", ttsPreset)
	}

	// Load config file if specified (overrides preset)
	if ttsConfigFile != "" {
		loadedConfig, err := ttsconfig.Load(ttsConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
		config.Merge(loadedConfig)
		logger.Info("loaded config", "file", ttsConfigFile)
	}

	// Apply CLI flags (highest priority)
	cliConfig := buildCLIConfig()
	config.Merge(cliConfig)

	// Validate voice ID
	if config.VoiceID == "" {
		return fmt.Errorf("voice ID is required (use -v flag or set in config file)")
	}

	// Save config if requested
	if ttsSaveConfig != "" {
		if err := ttsconfig.Save(ttsSaveConfig, config); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		logger.Info("saved config", "file", ttsSaveConfig)
	}

	// Read input file
	textBytes, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}
	text := strings.TrimSpace(string(textBytes))

	if text == "" {
		return fmt.Errorf("input file is empty")
	}

	// Strip SSML tags and emotion markers for word count
	textForCount := ttsconfig.StripMarkup(text)

	// Get speed for estimation
	speed := 1.0
	if config.VoiceSettings != nil && config.VoiceSettings.Speed != nil {
		speed = *config.VoiceSettings.Speed
	}

	// Estimate credits
	est := ttsconfig.Estimate(textForCount, speed)

	// If estimate only, show estimate and exit
	if ttsEstimate {
		logger.Info("credit estimate",
			"input", inputFile,
			"words", est.WordCount,
			"characters", len(text),
			"speed", speed,
			"estimated_duration", est.Duration(),
			"estimated_credits", est.Credits)
		return nil
	}

	// Check API key
	if os.Getenv("ELEVENLABS_API_KEY") == "" {
		return fmt.Errorf("ELEVENLABS_API_KEY environment variable is required")
	}

	// Determine output file extension based on format
	outputExt := ".mp3"
	if strings.HasPrefix(config.OutputFormat, "pcm_") {
		outputExt = ".wav"
	} else if strings.HasPrefix(config.OutputFormat, "opus_") {
		outputExt = ".opus"
	} else if strings.HasPrefix(config.OutputFormat, "ulaw_") || strings.HasPrefix(config.OutputFormat, "alaw_") {
		outputExt = ".raw"
	}

	// Determine output file
	outputFile := ttsOutputFile
	if outputFile == "" {
		ext := filepath.Ext(inputFile)
		base := strings.TrimSuffix(inputFile, ext)
		outputFile = base + outputExt
	}

	// Create client
	client, err := elevenlabs.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	ctx := context.Background()

	// Build voice settings from config
	voiceSettings := config.VoiceSettings.ToElevenLabsSettings()

	logger.Info("generating speech",
		"input", inputFile,
		"output", outputFile,
		"voice", config.VoiceID,
		"model", config.ModelID,
		"format", config.OutputFormat,
		"characters", len(text),
		"stability", voiceSettings.Stability,
		"similarity", voiceSettings.SimilarityBoost,
		"style", voiceSettings.Style,
		"speed", voiceSettings.Speed)

	// Generate speech
	resp, err := client.TextToSpeech().Generate(ctx, &elevenlabs.TTSRequest{
		VoiceID:       config.VoiceID,
		Text:          text,
		ModelID:       config.ModelID,
		OutputFormat:  config.OutputFormat,
		VoiceSettings: voiceSettings,
	})
	if err != nil {
		if apiErr := elevenlabs.ParseAPIError(err); apiErr != nil {
			return fmt.Errorf("API error (status %d): %s - %s", apiErr.StatusCode, apiErr.Message, apiErr.Detail)
		}
		return fmt.Errorf("failed to generate speech: %w", err)
	}

	// Write output file
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, resp.Audio)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	logger.Info("done", "bytes", written, "file", outputFile)
	return nil
}

// buildCLIConfig builds a config from CLI flags.
func buildCLIConfig() *ttsconfig.Config {
	cfg := &ttsconfig.Config{}

	if ttsVoiceID != "" {
		cfg.VoiceID = ttsVoiceID
	}
	if ttsModelID != "" {
		cfg.ModelID = ttsModelID
	}
	if ttsOutputFormat != "" {
		cfg.OutputFormat = ttsOutputFormat
	}

	// Only set voice settings if any flag was provided
	if ttsStability >= 0 || ttsSimilarityBoost >= 0 || ttsStyle >= 0 || ttsSpeed >= 0 {
		cfg.VoiceSettings = &ttsconfig.VoiceSettings{}
		if ttsStability >= 0 {
			cfg.VoiceSettings.Stability = &ttsStability
		}
		if ttsSimilarityBoost >= 0 {
			cfg.VoiceSettings.SimilarityBoost = &ttsSimilarityBoost
		}
		if ttsStyle >= 0 {
			cfg.VoiceSettings.Style = &ttsStyle
		}
		if ttsSpeed >= 0 {
			cfg.VoiceSettings.Speed = &ttsSpeed
		}
	}

	return cfg
}
