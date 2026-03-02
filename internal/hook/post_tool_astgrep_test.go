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

// mockFileAnalyzer는 테스트용 FileAnalyzer 목 구현체.
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
		ToolInput: json.RawMessage(`{"file_path": "` + tmpFile + `", "content": "package main\n"}`),
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
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}

	tests := []struct {
		name          string
		toolName      string
		toolInput     json.RawMessage
		analyzer      FileAnalyzer
		wantAstScan   bool
		wantMatches   int
		wantNonBlocking bool // 항상 true: 관찰 전용이므로 오류 발생해도 allow
	}{
		{
			name:     "Write 툴로 매치 없음",
			toolName: "Write",
			toolInput: json.RawMessage(`{"file_path": "` + tmpFile + `", "content": "pkg"}`),
			analyzer: &mockFileAnalyzer{
				scanFileFunc: func(_ context.Context, _ string, _ *astgrep.ScanConfig) (*astgrep.ScanResult, error) {
					return &astgrep.ScanResult{Matches: []astgrep.Match{}, Language: "go"}, nil
				},
			},
			wantAstScan: true,
			wantMatches: 0,
			wantNonBlocking: true,
		},
		{
			name:     "Edit 툴로 매치 있음",
			toolName: "Edit",
			toolInput: json.RawMessage(`{"file_path": "` + tmpFile + `", "old_string": "a", "new_string": "b"}`),
			analyzer: &mockFileAnalyzer{
				scanFileFunc: func(_ context.Context, _ string, _ *astgrep.ScanConfig) (*astgrep.ScanResult, error) {
					return &astgrep.ScanResult{
						Matches:  []astgrep.Match{{File: tmpFile, Line: 1}},
						Language: "go",
					}, nil
				},
			},
			wantAstScan: true,
			wantMatches: 1,
			wantNonBlocking: true,
		},
		{
			name:     "AST 스캔 오류는 무시됨 (관찰 전용)",
			toolName: "Write",
			toolInput: json.RawMessage(`{"file_path": "` + tmpFile + `", "content": "pkg"}`),
			analyzer: &mockFileAnalyzer{
				scanFileFunc: func(_ context.Context, _ string, _ *astgrep.ScanConfig) (*astgrep.ScanResult, error) {
					return nil, errors.New("sg CLI 오류")
				},
			},
			wantAstScan: false, // 오류 시 ast_scan 메트릭 없음
			wantNonBlocking: true,
		},
		{
			name:        "analyzer가 nil이면 ast_scan 없음",
			toolName:    "Write",
			toolInput:   json.RawMessage(`{"file_path": "` + tmpFile + `", "content": "pkg"}`),
			analyzer:    nil,
			wantAstScan: false,
			wantNonBlocking: true,
		},
		{
			name:        "Read 툴은 ast_scan 수행 안 함",
			toolName:    "Read",
			toolInput:   json.RawMessage(`{"file_path": "` + tmpFile + `"}`),
			analyzer:    &mockFileAnalyzer{},
			wantAstScan: false,
			wantNonBlocking: true,
		},
		{
			name:        "tool input에 file_path 없으면 ast_scan 없음",
			toolName:    "Write",
			toolInput:   json.RawMessage(`{"content": "pkg"}`),
			analyzer:    &mockFileAnalyzer{},
			wantAstScan: false,
			wantNonBlocking: true,
		},
		{
			name:        "tool input이 잘못된 JSON이면 ast_scan 없음",
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
