package cli

// managed_codex_tui_test.go — SPEC-FACTORY-MANAGED-TUI-001 acceptance
// coverage (card t1408): the operator TUI attach of the managed Codex owner.
//
// Fake codex: ONE re-exec of the test binary plays every program the owner
// launches (as in production, where the App Server and the TUI are the same
// codex binary). The first argument picks the role: `app-server` (a fake App
// Server over a loopback WebSocket), `resume --help` (the capability probe),
// `resume --remote ...` (the fake operator TUI). All roles append to ONE log
// file so cross-process ordering assertions read a single sequence, and the
// fake App Server is scripted through an append-only control file (frames to
// send, a connection to close, an armed request to send before the next turn
// response). POSIX sh shim, like the other managed fixtures; the Windows lane
// proves the owner by cross build.
//
// Every blocking wait is bounded: a test that cannot see the behavior fails on
// a NAMED assertion within seconds instead of hanging.

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/spf13/cobra"
)

// Environment names of the fake codex. The helper runs only when the role env
// is set, so the normal test run skips it.
const (
	tuiFakeRoleEnv       = "T1408_FAKE_CODEX_ROLE"
	tuiFakeLogEnv        = "T1408_FAKE_CODEX_LOG"
	tuiFakeControlEnv    = "T1408_FAKE_CODEX_CONTROL"
	tuiFakeHelpEnv       = "T1408_FAKE_CODEX_HELP"
	tuiFakeHelpFileEnv   = "T1408_FAKE_CODEX_HELP_FILE"
	tuiFakeStatusesEnv   = "T1408_FAKE_CODEX_STATUSES"
	tuiFakeHoldFromEnv   = "T1408_FAKE_CODEX_HOLD_FROM"
	tuiFakeNoOpFramesEnv = "T1408_FAKE_CODEX_NO_OPERATOR_FRAMES"
	tuiFakeTUIModeEnv    = "T1408_FAKE_CODEX_TUI_MODE"
	tuiFakeRootEnv       = "T1408_FAKE_CODEX_ROOT"
	tuiFakeRunEnv        = "T1408_FAKE_CODEX_RUN"
	// tuiFakeTurnDelayEnv delays every launcher turn/completed by that many ms,
	// so a TUI started early is seen to start before the priming turn completes
	// (process start-up alone is faster than an instant fake turn).
	tuiFakeTurnDelayEnv = "T1408_FAKE_CODEX_TURN_DELAY_MS"
)

const (
	// tuiWatchdog bounds a blocking in-process call (the HARDEN-001 convention).
	tuiWatchdog = 5 * time.Second
	// tuiAttachWait bounds the wait for the fake TUI process to appear.
	tuiAttachWait = 10 * time.Second
	// tuiRunWait bounds a whole owner run.
	tuiRunWait = 20 * time.Second
	// tuiQuiet is the "exactly once / nothing happens" window.
	tuiQuiet = 500 * time.Millisecond
	// tuiResumeHelpFixture is the vendored `codex resume --help` text of
	// codex-cli 0.161.0.
	tuiResumeHelpFixture = "testdata/codex-0.161.0/resume-help.txt"
)

// ---------------------------------------------------------------- fake codex

func tuiFakeWatchParent() {
	parent := os.Getppid()
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			if os.Getppid() != parent {
				os.Exit(0)
			}
		}
	}()
}

type tuiFakeConn struct {
	c  *websocket.Conn
	mu sync.Mutex
}

func (c *tuiFakeConn) send(v any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.WriteJSON(v)
}

type tuiFakeServer struct {
	logPath string
	mu      sync.Mutex
	// launcher is the connection that initialized as the MoAI owner; tui is any
	// other client.
	launcher, tui *tuiFakeConn
	turns         int
	armed         []json.RawMessage
	opTurns       int
	statuses      []string
	holdFrom      int
	noOpFrames    bool
}

func (s *tuiFakeServer) log(line string) { fakeAppendLog(s.logPath, line) }

func (s *tuiFakeServer) sendLauncher(v any) bool {
	s.mu.Lock()
	c := s.launcher
	s.mu.Unlock()
	if c == nil {
		s.log("control-error no launcher connection")
		return false
	}
	c.send(v)
	return true
}

func (s *tuiFakeServer) command(line string) {
	cmd, arg, _ := strings.Cut(strings.TrimSpace(line), " ")
	switch cmd {
	case "frame":
		s.sendLauncher(json.RawMessage(arg))
	case "frame-tui":
		s.mu.Lock()
		c := s.tui
		s.mu.Unlock()
		if c != nil {
			c.send(json.RawMessage(arg))
		}
	case "arm-request":
		s.mu.Lock()
		s.armed = append(s.armed, json.RawMessage(arg))
		s.mu.Unlock()
	case "close-launcher":
		s.mu.Lock()
		c := s.launcher
		s.mu.Unlock()
		if c != nil {
			_ = c.c.Close()
		}
	case "ack":
		s.log("control-ack " + arg)
	case "exit":
		os.Exit(0)
	}
}

func (s *tuiFakeServer) tailControl(path string) {
	var off int64
	for {
		time.Sleep(10 * time.Millisecond)
		b, err := os.ReadFile(path)
		if err != nil || int64(len(b)) <= off {
			continue
		}
		chunk := b[off:]
		nl := bytes.LastIndexByte(chunk, '\n')
		if nl < 0 {
			continue
		}
		off += int64(nl) + 1
		for _, line := range strings.Split(string(chunk[:nl]), "\n") {
			s.command(line)
		}
	}
}

// actOnBrokerPrompt is the loopback model: when the launcher injects a prompt
// carrying broker metadata, read each body by claim token and record the
// receipt, as the real model is instructed to.
func (s *tuiFakeServer) actOnBrokerPrompt(prompt string) {
	root, run := os.Getenv(tuiFakeRootEnv), os.Getenv(tuiFakeRunEnv)
	if root == "" || run == "" || !strings.Contains(prompt, "message_id=") {
		return
	}
	store, err := factorymsg.Open(root, run)
	if err != nil {
		s.log("broker-error " + err.Error())
		return
	}
	defer func() { _ = store.Close() }()
	lane, err := store.Peer(context.Background(), fakeAppServerThreadID)
	if err != nil {
		s.log("broker-error peer " + err.Error())
		return
	}
	if err := actOnInboxPrompt(store, lane, prompt, func(b []byte) { s.log("body-read " + string(b)) }); err != nil {
		s.log("broker-error act " + err.Error())
	}
}

func (s *tuiFakeServer) launcherTurn(c *tuiFakeConn, id json.RawMessage, params json.RawMessage) {
	var p struct {
		Input []struct {
			Text string `json:"text"`
		} `json:"input"`
	}
	_ = json.Unmarshal(params, &p)
	text := ""
	if len(p.Input) > 0 {
		text = p.Input[0].Text
	}
	s.mu.Lock()
	s.turns++
	n := s.turns
	armed := s.armed
	s.armed = nil
	s.mu.Unlock()
	s.log("turn-prompt " + strings.ReplaceAll(text, "\n", `\n`))
	turnID := "fake-turn-" + strconv.Itoa(n)
	s.actOnBrokerPrompt(text)
	for _, a := range armed {
		c.send(a)
		time.Sleep(20 * time.Millisecond)
	}
	c.send(map[string]any{"id": id, "result": map[string]any{"turn": map[string]string{"id": turnID}}})
	c.send(map[string]any{"method": "turn/started", "params": map[string]any{"turn": map[string]string{"id": turnID}}})
	if s.holdFrom > 0 && n >= s.holdFrom {
		s.log("turn-held " + turnID)
		return
	}
	if ms, _ := strconv.Atoi(os.Getenv(tuiFakeTurnDelayEnv)); ms > 0 {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
	status := "completed"
	if n <= len(s.statuses) && s.statuses[n-1] != "" {
		status = s.statuses[n-1]
	}
	c.send(map[string]any{"method": "turn/completed", "params": map[string]any{"turn": map[string]string{"id": turnID, "status": status}}})
	s.log("turn-completed " + turnID)
}

func (s *tuiFakeServer) operatorTurn(c *tuiFakeConn, id json.RawMessage) {
	s.mu.Lock()
	s.opTurns++
	turnID := "op-" + strconv.Itoa(s.opTurns)
	launcher := s.launcher
	s.mu.Unlock()
	c.send(map[string]any{"id": id, "result": map[string]any{"turn": map[string]string{"id": turnID}}})
	started := map[string]any{"method": "turn/started", "params": map[string]any{"turn": map[string]string{"id": turnID}}}
	done := map[string]any{"method": "turn/completed", "params": map[string]any{"turn": map[string]string{"id": turnID, "status": "completed"}}}
	c.send(started)
	c.send(done)
	if launcher != nil && !s.noOpFrames {
		launcher.send(started)
		launcher.send(done)
	}
	s.log("operator-turn " + turnID)
}

func (s *tuiFakeServer) serve(c *tuiFakeConn) {
	for {
		var raw json.RawMessage
		if err := c.c.ReadJSON(&raw); err != nil {
			return
		}
		var req struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		_ = json.Unmarshal(raw, &req)
		if req.Method == "" && len(req.ID) > 0 {
			var compact bytes.Buffer
			if json.Compact(&compact, raw) == nil {
				s.log("reply " + compact.String())
			}
			continue
		}
		s.log("method " + req.Method)
		switch req.Method {
		case "initialize":
			var p struct {
				ClientInfo struct {
					Name string `json:"name"`
				} `json:"clientInfo"`
			}
			_ = json.Unmarshal(req.Params, &p)
			s.mu.Lock()
			if p.ClientInfo.Name == "moai_factory" {
				s.launcher = c
			} else {
				s.tui = c
			}
			s.mu.Unlock()
			c.send(map[string]any{"id": req.ID, "result": map[string]string{"status": "ok"}})
		case "thread/start":
			var p struct {
				Cwd string `json:"cwd"`
			}
			_ = json.Unmarshal(req.Params, &p)
			s.log("thread-cwd " + p.Cwd)
			c.send(map[string]any{"id": req.ID, "result": map[string]any{"thread": map[string]string{"id": fakeAppServerThreadID}}})
		case "thread/resume":
			var p struct {
				ThreadID string `json:"threadId"`
			}
			_ = json.Unmarshal(req.Params, &p)
			s.log("thread-resume " + p.ThreadID)
			c.send(map[string]any{"id": req.ID, "result": map[string]any{"thread": map[string]string{"id": p.ThreadID}}})
		case "thread/name/set":
			c.send(map[string]any{"id": req.ID, "result": map[string]any{}})
		case "turn/start":
			s.mu.Lock()
			isLauncher := c == s.launcher
			s.mu.Unlock()
			if isLauncher {
				s.launcherTurn(c, req.ID, req.Params)
			} else {
				s.operatorTurn(c, req.ID)
			}
		case "initialized":
			// notification
		default:
			c.send(map[string]any{"id": req.ID, "error": map[string]any{"code": -32601, "message": "fake codex: unknown method"}})
		}
	}
}

func tuiFakeServerMain(args []string) {
	logPath := os.Getenv(tuiFakeLogEnv)
	var listen, tokenFile string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--listen":
			i++
			if i < len(args) {
				listen = args[i]
			}
		case "--ws-token-file":
			i++
			if i < len(args) {
				tokenFile = args[i]
			}
		}
	}
	fakeAppendLog(logPath, "appserver-argv "+strings.Join(args, " "))
	fakeAppendLog(logPath, "appserver-pid "+strconv.Itoa(os.Getpid()))
	token, err := os.ReadFile(tokenFile)
	if err != nil || listen == "" {
		fakeAppendLog(logPath, "appserver-error missing token file or listen address")
		os.Exit(2)
	}
	fakeAppendLog(logPath, "appserver-token "+string(token))
	fakeAppendLog(logPath, "appserver-token-file "+tokenFile)
	fmt.Fprintln(os.Stderr, "fake-appserver-stderr-line")
	tuiFakeWatchParent()

	srv := &tuiFakeServer{logPath: logPath, noOpFrames: os.Getenv(tuiFakeNoOpFramesEnv) != ""}
	srv.statuses = strings.Split(os.Getenv(tuiFakeStatusesEnv), ",")
	srv.holdFrom, _ = strconv.Atoi(os.Getenv(tuiFakeHoldFromEnv))
	if path := os.Getenv(tuiFakeControlEnv); path != "" {
		go srv.tailControl(path)
	}
	ln, err := net.Listen("tcp", strings.TrimPrefix(listen, "ws://"))
	if err != nil {
		fakeAppendLog(logPath, "appserver-error listen "+err.Error())
		os.Exit(2)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+string(token) {
			fakeAppendLog(logPath, "auth-refused")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		fakeAppendLog(logPath, "auth-ok")
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		srv.serve(&tuiFakeConn{c: conn})
	})
	_ = (&http.Server{Handler: mux}).Serve(ln)
}

