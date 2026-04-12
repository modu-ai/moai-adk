package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"syscall"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/config"
	"github.com/modu-ai/moai-adk/internal/lsp/subprocess"
	"github.com/modu-ai/moai-adk/internal/lsp/transport"
)

// launchFunc은 subprocess.Launcher.Launch와 동일한 시그니처의 함수형 타입.
// 테스트에서 의존성 주입을 위해 사용합니다.
type launchFunc func(ctx context.Context, cfg config.ServerConfig) (*subprocess.LaunchResult, error)

// transportFactory은 io.ReadWriteCloser를 받아 Transport를 생성하는 함수형 타입.
// 테스트에서 fakeTransport를 주입하기 위해 사용합니다.
type transportFactory func(stream io.ReadWriteCloser) transport.Transport

// supervisorIface abstracts subprocess.Supervisor for testability.
//
// @MX:NOTE: [AUTO] supervisorIface — allows faking process supervision in unit tests without real subprocesses
type supervisorIface interface {
	Watch(ctx context.Context) <-chan subprocess.ExitEvent
	Signal(sig os.Signal) error
	Kill() error
}

// Client is the public interface for interacting with a language server.
// It manages the full lifecycle: spawn, initialize, query, and shutdown.
//
// Implementations are safe for concurrent use after Start returns nil.
//
// @MX:ANCHOR: [AUTO] Client interface — primary LSP client contract consumed across the system
// @MX:REASON: fan_in >= 3 projected — Ralph engine, Quality Gates, LOOP command, MCP bridge, and integration tests all consume this interface
type Client interface {
	// Start spawns the language server subprocess and performs the LSP initialize
	// handshake. Returns an error if the binary is not found (ErrBinaryNotFound
	// surfaces to caller) or if initialization fails.
	Start(ctx context.Context) error

	// Shutdown performs a graceful LSP shutdown: sends shutdown request, exit
	// notification, sends SIGTERM to the subprocess, and closes the transport.
	// If the subprocess does not exit within shutdownTimeout, SIGKILL is sent.
	Shutdown(ctx context.Context) error

	// OpenFile notifies the server that a file has been opened or changed.
	// Sends textDocument/didOpen on first call for a path, or textDocument/didChange
	// when content differs from the last known version.
	// Same content on a previously opened file is a no-op (REQ-LC-006).
	OpenFile(ctx context.Context, path, content string) error

	// DidSave notifies the server that a tracked file has been saved (REQ-LC-023).
	// Returns an error if the file is not currently tracked by OpenFile.
	DidSave(ctx context.Context, path string) error

	// GetDiagnostics returns diagnostics for the given file path.
	// Uses push model: returns the latest diagnostics received via publishDiagnostics.
	// Returns ErrFileNotOpen if the file has not been opened via OpenFile.
	GetDiagnostics(ctx context.Context, path string) ([]lsp.Diagnostic, error)

	// FindReferences returns all references to the symbol at pos in the given file.
	// Returns ErrCapabilityUnsupported if the server does not advertise references support.
	FindReferences(ctx context.Context, path string, pos lsp.Position) ([]lsp.Location, error)

	// GotoDefinition returns the definition location for the symbol at pos.
	// Returns ErrCapabilityUnsupported if the server does not advertise definition support.
	GotoDefinition(ctx context.Context, path string, pos lsp.Position) ([]lsp.Location, error)

	// State returns the current lifecycle state of the client.
	State() ClientState

	// Capabilities returns the ServerCapabilities parsed from the LSP initialize response.
	// Returns zero-value ServerCapabilities when called before Start (REQ-LM-009).
	Capabilities() ServerCapabilities
}

