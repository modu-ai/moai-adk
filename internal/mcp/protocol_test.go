package mcp

import (
	"encoding/json"
	"testing"
)

func TestJSONRPCRequest_MarshalRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "request with string id",
			input: `{"jsonrpc":"2.0","id":"abc","method":"tools/list"}`,
		},
		{
			name:  "request with numeric id",
			input: `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		},
		{
			name:  "notification without id",
			input: `{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var req JSONRPCRequest
			if err := json.Unmarshal([]byte(tt.input), &req); err != nil {
				t.Fatalf("Unmarshal(%q): %v", tt.input, err)
			}

			data, err := json.Marshal(req)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}

			// Re-unmarshal and compare fields rather than raw bytes, since
			// JSON key ordering is not guaranteed.
			var got JSONRPCRequest
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("re-Unmarshal: %v", err)
			}

			if got.JSONRPC != req.JSONRPC {
				t.Errorf("JSONRPC: got %q want %q", got.JSONRPC, req.JSONRPC)
			}
			if got.Method != req.Method {
				t.Errorf("Method: got %q want %q", got.Method, req.Method)
			}
		})
	}
}

func TestJSONRPCResponse_ErrorAndResult(t *testing.T) {
	t.Parallel()

	t.Run("success response", func(t *testing.T) {
		t.Parallel()

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`1`),
			Result:  map[string]string{"status": "ok"},
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}

		var got JSONRPCResponse
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}

		if got.Error != nil {
			t.Errorf("unexpected error in success response: %v", got.Error)
		}
	})

	t.Run("error response", func(t *testing.T) {
		t.Parallel()

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`2`),
			Error: &JSONRPCError{
				Code:    -32603,
				Message: "internal error",
			},
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}

		var got JSONRPCResponse
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}

		if got.Error == nil {
			t.Fatal("expected error in error response, got nil")
		}
		if got.Error.Code != -32603 {
			t.Errorf("Error.Code: got %d want -32603", got.Error.Code)
		}
		if got.Error.Message != "internal error" {
			t.Errorf("Error.Message: got %q want %q", got.Error.Message, "internal error")
		}
	})
}

func TestToolSchema_Properties(t *testing.T) {
	t.Parallel()

	schema := ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"file": {Type: "string", Description: "Relative file path"},
			"line": {Type: "integer", Description: "1-based line number"},
		},
		Required: []string{"file", "line"},
	}

	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got ToolSchema
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.Type != "object" {
		t.Errorf("Type: got %q want %q", got.Type, "object")
	}
	if len(got.Properties) != 2 {
		t.Errorf("Properties count: got %d want 2", len(got.Properties))
	}
	if len(got.Required) != 2 {
		t.Errorf("Required count: got %d want 2", len(got.Required))
	}
}
