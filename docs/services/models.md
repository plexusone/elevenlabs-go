# Models

List and select text-to-speech models.

## List Available Models

```go
models, err := client.Models().ListTTSModels(ctx)
if err != nil {
    log.Fatal(err)
}

for _, m := range models {
    fmt.Printf("%s: %s\n", m.ModelID, m.Name)
    fmt.Printf("  Languages: %d\n", len(m.Languages))
    fmt.Printf("  Can TTS: %v\n", m.CanDoTTS)
}
```

## Model Object

| Field | Description |
|-------|-------------|
| `ModelID` | Unique identifier |
| `Name` | Display name |
| `Description` | Model description |
| `Languages` | Supported languages |
| `CanDoTTS` | Supports text-to-speech |
| `CanDoVoiceConversion` | Supports voice conversion |

## Available Models

| Model ID | Name | Best For |
|----------|------|----------|
| `eleven_multilingual_v2` | Multilingual v2 | Multiple languages, highest quality |
| `eleven_monolingual_v1` | English v1 | English only, fast |
| `eleven_turbo_v2` | Turbo v2 | Low latency |
| `eleven_turbo_v2_5` | Turbo v2.5 | Lowest latency |

## Choosing a Model

### For Quality

```go
// Best quality, supports 29 languages
modelID := "eleven_multilingual_v2"
```

### For Speed

```go
// Lowest latency for real-time applications
modelID := "eleven_turbo_v2_5"
```

### For English Only

```go
// Optimized for English
modelID := "eleven_monolingual_v1"
```

## Check Language Support

```go
models, _ := client.Models().ListTTSModels(ctx)

for _, m := range models {
    for _, lang := range m.Languages {
        if lang.LanguageID == "es" {  // Spanish
            fmt.Printf("%s supports Spanish\n", m.Name)
        }
    }
}
```

## Default Model

The SDK uses `eleven_multilingual_v2` as the default:

```go
elevenlabs.DefaultModelID  // "eleven_multilingual_v2"
```
