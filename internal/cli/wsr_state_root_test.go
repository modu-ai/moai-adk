package cli

// SPEC-WORKTREE-STATE-ROOT-001 — state and catalogue roots for a config-orphaned
// linked worktree. One test per acceptance criterion, named TestWSR<nnn>.
// Fixture F (untracked-.moai repository P + linked worktree W) is built by
// newWSRFixture on top of newUntrackedFixture. Every test here mutates package
// seams or the process environment, so none runs in parallel.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/verify"
)

// wsrWorkflowF is fixture F's primary workflow config: the codex audit gate is
// required and both Stop review gates are opted in.
const wsrWorkflowF = "workflow:\n" +
	"  audit:\n    gates:\n      codex: required\n" +
	"  multi:\n    review_gate:\n      enabled: true\n" +
	"  codex:\n    review_gate:\n      enabled: true\n"

// wsrSpecBody is a parseable SPEC frontmatter with no era field, so spec_audit
// and spec_drift emit an EraAutoDetected finding for it.
func wsrSpecBody(id, status string) string {
	return "---\nid: " + id + "\ntitle: \"fixture\"\nversion: \"0.1.0\"\nstatus: " + status + "\n" +
		"created: 2026-01-01\nupdated: 2026-01-01\nauthor: t\npriority: P3\nphase: \"v3.0.0\"\n" +
		"module: \"internal/x\"\nlifecycle: spec-anchored\ntags: \"x\"\n---\n\n# " + id + "\n"
}

// newWSRFixture builds fixture F: P carries workflowYAML and exactly one SPEC,
// SPEC-PRI-001; W is a linked worktree with no .moai.
func newWSRFixture(t *testing.T, workflowYAML string) untrackedFixture {
	t.Helper()
	fx := newUntrackedFixture(t, workflowYAML)
	if err := os.RemoveAll(filepath.Join(fx.P, ".moai", "specs", "SPEC-WTFIX-001")); err != nil {
		t.Fatal(err)
	}
	wtWriteFile(t, filepath.Join(fx.P, ".moai", "specs", "SPEC-PRI-001", "spec.md"), wsrSpecBody("SPEC-PRI-001", "draft"))
	return fx
}

// wsrAddWorktree adds a further linked worktree of fx.P (fixture F2).
func wsrAddWorktree(t *testing.T, fx untrackedFixture, name string) string {
	t.Helper()
	w := filepath.Join(filepath.Dir(fx.P), name)
	wtGit(t, fx.env, fx.P, "worktree", "add", "-q", "-b", strings.ToLower(name), w)
	return w
}

// wsrNoidAdminHead makes W's primary unidentifiable by deleting the admin HEAD
// (F-noid variant b); W gets an empty .moai so the validator still accepts it.
func wsrNoidAdminHead(t *testing.T, fx untrackedFixture) {
	t.Helper()
	mustMkdir(t, filepath.Join(fx.W, ".moai"))
	if err := os.Remove(filepath.Join(fx.P, ".git", "worktrees", filepath.Base(fx.W), "HEAD")); err != nil {
		t.Fatal(err)
	}
}

// wsrTickClock makes auditreceipt.Now strictly increasing, so a receipt minted
// after a start marker is never stamped with the same instant.
func wsrTickClock(t *testing.T) {
	t.Helper()
	base := time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC)
	n := 0
	prev := auditreceipt.Now
	auditreceipt.Now = func() time.Time { n++; return base.Add(time.Duration(n) * time.Second) }
	t.Cleanup(func() { auditreceipt.Now = prev })
}

// wsrStubCodexPass makes codex_audit return verdict pass without a binary.
func wsrStubCodexPass(t *testing.T) {
	t.Helper()
	prevLook, prevRPC := codexLookPath, codexReviewRPC
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	codexReviewRPC = func(context.Context, string, string, map[string]any) (ReviewOutput, error) {
		return ReviewOutput{Verdict: "pass", Summary: "stub pass", Findings: []Finding{}, NextSteps: []string{}}, nil
	}
	t.Cleanup(func() { codexLookPath, codexReviewRPC = prevLook, prevRPC })
}

// wsrCodexAudit calls codex_audit with project_root = root and decodes it.
func wsrCodexAudit(t *testing.T, root string) (ReviewOutput, string) {
	t.Helper()
	res, err := handleCodexAudit(context.Background(), newToolRequest(map[string]any{
		"project_root": root, "target": "uncommittedChanges",
	}))
	if err != nil {
		t.Fatalf("codex_audit hard error: %v", err)
	}
	text := resultTextOf(res)
	if res.IsError {
		t.Fatalf("codex_audit tool error on %s: %s", root, text)
	}
	var out ReviewOutput
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatalf("decode codex_audit: %v (%s)", err, text)
	}
	return out, text
}

// wsrAuditMulti calls audit_multi with project_root = root, session sid, and
// backend verdicts fixed by verdicts (claude + codex participate, glm off).
func wsrAuditMulti(t *testing.T, root, sid string, verdicts map[string]string) (ConvergenceResult, string) {
	t.Helper()
	t.Setenv(config.EnvMoaiLaunchProvider, "")
	withBackendCall(t, gateBlockCaller(verdicts))
	res, err := handleAuditMulti(context.Background(), newToolRequest(map[string]any{
		"project_root": root, "session_id": sid, "gates": map[string]any{"glm": "off"},
	}))
	if err != nil {
		t.Fatalf("audit_multi hard error: %v", err)
	}
	text := resultTextOf(res)
	if res.IsError {
		t.Fatalf("audit_multi tool error on %s: %s", root, text)
	}
	var r ConvergenceResult
	if err := json.Unmarshal([]byte(text), &r); err != nil {
		t.Fatalf("decode audit_multi: %v (%s)", err, text)
	}
	return r, text
}

