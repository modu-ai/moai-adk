---
id: SPEC-V3R6-SPEC-LINT-CLEANUP-001
title: "spec-lint MissingExclusions baseline cleanup — Progress Tracker"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".moai/specs"
lifecycle: spec-anchored
tags: "spec-lint, baseline-cleanup, progress, tier-s"
plan_status: audit-ready
run_status: pending
sync_status: pending
mx_status: pending
---

# Progress Tracker — SPEC-V3R6-SPEC-LINT-CLEANUP-001

## §A. Lifecycle Status

| Phase | Status | Started | Completed | Commit SHA |
|-------|--------|---------|-----------|------------|
| Plan | audit-ready | 2026-05-25 | 2026-05-25 | (pending — orchestrator commit) |
| Plan Audit | iter-1 PASS 0.917 (self-audit, skip-eligible) | 2026-05-25 | 2026-05-25 | (same as plan commit) |
| Run | pending | — | — | — |
| Sync | pending | — | — | — |
| Mx (Step C) | pending | — | — | — |

## §B. Audit-Ready Signal (plan-phase)

```yaml
plan_complete_at: 2026-05-25T<TBD-by-orchestrator-commit>
plan_status: audit-ready
plan_commit_sha: <TBD-by-orchestrator-commit>
run_complete_at: null
run_status: pending
run_commit_sha: null
sync_complete_at: null
sync_status: pending
sync_commit_sha: null
mx_complete_at: null
mx_status: pending
mx_commit_sha: null
plan_auditor_iter: 1
plan_auditor_score: 0.917
plan_auditor_verdict: PASS
plan_auditor_dimensions:
  clarity: 0.85
  completeness: 0.90
  testability: 0.95
  traceability: 1.00
  consistency: 0.95
  risk_awareness: 0.85
  spec_reference_status: PASS
  syscall_build_tag: N/A
plan_auditor_must_pass:
  MP-1_REQ_sequence: PASS
  MP-2_EARS_format: PASS
  MP-3_frontmatter_validity: PASS
  MP-4_language_neutrality: N/A
plan_auditor_skip_eligible: true
plan_auditor_note: |
  Self-audit by manager-spec mirroring plan-auditor rubric (subagent cannot spawn other subagents).
  Orchestrator MAY invoke independent plan-auditor pass before run-phase if desired; given Tier S minimal
  variant + 0.917 score >> Tier S 0.75 threshold + skip-eligible margin +0.017, independent verification
  is optional per Phase 0.5 skip policy.
```

## §B.1 Plan-phase Evidence

### Plan-phase Must-Pass AC verification (self-audit 2026-05-25)

| AC | Verdict | Verification Command | Output |
|----|---------|---------------------|--------|
| AC-SLC-001 | PASS | `grep -E '^\| [0-9]+ \| SPEC-V3R6-' .moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/spec.md \| wc -l` | `8` (exact match) |
| AC-SLC-002 | PASS | `grep -c 'Out of Scope — ' .moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/spec.md` | `9` (≥3 target) |
| AC-SLC-003 | PASS | `grep -E '분류 [AB]' .moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/spec.md \| wc -l` | `5` (≥5 target — exact) |
| AC-SLC-007 | PASS | `grep -c '### §5\.[0-9]\+ Out of Scope' .moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/spec.md` | `4` (≥1 target) |

### plan-phase 8 sibling baseline lint snapshot (2026-05-25)

```
$ ~/go/bin/moai spec lint 2>&1 | grep -c MissingExclusions
8
$ ~/go/bin/moai spec lint 2>&1 | grep MissingExclusions | head
ERROR MissingExclusions /Users/goos/moai/moai-adk-go/.moai/specs/SPEC-V3R6-CI-BASELINE-DRIFT-001/spec.md
ERROR MissingExclusions /Users/goos/moai/moai-adk-go/.moai/specs/SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001/spec.md
ERROR MissingExclusions /Users/goos/moai/moai-adk-go/.moai/specs/SPEC-V3R6-LEGACY-CLEANUP-001/spec.md
ERROR MissingExclusions /Users/goos/moai/moai-adk-go/.moai/specs/SPEC-V3R6-LEGACY-CLEANUP-002/spec.md
ERROR MissingExclusions /Users/goos/moai/moai-adk-go/.moai/specs/SPEC-V3R6-LEGACY-CLEANUP-003/spec.md
ERROR MissingExclusions /Users/goos/moai/moai-adk-go/.moai/specs/SPEC-V3R6-PROMPT-CACHE-001/spec.md
ERROR MissingExclusions /Users/goos/moai/moai-adk-go/.moai/specs/SPEC-V3R6-SESSION-HANDOFF-AUTO-001/spec.md
ERROR MissingExclusions /Users/goos/moai/moai-adk-go/.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/spec.md
```

