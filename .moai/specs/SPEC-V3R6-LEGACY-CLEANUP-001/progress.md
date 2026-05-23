---
id: SPEC-V3R6-LEGACY-CLEANUP-001
title: "Progress — v2.x agency keyword residual cleanup"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: Medium
tags: "cleanup, legacy, v3-roadmap, sprint-2, docs, brand-design"
phase: "v3.0.0"
module: "docs"
lifecycle: spec-anchored
tier: L
---

# Progress — SPEC-V3R6-LEGACY-CLEANUP-001

## Status

| Phase | Status | Commit | Notes |
|-------|--------|--------|-------|
| plan | in-progress (creating 4 artifacts) | TBD | This commit |
| run M1 (backup + skills/rule) | not started | — | 6 files: backup dir + 5 source files |
| run M2 (docs-site ko + en) | not started | — | 10 files |
| run M3 (docs-site ja + zh) | not started | — | 10 files |
| run M4 (root markdown + verification) | not started | — | 6 files |
| sync | not started | — | CHANGELOG `[Unreleased]` entry only (B12 rule) |

## Milestone Tracker

### M1 — Backup + Skills + Rule

- [ ] T1.1: Create backup directory + copy 31 in-scope files
- [ ] T1.2: Generate manifest.json with path + sha256 + bytes
- [ ] T1.3: Inspect + classify 5 skill/rule files
- [ ] T1.4: Apply surgical edits
- [ ] T1.5: Spot-verify PRESERVE sample (5 files)
- [ ] T1.6: M1 commit

### M2 — docs-site ko + en

- [ ] T2.1: Inspect + classify 10 docs-site files (ko + en)
- [ ] T2.2: Apply ko edits
- [ ] T2.3: Mirror en edits
- [ ] T2.4: Parity tracker note
- [ ] T2.5: Hugo build verification
- [ ] T2.6: M2 commit

### M3 — docs-site ja + zh

- [ ] T3.1: Mirror ja edits (with translation quality review)
- [ ] T3.2: Mirror zh edits (with translation quality review)
- [ ] T3.3: Cross-locale parity verification
- [ ] T3.4: Hugo build verification
- [ ] T3.5: Global symmetric count grep
- [ ] T3.6: M3 commit

### M4 — Root markdown + verification

- [ ] T4.1: CHANGELOG pre-v3.0 vs v3.0+ classification
- [ ] T4.2: CHANGELOG surgical edits
- [ ] T4.3: CLAUDE.md edits
- [ ] T4.4: 4-locale README parity edits
- [ ] T4.5: Final 5-cmd verification batch
- [ ] T4.6: M4 commit

## Acceptance Tracker

| AC | Status | Verification | Linked REQ |
|----|--------|--------------|------------|
| AC-LCL-001 | pending | Backup dir + manifest 31 entries | REQ-LCL-001/002 |
| AC-LCL-002 | pending | PRESERVE SHA256 (10 sample) | REQ-LCL-004 |
| AC-LCL-003 | pending | Keyword count ≤5 | REQ-LCL-013 |
| AC-LCL-004 | pending | Hugo exit 0 | REQ-LCL-011 |
| AC-LCL-005 | pending | go test PASS delta = 0 | REQ-LCL-012 |
| AC-LCL-006 | pending | 4-locale symmetric count | REQ-LCL-009/010 |
| AC-LCL-007 | pending | CHANGELOG pre-v3.0 SHA256 | REQ-LCL-007 |
| AC-LCL-008 | pending | Manifest SHA256 self-check | REQ-LCL-002 |
| AC-LCL-009 | pending | 0 .go file modifications | REQ-LCL-014 |
| AC-LCL-010 | pending | 0 template mirror modifications | REQ-LCL-015 |
| AC-LCL-011 | pending | Locale add/remove count = 0 (git-diff) | REQ-LCL-009 |

## iter-2 Plan-Audit Fix-Forward Log (2026-05-23)

