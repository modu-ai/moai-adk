package lsp

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- Mock types for server tests ---

// mockLauncher implements ServerLauncher for testing.
type mockLauncher struct {
	launchFn func(ctx context.Context, lang string) (Client, ProcessHandle, error)
}

func (l *mockLauncher) Launch(ctx context.Context, lang string) (Client, ProcessHandle, error) {
	return l.launchFn(ctx, lang)
}

// mockProcess implements ProcessHandle for testing.
// Wait() blocks until the done channel is closed (simulating process lifetime).
type mockProcess struct {
	killFn    func() error
	done      chan struct{}
	running   bool
	closeOnce sync.Once
}

func newMockProcess() *mockProcess {
	return &mockProcess{
		done:    make(chan struct{}),
		running: true,
	}
}

func (p *mockProcess) Kill() error {
	p.running = false
	p.closeOnce.Do(func() { close(p.done) })
	if p.killFn != nil {
		return p.killFn()
	}
	return nil
}

func (p *mockProcess) Wait() error {
	<-p.done
	return nil
}

func (p *mockProcess) IsRunning() bool {
	return p.running
}

// stop simulates process exit (crash or normal exit).
func (p *mockProcess) stop() {
	p.running = false
	p.closeOnce.Do(func() { close(p.done) })
}

// serverTestClient implements Client for server-level tests.
type serverTestClient struct {
	shutdownFn func(ctx context.Context) error
	diagFn     func(ctx context.Context, uri string) ([]Diagnostic, error)
}

// Compile-time interface compliance checks for serverTestClient.
var (
	_ Initializer         = (*serverTestClient)(nil)
	_ DiagnosticsProvider = (*serverTestClient)(nil)
	_ NavigationProvider  = (*serverTestClient)(nil)
	_ HoverProvider       = (*serverTestClient)(nil)
	_ SymbolsProvider     = (*serverTestClient)(nil)
	_ Client              = (*serverTestClient)(nil)
)

func (c *serverTestClient) Initialize(_ context.Context, _ string) error { return nil }

func (c *serverTestClient) Diagnostics(ctx context.Context, uri string) ([]Diagnostic, error) {
	if c.diagFn != nil {
		return c.diagFn(ctx, uri)
	}
	return []Diagnostic{}, nil
}

func (c *serverTestClient) References(_ context.Context, _ string, _ Position) ([]Location, error) {
	return []Location{}, nil
}

func (c *serverTestClient) Hover(_ context.Context, _ string, _ Position) (*HoverResult, error) {
	return nil, nil
}

func (c *serverTestClient) Definition(_ context.Context, _ string, _ Position) ([]Location, error) {
	return []Location{}, nil
}

func (c *serverTestClient) Symbols(_ context.Context, _ string) ([]DocumentSymbol, error) {
	return []DocumentSymbol{}, nil
}

func (c *serverTestClient) Shutdown(ctx context.Context) error {
	if c.shutdownFn != nil {
		return c.shutdownFn(ctx)
	}
	return nil
}

// cleanupProcess registers a t.Cleanup that stops the mock process,
// ensuring the watchProcess goroutine exits.
func cleanupProcess(t *testing.T, p *mockProcess) {
	t.Helper()
	t.Cleanup(func() { p.stop() })
}

// waitForCondition polls until fn returns true or timeout is reached.
func waitForCondition(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}

// --- Constructor tests ---

func TestNewServerManager(t *testing.T) {
	t.Parallel()

	launcher := &mockLauncher{}
	mgr := NewServerManager(launcher)

	if mgr == nil {
		t.Fatal("NewServerManager returned nil")
	}

	active := mgr.ActiveServers()
	if len(active) != 0 {
		t.Errorf("ActiveServers() = %v, want empty", active)
	}
}

