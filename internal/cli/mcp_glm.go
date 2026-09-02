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
	//
	// DERIVED from the tier default rather than restated as its own literal. A
	// second literal drifts: this fallback sat on a two-generation-old id while
	// the tier defaults moved on, and the non-GLM path — a Claude session calling
	// glm_audit for a cross-model second opinion, which is the common case — got
	// that stale model every time. Deriving keeps the fallback on whatever the
	// launcher actually injects.
	glmAuditDefaultModel = config.DefaultGLMHigh

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

// projectDirResolver is the project-root seam used by the GLM audit resolver to
// locate .moai/config/sections/llm.yaml. Wraps resolveProjectDir for testability.
var projectDirResolver = resolveProjectDir

// glmMessagesRequest is the Anthropic-compatible Messages API request body
// posted to z.ai.
type glmMessagesRequest struct {
	Model string `json:"model"`
	// ReasoningEffort is the audit path's reasoning directive
	// (SPEC-V3R6-AUDIT-MODEL-PIN-001 REQ-AMP-007): the top-level z.ai
	// reasoning_effort control — the delivery field SELECTED BY LIVE EVIDENCE
	// (AC-AMP-006's first differential ran hypothesis A, the Anthropic-style
	// thinking object, and measured budget_tokens IGNORED — output tokens
	// 3667 vs 3480 under budgets 3072 vs 1024, ratio 1.02; measured by card
	// t225's AC-AMP-006 differential). The state name is transmitted
	// VERBATIM; empty omits the field.
	ReasoningEffort string       `json:"reasoning_effort,omitempty"`
	MaxTokens       int          `json:"max_tokens"`
	System          string       `json:"system"`
	Messages        []glmMessage `json:"messages"`
}

