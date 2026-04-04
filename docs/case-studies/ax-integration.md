# Agent Experience (AX) Integration Case Study

## Executive Summary

This case study documents the integration of Agent Experience (AX) principles into elevenlabs-go, demonstrating how machine-readable metadata enables AI agents to handle errors, make retry decisions, and validate requests autonomously.

**Key Results:**

- 9 error codes discovered via API probing
- 236 operations mapped with retry policies
- 72 operations have required field definitions
- Error handling upgraded from string parsing to typed constants

## Background

### The Challenge

AI agents interact with APIs differently than human developers:

| Aspect | Human Developer | AI Agent |
|--------|-----------------|----------|
| Error handling | Read message, debug | Parse structure, recover |
| Retry logic | Intuitive judgment | Needs explicit rules |
| Validation | IDE warnings, docs | Pre-flight checks |
| Learning | Read examples | Execute iteratively |

Traditional SDKs optimize for human ergonomics. Agent-friendly SDKs need machine-readable metadata.

### The Project

elevenlabs-go is a Go SDK for the ElevenLabs API:

- **204 endpoints** covering text-to-speech, voice cloning, dubbing, and more
- **54,000 line OpenAPI specification**
- **330,000 lines of generated Go code** (via ogen)
- **Wrapper layer** providing ergonomic APIs over generated code

The SDK is used by AI agents for voice generation tasks.

## The Problem

When an agent encounters an API error:

```json
{
  "detail": {
    "status": "document_not_found",
    "message": "The requested document was not found"
  }
}
```

The agent faces several challenges:

1. **How to identify the error?** — Parse strings? Match substrings?
2. **Should it retry?** — Safe for reads, dangerous for writes
3. **What alternatives exist?** — No structured guidance
4. **Is the request valid?** — Only learns after calling API

### Before AX Integration

```go
// Fragile string matching
if strings.Contains(err.Error(), "not found") {
    // Handle... but which kind of "not found"?
}

// HTTP status codes lack specificity
if apiErr.StatusCode == 404 {
    // Is it a document? User? Workspace?
}

// No retry guidance
// Agent must guess or hardcode rules
```

## Solution: AX Integration

### DIRECT Principles

The integration follows the DIRECT principles for Agent Experience:

| Principle | Implementation |
|-----------|----------------|
| **Deterministic** | Typed error constants with predictable values |
| **Introspectable** | Error metadata (category, retryability) |
| **Recoverable** | Category-based recovery strategies |
| **Explicit** | Required fields documented per operation |
| **Consistent** | Uniform error handling patterns |
| **Testable** | Comprehensive unit tests |

### Implementation Steps

#### Step 1: API Discovery

The OpenAPI specification doesn't document all error codes. We used ax-spec's discovery feature to probe the actual API:

```bash
ax-spec enrich elevenlabs-openapi.json \
  --discover \
  --api-key $ELEVENLABS_API_KEY \
  --output elevenlabs-openapi-ax.json
```

**Discovered Error Codes:**

| Code | Category | Description | Retryable |
|------|----------|-------------|-----------|
| `DOCUMENT_NOT_FOUND` | not_found | Resource doesn't exist | No |
| `USER_NOT_FOUND` | not_found | User doesn't exist | No |
| `WORKSPACE_NOT_FOUND` | not_found | Workspace doesn't exist | No |
| `NOT_LOGGED_IN` | auth | User not authenticated | No |
| `NEEDS_AUTHORIZATION` | auth | Additional permissions required | No |
| `INVALID_UID` | validation | Invalid identifier format | No |
| `UNPROCESSABLE_ENTITY` | validation | Request validation failed | No |
| `MISSING_FEEDBACK` | validation | Required feedback not provided | No |
| `NO_EDIT_CHANGES` | validation | Edit request had no changes | No |

#### Step 2: Code Generation

Generated Go code from the enriched specification:

```bash
ax-spec gen elevenlabs-openapi-ax.json \
  --output ax/ \
  --package ax
```

