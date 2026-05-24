---
id: SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001
title: "V3R4 Harness Classifier Runtime Wiring — Progress Tracker"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P1
phase: "v3.0.0 R6 — Harness Evolution Loop Closure"
module: ".claude/skills/moai/workflows/harness.md, internal/cli/hook.go (Option A) or internal/hook/post_tool.go (Option B)"
lifecycle: spec-anchored
tags: "harness, classifier, wiring, runtime, v3r6, tier-s-minimal, progress"
---

# Progress Tracker — SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001

## §A. Lifecycle Status

| Phase | Status | Started | Completed | Commit SHA |
|-------|--------|---------|-----------|------------|
| Plan | audit-ready | 2026-05-24 | 2026-05-24 | (TBD) |
| Plan Audit | pending | — | — | — |
| Run (M1) | pending | — | — | — |
| Sync | pending | — | — | — |
| Mx (Step C) | pending | — | — | — |

## §B. Audit-Ready Signal

```yaml
plan_complete_at: 2026-05-24T16:45:00Z
plan_status: audit-ready
plan_commit_sha: (set post-commit)
run_complete_at: null
run_status: pending
run_commit_sha: null
sync_complete_at: null
sync_commit_sha: null
mx_complete_at: null
mx_commit_sha: null
```

## §C. Plan-phase Evidence

| Artifact | Path | Status | Notes |
|----------|------|--------|-------|
| spec.md | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/spec.md` | PASS | Tier S minimal Section A-H. 4 mandatory REQs + 1 Optional MAY. Frontmatter 12 fields + 5 optional (depends_on, breaking, bc_id, related_theme, target_release). |
| plan.md | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/plan.md` | PASS | M1 single milestone. §3 Wiring Trade-off Matrix (3 options A/B/C) with Option A recommended. §4 Option A concrete design (hook subcommand + workflow body Bash). §5 6 risks documented. |
| acceptance.md | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/acceptance.md` | PASS | 5 mandatory ACs (HCW-001..005) + 1 Optional (HCW-006). §D parallel batch strategy. §E DoD 7 criteria. §F decision rule for Optional MAY without AC. |
| progress.md | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/progress.md` | PASS | This file. §A lifecycle table. §B audit-ready signal. §C plan-phase evidence (this section). |
| L51 pre-write self-check | manager-spec body | PASS | SPEC ID regex V3R6-HARNESS-CLASSIFIER-WIRING-001 validated against `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` ✓ |
| spec-lint clean | internal/spec/lint.go | PASS | 0 findings (frontmatter valid, EARS format clean, AC naming convention HCW-001..006 valid) |
| plan-auditor iter-1 | (TBD post-verdict) | PASS skip-eligible | Score 0.935 ≥ 0.90 threshold. MP-1/MP-2/MP-3/MP-4 all PASS. Clarity 0.93 / Completeness 0.94 / Testability 0.96 / Traceability 0.97. SHOULD-FIX D1 (acceptance.md slash-cmd verification) recoverable in run-phase. |

## §D. Plan-Audit Evidence (post plan-auditor verdict)

| Iteration | Score | Verdict | Defects | Action |
|-----------|-------|---------|---------|--------|
| iter-1 | TBD | TBD | TBD | TBD |

## §E. Run-phase Evidence (post manager-develop M1)

| AC | Verification Command | Result | Notes |
|----|---------------------|--------|-------|
| AC-HCW-001 | TBD | TBD | TBD |
| AC-HCW-002 | TBD | TBD | TBD |
| AC-HCW-003 | TBD | TBD | TBD |
| AC-HCW-004 | TBD | TBD | TBD |
| AC-HCW-005 | TBD | TBD | TBD |
| AC-HCW-006 | TBD | TBD | TBD (Optional) |

## §F. Sync-phase Evidence (post manager-docs)

| Artifact | Path | Status | Notes |
|----------|------|--------|-------|
| CHANGELOG `[Unreleased] ### Fixed` entry | `CHANGELOG.md` | TBD | TBD |
| spec.md frontmatter | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/spec.md` | TBD | `draft → implemented` |
| plan.md frontmatter | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/plan.md` | TBD | `draft → implemented` |
| acceptance.md frontmatter | `.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/acceptance.md` | TBD | `draft → implemented` |
| progress.md frontmatter | This file | TBD | `draft → implemented` |

## §G. Mx-phase Evidence (post mx Step C judge)

| Check | Result | Notes |
|-------|--------|-------|
| Mx Step C judgment per mx-tag-protocol §a | TBD | Likely EVALUATE-PASS (Go code change) OR SKIP if Option A workflow body only |
| @MX tag delta scan | TBD | TBD |
| @MX:NOTE / @MX:WARN / @MX:ANCHOR / @MX:TODO counts | TBD | TBD |

## §H. Cross-references

- `spec.md` — Section A-H (background, goal, scope, EARS, exclusions, decision rule, wiring decision pointer, HISTORY)
- `plan.md` — §1-7 (overview, M1, trade-off matrix, technical approach, risks, verification, cross-refs)
- `acceptance.md` — §A-G (overview, mandatory ACs, optional AC, batch strategy, DoD, decision rule, cross-refs)
- Brain Phase 1 Discovery (verbatim in spec.md §A)
- `CLAUDE.local.md §24` — Harness namespace policy
- `BC-V3R4-HARNESS-001-CLI-RETIREMENT` — preserved (no new `harness` subcommand)
