package cli

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M4 RED, sub-step S1 — an unrun leg is
// classified, never failed (AC-AEC-002a; REQ-AEC-006; card t2017).
//
// When an audit leg returns inconclusive because the review could not be
// performed, its entry in per_backend_verdicts and the audit receipt carry an
// error_class from the closed set quota, timeout, auth, unavailable, other. The
// leg's verdict stays inconclusive, and a leg that reported pass or fail carries
// no class.
//
// The legs are stubbed at the backend seam with the REAL could-not-run
// constructors (glmInconclusive, inconclusiveReview*, claudeInconclusive), so the
// messages are the ones the backends produce. The result is decoded as generic
// JSON: the test needs no new production symbol.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// classWorld is a throwaway tree whose codex gate is declared required, so the
// audit receipt is filed and exposed.
func classWorld(t *testing.T) deltaRepo {
	t.Helper()
	d := newDeltaRepo(t)
	dir := filepath.Join(d.root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "workflow:\n  audit:\n    gates:\n      codex: required\n"
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return d
}

// callClassified runs audit_multi over d with every leg required and the given leg
// outputs (a missing backend passes), and returns the decoded result.
func callClassified(t *testing.T, d deltaRepo, legs map[string]ReviewOutput) map[string]any {
	t.Helper()
	prev := backendCall
	backendCall = func(_ context.Context, backend, _, _, _ string) ReviewOutput {
		if out, ok := legs[backend]; ok {
			return out
		}
		return claudeReview("pass")
	}
	t.Cleanup(func() { backendCall = prev })

	res, err := handleAuditMulti(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"target": codexTargetBaseBranch, "project_root": d.root,
		"gates": map[string]any{"claude": "required", "codex": "required", "glm": "required"},
	}}})
	if err != nil || res.IsError {
		t.Fatalf("handleAuditMulti: err=%v isError=%v %v", err, res.IsError, res.Content)
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
	return out
}

// legEntry returns the per_backend_verdicts entry of one backend.
func legEntry(t *testing.T, result map[string]any, backend string) map[string]any {
	t.Helper()
	list, _ := result["per_backend_verdicts"].([]any)
	for _, e := range list {
		m, _ := e.(map[string]any)
		if m["backend"] == backend {
			return m
		}
	}
	t.Fatalf("no per_backend_verdicts entry for %s in %v", backend, list)
	return nil
}

// AC-AEC-002a — the could-not-run causes the backends actually produce, one leg at
// a time: the leg stays inconclusive and carries the class.
func TestAEC002a_UnrunLeg_CarriesAnErrorClass_VerdictStaysInconclusive(t *testing.T) {
	rows := []struct {
		name    string
		backend string
		out     ReviewOutput
		want    string
	}{
		{"glm HTTP 429 is quota", BackendGLM, glmInconclusive("z.ai returned HTTP 429"), "quota"},
		{"codex usage limit is quota", BackendCodex, inconclusiveReview("codex turn failed: usage limit reached, quota exhausted"), "quota"},
		{"codex leg deadline is timeout", BackendCodex, inconclusiveReviewWithSummary("codex leg timed out after 5m0s"), "timeout"},
		{"glm request deadline is timeout", BackendGLM, glmInconclusive("z.ai request failed: Post \"https://api.z.ai/x\": context deadline exceeded"), "timeout"},
		{"claude subscription refusal is auth", BackendClaude, claudeInconclusive(claudeErrSubscriptionUnavailable, "Claude subscription authentication is unavailable", "", ""), "auth"},
		{"glm HTTP 401 is auth", BackendGLM, glmInconclusive("z.ai returned HTTP 401"), "auth"},
		{"codex binary missing is unavailable", BackendCodex, inconclusiveReview("codex binary not found in PATH"), "unavailable"},
		{"an unrecognized cause is other", BackendGLM, glmInconclusive("malformed z.ai response: invalid character"), "other"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			d := classWorld(t)
			result := callClassified(t, d, map[string]ReviewOutput{r.backend: r.out})
			e := legEntry(t, result, r.backend)
			if e["verdict"] != VerdictInconclusive {
				t.Errorf("the leg's verdict = %v, want it to stay %q", e["verdict"], VerdictInconclusive)
			}
			if got, _ := e["error_class"].(string); got != r.want {
				t.Errorf("error_class = %q, want %q (entry: %v)", got, r.want, e)
			}
		})
	}
}

