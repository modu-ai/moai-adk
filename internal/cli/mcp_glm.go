// Package cli — GLM audit backend (SPEC-MOAI-MCP-SERVER-001 M3, REQ-MCP-009).
//
// mcp_glm.go implements the `glm_audit` MCP tool. It calls the z.ai GLM API
// DIRECTLY at the Anthropic-compatible endpoint
// (https://api.z.ai/api/anthropic/v1/messages) — reusing the SAME credential
// (loadGLMKey → ~/.moai/.env.glm) and endpoint (config.DefaultGLMBaseURL) the
// GLM backend uses elsewhere. It is NOT the z.ai MCP server and NOT any z.ai
// gateway (REQ-MCP-009, design.md §3 M3).
//
// The result adopts the shared review-output.schema.json shape (progress.md
// §G.4) — the SAME ReviewOutput/Finding/VerdictInconclusive types the codex
// backend (mcp_codex.go) uses — so orchestrator translation is uniform
// (design.md DQ-2). Fail-open is mandatory (C2 / REQ-MCP-012): a missing or
// unauthenticated GLM yields a structured VerdictInconclusive ReviewOutput and
// the workflow falls back to the active auditor (claude) WITHOUT hard-blocking.
//
// It NEVER invokes AskUserQuestion (subagent boundary, REQ-MCP-014).
//
// @MX:SPEC: SPEC-MOAI-MCP-SERVER-001
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// GLM audit domain constants (§14 hardcoding prevention — domain identifiers +
// endpoint path live here alongside the backend, like the codex identifiers in
// mcp_codex.go; the 120s HTTP timeout lives below as a named constant rather
// than in defaults.go because it is a single-call ceiling, not a cross-package
// threshold).
const (
	// glmAuditAgentKey is the profile-matrix agent key used to resolve the
	// audit model + effort via the SSOT (template.ResolveAgentModelEffort,
	// REQ-MCP-013). "sync-auditor" is the auditor-shaped key present in the
	// matrix; the GLM backend reuses it so model selection goes through the
	// single interpreter rather than a forked read.
	glmAuditAgentKey = "sync-auditor"

	// glmAuditDefaultModel is the fallback GLM model id when the SSOT returns
	// mapped=false (no llm.yaml) or the resolved model is a Claude id the
	// z.ai endpoint cannot serve (non-GLM session). Named constant per §14.
	glmAuditDefaultModel = "glm-4.6"

	// glmMessagesPath is appended to config.DefaultGLMBaseURL to form the
	// Anthropic-compatible /v1/messages endpoint (z.ai accepts Anthropic
	// Messages API requests at this path).
	glmMessagesPath = "/v1/messages"

	// glmAnthropicVersion is the Anthropic API version header the z.ai
	// Anthropic-compatible surface expects.
	glmAnthropicVersion = "2023-06-01"

	// glmAuditMaxTokens bounds a single audit response. An audit is a focused
	// review pass, not an open-ended generation.
	glmAuditMaxTokens = 4096

	// glmAuditHTTPTimeout is the per-call HTTP ceiling. Bounded so a hung z.ai
	// connection cannot stall the MCP tool indefinitely (fail-open on timeout).
	glmAuditHTTPTimeout = 120 * time.Second
)

// glmHTTPDoer abstracts the POST so tests inject a canned response without any
// network. The production doer is an *http.Client with glmAuditHTTPTimeout.
// Mirrors the M2 codexCommandRunner injectable-seam pattern.
type glmHTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// glmHTTPClient is the HTTP seam. Tests swap it (with t.Cleanup to restore).
var glmHTTPClient glmHTTPDoer = &http.Client{Timeout: glmAuditHTTPTimeout}

// glmKeyLoader is the credential seam (wraps loadGLMKey so tests simulate a
// missing vs present key without touching ~/.moai/.env.glm or the env).
var glmKeyLoader = loadGLMKey

// projectDirResolver is the project-root seam used by resolveGLMAuditModel to
// locate .moai/config/sections/llm.yaml. Wraps resolveProjectDir for testability.
var projectDirResolver = resolveProjectDir

// glmMessagesRequest is the Anthropic-compatible Messages API request body
// posted to z.ai.
type glmMessagesRequest struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	System    string       `json:"system"`
	Messages  []glmMessage `json:"messages"`
}

type glmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// glmMessagesResponse is the Anthropic-compatible response envelope. Only the
// text content block is consumed; the audit prompt constrains the model to
// emit a JSON ReviewOutput there.
type glmMessagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// resolveGLMAuditModel resolves the GLM audit model id via the model/effort
// SSOT (template.ResolveAgentModelEffort, REQ-MCP-013 / AC-MCP-015) ONLY. It
// NEVER reads agent frontmatter or llm.agent_overrides directly.
//
// Resolution rule:
//  1. Load llm.yaml through the SAME loadLLMSectionOnly helper the launcher uses.
//  2. Resolve via ResolveAgentModelEffort(glmAuditAgentKey).
//  3. If the session backend is GLM (template.IsGLMBackend) and the SSOT returned
//     a non-empty mapped model, use it (the matrix carries a GLM model id).
//  4. Otherwise fall back to glmAuditDefaultModel — a Claude id cannot be served
//     by the z.ai endpoint, and a missing llm.yaml has nothing to resolve.
func resolveGLMAuditModel() string {
	sectionsDir := filepath.Join(projectDirResolver(), ".moai", "config", "sections")
	llm, err := loadLLMSectionOnly(sectionsDir)
	if err != nil {
		return glmAuditDefaultModel
	}
	me, mapped := template.ResolveAgentModelEffort(llm, glmAuditAgentKey)
	if !mapped || me.Model == "" {
		return glmAuditDefaultModel
	}
	if template.IsGLMBackend(llm) {
		return me.Model // GLM session ⇒ a GLM model id from the matrix.
	}
	// Non-GLM session: the SSOT returned a Claude id the z.ai endpoint cannot
	// serve. Fall back to the canonical GLM audit model so the tool is still
	// callable directly (the caller opted into glm_audit explicitly).
	return glmAuditDefaultModel
}

// handleGLMAudit is the thin-wrapper handler for the `glm_audit` MCP tool. It
// calls the z.ai GLM API directly, returns a review-output.schema.json-shaped
// result (the shared ReviewOutput), and fails open (VerdictInconclusive →
// claude fallback) on a missing/unauthenticated/erroring/malformed GLM.
//
// It NEVER invokes AskUserQuestion (subagent boundary, REQ-MCP-014).
func handleGLMAudit(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	key := glmKeyLoader()
	if key == "" {
		// Fail-open (C2 / REQ-MCP-012): a missing optional dependency must not
		// hard-block. The workflow falls back to the active auditor (claude).
		return reviewToolResult(glmInconclusive("GLM API key not configured (~/.moai/.env.glm)")), nil
	}

	model := req.GetString("model", "")
	if model == "" {
		model = resolveGLMAuditModel() // SSOT (REQ-MCP-013)
	}
	focus := req.GetString("focus", "")
	token := extractProgressToken(req)

	notifyMCPProgress(ctx, token, 0, "glm 감사 시작 — z.ai 요청 준비 중...")
	out := callGLMAudit(ctx, key, model, focus, token)
	return reviewToolResult(out), nil
}

