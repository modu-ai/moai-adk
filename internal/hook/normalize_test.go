package hook

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestReadInput_ClaudeCodeNativeFormat reproduces the bug from issue #474.
// Claude Code sends hook input in nested camelCase format (e.g. session.id,
// eventType, toolName), but the parser only accepted flat snake_case format
// (session_id, hook_event_name, tool_name), causing all hooks to fail with
// "missing required field session_id".
func TestReadInput_ClaudeCodeNativeFormat(t *testing.T) {
	t.Parallel()

	proto := &jsonProtocol{}

	tests := []struct {
		name  string
		input string
		check func(t *testing.T, got *HookInput)
	}{
		{
			name: "PreToolUse in Claude Code native camelCase format",
			input: `{
				"eventType": "PreToolUse",
				"toolName": "Bash",
				"toolInput": {"command": "echo test"},
				"session": {"id": "test-123", "cwd": "/tmp", "projectDir": "/tmp"}
			}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.HookEventName != "PreToolUse" {
					t.Errorf("HookEventName = %q, want %q", got.HookEventName, "PreToolUse")
				}
				if got.SessionID != "test-123" {
					t.Errorf("SessionID = %q, want %q", got.SessionID, "test-123")
				}
				if got.CWD != "/tmp" {
					t.Errorf("CWD = %q, want %q", got.CWD, "/tmp")
				}
				if got.ToolName != "Bash" {
					t.Errorf("ToolName = %q, want %q", got.ToolName, "Bash")
				}
				if got.ToolInput == nil {
					t.Error("ToolInput is nil, want non-nil")
				}
				if !json.Valid(got.ToolInput) {
					t.Errorf("ToolInput is not valid JSON: %s", got.ToolInput)
				}
			},
		},
		{
			name: "PostToolUse in Claude Code native camelCase format",
			input: `{
				"eventType": "PostToolUse",
				"toolName": "Write",
				"toolInput": {"file_path": "main.go", "content": "package main"},
				"toolOutput": {"success": true},
				"session": {"id": "sess-456", "cwd": "/project", "projectDir": "/project"}
			}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.HookEventName != "PostToolUse" {
					t.Errorf("HookEventName = %q, want %q", got.HookEventName, "PostToolUse")
				}
				if got.SessionID != "sess-456" {
					t.Errorf("SessionID = %q, want %q", got.SessionID, "sess-456")
				}
				if got.CWD != "/project" {
					t.Errorf("CWD = %q, want %q", got.CWD, "/project")
				}
				if got.ProjectDir != "/project" {
					t.Errorf("ProjectDir = %q, want %q", got.ProjectDir, "/project")
				}
				if got.ToolOutput == nil {
					t.Error("ToolOutput is nil, want non-nil")
				}
			},
		},
		{
			name: "SessionStart in Claude Code native camelCase format",
			input: `{
				"eventType": "SessionStart",
				"session": {"id": "sess-789", "cwd": "/home/user", "projectDir": "/home/user/project"}
			}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.HookEventName != "SessionStart" {
					t.Errorf("HookEventName = %q, want %q", got.HookEventName, "SessionStart")
				}
				if got.SessionID != "sess-789" {
					t.Errorf("SessionID = %q, want %q", got.SessionID, "sess-789")
				}
				if got.CWD != "/home/user" {
					t.Errorf("CWD = %q, want %q", got.CWD, "/home/user")
				}
				if got.ProjectDir != "/home/user/project" {
					t.Errorf("ProjectDir = %q, want %q", got.ProjectDir, "/home/user/project")
				}
			},
		},
		{
			name:  "exact input from issue #474 bug report",
			input: `{"eventType":"PreToolUse","toolName":"Bash","toolInput":{"command":"echo test"},"session":{"id":"test-123","cwd":"/tmp","projectDir":"/tmp"}}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.SessionID != "test-123" {
					t.Errorf("SessionID = %q, want %q", got.SessionID, "test-123")
				}
				if got.CWD != "/tmp" {
					t.Errorf("CWD = %q, want %q", got.CWD, "/tmp")
				}
				if got.HookEventName != "PreToolUse" {
					t.Errorf("HookEventName = %q, want %q", got.HookEventName, "PreToolUse")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := strings.NewReader(tt.input)
			got, err := proto.ReadInput(r)
			if err != nil {
				// This is the bug: native format is rejected with "missing required field session_id"
				t.Fatalf("ReadInput() returned unexpected error (issue #474): %v", err)
			}
			if got == nil {
				t.Fatal("ReadInput() returned nil HookInput")
			}
			tt.check(t, got)
		})
	}
}

