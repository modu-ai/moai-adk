package cli

// managed_codex_factory_test.go — SPEC-FACTORY-MANAGED-SESSION-001 M2
// acceptance coverage: AC-MS-001 (App Server handshake over a loopback WS
// with capability-token auth), AC-MS-002 (thread-id peer bind visible in the
// broker), AC-MS-016 (env-gated live round trip, skip in CI), plus the
// loopback-only dial guard and the managed Codex flag parser. The re-exec
// fake app server follows the codex_launcher_exec_posix_test.go precedent.

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// fakeAppServerRoleEnv / fakeAppServerLogEnv are this file's own re-exec
// trigger and evidence channel (the codex_launcher_exec_posix_test.go
// precedent keeps its trigger constant in the test file too): the fake App
// Server is the test binary re-invoked under -test.run, and it appends one
// token per line so the tests can assert the exact protocol surface they
// drove.
const (
	fakeAppServerRoleEnv = "T1375_FAKE_APP_SERVER_ROLE"
	fakeAppServerLogEnv  = "T1375_FAKE_APP_SERVER_LOG"
	// fakeAppServerFailRPCEnv / fakeAppServerTurnStatusEnv are the failure
	// modes the failure-path tests inject: the named method is answered with
	// a JSON-RPC error object, and turns complete under the given status
	// instead of "completed". fakeAppServerNoTurnIDEnv makes turn/start
	// return a result with no turn id; fakeAppServerNoCompletionEnv stops
	// after turn/started, so the turn never completes.
	fakeAppServerFailRPCEnv      = "T1375_FAKE_APP_SERVER_FAIL_RPC"
	fakeAppServerTurnStatusEnv   = "T1375_FAKE_APP_SERVER_TURN_STATUS"
	fakeAppServerNoTurnIDEnv     = "T1375_FAKE_APP_SERVER_NO_TURN_ID"
	fakeAppServerNoCompletionEnv = "T1375_FAKE_APP_SERVER_NO_COMPLETION"
	// fakeAppServerThreadID is the fixed thread id the fake issues, so the
	// bind test can assert the broker endpoint carries the thread id as its
	// session UUID.
	fakeAppServerThreadID = "fake-thread-1"
)

func fakeAppendLog(logPath, line string) {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.WriteString(line + "\n")
}

