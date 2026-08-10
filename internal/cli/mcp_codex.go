// Package cli — codex audit backend (SPEC-MOAI-MCP-SERVER-001 M2).
//
// mcp_codex.go implements the two Phase-1 codex tools registered alongside
// the M1 core surface: `codex_audit` (REQ-MCP-006 / AC-MCP-007) and
// `codex_setup` (REQ-MCP-007 / AC-MCP-008). It is a pure Go reimplementation
// — there is NO Node bridge of any kind (design.md §3 M2).
//
// The codex binary is OPTIONAL and experimental (R1): it may be absent from
// PATH or its `app-server` JSON-RPC surface may differ from the documented
// `review/start` / `turn/start` methods. Every code path therefore fails
// OPEN: a missing / erroring / malformed codex yields a structured
// VerdictInconclusive result, never a hard error and never a panic
// (REQ-MCP-012 preview; design.md §4.3 fail-open state machine). The full
// 3-way claude-fallback plumbing lands in M3; M2 only guarantees that the
// codex_audit tool + the review-gate hook never hard-crash on a missing codex.
//
// @MX:SPEC: SPEC-MOAI-MCP-SERVER-001
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"
)

// codex domain constants (§14 hardcoding prevention — domain identifiers live
// here alongside the backend, like moaiMCPServerName in mcp_server.go; the
// 900s review-gate timeout lives in internal/config/defaults.go as a threshold).
const (
	// codexBinaryName is the PATH-resolved codex binary (NO absolute path; §14).
	codexBinaryName = "codex"
	// codexAppServerSubcmd is the codex subcommand that speaks JSON-RPC over stdio.
	codexAppServerSubcmd = "app-server"

	// codex JSON-RPC methods (report §3.4 documented surface).
	codexMethodReviewStart = "review/start" // native audit mode
	codexMethodTurnStart   = "turn/start"   // adversarial audit mode (carries the adversarial-review prompt)

	// codex_audit mode enum (AC-MCP-007).
	codexModeNative      = "native"
	codexModeAdversarial = "adversarial"

	// codex_audit target enum (what codex reviews).
	codexTargetUncommitted = "uncommittedChanges"
	codexTargetBaseBranch  = "baseBranch"

	// codex_setup auth-provider classification tokens.
	codexAuthChatGPT  = "ChatGPT"
	codexAuthAPIKey   = "apiKey"
	codexAuthProvider = "provider"
	codexAuthUnknown  = "unknown"
)

// VerdictInconclusive is the fail-open verdict value. It rides the same
// `verdict` field as codex's native verdicts, as a STRUCTURED (non-error)
// result, never a hard error (design.md §4.3 / §G.4).
const VerdictInconclusive = "inconclusive"

// ReviewOutput is the adopted review-output.schema.json shape (progress.md
// §G.4 locked). It is shared across the codex (codex_audit) and — in M3 — the
// GLM backend, so orchestrator translation is uniform (design.md DQ-2).
type ReviewOutput struct {
	Verdict   string    `json:"verdict"`
	Summary   string    `json:"summary"`
	Findings  []Finding `json:"findings"`
	NextSteps []string  `json:"next_steps"`
}

// Finding is a single review finding (§G.4).
type Finding struct {
	Severity       string  `json:"severity"`
	Title          string  `json:"title"`
	Body           string  `json:"body"`
	File           string  `json:"file"`
	Line           int     `json:"line"`
	Confidence     float64 `json:"confidence"`
	Recommendation string  `json:"recommendation"`
}

// inconclusiveReview is the fail-open ReviewOutput factory: it ALWAYS returns a
// usable structured result carrying the reason + the claude-fallback next step.
// The full 3-way fallback plumbing is M3; M2 only guarantees no hard crash.
func inconclusiveReview(reason string) ReviewOutput {
	return ReviewOutput{
		Verdict:   VerdictInconclusive,
		Summary:   "codex unavailable: " + reason,
		Findings:  []Finding{},
		NextSteps: []string{"fall back to the active auditor (claude)"},
	}
}

// ─── injectable command-execution seams (cross-platform testable, no PATH stubs) ───

// codexCommandRunner abstracts shelling out to the codex binary so tests inject
// canned responses without spawning a process. The production runner uses
// exec.CommandContext; tests swap codexRunner (with t.Cleanup to restore).
type codexCommandRunner interface {
	run(ctx context.Context, binaryPath string, args []string, stdin string) (stdout string, err error)
}

// realCodexRunner is the production runner: spawns the binary, pipes stdin,
// captures stdout. A bounded context deadline is the caller's responsibility
// (the MCP tool call inherits the request ctx; the review-gate pins 900s).
type realCodexRunner struct{}