// client is the concrete implementation of Client.
type client struct {
	cfg             config.ServerConfig
	launcher        launchFunc
	trFactory       transportFactory
	tr              transport.Transport
	supervisor      supervisorIface
	state           *StateMachine
	router          *transport.NotificationRouter
	shutdownTimeout time.Duration
	idleTimeout     time.Duration
	logger          *slog.Logger
	serverCaps      ServerCapabilities

	// docs는 열린 문서 캐시입니다. T-011 이후 사용됩니다.
	docs *documentCache

	// diagnostics는 서버가 push한 textDocument/publishDiagnostics 결과를 저장합니다.
	// T-013 이후 사용됩니다.
	diagnostics   map[string][]lsp.Diagnostic
	diagnosticsMu sync.RWMutex
}

// Option is a functional option for configuring a client.
type Option func(*client)

// WithLauncherFunc sets a custom launch function (used for dependency injection in tests).
func WithLauncherFunc(fn launchFunc) Option {
	return func(c *client) {
		c.launcher = fn
	}
}

// WithTransportFactory sets a custom transport factory (used for dependency injection in tests).
func WithTransportFactory(fn transportFactory) Option {
	return func(c *client) {
		c.trFactory = fn
	}
}

// WithShutdownTimeout sets the timeout for graceful shutdown before SIGKILL.
func WithShutdownTimeout(d time.Duration) Option {
	return func(c *client) {
		c.shutdownTimeout = d
	}
}

// WithIdleTimeout sets the idle timeout after which an open file is sent textDocument/didClose
// by the idle reaper (REQ-LC-022). Default: 5 minutes.
func WithIdleTimeout(d time.Duration) Option {
	return func(c *client) {
		c.idleTimeout = d
	}
}

// WithLogger sets the logger for state transitions and lifecycle events.
func WithLogger(logger *slog.Logger) Option {
	return func(c *client) {
		c.logger = logger
		c.state = NewStateMachine(logger)
	}
}

// withSupervisorIface injects a supervisorIface — used for testing the Shutdown path.
// Not exported; use only in _test.go files.
func withSupervisorIface(sv supervisorIface) Option {
	return func(c *client) {
		c.supervisor = sv
	}
}

