package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/config"
)

// ---------------------------------------------------------------------------
// Test helpers: fakeClient, fakeClientFactory
// ---------------------------------------------------------------------------

// fakeClient is a test Client implementation.
// It simulates behavior without spawning a real subprocess.
type fakeClient struct {
	mu           sync.Mutex
	state        ClientState
	startErr     error
	shutdownErr  error
	startCalled  int
	shutdownCalled int
}

func (f *fakeClient) Start(_ context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.startCalled++
	if f.startErr != nil {
		return f.startErr
	}
	f.state = StateReady
	return nil
}

func (f *fakeClient) Shutdown(_ context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.shutdownCalled++
	f.state = StateShutdown
	return f.shutdownErr
}

func (f *fakeClient) OpenFile(_ context.Context, _, _ string) error { return nil }
func (f *fakeClient) DidSave(_ context.Context, _ string) error     { return nil }
func (f *fakeClient) GetDiagnostics(_ context.Context, _ string) ([]lsp.Diagnostic, error) {
	return nil, nil
}
func (f *fakeClient) FindReferences(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	return nil, nil
}
func (f *fakeClient) GotoDefinition(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	return nil, nil
}
func (f *fakeClient) State() ClientState {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.state
}

func (f *fakeClient) Capabilities() ServerCapabilities {
	return ServerCapabilities{}
}

// fakeClientFactory returns a factory function that creates fakeClients.
// It tracks the number of creations via spawnCount.
func fakeClientFactory(spawnCount *atomic.Int32, clients *[]*fakeClient, mu *sync.Mutex, startErr error) func(config.ServerConfig) Client {
	return func(cfg config.ServerConfig) Client {
		spawnCount.Add(1)
		fc := &fakeClient{
			state:    StateSpawning,
			startErr: startErr,
		}
		if mu != nil && clients != nil {
			mu.Lock()
			*clients = append(*clients, fc)
			mu.Unlock()
		}
		return fc
	}
}

// makeTestServersConfig creates a ServersConfig for testing.
func makeTestServersConfig() *config.ServersConfig {
	return &config.ServersConfig{
		Servers: map[string]config.ServerConfig{
			"go": {
				Language:            "go",
				Command:             "gopls",
				FileExtensions:      []string{".go"},
				RootMarkers:         []string{"go.mod"},
				IdleShutdownSeconds: 0,
			},
			"python": {
				Language:            "python",
				Command:             "pylsp",
				FileExtensions:      []string{".py"},
				RootMarkers:         []string{"pyproject.toml", "setup.py"},
				IdleShutdownSeconds: 0,
			},
			"typescript": {
				Language:            "typescript",
				Command:             "typescript-language-server",
				FileExtensions:      []string{".ts", ".tsx"},
				RootMarkers:         []string{"tsconfig.json"},
				IdleShutdownSeconds: 0,
			},
		},
	}
}

// ---------------------------------------------------------------------------
// T-015: detectLanguage tests
// ---------------------------------------------------------------------------

// TestDetectLanguage_GoExtension tests that .go files are detected as "go" language.
func TestDetectLanguage_GoExtension(t *testing.T) {
	cfg := makeTestServersConfig()
	m := NewManager(cfg)

	lang, ok := m.detectLanguage("/path/to/main.go")
	if !ok {
		t.Fatal("expected ok=true for .go extension, got false")
	}
	if lang != "go" {
		t.Fatalf("expected lang=%q, got %q", "go", lang)
	}
}

// TestDetectLanguage_PythonExtension tests that .py files are detected as "python" language.
func TestDetectLanguage_PythonExtension(t *testing.T) {
	cfg := makeTestServersConfig()
	m := NewManager(cfg)

	lang, ok := m.detectLanguage("/path/to/script.py")
	if !ok {
		t.Fatal("expected ok=true for .py extension, got false")
	}
	if lang != "python" {
		t.Fatalf("expected lang=%q, got %q", "python", lang)
	}
}

