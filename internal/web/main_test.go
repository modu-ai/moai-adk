package web

// Package-wide test entry point (SPEC-PROFILE-MEMORY-001 REQ-PM-022).
//
// WHY THIS FILE EXISTS.
//
// newApp wires recordLastProfile to the real profile ledger writer. Most tests
// build their app through newTestApp, which stubs that seam — but many
// construct newApp directly (board, agentfm, schema, project-config, profile
// CRUD, traversal), and those carry the real writer. With no override,
// profile.GetBaseDir() resolves to ~/.moai/claude-profiles, so a save handler
// exercised in a test rewrites the developer's real launch.yaml: the ledger a
// bare `moai cc` reads to decide which profile to launch.
//
// The failure is silent. Nothing in the test output indicates $HOME was
// touched, and the consequence shows up later as a launch using the wrong
// profile. internal/cli has carried this guard since that incident; this
// package needed it once newApp began writing project-scoped ledger entries.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// sandboxProfileBaseDir points profile.GetBaseDir at a throwaway directory for
// the whole package run. It mirrors the helper of the same name in
// internal/cli and internal/profile.
//
// t.TempDir() is unavailable in TestMain (no *testing.T), so the directory is
// created and removed manually.
func sandboxProfileBaseDir() func() {
	dir, err := os.MkdirTemp("", "moai-web-profiles-")
	if err != nil {
		// Fall back to a path under the OS temp dir rather than silently
		// leaving the real base dir in play.
		dir = filepath.Join(os.TempDir(), "moai-web-profiles-fallback")
		_ = os.MkdirAll(dir, 0o755)
	}
	orig := profile.BaseDirOverride
	profile.BaseDirOverride = dir
	return func() {
		profile.BaseDirOverride = orig
		_ = os.RemoveAll(dir)
	}
}

func TestMain(m *testing.M) {
	restore := sandboxProfileBaseDir()
	code := m.Run()
	restore()
	os.Exit(code)
}

// TestProfileBaseDirIsSandboxed is the guard for sandboxProfileBaseDir. It fails
// deterministically if the TestMain call is removed: with no override,
// profile.GetBaseDir() resolves to $HOME/.moai/claude-profiles and
// BaseDirOverride is "".
//
// The second assertion is the load-bearing one. An override pointed at the real
// base directory would satisfy a non-emptiness check while still writing $HOME,
// so the guard compares against the actual home-derived path rather than merely
// asserting that something was set.
func TestProfileBaseDirIsSandboxed(t *testing.T) {
	if profile.BaseDirOverride == "" {
		t.Fatal("profile.BaseDirOverride is empty: TestMain must call " +
			"sandboxProfileBaseDir() before m.Run(). Without it, any test " +
			"constructing newApp directly carries the real ledger writer and " +
			"can rewrite the developer's ~/.moai/claude-profiles/launch.yaml.")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory; sandbox comparison unavailable")
	}
	realBase := filepath.Join(home, ".moai", "claude-profiles")
	if got := profile.GetBaseDir(); got == realBase {
		t.Fatalf("profile.GetBaseDir() = %q, which is the real user profile "+
			"base. Tests in this package must never resolve to it.", got)
	}
}
