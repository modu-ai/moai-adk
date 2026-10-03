package cli

// managed_codex_tui.go — SPEC-FACTORY-MANAGED-TUI-001: the operator TUI attach
// of the managed Codex session owner.
//
// The owner is a headless App Server client (SPEC-FACTORY-MANAGED-SESSION-001
// Amendment 1). This file attaches the Codex TUI as a second client of the
// SAME owned App Server, resuming the thread the broker endpoint is bound to,
// and owns everything the attach changes: the capability probe and the
// fallback notices, who owns the terminal, where the launcher's own output
// goes, and the TUI child's start and stop.
//
// Zero-syscall (REQ-MT-012): os/exec, os.Interrupt, os.Process and isatty only,
// so the Windows cross build stays on one code path (os.Interrupt is
// unsupported there and the stop falls straight through to Kill).

import (
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
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mattn/go-isatty"
	"github.com/modu-ai/moai-adk/internal/config"
)

// Package-private seams. Production defaults are the real behavior; tests
// replace them so the timing and terminal rules run without a real terminal.
var (
	// managedTerminalCheck reports whether a stream is a terminal.
	managedTerminalCheck = func(f *os.File) bool { return f != nil && isatty.IsTerminal(f.Fd()) }
	// managedBusyWarnInterval is how often a long busy thread is logged.
	managedBusyWarnInterval = config.DefaultManagedCodexTurnTimeout
	// managedWriteDeadline bounds one App Server connection write.
	managedWriteDeadline = config.DefaultManagedCodexTurnTimeout
	// managedTUIStopGrace is the wait between the interrupt and the kill.
	managedTUIStopGrace = config.DefaultManagedCodexTUIStopGrace
	// managedProbeTimeout bounds the capability probe.
	managedProbeTimeout = config.DefaultManagedCodexProbeTimeout
	// managedCodexTUICommand creates the operator TUI child.
	managedCodexTUICommand = exec.Command
	// managedCodexRemoteProbe asks the codex binary whether `resume` offers the
	// remote options; the second result names the reason when it does not.
	managedCodexRemoteProbe = probeManagedCodexRemote
	// managedWriteBarrier is a test hook called at the entry of every App Server
	// connection write.
	managedWriteBarrier func(v any)
)

// errManagedCodexConnectionClosed is what the driver returns when the App
// Server connection ended under a running TUI.
var errManagedCodexConnectionClosed = errors.New("managed codex app server connection closed")

const (
	// managedTUINoticeFormat is the single stderr line a skipped or failed attach
	// writes (REQ-MT-004).
	managedTUINoticeFormat = "Factory managed session: operator TUI not attached (%s); continuing headless"
	// managedTUILogFormat names the session log file, printed once before the
	// TUI could start (REQ-MT-006).
	managedTUILogFormat = "Factory managed session log file: %s"
)

// managedOperatorSurface is the optional capability the Codex owner offers the
// delivery driver while an operator TUI is attached. The managedSession
// interface itself is unchanged.
type managedOperatorSurface interface {
	// AttachOperator starts the operator TUI when the session planned one. The
	// channel delivers the single end-of-TUI result; attached is false for a
	// headless session.
	AttachOperator() (done <-chan error, attached bool)
	// Busy reports whether the thread has an active turn.
	Busy() bool
}

var _ managedOperatorSurface = (*managedCodexSession)(nil)

var (
	managedRemoteOption      = regexp.MustCompile(`(?m)^\s*(?:-\w,\s*)?--remote(?:[\s=<\[]|$)`)
	managedRemoteTokenOption = regexp.MustCompile(`(?m)^\s*(?:-\w,\s*)?--remote-auth-token-env(?:[\s=<\[]|$)`)
)

// managedCodexRemoteSupport reports whether a `codex resume --help` text offers
// both the --remote and the --remote-auth-token-env option. Feature detection,
// not a version list: the help text is what the binary actually offers.
func managedCodexRemoteSupport(help string) bool {
	return managedRemoteOption.MatchString(help) && managedRemoteTokenOption.MatchString(help)
}

// probeManagedCodexRemote runs `codex resume --help` under the probe timeout
// and reads the options out of it.
func probeManagedCodexRemote(program string, env []string, dir string) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), managedProbeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, program, "resume", "--help")
	cmd.Env = env
	cmd.Dir = dir
	// A killed child must not leave Wait blocked on a pipe a grandchild holds.
	cmd.WaitDelay = managedProbeTimeout
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if ctx.Err() != nil {
		return false, "capability probe timed out after " + managedProbeTimeout.String()
	}
	if err != nil {
		return false, "capability probe failed: " + err.Error()
	}
	if !managedCodexRemoteSupport(out.String()) {
		return false, "codex resume does not offer --remote and --remote-auth-token-env"
	}
	return true, ""
}

