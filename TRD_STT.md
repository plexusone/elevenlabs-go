# TRD: WebSocket STT Update for scribe_v2_realtime

## Overview

The ElevenLabs WebSocket Speech-to-Text API has been updated. The current `websocketstt.go` implementation uses the deprecated `scribe_v1` message format which is no longer supported. This document outlines the changes needed to support the `scribe_v2_realtime` API.

## Current State

The current implementation in `websocketstt.go` uses:

- Message type `"config"` for initialization
- Message type `"audio"` with `audio` field for audio chunks
- Message type `"end_of_stream"` for stream termination
- Response parsing expects `text`, `is_final`, `confidence` fields

**Problem**: The server returns "You must be authenticated" errors after connection, and the model `scribe_v1` is no longer valid.

## Target State

Update to the `scribe_v2_realtime` API which uses:

- Query parameters for configuration (no init message needed)
- Message type `"input_audio_chunk"` for audio
- Different response message types: `partial_transcript`, `committed_transcript`

## API Reference

WebSocket URL: `wss://api.elevenlabs.io/v1/speech-to-text/realtime`

Documentation: https://elevenlabs.io/docs/api-reference/speech-to-text/v-1-speech-to-text-realtime

### Authentication

Two methods supported:

1. **Header**: `xi-api-key: <api_key>` (current implementation)
2. **Query Parameter**: `token=<single_use_token>` (for client-side use)

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `model_id` | string | - | Model to use (e.g., `scribe_v2_realtime`) |
| `audio_format` | enum | `pcm_16000` | `pcm_8000`, `pcm_16000`, `pcm_22050`, `pcm_24000`, `pcm_44100`, `pcm_48000`, `ulaw_8000` |
| `language_code` | string | - | ISO 639-1/639-3 code for language hint |
| `include_timestamps` | bool | false | Include word-level timestamps |
| `include_language_detection` | bool | false | Include detected language |
| `commit_strategy` | enum | `manual` | `manual` or `vad` |
| `vad_silence_threshold_secs` | float | 1.5 | Silence duration to trigger commit (VAD mode) |
| `vad_threshold` | float | 0.4 | VAD sensitivity threshold |
| `min_speech_duration_ms` | int | 100 | Minimum speech duration |
| `min_silence_duration_ms` | int | 100 | Minimum silence duration |

### Client-to-Server Messages

#### InputAudioChunk

```json
{
  "message_type": "input_audio_chunk",
  "audio_base_64": "<base64_encoded_audio>",
  "commit": false,
  "sample_rate": 16000,
  "previous_text": ""
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `message_type` | string | yes | Must be `"input_audio_chunk"` |
| `audio_base_64` | string | yes | Base64-encoded audio data |
| `commit` | bool | no | Force commit of current transcript |
| `sample_rate` | int | no | Sample rate if different from query param |
| `previous_text` | string | no | Context for better transcription |

### Server-to-Client Messages

#### SessionStarted

```json
{
  "message_type": "session_started",
  "session_id": "uuid",
  "config": {
    "sample_rate": 16000,
    "audio_format": "pcm_16000",
    "language_code": "en",
    "model_id": "scribe_v2_realtime",
    "vad_commit_strategy": false,
    "vad_silence_threshold_secs": 1.5,
    "vad_threshold": 0.4,
    "include_timestamps": true
  }
}
```

#### PartialTranscript

```json
{
  "message_type": "partial_transcript",
  "text": "partial transcription text"
}
```

#### CommittedTranscript

```json
{
  "message_type": "committed_transcript",
  "text": "final transcription text"
}
```

#### CommittedTranscriptWithTimestamps

```json
{
  "message_type": "committed_transcript_with_timestamps",
  "text": "final transcription text",
  "language_code": "en",
  "words": [
    {
      "text": "word",
      "start": 0.0,
      "end": 0.5,
      "type": "word",
      "speaker_id": "speaker_0",
      "logprob": -0.1,
      "characters": ["w", "o", "r", "d"]
    }
  ]
}
```

#### Error Messages

All error messages have:

```json
{
  "message_type": "<error_type>",
  "error": "error description"
}
```

Error types: `error`, `auth_error`, `quota_exceeded`, `commit_throttled`, `unaccepted_terms`, `rate_limited`, `queue_overflow`, `resource_exhausted`, `session_time_limit_exceeded`, `input_error`, `chunk_size_exceeded`, `insufficient_audio_activity`, `transcriber_error`

## Implementation Plan

### 1. Update WebSocketSTTOptions

```go
type WebSocketSTTOptions struct {
    // ModelID is the transcription model to use.
    // Default: "scribe_v2_realtime"
    ModelID string

    // AudioFormat specifies the audio encoding format.
    // Options: "pcm_8000", "pcm_16000", "pcm_22050", "pcm_24000", "pcm_44100", "pcm_48000", "ulaw_8000"
    // Default: "pcm_16000"
    AudioFormat string

    // LanguageCode is the expected language (e.g., "en", "es").
    // If not specified, language will be auto-detected.
    LanguageCode string

    // IncludeTimestamps enables word-level timing information.
    IncludeTimestamps bool

    // IncludeLanguageDetection includes detected language in responses.
    IncludeLanguageDetection bool

    // CommitStrategy determines how transcripts are committed.
    // Options: "manual" (default), "vad" (voice activity detection)
    CommitStrategy string

    // VAD settings (only used when CommitStrategy is "vad")
    VADSilenceThresholdSecs float64 // Default: 1.5
    VADThreshold            float64 // Default: 0.4
    MinSpeechDurationMs     int     // Default: 100
    MinSilenceDurationMs    int     // Default: 100
}
```

### 2. Update Message Structs

```go
// sttInputAudioChunk is the audio data message for v2 API.
type sttInputAudioChunk struct {
    MessageType  string `json:"message_type"` // "input_audio_chunk"
    AudioBase64  string `json:"audio_base_64"`
    Commit       bool   `json:"commit,omitempty"`
    SampleRate   int    `json:"sample_rate,omitempty"`
    PreviousText string `json:"previous_text,omitempty"`
}

