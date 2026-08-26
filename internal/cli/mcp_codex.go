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

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
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

	// codexMethodThreadResume re-opens an EXISTING thread by id, loading it from
	// codex's own on-disk store. It is what resume_last uses instead of
	// thread/start (REQ-CX2-008). ThreadResumeParams requires only {threadId}
	// and its response carries the same {thread:{id}} shape thread/start returns,
	// so extractThreadID reads both (codex-cli 0.146.1 generate-json-schema).
	codexMethodThreadResume = "thread/resume"

	// codexMethodTurnInterrupt cancels an in-flight turn. Its params are
	// {threadId, turnId} with BOTH required (M0 probe, progress.md §E.2 (a)) —
	// which is why the job record carries a turnId at all (REQ-CX2-003), and why
	// a job whose turn/started was never observed cannot be interrupted and
	// degrades to process termination instead (REQ-CX2-011).
	codexMethodTurnInterrupt = "turn/interrupt"

	// codexNotifyTurnStarted is the server→client notification announcing that a
	// turn began. Its turn.id is the ONLY source of the turnId that
	// turn/interrupt requires alongside threadId (REQ-CX2-003; M0 probe,
	// progress.md §E.2 (a)).
	codexNotifyTurnStarted = "turn/started"

	// codexRequestFileChangeApproval is a server→CLIENT request: it carries an
	// `id` and the turn does not advance until a response with that id comes
	// back. A live session (codex-cli 0.146.1) parked at
	// activeFlags:["waitingOnApproval"] and never returned because this driver
	// used to read the line and drop it (REQ-CX2-016; progress.md §E.2 § Live
	// protocol verification).
	codexRequestFileChangeApproval = "item/fileChange/requestApproval"

	// codexApprovalDecisionDecline is the answer this client always gives.
	//
	// The value is READ FROM the generated protocol schema, not inferred from
	// the request transcript (plan.md §F AP-9): FileChangeRequestApprovalResponse
	// requires a `decision` field whose FileChangeApprovalDecision variants are
	// accept | acceptForSession | decline | cancel. `decline` is documented as
	// "User denied the file changes. The agent will continue the turn"; `cancel`
	// denies AND interrupts the turn. Declining is the denial that still lets the
	// turn finish and report what it could not do, which is the outcome the
	// caller asked for.
	//
	// It is deliberately unconditional. REQ-CX2-016 adds no approval-GRANTING
	// path — workflow.codex.task.allow_write is the sole write opt-in
	// (REQ-CX2-007), and a turn that did opt in already runs workspaceWrite, so
	// an approval request reaching this driver is by construction one that must
	// not be granted (plan.md §F AP-7).
	codexApprovalDecisionDecline = "decline"

	// codexRPCMethodNotFound is the JSON-RPC 2.0 "Method not found" code. An
	// unrecognized client-bound request gets this rather than silence, so an
	// unknown request also unblocks the turn instead of stalling it
	// (REQ-CX2-016).
	codexRPCMethodNotFound = -32601

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

	// codex_setup auth-provider classification tokens. The spellings are the
	// wire tokens the web console maps to display labels — they are lowercase
	// on both sides so the two surfaces agree (internal/web/codex_state.go).
	// The former "provider" token is gone: the two-stage ladder maps only the
	// grammar's captured terms (chatgpt / api key) and treats a provider
	// phrase outside the grammar as unknown rather than guessing
	// (SPEC-CODEX-LAUNCHER-001, plan §C.1). internal/web keeps its own mirror
	// constant for legacy display mapping.
	codexAuthChatGPT = "chatgpt"
	codexAuthAPIKey  = "apiKey"
	codexAuthUnknown = "unknown"

	// codex sandbox-policy variants (REQ-CX2-007). SandboxPolicy is an
	// INTERNALLY-TAGGED UNION object — {"type":"readOnly"} — not a bare string,
	// exactly like `target` (codex-cli 0.146.1 generate-json-schema, .definitions
	// .SandboxPolicy: a oneOf of four objects each requiring `type`). A bare
	// string is rejected with JSON-RPC -32600, which is how the `target` shape
	// was originally learned. Only the two variants this SPEC transmits are named
	// here; dangerFullAccess and externalSandbox are deliberately unreachable.
	codexSandboxReadOnly       = "readOnly"
	codexSandboxWorkspaceWrite = "workspaceWrite"

	// codexAuditAgentKey is the profile-matrix agent key the codex backend
	// resolves its model + effort through (REQ-CX2-002). It is the SAME
	// auditor-shaped key the GLM sibling uses (mcp_glm.go glmAuditAgentKey), so
	// both backends read the single interpreter rather than forking a lookup.
	codexAuditAgentKey = "sync-auditor"
)

// codexServableModelPrefixes are the model-id families the codex app-server can
// actually serve (§14 — the families live here as a named constant rather than
// inline). The profile matrix is Claude-centric: its default cell for
// codexAuditAgentKey is {opus, high}, and handing "opus" to codex would break
// the review gate for every project that never opted in. A resolved model
// outside these families is therefore dropped, leaving the request byte-identical
// to the pre-M1 shape (C7 no-regression). This mirrors the GLM sibling, which
// filters its own SSOT result through IsGLMBackend before using it.
var codexServableModelPrefixes = []string{"gpt-", "o1", "o3", "o4", "codex"}

// codexServableModel reports whether a model id can plausibly be served by the
// codex app-server. An empty id is never servable (the field is then omitted so
// codex applies its own configured default).
func codexServableModel(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	if m == "" {
		return false
	}
	for _, p := range codexServableModelPrefixes {
		if strings.HasPrefix(m, p) {
			return true
		}
	}
	return false
}