**Generated Files:**

| File | Lines | Purpose |
|------|-------|---------|
| `ax/doc.go` | 25 | Package documentation |
| `ax/errors.go` | 140 | Error constants and metadata |
| `ax/retry.go` | 260 | Retry policies (236 operations) |
| `ax/validation.go` | 120 | Required fields (72 operations) |
| `ax/capabilities.go` | 100 | Operation capabilities |
| `ax/ax_test.go` | 180 | Unit tests |

#### Step 3: SDK Integration

Enhanced the existing `errors.go` with AX-aware methods:

```go
// Method on APIError to extract AX code
func (e *APIError) AXErrorCode() (string, bool) {
    for _, code := range ax.AllErrorCodes {
        if strings.Contains(e.Message, code) ||
           strings.Contains(e.Detail, code) {
            return code, true
        }
    }
    return "", false
}

// Top-level helpers
func IsAXError(err error, code string) bool
func GetAXErrorCode(err error) (string, bool)
```

## Results

### Error Handling Improvement

**Before:**

```go
resp, err := client.Voices().Get(ctx, voiceID)
if err != nil {
    if strings.Contains(err.Error(), "not found") {
        // Fragile, might miss variations
        return nil, ErrVoiceNotFound
    }
    return nil, err
}
```

**After:**

```go
resp, err := client.Voices().Get(ctx, voiceID)
if err != nil {
    if code, ok := elevenlabs.GetAXErrorCode(err); ok {
        switch code {
        case ax.ErrDocumentNotFound:
            // Specific, reliable match
            return nil, ErrVoiceNotFound
        case ax.ErrNeedsAuthorization:
            // Agent can request permissions
            return nil, ErrNeedsPermission
        case ax.ErrInvalidUID:
            // Agent can validate input
            return nil, ErrInvalidVoiceID
        }
    }
    return nil, err
}
```

### Retry Policy Coverage

236 operations now have documented retry safety:

```go
// Safe to retry with exponential backoff
if ax.IsRetryable("get_voices") {
    return retry.Do(ctx, func() error {
        _, err := client.Voices().List(ctx)
        return err
    }, retry.WithBackoff(time.Second, 30*time.Second))
}

// Do not retry - would create duplicates
if !ax.IsRetryable("create_voice") {
    return err // Return immediately
}
```

**Retry Policy Distribution:**

| Category | Count | Policy |
|----------|-------|--------|
| GET operations | 98 | Retryable |
| List operations | 22 | Retryable |
| POST (create) | 48 | Not retryable |
| PUT (update) | 26 | Not retryable |
| DELETE | 26 | Not retryable |
| Special (POST reads) | 16 | Varies |

### Pre-flight Validation

72 operations have required field definitions:

```go
// Before calling API, validate required fields
fields := map[string]bool{
    "text": true,
    // Missing: voice_id (if required)
}

if msg := ax.ValidateFields("text_to_speech_full", fields); msg != "" {
    // msg = "missing required fields: ..."
    return fmt.Errorf("validation failed: %s", msg)
}

// All required fields present - safe to call
resp, err := client.TextToSpeech().Generate(ctx, req)
```

### Error Category Handling

Error metadata enables category-based recovery:

```go
code, ok := elevenlabs.GetAXErrorCode(err)
if !ok {
    return err // Unknown error
}

info := ax.GetErrorInfo(code)

switch info.Category {
case "not_found":
    // Try alternative resource
    return tryAlternative(ctx)

case "auth":
    // Re-authenticate or request permissions
    return requestPermissions(ctx)

case "validation":
    // Fix request and retry
    fixedReq := fixRequest(req, code)
    return retryWith(ctx, fixedReq)
}
```

## Metrics

### Code Changes

