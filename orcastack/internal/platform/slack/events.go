package slack

import (
	"fmt"
	"time"
)

// SlackEvent represents different notification categories for Slack
type SlackEvent string

const (
	// Pipeline Events
	EventPipelineStart   SlackEvent = "pipeline.start"
	EventPipelineSuccess SlackEvent = "pipeline.success"
	EventPipelineFailure SlackEvent = "pipeline.failure"

	// Deployment Events
	EventDeploymentStart    SlackEvent = "deployment.start"
	EventDeploymentSuccess  SlackEvent = "deployment.success"
	EventDeploymentFailure  SlackEvent = "deployment.failure"
	EventDeploymentRollback SlackEvent = "deployment.rollback"

	// Repository Events
	EventRepoCreated    SlackEvent = "repo.created"
	EventBranchCreated  SlackEvent = "branch.created"
	EventMergeCompleted SlackEvent = "merge.completed"
	EventCommitPushed   SlackEvent = "commit.pushed"

	// Security Events
	EventSecurityViolation  SlackEvent = "security.violation"
	EventAuthenticationFail SlackEvent = "auth.failure"
	EventAccessDenied       SlackEvent = "access.denied"

	// System Health Events
	EventHealthWarning   SlackEvent = "health.warning"
	EventHealthCritical  SlackEvent = "health.critical"
	EventContainerRestart SlackEvent = "container.restart"

	// Identity Events
	EventUserSignup      SlackEvent = "user.signup"
	EventAccessRequest   SlackEvent = "access.request"
	EventAccessApproved  SlackEvent = "access.approved"
	EventAccessDeniedEvt SlackEvent = "access.denied"
)

// Severity defines the severity level of an event
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// NotificationEvent contains all information needed for a Slack notification
type NotificationEvent struct {
	EventType    SlackEvent
	Title        string
	Description  string
	Details      map[string]string
	Severity     Severity
	DashboardURL string
	Channel      string
	Timestamp    time.Time
}

// ToMessage converts a NotificationEvent to Slack blocks
func (ne *NotificationEvent) ToMessage() (string, []Block) {
	_ = ne.getColor() // Color code for future rich formatting
	
	// Build blocks
	blocks := []Block{
		NewHeaderBlock(ne.Title),
		NewSectionBlock(ne.Description),
		NewDividerBlock(),
	}

	// Add details as fields
	if len(ne.Details) > 0 {
		detailText := ""
		for key, value := range ne.Details {
			detailText += fmt.Sprintf("*%s:* %s\n", key, value)
		}
		blocks = append(blocks, NewSectionBlock(detailText))
	}

	// Add metadata
	metaText := fmt.Sprintf("*Event:* %s | *Severity:* %s | _Updated: %s_",
		ne.EventType, ne.Severity, ne.Timestamp.Format(time.RFC3339))
	blocks = append(blocks, NewSectionBlock(metaText))

	// Add dashboard link
	if ne.DashboardURL != "" {
		linkText := fmt.Sprintf("<@orcastack> <here|View in Dashboard>: %s", ne.DashboardURL)
		blocks = append(blocks, NewSectionBlock(linkText))
	}

	return ne.Title, blocks
}

// getColor returns the appropriate Slack color based on severity
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
func NewPipelineEvent(status, pipelineID, branch, commit, channel, dashboardURL string) *NotificationEvent {
	var eventType SlackEvent
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
		eventType = SlackEvent(status)
		severity = SeverityInfo
	}

	return &NotificationEvent{
		EventType:    eventType,
		Title:        fmt.Sprintf("Pipeline %s", status),
		Description:  fmt.Sprintf("Pipeline execution completed with status: `%s`", status),
		Severity:     severity,
		Channel:      channel,
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
func NewDeploymentEvent(status, env, service, version, channel, dashboardURL string) *NotificationEvent {
	var eventType SlackEvent
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
		eventType = SlackEvent(status)
		severity = SeverityInfo
	}

	return &NotificationEvent{
		EventType:    eventType,
		Title:        fmt.Sprintf("Deployment %s", status),
		Description:  fmt.Sprintf("Service deployment completed with status: `%s`", status),
		Severity:     severity,
		Channel:      channel,
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
func NewRepositoryEvent(action, repoName, details, channel, dashboardURL string) *NotificationEvent {
	var eventType SlackEvent

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
		eventType = SlackEvent(action)
	}

	return &NotificationEvent{
		EventType:    eventType,
		Title:        fmt.Sprintf("Repository Event: %s", action),
		Description:  fmt.Sprintf("Repository action: `%s`", action),
		Severity:     SeverityInfo,
		Channel:      channel,
		DashboardURL: dashboardURL,
		Timestamp:    time.Now(),
		Details: map[string]string{
			"Repository": repoName,
			"Details":    details,
		},
	}
}

// NewSecurityEvent creates a security notification event
func NewSecurityEvent(eventKind, violationType, username, details, channel, dashboardURL string) *NotificationEvent {
	var eventType SlackEvent
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
		eventType = SlackEvent(eventKind)
		severity = SeverityCritical
	}

	return &NotificationEvent{
		EventType:    eventType,
		Title:        fmt.Sprintf("🚨 Security Alert: %s", violationType),
		Description:  fmt.Sprintf("Security event detected: `%s`", violationType),
		Severity:     severity,
		Channel:      channel,
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
func NewHealthEvent(status, component, metric, value, channel, dashboardURL string) *NotificationEvent {
	var eventType SlackEvent
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
		eventType = SlackEvent(status)
		severity = SeverityInfo
	}

	return &NotificationEvent{
		EventType:    eventType,
		Title:        fmt.Sprintf("📊 System Health: %s", status),
		Description:  fmt.Sprintf("Health check alert: `%s`", component),
		Severity:     severity,
		Channel:      channel,
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
func NewUserEvent(action, username, email, details, channel, dashboardURL string) *NotificationEvent {
	var eventType SlackEvent

	switch action {
	case "signup":
		eventType = EventUserSignup
	case "access_request":
		eventType = EventAccessRequest
	case "access_approved":
		eventType = EventAccessApproved
	default:
		eventType = SlackEvent(action)
	}

	return &NotificationEvent{
		EventType:    eventType,
		Title:        fmt.Sprintf("👤 User Event: %s", action),
		Description:  fmt.Sprintf("User action: `%s`", action),
		Severity:     SeverityInfo,
		Channel:      channel,
		DashboardURL: dashboardURL,
		Timestamp:    time.Now(),
		Details: map[string]string{
			"Username": username,
			"Email":    email,
			"Details":  details,
		},
	}
}
