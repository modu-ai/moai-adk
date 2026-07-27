---
id: SPEC-GOAL-DOCS-RETIRE-001
title: Retire native /goal emission references from public and internal documentation across four locales
version: 1.0.0
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: MEDIUM
phase: "v3.1.0"
module: docs
lifecycle: spec-anchored
tags: "goal, docs-site, i18n, locale-parity, split-surface"
tier: M
depends_on: [SPEC-GOAL-SURFACE-UNIFY-001]
---

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifact set complete (Tier M): `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md`.

Origin: split from `SPEC-GOAL-SURFACE-UNIFY-001` after that SPEC's plan-audit iteration 2 returned FAIL 0.64 with a STOP signal (score regression from 0.71), and the user chose the auditor's scope-reduction proposal over iteration 3.

- SPEC ID regex check executed as Bash; observed output `SPEC-ID-CHECK: PASS`.
- ID confirmed unused before authoring (`ls .moai/specs/SPEC-GOAL-DOCS-RETIRE-001` → `No such file or directory`).
- **12** acceptance criteria; every judgment command executed and its verbatim baseline recorded. All 12 fail at baseline.
- **Every emission criterion carries a per-locale baseline**, and all five emission detectors measure symmetric (`distinct=1`). This is the requirement that closes the parent's audit finding N2.
- 11 REQs, all cited by ≥ 1 AC (`acceptance.md` §E). (10 at authoring; REQ-GDR-011 added closing audit iteration 2 finding B2-1.)
- 13 paths, single-owner across N1-N5 (`plan.md` §F.1). 41 retained files owned by no milestone.
- Four retention surfaces registered in this SPEC's own register (`plan.md` §A.2) — not inherited from the parent.

Design decisions recorded rather than deferred:

| Decision | Rationale |
|---|---|
| **Tier M**, not L | 13 files inside the 15-file ceiling; one layer; one verification regime (locale-symmetric grep repeated four times is the same regime, not four); fully reversible; no irreversibility surface. The split-surface difficulty is handled by plan-time detector design, not by added implementation layers |
| `.moai/docs/autonomous-workflow-strategy.md` belongs **here**, not in the parent | Its treatment (superseding note, sweep nothing) is a sync-phase documentation action and the parent has no sync phase post-split; its content is the same 3-engine comparison as `autonomous-loops.md`; its membership-test outcome matches the other retention rows. Counter-argument acknowledged: `.moai/docs/` is not public — but the retain-plus-note treatment does not depend on publication |
| `autonomous-loops.md` is a **split surface**, judged positively on both halves | A file-level `0` target would delete two sections, a comparison row, and the Axis-B justification in four locales — the parent's D4/N2 defect |
| AC-GDR-010 exists as a **meta-guard** | Per-detector locale diligence is not self-verifying; a criterion whose subject is the detectors makes the false-zero class mechanically impossible |

Carried-identifier provenance is recorded at `spec.md` §D; the parent's moved-identifier register is at `SPEC-GOAL-SURFACE-UNIFY-001/spec.md` §B.7.

### Plan-audit iteration 1 — FAIL 0.75 (Tier M threshold 0.80), remediated

The first audit confirmed Tier M correct, reproduced all 12 baselines with zero divergence, and found the retention register consistent and fully guarded. Both MUST-FIX were defects in AC-GDR-010's *specification*, not its measurement:

| Finding | Resolution |
|---|---|
| **B-1** — `distinct=1` admits a dead detector: a typo'd anchor reads `0/0/0/0` → `distinct=1` and passes, indistinguishable post-sweep from a genuine sweep | AC-GDR-010 gained a **liveness component (b)**: every detector must match non-zero content against the immutable base `e306e21a9`. New REQ-GDR-010. Dead-detector control executed: the typo'd anchor reads `live_min=0` and is rejected |
| **B-2** — the blanket symmetry rule contradicts REQ-GDR-008, with `hooks-reference.md` (`en:1 ja:0 ko:1 zh:0` → `distinct=2`) as a live example | REQ-GDR-004 split: per-locale recording stays unconditional; symmetry moved to new **REQ-GDR-009**, conditioned on content symmetry with a named-exemption carve-out. New §A.3 states both causes of asymmetry. Exemption list currently empty |
| **B-3** — §D provenance table mapped 4 of 7 rows to the wrong criterion | All rows re-derived from each criterion's actual subject; fixed jointly with the parent's A-1 so the two registers agree |
| **B-4** — `spec.md` stated 18 emission markers vs the measured 24 | Corrected to the command-derived `24`; the stale `18` used the disqualified `per-turn` detector's `2` and omitted L7's `4` |
| **B-5** | §A.1 now marks AC-GDR-004/005/006 as deliberate presence checks |
| **B-6** | Beyond-minimum artifact set documented as deliberate; tier stays M |
| **B-7** | AC-GDR-004's fallback re-specified as the locale-invariant triple co-occurrence, replacing the prose-adjacent `provides` (measured `1/1/1/1`) |

