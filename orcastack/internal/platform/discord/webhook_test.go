package discord

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebhookClientSend(t *testing.T) {
	// Create a test server that simulates Discord webhook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, _ := io.ReadAll(r.Body)
		var payload WebhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewWebhookClient(server.URL, true)

	payload := &WebhookPayload{
		Content: "Test message",
	}

	err := client.Send(context.Background(), payload)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
}

func TestWebhookClientAsync(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewWebhookClient(server.URL, true)

	payload := &WebhookPayload{
		Content: "Async message",
	}

	err := client.SendAsync(context.Background(), payload)
	if err != nil {
		t.Fatalf("SendAsync failed: %v", err)
	}

	// Wait for background goroutine
	time.Sleep(500 * time.Millisecond)

	if !serverCalled {
		t.Error("Server was not called")
	}
}

func TestWebhookClientDisabled(t *testing.T) {
	client := NewWebhookClient("", false)

	if client.IsEnabled() {
		t.Error("Client should be disabled")
	}

	payload := &WebhookPayload{Content: "Test"}
	err := client.SendAsync(context.Background(), payload)
	if err != nil {
		t.Fatalf("SendAsync should not error when disabled: %v", err)
	}
}

func TestWebhookClientRetry(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewWebhookClient(server.URL, true)
	client.maxRetries = 4 // Allow enough retries

	payload := &WebhookPayload{Content: "Retry test"}
	err := client.Send(context.Background(), payload)
	if err != nil {
		t.Fatalf("Send should succeed after retries: %v", err)
	}

	if attemptCount < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", attemptCount)
	}
}

func TestEmbedBuilder(t *testing.T) {
	embed := NewEmbed("Test Title", "Test Description", ColorSuccess)
	embed.AddField("Field1", "Value1", true)
	embed.AddField("Field2", "Value2", false)
	embed.SetURL("https://example.com")

	if embed.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", embed.Title)
	}

	if len(embed.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(embed.Fields))
	}

	if embed.URL != "https://example.com" {
		t.Errorf("Expected URL, got %s", embed.URL)
	}
}

func TestNotificationEventToPipeline(t *testing.T) {
	event := NewPipelineEvent("success", "pipe-123", "main", "abc123def", "https://orcastack.example.com/pipelines/pipe-123")

	if event.EventType != EventPipelineSuccess {
		t.Errorf("Expected EventPipelineSuccess, got %s", event.EventType)
	}

	if event.Severity != SeverityInfo {
		t.Errorf("Expected SeverityInfo, got %s", event.Severity)
	}

	payload := event.ToPayload()
	if len(payload.Embeds) != 1 {
		t.Fatalf("Expected 1 embed, got %d", len(payload.Embeds))
	}

	embed := payload.Embeds[0]
	if embed.Color != int(ColorInfo) {
		t.Errorf("Expected color %d, got %d", ColorInfo, embed.Color)
	}

	if embed.Title != "Pipeline success" {
		t.Errorf("Expected title 'Pipeline success', got '%s'", embed.Title)
	}
}

func TestNotificationEventToDeployment(t *testing.T) {
	event := NewDeploymentEvent("failure", "production", "api-service", "v1.2.3", "https://orcastack.example.com/deployments/dep-456")

	if event.EventType != EventDeploymentFailure {
		t.Errorf("Expected EventDeploymentFailure, got %s", event.EventType)
	}

	if event.Severity != SeverityCritical {
		t.Errorf("Expected SeverityCritical, got %s", event.Severity)
	}

	payload := event.ToPayload()
	embed := payload.Embeds[0]

	if embed.Color != int(ColorFailure) {
		t.Errorf("Expected color %d, got %d", ColorFailure, embed.Color)
	}
}

func TestNotificationEventToSecurity(t *testing.T) {
	event := NewSecurityEvent("violation", "RBAC violation", "user123", "Attempted access to forbidden resource", "https://orcastack.example.com/security")

	if event.Severity != SeverityCritical {
		t.Errorf("Expected SeverityCritical, got %s", event.Severity)
	}

	payload := event.ToPayload()
	embed := payload.Embeds[0]

	if embed.Color != int(ColorFailure) {
		t.Errorf("Expected color %d, got %d", ColorFailure, embed.Color)
	}
}

func TestNotificationEventToHealth(t *testing.T) {
	event := NewHealthEvent("critical", "memory", "usage_percent", "95.5", "https://orcastack.example.com/health")

	if event.EventType != EventHealthCritical {
		t.Errorf("Expected EventHealthCritical, got %s", event.EventType)
	}

	if event.Severity != SeverityCritical {
		t.Errorf("Expected SeverityCritical, got %s", event.Severity)
	}
}

func TestNotificationEventToUser(t *testing.T) {
	event := NewUserEvent("signup", "john.doe", "john@example.com", "New user registration", "https://orcastack.example.com/users")

	if event.EventType != EventUserSignup {
		t.Errorf("Expected EventUserSignup, got %s", event.EventType)
	}

	if event.Severity != SeverityInfo {
		t.Errorf("Expected SeverityInfo, got %s", event.Severity)
	}
}

func TestWebhookPayloadSerialization(t *testing.T) {
	embed := NewEmbed("Test", "Description", ColorSuccess)
	embed.AddField("Key", "Value", true)

	payload := &WebhookPayload{
		Content: "Test message",
		Embeds:  []Embed{*embed},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	var unmarshaled WebhookPayload
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if unmarshaled.Content != "Test message" {
		t.Errorf("Expected content 'Test message', got '%s'", unmarshaled.Content)
	}

	if len(unmarshaled.Embeds) != 1 {
		t.Errorf("Expected 1 embed, got %d", len(unmarshaled.Embeds))
	}
}
