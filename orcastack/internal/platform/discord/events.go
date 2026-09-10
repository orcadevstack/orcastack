package discord

import (
	"fmt"
	"time"
)

// EventType represents different notification categories
type EventType string

const (
	// Pipeline Events
	EventPipelineStart   EventType = "pipeline.start"
	EventPipelineSuccess EventType = "pipeline.success"
	EventPipelineFailure EventType = "pipeline.failure"

	// Deployment Events
	EventDeploymentStart    EventType = "deployment.start"
	EventDeploymentSuccess  EventType = "deployment.success"
	EventDeploymentFailure  EventType = "deployment.failure"
	EventDeploymentRollback EventType = "deployment.rollback"

	// Repository Events
	EventRepoCreated    EventType = "repo.created"
	EventBranchCreated  EventType = "branch.created"
	EventMergeCompleted EventType = "merge.completed"
	EventCommitPushed   EventType = "commit.pushed"

	// Security Events
	EventSecurityViolation  EventType = "security.violation"
	EventAuthenticationFail EventType = "auth.failure"
	EventAccessDenied       EventType = "access.denied"

	// System Health Events
	EventHealthWarning   EventType = "health.warning"
	EventHealthCritical  EventType = "health.critical"
	EventContainerRestart EventType = "container.restart"

	// User Events
	EventUserSignup      EventType = "user.signup"
	EventAccessRequest   EventType = "access.request"
	EventAccessApproved  EventType = "access.approved"
	EventAccessDeniedEvt EventType = "access.denied"
)

// NotificationEvent contains all information needed for a Discord notification
type NotificationEvent struct {
	EventType   EventType
	Title       string
	Description string
	Details     map[string]string
	Severity    Severity
	DashboardURL string
	Timestamp   time.Time
}

// Severity defines the severity level of an event
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// ToPayload converts a NotificationEvent to a WebhookPayload for Discord
func (ne *NotificationEvent) ToPayload() *WebhookPayload {
	color := ne.getColor()
	embed := NewEmbed(ne.Title, ne.Description, color)

	// Add details as fields
	for key, value := range ne.Details {
		embed.AddField(key, value, false)
	}

	// Add severity field
	embed.AddField("Severity", string(ne.Severity), true)

	// Add event type
	embed.AddField("Event Type", string(ne.EventType), true)

	// Set dashboard link if provided
	if ne.DashboardURL != "" {
		embed.SetURL(ne.DashboardURL)
		embed.AddField("Dashboard", fmt.Sprintf("[View Details](%s)", ne.DashboardURL), false)
	}

	return &WebhookPayload{
		Embeds: []Embed{*embed},
	}
}

// getColor returns the appropriate Discord color based on severity
func (ne *NotificationEvent) getColor() ColorCode {
	switch ne.Severity {
	case SeverityCritical:
		return ColorFailure
	case SeverityWarning:
		return ColorWarning
	default:
		return ColorInfo
	}
}

// NewPipelineEvent creates a pipeline notification event
func NewPipelineEvent(status string, pipelineID, branch, commit string, dashboardURL string) *NotificationEvent {
	var eventType EventType
	var severity Severity

	switch status {
	case "success":
		eventType = EventPipelineSuccess
		severity = SeverityInfo
	case "failure":
		eventType = EventPipelineFailure
		severity = SeverityCritical
	case "start":
		eventType = EventPipelineStart
		severity = SeverityInfo
	default:
		eventType = EventType(status)
		severity = SeverityInfo
	}

	return &NotificationEvent{
		EventType:    eventType,
		Title:        fmt.Sprintf("Pipeline %s", status),
		Description:  fmt.Sprintf("Pipeline execution completed with status: %s", status),
		Severity:     severity,
		DashboardURL: dashboardURL,
		Timestamp:    time.Now(),
		Details: map[string]string{
			"Pipeline ID": pipelineID,
			"Branch":      branch,
			"Commit":      commit,
		},
	}
}

