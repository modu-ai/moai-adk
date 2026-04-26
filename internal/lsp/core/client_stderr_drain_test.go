package core

import (
	"context"
	"io"
	"runtime"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
	"github.com/modu-ai/moai-adk/internal/lsp/subprocess"
	"github.com/modu-ai/moai-adk/internal/lsp/transport"
)

// ─── Test doubles ─────────────────────────────────────────────────────────────

// stderrSignalingReader is an io.ReadCloser that sends a signal on readCh on the first Read call.
// Used to observe whether the drain goroutine actually reads from stderr.
type stderrSignalingReader struct {
	inner  io.ReadCloser
	readCh chan struct{}
	once   bool
}

func newStderrSignalingReader(r io.ReadCloser) *stderrSignalingReader {
	return &stderrSignalingReader{inner: r, readCh: make(chan struct{}, 1)}
}

func (s *stderrSignalingReader) Read(p []byte) (int, error) {
	if !s.once {
		s.once = true
		select {
		case s.readCh <- struct{}{}:
		default:
		}
	}
	return s.inner.Read(p)
}

func (s *stderrSignalingReader) Close() error { return s.inner.Close() }

// nilStderrLauncher returns a LaunchResult with Stderr set to nil (AC-UTIL-003-002).
type nilStderrLauncher struct{}

func (n *nilStderrLauncher) Launch(_ context.Context, _ config.ServerConfig) (*subprocess.LaunchResult, error) {
	_, w1 := io.Pipe()
	r2, _ := io.Pipe()
	return &subprocess.LaunchResult{
		// Cmd is nil — Supervisor creation is skipped
		Stdin:  w1,
		Stdout: r2,
		Stderr: nil, // explicitly nil
	}, nil
}

// largeStderrLauncher returns a LaunchResult with a goroutine that writes 128 KiB + 1 byte to stderr.
// Without a drain goroutine, the write goroutine blocks permanently causing a deadlock (AC-UTIL-003-010).
type largeStderrLauncher struct {
	writeDone chan struct{} // closed when the stderr write completes
}

func newLargeStderrLauncher() *largeStderrLauncher {
	return &largeStderrLauncher{writeDone: make(chan struct{})}
}

func (l *largeStderrLauncher) Launch(_ context.Context, _ config.ServerConfig) (*subprocess.LaunchResult, error) {
	_, w1 := io.Pipe()
	r2, _ := io.Pipe()
	stderrR, stderrW := io.Pipe()

	// Simulate a "subprocess" writing 128 KiB + 1 byte to stderr.
	// io.Pipe has no buffer, so the first write blocks if nobody reads.
	// When a drain goroutine is present, it reads and allows the write to complete, then closes writeDone.
	go func() {
		defer close(l.writeDone)
		defer func() { _ = stderrW.Close() }()
		data := make([]byte, 128*1024+1)
		// Ignore write errors (pipe may be closed)
		_, _ = stderrW.Write(data)
	}()

	return &subprocess.LaunchResult{
		// Cmd is nil — Supervisor creation is skipped
		Stdin:  w1,
		Stdout: r2,
		Stderr: stderrR,
	}, nil
}

// ─── Tests ─────────────────────────────────────────────────────────────────

// TestStderrDrain_GoroutineObserved verifies that the drain goroutine actually reads
// from stderr when result.Stderr is non-nil (AC-UTIL-003-001).
func TestStderrDrain_GoroutineObserved(t *testing.T) {
	t.Parallel()

	stderrR, stderrW := io.Pipe()
	sigReader := newStderrSignalingReader(stderrR)

	ft := newFakeTransport()
	ft.setCallResponse("initialize", []byte(`{"capabilities":{}}`), nil)

	c := NewClient(
		config.ServerConfig{Language: "go"},
		WithLauncherFunc(func(_ context.Context, _ config.ServerConfig) (*subprocess.LaunchResult, error) {
			_, w1 := io.Pipe()
			r2, _ := io.Pipe()
			return &subprocess.LaunchResult{
				Stdin:  w1,
				Stdout: r2,
				Stderr: sigReader,
			}, nil
		}),
		WithTransportFactory(func(_ io.ReadWriteCloser) transport.Transport { return ft }),
	)

	ctx := context.Background()
	goroutinesBefore := runtime.NumGoroutine()
	if err := c.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v, want nil", err)
	}

	// Wait up to 2 seconds for the drain goroutine to read from stderr.
	// Verify via goroutine count increase OR signal channel.
	goroutinesAfter := runtime.NumGoroutine()

	// Write a small amount of data to stderr and wait for the readCh signal
	// to confirm that the goroutine count increased or the signal channel fired.
	go func() {
		defer func() { _ = stderrW.Close() }()
		stderrW.Write([]byte("drain me")) //nolint:errcheck
	}()

	select {
	case <-sigReader.readCh:
		// Drain goroutine actually read — pass
	case <-time.After(2 * time.Second):
		t.Errorf("drain goroutine did not read from stderr within 2s (goroutines: %d→%d)",
			goroutinesBefore, goroutinesAfter)
	}
}

// TestStderrDrain_NilGuard verifies that Start completes without panic when result.Stderr is nil (AC-UTIL-003-002).
func TestStderrDrain_NilGuard(t *testing.T) {
	t.Parallel()

	nl := &nilStderrLauncher{}
	ft := newFakeTransport()
	ft.setCallResponse("initialize", []byte(`{"capabilities":{}}`), nil)

	c := NewClient(
		config.ServerConfig{Language: "go"},
		WithLauncherFunc(nl.Launch),
		WithTransportFactory(func(_ io.ReadWriteCloser) transport.Transport { return ft }),
	)

	ctx := context.Background()
	// nil Stderr path — must not panic
	err := c.Start(ctx)
	if err != nil {
		t.Fatalf("Start() with nil Stderr error = %v, want nil", err)
	}
}

// TestStderrDrain_DeadlockPrevention verifies that when a subprocess writes 128 KiB to stderr,
// the drain goroutine allows completion within 5 seconds (AC-UTIL-003-010).
func TestStderrDrain_DeadlockPrevention(t *testing.T) {
	t.Parallel()

	ll := newLargeStderrLauncher()
	ft := newFakeTransport()
	ft.setCallResponse("initialize", []byte(`{"capabilities":{}}`), nil)

	c := NewClient(
		config.ServerConfig{Language: "go"},
		WithLauncherFunc(ll.Launch),
		WithTransportFactory(func(_ io.ReadWriteCloser) transport.Transport { return ft }),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start launches the drain goroutine. Without it, the largeStderrLauncher's
	// write goroutine blocks and writeDone is never closed.
	if err := c.Start(ctx); err != nil {
		// Non-nil from Start is allowed (depends on mock setup). The key is: no deadlock.
		t.Logf("Start() returned non-nil (expected for mock setup): %v", err)
	}

	// The 128 KiB write goroutine must complete within 5 seconds.
	// Without a drain goroutine, the first io.Pipe write blocks permanently.
	select {
	case <-ll.writeDone:
		// Normal: drain goroutine read from stderr, allowing the write goroutine to finish
	case <-ctx.Done():
		t.Fatal("deadlock: 128 KiB stderr write did not complete within 5s — drain goroutine missing?")
	}
}
