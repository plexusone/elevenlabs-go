# Twilio Integration

Phone call integration for ElevenLabs conversational AI agents.

## Overview

The Twilio integration enables:

- **Incoming Calls**: Route Twilio calls to ElevenLabs agents
- **Outbound Calls**: Initiate calls from ElevenLabs agents
- **SIP Trunks**: Connect via SIP infrastructure
- **Phone Number Management**: Manage agent phone numbers

## Registering Incoming Calls

When Twilio receives a call, register it with ElevenLabs:

```go
// In your Twilio webhook handler
func handleIncomingCall(w http.ResponseWriter, r *http.Request) {
    resp, err := client.Twilio().RegisterCall(ctx, &elevenlabs.TwilioRegisterCallRequest{
        AgentID: "your-agent-id",
    })
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Return TwiML to Twilio
    w.Header().Set("Content-Type", "application/xml")
    w.Write([]byte(resp.TwiML))
}
```

## With Dynamic Variables

Inject context into the agent conversation:

```go
resp, err := client.Twilio().RegisterCall(ctx, &elevenlabs.TwilioRegisterCallRequest{
    AgentID: "your-agent-id",

    // Dynamic variables for prompt injection
    DynamicVariables: map[string]string{
        "caller_name":    callerInfo.Name,
        "account_number": callerInfo.AccountNumber,
        "call_reason":    "support",
    },

    // Override first message
    FirstMessage: fmt.Sprintf("Hello %s, how can I help you today?", callerInfo.Name),

    // Override system prompt
    SystemPrompt: "You are a helpful customer support agent...",
})
```

## Making Outbound Calls

Initiate calls from your ElevenLabs agent:

```go
call, err := client.Twilio().OutboundCall(ctx, &elevenlabs.TwilioOutboundCallRequest{
    // Required fields
    AgentID:            "your-agent-id",
    AgentPhoneNumberID: "phone-number-id",
    ToNumber:           "+1234567890",

    // Optional overrides
    FirstMessage: "Hi, this is a call from your service.",
    DynamicVariables: map[string]string{
        "customer_name": "John",
        "order_id":      "12345",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Call SID: %s\n", call.CallSID)
fmt.Printf("Conversation ID: %s\n", call.ConversationID)
```

## SIP Trunk Outbound Calls

For SIP-based infrastructure:

```go
call, err := client.Twilio().SIPOutboundCall(ctx, &elevenlabs.SIPOutboundCallRequest{
    AgentID:    "your-agent-id",
    SIPTrunkID: "sip-trunk-id",
    ToNumber:   "+1234567890",
    FromNumber: "+0987654321", // Verified caller ID

    DynamicVariables: map[string]string{
        "context": "outbound_campaign",
    },
})
```

## Phone Number Management

### List Phone Numbers

```go
numbers, err := client.PhoneNumbers().List(ctx)
if err != nil {
    log.Fatal(err)
}

for _, num := range numbers {
    fmt.Printf("%s: %s (%s)\n", num.Label, num.PhoneNumber, num.Status)
}
```

### Get Phone Number Details

```go
number, err := client.PhoneNumbers().Get(ctx, "phone-number-id")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Number: %s\n", number.PhoneNumber)
fmt.Printf("Agent: %s\n", number.AgentID)
fmt.Printf("Provider: %s\n", number.Provider)
```

### Update Phone Number

```go
updated, err := client.PhoneNumbers().Update(ctx, "phone-number-id",
    &elevenlabs.UpdatePhoneNumberRequest{
        Label:   "Customer Support Line",
        AgentID: "new-agent-id",
    })
```

### Delete Phone Number

```go
err := client.PhoneNumbers().Delete(ctx, "phone-number-id")
```

## Request Types

### TwilioRegisterCallRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `AgentID` | string | Yes | ElevenLabs agent ID |
| `AgentPhoneNumberID` | string | No | Phone number ID |
| `DynamicVariables` | map[string]string | No | Prompt variables |
| `FirstMessage` | string | No | Override first message |
| `SystemPrompt` | string | No | Override system prompt |
| `CustomLLMExtraBody` | map[string]any | No | Extra LLM data |

### TwilioOutboundCallRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `AgentID` | string | Yes | ElevenLabs agent ID |
| `AgentPhoneNumberID` | string | Yes | From phone number ID |
| `ToNumber` | string | Yes | Destination (E.164 format) |
| `DynamicVariables` | map[string]string | No | Prompt variables |
| `FirstMessage` | string | No | Override first message |
| `SystemPrompt` | string | No | Override system prompt |

### SIPOutboundCallRequest

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `AgentID` | string | Yes | ElevenLabs agent ID |
| `SIPTrunkID` | string | Yes | SIP trunk ID |
| `ToNumber` | string | Yes | Destination (E.164 format) |
| `FromNumber` | string | No | Caller ID (must be verified) |
| `DynamicVariables` | map[string]string | No | Prompt variables |

## Example: Full Integration

```go
package main

import (
    "context"
    "log"
    "net/http"

    elevenlabs "github.com/agentplexus/go-elevenlabs"
)

var client *elevenlabs.Client

func main() {
    var err error
    client, err = elevenlabs.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    http.HandleFunc("/twilio/incoming", handleIncoming)
    http.HandleFunc("/api/call", handleOutbound)

    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIncoming(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Get caller info from Twilio parameters
    callerNumber := r.FormValue("From")

    resp, err := client.Twilio().RegisterCall(ctx, &elevenlabs.TwilioRegisterCallRequest{
        AgentID: "your-agent-id",
        DynamicVariables: map[string]string{
            "caller_number": callerNumber,
        },
    })
    if err != nil {
        log.Printf("register call error: %v", err)
        http.Error(w, "error", 500)
        return
    }

    w.Header().Set("Content-Type", "application/xml")
    w.Write([]byte(resp.TwiML))
}

func handleOutbound(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    toNumber := r.URL.Query().Get("to")

    call, err := client.Twilio().OutboundCall(ctx, &elevenlabs.TwilioOutboundCallRequest{
        AgentID:            "your-agent-id",
        AgentPhoneNumberID: "phone-number-id",
        ToNumber:           toNumber,
    })
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    w.Write([]byte("Call initiated: " + call.CallSID))
}
```
