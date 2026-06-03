# SPEC-LSPMCP-RETIRE-001 — Progress

## §A Overview

- **SPEC**: Retire dormant moai-lsp MCP bridge
- **Tier**: M (Medium)
- **plan-auditor verdict**: PASS 0.93 (iter 2/3; iter-1 FAIL 0.82 → D1 MissingExclusions fixed → monotonic +0.11)

## §E Phase 0.95 Mode Selection

Input parameters:
- tier: M
- scope: ~14 logical paths, deletion-dominant (~1197 LOC removed)
- domain count: Go package (`internal/mcp`) + CLI subcommand (`internal/cli/mcp.go`) + deployable template (`.mcp.json.tmpl`) + byte-mirrored docs pair (`settings-management.md`) + predecessor SPEC frontmatter
- file language mix: Go + Go-template + markdown
- concurrency benefit: LOW (coding-heavy)
- Agent Teams prereqs: not met

Mode evaluation:
- trivial — not selected (multi-file semantic removal, not a single-line change)
- background — not selected (write/delete operation, not read-only)
- agent-team — not selected (REQ-ATR-013 prereqs not met + coding-heavy)
- parallel — not selected (coding-heavy; Finding A4 caveat)
- sub-agent — **SELECTED**

**Decision: sub-agent**

Justification: Coding-heavy dead-code removal with a critical `internal/mcp` (remove) vs `internal/lsp` (PRESERVE — 10 consumers) boundary. Per Finding A4 (most coding tasks involve fewer truly parallelizable subtasks than research), sequential sub-agent (Mode 5) is the correct default. Executed as a single `manager-develop` delegation with the Tier M Section A-E template.

## §E.2 Run-phase Audit-Ready Signal

- **run_commit_sha**: `94d9fccfc` (pushed to origin/main; local fast-forward aligned at `a97206dc7` after absorbing a disjoint parallel-session commit)
- **cycle_type**: tdd (deletion → PRESERVE-reverify)
- **Removal delivered**: `internal/mcp/` (10 files) + `internal/cli/mcp.go` (`moai mcp` / `moai mcp lsp` cobra surface) + `.mcp.json.tmpl` moai-lsp entry + `internal/template/settings_test.go` (2 blocking functions) + `settings-management.md` byte-mirror pair (2 `moai-lsp` `alwaysLoad` refs) + predecessor SPEC-LSPMCP-001 `archived → superseded`

AC PASS/FAIL matrix (10/10 PASS):

| AC | Status | Evidence |
|----|--------|----------|
| AC-001 | PASS | `internal/mcp/` removed (`test ! -d` → REMOVED) |
| AC-002 | PASS | `internal/cli/mcp.go` removed; `moai --help` shows no `mcp` subcommand |
| AC-003 | PASS | `.mcp.json.tmpl` moai-lsp count 0; context7/staggeredStartup intact |
| AC-004 | PASS | `make build` ran; embed (`//go:embed all:templates`) consistent (NOCHANGE) |
| AC-005 | PASS | settings_test.go moai-lsp count 0; `go test ./internal/template/...` ok |
| AC-006 | PASS | settings-management.md mirror pair moai-lsp count 0; `TestRuleTemplateMirrorDrift` PASS |
| AC-007 | PASS | SPEC-LSPMCP-001 `status: superseded` + `superseded_by` set |
| AC-008 | PASS | `grep internal/mcp` import → no matches; both builds exit 0 |
| AC-009 | PASS | `TestMCPTemplate*` render + json.Unmarshal PASS (both template branches) |
| AC-010 | PASS | `git diff --quiet -- internal/lsp/` → PRESERVED (10 consumers untouched) |

- **Cross-platform build**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- **PRESERVE guard (AC-010)**: `internal/lsp/` byte-unchanged — ralph/mx/loop/hook/deps/lsp_doctor consumers intact
- **golangci-lint**: 0 issues
- **Full test suite**: 93 packages ok; 2 pre-existing flaky tests in `internal/hook/wrapper_test.go` (`signal: killed` under parallel contention; unrelated to this SPEC — STABILIZE-003 candidate; pass in isolation)
- **Dangling reference check**: `grep internal/mcp|moai mcp lsp|mcpCmd|mcpLSPCmd internal/ cmd/` → zero

Status: run-phase complete (`status: in-progress`). Next: sync-phase (in-progress → implemented + CHANGELOG + sync-auditor).