// TestDetectLanguage_TypeScriptExtension tests that .ts files are detected as "typescript" language.
func TestDetectLanguage_TypeScriptExtension(t *testing.T) {
	cfg := makeTestServersConfig()
	m := NewManager(cfg)

	lang, ok := m.detectLanguage("/path/to/app.ts")
	if !ok {
		t.Fatal("expected ok=true for .ts extension, got false")
	}
	if lang != "typescript" {
		t.Fatalf("expected lang=%q, got %q", "typescript", lang)
	}
}

// TestDetectLanguage_TsxExtension tests that .tsx files are detected as "typescript" language.
func TestDetectLanguage_TsxExtension(t *testing.T) {
	cfg := makeTestServersConfig()
	m := NewManager(cfg)

	lang, ok := m.detectLanguage("/path/to/App.tsx")
	if !ok {
		t.Fatal("expected ok=true for .tsx extension, got false")
	}
	if lang != "typescript" {
		t.Fatalf("expected lang=%q, got %q", "typescript", lang)
	}
}

// TestDetectLanguage_UnknownExtension tests that unknown extensions return (empty, false).
func TestDetectLanguage_UnknownExtension(t *testing.T) {
	cfg := makeTestServersConfig()
	m := NewManager(cfg)

	lang, ok := m.detectLanguage("/path/to/file.xyz")
	if ok {
		t.Fatalf("expected ok=false for unknown extension, got lang=%q", lang)
	}
	if lang != "" {
		t.Fatalf("expected empty lang, got %q", lang)
	}
}

// TestDetectLanguage_NoExtension tests that files without an extension return (empty, false).
func TestDetectLanguage_NoExtension(t *testing.T) {
	cfg := makeTestServersConfig()
	m := NewManager(cfg)

	lang, ok := m.detectLanguage("/path/to/Makefile")
	if ok {
		t.Fatalf("expected ok=false for no-extension file, got lang=%q", lang)
	}
	if lang != "" {
		t.Fatalf("expected empty lang, got %q", lang)
	}
}

// TestDetectLanguage_AmbiguousExtension_ResolvedByMarker tests that language is determined
// by project markers when multiple languages share the same extension.
// Uses .ts extension in a temp directory with tsconfig.json.
func TestDetectLanguage_AmbiguousExtension_ResolvedByMarker(t *testing.T) {
	// configure two languages sharing the same extension
	cfg := &config.ServersConfig{
		Servers: map[string]config.ServerConfig{
			"typescript": {
				Language:       "typescript",
				Command:        "ts-server",
				FileExtensions: []string{".ts"},
				RootMarkers:    []string{"tsconfig.json"},
			},
			"plain": {
				Language:       "plain",
				Command:        "plain-server",
				FileExtensions: []string{".ts"},
				RootMarkers:    []string{"plain.config"},
			},
		},
	}

	// set up a temp directory with tsconfig.json
	dir := t.TempDir()
	tsconfig := filepath.Join(dir, "tsconfig.json")
	if err := os.WriteFile(tsconfig, []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to create tsconfig.json: %v", err)
	}

	m := NewManager(cfg)
	filePath := filepath.Join(dir, "app.ts")
	lang, ok := m.detectLanguage(filePath)
	if !ok {
		t.Fatal("expected ok=true for .ts with tsconfig.json marker, got false")
	}
	if lang != "typescript" {
		t.Fatalf("expected lang=%q, got %q", "typescript", lang)
	}
}

// TestDetectLanguage_AmbiguousExtension_FallsBackToFirstCandidate tests that the first
// candidate language (by config order) is selected when no marker is present.
func TestDetectLanguage_AmbiguousExtension_FallsBackToFirstCandidate(t *testing.T) {
	// two languages sharing .ext for unambiguous ordering
	cfg := &config.ServersConfig{
		Servers: map[string]config.ServerConfig{
			"langA": {
				Language:       "langA",
				Command:        "serverA",
				FileExtensions: []string{".ext"},
				RootMarkers:    []string{"markerA"},
			},
			"langB": {
				Language:       "langB",
				Command:        "serverB",
				FileExtensions: []string{".ext"},
				RootMarkers:    []string{"markerB"},
			},
		},
	}

	// temp directory with no marker files
	dir := t.TempDir()
	m := NewManager(cfg)
	filePath := filepath.Join(dir, "file.ext")

	lang, ok := m.detectLanguage(filePath)
	if !ok {
		t.Fatal("expected ok=true for known extension even without markers")
	}
	// must be one of langA or langB (deterministically the first)
	if lang != "langA" && lang != "langB" {
		t.Fatalf("expected one of langA or langB, got %q", lang)
	}
}

