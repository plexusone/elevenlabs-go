# Audio Isolation

Extract vocals and speech from audio, removing background noise and music.

## Basic Usage

```go
file, _ := os.Open("mixed_audio.mp3")
defer file.Close()

isolated, err := client.AudioIsolation().IsolateFile(ctx, file, "mixed_audio.mp3")
if err != nil {
    log.Fatal(err)
}

// Save isolated vocals
output, _ := os.Create("vocals_only.mp3")
io.Copy(output, isolated)
```

## Streaming Isolation

For real-time processing:

```go
isolated, err := client.AudioIsolation().IsolateStream(ctx, &elevenlabs.AudioIsolationRequest{
    Audio:    audioReader,
    Filename: "audio.mp3",
})
```

## Full Options

```go
isolated, err := client.AudioIsolation().Isolate(ctx, &elevenlabs.AudioIsolationRequest{
    Audio:    audioFile,
    Filename: "podcast_with_music.mp3",
})
```

## Use Cases

### Podcast Cleanup

Remove background music and enhance speech clarity:

```go
// Original podcast has background music
podcastFile, _ := os.Open("podcast_episode.mp3")

// Extract just the voices
cleanAudio, err := client.AudioIsolation().IsolateFile(ctx,
    podcastFile, "podcast_episode.mp3")

// Save clean version
output, _ := os.Create("podcast_clean.mp3")
io.Copy(output, cleanAudio)
```

### Interview Processing

Extract speech from noisy interview recordings:

```go
// Field interview with ambient noise
interviewFile, _ := os.Open("street_interview.mp3")

// Isolate speaker voices
voices, err := client.AudioIsolation().IsolateFile(ctx,
    interviewFile, "street_interview.mp3")

// Use clean audio for transcription
result, _ := client.SpeechToText().Transcribe(ctx, &elevenlabs.TranscriptionRequest{
    File:     voices,
    Filename: "clean_interview.mp3",
})
```

### Music Vocal Extraction

Extract vocals from songs:

```go
songFile, _ := os.Open("song.mp3")

vocals, err := client.AudioIsolation().IsolateFile(ctx, songFile, "song.mp3")

// Save isolated vocals
vocalFile, _ := os.Create("vocals.mp3")
io.Copy(vocalFile, vocals)
```

### Video Audio Cleanup

Clean up audio tracks from video:

```go
// Extract audio from video (using ffmpeg externally)
// ffmpeg -i video.mp4 -vn audio.mp3

audioFile, _ := os.Open("audio.mp3")

cleanAudio, err := client.AudioIsolation().IsolateFile(ctx, audioFile, "audio.mp3")

// Save and remix back into video
cleanFile, _ := os.Create("clean_audio.mp3")
io.Copy(cleanFile, cleanAudio)
```

### Pre-processing for Voice Cloning

Get clean voice samples for cloning:

```go
// Sample may have background noise
sampleFile, _ := os.Open("voice_sample.mp3")

// Clean it up
cleanSample, err := client.AudioIsolation().IsolateFile(ctx,
    sampleFile, "voice_sample.mp3")

// Use clean sample for voice training
// (Professional Voice Cloning API)
```

## Pipeline Example

Combine with other services:

```go
// 1. Isolate vocals from noisy audio
file, _ := os.Open("noisy_recording.mp3")
clean, _ := client.AudioIsolation().IsolateFile(ctx, file, "noisy_recording.mp3")

// 2. Save to temp file for transcription
tmpFile, _ := os.CreateTemp("", "clean-*.mp3")
io.Copy(tmpFile, clean)
tmpFile.Seek(0, 0)

// 3. Transcribe the clean audio
transcript, _ := client.SpeechToText().Transcribe(ctx, &elevenlabs.TranscriptionRequest{
    File:     tmpFile,
    Filename: "clean.mp3",
})

// 4. Get word-level timestamps
tmpFile.Seek(0, 0)
alignment, _ := client.ForcedAlignment().AlignFile(ctx, tmpFile, "clean.mp3", transcript.Text)
```

## Supported Audio Formats

- MP3
- WAV
- M4A
- FLAC
- OGG
- WEBM

## Best Practices

1. **Use for noisy recordings** - Most effective when there's clear separation between voice and background
2. **Chain with transcription** - Clean audio produces better transcription results
3. **Save original files** - Keep originals as isolation is lossy
4. **Test on samples first** - Results vary based on audio content
