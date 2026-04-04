---
marp: true
theme: default
paginate: true
backgroundColor: #fff
---

# Agent Experience (AX) Case Study

## Integrating AX Spec into elevenlabs-go

Building Agent-Friendly SDKs with Machine-Readable Metadata

---

# The Challenge

AI agents interact with APIs differently than humans:

- **Agents execute** — humans interpret
- **Agents need structure** — humans tolerate ambiguity
- **Agents iterate** — humans debug manually

Traditional SDKs optimize for human ergonomics, not agent reliability.

---

# The Problem

When an agent encounters an error:

```json
{
  "detail": "Document not found"
}
```

**What can the agent do?**

- Parse the string? Fragile.
- Retry? Maybe unsafe.
- Try alternatives? Which ones?

---

# The Vision: Agent Experience (AX)

Design interfaces that enable agents to:

> **understand → call → recover → iterate**

Key insight: **Machine-readable metadata** enables autonomous error handling.

---

# DIRECT Principles

| Principle | Agent Benefit |
|-----------|---------------|
| **D**eterministic | Same input = same output |
| **I**ntrospectable | Discover capabilities programmatically |
| **R**ecoverable | Structured errors enable auto-fix |
| **E**xplicit | All constraints in specification |
| **C**onsistent | Patterns generalize across endpoints |
| **T**estable | Safe experimentation in sandbox |

---

# The Project: elevenlabs-go

An SDK for the ElevenLabs API:

- **204 endpoints** (text-to-speech, voice cloning, dubbing)
- **54K line OpenAPI spec**
- **330K lines of generated Go code**
- Used by AI agents for voice generation

---

# Implementation Approach

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  OpenAPI Spec   │────▶│   ax-spec CLI   │────▶│  Generated Go   │
│                 │     │                 │     │     Code        │
│ elevenlabs.json │     │ enrich + gen    │     │ ax/errors.go    │
│                 │     │                 │     │ ax/retry.go     │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌─────────────────┐
                        │  API Discovery  │
                        │                 │
                        │ Real API calls  │
                        │ to find errors  │
                        └─────────────────┘
```

---

# Step 1: API Discovery

Made real API calls to discover actual error codes:

```bash
ax-spec enrich elevenlabs-openapi.json \
  --discover \
  --api-key $ELEVENLABS_API_KEY
```

**Result:** 9 distinct error codes discovered

---

# Discovered Error Codes

| Code | Category | Description |
|------|----------|-------------|
| `DOCUMENT_NOT_FOUND` | not_found | Resource doesn't exist |
| `USER_NOT_FOUND` | not_found | User doesn't exist |
| `NOT_LOGGED_IN` | auth | Not authenticated |
| `NEEDS_AUTHORIZATION` | auth | Needs permissions |
| `INVALID_UID` | validation | Bad identifier |
| `UNPROCESSABLE_ENTITY` | validation | Request invalid |

---

# Step 2: Code Generation

```bash
ax-spec gen elevenlabs-openapi-ax.json \
  --output ax/ \
  --package ax
```

**Generated files:**

| File | Content |
|------|---------|
| `errors.go` | 9 error constants + metadata |
| `retry.go` | 236 retry policies |
| `validation.go` | 72 required field mappings |
| `capabilities.go` | Operation capabilities |

---

# Step 3: SDK Integration

Enhanced the existing `errors.go` with AX support:

```go
// New method on APIError
func (e *APIError) AXErrorCode() (string, bool)

// Top-level helpers
func IsAXError(err error, code string) bool
func GetAXErrorCode(err error) (string, bool)
```

---

# Before: String-Based Error Handling

```go
if strings.Contains(err.Error(), "not found") {
    // Hope this catches all cases...
}

if apiErr.StatusCode == 404 {
    // Is it a document? User? Workspace?
}
```

**Problems:**

- Fragile string matching
- No semantic understanding
- Can't distinguish error types

---

# After: AX Error Handling

```go
if elevenlabs.IsAXError(err, ax.ErrDocumentNotFound) {
    // Specific, reliable match
}

if code, ok := elevenlabs.GetAXErrorCode(err); ok {
    switch code {
    case ax.ErrDocumentNotFound:
        // Try alternative document
    case ax.ErrNeedsAuthorization:
        // Request permissions
    }
}
```

---

# Error Metadata

Each error code has rich metadata:

```go
info := ax.GetErrorInfo(ax.ErrDocumentNotFound)
// info.Code        = "DOCUMENT_NOT_FOUND"
// info.Category    = "not_found"
// info.Retryable   = false
// info.Description = "The requested document was not found"
```

Agents can make informed decisions:

- **not_found** → Try alternatives
- **auth** → Re-authenticate
- **validation** → Fix request

---

# Retry Policy

Before: Guess which operations are safe to retry

After: 236 operations with explicit retry safety

```go
if ax.IsRetryable("get_voices") {
    // Safe: GET request, no side effects
    retry.WithBackoff(...)
}

