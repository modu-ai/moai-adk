package homestate_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestProjectKeyArgumentEquivalence pins the property ProjectDir relies on
// when it passes the already-canonical root on one branch and the raw
// projectRoot on the other: ProjectKey canonicalizes its own argument, so both
// spellings reach the same key. The property holds only while
// CanonicalProjectRoot is idempotent, and idempotence is where a path-resolution
// change would break it silently — the two ProjectDir branches would start
// keying the same project into two directories, with no error on either path.
func TestProjectKeyArgumentEquivalence(t *testing.T) {
	for _, tc := range projectRootShapes(t) {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}
			raw := homestate.ProjectKey(tc.root)
			canonical := homestate.ProjectKey(homestate.CanonicalProjectRoot(tc.root))
			if raw != canonical {
				t.Fatalf("ProjectKey(%q) = %q, but ProjectKey(CanonicalProjectRoot(%q)) = %q — the two ProjectDir branches would key this project into two directories",
					tc.root, raw, tc.root, canonical)
			}
			t.Logf("root=%q key=%q", tc.root, raw)
		})
	}

	// Positive control. Without it every assertion above is satisfied by an
	// implementation that returns a constant, and the guard reads as
	// protection while protecting nothing.
	t.Run("positive control: distinct roots yield distinct keys", func(t *testing.T) {
		a := filepath.Join(t.TempDir(), "alpha")
		b := filepath.Join(t.TempDir(), "beta")
		for _, dir := range []string{a, b} {
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatalf("mkdir %s: %v", dir, err)
			}
		}
		if ka, kb := homestate.ProjectKey(a), homestate.ProjectKey(b); ka == kb {
			t.Fatalf("distinct roots produced the same key: %s and %s both -> %q", a, b, ka)
		}
	})
}

type rootShape struct {
	name string
	root string
	skip string
}

// projectRootShapes builds the six input shapes ProjectDir can be called with
// in practice. Fixtures live under t.TempDir(); the repository-derived shapes
// are read through CanonicalProjectRoot rather than hardcoded, so the test
// carries no machine-specific path. A shape that cannot be constructed here
// is reported as a skip with its reason, never dropped.
func projectRootShapes(t *testing.T) []rootShape {
	t.Helper()

	primary := homestate.CanonicalProjectRoot(".")
	shapes := []rootShape{
		{name: "primary checkout", root: primary},
		{name: "relative dot", root: "."},
		{name: "plain temp dir", root: t.TempDir()},
	}

	repo, worktree, gitSkip := gitFixtures(t)
	shapes = append(shapes,
		rootShape{name: "git-initialised temp dir", root: repo, skip: gitSkip},
		rootShape{name: "linked worktree", root: worktree, skip: gitSkip},
	)

	link := filepath.Join(t.TempDir(), "primary-link")
	linkSkip := ""
	if err := os.Symlink(primary, link); err != nil {
		linkSkip = "symlink unsupported in this environment: " + err.Error()
	}
	shapes = append(shapes, rootShape{name: "symlink to primary checkout", root: link, skip: linkSkip})

	return shapes
}

// gitFixtures creates a standalone repository and one linked worktree under
// t.TempDir(). It returns a skip reason instead of failing when git is absent
// or refuses, so the remaining shapes still run.
func gitFixtures(t *testing.T) (repo, worktree, skip string) {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		return "", "", "git not on PATH: " + err.Error()
	}

	base := t.TempDir()
	repo = filepath.Join(base, "repo")
	worktree = filepath.Join(base, "wt")
	if err := os.MkdirAll(repo, 0o700); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	steps := [][]string{
		{"init", "--initial-branch=main"},
		{"config", "user.email", "test@example.invalid"},
		{"config", "user.name", "test"},
		{"commit", "--allow-empty", "-m", "seed"},
		{"worktree", "add", worktree, "-b", "linked"},
	}
	for _, args := range steps {
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", "", "git " + args[0] + " failed: " + err.Error() + ": " + string(out)
		}
	}
	return repo, worktree, ""
}
