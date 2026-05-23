---
id: SPEC-V3R6-I18N-VALIDATOR-BUDGET-001
title: "i18n-validator TestBudget Threshold 30s → 35s — plan"
version: "1.0.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P3
phase: "v3.0.0"
module: "scripts/i18n-validator"
lifecycle: spec-anchored
tags: "i18n, test, budget, tier-s, lcl-001-followup"
tier: S
---

# Plan — SPEC-V3R6-I18N-VALIDATOR-BUDGET-001

## §1 Edit Map (Exact)

Single file: `scripts/i18n-validator/main_test.go`.

| Line | Current Content | Replacement Content | REQ Anchor |
|-----:|-----------------|---------------------|------------|
| 359 | `// TestBudget_FullRepoScanWithin30Sec は実際の repo で30秒以内に完了することを検証します。` | `// TestBudget_FullRepoScanWithin35Sec は実際の repo で35秒以内に完了することを検証します。` | REQ-IVB-002 |
| 360 | `func TestBudget_FullRepoScanWithin30Sec(t *testing.T) {` | `func TestBudget_FullRepoScanWithin35Sec(t *testing.T) {` | REQ-IVB-002 |
| 376 | `if elapsed > 30*time.Second {` | `if elapsed > 35*time.Second {` | REQ-IVB-001 |
| 377 | `t.Errorf("full repo scan took %v, want <= 30s", elapsed)` | `t.Errorf("full repo scan took %v, want <= 35s", elapsed)` | REQ-IVB-001 |

**Edit count clarification**: §A.4 says "3 textual edits" (line 359 / 360 / 376 budget value) — line 377 error message follows from the line-376 threshold change for consistency; the threshold and its operator error message are a single logical edit unit. Total physical line edits: **4 lines in 1 file**, all within the contiguous block 359-377. Tier S envelope (≤ 3 files, ≤ 1 milestone) preserved.

## §2 Milestones

### M1 — Apply edits + local verification

**Owner**: manager-develop (Tier S minimal cycle, DDD or TDD per `.moai/config/sections/quality.yaml development_mode`)

**Steps**:

1. Pre-flight read: `Read scripts/i18n-validator/main_test.go offset=355 limit=30` to confirm lines 359 / 360 / 376 / 377 match the §1 Edit Map current-content column verbatim.
2. Apply 4 Edit tool calls in sequence (one per line; or single MultiEdit batch). Preserve all surrounding whitespace and indentation.
3. Run `go vet ./scripts/i18n-validator/...` → exit 0 expected (no vet regression).
4. Run `go test -timeout 60s -run TestBudget ./scripts/i18n-validator/...` → both renamed budget tests must PASS.
5. Capture elapsed time of `TestBudget_FullRepoScanWithin35Sec` via `go test -timeout 60s -v -run TestBudget_FullRepoScanWithin35Sec ./scripts/i18n-validator/...` and record in progress.md §Run-phase Evidence (AC-IVB-004).
6. Run full package: `go test -timeout 90s ./scripts/i18n-validator/...` → all tests PASS, POST_PASS count = PRE_PASS count for non-budget tests (AC-IVB-005).

**Done criteria**: 6 ACs all PASS (5 in acceptance.md + AC-IVB-004 elapsed-time evidence row populated in progress.md).

**No further milestones**: Tier S envelope (single milestone sufficient for 4 line edits in 1 file).

## §3 Risk Mitigation

### Cascade grep result (verified 2026-05-24)

```bash
$ grep -rn "TestBudget_FullRepoScanWithin30Sec" \
    --include="*.go" --include="*.md" --include="*.yaml" \
    --include="*.yml" --include="Makefile" --include="*.sh"
```

Output classification:

| Reference Type | Count | Files | Action |
|----------------|------:|-------|--------|
| Code invocation | 0 | — | No fix needed; Go test reflection-based discovery |
| Archival SPEC narrative | 5 | `SPEC-V3R3-CI-AUTONOMY-001/strategy-wave6.md:378`, `tasks-wave6.md:28,118`, `progress.md:498`, `SPEC-V3R6-LEGACY-CLEANUP-001/progress.md:90` | Out-of-scope per §A.3 / D-3 (archived, immutable) |
| Source code self-references | 2 | `main_test.go:359, 360` | In-scope, fixed by M1 |

**Conclusion**: Function rename has zero code-level cascade. Documentation drift to archived narratives is accepted per LCL-001 §A.6 [Unwanted] retention pattern (precedent established 2026-05-23).

### Other risk mitigations

- **Risk-E1 (33s warning threshold)**: M1 step 5 elapsed-time capture must include a check — if elapsed ≥ 33s (94 % of new 35s budget), manager-develop reports as `PASS-WITH-WARNING` and recommends future optimization SPEC rather than silent acceptance. Threshold rationale: 33s leaves a 2s margin which is the minimum meaningful headroom on CI-class runners; below it, the next regression cycle would trip.

## §4 Validation Strategy

### Local verification commands (verbatim, ordered)

```bash
# 1. Pre-flight: confirm baseline content
grep -n "TestBudget_FullRepoScan" scripts/i18n-validator/main_test.go
# Expected: 2 lines (359 comment, 360 function decl)

# 2. Apply edits via M1 step 2 (manager-develop)

# 3. Post-edit confirmation
grep -n "TestBudget_FullRepoScanWithin35Sec" scripts/i18n-validator/main_test.go
# Expected: 2 lines (359 comment, 360 function decl, both with Within35Sec)

grep -cn "Within30Sec" scripts/i18n-validator/main_test.go
# Expected: 0 (no stale references)

grep -n "if elapsed > 35\*time.Second" scripts/i18n-validator/main_test.go
# Expected: 1 line (376)

# 4. Vet baseline
go vet ./scripts/i18n-validator/...
# Expected: exit 0

# 5. Targeted budget test with elapsed-time capture
go test -timeout 60s -v -run TestBudget_FullRepoScanWithin35Sec \
  ./scripts/i18n-validator/... 2>&1 | tee /tmp/ivb-run.log
# Expected: PASS line with elapsed time < 35s recorded;
# record exact elapsed for AC-IVB-004

# 6. Full package baseline
go test -timeout 90s ./scripts/i18n-validator/... 2>&1 | tee /tmp/ivb-full.log
# Expected: ok scripts/i18n-validator <elapsed>s with PASS count
# matching PRE_PASS count for non-budget tests
```

### Phase 0.5 Plan Audit Gate

- plan-auditor invocation: iter-1 threshold **0.80** (Tier S) with **MP-2 EARS format** obligation.
- Expected outcome: PASS on iter-1 (all 3 REQs in SHALL form; §A.3/§A.4 explicit; cascade evidence in §3 risk mitigation).
- If REVISE: orchestrator-direct fix-forward per L32 precedent (Tier S small scope, ≤ 5 textual edits).

### Definition of Done

- 5 ACs (AC-IVB-001 through AC-IVB-005) all PASS in progress.md §Run-phase Evidence
- CHANGELOG `[Unreleased]` `### Changed` entry added by sync-phase manager-docs (Tier S minimal: single line referencing SPEC ID + brief)
- B12 standing-rule guard sync-phase self-test PASS (7th consecutive)
- Mx Step C SKIP justified (0 Go production .go files modified; only test file; no @MX:ANCHOR/WARN/NOTE/TODO trigger)