// codexSSOTModelEffort resolves the codex model + effort through the model/effort
// SSOT (template.ResolveAgentModelEffort, REQ-CX2-002) ONLY — it NEVER reads
// agent frontmatter or the per-agent override map directly (C4; the negative
// guard is TestMCPAudit_NoDirectFrontmatterRead, the positive one is
// TestCodexSession_ResolvedModelReachesTransmittedParams).
//
// The cell is returned whole or not at all: when the resolved model is not
// codex-servable the paired effort is dropped with it, because an effort value
// from another backend's vocabulary is no more transmittable than its model id
// (ReasoningEffort is documented as "a non-empty reasoning effort value
// advertised by the model").
//
// projectDir is the tree being reviewed (the review gate passes the hook's
// project root, which need not equal the server's own cwd); an empty value falls
// back to the resolver seam.
func codexSSOTModelEffort(projectDir string) config.ModelEffort {
	if strings.TrimSpace(projectDir) == "" {
		projectDir = projectDirResolver()
	}
	llm, err := loadLLMSectionOnly(filepath.Join(projectDir, ".moai", "config", "sections"))
	if err != nil {
		return config.ModelEffort{}
	}
	me, mapped := template.ResolveAgentModelEffort(llm, codexAuditAgentKey)
	if !mapped || !codexServableModel(me.Model) {
		return config.ModelEffort{}
	}
	return me
}

// resolveCodexModelEffort resolves the model + effort for one codex request. An
// explicit caller-supplied `model` wins over the SSOT-resolved value and is sent
// verbatim — the caller opted into it deliberately, so the servability filter
// (which exists to protect callers who did NOT choose) does not apply.
func resolveCodexModelEffort(params map[string]any) config.ModelEffort {
	cwd, _ := params["cwd"].(string)
	me := codexSSOTModelEffort(cwd)
	if explicit, ok := params["model"].(string); ok && strings.TrimSpace(explicit) != "" {
		me.Model = strings.TrimSpace(explicit)
	}
	return me
}

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

// codexProcessConn is the OPTIONAL half of codexConn: a connection backed by a
// real OS process reports the pid of the process the session runner spawned. It
// is separate from codexConn so the canned test connections stay processless —
// a job record must never name a pid this server did not spawn (REQ-CX2-012).
type codexProcessConn interface {
	pid() int
}

// codexConnPID reports the pid of the codex process behind conn, or 0 when the
// connection has no backing process. Zero is the "no process of ours" value the
// job registry records and the cancel path (M4) refuses to signal.
func codexConnPID(conn codexConn) int {
	if p, ok := conn.(codexProcessConn); ok {
		return p.pid()
	}
	return 0
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

// pid reports the pid of the codex subprocess this connection spawned, or 0
// once the process is gone (satisfies codexProcessConn).
func (c *realCodexConn) pid() int {
	if c.cmd == nil || c.cmd.Process == nil {
		return 0
	}
	return c.cmd.Process.Pid
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

// codexSessionHandle is a REUSABLE codex app-server session (REQ-CX2-001): it
// retains the threadId obtained from thread/start so a second turn can be issued
// on the same thread without repeating the initialize + thread/start handshake.
//
// Before M1 the handshake was inlined in runCodexReviewRPC, which opened a
// session, drove exactly one turn, and tore the subprocess down on return — the
// threadId was a local variable discarded on return (spec.md §A.3 G1). Splitting
// it out changes nothing for the two existing consumers (runCodexReviewRPC is
// retained below as a thin caller) and gives the later job surface a session it
// can hold across turns.
//
// The handle owns the connection once openCodexSession returns successfully; the
// caller MUST close it.
type codexSessionHandle struct {
	conn     codexConn
	threadID string

	// mu guards nextID. A background job's goroutine allocates ids for its turn
	// while the M4 cancel path allocates one for turn/interrupt on the SAME
	// session from the calling goroutine, so the counter is genuinely shared.
	mu     sync.Mutex
	nextID int // monotonic JSON-RPC request id; the handshake consumed 1 and 2

	// turnID is the id of the most recent turn the session observed starting,
	// read from the turn/started notification's turn.id. It is the second
	// argument turn/interrupt requires alongside threadId (REQ-CX2-003; M0
	// probe, progress.md §E.2 (a)) — without it a turn cannot be addressed for
	// cancellation at all.
	turnID string

	// onTurnStarted, when non-nil, is invoked with the turn id the moment
	// turn/started is observed — MID-FLIGHT, before the turn completes. It is
	// the hook a background job goroutine installs so the turnId reaches the
	// job record while the turn is still running and therefore still
	// cancellable (plan.md §D M2).
	onTurnStarted func(turnID string)
}

// codexSessionError carries the fail-open summary text alongside the underlying
// cause. The handshake's failure paths each have their own operator-facing
// summary; returning them through this type lets openCodexSession be split out
// while runCodexReviewRPC keeps emitting the exact same ReviewOutput.Summary and
// the exact same error value it did before.
type codexSessionError struct {
	summary string
	cause   error
}

func (e *codexSessionError) Error() string { return e.cause.Error() }
func (e *codexSessionError) Unwrap() error { return e.cause }

// codexHandshakeFailure closes the half-open connection and wraps the cause with
// its fail-open summary.
func codexHandshakeFailure(conn codexConn, summary string, cause error) error {
	_ = conn.close()
	return &codexSessionError{summary: summary, cause: cause}
}

// openCodexSession spawns a codex app-server subprocess and completes the
// mandatory handshake — initialize (with clientInfo) → thread/start → threadId —
// returning a handle further turns can be issued on.
//
// params may carry "cwd" (string), threaded into thread/start so codex reviews
// the right working tree, and "model" (string). thread/start is the SESSION-level
// destination for the resolved model (REQ-CX2-002): ReviewStartParams declares
// only {delivery, target, threadId} — it has no model field at all — so for the
// review path the thread is the only place a model can reach codex.
func openCodexSession(ctx context.Context, binaryPath string, params map[string]any) (*codexSessionHandle, error) {
	return openCodexSessionOn(ctx, binaryPath, params, "")
}

// openCodexSessionOn is openCodexSession with an optional EXISTING thread to
// resume (REQ-CX2-008). When resumeThreadID is empty the handshake opens a new
// thread with thread/start — the unchanged path both existing callers take.
// When it is set, thread/resume re-opens that thread instead, so no new thread
// is started and the recorded id is the one every turn addresses.
func openCodexSessionOn(ctx context.Context, binaryPath string, params map[string]any, resumeThreadID string) (*codexSessionHandle, error) {
	conn, err := codexSession.start(ctx, binaryPath, []string{codexAppServerSubcmd})
	if err != nil {
		return nil, &codexSessionError{summary: "codex session start failed: " + err.Error(), cause: err}
	}

	// 1. initialize handshake (mandatory; clientInfo {name,version} required).
	const initID = 1
	if err := writeCodexRequest(conn, initID, codexMethodInitialize, map[string]any{
		"clientInfo": map[string]any{"name": codexClientName, "version": codexClientVersion},
	}); err != nil {
		return nil, codexHandshakeFailure(conn, "codex initialize write failed: "+err.Error(), err)
	}
	if _, err := awaitCodexResponse(conn, initID, ctx); err != nil {
		return nil, codexHandshakeFailure(conn, "codex initialize rejected: "+err.Error(), err)
	}

	// 2. thread/start (or thread/resume) → obtain threadId (required by
	//    review/start + turn/start).
	const threadIDReq = 2
	threadMethod := codexMethodThreadStart
	threadParams := map[string]any{}
	if resumeThreadID != "" {
		threadMethod = codexMethodThreadResume
		threadParams["threadId"] = resumeThreadID
	}
	if cwd, ok := params["cwd"].(string); ok && cwd != "" {
		threadParams["cwd"] = cwd
	}
	if me := resolveCodexModelEffort(params); me.Model != "" {
		threadParams["model"] = me.Model
	}
	if err := writeCodexRequest(conn, threadIDReq, threadMethod, threadParams); err != nil {
		return nil, codexHandshakeFailure(conn, "codex "+threadMethod+" write failed: "+err.Error(), err)
	}
	thrResp, err := awaitCodexResponse(conn, threadIDReq, ctx)
	if err != nil {
		return nil, codexHandshakeFailure(conn, "codex "+threadMethod+" rejected: "+err.Error(), err)
	}
	threadID := extractThreadID(thrResp.Result)
	if threadID == "" {
		return nil, codexHandshakeFailure(conn, "codex "+threadMethod+" returned no thread id",
			errors.New("codex "+threadMethod+": no thread id in result"))
	}

	return &codexSessionHandle{conn: conn, threadID: threadID, nextID: threadIDReq + 1}, nil
}

// close tears the session's subprocess down (stdin close, bounded wait, kill).
func (h *codexSessionHandle) close() error {
	if h == nil || h.conn == nil {
		return nil
	}
	return h.conn.close()
}

// pid reports the pid of the codex subprocess backing this session, or 0 when
// the session has no backing process (a canned session, or a closed one).
func (h *codexSessionHandle) pid() int {
	if h == nil || h.conn == nil {
		return 0
	}
	return codexConnPID(h.conn)
}

// noteTurnStarted records an observed turn id on the handle and fans it out to
// the mid-flight observer, if one is installed.
//
// The write is guarded because the turn loop no longer always finishes before
// its caller reads the id back: a turn abandoned at the REQ-CX2-017 bound keeps
// running in its own goroutine while codex_task builds a result from the
// handle. The observer fires OUTSIDE the lock — it writes a job record, which
// has no business happening under a session mutex.
func (h *codexSessionHandle) noteTurnStarted(turnID string) {
	if turnID == "" {
		return
	}
	h.mu.Lock()
	h.turnID = turnID
	observer := h.onTurnStarted
	h.mu.Unlock()

	if observer != nil {
		observer(turnID)
	}
}

// setTurnStartedObserver installs the mid-flight turn-id observer under the same
// lock noteTurnStarted reads it through, so the field has one discipline rather
// than one that happens to be safe because of where it is currently called.
func (h *codexSessionHandle) setTurnStartedObserver(fn func(string)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onTurnStarted = fn
}

// currentTurnID reports the turn id observed so far, or "" when no turn/started
// has arrived. It is the ONLY safe way to read the id from outside the turn
// loop (see noteTurnStarted).
func (h *codexSessionHandle) currentTurnID() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.turnID
}

// allocID hands out the next JSON-RPC request id. Ids must stay unique across
// the whole session: a reused id would let awaitCodexResponse match an earlier
// turn's response and return it as this turn's ack.
func (h *codexSessionHandle) allocID() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	id := h.nextID
	h.nextID++
	return id
}

