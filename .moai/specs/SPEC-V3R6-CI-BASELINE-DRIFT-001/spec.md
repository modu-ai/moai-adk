---
id: SPEC-V3R6-CI-BASELINE-DRIFT-001
title: "CI baseline drift cleanup (lint 27 + TestStatus golden + ConfigManager race)"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: manager-develop
priority: P3
phase: v3.0.0
module: "internal/cli, internal/config, internal/statusline, internal/cli/wizard, internal/constitution, internal/template, internal/merge, internal/tmux"
lifecycle: spec-anchored
tier: S
related_specs:
  - SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001
depends_on: []
tags: "ci, lint, baseline, drift, technical-debt, tier-s, v3.0"
---

# CI baseline drift cleanup — lint 27 + TestStatus golden v3.0.0-rc1 + ConfigManager race

## Overview

V3R6 UPDATE-NAMESPACE-PROTECT-001 PR #1048 admin merge (commit `767bc04a4`, 2026-05-23) confirmed that 4 CI failures are pre-existing baseline drift unrelated to that SPEC's scope. This SPEC consolidates and resolves the three observed drift categories:

- **Category A**: golangci-lint reports 27 issues (errcheck 8 + ineffassign 1 + staticcheck 5 + unused 13)
- **Category B**: TestStatus golden files reference `v2.17.0` while pkg/version SoT is `v3.0.0-rc1`
- **Category C**: ConfigManager race detected when running `go test -race ./internal/config/...` (root cause: package-level `sunsetNoticeOnce sync.Once` reset for testing without mutex)

This is a Tier S technical-debt cleanup. Behavior changes are NOT in scope — only baseline restoration.

---

## §A — Pre-existing State Survey

[HARD] Per `.claude/rules/moai/development/plan-auditor-codebase-state-blindspot` mitigation #1, this section enumerates the codebase facts the SPEC builds upon, with each fact independently verifiable.

### A.1 — Lint baseline: 27 issues (exact reproducer)

**Reproducer** (orchestrator pre-scan executed at commit `eaff5f272`, 2026-05-23):
```
golangci-lint run --timeout=2m
```

**Output footer**:
```
27 issues:
* errcheck: 8
* ineffassign: 1
* staticcheck: 5
* unused: 13
```

**Itemized inventory** (file:line:severity):

