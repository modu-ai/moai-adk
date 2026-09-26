package cli

// SPEC-MCP-WORKTREE-UNTRACKED-001 — acceptance criteria AC-MWU-002..012 for the
// linked-worktree validator branch and the tree operations it feeds. Fixture F
// and the isolated-git helpers live in mcp_project_root_worktree_test.go.

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/graph"
)

// noGitPath points PATH at an empty directory so git cannot be found. Callers
// are non-parallel (t.Setenv).
func noGitPath(t *testing.T) {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
	if _, err := exec.LookPath("git"); err == nil {
		t.Fatal("premise: git must not be resolvable from the emptied PATH")
	}
}

// codexAuditOn calls the codex_audit handler on root with a codex seam that
// produces no verdict (binary not found), and decodes the result.
func codexAuditOn(t *testing.T, root string) ReviewOutput {
	t.Helper()
	withCodexLookPath(t, func(string) (string, error) { return "", exec.ErrNotFound })
	res, err := handleCodexAudit(context.Background(), newToolRequest(map[string]any{
		"project_root": root,
		"target":       "uncommittedChanges",
	}))
	if err != nil {
		t.Fatalf("handleCodexAudit hard error: %v", err)
	}
	if res.IsError {
		t.Fatalf("handleCodexAudit returned a tool error on %s: %s", root, resultTextOf(res))
	}
	var out ReviewOutput
	if err := json.Unmarshal([]byte(resultTextOf(res)), &out); err != nil {
		t.Fatalf("decode codex_audit result: %v\n%s", err, resultTextOf(res))
	}
	return out
}

const wtWorkflowCodexRequired = "workflow:\n  audit:\n    gates:\n      codex: required\n"

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

// newSeparateGitDirRepo builds a repository created with --separate-git-dir
// whose working tree S has an (untracked) .moai, plus a linked worktree W3.
func newSeparateGitDirRepo(t *testing.T) (s, w3 string) {
	t.Helper()
	env := wtGitEnv(t)
	base := wtCanonTempDir(t)
	s = filepath.Join(base, "S")
	wtGit(t, env, base, "init", "-q", "-b", "main", "--separate-git-dir", filepath.Join(base, "sepgit"), s)
	wtWriteFile(t, filepath.Join(s, ".gitignore"), ".moai/\n")
	wtWriteFile(t, filepath.Join(s, "tracked.txt"), "base\n")
	mustMkdir(t, filepath.Join(s, ".moai"))
	wtGit(t, env, s, "add", ".gitignore", "tracked.txt")
	wtGit(t, env, s, "commit", "-q", "-m", "init")
	w3 = filepath.Join(base, "W3")
	wtGit(t, env, s, "worktree", "add", "-q", "-b", "w3", w3)
	return s, w3
}

// AC-MWU-002: tree operations act on the worktree — the review diff is W's,
// and codex receives W as its cwd.
func TestLinkedWorktree_TreeOperationsTargetTheWorktree(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	wtWriteFile(t, filepath.Join(fx.W, "tracked.txt"), "change-in-W\n")
	wtWriteFile(t, filepath.Join(fx.P, "tracked.txt"), "change-in-P\n")

	root, err := resolveToolProjectRoot(newToolRequest(map[string]any{"project_root": fx.W}))
	if err != nil {
		t.Fatalf("resolve W: %v", err)
	}
	diff, err := collectReviewDiff(root, "uncommittedChanges")
	if err != nil {
		t.Fatalf("collectReviewDiff: %v", err)
	}
	if !strings.Contains(diff, "change-in-W") || strings.Contains(diff, "change-in-P") {
		t.Fatalf("diff must carry W's change only; got:\n%s", diff)
	}

	sess := withCodexSession(t, codexSessionScript("No issues found."))
	res, err := handleCodexAudit(context.Background(), newToolRequest(map[string]any{
		"project_root": fx.W, "target": "uncommittedChanges",
	}))
	if err != nil || res.IsError {
		t.Fatalf("codex_audit on W: err=%v res=%s", err, resultTextOf(res))
	}
	wantCwd := `"cwd":` + strconv.Quote(fx.W)
	found := false
	for _, line := range sess.sent {
		if strings.Contains(line, wantCwd) {
			found = true
		}
	}
	if !found {
		t.Fatalf("codex was not handed cwd %s; sent=%v", fx.W, sess.sent)
	}
}

// AC-MWU-003: an unrelated non-repository directory is rejected.
func TestLinkedWorktree_RejectsUnrelatedDirectory(t *testing.T) {
	dir := wtCanonTempDir(t)
	if _, err := validateProjectRoot(dir); err == nil {
		t.Fatal("an unrelated directory without .moai was accepted")
	}
	if _, err := resolveToolProjectRoot(newToolRequest(map[string]any{"project_root": dir})); err == nil {
		t.Fatal("resolver substituted a root for a rejected project_root")
	}
}

