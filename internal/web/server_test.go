package web

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// newTestConfig returns a Config wired to a temp project root with a random
// free port and browser auto-open suppressed.
func newTestConfig(t *testing.T) Config {
	t.Helper()
	return Config{
		Port:           0, // random free port
		NoOpen:         true,
		ProjectRoot:    t.TempDir(),
		ProfileBaseDir: t.TempDir(),
	}
}

// TestNewServer_RequiresProjectRoot verifies the constructor rejects an empty
// ProjectRoot.
func TestNewServer_RequiresProjectRoot(t *testing.T) {
	_, err := NewServer(Config{ProjectRoot: ""})
	if err == nil {
		t.Fatal("expected error for empty ProjectRoot, got nil")
	}
}

// TestServer_LoopbackBindOnly verifies AC-WC-002 / REQ-WC-002: the listener binds
// to 127.0.0.1 and never to 0.0.0.0 or any non-loopback address.
func TestServer_LoopbackBindOnly(t *testing.T) {
	srv, err := NewServer(newTestConfig(t))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	if err := srv.bind(); err != nil {
		t.Fatalf("bind: %v", err)
	}
	t.Cleanup(func() { _ = srv.listener.Close() })

	addr := srv.Addr()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("SplitHostPort(%q): %v", addr, err)
	}

	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		t.Errorf("listener bound to %q, want a loopback address", addr)
	}
	if host == "0.0.0.0" || ip.IsUnspecified() {
		t.Errorf("listener bound to unspecified address %q — MUST be loopback only", addr)
	}
	if host != loopbackHost {
		t.Errorf("listener host = %q, want %q", host, loopbackHost)
	}
}

// TestServer_GracefulShutdownOnSignal verifies AC-WC-003 / REQ-WC-003:
// ListenAndServe returns nil (clean) when its context is cancelled (the same
// path a SIGINT/SIGTERM takes via signal.NotifyContext), draining within 5s.
func TestServer_GracefulShutdownOnSignal(t *testing.T) {
	srv, err := NewServer(newTestConfig(t))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	var serveErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		serveErr = srv.ListenAndServe(ctx)
	}()

	// Wait until the listener is bound.
	waitForAddr(t, srv)

	// Cancel the context — equivalent to receiving SIGINT/SIGTERM.
	cancel()

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
		if serveErr != nil {
			t.Errorf("ListenAndServe returned %v on graceful shutdown, want nil", serveErr)
		}
	case <-time.After(shutdownDrain + 2*time.Second):
		t.Fatal("ListenAndServe did not return within the drain window after cancellation")
	}
}

// TestServer_GracefulShutdownOnSIGTERM verifies that delivering a real SIGTERM to
// the process triggers the same clean shutdown path. The signal is caught by
// signal.NotifyContext inside ListenAndServe.
func TestServer_GracefulShutdownOnSIGTERM(t *testing.T) {
	srv, err := NewServer(newTestConfig(t))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	var wg sync.WaitGroup
	var serveErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		serveErr = srv.ListenAndServe(context.Background())
	}()

	waitForAddr(t, srv)

	// Deliver SIGTERM to our own process; signal.NotifyContext cancels.
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		t.Skipf("cannot deliver SIGTERM on this platform: %v", err)
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
		if serveErr != nil {
			t.Errorf("ListenAndServe returned %v on SIGTERM, want nil", serveErr)
		}
	case <-time.After(shutdownDrain + 2*time.Second):
		t.Fatal("ListenAndServe did not return after SIGTERM")
	}
}

// TestServer_PortConflictReturnsError verifies EC-1 / REQ-WC-002: binding a port
// already in use returns a clear error (not a panic).
func TestServer_PortConflictReturnsError(t *testing.T) {
	// Occupy a loopback port.
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("occupy port: %v", err)
	}
	t.Cleanup(func() { _ = occupied.Close() })

	_, port, _ := net.SplitHostPort(occupied.Addr().String())

	cfg := newTestConfig(t)
	cfg.Port = atoiOrZero(port)

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	err = srv.ListenAndServe(context.Background())
	if err == nil {
		t.Fatal("expected bind error on port conflict, got nil")
	}
	if !strings.Contains(err.Error(), "bind") {
		t.Errorf("error %q does not mention bind", err.Error())
	}
}

// TestServer_AutoOpenSuppressedByNoOpen verifies AC-WC-004 / REQ-WC-004: when
// NoOpen is set, the browser-open hook is never invoked.
func TestServer_AutoOpenSuppressedByNoOpen(t *testing.T) {
	var opened bool
	cfg := newTestConfig(t)
	cfg.NoOpen = true
	cfg.openBrowser = func(string) error { opened = true; return nil }

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); _ = srv.ListenAndServe(ctx) }()
	waitForAddr(t, srv)
	cancel()
	wg.Wait()

	if opened {
		t.Error("browser opener was invoked despite NoOpen=true")
	}
}

// TestServer_AutoOpenFailureNonFatal verifies AC-WC-004 / REQ-WC-004: when
// auto-open is enabled but the opener returns an error, the server keeps serving
// (failure is non-fatal) and shuts down cleanly.
func TestServer_AutoOpenFailureNonFatal(t *testing.T) {
	var attempted bool
	cfg := newTestConfig(t)
	cfg.NoOpen = false
	cfg.openBrowser = func(string) error {
		attempted = true
		return errors.New("no browser available")
	}

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	var serveErr error
	wg.Add(1)
	go func() { defer wg.Done(); serveErr = srv.ListenAndServe(ctx) }()
	waitForAddr(t, srv)

	// The server is still serving despite the (pending) open failure: a request
	// to / succeeds.
	resp, reqErr := httpGet(t, "http://"+srv.Addr()+"/")
	if reqErr != nil {
		t.Fatalf("GET / after auto-open failure: %v", reqErr)
	}
	_ = resp.Body.Close()

	cancel()
	wg.Wait()

	if !attempted {
		t.Error("browser opener was not attempted with NoOpen=false")
	}
	if serveErr != nil {
		t.Errorf("ListenAndServe returned %v, want nil (auto-open failure is non-fatal)", serveErr)
	}
}

// TestRun_RequiresProjectRoot verifies the package entry Run propagates the
// constructor's ProjectRoot requirement.
func TestRun_RequiresProjectRoot(t *testing.T) {
	err := Run(context.Background(), Config{ProjectRoot: ""})
	if err == nil {
		t.Fatal("expected error for empty ProjectRoot, got nil")
	}
}