func TestNewServerManagerWithOptions(t *testing.T) {
	t.Parallel()

	launcher := &mockLauncher{}
	mgr := NewServerManager(launcher, WithMaxParallel(8))

	sm, ok := mgr.(*serverManager)
	if !ok {
		t.Fatal("NewServerManager did not return *serverManager")
	}
	if sm.maxParallel != 8 {
		t.Errorf("maxParallel = %d, want 8", sm.maxParallel)
	}
}

func TestWithMaxParallelIgnoresNonPositive(t *testing.T) {
	t.Parallel()

	launcher := &mockLauncher{}
	mgr := NewServerManager(launcher, WithMaxParallel(0), WithMaxParallel(-1))

	sm := mgr.(*serverManager)
	if sm.maxParallel != defaultMaxParallel {
		t.Errorf("maxParallel = %d, want %d (default)", sm.maxParallel, defaultMaxParallel)
	}
}

// --- StartServer tests ---

func TestStartServer(t *testing.T) {
	t.Parallel()

	proc := newMockProcess()
	cleanupProcess(t, proc)

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, lang string) (Client, ProcessHandle, error) {
			if lang != "go" {
				t.Errorf("Launch called with lang=%q, want %q", lang, "go")
			}
			return &serverTestClient{}, proc, nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	if err := mgr.StartServer(ctx, "go"); err != nil {
		t.Fatalf("StartServer: %v", err)
	}

	active := mgr.ActiveServers()
	if len(active) != 1 || active[0] != "go" {
		t.Errorf("ActiveServers() = %v, want [go]", active)
	}
}

