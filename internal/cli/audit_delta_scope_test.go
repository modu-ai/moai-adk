package cli

// Card t1708 (defect 1) — a delta audit scoped to a previous audited SHA.
//
// codex_audit and audit_multi always reviewed `<merge-base>...HEAD`, so a delta
// re-audit re-surfaced every pre-existing finding. The explicit optional input
// `delta_base` (a commit that is an ancestor of HEAD) replaces the resolved merge
// base for the codex-native request and the Claude/GLM material; the codex
// adversarial/audit_multi leg can only be told in the prompt. Absent ⇒ unchanged.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/closure"
	"github.com/modu-ai/moai-adk/internal/closure/closuretest"
	"github.com/modu-ai/moai-adk/internal/config"
)

// deltaRepo is a MoAI-shaped repo with three commits on top of origin/main:
// A (seed, = origin/main), B (adds b.txt), C (adds c.txt). unreachable is a
// commit that is NOT an ancestor of HEAD.
type deltaRepo struct {
	root             string
	shaA, shaB, shaC string
	unreachable      string
}

func newDeltaRepo(t *testing.T) deltaRepo {
	t.Helper()
	repo := newReviewTargetRepo(t)
	seedRemoteMain(t, repo)
	d := deltaRepo{root: repo, shaA: reviewTargetGitOut(t, repo, "rev-parse", "HEAD")}
	commitFile := func(name, body string) string {
		t.Helper()
		if err := os.WriteFile(filepath.Join(repo, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		reviewTargetGit(t, repo, "add", name)
		reviewTargetGit(t, repo, "commit", "-m", "add "+name)
		return reviewTargetGitOut(t, repo, "rev-parse", "HEAD")
	}
	d.shaB = commitFile("b.txt", "canary_b_preexisting\n")
	d.shaC = commitFile("c.txt", "canary_c_delta\n")
	// A commit off to the side: reachable from no branch, so not an ancestor of HEAD.
	d.unreachable = reviewTargetGitOut(t, repo, "commit-tree", d.shaA+"^{tree}", "-p", d.shaA, "-m", "side")
	return d
}

func deltaAuditArgs(d deltaRepo, extra map[string]any) map[string]any {
	args := map[string]any{"mode": codexModeNative, "target": codexTargetBaseBranch, "project_root": d.root}
	for k, v := range extra {
		args[k] = v
	}
	return args
}

func callCodexAudit(t *testing.T, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := handleCodexAudit(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: args},
	})
	if err != nil {
		t.Fatalf("handleCodexAudit: %v", err)
	}
	return res
}

// sentTurnPrompt returns the prompt text of the turn/start the session sent.
func sentTurnPrompt(t *testing.T, sess *fakeCodexSession) string {
	t.Helper()
	for _, line := range sess.sent {
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) != nil || m["method"] != codexMethodTurnStart {
			continue
		}
		params, _ := m["params"].(map[string]any)
		input, _ := params["input"].([]any)
		if len(input) == 0 {
			continue
		}
		first, _ := input[0].(map[string]any)
		text, _ := first["text"].(string)
		return text
	}
	t.Fatalf("no turn/start was sent; sent=%v", sess.sent)
	return ""
}

func adversarialScript() []string {
	return []string{
		`{"id":1,"result":{"userAgent":"fake/1","codexHome":"/x","platformFamily":"unix","platformOs":"macos"}}`,
		`{"id":2,"result":{"thread":{"id":"tid-fake"}}}`,
		`{"id":3,"result":{"turn":{"id":"trn","status":"inProgress"}}}`,
		`{"method":"item/completed","params":{"threadId":"tid-fake","turnId":"trn","completedAtMs":1,"item":{"type":"agentMessage","id":"m1","text":"Verdict: pass"}}}`,
		`{"method":"turn/completed","params":{"threadId":"tid-fake","turn":{"id":"trn","status":"completed"}}}`,
	}
}

// ── codex native: delta_base reaches review/start ──