// serveFakeRPC answers the one App Server protocol surface the managed owner
// speaks: initialize, initialized, thread/start, thread/name/set, turn/start
// (completed immediately by the turn/started + turn/completed pair). Anything
// else is a test failure by absence — the log names what ran.
func serveFakeRPC(conn *websocket.Conn, logPath string) {
	turns := atomic.Int64{}
	failRPC := os.Getenv(fakeAppServerFailRPCEnv)
	turnStatus := os.Getenv(fakeAppServerTurnStatusEnv)
	if turnStatus == "" {
		turnStatus = "completed"
	}
	for {
		var req struct {
			ID     int             `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := conn.ReadJSON(&req); err != nil {
			return
		}
		fakeAppendLog(logPath, "method "+req.Method)
		switch req.Method {
		case "initialize":
			if req.Method == failRPC {
				_ = conn.WriteJSON(map[string]any{"id": req.ID, "error": map[string]any{"code": -32000, "message": "fake app server: injected failure"}})
				continue
			}
			_ = conn.WriteJSON(map[string]any{"id": req.ID, "result": map[string]string{"status": "ok"}})
		case "initialized":
			// notification: no reply
		case "thread/start":
			var params struct {
				Model string `json:"model"`
				Cwd   string `json:"cwd"`
			}
			_ = json.Unmarshal(req.Params, &params)
			fakeAppendLog(logPath, "thread-cwd "+params.Cwd)
			if params.Model != "" {
				fakeAppendLog(logPath, "model "+params.Model)
			}
			_ = conn.WriteJSON(map[string]any{"id": req.ID, "result": map[string]any{"thread": map[string]string{"id": fakeAppServerThreadID}}})
		case "thread/name/set":
			_ = conn.WriteJSON(map[string]any{"id": req.ID, "result": map[string]any{}})
		case "turn/start":
			if os.Getenv(fakeAppServerNoTurnIDEnv) != "" {
				_ = conn.WriteJSON(map[string]any{"id": req.ID, "result": map[string]any{"turn": map[string]string{}}})
				continue
			}
			turnID := "fake-turn-" + strconv.FormatInt(turns.Add(1), 10)
			_ = conn.WriteJSON(map[string]any{"id": req.ID, "result": map[string]any{"turn": map[string]string{"id": turnID}}})
			_ = conn.WriteJSON(map[string]any{"method": "turn/started", "params": map[string]any{"turn": map[string]string{"id": turnID}}})
			if os.Getenv(fakeAppServerNoCompletionEnv) != "" {
				continue
			}
			_ = conn.WriteJSON(map[string]any{"method": "turn/completed", "params": map[string]any{"turn": map[string]string{"id": turnID, "status": turnStatus}}})
		default:
			_ = conn.WriteJSON(map[string]any{"id": req.ID, "error": map[string]any{"code": -32601, "message": "fake app server: unknown method"}})
		}
	}
}

// TestManagedCodexFakeAppServer is the re-exec helper: the owner launches the
// script fakeAppServerScript wrote, which re-enters this binary with
// -test.run pinned to this test. Triggered only by the role env the owner's
// composed child env carries.
func TestManagedCodexFakeAppServer(t *testing.T) {
	if os.Getenv(fakeAppServerRoleEnv) != "appserver" {
		t.Skip("re-exec helper for the managed Codex handshake tests")
	}
	logPath := os.Getenv(fakeAppServerLogEnv)
	var listen, tokenFile string
	for i := 0; i < len(flag.Args()); i++ {
		switch flag.Arg(i) {
		case "--listen":
			i++
			if i < len(flag.Args()) {
				listen = flag.Arg(i)
			}
		case "--ws-token-file":
			i++
			if i < len(flag.Args()) {
				tokenFile = flag.Arg(i)
			}
		}
	}
	if listen == "" || tokenFile == "" {
		t.Fatalf("fake app server missing --listen/--ws-token-file: %v", flag.Args())
	}
	token, err := os.ReadFile(tokenFile)
	if err != nil {
		t.Fatalf("read capability token: %v", err)
	}
	fakeAppendLog(logPath, "token-loaded")
	if cwd, err := os.Getwd(); err == nil {
		fakeAppendLog(logPath, "server-cwd "+cwd)
	}

	addr := strings.TrimPrefix(listen, "ws://")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("listen on %s: %v", addr, err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// The capability token is the whole auth surface: a bearer mismatch is
		// refused before the upgrade, so a completed handshake below is itself
		// the evidence the owner dialed with the token file's content.
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
		serveFakeRPC(conn, logPath)
	})
	if err := (&http.Server{Handler: mux}).Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fakeAppendLog(logPath, "serve-error "+err.Error())
	}
}

// fakeAppServerScript writes the POSIX shim that turns this test binary into
// the fake App Server program (POSIX-only; the Windows lane proves the owner
// by crossbuild, per the writeManagedFakeBackend precedent).
func fakeAppServerScript(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX sh fixture backend")
	}
	script := "#!/bin/sh\nexec \"" + os.Args[0] + "\" -test.run='^TestManagedCodexFakeAppServer$' -- \"$@\"\n"
	path := filepath.Join(t.TempDir(), "fake-app-server.sh")
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestManagedCodexAppServerHandshake is AC-MS-001: over a temp project root
// and the fake App Server binary, the managed Codex session connects to the
// loopback WS with token auth, waits out /readyz, and completes initialize
// plus thread start.
func TestManagedCodexAppServerHandshake(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "handshake.log")
	backend := fakeAppServerScript(t)
	env := []string{
		fakeAppServerRoleEnv + "=appserver",
		fakeAppServerLogEnv + "=" + logPath,
	}
	s, err := newManagedCodexSession(backend, []string{backend, "-m", "fake-model"}, env)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatalf("managed Codex handshake: %v", err)
	}
	if s.threadID != fakeAppServerThreadID {
		t.Fatalf("thread id %q, want %q", s.threadID, fakeAppServerThreadID)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close after handshake: %v", err)
	}
	// A second Close must not panic or hang (idempotent teardown).
	if err := s.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("fake app server log: %v", err)
	}
	for _, want := range []string{"token-loaded", "auth-ok", "method initialize", "method initialized", "method thread/start", "method thread/name/set", "model fake-model"} {
		if !strings.Contains(string(log), want) {
			t.Errorf("handshake log misses %q:\n%s", want, log)
		}
	}
	if strings.Contains(string(log), "auth-refused") {
		t.Errorf("handshake hit an auth refusal:\n%s", log)
	}
}

// TestManagedCodexRegistersBoundPeer is AC-MS-002: after the handshake starts
// a thread, the broker shows the lane endpoint bound with the thread id as
// its session UUID — the launcher completed the bind half of REQ-MS-002
// itself (App Server turns run no TUI SessionStart hook).
func TestManagedCodexRegistersBoundPeer(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root, run := t.TempDir(), "managed-codex-bind"
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	activateManagedRun(t, root, run)
	logPath := filepath.Join(t.TempDir(), "bind.log")
	backend := fakeAppServerScript(t)
	env := []string{
		config.EnvMoaiKanbanID + "=" + run,
		config.EnvMoaiKanbanBackend + "=" + BackendCodex,
		config.EnvMoaiFactoryWorker + "=" + kanban.FactoryLaneLabel(1),
		fakeAppServerRoleEnv + "=appserver",
		fakeAppServerLogEnv + "=" + logPath,
	}
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stdinW.WriteString("/exit\n"); err != nil {
		t.Fatal(err)
	}
	if err := stdinW.Close(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stdinR.Close() })

	errCh := make(chan error, 1)
	go func() { errCh <- runManagedFactoryCodex(backend, []string{backend}, env, "", stdinR) }()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("managed codex run: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("managed codex run did not exit within 30s")
	}

	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("fake app server log: %v", err)
	}
	for _, want := range []string{"auth-ok", "method thread/start", "method turn/start"} {
		if !strings.Contains(string(log), want) {
			t.Errorf("bind-run log misses %q:\n%s", want, log)
		}
	}

	store, err := factorymsg.Open(root, run)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOnCleanup(t, "factory message broker", store)
	roster, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(roster.Lanes) != 1 {
		t.Fatalf("roster lanes=%+v, want exactly the managed lane", roster.Lanes)
	}
	lane := roster.Lanes[0]
	if lane.Slot != kanban.FactoryLaneLabel(1) || lane.BindingState != factorymsg.BindingBound || lane.SessionUUID != fakeAppServerThreadID {
		t.Fatalf("bound endpoint = slot %q state %q session %q; want lane-1/bound/%s",
			lane.Slot, lane.BindingState, lane.SessionUUID, fakeAppServerThreadID)
	}
}

// TestManagedCodexOwnerUsesLaunchDir pins the launch directory the divert
// carries (F1 of the t1375 sync audit): the App Server process and the thread
// cwd both use the directory the launcher resolved — project root or a -w
// worktree — never the process cwd, which here is a different directory.
func TestManagedCodexOwnerUsesLaunchDir(t *testing.T) {
	t.Setenv("MOAI_HOME", t.TempDir())
	root, run := t.TempDir(), "managed-codex-dir"
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	activateManagedRun(t, root, run)
	launchDir := t.TempDir()
	procCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if resolved, rerr := filepath.EvalSymlinks(launchDir); rerr != nil || resolved == procCwd {
		t.Fatalf("test needs a launch dir distinct from the process cwd: %q vs %q (%v)", launchDir, procCwd, rerr)
	}
	logPath := filepath.Join(t.TempDir(), "dir.log")
	backend := fakeAppServerScript(t)
	env := []string{
		config.EnvMoaiKanbanID + "=" + run,
		config.EnvMoaiKanbanBackend + "=" + BackendCodex,
		config.EnvMoaiFactoryWorker + "=" + kanban.FactoryLaneLabel(1),
		fakeAppServerRoleEnv + "=appserver",
		fakeAppServerLogEnv + "=" + logPath,
	}
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stdinW.WriteString("/exit\n"); err != nil {
		t.Fatal(err)
	}
	if err := stdinW.Close(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stdinR.Close() })

	errCh := make(chan error, 1)
	go func() { errCh <- runManagedFactoryCodex(backend, []string{backend}, env, launchDir, stdinR) }()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("managed codex run: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("managed codex run did not exit within 30s")
	}
	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	wantServer, _ := filepath.EvalSymlinks(launchDir)
	for _, want := range []string{"server-cwd " + wantServer, "thread-cwd " + launchDir} {
		if !strings.Contains(string(log), want+"\n") {
			t.Errorf("log misses %q (process cwd %q):\n%s", want, procCwd, log)
		}
	}
}

// TestManagedCodexTokenFileIsPrivate pins the capability-token secrecy the
// design promises (acceptance §3 Secured): while the session is up, the token
// file is 0600 and its directory grants nothing to group or other.
func TestManagedCodexTokenFileIsPrivate(t *testing.T) {
	backend := fakeAppServerScript(t)
	env := []string{
		fakeAppServerRoleEnv + "=appserver",
		fakeAppServerLogEnv + "=" + filepath.Join(t.TempDir(), "token.log"),
	}
	s, err := newManagedCodexSession(backend, []string{backend}, env)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() { _ = s.Close() }()
	file, err := os.Stat(filepath.Join(s.tokenDir, "token"))
	if err != nil {
		t.Fatal(err)
	}
	if got := file.Mode().Perm(); got != 0o600 {
		t.Errorf("token file mode = %o, want 600", got)
	}
	dir, err := os.Stat(s.tokenDir)
	if err != nil {
		t.Fatal(err)
	}
	if got := dir.Mode().Perm(); got&0o077 != 0 {
		t.Errorf("token directory mode = %o, grants access beyond the owner", got)
	}
}

// TestManagedCodexRefusesNonLoopbackDial proves the transport is structurally
// loopback: the dial guard refuses any non-loopback target (IP literals only
// — "localhost" is not an IP literal, so no DNS ever sits in the dial path)
// before any network activity.
func TestManagedCodexRefusesNonLoopbackDial(t *testing.T) {
	for _, bad := range []string{
		"ws://10.1.2.3:9000", "ws://172.16.0.9:8080", "ws://[2001:db8::1]:9000",
		"wss://127.0.0.1:9000", "http://127.0.0.1:9000", "ws://localhost:9000",
	} {
		if err := managedLoopbackURL(bad); !errors.Is(err, errManagedCodexNonLoopback) {
			t.Errorf("dial target %q = %v, want the loopback refusal", bad, err)
		}
	}
	for _, ok := range []string{"ws://127.0.0.1:9000", "ws://127.255.0.2:1", "ws://[::1]:9000"} {
		if err := managedLoopbackURL(ok); err != nil {
			t.Errorf("loopback target %q refused: %v", ok, err)
		}
	}
	// The guard sits inside the dial path: a non-loopback dial fails with the
	// sentinel — never a connection error, which would mean the guard never
	// fired.
	if _, _, err := managedCodexDial("ws://10.1.2.3:9000", nil); !errors.Is(err, errManagedCodexNonLoopback) {
		t.Fatalf("dial to a non-loopback endpoint = %v, want the loopback refusal", err)
	}
}

func TestManagedCodexOptions(t *testing.T) {
	appArgs, model, err := managedCodexOptions([]string{"codex", "-c", "model_reasoning=high", "--config=x.y=z", "-m", "gpt-5.3"})
	if err != nil {
		t.Fatal(err)
	}
	if model != "gpt-5.3" {
		t.Fatalf("model=%q, want gpt-5.3", model)
	}
	if len(appArgs) != 4 || appArgs[0] != "-c" || appArgs[1] != "model_reasoning=high" || appArgs[2] != "-c" || appArgs[3] != "x.y=z" {
		t.Fatalf("appArgs=%v, want both config values passed through", appArgs)
	}
	if _, model, err := managedCodexOptions([]string{"codex", "--model=fake-model"}); err != nil || model != "fake-model" {
		t.Fatalf("--model= form: model=%q err=%v", model, err)
	}
	if _, _, err := managedCodexOptions([]string{"codex", "--profile", "p"}); err == nil || !strings.Contains(err.Error(), "--profile") {
		t.Fatalf("unsupported flag refusal = %v, want an error naming --profile", err)
	}
	for _, bad := range [][]string{
		{"codex", "-c"},
		{"codex", "-m"},
		{"codex", "-m", " "},
		{"codex", "--model="},
	} {
		if _, _, err := managedCodexOptions(bad); err == nil {
			t.Errorf("args %v accepted without a value", bad)
		}
	}
}

// TestManagedCodexFailurePaths covers the owner's refusal and failure
// branches: a missing binary fails Start with the backend named, an RPC
// error object fails the handshake, a turn ending in a non-completed state
// fails the delivery, and a missing factory run id refuses before any child
// work.
func TestManagedCodexFailurePaths(t *testing.T) {
	backend := fakeAppServerScript(t)

	t.Run("missing binary fails Start", func(t *testing.T) {
		s, err := newManagedCodexSession("missing-codex-bin", []string{"missing-codex-bin"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		err = s.Start()
		if err == nil || !strings.Contains(err.Error(), "codex") {
			t.Fatalf("start error %v does not name the backend", err)
		}
		if err := s.Close(); err != nil {
			t.Fatalf("close of an unstarted session: %v", err)
		}
	})

	t.Run("rpc error object fails the handshake", func(t *testing.T) {
		env := []string{
			fakeAppServerRoleEnv + "=appserver",
			fakeAppServerLogEnv + "=" + filepath.Join(t.TempDir(), "failrpc.log"),
			fakeAppServerFailRPCEnv + "=initialize",
		}
		s, err := newManagedCodexSession(backend, []string{backend}, env)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = s.Close() }()
		err = s.Start()
		if err == nil || !strings.Contains(err.Error(), "injected failure") {
			t.Fatalf("handshake over an error object = %v, want the injected failure", err)
		}
	})

	t.Run("failed turn status fails the turn", func(t *testing.T) {
		env := []string{
			fakeAppServerRoleEnv + "=appserver",
			fakeAppServerLogEnv + "=" + filepath.Join(t.TempDir(), "failturn.log"),
			fakeAppServerTurnStatusEnv + "=failed",
		}
		s, err := newManagedCodexSession(backend, []string{backend}, env)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Start(); err != nil {
			t.Fatalf("handshake: %v", err)
		}
		defer func() { _ = s.Close() }()
		err = s.DeliverTurn("doomed-turn")
		if err == nil || !strings.Contains(err.Error(), "ended as failed") {
			t.Fatalf("failed-status turn = %v, want a turn failure", err)
		}
	})

	t.Run("missing run id refuses before any child work", func(t *testing.T) {
		err := runManagedFactoryCodex(backend, []string{backend}, nil, "", strings.NewReader(""))
		if err == nil || !strings.Contains(err.Error(), "factory run id") {
			t.Fatalf("run without a factory run id = %v, want the run-id refusal", err)
		}
	})

	t.Run("unsupported flag refuses before any broker or child work", func(t *testing.T) {
		env := []string{config.EnvMoaiKanbanID + "=refused-run"}
		err := runManagedFactoryCodex(backend, []string{backend, "--profile", "p"}, env, "", strings.NewReader(""))
		if err == nil || !strings.Contains(err.Error(), "--profile") {
			t.Fatalf("run with an unsupported flag = %v, want a refusal naming --profile", err)
		}
	})

	t.Run("turn without an id fails the turn", func(t *testing.T) {
		env := []string{
			fakeAppServerRoleEnv + "=appserver",
			fakeAppServerLogEnv + "=" + filepath.Join(t.TempDir(), "noturnid.log"),
			fakeAppServerNoTurnIDEnv + "=1",
		}
		s, err := newManagedCodexSession(backend, []string{backend}, env)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Start(); err != nil {
			t.Fatalf("handshake: %v", err)
		}
		defer func() { _ = s.Close() }()
		err = s.DeliverTurn("idless-turn")
		if err == nil || !strings.Contains(err.Error(), "no turn id") {
			t.Fatalf("idless turn/start = %v, want a turn failure", err)
		}
	})

	t.Run("unresponsive turn times out instead of holding the queue", func(t *testing.T) {
		env := []string{
			fakeAppServerRoleEnv + "=appserver",
			fakeAppServerLogEnv + "=" + filepath.Join(t.TempDir(), "nocompletion.log"),
			fakeAppServerNoCompletionEnv + "=1",
		}
		s, err := newManagedCodexSession(backend, []string{backend}, env)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Start(); err != nil {
			t.Fatalf("handshake: %v", err)
		}
		defer func() { _ = s.Close() }()
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		err = s.client.startTurn(ctx, s.threadID, "stalled-turn")
		if err == nil || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("stalled turn = %v, want a deadline expiration", err)
		}
	})

	t.Run("client calls survive a raced shutdown", func(t *testing.T) {
		env := []string{
			fakeAppServerRoleEnv + "=appserver",
			fakeAppServerLogEnv + "=" + filepath.Join(t.TempDir(), "shutdown.log"),
		}
		s, err := newManagedCodexSession(backend, []string{backend}, env)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Start(); err != nil {
			t.Fatalf("handshake: %v", err)
		}
		if err := s.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
		// A second shutdown and a call after close must fail cleanly, never
		// panic — teardown paths may converge here more than once.
		s.client.shutdown()
		if _, err := s.client.call(context.Background(), "initialize", map[string]any{}); err == nil {
			t.Fatal("call after close was accepted")
		}
	})
}

// TestManagedCodexFactoryBrokerLive is AC-MS-016: the live round trip runs
// only under the live gate env (a real project root carrying a live factory
// run) and with a real codex binary on PATH; without the env the test skips,
// which is the CI posture.
func TestManagedCodexFactoryBrokerLive(t *testing.T) {
	liveRoot, liveRun := os.Getenv("MOAI_FACTORY_LIVE_ROOT"), os.Getenv("MOAI_FACTORY_LIVE_RUN")
	if liveRoot == "" || liveRun == "" {
		t.Skip("live gate env (MOAI_FACTORY_LIVE_ROOT / MOAI_FACTORY_LIVE_RUN) not set — CI runs this test as a skip")
	}
	if runtime.GOOS == "windows" {
		t.Skip("live Codex session verified on POSIX only (design.md D-4 honest gap)")
	}
	codexBin, err := exec.LookPath("codex")
	if err != nil {
		t.Skipf("no codex binary on PATH: %v", err)
	}
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv("CLAUDE_PROJECT_DIR", liveRoot)
	env := []string{
		config.EnvMoaiKanbanID + "=" + liveRun,
		config.EnvMoaiKanbanBackend + "=" + BackendCodex,
		config.EnvMoaiFactoryWorker + "=" + kanban.FactoryLaneLabel(1),
	}
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stdinW.WriteString("/exit\n"); err != nil {
		t.Fatal(err)
	}
	if err := stdinW.Close(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stdinR.Close() })

	errCh := make(chan error, 1)
	go func() { errCh <- runManagedFactoryCodex(codexBin, []string{codexBin}, env, "", stdinR) }()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("live managed codex run: %v", err)
		}
	case <-time.After(10 * time.Minute):
		t.Fatal("live managed codex run did not exit within 10m")
	}
	store, err := factorymsg.Open(liveRoot, liveRun)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOnCleanup(t, "factory message broker", store)
	roster, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	bound := 0
	for _, lane := range roster.Lanes {
		if lane.BindingState == factorymsg.BindingBound {
			bound++
		}
	}
	if bound == 0 {
		t.Fatalf("live roster has no bound endpoint: %+v", roster.Lanes)
	}
}
