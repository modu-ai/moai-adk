package cli

// managed_codex_factory.go — SPEC-FACTORY-MANAGED-SESSION-001 M2: the Codex
// App Server session owner, rewritten from the PR #1722 reference onto the
// M1 managedSession interface, the current broker API, and the canonical
// lane/leader vocabulary.
//
// Transport (design.md D-2): gorilla/websocket, scoped to the Codex App
// Server loopback client — dial + JSON read/write. The endpoint lives on
// 127.0.0.1 because the launcher owns the codex process on its own host;
// "cross-host" names distinct launcher/backend combinations, never a machine
// boundary. The dial guard (errManagedCodexNonLoopback) enforces that
// structurally: only loopback IP literals dial, so no hostname resolution
// ever sits in the dial path.
//
// Delivery-only boundary (design.md D-1): this owner drives the claim →
// metadata injection → body read (claim token) → receipt flow over the
// internal/factorymsg broker; card disposition and merge windows are
// controller-owned surfaces it never touches, and a receipt is delivery
// evidence only (REQ-MS-015).
//
// Ownership (design.md D-4): the app-server child is exec.Command + Wait —
// the launcher keeps its PID, owns the process, and tears it down in Close.
// No process-replacement exec primitive, so the Windows crossbuild rides the
// same code path.

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

var errManagedCodexNonLoopback = errors.New("managed codex app server endpoint must be a loopback IP literal (127.0.0.0/8 or ::1) over ws://")

// managedLoopbackURL admits only ws:// URLs whose host is a loopback IP
// literal. "localhost" is refused on purpose: accepting it would put name
// resolution back into the dial path, and the AC is a structural property,
// not a resolvable one.
func managedLoopbackURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "ws" {
		return fmt.Errorf("%w: %q", errManagedCodexNonLoopback, raw)
	}
	ip := net.ParseIP(u.Hostname())
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("%w: %q", errManagedCodexNonLoopback, raw)
	}
	return nil
}

// Test seams: tests force the early Start failures that a real filesystem and
// loopback cannot produce on demand.
var (
	managedCodexWriteToken  = os.WriteFile
	managedCodexAllocateURL = managedCodexAppEndpoint
)

// managedCodexAppEndpoint allocates the ephemeral loopback port the app
// server listens on. The owner closes its probe listener before the child
// binds — the standard ephemeral-port handoff, same shape as the reference.
func managedCodexAppEndpoint() (string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	if err := ln.Close(); err != nil {
		return "", err
	}
	return "ws://127.0.0.1:" + strconv.Itoa(port), nil
}

// managedCodexAppToken mints the per-launch capability token (design.md
// D-2): 32 random bytes, hex-encoded. The token lives only for the session:
// the app server reads it from a 0600 token file, the owner dials with it as
// a Bearer header, and the token directory is removed at teardown.
func managedCodexAppToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// managedCodexAppReady polls the app server's /readyz until it answers 200
// or the context ends. The probe target passes the same loopback guard the
// dial does.
func managedCodexAppReady(ctx context.Context, raw string) error {
	if err := managedLoopbackURL(raw); err != nil {
		return err
	}
	client := &http.Client{
		Timeout: 300 * time.Millisecond,
		// A redirect target must pass the same loopback rule as the probe
		// target, or a loopback endpoint could steer the probe off the machine.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("managed codex readiness probe: too many redirects")
			}
			return managedLoopbackURL(strings.Replace(req.URL.String(), "http", "ws", 1))
		},
	}
	endpoint := strings.Replace(raw, "ws://", "http://", 1) + "/readyz"
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if errors.Is(err, errManagedCodexNonLoopback) {
			return err
		}
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// managedCodexDial dials the app server WS after the loopback guard. The
// guard fires before any network activity, so a non-loopback target can
// never leave the process.
func managedCodexDial(raw string, header http.Header) (*websocket.Conn, *http.Response, error) {
	if err := managedLoopbackURL(raw); err != nil {
		return nil, nil, err
	}
	dialer := &websocket.Dialer{HandshakeTimeout: config.DefaultManagedCodexReadyTimeout}
	return dialer.Dial(raw, header)
}

