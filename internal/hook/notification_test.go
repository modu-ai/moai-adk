package hook

import (
	"context"
	"testing"
)

func TestNotificationHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewNotificationHandler()

	if got := h.EventType(); got != EventNotification {
		t.Errorf("EventType() = %q, want %q", got, EventNotification)
	}
}

func TestNotificationHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "notification with all fields",
			input: &HookInput{
				SessionID:        "sess-notif-1",
				Title:            "Build Complete",
				Message:          "All tests passed",
				NotificationType: "success",
				HookEventName:    "Notification",
			},
		},
		{
			name: "notification without title",
			input: &HookInput{
				SessionID:     "sess-notif-2",
				Message:       "Info msg",
				HookEventName: "Notification",
			},
		},
		{
			name: "empty notification",
			input: &HookInput{
				SessionID: "sess-notif-3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewNotificationHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.HookSpecificOutput != nil {
				t.Error("Notification hook should not set hookSpecificOutput")
			}
		})
	}
}
