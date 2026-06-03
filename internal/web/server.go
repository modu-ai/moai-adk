// Package web implements the MoAI Web Console — a loopback-only browser surface
// for reading and writing the named MoAI settings (profile preferences plus the
// project-level user.yaml / language.yaml / statusline.yaml sections).
//
// The Console is a thin equivalent of the terminal profile wizard: it reuses the
// existing internal/profile read/write/sync functions and the canonical
// validation value lists. It never writes YAML directly and never introduces a
// parallel validation rule set.
//
// Scope boundary (REQ-WC-012): the Console touches ONLY profile preferences plus
// user.yaml / language.yaml / statusline.yaml. It never reads or writes
// quality.yaml, workflow.yaml, harness.yaml, git-strategy.yaml, or any other
// config section.
package web

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// loopbackHost is the only interface the Console ever binds to (REQ-WC-002).
// Binding to 0.0.0.0 or any non-loopback address is a forbidden anti-pattern.
const loopbackHost = "127.0.0.1"

// shutdownDrain is the maximum time the server waits for in-flight requests to
// drain on SIGINT/SIGTERM before closing the listener (REQ-WC-003).
const shutdownDrain = 5 * time.Second

// Config holds the runtime configuration for a Console server. ProjectRoot and
// ProfileBaseDir are injectable so tests can point the Console at temp dirs
// (t.TempDir) and never touch the real ~/.moai or project root.
type Config struct {
	// Port is the loopback TCP port to bind. 0 selects a random free port
	// (used by tests via httptest-style binding).
	Port int

	// NoOpen suppresses the browser auto-open attempt when true (REQ-WC-004).
	NoOpen bool

	// ProjectRoot is the directory whose .moai/config/sections/*.yaml files the
	// Console reads and writes via SyncToProjectConfig. Required.
	ProjectRoot string

	// ProfileBaseDir overrides the profile store base directory
	// (~/.moai/claude-profiles by default). When empty, the process default is
	// used. Tests set this to a temp dir for isolation.
	ProfileBaseDir string

	// ProfileName selects which profile's preferences the form reads/writes.
	// Empty means the "default" profile.
	ProfileName string

	// openBrowser is the browser-open hook. Injectable for tests; when nil the
	// real default-browser opener is used. It MUST NOT be fatal on failure
	// (REQ-WC-004): a returned error is logged and the server continues.
	openBrowser func(url string) error
}

// @MX:ANCHOR: [AUTO] Console HTTP 서버 엔트리 포인트 — 루프백 바인드 + graceful shutdown 불변식을 보장한다.
// @MX:REASON: [AUTO] fan_in≥3 (CLI web 서브커맨드 + 단위 테스트 + 통합 테스트가 호출). 127.0.0.1 전용 바인드(REQ-WC-002)와
// SIGINT/SIGTERM 5초 드레인 후 exit 0(REQ-WC-003) 계약을 이 함수가 단독으로 책임진다 — 0.0.0.0 바인드는 금지된 안티패턴.
//
// Run starts the Console server, blocks until SIGINT/SIGTERM (or ctx
// cancellation), then drains in-flight requests for up to shutdownDrain and
// returns nil on a clean shutdown. It binds exclusively to 127.0.0.1:<port>.
func Run(ctx context.Context, cfg Config) error {
	srv, err := NewServer(cfg)
	if err != nil {
		return err
	}
	return srv.ListenAndServe(ctx)
}

// Server is a running (or runnable) Console instance.
type Server struct {
	cfg     Config
	handler http.Handler

	// mu guards listener: bind() writes it from the serving goroutine while
	// Addr() may read it concurrently from another goroutine (tests, signal
	// handling). Race-free access is required.
	mu       sync.RWMutex
	listener net.Listener
}