func TestCodexAudit_DeltaBase_NativeSendsDeltaBaseAsReviewBase(t *testing.T) {
	d := newDeltaRepo(t)
	sess := withCodexSession(t, codexSessionScript("clean change, no findings"))

	res := callCodexAudit(t, deltaAuditArgs(d, map[string]any{"delta_base": d.shaB}))

	if res.IsError {
		t.Fatalf("a valid delta_base must not be a tool error: %v", res.Content)
	}
	if got := sentTargetBranch(t, sess); got != d.shaB {
		t.Errorf("review/start target.branch = %q, want the delta base %s (not the merge base %s)", got, d.shaB, d.shaA)
	}
	rb := reviewOutputField(res, "review_base")
	if !strings.Contains(rb, d.shaB) || !strings.Contains(rb, "delta_base") {
		t.Errorf("review_base = %q, want it to name delta_base and %s", rb, d.shaB)
	}
}

func TestCodexAudit_DeltaBase_AbsentKeepsTheResolvedMergeBase(t *testing.T) {
	d := newDeltaRepo(t)
	sess := withCodexSession(t, codexSessionScript("clean change, no findings"))

	res := callCodexAudit(t, deltaAuditArgs(d, nil))

	if got := sentTargetBranch(t, sess); got != d.shaA {
		t.Errorf("target.branch = %q, want the resolved merge base %s", got, d.shaA)
	}
	if rb, want := reviewOutputField(res, "review_base"), "origin/main (merge base "+d.shaA+")"; rb != want {
		t.Errorf("review_base = %q, want %q", rb, want)
	}
}

// ── codex adversarial: the base can only be asked for in prose ──

func TestCodexAudit_DeltaBase_AdversarialPromptCarriesBaseAndSaysProseOnly(t *testing.T) {
	d := newDeltaRepo(t)
	sess := withCodexSession(t, adversarialScript())

	res := callCodexAudit(t, deltaAuditArgs(d, map[string]any{"mode": codexModeAdversarial, "delta_base": d.shaB}))

	if res.IsError {
		t.Fatalf("a valid delta_base must not be a tool error: %v", res.Content)
	}
	prompt := sentTurnPrompt(t, sess)
	if !strings.Contains(prompt, "git diff "+d.shaB+"...HEAD") {
		t.Errorf("adversarial prompt does not carry the delta scope; prompt:\n%s", prompt)
	}
	if !strings.HasPrefix(prompt, codexAdversarialReviewPrompt("")) {
		t.Errorf("the delta clause must extend the standard prompt, not replace it")
	}
	rb := reviewOutputField(res, "review_base")
	if !strings.Contains(rb, d.shaB) || !strings.Contains(rb, "prose-only") {
		t.Errorf("review_base = %q, want the result to state that the scope was requested in prose only", rb)
	}
}

func TestCodexAudit_DeltaBase_AdversarialAbsentPromptIsUnchanged(t *testing.T) {
	d := newDeltaRepo(t)
	sess := withCodexSession(t, adversarialScript())

	// REQ-MN-019: a baseBranch adversarial audit now carries the resolved base in
	// its prompt, so the byte-identical claim holds for a target with no base.
	res := callCodexAudit(t, deltaAuditArgs(d, map[string]any{"mode": codexModeAdversarial, "target": codexTargetUncommitted}))

	if got, want := sentTurnPrompt(t, sess), codexAdversarialReviewPrompt(""); got != want {
		t.Errorf("without delta_base the prompt must be byte-identical.\ngot:  %q\nwant: %q", got, want)
	}
	if rb := reviewOutputField(res, "review_base"); rb != "" {
		t.Errorf("review_base = %q without delta_base on an adversarial audit; want it omitted", rb)
	}
}

// ── invalid delta_base is a tool error, never a silent full-diff fallback ──

func TestCodexAudit_DeltaBase_InvalidIsAToolErrorNotAFallback(t *testing.T) {
	d := newDeltaRepo(t)
	cases := map[string]string{
		"not hex":             "not-a-sha",
		"option shaped":       "--output=/tmp/x",
		"hex but no commit":   "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		"not an ancestor":     d.unreachable,
		"too short":           "abc",
		"revision expression": "HEAD~1",
	}
	for name, v := range cases {
		t.Run(name, func(t *testing.T) {
			sess := withCodexSession(t, codexSessionScript("clean change, no findings"))
			res := callCodexAudit(t, deltaAuditArgs(d, map[string]any{"delta_base": v}))
			if !res.IsError {
				t.Fatalf("delta_base %q must be a tool error; got a result: %v", v, res.Content)
			}
			if len(sess.sent) != 0 {
				t.Errorf("no codex request may be sent for an invalid delta_base; sent=%v", sess.sent)
			}
		})
	}
}

