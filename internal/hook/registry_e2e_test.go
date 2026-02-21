package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

// TestRegistryDispatch_AllNewEventTypes verifies that all 7 new event types
// dispatch correctly through real handlers registered in the registry.
// Each handler is constructed via its exported constructor (not mocks).
func TestRegistryDispatch_AllNewEventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		handler           Handler
		event             EventType
		input             *HookInput
		wantNilHSO        bool   // expect HookSpecificOutput to be nil
		wantPermDecision  string // expected permissionDecision (if HSO present)
		wantExitCode      int
		wantHookEventName string // expected hookEventName in HSO (if present)
	}{
		{
			name:    "PostToolUseFailure dispatches with real handler",
			handler: NewPostToolUseFailureHandler(),
			event:   EventPostToolUseFailure,
			input: &HookInput{
				SessionID:     "sess-e2e-1",
				CWD:           "/tmp",
				HookEventName: "PostToolUseFailure",
				ToolName:      "Bash",
				ToolUseID:     "tu-123",
				Error:         "command not found",
				IsInterrupt:   false,
			},
			wantNilHSO: true,
		},
		{
			name:    "Notification dispatches with real handler",
			handler: NewNotificationHandler(),
			event:   EventNotification,
			input: &HookInput{
				SessionID:        "sess-e2e-2",
				CWD:              "/tmp",
				HookEventName:    "Notification",
				Title:            "Build Complete",
				Message:          "All tests pass",
				NotificationType: "info",
			},
			wantNilHSO: true,
		},
		{
			name:    "SubagentStart dispatches with real handler",
			handler: NewSubagentStartHandler(),
			event:   EventSubagentStart,
			input: &HookInput{
				SessionID:           "sess-e2e-3",
				CWD:                 "/tmp",
				HookEventName:       "SubagentStart",
				AgentID:             "agent-backend-1",
				AgentTranscriptPath: "/tmp/transcript.jsonl",
			},
			wantNilHSO: true,
		},
		{
			name:    "UserPromptSubmit dispatches with real handler",
			handler: NewUserPromptSubmitHandler(),
			event:   EventUserPromptSubmit,
			input: &HookInput{
				SessionID:     "sess-e2e-4",
				CWD:           "/tmp",
				HookEventName: "UserPromptSubmit",
				Prompt:        "Help me fix this bug",
			},
			wantNilHSO: true,
		},
		{
			name:    "PermissionRequest dispatches with real handler and returns ask",
			handler: NewPermissionRequestHandler(),
			event:   EventPermissionRequest,
			input: &HookInput{
				SessionID:     "sess-e2e-5",
				CWD:           "/tmp",
				HookEventName: "PermissionRequest",
				ToolName:      "Bash",
			},
			wantNilHSO:        false,
			wantPermDecision:  DecisionAsk,
			wantHookEventName: "PreToolUse",
		},
		{
			name:    "TeammateIdle dispatches with real handler and exits 0",
			handler: NewTeammateIdleHandler(),
			event:   EventTeammateIdle,
			input: &HookInput{
				SessionID:     "sess-e2e-6",
				CWD:           "/tmp",
				HookEventName: "TeammateIdle",
				AgentID:       "teammate-researcher",
			},
			wantNilHSO:   true,
			wantExitCode: 0,
		},
		{
			name:    "TaskCompleted dispatches with real handler and exits 0",
			handler: NewTaskCompletedHandler(),
			event:   EventTaskCompleted,
			input: &HookInput{
				SessionID:     "sess-e2e-7",
				CWD:           "/tmp",
				HookEventName: "TaskCompleted",
			},
			wantNilHSO:   true,
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfigProvider{cfg: newTestConfig()}
			reg := NewRegistry(cfg)
			reg.Register(tt.handler)

			ctx := context.Background()
			got, err := reg.Dispatch(ctx, tt.event, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output, want non-nil")
			}

			// Verify HookSpecificOutput presence
			if tt.wantNilHSO {
				if got.HookSpecificOutput != nil {
					t.Errorf("HookSpecificOutput = %+v, want nil", got.HookSpecificOutput)
				}
			} else {
				if got.HookSpecificOutput == nil {
					t.Fatal("HookSpecificOutput is nil, want non-nil")
				}
				if tt.wantPermDecision != "" && got.HookSpecificOutput.PermissionDecision != tt.wantPermDecision {
					t.Errorf("PermissionDecision = %q, want %q",
						got.HookSpecificOutput.PermissionDecision, tt.wantPermDecision)
				}
				if tt.wantHookEventName != "" && got.HookSpecificOutput.HookEventName != tt.wantHookEventName {
					t.Errorf("HookEventName = %q, want %q",
						got.HookSpecificOutput.HookEventName, tt.wantHookEventName)
				}
			}

			// Verify exit code
			if got.ExitCode != tt.wantExitCode {
				t.Errorf("ExitCode = %d, want %d", got.ExitCode, tt.wantExitCode)
			}
		})
	}
}

