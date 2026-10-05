package cli

// Card t1426 review follow-up (independent codex review of f82a078bb, two P2):
//
//  1. Branch name and merge base must come from ONE resolution, so codex (which
//     is sent a name) and GLM (which is sent a diff from a merge base) never
//     disagree. A configured base that exists but shares no history with HEAD
//     must make BOTH fall through to the same next step.
//  2. review_base must name the base the reviewed material was collected
//     against — captured before the backend is called — not a value recomputed
//     after the response, when the base ref may have moved.

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// newUnrelatedBaseRepo: main (= origin/main = origin/HEAD) carries the config
// naming `develop`; `develop` is an ORPHAN branch sharing no history with main;
// the card branch is cut from main with one change.
func newUnrelatedBaseRepo(t *testing.T) string {
	t.Helper()
	repo := newReviewTargetRepo(t)
	reviewTargetGit(t, repo, "branch", "-M", "main")
	writeWorktreeBaseBranchConfig(t, repo, "develop")
	reviewTargetGit(t, repo, "add", ".moai")
	reviewTargetGit(t, repo, "commit", "-m", "config")
	seedRemoteMain(t, repo)

	reviewTargetGit(t, repo, "checkout", "--orphan", "develop")
	reviewTargetGit(t, repo, "rm", "-rf", "--cached", ".")
	writeBaseFixtureFile(t, repo, "develop-only.txt", "unrelated history\n")
	reviewTargetGit(t, repo, "add", "develop-only.txt")
	reviewTargetGit(t, repo, "commit", "-m", "orphan develop")

	reviewTargetGit(t, repo, "checkout", "-f", "main")
	reviewTargetGit(t, repo, "clean", "-fdq")
	reviewTargetGit(t, repo, "checkout", "-b", "WT-card-fixture")
	writeBaseFixtureFile(t, repo, "card.txt", "the card's own change\n")
	reviewTargetGit(t, repo, "add", "card.txt")
	reviewTargetGit(t, repo, "commit", "-m", "card change")
	return repo
}

func TestAuditBaseBranch_UnrelatedConfiguredBaseFallsThroughForBothBackends(t *testing.T) {
	repo := newUnrelatedBaseRepo(t)
	mainSHA := reviewTargetGitOut(t, repo, "rev-parse", "main")

	sess, res := runNativeAudit(t, repo, codexTargetBaseBranch)
	codexBranch := sentTargetBranch(t, sess)
	glmBase, err := resolveReviewMergeBase(repo)
	if err != nil {
		t.Fatalf("resolveReviewMergeBase: %v", err)
	}
	if codexBranch != mainSHA || glmBase != mainSHA {
		t.Errorf("backends disagree or did not fall through: codex target.branch = %q, GLM merge base = %s; want both on main's merge base %s", codexBranch, glmBase, mainSHA)
	}
	if base := reviewOutputField(res, "review_base"); !strings.Contains(base, "main") || !strings.Contains(base, mainSHA) {
		t.Errorf("review_base = %q, want main with merge base %s", base, mainSHA)
	}
}

// movingRefGLMDoer moves `develop` to the card tip while the "model" is
// thinking, then answers pass.
type movingRefGLMDoer struct {
	t      *testing.T
	repo   string
	toSHA  string
	answer string
}

func (d movingRefGLMDoer) Do(req *http.Request) (*http.Response, error) {
	moveRef(d.t, d.repo, "refs/heads/develop", d.toSHA)
	return (&stubGLMDoer{body: d.answer}).Do(req)
}

func moveRef(t *testing.T, repo, ref, sha string) {
	t.Helper()
	if out, err := exec.Command("git", "-C", repo, "update-ref", ref, sha).CombinedOutput(); err != nil {
		t.Errorf("update-ref %s: %v %s", ref, err, out)
	}
}

func TestGLMAudit_ReviewBaseIsTheCollectedBase(t *testing.T) {
	repo, developTip := newIntegrationBaseRepo(t, "develop")
	cardTip := reviewTargetGitOut(t, repo, "rev-parse", "HEAD")
	answer := glmMessagesResp(t, ReviewOutput{Verdict: "pass", Summary: "ok", Findings: []Finding{}, NextSteps: []string{}})
	withGLMSeams(t, "test-key", movingRefGLMDoer{t: t, repo: repo, toSHA: cardTip, answer: answer})

	res, err := handleGLMAudit(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"project_root": repo, "target": codexTargetBaseBranch,
	}}})
	if err != nil {
		t.Fatalf("handleGLMAudit: %v", err)
	}
	base := reviewOutputField(res, "review_base")
	if !strings.Contains(base, developTip) {
		t.Errorf("review_base = %q, want the collected merge base %s (the ref moved to %s during the call)", base, developTip, cardTip)
	}
}

// movingRefCodexSession moves `develop` when review/start is sent.
type movingRefCodexSession struct {
	fakeCodexSession
	t     *testing.T
	repo  string
	toSHA string
}

func (s *movingRefCodexSession) start(context.Context, string, []string) (codexConn, error) {
	return &movingRefCodexConn{fakeCodexConn: fakeCodexConn{lines: s.lines, sent: &s.sent}, s: s}, nil
}

type movingRefCodexConn struct {
	fakeCodexConn
	s *movingRefCodexSession
}

func (c *movingRefCodexConn) send(line string) error {
	if strings.Contains(line, codexMethodReviewStart) {
		moveRef(c.s.t, c.s.repo, "refs/heads/develop", c.s.toSHA)
	}
	return c.fakeCodexConn.send(line)
}

func TestCodexAudit_ReviewBaseIsTheBaseResolvedBeforeTheCall(t *testing.T) {
	repo, developTip := newIntegrationBaseRepo(t, "develop")
	cardTip := reviewTargetGitOut(t, repo, "rev-parse", "HEAD")
	withCodexSession(t, nil) // install LookPath/runner seams, then swap the session
	codexSession = &movingRefCodexSession{
		fakeCodexSession: fakeCodexSession{lines: codexSessionScript("clean change, no findings")},
		t:                t, repo: repo, toSHA: cardTip,
	}

	res, err := handleCodexAudit(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"mode": codexModeNative, "target": codexTargetBaseBranch, "project_root": repo,
	}}})
	if err != nil {
		t.Fatalf("handleCodexAudit: %v", err)
	}
	base := reviewOutputField(res, "review_base")
	if !strings.Contains(base, developTip) {
		t.Errorf("review_base = %q, want the merge base resolved before the call %s (the ref moved to %s during the call)", base, developTip, cardTip)
	}
}
