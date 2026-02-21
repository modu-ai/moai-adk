package hook

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
)

func TestPostToolHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewPostToolHandler()

	if got := h.EventType(); got != EventPostToolUse {
		t.Errorf("EventType() = %q, want %q", got, EventPostToolUse)
	}
}

func TestPostToolHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        *HookInput
		wantDecision string
		checkData    bool
	}{
		{
			name: "normal tool output with metrics",
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PostToolUse",
				ToolName:      "Write",
				ToolInput:     json.RawMessage(`{"file_path": "main.go"}`),
				ToolOutput:    json.RawMessage(`{"success": true, "path": "main.go"}`),
			},
			wantDecision: DecisionAllow,
			checkData:    true,
		},
		{
			name: "empty tool output",
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PostToolUse",
				ToolName:      "Read",
			},
			wantDecision: DecisionAllow,
		},
		{
			name: "large tool output",
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PostToolUse",
				ToolName:      "Bash",
				ToolOutput:    json.RawMessage(`{"output": "` + strings.Repeat("x", 10000) + `"}`),
			},
			wantDecision: DecisionAllow,
			checkData:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewPostToolHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			// PostToolUse handler is observation-only and uses hookSpecificOutput.hookEventName
			if got.HookSpecificOutput == nil {
				t.Fatal("HookSpecificOutput is nil")
			}
			if got.HookSpecificOutput.HookEventName != "PostToolUse" {
				t.Errorf("HookEventName = %q, want %q", got.HookSpecificOutput.HookEventName, "PostToolUse")
			}
			if tt.checkData && got.Data != nil {
				if !json.Valid(got.Data) {
					t.Errorf("Data is not valid JSON: %s", got.Data)
				}
			}
		})
	}
}

func TestLogTaskMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        *HookInput
		setupMoaiDir bool // whether to pre-create .moai/ so resolveProjectRoot accepts tmpDir
		wantFile     bool // whether task-metrics.jsonl should be created
		wantTokens   int
		wantTools    int
		wantSeconds  float64
	}{
		{
			name: "task tool with valid metrics creates JSONL record",
			input: &HookInput{
				SessionID:    "sess-metrics-001",
				ToolName:     "Task",
				ToolResponse: json.RawMessage(`{"status":"completed","output":"done","metrics":{"tokensUsed":12450,"toolUses":8,"durationSeconds":45.2}}`),
			},
			setupMoaiDir: true,
			wantFile:     true,
			wantTokens:   12450,
			wantTools:    8,
			wantSeconds:  45.2,
		},
		{
			name: "task tool with valid metrics but no .moai dir writes no file",
			input: &HookInput{
				SessionID:    "sess-metrics-guard",
				ToolName:     "Task",
				ToolResponse: json.RawMessage(`{"status":"completed","output":"done","metrics":{"tokensUsed":100,"toolUses":1,"durationSeconds":1.0}}`),
			},
			setupMoaiDir: false,
			wantFile:     false,
		},
		{
			name: "task tool with missing metrics field writes no file",
			input: &HookInput{
				SessionID:    "sess-metrics-002",
				ToolName:     "Task",
				ToolResponse: json.RawMessage(`{"status":"completed","output":"done"}`),
			},
			wantFile: false,
		},
		{
			name: "non-task tool writes no file",
			input: &HookInput{
				SessionID:    "sess-metrics-003",
				ToolName:     "Write",
				ToolResponse: json.RawMessage(`{"success":true}`),
			},
			wantFile: false,
		},
		{
			name: "task tool with empty ToolResponse writes no file",
			input: &HookInput{
				SessionID: "sess-metrics-004",
				ToolName:  "Task",
			},
			wantFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Use a temp directory as CWD for isolation.
			// CLAUDE_PROJECT_DIR is not set in normal test runs, so resolveProjectRoot
			// falls back to input.CWD. Pre-create .moai/ when the test expects success.
			tmpDir := t.TempDir()
			tt.input.CWD = tmpDir

			if tt.setupMoaiDir {
				if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
					t.Fatalf("pre-create .moai: %v", err)
				}
			}

			logPath := filepath.Join(tmpDir, ".moai", "logs", "task-metrics.jsonl")

			// Only Task tool calls logTaskMetrics; replicate Handle() routing.
			if tt.input.ToolName == "Task" {
				logTaskMetrics(tt.input)
			}

			_, statErr := os.Stat(logPath)
			fileExists := statErr == nil

			if tt.wantFile && !fileExists {
				t.Fatalf("expected JSONL file at %s, but it was not created", logPath)
			}
			if !tt.wantFile && fileExists {
				t.Fatalf("expected no JSONL file, but found one at %s", logPath)
			}

			if !tt.wantFile {
				return
			}

			// Parse the single JSONL line.
			f, err := os.Open(logPath)
			if err != nil {
				t.Fatalf("failed to open JSONL file: %v", err)
			}
			defer func() { _ = f.Close() }()

			scanner := bufio.NewScanner(f)
			if !scanner.Scan() {
				t.Fatal("JSONL file is empty")
			}
			line := scanner.Bytes()

			var rec taskMetricsRecord
			if err := json.Unmarshal(line, &rec); err != nil {
				t.Fatalf("failed to unmarshal JSONL record: %v", err)
			}

			if rec.SessionID != tt.input.SessionID {
				t.Errorf("session_id = %q, want %q", rec.SessionID, tt.input.SessionID)
			}
			if rec.ToolName != "Task" {
				t.Errorf("tool_name = %q, want %q", rec.ToolName, "Task")
			}
			if rec.TokensUsed != tt.wantTokens {
				t.Errorf("tokens_used = %d, want %d", rec.TokensUsed, tt.wantTokens)
			}
			if rec.ToolUses != tt.wantTools {
				t.Errorf("tool_uses = %d, want %d", rec.ToolUses, tt.wantTools)
			}
			if rec.DurationSeconds != tt.wantSeconds {
				t.Errorf("duration_seconds = %f, want %f", rec.DurationSeconds, tt.wantSeconds)
			}
			if rec.Timestamp == "" {
				t.Error("timestamp should not be empty")
			}
		})
	}
}