// sendTurnInterrupt sends turn/interrupt for the given thread and turn
// (REQ-CX2-011). Both arguments are required by the method, so the caller must
// have a recorded turnId — an empty one is a caller error rather than a
// degradation to be papered over here.
//
// The response is deliberately NOT awaited. The job's own goroutine is inside
// awaitCodexTurnReview draining this connection; a second reader would race it
// for lines and could swallow the turn/completed notification the turn loop is
// waiting for. What cancellation needs is the request to reach codex — the
// outcome is observed as the turn ending, which the caller waits for through the
// grace window instead.
func (h *codexSessionHandle) sendTurnInterrupt(threadID, turnID string) error {
	if threadID == "" || turnID == "" {
		return fmt.Errorf("%s requires both threadId and turnId (got %q / %q)",
			codexMethodTurnInterrupt, threadID, turnID)
	}
	return writeCodexRequest(h.conn, h.allocID(), codexMethodTurnInterrupt, map[string]any{
		"threadId": threadID,
		"turnId":   turnID,
	})
}

// runTurn drives ONE turn on the session's existing thread: send the caller's
// review/turn method with the corrected request shape → await the ack → await the
// asynchronous turn completion → synthesize a ReviewOutput from codex's review
// prose. On ANY error it returns an inconclusive ReviewOutput AND the error, so
// the caller always has a fail-open-usable struct while still seeing the cause
// (the error arm in awaitCodexResponse surfaces a JSON-RPC rejection verbatim).
func (h *codexSessionHandle) runTurn(ctx context.Context, method string, params map[string]any) (ReviewOutput, error) {
	id := h.allocID()
	finalParams := buildCodexReviewParams(method, params, h.threadID)
	if err := writeCodexRequest(h.conn, id, method, finalParams); err != nil {
		return inconclusiveReview("codex " + method + " write failed: " + err.Error()), err
	}
	// The review/start ack only confirms the turn started; the verdict comes later.
	if _, err := awaitCodexResponse(h.conn, id, ctx); err != nil {
		return inconclusiveReview("codex " + method + " rejected: " + err.Error()), err
	}

	reviewText, turnErr := awaitCodexTurnReview(h.conn, h.threadID, ctx, h.noteTurnStarted)
	if turnErr != nil {
		// The turn ENDED in a non-completed terminal state (failed / interrupted
		// — quota exhaustion, stream error, interruption). Whatever review text
		// was collected is codex's PLACEHOLDER ("Reviewer failed to output a
		// response."), not a verdict, so it must never reach the synthesizer: a
		// bullet-less placeholder there synthesizes "pass" and launders a review
		// that never happened into a clean one (card t52 / the exact hazard the
		// fail-open comment in HandleCodexReviewGate names).
		return inconclusiveReview(turnErr.Error()), turnErr
	}
	if reviewText == "" {
		return inconclusiveReview("codex review produced no verdict text"), errors.New("codex review produced no verdict text")
	}
	return synthesizeReviewOutput(reviewText), nil
}

