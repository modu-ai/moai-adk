# SPEC-OBSERVE-HYGIENE-001 — Progress

> Plan-phase artifact. Run-phase evidence (§E.2, §E.3) and sync-phase evidence (§E.4) are populated by manager-develop and manager-docs respectively. This file is parser-load-bearing: `internal/spec/era.go` `hasAnyProgressMarker` greps for the literal `§E.2` / `§E.3` / `§E.4` substrings — the headings below MUST be preserved verbatim.

---

## §E.1 Plan-phase Audit-Ready Signal

```
plan_status: audit-ready
plan_complete_at: 2026-07-09
tier: S
epic: Workflow-Reflex (6 of 6)
artifacts: 4 (spec.md + plan.md + acceptance.md + progress.md)
gears_requirements: 6
acceptance_criteria: 7
out_of_scope_topics: 6
audit_findings_traced: H4, H5, H6, L7
related_specs: SPEC-HARNESS-RATCHET-REWIRE-001 (harness boundary), SPEC-TOKEN-BUDGET-STOP-001 (verify-dir boundary)
open_decision_points: D1 (audit-log consumer vs document — (a) recommended), D2 (retention threshold + task-metrics disposition), D3 (sync-gate blocking promotion — user confirmation required before M3)
brief_correction: model_upgrade_review HAS a consumer (emitModelUpgradeReminder, REQ-HRN-001-016); dormancy is at the never-set CLAUDE_MODEL_PREVIOUS trigger env — annotation wording adjusted accordingly
sibling_surface_coordination: loop.md (three-way with SPEC-LOOP-VERDICT-CONTRACT-001 + SPEC-ADVISOR-RUNG-001 — orchestrator sequences)
spec_id_self_check: PASS (SPEC-OBSERVE-HYGIENE-001 → ^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$)
```

**Plan-phase self-verification**:
- [x] All GEARS requirements use valid patterns (Ubiquitous / When / Where; no IF/THEN).
- [x] All gap claims cite measured source + observed pattern (vci §2; measured 2026-07-09; log sizes, zero-consumer greps, dormancy layers, and the model_upgrade_review consumer re-verified — brief claim corrected rather than propagated per vci §1.1 surface 3).
- [x] §Out of Scope has 6 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.
- [x] 12 canonical frontmatter fields + era + tier + related_specs; no snake_case aliases.
- [x] Per-sink disposition rule (exactly one of consumer/document/prune) encoded in the requirements header + DoD #3.
- [x] Hook signaling contract (exit-0 stdout block channel) + runtime-recovery §4 carve-out stated as preserved invariants.
- [x] Template mirrors verified present for all mirrored edit surfaces (sunset.yaml, harness.yaml, both hook scripts, loop.md).

---

## §E.2 Run-phase Evidence

_<pending run-phase — populated by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_