if !ax.IsRetryable("create_voice") {
    // Unsafe: Would create duplicates
    return err
}
```

---

# Retry Policy Distribution

| Category | Count | Retryable |
|----------|-------|-----------|
| GET operations | 98 | Yes |
| POST/PUT operations | 92 | No |
| DELETE operations | 26 | No |
| Special cases | 20 | Varies |

**Key insight:** Only GET operations default to retryable.

---

# Pre-flight Validation

Before: Discover required fields at runtime (via API errors)

After: Validate before calling

```go
fields := map[string]bool{
    "text": true,
    // Missing: voice_id
}

msg := ax.ValidateFields("text_to_speech_full", fields)
// msg = "missing required fields: voice_id"

// Don't call API - fix first
```

---

# Required Fields Coverage

72 operations have required field definitions:

```go
var RequiredFields = map[string][]string{
    "text_to_speech_full": {"text"},
    "create_voice":        {"voice_name", "voice_description", "generated_voice_id"},
    "create_batch_call":   {"call_name", "agent_id", "recipients"},
    // ... 69 more
}
```

---

# Results Summary

| Metric | Before | After |
|--------|--------|-------|
| Error handling | String parsing | Typed constants |
| Retry decisions | Hardcoded | 236 policies |
| Required fields | Runtime errors | Pre-validation |
| Error categories | Manual mapping | Auto-classified |
| Agent behavior | Guesswork | Deterministic |

---

# Agent Reliability Improvement

**Before AX:**

```
Agent: Call API
API: Error "document not found"
Agent: Parse string... maybe retry?
Agent: Retry
API: Same error
Agent: Give up
```

**After AX:**

```
Agent: Call API
API: Error DOCUMENT_NOT_FOUND
Agent: Check ax.GetErrorInfo() → category=not_found, retryable=false
Agent: Try alternative document
Agent: Success
```

---

# Code Impact

**New files:**

- `ax/doc.go` - Package documentation
- `ax/errors.go` - 9 error constants
- `ax/retry.go` - 236 retry policies
- `ax/validation.go` - 72 required field maps
- `ax/capabilities.go` - Operation capabilities
- `ax/ax_test.go` - Unit tests

**Modified files:**

- `errors.go` - Added AX integration methods

---

# Example: Complete Error Handling

```go
resp, err := client.Voices().Get(ctx, voiceID)
if err != nil {
    if code, ok := elevenlabs.GetAXErrorCode(err); ok {
        info := ax.GetErrorInfo(code)

        log.Printf("Error: %s (category=%s, retryable=%v)",
            code, info.Category, info.Retryable)

        if info.Retryable {
            return retry(ctx, func() error { ... })
        }

        if ax.IsNotFoundError(code) {
            return useDefaultVoice(ctx)
        }
    }
    return err
}
```

---

# Key Learnings

1. **API discovery is essential** — Specs don't document all error codes
2. **Code generation scales** — 204 endpoints, one command
3. **Metadata enables decisions** — Categories, retry safety, descriptions
4. **Layers work together** — AX enhances existing SDK, doesn't replace
5. **Tests validate integration** — All 30+ tests pass

---

# The AX Workflow

```bash
# 1. Lint spec for AX compliance
ax-spec lint api.yaml

# 2. Enrich with x-ax-* extensions + discover errors
ax-spec enrich api.yaml --discover --api-key $KEY

# 3. Generate SDK code
ax-spec gen api-ax.yaml --output ax/

# 4. Integrate with existing SDK
# Add IsAXError(), GetAXErrorCode() helpers
```

---

# What's Next

- **Expand error discovery** — More endpoints, more codes
- **Idempotency tracking** — `x-ax-idempotent` for safe retries
- **Cost metadata** — `x-ax-cost-estimate` for resource planning
- **Capability graphs** — Related operations for agent navigation

---

# Resources

- **DIRECT Principles:** github.com/grokify/direct-principles
- **AX Spec:** github.com/grokify/ax-spec
- **elevenlabs-go:** github.com/plexusone/elevenlabs-go
- **Article:** grokify-articles/agent-experience-ax

---

# Summary

Agent Experience (AX) transforms how SDKs support AI agents:

| Traditional SDK | AX-Enhanced SDK |
|-----------------|-----------------|
| Human-centric | Agent-friendly |
| String errors | Typed error codes |
| Implicit rules | Explicit metadata |
| Trial and error | Informed decisions |

**AX is not just better DX — it's a new foundation for agent interoperability.**

---

# Thank You

Questions?

---

# Appendix: Error Code Reference

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
