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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

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

	// codex JSON-RPC methods. review/start (native audit) and turn/start
	// (adversarial prompt) are the review methods; initialize + thread/start are
	// the mandatory session handshake the app-server (codex-cli 0.146.1) requires
	// before either will answer (see runCodexReviewRPC doc).
	codexMethodReviewStart = "review/start" // native audit mode
	codexMethodTurnStart   = "turn/start"   // adversarial audit mode (carries the adversarial-review prompt)
	codexMethodInitialize  = "initialize"   // mandatory handshake opener
	codexMethodThreadStart = "thread/start" // opens a thread; its result.thread.id is the review threadId

	// codex client identity sent in the initialize handshake (the server rejects
	// initialize without a clientInfo {name,version} object).
	codexClientName    = "moai-codex-gate"
	codexClientVersion = "0.1.0"

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

// codexCommandRunner abstracts a SINGLE-SHOT shelling out to the codex binary
// for the codex_setup / auth probes (`codex --version`, `codex login status`),
// which are not JSON-RPC calls. The interactive review-gate session uses the
// separate codexSession seam below. Tests inject canned responses without
// spawning a process; the production runner uses exec.CommandContext.
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

// codexRunner is the single-shot command-execution seam (codex_setup / auth).
// Tests swap it (with t.Cleanup).
var codexRunner codexCommandRunner = realCodexRunner{}

// codexLookPath is the binary-resolution seam (wraps exec.LookPath so tests
// simulate a missing codex without PATH manipulation).
var codexLookPath = exec.LookPath

// ─── low-level codex app-server JSON-RPC session client ───
//
// codex app-server speaks an async, session-oriented JSON-RPC over stdio
// (verified against codex-cli 0.146.1). Four protocol facts shape this client
// (each was the cause of a prior structural ALLOW — the gate could not reach a
// BLOCK even on an obvious injection + hardcoded-AWS-key change):
//
//  1. initialize handshake FIRST. A bare review/start gets no response at all;
//     initialize with a clientInfo {name,version} object is the required opener.
//  2. threadId is required by review/start and turn/start. It is obtained from
//     thread/start's result (result.thread.id) and threaded into the review call.
//  3. target is an INTERNALLY TAGGED OBJECT {"type":"uncommittedChanges"}, not a
//     bare string. A bare string is rejected with JSON-RPC -32600.
//  4. Responses are asynchronous NDJSON. Notifications (no id) and the final
//     result (with id) arrive as LATER lines on stdout; the verdict itself is
//     NOT in the review/start ack (which only says the turn started) — it
//     arrives later as an `exitedReviewMode` item followed by `turn/completed`.
//     The client keeps the subprocess's stdin/stdout open and reads NDJSON lines
//     until the turn completes, then synthesizes a verdict from codex's review
//     prose.

// codexConn is the interactive JSON-RPC session handle: send writes one NDJSON
// request line to the server's stdin; recv returns the next NDJSON line the
// server wrote to stdout (response OR notification), or (_, false) on EOF.
type codexConn interface {
	send(line string) error
	recv() (string, bool)
	close() error
}

// codexSessionRunner abstracts spawning the codex app-server subprocess so tests
// inject a canned line exchange without spawning a process. The production
// runner (realCodexSessionRunner) pipes stdin/stdout and reads stdout
// concurrently into a buffered line channel.
type codexSessionRunner interface {
	start(ctx context.Context, binaryPath string, args []string) (codexConn, error)
}

// codexSession is the session-spawning seam. Tests swap it (with t.Cleanup).
var codexSession codexSessionRunner = realCodexSessionRunner{}

// realCodexSessionRunner is the production session runner.
type realCodexSessionRunner struct{}

func (realCodexSessionRunner) start(ctx context.Context, binaryPath string, args []string) (codexConn, error) {
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("codex stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("codex stdout pipe: %w", err)
	}
	// Discard stderr so a noisy codex never corrupts the stdout NDJSON parse.
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("codex start: %w", err)
	}
	c := &realCodexConn{cmd: cmd, stdin: stdin, stdout: stdout, lines: make(chan string, 128)}
	go c.readLoop(ctx)
	return c, nil
}