// managedCodexTUI is the operator TUI of one managed Codex session: the
// session log file the launcher's output moves to, and the TUI child once it
// runs.
type managedCodexTUI struct {
	stdin     *os.File
	logFile   *os.File
	logPath   string
	stopGrace time.Duration

	// attached is set once the TUI child runs; it gates every behavior that only
	// holds while a human is at the terminal.
	attached atomic.Bool

	// stopMu serializes stop: the connection-lost path and Close can both reach
	// it, and a second caller must find the child already reaped.
	stopMu sync.Mutex

	mu       sync.Mutex
	cmd      *exec.Cmd
	done     chan error
	reaped   chan struct{}
	stopping bool
	exited   bool

	// The sink pair is written by attach before the child's wait goroutine
	// starts and read after it; mu orders the other readers.
	sinkPtr *io.Writer
	prevLog *io.Writer

	closeLogOnce sync.Once
	finishOnce   sync.Once
}

// planOperatorTUI makes the attach decision before Start (REQ-MT-003,
// REQ-MT-004): any failed precondition prints exactly one notice and leaves the
// session headless. When the TUI is planned the session log file is opened now,
// because the App Server child's stderr is fixed when it starts.
func (s *managedCodexSession) planOperatorTUI(root, runID string, stdin io.Reader) {
	in, reason := s.operatorTUIPreconditions(stdin)
	if reason != "" {
		managedLogf(managedTUINoticeFormat, reason)
		return
	}
	file, path, err := openManagedTUILog(root, runID, s.label)
	if err != nil {
		managedLogf(managedTUINoticeFormat, "session log file unavailable: "+err.Error())
		return
	}
	managedLogf(managedTUILogFormat, path)
	s.tui = &managedCodexTUI{stdin: in, logFile: file, logPath: path, stopGrace: managedTUIStopGrace}
}

// operatorTUIPreconditions evaluates the three attach preconditions cheapest
// first, so an opt-out or a missing terminal never spawns the probe.
func (s *managedCodexSession) operatorTUIPreconditions(stdin io.Reader) (*os.File, string) {
	switch v := strings.ToLower(strings.TrimSpace(launchEnvValue(s.env, config.EnvMoaiFactoryManagedTUI))); v {
	case "0", "false", "off":
		return nil, fmt.Sprintf("disabled by %s=%s", config.EnvMoaiFactoryManagedTUI, v)
	}
	in, ok := stdin.(*os.File)
	if !ok || in == nil {
		return nil, "stdin is not a file"
	}
	if !managedTerminalCheck(in) {
		return nil, "stdin is not a terminal"
	}
	if !managedTerminalCheck(os.Stdout) {
		return nil, "stdout is not a terminal"
	}
	if supported, why := managedCodexRemoteProbe(s.program, s.env, s.dir); !supported {
		return nil, why
	}
	return in, ""
}

// openManagedTUILog opens the session log file under the project's
// .moai/logs/ directory (gitignored), mode 0600, append.
func openManagedTUILog(root, runID, label string) (*os.File, string, error) {
	dir := filepath.Join(root, ".moai", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, "", err
	}
	path := filepath.Join(dir, "factory-managed-"+managedLogNamePart(runID)+"-"+managedLogNamePart(label)+".log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, "", err
	}
	return f, path, nil
}

// managedLogNamePart keeps a run id or lane label usable as a file name part.
func managedLogNamePart(v string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			return r
		}
		return '-'
	}, v)
}

// appServerStderr is where the App Server child's stderr goes: the session log
// file when a TUI is planned (the terminal belongs to the TUI), else the
// terminal.
func (s *managedCodexSession) appServerStderr() io.Writer {
	if s.tui != nil && s.tui.logFile != nil {
		return s.tui.logFile
	}
	return os.Stderr
}

// managedCodexListenURL reads the --listen address out of the App Server
// command line the owner generated.
func managedCodexListenURL(args []string) string {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--listen" {
			return args[i+1]
		}
	}
	return ""
}