// AC-MWU-004: a non-MoAI repository's worktree, that repository's primary, and
// a subdirectory of W are all rejected, each naming its failed condition.
func TestLinkedWorktree_RejectsNonMoaiPrimarySelfAndSubdirectory(t *testing.T) {
	env := wtGitEnv(t)
	base := wtCanonTempDir(t)
	q := filepath.Join(base, "Q")
	mustMkdir(t, q)
	wtGit(t, env, q, "init", "-q", "-b", "main")
	wtWriteFile(t, filepath.Join(q, "f.txt"), "x\n")
	wtGit(t, env, q, "add", "f.txt")
	wtGit(t, env, q, "commit", "-q", "-m", "init")
	qw := filepath.Join(base, "QW")
	wtGit(t, env, q, "worktree", "add", "-q", "-b", "qw", qw)

	_, err := validateProjectRoot(qw)
	if err == nil || !strings.Contains(err.Error(), "primary checkout") || !strings.Contains(err.Error(), "has no .moai") {
		t.Fatalf("non-MoAI worktree: want error naming the primary's missing .moai, got %v", err)
	}
	t.Logf("non-MoAI worktree: %v", err)
	_, err = validateProjectRoot(q)
	if err == nil {
		t.Fatal("non-MoAI primary itself was accepted")
	}
	t.Logf("non-MoAI primary: %v", err)

	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	sub := filepath.Join(fx.W, "sub")
	mustMkdir(t, sub)
	_, err = validateProjectRoot(sub)
	if err == nil || !strings.Contains(err.Error(), "not the top level") {
		t.Fatalf("subdirectory of W: want non-top-level rejection, got %v", err)
	}
	t.Logf("subdirectory: %v", err)
}

// AC-MWU-005: once W's admin entry is removed, W is rejected as unregistered or
// unreadable.
func TestLinkedWorktree_RejectsUnregisteredWorktree(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	if err := os.RemoveAll(filepath.Join(fx.P, ".git", "worktrees", "W")); err != nil {
		t.Fatal(err)
	}
	_, err := validateProjectRoot(fx.W)
	if err == nil || !strings.Contains(err.Error(), "unregistered or unreadable") {
		t.Fatalf("want rejection naming the unregistered/unreadable worktree, got %v", err)
	}
	t.Logf("unregistered: %v", err)
}

// AC-MWU-006: a --separate-git-dir repository's worktree is an ambiguous layout.
func TestLinkedWorktree_RejectsSeparateGitDirLayout(t *testing.T) {
	_, w3 := newSeparateGitDirRepo(t)
	_, err := validateProjectRoot(w3)
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("want ambiguous-layout rejection, got %v", err)
	}
	t.Logf("separate git dir: %v", err)
}

// AC-MWU-007: a symlink to W resolves to W; a symlink to an unrelated
// directory is rejected.
func TestLinkedWorktree_SymlinkCanonicalization(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	links := wtCanonTempDir(t)
	l := filepath.Join(links, "L")
	if err := os.Symlink(fx.W, l); err != nil {
		t.Fatal(err)
	}
	got, err := validateProjectRoot(l)
	if err != nil || got != fx.W {
		t.Fatalf("symlink to W: got (%q, %v), want (%q, nil)", got, err, fx.W)
	}
	unrelated := filepath.Join(links, "U")
	if err := os.Symlink(wtCanonTempDir(t), unrelated); err != nil {
		t.Fatal(err)
	}
	if _, err := validateProjectRoot(unrelated); err == nil {
		t.Fatal("symlink to an unrelated directory was accepted")
	}
}

// AC-MWU-008: inherited GIT_DIR / GIT_WORK_TREE naming P do not redirect the
// inspection. Non-parallel (t.Setenv).
func TestLinkedWorktree_IgnoresInheritedGitEnvironment(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	unrelated := wtCanonTempDir(t)
	t.Setenv("GIT_DIR", filepath.Join(fx.P, ".git"))
	t.Setenv("GIT_WORK_TREE", fx.P)
	got, err := validateProjectRoot(fx.W)
	if err != nil || got != fx.W {
		t.Fatalf("W under hostile GIT_DIR: got (%q, %v), want (%q, nil)", got, err, fx.W)
	}
	if _, err := validateProjectRoot(unrelated); err == nil {
		t.Fatal("unrelated directory accepted under hostile GIT_DIR")
	}
}

// AC-MWU-009: with git unavailable, W is rejected (fail-closed).
func TestLinkedWorktree_FailsClosedWithoutGit(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	noGitPath(t)
	if got, err := validateProjectRoot(fx.W); err == nil {
		t.Fatalf("W accepted without git: %q", got)
	}
	if _, err := resolveToolProjectRoot(newToolRequest(map[string]any{"project_root": fx.W})); err == nil {
		t.Fatal("resolver substituted a default root without git")
	}
}

