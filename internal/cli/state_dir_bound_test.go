package cli

// SPEC-CLI-STATE-DIR-BOUND-001 — regression tests for the state-directory
// resolution contract: CLAUDE_PROJECT_DIR first for read/append consumers,
// otherwise the guarded project-root convention that never crosses $HOME.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// normPath normalizes an EXPECTED path so it can be compared against a raw
// actual value. Per acceptance.md §D.2 the actual value is never normalized:
// normalizing both sides passes under an implementation that does not
// normalize at all, which is precisely the contract under test (REQ-8).
func normPath(t *testing.T, p string) string {
	t.Helper()
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return p
}

// TestStateDirDoesNotCrossHomeBoundary is the regression for the observed
// asymmetry: a caller under $HOME silently resolved to ~/.moai/state while the
// same call under /tmp failed. The home directory is relocated into a
// fully-owned temp tree so the decoy is the test's own, not the machine's.
//
// t.Setenv makes this test non-parallel by construction, which is the
// isolation this needs — HOME is process-global state.
func TestStateDirDoesNotCrossHomeBoundary(t *testing.T) {
	root := t.TempDir()
	home := mustMkdirAll(t, filepath.Join(root, "home"))
	decoy := mustMkdirAll(t, filepath.Join(home, ".moai", "state"))
	start := mustMkdirAll(t, filepath.Join(home, "probe", "sub"))

	t.Setenv("HOME", home)
	t.Setenv(config.EnvClaudeProjectDir, "")

	got, err := findStateDirFrom(start)
	if err == nil {
		t.Errorf("resolution succeeded at %q; it must stop at the home boundary", got)
		if got == normPath(t, decoy) {
			t.Errorf("claimed the home-band decoy %q", decoy)
		}
	}
}

// TestStateDirStopsAtProjectRoot pins REQ-3 class A: a project root holding a
// bare .moai (no state/) fails there instead of climbing to an ancestor that
// happens to have one.
func TestStateDirStopsAtProjectRoot(t *testing.T) {
	outer := t.TempDir()
	mustMkdirAll(t, filepath.Join(outer, ".moai", "state"))
	root := mustMkdirAll(t, filepath.Join(outer, "project"))
	mustMkdirAll(t, filepath.Join(root, ".moai"))
	start := mustMkdirAll(t, filepath.Join(root, "sub"))

	t.Setenv(config.EnvClaudeProjectDir, "")

	got, err := findStateDirFrom(start)
	if err == nil {
		t.Fatalf("got %q; resolution must fail at the project root that has no state dir", got)
	}
	if !strings.Contains(err.Error(), normPath(t, root)) {
		t.Errorf("error %q does not name the directory it stopped at (%q)", err, normPath(t, root))
	}
}

// TestStateDirStopsAtNestedBareMoai pins REQ-3 class B — a deliberate
// regression. A bare .moai in a subdirectory of a valid project root used to
// be skipped; it now stops the resolution there. The correct answer is
// reachable and still refused, which is a decision rather than an accident,
// so it is pinned.
func TestStateDirStopsAtNestedBareMoai(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, ".moai", "state"))
	nested := mustMkdirAll(t, filepath.Join(root, "nested"))
	mustMkdirAll(t, filepath.Join(nested, ".moai"))
	start := mustMkdirAll(t, filepath.Join(nested, "sub"))

	t.Setenv(config.EnvClaudeProjectDir, "")

	got, err := findStateDirFrom(start)
	if err == nil {
		t.Fatalf("got %q; the bare .moai in the subdirectory must stop the resolution", got)
	}
	if !strings.Contains(err.Error(), normPath(t, nested)) {
		t.Errorf("error %q does not name the directory it stopped at (%q)", err, normPath(t, nested))
	}
}

