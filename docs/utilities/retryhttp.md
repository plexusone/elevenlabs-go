# Retry HTTP Transport

The `retryhttp` package provides an HTTP RoundTripper with exponential backoff retry logic, designed to work seamlessly with the ElevenLabs client and any other HTTP client.

This package is part of [mogo](https://github.com/grokify/mogo) and can be used with any HTTP client.

## Installation

```go
import "github.com/grokify/mogo/net/http/retryhttp"
```

## Quick Start

### Basic Usage

```go
import (
    elevenlabs "github.com/agentplexus/go-elevenlabs"
    "github.com/grokify/mogo/net/http/retryhttp"
)

// Create retry transport with defaults
rt := retryhttp.New()

// Use with ElevenLabs client
client, _ := elevenlabs.NewClient(
    elevenlabs.WithHTTPClient(rt.Client()),
)

// API calls now automatically retry on rate limits
audio, err := client.TextToSpeech().Simple(ctx, voiceID, "Hello world")
```

### With Custom Options

```go
rt := retryhttp.NewWithOptions(
    retryhttp.WithMaxRetries(5),
    retryhttp.WithInitialBackoff(500*time.Millisecond),
    retryhttp.WithMaxBackoff(30*time.Second),
    retryhttp.WithBackoffMultiplier(2.0),
    retryhttp.WithJitter(0.1),
)
```

## Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithMaxRetries(n)` | 3 | Maximum retry attempts |
| `WithInitialBackoff(d)` | 1s | Initial backoff duration |
| `WithMaxBackoff(d)` | 30s | Maximum backoff duration |
| `WithBackoffMultiplier(m)` | 2.0 | Backoff multiplier per retry |
| `WithJitter(j)` | 0.1 | Jitter factor (0.0-1.0) |
| `WithTransport(t)` | http.DefaultTransport | Underlying transport |
| `WithRetryableStatusCodes(codes)` | 429, 500, 502, 503, 504 | Status codes to retry |
| `WithShouldRetry(fn)` | - | Custom retry decision function |
| `WithOnRetry(fn)` | - | Callback before each retry |
| `WithLogger(l)` | nil (silent) | Injectable `*slog.Logger` for error logging |

## Features

### Exponential Backoff

Backoff increases exponentially with each retry:

```
Attempt 0: 1s
Attempt 1: 2s (1s × 2)
Attempt 2: 4s (2s × 2)
Attempt 3: 8s (4s × 2)
...capped at MaxBackoff
```

### Jitter

Random jitter prevents thundering herd when multiple clients retry simultaneously:

```go
rt := retryhttp.NewWithOptions(
    retryhttp.WithJitter(0.2), // ±20% randomness
)
```

### Retry-After Header

The transport automatically respects `Retry-After` headers from the server:

```go
// If server returns:
// HTTP/1.1 429 Too Many Requests
// Retry-After: 60
//
// The transport will wait 60 seconds before retrying
```

### Retry Callback

Monitor retry behavior with a callback:

```go
rt := retryhttp.NewWithOptions(
    retryhttp.WithOnRetry(func(attempt int, req *http.Request, resp *http.Response, err error, backoff time.Duration) {
        log.Printf("Retry %d: status=%d, waiting %v", attempt, resp.StatusCode, backoff)
    }),
)
```

### Custom Retry Logic

Override the default retry decision:

```go
rt := retryhttp.NewWithOptions(
    retryhttp.WithShouldRetry(func(resp *http.Response, err error) bool {
        if err != nil {
            return true // Retry connection errors
        }
        // Custom logic: retry on specific error codes
        return resp.StatusCode == 429 || resp.StatusCode >= 500
    }),
)
```

### Logging

Inject a `*slog.Logger` to capture internal errors (e.g., response body drain failures):

```go
import "log/slog"

rt := retryhttp.NewWithOptions(
    retryhttp.WithMaxRetries(3),
    retryhttp.WithLogger(slog.Default()),
)
```

If no logger is provided, internal errors are silently discarded using a null logger.

## Wrapping Existing Clients

Wrap an existing `*http.Client` with retry logic:

```go
existingClient := &http.Client{
    Timeout: 30 * time.Second,
}

retryClient := retryhttp.WrapClient(existingClient,
    retryhttp.WithMaxRetries(3),
)

// Preserves Timeout, CheckRedirect, and Jar from original
```

## Default Retryable Status Codes

By default, these status codes trigger a retry:

- **429** - Too Many Requests (rate limited)
- **500** - Internal Server Error
- **502** - Bad Gateway
- **503** - Service Unavailable
- **504** - Gateway Timeout

## Complete Example

```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    elevenlabs "github.com/agentplexus/go-elevenlabs"
    "github.com/agentplexus/go-elevenlabs/voices"
    "github.com/grokify/mogo/net/http/retryhttp"
)

func main() {
    // Create retry transport
    rt := retryhttp.NewWithOptions(
        retryhttp.WithMaxRetries(3),
        retryhttp.WithInitialBackoff(1*time.Second),
        retryhttp.WithOnRetry(func(attempt int, req *http.Request, resp *http.Response, err error, backoff time.Duration) {
            log.Printf("Retry %d after %v", attempt, backoff)
        }),
    )

    // Create client with retry support
    client, _ := elevenlabs.NewClient(
        elevenlabs.WithHTTPClient(rt.Client()),
    )

    ctx := context.Background()

    // API calls automatically retry on transient failures
    audio, err := client.TextToSpeech().Simple(ctx, voices.Rachel, "Hello!")
    if err != nil {
        log.Fatal(err)
    }

    // Use audio...
}
```

## Architecture

The package implements `http.RoundTripper`, making it composable with any HTTP client:

```
┌─────────────────┐
│ ElevenLabs      │
│ Client          │
└────────┬────────┘
         │
┌────────▼────────┐
│ http.Client     │
└────────┬────────┘
         │
┌────────▼────────┐
│ RetryTransport  │ ◄── Handles retries
└────────┬────────┘
         │
┌────────▼────────┐
│ http.Default    │
│ Transport       │
└─────────────────┘
```

## Compatibility

The package implements Go's standard `http.RoundTripper` interface, making it compatible with any HTTP client:

| SDK Generator | Usage |
|---------------|-------|
| **ogen** | `elevenlabs.WithHTTPClient(rt.Client())` |
| **OpenAPI-Generator** | `cfg.HTTPClient = rt.Client()` |
| **oapi-codegen** | `api.WithHTTPClient(rt.Client())` |
| **go-swagger** | Transport configuration |
| **Any `*http.Client`** | Set `Transport` field |

## Best Practices

1. **Set appropriate timeouts** - Use context timeouts to prevent infinite retry loops
2. **Monitor retries** - Use `WithOnRetry` to log retry attempts for debugging
3. **Tune backoff** - Adjust `InitialBackoff` and `MaxBackoff` based on API characteristics
4. **Use jitter** - Keep jitter enabled to prevent synchronized retries across clients
5. **Enable logging in production** - Use `WithLogger` to capture internal errors
