package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/graph"
)

// SPEC-GRAPH-REPORT-001 REQ-GR-010/011 — the SINGLE rebuild path exercised
// directly. The SessionStart deferred edges refresh was removed (card t483,
// t448 Option-1 consumer-pays migration): consumers pay at query time, and
// the rebuild itself is owned by refreshEdgesArtifact. These tests pin the
// rebuild contract the preserved consumer path depends on: a stale tree is
// brought back in sync (staleness predicate flips false) and nothing lands
// in the git index.

// initFixtureGitRepo turns the fixture root into a git repository so the
// staged-entry assertion can read `git status --porcelain`.
func initFixtureGitRepo(t *testing.T, root string) {
	t.Helper()
	cmd := exec.Command("git", "init", "-q", root)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v (%s)", err, out)
	}
}

// stagedPorcelainLines returns the porcelain lines representing STAGED entries
// (index status in column 1, excluding untracked '??').
func stagedPorcelainLines(t *testing.T, root string) []string {
	t.Helper()
	out, err := exec.Command("git", "-C", root, "status", "--porcelain").CombinedOutput()
	if err != nil {
		t.Fatalf("git status --porcelain: %v (%s)", err, out)
	}
	var staged []string
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) >= 2 && line[0] != ' ' && line[0] != '?' {
			staged = append(staged, line)
		}
	}
	return staged
}

// AC-GR-015 stale leg: the rebuild path over a stale tree rebuilds the
// default artifact so both staleness predicates read false afterward, and
// nothing lands in the git index.
func TestRefreshEdgesArtifact_StaleRefreshesAndStagesNothing(t *testing.T) {
	root := graphFixtureProject(t)
	initFixtureGitRepo(t, root)
	edgesFile := filepath.Join(root, ".moai", "project", "graph", "edges.jsonl")
	if _, err := os.Stat(edgesFile); !os.IsNotExist(err) {
		t.Fatalf("precondition: fixture must start without an edges artifact (err: %v)", err)
	}

	if _, err := refreshEdgesArtifact(root, edgesFile); err != nil {
		t.Fatalf("refreshEdgesArtifact: %v", err)
	}

	if _, err := os.Stat(edgesFile); err != nil {
		t.Fatalf("stale tree: the refresh must have written the edges artifact: %v", err)
	}
	if edgesRefreshNeeded(root, edgesFile, graph.DefaultThresholds().MXIndexChangedFiles) {
		t.Error("staleness predicate must read false after the refresh (REQ-GR-010)")
	}
	if staged := stagedPorcelainLines(t, root); len(staged) != 0 {
		t.Errorf("REQ-GR-011: no entry may be staged by the refresh, staged: %v", staged)
	}
}
