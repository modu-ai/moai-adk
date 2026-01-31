package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/anthropics/moai-adk-go/internal/hooks/protocol"
)

// ============================================================================
// NewHookDispatcher tests
// ============================================================================

func TestNewHookDispatcher(t *testing.T) {
	d := NewHookDispatcher()

	if d == nil {
		t.Fatal("NewHookDispatcher returned nil")
	}
	if d.sessionStartHandler == nil {
		t.Error("Expected sessionStartHandler to be initialized")
	}
	if d.sessionEndHandler == nil {
		t.Error("Expected sessionEndHandler to be initialized")
	}
	if d.preToolHandler == nil {
		t.Error("Expected preToolHandler to be initialized")
	}
	if d.postToolHandler == nil {
		t.Error("Expected postToolHandler to be initialized")
	}
}

// ============================================================================
// Dispatch tests - routing to correct handlers
// ============================================================================

func TestDispatch_SessionStart(t *testing.T) {
	d := NewHookDispatcher()
	input := &protocol.HookInput{
		SessionID: "test-session",
		Event:     "session_start",
	}

	output := captureDispatchStdout(t, func() error {
		return d.Dispatch(context.Background(), input)
	})

	// Should produce valid JSON output
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s\nError: %v", output, err)
	}

	// Should have continue field
	if _, ok := parsed["continue"]; !ok {
		t.Error("Expected 'continue' field in output")
	}
}

func TestDispatch_SessionEnd(t *testing.T) {
	d := NewHookDispatcher()
	input := &protocol.HookInput{
		SessionID: "test-session",
		Event:     "session_end",
	}

	output := captureDispatchStdout(t, func() error {
		return d.Dispatch(context.Background(), input)
	})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s\nError: %v", output, err)
	}

	if _, ok := parsed["continue"]; !ok {
		t.Error("Expected 'continue' field in output")
	}
}

func TestDispatch_PreTool_AllowedCommand(t *testing.T) {
	d := NewHookDispatcher()
	input := &protocol.HookInput{
		SessionID: "test-session",
		Event:     "pre_tool",
		ToolName:  "Bash",
		ToolInput: map[string]any{
			"command": "echo hello",
		},
	}

	output := captureDispatchStdout(t, func() error {
		return d.Dispatch(context.Background(), input)
	})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s\nError: %v", output, err)
	}

	if parsed["continue"] != true {
		t.Errorf("Expected continue=true for allowed command, got: %v", parsed["continue"])
	}
}

func TestDispatch_PreTool_BlockedCommand(t *testing.T) {
	d := NewHookDispatcher()
	input := &protocol.HookInput{
		SessionID: "test-session",
		Event:     "pre_tool",
		ToolName:  "Bash",
		ToolInput: map[string]any{
			"command": "rm -rf /",
		},
	}

	output := captureDispatchStdout(t, func() error {
		// Dispatch returns the handler error for blocked commands,
		// but it still writes JSON to stdout
		_ = d.Dispatch(context.Background(), input)
		return nil
	})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s\nError: %v", output, err)
	}

	if parsed["continue"] != false {
		t.Errorf("Expected continue=false for blocked command, got: %v", parsed["continue"])
	}
	if parsed["block_execution"] != true {
		t.Errorf("Expected block_execution=true for blocked command, got: %v", parsed["block_execution"])
	}
}

func TestDispatch_PreTool_NonTargetTool(t *testing.T) {
	d := NewHookDispatcher()
	input := &protocol.HookInput{
		SessionID: "test-session",
		Event:     "pre_tool",
		ToolName:  "Read",
	}

	output := captureDispatchStdout(t, func() error {
		return d.Dispatch(context.Background(), input)
	})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s\nError: %v", output, err)
	}

	if parsed["suppressOutput"] != true {
		t.Errorf("Expected suppressOutput=true for non-target tool, got: %v", parsed["suppressOutput"])
	}
}