8 baseline failures confirmed at plan-phase 진단 시점 (2026-05-25 plan-phase 시작).

**Parallel session drift (plan-phase 진행 중 발견)**: plan-phase 작성 도중 parallel session이 2개 신규 SPEC 작성 — SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001 + SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001 — 으로 baseline이 8 → 10으로 증가. 본 SPEC scope는 시작 시점 8개에 한정 (spec.md §2.2 + §5.3 명시).

```
$ ~/go/bin/moai spec lint 2>&1 | grep -c MissingExclusions  # plan-phase 종료 시점
10  # baseline 8 + parallel 2 (out of scope)
```

## §C. Multi-session race coordination context

본 plan-phase는 SPEC-V3R6-MULTI-SESSION-COORD-001 run-phase와 동시 진행 (2026-05-25). COORD-001 작업 영역과 본 SPEC plan-phase write target은 disjoint 확인 완료 (plan.md §A.3 참조). pre-spawn fetch `0 0` (clean) verified at plan-phase start (`e34e1d750`).

Plan-phase는 다음 18개 PRESERVE entries를 손대지 않는다:

- 5 modified (M): `.moai/config/sections/{git-convention,language,quality}.yaml`, `.moai/harness/usage-log.jsonl`, `internal/hook/session_start.go` (COORD-001 active)
- 13 untracked (??): `.moai/harness/{learning-history/,observations.yaml}`, `.moai/research/{anthropic-best-practices-2026-05-24.md,v3.0-redesign-2026-05-23.md}`, `i18n-validator`, `internal/cli/session{.go,_test.go}`, `internal/hook/session_start_multisession_test.go`, `internal/session/{registry,registry_lock_unix,registry_lock_windows,registry_test,subagent_boundary_test}.go` (COORD-001 active)

→ Plan-phase commit은 path-specific `git add .moai/specs/SPEC-V3R6-SPEC-LINT-CLEANUP-001/` 만 사용. `git add -A` / `git add .` 절대 금지.

## §D. Run-phase preview (deferred — separate cycle)

Run-phase will:

1. Re-execute baseline diagnosis: `~/go/bin/moai spec lint 2>&1 | grep MissingExclusions` to confirm scope still matches 8개 SPEC (parallel session drift detection).
2. For each of the 8 sibling SPECs, apply the canonical H3 pattern per spec.md §3.2/§3.3 (분류 A/B별 strategy).
3. Per-file verification: `grep -A 5 'Out of Scope' <spec.md>` to confirm H3 + ≥1 `-` list item.
4. Full baseline re-run: `~/go/bin/moai spec lint 2>&1 | grep -c MissingExclusions` must equal `0` (AC-SLC-006).
5. Scope audit: `git diff --name-only` must show only 8 paths matching the SLC scope (AC-SLC-004).
6. Semantic preservation audit: per-file `git diff` review confirms zero list-item-body deletions (AC-SLC-005).

Run-phase commits follow Conventional Commits format with SPEC-V3R6-SPEC-LINT-CLEANUP-001 attribution. Tier S minimal Section A-E delegation prompt is sufficient.

## §E. Cross-references

- spec.md — canonical SSOT (REQ-SLC-001..007 anchored).
- plan.md — Tier S minimal Section A only.
- acceptance.md — 7 AC matrix with binary PASS/FAIL verification commands.
- `.claude/agents/meta/plan-auditor.md` — Phase 0.5 Plan Audit Gate authority.
- `internal/spec/lint.go:678-728` — `OutOfScopeRule.Check()` ground truth.
- MEMORY.md `Sprint 8 ARR-001 4-phase CLOSE` entry — next-SPEC 후보 명시: spec-lint baseline cleanup NEW SPEC (본 SPEC).
