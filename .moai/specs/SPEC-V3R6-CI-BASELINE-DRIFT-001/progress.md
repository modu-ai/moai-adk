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
| Plan-auditor | PENDING | Awaiting orchestrator delegation |
| Run-phase | NOT STARTED | manager-develop (cycle_type=ddd, Tier S minimal) |
| Sync-phase | NOT STARTED | manager-docs |

## Milestone Tracker

| Milestone | Description | Status | Verification | AC covered |
|-----------|-------------|--------|--------------|------------|
| M1 | TestStatus golden v2.17.0 → v3.0.0-rc1 (3 files) | NOT STARTED | `go test -count=1 -run TestStatus ./internal/cli/...` | AC-CBD-001 |
| M2.1 | errcheck 8 → 0 | NOT STARTED | `golangci-lint run --timeout=2m \| grep errcheck` | AC-CBD-002 |
| M2.2 | ineffassign 1 → 0 | NOT STARTED | `golangci-lint run --timeout=2m \| grep ineffassign` | AC-CBD-003 |
| M2.3 | staticcheck 5 → 0 | NOT STARTED | `golangci-lint run --timeout=2m \| grep staticcheck` | AC-CBD-004 |
| M2.4 | unused 13 → ≤ scope-deferred (with nolint directives) | NOT STARTED | `golangci-lint run --timeout=2m \| grep unused` | AC-CBD-005 |
| M3 | ConfigManager race fix (sunset_notice.go mutex) | NOT STARTED | `go test -race -count=1 ./internal/config/...` | AC-CBD-006, AC-CBD-007 |
| Final | Full suite under -race | NOT STARTED | `go test -race -count=1 ./...` | AC-CBD-008 |

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

## Notable Implementation Decisions (TBD by manager-develop in run-phase)

To be populated by manager-develop:
- Final errcheck count (initial estimate 8, may differ after re-survey)
- M2.4 final defer list (per-symbol decisions for `branch_protection.go` and `init_layout.go`)
- M3 design decision rationale (Option 1 vs 2 vs 3)
- Follow-up SPEC stubs created (SPEC IDs)
