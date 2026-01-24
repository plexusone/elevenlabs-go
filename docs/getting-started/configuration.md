# Configuration

## Client Options

The client supports several configuration options using functional options pattern.

### API Key

```go
// From environment variable (default)
client, _ := elevenlabs.NewClient()

// Explicitly set
client, _ := elevenlabs.NewClient(
    elevenlabs.WithAPIKey("your-api-key"),
)
```

### Custom Base URL

```go
client, _ := elevenlabs.NewClient(
    elevenlabs.WithBaseURL("https://custom-endpoint.example.com"),
)
```

### Custom HTTP Client

```go
httpClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        10,
        IdleConnTimeout:     30 * time.Second,
    },
}

client, _ := elevenlabs.NewClient(
    elevenlabs.WithHTTPClient(httpClient),
)
```

### Request Timeout

```go
client, _ := elevenlabs.NewClient(
    elevenlabs.WithTimeout(5 * time.Minute),  // For long audio generation
)
```

### Multiple Options

```go
client, _ := elevenlabs.NewClient(
    elevenlabs.WithAPIKey("your-api-key"),
    elevenlabs.WithTimeout(3 * time.Minute),
    elevenlabs.WithBaseURL("https://api.elevenlabs.io"),
)
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ELEVENLABS_API_KEY` | API key for authentication |

## Accessing the Raw API Client

For advanced use cases not covered by the wrapper, access the underlying ogen-generated client:

```go
client, _ := elevenlabs.NewClient()

// Access raw API for advanced operations
rawClient := client.API()

// Use raw client methods directly
resp, err := rawClient.SomeAdvancedMethod(ctx, params)
```

## Constants

```go
// SDK version
elevenlabs.Version  // "0.1.0"

// Default base URL
elevenlabs.DefaultBaseURL  // "https://api.elevenlabs.io"

// Recommended model
elevenlabs.DefaultModelID  // "eleven_multilingual_v2"
```
