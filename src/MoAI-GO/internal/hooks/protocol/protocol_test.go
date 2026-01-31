package protocol

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

// --- ReadFromStdin tests ---

func TestReadFromStdin_ValidInput(t *testing.T) {
	input := `{"session_id":"sess-123","event":"session_start","tool_name":"Write","tool_input":{"file_path":"/tmp/test.py"}}`
	restoreStdin := overrideStdin(t, input)
	defer restoreStdin()

	result, err := ReadFromStdin()
	if err != nil {
		t.Fatalf("ReadFromStdin returned error: %v", err)
	}

	if result.SessionID != "sess-123" {
		t.Errorf("Expected SessionID 'sess-123', got '%s'", result.SessionID)
	}
	if result.Event != "session_start" {
		t.Errorf("Expected Event 'session_start', got '%s'", result.Event)
	}
	if result.ToolName != "Write" {
		t.Errorf("Expected ToolName 'Write', got '%s'", result.ToolName)
	}
	if result.GetToolInputPath() != "/tmp/test.py" {
		t.Errorf("Expected file_path '/tmp/test.py', got '%s'", result.GetToolInputPath())
	}
}

func TestReadFromStdin_AllEventTypes(t *testing.T) {
	events := []string{
		"session_start", "pre_tool", "post_tool", "session_end",
		"pre_compact", "stop", "notification", "quality_gate",
		"commit", "push", "compact",
	}

	for _, event := range events {
		t.Run(event, func(t *testing.T) {
			input := `{"session_id":"s","event":"` + event + `"}`
			restoreStdin := overrideStdin(t, input)
			defer restoreStdin()

			result, err := ReadFromStdin()
			if err != nil {
				t.Fatalf("ReadFromStdin returned error for event '%s': %v", event, err)
			}
			if result.Event != event {
				t.Errorf("Expected Event '%s', got '%s'", event, result.Event)
			}
		})
	}
}

func TestReadFromStdin_MissingSessionID(t *testing.T) {
	input := `{"event":"session_start"}`
	restoreStdin := overrideStdin(t, input)
	defer restoreStdin()

	_, err := ReadFromStdin()
	if err == nil {
		t.Fatal("Expected error for missing session_id, got nil")
	}
	if !strings.Contains(err.Error(), "session_id") {
		t.Errorf("Expected error about session_id, got: %v", err)
	}
}

func TestReadFromStdin_MissingEvent(t *testing.T) {
	input := `{"session_id":"s"}`
	restoreStdin := overrideStdin(t, input)
	defer restoreStdin()

	_, err := ReadFromStdin()
	if err == nil {
		t.Fatal("Expected error for missing event, got nil")
	}
	if !strings.Contains(err.Error(), "event") {
		t.Errorf("Expected error about event, got: %v", err)
	}
}

func TestReadFromStdin_InvalidEvent(t *testing.T) {
	input := `{"session_id":"s","event":"invalid_event_type"}`
	restoreStdin := overrideStdin(t, input)
	defer restoreStdin()

	_, err := ReadFromStdin()
	if err == nil {
		t.Fatal("Expected error for invalid event, got nil")
	}
	if !strings.Contains(err.Error(), "invalid event type") {
		t.Errorf("Expected error about invalid event type, got: %v", err)
	}
}

func TestReadFromStdin_InvalidJSON(t *testing.T) {
	input := `{not valid json`
	restoreStdin := overrideStdin(t, input)
	defer restoreStdin()

	_, err := ReadFromStdin()
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parse JSON") {
		t.Errorf("Expected JSON parse error, got: %v", err)
	}
}

func TestReadFromStdin_EmptyInput(t *testing.T) {
	restoreStdin := overrideStdin(t, "")
	defer restoreStdin()

	_, err := ReadFromStdin()
	if err == nil {
		t.Fatal("Expected error for empty input, got nil")
	}
}

func TestReadFromStdin_NilToolMapsInitialized(t *testing.T) {
	input := `{"session_id":"s","event":"session_start"}`
	restoreStdin := overrideStdin(t, input)
	defer restoreStdin()

	result, err := ReadFromStdin()
	if err != nil {
		t.Fatalf("ReadFromStdin returned error: %v", err)
	}

	if result.ToolInput == nil {
		t.Error("Expected ToolInput to be initialized (not nil)")
	}
	if result.ToolOutput == nil {
		t.Error("Expected ToolOutput to be initialized (not nil)")
	}
}

