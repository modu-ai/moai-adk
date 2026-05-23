//go:build !windows

// SPEC-V3R6-WORKTREE-TEAM-LAUNCH-001 — M2 launchP3 POSIX execution tests.
//
// Verifies the in-process P3 dispatch path against AC-WTL-003. The test
// injects both `syscallExecFn` and `lookPathFn` to capture argv / cwd / env
// without actually replacing the process and without depending on the `moai`
// binary being present in $PATH (CI containers usually lack it).
//
// Windows variant: see team_launch_windows_test.go (handoff fallback).

package worktree

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLaunchP3_CapturesArgvAndCwd_CC verifies P3 (no-tmux) dispatch performs:
//
//  1. lookPathFn("moai") to resolve binary path
//  2. os.Chdir(cfg.WorktreePath) — the new process starts in the worktree
//  3. syscallExecFn(binPath, ["moai", llm], env) — actual replacement call
//
// AC-WTL-003: argv must equal [moai, cc] for cc mode.
func TestLaunchP3_CapturesArgvAndCwd_CC(t *testing.T) {
	// Save and restore injection seams + cwd so the test does not leak.
	origExec := syscallExecFn
	origLookPath := lookPathFn
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	defer func() {
		syscallExecFn = origExec
		lookPathFn = origLookPath
		_ = os.Chdir(origCwd)
	}()

	// Pretend `moai` lives at a fixed path — independent of $PATH state.
	lookPathFn = func(name string) (string, error) {
		if name != "moai" {
			return "", os.ErrNotExist
		}
		return "/fake/bin/moai", nil
	}
	// Capturing fake — record args without actually replacing the process.
	var capturedBin string
	var capturedArgs []string
	var capturedEnv []string
	syscallExecFn = func(bin string, args []string, env []string) error {
		capturedBin = bin
		capturedArgs = args
		capturedEnv = env
		return nil
	}

	wt := t.TempDir()
	cfg := TeamLaunchConfig{
		Pattern:      PatternP3InProgress,
		WorktreePath: wt,
		LLM:          "cc",
	}

	if err := launchP3(cfg); err != nil {
		t.Fatalf("launchP3 returned unexpected error: %v", err)
	}

	if capturedBin != "/fake/bin/moai" {
		t.Errorf("captured binary path = %q, want /fake/bin/moai", capturedBin)
	}
	if want := []string{"moai", "cc"}; !equalStringSlices(capturedArgs, want) {
		t.Errorf("captured argv = %v, want %v", capturedArgs, want)
	}
	// After Chdir, cwd should match the worktree (resolve symlinks because
	// t.TempDir() on macOS lives under /var/folders/... which is a symlink to
	// /private/var/folders/...).
	gotCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd post-launch: %v", err)
	}
	resolved, err := absoluteResolved(wt)
	if err != nil {
		t.Fatalf("resolve wt path: %v", err)
	}
	resolvedCwd, err := absoluteResolved(gotCwd)
	if err != nil {
		t.Fatalf("resolve cwd: %v", err)
	}
	if resolvedCwd != resolved {
		t.Errorf("cwd after launchP3 = %q, want %q", resolvedCwd, resolved)
	}
	if len(capturedEnv) == 0 {
		t.Errorf("captured env must be non-empty (os.Environ inherited)")
	}
}

// TestLaunchP3_CapturesArgvAndCwd_GLM mirrors the CC test for GLM mode —
// argv must be [moai, glm].
func TestLaunchP3_CapturesArgvAndCwd_GLM(t *testing.T) {
	origExec := syscallExecFn
	origLookPath := lookPathFn
	origCwd, _ := os.Getwd()
	defer func() {
		syscallExecFn = origExec
		lookPathFn = origLookPath
		_ = os.Chdir(origCwd)
	}()

	lookPathFn = func(name string) (string, error) {
		return "/fake/bin/moai", nil
	}
	var capturedArgs []string
	syscallExecFn = func(bin string, args []string, env []string) error {
		capturedArgs = args
		return nil
	}

	cfg := TeamLaunchConfig{
		Pattern:      PatternP3InProgress,
		WorktreePath: t.TempDir(),
		LLM:          "glm",
	}
	if err := launchP3(cfg); err != nil {
		t.Fatalf("launchP3 returned unexpected error: %v", err)
	}
	if want := []string{"moai", "glm"}; !equalStringSlices(capturedArgs, want) {
		t.Errorf("captured argv = %v, want %v", capturedArgs, want)
	}
}

// TestLaunchP3_ChdirFails verifies that an invalid worktree path produces a
// chdir error before any exec attempt is made.
func TestLaunchP3_ChdirFails(t *testing.T) {
	origExec := syscallExecFn
	origLookPath := lookPathFn
	origCwd, _ := os.Getwd()
	defer func() {
		syscallExecFn = origExec
		lookPathFn = origLookPath
		_ = os.Chdir(origCwd)
	}()

	lookPathFn = func(name string) (string, error) {
		return "/fake/bin/moai", nil
	}
	execCalled := false
	syscallExecFn = func(bin string, args []string, env []string) error {
		execCalled = true
		return nil
	}

	cfg := TeamLaunchConfig{
		Pattern:      PatternP3InProgress,
		WorktreePath: "/this/path/does/not/exist/" + t.Name(),
		LLM:          "cc",
	}
	err := launchP3(cfg)
	if err == nil {
		t.Errorf("launchP3 with invalid path should return error")
	}
	if execCalled {
		t.Errorf("syscallExecFn must NOT be called when chdir fails")
	}
}

// TestLaunchP3_LookPathFails verifies that when `moai` is not in PATH, no
// chdir is performed and no exec is attempted. Acceptance.md §2 edge case:
// "claude binary not in PATH (P3) → syscall.Exec LookPath fails; print
// error on stderr; exit non-zero".
func TestLaunchP3_LookPathFails(t *testing.T) {
	origExec := syscallExecFn
	origLookPath := lookPathFn
	origCwd, _ := os.Getwd()
	defer func() {
		syscallExecFn = origExec
		lookPathFn = origLookPath
		_ = os.Chdir(origCwd)
	}()

	lookPathFn = func(name string) (string, error) {
		return "", os.ErrNotExist
	}
	execCalled := false
	syscallExecFn = func(bin string, args []string, env []string) error {
		execCalled = true
		return nil
	}

	wt := t.TempDir()
	cfg := TeamLaunchConfig{
		Pattern:      PatternP3InProgress,
		WorktreePath: wt,
		LLM:          "cc",
	}
	err := launchP3(cfg)
	if err == nil {
		t.Errorf("launchP3 with missing moai binary should return error")
	}
	// Verify chdir was NOT called (cwd unchanged from origCwd).
	got, _ := os.Getwd()
	gotResolved, _ := absoluteResolved(got)
	origResolved, _ := absoluteResolved(origCwd)
	if gotResolved != origResolved {
		t.Errorf("os.Chdir must NOT be called when lookPathFn fails: cwd=%q, want=%q", gotResolved, origResolved)
	}
	if execCalled {
		t.Errorf("syscallExecFn must NOT be called when lookPathFn fails")
	}
}

// equalStringSlices is a small helper to keep tests free of reflect imports.
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// absoluteResolved returns the symlink-resolved absolute path. macOS /tmp →
// /private/tmp symlink would otherwise make cwd assertions flaky.
func absoluteResolved(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	return resolved, nil
}
