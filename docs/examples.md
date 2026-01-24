# Examples

The `go-elevenlabs` SDK includes working examples in the [`examples/`](https://github.com/agentplexus/go-elevenlabs/tree/main/examples) directory.

## Running Examples

All examples require an ElevenLabs API key:

```bash
export ELEVENLABS_API_KEY="your-api-key"
cd examples/<example-name>
go run main.go
```

---

## Runnable Examples

### Basic Usage

**Location:** [`examples/basic/`](https://github.com/agentplexus/go-elevenlabs/tree/main/examples/basic)

Demonstrates common SDK operations: listing voices, models, checking subscription, generating speech, and working with projects.

```bash
go run examples/basic/main.go
```

### WebSocket TTS (Real-Time)

**Location:** [`examples/websocket-tts/`](https://github.com/agentplexus/go-elevenlabs/tree/main/examples/websocket-tts)

Real-time text-to-speech streaming via WebSocket. Ideal for LLM integration.

```bash
go run examples/websocket-tts/main.go
# Output: websocket_output.mp3
```

**Related docs:** [WebSocket TTS Service](services/websocket-tts.md)

### WebSocket STT (Real-Time)

**Location:** [`examples/websocket-stt/`](https://github.com/agentplexus/go-elevenlabs/tree/main/examples/websocket-stt)

Real-time speech-to-text transcription with partial results and word timing.

```bash
go run examples/websocket-stt/main.go <audio-file.wav>
```

**Related docs:** [WebSocket STT Service](services/websocket-stt.md)

### Speech-to-Speech

**Location:** [`examples/speech-to-speech/`](https://github.com/agentplexus/go-elevenlabs/tree/main/examples/speech-to-speech)

Voice conversion - transform audio from one voice to another.

```bash
go run examples/speech-to-speech/main.go input.mp3 output.mp3
```

**Related docs:** [Speech-to-Speech Service](services/speech-to-speech.md)

### Twilio Integration

**Location:** [`examples/twilio/`](https://github.com/agentplexus/go-elevenlabs/tree/main/examples/twilio)

Phone call integration for voice agent applications.

```bash
export ELEVENLABS_AGENT_ID="your-agent-id"
go run examples/twilio/main.go
# Server starts on :8080
```

**Related docs:** [Twilio Integration](services/twilio.md)

### TTS Script

**Location:** [`examples/ttsscript/`](https://github.com/agentplexus/go-elevenlabs/tree/main/examples/ttsscript)

Multi-voice, multi-chapter audio content with SSML-like markup.

```bash
go run examples/ttsscript/main.go
```

**Related docs:** [TTS Script Guide](guides/ttsscript.md)

### Retry HTTP Transport

**Location:** [`examples/retryhttp/`](https://github.com/agentplexus/go-elevenlabs/tree/main/examples/retryhttp)

Retry-capable HTTP transport for resilient API calls.

```bash
go run examples/retryhttp/main.go
```

**Related docs:** [Retry HTTP Transport](utilities/retryhttp.md)

---

## Code Snippets

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "io"
    "log"
    "os"

    elevenlabs "github.com/agentplexus/go-elevenlabs"
)

func main() {
    client, err := elevenlabs.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // List voices
    voices, _ := client.Voices().List(ctx)
    fmt.Printf("Found %d voices\n", len(voices))

    // Generate speech
    if len(voices) > 0 {
        audio, _ := client.TextToSpeech().Simple(ctx,
            voices[0].VoiceID,
            "Hello from go-elevenlabs!")

        f, _ := os.Create("hello.mp3")
        defer f.Close()
        io.Copy(f, audio)
    }
}
```

### Text-to-Speech with Options

```go
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
if err != nil {
    log.Fatal(err)
}

f, _ := os.Create("custom.mp3")
defer f.Close()
io.Copy(f, resp.Audio)
```

### WebSocket TTS (LLM Integration)

```go
// Connect with low-latency settings
conn, err := client.WebSocketTTS().Connect(ctx, voiceID, &elevenlabs.WebSocketTTSOptions{
    ModelID:                  "eleven_turbo_v2_5",
    OutputFormat:             "pcm_16000",
    OptimizeStreamingLatency: 3,
})
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Stream text from LLM
for token := range llmStream {
    conn.SendText(token)
}
conn.Flush()

// Receive audio chunks
for audio := range conn.Audio() {
    player.Write(audio)
}
```

### WebSocket STT (Live Transcription)

```go
conn, err := client.WebSocketSTT().Connect(ctx, &elevenlabs.WebSocketSTTOptions{
    ModelID:              "scribe_v1",
    SampleRate:           16000,
    EnablePartials:       true,
    EnableWordTimestamps: true,
})
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Stream audio
go func() {
    for audioChunk := range microphoneStream {
        conn.SendAudio(audioChunk)
    }
    conn.EndStream()
}()

// Receive transcripts
for transcript := range conn.Transcripts() {
    if transcript.IsFinal {
        fmt.Println(transcript.Text)
    }
}
```

### Sound Effects

```go
// Simple sound effect
thunder, _ := client.SoundEffects().Simple(ctx, "thunder and rain storm")

// With options
sfx, _ := client.SoundEffects().Generate(ctx, &elevenlabs.SoundEffectRequest{
    Text:            "spaceship engine humming",
    DurationSeconds: 10,
    PromptInfluence: 0.5,
    Loop:            true,
})

// Looping background
ambience, _ := client.SoundEffects().GenerateLoop(ctx,
    "peaceful forest with birds", 30)
```

### Pronunciation Dictionary

```go
// From a map (simplest)
dict, _ := client.Pronunciation().CreateFromMap(ctx, "Tech Terms", map[string]string{
    "API":     "A P I",
    "kubectl": "kube control",
    "nginx":   "engine X",
})

// From JSON file
dict, _ := client.Pronunciation().CreateFromJSON(ctx, "Terms", "terms.json")

// With full options
rules := elevenlabs.PronunciationRules{
    {Grapheme: "API", Alias: "A P I"},
    {Grapheme: "nginx", Phoneme: "ˈɛndʒɪnˈɛks"},
}

dict, _ := client.Pronunciation().Create(ctx, &elevenlabs.CreatePronunciationDictionaryRequest{
    Name:        "Custom Terms",
    Description: "Technical vocabulary",
    Rules:       rules,
    Language:    "en-US",
})
```

### Projects (Long-form Content)

```go
// Create project
project, _ := client.Projects().Create(ctx, &elevenlabs.CreateProjectRequest{
    Name:                    "My Audiobook",
    DefaultModelID:          "eleven_multilingual_v2",
    DefaultParagraphVoiceID: "21m00Tcm4TlvDq8ikWAM",
    DefaultTitleVoiceID:     "21m00Tcm4TlvDq8ikWAM",
})

// List chapters
chapters, _ := client.Projects().ListChapters(ctx, project.ProjectID)

// Convert project to audio
client.Projects().Convert(ctx, project.ProjectID)

// Download completed audio
snapshots, _ := client.Projects().ListSnapshots(ctx, project.ProjectID)
if len(snapshots) > 0 {
    reader, _ := client.Projects().DownloadSnapshotArchive(ctx,
        project.ProjectID, snapshots[0].ProjectSnapshotID)

    f, _ := os.Create("audiobook.zip")
    io.Copy(f, reader)
    f.Close()
}
```

### Error Handling

```go
audio, err := client.TextToSpeech().Simple(ctx, voiceID, text)
if err != nil {
    if elevenlabs.IsRateLimitError(err) {
        log.Println("Rate limited, waiting...")
        time.Sleep(time.Minute)
        // Retry...
    } else if elevenlabs.IsUnauthorizedError(err) {
        log.Fatal("Invalid API key")
    } else if elevenlabs.IsNotFoundError(err) {
        log.Fatal("Voice not found")
    } else {
        log.Fatalf("Error: %v", err)
    }
}
```

---

## Use Case Examples

### Voice Agents (Twilio + WebSocket)

Build phone-based voice agents:

1. Use **Twilio example** to handle incoming/outgoing calls
2. Use **WebSocket STT** to transcribe caller speech
3. Use **WebSocket TTS** to generate agent responses

### LLM Integration

Stream LLM responses to audio:

```go
conn, _ := client.WebSocketTTS().Connect(ctx, voiceID, nil)
for token := range llmStream {
    conn.SendText(token)
}
conn.Flush()
```

### Audio Content Creation

Create audiobooks, courses, or podcasts using TTS Script format for multi-voice, multi-chapter content.

### Voice Conversion Pipeline

1. Record original audio
2. Use Speech-to-Speech to convert voice
3. Optionally clean up with Audio Isolation

---

## Logging Pattern

All real-time examples use context-based structured logging with `slog` and `slogutil` from `github.com/grokify/mogo`. This pattern provides:

- **Request-scoped logging** - Each HTTP request or operation gets its own logger via context
- **Silent by default** - Uses `slogutil.Null()` fallback for quiet operation
- **Testable** - Easy to inject mock loggers for testing
- **No global state** - Loggers flow through context, not package variables

### Pattern

```go
import (
    "context"
    "log/slog"
    "github.com/grokify/mogo/log/slogutil"
)

func main() {
    // Attach logger to context
    ctx := slogutil.ContextWithLogger(context.Background(), slog.Default())

    // Pass context to functions
    doWork(ctx)
}

// Helper functions retrieve logger from context
func logInfo(ctx context.Context, msg string, args ...any) {
    slogutil.LoggerFromContext(ctx, slogutil.Null()).Info(msg, args...)
}

func logError(ctx context.Context, msg string, err error, args ...any) {
    logger := slogutil.LoggerFromContext(ctx, slogutil.Null())
    if err != nil {
        args = append([]any{"error", err}, args...)
    }
    logger.Error(msg, args...)
}
```

### HTTP Middleware

For HTTP handlers, use middleware to attach loggers to request context:

```go
func withLogger(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := slogutil.ContextWithLogger(r.Context(), slog.Default())
        next(w, r.WithContext(ctx))
    }
}

// Usage
http.HandleFunc("/api/endpoint", withLogger(handleEndpoint))
```
