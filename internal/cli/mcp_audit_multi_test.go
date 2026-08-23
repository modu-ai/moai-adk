// Package cli — tests for the `audit_multi` MCP tool handler
// (SPEC-AUDIT-MULTI-MODEL-001 M3, AC-AMM-012 / AC-AMM-013 / AC-AMM-014).
//
// These tests are the TDD RED-GREEN drivers for the audit_multi MCP tool
// surface. The handler MUST be a THIN WRAPPER over runMultiAudit
// (mcp_convergence.go): it maps JSON-RPC params → (claude_verdict, target,
// focus, MultiAuditConfig), calls runMultiAudit, and shapes the
// ConvergenceResult into the tool's declared JSON output. NO backend
// re-implementation, NO AskUserQuestion, NO hard error.
//
// @MX:SPEC: SPEC-AUDIT-MULTI-MODEL-001
package cli

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
)

// recordingCallerMulti is the test-only backendCall seam that records every
// secondary call AND can be programmed per-backend (so tests assert the handler
// did NOT re-implement the backend — it delegated through runMultiAudit).
type recordingCallerMulti struct {
	mu        sync.Mutex // protects calls; runMultiAudit fans out to backends concurrently (AC-AMM-013 race fix)
	calls     []call
	verdictBy map[string]ReviewOutput // backend → verdict
}

type call struct {
	backend, target, focus, projectRoot string
}

func (r *recordingCallerMulti) call(ctx context.Context, backend, target, focus, projectRoot string) ReviewOutput {
	r.mu.Lock()
	r.calls = append(r.calls, call{backend: backend, target: target, focus: focus, projectRoot: projectRoot})
	r.mu.Unlock()
	if r.verdictBy == nil {
		return ReviewOutput{Verdict: "pass", Summary: backend + ":pass", Findings: []Finding{}, NextSteps: []string{}}
	}
	if v, ok := r.verdictBy[backend]; ok {
		return v
	}
	return ReviewOutput{Verdict: "pass", Summary: backend + ":pass", Findings: []Finding{}, NextSteps: []string{}}
}

// callToolAuditMulti builds a CallToolRequest for audit_multi with the given
// claude verdict object + args, so each test stays readable.
func callToolAuditMulti(t *testing.T, claudeVerdict map[string]any, extra map[string]any) (*mcp.CallToolResult, error) {
	t.Helper()
	args := map[string]any{}
	if claudeVerdict != nil {
		args["claude_verdict"] = claudeVerdict
	}
	for k, v := range extra {
		args[k] = v
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = "audit_multi"
	req.Params.Arguments = args
	return handleAuditMulti(context.Background(), req)
}

// AC-AMM-012 / AC-AMM-013: tools/list declares audit_multi with a JSON Schema
// including claude_verdict (object), target, focus, and a ConvergenceResult
// output shape. Verified structurally — the tool IS registered on the server.
func TestAuditMulti_RegisteredWithSchema_AC_AMM_012(t *testing.T) {
	srv := newMoaiMCPServer()
	ctx := context.Background()
	c, err := client.NewInProcessClient(srv)
	if err != nil {
		t.Fatalf("NewInProcessClient: %v", err)
	}
	defer closeInProcessClient(c)
	if _, err := c.Initialize(ctx, mcp.InitializeRequest{}); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	var found bool
	for _, tool := range res.Tools {
		if tool.Name != "audit_multi" {
			continue
		}
		found = true
		// Schema MUST mention claude_verdict + target + focus (focus optional).
		b, _ := json.Marshal(tool.InputSchema)
		schema := string(b)
		for _, want := range []string{"claude_verdict", "target", "focus"} {
			if !strings.Contains(schema, want) {
				t.Errorf("audit_multi inputSchema missing %q; schema=%s", want, schema)
			}
		}
	}
	if !found {
		t.Fatal("audit_multi tool NOT registered on the moai MCP server (AC-AMM-012)")
	}
}

// AC-AMM-013: the handler is a THIN WRAPPER — it delegates to runMultiAudit
// (which fans out through backendCall) and does NOT re-implement the backend.
// We swap backendCall with a recordingCallerMulti; if the handler bypasses
// runMultiAudit, no call would be recorded.
func TestAuditMulti_DelegatesToRunMultiAudit_AC_AMM_013(t *testing.T) {
	rc := &recordingCallerMulti{}
	orig := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = orig })

	claude := map[string]any{
		"verdict": "pass", "summary": "claude pass", "findings": []any{}, "next_steps": []any{},
	}
	res, err := callToolAuditMulti(t, claude, map[string]any{
		"target": "uncommittedChanges",
		"focus":  "concurrency",
	})
	if err != nil {
		t.Fatalf("handleAuditMulti returned Go error: %v (the handler MUST return a structured result, never a hard error)", err)
	}
	if res == nil {
		t.Fatal("nil result")
	}
	if res.IsError {
		t.Fatalf("IsError=true; content=%v (a passing claude anchor + passing secondaries MUST be a clean result)", res.Content)
	}
	// The handler delegated to runMultiAudit ⇒ it invoked ≥1 secondary backend.
	if len(rc.calls) == 0 {
		t.Error("no secondary backend was invoked — the handler did NOT delegate through runMultiAudit (AC-AMM-013 thin-wrapper)")
	}
}

