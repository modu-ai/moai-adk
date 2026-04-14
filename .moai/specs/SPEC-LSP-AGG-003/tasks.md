## Task Decomposition
SPEC: SPEC-LSP-AGG-003
Methodology: TDD (RED-GREEN-REFACTOR)
Mode: sub-agent (Standard Mode)

### Task Table

| Task ID | Description | Requirement | Dependencies | Planned Files | Status |
|---------|-------------|-------------|--------------|---------------|--------|
| T-001 | CacheEntry type + IsExpired() | REQ-AGG-003, REQ-AGG-004 | — | cache/types.go, cache/types_test.go | pending |
| T-002 | DiagnosticCache.Get/Set — (uri, version) key, version mismatch = miss | REQ-AGG-003, REQ-AGG-006 | T-001 | cache/cache.go, cache/cache_test.go | pending |
| T-003 | DiagnosticCache.Invalidate + TTL expiry (default 5s, configurable) | REQ-AGG-004, REQ-AGG-005 | T-002 | cache/cache.go (ext), cache/cache_test.go (ext) | pending |
| T-004 | Cache cleanup goroutine Start/Stop lifecycle | REQ-AGG-004 | T-003 | cache/cache.go (ext), cache/cache_test.go (ext) | pending |
| T-005 | Manager.routeFor → RouteFor export + callsite update | REQ-AGG-002 | — | core/manager.go (mod), core/manager_test.go (mod) | pending |
| T-006 | Aggregator type + NewAggregator + GetDiagnostics basic (cache miss → upstream) | REQ-AGG-001, REQ-AGG-002 | T-002, T-005 | aggregator/aggregator.go, aggregator/aggregator_test.go | done |
| T-007 | singleflight integration — concurrent dedup for same file | REQ-AGG-007, REQ-AGG-010 | T-006 | aggregator/aggregator.go (ext), aggregator/aggregator_test.go (ext) | done |
| T-008 | Cache integration — hit skip upstream, version refresh, Invalidate | REQ-AGG-003, REQ-AGG-006 | T-003, T-006 | aggregator/aggregator.go (ext), aggregator/aggregator_test.go (ext) | done |
| T-009 | CircuitBreaker integration — per-server, 3 fail → open, 30s | REQ-AGG-009 | T-006 | aggregator/aggregator.go (ext), aggregator/aggregator_test.go (ext) | done |
| T-010 | Query timeout + graceful degradation — timeout → cache fallback, else empty | REQ-AGG-008, REQ-AGG-010 | T-008, T-009 | aggregator/aggregator.go (ext), aggregator/aggregator_test.go (ext) | done |

### Execution Order
T-001 > T-002 > T-003 > T-004 > T-005 > T-006 > T-007 > T-008 > T-009 > T-010

### Key Design Decisions
- CircuitBreaker: reuse internal/resilience.CircuitBreaker (Threshold=3, Timeout=30s)
- singleflight: golang.org/x/sync/singleflight (new dep)
- Manager access: RouteFor export (public method)
- Fake strategy: interface-based fake Router/Client for unit test isolation

### Success Criteria
- [ ] go test -race ./internal/lsp/cache/... PASS
- [ ] go test -race ./internal/lsp/aggregator/... PASS
- [ ] go test -race ./internal/lsp/core/... zero regression (RouteFor export)
- [ ] cache + aggregator per-package coverage >= 85%
- [ ] go vet clean
- [ ] REQ-AGG-001~010 all implemented
- [ ] resilience.CircuitBreaker reused (Threshold=3, Timeout=30s)
- [ ] singleflight integrated
- [ ] Cache cleanup goroutine leak-free (Start/Stop pair verified)
- [ ] @MX:ANCHOR on Aggregator.GetDiagnostics, DiagnosticCache.Get (fan_in>=3)