// realCodexConn pipes a single codex app-server subprocess. readLoop drains
// stdout line-by-line into the lines channel; closing stdin + killing the
// process on close prevents a hung codex from stalling the Stop hook.
type realCodexConn struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.Reader
	lines  chan string
	mu     sync.Mutex
}

func (c *realCodexConn) readLoop(ctx context.Context) {
	defer close(c.lines)
	sc := bufio.NewScanner(c.stdout)
	// Review turns can carry large patches in aggregatedOutput; raise the cap.
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		select {
		case c.lines <- sc.Text():
		case <-ctx.Done():
			return
		}
	}
	// scanner stopped (EOF or read error) — the channel close signals recv EOF.
}

func (c *realCodexConn) send(line string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := fmt.Fprintln(c.stdin, line); err != nil {
		return fmt.Errorf("codex write: %w", err)
	}
	return nil
}

func (c *realCodexConn) recv() (string, bool) {
	line, ok := <-c.lines
	return line, ok
}

func (c *realCodexConn) close() error {
	_ = c.stdin.Close()
	done := make(chan error, 1)
	go func() { done <- c.cmd.Wait() }()
	select {
	case err := <-done:
		return err
	case <-time.After(3 * time.Second):
		_ = c.cmd.Process.Kill()
		return <-done
	}
}