func tuiFakeHelpMain() {
	switch os.Getenv(tuiFakeHelpEnv) {
	case "lacks":
		fmt.Println("Usage: codex resume [OPTIONS] [SESSION_ID]\n\nOptions:\n  -h, --help  Print help")
	case "hang":
		time.Sleep(time.Hour)
	case "fail":
		fmt.Fprintln(os.Stderr, "error: fake codex cannot print help")
		os.Exit(2)
	case "empty":
	default:
		b, err := os.ReadFile(os.Getenv(tuiFakeHelpFileEnv))
		if err != nil {
			fmt.Fprintln(os.Stderr, "fake codex: help fixture unreadable:", err)
			os.Exit(2)
		}
		_, _ = os.Stdout.Write(b)
	}
	os.Exit(0)
}

func tuiFakeTUIMain(args []string) {
	logPath := os.Getenv(tuiFakeLogEnv)
	tuiFakeWatchParent()
	bound := "none"
	if root, run := os.Getenv(tuiFakeRootEnv), os.Getenv(tuiFakeRunEnv); root != "" && run != "" {
		store, err := factorymsg.Open(root, run)
		if err != nil {
			fakeAppendLog(logPath, "tui-broker-error open "+err.Error())
		} else {
			st, err := store.Status(context.Background())
			if err != nil {
				fakeAppendLog(logPath, "tui-broker-error status "+err.Error())
			}
			for _, lane := range st.Lanes {
				if lane.BindingState == factorymsg.BindingBound {
					bound = lane.SessionUUID
				}
			}
			_ = store.Close()
		}
	}
	fakeAppendLog(logPath, "tui-start bound="+bound)
	argv, _ := json.Marshal(args)
	fakeAppendLog(logPath, "tui-argv "+string(argv))
	fakeAppendLog(logPath, "tui-env-token "+os.Getenv(config.EnvMoaiFactoryAppServerToken))
	if cwd, err := os.Getwd(); err == nil {
		fakeAppendLog(logPath, "tui-cwd "+cwd)
	}
	fakeAppendLog(logPath, "tui-pid "+strconv.Itoa(os.Getpid()))

	mode := os.Getenv(tuiFakeTUIModeEnv)
	if strings.Contains(mode, "ignore-interrupt") {
		signal.Ignore(os.Interrupt)
	} else {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt)
		go func() {
			<-sigs
			fakeAppendLog(logPath, "tui-interrupted")
			os.Exit(3)
		}()
	}
	if strings.Contains(mode, "ws") {
		var remote string
		for i := 0; i+1 < len(args); i++ {
			if args[i] == "--remote" {
				remote = args[i+1]
			}
		}
		header := http.Header{"Authorization": []string{"Bearer " + os.Getenv(config.EnvMoaiFactoryAppServerToken)}}
		conn, _, err := websocket.DefaultDialer.Dial(remote, header)
		if err != nil {
			fakeAppendLog(logPath, "tui-ws-error "+err.Error())
		} else {
			_ = conn.WriteJSON(map[string]any{"id": 1, "method": "initialize", "params": map[string]any{"clientInfo": map[string]string{"name": "fake-tui"}}})
			_ = conn.WriteJSON(map[string]any{"id": 2, "method": "thread/resume", "params": map[string]any{"threadId": args[len(args)-1]}})
			fakeAppendLog(logPath, "tui-ws-resumed")
			if strings.Contains(mode, "op-turn") {
				_ = conn.WriteJSON(map[string]any{"id": 3, "method": "turn/start", "params": map[string]any{
					"threadId": args[len(args)-1], "input": []map[string]string{{"type": "text", "text": "operator-hello"}}}})
			}
			go func() {
				for {
					var discard json.RawMessage
					if conn.ReadJSON(&discard) != nil {
						return
					}
				}
			}()
		}
	}
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		line := sc.Text()
		fakeAppendLog(logPath, "tui-stdin "+line)
		switch line {
		case "bye":
			fakeAppendLog(logPath, "tui-exit 0")
			os.Exit(0)
		case "exit7":
			fakeAppendLog(logPath, "tui-exit 7")
			os.Exit(7)
		case "selfkill":
			if p, err := os.FindProcess(os.Getpid()); err == nil {
				_ = p.Kill()
			}
			time.Sleep(time.Minute)
		}
	}
	fakeAppendLog(logPath, "tui-stdin-eof")
	os.Exit(0)
}

// TestManagedCodexTUIFakeCodex is the re-exec helper: the program the owner
// launches for both `app-server` and `resume`.
func TestManagedCodexTUIFakeCodex(t *testing.T) {
	if os.Getenv(tuiFakeRoleEnv) != "1" {
		t.Skip("re-exec helper for the managed Codex TUI tests")
	}
	args := flag.Args()
	switch {
	case len(args) > 0 && args[0] == "app-server":
		tuiFakeServerMain(args[1:])
	case len(args) > 0 && args[0] == "resume":
		for _, a := range args {
			if a == "--help" {
				tuiFakeHelpMain()
			}
		}
		tuiFakeTUIMain(args)
	default:
		t.Fatalf("fake codex: unexpected arguments %v", args)
	}
}

// ---------------------------------------------------------------- test harness