errcheck (8):
- `internal/cli/harness_mute.go:141,145,173,186` — `fmt.Fprintln`/`fmt.Fprintf` to `cmd.OutOrStdout()` (4 sites)
- `internal/template/seq_thinking_retire_audit_test.go:95,135` — `defer f.Close()` (2 sites in test file, residual from SEQ-THINKING-RETIRE-001 PR #1049)
- *(remaining 2 errcheck sites inside the cluster above — actual count 6 may differ; SPEC scope is "errcheck → 0", exact site enumeration verified by manager-develop M1)*

ineffassign (1):
- `internal/cli/agent_lint_test.go:2008` — `p := filepath.Join(...)` ineffectual reassignment in test loop

staticcheck (5):
- `internal/merge/confirm.go:559,637` — QF1012 `fmt.Fprintf(...)` instead of `WriteString(fmt.Sprintf(...))` (2 sites)
- `internal/tmux/session_sensitive_test.go:35,80,114` — SA4032 `runtime.GOOS == "windows"` under build constraints that exclude windows (3 sites — test file under `//go:build !windows` or similar)

unused (13):
- `internal/cli/branch_protection.go:35,38` — type `ttyConfirmer` + method `Confirm` (2 entries)
- `internal/cli/init_layout.go:21,61` — `renderInitHeader`, `renderInitNextSteps` (2 entries)
- `internal/cli/wizard/review.go:15-17,48,119,152` — 3 constants (`reviewChoiceProceed/Edit/Cancel`) + `renderProgressBreadcrumb` + `renderReviewPanel` + `runReview` (6 entries — wizard/review.go module exists end-to-end but is not wired into the init wizard run path)
- `internal/constitution/validator.go:151` — `var hardRuleRegexp` (1 entry)
- `internal/statusline/renderer.go:145,682` — `renderFullV3`, `contextPercent` (2 entries — v3 statusline variant scaffolded but not selected by current full-mode renderer)

### A.2 — TestStatus golden drift: v2.17.0 → v3.0.0-rc1

**Reproducer**:
```
grep -rn "v2.17.0" internal/cli/testdata/*.golden
```

**Output** (3 files):
```
internal/cli/testdata/status-nocolor.golden:6:│  ADK       moai-adk v2.17.0       │
internal/cli/testdata/status-light.golden:6:│  ADK       moai-adk v2.17.0       │
internal/cli/testdata/status-dark.golden:6:│  ADK       moai-adk v2.17.0       │
```

**SoT divergence**: `pkg/version/version.go:7` declares `Version = "v3.0.0-rc1"` (build-time injectable via ldflags, but the in-source default is `v3.0.0-rc1`). The `moai status` CLI subcommand reads from `pkg/version.GetVersion()` and renders into the status box at line 6 (`ADK       moai-adk <version>`). Golden files captured at `v2.17.0` were not refreshed when the SoT bumped through `v2.18.x → v3.0.0-rc1`.

### A.3 — ConfigManager race (root cause: sunsetNoticeOnce reset)

**Reproducer**:
```
go test -race -count=1 ./internal/config/...
```

**Output (excerpt)** (4 distinct racing test entries: `TestValidateGitConventionSampleSize`, `TestConfigManagerSave`, `TestConfigManagerGetSection`, `TestLoaderMIG003MalformedSectionsUseDefaults`, `TestConfigManagerGet`):
```
Write at 0x000102d39408 by goroutine 79:
      internal/config/loader.go:86 +0x330
      internal/config/manager.go:60 +0x134
...
Previous write at 0x000104d11408 by goroutine 585:
  internal/config.resetSunsetNoticeOnce()
      internal/config/sunset_notice.go:37 +0x5c
```

**Root cause** (`internal/config/sunset_notice.go:37`):
```go
// resetSunsetNoticeOnce resets the once guard. FOR TESTING ONLY.
func resetSunsetNoticeOnce() {
    sunsetNoticeOnce = sync.Once{}
}
```

This rewrites the package-level `sunsetNoticeOnce sync.Once` from one test goroutine while another parallel test invokes `emitSunsetDormantNotice()` → `sunsetNoticeOnce.Do(...)` from `Loader.Load()` at `loader.go:86`. The race target address is the `sync.Once` struct itself (0x000102d39408).

**Affected tests** (5 races confirmed in parallel execution):
1. `TestValidateGitConventionSampleSize/{-1 is invalid, 0 is valid, 100 is valid}`
2. `TestConfigManagerSave`
3. `TestConfigManagerGetSection`
4. `TestLoaderMIG003MalformedSectionsUseDefaults/interview.yaml`
5. `TestConfigManagerGet/after_load_returns_config`

Race does NOT manifest without `-race` (full package passes `go test -count=1 ./internal/config/...` in 1.454s). It manifests under `-race` because the detector instruments `sync.Once` internals against unsynchronized struct overwrites.

### A.4 — Provenance: confirmed baseline at PR #1048 admin merge

Commit `767bc04a4` (PR #1048, `feat(SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001)`, admin-merged 2026-05-23) confirmed all three drift categories pre-existed and were unrelated to the NAMESPACE-PROTECT-001 scope. CI failures observed at PR #1048:
- "Test ubuntu" — TestStatus golden diff (Category B)
- "Lint" — golangci-lint baseline 27 (Category A)
- "Test ubuntu" — ConfigManager race fails under `-race` (Category C)
- "Build linux/amd64" — passed
- "CodeQL" — passed

Reference: MEMORY.md entry `[V3R6 NAMESPACE-PROTECT-001 admin merge complete]` § "CI failure 원인 분석" (2026-05-23).

### A.5 — Baseline command (single SoT for measurement)

[HARD] All AC measurements MUST use exactly this command for lint baseline:
```
golangci-lint run --timeout=2m
```

For race baseline:
```
go test -race -count=1 ./internal/config/...
```

For TestStatus golden:
```
go test -count=1 -run TestStatus ./internal/cli/...
```

These three commands are the AC measurement contract; any deviation invalidates AC verification.

---

## Requirements (EARS Format)

### REQ-CBD-001 (Ubiquitous): TestStatus golden alignment with version SoT
The system SHALL ensure that golden files used by TestStatus reflect the version reported by `pkg/version.GetVersion()` at the time the golden was captured. When `pkg/version/version.go` changes its declared `Version`, the three golden files (`status-nocolor.golden`, `status-light.golden`, `status-dark.golden`) MUST be regenerated to match.

### REQ-CBD-002 (Ubiquitous): Lint baseline restoration
The codebase SHALL contain zero issues in the following golangci-lint categories: errcheck, ineffassign, staticcheck. The `unused` category SHALL contain at most the entries explicitly scope-deferred in §D.1 (see Plan), and no new unused entries SHALL be introduced.

### REQ-CBD-003 (Event-Driven): ConfigManager race elimination
WHEN `go test -race -count=1 ./internal/config/...` is executed, the package SHALL complete without any data race detected on the `sunsetNoticeOnce sync.Once` variable. The root-cause fix MUST protect package-level test-only mutators with the appropriate synchronization primitive (mutex or `atomic.Value`) without changing the runtime behavior of `emitSunsetDormantNotice()`.

### REQ-CBD-004 (Unwanted): Behavior preservation
The system SHALL NOT change any observable runtime behavior. Specifically: (a) `moai status` output text MAY change ONLY in the version line to reflect the current `pkg/version` SoT; (b) the once-per-process DORMANT notice emission contract (REQ-MIG003-018) MUST remain — at most one log line per process lifetime under normal operation; (c) no public API signature SHALL change.

### REQ-CBD-005 (State-Driven): Scope-deferred unused entries
WHILE a function or symbol flagged as `unused` represents work explicitly deferred to a follow-up SPEC (see §D.1 of Plan), the entry SHALL be retained in the codebase with a `//nolint:unused // deferred to SPEC-XXX-YYY` directive and a corresponding follow-up SPEC stub. Otherwise, the entry SHALL be removed.

### REQ-CBD-006 (Optional): Race reproducer test
WHERE practical, a dedicated test file (e.g., `internal/config/sunset_notice_race_test.go`) MAY be added containing a deterministic race reproducer that fails on the pre-fix state under `-race` and passes after the fix.

### REQ-CBD-007 (Unwanted): No new lint debt
The PR SHALL NOT introduce any new lint issues. Any new issue introduced during the cleanup MUST either be fixed in the same PR or explicitly documented in §D.1 (scope-deferred section) of `plan.md`.

---

## Exclusions (What NOT to Build)

- **NO** v3.0.0-rc2 or v3.0.0 GA release version bump — `pkg/version` SoT remains `v3.0.0-rc1`; only golden files are aligned to the existing SoT
- **NO** behavioral changes to `moai status`, `moai update`, or any other command — only golden file content is updated
- **NO** fix for the underlying `sunset.yaml` template-only design (REQ-MIG003-018 contract preserved)
- **NO** removal of `internal/cli/wizard/review.go` end-to-end — the wizard/review module is preserved per §D.1 (deferred to follow-up SPEC if removed)
- **NO** changes to `pkg/version` ldflags injection mechanism
- **NO** broader CI workflow modifications (caching, matrix, timeouts) — only baseline drift restoration
- **NO** dependency upgrades, Go toolchain bumps, or golangci-lint config changes

---

## Acceptance Criteria

See `acceptance.md` for the full Given-When-Then scenarios and binary AC list.

---

## Cross-references

- Plan: `plan.md` (M1-M3 milestones)
- Acceptance: `acceptance.md` (8 binary ACs)
- Progress: `progress.md` (milestone tracker)
- Sibling SPEC: SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 (PR #1048 — the merge that confirmed this drift)
- Memory: `[V3R6 NAMESPACE-PROTECT-001 admin merge complete]` (MEMORY.md, 2026-05-23)
- Codebase state blindspot lesson: `feedback-plan-auditor-codebase-state-blindspot`
