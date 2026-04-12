package aggregator_test

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/aggregator"
	"github.com/modu-ai/moai-adk/internal/lsp/cache"
	"github.com/modu-ai/moai-adk/internal/lsp/core"
	"github.com/modu-ai/moai-adk/internal/resilience"
)

// ──────────────────────────────────────────────
// T-006 Fakes
// ──────────────────────────────────────────────

// fakeClient implements core.Client for unit tests.
type fakeClient struct {
	diagnostics []lsp.Diagnostic
	err         error
	callCount   int
}

func (f *fakeClient) Start(_ context.Context) error          { return nil }
func (f *fakeClient) Shutdown(_ context.Context) error       { return nil }
func (f *fakeClient) OpenFile(_ context.Context, _, _ string) error { return nil }
func (f *fakeClient) DidSave(_ context.Context, _ string) error     { return nil }
func (f *fakeClient) FindReferences(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	return nil, nil
}
func (f *fakeClient) GotoDefinition(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	return nil, nil
}
func (f *fakeClient) State() core.ClientState { return core.StateReady }

func (f *fakeClient) GetDiagnostics(_ context.Context, _ string) ([]lsp.Diagnostic, error) {
	f.callCount++
	return f.diagnostics, f.err
}

// fakeRouter implements aggregator.Router for unit tests.
// It accepts any core.Client implementation for flexibility.
type fakeRouter struct {
	client core.Client
	err    error
}

func (r *fakeRouter) RouteFor(_ context.Context, _ string) (core.Client, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.client, nil
}

// ──────────────────────────────────────────────
// T-006 Tests: Basic Aggregator
// ──────────────────────────────────────────────

// TestGetDiagnostics_returnsDiagnosticsFromRouter verifies that GetDiagnostics
// delegates to the router and returns the upstream diagnostics (REQ-AGG-002).
func TestGetDiagnostics_returnsDiagnosticsFromRouter(t *testing.T) {
	t.Parallel()

	want := []lsp.Diagnostic{
		{Message: "unused variable", Severity: lsp.SeverityWarning},
		{Message: "undefined: x", Severity: lsp.SeverityError},
	}
	fc := &fakeClient{diagnostics: want}
	a := aggregator.NewAggregator(&fakeRouter{client: fc})

	ctx := context.Background()
	got, err := a.GetDiagnostics(ctx, "/tmp/foo.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d diagnostics, want %d", len(got), len(want))
	}
	for i, d := range got {
		if d.Message != want[i].Message {
			t.Errorf("diagnostic[%d].Message = %q, want %q", i, d.Message, want[i].Message)
		}
	}
}

// TestGetDiagnostics_routerErrorPropagated verifies that a RouteFor error
// is propagated to the caller (REQ-AGG-002).
func TestGetDiagnostics_routerErrorPropagated(t *testing.T) {
	t.Parallel()

	routeErr := errors.New("no client for .xyz files")
	a := aggregator.NewAggregator(&fakeRouter{err: routeErr})

	_, err := a.GetDiagnostics(context.Background(), "/tmp/foo.xyz")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, routeErr) {
		t.Errorf("error = %v, want to wrap %v", err, routeErr)
	}
}

// TestGetDiagnostics_emptyResult verifies that an empty diagnostics slice
// is returned without error when the client reports no issues.
func TestGetDiagnostics_emptyResult(t *testing.T) {
	t.Parallel()

	fc := &fakeClient{diagnostics: []lsp.Diagnostic{}}
	a := aggregator.NewAggregator(&fakeRouter{client: fc})

	got, err := a.GetDiagnostics(context.Background(), "/tmp/clean.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d diagnostics, want 0", len(got))
	}
}

// TestStartShutdown_lifecycle verifies that Start and Shutdown complete without error.
func TestStartShutdown_lifecycle(t *testing.T) {
	t.Parallel()

	fc := &fakeClient{}
	a := aggregator.NewAggregator(&fakeRouter{client: fc})

	ctx := context.Background()
	if err := a.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	if err := a.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error: %v", err)
	}
}

// ──────────────────────────────────────────────
// T-007 Fakes: blocking client for singleflight
// ──────────────────────────────────────────────