// delta_base is a base-branch scope: with uncommittedChanges there is no base to
// move, and ignoring the value would silently review something else.
func TestCodexAudit_DeltaBase_WithUncommittedTargetIsAToolError(t *testing.T) {
	d := newDeltaRepo(t)
	sess := withCodexSession(t, codexSessionScript("clean change, no findings"))

	res := callCodexAudit(t, deltaAuditArgs(d, map[string]any{"target": codexTargetUncommitted, "delta_base": d.shaB}))

	if !res.IsError {
		t.Fatalf("delta_base with target uncommittedChanges must be a tool error; got %v", res.Content)
	}
	if len(sess.sent) != 0 {
		t.Errorf("no codex request may be sent; sent=%v", sess.sent)
	}
}

// ── Claude / GLM material ──

func TestCollectReviewDiffCtx_DeltaBaseScopesTheMaterial(t *testing.T) {
	d := newDeltaRepo(t)

	full, err := collectReviewDiffCtx(context.Background(), d.root, codexTargetBaseBranch)
	if err != nil {
		t.Fatalf("collectReviewDiffCtx without delta: %v", err)
	}
	if !strings.Contains(full, "canary_b_preexisting") || !strings.Contains(full, "canary_c_delta") {
		t.Fatalf("the unscoped diff must carry both commits; got:\n%s", full)
	}

	delta, err := collectReviewDiffCtx(withDeltaBase(context.Background(), d.shaB), d.root, codexTargetBaseBranch)
	if err != nil {
		t.Fatalf("collectReviewDiffCtx with delta: %v", err)
	}
	if !strings.Contains(delta, "canary_c_delta") {
		t.Errorf("the delta diff must carry the change after %s; got:\n%s", d.shaB, delta)
	}
	if strings.Contains(delta, "canary_b_preexisting") {
		t.Errorf("the delta diff must not carry the already-audited change; got:\n%s", delta)
	}
}

func TestGLMAudit_DeltaBaseReachesTheZaiRequest(t *testing.T) {
	d := newDeltaRepo(t)
	doer := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass", Summary: "ok"})}
	withGLMSeams(t, "test-key", doer)

	out := performGLMAudit(withDeltaBase(context.Background(), d.shaB), codexTargetBaseBranch, "", d.root)

	if out.Verdict == VerdictInconclusive {
		t.Fatalf("expected a real verdict; got inconclusive: %s", out.Summary)
	}
	if !strings.Contains(doer.gotBody, "canary_c_delta") || strings.Contains(doer.gotBody, "canary_b_preexisting") {
		t.Errorf("z.ai request must carry only the delta; body:\n%s", doer.gotBody)
	}
}

// ── codex leg of audit_multi: the base rides the prompt ──

func TestPerformCodexAudit_DeltaBaseRidesThePrompt(t *testing.T) {
	d := newDeltaRepo(t)
	sess := withCodexSession(t, adversarialScript())

	_ = performCodexAudit(withDeltaBase(context.Background(), d.shaB), codexTargetBaseBranch, "", d.root)

	prompt := sentTurnPrompt(t, sess)
	if !strings.Contains(prompt, "git diff "+d.shaB+"...HEAD") {
		t.Errorf("codex leg prompt does not carry the delta scope; prompt:\n%s", prompt)
	}
}

// ── audit_multi: ctx carriage, validation, no second-review record ──