// TestRegistryDispatch_NewEventDefaults verifies that unregistered new events
// return correct default outputs via defaultOutputForEvent().
func TestRegistryDispatch_NewEventDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		event             EventType
		wantNilHSO        bool
		wantPermDecision  string
		wantHookEventName string
	}{
		{
			name:       "PostToolUseFailure default is empty output",
			event:      EventPostToolUseFailure,
			wantNilHSO: true,
		},
		{
			name:       "Notification default is empty output",
			event:      EventNotification,
			wantNilHSO: true,
		},
		{
			name:       "SubagentStart default is empty output",
			event:      EventSubagentStart,
			wantNilHSO: true,
		},
		{
			name:       "UserPromptSubmit default is empty output",
			event:      EventUserPromptSubmit,
			wantNilHSO: true,
		},
		{
			name:              "PermissionRequest default is ask",
			event:             EventPermissionRequest,
			wantNilHSO:        false,
			wantPermDecision:  DecisionAsk,
			wantHookEventName: "PreToolUse",
		},
		{
			name:       "TeammateIdle default is empty output",
			event:      EventTeammateIdle,
			wantNilHSO: true,
		},
		{
			name:       "TaskCompleted default is empty output",
			event:      EventTaskCompleted,
			wantNilHSO: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfigProvider{cfg: newTestConfig()}
			reg := NewRegistry(cfg)

			// Register only a SessionStart handler to prove it has no effect on new events
			reg.Register(&mockHandler{
				event:  EventSessionStart,
				output: &HookOutput{},
			})

			ctx := context.Background()
			input := &HookInput{
				SessionID:     "sess-default",
				CWD:           "/tmp",
				HookEventName: string(tt.event),
			}

			got, err := reg.Dispatch(ctx, tt.event, input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output, want non-nil")
			}

			if tt.wantNilHSO {
				if got.HookSpecificOutput != nil {
					t.Errorf("HookSpecificOutput = %+v, want nil for default %s", got.HookSpecificOutput, tt.event)
				}
			} else {
				if got.HookSpecificOutput == nil {
					t.Fatalf("HookSpecificOutput is nil for default %s, want non-nil", tt.event)
				}
				if got.HookSpecificOutput.PermissionDecision != tt.wantPermDecision {
					t.Errorf("PermissionDecision = %q, want %q for default %s",
						got.HookSpecificOutput.PermissionDecision, tt.wantPermDecision, tt.event)
				}
				if got.HookSpecificOutput.HookEventName != tt.wantHookEventName {
					t.Errorf("HookEventName = %q, want %q for default %s",
						got.HookSpecificOutput.HookEventName, tt.wantHookEventName, tt.event)
				}
			}
		})
	}
}

