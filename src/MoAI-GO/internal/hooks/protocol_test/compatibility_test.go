package protocol_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/anthropics/moai-adk-go/internal/hooks/protocol"
)

// TestHookInputParsing tests parsing JSON input from stdin
func TestHookInputParsing(t *testing.T) {
	// Test valid input
	validJSON := `{
		"session_id": "test-session-123",
		"event": "session_start",
		"tool_name": "Write",
		"tool_input": {
			"file_path": "/path/to/file.txt"
		}
	}`

	// Create temp file with input
	tmpfile, err := os.CreateTemp("", "test-input-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	if _, err := tmpfile.WriteString(validJSON); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Read and parse
	var input protocol.HookInput
	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	if err := json.Unmarshal(data, &input); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify parsed data
	if input.SessionID != "test-session-123" {
		t.Errorf("Expected SessionID 'test-session-123', got '%s'", input.SessionID)
	}

	if input.Event != "session_start" {
		t.Errorf("Expected Event 'session_start', got '%s'", input.Event)
	}

	if input.ToolName != "Write" {
		t.Errorf("Expected ToolName 'Write', got '%s'", input.ToolName)
	}
}

// TestHookResponseOutput tests JSON output generation
func TestHookResponseOutput(t *testing.T) {
	// Test security response
	response := protocol.NewSecurityResponse(protocol.DecisionBlock, "Test block reason")

	// Verify response fields
	if !response.BlockExecution {
		t.Error("Expected BlockExecution to be true for block decision")
	}

	if response.ContinueExecution {
		t.Error("Expected ContinueExecution to be false for block decision")
	}

	if response.HookSpecificOutput == nil {
		t.Error("Expected HookSpecificOutput to be populated")
	} else {
		if decision, ok := response.HookSpecificOutput["permissionDecision"].(string); !ok || decision != "block" {
			t.Errorf("Expected permissionDecision 'block', got %v", decision)
		}
	}
}

// TestPythonCompatibility tests JSON schema compatibility with Python hooks
func TestPythonCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		response *protocol.HookResponse
		expected string // Expected JSON snippet
	}{
		{
			name:     "Security block response",
			response: protocol.NewSecurityResponse(protocol.DecisionBlock, "Blocked for security"),
			expected: `"permissionDecision":"block"`,
		},
		{
			name:     "Security warn response",
			response: protocol.NewSecurityResponse(protocol.DecisionWarn, "Config file modification"),
			expected: `"permissionDecision":"warn"`,
		},
		{
			name:     "Security allow response",
			response: protocol.NewSecurityResponse(protocol.DecisionAllow, ""),
			expected: `"suppressOutput":true`,
		},
		{
			name:     "Message response",
			response: protocol.NewMessageResponse("Test message", false),
			expected: `"systemMessage":"Test message"`,
		},
		{
			name:     "Error response",
			response: protocol.NewErrorResponse("Test error"),
			expected: `"error":"Test error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Failed to marshal response: %v", err)
			}

			if !strings.Contains(string(data), tt.expected) {
				t.Errorf("Expected JSON to contain '%s', got:\n%s", tt.expected, string(data))
			}
		})
	}
}

// TestEventValidation tests hook event validation
func TestEventValidation(t *testing.T) {
	validEvents := []protocol.HookEvent{
		protocol.EventSessionStart,
		protocol.EventPreTool,
		protocol.EventPostTool,
		protocol.EventSessionEnd,
		protocol.EventPreCompact,
		protocol.EventStop,
		protocol.EventNotification,
		protocol.EventQualityGate,
		protocol.EventCommit,
		protocol.EventPush,
		protocol.EventCompact,
	}

	for _, event := range validEvents {
		t.Run(string(event), func(t *testing.T) {
			if !event.IsValid() {
				t.Errorf("Expected event '%s' to be valid", event)
			}
		})
	}

	// Test invalid event
	invalidEvent := protocol.HookEvent("invalid_event")
	if invalidEvent.IsValid() {
		t.Error("Expected invalid_event to be invalid")
	}
}

// TestTimeoutConfiguration tests timeout values
func TestTimeoutConfiguration(t *testing.T) {
	expectedTimeouts := map[protocol.HookEvent]int{
		protocol.EventSessionStart: 5,
		protocol.EventPreTool:      5,
		protocol.EventPostTool:     30,
		protocol.EventSessionEnd:   5,
		protocol.EventPreCompact:   3,
		protocol.EventStop:         5,
		protocol.EventNotification: 5,
		protocol.EventQualityGate:  10,
		protocol.EventCommit:       5,
		protocol.EventPush:         5,
		protocol.EventCompact:      5,
	}

	for event, expectedTimeout := range expectedTimeouts {
		t.Run(string(event), func(t *testing.T) {
			timeout := protocol.GetTimeout(event)
			if timeout != expectedTimeout {
				t.Errorf("Expected timeout %d for event '%s', got %d", expectedTimeout, event, timeout)
			}
		})
	}
}

// TestExitCodes tests that exit codes match Python hooks
func TestExitCodes(t *testing.T) {
	// Exit codes are handled at the process level, but we can verify
	// that the response structure indicates success/failure correctly

	t.Run("Success response indicates success", func(t *testing.T) {
		response := protocol.NewMessageResponse("Success", true)
		if !response.ContinueExecution {
			t.Error("Expected ContinueExecution to be true for success")
		}
		if response.Error != "" {
			t.Error("Expected no error for success response")
		}
	})

	t.Run("Error response contains error", func(t *testing.T) {
		response := protocol.NewErrorResponse("Test error")
		if response.Error == "" {
			t.Error("Expected Error to be populated for error response")
		}
	})
}
