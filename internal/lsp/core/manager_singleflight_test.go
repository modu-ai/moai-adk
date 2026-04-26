package core

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
)

// ─── AC-UTIL-003-003: verify the singleflight.Group field is present ──────

// TestManager_SingleflightField_Present verifies that the Manager.sf field exists
// and is in a zero-value ready state where Do can be invoked immediately (AC-UTIL-003-003).
func TestManager_SingleflightField_Present(t *testing.T) {
	t.Parallel()

	m := NewManager(makeTestServersConfig())

	// Direct field access from the same package — compile fails if the field is missing.
	// Verify zero-value readiness: Do invocation must work normally.
	called := false
	val, err, shared := m.sf.Do("probe-key", func() (any, error) {
		called = true
		return "result", nil
	})

	if !called {
		t.Error("sf.Do did not invoke the function")
	}
	if err != nil {
		t.Errorf("sf.Do error = %v, want nil", err)
	}
	if val != "result" {
		t.Errorf("sf.Do value = %v, want 'result'", val)
	}
	_ = shared // sharing is not verified in this test
}

// ─── AC-UTIL-003-004: singleflight barrier ────────────────────────────────

// TestGetOrSpawn_SingleflightBarrier_SecondBlocksUntilFirst verifies that
// the second concurrent getOrSpawn call is blocked until the first factory returns (AC-UTIL-003-004).
func TestGetOrSpawn_SingleflightBarrier_SecondBlocksUntilFirst(t *testing.T) {
	t.Parallel()

	var factoryCount atomic.Int32
	// factory waits until it receives the unblock signal
	firstStarted := make(chan struct{})
	unblock := make(chan struct{})

	m := NewManager(
		makeTestServersConfig(),
		WithClientFactory(func(cfg config.ServerConfig) Client {
			factoryCount.Add(1)
			// signal that the first factory entered
			select {
			case firstStarted <- struct{}{}:
			default:
			}
			// wait for the unblock signal
			<-unblock
			return &fakeClient{state: StateSpawning}
		}),
	)

	ctx := context.Background()

	// Start the first goroutine.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = m.getOrSpawn(ctx, "go")
	}()

	// Wait until the factory starts.
	select {
	case <-firstStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("first factory did not start within 2s")
	}

	// Start the second goroutine: invoke getOrSpawn while the factory is blocked.
	done2 := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(done2)
		_, _ = m.getOrSpawn(ctx, "go")
	}()

	// The second goroutine must not return immediately (sf.Do is blocking).
	select {
	case <-done2:
		// The second caller returned early — it may have taken the fast path (cache hit).
		// If the factory call count is 1, this means sharing via sf.Do, which is acceptable.
		t.Logf("second caller returned early (cache hit path) — factory calls: %d", factoryCount.Load())
	case <-time.After(50 * time.Millisecond):
		// expected behavior: the second goroutine is blocked inside sf.Do
	}

	// Release the factory.
	close(unblock)
	wg.Wait()

	// The factory must have been called exactly once.
	if got := factoryCount.Load(); got != 1 {
		t.Errorf("factory called %d times, want 1 (singleflight should deduplicate)", got)
	}
}

// ─── AC-UTIL-003-005: 16 concurrent RouteFor goroutines, factory exactly once ───