func TestReadFromStdin_WithToolOutput(t *testing.T) {
	input := `{"session_id":"s","event":"post_tool","tool_name":"Write","tool_input":{"file_path":"/tmp/f.py"},"tool_output":{"result":"ok"}}`
	restoreStdin := overrideStdin(t, input)
	defer restoreStdin()

	result, err := ReadFromStdin()
	if err != nil {
		t.Fatalf("ReadFromStdin returned error: %v", err)
	}
	if result.ToolOutput["result"] != "ok" {
		t.Errorf("Expected tool_output.result 'ok', got '%v'", result.ToolOutput["result"])
	}
}

// --- HookEvent.IsValid tests ---

func TestHookEvent_IsValid(t *testing.T) {
	validEvents := []HookEvent{
		EventSessionStart, EventPreTool, EventPostTool, EventSessionEnd,
		EventPreCompact, EventStop, EventNotification, EventQualityGate,
		EventCommit, EventPush, EventCompact,
	}

	for _, event := range validEvents {
		t.Run("valid_"+string(event), func(t *testing.T) {
			if !event.IsValid() {
				t.Errorf("Expected event '%s' to be valid", event)
			}
		})
	}

	invalidEvents := []HookEvent{
		"", "unknown", "pre_tool_use", "PostTool", "SESSION_START",
	}

	for _, event := range invalidEvents {
		name := string(event)
		if name == "" {
			name = "empty"
		}
		t.Run("invalid_"+name, func(t *testing.T) {
			if event.IsValid() {
				t.Errorf("Expected event '%s' to be invalid", event)
			}
		})
	}
}

// --- GetTimeout tests ---

func TestGetTimeout_AllEvents(t *testing.T) {
	expected := map[HookEvent]int{
		EventSessionStart: 5,
		EventPreTool:      5,
		EventPostTool:     30,
		EventSessionEnd:   5,
		EventPreCompact:   3,
		EventStop:         5,
		EventNotification: 5,
		EventQualityGate:  10,
		EventCommit:       5,
		EventPush:         5,
		EventCompact:      5,
	}

	for event, expectedTimeout := range expected {
		t.Run(string(event), func(t *testing.T) {
			timeout := GetTimeout(event)
			if timeout != expectedTimeout {
				t.Errorf("Expected timeout %d for event '%s', got %d", expectedTimeout, event, timeout)
			}
		})
	}
}

func TestGetTimeout_UnknownEvent(t *testing.T) {
	timeout := GetTimeout("unknown_event")
	if timeout != 5 {
		t.Errorf("Expected default timeout 5 for unknown event, got %d", timeout)
	}
}

// --- GetToolInputString tests ---

func TestGetToolInputString(t *testing.T) {
	input := &HookInput{
		ToolInput: map[string]interface{}{
			"string_key": "hello",
			"int_key":    42,
			"nil_key":    nil,
		},
	}

	t.Run("existing string key", func(t *testing.T) {
		val := input.GetToolInputString("string_key")
		if val != "hello" {
			t.Errorf("Expected 'hello', got '%s'", val)
		}
	})

	t.Run("non-string value returns empty", func(t *testing.T) {
		val := input.GetToolInputString("int_key")
		if val != "" {
			t.Errorf("Expected empty string for non-string value, got '%s'", val)
		}
	})

	t.Run("nil value returns empty", func(t *testing.T) {
		val := input.GetToolInputString("nil_key")
		if val != "" {
			t.Errorf("Expected empty string for nil value, got '%s'", val)
		}
	})

	t.Run("missing key returns empty", func(t *testing.T) {
		val := input.GetToolInputString("nonexistent")
		if val != "" {
			t.Errorf("Expected empty string for missing key, got '%s'", val)
		}
	})
}

func TestGetToolInputPath(t *testing.T) {
	input := &HookInput{
		ToolInput: map[string]interface{}{
			"file_path": "/some/path.go",
		},
	}
	if path := input.GetToolInputPath(); path != "/some/path.go" {
		t.Errorf("Expected '/some/path.go', got '%s'", path)
	}
}

func TestGetToolInputCommand(t *testing.T) {
	input := &HookInput{
		ToolInput: map[string]interface{}{
			"command": "echo hello",
		},
	}
	if cmd := input.GetToolInputCommand(); cmd != "echo hello" {
		t.Errorf("Expected 'echo hello', got '%s'", cmd)
	}
}

func TestGetToolInputMatcher(t *testing.T) {
	input := &HookInput{
		ToolInput: map[string]interface{}{
			"matcher": "Write",
		},
	}
	if matcher := input.GetToolInputMatcher(); matcher != "Write" {
		t.Errorf("Expected 'Write', got '%s'", matcher)
	}
}

