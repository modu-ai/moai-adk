---
id: SPEC-LSP-AGG-003
version: "1.0.0"
status: draft
created: "2026-04-11"
updated: "2026-04-11"
author: GOOS
priority: P2
issue_number: 0
phase: "Phase 4 - Multi-Language LSP"
module: "internal/lsp/aggregator/, internal/lsp/cache/"
estimated_loc: 900
dependencies:
  - SPEC-LSP-CORE-002
lifecycle: spec-anchored
tags: lsp, aggregator, cache, ttl, parallel, diagnostics
---

# SPEC-LSP-AGG-003: Multi-Server Diagnostic Aggregator with TTL Cache

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-04-11 | 1.0.0 | Initial draft |

---

## Overview

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-LSP-AGG-003 |
| Title | Multi-Server Diagnostic Aggregator with TTL Cache |
| Priority | P2 |
| Depends on | SPEC-LSP-CORE-002 |

## Problem Statement

With multiple LSP servers spawned (one per user's project language), diagnostics calls must be aggregated efficiently. Without caching, every hook invocation re-queries all servers. Without parallelism, slow servers block fast ones. The existing `internal/lsp/` package has no aggregator.

Research findings (R2/C1/C2):
- **Biome**: AST cache + invalidation epoch
- **golangci-lint**: parallel linter orchestration with result merging
- **Deno**: TTL-based result cache (8s → <1s optimization)
- **rust-analyzer**: Salsa incremental (too heavy for us)

## Goal

Provide an Aggregator that:

1. Queries multiple LSP clients (one per language) **in parallel**
2. Caches results per file with **TTL + version-based invalidation**
3. Merges and deduplicates diagnostics from all servers
4. Degrades gracefully when individual servers fail or time out

## Requirements (EARS Format)

**REQ-AGG-001**: The system SHALL provide an `Aggregator` type that accepts a `Manager` (from SPEC-LSP-CORE-002) and coordinates parallel queries.

**REQ-AGG-002**: When `Aggregator.GetDiagnostics(ctx, path)` is called, the system SHALL detect the file's language via extension and project markers, then query the matching `Client` via the Manager.

**REQ-AGG-003**: Results SHALL be cached in a `DiagnosticCache` keyed by `(uri, version)` tuple.

**REQ-AGG-004**: Cache entries SHALL have a TTL (default 5 seconds, configurable via `lsp.aggregator.cache_ttl_seconds`).

**REQ-AGG-005**: The cache SHALL support manual invalidation: `Cache.Invalidate(uri)` removes the entry immediately.

**REQ-AGG-006**: When a file version changes (detected via document version field), the cache entry SHALL be considered stale and refreshed.

**REQ-AGG-007**: Multiple concurrent calls to `GetDiagnostics` for the same file SHALL be deduplicated via singleflight pattern (only one upstream query per cache miss).

**REQ-AGG-008**: Per-query timeouts SHALL be configurable (default 5 seconds); on timeout, `Aggregator` returns cached results if any, else empty.

**REQ-AGG-009**: Circuit breaker SHALL open when a specific server fails 3 consecutive times; open duration is 30 seconds (config).

**REQ-AGG-010**: The Aggregator SHALL be safe for concurrent use from multiple goroutines.

## Non-Goals

- Phase-aware quality gate enforcement (that's SPEC-LSP-QGATE-004)
- Persistent disk cache (in-memory only for v1)
- Cross-session cache reuse

## Architecture

```
Aggregator
├── *Manager (from SPEC-LSP-CORE-002)
├── *DiagnosticCache (TTL + version keyed)
├── *CircuitBreaker (per-server failure tracking)
└── singleflight.Group (dedupe concurrent requests)

DiagnosticCache
├── sync.RWMutex
├── map[string]*CacheEntry  // key: uri
└── cleanupLoop (periodic expired entry removal)

CacheEntry {
    Diagnostics []Diagnostic
    Version     int64
    ExpiresAt   time.Time
}
```

## References

- Phase 1 reports R2 (reference tool architectures)
- Deno optimization blog (TTL cache pattern)
- Golang singleflight package
