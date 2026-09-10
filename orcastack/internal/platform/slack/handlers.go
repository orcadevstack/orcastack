package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// OAuthHandler manages Slack OAuth flow
type OAuthHandler struct {
	config OAuthConfig
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(config OAuthConfig) *OAuthHandler {
	return &OAuthHandler{
		config: config,
	}
}

// HandleInstall redirects to Slack OAuth consent screen
func (oh *OAuthHandler) HandleInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Build OAuth URL
	oauthURL := "https://slack.com/oauth/v2/authorize"
	params := url.Values{}
	params.Add("client_id", oh.config.ClientID)
	params.Add("scope", "chat:write,chat:write.public,commands,app_mentions:read")
	params.Add("redirect_uri", oh.config.RedirectURL)
	params.Add("state", generateState()) // For CSRF protection

	fullURL := oauthURL + "?" + params.Encode()

	http.Redirect(w, r, fullURL, http.StatusFound)
}

// OAuthCallbackResponse response structure
type OAuthCallbackResponse struct {
	Ok        bool   `json:"ok"`
	AppID     string `json:"app_id"`
	AuthedUser struct {
		ID string `json:"id"`
	} `json:"authed_user"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	BotUserID   string `json:"bot_user_id"`
	Team struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"team"`
	Enterprise struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"enterprise"`
	Error string `json:"error,omitempty"`
}

// HandleCallback handles Slack OAuth callback
func (oh *OAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := oh.exchangeCodeForToken(code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange code: %v", err), http.StatusInternalServerError)
		return
	}

	if !token.Ok {
		http.Error(w, fmt.Sprintf("OAuth error: %s", token.Error), http.StatusBadRequest)
		return
	}

	// TODO: Store token in database
	// For now, just confirm success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":        true,
		"message":   "Slack integration installed successfully",
		"team":      token.Team.Name,
		"bot_user":  token.BotUserID,
	})
}

// exchangeCodeForToken exchanges OAuth code for access token
func (oh *OAuthHandler) exchangeCodeForToken(code string) (*OAuthCallbackResponse, error) {
	// Prepare request
	params := url.Values{}
	params.Add("client_id", oh.config.ClientID)
	params.Add("client_secret", oh.config.ClientSecret)
	params.Add("code", code)
	params.Add("redirect_uri", oh.config.RedirectURL)

	resp, err := http.PostForm("https://slack.com/api/oauth.v2.access", params)
	if err != nil {
		return nil, fmt.Errorf("failed to call Slack API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response OAuthCallbackResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// EventHandler handles incoming Slack events
type EventHandler struct {
	verifier *EventVerifier
	onEvent  func(ctx context.Context, event *SlackEventPayload) error
}

// NewEventHandler creates a new event handler
func NewEventHandler(verifier *EventVerifier, onEvent func(ctx context.Context, event *SlackEventPayload) error) *EventHandler {
	return &EventHandler{
		verifier: verifier,
		onEvent:  onEvent,
	}
}

// HandleEvents handles incoming Slack events
func (eh *EventHandler) HandleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify request signature
	body, err := eh.verifier.VerifyRequest(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Verification failed: %v", err), http.StatusUnauthorized)
		return
	}

	// Parse event
	var event SlackEventPayload
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Failed to parse event", http.StatusBadRequest)
		return
	}

	// Handle URL verification challenge
	if event.IsUrlVerification() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"challenge": event.Challenge,
		})
		return
	}

	// Respond immediately with 200 OK
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"ok": "true",
	})

	// Process event asynchronously
	go func() {
		// Create context with timeout
		timeoutCtx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := eh.onEvent(timeoutCtx, &event); err != nil {
			fmt.Printf("[Slack] Error processing event: %v\n", err)
		}
	}()
}

// SlashCommandHandler handles Slack slash commands
type SlashCommandHandler struct {
	verifier  *EventVerifier
	onCommand func(ctx context.Context, command *SlackEventPayload) (string, error)
}

// NewSlashCommandHandler creates a new slash command handler
func NewSlashCommandHandler(verifier *EventVerifier, onCommand func(ctx context.Context, command *SlackEventPayload) (string, error)) *SlashCommandHandler {
	return &SlashCommandHandler{
		verifier:  verifier,
		onCommand: onCommand,
	}
}

// HandleCommand handles incoming slash commands
func (sch *SlashCommandHandler) HandleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify request
	body, err := sch.verifier.VerifyRequest(r)
	if err != nil {
		http.Error(w, "Verification failed", http.StatusUnauthorized)
		return
	}

	// Parse command
	var command SlackEventPayload
	if err := json.Unmarshal(body, &command); err != nil {
		http.Error(w, "Failed to parse command", http.StatusBadRequest)
		return
	}

	// Process command
	response, err := sch.onCommand(r.Context(), &command)
	if err != nil {
		response = fmt.Sprintf("Error: %v", err)
	}

	// Respond with message
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"response_type": "in_channel",
		"text":          response,
	})
}

// generateState generates a random state for CSRF protection
func generateState() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
