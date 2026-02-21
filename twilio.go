package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// TwilioService handles Twilio phone integration for conversational AI.
type TwilioService struct {
	client *Client
}

// postJSON is a helper for making JSON POST requests.
func (s *TwilioService) postJSON(ctx context.Context, path string, req any, result any) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		s.client.baseURL+path,
		bytes.NewReader(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("xi-api-key", s.client.apiKey)

	resp, err := http.DefaultClient.Do(httpReq) //nolint:gosec // G704: API client, URL is fixed ElevenLabs endpoint
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// TwilioRegisterCallRequest is the request to register an incoming Twilio call.
type TwilioRegisterCallRequest struct {
	// AgentID is the ElevenLabs agent ID to handle the call.
	AgentID string `json:"agent_id"`

	// AgentPhoneNumberID is the ElevenLabs phone number ID (if using imported number).
	AgentPhoneNumberID string `json:"agent_phone_number_id,omitempty"`

	// CustomLLMExtraBody is additional data to pass to the LLM.
	CustomLLMExtraBody map[string]any `json:"custom_llm_extra_body,omitempty"`

	// DynamicVariables are variables to inject into the agent prompt.
	DynamicVariables map[string]string `json:"dynamic_variables,omitempty"`

	// FirstMessage overrides the agent's default first message.
	FirstMessage string `json:"first_message,omitempty"`

	// SystemPrompt overrides the agent's system prompt.
	SystemPrompt string `json:"system_prompt,omitempty"`
}

// TwilioRegisterCallResponse is the response from registering a call.
type TwilioRegisterCallResponse struct {
	// TwiML is the TwiML response to return to Twilio.
	TwiML string `json:"twiml"`

	// ConversationID is the ElevenLabs conversation ID for this call.
	ConversationID string `json:"conversation_id,omitempty"`
}

// TwilioOutboundCallRequest is the request to make an outbound call via Twilio.
type TwilioOutboundCallRequest struct {
	// AgentID is the ElevenLabs agent ID to handle the call.
	AgentID string `json:"agent_id"`

	// AgentPhoneNumberID is the ElevenLabs phone number ID to call from.
	AgentPhoneNumberID string `json:"agent_phone_number_id"`

	// ToNumber is the phone number to call (E.164 format).
	ToNumber string `json:"to_number"`

	// CustomLLMExtraBody is additional data to pass to the LLM.
	CustomLLMExtraBody map[string]any `json:"custom_llm_extra_body,omitempty"`

	// DynamicVariables are variables to inject into the agent prompt.
	DynamicVariables map[string]string `json:"dynamic_variables,omitempty"`

	// FirstMessage overrides the agent's default first message.
	FirstMessage string `json:"first_message,omitempty"`

	// SystemPrompt overrides the agent's system prompt.
	SystemPrompt string `json:"system_prompt,omitempty"`
}

// TwilioOutboundCallResponse is the response from making an outbound call.
type TwilioOutboundCallResponse struct {
	// CallSID is the Twilio call SID.
	CallSID string `json:"call_sid"`

	// ConversationID is the ElevenLabs conversation ID for this call.
	ConversationID string `json:"conversation_id"`

	// Status is the initial call status.
	Status string `json:"status"`
}

// SIPOutboundCallRequest is the request to make an outbound call via SIP trunk.
type SIPOutboundCallRequest struct {
	// AgentID is the ElevenLabs agent ID to handle the call.
	AgentID string `json:"agent_id"`

	// ToNumber is the phone number to call (E.164 format).
	ToNumber string `json:"to_number"`

	// SIPTrunkID is the SIP trunk ID to use.
	SIPTrunkID string `json:"sip_trunk_id"`

	// FromNumber is the caller ID to display (must be verified).
	FromNumber string `json:"from_number,omitempty"`

	// CustomLLMExtraBody is additional data to pass to the LLM.
	CustomLLMExtraBody map[string]any `json:"custom_llm_extra_body,omitempty"`

	// DynamicVariables are variables to inject into the agent prompt.
	DynamicVariables map[string]string `json:"dynamic_variables,omitempty"`

	// FirstMessage overrides the agent's default first message.
	FirstMessage string `json:"first_message,omitempty"`

	// SystemPrompt overrides the agent's system prompt.
	SystemPrompt string `json:"system_prompt,omitempty"`
}

// SIPOutboundCallResponse is the response from making a SIP outbound call.
type SIPOutboundCallResponse struct {
	// ConversationID is the ElevenLabs conversation ID for this call.
	ConversationID string `json:"conversation_id"`

	// Status is the initial call status.
	Status string `json:"status"`
}

// RegisterCall registers an incoming Twilio call with ElevenLabs.
// Returns TwiML that should be returned to Twilio's webhook.
func (s *TwilioService) RegisterCall(ctx context.Context, req *TwilioRegisterCallRequest) (*TwilioRegisterCallResponse, error) {
	if req.AgentID == "" {
		return nil, &APIError{Message: "agent_id is required"}
	}

	var result TwilioRegisterCallResponse
	if err := s.postJSON(ctx, "/v1/convai/twilio/register-call", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// OutboundCall initiates an outbound call via Twilio.
func (s *TwilioService) OutboundCall(ctx context.Context, req *TwilioOutboundCallRequest) (*TwilioOutboundCallResponse, error) {
	if req.AgentID == "" {
		return nil, &APIError{Message: "agent_id is required"}
	}
	if req.AgentPhoneNumberID == "" {
		return nil, &APIError{Message: "agent_phone_number_id is required"}
	}
	if req.ToNumber == "" {
		return nil, &APIError{Message: "to_number is required"}
	}

	var result TwilioOutboundCallResponse
	if err := s.postJSON(ctx, "/v1/convai/twilio/outbound-call", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SIPOutboundCall initiates an outbound call via SIP trunk.
func (s *TwilioService) SIPOutboundCall(ctx context.Context, req *SIPOutboundCallRequest) (*SIPOutboundCallResponse, error) {
	if req.AgentID == "" {
		return nil, &APIError{Message: "agent_id is required"}
	}
	if req.SIPTrunkID == "" {
		return nil, &APIError{Message: "sip_trunk_id is required"}
	}
	if req.ToNumber == "" {
		return nil, &APIError{Message: "to_number is required"}
	}

	var result SIPOutboundCallResponse
	if err := s.postJSON(ctx, "/v1/convai/sip-trunk/outbound-call", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PhoneNumberService handles phone number management.
type PhoneNumberService struct {
	client *Client
}

// PhoneNumber represents an ElevenLabs phone number.
type PhoneNumber struct {
	ID          string `json:"phone_number_id"`
	PhoneNumber string `json:"phone_number"`
	Label       string `json:"label"`
	AgentID     string `json:"agent_id,omitempty"`
	Provider    string `json:"provider"` // "twilio", "sip"
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// ListPhoneNumbersResponse is the response from listing phone numbers.
type ListPhoneNumbersResponse struct {
	PhoneNumbers []PhoneNumber `json:"phone_numbers"`
}

// List lists all phone numbers in the workspace.
func (s *PhoneNumberService) List(ctx context.Context) ([]PhoneNumber, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET",
		s.client.baseURL+"/v1/convai/phone-numbers",
		nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("xi-api-key", s.client.apiKey)

	resp, err := http.DefaultClient.Do(httpReq) //nolint:gosec // G704: API client, URL is fixed ElevenLabs endpoint
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	var result ListPhoneNumbersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.PhoneNumbers, nil
}

// Get retrieves a specific phone number by ID.
func (s *PhoneNumberService) Get(ctx context.Context, phoneNumberID string) (*PhoneNumber, error) {
	if phoneNumberID == "" {
		return nil, &APIError{Message: "phone_number_id is required"}
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET",
		s.client.baseURL+"/v1/convai/phone-numbers/"+phoneNumberID,
		nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("xi-api-key", s.client.apiKey)

	resp, err := http.DefaultClient.Do(httpReq) //nolint:gosec // G704: API client, URL is fixed ElevenLabs endpoint
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	var result PhoneNumber
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UpdatePhoneNumberRequest is the request to update a phone number.
type UpdatePhoneNumberRequest struct {
	// Label is a descriptive label for the phone number.
	Label string `json:"label,omitempty"`

	// AgentID is the agent to associate with this phone number.
	AgentID string `json:"agent_id,omitempty"`
}

// Update updates a phone number's settings.
func (s *PhoneNumberService) Update(ctx context.Context, phoneNumberID string, req *UpdatePhoneNumberRequest) (*PhoneNumber, error) {
	if phoneNumberID == "" {
		return nil, &APIError{Message: "phone_number_id is required"}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PATCH",
		s.client.baseURL+"/v1/convai/phone-numbers/"+phoneNumberID,
		bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("xi-api-key", s.client.apiKey)

	resp, err := http.DefaultClient.Do(httpReq) //nolint:gosec // G704: API client, URL is fixed ElevenLabs endpoint
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	var result PhoneNumber
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Delete removes a phone number from the workspace.
func (s *PhoneNumberService) Delete(ctx context.Context, phoneNumberID string) error {
	if phoneNumberID == "" {
		return &APIError{Message: "phone_number_id is required"}
	}

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE",
		s.client.baseURL+"/v1/convai/phone-numbers/"+phoneNumberID,
		nil)
	if err != nil {
		return err
	}

	httpReq.Header.Set("xi-api-key", s.client.apiKey)

	resp, err := http.DefaultClient.Do(httpReq) //nolint:gosec // G704: API client, URL is fixed ElevenLabs endpoint
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return nil
}
