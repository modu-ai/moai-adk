package hook

import (
	"context"
	"log/slog"
)

// notificationHandler processes Notification events.
// It logs notifications sent by Claude Code.
type notificationHandler struct{}

// NewNotificationHandler creates a new Notification event handler.
func NewNotificationHandler() Handler {
	return &notificationHandler{}
}

// EventType returns EventNotification.
func (h *notificationHandler) EventType() EventType {
	return EventNotification
}

// Handle processes a Notification event. It logs the notification details.
func (h *notificationHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("notification received",
		"session_id", input.SessionID,
		"title", input.Title,
		"message", input.Message,
		"type", input.NotificationType,
	)
	return &HookOutput{}, nil
}
