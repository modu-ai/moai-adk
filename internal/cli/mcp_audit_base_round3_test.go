package cli

// Card t1426, second independent review of ea0a3b22b (three P2):
//
//  1. A base that exists only as a remote-tracking ref (origin/develop, no
//     local develop) must be passed as the ref that was actually selected —
//     never as a bare name that does not resolve in the tree.
//  2. codex must be sent the captured merge-base SHA, not a branch name it
//     re-resolves later, so review_base names exactly what codex compared.
//  3. An adversarial (turn/start) audit carries no target and names no base,
//     so it must not claim a review_base.

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestCodexAudit_RemoteOnlyBaseIsSentAsAResolvableRevision(t *testing.T) {
	repo, developTip := newIntegrationBaseRepo(t, "develop")
	reviewTargetGit(t, repo, "update-ref", "refs/remotes/origin/develop", developTip)
	reviewTargetGit(t, repo, "branch", "-D", "develop")

	sess, res := runNativeAudit(t, repo, codexTargetBaseBranch)
	sent := sentTargetBranch(t, sess)
	out, err := exec.Command("git", "-C", repo, "rev-parse", "--verify", "--quiet", sent+"^{commit}").Output()
	if err != nil {
		t.Errorf("codex target.branch %q does not resolve in the tree: %v", sent, err)
	} else if got := strings.TrimSpace(string(out)); got != developTip {
		t.Errorf("codex target.branch %q resolves to %s, want the develop merge base %s", sent, got, developTip)
	}
	if base := reviewOutputField(res, "review_base"); !strings.Contains(base, "origin/develop") || !strings.Contains(base, developTip) {
		t.Errorf("review_base = %q, want the selected ref origin/develop and merge base %s", base, developTip)
	}
	diff, err := collectReviewDiff(repo, codexTargetBaseBranch)
	if err != nil {
		t.Fatalf("collectReviewDiff: %v", err)
	}
	if strings.Contains(diff, "develop-only.txt") || !strings.Contains(diff, "card.txt") {
		t.Errorf("GLM material must be the card diff only; got:\n%s", diff)
	}
}

func TestCodexAudit_SendsTheCapturedMergeBaseNotABranchName(t *testing.T) {
	repo, developTip := newIntegrationBaseRepo(t, "develop")
	cardTip := reviewTargetGitOut(t, repo, "rev-parse", "HEAD")
	withCodexSession(t, nil)
	sess := &movingRefCodexSession{
		fakeCodexSession: fakeCodexSession{lines: codexSessionScript("clean change, no findings")},
		t:                t, repo: repo, toSHA: cardTip,
	}
	codexSession = sess

	res, err := handleCodexAudit(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"mode": codexModeNative, "target": codexTargetBaseBranch, "project_root": repo,
	}}})
	if err != nil {
		t.Fatalf("handleCodexAudit: %v", err)
	}
	sent := sentTargetBranch(t, &sess.fakeCodexSession)
	if sent != developTip {
		t.Errorf("codex target.branch = %q, want the captured merge base %s (develop moved to %s at send time)", sent, developTip, cardTip)
	}
	if base := reviewOutputField(res, "review_base"); !strings.Contains(base, sent) {
		t.Errorf("review_base = %q does not name the SHA codex was sent (%s)", base, sent)
	}
}

func TestCodexAudit_AdversarialBaseBranchClaimsNoReviewBase(t *testing.T) {
	repo, _ := newIntegrationBaseRepo(t, "develop")
	sess := withCodexSession(t, codexSessionScript("clean change, no findings"))

	res, err := handleCodexAudit(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"mode": codexModeAdversarial, "target": codexTargetBaseBranch, "project_root": repo,
	}}})
	if err != nil {
		t.Fatalf("handleCodexAudit: %v", err)
	}
	var turnStart map[string]any
	for _, line := range sess.sent {
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) == nil && m["method"] == codexMethodTurnStart {
			turnStart, _ = m["params"].(map[string]any)
		}
	}
	if turnStart == nil {
		t.Fatalf("no turn/start was sent; sent=%v", sess.sent)
	}
	if _, has := turnStart["target"]; has {
		t.Errorf("turn/start must carry no target; params=%v", turnStart)
	}
	if base := reviewOutputField(res, "review_base"); base != "" {
		t.Errorf("review_base = %q on an adversarial audit whose turn/start names no base; want it omitted", base)
	}
}
