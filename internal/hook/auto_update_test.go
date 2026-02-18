package hook

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
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

// TestAutoUpdateHandler_IndependentContext verifies that Handle uses an
// independent context for the update function, not the caller's context.
// This is critical for issue #397: when the registry's outer context
// expires, the handler must still return cleanly rather than propagating
// a timeout error that causes "SessionStart:startup hook error".
func TestAutoUpdateHandler_IndependentContext(t *testing.T) {
	// The update function checks whether its context is already cancelled
	// at the moment the function is invoked. If Handle forwards the caller's
	// (pre-cancelled) context, the function will see ctx.Err() != nil
	// immediately upon entry. If Handle creates an independent context from
	// context.Background(), the function will see ctx.Err() == nil.
	ctxErrAtInvocation := make(chan error, 1)
	fn := func(ctx context.Context) (*AutoUpdateResult, error) {
		ctxErrAtInvocation <- ctx.Err()
		return &AutoUpdateResult{Updated: false}, nil
	}

	// Pre-cancel the caller context to simulate a timed-out registry context.
	callerCtx, cancel := context.WithCancel(context.Background())
	cancel()

	h := NewAutoUpdateHandler(fn)
	output, err := h.Handle(callerCtx, &HookInput{})
	if err != nil {
		t.Fatalf("Handle() should not propagate errors, got: %v", err)
	}
	if output == nil {
		t.Fatal("Handle() returned nil output")
	}

	// The update function must have been called with a non-cancelled context,
	// proving Handle creates an independent context rather than forwarding
	// the (expired) caller context.
	select {
	case ctxErr := <-ctxErrAtInvocation:
		if ctxErr != nil {
			t.Errorf("update function received a cancelled context at invocation (%v), "+
				"want independent non-cancelled context", ctxErr)
		}
	default:
		t.Fatal("update function was never called")
	}
}

// TestAutoUpdateHandler_UpdateContextHasDeadline verifies that the update
// function receives a context with a deadline set by autoUpdateTimeout.
func TestAutoUpdateHandler_UpdateContextHasDeadline(t *testing.T) {
	deadlineSet := make(chan bool, 1)
	fn := func(ctx context.Context) (*AutoUpdateResult, error) {
		_, hasDeadline := ctx.Deadline()
		deadlineSet <- hasDeadline
		return &AutoUpdateResult{Updated: false}, nil
	}

	h := NewAutoUpdateHandler(fn)
	_, err := h.Handle(context.Background(), &HookInput{})
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if !<-deadlineSet {
		t.Error("update function context has no deadline, want autoUpdateTimeout deadline")
	}
}

// TestAutoUpdateTimeout_Value documents and asserts the expected timeout
// constant so that accidental changes are caught during code review.
func TestAutoUpdateTimeout_Value(t *testing.T) {
	if autoUpdateTimeout != 25*time.Second {
		t.Errorf("autoUpdateTimeout = %v, want 25s; this constant must stay below the SessionStart hook timeout (30s)",
			autoUpdateTimeout)
	}
}
