# STT Provider

The STT provider implements `stt.Provider` and `stt.StreamingProvider` interfaces using the ElevenLabs Speech-to-Text API.

## Features

- Synchronous transcription (audio bytes, file, or URL)
- Streaming transcription (real-time via WebSocket)
- Speaker diarization
- Word-level timestamps
- Language detection

## Installation

```go
import elevenstt "github.com/agentplexus/go-elevenlabs/omnivoice/stt"
```

## Creating a Provider

```go
// Using environment variable (ELEVENLABS_API_KEY)
provider, err := elevenstt.New()

// With explicit API key
provider, err := elevenstt.New(
    elevenstt.WithAPIKey("your-api-key"),
)

// With existing ElevenLabs client
client, _ := elevenlabs.NewClient()
provider := elevenstt.NewWithClient(client)
```

## Synchronous Transcription

### From Audio Bytes

```go
audioData, _ := os.ReadFile("audio.mp3")

result, err := provider.Transcribe(ctx, audioData, stt.TranscriptionConfig{
    Language: "en",
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Text)
```

### From File Path

```go
result, err := provider.TranscribeFile(ctx, "/path/to/audio.mp3", stt.TranscriptionConfig{
    Language: "en",
})
```

### From URL

```go
result, err := provider.TranscribeURL(ctx, "https://example.com/audio.mp3", stt.TranscriptionConfig{
    Language: "en",
})
```

## Configuration Options

```go
config := stt.TranscriptionConfig{
    // Language code (e.g., "en", "es", "fr")
    Language: "en",

    // Model ID (optional, uses default if empty)
    Model: "scribe_v1",

    // Enable speaker diarization
    EnableSpeakerDiarization: true,
    MaxSpeakers:              2,

    // Enable word-level timestamps
    EnableWordTimestamps: true,
}
```

## Streaming Transcription

Get real-time transcription results as audio is being processed:

```go
writer, events, err := provider.TranscribeStream(ctx, stt.TranscriptionConfig{
    Language:             "en",
    EnableWordTimestamps: true,
})
if err != nil {
    log.Fatal(err)
}

// Send audio in a goroutine
go func() {
    defer writer.Close()

    // Stream audio chunks (e.g., from microphone)
    for chunk := range audioSource {
        if _, err := writer.Write(chunk); err != nil {
            log.Printf("Write error: %v", err)
            return
        }
    }
}()

// Process transcription events
for event := range events {
    switch event.Type {
    case stt.EventTranscript:
        if event.IsFinal {
            fmt.Printf("[Final] %s\n", event.Transcript)
        } else {
            fmt.Printf("[Partial] %s\n", event.Transcript)
        }
    case stt.EventError:
        log.Printf("Error: %v", event.Error)
    }
}
```

## Working with Results

### Basic Transcription

```go
result, _ := provider.Transcribe(ctx, audioData, config)

fmt.Printf("Text: %s\n", result.Text)
fmt.Printf("Language: %s\n", result.Language)
```

### Word-Level Timestamps

```go
config := stt.TranscriptionConfig{
    EnableWordTimestamps: true,
}

result, _ := provider.Transcribe(ctx, audioData, config)

for _, segment := range result.Segments {
    for _, word := range segment.Words {
        fmt.Printf("%s [%.2fs - %.2fs] (%.2f confidence)\n",
            word.Text,
            word.StartTime.Seconds(),
            word.EndTime.Seconds(),
            word.Confidence,
        )
    }
}
```

### Speaker Diarization

```go
config := stt.TranscriptionConfig{
    EnableSpeakerDiarization: true,
    MaxSpeakers:              3,
}

result, _ := provider.Transcribe(ctx, audioData, config)

for _, segment := range result.Segments {
    fmt.Printf("[Speaker %s] %s\n", segment.Speaker, segment.Text)
}
```

## Audio Encoding

The provider supports these audio encodings for streaming:

| OmniVoice Encoding | ElevenLabs Encoding | Description |
|--------------------|---------------------|-------------|
| `pcm`, `pcm_s16le` | `pcm_s16le` | 16-bit signed little-endian PCM |
| `mulaw`, `pcm_mulaw` | `pcm_mulaw` | μ-law encoded audio |

Configure encoding in the transcription config:

```go
config := stt.TranscriptionConfig{
    Encoding:   "pcm_s16le",
    SampleRate: 16000,
}
```

## Accessing the Underlying Client

For ElevenLabs-specific features:

```go
client := provider.Client()

// Use any ElevenLabs SDK method directly
resp, _ := client.SpeechToText().TranscribeWithTags(ctx, req)
```

## Error Handling

```go
result, err := provider.Transcribe(ctx, audioData, config)
if err != nil {
    if elevenlabs.IsForbiddenError(err) {
        log.Fatal("Invalid API key or insufficient permissions")
    }
    log.Fatalf("Transcription failed: %v", err)
}
```
