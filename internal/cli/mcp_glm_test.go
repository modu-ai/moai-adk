package cli

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// SPEC-MOAI-MCP-SERVER-001 M3 — glm_audit backend (REQ-MCP-009/011/012/013/014,
// AC-MCP-011/013/014/015/016). RED until internal/cli/mcp_glm.go exists.
//
// Tests MUST NOT require a real GLM key or network: the HTTP doer + key loader
// are injectable seams (mirroring the M2 codexRunner/codexLookPath pattern).
// Every fail-open path returns a structured VerdictInconclusive ReviewOutput,
// never a hard error, never AskUserQuestion (subagent boundary).

// stubGLMDoer is the injectable HTTP doer. It captures the request body so a
// test can assert what was sent to z.ai, and returns a canned response.
type stubGLMDoer struct {
	status  int
	body    string
	err     error
	gotBody string
	gotURL  string
}

func (s *stubGLMDoer) Do(req *http.Request) (*http.Response, error) {
	if s.err != nil {
		return nil, s.err
	}
	s.gotURL = req.URL.String()
	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		s.gotBody = string(b)
		_ = req.Body.Close()
	}
	code := s.status
	if code == 0 {
		code = http.StatusOK
	}
	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(strings.NewReader(s.body)),
		Header:     make(http.Header),
	}, nil
}

// glmMessagesResp builds an Anthropic-compatible /v1/messages response whose
// single text block carries the given (JSON-encoded ReviewOutput) payload.
func glmMessagesResp(t *testing.T, review ReviewOutput) string {
	t.Helper()
	text, err := json.Marshal(review)
	if err != nil {
		t.Fatalf("marshal review: %v", err)
	}
	resp := map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(text)},
		},
	}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal resp: %v", err)
	}
	return string(b)
}

// withGLMSeams swaps the key loader + HTTP client and restores them on cleanup.
func withGLMSeams(t *testing.T, key string, doer glmHTTPDoer) {
	t.Helper()
	oldKey, oldHTTP := glmKeyLoader, glmHTTPClient
	glmKeyLoader = func() string { return key }
	if doer != nil {
		glmHTTPClient = doer
	}
	t.Cleanup(func() {
		glmKeyLoader = oldKey
		glmHTTPClient = oldHTTP
	})
}

func TestGLMAudit_FailOpenOnMissingKey(t *testing.T) {
	stub := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass"})}
	withGLMSeams(t, "", stub) // empty key ⇒ missing/unauthenticated

	res, err := handleGLMAudit(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("handler returned Go error (must fail-open, not error): %v", err)
	}
	if res.IsError {
		t.Fatalf("handler returned IsError (must be a structured inconclusive result)")
	}
	// No HTTP call must have been made — a missing key short-circuits before z.ai.
	if stub.gotURL != "" {
		t.Errorf("missing key still hit z.ai at %q", stub.gotURL)
	}
}

func TestGLMAudit_StubReturnsVerdict(t *testing.T) {
	want := ReviewOutput{
		Verdict:   "pass",
		Summary:   "no blocking findings",
		Findings:  []Finding{},
		NextSteps: []string{"proceed to merge"},
	}
	stub := &stubGLMDoer{body: glmMessagesResp(t, want)}
	withGLMSeams(t, "test-glm-key", stub)

	res, err := handleGLMAudit(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("handler IsError on a passing stub: %+v", res)
	}
	// The result must carry a ReviewOutput-shaped structured payload whose
	// verdict is the one the stub returned (uniform §G.4 shape, AC-MCP-011).
	if len(res.Content) == 0 {
		t.Fatal("no content blocks")
	}
	tc, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content[0] type = %T, want TextContent", res.Content[0])
	}
	var got ReviewOutput
	if err := json.Unmarshal([]byte(tc.Text), &got); err != nil {
		t.Fatalf("result text is not a ReviewOutput JSON: %v (text=%q)", err, tc.Text)
	}
	if got.Verdict != "pass" {
		t.Errorf("verdict = %q, want pass", got.Verdict)
	}
	if len(got.NextSteps) != 1 || got.NextSteps[0] != "proceed to merge" {
		t.Errorf("next_steps = %v, want [proceed to merge]", got.NextSteps)
	}
	// AC-MCP-011: the request must target the z.ai Anthropic-compatible endpoint.
	if !strings.Contains(stub.gotURL, "api.z.ai") {
		t.Errorf("request URL %q does not hit z.ai (api.z.ai)", stub.gotURL)
	}
	if !strings.Contains(stub.gotURL, "/v1/messages") {
		t.Errorf("request URL %q is not the Anthropic /v1/messages path", stub.gotURL)
	}
}

