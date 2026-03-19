// Package integrations provides external service integration for MoAI-ADK,
// including Slack notifications, Linear issue sync, and webhook handling.
package integrations

// NotifyEvent represents an event that triggers external notifications.
type NotifyEvent struct {
	Type    string            `json:"type"`    // "spec_complete", "quality_failure", "pr_created", "budget_alert"
	SpecID  string            `json:"spec_id,omitempty"`
	Title   string            `json:"title"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Notifier defines the interface for sending notifications to external services.
type Notifier interface {
	// Send sends a notification event to the external service.
	Send(event NotifyEvent) error

	// Test sends a test notification to verify connectivity.
	Test() error

	// Name returns the notifier's display name.
	Name() string

	// IsEnabled returns whether the notifier is configured and enabled.
	IsEnabled() bool
}

// NotifyResult represents the result of a notification attempt.
type NotifyResult struct {
	Service string `json:"service"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
