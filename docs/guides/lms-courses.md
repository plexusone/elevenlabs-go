# Creating LMS/Udemy Courses

A complete guide to using go-elevenlabs for creating professional online course audio.

## Overview

This SDK provides everything you need to produce professional course audio:

| Feature | Service | Use Case |
|---------|---------|----------|
| Narration | Text-to-Speech | Generate voiceovers from scripts |
| Voice Selection | Voices | Choose the right narrator |
| Course Structure | Projects | Organize into chapters |
| Technical Terms | Pronunciation | Handle jargon correctly |
| Sound Effects | SoundEffects | Intros, transitions |
| Background Music | Music | Generate custom background tracks |
| Multi-Language | Dubbing | Translate to other languages |
| Usage Tracking | User | Monitor character consumption |

## Workflow

### 1. Set Up Your Project

```go
package main

import (
    "context"
    "log"

    elevenlabs "github.com/agentplexus/go-elevenlabs"
)

func main() {
    client, err := elevenlabs.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Check available characters
    sub, _ := client.User().GetSubscription(ctx)
    fmt.Printf("Characters available: %d\n", sub.CharactersRemaining())
}
```

### 2. Create Pronunciation Dictionary

Handle technical terms before generating audio:

```json
// terms.json
[
  {"grapheme": "API", "alias": "A P I"},
  {"grapheme": "SDK", "alias": "S D K"},
  {"grapheme": "JSON", "alias": "jay son"},
  {"grapheme": "CLI", "alias": "C L I"},
  {"grapheme": "GUI", "alias": "gooey"},
  {"grapheme": "SQL", "alias": "sequel"},
  {"grapheme": "OAuth", "alias": "oh auth"},
  {"grapheme": "nginx", "alias": "engine X"},
  {"grapheme": "kubectl", "alias": "kube control"}
]
```

```go
dict, err := client.Pronunciation().CreateFromJSON(ctx, "Tech Terms", "terms.json")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created dictionary: %s\n", dict.ID)
```

### 3. Select Your Voice

```go
voices, _ := client.Voices().List(ctx)

// Find a voice suitable for narration
for _, v := range voices {
    fmt.Printf("%s: %s\n", v.VoiceID, v.Name)
}

// Popular choices for courses:
// - Rachel (21m00Tcm4TlvDq8ikWAM): Calm, clear
// - Antoni (ErXwobaYiN019PkySvjV): Professional, warm
```

### 4. Generate Chapter Audio

```go
chapters := []struct {
    Title   string
    Script  string
}{
    {
        Title:  "01-introduction",
        Script: "Welcome to this course on building APIs with Go...",
    },
    {
        Title:  "02-setup",
        Script: "In this chapter, we'll set up our development environment...",
    },
}

voiceID := "21m00Tcm4TlvDq8ikWAM"

for _, ch := range chapters {
    audio, err := client.TextToSpeech().Generate(ctx, &elevenlabs.TTSRequest{
        VoiceID: voiceID,
        Text:    ch.Script,
        ModelID: "eleven_multilingual_v2",
        VoiceSettings: &elevenlabs.VoiceSettings{
            Stability:       0.6,  // Slightly more consistent for courses
            SimilarityBoost: 0.75,
        },
    })
    if err != nil {
        log.Printf("Failed to generate %s: %v", ch.Title, err)
        continue
    }

    filename := fmt.Sprintf("%s.mp3", ch.Title)
    f, _ := os.Create(filename)
    io.Copy(f, audio.Audio)
    f.Close()

    fmt.Printf("Generated: %s\n", filename)
}
```

### 5. Create Sound Effects & Music

```go
// Intro jingle
intro, _ := client.SoundEffects().Simple(ctx, "professional podcast intro with subtle music")

// Chapter transition
transition, _ := client.SoundEffects().Simple(ctx, "soft whoosh transition sound")

// Save them
saveAudio(intro, "sfx-intro.mp3")
saveAudio(transition, "sfx-transition.mp3")

// Generate background music with the Music service
background, _ := client.Music().Generate(ctx, &elevenlabs.MusicRequest{
    Prompt:            "calm ambient background music for educational content",
    DurationMs:        300000, // 5 minutes
    ForceInstrumental: true,
})
saveAudio(background.Audio, "background-music.mp3")

// For more control, use composition plans
plan, _ := client.Music().GeneratePlan(ctx, &elevenlabs.CompositionPlanRequest{
    Prompt:     "educational video intro music, friendly and professional",
    DurationMs: 10000, // 10 seconds
})
introMusic, _ := client.Music().GenerateDetailed(ctx, &elevenlabs.MusicDetailedRequest{
    CompositionPlan:   plan,
    ForceInstrumental: true,
})
saveAudio(introMusic.Audio, "intro-music.mp3")
```

