package hook

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		name        string
		input       *HookInput
		wantFile    bool // whether task-metrics.jsonl should be created
		wantTokens  int
		wantTools   int
		wantSeconds float64
	}{
		{
			name: "task tool with valid metrics creates JSONL record",
			input: &HookInput{
				SessionID:    "sess-metrics-001",
				ToolName:     "Task",
				ToolResponse: json.RawMessage(`{"status":"completed","output":"done","metrics":{"tokensUsed":12450,"toolUses":8,"durationSeconds":45.2}}`),
			},
			wantFile:    true,
			wantTokens:  12450,
			wantTools:   8,
			wantSeconds: 45.2,
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
			tmpDir := t.TempDir()
			tt.input.CWD = tmpDir

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
			defer f.Close()

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
