# Music

Generate music from text prompts.

## Basic Usage

```go
audio, err := client.Music().Simple(ctx, "upbeat electronic music for a tech video")
if err != nil {
    log.Fatal(err)
}

f, _ := os.Create("music.mp3")
io.Copy(f, audio)
```

## Full Options

```go
resp, err := client.Music().Generate(ctx, &elevenlabs.MusicRequest{
    Prompt:            "calm piano melody with soft strings",
    DurationMs:        30000,  // 30 seconds
    ForceInstrumental: true,   // No vocals
    Seed:              12345,  // For reproducibility
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Song ID: %s\n", resp.SongID)

f, _ := os.Create("calm_piano.mp3")
io.Copy(f, resp.Audio)
```

## Instrumental Only

Ensure the generated music has no vocals:

```go
audio, err := client.Music().GenerateInstrumental(ctx,
    "epic orchestral music for movie trailer",
    60000,  // 60 seconds
)
```

## Streaming

For real-time playback:

```go
resp, err := client.Music().GenerateStream(ctx, &elevenlabs.MusicRequest{
    Prompt:     "lofi hip hop beats",
    DurationMs: 120000,  // 2 minutes
})
```

## Request Options

| Option | Type | Description |
|--------|------|-------------|
| `Prompt` | string | Text description of the music |
| `DurationMs` | int | Duration in milliseconds (3000-600000) |
| `ForceInstrumental` | bool | Ensure no vocals |
| `Seed` | int | For reproducible generation |

## Response Structure

```go
type MusicRequest struct {
    Prompt            string
    DurationMs        int
    ForceInstrumental bool
    Seed              int
}

type MusicResponse struct {
    Audio  io.Reader // Generated music
    SongID string    // Unique identifier
}
```

## Duration Guidelines

| Duration | Use Case |
|----------|----------|
| 3-10 seconds | Sound logos, jingles |
| 10-30 seconds | Intro/outro music |
| 30-60 seconds | Background music |
| 1-5 minutes | Full tracks |
| 5-10 minutes | Extended ambient |

## Prompt Examples

### By Genre

```go
// Electronic
"upbeat EDM with synthesizers and heavy bass drops"
"ambient electronic music with soft pads"
"retro 80s synthwave with arpeggios"

// Orchestral
"epic cinematic orchestra with brass and strings"
"soft classical piano solo"
"dramatic film score with timpani"

// Modern
"lofi hip hop beats to study to"
"indie folk with acoustic guitar"
"jazz trio with piano bass and drums"

// Ambient
"peaceful nature soundscape with birds"
"space ambient with ethereal pads"
"meditation music with singing bowls"
```

### By Mood

```go
// Energetic
"high energy workout music with driving beat"
"exciting action music for video games"
"uplifting pop rock anthem"

// Calm
"relaxing spa music with soft piano"
"gentle lullaby for sleep"
"peaceful morning coffee music"

// Dramatic
"tense thriller soundtrack"
"emotional sad piano piece"
"triumphant victory fanfare"
```

### By Use Case

```go
// Video Content
"YouTube intro music, modern and catchy, 10 seconds"
"podcast background music, subtle and professional"
"tutorial video music, friendly and upbeat"

// Business
"corporate presentation background music"
"hold music, pleasant and non-intrusive"
"product launch reveal music"

// Gaming
"RPG exploration theme, adventurous"
"boss battle music, intense and fast"
"game menu music, mysterious ambient"
```

## Use Cases

### Video Intro Music

```go
intro, err := client.Music().Generate(ctx, &elevenlabs.MusicRequest{
    Prompt:            "modern tech startup intro jingle, professional and innovative",
    DurationMs:        8000,  // 8 seconds
    ForceInstrumental: true,
})
```

### Podcast Background

```go
background, err := client.Music().Generate(ctx, &elevenlabs.MusicRequest{
    Prompt:            "subtle podcast background music, warm and conversational",
    DurationMs:        300000,  // 5 minutes
    ForceInstrumental: true,
})
```

### Course Content

```go
// Intro music
intro, _ := client.Music().GenerateInstrumental(ctx,
    "educational video intro, friendly and engaging", 5000)

// Transition music
transition, _ := client.Music().GenerateInstrumental(ctx,
    "soft transition sound, brief whoosh with melody", 2000)

// Background for explanations
background, _ := client.Music().GenerateInstrumental(ctx,
    "thinking music, curious and light", 60000)
```

