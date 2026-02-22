package hook

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// TestExitCodeBehavior_TeammateKeepWorking verifies that NewTeammateKeepWorkingOutput
// sets ExitCode=2 internally but does NOT serialize it to JSON.
func TestExitCodeBehavior_TeammateKeepWorking(t *testing.T) {
	t.Parallel()

	output := NewTeammateKeepWorkingOutput()
	if output.ExitCode != 2 {
		t.Errorf("ExitCode = %d, want 2", output.ExitCode)
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	s := string(data)
	if strings.Contains(s, "exitCode") || strings.Contains(s, "ExitCode") {
		t.Errorf("ExitCode should not be in JSON: %s", s)
	}

	if !json.Valid(data) {
		t.Fatalf("output is not valid JSON: %s", data)
	}
}

// TestExitCodeBehavior_TaskRejected verifies that NewTaskRejectedOutput
// sets ExitCode=2 internally but does NOT serialize it to JSON.
func TestExitCodeBehavior_TaskRejected(t *testing.T) {
	t.Parallel()

	output := NewTaskRejectedOutput()
	if output.ExitCode != 2 {
		t.Errorf("ExitCode = %d, want 2", output.ExitCode)
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	s := string(data)
	if strings.Contains(s, "exitCode") || strings.Contains(s, "ExitCode") {
		t.Errorf("ExitCode should not be in JSON: %s", s)
	}

	if !json.Valid(data) {
		t.Fatalf("output is not valid JSON: %s", data)
	}
}

// TestExitCodeBehavior_DefaultHandlers verifies that the real handlers for
// TeammateIdle and TaskCompleted return ExitCode 0 by default (accept idle / accept completion).
func TestExitCodeBehavior_DefaultHandlers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler Handler
		event   EventType
	}{
		{
			name:    "TeammateIdleHandler returns ExitCode 0",
			handler: NewTeammateIdleHandler(),
			event:   EventTeammateIdle,
		},
		{
			name:    "TaskCompletedHandler returns ExitCode 0",
			handler: NewTaskCompletedHandler(),
			event:   EventTaskCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.handler.EventType() != tt.event {
				t.Errorf("EventType() = %q, want %q", tt.handler.EventType(), tt.event)
			}

			input := &HookInput{
				SessionID:     "test-exit-code",
				CWD:           "/tmp",
				HookEventName: string(tt.event),
				AgentID:       "test-agent",
			}

			output, err := tt.handler.Handle(context.Background(), input)
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}
			if output == nil {
				t.Fatal("Handle() returned nil output")
			}
			if output.ExitCode != 0 {
				t.Errorf("ExitCode = %d, want 0 (default accept)", output.ExitCode)
			}
		})
	}
}