// glmAuditReasoningEffort validates the pinned effort under REQ-AMP-006's
// single-reading rule: the valid set is EXACTLY {low, high, max}
// (template.GLMState* — the z.ai state names, stored and transmitted
// verbatim); any other non-empty value returns "" — the reasoning directive
// is omitted while the model pin still applies. NO effort collapse runs on
// this path (CollapseClaudeEffortToGLM is the vocabulary reference only, not
// a runtime dependency here).
func glmAuditReasoningEffort(effort string) string {
	switch effort {
	case template.GLMStateLow, template.GLMStateHigh, template.GLMStateMax:
		return effort
	default:
		return ""
	}
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

// resolveGLMAuditModelEffort resolves the GLM audit {model, effort} pair
// (SPEC-V3R6-AUDIT-MODEL-PIN-001 REQ-AMP-003). Precedence:
//
//  1. The workflow.audit.glm pin — a non-empty pin model returns the pair
//     VERBATIM, bypassing the IsGLMBackend session check: a pin is
//     by-construction a GLM id, and a wrong id degrades via the existing
//     z.ai-4xx fail-open to VerdictInconclusive (design decision D3), never a
//     hard error. Effort rides the pin only when the model is pinned (the
//     model is the gate — effort alone pins nothing).
//  2. Otherwise the legacy SSOT resolution: the sync-auditor cell through
//     resolveGLMModelForAgent, with an EMPTY effort (the pre-SPEC body carried
//     no reasoning field, and the SSOT effort is Claude-vocabulary — never
//     transmittable under the single-reading rule).
//
// It NEVER reads agent frontmatter or llm.agent_overrides directly (REQ-MCP-013
// / AC-MCP-015), and glm_task never calls it (REQ-AMP-008).
//
// projectRoot names the tree being reviewed (the project_root doctrine,
// .claude/rules/moai/core/moai-mcp-tools.md): a worktree session MUST pass its
// own root, because projectDirResolver() names the primary checkout and would
// read the pin from a DIFFERENT tree than the diff under review — the same
// caller-named-root contract the codex counterpart honors via params["cwd"].
// Empty falls back to projectDirResolver() (the pre-CR behavior).
func resolveGLMAuditModelEffort(projectRoot string) config.ModelEffort {
	root := strings.TrimSpace(projectRoot)
	if root == "" {
		root = projectDirResolver()
	}
	if pin := workflowAuditPins(root).GLM; pin.Model != "" {
		return pin
	}
	return config.ModelEffort{Model: resolveGLMModelForAgent(glmAuditAgentKey)}
}

// resolveGLMModelForAgent resolves a GLM model id for the given profile-matrix
// agent key via the model/effort SSOT (template.ResolveAgentModelEffort). It is
// the shared body behind the GLM audit resolver and the glm_task resolver: the
// audit path keys on the auditor-shaped cell, the task path on its consumer
// (super-advisor), and the resolution rule is otherwise identical.
//
// Resolution rule:
//  1. Load llm.yaml through the SAME loadLLMSectionOnly helper the launcher uses.
//  2. Resolve via ResolveAgentModelEffort(agentKey).
//  3. If the session backend is GLM (template.IsGLMBackend) and the SSOT returned
//     a non-empty mapped model, use it (the matrix carries a GLM model id).
//  4. Otherwise fall back to glmAuditDefaultModel — a Claude id cannot be served
//     by the z.ai endpoint, and a missing llm.yaml has nothing to resolve.
func resolveGLMModelForAgent(agentKey string) string {
	sectionsDir := filepath.Join(projectDirResolver(), ".moai", "config", "sections")
	llm, err := loadLLMSectionOnly(sectionsDir)
	if err != nil {
		return glmAuditDefaultModel
	}
	me, mapped := template.ResolveAgentModelEffort(llm, agentKey)
	if !mapped || me.Model == "" {
		return glmAuditDefaultModel
	}
	if template.IsGLMBackend(llm) {
		return me.Model // GLM session ⇒ a GLM model id from the matrix.
	}
	// Non-GLM session: the SSOT returned a Claude id the z.ai endpoint cannot
	// serve. Fall back to the canonical GLM model so the tool is still callable
	// directly (the caller opted into the GLM tool explicitly).
	return glmAuditDefaultModel
}

// handleGLMAudit is the thin-wrapper handler for the `glm_audit` MCP tool. It
// calls the z.ai GLM API directly, returns a review-output.schema.json-shaped
// result (the shared ReviewOutput), and fails open (VerdictInconclusive →
// claude fallback) on a missing/unauthenticated/erroring/malformed GLM.
//
// It NEVER invokes AskUserQuestion (subagent boundary, REQ-MCP-014).
func handleGLMAudit(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// The tree comes from resolveToolProjectRoot — the same caller-named-root
	// contract codex_audit has (SPEC-MCP-WORKTREE-ROOT-001): a worktree session
	// MUST be able to name its own tree, because the server's own resolution
	// names the primary checkout. Absent ⇒ resolveProjectDir(), exactly what
	// this call did before the parameter existed. An unusable path is a tool
	// error, not the fail-open verdict — fail-open covers a broken GLM, not a
	// caller input the caller can correct. Resolved BEFORE the key check so a
	// caller-input error keeps its precedence (codex_audit resolves the same
	// way before touching its backend).
	root, rootErr := resolveToolProjectRoot(req)
	if rootErr != nil {
		return toolErr("glm_audit", rootErr), nil
	}

	// Build identity is assembled ONCE here (REQ-ABI-007) and rides every
	// result this handler returns below, fail-open exits included — an
	// inconclusive verdict is still a verdict that must name its binary.
	buildCommit, buildLag := auditBuildIdentity(ctx, root)
	// review shapes every glm_audit result this handler returns, identity
	// fields attached — one assembly point for the whole handler.
	review := func(out ReviewOutput) *mcp.CallToolResult {
		out.BuildCommit, out.BuildLag = buildCommit, buildLag
		return reviewToolResult(out)
	}

	key := glmKeyLoader()
	if key == "" {
		// Fail-open (C2 / REQ-MCP-012): a missing optional dependency must not
		// hard-block. The workflow falls back to the active auditor (claude).
		return review(glmInconclusive("GLM API key not configured (~/.moai/.env.glm)")), nil
	}

	focus := req.GetString("focus", "")
	target := req.GetString("target", codexTargetUncommitted)
	token := extractProgressToken(req)

	notifyMCPProgress(ctx, token, 0, "glm 감사 시작 — 리뷰 대상 diff 수집 중...")
	// GLM cannot read the tree, so the change has to be collected here and
	// carried in the request. No diff ⇒ no review: fail open to inconclusive
	// rather than ask z.ai for a verdict on code it will never see (card t178).
	//
	// Resolved BELOW the root (CR #8): the pin must come from the SAME tree as
	// the diff under review — a worktree session names its own root, and
	// resolving through projectDirResolver() here could read a different
	// tree's workflow.yaml.
	me := resolveGLMAuditModelEffort(root) // pin > SSOT (REQ-AMP-003)
	if explicit := req.GetString("model", ""); strings.TrimSpace(explicit) != "" {
		me.Model = strings.TrimSpace(explicit) // explicit caller model outranks the pin
	}
	diff, err := collectReviewDiff(root, target)
	if err != nil {
		return review(glmInconclusive("no reviewable change: " + err.Error())), nil
	}
	if strings.TrimSpace(diff) == "" {
		return review(glmInconclusive("no reviewable change: target " + target + " produced an empty diff")), nil
	}

	notifyMCPProgress(ctx, token, 0.05, "glm 감사 — z.ai 요청 준비 중...")
	out := callGLMAudit(ctx, key, me.Model, me.Effort, focus, diff, token)
	return review(out), nil
}

// callGLMAudit posts the audit prompt to z.ai and parses the response into a
// ReviewOutput. Every error path fails open to a VerdictInconclusive — the
// caller always receives a usable structured result.
//
// effort is the pinned z.ai reasoning state (low|high|max, single reading per
// REQ-AMP-006); an invalid/empty value omits the reasoning directive while the
// model still applies. Whether the delivered field is honored by the endpoint
// is arbitrated by the AC-AMP-006 live differential, not assumed.
func callGLMAudit(ctx context.Context, key, model, effort, focus, diff string, token mcp.ProgressToken) ReviewOutput {
	body, err := json.Marshal(glmMessagesRequest{
		Model:           model,
		ReasoningEffort: glmAuditReasoningEffort(effort),
		MaxTokens:       glmAuditMaxTokens,
		System:          glmAuditSystemPrompt(),
		Messages:        []glmMessage{{Role: "user", Content: glmAuditUserPrompt(focus, diff)}},
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
// TEXT content block — not the first block: under a delivered reasoning
// directive z.ai prefixes the response with thinking content blocks (whose
// payload lives in `thinking`, not `text`), and reading Content[0] blindly
// failed open to inconclusive while a full review sat in the next block
// (live-gate finding, SPEC-V3R6-AUDIT-MODEL-PIN-001 M5). The audit prompt
// constrains the model to JSON; a malformed or empty-verdict response fails
// open to VerdictInconclusive (never a hard error).
func parseGLMReview(raw []byte) ReviewOutput {
	var env glmMessagesResponse
	if err := json.Unmarshal(raw, &env); err != nil {
		return glmInconclusive("malformed z.ai response: " + err.Error())
	}
	text := ""
	for i := range env.Content {
		// An empty Type is a text block too: the Anthropic-compatible shapes
		// this parser sees include envelopes whose text blocks carry NO type
		// field at all (TestParseGLMReview_StripsMarkdownFence's
		// production-envelope fixture). Thinking/reasoning blocks DO carry a
		// type, so the skip semantics hold.
		if (env.Content[i].Type == "" || env.Content[i].Type == "text") && env.Content[i].Text != "" {
			text = env.Content[i].Text
			break
		}
	}
	if text == "" {
		return glmInconclusive("z.ai response carried no text content")
	}
	var out ReviewOutput
	// z.ai occasionally wraps the JSON in a markdown code fence despite the
	// prompt's "no code fences" constraint; strip fences + surrounding prose so
	// the Unmarshal sees a bare object.
	jsonBody := extractJSONObject(text)
	if err := json.Unmarshal([]byte(jsonBody), &out); err != nil {
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

// extractJSONObject strips a leading markdown code fence and any surrounding
// prose from a model's text response, returning the substring spanning the
// first '{' to the last '}'. z.ai occasionally wraps the JSON object in
// ```json ... ``` despite the prompt's "no code fences" constraint; this
// recovers the bare object so json.Unmarshal succeeds. Returns the input
// unchanged when no brace boundary is found (the caller's Unmarshal then
// reports the original error — no silent masking).
func extractJSONObject(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// Drop the opening fence line (``` or ```json).
		if nl := strings.IndexByte(s, '\n'); nl >= 0 {
			s = s[nl+1:]
		}
		s = strings.TrimSpace(s)
		// Drop a trailing closing fence.
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	start := strings.IndexByte(s, '{')
	end := strings.LastIndexByte(s, '}')
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
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
//
// diff is the change under review, and it is NOT optional in practice: GLM is an
// HTTPS call with no filesystem, so a prompt without it asks the model to review
// code it has no way to see, and the model answers anyway (card t178). Callers
// fail open to inconclusive before reaching here when no diff could be produced;
// the closing instruction below is the second line of defence for the case where
// one arrives empty regardless.
func glmAuditUserPrompt(focus, diff string) string {
	prompt := "Review the proposed change for security flaws (injection, auth " +
		"bypass, secret leakage), correctness bugs (edge cases, error handling, " +
		"concurrency), and scope creep. Return the ReviewOutput JSON."
	if strings.TrimSpace(focus) != "" {
		prompt += " Focus area: " + strings.TrimSpace(focus) + "."
	}
	prompt += "\n\nGround every finding in the diff below — cite the file and line " +
		"it appears at. Report nothing you cannot point to in this diff; if it " +
		"does not let you judge, return verdict 'inconclusive'.\n\n" +
		"--- BEGIN DIFF ---\n" + diff + "\n--- END DIFF ---\n"
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
