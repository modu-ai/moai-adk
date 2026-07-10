---
id: SPEC-CLIFIX-CONTRACT-001
title: "CLI Contract Remediation — ExitCoder adoption, dead flags, exit-code/stream contracts (P1)"
version: "0.1.0"
status: draft
created: 2026-07-10
updated: 2026-07-10
author: manager-spec
priority: P1
phase: "v3.0.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "cli, audit-remediation, exit-codes, contract-drift, p1"
era: V3R6
tier: M
depends_on: [SPEC-CLIFIX-CRITICAL-001]
---

# SPEC-CLIFIX-CONTRACT-001 — CLI Contract Remediation (P1)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-10 | manager-spec | Initial draft from CLI audit 2026-07-10 §4 (contract-drift + os.Exit patterns) + §5 P1 roadmap row |

## §A Context

The audit's P1 row targets declaration-vs-implementation drift: 8+ `os.Exit` calls inside RunE/PostRunE bodies bypass the existing `ExitCoder` mechanism at the main.go boundary (defeating defers and making commands untestable), registered flags that no code reads (`github --dry-run`, `spec_status --confirm`), documented exit codes that are never produced (astgrep json/sarif, constitution 2, spec_lint 3, spec_audit 2), pre-push violation details printed to stdout where Claude Code only surfaces stderr on exit 2, and a flock test that never runs because its filename does not end in `_test.go`.

Findings SSOT: audit §3 clusters 3/4/5 (contract rows) + §4 rows "RunE 내 os.Exit" and "exit-code/플래그 문서-구현 불일치". Re-verify all anchors against the live tree at run time.

## §B Requirements (GEARS)

- REQ-CONT-001-001: The CLI shall produce all non-zero exit codes through the existing `ExitCoder` mechanism mapped at the main.go boundary, and RunE/PostRunE bodies shall not call `os.Exit` — removing the calls in hook.go, hook_pre_push.go, astgrep.go, spec_lint.go, spec_drift.go, migrate_agency.go (:594), harness/execute.go (:327), and agentlint/workflow_lint.go (:159).
- REQ-CONT-001-002: When a `moai github` subcommand runs with `--dry-run` (github.go:97,103), the CLI shall perform no registry writes and shall print the planned mutations instead.
- REQ-CONT-001-003: When `moai spec status --sync-git` would mutate git state (spec_status.go:205-235), the CLI shall gate the mutation on the `--confirm` (or `--yes`) flag and shall abort with a diagnostic in a non-TTY context instead of blocking on `fmt.Scanln`.
- REQ-CONT-001-004: When `moai astgrep` emits `--json` or `--sarif` output and the result set contains errors (astgrep.go:107-122), the process shall exit 1 exactly as in text format, with the HasErrors decision evaluated independently of the output-format branch.
- REQ-CONT-001-005: The CLI shall implement the documented exit-code contracts end-to-end: constitution check failure shall exit 2 (constitution.go:296-319 exitCodeError mapped at the Execute boundary), `moai spec lint` invalid arguments shall exit 3, and `moai spec audit` MUST-FIX drift shall exit 2.
- REQ-CONT-001-006: When the pre-push hook blocks a push with exit 2 (hook_pre_push.go:180-198), the CLI shall write the violation details to stderr so the reason is surfaced by Claude Code and git.
- REQ-CONT-001-007: The CLI shall rename `team_spawn_lock_test_unix.go` to `team_spawn_lock_unix_test.go` so the flock test compiles into the test binary and runs, and the `testing` package shall no longer be linked into the production binary.
- REQ-CONT-001-008: The run-phase implementation shall add a contract test per changed command asserting that the exit code declared in help text matches the exit code actually produced.

## §C Scope

### In Scope

- ExitCoder boundary mapping + os.Exit removal at the 8 listed sites; dead-flag wiring (github --dry-run, spec_status --confirm); exit-code contract implementation (astgrep/constitution/spec_lint/spec_audit); pre-push stderr routing; test filename correction; per-command contract tests.

### Out of Scope — Critical data-loss fixes

- The 8 P0 defects (including hook.go tier-promotion growth and migrate_agency rollback) belong to SPEC-CLIFIX-CRITICAL-001, which lands first; this SPEC rebases on its result for the shared files hook.go and migrate_agency.go.

### Out of Scope — New exit-code taxonomy

- No new exit codes are introduced and no help text is rewritten beyond aligning declaration with implementation. Designing a unified CLI-wide exit-code scheme is future work.

### Out of Scope — Concurrency and hygiene work

- Locked-writer consolidation (SPEC-CLIFIX-CONCURRENCY-001) and dead-code/hardcoding sweeps (SPEC-CLIFIX-HYGIENE-001) are excluded, including inside files this SPEC touches.

## §D Acceptance Criteria

- AC-CONT-001-001: Given the internal/cli tree after implementation, When grepping RunE/PostRunE bodies for os.Exit, Then zero call sites remain outside main.go boundary mapping and cmd wiring approved exceptions (maps REQ-CONT-001-001)
- AC-CONT-001-002: Given a registry-mutating github subcommand invoked with --dry-run, When it completes, Then the registry file is byte-identical and the planned mutation is printed (maps REQ-CONT-001-002)
- AC-CONT-001-003: Given a non-TTY invocation of spec status --sync-git without --confirm, When it runs, Then it aborts with a diagnostic and performs no git mutation (maps REQ-CONT-001-003)
- AC-CONT-001-004: Given an astgrep scan whose findings include errors, When run with --json and with --sarif, Then both invocations exit 1 (maps REQ-CONT-001-004)
- AC-CONT-001-005: Given a failing constitution check, an invalid spec lint argument set, and a MUST-FIX spec audit drift fixture, When each command runs, Then exit codes are 2, 3, and 2 respectively (maps REQ-CONT-001-005)
- AC-CONT-001-006: Given a pre-push violation, When the hook blocks with exit 2, Then the violation detail text appears on stderr and not solely on stdout (maps REQ-CONT-001-006)
- AC-CONT-001-007: Given the renamed test file, When `go test ./internal/cli/ -run Flock` executes on unix, Then the flock test is discovered and runs, and the production binary no longer imports testing (maps REQ-CONT-001-007)
- AC-CONT-001-008: Given the contract test suite, When it runs, Then every changed command's declared exit codes are asserted against produced exit codes (maps REQ-CONT-001-008)

Machine-verifiable commands and expected outcomes per AC: see `acceptance.md` (§D AC Matrix).

## §E Non-Goals and Dependencies

- Dependencies: SPEC-CLIFIX-CRITICAL-001 (shared files hook.go, migrate_agency.go, team_spawn* — P0 lands first; this SPEC rebases on it).
- Non-goal: changing what conditions constitute a violation for any command — only how the verdict is communicated (exit code, stream) is corrected.
