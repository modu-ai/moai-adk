package cli

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M4 RED, sub-step S3 — re-run only the named
// legs and carry the others from the server's own persisted result (AC-AEC-002c;
// REQ-AEC-008; card t2017).
//
// A call that names legs to re-run (the input `rerun_legs`, an array of backend
// names) runs only those legs and carries every other leg's verdict from the
// result the server persisted for the same session, provided that persisted result
// records the same head commit, the same tree digest and the same scope (target,
// resolved base, delta base). When no such result exists or any of those differ,
// every leg runs. A carried verdict is never read from a caller-supplied argument.
//
// The legs are stubbed at the backendCall seam, which counts calls per backend; the
// result is decoded as generic JSON. The test pins the input name `rerun_legs` (a
// lane decision; the SPEC names none) and nothing about the persisted field names:
// the first call persists whatever it persists and the second call reads it back.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

const rerunSession = "sess-aec-002c"

// rerunCalls counts the calls the backend seam received, per backend.
type rerunCalls struct {
	mu sync.Mutex
	n  map[string]int
}

func (c *rerunCalls) get(backend string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n[backend]
}

// rerunLegOut is a leg output with a summary that names the call it came from, so a
// carried verdict is told apart from a re-run one.
func rerunLegOut(verdict, summary string) ReviewOutput {
	out := claudeReview(verdict)
	out.Summary = summary
	return out
}

// callRerun runs audit_multi over d for the session, with every leg required and the
// given leg outputs (a backend not listed passes with the summary "<backend>:default"),
// and returns the raw tool result, the decoded result and the per-backend call counts.
func callRerun(t *testing.T, d deltaRepo, session string, extra map[string]any, outs map[string]ReviewOutput) (*mcp.CallToolResult, map[string]any, *rerunCalls) {
	t.Helper()
	calls := &rerunCalls{n: map[string]int{}}
	prev := backendCall
	backendCall = func(_ context.Context, backend, _, _, _ string) ReviewOutput {
		calls.mu.Lock()
		calls.n[backend]++
		calls.mu.Unlock()
		if out, ok := outs[backend]; ok {
			return out
		}
		return rerunLegOut("pass", backend+":default")
	}
	t.Cleanup(func() { backendCall = prev })

	args := map[string]any{
		"target": codexTargetBaseBranch, "project_root": d.root, "session_id": session,
		"gates": map[string]any{"claude": "required", "codex": "required", "glm": "required"},
	}
	for k, v := range extra {
		args[k] = v
	}
	res, err := handleAuditMulti(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}})
	if err != nil {
		t.Fatalf("handleAuditMulti: %v", err)
	}
	if res.IsError {
		return res, nil, calls
	}
	var raw string
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			raw += tc.Text
		}
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("decode: %v\n%s", err, raw)
	}
	return res, out, calls
}

// requireAllLegsRan fails unless each of the three backends was called exactly once.
func requireAllLegsRan(t *testing.T, calls *rerunCalls) {
	t.Helper()
	for _, b := range []string{BackendClaude, BackendCodex, BackendGLM} {
		if got := calls.get(b); got != 1 {
			t.Errorf("%s leg was called %d times, want 1 (every leg runs)", b, got)
		}
	}
}

// firstFullCall runs the session's first call: every leg runs, codex fails, claude
// and glm pass. It checks the control facts the later steps rest on and returns the
// decoded result.
func firstFullCall(t *testing.T, d deltaRepo, session string) map[string]any {
	t.Helper()
	_, first, calls := callRerun(t, d, session, nil, map[string]ReviewOutput{
		BackendClaude: rerunLegOut("pass", "claude:first"),
		BackendCodex:  rerunLegOut("fail", "codex:first"),
		BackendGLM:    rerunLegOut("pass", "glm:first"),
	})
	requireAllLegsRan(t, calls)
	if first["overall_verdict"] != "fail" {
		t.Fatalf("first call overall_verdict = %v, want fail (codex failed)", first["overall_verdict"])
	}
	if _, err := os.Stat(filepath.Join(d.root, ".moai", "state", "audit-multi", session+".json")); err != nil {
		t.Fatalf("the first call must persist the session's result: %v", err)
	}
	return first
}