func TestGLMAudit_HitsZaiEndpointDirectly(t *testing.T) {
	// NOT the z.ai MCP server (api.z.ai/api/mcp/...), NOT any gateway — the
	// Anthropic-compatible api.z.ai/api/anthropic/v1/messages surface.
	stub := &stubGLMDoer{body: glmMessagesResp(t, ReviewOutput{Verdict: "pass"})}
	withGLMSeams(t, "k", stub)

	_, _ = handleGLMAudit(context.Background(), mcp.CallToolRequest{})
	if !strings.HasPrefix(stub.gotURL, "https://api.z.ai/api/anthropic/v1/messages") {
		t.Errorf("URL = %q, want prefix https://api.z.ai/api/anthropic/v1/messages", stub.gotURL)
	}
	if strings.Contains(stub.gotURL, "/api/mcp/") {
		t.Errorf("URL %q hits the z.ai MCP server, not the direct Anthropic endpoint", stub.gotURL)
	}
}

func TestGLMAudit_FailOpenOnHTTPError(t *testing.T) {
	stub := &stubGLMDoer{err: errBoom}
	withGLMSeams(t, "k", stub)

	res, _ := handleGLMAudit(context.Background(), mcp.CallToolRequest{})
	if res.IsError {
		t.Fatalf("transport error must fail-open to a structured result, not IsError")
	}
	got := decodeReview(t, res)
	if got.Verdict != VerdictInconclusive {
		t.Errorf("verdict = %q, want %q (fail-open)", got.Verdict, VerdictInconclusive)
	}
}

func TestGLMAudit_FailOpenOnMalformedResponse(t *testing.T) {
	stub := &stubGLMDoer{body: "<<<not-json>>>"}
	withGLMSeams(t, "k", stub)

	res, _ := handleGLMAudit(context.Background(), mcp.CallToolRequest{})
	got := decodeReview(t, res)
	if got.Verdict != VerdictInconclusive {
		t.Errorf("verdict = %q, want %q (malformed ⇒ fail-open)", got.Verdict, VerdictInconclusive)
	}
}

func TestGLMAudit_FailOpenOnUnauthenticatedStatus(t *testing.T) {
	// AC-MCP-014: an unauthenticated GLM (401) is a missing/unauthenticated
	// auditor ⇒ VerdictInconclusive ⇒ claude fallback, NOT a hard block.
	stub := &stubGLMDoer{status: http.StatusUnauthorized, body: "{}"}
	withGLMSeams(t, "k", stub)

	res, _ := handleGLMAudit(context.Background(), mcp.CallToolRequest{})
	got := decodeReview(t, res)
	if got.Verdict != VerdictInconclusive {
		t.Errorf("verdict = %q, want %q on 401 (unauthenticated fail-open)", got.Verdict, VerdictInconclusive)
	}
}

// decodeReview extracts the ReviewOutput a handler result carries in its first
// text content block. Test-only helper.
func decodeReview(t *testing.T, res *mcp.CallToolResult) ReviewOutput {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatal("no content blocks")
	}
	tc, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content[0] type = %T", res.Content[0])
	}
	var got ReviewOutput
	if err := json.Unmarshal([]byte(tc.Text), &got); err != nil {
		t.Fatalf("result not a ReviewOutput JSON: %v (text=%q)", err, tc.Text)
	}
	return got
}

var errBoom = &boomErr{}

type boomErr struct{}

func (boomErr) Error() string { return "boom: simulated transport failure" }

func TestResolveGLMAuditModel_SSOT(t *testing.T) {
	// AC-MCP-015: model resolution MUST go through template.ResolveAgentModelEffort
	// (the SSOT), never read agent frontmatter / llm.agent_overrides directly.
	// With no llm.yaml, the resolver returns the documented GLM default.
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	old := projectDirResolver
	projectDirResolver = func() string { return "" } // no sections dir available
	t.Cleanup(func() { projectDirResolver = old })

	m := resolveGLMAuditModel()
	if m == "" {
		t.Fatal("resolveGLMAuditModel returned empty for a missing llm.yaml (want the GLM default)")
	}
	if m == "opus" || strings.HasPrefix(m, "claude") {
		t.Errorf("resolveGLMAuditModel = %q; a Claude id cannot be a GLM default", m)
	}
}

func TestGLMAudit_NoAskUserQuestion(t *testing.T) {
	// AC-MCP-016 / C-HRA-008: the glm_audit handler package MUST NOT invoke
	// AskUserQuestion. Verified structurally — see also mcp_audit_test.go's
	// package-wide grep.
	// (This test is a placeholder for the grep-driven guard; the real guard
	// lives in mcp_audit_test.go::TestMCPAudit_NoAskUserQuestion so a single
	// assertion covers all M3 audit files.)
	t.Skip("structural guard lives in mcp_audit_test.go package-wide grep")
}
