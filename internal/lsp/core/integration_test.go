//go:build integration

package core_test

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
	"github.com/modu-ai/moai-adk/internal/lsp/core"
)

// goplsConfig returns the standard ServerConfig used for gopls integration tests.
// Omits rootUri when rootDir is empty.
func goplsConfig() config.ServerConfig {
	return config.ServerConfig{
		Language:       "go",
		Command:        "gopls",
		Args:           []string{"serve"},
		FileExtensions: []string{".go"},
		RootMarkers:    []string{"go.mod"},
	}
}

// goplsConfigWithRoot returns a gopls ServerConfig with rootDir set.
// rootDir must be a temporary directory containing go.mod.
func goplsConfigWithRoot(rootDir string) config.ServerConfig {
	cfg := goplsConfig()
	cfg.RootDir = rootDir
	return cfg
}

// TestIntegration_Gopls_InitializeAndShutdown verifies that gopls starts, reaches StateReady,
// and shuts down cleanly (REQ-LC-010).
func TestIntegration_Gopls_InitializeAndShutdown(t *testing.T) {
	skipIfBinaryMissing(t, "gopls")

	before := runtime.NumGoroutine()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cl := core.NewClient(goplsConfig())

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

	// Check for goroutine leaks: allow up to 2 additional goroutines after shutdown
	// (temporary fluctuation from runtime-internal GC goroutines, etc. is allowed)
	after := runtime.NumGoroutine()
	if delta := after - before; delta > 2 {
		t.Errorf("goroutine leak suspected: before=%d after=%d delta=%d", before, after, delta)
	}
}

// TestIntegration_Gopls_OpenFileAndGetDiagnostics verifies that gopls pushes diagnostics
// for compile errors in a Go source file (REQ-LC-010).
func TestIntegration_Gopls_OpenFileAndGetDiagnostics(t *testing.T) {
	skipIfBinaryMissing(t, "gopls")

	// makeGoModuleDir: create under /tmp to bypass the macOS system temp root restriction
	tmpDir := makeGoModuleDir(t)

	// Source referencing an undefined variable: gopls must push a diagnostic
	mainContent := `package main

func main() {
	_ = undefinedVariable
}
`
	_, mainGoPath := writeGoModule(t, tmpDir, "testmod", mainContent)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set RootDir to tmpDir: passes rootUri+workspaceFolders to gopls
	cl := core.NewClient(goplsConfigWithRoot(tmpDir))

	if err := cl.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		_ = cl.Shutdown(shutCtx)
	})

	// Send didOpen
	if err := cl.OpenFile(ctx, mainGoPath, mainContent); err != nil {
		t.Fatalf("OpenFile: %v", err)
	}

	// Poll for diagnostics: gopls pushes publishDiagnostics asynchronously
	diags := waitForDiagnostics(ctx, cl, mainGoPath, 1, 100*time.Millisecond, 10*time.Second)
	if len(diags) == 0 {
		t.Fatal("expected at least 1 diagnostic from gopls for undefined variable, got 0")
	}

	// Verify that at least one diagnostic message contains "undefined" (case-insensitive)
	found := false
	for _, d := range diags {
		if strings.Contains(strings.ToLower(d.Message), "undefined") {
			found = true
			break
		}
	}
	if !found {
		msgs := make([]string, len(diags))
		for i, d := range diags {
			msgs[i] = d.Message
		}
		t.Errorf("expected diagnostic with 'undefined' in message, got: %v", msgs)
	}
}

// TestIntegration_Gopls_FindReferences verifies that gopls returns symbol reference locations
// (REQ-LC-010, REQ-LC-002b).
func TestIntegration_Gopls_FindReferences(t *testing.T) {
	skipIfBinaryMissing(t, "gopls")

	// makeGoModuleDir: create under /tmp to bypass the macOS system temp root restriction
	tmpDir := makeGoModuleDir(t)
	// foo() definition (line 2, 0-indexed) + foo() call (line 5, 0-indexed)
	mainContent := `package main

func foo() {}

func main() {
	foo()
}
`
	_, mainGoPath := writeGoModule(t, tmpDir, "testmod", mainContent)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set RootDir to tmpDir: passes rootUri+workspaceFolders to gopls
	cl := core.NewClient(goplsConfigWithRoot(tmpDir))
	if err := cl.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		_ = cl.Shutdown(shutCtx)
	})

	if err := cl.OpenFile(ctx, mainGoPath, mainContent); err != nil {
		t.Fatalf("OpenFile: %v", err)
	}

	// Wait for gopls indexing (initial analysis time needed after didOpen)
	time.Sleep(500 * time.Millisecond)

	// foo definition location (line=2, character=5): position of 'f' in "func foo()"
	locs, err := cl.FindReferences(ctx, mainGoPath, lspPosition(2, 5))
	if err != nil {
		// ErrCapabilityUnsupported means gopls version issue — skip
		if strings.Contains(err.Error(), "capability") {
			t.Skipf("gopls does not advertise references capability: %v", err)
		}
		t.Fatalf("FindReferences: %v", err)
	}

	if len(locs) < 2 {
		t.Errorf("FindReferences: expected >= 2 locations (definition + call site), got %d: %v", len(locs), locs)
	}
}

// TestIntegration_Gopls_GotoDefinition verifies that gopls returns the symbol definition location
// (REQ-LC-010, REQ-LC-002b).
func TestIntegration_Gopls_GotoDefinition(t *testing.T) {
	skipIfBinaryMissing(t, "gopls")

	// makeGoModuleDir: create under /tmp to bypass the macOS system temp root restriction
	tmpDir := makeGoModuleDir(t)
	// foo() definition (line 2) + foo() call (line 5)
	mainContent := `package main

func foo() {}

func main() {
	foo()
}
`
	_, mainGoPath := writeGoModule(t, tmpDir, "testmod", mainContent)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set RootDir to tmpDir: passes rootUri+workspaceFolders to gopls
	cl := core.NewClient(goplsConfigWithRoot(tmpDir))
	if err := cl.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		_ = cl.Shutdown(shutCtx)
	})

	if err := cl.OpenFile(ctx, mainGoPath, mainContent); err != nil {
		t.Fatalf("OpenFile: %v", err)
	}

	// Wait for gopls indexing
	time.Sleep(500 * time.Millisecond)

	// foo() call location (line=5, character=1): position of 'f' in "\tfoo()"
	locs, err := cl.GotoDefinition(ctx, mainGoPath, lspPosition(5, 1))
	if err != nil {
		if strings.Contains(err.Error(), "capability") {
			t.Skipf("gopls does not advertise definition capability: %v", err)
		}
		t.Fatalf("GotoDefinition: %v", err)
	}

	if len(locs) == 0 {
		t.Fatal("GotoDefinition: expected at least 1 location, got 0")
	}

	// Returned location must include the definition line (line=2)
	found := false
	for _, loc := range locs {
		if loc.Range.Start.Line == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GotoDefinition: expected location on line 2 (foo definition), got: %v", locs)
	}
}
