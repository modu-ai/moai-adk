package factory

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// t1589 — the integration-tree cleanliness check reads the repository named by
// an inherited GIT_DIR instead of the tree it was given. A dirty primary then
// reads as clean and the merge step proceeds.
func TestGitIntegrationWorktreeCleanIgnoresInheritedGitEnv(t *testing.T) {
	dirty := t1589FactoryRepo(t)
	if err := os.WriteFile(filepath.Join(dirty, "untracked.txt"), []byte("work\n"), 0o644); err != nil {
		t.Fatalf("write untracked file: %v", err)
	}
	clean := t1589FactoryRepo(t)
	t.Setenv("GIT_DIR", filepath.Join(clean, ".git"))
	t.Setenv("GIT_WORK_TREE", clean)

	ok, status, err := gitIntegrationWorktreeClean(dirty)
	if err != nil {
		t.Fatalf("gitIntegrationWorktreeClean() error = %v", err)
	}
	if ok {
		t.Errorf("gitIntegrationWorktreeClean(dirty) reported clean (status %q); the inherited GIT_DIR read the clean repository", status)
	}
}

// t1589FactoryRepo makes a one-commit repository in a temp directory.
func t1589FactoryRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"},
		{"-c", "user.name=test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "--no-gpg-sign", "-m", "init"},
	} {
		full := append([]string{"-C", dir}, args...)
		if out, err := exec.Command("git", full...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

// t1834G1Git runs git in dir with the repo-scoping environment removed, so the
// fixture setup and the expected values never read the inherited decoy.
func t1834G1Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(gitenv.Env(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t.local",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t.local",
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v: %s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// t1834 sync-audit iter9 G1 — RunRemeasure's own HEAD, merge-base and tree reads
// ran git without the repo-scoping scrub the clean check (t1589) and the gate's
// runner apply. With GIT_DIR / GIT_WORK_TREE naming ANOTHER repository, the
// record was keyed by that repository's tree, and the base-freshness gate — which
// reads the candidate's own tree — reported "no valid re-measure record" for a
// candidate that had just been remeasured.
func TestRemeasureKeysRecordByOwnTreeUnderInheritedGitEnv(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	primary, card, decoy := filepath.Join(root, "primary"), filepath.Join(root, "card"), filepath.Join(root, "decoy")
	store := filepath.Join(root, "store")
	for _, dir := range []string{primary, decoy, store} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeFile := func(dir, name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// The repository under test: main carries the landing the card has absorbed;
	// the card is a linked worktree one commit past it.
	t1834G1Git(t, primary, "init", "-q", "-b", "main")
	writeFile(primary, "base.txt", "base\n")
	t1834G1Git(t, primary, "add", "base.txt")
	t1834G1Git(t, primary, "commit", "-q", "-m", "base")
	t1834G1Git(t, primary, "commit", "-q", "--allow-empty", "-m", "landing")
	t1834G1Git(t, primary, "worktree", "add", "-q", "-b", "WT-card", card, "main")
	writeFile(card, "card.txt", "card work\n")
	t1834G1Git(t, card, "add", "card.txt")
	t1834G1Git(t, card, "commit", "-q", "-m", "card work")
	wantTree := t1834G1Git(t, card, "rev-parse", "HEAD^{tree}")
	wantBase := t1834G1Git(t, primary, "rev-parse", "main")

	// A different repository, also with a `main`, whose HEAD tree is not the card's.
	t1834G1Git(t, decoy, "init", "-q", "-b", "main")
	writeFile(decoy, "other.txt", "other repository\n")
	t1834G1Git(t, decoy, "add", "other.txt")
	t1834G1Git(t, decoy, "commit", "-q", "-m", "other")
	decoyTree := t1834G1Git(t, decoy, "rev-parse", "HEAD^{tree}")
	if decoyTree == wantTree {
		t.Fatalf("fixture: the decoy tree must differ from the card tree")
	}

	t.Setenv("GIT_DIR", filepath.Join(decoy, ".git"))
	t.Setenv("GIT_WORK_TREE", decoy)

	rec, err := RunRemeasure(store, card, "main", "true")
	if err != nil {
		t.Fatalf("remeasure of a clean linked worktree must succeed under an inherited GIT_DIR: %v", err)
	}
	if rec.Tree != wantTree {
		t.Errorf("record keyed by tree %s, want the card's own tree %s (the decoy's tree is %s)", rec.Tree, wantTree, decoyTree)
	}
	if rec.Base != wantBase {
		t.Errorf("record names base %s, want the card repository's absorbed main tip %s", rec.Base, wantBase)
	}

	// The gate's own read of the candidate (ExecGitRunner, scrubbed) must find it.
	res := EvaluateBaseFreshness(BaseFreshnessInput{
		Root: store, Dir: card, CardID: "t1", IntegrationBranch: "main", K: 0, Enabled: true,
	}, BaseFreshnessSeams{})
	if !res.Admit {
		t.Errorf("the gate must find the record the remeasure just wrote, refused: %s", res.Refusal)
	}
}
