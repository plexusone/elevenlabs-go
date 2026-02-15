# tts Command

Generate speech from a text file using ElevenLabs TTS.

## Synopsis

```bash
elevenlabs tts [flags] <text-file>
```

## Description

The `tts` command converts a text file to speech using the ElevenLabs API. It supports:

- Plain text input
- SSML `<break>` tags for pauses
- Emotion tags for the v3 model (`[calm]`, `[excited]`, etc.)
- CAPITALIZED words for emphasis
- YAML configuration files for reusable settings
- Built-in presets for common use cases

## Flags

### Required

| Flag | Short | Description |
|------|-------|-------------|
| `--voice` | `-v` | Voice ID (required unless specified in config) |

### Output

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `<input>.mp3` | Output file path |
| `--format` | `-f` | `mp3_44100_128` | Audio format |
| `--model` | `-m` | `eleven_v3` | Model ID |

### Voice Settings

| Flag | Range | Default | Description |
|------|-------|---------|-------------|
| `--stability` | 0.0-1.0 | 0.5 | Voice consistency (lower = more expressive) |
| `--similarity` | 0.0-1.0 | 0.75 | Adherence to original voice |
| `--style` | 0.0-1.0 | 0.0 | Style exaggeration |
| `--speed` | 0.25-4.0 | 1.0 | Speech rate multiplier |

### Configuration

| Flag | Short | Description |
|------|-------|-------------|
| `--config` | `-c` | Load settings from YAML config file |
| `--preset` | `-p` | Use preset: `oratory`, `podcast`, `audiobook` |
| `--save-config` | | Save current settings to YAML file |

### Utility

| Flag | Description |
|------|-------------|
| `--estimate` | Estimate credits without calling API |

## Output Formats

| Format | Quality | Use Case |
|--------|---------|----------|
| `mp3_22050_32` | Low | Small files, previews |
| `mp3_44100_128` | Good | General use (default) |
| `mp3_44100_192` | High | High-quality MP3 |
| `pcm_44100` | Lossless | CD quality |
| `pcm_48000` | Lossless | Studio quality (recommended) |
| `opus_48000_128` | Efficient | Streaming, web |
| `ulaw_8000` | Telephony | Phone systems |

## Presets

### Oratory

Optimized for speeches, presentations, and formal content:

```yaml
stability: 0.4      # More expressive
style: 0.3          # Moderate dramatic emphasis
speed: 0.95         # Slightly slower for gravitas
format: pcm_48000   # Studio quality
```

### Podcast

Optimized for conversational, natural delivery:

```yaml
stability: 0.5      # Balanced
style: 0.0          # Natural, no exaggeration
speed: 1.0          # Normal pace
format: mp3_44100_128
```

### Audiobook

Optimized for long-form narration:

```yaml
stability: 0.6      # Consistent for long content
style: 0.1          # Subtle character
speed: 0.95         # Comfortable listening pace
format: pcm_48000   # Studio quality
```

## Examples

### Basic Usage

```bash
# Simple text-to-speech
elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko speech.txt

# Specify output file
elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko -o output.mp3 speech.txt
```

### Using Presets

```bash
# Oratory preset for speeches
elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko --preset oratory speech.txt

# Podcast preset
elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko --preset podcast episode.txt
```

### Custom Voice Settings

```bash
# More expressive with slower pace
elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko \
  --stability 0.3 \
  --style 0.4 \
  --speed 0.9 \
  speech.txt
```

### Configuration Files

```bash
# Use config file
elevenlabs tts --config tts-config.yaml speech.txt

# Save current settings to config
elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko --preset oratory \
  --save-config my-config.yaml speech.txt

# Config overrides preset, CLI flags override config
elevenlabs tts --preset podcast --config custom.yaml --speed 1.1 speech.txt
```

### High-Quality Output

```bash
# Studio quality PCM (48kHz, 16-bit, mono)
elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko -f pcm_48000 speech.txt
# Output: speech.wav
```

### Credit Estimation

```bash
# Estimate credits before generating (no API call)
elevenlabs tts -v IT8nQhZJj9jzRwmC46Ko --estimate speech.txt

# With config file
elevenlabs tts --config tts-config.yaml --estimate speech.txt

# Output:
# INFO credit estimate input=speech.txt words=453 characters=4600
#      speed=0.95 estimated_duration="3m 10s" estimated_credits=3178
```

The estimator calculates:

- **Words**: Excluding SSML `<break>` tags and `[emotion]` markers
- **Duration**: Based on ~150 WPM, adjusted for speed setting
- **Credits**: ~1,000 credits per minute of audio

## Input File Format

### Plain Text

```
Hello, this is a simple text file.
It will be converted to speech.
```

### With SSML Breaks

```
Hello. <break time="0.5s"/> This is a pause.
<break time="1s"/>
A longer pause before this sentence.
```

### With Emotion Tags (v3 model)

```
[calm] Welcome to our presentation.
[excited] This is amazing news!
[firm] We must act now.
```

### With Emphasis (Capitalization)

```
This is VERY important.
We are NO LONGER alone in intelligence.
```

### Combined Example

```
[calm] <break time="1s"/>
There are moments <break time="0.15s"/>
in history <break time="0.2s"/>
when humanity does not merely advance —
it TRANSFORMS.
<break time="1s"/>

[serious] And so we must ask:
What makes us INDISPENSABLE?
```

## Priority Order

Settings are applied in this order (later overrides earlier):

1. **Defaults** — Built-in default values
2. **Preset** — `--preset` flag values
3. **Config file** — `--config` file values
4. **CLI flags** — Individual flag values

## See Also

- [Configuration](configuration.md) — YAML config file format
- [ttsscript](ttsscript.md) — JSON script-based TTS
- [Voice Settings](../utilities/voicesettings.md) — Programmatic presets
