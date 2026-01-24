# Text-to-Speech

The Text-to-Speech service converts text into natural-sounding speech.

## Basic Usage

### Simple Generation

```go
audio, err := client.TextToSpeech().Simple(ctx, voiceID, "Your text here")
if err != nil {
    log.Fatal(err)
}

// audio is an io.Reader - save or stream it
f, _ := os.Create("output.mp3")
io.Copy(f, audio)
```

### Full Control

```go
resp, err := client.TextToSpeech().Generate(ctx, &elevenlabs.TTSRequest{
    VoiceID: "21m00Tcm4TlvDq8ikWAM",
    Text:    "Hello world!",
    ModelID: "eleven_multilingual_v2",
    VoiceSettings: &elevenlabs.VoiceSettings{
        Stability:       0.5,
        SimilarityBoost: 0.75,
        Style:           0.0,
        SpeakerBoost:    true,
    },
    OutputFormat: "mp3_44100_128",
})
```

## Voice Settings

| Setting | Range | Description |
|---------|-------|-------------|
| `Stability` | 0.0-1.0 | Higher = more consistent, lower = more expressive |
| `SimilarityBoost` | 0.0-1.0 | Higher = closer to original voice |
| `Style` | 0.0-1.0 | Style exaggeration (use sparingly) |
| `SpeakerBoost` | bool | Enhance speaker similarity |
| `Speed` | 0.7-1.2 | Speech speed multiplier |

### Default Settings

```go
settings := elevenlabs.DefaultVoiceSettings()
// Stability: 0.5, SimilarityBoost: 0.75, Style: 0, SpeakerBoost: true
```

## Output Formats

| Format | Description |
|--------|-------------|
| `mp3_44100_128` | MP3, 44.1kHz, 128kbps (default) |
| `mp3_44100_192` | MP3, 44.1kHz, 192kbps |
| `pcm_16000` | PCM, 16kHz |
| `pcm_22050` | PCM, 22.05kHz |
| `pcm_24000` | PCM, 24kHz |
| `pcm_44100` | PCM, 44.1kHz |
| `ulaw_8000` | u-law, 8kHz |

## Models

| Model ID | Best For |
|----------|----------|
| `eleven_multilingual_v2` | Multiple languages, highest quality |
| `eleven_monolingual_v1` | English only, fast |
| `eleven_turbo_v2` | Low latency applications |
| `eleven_turbo_v2_5` | Lowest latency |

## Streaming

For real-time applications, use streaming:

```go
resp, err := client.TextToSpeech().GenerateStream(ctx, &elevenlabs.TTSRequest{
    VoiceID: voiceID,
    Text:    "Long text to stream...",
})
if err != nil {
    log.Fatal(err)
}

// Stream audio chunks as they arrive
for chunk := range resp.Chunks {
    // Process chunk
}
```

## Error Handling

```go
audio, err := client.TextToSpeech().Simple(ctx, voiceID, text)
if err != nil {
    if elevenlabs.IsRateLimitError(err) {
        // Wait and retry
        time.Sleep(time.Minute)
    } else if elevenlabs.IsUnauthorizedError(err) {
        // Check API key
    }
    return err
}
```

## Best Practices

1. **Reuse the client** - Create one client and reuse it
2. **Check character limits** - Monitor `CharactersRemaining()` before generation
3. **Use appropriate models** - `turbo` for speed, `multilingual_v2` for quality
4. **Batch when possible** - Combine short texts to reduce API calls