// AttachOperator starts the planned operator TUI (REQ-MT-001): after the
// priming turn, on the thread the broker endpoint is bound to. The token
// travels only as the value of an environment variable whose name is on the
// command line (REQ-MT-002); none of the launcher-generated approval overrides
// reach the TUI (REQ-MT-009).
func (s *managedCodexSession) AttachOperator() (<-chan error, bool) {
	t := s.tui
	if t == nil || s.cmd == nil || s.client == nil {
		return nil, false
	}
	return t.attach(s)
}

func (t *managedCodexTUI) attach(s *managedCodexSession) (<-chan error, bool) {
	t.mu.Lock()
	if t.cmd != nil {
		done := t.done
		t.mu.Unlock()
		return done, true
	}
	t.mu.Unlock()

	token, err := os.ReadFile(filepath.Join(s.tokenDir, "token"))
	url := managedCodexListenURL(s.cmd.Args)
	if err != nil || url == "" {
		t.giveUp("capability token or address unavailable")
		return nil, false
	}
	args := []string{"resume", "--remote", url, "--remote-auth-token-env", config.EnvMoaiFactoryAppServerToken}
	args = append(args, s.appArgs...)
	if s.model != "" {
		args = append(args, "-m", s.model)
	}
	args = append(args, s.threadID)
	cmd := managedCodexTUICommand(s.program, args...)
	cmd.Env = append(append([]string(nil), s.env...), config.EnvMoaiFactoryAppServerToken+"="+string(token))
	cmd.Dir = s.dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = t.stdin, os.Stdout, os.Stderr

	// The launcher stops writing to the terminal before the TUI draws on it.
	var sink io.Writer = t.logFile
	t.sinkPtr = &sink
	t.prevLog = managedLogOutput.Swap(t.sinkPtr)
	s.client.enableOperatorScoping(managedBusyWarnInterval)
	if err := cmd.Start(); err != nil {
		s.client.disableOperatorScoping()
		t.giveUp("TUI could not start: " + err.Error())
		return nil, false
	}
	done := make(chan error, 1)
	reaped := make(chan struct{})
	t.mu.Lock()
	t.cmd, t.done, t.reaped = cmd, done, reaped
	t.mu.Unlock()
	t.attached.Store(true)
	go t.wait(cmd, s.client, reaped)
	return done, true
}

// giveUp falls back to the headless session after a late attach failure: the
// terminal sink is restored, the notice is the one stderr line, and the App
// Server's stderr stays in the log file it already writes (REQ-MT-004).
func (t *managedCodexTUI) giveUp(reason string) {
	t.restoreLog()
	t.closeLog()
	managedLogf(managedTUINoticeFormat, reason)
}

// wait reaps the TUI child. The sink is restored the moment the child is gone,
// so the teardown lines that follow reach the terminal again (REQ-MT-006).
//
// When the TUI ends on its own (the owner did not ask it to stop) its exit
// status is recorded first and the App Server connection is closed after: a
// blocked turn/start write or a waiting turn can only be released by closing
// the connection, and the driver then finds the status already recorded and
// returns it instead of the connection error (REQ-MT-010). This goroutine
// calls only the connection's Close, never the session's, which stays
// single-goroutine and idempotent.
//
// @MX:WARN: [AUTO] goroutine that outlives the attach call; it ends only when the child exits, and stop waits on its reaped channel
// @MX:REASON: stop and Close must not return before this goroutine has reaped the child, or a TUI process could outlive the launcher
// @MX:SPEC: SPEC-FACTORY-MANAGED-TUI-001
func (t *managedCodexTUI) wait(cmd *exec.Cmd, client *managedCodexAppClient, reaped chan<- struct{}) {
	_ = cmd.Wait()
	t.mu.Lock()
	stopping := t.stopping
	t.exited = true
	t.mu.Unlock()
	t.restoreLog()
	if !stopping {
		t.finish(managedTUIExitError(cmd.ProcessState))
		client.closeConnection()
	}
	close(reaped)
}

// finish delivers the single end-of-TUI result to the driver; the first caller
// wins (the TUI's own exit, or the lost-connection error).
func (t *managedCodexTUI) finish(err error) {
	t.finishOnce.Do(func() {
		t.mu.Lock()
		done := t.done
		t.mu.Unlock()
		done <- err
	})
}

