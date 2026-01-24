# Quick Start

This guide will have you generating audio in under 5 minutes.

## Create a Client

```go
import elevenlabs "github.com/agentplexus/go-elevenlabs"

// Uses ELEVENLABS_API_KEY environment variable
client, err := elevenlabs.NewClient()
if err != nil {
    log.Fatal(err)
}
```

## List Available Voices

```go
ctx := context.Background()

voices, err := client.Voices().List(ctx)
if err != nil {
    log.Fatal(err)
}

for _, v := range voices {
    fmt.Printf("%s: %s\n", v.VoiceID, v.Name)
}
```

## Generate Speech

### Simple Method

```go
// Generate speech with default settings
audio, err := client.TextToSpeech().Simple(ctx,
    "21m00Tcm4TlvDq8ikWAM",  // Voice ID (Rachel)
    "Hello, this is a test of the ElevenLabs API.")
if err != nil {
    log.Fatal(err)
}

// Save to file
f, _ := os.Create("output.mp3")
defer f.Close()
io.Copy(f, audio)
```

### With Options

```go
resp, err := client.TextToSpeech().Generate(ctx, &elevenlabs.TTSRequest{
    VoiceID: "21m00Tcm4TlvDq8ikWAM",
    Text:    "Hello with custom settings!",
    ModelID: "eleven_multilingual_v2",
    VoiceSettings: &elevenlabs.VoiceSettings{
        Stability:       0.5,
        SimilarityBoost: 0.75,
        Style:           0.2,
    },
    OutputFormat: "mp3_44100_128",
})
```

## Generate Sound Effects

```go
audio, err := client.SoundEffects().Simple(ctx, "thunder and rain storm")
if err != nil {
    log.Fatal(err)
}

f, _ := os.Create("thunder.mp3")
defer f.Close()
io.Copy(f, audio)
```

## Check Your Usage

```go
sub, err := client.User().GetSubscription(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Characters used: %d / %d\n",
    sub.CharacterCount, sub.CharacterLimit)
fmt.Printf("Remaining: %d\n", sub.CharactersRemaining())
```

## Complete Example

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

    // Generate speech using first voice
    if len(voices) > 0 {
        audio, err := client.TextToSpeech().Simple(ctx,
            voices[0].VoiceID,
            "Hello from go-elevenlabs!")
        if err != nil {
            log.Fatal(err)
        }

        f, _ := os.Create("hello.mp3")
        defer f.Close()
        n, _ := io.Copy(f, audio)
        fmt.Printf("Saved %d bytes to hello.mp3\n", n)
    }
}
```

## Next Steps

- [Configuration](configuration.md) - Custom HTTP clients, timeouts
- [Text-to-Speech](../services/text-to-speech.md) - Full TTS documentation
- [Voices](../services/voices.md) - Voice management
