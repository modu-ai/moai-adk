# progress.md — SPEC-FACTORY-BOOTSTRAP-001

> Phase evidence and audit-ready signals. §E.1 is populated at plan-phase; §E.2–§E.4 are placeholder headings to be populated by manager-develop (run-phase) and manager-docs (sync-phase) per the Forbidden-modifications matrix.

---

## §A. Plan-Phase Artifact Set

| Artifact | Path | Status |
|---|---|---|
| spec.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/spec.md` | emitted |
| plan.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/plan.md` | emitted |
| acceptance.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/acceptance.md` | emitted |
| design.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/design.md` | emitted (Tier L) |
| research.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/research.md` | emitted (Tier L) |
| progress.md | `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/progress.md` | emitted (this file) |

---

## §B. Branch State (at plan-phase emission)

- **Branch**: `feat/factory-bootstrap-guidance`
- **HEAD**: `94025ce0a` (prior-art commit "feat(factory): announce companion session bootstrap from the SessionStart hook")
- **Base**: `chore/revert-kanban-rename` (`24c4674b5`) ← `origin/main`
- **Local ahead by**: 2 (no race)
- **Tree status at emission**: clean working tree, plan-phase artifacts are the only new files

---

## §C. Requirements Budget

- **REQ count**: 18 (REQ-FB-001..018) against the Tier L ceiling of 25.
- **AC count**: 27 against 18 REQ (9-criterion surplus reflects the multi-surface span — env / source / notice / help / docs-site / template-neutrality / sibling-boundary / worktree-isolation / breaking-change-pin). AC-FB-016a is a fail-open sub-clause paired with AC-FB-016, not a separate criterion.
- **MUST criteria**: 25; **SHOULD**: 1 (AC-FB-020); **meta**: 1 (AC-FB-026 worktree isolation).

---

## §D. Pre-Plan-Audit Self-Check

- [x] SPEC ID regex: `SPEC-FACTORY-BOOTSTRAP-001` matches `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` (Bash output: `PASS`).
- [x] Frontmatter 12 canonical fields present (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags) + tier L.
- [x] ID uniqueness: no prior `SPEC-FACTORY-BOOTSTRAP-001` directory exists at `.moai/specs/`.
- [x] Requirements in GEARS notation (Where/When/While + shall).
- [x] Out of Scope: §C carries ≥ 1 `### Out of Scope — <topic>` H3 heading with `-` bullets (satisfies `OutOfScopeRule`).
- [x] Artifact set matches Tier L (spec + plan + acceptance + design + research + progress).
- [x] spec.md carries no implementation detail (no function signatures, no API schemas — only observable behaviors, env-var names, file paths as evidence anchors).
- [x] Prior-art commit `94025ce0a` recorded in HISTORY and §A.1 as revised-not-reverted.
- [x] Sibling boundary one-sided (no edits to `.moai/specs/SPEC-KANBAN-*`).
- [x] AC003 preserve-tests named in plan.md §A.5 with package path.

---

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
spec_version: 0.2.0
spec_id: SPEC-FACTORY-BOOTSTRAP-001
tier: L
target_auditor_threshold: 0.85
requirements_count: 18
requirements_ceiling: 25
acceptance_criteria_count: 27
must_criteria: 25
should_criteria: 1
meta_criteria: 1
prior_art_commit: 94025ce0a
prior_art_relation: revised-not-reverted
predecessor_spec: SPEC-FACTORY-MODE-001
predecessor_status: completed
sibling_spec: SPEC-KANBAN-BOOTSTRAP-001
sibling_boundary: one-sided-from-this-side
out_of_scope_sections:
  - "Out of Scope — Topology-config-gated quorum and dispatch"
  - "Out of Scope — Forward reference to the sibling's supersedence"
  - "Out of Scope — Run-phase decisions"
  - "Out of Scope — Single-session factory mode"
preserve_tests:
  - "internal/cli/launcher_blockcap_infinite_test.go::TestAC003_LauncherInjectsRaisedBlockCapForInfiniteGoal"
  - "internal/cli/launcher_blockcap_infinite_test.go::TestAC003_BlockCapDoctrineClauseSpecific"
baseline_attribution:
  worktree: "/Users/goos/.moai/worktrees/kanban"
  branch: "feat/factory-bootstrap-guidance"
  head: "94025ce0a"
  base: "24c4674b5"
  tree_status: "clean"
audit_revision:
  iteration: 2
  prior_verdict: FAIL
  prior_score: 0.81
  findings_closed: [D1, D2, D3, D4, D5]
  breaking_change_94025ce0a: "companion-shape --name alone reclassified from companion entry to no-op (spec.md §A.2.1, REQ-FB-001 no--f clause, AC-FB-027)"
frontmatter_repo_conventions:
  - "related_specs: [SPEC-FACTORY-MODE-001, SPEC-KANBAN-BOOTSTRAP-001] is a repo convention NOT codified in the canonical 12-field or optional schema; spec-lint passes (moai spec lint → 0 findings) and the sibling SPEC-KANBAN-BOOTSTRAP-001 carries the same field. Schema codification is a separate SPEC."
open_clarifications: 0
blocker_report: none
```

---

## §E.2 Run-phase Evidence

_<pending run-phase — populated by manager-develop>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_
