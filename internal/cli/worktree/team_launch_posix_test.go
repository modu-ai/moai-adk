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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

// =============================================================================
// M3 — Pattern P1/P2 tmux window spawn tests
// =============================================================================
//
// These tests inject `tmuxNewWindowFn` to capture the cwd/command arguments
// that production code would otherwise pass to `tmux new-window`. The injection
// seam lets us assert argv correctness, pane_id format, and the fallback
// behavior on tmux failure without depending on an actual tmux server.

// TestLaunchP1_CG_TmuxGLMWindow verifies AC-WTL-001: when --team is set inside
// a tmux session with CG mode active, launchP1P2 must spawn a new tmux window
// with cwd=worktree-path and command `moai glm`. The pane_id returned by tmux
// must be propagated to the caller for swarm registry use (M4).
//
// REQ-WTL-001: P1 dispatch = tmux + CG → moai glm window.
func TestLaunchP1_CG_TmuxGLMWindow(t *testing.T) {
	origFn := tmuxNewWindowFn
	defer func() { tmuxNewWindowFn = origFn }()

	var capturedCwd, capturedCommand string
	tmuxNewWindowFn = func(cwd, command string) (string, error) {
		capturedCwd = cwd
		capturedCommand = command
		return "%5", nil
	}

	wt := t.TempDir()
	cfg := TeamLaunchConfig{
		Pattern:      PatternP1TmuxGLM,
		WorktreePath: wt,
		SpecID:       "SPEC-WTL-DEMO-001",
		Branch:       "feature/SPEC-WTL-DEMO-001",
		LLM:          "glm",
	}
	paneID, err := launchP1P2(cfg)
	if err != nil {
		t.Fatalf("launchP1P2 returned unexpected error: %v", err)
	}
	if paneID != "%5" {
		t.Errorf("paneID = %q, want %q", paneID, "%5")
	}
	if capturedCwd != wt {
		t.Errorf("captured cwd = %q, want %q", capturedCwd, wt)
	}
	if !strings.Contains(capturedCommand, "moai glm") {
		t.Errorf("captured command = %q, want substring %q", capturedCommand, "moai glm")
	}
	if strings.Contains(capturedCommand, "moai cc") {
		t.Errorf("captured command = %q must NOT contain %q", capturedCommand, "moai cc")
	}
}

// TestLaunchP2_NoCG_TmuxCCWindow verifies AC-WTL-002: when --team is set
// inside a tmux session but CG mode is NOT active, launchP1P2 must spawn a
// new tmux window with cwd=worktree-path and command `moai cc`.
//
// REQ-WTL-002: P2 dispatch = tmux + CC → moai cc window.
func TestLaunchP2_NoCG_TmuxCCWindow(t *testing.T) {
	origFn := tmuxNewWindowFn
	defer func() { tmuxNewWindowFn = origFn }()

	var capturedCwd, capturedCommand string
	tmuxNewWindowFn = func(cwd, command string) (string, error) {
		capturedCwd = cwd
		capturedCommand = command
		return "%7", nil
	}

	wt := t.TempDir()
	cfg := TeamLaunchConfig{
		Pattern:      PatternP2TmuxCC,
		WorktreePath: wt,
		SpecID:       "SPEC-WTL-DEMO-002",
		Branch:       "feature/SPEC-WTL-DEMO-002",
		LLM:          "cc",
	}
	paneID, err := launchP1P2(cfg)
	if err != nil {
		t.Fatalf("launchP1P2 returned unexpected error: %v", err)
	}
	if paneID != "%7" {
		t.Errorf("paneID = %q, want %q", paneID, "%7")
	}
	if capturedCwd != wt {
		t.Errorf("captured cwd = %q, want %q", capturedCwd, wt)
	}
	if !strings.Contains(capturedCommand, "moai cc") {
		t.Errorf("captured command = %q, want substring %q", capturedCommand, "moai cc")
	}
	if strings.Contains(capturedCommand, "moai glm") {
		t.Errorf("captured command = %q must NOT contain %q", capturedCommand, "moai glm")
	}
}