### 6. Using Projects for Long-Form Content

For complete courses, use the Projects/Studio feature:

```go
// Create project
project, err := client.Projects().Create(ctx, &elevenlabs.CreateProjectRequest{
    Name:                    "Go API Development Course",
    Description:             "Complete guide to building REST APIs in Go",
    Language:                "en",
    DefaultModelID:          "eleven_multilingual_v2",
    DefaultParagraphVoiceID: "21m00Tcm4TlvDq8ikWAM",
    DefaultTitleVoiceID:     "21m00Tcm4TlvDq8ikWAM",
    QualityPreset:           "high",
})

// Chapters are added via web UI or content upload
// Then convert:
err = client.Projects().Convert(ctx, project.ProjectID)

// Download when complete
snapshots, _ := client.Projects().ListSnapshots(ctx, project.ProjectID)
if len(snapshots) > 0 {
    reader, _ := client.Projects().DownloadSnapshotArchive(ctx,
        project.ProjectID, snapshots[0].ProjectSnapshotID)

    f, _ := os.Create("course.zip")
    io.Copy(f, reader)
    f.Close()
}
```

### 7. Translate to Other Languages

```go
// Dub English course to Spanish
dub, err := client.Dubbing().Create(ctx, &elevenlabs.DubbingRequest{
    SourceURL:      "https://storage.example.com/course-intro.mp4",
    TargetLanguage: "es",
    Name:           "Go API Course - Spanish",
})

// Wait for completion, then download
// ...
```

## Best Practices

### Script Writing

1. **Write for audio** - Use conversational language
2. **Spell out acronyms** in pronunciation dictionary
3. **Add pauses** with punctuation (periods, commas)
4. **Keep sentences short** - Easier to follow

### Voice Settings for Courses

Use the built-in presets for different platforms:

```go
// Platform-specific presets
settings := elevenlabs.VoiceSettingsForUdemy()     // Neutral, clear, consistent
settings := elevenlabs.VoiceSettingsForCoursera()  // Slightly expressive, engaging
settings := elevenlabs.VoiceSettingsForEdX()       // Very stable, highly intelligible

// Use with TTS
audio, err := client.TextToSpeech().Generate(ctx, &elevenlabs.TTSRequest{
    VoiceID:       voiceID,
    Text:          script,
    VoiceSettings: elevenlabs.VoiceSettingsForUdemy(),
})
```

See [Voice Settings Presets](../utilities/voicesettings.md) for all available presets.

### Quality Presets

| Content Type | Recommended Preset |
|--------------|-------------------|
| Preview/Draft | `standard` |
| Final Course | `high` or `ultra` |
| Podcast | `high` |
| Audiobook | `ultra` |

### Cost Optimization

```go
// Check before generating
sub, _ := client.User().GetSubscription(ctx)
scriptLength := len(script)

if sub.CharactersRemaining() < scriptLength {
    log.Fatal("Insufficient characters")
}

// For drafts, use shorter clips
// Only generate full audio for final version
```

## Complete Example

See the [examples directory](https://github.com/agentplexus/go-elevenlabs/tree/main/examples) for complete working examples.

```go
// Full course generation workflow
func generateCourse(client *elevenlabs.Client, courseName string, chapters []Chapter) error {
    ctx := context.Background()

    // 1. Create pronunciation dictionary
    dict, err := client.Pronunciation().CreateFromJSON(ctx, courseName+" Terms", "terms.json")
    if err != nil {
        return err
    }
    log.Printf("Dictionary: %s", dict.ID)

    // 2. Generate intro sound effect
    intro, _ := client.SoundEffects().Simple(ctx, "professional course intro")
    saveAudio(intro, "course-intro.mp3")

    // 3. Generate each chapter
    for i, ch := range chapters {
        log.Printf("Generating chapter %d: %s", i+1, ch.Title)

        audio, err := client.TextToSpeech().Simple(ctx, voiceID, ch.Script)
        if err != nil {
            return fmt.Errorf("chapter %d: %w", i+1, err)
        }

        filename := fmt.Sprintf("chapter-%02d.mp3", i+1)
        saveAudio(audio, filename)
    }

    log.Println("Course generation complete!")
    return nil
}
```
