# TTS Script Authoring

A guide to authoring multilingual TTS scripts using the `ttsscript` package.

## Why Use ttsscript?

Instead of storing raw SSML (which is engine-specific and hard to edit), author your scripts in a structured JSON format that:

- Supports **multiple languages** in a single file
- Handles **pronunciations** separately from content
- Can be **compiled to any TTS engine** format
- Is **easy to edit** and version control

## Quick Start

### 1. Create a Script JSON File

```json
{
  "title": "My Course",
  "default_voices": {
    "en": "21m00Tcm4TlvDq8ikWAM",
    "es": "EXAVITQu4vr4xnSDxMaL"
  },
  "pronunciations": {
    "API": {"en": "A P I", "es": "A P I"},
    "SDK": {"en": "S D K", "es": "S D K"}
  },
  "slides": [
    {
      "title": "Introduction",
      "segments": [
        {
          "text": {
            "en": "Welcome to the API course.",
            "es": "Bienvenidos al curso de API."
          },
          "pause_after": "500ms"
        }
      ]
    }
  ]
}
```

### 2. Load and Compile

```go
import "github.com/agentplexus/go-elevenlabs/ttsscript"

// Load script
script, err := ttsscript.LoadScript("script.json")
if err != nil {
    log.Fatal(err)
}

// Compile for English
compiler := ttsscript.NewCompiler()
segments, err := compiler.Compile(script, "en")
```

### 3. Generate Audio

```go
import elevenlabs "github.com/agentplexus/go-elevenlabs"

client, _ := elevenlabs.NewClient()
formatter := ttsscript.NewElevenLabsFormatter()
jobs := formatter.Format(segments)

for _, job := range jobs {
    audio, _ := client.TextToSpeech().Simple(ctx, job.VoiceID, job.Text)
    // Save audio file...
}
```

## Script Structure

### Top-Level Fields

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Script title |
| `description` | string | Optional description |
| `default_language` | string | Primary language code |
| `default_voices` | map | Voice IDs by language |
| `pronunciations` | map | Global pronunciation rules |
| `slides` | array | Ordered list of slides |

### Slide Fields

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Slide title (for reference) |
| `notes` | string | Speaker notes (not rendered) |
| `segments` | array | Audio segments |

### Segment Fields

| Field | Type | Description |
|-------|------|-------------|
| `text` | map | Text by language code |
| `voice` | map | Voice override by language |
| `pause_before` | string | Pause before (e.g., "500ms") |
| `pause_after` | string | Pause after (e.g., "1s") |
| `emphasis` | string | "strong", "moderate", "reduced" |
| `rate` | string | "slow", "medium", "fast", or "80%" |
| `pitch` | string | "low", "medium", "high", or "+10%" |
| `pronunciations` | map | Segment-specific pronunciations |

## Pronunciations

Pronunciations are applied automatically during compilation:

```json
{
  "pronunciations": {
    "API": {"en": "A P I", "es": "A P I"},
    "kubectl": {"en": "kube control"},
    "nginx": {"en": "engine X"}
  }
}
```

### Priority Order

1. **Compiler-level** - Added via `compiler.AddPronunciation()`
2. **Segment-level** - In `segment.pronunciations`
3. **Script-level** - In `script.pronunciations`

Higher priority overrides lower.

### Add Pronunciations at Runtime

```go
compiler := ttsscript.NewCompiler()
compiler.AddPronunciation("goroutine", "en", "go routine")
compiler.AddPronunciations("en", map[string]string{
    "API": "A P I",
    "SDK": "S D K",
})
```

## Output Formats

### ElevenLabs

```go
formatter := ttsscript.NewElevenLabsFormatter()
jobs := formatter.Format(segments)

for _, job := range jobs {
    fmt.Printf("Voice: %s\n", job.VoiceID)
    fmt.Printf("Text: %s\n", job.Text)
    fmt.Printf("Pause after: %dms\n", job.PauseAfterMs)
}
```

### SSML (Google, Amazon, Azure)

```go
formatter := ttsscript.NewSSMLFormatter()
ssml, err := formatter.FormatScript(script, "en")
// Use with Google Cloud TTS, Amazon Polly, or Azure TTS
```

Example SSML output:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<speak version="1.1" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="en">
  <!-- Slide 1: Introduction -->
  Welcome to the A P I course.
  <break time="500ms"/>
</speak>
```

## Batch Processing

### Generate Manifest

```go
config := ttsscript.NewBatchConfig("./output")
config.IncludeLanguageInFilename = true

manifest := ttsscript.GenerateManifest(jobs, config, "en")
// Returns []ManifestEntry with output filenames
```

### Group by Voice

```go
groups := formatter.GroupByVoice(jobs)
for voiceID, voiceJobs := range groups {
    // Process all jobs for this voice together
}
```

## Multilingual Workflow

### 1. Author Once

```json
{
  "slides": [{
    "segments": [{
      "text": {
        "en": "Hello world",
        "es": "Hola mundo",
        "fr": "Bonjour le monde"
      }
    }]
  }]
}
```

### 2. Compile for Each Language

```go
languages := script.Languages() // ["en", "es", "fr"]

for _, lang := range languages {
    segments, _ := compiler.Compile(script, lang)
    jobs := formatter.Format(segments)

    // Generate audio for this language
    for _, job := range jobs {
        audio, _ := client.TextToSpeech().Simple(ctx, job.VoiceID, job.Text)
        // Save with language suffix
    }
}
```

## Best Practices

1. **Version control your scripts** - JSON is easy to diff and merge
2. **Separate pronunciations** - Keep them in the script, not embedded in text
3. **Use meaningful slide titles** - They appear in comments and manifests
4. **Test with one language first** - Verify before generating all languages
5. **Use consistent pause durations** - Create a style guide for your project

## Integration with Marp

For presentations, you can embed TTS annotations in Marp comments:

```markdown
---
marp: true
---

<!--
tts: Welcome to the presentation.
pause: 500ms
-->

# Slide Title

Content here...
```

Then parse and convert to ttsscript format for audio generation.