// wsrJSONFiles lists the *.json files directly under dir (none when absent).
func wsrJSONFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read %s: %v", dir, err)
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(out)
	return out
}

// wsrReadMap decodes one JSON object file.
func wsrReadMap(t *testing.T, path string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return m
}

// wsrAssertNoMoai asserts the worktree has no .moai path at all.
func wsrAssertNoMoai(t *testing.T, w string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(w, ".moai")); !os.IsNotExist(err) {
		t.Errorf("%s/.moai must not exist (stat err=%v)", w, err)
	}
}

// wsrAssertEmptyMoai asserts the worktree's .moai is an empty directory.
func wsrAssertEmptyMoai(t *testing.T, w string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(w, ".moai"))
	if err != nil || len(entries) != 0 {
		t.Errorf("%s/.moai must still be an empty directory (entries=%d err=%v)", w, len(entries), err)
	}
}

// wsrConvergenceFiles returns the convergence result files for session sid in
// root's .moai/state/audit-multi (S.json and any tree-qualified S--*.json).
func wsrConvergenceFiles(t *testing.T, root, sid string) []string {
	t.Helper()
	var out []string
	for _, f := range wsrJSONFiles(t, filepath.Join(root, ".moai", "state", "audit-multi")) {
		base := filepath.Base(f)
		if base == sid+".json" || strings.HasPrefix(base, sid+"--") {
			out = append(out, f)
		}
	}
	return out
}

// wsrDetector replaces the review-gate change detector with a recorder that
// always reports a reviewable change.
func wsrDetector(t *testing.T) *[]string {
	t.Helper()
	calls := &[]string{}
	prev := reviewGateChangeDetector
	reviewGateChangeDetector = func(dir string) bool { *calls = append(*calls, dir); return true }
	t.Cleanup(func() { reviewGateChangeDetector = prev })
	return calls
}

// wsrRunGate runs one Stop review gate RunE on payload and returns its stdout.
func wsrRunGate(t *testing.T, multi bool, payload map[string]any) string {
	t.Helper()
	blob, _ := json.Marshal(payload)
	cmd, out, _ := newCodexGateCmd(string(blob))
	if multi {
		cmd, out, _ = newGateCmd(string(blob))
	}
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("gate RunE: %v", err)
	}
	return strings.TrimSpace(out.String())
}

// AC-WSR-001: codex_audit on W files its receipt in P's store with tree
// identity W and creates nothing under W.
func TestWSR001_ReceiptInPrimaryStoreWithWorktreeIdentity(t *testing.T) {
	fx := newWSRFixture(t, wsrWorkflowF)
	wsrStubCodexPass(t)
	wsrCodexAudit(t, fx.W)

	files := wsrJSONFiles(t, filepath.Join(fx.P, ".moai", "state", "audit-receipts", "receipts"))
	if len(files) != 1 {
		t.Fatalf("want exactly one receipt under P/.moai/state/audit-receipts/receipts, got %d: %v", len(files), files)
	}
	if got := wsrReadMap(t, files[0])["tree_root"]; got != fx.W {
		t.Errorf("receipt tree_root = %v, want canonical W %s", got, fx.W)
	}
	wsrAssertNoMoai(t, fx.W)
}

// AC-WSR-002: audit_multi on W persists its result in P's store with tree
// identity W; results of W and W2 for one session coexist; the multi gate
// blocks for a W2 input (2a) and for a P input (2b).
func TestWSR002_ConvergenceStoreIdentityAndCoexistence(t *testing.T) {
	const sid = "S"
	t.Run("cell1", func(t *testing.T) {
		fx := newWSRFixture(t, wsrWorkflowF)
		r, _ := wsrAuditMulti(t, fx.W, sid, map[string]string{BackendClaude: "fail", BackendCodex: "pass"})
		if r.OverallVerdict != overallVerdictFail {
			t.Fatalf("premise: stubs determine overall fail, got %q", r.OverallVerdict)
		}
		files := wsrConvergenceFiles(t, fx.P, sid)
		if len(files) != 1 {
			t.Fatalf("want exactly one convergence result for %s under P's store, got %d: %v", sid, len(files), files)
		}
		m := wsrReadMap(t, files[0])
		if m["overall_verdict"] != "fail" || m["tree_root"] != fx.W {
			t.Errorf("want overall_verdict fail and tree_root W, got overall=%v tree_root=%v", m["overall_verdict"], m["tree_root"])
		}
		wsrAssertNoMoai(t, fx.W)
	})

	// Cell 2 — fixture F2. Setup runs once; each predicate is its own subtest
	// so cell 2b's outcome is observable on its own (-run TestWSR002/cell2b).
	fx := newWSRFixture(t, wsrWorkflowF)
	w2 := wsrAddWorktree(t, fx, "W2")
	wsrAuditMulti(t, fx.W, sid, map[string]string{BackendClaude: "fail", BackendCodex: "pass"})
	wsrAuditMulti(t, w2, sid, map[string]string{BackendClaude: "pass", BackendCodex: "pass"})
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	calls := wsrDetector(t)

	t.Run("cell2-store", func(t *testing.T) {
		got := map[string]any{}
		for _, f := range wsrConvergenceFiles(t, fx.P, sid) {
			m := wsrReadMap(t, f)
			tr, _ := m["tree_root"].(string)
			got[tr] = m["overall_verdict"]
		}
		if len(got) != 2 || got[fx.W] != "fail" || got[w2] != "pass" {
			t.Errorf("want two results for %s in P's store (W fail, W2 pass), got %v", sid, got)
		}
	})
	t.Run("cell2a", func(t *testing.T) {
		out := wsrRunGate(t, true, map[string]any{"session_id": sid, "cwd": w2})
		if !strings.Contains(out, `"decision":"block"`) {
			t.Errorf("2a (cwd=W2): want block, got %s (detector calls %v)", out, *calls)
		}
	})
	t.Run("cell2b", func(t *testing.T) {
		out := wsrRunGate(t, true, map[string]any{"session_id": sid, "cwd": fx.P})
		if !strings.Contains(out, `"decision":"block"`) {
			t.Errorf("2b (cwd=P): want block on the fail audit_multi wrote for W, got %s (detector calls %v)", out, *calls)
		}
		t.Logf("2b observed: output=%s", out)
	})
}

