package hook

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidEventTypes(t *testing.T) {
	t.Parallel()

	events := ValidEventTypes()
	if len(events) != 30 {
		t.Errorf("ValidEventTypes() returned %d events, want 30 (3 observe-only events added by SPEC-HOOK-EVENT-REGISTRY-001 + official Setup recognition)", len(events))
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
		EventSetup:               true,
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

// TestNewSafeDefaultOutput verifies the mode-aware PreToolUse safe-path
// decision: cautious modes ("default", "plan") defer to Claude Code's normal
// permission flow (empty no-opinion output), while autonomous modes and any
// empty/unrecognized mode preserve the existing "allow" behavior.
func TestNewSafeDefaultOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		permissionMode string
		wantEmpty      bool // true: expect &HookOutput{} (no permissionDecision)
	}{
		{name: "default mode defers to normal permission flow", permissionMode: PermissionModeDefault, wantEmpty: true},
		{name: "plan mode defers to normal permission flow", permissionMode: PermissionModePlan, wantEmpty: true},
		{name: "acceptEdits mode allows", permissionMode: PermissionModeAcceptEdits, wantEmpty: false},
		{name: "bypassPermissions mode allows", permissionMode: PermissionModeBypassPermissions, wantEmpty: false},
		{name: "auto mode allows", permissionMode: PermissionModeAuto, wantEmpty: false},
		{name: "dontAsk mode allows", permissionMode: PermissionModeDontAsk, wantEmpty: false},
		{name: "empty mode allows (autonomous fallback)", permissionMode: "", wantEmpty: false},
		{name: "unrecognized mode allows (autonomous fallback)", permissionMode: "some-future-mode", wantEmpty: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewSafeDefaultOutput(tt.permissionMode)
			if got == nil {
				t.Fatal("NewSafeDefaultOutput() returned nil")
			}

			if tt.wantEmpty {
				if got.HookSpecificOutput != nil {
					t.Errorf("HookSpecificOutput = %+v, want nil for mode %q", got.HookSpecificOutput, tt.permissionMode)
				}
				if got.Decision != "" {
					t.Errorf("Decision = %q, want empty for mode %q", got.Decision, tt.permissionMode)
				}
			} else {
				if got.HookSpecificOutput == nil {
					t.Fatalf("HookSpecificOutput is nil for mode %q, want allow decision", tt.permissionMode)
				}
				if got.HookSpecificOutput.PermissionDecision != DecisionAllow {
					t.Errorf("PermissionDecision = %q, want %q for mode %q",
						got.HookSpecificOutput.PermissionDecision, DecisionAllow, tt.permissionMode)
				}
			}
		})
	}
}

