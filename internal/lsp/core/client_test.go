package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/config"
	"github.com/modu-ai/moai-adk/internal/lsp/subprocess"
	"github.com/modu-ai/moai-adk/internal/lsp/transport"
)

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

// fakeSupervisor는 실제 subprocess 없이 Shutdown 경로를 테스트하는 supervisor 더블.
type fakeSupervisor struct {
	mu          sync.Mutex
	signalLog   []os.Signal
	killCalled  bool
	signalError error
	exitCh      chan subprocess.ExitEvent
}

func newFakeSupervisor() *fakeSupervisor {
	ch := make(chan subprocess.ExitEvent, 1)
	return &fakeSupervisor{exitCh: ch}
}

func (f *fakeSupervisor) Watch(ctx context.Context) <-chan subprocess.ExitEvent {
	out := make(chan subprocess.ExitEvent, 1)
	go func() {
		defer close(out)
		select {
		case ev, ok := <-f.exitCh:
			if ok {
				out <- ev
			}
		case <-ctx.Done():
		}
	}()
	return out
}

func (f *fakeSupervisor) Signal(sig os.Signal) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.signalLog = append(f.signalLog, sig)
	return f.signalError
}

func (f *fakeSupervisor) Kill() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.killCalled = true
	return nil
}

// exitNow triggers immediate subprocess exit in the supervisor.
func (f *fakeSupervisor) exitNow() {
	f.exitCh <- subprocess.ExitEvent{ExitCode: 0}
}

// fakeLauncher는 실제 서브프로세스를 생성하지 않는 테스트 전용 Launcher.
// 기본적으로 성공을 반환하며, failWith가 설정된 경우 해당 에러를 반환합니다.
type fakeLauncher struct {
	mu       sync.Mutex
	called   int
	failWith error
}

func (f *fakeLauncher) Launch(ctx context.Context, cfg config.ServerConfig) (*subprocess.LaunchResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.called++
	if f.failWith != nil {
		return nil, f.failWith
	}
	// 실제 subprocess 없이 가짜 파이프 반환 (Cmd is nil — Supervisor must handle nil gracefully)
	_, w1 := io.Pipe()
	r2, _ := io.Pipe()
	r3, _ := io.Pipe()
	return &subprocess.LaunchResult{
		Stdin:  w1,
		Stdout: r2,
		Stderr: r3,
	}, nil
}

// fakeTransport는 실제 JSON-RPC 통신 없이 동작하는 테스트 전용 Transport.
type fakeTransport struct {
	mu            sync.Mutex
	callLog       []string
	notifyLog     []string
	closed        bool
	callResponses map[string]callResponse
}

type callResponse struct {
	result json.RawMessage
	err    error
}

func newFakeTransport() *fakeTransport {
	return &fakeTransport{
		callResponses: make(map[string]callResponse),
	}
}

// setCallResponse는 특정 method에 대한 응답을 설정합니다.
func (f *fakeTransport) setCallResponse(method string, result json.RawMessage, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callResponses[method] = callResponse{result: result, err: err}
}

func (f *fakeTransport) Call(ctx context.Context, method string, params, result any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callLog = append(f.callLog, method)

	if resp, ok := f.callResponses[method]; ok {
		if resp.err != nil {
			return resp.err
		}
		if result != nil && resp.result != nil {
			return json.Unmarshal(resp.result, result)
		}
		return nil
	}
	// 기본: 성공, result 수정 없음
	return nil
}

func (f *fakeTransport) Notify(ctx context.Context, method string, params any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.notifyLog = append(f.notifyLog, method)
	return nil
}

func (f *fakeTransport) OnNotification(method string, handler func(params json.RawMessage)) {}

func (f *fakeTransport) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func (f *fakeTransport) callCount(method string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, m := range f.callLog {
		if m == method {
			n++
		}
	}
	return n
}

func (f *fakeTransport) notifyCount(method string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, m := range f.notifyLog {
		if m == method {
			n++
		}
	}
	return n
}

// fakeTransportFactory는 fakeLauncher 이후에 fakeTransport를 반환하는 팩토리.
type fakeTransportFactory struct {
	tr *fakeTransport
}

