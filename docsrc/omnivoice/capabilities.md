# OmniVoice Capabilities

This page documents OmniVoice interface capabilities and their implementation status in go-elevenlabs.

## Overview

[OmniVoice](https://github.com/agentplexus/omnivoice) provides vendor-agnostic interfaces for voice AI services. The go-elevenlabs SDK implements these interfaces, allowing ElevenLabs to be used as a drop-in provider.

## TTS (Text-to-Speech) Provider

### Interface: `tts.Provider`

| Method | Description | Implemented | Conformance Test |
|--------|-------------|:-----------:|:----------------:|
| `Name()` | Returns provider name | :white_check_mark: | :white_check_mark: |
| `Synthesize()` | Convert text to audio (batch) | :white_check_mark: | :white_check_mark: |
| `SynthesizeStream()` | Convert text to streaming audio | :white_check_mark: | :white_check_mark: |
| `ListVoices()` | List available voices | :white_check_mark: | :white_check_mark: |
| `GetVoice()` | Get voice by ID | :white_check_mark: | :white_check_mark: |

### Interface: `tts.StreamingProvider`

| Method | Description | Implemented | Conformance Test |
|--------|-------------|:-----------:|:----------------:|
| `SynthesizeFromReader()` | Stream text input to audio output | :white_check_mark: | :white_check_mark: |

### Synthesis Configuration

| Config Field | Description | Supported |
|--------------|-------------|:---------:|
| `VoiceID` | Voice identifier | :white_check_mark: |
| `Model` | TTS model (e.g., `eleven_turbo_v2_5`) | :white_check_mark: |
| `OutputFormat` | Audio format (`mp3`, `pcm`, `wav`, `opus`) | :white_check_mark: |
| `SampleRate` | Audio sample rate | :white_check_mark: |
| `Speed` | Speech speed multiplier | :white_check_mark: |
| `Pitch` | Voice pitch adjustment | :x: Not supported by ElevenLabs |
| `Stability` | Voice stability (0.0-1.0) | :white_check_mark: |
| `SimilarityBoost` | Voice similarity boost (0.0-1.0) | :white_check_mark: |

## STT (Speech-to-Text) Provider

### Interface: `stt.Provider`

| Method | Description | Implemented | Conformance Test |
|--------|-------------|:-----------:|:----------------:|
| `Name()` | Returns provider name | :white_check_mark: | :white_check_mark: |
| `Transcribe()` | Transcribe audio bytes | :white_check_mark: | :white_check_mark: |
| `TranscribeFile()` | Transcribe from file path | :white_check_mark: | - |
| `TranscribeURL()` | Transcribe from URL | :white_check_mark: | - |

### Interface: `stt.StreamingProvider`

| Method | Description | Implemented | Conformance Test |
|--------|-------------|:-----------:|:----------------:|
| `TranscribeStream()` | Real-time streaming transcription | :white_check_mark: | :white_check_mark: |

### Transcription Configuration

| Config Field | Description | Supported |
|--------------|-------------|:---------:|
| `Language` | BCP-47 language code | :white_check_mark: |
| `Model` | STT model (`scribe_v2_realtime`) | :white_check_mark: |
| `SampleRate` | Audio sample rate | :white_check_mark: (via `AudioFormat`) |
| `Channels` | Audio channels | :white_check_mark: (mono only) |
| `Encoding` | Audio encoding (`pcm`, `mulaw`) | :white_check_mark: (via `AudioFormat`) |
| `EnablePunctuation` | Add punctuation | :white_check_mark: (always enabled) |
| `EnableWordTimestamps` | Word-level timing | :white_check_mark: |
| `EnableSpeakerDiarization` | Speaker identification | :white_check_mark: (batch API only) |
| `MaxSpeakers` | Maximum speakers to detect | :white_check_mark: (batch API only) |
| `Keywords` | Recognition hints | :x: Not supported |
| `VocabularyID` | Custom vocabulary | :x: Not supported |

### WebSocket STT Audio Formats

The WebSocket STT API supports these audio formats:

| Format | Sample Rate | Use Case |
|--------|-------------|----------|
| `pcm_8000` | 8 kHz | Telephony |
| `pcm_16000` | 16 kHz | Standard (default) |
| `pcm_22050` | 22.05 kHz | Higher quality |
| `pcm_24000` | 24 kHz | High quality |
| `pcm_44100` | 44.1 kHz | CD quality |
| `pcm_48000` | 48 kHz | Professional |
| `ulaw_8000` | 8 kHz | Twilio/telephony |

## Agent Provider

### Interface: `agent.Provider`

| Method | Description | Implemented | Conformance Test |
|--------|-------------|:-----------:|:----------------:|
| `Name()` | Returns provider name | :white_check_mark: | - |
| `CreateSession()` | Create voice session | :white_check_mark: | - |
| `GetSession()` | Get session by ID | :white_check_mark: | - |
| `ListSessions()` | List active sessions | :white_check_mark: | - |

### Interface: `agent.Session`

| Method | Description | Implemented | Conformance Test |
|--------|-------------|:-----------:|:----------------:|
| `ID()` | Session identifier | :white_check_mark: | - |
| `Start()` | Begin voice session | :white_check_mark: | - |
| `Stop()` | End voice session | :white_check_mark: | - |
| `SendAudio()` | Send audio to agent | :white_check_mark: | - |
| `ReceiveAudio()` | Receive agent audio | :white_check_mark: | - |
| `SendText()` | Send text (bypass STT) | :white_check_mark: | - |
| `Events()` | Session event channel | :white_check_mark: | - |
| `Transcript()` | Conversation history | :white_check_mark: | - |
| `Metrics()` | Performance metrics | :white_check_mark: | - |

### Agent Configuration

| Config Field | Description | Supported |
|--------------|-------------|:---------:|
| `Name` | Agent name | :white_check_mark: |
| `SystemPrompt` | LLM system prompt | :x: (no LLM integration) |
| `VoiceID` | TTS voice | :white_check_mark: |
| `Language` | Primary language | :white_check_mark: |
| `STTProvider` | STT provider name | :x: (uses ElevenLabs) |
| `TTSProvider` | TTS provider name | :x: (uses ElevenLabs) |
| `LLMProvider` | LLM provider name | :x: (no LLM integration) |
| `InterruptionMode` | How to handle interruptions | :x: (not implemented) |
| `Tools` | Function calling | :x: (no LLM integration) |

!!! note "Agent Provider Limitations"
    The ElevenLabs agent provider combines WebSocket TTS and STT for bidirectional audio, but does **not** include LLM integration. You must handle:

    - Processing user transcripts
    - Generating agent responses
    - Calling `SpeakText()` to vocalize responses

    For full conversational AI, integrate with an LLM provider separately.

## Conformance Test Status

### Running Tests

```bash
# Run all conformance tests (requires API key)
export ELEVENLABS_API_KEY="your-api-key"
go test -v ./omnivoice/...
```

### Test Categories

| Category | Description | TTS | STT | Agent |
|----------|-------------|:---:|:---:|:-----:|
| **Interface** | Basic interface compliance | :white_check_mark: | :white_check_mark: | - |
| **Behavior** | Edge cases (empty input, cancellation) | :white_check_mark: | :white_check_mark: | - |
| **Integration** | Real API calls | :white_check_mark: | :white_check_mark: | - |

### Test Results Summary

| Provider | Interface | Behavior | Integration | Overall |
|----------|:---------:|:--------:|:-----------:|:-------:|
| TTS | :white_check_mark: Pass | :white_check_mark: Pass | :white_check_mark: Pass | **Pass** |
| STT | :white_check_mark: Pass | :white_check_mark: Pass | :white_check_mark: Pass | **Pass** |
| Agent | - | - | - | *Not tested* |

## Feature Comparison

### vs Direct SDK Usage

| Feature | OmniVoice Provider | Direct SDK |
|---------|:------------------:|:----------:|
| Vendor portability | :white_check_mark: | :x: |
| Consistent API | :white_check_mark: | :x: |
| Voice cloning | :x: | :white_check_mark: |
| Pronunciation dictionaries | :x: | :white_check_mark: |
| Projects (Studio) | :x: | :white_check_mark: |
| Audio isolation | :x: | :white_check_mark: |
| Sound effects | :x: | :white_check_mark: |
| Music generation | :x: | :white_check_mark: |
| Full API parameters | :x: | :white_check_mark: |

!!! tip "When to Use OmniVoice"
    Use OmniVoice providers when you need vendor portability or a consistent API across providers. Use the SDK directly when you need ElevenLabs-specific features.

## Version Compatibility

| go-elevenlabs | OmniVoice | Notes |
|---------------|-----------|-------|
| v0.7.0+ | v0.2.0+ | WebSocket STT uses scribe_v2_realtime |
| v0.5.0-v0.6.x | v0.1.0+ | WebSocket STT uses deprecated scribe_v1 |

## See Also

- [TTS Provider](tts.md) - Detailed TTS documentation
- [STT Provider](stt.md) - Detailed STT documentation
- [Agent Provider](agent.md) - Detailed Agent documentation
- [OmniVoice GitHub](https://github.com/agentplexus/omnivoice) - Main OmniVoice repository