// AC-WSR-006: the Stop review gates' input-to-root matrix. P's own store holds
// a fail for S; W's own store holds a decoy pass for S.
func TestWSR006_ReviewGateRootMatrix(t *testing.T) {
	const sid = "S"
	fx := newWSRFixture(t, wsrWorkflowF)
	failRec := `{"per_backend_verdicts":[],"overall_verdict":"fail","disagreement_flag":null,"participant_count":0,"residual_risk_note":"P fail record","fail_open_backends":[]}`
	decoy := `{"per_backend_verdicts":[],"overall_verdict":"pass","disagreement_flag":null,"participant_count":0,"residual_risk_note":"W decoy","fail_open_backends":[]}`
	wtWriteFile(t, filepath.Join(fx.P, ".moai", "state", "audit-multi", sid+".json"), failRec)
	wtWriteFile(t, filepath.Join(fx.W, ".moai", "state", "audit-multi", sid+".json"), decoy)
	withCodexLookPath(t, func(string) (string, error) { return "", os.ErrNotExist })

	val := map[string]string{"-": "", "W": fx.W, "P": fx.P}
	rows := []struct{ pd, cwd, env string }{
		{"-", "-", "-"}, {"-", "-", "W"}, {"-", "-", "P"},
		{"-", "W", "-"}, {"-", "W", "W"}, {"-", "W", "P"},
		{"-", "P", "-"}, {"-", "P", "W"}, {"-", "P", "P"},
		{"W", "-", "-"}, {"W", "-", "W"}, {"W", "-", "P"},
		{"W", "W", "-"}, {"W", "W", "W"}, {"W", "W", "P"},
		{"W", "P", "-"}, {"W", "P", "W"}, {"W", "P", "P"},
		{"P", "-", "-"}, {"P", "-", "W"}, {"P", "-", "P"},
		{"P", "W", "-"}, {"P", "W", "W"}, {"P", "W", "P"},
		{"P", "P", "-"}, {"P", "P", "W"}, {"P", "P", "P"},
	}
	type outcome struct {
		codexCalls, multiCalls []string
		multiBlock             bool
	}
	observe := func(t *testing.T, pd, cwd, env string) outcome {
		t.Helper()
		t.Setenv("CLAUDE_PROJECT_DIR", env)
		payload := map[string]any{"session_id": sid}
		if pd != "" {
			payload["project_dir"] = pd
		}
		if cwd != "" {
			payload["cwd"] = cwd
		}
		calls := wsrDetector(t)
		wsrRunGate(t, false, payload)
		codexCalls := append([]string{}, *calls...)
		*calls = (*calls)[:0]
		out := wsrRunGate(t, true, payload)
		return outcome{codexCalls: codexCalls, multiCalls: append([]string{}, *calls...), multiBlock: strings.Contains(out, `"decision":"block"`)}
	}
	classify := func(o outcome, r string) string {
		switch {
		case len(o.codexCalls) == 0 && len(o.multiCalls) == 0 && !o.multiBlock:
			if r == "" {
				return "N"
			}
			return "Wt"
		case len(o.codexCalls) == 1 && len(o.multiCalls) == 1 && o.codexCalls[0] == r && o.multiCalls[0] == r && o.multiBlock:
			if r == fx.P {
				return "Pp"
			}
			return "Wr"
		case len(o.multiCalls) == 1 && !o.multiBlock:
			return "decoy-read"
		}
		return "other"
	}
	for i, row := range rows {
		r := val[row.pd]
		if r == "" {
			r = val[row.cwd]
		}
		if r == "" {
			r = val[row.env]
		}
		want := map[string]string{"": "N", fx.W: "Wr", fx.P: "Pp"}[r]
		o := observe(t, val[row.pd], val[row.cwd], val[row.env])
		got := classify(o, r)
		t.Logf("row %2d pd=%s cwd=%s env=%s R=%s observed=%s (codex calls=%v multi calls=%v multi block=%v) required=%s",
			i+1, row.pd, row.cwd, row.env, map[string]string{"": "empty", fx.W: "W", fx.P: "P"}[r], got, o.codexCalls, o.multiCalls, o.multiBlock, want)
		if got != want {
			t.Errorf("row %d: observed %s, required %s", i+1, got, want)
		}
	}

	t.Run("F-noid row 4", func(t *testing.T) {
		for _, variant := range []string{"admin HEAD deleted", "no git"} {
			nfx := newWSRFixture(t, wsrWorkflowF)
			if variant == "no git" {
				mustMkdir(t, filepath.Join(nfx.W, ".moai"))
				noGitPath(t)
			} else {
				wsrNoidAdminHead(t, nfx)
			}
			t.Setenv("CLAUDE_PROJECT_DIR", "")
			calls := wsrDetector(t)
			codexOut := wsrRunGate(t, false, map[string]any{"session_id": sid, "cwd": nfx.W})
			multiOut := wsrRunGate(t, true, map[string]any{"session_id": sid, "cwd": nfx.W})
			if len(*calls) != 0 || codexOut != "{}" || multiOut != "{}" {
				t.Errorf("%s: want N′ (both {} and no detector call), got codex=%s multi=%s calls=%v", variant, codexOut, multiOut, *calls)
			}
		}
	})
}

