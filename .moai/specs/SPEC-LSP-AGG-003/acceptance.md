# SPEC-LSP-AGG-003: Acceptance Criteria

## Functional

### AC1: Cache Operations
- [ ] `Cache.Get(uri, version)` returns hit for matching entry within TTL (REQ-AGG-003, REQ-AGG-004)
- [ ] `Cache.Get(uri, version)` returns miss when version differs (REQ-AGG-006)
- [ ] `Cache.Get(uri, version)` returns miss when TTL expired (REQ-AGG-004)
- [ ] `Cache.Set(uri, diags, version, ttl)` stores entry (REQ-AGG-003)
- [ ] `Cache.Invalidate(uri)` removes entry (REQ-AGG-005)
- [ ] Background cleanup removes expired entries periodically (REQ-AGG-004)

### AC2: Circuit Breaker
- [ ] 3 consecutive failures open the circuit for that server (REQ-AGG-009)
- [ ] Open circuit returns cached results without upstream call (REQ-AGG-008, REQ-AGG-009)
- [ ] After 30s open duration, circuit transitions to half-open (REQ-AGG-009)
- [ ] Single successful call in half-open closes the circuit (REQ-AGG-009)
- [ ] Single failure in half-open reopens the circuit (REQ-AGG-009)

### AC3: Aggregator Core
- [ ] `Aggregator.GetDiagnostics(ctx, path)` returns merged diagnostics (REQ-AGG-001, REQ-AGG-002)
- [ ] Multiple concurrent calls for same uri are deduped (singleflight) (REQ-AGG-007)
- [ ] Per-query timeout honored; returns cached results on timeout (REQ-AGG-008)
- [ ] Graceful degradation on individual server failure (REQ-AGG-009)

### AC4: Merge + Dedupe
- [ ] Duplicate diagnostics (same range, code, source, message) are deduped (REQ-AGG-001)
- [ ] Results sorted by severity desc, then range asc (REQ-AGG-001)
- [ ] Different severities for same location are preserved (both kept) (REQ-AGG-001)

### AC5: Parallel Execution
- [ ] When a file has bindings to multiple servers (e.g., a .tsx file to tsserver + eslint-LS), queries run in parallel (REQ-AGG-001, REQ-AGG-010)
- [ ] Slow server does not block fast server's results from being cached (REQ-AGG-010)
- [ ] `go test -race` passes (REQ-AGG-010)

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