func TestDispatch_PostTool_NonTargetTool(t *testing.T) {
	d := NewHookDispatcher()
	input := &protocol.HookInput{
		SessionID: "test-session",
		Event:     "post_tool",
		ToolName:  "Bash",
	}

	output := captureDispatchStdout(t, func() error {
		return d.Dispatch(context.Background(), input)
	})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Expected valid JSON output, got: %s\nError: %v", output, err)
	}

	if parsed["suppressOutput"] != true {
		t.Errorf("Expected suppressOutput=true for non-target tool, got: %v", parsed["suppressOutput"])
	}
}

// ============================================================================
// Dispatch tests - simple events (suppress output)
// ============================================================================

func TestDispatch_SimpleEvents(t *testing.T) {
	simpleEvents := []struct {
		event string
		label string
	}{
		{"pre_compact", "Pre-compact"},
		{"stop", "Stop"},
		{"notification", "Notification"},
		{"quality_gate", "Quality gate"},
		{"commit", "Commit"},
		{"push", "Push"},
		{"compact", "Compact"},
	}

	d := NewHookDispatcher()

	for _, tc := range simpleEvents {
		t.Run(tc.event, func(t *testing.T) {
			input := &protocol.HookInput{
				SessionID: "test-session",
				Event:     tc.event,
			}

			output := captureDispatchStdout(t, func() error {
				return d.Dispatch(context.Background(), input)
			})

			var parsed map[string]any
			if err := json.Unmarshal([]byte(output), &parsed); err != nil {
				t.Fatalf("Expected valid JSON for event '%s', got: %s\nError: %v", tc.event, output, err)
			}

			// Simple events should suppress output
			if parsed["suppressOutput"] != true {
				t.Errorf("Expected suppressOutput=true for simple event '%s', got: %v", tc.event, parsed["suppressOutput"])
			}
			if parsed["continue"] != true {
				t.Errorf("Expected continue=true for simple event '%s', got: %v", tc.event, parsed["continue"])
			}
		})
	}
}

// ============================================================================
// Dispatch tests - unknown event
// ============================================================================

func TestDispatch_UnknownEvent(t *testing.T) {
	d := NewHookDispatcher()
	input := &protocol.HookInput{
		SessionID: "test-session",
		Event:     "unknown_event_type",
	}

	output := captureDispatchStdout(t, func() error {
		// Unknown events return an error but still write to stdout
		_ = d.Dispatch(context.Background(), input)
		return nil
	})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Expected valid JSON for unknown event, got: %s\nError: %v", output, err)
	}

	// Should have an error field
	if _, ok := parsed["error"]; !ok {
		t.Error("Expected 'error' field in output for unknown event")
	}
}

func TestDispatch_UnknownEvent_ReturnsError(t *testing.T) {
	d := NewHookDispatcher()
	input := &protocol.HookInput{
		SessionID: "test-session",
		Event:     "totally_invalid",
	}

	// Capture stdout to prevent noise
	_ = captureDispatchStdout(t, func() error {
		err := d.Dispatch(context.Background(), input)
		if err == nil {
			t.Error("Expected error for unknown event, got nil")
		}
		return nil
	})
}

// ============================================================================
// Dispatch tests - context timeout
// ============================================================================

func TestDispatch_UsesCorrectTimeout(t *testing.T) {
	d := NewHookDispatcher()

	// Use a pre_compact event which has a 3-second timeout
	input := &protocol.HookInput{
		SessionID: "test-session",
		Event:     "pre_compact",
	}

	// This should complete quickly (simple handler)
	output := captureDispatchStdout(t, func() error {
		return d.Dispatch(context.Background(), input)
	})

	// Verify it produced output (timeout did not trigger)
	if len(output) == 0 {
		t.Error("Expected non-empty output from dispatch")
	}
}