// codexRPCError is the JSON-RPC error object. Both fields are surfaced verbatim
// so the operator sees the server's own words rather than our paraphrase. The
// arm is load-bearing: without it a rejected request reads as "no opinion" and
// the gate fail-opens silently with nothing anywhere saying why (the #1421 fix
// preserves this surface).
type codexRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// rpcMessage is a permissive NDJSON decoder that captures the fields the
// session driver needs: id (request vs notification discriminator), method
// (notification name), result (response payload), error (rejection arm), and
// params (notification payload). Lines that fail to parse are skipped as noise.
type rpcMessage struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *codexRPCError  `json:"error,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
}

// runCodexReviewRPC drives a full codex app-server JSON-RPC session: the
// initialize handshake → thread/start (obtain threadId) → the caller's review
// or turn method with the corrected request shape → await the asynchronous turn
// result → synthesize a ReviewOutput from codex's review prose. On ANY error
// (spawn failure, deadline, JSON-RPC rejection, no verdict text) it returns an
// inconclusive ReviewOutput AND the error, so the caller always has a
// fail-open-usable struct while still seeing the cause (fail-open is the gate's
// invariant; the #1421 error-surfacing invariant is preserved by the error arm
// in awaitResponse).
//
// The caller passes the FINAL review/turn method; the handshake (initialize +
// thread/start) is owned by this function. params may carry "cwd" (string),
// which is threaded into thread/start so codex reviews the right working tree.
func runCodexReviewRPC(ctx context.Context, binaryPath, method string, params map[string]any) (ReviewOutput, error) {
	conn, err := codexSession.start(ctx, binaryPath, []string{codexAppServerSubcmd})
	if err != nil {
		return inconclusiveReview("codex session start failed: " + err.Error()), err
	}
	defer func() { _ = conn.close() }()

	// 1. initialize handshake (mandatory; clientInfo {name,version} required).
	const initID = 1
	if err := writeCodexRequest(conn, initID, codexMethodInitialize, map[string]any{
		"clientInfo": map[string]any{"name": codexClientName, "version": codexClientVersion},
	}); err != nil {
		return inconclusiveReview("codex initialize write failed: " + err.Error()), err
	}
	if _, err := awaitCodexResponse(conn, initID, ctx); err != nil {
		return inconclusiveReview("codex initialize rejected: " + err.Error()), err
	}

	// 2. thread/start → obtain threadId (required by review/start + turn/start).
	const threadIDReq = 2
	threadParams := map[string]any{}
	if cwd, ok := params["cwd"].(string); ok && cwd != "" {
		threadParams["cwd"] = cwd
	}
	if err := writeCodexRequest(conn, threadIDReq, codexMethodThreadStart, threadParams); err != nil {
		return inconclusiveReview("codex thread/start write failed: " + err.Error()), err
	}
	thrResp, err := awaitCodexResponse(conn, threadIDReq, ctx)
	if err != nil {
		return inconclusiveReview("codex thread/start rejected: " + err.Error()), err
	}
	threadID := extractThreadID(thrResp.Result)
	if threadID == "" {
		return inconclusiveReview("codex thread/start returned no thread id"), errors.New("codex thread/start: no thread id in result")
	}

	// 3. caller's review/turn method with the corrected request shape.
	const reviewID = 3
	finalParams := buildCodexReviewParams(method, params, threadID)
	if err := writeCodexRequest(conn, reviewID, method, finalParams); err != nil {
		return inconclusiveReview("codex " + method + " write failed: " + err.Error()), err
	}
	// The review/start ack only confirms the turn started; the verdict comes later.
	if _, err := awaitCodexResponse(conn, reviewID, ctx); err != nil {
		return inconclusiveReview("codex " + method + " rejected: " + err.Error()), err
	}

	// 4. await the asynchronous turn completion and synthesize the verdict.
	reviewText := awaitCodexTurnReview(conn, threadID, ctx)
	if reviewText == "" {
		return inconclusiveReview("codex review produced no verdict text"), errors.New("codex review produced no verdict text")
	}
	return synthesizeReviewOutput(reviewText), nil
}

// writeCodexRequest marshals and sends one JSON-RPC request line.
func writeCodexRequest(conn codexConn, id int, method string, params map[string]any) error {
	req := map[string]any{"jsonrpc": "2.0", "method": method, "params": params, "id": id}
	b, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return conn.send(string(b))
}

// awaitCodexResponse reads NDJSON lines until one carries the wanted id,
// surfacing a JSON-RPC error arm verbatim (the #1421 invariant). Notification
// lines (no id) and lines for other ids are consumed and discarded.
func awaitCodexResponse(conn codexConn, wantID int, ctx context.Context) (rpcMessage, error) {
	for {
		if err := ctx.Err(); err != nil {
			return rpcMessage{}, err
		}
		line, ok := conn.recv()
		if !ok {
			return rpcMessage{}, fmt.Errorf("codex stdout closed before response to id=%d", wantID)
		}
		var msg rpcMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue // unparseable line (stray stderr leak / noise) — skip
		}
		if !codexIDMatches(msg.ID, wantID) {
			continue // a notification or a different request's response
		}
		if msg.Error != nil {
			return msg, fmt.Errorf("codex rejected the request (JSON-RPC %d): %s", msg.Error.Code, msg.Error.Message)
		}
		return msg, nil
	}
}

// codexIDMatches reports whether a JSON-RPC id (integer or string) equals want.
func codexIDMatches(raw json.RawMessage, want int) bool {
	if len(raw) == 0 {
		return false
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n == want
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s == strconv.Itoa(want)
	}
	return false
}

// extractThreadID reads result.thread.id from a thread/start response.
func extractThreadID(result json.RawMessage) string {
	if len(result) == 0 {
		return ""
	}
	var doc struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err := json.Unmarshal(result, &doc); err != nil {
		return ""
	}
	return doc.Thread.ID
}

// buildCodexReviewParams shapes the final review/turn request params with the
// corrected protocol envelope: threadId is always added; review/start's target
// is coerced to the internally-tagged object; turn/start's prompt is wrapped in
// the input array the UserInput schema requires.
func buildCodexReviewParams(method string, params map[string]any, threadID string) map[string]any {
	out := map[string]any{"threadId": threadID}
	switch method {
	case codexMethodReviewStart:
		out["target"] = coerceCodexReviewTarget(params["target"])
	case codexMethodTurnStart:
		prompt, _ := params["prompt"].(string)
		if strings.TrimSpace(prompt) == "" {
			prompt = codexAdversarialReviewPrompt("")
		}
		out["input"] = []map[string]any{{"type": "text", "text": prompt}}
	}
	return out
}

// coerceCodexReviewTarget normalizes the target value to the internally-tagged
// object shape codex requires ({"type":"uncommittedChanges"}). A bare string
// ("uncommittedChanges" / "baseBranch") is lifted into the object; an
// already-correct object is passed through; anything else defaults to reviewing
// uncommitted changes.
func coerceCodexReviewTarget(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		if _, has := m["type"]; has {
			return m
		}
	}
	if s, ok := v.(string); ok && s != "" {
		return map[string]any{"type": s}
	}
	return map[string]any{"type": codexTargetUncommitted}
}

// awaitCodexTurnReview reads notifications until turn/completed for threadID,
// collecting the review prose from the exitedReviewMode item (preferred) or the
// final agentMessage text (fallback). Returns "" on EOF / deadline (fail-open).
func awaitCodexTurnReview(conn codexConn, threadID string, ctx context.Context) string {
	reviewText, agentText := "", ""
	for {
		if err := ctx.Err(); err != nil {
			return bestCodexReviewText(reviewText, agentText)
		}
		line, ok := conn.recv()
		if !ok {
			return bestCodexReviewText(reviewText, agentText)
		}
		var msg rpcMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		if msg.Method == "" { // a response, not a notification
			continue
		}
		if codexThreadIDOf(msg.Params) != threadID {
			continue
		}
		switch msg.Method {
		case "item/completed":
			var p struct {
				Item struct {
					Type   string `json:"type"`
					Review string `json:"review"`
					Text   string `json:"text"`
				} `json:"item"`
			}
			_ = json.Unmarshal(msg.Params, &p)
			if p.Item.Type == "exitedReviewMode" && p.Item.Review != "" {
				reviewText = p.Item.Review
			}
			if p.Item.Type == "agentMessage" && p.Item.Text != "" {
				agentText = p.Item.Text
			}
		case "turn/completed":
			return bestCodexReviewText(reviewText, agentText)
		}
	}
}

// bestCodexReviewText prefers the structured exitedReviewMode review over the
// free-form agentMessage text (both carry the same prose in practice).
func bestCodexReviewText(review, agent string) string {
	if review != "" {
		return review
	}
	return agent
}

// codexThreadIDOf reads threadId from a notification's params.
func codexThreadIDOf(params json.RawMessage) string {
	if len(params) == 0 {
		return ""
	}
	var p struct {
		ThreadID string `json:"threadId"`
	}
	_ = json.Unmarshal(params, &p)
	return p.ThreadID
}

// codexFindingBullet matches codex review finding bullets, which carry a
// severity tag in brackets (e.g. "- [P1] Avoid ..."). Verified against
// codex-cli 0.146.1's review-mode output on BOTH a clean change (no bullets ⇒
// pass: "introduces no identifiable correctness or blocking issues") and an
// injection+secret change ("- [P1] ..." bullets ⇒ fail).
var codexFindingBullet = regexp.MustCompile(`(?m)^\s*[-*]\s+\[[A-Za-z]+\d+\]`)

// synthesizeReviewOutput maps codex's review prose into the review-output
// schema. codex does NOT return a structured verdict enum — it returns free-form
// prose whose presence of severity-tagged finding bullets (codex's own format)
// signals a failure; a clean review (no finding bullets) is a pass. The summary
// carries the verbatim review text so the operator sees codex's own words.
func synthesizeReviewOutput(reviewText string) ReviewOutput {
	verdict := "pass"
	if codexFindingBullet.MatchString(reviewText) {
		verdict = "fail"
	}
	return ReviewOutput{
		Verdict:   verdict,
		Summary:   strings.TrimSpace(reviewText),
		Findings:  []Finding{},
		NextSteps: []string{},
	}
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
		"cwd":    resolveProjectDir(),
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
