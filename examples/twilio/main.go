// Example: Twilio Integration - Phone call handling
//
// This example demonstrates integrating ElevenLabs with Twilio for
// voice agent phone calls. It shows how to:
// - Register incoming calls with ElevenLabs agents
// - Make outbound calls
// - Manage phone numbers
//
// Usage:
//
//	export ELEVENLABS_API_KEY="your-api-key"
//	go run main.go
//
// For production, you would run this as an HTTP server handling
// Twilio webhooks.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/grokify/mogo/log/slogutil"
	elevenlabs "github.com/plexusone/elevenlabs-go"
)

var client *elevenlabs.Client

func main() {
	// Create base context with logger
	ctx := slogutil.ContextWithLogger(context.Background(), slog.Default())

	var err error
	client, err = elevenlabs.NewClient()
	if err != nil {
		logError(ctx, "Failed to create client", err)
		os.Exit(1)
	}

	// Demo: List phone numbers
	listPhoneNumbers(ctx)

	// Start HTTP server for Twilio webhooks
	fmt.Println("\nStarting webhook server on :8080...")
	fmt.Println("Configure Twilio webhook URL: http://your-server:8080/twilio/incoming")

	http.HandleFunc("/twilio/incoming", withLogger(handleIncomingCall))
	http.HandleFunc("/api/outbound", withLogger(handleOutboundCall))
	http.HandleFunc("/api/phone-numbers", withLogger(handleListPhoneNumbers))

	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 10 * time.Second,
	}
	logError(ctx, "Server stopped", server.ListenAndServe())
}

// withLogger is middleware that attaches a logger to the request context.
func withLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := slogutil.ContextWithLogger(r.Context(), slog.Default())
		next(w, r.WithContext(ctx))
	}
}

// listPhoneNumbers demonstrates phone number management
func listPhoneNumbers(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	numbers, err := client.PhoneNumbers().List(ctx)
	if err != nil {
		logError(ctx, "Failed to list phone numbers", err)
		return
	}

	fmt.Printf("\nPhone Numbers (%d):\n", len(numbers))
	for _, num := range numbers {
		fmt.Printf("  - %s: %s (Provider: %s, Status: %s)\n",
			num.Label, num.PhoneNumber, num.Provider, num.Status)
		if num.AgentID != "" {
			fmt.Printf("    Agent: %s\n", num.AgentID)
		}
	}
}

// handleIncomingCall handles Twilio webhook for incoming calls
func handleIncomingCall(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse Twilio parameters (limit body to 64KB for webhook)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	callerNumber := r.Form.Get("From")
	calledNumber := r.Form.Get("To")
	callSid := r.Form.Get("CallSid")

	logInfo(ctx, "Incoming call", "from", callerNumber, "to", calledNumber, "sid", callSid)

	// Get agent ID from environment or configuration
	agentID := os.Getenv("ELEVENLABS_AGENT_ID")
	if agentID == "" {
		// Return error TwiML
		w.Header().Set("Content-Type", "application/xml")
		if _, err := w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Say>Sorry, no agent is configured.</Say>
    <Hangup/>
</Response>`)); err != nil {
			logError(ctx, "Failed to write response", err)
		}
		return
	}

	// Register call with ElevenLabs
	resp, err := client.Twilio().RegisterCall(ctx, &elevenlabs.TwilioRegisterCallRequest{
		AgentID: agentID,

		// Inject caller info as dynamic variables
		DynamicVariables: map[string]string{
			"caller_number": callerNumber,
			"call_sid":      callSid,
		},

		// Optional: customize first message
		// FirstMessage: fmt.Sprintf("Hello! I see you're calling from %s.", callerNumber),
	})
	if err != nil {
		logError(ctx, "Failed to register call", err, "agent_id", agentID)
		w.Header().Set("Content-Type", "application/xml")
		if _, err := w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Say>Sorry, there was an error connecting your call.</Say>
    <Hangup/>
</Response>`)); err != nil {
			logError(ctx, "Failed to write response", err)
		}
		return
	}

	logInfo(ctx, "Call registered", "conversation_id", resp.ConversationID)

	// Return TwiML to Twilio
	w.Header().Set("Content-Type", "application/xml")
	if _, err := w.Write([]byte(resp.TwiML)); err != nil { //nolint:gosec // G705: TwiML XML response to Twilio callback, not browser
		logError(ctx, "Failed to write TwiML response", err)
	}
}

