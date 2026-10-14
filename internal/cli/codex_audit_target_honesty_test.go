// codex_audit_target_honesty_test.go — SPEC-MCP-NEUTRALITY-001 M6 (card
// t1632, REQ-MN-019): when a backend audit is asked for a declared target,
// the leg applies the resolved target — the resolved base SHA and the diff
// instruction present in the prompt — or returns inconclusive stating the
// target was not applied. An adversarial codex pass no longer silently
// ignores a baseBranch target.
package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// adversarialTurnStart extracts the turn/start params the fake session saw.
func adversarialTurnStart(t *testing.T, sess *fakeCodexSession) map[string]any {
	t.Helper()
	for _, line := range sess.sent {
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) == nil && m["method"] == codexMethodTurnStart {
			params, _ := m["params"].(map[string]any)
			return params
		}
	}
	t.Fatalf("no turn/start was sent; sent=%v", sess.sent)
	return nil
}

// turnStartPromptText extracts the prompt text the turn/start request carried
// (the params prompt rides the wire as the input array's first text item).
func turnStartPromptText(t *testing.T, params map[string]any) string {
	t.Helper()
	if p, ok := params["prompt"].(string); ok && p != "" {
		return p
	}
	input, _ := params["input"].([]any)
	for _, item := range input {
		m, _ := item.(map[string]any)
		if s, _ := m["text"].(string); s != "" {
			return s
		}
	}
	return ""
}

// TestCodexAudit_AdversarialBaseBranchAppliesTheTarget pins REQ-MN-019: an
// adversarial audit asked for baseBranch resolves the base and carries the
// resolved base SHA and the diff instruction IN THE PROMPT — the target is
// applied, not ignored — and the result names the base it judged.
func TestCodexAudit_AdversarialBaseBranchAppliesTheTarget(t *testing.T) {
	repo, _ := newIntegrationBaseRepo(t, "develop")
	mergeBase := reviewTargetGitOut(t, repo, "merge-base", "develop", "HEAD")
	sess := withCodexSession(t, codexSessionScript("clean change, no findings"))

	res, err := handleCodexAudit(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"mode": codexModeAdversarial, "target": codexTargetBaseBranch, "project_root": repo,
	}}})
	if err != nil {
		t.Fatalf("handleCodexAudit: %v", err)
	}
	turnStart := adversarialTurnStart(t, sess)
	prompt := turnStartPromptText(t, turnStart)
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
// honesty arm: when the base cannot be resolved (a repository with no commit
// to diff against), the result is inconclusive stating the target was not
// applied — never a verdict about something else.
func TestCodexAudit_AdversarialBaseBranchUnresolvableStatesNotApplied(t *testing.T) {
	// An EMPTY git repository (init, no commit): no base SHA exists to diff
	// against, so the target cannot be applied.
	repo := t.TempDir()
	cardScopeGit(t, repo, "init", "-q", "-b", "main")
	if err := os.MkdirAll(filepath.Join(repo, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}
	sess := withCodexSession(t, codexSessionScript("clean change, no findings"))
	_ = sess

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
}