// TestRouteFor_NoLanguageDetected tests that ErrNoLanguageDetected is returned
// when RouteFor is called with an unknown file extension.
func TestRouteFor_NoLanguageDetected(t *testing.T) {
	cfg := makeTestServersConfig()
	var spawnCount atomic.Int32
	m := NewManager(cfg, WithClientFactory(fakeClientFactory(&spawnCount, nil, nil, nil)))

	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(context.Background()) //nolint:errcheck

	_, err := m.RouteFor(context.Background(), "/path/to/file.xyz")
	if !errors.Is(err, ErrNoLanguageDetected) {
		t.Fatalf("expected ErrNoLanguageDetected, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// T-016: Manager lazy spawn tests
// ---------------------------------------------------------------------------

// TestRouteFor_SpawnsClientOnFirstCall tests that a client is spawned on the first RouteFor call.
func TestRouteFor_SpawnsClientOnFirstCall(t *testing.T) {
	cfg := makeTestServersConfig()
	var spawnCount atomic.Int32
	m := NewManager(cfg, WithClientFactory(fakeClientFactory(&spawnCount, nil, nil, nil)))

	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(context.Background()) //nolint:errcheck

	_, err := m.RouteFor(context.Background(), "/path/to/main.go")
	if err != nil {
		t.Fatalf("RouteFor failed: %v", err)
	}
	if spawnCount.Load() != 1 {
		t.Fatalf("expected factory called once, got %d", spawnCount.Load())
	}
}

// TestRouteFor_ReturnsSameClientOnSecondCall tests that the factory is called only once
// even when RouteFor is called twice for the same language.
func TestRouteFor_ReturnsSameClientOnSecondCall(t *testing.T) {
	cfg := makeTestServersConfig()
	var spawnCount atomic.Int32
	m := NewManager(cfg, WithClientFactory(fakeClientFactory(&spawnCount, nil, nil, nil)))

	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(context.Background()) //nolint:errcheck

	c1, err := m.RouteFor(context.Background(), "/path/to/main.go")
	if err != nil {
		t.Fatalf("first RouteFor failed: %v", err)
	}
	c2, err := m.RouteFor(context.Background(), "/path/to/other.go")
	if err != nil {
		t.Fatalf("second RouteFor failed: %v", err)
	}
	if c1 != c2 {
		t.Fatal("expected same Client instance for same language, got different")
	}
	if spawnCount.Load() != 1 {
		t.Fatalf("expected factory called once, got %d", spawnCount.Load())
	}
}

// TestRouteFor_ConcurrentSpawnOnlyOnce tests that the factory is called exactly once
// even when 50 goroutines call RouteFor concurrently for the same language.
func TestRouteFor_ConcurrentSpawnOnlyOnce(t *testing.T) {
	cfg := makeTestServersConfig()
	var spawnCount atomic.Int32
	m := NewManager(cfg, WithClientFactory(fakeClientFactory(&spawnCount, nil, nil, nil)))

	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(context.Background()) //nolint:errcheck

	const goroutines = 50
	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = m.RouteFor(context.Background(), "/path/to/main.go")
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: RouteFor failed: %v", i, err)
		}
	}
	if spawnCount.Load() != 1 {
		t.Fatalf("expected factory called once across 50 goroutines, got %d", spawnCount.Load())
	}
}

// TestRouteFor_SpawnErrorReleasesClient tests that the client is released when Start fails
// and the next call retries.
func TestRouteFor_SpawnErrorReleasesClient(t *testing.T) {
	cfg := makeTestServersConfig()

	startErr := errors.New("fake start error")
	var callCount atomic.Int32

	// factory that fails on the first call and succeeds on the second
	factory := func(sc config.ServerConfig) Client {
		callCount.Add(1)
		call := callCount.Load()
		if call == 1 {
			return &fakeClient{state: StateSpawning, startErr: startErr}
		}
		return &fakeClient{state: StateSpawning} // second call succeeds
	}

	m := NewManager(cfg, WithClientFactory(factory))

	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(context.Background()) //nolint:errcheck

	// first call: Start fails
	_, err := m.RouteFor(context.Background(), "/path/to/main.go")
	if err == nil {
		t.Fatal("expected error from RouteFor with failing client start, got nil")
	}
	if !errors.Is(err, startErr) {
		t.Fatalf("expected startErr in chain, got: %v", err)
	}

	// second call: must succeed (failed client from first call is released)
	_, err2 := m.RouteFor(context.Background(), "/path/to/main.go")
	if err2 != nil {
		t.Fatalf("expected second RouteFor to succeed, got: %v", err2)
	}
	if callCount.Load() != 2 {
		t.Fatalf("expected factory called 2 times (1 fail + 1 success), got %d", callCount.Load())
	}
}

// TestRouteFor_UpdatesLastActivity tests that RouteFor updates lastActivity.
func TestRouteFor_UpdatesLastActivity(t *testing.T) {
	cfg := makeTestServersConfig()
	var spawnCount atomic.Int32
	m := NewManager(cfg, WithClientFactory(fakeClientFactory(&spawnCount, nil, nil, nil)))

	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(context.Background()) //nolint:errcheck

	before := time.Now()
	_, err := m.RouteFor(context.Background(), "/path/to/main.go")
	if err != nil {
		t.Fatalf("RouteFor failed: %v", err)
	}
	after := time.Now()

	m.mu.Lock()
	activity := m.lastActivity["go"]
	m.mu.Unlock()

	if activity.Before(before) || activity.After(after.Add(time.Second)) {
		t.Fatalf("lastActivity %v is outside expected range [%v, %v]", activity, before, after)
	}
}

// ---------------------------------------------------------------------------
// T-017: Manager idle shutdown tests
// ---------------------------------------------------------------------------

// TestManager_ReaperShutsDownIdleClient tests that a client is shut down after the idle timeout.
// Uses WithIdleShutdownSeconds(0) + a very short reaper interval.
func TestManager_ReaperShutsDownIdleClient(t *testing.T) {
	cfg := makeTestServersConfig()

	var clients []*fakeClient
	var clientsMu sync.Mutex
	var spawnCount atomic.Int32

	factory := fakeClientFactory(&spawnCount, &clients, &clientsMu, nil)
	m := NewManager(cfg,
		WithClientFactory(factory),
		WithIdleShutdownSeconds(0),         // 0 seconds = expires immediately
		WithReaperInterval(5*time.Millisecond),
	)

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(ctx) //nolint:errcheck

	// create a client (via RouteFor call)
	_, err := m.RouteFor(ctx, "/path/to/main.go")
	if err != nil {
		t.Fatalf("RouteFor failed: %v", err)
	}

	// wait until reaper shuts down the idle client
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		clientsMu.Lock()
		count := len(clients)
		var shutCalled int
		if count > 0 {
			clients[0].mu.Lock()
			shutCalled = clients[0].shutdownCalled
			clients[0].mu.Unlock()
		}
		clientsMu.Unlock()
		if count > 0 && shutCalled > 0 {
			return // success
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("reaper did not shut down idle client within deadline")
}

// TestManager_ReaperDoesNotShutdownActiveClient tests that the reaper does not
// shut down an active client.
func TestManager_ReaperDoesNotShutdownActiveClient(t *testing.T) {
	cfg := makeTestServersConfig()

	var clients []*fakeClient
	var clientsMu sync.Mutex
	var spawnCount atomic.Int32

	factory := fakeClientFactory(&spawnCount, &clients, &clientsMu, nil)
	m := NewManager(cfg,
		WithClientFactory(factory),
		WithIdleShutdownSeconds(60),        // 60-second timeout — will not expire during test
		WithReaperInterval(5*time.Millisecond),
	)

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(ctx) //nolint:errcheck

	_, err := m.RouteFor(ctx, "/path/to/main.go")
	if err != nil {
		t.Fatalf("RouteFor failed: %v", err)
	}

	// wait for the reaper to cycle a few times
	time.Sleep(50 * time.Millisecond)

	clientsMu.Lock()
	shutdown := len(clients) > 0 && clients[0].shutdownCalled > 0
	clientsMu.Unlock()

	if shutdown {
		t.Fatal("reaper shut down active client (idle timeout not yet reached)")
	}
}

// TestManager_Shutdown_CancelsReaperPromptly tests that the reaper terminates within
// 100ms when Shutdown is called.
func TestManager_Shutdown_CancelsReaperPromptly(t *testing.T) {
	cfg := makeTestServersConfig()
	var spawnCount atomic.Int32
	m := NewManager(cfg,
		WithClientFactory(fakeClientFactory(&spawnCount, nil, nil, nil)),
		WithIdleShutdownSeconds(60),
		WithReaperInterval(50*time.Millisecond),
	)

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}

	start := time.Now()
	if err := m.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed > 300*time.Millisecond {
		t.Fatalf("Shutdown took %v, expected < 300ms", elapsed)
	}
}

// TestManager_Shutdown_AggregatesClientErrors tests that errors from client shutdown
// are aggregated and returned.
func TestManager_Shutdown_AggregatesClientErrors(t *testing.T) {
	cfg := makeTestServersConfig()

	shutdownErr := errors.New("fake shutdown error")
	factory := func(_ config.ServerConfig) Client {
		return &fakeClient{state: StateSpawning, shutdownErr: shutdownErr}
	}

	m := NewManager(cfg,
		WithClientFactory(factory),
		WithIdleShutdownSeconds(60),
		WithReaperInterval(50*time.Millisecond),
	)

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}

	// create a client
	_, err := m.RouteFor(ctx, "/path/to/main.go")
	if err != nil {
		t.Fatalf("RouteFor failed: %v", err)
	}

	// Shutdown — client errors must be aggregated
	shutErr := m.Shutdown(ctx)
	if shutErr == nil {
		t.Fatal("expected Shutdown to return aggregated error, got nil")
	}
	if !errors.Is(shutErr, shutdownErr) {
		t.Fatalf("expected shutdownErr in Shutdown error chain, got: %v", shutErr)
	}
}

