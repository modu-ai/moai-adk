package cli

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

func TestReviewCodexGrantsKeepSiblingProjectWorktreesOutside(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	home := m2Real(t, withRealCodexMoaiHomeAddDir(t))
	credentials := filepath.Join(home, "credentials")
	if err := os.Mkdir(credentials, 0700); err != nil {
		t.Fatal(err)
	}
	repo := m3SeedRepo(t, "selected")
	sibling := m3SeedRepo(t, "sibling")
	withCodexProjectRoot(t, repo)
	siblingTrees, err := worktreeroot.Root(sibling)
	if err != nil {
		t.Fatal(err)
	}
	siblingTrees = m2Real(t, siblingTrees)
	ownTrees, err := worktreeroot.Root(repo)
	if err != nil {
		t.Fatal(err)
	}
	ownTrees = m2Real(t, ownTrees)
	projectDB, err := homestate.ProjectDir(repo)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectDB, 0700); err != nil {
		t.Fatal(err)
	}
	projectDB = m2Real(t, projectDB)
	db, err := sql.Open("sqlite", filepath.Join(projectDB, "review.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err := db.Exec("PRAGMA journal_mode=WAL; CREATE TABLE sample (value TEXT); INSERT INTO sample VALUES ('retained')"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(projectDB, "review.db-wal")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runCodexCmd(t, "cli"); err != nil {
		t.Fatal(err)
	}
	argv := codexOnlyRecord(t, cap).Argv
	covered := false
	ownCovered := false
	for i := 0; i+1 < len(argv); i++ {
		if argv[i] != "--add-dir" {
			continue
		}
		grant := argv[i+1]
		for _, blocked := range []string{siblingTrees, filepath.Dir(siblingTrees), credentials} {
			if rel, err := filepath.Rel(grant, blocked); err == nil && rel != ".." && !filepath.IsAbs(rel) && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				t.Errorf("grant %q includes another project's or shared namespace %q", grant, blocked)
			}
		}
		if rel, err := filepath.Rel(grant, projectDB); err == nil && rel == "." {
			covered = true
		}
		if rel, err := filepath.Rel(grant, ownTrees); err == nil && rel == "." {
			ownCovered = true
		}
	}
	if !covered {
		t.Errorf("argv %q lacks the selected project's SQLite/WAL directory %q", argv, projectDB)
	}
	if !ownCovered {
		t.Errorf("argv %q lacks the selected project's worktree-creation namespace %q", argv, ownTrees)
	}
	// Bind the namespace to a real creation target, not only a derived name.
	newTree, err := worktreeroot.PathFor(repo, "next-card")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(newTree, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newTree, "probe.txt"), []byte("selected project\n"), 0600); err != nil {
		t.Fatal(err)
	}
}
