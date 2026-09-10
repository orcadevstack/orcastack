package slack

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

// NotifierClient handles Slack message sending
type NotifierClient struct {
	botToken   string
	apiURL     string
	httpClient *http.Client
	mu         sync.RWMutex
	enabled    bool
	maxRetries int
}

// NewNotifierClient creates a new Slack notifier
func NewNotifierClient(botToken string, enabled bool) *NotifierClient {
	return &NotifierClient{
		botToken: botToken,
		apiURL:   "https://slack.com/api/chat.postMessage",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		enabled:    enabled,
		maxRetries: 3,
	}
}

// IsEnabled returns whether Slack notifications are enabled
func (nc *NotifierClient) IsEnabled() bool {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.enabled && nc.botToken != ""
}

// Block represents a Slack block in message formatting
type Block struct {
	Type      string                 `json:"type"`
	Text      *TextBlock             `json:"text,omitempty"`
	Elements  []interface{}          `json:"elements,omitempty"`
	Fields    []*TextBlock           `json:"fields,omitempty"`
	Accessory interface{}            `json:"accessory,omitempty"`
	Image     *ImageBlock            `json:"image_url,omitempty"`
	Label     *TextBlock             `json:"label,omitempty"`
	Optional  *bool                  `json:"optional,omitempty"`
	BlockID   string                 `json:"block_id,omitempty"`
	Confirm   map[string]interface{} `json:"confirm,omitempty"`
}

// TextBlock represents text in a Slack block
type TextBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Emoji    *bool  `json:"emoji,omitempty"`
	Verbatim *bool  `json:"verbatim,omitempty"`
}

// ImageBlock represents an image in a Slack block
type ImageBlock struct {
	URL     string `json:"url"`
	AltText string `json:"alt_text"`
}

// Attachment represents a legacy Slack attachment
type Attachment struct {
	Color      string            `json:"color,omitempty"`
	Title      string            `json:"title,omitempty"`
	TitleLink  string            `json:"title_link,omitempty"`
	Text       string            `json:"text,omitempty"`
	Fields     []AttachmentField `json:"fields,omitempty"`
	Footer     string            `json:"footer,omitempty"`
	FooterIcon string            `json:"footer_icon,omitempty"`
	Timestamp  int64             `json:"ts,omitempty"`
}

// AttachmentField represents a field in an attachment
type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Message represents a Slack message
type Message struct {
	Channel     string       `json:"channel"`
	ThreadTS    string       `json:"thread_ts,omitempty"`
	Text        string       `json:"text"`
	Blocks      []Block      `json:"blocks,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// SendAsync sends a message asynchronously with retry logic
func (nc *NotifierClient) SendAsync(ctx context.Context, channel, text string, blocks []Block) error {
	if !nc.IsEnabled() {
		return nil
	}

	go func() {
		msg := &Message{
			Channel: channel,
			Text:    text,
			Blocks:  blocks,
		}

		err := nc.send(ctx, msg)
		if err != nil {
			fmt.Printf("[Slack] Failed to send message: %v\n", err)
		}
	}()

	return nil
}

// Send synchronously sends a message (for testing)
func (nc *NotifierClient) Send(ctx context.Context, channel, text string, blocks []Block) error {
	if !nc.IsEnabled() {
		return fmt.Errorf("slack notifier not enabled")
	}

	msg := &Message{
		Channel: channel,
		Text:    text,
		Blocks:  blocks,
	}

	return nc.send(ctx, msg)
}

// send handles the actual HTTP request with retry logic
func (nc *NotifierClient) send(ctx context.Context, msg *Message) error {
	nc.mu.RLock()
	botToken := nc.botToken
	maxRetries := nc.maxRetries
	nc.mu.RUnlock()

	jsonPayload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err = nc.sendRequest(ctx, botToken, jsonPayload)
		if err == nil {
			return nil
		}

		lastErr = err

		// Exponential backoff
		if attempt < maxRetries-1 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("failed to send message after %d attempts: %w", maxRetries, lastErr)
}

// sendRequest performs a single HTTP POST request to Slack
func (nc *NotifierClient) sendRequest(ctx context.Context, botToken string, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, nc.apiURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", botToken))
	req.Header.Set("User-Agent", "ORCASTACK/1.0")

	resp, err := nc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var slackResp map[string]interface{}
	if err := json.Unmarshal(body, &slackResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if Slack reported an error
	if ok, exists := slackResp["ok"].(bool); exists && !ok {
		if errMsg, exists := slackResp["error"].(string); exists {
			return fmt.Errorf("slack error: %s", errMsg)
		}
	}

	return nil
}

// NewTextBlock creates a text block
func NewTextBlock(text string) *TextBlock {
	return &TextBlock{
		Type: "mrkdwn",
		Text: text,
	}
}

// NewSectionBlock creates a section block with text
func NewSectionBlock(text string) Block {
	return Block{
		Type: "section",
		Text: NewTextBlock(text),
	}
}

// NewFieldsBlock creates a fields block
func NewFieldsBlock(fields []*TextBlock) Block {
	return Block{
		Type:   "section",
		Fields: fields,
	}
}

// NewDividerBlock creates a divider block
func NewDividerBlock() Block {
	return Block{
		Type: "divider",
	}
}

// NewHeaderBlock creates a header block
func NewHeaderBlock(text string) Block {
	return Block{
		Type: "header",
		Text: &TextBlock{
			Type: "plain_text",
			Text: text,
		},
	}
}

// ColorCode represents Slack message colors
type ColorCode string

const (
	ColorSuccess ColorCode = "#36a64f" // Green
	ColorWarning ColorCode = "#ffa500" // Orange
	ColorFailure ColorCode = "#ff0000" // Red
	ColorInfo    ColorCode = "#439FE0" // Blue
)

// NewAttachment creates an attachment with color and fields
func NewAttachment(title, color string) Attachment {
	return Attachment{
		Color:     color,
		Title:     title,
		Timestamp: time.Now().Unix(),
	}
}

// AddField adds a field to an attachment
func (a *Attachment) AddField(title, value string, short bool) *Attachment {
	a.Fields = append(a.Fields, AttachmentField{
		Title: title,
		Value: value,
		Short: short,
	})
	return a
}

// SetFooter sets the footer for an attachment
func (a *Attachment) SetFooter(footer string) *Attachment {
	a.Footer = footer
	return a
}

// SetTitleLink sets the title link for an attachment
func (a *Attachment) SetTitleLink(link string) *Attachment {
	a.TitleLink = link
	return a
}