// managedCodexRPCError is one App Server JSON-RPC error object.
type managedCodexRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// managedCodexAppReply is one decoded frame, classified by field presence: an
// id and a method make a server-originated request, a method alone a
// notification, an id alone a response. The id stays raw JSON because a server
// request may carry an integer or a string id, and the answer must echo it
// verbatim.
type managedCodexAppReply struct {
	ID     json.RawMessage       `json:"id"`
	Method string                `json:"method"`
	Params json.RawMessage       `json:"params"`
	Result json.RawMessage       `json:"result"`
	Error  *managedCodexRPCError `json:"error"`

	// brokerDeclined is the read goroutine's verdict on a turn/completed event:
	// true when a declined MoAI broker elicitation was attributed to the turn
	// that just completed. It is never decoded from the wire.
	brokerDeclined bool
}

// hasID reports whether the frame carries a usable id; a missing id and a
// literal JSON null both count as no id.
func (f managedCodexAppReply) hasID() bool {
	return len(f.ID) > 0 && string(f.ID) != "null"
}

// idIs reports whether the frame id is the integer id of a pending client
// call. A string id never matches, so a server request cannot be taken for a
// response.
func (f managedCodexAppReply) idIs(id int) bool {
	n, err := strconv.Atoi(string(f.ID))
	return err == nil && n == id
}

// managedLogOutput is the single destination for the owner's operator-facing
// log lines, and the only test seam of this file: a test stores a synchronized
// writer here, production leaves it nil and the lines go to os.Stderr (stdout
// belongs to model output). The atomic pointer protects the pointer load only,
// so a writer stored here must synchronize its own Write.
var managedLogOutput atomic.Pointer[io.Writer]

func managedLogf(format string, args ...any) {
	var out io.Writer = os.Stderr
	if w := managedLogOutput.Load(); w != nil {
		out = *w
	}
	_, _ = fmt.Fprintf(out, format+"\n", args...)
}

// managedCodexAppClient is one App Server WS connection. The read goroutine
// answers server-originated requests while DeliverTurn writes its own calls,
// so every connection write goes through write, which holds writeMu around the
// WriteJSON call and nothing else.
type managedCodexAppClient struct {
	conn      *websocket.Conn
	events    chan managedCodexAppReply
	done      chan struct{}
	nextID    int
	completed map[string]string
	// brokerFailed holds the read goroutine's verdict per completed turn id; only
	// the consumer (call and waitTurn) touches it, like completed.
	brokerFailed map[string]bool

	writeMu sync.Mutex

	// The turn window is owned by the read goroutine's frame order: it opens
	// in armTurn before turn/start is written and closes when that turn's
	// turn/completed frame is read. mu guards it and is never held across a
	// connection write.
	mu             sync.Mutex
	open           bool
	turnID         string
	prevTurnID     string
	brokerDeclined int
}

// write sends one JSON frame; gorilla/websocket allows a single concurrent
// writer, and the call loop and the read goroutine are two.
//
// @MX:ANCHOR: [AUTO] the single path for every connection write (call, the initialized notification, server-request replies) — fan_in 3
// @MX:REASON: gorilla/websocket forbids concurrent writers, so a write that bypasses writeMu races the read goroutine's replies with the call loop
// @MX:SPEC: SPEC-FACTORY-MANAGED-HARDEN-001
func (c *managedCodexAppClient) write(v any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteJSON(v)
}

// armTurn opens a fresh turn window. It runs before turn/start is written so a
// frame belonging to the new turn can never arrive while the window is closed.
// prevTurnID survives on purpose: it identifies late frames of the turn that
// just ended.
func (c *managedCodexAppClient) armTurn() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.open = true
	c.turnID = ""
	c.brokerDeclined = 0
}