// AC-AMM-003 / EC-6 (independence regression at the tool-handler layer): the
// claude_verdict passed to the tool MUST NOT reach the secondary backend. The
// recordingCallerMulti structurally excludes it from its signature; this test
// adds a distinctive secret into the claude verdict and asserts it does NOT
// surface in any recorded (target, focus) pair.
func TestAuditMulti_ClaudeVerdictNeverInSecondaryPayload_AC_AMM_003(t *testing.T) {
	rc := &recordingCallerMulti{}
	orig := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = orig })

	const secret = "CLAUDE-ANCHOR-SECRET-must-not-leak-98765"
	claude := map[string]any{
		"verdict": "pass", "summary": secret, "findings": []any{}, "next_steps": []any{secret},
	}
	if _, err := callToolAuditMulti(t, claude, map[string]any{"target": "uncommittedChanges"}); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	for _, c := range rc.calls {
		if strings.Contains(c.target, secret) || strings.Contains(c.focus, secret) {
			t.Errorf("claude verdict secret leaked into %s backend payload: target=%q focus=%q", c.backend, c.target, c.focus)
		}
	}
}

// AC-AMM-014: per-auditor audit_gate respected. With codex gate == off, the
// codex backend is NOT invoked and per_backend_verdicts carries no codex entry.
func TestAuditMulti_RespectsCodexGateOff_AC_AMM_014(t *testing.T) {
	rc := &recordingCallerMulti{}
	orig := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = orig })

	claude := map[string]any{
		"verdict": "pass", "summary": "ok", "findings": []any{}, "next_steps": []any{},
	}
	gates := map[string]any{
		"claude": config.AuditGateRequired,
		"codex":  config.AuditGateOff,
		"glm":    config.AuditGateAdvisory,
	}
	res, err := callToolAuditMulti(t, claude, map[string]any{"gates": gates})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("IsError=true; content=%v", res.Content)
	}
	for _, c := range rc.calls {
		if c.backend == BackendCodex {
			t.Errorf("codex backend invoked despite gate=off (AC-AMM-014); call=%+v", c)
		}
	}
	// And the result text must not contain a codex entry in per_backend_verdicts.
	if len(res.Content) == 0 {
		t.Fatal("empty content")
	}
	body := toolResultText(res)
	if strings.Contains(body, `"backend":"codex"`) {
		t.Errorf("codex entry present in per_backend_verdicts despite gate=off; body=%s", body)
	}
}

// DQ-2 surfacing: a missing claude_verdict anchor produces a structured result
// (overall = fail + a residual_risk_note explaining the missing anchor), NEVER
// a hard error. The tool-handler surface must preserve the fail-open direction.
func TestAuditMulti_MissingClaudeAnchor_StructuredRefusal(t *testing.T) {
	rc := &recordingCallerMulti{}
	orig := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = orig })

	// No claude_verdict supplied (empty verdict token).
	res, err := callToolAuditMulti(t, map[string]any{"verdict": ""}, nil)
	if err != nil {
		t.Fatalf("handler returned Go error on missing anchor: %v (must be a structured result)", err)
	}
	if res == nil || len(res.Content) == 0 {
		t.Fatal("nil / empty result for missing-anchor case")
	}
	body := toolResultText(res)
	if !strings.Contains(body, `"overall_verdict":"fail"`) {
		t.Errorf("missing-anchor case: expected overall_verdict=fail in result; body=%s", body)
	}
	if len(rc.calls) != 0 {
		t.Errorf("missing-anchor case: secondary backend invoked %d time(s); want 0 (refuse BEFORE fan-out)", len(rc.calls))
	}
}

// AC-AMM-024 (cross-cutting): the handler NEVER calls AskUserQuestion. Verified
// statically via grep elsewhere; here we verify behaviorally — every fail-open
// path produces a STRUCTURED result, not a panic / hard error.
func TestAuditMulti_NoHardErrorPath_AC_AMM_024(t *testing.T) {
	rc := &recordingCallerMulti{
		verdictBy: map[string]ReviewOutput{
			BackendCodex: {Verdict: VerdictInconclusive, Summary: "codex missing", Findings: []Finding{}, NextSteps: []string{}},
			BackendGLM:   {Verdict: VerdictInconclusive, Summary: "glm missing", Findings: []Finding{}, NextSteps: []string{}},
		},
	}
	orig := backendCall
	backendCall = rc.call
	t.Cleanup(func() { backendCall = orig })

	claude := map[string]any{
		"verdict": "pass", "summary": "claude ok", "findings": []any{}, "next_steps": []any{},
	}
	res, err := callToolAuditMulti(t, claude, nil)
	if err != nil {
		t.Fatalf("handler returned Go error on all-secondaries-inconclusive: %v", err)
	}
	if res.IsError {
		t.Errorf("IsError=true on all-secondaries-inconclusive (must fail-open to a structured ALLOW-shaped result); content=%v", res.Content)
	}
}

// toolResultText extracts the concatenated text content from a CallToolResult
// so the gate-off test can grep for the codex entry token without depending on
// the SDK's structured-content envelope.
func toolResultText(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}