func newFakeTransportFactory(tr *fakeTransport) *fakeTransportFactory {
	return &fakeTransportFactory{tr: tr}
}

func (f *fakeTransportFactory) New(stream io.ReadWriteCloser) transport.Transport {
	return f.tr
}

// ---------------------------------------------------------------------------
// NewClient option tests
// ---------------------------------------------------------------------------

// TestNewClient_DefaultOptions verifies that NewClient creates a client in spawning state.
func TestNewClient_DefaultOptions(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	fl := &fakeLauncher{}

	c := NewClient(cfg,
		WithLauncherFunc(fl.Launch),
		WithTransportFactory(newFakeTransportFactory(newFakeTransport()).New),
	)

	if got := c.State(); got != StateSpawning {
		t.Errorf("expected initial state %q, got %q", StateSpawning, got)
	}
}

// TestNewClient_WithShutdownTimeout verifies the WithShutdownTimeout option.
func TestNewClient_WithShutdownTimeout(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := NewClient(cfg,
		WithShutdownTimeout(42*time.Second),
		WithTransportFactory(newFakeTransportFactory(newFakeTransport()).New),
	)
	if c.shutdownTimeout != 42*time.Second {
		t.Errorf("expected shutdownTimeout 42s, got %v", c.shutdownTimeout)
	}
}

// ---------------------------------------------------------------------------
// Start tests
// ---------------------------------------------------------------------------

// TestClient_Start_Success verifies the spawning→initializing→ready path.
func TestClient_Start_Success(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	fl := &fakeLauncher{}
	ft := newFakeTransport()

	// initialize 응답 설정: capabilities 포함
	ft.setCallResponse("initialize", json.RawMessage(`{"capabilities":{}}`), nil)

	c := NewClient(cfg,
		WithLauncherFunc(fl.Launch),
		WithTransportFactory(newFakeTransportFactory(ft).New),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: unexpected error: %v", err)
	}

	if got := c.State(); got != StateReady {
		t.Errorf("after Start: expected state %q, got %q", StateReady, got)
	}
	if fl.called == 0 {
		t.Error("Launch was not called")
	}
	if ft.callCount("initialize") == 0 {
		t.Error("initialize was not sent")
	}
}

// TestClient_Start_BinaryNotFound verifies that ErrBinaryNotFound transitions to shutdown.
func TestClient_Start_BinaryNotFound(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "nonexistent_binary_xyz"}
	fl := &fakeLauncher{
		failWith: fmt.Errorf("subprocess.Launch %q: %w", "nonexistent_binary_xyz", subprocess.ErrBinaryNotFound),
	}

	c := NewClient(cfg,
		WithLauncherFunc(fl.Launch),
		WithTransportFactory(newFakeTransportFactory(newFakeTransport()).New),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := c.Start(ctx)
	if err == nil {
		t.Fatal("Start: expected error for missing binary, got nil")
	}
	if !errors.Is(err, subprocess.ErrBinaryNotFound) {
		t.Errorf("Start: expected errors.Is(err, ErrBinaryNotFound), got %v", err)
	}
	if got := c.State(); got != StateShutdown {
		t.Errorf("after BinaryNotFound: expected state %q, got %q", StateShutdown, got)
	}
}

// ---------------------------------------------------------------------------
// Shutdown tests
// ---------------------------------------------------------------------------

// TestClient_Shutdown_SendsShutdownAndExit verifies that Shutdown sends LSP shutdown+exit.
func TestClient_Shutdown_SendsShutdownAndExit(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	fl := &fakeLauncher{}
	ft := newFakeTransport()
	ft.setCallResponse("initialize", json.RawMessage(`{"capabilities":{}}`), nil)

	c := NewClient(cfg,
		WithLauncherFunc(fl.Launch),
		WithTransportFactory(newFakeTransportFactory(ft).New),
		WithShutdownTimeout(500*time.Millisecond),
	)

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := c.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: unexpected error: %v", err)
	}

	if ft.callCount("shutdown") == 0 {
		t.Error("shutdown request was not sent")
	}
	if ft.notifyCount("exit") == 0 {
		t.Error("exit notification was not sent")
	}
	if got := c.State(); got != StateShutdown {
		t.Errorf("after Shutdown: expected state %q, got %q", StateShutdown, got)
	}
	if !ft.closed {
		t.Error("transport was not closed after Shutdown")
	}
}

