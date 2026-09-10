package cli

// SPEC-CODEX-REVIEW-TARGET-001 (card t399, issue modu-ai/moai-adk#1632) —
// the codex `review/start` target must satisfy the required-field set of its own
// variant, as declared by the measured ReviewStartParams schema (codex-cli
// 0.150.1, .moai/reports/t399/schema/v2/ReviewStartParams.json):
//
//	uncommittedChanges  required [type]                 ← a bare string IS valid
//	baseBranch          required [branch, type]         ← a bare string is NOT
//	commit              required [sha, type]            ← a bare string is NOT
//	custom              required [instructions, type]   ← a bare string is NOT
//
// Every assertion here observes the SERIALIZED REQUEST BYTES (sess.sent[2]),
// never the stub's return value: the script stub replays a canned response
// regardless of what was asked, so a malformed request still "passes" at the
// stub (spec.md §A.5, acceptance.md §A). Counting `inconclusive` as a pass is
// likewise prohibited — it is the defect this SPEC closes.

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/modu-ai/moai-adk/internal/config"
)

// ─── §B schema table — the classification that drives AC-CRT-006 / 006b ───

// codexReviewTargetVariant is one row of acceptance.md §B. `serializable` is the
// `분류` column: whether the server can populate every required field of that
// variant. A fifth variant is added by appending a row; neither property
// assertion changes.
type codexReviewTargetVariant struct {
	name         string
	required     []string
	serializable bool
}

var codexReviewTargetVariants = []codexReviewTargetVariant{
	{name: codexTargetUncommitted, required: []string{"type"}, serializable: true},
	{name: codexTargetBaseBranch, required: []string{"branch", "type"}, serializable: true},
	{name: "commit", required: []string{"sha", "type"}, serializable: false},
	{name: "custom", required: []string{"instructions", "type"}, serializable: false},
}

// ─── fixtures ───

// reviewTargetGit runs one git command in repo, failing the test on error.
func reviewTargetGit(t *testing.T, repo string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// newReviewTargetRepo seeds a MoAI-shaped temp git repo with one commit. The
// `.moai` directory is required: validateProjectRoot rejects a project_root
// without one.
func newReviewTargetRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	reviewTargetGit(t, repo, "init")
	reviewTargetGit(t, repo, "config", "user.email", "t@t.test")
	reviewTargetGit(t, repo, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(repo, "seed.txt"), []byte("seed\n"), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	reviewTargetGit(t, repo, "add", "seed.txt", ".moai")
	reviewTargetGit(t, repo, "commit", "-m", "seed")
	// Whatever the local git default-branch setting is, pin the seeded branch to
	// a name that is NOT a chain step, so `main` only ever resolves because a
	// fixture deliberately created it.
	reviewTargetGit(t, repo, "branch", "-M", "wt-fixture")
	return repo
}

// seedRemoteMain adds refs/remotes/origin/main and points refs/remotes/origin/HEAD
// at it — the chain's step 1.
func seedRemoteMain(t *testing.T, repo string) {
	t.Helper()
	sha := reviewTargetGitOut(t, repo, "rev-parse", "HEAD")
	reviewTargetGit(t, repo, "update-ref", "refs/remotes/origin/main", sha)
	reviewTargetGit(t, repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
}

func reviewTargetGitOut(t *testing.T, repo string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out))
}

// writeWorktreeBaseBranchConfig writes git_strategy.worktree_base_branch so the
// AC-CRT-002 [HARD] non-read clause has a fixture where the config value and
// origin/HEAD DIVERGE. In a tree where the two coincide the two designs are
// indistinguishable, so the check would have no teeth (spec.md §A.7).
func writeWorktreeBaseBranchConfig(t *testing.T, repo, value string) {
	t.Helper()
	dir := filepath.Join(repo, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	body := "git_strategy:\n    worktree_base_branch: " + value + "\n"
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}
}

// ─── observation helpers (serialized bytes only) ───

// sentReviewStart returns the parsed params of the review/start request the
// session actually serialized, and whether one was sent at all.
func sentReviewStart(t *testing.T, sess *fakeCodexSession) (map[string]any, bool) {
	t.Helper()
	for _, line := range sess.sent {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		if method, _ := m["method"].(string); method != codexMethodReviewStart {
			continue
		}
		params, _ := m["params"].(map[string]any)
		return params, true
	}
	return nil, false
}

