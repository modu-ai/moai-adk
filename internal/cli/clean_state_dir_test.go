package cli

// SPEC-CLI-STATE-DIR-BOUND-001 REQ-9 + REQ-10 — `moai clean` decides what to
// delete from the working directory alone, and says which project that is
// before it enumerates anything.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/config"
)

// cleanFixture builds two sibling projects under a shared home boundary: A
// holds a run old enough to be deleted, B is the project an inherited
// CLAUDE_PROJECT_DIR names. The working directory is inside A.
func cleanFixture(t *testing.T) (projectA, projectB string) {
	t.Helper()
	home := t.TempDir()
	projectA = mustMkdirAll(t, filepath.Join(home, "A"))
	projectB = mustMkdirAll(t, filepath.Join(home, "B"))

	for _, root := range []string{projectA, projectB} {
		mustMkdirAll(t, filepath.Join(root, ".moai", "state", "runs"))
		sections := mustMkdirAll(t, filepath.Join(root, ".moai", "config", "sections"))
		if err := os.WriteFile(filepath.Join(sections, "state.yaml"), []byte("state:\n  retention_days: 1\n"), 0o644); err != nil {
			t.Fatalf("write state.yaml: %v", err)
		}
		old := mustMkdirAll(t, filepath.Join(root, ".moai", "state", "runs", "2020-01-01"))
		stale := time.Now().AddDate(0, 0, -30)
		if err := os.Chtimes(old, stale, stale); err != nil {
			t.Fatalf("chtimes %s: %v", old, err)
		}
	}

	t.Setenv("HOME", home)
	t.Chdir(mustMkdirAll(t, filepath.Join(projectA, "sub")))
	t.Setenv(config.EnvClaudeProjectDir, projectB)
	return projectA, projectB
}

// TestCleanIgnoresProjectDirEnv is the door REQ-9 closes. Handing every
// consumer the environment variable would have handed it to os.RemoveAll too,
// and this repository already has a record of CLAUDE_PROJECT_DIR failing to
// track a worktree-resident agent's actual directory — the deletion would land
// in the primary checkout.
func TestCleanIgnoresProjectDirEnv(t *testing.T) {
	projectA, projectB := cleanFixture(t)

	got, err := findStateDirNoEnv()
	if err != nil {
		t.Fatalf("findStateDirNoEnv: %v", err)
	}
	wantA := filepath.Join(normPath(t, projectA), ".moai", "state")
	notB := filepath.Join(normPath(t, projectB), ".moai", "state")
	if got != wantA {
		t.Errorf("got %q, want the working directory's project %q", got, wantA)
	}
	if got == notB {
		t.Errorf("resolution followed CLAUDE_PROJECT_DIR to %q", notB)
	}

	var stdout, stderr bytes.Buffer
	p := printer.New(printer.WithWriters(&stdout, &stderr), printer.WithMode(printer.ModePlain))
	if err := runClean(p, true); err != nil {
		t.Fatalf("runClean: %v", err)
	}

	if _, err := os.Stat(filepath.Join(projectB, ".moai", "state", "runs", "2020-01-01")); err != nil {
		t.Errorf("a run under the named-but-not-current project was removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectA, ".moai", "state", "runs", "2020-01-01")); !os.IsNotExist(err) {
		t.Errorf("the working directory's own old run survived; stat err = %v", err)
	}
}

// TestCleanAnnouncesResolvedRoot pins REQ-10 where it matters most. Read
// commands honour CLAUDE_PROJECT_DIR and clean does not, so within one session
// `moai state dump` can inspect one project while `moai clean` deletes from
// another. That divergence is accepted; this line is the only thing that makes
// it visible, and it is worth nothing after the file list has been printed.
func TestCleanAnnouncesResolvedRoot(t *testing.T) {
	projectA, projectB := cleanFixture(t)

	var stdout, stderr bytes.Buffer
	p := printer.New(printer.WithWriters(&stdout, &stderr), printer.WithMode(printer.ModePlain))
	if err := runClean(p, false); err != nil {
		t.Fatalf("runClean: %v", err)
	}

	out := stderr.String()
	announced := strings.Index(out, "resolved project root: "+normPath(t, projectA))
	if announced < 0 {
		t.Fatalf("no line naming the resolved root %q in %q", normPath(t, projectA), out)
	}
	if strings.Contains(out, normPath(t, projectB)) {
		t.Errorf("output names the project it did not resolve (%q): %q", normPath(t, projectB), out)
	}

	candidate := strings.Index(out, "Would delete:")
	if candidate < 0 {
		t.Fatalf("no deletion candidate was enumerated, so the ordering cannot be observed: %q", out)
	}
	if announced > candidate {
		t.Errorf("the resolved root was announced after the deletion candidates:\n%s", out)
	}
}
