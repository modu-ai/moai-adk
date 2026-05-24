# SPEC-V3R5-ATOMIC-WRITE-001 — Lifecycle Progress

## Session Summary

- **SPEC**: SPEC-V3R5-ATOMIC-WRITE-001 (Tier S, P0, v2.20.0-rc1 release blocker 4/8)
- **cycle_type**: tdd (per quality.yaml development_mode)
- **Branch**: main (1인 OSS Hybrid Trunk, direct push per CLAUDE.local.md §23)
- **Status**: implemented (v0.2.0)
- **Lifecycle commits** (main):
  - `e9da15706` plan(SPEC-V3R5-ATOMIC-WRITE-001): P0-4 atomicWrite 안전성 위반 정정 SPEC (Tier S, v0.1.0)
  - `5fccdb7a6` + `c25a718f5` run-phase 구현 (cherry-pick source)
  - `d49a0a7db` sync(SPEC-V3R5-ATOMIC-WRITE-001): CHANGELOG entry + status implemented (v0.2.0)

## Implementation Evidence (verified 2026-05-25)

### REQ-AWR-001..006 Implementation

`internal/migrate/hook_cleanup.go:124-164` — `atomicWrite(path string, data []byte, perm os.FileMode) error`:

- Sequence: `os.CreateTemp(dir, ".hook_cleanup_tmp_*")` → `tmp.Write(data)` → `tmp.Sync()` → `tmp.Close()` → `os.Rename(tmpName, path)` → `os.Chmod(path, perm)` (REQ-AWR-001 ✓)
- Signature preserved: `(path string, data []byte, perm os.FileMode) error` (REQ-AWR-002 ✓)
- Error path: each step wraps error with step name + `os.Remove(tmpName)` best-effort cleanup (REQ-AWR-003 ✓)
- Explicit `os.Chmod(path, perm)` post-rename preserves caller-supplied 0o644 vs CreateTemp default 0o600 (REQ-AWR-004 ✓)
- Naming-semantics match restored — no `os.WriteFile` primitive (REQ-AWR-005 ✓)
- `@MX:WARN` + `@MX:REASON` annotation (P0-4 reference) at L126-130 per mx-tag-protocol.md mandatory-fields (REQ-AWR-006 ✓)

### Regression Test Coverage

`internal/migrate/hook_cleanup_test.go` (613 LOC) — 6 SPEC-V3R5-ATOMIC-WRITE-001 regression tests at L344-540:

| Test | Scenario | AC |
|------|----------|----|
| TestAtomicWrite_HappyPath | Scenario 1: fresh path write | AC-AWR-001, AC-AWR-005 |
| TestAtomicWrite_ReplaceExisting | Scenario 2: replace existing | AC-AWR-001 |
| TestAtomicWrite_PreservesPerm | Scenario 5: perm preservation post-CreateTemp 0o600 default | AC-AWR-005 |
| TestAtomicWrite_CleanupOnError | Scenarios 3+4: rename error → temp cleanup | AC-AWR-003 |
| TestAtomicWrite_CreateTempError | CreateTemp failure path | AC-AWR-003 |
| TestCleanupUserSettings_RegressionAfterAtomicWriteFix | Caller-level (`hook_cleanup.go:109` settings.json + `:121` archive) regression | AC-AWR-002 |

### AC Binary Matrix (sync verdict, d49a0a7db)

| AC | Status | Verification |
|----|--------|--------------|
| AC-AWR-001 | PASS | tmp+rename sequence (TestAtomicWrite_HappyPath PASS) |
| AC-AWR-002 | PASS | Signature preserved + 2 callers compile (TestCleanupUserSettings_RegressionAfterAtomicWriteFix PASS) |
| AC-AWR-003 | PASS | Error path cleanup (TestAtomicWrite_CleanupOnError + CreateTempError PASS) |
| AC-AWR-004 | PASS | Explicit os.Chmod post-rename (TestAtomicWrite_PreservesPerm PASS) |
| AC-AWR-005 | PASS | Naming-semantics restored (no os.WriteFile primitive) |
| AC-AWR-006 | PASS | @MX:WARN + @MX:REASON L126-130 |
| AC-AWR-007 | CONDITIONAL | Package coverage 77.6%; `atomicWrite` function 59.1% > runtime/persist.go::atomicWrite 50.0% baseline by +9.1pp. Deeper structural coverage requires deferred `pkg/atomicio` mocking framework per EXCL-AWR-003 Tier L. Sync judged CONDITIONAL PASS — coverage exceeds reference baseline; structural mocking deferred. |
| AC-AWR-008 | PASS | Cross-platform PASS (darwin native + GOOS=windows amd64 + GOOS=linux amd64); race detector PASS; NEW lint=0 |