func tuiFakeScript(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh fixture backend; the Windows lane proves the owner by cross build")
	}
	script := "#!/bin/sh\nexec \"" + os.Args[0] + "\" -test.run='^TestManagedCodexTUIFakeCodex$' -- \"$@\"\n"
	path := filepath.Join(t.TempDir(), "fake-codex.sh")
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func waitUntil(cond func() bool, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for {
		if cond() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func tuiLines(log, prefix string) []string {
	var out []string
	for _, l := range strings.Split(log, "\n") {
		if strings.HasPrefix(l, prefix) {
			out = append(out, strings.TrimPrefix(l, prefix))
		}
	}
	return out
}

func tuiJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func tuiPIDAlive(pid int) bool {
	return exec.Command("kill", "-0", strconv.Itoa(pid)).Run() == nil
}

// tuiTerminal sets the terminal predicate: the stdout stream is told apart from
// the operator stdin by identity with os.Stdout.
func tuiTerminal(t *testing.T, stdin, stdout bool) {
	t.Helper()
	orig := managedTerminalCheck
	managedTerminalCheck = func(f *os.File) bool {
		if f == os.Stdout {
			return stdout
		}
		return stdin
	}
	t.Cleanup(func() { managedTerminalCheck = orig })
}

// tuiSetDuration overrides one package-private duration seam for the test.
func tuiSetDuration(t *testing.T, seam *time.Duration, v time.Duration) {
	t.Helper()
	orig := *seam
	*seam = v
	t.Cleanup(func() { *seam = orig })
}

type tuiFake struct {
	t                   *testing.T
	root, run, dir      string
	program             string
	logPath, controlLog string
	env                 []string
	args                []string
	stdinR, stdinW      *os.File
	term                *hardenLogSink
}

// newTUIFake builds one fake factory project: a temp project root with an
// active run, the fake codex shim, an operator stdin pipe, and the terminal
// stand-in sink (the owner's pre-attach log destination).
func newTUIFake(t *testing.T, run string, extraEnv ...string) *tuiFake {
	t.Helper()
	program := tuiFakeScript(t)
	// Keep the child environment isolated: broker identity resolution needs real
	// Git, while every Codex process still uses the absolute fake program.
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	gitPath, err = filepath.Abs(gitPath)
	if err != nil {
		t.Fatal(err)
	}
	gitDir := filepath.Join(t.TempDir(), "git tools")
	if err := os.Mkdir(gitDir, 0o700); err != nil {
		t.Fatal(err)
	}
	gitShim := "#!/bin/sh\nexec " + shellQuote(gitPath) + " \"$@\"\n"
	if err := os.WriteFile(filepath.Join(gitDir, "git"), []byte(gitShim), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	activateManagedRun(t, root, run)
	term := hardenLogCapture(t)
	tuiTerminal(t, true, true)
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stdinR.Close(); _ = stdinW.Close() })
	helpFile, err := filepath.Abs(tuiResumeHelpFixture)
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	f := &tuiFake{
		t: t, root: root, run: run, dir: t.TempDir(), program: program,
		logPath: filepath.Join(tmp, "fake.log"), controlLog: filepath.Join(tmp, "control.log"),
		stdinR: stdinR, stdinW: stdinW, term: term,
	}
	if err := os.WriteFile(f.controlLog, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	f.env = append([]string{
		"PATH=" + gitDir,
		config.EnvFactoryRunID + "=" + run,
		config.EnvFactoryBackend + "=" + BackendCodex,
		config.EnvMoaiFactoryWorker + "=" + factory.FactoryLaneLabel(1),
		tuiFakeRoleEnv + "=1",
		tuiFakeLogEnv + "=" + f.logPath,
		tuiFakeControlEnv + "=" + f.controlLog,
		tuiFakeHelpFileEnv + "=" + helpFile,
		config.EnvMoaiFactoryManaged + "=1",
		// The fake processes open the same on-disk broker store as the owner: the
		// sandbox marker makes the re-executed test binary keep this MOAI_HOME
		// instead of creating its own (main_test.go sandboxMoaiHome).
		"MOAI_HOME=" + os.Getenv("MOAI_HOME"),
		moaiHomeSandboxEnv + "=" + os.Getenv("MOAI_HOME"),
	}, extraEnv...)
	return f
}

func (f *tuiFake) logText() string {
	b, _ := os.ReadFile(f.logPath)
	return string(b)
}

func (f *tuiFake) waitLog(substr string, d time.Duration) bool {
	return waitUntil(func() bool { return strings.Contains(f.logText(), substr) }, d)
}

func (f *tuiFake) control(cmd string) {
	f.t.Helper()
	fakeAppendLog(f.controlLog, cmd)
}

// controlSync appends an ack command and waits for the fake server to log it:
// every control line before it has then been executed.
func (f *tuiFake) controlSync(tag string) {
	f.t.Helper()
	f.control("ack " + tag)
	if !f.waitLog("control-ack "+tag, tuiWatchdog) {
		f.t.Errorf("the fake App Server did not execute the control lines before ack %s", tag)
	}
}

func (f *tuiFake) say(s string) {
	f.t.Helper()
	if _, err := f.stdinW.WriteString(s); err != nil {
		f.t.Errorf("write to the operator stdin pipe: %v", err)
	}
}

// sessionLog returns the owner's session log file ("" when none exists).
func (f *tuiFake) sessionLog() (path, text string) {
	matches, _ := filepath.Glob(filepath.Join(f.root, ".moai", "logs", "factory-managed-*.log"))
	if len(matches) == 0 {
		return "", ""
	}
	b, _ := os.ReadFile(matches[0])
	return matches[0], string(b)
}

func (f *tuiFake) runOwnerWith(stdin io.Reader) <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- runManagedFactoryCodex(f.program, append([]string{f.program}, f.args...), f.env, f.dir, stdin)
	}()
	return ch
}

func (f *tuiFake) runOwner() <-chan error { return f.runOwnerWith(f.stdinR) }

// tuiAwait waits for a result under d; ok is false when none arrived.
func tuiAwait(ch <-chan error, d time.Duration) (ok bool, err error) {
	select {
	case err = <-ch:
		return true, err
	case <-time.After(d):
		return false, nil
	}
}

// finish ends whatever the session is: the `bye` line ends an attached fake TUI
// and is a harmless operator line for a headless session, whose `/exit` ends it.
func (f *tuiFake) finish(ch <-chan error) {
	f.t.Helper()
	_, _ = f.stdinW.WriteString("bye\n/exit\n")
	if ok, _ := tuiAwait(ch, tuiRunWait); !ok {
		f.t.Errorf("the owner did not return within %s of the end-of-session lines", tuiRunWait)
	}
}

// newSession plans, builds and starts one managed Codex session by hand (the
// steps runManagedFactoryCodex performs), for the tests that drive the driver
// or the client directly.
func (f *tuiFake) newSession() *managedCodexSession {
	f.t.Helper()
	s, err := newManagedCodexSession(f.program, append([]string{f.program}, f.args...), f.env)
	if err != nil {
		f.t.Fatal(err)
	}
	s.dir = f.dir
	s.planOperatorTUI(f.root, f.run, f.stdinR)
	if err := s.Start(); err != nil {
		_ = s.Close()
		f.t.Fatalf("start the fake session: %v", err)
	}
	f.t.Cleanup(func() { _ = s.Close() })
	return s
}

// tuiClaims scripts the broker claim step.
type tuiClaims struct {
	calls   atomic.Int64
	release atomic.Bool // when set, claims hand out a batch
	always  bool        // hand out a batch on every call instead of once
	errOnce atomic.Bool // when set, the next call returns an inbox error once
	sent    atomic.Int64
	big     string // when non-empty the batch prompt is this text
}

func (c *tuiClaims) claim() ([]factorymsg.Claim, error) {
	c.calls.Add(1)
	if c.errOnce.CompareAndSwap(true, false) {
		return nil, errors.New("inbox boom")
	}
	if !c.release.Load() {
		return nil, nil
	}
	if c.always || c.sent.CompareAndSwap(0, 1) {
		c.sent.Add(1)
		return []factorymsg.Claim{{ClaimToken: "tok"}}, nil
	}
	return nil, nil
}

func (c *tuiClaims) prompt([]factorymsg.Claim) string {
	if c.big != "" {
		return c.big
	}
	return "BATCH"
}

// drive runs the real delivery driver over sess with scripted claims and a
// 10 ms idle tick.
func (f *tuiFake) drive(sess managedSession, claims *tuiClaims) <-chan error {
	f.t.Helper()
	idle := make(chan time.Time)
	stop := make(chan struct{})
	f.t.Cleanup(func() { close(stop) })
	go func() {
		for {
			select {
			case idle <- time.Now():
				time.Sleep(10 * time.Millisecond)
			case <-stop:
				return
			}
		}
	}()
	ch := make(chan error, 1)
	go func() { ch <- driveManagedFactorySession(sess, f.stdinR, idle, claims.claim, claims.prompt) }()
	return ch
}

func tuiTurnFrame(method, turn string) string {
	body := map[string]any{"turn": map[string]string{"id": turn}}
	if method == "turn/completed" {
		body = map[string]any{"turn": map[string]string{"id": turn, "status": "completed"}}
	}
	return "frame " + tuiJSON(map[string]any{"method": method, "params": body})
}

// tuiRequest builds one server request naming a turn ("" omits turnId, "null"
// sends a JSON null).
func tuiRequest(id, method, turn string) map[string]any {
	params := map[string]any{"threadId": fakeAppServerThreadID, "serverName": "other-server"}
	switch turn {
	case "":
	case "null":
		params["turnId"] = nil
	default:
		params["turnId"] = turn
	}
	return map[string]any{"id": id, "method": method, "params": params}
}

// reply returns the raw JSON of the launcher's answer to a server request id.
func (f *tuiFake) reply(id string) (string, bool) {
	for _, l := range tuiLines(f.logText(), "reply ") {
		if strings.Contains(l, `"id":"`+id+`"`) {
			return l, true
		}
	}
	return "", false
}

func (f *tuiFake) waitReply(id string, d time.Duration) (string, bool) {
	var out string
	ok := waitUntil(func() bool { r, ok := f.reply(id); out = r; return ok }, d)
	return out, ok
}

var (
	tuiTurnKinds = []string{
		"item/commandExecution/requestApproval",
		"item/fileChange/requestApproval",
		"item/tool/requestUserInput",
		"mcpServer/elicitation/request",
		"item/permissions/requestApproval",
		"item/tool/call",
	}
	tuiNoTurnKinds = []string{
		"account/chatgptAuthTokens/refresh",
		"attestation/generate",
		"applyPatchApproval",
		"execCommandApproval",
	}
)

func tuiListen(log string) string {
	for _, l := range tuiLines(log, "appserver-argv ") {
		fields := strings.Fields(l)
		for i, a := range fields {
			if a == "--listen" && i+1 < len(fields) {
				return fields[i+1]
			}
		}
	}
	return ""
}

func tuiServerPID(log string) int {
	for _, l := range tuiLines(log, "appserver-pid ") {
		if n, err := strconv.Atoi(strings.TrimSpace(l)); err == nil {
			return n
		}
	}
	return 0
}

func tuiTUIPID(log string) int {
	for _, l := range tuiLines(log, "tui-pid ") {
		if n, err := strconv.Atoi(strings.TrimSpace(l)); err == nil {
			return n
		}
	}
	return 0
}

// ---------------------------------------------------------------- AC-MT-001

