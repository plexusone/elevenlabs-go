# Forced Alignment

Get precise word-level and character-level timestamps for audio when you already have the transcript.

## Basic Usage

```go
file, _ := os.Open("speech.mp3")
defer file.Close()

result, err := client.ForcedAlignment().AlignFile(ctx, file, "speech.mp3",
    "The text that was spoken in the audio")
if err != nil {
    log.Fatal(err)
}

for _, word := range result.Words {
    fmt.Printf("%s: %.2fs - %.2fs\n", word.Text, word.Start, word.End)
}
```

## Full Options

```go
result, err := client.ForcedAlignment().Align(ctx, &elevenlabs.ForcedAlignmentRequest{
    File:     audioFile,
    Filename: "narration.mp3",
    Text:     "The complete transcript of the audio file",
})
```

## Response Structure

```go
type ForcedAlignmentResponse struct {
    Loss       float64              // Overall alignment loss/confidence
    Words      []AlignmentWord      // Word-level timestamps
    Characters []AlignmentCharacter // Character-level timestamps
}

type AlignmentWord struct {
    Text  string  // The word
    Start float64 // Start time in seconds
    End   float64 // End time in seconds
    Loss  float64 // Alignment confidence for this word
}

type AlignmentCharacter struct {
    Text  string  // The character
    Start float64 // Start time in seconds
    End   float64 // End time in seconds
}
```

## Use Cases

### Karaoke-Style Subtitles

```go
result, err := client.ForcedAlignment().AlignFile(ctx, audioFile, "song.mp3", lyrics)

// Highlight words as they're spoken
for _, word := range result.Words {
    fmt.Printf("At %.2fs, highlight: %s\n", word.Start, word.Text)
}
```

### Video Caption Sync

```go
// Align narration with known script
result, err := client.ForcedAlignment().AlignFile(ctx,
    narrationFile, "narration.mp3", script)

// Generate precise captions
for _, word := range result.Words {
    caption := Caption{
        Text:      word.Text,
        StartTime: word.Start,
        EndTime:   word.End,
    }
    captions = append(captions, caption)
}
```

### Audiobook Chapter Markers

```go
// Split transcript into chapters
chapters := []string{
    "Chapter one begins here...",
    "Chapter two continues...",
}

var chapterMarkers []float64
currentPos := 0

result, err := client.ForcedAlignment().AlignFile(ctx, bookFile, "book.mp3", fullText)

// Find chapter start times
for _, word := range result.Words {
    // Check if this word starts a new chapter
    for i, chapter := range chapters {
        if strings.HasPrefix(chapter, word.Text) {
            chapterMarkers = append(chapterMarkers, word.Start)
        }
    }
}
```

### Quality Check for TTS

```go
// Generate speech
audio, _ := client.TextToSpeech().Simple(ctx, voiceID, text)

// Save audio
audioFile, _ := os.CreateTemp("", "tts-*.mp3")
io.Copy(audioFile, audio)
audioFile.Seek(0, 0)

// Verify alignment
result, _ := client.ForcedAlignment().AlignFile(ctx, audioFile, "tts.mp3", text)

// Check alignment quality
if result.Loss > 0.5 {
    fmt.Println("Warning: Poor alignment - audio may not match text well")
}
```

## Alignment Loss

The `Loss` field indicates alignment confidence:

| Loss Value | Quality |
|------------|---------|
| 0.0 - 0.1 | Excellent alignment |
| 0.1 - 0.3 | Good alignment |
| 0.3 - 0.5 | Acceptable |
| > 0.5 | Poor - check audio/text match |

## Supported Audio Formats

- MP3
- WAV
- M4A
- FLAC
- OGG

## Best Practices

1. **Ensure text matches audio exactly** - Mismatched text will increase loss
2. **Use clean audio** - Background noise affects alignment accuracy
3. **Check the Loss value** - High loss indicates poor alignment
4. **Use character-level for precise sync** - Useful for karaoke effects