func TestPostToolHandler_Handle_TaskMetrics_DoesNotFail(t *testing.T) {
	t.Parallel()

	// Even when Task metrics logging would fail (e.g. bad CWD),
	// the hook must still return a successful allow response.
	tmpDir := t.TempDir()
	input := &HookInput{
		SessionID: "sess-resilience-001",
		CWD:       tmpDir,
		ToolName:  "Task",
		// Intentionally malformed JSON to trigger parse failure.
		ToolResponse: json.RawMessage(`{not valid json`),
	}

	h := NewPostToolHandler()
	ctx := context.Background()
	got, err := h.Handle(ctx, input)

	if err != nil {
		t.Fatalf("Handle() returned unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("Handle() returned nil output")
	}
	if got.HookSpecificOutput == nil || got.HookSpecificOutput.HookEventName != "PostToolUse" {
		t.Errorf("expected PostToolUse output, got: %+v", got)
	}
}

func TestNewPostToolHandlerWithDiagnostics_NilDiagnostics(t *testing.T) {
	t.Parallel()

	h := NewPostToolHandlerWithDiagnostics(nil)
	if h == nil {
		t.Fatal("NewPostToolHandlerWithDiagnostics(nil) returned nil")
	}
	if h.EventType() != EventPostToolUse {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventPostToolUse)
	}
}

func TestPostToolHandler_Handle_WithInputAndOutput(t *testing.T) {
	t.Parallel()

	h := NewPostToolHandler()
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-both-sizes",
		CWD:           "/tmp",
		HookEventName: "PostToolUse",
		ToolName:      "Edit",
		ToolInput:     json.RawMessage(`{"file_path": "/tmp/main.go", "old_string": "a", "new_string": "b"}`),
		ToolOutput:    json.RawMessage(`{"result": "success"}`),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if got == nil || got.Data == nil {
		t.Fatal("expected non-nil Data with both input and output")
	}

	var metrics map[string]any
	if err := json.Unmarshal(got.Data, &metrics); err != nil {
		t.Fatalf("unmarshal Data: %v", err)
	}
	if _, ok := metrics["input_size"]; !ok {
		t.Error("metrics should contain input_size")
	}
	if _, ok := metrics["output_size"]; !ok {
		t.Error("metrics should contain output_size")
	}
	if metrics["tool_name"] != "Edit" {
		t.Errorf("tool_name = %v, want Edit", metrics["tool_name"])
	}
	if metrics["session_id"] != "sess-both-sizes" {
		t.Errorf("session_id = %v, want sess-both-sizes", metrics["session_id"])
	}
}

