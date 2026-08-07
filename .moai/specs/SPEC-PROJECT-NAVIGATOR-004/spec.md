---
id: SPEC-PROJECT-NAVIGATOR-004
title: "Fix Next-task misclassification — exclude implemented SPECs, prefer in-progress"
version: 0.1.0
status: completed
created: 2026-08-07
updated: 2026-08-07
author: manager-spec
priority: P1
phase: "v3.3 target"
module: navigator
lifecycle: spec-anchored
tier: M
tags: "navigator, next-task, template, bash, bugfix"
related_specs: [SPEC-PROJECT-NAVIGATOR-001, SPEC-PROJECT-NAVIGATOR-002, SPEC-PROJECT-NAVIGATOR-003]
---

# SPEC-PROJECT-NAVIGATOR-004 — Fix Next-task misclassification

## HISTORY

- 2026-08-07 — Plan-phase authored. Bug repro verified against worktree HEAD `adaff36e5` (`internal/template/templates/.claude/skills/moai-workflow-project/scripts/navigator-regen.sh` lines 211–224). Template/mirror byte-parity confirmed via `cmp`.

## §A. User Story

**As a** MoAI maintainer navigating a large SPEC registry,
**I want** the Project Navigator's "Next task" line to surface the genuinely active SPEC (the one `in-progress`, else the alphabetically-first `draft`),
**so that** the post-completion next-step suggestion does not stall on a long-`implemented` legacy SPEC while the real work-in-progress goes unnamed.

### Observed defect

After PR #1365 merged, the Navigator's "Next task" pointed at `SPEC-AGENCY-ABSORB-001` (status `implemented`, a 2026-05 legacy SPEC), while the single `in-progress` SPEC and the autonomy-epic critical-path SPEC were both missed. Originating incident captured in `feedback_navigator_next_task_misclassification.md`.

## §B. Requirements (GEARS)

### REQ-PN-004-001 — Next task excludes implemented SPECs (Ubiquitous)

The Navigator regeneration script shall exclude any SPEC whose status is `implemented` from the "Next task" candidate set, in addition to the existing `completed` / `superseded` / `archived` / `rejected` exclusions.

### REQ-PN-004-002 — Next task prefers in-progress over draft (State-driven)

**While** two or more SPECs satisfy the "Next task" candidate filter (REQ-PN-004-001), the Navigator regeneration script shall select the `in-progress` SPEC ahead of any `draft` SPEC, with alphabetical sort (`sort -k1`) applied only as a tiebreaker within a status tier.

### REQ-PN-004-003 — Current frontier display remains non-terminal-inclusive (State-driven)

**While** the "Current frontier" list is rendered, the Navigator regeneration script shall continue to list every SPEC whose status is not `completed` / `superseded` / `archived` / `rejected`, including `implemented` SPECs, so that recently-implemented work remains visible for context without being eligible for "Next task" recommendation.

> Scope note: REQ-PN-004-003 preserves the frontier's existing inclusive semantics — it codifies the deliberate split between the *display* list (broad, includes implemented) and the *recommendation* target (narrow, excludes implemented). REQ-PN-004-001 + REQ-PN-004-002 bind the recommendation only.

### REQ-PN-004-004 — Downstream propagation requires no hook change (Capability gate)

**Where** the SessionStart hook `handle-session-start-navigator.sh` consumes the regenerated `navigator.md`, the fix shall propagate automatically with no hook-side change, because the hook reads the "Next task" line from the generated document rather than recomputing it.

### REQ-PN-004-005 — Template and mirror stay byte-identical (Ubiquitous)

The Navigator regeneration script edits shall land in both `internal/template/templates/.claude/skills/moai-workflow-project/scripts/navigator-regen.sh` (template source) and `.claude/skills/moai-workflow-project/scripts/navigator-regen.sh` (local mirror) as byte-identical content within a single commit.

### REQ-PN-004-006 — Regression test fails before fix, passes after (Event-driven)

**When** the navigator regression test suite is run against the pre-fix script logic, a new fixture containing a mix of `implemented` + `in-progress` + `draft` SPECs shall demonstrate the pre-fix misclassification (the `implemented` SPEC selected as "Next task"); **when** the same suite is run against the post-fix logic, the `in-progress` SPEC shall be selected and the `implemented` SPEC excluded.

## §C. Constraints

