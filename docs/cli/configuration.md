# CLI Configuration

The `elevenlabs` CLI supports YAML configuration files for reusable TTS settings.

## Configuration File Format

```yaml
# ElevenLabs TTS Configuration

# Voice ID: Your ElevenLabs voice ID (custom or premade)
# Find yours at: https://elevenlabs.io/app/voice-library
voice_id: IT8nQhZJj9jzRwmC46Ko

# Model ID: The TTS model to use
# Options:
#   - eleven_v3: Latest model with emotion tags
#   - eleven_multilingual_v2: Default, multilingual support
#   - eleven_turbo_v2_5: Low-latency for real-time
model_id: eleven_v3

# Output Format: Audio encoding format
# Options:
#   - mp3_44100_128: Good quality MP3 (default)
#   - mp3_44100_192: High quality MP3
#   - pcm_48000: Studio quality lossless
#   - opus_48000_128: Efficient streaming
output_format: pcm_48000

# Voice Settings: Fine-tune voice characteristics
voice_settings:
  # Stability: Voice consistency (0.0 to 1.0)
  # Lower = more expressive, emotional variation
  # Higher = more consistent, predictable delivery
  stability: 0.4

  # Similarity Boost: Match to original voice (0.0 to 1.0)
  # Higher = closer to original voice character
  similarity_boost: 0.75

  # Style: Exaggerates speaker's natural style (0.0 to 1.0)
  # Higher = more dramatic delivery
  style: 0.3

  # Speed: Playback rate multiplier (0.25 to 4.0)
  # 0.95 = slightly slower (adds gravitas)
  # 1.0 = normal
  # 1.1 = slightly faster (more energetic)
  speed: 0.95
```

## Creating a Config File

### From Preset

```bash
# Save oratory preset to file
elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko \
  --preset oratory \
  --save-config oratory.yaml \
  input.txt
```

### From Custom Settings

```bash
# Save custom settings
elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko \
  --model eleven_v3 \
  --format pcm_48000 \
  --stability 0.35 \
  --style 0.25 \
  --speed 0.92 \
  --save-config custom.yaml \
  input.txt
```

## Using Config Files

```bash
# Load config file
elevenlabs tts --config tts-config.yaml input.txt

# Override specific settings
elevenlabs tts --config tts-config.yaml --speed 1.0 input.txt

# Config can include voice_id
elevenlabs tts --config full-config.yaml input.txt
```

## Priority Order

Settings are applied in this order (later overrides earlier):

1. **Built-in defaults**
2. **Preset** (`--preset`)
3. **Config file** (`--config`)
4. **CLI flags** (`--stability`, `--speed`, etc.)

## Example Configs

### Oratory/Speech

For presentations, speeches, and formal content:

```yaml
voice_id: IT8nQhZJj9jzRwmC46Ko
model_id: eleven_v3
output_format: pcm_48000
voice_settings:
  stability: 0.4
  similarity_boost: 0.75
  style: 0.3
  speed: 0.95
```

### Podcast

For conversational, natural delivery:

```yaml
voice_id: IT8nQhZJj9jzRwmC46Ko
model_id: eleven_v3
output_format: mp3_44100_128
voice_settings:
  stability: 0.5
  similarity_boost: 0.75
  style: 0.0
  speed: 1.0
```

### Audiobook

For long-form narration:

```yaml
voice_id: IT8nQhZJj9jzRwmC46Ko
model_id: eleven_v3
output_format: pcm_48000
voice_settings:
  stability: 0.6
  similarity_boost: 0.8
  style: 0.1
  speed: 0.95
```

### Energetic (TikTok/Reels)

For short-form, high-energy content:

```yaml
voice_id: IT8nQhZJj9jzRwmC46Ko
model_id: eleven_v3
output_format: mp3_44100_128
voice_settings:
  stability: 0.4
  similarity_boost: 0.7
  style: 0.5
  speed: 1.15
```

## Programmatic Usage

The config structs can be used programmatically:

```go
import "github.com/agentplexus/go-elevenlabs/cmd/elevenlabs"

// Load config
config, err := main.LoadTTSConfig("tts-config.yaml")

// Use preset
config := main.NewOratoryConfig()
config.VoiceID = "IT8nQhZJj9jzRwmC46Ko"

// Convert to VoiceSettings
settings := config.VoiceSettings.ToVoiceSettings()

// Save config
err := main.SaveTTSConfig("output.yaml", config)
```

## See Also

- [tts Command](tts.md) — Full tts command reference
- [Voice Settings](../utilities/voicesettings.md) — Programmatic presets