// TestManagedCodexTUIAttachCommand pins the TUI child's command line, token
// delivery and working directory (REQ-MT-001, REQ-MT-002, REQ-MT-009).
func TestManagedCodexTUIAttachCommand(t *testing.T) {
	f := newTUIFake(t, "tui-attach-cmd")
	f.args = []string{"-c", `model_reasoning_effort="high"`, "-m", "fake-model"}
	done := f.runOwner()
	attached := f.waitLog("tui-argv", tuiAttachWait)
	if !attached {
		t.Errorf("the TUI child never started within %s (no tui-argv line in the fake log)", tuiAttachWait)
	}
	f.finish(done)
	log := f.logText()

	t.Run("argv_exact", func(t *testing.T) {
		listen := tuiListen(log)
		want := []string{"resume", "--remote", listen, "--remote-auth-token-env", config.EnvMoaiFactoryAppServerToken,
			"-c", `model_reasoning_effort="high"`, "-m", "fake-model", fakeAppServerThreadID}
		lines := tuiLines(log, "tui-argv ")
		if len(lines) != 1 {
			t.Fatalf("tui-argv lines = %d, want 1", len(lines))
		}
		var got []string
		if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
			t.Fatal(err)
		}
		if listen == "" || strings.Join(got, "\x00") != strings.Join(want, "\x00") {
			t.Errorf("TUI argv = %q, want %q", got, want)
		}
	})
	t.Run("token_not_in_argv", func(t *testing.T) {
		tokens := tuiLines(log, "appserver-token ")
		if len(tokens) != 1 || tokens[0] == "" {
			t.Fatalf("the fake App Server did not record its token: %q", tokens)
		}
		for _, prefix := range []string{"tui-argv ", "appserver-argv "} {
			for _, l := range tuiLines(log, prefix) {
				if strings.Contains(l, tokens[0]) {
					t.Errorf("the token value appears on a command line (%s)", prefix)
				}
			}
		}
		if len(tuiLines(log, "tui-argv ")) == 0 {
			t.Errorf("no TUI command line was logged, so its token handling is unobserved")
		}
	})
	t.Run("token_in_env", func(t *testing.T) {
		tokens := tuiLines(log, "appserver-token ")
		envTokens := tuiLines(log, "tui-env-token ")
		if len(tokens) != 1 || len(envTokens) != 1 || tokens[0] == "" || envTokens[0] != tokens[0] {
			t.Errorf("TUI env token %q, App Server token %q: want the same non-empty value", envTokens, tokens)
		}
	})
	t.Run("app_server_keeps_token_file", func(t *testing.T) {
		lines := tuiLines(log, "appserver-argv ")
		if len(lines) != 1 || !strings.Contains(lines[0], "--ws-token-file") {
			t.Errorf("App Server argv lost --ws-token-file: %q", lines)
		}
	})
	t.Run("no_generated_approval_args", func(t *testing.T) {
		tui := strings.Join(tuiLines(log, "tui-argv "), "")
		if tui == "" {
			t.Fatalf("no TUI command line was logged")
		}
		for _, banned := range []string{"mcp_servers.moai", "default_tools_approval_mode"} {
			if strings.Contains(tui, banned) {
				t.Errorf("TUI argv carries the generated override %q: %s", banned, tui)
			}
		}
		server := strings.Join(tuiLines(log, "appserver-argv "), "")
		if !strings.Contains(server, `mcp_servers.moai.default_tools_approval_mode="approve"`) {
			t.Errorf("App Server argv lost its generated approval overrides: %s", server)
		}
	})
	t.Run("operator_args_forwarded", func(t *testing.T) {
		tui := strings.Join(tuiLines(log, "tui-argv "), "")
		for _, want := range []string{`model_reasoning_effort=\"high\"`, `"-m","fake-model"`} {
			if !strings.Contains(tui, want) {
				t.Errorf("TUI argv misses the operator argument %q: %s", want, tui)
			}
		}
	})
	t.Run("cwd_matches_thread", func(t *testing.T) {
		cwds, threads := tuiLines(log, "tui-cwd "), tuiLines(log, "thread-cwd ")
		if len(cwds) != 1 || len(threads) != 1 {
			t.Fatalf("tui-cwd %q, thread-cwd %q: want one of each", cwds, threads)
		}
		a, _ := filepath.EvalSymlinks(cwds[0])
		b, _ := filepath.EvalSymlinks(threads[0])
		if a == "" || a != b {
			t.Errorf("TUI cwd %q differs from the thread cwd %q", cwds[0], threads[0])
		}
	})
}

// ---------------------------------------------------------------- AC-MT-002

func TestManagedCodexTUIStartsAfterPrimingAndBind(t *testing.T) {
	t.Run("order", func(t *testing.T) {
		f := newTUIFake(t, "tui-order", tuiFakeTurnDelayEnv+"=1500")
		f.env = append(f.env, tuiFakeRootEnv+"="+f.root, tuiFakeRunEnv+"="+f.run)
		done := f.runOwner()
		if !f.waitLog("tui-start", tuiAttachWait) {
			t.Errorf("the TUI never started within %s", tuiAttachWait)
		}
		f.finish(done)
		log := f.logText()
		completed := strings.Index(log, "turn-completed fake-turn-1")
		started := strings.Index(log, "tui-start")
		if completed < 0 || started < 0 || started < completed {
			t.Errorf("tui-start (offset %d) must follow the priming turn/completed (offset %d)", started, completed)
		}
		bound := tuiLines(log, "tui-start bound=")
		if len(bound) != 1 || bound[0] != fakeAppServerThreadID {
			t.Errorf("at TUI start the broker endpoint was bound to %q, want %q:\n%s", bound, fakeAppServerThreadID, log)
		}
	})
	t.Run("priming_failure_never_attaches", func(t *testing.T) {
		f := newTUIFake(t, "tui-prime-fail", tuiFakeStatusesEnv+"=failed")
		done := f.runOwner()
		ok, err := tuiAwait(done, tuiRunWait)
		if !ok {
			f.finish(done)
		}
		log := f.logText()
		if !strings.Contains(log, "turn-prompt ") {
			t.Fatalf("the priming turn never ran, so the failure path is unobserved:\n%s", log)
		}
		if ok && err == nil {
			t.Errorf("a failing priming turn must end the owner with an error")
		}
		if strings.Contains(log, "tui-start") {
			t.Errorf("the TUI started although the priming turn failed:\n%s", log)
		}
	})
}

// ---------------------------------------------------------------- AC-MT-003

func TestManagedCodexRemoteSupportProbe(t *testing.T) {
	t.Run("real_help_supported", func(t *testing.T) {
		b, err := os.ReadFile(tuiResumeHelpFixture)
		if err != nil {
			t.Fatal(err)
		}
		if !managedCodexRemoteSupport(string(b)) {
			t.Errorf("the vendored codex-cli 0.161.0 resume help must be read as supporting --remote and --remote-auth-token-env")
		}
	})
	t.Run("no_options_unsupported", func(t *testing.T) {
		help := "Usage: codex resume [OPTIONS] [SESSION_ID]\n\nOptions:\n  -h, --help  Print help\n"
		if managedCodexRemoteSupport(help) {
			t.Errorf("a help text without either option was read as supported")
		}
		// Each option alone is not enough, and --remote-auth-token-env must not
		// satisfy the --remote requirement.
		if managedCodexRemoteSupport("      --remote-auth-token-env <ENV_VAR>\n") {
			t.Errorf("--remote-auth-token-env alone was read as supported")
		}
		if managedCodexRemoteSupport("      --remote <ADDR>\n") {
			t.Errorf("--remote alone was read as supported")
		}
	})
	t.Run("empty_unsupported", func(t *testing.T) {
		if managedCodexRemoteSupport("") {
			t.Errorf("empty help output was read as supported")
		}
	})
}

func TestManagedCodexTUIPreconditionsAndFallback(t *testing.T) {
	cases := []struct {
		name   string
		reason string
		setup  func(t *testing.T, f *tuiFake)
		stdin  func(f *tuiFake) io.Reader
	}{
		{"stdin_not_terminal", "stdin is not a terminal", func(t *testing.T, f *tuiFake) { tuiTerminal(t, false, true) }, nil},
		{"stdin_not_file", "stdin is not a file", nil, func(f *tuiFake) io.Reader { return strings.NewReader("hello\n/exit\n") }},
		{"stdout_not_terminal", "stdout is not a terminal", func(t *testing.T, f *tuiFake) { tuiTerminal(t, true, false) }, nil},
		{"probe_lacks_options", "--remote", func(t *testing.T, f *tuiFake) { f.env = append(f.env, tuiFakeHelpEnv+"=lacks") }, nil},
		{"probe_times_out", "timed out", func(t *testing.T, f *tuiFake) {
			f.env = append(f.env, tuiFakeHelpEnv+"=hang")
			tuiSetDuration(t, &managedProbeTimeout, 300*time.Millisecond)
		}, nil},
		{"probe_fails_to_run", "failed", func(t *testing.T, f *tuiFake) { f.env = append(f.env, tuiFakeHelpEnv+"=fail") }, nil},
		{"opt_out_0", config.EnvMoaiFactoryManagedTUI, func(t *testing.T, f *tuiFake) { f.env = append(f.env, config.EnvMoaiFactoryManagedTUI+"=0") }, nil},
		{"opt_out_false", config.EnvMoaiFactoryManagedTUI, func(t *testing.T, f *tuiFake) { f.env = append(f.env, config.EnvMoaiFactoryManagedTUI+"=false") }, nil},
		{"opt_out_off", config.EnvMoaiFactoryManagedTUI, func(t *testing.T, f *tuiFake) { f.env = append(f.env, config.EnvMoaiFactoryManagedTUI+"=off") }, nil},
		{"opt_out_uppercase_padded", config.EnvMoaiFactoryManagedTUI, func(t *testing.T, f *tuiFake) { f.env = append(f.env, config.EnvMoaiFactoryManagedTUI+"=  OFF ") }, nil},
		{"tui_start_fails", "could not start", func(t *testing.T, f *tuiFake) {
			orig := managedCodexTUICommand
			managedCodexTUICommand = func(string, ...string) *exec.Cmd { return exec.Command(filepath.Join(t.TempDir(), "no-such-tui")) }
			t.Cleanup(func() { managedCodexTUICommand = orig })
		}, nil},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			f := newTUIFake(t, "tui-fallback-"+strings.ReplaceAll(tc.name, "_", "-"))
			if tc.setup != nil {
				tc.setup(t, f)
			}
			stdin := io.Reader(f.stdinR)
			if tc.stdin != nil {
				stdin = tc.stdin(f)
			} else {
				f.say("hello\n/exit\n")
			}
			done := f.runOwnerWith(stdin)
			ok, err := tuiAwait(done, tuiRunWait)
			if !ok {
				t.Fatalf("the headless fallback did not end on /exit within %s", tuiRunWait)
			}
			if err != nil {
				t.Errorf("a fallback must not change the exit status: %v", err)
			}
			log := f.logText()
			if !strings.Contains(log, "turn-prompt hello") {
				t.Errorf("an operator stdin line must still become a turn in the headless fallback:\n%s", log)
			}
			if strings.Contains(log, "tui-start") {
				t.Errorf("a TUI ran although a precondition failed")
			}
			term := f.term.snapshot()
			notices := 0
			for _, l := range strings.Split(term, "\n") {
				if strings.Contains(l, "operator TUI not attached") {
					notices++
					if !strings.Contains(l, tc.reason) {
						t.Errorf("notice %q does not name the reason %q", l, tc.reason)
					}
				}
			}
			if notices != 1 {
				t.Errorf("stderr holds %d fallback notices, want exactly 1:\n%s", notices, term)
			}
			if tc.name == "tui_start_fails" {
				path, text := f.sessionLog()
				if !strings.Contains(text, "fake-appserver-stderr-line") {
					t.Errorf("after a failed TUI start the App Server stderr must be in the session log file %q:\n%s", path, text)
				}
				if strings.Contains(term, "fake-appserver-stderr-line") {
					t.Errorf("the App Server stderr reached the terminal stand-in")
				}
				if n := strings.Count(term, "log file:"); n != 1 {
					t.Errorf("the log file path was printed %d times, want once:\n%s", n, term)
				}
			}
		})
	}
}