// --- NewMessageResponse tests ---

func TestNewMessageResponse(t *testing.T) {
	t.Run("with message and no suppress", func(t *testing.T) {
		resp := NewMessageResponse("hello", false)
		if resp.SystemMessage != "hello" {
			t.Errorf("Expected SystemMessage 'hello', got '%s'", resp.SystemMessage)
		}
		if !resp.ContinueExecution {
			t.Error("Expected ContinueExecution to be true")
		}
		if resp.SuppressOutput {
			t.Error("Expected SuppressOutput to be false")
		}
	})

	t.Run("with suppress", func(t *testing.T) {
		resp := NewMessageResponse("", true)
		if resp.SystemMessage != "" {
			t.Errorf("Expected empty SystemMessage, got '%s'", resp.SystemMessage)
		}
		if !resp.SuppressOutput {
			t.Error("Expected SuppressOutput to be true")
		}
	})
}

// --- NewErrorResponse tests ---

func TestNewErrorResponse(t *testing.T) {
	resp := NewErrorResponse("something went wrong")
	if resp.Error != "something went wrong" {
		t.Errorf("Expected Error 'something went wrong', got '%s'", resp.Error)
	}
	if !resp.ContinueExecution {
		t.Error("Expected ContinueExecution to be true (graceful degradation)")
	}
	if resp.BlockExecution {
		t.Error("Expected BlockExecution to be false for error response")
	}
}

// --- NewSecurityResponse tests ---

func TestNewSecurityResponse_Block(t *testing.T) {
	resp := NewSecurityResponse(DecisionBlock, "blocked reason")

	if resp.ContinueExecution {
		t.Error("Expected ContinueExecution to be false for block")
	}
	if !resp.BlockExecution {
		t.Error("Expected BlockExecution to be true for block")
	}
	if resp.SuppressOutput {
		t.Error("Expected SuppressOutput to be false for block")
	}
	if resp.HookSpecificOutput == nil {
		t.Fatal("Expected HookSpecificOutput to be set")
	}
	if resp.HookSpecificOutput["permissionDecision"] != "block" {
		t.Errorf("Expected permissionDecision 'block', got '%v'", resp.HookSpecificOutput["permissionDecision"])
	}
	if resp.HookSpecificOutput["permissionDecisionReason"] != "blocked reason" {
		t.Errorf("Expected reason 'blocked reason', got '%v'", resp.HookSpecificOutput["permissionDecisionReason"])
	}
	if resp.HookSpecificOutput["hookEventName"] != "PreToolUse" {
		t.Errorf("Expected hookEventName 'PreToolUse', got '%v'", resp.HookSpecificOutput["hookEventName"])
	}
}

func TestNewSecurityResponse_Warn(t *testing.T) {
	resp := NewSecurityResponse(DecisionWarn, "warning reason")

	if !resp.ContinueExecution {
		t.Error("Expected ContinueExecution to be true for warn")
	}
	if resp.BlockExecution {
		t.Error("Expected BlockExecution to be false for warn")
	}
	if resp.SuppressOutput {
		t.Error("Expected SuppressOutput to be false for warn")
	}
	if resp.HookSpecificOutput == nil {
		t.Fatal("Expected HookSpecificOutput to be set")
	}
	if resp.HookSpecificOutput["permissionDecision"] != "warn" {
		t.Errorf("Expected permissionDecision 'warn', got '%v'", resp.HookSpecificOutput["permissionDecision"])
	}
}

func TestNewSecurityResponse_Allow(t *testing.T) {
	resp := NewSecurityResponse(DecisionAllow, "")

	if !resp.ContinueExecution {
		t.Error("Expected ContinueExecution to be true for allow")
	}
	if resp.BlockExecution {
		t.Error("Expected BlockExecution to be false for allow")
	}
	if !resp.SuppressOutput {
		t.Error("Expected SuppressOutput to be true for allow")
	}
	if resp.HookSpecificOutput != nil {
		t.Error("Expected HookSpecificOutput to be nil for allow")
	}
}

// --- WriteToStdout tests ---

func TestWriteToStdout_MessageResponse(t *testing.T) {
	resp := NewMessageResponse("test msg", false)
	output := captureStdout(t, func() {
		if err := resp.WriteToStdout(); err != nil {
			t.Fatalf("WriteToStdout returned error: %v", err)
		}
	})

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if parsed["systemMessage"] != "test msg" {
		t.Errorf("Expected systemMessage 'test msg', got '%v'", parsed["systemMessage"])
	}
	if parsed["continue"] != true {
		t.Errorf("Expected continue true, got '%v'", parsed["continue"])
	}
}

