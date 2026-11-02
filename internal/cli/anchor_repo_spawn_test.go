package cli

// anchor_repo_spawn_test.go — the launcher-anchor fixture's git process count
// (Issue #1948, CLI package timeout).
//
// Defect answer: this catches the fixture going back to one git process per
// config key, per branch query and per commit-time maintenance run. Each
// launcher-anchor cell builds a fresh repository, so every process the fixture
// spawns is paid once per cell, and a process costs several times more on a
// loaded machine. It also catches the fixture losing a property the cells rely
// on: the seed commit on `main`, the committer identity, no signing.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// maxAnchorRepoGitProcs is the repository build: init, add, commit. The three
// worktrees of newLeaderFixture are counted separately.
const maxAnchorRepoGitProcs = 3

// gitProcessStarts counts the git processes a trace2 event file recorded.
func gitProcessStarts(t *testing.T, trace string) int {
	t.Helper()
	b, err := os.ReadFile(trace)
	if err != nil {
		t.Fatalf("read the git trace: %v", err)
	}
	return strings.Count(string(b), `"event":"start"`)
}

func TestAnchorRepo_BuildSpawnsFewGitProcessesAndKeepsItsProperties(t *testing.T) {
	trace := filepath.Join(t.TempDir(), "trace2.json")
	t.Setenv("GIT_TRACE2_EVENT", trace)

	// Positive control: the instrument sees a git process at all. Without it a
	// dead trace would read as zero spawns and pass the bound below.
	probe := t.TempDir()
	anchorGit(t, probe, "init", "-q")
	if got := gitProcessStarts(t, trace); got < 1 {
		t.Fatalf("control: the trace recorded %d git processes for a git init, want at least 1", got)
	}
	before := gitProcessStarts(t, trace)

	repo := newAnchorRepo(t)
	spawned := gitProcessStarts(t, trace) - before
	t.Logf("MEASURED newAnchorRepo spawned %d git processes (bound %d)", spawned, maxAnchorRepoGitProcs)
	if spawned > maxAnchorRepoGitProcs {
		t.Errorf("newAnchorRepo spawned %d git processes, want at most %d", spawned, maxAnchorRepoGitProcs)
	}

	// The properties the cells rely on, read with the trace off.
	t.Setenv("GIT_TRACE2_EVENT", "")
	if repo.base != "main" {
		t.Errorf("base = %q, want main", repo.base)
	}
	if got := anchorGit(t, repo.root, "branch", "--show-current"); got != "main" {
		t.Errorf("current branch = %q, want main", got)
	}
	if got := anchorGit(t, repo.root, "log", "-1", "--format=%s"); got != "seed" {
		t.Errorf("seed commit subject = %q, want seed", got)
	}
	if got := anchorGit(t, repo.root, "log", "-1", "--format=%ae %an"); got != "anchor-test@example.com Anchor Test" {
		t.Errorf("committer = %q, want the fixture identity", got)
	}
	if got := anchorGit(t, repo.root, "config", "--get", "commit.gpgsign"); got != "false" {
		t.Errorf("commit.gpgsign = %q, want false", got)
	}
	if _, err := os.Stat(filepath.Join(repo.root, ".moai")); err != nil {
		t.Errorf("the project marker .moai is missing: %v", err)
	}
	if got := anchorGit(t, repo.root, "status", "--porcelain"); got != "" {
		t.Errorf("the fixture is not clean after the seed commit: %q", got)
	}
}
