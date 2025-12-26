//go:build ignore

package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	defaultURL    = "http://localhost:3000/api/v1/webhooks/logto"
	defaultSecret = "test_secret_key_12345"
)

// LogtoWebhookEvent represents a webhook event from Logto
type LogtoWebhookEvent struct {
	Event            string                 `json:"event"`
	CreatedAt        string                 `json:"createdAt"`
	SessionID        string                 `json:"sessionId,omitempty"`
	UserAgent        string                 `json:"userAgent,omitempty"`
	IP               string                 `json:"ip,omitempty"`
	InteractionEvent bool                   `json:"interactionEvent,omitempty"`
	Data             map[string]interface{} `json:"data"`
}

func main() {
	// Command line flags
	webhookURL := flag.String("url", defaultURL, "Webhook endpoint URL")
	secret := flag.String("secret", "", "Webhook signing secret (uses LOGTO_WEBHOOK_SECRET env var if not provided)")
	eventType := flag.String("event", "User.Created", "Event type to test (User.Created, User.Data.Updated, User.Deleted, User.SuspensionStatus.Updated)")
	skipSignature := flag.Bool("skip-sig", false, "Skip signature verification (for testing)")
	flag.Parse()

	// Get secret from env if not provided
	signingSecret := *secret
	if signingSecret == "" {
		if envSecret := os.Getenv("LOGTO_WEBHOOK_SECRET"); envSecret != "" {
			signingSecret = envSecret
		} else {
			signingSecret = defaultSecret
			fmt.Printf("⚠️  Using default test secret. Set LOGTO_WEBHOOK_SECRET env var or use -secret flag for production\n")
		}
	}

	// Create test event based on type
	var event LogtoWebhookEvent
	switch *eventType {
	case "User.Created":
		event = createUserCreatedEvent()
	case "User.Data.Updated":
		event = createUserUpdatedEvent()
	case "User.Deleted":
		event = createUserDeletedEvent()
	case "User.SuspensionStatus.Updated":
		event = createUserSuspensionEvent()
	default:
		fmt.Printf("❌ Unknown event type: %s\n", *eventType)
		fmt.Println("Valid types: User.Created, User.Data.Updated, User.Deleted, User.SuspensionStatus.Updated")
		os.Exit(1)
	}

	// Send webhook request
	if err := sendWebhookRequest(*webhookURL, event, signingSecret, *skipSignature); err != nil {
		fmt.Printf("❌ Error sending webhook: %v\n", err)
		os.Exit(1)
	}
}

func createUserCreatedEvent() LogtoWebhookEvent {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	userID := fmt.Sprintf("test_user_%d", time.Now().Unix())

	return LogtoWebhookEvent{
		Event:            "User.Created",
		CreatedAt:        timestamp,
		SessionID:        "test_session_123",
		UserAgent:        "Test-Webhook-Client/1.0",
		IP:               "127.0.0.1",
		InteractionEvent: false,
		Data: map[string]interface{}{
			"id":           userID,
			"primaryEmail": fmt.Sprintf("testuser%d@example.com", time.Now().Unix()),
			"primaryPhone": "+1234567890",
			"name":         "Test User",
			"avatar":       "https://example.com/avatar.jpg",
			"customData": map[string]interface{}{
				"role":             "customer",
				"is_platform_user": false,
			},
			"profile": map[string]interface{}{
				"bio": "Test user bio",
			},
			"applicationId":   "test_app_123",
			"isSuspended":     false,
			"lastSignInAt":    time.Now().UnixMilli(),
			"createdAt":       time.Now().UnixMilli(),
			"updatedAt":       time.Now().UnixMilli(),
			"isEmailVerified": true,
			"isPhoneVerified": false,
		},
	}
}

func createUserUpdatedEvent() LogtoWebhookEvent {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	// Use a known user ID that should already exist
	userID := "test_user_existing"

	return LogtoWebhookEvent{
		Event:            "User.Data.Updated",
		CreatedAt:        timestamp,
		SessionID:        "test_session_456",
		UserAgent:        "Test-Webhook-Client/1.0",
		IP:               "127.0.0.1",
		InteractionEvent: false,
		Data: map[string]interface{}{
			"id":           userID,
			"primaryEmail": "updated.email@example.com",
			"name":         "Updated Test User",
			"avatar":       "https://example.com/new-avatar.jpg",
			"customData": map[string]interface{}{
				"role":        "artisan",
				"updated_via": "webhook",
			},
			"isSuspended":     false,
			"updatedAt":       time.Now().UnixMilli(),
			"isEmailVerified": true,
			"isPhoneVerified": true,
		},
	}
}

func createUserDeletedEvent() LogtoWebhookEvent {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	userID := "test_user_to_delete"

	return LogtoWebhookEvent{
		Event:            "User.Deleted",
		CreatedAt:        timestamp,
		SessionID:        "test_session_789",
		UserAgent:        "Test-Webhook-Client/1.0",
		IP:               "127.0.0.1",
		InteractionEvent: false,
		Data: map[string]interface{}{
			"id": userID,
		},
	}
}

func createUserSuspensionEvent() LogtoWebhookEvent {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	userID := "test_user_suspension"

	return LogtoWebhookEvent{
		Event:            "User.SuspensionStatus.Updated",
		CreatedAt:        timestamp,
		SessionID:        "test_session_999",
		UserAgent:        "Test-Webhook-Client/1.0",
		IP:               "127.0.0.1",
		InteractionEvent: false,
		Data: map[string]interface{}{
			"id":          userID,
			"isSuspended": true,
		},
	}
}

func sendWebhookRequest(url string, event LogtoWebhookEvent, secret string, skipSignature bool) error {
	// Marshal event to JSON
	payload, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	fmt.Printf("📤 Sending webhook request to: %s\n", url)
	fmt.Printf("📋 Event type: %s\n", event.Event)
	fmt.Printf("📄 Payload:\n%s\n\n", string(payload))

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Calculate and set signature if not skipping
	if !skipSignature {
		signature := calculateSignature(payload, secret)
		req.Header.Set("Logto-Signature-Sha-256", signature)
		fmt.Printf("🔐 Signature: %s\n\n", signature)
	} else {
		fmt.Printf("⚠️  Skipping signature (testing error handling)\n\n")
	}

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Format response JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		body = prettyJSON.Bytes()
	}

	// Print response
	fmt.Printf("📥 Response Status: %s\n", resp.Status)
	fmt.Printf("📥 Response Body:\n%s\n", string(body))

	// Check status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("\n✅ Webhook sent successfully!\n")
		return nil
	} else {
		fmt.Printf("\n⚠️  Webhook returned non-success status: %d\n", resp.StatusCode)
		return nil
	}
}

func calculateSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))
	return signature
}
