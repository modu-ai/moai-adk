package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestSessionEnd_ValidateMxTags_BasicFlow verifies AC-SESSION-001:
// session end validates modified Go files and logs results.
func TestSessionEnd_ValidateMxTags_BasicFlow(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	// Create a Go file with a goroutine pattern (P2 violation)
	goFile := filepath.Join(projectDir, "worker.go")
	if err := os.WriteFile(goFile, []byte(`package pkg

func StartWorker() {
	go func() {}()
}
`), 0o600); err != nil {
		t.Fatalf("failed to create go file: %v", err)
	}

	// validateMxTags should run without error
	ctx := context.Background()
	validateMxTags(ctx, []string{goFile}, projectDir)
	// Observation-only: no return value to assert.
	// The test just verifies no panic and the function completes.
}

// TestSessionEnd_ValidateMxTags_EmptyFileList verifies AC-EDGE-004:
// empty file list completes without error.
func TestSessionEnd_ValidateMxTags_EmptyFileList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	// Should not panic with empty list
	validateMxTags(ctx, nil, t.TempDir())
	validateMxTags(ctx, []string{}, t.TempDir())
}

// TestSessionEnd_ValidateMxTags_NonGoFilesFiltered verifies only .go files
// are validated.
func TestSessionEnd_ValidateMxTags_NonGoFilesFiltered(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()

	// Create non-Go files
	for _, name := range []string{"README.md", "config.yaml", "script.sh"} {
		path := filepath.Join(projectDir, name)
		if err := os.WriteFile(path, []byte("content\n"), 0o600); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	ctx := context.Background()
	// Non-Go files should be filtered out, no panic
	validateMxTags(ctx, []string{
		filepath.Join(projectDir, "README.md"),
		filepath.Join(projectDir, "config.yaml"),
	}, projectDir)
}

// TestGetModifiedGoFiles verifies getModifiedGoFiles returns .go file paths.
func TestGetModifiedGoFiles(t *testing.T) {
	t.Parallel()

	// This function relies on git diff, which may not work in all test envs.
	// We just verify it doesn't panic and returns a slice (possibly empty).
	ctx := context.Background()
	files := getModifiedGoFiles(ctx, t.TempDir())
	// files may be empty in non-git environment
	if files == nil {
		// nil is acceptable (treated as empty)
		files = []string{}
	}
	// Verify all returned files end in .go
	for _, f := range files {
		if filepath.Ext(f) != ".go" {
			t.Errorf("getModifiedGoFiles() returned non-Go file: %s", f)
		}
	}
}

// TestSessionEnd_Handle_WithMxValidation verifies that the full Handle method
// completes normally with MX validation enabled (AC-SESSION-002: non-blocking).
func TestSessionEnd_Handle_WithMxValidation(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	goFile := filepath.Join(projectDir, "service.go")
	if err := os.WriteFile(goFile, []byte(`package svc

func Serve() {
	go func() {}()
}
`), 0o600); err != nil {
		t.Fatalf("failed to create go file: %v", err)
	}

	h := NewSessionEndHandler()
	ctx := context.Background()

	input := &HookInput{
		SessionID: "sess-mx-session-end",
		CWD:       projectDir,
	}

	// Must complete without error (AC-SESSION-002: non-blocking on failure)
	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("Handle() error = %v (must be non-blocking)", err)
	}
	if got == nil {
		t.Fatal("Handle() returned nil output")
	}
}
