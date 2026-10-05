---
id: SPEC-JEV-INTEGRATION-001
title: "Jev integration — RETIRED, split into four successors"
version: "0.2.0"
status: superseded
created: 2026-09-20
updated: 2026-09-20
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: "internal/jev + internal/cli + internal/web + internal/template/templates"
lifecycle: spec-anchored
tags: "jev, typesafe, retired, superseded, split"
tier: L
partially_superseded_by: [SPEC-JEV-CORE-001, SPEC-JEV-OPTIN-MEASURE-001, SPEC-JEV-CONSUMERS-001, SPEC-JEV-GOAL-DIST-001]
---

# SPEC-JEV-INTEGRATION-001 — RETIRED

This SPEC was authored, then retired the same day without ever entering run-phase. Its scope is carried, whole and unreduced, by four successors.

## Why it was retired

It carried 68 requirements and 56 acceptance criteria against a Tier L ceiling of 25 and 25 (`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier). Tier L is the top tier, so tiering up was not available and the ceiling's own instruction applied: split, never relax the budget. `moai spec lint` does not check the ceiling, which is why the violation survived authoring — the count is the evidence, not the linter.

Nothing was cut. Every requirement and acceptance criterion survives into exactly one successor; the operator explicitly rejected reducing scope.

## Where its content went

The split follows the milestone seams. The dependency chain was linear, so the partition is by contiguous prefix.

| Successor | Tier | Seams | Carries |
|---|---|---|---|
| `SPEC-JEV-CORE-001` | L | M1 | The `internal/jev` package, the pinned model id, size bounds, secret screening, the credential at `~/.moai/.env.typesafe`, the `workflow.jev.enabled` gate, the fail-open contract, the display-only invariant, and the `moai doctor` check |
| `SPEC-JEV-OPTIN-MEASURE-001` | L | M2 + M3 | The `moai init` wizard question (init-only, per operator decision) and the `moai web` settings section over one shared persistence path; the measurement harness, constant-answer baselines, two language arms, threshold fitting, and the question-design rules |
| `SPEC-JEV-CONSUMERS-001` | L | M4 + M5 + M6 | Near-duplicate card marking (with the third finding source constant), lane-question routing, and skill suggestion — each gated on its own measurement |
| `SPEC-JEV-GOAL-DIST-001` | M | M7 + M8 | Both `/moai goal --auto` seats, the MCP tool wrapper, the reference skill, the template and documentation surfaces, and the preserved negative result |

Execution order is the table's order: each successor declares its predecessor in `depends_on`, and no two depend on each other.

## Requirement and criterion mapping

| Original | Successor | Re-numbered as |
|---|---|---|
| REQ-JEV-001 … 020, 024, 026 | `SPEC-JEV-CORE-001` | REQ-JEVC-001 … 022 |
| REQ-JEV-021, 022, 023, 025 | `SPEC-JEV-OPTIN-MEASURE-001` | REQ-JEVO-001 … 004 |
| REQ-JEV-027 … 040 | `SPEC-JEV-OPTIN-MEASURE-001` | REQ-JEVO-008 … 021 |
| REQ-JEV-041 … 054 | `SPEC-JEV-CONSUMERS-001` | REQ-JEVN-001 … 014 |
| REQ-JEV-055 … 068 | `SPEC-JEV-GOAL-DIST-001` | REQ-JEVG-001 … 014 |

Three requirements are **new**, not re-numbered: `REQ-JEVO-005`, `REQ-JEVO-006`, and `REQ-JEVO-007` encode the operator's init-only placement decision, its consequence (the setting is unreachable from `moai update --reconfigure`), and the `moai web` pointer that keeps the switch discoverable. Three matching acceptance criteria (`AC-JEVO-007` … `AC-JEVO-009`) were added with them.

## Exclusions

### Out of Scope — everything

- This document specifies no behaviour. It is a pointer, retained so a reader arriving at this SPEC ID finds the successors rather than a dead path. Every exclusion the original declared is restated in whichever successor inherited its scope.

## Sibling artifacts

The original `plan.md`, `acceptance.md`, `design.md`, `research.md`, and `progress.md` were removed when this stub replaced the body: their content now lives in the successors, and keeping a second copy would let the two drift. They remain in git history at commit `51ababad4` for anyone tracing the split.
