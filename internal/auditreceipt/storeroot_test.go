package auditreceipt

// SPEC-WORKTREE-STATE-ROOT-001 — the store-root answer and the per-tree
// rejection records. Non-parallel (t.Setenv in the no-git case).

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func srGitEnv(t *testing.T) []string {
	t.Helper()
	empty := filepath.Join(t.TempDir(), "empty-gitconfig")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_") {
			env = append(env, kv)
		}
	}
	return append(env, "GIT_CONFIG_GLOBAL="+empty, "GIT_CONFIG_SYSTEM="+empty, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid")
}

func srGit(t *testing.T, env []string, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func srWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// srRepo builds a repository P (tracking the workflow config when tracked is
// true, ignoring .moai otherwise) with one linked worktree W.
func srRepo(t *testing.T, tracked bool) (p, w string) {
	t.Helper()
	env := srGitEnv(t)
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p = filepath.Join(base, "P")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	srGit(t, env, p, "init", "-q", "-b", "main")
	srWrite(t, filepath.Join(p, ".moai", "config", "sections", "workflow.yaml"), "workflow: {}\n")
	if tracked {
		srGit(t, env, p, "add", ".moai/config/sections/workflow.yaml")
	} else {
		srWrite(t, filepath.Join(p, ".gitignore"), ".moai/\n")
		srGit(t, env, p, "add", ".gitignore")
	}
	srGit(t, env, p, "commit", "-q", "-m", "init")
	w = filepath.Join(base, "W")
	srGit(t, env, p, "worktree", "add", "-q", "-b", "wt", w)
	return p, w
}

// AC-WSR-014 (store-root half): every root maps to one store root — P for a
// config-orphaned W, itself otherwise, unresolved when W's primary cannot be
// identified.
func TestWSR014_StoreRootPerRoot(t *testing.T) {
	p, w := srRepo(t, false)
	tp, wt := srRepo(t, true)
	b, _ := filepath.EvalSymlinks(t.TempDir())
	srWrite(t, filepath.Join(b, ".moai", "config", "sections", "workflow.yaml"), "workflow: {}\n")

	for _, tc := range []struct{ label, root, want string }{
		{"P", p, p}, {"W", w, p}, {"WT", wt, wt}, {"T primary", tp, tp}, {"B", b, b},
	} {
		got, err := StoreRoot(tc.root)
		if err != nil || got != tc.want {
			t.Errorf("%s: StoreRoot = %q, %v; want %q", tc.label, got, err, tc.want)
		}
	}

	if err := os.Remove(filepath.Join(p, ".git", "worktrees", "W", "HEAD")); err != nil {
		t.Fatal(err)
	}
	got, err := StoreRoot(w)
	var unresolved *UnresolvedStoreError
	if got != "" || !errors.As(err, &unresolved) || !strings.Contains(err.Error(), "primary checkout") {
		t.Errorf("W with primary unidentifiable: StoreRoot = %q, %v; want unresolved naming the primary checkout", got, err)
	}

	// A root that is not config-orphaned runs no git: it resolves identically
	// with git unavailable.
	t.Setenv("PATH", t.TempDir())
	if got, err := StoreRoot(wt); err != nil || got != wt {
		t.Errorf("WT without git: StoreRoot = %q, %v; want %q", got, err, wt)
	}
}

// Records of different trees coexist in one store; list and clear are per
// tree; a legacy record without tree identity belongs to the store root.
func TestRejectionsInSharedStoreArePerTree(t *testing.T) {
	store := t.TempDir()
	w, w2 := filepath.Join(store, "W"), filepath.Join(store, "W2")
	for _, tree := range []string{w, w2, store} {
		if err := WriteRejectionIn(store, &Rejection{AgentType: AgentPlanAuditor, SpecID: "SPEC-X-001", Cause: "c", TreeRoot: tree}); err != nil {
			t.Fatal(err)
		}
	}
	legacy := Rejection{AgentType: AgentSyncAuditor, SpecID: "SPEC-L-001", Cause: "old"}
	if err := WriteRejection(store, &legacy); err != nil {
		t.Fatal(err)
	}
	all, err := ListRejections(store)
	if err != nil || len(all) != 4 {
		t.Fatalf("want 4 records in the store, got %d (%v)", len(all), err)
	}
	if got, _ := ListRejectionsForTree(store, w); len(got) != 1 || got[0].TreeRoot != w {
		t.Errorf("W's list = %+v, want W's record only", got)
	}
	if got, _ := ListRejectionsForTree(store, store); len(got) != 2 {
		t.Errorf("store root's own list = %+v, want its own record plus the legacy one", got)
	}
	if err := ClearRejectionsForRoleInTree(store, w, AgentPlanAuditor); err != nil {
		t.Fatal(err)
	}
	if got, _ := ListRejectionsForTree(store, w); len(got) != 0 {
		t.Errorf("W's rejection must be cleared, got %+v", got)
	}
	if got, _ := ListRejectionsForTree(store, w2); len(got) != 1 {
		t.Errorf("W2's rejection must survive W's clear, got %+v", got)
	}
	if _, err := ReadRejectionIn(store, store, AgentPlanAuditor, "SPEC-X-001"); err != nil {
		t.Errorf("the store root's own record must survive W's clear and keep the legacy file name: %v", err)
	}
	if _, err := os.Stat(filepath.Join(StateDir(store), rejectionsRel, rejectionFileName(AgentPlanAuditor, "SPEC-X-001"))); err != nil {
		t.Errorf("the store root's own record must use the pre-existing file name: %v", err)
	}
}