// wsrNoidVariants runs fn once per F-noid variant: (a) git absent from PATH,
// (b) W's admin HEAD deleted. W carries an empty .moai in both.
func wsrNoidVariants(t *testing.T, workflowYAML string, fn func(t *testing.T, fx untrackedFixture)) {
	t.Helper()
	t.Run("admin HEAD deleted", func(t *testing.T) {
		fx := newWSRFixture(t, workflowYAML)
		wsrNoidAdminHead(t, fx)
		fn(t, fx)
	})
	t.Run("no git", func(t *testing.T) {
		fx := newWSRFixture(t, workflowYAML)
		mustMkdir(t, filepath.Join(fx.W, ".moai"))
		noGitPath(t)
		fn(t, fx)
	})
}

// wsrToolText calls a handler and returns (isError, text).
func wsrToolText(t *testing.T, handler func(context.Context, mcpReq) (*mcpRes, error), args map[string]any) (bool, string) {
	t.Helper()
	res, err := handler(context.Background(), newToolRequest(args))
	if err != nil {
		t.Fatalf("handler hard error: %v", err)
	}
	blob, _ := json.Marshal(res) // text and structured content both
	return res.IsError, string(blob)
}

// AC-WSR-003: with W's primary unidentifiable no state write falls back to
// another root: the verify tools error, the audits keep their verdicts and
// carry a notice naming the skipped write and cause, and nothing is written.
func TestWSR003_PrimaryUnidentifiedNoFallback(t *testing.T) {
	wsrNoidVariants(t, wsrWorkflowF, func(t *testing.T, fx untrackedFixture) {
		for name, h := range map[string]func(context.Context, mcpReq) (*mcpRes, error){
			"verify_snapshot": handleVerifySnapshot, "verify_trend": handleVerifyTrend,
		} {
			isErr, text := wsrToolText(t, h, map[string]any{"project_root": fx.W, "key": "k1", "command": "c"})
			if !isErr || !strings.Contains(text, "primary checkout") || !strings.Contains(text, "could not be identified") {
				t.Errorf("%s: want a tool error naming the unidentifiable primary checkout, got isError=%v %s", name, isErr, text)
			}
		}
		wsrStubCodexPass(t)
		out, _ := wsrCodexAudit(t, fx.W)
		if out.Verdict != "pass" {
			t.Errorf("codex_audit verdict = %q, want the seam's pass", out.Verdict)
		}
		r, text := wsrAuditMulti(t, fx.W, "S", map[string]string{BackendClaude: "fail", BackendCodex: "pass"})
		if r.OverallVerdict != overallVerdictFail {
			t.Errorf("audit_multi overall = %q, want fail", r.OverallVerdict)
		}
		for name, notice := range map[string]string{"codex_audit": out.StateNotice, "audit_multi": r.StateNotice} {
			if !strings.Contains(notice, "receipt") || !strings.Contains(notice, "could not be identified") {
				t.Errorf("%s: want a notice naming the skipped receipt write and the cause, got %q", name, notice)
			}
		}
		if !strings.Contains(r.StateNotice, "convergence") {
			t.Errorf("audit_multi notice must name the skipped convergence write, got %q (%s)", r.StateNotice, text)
		}
		for _, root := range []string{fx.P, fx.W} {
			if n := len(wsrJSONFiles(t, filepath.Join(root, ".moai", "state", "audit-receipts", "receipts"))); n != 0 {
				t.Errorf("no receipt may exist under %s, found %d", root, n)
			}
			if n := len(wsrConvergenceFiles(t, root, "S")); n != 0 {
				t.Errorf("no convergence result may exist under %s, found %d", root, n)
			}
		}
		wsrAssertEmptyMoai(t, fx.W)
	})
}

// AC-WSR-004: a root that is not config-orphaned — P, a tracked-.moai
// worktree WT, a bare non-git B — keeps every path it uses today, with and
// without git on PATH.
func TestWSR004_OtherRootsUnchanged(t *testing.T) {
	for _, withGit := range []bool{true, false} {
		fx := newWSRFixture(t, wsrWorkflowF)
		_, wt := newTrackedFixture(t)
		b := wtCanonTempDir(t)
		mustMkdir(t, filepath.Join(b, ".moai"))
		wsrStubCodexPass(t)
		if !withGit {
			noGitPath(t)
		}
		for _, root := range []string{fx.P, wt, b} {
			out, _ := wsrCodexAudit(t, root)
			rcpts := wsrJSONFiles(t, filepath.Join(root, ".moai", "state", "audit-receipts", "receipts"))
			if len(rcpts) != 1 {
				t.Errorf("git=%v %s: want the receipt under its own store, got %v", withGit, root, rcpts)
			}
			_ = out
			wsrAuditMulti(t, root, "S", map[string]string{BackendClaude: "pass", BackendCodex: "pass"})
			if _, err := os.Stat(filepath.Join(root, ".moai", "state", "audit-multi", "S.json")); err != nil {
				t.Errorf("git=%v %s: convergence result not at the base path: %v", withGit, root, err)
			}
			if isErr, text := wsrToolText(t, handleVerifySnapshot, map[string]any{"project_root": root, "key": "h:d", "command": "c"}); isErr {
				t.Errorf("git=%v %s: verify_snapshot record failed: %s", withGit, root, text)
			} else if _, err := os.Stat(verify.SnapshotPath(root, "h:d")); err != nil {
				t.Errorf("git=%v %s: snapshot not at the base path: %v", withGit, root, err)
			}
		}
	}
}

