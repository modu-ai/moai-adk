package aggregator

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/cache"
	"github.com/modu-ai/moai-adk/internal/lsp/core"
	"github.com/modu-ai/moai-adk/internal/resilience"
	"golang.org/x/sync/singleflight"
)

// defaultQueryTimeout is applied when no WithQueryTimeout option is given.
const defaultQueryTimeout = 5 * time.Second

// defaultCacheTTL is the TTL used when NewAggregator creates its own cache.
const defaultCacheTTL = 5 * time.Second

// defaultCBThreshold is the consecutive-failure count that opens a circuit breaker.
const defaultCBThreshold = 3

// defaultCBTimeout is the open-circuit duration before transitioning to half-open.
const defaultCBTimeout = 30 * time.Second

// Router is the interface used by Aggregator to obtain a Client for a file path.
// Manager from SPEC-LSP-CORE-002 satisfies this interface because it exports
// RouteFor(ctx, path) (core.Client, error).
//
// @MX:NOTE: [AUTO] Router interface is a DI boundary that allows injecting fakes in tests without coupling to Manager
type Router interface {
	RouteFor(ctx context.Context, path string) (core.Client, error)
}

// Aggregator coordinates parallel LSP diagnostic queries across multiple language
// servers. It integrates a TTL-based cache, singleflight deduplication, per-language
// circuit breakers, and per-query timeouts for graceful degradation.
//
// @MX:ANCHOR: [AUTO] Aggregator — core diagnostic collection type shared by Ralph, QGATE, LOOP, MCP and other callers
// @MX:REASON: fan_in >= 3 — Ralph engine, Quality Gate, LOOP command, and MCP bridge all obtain diagnostics through Aggregator
type Aggregator struct {
	router       Router
	cache        *cache.DiagnosticCache
	sf           singleflight.Group
	breakers     map[string]*resilience.CircuitBreaker
	mu           sync.RWMutex
	queryTimeout time.Duration
	cbConfig     resilience.CircuitBreakerConfig
	logger       *slog.Logger
}

// Option configures an Aggregator at construction time.
type Option func(*Aggregator)

// WithCache replaces the default DiagnosticCache with the provided one.
func WithCache(c *cache.DiagnosticCache) Option {
	return func(a *Aggregator) {
		a.cache = c
	}
}

// WithQueryTimeout sets the per-query upstream timeout.
// The default is 5 seconds.
func WithQueryTimeout(d time.Duration) Option {
	return func(a *Aggregator) {
		a.queryTimeout = d
	}
}

// WithLogger sets the logger for lifecycle events.
// @MX:ANCHOR: [AUTO] WithLogger is a functional option used across 3+ packages
// @MX:REASON: Multi-package dependency — signature changes affect Ralph, QGATE, LOOP
func WithLogger(logger *slog.Logger) Option {
	return func(a *Aggregator) {
		a.logger = logger
	}
}

// WithCircuitBreakerConfig sets the circuit breaker configuration applied to
// each per-language breaker.
func WithCircuitBreakerConfig(cfg resilience.CircuitBreakerConfig) Option {
	return func(a *Aggregator) {
		a.cbConfig = cfg
	}
}

// NewAggregator constructs a new Aggregator backed by the given router.
//
// @MX:ANCHOR: [AUTO] NewAggregator — sole entry point for constructing an Aggregator
// @MX:REASON: fan_in >= 3 — Ralph engine, Quality Gate, integration tests, and MCP bridge all call NewAggregator
func NewAggregator(router Router, opts ...Option) *Aggregator {
	a := &Aggregator{
		router:       router,
		breakers:     make(map[string]*resilience.CircuitBreaker),
		queryTimeout: defaultQueryTimeout,
		cbConfig: resilience.CircuitBreakerConfig{
			Threshold: defaultCBThreshold,
			Timeout:   defaultCBTimeout,
		},
		logger: slog.Default(),
	}
	for _, opt := range opts {
		opt(a)
	}
	// Create a default cache if none was injected.
	if a.cache == nil {
		a.cache = cache.NewDiagnosticCache(defaultCacheTTL)
	}
	return a
}

// Start begins the cache cleanup goroutine.
// It is safe to call GetDiagnostics before Start, but cache cleanup will not
// run until Start is called.
func (a *Aggregator) Start(ctx context.Context) error {
	a.cache.Start(ctx)
	return nil
}

// Shutdown stops the cache cleanup goroutine.
func (a *Aggregator) Shutdown(_ context.Context) error {
	a.cache.Stop()
	return nil
}

