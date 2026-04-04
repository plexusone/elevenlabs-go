# AX Package

The `ax` package provides Agent Experience (AX) metadata for the ElevenLabs API. It enables AI agents to handle errors, make retry decisions, and validate requests programmatically.

## Installation

The ax package is included in elevenlabs-go:

```go
import "github.com/plexusone/elevenlabs-go/ax"
```

## Error Handling

### Error Code Constants

```go
const (
    ErrDocumentNotFound    = "DOCUMENT_NOT_FOUND"
    ErrInvalidUID          = "INVALID_UID"
    ErrMissingFeedback     = "MISSING_FEEDBACK"
    ErrNeedsAuthorization  = "NEEDS_AUTHORIZATION"
    ErrNotLoggedIn         = "NOT_LOGGED_IN"
    ErrNoEditChanges       = "NO_EDIT_CHANGES"
    ErrUnprocessableEntity = "UNPROCESSABLE_ENTITY"
    ErrUserNotFound        = "USER_NOT_FOUND"
    ErrWorkspaceNotFound   = "WORKSPACE_NOT_FOUND"
)
```

### Checking Errors

```go
// Using the main package helpers
if elevenlabs.IsAXError(err, ax.ErrDocumentNotFound) {
    // Handle document not found
}

// Extract error code
if code, ok := elevenlabs.GetAXErrorCode(err); ok {
    switch code {
    case ax.ErrDocumentNotFound:
        // ...
    case ax.ErrNeedsAuthorization:
        // ...
    }
}
```

### Error Metadata

```go
info := ax.GetErrorInfo(ax.ErrDocumentNotFound)
// info.Code        = "DOCUMENT_NOT_FOUND"
// info.Category    = "not_found"
// info.Retryable   = false
// info.Description = "The requested document was not found"
```

### Error Categories

```go
// Check error category
ax.IsAuthError(code)       // auth: NOT_LOGGED_IN, NEEDS_AUTHORIZATION
ax.IsNotFoundError(code)   // not_found: DOCUMENT_NOT_FOUND, USER_NOT_FOUND, WORKSPACE_NOT_FOUND
ax.IsValidationError(code) // validation: INVALID_UID, UNPROCESSABLE_ENTITY, etc.
```

## Retry Policies

### Checking Retryability

```go
if ax.IsRetryable("get_voices") {
    // Safe to retry - GET operation
}

if !ax.IsRetryable("create_voice") {
    // Not safe to retry - would create duplicates
}
```

### Retry Policy Map

The `RetryPolicy` map contains 236 operation IDs:

```go
// Example entries
var RetryPolicy = map[string]bool{
    "get_voices":        true,  // Safe to retry
    "get_models":        true,  // Safe to retry
    "create_voice":      false, // Not safe
    "delete_voice":      false, // Not safe
    "text_to_speech_full": false, // Not safe (consumes credits)
}
```

### Getting All Retryable Operations

```go
ops := ax.GetRetryableOperations()
// Returns all operation IDs where retry is safe
```

## Required Fields Validation

### Pre-flight Validation

```go
// Check which fields are required
fields := ax.GetRequiredFields("text_to_speech_full")
// fields = []string{"text"}

// Validate before API call
present := map[string]bool{"text": true}
if msg := ax.ValidateFields("text_to_speech_full", present); msg != "" {
    return fmt.Errorf("validation failed: %s", msg)
}
```

### Required Fields Map

```go
var RequiredFields = map[string][]string{
    "text_to_speech_full":      {"text"},
    "create_voice":             {"voice_name", "voice_description", "generated_voice_id"},
    "create_batch_call":        {"call_name", "agent_id", "recipients"},
    "sound_generation":         {"text"},
    // ... 72 operations total
}
```

## Capabilities

### Checking Capabilities

```go
caps := ax.GetCapabilities("get_voices")
// caps = []Capability{CapRead}

if ax.HasCapability("delete_voice", ax.CapDelete) {
    // Operation can delete resources
}

if ax.IsReadOnly("get_voices") {
    // Operation only reads, doesn't modify
}

if ax.RequiresAdmin("invite_user") {
    // Operation requires admin permissions
}
```

### Capability Constants

```go
const (
    CapRead   Capability = "read"   // Retrieves data
    CapWrite  Capability = "write"  // Creates or modifies data
    CapDelete Capability = "delete" // Removes data
    CapAdmin  Capability = "admin"  // Requires elevated permissions
)
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    elevenlabs "github.com/plexusone/elevenlabs-go"
    "github.com/plexusone/elevenlabs-go/ax"
)

func main() {
    client, err := elevenlabs.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    voiceID := "some-voice-id"

    // Attempt to get a voice
    voice, err := client.Voices().Get(ctx, voiceID)
    if err != nil {
        handleError(err)
        return
    }

    fmt.Printf("Voice: %s\n", voice.Name)
}

func handleError(err error) {
    // Extract AX error code
    code, ok := elevenlabs.GetAXErrorCode(err)
    if !ok {
        // Not an AX-recognized error
        log.Printf("Unknown error: %v", err)
        return
    }

    // Get error metadata
    info := ax.GetErrorInfo(code)
    log.Printf("Error: %s (category=%s, retryable=%v)",
        code, info.Category, info.Retryable)

    // Handle by category
    switch info.Category {
    case "not_found":
        log.Println("Resource not found - try alternative")
    case "auth":
        log.Println("Authentication issue - re-authenticate")
    case "validation":
        log.Println("Validation failed - fix request")
    }
}
```

## API Reference

### Functions

| Function | Description |
|----------|-------------|
| `IsErrorCode(err, code)` | Check if error contains code |
| `ContainsErrorCode(err)` | Extract error code from error |
| `GetErrorInfo(code)` | Get metadata for error code |
| `IsAuthError(code)` | Check if code is auth category |
| `IsNotFoundError(code)` | Check if code is not_found category |
| `IsValidationError(code)` | Check if code is validation category |
| `IsRetryable(opID)` | Check if operation is safe to retry |
| `GetRetryableOperations()` | Get all retryable operation IDs |
| `GetRequiredFields(opID)` | Get required fields for operation |
| `HasRequiredFields(opID)` | Check if operation has required fields |
| `MissingFields(opID, present)` | Get list of missing required fields |
| `ValidateFields(opID, present)` | Get validation error message |
| `GetCapabilities(opID)` | Get capabilities for operation |
| `HasCapability(opID, cap)` | Check if operation has capability |
| `IsReadOnly(opID)` | Check if operation only reads |
| `RequiresAdmin(opID)` | Check if operation needs admin |

### Types

| Type | Description |
|------|-------------|
| `Capability` | Operation capability (read, write, delete, admin) |
| `ErrorCodeInfo` | Metadata about an error code |

## See Also

- [Errors Documentation](errors.md) - Main error handling
- [AX Integration Case Study](../case-studies/ax-integration.md) - Full case study
- [DIRECT Principles](https://github.com/grokify/direct-principles) - Design principles
- [AX Spec](https://github.com/grokify/ax-spec) - Code generation tooling