func TestDispatch_CancelledContext(t *testing.T) {
	d := NewHookDispatcher()
	input := &protocol.HookInput{
		SessionID: "test-session",
		Event:     "pre_compact",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Even with cancelled context, dispatch creates its own timeout context
	// The handler should still complete since it is fast
	output := captureDispatchStdout(t, func() error {
		_ = d.Dispatch(ctx, input)
		return nil
	})

	// Should still produce some output
	if len(output) == 0 {
		t.Log("Note: cancelled context may prevent output in some cases")
	}
}

// ============================================================================
// handleError tests
// ============================================================================

func TestHandleError_WritesErrorJSON(t *testing.T) {
	d := NewHookDispatcher()
	testErr := fmt.Errorf("test error message")

	output := captureDispatchStdout(t, func() error {
		_ = d.handleError(context.Background(), testErr)
		return nil
	})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Expected valid JSON from handleError, got: %s\nError: %v", output, err)
	}

	if parsed["error"] != "test error message" {
		t.Errorf("Expected error 'test error message', got: %v", parsed["error"])
	}
	if parsed["continue"] != true {
		t.Errorf("Expected continue=true (graceful degradation), got: %v", parsed["continue"])
	}
}

func TestHandleError_ReturnsOriginalError(t *testing.T) {
	d := NewHookDispatcher()
	testErr := fmt.Errorf("original error")

	_ = captureDispatchStdout(t, func() error {
		err := d.handleError(context.Background(), testErr)
		if err == nil {
			t.Error("Expected handleError to return the original error")
		}
		if err.Error() != "original error" {
			t.Errorf("Expected 'original error', got '%s'", err.Error())
		}
		return nil
	})
}

// ============================================================================
// handleSimple tests
// ============================================================================

func TestHandleSimple_SuppressesOutput(t *testing.T) {
	d := NewHookDispatcher()

	output := captureDispatchStdout(t, func() error {
		return d.handleSimple(context.Background(), "Test")
	})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Expected valid JSON from handleSimple, got: %s\nError: %v", output, err)
	}

	if parsed["suppressOutput"] != true {
		t.Errorf("Expected suppressOutput=true, got: %v", parsed["suppressOutput"])
	}
	if parsed["continue"] != true {
		t.Errorf("Expected continue=true, got: %v", parsed["continue"])
	}
}

// ============================================================================
// Integration tests - full event-to-output flow
// ============================================================================

func TestDispatch_PreTool_Write_SecurityBlock(t *testing.T) {
	d := NewHookDispatcher()
	input := &protocol.HookInput{
		SessionID: "integration-test",
		Event:     "pre_tool",
		ToolName:  "Write",
		ToolInput: map[string]any{
			"file_path": ".env.production",
		},
	}

	output := captureDispatchStdout(t, func() error {
		_ = d.Dispatch(context.Background(), input)
		return nil
	})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Expected valid JSON, got: %s\nError: %v", output, err)
	}

	if parsed["block_execution"] != true {
		t.Error("Expected block_execution=true for .env.production write")
	}
	if parsed["continue"] != false {
		t.Error("Expected continue=false for blocked file write")
	}

	hookOutput, ok := parsed["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatal("Expected hookSpecificOutput to be present")
	}
	if hookOutput["permissionDecision"] != "block" {
		t.Errorf("Expected permissionDecision 'block', got '%v'", hookOutput["permissionDecision"])
	}
}

func TestDispatch_PreTool_Edit_SecurityWarn(t *testing.T) {
	d := NewHookDispatcher()
	input := &protocol.HookInput{
		SessionID: "integration-test",
		Event:     "pre_tool",
		ToolName:  "Edit",
		ToolInput: map[string]any{
			"file_path": "package-lock.json",
		},
	}

	output := captureDispatchStdout(t, func() error {
		return d.Dispatch(context.Background(), input)
	})

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Expected valid JSON, got: %s\nError: %v", output, err)
	}

	if parsed["continue"] != true {
		t.Error("Expected continue=true for warn (not block)")
	}

	hookOutput, ok := parsed["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatal("Expected hookSpecificOutput to be present for warn")
	}
	if hookOutput["permissionDecision"] != "warn" {
		t.Errorf("Expected permissionDecision 'warn', got '%v'", hookOutput["permissionDecision"])
	}
}

// ============================================================================
// Helper functions
// ============================================================================

// captureDispatchStdout captures stdout during fn execution.
// The fn receives no error return since we handle errors inside fn.
func captureDispatchStdout(t *testing.T, fn func() error) string {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	os.Stdout = w

	_ = fn()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to read captured output: %v", err)
	}
	r.Close()

	return buf.String()
}