// callGLMAudit posts the audit prompt to z.ai and parses the response into a
// ReviewOutput. Every error path fails open to a VerdictInconclusive — the
// caller always receives a usable structured result.
func callGLMAudit(ctx context.Context, key, model, focus string, token mcp.ProgressToken) ReviewOutput {
	body, err := json.Marshal(glmMessagesRequest{
		Model:     model,
		MaxTokens: glmAuditMaxTokens,
		System:    glmAuditSystemPrompt(),
		Messages:  []glmMessage{{Role: "user", Content: glmAuditUserPrompt(focus)}},
	})
	if err != nil {
		return glmInconclusive("cannot build z.ai request: " + err.Error())
	}

	url := config.DefaultGLMBaseURL + glmMessagesPath
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return glmInconclusive("cannot build z.ai request: " + err.Error())
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", key)
	httpReq.Header.Set("anthropic-version", glmAnthropicVersion)

	notifyMCPProgress(ctx, token, 0.1, "z.ai에 감사 요청 전송 중...")
	resp, err := glmHTTPClient.Do(httpReq)
	if err != nil {
		return glmInconclusive("z.ai request failed: " + err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Unauthenticated (401/403) / rate-limited / server error ⇒ fail-open.
		// AC-MCP-014: an unauthenticated GLM is a missing/unauthenticated auditor.
		return glmInconclusive(fmt.Sprintf("z.ai returned HTTP %d", resp.StatusCode))
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return glmInconclusive("cannot read z.ai response: " + err.Error())
	}
	notifyMCPProgress(ctx, token, 0.8, "z.ai 응답 수신 — ReviewOutput 파싱 중...")
	return parseGLMReview(raw)
}

// parseGLMReview extracts the ReviewOutput the GLM model emitted in its first
// text content block. The audit prompt constrains the model to JSON; a malformed
// or empty-verdict response fails open to VerdictInconclusive (never a hard error).
func parseGLMReview(raw []byte) ReviewOutput {
	var env glmMessagesResponse
	if err := json.Unmarshal(raw, &env); err != nil {
		return glmInconclusive("malformed z.ai response: " + err.Error())
	}
	if len(env.Content) == 0 || env.Content[0].Text == "" {
		return glmInconclusive("z.ai response carried no content")
	}
	var out ReviewOutput
	if err := json.Unmarshal([]byte(env.Content[0].Text), &out); err != nil {
		return glmInconclusive("z.ai content was not a ReviewOutput JSON: " + err.Error())
	}
	if strings.TrimSpace(out.Verdict) == "" {
		return glmInconclusive("z.ai response carried no verdict")
	}
	if out.Findings == nil {
		out.Findings = []Finding{} // normalize empty to [] for §G.4 uniformity
	}
	if out.NextSteps == nil {
		out.NextSteps = []string{}
	}
	return out
}

// glmAuditSystemPrompt constrains the GLM model to emit ONLY a ReviewOutput
// JSON object, so the result is schema-uniform (§G.4) and parseable. Defensive
// output-validation per the LLM-security reference (LLM05 improper output
// handling): the model's free-form text is schema-validated before use, never
// acted on raw.
func glmAuditSystemPrompt() string {
	return "You are a meticulous code reviewer. Respond with ONLY a single JSON " +
		"object (no prose, no code fences) matching exactly this schema: " +
		`{"verdict":"pass|fail|inconclusive","summary":"<string>",` +
		`"findings":[{"severity":"high|medium|low","title":"<string>","body":"<string>",` +
		`"file":"<string>","line":0,"confidence":0.0,"recommendation":"<string>"}],` +
		`"next_steps":["<string>"]}. ` +
		"verdict 'pass' means no blocking findings; 'fail' means at least one " +
		"blocking finding; 'inconclusive' means you could not determine. Omit " +
		"findings you cannot substantiate. Be concrete about file/line."
}

// glmAuditUserPrompt builds the per-audit user turn. An optional focus narrows
// the review scope (e.g. "concurrency", "auth", "secret handling").
func glmAuditUserPrompt(focus string) string {
	prompt := "Review the proposed change for security flaws (injection, auth " +
		"bypass, secret leakage), correctness bugs (edge cases, error handling, " +
		"concurrency), and scope creep. Return the ReviewOutput JSON."
	if strings.TrimSpace(focus) != "" {
		prompt += " Focus area: " + strings.TrimSpace(focus) + "."
	}
	return prompt
}

// glmInconclusive is the GLM-backend fail-open ReviewOutput factory: it ALWAYS
// returns a usable structured result carrying the reason + the claude-fallback
// next step. Sibling of mcp_codex.go's inconclusiveReview.
func glmInconclusive(reason string) ReviewOutput {
	return ReviewOutput{
		Verdict:   VerdictInconclusive,
		Summary:   "glm unavailable: " + reason,
		Findings:  []Finding{},
		NextSteps: []string{"fall back to the active auditor (claude)"},
	}
}

// reviewToolResult shapes a ReviewOutput as the schema-typed MCP result. Shared
// with the codex backend's codexReviewToolResult shape so both backends emit
// identical result structure (uniform orchestrator translation, §G.4).
func reviewToolResult(out ReviewOutput) *mcp.CallToolResult {
	res, err := mcp.NewToolResultJSON(out)
	if err != nil {
		// Marshal of a plain struct cannot fail; degrade to text rather than
		// losing the verdict.
		return mcp.NewToolResultText("glm_audit: " + out.Verdict + " (" + out.Summary + ")")
	}
	return res
}
