# Voice Design

Generate custom AI voices with specific characteristics like gender, age, and accent.

## Basic Usage

```go
// Generate a voice preview
resp, err := client.VoiceDesign().Simple(ctx,
    elevenlabs.VoiceGenderFemale,
    elevenlabs.VoiceAgeYoung,
    elevenlabs.VoiceAccentAmerican,
    "This is a preview of the generated voice. It should be at least one hundred characters long for the best quality results.",
)
if err != nil {
    log.Fatal(err)
}

// Listen to the preview
f, _ := os.Create("voice_preview.mp3")
io.Copy(f, resp.Audio)
```

## Full Options

```go
resp, err := client.VoiceDesign().GeneratePreview(ctx, &elevenlabs.VoiceDesignRequest{
    Gender:         elevenlabs.VoiceGenderFemale,
    Age:            elevenlabs.VoiceAgeYoung,
    Accent:         elevenlabs.VoiceAccentBritish,
    AccentStrength: 1.5,  // 0.3 to 2.0
    Text:           "This is a preview text that must be between one hundred and one thousand characters long for optimal voice generation quality.",
})
```

## Save Generated Voice

Once you like a preview, save it to your voice library:

```go
// Generate preview
preview, _ := client.VoiceDesign().GeneratePreview(ctx, &elevenlabs.VoiceDesignRequest{
    Gender: elevenlabs.VoiceGenderMale,
    Age:    elevenlabs.VoiceAgeMiddleAged,
    Accent: elevenlabs.VoiceAccentBritish,
    Text:   sampleText,
})

// Save to library
voice, err := client.VoiceDesign().SaveVoice(ctx, &elevenlabs.SaveVoiceRequest{
    GeneratedVoiceID: preview.GeneratedVoiceID,
    VoiceName:        "British Narrator",
    VoiceDescription: "Professional British male voice for narration",
    Labels: map[string]string{
        "use_case": "narration",
        "style":    "professional",
    },
})

fmt.Printf("Saved voice ID: %s\n", voice.VoiceID)
```

## Voice Options

### Gender

```go
elevenlabs.VoiceGenderFemale
elevenlabs.VoiceGenderMale
```

### Age

```go
elevenlabs.VoiceAgeYoung       // Young adult
elevenlabs.VoiceAgeMiddleAged  // Middle-aged
elevenlabs.VoiceAgeOld         // Elderly
```

### Accent

```go
elevenlabs.VoiceAccentAmerican
elevenlabs.VoiceAccentBritish
elevenlabs.VoiceAccentAustralian
elevenlabs.VoiceAccentIndian
elevenlabs.VoiceAccentAfrican
```

### Accent Strength

| Value | Effect |
|-------|--------|
| 0.3 | Subtle accent |
| 1.0 | Normal (default) |
| 1.5 | Strong accent |
| 2.0 | Very strong accent |

## Request Structure

```go
type VoiceDesignRequest struct {
    Gender         VoiceGender  // Required
    Age            VoiceAge     // Required
    Accent         VoiceAccent  // Required
    AccentStrength float64      // 0.3 to 2.0 (default: 1.0)
    Text           string       // 100-1000 characters
}

type VoiceDesignResponse struct {
    Audio            io.Reader // Preview audio
    GeneratedVoiceID string    // ID to save the voice
}

type SaveVoiceRequest struct {
    GeneratedVoiceID string            // From preview response
    VoiceName        string            // Name for saved voice
    VoiceDescription string            // Optional description
    Labels           map[string]string // Optional metadata
}
```

## Use Cases

### Create Character Voices

```go
characters := []struct {
    Name   string
    Gender elevenlabs.VoiceGender
    Age    elevenlabs.VoiceAge
    Accent elevenlabs.VoiceAccent
}{
    {"Hero", elevenlabs.VoiceGenderMale, elevenlabs.VoiceAgeYoung, elevenlabs.VoiceAccentAmerican},
    {"Mentor", elevenlabs.VoiceGenderMale, elevenlabs.VoiceAgeOld, elevenlabs.VoiceAccentBritish},
    {"Sidekick", elevenlabs.VoiceGenderFemale, elevenlabs.VoiceAgeYoung, elevenlabs.VoiceAccentAustralian},
}

sampleText := "This is a sample of how this character will sound in the story. The text needs to be long enough for quality generation."

for _, char := range characters {
    preview, _ := client.VoiceDesign().Simple(ctx, char.Gender, char.Age, char.Accent, sampleText)

    // Save if satisfied
    voice, _ := client.VoiceDesign().SaveVoice(ctx, &elevenlabs.SaveVoiceRequest{
        GeneratedVoiceID: preview.GeneratedVoiceID,
        VoiceName:        char.Name,
        Labels:           map[string]string{"project": "audiobook"},
    })

    fmt.Printf("Created %s voice: %s\n", char.Name, voice.VoiceID)
}
```

### A/B Test Voices

```go
// Generate multiple previews with same parameters
var previews []*elevenlabs.VoiceDesignResponse

for i := 0; i < 3; i++ {
    preview, _ := client.VoiceDesign().GeneratePreview(ctx, &elevenlabs.VoiceDesignRequest{
        Gender: elevenlabs.VoiceGenderFemale,
        Age:    elevenlabs.VoiceAgeYoung,
        Accent: elevenlabs.VoiceAccentAmerican,
        Text:   sampleText,
    })
    previews = append(previews, preview)

    // Save preview audio for comparison
    f, _ := os.Create(fmt.Sprintf("preview_%d.mp3", i))
    io.Copy(f, preview.Audio)
    f.Close()
}

// Listen to all previews and save the best one
```

### Brand Voice Creation

```go
// Define brand voice characteristics
brandVoice := elevenlabs.VoiceDesignRequest{
    Gender:         elevenlabs.VoiceGenderFemale,
    Age:            elevenlabs.VoiceAgeMiddleAged,
    Accent:         elevenlabs.VoiceAccentAmerican,
    AccentStrength: 0.5,  // Subtle accent for professionalism
    Text:           "Welcome to our service. We're here to help you succeed. Our team is dedicated to providing the best experience possible for all our customers.",
}

preview, _ := client.VoiceDesign().GeneratePreview(ctx, &brandVoice)

voice, _ := client.VoiceDesign().SaveVoice(ctx, &elevenlabs.SaveVoiceRequest{
    GeneratedVoiceID: preview.GeneratedVoiceID,
    VoiceName:        "Brand Voice - Main",
    VoiceDescription: "Official brand voice for customer communications",
    Labels: map[string]string{
        "brand":    "true",
        "approved": "true",
        "use_case": "customer_service",
    },
})
```

## Text Requirements

The preview text must be:

- **Minimum:** 100 characters
- **Maximum:** 1,000 characters
- **Content:** Representative of intended use

Good preview text:
```go
text := `This is a sample of how this voice will sound when reading longer content.
The text should be representative of the actual content you plan to generate,
including the style, tone, and type of vocabulary you'll be using.`
```

## Best Practices

1. **Use representative text** - Preview text should match your intended use case
2. **Generate multiple previews** - Each generation is unique; try several
3. **Test accent strength** - Adjust for natural-sounding results
4. **Add descriptive labels** - Makes organizing voices easier
5. **Save good voices immediately** - Generated voice IDs may expire
