# SPEC-WEB-CODEX-PANEL-001 — progress

Card: t509 (axis B1) · Tree: `.claude/worktrees/t509` · Branch: `WT-codex-model-config`

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, this file.

- SPEC ID `SPEC-WEB-CODEX-PANEL-001` — regex check executed as Bash, output `PASS`; collision check
  against `.moai/specs/` in this worktree and in the primary checkout returned no existing
  directory.
- Mechanism fixed by operator ruling (option A, read-only mirror) — not reopened.
- Three dispatch premises re-measured against HEAD `8a6e21d98`; two corrected in spec.md §B, both
  in the direction of less work. Both corrections are reported to the lead alongside these
  artifacts.
- Status: `draft`. Awaiting plan audit and Implementation Kickoff Approval.

Plan-audit iteration 1 (`.moai/reports/t509/plan-audit.md`, FAIL 0.69, no must-pass failure) —
repaired, artifacts at version 0.2.0:

- D1 → AC-WCP-005 restated as the per-field/per-panel invariant; `boolSegment`'s radio pair
  re-read in this tree, spec.md §B.1 mechanism corrected.
- D2 → AC-WCP-011 split into ref-resolution, non-empty-diff, and filter steps. The `rc=1` trap was
  reproduced independently in this tree before rewriting.
- D3 → AC-WCP-008 scoped to `panelHTML(t, html, "codex")`; whole-body arm repurposed to prove the
  MCP surface unchanged.
- D4/D5 → `workflow.audit.model` decision closed in spec.md §C.1 as one declared exception;
  AC-WCP-006 now compares against an independent pinned list in both directions, AC-WCP-014 covers
  the exception, the "≥12" floor is gone.
- D6 → `internal/web/settings_shell.go` named in spec.md §F, rail count decided as zero in §C.2,
  REQ-WCP-012 + AC-WCP-013 + MU-6 added.
- D7 → AC-WCP-012 and AC-WCP-009 now compare extracted function bodies across revisions.
- D8/D9/D10 → mutant-split rationale corrected, `name="` + control-tag sweep and the not-last
  placement dependency pinned, icon step made actionable.

Plan-audit iteration 2 (`.moai/reports/t509/plan-audit-iter2.md`, 0.84 over the 0.80 threshold;
FAIL on the retry-contract regression clause, not on score) — two edits, both in `acceptance.md`:

- D2-1 → AC-WCP-012's extractor anchor made receiver-tolerant AND gated on a non-zero extraction
  count per side per target, with MU-8 pinning the mis-anchor mutant. `handleSave` is a method
  (`internal/web/handlers.go:350`), so the prior `^func handleSave\(` anchor matched 0 and both
  sides extracted nothing — a vacuous `IDENTICAL` on the function guarding REQ-WCP-011. The
  extraction loop was **run in this tree**: at merge-base `c068667ad`, 209/209, 67/67, 53/53.
- D2-2 → the baseline is now an explicitly computed `git merge-base origin/develop HEAD`, not a
  `git show origin/develop:<file>` read of a moving tip.

Nothing else was touched: `spec.md` and `plan.md` are unchanged at 0.2.0, and no criterion that
passed iteration 2 was edited.

Iteration-2 reinforcement (lead, same round, still `acceptance.md` only):

- All three targets re-measured in both directions (method-form and plain-form controls), each row
  summing to exactly 1: `handleSave` is a method, `parseSchemaForm` and `ApplySchemaEdits` are plain
  functions. The non-zero assertion binds all three — the anchor matching two of them today is a
  coincidence of shape, not a property.
- The vacuous pass is now an **observation**: naive anchor on `handleSave`, base 0 / head 0 /
  `diff` exit 0. The loop itself remains unrunnable in a worktree session — refused twice, first
  for a compound `git` form, then for a non-literal `awk` program — so the criterion is written as
  six plain per-side-per-target commands, all of which ran.
- Genealogy line added to AC-WCP-012: the sibling class is "the fact the verdict rests on does not
  yet exist"; this one is "the thing the predicate points at does not exist in that shape".
- New prose in this round cites function names and section numbers, never `file:line`. The
  `handlers.go:350` citation introduced in the previous round was converted; the six pre-existing
  `file:line` citations are left untouched for the lead's post-absorption re-measurement.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
