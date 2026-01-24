# Sound Effects

Generate sound effects from text descriptions.

## Basic Usage

```go
audio, err := client.SoundEffects().Simple(ctx, "thunder and rain storm")
if err != nil {
    log.Fatal(err)
}

f, _ := os.Create("thunder.mp3")
io.Copy(f, audio)
```

## Full Control

```go
resp, err := client.SoundEffects().Generate(ctx, &elevenlabs.SoundEffectRequest{
    Text:            "car engine starting and revving",
    DurationSeconds: 5.0,      // 0.5 to 30 seconds
    PromptInfluence: 0.5,      // 0.0 to 1.0
    OutputFormat:    "mp3_44100_128",
})
```

## Looping Sound Effects

Create seamlessly looping audio for backgrounds:

```go
audio, err := client.SoundEffects().GenerateLoop(ctx,
    "gentle rain on window",
    10.0,  // duration in seconds
)
```

## Request Options

| Option | Range | Description |
|--------|-------|-------------|
| `Text` | required | Description of the sound effect |
| `DurationSeconds` | 0.5-30 | Target duration (auto if not set) |
| `PromptInfluence` | 0.0-1.0 | How closely to follow the prompt |
| `Loop` | bool | Create seamless loop |
| `OutputFormat` | string | Audio format |

## Prompt Influence

- **Low (0.0-0.3)**: More variation, creative interpretation
- **Medium (0.3-0.6)**: Balanced (default: 0.3)
- **High (0.6-1.0)**: Strictly follows prompt

## Example Prompts

### Nature
```go
"gentle rain on a tin roof"
"thunderstorm with distant lightning"
"ocean waves on a beach"
"wind blowing through trees"
"birds chirping in a forest"
```

### Urban
```go
"busy city traffic"
"car horn honking"
"subway train arriving"
"crowd murmuring in a cafe"
```

### Technology
```go
"computer keyboard typing"
"notification chime"
"sci-fi door opening"
"robot powering up"
```

### Music/Transitions
```go
"dramatic orchestral hit"
"soft piano transition"
"whoosh transition sound"
"upbeat intro jingle"
```

## Use Cases

### Course Production

```go
// Intro sound
intro, _ := client.SoundEffects().Simple(ctx, "professional podcast intro jingle")

// Transition
transition, _ := client.SoundEffects().Simple(ctx, "soft whoosh transition")

// Background ambience (looping)
ambience, _ := client.SoundEffects().GenerateLoop(ctx, "quiet office ambience", 30)
```

### Video Production

```go
// Action sounds
punch, _ := client.SoundEffects().Simple(ctx, "punch impact sound")

// Ambient backgrounds
forest, _ := client.SoundEffects().GenerateLoop(ctx, "peaceful forest ambience", 60)
```

## Best Practices

1. **Be specific** - "car engine starting cold" vs just "car"
2. **Include context** - "footsteps on wooden floor" vs "footsteps"
3. **Use loops for backgrounds** - Enables seamless repetition
4. **Test prompt influence** - Adjust based on desired creativity