// AC-WSR-005: verify_snapshot/verify_trend and `moai verify record|check` on W
// all use P's store, and agree.
func TestWSR005_VerifyToolsAndCLIAgree(t *testing.T) {
	fx := newWSRFixture(t, wtWorkflowNoGate)
	wtWriteFile(t, filepath.Join(fx.W, "tracked.txt"), "uncommitted change in W\n")
	key, err := verify.Key(context.Background(), fx.W)
	if err != nil {
		t.Fatal(err)
	}
	if isErr, text := wsrToolText(t, handleVerifySnapshot, map[string]any{"project_root": fx.W, "key": key, "command": "c1"}); isErr {
		t.Fatalf("verify_snapshot record: %s", text)
	}
	run := func(args ...string) (string, error) {
		cmd := newVerifyCmd()
		var out strings.Builder
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs(append([]string{"--project-root", fx.W}, args...))
		err := cmd.Execute()
		return out.String(), err
	}
	if out, err := run("record", "--check-id", "c2", "--command", "c2"); err != nil {
		t.Fatalf("moai verify record: %v (%s)", err, out)
	}
	snaps := wsrJSONFiles(t, filepath.Join(fx.P, ".moai", "state", "verify", "snapshots"))
	if len(snaps) != 1 {
		t.Errorf("want exactly one snapshot under P, got %v", snaps)
	}
	for name, h := range map[string]func(context.Context, mcpReq) (*mcpRes, error){
		"verify_snapshot": handleVerifySnapshot, "verify_trend": handleVerifyTrend,
	} {
		_, text := wsrToolText(t, h, map[string]any{"project_root": fx.W, "key": key})
		if !strings.Contains(text, `"c1"`) && !strings.Contains(text, "c1") || !strings.Contains(text, "c2") {
			t.Errorf("%s load on W must return c1 and c2, got %s", name, text)
		}
	}
	if out, err := run("check"); err != nil || !strings.Contains(out, `"fresh": true`) {
		t.Errorf("moai verify check on W must report the snapshot present and fresh, got err=%v %s", err, out)
	}
	wsrAssertNoMoai(t, fx.W)
}

// AC-WSR-010: on W the catalogue tools answer over the union of W's and P's
// .moai/specs, and every record and finding names its source.
func TestWSR010_UnionCatalogue(t *testing.T) {
	fx := newWSRFixture(t, wtWorkflowNoGate)
	wtWriteFile(t, filepath.Join(fx.W, ".moai", "specs", "SPEC-WTL-001", "spec.md"), wsrSpecBody("SPEC-WTL-001", "draft"))
	wantSource := map[string]string{"SPEC-WTL-001": "worktree", "SPEC-PRI-001": "primary"}

	var progress struct {
		Count int              `json:"count"`
		Specs []map[string]any `json:"specs"`
	}
	decodeTool(t, handleSpecProgress, map[string]any{"project_root": fx.W}, &progress)
	if progress.Count != 2 {
		t.Errorf("spec_progress count = %d, want 2", progress.Count)
	}
	seen := map[string]string{}
	for _, rec := range progress.Specs {
		seen[wsrSpecIDOfPath(rec["Path"])], _ = rec["source"].(string)
	}
	for id, src := range wantSource {
		if seen[id] != src {
			t.Errorf("spec_progress: %s source = %q, want %q (records %v)", id, seen[id], src, seen)
		}
	}

	for name, handler := range map[string]func(context.Context, mcpReq) (*mcpRes, error){
		"spec_drift": handleSpecDrift, "spec_audit": handleSpecAudit,
	} {
		var body struct {
			TotalSpecs int              `json:"total_specs"`
			Findings   []map[string]any `json:"drift_findings"`
		}
		decodeTool(t, handler, map[string]any{"project_root": fx.W}, &body)
		if body.TotalSpecs != 2 {
			t.Errorf("%s total_specs = %d, want 2", name, body.TotalSpecs)
		}
		ids := map[string]bool{}
		for _, f := range body.Findings {
			ids[f["spec_id"].(string)] = true
		}
		if !ids["SPEC-WTL-001"] || !ids["SPEC-PRI-001"] {
			t.Errorf("%s premise: want findings for both SPEC-WTL-001 and SPEC-PRI-001, got %v", name, ids)
			continue
		}
		for _, f := range body.Findings {
			id := f["spec_id"].(string)
			if want, ok := wantSource[id]; ok && f["source"] != want {
				t.Errorf("%s finding %s/%v source = %v, want %s", name, id, f["finding_type"], f["source"], want)
			}
		}
	}
}

type (
	mcpReq = mcp.CallToolRequest
	mcpRes = mcp.CallToolResult
)

// decodeTool calls a structured-result tool handler and decodes its
// structured content into v, failing on a tool error.
func decodeTool(t *testing.T, handler func(context.Context, mcpReq) (*mcpRes, error), args map[string]any, v any) {
	t.Helper()
	res, err := handler(context.Background(), newToolRequest(args))
	if err != nil {
		t.Fatalf("handler hard error: %v", err)
	}
	if res.IsError {
		t.Fatalf("handler returned a tool error: %s", resultTextOf(res))
	}
	blob, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(blob, v); err != nil {
		t.Fatalf("decode: %v (%s)", err, blob)
	}
}

