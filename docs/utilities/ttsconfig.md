# TTS Config Package

The `ttsconfig` package provides configuration types and utilities for ElevenLabs TTS generation.

## Import

```go
import "github.com/agentplexus/go-elevenlabs/ttsconfig"
```

## Types

### Config

The main configuration structure for TTS settings:

```go
type Config struct {
    VoiceID       string          `yaml:"voice_id,omitempty"`
    ModelID       string          `yaml:"model_id,omitempty"`
    OutputFormat  string          `yaml:"output_format,omitempty"`
    VoiceSettings *VoiceSettings  `yaml:"voice_settings,omitempty"`
}
```

### VoiceSettings

Voice parameter configuration:

```go
type VoiceSettings struct {
    Stability       *float64 `yaml:"stability,omitempty"`        // 0.0-1.0
    SimilarityBoost *float64 `yaml:"similarity_boost,omitempty"` // 0.0-1.0
    Style           *float64 `yaml:"style,omitempty"`            // 0.0-1.0
    Speed           *float64 `yaml:"speed,omitempty"`            // 0.25-4.0
}
```

## Presets

Built-in presets for common use cases:

### Oratory

Optimized for speeches, presentations, and formal content:

```go
cfg := ttsconfig.Oratory()
// stability: 0.4 (expressive), style: 0.3 (dramatic), speed: 0.95 (gravitas)
// format: pcm_48000 (studio quality)
```

### Podcast

Optimized for conversational, natural delivery:

```go
cfg := ttsconfig.Podcast()
// stability: 0.5 (balanced), style: 0.0 (natural), speed: 1.0 (normal)
// format: mp3_44100_128 (good quality)
```

### Audiobook

Optimized for long-form narration:

```go
cfg := ttsconfig.Audiobook()
// stability: 0.6 (consistent), style: 0.1 (subtle), speed: 0.95
// format: pcm_48000 (studio quality)
```

### Preset Lookup

```go
// Get preset by name
cfg := ttsconfig.GetPreset("oratory") // returns nil if not found

// List available presets
names := ttsconfig.PresetNames() // ["oratory", "podcast", "audiobook"]
```

## Configuration Files

Load and save YAML configuration files:

```go
// Load from file
cfg, err := ttsconfig.Load("tts-config.yaml")
if err != nil {
    log.Fatal(err)
}

// Save to file (includes helpful comments)
err = ttsconfig.Save("tts-config.yaml", cfg)
```

### YAML Format

```yaml
# ElevenLabs TTS Configuration
voice_id: IT8nQhZJj9jzRwmC46Ko
model_id: eleven_v3
output_format: pcm_48000

voice_settings:
  stability: 0.4
  similarity_boost: 0.75
  style: 0.3
  speed: 0.95
```

## Merging Configurations

Merge settings from multiple sources (CLI flags override config file):

```go
// Start with defaults
config := ttsconfig.Default()

// Apply preset
config = ttsconfig.Oratory()

// Merge config file values
fileConfig, _ := ttsconfig.Load("config.yaml")
config.Merge(fileConfig)

// Merge CLI overrides
cliConfig := &ttsconfig.Config{VoiceID: "custom-voice"}
config.Merge(cliConfig)
```

## Credit Estimation

Estimate API credits without making an API call:

```go
text := "Hello, this is a test."
speed := 0.95

// Simple estimation
credits, durationSecs, wordCount := ttsconfig.EstimateCredits(text, speed)

// Full estimate with all details
est := ttsconfig.Estimate(text, speed)
fmt.Printf("Words: %d\n", est.WordCount)
fmt.Printf("Duration: %s\n", est.Duration())
fmt.Printf("Credits: %d\n", est.Credits)
```

### How Credits Are Calculated

- Base speaking rate: ~150 words per minute
- Adjusted by speed setting (lower speed = longer duration)
- Credits: ~1,000 credits per minute of audio

| Text Length | Speed | Duration | Credits |
|-------------|-------|----------|---------|
| 150 words | 1.0 | 1m 0s | 1,000 |
| 150 words | 0.95 | 1m 3s | 1,052 |
| 300 words | 1.0 | 2m 0s | 2,000 |

## Markup Stripping

Remove SSML tags and emotion markers for accurate word counting:

```go
text := `[calm] Hello <break time="1s"/> world [excited]`
clean := ttsconfig.StripMarkup(text)
// Result: "  Hello   world  " (with spaces where tags were)
```

## Converting to ElevenLabs Settings

Convert `VoiceSettings` to the ElevenLabs SDK type:

```go
cfg := ttsconfig.Oratory()
settings := cfg.VoiceSettings.ToElevenLabsSettings()

// Use with ElevenLabs client
resp, err := client.TextToSpeech().Generate(ctx, &elevenlabs.TTSRequest{
    VoiceID:       cfg.VoiceID,
    ModelID:       cfg.ModelID,
    OutputFormat:  cfg.OutputFormat,
    VoiceSettings: settings,
    Text:          "Hello, world!",
})
```

## Duration Formatting

Format seconds into human-readable duration:

```go
ttsconfig.FormatDuration(90)   // "1m 30s"
ttsconfig.FormatDuration(45)   // "45s"
ttsconfig.FormatDuration(3600) // "60m 0s"
```

## See Also

- [CLI Configuration](../cli/configuration.md) — Using config files with the CLI
- [tts Command](../cli/tts.md) — CLI usage
- [Voice Settings Presets](voicesettings.md) — SDK-level voice presets
