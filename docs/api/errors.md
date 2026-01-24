# Errors

Error types and handling for the ElevenLabs SDK.

## Error Types

### ValidationError

Returned when request validation fails before making an API call.

```go
type ValidationError struct {
    Field   string
    Message string
}
```

**Example:**

```go
audio, err := client.TextToSpeech().Simple(ctx, "", "Hello")
if err != nil {
    var valErr *elevenlabs.ValidationError
    if errors.As(err, &valErr) {
        fmt.Printf("Validation error on %s: %s\n", valErr.Field, valErr.Message)
        // Output: Validation error on voice_id: cannot be empty
    }
}
```

### APIError

Returned when the API returns an error response.

```go
type APIError struct {
    StatusCode int
    Message    string
    Detail     string
}
```

**Example:**

```go
audio, err := client.TextToSpeech().Simple(ctx, "invalid-voice", "Hello")
if err != nil {
    var apiErr *elevenlabs.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API error %d: %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

## Sentinel Errors

```go
var ErrEmptyVoiceID = errors.New("elevenlabs: voice_id cannot be empty")
var ErrEmptyText    = errors.New("elevenlabs: text cannot be empty")
```

## Error Helper Functions

### IsNotFoundError

```go
func IsNotFoundError(err error) bool
```

Returns true if the error is a 404 Not Found response.

```go
voice, err := client.Voices().Get(ctx, "nonexistent")
if elevenlabs.IsNotFoundError(err) {
    fmt.Println("Voice not found")
}
```

### IsUnauthorizedError

```go
func IsUnauthorizedError(err error) bool
```

Returns true if the error is a 401 Unauthorized response.

```go
if elevenlabs.IsUnauthorizedError(err) {
    fmt.Println("Invalid API key")
}
```

### IsRateLimitError

```go
func IsRateLimitError(err error) bool
```

Returns true if the error is a 429 Too Many Requests response.

```go
if elevenlabs.IsRateLimitError(err) {
    fmt.Println("Rate limited, waiting...")
    time.Sleep(time.Minute)
}
```

## Error Handling Patterns

### Complete Error Handling

```go
audio, err := client.TextToSpeech().Simple(ctx, voiceID, text)
if err != nil {
    // Check for validation errors first
    var valErr *elevenlabs.ValidationError
    if errors.As(err, &valErr) {
        return fmt.Errorf("invalid request: %s - %s", valErr.Field, valErr.Message)
    }

    // Check for specific API errors
    if elevenlabs.IsUnauthorizedError(err) {
        return errors.New("invalid API key")
    }
    if elevenlabs.IsRateLimitError(err) {
        return errors.New("rate limited, try again later")
    }
    if elevenlabs.IsNotFoundError(err) {
        return errors.New("voice not found")
    }

    // Generic API error
    var apiErr *elevenlabs.APIError
    if errors.As(err, &apiErr) {
        return fmt.Errorf("API error %d: %s", apiErr.StatusCode, apiErr.Message)
    }

    // Unknown error
    return fmt.Errorf("unexpected error: %w", err)
}
```

### Retry Pattern

```go
func generateWithRetry(client *elevenlabs.Client, voiceID, text string, maxRetries int) (io.Reader, error) {
    ctx := context.Background()

    for i := 0; i < maxRetries; i++ {
        audio, err := client.TextToSpeech().Simple(ctx, voiceID, text)
        if err == nil {
            return audio, nil
        }

        if elevenlabs.IsRateLimitError(err) {
            backoff := time.Duration(i+1) * 30 * time.Second
            log.Printf("Rate limited, waiting %v...", backoff)
            time.Sleep(backoff)
            continue
        }

        // Non-retryable error
        return nil, err
    }

    return nil, errors.New("max retries exceeded")
}
```

### Pre-flight Validation

```go
func generateSafely(client *elevenlabs.Client, voiceID, text string) (io.Reader, error) {
    // Validate inputs before API call
    if voiceID == "" {
        return nil, &elevenlabs.ValidationError{
            Field:   "voice_id",
            Message: "cannot be empty",
        }
    }
    if text == "" {
        return nil, &elevenlabs.ValidationError{
            Field:   "text",
            Message: "cannot be empty",
        }
    }

    // Check character limit
    sub, err := client.User().GetSubscription(context.Background())
    if err != nil {
        return nil, err
    }
    if sub.CharactersRemaining() < len(text) {
        return nil, errors.New("insufficient characters remaining")
    }

    // Safe to proceed
    return client.TextToSpeech().Simple(context.Background(), voiceID, text)
}
```
