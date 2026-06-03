package hook

import (
	"encoding/json"
	"testing"
)

func TestValidEventTypes(t *testing.T) {
	t.Parallel()

	events := ValidEventTypes()
	if len(events) != 29 {
		t.Errorf("ValidEventTypes() returned %d events, want 29 (3 observe-only events added by SPEC-HOOK-EVENT-REGISTRY-001)", len(events))
	}

	expected := map[EventType]bool{
		EventSessionStart:        true,
		EventPreToolUse:          true,
		EventPostToolUse:         true,
		EventSessionEnd:          true,
		EventStop:                true,
		EventSubagentStop:        true,
		EventPreCompact:          true,
		EventPostToolUseFailure:  true,
		EventNotification:        true,
		EventSubagentStart:       true,
		EventUserPromptSubmit:    true,
		EventPermissionRequest:   true,
		EventTeammateIdle:        true,
		EventTaskCompleted:       true,
		EventWorktreeCreate:      true,
		EventWorktreeRemove:      true,
		EventPostCompact:         true,
		EventInstructionsLoaded:  true,
		EventStopFailure:         true,
		EventConfigChange:        true,
		EventTaskCreated:         true,
		EventCwdChanged:          true,
		EventFileChanged:         true,
		EventElicitation:         true,
		EventElicitationResult:   true,
		EventPermissionDenied:    true,
		EventPostToolBatch:       true,
		EventUserPromptExpansion: true,
		EventMessageDisplay:      true,
	}

	for _, et := range events {
		if !expected[et] {
			t.Errorf("unexpected event type: %q", et)
		}
	}
}

// TestCoverageTableLen asserts the CoverageTable row count after the 3
// observe-only events were added by SPEC-HOOK-EVENT-REGISTRY-001.
// NOTE: this is 30 rows, NOT 29 — the table carries 26 real events + 1
// synthetic COMPOSITE row ("AutoUpdate") + 3 new observe-only rows = 30.
// The "29-event" figure in coverage_table.go is the event-semantic count,
// distinct from len(CoverageTable).
func TestCoverageTableLen(t *testing.T) {
	t.Parallel()

	if len(CoverageTable) != 30 {
		t.Errorf("len(CoverageTable) = %d, want 30 (27 prior rows + 3 observe-only events)", len(CoverageTable))
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
		{"WorktreeCreate is valid", EventWorktreeCreate, true},
		{"WorktreeRemove is valid", EventWorktreeRemove, true},
		{"PostToolBatch is valid", EventPostToolBatch, true},
		{"UserPromptExpansion is valid", EventUserPromptExpansion, true},
		{"MessageDisplay is valid", EventMessageDisplay, true},
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

func TestNewPermissionRequestOutput(t *testing.T) {
	t.Parallel()

	out := NewPermissionRequestOutput(DecisionAllow, "auto-approved tool")
	if out.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput is nil")
	}
	if out.HookSpecificOutput.PermissionDecision != DecisionAllow {
		t.Errorf("PermissionDecision = %q, want %q", out.HookSpecificOutput.PermissionDecision, DecisionAllow)
	}
	if out.HookSpecificOutput.PermissionDecisionReason != "auto-approved tool" {
		t.Errorf("PermissionDecisionReason = %q, want %q", out.HookSpecificOutput.PermissionDecisionReason, "auto-approved tool")
	}
	if out.HookSpecificOutput.HookEventName != "PermissionRequest" {
		t.Errorf("HookEventName = %q, want %q", out.HookSpecificOutput.HookEventName, "PermissionRequest")
	}
}

func TestNewUserPromptBlockOutput(t *testing.T) {
	t.Parallel()

	out := NewUserPromptBlockOutput("blocked for safety")
	if out.Decision != DecisionBlock {
		t.Errorf("Decision = %q, want %q", out.Decision, DecisionBlock)
	}
	if out.Reason != "blocked for safety" {
		t.Errorf("Reason = %q, want %q", out.Reason, "blocked for safety")
	}
}

func TestNewTeammateKeepWorkingOutput(t *testing.T) {
	t.Parallel()

	out := NewTeammateKeepWorkingOutput()
	if out.ExitCode != 2 {
		t.Errorf("ExitCode = %d, want 2", out.ExitCode)
	}
}

func TestNewTaskRejectedOutput(t *testing.T) {
	t.Parallel()

	out := NewTaskRejectedOutput()
	if out.ExitCode != 2 {
		t.Errorf("ExitCode = %d, want 2", out.ExitCode)
	}
}

func TestHookOutput_UpdatedInput_JSON(t *testing.T) {
	t.Parallel()

	out := &HookOutput{
		UpdatedInput: "modified prompt",
	}
	data, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if v, ok := m["updatedInput"]; !ok {
		t.Error("updatedInput key missing from JSON output")
	} else if v != "modified prompt" {
		t.Errorf("updatedInput = %q, want %q", v, "modified prompt")
	}
}

func TestHookOutput_ExitCode_NotSerialized(t *testing.T) {
	t.Parallel()

	out := &HookOutput{
		ExitCode: 2,
	}
	data, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if _, ok := m["ExitCode"]; ok {
		t.Error("ExitCode should not be serialized to JSON (json:\"-\" tag)")
	}
	if _, ok := m["exitCode"]; ok {
		t.Error("exitCode should not be serialized to JSON (json:\"-\" tag)")
	}
}