func TestWriteToStdout_SuppressedResponse(t *testing.T) {
	resp := NewMessageResponse("", true)
	output := captureStdout(t, func() {
		if err := resp.WriteToStdout(); err != nil {
			t.Fatalf("WriteToStdout returned error: %v", err)
		}
	})

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// systemMessage should be cleaned (empty, so omitted)
	if _, ok := parsed["systemMessage"]; ok {
		t.Error("Expected systemMessage to be omitted when empty")
	}
	if parsed["suppressOutput"] != true {
		t.Errorf("Expected suppressOutput true, got '%v'", parsed["suppressOutput"])
	}
}

func TestWriteToStdout_ErrorResponse(t *testing.T) {
	resp := NewErrorResponse("test error")
	output := captureStdout(t, func() {
		if err := resp.WriteToStdout(); err != nil {
			t.Fatalf("WriteToStdout returned error: %v", err)
		}
	})

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if parsed["error"] != "test error" {
		t.Errorf("Expected error 'test error', got '%v'", parsed["error"])
	}
}

func TestWriteToStdout_SecurityBlockResponse(t *testing.T) {
	resp := NewSecurityResponse(DecisionBlock, "security reason")
	output := captureStdout(t, func() {
		if err := resp.WriteToStdout(); err != nil {
			t.Fatalf("WriteToStdout returned error: %v", err)
		}
	})

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if parsed["continue"] != false {
		t.Errorf("Expected continue false for block, got '%v'", parsed["continue"])
	}
	if parsed["block_execution"] != true {
		t.Errorf("Expected block_execution true, got '%v'", parsed["block_execution"])
	}

	hookOutput, ok := parsed["hookSpecificOutput"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected hookSpecificOutput to be a map")
	}
	if hookOutput["permissionDecision"] != "block" {
		t.Errorf("Expected permissionDecision 'block', got '%v'", hookOutput["permissionDecision"])
	}
}

// --- clean() tests ---

func TestClean_RemovesEmptyFields(t *testing.T) {
	resp := &HookResponse{
		SystemMessage:      "",
		ContinueExecution:  true,
		ContextFiles:       nil,
		HookSpecificOutput: nil,
		BlockExecution:     false,
		SuppressOutput:     false,
		Error:              "",
	}

	cleaned := resp.clean()

	if cleaned.SystemMessage != "" {
		t.Error("Expected SystemMessage to remain empty")
	}
	if cleaned.ContextFiles != nil {
		t.Error("Expected ContextFiles to be nil")
	}
	if cleaned.HookSpecificOutput != nil {
		t.Error("Expected HookSpecificOutput to be nil")
	}
	if cleaned.BlockExecution {
		t.Error("Expected BlockExecution to be false")
	}
	if cleaned.SuppressOutput {
		t.Error("Expected SuppressOutput to be false")
	}
	if cleaned.Error != "" {
		t.Error("Expected Error to be empty")
	}
	if !cleaned.ContinueExecution {
		t.Error("Expected ContinueExecution to be preserved as true")
	}
}

func TestClean_PreservesPopulatedFields(t *testing.T) {
	resp := &HookResponse{
		SystemMessage:      "message",
		ContinueExecution:  false,
		ContextFiles:       []string{"file1.go", "file2.go"},
		HookSpecificOutput: map[string]any{"key": "value"},
		BlockExecution:     true,
		SuppressOutput:     true,
		Error:              "some error",
	}

	cleaned := resp.clean()

	if cleaned.SystemMessage != "message" {
		t.Errorf("Expected SystemMessage 'message', got '%s'", cleaned.SystemMessage)
	}
	if cleaned.ContinueExecution {
		t.Error("Expected ContinueExecution to be false")
	}
	if len(cleaned.ContextFiles) != 2 {
		t.Errorf("Expected 2 ContextFiles, got %d", len(cleaned.ContextFiles))
	}
	if cleaned.HookSpecificOutput["key"] != "value" {
		t.Error("Expected HookSpecificOutput to contain key=value")
	}
	if !cleaned.BlockExecution {
		t.Error("Expected BlockExecution to be true")
	}
	if !cleaned.SuppressOutput {
		t.Error("Expected SuppressOutput to be true")
	}
	if cleaned.Error != "some error" {
		t.Errorf("Expected Error 'some error', got '%s'", cleaned.Error)
	}
}

// --- WithContext / WithHookOutput tests ---