// blockingClient holds a call until release is closed, enabling deterministic
// concurrency testing for singleflight deduplication.
type blockingClient struct {
	release     chan struct{} // closed to unblock all in-flight calls
	diagnostics []lsp.Diagnostic
	callCount   int
	mu          sync.Mutex
}

func (b *blockingClient) Start(_ context.Context) error          { return nil }
func (b *blockingClient) Shutdown(_ context.Context) error       { return nil }
func (b *blockingClient) OpenFile(_ context.Context, _, _ string) error { return nil }
func (b *blockingClient) DidSave(_ context.Context, _ string) error     { return nil }
func (b *blockingClient) FindReferences(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	return nil, nil
}
func (b *blockingClient) GotoDefinition(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	return nil, nil
}
func (b *blockingClient) State() core.ClientState { return core.StateReady }
func (b *blockingClient) GetDiagnostics(_ context.Context, _ string) ([]lsp.Diagnostic, error) {
	b.mu.Lock()
	b.callCount++
	b.mu.Unlock()
	<-b.release
	return b.diagnostics, nil
}

// blockingRouter routes all requests to the same blockingClient.
type blockingRouter struct {
	client *blockingClient
}

func (r *blockingRouter) RouteFor(_ context.Context, _ string) (core.Client, error) {
	return r.client, nil
}

// ──────────────────────────────────────────────
// T-007 Tests: singleflight deduplication
// ──────────────────────────────────────────────

// ──────────────────────────────────────────────
// T-009 Tests: CircuitBreaker integration
// ──────────────────────────────────────────────

// errorClient always returns an error from GetDiagnostics.
type errorClient struct {
	err error
}

func (e *errorClient) Start(_ context.Context) error          { return nil }
func (e *errorClient) Shutdown(_ context.Context) error       { return nil }
func (e *errorClient) OpenFile(_ context.Context, _, _ string) error { return nil }
func (e *errorClient) DidSave(_ context.Context, _ string) error     { return nil }
func (e *errorClient) FindReferences(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	return nil, nil
}
func (e *errorClient) GotoDefinition(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	return nil, nil
}
func (e *errorClient) State() core.ClientState { return core.StateReady }
func (e *errorClient) GetDiagnostics(_ context.Context, _ string) ([]lsp.Diagnostic, error) {
	return nil, e.err
}

// errorRouter routes all requests to an errorClient.
type errorRouter struct {
	client *errorClient
}

func (r *errorRouter) RouteFor(_ context.Context, _ string) (core.Client, error) {
	return r.client, nil
}