// TestLaunchP1P2_PaneSpawnFailure_FallbackHandoff verifies AC-WTL-007: when
// tmux new-window returns a non-zero exit, launchP1P2 returns a wrapped error
// containing "tmux" so the caller (dispatchTeamLaunch in new.go) can fall
// back to the P4 handoff path with a stderr notice.
//
// REQ-WTL-007: Pane spawn failure → P4 fallback.
func TestLaunchP1P2_PaneSpawnFailure_FallbackHandoff(t *testing.T) {
	origFn := tmuxNewWindowFn
	defer func() { tmuxNewWindowFn = origFn }()

	tmuxNewWindowFn = func(cwd, command string) (string, error) {
		return "", errors.New("tmux: no server running on /tmp/tmux-1000/default")
	}

	cfg := TeamLaunchConfig{
		Pattern:      PatternP1TmuxGLM,
		WorktreePath: t.TempDir(),
		SpecID:       "SPEC-WTL-DEMO-003",
		LLM:          "glm",
	}
	paneID, err := launchP1P2(cfg)
	if err == nil {
		t.Fatalf("launchP1P2 should return error when tmux fails; got paneID=%q", paneID)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "tmux") {
		t.Errorf("error message = %q, want substring %q", err.Error(), "tmux")
	}
	if paneID != "" {
		t.Errorf("paneID on failure = %q, want empty", paneID)
	}
}

// TestLaunchP1P2_PaneIDCaptured verifies that the pane_id returned by tmux
// matches the canonical `%N` format (tmux convention) and is propagated to
// the caller. This pane_id is the swarm registry's "success signal" for
// P1/P2 entries (M4 REQ-WTL-008).
func TestLaunchP1P2_PaneIDCaptured(t *testing.T) {
	origFn := tmuxNewWindowFn
	defer func() { tmuxNewWindowFn = origFn }()

	tmuxNewWindowFn = func(cwd, command string) (string, error) {
		return "%17", nil
	}

	cfg := TeamLaunchConfig{
		Pattern:      PatternP1TmuxGLM,
		WorktreePath: t.TempDir(),
		SpecID:       "SPEC-WTL-DEMO-004",
		LLM:          "glm",
	}
	paneID, err := launchP1P2(cfg)
	if err != nil {
		t.Fatalf("launchP1P2 returned unexpected error: %v", err)
	}
	paneIDPattern := regexp.MustCompile(`^%[0-9]+$`)
	if !paneIDPattern.MatchString(paneID) {
		t.Errorf("paneID = %q does not match ^%%[0-9]+$ (tmux pane_id format)", paneID)
	}
}

// TestLaunchP1P2_SettingsLocalJSON_ByteIdentical verifies AC-WTL-001 and
// AC-WTL-002 sub-assertion: launchP1P2 must NEVER mutate settings.local.json
// (R7 HARD constraint — settings.local.json is runtime-managed by `moai cg`
// and SessionStart hook, not by team launch).
//
// The test creates a settings.local.json, hashes it before launch, runs
// launchP1P2 with a tmux fake, and asserts the file's SHA-256 is unchanged.
func TestLaunchP1P2_SettingsLocalJSON_ByteIdentical(t *testing.T) {
	origFn := tmuxNewWindowFn
	defer func() { tmuxNewWindowFn = origFn }()

	tmuxNewWindowFn = func(cwd, command string) (string, error) {
		return "%9", nil
	}

	// Prepare a settings.local.json with realistic content (teammateMode +
	// env vars). The HARD constraint is: launchP1P2 must not write to this
	// file at all.
	wt := t.TempDir()
	settingsDir := filepath.Join(wt, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatalf("mkdir settings dir: %v", err)
	}
	settingsPath := filepath.Join(settingsDir, "settings.local.json")
	originalContent := []byte(`{
  "teammateMode": "tmux",
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "sk-fake-token-do-not-mutate",
    "ANTHROPIC_BASE_URL": "https://api.fake.example.com"
  }
}
`)
	if err := os.WriteFile(settingsPath, originalContent, 0o644); err != nil {
		t.Fatalf("write settings: %v", err)
	}

	hashFile := func(path string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read settings: %v", err)
		}
		sum := sha256.Sum256(data)
		return hex.EncodeToString(sum[:])
	}

	before := hashFile(settingsPath)

	cfg := TeamLaunchConfig{
		Pattern:      PatternP1TmuxGLM,
		WorktreePath: wt,
		SpecID:       "SPEC-WTL-DEMO-005",
		LLM:          "glm",
	}
	if _, err := launchP1P2(cfg); err != nil {
		t.Fatalf("launchP1P2 returned unexpected error: %v", err)
	}

	after := hashFile(settingsPath)
	if before != after {
		t.Errorf("settings.local.json mutated by launchP1P2: before=%q, after=%q", before, after)
	}
}

