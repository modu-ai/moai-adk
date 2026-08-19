package hook

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/goal"
)

// armTestGoal writes an armed goal for sessionID under root and returns nothing
// but a failed test if the write did not take. It exists so each case below
// starts from a state the production code would actually see.
func armTestGoal(t *testing.T, root, sessionID string) {
	t.Helper()
	g := goal.NewGoal(sessionID, "the suite is green", []goal.Condition{
		{Type: goal.ConditionMechanical, Cmd: "true", ExpectExit: 0},
	})
	if err := goal.SaveGoal(root, g); err != nil {
		t.Fatalf("arming the fixture goal: %v", err)
	}
	if _, err := goal.LoadGoal(root, sessionID); err != nil {
		t.Fatalf("fixture goal did not persist: %v", err)
	}
}

// goalIsArmed reports whether a goal state file is still loadable for sessionID.
func goalIsArmed(t *testing.T, root, sessionID string) bool {
	t.Helper()
	g, err := goal.LoadGoal(root, sessionID)
	return err == nil && g != nil
}

// TestStopFailureHandler_DisarmsGoalOnUnrecoverable is the reproduction: a turn
// dies on a credential or billing failure, and the goal it was working toward
// stays armed with nothing left to advance it — so every later turn-end finds
// the condition unmet and burns an iteration against the ceiling.
func TestStopFailureHandler_DisarmsGoalOnUnrecoverable(t *testing.T) {
	for _, errType := range []string{"authentication_failed", "oauth_org_not_allowed", "billing_error"} {
		t.Run(errType, func(t *testing.T) {
			root := t.TempDir()
			const sessionID = "sess-unrecoverable"
			armTestGoal(t, root, sessionID)

			h := NewStopFailureHandler()
			out, err := h.Handle(context.Background(), &HookInput{
				HookEventName: string(EventStopFailure),
				SessionID:     sessionID,
				CWD:           root,
				ErrorType:     errType,
				ErrorMessage:  "the turn ended here",
			})
			if err != nil {
				t.Fatalf("Handle returned an error; the handler must stay non-blocking: %v", err)
			}
			if goalIsArmed(t, root, sessionID) {
				t.Errorf("goal still armed after an unrecoverable %s; it will spin idle turns to the ceiling", errType)
			}
			if !strings.Contains(strings.ToLower(out.SystemMessage), "goal") {
				t.Errorf("systemMessage does not mention the goal that was disarmed: %q", out.SystemMessage)
			}
		})
	}
}

// TestStopFailureHandler_KeepsGoalOnRecoverable is the other half, and the one
// that stops a later simplification to "clear on any StopFailure". These types
// resolve on retry, and the goal is exactly the state that must survive to see
// it — disarming here destroys live state on a condition that fixes itself.
func TestStopFailureHandler_KeepsGoalOnRecoverable(t *testing.T) {
	for _, errType := range []string{"rate_limit", "overloaded", "server_error", "max_output_tokens", "invalid_request", "model_not_found", "unknown", ""} {
		name := errType
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			const sessionID = "sess-recoverable"
			armTestGoal(t, root, sessionID)

			h := NewStopFailureHandler()
			if _, err := h.Handle(context.Background(), &HookInput{
				HookEventName: string(EventStopFailure),
				SessionID:     sessionID,
				CWD:           root,
				ErrorType:     errType,
			}); err != nil {
				t.Fatalf("Handle returned an error: %v", err)
			}
			if !goalIsArmed(t, root, sessionID) {
				t.Errorf("goal disarmed on a recoverable %q; the work resumes on retry and the goal must survive it", errType)
			}
		})
	}
}

// TestStopFailureHandler_FailOpen covers the degradations: no goal armed, no
// resolvable root, no session id. Each must leave the pre-existing error-class
// message intact rather than failing the hook.
func TestStopFailureHandler_FailOpen(t *testing.T) {
	for _, c := range []struct {
		name  string
		input *HookInput
	}{
		{"no goal armed", &HookInput{SessionID: "nobody", CWD: t.TempDir(), ErrorType: "billing_error"}},
		{"no session id", &HookInput{CWD: t.TempDir(), ErrorType: "billing_error"}},
		{"unwritable root", &HookInput{SessionID: "s", CWD: "/nonexistent/moai-t139", ErrorType: "authentication_failed"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			out, err := NewStopFailureHandler().Handle(context.Background(), c.input)
			if err != nil {
				t.Fatalf("Handle returned an error on the %s path: %v", c.name, err)
			}
			if out == nil {
				t.Fatal("Handle returned a nil output")
			}
			// Nothing was armed on any of these paths, so nothing may be
			// announced as disarmed. LoadGoal reports an absent goal as
			// (nil, nil) rather than an error, which is how an error-only
			// guard ends up claiming a disarm that never happened.
			if strings.Contains(out.SystemMessage, "disarmed") {
				t.Errorf("announced a disarm with no goal armed: %q", out.SystemMessage)
			}
		})
	}
}

// TestIsUnrecoverableStopFailure pins a decision for every documented error
// type, so a type added upstream lands as a deliberate choice rather than
// inheriting whatever the default branch happens to do.
func TestIsUnrecoverableStopFailure(t *testing.T) {
	for errType, want := range map[string]bool{
		"authentication_failed": true,
		"oauth_org_not_allowed": true,
		"billing_error":         true,
		"rate_limit":            false,
		"overloaded":            false,
		"invalid_request":       false,
		"model_not_found":       false,
		"server_error":          false,
		"max_output_tokens":     false,
		"unknown":               false,
		"":                      false,
	} {
		if got := goal.IsUnrecoverableStopFailure(errType); got != want {
			t.Errorf("IsUnrecoverableStopFailure(%q) = %v, want %v", errType, got, want)
		}
	}
}
