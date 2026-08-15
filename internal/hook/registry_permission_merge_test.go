// Regression tests for the permission-decision merge in registry.Dispatch.
//
// The defect these pin: Dispatch pre-seeds the merged output with the event
// default — already permissionDecision "allow" for PreToolUse under the
// autonomous permission modes — and mergeHandlerOutput copied every
// hookSpecificOutput field EXCEPT PermissionDecision. A handler's "ask" was
// therefore dropped on the floor and the pre-seeded "allow" reached the wire,
// silently downgrading all 27 AskPatterns and all 10 AskBashPatterns.
//
// It shipped green because every pre-existing "ask" assertion called
// preToolHandler.Handle directly and never crossed registry.Dispatch: the
// producer was tested, the consumer was not. These tests cross Dispatch.
package hook

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
)

// dispatchPreTool runs a real preToolHandler through a real registry and
// returns the dispatched output — the consumer-side path the handler-level
// tests never exercised.
func dispatchPreTool(t *testing.T, projectDir string, input *HookInput) *HookOutput {
	t.Helper()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	reg := NewRegistry(cfg)
	reg.Register(&preToolHandler{
		cfg:        cfg,
		policy:     DefaultSecurityPolicy(),
		projectDir: projectDir,
	})

	got, err := reg.Dispatch(context.Background(), EventPreToolUse, input)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if got == nil || got.HookSpecificOutput == nil {
		t.Fatal("Dispatch returned nil output or nil HookSpecificOutput")
	}
	return got
}

