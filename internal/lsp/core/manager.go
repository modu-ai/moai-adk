package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/modu-ai/moai-adk/internal/lsp/config"
)

// ErrNoLanguageDetected is the sentinel error returned by RouteFor when the file
// extension does not map to any configured language server.
//
// @MX:ANCHOR: [AUTO] ErrNoLanguageDetected — public sentinel error for Manager.RouteFor
// @MX:REASON: fan_in >= 3 — RouteFor, test assertions, Ralph engine, MCP bridge, and other callers all branch on this sentinel
var ErrNoLanguageDetected = errors.New("lsp: no language server configured for this file type")

// Manager coordinates multiple language server Client instances.
// Routes requests to the appropriate Client based on file extension and project markers (REQ-LC-008).
//
// @MX:ANCHOR: [AUTO] Manager — multi-language LSP client coordinator (REQ-LC-008, REQ-LC-009, REQ-LC-050)
// @MX:REASON: fan_in >= 3 — Ralph engine, Quality Gates, LOOP command, and MCP bridge all access LSP through Manager
type Manager struct {
	// servers is a map of language name (e.g. "go") → ServerConfig.
	servers map[string]config.ServerConfig

	// clients is a map of language name → Client. Clients are lazy-spawned (T-016).
	clients map[string]Client

	// mu protects access to servers, clients, and lastActivity.
	mu sync.Mutex

	// sf is a singleflight group that serializes concurrent getOrSpawn calls per language (REQ-UTIL-003-003).
	// Initialized as a zero-value and ready to use without an additional constructor call.
	// @MX:NOTE: [AUTO] sf — singleflight.Group; prevents duplicate clientFactory+Start calls for the same language (REQ-UTIL-003-004, REQ-UTIL-003-005)
	sf singleflight.Group

	// clientFactory is the Client construction function; can be replaced via DI in tests.
	// Default: NewClient.
	clientFactory func(config.ServerConfig) Client

	// lastActivity tracks the last RouteFor call time per language (REQ-LC-050).
	lastActivity map[string]time.Time

	// idleShutdownSeconds is the inactivity threshold in seconds for client shutdown.
	// A value of 0 is treated as immediately expired.
	idleShutdownSeconds int

	// reaperInterval is the tick interval of the reaper goroutine (can be shortened via WithReaperInterval in tests).
	reaperInterval time.Duration

	// logger is used for state transitions and lifecycle events.
	logger *slog.Logger

	// ctx and cancel control the lifecycle of Manager's internal goroutine (reaper).
	ctx    context.Context
	cancel context.CancelFunc

	// reaperDone is a channel that signals when the reaper goroutine has finished.
	reaperDone chan struct{}

	// statFn is the function used to check file existence (can be injected via DI in tests).
	// Default: os.Stat.
	statFn func(string) (os.FileInfo, error)
}

// ManagerOption is a functional option for configuring a Manager.
type ManagerOption func(*Manager)

// WithClientFactory is an option for injecting a fake Client via DI in tests.
func WithClientFactory(f func(config.ServerConfig) Client) ManagerOption {
	return func(m *Manager) {
		m.clientFactory = f
	}
}

// WithIdleShutdownSeconds sets the idle shutdown timeout for the Manager.
// The default is 600 seconds per REQ-LC-050.
func WithIdleShutdownSeconds(seconds int) ManagerOption {
	return func(m *Manager) {
		m.idleShutdownSeconds = seconds
	}
}

// WithManagerLogger sets the logger for the Manager.
func WithManagerLogger(logger *slog.Logger) ManagerOption {
	return func(m *Manager) {
		m.logger = logger
	}
}

// WithReaperInterval sets the tick interval of the reaper goroutine.
// Used in tests to shorten the interval for fast idle detection.
// Default: 30 seconds.
func WithReaperInterval(d time.Duration) ManagerOption {
	return func(m *Manager) {
		m.reaperInterval = d
	}
}

// defaultReaperInterval is the default tick interval for the reaper.
const defaultReaperInterval = 30 * time.Second

// defaultIdleShutdownSeconds is the default idle shutdown timeout per REQ-LC-050.
const defaultIdleShutdownSeconds = 600

