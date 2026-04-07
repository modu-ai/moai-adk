package mcp

import (
	"context"
	"encoding/json"
	"testing"
)

func TestLSPHandler_Initialize(t *testing.T) {
	t.Parallel()

	h := NewLSPHandler("/tmp/test")
	result, err := h.HandleMethod(context.Background(), "initialize", nil)
	if err != nil {
		t.Fatalf("HandleMethod(initialize): %v", err)
	}

	init, ok := result.(*InitializeResult)
	if !ok {
		t.Fatalf("expected *InitializeResult, got %T", result)
	}

	if init.ProtocolVersion == "" {
		t.Error("ProtocolVersion should not be empty")
	}
	if init.ServerInfo.Name != "moai-lsp" {
		t.Errorf("ServerInfo.Name: got %q want %q", init.ServerInfo.Name, "moai-lsp")
	}
	if init.Capabilities.Tools == nil {
		t.Error("Capabilities.Tools should not be nil")
	}
}

func TestLSPHandler_ToolsList(t *testing.T) {
	t.Parallel()

	h := NewLSPHandler("/tmp/test")
	result, err := h.HandleMethod(context.Background(), "tools/list", nil)
	if err != nil {
		t.Fatalf("HandleMethod(tools/list): %v", err)
	}

	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}

	tools, ok := m["tools"]
	if !ok {
		t.Fatal("response missing 'tools' key")
	}

	toolList, ok := tools.([]Tool)
	if !ok {
		t.Fatalf("tools: expected []Tool, got %T", tools)
	}

	wantTools := []string{
		"goto_definition", "find_references", "hover",
		"document_symbols", "diagnostics", "rename",
	}
	if len(toolList) != len(wantTools) {
		t.Errorf("tool count: got %d want %d", len(toolList), len(wantTools))
	}

	names := make(map[string]bool)
	for _, tool := range toolList {
		names[tool.Name] = true
	}
	for _, want := range wantTools {
		if !names[want] {
			t.Errorf("missing tool %q", want)
		}
	}
}

func TestLSPHandler_ToolsCall_Dispatch(t *testing.T) {
	t.Parallel()

	h := NewLSPHandler("/tmp/test")

	tests := []struct {
		name       string
		toolName   string
		args       map[string]any
		wantIsErr  bool
		wantInText string
	}{
		{
			name:       "goto_definition returns stub",
			toolName:   "goto_definition",
			args:       map[string]any{"file": "main.go", "line": float64(10), "column": float64(5)},
			wantIsErr:  true,
			wantInText: "goto_definition",
		},
		{
			name:       "find_references returns stub",
			toolName:   "find_references",
			args:       map[string]any{"file": "main.go", "line": float64(1), "column": float64(1)},
			wantIsErr:  true,
			wantInText: "find_references",
		},
		{
			name:       "hover returns stub",
			toolName:   "hover",
			args:       map[string]any{"file": "main.go", "line": float64(1), "column": float64(1)},
			wantIsErr:  true,
			wantInText: "hover",
		},
		{
			name:       "document_symbols returns stub",
			toolName:   "document_symbols",
			args:       map[string]any{"file": "main.go"},
			wantIsErr:  true,
			wantInText: "document_symbols",
		},
		{
			name:       "diagnostics returns stub",
			toolName:   "diagnostics",
			args:       map[string]any{"file": "main.go"},
			wantIsErr:  true,
			wantInText: "diagnostics",
		},
		{
			name:       "rename returns stub",
			toolName:   "rename",
			args:       map[string]any{"file": "main.go", "line": float64(1), "column": float64(1), "new_name": "newFunc"},
			wantIsErr:  true,
			wantInText: "rename",
		},
		{
			name:       "unknown tool returns error result",
			toolName:   "nonexistent",
			args:       map[string]any{},
			wantIsErr:  true,
			wantInText: "unknown tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, err := json.Marshal(ToolCallParams{Name: tt.toolName, Arguments: tt.args})
			if err != nil {
				t.Fatalf("marshal params: %v", err)
			}

			result, callErr := h.HandleMethod(context.Background(), "tools/call", params)
			if callErr != nil {
				t.Fatalf("HandleMethod: unexpected error %v", callErr)
			}

			toolResult, ok := result.(*ToolResult)
			if !ok {
				t.Fatalf("expected *ToolResult, got %T", result)
			}

			if toolResult.IsError != tt.wantIsErr {
				t.Errorf("IsError: got %v want %v", toolResult.IsError, tt.wantIsErr)
			}

			if len(toolResult.Content) == 0 {
				t.Fatal("expected at least one content item")
			}

			found := false
			for _, c := range toolResult.Content {
				if contains(c.Text, tt.wantInText) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("content text does not contain %q; got: %s", tt.wantInText, toolResult.Content[0].Text)
			}
		})
	}
}

func TestLSPHandler_UnknownMethod(t *testing.T) {
	t.Parallel()

	h := NewLSPHandler("/tmp/test")
	_, err := h.HandleMethod(context.Background(), "no/such/method", nil)
	if err == nil {
		t.Fatal("expected error for unknown method, got nil")
	}
}

func TestLSPHandler_ToolsCall_InvalidParams(t *testing.T) {
	t.Parallel()

	h := NewLSPHandler("/tmp/test")
	_, err := h.HandleMethod(context.Background(), "tools/call", json.RawMessage(`not-json`))
	if err == nil {
		t.Fatal("expected error for invalid params, got nil")
	}
}

// contains reports whether s contains substr.
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && indexString(s, substr) >= 0)
}

func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
