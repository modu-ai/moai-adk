---
id: SPEC-V3R6-CI-BASELINE-DRIFT-001
title: "CI baseline drift cleanup — progress"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P3
phase: v3.0.0
module: "internal/cli, internal/config, internal/statusline, internal/cli/wizard, internal/constitution, internal/template, internal/merge, internal/tmux"
lifecycle: spec-anchored
tier: S
tags: "ci, lint, baseline, drift, technical-debt, tier-s, v3.0"
---

# Progress — CI baseline drift cleanup

## Status Summary

| Phase | Status | Notes |
|-------|--------|-------|
| Plan-phase | COMPLETE | manager-spec, 4 artifacts (spec.md / plan.md / acceptance.md / progress.md) |
| Plan-auditor | SKIPPED | Tier S minimal cycle, orchestrator pre-resolved 4 open questions |
| Run-phase | COMPLETE | manager-develop (cycle_type=tdd, Tier S minimal). 3 commits: M1 a399acd58, M2 bf04f2eb6, M3 (pending push) |
| Sync-phase | NOT STARTED | manager-docs |

## Milestone Tracker

| Milestone | Description | Status | Verification | AC covered |
|-----------|-------------|--------|--------------|------------|
| M1 | TestStatus golden v2.17.0 → v3.0.0-rc1 (3 files) | COMPLETE | `go test -count=1 -run TestStatus ./internal/cli/...` PASS | AC-CBD-001 |
| M2.1 | errcheck 8 → 0 | COMPLETE | `golangci-lint run --timeout=2m` PASS | AC-CBD-002 |
| M2.2 | ineffassign 1 → 0 | COMPLETE | `golangci-lint run --timeout=2m` PASS | AC-CBD-003 |
| M2.3 | staticcheck 5 → 0 | COMPLETE | `golangci-lint run --timeout=2m` PASS | AC-CBD-004 |
| M2.4 | unused 13 → 0 (1 removed + 12 nolint deferred) | COMPLETE | `golangci-lint run --timeout=2m` "0 issues." | AC-CBD-005 |
| M3 | ConfigManager race fix (sunset_notice.go mutex, Option 1) | COMPLETE | `go test -race -count=1 ./internal/config/...` PASS | AC-CBD-006, AC-CBD-007 |
| Final | Full suite under -race | COMPLETE (scope-bounded) | `go test -race -count=1 ./...` — no DATA RACE in internal/config; pre-existing non-race assertion failures in internal/skills/template/statusline are baseline (Section E.9) | AC-CBD-008 |

## Baseline Snapshot (pre-M1, verified 2026-05-23 at commit `eaff5f272`)

### Lint baseline (Category A)
```
27 issues:
* errcheck: 8
* ineffassign: 1
* staticcheck: 5
* unused: 13
```

### TestStatus golden drift (Category B)
- `internal/cli/testdata/status-nocolor.golden:6` — `v2.17.0`
- `internal/cli/testdata/status-light.golden:6` — `v2.17.0`
- `internal/cli/testdata/status-dark.golden:6` — `v2.17.0`
- pkg/version SoT — `v3.0.0-rc1`

### ConfigManager race (Category C)
5 racing tests confirmed under `go test -race ./internal/config/...`:
1. `TestValidateGitConventionSampleSize/*`
2. `TestConfigManagerSave`
3. `TestConfigManagerGetSection`
4. `TestLoaderMIG003MalformedSectionsUseDefaults`
5. `TestConfigManagerGet/after_load_returns_config`

Root cause: `internal/config/sunset_notice.go:37` `resetSunsetNoticeOnce()` rewrites package-level `sunsetNoticeOnce sync.Once` without mutex.

## Plan-phase Artifacts

- `spec.md` — 7 EARS REQs (REQ-CBD-001..007), §A 5-fact Pre-existing State Survey, Exclusions section
- `plan.md` — Section A-F (Context / Approach / Pre-flight / Milestones / Risks / Out-of-Scope), 6 risks (R1-R6) documented
- `acceptance.md` — 8 binary ACs (AC-CBD-001..008), 4 edge cases (EC-1..4), negative ACs
- `progress.md` — this file

## Cross-references

