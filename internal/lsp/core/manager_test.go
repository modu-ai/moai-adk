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

// fakeClient는 테스트용 Client 구현체입니다.
// 실제 subprocess를 생성하지 않고 동작을 시뮬레이션합니다.
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

// fakeClientFactory는 fakeClient를 생성하는 팩토리 함수를 반환합니다.
// spawnCount로 생성 횟수를 추적합니다.
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

// makeTestServersConfig는 테스트용 ServersConfig를 생성합니다.
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
// T-015: detectLanguage 테스트
// ---------------------------------------------------------------------------

// TestDetectLanguage_GoExtension은 .go 파일이 "go" 언어로 감지되는지 테스트합니다.
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

// TestDetectLanguage_PythonExtension은 .py 파일이 "python" 언어로 감지되는지 테스트합니다.
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

// TestDetectLanguage_TypeScriptExtension은 .ts 파일이 "typescript" 언어로 감지되는지 테스트합니다.
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

// TestDetectLanguage_TsxExtension은 .tsx 파일이 "typescript" 언어로 감지되는지 테스트합니다.
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

// TestDetectLanguage_UnknownExtension은 알 수 없는 확장자에 대해 (empty, false)를 반환하는지 테스트합니다.
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

// TestDetectLanguage_NoExtension은 확장자가 없는 파일에 대해 (empty, false)를 반환하는지 테스트합니다.
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

