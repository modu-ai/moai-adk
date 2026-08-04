---
id: SPEC-PROJECT-NAVIGATOR-002
title: "Project Navigator — drift / completeness audit (`--audit` mode: design-intent vs implemented feature diff)"
version: "0.0.0"
status: draft
created: 2026-08-05
updated: 2026-08-05
author: manager-spec
priority: P2
phase: "v3.2 target"
module: project-navigator
lifecycle: spec-anchored
tags: "navigator, audit, drift, missing-spec, completeness"
related_specs: [SPEC-PROJECT-NAVIGATOR-001, SPEC-PROJECT-NAVIGATOR-003]
---

# SPEC-PROJECT-NAVIGATOR-002 — Navigator `--audit` drift check (STUB)

> **STUB — not authored at plan-phase.** This entry exists to reserve the SPEC ID and record the boundary decision. It will be fully authored (spec.md + plan.md + acceptance.md + research.md + progress.md) after SPEC-PROJECT-NAVIGATOR-001 lands its artifact set, because the audit algorithm depends on what 001's capability-map actually surfaces.

## Problem (placeholder)

The user-approved Project Navigator scope names a P2 — a drift / completeness audit that diffs design intent (`product.md` / `structure.md` / `tech.md` — the intended features) against the Navigator's capability-map (the implemented features) and flags:

1. **Missing SPECs** — features named in design docs that have no corresponding SPEC in `.moai/specs/` (명세 누락 detection).
2. **Unimplemented features** — SPECs whose status has stalled (e.g. `draft` for >N sync cycles) or whose owning code path is absent from the capability-map.
3. **Stale design** — design-doc features that have been retired in implementation but never removed from `product.md` (design rot).

## Mechanism (decided in 001's plan.md §B.1)

- **Skill mode only**: `/moai project --audit` (NOT a hook — audit output can be large and is not latency-sensitive; making it a hook would violate Advisory-Check Discipline).
- Invoked on-demand OR chained into sync-phase as an optional advisory step.

## Boundary vs 001

- **001 owns**: the artifact set (capability-map, progress-map, navigator.md) + regeneration + reorientation.
- **002 owns**: the audit algorithm that READS 001's artifacts + design docs and emits a drift report.
- **003 owns**: tree-sitter auto-derivation that feeds enriched rows into 001's capability-map.

## Why this is a separate SPEC (not folded into 001)

- The audit algorithm is exploratory — its exact shape depends on what 001's capability-map actually contains, which is decided in 001's run-phase.
- Different changeability profile: 001 ships the artifact set; 002 ships the diff algorithm. Bundling them would force one plan-audit gate over two orthogonal concerns.
- Tier classification: likely **Tier S or M** (single skill mode + report format; no new infrastructure) — to be confirmed at 002's own plan-phase.

## Out of Scope (for this stub)

### Out of Scope — artifact generation

- Generating the Navigator artifact set is owned by SPEC-PROJECT-NAVIGATOR-001.

### Out of Scope — tree-sitter auto-derivation

- Auto-deriving capability rows from AST analysis is owned by SPEC-PROJECT-NAVIGATOR-003.