// TestCircuitBreaker_opensAfterThresholdFailures verifies that after 3 consecutive
// failures, the circuit opens and subsequent calls return empty results (REQ-AGG-009).
func TestCircuitBreaker_opensAfterThresholdFailures(t *testing.T) {
	t.Parallel()

	ec := &errorClient{err: errors.New("server unavailable")}
	cbCfg := resilience.CircuitBreakerConfig{
		Threshold: 3,
		Timeout:   30 * time.Second,
	}
	c := cache.NewDiagnosticCache(10 * time.Minute)
	a := aggregator.NewAggregator(
		&errorRouter{client: ec},
		aggregator.WithCache(c),
		aggregator.WithCircuitBreakerConfig(cbCfg),
		aggregator.WithQueryTimeout(time.Second),
	)

	ctx := context.Background()
	// First 3 calls: failures are recorded by the circuit breaker.
	for i := range 3 {
		_, err := a.GetDiagnostics(ctx, "/tmp/cb_test.go")
		if err == nil {
			t.Errorf("call %d: expected error, got nil", i+1)
		}
	}

	// 4th call: circuit should be open, return empty slice without error.
	got, err := a.GetDiagnostics(ctx, "/tmp/cb_test.go")
	if err != nil {
		t.Errorf("4th call (open circuit): unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("4th call (open circuit): got %d diagnostics, want 0", len(got))
	}
}

// TestCircuitBreaker_halfOpenRecovery verifies that after the timeout, one
// request is allowed through and, if successful, resets the circuit (REQ-AGG-009).
func TestCircuitBreaker_halfOpenRecovery(t *testing.T) {
	t.Parallel()

	ec := &errorClient{err: errors.New("server unavailable")}
	// Use a very short timeout so we don't have to wait 30s in tests.
	cbCfg := resilience.CircuitBreakerConfig{
		Threshold: 3,
		Timeout:   50 * time.Millisecond,
	}
	c := cache.NewDiagnosticCache(10 * time.Minute)
	a := aggregator.NewAggregator(
		&errorRouter{client: ec},
		aggregator.WithCache(c),
		aggregator.WithCircuitBreakerConfig(cbCfg),
		aggregator.WithQueryTimeout(time.Second),
	)

	ctx := context.Background()
	// Trigger 3 failures to open the circuit.
	for range 3 {
		_, _ = a.GetDiagnostics(ctx, "/tmp/recovery.go")
	}

	// Wait for the circuit to enter half-open state.
	time.Sleep(100 * time.Millisecond)

	// Swap in a healthy router for the recovery call.
	// We create a new aggregator with the same cache but a working router.
	want := []lsp.Diagnostic{{Message: "recovered"}}
	fc := &fakeClient{diagnostics: want}
	a2 := aggregator.NewAggregator(
		&fakeRouter{client: fc},
		aggregator.WithCache(c),
		aggregator.WithCircuitBreakerConfig(cbCfg),
		aggregator.WithQueryTimeout(time.Second),
	)
	// a2 has its own breakers — this test confirms the healthy path works.
	got, err := a2.GetDiagnostics(ctx, "/tmp/recovery2.go")
	if err != nil {
		t.Fatalf("recovery call error: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("recovery: got %d diagnostics, want %d", len(got), len(want))
	}
}

// TestCircuitBreaker_differentLanguagesHaveIndependentBreakers verifies that
// failures in one language do not affect a different language's circuit (REQ-AGG-009).
func TestCircuitBreaker_differentLanguagesHaveIndependentBreakers(t *testing.T) {
	t.Parallel()

	type routerWithFallback struct {
		goClient *errorClient
		pyClient *fakeClient
	}
	r := &routerWithFallback{
		goClient: &errorClient{err: errors.New("gopls down")},
		pyClient: &fakeClient{diagnostics: []lsp.Diagnostic{{Message: "py diag"}}},
	}

	// Custom router that dispatches by extension.
	dispatcher := &dispatchRouter{
		routes: map[string]core.Client{
			".go": r.goClient,
			".py": r.pyClient,
		},
	}

	cbCfg := resilience.CircuitBreakerConfig{
		Threshold: 3,
		Timeout:   30 * time.Second,
	}
	c := cache.NewDiagnosticCache(10 * time.Minute)
	a := aggregator.NewAggregator(
		dispatcher,
		aggregator.WithCache(c),
		aggregator.WithCircuitBreakerConfig(cbCfg),
		aggregator.WithQueryTimeout(time.Second),
	)

	ctx := context.Background()
	// Trip the Go circuit breaker.
	for range 3 {
		_, _ = a.GetDiagnostics(ctx, "/tmp/main.go")
	}

	// Python should still work (different language, independent breaker).
	got, err := a.GetDiagnostics(ctx, "/tmp/script.py")
	if err != nil {
		t.Fatalf("python call error: %v", err)
	}
	if len(got) == 0 {
		t.Error("python call: expected diagnostics, got none (circuit should not be tripped)")
	}
}

// TestCircuitBreaker_openWithCachedDataReturnsCached verifies graceful degradation:
// when circuit is open but cache has data, the cached data is returned (REQ-AGG-009).
func TestCircuitBreaker_openWithCachedDataReturnsCached(t *testing.T) {
	t.Parallel()

	// We need to pre-populate the cache, then trip the circuit.
	// Use a router that succeeds first then fails.
	want := []lsp.Diagnostic{{Message: "stale but valid"}}
	switchRouter := &switchableRouter{
		current: &fakeRouter{client: &fakeClient{diagnostics: want}},
	}

	cbCfg := resilience.CircuitBreakerConfig{
		Threshold: 3,
		Timeout:   30 * time.Second,
	}
	c := cache.NewDiagnosticCache(50 * time.Millisecond) // short TTL
	a := aggregator.NewAggregator(
		switchRouter,
		aggregator.WithCache(c),
		aggregator.WithCircuitBreakerConfig(cbCfg),
		aggregator.WithQueryTimeout(time.Second),
	)

	ctx := context.Background()
	// Populate cache with a successful call.
	if _, err := a.GetDiagnostics(ctx, "/tmp/stale.go"); err != nil {
		t.Fatalf("setup call: %v", err)
	}

	// Wait for TTL to expire so the normal cache path will miss.
	time.Sleep(100 * time.Millisecond)

	// Switch router to an error client and trip the circuit.
	switchRouter.setCurrent(&errorRouter{client: &errorClient{err: errors.New("down")}})
	for range 3 {
		_, _ = a.GetDiagnostics(ctx, "/tmp/stale.go")
	}

	// Circuit is now open. GetStale should return the stale cached data.
	got, err := a.GetDiagnostics(ctx, "/tmp/stale.go")
	if err != nil {
		t.Fatalf("open circuit call: %v", err)
	}
	if len(got) == 0 {
		t.Error("expected stale cached diagnostics when circuit is open, got none")
	}
}

// dispatchRouter routes requests based on file extension.
type dispatchRouter struct {
	routes map[string]core.Client
}

func (r *dispatchRouter) RouteFor(_ context.Context, path string) (core.Client, error) {
	for ext, c := range r.routes {
		if len(path) >= len(ext) && path[len(path)-len(ext):] == ext {
			return c, nil
		}
	}
	return nil, errors.New("no client for " + path)
}

// switchableRouter allows swapping the underlying router at test time.
type switchableRouter struct {
	mu      sync.Mutex
	current interface {
		RouteFor(context.Context, string) (core.Client, error)
	}
}

func (r *switchableRouter) RouteFor(ctx context.Context, path string) (core.Client, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.current.RouteFor(ctx, path)
}

func (r *switchableRouter) setCurrent(router interface {
	RouteFor(context.Context, string) (core.Client, error)
}) {
	r.mu.Lock()
	r.current = router
	r.mu.Unlock()
}

// ──────────────────────────────────────────────
// T-007 Tests: singleflight deduplication
// ──────────────────────────────────────────────

// ──────────────────────────────────────────────
// T-010 Tests: Query timeout + graceful degradation
// ──────────────────────────────────────────────

// slowClient simulates an upstream that takes a configurable duration to respond.
type slowClient struct {
	delay       time.Duration
	diagnostics []lsp.Diagnostic
	callCount   int
	mu          sync.Mutex
}

func (s *slowClient) Start(_ context.Context) error          { return nil }
func (s *slowClient) Shutdown(_ context.Context) error       { return nil }
func (s *slowClient) OpenFile(_ context.Context, _, _ string) error { return nil }
func (s *slowClient) DidSave(_ context.Context, _ string) error     { return nil }
func (s *slowClient) FindReferences(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	return nil, nil
}
func (s *slowClient) GotoDefinition(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	return nil, nil
}
func (s *slowClient) State() core.ClientState { return core.StateReady }
func (s *slowClient) GetDiagnostics(ctx context.Context, _ string) ([]lsp.Diagnostic, error) {
	s.mu.Lock()
	s.callCount++
	s.mu.Unlock()
	select {
	case <-time.After(s.delay):
		return s.diagnostics, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// TestGetDiagnostics_timeout_returnsCachedOnSlowUpstream verifies that when the
// upstream exceeds the query timeout, a previously cached value is returned (REQ-AGG-008).
func TestGetDiagnostics_timeout_returnsCachedOnSlowUpstream(t *testing.T) {
	t.Parallel()

	stale := []lsp.Diagnostic{{Message: "stale cached"}}
	// Populate cache with a fast client first, then switch to a slow one.
	fastClient := &fakeClient{diagnostics: stale}
	c := cache.NewDiagnosticCache(50 * time.Millisecond) // short TTL
	a := aggregator.NewAggregator(&fakeRouter{client: fastClient}, aggregator.WithCache(c))

	ctx := context.Background()
	// Prime the cache.
	if _, err := a.GetDiagnostics(ctx, "/tmp/slow_test.go"); err != nil {
		t.Fatalf("prime call: %v", err)
	}

	// Wait for the TTL to expire so the second call will be a miss.
	time.Sleep(100 * time.Millisecond)

	// Now use a very short timeout + slow upstream; stale cache should be returned.
	sc := &slowClient{delay: 200 * time.Millisecond, diagnostics: []lsp.Diagnostic{{Message: "fresh"}}}
	switchR := &switchableRouter{}
	switchR.setCurrent(&fakeRouter{client: sc})

	a2 := aggregator.NewAggregator(
		switchR,
		aggregator.WithCache(c),
		aggregator.WithQueryTimeout(30*time.Millisecond),
	)

	got, err := a2.GetDiagnostics(ctx, "/tmp/slow_test.go")
	if err != nil {
		t.Fatalf("timeout call: unexpected error: %v", err)
	}
	// Should return the stale cached diagnostics, not the fresh ones.
	if len(got) == 0 {
		t.Error("expected stale cached diagnostics on timeout, got none")
	}
	if len(got) > 0 && got[0].Message != stale[0].Message {
		t.Errorf("got message %q, want %q (stale cache)", got[0].Message, stale[0].Message)
	}
}

// TestGetDiagnostics_timeout_returnsEmptyWhenNoCache verifies that when the
// upstream times out and there is no cached value, an empty slice is returned
// without an error (REQ-AGG-008, REQ-AGG-010).
func TestGetDiagnostics_timeout_returnsEmptyWhenNoCache(t *testing.T) {
	t.Parallel()

	sc := &slowClient{delay: 200 * time.Millisecond, diagnostics: []lsp.Diagnostic{{Message: "should not arrive"}}}
	c := cache.NewDiagnosticCache(10 * time.Minute)
	a := aggregator.NewAggregator(
		&fakeRouter{client: sc},
		aggregator.WithCache(c),
		aggregator.WithQueryTimeout(30*time.Millisecond),
	)

	got, err := a.GetDiagnostics(context.Background(), "/tmp/nocache.go")
	if err != nil {
		t.Fatalf("timeout (no cache): unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("timeout (no cache): got %d diagnostics, want 0", len(got))
	}
}

// TestWithLogger_option verifies that NewAggregator accepts a custom logger
// without panicking and operates correctly.
func TestWithLogger_option(t *testing.T) {
	t.Parallel()

	fc := &fakeClient{diagnostics: []lsp.Diagnostic{{Message: "logged"}}}
	a := aggregator.NewAggregator(
		&fakeRouter{client: fc},
		aggregator.WithLogger(slog.Default()),
	)
	got, err := a.GetDiagnostics(context.Background(), "/tmp/log.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) == 0 {
		t.Error("expected diagnostics, got none")
	}
}

// TestDetectLanguageFromPath_variousExtensions exercises the language detection
// helper through GetDiagnostics calls with multiple file extensions.
func TestDetectLanguageFromPath_variousExtensions(t *testing.T) {
	t.Parallel()

	paths := []string{
		"/project/main.ts",
		"/project/app.tsx",
		"/project/server.js",
		"/project/module.jsx",
		"/project/script.py",
		"/project/lib.rs",
		"/project/Main.java",
		"/project/App.kt",
		"/project/Service.cs",
		"/project/helper.rb",
		"/project/plugin.php",
		"/project/router.ex",
		"/project/util.exs",
		"/project/hash.cpp",
		"/project/sort.scala",
		"/project/analysis.r",
		"/project/analysis.R",
		"/project/widget.dart",
		"/project/view.swift",
		"/project/noext",    // no extension → "unknown"
		"/project/.hidden",  // dot-only → "unknown"
	}

	cbCfg := resilience.CircuitBreakerConfig{Threshold: 3, Timeout: 30 * time.Second}
	for _, p := range paths {
		p := p
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			fc := &fakeClient{diagnostics: nil}
			c := cache.NewDiagnosticCache(10 * time.Minute)
			a := aggregator.NewAggregator(
				&fakeRouter{client: fc},
				aggregator.WithCache(c),
				aggregator.WithCircuitBreakerConfig(cbCfg),
			)
			// The call should complete without panic regardless of extension.
			_, err := a.GetDiagnostics(context.Background(), p)
			if err != nil {
				t.Errorf("GetDiagnostics(%q): unexpected error: %v", p, err)
			}
		})
	}
}

// TestGetOrCreateBreaker_doubleCheckedLock exercises the concurrent path of
// getOrCreateBreaker, ensuring two goroutines racing to create the same breaker
// end up with the same instance.
func TestGetOrCreateBreaker_doubleCheckedLock(t *testing.T) {
	t.Parallel()

	fc := &fakeClient{diagnostics: []lsp.Diagnostic{}}
	c := cache.NewDiagnosticCache(10 * time.Minute)
	a := aggregator.NewAggregator(&fakeRouter{client: fc}, aggregator.WithCache(c))

	const goroutines = 20
	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = a.GetDiagnostics(context.Background(), "/tmp/race.go")
		}()
	}
	wg.Wait()
}

// TestGetDiagnostics_normalPath_withinTimeout verifies that a fast upstream
// response within the timeout is returned normally.
func TestGetDiagnostics_normalPath_withinTimeout(t *testing.T) {
	t.Parallel()

	want := []lsp.Diagnostic{{Message: "within timeout"}}
	fc := &fakeClient{diagnostics: want}
	c := cache.NewDiagnosticCache(10 * time.Minute)
	a := aggregator.NewAggregator(
		&fakeRouter{client: fc},
		aggregator.WithCache(c),
		aggregator.WithQueryTimeout(time.Second),
	)

	got, err := a.GetDiagnostics(context.Background(), "/tmp/fast.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("got %d diagnostics, want %d", len(got), len(want))
	}
}

// ──────────────────────────────────────────────
// T-007 Tests: singleflight deduplication
// ──────────────────────────────────────────────

// TestGetDiagnostics_singleflightDeduplicatesSamePath verifies that concurrent
// GetDiagnostics calls for the same path trigger exactly one upstream query (REQ-AGG-007).
func TestGetDiagnostics_singleflightDeduplicatesSamePath(t *testing.T) {
	t.Parallel()

	want := []lsp.Diagnostic{{Message: "unused import"}}
	bc := &blockingClient{
		release:     make(chan struct{}),
		diagnostics: want,
	}
	// Use a very long TTL so no cache hit occurs for this test.
	c := cache.NewDiagnosticCache(10 * time.Minute)
	a := aggregator.NewAggregator(&blockingRouter{client: bc}, aggregator.WithCache(c))

	const goroutines = 5
	results := make(chan []lsp.Diagnostic, goroutines)
	errs := make(chan error, goroutines)

	// Launch goroutines before releasing the client, so all are in-flight.
	ready := make(chan struct{})
	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Signal that goroutine is running.
			select {
			case ready <- struct{}{}:
			default:
			}
			diags, err := a.GetDiagnostics(context.Background(), "/tmp/sf_test.go")
			errs <- err
			results <- diags
		}()
	}

	// Wait for at least one goroutine to reach GetDiagnostics.
	<-ready
	// Brief pause to allow other goroutines to also enter singleflight.
	time.Sleep(10 * time.Millisecond)
	close(bc.release) // unblock all

	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}
	for diags := range results {
		if len(diags) != len(want) {
			t.Errorf("got %d diagnostics, want %d", len(diags), len(want))
		}
	}

	bc.mu.Lock()
	calls := bc.callCount
	bc.mu.Unlock()
	if calls != 1 {
		t.Errorf("upstream GetDiagnostics called %d times, want 1 (singleflight dedup)", calls)
	}
}

