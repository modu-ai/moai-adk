---
id: SPEC-LSP-CORE-002
version: "1.0.0"
status: draft
created: "2026-04-11"
updated: "2026-04-11"
author: GOOS
priority: P2
issue_number: 0
phase: "Phase 4 - Multi-Language LSP"
module: "internal/lsp/core/, internal/lsp/transport/, internal/lsp/subprocess/"
estimated_loc: 2500
dependencies:
  - SPEC-GOPLS-BRIDGE-001
lifecycle: spec-anchored
tags: lsp, powernap, json-rpc, multi-language, client
supersedes: SPEC-LSP-001
---

# SPEC-LSP-CORE-002: Multi-Language LSP Client Core (powernap-based)

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-04-11 | 1.0.0 | Initial draft; supersedes SPEC-LSP-001 (which was marked Completed but never actually implemented) |

---

## Overview

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-LSP-CORE-002 |
| Title | Multi-Language LSP Client Core |
| Status | Draft |
| Priority | P2 |
| Research Source | R1/R2/C1/C2/C3 (2026-04-11) |
| Library Decision | `github.com/charmbracelet/powernap` (crush-validated) |

## Problem Statement

SPEC-LSP-001 (2026-02-03) is marked `Completed` but the actual `internal/lsp/` package contains only interface stubs. Real LSP client implementation is nil. SPEC-GOPLS-BRIDGE-001 addresses Go-only needs; this SPEC generalizes to a **multi-language LSP client core** that will serve as the runtime foundation for all 16 MoAI-supported languages.

Path A vs Path B vs Path C tradeoff (from C1-C3 research):
- **Path A (own JSON-RPC)**: chosen for SPEC-GOPLS-BRIDGE-001 (Go-specific, small scope)
- **Path B (MCP bridge)**: exists in SPEC-LSPMCP-001 (draft), complementary
- **Path C (powernap)**: this SPEC, for multi-language production-grade client

## Goal

Deliver a production-grade LSP client core that:

1. Uses `github.com/charmbracelet/powernap` (battle-tested in charmbracelet/crush, 23k+ stars)
2. Supports N language servers spawned lazily per user project
3. Provides a unified `Client` interface consumable by Ralph, Quality Gates, and MCP bridges
4. Handles lifecycle (spawn, initialize, shutdown, crash recovery) uniformly
5. Isolates each language server in its own subprocess

## Requirements (EARS Format)

**REQ-LC-001**: The system SHALL depend on `github.com/charmbracelet/powernap` at a pinned version recorded in `go.mod`.

**REQ-LC-002**: The system SHALL provide a `Client` interface with methods: `Start(ctx)`, `OpenFile(ctx, path, content)`, `GetDiagnostics(ctx, path)`, `FindReferences(ctx, path, position)`, `GotoDefinition(ctx, path, position)`, `Shutdown(ctx)`.

**REQ-LC-003**: Language server configuration SHALL be read from `.moai/config/sections/lsp.yaml` `lsp.servers.<language>` entries (already Section 22 compliant).

**REQ-LC-004**: When a language server binary is missing in PATH, the client SHALL log a `warn_and_skip` and return an unavailable sentinel — not crash.

**REQ-LC-005**: Each language server instance SHALL run in its own subprocess with isolated stdio pipes.

**REQ-LC-006**: The client SHALL cache open file state per subprocess to avoid redundant `didOpen` on repeated requests.

**REQ-LC-007**: The client SHALL support graceful shutdown with configurable timeout; on timeout, SIGKILL is sent.

**REQ-LC-008**: A single `Manager` type SHALL coordinate multiple `Client` instances (one per language) and delegate calls based on file extension + project markers.

**REQ-LC-009**: The `Manager` SHALL respect lazy initialization: a language server is spawned only when the first file of that language is processed.

**REQ-LC-010**: `SPEC-GOPLS-BRIDGE-001` SHALL remain available as an alternative path; SPEC-LSP-CORE-002 does NOT deprecate it. Users can choose between hand-rolled (GOPLS-BRIDGE) and powernap-based (LSP-CORE) via config.

## Non-Goals

- Aggregation across servers (that's SPEC-LSP-AGG-003)
- Phase-aware quality gates (that's SPEC-LSP-QGATE-004)
- Loop/Ralph integration (that's SPEC-LSP-LOOP-005)
- 16-language config matrix (that's SPEC-LSP-MULTI-006; this SPEC ships with Go, Python, TypeScript as validation targets)

## References

- Phase 1 reports R1, R2
- Phase 2 reports C1 (sst/opencode), C2 (charmbracelet/crush — uses powernap)
- [powernap library](https://github.com/charmbracelet/powernap)
- SPEC-LSP-001 (superseded)
- SPEC-GOPLS-BRIDGE-001 (complementary)
