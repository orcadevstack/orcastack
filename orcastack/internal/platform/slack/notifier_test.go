package slack

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestNotifierClientSend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		body, _ := io.ReadAll(r.Body)
		var msg Message
		if err := json.Unmarshal(body, &msg); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp := map[string]interface{}{
			"ok": true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create notifier with test server URL
	notifier := NewNotifierClient("test-token", true)
	notifier.httpClient = server.Client()
	notifier.apiURL = server.URL

	blocks := []Block{
		NewSectionBlock("Test message"),
	}

	err := notifier.Send(context.Background(), "#test", "Test", blocks)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
}

func TestNotifierClientDisabled(t *testing.T) {
	notifier := NewNotifierClient("", false)

	if notifier.IsEnabled() {
		t.Error("Notifier should be disabled")
	}

	blocks := []Block{}
	err := notifier.SendAsync(context.Background(), "#test", "Test", blocks)
	if err != nil {
		t.Fatalf("SendAsync should not error when disabled: %v", err)
	}
}

func TestEventVerifierValidSignature(t *testing.T) {
	signingSecret := "test-secret"
	verifier := NewEventVerifier(signingSecret)

	// Create a valid request
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"type":"url_verification"}`)

	baseString := "v0:" + timestamp + ":" + string(body)
	hash := hmac.New(sha256.New, []byte(signingSecret))
	hash.Write([]byte(baseString))
	signature := "v0=" + hex.EncodeToString(hash.Sum(nil))

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("X-Slack-Signature", signature)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)

	respBody, err := verifier.VerifyRequest(req)
	if err != nil {
		t.Fatalf("Valid signature should not error: %v", err)
	}

	if string(respBody) != string(body) {
		t.Errorf("Response body mismatch")
	}
}

func TestEventVerifierInvalidSignature(t *testing.T) {
	verifier := NewEventVerifier("test-secret")

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	body := []byte(`{"type":"url_verification"}`)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("X-Slack-Signature", "v0=invalid")
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)

	_, err := verifier.VerifyRequest(req)
	if err == nil {
		t.Error("Invalid signature should return error")
	}
}

func TestEventVerifierOldTimestamp(t *testing.T) {
	verifier := NewEventVerifier("test-secret")

	// Use timestamp from 10 minutes ago
	oldTimestamp := strconv.FormatInt(time.Now().Unix()-600, 10)
	body := []byte(`{"type":"url_verification"}`)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("X-Slack-Signature", "v0=test")
	req.Header.Set("X-Slack-Request-Timestamp", oldTimestamp)

	_, err := verifier.VerifyRequest(req)
	if err == nil {
		t.Error("Old timestamp should return error")
	}
}

func TestNotificationEventToPipelineMessage(t *testing.T) {
	event := NewPipelineEvent("success", "pipe-123", "main", "abc123", "#test", "https://orcastack.example.com/pipes/123")

	text, blocks := event.ToMessage()

	if text != "Pipeline success" {
		t.Errorf("Expected text 'Pipeline success', got '%s'", text)
	}

	if len(blocks) < 3 {
		t.Errorf("Expected at least 3 blocks, got %d", len(blocks))
	}
}

func TestNotificationEventToDeploymentMessage(t *testing.T) {
	event := NewDeploymentEvent("failure", "production", "api-service", "v1.2.3", "#test", "https://orcastack.example.com")

	_, blocks := event.ToMessage()

	if event.Severity != SeverityCritical {
		t.Errorf("Expected severity critical, got %s", event.Severity)
	}

	if len(blocks) == 0 {
		t.Error("Expected blocks to be generated")
	}
}

func TestNotificationEventToSecurityMessage(t *testing.T) {
	event := NewSecurityEvent("violation", "RBAC", "user123", "Access denied", "#test", "https://orcastack.example.com")

	_, blocks := event.ToMessage()

	if len(blocks) == 0 {
	}

	if len(blocks) == 0 {
		t.Error("Expected blocks to be generated")
	}
}

func TestNotificationEventToHealthMessage(t *testing.T) {
	event := NewHealthEvent("critical", "memory", "usage_percent", "95.5", "#test", "https://orcastack.example.com")

	if event.EventType != EventHealthCritical {
		t.Errorf("Expected EventHealthCritical, got %s", event.EventType)
	}

	if event.Severity != SeverityCritical {
		t.Errorf("Expected severity critical, got %s", event.Severity)
	}
}

func TestNotificationEventToUserMessage(t *testing.T) {
	event := NewUserEvent("signup", "john.doe", "john@example.com", "New user", "#test", "https://orcastack.example.com")

	text, blocks := event.ToMessage()

	if !bytes.Contains([]byte(text), []byte("User Event")) {
		t.Errorf("Expected 'User Event' in title, got '%s'", text)
	}

	if len(blocks) == 0 {
		t.Error("Expected blocks to be generated")
	}
}

func TestSlackEventPayloadParsing(t *testing.T) {
	payload := &SlackEventPayload{
		Type:      "url_verification",
		Challenge: "test-challenge",
	}

	if !payload.IsUrlVerification() {
		t.Error("Should identify URL verification")
	}

	if payload.IsEventCallback() {
		t.Error("Should not be event callback")
	}
}

func TestSlackEventPayloadEventCallback(t *testing.T) {
	payload := &SlackEventPayload{
		Type: "event_callback",
		Event: map[string]interface{}{
			"type":    "message",
			"user":    "U123456",
			"channel": "C123456",
		},
	}

	if !payload.IsEventCallback() {
		t.Error("Should identify event callback")
	}

	if !payload.IsMessageEvent() {
		t.Error("Should identify message event")
	}

	if payload.GetUserID() != "U123456" {
		t.Errorf("Expected user ID U123456, got %s", payload.GetUserID())
	}

	if payload.GetChannelID() != "C123456" {
		t.Errorf("Expected channel ID C123456, got %s", payload.GetChannelID())
	}
}

func TestBlockBuilders(t *testing.T) {
	// Test section block
	section := NewSectionBlock("Test text")
	if section.Type != "section" {
		t.Errorf("Expected type 'section', got '%s'", section.Type)
	}

	// Test divider block
	divider := NewDividerBlock()
	if divider.Type != "divider" {
		t.Errorf("Expected type 'divider', got '%s'", divider.Type)
	}

	// Test header block
	header := NewHeaderBlock("Header")
	if header.Type != "header" {
		t.Errorf("Expected type 'header', got '%s'", header.Type)
	}
}

func TestAttachmentBuilder(t *testing.T) {
	attachment := NewAttachment("Title", string(ColorSuccess))
	attachment.AddField("Field1", "Value1", true)
	attachment.SetFooter("Footer")
	attachment.SetTitleLink("https://example.com")

	if attachment.Title != "Title" {
		t.Errorf("Expected title 'Title', got '%s'", attachment.Title)
	}

	if len(attachment.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(attachment.Fields))
	}

	if attachment.Footer != "Footer" {
		t.Errorf("Expected footer 'Footer', got '%s'", attachment.Footer)
	}

	if attachment.TitleLink != "https://example.com" {
		t.Errorf("Expected title link, got '%s'", attachment.TitleLink)
	}
}