| Component | Files Changed | Lines Added |
|-----------|---------------|-------------|
| ax package | 6 new files | ~825 lines |
| errors.go | 1 modified | ~50 lines |
| errors_test.go | 1 modified | ~100 lines |
| Example | 1 new file | ~100 lines |
| **Total** | **9 files** | **~1,075 lines** |

### Test Coverage

All new code is tested:

```bash
$ go test -v ./ax/...
=== RUN   TestIsErrorCode
--- PASS: TestIsErrorCode (0.00s)
=== RUN   TestContainsErrorCode
--- PASS: TestContainsErrorCode (0.00s)
=== RUN   TestGetErrorInfo
--- PASS: TestGetErrorInfo (0.00s)
=== RUN   TestErrorCategoryHelpers
--- PASS: TestErrorCategoryHelpers (0.00s)
=== RUN   TestIsRetryable
--- PASS: TestIsRetryable (0.00s)
=== RUN   TestGetRequiredFields
--- PASS: TestGetRequiredFields (0.00s)
...
PASS
ok      github.com/plexusone/elevenlabs-go/ax
```

### API Coverage

| Metadata Type | Operations Covered | Percentage |
|---------------|-------------------|------------|
| Retry policy | 236 / 204 | 100%+ (includes internal) |
| Required fields | 72 / 204 | 35% (POST/PUT/PATCH only) |
| Error codes | 9 unique | N/A (discovered) |
| Capabilities | 50+ | ~25% (key operations) |

## Key Learnings

### 1. API Discovery is Essential

The OpenAPI specification did not document all error codes. Real API calls were necessary to discover actual error responses.

**Recommendation:** Always probe production APIs to discover undocumented behavior.

### 2. Code Generation Scales

With 204 endpoints, manual metadata would be error-prone. Code generation from the enriched spec ensures consistency.

**Recommendation:** Invest in tooling that generates code from specifications.

### 3. Categories Enable Strategies

Individual error codes are useful, but categories enable broader strategies:

- **not_found** → Try alternatives, report to user
- **auth** → Re-authenticate, escalate permissions
- **validation** → Fix input, retry

**Recommendation:** Define error categories, not just error codes.

### 4. Layers Work Together

The AX integration enhances the existing SDK without replacing it:

- Foundation layer (ogen): Accurate API calls
- Wrapper layer: Ergonomic human APIs
- AX layer: Agent metadata and helpers

**Recommendation:** Add AX as a complementary layer, not a replacement.

### 5. Tests Validate Integration

Comprehensive tests ensure the integration works correctly across all scenarios.

**Recommendation:** Test error extraction, retry logic, and validation separately.

## Future Work

### Expanded Error Discovery

- Probe more endpoints systematically
- Track error code changes over API versions
- Automate discovery in CI/CD

### Idempotency Tracking

- Add `x-ax-idempotent` extension
- Generate idempotency key helpers
- Enable safe retries for more operations

### Cost Metadata

- Add `x-ax-cost-estimate` extension
- Track credit usage per operation
- Enable cost-aware agent decisions

### Capability Graphs

- Map related operations
- Enable agent navigation
- Support "what can I do next?" queries

## Conclusion

The AX integration transforms elevenlabs-go from a human-centric SDK to an agent-friendly one:

| Aspect | Before | After |
|--------|--------|-------|
| Error handling | String parsing | Typed constants |
| Retry decisions | Hardcoded | 236 documented policies |
| Validation | Runtime errors | Pre-flight checks |
| Error semantics | None | Category metadata |
| Agent reliability | Guesswork | Deterministic |

By following DIRECT principles and using ax-spec tooling, we've demonstrated that existing SDKs can be enhanced for AI agent consumption with modest effort and significant value.

## References

- [DIRECT Principles](https://github.com/grokify/direct-principles)
- [AX Spec](https://github.com/grokify/ax-spec)
- [Agent Experience Article](https://github.com/grokify/grokify-articles/tree/main/agent-experience-ax)
- [elevenlabs-go ax package](https://github.com/plexusone/elevenlabs-go/tree/main/ax)