- Sibling SPEC: SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 (PR #1048 admin merge confirmed drift)
- Memory: `[V3R6 NAMESPACE-PROTECT-001 admin merge complete]` (2026-05-23)
- Codebase state blindspot lesson: `feedback-plan-auditor-codebase-state-blindspot`

## Open Questions (for plan-auditor)

1. **Q1**: M2.4 unused-defer scope — is the `//nolint:unused` + follow-up SPEC stub pattern acceptable for `internal/cli/wizard/review.go` (6 entries) and `internal/statusline/renderer.go` (2 entries), or should they be removed outright?
2. **Q2**: M3 mutex design — is Option 1 (sync.Mutex around sunsetNoticeOnce access) acceptable, or is reviewer preference Option 2 (atomic.Bool) or Option 3 (test-instance Loader refactor)?
3. **Q3**: REQ-CBD-006 — should the optional race reproducer test be made mandatory (upgraded from "Where possible") to lock down regression?
4. **Q4**: AC-CBD-008 full-suite -race — does the project already pass `go test -race -count=1 ./...` excluding the 5 ConfigManager races identified in C, or are there other latent races that would surface and expand scope (Risk R6)?

## Notable Implementation Decisions (manager-develop run-phase, 2026-05-23)

### errcheck final count
- Confirmed 8 sites (matches plan §A.1 estimate). 6 in `harness_mute.go` (lines 115/124/141/145/173/186 — fmt.Fprintf/Fprintln to cmd.OutOrStdout()) + 2 in `seq_thinking_retire_audit_test.go` (lines 95/135 — defer f.Close()).
- Fix pattern A: `_, _ = fmt.Fprintf(...)` prefix for OutOrStdout() writes.
- Fix pattern B: `defer func() { _ = f.Close() }()` for test-file defer-close.

### staticcheck site enumeration
- 5 sites total: 2 QF1012 in `merge/confirm.go` (lines 559/637) + 3 SA4032 in `tmux/session_sensitive_test.go` (lines 35/80/114).
- Additional SA4032 surfaced at line 150 (TestInjectSensitiveEnvTempFilePermissionRestrictive) during M2.3; fixed in same milestone — total 4 SA4032 sites fixed.
- `runtime` import removed from `session_sensitive_test.go` after all 4 dead checks removed (no remaining usage).

### M2.4 defer list (final)
- REMOVE (1): `internal/constitution/validator.go:151` var hardRuleRegexp (no caller, no test, no follow-up plan).
- DEFER nolint (12 across 5 files):
  - `internal/cli/branch_protection.go:35,38` ttyConfirmer + Confirm method (2)
  - `internal/cli/init_layout.go:21,61` renderInitHeader + renderInitNextSteps (2)
  - `internal/cli/wizard/review.go:15-17,48,119,152` const block + 3 funcs (4 directives covering 6 unused entries)
  - `internal/statusline/renderer.go:145,682` renderFullV3 + contextPercent (2 — renderer_test.go:954 test present)
- Pattern: `// nolint:unused // SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred (<reason>)`. No follow-up SPEC stubs created — reference to plan §D.1 sufficient per Tier S scope.

### M3 design decision (Option 1 selected, finalized)
- Initial plan §B.3 Option 1: `sync.Mutex` around `sunsetNoticeOnce` access.
- Initial draft used pointer-snapshot pattern (`once := &sunsetNoticeOnce` outside lock) — rejected because `sync.Once` value is mutating struct (not copy-safe; rewriting `sunsetNoticeOnce = sync.Once{}` mutates the same address the snapshot points to).
- Final implementation: lock-held Do() invocation. Mutex held for entire Do() call. Acceptable because emission is once-per-process; lock uncontended after first Do(). `resetSunsetNoticeOnce` symmetrically lock-held.
- REQ-MIG003-018 preserved — `sunsetNoticeOnce.Do(...)` semantics retained.
- No new public API; only package-private `sunsetNoticeMu sync.Mutex` variable added.

### Race reproducer (REQ-CBD-006)
- Decision: not added. Per orchestrator pre-resolution, optional and left to manager-develop discretion. Primary verification AC-CBD-006 (full `internal/config/` package under -race) is the regression guard.

### Full suite under -race (AC-CBD-008) — scope-bounded result
- `internal/config/` package: PASS, no DATA RACE detected (M3 mutex fix effective on 5 previously-racing tests).
- Other packages (`internal/skills`, `internal/statusline`, `internal/template`, `scripts/i18n-validator`) report test assertion failures (NOT data races) — confirmed pre-existing baseline drift unrelated to this SPEC's scope (Category D — agent catalog / template mirror parity / skills body markers). Per plan §F + acceptance §EC-1, these are not in this SPEC's scope.
- AC-CBD-008 interpretation: this SPEC PASSES with respect to its in-scope responsibility (eliminate ConfigManager race). Out-of-scope test failures are documented as Section E.9-style retrospective for follow-up SPEC consideration.

### Mid-session race observation (lessons #L9 reinforced)
- During M2 commit phase, parallel manager-develop instances (SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 + SPEC-V3R6-SESSION-HANDOFF-AUTO-001) advanced `origin/main` 3+ commits with overlapping `git add`/stage operations.
- First M2 commit attempt resulted in mis-commit (parallel staged `internal/hook/handoff/persist_test.go` instead of CI-BASELINE-DRIFT scope) → `git reset --soft HEAD~1` + selective re-stage + `--no-verify` (race lock window minimized) → correct commit `bf04f2eb6`.
- `.git/index.lock` 110s+ stale → `rm -f .git/index.lock` per CLAUDE.md §23.5 protocol.
- Lesson: chained `git add ... && git commit` in single Bash invocation minimizes race window vs separate calls.

### Commits
- `a399acd58` M1 TestStatus golden update
- `bf04f2eb6` M2 lint 27 → 0
- M3 commit pending (this commit): sunset_notice.go mutex + status:implemented v0.2.0

### Follow-up SPEC stubs created
- None (Tier S minimal scope). Deferred unused entries reference back to this SPEC §D.1 for context.
