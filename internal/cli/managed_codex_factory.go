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
	"time"

	"github.com/gorilla/websocket"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
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
	client := &http.Client{Timeout: 300 * time.Millisecond}
	endpoint := strings.Replace(raw, "ws://", "http://", 1) + "/readyz"
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
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

// managedCodexAppReply is one frame: an RPC response (ID set), a lifecycle
// notification (Method set, ID 0), or a response carrying an error object.
type managedCodexAppReply struct {
	ID     int                   `json:"id"`
	Method string                `json:"method"`
	Params json.RawMessage       `json:"params"`
	Result json.RawMessage       `json:"result"`
	Error  *managedCodexRPCError `json:"error"`
}

// managedCodexAppClient is one App Server WS connection. DeliverTurn is the
// only writer and the M1 driver serializes turns, so the connection needs no
// write mutex; reads run on their own goroutine.
type managedCodexAppClient struct {
	conn      *websocket.Conn
	events    chan managedCodexAppReply
	done      chan struct{}
	nextID    int
	completed map[string]string
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
		// The turn loop consumes lifecycle notifications and RPC replies;
		// streaming item deltas belong to the interactive surface, not the
		// delivery path.
		if event.ID == 0 && event.Method != "turn/started" && event.Method != "turn/completed" {
			continue
		}
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
	}
}

func (c *managedCodexAppClient) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	c.nextID++
	id := c.nextID
	if err := c.conn.WriteJSON(map[string]any{"id": id, "method": method, "params": params}); err != nil {
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
			if event.ID != id {
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
// folded into the table resolves immediately; a turn that ended in any other
// state fails the delivery.
func (c *managedCodexAppClient) waitTurn(ctx context.Context, turnID string) error {
	for {
		if status, ok := c.completed[turnID]; ok {
			delete(c.completed, turnID)
			if status != "completed" {
				return fmt.Errorf("managed codex turn %s ended as %s", turnID, status)
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
// delivery loop's backend; the interactive TUI attach is later-milestone
// launcher wiring (plan.md M3/M4), so model output renders wherever the
// app server's own thread view renders it, not on our stdout.
type managedCodexSession struct {
	program  string
	appArgs  []string
	model    string
	env      []string
	label    string
	cmd      *exec.Cmd
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
		label = kanban.RoleLeader
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
	tokenFile := filepath.Join(tokenDir, "token")
	if err := os.WriteFile(tokenFile, []byte(token), 0o600); err != nil {
		return err
	}
	url, err := managedCodexAppEndpoint()
	if err != nil {
		return err
	}
	args := managedCodexAppServerArgs(url, tokenFile, s.appArgs)
	s.cmd = exec.Command(s.program, args...)
	s.cmd.Env = s.env
	s.cmd.Stderr = os.Stderr
	if err := s.cmd.Start(); err != nil {
		// No child exists, so the deferred Close will not run its teardown:
		// drop the token directory here instead.
		_ = os.RemoveAll(s.tokenDir)
		s.tokenDir = ""
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
	if err := conn.WriteJSON(map[string]any{"method": "initialized", "params": map[string]any{}}); err != nil {
		return err
	}

	threadParams := map[string]any{}
	if cwd, cwdErr := os.Getwd(); cwdErr == nil {
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
func runManagedFactoryCodex(bin string, args, env []string, stdin io.Reader) (err error) {
	root := launchProjectRoot()
	runID := launchEnvValue(env, config.EnvMoaiKanbanID)
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

	claim := func() ([]factorymsg.Claim, error) { return claimManagedFactoryInbox(store, ownerPID, ownerStart) }
	toPrompt := func(claims []factorymsg.Claim) string { return managedFactoryInboxPrompt(runID, claims) }
	ticker := time.NewTicker(config.DefaultManagedSessionPollInterval)
	defer ticker.Stop()
	return driveManagedFactorySession(session, stdin, ticker.C, claim, toPrompt)
}