func TestStartServerAlreadyRunning(t *testing.T) {
	t.Parallel()

	proc := newMockProcess()
	cleanupProcess(t, proc)

	launchCount := 0
	launcher := &mockLauncher{
		launchFn: func(_ context.Context, _ string) (Client, ProcessHandle, error) {
			launchCount++
			return &serverTestClient{}, proc, nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	if err := mgr.StartServer(ctx, "go"); err != nil {
		t.Fatalf("first StartServer: %v", err)
	}
	if err := mgr.StartServer(ctx, "go"); err != nil {
		t.Fatalf("second StartServer: %v", err)
	}

	if launchCount != 1 {
		t.Errorf("Launch called %d times, want 1 (idempotent)", launchCount)
	}

	active := mgr.ActiveServers()
	if len(active) != 1 {
		t.Errorf("ActiveServers() has %d entries, want 1", len(active))
	}
}

func TestStartServerLaunchError(t *testing.T) {
	t.Parallel()

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, _ string) (Client, ProcessHandle, error) {
			return nil, nil, ErrServerStartFailed
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	err := mgr.StartServer(ctx, "haskell")
	if err == nil {
		t.Fatal("StartServer should return error for failed launch")
	}
	if !errors.Is(err, ErrServerStartFailed) {
		t.Errorf("error = %v, want ErrServerStartFailed", err)
	}

	active := mgr.ActiveServers()
	if len(active) != 0 {
		t.Errorf("ActiveServers() = %v, want empty after failed start", active)
	}
}

// --- StopServer tests ---

func TestStopServer(t *testing.T) {
	t.Parallel()

	proc := newMockProcess()
	cleanupProcess(t, proc)
	shutdownCalled := false

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, _ string) (Client, ProcessHandle, error) {
			client := &serverTestClient{
				shutdownFn: func(_ context.Context) error {
					shutdownCalled = true
					return nil
				},
			}
			return client, proc, nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	if err := mgr.StartServer(ctx, "go"); err != nil {
		t.Fatalf("StartServer: %v", err)
	}
	if err := mgr.StopServer(ctx, "go"); err != nil {
		t.Fatalf("StopServer: %v", err)
	}

	if !shutdownCalled {
		t.Error("client.Shutdown was not called")
	}

	active := mgr.ActiveServers()
	if len(active) != 0 {
		t.Errorf("ActiveServers() = %v, want empty after stop", active)
	}
}

func TestStopServerNotRunning(t *testing.T) {
	t.Parallel()

	launcher := &mockLauncher{}
	mgr := NewServerManager(launcher)

	err := mgr.StopServer(context.Background(), "haskell")
	if err != nil {
		t.Errorf("StopServer for non-running server: %v, want nil", err)
	}
}

func TestStopServerShutdownError(t *testing.T) {
	t.Parallel()

	proc := newMockProcess()
	cleanupProcess(t, proc)
	killCalled := false

	proc.killFn = func() error {
		killCalled = true
		return nil
	}

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, _ string) (Client, ProcessHandle, error) {
			client := &serverTestClient{
				shutdownFn: func(_ context.Context) error {
					return errors.New("shutdown failed")
				},
			}
			return client, proc, nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	if err := mgr.StartServer(ctx, "go"); err != nil {
		t.Fatalf("StartServer: %v", err)
	}
	if err := mgr.StopServer(ctx, "go"); err != nil {
		t.Fatalf("StopServer: %v, want nil (graceful degradation)", err)
	}

	if !killCalled {
		t.Error("process.Kill was not called after shutdown failure")
	}
}

// --- GetClient tests ---

func TestGetClient(t *testing.T) {
	t.Parallel()

	proc := newMockProcess()
	cleanupProcess(t, proc)
	expectedClient := &serverTestClient{}

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, _ string) (Client, ProcessHandle, error) {
			return expectedClient, proc, nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	if err := mgr.StartServer(ctx, "go"); err != nil {
		t.Fatalf("StartServer: %v", err)
	}

	client, err := mgr.GetClient("go")
	if err != nil {
		t.Fatalf("GetClient: %v", err)
	}
	if client != expectedClient {
		t.Error("GetClient returned different client instance")
	}
}

func TestGetClientNotRunning(t *testing.T) {
	t.Parallel()

	launcher := &mockLauncher{}
	mgr := NewServerManager(launcher)

	_, err := mgr.GetClient("python")
	if err == nil {
		t.Fatal("GetClient should return error for non-running server")
	}
	if !errors.Is(err, ErrServerNotRunning) {
		t.Errorf("error = %v, want ErrServerNotRunning", err)
	}
}

// --- ActiveServers tests ---

func TestActiveServers(t *testing.T) {
	t.Parallel()

	processes := make([]*mockProcess, 3)
	for i := range processes {
		processes[i] = newMockProcess()
		cleanupProcess(t, processes[i])
	}

	idx := 0
	launcher := &mockLauncher{
		launchFn: func(_ context.Context, _ string) (Client, ProcessHandle, error) {
			p := processes[idx]
			idx++
			return &serverTestClient{}, p, nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	for _, lang := range []string{"typescript", "go", "python"} {
		if err := mgr.StartServer(ctx, lang); err != nil {
			t.Fatalf("StartServer(%s): %v", lang, err)
		}
	}

	active := mgr.ActiveServers()
	want := []string{"go", "python", "typescript"} // sorted
	if len(active) != len(want) {
		t.Fatalf("ActiveServers() = %v, want %v", active, want)
	}
	for i, lang := range active {
		if lang != want[i] {
			t.Errorf("ActiveServers()[%d] = %q, want %q", i, lang, want[i])
		}
	}
}

func TestActiveServersEmpty(t *testing.T) {
	t.Parallel()

	launcher := &mockLauncher{}
	mgr := NewServerManager(launcher)

	active := mgr.ActiveServers()
	if len(active) != 0 {
		t.Errorf("ActiveServers() = %v, want empty", active)
	}
}

// --- HealthCheck tests ---

func TestHealthCheckAllHealthy(t *testing.T) {
	t.Parallel()

	procs := make(map[string]*mockProcess)
	for _, lang := range []string{"go", "python", "typescript", "rust"} {
		p := newMockProcess()
		cleanupProcess(t, p)
		procs[lang] = p
	}

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, lang string) (Client, ProcessHandle, error) {
			return &serverTestClient{}, procs[lang], nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	for lang := range procs {
		if err := mgr.StartServer(ctx, lang); err != nil {
			t.Fatalf("StartServer(%s): %v", lang, err)
		}
	}

	result := mgr.HealthCheck(ctx)
	if len(result) != 4 {
		t.Fatalf("HealthCheck returned %d entries, want 4", len(result))
	}
	for lang, err := range result {
		if err != nil {
			t.Errorf("HealthCheck[%s] = %v, want nil", lang, err)
		}
	}
}

func TestHealthCheckUnhealthy(t *testing.T) {
	t.Parallel()

	goProc := newMockProcess()
	cleanupProcess(t, goProc)

	rustProc := newMockProcess()
	cleanupProcess(t, rustProc)
	rustProc.running = false // simulate crashed process

	procs := map[string]*mockProcess{"go": goProc, "rust": rustProc}

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, lang string) (Client, ProcessHandle, error) {
			return &serverTestClient{}, procs[lang], nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	for lang := range procs {
		if err := mgr.StartServer(ctx, lang); err != nil {
			t.Fatalf("StartServer(%s): %v", lang, err)
		}
	}

	result := mgr.HealthCheck(ctx)
	if result["go"] != nil {
		t.Errorf("HealthCheck[go] = %v, want nil", result["go"])
	}
	if result["rust"] == nil {
		t.Error("HealthCheck[rust] = nil, want error")
	}
}

func TestHealthCheckEmpty(t *testing.T) {
	t.Parallel()

	launcher := &mockLauncher{}
	mgr := NewServerManager(launcher)

	result := mgr.HealthCheck(context.Background())
	if len(result) != 0 {
		t.Errorf("HealthCheck = %v, want empty map", result)
	}
}

// --- StartAll tests ---

func TestStartAll(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	procs := make(map[string]*mockProcess)

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, lang string) (Client, ProcessHandle, error) {
			p := newMockProcess()
			cleanupProcess(t, p)

			mu.Lock()
			procs[lang] = p
			mu.Unlock()

			return &serverTestClient{}, p, nil
		},
	}

	mgr := NewServerManager(launcher)
	langs := []string{"go", "python", "typescript", "rust"}

	if err := mgr.StartAll(context.Background(), langs); err != nil {
		t.Fatalf("StartAll: %v", err)
	}

	active := mgr.ActiveServers()
	if len(active) != 4 {
		t.Errorf("ActiveServers() has %d entries, want 4", len(active))
	}
}

func TestStartAllConcurrencyLimit(t *testing.T) {
	t.Parallel()

	var concurrent atomic.Int32
	var peak atomic.Int32

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, _ string) (Client, ProcessHandle, error) {
			n := concurrent.Add(1)
			defer concurrent.Add(-1)

			// Update peak concurrency.
			for {
				old := peak.Load()
				if n <= old || peak.CompareAndSwap(old, n) {
					break
				}
			}

			// Simulate server startup time.
			time.Sleep(20 * time.Millisecond)

			p := newMockProcess()
			cleanupProcess(t, p)
			return &serverTestClient{}, p, nil
		},
	}

	mgr := NewServerManager(launcher, WithMaxParallel(2))
	langs := []string{"go", "python", "typescript", "rust", "java", "kotlin"}

	if err := mgr.StartAll(context.Background(), langs); err != nil {
		t.Fatalf("StartAll: %v", err)
	}

	if peak.Load() > 2 {
		t.Errorf("peak concurrent launches = %d, want <= 2", peak.Load())
	}

	active := mgr.ActiveServers()
	if len(active) != 6 {
		t.Errorf("ActiveServers() has %d entries, want 6", len(active))
	}
}

