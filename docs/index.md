# go-elevenlabs

A Go SDK for the [ElevenLabs](https://elevenlabs.io/) AI Audio API.

## Features

- **Text-to-Speech** - Generate natural-sounding speech from text
- **Speech-to-Text** - Transcribe audio with speaker diarization
- **Speech-to-Speech** - Voice conversion to transform any voice
- **Voice Selection** - Access pre-made and cloned voices
- **Sound Effects** - Generate sound effects from text descriptions
- **Music Composition** - Generate music from text prompts
- **Projects (Studio)** - Create long-form content with chapters
- **Pronunciation Dictionaries** - Ensure correct pronunciation of technical terms
- **Dubbing** - Translate audio/video to other languages

### Real-Time Services

- **WebSocket TTS** - Low-latency streaming text-to-speech for LLM integration
- **WebSocket STT** - Real-time speech-to-text with partial results
- **Twilio Integration** - Phone call integration for voice agents
- **Phone Numbers** - Manage phone numbers for conversational AI

## Installation

```bash
go get github.com/agentplexus/go-elevenlabs
```

## Quick Example

```go
package main

import (
    "context"
    "io"
    "os"

    elevenlabs "github.com/agentplexus/go-elevenlabs"
)

func main() {
    // Create client (uses ELEVENLABS_API_KEY env var)
    client, _ := elevenlabs.NewClient()
    ctx := context.Background()

    // Generate speech
    audio, _ := client.TextToSpeech().Simple(ctx,
        "21m00Tcm4TlvDq8ikWAM",  // Voice ID
        "Hello, welcome to ElevenLabs!")

    // Save to file
    f, _ := os.Create("output.mp3")
    defer f.Close()
    io.Copy(f, audio)
}
```

## Use Cases

This SDK is particularly well-suited for:

- **Voice Agents** - Build conversational AI agents with real-time TTS/STT
- **Phone Integration** - Create voice bots with Twilio phone calls
- **Online Courses** - Generate professional narration for Udemy, LMS platforms
- **Audiobooks** - Create chapter-organized audio content
- **Podcasts** - Produce consistent, high-quality audio
- **Video Production** - Add voiceovers and sound effects
- **LLM Applications** - Stream text from LLMs directly to speech

## Documentation

- [Getting Started](getting-started/installation.md) - Installation and setup
- [Services](services/text-to-speech.md) - API service documentation
- [Guides](guides/lms-courses.md) - Use case guides
- [Examples](examples.md) - Code examples

## License

MIT License - see [LICENSE](https://github.com/agentplexus/go-elevenlabs/blob/main/LICENSE) for details.