// sttSessionStarted is the session started response.
type sttSessionStarted struct {
    MessageType string                 `json:"message_type"`
    SessionID   string                 `json:"session_id"`
    Config      map[string]interface{} `json:"config"`
}

// sttPartialTranscript is a partial transcription result.
type sttPartialTranscript struct {
    MessageType string `json:"message_type"`
    Text        string `json:"text"`
}

// sttCommittedTranscript is a final transcription result.
type sttCommittedTranscript struct {
    MessageType  string    `json:"message_type"`
    Text         string    `json:"text"`
    LanguageCode string    `json:"language_code,omitempty"`
    Words        []STTWord `json:"words,omitempty"`
}

// sttErrorResponse is an error from the server.
type sttErrorResponse struct {
    MessageType string `json:"message_type"`
    Error       string `json:"error"`
}
```

### 3. Update buildWebSocketURL

Add all query parameters:

```go
func (s *WebSocketSTTService) buildWebSocketURL(opts *WebSocketSTTOptions) (string, error) {
    // ... existing URL building ...

    q := u.Query()
    if opts.ModelID != "" {
        q.Set("model_id", opts.ModelID)
    }
    if opts.AudioFormat != "" {
        q.Set("audio_format", opts.AudioFormat)
    }
    if opts.LanguageCode != "" {
        q.Set("language_code", opts.LanguageCode)
    }
    if opts.IncludeTimestamps {
        q.Set("include_timestamps", "true")
    }
    if opts.IncludeLanguageDetection {
        q.Set("include_language_detection", "true")
    }
    if opts.CommitStrategy != "" {
        q.Set("commit_strategy", opts.CommitStrategy)
    }
    if opts.CommitStrategy == "vad" {
        if opts.VADSilenceThresholdSecs > 0 {
            q.Set("vad_silence_threshold_secs", fmt.Sprintf("%.2f", opts.VADSilenceThresholdSecs))
        }
        if opts.VADThreshold > 0 {
            q.Set("vad_threshold", fmt.Sprintf("%.2f", opts.VADThreshold))
        }
    }
    u.RawQuery = q.Encode()

    return u.String(), nil
}
```

### 4. Remove sendInit()

The v2 API uses query parameters for configuration, so no initialization message is needed. Remove the `sendInit()` call from `Connect()`.

### 5. Update SendAudio

```go
func (wsc *WebSocketSTTConnection) SendAudio(audio []byte) error {
    msg := sttInputAudioChunk{
        MessageType: "input_audio_chunk",
        AudioBase64: base64.StdEncoding.EncodeToString(audio),
    }
    return wsc.sendJSON(msg)
}