// ---------------------------------------------------------------- AC-MT-004

func TestManagedCodexTUIOwnsOperatorStdin(t *testing.T) {
	f := newTUIFake(t, "tui-stdin")
	done := f.runOwner()
	if !f.waitLog("tui-start", tuiAttachWait) {
		t.Errorf("the TUI never started within %s", tuiAttachWait)
	}
	f.say("hello\n/exit\n")
	f.waitLog("tui-stdin /exit", 2*time.Second)
	ok, err := tuiAwait(done, tuiQuiet)
	if ok {
		t.Errorf("the session ended on /exit while the TUI owns the terminal")
	} else {
		f.say("bye\n")
		ok, err = tuiAwait(done, tuiRunWait)
		if !ok {
			t.Errorf("the session did not end after the TUI exited")
			f.finish(done)
		}
	}
	if ok && err != nil {
		t.Errorf("TUI exit 0 must end the session with nil, got %v", err)
	}
	log := f.logText()
	for _, want := range []string{"tui-stdin hello", "tui-stdin /exit"} {
		if !strings.Contains(log, want) {
			t.Errorf("the fake TUI did not read %q from the shared stdin:\n%s", want, log)
		}
	}
	if strings.Contains(log, "turn-prompt hello") {
		t.Errorf("the launcher turned an operator stdin line into a turn while the TUI is attached")
	}
}

// ---------------------------------------------------------------- AC-MT-005

func TestManagedCodexTUIKeepsTerminalClean(t *testing.T) {
	f := newTUIFake(t, "tui-clean", tuiFakeStatusesEnv+"=completed,failed")
	sess := f.newSession()
	claims := &tuiClaims{}
	done := f.drive(sess, claims)
	if !f.waitLog("tui-start", tuiAttachWait) {
		t.Errorf("the TUI never started within %s", tuiAttachWait)
	}
	// A server request the launcher answers (it names the owned priming turn).
	f.control("frame " + tuiJSON(tuiRequest("clean-1", "item/commandExecution/requestApproval", "fake-turn-1")))
	f.waitReply("clean-1", tuiWatchdog)
	// A turn failure: the second launcher turn completes as failed.
	claims.release.Store(true)
	f.waitLog("turn-completed fake-turn-2", tuiWatchdog)
	// A driver inbox error.
	claims.errOnce.Store(true)
	waitUntil(func() bool { _, text := f.sessionLog(); return strings.Contains(text, "Factory inbox:") }, tuiWatchdog)

	path, text := f.sessionLog()
	term := f.term.snapshot()

	t.Run("attached_lines_in_file", func(t *testing.T) {
		if path == "" {
			t.Fatalf("no session log file exists under .moai/logs")
		}
		for _, want := range []string{"fake-appserver-stderr-line", "Factory server request answered", "Factory turn failed", "Factory inbox:"} {
			if !strings.Contains(text, want) {
				t.Errorf("session log file %s misses %q:\n%s", path, want, text)
			}
		}
	})
	t.Run("terminal_only_has_path_line", func(t *testing.T) {
		lines := strings.Split(strings.TrimSpace(term), "\n")
		if len(lines) != 1 || !strings.Contains(lines[0], "log file") || (path != "" && !strings.Contains(lines[0], path)) {
			t.Errorf("terminal stderr must hold exactly the one line naming the log file, got:\n%s", term)
		}
	})
	t.Run("post_reap_line_on_terminal", func(t *testing.T) {
		sess.stopTUI()
		f.control("frame " + tuiJSON(tuiRequest("clean-2", "item/fileChange/requestApproval", "fake-turn-1")))
		f.waitReply("clean-2", tuiWatchdog)
		reachedTerminal := waitUntil(func() bool { return strings.Contains(f.term.snapshot(), "item/fileChange/requestApproval") }, tuiWatchdog)
		_, after := f.sessionLog()
		if !reachedTerminal {
			t.Errorf("the owner log line written after the TUI was reaped did not reach the terminal stand-in:\n%s", f.term.snapshot())
		}
		if strings.Contains(after, "item/fileChange/requestApproval") {
			t.Errorf("the owner log line written after the TUI was reaped still went to the session log file")
		}
		if !strings.Contains(after, "fake-appserver-stderr-line") {
			t.Errorf("the App Server child's stderr must stay in the file")
		}
	})
	_ = done
}

// ---------------------------------------------------------------- AC-MT-006

func tuiCountPrompts(f *tuiFake) int { return len(tuiLines(f.logText(), "turn-prompt ")) }

func TestManagedCodexTUIDefersDeliveryWhileBusy(t *testing.T) {
	t.Run("deferred_then_delivered", func(t *testing.T) {
		f := newTUIFake(t, "tui-busy-defer")
		sess := f.newSession()
		claims := &tuiClaims{}
		done := f.drive(sess, claims)
		if !f.waitLog("tui-start", tuiAttachWait) {
			t.Errorf("the TUI never started within %s", tuiAttachWait)
		}
		f.control(tuiTurnFrame("turn/started", "op-1"))
		busy := waitUntil(sess.Busy, 3*time.Second)
		if !busy {
			t.Errorf("the owner never saw the operator-started turn as busy")
		}
		claims.release.Store(true)
		time.Sleep(50 * time.Millisecond)
		before := claims.calls.Load()
		time.Sleep(tuiQuiet)
		if busy && claims.calls.Load() != before {
			t.Errorf("the driver claimed broker messages %d times while the thread was busy", claims.calls.Load()-before)
		}
		if tuiCountPrompts(f) != 1 {
			t.Errorf("a launcher turn started while the operator turn ran (prompts: %d)", tuiCountPrompts(f))
		}
		f.control(tuiTurnFrame("turn/completed", "op-1"))
		if !f.waitLog("turn-prompt BATCH", tuiAttachWait) {
			t.Errorf("the deferred batch was not delivered after the operator turn completed")
		}
		time.Sleep(tuiQuiet)
		if n := len(tuiLines(f.logText(), "turn-prompt BATCH")); n != 1 {
			t.Errorf("the batch was delivered %d times, want exactly 1", n)
		}
		f.say("bye\n/exit\n")
		_, _ = tuiAwait(done, tuiRunWait)
	})
	t.Run("frames_not_delivered", func(t *testing.T) {
		// P6 false: the launcher's connection never receives lifecycle frames of
		// turns the TUI starts. The launcher must neither deadlock nor hang.
		f := newTUIFake(t, "tui-busy-noframes", tuiFakeNoOpFramesEnv+"=1", tuiFakeTUIModeEnv+"=ws,op-turn")
		sess := f.newSession()
		claims := &tuiClaims{}
		claims.release.Store(true)
		done := f.drive(sess, claims)
		if !f.waitLog("turn-prompt BATCH", tuiAttachWait) {
			t.Errorf("without lifecycle frames of operator turns the launcher must still deliver (degraded path)")
		}
		f.say("bye\n/exit\n")
		if ok, _ := tuiAwait(done, tuiRunWait); !ok {
			t.Errorf("the driver hung after the degraded delivery")
		}
	})
}

func TestManagedCodexReaderSurvivesOperatorTurns(t *testing.T) {
	f := newTUIFake(t, "tui-reader")
	sess := f.newSession()
	claims := &tuiClaims{}
	done := f.drive(sess, claims)
	if !f.waitLog("tui-start", tuiAttachWait) {
		t.Errorf("the TUI never started within %s", tuiAttachWait)
	}
	// 40 operator turns between launcher turns: far more frame pairs than the
	// 32-slot event channel holds, with no consumer.
	for i := 1; i <= 40; i++ {
		id := "op-" + strconv.Itoa(i)
		f.control(tuiTurnFrame("turn/started", id))
		f.control(tuiTurnFrame("turn/completed", id))
	}
	f.control("frame " + tuiJSON(tuiRequest("reader-1", "item/commandExecution/requestApproval", "fake-turn-1")))
	if _, ok := f.waitReply("reader-1", tuiWatchdog); !ok {
		t.Errorf("a server request after 40 operator turns was not answered within %s: the reader stalled behind the event channel", tuiWatchdog)
	}
	f.say("bye\n/exit\n")
	_, _ = tuiAwait(done, tuiWatchdog)
}

func TestManagedCodexBusyLongTurnKeepsDeferring(t *testing.T) {
	tuiSetDuration(t, &managedBusyWarnInterval, 100*time.Millisecond)
	f := newTUIFake(t, "tui-busy-long")
	sess := f.newSession()
	claims := &tuiClaims{}
	claims.release.Store(true)
	done := f.drive(sess, claims)
	if !f.waitLog("tui-start", tuiAttachWait) {
		t.Errorf("the TUI never started within %s", tuiAttachWait)
	}
	// The operator turn starts and never completes.
	f.control(tuiTurnFrame("turn/started", "op-1"))
	waitUntil(sess.Busy, 3*time.Second)
	promptsBefore := tuiCountPrompts(f)
	warned := waitUntil(func() bool {
		_, text := f.sessionLog()
		return strings.Count(text, "active turn for") >= 2
	}, tuiWatchdog)
	if !warned {
		_, text := f.sessionLog()
		t.Errorf("a busy turn that outlasts the warn interval must log one line per elapsed interval (want at least 2):\n%s", text)
	}
	time.Sleep(tuiQuiet)
	if tuiCountPrompts(f) != promptsBefore {
		t.Errorf("the busy flag cleared by time: a launcher turn started while the operator turn was still active")
	}
	f.control(tuiTurnFrame("turn/completed", "op-1"))
	if !f.waitLog("turn-prompt BATCH", tuiAttachWait) {
		t.Errorf("delivery did not resume when the long turn completed")
	}
	f.say("bye\n/exit\n")
	_, _ = tuiAwait(done, tuiRunWait)
}

// ---------------------------------------------------------------- AC-CONF-006

