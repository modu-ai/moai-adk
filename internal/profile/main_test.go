package profile

// Package-wide test entry point (SPEC-PROFILE-MEMORY-001 REQ-PM-022).
//
// WHY THIS FILE EXISTS.
//
// This package owns GetBaseDir, which resolves to ~/.moai/claude-profiles
// whenever BaseDirOverride is empty. Every write path here — the launch ledger
// recorder above all — targets that directory. A test that forgets to set an
// override does not fail; it silently rewrites the developer's real
// launch.yaml, and the damage surfaces later as a `moai cc` that launches the
// wrong profile (or none). internal/cli has carried this guard since that
// exact incident; internal/profile, which is where the writes actually live,
// had none.
//
// The override here is a floor, not a substitute for per-test sandboxing:
// tests still set their own BaseDirOverride so they do not share state. What
// this adds is a package-wide guarantee that a test which forgets cannot reach
// $HOME.
//
// NOTE ON COVERAGE: this TestMain guard runs BEFORE every test in the package,
// so it cannot observe contamination introduced by a test that clears the
// override mid-run. Go runs test files in lexicographic order, so the guard in
// this file always executes first and always passes. That blind spot is
// covered separately by TestSandboxSurvivesPackageRun in
// zz_sandbox_guard_test.go, which sorts last.

import (
	"os"
	"path/filepath"
	"testing"
)

// sandboxProfileBaseDir points GetBaseDir at a throwaway directory for the whole
// package run. It mirrors the helper of the same name in internal/cli.
//
// t.TempDir() is unavailable in TestMain (no *testing.T), so the directory is
// created and removed manually.
func sandboxProfileBaseDir() func() {
	dir, err := os.MkdirTemp("", "moai-profile-pkg-")
	if err != nil {
		// Fall back to a path under the OS temp dir rather than silently
		// leaving the real base dir in play.
		dir = filepath.Join(os.TempDir(), "moai-profile-pkg-fallback")
		_ = os.MkdirAll(dir, 0o755)
	}
	orig := BaseDirOverride
	BaseDirOverride = dir
	return func() {
		BaseDirOverride = orig
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
// GetBaseDir() resolves to $HOME/.moai/claude-profiles and BaseDirOverride is "".
//
// The second assertion is the load-bearing one. An override pointed at the real
// base directory would satisfy a non-emptiness check while still writing $HOME,
// so the guard compares against the actual home-derived path rather than merely
// asserting that something was set.
func TestProfileBaseDirIsSandboxed(t *testing.T) {
	if BaseDirOverride == "" {
		t.Fatal("BaseDirOverride is empty: TestMain must call sandboxProfileBaseDir() " +
			"before m.Run(). Without it, any test reaching RecordLastUsedProfile " +
			"rewrites the developer's real ~/.moai/claude-profiles/launch.yaml.")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory; sandbox comparison unavailable")
	}
	realBase := filepath.Join(home, ".moai", "claude-profiles")
	if got := GetBaseDir(); got == realBase {
		t.Fatalf("GetBaseDir() = %q, which is the real user profile base. "+
			"Tests in this package must never resolve to it.", got)
	}
}