// SendAudioWithCommit sends audio and optionally commits the transcript.
func (wsc *WebSocketSTTConnection) SendAudioWithCommit(audio []byte, commit bool) error {
    msg := sttInputAudioChunk{
        MessageType: "input_audio_chunk",
        AudioBase64: base64.StdEncoding.EncodeToString(audio),
        Commit:      commit,
    }
    return wsc.sendJSON(msg)
}
```

### 6. Update readLoop

Handle the new message types:

```go
func (wsc *WebSocketSTTConnection) readLoop() {
    defer wsc.closeChannels()

    for {
        select {
        case <-wsc.closeChan:
            return
        default:
        }

        _, message, err := wsc.conn.ReadMessage()
        if err != nil {
            // ... error handling ...
            return
        }

        // Parse message type first
        var msgType struct {
            MessageType string `json:"message_type"`
        }
        if err := json.Unmarshal(message, &msgType); err != nil {
            continue
        }

        switch msgType.MessageType {
        case "session_started":
            // Session established, could store session_id if needed
            continue

        case "partial_transcript":
            var partial sttPartialTranscript
            if err := json.Unmarshal(message, &partial); err != nil {
                continue
            }
            wsc.sendTranscript(&STTTranscript{
                Text:    partial.Text,
                IsFinal: false,
            })

        case "committed_transcript", "committed_transcript_with_timestamps":
            var committed sttCommittedTranscript
            if err := json.Unmarshal(message, &committed); err != nil {
                continue
            }
            wsc.sendTranscript(&STTTranscript{
                Text:         committed.Text,
                IsFinal:      true,
                LanguageCode: committed.LanguageCode,
                Words:        committed.Words,
            })

        case "error", "auth_error", "quota_exceeded", "rate_limited",
             "input_error", "transcriber_error":
            var errResp sttErrorResponse
            if err := json.Unmarshal(message, &errResp); err != nil {
                continue
            }
            wsc.sendError(fmt.Errorf("server error (%s): %s", errResp.MessageType, errResp.Error))

        default:
            // Unknown message type, log and continue
        }
    }
}
```

### 7. Update STTTranscript struct

```go
type STTTranscript struct {
    Text         string    `json:"text"`
    IsFinal      bool      `json:"is_final"`
    LanguageCode string    `json:"language_code,omitempty"`
    Words        []STTWord `json:"words,omitempty"`
}

type STTWord struct {
    Text      string  `json:"text"`
    Start     float64 `json:"start"`
    End       float64 `json:"end"`
    Type      string  `json:"type,omitempty"`      // "word" or "spacing"
    SpeakerID string  `json:"speaker_id,omitempty"`
    Logprob   float64 `json:"logprob,omitempty"`
}
```

### 8. Update DefaultWebSocketSTTOptions

```go
func DefaultWebSocketSTTOptions() *WebSocketSTTOptions {
    return &WebSocketSTTOptions{
        ModelID:         "scribe_v2_realtime",
        AudioFormat:     "pcm_16000",
        CommitStrategy:  "manual",
        IncludeTimestamps: true,
    }
}
```

## Testing Plan

1. **Unit Tests**: Update existing tests in `websocketstt_test.go`
2. **Integration Tests**: Enable STT conformance tests in `omnivoice/stt/provider_conformance_test.go`
3. **Manual Testing**: Test with real audio using the example in `examples/websocket-stt/`

## Migration Notes

- The old `Encoding` field is replaced by `AudioFormat` query parameter
- The old `SampleRate` field is now derived from `AudioFormat`
- `EnablePartials` is now always enabled (partials are sent by default)
- `EnableWordTimestamps` is now `IncludeTimestamps` query parameter
- No initialization message is sent; configuration is via query parameters

## References

- [ElevenLabs Realtime STT API](https://elevenlabs.io/docs/api-reference/speech-to-text/v-1-speech-to-text-realtime)
- [Server-side Streaming Guide](https://elevenlabs.io/docs/developers/guides/cookbooks/speech-to-text/realtime/server-side-streaming)
- [Scribe v2 Realtime Announcement](https://elevenlabs.io/blog/introducing-scribe-v2-realtime)