// managedTurnFrame decodes the turn object of a turn/started or turn/completed
// notification.
func managedTurnFrame(params json.RawMessage) (id, status string) {
	var payload struct {
		Turn struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"turn"`
	}
	if json.Unmarshal(params, &payload) != nil {
		return "", ""
	}
	return payload.Turn.ID, payload.Turn.Status
}

// noteTurnStarted records the id of the turn the open window belongs to.
func (c *managedCodexAppClient) noteTurnStarted(params json.RawMessage) {
	id, _ := managedTurnFrame(params)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.open && c.turnID == "" && id != "" {
		c.turnID = id
	}
}

// noteTurnCompleted closes the window for the completed turn and returns the
// verdict the completion event carries. A completion of the previous turn (a
// duplicate or late frame) and a completion of some other turn leave the window
// alone and carry no verdict.
func (c *managedCodexAppClient) noteTurnCompleted(params json.RawMessage) bool {
	id, _ := managedTurnFrame(params)
	c.mu.Lock()
	defer c.mu.Unlock()
	if id == "" || id == c.prevTurnID || !c.open || (c.turnID != "" && c.turnID != id) {
		return false
	}
	declined := c.brokerDeclined > 0
	c.prevTurnID = id
	c.open = false
	return declined
}

// noteElicitation attributes one answered mcpServer/elicitation request to the
// turn window. Only a request from the MoAI broker server that arrives while
// the window is open counts, and only when its turnId is neither the previous
// turn's nor a turn other than the one the window already knows. It returns
// the attributed turn id ("none" when the request was not counted) and the
// window's declined count after this request.
func (c *managedCodexAppClient) noteElicitation(serverName string, requestTurn string) (turn string, declined int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	turn = "none"
	if !c.open || serverName != moaiMCPServerKey {
		return turn, c.brokerDeclined
	}
	switch {
	case requestTurn != "" && requestTurn == c.prevTurnID:
		// A late request of the turn that just ended.
		return turn, c.brokerDeclined
	case c.turnID != "" && requestTurn != "" && requestTurn != c.turnID:
		// A request of some turn other than the one this window belongs to.
		return turn, c.brokerDeclined
	}
	c.brokerDeclined++
	switch {
	case c.turnID != "":
		turn = c.turnID
	case requestTurn != "":
		turn = requestTurn
	}
	return turn, c.brokerDeclined
}

// JSON-RPC error codes the owner answers server requests with.
const (
	managedRPCServerError    = -32000
	managedRPCMethodNotFound = -32601
)

const managedElicitationMethod = "mcpServer/elicitation/request"

// managedLegacyApprovalRejection is the refusal text of the two legacy approval
// requests (applyPatchApproval, execCommandApproval). Codex's ReviewDecision
// carries a denial as an object, {"denied":{"rejection":<text>}}, never as the
// bare string "denied". The text names no id, path or secret.
const managedLegacyApprovalRejection = "managed Factory layer is delivery-only and never grants approvals"

// managedServerRequestPolicy is how the owner answers one server request kind.
// result nil means a JSON-RPC error with errCode; outcome is the token the log
// line carries.
type managedServerRequestPolicy struct {
	outcome string
	result  any
	errCode int
}

// managedServerRequestPolicies is the response policy (design.md D-1): decline
// or grant nothing. No entry accepts anything and none uses cancel or abort,
// which would interrupt the turn the model still needs to write its receipt in.
// Values the owner cannot invent (tokens, attestations, user answers) answer
// with an error.
//
// @MX:NOTE: [AUTO] the only place a server request is mapped to an answer, and the only place its log token is chosen — keep table and log in step
var managedServerRequestPolicies = map[string]managedServerRequestPolicy{
	"item/commandExecution/requestApproval": {outcome: "decline", result: map[string]any{"decision": "decline"}},
	"item/fileChange/requestApproval":       {outcome: "decline", result: map[string]any{"decision": "decline"}},
	"item/permissions/requestApproval":      {outcome: "empty", result: map[string]any{"permissions": map[string]any{}}},
	managedElicitationMethod:                {outcome: "decline", result: map[string]any{"action": "decline"}},
	"item/tool/requestUserInput":            {outcome: "error", errCode: managedRPCServerError},
	"item/tool/call": {outcome: "failed", result: map[string]any{
		"contentItems": []map[string]string{{"type": "inputText", "text": "managed Factory session registers no dynamic tools"}},
		"success":      false,
	}},
	"account/chatgptAuthTokens/refresh": {outcome: "error", errCode: managedRPCServerError},
	"attestation/generate":              {outcome: "error", errCode: managedRPCServerError},
	"applyPatchApproval":                {outcome: "denied", result: managedLegacyDeniedResult()},
	"execCommandApproval":               {outcome: "denied", result: managedLegacyDeniedResult()},
}

// managedLegacyDeniedResult builds the schema-valid refusal of a legacy approval
// request: ReviewDecision's denied variant is an object with a required
// rejection string. It declines the action and keeps the session alive, unlike
// the bare string "abort", which would interrupt the turn.
func managedLegacyDeniedResult() map[string]any {
	return map[string]any{"decision": map[string]any{"denied": map[string]any{"rejection": managedLegacyApprovalRejection}}}
}

// managedUnknownRequestPolicy answers a method the table does not name: silence
// would leave the server waiting forever.
var managedUnknownRequestPolicy = managedServerRequestPolicy{outcome: "error", errCode: managedRPCMethodNotFound}

// answerServerRequest answers one server-originated request from the read
// goroutine. The log line is written before the reply so an operator-visible
// trace exists by the time the server sees the answer; the reply echoes the
// request id verbatim. The turn-window lock is released before the write.
func (c *managedCodexAppClient) answerServerRequest(req managedCodexAppReply) error {
	policy, ok := managedServerRequestPolicies[req.Method]
	if !ok {
		policy = managedUnknownRequestPolicy
	}
	line := fmt.Sprintf("Factory server request answered: %q -> %s", req.Method, policy.outcome)
	if req.Method == managedElicitationMethod {
		var params struct {
			ServerName string  `json:"serverName"`
			TurnID     *string `json:"turnId"`
		}
		_ = json.Unmarshal(req.Params, &params)
		requestTurn := ""
		if params.TurnID != nil {
			requestTurn = *params.TurnID
		}
		turn, declined := c.noteElicitation(params.ServerName, requestTurn)
		line += fmt.Sprintf(" serverName=%q turn=%s broker_declined=%d", params.ServerName, turn, declined)
	}
	managedLogf("%s", line)

	reply := map[string]any{"id": req.ID}
	if policy.result != nil {
		reply["result"] = policy.result
	} else {
		reply["error"] = managedCodexRPCError{
			Code:    policy.errCode,
			Message: fmt.Sprintf("managed Factory session cannot answer %q: no operator is attached", req.Method),
		}
	}
	return c.write(reply)
}

// @MX:WARN: [AUTO] the read goroutine outlives the calls that start it and its only stop signal is the done channel
// @MX:REASON: a blocked events send would leak the goroutine after Close if it did not select on done — keep the select when touching read()
func (c *managedCodexAppClient) read() {
	defer close(c.events)
	for {
		var event managedCodexAppReply
		if err := c.conn.ReadJSON(&event); err != nil {
			return
		}
		switch {
		case event.hasID() && event.Method != "":
			// A server-originated request never reaches the call loop: it is
			// answered here, even between turns. A failed answer means the
			// connection is dead.
			if err := c.answerServerRequest(event); err != nil {
				return
			}
			continue
		case event.Method == "turn/started":
			c.noteTurnStarted(event.Params)
		case event.Method == "turn/completed":
			event.brokerDeclined = c.noteTurnCompleted(event.Params)
		case event.Method != "" || !event.hasID():
			// Streaming item deltas and id-less frames belong to the
			// interactive surface, not the delivery path.
			continue
		}
		// What remains for the turn loop: lifecycle notifications and RPC
		// responses.
		select {
		case c.events <- event:
		case <-c.done:
			return
		}
	}
}

// observe folds one lifecycle notification into the client's turn-completion
// table; the delivery driver's turn waits resolve against it.
func (c *managedCodexAppClient) observe(event managedCodexAppReply) {
	if event.Method != "turn/completed" {
		return
	}
	var payload struct {
		Turn struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"turn"`
	}
	if json.Unmarshal(event.Params, &payload) == nil && payload.Turn.ID != "" {
		if c.completed == nil {
			c.completed = map[string]string{}
		}
		c.completed[payload.Turn.ID] = payload.Turn.Status
		if event.brokerDeclined {
			if c.brokerFailed == nil {
				c.brokerFailed = map[string]bool{}
			}
			c.brokerFailed[payload.Turn.ID] = true
		}
	}
}