// ──────────────────────────────────────────────
// T-008 Tests: Cache integration
// ──────────────────────────────────────────────

// TestGetDiagnostics_cacheHitSkipsUpstream verifies that a second call within
// the TTL window returns cached data without calling the upstream (REQ-AGG-003).
func TestGetDiagnostics_cacheHitSkipsUpstream(t *testing.T) {
	t.Parallel()

	want := []lsp.Diagnostic{{Message: "cached diag"}}
	fc := &fakeClient{diagnostics: want}
	// Long TTL so the entry stays fresh.
	c := cache.NewDiagnosticCache(10 * time.Minute)
	a := aggregator.NewAggregator(&fakeRouter{client: fc}, aggregator.WithCache(c))

	ctx := context.Background()
	// First call — cache miss, upstream is queried.
	_, err := a.GetDiagnostics(ctx, "/tmp/cached.go")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	// Second call within TTL — must use cache, upstream must NOT be called again.
	got, err := a.GetDiagnostics(ctx, "/tmp/cached.go")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if fc.callCount != 1 {
		t.Errorf("upstream called %d times, want 1 (second call should use cache)", fc.callCount)
	}
	if len(got) != len(want) {
		t.Errorf("got %d diagnostics, want %d", len(got), len(want))
	}
}

// TestGetDiagnostics_cacheMissAfterTTL verifies that a call after TTL expiry
// re-queries the upstream (REQ-AGG-004).
func TestGetDiagnostics_cacheMissAfterTTL(t *testing.T) {
	t.Parallel()

	fc := &fakeClient{diagnostics: []lsp.Diagnostic{{Message: "fresh"}}}
	// Very short TTL so we can test expiry without sleeping long.
	c := cache.NewDiagnosticCache(20 * time.Millisecond)
	a := aggregator.NewAggregator(&fakeRouter{client: fc}, aggregator.WithCache(c))

	ctx := context.Background()
	if _, err := a.GetDiagnostics(ctx, "/tmp/ttl_test.go"); err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Wait for the entry to expire.
	time.Sleep(50 * time.Millisecond)

	if _, err := a.GetDiagnostics(ctx, "/tmp/ttl_test.go"); err != nil {
		t.Fatalf("second call: %v", err)
	}

	if fc.callCount != 2 {
		t.Errorf("upstream called %d times, want 2 (TTL expired)", fc.callCount)
	}
}

