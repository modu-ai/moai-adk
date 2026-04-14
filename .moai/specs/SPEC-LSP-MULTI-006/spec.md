---
id: SPEC-LSP-MULTI-006
version: "1.0.0"
status: completed
created: "2026-04-11"
updated: "2026-04-11"
author: GOOS
priority: P3
issue_number: 0
phase: "Phase 5 - Full Language Matrix"
module: "internal/lsp/core/, .moai/config/sections/lsp.yaml"
estimated_loc: 2900
dependencies:
  - SPEC-LSP-CORE-002
  - SPEC-LSP-AGG-003
lifecycle: spec-anchored
tags: lsp, multi-language, 16-languages, neutrality, discovery
---

# SPEC-LSP-MULTI-006: 16-Language LSP Server Matrix

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-04-11 | 1.0.0 | Initial draft — activates all 16 languages |

---

## Overview

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-LSP-MULTI-006 |
| Title | 16-Language LSP Server Matrix |
| Status | Draft |
| Priority | P3 |
| Depends on | SPEC-LSP-CORE-002, SPEC-LSP-AGG-003 |

## Problem Statement

The `lsp.yaml.tmpl` (PR #625, PR #627) declares 16 language server entries for documentation purposes, but the runtime code only validates Go (SPEC-GOPLS-BRIDGE-001) and Go/Python/TypeScript (SPEC-LSP-CORE-002 initial validation). This SPEC activates the full 16-language matrix and adds discovery + user-friendly onboarding.

## Goal

1. **Validate all 16 language servers** can be spawned and produce diagnostics via integration tests
2. **Discovery logic** detects user's project language via marker files and activates the corresponding server
3. **User onboarding** provides install hints when a language server is missing
4. **Section 22 compliance** maintained: no language receives priority over another

## Requirements (EARS Format)

**REQ-LM-001**: The `Manager` SHALL detect project language(s) via `project_markers` defined in `lsp.yaml`.

**REQ-LM-002**: The `Manager` SHALL spawn ONLY language servers matching detected markers (not all 16 eagerly).

**REQ-LM-003**: For each of the 16 languages, the system SHALL have an integration test that:
  - Skips gracefully when the binary is missing
  - Spawns the real server when available
  - Produces at least 1 diagnostic from a fixture error file

**REQ-LM-004**: When a language server binary is missing, the system SHALL log a `warn_and_skip` with the install hint from `lsp.yaml.servers.<lang>.install_hint`.

**REQ-LM-005**: Discovery SHALL prefer project-local servers (`node_modules/.bin`, `.venv/bin`) over global PATH.

**REQ-LM-006**: The 16 languages supported SHALL exactly match the canonical list in `.claude/skills/moai/workflows/sync.md` Phase 0.6.1 Language Detection table: `cpp, csharp, elixir, flutter, go, java, javascript, kotlin, php, python, r, ruby, rust, scala, swift, typescript`.

**REQ-LM-007**: A diagnostic "doctor" subcommand `moai lsp doctor` SHALL report:
  - Which languages the current project uses
  - Which language servers are installed
  - Which are missing + install hints
  - Aggregate readiness status

**REQ-LM-008**: Fallback binaries (e.g., `pyright-langserver` when `pylsp` missing) SHALL be tried in order as defined in `lsp.yaml.servers.<lang>.fallback_binaries`.

**REQ-LM-009**: Server capability differences (e.g., some servers don't support `textDocument/rename`) SHALL be tracked per-server and exposed via `Client.Capabilities()`.

**REQ-LM-010**: The aggregator SHALL NOT error when a language's server is unavailable; it SHALL skip that language's diagnostics silently (or with a single warn at startup).

## Non-Goals

- Custom rules per language in LSP settings (future enhancement)
- Bidirectional file watchers (LSP 3.17 provides this; wait for concrete use case)
- Auto-install of language servers (install hint printed only; user runs installer)

## References

- CLAUDE.local.md Section 22 (Template Language Neutrality)
- `.claude/skills/moai/workflows/sync.md` Phase 0.6.1 (canonical 16-language list)
- `.moai/config/sections/lsp.yaml.tmpl` (language matrix)
- Section 22 audit reports (PRs #625, #627, #628)