// TestStateDirHonoursProjectDirEnv pins REQ-1: an explicitly named project
// wins with no walk, and the named value is normalized before use (REQ-8).
func TestStateDirHonoursProjectDirEnv(t *testing.T) {
	root := t.TempDir()
	decoy := mustMkdirAll(t, filepath.Join(root, ".moai", "state"))
	explicit := mustMkdirAll(t, filepath.Join(root, "explicit"))
	want := mustMkdirAll(t, filepath.Join(explicit, ".moai", "state"))
	start := mustMkdirAll(t, filepath.Join(root, "sub"))

	t.Chdir(start)
	// The raw, un-normalized form t.TempDir() handed out: on darwin this is
	// /var/... while the resolved form is /private/var/....
	t.Setenv(config.EnvClaudeProjectDir, explicit)

	got, err := findStateDir()
	if err != nil {
		t.Fatalf("findStateDir: %v", err)
	}
	if got != normPath(t, want) {
		t.Errorf("got %q, want the named project's state dir %q", got, normPath(t, want))
	}
	if got == normPath(t, decoy) {
		t.Errorf("resolution walked up to %q instead of honouring the named project", decoy)
	}
}

// TestStateDirReturnsNormalizedPath pins REQ-8 on both branches. The assertion
// holds on every platform (a normalized path re-normalizes to itself); on
// darwin, where /var is a symlink to private/var, branch B additionally
// distinguishes an implementation that passes CLAUDE_PROJECT_DIR through raw.
func TestStateDirReturnsNormalizedPath(t *testing.T) {
	t.Run("walk branch", func(t *testing.T) {
		root := t.TempDir()
		mustMkdirAll(t, filepath.Join(root, ".moai", "state"))
		start := mustMkdirAll(t, filepath.Join(root, "sub"))

		t.Setenv(config.EnvClaudeProjectDir, "")

		got, err := findStateDirFrom(start)
		if err != nil {
			t.Fatalf("findStateDirFrom: %v", err)
		}
		resolved, err := filepath.EvalSymlinks(got)
		if err != nil {
			t.Fatalf("EvalSymlinks(%q): %v", got, err)
		}
		if got != resolved {
			t.Errorf("got %q, which is not its own normalized form %q", got, resolved)
		}
	})

	t.Run("env branch", func(t *testing.T) {
		root := t.TempDir()
		explicit := mustMkdirAll(t, filepath.Join(root, "explicit"))
		mustMkdirAll(t, filepath.Join(explicit, ".moai", "state"))
		start := mustMkdirAll(t, filepath.Join(root, "sub"))

		t.Chdir(start)
		t.Setenv(config.EnvClaudeProjectDir, explicit)

		got, err := findStateDir()
		if err != nil {
			t.Fatalf("findStateDir: %v", err)
		}
		resolved, err := filepath.EvalSymlinks(got)
		if err != nil {
			t.Fatalf("EvalSymlinks(%q): %v", got, err)
		}
		if got != resolved {
			t.Errorf("got %q, which is not its own normalized form %q", got, resolved)
		}
		if normPath(t, explicit) != explicit && strings.HasPrefix(got, explicit) {
			t.Errorf("got %q still carries the raw CLAUDE_PROJECT_DIR prefix %q", got, explicit)
		}
	})
}

// TestResolveTokensStateDirFallsBackToCwd pins REQ-4: without this fallback
// `moai tokens record` breaks on a fresh checkout that has no .moai yet.
func TestResolveTokensStateDirFallsBackToCwd(t *testing.T) {
	cwd := t.TempDir()
	t.Chdir(cwd)
	t.Setenv(config.EnvClaudeProjectDir, "")

	got, err := resolveTokensStateDir()
	if err != nil {
		t.Fatalf("resolveTokensStateDir: %v", err)
	}
	want := normPath(t, filepath.Join(cwd, ".moai", "state"))
	if got != want {
		t.Errorf("got %q, want the cwd fallback %q", got, want)
	}
}

// TestLoadRegistryForOverlayFailsOpen pins the fail-open constraint. REQ-3
// deliberately increases the number of directories where resolution fails, so
// this path sees more traffic than it used to.
func TestLoadRegistryForOverlayFailsOpen(t *testing.T) {
	root := t.TempDir()
	start := mustMkdirAll(t, filepath.Join(root, "sub"))
	t.Chdir(start)
	t.Setenv("HOME", root)
	t.Setenv(config.EnvClaudeProjectDir, "")

	if got := loadRegistryForOverlay(); got != nil {
		t.Errorf("got %v, want nil when the state dir cannot be resolved", got)
	}
}