func (c *managedCodexAppClient) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	c.nextID++
	id := c.nextID
	if err := c.write(map[string]any{"id": id, "method": method, "params": params}); err != nil {
		return nil, err
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-c.done:
			return nil, errors.New("managed codex app server connection closed")
		case event, ok := <-c.events:
			if !ok {
				return nil, errors.New("managed codex app server connection closed")
			}
			c.observe(event)
			if !event.idIs(id) {
				continue
			}
			if event.Error != nil {
				return nil, fmt.Errorf("%s: %s (%d)", method, event.Error.Message, event.Error.Code)
			}
			return event.Result, nil
		}
	}
}

// waitTurn blocks until the named turn completes. A completion already
// folded into the table resolves immediately. Two outcomes are turn-scoped
// failures (errManagedTurnFailed): a turn that ended in any state other than
// completed, and a turn the App Server marked completed but that carries a
// declined MoAI broker elicitation — the receipt tool may never have run. The
// timeout and connection errors below stay unmarked, so they remain
// session-fatal.
func (c *managedCodexAppClient) waitTurn(ctx context.Context, turnID string) error {
	for {
		if status, ok := c.completed[turnID]; ok {
			brokerFailed := c.brokerFailed[turnID]
			delete(c.completed, turnID)
			delete(c.brokerFailed, turnID)
			if status != "completed" {
				return fmt.Errorf("%w: managed codex turn %s ended as %s", errManagedTurnFailed, turnID, status)
			}
			if brokerFailed {
				return fmt.Errorf("%w: managed codex turn %s completed but a MoAI broker elicitation was declined during it", errManagedTurnFailed, turnID)
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.done:
			return errors.New("managed codex app server connection closed")
		case event, ok := <-c.events:
			if !ok {
				return errors.New("managed codex app server connection closed")
			}
			c.observe(event)
		}
	}
}

func (c *managedCodexAppClient) startTurn(ctx context.Context, threadID, prompt string) error {
	c.armTurn()
	result, err := c.call(ctx, "turn/start", map[string]any{
		"threadId": threadID,
		"input":    []map[string]string{{"type": "text", "text": prompt}},
	})
	if err != nil {
		return err
	}
	var reply struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	if err := json.Unmarshal(result, &reply); err != nil || reply.Turn.ID == "" {
		return errors.New("managed codex turn/start returned no turn id")
	}
	return c.waitTurn(ctx, reply.Turn.ID)
}

// shutdown stops the read goroutine and the connection exactly once.
func (c *managedCodexAppClient) shutdown() {
	select {
	case <-c.done:
	default:
		close(c.done)
	}
	_ = c.conn.Close()
}

// managedCodexOptions parses the operator arguments the managed codex launch
// admits: config overrides pass through to the app server, the model moves
// to thread/start, everything else is refused with the offending flag named.
// args[0] is the program name and is never scanned.
func managedCodexOptions(args []string) (appArgs []string, model string, err error) {
	for i := 1; i < len(args); i++ {
		switch arg := args[i]; {
		case arg == "-c" || arg == "--config":
			if i+1 >= len(args) {
				return nil, "", fmt.Errorf("managed codex launch %s requires a value", arg)
			}
			i++
			appArgs = append(appArgs, "-c", args[i])
		case strings.HasPrefix(arg, "--config="):
			appArgs = append(appArgs, "-c", strings.TrimPrefix(arg, "--config="))
		case arg == "-m" || arg == "--model":
			if i+1 >= len(args) || strings.TrimSpace(args[i+1]) == "" {
				return nil, "", fmt.Errorf("managed codex launch %s requires a model", arg)
			}
			i++
			model = args[i]
		case strings.HasPrefix(arg, "--model="):
			model = strings.TrimPrefix(arg, "--model=")
			if model == "" {
				return nil, "", errors.New("managed codex launch --model requires a model")
			}
		default:
			return nil, "", fmt.Errorf("managed codex launch does not support %s", arg)
		}
	}
	return appArgs, model, nil
}

// factoryMoAIMCPApprovalArgs are the config overrides that let the owned
// App Server take MoAI broker tool calls without an approval dialog
// (REQ-MS-009, D-5). Every override addresses mcp_servers.moai only, and the
// two broker tools are approved individually. They ride the owned process
// command line and never reach the project config, whose generated
// default_tools_approval_mode stays "writes".
func factoryMoAIMCPApprovalArgs() []string {
	return []string{
		"-c", `mcp_servers.moai.default_tools_approval_mode="approve"`,
		"-c", `mcp_servers.moai.tools.factory_msg_send.approval_mode="approve"`,
		"-c", `mcp_servers.moai.tools.factory_msg_receipt.approval_mode="approve"`,
	}
}

// managedCodexAppServerArgs builds the owned App Server command line: the
// listen/auth flags, then the approval scoping, then the operator's own
// overrides last so an explicit operator value still wins.
func managedCodexAppServerArgs(url, tokenFile string, operatorArgs []string) []string {
	args := []string{"app-server", "--listen", url, "--ws-auth", "capability-token", "--ws-token-file", tokenFile}
	args = append(args, factoryMoAIMCPApprovalArgs()...)
	return append(args, operatorArgs...)
}

// managedCodexSession owns one Codex App Server child process and its WS
// control connection. It implements the M1 managedSession interface: Start
// spawns and handshakes, DeliverTurn injects one turn, Close tears the child
// down. This surface is headless on purpose — the managed session is the
// delivery loop's backend. TUI attach is not delivered by
// SPEC-FACTORY-MANAGED-SESSION-001 and is owed to a follow-up card, so model
// output renders wherever the app server's own thread view renders it, not
// on our stdout.
type managedCodexSession struct {
	program  string
	appArgs  []string
	model    string
	env      []string
	label    string
	cmd      *exec.Cmd
	dir      string // launch directory (project root or -w worktree); empty inherits the process cwd
	tokenDir string
	client   *managedCodexAppClient
	threadID string
	started  bool
	closed   bool
}

// newManagedCodexSession builds the session without starting it: operator
// arguments are parsed once, and the thread label comes from the factory
// environment (lane-<n>) with the leader fallback.
func newManagedCodexSession(bin string, args, env []string) (*managedCodexSession, error) {
	appArgs, model, err := managedCodexOptions(args)
	if err != nil {
		return nil, err
	}
	label := launchEnvValue(env, config.EnvMoaiFactoryWorker)
	if label == "" {
		label = factory.RoleLeader
	}
	return &managedCodexSession{program: bin, appArgs: appArgs, model: model, env: env, label: label}, nil
}

// Start spawns the app server, waits out its readiness, dials the loopback
// WS with the capability token, completes the JSON-RPC handshake, and starts
// the session thread. Everything here is launch-lifecycle; the entry point
// binds the broker endpoint afterwards.
func (s *managedCodexSession) Start() error {
	token, err := managedCodexAppToken()
	if err != nil {
		return err
	}
	tokenDir, err := os.MkdirTemp("", "moai-factory-app-")
	if err != nil {
		return err
	}
	s.tokenDir = tokenDir
	// Until the child exists, Close is a no-op, so any failure before
	// cmd.Start must drop the token directory here.
	defer func() {
		if !s.started {
			_ = os.RemoveAll(s.tokenDir)
			s.tokenDir = ""
		}
	}()
	tokenFile := filepath.Join(tokenDir, "token")
	if err := managedCodexWriteToken(tokenFile, []byte(token), 0o600); err != nil {
		return err
	}
	url, err := managedCodexAllocateURL()
	if err != nil {
		return err
	}
	args := managedCodexAppServerArgs(url, tokenFile, s.appArgs)
	s.cmd = exec.Command(s.program, args...)
	s.cmd.Env = s.env
	s.cmd.Dir = s.dir
	s.cmd.Stderr = os.Stderr
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("start managed Factory codex app server: %w", err)
	}
	s.started = true

	readyCtx, cancel := context.WithTimeout(context.Background(), config.DefaultManagedCodexReadyTimeout)
	defer cancel()
	if err := managedCodexAppReady(readyCtx, url); err != nil {
		return fmt.Errorf("managed codex app server readiness: %w", err)
	}
	conn, _, err := managedCodexDial(url, http.Header{"Authorization": []string{"Bearer " + token}})
	if err != nil {
		return err
	}
	s.client = &managedCodexAppClient{conn: conn, events: make(chan managedCodexAppReply, 32), done: make(chan struct{})}
	go s.client.read()
	if _, err := s.client.call(readyCtx, "initialize", map[string]any{
		"clientInfo": map[string]string{"name": "moai_factory", "title": "MoAI Factory", "version": "1"},
	}); err != nil {
		return err
	}
	if err := s.client.write(map[string]any{"method": "initialized", "params": map[string]any{}}); err != nil {
		return err
	}

	threadParams := map[string]any{}
	if s.dir != "" {
		threadParams["cwd"] = s.dir
	} else if cwd, cwdErr := os.Getwd(); cwdErr == nil {
		threadParams["cwd"] = cwd
	}
	if s.model != "" {
		threadParams["model"] = s.model
	}
	threadResult, err := s.client.call(readyCtx, "thread/start", threadParams)
	if err != nil {
		return err
	}
	var thread struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err := json.Unmarshal(threadResult, &thread); err != nil || thread.Thread.ID == "" {
		return errors.New("managed codex thread/start returned no thread id")
	}
	s.threadID = thread.Thread.ID
	_, err = s.client.call(readyCtx, "thread/name/set", map[string]any{"threadId": s.threadID, "name": s.label})
	return err
}