// TestLaunchP1P2_GLMDriftFallback_P2 verifies REQ-WTL-009 drift case
// propagation into M3: when CG mode is configured (teammateMode=tmux) but
// the GLM env vars are missing, IsCGMode returns false (with stderr warning).
// dispatchTeamLaunch then selects PatternP2TmuxCC and constructs a
// TeamLaunchConfig with LLM="cc" — NOT "glm". This test verifies that, given
// such a config (P2 + LLM=cc), launchP1P2 dispatches `moai cc` even though
// the original user intent (per settings.local.json) was GLM.
//
// REQ-WTL-009: teammateMode=tmux + no GLM env → P2 fallback (cc, not glm).
func TestLaunchP1P2_GLMDriftFallback_P2(t *testing.T) {
	origFn := tmuxNewWindowFn
	defer func() { tmuxNewWindowFn = origFn }()

	var capturedCommand string
	tmuxNewWindowFn = func(cwd, command string) (string, error) {
		capturedCommand = command
		return "%11", nil
	}

	// Simulate the post-dispatchTeamLaunch state: P2 was chosen because
	// IsCGMode returned false (drift case), LLM was set to "cc" not "glm".
	cfg := TeamLaunchConfig{
		Pattern:      PatternP2TmuxCC,
		WorktreePath: t.TempDir(),
		SpecID:       "SPEC-WTL-DEMO-006",
		Branch:       "feature/SPEC-WTL-DEMO-006",
		LLM:          "cc", // drift fallback already applied upstream
	}
	if _, err := launchP1P2(cfg); err != nil {
		t.Fatalf("launchP1P2 returned unexpected error: %v", err)
	}
	if !strings.Contains(capturedCommand, "moai cc") {
		t.Errorf("captured command = %q must contain %q (drift fallback)", capturedCommand, "moai cc")
	}
	if strings.Contains(capturedCommand, "moai glm") {
		t.Errorf("captured command = %q must NOT contain %q (drift means no GLM)", capturedCommand, "moai glm")
	}
}

// TestLaunchP1P2_EmptyLLMDefaultsToCC verifies the defensive `llm == ""`
// guard: if a TeamLaunchConfig is constructed without setting LLM (e.g., by
// a future caller that bypasses dispatchTeamLaunch), launchP1P2 defaults to
// `moai cc` rather than producing an empty command string like `moai`.
//
// This is a defense-in-depth guarantee — dispatchTeamLaunch always populates
// cfg.LLM today, so this branch is unreachable in normal flow.
func TestLaunchP1P2_EmptyLLMDefaultsToCC(t *testing.T) {
	origFn := tmuxNewWindowFn
	defer func() { tmuxNewWindowFn = origFn }()

	var capturedCommand string
	tmuxNewWindowFn = func(cwd, command string) (string, error) {
		capturedCommand = command
		return "%99", nil
	}

	cfg := TeamLaunchConfig{
		Pattern:      PatternP2TmuxCC,
		WorktreePath: t.TempDir(),
		SpecID:       "SPEC-WTL-DEMO-007",
		// LLM intentionally omitted — defensive default should kick in.
	}
	if _, err := launchP1P2(cfg); err != nil {
		t.Fatalf("launchP1P2 returned unexpected error: %v", err)
	}
	if !strings.Contains(capturedCommand, "moai cc") {
		t.Errorf("captured command = %q must default to %q when cfg.LLM is empty", capturedCommand, "moai cc")
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