// NewClient creates a new client for the given ServerConfig.
// Default values: shutdownTimeout=10s, launcher=subprocess.NewLauncher().Launch,
// trFactory=transport.NewPowernapTransport, logger=slog.Default().
//
// @MX:ANCHOR: [AUTO] NewClient — factory function, primary entry point for client construction
// @MX:REASON: fan_in >= 3 — Manager, integration tests, CLI diagnostics command, and MCP bridge all call NewClient
func NewClient(cfg config.ServerConfig, opts ...Option) *client {
	c := &client{
		cfg:             cfg,
		shutdownTimeout: 10 * time.Second,
		idleTimeout:     5 * time.Minute,
		logger:          slog.Default(),
		router:          transport.NewNotificationRouter(),
		docs:            newDocumentCache(),
		diagnostics:     make(map[string][]lsp.Diagnostic),
	}
	c.state = NewStateMachine(c.logger)

	// 기본 launcher: 실제 subprocess.Launcher 사용
	defaultLauncher := subprocess.NewLauncher()
	c.launcher = defaultLauncher.Launch

	// 기본 transport factory: powernap Transport
	c.trFactory = transport.NewPowernapTransport

	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Start spawns the language server and performs the LSP initialize handshake.
//
// State transitions:
//   - spawning → (launch) → initializing → (initialize) → ready
//   - on ErrBinaryNotFound: spawning → shutdown
//   - on initialize failure: initializing → degraded
func (c *client) Start(ctx context.Context) error {
	// subprocess 실행
	result, err := c.launcher(ctx, c.cfg)
	if err != nil {
		// ErrBinaryNotFound: warn_and_skip 패턴 (REQ-LC-004)
		if errors.Is(err, subprocess.ErrBinaryNotFound) {
			_ = c.state.Transition(StateShutdown)
			return fmt.Errorf("lsp client start (lang=%s): %w", c.cfg.Language, err)
		}
		_ = c.state.Transition(StateShutdown)
		return fmt.Errorf("lsp client start (lang=%s): launch: %w", c.cfg.Language, err)
	}

	// Supervisor 생성 (Cmd가 nil이면 skip — 테스트 전용 경로)
	if result.Cmd != nil {
		c.supervisor = subprocess.NewSupervisor(result)
	}

	// subprocess stdio로 Transport 생성
	stream := &readWriteCloser{
		r: result.Stdout,
		w: result.Stdin,
		closers: []io.Closer{result.Stdin, result.Stdout},
	}
	c.tr = c.trFactory(stream)

	// initializing 상태 전환
	if err := c.state.Transition(StateInitializing); err != nil {
		c.cleanupTransport()
		return fmt.Errorf("lsp client start: state transition: %w", err)
	}

	// LSP initialize 요청 전송
	if err := c.initialize(ctx); err != nil {
		_ = c.state.Transition(StateDegraded)
		return fmt.Errorf("lsp client initialize (lang=%s): %w", c.cfg.Language, err)
	}

	// publishDiagnostics 핸들러 등록: 서버가 push한 진단 결과를 diagnostics 캐시에 저장합니다.
	// @MX:NOTE: [AUTO] NotificationRouter를 통해 비동기 publishDiagnostics 알림을 진단 캐시로 라우팅
	_ = c.router.RegisterPublishDiagnostics(func(uri string, diags []lsp.Diagnostic) error {
		c.diagnosticsMu.Lock()
		c.diagnostics[uri] = diags
		c.diagnosticsMu.Unlock()
		return nil
	})
	c.router.Attach(c.tr)

	// ready 상태 전환
	if err := c.state.Transition(StateReady); err != nil {
		return fmt.Errorf("lsp client start: state transition to ready: %w", err)
	}

	return nil
}

// initialize sends the LSP initialize request and parses server capabilities.
func (c *client) initialize(ctx context.Context) error {
	caps := DefaultClientCapabilities()

	// rootUri + workspaceFolders: cfg.RootDir이 설정되어 있으면 file:// URI로 변환.
	// gopls 등 서버는 workspaceFolders를 더 신뢰성 있게 처리하므로 두 필드 모두 전달.
	var rootURI any
	var workspaceFolders any
	if c.cfg.RootDir != "" {
		uri := pathToURI(c.cfg.RootDir)
		rootURI = uri
		workspaceFolders = []map[string]any{
			{"uri": uri, "name": c.cfg.Language},
		}
	}

	params := map[string]any{
		"processId":        nil,
		"capabilities":     caps,
		"rootUri":          rootURI,
		"workspaceFolders": workspaceFolders,
	}
	if len(c.cfg.InitOptions) > 0 {
		params["initializationOptions"] = c.cfg.InitOptions
	}

	var result lsp.InitializeResult
	if err := transport.CallWithTimeout(ctx, c.tr, "initialize", params, &result, c.cfg.Language); err != nil {
		return fmt.Errorf("initialize request: %w", err)
	}

	// LSP 프로토콜: initialize 응답 수신 후 반드시 initialized 알림 전송 (fire-and-forget)
	// 이 알림 없이는 gopls 등 서버가 workspace를 활성화하지 않음
	if notifyErr := c.tr.Notify(ctx, "initialized", map[string]any{}); notifyErr != nil {
		c.logger.Warn("lsp: initialized notification failed",
			slog.String("language", c.cfg.Language),
			slog.String("error", notifyErr.Error()),
		)
	}

	// サーバーケイパビリティ解析
	serverCaps, err := ParseServerCapabilities(result.Capabilities)
	if err != nil {
		// Parse failure is non-fatal — degrade gracefully
		c.logger.Warn("lsp: failed to parse server capabilities; using empty defaults",
			slog.String("language", c.cfg.Language),
			slog.String("error", err.Error()),
		)
	} else {
		c.serverCaps = serverCaps
	}

	return nil
}

// Shutdown performs graceful shutdown of the language server.
//
// State transitions: any → shutdown
func (c *client) Shutdown(ctx context.Context) error {
	// transport이 없으면 (Start 전 호출) 즉시 shutdown 전환
	if c.tr == nil {
		_ = c.state.Transition(StateShutdown)
		return nil
	}

	// LSP shutdown 요청 (best-effort: 에러는 로그만)
	shutCtx, cancel := context.WithTimeout(ctx, c.shutdownTimeout)
	defer cancel()

	if err := transport.CallWithTimeout(shutCtx, c.tr, "shutdown", nil, nil, c.cfg.Language); err != nil {
		c.logger.Warn("lsp: shutdown request failed",
			slog.String("language", c.cfg.Language),
			slog.String("error", err.Error()),
		)
	}

	// exit 알림 (fire-and-forget)
	_ = c.tr.Notify(ctx, "exit", nil)

	// Supervisor를 통해 프로세스 종료
	if c.supervisor != nil {
		if err := c.supervisor.Signal(syscall.SIGTERM); err != nil {
			c.logger.Warn("lsp: SIGTERM failed, sending SIGKILL",
				slog.String("language", c.cfg.Language),
				slog.String("error", err.Error()),
			)
			_ = c.supervisor.Kill()
		} else {
			// shutdownTimeout 내에 종료되지 않으면 SIGKILL
			killTimer := time.NewTimer(c.shutdownTimeout)
			defer killTimer.Stop()

			watchCtx, watchCancel := context.WithCancel(context.Background())
			defer watchCancel()

			exitCh := c.supervisor.Watch(watchCtx)
			select {
			case <-exitCh:
				// 정상 종료
			case <-killTimer.C:
				_ = c.supervisor.Kill()
			}
		}
	}

	c.cleanupTransport()
	_ = c.state.Transition(StateShutdown)
	return nil
}

// cleanupTransport closes the transport if it exists.
func (c *client) cleanupTransport() {
	if c.tr != nil {
		_ = c.tr.Close()
	}
}

// OpenFile notifies the server that a file has been opened or changed (REQ-LC-002a, REQ-LC-020, REQ-LC-021).
// Delegates to documentCache.openOrChange which selects didOpen or didChange based on cache state.
func (c *client) OpenFile(ctx context.Context, path, content string) error {
	uri := pathToURI(path)
	langID := resolveLanguageID(c.cfg.Language)
	return c.docs.openOrChange(ctx, c.tr, uri, langID, content)
}

// DidSave notifies the server that a tracked file has been saved (REQ-LC-023).
func (c *client) DidSave(ctx context.Context, path string) error {
	uri := pathToURI(path)
	return c.docs.didSave(ctx, c.tr, uri)
}

// GetDiagnostics is implemented in queries.go (T-013).

// FindReferences and GotoDefinition are implemented in queries.go (T-014).

// State returns the current lifecycle state.
func (c *client) State() ClientState {
	return c.state.Current()
}

// Capabilities returns the ServerCapabilities parsed from the LSP initialize response
// (REQ-LM-009). Returns a zero-value ServerCapabilities before Start completes.
func (c *client) Capabilities() ServerCapabilities {
	return c.serverCaps
}

// readWriteCloser combines a reader and writer into io.ReadWriteCloser for transport.
type readWriteCloser struct {
	r       io.ReadCloser
	w       io.WriteCloser
	closers []io.Closer
}

func (rwc *readWriteCloser) Read(p []byte) (n int, err error) {
	return rwc.r.Read(p)
}

func (rwc *readWriteCloser) Write(p []byte) (n int, err error) {
	return rwc.w.Write(p)
}

func (rwc *readWriteCloser) Close() error {
	var lastErr error
	for _, c := range rwc.closers {
		if err := c.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