// runNativeAudit drives handleCodexAudit in native mode against a stub session
// and returns the session (for byte observation) plus the tool result.
func runNativeAudit(t *testing.T, root, target string) (*fakeCodexSession, *mcp.CallToolResult) {
	t.Helper()
	sess := withCodexSession(t, codexSessionScript("clean change, no findings"))
	args := map[string]any{"mode": codexModeNative, "target": target}
	if root != "" {
		args["project_root"] = root
	}
	res, err := handleCodexAudit(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: args},
	})
	if err != nil {
		t.Fatalf("handleCodexAudit(%s): %v", target, err)
	}
	return sess, res
}

// reviewOutputField reads a top-level string field out of the tool result.
func reviewOutputField(res *mcp.CallToolResult, key string) string {
	for _, m := range resultMaps(res) {
		if v, ok := m[key].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// ─── AC-CRT-001 ───

// TestCodexAudit_NativeBaseBranchCarriesBranch — AC-CRT-001 / REQ-CRT-002.
// The 3rd sent request is review/start; its params.target is an object whose
// type is baseBranch AND whose `branch` is present and non-empty.
func TestCodexAudit_NativeBaseBranchCarriesBranch(t *testing.T) {
	repo := newReviewTargetRepo(t)
	seedRemoteMain(t, repo)

	sess, _ := runNativeAudit(t, repo, codexTargetBaseBranch)

	params, ok := sentReviewStart(t, sess)
	if !ok {
		t.Fatalf("no review/start was serialized; sent=%v", sess.sent)
	}
	target, ok := params["target"].(map[string]any)
	if !ok {
		t.Fatalf("review/start target must be a JSON object; got %T (%v)", params["target"], params["target"])
	}
	if got, _ := target["type"].(string); got != codexTargetBaseBranch {
		t.Errorf("target.type = %q, want %q", got, codexTargetBaseBranch)
	}
	branch, _ := target["branch"].(string)
	if strings.TrimSpace(branch) == "" {
		t.Errorf("baseBranch target must carry a non-empty `branch` (schema required set is [branch, type]); target=%v", target)
	}
}

// ─── AC-CRT-002 ───

// TestCodexAudit_BaseBranchResolutionChain — AC-CRT-002 / REQ-CRT-003.
// The chain is the one resolveReviewMergeBase uses, read at the NAME layer:
// the remote default head, then `main`. git_strategy.worktree_base_branch is
// NOT a step (spec.md §A.7).
func TestCodexAudit_BaseBranchResolutionChain(t *testing.T) {
	t.Run("step1_remote_default_head", func(t *testing.T) {
		repo := newReviewTargetRepo(t)
		seedRemoteMain(t, repo)

		sess, _ := runNativeAudit(t, repo, codexTargetBaseBranch)
		if got := sentTargetBranch(t, sess); got != "main" {
			t.Errorf("target.branch = %q, want %q (origin/HEAD → origin/main, prefix stripped)", got, "main")
		}
	})

	t.Run("step2_main_when_remote_head_absent", func(t *testing.T) {
		repo := newReviewTargetRepo(t)
		// No origin/HEAD at all; `main` exists as a local branch.
		reviewTargetGit(t, repo, "branch", "main")

		sess, _ := runNativeAudit(t, repo, codexTargetBaseBranch)
		if got := sentTargetBranch(t, sess); got != "main" {
			t.Errorf("target.branch = %q, want %q — a missing step 1 must not become a silent skip", got, "main")
		}
	})

	// [HARD] the non-read clause. The fixture DIVERGES: the config key names a
	// branch that exists (so a config-reading resolver would happily return it)
	// while origin/HEAD points at main. Only a fixture where the two differ can
	// tell the two designs apart.
	t.Run("worktree_base_branch_is_not_read", func(t *testing.T) {
		repo := newReviewTargetRepo(t)
		seedRemoteMain(t, repo)
		reviewTargetGit(t, repo, "branch", "divergent-base")
		writeWorktreeBaseBranchConfig(t, repo, "divergent-base")

		// Prove the fixture actually diverges — otherwise this sub-test asserts
		// nothing (the coinciding-values trap spec.md §A.7 names).
		if got := config.LoadWorktreeBaseBranch(repo); got != "divergent-base" {
			t.Fatalf("fixture is not divergent: LoadWorktreeBaseBranch = %q, want %q", got, "divergent-base")
		}

		sess, _ := runNativeAudit(t, repo, codexTargetBaseBranch)
		if got := sentTargetBranch(t, sess); got != "main" {
			t.Errorf("target.branch = %q, want %q — worktree_base_branch must NOT be read on this path (spec.md §A.7)", got, "main")
		}
	})

	// Each step confirms the name it is about to return resolves as a ref in
	// that tree (REQ-CRT-003 second half). A dangling origin/HEAD is the case
	// that separates "confirmed" from "assumed": the symbolic-ref reads fine,
	// the branch it names does not exist.
	t.Run("dangling_remote_head_falls_through", func(t *testing.T) {
		repo := newReviewTargetRepo(t)
		reviewTargetGit(t, repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/ghost")
		reviewTargetGit(t, repo, "branch", "main")

		sess, _ := runNativeAudit(t, repo, codexTargetBaseBranch)
		if got := sentTargetBranch(t, sess); got != "main" {
			t.Errorf("target.branch = %q, want %q — an unresolvable name must not be returned", got, "main")
		}
	})
}

// sentTargetBranch extracts params.target.branch from the serialized
// review/start, failing when no review/start was sent.
func sentTargetBranch(t *testing.T, sess *fakeCodexSession) string {
	t.Helper()
	params, ok := sentReviewStart(t, sess)
	if !ok {
		t.Fatalf("no review/start was serialized; sent=%v", sess.sent)
	}
	target, _ := params["target"].(map[string]any)
	branch, _ := target["branch"].(string)
	return branch
}

// ─── AC-CRT-003 (regression line — GREEN before the change) ───

// TestCodexAudit_UncommittedChangesShapeUnchanged — AC-CRT-003 / REQ-CRT-006.
// The uncommittedChanges request stays shape-identical to its form at
// 442da4f06: exactly one key, `type`, and none of branch / sha / instructions.
func TestCodexAudit_UncommittedChangesShapeUnchanged(t *testing.T) {
	repo := newReviewTargetRepo(t)
	seedRemoteMain(t, repo)

	sess, _ := runNativeAudit(t, repo, codexTargetUncommitted)

	params, ok := sentReviewStart(t, sess)
	if !ok {
		t.Fatalf("no review/start was serialized; sent=%v", sess.sent)
	}
	target, ok := params["target"].(map[string]any)
	if !ok {
		t.Fatalf("review/start target must be a JSON object; got %T", params["target"])
	}
	if len(target) != 1 {
		t.Errorf("uncommittedChanges target must carry exactly one key; got %v", target)
	}
	if got, _ := target["type"].(string); got != codexTargetUncommitted {
		t.Errorf("target.type = %q, want %q", got, codexTargetUncommitted)
	}
	for _, forbidden := range []string{"branch", "sha", "instructions"} {
		if _, has := target[forbidden]; has {
			t.Errorf("uncommittedChanges target must not carry %q; got %v", forbidden, target)
		}
	}
}

// ─── AC-CRT-004 ───

// TestCodexAudit_UnresolvableBaseBranchIsNotASilentOtherReview — AC-CRT-004 /
// REQ-CRT-004 + REQ-CRT-008. Four things are judged: no review/start line, no
// substituted uncommittedChanges request, the named output field's value, and
// its distinguishability from every other fail-open cause.
func TestCodexAudit_UnresolvableBaseBranchIsNotASilentOtherReview(t *testing.T) {
	repo := newReviewTargetRepo(t) // no origin/HEAD, no origin/main, no main

	sess, res := runNativeAudit(t, repo, codexTargetBaseBranch)

	if params, ok := sentReviewStart(t, sess); ok {
		t.Errorf("review/start must NOT be sent when the base branch cannot be resolved; params=%v", params)
	}
	for _, line := range sess.sent {
		if strings.Contains(line, `"`+codexTargetUncommitted+`"`) {
			t.Errorf("an unresolvable baseBranch must not be substituted by uncommittedChanges; line:\n%s", line)
		}
	}

	// Candidate (가) from plan.md §B: a cause-naming inconclusive that still
	// passes through applyGateUnmet.
	if got := reviewOutputField(res, "verdict"); got != VerdictInconclusive {
		t.Errorf("verdict = %q, want %q", got, VerdictInconclusive)
	}
	summary := reviewOutputField(res, "summary")
	if !strings.Contains(strings.ToLower(summary), "base branch") {
		t.Errorf("summary must name the unresolvable base branch as the cause; got %q", summary)
	}
	if summary == "codex binary not found in PATH" {
		t.Errorf("summary must be distinguishable from every other fail-open cause; got %q", summary)
	}
}

// ─── AC-CRT-005 ───

// TestCodexAudit_IncompleteVariantsAreNotSerialized — AC-CRT-005 / REQ-CRT-005.
// A bare string naming a variant whose required fields the server cannot
// populate must never be lifted into an incomplete object.
func TestCodexAudit_IncompleteVariantsAreNotSerialized(t *testing.T) {
	repo := newReviewTargetRepo(t)
	seedRemoteMain(t, repo)

	for _, v := range codexReviewTargetVariants {
		if v.serializable {
			continue
		}
		t.Run(v.name, func(t *testing.T) {
			sess, _ := runNativeAudit(t, repo, v.name)
			for _, line := range sess.sent {
				if strings.Contains(line, `"type":"`+v.name+`"`) {
					t.Errorf("an incomplete %s target must never be serialized (required %v); line:\n%s",
						v.name, v.required, line)
				}
			}
		})
	}
}

// ─── AC-CRT-006 ───

// TestCodexReviewTarget_SerializableVariantsSatisfyRequiredSet — AC-CRT-006 /
// REQ-CRT-001. Property assertion over the §B rows classified serializable.
func TestCodexReviewTarget_SerializableVariantsSatisfyRequiredSet(t *testing.T) {
	repo := newReviewTargetRepo(t)
	seedRemoteMain(t, repo)

	visited := 0
	for _, v := range codexReviewTargetVariants {
		if !v.serializable {
			continue
		}
		visited++
		t.Run(v.name, func(t *testing.T) {
			sess, _ := runNativeAudit(t, repo, v.name)
			params, ok := sentReviewStart(t, sess)
			if !ok {
				t.Fatalf("no review/start was serialized for %s; sent=%v", v.name, sess.sent)
			}
			target, ok := params["target"].(map[string]any)
			if !ok {
				t.Fatalf("target must be a JSON object; got %T", params["target"])
			}
			for _, key := range v.required {
				val, has := target[key]
				if !has {
					t.Errorf("%s target is missing required key %q; got %v", v.name, key, target)
					continue
				}
				if s, isStr := val.(string); isStr && strings.TrimSpace(s) == "" {
					t.Errorf("%s target's required key %q is empty", v.name, key)
				}
			}
		})
	}
	// A classification that selected nothing asserts nothing — that is a defect,
	// not a pass (acceptance.md §B / §A).
	if visited == 0 {
		t.Fatal("no serializable variant rows were traversed; the §B classification selected an empty set")
	}
}

// ─── AC-CRT-006b ───

// TestCodexReviewTarget_UnserializableVariantsLeaveNoTarget — AC-CRT-006b /
// REQ-CRT-001 + REQ-CRT-005. The paired property: the variant's own target must
// not appear, AND no review/start carrying another variant's target may be sent
// in its place. Without the second clause the current default branch
// (mcp_codex.go:1004 → uncommittedChanges) satisfies the first one silently.
func TestCodexReviewTarget_UnserializableVariantsLeaveNoTarget(t *testing.T) {
	repo := newReviewTargetRepo(t)
	seedRemoteMain(t, repo)

	visited := 0
	for _, v := range codexReviewTargetVariants {
		if v.serializable {
			continue
		}
		visited++
		t.Run(v.name, func(t *testing.T) {
			sess, _ := runNativeAudit(t, repo, v.name)
			params, sent := sentReviewStart(t, sess)
			if !sent {
				return // nothing serialized at all — both clauses hold
			}
			target, _ := params["target"].(map[string]any)
			got, _ := target["type"].(string)
			if got == v.name {
				t.Errorf("the %s target object must not appear; target=%v", v.name, target)
			}
			t.Errorf("no review/start may be sent in place of an unsupported %s target; substituted target=%v",
				v.name, target)
		})
	}
	if visited == 0 {
		t.Fatal("no unserializable variant rows were traversed; the §B classification selected an empty set")
	}
}

// ─── AC-CRT-009 ───

// TestCodexAuditToolSurface_DescribesServerSideBranchResolution — AC-CRT-009 /
// REQ-CRT-007. The caller cannot supply a branch (the target parameter is a bare
// string enum with no companion), so a description that omits the server-side
// resolution implies a value the caller has no way to provide.
func TestCodexAuditToolSurface_DescribesServerSideBranchResolution(t *testing.T) {
	srv := newMoaiMCPServer()
	if srv == nil {
		t.Fatal("newMoaiMCPServer returned nil")
	}
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)

	ctx := context.Background()
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	tools, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}

	var desc string
	for _, tool := range tools.Tools {
		if tool.Name != "codex_audit" {
			continue
		}
		props, _ := tool.InputSchema.Properties["target"].(map[string]any)
		desc, _ = props["description"].(string)
	}
	if desc == "" {
		t.Fatal("codex_audit declares no description for its target parameter")
	}
	lower := strings.ToLower(desc)
	// The fact: the branch name is resolved server-side, not supplied.
	if !strings.Contains(lower, "server-side") {
		t.Errorf("target description must state that the baseBranch branch name is resolved server-side; got %q", desc)
	}
	// The source: which chain that resolution reads.
	if !strings.Contains(lower, "remote default head") || !strings.Contains(lower, "main") {
		t.Errorf("target description must name the resolution source (the remote default head, then main); got %q", desc)
	}
}