func entryJSON(t *testing.T, result map[string]any, backend string) string {
	t.Helper()
	return marshalForClassCheck(legEntry(t, result, backend))
}

// AC-AEC-002c — the re-run: only the named leg runs, the other legs' verdicts are
// carried, and the overall verdict follows the combination.
func TestAEC002c_RerunNamedLeg_CarriesTheOthers(t *testing.T) {
	d := classWorld(t)
	first := firstFullCall(t, d, rerunSession)

	_, second, calls := callRerun(t, d, rerunSession,
		map[string]any{"rerun_legs": []any{"codex"}},
		map[string]ReviewOutput{BackendCodex: rerunLegOut("pass", "codex:second")})

	if calls.get(BackendClaude) != 0 || calls.get(BackendGLM) != 0 {
		t.Errorf("claude and glm calls = %d and %d, want 0 and 0 (their verdicts are carried)", calls.get(BackendClaude), calls.get(BackendGLM))
	}
	if calls.get(BackendCodex) != 1 {
		t.Errorf("codex leg was called %d times, want 1 (it is the named leg)", calls.get(BackendCodex))
	}
	for _, carried := range []string{BackendClaude, BackendGLM} {
		if got, want := entryJSON(t, second, carried), entryJSON(t, first, carried); got != want {
			t.Errorf("%s verdict was not carried unchanged:\n got %s\nwant %s", carried, got, want)
		}
	}
	if got := legEntry(t, second, BackendCodex)["summary"]; got != "codex:second" {
		t.Errorf("codex summary = %v, want the re-run's codex:second", got)
	}
	if second["overall_verdict"] != "pass" {
		t.Errorf("overall_verdict = %v, want pass (claude and glm carried as pass, codex re-run passes)", second["overall_verdict"])
	}
}

// The persisted result of the re-run carries the new combination, so a further
// re-run builds on it.
func TestAEC002c_RerunResultIsPersistedForTheNextRerun(t *testing.T) {
	d := classWorld(t)
	firstFullCall(t, d, rerunSession)
	callRerun(t, d, rerunSession,
		map[string]any{"rerun_legs": []any{"codex"}},
		map[string]ReviewOutput{BackendCodex: rerunLegOut("pass", "codex:second")})

	_, third, calls := callRerun(t, d, rerunSession,
		map[string]any{"rerun_legs": []any{"glm"}},
		map[string]ReviewOutput{BackendGLM: rerunLegOut("pass", "glm:third")})

	if calls.get(BackendClaude) != 0 || calls.get(BackendCodex) != 0 || calls.get(BackendGLM) != 1 {
		t.Errorf("calls claude/codex/glm = %d/%d/%d, want 0/0/1", calls.get(BackendClaude), calls.get(BackendCodex), calls.get(BackendGLM))
	}
	if got := legEntry(t, third, BackendCodex)["summary"]; got != "codex:second" {
		t.Errorf("the codex verdict carried from the second call = %v, want codex:second", got)
	}
}

// A carried codex leg took no new run, so no new audit receipt is recorded: the
// result keeps the receipt id the codex run itself was filed under.
func TestAEC002c_CarriedCodexLeg_KeepsItsReceipt(t *testing.T) {
	d := classWorld(t)
	first := firstFullCall(t, d, rerunSession)
	if first["audit_receipt"] == nil || first["audit_receipt"] == "" {
		t.Fatalf("the first call must file a codex receipt (codex gate required): %v", first)
	}

	_, second, calls := callRerun(t, d, rerunSession,
		map[string]any{"rerun_legs": []any{"glm"}},
		map[string]ReviewOutput{BackendGLM: rerunLegOut("pass", "glm:second")})

	if calls.get(BackendCodex) != 0 {
		t.Errorf("codex leg was called %d times, want 0 (carried)", calls.get(BackendCodex))
	}
	if second["audit_receipt"] != first["audit_receipt"] {
		t.Errorf("audit_receipt = %v, want the first call's %v (no new receipt for a carried codex leg)", second["audit_receipt"], first["audit_receipt"])
	}
}