func TestStartAllPartialFailure(t *testing.T) {
	t.Parallel()

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, lang string) (Client, ProcessHandle, error) {
			if lang == "zig" || lang == "haskell" {
				return nil, nil, ErrServerStartFailed
			}
			p := newMockProcess()
			cleanupProcess(t, p)
			return &serverTestClient{}, p, nil
		},
	}

	mgr := NewServerManager(launcher)
	langs := []string{"go", "zig", "python", "haskell"}

	err := mgr.StartAll(context.Background(), langs)
	if err != nil {
		t.Errorf("StartAll should return nil on partial failure, got %v", err)
	}

	active := mgr.ActiveServers()
	if len(active) != 2 {
		t.Errorf("ActiveServers() has %d entries, want 2 (go, python)", len(active))
	}

	// Verify correct servers are running.
	for _, lang := range active {
		if lang != "go" && lang != "python" {
			t.Errorf("unexpected active server: %s", lang)
		}
	}
}

// --- StopAll tests ---

func TestStopAll(t *testing.T) {
	t.Parallel()

	var shutdownCount atomic.Int32

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, _ string) (Client, ProcessHandle, error) {
			p := newMockProcess()
			cleanupProcess(t, p)
			client := &serverTestClient{
				shutdownFn: func(_ context.Context) error {
					shutdownCount.Add(1)
					return nil
				},
			}
			return client, p, nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	for _, lang := range []string{"go", "python", "typescript"} {
		if err := mgr.StartServer(ctx, lang); err != nil {
			t.Fatalf("StartServer(%s): %v", lang, err)
		}
	}

	if err := mgr.StopAll(ctx); err != nil {
		t.Fatalf("StopAll: %v", err)
	}

	active := mgr.ActiveServers()
	if len(active) != 0 {
		t.Errorf("ActiveServers() = %v, want empty after StopAll", active)
	}

	if shutdownCount.Load() != 3 {
		t.Errorf("Shutdown called %d times, want 3", shutdownCount.Load())
	}
}