func (realCodexRunner) run(ctx context.Context, binaryPath string, args []string, stdin string) (string, error) {
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Stdin = strings.NewReader(stdin)
	var out bytes.Buffer
	cmd.Stdout = &out
	// Discard stderr so a noisy codex never corrupts the JSON-RPC stdout parse.
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

// codexRunner is the command-execution seam. Tests swap it (with t.Cleanup).
var codexRunner codexCommandRunner = realCodexRunner{}

// codexLookPath is the binary-resolution seam (wraps exec.LookPath so tests
// simulate a missing codex without PATH manipulation).
var codexLookPath = exec.LookPath

// ─── low-level codex JSON-RPC client ───

// codexRPCEnvelope is the newline-delimited JSON-RPC response envelope the
// app-server returns — either `{"id":N,"result":<ReviewOutput>}` on success or
// `{"id":N,"error":{"code":C,"message":M}}` on a protocol-level rejection.
//
// The error arm is load-bearing, not decoration. Without it a rejected request
// decodes into a zero-valued Result, which is indistinguishable from "the
// reviewer had no opinion" — so a protocol mismatch reads as an inconclusive
// review and the gate fail-opens silently. An operator who turned the gate on
// then sees nothing happen, with nothing anywhere saying why.
type codexRPCEnvelope struct {
	Result ReviewOutput   `json:"result"`
	Error  *codexRPCError `json:"error"`
}

// codexRPCError is the JSON-RPC error object. Both fields are surfaced verbatim
// so the operator sees the server's own words rather than our paraphrase.
type codexRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// runCodexReviewRPC shells out to the codex binary's app-server JSON-RPC mode
// and invokes a single review method. It returns the parsed ReviewOutput on
// success. On ANY error (spawn failure, non-zero exit, deadline, malformed
// response) it returns an inconclusive ReviewOutput AND the error, so the
// caller always has a fail-open-usable struct while still seeing the cause.
//
// The NDJSON (newline-delimited) framing is the documented app-server shape; if
// codex's actual framing differs, the malformed-response path fails open
// (REQ-MCP-012 preview). The request is `{"jsonrpc":"2.0","method":<m>,"params":<p>,"id":1}`.
func runCodexReviewRPC(ctx context.Context, binaryPath, method string, params map[string]any) (ReviewOutput, error) {
	reqBytes, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	})
	if err != nil {
		return inconclusiveReview("cannot marshal codex request: " + err.Error()), err
	}
	stdout, runErr := codexRunner.run(ctx, binaryPath, []string{codexAppServerSubcmd}, string(reqBytes)+"\n")
	if runErr != nil {
		return inconclusiveReview("codex invocation failed: " + runErr.Error()), runErr
	}
	var env codexRPCEnvelope
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		return inconclusiveReview("malformed codex response: " + err.Error()), err
	}
	if env.Error != nil {
		reason := fmt.Sprintf("codex rejected the request (JSON-RPC %d): %s", env.Error.Code, env.Error.Message)
		return inconclusiveReview(reason), errors.New(reason)
	}
	if env.Result.Verdict == "" {
		return inconclusiveReview("codex response carried no verdict"), fmt.Errorf("codex response carried no verdict")
	}
	return env.Result, nil
}

// codexAdversarialReviewPrompt builds the adversarial-review prompt text the
// adversarial mode sends to codex turn/start (design.md §3 M2 / report §3.4).
// Generic + focused: a red-team security + correctness review of the change.
func codexAdversarialReviewPrompt(focus string) string {
	prompt := "Perform an adversarial code review of the proposed change. " +
		"Hunt for security flaws (injection, auth bypass, secret leakage, unsafe " +
		"destructive operations), correctness bugs (edge cases, error handling, " +
		"concurrency), and scope creep. Report concrete findings with severity, " +
		"file/line, confidence, and a recommendation. If the change is sound, " +
		"return verdict pass with an empty findings list."
	if focus != "" {
		prompt += " Focus area: " + focus + "."
	}
	return prompt
}

// ─── codex_audit handler (AC-MCP-007) ───

// handleCodexAudit is the thin-wrapper handler for the `codex_audit` MCP tool.
// mode=native → codex review/start; mode=adversarial → codex turn/start +
// the adversarial-review prompt. Both return a review-output.schema.json-shaped
// result. fail-open on missing / erroring / malformed codex (VerdictInconclusive).
//
// It NEVER invokes AskUserQuestion (subagent boundary, REQ-MCP-014): a missing
// or inconclusive codex is a structured result, surfaced for the orchestrator
// to translate through its own user-interaction channel.
func handleCodexAudit(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mode := req.GetString("mode", codexModeNative)
	target := req.GetString("target", codexTargetUncommitted)
	focus := req.GetString("focus", "")
	model := req.GetString("model", "")

	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		return codexReviewToolResult(inconclusiveReview("codex binary not found in PATH")), nil
	}

	method := codexMethodReviewStart
	params := map[string]any{
		"target": target,
		"model":  model,
	}
	if mode == codexModeAdversarial {
		method = codexMethodTurnStart
		params["prompt"] = codexAdversarialReviewPrompt(focus)
		if focus != "" {
			params["focus"] = focus
		}
	}

	out, _ := runCodexReviewRPC(ctx, binaryPath, method, params) // fail-open inside
	return codexReviewToolResult(out), nil
}

