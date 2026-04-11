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

### Dependencies

**REQ-LC-001**: The system SHALL depend on `github.com/charmbracelet/powernap` at a pinned version recorded in `go.mod`.

**REQ-LC-001a**: The pinned version SHALL be documented in a `.claude/rules/moai/core/lsp-client.md` rationale section, and upgrades SHALL require integration test re-validation against all 3 languages (Go, Python, TypeScript) before merge.

### Client Interface — Lifecycle Operations

**REQ-LC-002**: The system SHALL provide a `Client` interface exposing lifecycle operations `Start(ctx)` and `Shutdown(ctx)`, each accepting a `context.Context` as the first parameter.

**REQ-LC-002a**: The system SHALL provide a `Client` interface exposing file synchronization operations `OpenFile(ctx, path, content)` and an internal companion for handling document changes (detailed in REQ-LC-020..023).

**REQ-LC-002b**: The system SHALL provide a `Client` interface exposing query operations `GetDiagnostics(ctx, path)`, `FindReferences(ctx, path, position)`, and `GotoDefinition(ctx, path, position)`.

### Configuration

**REQ-LC-003**: Language server configuration SHALL be read from `.moai/config/sections/lsp.yaml` `lsp.servers.<language>` entries (already Section 22 compliant).

**REQ-LC-003a**: Per-language `initializationOptions` SHALL be merged from the `lsp.servers.<language>.init_options` field into the LSP `initialize` request payload sent to each server.

### Graceful Degradation

**REQ-LC-004**: When a language server binary is missing in PATH, the client SHALL log a `warn_and_skip` and return an unavailable sentinel — not crash.

### Subprocess Isolation

**REQ-LC-005**: Each language server instance SHALL run in its own subprocess with isolated stdio pipes.

### Document Synchronization

**REQ-LC-006**: The client SHALL cache open file state per subprocess to avoid redundant `didOpen` on repeated requests.

**REQ-LC-020**: When a previously unopened file is queried, the client SHALL send `textDocument/didOpen` with the file URI, language ID, version `1`, and current content.

**REQ-LC-021**: When a file's content changes between queries, the client SHALL send `textDocument/didChange` with an incremented version number and the updated content (full document sync mode in v1).

**REQ-LC-022**: When a file is no longer tracked (configurable idle timeout, default 5 minutes), the client SHALL send `textDocument/didClose` to release server-side resources.

**REQ-LC-023**: On request of the caller, the client SHALL send `textDocument/didSave` to inform the server of persisted changes; this is optional and used only for servers that gate diagnostics on save (e.g., some Java servers).

### Lifecycle State Machine

**REQ-LC-007**: The client SHALL support graceful shutdown with configurable timeout; on timeout, SIGKILL is sent.

**REQ-LC-030**: Each `Client` SHALL maintain an internal state machine with states `spawning`, `initializing`, `ready`, `degraded`, and `shutdown`; state transitions SHALL be logged at `slog.Debug` level.

**REQ-LC-031**: In the `degraded` state (server crash recovery in progress), the client SHALL return empty results for queries without error, allowing upstream callers to continue.

### Capability Negotiation

**REQ-LC-032**: During initialization, the client SHALL send `ClientCapabilities` listing support for `textDocument/publishDiagnostics`, `textDocument/references`, and `textDocument/definition`. Additional capabilities MAY be declared per-language.

**REQ-LC-033**: The client SHALL parse and store the server's `ServerCapabilities` from the `initialize` response; query methods SHALL check capability support before sending a request and return a structured `ErrCapabilityUnsupported` error for unsupported operations.

### Error Handling

**REQ-LC-040**: Each LSP method invocation SHALL wrap protocol errors with contextual information (method name, file URI, server language) before returning to the caller.

**REQ-LC-041**: Request timeouts SHALL be enforced per-method via the provided `context.Context`; on timeout, the pending request SHALL be removed from the correlation map to prevent memory leaks.

### Manager (Multi-Server Coordinator)

**REQ-LC-008**: A single `Manager` type SHALL coordinate multiple `Client` instances (one per language) and delegate calls based on file extension + project markers.

**REQ-LC-009**: The `Manager` SHALL respect lazy initialization: a language server is spawned only when the first file of that language is processed.

**REQ-LC-050**: The `Manager` SHALL implement an idle-server cleanup policy: a `Client` with no activity for the configured `idle_shutdown_seconds` (default 600) SHALL be gracefully shut down to release subprocess resources.

### Compatibility

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
