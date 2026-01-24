# WebSocket TTS

Real-time text-to-speech streaming via WebSocket for low-latency voice synthesis.

## Overview

The WebSocket TTS service enables streaming text to speech in real-time, making it ideal for:

- **LLM Integration**: Stream text from language models as it's generated
- **Interactive Applications**: Voice assistants, chatbots, real-time narration
- **Low Latency**: Get audio output before the full text is available

## Basic Usage

```go
// Connect to WebSocket TTS
conn, err := client.WebSocketTTS().Connect(ctx, voiceID, nil)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Send text
conn.SendText("Hello, ")
conn.SendText("this is streaming ")
conn.SendText("text to speech!")

// Flush to finalize
conn.Flush()

// Receive audio chunks
for audio := range conn.Audio() {
    // Play or save audio chunks
    player.Write(audio)
}
```

## With Options

```go
opts := &elevenlabs.WebSocketTTSOptions{
    // Use turbo model for lowest latency
    ModelID: "eleven_turbo_v2_5",

    // PCM format for real-time playback
    OutputFormat: "pcm_16000",

    // Latency optimization (0-4, higher = faster but lower quality)
    OptimizeStreamingLatency: 3,

    // Enable SSML parsing
    EnableSSMLParsing: true,

    // Voice settings
    VoiceSettings: &elevenlabs.VoiceSettings{
        Stability:       0.5,
        SimilarityBoost: 0.75,
    },
}

conn, err := client.WebSocketTTS().Connect(ctx, voiceID, opts)
```

## Streaming from LLM

```go
// Connect to TTS
conn, err := client.WebSocketTTS().Connect(ctx, voiceID, &elevenlabs.WebSocketTTSOptions{
    ModelID:                  "eleven_turbo_v2_5",
    OutputFormat:             "pcm_16000",
    OptimizeStreamingLatency: 3,
})
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Stream LLM output to TTS
go func() {
    for chunk := range llmOutputStream {
        if err := conn.SendText(chunk); err != nil {
            log.Printf("send error: %v", err)
            return
        }
    }
    conn.Flush()
}()

// Play audio as it arrives
for audio := range conn.Audio() {
    audioPlayer.Write(audio)
}
```

## Using StreamText Helper

```go
// Create a channel of text chunks
textStream := make(chan string)

// Start streaming (this handles flushing automatically)
audioOut, errOut := conn.StreamText(ctx, textStream)

// Send text chunks
go func() {
    defer close(textStream)
    textStream <- "Hello, "
    textStream <- "world!"
}()

// Receive audio
for audio := range audioOut {
    // Process audio
}

// Check for errors
if err := <-errOut; err != nil {
    log.Printf("streaming error: %v", err)
}
```

## Word Alignments

```go
// Receive word-level timing
go func() {
    for align := range conn.Alignments() {
        for i, char := range align.Characters {
            fmt.Printf("%s: %.3fs - %.3fs\n",
                char,
                align.CharacterStart[i],
                align.CharacterEnd[i])
        }
    }
}()
```

## Error Handling

```go
// Monitor errors
go func() {
    for err := range conn.Errors() {
        log.Printf("WebSocket error: %v", err)
    }
}()
```

## Stream Completion Behavior

ElevenLabs WebSocket TTS does **not** send an explicit "end of stream" signal. After calling `Flush()`, the server generates any remaining audio and then waits for more input. If no input arrives within the inactivity timeout (default 20 seconds), the server sends an `input_timeout_exceeded` error and closes the connection.

This behavior has implications for detecting when audio generation is complete:

### Default Behavior

With the default 20-second timeout, your application will wait up to 20 seconds after the last audio chunk before the connection closes:

```go
conn.Flush()

// This loop will block for up to 20 seconds after last audio
for audio := range conn.Audio() {
    player.Write(audio)
}
```

### Faster Completion Detection

For applications that need faster stream completion, set a shorter `InactivityTimeout` and treat the timeout as successful completion:

```go
opts := &elevenlabs.WebSocketTTSOptions{
    ModelID:           "eleven_turbo_v2_5",
    OutputFormat:      "pcm_16000",
    InactivityTimeout: 5, // 5 seconds instead of 20
}

conn, err := client.WebSocketTTS().Connect(ctx, voiceID, opts)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Send text and flush
conn.SendText("Hello, world!")
conn.Flush()

// Use Done() channel to detect completion
var receivedAudio bool
for {
    select {
    case audio, ok := <-conn.Audio():
        if !ok {
            return // Channel closed
        }
        receivedAudio = true
        player.Write(audio)
    case <-conn.Done():
        // All audio received after flush
        return
    case err := <-conn.Errors():
        // Treat timeout as success if we received audio
        if receivedAudio && strings.Contains(err.Error(), "input_timeout_exceeded") {
            return // Stream completed successfully
        }
        log.Printf("error: %v", err)
        return
    }
}
```

### OmniVoice Provider

The OmniVoice TTS provider handles this automatically by setting a 5-second inactivity timeout and treating the timeout as successful completion when audio was received after flush.

## Options Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `ModelID` | string | `eleven_turbo_v2_5` | TTS model to use |
| `OutputFormat` | string | `pcm_16000` | Audio format |
| `VoiceSettings` | *VoiceSettings | nil | Voice parameters |
| `OptimizeStreamingLatency` | int | 3 | Latency vs quality (0-4) |
| `EnableSSMLParsing` | bool | false | Parse SSML in text |
| `LanguageCode` | string | "" | ISO language code |
| `ChunkLengthSchedule` | []int | nil | Custom chunking |
| `InactivityTimeout` | int | 20 | Timeout in seconds |

## Output Formats

For real-time playback, PCM formats are recommended:

- `pcm_16000` - 16kHz PCM (lowest latency)
- `pcm_22050` - 22.05kHz PCM
- `pcm_24000` - 24kHz PCM
- `pcm_44100` - 44.1kHz PCM (highest quality)

MP3 formats are also available but add encoding latency:

- `mp3_44100_64` - 64kbps MP3
- `mp3_44100_128` - 128kbps MP3
- `mp3_44100_192` - 192kbps MP3
