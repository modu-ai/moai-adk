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

	// 실제 임시 파일 생성 (ScanFile이 파일 존재 여부를 확인하므로)
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
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
		t.Fatalf("Handle() 오류: %v", err)
	}
	if got == nil || got.Data == nil {
		t.Fatal("Data가 nil이면 안 됨")
	}

	var metrics map[string]any
	if err := json.Unmarshal(got.Data, &metrics); err != nil {
		t.Fatalf("메트릭 언마샬 실패: %v", err)
	}

	astScan, ok := metrics["ast_scan"]
	if !ok {
		t.Fatal("metrics에 ast_scan이 없음")
	}
	astMap, ok := astScan.(map[string]any)
	if !ok {
		t.Fatalf("ast_scan이 map이 아님: %T", astScan)
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
		name          string
		toolName      string
		toolInput     json.RawMessage
		analyzer      FileAnalyzer
		wantAstScan   bool
		wantMatches   int
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
			name:        "no ast_scan when tool input lacks file_path",
			toolName:    "Write",
			toolInput:   json.RawMessage(`{"content": "pkg"}`),
			analyzer:    &mockFileAnalyzer{},
			wantAstScan: false,
			wantNonBlocking: true,
		},
		{
			name:        "no ast_scan when tool input is invalid JSON",
			toolName:    "Write",
			toolInput:   json.RawMessage(`{invalid json`),
			analyzer:    &mockFileAnalyzer{},
			wantAstScan: false,
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

			// 관찰 전용: 오류 없이 항상 allow 반환
			if err != nil {
				t.Fatalf("Handle() 오류 반환 (관찰 전용이어야 함): %v", err)
			}
			if got == nil {
				t.Fatal("nil output 반환")
			}

			// HookSpecificOutput이 PostToolUse로 설정돼야 함
			if got.HookSpecificOutput == nil || got.HookSpecificOutput.HookEventName != "PostToolUse" {
				t.Errorf("HookEventName이 PostToolUse가 아님: %+v", got.HookSpecificOutput)
			}

			// ast_scan 메트릭 확인
			if got.Data != nil {
				var metrics map[string]any
				if err := json.Unmarshal(got.Data, &metrics); err != nil {
					t.Fatalf("메트릭 언마샬 실패: %v", err)
				}

				_, hasAstScan := metrics["ast_scan"]
				if tt.wantAstScan && !hasAstScan {
					t.Error("ast_scan이 metrics에 있어야 함")
				}
				if !tt.wantAstScan && hasAstScan {
					t.Error("ast_scan이 metrics에 없어야 함")
				}

				if tt.wantAstScan && hasAstScan {
					astMap := metrics["ast_scan"].(map[string]any)
					if astMap["matches"] != float64(tt.wantMatches) {
						t.Errorf("matches = %v, want %d", astMap["matches"], tt.wantMatches)
					}
				}
			} else if tt.wantAstScan {
				t.Error("Data가 nil인데 ast_scan을 기대함")
			}
		})
	}
}
