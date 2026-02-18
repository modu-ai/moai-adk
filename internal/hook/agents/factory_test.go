package agents

import (
	"context"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

func TestNewFactory(t *testing.T) {
	t.Parallel()

	f := NewFactory()
	if f == nil {
		t.Fatal("NewFactory() returned nil")
	}
}

func TestFactory_CreateHandler_ValidActions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		action    string
		wantEvent hook.EventType
	}{
		// DDD handler actions
		{"ddd-pre-transformation", hook.EventPreToolUse},
		{"ddd-post-transformation", hook.EventPostToolUse},
		{"ddd-completion", hook.EventSubagentStop},

		// TDD handler actions
		{"tdd-pre-implementation", hook.EventPreToolUse},
		{"tdd-post-implementation", hook.EventPostToolUse},
		{"tdd-completion", hook.EventSubagentStop},

		// Backend handler actions
		{"backend-validation", hook.EventPreToolUse},
		{"backend-verification", hook.EventPostToolUse},

		// Frontend handler actions
		{"frontend-validation", hook.EventPreToolUse},
		{"frontend-verification", hook.EventPostToolUse},

		// Testing handler actions
		{"testing-verification", hook.EventPostToolUse},
		{"testing-completion", hook.EventSubagentStop},

		// Debug handler actions
		{"debug-verification", hook.EventPostToolUse},
		{"debug-completion", hook.EventSubagentStop},

		// DevOps handler actions
		{"devops-verification", hook.EventPostToolUse},
		{"devops-completion", hook.EventSubagentStop},

		// Quality handler actions
		{"quality-completion", hook.EventSubagentStop},

		// Spec handler actions
		{"spec-completion", hook.EventSubagentStop},

		// Docs handler actions
		{"docs-verification", hook.EventPostToolUse},
		{"docs-completion", hook.EventSubagentStop},
	}

	f := NewFactory()

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			t.Parallel()

			h, err := f.CreateHandler(tt.action)
			if err != nil {
				t.Fatalf("CreateHandler(%q) returned error: %v", tt.action, err)
			}
			if h == nil {
				t.Fatalf("CreateHandler(%q) returned nil handler", tt.action)
			}
			if got := h.EventType(); got != tt.wantEvent {
				t.Errorf("CreateHandler(%q).EventType() = %q, want %q", tt.action, got, tt.wantEvent)
			}
		})
	}
}

func TestFactory_CreateHandler_DefaultHandler(t *testing.T) {
	t.Parallel()

	f := NewFactory()
	h, err := f.CreateHandler("unknown-action")
	if err != nil {
		t.Fatalf("CreateHandler(\"unknown-action\") returned error: %v", err)
	}
	if h == nil {
		t.Fatal("CreateHandler(\"unknown-action\") returned nil handler")
	}
	// Default handler uses PreToolUse.
	if got := h.EventType(); got != hook.EventPreToolUse {
		t.Errorf("default handler EventType() = %q, want %q", got, hook.EventPreToolUse)
	}
}

func TestFactory_CreateHandler_InvalidFormat(t *testing.T) {
	t.Parallel()

	f := NewFactory()

	invalidActions := []string{
		"nodelimiter",
		"",
		"single",
	}

	for _, action := range invalidActions {
		t.Run("action_"+action, func(t *testing.T) {
			t.Parallel()

			_, err := f.CreateHandler(action)
			if err == nil {
				t.Errorf("CreateHandler(%q) should return error for invalid format", action)
			}
		})
	}
}