// connectionLost runs when the App Server connection closed while the TUI
// might still run: the TUI is stopped (interrupt, grace, kill) and the driver
// is told the connection is gone — only after the child is reaped, so the owner
// never returns while a TUI process is alive (REQ-MT-011).
//
// @MX:WARN: [AUTO] started from the connection's read goroutine exit and outlives it; it blocks until the TUI child is reaped
// @MX:REASON: the stop can wait the whole grace period, which must never run on the read goroutine's own exit path
// @MX:SPEC: SPEC-FACTORY-MANAGED-TUI-001
func (t *managedCodexTUI) connectionLost() {
	t.stop()
	t.finish(errManagedCodexConnectionClosed)
}

// tuiConnectionLost is the client's exit signal: the server-death monitor. It
// does nothing for a session without a running TUI, and nothing when the
// connection ended because the TUI exited or the owner asked it to stop.
func (s *managedCodexSession) tuiConnectionLost() {
	t := s.tui
	if t == nil || !t.attached.Load() {
		return
	}
	t.mu.Lock()
	over := t.exited || t.stopping
	t.mu.Unlock()
	if !over {
		go t.connectionLost()
	}
}

// newAppClient builds the App Server client with the operator-attach hooks:
// the seams are snapshotted here so no goroutine reads a package variable.
func (s *managedCodexSession) newAppClient(conn *websocket.Conn) *managedCodexAppClient {
	return &managedCodexAppClient{
		conn: conn, events: make(chan managedCodexAppReply, 32), done: make(chan struct{}),
		writeDeadline: managedWriteDeadline, writeBarrier: managedWriteBarrier, onExit: s.tuiConnectionLost,
	}
}

// closeConnection closes the WebSocket connection itself, which gorilla allows
// concurrently with a blocked write: the write returns an error and the read
// goroutine ends, releasing a waiting turn.
func (c *managedCodexAppClient) closeConnection() { _ = c.conn.Close() }

// stop interrupts the TUI, waits the grace period, kills it, and waits for the
// reap. Idempotent: the connection-lost path and Close can both reach it.
func (t *managedCodexTUI) stop() {
	t.stopMu.Lock()
	defer t.stopMu.Unlock()
	t.mu.Lock()
	cmd, reaped := t.cmd, t.reaped
	if cmd != nil {
		t.stopping = true
	}
	t.mu.Unlock()
	if cmd != nil {
		select {
		case <-reaped:
		default:
			// An interrupt lets a TUI that handles it restore the terminal;
			// where the platform cannot deliver one, the child is killed.
			if err := cmd.Process.Signal(os.Interrupt); err != nil {
				_ = cmd.Process.Kill()
			}
			select {
			case <-reaped:
			case <-time.After(t.stopGrace):
				_ = cmd.Process.Kill()
				<-reaped
			}
		}
	}
	t.restoreLog()
	t.closeLog()
}

// restoreLog puts the log destination back, but only if it is still ours.
func (t *managedCodexTUI) restoreLog() {
	if t.sinkPtr != nil {
		managedLogOutput.CompareAndSwap(t.sinkPtr, t.prevLog)
	}
}

func (t *managedCodexTUI) closeLog() {
	t.closeLogOnce.Do(func() {
		if t.logFile != nil {
			_ = t.logFile.Close()
		}
	})
}

// stopTUI is the first step of the session teardown: the TUI goes first so it
// does not render a dying server and the terminal returns before the launcher
// exits.
func (s *managedCodexSession) stopTUI() {
	if s.tui != nil {
		s.tui.stop()
	}
}

// Busy reports whether the thread has an active turn, whichever client started
// it. The driver claims broker messages only while it is false.
func (s *managedCodexSession) Busy() bool {
	if t := s.tui; t == nil || !t.attached.Load() || s.client == nil {
		return false
	}
	return s.client.busy(time.Now())
}

// ---- shared-thread bookkeeping of the App Server client (REQ-MT-007, REQ-MT-008)

// enableOperatorScoping turns on the rules that only hold while a human is at
// the terminal: foreign lifecycle frames stay off the owner's event channel and
// requests of turns the owner did not start are left to the operator.
func (c *managedCodexAppClient) enableOperatorScoping(busyWarn time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scoping, c.busyWarn = true, busyWarn
}

func (c *managedCodexAppClient) disableOperatorScoping() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scoping = false
}

// trackStartedLocked records a started turn of any client.
func (c *managedCodexAppClient) trackStartedLocked(id string) {
	if id == "" {
		return
	}
	if c.active == nil {
		c.active = map[string]struct{}{}
	}
	if len(c.active) == 0 {
		c.busySince = time.Now()
		c.lastWarn = c.busySince
	}
	c.active[id] = struct{}{}
}