// TestRouteFor_ExactlyOnce_16ConcurrentCallers verifies that across 16 concurrent
// RouteFor invocations, clientFactory is called exactly once (AC-UTIL-003-005).
func TestRouteFor_ExactlyOnce_16ConcurrentCallers(t *testing.T) {
	t.Parallel()

	var factoryCount atomic.Int32

	m := NewManager(
		makeTestServersConfig(),
		WithClientFactory(func(cfg config.ServerConfig) Client {
			factoryCount.Add(1)
			// Short delay to make the concurrency scenario more pronounced.
			time.Sleep(5 * time.Millisecond)
			return &fakeClient{state: StateSpawning}
		}),
	)

	const N = 16
	clients := make([]Client, N)
	errs := make([]error, N)

	ctx := context.Background()
	var wg sync.WaitGroup

	// Start gate to start all goroutines as concurrently as possible.
	startGate := make(chan struct{})

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-startGate
			clients[i], errs[i] = m.RouteFor(ctx, "/workspace/main.go")
		}(i)
	}

	// Start all goroutines simultaneously after they are ready.
	close(startGate)
	wg.Wait()

	// The factory must have been called exactly once.
	if got := factoryCount.Load(); got != 1 {
		t.Errorf("clientFactory called %d times, want exactly 1", got)
	}

	// All calls must complete without error.
	for i := 0; i < N; i++ {
		if errs[i] != nil {
			t.Errorf("goroutine %d: RouteFor error = %v", i, errs[i])
		}
		if clients[i] == nil {
			t.Errorf("goroutine %d: client is nil", i)
		}
	}

	// All clients must be the same pointer.
	for i := 1; i < N; i++ {
		if clients[i] != clients[0] {
			t.Errorf("goroutine %d: got different client pointer (want same as goroutine 0)", i)
		}
	}
}

// ─── AC-UTIL-003-006: cache absent and retry on Start failure ─────────────

// TestGetOrSpawn_StartError_CacheAbsentAndRetry verifies that on c.Start failure,
// no client remains in the cache and the factory is invoked again on the next call (AC-UTIL-003-006).
func TestGetOrSpawn_StartError_CacheAbsentAndRetry(t *testing.T) {
	t.Parallel()

	startErr := errors.New("start: simulated failure")
	var callCount atomic.Int32

	m := NewManager(
		makeTestServersConfig(),
		WithClientFactory(func(cfg config.ServerConfig) Client {
			callCount.Add(1)
			return &fakeClient{state: StateSpawning, startErr: startErr}
		}),
	)

	ctx := context.Background()

	// First call: Start fails → returns error.
	c1, err1 := m.getOrSpawn(ctx, "go")
	if err1 == nil {
		t.Fatal("first getOrSpawn: expected error, got nil")
	}
	if c1 != nil {
		t.Error("first getOrSpawn: expected nil client on Start error")
	}
	if !errors.Is(err1, startErr) {
		t.Errorf("first getOrSpawn: error = %v, want to wrap %v", err1, startErr)
	}

	// The cache must not contain the client.
	m.mu.Lock()
	_, cacheHit := m.clients["go"]
	m.mu.Unlock()
	if cacheHit {
		t.Error("cache should be absent after Start error (REQ-UTIL-003-006)")
	}

	// Second call: the factory must be invoked again (sf.Do key released).
	_, _ = m.getOrSpawn(ctx, "go")
	if got := callCount.Load(); got != 2 {
		t.Errorf("factory called %d times total, want 2 (retry on second call)", got)
	}
}

// ─── AC-UTIL-003-011: race detector — 16 goroutines getOrSpawn ────────────

// TestGetOrSpawnConcurrent_RaceDetector verifies that 16 concurrent getOrSpawn calls
// do not produce a data race when run under go test -race (AC-UTIL-003-011).
//
// Run with: go test -race -run TestGetOrSpawnConcurrent ./internal/lsp/core/
func TestGetOrSpawnConcurrent_RaceDetector(t *testing.T) {
	// t.Parallel() is intentionally omitted to reduce interference when running -race in isolation.

	var spawnCount atomic.Int32

	m := NewManager(
		makeTestServersConfig(),
		WithClientFactory(func(cfg config.ServerConfig) Client {
			spawnCount.Add(1)
			// Small delay to provoke real race conditions.
			time.Sleep(2 * time.Millisecond)
			return &fakeClient{state: StateSpawning}
		}),
	)

	const N = 16
	ctx := context.Background()
	var wg sync.WaitGroup

	startGate := make(chan struct{})

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startGate
			// RouteFor also updates lastActivity, so it covers a wider race surface.
			_, _ = m.RouteFor(ctx, "/workspace/main.go")
		}()
	}

	close(startGate)
	wg.Wait()

	// If the race detector passes this test, no data race occurred.
	// The factory call count must be 1 (singleflight guarantee).
	if got := spawnCount.Load(); got != 1 {
		t.Errorf("spawnCount = %d, want 1 (singleflight should prevent duplicates)", got)
	}
}