// TestRegistryDispatch_PermissionRequestBlockChain verifies that a blocking
// (deny) decision for PermissionRequest propagates correctly through the
// dispatch chain, short-circuiting subsequent handlers.
func TestRegistryDispatch_PermissionRequestBlockChain(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	reg := NewRegistry(cfg)

	// Register a mock handler that returns deny for PermissionRequest
	denyHandler := &mockHandler{
		event: EventPermissionRequest,
		output: &HookOutput{
			HookSpecificOutput: &HookSpecificOutput{
				HookEventName:            "PreToolUse",
				PermissionDecision:       DecisionDeny,
				PermissionDecisionReason: "security policy violation",
			},
		},
	}
	// Register a second handler that should NOT be called
	secondHandler := &mockHandler{
		event:  EventPermissionRequest,
		output: &HookOutput{},
	}

	reg.Register(denyHandler)
	reg.Register(secondHandler)

	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-perm-block",
		CWD:           "/tmp",
		HookEventName: "PermissionRequest",
		ToolName:      "Write",
	}

	got, err := reg.Dispatch(ctx, EventPermissionRequest, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output, want non-nil")
	}

	// Verify deny decision propagated
	if got.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput is nil, want deny decision")
	}
	if got.HookSpecificOutput.PermissionDecision != DecisionDeny {
		t.Errorf("PermissionDecision = %q, want %q",
			got.HookSpecificOutput.PermissionDecision, DecisionDeny)
	}
	if got.HookSpecificOutput.PermissionDecisionReason != "security policy violation" {
		t.Errorf("PermissionDecisionReason = %q, want %q",
			got.HookSpecificOutput.PermissionDecisionReason, "security policy violation")
	}

	// Verify short-circuit: first handler called, second not called
	if !denyHandler.called {
		t.Error("deny handler should have been called")
	}
	if secondHandler.called {
		t.Error("second handler should NOT have been called after deny")
	}
}

// TestRegistryDispatch_MultipleHandlersNewEvents verifies that multiple
// handlers for new events execute in sequence (when none blocks).
func TestRegistryDispatch_MultipleHandlersNewEvents(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	reg := NewRegistry(cfg)

	// Register a mock handler that returns empty output (no block)
	mockH := &mockHandler{
		event:  EventPostToolUseFailure,
		output: &HookOutput{},
	}
	// Register the real handler
	realH := NewPostToolUseFailureHandler()

	reg.Register(mockH)
	reg.Register(realH)

	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-multi",
		CWD:           "/tmp",
		HookEventName: "PostToolUseFailure",
		ToolName:      "Bash",
		ToolUseID:     "tu-multi-1",
		Error:         "permission denied",
		IsInterrupt:   false,
	}

	got, err := reg.Dispatch(ctx, EventPostToolUseFailure, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output, want non-nil")
	}

	// Verify mock handler was called (both handlers should execute)
	if !mockH.called {
		t.Error("mock handler should have been called")
	}

	// The final output comes from defaultOutputForEvent since no handler blocked.
	// PostToolUseFailure default is empty HookOutput.
	if got.HookSpecificOutput != nil {
		t.Errorf("HookSpecificOutput = %+v, want nil for PostToolUseFailure", got.HookSpecificOutput)
	}
}

