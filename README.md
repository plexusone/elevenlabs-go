# ElevenLabs Go SDK

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/elevenlabs-go/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/elevenlabs-go/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/elevenlabs-go/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/elevenlabs-go/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/elevenlabs-go/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/elevenlabs-go/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/elevenlabs-go
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/elevenlabs-go
 [docs-mkdoc-svg]: https://img.shields.io/badge/Go-dev%20guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/elevenlabs-go
 [viz-svg]: https://img.shields.io/badge/visualizaton-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fomnillm
 [loc-svg]: https://tokei.rs/b1/github/plexusone/elevenlabs-go
 [repo-url]: https://github.com/plexusone/elevenlabs-go
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/elevenlabs-go/blob/main/LICENSE

Go SDK for the [ElevenLabs API](https://elevenlabs.io/).

## Features

- 🗣️ **Text-to-Speech**: Convert text to realistic speech with multiple voices and models
- 📝 **Speech-to-Text**: Transcribe audio with speaker diarization support
- 🎙️ **Speech-to-Speech**: Voice conversion - transform speech to a different voice
- 🔊 **Sound Effects**: Generate sound effects from text descriptions
- 🎨 **Voice Design**: Create custom AI voices with specific characteristics
- 🎵 **Music Composition**: Generate music from text prompts
- 🎙️ **Audio Isolation**: Extract vocals/speech from audio
- ⏱️ **Forced Alignment**: Get word-level timestamps for audio
- 💬 **Text-to-Dialogue**: Generate multi-speaker conversations
- 🌍 **Dubbing**: Translate and dub video/audio content
- 📚 **Projects**: Manage long-form audio content (audiobooks, podcasts)
- 📖 **Pronunciation Dictionaries**: Control pronunciation of specific terms

### Conversational AI

- 🤖 **Agents**: Create and manage conversational AI agents with branching and deployment
- 🔀 **Branches**: Version control for agent configurations with traffic splitting
- 💬 **Conversations**: Access conversation history, transcripts, audio, and analysis
- 📚 **Knowledge Base**: RAG document management (files, text, URLs) for agent context
- 📞 **Batch Calling**: Schedule and manage bulk outbound calls
- 🧪 **Agent Testing**: Organize test folders and run response tests
- 📊 **Analytics**: Live conversation counts and agent insights

### Real-Time Services

- ⚡ **WebSocket TTS**: Low-latency text-to-speech streaming for real-time voice synthesis
- ⚡ **WebSocket STT**: Real-time speech-to-text with partial results
- 📞 **Twilio Integration**: Phone call integration for conversational AI agents
- 📱 **Phone Numbers**: Manage phone numbers for voice agents

### Command Line Interface

- 🖥️ **`elevenlabs tts`**: Generate speech from text files with YAML config support
- 📜 **`elevenlabs ttsscript`**: Batch TTS from JSON scripts with per-slide output
- 🎛️ **Presets**: Built-in configurations for oratory, podcast, audiobook styles

### OmniVoice Integration

- 🔌 **[OmniVoice](https://github.com/plexusone/omnivoice-core) Providers**: Use ElevenLabs as a drop-in backend for the vendor-agnostic OmniVoice interface
- 🔄 **Portable Code**: Swap voice providers (ElevenLabs, OpenAI, Google) without changing application logic
- 🧪 **TTS, STT, Agent**: Full provider implementations for text-to-speech, speech-to-text, and voice agents

### Agent Experience (AX)

- 🤖 **Machine-Readable Errors**: Error codes (`DOCUMENT_NOT_FOUND`, `NOT_LOGGED_IN`) for programmatic handling
- 🔄 **Automatic Retry**: TTS provider retries transient errors (429, 500) with exponential backoff
- 📊 **Error Classification**: 8 categories (auth, validation, rate_limit, etc.) for smart error handling
- ✅ **Pre-flight Validation**: Check required fields before making API calls
- 🔧 **Retry Policies**: Know which operations are safe to retry automatically

## Installation

```bash
go get github.com/plexusone/elevenlabs-go
```

### CLI Installation

```bash
go install github.com/plexusone/elevenlabs-go/cmd/elevenlabs@latest
```

## Quick Start

### Basic Text-to-Speech

```go
package main

import (
    "context"
    "io"
    "log"
    "os"

    elevenlabs "github.com/plexusone/elevenlabs-go"
)

func main() {
    // Create client (uses ELEVENLABS_API_KEY env var)
    client, err := elevenlabs.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // List available voices
    voices, err := client.Voice().List(ctx)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Found %d voices", len(voices))

    // Generate speech
    if len(voices) > 0 {
        audio, err := client.TTS().Simple(ctx,
            voices[0].VoiceID,
            "Hello from the ElevenLabs Go SDK!")
        if err != nil {
            log.Fatal(err)
        }

        // Save to file
        f, _ := os.Create("hello.mp3")
        defer f.Close()
        io.Copy(f, audio)
    }
}
```

### With Custom Options

```go
client, err := elevenlabs.NewClient(
    elevenlabs.WithAPIKey("your-api-key"),
    elevenlabs.WithTimeout(5 * time.Minute),
)
```

## Services

### Text-to-Speech

```go
import "github.com/plexusone/elevenlabs-go/tts"

// Simple generation
audio, err := client.TTS().Simple(ctx, voiceID, "Hello world")

// With full options
resp, err := client.TTS().Generate(ctx, &tts.Request{
    VoiceID: "21m00Tcm4TlvDq8ikWAM",
    Text:    "Hello with custom settings!",
    ModelID: "eleven_multilingual_v2",
    VoiceSettings: &elevenlabs.VoiceSettings{
        Stability:       0.6,
        SimilarityBoost: 0.8,
        Style:           0.1,
        SpeakerBoost:    true,
    },
    OutputFormat: "mp3_44100_192",
})
```

### Speech-to-Text

```go
// Transcribe from URL
result, err := client.STT().TranscribeURL(ctx, "https://example.com/audio.mp3")
fmt.Printf("Text: %s\n", result.Text)
fmt.Printf("Language: %s\n", result.LanguageCode)

// With speaker diarization
result, err := client.STT().TranscribeWithDiarization(ctx, audioURL)
for _, word := range result.Words {
    fmt.Printf("[%s] %s (%.2fs - %.2fs)\n", word.Speaker, word.Text, word.Start, word.End)
}
```

### Sound Effects

```go
import "github.com/plexusone/elevenlabs-go/audio"

// Simple sound effect
sfx, err := client.Audio().GenerateSoundEffect(ctx, &audio.SoundEffectRequest{
    Text:            "thunder and rain storm",
    DurationSeconds: 5,
})

// With options
sfx, err := client.Audio().GenerateSoundEffect(ctx, &audio.SoundEffectRequest{
    Text:            "spaceship engine humming",
    DurationSeconds: 10,
    PromptInfluence: 0.5,
})
```

### Music Composition

```go
import "github.com/plexusone/elevenlabs-go/content"

// Generate music from prompt
resp, err := client.Content().GenerateMusic(ctx, &content.MusicRequest{
    Prompt:     "upbeat electronic music for a tech video",
    DurationMs: 30000,
})

// Instrumental only
audio, err := client.Content().GenerateMusicInstrumental(ctx, "calm piano melody", 60000)

// Generate with composition plan for fine-grained control
plan, _ := client.Content().GenerateMusicPlan(ctx, &content.CompositionPlanRequest{
    Prompt:     "pop song about summer",
    DurationMs: 180000,
})
resp, err := client.Content().GenerateMusicDetailed(ctx, &content.MusicDetailedRequest{
    CompositionPlan: plan,
})

// Separate stems (vocals, drums, bass, etc.)
f, _ := os.Open("song.mp3")
stems, err := client.Content().SeparateStems(ctx, &content.StemSeparationRequest{
    File:     f,
    Filename: "song.mp3",
})
```

### Audio Isolation

```go
// Extract vocals from audio file
f, _ := os.Open("mixed_audio.mp3")
isolated, err := client.Audio().Isolate(ctx, f, "mixed_audio.mp3")
```

### Forced Alignment

```go
// Get word-level timestamps
f, _ := os.Open("speech.mp3")
result, err := client.Audio().Align(ctx, f, "speech.mp3",
    "The text that was spoken in the audio")

for _, word := range result.Words {
    fmt.Printf("%s: %.2fs - %.2fs\n", word.Text, word.Start, word.End)
}
```

### Text-to-Dialogue

```go
import "github.com/plexusone/elevenlabs-go/tts"

// Generate multi-speaker dialogue
audio, err := client.TTS().GenerateDialogue(ctx, []tts.DialogueInput{
    {Text: "Hello, how are you?", VoiceID: "voice1"},
    {Text: "I'm doing great, thanks!", VoiceID: "voice2"},
})
```

### Voice Design

```go
import "github.com/plexusone/elevenlabs-go/voice"

// Generate a custom voice
resp, err := client.Voice().Design(ctx, &voice.DesignRequest{
    Gender:         voice.GenderFemale,
    Age:            voice.AgeYoung,
    Accent:         voice.AccentAmerican,
    AccentStrength: 1.0,
    Text:           "This is a preview of the generated voice. It should be at least one hundred characters long for best results.",
})
```

### Pronunciation Dictionaries

```go
// Create from a map
dict, err := client.Account().CreatePronunciation(ctx, "Tech Terms", map[string]string{
    "API":     "A P I",
    "kubectl": "kube control",
    "nginx":   "engine X",
})

// Create from JSON file
dict, err := client.Account().CreatePronunciationFromJSON(ctx, "Terms", "pronunciation.json")
```

Use a dictionary in a request, and give the text around it so the delivery
stays continuous across requests:

```go
resp, err := client.TTS().Generate(ctx, &tts.Request{
    VoiceID:                   voiceID,
    Text:                      "The Miura was shown at Geneva.",
    PronunciationDictionaries: []tts.PronunciationDictionaryLocator{{ID: dict.ID}},
    PreviousText:              "With a body by Bertone, it became the Miura.",
})
```

Through omnivoice, pass `SynthesisConfig.Pronunciations` instead. Terms with
IPA are sent as SSML phoneme tags on models that honour them
(`eleven_flash_v2`, `eleven_turbo_v2`, `eleven_monolingual_v1`; see
`tts.SupportsPhonemes`), and as respellings on other models.
`DictionaryRules` in `omnivoice/tts` turns the same pronunciations into
dictionary rules.

### Dubbing

```go
import "github.com/plexusone/elevenlabs-go/content"

// Create dubbing job
dub, err := client.Content().CreateDubbing(ctx, &content.DubbingRequest{
    SourceURL:      "https://example.com/video.mp4",
    TargetLanguage: "es",
    Name:           "Video - Spanish",
})

// Check status
status, err := client.Content().GetDubbing(ctx, dub.DubbingID)
```

### Projects (Studio)

```go
import "github.com/plexusone/elevenlabs-go/content"

// Create a project for long-form content
project, err := client.Content().CreateProject(ctx, &content.CreateProjectRequest{
    Name:                    "My Audiobook",
    DefaultModelID:          "eleven_multilingual_v2",
    DefaultParagraphVoiceID: voiceID,
})

// Convert to audio
err = client.Content().ConvertProject(ctx, project.ProjectID)
```

### Speech-to-Speech (Voice Conversion)

```go
import "github.com/plexusone/elevenlabs-go/tts"

// Convert speech from one voice to another
f, _ := os.Open("input.mp3")
resp, err := client.TTS().ConvertSpeech(ctx, &tts.SpeechToSpeechRequest{
    VoiceID: targetVoiceID,
    Audio:   f,
})

// Simple conversion
output, err := client.TTS().ConvertSpeechSimple(ctx, targetVoiceID, audioReader)
```

### WebSocket TTS (Real-Time Streaming)

```go
import "github.com/plexusone/elevenlabs-go/realtime"

// Connect for low-latency TTS (ideal for LLM output)
conn, err := client.Realtime().ConnectTTS(ctx, voiceID, &realtime.TTSOptions{
    ModelID:                  "eleven_turbo_v2_5",
    OutputFormat:             "pcm_16000",
    OptimizeStreamingLatency: 3,
})
defer conn.Close()

// Stream text as it arrives (e.g., from LLM)
for text := range llmOutputStream {
    conn.SendText(text)
}
conn.Flush()

// Receive audio chunks
for audio := range conn.Audio() {
    // Play or save audio chunks
}
```

### WebSocket STT (Real-Time Transcription)

```go
import "github.com/plexusone/elevenlabs-go/realtime"

// Connect for live transcription
conn, err := client.Realtime().ConnectSTT(ctx, &realtime.STTOptions{
    SampleRate:     16000,
    EnablePartials: true,
})
defer conn.Close()

// Send audio chunks
go func() {
    for audioChunk := range microphoneInput {
        conn.SendAudio(audioChunk)
    }
    conn.EndStream()
}()

// Receive transcripts
for transcript := range conn.Transcripts() {
    if transcript.IsFinal {
        fmt.Println("Final:", transcript.Text)
    } else {
        fmt.Println("Partial:", transcript.Text)
    }
}
```

### Twilio Integration (Phone Calls)

```go
import "github.com/plexusone/elevenlabs-go/telephony"

// Register incoming Twilio call with an ElevenLabs agent
resp, err := client.Telephony().RegisterCall(ctx, &telephony.RegisterCallRequest{
    AgentID: "your-agent-id",
})
// Return resp.TwiML to Twilio webhook

// Make outbound call
call, err := client.Telephony().OutboundCall(ctx, &telephony.OutboundCallRequest{
    AgentID:            "your-agent-id",
    AgentPhoneNumberID: "phone-number-id",
    ToNumber:           "+1234567890",
})

// List phone numbers
numbers, err := client.Telephony().ListPhoneNumbers(ctx)
```

### Conversational AI Agents

```go
import "github.com/plexusone/elevenlabs-go/agents"

// Create an agent
agent, err := client.Agents().Create(ctx, &agents.CreateAgentRequest{
    Name: "Support Agent",
    Tags: []string{"support"},
    ConversationConfig: map[string]any{
        "agent": map[string]any{
            "prompt": map[string]any{
                "prompt": "You are a helpful support agent.",
            },
            "first_message": "Hello! How can I help?",
        },
    },
})

// List agents
resp, err := client.Agents().List(ctx, &agents.ListAgentsOptions{
    PageSize: 10,
    Search:   "support",
})

// Get agent branches
branches, err := client.Agents().ListBranches(ctx, agentID)
for _, b := range branches {
    fmt.Printf("Branch: %s (%.0f%% live)\n", b.Name, b.CurrentLivePercentage)
}

// Deploy with traffic splitting (A/B testing)
err = client.Agents().Deploy(ctx, agentID, []agents.DeploymentRequest{
    {BranchID: mainBranchID, Percentage: 80.0},
    {BranchID: testBranchID, Percentage: 20.0},
})

// Get conversation topics
topics, err := client.Agents().GetTopics(ctx, agentID)

// Delete agent
err = client.Agents().Delete(ctx, agentID)
```

### Conversations

```go
import "github.com/plexusone/elevenlabs-go/agents"

// List conversations with filters
resp, err := client.Agents().ListConversations(ctx, &agents.ListConversationsOptions{
    AgentID:        agentID,
    CallSuccessful: "success",
    PageSize:       10,
})

// Get conversation details with transcript
conv, err := client.Agents().GetConversation(ctx, conversationID)
for _, msg := range conv.Transcript {
    fmt.Printf("[%s] %s\n", msg.Role, msg.Message)
}

// Get conversation audio recording
audio, err := client.Agents().GetConversationAudio(ctx, conversationID)

// Search conversation transcripts
results, err := client.Agents().SearchConversations(ctx, "refund policy", nil)
```

### Knowledge Base

```go
import "github.com/plexusone/elevenlabs-go/agents"

// Upload a file document for RAG
f, _ := os.Open("product-manual.pdf")
doc, err := client.Agents().CreateFileDocument(ctx, &agents.CreateFileDocumentRequest{
    File:     f,
    Filename: "product-manual.pdf",
    Name:     "Product Manual",
})

// Create a text document
doc, err := client.Agents().CreateTextDocument(ctx, &agents.CreateTextDocumentRequest{
    Content: "Your company FAQ content here...",
    Name:    "Company FAQ",
})

// Index a URL
doc, err := client.Agents().CreateURLDocument(ctx, &agents.CreateURLDocumentRequest{
    URL:  "https://docs.example.com/api",
    Name: "API Documentation",
})

// List documents
resp, err := client.Agents().ListDocuments(ctx, &agents.ListDocumentsOptions{
    PageSize: 20,
})

// Get document chunks for RAG
chunks, err := client.Agents().GetDocumentChunks(ctx, documentID)
```

### Batch Calling

```go
import "github.com/plexusone/elevenlabs-go/agents"

// Create a batch call job
batch, err := client.Agents().CreateBatchCall(ctx, &agents.CreateBatchCallRequest{
    Name:               "Customer Outreach Campaign",
    AgentID:            agentID,
    AgentPhoneNumberID: phoneNumberID,
    Recipients: []agents.BatchCallRecipient{
        {PhoneNumber: "+14155551234"},
        {PhoneNumber: "+14155555678"},
    },
    Timezone: "America/New_York",
})

// List batch calls
batches, err := client.Agents().ListBatchCalls(ctx, nil)

// Cancel or retry a batch
err = client.Agents().CancelBatchCall(ctx, batchID)
_, err = client.Agents().RetryBatchCall(ctx, batchID)
```

### Agent Testing

```go
import "github.com/plexusone/elevenlabs-go/agents"

// Create a test folder
folder, err := client.Agents().CreateTestFolder(ctx, &agents.CreateTestFolderRequest{
    Name: "Regression Tests",
})

// List tests and folders
tests, err := client.Agents().ListTests(ctx, &agents.ListTestsOptions{
    PageSize: 20,
})

// Get test details
test, err := client.Agents().GetResponseTest(ctx, testID)

// Bulk move tests to a folder
err = client.Agents().BulkMoveTests(ctx, &agents.BulkMoveTestsRequest{
    TestIDs:        []string{"test1", "test2"},
    TargetFolderID: folderID,
})
```

### Analytics

```go
// Get number of active conversations
count, err := client.Agents().GetLiveCount(ctx)
fmt.Printf("Active conversations: %d\n", count)
```

## Examples

See the [`examples/`](https://github.com/plexusone/elevenlabs-go/tree/main/examples) directory for runnable examples:

| Example | Description |
|---------|-------------|
| `basic/` | Common SDK operations |
| `ax-error-handling/` | AX error codes for machine-readable error handling |
| `websocket-tts/` | Real-time TTS streaming for LLM integration |
| `websocket-stt/` | Live transcription with partial results |
| `speech-to-speech/` | Voice conversion |
| `twilio/` | Phone call integration with Twilio |
| `ttsscript/` | Multi-voice script authoring |
| `retryhttp/` | Retry-capable HTTP transport |

```bash
export ELEVENLABS_API_KEY="your-api-key"
go run examples/basic/main.go
```

## Command Line Interface

The `elevenlabs` CLI provides text-to-speech generation from the command line.

### Basic Usage

```bash
# Generate speech from a text file
elevenlabs tts -v <voice-id> speech.txt

# Use a preset (oratory, podcast, audiobook)
elevenlabs tts -v <voice-id> --preset oratory speech.txt

# High-quality PCM output
elevenlabs tts -v <voice-id> -f pcm_48000 -o output.wav speech.txt

# Estimate credits without calling API
elevenlabs tts -v <voice-id> --estimate speech.txt
```

### Configuration Files

Save and reuse TTS settings with YAML config files:

```bash
# Use config file
elevenlabs tts --config tts-config.yaml speech.txt

# Save current settings to config
elevenlabs tts -v <voice-id> --preset oratory --save-config my-config.yaml speech.txt
```

Example config file:

```yaml
voice_id: IT8nQhZJj9jzRwmC46Ko
model_id: eleven_v3
output_format: pcm_48000

voice_settings:
  stability: 0.4        # Lower = more expressive
  similarity_boost: 0.75
  style: 0.3            # Higher = more dramatic
  speed: 0.95           # Slightly slower for gravitas
```

### Presets

| Preset | Stability | Style | Speed | Format | Use Case |
|--------|-----------|-------|-------|--------|----------|
| `oratory` | 0.4 | 0.3 | 0.95 | pcm_48000 | Speeches, presentations |
| `podcast` | 0.5 | 0.0 | 1.0 | mp3_44100_128 | Conversational content |
| `audiobook` | 0.6 | 0.1 | 0.95 | pcm_48000 | Long-form narration |

### Input Format

Text files support ElevenLabs formatting:

```
[calm] <break time="1s"/>
There are moments in history when humanity TRANSFORMS.
<break time="0.5s"/>
[excited] This is AMAZING news!
```

- SSML `<break>` tags for pauses
- Emotion tags (`[calm]`, `[excited]`, `[firm]`) for v3 model
- CAPITALIZED words for emphasis

## Error Handling

### Basic Error Handling

```go
audio, err := client.TTS().Simple(ctx, voiceID, text)
if err != nil {
    if elevenlabs.IsRateLimitError(err) {
        log.Println("Rate limited, waiting...")
        time.Sleep(time.Minute)
    } else if elevenlabs.IsUnauthorizedError(err) {
        log.Fatal("Invalid API key")
    } else if elevenlabs.IsNotFoundError(err) {
        log.Fatal("Voice not found")
    } else {
        log.Fatalf("Error: %v", err)
    }
}
```

### AX Error Codes (Machine-Readable)

For AI agents and automated systems, use AX error codes for precise error handling:

```go
import "github.com/plexusone/elevenlabs-go/ax"

_, err := client.Voice().Get(ctx, voiceID)
if err != nil {
    // Extract AX error code
    if code, ok := elevenlabs.GetAXErrorCode(err); ok {
        switch code {
        case ax.ErrDocumentNotFound:
            // Handle not found - try alternative resource
        case ax.ErrNotLoggedIn, ax.ErrNeedsAuthorization:
            // Handle auth - re-authenticate
        case ax.ErrInvalidUID:
            // Handle validation - fix input
        }

        // Get error metadata
        if info := ax.GetErrorInfo(code); info != nil {
            log.Printf("Category: %s, Retryable: %v", info.Category, info.Retryable)
        }
    }
}
```

## Environment Variables

- `ELEVENLABS_API_KEY`: Your ElevenLabs API key (used automatically if not provided via `WithAPIKey`)

## Documentation

- [API Reference](https://plexusone.github.io/elevenlabs-go/)
- [ElevenLabs API Docs](https://elevenlabs.io/docs)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License