// NewManager creates a new Manager from the given ServersConfig.
// The reaper goroutine is not started until Start is called.
//
// @MX:ANCHOR: [AUTO] NewManager — Manager factory and sole entry point for constructing a Manager
// @MX:REASON: fan_in >= 3 — CLI, integration tests, Ralph engine, and MCP bridge all call NewManager
func NewManager(cfg *config.ServersConfig, opts ...ManagerOption) *Manager {
	m := &Manager{
		servers:             make(map[string]config.ServerConfig),
		clients:             make(map[string]Client),
		lastActivity:        make(map[string]time.Time),
		idleShutdownSeconds: defaultIdleShutdownSeconds,
		reaperInterval:      defaultReaperInterval,
		logger:              slog.Default(),
	}

	// Initialize the servers map from ServersConfig.
	if cfg != nil {
		for lang, sc := range cfg.Servers {
			// Fill Language field from the map key if it is empty.
			if sc.Language == "" {
				sc.Language = lang
			}
			m.servers[lang] = sc
		}
	}

	// Default clientFactory: use NewClient.
	m.clientFactory = func(sc config.ServerConfig) Client {
		return NewClient(sc)
	}
	// Default statFn: use os.Stat.
	m.statFn = os.Stat

	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Start initializes the Manager's internal context and starts the reaper goroutine.
// This method must be called before using the Manager.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		// Already started — no-op.
		return nil
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.reaperDone = make(chan struct{})

	// @MX:WARN: [AUTO] reaper goroutine — long-running background goroutine that cleans up idle clients during the Manager's lifetime
	// @MX:REASON: goroutine leak may occur without ctx cancellation. Shutdown must cancel ctx and wait for reaperDone.
	go m.reaper(m.ctx, m.reaperDone)
	return nil
}

// Shutdown cancels the reaper goroutine and shuts down all cached Clients in parallel.
// Client shutdown errors are aggregated via errors.Join (REQ-LC-050).
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	if m.cancel != nil {
		m.cancel()
	}
	reaperDone := m.reaperDone
	m.mu.Unlock()

	// Wait for the reaper goroutine to finish (respecting ctx timeout).
	if reaperDone != nil {
		select {
		case <-reaperDone:
		case <-ctx.Done():
			// Continue even if ctx expires (best-effort).
		}
	}

	// Shut down all clients in parallel.
	m.mu.Lock()
	snapshot := make(map[string]Client, len(m.clients))
	for lang, c := range m.clients {
		snapshot[lang] = c
	}
	m.clients = make(map[string]Client)
	m.mu.Unlock()

	var shutdownErrs []error
	var wg sync.WaitGroup
	var errsMu sync.Mutex

	for lang, c := range snapshot {
		wg.Add(1)
		go func(language string, cl Client) {
			defer wg.Done()
			if err := cl.Shutdown(ctx); err != nil {
				errsMu.Lock()
				shutdownErrs = append(shutdownErrs, fmt.Errorf("language %s: %w", language, err))
				errsMu.Unlock()
			}
		}(lang, c)
	}
	wg.Wait()

	if len(shutdownErrs) > 0 {
		return errors.Join(shutdownErrs...)
	}
	return nil
}

// detectLanguage maps the file path extension to a configured language server.
//
// Mapping strategy:
//  1. If the extension maps to exactly one language, return it immediately.
//  2. If multiple languages share the same extension, disambiguate using project markers:
//     - Walk up from the file's directory to the root, checking each language's RootMarker files.
//     - If exactly one language has a matching marker, return that language.
//     - If markers cannot disambiguate, return the first candidate deterministically.
//  3. If no extension mapping exists, return ("", false).
func (m *Manager) detectLanguage(path string) (string, bool) {
	ext := filepath.Ext(path)
	if ext == "" {
		return "", false
	}

	// Collect candidate languages by extension.
	// Sort language names for deterministic ordering.
	var candidates []string
	for lang, sc := range m.servers {
		for _, fe := range sc.FileExtensions {
			if fe == ext {
				candidates = append(candidates, lang)
				break
			}
		}
	}

	if len(candidates) == 0 {
		return "", false
	}
	if len(candidates) == 1 {
		return candidates[0], true
	}

	// Sort for deterministic ordering.
	sort.Strings(candidates)

	// Disambiguate using project markers.
	dir := filepath.Dir(path)
	matched := findLanguageByMarkers(dir, candidates, m.servers, m.statFn)
	if matched != "" {
		return matched, true
	}

	// Cannot disambiguate via markers: return the first candidate (deterministic).
	return candidates[0], true
}

// findLanguageByMarkers walks up from dir to the filesystem root, checking whether
// each candidate language's RootMarker files exist.
// Returns the matched language if exactly one language has a marker, otherwise an empty string.
func findLanguageByMarkers(dir string, candidates []string, servers map[string]config.ServerConfig, stat func(string) (os.FileInfo, error)) string {
	// Walk up the path searching for markers.
	for {
		var matchedLangs []string
		for _, lang := range candidates {
			sc := servers[lang]
			for _, marker := range sc.RootMarkers {
				markerPath := filepath.Join(dir, marker)
				if _, statErr := stat(markerPath); statErr == nil {
					matchedLangs = append(matchedLangs, lang)
					break
				}
			}
		}

		if len(matchedLangs) == 1 {
			return matchedLangs[0]
		}

		// Move to the parent directory.
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the filesystem root.
			break
		}
		dir = parent
	}
	return ""
}

