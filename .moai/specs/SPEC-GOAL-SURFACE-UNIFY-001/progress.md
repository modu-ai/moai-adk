---
id: SPEC-GOAL-SURFACE-UNIFY-001
title: Unify the goal surface on /moai goal and relocate goal presentation to the Implementation Kickoff Approval gate
version: 1.3.0
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: HIGH
phase: "v3.1.0"
module: doctrine
lifecycle: spec-anchored
tags: "goal, doctrine, session-handoff, slash-command, template-mirror"
tier: L
---

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifact set complete (Tier L): `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md`.

- SPEC ID regex check executed as Bash; observed output `SPEC-ID-CHECK: PASS`.
- **30** acceptance criteria retained after the iteration-2 scope reduction; every judgment command executed against the worktree at `origin/main` = `e306e21a9` and its verbatim output recorded as the baseline in `acceptance.md` §B.
- All 30 ACs FAIL at the recorded baseline.
- Milestone ownership verified single-owner across all **37** run-phase paths (`plan.md` §F.1). No sync-phase paths remain.
- Scope spans two layers: doctrine (M1-M6) and Go emission paths (M7, `cycle_type: tdd`). The public-docs layer was split out to `SPEC-GOAL-DOCS-RETIRE-001`.

Five brief corrections surfaced and resolved during authoring, each verified by an executed command (see `research.md`):

| # | Correction | Evidence |
|---|---|---|
| 1 | Doctrine scope is 28 existing files, not 26 — the root `CLAUDE.md` pair was missed by a `.claude/`-scoped grep | `research.md` §A.3 |
| 2 | Four retention surfaces exist, not one — `internal/goal/evaluate.go`'s native-`/goal` yield invariant was mis-classified in the D4 brief as "implementation" | `research.md` §F.2 |
| 3 | The D4 "17 implementation references" figure is not reproducible (nearest pattern gives 35); no AC is built on it | `research.md` §F.3 |
| 4 | Sync-phase affected set is 13 files, not 9 — `advanced/self-evolving.md` (×4) is MoAI-surface primitive naming, not factual contrast | `research.md` §G.1 |
| 5 | `advanced/hooks-reference.md` exists in all four locales, not only en/ko — the gap is a content gap inside existing pages | `research.md` §G.3 |

Two D4 preconditions resolved in the safe direction: `PrimitiveGoal` occupancy measured **0** (hard rename safe, no back-compat needed), and the renderer confirmed to have **no test** (M7 must author the RED gate).

### Plan-audit iteration 1 — FAIL 0.71, remediated

`plan-audit.md` returned FAIL (0.71 against the Tier L 0.85 threshold), forced by MP-3. The auditor re-executed all 29 judgment commands and **every recorded baseline reproduced verbatim, zero divergence**; the defects were structural, not measurement errors. All six MUST-FIX and the material SHOULD-FIX are addressed:

| Finding | Resolution |
|---|---|
| D1 — `tags:` YAML sequence broke SPEC parsing on all six artifacts (MP-3) | Converted to the comma-separated string form. `moai spec lint` now reports **0 errors** on all six (`spec.md` → `✓ No findings`; the other five carry one grandfathered `MissingExclusions` warning each, since Out-of-Scope correctly lives only in `spec.md`) |
| D2 — `plan.md` D3 [HARD] said "two" retention surfaces vs four elsewhere | §A.2 is now the canonical **retention register** with **six** rows; D3 reads "six" and adds a per-AC retention test obligation. `research.md` corrected. `design.md:99`'s different "two surfaces" (the two files owning the W2 rule halves) deliberately left alone |
| D3 — M7 breaks `runner_template_test.go`, owned by no milestone | Added to M7 (now 5 Go files); REQ-GSU-028 + AC-GSU-030 added |
| D4 — AC-GSU-028's blanket `0` would destroy a retention surface | `autonomous-loops.md` reclassified **split**, `autonomous-workflow-strategy.md` **retain + note**; AC-GSU-028 rewritten as emission-markers-positive + 3 retention pins; AC-GSU-032 added. Sweep total 89 → **18 emission markers** |
| D5 — 8 unbackticked refs invisible to every AC | Union detector adopted; AC-GSU-002 rescoped, AC-GSU-031 added over the five no-retention owned files |
| D6 — `moai-meta-harness/SKILL.md` + mirror owned by no milestone | Added to M4 (9 local) and M6 (15 mirrors) |
| S1/S2/S3/S12 | Cross-ref counts, `version: 1.2.0`, canonical **50**-path total, `phase: "v3.1.0"` |
| S4 | B1 corrected: mirror parity is CI-enforced for **2 of 15** pairs, raising AC-GSU-019's importance |
| S5/S6 | AC-GSU-023 rescoped to locale-named subtests; the ordering overclaim dropped and replaced by AC-GSU-033 (RED output recorded in §E.2) |
| S7 | REQ↔AC traceability matrix added at `acceptance.md` §F — **28/28 REQs** cited |
| S9/S10/S11 | `.moai/specs/**` exclusion stated; M4 must re-derive (not transpose) the v2.1.139 availability condition; M1 relocates the comparison-table native row into the prohibition section |

