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
- 8 REQs, all cited by ≥ 1 AC (`acceptance.md` §E).
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

_Awaiting first plan-audit for this SPEC. N3 additionally blocked on the parent's M7 landing (`acceptance.md` §C dependency gate)._

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
