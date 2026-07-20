# SPEC-ADVISOR-RUNG-001 — Progress

> Plan-phase artifact. Run-phase evidence (§E.2, §E.3) and sync-phase evidence (§E.4) are populated by manager-develop and manager-docs respectively. This file is parser-load-bearing: `internal/spec/era.go` `hasAnyProgressMarker` greps for the literal `§E.2` / `§E.3` / `§E.4` substrings — the headings below MUST be preserved verbatim.

---

## §E.1 Plan-phase Audit-Ready Signal

```
plan_status: audit-ready
plan_complete_at: 2026-07-09
tier: S
epic: Workflow-Reflex (4 of 6)
artifacts: 4 (spec.md + plan.md + acceptance.md + progress.md)
gears_requirements: 6
acceptance_criteria: 7
out_of_scope_topics: 5
audit_findings_traced: R3, R6
depends_on: SPEC-MODEL-ROUTING-WIRE-001
open_decision_points: D1 (advisor doctrine placement), D2 (ci-autofix touch scope), D3 (same-diagnostic identity), D4 (advisor model default)
sibling_surface_coordination: loop.md (SPEC-LOOP-VERDICT-CONTRACT-001 run-phase must land first or be coordinated)
spec_id_self_check: PASS (SPEC-ADVISOR-RUNG-001 → ^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$)
```

**Plan-phase self-verification**:
- [x] All GEARS requirements use valid patterns (Ubiquitous / When / Where / Unwanted; no IF/THEN).
- [x] All gap claims cite measured source + observed pattern (vci §2; measured 2026-07-09; fix.md:312-314 + glm.md mode table + Avoid-grep-0 re-verified).
- [x] §Out of Scope has 5 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.
- [x] 12 canonical frontmatter fields + era + tier + depends_on; no snake_case aliases.
- [x] Frozen-clause preservation (ci-autofix 3-iteration ceiling) stated as HARD constraint + AC-ADV-004.
- [x] Template mirrors verified present for all 5 edit surfaces (fix.md, loop.md, ci-autofix-protocol.md, team/glm.md, team/run.md).

---

## §E.2 Run-phase Evidence

_<pending run-phase — populated by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_
