---
id: SPEC-V3R6-CI-FLAKY-STABILIZE-001
title: "CI Flaky Test Stabilization — internal/spec race + Windows merge-TUI hang"
version: "0.2.0"
status: implemented
created: 2026-05-31
updated: 2026-05-31
author: manager-spec
priority: P0
phase: "v3.0.0"
module: "internal/spec, internal/cli, internal/merge"
lifecycle: spec-anchored
tags: "ci, flaky-test, race-condition, windows, bubbletea, stabilization"
tier: M
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-31 | manager-spec | Initial plan-phase draft (3-artifact LEAN). Two CI-only flaky failures diagnosed: FLAKY-1 (internal/spec git-add race under -race parallel), FLAKY-2 (Windows merge-TUI ReadConsole hang → 600s timeout). |

## §A — Context and Problem Statement

`origin/main` CI is red. Two CI-only flaky test failures are the Go-test-infrastructure subset of the residual main-RED debt. Neither was introduced by recent work — both predate it and only manifest on specific CI runner configurations. They block `go test -race ./...` from passing deterministically across the ubuntu / macos / windows CI matrix.

This SPEC stabilizes both so the Go test matrix passes reliably. A sibling SPEC (`SPEC-V3R6-DOCS-I18N-PARITY-001`) owns the docs-site i18n CI failure separately; that concern is explicitly out of scope here.

### FLAKY-1 — internal/spec git-add race (ubuntu + macos CI race-detector only)

The atomic-close transaction `performAtomicClose()` (`internal/spec/closer.go`) shells out to `git add <absolutePath>` with `cmd.Dir = baseDir`. When the package's many sibling tests run in parallel (15 `t.Parallel()` siblings observed in `closer_test.go`), git's resolution of an absolute path against a `cmd.Dir` working directory hits a filesystem-timing race that surfaces only under the CI race-detector on ubuntu/macos. `baseDir` additionally defaults to the relative `"."` (`closer.go` `Close()`), compounding inconsistent path resolution. Locally the package runs serially-enough to never manifest — `go test -race ./internal/spec/ -run TestClose_FullClose_ProducesCommit -count=5` passes clean.

The targeted test (`TestClose_FullClose_ProducesCommit`) is intentionally NOT `t.Parallel()` because the close transaction is atomic; that property MUST be preserved.

### FLAKY-2 — Windows internal/cli 600s timeout + internal/harness failures

On `windows-latest` CI, the `internal/cli` package hits a 600s wall-clock timeout (one test hangs) and `internal/harness` tests fail. The hang originates in `merge.ConfirmMerge()` (`internal/merge/confirm.go`), which calls `tea.NewProgram(m).Run()` (Bubble Tea TUI). On a Windows non-TTY CI stdin, Bubble Tea blocks indefinitely on the Windows console `ReadConsole` syscall. A test reaches `ConfirmMerge` via `runTemplateSyncWithProgress` (`internal/cli/update.go`) when `autoConfirm` (the `--yes` flag) is false and a version mismatch forces the merge path.

Evidence: a sibling test `TestRunTemplateSyncWithProgress_ForceFlagBypassesVersionCheck` (`internal/cli/target_coverage_test.go`) ALREADY carries a Windows skip with the reason "charmbracelet/bubbletea blocks on Windows console ReadConsole in non-TTY CI" — confirming the pattern and the existing-precedent skip idiom.

Windows cannot be fully reproduced locally (no Windows runner in this environment); see §C and acceptance.md for the honest verifiability boundary.

## §B — GEARS Requirements

### FLAKY-1 — internal/spec path-resolution stabilization

**REQ-CFS-001** (Event-driven): When `performAtomicClose` stages a SPEC's `spec.md` / `progress.md` via `git add`, the close orchestrator shall pass each path resolved RELATIVE to `baseDir` (via `filepath.Rel`) so git resolves the path consistently repo-root-relative regardless of the `cmd.Dir` filesystem-timing environment.

