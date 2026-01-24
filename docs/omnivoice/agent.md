# Agent Provider

The Agent provider implements `agent.Provider` interface for building interactive voice agents using ElevenLabs real-time WebSocket TTS and STT services.

## Features

- Bidirectional real-time audio streaming
- Session management with unique IDs
- Event-driven architecture
- Conversation transcript tracking
- Performance metrics

## Installation

```go
import elevenagent "github.com/agentplexus/go-elevenlabs/omnivoice/agent"
```

## Creating a Provider

```go
// Using environment variable (ELEVENLABS_API_KEY)
provider, err := elevenagent.New()

// With explicit API key
provider, err := elevenagent.New(
    elevenagent.WithAPIKey("your-api-key"),
)

// With existing ElevenLabs client
client, _ := elevenlabs.NewClient()
provider := elevenagent.NewWithClient(client)
```

## Session Lifecycle

### Create and Start a Session

```go
// Create session with configuration
session, err := provider.CreateSession(ctx, agent.Config{
    VoiceID:  "21m00Tcm4TlvDq8ikWAM", // Rachel
    Language: "en",
})
if err != nil {
    log.Fatal(err)
}

// Start the session (connects WebSockets)
if err := session.Start(ctx); err != nil {
    log.Fatal(err)
}
defer session.Stop(ctx)
```

### Session Configuration

```go
config := agent.Config{
    // Voice ID for TTS output
    VoiceID: "21m00Tcm4TlvDq8ikWAM",

    // Language for STT
    Language: "en",

    // System prompt for the agent (used by your LLM)
    SystemPrompt: "You are a helpful assistant.",

    // Custom metadata
    Metadata: map[string]any{
        "user_id": "12345",
    },
}
```

## Audio I/O

### Sending Audio (User Input)

```go
// Send audio chunks (e.g., from microphone)
go func() {
    for {
        chunk := readFromMicrophone()
        if err := session.SendAudio(chunk); err != nil {
            log.Printf("Send error: %v", err)
            return
        }
    }
}()
```

### Receiving Audio (Agent Output)

```go
// Receive and play agent audio
go func() {
    for audio := range session.ReceiveAudio() {
        playToSpeaker(audio)
    }
}()
```

## Event Handling

The session emits events for all voice interactions:

```go
for event := range session.Events() {
    switch event.Type {
    case agent.EventSessionStarted:
        fmt.Println("Session started")

    case agent.EventUserTranscript:
        fmt.Printf("User said: %s\n", event.Data)

        // Process with your LLM and respond
        response := processWithLLM(event.Data.(string))
        session.SpeakText(response)

    case agent.EventAgentTranscript:
        fmt.Printf("Agent said: %s\n", event.Data)

    case agent.EventAgentSpeechStart:
        fmt.Println("Agent started speaking")

    case agent.EventSessionEnded:
        fmt.Println("Session ended")
        return

    case agent.EventError:
        log.Printf("Error: %v", event.Error)
    }
}
```

### Event Types

| Event | Description |
|-------|-------------|
| `EventSessionStarted` | Session has started successfully |
| `EventSessionEnded` | Session has ended |
| `EventUserTranscript` | User speech transcribed |
| `EventAgentTranscript` | Agent response text |
| `EventAgentSpeechStart` | Agent started speaking |
| `EventError` | An error occurred |

## Text-Based Interaction

You can also send text directly (bypassing STT):

```go
// Send text as user input
session.SendText("What's the weather like?")

// Make the agent speak
session.SpeakText("The weather is sunny and 72 degrees.")
```

## Transcript and Metrics

### Conversation Transcript

```go
transcript := session.Transcript()

for _, turn := range transcript {
    fmt.Printf("[%s] %s: %s\n",
        turn.Timestamp.Format(time.Kitchen),
        turn.Role,
        turn.Text,
    )
}
```

### Performance Metrics

```go
metrics := session.Metrics()

fmt.Printf("Session duration: %d ms\n", metrics.SessionDurationMs)
fmt.Printf("Turn count: %d\n", metrics.TurnCount)
```

## Session Management

### List Active Sessions

```go
sessionIDs, err := provider.ListSessions(ctx)
for _, id := range sessionIDs {
    fmt.Println(id)
}
```

### Get Existing Session

```go
session, err := provider.GetSession(ctx, "elevenlabs-1234567890-1")
if err != nil {
    log.Fatal("Session not found")
}
```

### Cleanup

```go
// Stop the session
session.Stop(ctx)

// Remove from provider's registry
provider.RemoveSession(session.ID())
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/agentplexus/omnivoice/agent"
    elevenagent "github.com/agentplexus/go-elevenlabs/omnivoice/agent"
)

func main() {
    ctx := context.Background()

    // Create provider
    provider, err := elevenagent.New()
    if err != nil {
        log.Fatal(err)
    }

    // Create session
    session, err := provider.CreateSession(ctx, agent.Config{
        VoiceID:  "21m00Tcm4TlvDq8ikWAM",
        Language: "en",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Start session
    if err := session.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer session.Stop(ctx)

    // Handle audio I/O in background
    go streamAudioIn(session)
    go streamAudioOut(session)

    // Process events
    for event := range session.Events() {
        switch event.Type {
        case agent.EventUserTranscript:
            userText := event.Data.(string)
            fmt.Printf("User: %s\n", userText)

            // Your LLM logic here
            response := "I heard you say: " + userText
            session.SpeakText(response)

        case agent.EventAgentTranscript:
            fmt.Printf("Agent: %s\n", event.Data)

        case agent.EventSessionEnded:
            return

        case agent.EventError:
            log.Printf("Error: %v", event.Error)
        }
    }
}
```

## Accessing the Underlying Client

For ElevenLabs-specific features:

```go
client := provider.Client()

// Access any ElevenLabs service directly
voices, _ := client.Voices().List(ctx)
```