func TestPostToolHandler_Handle_TaskToolRoutesToLogTaskMetrics(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Pre-create .moai/ so resolveProjectRoot accepts tmpDir as a MoAI project root.
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatalf("pre-create .moai: %v", err)
	}
	h := NewPostToolHandler()
	ctx := context.Background()

	input := &HookInput{
		SessionID:    "sess-task-route",
		CWD:          tmpDir,
		ToolName:     "Task",
		ToolResponse: json.RawMessage(`{"status":"completed","output":"done","metrics":{"tokensUsed":500,"toolUses":3,"durationSeconds":12.5}}`),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if got == nil {
		t.Fatal("Handle() returned nil")
	}

	// Verify the metrics JSONL was created
	logPath := filepath.Join(tmpDir, ".moai", "logs", "task-metrics.jsonl")
	if _, err := os.Stat(logPath); err != nil {
		t.Errorf("task-metrics.jsonl should be created: %v", err)
	}

	// Verify JSONL content
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read JSONL: %v", err)
	}

	var rec taskMetricsRecord
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	if !scanner.Scan() {
		t.Fatal("JSONL file is empty")
	}
	if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
		t.Fatalf("unmarshal JSONL: %v", err)
	}
	if rec.TokensUsed != 500 {
		t.Errorf("tokens_used = %d, want 500", rec.TokensUsed)
	}
	if rec.ToolUses != 3 {
		t.Errorf("tool_uses = %d, want 3", rec.ToolUses)
	}
}

func TestLogTaskMetrics_AppendMultipleRecords(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Pre-create .moai/ so resolveProjectRoot accepts tmpDir as a MoAI project root.
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatalf("pre-create .moai: %v", err)
	}

	for i := range 3 {
		input := &HookInput{
			SessionID:    fmt.Sprintf("sess-multi-%d", i),
			CWD:          tmpDir,
			ToolName:     "Task",
			ToolResponse: json.RawMessage(fmt.Sprintf(`{"metrics":{"tokensUsed":%d,"toolUses":%d,"durationSeconds":%d}}`, (i+1)*100, i+1, i+1)),
		}
		logTaskMetrics(input)
	}

	logPath := filepath.Join(tmpDir, ".moai", "logs", "task-metrics.jsonl")
	f, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("open JSONL: %v", err)
	}
	defer func() { _ = f.Close() }()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		count++
		var rec taskMetricsRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			t.Fatalf("unmarshal record %d: %v", count, err)
		}
		if rec.SessionID == "" {
			t.Errorf("record %d: session_id should not be empty", count)
		}
	}
	if count != 3 {
		t.Errorf("expected 3 records, got %d", count)
	}
}

// --- collectDiagnostics mock and tests ---

type mockDiagnosticsCollector struct {
	getDiagnosticsFunc   func(ctx context.Context, filePath string) ([]lsphook.Diagnostic, error)
	getSeverityCountFunc func(diagnostics []lsphook.Diagnostic) lsphook.SeverityCounts
}

func (m *mockDiagnosticsCollector) GetDiagnostics(ctx context.Context, filePath string) ([]lsphook.Diagnostic, error) {
	if m.getDiagnosticsFunc != nil {
		return m.getDiagnosticsFunc(ctx, filePath)
	}
	return nil, nil
}

func (m *mockDiagnosticsCollector) GetSeverityCounts(diagnostics []lsphook.Diagnostic) lsphook.SeverityCounts {
	if m.getSeverityCountFunc != nil {
		return m.getSeverityCountFunc(diagnostics)
	}
	return lsphook.SeverityCounts{}
}

func TestPostToolHandler_CollectDiagnostics_WriteToolWithErrors(t *testing.T) {
	t.Parallel()

	mock := &mockDiagnosticsCollector{
		getDiagnosticsFunc: func(_ context.Context, _ string) ([]lsphook.Diagnostic, error) {
			return []lsphook.Diagnostic{
				{Message: "undefined: foo", Severity: lsphook.SeverityError},
				{Message: "unused variable", Severity: lsphook.SeverityWarning},
			}, nil
		},
		getSeverityCountFunc: func(_ []lsphook.Diagnostic) lsphook.SeverityCounts {
			return lsphook.SeverityCounts{Errors: 1, Warnings: 1}
		},
	}

	h := NewPostToolHandlerWithDiagnostics(mock)
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-diag-errors",
		CWD:           "/tmp",
		HookEventName: "PostToolUse",
		ToolName:      "Write",
		ToolInput:     json.RawMessage(`{"file_path": "/tmp/test.go", "content": "package main"}`),
		ToolOutput:    json.RawMessage(`{"success": true}`),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if got == nil || got.Data == nil {
		t.Fatal("expected non-nil Data")
	}

	var metrics map[string]any
	if err := json.Unmarshal(got.Data, &metrics); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	diag, ok := metrics["lsp_diagnostics"]
	if !ok {
		t.Fatal("expected lsp_diagnostics in metrics")
	}
	diagMap, ok := diag.(map[string]any)
	if !ok {
		t.Fatalf("lsp_diagnostics is not a map, got %T", diag)
	}
	if diagMap["errors"] != float64(1) {
		t.Errorf("errors = %v, want 1", diagMap["errors"])
	}
	if diagMap["warnings"] != float64(1) {
		t.Errorf("warnings = %v, want 1", diagMap["warnings"])
	}
	if diagMap["has_issues"] != true {
		t.Errorf("has_issues = %v, want true", diagMap["has_issues"])
	}
}

