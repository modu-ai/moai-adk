---
id: SPEC-V3R6-I18N-VALIDATOR-BUDGET-001
title: "i18n-validator TestBudget Threshold 30s → 35s — progress"
version: "1.0.0"
status: draft
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

# Progress — SPEC-V3R6-I18N-VALIDATOR-BUDGET-001

## Plan-phase Evidence

| Item | Status | Note |
|------|--------|------|
| spec.md created | DONE | manager-spec authoring 2026-05-24 (101 lines, ~10KB after fix-forward) |
| plan.md created | DONE | 119 lines, 6296B — edit map for 4 lines in 1 file (M1 single milestone) |
| acceptance.md created | DONE | 109 lines, 5243B — 5 ACs (AC-IVB-001 through AC-IVB-005) |
| progress.md scaffold | DONE | this file, 54 lines, 2392B |
| Cascade grep verified | DONE | auditor independent: 0 code invocations, 5 archival .md refs (out of scope per §A.3), 7 self-refs in own SPEC artifacts excluded |
| plan-auditor iter-1 | PASS 0.87 | Tier S threshold 0.80 +0.07 margin, MP-2 EARS PASS, 0 BLOCKING, 5 SHOULD-FIX, 3 NICE-TO-HAVE |
| Orchestrator-direct fix-forward (S1+S2+S5) | DONE | 7 edits: 4 frontmatter (S1+S2 canonical schema: created_at→created, updated_at→updated, labels→tags CSV, +title+phase+module+lifecycle, Low→P3, related_specs→depends_on YAML list) + spec.md §A.4 table row 49 + §A.5 line 60 edit count 3→4 (S5) + §A.3 heading "Out-of-scope"→"Out of Scope" + body consistency. spec-lint final: ✓ zero findings |
| Plan commit on main | TBD | next step (Hybrid Trunk 1-person OSS direct push per CLAUDE.local.md §23.7) |

## Run-phase Evidence

| AC | Status | Command / Output | Note |
|----|--------|------------------|------|
| AC-IVB-001 | TBD | `grep -n "if elapsed > 35\*time.Second" scripts/i18n-validator/main_test.go` | Budget threshold at line 376 |
| AC-IVB-002 | TBD | `grep -n "TestBudget_FullRepoScanWithin35Sec" scripts/i18n-validator/main_test.go` | Function renamed + JP comment updated |
| AC-IVB-003 | TBD | `go test -timeout 60s -v -run TestBudget_FullRepoScanWithin35Sec ./scripts/i18n-validator/...` | Renamed test passes |
| AC-IVB-004 | TBD | elapsed = `<X.YYs>` recorded here | Risk-E1 warning if ≥ 33.00s |
| AC-IVB-005 | TBD | `go test -timeout 90s -count=1 ./scripts/i18n-validator/...` | No regression in non-budget tests |

## Sync-phase Evidence

| Item | Status | Note |
|------|--------|------|
| CHANGELOG `[Unreleased]` `### Changed` entry | TBD | single-line SPEC ID + brief per Tier S minimal |
| B12 standing-rule guard self-test | TBD | 7th consecutive PASS expected |
| Frontmatter status draft → completed | TBD | all 4 artifacts (spec/plan/acceptance/progress) |

## Mx-phase Evidence

| Item | Status | Note |
|------|--------|------|
| Step C SKIP justification | TBD | 0 Go production .go files modified; only test file `main_test.go`; no @MX:ANCHOR/WARN/NOTE/TODO trigger; existing @MX tags (if any) preserved verbatim per §a exemption pattern |

## Exemptions / Carry-over

- None expected. All scope edits confined to `scripts/i18n-validator/main_test.go` lines 359-377.
- LCL-001 AC-LCL-005 PASS-WITH-DEBT cleared by this SPEC's completion.
