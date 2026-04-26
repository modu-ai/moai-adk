//go:build integration

package core_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
	"github.com/modu-ai/moai-adk/internal/lsp/core"
)

// pyrightConfig returns the ServerConfig for pyright-langserver integration tests.
func pyrightConfig() config.ServerConfig {
	return config.ServerConfig{
		Language:       "python",
		Command:        "pyright-langserver",
		Args:           []string{"--stdio"},
		FileExtensions: []string{".py"},
		RootMarkers:    []string{"pyproject.toml", "setup.py", "setup.cfg"},
	}
}

// pyrightConfigWithRoot returns a pyright ServerConfig with rootDir set.
func pyrightConfigWithRoot(rootDir string) config.ServerConfig {
	cfg := pyrightConfig()
	cfg.RootDir = rootDir
	return cfg
}

// TestIntegration_Pyright_InitializeAndShutdown verifies that pyright-langserver starts,
// reaches StateReady, and shuts down cleanly.
func TestIntegration_Pyright_InitializeAndShutdown(t *testing.T) {
	skipIfBinaryMissing(t, "pyright-langserver")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cl := core.NewClient(pyrightConfig())

	if err := cl.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	assertStateEquals(t, cl, core.StateReady)

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()

	if err := cl.Shutdown(shutCtx); err != nil {
		t.Errorf("Shutdown: %v", err)
	}
	assertStateEquals(t, cl, core.StateShutdown)
}

// TestIntegration_Pyright_OpenFileAndGetDiagnostics verifies that pyright-langserver pushes
// diagnostics for type errors in a Python source file.
//
// Pyright has a longer initial analysis time, so a polling window of up to 15s is used.
func TestIntegration_Pyright_OpenFileAndGetDiagnostics(t *testing.T) {
	skipIfBinaryMissing(t, "pyright-langserver")

	tmpDir := t.TempDir()

	// Create a Python file with a type error
	// pyright statically detects type mismatches
	pyContent := `x: int = "this is a string"
`
	pyPath := writeTempFile(t, tmpDir, "main.py", pyContent)

	// Create pyproject.toml (pyright root marker)
	writeTempFile(t, tmpDir, "pyproject.toml", "[tool.pyright]\n")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cl core.Client
	var startErr error

	// Retry pyright startup (up to 2 times) in case the first attempt fails.
	// Set RootDir to tmpDir to pass the workspace root to pyright.
	for attempt := 0; attempt < 2; attempt++ {
		cl = core.NewClient(pyrightConfigWithRoot(tmpDir))
		startErr = cl.Start(ctx)
		if startErr == nil {
			break
		}
		t.Logf("Pyright Start attempt %d failed: %v — retrying", attempt+1, startErr)
		time.Sleep(500 * time.Millisecond)
	}
	if startErr != nil {
		t.Fatalf("Start (after 2 attempts): %v", startErr)
	}

	t.Cleanup(func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		_ = cl.Shutdown(shutCtx)
	})

	if err := cl.OpenFile(ctx, pyPath, pyContent); err != nil {
		t.Fatalf("OpenFile: %v", err)
	}

	// Poll for diagnostics: pyright is a slower analyzer than gopls — wait up to 15s
	diags := waitForDiagnostics(ctx, cl, pyPath, 1, 200*time.Millisecond, 15*time.Second)
	if len(diags) == 0 {
		t.Fatal("expected at least 1 diagnostic from pyright for type mismatch, got 0")
	}

	// Verify that at least one diagnostic contains a type-error keyword (case-insensitive)
	// pyright outputs messages like "Expression of type" or "Cannot assign"
	typeErrorKeywords := []string{"type", "assign", "str", "int", "expression"}
	found := false
outer:
	for _, d := range diags {
		lower := strings.ToLower(d.Message)
		for _, kw := range typeErrorKeywords {
			if strings.Contains(lower, kw) {
				found = true
				break outer
			}
		}
	}
	if !found {
		msgs := make([]string, len(diags))
		for i, d := range diags {
			msgs[i] = d.Message
		}
		t.Errorf("expected type-error diagnostic from pyright, got: %v", msgs)
	}
}