func TestFactory_CreateHandler_HandleReturnsAllowOutput(t *testing.T) {
	t.Parallel()

	// Every known handler should return a non-nil output with no error from Handle().
	actions := []string{
		"ddd-pre-transformation",
		"ddd-post-transformation",
		"ddd-completion",
		"tdd-pre-implementation",
		"tdd-post-implementation",
		"tdd-completion",
		"backend-validation",
		"backend-verification",
		"frontend-validation",
		"frontend-verification",
		"testing-verification",
		"testing-completion",
		"debug-verification",
		"debug-completion",
		"devops-verification",
		"devops-completion",
		"quality-completion",
		"spec-completion",
		"docs-verification",
		"docs-completion",
		"unknown-action",
	}

	f := NewFactory()
	ctx := context.Background()

	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			t.Parallel()

			h, err := f.CreateHandler(action)
			if err != nil {
				t.Fatalf("CreateHandler(%q) error: %v", action, err)
			}

			output, err := h.Handle(ctx, &hook.HookInput{})
			if err != nil {
				t.Fatalf("Handle() for %q returned error: %v", action, err)
			}
			if output == nil {
				t.Fatalf("Handle() for %q returned nil output", action)
			}
		})
	}
}

func TestDDDHandler_ActionBranches(t *testing.T) {
	t.Parallel()

	// DDD handler has a switch statement on action; exercise every branch.
	actions := []string{
		"pre-transformation",
		"post-transformation",
		"completion",
		"unknown-ddd-action", // default branch
	}

	ctx := context.Background()

	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			t.Parallel()

			h := NewDDDHandler(action)
			output, err := h.Handle(ctx, &hook.HookInput{})
			if err != nil {
				t.Fatalf("DDDHandler.Handle(%q) error: %v", action, err)
			}
			if output == nil {
				t.Fatalf("DDDHandler.Handle(%q) returned nil output", action)
			}
		})
	}
}

func TestDDDHandler_EventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		action    string
		wantEvent hook.EventType
	}{
		{"pre-transformation", hook.EventPreToolUse},
		{"post-transformation", hook.EventPostToolUse},
		{"completion", hook.EventSubagentStop},
		{"unknown", hook.EventPreToolUse}, // default
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			t.Parallel()

			h := NewDDDHandler(tt.action)
			if got := h.EventType(); got != tt.wantEvent {
				t.Errorf("NewDDDHandler(%q).EventType() = %q, want %q", tt.action, got, tt.wantEvent)
			}
		})
	}
}

func TestTDDHandler_EventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		action    string
		wantEvent hook.EventType
	}{
		{"pre-implementation", hook.EventPreToolUse},
		{"post-implementation", hook.EventPostToolUse},
		{"completion", hook.EventSubagentStop},
		{"unknown", hook.EventPreToolUse}, // default
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			t.Parallel()

			h := NewTDDHandler(tt.action)
			if got := h.EventType(); got != tt.wantEvent {
				t.Errorf("NewTDDHandler(%q).EventType() = %q, want %q", tt.action, got, tt.wantEvent)
			}
		})
	}
}

func TestBackendHandler_EventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		action    string
		wantEvent hook.EventType
	}{
		{"validation", hook.EventPreToolUse},
		{"verification", hook.EventPostToolUse},
		{"other", hook.EventPreToolUse}, // default
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			t.Parallel()

			h := NewBackendHandler(tt.action)
			if got := h.EventType(); got != tt.wantEvent {
				t.Errorf("NewBackendHandler(%q).EventType() = %q, want %q", tt.action, got, tt.wantEvent)
			}
		})
	}
}

func TestFrontendHandler_EventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		action    string
		wantEvent hook.EventType
	}{
		{"validation", hook.EventPreToolUse},
		{"verification", hook.EventPostToolUse},
		{"other", hook.EventPreToolUse}, // default
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			t.Parallel()

			h := NewFrontendHandler(tt.action)
			if got := h.EventType(); got != tt.wantEvent {
				t.Errorf("NewFrontendHandler(%q).EventType() = %q, want %q", tt.action, got, tt.wantEvent)
			}
		})
	}
}

func TestDebugHandler_EventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		action    string
		wantEvent hook.EventType
	}{
		{"verification", hook.EventPostToolUse},
		{"completion", hook.EventSubagentStop},
		{"other", hook.EventPostToolUse}, // default (not "completion")
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			t.Parallel()

			h := NewDebugHandler(tt.action)
			if got := h.EventType(); got != tt.wantEvent {
				t.Errorf("NewDebugHandler(%q).EventType() = %q, want %q", tt.action, got, tt.wantEvent)
			}
		})
	}
}

