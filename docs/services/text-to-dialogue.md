# Text-to-Dialogue

Generate multi-speaker conversations with different voices for each speaker.

## Basic Usage

```go
audio, err := client.TextToDialogue().Simple(ctx, []elevenlabs.DialogueInput{
    {Text: "Hello, how are you today?", VoiceID: "voice1"},
    {Text: "I'm doing great, thanks for asking!", VoiceID: "voice2"},
    {Text: "That's wonderful to hear.", VoiceID: "voice1"},
})
if err != nil {
    log.Fatal(err)
}

f, _ := os.Create("dialogue.mp3")
io.Copy(f, audio)
```

## Full Options

```go
audio, err := client.TextToDialogue().Generate(ctx, &elevenlabs.DialogueRequest{
    Inputs: []elevenlabs.DialogueInput{
        {Text: "Welcome to the show!", VoiceID: hostVoiceID},
        {Text: "Thanks for having me.", VoiceID: guestVoiceID},
    },
    ModelID:      "eleven_multilingual_v2",
    LanguageCode: "en",
    Seed:         42,  // For reproducible output
})
```

## With Timestamps

Get timing information for each segment:

```go
resp, err := client.TextToDialogue().GenerateWithTimestamps(ctx, &elevenlabs.DialogueRequest{
    Inputs: []elevenlabs.DialogueInput{
        {Text: "First speaker talks.", VoiceID: voice1},
        {Text: "Second speaker responds.", VoiceID: voice2},
    },
})

fmt.Printf("Audio (base64): %s\n", resp.AudioBase64)

for _, seg := range resp.VoiceSegments {
    fmt.Printf("Voice %s: %.2fs - %.2fs\n", seg.VoiceID, seg.StartTime, seg.EndTime)
}
```

## Streaming

For real-time playback:

```go
stream, err := client.TextToDialogue().GenerateStream(ctx, &elevenlabs.DialogueRequest{
    Inputs: dialogueInputs,
})
```

## Response Structures

```go
type DialogueInput struct {
    Text    string // Text to speak
    VoiceID string // Voice ID for this line
}

type DialogueRequest struct {
    Inputs       []DialogueInput
    ModelID      string // TTS model
    LanguageCode string // ISO 639-1 code
    Seed         int    // For reproducibility
}

type DialogueResponse struct {
    AudioBase64   string         // Base64-encoded audio
    VoiceSegments []VoiceSegment // Timing info
}

type VoiceSegment struct {
    VoiceID   string
    StartTime float64
    EndTime   float64
}
```

## Use Cases

### Podcast Conversations

```go
hostVoice := "21m00Tcm4TlvDq8ikWAM"
guestVoice := "AZnzlk1XvdvUeBnXmlld"

dialogue := []elevenlabs.DialogueInput{
    {Text: "Welcome back to Tech Talk! Today we're discussing AI.", VoiceID: hostVoice},
    {Text: "Thanks for having me. AI has really transformed everything.", VoiceID: guestVoice},
    {Text: "Let's dive into the specifics. What excites you most?", VoiceID: hostVoice},
    {Text: "Definitely the creative applications - music, art, writing.", VoiceID: guestVoice},
}

audio, _ := client.TextToDialogue().Simple(ctx, dialogue)
```

### Educational Content

```go
teacher := "teacher-voice-id"
student := "student-voice-id"

lesson := []elevenlabs.DialogueInput{
    {Text: "Today we'll learn about photosynthesis.", VoiceID: teacher},
    {Text: "What exactly is photosynthesis?", VoiceID: student},
    {Text: "It's how plants convert sunlight into energy.", VoiceID: teacher},
    {Text: "So plants are like solar panels?", VoiceID: student},
    {Text: "That's a great analogy! Let me explain further.", VoiceID: teacher},
}

audio, _ := client.TextToDialogue().Simple(ctx, lesson)
```

### Audiobook Dialogues

```go
narrator := "narrator-voice-id"
character1 := "character1-voice-id"
character2 := "character2-voice-id"

story := []elevenlabs.DialogueInput{
    {Text: "The detective entered the room slowly.", VoiceID: narrator},
    {Text: "I've been expecting you.", VoiceID: character1},
    {Text: "Then you know why I'm here.", VoiceID: character2},
    {Text: "The tension was palpable.", VoiceID: narrator},
}

audio, _ := client.TextToDialogue().Simple(ctx, story)
```

### Customer Service Demos

```go
agent := "agent-voice-id"
customer := "customer-voice-id"

demo := []elevenlabs.DialogueInput{
    {Text: "Thank you for calling support. How can I help?", VoiceID: agent},
    {Text: "I'm having trouble with my account.", VoiceID: customer},
    {Text: "I'd be happy to help. Can I have your account number?", VoiceID: agent},
    {Text: "Sure, it's 12345.", VoiceID: customer},
    {Text: "Perfect, I can see your account now.", VoiceID: agent},
}

audio, _ := client.TextToDialogue().Simple(ctx, demo)
```

### Interview Simulation

```go
// Get timestamps for video sync
resp, _ := client.TextToDialogue().GenerateWithTimestamps(ctx, &elevenlabs.DialogueRequest{
    Inputs: interviewDialogue,
})

// Decode audio
audioData, _ := base64.StdEncoding.DecodeString(resp.AudioBase64)

// Use segments for visual indicators
for _, seg := range resp.VoiceSegments {
    fmt.Printf("Show speaker %s avatar from %.2fs to %.2fs\n",
        seg.VoiceID, seg.StartTime, seg.EndTime)
}
```

## Voice Selection Tips

1. **Use contrasting voices** - Different genders, accents, or tones help distinguish speakers
2. **Match voice to character** - Select voices that fit the persona
3. **Test combinations** - Some voice pairs work better together

```go
// Get available voices
voices, _ := client.Voices().List(ctx)

// Filter by characteristics
var maleVoices, femaleVoices []elevenlabs.Voice
for _, v := range voices {
    if v.Labels["gender"] == "male" {
        maleVoices = append(maleVoices, v)
    } else {
        femaleVoices = append(femaleVoices, v)
    }
}
```

## Best Practices

1. **Keep turns natural** - Avoid very long monologues
2. **Use appropriate voices** - Match voice characteristics to roles
3. **Add pauses naturally** - Include "..." or commas for natural pauses
4. **Test with timestamps** - Verify timing for video sync use cases
5. **Use consistent voice IDs** - Don't mix up which voice is which speaker