// runCodexReviewRPC drives a full single-turn codex app-server JSON-RPC session:
// the handshake (openCodexSession) → one turn (runTurn) → tear-down. It is
// RETAINED as a thin caller over the reusable session so both existing consumers
// — handleCodexAudit and the Stop-hook gate HandleCodexReviewGate — keep their
// current behavior and signature semantics unchanged (C7 / plan.md §F AP-2: the
// gate's fail-open depends on the (ReviewOutput, error) pair where the output is
// always usable).
//
// The caller passes the FINAL review/turn method; the handshake is owned by
// openCodexSession.
//
// codexReviewRPC is the injectable seam over it, wired to runCodexReviewRPC in
// production. Both codex AUDIT paths — handleCodexAudit here and
// performCodexAudit in mcp_convergence.go — dispatch through it, so a test can
// assert what the params map actually carries with no live backend
// (SPEC-MCP-WORKTREE-ROOT-001 AC-1b). The Stop-hook review gate keeps calling
// runCodexReviewRPC directly: it is not an audit path and is out of scope here.
var codexReviewRPC = runCodexReviewRPC

func runCodexReviewRPC(ctx context.Context, binaryPath, method string, params map[string]any) (ReviewOutput, error) {
	sess, err := openCodexSession(ctx, binaryPath, params)
	if err != nil {
		var sErr *codexSessionError
		if errors.As(err, &sErr) {
			return inconclusiveReview(sErr.summary), sErr.cause
		}
		return inconclusiveReview(err.Error()), err
	}
	defer func() { _ = sess.close() }()

	return sess.runTurn(ctx, method, params)
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
// the input array the UserInput schema requires, and carries the resolved model
// + effort.
//
// The model/effort destination is method-specific, and NOT a matter of taste —
// it is what the protocol declares (codex-cli 0.146.1 `codex app-server
// generate-json-schema`):
//
//   - TurnStartParams declares BOTH `model` (string|null) and `effort`
//     (ReasoningEffort|null), so a turn carries them directly.
//   - ReviewStartParams declares ONLY {delivery, target, threadId} — neither
//     field exists. Injecting them would put unknown fields on the Stop-hook
//     gate's own request path, so review/start carries neither and the session
//     model rides thread/start instead (see openCodexSession).
//
// Before M1 this function built a fresh map that dropped the caller's `model`
// entirely, so the `model` parameter advertised by codex_audit never reached
// codex (spec.md §A.3 G3 — a live defect, not a missing feature).
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
		me := resolveCodexModelEffort(params)
		if me.Model != "" {
			out["model"] = me.Model
		}
		if me.Effort != "" {
			out["effort"] = me.Effort
		}
		// sandboxPolicy is forwarded ONLY when the caller supplied one, so the
		// codex_audit / review-gate request stays byte-identical to its pre-M3
		// shape (C7). codex_task supplies it on EVERY turn — see
		// codexSandboxPolicy for why the non-writing turn is the one that must.
		if policy, ok := params["sandboxPolicy"]; ok && policy != nil {
			out["sandboxPolicy"] = policy
		}
	}
	return out
}

