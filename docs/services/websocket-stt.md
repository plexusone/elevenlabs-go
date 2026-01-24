# WebSocket STT

Real-time speech-to-text streaming via WebSocket for live transcription.

## Overview

The WebSocket STT service enables real-time audio transcription with:

- **Partial Results**: Get interim transcripts for responsive UIs
- **Word Timing**: Word-level timestamps and confidence scores
- **Language Detection**: Automatic language identification
- **Low Latency**: Stream audio as it's captured

## Basic Usage

```go
// Connect to WebSocket STT
conn, err := client.WebSocketSTT().Connect(ctx, nil)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Send audio chunks
go func() {
    for {
        audioChunk := captureAudio() // Your audio capture function
        if err := conn.SendAudio(audioChunk); err != nil {
            return
        }
    }
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

## With Options

```go
opts := &elevenlabs.WebSocketSTTOptions{
    // Scribe model for transcription
    ModelID: "scribe_v1",

    // Audio settings
    SampleRate: 16000,
    Encoding:   "pcm_s16le",

    // Enable partial/interim results
    EnablePartials: true,

    // Get word-level timing
    EnableWordTimestamps: true,

    // Specify language (or leave empty for auto-detect)
    LanguageCode: "en",
}

conn, err := client.WebSocketSTT().Connect(ctx, opts)
```

## Streaming from Microphone

```go
// Connect to STT
conn, err := client.WebSocketSTT().Connect(ctx, &elevenlabs.WebSocketSTTOptions{
    SampleRate:           16000,
    Encoding:             "pcm_s16le",
    EnablePartials:       true,
    EnableWordTimestamps: true,
})
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Stream microphone audio
go func() {
    for {
        chunk, err := microphone.Read()
        if err != nil {
            conn.EndStream()
            return
        }
        conn.SendAudio(chunk)
    }
}()

// Display transcripts
for transcript := range conn.Transcripts() {
    if transcript.IsFinal {
        fmt.Printf("\n[FINAL] %s\n", transcript.Text)
    } else {
        fmt.Printf("\r[...] %s", transcript.Text)
    }
}
```

## Using StreamAudio Helper

```go
// Create audio input channel
audioStream := make(chan []byte)

// Start streaming (handles EndStream automatically)
transcriptOut, errOut := conn.StreamAudio(ctx, audioStream)

// Send audio chunks
go func() {
    defer close(audioStream)
    for {
        chunk := captureAudio()
        audioStream <- chunk
    }
}()

// Receive transcripts
for transcript := range transcriptOut {
    fmt.Println(transcript.Text)
}

// Check for errors
if err := <-errOut; err != nil {
    log.Printf("streaming error: %v", err)
}
```

## Word-Level Timing

```go
for transcript := range conn.Transcripts() {
    if transcript.IsFinal && len(transcript.Words) > 0 {
        fmt.Printf("Transcript: %s\n", transcript.Text)
        fmt.Printf("Language: %s\n", transcript.LanguageCode)
        fmt.Printf("Confidence: %.2f\n", transcript.Confidence)

        for _, word := range transcript.Words {
            fmt.Printf("  '%s': %.3fs - %.3fs (conf: %.2f)\n",
                word.Word,
                word.Start,
                word.End,
                word.Confidence)
        }
    }
}
```

## Error Handling

```go
// Monitor errors
go func() {
    for err := range conn.Errors() {
        log.Printf("WebSocket STT error: %v", err)
    }
}()
```

## Options Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `ModelID` | string | `scribe_v1` | Transcription model |
| `SampleRate` | int | 16000 | Audio sample rate in Hz |
| `Encoding` | string | `pcm_s16le` | Audio encoding format |
| `LanguageCode` | string | "" | Expected language (auto-detect if empty) |
| `EnablePartials` | bool | true | Enable interim results |
| `EnableWordTimestamps` | bool | true | Include word timing |
| `MaxAlternatives` | int | 0 | Number of alternative transcripts |

## Transcript Fields

| Field | Type | Description |
|-------|------|-------------|
| `Text` | string | Transcribed text |
| `IsFinal` | bool | True if final result |
| `Confidence` | float64 | Overall confidence (0-1) |
| `Words` | []STTWord | Word-level details |
| `LanguageCode` | string | Detected language |
| `StartTime` | float64 | Start time in seconds |
| `EndTime` | float64 | End time in seconds |

## Audio Formats

Supported audio encodings:

- `pcm_s16le` - 16-bit signed little-endian PCM (recommended)
- `pcm_mulaw` - 8-bit mu-law PCM (telephony)

Common sample rates:

- 8000 Hz - Telephony
- 16000 Hz - Voice (recommended)
- 22050 Hz - High quality voice
- 44100 Hz - CD quality
