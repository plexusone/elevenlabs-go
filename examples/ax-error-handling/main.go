// Example demonstrating AX (Agent Experience) error handling
//
// This example shows how to use the ax package for machine-readable
// error handling, which is essential for building robust AI agents
// that can programmatically respond to specific error conditions.
package main

import (
	"context"
	"fmt"
	"log"

	elevenlabs "github.com/plexusone/elevenlabs-go"
	"github.com/plexusone/elevenlabs-go/ax"
)

func main() {
	// ELEVENLABS_API_KEY is read from environment automatically
	client, err := elevenlabs.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Example: Attempt to get a non-existent voice
	_, err = client.Voices().Get(ctx, "non-existent-voice-id")
	if err != nil {
		handleError(err)
	}
}

// handleError demonstrates AX-powered error handling for agents.
// Instead of parsing error messages, agents can use error codes.
func handleError(err error) {
	// Method 1: Use the high-level helper
	if elevenlabs.IsAXError(err, ax.ErrDocumentNotFound) {
		fmt.Println("Resource not found - agent should try a different resource")
		return
	}

	// Method 2: Extract the error code for switch-based handling
	if code, ok := elevenlabs.GetAXErrorCode(err); ok {
		switch code {
		case ax.ErrDocumentNotFound, ax.ErrUserNotFound, ax.ErrWorkspaceNotFound:
			fmt.Printf("Not found error: %s\n", code)
			// Agent action: Try alternative resources or report to user

		case ax.ErrNotLoggedIn, ax.ErrNeedsAuthorization:
			fmt.Printf("Auth error: %s\n", code)
			// Agent action: Re-authenticate or request permissions

		case ax.ErrInvalidUID:
			fmt.Printf("Validation error: %s\n", code)
			// Agent action: Fix the input and retry

		case ax.ErrUnprocessableEntity:
			fmt.Printf("Request validation failed: %s\n", code)
			// Agent action: Check required fields and fix request

		default:
			fmt.Printf("Known error code: %s\n", code)
		}

		// Get additional metadata about the error
		if info := ax.GetErrorInfo(code); info != nil {
			fmt.Printf("  Category: %s\n", info.Category)
			fmt.Printf("  Retryable: %v\n", info.Retryable)
			fmt.Printf("  Description: %s\n", info.Description)
		}
		return
	}

	// Method 3: Use structured APIError for HTTP status codes
	if apiErr := elevenlabs.ParseAPIError(err); apiErr != nil {
		// Check for AX code within the APIError
		if axCode, ok := apiErr.AXErrorCode(); ok {
			fmt.Printf("API error with AX code: %s (HTTP %d)\n", axCode, apiErr.StatusCode)
		} else {
			// Fall back to HTTP status code based handling
			switch apiErr.StatusCode {
			case 401:
				fmt.Println("Unauthorized - check API key")
			case 403:
				fmt.Println("Forbidden - insufficient permissions")
			case 429:
				fmt.Println("Rate limited - back off and retry")
			default:
				fmt.Printf("API error: %s\n", apiErr.Error())
			}
		}
		return
	}

	// Generic error (network issues, etc.)
	fmt.Printf("Error: %v\n", err)
}

// Example: Using retry policy for automatic retries
//
//nolint:unused // Example function for users to copy
func shouldRetry(operationID string, err error) bool {
	// Check if the operation is safe to retry
	if !ax.IsRetryable(operationID) {
		return false
	}

	// Check if the error is retryable
	if apiErr := elevenlabs.ParseAPIError(err); apiErr != nil {
		// Retry on rate limits or server errors
		if apiErr.StatusCode == 429 || apiErr.StatusCode >= 500 {
			return true
		}

		// Check AX error code
		if code, ok := apiErr.AXErrorCode(); ok {
			if info := ax.GetErrorInfo(code); info != nil {
				return info.Retryable
			}
		}
	}

	return false
}

// Example: Pre-flight validation using required fields
//
//nolint:unused // Example function for users to copy
func validateRequest(operationID string, fields map[string]bool) error {
	if msg := ax.ValidateFields(operationID, fields); msg != "" {
		return fmt.Errorf("validation failed: %s", msg)
	}
	return nil
}
