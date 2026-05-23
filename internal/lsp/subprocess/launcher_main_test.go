package subprocess_test

// @MX:NOTE: [AUTO] TestMain — package-wide shared fake binary initialization entry point
// @MX:SPEC: SPEC-LSP-FLAKY-001, SPEC-LSP-FLAKY-002
//
// This file defines TestMain, the entry point for the package test process.
// TestMain writes the shared fake binary *before* m.Run() is called, and any
// t.Parallel() tests only start after every writer fd has been closed. This
// fundamentally eliminates the Linux fork-exec ETXTBSY race.
//
// History:
// - SPEC-LSP-FLAKY-001 (first attempt): sync.OnceValues lazy initialization.
//   A race remained where supervisor_test goroutines, when forking during
//   t.Parallel(), briefly inherited the launcher OnceValues writer fd, so
//   flakes still appeared in Ubuntu CI.
// - SPEC-LSP-FLAKY-002 (current): eager write inside TestMain. By the time
//   m.Run() begins, the writer fd is already closed, so no fork by any
//   t.Parallel() goroutine can inherit the shared binary's writer fd.

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// pkgTempDir is the package-wide temporary directory created by TestMain.
// It is cleaned up via os.RemoveAll when the test process exits.
var pkgTempDir string

// pkgSharedBinaryPath is the absolute path of the shared fake LSP stub binary
// that TestMain writes before calling m.Run(). On Windows this is the empty
// string (callers handle this via t.Skip).
//
// @MX:ANCHOR: [AUTO] pkgSharedBinaryPath — shared binary path used to eliminate the ETXTBSY race
// @MX:REASON: fan_in >= 3 — TestLauncher_Launch_HappyPath, _StdioPipesNonNil, and _WithArgs all reference it
// @MX:SPEC: SPEC-LSP-FLAKY-002 REQ-001
//
// Core invariant: after TestMain exits, the file at this path is never rewritten.
// All callers must use it read+exec only and must not modify the file contents.
var pkgSharedBinaryPath string

// sharedFakeBinaryPath returns the package-wide shared fake LSP stub binary path.
//
// Usage conditions:
//   - Call this only in tests that use the binary read+exec only.
//   - Tests that modify the binary contents or permissions must use writeFakeBinary instead.
//
// @MX:SPEC: SPEC-LSP-FLAKY-002 REQ-001
func sharedFakeBinaryPath(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script stub은 Windows에서 지원되지 않음")
	}
	if pkgSharedBinaryPath == "" {
		t.Fatal("pkgSharedBinaryPath 가 초기화되지 않음 (TestMain 에서 setup 실패)")
	}
	return pkgSharedBinaryPath
}

// buildSharedBinary writes the fake LSP stub binary during the TestMain stage.
// Write order: Create → Write → Sync → Close → Chmod.
// At return, all writer fds are guaranteed to be closed.
//
// This function must be called *only* before m.Run(), at a point where no
// t.Parallel() goroutine has started. As a result, the fork-exec ETXTBSY race
// cannot occur.
func buildSharedBinary(dir string) (string, error) {
	if runtime.GOOS == "windows" {
		// Windows does not support the shell script stub; callers handle via t.Skip.
		return "", nil
	}
	path := filepath.Join(dir, "shared-fake-lsp")

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("buildSharedBinary create: %w", err)
	}
	if _, err := f.Write([]byte("#!/bin/sh\ncat\n")); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("buildSharedBinary write: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("buildSharedBinary sync: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("buildSharedBinary close: %w", err)
	}
	if err := os.Chmod(path, 0o755); err != nil {
		return "", fmt.Errorf("buildSharedBinary chmod: %w", err)
	}
	return path, nil
}

// TestMain is the entry point for the package test process.
//
// Responsibilities:
//  1. Create the package-wide temporary directory
//  2. Write the shared fake binary (before m.Run(), before t.Parallel() starts)
//  3. Run every test via m.Run() (with the writer fd already closed)
//  4. Clean up the temporary directory
//
// Core design principle: all binary writes must complete before m.Run() is
// called, so that the writer fd is closed. That way, even when t.Parallel()
// tests call fork-exec concurrently, no child process can inherit the
// shared-fake-lsp writer fd.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "moai-lsp-subprocess-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: 임시 디렉터리 생성 실패: %v\n", err)
		os.Exit(1)
	}
	pkgTempDir = dir

	// CRITICAL: eagerly write the binary *before* calling m.Run().
	// At this point no t.Parallel() goroutine has started, so the
	// fork-exec ETXTBSY race cannot occur.
	binPath, err := buildSharedBinary(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: 공유 fake binary 작성 실패: %v\n", err)
		_ = os.RemoveAll(dir)
		os.Exit(1)
	}
	pkgSharedBinaryPath = binPath

	code := m.Run()

	_ = os.RemoveAll(pkgTempDir)
	os.Exit(code)
}