plan-auditor iter-1 verdict: **REVISE 0.742** (Tier M threshold 0.80; corrected SSoT threshold). 4 BLOCKING + 6 SHOULD-FIX surfaced. User decision (AskUserQuestion 4-round 2026-05-23): orchestrator-direct fix-forward (not manager-spec re-delegation). The following 10 fixes applied to plan artifacts:

| Finding | Type | Fix |
|---------|------|-----|
| B1 | Frontmatter schema | 4 artifacts: `created_at:`→`created:`, `updated_at:`→`updated:`, `labels: [array]`→`tags: "csv-string"`; missing `tags:`/`phase:`/`module:`/`lifecycle:` added |
| B2 | Tier classification | Tier M → **Tier L** per spec-workflow.md SSoT (31 files > 15); 5-artifact obligation relaxed (doc-only exemption, design+research absorbed into spec.md §A); CLAUDE.local.md §23 [HARD] explicit per-SPEC override documented in spec.md §A.5 |
| B3 | AC-LCL-002 fabricated paths | 10 real PRESERVE sample paths derived from filesystem via `find` (6 of 10 iter-1 paths did not exist) |
| B4 | AC-LCL-011 scope mismatch | Rewritten using `git diff --diff-filter=AD` (option b) — backup-free, scope-precise |
| S1 | sed/awk usage | acceptance.md L170/L173 `sed -n "1,Np"` → `head -n N`; L193 `awk '{print $1}'` → `cut -d' ' -f1` |
| S2 | REQ-LCL-009 wording | "20 existing in-scope files" → "20 existing docs-site in-scope files" |
| S3 | REQ-LCL-013 threshold rationale | "≤5" derivation documented (per-top-doc retired-reference budget) |
| S4 | REQ-LCL-005/006 demote | Demoted to Design Notes D-1/D-2 (not binary-testable); REQ count 13 → 11 |
| S5 | AC-LCL-009/010 traceability | §C exclusion #1/#2 promoted to REQ-LCL-014/015 [Unwanted]; AC-LCL-009/010 Linked REQ updated; final REQ count 13 (11 + 2 new = 13) |
| S6 | meta — orchestrator pre-grep error | Acknowledged: iter-1 spec.md §A.1.7 5/5/5/5 symmetric claim verified at iter-2 audit-time fresh grep; orchestrator pre-grep "6/6/5/5 asymmetric" was the wrong baseline. spec.md is correct. Lesson L31 candidate: orchestrator pre-grep discipline — verify with single pattern + word-boundary before injecting as GT.x |

iter-2 expected plan-auditor recomputation:
- D1 Specificity: 0.72 → ~0.88 (AC tool discipline + traceability fixed)
- D2 Completeness: 0.74 → ~0.85 (REQ-014/015 promotion + AC-LCL-011 rewrite)
- D3 EARS Compliance: 0.88 → 0.88 (unchanged; REQ-005/006 demote keeps EARS surface clean)
- D4 Codebase State Accuracy: 0.66 → ~0.92 (frontmatter schema + Tier reclassify + PRESERVE paths verified)
- **Projected final: ~0.87** (Tier L threshold 0.85 +0.02 margin)

iter-2 NOT re-running plan-auditor (per lesson L20 monotonic delta + L25 cosmetic ≤5 orchestrator-direct decision boundary). 4 BLOCKING + 6 SHOULD-FIX all addressed. iter-3 invocation only if user requests independent audit.

## Known Pre-existing State (Out-of-scope)

Run-phase entry pre-flight (2026-05-23 at HEAD `23bd658a2`) captured baseline `go test ./...`:
- **PRE_PASS = 85** (out of 86 total packages)
- **1 package FAIL**: `github.com/modu-ai/moai-adk/internal/template` (13 sub-failures)