// TestPermissionModeOf verifies the nil-safety helper used by both pre_tool.go
// safe-path returns and registry.go's PreToolUse fallback.
func TestPermissionModeOf(t *testing.T) {
	t.Parallel()

	if got := permissionModeOf(nil); got != "" {
		t.Errorf("permissionModeOf(nil) = %q, want empty", got)
	}

	input := &HookInput{PermissionMode: PermissionModeDefault}
	if got := permissionModeOf(input); got != PermissionModeDefault {
		t.Errorf("permissionModeOf(input) = %q, want %q", got, PermissionModeDefault)
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
	// Official PermissionRequest schema: hookSpecificOutput.decision{behavior},
	// NOT permissionDecision (that field is PreToolUse-only).
	if out.HookSpecificOutput.Decision == nil {
		t.Fatal("Decision is nil, want official decision object")
	}
	if out.HookSpecificOutput.Decision.Behavior != DecisionAllow {
		t.Errorf("Decision.Behavior = %q, want %q", out.HookSpecificOutput.Decision.Behavior, DecisionAllow)
	}
	if out.HookSpecificOutput.PermissionDecision != "" {
		t.Errorf("PermissionDecision = %q, want empty (non-schema for PermissionRequest)", out.HookSpecificOutput.PermissionDecision)
	}
	// The reason rides systemMessage (the decision object has no reason slot).
	if out.SystemMessage != "auto-approved tool" {
		t.Errorf("SystemMessage = %q, want %q", out.SystemMessage, "auto-approved tool")
	}
	if out.HookSpecificOutput.HookEventName != "PermissionRequest" {
		t.Errorf("HookEventName = %q, want %q", out.HookSpecificOutput.HookEventName, "PermissionRequest")
	}

	// Wire-format check: JSON must carry decision.behavior, not permissionDecision.
	data, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, `"decision":{"behavior":"allow"}`) {
		t.Errorf("JSON missing official decision object: %s", s)
	}
	if strings.Contains(s, "permissionDecision") {
		t.Errorf("JSON must not carry permissionDecision for PermissionRequest: %s", s)
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

	// Official TeammateIdle blocking channel: decision:"block" + reason JSON
	// with exit 0 (NOT ExitCode=2, whose stdout JSON is ignored by Claude Code).
	out := NewTeammateKeepWorkingOutput("quality gate failed")
	if out.Decision != DecisionBlock {
		t.Errorf("Decision = %q, want %q", out.Decision, DecisionBlock)
	}
	if out.Reason != "quality gate failed" {
		t.Errorf("Reason = %q, want %q", out.Reason, "quality gate failed")
	}
	if out.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0 (block rides JSON, not exit code)", out.ExitCode)
	}
}

func TestNewTaskRejectedOutput(t *testing.T) {
	t.Parallel()

	// Official TaskCompleted rejection channel: decision:"block" + reason JSON
	// with exit 0 (NOT ExitCode=2, whose stdout JSON is ignored by Claude Code).
	out := NewTaskRejectedOutput("unchecked acceptance criteria")
	if out.Decision != DecisionBlock {
		t.Errorf("Decision = %q, want %q", out.Decision, DecisionBlock)
	}
	if out.Reason != "unchecked acceptance criteria" {
		t.Errorf("Reason = %q, want %q", out.Reason, "unchecked acceptance criteria")
	}
	if out.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0 (block rides JSON, not exit code)", out.ExitCode)
	}
}

// TestHookOutputContinueMarshaling verifies the G1 fix: "continue": false is
// the only meaningful value per the official spec, so the field must be
// absent by default (nil) and representable when explicitly set to false —
// the old plain-bool + omitempty could never emit false.
func TestHookOutputContinueMarshaling(t *testing.T) {
	t.Parallel()

	t.Run("absent by default", func(t *testing.T) {
		t.Parallel()
		data, err := json.Marshal(&HookOutput{})
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		if strings.Contains(string(data), `"continue"`) {
			t.Errorf("continue must be absent from default output: %s", data)
		}
	})

	t.Run("explicit false is emitted", func(t *testing.T) {
		t.Parallel()
		out := &HookOutput{StopReason: "halting"}
		out.SetContinue(false)
		data, err := json.Marshal(out)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		if !strings.Contains(string(data), `"continue":false`) {
			t.Errorf(`JSON must carry "continue":false when explicitly halted: %s`, data)
		}
	})

	t.Run("BoolPtr helper", func(t *testing.T) {
		t.Parallel()
		p := BoolPtr(false)
		if p == nil || *p != false {
			t.Errorf("BoolPtr(false) = %v, want pointer to false", p)
		}
	})
}

// TestEventSetupRecognized verifies the official Setup event validates
// (recognition-only; no handler, no CLI binding).
func TestEventSetupRecognized(t *testing.T) {
	t.Parallel()

	if !IsValidEventType(EventSetup) {
		t.Error("IsValidEventType(EventSetup) = false, want true (official Setup event)")
	}
	if EventSetup != "Setup" {
		t.Errorf("EventSetup = %q, want %q", EventSetup, "Setup")
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
