package factory

// m3_bases_test.go — SPEC-WORKTREE-HOME-ROOT-001 M3 RED (card t1640):
// AC-WHR-007's factory piece — DefaultWorktreeBases and ReadCardStatus must
// scan the project's home subdirectory
// (<MoaiHome>/worktrees/projects/<key>/<card>), not the bare
// ~/.moai/worktrees root.
//
// Written before the scan learns the home root: today a card whose live
// worktree stands at the home root resolves to the primary checkout's
// status (the scan misses it), which is the RED. The legacy bases are
// pinned by the existing suite and stay first-class.

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

// m3BaseFixture creates a scratch repository and a registered home-root
// worktree for card t7 on its own branch.
func m3BaseFixture(t *testing.T) (repo, homeTree string) {
	t.Helper()
	repo = filepath.Join(m3Real(t, t.TempDir()), "base-repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	git := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
		}
	}
	git(repo, "init", "-q", "--template=", "-b", "main")
	git(repo, "config", "user.email", "bases@example.com")
	git(repo, "config", "user.name", "Bases Test")
	git(repo, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	git(repo, "add", "README.md")
	git(repo, "commit", "-q", "-m", "seed")

	home := m3Real(t, t.TempDir())
	t.Setenv(paths.EnvHome, home)
	homeTree = filepath.Join(home, "worktrees", "projects", filepath.Base(repo), "t7")
	if err := os.MkdirAll(filepath.Dir(homeTree), 0o755); err != nil {
		t.Fatalf("mkdir home projects: %v", err)
	}
	// The branch names the card per REQ-KW-003's recognition rule (the
	// segment after its type prefix is the SPEC id): feature/t7, not WT-t7.
	git(repo, "worktree", "add", "-q", "-b", "feature/t7", "--", homeTree)
	// The fixture's project folder is claimed the production way (the
	// resolver's owner marker + repository-side record): trees under the
	// home root in production always arrive through a resolver path, and
	// the M8R key-scoped scan (card t1640 review P1-5) reads that record.
	if _, err := worktreeroot.Resolve(repo); err != nil {
		t.Fatalf("resolve the repository's key: %v", err)
	}
	return repo, homeTree
}

// m3Real resolves symlinks (the macOS /var alias) so git-derived paths
// agree with the asserted ones.
func m3Real(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatalf("evalsymlinks %s: %v", p, err)
	}
	return r
}

// TestReadCardStatus_FindsHomeRootWorktree pins the home scan: a card
// whose live worktree stands at the home root is reported from that
// worktree's branch, with the home path as the observed worktree.
func TestReadCardStatus_FindsHomeRootWorktree(t *testing.T) {
	repo, homeTree := m3BaseFixture(t)
	status, err := ReadCardStatus(repo, "t7", DefaultWorktreeBases(repo))
	if err != nil {
		t.Fatalf("ReadCardStatus: %v", err)
	}
	if status.WorktreePath != homeTree {
		t.Errorf("WorktreePath = %q, want the home-root worktree %q", status.WorktreePath, homeTree)
	}
	if status.WorktreeBranch != "feature/t7" {
		t.Errorf("WorktreeBranch = %q, want feature/t7", status.WorktreeBranch)
	}
	if status.Status == "unresolved" {
		t.Errorf("status = unresolved with candidates %v; the home-root worktree must resolve", status.Candidates)
	}
}

// TestDefaultWorktreeBases_LegacyBasesPinned pins the two legacy bases the
// M3 change must preserve (the home scan is exercised end-to-end by
// TestReadCardStatus_FindsHomeRootWorktree; the base-set shape it adds is
// an implementation detail of that observable).
func TestDefaultWorktreeBases_LegacyBasesPinned(t *testing.T) {
	repo, _ := m3BaseFixture(t)
	bases := DefaultWorktreeBases(repo)
	if bases.Claude != filepath.Join(repo, ".claude", "worktrees") {
		t.Errorf("bases.Claude = %q, want the legacy claude base", bases.Claude)
	}
	moaiBase, err := paths.WorktreesDir()
	if err != nil {
		t.Fatalf("WorktreesDir: %v", err)
	}
	if bases.MoAI != moaiBase {
		t.Errorf("bases.MoAI = %q, want %q", bases.MoAI, moaiBase)
	}
}

// TestReadCardStatus_HomeScanIsKeyScoped — card t1640 review P1-5: the
// home-root scan of a repository WITH a recorded key is scoped to that
// key's project folder; another project folder carrying the same specID —
// even one sorting first — is never selected (the home root is cross-
// project and the specID alone does not name a tree).
func TestReadCardStatus_HomeScanIsKeyScoped(t *testing.T) {
	repo, homeTree := m3BaseFixture(t)
	home := filepath.Dir(filepath.Dir(homeTree)) // .../worktrees/projects
	// The fixture created the project folder by hand; give the repository
	// its real recorded assignment (the resolver claims the same base name
	// and writes the marker and the record).
	if _, err := worktreeroot.Resolve(repo); err != nil {
		t.Fatalf("resolve the repository's key: %v", err)
	}
	// The decoy: a project folder that sorts FIRST and carries the same
	// specID on a token-matching branch — another repository's tree in the
	// shared root.
	decoy := filepath.Join(home, "000aaa", "t7")
	if err := os.MkdirAll(filepath.Dir(decoy), 0o755); err != nil {
		t.Fatalf("mkdir decoy: %v", err)
	}
	git := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
		}
	}
	git(repo, "worktree", "add", "-q", "-b", "feature/t7-decoy", "--", decoy)

	status, err := ReadCardStatus(repo, "t7", DefaultWorktreeBases(repo))
	if err != nil {
		t.Fatalf("ReadCardStatus: %v", err)
	}
	if status.WorktreePath != homeTree {
		t.Errorf("WorktreePath = %q, want this repository's tree %q; the scan selected a foreign project folder", status.WorktreePath, homeTree)
	}
	if status.WorktreeBranch != "feature/t7" {
		t.Errorf("WorktreeBranch = %q, want feature/t7 (the decoy's branch leaked in)", status.WorktreeBranch)
	}
}
