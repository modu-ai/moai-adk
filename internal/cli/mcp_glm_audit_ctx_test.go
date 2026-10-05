package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// SPEC-CTX-BLIND-DOUBLE-001 M2 (card t539) — the GLM AUDIT path had no
// cancellation guard, and the reason was the double, not the production code.
//
// mcp_glm.go:305 already builds its request with http.NewRequestWithContext,
// so the context reaches the seam correctly. What was missing is a test that
// can TELL. Every audit-path test drives handleGLMAudit through stubGLMDoer
// (mcp_glm_test.go:26), which reads req.URL and req.Body and never calls
// req.Context() — so it answers a request whose context is already dead as
// readily as a live one. That makes the stub STRONGER than the thing it
// models: a real *http.Client returns context.Canceled there.
//
// Measured consequence: mutating line 305 to http.NewRequest (dropping ctx)
// left `-run GLM` (223 tests), `-run Audit` (99) and `-run Converg` (32) all
// green. A defect that detaches the audit request from its context is
// invisible to the whole selector set — the mutant survives.
//
// The two tests below are that missing discrimination. The first drives the
// handler with an ALREADY-CANCELLED context through ctxAwareGLMDoer (the
// faithful double from glm_task_bg_context_test.go:48 — same package, reused
// rather than duplicated) and asserts the outcome a real http.Client would
// produce. The second is its mirror image on a LIVE context, so the first
// cannot be satisfied by any failure whatever: together they assert that the
// context is what decides, not the weather.
//
// Scope note: stubGLMDoer is deliberately LEFT BLIND. It has 31 call sites
// across 7 files whose subject is the URL, the request body, and the fail-open
// verdict — none of them is about cancellation, and making the shared stub
// context-aware would change what those tests mean in order to guard a
// dimension only this file is about. The ctx dimension is covered here, on its
// own doubles. (SPEC §G "일괄 교체" anti-pattern.)

// glmAuditReview drives handleGLMAudit and decodes the ReviewOutput out of the
// structured result, failing the test on any shape it cannot read.
func glmAuditReview(t *testing.T, ctx context.Context, root string) ReviewOutput {
	t.Helper()
	res, err := handleGLMAudit(ctx, glmAuditReq(root))
	if err != nil {
		t.Fatalf("handleGLMAudit returned a Go error (must fail open, not error): %v", err)
	}
	if res.IsError {
		t.Fatalf("handleGLMAudit returned IsError (must be a structured result): %+v", res)
	}
	if len(res.Content) == 0 {
		t.Fatal("handleGLMAudit returned no content blocks")
	}
	tc, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content[0] type = %T, want mcp.TextContent", res.Content[0])
	}
	var out ReviewOutput
	if err := json.Unmarshal([]byte(tc.Text), &out); err != nil {
		t.Fatalf("result text is not a ReviewOutput JSON: %v (text=%q)", err, tc.Text)
	}
	return out
}

// TestGLMAudit_CancelledContext_IsNotSwallowed is the guard the survived
// mutant named. A dead context must reach the transport and come back as the
// fail-open inconclusive verdict — NEVER as the canned success the doer would
// hand a live request.
func TestGLMAudit_CancelledContext_IsNotSwallowed(t *testing.T) {
	// The faithful double: answers only a request whose context is still
	// alive, exactly as *http.Client does.
	withGLMSeams(t, "test-glm-key", &ctxAwareGLMDoer{
		body: glmMessagesResp(t, ReviewOutput{Verdict: "pass", Summary: "canned success"}),
	})
	root := newGLMReviewTree(t, true) // one uncommitted change ⇒ real diff material

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // the host revoked the request scope before the call

	out := glmAuditReview(t, ctx, root)

	// The canned success must NOT survive a dead context. This is the
	// assertion the mutant kills: with ctx detached at mcp_glm.go:305 the
	// doer sees a live background context and answers "pass".
	if out.Verdict == "pass" {
		t.Fatalf("a cancelled audit returned the canned success verdict %q — "+
			"the request reached the transport detached from its context "+
			"(check http.NewRequestWithContext at mcp_glm.go:305)", out.Verdict)
	}
	if out.Verdict != VerdictInconclusive {
		t.Fatalf("verdict = %q, want %q on a cancelled context", out.Verdict, VerdictInconclusive)
	}
	// The reason must name the context failure rather than some other
	// fail-open path (a missing key or an empty diff would also be
	// inconclusive, and neither would prove the context travelled).
	if !strings.Contains(out.Summary, context.Canceled.Error()) {
		t.Errorf("summary = %q, want it to name %q — an inconclusive verdict "+
			"that does not name the context error does not establish that the "+
			"context reached the transport", out.Summary, context.Canceled.Error())
	}
}

// TestGLMAudit_LiveContext_StillReachesTheDoer is the mirror image. Without
// it the test above would pass just as well against a handler that fails on
// EVERY audit — an assertion satisfied by any failure asserts nothing about
// cancellation.
func TestGLMAudit_LiveContext_StillReachesTheDoer(t *testing.T) {
	withGLMSeams(t, "test-glm-key", &ctxAwareGLMDoer{
		body: glmMessagesResp(t, ReviewOutput{Verdict: "pass", Summary: "canned success"}),
	})
	root := newGLMReviewTree(t, true)

	out := glmAuditReview(t, context.Background(), root)

	if out.Verdict != "pass" {
		t.Fatalf("verdict = %q on a LIVE context, want pass — the same doer that "+
			"must refuse a dead request has to answer a live one, or the "+
			"cancellation test is passing for the wrong reason (summary=%q)",
			out.Verdict, out.Summary)
	}
}