// TestReadInput_BackwardCompatibility ensures flat snake_case format (the
// original format) continues to work after the normalization change.
func TestReadInput_BackwardCompatibility(t *testing.T) {
	t.Parallel()

	proto := &jsonProtocol{}

	tests := []struct {
		name  string
		input string
		check func(t *testing.T, got *HookInput)
	}{
		{
			name: "flat snake_case SessionStart still works",
			input: `{
				"session_id": "legacy-sess-1",
				"cwd": "/legacy/path",
				"hook_event_name": "SessionStart",
				"project_dir": "/legacy/path"
			}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.SessionID != "legacy-sess-1" {
					t.Errorf("SessionID = %q, want %q", got.SessionID, "legacy-sess-1")
				}
				if got.CWD != "/legacy/path" {
					t.Errorf("CWD = %q, want %q", got.CWD, "/legacy/path")
				}
				if got.HookEventName != "SessionStart" {
					t.Errorf("HookEventName = %q, want %q", got.HookEventName, "SessionStart")
				}
			},
		},
		{
			name: "flat snake_case PreToolUse still works",
			input: `{
				"session_id": "legacy-sess-2",
				"cwd": "/tmp",
				"hook_event_name": "PreToolUse",
				"tool_name": "Write",
				"tool_input": {"file_path": "/tmp/test.go"}
			}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.ToolName != "Write" {
					t.Errorf("ToolName = %q, want %q", got.ToolName, "Write")
				}
				if got.ToolInput == nil {
					t.Error("ToolInput is nil, want non-nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := strings.NewReader(tt.input)
			got, err := proto.ReadInput(r)
			if err != nil {
				t.Fatalf("ReadInput() backward compatibility broken: %v", err)
			}
			if got == nil {
				t.Fatal("ReadInput() returned nil HookInput")
			}
			tt.check(t, got)
		})
	}
}

// TestNormalizeHookInput_NativeFormat tests the normalizeHookInput function
// directly with Claude Code's native camelCase format.
func TestNormalizeHookInput_NativeFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         string
		wantSessionID string
		wantCWD       string
		wantEventName string
		wantToolName  string
	}{
		{
			name:          "session fields are flattened",
			input:         `{"eventType":"PreToolUse","toolName":"Bash","session":{"id":"s1","cwd":"/tmp","projectDir":"/proj"}}`,
			wantSessionID: "s1",
			wantCWD:       "/tmp",
			wantEventName: "PreToolUse",
			wantToolName:  "Bash",
		},
		{
			name:          "session without projectDir",
			input:         `{"eventType":"SessionStart","session":{"id":"s2","cwd":"/home"}}`,
			wantSessionID: "s2",
			wantCWD:       "/home",
			wantEventName: "SessionStart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := normalizeHookInput([]byte(tt.input))
			if err != nil {
				t.Fatalf("normalizeHookInput() error: %v", err)
			}

			var m map[string]json.RawMessage
			if err := json.Unmarshal(result, &m); err != nil {
				t.Fatalf("result is not valid JSON: %v", err)
			}

			if tt.wantSessionID != "" {
				v, ok := m["session_id"]
				if !ok {
					t.Error("session_id missing from normalized output")
				} else {
					var s string
					if err := json.Unmarshal(v, &s); err != nil || s != tt.wantSessionID {
						t.Errorf("session_id = %q, want %q", s, tt.wantSessionID)
					}
				}
			}

			if tt.wantCWD != "" {
				v, ok := m["cwd"]
				if !ok {
					t.Error("cwd missing from normalized output")
				} else {
					var s string
					if err := json.Unmarshal(v, &s); err != nil || s != tt.wantCWD {
						t.Errorf("cwd = %q, want %q", s, tt.wantCWD)
					}
				}
			}

			if tt.wantEventName != "" {
				v, ok := m["hook_event_name"]
				if !ok {
					t.Error("hook_event_name missing from normalized output")
				} else {
					var s string
					if err := json.Unmarshal(v, &s); err != nil || s != tt.wantEventName {
						t.Errorf("hook_event_name = %q, want %q", s, tt.wantEventName)
					}
				}
			}

			if tt.wantToolName != "" {
				v, ok := m["tool_name"]
				if !ok {
					t.Error("tool_name missing from normalized output")
				} else {
					var s string
					if err := json.Unmarshal(v, &s); err != nil || s != tt.wantToolName {
						t.Errorf("tool_name = %q, want %q", s, tt.wantToolName)
					}
				}
			}

			// Ensure original camelCase keys are removed
			if _, ok := m["eventType"]; ok {
				t.Error("eventType should be removed after normalization")
			}
			if _, ok := m["toolName"]; ok {
				t.Error("toolName should be removed after normalization")
			}
			if _, ok := m["session"]; ok {
				t.Error("session object should be removed after normalization (fields flattened)")
			}
		})
	}
}