Iteration-2 entry state was 33 ACs / 28 REQs / 6 retention surfaces / 50 paths.

### Plan-audit iteration 2 — FAIL 0.64, STOP; scope reduction executed

Iteration 2 (`plan-audit-2.md`) returned FAIL 0.64 — a regression from 0.71 — and emitted **STOP** rather than a fix list. All must-pass criteria now pass (MP-3 genuinely closed: `spec lint --strict spec.md` → `✓ No findings`, and the SPEC contributes zero findings to the repo-wide run), and across both iterations the auditor re-executed **62 judgment commands with zero divergence**. The regression was consistency and coverage in a scope that had outgrown single-pass remediation: iteration 1's remediation updated some surfaces and missed others, and the misses were themselves new defects.

The user chose the auditor's **scope-reduction** proposal over iteration 3. The split runs along the run/sync seam:

| | SPEC-A (this SPEC) | SPEC-B |
|---|---|---|
| ID | `SPEC-GOAL-SURFACE-UNIFY-001` | `SPEC-GOAL-DOCS-RETIRE-001` |
| Scope | doctrine + Go, run-phase, M1-M7 | public docs, sync-phase |
| Tier | L | M |
| REQs | 25 (001-024 + 028) | 8 |
| ACs | 30 (001-027 + 030, 031, 033) | 12 |
| Retention surfaces | 3 (doctrine ×2, Go ×1) | 4 (docs) |

Iteration-2 findings closed in SPEC-A:

| Finding | Resolution |
|---|---|
| N1 — `spec.md` REQ-GSU-004 bound retention to FOUR while four other artifacts said SIX | **Re-derived, not edited.** REQ-GSU-004 now binds by reference to the `plan.md` §A.2 register instead of restating a table, so the two cannot drift again. Count derived from post-split layer membership: 6 − 3 docs-layer surfaces = **3**. All five surfaces agree (verified in §E.1 evidence below) |
| N2 — AC-028's a3 detector was English-only (`en:2 ja:0 ko:0 zh:0`) | Moved to SPEC-B and re-anchored on locale-invariant literals; every SPEC-B detector now carries a per-locale baseline proving symmetry |
| N3 — `plan.md` asserted a guard that did not exist for `orchestration-mode-selection.md`'s 4 residue tokens | File added to AC-GSU-031's list (baseline `99` → `112`), retention-tested first |
| N4 — the `moai-meta-harness` mirror had no guard | Added to AC-GSU-019's parity loop (`same=13` → `same=14`) |
| N5 — `acceptance.md` contradicted itself about AC-023 | The overclaim was **removed**, not disclaimed; the edge case now points at AC-GSU-033 |
| S-new-1 | AC-GSU-004 and AC-GSU-018 strengthened with content assertions, closing two proxy-coverage matrix rows |
| S-new-2 | Both duplicated paragraph pairs deleted |
| S-new-4/5/6/7 | Stale counts reconciled against `plan.md` §F.1 as the arithmetic SSOT; HISTORY rows 1.2.0 and 1.3.0 added; `version: 1.3.0` |
| S-old-S10 | AC-GSU-031 gained a `grep -c '2.1.139' run.md` → `0` component |

Also decided rather than deferred (the auditor's flagged judgment call): `session-handoff.md:163`'s Paste-Time Activation Matrix classification statement keeps its classification but drops the `/goal` token, since the classification survives canonically in retention register row 2 (`native-invocation-model.md`). Recorded in `plan.md` M2.

_Awaiting plan-audit iteration 3 (SPEC-A) and a first audit for SPEC-B, then Implementation Kickoff Approval._

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