func TestDevOpsHandler_EventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		action    string
		wantEvent hook.EventType
	}{
		{"verification", hook.EventPostToolUse},
		{"completion", hook.EventSubagentStop},
		{"other", hook.EventPostToolUse}, // default
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			t.Parallel()

			h := NewDevOpsHandler(tt.action)
			if got := h.EventType(); got != tt.wantEvent {
				t.Errorf("NewDevOpsHandler(%q).EventType() = %q, want %q", tt.action, got, tt.wantEvent)
			}
		})
	}
}

func TestDocsHandler_EventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		action    string
		wantEvent hook.EventType
	}{
		{"verification", hook.EventPostToolUse},
		{"completion", hook.EventSubagentStop},
		{"other", hook.EventPostToolUse}, // default
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			t.Parallel()

			h := NewDocsHandler(tt.action)
			if got := h.EventType(); got != tt.wantEvent {
				t.Errorf("NewDocsHandler(%q).EventType() = %q, want %q", tt.action, got, tt.wantEvent)
			}
		})
	}
}

func TestTestingHandler_EventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		action    string
		wantEvent hook.EventType
	}{
		{"verification", hook.EventPostToolUse},
		{"completion", hook.EventSubagentStop},
		{"other", hook.EventPostToolUse}, // default
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			t.Parallel()

			h := NewTestingHandler(tt.action)
			if got := h.EventType(); got != tt.wantEvent {
				t.Errorf("NewTestingHandler(%q).EventType() = %q, want %q", tt.action, got, tt.wantEvent)
			}
		})
	}
}

func TestQualityHandler_AlwaysSubagentStop(t *testing.T) {
	t.Parallel()

	actions := []string{"completion", "anything", ""}
	for _, action := range actions {
		t.Run("action_"+action, func(t *testing.T) {
			t.Parallel()

			h := NewQualityHandler(action)
			if got := h.EventType(); got != hook.EventSubagentStop {
				t.Errorf("NewQualityHandler(%q).EventType() = %q, want %q", action, got, hook.EventSubagentStop)
			}
		})
	}
}

func TestSpecHandler_AlwaysSubagentStop(t *testing.T) {
	t.Parallel()

	actions := []string{"completion", "anything", ""}
	for _, action := range actions {
		t.Run("action_"+action, func(t *testing.T) {
			t.Parallel()

			h := NewSpecHandler(action)
			if got := h.EventType(); got != hook.EventSubagentStop {
				t.Errorf("NewSpecHandler(%q).EventType() = %q, want %q", action, got, hook.EventSubagentStop)
			}
		})
	}
}

func TestDefaultHandler(t *testing.T) {
	t.Parallel()

	h := NewDefaultHandler("some-action")

	if got := h.EventType(); got != hook.EventPreToolUse {
		t.Errorf("DefaultHandler.EventType() = %q, want %q", got, hook.EventPreToolUse)
	}

	output, err := h.Handle(context.Background(), &hook.HookInput{})
	if err != nil {
		t.Fatalf("DefaultHandler.Handle() error: %v", err)
	}
	if output == nil {
		t.Fatal("DefaultHandler.Handle() returned nil output")
	}
}

func TestBaseHandler_Handle(t *testing.T) {
	t.Parallel()

	// The baseHandler is embedded; verify it works through a handler
	// that does NOT override Handle (none currently, but test the base directly
	// through the factory's default path).
	f := NewFactory()
	h, err := f.CreateHandler("unknown-agent-action")
	if err != nil {
		t.Fatalf("CreateHandler for default: %v", err)
	}

	output, err := h.Handle(context.Background(), nil)
	if err != nil {
		t.Fatalf("baseHandler.Handle() error: %v", err)
	}
	if output == nil {
		t.Fatal("baseHandler.Handle() returned nil output")
	}
}