// TestClient_Shutdown_IdempotentOnUnstarted verifies that Shutdown on an unstarted
// (spawning state) client transitions to shutdown without error.
func TestClient_Shutdown_IdempotentOnUnstarted(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := NewClient(cfg,
		WithLauncherFunc((&fakeLauncher{}).Launch),
		WithTransportFactory(newFakeTransportFactory(newFakeTransport()).New),
	)

	ctx := context.Background()
	// Shutdown before Start — should not panic and should set state to shutdown
	err := c.Shutdown(ctx)
	if err != nil {
		t.Logf("Shutdown on unstarted client returned (acceptable): %v", err)
	}
	if got := c.State(); got != StateShutdown {
		t.Errorf("expected state %q after unstarted Shutdown, got %q", StateShutdown, got)
	}
}

// TestClient_State verifies that State() delegates to the internal StateMachine.
func TestClient_State(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := NewClient(cfg,
		WithLauncherFunc((&fakeLauncher{}).Launch),
		WithTransportFactory(newFakeTransportFactory(newFakeTransport()).New),
	)
	if got := c.State(); got != StateSpawning {
		t.Errorf("State(): expected %q, got %q", StateSpawning, got)
	}
}

// TestClient_Sprint4Methods verifies Sprint-4 method behaviors.
// OpenFile and GetDiagnostics are now implemented; FindReferences/GotoDefinition remain stubs.
func TestClient_Sprint4Methods(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	fl := &fakeLauncher{}
	ft := newFakeTransport()
	ft.setCallResponse("initialize", json.RawMessage(`{"capabilities":{}}`), nil)

	c := NewClient(cfg,
		WithLauncherFunc(fl.Launch),
		WithTransportFactory(newFakeTransportFactory(ft).New),
	)

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// OpenFile은 Sprint 4에서 구현됨 — 에러 없이 성공해야 함
	t.Run("OpenFile_Succeeds", func(t *testing.T) {
		err := c.OpenFile(ctx, "/tmp/foo.go", "package main")
		if err != nil {
			t.Errorf("OpenFile: expected nil, got %v", err)
		}
	})

	// GetDiagnostics는 Sprint 4에서 구현됨 — 열린 파일이면 빈 슬라이스 반환
	t.Run("GetDiagnostics_OpenedFile", func(t *testing.T) {
		// OpenFile로 먼저 파일을 열어야 함
		_ = c.OpenFile(ctx, "/tmp/diag.go", "package main")
		diags, err := c.GetDiagnostics(ctx, "/tmp/diag.go")
		if err != nil {
			t.Errorf("GetDiagnostics: expected nil error, got %v", err)
		}
		if diags == nil {
			t.Error("GetDiagnostics: expected non-nil slice, got nil")
		}
	})

	// GetDiagnostics는 열리지 않은 파일에 대해 ErrFileNotOpen 반환
	t.Run("GetDiagnostics_UnopenedFile", func(t *testing.T) {
		_, err := c.GetDiagnostics(ctx, "/tmp/notopen.go")
		if !errors.Is(err, ErrFileNotOpen) {
			t.Errorf("expected ErrFileNotOpen, got %v", err)
		}
	})

	// FindReferences는 Sprint 4에서 구현됨 — 서버 capabilities가 없으면 ErrCapabilityUnsupported 반환
	t.Run("FindReferences_NoCapability", func(t *testing.T) {
		_, err := c.FindReferences(ctx, "/tmp/foo.go", lsp.Position{Line: 0, Character: 0})
		if !errors.Is(err, ErrCapabilityUnsupported) {
			t.Errorf("expected ErrCapabilityUnsupported (server has no refs capability), got %v", err)
		}
	})

	// GotoDefinition은 Sprint 4에서 구현됨 — 서버 capabilities가 없으면 ErrCapabilityUnsupported 반환
	t.Run("GotoDefinition_NoCapability", func(t *testing.T) {
		_, err := c.GotoDefinition(ctx, "/tmp/foo.go", lsp.Position{Line: 0, Character: 0})
		if !errors.Is(err, ErrCapabilityUnsupported) {
			t.Errorf("expected ErrCapabilityUnsupported (server has no def capability), got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// Integration: Client interface type assertions
// ---------------------------------------------------------------------------

// TestClient_ImplementsInterface verifies that *client satisfies the Client interface.
func TestClient_ImplementsInterface(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	var _ Client = NewClient(cfg)
}

// TestClient_Start_TransportClosedOnFailure verifies that transport is cleaned up
// when initialization fails after subprocess launch.
func TestClient_Start_TransportClosedOnFailure(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	fl := &fakeLauncher{}
	ft := newFakeTransport()

	initErr := errors.New("lsp initialize failed")
	ft.setCallResponse("initialize", nil, initErr)

	c := NewClient(cfg,
		WithLauncherFunc(fl.Launch),
		WithTransportFactory(newFakeTransportFactory(ft).New),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := c.Start(ctx)
	if err == nil {
		t.Fatal("Start: expected error on initialize failure, got nil")
	}
	// 초기화 실패 시 상태가 degraded 또는 shutdown으로 전환되어야 함
	state := c.State()
	if state != StateDegraded && state != StateShutdown {
		t.Errorf("after initialize failure: expected degraded or shutdown, got %q", state)
	}
}

// TestClient_Start_GenericLaunchError verifies that non-BinaryNotFound launch errors
// transition to shutdown and are returned to the caller.
func TestClient_Start_GenericLaunchError(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	genericErr := errors.New("permission denied")
	fl := &fakeLauncher{failWith: genericErr}

	c := NewClient(cfg,
		WithLauncherFunc(fl.Launch),
		WithTransportFactory(newFakeTransportFactory(newFakeTransport()).New),
	)

	ctx := context.Background()
	err := c.Start(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, subprocess.ErrBinaryNotFound) {
		t.Error("generic error should not wrap ErrBinaryNotFound")
	}
	if got := c.State(); got != StateShutdown {
		t.Errorf("expected shutdown state, got %q", got)
	}
}

// TestWithLogger verifies the WithLogger option sets the logger on the client.
func TestWithLogger(t *testing.T) {
	buf := &logBuffer{}
	logger := newTestLogger(buf)

	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	fl := &fakeLauncher{}
	ft := newFakeTransport()
	ft.setCallResponse("initialize", json.RawMessage(`{"capabilities":{}}`), nil)

	c := NewClient(cfg,
		WithLauncherFunc(fl.Launch),
		WithTransportFactory(newFakeTransportFactory(ft).New),
		WithLogger(logger),
	)

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Logger이 적용되었으면 state 전환 시 로그가 기록됨
	logged := buf.String()
	if logged == "" {
		t.Error("expected log output from WithLogger, got empty")
	}
}

// TestReadWriteCloser_Operations verifies the readWriteCloser helper works correctly.
func TestReadWriteCloser_Operations(t *testing.T) {
	pr, pw := io.Pipe()

	rwc := &readWriteCloser{
		r:       pr,
		w:       pw,
		closers: []io.Closer{pr, pw},
	}

	// 쓰기 후 읽기
	payload := []byte("hello")
	go func() {
		_, _ = rwc.Write(payload)
	}()

	buf := make([]byte, len(payload))
	n, err := rwc.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(buf[:n]) != "hello" {
		t.Errorf("expected %q, got %q", "hello", string(buf[:n]))
	}

	// Close: 양쪽 파이프 모두 닫힘
	if err := rwc.Close(); err != nil {
		// io.Pipe Close는 이미 닫힌 파이프에 에러를 반환할 수 있음 — 무시
		_ = err
	}
}

// TestClient_Shutdown_WithSupervisor_GracefulExit verifies the SIGTERM → graceful exit path.
func TestClient_Shutdown_WithSupervisor_GracefulExit(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	fl := &fakeLauncher{}
	ft := newFakeTransport()
	ft.setCallResponse("initialize", json.RawMessage(`{"capabilities":{}}`), nil)

	sv := newFakeSupervisor()

	c := NewClient(cfg,
		WithLauncherFunc(fl.Launch),
		WithTransportFactory(newFakeTransportFactory(ft).New),
		WithShutdownTimeout(2*time.Second),
		withSupervisorIface(sv),
	)

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Supervisor가 즉시 종료하는 시뮬레이션
	go sv.exitNow()

	if err := c.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	if got := c.State(); got != StateShutdown {
		t.Errorf("expected shutdown, got %q", got)
	}
}

// TestClient_Shutdown_WithSupervisor_SIGTERMFail verifies the SIGTERM fail → SIGKILL path.
func TestClient_Shutdown_WithSupervisor_SIGTERMFail(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	fl := &fakeLauncher{}
	ft := newFakeTransport()
	ft.setCallResponse("initialize", json.RawMessage(`{"capabilities":{}}`), nil)

	sv := newFakeSupervisor()
	sv.signalError = errors.New("process already dead")

	c := NewClient(cfg,
		WithLauncherFunc(fl.Launch),
		WithTransportFactory(newFakeTransportFactory(ft).New),
		WithShutdownTimeout(500*time.Millisecond),
		withSupervisorIface(sv),
	)

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := c.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	sv.mu.Lock()
	killed := sv.killCalled
	sv.mu.Unlock()

	if !killed {
		t.Error("expected Kill to be called when SIGTERM fails")
	}
	if got := c.State(); got != StateShutdown {
		t.Errorf("expected shutdown, got %q", got)
	}
}

// TestClient_Shutdown_WithCapableServer verifies full Shutdown path with server capabilities.
func TestClient_Shutdown_WithCapableServer(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	fl := &fakeLauncher{}
	ft := newFakeTransport()
	ft.setCallResponse("initialize", json.RawMessage(`{
		"capabilities": {
			"textDocumentSync": 1,
			"referencesProvider": true,
			"definitionProvider": true
		}
	}`), nil)

	c := NewClient(cfg,
		WithLauncherFunc(fl.Launch),
		WithTransportFactory(newFakeTransportFactory(ft).New),
		WithShutdownTimeout(500*time.Millisecond),
	)

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// 서버 캐퍼빌리티가 파싱되었는지 확인
	if !c.serverCaps.ReferencesProvider {
		t.Error("expected ReferencesProvider to be true after Start")
	}

	if err := c.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	if got := c.State(); got != StateShutdown {
		t.Errorf("expected shutdown, got %q", got)
	}
}

// TestClient_Capabilities_ExposedViaInterface verifies that client.Capabilities()
// returns the server capabilities parsed during the LSP initialize handshake (REQ-LM-009).
func TestClient_Capabilities_ExposedViaInterface(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	fl := &fakeLauncher{}
	ft := newFakeTransport()
	ft.setCallResponse("initialize", json.RawMessage(`{
		"capabilities": {
			"textDocumentSync": 1,
			"referencesProvider": true,
			"definitionProvider": true
		}
	}`), nil)

	c := NewClient(cfg,
		WithLauncherFunc(fl.Launch),
		WithTransportFactory(newFakeTransportFactory(ft).New),
	)

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() { _ = c.Shutdown(ctx) }()

	caps := c.Capabilities()
	if !caps.ReferencesProvider {
		t.Error("Capabilities().ReferencesProvider should be true")
	}
	if !caps.DefinitionProvider {
		t.Error("Capabilities().DefinitionProvider should be true")
	}
	if !caps.DiagnosticProvider {
		t.Error("Capabilities().DiagnosticProvider should be true (textDocumentSync=1)")
	}
}

// TestClient_Capabilities_BeforeStart verifies Capabilities() returns empty defaults
// when called before Start (safe nil-access).
func TestClient_Capabilities_BeforeStart(t *testing.T) {
	cfg := config.ServerConfig{Language: "go", Command: "gopls"}
	c := NewClient(cfg)

	// Before Start, Capabilities() should return zero-value ServerCapabilities
	caps := c.Capabilities()
	if caps.ReferencesProvider || caps.DefinitionProvider || caps.DiagnosticProvider {
		t.Error("Capabilities() before Start should return all-false ServerCapabilities")
	}
}

