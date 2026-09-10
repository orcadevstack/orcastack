package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// EventVerifier validates incoming Slack events using signature verification
type EventVerifier struct {
	signingSecret string
}

// NewEventVerifier creates a new event verifier
func NewEventVerifier(signingSecret string) *EventVerifier {
	return &EventVerifier{
		signingSecret: signingSecret,
	}
}

// VerifyRequest verifies the Slack request signature
// Returns the request body if valid, error otherwise
func (ev *EventVerifier) VerifyRequest(r *http.Request) ([]byte, error) {
	// Get signature components from headers
	signature := r.Header.Get("X-Slack-Signature")
	timestamp := r.Header.Get("X-Slack-Request-Timestamp")

	if signature == "" || timestamp == "" {
		return nil, fmt.Errorf("missing slack signature or timestamp")
	}

	// Verify timestamp is recent (within 5 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}

	now := time.Now().Unix()
	if now-ts > 300 { // 5 minutes
		return nil, fmt.Errorf("request timestamp too old")
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	// Verify signature
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
	hash := hmac.New(sha256.New, []byte(ev.signingSecret))
	hash.Write([]byte(baseString))
	computedSignature := "v0=" + hex.EncodeToString(hash.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(computedSignature)) {
		return nil, fmt.Errorf("invalid signature")
	}

	return body, nil
}

// SlackEventPayload represents the structure of a Slack event
type SlackEventPayload struct {
	Token       string                 `json:"token"`
	Challenge   string                 `json:"challenge"`
	Type        string                 `json:"type"`
	EventID     string                 `json:"event_id"`
	Event       map[string]interface{} `json:"event"`
	Actions     []map[string]interface{} `json:"actions,omitempty"`
	ResponseURL string                 `json:"response_url,omitempty"`
	TriggerID   string                 `json:"trigger_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	ChannelID   string                 `json:"channel_id,omitempty"`
	TeamID      string                 `json:"team_id,omitempty"`
	EnterpriseID string                 `json:"enterprise_id,omitempty"`
	Command     string                 `json:"command,omitempty"`
	Text        string                 `json:"text,omitempty"`
}

// IsUrlVerification checks if this is a URL verification challenge from Slack
func (ep *SlackEventPayload) IsUrlVerification() bool {
	return ep.Type == "url_verification"
}

// IsEventCallback checks if this is an event callback
func (ep *SlackEventPayload) IsEventCallback() bool {
	return ep.Type == "event_callback"
}

// IsSlashCommand checks if this is a slash command
func (ep *SlackEventPayload) IsSlashCommand() bool {
	return !strings.HasPrefix(ep.Command, "") && ep.Command != "" // If Command field is present
}

// GetEventType returns the event type from the event
func (ep *SlackEventPayload) GetEventType() string {
	if ep.Event == nil {
		return ""
	}
	if eventType, ok := ep.Event["type"].(string); ok {
		return eventType
	}
	return ""
}

// IsMessageEvent checks if event is a message event
func (ep *SlackEventPayload) IsMessageEvent() bool {
	return ep.GetEventType() == "message"
}

// IsAppMentionEvent checks if event is an app mention
func (ep *SlackEventPayload) IsAppMentionEvent() bool {
	return ep.GetEventType() == "app_mention"
}

// IsReactionAddedEvent checks if event is a reaction added
func (ep *SlackEventPayload) IsReactionAddedEvent() bool {
	return ep.GetEventType() == "reaction_added"
}

// GetUserID extracts user ID from event
func (ep *SlackEventPayload) GetUserID() string {
	if ep.UserID != "" {
		return ep.UserID
	}

	if ep.Event != nil {
		if userID, ok := ep.Event["user"].(string); ok {
			return userID
		}
	}

	return ""
}

// GetChannelID extracts channel ID from event
func (ep *SlackEventPayload) GetChannelID() string {
	if ep.ChannelID != "" {
		return ep.ChannelID
	}

	if ep.Event != nil {
		if channelID, ok := ep.Event["channel"].(string); ok {
			return channelID
		}
	}

	return ""
}

// GetText extracts text from event or command
func (ep *SlackEventPayload) GetText() string {
	if ep.Text != "" {
		return ep.Text
	}

	if ep.Event != nil {
		if text, ok := ep.Event["text"].(string); ok {
			return text
		}
	}

	return ""
}