// tuiReplayFrame0161 builds one id-less notification frame as the wire carries
// it. The REPLAY payload sets below are derived from the codex-cli 0.161.0
// binary's own generated protocol schemas — the same generator output the
// vendored fixtures come from (v2/ItemStartedNotification.json,
// v2/ItemCompletedNotification.json, v2/TurnCompletedNotification.json,
// v2/ThreadTokenUsageUpdatedNotification.json, v2/TurnDiffUpdatedNotification.json;
// rust-v0.161.0) — NOT from this fake server's own responses, which AC-CONF-006
// refuses as provenance (a self-generated circle would pass invented fields).
// The adapter models none of these payloads; REQ-CONF-005 requires a resumed
// thread replaying them to be consumed without a decode failure and without a
// marked turn failure.
func tuiReplayFrame0161(method string, params map[string]any) string {
	return tuiJSON(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

// tuiReplayBurst0161 is the resumed-thread replay: history items and turn
// lifecycle frames of a thread resumed on another client, streamed to the
// owner's connection while it is mid-session.
func tuiReplayBurst0161() []string {
	return []string{
		// item/started of a replayed agentMessage item, carrying every extended
		// field the 0.161.0 ThreadItem::agentMessage schema names (phase,
		// memoryCitation, delivery, questions).
		tuiReplayFrame0161("item/started", map[string]any{
			"item": map[string]any{
				"type":           "agentMessage",
				"id":             "replay-item-1",
				"text":           "replayed answer from the resumed thread",
				"phase":          "final_answer",
				"memoryCitation": map[string]any{"entries": []any{}, "threadIds": []any{}},
				"delivery":       "async",
				"questions":      []any{},
			},
			"threadId":    "replay-thread",
			"turnId":      "replay-turn-1",
			"startedAtMs": 1728441600000,
		}),
		// item/completed of a replayed commandExecution item: the 0.161.0 field
		// set (commandActions, aggregatedOutput, durationMs, pluginId, ...).
		tuiReplayFrame0161("item/completed", map[string]any{
			"item": map[string]any{
				"type":             "commandExecution",
				"id":               "replay-item-2",
				"command":          "ls",
				"cwd":              "/tmp",
				"source":           "agent",
				"status":           "completed",
				"commandActions":   []any{map[string]any{"type": "listFiles", "command": "ls", "path": "/tmp"}},
				"aggregatedOutput": "",
				"exitCode":         0,
				"durationMs":       12,
				"pluginId":         nil,
				"scriptPath":       nil,
				"processId":        nil,
			},
			"threadId":      "replay-thread",
			"turnId":        "replay-turn-1",
			"completedAtMs": 1728441601000,
		}),
		// The replayed foreign turn's own completion: the 0.161.0 Turn shape
		// (id/items/status required; startedAt/completedAt/durationMs optional).
		// The owner must neither forward it nor count it as its own.
		tuiReplayFrame0161("turn/completed", map[string]any{
			"threadId": "replay-thread",
			"turn": map[string]any{
				"id": "replay-turn-1", "status": "completed", "items": []any{},
				"startedAt": nil, "completedAt": nil, "durationMs": nil,
			},
		}),
		// A 0.161.0 notification method the adapter has never answered or
		// tracked (ThreadTokenUsageUpdatedNotification).
		tuiReplayFrame0161("thread/tokenUsage/updated", map[string]any{
			"threadId": "replay-thread",
			"turnId":   "replay-turn-1",
			"tokenUsage": map[string]any{
				"total": map[string]any{"inputTokens": 1, "cachedInputTokens": 0, "outputTokens": 2, "reasoningOutputTokens": 0, "totalTokens": 3},
				"last":  map[string]any{"inputTokens": 1, "cachedInputTokens": 0, "outputTokens": 2, "reasoningOutputTokens": 0, "totalTokens": 3},
			},
		}),
		tuiReplayFrame0161("turn/diff/updated", map[string]any{
			"threadId": "replay-thread", "turnId": "replay-turn-1", "diff": "",
		}),
	}
}

// TestManagedCodexTUIConsumesResumedThreadReplayFrames pins REQ-CONF-005: a
// resumed thread's replay history — frames carrying fields the adapter does
// not model (#49599 class) — is consumed without a decode failure and without
// a marked turn failure, and a normal owner turn issued AFTER the replay burst
// still completes (the positive control that proves consumption actually
// happened, not merely that the frames were dropped by a dead reader).
func TestManagedCodexTUIConsumesResumedThreadReplayFrames(t *testing.T) {
	t.Run("replay_frames_keep_the_owner_turn_alive", func(t *testing.T) {
		f := newTUIFake(t, "tui-replay-consume")
		sess := f.newSession()
		claims := &tuiClaims{}
		done := f.drive(sess, claims)
		if !f.waitLog("tui-start", tuiAttachWait) {
			t.Errorf("the TUI never started within %s", tuiAttachWait)
		}
		for _, frame := range tuiReplayBurst0161() {
			f.control("frame " + frame)
		}
		f.controlSync("replay-burst-sent")
		// The replay frames must leave the read loop live: a server request
		// issued right after the burst is still answered.
		f.control("frame " + tuiJSON(tuiRequest("replay-alive-1", "item/commandExecution/requestApproval", "fake-turn-1")))
		if _, ok := f.waitReply("replay-alive-1", tuiWatchdog); !ok {
			t.Errorf("a server request after the replay burst was not answered: the reader died on the replay frames")
		}
		// Positive control: the owner's own turn issued after the burst still
		// completes.
		claims.release.Store(true)
		if !f.waitLog("turn-completed fake-turn-2", tuiWatchdog) {
			t.Errorf("the owner's own turn after the replay burst never completed:\n%s", f.logText())
		}
		time.Sleep(tuiQuiet)
		if _, text := f.sessionLog(); strings.Contains(text, "Factory turn failed") {
			t.Errorf("replay frames marked a Factory turn failed:\n%s", text)
		}
		f.say("bye\n/exit\n")
		if ok, err := tuiAwait(done, tuiRunWait); !ok {
			t.Errorf("the driver hung after the replay burst")
		} else if err != nil {
			t.Errorf("the driver errored after the replay burst: %v", err)
		}
	})
	t.Run("failed_own_turn_still_marks_Factory_turn_failed", func(t *testing.T) {
		// Canary: the same replay burst with the owner's own turn scripted to
		// fail. The tolerance sub-test asserts the ABSENCE of "Factory turn
		// failed" — this sub-test observes its PRESENCE on a known failing
		// input, so the absence assertion is wired to a live signal.
		f := newTUIFake(t, "tui-replay-canary", tuiFakeStatusesEnv+"=completed,failed")
		sess := f.newSession()
		claims := &tuiClaims{}
		done := f.drive(sess, claims)
		if !f.waitLog("tui-start", tuiAttachWait) {
			t.Errorf("the TUI never started within %s", tuiAttachWait)
		}
		for _, frame := range tuiReplayBurst0161() {
			f.control("frame " + frame)
		}
		f.controlSync("replay-burst-sent")
		claims.release.Store(true)
		if !f.waitLog("turn-completed fake-turn-2", tuiWatchdog) {
			t.Errorf("the scripted-failure turn never completed:\n%s", f.logText())
		}
		if !waitUntil(func() bool {
			_, text := f.sessionLog()
			return strings.Contains(text, "Factory turn failed")
		}, tuiWatchdog) {
			_, text := f.sessionLog()
			t.Errorf("a failed own turn did not mark Factory turn failed — the tolerance assertion above would pass vacuously:\n%s", text)
		}
		f.say("bye\n/exit\n")
		_, _ = tuiAwait(done, tuiRunWait)
	})
}

// ---------------------------------------------------------------- AC-MT-007

func TestManagedCodexServerRequestScopingWithTUI(t *testing.T) {
	// attachedSession starts a session with the TUI attached and the priming
	// turn done, so fake-turn-1 is an owned turn.
	attachedSession := func(t *testing.T, run string, extraEnv ...string) (*tuiFake, *managedCodexSession) {
		f := newTUIFake(t, run, extraEnv...)
		sess := f.newSession()
		if err := sess.DeliverTurn(managedPrimingPrompt); err != nil {
			t.Fatalf("priming turn: %v", err)
		}
		_, attached := sess.AttachOperator()
		if attached && !f.waitLog("tui-start", tuiAttachWait) {
			t.Errorf("the TUI never started")
		}
		if !attached {
			t.Errorf("the session did not attach the TUI")
		}
		return f, sess
	}

	f, sess := attachedSession(t, "tui-scoping")

	t.Run("owned_turn_answered", func(t *testing.T) {
		for i, kind := range tuiTurnKinds {
			id := "own-" + strconv.Itoa(i)
			f.control("frame " + tuiJSON(tuiRequest(id, kind, "fake-turn-1")))
			if _, ok := f.waitReply(id, tuiWatchdog); !ok {
				t.Errorf("%s naming the launcher's own turn was not answered", kind)
			}
		}
	})
	t.Run("operator_turn_unanswered", func(t *testing.T) {
		for i, kind := range tuiTurnKinds {
			f.control("frame " + tuiJSON(tuiRequest("op-req-"+strconv.Itoa(i), kind, "op-1")))
		}
		time.Sleep(tuiQuiet)
		for i, kind := range tuiTurnKinds {
			if r, ok := f.reply("op-req-" + strconv.Itoa(i)); ok {
				t.Errorf("%s naming an operator-started turn was answered by the launcher: %s", kind, r)
			}
		}
		// Positive control: the reader is alive and answers an owned request.
		f.control("frame " + tuiJSON(tuiRequest("op-control", "item/commandExecution/requestApproval", "fake-turn-1")))
		if _, ok := f.waitReply("op-control", tuiWatchdog); !ok {
			t.Errorf("positive control failed: the reader answered nothing after the operator-turn requests")
		}
	})
	t.Run("armed_window_answered", func(t *testing.T) {
		// Requests that arrive while the launcher's own turn/start is
		// outstanding and its turn id is not yet known are answered.
		for i, kind := range tuiTurnKinds {
			f.control("arm-request " + tuiJSON(tuiRequest("armed-"+strconv.Itoa(i), kind, "op-9")))
		}
		f.controlSync("armed")
		if err := sess.DeliverTurn("second"); err != nil {
			t.Fatalf("second launcher turn: %v", err)
		}
		for i, kind := range tuiTurnKinds {
			if _, ok := f.waitReply("armed-"+strconv.Itoa(i), tuiWatchdog); !ok {
				t.Errorf("%s arriving while the owner's turn/start was outstanding was not answered", kind)
			}
		}
	})
	t.Run("no_turn_id_kinds_answered", func(t *testing.T) {
		for i, kind := range tuiNoTurnKinds {
			id := "noid-" + strconv.Itoa(i)
			f.control("frame " + tuiJSON(tuiRequest(id, kind, "")))
			if _, ok := f.waitReply(id, tuiWatchdog); !ok {
				t.Errorf("%s (no turnId) was not answered", kind)
			}
		}
		f.control("frame " + tuiJSON(tuiRequest("noid-null", "mcpServer/elicitation/request", "null")))
		if _, ok := f.waitReply("noid-null", tuiWatchdog); !ok {
			t.Errorf("an elicitation with a null turnId was not answered")
		}
	})
	t.Run("never_accepts", func(t *testing.T) {
		replies := tuiLines(f.logText(), "reply ")
		if len(replies) == 0 {
			t.Fatalf("no launcher answer was observed")
		}
		for _, r := range replies {
			low := strings.ToLower(r)
			for _, banned := range []string{"accept", "approved", "allow", `"decision":"approve`} {
				if strings.Contains(low, banned) {
					t.Errorf("an answer carries an accepting decision (%q): %s", banned, r)
				}
			}
		}
	})
	t.Run("detached_unchanged", func(t *testing.T) {
		d := newTUIFake(t, "tui-scoping-detached", config.EnvMoaiFactoryManagedTUI+"=off")
		dsess := d.newSession()
		if err := dsess.DeliverTurn(managedPrimingPrompt); err != nil {
			t.Fatal(err)
		}
		d.control("frame " + tuiJSON(tuiRequest("det-1", "item/commandExecution/requestApproval", "op-1")))
		reply, ok := d.waitReply("det-1", tuiWatchdog)
		if !ok || !strings.Contains(reply, `"decision":"decline"`) {
			t.Errorf("with no TUI attached a request naming another turn must be declined as before, got %q (answered=%v)", reply, ok)
		}
	})
}

// ---------------------------------------------------------------- AC-MT-008

func tuiExitCode(err error) (int, bool) { return ResolveExitCode(err) }

func TestManagedCodexTUIExitEndsSession(t *testing.T) {
	runExit := func(t *testing.T, run, line string) (*tuiFake, error) {
		f := newTUIFake(t, run)
		done := f.runOwner()
		if !f.waitLog("tui-start", tuiAttachWait) {
			t.Errorf("the TUI never started within %s", tuiAttachWait)
		}
		f.say(line + "\n/exit\n")
		ok, err := tuiAwait(done, tuiRunWait)
		if !ok {
			f.say("bye\n")
			t.Errorf("the owner did not return after the TUI exited")
		}
		return f, err
	}
	reaped := func(t *testing.T, f *tuiFake) {
		t.Helper()
		log := f.logText()
		pid := tuiServerPID(log)
		if pid == 0 || tuiPIDAlive(pid) {
			t.Errorf("the App Server child (pid %d) is not reaped", pid)
		}
		for _, p := range tuiLines(log, "appserver-token-file ") {
			if _, err := os.Stat(filepath.Dir(p)); err == nil {
				t.Errorf("the token directory %s still exists", filepath.Dir(p))
			}
		}
	}

	t.Run("exit_zero", func(t *testing.T) {
		f, err := runExit(t, "tui-exit-0", "bye")
		if err != nil {
			t.Errorf("TUI exit 0 must end the session with nil, got %v", err)
		}
		reaped(t, f)
	})
	t.Run("exit_seven", func(t *testing.T) {
		f, err := runExit(t, "tui-exit-7", "exit7")
		if code, ok := tuiExitCode(err); !ok || code != 7 {
			t.Errorf("TUI exit 7 must surface as ExitCode 7, got %v (code %d, coder=%v)", err, code, ok)
		}
		reaped(t, f)
	})
	t.Run("signaled", func(t *testing.T) {
		f, err := runExit(t, "tui-signaled", "selfkill")
		if code, ok := tuiExitCode(err); !ok || code != 1 {
			t.Errorf("a TUI ended by a signal must surface as ExitCode 1, got %v (code %d, coder=%v)", err, code, ok)
		}
		reaped(t, f)
	})
	t.Run("session_error_wins", func(t *testing.T) {
		f := newTUIFake(t, "tui-session-error", tuiFakeStatusesEnv+"=completed,failed,failed,failed")
		sess := f.newSession()
		claims := &tuiClaims{always: true}
		done := f.drive(sess, claims)
		// The failing batches start only once the TUI is up, so the session ends
		// under a running TUI.
		if !f.waitLog("tui-pid", tuiAttachWait) {
			t.Errorf("the TUI never started, so the precedence rule is unobserved")
		}
		claims.release.Store(true)
		ok, driveErr := tuiAwait(done, tuiRunWait)
		if !ok {
			t.Fatalf("the driver did not end after the consecutive-failure ceiling")
		}
		err := errors.Join(driveErr, sess.Close())
		if !errors.Is(err, errManagedTurnFailed) {
			t.Errorf("the driver's own error must stand, got %v", err)
		}
		if code, isCoder := tuiExitCode(err); isCoder {
			t.Errorf("the interrupt-induced TUI status (%d) must not replace the driver's error: %v", code, err)
		}
		// Positive control: the owner's stop did interrupt the TUI (its exit
		// status 3), so the status that did not surface existed.
		if !strings.Contains(f.logText(), "tui-interrupted") {
			t.Errorf("the owner's stop never interrupted the TUI, so the precedence rule is unobserved")
		}
	})

	// Blocked-call release: the fake App Server stops reading (SIGSTOP), and the
	// TUI exits 7.
	stop := func(t *testing.T, pid int, sig string) {
		t.Helper()
		if err := exec.Command("kill", "-"+sig, strconv.Itoa(pid)).Run(); err != nil {
			t.Fatalf("kill -%s %d: %v", sig, pid, err)
		}
	}
	t.Run("blocked_turn_start_released", func(t *testing.T) {
		tuiSetDuration(t, &managedWriteDeadline, 30*time.Second)
		var entered atomic.Int64
		origBarrier := managedWriteBarrier
		managedWriteBarrier = func(v any) {
			if m, ok := v.(map[string]any); ok {
				if method, _ := m["method"].(string); method == "turn/start" || method == "tui-second-sender" {
					entered.Add(1)
				}
			}
		}
		t.Cleanup(func() { managedWriteBarrier = origBarrier })

		f := newTUIFake(t, "tui-blocked-write")
		sess := f.newSession()
		claims := &tuiClaims{big: strings.Repeat("x", 48<<20)}
		done := f.drive(sess, claims)
		if !f.waitLog("tui-start", tuiAttachWait) {
			t.Errorf("the TUI never started within %s", tuiAttachWait)
		}
		pid := tuiServerPID(f.logText())
		if pid == 0 {
			t.Fatalf("no App Server pid in the fake log")
		}
		stop(t, pid, "STOP")
		t.Cleanup(func() { _ = exec.Command("kill", "-CONT", strconv.Itoa(pid)).Run() })
		claims.release.Store(true)
		if !waitUntil(func() bool { return entered.Load() >= 1 }, tuiWatchdog) {
			t.Errorf("barrier: the launcher's turn/start write never entered write()")
		}
		// A second sender queues behind the blocked write.
		secondDone := make(chan error, 1)
		go func() {
			secondDone <- sess.client.write(map[string]any{"method": "tui-second-sender", "params": map[string]any{}})
		}()
		if !waitUntil(func() bool { return entered.Load() >= 2 }, tuiWatchdog) {
			t.Errorf("barrier: the second sender never entered write()")
		}
		if returned, _ := tuiAwait(done, 300*time.Millisecond); returned {
			t.Errorf("the driver returned although the turn/start write is blocked: the write was not actually blocked")
		}
		f.say("exit7\n")
		ok, driveErr := tuiAwait(done, tuiWatchdog)
		if !ok {
			t.Errorf("the TUI exit did not release the blocked turn/start write within %s", tuiWatchdog)
			stop(t, pid, "CONT")
		} else if code, isCoder := tuiExitCode(driveErr); !isCoder || code != 7 {
			t.Errorf("the driver must return the TUI's result (exit code 7), got %v", driveErr)
		}
		if ok, _ := tuiAwait(secondDone, tuiWatchdog); !ok {
			t.Errorf("the second sender queued on the write lock was not released")
		}
	})
	t.Run("blocked_wait_turn_released", func(t *testing.T) {
		f := newTUIFake(t, "tui-blocked-wait", tuiFakeHoldFromEnv+"=2")
		sess := f.newSession()
		claims := &tuiClaims{}
		claims.release.Store(true)
		done := f.drive(sess, claims)
		if !f.waitLog("turn-held fake-turn-2", tuiAttachWait) {
			t.Errorf("the launcher turn waiting for a completion never started")
		}
		if returned, _ := tuiAwait(done, 300*time.Millisecond); returned {
			t.Errorf("the driver returned although its turn never completed")
		}
		f.say("exit7\n")
		ok, driveErr := tuiAwait(done, tuiWatchdog)
		if !ok {
			t.Errorf("the TUI exit did not release the waiting turn within %s", tuiWatchdog)
		} else if code, isCoder := tuiExitCode(driveErr); !isCoder || code != 7 {
			t.Errorf("the driver must return the TUI's result (exit code 7), got %v", driveErr)
		}
	})
}

// ---------------------------------------------------------------- AC-MT-009

func TestManagedCodexServerDeathStopsTUI(t *testing.T) {
	for _, tc := range []struct {
		name, mode string
		interrupts bool
	}{{"handles_interrupt", "", true}, {"ignores_interrupt", "ignore-interrupt", false}} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tuiSetDuration(t, &managedTUIStopGrace, 400*time.Millisecond)
			f := newTUIFake(t, "tui-death-"+strings.ReplaceAll(tc.name, "_", "-"), tuiFakeTUIModeEnv+"="+tc.mode)
			done := f.runOwner()
			if !f.waitLog("tui-start", tuiAttachWait) {
				t.Errorf("the TUI never started within %s", tuiAttachWait)
			}
			started := time.Now()
			f.control("close-launcher")
			ok, err := tuiAwait(done, tuiAttachWait)
			elapsed := time.Since(started)
			if !ok {
				t.Errorf("the owner did not return after the App Server connection closed")
				f.finish(done)
			} else if err == nil || !strings.Contains(err.Error(), "connection closed") {
				t.Errorf("the owner must return the connection error, got %v", err)
			}
			log := f.logText()
			pid := tuiTUIPID(log)
			if pid == 0 || tuiPIDAlive(pid) {
				t.Errorf("the TUI child (pid %d) outlived the owner's return", pid)
			}
			interrupted := strings.Contains(log, "tui-interrupted")
			if tc.interrupts && !interrupted {
				t.Errorf("a TUI that handles the interrupt was not interrupted")
			}
			if !tc.interrupts && (interrupted || elapsed < managedTUIStopGrace) {
				t.Errorf("a TUI that ignores the interrupt must be killed only after the grace (elapsed %s, grace %s, interrupted=%v)", elapsed, managedTUIStopGrace, interrupted)
			}
		})
	}
}