**REQ-CFS-002** (Ubiquitous): The close orchestrator shall normalize `baseDir` to an absolute path (via `filepath.Abs`) at the start of `Close()` so downstream `git add` / `git restore --staged` / `git commit` operations never depend on a relative `"."` default.

**REQ-CFS-003** (Ubiquitous): The `git restore --staged` rollback path in `performAtomicClose` shall use the same relative-path resolution as the staging path so rollback unstages exactly the paths that were staged.

**REQ-CFS-004** (State-driven): While `TestClose_FullClose_ProducesCommit` exercises the full close transaction, the test shall remain non-parallel (no `t.Parallel()` call) because the close transaction is atomic and serialization is the documented intent.

**REQ-CFS-005** (Event-driven): When the test suite runs `go test -race ./internal/spec/ -count=20`, the close orchestrator shall produce zero race-detector findings and zero test failures.

### FLAKY-2 — Windows merge-TUI hang prevention

**REQ-CFS-006** (Event-driven): When a test invokes `runTemplateSyncWithProgress` (or any path reaching `merge.ConfirmMerge`) without guaranteed `--yes`/auto-confirm propagation on a Windows runner, the test shall skip via `if runtime.GOOS == "windows" { t.Skip(...) }` carrying a reason that names the bubbletea/ReadConsole non-TTY hang.

**REQ-CFS-007** (Capability gate): Where the runtime is Windows and stdin is not a terminal, `merge.ConfirmMerge` shall fail open with an error (rather than calling `tea.NewProgram().Run()` and blocking on `ReadConsole`) before entering the Bubble Tea event loop.

**REQ-CFS-008** (Ubiquitous): The `merge.ConfirmMerge` TTY guard shall reuse the already-vendored `github.com/mattn/go-isatty` dependency and the established `isatty.IsTerminal(os.Stdin.Fd())` idiom (as used in `internal/cli/update.go` and `init.go`) — no new dependency shall be introduced.

**REQ-CFS-009** (Event-driven): When the test author audits `runTemplateSyncWithProgress` callers in `internal/cli/coverage_improvement_test.go`, the audit shall identify every caller that reaches `ConfirmMerge` without `--yes` set and add the Windows skip to each such caller.

### Green gate

**REQ-CFS-010** (Event-driven): When `go test -race ./...` runs locally (darwin), the test suite shall pass deterministically with zero failures, establishing the necessary (but not sufficient) local gate.

**REQ-CFS-011** (Ubiquitous): The Windows CI job completing under the timeout shall be the true completion gate for FLAKY-2; the SPEC shall NOT claim FLAKY-2 closed on local evidence alone.

## §C — Exclusions (What NOT to Build)

### Out of Scope — docs-site i18n CI failure
The docs-site i18n parity CI failure is owned by the sibling `SPEC-V3R6-DOCS-I18N-PARITY-001`. This SPEC MUST NOT modify `docs-site/` or any i18n locale files, and MUST NOT entangle the docs-i18n failure into its green gate.

### Out of Scope — internal/template mirror-drift failures
The `internal/template` mirror-drift and neutrality CI failures are already addressed by `SPEC-V3R6-MAIN-RED-REMEDIATION-001`. This SPEC MUST NOT touch `internal/template/templates/` or the mirror-drift test fixtures.

### Out of Scope — broader flaky-test refactor
The 3 pre-existing flaky tests' broader refactor (test-isolation overhaul, fixture restructuring beyond the two diagnosed failures) is out of scope. This SPEC fixes ONLY the two CI-only failures diagnosed in §A.

### Out of Scope — Bubble Tea / charmbracelet upgrade
Upgrading `charmbracelet/bubbletea` or replacing the merge-confirm TUI with a non-TUI prompt is out of scope. The fix is a minimal TTY guard + test skips, not a TUI rewrite.

### Out of Scope — implementation details
Per SPEC scope discipline, this document specifies observable behaviors and constraints, not exact function signatures, helper-function names, or line-level diffs. Those are decided at run-phase.