### Game Audio

```go
// Menu theme
menuTheme, _ := client.Music().Generate(ctx, &elevenlabs.MusicRequest{
    Prompt:     "fantasy RPG main menu theme, epic and adventurous",
    DurationMs: 120000,
    Seed:       42,  // Reproducible for version control
})

// Battle music
battleMusic, _ := client.Music().Generate(ctx, &elevenlabs.MusicRequest{
    Prompt:            "intense battle music, fast drums and aggressive strings",
    DurationMs:        90000,
    ForceInstrumental: true,
})
```

### Reproducible Generation

Use seeds for consistent results:

```go
// Generate same music twice
seed := 12345

music1, _ := client.Music().Generate(ctx, &elevenlabs.MusicRequest{
    Prompt:     "calm acoustic guitar",
    DurationMs: 30000,
    Seed:       seed,
})

music2, _ := client.Music().Generate(ctx, &elevenlabs.MusicRequest{
    Prompt:     "calm acoustic guitar",
    DurationMs: 30000,
    Seed:       seed,
})

// music1 and music2 will be identical
```

## Composition Plans

For fine-grained control over song structure, generate a composition plan first:

```go
// Generate a plan from a prompt
plan, err := client.Music().GeneratePlan(ctx, &elevenlabs.CompositionPlanRequest{
    Prompt:     "upbeat pop song about summer",
    DurationMs: 180000, // 3 minutes
})
if err != nil {
    log.Fatal(err)
}

// Modify the plan if needed
plan.Sections[0].Lines = []string{"Custom lyrics for the intro"}
plan.PositiveGlobalStyles = append(plan.PositiveGlobalStyles, "energetic")

// Generate music from the modified plan
resp, err := client.Music().GenerateDetailed(ctx, &elevenlabs.MusicDetailedRequest{
    CompositionPlan: plan,
    WithTimestamps:  true,
})
```

### Composition Plan Structure

```go
type CompositionPlan struct {
    PositiveGlobalStyles []string     // Styles present throughout
    NegativeGlobalStyles []string     // Styles to avoid
    Sections             []SongSection
}

type SongSection struct {
    SectionName         string   // "intro", "verse", "chorus"
    DurationMs          int      // 3000-120000 per section
    Lines               []string // Lyrics (max 200 chars per line)
    PositiveLocalStyles []string // Styles for this section
    NegativeLocalStyles []string // Styles to avoid in this section
}
```

## Detailed Generation

Generate music with detailed options and metadata:

```go
// With a simple prompt
resp, err := client.Music().GenerateDetailed(ctx, &elevenlabs.MusicDetailedRequest{
    Prompt:            "epic orchestral music",
    DurationMs:        60000,
    ForceInstrumental: true,
    WithTimestamps:    true,  // Get word timestamps
})

// With a composition plan (see above)
resp, err := client.Music().GenerateDetailed(ctx, &elevenlabs.MusicDetailedRequest{
    CompositionPlan: plan,
})
```

## Stem Separation

Separate a song into individual stems (vocals, drums, bass, etc.):

```go
f, _ := os.Open("song.mp3")
defer f.Close()

stems, err := client.Music().SeparateStems(ctx, &elevenlabs.StemSeparationRequest{
    File:     f,
    Filename: "song.mp3",
    // Options: "two_stems_v1" (vocals + music)
    //          "six_stems_v1" (vocals, drums, bass, other - default)
    StemVariation: "six_stems_v1",
})
if err != nil {
    log.Fatal(err)
}

// Save the stems (returned as a zip file)
output, _ := os.Create("stems.zip")
io.Copy(output, stems)
```

### Stem Variation Options

| Option | Output |
|--------|--------|
| `two_stems_v1` | Vocals + instrumental |
| `six_stems_v1` | Vocals, drums, bass, and other (default) |

## Best Practices

1. **Be specific** - "upbeat jazz piano trio" vs just "jazz"
2. **Include tempo hints** - "fast", "slow", "moderate tempo"
3. **Specify instruments** - "acoustic guitar and violin" for precise results
4. **Use ForceInstrumental** - When you don't want any vocals
5. **Test with seeds** - Find good seeds and save them for consistency
6. **Match duration to use case** - Don't generate 5 minutes for a 10-second intro
7. **Use composition plans** - For complex songs with multiple sections
8. **Stem separation** - Extract vocals for remixing or karaoke