// TestNormalizeHookInput_LegacyFormat verifies that flat snake_case inputs
// pass through normalizeHookInput unchanged.
func TestNormalizeHookInput_LegacyFormat(t *testing.T) {
	t.Parallel()

	input := `{"session_id":"s1","cwd":"/tmp","hook_event_name":"SessionStart"}`
	result, err := normalizeHookInput([]byte(input))
	if err != nil {
		t.Fatalf("normalizeHookInput() error: %v", err)
	}

	// Result should be identical to input (no transformation)
	if string(result) != input {
		t.Errorf("legacy format was modified unexpectedly:\n  got:  %s\n  want: %s", result, input)
	}
}

// TestNormalizeHookInput_AdditionalCamelCaseFields verifies that additional
// camelCase fields beyond the session mapping are also normalized.
func TestNormalizeHookInput_AdditionalCamelCaseFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantKey   string
		wantValue string
	}{
		{
			name:      "toolUseId maps to tool_use_id",
			input:     `{"eventType":"PreToolUse","session":{"id":"s1","cwd":"/tmp"},"toolName":"Bash","toolUseId":"tu-001"}`,
			wantKey:   "tool_use_id",
			wantValue: "tu-001",
		},
		{
			name:      "transcriptPath maps to transcript_path",
			input:     `{"eventType":"SessionStart","session":{"id":"s1","cwd":"/tmp"},"transcriptPath":"/tmp/t.jsonl"}`,
			wantKey:   "transcript_path",
			wantValue: "/tmp/t.jsonl",
		},
		{
			name:      "agentId maps to agent_id",
			input:     `{"eventType":"SubagentStart","session":{"id":"s1","cwd":"/tmp"},"agentId":"agent-1"}`,
			wantKey:   "agent_id",
			wantValue: "agent-1",
		},
		{
			name:      "notificationType maps to notification_type",
			input:     `{"eventType":"Notification","session":{"id":"s1","cwd":"/tmp"},"notificationType":"info"}`,
			wantKey:   "notification_type",
			wantValue: "info",
		},
		{
			name:      "teamName maps to team_name",
			input:     `{"eventType":"TeammateIdle","session":{"id":"s1","cwd":"/tmp"},"teamName":"my-team"}`,
			wantKey:   "team_name",
			wantValue: "my-team",
		},
		{
			name:      "worktreePath maps to worktree_path",
			input:     `{"eventType":"WorktreeCreate","session":{"id":"s1","cwd":"/tmp"},"worktreePath":"/tmp/wt-1"}`,
			wantKey:   "worktree_path",
			wantValue: "/tmp/wt-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := normalizeHookInput([]byte(tt.input))
			if err != nil {
				t.Fatalf("normalizeHookInput() error: %v", err)
			}

			var m map[string]json.RawMessage
			if err := json.Unmarshal(result, &m); err != nil {
				t.Fatalf("result is invalid JSON: %v", err)
			}

			v, ok := m[tt.wantKey]
			if !ok {
				t.Errorf("key %q missing from normalized output; got keys: %v", tt.wantKey, normalizeTestMapKeys(m))
				return
			}

			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				t.Fatalf("value for %q is not a string: %v", tt.wantKey, err)
			}
			if s != tt.wantValue {
				t.Errorf("%s = %q, want %q", tt.wantKey, s, tt.wantValue)
			}
		})
	}
}

// TestNormalizeHookInput_InvalidJSON verifies that invalid JSON returns an error.
func TestNormalizeHookInput_InvalidJSON(t *testing.T) {
	t.Parallel()

	inputs := []string{
		``,
		`   `,
		`not json`,
		`{invalid}`,
		`{"unclosed": "object"`,
	}

	for _, input := range inputs {
		input := input
		t.Run("input="+strings.TrimSpace(input), func(t *testing.T) {
			t.Parallel()

			_, err := normalizeHookInput([]byte(input))
			if err == nil {
				t.Errorf("normalizeHookInput(%q) = nil error, want error", input)
			}
		})
	}
}

// normalizeTestMapKeys returns the keys of a map for test diagnostics.
func normalizeTestMapKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
