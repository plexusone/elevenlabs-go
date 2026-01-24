# Testing

This SDK includes both unit tests and integration tests.

## Unit Tests

Run unit tests (no API key required):

```bash
go test ./...
```

Unit tests use mocked HTTP responses and don't make real API calls.

## Integration Tests

Integration tests make real API calls to verify the SDK works correctly with the ElevenLabs API. They catch issues like:

- Nullable field handling in API responses
- API response format changes
- Authentication and permissions

### Running Integration Tests

```bash
ELEVENLABS_API_KEY=your_key go test -v -tags=integration ./...
```

### How They Work

Integration tests use Go build tags to ensure they only run when explicitly requested:

```go
//go:build integration

package elevenlabs
```

They also skip automatically if no API key is set:

```go
func skipIfNoAPIKey(t *testing.T) {
    if os.Getenv("ELEVENLABS_API_KEY") == "" {
        t.Skip("ELEVENLABS_API_KEY not set, skipping integration test")
    }
}
```

### Available Tests

| Test | Endpoint | Purpose |
|------|----------|---------|
| `TestVoicesListNullHandling` | `/v1/voices` | Verifies nullable fields in voice responses |
| `TestModelsListNullHandling` | `/v1/models` | Verifies nullable fields in model responses |
| `TestUserGetNullHandling` | `/v1/user` | Verifies nullable fields in user/subscription |
| `TestHistoryListNullHandling` | `/v1/history` | Verifies nullable fields in history items |

### API Key Permissions

Some tests may skip if your API key doesn't have access to certain endpoints. This is expected behavior - the tests handle 401 errors gracefully:

```go
func skipOn401(t *testing.T, err error) {
    if err != nil && strings.Contains(err.Error(), "401") {
        t.Skipf("API key does not have access to this endpoint: %v", err)
    }
}
```

### CI Integration

For GitHub Actions, add the API key as a secret and run integration tests:

```yaml
- name: Run integration tests
  if: ${{ secrets.ELEVENLABS_API_KEY != '' }}
  env:
    ELEVENLABS_API_KEY: ${{ secrets.ELEVENLABS_API_KEY }}
  run: go test -v -tags=integration ./...
```

## Why Integration Tests Matter

The SDK uses [ogen](https://github.com/ogen-go/ogen) to generate API client code from the ElevenLabs OpenAPI spec. There's a known issue ([ogen-go/ogen#1358](https://github.com/ogen-go/ogen/issues/1358)) where nullable `$ref` fields don't decode `null` values correctly.

We use [ogen-tools](https://github.com/agentplexus/ogen-tools) to post-process the generated code, and integration tests verify these fixes work against the real API.

### Example: Null Handling Issue

Without the fix, this API response would fail:

```json
{
  "voices": [{
    "voice_id": "abc123",
    "name": "Test Voice",
    "fine_tuning": {
      "manual_verification": null  // This null caused decode errors
    }
  }]
}
```

Error:
```
decode ManualVerificationResponseModel: "{" expected: unexpected byte 110 'n'
```

Integration tests catch these issues before users do.

## Test File Location

Integration tests are in `integration_test.go` at the project root (not in `internal/api/`) so they won't be overwritten when regenerating ogen code with `--clean`.
