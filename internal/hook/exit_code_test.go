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

// TestExitCodeBehavior_RegistryPreservesExitCode verifies the registry's behavior
// with ExitCode. The registry dispatch uses block-decision short-circuiting: if a
// handler returns a blocking decision, that output (including ExitCode) is returned
// directly. For non-blocking outputs, the registry returns defaultOutputForEvent.
//
// ExitCode=2 without a block Decision is a process-level signal (not JSON-level),
// so the registry's default output path returns ExitCode=0. The CLI layer is
// responsible for reading ExitCode from individual handler results.
func TestExitCodeBehavior_RegistryPreservesExitCode(t *testing.T) {
	t.Parallel()

	t.Run("non-blocking handler ExitCode is not preserved through dispatch default path", func(t *testing.T) {
		t.Parallel()

		// ExitCode=2 without a block Decision: registry returns defaultOutputForEvent
		handler := &mockHandler{
			event:  EventTeammateIdle,
			output: NewTeammateKeepWorkingOutput(), // ExitCode=2, no Decision
		}

		cfg := &mockConfigProvider{cfg: newTestConfig()}
		reg := NewRegistry(cfg)
		reg.Register(handler)

		input := &HookInput{
			SessionID:     "test-exit-default",
			CWD:           "/tmp",
			HookEventName: string(EventTeammateIdle),
			AgentID:       "tm-default",
		}

		got, err := reg.Dispatch(context.Background(), EventTeammateIdle, input)
		if err != nil {
			t.Fatalf("Dispatch() unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("Dispatch() returned nil output")
		}

		// Registry returns defaultOutputForEvent which has ExitCode=0
		if got.ExitCode != 0 {
			t.Errorf("ExitCode = %d, want 0 (default output path)", got.ExitCode)
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
