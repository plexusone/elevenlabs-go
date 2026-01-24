# TTS Provider

The TTS provider implements `tts.Provider` and `tts.StreamingProvider` interfaces using the ElevenLabs Text-to-Speech API.

## Features

- Synchronous synthesis (returns complete audio)
- Streaming synthesis (real-time audio chunks via WebSocket)
- LLM output streaming (pipe reader directly to TTS)
- Voice listing and caching
- Output format mapping

## Installation

```go
import eleventts "github.com/agentplexus/go-elevenlabs/omnivoice/tts"
```

## Creating a Provider

```go
// Using environment variable (ELEVENLABS_API_KEY)
provider, err := eleventts.New()

// With explicit API key
provider, err := eleventts.New(
    eleventts.WithAPIKey("your-api-key"),
)

// With existing ElevenLabs client
client, _ := elevenlabs.NewClient()
provider := eleventts.NewWithClient(client)
```

## Synchronous Synthesis

Convert text to audio in a single request:

```go
result, err := provider.Synthesize(ctx, "Hello, world!", tts.SynthesisConfig{
    VoiceID:      "21m00Tcm4TlvDq8ikWAM", // Rachel
    Model:        "eleven_turbo_v2_5",
    OutputFormat: "mp3",
    SampleRate:   44100,
})
if err != nil {
    log.Fatal(err)
}

// result.Audio contains the complete audio data
os.WriteFile("output.mp3", result.Audio, 0644)
```

### Voice Settings

Fine-tune the voice output:

```go
result, err := provider.Synthesize(ctx, text, tts.SynthesisConfig{
    VoiceID:         "21m00Tcm4TlvDq8ikWAM",
    Stability:       0.5,  // 0-1, lower = more expressive
    SimilarityBoost: 0.75, // 0-1, higher = closer to original
    Speed:           1.0,  // 0.5-2.0, speech rate
})
```

## Streaming Synthesis

Get audio chunks in real-time via WebSocket:

```go
chunks, err := provider.SynthesizeStream(ctx, "Hello, world!", tts.SynthesisConfig{
    VoiceID: "21m00Tcm4TlvDq8ikWAM",
})
if err != nil {
    log.Fatal(err)
}

for chunk := range chunks {
    if chunk.Error != nil {
        log.Printf("Error: %v", chunk.Error)
        break
    }
    if chunk.IsFinal {
        break
    }
    // Process audio chunk
    playAudio(chunk.Audio)
}
```

## LLM Output Streaming

Stream text from an LLM directly to TTS:

```go
// Create a pipe for LLM output
pr, pw := io.Pipe()

// Start streaming LLM output to the pipe
go func() {
    defer pw.Close()
    for token := range llmTokens {
        pw.Write([]byte(token))
    }
}()

// Stream pipe content to TTS
chunks, err := provider.SynthesizeFromReader(ctx, pr, tts.SynthesisConfig{
    VoiceID: "21m00Tcm4TlvDq8ikWAM",
})
```

This enables ultra-low latency voice responses where TTS begins before the LLM finishes generating.

## Voice Management

### List Available Voices

```go
voices, err := provider.ListVoices(ctx)
if err != nil {
    log.Fatal(err)
}

for _, v := range voices {
    fmt.Printf("%s: %s (%s)\n", v.ID, v.Name, v.Gender)
}
```

Voices are cached after the first call. Clear the cache if needed:

```go
provider.ClearVoiceCache()
```

### Get Specific Voice

```go
voice, err := provider.GetVoice(ctx, "21m00Tcm4TlvDq8ikWAM")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Name: %s\n", voice.Name)
fmt.Printf("Gender: %s\n", voice.Gender)
fmt.Printf("Language: %s\n", voice.Language)
```

## Output Formats

The provider maps OmniVoice format names to ElevenLabs formats:

| OmniVoice Format | ElevenLabs Format | Notes |
|------------------|-------------------|-------|
| `mp3` | `mp3_44100_128` | Default, 128kbps |
| `pcm` | `pcm_16000` | Raw PCM, 16kHz |
| `wav` | `pcm_44100` | Raw PCM (add WAV header manually) |
| `opus` | `mp3_44100_128` | Fallback (not supported) |

You can also use ElevenLabs format strings directly:

```go
config := tts.SynthesisConfig{
    OutputFormat: "mp3_22050_32", // Use ElevenLabs format directly
}
```

## Accessing the Underlying Client

For ElevenLabs-specific features not exposed by OmniVoice:

```go
client := provider.Client()

// Use any ElevenLabs SDK method
voices, _ := client.Voices().List(ctx)
```

## Error Handling

```go
result, err := provider.Synthesize(ctx, text, config)
if err != nil {
    // Check for specific error types
    if elevenlabs.IsForbiddenError(err) {
        log.Fatal("Invalid API key or insufficient permissions")
    }
    log.Fatalf("TTS failed: %v", err)
}
```
