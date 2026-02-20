package hook

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestReadInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantErr   bool
		errTarget error
		check     func(t *testing.T, got *HookInput)
	}{
		{
			name: "valid SessionStart input",
			input: `{
				"session_id": "sess-abc-123",
				"cwd": "/Users/goos/project",
				"hook_event_name": "SessionStart",
				"project_dir": "/Users/goos/project"
			}`,
			wantErr: false,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.SessionID != "sess-abc-123" {
					t.Errorf("SessionID = %q, want %q", got.SessionID, "sess-abc-123")
				}
				if got.CWD != "/Users/goos/project" {
					t.Errorf("CWD = %q, want %q", got.CWD, "/Users/goos/project")
				}
				if got.HookEventName != "SessionStart" {
					t.Errorf("HookEventName = %q, want %q", got.HookEventName, "SessionStart")
				}
				if got.ProjectDir != "/Users/goos/project" {
					t.Errorf("ProjectDir = %q, want %q", got.ProjectDir, "/Users/goos/project")
				}
			},
		},
		{
			name: "valid PreToolUse input with tool_name and tool_input",
			input: `{
				"session_id": "sess-abc-123",
				"cwd": "/Users/goos/project",
				"hook_event_name": "PreToolUse",
				"tool_name": "Write",
				"tool_input": {"file_path": "/tmp/test.go", "content": "package main"}
			}`,
			wantErr: false,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.ToolName != "Write" {
					t.Errorf("ToolName = %q, want %q", got.ToolName, "Write")
				}
				if got.ToolInput == nil {
					t.Fatal("ToolInput is nil, want non-nil")
				}
				if !json.Valid(got.ToolInput) {
					t.Errorf("ToolInput is not valid JSON: %s", got.ToolInput)
				}
			},
		},
		{
			name: "valid PostToolUse input with tool_output",
			input: `{
				"session_id": "sess-abc-123",
				"cwd": "/Users/goos/project",
				"hook_event_name": "PostToolUse",
				"tool_name": "Write",
				"tool_input": {"file_path": "main.go"},
				"tool_output": {"success": true, "path": "main.go"}
			}`,
			wantErr: false,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.ToolOutput == nil {
					t.Fatal("ToolOutput is nil, want non-nil")
				}
				if !json.Valid(got.ToolOutput) {
					t.Errorf("ToolOutput is not valid JSON: %s", got.ToolOutput)
				}
			},
		},
		{
			name:      "missing session_id",
			input:     `{"cwd": "/tmp", "hook_event_name": "SessionStart"}`,
			wantErr:   true,
			errTarget: ErrHookInvalidInput,
		},
		{
			name:      "missing cwd",
			input:     `{"session_id": "sess-1", "hook_event_name": "SessionStart"}`,
			wantErr:   true,
			errTarget: ErrHookInvalidInput,
		},
		{
			name:      "missing hook_event_name",
			input:     `{"session_id": "sess-1", "cwd": "/tmp"}`,
			wantErr:   true,
			errTarget: ErrHookInvalidInput,
		},
		{
			name:      "empty JSON object",
			input:     `{}`,
			wantErr:   true,
			errTarget: ErrHookInvalidInput,
		},
		{
			name:      "malformed JSON",
			input:     `this is not json at all`,
			wantErr:   true,
			errTarget: ErrHookInvalidInput,
		},
		{
			name:      "empty stdin",
			input:     ``,
			wantErr:   true,
			errTarget: ErrHookInvalidInput,
		},
		{
			name:      "whitespace only stdin",
			input:     `   `,
			wantErr:   true,
			errTarget: ErrHookInvalidInput,
		},
		{
			name: "extra unknown fields are ignored",
			input: `{
				"session_id": "sess-1",
				"cwd": "/tmp",
				"hook_event_name": "SessionStart",
				"unknown_field": "should be ignored",
				"another_field": 42
			}`,
			wantErr: false,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.SessionID != "sess-1" {
					t.Errorf("SessionID = %q, want %q", got.SessionID, "sess-1")
				}
			},
		},
		{
			name:    "large tool_output payload",
			input:   buildLargePayload(t),
			wantErr: false,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.ToolOutput == nil {
					t.Fatal("ToolOutput is nil for large payload")
				}
			},
		},
	}

	proto := &jsonProtocol{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := strings.NewReader(tt.input)
			got, err := proto.ReadInput(r)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errTarget != nil && !errors.Is(err, tt.errTarget) {
					t.Errorf("error = %v, want errors.Is(%v)", err, tt.errTarget)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil HookInput, want non-nil")
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestWriteOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		output  *HookOutput
		wantErr bool
		check   func(t *testing.T, written []byte)
	}{
		{
			name:   "allow output",
			output: NewAllowOutput(),
			check: func(t *testing.T, written []byte) {
				t.Helper()
				if !json.Valid(written) {
					t.Fatalf("output is not valid JSON: %s", written)
				}
				var parsed HookOutput
				if err := json.Unmarshal(written, &parsed); err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}
				// Check hookSpecificOutput.permissionDecision per Claude Code protocol
				if parsed.HookSpecificOutput == nil {
					t.Fatal("HookSpecificOutput is nil, want non-nil")
				}
				if parsed.HookSpecificOutput.PermissionDecision != DecisionAllow {
					t.Errorf("PermissionDecision = %q, want %q", parsed.HookSpecificOutput.PermissionDecision, DecisionAllow)
				}
			},
		},
		{
			name:   "block output with reason",
			output: NewBlockOutput("security policy violation"),
			check: func(t *testing.T, written []byte) {
				t.Helper()
				var parsed HookOutput
				if err := json.Unmarshal(written, &parsed); err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}
				// Check hookSpecificOutput per Claude Code protocol
				if parsed.HookSpecificOutput == nil {
					t.Fatal("HookSpecificOutput is nil, want non-nil")
				}
				if parsed.HookSpecificOutput.PermissionDecision != DecisionDeny {
					t.Errorf("PermissionDecision = %q, want %q", parsed.HookSpecificOutput.PermissionDecision, DecisionDeny)
				}
				if parsed.HookSpecificOutput.PermissionDecisionReason != "security policy violation" {
					t.Errorf("PermissionDecisionReason = %q, want %q", parsed.HookSpecificOutput.PermissionDecisionReason, "security policy violation")
				}
			},
		},
		{
			name:   "allow output with data",
			output: NewAllowOutputWithData(json.RawMessage(`{"project":"moai","version":"1.0.0"}`)),
			check: func(t *testing.T, written []byte) {
				t.Helper()
				var parsed HookOutput
				if err := json.Unmarshal(written, &parsed); err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}
				// Check hookSpecificOutput per Claude Code protocol
				if parsed.HookSpecificOutput == nil {
					t.Fatal("HookSpecificOutput is nil, want non-nil")
				}
				if parsed.HookSpecificOutput.PermissionDecision != DecisionAllow {
					t.Errorf("PermissionDecision = %q, want %q", parsed.HookSpecificOutput.PermissionDecision, DecisionAllow)
				}
			},
		},
		{
			name:   "empty output",
			output: &HookOutput{},
			check: func(t *testing.T, written []byte) {
				t.Helper()
				if !json.Valid(written) {
					t.Fatalf("output is not valid JSON: %s", written)
				}
			},
		},
		{
			name:   "nil output",
			output: nil,
			check: func(t *testing.T, written []byte) {
				t.Helper()
				if !json.Valid(written) {
					t.Fatalf("output is not valid JSON: %s", written)
				}
			},
		},
		{
			name:   "suppress output",
			output: NewSuppressOutput(),
			check: func(t *testing.T, written []byte) {
				t.Helper()
				var parsed HookOutput
				if err := json.Unmarshal(written, &parsed); err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}
				if !parsed.SuppressOutput {
					t.Error("SuppressOutput = false, want true")
				}
			},
		},
		{
			name:   "session output",
			output: NewSessionOutput(true, "Session started"),
			check: func(t *testing.T, written []byte) {
				t.Helper()
				var parsed HookOutput
				if err := json.Unmarshal(written, &parsed); err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}
				if !parsed.Continue {
					t.Error("Continue = false, want true")
				}
				if parsed.SystemMessage != "Session started" {
					t.Errorf("SystemMessage = %q, want %q", parsed.SystemMessage, "Session started")
				}
			},
		},
		{
			name:   "post tool output",
			output: NewPostToolOutput("Formatted with gofmt"),
			check: func(t *testing.T, written []byte) {
				t.Helper()
				var parsed HookOutput
				if err := json.Unmarshal(written, &parsed); err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}
				if parsed.HookSpecificOutput == nil {
					t.Fatal("HookSpecificOutput is nil, want non-nil")
				}
				if parsed.HookSpecificOutput.AdditionalContext != "Formatted with gofmt" {
					t.Errorf("AdditionalContext = %q, want %q", parsed.HookSpecificOutput.AdditionalContext, "Formatted with gofmt")
				}
			},
		},
	}

	proto := &jsonProtocol{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := proto.WriteOutput(&buf, tt.output)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, buf.Bytes())
			}
		})
	}
}

func TestJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := &HookOutput{
		Decision: DecisionAllow,
		Reason:   "test reason",
		Data:     json.RawMessage(`{"key":"value","nested":{"a":1}}`),
	}

	// First serialization
	first, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("first marshal: %v", err)
	}
	if !json.Valid(first) {
		t.Fatalf("first marshal produced invalid JSON: %s", first)
	}

	// Deserialize
	var parsed HookOutput
	if err := json.Unmarshal(first, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Re-serialize
	second, err := json.Marshal(parsed)
	if err != nil {
		t.Fatalf("second marshal: %v", err)
	}
	if !json.Valid(second) {
		t.Fatalf("second marshal produced invalid JSON: %s", second)
	}

	// Compare
	if !bytes.Equal(first, second) {
		t.Errorf("round-trip mismatch:\n  first:  %s\n  second: %s", first, second)
	}
}

func TestWriteOutputNoExtraneousContent(t *testing.T) {
	t.Parallel()

	proto := &jsonProtocol{}
	var buf bytes.Buffer

	err := proto.WriteOutput(&buf, NewAllowOutput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Output should be valid JSON followed by a newline only
	output := buf.String()
	trimmed := strings.TrimSpace(output)
	if !json.Valid([]byte(trimmed)) {
		t.Errorf("trimmed output is not valid JSON: %q", trimmed)
	}
}

// buildLargePayload creates a valid hook JSON with a large tool_output field.
func buildLargePayload(t *testing.T) string {
	t.Helper()

	// Create ~100KB of tool output data
	largeData := make(map[string]string)
	for i := range 1000 {
		key := strings.Repeat("k", 10)
		val := strings.Repeat("v", 90)
		largeData[key+string(rune('0'+i%10))] = val
	}

	toolOutput, err := json.Marshal(largeData)
	if err != nil {
		t.Fatalf("marshal large data: %v", err)
	}

	input := map[string]any{
		"session_id":      "sess-large",
		"cwd":             "/tmp",
		"hook_event_name": "PostToolUse",
		"tool_name":       "Write",
		"tool_output":     json.RawMessage(toolOutput),
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal large payload: %v", err)
	}
	return string(data)
}