7/8 binary PASS + 1 CONDITIONAL (AC-AWR-007 documented EXCL-AWR-003 Tier L deferral).

### Cross-Platform Verification

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
$ GOOS=linux GOARCH=amd64 go build ./...   → exit 0
$ go test -race ./internal/migrate/...     → PASS
```

### MX Tag Status (Mx Step C judgment)

`internal/migrate/hook_cleanup.go::atomicWrite` carries `@MX:WARN` + `@MX:REASON` per REQ-AWR-006. No additional MX annotations required — single function, single caller-graph fan_in, no goroutines, no complexity ≥ 15. Step C judge: **EVALUATE-PASS** per mx-tag-protocol.md §a.

## Frontmatter Drift Backfill Note (2026-05-25)

Observed during Sprint 10 cleanup: `spec.md` frontmatter had reverted to `version: "0.1.0" / status: draft / updated: 2026-05-20` despite the canonical sync commit `d49a0a7db` (2026-05-21) advancing frontmatter to `version: "0.2.0" / status: implemented / updated: 2026-05-21` and adding the v0.2.0 HISTORY row. Root cause likely the `720a636b5 chore(main-sync): 23 parallel-session + lifecycle cleanup commits (2026-05-22) (#1037)` mass-merge — conflict resolution silently kept the pre-sync version of `spec.md`. Yesterday's batch cleanup `3b7ce18b6` (22 archived + 4 implemented frontmatter drift 정정) audited 39 missing-implementation SPECs but did NOT cover this SPEC (status: draft on the surface, while git log showed sync commit + production code).

Backfill action (this chore, `chore(specs): backfill 2 implemented frontmatter drift (ATOMIC-WRITE-001 + HOOK-CONTRACT-FIX-001)`):
- `spec.md` frontmatter restored to canonical sync-time values (v0.2.0 / implemented / 2026-05-21)
- HISTORY v0.2.0 row restored verbatim from `git show d49a0a7db:.moai/specs/SPEC-V3R5-ATOMIC-WRITE-001/spec.md`
- `progress.md` newly authored (sync commit `d49a0a7db` did not include progress.md — Tier S minimal sync scope was 2 files only: CHANGELOG.md + spec.md)
- CHANGELOG.md ledger entry (this backfill) added to `[Unreleased]` § Changed

L48 SSOT canary: `spec.md` body absolutely untouched — only frontmatter + HISTORY table modified (manager-docs scope per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix).

## Cross-references

- Source SPEC: `.moai/specs/SPEC-V3R5-ATOMIC-WRITE-001/spec.md`
- Implementation: `internal/migrate/hook_cleanup.go:124-164`
- Tests: `internal/migrate/hook_cleanup_test.go:344-540` (6 TestAtomicWrite_* + caller regression)
- Reference pattern: `internal/runtime/persist.go:62-85`
- Sync commit: `d49a0a7db` (2026-05-21)
- Plan commit: `e9da15706` (2026-05-20)
- Defect source: `.moai/reports/review-v214-to-HEAD.md` L53-58 (P0-4 v2.20.0-rc1 release blocker)
- Related SPEC: SPEC-V3R5-SECURITY-CRIT-001 (PR #1032 — P0-1/P0-2/P0-3 fixes, merged 2026-05-20T13:05:12Z)
- Pattern doctrine: `.claude/rules/moai/workflow/mx-tag-protocol.md` § Mandatory Fields
- Tier S policy: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier
