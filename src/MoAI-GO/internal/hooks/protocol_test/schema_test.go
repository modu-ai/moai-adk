package protocol_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/anthropics/moai-adk-go/internal/hooks/protocol"
)

// TestHookInputHelpers tests helper methods for tool input
func TestHookInputHelpers(t *testing.T) {
	input := &protocol.HookInput{
		SessionID: "test-session",
		Event:     "pre_tool",
		ToolName:  "Write",
		ToolInput: map[string]any{
			"file_path": "/path/to/test.py",
			"matcher":   "Write",
		},
	}

	t.Run("GetToolInputPath", func(t *testing.T) {
		path := input.GetToolInputPath()
		if path != "/path/to/test.py" {
			t.Errorf("Expected path '/path/to/test.py', got '%s'", path)
		}
	})

	t.Run("GetToolInputMatcher", func(t *testing.T) {
		matcher := input.GetToolInputMatcher()
		if matcher != "Write" {
			t.Errorf("Expected matcher 'Write', got '%s'", matcher)
		}
	})

	t.Run("GetToolInputCommand", func(t *testing.T) {
		command := input.GetToolInputCommand()
		if command != "" {
			t.Errorf("Expected empty command, got '%s'", command)
		}
	})

	// Test with bash command
	bashInput := &protocol.HookInput{
		SessionID: "test-session",
		Event:     "pre_tool",
		ToolName:  "Bash",
		ToolInput: map[string]any{
			"command": "echo 'test'",
		},
	}

	t.Run("GetToolInputCommand with bash", func(t *testing.T) {
		command := bashInput.GetToolInputCommand()
		if command != "echo 'test'" {
			t.Errorf("Expected command \"echo 'test'\", got '%s'", command)
		}
	})
}

// TestHookResponseClean tests that empty values are removed from JSON
func TestHookResponseClean(t *testing.T) {
	t.Run("Full response", func(t *testing.T) {
		response := protocol.NewMessageResponse("Test message", false)

		// Verify response has expected fields
		if response.SystemMessage == "" {
			t.Error("Expected SystemMessage to be set")
		}
		if !response.ContinueExecution {
			t.Error("Expected ContinueExecution to be true")
		}
	})

	t.Run("Empty response", func(t *testing.T) {
		response := protocol.NewMessageResponse("", true)

		// Verify empty message
		if response.SystemMessage != "" {
			t.Error("Expected SystemMessage to be empty")
		}
		if !response.ContinueExecution {
			t.Error("Expected ContinueExecution to be true")
		}
		if !response.SuppressOutput {
			t.Error("Expected SuppressOutput to be true")
		}
	})
}

// TestSecurityDecisionCreation tests security decision responses
func TestSecurityDecisionCreation(t *testing.T) {
	t.Run("Block decision", func(t *testing.T) {
		response := protocol.NewSecurityResponse(protocol.DecisionBlock, "Test block")
		if response.BlockExecution != true {
			t.Error("Expected BlockExecution to be true")
		}
		if response.ContinueExecution != false {
			t.Error("Expected ContinueExecution to be false")
		}
		if response.SuppressOutput != false {
			t.Error("Expected SuppressOutput to be false")
		}
	})

	t.Run("Allow decision", func(t *testing.T) {
		response := protocol.NewSecurityResponse(protocol.DecisionAllow, "")
		if response.BlockExecution != false {
			t.Error("Expected BlockExecution to be false")
		}
		if response.ContinueExecution != true {
			t.Error("Expected ContinueExecution to be true")
		}
		if response.SuppressOutput != true {
			t.Error("Expected SuppressOutput to be true for allow")
		}
	})

	t.Run("Warn decision", func(t *testing.T) {
		response := protocol.NewSecurityResponse(protocol.DecisionWarn, "Test warn")
		if response.BlockExecution != false {
			t.Error("Expected BlockExecution to be false")
		}
		if response.ContinueExecution != true {
			t.Error("Expected ContinueExecution to be true")
		}
		if response.SuppressOutput != false {
			t.Error("Expected SuppressOutput to be false")
		}
	})
}

// TestJSONSchemaMatching tests that JSON output matches Python schema exactly
func TestJSONSchemaMatching(t *testing.T) {
	// This test ensures the Go output matches Python hook output format

	t.Run("Security block JSON schema", func(t *testing.T) {
		response := protocol.NewSecurityResponse(protocol.DecisionBlock, "Test block reason")

		data, err := toJSON(response)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		jsonStr := string(data)
		requiredFields := []string{
			`"continue":false`,
			`"block_execution":true`,
			`"hookSpecificOutput"`,
			`"permissionDecision":"block"`,
			`"permissionDecisionReason":"Test block reason"`,
		}

		for _, field := range requiredFields {
			if !strings.Contains(jsonStr, field) {
				t.Errorf("Expected JSON to contain '%s', got:\n%s", field, jsonStr)
			}
		}
	})

	t.Run("Message JSON schema", func(t *testing.T) {
		response := protocol.NewMessageResponse("Test message", false)

		data, err := toJSON(response)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		jsonStr := string(data)
		if !strings.Contains(jsonStr, `"systemMessage":"Test message"`) {
			t.Errorf("Expected systemMessage in JSON, got:\n%s", jsonStr)
		}
		if !strings.Contains(jsonStr, `"continue":true`) {
			t.Errorf("Expected continue:true in JSON, got:\n%s", jsonStr)
		}
	})
}

// toJSON is a helper method to convert response to JSON
func toJSON(h *protocol.HookResponse) ([]byte, error) {
	return json.Marshal(h)
}