// TestExitCodeBehavior_RegistryPreservesExitCode verifies that the registry
// short-circuits and returns a handler's output when ExitCode==2, so the CLI
// can exit with code 2 for TeammateIdle (keep-working) and TaskCompleted (reject).
func TestExitCodeBehavior_RegistryPreservesExitCode(t *testing.T) {
	t.Parallel()

	t.Run("handler ExitCode=2 is preserved through dispatch (TeammateIdle keep-working)", func(t *testing.T) {
		t.Parallel()

		// NewTeammateKeepWorkingOutput returns ExitCode=2 without a block Decision.
		// The registry must short-circuit and return this output so the CLI can exit(2).
		handler := &mockHandler{
			event:  EventTeammateIdle,
			output: NewTeammateKeepWorkingOutput(), // ExitCode=2, no block Decision
		}

		cfg := &mockConfigProvider{cfg: newTestConfig()}
		reg := NewRegistry(cfg)
		reg.Register(handler)

		input := &HookInput{
			SessionID:     "test-exit-keep-working",
			CWD:           "/tmp",
			HookEventName: string(EventTeammateIdle),
			AgentID:       "tm-keep-working",
		}

		got, err := reg.Dispatch(context.Background(), EventTeammateIdle, input)
		if err != nil {
			t.Fatalf("Dispatch() unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("Dispatch() returned nil output")
		}

		// ExitCode=2 must be preserved so the CLI can call os.Exit(2).
		if got.ExitCode != 2 {
			t.Errorf("ExitCode = %d, want 2 (TeammateIdle keep-working signal)", got.ExitCode)
		}
		if !handler.called {
			t.Error("handler was not called")
		}
	})

	t.Run("handler ExitCode=2 is preserved through dispatch (TaskCompleted reject)", func(t *testing.T) {
		t.Parallel()

		handler := &mockHandler{
			event:  EventTaskCompleted,
			output: NewTaskRejectedOutput(), // ExitCode=2, no block Decision
		}

		cfg := &mockConfigProvider{cfg: newTestConfig()}
		reg := NewRegistry(cfg)
		reg.Register(handler)

		input := &HookInput{
			SessionID:     "test-exit-task-rejected",
			CWD:           "/tmp",
			HookEventName: string(EventTaskCompleted),
			AgentID:       "tm-reject",
		}

		got, err := reg.Dispatch(context.Background(), EventTaskCompleted, input)
		if err != nil {
			t.Fatalf("Dispatch() unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("Dispatch() returned nil output")
		}

		// ExitCode=2 must be preserved so the CLI can call os.Exit(2).
		if got.ExitCode != 2 {
			t.Errorf("ExitCode = %d, want 2 (TaskCompleted reject signal)", got.ExitCode)
		}
		if !handler.called {
			t.Error("handler was not called")
		}
	})

	t.Run("blocking handler output with ExitCode is preserved through dispatch", func(t *testing.T) {
		t.Parallel()

		// Block Decision causes short-circuit: handler output is returned directly
		blockOutput := &HookOutput{
			Decision: DecisionBlock,
			Reason:   "rejected",
			ExitCode: 2,
		}
		handler := &mockHandler{
			event:  EventTeammateIdle,
			output: blockOutput,
		}

		cfg := &mockConfigProvider{cfg: newTestConfig()}
		reg := NewRegistry(cfg)
		reg.Register(handler)

		input := &HookInput{
			SessionID:     "test-exit-block",
			CWD:           "/tmp",
			HookEventName: string(EventTeammateIdle),
			AgentID:       "tm-block",
		}

		got, err := reg.Dispatch(context.Background(), EventTeammateIdle, input)
		if err != nil {
			t.Fatalf("Dispatch() unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("Dispatch() returned nil output")
		}

		// Block decision short-circuits: handler output is returned directly
		if got.ExitCode != 2 {
			t.Errorf("ExitCode = %d, want 2 (preserved from block handler)", got.ExitCode)
		}
		if got.Decision != DecisionBlock {
			t.Errorf("Decision = %q, want %q", got.Decision, DecisionBlock)
		}
		if !handler.called {
			t.Error("handler was not called")
		}
	})
}

// TestExitCodeVsDecisionPriority verifies that ExitCode and Decision can coexist
// on the same HookOutput. The CLI checks ExitCode first (process-level), then
// Decision (JSON-level). Both are preserved internally, but only Decision appears
// in the serialized JSON because ExitCode uses json:"-".
func TestExitCodeVsDecisionPriority(t *testing.T) {
	t.Parallel()

	output := &HookOutput{ExitCode: 2, Decision: DecisionAllow}

	// Both fields are preserved in memory
	if output.ExitCode != 2 {
		t.Errorf("ExitCode = %d, want 2", output.ExitCode)
	}
	if output.Decision != DecisionAllow {
		t.Errorf("Decision = %q, want %q", output.Decision, DecisionAllow)
	}

	// Serialize to JSON
	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	s := string(data)

	// JSON should contain "decision" but NOT "exitCode"
	if !strings.Contains(s, `"decision"`) {
		t.Errorf("JSON should contain \"decision\": %s", s)
	}
	if strings.Contains(s, "exitCode") || strings.Contains(s, "ExitCode") {
		t.Errorf("JSON should NOT contain exitCode: %s", s)
	}

	// Verify the deserialized Decision is correct
	var parsed HookOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if parsed.Decision != DecisionAllow {
		t.Errorf("parsed Decision = %q, want %q", parsed.Decision, DecisionAllow)
	}
	// ExitCode is lost after JSON round-trip (json:"-")
	if parsed.ExitCode != 0 {
		t.Errorf("parsed ExitCode = %d, want 0 (lost in JSON round-trip)", parsed.ExitCode)
	}
}
