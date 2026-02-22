package hook

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook/security"
)

// TestScanWriteContent_InvalidJSON tests scanWriteContent with invalid JSON input.
func TestScanWriteContent_InvalidJSON(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()
	scanner := security.NewSecurityScanner()

	handler := &preToolHandler{
		cfg:        cfg,
		policy:     policy,
		scanner:    scanner,
		projectDir: t.TempDir(),
	}

	decision, reason := handler.scanWriteContent(context.Background(), json.RawMessage(`{invalid}`))
	if decision != "" {
		t.Errorf("expected empty decision for invalid JSON, got %q (reason: %q)", decision, reason)
	}
}

// TestScanWriteContent_MissingFilePath tests scanWriteContent when file_path is missing.
func TestScanWriteContent_MissingFilePath(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()
	scanner := security.NewSecurityScanner()

	handler := &preToolHandler{
		cfg:        cfg,
		policy:     policy,
		scanner:    scanner,
		projectDir: t.TempDir(),
	}

	toolInput, _ := json.Marshal(map[string]string{
		"content": "package main\nfunc main() {}",
	})
	decision, reason := handler.scanWriteContent(context.Background(), json.RawMessage(toolInput))
	if decision != "" {
		t.Errorf("expected empty decision for missing file_path, got %q (reason: %q)", decision, reason)
	}
}

// TestScanWriteContent_MissingContent tests scanWriteContent when content is missing.
func TestScanWriteContent_MissingContent(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()
	scanner := security.NewSecurityScanner()

	handler := &preToolHandler{
		cfg:        cfg,
		policy:     policy,
		scanner:    scanner,
		projectDir: t.TempDir(),
	}

	projectDir := t.TempDir()
	toolInput, _ := json.Marshal(map[string]string{
		"file_path": filepath.Join(projectDir, "test.go"),
	})
	decision, reason := handler.scanWriteContent(context.Background(), json.RawMessage(toolInput))
	if decision != "" {
		t.Errorf("expected empty decision for missing content, got %q (reason: %q)", decision, reason)
	}
}

// TestScanWriteContent_UnsupportedExtension tests scanWriteContent with unsupported file extension.
func TestScanWriteContent_UnsupportedExtension(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()
	scanner := security.NewSecurityScanner()

	handler := &preToolHandler{
		cfg:        cfg,
		policy:     policy,
		scanner:    scanner,
		projectDir: t.TempDir(),
	}

	projectDir := t.TempDir()
	toolInput, _ := json.Marshal(map[string]string{
		"file_path": filepath.Join(projectDir, "test.unknownext"),
		"content":   "some content here",
	})
	decision, reason := handler.scanWriteContent(context.Background(), json.RawMessage(toolInput))
	// Unsupported extension should result in empty decision (allow)
	if decision != "" {
		t.Errorf("expected empty decision for unsupported extension, got %q (reason: %q)", decision, reason)
	}
}

// TestScanWriteContent_SupportedExtension_NoScannerAvailable tests when scanner is not available.
func TestScanWriteContent_SupportedExtension_ScannerNotAvailable(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()
	scanner := security.NewSecurityScanner()

	// Only run if scanner is not available
	if scanner.IsAvailable() {
		t.Skip("ast-grep is available; test requires it to be absent")
	}

	handler := &preToolHandler{
		cfg:        cfg,
		policy:     policy,
		scanner:    scanner,
		projectDir: t.TempDir(),
	}

	projectDir := t.TempDir()
	toolInput, _ := json.Marshal(map[string]string{
		"file_path": filepath.Join(projectDir, "test.go"),
		"content":   "package main\nfunc main() {}",
	})
	decision, reason := handler.scanWriteContent(context.Background(), json.RawMessage(toolInput))
	// Without ast-grep available, scanner returns empty result
	if decision != "" {
		t.Errorf("expected empty decision when scanner unavailable, got %q (reason: %q)", decision, reason)
	}
}

// TestNewPreToolHandlerWithScanner_ScannerUnavailable tests the handler creation
// when the scanner IsAvailable() returns false.
func TestNewPreToolHandlerWithScanner_ScannerUnavailable(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	policy := DefaultSecurityPolicy()

	// Create a scanner that is always unavailable
	scanner := security.NewSecurityScanner()
	// If scanner is available, use nil to simulate unavailability
	var testScanner *security.SecurityScanner
	if scanner.IsAvailable() {
		// Scanner is available - test the path when passing non-nil available scanner
		testScanner = scanner
	}
	// else testScanner remains nil

	h := NewPreToolHandlerWithScanner(cfg, policy, testScanner)
	if h == nil {
		t.Fatal("NewPreToolHandlerWithScanner returned nil")
	}

	// Handler should still work
	ctx := context.Background()
	input := &HookInput{
		SessionID:     "test",
		ToolName:      "Read",
		HookEventName: "PreToolUse",
	}
	output, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}
	if output == nil {
		t.Fatal("Handle() returned nil output")
	}
}
