# Installation

## Requirements

- Go 1.21 or later
- An ElevenLabs API key ([get one here](https://elevenlabs.io/))

## Install the SDK

```bash
go get github.com/agentplexus/go-elevenlabs
```

## Get Your API Key

1. Sign up at [elevenlabs.io](https://elevenlabs.io/)
2. Go to your Profile Settings
3. Copy your API key

## Set Up Authentication

The SDK can read your API key from the environment:

```bash
export ELEVENLABS_API_KEY=your-api-key-here
```

Or pass it directly when creating the client:

```go
client, err := elevenlabs.NewClient(
    elevenlabs.WithAPIKey("your-api-key-here"),
)
```

## Verify Installation

```go
package main

import (
    "context"
    "fmt"
    "log"

    elevenlabs "github.com/agentplexus/go-elevenlabs"
)

func main() {
    client, err := elevenlabs.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    // Check subscription
    sub, err := client.User().GetSubscription(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Connected! Tier: %s, Characters remaining: %d\n",
        sub.Tier, sub.CharactersRemaining())
}
```

## Next Steps

- [Quick Start](quickstart.md) - Generate your first audio
- [Configuration](configuration.md) - Advanced client options