// AC-WSR-011: a SPEC ID present in both catalogues is reported once, from the
// worktree, and the shadowed primary copy is named.
func TestWSR011_SameSpecInBoth(t *testing.T) {
	fx := newWSRFixture(t, wtWorkflowNoGate)
	wtWriteFile(t, filepath.Join(fx.W, ".moai", "specs", "SPEC-DUP-001", "spec.md"), wsrSpecBody("SPEC-DUP-001", "draft"))
	wtWriteFile(t, filepath.Join(fx.P, ".moai", "specs", "SPEC-DUP-001", "spec.md"), wsrSpecBody("SPEC-DUP-001", "completed"))
	namesShadow := func(body map[string]any) bool {
		list, _ := body["shadowed"].([]any)
		for _, s := range list {
			if m, ok := s.(map[string]any); ok && m["spec_id"] == "SPEC-DUP-001" && m["source"] == "primary" {
				return true
			}
		}
		return false
	}

	var progress map[string]any
	decodeTool(t, handleSpecProgress, map[string]any{"project_root": fx.W}, &progress)
	dup := 0
	for _, s := range progress["specs"].([]any) {
		rec := s.(map[string]any)
		if wsrSpecIDOfPath(rec["Path"]) != "SPEC-DUP-001" {
			continue
		}
		dup++
		fm, _ := rec["Frontmatter"].(map[string]any)
		if fm["Status"] != "draft" || rec["source"] != "worktree" {
			t.Errorf("spec_progress SPEC-DUP-001: want status draft source worktree, got status=%v source=%v", fm["Status"], rec["source"])
		}
	}
	if dup != 1 || progress["count"] != float64(2) {
		t.Errorf("spec_progress: want SPEC-DUP-001 once and count 2, got %d and count %v", dup, progress["count"])
	}
	if !namesShadow(progress) {
		t.Errorf("spec_progress must name the shadowed primary copy, got %v", progress["shadowed"])
	}
	for name, h := range map[string]func(context.Context, mcpReq) (*mcpRes, error){
		"spec_drift": handleSpecDrift, "spec_audit": handleSpecAudit,
	} {
		var body map[string]any
		decodeTool(t, h, map[string]any{"project_root": fx.W}, &body)
		if body["total_specs"] != float64(2) {
			t.Errorf("%s total_specs = %v, want 2", name, body["total_specs"])
		}
		for _, f := range body["drift_findings"].([]any) {
			m := f.(map[string]any)
			if m["spec_id"] == "SPEC-DUP-001" && m["source"] != "worktree" {
				t.Errorf("%s: SPEC-DUP-001 finding must come from the worktree only, got source %v", name, m["source"])
			}
		}
		if !namesShadow(body) {
			t.Errorf("%s must name the shadowed primary copy, got %v", name, body["shadowed"])
		}
	}
}

// AC-WSR-012: with the primary unidentifiable the catalogue tools answer over
// the worktree only and say, in _root, that the primary catalogue was not read.
func TestWSR012_PrimaryUnidentifiedOnRead(t *testing.T) {
	wsrNoidVariants(t, wtWorkflowNoGate, func(t *testing.T, fx untrackedFixture) {
		wtWriteFile(t, filepath.Join(fx.W, ".moai", "specs", "SPEC-WTL-001", "spec.md"), wsrSpecBody("SPEC-WTL-001", "draft"))
		states := func(body map[string]any) bool {
			root, _ := body["_root"].(map[string]any)
			s, _ := root["primary_catalogue"].(string)
			return strings.Contains(s, "not read") && strings.Contains(s, "could not be identified")
		}
		var progress map[string]any
		decodeTool(t, handleSpecProgress, map[string]any{"project_root": fx.W}, &progress)
		if progress["count"] != float64(1) || !states(progress) {
			t.Errorf("spec_progress: want count 1 and a primary-not-read statement, got count=%v _root=%v", progress["count"], progress["_root"])
		}
		for name, h := range map[string]func(context.Context, mcpReq) (*mcpRes, error){
			"spec_drift": handleSpecDrift, "spec_audit": handleSpecAudit,
		} {
			var body map[string]any
			decodeTool(t, h, map[string]any{"project_root": fx.W}, &body)
			if body["total_specs"] != float64(1) || !states(body) {
				t.Errorf("%s: want total_specs 1 and a primary-not-read statement, got %v _root=%v", name, body["total_specs"], body["_root"])
			}
		}
	})
}

// baseFallbackWarning is the _root.warning text of the base tree, verbatim.
func baseFallbackWarning(source string) string {
	return "project_root not passed — resolved from " + source +
		", which froze at server spawn; a session that moved worktrees is reading another tree; " +
		"pass project_root = git rev-parse --show-toplevel"
}

// AC-WSR-013: on W every _root lists the sources read, the worktree warning
// no longer says the answer came from the worktree tree, and the fallback
// warning keeps its key and text.
func TestWSR013_RootProvenance(t *testing.T) {
	fx := newWSRFixture(t, wtWorkflowNoGate)
	sourcesOf := func(body map[string]any) map[string]string {
		root, _ := body["_root"].(map[string]any)
		out := map[string]string{}
		list, _ := root["sources"].([]any)
		for _, s := range list {
			m, _ := s.(map[string]any)
			src, _ := m["source"].(string)
			dir, _ := m["dir"].(string)
			out[src] = dir
		}
		return out
	}
	for name, h := range map[string]func(context.Context, mcpReq) (*mcpRes, error){
		"spec_progress": handleSpecProgress, "spec_audit": handleSpecAudit, "spec_drift": handleSpecDrift,
		"verify_snapshot": handleVerifySnapshot, "verify_trend": handleVerifyTrend,
	} {
		var body map[string]any
		decodeTool(t, h, map[string]any{"project_root": fx.W, "key": "h:d"}, &body)
		src := sourcesOf(body)
		if strings.HasPrefix(name, "spec_") {
			if src["worktree"] != fx.W || src["primary"] != fx.P {
				t.Errorf("%s: _root.sources must list worktree W and primary P, got %v", name, src)
			}
		} else if src["store"] != fx.P {
			t.Errorf("%s: _root.sources must list the store root P, got %v", name, src)
		}
		root, _ := body["_root"].(map[string]any)
		if w, _ := root["worktree_warning"].(string); w == "" || strings.Contains(w, "read from the worktree tree") {
			t.Errorf("%s: worktree_warning must be present and no longer say it was read from the worktree tree, got %q", name, w)
		}
	}
	t.Setenv("CLAUDE_PROJECT_DIR", fx.W)
	var body map[string]any
	decodeTool(t, handleSpecProgress, map[string]any{}, &body)
	root, _ := body["_root"].(map[string]any)
	if root["warning"] != baseFallbackWarning("env:CLAUDE_PROJECT_DIR") {
		t.Errorf("fallback _root.warning changed: %v", root["warning"])
	}
}

