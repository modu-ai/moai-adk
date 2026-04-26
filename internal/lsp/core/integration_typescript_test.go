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

// tsConfig returns a ServerConfig for typescript-language-server integration tests.
func tsConfig() config.ServerConfig {
	return config.ServerConfig{
		Language:       "typescript",
		Command:        "typescript-language-server",
		Args:           []string{"--stdio"},
		FileExtensions: []string{".ts", ".tsx"},
		RootMarkers:    []string{"tsconfig.json", "package.json"},
	}
}

// TestIntegration_TypeScript_InitializeAndShutdown actually launches
// typescript-language-server, waits for StateReady, and verifies normal shutdown.
//
// Skipped if typescript-language-server is not on PATH.
func TestIntegration_TypeScript_InitializeAndShutdown(t *testing.T) {
	skipIfBinaryMissing(t, "typescript-language-server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cl := core.NewClient(tsConfig())

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

// TestIntegration_TypeScript_OpenFileAndDiagnostics verifies that
// typescript-language-server pushes diagnostics for type errors in a TypeScript
// source file.
//
// Skipped if typescript-language-server is not on PATH.
func TestIntegration_TypeScript_OpenFileAndDiagnostics(t *testing.T) {
	skipIfBinaryMissing(t, "typescript-language-server")

	tmpDir := t.TempDir()

	// Create a TypeScript file with a type error.
	tsContent := "const x: number = \"string\";\n"
	tsPath := writeTempFile(t, tmpDir, "main.ts", tsContent)

	// Create tsconfig.json (root marker).
	writeTempFile(t, tmpDir, "tsconfig.json", `{"compilerOptions":{"strict":true}}`+"\n")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cl := core.NewClient(tsConfig())

	if err := cl.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		_ = cl.Shutdown(shutCtx)
	})

	if err := cl.OpenFile(ctx, tsPath, tsContent); err != nil {
		t.Fatalf("OpenFile: %v", err)
	}

	// Poll for diagnostics: wait up to 10s.
	diags := waitForDiagnostics(ctx, cl, tsPath, 1, 200*time.Millisecond, 10*time.Second)
	if len(diags) == 0 {
		t.Fatal("expected at least 1 diagnostic from typescript-language-server for type mismatch, got 0")
	}

	// Verify the diagnostic message contains type-error keywords (case-insensitive).
	typeErrorKeywords := []string{"type", "string", "number", "assignable"}
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
		t.Errorf("expected type-error diagnostic, got: %v", msgs)
	}
}