// DeliverTurn injects one prompt as the thread's next user turn and blocks
// until turn/completed. Turn completion is transport fact; the broker
// receipt the model writes afterwards remains the delivery evidence
// (REQ-MS-015).
func (s *managedCodexSession) DeliverTurn(prompt string) error {
	if s.client == nil {
		return errors.New("managed codex session not started")
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultManagedCodexTurnTimeout)
	defer cancel()
	return s.client.startTurn(ctx, s.threadID, prompt)
}

// Close tears the child down: WS close, kill if still running, wait, and the
// token directory removed — the ownership model's teardown half, idempotent
// like its stream sibling.
func (s *managedCodexSession) Close() error {
	if !s.started || s.closed {
		return nil
	}
	s.closed = true
	if s.client != nil {
		s.client.shutdown()
	}
	killed := false
	if s.cmd.ProcessState == nil {
		killed = s.cmd.Process.Kill() == nil
	}
	waitErr := s.cmd.Wait()
	removeErr := os.RemoveAll(s.tokenDir)
	if !killed && waitErr != nil {
		return errors.Join(waitErr, removeErr)
	}
	return removeErr
}

// runManagedFactoryCodex owns one managed Factory Codex session: flag
// parsing, launch-pending registration with rollback on a failed start, the
// App Server handshake, the launcher-side peer bind (thread id as session
// UUID), then the shared M1 delivery driver over the operator's stdin and
// the broker inbox. stdin is a parameter so tests can drive a real session
// without process-global mutation (the entry points pass os.Stdin).
func runManagedFactoryCodex(bin string, args, env []string, dir string, stdin io.Reader) (err error) {
	root := launchProjectRoot()
	runID := launchEnvValue(env, config.EnvFactoryRunID)
	if runID == "" {
		return errors.New("factory managed session requires a factory run id")
	}
	ownerPID := os.Getpid()
	ownerStart := homestate.CurrentProcessFingerprint()
	if ownerStart == "" {
		return errors.New("factory managed session owner identity unavailable")
	}
	launchEnv := withSessionPID(env, ownerPID)
	// Operator arguments are parsed before any broker or child work: a
	// refusal never leaves state behind.
	session, err := newManagedCodexSession(bin, args, launchEnv)
	if err != nil {
		return err
	}
	session.dir = dir
	pending, err := registerFactoryLaunchPending(context.Background(), root, launchEnv, ownerPID, ownerStart)
	if err != nil {
		return err
	}
	// REQ-MS-002: until the launcher binds the thread id itself, a failed
	// handshake leaves no launch-pending row behind.
	bound := false
	defer func() {
		if !bound {
			err = errors.Join(err, rollbackFactoryLaunchPending(context.Background(), root, pending))
		}
	}()

	defer func() { err = errors.Join(err, session.Close()) }()
	if err = session.Start(); err != nil {
		return err
	}

	store, err := factorymsg.Open(root, runID)
	if err != nil {
		return err
	}
	defer closeFactoryToolStore("managed_codex_factory", store)
	// App Server turns run no TUI SessionStart hook, so the launcher — which
	// owns both the thread id and the process identity — completes the bind
	// half of REQ-MS-002 itself over the store's atomic launch-pending bind.
	if pending.Slot != "" {
		bindPeer := pending
		bindPeer.SessionUUID = session.threadID
		if _, _, err := store.BindLaunchPending(context.Background(), bindPeer); err != nil {
			return fmt.Errorf("managed codex session bind: %w", err)
		}
	}
	bound = true

	claim, toPrompt := managedFactoryInboxWiring(store, ownerPID, ownerStart, runID)
	ticker := time.NewTicker(config.DefaultManagedSessionPollInterval)
	defer ticker.Stop()
	return driveManagedFactorySession(session, stdin, ticker.C, claim, toPrompt)
}