// wsrRepoRoot returns the repository root from the package directory.
func wsrRepoRoot(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

// AC-WSR-015: the rule section, its template mirror, the shared project_root
// description, and the tool descriptions state the new behaviour.
func TestWSR015_Documentation(t *testing.T) {
	repo := wsrRepoRoot(t)
	rel := filepath.Join(".claude", "rules", "moai", "core", "moai-mcp-tools-catalogue.md")
	local, err := os.ReadFile(filepath.Join(repo, rel))
	if err != nil {
		t.Fatal(err)
	}
	mirror, err := os.ReadFile(filepath.Join(repo, "internal", "template", "templates", rel))
	if err != nil {
		t.Fatal(err)
	}
	if string(local) != string(mirror) {
		t.Errorf("rule file and template mirror differ")
	}
	sectionStart := strings.Index(string(local), "## Linked worktrees of a repository that keeps `.moai` untracked")
	if sectionStart < 0 {
		t.Fatal("catalogue section heading not found")
	}
	section := string(local)[sectionStart:]
	for _, want := range []string{"primary checkout's `.moai/state`", "tree identity", "union", "receipt guard"} {
		if !strings.Contains(section, want) {
			t.Errorf("catalogue section lacks %q", want)
		}
	}
	if strings.Contains(section, "still read from the accepted tree") {
		t.Errorf("catalogue section still says state and catalogue are read from the accepted tree")
	}
	for _, pat := range []string{`SPEC-[A-Z]`, `\bt[0-9]{2,5}\b`, `20[0-9]{2}-[0-9]{2}-[0-9]{2}`} {
		if regexp.MustCompile(pat).Match(mirror) {
			t.Errorf("template copy matches forbidden pattern %s", pat)
		}
	}
	if strings.Contains(projectRootDescCommon, "still read from the accepted tree") {
		t.Errorf("projectRootDescCommon still says state is read from the accepted tree")
	}
	srv := newMoaiMCPServer()
	desc := func(name string) string {
		tool := srv.GetTool(name)
		if tool == nil {
			t.Fatalf("%s not registered", name)
		}
		return tool.Tool.Description
	}
	for _, name := range []string{"spec_progress", "spec_drift", "spec_audit"} {
		if d := desc(name); !strings.Contains(d, "primary checkout") || !strings.Contains(d, "union") {
			t.Errorf("%s description lacks the union catalogue statement: %s", name, d)
		}
	}
	for _, name := range []string{"verify_snapshot", "verify_trend", "codex_audit", auditMultiToolName} {
		if d := desc(name); !strings.Contains(d, "primary checkout") || !strings.Contains(d, ".moai/state") {
			t.Errorf("%s description lacks the primary-checkout state statement: %s", name, d)
		}
	}
	if d := desc("codex_audit"); strings.Contains(d, "that refusal is not guaranteed") {
		t.Errorf("codex_audit description still says the refusal is not guaranteed")
	}
}

// wsrSpecIDOfPath extracts the SPEC directory name from a DocRecord path.
func wsrSpecIDOfPath(p any) string {
	s, _ := p.(string)
	return filepath.Base(filepath.Dir(s))
}

// AC-WSR-014: the MCP writer, the hook writer, and the CLI writer agree on the
// store each root maps to, or on refusing to write.
func TestWSR014_OneStoreRootEverywhere(t *testing.T) {
	wsrTickClock(t)
	wsrStubCodexPass(t)
	scratch := wtCanonTempDir(t)

	fx := newWSRFixture(t, wsrWorkflowF)
	nfx := newWSRFixture(t, wsrWorkflowF)
	wsrNoidAdminHead(t, nfx)
	tP, wt := newTrackedFixture(t)
	b := wtCanonTempDir(t)
	wtWriteFile(t, filepath.Join(b, ".moai", "config", "sections", "workflow.yaml"), wtWorkflowCodexRequired)

	rows := []struct {
		label, root, want string // want "" = unresolved
	}{
		{"P", fx.P, fx.P}, {"W", fx.W, fx.P}, {"W F-noid", nfx.W, ""},
		{"WT", wt, wt}, {"T primary", tP, tP}, {"B", b, b},
	}
	for i, row := range rows {
		agent := "a14-" + string(rune('a'+i))
		// MCP writer.
		out, text := wsrCodexAudit(t, row.root)
		// Hook writer.
		t.Setenv("CLAUDE_PROJECT_DIR", scratch)
		if _, err := hook.NewSubagentStartHandler().Handle(context.Background(), &hook.HookInput{
			CWD: row.root, AgentID: agent, AgentType: auditreceipt.AgentPlanAuditor, SessionID: "s14",
		}); err != nil {
			t.Fatalf("%s: SubagentStart: %v", row.label, err)
		}
		// CLI writer.
		cmd := newVerifyCmd()
		var cliOut strings.Builder
		cmd.SetOut(&cliOut)
		cmd.SetErr(&cliOut)
		cmd.SetArgs([]string{"--project-root", row.root, "record", "--check-id", "c", "--command", "c-" + row.label})
		cliErr := cmd.Execute()

		if row.want == "" {
			if !strings.Contains(text, "primary checkout") || strings.Contains(text, `"audit_receipt"`) {
				t.Errorf("%s: codex_audit must carry a skip notice naming the primary checkout and no receipt; got %s", row.label, text)
			}
			for _, r := range []string{nfx.P, nfx.W} {
				if n := len(wsrJSONFiles(t, filepath.Join(r, ".moai", "state", "audit-receipts", "starts"))); n != 0 {
					t.Errorf("%s: no start marker may exist under %s, found %d", row.label, r, n)
				}
			}
			if cliErr == nil || !strings.Contains(cliErr.Error(), "primary checkout") {
				t.Errorf("%s: moai verify record must fail naming the primary checkout, got err=%v", row.label, cliErr)
			}
			_ = out
			wsrAssertEmptyMoai(t, nfx.W)
			continue
		}
		store := filepath.Join(row.want, ".moai", "state")
		mine := 0
		for _, f := range wsrJSONFiles(t, filepath.Join(store, "audit-receipts", "receipts")) {
			if wsrReadMap(t, f)["tree_root"] == row.root {
				mine++
			}
		}
		if mine != 1 {
			t.Errorf("%s: want one MCP receipt with tree_root %s under %s, found %d", row.label, row.root, store, mine)
		}
		if _, err := os.Stat(filepath.Join(store, "audit-receipts", "starts", agent+".json")); err != nil {
			t.Errorf("%s: hook start marker not under %s: %v", row.label, store, err)
		}
		if row.label == "B" {
			// B is not a git repository: `moai verify record` cannot compute a
			// working-tree key there on any tree, before a store is chosen.
			if cliErr == nil {
				t.Errorf("%s: premise: verify record on a non-git root was expected to fail at key computation", row.label)
			}
			continue
		}
		if cliErr != nil {
			t.Errorf("%s: moai verify record failed: %v (%s)", row.label, cliErr, cliOut.String())
		} else if key, kerr := verify.Key(context.Background(), row.root); kerr != nil {
			t.Errorf("%s: premise: verify.Key: %v", row.label, kerr)
		} else if snap, lerr := verify.Load(row.want, key); lerr != nil || snap == nil || !wsrHasCommand(snap.Checks, "c-"+row.label) {
			t.Errorf("%s: CLI check %q not recorded in the snapshot under %s (err=%v)", row.label, "c-"+row.label, store, lerr)
		}
	}
	wsrAssertNoMoai(t, fx.W)
}

// wsrHasCommand reports whether a snapshot's checks include command.
func wsrHasCommand(checks []verify.CheckEntry, command string) bool {
	for _, c := range checks {
		if c.Command == command {
			return true
		}
	}
	return false
}

// newTrackedFixture builds fixture T: a repository that commits
// .moai/config/sections/workflow.yaml (codex required), plus a linked worktree.
func newTrackedFixture(t *testing.T) (primary, wt string) {
	t.Helper()
	env := wtGitEnv(t)
	base := wtCanonTempDir(t)
	primary = filepath.Join(base, "T")
	mustMkdir(t, primary)
	wtGit(t, env, primary, "init", "-q", "-b", "main")
	wtWriteFile(t, filepath.Join(primary, ".moai", "config", "sections", "workflow.yaml"), wtWorkflowCodexRequired)
	// Runtime state stays ignored, as in a real tracked-.moai repository, so a
	// state write does not move the working-tree verify key.
	wtWriteFile(t, filepath.Join(primary, ".gitignore"), ".moai/state/\n")
	wtGit(t, env, primary, "add", ".gitignore", ".moai/config/sections/workflow.yaml")
	wtGit(t, env, primary, "commit", "-q", "-m", "init")
	wt = filepath.Join(base, "WT")
	wtGit(t, env, primary, "worktree", "add", "-q", "-b", "wt", wt)
	return primary, wt
}

// AC-WSR-016: end to end in the untracked-.moai shape.
func TestWSR016_EndToEndUntrackedShape(t *testing.T) {
	wsrTickClock(t)
	wsrStubCodexPass(t)
	fx := newWSRFixture(t, wsrWorkflowF)
	t.Setenv("CLAUDE_PROJECT_DIR", wtCanonTempDir(t))
	const agent = "a16"
	if _, err := hook.NewSubagentStartHandler().Handle(context.Background(), &hook.HookInput{
		CWD: fx.W, AgentID: agent, AgentType: auditreceipt.AgentPlanAuditor, SessionID: "s16",
	}); err != nil {
		t.Fatal(err)
	}
	out, _ := wsrCodexAudit(t, fx.W)
	if out.AuditReceipt == "" {
		t.Fatalf("premise: codex_audit on W must expose a receipt id (P's gate is required)")
	}
	stop, err := hook.NewSubagentStopHandler().Handle(context.Background(), &hook.HookInput{
		CWD: fx.W, AgentID: agent, AgentType: auditreceipt.AgentPlanAuditor,
		LastAssistantMessage: "done\nAUDIT-VERDICT: PASS spec=SPEC-PRI-001 receipts=" + out.AuditReceipt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if stop != nil && stop.Decision == "block" {
		t.Errorf("the stop citing the receipt must be accepted, got block: %s", stop.Reason)
	}
	var progress struct {
		Specs []map[string]any `json:"specs"`
	}
	decodeTool(t, handleSpecProgress, map[string]any{"project_root": fx.W}, &progress)
	found := false
	for _, rec := range progress.Specs {
		if wsrSpecIDOfPath(rec["Path"]) == "SPEC-PRI-001" {
			found = true
		}
	}
	if !found {
		t.Errorf("spec_progress on W must list SPEC-PRI-001")
	}
	wsrAssertNoMoai(t, fx.W)
}