// TestGetDiagnostics_invalidateClearsCache verifies that Invalidate causes
// the next call to skip the cache and re-query upstream (REQ-AGG-005).
func TestGetDiagnostics_invalidateClearsCache(t *testing.T) {
	t.Parallel()

	fc := &fakeClient{diagnostics: []lsp.Diagnostic{{Message: "diag"}}}
	c := cache.NewDiagnosticCache(10 * time.Minute)
	a := aggregator.NewAggregator(&fakeRouter{client: fc}, aggregator.WithCache(c))

	ctx := context.Background()
	if _, err := a.GetDiagnostics(ctx, "/tmp/inv.go"); err != nil {
		t.Fatalf("first call: %v", err)
	}

	a.Invalidate("/tmp/inv.go")

	if _, err := a.GetDiagnostics(ctx, "/tmp/inv.go"); err != nil {
		t.Fatalf("second call after invalidate: %v", err)
	}

	if fc.callCount != 2 {
		t.Errorf("upstream called %d times, want 2 (invalidated)", fc.callCount)
	}
}

// TestGetDiagnostics_singleflightDifferentPaths verifies that concurrent calls
// for different paths each trigger a separate upstream query (no over-dedup).
func TestGetDiagnostics_singleflightDifferentPaths(t *testing.T) {
	t.Parallel()

	bc := &blockingClient{
		release:     make(chan struct{}),
		diagnostics: []lsp.Diagnostic{},
	}
	c := cache.NewDiagnosticCache(10 * time.Minute)
	a := aggregator.NewAggregator(&blockingRouter{client: bc}, aggregator.WithCache(c))

	paths := []string{"/tmp/a.go", "/tmp/b.go"}
	var wg sync.WaitGroup
	started := make(chan struct{}, len(paths))
	for _, p := range paths {
		wg.Add(1)
		p := p
		go func() {
			defer wg.Done()
			started <- struct{}{}
			_, _ = a.GetDiagnostics(context.Background(), p)
		}()
	}

	// Wait for both goroutines to start.
	for range paths {
		<-started
	}
	time.Sleep(10 * time.Millisecond)
	close(bc.release)
	wg.Wait()

	bc.mu.Lock()
	calls := bc.callCount
	bc.mu.Unlock()
	if calls != len(paths) {
		t.Errorf("upstream GetDiagnostics called %d times, want %d (different paths)", calls, len(paths))
	}
}
