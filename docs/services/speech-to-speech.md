# Speech-to-Speech

Voice conversion service that transforms speech from one voice to another while preserving the content.

## Overview

The Speech-to-Speech service enables:

- **Voice Conversion**: Transform any voice to a target voice
- **Content Preservation**: Keep the original speech content
- **Background Noise Removal**: Clean up source audio
- **Streaming**: Real-time voice conversion

## Basic Usage

```go
// Open source audio file
f, err := os.Open("source_audio.mp3")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

// Convert to target voice
resp, err := client.SpeechToSpeech().Convert(ctx, &elevenlabs.SpeechToSpeechRequest{
    VoiceID: "target-voice-id",
    Audio:   f,
})
if err != nil {
    log.Fatal(err)
}

// Save converted audio
out, _ := os.Create("converted.mp3")
defer out.Close()
io.Copy(out, resp.Audio)
```

## Simple Conversion

```go
// One-line conversion
f, _ := os.Open("input.mp3")
audio, err := client.SpeechToSpeech().Simple(ctx, targetVoiceID, f)
if err != nil {
    log.Fatal(err)
}

// Save output
out, _ := os.Create("output.mp3")
io.Copy(out, audio)
```

## With Full Options

```go
sourceFile, _ := os.Open("speaker_a.mp3")

resp, err := client.SpeechToSpeech().Convert(ctx, &elevenlabs.SpeechToSpeechRequest{
    // Target voice
    VoiceID: "21m00Tcm4TlvDq8ikWAM",

    // Source audio
    Audio:         sourceFile,
    AudioFilename: "speaker_a.mp3", // Helps with format detection

    // Model selection
    ModelID: "eleven_english_sts_v2",

    // Voice settings
    VoiceSettings: &elevenlabs.VoiceSettings{
        Stability:       0.5,
        SimilarityBoost: 0.8,
        Style:           0.0,
        UseSpeakerBoost: true,
    },

    // Output format
    OutputFormat: "mp3_44100_128",

    // Remove background noise from source
    RemoveBackgroundNoise: true,
})
```

## Streaming Conversion

For real-time voice conversion:

```go
resp, err := client.SpeechToSpeech().ConvertStream(ctx, &elevenlabs.SpeechToSpeechRequest{
    VoiceID: targetVoiceID,
    Audio:   sourceAudio,
    OutputFormat: "pcm_22050",
})
if err != nil {
    log.Fatal(err)
}

// Stream to audio player
player := audio.NewPlayer(22050)
io.Copy(player, resp.Audio)
```

## With Seed Audio

Use seed audio for more consistent conversions:

```go
sourceFile, _ := os.Open("input.mp3")
seedFile, _ := os.Open("seed_sample.mp3")

resp, err := client.SpeechToSpeech().Convert(ctx, &elevenlabs.SpeechToSpeechRequest{
    VoiceID: targetVoiceID,
    Audio:   sourceFile,

    // Seed audio influences the conversion style
    SeedAudio:         seedFile,
    SeedAudioFilename: "seed_sample.mp3",
})
```

## Request Options

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `VoiceID` | string | Yes | Target voice ID |
| `Audio` | io.Reader | Yes | Source audio data |
| `AudioFilename` | string | No | Source filename hint |
| `ModelID` | string | No | Model (default: `eleven_english_sts_v2`) |
| `VoiceSettings` | *VoiceSettings | No | Voice parameters |
| `OutputFormat` | string | No | Output audio format |
| `RemoveBackgroundNoise` | bool | No | Clean source audio |
| `SeedAudio` | io.Reader | No | Reference audio for style |
| `SeedAudioFilename` | string | No | Seed filename hint |

## Output Formats

Available output formats:

**MP3 Formats:**
- `mp3_44100_64` - 64kbps MP3
- `mp3_44100_96` - 96kbps MP3
- `mp3_44100_128` - 128kbps MP3 (recommended)
- `mp3_44100_192` - 192kbps MP3

**PCM Formats (for streaming):**
- `pcm_16000` - 16kHz PCM
- `pcm_22050` - 22.05kHz PCM
- `pcm_24000` - 24kHz PCM
- `pcm_44100` - 44.1kHz PCM

## Use Cases

### Voice Dubbing
```go
// Convert foreign language audio to your voice library
resp, err := client.SpeechToSpeech().Convert(ctx, &elevenlabs.SpeechToSpeechRequest{
    VoiceID:               englishVoiceID,
    Audio:                 foreignAudio,
    RemoveBackgroundNoise: true,
})
```

### Voice Anonymization
```go
// Convert voice for privacy
resp, err := client.SpeechToSpeech().Convert(ctx, &elevenlabs.SpeechToSpeechRequest{
    VoiceID: anonymousVoiceID,
    Audio:   originalRecording,
})
```

### Real-time Voice Changer
```go
// Stream microphone through voice conversion
resp, err := client.SpeechToSpeech().ConvertStream(ctx, &elevenlabs.SpeechToSpeechRequest{
    VoiceID:      characterVoiceID,
    Audio:        microphoneStream,
    OutputFormat: "pcm_22050",
})
// Pipe to speakers
io.Copy(speakers, resp.Audio)
```