// RouteFor returns the LSP Client responsible for the given file path.
//
// Behavior:
//  1. Determine language via detectLanguage.
//  2. Obtain Client via getOrSpawn (lazy spawn).
//  3. Update lastActivity.
//
// @MX:ANCHOR: [AUTO] Manager.RouteFor — core path for file-path-based LSP client routing
// @MX:REASON: fan_in >= 3 — Ralph engine, Quality Gates, LOOP command, and Aggregator all obtain a Client through RouteFor
func (m *Manager) RouteFor(ctx context.Context, path string) (Client, error) {
	lang, ok := m.detectLanguage(path)
	if !ok {
		return nil, fmt.Errorf("lsp manager: %w (path=%q)", ErrNoLanguageDetected, path)
	}

	c, err := m.getOrSpawn(ctx, lang)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.lastActivity[lang] = time.Now()
	m.mu.Unlock()

	return c, nil
}

// getOrSpawn returns a cached Client or creates a new one if none exists.
//
// Concurrency safety (REQ-UTIL-003-004, REQ-UTIL-003-005, REQ-UTIL-003-006):
//  1. Fast path: check the cache under the mu lock and return immediately if a non-shutdown client exists.
//  2. Cache miss: enter the singleflight barrier (m.sf.Do) to serialize concurrent calls for the same language.
//     - Double-check inside sf.Do: another goroutine may have already completed.
//     - clientFactory + c.Start run only inside sf.Do → guaranteed exactly once per language.
//     - Insert into cache only after a successful Start → allows retry on Start failure (REQ-UTIL-003-006).
func (m *Manager) getOrSpawn(ctx context.Context, language string) (Client, error) {
	// Fast path: return immediately if a ready client is in the cache (lock → unlock → return).
	m.mu.Lock()
	existing, ok := m.clients[language]
	if ok && existing.State() != StateShutdown {
		m.mu.Unlock()
		return existing, nil
	}
	m.mu.Unlock()

	// singleflight barrier: serializes concurrent factory+Start calls for the same language (REQ-UTIL-003-004).
	// sf.Do shares the result of an in-flight call with all waiting callers for the same key.
	// After completion the key is automatically removed, so the next call runs fresh (REQ-UTIL-003-006 retry guarantee).
	v, err, _ := m.sf.Do(language, func() (any, error) {
		// Double-check inside sf.Do: another goroutine may have already inserted the client into the cache.
		m.mu.Lock()
		existing2, ok2 := m.clients[language]
		if ok2 && existing2.State() != StateShutdown {
			m.mu.Unlock()
			return existing2, nil
		}
		sc, hasCfg := m.servers[language]
		m.mu.Unlock()

		if !hasCfg {
			return nil, fmt.Errorf("lsp manager: no server config for language %q", language)
		}

		// Call factory: runs exactly once per language (REQ-UTIL-003-005).
		c := m.clientFactory(sc)

		// Call Start: runs outside the lock (may block on I/O).
		if err := c.Start(ctx); err != nil {
			// Start failed: do not insert into cache → next call can retry (REQ-UTIL-003-006).
			return nil, fmt.Errorf("lsp manager: failed to start client for language %q: %w", language, err)
		}

		// Insert into cache only after a successful Start (REQ-UTIL-003-006).
		m.mu.Lock()
		m.clients[language] = c
		m.mu.Unlock()
		return c, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(Client), nil
}

// reaper is the background goroutine that periodically shuts down idle clients.
//
// Behavior:
//   - Every reaperInterval, check lastActivity for each language.
//   - If time.Since(lastActivity) > idleShutdownSeconds, perform a graceful Shutdown.
//   - On ctx.Done(), close reaperDone and exit.
func (m *Manager) reaper(ctx context.Context, done chan struct{}) {
	defer close(done)

	ticker := time.NewTicker(m.reaperInterval)
	defer ticker.Stop()

	// shutdownTimeout: used for individual client graceful shutdown.
	const shutdownTimeout = 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.reapIdleClients(ctx, shutdownTimeout)
		}
	}
}

// reapIdleClients finds and shuts down clients that have exceeded the idle timeout.
func (m *Manager) reapIdleClients(ctx context.Context, shutdownTimeout time.Duration) {
	now := time.Now()

	m.mu.Lock()
	var toShutdown []struct {
		lang   string
		client Client
	}

	for lang, c := range m.clients {
		activity, hasActivity := m.lastActivity[lang]
		if !hasActivity {
			continue
		}

		elapsed := now.Sub(activity)
		threshold := time.Duration(m.idleShutdownSeconds) * time.Second

		if elapsed > threshold {
			toShutdown = append(toShutdown, struct {
				lang   string
				client Client
			}{lang: lang, client: c})
			delete(m.clients, lang)
			delete(m.lastActivity, lang)
		}
	}
	m.mu.Unlock()

	for _, item := range toShutdown {
		shutCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
		if err := item.client.Shutdown(shutCtx); err != nil {
			m.logger.Info("lsp manager: idle client shutdown error",
				slog.String("language", item.lang),
				slog.String("error", err.Error()),
			)
		} else {
			m.logger.Info("lsp manager: idle client shut down",
				slog.String("language", item.lang),
			)
		}
		cancel()
	}
}