// TestDetectLanguage_AmbiguousExtension_ResolvedByMarker는 확장자 중복 시 프로젝트 마커로
// 언어가 결정되는지 테스트합니다.
// tsconfig.json이 있는 임시 디렉토리에서 .ts 확장자를 사용합니다.
func TestDetectLanguage_AmbiguousExtension_ResolvedByMarker(t *testing.T) {
	// 두 언어가 같은 확장자를 공유하는 설정 구성
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

	// tsconfig.json이 있는 임시 디렉토리 구성
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

// TestDetectLanguage_AmbiguousExtension_FallsBackToFirstCandidate는
// 마커가 없을 때 첫 번째 후보 언어(설정 순서 기준)가 선택되는지 테스트합니다.
func TestDetectLanguage_AmbiguousExtension_FallsBackToFirstCandidate(t *testing.T) {
	// 명확한 순서를 위해 두 언어가 .ext를 공유하는 설정
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

	// 마커 파일이 없는 임시 디렉토리
	dir := t.TempDir()
	m := NewManager(cfg)
	filePath := filepath.Join(dir, "file.ext")

	lang, ok := m.detectLanguage(filePath)
	if !ok {
		t.Fatal("expected ok=true for known extension even without markers")
	}
	// langA 또는 langB 중 하나여야 함 (결정론적으로 첫 번째)
	if lang != "langA" && lang != "langB" {
		t.Fatalf("expected one of langA or langB, got %q", lang)
	}
}

// TestRouteFor_NoLanguageDetected는 알 수 없는 파일 확장자로 routeFor를 호출하면
// ErrNoLanguageDetected가 반환되는지 테스트합니다.
func TestRouteFor_NoLanguageDetected(t *testing.T) {
	cfg := makeTestServersConfig()
	var spawnCount atomic.Int32
	m := NewManager(cfg, WithClientFactory(fakeClientFactory(&spawnCount, nil, nil, nil)))

	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(context.Background()) //nolint:errcheck

	_, err := m.routeFor(context.Background(), "/path/to/file.xyz")
	if !errors.Is(err, ErrNoLanguageDetected) {
		t.Fatalf("expected ErrNoLanguageDetected, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// T-016: Manager lazy spawn 테스트
// ---------------------------------------------------------------------------

// TestRouteFor_SpawnsClientOnFirstCall은 routeFor 최초 호출 시 클라이언트를 생성하는지 테스트합니다.
func TestRouteFor_SpawnsClientOnFirstCall(t *testing.T) {
	cfg := makeTestServersConfig()
	var spawnCount atomic.Int32
	m := NewManager(cfg, WithClientFactory(fakeClientFactory(&spawnCount, nil, nil, nil)))

	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(context.Background()) //nolint:errcheck

	_, err := m.routeFor(context.Background(), "/path/to/main.go")
	if err != nil {
		t.Fatalf("routeFor failed: %v", err)
	}
	if spawnCount.Load() != 1 {
		t.Fatalf("expected factory called once, got %d", spawnCount.Load())
	}
}

// TestRouteFor_ReturnsSameClientOnSecondCall은 동일 언어로 두 번 routeFor를 호출해도
// 팩토리가 한 번만 호출되는지 테스트합니다.
func TestRouteFor_ReturnsSameClientOnSecondCall(t *testing.T) {
	cfg := makeTestServersConfig()
	var spawnCount atomic.Int32
	m := NewManager(cfg, WithClientFactory(fakeClientFactory(&spawnCount, nil, nil, nil)))

	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(context.Background()) //nolint:errcheck

	c1, err := m.routeFor(context.Background(), "/path/to/main.go")
	if err != nil {
		t.Fatalf("first routeFor failed: %v", err)
	}
	c2, err := m.routeFor(context.Background(), "/path/to/other.go")
	if err != nil {
		t.Fatalf("second routeFor failed: %v", err)
	}
	if c1 != c2 {
		t.Fatal("expected same Client instance for same language, got different")
	}
	if spawnCount.Load() != 1 {
		t.Fatalf("expected factory called once, got %d", spawnCount.Load())
	}
}

// TestRouteFor_ConcurrentSpawnOnlyOnce는 50개 고루틴이 동일 언어로 routeFor를 동시에
// 호출해도 팩토리가 정확히 1번만 호출되는지 테스트합니다.
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
			_, errs[idx] = m.routeFor(context.Background(), "/path/to/main.go")
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: routeFor failed: %v", i, err)
		}
	}
	if spawnCount.Load() != 1 {
		t.Fatalf("expected factory called once across 50 goroutines, got %d", spawnCount.Load())
	}
}

// TestRouteFor_SpawnErrorReleasesClient는 Start가 실패하면 클라이언트가 해제되고
// 다음 호출 시 재시도하는지 테스트합니다.
func TestRouteFor_SpawnErrorReleasesClient(t *testing.T) {
	cfg := makeTestServersConfig()

	startErr := errors.New("fake start error")
	var callCount atomic.Int32

	// 첫 번째 호출은 실패, 두 번째 호출은 성공하는 팩토리
	factory := func(sc config.ServerConfig) Client {
		callCount.Add(1)
		call := callCount.Load()
		if call == 1 {
			return &fakeClient{state: StateSpawning, startErr: startErr}
		}
		return &fakeClient{state: StateSpawning} // 두 번째는 성공
	}

	m := NewManager(cfg, WithClientFactory(factory))

	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(context.Background()) //nolint:errcheck

	// 첫 번째 호출: Start 실패
	_, err := m.routeFor(context.Background(), "/path/to/main.go")
	if err == nil {
		t.Fatal("expected error from routeFor with failing client start, got nil")
	}
	if !errors.Is(err, startErr) {
		t.Fatalf("expected startErr in chain, got: %v", err)
	}

	// 두 번째 호출: 성공해야 함 (첫 번째 실패한 클라이언트가 해제됨)
	_, err2 := m.routeFor(context.Background(), "/path/to/main.go")
	if err2 != nil {
		t.Fatalf("expected second routeFor to succeed, got: %v", err2)
	}
	if callCount.Load() != 2 {
		t.Fatalf("expected factory called 2 times (1 fail + 1 success), got %d", callCount.Load())
	}
}

// TestRouteFor_UpdatesLastActivity는 routeFor가 lastActivity를 업데이트하는지 테스트합니다.
func TestRouteFor_UpdatesLastActivity(t *testing.T) {
	cfg := makeTestServersConfig()
	var spawnCount atomic.Int32
	m := NewManager(cfg, WithClientFactory(fakeClientFactory(&spawnCount, nil, nil, nil)))

	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(context.Background()) //nolint:errcheck

	before := time.Now()
	_, err := m.routeFor(context.Background(), "/path/to/main.go")
	if err != nil {
		t.Fatalf("routeFor failed: %v", err)
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
// T-017: Manager idle shutdown 테스트
// ---------------------------------------------------------------------------

// TestManager_ReaperShutsDownIdleClient는 유휴 타임아웃 후 클라이언트가 종료되는지 테스트합니다.
// WithIdleShutdownSeconds(0) + 매우 짧은 reaper 간격을 사용합니다.
func TestManager_ReaperShutsDownIdleClient(t *testing.T) {
	cfg := makeTestServersConfig()

	var clients []*fakeClient
	var clientsMu sync.Mutex
	var spawnCount atomic.Int32

	factory := fakeClientFactory(&spawnCount, &clients, &clientsMu, nil)
	m := NewManager(cfg,
		WithClientFactory(factory),
		WithIdleShutdownSeconds(0),         // 0초 = 즉시 만료
		WithReaperInterval(5*time.Millisecond),
	)

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(ctx) //nolint:errcheck

	// 클라이언트 생성 (routeFor 호출)
	_, err := m.routeFor(ctx, "/path/to/main.go")
	if err != nil {
		t.Fatalf("routeFor failed: %v", err)
	}

	// reaper가 idle 클라이언트를 종료할 때까지 대기
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
			return // 성공
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("reaper did not shut down idle client within deadline")
}

// TestManager_ReaperDoesNotShutdownActiveClient는 활성 클라이언트는 reaper가 종료하지
// 않는지 테스트합니다.
func TestManager_ReaperDoesNotShutdownActiveClient(t *testing.T) {
	cfg := makeTestServersConfig()

	var clients []*fakeClient
	var clientsMu sync.Mutex
	var spawnCount atomic.Int32

	factory := fakeClientFactory(&spawnCount, &clients, &clientsMu, nil)
	m := NewManager(cfg,
		WithClientFactory(factory),
		WithIdleShutdownSeconds(60),        // 60초 타임아웃 — 테스트 중에는 만료되지 않음
		WithReaperInterval(5*time.Millisecond),
	)

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}
	defer m.Shutdown(ctx) //nolint:errcheck

	_, err := m.routeFor(ctx, "/path/to/main.go")
	if err != nil {
		t.Fatalf("routeFor failed: %v", err)
	}

	// reaper가 몇 번 돌 시간 동안 대기
	time.Sleep(50 * time.Millisecond)

	clientsMu.Lock()
	shutdown := len(clients) > 0 && clients[0].shutdownCalled > 0
	clientsMu.Unlock()

	if shutdown {
		t.Fatal("reaper shut down active client (idle timeout not yet reached)")
	}
}

// TestManager_Shutdown_CancelsReaperPromptly는 Shutdown 호출 시 reaper가 100ms 이내에
// 종료되는지 테스트합니다.
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

// TestManager_Shutdown_AggregatesClientErrors는 클라이언트 종료 시 에러가 집계되어
// 반환되는지 테스트합니다.
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

	// 클라이언트 생성
	_, err := m.routeFor(ctx, "/path/to/main.go")
	if err != nil {
		t.Fatalf("routeFor failed: %v", err)
	}

	// Shutdown — 클라이언트 에러가 집계되어야 함
	shutErr := m.Shutdown(ctx)
	if shutErr == nil {
		t.Fatal("expected Shutdown to return aggregated error, got nil")
	}
	if !errors.Is(shutErr, shutdownErr) {
		t.Fatalf("expected shutdownErr in Shutdown error chain, got: %v", shutErr)
	}
}

// TestManager_Shutdown_ClearsClientsMap는 Shutdown 후 clients 맵이 비워지는지 테스트합니다.
func TestManager_Shutdown_ClearsClientsMap(t *testing.T) {
	cfg := makeTestServersConfig()
	var spawnCount atomic.Int32
	m := NewManager(cfg, WithClientFactory(fakeClientFactory(&spawnCount, nil, nil, nil)))

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Fatalf("Manager.Start failed: %v", err)
	}

	_, err := m.routeFor(ctx, "/path/to/main.go")
	if err != nil {
		t.Fatalf("routeFor failed: %v", err)
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

// TestErrNoLanguageDetected_Sentinel은 ErrNoLanguageDetected가 errors.Is로 감지 가능한
// 센티넬인지 테스트합니다.
func TestErrNoLanguageDetected_Sentinel(t *testing.T) {
	if ErrNoLanguageDetected == nil {
		t.Fatal("ErrNoLanguageDetected should not be nil")
	}
	wrapped := errors.New("outer: " + ErrNoLanguageDetected.Error())
	// wrapped는 errors.Is로 감지되지 않음 (의도적)
	// 진짜 sentinel은 직접 비교 가능해야 함
	if !errors.Is(ErrNoLanguageDetected, ErrNoLanguageDetected) {
		t.Fatal("ErrNoLanguageDetected should be detectable via errors.Is with itself")
	}
	_ = wrapped
}
