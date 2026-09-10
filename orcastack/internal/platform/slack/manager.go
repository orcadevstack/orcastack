package slack

import (
	"context"
	"fmt"

	"github.com/orcastack/orcastackapi/internal/platform/config"
)

// Manager manages Slack notifications globally
type Manager struct {
	notifier *NotifierClient
	verifier *EventVerifier
	botToken string
	enabled  bool
}

// NewManager creates a new Slack notification manager
func NewManager() *Manager {
	botToken := config.String("SLACK_BOT_TOKEN", "")
	signingSecret := config.String("SLACK_SIGNING_SECRET", "")
	enabled := config.Bool("SLACK_NOTIFICATIONS_ENABLED", false)

	notifier := NewNotifierClient(botToken, enabled)
	verifier := NewEventVerifier(signingSecret)

	return &Manager{
		notifier: notifier,
		verifier: verifier,
		botToken: botToken,
		enabled:  enabled,
	}
}

// IsEnabled checks if Slack notifications are enabled
func (m *Manager) IsEnabled() bool {
	return m.enabled && m.notifier.IsEnabled()
}

// GetVerifier returns the event verifier
func (m *Manager) GetVerifier() *EventVerifier {
	return m.verifier
}

// NotifyPipelineEvent sends a pipeline event notification
func (m *Manager) NotifyPipelineEvent(ctx context.Context, status, pipelineID, branch, commit, channel, dashboardURL string) error {
	if !m.notifier.IsEnabled() {
		return nil
	}

	event := NewPipelineEvent(status, pipelineID, branch, commit, channel, dashboardURL)
	text, blocks := event.ToMessage()

	if channel == "" {
		channel = "#orcastack-pipelines"
	}

	return m.notifier.SendAsync(ctx, channel, text, blocks)
}

// NotifyDeploymentEvent sends a deployment event notification
func (m *Manager) NotifyDeploymentEvent(ctx context.Context, status, env, service, version, channel, dashboardURL string) error {
	if !m.notifier.IsEnabled() {
		return nil
	}

	event := NewDeploymentEvent(status, env, service, version, channel, dashboardURL)
	text, blocks := event.ToMessage()

	if channel == "" {
		channel = "#orcastack-deployments"
	}

	return m.notifier.SendAsync(ctx, channel, text, blocks)
}

// NotifyRepositoryEvent sends a repository event notification
func (m *Manager) NotifyRepositoryEvent(ctx context.Context, action, repoName, details, channel, dashboardURL string) error {
	if !m.notifier.IsEnabled() {
		return nil
	}

	event := NewRepositoryEvent(action, repoName, details, channel, dashboardURL)
	text, blocks := event.ToMessage()

	if channel == "" {
		channel = "#orcastack-repos"
	}

	return m.notifier.SendAsync(ctx, channel, text, blocks)
}

// NotifySecurityEvent sends a security event notification
func (m *Manager) NotifySecurityEvent(ctx context.Context, eventKind, violationType, username, details, channel, dashboardURL string) error {
	if !m.notifier.IsEnabled() {
		return nil
	}

	event := NewSecurityEvent(eventKind, violationType, username, details, channel, dashboardURL)
	text, blocks := event.ToMessage()

	if channel == "" {
		channel = "#orcastack-security"
	}

	return m.notifier.SendAsync(ctx, channel, text, blocks)
}

// NotifyHealthEvent sends a health event notification
func (m *Manager) NotifyHealthEvent(ctx context.Context, status, component, metric, value, channel, dashboardURL string) error {
	if !m.notifier.IsEnabled() {
		return nil
	}

	event := NewHealthEvent(status, component, metric, value, channel, dashboardURL)
	text, blocks := event.ToMessage()

	if channel == "" {
		channel = "#orcastack-health"
	}

	return m.notifier.SendAsync(ctx, channel, text, blocks)
}

// NotifyUserEvent sends a user event notification
func (m *Manager) NotifyUserEvent(ctx context.Context, action, username, email, details, channel, dashboardURL string) error {
	if !m.notifier.IsEnabled() {
		return nil
	}

	event := NewUserEvent(action, username, email, details, channel, dashboardURL)
	text, blocks := event.ToMessage()

	if channel == "" {
		channel = "#orcastack-users"
	}

	return m.notifier.SendAsync(ctx, channel, text, blocks)
}

// NotifyCustomEvent sends a custom notification event
func (m *Manager) NotifyCustomEvent(ctx context.Context, channel, text string, blocks []Block) error {
	if !m.notifier.IsEnabled() {
		return nil
	}

	if channel == "" {
		channel = "#orcastack-general"
	}

	return m.notifier.SendAsync(ctx, channel, text, blocks)
}

// SendDirectMessage sends a direct message to a user
func (m *Manager) SendDirectMessage(ctx context.Context, userID, text string, blocks []Block) error {
	if !m.notifier.IsEnabled() {
		return nil
	}

	// Direct messages use @userID format
	channel := fmt.Sprintf("@%s", userID)
	return m.notifier.SendAsync(ctx, channel, text, blocks)
}

// Global singleton manager
var globalManager *Manager

// Initialize sets up the global Slack notification manager
func Initialize() {
	globalManager = NewManager()
	if globalManager.IsEnabled() {
		fmt.Println("[Slack] Notifications enabled")
	}
}

// GetManager returns the global Slack manager
func GetManager() *Manager {
	if globalManager == nil {
		Initialize()
	}
	return globalManager
}
