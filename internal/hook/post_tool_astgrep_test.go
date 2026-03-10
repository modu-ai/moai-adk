package hook

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	astgrep "github.com/modu-ai/moai-adk/internal/astgrep"
)

// makeToolInput safely JSON-encodes key-value pairs for use as ToolInput.
// It handles OS-specific path separators (e.g., Windows backslashes in file paths).
func makeToolInput(fields map[string]string) json.RawMessage {
	data, _ := json.Marshal(fields)
	return json.RawMessage(data)
}

// mockFileAnalyzer is a mock FileAnalyzer implementation for testing.
type mockFileAnalyzer struct {
	scanFileFunc func(ctx context.Context, filePath string, config *astgrep.ScanConfig) (*astgrep.ScanResult, error)
}

func (m *mockFileAnalyzer) ScanFile(ctx context.Context, filePath string, config *astgrep.ScanConfig) (*astgrep.ScanResult, error) {
	if m.scanFileFunc != nil {
		return m.scanFileFunc(ctx, filePath, config)
	}
	return &astgrep.ScanResult{Matches: []astgrep.Match{}}, nil
}

func TestNewPostToolHandlerWithAstgrep(t *testing.T) {
	t.Parallel()

	h := NewPostToolHandlerWithAstgrep(nil, nil)
	if h == nil {
		t.Fatal("NewPostToolHandlerWithAstgrep returned nil")
	}
	if h.EventType() != EventPostToolUse {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventPostToolUse)
	}
}

