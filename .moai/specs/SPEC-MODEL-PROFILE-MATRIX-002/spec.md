---
id: SPEC-MODEL-PROFILE-MATRIX-002
title: "Agent-direct profile matrix + effort actualization — RETIRED, split into four successors"
version: "0.2.0"
status: superseded
created: 2026-07-23
updated: 2026-09-20
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "internal/template, internal/config, internal/cli, internal/web, internal/harness/v4manifest, docs-site"
lifecycle: spec-anchored
tags: "model-profile, agent-matrix, effort-injection, retired, superseded, split"
tier: L
era: V3R6
partially_superseded_by: [SPEC-MODEL-MATRIX-CORE-001, SPEC-MODEL-MATRIX-CONFIG-001, SPEC-MODEL-MATRIX-SURFACES-001, SPEC-MODEL-MATRIX-DOCS-001]
related_specs: [SPEC-MODEL-PROFILE-MATRIX-001, SPEC-MODEL-TIER-PLANTYPE-001, SPEC-AGENT-ARCH-V2-001, SPEC-WEBCONF-SIMPLIFY-001]
---

# SPEC-MODEL-PROFILE-MATRIX-002 — RETIRED

This SPEC was authored on 2026-07-23 and entered run-phase; two of its eight units of work landed before it was retired on 2026-09-20. Its scope is carried, whole and unreduced, by four successors.

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-07-23 | manager-spec | Initial draft. Supersedes the 6-group abstraction of SPEC-MODEL-PROFILE-MATRIX-001 with a direct 33-cell profile→agent matrix; revives frontmatter `effort:` rewrite in a narrowed form; retires the 36-cell Tier×Phase axis and the `llm.yaml profiles:` mirror; corrects SPEC-001 DECISION-001's effort-injection claim. |
| 0.2.0 | 2026-09-20 | manager-spec | Retired in place to a pointer stub and split into four successors on the milestone seams (card t1036). Sibling artifacts removed; scope preserved in full. |

## Why it was retired

