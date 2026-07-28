// launcher_worktree_l2_test.go: tests for the launcher-side L2 absolute-path
// pre-resolution step (SPEC-WORKTREE-ENTRY-STRATEGY-001 M3a, REQ-WES-010).
//
// AC-WES-010a: an absolute path under ~/.moai/worktrees/<project>/... resolves
//   to the L2 worktree (accepted; passes through unchanged so claude uses the
//   absolute path directly rather than treating it as a .claude/worktrees/
//   short name).
// AC-WES-010b: legacy short-name token-normalization behavior preserved
//   (covered by the existing TestNormalizeWorktreeFlag; this file asserts the
//   pre-resolution step does not interfere with short-name inputs).
// AC-WES-010c: an absolute path NOT under ~/.moai/worktrees/ or
//   .claude/worktrees/ is rejected with a clear error.
//
// NOTE: does not call t.Parallel() because it sets HOME via t.Setenv (process-
// global state); follows the TestCleanupMoaiWorktrees_GlobalPath convention.
package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

// runResolveWorktreeL2Path sets HOME to homeDir via t.Setenv (deterministic,
// parallel-safe because this test file forgoes t.Parallel) and invokes the
// production resolveWorktreeL2Path. The project root is resolved from the
// real findProjectRoot() — this test file lives in the repo, so
// <repo>/.claude/worktrees/ is the L1 prefix used for the L1-vs-L2 check.
func runResolveWorktreeL2Path(t *testing.T, homeDir string, args []string) error {
	t.Helper()
	t.Setenv("HOME", homeDir)
	return resolveWorktreeL2Path(args)
}

// TestLauncherWorktreeL2AbsPath covers AC-WES-010a: absolute paths under
// ~/.moai/worktrees/ are accepted (the pre-resolution step returns nil and
// leaves args unchanged so normalizeWorktreeFlag can do its token rewriting
// and claude then receives the absolute path via the canonical --worktree
// <abs-path> two-token form).
func TestLauncherWorktreeL2AbsPath(t *testing.T) {
	tmpHome := t.TempDir()
	l2Base := filepath.Join(tmpHome, ".moai", "worktrees")
	// Synthesize a representative L2 absolute path matching the auto-isolation
	// naming scheme documented in REQ-WES-005 (auto-<session-short>-<spec-id>).
	l2Path := filepath.Join(l2Base, "moai-adk-go", "auto-6558cd02-WES-001")

	tests := []struct {
		name string
		args []string
	}{
		{"-w <abs-path>", []string{"-w", l2Path}},
		{"--worktree <abs-path>", []string{"--worktree", l2Path}},
		{"--worktree=<abs-path>", []string{"--worktree=" + l2Path}},
		{"-w=<abs-path>", []string{"-w=" + l2Path}},
		{"-w <abs-path> with other flags", []string{"-b", "-w", l2Path, "--model", "opus"}},
		{"-w <abs-path> before pass-through marker", []string{"-w", l2Path, "--", "--print"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runResolveWorktreeL2Path(t, tmpHome, tt.args)
			if err != nil {
				t.Fatalf("resolveWorktreeL2Path(%v) returned error for L2 absolute path: %v", tt.args, err)
			}
		})
	}
}

// TestLauncherWorktreeShortNamePreserved covers AC-WES-010b: short-name
// values (non-absolute paths) MUST pass through the pre-resolution step
// without error so normalizeWorktreeFlag can handle the token rewriting and
// claude resolves the short name against .claude/worktrees/<name>.
func TestLauncherWorktreeShortNamePreserved(t *testing.T) {
	tmpHome := t.TempDir()

	tests := []struct {
		name string
		args []string
	}{
		{"short name feat-login", []string{"-w", "feat-login"}},
		{"bare -w (auto-name)", []string{"-w"}},
		{"--worktree= empty (auto-name)", []string{"--worktree="}},
		{"no -w flag at all", []string{"-b", "--model", "opus"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runResolveWorktreeL2Path(t, tmpHome, tt.args)
			if err != nil {
				t.Fatalf("resolveWorktreeL2Path(%v) returned error for short-name input: %v", tt.args, err)
			}
		})
	}
}

// TestLauncherWorktreeReject covers AC-WES-010c: absolute paths NOT under
// ~/.moai/worktrees/ or <project>/.claude/worktrees/ MUST be rejected with a
// clear error (no silent fall-through to creating a new worktree).
func TestLauncherWorktreeReject(t *testing.T) {
	tmpHome := t.TempDir()

	// An absolute path that is neither under tmpHome/.moai/worktrees/ (the L2
	// prefix when HOME=tmpHome) nor under the real repo's .claude/worktrees/
	// (it points into the test temp dir tree, which is disjoint from the repo).
	outOfPrefix := filepath.Join(tmpHome, "random-dir", "not-a-worktree")

	tests := []struct {
		name string
		args []string
	}{
		{"-w <out-of-prefix abs-path>", []string{"-w", outOfPrefix}},
		{"--worktree <out-of-prefix abs-path>", []string{"--worktree", outOfPrefix}},
		{"--worktree=<out-of-prefix abs-path>", []string{"--worktree=" + outOfPrefix}},
		{"-w=<out-of-prefix abs-path>", []string{"-w=" + outOfPrefix}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runResolveWorktreeL2Path(t, tmpHome, tt.args)
			if err == nil {
				t.Fatalf("resolveWorktreeL2Path(%v) returned nil for out-of-prefix absolute path; want non-nil error", tt.args)
			}
			// The error message MUST name the accepted prefixes so the user
			// knows how to fix the invocation (no silent fall-through).
			msg := err.Error()
			for _, want := range []string{".moai/worktrees", ".claude/worktrees"} {
				if !strings.Contains(msg, want) {
					t.Errorf("error message %q does not name accepted prefix %q", msg, want)
				}
			}
		})
	}
}

// TestLauncherWorktreeL2AfterPassThroughMarker covers Edge: a -w value
// appearing AFTER the "--" pass-through marker is verbatim pass-through and
// MUST NOT be validated by the pre-resolution step (claude owns those tokens).
func TestLauncherWorktreeL2AfterPassThroughMarker(t *testing.T) {
	tmpHome := t.TempDir()
	outOfPrefix := filepath.Join(tmpHome, "random-dir", "not-a-worktree")

	// After "--", the -w token is verbatim pass-through; no validation.
	args := []string{"--", "-w", outOfPrefix}
	err := runResolveWorktreeL2Path(t, tmpHome, args)
	if err != nil {
		t.Fatalf("resolveWorktreeL2Path(%v) must not validate tokens after --: got %v", args, err)
	}
}
