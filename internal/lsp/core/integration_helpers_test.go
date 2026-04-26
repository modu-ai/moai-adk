//go:build integration

package core_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/core"
)

// skipIfBinaryMissing skips the test when the cmd binary is not found in PATH.
func skipIfBinaryMissing(t *testing.T, cmd string) {
	t.Helper()
	if _, err := exec.LookPath(cmd); err != nil {
		t.Skipf("integration: binary %q not found in PATH — skipping (install to run this test)", cmd)
	}
}

// waitForDiagnostics polls GetDiagnostics at pollInterval intervals and waits up to totalTimeout
// until at least minCount diagnostics are returned.
// Returns an empty slice if the condition is not met within totalTimeout.
func waitForDiagnostics(
	ctx context.Context,
	cl core.Client,
	path string,
	minCount int,
	pollInterval, totalTimeout time.Duration,
) []lsp.Diagnostic {
	deadline := time.Now().Add(totalTimeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		diags, err := cl.GetDiagnostics(ctx, path)
		if err == nil && len(diags) >= minCount {
			return diags
		}
		time.Sleep(pollInterval)
	}
	// One final attempt
	diags, err := cl.GetDiagnostics(ctx, path)
	if err == nil {
		return diags
	}
	return nil
}

// writeTempFile writes content to dir/name and returns the absolute path.
func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTempFile: failed to write %s: %v", path, err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("writeTempFile: failed to resolve absolute path for %s: %v", path, err)
	}
	return abs
}

// assertStateEquals verifies that Client.State() equals expected.
func assertStateEquals(t *testing.T, cl core.Client, expected core.ClientState) {
	t.Helper()
	got := cl.State()
	if got != expected {
		t.Errorf("client state: got %q, want %q", got, expected)
	}
}

// lspPosition converts 0-based line and character to an lsp.Position.
func lspPosition(line, character int) lsp.Position {
	return lsp.Position{Line: line, Character: character}
}

// writeGoModule creates go.mod and main.go in tmpDir and returns their absolute paths.
// If mainContent is empty, a default main.go is used.
func writeGoModule(t *testing.T, tmpDir, moduleName, mainContent string) (goModPath, mainGoPath string) {
	t.Helper()

	goModContent := fmt.Sprintf("module %s\n\ngo 1.21\n", moduleName)
	goModPath = writeTempFile(t, tmpDir, "go.mod", goModContent)

	if mainContent == "" {
		mainContent = "package main\n\nfunc main() {}\n"
	}
	mainGoPath = writeTempFile(t, tmpDir, "main.go", mainContent)
	return goModPath, mainGoPath
}

// makeGoModuleDir creates a temporary directory under /tmp for a Go module that gopls can recognize.
//
// On macOS, os.TempDir() returns /var/folders/..., and Go tools ignore go.mod files within the
// system temp root. This function explicitly creates the directory under /tmp so gopls
// correctly recognizes the Go module.
//
// Automatically deleted via t.Cleanup.
func makeGoModuleDir(t *testing.T) string {
	t.Helper()
	// Create under /tmp: bypass the macOS system temp root (/var/folders)
	dir, err := os.MkdirTemp("/tmp", "lsp_integration_*")
	if err != nil {
		t.Fatalf("makeGoModuleDir: MkdirTemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	// Resolve symlinks: ensures consistency because gopls constructs URIs from the real path
	realDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return dir
	}
	return realDir
}
