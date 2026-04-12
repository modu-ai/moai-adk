package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
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

	// OpenFile notifies the server that a file has been opened.
	// Not implemented in Sprint 3 — returns ErrNotImplemented.
	OpenFile(ctx context.Context, path, content string) error

	// GetDiagnostics returns diagnostics for the given file path.
	// Not implemented in Sprint 3 — returns ErrNotImplemented.
	GetDiagnostics(ctx context.Context, path string) ([]lsp.Diagnostic, error)

	// FindReferences returns all references to the symbol at pos in the given file.
	// Not implemented in Sprint 3 — returns ErrNotImplemented.
	FindReferences(ctx context.Context, path string, pos lsp.Position) ([]lsp.Location, error)

	// GotoDefinition returns the definition location for the symbol at pos.
	// Not implemented in Sprint 3 — returns ErrNotImplemented.
	GotoDefinition(ctx context.Context, path string, pos lsp.Position) ([]lsp.Location, error)

	// State returns the current lifecycle state of the client.
	State() ClientState
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
	logger          *slog.Logger
	serverCaps      ServerCapabilities
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
		logger:          slog.Default(),
		router:          transport.NewNotificationRouter(),
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

	// ready 상태 전환
	if err := c.state.Transition(StateReady); err != nil {
		return fmt.Errorf("lsp client start: state transition to ready: %w", err)
	}

	return nil
}

// initialize sends the LSP initialize request and parses server capabilities.
func (c *client) initialize(ctx context.Context) error {
	caps := DefaultClientCapabilities()

	params := map[string]any{
		"processId":    nil,
		"capabilities": caps,
		"rootUri":      nil,
	}
	if len(c.cfg.InitOptions) > 0 {
		params["initializationOptions"] = c.cfg.InitOptions
	}

	var result lsp.InitializeResult
	if err := transport.CallWithTimeout(ctx, c.tr, "initialize", params, &result, c.cfg.Language); err != nil {
		return fmt.Errorf("initialize request: %w", err)
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

// OpenFile is not implemented in Sprint 3.
func (c *client) OpenFile(_ context.Context, _, _ string) error {
	return fmt.Errorf("OpenFile: %w", ErrNotImplemented)
}

// GetDiagnostics is not implemented in Sprint 3.
func (c *client) GetDiagnostics(_ context.Context, _ string) ([]lsp.Diagnostic, error) {
	return nil, fmt.Errorf("GetDiagnostics: %w", ErrNotImplemented)
}

// FindReferences is not implemented in Sprint 3.
func (c *client) FindReferences(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	return nil, fmt.Errorf("FindReferences: %w", ErrNotImplemented)
}

// GotoDefinition is not implemented in Sprint 3.
func (c *client) GotoDefinition(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	return nil, fmt.Errorf("GotoDefinition: %w", ErrNotImplemented)
}

// State returns the current lifecycle state.
func (c *client) State() ClientState {
	return c.state.Current()
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
