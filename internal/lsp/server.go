package lsp

import (
	"context"
	"fmt"
	"maps"
	"sort"
	"sync"
	"time"
)

// ServerLauncher abstracts launching a language server process.
// Implementations handle process creation, transport setup, and LSP initialization.
//
// Example implementation:
//
//	type stdioLauncher struct{}
//
//	func (l *stdioLauncher) Launch(ctx context.Context, lang string) (Client, ProcessHandle, error) {
//	    cmd := exec.CommandContext(ctx, serverPath, args...)
//	    stdin, _ := cmd.StdinPipe()
//	    stdout, _ := cmd.StdoutPipe()
//	    cmd.Start()
//	    transport := NewStreamTransport(stdout, stdin, cmd)
//	    conn := NewConn(transport)
//	    client := NewClient(conn)
//	    client.Initialize(ctx, rootURI)
//	    return client, &processHandle{cmd: cmd}, nil
//	}
type ServerLauncher interface {
	// Launch starts a language server for the given language and returns
	// the initialized client and a handle to the server process.
	Launch(ctx context.Context, lang string) (Client, ProcessHandle, error)
}

// ProcessHandle abstracts a language server process for lifecycle management.
type ProcessHandle interface {
	// Kill forcefully terminates the server process.
	Kill() error

	// Wait blocks until the server process exits.
	Wait() error

	// IsRunning reports whether the server process is still running.
	IsRunning() bool
}

// @MX:ANCHOR: [AUTO] ServerManager는 다중 언어 서버의 동시성 라이프사이클을 관리하는 핵심 인터페이스입니다. 모든 메서드는 스레드 안전하게 설계되었습니다.
// @MX:REASON: fan_in=10+, LSP 서버 관리의 진입점이며 여러 곳에서 호출됩니다
// ServerManager manages multiple language server lifecycles concurrently.
// All methods are safe for concurrent use.
//
// Example usage:
//
//	mgr := lsp.NewServerManager(launcher, lsp.WithMaxParallel(4))
//	err := mgr.StartAll(ctx, []string{"go", "python", "typescript"})
//	diags, err := mgr.CollectAllDiagnostics(ctx, "file:///project/main.go")
//	err = mgr.StopAll(ctx)
type ServerManager interface {
	// StartServer starts a language server for the given language.
	// Returns nil if the server is already running (idempotent).
	StartServer(ctx context.Context, lang string) error

	// StopServer stops the language server for the given language.
	// Returns nil if the server is not running (idempotent).
	StopServer(ctx context.Context, lang string) error

	// GetClient returns the LSP client for the given language.
	// Returns ErrServerNotRunning if the server is not running.
	GetClient(lang string) (Client, error)

	// ActiveServers returns a sorted list of languages with running servers.
	ActiveServers() []string

	// HealthCheck checks the health of all active servers.
	// Returns a map from language to error (nil means healthy).
	HealthCheck(ctx context.Context) map[string]error

	// StartAll starts servers for all given languages concurrently.
	// Individual failures are non-fatal (graceful degradation).
	StartAll(ctx context.Context, langs []string) error

	// StopAll stops all running servers.
	StopAll(ctx context.Context) error

	// CollectAllDiagnostics collects diagnostics from all active servers concurrently.
	// Individual server errors are non-fatal.
	CollectAllDiagnostics(ctx context.Context, uri string) ([]Diagnostic, error)
}

// ManagerOption configures a serverManager.
type ManagerOption func(*serverManager)

// WithMaxParallel sets the maximum number of concurrent server startups.
// The default is 4.
func WithMaxParallel(n int) ManagerOption {
	return func(m *serverManager) {
		if n > 0 {
			m.maxParallel = n
		}
	}
}

const defaultMaxParallel = 4

// managedServer holds a running language server and its process handle.
type managedServer struct {
	client    Client
	process   ProcessHandle
	lang      string
	startedAt time.Time
}

// serverManager implements ServerManager.
type serverManager struct {
	servers     map[string]*managedServer
	mu          sync.RWMutex
	launcher    ServerLauncher
	maxParallel int
}