// AC-MWU-010: a dangling sibling worktree is not a reason to reject W.
func TestLinkedWorktree_DanglingSiblingDoesNotRejectW(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	w2 := filepath.Join(filepath.Dir(fx.W), "W2")
	wtGit(t, fx.env, fx.P, "worktree", "add", "-q", "-b", "w2", w2)
	if err := os.RemoveAll(w2); err != nil {
		t.Fatal(err)
	}
	t.Logf("porcelain after deleting W2:\n%s", wtGit(t, fx.env, fx.P, "worktree", "list", "--porcelain"))
	got, err := validateProjectRoot(fx.W)
	if err != nil || got != fx.W {
		t.Fatalf("W with a dangling sibling: got (%q, %v), want (%q, nil)", got, err, fx.W)
	}
}

// AC-MWU-011: a graph query on W never answers from P's graph artifact.
func TestLinkedWorktree_GraphNeverFromPrimary(t *testing.T) {
	fx := newUntrackedFixture(t, wtWorkflowNoGate)
	wtWriteFile(t, filepath.Join(fx.P, "internal", "svc", "svc.go"),
		"package svc\n\nfunc Run() { Finish() }\n\nfunc Finish() {}\n")
	edges, _, _, err := graph.BuildWithCodeLayers(fx.P)
	if err != nil {
		t.Fatal(err)
	}
	if err := graph.WriteJSONL(filepath.Join(fx.P, ".moai", "project", "graph", "edges.jsonl"), edges); err != nil {
		t.Fatal(err)
	}
	// Positive control: the same query on P does answer from P's graph.
	preq := mcp.CallToolRequest{}
	preq.Params.Arguments = map[string]any{"query": "Finish", "project_root": fx.P}
	pres, err := handleGraphFindCode(context.Background(), preq)
	if err != nil {
		t.Fatalf("positive control hard error: %v", err)
	}
	if body := graphToolJSON(t, pres); !strings.Contains(body, "svc.go") {
		t.Fatalf("positive control: P's graph should answer; body=%s", body)
	}

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"query": "Finish", "project_root": fx.W}
	res, err := handleGraphFindCode(context.Background(), req)
	if err != nil {
		t.Fatalf("handler hard error: %v", err)
	}
	blob, _ := json.Marshal(res)
	text := string(blob)
	if !res.IsError || !strings.Contains(text, "graph layer absent") {
		t.Fatalf("want the graph-layer-absent error for W, got: %s", text)
	}
	if strings.Contains(text, "svc.go") {
		t.Fatalf("W's graph query answered from P's graph: %s", text)
	}
}

// AC-MWU-013 (description half): the shared project_root text and the
// codex_audit description state the linked-worktree acceptance and where the
// audit gate comes from, without promising the receipt refusal on a worktree.
func TestLinkedWorktree_DescriptionsStateTheGateSource(t *testing.T) {
	for _, want := range []string{
		"linked worktree",
		"audit gate (workflow.audit.gates) is read from the primary checkout",
		"still read from the accepted tree",
	} {
		if !strings.Contains(projectRootDescCommon, want) {
			t.Errorf("projectRootDescCommon lacks %q", want)
		}
	}
	srv := newMoaiMCPServer()
	tool := srv.GetTool("codex_audit")
	if tool == nil {
		t.Fatal("codex_audit not registered")
	}
	desc := tool.Tool.Description
	if strings.Contains(desc, "where the reviewed tree explicitly sets") {
		t.Errorf("codex_audit description still says the gate comes only from the reviewed tree: %s", desc)
	}
	for _, want := range []string{"takes it from the primary checkout", "that refusal is not guaranteed"} {
		if !strings.Contains(desc, want) {
			t.Errorf("codex_audit description lacks %q: %s", want, desc)
		}
	}
}

// AC-MWU-012: today's accepted set is preserved and needs no git — a non-git
// directory with only .moai (or only .moai/specs/<id>) is accepted with git
// unavailable and GIT_DIR pointing nowhere. Non-parallel (t.Setenv).
func TestValidateProjectRoot_BareMoaiNeedsNoGit(t *testing.T) {
	bare := wtCanonTempDir(t)
	mustMkdir(t, filepath.Join(bare, ".moai"))
	specsOnly := newProbeProject(t, "SPEC-WTBARE-001")
	for _, dir := range []string{bare, specsOnly} {
		if got, err := validateProjectRoot(dir); err != nil || got != dir {
			t.Fatalf("normal run: got (%q, %v), want (%q, nil)", got, err, dir)
		}
	}
	noGitPath(t)
	t.Setenv("GIT_DIR", filepath.Join(t.TempDir(), "does-not-exist"))
	for _, dir := range []string{bare, specsOnly} {
		if got, err := validateProjectRoot(dir); err != nil || got != dir {
			t.Fatalf("git-less run: got (%q, %v), want (%q, nil)", got, err, dir)
		}
	}
}
