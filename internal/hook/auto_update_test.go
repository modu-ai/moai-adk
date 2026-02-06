package hook

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestAutoUpdateHandler_EventType(t *testing.T) {
	h := NewAutoUpdateHandler(nil)
	if got := h.EventType(); got != EventSessionStart {
		t.Errorf("EventType() = %v, want %v", got, EventSessionStart)
	}
}

func TestAutoUpdateHandler_Handle(t *testing.T) {
	tests := []struct {
		name              string
		fn                AutoUpdateFunc
		wantSystemMessage bool
		wantContains      string
	}{
		{
			name:              "nil function",
			fn:                nil,
			wantSystemMessage: false,
		},
		{
			name: "no update available",
			fn: func(ctx context.Context) (*AutoUpdateResult, error) {
				return &AutoUpdateResult{Updated: false}, nil
			},
			wantSystemMessage: false,
		},
		{
			name: "nil result",
			fn: func(ctx context.Context) (*AutoUpdateResult, error) {
				return nil, nil
			},
			wantSystemMessage: false,
		},
		{
			name: "update error swallowed",
			fn: func(ctx context.Context) (*AutoUpdateResult, error) {
				return nil, errors.New("network timeout")
			},
			wantSystemMessage: false,
		},
		{
			name: "successful update",
			fn: func(ctx context.Context) (*AutoUpdateResult, error) {
				return &AutoUpdateResult{
					Updated:         true,
					PreviousVersion: "v2.0.0",
					NewVersion:      "v2.0.1",
				}, nil
			},
			wantSystemMessage: true,
			wantContains:      "v2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAutoUpdateHandler(tt.fn)
			input := &HookInput{SessionID: "test-session"}

			output, err := h.Handle(context.Background(), input)
			if err != nil {
				t.Fatalf("Handle() returned error: %v", err)
			}

			if output == nil {
				t.Fatal("Handle() returned nil output")
			}

			hasMsg := output.SystemMessage != ""
			if hasMsg != tt.wantSystemMessage {
				t.Errorf("SystemMessage present = %v, want %v (msg: %q)",
					hasMsg, tt.wantSystemMessage, output.SystemMessage)
			}

			if tt.wantContains != "" && !strings.Contains(output.SystemMessage, tt.wantContains) {
				t.Errorf("SystemMessage %q should contain %q",
					output.SystemMessage, tt.wantContains)
			}
		})
	}
}

func TestAutoUpdateHandler_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Pre-cancel

	fn := func(ctx context.Context) (*AutoUpdateResult, error) {
		return nil, ctx.Err()
	}

	h := NewAutoUpdateHandler(fn)
	output, err := h.Handle(ctx, &HookInput{})
	if err != nil {
		t.Fatalf("Handle() should not propagate errors, got: %v", err)
	}
	if output.SystemMessage != "" {
		t.Errorf("cancelled context should not produce a SystemMessage, got: %q",
			output.SystemMessage)
	}
}
