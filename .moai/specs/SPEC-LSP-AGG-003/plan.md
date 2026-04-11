# SPEC-LSP-AGG-003: Implementation Plan

## Phase Breakdown

### Phase 1: DiagnosticCache
- `internal/lsp/cache/diagnostics.go`: TTL cache with version key
- `internal/lsp/cache/diagnostics_test.go`: TTL expiration, version invalidation, concurrent access
- Use sync.RWMutex; background cleanup goroutine
- **Estimated LOC**: +350

### Phase 2: Circuit Breaker
- `internal/lsp/aggregator/circuit_breaker.go`: per-server state (closed/open/half-open)
- Consecutive failure tracking with reset on success
- Configurable threshold and open_duration
- **Estimated LOC**: +250

### Phase 3: Aggregator Core
- `internal/lsp/aggregator/aggregator.go`: main type wiring Manager + Cache + CircuitBreaker
- `singleflight.Group` for request deduplication
- Parallel query when multiple servers serve the same file (rare, but supported)
- **Estimated LOC**: +400

### Phase 4: Merge + Deduplicate
- `internal/lsp/aggregator/merge.go`: merge diagnostics from multiple sources
- Dedupe by `(range, code, source, message)` tuple
- Sort by severity then range
- **Estimated LOC**: +200

### Phase 5: Integration Tests
- `internal/lsp/aggregator/integration_test.go`: exercise full path with mock Manager
- Verify cache hits, misses, concurrent access, circuit breaker behavior
- **Estimated LOC**: +400

### Phase 6: deps.go Wiring
- Wire Aggregator into `GoFeedbackGenerator` and quality gate
- Feature flag: `lsp.aggregator.enabled`
- **Estimated LOC**: +50 modified

## Dependencies

- **Hard**: SPEC-LSP-CORE-002 complete
- **Blocks**: SPEC-LSP-QGATE-004, SPEC-LSP-LOOP-005

## Risks

- singleflight deadlock on context cancellation: test explicitly with canceled contexts
- Memory growth from cache: bounded by LRU eviction with max_entries config
- Circuit breaker flapping: hysteresis with half-open recovery state
