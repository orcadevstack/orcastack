package discord

import (
	"context"
	"fmt"

	"github.com/orcastack/orcastackapi/internal/platform/config"
)

// Manager manages Discord webhook notifications globally
type Manager struct {
	client *WebhookClient
}

// NewManager creates a new Discord notification manager
func NewManager() *Manager {
	webhookURL := config.String("DISCORD_WEBHOOK_URL", "")
	enabled := config.Bool("DISCORD_NOTIFICATIONS_ENABLED", false)
	_ = config.String("DISCORD_ENVIRONMENT", "development") // For future use

	client := NewWebhookClient(webhookURL, enabled)

	return &Manager{
		client: client,
	}
}

// IsEnabled checks if Discord notifications are enabled
func (m *Manager) IsEnabled() bool {
	return m.client.IsEnabled()
}

// NotifyPipelineEvent sends a pipeline event notification
func (m *Manager) NotifyPipelineEvent(ctx context.Context, status, pipelineID, branch, commit, dashboardURL string) error {
	if !m.client.IsEnabled() {
		return nil
	}

	event := NewPipelineEvent(status, pipelineID, branch, commit, dashboardURL)
	payload := event.ToPayload()

	return m.client.SendAsync(ctx, payload)
}

// NotifyDeploymentEvent sends a deployment event notification
func (m *Manager) NotifyDeploymentEvent(ctx context.Context, status, env, service, version, dashboardURL string) error {
	if !m.client.IsEnabled() {
		return nil
	}

	event := NewDeploymentEvent(status, env, service, version, dashboardURL)
	payload := event.ToPayload()

	return m.client.SendAsync(ctx, payload)
}

// NotifyRepositoryEvent sends a repository event notification
func (m *Manager) NotifyRepositoryEvent(ctx context.Context, action, repoName, details, dashboardURL string) error {
	if !m.client.IsEnabled() {
		return nil
	}

	event := NewRepositoryEvent(action, repoName, details, dashboardURL)
	payload := event.ToPayload()

	return m.client.SendAsync(ctx, payload)
}

// NotifySecurityEvent sends a security event notification
func (m *Manager) NotifySecurityEvent(ctx context.Context, eventKind, violationType, username, details, dashboardURL string) error {
	if !m.client.IsEnabled() {
		return nil
	}

	event := NewSecurityEvent(eventKind, violationType, username, details, dashboardURL)
	payload := event.ToPayload()

	return m.client.SendAsync(ctx, payload)
}

// NotifyHealthEvent sends a health event notification
func (m *Manager) NotifyHealthEvent(ctx context.Context, status, component, metric, value, dashboardURL string) error {
	if !m.client.IsEnabled() {
		return nil
	}

	event := NewHealthEvent(status, component, metric, value, dashboardURL)
	payload := event.ToPayload()

	return m.client.SendAsync(ctx, payload)
}

// NotifyUserEvent sends a user event notification
func (m *Manager) NotifyUserEvent(ctx context.Context, action, username, email, details, dashboardURL string) error {
	if !m.client.IsEnabled() {
		return nil
	}

	event := NewUserEvent(action, username, email, details, dashboardURL)
	payload := event.ToPayload()

	return m.client.SendAsync(ctx, payload)
}

// NotifyCustomEvent sends a custom notification event
func (m *Manager) NotifyCustomEvent(ctx context.Context, event *NotificationEvent) error {
	if !m.client.IsEnabled() {
		return nil
	}

	payload := event.ToPayload()
	return m.client.SendAsync(ctx, payload)
}

// SetWebhookURL updates the webhook URL (mainly for testing)
func (m *Manager) SetWebhookURL(url string) {
	m.client.mu.Lock()
	defer m.client.mu.Unlock()
	m.client.webhookURL = url
}

// Global singleton manager
var globalManager *Manager

// Initialize sets up the global Discord notification manager
func Initialize() {
	globalManager = NewManager()
	if globalManager.IsEnabled() {
		fmt.Println("[Discord] Notifications enabled")
	}
}

// GetManager returns the global Discord manager
func GetManager() *Manager {
	if globalManager == nil {
		Initialize()
	}
	return globalManager
}
