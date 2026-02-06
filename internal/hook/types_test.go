package hook

import (
	"testing"
)

func TestValidEventTypes(t *testing.T) {
	t.Parallel()

	events := ValidEventTypes()
	if len(events) != 7 {
		t.Errorf("ValidEventTypes() returned %d events, want 7", len(events))
	}

	expected := map[EventType]bool{
		EventSessionStart: true,
		EventPreToolUse:   true,
		EventPostToolUse:  true,
		EventSessionEnd:   true,
		EventStop:         true,
		EventSubagentStop: true,
		EventPreCompact:   true,
	}

	for _, et := range events {
		if !expected[et] {
			t.Errorf("unexpected event type: %q", et)
		}
	}
}

func TestIsValidEventType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		event EventType
		want  bool
	}{
		{"SessionStart is valid", EventSessionStart, true},
		{"PreToolUse is valid", EventPreToolUse, true},
		{"PostToolUse is valid", EventPostToolUse, true},
		{"SessionEnd is valid", EventSessionEnd, true},
		{"Stop is valid", EventStop, true},
		{"SubagentStop is valid", EventSubagentStop, true},
		{"PreCompact is valid", EventPreCompact, true},
		{"empty string is invalid", EventType(""), false},
		{"unknown event is invalid", EventType("UnknownEvent"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsValidEventType(tt.event); got != tt.want {
				t.Errorf("IsValidEventType(%q) = %v, want %v", tt.event, got, tt.want)
			}
		})
	}
}

func TestNewAllowOutput(t *testing.T) {
	t.Parallel()

	out := NewAllowOutput()
	// PreToolUse uses hookSpecificOutput.permissionDecision, not top-level Decision
	if out.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput is nil")
	}
	if out.HookSpecificOutput.PermissionDecision != DecisionAllow {
		t.Errorf("PermissionDecision = %q, want %q", out.HookSpecificOutput.PermissionDecision, DecisionAllow)
	}
	// Top-level Decision should be empty for PreToolUse
	if out.Decision != "" {
		t.Errorf("Decision = %q, want empty for PreToolUse", out.Decision)
	}
}

func TestNewBlockOutput(t *testing.T) {
	t.Parallel()

	out := NewBlockOutput("test reason")
	// PreToolUse uses hookSpecificOutput.permissionDecision = "deny", not top-level "block"
	if out.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput is nil")
	}
	if out.HookSpecificOutput.PermissionDecision != DecisionDeny {
		t.Errorf("PermissionDecision = %q, want %q", out.HookSpecificOutput.PermissionDecision, DecisionDeny)
	}
	if out.HookSpecificOutput.PermissionDecisionReason != "test reason" {
		t.Errorf("PermissionDecisionReason = %q, want %q", out.HookSpecificOutput.PermissionDecisionReason, "test reason")
	}
}

func TestNewStopBlockOutput(t *testing.T) {
	t.Parallel()

	out := NewStopBlockOutput("continue working")
	// Stop hooks use top-level decision = "block", not hookSpecificOutput
	if out.Decision != DecisionBlock {
		t.Errorf("Decision = %q, want %q", out.Decision, DecisionBlock)
	}
	if out.Reason != "continue working" {
		t.Errorf("Reason = %q, want %q", out.Reason, "continue working")
	}
	// hookSpecificOutput should be nil for Stop hooks
	if out.HookSpecificOutput != nil {
		t.Error("HookSpecificOutput should be nil for Stop hooks")
	}
}

func TestNewPostToolBlockOutput(t *testing.T) {
	t.Parallel()

	out := NewPostToolBlockOutput("test failed", "additional info")
	// PostToolUse uses top-level decision = "block"
	if out.Decision != DecisionBlock {
		t.Errorf("Decision = %q, want %q", out.Decision, DecisionBlock)
	}
	if out.Reason != "test failed" {
		t.Errorf("Reason = %q, want %q", out.Reason, "test failed")
	}
	// PostToolUse can also have hookSpecificOutput.additionalContext
	if out.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput is nil")
	}
	if out.HookSpecificOutput.AdditionalContext != "additional info" {
		t.Errorf("AdditionalContext = %q, want %q", out.HookSpecificOutput.AdditionalContext, "additional info")
	}
}

func TestNewProtocol(t *testing.T) {
	t.Parallel()

	proto := NewProtocol()
	if proto == nil {
		t.Fatal("NewProtocol() returned nil")
	}
}