func TestPostToolHandler_AstScan_WriteToolWithMatches(t *testing.T) {
	t.Parallel()

	// Create a real temp file because ScanFile checks whether the file exists.
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	analyzer := &mockFileAnalyzer{
		scanFileFunc: func(_ context.Context, _ string, _ *astgrep.ScanConfig) (*astgrep.ScanResult, error) {
			return &astgrep.ScanResult{
				Matches:  []astgrep.Match{{File: tmpFile, Line: 1, Text: "match", Rule: "test-rule"}},
				Language: "go",
			}, nil
		},
	}

	h := NewPostToolHandlerWithAstgrep(nil, analyzer)
	ctx := context.Background()

	input := &HookInput{
		SessionID: "sess-ast-matches",
		CWD:       tmpDir,
		ToolName:  "Write",
		ToolInput: makeToolInput(map[string]string{"file_path": tmpFile, "content": "package main\n"}),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if got == nil || got.Data == nil {
		t.Fatal("Data must not be nil")
	}

	var metrics map[string]any
	if err := json.Unmarshal(got.Data, &metrics); err != nil {
		t.Fatalf("failed to unmarshal metrics: %v", err)
	}

	astScan, ok := metrics["ast_scan"]
	if !ok {
		t.Fatal("ast_scan not found in metrics")
	}
	astMap, ok := astScan.(map[string]any)
	if !ok {
		t.Fatalf("ast_scan is not a map: %T", astScan)
	}
	if astMap["matches"] != float64(1) {
		t.Errorf("matches = %v, want 1", astMap["matches"])
	}
	if astMap["lang"] != "go" {
		t.Errorf("lang = %v, want go", astMap["lang"])
	}
}

func TestPostToolHandler_AstScan_TableDriven(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "example.go")
	if err := os.WriteFile(tmpFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		name            string
		toolName        string
		toolInput       json.RawMessage
		analyzer        FileAnalyzer
		wantAstScan     bool
		wantMatches     int
		wantNonBlocking bool // always true: observation-only, allow even on error
	}{
		{
			name:      "Write tool with no matches",
			toolName:  "Write",
			toolInput: makeToolInput(map[string]string{"file_path": tmpFile, "content": "pkg"}),
			analyzer: &mockFileAnalyzer{
				scanFileFunc: func(_ context.Context, _ string, _ *astgrep.ScanConfig) (*astgrep.ScanResult, error) {
					return &astgrep.ScanResult{Matches: []astgrep.Match{}, Language: "go"}, nil
				},
			},
			wantAstScan:     true,
			wantMatches:     0,
			wantNonBlocking: true,
		},
		{
			name:      "Edit tool with matches",
			toolName:  "Edit",
			toolInput: makeToolInput(map[string]string{"file_path": tmpFile, "old_string": "a", "new_string": "b"}),
			analyzer: &mockFileAnalyzer{
				scanFileFunc: func(_ context.Context, _ string, _ *astgrep.ScanConfig) (*astgrep.ScanResult, error) {
					return &astgrep.ScanResult{
						Matches:  []astgrep.Match{{File: tmpFile, Line: 1}},
						Language: "go",
					}, nil
				},
			},
			wantAstScan:     true,
			wantMatches:     1,
			wantNonBlocking: true,
		},
		{
			name:      "AST scan error is ignored (observation-only)",
			toolName:  "Write",
			toolInput: makeToolInput(map[string]string{"file_path": tmpFile, "content": "pkg"}),
			analyzer: &mockFileAnalyzer{
				scanFileFunc: func(_ context.Context, _ string, _ *astgrep.ScanConfig) (*astgrep.ScanResult, error) {
					return nil, errors.New("sg CLI error")
				},
			},
			wantAstScan:     false, // no ast_scan metric on error
			wantNonBlocking: true,
		},
		{
			name:            "no ast_scan when analyzer is nil",
			toolName:        "Write",
			toolInput:       makeToolInput(map[string]string{"file_path": tmpFile, "content": "pkg"}),
			analyzer:        nil,
			wantAstScan:     false,
			wantNonBlocking: true,
		},
		{
			name:            "Read tool does not perform ast_scan",
			toolName:        "Read",
			toolInput:       makeToolInput(map[string]string{"file_path": tmpFile}),
			analyzer:        &mockFileAnalyzer{},
			wantAstScan:     false,
			wantNonBlocking: true,
		},
		{
			name:            "no ast_scan when tool input lacks file_path",
			toolName:        "Write",
			toolInput:       json.RawMessage(`{"content": "pkg"}`),
			analyzer:        &mockFileAnalyzer{},
			wantAstScan:     false,
			wantNonBlocking: true,
		},
		{
			name:            "no ast_scan when tool input is invalid JSON",
			toolName:        "Write",
			toolInput:       json.RawMessage(`{invalid json`),
			analyzer:        &mockFileAnalyzer{},
			wantAstScan:     false,
			wantNonBlocking: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewPostToolHandlerWithAstgrep(nil, tt.analyzer)
			ctx := context.Background()

			input := &HookInput{
				SessionID: "sess-ast-table",
				CWD:       tmpDir,
				ToolName:  tt.toolName,
				ToolInput: tt.toolInput,
			}

			got, err := h.Handle(ctx, input)

			// Observation-only: always returns allow without error.
			if err != nil {
				t.Fatalf("Handle() returned error (must be observation-only): %v", err)
			}
			if got == nil {
				t.Fatal("returned nil output")
			}

			// HookSpecificOutput must be set to PostToolUse.
			if got.HookSpecificOutput == nil || got.HookSpecificOutput.HookEventName != "PostToolUse" {
				t.Errorf("HookEventName is not PostToolUse: %+v", got.HookSpecificOutput)
			}

			// Check ast_scan metric.
			if got.Data != nil {
				var metrics map[string]any
				if err := json.Unmarshal(got.Data, &metrics); err != nil {
					t.Fatalf("failed to unmarshal metrics: %v", err)
				}

				_, hasAstScan := metrics["ast_scan"]
				if tt.wantAstScan && !hasAstScan {
					t.Error("ast_scan should be present in metrics")
				}
				if !tt.wantAstScan && hasAstScan {
					t.Error("ast_scan should not be present in metrics")
				}

				if tt.wantAstScan && hasAstScan {
					astMap := metrics["ast_scan"].(map[string]any)
					if astMap["matches"] != float64(tt.wantMatches) {
						t.Errorf("matches = %v, want %d", astMap["matches"], tt.wantMatches)
					}
				}
			} else if tt.wantAstScan {
				t.Error("Data is nil but ast_scan was expected")
			}
		})
	}
}