// TestRegistryDispatch_FullPipeline_JSONRoundTrip tests the complete dispatch
// pipeline for each new event: create input, register real handler, dispatch,
// serialize output to JSON via protocol.WriteOutput, deserialize back, and
// verify round-trip fidelity.
func TestRegistryDispatch_FullPipeline_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		handler           Handler
		event             EventType
		input             *HookInput
		wantNilHSO        bool
		wantPermDecision  string
		wantHookEventName string
	}{
		{
			name:    "PostToolUseFailure round-trip",
			handler: NewPostToolUseFailureHandler(),
			event:   EventPostToolUseFailure,
			input: &HookInput{
				SessionID:     "sess-e2e-1",
				CWD:           "/tmp",
				HookEventName: "PostToolUseFailure",
				ToolName:      "Bash",
				ToolUseID:     "tu-123",
				Error:         "command not found",
				IsInterrupt:   false,
			},
			wantNilHSO: true,
		},
		{
			name:    "Notification round-trip",
			handler: NewNotificationHandler(),
			event:   EventNotification,
			input: &HookInput{
				SessionID:        "sess-e2e-2",
				CWD:              "/tmp",
				HookEventName:    "Notification",
				Title:            "Build Complete",
				Message:          "All tests pass",
				NotificationType: "info",
			},
			wantNilHSO: true,
		},
		{
			name:    "SubagentStart round-trip",
			handler: NewSubagentStartHandler(),
			event:   EventSubagentStart,
			input: &HookInput{
				SessionID:           "sess-e2e-3",
				CWD:                 "/tmp",
				HookEventName:       "SubagentStart",
				AgentID:             "agent-backend-1",
				AgentTranscriptPath: "/tmp/transcript.jsonl",
			},
			wantNilHSO: true,
		},
		{
			name:    "UserPromptSubmit round-trip",
			handler: NewUserPromptSubmitHandler(),
			event:   EventUserPromptSubmit,
			input: &HookInput{
				SessionID:     "sess-e2e-4",
				CWD:           "/tmp",
				HookEventName: "UserPromptSubmit",
				Prompt:        "Help me fix this bug",
			},
			wantNilHSO: true,
		},
		{
			name:    "PermissionRequest round-trip",
			handler: NewPermissionRequestHandler(),
			event:   EventPermissionRequest,
			input: &HookInput{
				SessionID:     "sess-e2e-5",
				CWD:           "/tmp",
				HookEventName: "PermissionRequest",
				ToolName:      "Bash",
			},
			wantNilHSO:        false,
			wantPermDecision:  DecisionAsk,
			wantHookEventName: "PreToolUse",
		},
		{
			name:    "TeammateIdle round-trip",
			handler: NewTeammateIdleHandler(),
			event:   EventTeammateIdle,
			input: &HookInput{
				SessionID:     "sess-e2e-6",
				CWD:           "/tmp",
				HookEventName: "TeammateIdle",
				AgentID:       "teammate-researcher",
			},
			wantNilHSO: true,
		},
		{
			name:    "TaskCompleted round-trip",
			handler: NewTaskCompletedHandler(),
			event:   EventTaskCompleted,
			input: &HookInput{
				SessionID:     "sess-e2e-7",
				CWD:           "/tmp",
				HookEventName: "TaskCompleted",
			},
			wantNilHSO: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Step 1: Create registry and register real handler
			cfg := &mockConfigProvider{cfg: newTestConfig()}
			reg := NewRegistry(cfg)
			reg.Register(tt.handler)

			// Step 2: Dispatch through registry
			ctx := context.Background()
			got, err := reg.Dispatch(ctx, tt.event, tt.input)
			if err != nil {
				t.Fatalf("dispatch error: %v", err)
			}
			if got == nil {
				t.Fatal("dispatch returned nil output")
			}

			// Step 3: Serialize output to JSON via protocol.WriteOutput
			proto := NewProtocol()
			var buf bytes.Buffer
			if err := proto.WriteOutput(&buf, got); err != nil {
				t.Fatalf("WriteOutput error: %v", err)
			}

			// Step 4: Deserialize JSON back to HookOutput
			var roundTripped HookOutput
			if err := json.Unmarshal(buf.Bytes(), &roundTripped); err != nil {
				t.Fatalf("json.Unmarshal error: %v (json: %s)", err, buf.String())
			}

			// Step 5: Verify round-trip fidelity
			if tt.wantNilHSO {
				if roundTripped.HookSpecificOutput != nil {
					t.Errorf("round-trip HookSpecificOutput = %+v, want nil",
						roundTripped.HookSpecificOutput)
				}
			} else {
				if roundTripped.HookSpecificOutput == nil {
					t.Fatal("round-trip HookSpecificOutput is nil, want non-nil")
				}
				if roundTripped.HookSpecificOutput.PermissionDecision != tt.wantPermDecision {
					t.Errorf("round-trip PermissionDecision = %q, want %q",
						roundTripped.HookSpecificOutput.PermissionDecision, tt.wantPermDecision)
				}
				if roundTripped.HookSpecificOutput.HookEventName != tt.wantHookEventName {
					t.Errorf("round-trip HookEventName = %q, want %q",
						roundTripped.HookSpecificOutput.HookEventName, tt.wantHookEventName)
				}
			}

			// Verify no unexpected fields leaked through
			if roundTripped.Decision != got.Decision {
				t.Errorf("round-trip Decision = %q, want %q", roundTripped.Decision, got.Decision)
			}
			if roundTripped.Reason != got.Reason {
				t.Errorf("round-trip Reason = %q, want %q", roundTripped.Reason, got.Reason)
			}
			if roundTripped.Continue != got.Continue {
				t.Errorf("round-trip Continue = %v, want %v", roundTripped.Continue, got.Continue)
			}
			if roundTripped.SuppressOutput != got.SuppressOutput {
				t.Errorf("round-trip SuppressOutput = %v, want %v",
					roundTripped.SuppressOutput, got.SuppressOutput)
			}
		})
	}
}
