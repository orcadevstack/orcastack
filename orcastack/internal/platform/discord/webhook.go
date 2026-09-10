package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// WebhookClient handles Discord webhook communications
type WebhookClient struct {
	webhookURL string
	httpClient *http.Client
	mu         sync.RWMutex
	enabled    bool
	maxRetries int
}

// Embed represents a Discord embed message
type Embed struct {
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	Color       int          `json:"color,omitempty"`
	Timestamp   string       `json:"timestamp,omitempty"`
	Fields      []EmbedField `json:"fields,omitempty"`
	URL         string       `json:"url,omitempty"`
}

// EmbedField represents a field in a Discord embed
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// WebhookPayload represents the structure sent to Discord
type WebhookPayload struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds,omitempty"`
}

// NewWebhookClient creates a new Discord webhook client
func NewWebhookClient(webhookURL string, enabled bool) *WebhookClient {
	return &WebhookClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		enabled:    enabled,
		maxRetries: 3,
	}
}

// IsEnabled returns whether Discord notifications are enabled
func (wc *WebhookClient) IsEnabled() bool {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	return wc.enabled && wc.webhookURL != ""
}

// SendAsync sends a webhook message asynchronously with retry logic
func (wc *WebhookClient) SendAsync(ctx context.Context, payload *WebhookPayload) error {
	if !wc.IsEnabled() {
		return nil
	}

	// Run in background goroutine to avoid blocking
	go func() {
		err := wc.send(ctx, payload)
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("[Discord] Failed to send webhook: %v\n", err)
		}
	}()

	return nil
}

// Send synchronously sends a webhook message (for testing)
func (wc *WebhookClient) Send(ctx context.Context, payload *WebhookPayload) error {
	if !wc.IsEnabled() {
		return fmt.Errorf("discord webhook not enabled")
	}

	return wc.send(ctx, payload)
}

// send handles the actual HTTP request with retry logic
func (wc *WebhookClient) send(ctx context.Context, payload *WebhookPayload) error {
	wc.mu.RLock()
	webhookURL := wc.webhookURL
	maxRetries := wc.maxRetries
	wc.mu.RUnlock()

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err = wc.sendRequest(ctx, webhookURL, jsonPayload)
		if err == nil {
			return nil
		}

		lastErr = err

		// Exponential backoff: 1s, 2s, 4s
		if attempt < maxRetries-1 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("failed to send webhook after %d attempts: %w", maxRetries, lastErr)
}

// sendRequest performs a single HTTP POST request to the Discord webhook
func (wc *WebhookClient) sendRequest(ctx context.Context, url string, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ORCASTACK/1.0")

	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	body, _ := io.ReadAll(resp.Body)

	// Discord returns 204 No Content on success
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ColorCode represents Discord message colors
type ColorCode int

const (
	ColorSuccess ColorCode = 0x00AA00 // Green
	ColorWarning ColorCode = 0xFFAA00 // Yellow
	ColorFailure ColorCode = 0xFF0000 // Red
	ColorInfo    ColorCode = 0x0099FF // Blue
)

// NewEmbed creates a new embed with common fields
func NewEmbed(title, description string, color ColorCode) *Embed {
	return &Embed{
		Title:       title,
		Description: description,
		Color:       int(color),
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}
}

// AddField adds a field to an embed
func (e *Embed) AddField(name, value string, inline bool) *Embed {
	e.Fields = append(e.Fields, EmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return e
}

// SetURL sets the URL for the embed
func (e *Embed) SetURL(url string) *Embed {
	e.URL = url
	return e
}
