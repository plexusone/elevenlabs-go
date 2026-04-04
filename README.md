# ElevenLabs Go SDK

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/elevenlabs-go/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/elevenlabs-go/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/elevenlabs-go/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/elevenlabs-go/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/elevenlabs-go/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/elevenlabs-go/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/elevenlabs-go
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/elevenlabs-go
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/elevenlabs-go
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/elevenlabs-go
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Felevenlabs-go
 [loc-svg]: https://tokei.rs/b1/github/plexusone/elevenlabs-go
 [repo-url]: https://github.com/plexusone/elevenlabs-go
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/elevenlabs-go/blob/master/LICENSE

Go SDK for the [ElevenLabs API](https://elevenlabs.io/).

## Features

- 🗣️ **Text-to-Speech**: Convert text to realistic speech with multiple voices and models
- 📝 **Speech-to-Text**: Transcribe audio with speaker diarization support
- 🎙️ **Speech-to-Speech**: Voice conversion - transform speech to a different voice
- 🔊 **Sound Effects**: Generate sound effects from text descriptions
- 🎨 **Voice Design**: Create custom AI voices with specific characteristics
- 🎵 **Music Composition**: Generate music from text prompts
- 🎙️ **Audio Isolation**: Extract vocals/speech from audio
- ⏱️ **Forced Alignment**: Get word-level timestamps for audio
- 💬 **Text-to-Dialogue**: Generate multi-speaker conversations
- 🌍 **Dubbing**: Translate and dub video/audio content
- 📚 **Projects**: Manage long-form audio content (audiobooks, podcasts)
- 📖 **Pronunciation Dictionaries**: Control pronunciation of specific terms

### Real-Time Services

- ⚡ **WebSocket TTS**: Low-latency text-to-speech streaming for real-time voice synthesis
- ⚡ **WebSocket STT**: Real-time speech-to-text with partial results
- 📞 **Twilio Integration**: Phone call integration for conversational AI agents
- 📱 **Phone Numbers**: Manage phone numbers for voice agents

### Command Line Interface

- 🖥️ **`elevenlabs tts`**: Generate speech from text files with YAML config support
- 📜 **`elevenlabs ttsscript`**: Batch TTS from JSON scripts with per-slide output
- 🎛️ **Presets**: Built-in configurations for oratory, podcast, audiobook styles

### OmniVoice Integration

- 🔌 **[OmniVoice](https://github.com/plexusone/omnivoice-core) Providers**: Use ElevenLabs as a drop-in backend for the vendor-agnostic OmniVoice interface
- 🔄 **Portable Code**: Swap voice providers (ElevenLabs, OpenAI, Google) without changing application logic
- 🧪 **TTS, STT, Agent**: Full provider implementations for text-to-speech, speech-to-text, and voice agents

### Agent Experience (AX)

- 🤖 **Machine-Readable Errors**: Error codes (`DOCUMENT_NOT_FOUND`, `NOT_LOGGED_IN`) for programmatic handling
- 🔄 **Automatic Retry**: TTS provider retries transient errors (429, 500) with exponential backoff
- 📊 **Error Classification**: 8 categories (auth, validation, rate_limit, etc.) for smart error handling
- ✅ **Pre-flight Validation**: Check required fields before making API calls
- 🔧 **Retry Policies**: Know which operations are safe to retry automatically

## Installation

```bash
go get github.com/plexusone/elevenlabs-go
```

### CLI Installation

```bash
go install github.com/plexusone/elevenlabs-go/cmd/elevenlabs@latest
```

## Quick Start

### Basic Text-to-Speech

```go
package main

import (
    "context"
    "io"
    "log"
    "os"

    elevenlabs "github.com/plexusone/elevenlabs-go"
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

    // Generate speech
    if len(voices) > 0 {
        audio, err := client.TextToSpeech().Simple(ctx,
            voices[0].VoiceID,
            "Hello from the ElevenLabs Go SDK!")
        if err != nil {
            log.Fatal(err)
        }

        // Save to file
        f, _ := os.Create("hello.mp3")
        defer f.Close()
        io.Copy(f, audio)
    }
}
```

### With Custom Options

```go
client, err := elevenlabs.NewClient(
    elevenlabs.WithAPIKey("your-api-key"),
    elevenlabs.WithTimeout(5 * time.Minute),
)
```

## Services

### Text-to-Speech

```go
// Simple generation
audio, err := client.TextToSpeech().Simple(ctx, voiceID, "Hello world")

// With full options
resp, err := client.TextToSpeech().Generate(ctx, &elevenlabs.TTSRequest{
    VoiceID: "21m00Tcm4TlvDq8ikWAM",
    Text:    "Hello with custom settings!",
    ModelID: "eleven_multilingual_v2",
    VoiceSettings: &elevenlabs.VoiceSettings{
        Stability:       0.6,
        SimilarityBoost: 0.8,
        Style:           0.1,
        SpeakerBoost:    true,
    },
    OutputFormat: "mp3_44100_192",
})
```

### Speech-to-Text

```go
// Transcribe from URL
result, err := client.SpeechToText().TranscribeURL(ctx, "https://example.com/audio.mp3")
fmt.Printf("Text: %s\n", result.Text)
fmt.Printf("Language: %s\n", result.LanguageCode)

// With speaker diarization
result, err := client.SpeechToText().TranscribeWithDiarization(ctx, audioURL)
for _, word := range result.Words {
    fmt.Printf("[%s] %s (%.2fs - %.2fs)\n", word.Speaker, word.Text, word.Start, word.End)
}
```

### Sound Effects

```go
// Simple sound effect
audio, err := client.SoundEffects().Simple(ctx, "thunder and rain storm")

// With options
sfx, err := client.SoundEffects().Generate(ctx, &elevenlabs.SoundEffectRequest{
    Text:            "spaceship engine humming",
    DurationSeconds: 10,
    PromptInfluence: 0.5,
})
```

### Music Composition

```go
// Generate music from prompt
resp, err := client.Music().Generate(ctx, &elevenlabs.MusicRequest{
    Prompt:     "upbeat electronic music for a tech video",
    DurationMs: 30000,
})

// Instrumental only
audio, err := client.Music().GenerateInstrumental(ctx, "calm piano melody", 60000)

// Generate with composition plan for fine-grained control
plan, _ := client.Music().GeneratePlan(ctx, &elevenlabs.CompositionPlanRequest{
    Prompt:     "pop song about summer",
    DurationMs: 180000,
})
resp, err := client.Music().GenerateDetailed(ctx, &elevenlabs.MusicDetailedRequest{
    CompositionPlan: plan,
})

// Separate stems (vocals, drums, bass, etc.)
f, _ := os.Open("song.mp3")
stems, err := client.Music().SeparateStems(ctx, &elevenlabs.StemSeparationRequest{
    File:     f,
    Filename: "song.mp3",
})
```

### Audio Isolation

```go
// Extract vocals from audio file
f, _ := os.Open("mixed_audio.mp3")
isolated, err := client.AudioIsolation().IsolateFile(ctx, f, "mixed_audio.mp3")
```

### Forced Alignment

```go
// Get word-level timestamps
f, _ := os.Open("speech.mp3")
result, err := client.ForcedAlignment().AlignFile(ctx, f, "speech.mp3",
    "The text that was spoken in the audio")

for _, word := range result.Words {
    fmt.Printf("%s: %.2fs - %.2fs\n", word.Text, word.Start, word.End)
}
```

### Text-to-Dialogue

```go
// Generate multi-speaker dialogue
audio, err := client.TextToDialogue().Simple(ctx, []elevenlabs.DialogueInput{
    {Text: "Hello, how are you?", VoiceID: "voice1"},
    {Text: "I'm doing great, thanks!", VoiceID: "voice2"},
})
```

### Voice Design

```go
// Generate a custom voice
resp, err := client.VoiceDesign().GeneratePreview(ctx, &elevenlabs.VoiceDesignRequest{
    Gender:         elevenlabs.VoiceGenderFemale,
    Age:            elevenlabs.VoiceAgeYoung,
    Accent:         elevenlabs.VoiceAccentAmerican,
    AccentStrength: 1.0,
    Text:           "This is a preview of the generated voice. It should be at least one hundred characters long for best results.",
})
```

### Pronunciation Dictionaries

```go
// Create from a map
dict, err := client.Pronunciation().CreateFromMap(ctx, "Tech Terms", map[string]string{
    "API":     "A P I",
    "kubectl": "kube control",
    "nginx":   "engine X",
})

// Create from JSON file
dict, err := client.Pronunciation().CreateFromJSON(ctx, "Terms", "pronunciation.json")
```

### Dubbing

```go
// Create dubbing job
dub, err := client.Dubbing().Create(ctx, &elevenlabs.DubbingRequest{
    SourceURL:      "https://example.com/video.mp4",
    TargetLanguage: "es",
    Name:           "Video - Spanish",
})

// Check status
status, err := client.Dubbing().GetStatus(ctx, dub.DubbingID)
```

### Projects (Studio)

```go
// Create a project for long-form content
project, err := client.Projects().Create(ctx, &elevenlabs.CreateProjectRequest{
    Name:                    "My Audiobook",
    DefaultModelID:          "eleven_multilingual_v2",
    DefaultParagraphVoiceID: voiceID,
})

// Convert to audio
err = client.Projects().Convert(ctx, project.ProjectID)
```

### Speech-to-Speech (Voice Conversion)

```go
// Convert speech from one voice to another
f, _ := os.Open("input.mp3")
resp, err := client.SpeechToSpeech().Convert(ctx, &elevenlabs.SpeechToSpeechRequest{
    VoiceID: targetVoiceID,
    Audio:   f,
})

// Simple conversion
output, err := client.SpeechToSpeech().Simple(ctx, targetVoiceID, audioReader)
```

### WebSocket TTS (Real-Time Streaming)

```go
// Connect for low-latency TTS (ideal for LLM output)
conn, err := client.WebSocketTTS().Connect(ctx, voiceID, &elevenlabs.WebSocketTTSOptions{
    ModelID:                  "eleven_turbo_v2_5",
    OutputFormat:             "pcm_16000",
    OptimizeStreamingLatency: 3,
})
defer conn.Close()

// Stream text as it arrives (e.g., from LLM)
for text := range llmOutputStream {
    conn.SendText(text)
}
conn.Flush()

// Receive audio chunks
for audio := range conn.Audio() {
    // Play or save audio chunks
}
```

### WebSocket STT (Real-Time Transcription)

```go
// Connect for live transcription
conn, err := client.WebSocketSTT().Connect(ctx, &elevenlabs.WebSocketSTTOptions{
    SampleRate:     16000,
    EnablePartials: true,
})
defer conn.Close()

// Send audio chunks
go func() {
    for audioChunk := range microphoneInput {
        conn.SendAudio(audioChunk)
    }
    conn.EndStream()
}()

// Receive transcripts
for transcript := range conn.Transcripts() {
    if transcript.IsFinal {
        fmt.Println("Final:", transcript.Text)
    } else {
        fmt.Println("Partial:", transcript.Text)
    }
}
```

### Twilio Integration (Phone Calls)

```go
// Register incoming Twilio call with an ElevenLabs agent
resp, err := client.Twilio().RegisterCall(ctx, &elevenlabs.TwilioRegisterCallRequest{
    AgentID: "your-agent-id",
})
// Return resp.TwiML to Twilio webhook

// Make outbound call
call, err := client.Twilio().OutboundCall(ctx, &elevenlabs.TwilioOutboundCallRequest{
    AgentID:            "your-agent-id",
    AgentPhoneNumberID: "phone-number-id",
    ToNumber:           "+1234567890",
})

// List phone numbers
numbers, err := client.PhoneNumbers().List(ctx)
```

## Examples

See the [`examples/`](https://github.com/plexusone/elevenlabs-go/tree/main/examples) directory for runnable examples:

| Example | Description |
|---------|-------------|
| `basic/` | Common SDK operations |
| `ax-error-handling/` | AX error codes for machine-readable error handling |
| `websocket-tts/` | Real-time TTS streaming for LLM integration |
| `websocket-stt/` | Live transcription with partial results |
| `speech-to-speech/` | Voice conversion |
| `twilio/` | Phone call integration with Twilio |
| `ttsscript/` | Multi-voice script authoring |
| `retryhttp/` | Retry-capable HTTP transport |

```bash
export ELEVENLABS_API_KEY="your-api-key"
go run examples/basic/main.go
```

## Command Line Interface

The `elevenlabs` CLI provides text-to-speech generation from the command line.

### Basic Usage

```bash
# Generate speech from a text file
elevenlabs tts -v <voice-id> speech.txt

# Use a preset (oratory, podcast, audiobook)
elevenlabs tts -v <voice-id> --preset oratory speech.txt

# High-quality PCM output
elevenlabs tts -v <voice-id> -f pcm_48000 -o output.wav speech.txt

# Estimate credits without calling API
elevenlabs tts -v <voice-id> --estimate speech.txt
```

### Configuration Files

Save and reuse TTS settings with YAML config files:

```bash
# Use config file
elevenlabs tts --config tts-config.yaml speech.txt

# Save current settings to config
elevenlabs tts -v <voice-id> --preset oratory --save-config my-config.yaml speech.txt
```

Example config file:

```yaml
voice_id: IT8nQhZJj9jzRwmC46Ko
model_id: eleven_v3
output_format: pcm_48000

voice_settings:
  stability: 0.4        # Lower = more expressive
  similarity_boost: 0.75
  style: 0.3            # Higher = more dramatic
  speed: 0.95           # Slightly slower for gravitas
```

### Presets

| Preset | Stability | Style | Speed | Format | Use Case |
|--------|-----------|-------|-------|--------|----------|
| `oratory` | 0.4 | 0.3 | 0.95 | pcm_48000 | Speeches, presentations |
| `podcast` | 0.5 | 0.0 | 1.0 | mp3_44100_128 | Conversational content |
| `audiobook` | 0.6 | 0.1 | 0.95 | pcm_48000 | Long-form narration |

### Input Format

Text files support ElevenLabs formatting:

```
[calm] <break time="1s"/>
There are moments in history when humanity TRANSFORMS.
<break time="0.5s"/>
[excited] This is AMAZING news!
```

- SSML `<break>` tags for pauses
- Emotion tags (`[calm]`, `[excited]`, `[firm]`) for v3 model
- CAPITALIZED words for emphasis

## Error Handling

### Basic Error Handling

```go
audio, err := client.TextToSpeech().Simple(ctx, voiceID, text)
if err != nil {
    if elevenlabs.IsRateLimitError(err) {
        log.Println("Rate limited, waiting...")
        time.Sleep(time.Minute)
    } else if elevenlabs.IsUnauthorizedError(err) {
        log.Fatal("Invalid API key")
    } else if elevenlabs.IsNotFoundError(err) {
        log.Fatal("Voice not found")
    } else {
        log.Fatalf("Error: %v", err)
    }
}
```

### AX Error Codes (Machine-Readable)

For AI agents and automated systems, use AX error codes for precise error handling:

```go
import "github.com/plexusone/elevenlabs-go/ax"

_, err := client.Voices().Get(ctx, voiceID)
if err != nil {
    // Extract AX error code
    if code, ok := elevenlabs.GetAXErrorCode(err); ok {
        switch code {
        case ax.ErrDocumentNotFound:
            // Handle not found - try alternative resource
        case ax.ErrNotLoggedIn, ax.ErrNeedsAuthorization:
            // Handle auth - re-authenticate
        case ax.ErrInvalidUID:
            // Handle validation - fix input
        }

        // Get error metadata
        if info := ax.GetErrorInfo(code); info != nil {
            log.Printf("Category: %s, Retryable: %v", info.Category, info.Retryable)
        }
    }
}
```

## Environment Variables

- `ELEVENLABS_API_KEY`: Your ElevenLabs API key (used automatically if not provided via `WithAPIKey`)

## Documentation

- [API Reference](https://plexusone.github.io/elevenlabs-go/)
- [ElevenLabs API Docs](https://elevenlabs.io/docs)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License