// handleOutboundCall initiates an outbound call
func handleOutboundCall(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req struct {
		ToNumber           string            `json:"to_number"`
		AgentID            string            `json:"agent_id"`
		AgentPhoneNumberID string            `json:"agent_phone_number_id"`
		FirstMessage       string            `json:"first_message,omitempty"`
		Variables          map[string]string `json:"variables,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ToNumber == "" || req.AgentID == "" || req.AgentPhoneNumberID == "" {
		http.Error(w, "Missing required fields: to_number, agent_id, agent_phone_number_id", http.StatusBadRequest)
		return
	}

	// Make outbound call
	call, err := client.Twilio().OutboundCall(ctx, &elevenlabs.TwilioOutboundCallRequest{
		AgentID:            req.AgentID,
		AgentPhoneNumberID: req.AgentPhoneNumberID,
		ToNumber:           req.ToNumber,
		FirstMessage:       req.FirstMessage,
		DynamicVariables:   req.Variables,
	})
	if err != nil {
		logError(ctx, "Failed to make outbound call", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logInfo(ctx, "Outbound call initiated", "sid", call.CallSID, "conversation_id", call.ConversationID)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"call_sid":        call.CallSID,
		"conversation_id": call.ConversationID,
		"status":          call.Status,
	}); err != nil {
		logError(ctx, "Failed to encode response", err)
	}
}

// handleListPhoneNumbers returns available phone numbers as JSON
func handleListPhoneNumbers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	numbers, err := client.PhoneNumbers().List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(numbers); err != nil {
		logError(ctx, "Failed to encode response", err)
	}
}

// logInfo logs an info message using the logger from context.
func logInfo(ctx context.Context, msg string, args ...any) {
	slogutil.LoggerFromContext(ctx, slogutil.Null()).Info(msg, args...)
}

// logError logs an error message using the logger from context.
func logError(ctx context.Context, msg string, err error, args ...any) {
	logger := slogutil.LoggerFromContext(ctx, slogutil.Null())
	if err != nil {
		args = append([]any{"error", err}, args...)
	}
	logger.Error(msg, args...)
}

// Example: SIP trunk outbound call
//
//nolint:unused // Example function for documentation
func sipOutboundExample(ctx context.Context, client *elevenlabs.Client) {
	call, err := client.Twilio().SIPOutboundCall(ctx, &elevenlabs.SIPOutboundCallRequest{
		AgentID:    "your-agent-id",
		SIPTrunkID: "your-sip-trunk-id",
		ToNumber:   "+1234567890",
		FromNumber: "+0987654321", // Must be verified

		DynamicVariables: map[string]string{
			"customer_name": "John",
			"order_id":      "12345",
		},
	})
	if err != nil {
		logError(ctx, "SIP call failed", err)
		os.Exit(1)
	}

	logInfo(ctx, "SIP call initiated", "conversation_id", call.ConversationID)
}

// Example: Update phone number settings
//
//nolint:unused // Example function for documentation
func updatePhoneNumberExample(ctx context.Context, client *elevenlabs.Client) {
	updated, err := client.PhoneNumbers().Update(ctx, "phone-number-id",
		&elevenlabs.UpdatePhoneNumberRequest{
			Label:   "Customer Support Line",
			AgentID: "new-agent-id",
		})
	if err != nil {
		logError(ctx, "Failed to update phone number", err)
		os.Exit(1)
	}

	logInfo(ctx, "Updated phone number", "number", updated.PhoneNumber)
}
