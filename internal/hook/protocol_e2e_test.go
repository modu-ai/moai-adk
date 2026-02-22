package hook

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestProtocol_ReadInput_NewEvents verifies that ReadInput correctly deserializes
// JSON payloads for all 7 new hook event types, including their event-specific fields.
func TestProtocol_ReadInput_NewEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		check func(t *testing.T, got *HookInput)
	}{
		{
			name:  "PostToolUseFailure with all fields",
			input: `{"session_id":"s1","cwd":"/tmp","hook_event_name":"PostToolUseFailure","tool_name":"Bash","tool_use_id":"tu-1","error":"command failed","is_interrupt":false}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.HookEventName != "PostToolUseFailure" {
					t.Errorf("HookEventName = %q, want %q", got.HookEventName, "PostToolUseFailure")
				}
				if got.ToolName != "Bash" {
					t.Errorf("ToolName = %q, want %q", got.ToolName, "Bash")
				}
				if got.ToolUseID != "tu-1" {
					t.Errorf("ToolUseID = %q, want %q", got.ToolUseID, "tu-1")
				}
				if got.Error != "command failed" {
					t.Errorf("Error = %q, want %q", got.Error, "command failed")
				}
				if got.IsInterrupt {
					t.Error("IsInterrupt = true, want false")
				}
			},
		},
		{
			name:  "PostToolUseFailure with is_interrupt true",
			input: `{"session_id":"s1b","cwd":"/tmp","hook_event_name":"PostToolUseFailure","tool_name":"Write","tool_use_id":"tu-2","error":"interrupted by user","is_interrupt":true}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if !got.IsInterrupt {
					t.Error("IsInterrupt = false, want true")
				}
				if got.Error != "interrupted by user" {
					t.Errorf("Error = %q, want %q", got.Error, "interrupted by user")
				}
			},
		},
		{
			name:  "Notification with all fields",
			input: `{"session_id":"s2","cwd":"/tmp","hook_event_name":"Notification","title":"Build Done","message":"Success","notification_type":"info"}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.HookEventName != "Notification" {
					t.Errorf("HookEventName = %q, want %q", got.HookEventName, "Notification")
				}
				if got.Title != "Build Done" {
					t.Errorf("Title = %q, want %q", got.Title, "Build Done")
				}
				if got.Message != "Success" {
					t.Errorf("Message = %q, want %q", got.Message, "Success")
				}
				if got.NotificationType != "info" {
					t.Errorf("NotificationType = %q, want %q", got.NotificationType, "info")
				}
			},
		},
		{
			name:  "SubagentStart with all fields",
			input: `{"session_id":"s3","cwd":"/tmp","hook_event_name":"SubagentStart","agent_id":"agent-1","agent_transcript_path":"/tmp/t.jsonl"}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.HookEventName != "SubagentStart" {
					t.Errorf("HookEventName = %q, want %q", got.HookEventName, "SubagentStart")
				}
				if got.AgentID != "agent-1" {
					t.Errorf("AgentID = %q, want %q", got.AgentID, "agent-1")
				}
				if got.AgentTranscriptPath != "/tmp/t.jsonl" {
					t.Errorf("AgentTranscriptPath = %q, want %q", got.AgentTranscriptPath, "/tmp/t.jsonl")
				}
			},
		},
		{
			name:  "UserPromptSubmit with prompt",
			input: `{"session_id":"s4","cwd":"/tmp","hook_event_name":"UserPromptSubmit","prompt":"fix this bug"}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.HookEventName != "UserPromptSubmit" {
					t.Errorf("HookEventName = %q, want %q", got.HookEventName, "UserPromptSubmit")
				}
				if got.Prompt != "fix this bug" {
					t.Errorf("Prompt = %q, want %q", got.Prompt, "fix this bug")
				}
			},
		},
		{
			name:  "PermissionRequest with tool_name",
			input: `{"session_id":"s5","cwd":"/tmp","hook_event_name":"PermissionRequest","tool_name":"Write"}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.HookEventName != "PermissionRequest" {
					t.Errorf("HookEventName = %q, want %q", got.HookEventName, "PermissionRequest")
				}
				if got.ToolName != "Write" {
					t.Errorf("ToolName = %q, want %q", got.ToolName, "Write")
				}
			},
		},
		{
			name:  "TeammateIdle with agent_id",
			input: `{"session_id":"s6","cwd":"/tmp","hook_event_name":"TeammateIdle","agent_id":"tm-1"}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.HookEventName != "TeammateIdle" {
					t.Errorf("HookEventName = %q, want %q", got.HookEventName, "TeammateIdle")
				}
				if got.AgentID != "tm-1" {
					t.Errorf("AgentID = %q, want %q", got.AgentID, "tm-1")
				}
			},
		},
		{
			name:  "TaskCompleted minimal",
			input: `{"session_id":"s7","cwd":"/tmp","hook_event_name":"TaskCompleted"}`,
			check: func(t *testing.T, got *HookInput) {
				t.Helper()
				if got.HookEventName != "TaskCompleted" {
					t.Errorf("HookEventName = %q, want %q", got.HookEventName, "TaskCompleted")
				}
				if got.SessionID != "s7" {
					t.Errorf("SessionID = %q, want %q", got.SessionID, "s7")
				}
				if got.CWD != "/tmp" {
					t.Errorf("CWD = %q, want %q", got.CWD, "/tmp")
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
			if err != nil {
				t.Fatalf("ReadInput() unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("ReadInput() returned nil, want non-nil")
			}
			tt.check(t, got)
		})
	}
}

// TestProtocol_WriteOutput_NewEventOutputs tests JSON serialization of all new
// output types, verifying hookSpecificOutput presence, ExitCode exclusion, and
// field correctness in the serialized JSON.
func TestProtocol_WriteOutput_NewEventOutputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		output *HookOutput
		check  func(t *testing.T, data []byte)
	}{
		{
			name:   "empty HookOutput produces valid JSON without hookSpecificOutput",
			output: &HookOutput{},
			check: func(t *testing.T, data []byte) {
				t.Helper()
				if !json.Valid(data) {
					t.Fatalf("output is not valid JSON: %s", data)
				}
				var m map[string]any
				if err := json.Unmarshal(data, &m); err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}
				if _, ok := m["hookSpecificOutput"]; ok {
					t.Error("hookSpecificOutput should not be present in empty output")
				}
			},
		},
		{
			name:   "PermissionRequest output with ask decision",
			output: NewPermissionRequestOutput(DecisionAsk, "deferred to user"),
			check: func(t *testing.T, data []byte) {
				t.Helper()
				var m map[string]any
				if err := json.Unmarshal(data, &m); err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}
				hso, ok := m["hookSpecificOutput"].(map[string]any)
				if !ok {
					t.Fatal("hookSpecificOutput missing or wrong type")
				}
				if hso["permissionDecision"] != DecisionAsk {
					t.Errorf("permissionDecision = %v, want %q", hso["permissionDecision"], DecisionAsk)
				}
				if hso["permissionDecisionReason"] != "deferred to user" {
					t.Errorf("permissionDecisionReason = %v, want %q", hso["permissionDecisionReason"], "deferred to user")
				}
				if hso["hookEventName"] != "PreToolUse" {
					t.Errorf("hookEventName = %v, want %q", hso["hookEventName"], "PreToolUse")
				}
			},
		},
		{
			name:   "TeammateKeepWorkingOutput excludes ExitCode from JSON",
			output: NewTeammateKeepWorkingOutput(),
			check: func(t *testing.T, data []byte) {
				t.Helper()
				s := string(data)
				if strings.Contains(s, "exitCode") || strings.Contains(s, "ExitCode") {
					t.Errorf("ExitCode should not appear in JSON: %s", s)
				}
				if !json.Valid(data) {
					t.Fatalf("output is not valid JSON: %s", data)
				}
			},
		},
		{
			name:   "TaskRejectedOutput excludes ExitCode from JSON",
			output: NewTaskRejectedOutput(),
			check: func(t *testing.T, data []byte) {
				t.Helper()
				s := string(data)
				if strings.Contains(s, "exitCode") || strings.Contains(s, "ExitCode") {
					t.Errorf("ExitCode should not appear in JSON: %s", s)
				}
				if !json.Valid(data) {
					t.Fatalf("output is not valid JSON: %s", data)
				}
			},
		},
		{
			name:   "PermissionRequest allow output with reason",
			output: NewPermissionRequestOutput(DecisionAllow, "approved by policy"),
			check: func(t *testing.T, data []byte) {
				t.Helper()
				var m map[string]any
				if err := json.Unmarshal(data, &m); err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}
				hso, ok := m["hookSpecificOutput"].(map[string]any)
				if !ok {
					t.Fatal("hookSpecificOutput missing or wrong type")
				}
				if hso["permissionDecision"] != DecisionAllow {
					t.Errorf("permissionDecision = %v, want %q", hso["permissionDecision"], DecisionAllow)
				}
				if hso["permissionDecisionReason"] != "approved by policy" {
					t.Errorf("permissionDecisionReason = %v, want %q", hso["permissionDecisionReason"], "approved by policy")
				}
			},
		},
		{
			name:   "UserPromptBlock output has decision block and reason",
			output: NewUserPromptBlockOutput("blocked"),
			check: func(t *testing.T, data []byte) {
				t.Helper()
				var parsed HookOutput
				if err := json.Unmarshal(data, &parsed); err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}
				if parsed.Decision != DecisionBlock {
					t.Errorf("Decision = %q, want %q", parsed.Decision, DecisionBlock)
				}
				if parsed.Reason != "blocked" {
					t.Errorf("Reason = %q, want %q", parsed.Reason, "blocked")
				}
			},
		},
		{
			name:   "output with UpdatedInput field appears in JSON",
			output: &HookOutput{UpdatedInput: "modified prompt text"},
			check: func(t *testing.T, data []byte) {
				t.Helper()
				var m map[string]any
				if err := json.Unmarshal(data, &m); err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}
				v, ok := m["updatedInput"]
				if !ok {
					t.Fatal("updatedInput key missing from JSON output")
				}
				if v != "modified prompt text" {
					t.Errorf("updatedInput = %v, want %q", v, "modified prompt text")
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
			if err != nil {
				t.Fatalf("WriteOutput() unexpected error: %v", err)
			}
			tt.check(t, buf.Bytes())
		})
	}
}

// TestProtocol_FullRoundTrip_NewEvents verifies full round-trip integrity for each
// of the 7 new event types: create input JSON -> ReadInput -> create appropriate
// HookOutput -> WriteOutput -> parse JSON -> verify structure.
func TestProtocol_FullRoundTrip_NewEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		inputJSON   string
		makeOutput  func(input *HookInput) *HookOutput
		checkOutput func(t *testing.T, parsed map[string]any)
	}{
		{
			name:      "PostToolUseFailure round-trip",
			inputJSON: `{"session_id":"rt1","cwd":"/tmp","hook_event_name":"PostToolUseFailure","tool_name":"Bash","error":"exit 1"}`,
			makeOutput: func(_ *HookInput) *HookOutput {
				// PostToolUseFailure handler returns empty output
				return &HookOutput{}
			},
			checkOutput: func(t *testing.T, parsed map[string]any) {
				t.Helper()
				if _, ok := parsed["hookSpecificOutput"]; ok {
					t.Error("PostToolUseFailure should not have hookSpecificOutput")
				}
			},
		},
		{
			name:      "Notification round-trip",
			inputJSON: `{"session_id":"rt2","cwd":"/tmp","hook_event_name":"Notification","title":"Alert","message":"Disk full","notification_type":"warning"}`,
			makeOutput: func(_ *HookInput) *HookOutput {
				return &HookOutput{}
			},
			checkOutput: func(t *testing.T, parsed map[string]any) {
				t.Helper()
				if _, ok := parsed["hookSpecificOutput"]; ok {
					t.Error("Notification should not have hookSpecificOutput")
				}
			},
		},
		{
			name:      "SubagentStart round-trip",
			inputJSON: `{"session_id":"rt3","cwd":"/tmp","hook_event_name":"SubagentStart","agent_id":"sa-1","agent_transcript_path":"/tmp/sa.jsonl"}`,
			makeOutput: func(_ *HookInput) *HookOutput {
				return &HookOutput{}
			},
			checkOutput: func(t *testing.T, parsed map[string]any) {
				t.Helper()
				if _, ok := parsed["hookSpecificOutput"]; ok {
					t.Error("SubagentStart should not have hookSpecificOutput")
				}
			},
		},
		{
			name:      "UserPromptSubmit round-trip with updatedInput",
			inputJSON: `{"session_id":"rt4","cwd":"/tmp","hook_event_name":"UserPromptSubmit","prompt":"hello world"}`,
			makeOutput: func(_ *HookInput) *HookOutput {
				return &HookOutput{UpdatedInput: "hello world (sanitized)"}
			},
			checkOutput: func(t *testing.T, parsed map[string]any) {
				t.Helper()
				v, ok := parsed["updatedInput"]
				if !ok {
					t.Fatal("updatedInput key missing from round-trip output")
				}
				if v != "hello world (sanitized)" {
					t.Errorf("updatedInput = %v, want %q", v, "hello world (sanitized)")
				}
			},
		},
		{
			name:      "PermissionRequest round-trip with allow",
			inputJSON: `{"session_id":"rt5","cwd":"/tmp","hook_event_name":"PermissionRequest","tool_name":"Read"}`,
			makeOutput: func(_ *HookInput) *HookOutput {
				return NewPermissionRequestOutput(DecisionAllow, "safe tool")
			},
			checkOutput: func(t *testing.T, parsed map[string]any) {
				t.Helper()
				hso, ok := parsed["hookSpecificOutput"].(map[string]any)
				if !ok {
					t.Fatal("hookSpecificOutput missing for PermissionRequest")
				}
				if hso["permissionDecision"] != DecisionAllow {
					t.Errorf("permissionDecision = %v, want %q", hso["permissionDecision"], DecisionAllow)
				}
				if hso["hookEventName"] != "PreToolUse" {
					t.Errorf("hookEventName = %v, want %q", hso["hookEventName"], "PreToolUse")
				}
			},
		},
		{
			name:      "TeammateIdle round-trip with keep-working exit code",
			inputJSON: `{"session_id":"rt6","cwd":"/tmp","hook_event_name":"TeammateIdle","agent_id":"tm-2"}`,
			makeOutput: func(_ *HookInput) *HookOutput {
				return NewTeammateKeepWorkingOutput()
			},
			checkOutput: func(t *testing.T, parsed map[string]any) {
				t.Helper()
				// ExitCode should not appear in JSON
				if _, ok := parsed["exitCode"]; ok {
					t.Error("exitCode should not be in JSON output")
				}
				if _, ok := parsed["ExitCode"]; ok {
					t.Error("ExitCode should not be in JSON output")
				}
			},
		},
		{
			name:      "TaskCompleted round-trip with rejection exit code",
			inputJSON: `{"session_id":"rt7","cwd":"/tmp","hook_event_name":"TaskCompleted"}`,
			makeOutput: func(_ *HookInput) *HookOutput {
				return NewTaskRejectedOutput()
			},
			checkOutput: func(t *testing.T, parsed map[string]any) {
				t.Helper()
				// ExitCode should not appear in JSON
				if _, ok := parsed["exitCode"]; ok {
					t.Error("exitCode should not be in JSON output")
				}
				if _, ok := parsed["ExitCode"]; ok {
					t.Error("ExitCode should not be in JSON output")
				}
			},
		},
	}

	proto := &jsonProtocol{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Step 1: ReadInput
			r := strings.NewReader(tt.inputJSON)
			input, err := proto.ReadInput(r)
			if err != nil {
				t.Fatalf("ReadInput() unexpected error: %v", err)
			}
			if input == nil {
				t.Fatal("ReadInput() returned nil")
			}

			// Step 2: Create appropriate output
			output := tt.makeOutput(input)

			// Step 3: WriteOutput
			var buf bytes.Buffer
			if err := proto.WriteOutput(&buf, output); err != nil {
				t.Fatalf("WriteOutput() unexpected error: %v", err)
			}

			// Step 4: Parse buffer as JSON
			written := buf.Bytes()
			if !json.Valid(written) {
				t.Fatalf("WriteOutput produced invalid JSON: %s", written)
			}

			var parsed map[string]any
			if err := json.Unmarshal(written, &parsed); err != nil {
				t.Fatalf("unmarshal round-trip output: %v", err)
			}

			// Step 5: Verify structure
			tt.checkOutput(t, parsed)
		})
	}
}