// AC-AEC-002c — when anything the persisted result recorded differs, or it recorded
// nothing, every leg runs. Each row names the change made between the two calls.
func TestAEC002c_AnyDifference_RunsEveryLeg(t *testing.T) {
	rows := []struct {
		name string
		// change mutates the world between the first and the second call and
		// returns the second call's session and extra arguments.
		change func(t *testing.T, d deltaRepo) (session string, extra map[string]any)
	}{
		{"a new commit moves the head", func(t *testing.T, d deltaRepo) (string, map[string]any) {
			if err := os.WriteFile(filepath.Join(d.root, "d.txt"), []byte("canary_d\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			reviewTargetGit(t, d.root, "add", "d.txt")
			reviewTargetGit(t, d.root, "commit", "-m", "add d.txt")
			return rerunSession, nil
		}},
		{"an uncommitted edit moves the tree digest", func(t *testing.T, d deltaRepo) (string, map[string]any) {
			if err := os.WriteFile(filepath.Join(d.root, "b.txt"), []byte("canary_b_edited\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			return rerunSession, nil
		}},
		{"a different target", func(t *testing.T, d deltaRepo) (string, map[string]any) {
			return rerunSession, map[string]any{"target": codexTargetUncommitted}
		}},
		{"a different resolved base", func(t *testing.T, d deltaRepo) (string, map[string]any) {
			reviewTargetGit(t, d.root, "update-ref", "refs/remotes/origin/main", d.shaB)
			return rerunSession, nil
		}},
		{"a different delta base", func(t *testing.T, d deltaRepo) (string, map[string]any) {
			return rerunSession, map[string]any{"delta_base": d.shaB}
		}},
		{"a different session has no persisted result", func(t *testing.T, d deltaRepo) (string, map[string]any) {
			return "sess-other", nil
		}},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			d := classWorld(t)
			firstFullCall(t, d, rerunSession)
			session, extra := r.change(t, d)
			if extra == nil {
				extra = map[string]any{}
			}
			extra["rerun_legs"] = []any{"codex"}
			_, _, calls := callRerun(t, d, session, extra, nil)
			requireAllLegsRan(t, calls)
		})
	}
}

// A persisted result that predates the fields the comparison needs (it carries no
// head, digest or scope) is never carried from.
func TestAEC002c_LegacyPersistedResult_RunsEveryLeg(t *testing.T) {
	d := classWorld(t)
	dir := filepath.Join(d.root, ".moai", "state", "audit-multi")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := `{"per_backend_verdicts":[` +
		`{"backend":"claude","gate":"required","verdict":"pass","summary":"claude:legacy","findings":[],"next_steps":[]},` +
		`{"backend":"codex","gate":"required","verdict":"pass","summary":"codex:legacy","findings":[],"next_steps":[]},` +
		`{"backend":"glm","gate":"required","verdict":"pass","summary":"glm:legacy","findings":[],"next_steps":[]}],` +
		`"overall_verdict":"pass","disagreement_flag":false,"participant_count":3,"residual_risk_note":"","fail_open_backends":[]}`
	if err := os.WriteFile(filepath.Join(dir, rerunSession+".json"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	_, result, calls := callRerun(t, d, rerunSession, map[string]any{"rerun_legs": []any{"codex"}}, nil)
	requireAllLegsRan(t, calls)
	for _, b := range []string{BackendClaude, BackendGLM} {
		if got := legEntry(t, result, b)["summary"]; got == b+":legacy" {
			t.Errorf("%s verdict was carried from a persisted result that records no head, digest or scope", b)
		}
	}
}

// AC-AEC-002c — a verdict the caller supplies for a carried leg is ignored: the
// carried leg keeps the verdict the server persisted.
func TestAEC002c_CallerVerdictForACarriedLeg_IsIgnored(t *testing.T) {
	t.Setenv(config.EnvMoaiLaunchProvider, BackendClaude)
	d := classWorld(t)
	first := firstFullCall(t, d, rerunSession)
	callerVerdict := map[string]any{"verdict": "fail", "summary": "claude:caller-supplied", "findings": []any{}, "next_steps": []any{}}

	_, second, calls := callRerun(t, d, rerunSession,
		map[string]any{"rerun_legs": []any{"codex"}, "claude_verdict": callerVerdict},
		map[string]ReviewOutput{BackendCodex: rerunLegOut("pass", "codex:second")})

	if calls.get(BackendClaude) != 0 {
		t.Errorf("claude leg was called %d times, want 0 (carried)", calls.get(BackendClaude))
	}
	if got, want := entryJSON(t, second, BackendClaude), entryJSON(t, first, BackendClaude); got != want {
		t.Errorf("the carried claude verdict was replaced by the caller's:\n got %s\nwant %s", got, want)
	}
	if second["overall_verdict"] != "pass" {
		t.Errorf("overall_verdict = %v, want pass (the caller's fail must not count)", second["overall_verdict"])
	}
}

// Control, passes today: without rerun_legs the same caller verdict IS honored as the
// in-session anchor in a Claude-origin session, so the test above is a real
// discrimination and not a vacuous one.
func TestAEC002c_CallerVerdictWithoutRerun_IsTheAnchor_Control(t *testing.T) {
	t.Setenv(config.EnvMoaiLaunchProvider, BackendClaude)
	d := classWorld(t)
	callerVerdict := map[string]any{"verdict": "fail", "summary": "claude:caller-supplied", "findings": []any{}, "next_steps": []any{}}
	_, result, _ := callRerun(t, d, rerunSession, map[string]any{"claude_verdict": callerVerdict}, nil)
	if got := legEntry(t, result, BackendClaude)["summary"]; got != "claude:caller-supplied" {
		t.Errorf("claude anchor summary = %v, want the caller's verdict honored when no leg is carried", got)
	}
}

// Controls, pass today: no rerun_legs, or an empty list, runs every leg.
func TestAEC002c_NoRerunRequest_RunsEveryLeg_Control(t *testing.T) {
	for name, extra := range map[string]map[string]any{
		"no rerun_legs":    nil,
		"empty rerun_legs": {"rerun_legs": []any{}},
	} {
		t.Run(name, func(t *testing.T) {
			d := classWorld(t)
			firstFullCall(t, d, rerunSession)
			_, _, calls := callRerun(t, d, rerunSession, extra, nil)
			requireAllLegsRan(t, calls)
		})
	}
}

// rerun_legs names backends: a name outside claude, codex, glm is a caller error, and
// it is refused before any leg runs.
func TestAEC002c_RerunLegsRejectsUnknownNames(t *testing.T) {
	for name, legs := range map[string]any{
		"unknown backend":   []any{"codex", "gemini"},
		"not an array":      "codex",
		"non-string member": []any{7},
	} {
		t.Run(name, func(t *testing.T) {
			d := classWorld(t)
			firstFullCall(t, d, rerunSession)
			res, _, calls := callRerun(t, d, rerunSession, map[string]any{"rerun_legs": legs}, nil)
			if !res.IsError {
				t.Errorf("rerun_legs %v was accepted, want a tool error", legs)
			}
			for _, b := range []string{BackendClaude, BackendCodex, BackendGLM} {
				if calls.get(b) != 0 {
					t.Errorf("%s leg ran (%d) although the input was refused", b, calls.get(b))
				}
			}
		})
	}
}