func TestManagedCodexSessionEndStopsTUI(t *testing.T) {
	for _, tc := range []struct {
		name, mode string
		interrupts bool
	}{{"handles_interrupt", "", true}, {"ignores_interrupt", "ignore-interrupt", false}} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tuiSetDuration(t, &managedTUIStopGrace, 400*time.Millisecond)
			f := newTUIFake(t, "tui-end-"+strings.ReplaceAll(tc.name, "_", "-"),
				tuiFakeTUIModeEnv+"="+tc.mode, tuiFakeStatusesEnv+"=completed,failed,failed,failed")
			sess := f.newSession()
			claims := &tuiClaims{always: true}
			done := f.drive(sess, claims)
			if !f.waitLog("tui-pid", tuiAttachWait) {
				t.Errorf("the TUI never started within %s", tuiAttachWait)
			}
			claims.release.Store(true)
			ok, driveErr := tuiAwait(done, tuiRunWait)
			if !ok {
				t.Fatalf("the driver did not end after the consecutive-failure ceiling")
			}
			if !errors.Is(driveErr, errManagedTurnFailed) {
				t.Errorf("driver error = %v, want the consecutive-failure error", driveErr)
			}
			started := time.Now()
			_ = sess.Close()
			elapsed := time.Since(started)
			log := f.logText()
			pid := tuiTUIPID(log)
			if pid == 0 || tuiPIDAlive(pid) {
				t.Errorf("the TUI child (pid %d) is still alive after Close returned", pid)
			}
			interrupted := strings.Contains(log, "tui-interrupted")
			if tc.interrupts && !interrupted {
				t.Errorf("a TUI that handles the interrupt was not interrupted at session end")
			}
			if !tc.interrupts && (interrupted || elapsed < managedTUIStopGrace) {
				t.Errorf("a TUI that ignores the interrupt must be killed only after the grace (elapsed %s, interrupted=%v)", elapsed, interrupted)
			}
		})
	}
}