// codexReviewToolResult shapes a ReviewOutput as the schema-typed MCP result
// (mcp.NewToolResultJSON carries the JSON in TextContent AND the typed struct
// in StructuredContent, so a caller can read either).
func codexReviewToolResult(out ReviewOutput) *mcp.CallToolResult {
	res, err := mcp.NewToolResultJSON(out)
	if err != nil {
		// Marshal of a plain struct cannot fail; if it ever does, degrade to a
		// text result rather than losing the verdict.
		return mcp.NewToolResultText("codex_audit: " + out.Verdict + " (" + out.Summary + ")")
	}
	return res
}

// ─── codex_setup handler (AC-MCP-008) ───

// handleCodexSetup is the thin-wrapper handler for the `codex_setup` MCP tool.
// It is a pure Go probe — exec.LookPath("codex") + codex --version + an
// auth-provider classification + the current enable_review_gate config state.
// NO Node bridge of any kind (REQ-MCP-007). Read-only: it REPORTS the toggle
// state; mutating it is a heavier, wizard-owned concern (M4).
func handleCodexSetup(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	binaryPath, err := codexLookPath(codexBinaryName)
	installed := err == nil && binaryPath != ""
	result := map[string]any{
		"installed":          installed,
		"auth_provider":      codexAuthUnknown,
		"enable_review_gate": readCodexReviewGateEnabled(resolveProjectDir()),
		"node_bridge":        false, // explicit: REQ-MCP-007 Go-only reimplementation
	}
	if !installed {
		return toolJSON("codex_setup", result), nil
	}
	result["binary"] = binaryPath

	// codex --version (best-effort; failure leaves version empty, not an error).
	if ver, vErr := codexRunner.run(ctx, binaryPath, []string{"--version"}, ""); vErr == nil {
		result["version"] = strings.TrimSpace(ver)
	}

	// Auth-provider classification (heuristic — codex's auth surface is
	// undocumented per R1; fail-open to "unknown" on any uncertainty).
	result["auth_provider"] = classifyCodexAuth(ctx, binaryPath)

	return toolJSON("codex_setup", result), nil
}

// classifyCodexAuth probes codex's auth state and maps it to one of the
// ChatGPT / apiKey / provider / unknown tokens (AC-MCP-008). It runs
// `codex login status` and pattern-matches the output; any error or
// non-matching output degrades to codexAuthUnknown (fail-open, R1).
func classifyCodexAuth(ctx context.Context, binaryPath string) string {
	out, err := codexRunner.run(ctx, binaryPath, []string{"login", "status"}, "")
	if err != nil || out == "" {
		return codexAuthUnknown
	}
	low := strings.ToLower(out)
	switch {
	case strings.Contains(low, "chatgpt"):
		return codexAuthChatGPT
	case strings.Contains(low, "api key"), strings.Contains(low, "apikey"):
		return codexAuthAPIKey
	case strings.Contains(low, "provider"):
		return codexAuthProvider
	default:
		return codexAuthUnknown
	}
}

// ─── config gate reader (shared by codex_setup + the review-gate subcommand) ───

// readCodexReviewGateEnabled reads workflow.codex.review_gate.enabled from
// `.moai/config/sections/workflow.yaml`. Truth table (fail-CLOSED — the opt-in
// default is OFF, matching the HOI precedent isHookOptInEnabled, NOT the
// fail-open learning gate):
//
//   - file missing / unreadable                   → false (default disabled)
//   - YAML parse error                            → false (default disabled)
//   - `workflow.codex` block absent               → false (Go zero-value default)
//   - `workflow.codex.review_gate.enabled: false` → false
//   - `workflow.codex.review_gate.enabled: true`  → true
//
// The key path is NESTED under the file's `workflow:` root — the shape
// config.Loader.loadWorkflowSection expects (config.workflowFileWrapper) and
// the shape the shipped template deploys. An earlier revision unmarshalled
// `codex:` at the TOP level, which no deployed file ever carries, so the toggle
// could never read true. The flat form is NOT accepted as an alias: one
// spelling only.
//
// This stays a small hand-rolled read rather than a config.Loader.Load call on
// purpose. The gate runs at every turn-end, so it should touch exactly one file
// and no defaults/env-override machinery; and each failure path here returns
// false explicitly, whereas Load returns populated defaults plus a nil error
// and would blur "unreadable" into "configured off". The shape is pinned
// against the real loader by TestReviewGateReaders_AgreeWithConfigLoader.
//
// The distributed default is false (C6 / AC-MCP-010 opt-in). A maintainer opts
// in via local config; the template NEVER carries `enabled: true` (§25).
func readCodexReviewGateEnabled(projectDir string) bool {
	if projectDir == "" {
		return false
	}
	configPath := filepath.Join(projectDir, ".moai", "config", "sections", "workflow.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}
	var doc struct {
		Workflow struct {
			Codex struct {
				ReviewGate struct {
					Enabled bool `yaml:"enabled"`
				} `yaml:"review_gate"`
			} `yaml:"codex"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false
	}
	return doc.Workflow.Codex.ReviewGate.Enabled
}
