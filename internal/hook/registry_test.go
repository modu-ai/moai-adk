package hook

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk-go/internal/config"
)

// mockHandler is a test helper that implements the Handler interface.
type mockHandler struct {
	event  EventType
	output *HookOutput
	err    error
	called bool
}

func (h *mockHandler) EventType() EventType { return h.event }

func (h *mockHandler) Handle(_ context.Context, _ *HookInput) (*HookOutput, error) {
	h.called = true
	return h.output, h.err
}

// slowHandler simulates a handler that takes a long time to execute.
type slowHandler struct {
	event    EventType
	duration time.Duration
}

func (h *slowHandler) EventType() EventType { return h.event }

func (h *slowHandler) Handle(ctx context.Context, _ *HookInput) (*HookOutput, error) {
	select {
	case <-time.After(h.duration):
		return NewAllowOutput(), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// mockConfigProvider implements ConfigProvider for testing.
type mockConfigProvider struct {
	cfg *config.Config
}

func (m *mockConfigProvider) Get() *config.Config { return m.cfg }

func newTestConfig() *config.Config {
	return config.NewDefaultConfig()
}

func TestRegistryRegisterAndHandlers(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	reg := NewRegistry(cfg)

	h1 := &mockHandler{event: EventSessionStart, output: NewAllowOutput()}
	h2 := &mockHandler{event: EventSessionStart, output: NewAllowOutput()}
	h3 := &mockHandler{event: EventPreToolUse, output: NewAllowOutput()}

	reg.Register(h1)
	reg.Register(h2)
	reg.Register(h3)

	// SessionStart should have 2 handlers
	handlers := reg.Handlers(EventSessionStart)
	if len(handlers) != 2 {
		t.Errorf("SessionStart handlers = %d, want 2", len(handlers))
	}

	// PreToolUse should have 1 handler
	handlers = reg.Handlers(EventPreToolUse)
	if len(handlers) != 1 {
		t.Errorf("PreToolUse handlers = %d, want 1", len(handlers))
	}

	// PostToolUse should have 0 handlers
	handlers = reg.Handlers(EventPostToolUse)
	if len(handlers) != 0 {
		t.Errorf("PostToolUse handlers = %d, want 0", len(handlers))
	}
}

func TestRegistryDispatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		handlers     []*mockHandler
		event        EventType
		wantDecision string
		wantErr      bool
		errTarget    error
		checkCalled  func(t *testing.T, handlers []*mockHandler)
	}{
		{
			name:         "empty registry returns allow for SessionStart",
			handlers:     nil,
			event:        EventSessionStart,
			wantDecision: DecisionAllow,
		},
		{
			name:         "empty registry returns empty for SessionEnd",
			handlers:     nil,
			event:        EventSessionEnd,
			wantDecision: "", // SessionEnd returns empty JSON per Claude Code protocol
		},
		{
			name:         "empty registry returns empty for Stop",
			handlers:     nil,
			event:        EventStop,
			wantDecision: "", // Stop returns empty JSON per Claude Code protocol
		},
		{
			name: "single allow handler",
			handlers: []*mockHandler{
				{event: EventSessionStart, output: NewAllowOutput()},
			},
			event:        EventSessionStart,
			wantDecision: DecisionAllow,
		},
		{
			name: "multiple allow handlers all execute",
			handlers: []*mockHandler{
				{event: EventPostToolUse, output: NewAllowOutput()},
				{event: EventPostToolUse, output: NewAllowOutput()},
				{event: EventPostToolUse, output: NewAllowOutput()},
			},
			event:        EventPostToolUse,
			wantDecision: DecisionAllow,
			checkCalled: func(t *testing.T, handlers []*mockHandler) {
				t.Helper()
				for i, h := range handlers {
					if !h.called {
						t.Errorf("handler[%d] was not called", i)
					}
				}
			},
		},
		{
			name: "block short-circuits remaining handlers",
			handlers: []*mockHandler{
				{event: EventPreToolUse, output: NewAllowOutput()},
				{event: EventPreToolUse, output: NewBlockOutput("blocked")},
				{event: EventPreToolUse, output: NewAllowOutput()},
			},
			event:        EventPreToolUse,
			wantDecision: DecisionBlock,
			checkCalled: func(t *testing.T, handlers []*mockHandler) {
				t.Helper()
				if !handlers[0].called {
					t.Error("handler[0] should have been called")
				}
				if !handlers[1].called {
					t.Error("handler[1] should have been called")
				}
				if handlers[2].called {
					t.Error("handler[2] should NOT have been called after block")
				}
			},
		},
		{
			name: "handler error stops dispatch chain",
			handlers: []*mockHandler{
				{event: EventSessionStart, output: NewAllowOutput()},
				{event: EventSessionStart, err: errors.New("handler failed")},
				{event: EventSessionStart, output: NewAllowOutput()},
			},
			event:   EventSessionStart,
			wantErr: true,
			checkCalled: func(t *testing.T, handlers []*mockHandler) {
				t.Helper()
				if !handlers[0].called {
					t.Error("handler[0] should have been called")
				}
				if !handlers[1].called {
					t.Error("handler[1] should have been called")
				}
				if handlers[2].called {
					t.Error("handler[2] should NOT have been called after error")
				}
			},
		},
		{
			name: "dispatch to unregistered event returns allow for PreToolUse",
			handlers: []*mockHandler{
				{event: EventSessionStart, output: NewAllowOutput()},
			},
			event:        EventPreToolUse,
			wantDecision: DecisionAllow,
		},
		{
			name: "dispatch to unregistered SessionEnd returns empty output",
			handlers: []*mockHandler{
				{event: EventSessionStart, output: NewAllowOutput()},
			},
			event:        EventSessionEnd,
			wantDecision: "", // SessionEnd should return empty JSON per Claude Code protocol
		},
		{
			name: "dispatch to unregistered Stop returns empty output",
			handlers: []*mockHandler{
				{event: EventSessionStart, output: NewAllowOutput()},
			},
			event:        EventStop,
			wantDecision: "", // Stop should return empty JSON per Claude Code protocol
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfigProvider{cfg: newTestConfig()}
			reg := NewRegistry(cfg)

			for _, h := range tt.handlers {
				reg.Register(h)
			}

			ctx := context.Background()
			input := &HookInput{
				SessionID:     "test-session",
				CWD:           "/tmp",
				HookEventName: string(tt.event),
			}

			got, err := reg.Dispatch(ctx, tt.event, input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errTarget != nil && !errors.Is(err, tt.errTarget) {
					t.Errorf("error = %v, want errors.Is(%v)", err, tt.errTarget)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got == nil {
					t.Fatal("got nil output, want non-nil")
				}
				if got.Decision != tt.wantDecision {
					t.Errorf("Decision = %q, want %q", got.Decision, tt.wantDecision)
				}
			}

			if tt.checkCalled != nil {
				tt.checkCalled(t, tt.handlers)
			}
		})
	}
}

func TestRegistryDispatchTimeout(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	reg := NewRegistryWithTimeout(cfg, 50*time.Millisecond)

	slow := &slowHandler{event: EventSessionStart, duration: 5 * time.Second}
	reg.Register(slow)

	ctx := context.Background()
	input := &HookInput{
		SessionID:     "test-timeout",
		CWD:           "/tmp",
		HookEventName: "SessionStart",
	}

	_, err := reg.Dispatch(ctx, EventSessionStart, input)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, ErrHookTimeout) {
		t.Errorf("error = %v, want errors.Is(ErrHookTimeout)", err)
	}
}

func TestRegistryDispatchContextCancellation(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	reg := NewRegistry(cfg)

	slow := &slowHandler{event: EventSessionStart, duration: 5 * time.Second}
	reg.Register(slow)

	ctx, cancel := context.WithCancel(context.Background())
	input := &HookInput{
		SessionID:     "test-cancel",
		CWD:           "/tmp",
		HookEventName: "SessionStart",
	}

	// Cancel after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := reg.Dispatch(ctx, EventSessionStart, input)
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}