// ---------------------------------------------------------------- AC-MT-011

func TestManagedCodexTUILoopbackRoundTrip(t *testing.T) {
	const body = "TUI_LOOPBACK_SECRET_BODY"
	f := newTUIFake(t, "tui-loopback", tuiFakeTUIModeEnv+"=ws")
	f.env = append(f.env, tuiFakeRootEnv+"="+f.root, tuiFakeRunEnv+"="+f.run)
	done := f.runOwner()
	if !f.waitLog("tui-ws-resumed", tuiAttachWait) {
		t.Errorf("the fake TUI never attached with the env token and its own WebSocket connection")
	}
	store, err := factorymsg.Open(f.root, f.run)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOnCleanup(t, "factory message broker", store)
	ctx := context.Background()
	lane, err := store.Peer(ctx, fakeAppServerThreadID)
	if err != nil {
		t.Fatalf("the managed lane endpoint is not bound to the thread id: %v", err)
	}
	leader := lane
	leader.Role, leader.Slot = factory.RoleLeader, factory.RoleLeader
	leader.SessionUUID = "tui-loopback-leader"
	leader.Generation = 1
	leader, err = registerTUILoopbackLeader(ctx, t, store, leader)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Send(ctx, factorymsg.SendRequest{
		From: leader, To: lane, Kind: factorymsg.KindStatusRequest,
		IdempotencyKey: "tui-once", TaskRef: "t1", CorrelationID: "c1", TTL: time.Minute, Payload: []byte(body),
	}); err != nil {
		t.Fatal(err)
	}
	settled := waitUntil(func() bool {
		st, err := store.Status(ctx)
		return err == nil && st.Acknowledged >= 1
	}, tuiRunWait)
	if !settled {
		t.Errorf("the injected message was never acknowledged through the attached session")
	}
	f.say("bye\n")
	ok, err := tuiAwait(done, tuiRunWait)
	if !ok {
		t.Errorf("the session did not end after the fake TUI exited")
	} else if err != nil {
		t.Errorf("the loopback session must end with nil, got %v", err)
	}
	log := f.logText()
	st, _ := store.Status(ctx)
	if st.Acknowledged != 1 || st.Pending != 0 {
		t.Errorf("acknowledged=%d pending=%d, want 1/0", st.Acknowledged, st.Pending)
	}
	if !strings.Contains(log, "body-read "+body) {
		t.Errorf("the model did not read the body by claim token:\n%s", log)
	}
	for _, l := range tuiLines(log, "turn-prompt ") {
		if strings.Contains(l, body) {
			t.Errorf("an injected prompt carries the message body")
		}
	}
	if pid := tuiServerPID(log); pid == 0 || tuiPIDAlive(pid) {
		t.Errorf("teardown incomplete: App Server pid %d alive", pid)
	}
}

// registerTUILoopbackLeader registers the leader sender the way seedManagedInbox
// does, with the test parent's process identity.
func registerTUILoopbackLeader(ctx context.Context, t *testing.T, store *factorymsg.Store, leader factorymsg.Peer) (factorymsg.Peer, error) {
	t.Helper()
	start, state := homestate.ProbeProcessIdentity(os.Getppid())
	if state != homestate.ProcessIdentityLive || start == "" {
		t.Fatal("test parent process identity unavailable")
	}
	leader.PID, leader.ProcessStart = os.Getppid(), start
	return store.RegisterPeer(ctx, leader)
}

// ---------------------------------------------------------------- AC-MT-016

func TestManagedTUINeverReachedWithoutOptIn(t *testing.T) {
	// ownerRoute restores the production owner behind the divert seam, with the
	// probe and TUI spawn seams counting.
	ownerRoute := func(t *testing.T) (probes, spawns *atomic.Int64, direct *int) {
		withManagedCodexWiring(t)
		direct = captureCodexDirectDoor(t)
		managedFactoryCodexLaunchFunc = func(bin string, args, env []string, dir string) error {
			return runManagedFactoryCodex(bin, args, env, dir, os.Stdin)
		}
		probes, spawns = &atomic.Int64{}, &atomic.Int64{}
		origProbe, origCmd := managedCodexRemoteProbe, managedCodexTUICommand
		managedCodexRemoteProbe = func(string, []string, string) (bool, string) { probes.Add(1); return false, "counting probe" }
		managedCodexTUICommand = func(name string, arg ...string) *exec.Cmd { spawns.Add(1); return origCmd(name, arg...) }
		orig := managedTerminalCheck
		managedTerminalCheck = func(*os.File) bool { return true }
		t.Cleanup(func() {
			managedCodexRemoteProbe, managedCodexTUICommand, managedTerminalCheck = origProbe, origCmd, orig
		})
		t.Setenv("MOAI_HOME", t.TempDir())
		return probes, spawns, direct
	}
	launch := func(t *testing.T) error {
		return runCodex(&cobra.Command{Use: "codex"}, []string{"cli"})
	}
	// ownerDirect reaches the owner entry itself with a launch environment that
	// is outside the gate: the owner enforces the same gate as the divert, so a
	// direct caller never triggers the probe either. The sentinel binary does not
	// exist, so a run that gets past the decision ends at its Start.
	ownerDirect := func(env ...string) {
		_ = runManagedFactoryCodex(sentinelCodexBinaryPath, []string{sentinelCodexBinaryPath}, env, "", os.Stdin)
	}
	const run = config.EnvFactoryRunID + "=run-wire0001"
	const lane = config.EnvMoaiFactoryWorker + "=lane-2"

	t.Run("no_switch", func(t *testing.T) {
		probes, spawns, direct := ownerRoute(t)
		if err := launch(t); err != nil {
			t.Fatal(err)
		}
		ownerDirect(run, lane)
		if probes.Load() != 0 || spawns.Load() != 0 || *direct != 1 {
			t.Errorf("probe=%d tui spawns=%d direct door=%d, want 0/0/1", probes.Load(), spawns.Load(), *direct)
		}
	})
	t.Run("switch_only", func(t *testing.T) {
		probes, spawns, direct := ownerRoute(t)
		managedOptIn(t)
		if err := launch(t); err != nil {
			t.Fatal(err)
		}
		ownerDirect(config.EnvMoaiFactoryManaged+"=1", run)
		if probes.Load() != 0 || spawns.Load() != 0 || *direct != 1 {
			t.Errorf("probe=%d tui spawns=%d direct door=%d, want 0/0/1", probes.Load(), spawns.Load(), *direct)
		}
	})
	t.Run("stamps_only", func(t *testing.T) {
		probes, spawns, direct := ownerRoute(t)
		factoryLaneEnv(t)
		if err := launch(t); err != nil {
			t.Fatal(err)
		}
		ownerDirect(run, lane)
		if probes.Load() != 0 || spawns.Load() != 0 || *direct != 1 {
			t.Errorf("probe=%d tui spawns=%d direct door=%d, want 0/0/1", probes.Load(), spawns.Load(), *direct)
		}
	})
	t.Run("opted_in_probes_once", func(t *testing.T) {
		probes, spawns, direct := ownerRoute(t)
		factoryLaneEnv(t)
		managedOptIn(t)
		// The sentinel codex binary does not exist, so the owner ends at its
		// Start; the probe decision comes first.
		_ = launch(t)
		if probes.Load() != 1 || spawns.Load() != 0 || *direct != 0 {
			t.Errorf("probe=%d tui spawns=%d direct door=%d, want 1/0/0", probes.Load(), spawns.Load(), *direct)
		}
	})
}