func TestPostToolHandler_CollectDiagnostics_EditToolClean(t *testing.T) {
	t.Parallel()

	mock := &mockDiagnosticsCollector{
		getDiagnosticsFunc: func(_ context.Context, _ string) ([]lsphook.Diagnostic, error) {
			return []lsphook.Diagnostic{}, nil
		},
		getSeverityCountFunc: func(_ []lsphook.Diagnostic) lsphook.SeverityCounts {
			return lsphook.SeverityCounts{}
		},
	}

	h := NewPostToolHandlerWithDiagnostics(mock)
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-diag-clean",
		CWD:           "/tmp",
		HookEventName: "PostToolUse",
		ToolName:      "Edit",
		ToolInput:     json.RawMessage(`{"file_path": "/tmp/clean.go", "old_string": "a", "new_string": "b"}`),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	var metrics map[string]any
	if err := json.Unmarshal(got.Data, &metrics); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	diag, ok := metrics["lsp_diagnostics"]
	if !ok {
		t.Fatal("expected lsp_diagnostics in metrics")
	}
	diagMap := diag.(map[string]any)
	if diagMap["has_issues"] != false {
		t.Errorf("has_issues = %v, want false", diagMap["has_issues"])
	}
}

func TestPostToolHandler_CollectDiagnostics_Error(t *testing.T) {
	t.Parallel()

	mock := &mockDiagnosticsCollector{
		getDiagnosticsFunc: func(_ context.Context, _ string) ([]lsphook.Diagnostic, error) {
			return nil, errors.New("LSP unavailable")
		},
	}

	h := NewPostToolHandlerWithDiagnostics(mock)
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-diag-err",
		CWD:           "/tmp",
		HookEventName: "PostToolUse",
		ToolName:      "Write",
		ToolInput:     json.RawMessage(`{"file_path": "/tmp/err.go", "content": "pkg"}`),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	// Should still succeed (observation only)
	if got == nil {
		t.Fatal("expected non-nil output")
	}

	var metrics map[string]any
	if err := json.Unmarshal(got.Data, &metrics); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// No lsp_diagnostics when collector fails
	if _, ok := metrics["lsp_diagnostics"]; ok {
		t.Error("expected no lsp_diagnostics when collector errors")
	}
}

func TestPostToolHandler_CollectDiagnostics_InvalidToolInput(t *testing.T) {
	t.Parallel()

	mock := &mockDiagnosticsCollector{}
	h := NewPostToolHandlerWithDiagnostics(mock)
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-diag-badinput",
		CWD:           "/tmp",
		HookEventName: "PostToolUse",
		ToolName:      "Write",
		ToolInput:     json.RawMessage(`{invalid json`),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil output even with bad input")
	}
}

func TestPostToolHandler_CollectDiagnostics_NoFilePath(t *testing.T) {
	t.Parallel()

	mock := &mockDiagnosticsCollector{}
	h := NewPostToolHandlerWithDiagnostics(mock)
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-diag-nopath",
		CWD:           "/tmp",
		HookEventName: "PostToolUse",
		ToolName:      "Edit",
		ToolInput:     json.RawMessage(`{"old_string": "a", "new_string": "b"}`),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil output")
	}
}

func TestPostToolHandler_NonWriteEditTool_NoDiagnostics(t *testing.T) {
	t.Parallel()

	mock := &mockDiagnosticsCollector{
		getDiagnosticsFunc: func(_ context.Context, _ string) ([]lsphook.Diagnostic, error) {
			t.Error("GetDiagnostics should not be called for non-Write/Edit tools")
			return nil, nil
		},
	}

	h := NewPostToolHandlerWithDiagnostics(mock)
	ctx := context.Background()

	input := &HookInput{
		SessionID:     "sess-diag-read",
		CWD:           "/tmp",
		HookEventName: "PostToolUse",
		ToolName:      "Read",
		ToolInput:     json.RawMessage(`{"file_path": "/tmp/test.go"}`),
	}

	_, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
}