### Plan-audit iteration 2 — PASS 0.86 (Tier M threshold 0.80), MUST-FIX B2-1 closed

Iteration 2 reproduced 12 of 12 baselines with zero divergence and scored Clarity 0.90 / Completeness 0.85 / Testability 0.80 / Traceability 0.90 (harmonic mean 0.8605). One MUST-FIX and two SHOULD-FIX were raised.

| Finding | Severity | Resolution |
|---|---|---|
| **B2-1** — AC-GDR-010 validated liveness and symmetry but not *aptness*: a detector aimed at an adjacent token (`ac_converge`) or a compound regex with its `` `/goal` `` half dropped passes (a) and (b) while matching no emission reference | MUST-FIX | **CLOSED.** New **REQ-GDR-011**; AC-GDR-010 gained component **(d)**, asserting each detector's pattern carries a literal `/goal` token from a single `p=` source shared by `w()` and the assertion. Three controls executed: the hyphenated dead detector is rejected by (b); `ac_converge` and the weakened `auto mode` half are rejected by (d) |
| **B2-2** — liveness against a frozen base can be stale-true; no criterion asserts the base is still current for the scoped pages | SHOULD-FIX | Open — deferred as follow-on |
| **B2-3** — the empty exemption list is guarded by prose, not by a check | SHOULD-FIX | Open — deferred as follow-on |

**Auditor's stronger alternative for B2-1 was executed and refuted.** The audit proposed asserting each detector's base-match line set is a subset of the base lines carrying `` `/goal` ``. Run against `e306e21a9`, both attack detectors yield `leak=0` and pass: `ac_converge` and `auto mode` occur on the *same line* as `` `/goal` `` (`en/advanced/autonomous-loops.md` line 7). Line granularity is precisely the granularity the adjacent-token attack exploits, so the subset form cannot discriminate. The pattern-literal form was adopted instead, with the single-`p=`-source discipline closing its own divergence risk.

Residual (not in B2-1's required fix): AC-GDR-012's aggregate still carries its own inline copies of the five detector definitions, so the single-source discipline does not extend to the aggregate path.

N3 remains blocked on the parent's M7 landing (`acceptance.md` §C dependency gate) — the parent is now `completed`, so the gate is expected to clear at run-phase entry and must be re-measured there.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Decision: sub-agent** (Mode 5, sequential per-milestone delegation)

**Input parameters**

| Signal | Value |
|---|---|
| tier | M (3-artifact set; 5 authored) |
| scope (file count) | 8 sweep-target documentation files; 12 scoped locale files across 3 pages |
| domain count | 1 — docs-site markdown content |
| file language mix | 100% markdown (4 locales: en / ja / ko / zh) |
| concurrency benefit | LOW — per-locale prose differs, and 4-locale parity requires coordinated edits within one reasoning context |

**Mode evaluation**

| Mode | Selected | Rationale |
|---|---|---|
| 1 `trivial` | no | Multi-file semantic sweep with 12 acceptance criteria; not a typo class |
| 2 `background` | no | Work is write-heavy (Edit on scoped pages), not read-only analysis |
| 3 `agent-team` | no | RETIRED tombstone; never selected by the decision tree |
| 4 `parallel` | no | Single domain (< 3) and edit-heavy, not research-heavy; the coding-task parallelism caveat routes this to Mode 5 |
| 5 `sub-agent` | **YES** | Default fallback; sequential per-milestone delegation preserves 4-locale coordination |
| 6 `workflow` | no | Scope is ~12 files, below the ~30-file entry threshold, and the transform is not one uniform mechanical rule — each locale's surrounding prose differs |

**Decision: sub-agent**

Mode 5 is selected. The sweep is edit-heavy rather than research-heavy, touches one domain, and its hardest constraint is locale parity — four files must change consistently against a shared detector set, which a single sequential agent holds in one reasoning context. Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research") applies directly, and the Mode 6 file-count threshold is not met.

**Boundary case** — none. No signal sat at a threshold ±1: domain count 1 (vs 3), file count ~12 (vs 30), and the transform-kind test fails Mode 6 independently of count.

**Gate provenance** — Implementation Kickoff Approval obtained from the user before this section was written. Progression mode selected at the same gate: **autonomous** (`/moai goal` armed after the gate, per `goal-directive.md` § Goal-Presentation Timing — arming pairs with the work-starting command, never replaces it).