// NewServer constructs a Console server from cfg. It validates required fields
// and builds the HTTP handler (routes + Host-check middleware) but does not yet
// bind a listener — call ListenAndServe (or Start for tests).
func NewServer(cfg Config) (*Server, error) {
	if cfg.ProjectRoot == "" {
		return nil, errors.New("web: ProjectRoot is required")
	}
	a := newApp(cfg)
	srv := &Server{
		cfg:     cfg,
		handler: a.routes(),
	}
	// REQ-WC4-005: wire the loopback-indicator address accessor. The app renders
	// {{.BindAddr}} from this closure at request time, so once the listener is
	// bound (including a random :0 test port) the indicator shows the real
	// 127.0.0.1:<port> address rather than a hardcoded placeholder.
	a.bindAddr = srv.displayBindAddr
	return srv, nil
}

// displayBindAddr returns the address the loopback indicator should show. After
// the listener is bound it returns the real 127.0.0.1:<port>; before bind (or
// when Port is 0 and no listener exists yet) it falls back to the configured
// loopback host + port so the GET / render is never blank (REQ-WC4-005).
func (s *Server) displayBindAddr() string {
	if addr := s.listenerAddr(); addr != "" {
		return addr
	}
	if s.cfg.Port > 0 {
		return fmt.Sprintf("%s:%d", loopbackHost, s.cfg.Port)
	}
	return loopbackHost
}

// Handler returns the fully-wired HTTP handler (with Host-check middleware).
// Exposed for httptest-based integration tests.
func (s *Server) Handler() http.Handler {
	return s.handler
}

// Addr returns the bound listener address, or an empty string when the server
// has not been started. Used by tests to assert the loopback bind (REQ-WC-002).
func (s *Server) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// bind opens the loopback listener for the configured port. It is the single
// place that constructs the bind address, guaranteeing the 127.0.0.1-only
// invariant (REQ-WC-002).
func (s *Server) bind() error {
	addr := fmt.Sprintf("%s:%d", loopbackHost, s.cfg.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("web: bind %s: %w", addr, err)
	}
	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()
	return nil
}

// listenerAddr returns the listener's address string under the read lock. Used
// internally by tryOpenBrowser to avoid a direct unsynchronized field read.
func (s *Server) listenerAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// ListenAndServe binds the loopback listener, optionally opens the browser, and
// serves until ctx is cancelled or a SIGINT/SIGTERM arrives. On signal it drains
// in-flight requests for up to shutdownDrain and returns nil (REQ-WC-003).
func (s *Server) ListenAndServe(ctx context.Context) error {
	if err := s.bind(); err != nil {
		return err
	}

	s.mu.RLock()
	ln := s.listener
	s.mu.RUnlock()

	httpSrv := &http.Server{
		Handler:           s.handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Browser auto-open (REQ-WC-004): attempt unless suppressed; never fatal.
	if !s.cfg.NoOpen {
		s.tryOpenBrowser()
	}

	// Signal-aware context: cancel on SIGINT/SIGTERM (REQ-WC-003).
	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		err := httpSrv.Serve(ln)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serveErr <- err
	}()

	select {
	case <-sigCtx.Done():
		// Drain in-flight requests for up to shutdownDrain, then close.
		drainCtx, cancel := context.WithTimeout(context.Background(), shutdownDrain)
		defer cancel()
		if err := httpSrv.Shutdown(drainCtx); err != nil {
			return fmt.Errorf("web: shutdown: %w", err)
		}
		<-serveErr
		return nil
	case err := <-serveErr:
		return err
	}
}

// tryOpenBrowser attempts to open the Console URL in the default browser. Any
// failure is non-fatal (REQ-WC-004): it is reported to stderr and the server
// continues serving.
func (s *Server) tryOpenBrowser() {
	url := fmt.Sprintf("http://%s", s.listenerAddr())
	open := s.cfg.openBrowser
	if open == nil {
		open = openDefaultBrowser
	}
	if err := open(url); err != nil {
		fmt.Fprintf(os.Stderr, "moai web: could not open browser (%v); open %s manually\n", err, url)
	}
}