// GetDiagnostics returns diagnostics for the file at path.
// On a cache miss it queries the upstream client via the router.
// Concurrent calls for the same path are deduplicated via singleflight.
// If the circuit breaker for the file's language is open, cached or empty
// results are returned (graceful degradation).
// A per-query timeout is applied; on deadline the latest cached value is
// returned if available, otherwise an empty slice is returned (REQ-AGG-008).
//
// @MX:ANCHOR: [AUTO] Aggregator.GetDiagnostics — core path for collecting diagnostics
// @MX:REASON: fan_in >= 3 — Ralph engine, Quality Gate, LOOP command, and MCP bridge all request diagnostics through this method
func (a *Aggregator) GetDiagnostics(ctx context.Context, path string) ([]lsp.Diagnostic, error) {
	uri := path // v1: use path directly as cache key

	// Apply per-query timeout (REQ-AGG-008).
	qCtx, cancel := context.WithTimeout(ctx, a.queryTimeout)
	defer cancel()

	// Check cache first (TTL-only, version=0 always for v1).
	if cached, ok := a.cache.Get(uri, 0); ok {
		return cached, nil
	}

	// Determine language for circuit breaker lookup.
	lang := detectLanguageFromPath(path)

	// Check circuit breaker state (REQ-AGG-009).
	cb := a.getOrCreateBreaker(lang)
	if cb.State() == resilience.StateOpen {
		return a.degradeOrEmpty(uri), nil
	}

	// Deduplicate concurrent requests for the same URI (REQ-AGG-007).
	type result struct {
		diags []lsp.Diagnostic
		err   error
	}
	v, err, _ := a.sf.Do(uri, func() (any, error) {
		client, routeErr := a.router.RouteFor(qCtx, path)
		if routeErr != nil {
			return result{nil, routeErr}, nil
		}

		var cbErr error
		var diags []lsp.Diagnostic
		cbErr = cb.Call(qCtx, func() error {
			var getErr error
			diags, getErr = client.GetDiagnostics(qCtx, path)
			return getErr
		})
		if cbErr != nil {
			return result{nil, cbErr}, nil
		}

		// Store in cache on success.
		a.cache.Set(uri, 0, diags)
		return result{diags, nil}, nil
	})

	// Handle context deadline — return stale cache or empty.
	if err == nil && qCtx.Err() != nil {
		return a.degradeOrEmpty(uri), nil
	}
	if err != nil {
		return nil, err
	}

	r := v.(result)
	if r.err != nil {
		if r.err == resilience.ErrCircuitOpen {
			// Circuit opened during the call — graceful degradation.
			return a.degradeOrEmpty(uri), nil
		}
		return nil, fmt.Errorf("aggregator: get diagnostics for %q: %w", path, r.err)
	}
	return r.diags, nil
}

// Invalidate removes the cache entry for the given file URI (REQ-AGG-005).
func (a *Aggregator) Invalidate(uri string) {
	a.cache.Invalidate(uri)
}

// degradeOrEmpty returns stale cached diagnostics if available, otherwise an
// empty slice. It is the graceful-degradation fallback for circuit-open and
// timeout scenarios (REQ-AGG-008, REQ-AGG-009, REQ-AGG-010).
func (a *Aggregator) degradeOrEmpty(uri string) []lsp.Diagnostic {
	if stale, ok := a.cache.GetStale(uri); ok {
		return stale
	}
	return []lsp.Diagnostic{}
}

// getOrCreateBreaker returns the CircuitBreaker for the given language,
// creating one with the configured settings if it does not yet exist.
func (a *Aggregator) getOrCreateBreaker(lang string) *resilience.CircuitBreaker {
	a.mu.RLock()
	cb, ok := a.breakers[lang]
	a.mu.RUnlock()
	if ok {
		return cb
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	// Double-check after acquiring write lock.
	if cb, ok = a.breakers[lang]; ok {
		return cb
	}
	cb = resilience.NewCircuitBreaker(a.cbConfig)
	a.breakers[lang] = cb
	return cb
}

// detectLanguageFromPath maps common file extensions to language identifiers.
// This avoids coupling Aggregator to Manager's unexported detectLanguage logic.
//
// @MX:NOTE: [AUTO] Extension-based language detection helper — determines the CircuitBreaker map key without coupling to Manager
func detectLanguageFromPath(path string) string {
	// Find the last dot in the base name only.
	base := path
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			base = path[i+1:]
			break
		}
	}
	for i := len(base) - 1; i >= 0; i-- {
		if base[i] == '.' {
			ext := base[i:]
			switch ext {
			case ".go":
				return "go"
			case ".ts", ".tsx":
				return "typescript"
			case ".js", ".jsx", ".mjs", ".cjs":
				return "javascript"
			case ".py":
				return "python"
			case ".rs":
				return "rust"
			case ".java":
				return "java"
			case ".kt", ".kts":
				return "kotlin"
			case ".cs":
				return "csharp"
			case ".rb":
				return "ruby"
			case ".php":
				return "php"
			case ".ex", ".exs":
				return "elixir"
			case ".cpp", ".cc", ".cxx", ".hpp":
				return "cpp"
			case ".scala":
				return "scala"
			case ".r", ".R":
				return "r"
			case ".dart":
				return "flutter"
			case ".swift":
				return "swift"
			}
			break
		}
	}
	return "unknown"
}
