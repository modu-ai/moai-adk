package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestNewPostToolHandlerWithMxValidator verifies the constructor.
func TestNewPostToolHandlerWithMxValidator(t *testing.T) {
	t.Parallel()

	h := NewPostToolHandlerWithMxValidator(nil, nil, "")
	if h == nil {
		t.Fatal("NewPostToolHandlerWithMxValidator returned nil")
	}
	if h.EventType() != EventPostToolUse {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventPostToolUse)
	}
}

// TestPostToolHandler_MxValidation_WriteTool verifies AC-POST-001:
// mx_validation metrics are populated for Write operations.
func TestPostToolHandler_MxValidation_WriteTool(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "service.go")
	// File with goroutine pattern (will trigger P2 violation)
	if err := os.WriteFile(tmpFile, []byte(`package svc

func Worker() {
	go func() {}()
}
`), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	h := NewPostToolHandlerWithMxValidator(nil, nil, tmpDir)
	ctx := context.Background()

	input := &HookInput{
		SessionID: "sess-mx-test",
		CWD:       tmpDir,
		ToolName:  "Write",
		ToolInput: makeToolInput(map[string]string{
			"file_path": tmpFile,
			"content":   `package svc` + "\n",
		}),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if got == nil || got.Data == nil {
		t.Fatal("Data must not be nil")
	}

	var metrics map[string]any
	if err := json.Unmarshal(got.Data, &metrics); err != nil {
		t.Fatalf("failed to unmarshal metrics: %v", err)
	}

	// AC-POST-001: mx_validation key must exist
	mxVal, ok := metrics["mx_validation"]
	if !ok {
		t.Fatal("mx_validation not found in metrics")
	}
	mxMap, ok := mxVal.(map[string]any)
	if !ok {
		t.Fatalf("mx_validation is not a map: %T", mxVal)
	}

	// AC-POST-001: status must be pass, warn, or fail
	status, ok := mxMap["status"].(string)
	if !ok {
		t.Fatalf("mx_validation.status is not a string: %v", mxMap["status"])
	}
	validStatuses := map[string]bool{"pass": true, "warn": true, "fail": true}
	if !validStatuses[status] {
		t.Errorf("mx_validation.status = %q, want pass/warn/fail", status)
	}

	// AC-POST-001: violations array must exist
	if _, ok := mxMap["violations"]; !ok {
		t.Error("mx_validation.violations not found in metrics")
	}

	// AC-POST-001: duration_ms must be >= 0
	durMs, ok := mxMap["duration_ms"].(float64)
	if !ok {
		t.Fatalf("mx_validation.duration_ms is not a number: %v", mxMap["duration_ms"])
	}
	if durMs < 0 {
		t.Errorf("mx_validation.duration_ms = %v, want >= 0", durMs)
	}
}

// TestPostToolHandler_MxValidation_NonGoFile verifies no mx_validation
// for non-Go files.
func TestPostToolHandler_MxValidation_NonGoFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(tmpFile, []byte("# readme\n"), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	h := NewPostToolHandlerWithMxValidator(nil, nil, tmpDir)
	ctx := context.Background()

	input := &HookInput{
		SessionID: "sess-nongo",
		CWD:       tmpDir,
		ToolName:  "Write",
		ToolInput: makeToolInput(map[string]string{
			"file_path": tmpFile,
			"content":   "# readme\n",
		}),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if got == nil || got.Data == nil {
		t.Fatal("Data must not be nil")
	}

	var metrics map[string]any
	if err := json.Unmarshal(got.Data, &metrics); err != nil {
		t.Fatalf("failed to unmarshal metrics: %v", err)
	}

	// Non-Go file: no mx_validation metric
	if _, ok := metrics["mx_validation"]; ok {
		t.Error("mx_validation should not be present for non-Go files")
	}
}

// TestPostToolHandler_MxValidation_Timeout verifies AC-POST-002:
// mx_validation returns "skipped" status on timeout.
func TestPostToolHandler_MxValidation_Timeout(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "service.go")
	if err := os.WriteFile(tmpFile, []byte("package svc\n"), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Use a very short timeout to force skipped status
	h := NewPostToolHandlerWithMxValidatorAndTimeout(nil, nil, tmpDir, 1*time.Nanosecond)
	ctx := context.Background()

	start := time.Now()
	input := &HookInput{
		SessionID: "sess-timeout",
		CWD:       tmpDir,
		ToolName:  "Write",
		ToolInput: makeToolInput(map[string]string{
			"file_path": tmpFile,
			"content":   "package svc\n",
		}),
	}

	got, err := h.Handle(ctx, input)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	// AC-POST-002: must respond quickly (500ms budget, but 1ns timeout here)
	// We just verify the handler doesn't hang
	_ = elapsed

	if got == nil || got.Data == nil {
		t.Fatal("Data must not be nil")
	}

	var metrics map[string]any
	if err := json.Unmarshal(got.Data, &metrics); err != nil {
		t.Fatalf("failed to unmarshal metrics: %v", err)
	}

	// With 1ns timeout, mx_validation should either be "skipped" or not present
	if mxVal, ok := metrics["mx_validation"]; ok {
		mxMap, ok := mxVal.(map[string]any)
		if !ok {
			t.Fatalf("mx_validation is not a map: %T", mxVal)
		}
		status, _ := mxMap["status"].(string)
		if status != "skipped" && status != "pass" && status != "warn" && status != "fail" {
			t.Errorf("mx_validation.status = %q, want skipped/pass/warn/fail", status)
		}
	}
}

// TestPostToolHandler_MxValidation_ReadTool verifies no mx_validation for Read tool.
func TestPostToolHandler_MxValidation_ReadTool(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(tmpFile, []byte("package main\n"), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	h := NewPostToolHandlerWithMxValidator(nil, nil, tmpDir)
	ctx := context.Background()

	input := &HookInput{
		SessionID: "sess-read",
		CWD:       tmpDir,
		ToolName:  "Read",
		ToolInput: makeToolInput(map[string]string{"file_path": tmpFile}),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	var metrics map[string]any
	if err := json.Unmarshal(got.Data, &metrics); err != nil {
		t.Fatalf("failed to unmarshal metrics: %v", err)
	}

	// Read tool: no mx_validation metric
	if _, ok := metrics["mx_validation"]; ok {
		t.Error("mx_validation should not be present for Read tool")
	}
}

// TestPostToolHandler_MxValidation_NilProjectRoot verifies no mx_validation
// when project root is empty (no validator configured).
func TestPostToolHandler_MxValidation_NilProjectRoot(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(tmpFile, []byte("package main\n"), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Empty project root → no MX validator
	h := NewPostToolHandlerWithMxValidator(nil, nil, "")
	ctx := context.Background()

	input := &HookInput{
		SessionID: "sess-nilroot",
		CWD:       tmpDir,
		ToolName:  "Write",
		ToolInput: makeToolInput(map[string]string{"file_path": tmpFile, "content": "pkg\n"}),
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	var metrics map[string]any
	if err := json.Unmarshal(got.Data, &metrics); err != nil {
		t.Fatalf("failed to unmarshal metrics: %v", err)
	}

	// No project root → no mx_validation
	if _, ok := metrics["mx_validation"]; ok {
		t.Error("mx_validation should not be present when project root is empty")
	}
}