// codexSandboxPolicy builds the internally-tagged SandboxPolicy object for a
// turn: workspaceWrite when the write opt-in is granted, readOnly otherwise.
//
// It is called for EVERY turn codex_task starts, including the ones that did
// not ask to write — and that is the whole point rather than a redundancy. The
// protocol documents sandboxPolicy as overriding the policy "for this turn AND
// SUBSEQUENT TURNS": the field is sticky on the thread, not scoped to one turn.
// Under resume_last thread reuse (REQ-CX2-008), omitting the field on a
// non-writing turn would let it inherit a write-enabled policy left behind by an
// earlier turn that DID opt in — a route around the allow_write gate, since the
// gate is read at request time while its effect outlives the request. Sending
// readOnly explicitly is what closes it (REQ-CX2-007; plan.md §D M3 hazard).
func codexSandboxPolicy(allowWrite bool) map[string]any {
	if allowWrite {
		return map[string]any{"type": codexSandboxWorkspaceWrite}
	}
	return map[string]any{"type": codexSandboxReadOnly}
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
//
// The second return value is the TURN FAILURE: non-nil exactly when the
// turn/completed terminal state is anything other than "completed" (the
// TurnStatus enum is completed | interrupted | failed | inProgress, and
// turn.error is documented as populated only when status is failed — codex-cli
// app-server generate-json-schema, verified on 0.147.0). A failed or
// interrupted turn never produced a verdict, so the caller must treat the
// collected text as a placeholder and fail open WITH the error surfaced rather
// than synthesize a verdict from it. Only the turn/completed state is
// authoritative: a bare `error` notification with willRetry semantics whose
// turn later completes is a REAL review (the error was transient).
//
// onTurnStarted, when non-nil, is invoked with the turn id carried by the
// turn/started notification — the ONLY source of the turnId that turn/interrupt
// requires (REQ-CX2-003). It fires mid-loop, so a background job persists the
// id while the turn is still running rather than after it completes.
func awaitCodexTurnReview(conn codexConn, threadID string, ctx context.Context, onTurnStarted func(string)) (string, error) {
	reviewText, agentText := "", ""
	for {
		if err := ctx.Err(); err != nil {
			return bestCodexReviewText(reviewText, agentText), nil
		}
		line, ok := conn.recv()
		if !ok {
			return bestCodexReviewText(reviewText, agentText), nil
		}
		var msg rpcMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		// A line carrying BOTH a method and an id is a server→client REQUEST, and
		// the turn stops until it is answered (REQ-CX2-016). It is handled before
		// the thread filter below on purpose: an unrecognized request need not
		// carry a threadId, and a request dropped for want of one stalls the turn
		// exactly as surely as one dropped for want of a case.
		if codexIsClientRequest(msg) {
			answerCodexClientRequest(conn, msg)
			continue
		}
		if msg.Method == "" { // a response, not a notification
			continue
		}
		if codexThreadIDOf(msg.Params) != threadID {
			continue
		}
		switch msg.Method {
		case codexNotifyTurnStarted:
			if onTurnStarted != nil {
				onTurnStarted(codexTurnIDOf(msg.Params))
			}
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
			if failure := codexTurnFailure(msg.Params); failure != "" {
				return bestCodexReviewText(reviewText, agentText),
					fmt.Errorf("codex review turn %s", failure)
			}
			return bestCodexReviewText(reviewText, agentText), nil
		}
	}
}

// codexTurnFailure reads the terminal turn state from a turn/completed
// notification's params and returns "" when the turn COMPLETED (a real verdict
// exists), otherwise a short human description carrying codex's own error
// message when one is present. Keys off turn.status (and turn.error) — the
// authoritative terminal signal — so transient `error` notifications that
// codex retried internally do not poison a review that did complete.
func codexTurnFailure(params json.RawMessage) string {
	if len(params) == 0 {
		return ""
	}
	var p struct {
		Turn struct {
			Status string `json:"status"`
			Error  *struct {
				Message string `json:"message"`
			} `json:"error"`
		} `json:"turn"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return ""
	}
	switch p.Turn.Status {
	case "completed", "":
		return "" // completed turn ("" = malformed/absent status — not our call)
	case "inProgress":
		return "" // a non-terminal state on a terminal notification — trust the notification
	}
	if p.Turn.Error != nil && strings.TrimSpace(p.Turn.Error.Message) != "" {
		return p.Turn.Status + ": " + p.Turn.Error.Message
	}
	return p.Turn.Status
}

// codexIsClientRequest reports whether a decoded line is a server→client
// REQUEST: a method AND an id. An explicit `"id":null` is a notification with a
// noisy encoder, not a request — answering it would address nobody.
func codexIsClientRequest(msg rpcMessage) bool {
	if msg.Method == "" || len(msg.ID) == 0 {
		return false
	}
	return strings.TrimSpace(string(msg.ID)) != "null"
}

// answerCodexClientRequest responds to a server→client request so the turn can
// advance (REQ-CX2-016). A recognized approval request is DENIED; anything else
// gets a JSON-RPC error arm, which unblocks the turn just as well as a result
// would and states plainly that this client does not implement the method.
//
// The write happens from inside the read loop while the turn's caller is
// blocked in it. That is the established shape here, not a new hazard:
// sendTurnInterrupt already writes from a second goroutine while the turn's own
// goroutine sits in this loop, and a live session exercised exactly that
// (progress.md §E.2 Item 2). No locking is added on the assumption it is
// needed.
//
// A write failure is deliberately swallowed: the connection is already broken,
// the turn will end by EOF, and there is no caller here to hand an error to.
func answerCodexClientRequest(conn codexConn, msg rpcMessage) {
	if msg.Method == codexRequestFileChangeApproval {
		_ = writeCodexResponse(conn, msg.ID, map[string]any{"decision": codexApprovalDecisionDecline})
		return
	}
	_ = writeCodexErrorResponse(conn, msg.ID, codexRPCMethodNotFound,
		"moai codex client does not implement "+msg.Method)
}

// writeCodexResponse sends one JSON-RPC RESULT response line, echoing the
// request's id verbatim. It is a second small writer rather than a reuse of
// writeCodexRequest, which builds a request envelope (method + params + id);
// a response carries id + result and no method.
//
// The id is echoed as raw JSON so an integer id stays an integer and a string
// id stays a string — the server matches on the value it sent, and codex sends
// integer ids from its own counter (id 0 was the observed first one).
func writeCodexResponse(conn codexConn, id json.RawMessage, result any) error {
	return writeCodexEnvelope(conn, map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	})
}

// writeCodexErrorResponse sends one JSON-RPC ERROR response line.
func writeCodexErrorResponse(conn codexConn, id json.RawMessage, code int, message string) error {
	return writeCodexEnvelope(conn, map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error":   map[string]any{"code": code, "message": message},
	})
}

// writeCodexEnvelope marshals and sends one NDJSON line.
func writeCodexEnvelope(conn codexConn, envelope map[string]any) error {
	b, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return conn.send(string(b))
}

// bestCodexReviewText prefers the structured exitedReviewMode review over the
// free-form agentMessage text (both carry the same prose in practice).
func bestCodexReviewText(review, agent string) string {
	if review != "" {
		return review
	}
	return agent
}

// codexTurnIDOf reads turn.id from a turn/started notification's params. The
// generated protocol schema (codex-cli 0.146.1) declares TurnStartedParams as
// {threadId, turn} with Turn.id required; that id is what turn/interrupt
// expects as its turnId (M0 probe, progress.md §E.2 (a)).
func codexTurnIDOf(params json.RawMessage) string {
	if len(params) == 0 {
		return ""
	}
	var p struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	_ = json.Unmarshal(params, &p)
	return p.Turn.ID
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

// codexStatedVerdict matches a verdict codex states in its own words, which is
// what the ADVERSARIAL path produces: a leading line reading "Verdict: fail —
// merge blocked." The bullet heuristic above was measured against REVIEW-mode
// output only, and adversarial prose carries none of those bullets — so a
// merge-blocking review was synthesized as a pass, and a required gate inverted
// with nothing in the result saying so (card t178).
//
// The match is deliberately narrow: the label must open a line (optionally
// wrapped in markdown emphasis) and name the verdict on that same line. "I could
// not reach a verdict on the caching layer" is prose about a verdict, not a
// verdict, and must not be read as one. All three schema verdicts are captured
// (card t186): a stated inconclusive is codex saying it could not tell, and
// reading that as the default pass launders uncertainty into a clean verdict.
var codexStatedVerdict = regexp.MustCompile(`(?mi)^[\s>#]*[*_]{0,2}verdict[*_]{0,2}\s*[:\-–—]+[*_]{0,2}\s*[*_]{0,2}(pass|fail|inconclusive)\b`)

// synthesizeReviewOutput maps codex's review prose into the review-output
// schema. codex does NOT return a structured verdict enum — it returns free-form
// prose, and the two review paths express their verdict differently: review mode
// emits severity-tagged finding bullets ("- [P1] ..."), adversarial mode states
// the verdict in a line of its own. Both are read here.
//
// The two signals combine fail-biased: a stated "pass" does NOT clear finding
// bullets, and neither does a stated "inconclusive" — concrete findings outrank
// a stated inability to judge. A stated inconclusive with no bullets is kept as
// inconclusive: uncertainty is never laundered into the clean verdict (t186).
// The summary carries the verbatim review text either way, so the operator sees
// codex's own words rather than only this function's reading of them.
func synthesizeReviewOutput(reviewText string) ReviewOutput {
	verdict := "pass"
	if m := codexStatedVerdict.FindStringSubmatch(reviewText); m != nil {
		verdict = strings.ToLower(m[1])
	}
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
	token := extractProgressToken(req)

	// SPEC-MCP-WORKTREE-ROOT-001 REQ-1/REQ-3: the caller may name the tree codex
	// should review. Absent ⇒ resolveProjectDir(), exactly as before. An unusable
	// path is REJECTED here rather than replaced by the default — a fallback would
	// review the primary checkout while reporting success, which is the defect the
	// parameter exists to fix. The rejection is a tool error, not the fail-open
	// inconclusive verdict: fail-open covers an absent or broken codex, not a
	// caller input the caller can correct.
	root, rootErr := resolveToolProjectRoot(req)
	if rootErr != nil {
		return toolErr("codex_audit", rootErr), nil
	}

	notifyMCPProgress(ctx, token, 0, "codex 감사 시작 — 모드: "+mode+", target: "+target)
	binaryPath, err := codexLookPath(codexBinaryName)
	if err != nil {
		return codexReviewToolResult(inconclusiveReview("codex binary not found in PATH")), nil
	}
	notifyMCPProgress(ctx, token, 0.1, "codex 바이너리 확인 — 리뷰 요청 준비 중...")

	method := codexMethodReviewStart
	params := map[string]any{
		"target": target,
		"model":  model,
		"cwd":    root,
	}
	if mode == codexModeAdversarial {
		method = codexMethodTurnStart
		params["prompt"] = codexAdversarialReviewPrompt(focus)
		if focus != "" {
			params["focus"] = focus
		}
	}

	notifyMCPProgress(ctx, token, 0.2, "codex에 리뷰 요청 전송 중... (수분 소요 가능)")
	out, _ := codexReviewRPC(ctx, binaryPath, method, params) // fail-open inside
	notifyMCPProgress(ctx, token, 0.9, "codex 응답 수신 — 결과 조립 중...")
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

// CodexSetupResult is the plain-Go view of the codex_setup probe state. It is
// shared by the MCP codex_setup tool handler and the web console (via
// internal/cli/web.go injection) so the console CONSUMES the probe rather than
// reimplementing the classification (SPEC-MCP-CONSOLE-001 AC-C-006 — no second
// auth-classification in internal/web).
type CodexSetupResult struct {
	Installed        bool
	Binary           string
	Version          string
	AuthProvider     string
	EnableReviewGate bool
	AllowWrite       bool
}

// ProbeCodexSetup runs the same probe handleCodexSetup runs and returns the
// state as a typed struct. It is the single entry point both the MCP tool
// handler and the web console consume — the auth classification
// (classifyCodexAuth) is defined ONCE here and never forked into internal/web.
// The fail-closed readers (readCodexReviewGateEnabled / readCodexTaskAllowWrite)
// supply the toggle state through their shared workflow.yaml seam.
func ProbeCodexSetup(ctx context.Context) CodexSetupResult {
	binaryPath, err := codexLookPath(codexBinaryName)
	installed := err == nil && binaryPath != ""
	projectDir := projectDirResolver()
	result := CodexSetupResult{
		Installed:        installed,
		AuthProvider:     codexAuthUnknown,
		EnableReviewGate: readCodexReviewGateEnabled(projectDir),
		AllowWrite:       readCodexTaskAllowWrite(projectDir),
	}
	if !installed {
		return result
	}
	result.Binary = binaryPath

	// codex --version (best-effort; failure leaves version empty, not an error).
	if ver, vErr := codexRunner.run(ctx, binaryPath, []string{"--version"}, ""); vErr == nil {
		result.Version = strings.TrimSpace(ver)
	}

	// Auth-provider classification (heuristic — codex's auth surface is
	// undocumented per R1; fail-open to "unknown" on any uncertainty).
	result.AuthProvider = classifyCodexAuth(ctx, binaryPath)
	return result
}

// handleCodexSetup is the thin-wrapper handler for the `codex_setup` MCP tool.
// It delegates to ProbeCodexSetup (the shared probe) and maps the typed result
// to the MCP tool's JSON output. NO Node bridge of any kind (REQ-MCP-007).
// Read-only: it REPORTS the toggle state; mutating it is a heavier,
// wizard-owned concern (M4).
func handleCodexSetup(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Consumed through the codexSetupProbe seam (NOT the direct function) so
	// the sentinel cross-bridge of AC-CL-007 can stub ONE probe and observe
	// the value on every consuming surface — the launcher readout (a) and
	// this tool's response (b) — through the same injection point.
	s := codexSetupProbe(ctx)
	result := map[string]any{
		"installed":     s.Installed,
		"auth_provider": s.AuthProvider,
		// enable_review_gate and allow_write are the two codex opt-ins, reported
		// through the SAME fail-closed reads the gates themselves perform.
		// allow_write's inspectability here is load-bearing, not a convenience:
		// the write-mode opt-in was made a config key rather than an environment
		// variable precisely BECAUSE the key is visible to codex_setup
		// (REQ-CX2-007; plan.md §D M0 write-mode decision, rejected alternative).
		"enable_review_gate": s.EnableReviewGate,
		"allow_write":        s.AllowWrite,
		"node_bridge":        false, // explicit: REQ-MCP-007 Go-only reimplementation
	}
	if s.Installed {
		result["binary"] = s.Binary
		if s.Version != "" {
			result["version"] = s.Version
		}
	}
	return toolJSON("codex_setup", result), nil
}

// ─── two-stage auth classification ladder (SPEC-CODEX-LAUNCHER-001 M1) ───
//
// Stage 1 reads <CODEX_HOME>/auth.json — the structured source `codex doctor`
// itself consults — and classifies ONLY when the mode is a known value AND the
// credential material that mode implies is present. Stage 2 falls back to
// `codex login status` and accepts a provider only from a WHOLE-LINE grammar
// match. Everything else is a gap reported as unknown, never a verdict.
//
// The prior classifier substring-matched a SINGLE stream, so it read
// "API key missing" as a successful apiKey login while discarding the stderr
// the answer actually arrives on (spec §A.2).

// codexAuthFileName is the credential file inside CODEX_HOME.
const codexAuthFileName = "auth.json"

// codexHomeEnvVar / codexHomeDirName drive CODEX_HOME resolution (REQ-CL-005).
const (
	codexHomeEnvVar  = "CODEX_HOME"
	codexHomeDirName = ".codex"

	codexHomeSourceEnv     = "env"
	codexHomeSourceDefault = "default"
)

// codexAuthMode* are the auth_mode enum values stage 1 recognises. An
// unrecognised mode descends to stage 2 rather than guessing.
const (
	codexAuthModeChatGPT = "chatgpt"
	codexAuthModeAPIKey  = "apikey"
)

// errCodexAuthUnparseable is the SANITISED parse-failure reason. The
// underlying encoding/json error is deliberately NOT wrapped: REQ-CL-008
// forbids a credential being retained, logged, or wrapped, and a decoder error
// can quote the offending bytes.
var errCodexAuthUnparseable = errors.New("auth file is not valid JSON for the expected shape")

// codexUserHomeDir is the home-resolution seam. Resolution goes through
// os.UserHomeDir rather than reading $HOME directly, so it behaves on every
// platform (spec §E cross-platform).
var codexUserHomeDir = os.UserHomeDir

// resolveCodexHomeDir returns the resolved CODEX_HOME and which source
// supplied it. An env var that is unset, empty, or whitespace-only falls back
// to the default — reading only LookupEnv's ok flag would accept a blank path.
func resolveCodexHomeDir() (path string, source string) {
	if v := os.Getenv(codexHomeEnvVar); strings.TrimSpace(v) != "" {
		return v, codexHomeSourceEnv
	}
	home, err := codexUserHomeDir()
	if err != nil {
		return "", codexHomeSourceDefault
	}
	return filepath.Join(home, codexHomeDirName), codexHomeSourceDefault
}

// nonEmptyString records ONLY whether a JSON string field carried a non-blank
// value; the value itself dies on the stack. Any other JSON type (null, false,
// a number, an array, an object) counts as absent, and a type mismatch does
// not fail the surrounding document.
type nonEmptyString bool

// UnmarshalJSON implements json.Unmarshaler, keeping the presence bit and
// discarding the value.
func (n *nonEmptyString) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		*n = false // not a string ⇒ not credential material
		return nil
	}
	*n = nonEmptyString(strings.TrimSpace(s) != "")
	return nil // s ends here — it is stored nowhere
}

// codexChatGPTCredentialKeys are the token keys that actually carry a login
// credential. Account metadata such as account_id is deliberately absent: a
// file holding only metadata is not a logged-in state.
var codexChatGPTCredentialKeys = map[string]bool{
	"access_token":  true,
	"id_token":      true,
	"refresh_token": true,
}

// codexTokenSet counts recognised credential keys whose value is a non-blank
// JSON string. It stores a COUNT, never a token.
type codexTokenSet struct{ credentialCount int }

// UnmarshalJSON implements json.Unmarshaler. A non-object value yields a zero
// count rather than an error, so a stale file descends instead of failing.
func (t *codexTokenSet) UnmarshalJSON(b []byte) error {
	var m map[string]nonEmptyString
	if err := json.Unmarshal(b, &m); err != nil {
		t.credentialCount = 0 // not an object ⇒ no credential material
		return nil
	}
	for k, v := range m {
		if codexChatGPTCredentialKeys[k] && bool(v) {
			t.credentialCount++ // ignoring the key would let {"irrelevant":"x"} pass
		}
	}
	return nil
}

// codexAuthFile is the deserialisation target. Every field is a value-free
// kind except auth_mode, which is an enum rather than a secret — AC-CL-008
// asserts that closed set by reflection over this type and its nested types.
type codexAuthFile struct {
	AuthMode string         `json:"auth_mode"`
	APIKey   nonEmptyString `json:"OPENAI_API_KEY"`
	Tokens   codexTokenSet  `json:"tokens"`
}

// classifyCodexAuthFile is stage 1: a PURE judgement over the file's bytes,
// with no disk access. ok=false is a descent to stage 2, never an outcome.
func classifyCodexAuthFile(raw []byte) (provider string, ok bool, err error) {
	var f codexAuthFile
	if uErr := json.Unmarshal(raw, &f); uErr != nil {
		return "", false, errCodexAuthUnparseable
	}
	switch f.AuthMode {
	case codexAuthModeChatGPT:
		if f.Tokens.credentialCount >= 1 {
			return codexAuthChatGPT, true, nil
		}
	case codexAuthModeAPIKey:
		if bool(f.APIKey) {
			return codexAuthAPIKey, true, nil
		}
	}
	return "", false, nil
}

// readCodexAuthFile is the ONLY layer that knows the path. Its errors name the
// path and a reason and nothing else — never the file's contents.
func readCodexAuthFile(codexHome string) (provider string, ok bool, err error) {
	path := filepath.Join(codexHome, codexAuthFileName)
	raw, readErr := os.ReadFile(path) //nolint:gosec // path is the resolved CODEX_HOME, not user input
	if readErr != nil {
		return "", false, fmt.Errorf("codex auth file %s: %w", path, readErr)
	}
	provider, ok, cErr := classifyCodexAuthFile(raw)
	if cErr != nil {
		return "", false, fmt.Errorf("codex auth file %s: %w", path, cErr)
	}
	return provider, ok, nil
}

// codexLoginStatusRunnerFunc is the low-level execution seam for stage 2. It
// returns the two streams SEPARATELY and on purpose: a seam handing back
// pre-combined bytes cannot express "stdout empty, stderr carries the answer",
// which is precisely the state this SPEC repairs — the regression test would
// pass against the defective implementation.
type codexLoginStatusRunnerFunc func(ctx context.Context, binaryPath string) (stdout, stderr []byte, exitCode int, err error)

// codexLoginStatusRunner is the swappable seam (tests replace it with cleanup).
var codexLoginStatusRunner codexLoginStatusRunnerFunc = defaultLoginStatusRunner

// defaultLoginStatusRunner spawns `codex login status` and collects both
// streams. A non-zero exit is DATA, not a failure: codex reports a logged-out
// or degraded state that way and the status line still arrives.
func defaultLoginStatusRunner(ctx context.Context, binaryPath string) ([]byte, []byte, int, error) {
	cmd := exec.CommandContext(ctx, binaryPath, "login", "status")
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if runErr := cmd.Run(); runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			return outBuf.Bytes(), errBuf.Bytes(), exitErr.ExitCode(), nil
		}
		return nil, nil, -1, runErr
	}
	return outBuf.Bytes(), errBuf.Bytes(), 0, nil
}

// combineCodexStreams merges both streams into a newline-joined block of the
// non-blank lines, stdout first. Combining is a named, PURE function so the
// combine rule itself is under test — the defect this SPEC repairs lived in a
// production path that silently dropped one stream.
func combineCodexStreams(stdout, stderr []byte) []byte {
	var lines []string
	for _, chunk := range [][]byte{stdout, stderr} {
		for _, ln := range strings.Split(string(chunk), "\n") {
			if strings.TrimSpace(ln) == "" {
				continue
			}
			lines = append(lines, ln)
		}
	}
	if len(lines) == 0 {
		return nil
	}
	return []byte(strings.Join(lines, "\n"))
}

// codexAuthStatusLine is the WHOLE-LINE grammar of stage 2. Anchoring on
// "logged in" alone is not enough: "Logged in state unavailable: API key
// missing" starts that way and would still be misread. Only the captured term
// classifies, so a line merely CONTAINING a provider word never does.
var codexAuthStatusLine = regexp.MustCompile(`(?i)^[ \t]*logged in using (chatgpt|api key)[ \t]*\r?$`)

// parseCodexAuthLine is stage 2: a PURE parser over the combined block.
//
// exitCode is accepted so callers pass the probe's full result, but it is
// deliberately NOT a discriminator: REQ-CL-009 classifies a non-zero exit
// exactly when a grammar-matching line is present, which is the same rule as
// for a zero exit. Two matching lines naming DIFFERENT providers are a
// conflict, and a conflict is a gap rather than a guess.
func parseCodexAuthLine(combined []byte, exitCode int) string {
	seen := map[string]bool{}
	for _, ln := range strings.Split(string(combined), "\n") {
		m := codexAuthStatusLine.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		seen[strings.ToLower(m[1])] = true
	}
	if len(seen) != 1 {
		return codexAuthUnknown
	}
	if seen[codexAuthModeChatGPT] {
		return codexAuthChatGPT
	}
	return codexAuthAPIKey
}

// classifyCodexAuth is the THIN assembly of the two-stage ladder, and the
// single classification path every consumer shares (the codex_setup MCP tool,
// the web console card, and the launcher readout — REQ-CL-010).
//
// Stage 1 is the structured file; a rejected file — unknown mode, blank
// credential material, unparseable, or absent — is a DESCENT to stage 2, never
// an outcome. Stage 2 reads BOTH streams, because codex writes the status line
// entirely to stderr on this platform and the previous implementation
// discarded it (spec §A.2). Anything unreadable degrades to
// codexAuthUnknown: an unreadable probe is a gap, not a verdict.
func classifyCodexAuth(ctx context.Context, binaryPath string) string {
	if codexHome, _ := resolveCodexHomeDir(); codexHome != "" {
		if provider, ok, _ := readCodexAuthFile(codexHome); ok {
			return provider
		}
	}
	stdout, stderr, exitCode, err := codexLoginStatusRunner(ctx, binaryPath)
	if err != nil {
		return codexAuthUnknown
	}
	return parseCodexAuthLine(combineCodexStreams(stdout, stderr), exitCode)
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

// readCodexTaskAllowWrite reads workflow.codex.task.allow_write from
// `.moai/config/sections/workflow.yaml` (REQ-CX2-007). It is the sibling of
// readCodexReviewGateEnabled above and shares its truth table verbatim
// (fail-CLOSED — the opt-in default is OFF):
//
//   - file missing / unreadable                   → false
//   - YAML parse error                            → false
//   - `workflow.codex.task` block absent          → false (Go zero-value default)
//   - `workflow.codex.task.allow_write: false`    → false
//   - `workflow.codex.task.allow_write: true`     → true
//
// The key path is NESTED under the file's `workflow:` root, matching the shape
// config.Loader expects and the shape the sibling gate reads; the flat form is
// NOT accepted as an alias. Failing closed here is the difference between a
// misconfigured project running codex read-only and a misconfigured project
// handing an external MCP host write access to its working tree (spec.md §H R3).
//
// The distributed default is false and lives in internal/config/defaults.go
// (CodexTaskConfig.AllowWrite); the typed field is internal/config/types.go.
// No template file carries this key (§25 / REQ-CX2-015).
func readCodexTaskAllowWrite(projectDir string) bool {
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
				Task struct {
					AllowWrite bool `yaml:"allow_write"`
				} `yaml:"task"`
			} `yaml:"codex"`
		} `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false
	}
	return doc.Workflow.Codex.Task.AllowWrite
}