func TestRunMultiAudit_DeltaBase_CarriedOnCtxToEveryLeg(t *testing.T) {
	var mu sync.Mutex
	seen := map[string]string{}
	prev := backendCall
	backendCall = func(ctx context.Context, backend, _, _, _ string) ReviewOutput {
		mu.Lock()
		seen[backend] = deltaBaseFromContext(ctx)
		mu.Unlock()
		return claudeReview("pass")
	}
	t.Cleanup(func() { backendCall = prev })

	gates := config.AuditGates{Claude: config.AuditGateRequired, Codex: config.AuditGateRequired, GLM: config.AuditGateAdvisory}
	sha := strings.Repeat("a", 40)
	_ = runMultiAudit(context.Background(), ReviewOutput{}, codexTargetBaseBranch, "", MultiAuditConfig{Gates: gates, DeltaBase: sha}, nil)
	for _, b := range []string{BackendClaude, BackendCodex, BackendGLM} {
		if seen[b] != sha {
			t.Errorf("%s leg saw delta base %q on its ctx, want %q", b, seen[b], sha)
		}
	}

	seen = map[string]string{}
	_ = runMultiAudit(context.Background(), ReviewOutput{}, codexTargetBaseBranch, "", MultiAuditConfig{Gates: gates}, nil)
	for b, v := range seen {
		if v != "" {
			t.Errorf("%s leg saw delta base %q without delta_base; want none", b, v)
		}
	}
}

func TestRunMultiAudit_DeltaBase_WritesNoSecondReviewRecord(t *testing.T) {
	f := closuretest.New(t)
	fixtureSeams(t, f)
	queueWithCards(t, f, map[string]string{"c1": closuretest.SpecID})
	recordBackends(t)

	result := runMultiAudit(context.Background(), ReviewOutput{}, codexTargetBaseBranch, "", MultiAuditConfig{
		ProjectRoot: f.CardDir,
		CardID:      "c1",
		DeltaBase:   f.Head(),
	}, nil)

	path := filepath.Join(f.EvidenceDir, closure.SecondReviewFile)
	if _, err := os.Stat(path); err == nil {
		t.Errorf("a delta call must not write %s: closure readiness would count it as full-scope review", closure.SecondReviewFile)
	}
	if result.SecondReviewRecordError != "" {
		t.Errorf("skipping the record is not an error; second_review_record_error = %q", result.SecondReviewRecordError)
	}
	if !strings.Contains(result.ResidualRiskNote, "delta_base") {
		t.Errorf("residual_risk_note = %q, want the delta scope (and the codex prose-only limit) named", result.ResidualRiskNote)
	}
}

func TestAuditMulti_DeltaBase_HandlerValidatesAndCarries(t *testing.T) {
	d := newDeltaRepo(t)
	var mu sync.Mutex
	var calls int
	var seen []string
	prev := backendCall
	backendCall = func(ctx context.Context, backend, _, _, _ string) ReviewOutput {
		mu.Lock()
		calls++
		seen = append(seen, deltaBaseFromContext(ctx))
		mu.Unlock()
		return claudeReview("pass")
	}
	t.Cleanup(func() { backendCall = prev })

	call := func(extra map[string]any) *mcp.CallToolResult {
		args := map[string]any{
			"target": codexTargetBaseBranch, "project_root": d.root,
			"gates": map[string]any{"claude": "required", "codex": "required", "glm": "off"},
		}
		for k, v := range extra {
			args[k] = v
		}
		res, err := handleAuditMulti(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}})
		if err != nil {
			t.Fatalf("handleAuditMulti: %v", err)
		}
		return res
	}

	if res := call(map[string]any{"delta_base": d.shaB}); res.IsError {
		t.Fatalf("a valid delta_base must not be a tool error: %v", res.Content)
	}
	if calls == 0 {
		t.Fatal("no backend was called")
	}
	for _, v := range seen {
		if v != d.shaB {
			t.Errorf("a leg saw delta base %q, want the validated full SHA %s", v, d.shaB)
		}
	}

	calls = 0
	for _, bad := range []string{"not-a-sha", d.unreachable, "HEAD~1"} {
		if res := call(map[string]any{"delta_base": bad}); !res.IsError {
			t.Errorf("delta_base %q must be a tool error", bad)
		}
	}
	if calls != 0 {
		t.Errorf("no backend may run for an invalid delta_base; %d calls", calls)
	}
}