func TestWithContext(t *testing.T) {
	resp := NewMessageResponse("test", false)
	result := resp.WithContext([]string{"a.go", "b.go"})

	if result != resp {
		t.Error("Expected WithContext to return same pointer (builder pattern)")
	}
	if len(resp.ContextFiles) != 2 {
		t.Errorf("Expected 2 context files, got %d", len(resp.ContextFiles))
	}
	if resp.ContextFiles[0] != "a.go" || resp.ContextFiles[1] != "b.go" {
		t.Errorf("Unexpected context files: %v", resp.ContextFiles)
	}
}

func TestWithHookOutput(t *testing.T) {
	resp := NewMessageResponse("test", false)

	// First call should initialize the map
	result := resp.WithHookOutput("key1", "value1")
	if result != resp {
		t.Error("Expected WithHookOutput to return same pointer (builder pattern)")
	}
	if resp.HookSpecificOutput == nil {
		t.Fatal("Expected HookSpecificOutput to be initialized")
	}
	if resp.HookSpecificOutput["key1"] != "value1" {
		t.Errorf("Expected key1='value1', got '%v'", resp.HookSpecificOutput["key1"])
	}

	// Second call should add to existing map
	resp.WithHookOutput("key2", 42)
	if resp.HookSpecificOutput["key2"] != 42 {
		t.Errorf("Expected key2=42, got '%v'", resp.HookSpecificOutput["key2"])
	}
}

func TestWithHookOutput_NilMap(t *testing.T) {
	resp := &HookResponse{}
	resp.WithHookOutput("test", "value")
	if resp.HookSpecificOutput == nil {
		t.Error("Expected map to be created")
	}
	if resp.HookSpecificOutput["test"] != "value" {
		t.Errorf("Expected 'value', got '%v'", resp.HookSpecificOutput["test"])
	}
}

// --- SecurityDecision constant tests ---

func TestSecurityDecisionConstants(t *testing.T) {
	if string(DecisionAllow) != "allow" {
		t.Errorf("Expected DecisionAllow to be 'allow', got '%s'", DecisionAllow)
	}
	if string(DecisionBlock) != "block" {
		t.Errorf("Expected DecisionBlock to be 'block', got '%s'", DecisionBlock)
	}
	if string(DecisionWarn) != "warn" {
		t.Errorf("Expected DecisionWarn to be 'warn', got '%s'", DecisionWarn)
	}
}

// --- HookEvent constant tests ---

func TestHookEventConstants(t *testing.T) {
	expected := map[HookEvent]string{
		EventSessionStart: "session_start",
		EventPreTool:      "pre_tool",
		EventPostTool:     "post_tool",
		EventSessionEnd:   "session_end",
		EventPreCompact:   "pre_compact",
		EventStop:         "stop",
		EventNotification: "notification",
		EventQualityGate:  "quality_gate",
		EventCommit:       "commit",
		EventPush:         "push",
		EventCompact:      "compact",
	}

	for event, expectedStr := range expected {
		if string(event) != expectedStr {
			t.Errorf("Expected '%s', got '%s'", expectedStr, string(event))
		}
	}
}

// --- HookTimeouts map tests ---

func TestHookTimeouts_AllEventsHaveTimeouts(t *testing.T) {
	events := []HookEvent{
		EventSessionStart, EventPreTool, EventPostTool, EventSessionEnd,
		EventPreCompact, EventStop, EventNotification, EventQualityGate,
		EventCommit, EventPush, EventCompact,
	}

	for _, event := range events {
		if _, ok := HookTimeouts[event]; !ok {
			t.Errorf("Missing timeout for event '%s'", event)
		}
	}
}

func TestHookTimeouts_PostToolHasLongestTimeout(t *testing.T) {
	postToolTimeout := HookTimeouts[EventPostTool]
	for event, timeout := range HookTimeouts {
		if event != EventPostTool && event != EventQualityGate && timeout > postToolTimeout {
			t.Errorf("Event '%s' has timeout %d > PostTool timeout %d", event, timeout, postToolTimeout)
		}
	}
}

// --- Helper functions ---

// overrideStdin replaces os.Stdin with a reader containing the given data.
// Returns a function to restore the original stdin.
func overrideStdin(t *testing.T, data string) func() {
	t.Helper()

	origStdin := os.Stdin

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	_, err = w.WriteString(data)
	if err != nil {
		t.Fatalf("Failed to write to pipe: %v", err)
	}
	w.Close()

	os.Stdin = r

	return func() {
		os.Stdin = origStdin
		r.Close()
	}
}

// captureStdout captures everything written to os.Stdout during fn execution.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to read captured output: %v", err)
	}
	r.Close()

	return buf.String()
}
