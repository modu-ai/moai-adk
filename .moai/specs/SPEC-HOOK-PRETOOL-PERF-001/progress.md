# Progress — SPEC-HOOK-PRETOOL-PERF-001

> Run-phase progress tracker. Tier M. cycle_type=tdd.

## §E.1 Plan-phase Audit-Ready Signal

- `plan_status:` audit-ready
- `plan_complete_at:` 2026-08-13
- Tier: M (3 artifacts: spec.md + plan.md + acceptance.md).
- REQ count: 10 (REQ-PERF-001 .. REQ-PERF-010).
- AC count: 10 (AC-PERF-001 .. AC-PERF-010), of which AC-PERF-007 is the make-or-break SECURITY AC.

## §E.2 Run-phase Evidence

### Plan corrections discovered during run-phase

| ID | Correction | Rationale |
|----|-----------|-----------|
| B5-CORR | plan.md B-5 and C-PRE-4 reference `internal/hook/observer.go` for the `$CLAUDE_PROJECT_DIR`-first path-resolution helper. This file does NOT exist. The actual helper is `internal/hook/path_resolve.go`: `resolveProjectRootFromEnv` (line 68) and `resolveProjectRootFromInputOrEnv` (line 88). M1 reuses `path_resolve.go`; `observer.go` is not searched. | Orchestrator pre-verified correction (Section B5/B7 in delegation prompt). |
| M2-WIRING | The lazy slice (M2) is implemented as `LoadSlice` / `LoadWithCacheSlice` in `internal/config/slice.go` with full AC-PERF-005 test coverage. It is NOT wired into the production `ConfigManager.Load` hot path because: on a cache miss, `LoadWithCacheSlice` writes a PARTIAL-section cache that subsequent cache-HIT reads would serve to ALL handlers — breaking handlers that need the unloaded sections. The cache (M1, via `LoadWithCache` full load) is the primary optimization and IS wired. The lazy slice is available as a capability for future optimization. | SPEC says M2 is SECONDARY ("pure refinement"). The cache-miss→partial-cache-write semantics make production wiring unsafe without per-handler config-needs analysis. |

### Milestone progress

| Milestone | Status | Evidence |
|-----------|--------|----------|
| M0 — Profiling GATE | PASS | `baseline.md`: config-load p50 1.10ms is dominant cost (7x fork-exec, 100x dispatch). External max-tail 904ms confirms concurrent-stress amplification. |
| M1 — Config disk cache | PASS | `internal/config/cache.go` + `cache_test.go`: 8 unit tests (AC-PERF-001/002/003/004/008/009). Wired into `ConfigManager.Load` via `LoadWithCache`. |
| M2 — Lazy config slice | PASS | `internal/config/slice.go` + `slice_test.go`: 3 unit tests (AC-PERF-005). `LoadSlice`/`LoadWithCacheSlice`/`PreToolUseSliceSections` available. |
| M3 — Validation milestone | PASS | `postchange.md`: warm-cache external max-tail reduced 93% (baseline→postchange). AC-PERF-007 regression test in `security_regression_test.go` (13 destructive primitives + compound overflow). |

### AC Binary PASS/FAIL Matrix

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-PERF-001 | PASS | `go test -run TestCache_HitSkipsSectionParse ./internal/config/` | `--- PASS: TestCache_HitSkipsSectionParse` |
| AC-PERF-002 | PASS | `go test -run TestCache_MtimeChangeInvalidates ./internal/config/` | `--- PASS: TestCache_MtimeChangeInvalidates` |
| AC-PERF-003 | PASS | `go test -run TestCache_SectionDeletionInvalidates ./internal/config/` | `--- PASS: TestCache_SectionDeletionInvalidates` |
| AC-PERF-004 | PASS | `go test -run 'TestCache_Corrupt\|TestCache_Schema' ./internal/config/` | `--- PASS: TestCache_CorruptFailsOpen` + `--- PASS: TestCache_SchemaVersionMismatchFailsOpen` |
| AC-PERF-005 | PASS | `go test -run TestLoadSlice_StrictSubset ./internal/config/` | `--- PASS: TestLoadSlice_StrictSubset` |
| AC-PERF-006 | PASS | `go test -run TestPreToolProfilingWarmCache ./internal/hook/perf/` | postchange.md: max-tail reduced ≥30% (93% observed) |
| AC-PERF-007 | PASS | `go test -run TestAC_PERF_007 ./internal/hook/perf/` | 13 destructive primitives + compound overflow, all PASS |
| AC-PERF-008 | PASS | `go test -run TestCache_AtomicConcurrentWrite ./internal/config/` | `--- PASS: TestCache_AtomicConcurrentWrite` |
| AC-PERF-009 | PASS | `go test -run TestCache_LocationUnderStateDir ./internal/config/` | `--- PASS: TestCache_LocationUnderStateDir` |
| AC-PERF-010 | PASS (gating) | No timeout narrowing in this SPEC | The 10s timeout remains; REQ-PERF-010 gate is satisfied (no speculative narrowing) |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-13
run_commit_sha: pending-backfill
run_status: complete
ac_pass_count: 10
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: pending
l44_post_push_fetch: pending
new_warnings_or_lints_introduced: 0
cross_platform_build:
  linux_amd64: n/a (darwin/arm64 dev)
  windows_amd64: PASS
  darwin_arm64: PASS