// NewServerManager creates a new ServerManager with the given launcher and options.
func NewServerManager(launcher ServerLauncher, opts ...ManagerOption) ServerManager {
	m := &serverManager{
		servers:     make(map[string]*managedServer),
		launcher:    launcher,
		maxParallel: defaultMaxParallel,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// StartServer starts a language server for the given language.
// If the server is already running, it returns nil (idempotent).
func (m *serverManager) StartServer(ctx context.Context, lang string) error {
	// Fast path: check if already running.
	if m.isRunning(lang) {
		return nil
	}

	client, process, err := m.launcher.Launch(ctx, lang)
	if err != nil {
		return fmt.Errorf("start %s server: %w", lang, err)
	}

	registered := func() bool {
		m.mu.Lock()
		defer m.mu.Unlock()

		// Double-check after acquiring write lock to handle concurrent starts.
		if _, exists := m.servers[lang]; exists {
			return false
		}

		m.servers[lang] = &managedServer{
			client:    client,
			process:   process,
			lang:      lang,
			startedAt: time.Now(),
		}
		return true
	}()

	if !registered {
		// Another goroutine started it. Clean up our duplicate.
		_ = client.Shutdown(ctx)
		_ = process.Kill()
		return nil
	}

	// Start background goroutine to detect process crashes.
	go m.watchProcess(lang, process)

	return nil
}

// StopServer stops the language server for the given language.
// If the server is not running, it returns nil (idempotent).
func (m *serverManager) StopServer(ctx context.Context, lang string) error {
	srv := func() *managedServer {
		m.mu.Lock()
		defer m.mu.Unlock()

		srv, exists := m.servers[lang]
		if !exists {
			return nil
		}
		delete(m.servers, lang)
		return srv
	}()

	if srv == nil {
		return nil
	}

	// Attempt graceful shutdown.
	if err := srv.client.Shutdown(ctx); err != nil {
		// Graceful shutdown failed; force kill.
		_ = srv.process.Kill()
	}

	return nil
}

// GetClient returns the LSP client for the given language.
// Returns ErrServerNotRunning if the server is not running.
func (m *serverManager) GetClient(lang string) (Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	srv, exists := m.servers[lang]
	if !exists {
		return nil, fmt.Errorf("get client for %s: %w", lang, ErrServerNotRunning)
	}
	return srv.client, nil
}

// ActiveServers returns a sorted list of languages with running servers.
func (m *serverManager) ActiveServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	langs := make([]string, 0, len(m.servers))
	for lang := range m.servers {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}

// HealthCheck checks the health of all active servers by verifying
// their processes are still running.
func (m *serverManager) HealthCheck(_ context.Context) map[string]error {
	snapshot := func() map[string]*managedServer {
		m.mu.RLock()
		defer m.mu.RUnlock()

		s := make(map[string]*managedServer, len(m.servers))
		maps.Copy(s, m.servers)
		return s
	}()

	result := make(map[string]error, len(snapshot))
	for lang, srv := range snapshot {
		if !srv.process.IsRunning() {
			result[lang] = fmt.Errorf("%s server: process not running", lang)
		} else {
			result[lang] = nil
		}
	}
	return result
}

// @MX:WARN: [AUTO] 고루틴 풀을 사용하여 병렬로 서버를 시작합니다. 셈포어는 maxParallel로 제한됩니다.
// @MX:REASON: [AUTO] 고루틴 수가 langs 길이만큼 생성되어 리소스 부하 가능성
// @MX:NOTE: [AUTO] 개별 서버 실패는 치명적이지 않습니다 (graceful degradation).
// StartAll starts servers for all given languages concurrently,
// limited by maxParallel. Individual failures are non-fatal (graceful degradation).
func (m *serverManager) StartAll(ctx context.Context, langs []string) error {
	sem := make(chan struct{}, m.maxParallel)
	var wg sync.WaitGroup

	for _, lang := range langs {
		wg.Add(1)
		go func(l string) {
			defer wg.Done()
			sem <- struct{}{}        // acquire semaphore
			defer func() { <-sem }() // release semaphore

			// Individual server start errors are non-fatal.
			_ = m.StartServer(ctx, l)
		}(lang)
	}

	wg.Wait()
	return nil
}

// StopAll stops all running servers.
func (m *serverManager) StopAll(ctx context.Context) error {
	langs := m.ActiveServers()
	for _, lang := range langs {
		_ = m.StopServer(ctx, lang)
	}
	return nil
}

// @MX:WARN: [AUTO] 고루틴을 사용하여 모든 활성 서버에서 병렬로 진단을 수집합니다.
// @MX:REASON: [AUTO] 서버 수만큼 고루틴이 생성되어 리소스 부하 가능성
// CollectAllDiagnostics collects diagnostics from all active servers concurrently.
// Individual server errors are non-fatal; results from successful servers are returned.
func (m *serverManager) CollectAllDiagnostics(ctx context.Context, uri string) ([]Diagnostic, error) {
	snapshot := func() []*managedServer {
		m.mu.RLock()
		defer m.mu.RUnlock()

		servers := make([]*managedServer, 0, len(m.servers))
		for _, srv := range m.servers {
			servers = append(servers, srv)
		}
		return servers
	}()

	if len(snapshot) == 0 {
		return []Diagnostic{}, nil
	}

	var mu sync.Mutex
	var allDiags []Diagnostic
	var wg sync.WaitGroup

	for _, srv := range snapshot {
		wg.Add(1)
		go func(s *managedServer) {
			defer wg.Done()

			diags, err := s.client.Diagnostics(ctx, uri)
			if err != nil {
				// Non-fatal: skip this server's diagnostics.
				return
			}

			func() {
				mu.Lock()
				defer mu.Unlock()
				allDiags = append(allDiags, diags...)
			}()
		}(srv)
	}

	wg.Wait()

	if allDiags == nil {
		return []Diagnostic{}, nil
	}
	return allDiags, nil
}

// isRunning checks if a server for the given language is currently running.
func (m *serverManager) isRunning(lang string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.servers[lang]
	return exists
}

// @MX:WARN: [AUTO] 고루틴에서 프로세스 종료를 차단하고 있습니다. 컨텍스트 취소에 응답하지 않습니다.
// @MX:REASON: [AUTO] 고루틴이 process.Wait()에서 영구 차단될 수 있습니다
// watchProcess monitors a server process and removes it from the registry
// when the process exits unexpectedly (crash detection).
func (m *serverManager) watchProcess(lang string, process ProcessHandle) {
	_ = process.Wait() // blocks until process exits

	m.mu.Lock()
	defer m.mu.Unlock()

	// Only remove if it's the same server instance (not replaced by a new start).
	if srv, ok := m.servers[lang]; ok && srv.process == process {
		delete(m.servers, lang)
	}
}