It carried **72 requirements and 64 acceptance criteria** against a Tier L ceiling of 25 and 25 (`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier). Tier L is the top tier, so tiering up was not available and the ceiling's own instruction applied: split, never relax the budget. `moai spec lint` does not check the ceiling, which is why the violation survived authoring and an entire run-phase — the count is the evidence, not the linter.

The excess dates from the 2026-07-23 authoring, not from any later absorption. This was dividing what was already carried, not shedding new scope.

Nothing was cut. Every requirement and acceptance criterion survives into exactly one successor, and unlike the `SPEC-JEV-INTEGRATION-001` precedent, **no requirement was added** — no operator decision accompanied this split.

## Where its content went

The split follows the milestone seams. The dependency structure is a **DAG, not a chain**, so the buckets are deliberately not contiguous prefixes: S0 blocks only M6; M4 and M5 are independent of everything; M1 blocks M2, M3, and M7.

| Successor | Tier | Seams | Carries |
|---|---|---|---|
| `SPEC-MODEL-MATRIX-CORE-001` | L | S0 + M1 | The leaderboard verification record and the 33-cell agent-direct matrix: the Go SSOT, the resolver precedence, the explicit `Explore` row, the group removal, and the amended resolver tests |
| `SPEC-MODEL-MATRIX-CONFIG-001` | L | M2 + M3 | The 36-cell axis retirement, the `llm.yaml profiles:` mirror removal with its no-silent-drop migration, and effort actualization on both channels — the frontmatter rewrite and the Workflow-path lookup route |
| `SPEC-MODEL-MATRIX-SURFACES-001` | M | M4 + M5 | The `moai init` model-policy question with its four-locale translations, and the web console cleanup — the agentfm model option set, the v4manifest lightblue suggestion, and the two i18n key families |
| `SPEC-MODEL-MATRIX-DOCS-001` | L | M6 + M7 | The four-locale documentation set with the naming-inversion disclosure and the Max-Opus rationale, the rules-file reconciliation, and the haiku-residual guard realignment plus full verification |

Execution order follows the dependency edges rather than the table order. `SPEC-MODEL-MATRIX-CORE-001` and `SPEC-MODEL-MATRIX-SURFACES-001` are both roots and may proceed in parallel; `SPEC-MODEL-MATRIX-CONFIG-001` depends on CORE; `SPEC-MODEL-MATRIX-DOCS-001` depends on all three. The graph is acyclic.

### Why four rather than three

Three is arithmetically feasible on counts alone — a 24/25/23 partition exists — but only by severing the cross-cutting requirements from the milestones they attach to and by packing two buckets to 24 and exactly 25, with no headroom. Those buckets would have been formed by arithmetic rather than by cohesion.

Three was therefore rejected on **attachment integrity and headroom**, not on feasibility. Four keeps each cross-cutting item with its attachment milestone and leaves every bucket below its ceiling with room to spare.

### Why S0 is not with the milestone it blocks

S0 blocks M6, and the natural reading is that they should travel together. Placing S0 in the DOCS successor would make that bucket **27 requirements and 26 criteria** — over the Tier L ceiling this split exists to satisfy. S0 stays in `SPEC-MODEL-MATRIX-CORE-001`, and the blocking relation survives structurally as DOCS's `depends_on` edge on CORE, which DOCS already carries via M1.

`SPEC-MODEL-MATRIX-DOCS-001` states explicitly in its own body that its blocking precondition lives in a predecessor and is **undischarged**.

## Run-phase state at retirement

This SPEC was `in-progress`, not untouched. Two of eight units landed in squash `31da99a7b` (PR #1163):

| Unit | State | Inherited by |
|---|---|---|
| M1 — 33-cell matrix redesign | **landed** | `SPEC-MODEL-MATRIX-CORE-001` |
| M6 — 4-locale documentation + README ×4 | **landed** | `SPEC-MODEL-MATRIX-DOCS-001` |
| S0 — leaderboard verification (BLOCKING) | **not discharged** | `SPEC-MODEL-MATRIX-CORE-001` |
| M2, M3 | not started | `SPEC-MODEL-MATRIX-CONFIG-001` |
| M4, M5 | not started | `SPEC-MODEL-MATRIX-SURFACES-001` |
| M7 | not started | `SPEC-MODEL-MATRIX-DOCS-001` |

Three facts travel with that state and are recorded in the inheriting successors rather than lost here:

- **S0 blocked M6, and M6 shipped anyway.** M1 and M6 proceeded on the user-supplied per-effort measurements; S0's own verification was never performed. Discharging S0 is now remediation behind live pages rather than a gate ahead of them.
- **M1's plan steps 2 and 3 were never executed** despite M1 being recorded as landed: `agentGroupMembership` and `AgentGroup` are still alive, and `internal/web/agentfm.go:491` calls `AgentGroup` as a live gate. Card **t1037** owns the disposition — is the landing record wrong, or is the plan stale? `SPEC-MODEL-MATRIX-CORE-001` records the remainder as its own and does not decide it.
- **64 acceptance criteria were formally verified: zero.** `internal/` carries no `REQ-MPM2` / `AC-MPM2` markers, so requirement-to-code traceability is unestablished for the landed work.

## Requirement and criterion mapping

Every one of the 72 requirements and 64 criteria appears exactly once below.

### Requirements

| Original | Successor | Re-numbered as |
|---|---|---|
| REQ-MPM2-001 … 004 (S0) | `SPEC-MODEL-MATRIX-CORE-001` | REQ-MPMC-001 … 004 |
| REQ-MPM2-010 … 023 (M1) | `SPEC-MODEL-MATRIX-CORE-001` | REQ-MPMC-005 … 018 |
| REQ-MPM2-030 … 037 (M2) | `SPEC-MODEL-MATRIX-CONFIG-001` | REQ-MPME-001 … 008 |
| REQ-MPM2-113 (cross-cutting → M2) | `SPEC-MODEL-MATRIX-CONFIG-001` | REQ-MPME-009 |
| REQ-MPM2-040 … 051 (M3) | `SPEC-MODEL-MATRIX-CONFIG-001` | REQ-MPME-010 … 021 |
| REQ-MPM2-112 (cross-cutting → M3) | `SPEC-MODEL-MATRIX-CONFIG-001` | REQ-MPME-022 |
| REQ-MPM2-060 … 065 (M4) | `SPEC-MODEL-MATRIX-SURFACES-001` | REQ-MPMS-001 … 006 |
| REQ-MPM2-070 … 074 (M5) | `SPEC-MODEL-MATRIX-SURFACES-001` | REQ-MPMS-007 … 011 |
| REQ-MPM2-080 … 092 (M6) | `SPEC-MODEL-MATRIX-DOCS-001` | REQ-MPMD-001 … 013 |
| REQ-MPM2-110, 111 (cross-cutting → M6) | `SPEC-MODEL-MATRIX-DOCS-001` | REQ-MPMD-014, 015 |
| REQ-MPM2-100 … 105 (M7) | `SPEC-MODEL-MATRIX-DOCS-001` | REQ-MPMD-016 … 021 |

Totals: 18 + 22 + 11 + 21 = **72**.

> Note: `REQ-MPM-024` appears once in this SPEC's original body. It is a reference to the predecessor `SPEC-MODEL-PROFILE-MATRIX-001` (prefix `MPM`, not `MPM2`) and was never one of this SPEC's own requirements. A naive `REQ-` grep returns 73; the owned set is the 72 `REQ-MPM2-*` above.

### Acceptance criteria

| Original | Successor | Re-numbered as |
|---|---|---|
| AC-MPM2-001 … 003 (S0) | `SPEC-MODEL-MATRIX-CORE-001` | AC-MPMC-001 … 003 |
| AC-MPM2-010 … 019 (M1) | `SPEC-MODEL-MATRIX-CORE-001` | AC-MPMC-004 … 013 |
| AC-MPM2-020 … 026 (M2) | `SPEC-MODEL-MATRIX-CONFIG-001` | AC-MPME-001 … 007 |
| AC-MPM2-103 (cross-cutting → M2) | `SPEC-MODEL-MATRIX-CONFIG-001` | AC-MPME-008 |
| AC-MPM2-030 … 040 (M3) | `SPEC-MODEL-MATRIX-CONFIG-001` | AC-MPME-009 … 019 |
| AC-MPM2-102 (cross-cutting → M3) | `SPEC-MODEL-MATRIX-CONFIG-001` | AC-MPME-020 |
| AC-MPM2-050 … 054 (M4) | `SPEC-MODEL-MATRIX-SURFACES-001` | AC-MPMS-001 … 005 |
| AC-MPM2-060 … 064 (M5) | `SPEC-MODEL-MATRIX-SURFACES-001` | AC-MPMS-006 … 010 |
| AC-MPM2-070 … 082 (M6) | `SPEC-MODEL-MATRIX-DOCS-001` | AC-MPMD-001 … 011, 013, 014 |
| AC-MPM2-100, 101 (cross-cutting → M6) | `SPEC-MODEL-MATRIX-DOCS-001` | AC-MPMD-015 |
| AC-MPM2-090 … 095 (M7) | `SPEC-MODEL-MATRIX-DOCS-001` | AC-MPMD-016 … 021 |

Totals: 13 + 20 + 10 + 21 = **64**.

> Two shape notes in the DOCS column. `AC-MPMD-012` is **not** a re-numbered original — it is the correction criterion the split's own analysis required, because M6 shipped ahead of its blocking precondition and the original SPEC had no criterion for correcting an already-published figure. And `AC-MPM2-100` and `AC-MPM2-101` merge into the single `AC-MPMD-015`, which asserts the Fable-fallback and the three GLM statements together, since both are read from the same documentation section. The 13 M6 criteria plus that merged pair plus the new correction criterion give DOCS its 21.

### Where the open clarifications and unverified inputs went

The four `[NEEDS CLARIFICATION]` items from `research.md` §F:

| Item | Successor |
|---|---|
| `Group` field in `moai model profile --json` | `SPEC-MODEL-MATRIX-CORE-001` |
| `profiles:` migration representability | `SPEC-MODEL-MATRIX-CONFIG-001` |
| Wizard option value vocabulary | `SPEC-MODEL-MATRIX-SURFACES-001` |
| Infographic disposition | `SPEC-MODEL-MATRIX-DOCS-001` |

The eight unverified carried-forward inputs from `research.md` §G: the three S0 leaderboard items to `SPEC-MODEL-MATRIX-CORE-001`; the `mp.*` i18n orphan status to `SPEC-MODEL-MATRIX-SURFACES-001`; the ja/ko `multi-llm/model-policy.md` shape, the `advanced/*` line ranges, the infographic contents, and cross-platform build status to `SPEC-MODEL-MATRIX-DOCS-001`.

### Downstream card addresses

- Card **t1031**'s M5 expansion entry lands in `SPEC-MODEL-MATRIX-SURFACES-001`. That is the fixed address it was waiting on.
- Card **t1037** owns the unexecuted M1 remainder, which `SPEC-MODEL-MATRIX-CORE-001` records as its own.

## Exclusions

### Out of Scope — everything

- This document specifies no behaviour. It is a pointer, retained so a reader arriving at this SPEC ID finds the successors rather than a dead path. Every exclusion the original declared is restated in whichever successor inherited its scope.

## Sibling artifacts

The original `plan.md`, `acceptance.md`, `design.md`, `research.md`, and `progress.md` were removed when this stub replaced the body: their content now lives in the successors, and keeping a second copy would let the two drift. They remain in git history at commit `c49660dd1` for anyone tracing the split.
