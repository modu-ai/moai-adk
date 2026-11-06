// codex_audit_target_honesty_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card
// t1632, REQ-MN-019 / AC-MN-021): when a backend audit is asked for a declared
// target, the leg applies the resolved target — the resolved base SHA and the
// diff instruction present in the prompt, the result naming what was judged —
// or returns inconclusive stating the target was not applied. An adversarial
// codex pass (the codex_audit adversarial mode and the audit_multi codex leg)
// never silently reviews something else and returns a real pass/fail.
package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// emptyGitRepo is an initialised repository with no commit: no base SHA exists
// to diff against, so a baseBranch target cannot be applied there.
func emptyGitRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	cardScopeGit(t, repo, "init", "-q", "-b", "main")
	if err := os.MkdirAll(filepath.Join(repo, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	return repo
}

// TestCodexAudit_AdversarialBaseBranchAppliesTheTarget pins REQ-MN-019: an
// adversarial audit asked for baseBranch resolves the base and carries the
// resolved base SHA and the diff instruction IN THE PROMPT — the target is
// applied, not ignored — and the result names the base it judged.
func TestCodexAudit_AdversarialBaseBranchAppliesTheTarget(t *testing.T) {
	repo, _ := newIntegrationBaseRepo(t, "develop")
	mergeBase := reviewTargetGitOut(t, repo, "merge-base", "develop", "HEAD")
	sess := withCodexSession(t, adversarialScript())

	res, err := handleCodexAudit(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"mode": codexModeAdversarial, "target": codexTargetBaseBranch, "project_root": repo,
	}}})
	if err != nil {
		t.Fatalf("handleCodexAudit: %v", err)
	}
	prompt := sentTurnPrompt(t, sess)
	if !strings.Contains(prompt, mergeBase) {
		t.Errorf("turn/start prompt = %q, want it to carry the resolved base SHA %s (REQ-MN-019: the target is applied)", prompt, mergeBase)
	}
	if !strings.Contains(strings.ToLower(prompt), "diff") {
		t.Errorf("turn/start prompt = %q, want it to carry a diff instruction", prompt)
	}
	if base := reviewOutputField(res, "review_base"); !strings.Contains(base, mergeBase) {
		t.Errorf("review_base = %q, want it to name the base the prompt carried (%s)", base, mergeBase)
	}
}

// TestCodexAudit_AdversarialBaseBranchUnresolvableStatesNotApplied pins the
// honesty arm: when the base cannot be resolved the result is inconclusive
// stating the target was not applied — never a verdict about something else.
func TestCodexAudit_AdversarialBaseBranchUnresolvableStatesNotApplied(t *testing.T) {
	repo := emptyGitRepo(t)
	sess := withCodexSession(t, adversarialScript())

	res, err := handleCodexAudit(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"mode": codexModeAdversarial, "target": codexTargetBaseBranch, "project_root": repo,
	}}})
	if err != nil {
		t.Fatalf("handleCodexAudit: %v", err)
	}
	got := structuredMap(t, res)
	if v, _ := got["verdict"].(string); v != VerdictInconclusive {
		t.Errorf("verdict = %q, want inconclusive — an unresolvable target must not produce a verdict (REQ-MN-019)", v)
	}
	if summary, _ := got["summary"].(string); !strings.Contains(summary, "target not applied") {
		t.Errorf("summary = %q, want it to state the target was not applied", summary)
	}
	if len(sess.sent) != 0 {
		t.Errorf("no codex request may be sent when the target cannot be applied; sent=%v", sess.sent)
	}
}

// TestCodexAudit_AdversarialBaseBranchWithDeltaBaseComposes pins the
// composition with card t1708: when the request also carries delta_base the
// delta-scope clause is what applies the target (the delta SHA and its diff
// instruction ride the prompt, no contradicting merge-base diff instruction),
// and review_base names the delta base, not the resolved merge base.
func TestCodexAudit_AdversarialBaseBranchWithDeltaBaseComposes(t *testing.T) {
	d := newDeltaRepo(t)
	sess := withCodexSession(t, adversarialScript())

	res := callCodexAudit(t, deltaAuditArgs(d, map[string]any{"mode": codexModeAdversarial, "delta_base": d.shaB}))

	prompt := sentTurnPrompt(t, sess)
	if !strings.Contains(prompt, "git diff "+d.shaB+"...HEAD") {
		t.Errorf("prompt does not carry the delta scope; prompt:\n%s", prompt)
	}
	if strings.Contains(prompt, "git diff "+d.shaA+"..HEAD") {
		t.Errorf("prompt must not also instruct the merge-base diff %s; prompt:\n%s", d.shaA, prompt)
	}
	rb := reviewOutputField(res, "review_base")
	if !strings.Contains(rb, d.shaB) || !strings.Contains(rb, "delta_base") {
		t.Errorf("review_base = %q, want it to name delta_base %s (not the merge base %s)", rb, d.shaB, d.shaA)
	}
}

// TestPerformCodexAudit_BaseBranchAppliesTheTarget pins the same contract on the
// audit_multi codex leg: a baseBranch target carries the resolved base SHA and
// the diff instruction in the prompt and the leg result names the judged base.
func TestPerformCodexAudit_BaseBranchAppliesTheTarget(t *testing.T) {
	repo, _ := newIntegrationBaseRepo(t, "develop")
	mergeBase := reviewTargetGitOut(t, repo, "merge-base", "develop", "HEAD")
	sess := withCodexSession(t, adversarialScript())

	out := performCodexAudit(context.Background(), codexTargetBaseBranch, "", repo)

	prompt := sentTurnPrompt(t, sess)
	if !strings.Contains(prompt, mergeBase) {
		t.Errorf("codex leg prompt = %q, want it to carry the resolved base SHA %s (REQ-MN-019)", prompt, mergeBase)
	}
	if !strings.Contains(strings.ToLower(prompt), "diff") {
		t.Errorf("codex leg prompt = %q, want it to carry a diff instruction", prompt)
	}
	if !strings.Contains(out.ReviewBase, mergeBase) {
		t.Errorf("review_base = %q, want it to name the base the prompt carried (%s)", out.ReviewBase, mergeBase)
	}
}

// TestPerformCodexAudit_BaseBranchUnresolvableStatesNotApplied: an unresolvable
// base on the audit_multi codex leg is inconclusive naming the unapplied
// target, and codex is never asked.
func TestPerformCodexAudit_BaseBranchUnresolvableStatesNotApplied(t *testing.T) {
	repo := emptyGitRepo(t)
	sess := withCodexSession(t, adversarialScript())

	out := performCodexAudit(context.Background(), codexTargetBaseBranch, "", repo)

	if out.Verdict != VerdictInconclusive {
		t.Errorf("verdict = %q, want inconclusive — an unresolvable target must not produce a verdict (REQ-MN-019)", out.Verdict)
	}
	if !strings.Contains(out.Summary, "target not applied") {
		t.Errorf("summary = %q, want it to state the target was not applied", out.Summary)
	}
	if len(sess.sent) != 0 {
		t.Errorf("no codex request may be sent when the target cannot be applied; sent=%v", sess.sent)
	}
}