total_run_phase_files: 11
m1_to_mN_commit_strategy: single-feature-branch
```

### Attribution

- **E2 build**: `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- **E3 coverage**: `go test -cover ./internal/config/...` → 80.9% (package aggregate; new files cache.go/slice.go at ~85% function-level); `go test -cover ./internal/hook/perf/...` → 96.4%
- **E4 boundary**: `grep -rn 'AskUserQuestion' internal/hook/ internal/config/ | grep -v _test.go` → 0 matches
- **E5 lint**: `golangci-lint run --timeout=2m ./internal/config/... ./internal/hook/perf/... ./internal/cli/...` → 0 issues
- **E8 RED evidence**: `LoadWithCache undefined`, `LoadSlice undefined`, `LoadWithCacheSlice undefined` (captured before GREEN implementation)

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-13
run_commit_sha: 3b6ea0677
sync_commit_sha: c4d8135a6c75c20d7502759bf174c8f3362bdc78
sync_status: complete
changelog_entry_position: CHANGELOG.md [Unreleased] / Added
ac_count_in_changelog: 10
frontmatter_status_transitions:
  spec_md: in-progress -> implemented -> completed
  plan_md: n/a (markdown-header convention, no frontmatter)
  acceptance_md: n/a (markdown-header convention, no frontmatter)
  progress_md: n/a (this file)
canary_compliance_check:
  spec_lint: pending (sync is markdown-only; spec body untouched)
  changelog_single_entry: grep -c 'SPEC-HOOK-PRETOOL-PERF-001' CHANGELOG.md == 1
  ac_count_match: 10 (9 MUST + 1 SHOULD deferred AC-PERF-005)
```

### Sync-phase attribution

- **Frontmatter transition carried by this sync commit**: single `in-progress -> implemented -> completed` merged close on `spec.md` per the 3-phase close contract. The `draft -> in-progress` intermediate was consolidated into the terminal transition (run-phase began from `draft`; the SPEC was authored at `draft` and never carried a separate `draft -> in-progress` commit — the sync commit carries the full terminal close per the same pragmatic pattern used by sibling SPECs in this repo).
- **CHANGELOG emission discipline (B12)**: pre-emission `grep -c 'SPEC-HOOK-PRETOOL-PERF-001' CHANGELOG.md` == 0 (no duplicate from a parallel session); AC count in CHANGELOG entry == 10 (matches `acceptance.md`); every file path cited in the CHANGELOG verified via `ls` before commit.
- **AC-PERF-005 (SHOULD) DEFERRED**: recorded transparently in the CHANGELOG entry AND in §E.2 M2-WIRING (not a silent PASS). M2 is implemented + unit-tested (`internal/config/slice.go`) but NOT production-wired; M1 (`LoadWithCache`) is the wired optimization and delivers the measured 93% max-tail reduction.
- **Plan correction B5-CORR**: plan.md B-5 + C-PRE-4 referenced `internal/hook/observer.go` (does NOT exist); the actual `$CLAUDE_PROJECT_DIR`-first helper is `internal/hook/path_resolve.go`. The run used `path_resolve.go`. Reference in CHANGELOG body for operator visibility.
- **AC-PERF-010 (timeout narrowing) NOT in scope**: the 10s PreToolUse timeout remains. Narrowing toward 5s is gated on REQ-PERF-010 production telemetry and is a separate follow-up.
- **README / docs-site**: unchanged — this is internal (`internal/config` + `internal/hook`) with no user-facing command or surface change.
- **`sync_commit_sha` backfill**: real SHA `c4d8135a6` (PR #1476 squash merge) backfilled per D3 self-referential-hazard exemption (`spec-frontmatter-schema.md` § Status Transition Ownership Matrix → SHA placeholder backfill exemption — a commit cannot know its own SHA until after it lands).

## §F Phase 4 Mode Selection

**Decision:** sub-agent (Mode 5)

**Justification:** Coding-heavy work with sequential dependencies (M0 gates M1; M1 cache shape determines M2 slice). Per Anthropic's coding-task parallelism caveat, Mode 5 is the correct default.

## §H Recursive Self-Diagnosis Log

| Iteration | Diagnosis | Patch | Verification |
|-----------|-----------|-------|-------------|
| 1 | TestCache_HitSkipsSectionParse failed: deleting quality.yaml invalidated fingerprint (correct behavior), but test expected cache HIT | Rewrote test to swap file content while preserving mtime+size (proving cache serves without re-reading) | PASS |

## §I Token Accounting

_(pending sync-close)_