- Template-First (CLAUDE.local.md §2 [HARD]): template source + mirror edited in one commit, byte-identical, then `make build` recompiles the embedded FS.
- §25 template neutrality: the fix is generic selection logic — no SPEC IDs, dates, commit SHAs, or moai-adk-internal references in the script body. Run the 5-item pre-commit self-check (`.moai/docs/template-internal-isolation-doctrine.md` §25.3).
- Reproduction-First (CLAUDE.md §7 Rule 4): the regression test must fail on the current logic before the script fix is applied; it must pass after.
- No new dependencies; pure bash + awk.
- `@MX` tags: not applicable (the change is in a shell script, not Go).
- Backward compatibility: the `Current frontier` display list keeps its existing inclusive semantics (REQ-PN-004-003) — no display regression for users who rely on seeing implemented SPECs for context.

## §D. Acceptance Criteria (summary — full Given-When-Then in acceptance.md)

- AC-001 — implemented SPEC excluded from "Next task" (REQ-PN-004-001)
- AC-002 — in-progress prioritized over draft when both are candidates (REQ-PN-004-002)
- AC-003 — `Current frontier` display still lists implemented SPECs (REQ-PN-004-003)
- AC-004 — SessionStart hook source unchanged yet propagates the fix (REQ-PN-004-004)
- AC-005 — template/mirror byte-parity after the edit (REQ-PN-004-005)
- AC-006 — regression test red before fix, green after fix (REQ-PN-004-006)
- AC-007 — §25 neutrality pre-commit self-check clean (constraint)

## §E. Out of Scope

### Out of Scope — completion-triggered Navigator re-run

- Auto-re-running the Navigator immediately after a SPEC's sync-phase completion (the "completion → Navigator → next identification → suggest" loop the memory's How-to-apply mentions as a follow-up). This is a separate enhancement candidate; it lives in the orchestrator, not in the regeneration script, and is deferred.

### Out of Scope — phase / milestone alignment for Next task

- Choosing "Next task" by phase ordering or milestone-critical-path alignment (e.g. prefer a SPEC at M0 over a SPEC at M3). REQ-PN-004-002 binds only the status tier (in-progress > draft); deeper phase-aware selection is deferred.

### Out of Scope — navigator.md prose restructure

- Restructuring the "Next task" prose block (e.g. showing the runner-up, or explaining why a SPEC was chosen). The fix changes only the selection, not the wording.

### Out of Scope — non-English locale rendering of the "Next task" line

- The "Next task" line is a status-agnostic prose template; the selection fix does not touch locale-rendered text.

## §F. Technical Approach (summary — see plan.md)

- Edit `navigator-regen.sh` "Current frontier" + "Next task" block (lines ~205–224): introduce a status-tier-aware "Next task" selection — exclude `implemented` from the candidate set for the recommendation, then prefer `in-progress` over `draft` with alphabetical tiebreak.
- Keep the `Current frontier` display filter unchanged.
- Extend `internal/template/navigator_regen_test.go` with a fixture (mix of implemented + in-progress + draft) and a new test asserting AC-001 + AC-002.
- Optional one-line note in `references/navigator.md` §2 schema noting the status-tier preference.
- `make build` to recompile the embedded FS.

## §G. Risks

- **R1 — Filter asymmetry drift**: the "Next task" filter and the "Current frontier" filter now differ (frontier includes implemented, next-task excludes). Risk: a future maintainer re-unifies them and re-introduces the bug. Mitigation: an inline comment at the filter site explaining the deliberate asymmetry, + the regression test.
- **R2 — Status enum drift**: if a new non-terminal status is added to the SPEC frontmatter enum later, the awk filter must be updated. Mitigation: the filter uses a positive "in-progress OR draft" recommendation predicate, not a negative exclusion list, so new statuses default to excluded from "Next task" (safe).
- **R3 — Tier-4 harness-learner misfire**: not applicable (no LSEL surface touched).

## §H. Cross-References

- Bug report: `feedback_navigator_next_task_misclassification.md` (memory).
- Originating feature SPECs: SPEC-PROJECT-NAVIGATOR-001 / -002 / -003 (all closed).
- Epic context: `project_autonomy_workflow_epic.md` (this closes one remaining autonomy-epic item).
- Template-First Rule: CLAUDE.local.md §2.
- §25 neutrality doctrine: `.moai/docs/template-internal-isolation-doctrine.md`.