// --- CollectAllDiagnostics tests ---

func TestCollectAllDiagnostics(t *testing.T) {
	t.Parallel()

	goDiags := []Diagnostic{
		{Range: Range{Start: Position{Line: 1}, End: Position{Line: 1}}, Severity: SeverityError, Source: "gopls", Message: "undefined: foo"},
	}
	pyDiags := []Diagnostic{
		{Range: Range{Start: Position{Line: 5}, End: Position{Line: 5}}, Severity: SeverityWarning, Source: "pyright", Message: "unused import"},
	}

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, lang string) (Client, ProcessHandle, error) {
			p := newMockProcess()
			cleanupProcess(t, p)

			var diags []Diagnostic
			switch lang {
			case "go":
				diags = goDiags
			case "python":
				diags = pyDiags
			}

			client := &serverTestClient{
				diagFn: func(_ context.Context, _ string) ([]Diagnostic, error) {
					return diags, nil
				},
			}
			return client, p, nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	for _, lang := range []string{"go", "python"} {
		if err := mgr.StartServer(ctx, lang); err != nil {
			t.Fatalf("StartServer(%s): %v", lang, err)
		}
	}

	result, err := mgr.CollectAllDiagnostics(ctx, "file:///project/main.go")
	if err != nil {
		t.Fatalf("CollectAllDiagnostics: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("got %d diagnostics, want 2", len(result))
	}

	// Verify both sources are represented.
	sources := make(map[string]bool)
	for _, d := range result {
		sources[d.Source] = true
	}
	if !sources["gopls"] || !sources["pyright"] {
		t.Errorf("sources = %v, want gopls and pyright", sources)
	}
}

func TestCollectAllDiagnosticsPartialError(t *testing.T) {
	t.Parallel()

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, lang string) (Client, ProcessHandle, error) {
			p := newMockProcess()
			cleanupProcess(t, p)

			client := &serverTestClient{
				diagFn: func(_ context.Context, _ string) ([]Diagnostic, error) {
					if lang == "rust" {
						return nil, errors.New("server error")
					}
					return []Diagnostic{
						{Source: lang, Message: "test diagnostic"},
					}, nil
				},
			}
			return client, p, nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	for _, lang := range []string{"go", "python", "rust"} {
		if err := mgr.StartServer(ctx, lang); err != nil {
			t.Fatalf("StartServer(%s): %v", lang, err)
		}
	}

	result, err := mgr.CollectAllDiagnostics(ctx, "file:///project/main.go")
	if err != nil {
		t.Errorf("CollectAllDiagnostics error = %v, want nil (non-fatal)", err)
	}

	// Should have diagnostics from go and python but not rust.
	if len(result) != 2 {
		t.Errorf("got %d diagnostics, want 2 (rust failed)", len(result))
	}
}

func TestCollectAllDiagnosticsNoServers(t *testing.T) {
	t.Parallel()

	launcher := &mockLauncher{}
	mgr := NewServerManager(launcher)

	result, err := mgr.CollectAllDiagnostics(context.Background(), "file:///project/main.go")
	if err != nil {
		t.Fatalf("CollectAllDiagnostics: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("got %d diagnostics, want 0", len(result))
	}
}

// --- Process crash detection test ---

func TestProcessCrashDetection(t *testing.T) {
	t.Parallel()

	proc := newMockProcess()
	// Do NOT use cleanupProcess here; we manually control the process lifecycle.

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, _ string) (Client, ProcessHandle, error) {
			return &serverTestClient{}, proc, nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	if err := mgr.StartServer(ctx, "go"); err != nil {
		t.Fatalf("StartServer: %v", err)
	}

	// Verify server is running.
	if _, err := mgr.GetClient("go"); err != nil {
		t.Fatalf("GetClient before crash: %v", err)
	}

	// Simulate process crash.
	proc.stop()

	// Wait for the watchProcess goroutine to detect the crash and remove the server.
	waitForCondition(t, 500*time.Millisecond, func() bool {
		_, err := mgr.GetClient("go")
		return errors.Is(err, ErrServerNotRunning)
	})

	// Verify server is removed from registry.
	active := mgr.ActiveServers()
	if len(active) != 0 {
		t.Errorf("ActiveServers() = %v, want empty after crash", active)
	}
}

// --- Concurrent access tests ---

func TestConcurrentGetClient(t *testing.T) {
	t.Parallel()

	proc := newMockProcess()
	cleanupProcess(t, proc)
	expectedClient := &serverTestClient{}

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, _ string) (Client, ProcessHandle, error) {
			return expectedClient, proc, nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	if err := mgr.StartServer(ctx, "go"); err != nil {
		t.Fatalf("StartServer: %v", err)
	}

	// Run 10 concurrent GetClient calls.
	var wg sync.WaitGroup
	errs := make(chan error, 10)

	for range 10 {
		wg.Go(func() {
			client, err := mgr.GetClient("go")
			if err != nil {
				errs <- err
				return
			}
			if client != expectedClient {
				errs <- errors.New("wrong client returned")
			}
		})
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent GetClient error: %v", err)
	}
}

func TestConcurrentStartStop(t *testing.T) {
	t.Parallel()

	launcher := &mockLauncher{
		launchFn: func(_ context.Context, _ string) (Client, ProcessHandle, error) {
			p := newMockProcess()
			cleanupProcess(t, p)
			return &serverTestClient{}, p, nil
		},
	}

	mgr := NewServerManager(launcher)
	ctx := context.Background()

	// Run concurrent start and stop operations.
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			lang := "go"
			if n%2 == 0 {
				_ = mgr.StartServer(ctx, lang)
			} else {
				_ = mgr.StopServer(ctx, lang)
			}
		}(i)
	}

	wg.Wait()

	// No panic, no race condition - test passes with -race flag.
}