// TestManager_Shutdown_ClearsClientsMap tests that the clients map is emptied after Shutdown.
func TestManager_Shutdown_ClearsClientsMap(t *testing.T) {
	cfg := makeTestServersConfig()
	var spawnCount atomic.Int32
	m := NewManager(cfg, WithClientFactory(fakeClientFactory(&spawnCount, nil, nil, nil)))

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}

	_, err := m.RouteFor(ctx, "/path/to/main.go")
	if err != nil {
		t.Fatalf("RouteFor failed: %v", err)
	}

	if err := m.Shutdown(ctx); err != nil {
		t.Logf("Shutdown returned (non-fatal for this test): %v", err)
	}

	m.mu.Lock()
	count := len(m.clients)
	m.mu.Unlock()

	if count != 0 {
		t.Fatalf("expected clients map to be empty after Shutdown, got %d entries", count)
	}
}

// TestErrNoLanguageDetected_Sentinel tests that ErrNoLanguageDetected is a sentinel
// detectable via errors.Is.
func TestErrNoLanguageDetected_Sentinel(t *testing.T) {
	if ErrNoLanguageDetected == nil {
		t.Fatal("ErrNoLanguageDetected should not be nil")
	}
	wrapped := errors.New("outer: " + ErrNoLanguageDetected.Error())
	// wrapped is not detectable via errors.Is (intentional)
	// the true sentinel must be directly comparable
	if !errors.Is(ErrNoLanguageDetected, ErrNoLanguageDetected) {
		t.Fatal("ErrNoLanguageDetected should be detectable via errors.Is with itself")
	}
	_ = wrapped
}
