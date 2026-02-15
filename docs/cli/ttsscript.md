# ttsscript Command

Generate TTS audio from a JSON script file.

## Synopsis

```bash
elevenlabs ttsscript [flags] <script.json>
```

## Description

The `ttsscript` command generates speech from a structured JSON script file. It supports:

- Multiple slides with segments
- Voice assignments per segment
- Pause timing control
- Multilingual content
- Per-slide audio concatenation (requires ffmpeg)
- Manifest file generation

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--lang` | `-l` | `en` | Language code to generate |
| `--output` | `-o` | `./output` | Output directory |
| `--model` | `-m` | `eleven_multilingual_v2` | ElevenLabs model ID |
| `--per-slide` | | `false` | Concatenate segments into per-slide files |
| `--manifest` | | `true` | Generate manifest JSON file |
| `--dry-run` | | `false` | Show what would be generated |

## Script Format

```json
{
  "title": "My Presentation",
  "voices": {
    "narrator": "IT8nQhZJj9jzRwmC46Ko",
    "expert": "CwhRBWXzGAHq8TQ4Fs17"
  },
  "slides": [
    {
      "title": {
        "voice": "narrator",
        "text": {
          "en": "Introduction",
          "es": "Introducción"
        }
      },
      "segments": [
        {
          "voice": "narrator",
          "text": {
            "en": "Welcome to our presentation.",
            "es": "Bienvenidos a nuestra presentación."
          },
          "pause_after_ms": 500
        },
        {
          "voice": "expert",
          "text": {
            "en": "Let me explain the details.",
            "es": "Permítanme explicar los detalles."
          }
        }
      ]
    }
  ]
}
```

## Examples

### Basic Usage

```bash
# Generate English audio
elevenlabs ttsscript script.json

# Generate Spanish audio
elevenlabs ttsscript -l es script.json
```

### Output Options

```bash
# Custom output directory
elevenlabs ttsscript -o ./audio script.json

# Generate per-slide concatenated files
elevenlabs ttsscript --per-slide script.json

# Dry run to preview
elevenlabs ttsscript --dry-run script.json
```

### Different Model

```bash
# Use v3 model
elevenlabs ttsscript -m eleven_v3 script.json
```

## Output Structure

```
output/
├── slide01_seg00_title_en.mp3
├── slide01_seg01_en.mp3
├── slide01_seg02_en.mp3
├── slide02_seg00_title_en.mp3
├── slide02_seg01_en.mp3
├── manifest_en.json
└── (with --per-slide)
    ├── slide01_en.mp3
    └── slide02_en.mp3
```

## Manifest File

The manifest JSON contains metadata for each generated audio file:

```json
[
  {
    "slide_index": 0,
    "segment_index": -1,
    "is_title_segment": true,
    "output_file": "output/slide01_seg00_title_en.mp3",
    "text": "Introduction",
    "voice_id": "IT8nQhZJj9jzRwmC46Ko",
    "pause_before_ms": 0,
    "pause_after_ms": 0
  }
]
```

## Requirements

- `ELEVENLABS_API_KEY` environment variable
- `ffmpeg` (only if using `--per-slide`)

## See Also

- [TTS Script Authoring](../guides/ttsscript.md) — Script format guide
- [tts](tts.md) — Simple text file TTS
