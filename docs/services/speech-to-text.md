# Speech-to-Text

Transcribe audio files with optional speaker diarization.

## Basic Usage

```go
// Transcribe from URL
result, err := client.SpeechToText().TranscribeURL(ctx, "https://example.com/audio.mp3")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Text: %s\n", result.Text)
fmt.Printf("Language: %s\n", result.LanguageCode)
```

## Transcribe with File Upload

```go
file, _ := os.Open("audio.mp3")
defer file.Close()

result, err := client.SpeechToText().Transcribe(ctx, &elevenlabs.TranscriptionRequest{
    File:     file,
    Filename: "audio.mp3",
    ModelID:  "scribe_v1",
})
```

## Speaker Diarization

Identify different speakers in the audio:

```go
result, err := client.SpeechToText().TranscribeWithDiarization(ctx, audioURL)
if err != nil {
    log.Fatal(err)
}

for _, word := range result.Words {
    fmt.Printf("[%s] %s (%.2fs - %.2fs)\n",
        word.Speaker, word.Text, word.Start, word.End)
}
```

## Full Options

```go
result, err := client.SpeechToText().Transcribe(ctx, &elevenlabs.TranscriptionRequest{
    File:              file,
    Filename:          "interview.mp3",
    ModelID:           "scribe_v1",
    LanguageCode:      "en",           // ISO 639-1 code
    Diarize:           true,           // Enable speaker detection
    TagAudioEvents:    true,           // Tag laughter, music, etc.
    NumSpeakers:       2,              // Expected number of speakers
})
```

## Request Options

| Option | Type | Description |
|--------|------|-------------|
| `File` | io.Reader | Audio file to transcribe |
| `Filename` | string | Name of the audio file |
| `AudioURL` | string | URL to audio (alternative to file) |
| `ModelID` | string | Transcription model (default: scribe_v1) |
| `LanguageCode` | string | ISO 639-1 language code |
| `Diarize` | bool | Enable speaker diarization |
| `TagAudioEvents` | bool | Tag non-speech audio events |
| `NumSpeakers` | int | Expected number of speakers |

## Response Structure

```go
type TranscriptionResponse struct {
    Text         string              // Full transcription text
    LanguageCode string              // Detected language
    Words        []TranscriptionWord // Word-level timestamps
}

type TranscriptionWord struct {
    Text    string  // The word
    Start   float64 // Start time in seconds
    End     float64 // End time in seconds
    Speaker string  // Speaker ID (if diarization enabled)
}
```

## Use Cases

### Meeting Transcription

```go
result, err := client.SpeechToText().Transcribe(ctx, &elevenlabs.TranscriptionRequest{
    File:        meetingFile,
    Filename:    "meeting.mp3",
    Diarize:     true,
    NumSpeakers: 4,
})

// Group by speaker
speakers := make(map[string][]string)
for _, word := range result.Words {
    speakers[word.Speaker] = append(speakers[word.Speaker], word.Text)
}
```

### Subtitle Generation

```go
result, err := client.SpeechToText().TranscribeURL(ctx, videoAudioURL)

// Generate SRT format
for i, word := range result.Words {
    fmt.Printf("%d\n", i+1)
    fmt.Printf("%s --> %s\n", formatTime(word.Start), formatTime(word.End))
    fmt.Printf("%s\n\n", word.Text)
}
```

### Podcast Processing

```go
// Transcribe podcast episode
result, err := client.SpeechToText().Transcribe(ctx, &elevenlabs.TranscriptionRequest{
    File:           podcastFile,
    Filename:       "episode.mp3",
    Diarize:        true,
    TagAudioEvents: true,  // Detect music, laughter, etc.
})
```

## Supported Audio Formats

- MP3
- WAV
- M4A
- FLAC
- OGG
- WEBM

## Best Practices

1. **Use diarization for multi-speaker content** - Interviews, meetings, podcasts
2. **Specify language** when known for better accuracy
3. **Set expected speaker count** for more accurate diarization
4. **Enable audio event tagging** for richer metadata
