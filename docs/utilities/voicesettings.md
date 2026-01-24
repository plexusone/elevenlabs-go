# Voice Settings Presets

Pre-configured voice settings optimized for different platforms and use cases.

## Overview

Voice settings control how ElevenLabs synthesizes speech. The right settings depend on your platform and content type:

| Setting | Range | Effect |
|---------|-------|--------|
| Stability | 0.0-1.0 | Higher = more consistent, lower = more expressive |
| SimilarityBoost | 0.0-1.0 | Higher = closer to original voice |
| Style | 0.0-1.0 | Higher = more stylized/exaggerated |
| Speed | 0.25-4.0 | Playback speed multiplier |
| UseSpeakerBoost | bool | Enhanced clarity and presence |

## Available Presets

### Educational Platforms

#### Udemy
```go
settings := elevenlabs.VoiceSettingsForUdemy()
```

| Setting | Value | Rationale |
|---------|-------|-----------|
| Stability | 0.5 | Balanced - not robotic, not erratic |
| SimilarityBoost | 0.75 | Natural sound |
| Style | 0.05 | Minimal style to avoid flat delivery |
| Speed | 1.0 | Normal pace |
| UseSpeakerBoost | true | Clear audio |

**Best for:** Long-form courses, technical tutorials, consistent narration.

#### Coursera
```go
settings := elevenlabs.VoiceSettingsForCoursera()
```

| Setting | Value | Rationale |
|---------|-------|-----------|
| Stability | 0.7 | More consistent for academic content |
| SimilarityBoost | 0.85 | Professional sound |
| Style | 0.2 | Slight engagement |
| Speed | 1.0 | Normal pace |
| UseSpeakerBoost | true | Clear audio |

**Best for:** University-style courses, mixed media content, professional development.

#### edX
```go
settings := elevenlabs.VoiceSettingsForEdX()
```

| Setting | Value | Rationale |
|---------|-------|-----------|
| Stability | 0.8 | Very stable for academic rigor |
| SimilarityBoost | 0.9 | Highly consistent |
| Style | 0.15 | Subtle engagement without distraction |
| Speed | 1.05 | Slightly faster for dense content |
| UseSpeakerBoost | true | Maximum clarity |

**Best for:** Academic courses, technical subjects, multi-hour content.

### Social Media

#### Instagram
```go
settings := elevenlabs.VoiceSettingsForInstagram()
```

| Setting | Value | Rationale |
|---------|-------|-----------|
| Stability | 0.4 | More dynamic |
| SimilarityBoost | 0.85 | Polished sound |
| Style | 0.35 | Energetic but professional |
| Speed | 1.1 | Slightly faster for short attention spans |
| UseSpeakerBoost | true | Punchy audio |

**Best for:** Reels, brand content, product showcases, 15-60 second clips.

#### TikTok
```go
settings := elevenlabs.VoiceSettingsForTikTok()
```

| Setting | Value | Rationale |
|---------|-------|-----------|
| Stability | 0.3 | Expressive, attention-grabbing |
| SimilarityBoost | 0.85 | Consistent quality |
| Style | 0.45 | High energy |
| Speed | 1.15 | Fast pace for engagement |
| UseSpeakerBoost | true | Cuts through background music |

**Best for:** Short-form viral content, hook-driven intros, trend content.

#### YouTube
```go
settings := elevenlabs.VoiceSettingsForYouTube()
```

| Setting | Value | Rationale |
|---------|-------|-----------|
| Stability | 0.45 | Natural variation |
| SimilarityBoost | 0.8 | Human-like quality |
| Style | 0.2 | Engaging but sustainable |
| Speed | 1.05 | Slightly energized |
| UseSpeakerBoost | true | Professional production |

**Best for:** 5-20 minute videos, tutorials, reviews, educational content.

### Long-Form Audio

#### Podcast
```go
settings := elevenlabs.VoiceSettingsForPodcast()
```

| Setting | Value | Rationale |
|---------|-------|-----------|
| Stability | 0.55 | Conversational variation |
| SimilarityBoost | 0.75 | Natural, not synthetic |
| Style | 0.15 | Subtle personality |
| Speed | 1.0 | Relaxed listening pace |
| UseSpeakerBoost | true | Clean audio |

**Best for:** Podcast episodes, interview-style content, conversational narration.

#### Audiobook
```go
settings := elevenlabs.VoiceSettingsForAudiobook()
```

| Setting | Value | Rationale |
|---------|-------|-----------|
| Stability | 0.65 | Consistent for hours of listening |
| SimilarityBoost | 0.8 | Familiar, comfortable voice |
| Style | 0.1 | Minimal distraction |
| Speed | 0.95 | Slightly slower for comprehension |
| UseSpeakerBoost | true | Fatigue-free listening |

**Best for:** Books, long-form narration, bedtime stories, extended listening.

## Usage Examples

### Basic Usage
```go
audio, err := client.TextToSpeech().Generate(ctx, &elevenlabs.TTSRequest{
    VoiceID:       "21m00Tcm4TlvDq8ikWAM",
    Text:          "Welcome to this course!",
    VoiceSettings: elevenlabs.VoiceSettingsForUdemy(),
})
```

### With Simple Method
```go
// The Simple method uses default settings
// For custom settings, use Generate with a preset
audio, err := client.TextToSpeech().Generate(ctx, &elevenlabs.TTSRequest{
    VoiceID:       voiceID,
    Text:          script,
    ModelID:       "eleven_multilingual_v2",
    VoiceSettings: elevenlabs.VoiceSettingsForYouTube(),
})
```

### Customizing a Preset
```go
// Start with a preset and modify
settings := elevenlabs.VoiceSettingsForUdemy()
settings.Speed = 1.1  // Speed up slightly

audio, err := client.TextToSpeech().Generate(ctx, &elevenlabs.TTSRequest{
    VoiceID:       voiceID,
    Text:          script,
    VoiceSettings: settings,
})
```

## Choosing the Right Preset

| Content Type | Recommended Preset | Why |
|--------------|-------------------|-----|
| Online course | `VoiceSettingsForUdemy()` | Neutral, clear, won't fatigue listeners |
| Academic lecture | `VoiceSettingsForEdX()` | Stable, intelligible, professional |
| Short social clip | `VoiceSettingsForTikTok()` | Grabs attention immediately |
| YouTube tutorial | `VoiceSettingsForYouTube()` | Engaging but sustainable |
| Podcast episode | `VoiceSettingsForPodcast()` | Conversational, natural |
| Audiobook | `VoiceSettingsForAudiobook()` | Easy to listen to for hours |

## Tips

1. **Test with your voice** - Presets are starting points. Different voices respond differently to settings.

2. **Match your content** - A coding tutorial might use Udemy settings even on YouTube.

3. **Consider duration** - Longer content benefits from higher stability.

4. **Platform expectations** - TikTok audiences expect energy; podcast listeners expect calm.

5. **A/B test** - Try different presets and gather feedback.