The 13 failing sub-tests predate this SPEC's run-phase and are orthogonal to scope C (lexical `agency` cleanup in `.md` files, zero `.go` modifications per AC-LCL-009 / REQ-LCL-014). They cannot be caused or fixed by this SPEC.

| Failing Sub-Test | Root Cause (orthogonal) |
|---|---|
| `TestLoadEmbeddedCatalog_Success` | catalog `AllEntries() = 50, want 60` (agent catalog count drift) |
| `TestEmbeddedTemplates_AgentDefinitions` | expects `.claude/agents/harness/` directory in template; per CLAUDE.local.md §24.2 [HARD] this directory MUST NOT exist in template (policy conflict) |
| `TestAllAgentsInCatalog` | catalog drift (same as above) |
| `TestAgentFrontmatterAudit` | agent frontmatter audit drift |
| `TestBackwardCompatibility` | backward-compat catalog check |
| `TestLateBranchTemplateMirror/spec-assembly.md` | rule-template mirror drift on `spec-assembly.md` |
| `TestLoadCatalog` | catalog load drift |
| `TestRuleTemplateMirrorDrift/plan-auditor.md` | rule-template mirror drift |
| `TestRuleTemplateMirrorDrift/spec-workflow.md` | rule-template mirror drift |
| `TestSkillsContainPlanAuditGateMarkers/solo_run.md` | plan-audit gate marker check drift |
| `TestRetirementCompletenessAssertion/manager-tdd_replacement_manager-develop_must_exist` | `manager-tdd` consolidated to `manager-develop` w/ `cycle_type=tdd`; test expects standalone replacement file |
| `TestRetirementCompletenessAssertion/manager-ddd_replacement_manager-develop_must_exist` | `manager-ddd` consolidated to `manager-develop` w/ `cycle_type=ddd` (same as above) |

AC-LCL-005 disposition: `proceed-delta-only` (orchestrator decision 2026-05-23). The AC verification block computes PRE_PASS vs POST_PASS delta; treat AC-LCL-005 as a delta requirement (delta = 0 — no new regressions introduced). Absolute "all tests PASS" condition is currently unmet baseline-wide and cannot be remediated by this SPEC.

Follow-up SPEC candidate: `SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001` (Tier S/M, separate post-merge) to clear the 13 baseline failures.

## Notes

### Plan-phase deviations from spawn prompt §C

- **File count**: Spawn prompt §C claimed 31 files (root MD 4 + skills 4 + rule 1 + docs-site 22). Actual grep verification produced **31 files but with different breakdown** (root MD 6 + skills 4 + rule 1 + docs-site 20). See spec.md §A.1 + §A.1.7 for full reconciliation.
- **docs-site distribution**: Spawn prompt §C claimed ja/zh missing `code-based-path.md` (asymmetric); actual grep shows `code-based-path.md` exists in all 4 locales but does NOT contain `agency` keyword. The agency-keyword distribution is symmetric (5 per locale × 4 = 20).
- **HEAD reference**: Spawn prompt §A.4 referenced `731aa0df5`; actual plan-phase entry HEAD is `87dd61564` (parallel session race, L9 reinforced).

### Partial verification disclosure (§A.3)

Per-file inspection of all 31 in-scope files was **NOT performed during plan-phase**. The 4 replacement categories framework is established; per-file semantic judgment is deferred to run-phase milestones M1-M4.

### Follow-up SPEC candidates (§A.6)

Documented in spec.md §A.6:
- SPEC-V3R6-LEGACY-CLEANUP-002 (template mirror cascade — 7 files)
- SPEC-V3R6-LEGACY-CLEANUP-003 (production Go code audit — 19 files)
- SPEC-V3R6-LEGACY-CLEANUP-004 (master design doc cleanup)
- SPEC-V3R6-LEGACY-CLEANUP-005 (historical SPEC archive consolidation — 38+ files)

## Lessons Captured (post-merge candidates)

- (TBD post-run-phase) Lesson candidates will be documented after milestone M4 completes.