func (c *managedCodexAppClient) markOwnedLocked(id string) {
	if id == "" {
		return
	}
	if c.owned == nil {
		c.owned = map[string]struct{}{}
	}
	c.owned[id] = struct{}{}
}

// forwardsLocked reports whether a lifecycle frame of turn id belongs on the
// owner's event channel.
func (c *managedCodexAppClient) forwardsLocked(id string) bool {
	if !c.scoping {
		return true
	}
	_, owned := c.owned[id]
	return owned
}

// trackTurnCompleted drops a completed turn from the active set and reports
// whether its completion frame belongs on the event channel.
func (c *managedCodexAppClient) trackTurnCompleted(params json.RawMessage) (forward bool) {
	id, _ := managedTurnFrame(params)
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.active, id)
	return c.forwardsLocked(id)
}

// noteTurnStartResponse claims the turn id a turn/start response carries for
// the owner while its window is open: the response is the one place a turn id
// is attributed to the owner's own call.
func (c *managedCodexAppClient) noteTurnStartResponse(result json.RawMessage) {
	var reply struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	if json.Unmarshal(result, &reply) != nil || reply.Turn.ID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.open {
		return
	}
	c.markOwnedLocked(reply.Turn.ID)
	if c.scoping && c.turnID == "" {
		c.turnID = reply.Turn.ID
	}
}

// busy reports whether any turn is active. It never clears by time: a turn that
// stays active longer than the warn interval only produces one log line per
// elapsed interval, because releasing a stale flag would steer a broker prompt
// into a legitimate long operator turn (known debt 13).
func (c *managedCodexAppClient) busy(now time.Time) bool {
	c.mu.Lock()
	if len(c.active) == 0 {
		c.mu.Unlock()
		return false
	}
	var line string
	if c.busyWarn > 0 && now.Sub(c.lastWarn) >= c.busyWarn {
		c.lastWarn = now
		line = fmt.Sprintf("Factory thread has had an active turn for %s; broker delivery stays deferred until it completes", now.Sub(c.busySince).Round(time.Millisecond))
	}
	c.mu.Unlock()
	if line != "" {
		managedLogf("%s", line)
	}
	return true
}

// managedTurnScopedRequests are the server request kinds whose params carry a
// turnId (schema-measured on codex-cli 0.160.0).
var managedTurnScopedRequests = map[string]struct{}{
	"item/commandExecution/requestApproval": {},
	"item/fileChange/requestApproval":       {},
	"item/tool/requestUserInput":            {},
	managedElicitationMethod:                {},
	"item/permissions/requestApproval":      {},
	"item/tool/call":                        {},
}

// leavesForOperator reports whether a server request is left unanswered so the
// operator answers it in the TUI: it names a turn the owner did not start. A
// request without a turn id keeps its fixed answer, and so does one that
// arrives while the owner's own turn/start is outstanding and its turn id is
// not yet known (benefit of the doubt goes to the owner's turn).
func (c *managedCodexAppClient) leavesForOperator(req managedCodexAppReply) bool {
	if _, scoped := managedTurnScopedRequests[req.Method]; !scoped {
		return false
	}
	var params struct {
		TurnID *string `json:"turnId"`
	}
	if json.Unmarshal(req.Params, &params) != nil || params.TurnID == nil || *params.TurnID == "" {
		return false
	}
	turn := *params.TurnID
	c.mu.Lock()
	_, owned := c.owned[turn]
	left := c.scoping && !owned && !(c.open && c.turnID == "")
	c.mu.Unlock()
	if left {
		managedLogf("Factory server request left for the operator: %q turn=%s", req.Method, turn)
	}
	return left
}

// managedTUIExitError maps the TUI's exit to the launcher's: nil for status 0,
// the TUI's own status otherwise, and 1 when a signal ended it (Go reports -1
// for a signaled process; no platform primitive is needed).
func managedTUIExitError(state *os.ProcessState) error {
	if state == nil {
		return &exitCodeError{code: 1, msg: "operator TUI ended without an exit status"}
	}
	switch code := state.ExitCode(); {
	case code == 0:
		return nil
	case code < 0:
		return &exitCodeError{code: 1, msg: "operator TUI ended by a signal"}
	default:
		return &exitCodeError{code: code, msg: fmt.Sprintf("operator TUI exited with status %d", code)}
	}
}