// NewDeploymentEvent creates a deployment notification event
func NewDeploymentEvent(status, env, service, version string, dashboardURL string) *NotificationEvent {
	var eventType EventType
	var severity Severity

	switch status {
	case "success":
		eventType = EventDeploymentSuccess
		severity = SeverityInfo
	case "failure":
		eventType = EventDeploymentFailure
		severity = SeverityCritical
	case "rollback":
		eventType = EventDeploymentRollback
		severity = SeverityWarning
	default:
		eventType = EventType(status)
		severity = SeverityInfo
	}

	return &NotificationEvent{
		EventType:    eventType,
		Title:        fmt.Sprintf("Deployment %s", status),
		Description:  fmt.Sprintf("Service deployment completed with status: %s", status),
		Severity:     severity,
		DashboardURL: dashboardURL,
		Timestamp:    time.Now(),
		Details: map[string]string{
			"Environment": env,
			"Service":     service,
			"Version":     version,
		},
	}
}

// NewRepositoryEvent creates a repository notification event
func NewRepositoryEvent(action, repoName, details string, dashboardURL string) *NotificationEvent {
	var eventType EventType

	switch action {
	case "created":
		eventType = EventRepoCreated
	case "branch_created":
		eventType = EventBranchCreated
	case "merged":
		eventType = EventMergeCompleted
	case "commit":
		eventType = EventCommitPushed
	default:
		eventType = EventType(action)
	}

	return &NotificationEvent{
		EventType:    eventType,
		Title:        fmt.Sprintf("Repository Event: %s", action),
		Description:  fmt.Sprintf("Repository action: %s", action),
		Severity:     SeverityInfo,
		DashboardURL: dashboardURL,
		Timestamp:    time.Now(),
		Details: map[string]string{
			"Repository": repoName,
			"Details":    details,
		},
	}
}

// NewSecurityEvent creates a security notification event
func NewSecurityEvent(eventKind, violationType, username, details string, dashboardURL string) *NotificationEvent {
	var eventType EventType
	var severity Severity

	switch eventKind {
	case "violation":
		eventType = EventSecurityViolation
		severity = SeverityCritical
	case "auth_failure":
		eventType = EventAuthenticationFail
		severity = SeverityWarning
	case "access_denied":
		eventType = EventAccessDenied
		severity = SeverityWarning
	default:
		eventType = EventType(eventKind)
		severity = SeverityCritical
	}

	return &NotificationEvent{
		EventType:    eventType,
		Title:        fmt.Sprintf("Security Alert: %s", violationType),
		Description:  fmt.Sprintf("Security event detected: %s", violationType),
		Severity:     severity,
		DashboardURL: dashboardURL,
		Timestamp:    time.Now(),
		Details: map[string]string{
			"Type":     violationType,
			"Username": username,
			"Details":  details,
		},
	}
}

// NewHealthEvent creates a system health notification event
func NewHealthEvent(status, component, metric, value string, dashboardURL string) *NotificationEvent {
	var eventType EventType
	var severity Severity

	switch status {
	case "warning":
		eventType = EventHealthWarning
		severity = SeverityWarning
	case "critical":
		eventType = EventHealthCritical
		severity = SeverityCritical
	case "restart":
		eventType = EventContainerRestart
		severity = SeverityWarning
	default:
		eventType = EventType(status)
		severity = SeverityInfo
	}

	return &NotificationEvent{
		EventType:    eventType,
		Title:        fmt.Sprintf("System Health: %s", status),
		Description:  fmt.Sprintf("Health check alert: %s", component),
		Severity:     severity,
		DashboardURL: dashboardURL,
		Timestamp:    time.Now(),
		Details: map[string]string{
			"Component": component,
			"Metric":    metric,
			"Value":     value,
		},
	}
}

// NewUserEvent creates a user-related notification event
func NewUserEvent(action, username, email, details string, dashboardURL string) *NotificationEvent {
	var eventType EventType

	switch action {
	case "signup":
		eventType = EventUserSignup
	case "access_request":
		eventType = EventAccessRequest
	case "access_approved":
		eventType = EventAccessApproved
	default:
		eventType = EventType(action)
	}

	return &NotificationEvent{
		EventType:    eventType,
		Title:        fmt.Sprintf("User Event: %s", action),
		Description:  fmt.Sprintf("User action: %s", action),
		Severity:     SeverityInfo,
		DashboardURL: dashboardURL,
		Timestamp:    time.Now(),
		Details: map[string]string{
			"Username": username,
			"Email":    email,
			"Details":  details,
		},
	}
}
