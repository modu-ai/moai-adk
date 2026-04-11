# SPEC-LSP-AGG-003: Acceptance Criteria

## Functional

### AC1: Cache Operations
- [ ] `Cache.Get(uri, version)` returns hit for matching entry within TTL
- [ ] `Cache.Get(uri, version)` returns miss when version differs
- [ ] `Cache.Get(uri, version)` returns miss when TTL expired
- [ ] `Cache.Set(uri, diags, version, ttl)` stores entry
- [ ] `Cache.Invalidate(uri)` removes entry
- [ ] Background cleanup removes expired entries periodically

### AC2: Circuit Breaker
- [ ] 3 consecutive failures open the circuit for that server
- [ ] Open circuit returns cached results without upstream call
- [ ] After 30s open duration, circuit transitions to half-open
- [ ] Single successful call in half-open closes the circuit
- [ ] Single failure in half-open reopens the circuit

### AC3: Aggregator Core
- [ ] `Aggregator.GetDiagnostics(ctx, path)` returns merged diagnostics
- [ ] Multiple concurrent calls for same uri are deduped (singleflight)
- [ ] Per-query timeout honored; returns cached results on timeout
- [ ] Graceful degradation on individual server failure

### AC4: Merge + Dedupe
- [ ] Duplicate diagnostics (same range, code, source, message) are deduped
- [ ] Results sorted by severity desc, then range asc
- [ ] Different severities for same location are preserved (both kept)

### AC5: Parallel Execution
- [ ] When a file has bindings to multiple servers (e.g., a .tsx file to tsserver + eslint-LS), queries run in parallel
- [ ] Slow server does not block fast server's results from being cached
- [ ] `go test -race` passes

## Quality (TRUST 5)

### Tested
- [ ] ≥ 85% coverage for `internal/lsp/aggregator/` and `internal/lsp/cache/`
- [ ] Race detector tests for concurrent cache access
- [ ] Circuit breaker state transition tests

### Readable
- [ ] Architecture diagram in `.claude/rules/moai/core/lsp-aggregator.md`
- [ ] Godoc on all exported types

### Unified
- [ ] Cache type reusable by future features (e.g., reference cache)
- [ ] Circuit breaker shared with existing `internal/resilience/` package

### Secured
- [ ] No goroutine leaks on context cancellation
- [ ] Cache memory bounded by max_entries config

### Trackable
- [ ] `@MX:ANCHOR` on `Aggregator.GetDiagnostics`
- [ ] SPEC-LSP-AGG-003 referenced in commits

## Deliverables

- [ ] 6+ new Go files
- [ ] 4+ test files (unit + integration)
- [ ] 1 architecture doc
- [ ] CHANGELOG entry