// AC-AEC-002a — the acceptance fixture itself: three legs fail at once with a
// quota message, a deadline and an authentication refusal; each carries its own
// class. A mutant that sets every class to "other" fails here.
func TestAEC002a_ThreeLegsFailAtOnce_ThreeDistinctClasses(t *testing.T) {
	d := classWorld(t)
	result := callClassified(t, d, map[string]ReviewOutput{
		BackendCodex:  inconclusiveReview("codex turn failed: usage limit reached, quota exhausted"),
		BackendGLM:    glmInconclusive("z.ai request failed: context deadline exceeded"),
		BackendClaude: claudeInconclusive(claudeErrSubscriptionUnavailable, "Claude subscription authentication is unavailable", "", ""),
	})
	for backend, want := range map[string]string{BackendCodex: "quota", BackendGLM: "timeout", BackendClaude: "auth"} {
		e := legEntry(t, result, backend)
		if got, _ := e["error_class"].(string); got != want {
			t.Errorf("%s error_class = %q, want %q", backend, got, want)
		}
	}
}

// AC-AEC-002a — a leg that reported pass or fail carries no class at all.
func TestAEC002a_PassAndFailLegs_CarryNoErrorClass(t *testing.T) {
	d := classWorld(t)
	result := callClassified(t, d, map[string]ReviewOutput{
		BackendCodex: {Verdict: "fail", Summary: "stub: codex found a defect", Findings: []Finding{}, NextSteps: []string{}},
		BackendGLM:   glmInconclusive("z.ai returned HTTP 429"), // one unrun leg beside them
	})
	for _, backend := range []string{BackendClaude, BackendCodex} {
		e := legEntry(t, result, backend)
		if _, has := e["error_class"]; has {
			t.Errorf("%s reported %v and must carry no error_class: %v", backend, e["verdict"], e)
		}
	}
}

// AC-AEC-002a — the audit receipt carries the class of an unrun codex leg, and a
// receipt of a codex leg that ran carries none.
func TestAEC002a_Receipt_CarriesTheCodexLegsErrorClass(t *testing.T) {
	readReceipt := func(t *testing.T, root string) map[string]any {
		t.Helper()
		dir := filepath.Join(root, ".moai", "state", "audit-receipts", "receipts")
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) != 1 {
			t.Fatalf("want exactly one receipt in the test tree, got %d (err %v)", len(entries), err)
		}
		b, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("decode receipt: %v", err)
		}
		return m
	}

	t.Run("an unrun codex leg", func(t *testing.T) {
		d := classWorld(t)
		callClassified(t, d, map[string]ReviewOutput{BackendCodex: inconclusiveReview("codex turn failed: usage limit reached, quota exhausted")})
		r := readReceipt(t, d.root)
		if r["codex_verdict"] != VerdictInconclusive {
			t.Fatalf("setup: codex_verdict = %v", r["codex_verdict"])
		}
		if got, _ := r["error_class"].(string); got != "quota" {
			t.Errorf("the receipt's error_class = %q, want quota; receipt: %v", got, r)
		}
	})
	t.Run("a codex leg that ran", func(t *testing.T) {
		d := classWorld(t)
		callClassified(t, d, nil)
		r := readReceipt(t, d.root)
		if _, has := r["error_class"]; has {
			t.Errorf("a receipt of a codex leg that ran must carry no error_class: %v", r)
		}
	})
}

// AC-AEC-002a — the single-backend codex audit files the same receipt field.
func TestAEC002a_SingleBackendReceiptsDoNotInventAClass_Control(t *testing.T) {
	// Control (passes today): the audit_multi result of an all-pass fan-out has no
	// error_class anywhere, so a consumer that reads the field sees it only where
	// a leg could not run.
	d := classWorld(t)
	result := callClassified(t, d, nil)
	list, _ := result["per_backend_verdicts"].([]any)
	if len(list) != 3 {
		t.Fatalf("want three legs, got %d", len(list))
	}
	for _, e := range list {
		if strings.Contains(marshalForClassCheck(e), "error_class") {
			t.Errorf("an all-pass leg carries error_class: %v", e)
		}
	}
}

func marshalForClassCheck(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