// TestDispatch_PreToolUse_AskPatternSurvivesMerge asserts a file-pattern "ask"
// reaches the dispatched output with its reason intact. Before the merge fix
// this returned "allow".
func TestDispatch_PreToolUse_AskPatternSurvivesMerge(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	toolInput, err := json.Marshal(map[string]string{
		"file_path":  filepath.Join(projectDir, "package.json"),
		"old_string": "a",
		"new_string": "b",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	got := dispatchPreTool(t, projectDir, &HookInput{
		SessionID:     "sess-dispatch-ask-file",
		CWD:           projectDir,
		HookEventName: "PreToolUse",
		ToolName:      "Edit",
		ToolInput:     json.RawMessage(toolInput),
	})

	if got.HookSpecificOutput.PermissionDecision != DecisionAsk {
		t.Errorf("permissionDecision = %q, want %q (the handler's ask was lost in the merge)",
			got.HookSpecificOutput.PermissionDecision, DecisionAsk)
	}
	if got.HookSpecificOutput.PermissionDecisionReason == "" {
		t.Error("permissionDecisionReason is empty; an ask without a reason shows the user a dialog that explains nothing")
	}
}

// TestDispatch_PreToolUse_AskBashPatternSurvivesMerge asserts the Bash side of
// the same defect: `git reset --hard` is an AskBashPattern and was equally
// downgraded to "allow".
func TestDispatch_PreToolUse_AskBashPatternSurvivesMerge(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	toolInput, err := json.Marshal(map[string]string{
		"command": "git reset --hard HEAD~1",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	got := dispatchPreTool(t, projectDir, &HookInput{
		SessionID:     "sess-dispatch-ask-bash",
		CWD:           projectDir,
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		ToolInput:     json.RawMessage(toolInput),
	})

	if got.HookSpecificOutput.PermissionDecision != DecisionAsk {
		t.Errorf("permissionDecision = %q, want %q for `git reset --hard`",
			got.HookSpecificOutput.PermissionDecision, DecisionAsk)
	}
	if got.HookSpecificOutput.PermissionDecisionReason == "" {
		t.Error("permissionDecisionReason is empty for the ask-bash path")
	}
}

// TestDispatch_PreToolUse_DenyStillShortCircuits guards the decision that was
// already correct: deny short-circuits before the merge and must stay that way.
func TestDispatch_PreToolUse_DenyStillShortCircuits(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	toolInput, err := json.Marshal(map[string]string{
		"file_path":  filepath.Join(projectDir, ".git", "config"),
		"old_string": "a",
		"new_string": "b",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	got := dispatchPreTool(t, projectDir, &HookInput{
		SessionID:     "sess-dispatch-deny",
		CWD:           projectDir,
		HookEventName: "PreToolUse",
		ToolName:      "Edit",
		ToolInput:     json.RawMessage(toolInput),
	})

	if got.HookSpecificOutput.PermissionDecision != DecisionDeny {
		t.Errorf("permissionDecision = %q, want %q", got.HookSpecificOutput.PermissionDecision, DecisionDeny)
	}
}

// TestMergeHandlerOutput_PermissionPrecedence pins the ladder itself:
// deny > ask > allow > unset, with the reason travelling alongside the
// decision. A blind field copy would let a later handler's allow clobber an
// earlier handler's ask — this is why the fix is a ladder and not a copy.
func TestMergeHandlerOutput_PermissionPrecedence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dst        string
		dstReason  string
		src        string
		srcReason  string
		want       string
		wantReason string
	}{
		{
			name: "ask beats the pre-seeded allow", dst: DecisionAllow, src: DecisionAsk,
			srcReason: "critical config file", want: DecisionAsk, wantReason: "critical config file",
		},
		{
			name: "allow never clobbers an earlier ask", dst: DecisionAsk, dstReason: "keep me",
			src: DecisionAllow, want: DecisionAsk, wantReason: "keep me",
		},
		{
			name: "allow never clobbers an earlier deny", dst: DecisionDeny, dstReason: "protected",
			src: DecisionAllow, want: DecisionDeny, wantReason: "protected",
		},
		{
			name: "ask never clobbers an earlier deny", dst: DecisionDeny, dstReason: "protected",
			src: DecisionAsk, srcReason: "confirm", want: DecisionDeny, wantReason: "protected",
		},
		{
			name: "deny beats an earlier ask", dst: DecisionAsk, dstReason: "confirm",
			src: DecisionDeny, srcReason: "protected", want: DecisionDeny, wantReason: "protected",
		},
		{
			name: "an unset src leaves the accumulator untouched", dst: DecisionAsk, dstReason: "confirm",
			src: "", want: DecisionAsk, wantReason: "confirm",
		},
		{
			name: "an explicit allow fills an unset accumulator", dst: "", src: DecisionAllow,
			want: DecisionAllow, wantReason: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			merged := &HookOutput{HookSpecificOutput: &HookSpecificOutput{
				HookEventName:            "PreToolUse",
				PermissionDecision:       tt.dst,
				PermissionDecisionReason: tt.dstReason,
			}}
			output := &HookOutput{HookSpecificOutput: &HookSpecificOutput{
				HookEventName:            "PreToolUse",
				PermissionDecision:       tt.src,
				PermissionDecisionReason: tt.srcReason,
			}}

			mergeHandlerOutput(merged, output)

			if got := merged.HookSpecificOutput.PermissionDecision; got != tt.want {
				t.Errorf("permissionDecision = %q, want %q", got, tt.want)
			}
			if got := merged.HookSpecificOutput.PermissionDecisionReason; got != tt.wantReason {
				t.Errorf("permissionDecisionReason = %q, want %q", got, tt.wantReason)
			}
		})
	}
}

// TestMergeHandlerOutput_PermissionRequestDecision covers the sibling field the
// same omission pattern left uncopied: hookSpecificOutput.decision.behavior,
// used by PermissionRequest (which has no "ask" behavior — only allow/deny).
//
// This was NOT reachable as a live defect: defaultOutputForEvent returns a bare
// &HookOutput{} for PermissionRequest, so the nil-HookSpecificOutput
// early-assign branch carried the whole object through, and only one handler is
// registered for the event. It was sound by accident on two counts. The ladder
// now covers it explicitly, and these cases pin it.
func TestMergeHandlerOutput_PermissionRequestDecision(t *testing.T) {
	t.Parallel()

	t.Run("decision survives into a non-nil accumulator", func(t *testing.T) {
		t.Parallel()

		merged := &HookOutput{HookSpecificOutput: &HookSpecificOutput{HookEventName: "PermissionRequest"}}
		output := &HookOutput{HookSpecificOutput: &HookSpecificOutput{
			HookEventName: "PermissionRequest",
			Decision:      &PermissionRequestDecision{Behavior: DecisionDeny},
		}}

		mergeHandlerOutput(merged, output)

		if merged.HookSpecificOutput.Decision == nil {
			t.Fatal("decision was dropped by the merge")
		}
		if got := merged.HookSpecificOutput.Decision.Behavior; got != DecisionDeny {
			t.Errorf("decision.behavior = %q, want %q", got, DecisionDeny)
		}
	})

	t.Run("allow never clobbers an earlier deny", func(t *testing.T) {
		t.Parallel()

		merged := &HookOutput{HookSpecificOutput: &HookSpecificOutput{
			HookEventName: "PermissionRequest",
			Decision:      &PermissionRequestDecision{Behavior: DecisionDeny},
		}}
		output := &HookOutput{HookSpecificOutput: &HookSpecificOutput{
			HookEventName: "PermissionRequest",
			Decision:      &PermissionRequestDecision{Behavior: DecisionAllow},
		}}

		mergeHandlerOutput(merged, output)

		if got := merged.HookSpecificOutput.Decision.Behavior; got != DecisionDeny {
			t.Errorf("decision.behavior = %q, want %q", got, DecisionDeny)
		}
	})
}
