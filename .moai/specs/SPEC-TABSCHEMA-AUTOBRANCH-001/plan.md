# SPEC-TABSCHEMA-AUTOBRANCH-001 — Implementation Plan

Tier S. One key alignment on one JSON file, mirrored across a template/local pair.

## A. Context

See `spec.md` §1. The decision (delete, do not rebind) is made and is not re-litigated in run phase.

## B. Constraints

- **Template-First is mandatory.** The template copy is edited first, then `make build` regenerates
  the embedded assets, then the local mirror is synced. A local-only edit is reverted by the next
  `moai update`. Both copies must end byte-identical, and that identity is itself an acceptance
  criterion (AC-TSA-005).
- **Deletion only.** No rebinding, no reordering, no reformatting of untouched regions. The two
  question objects come out; nothing else moves.
- **No `go test ./...`.** Run the targeted checks in `acceptance.md` only, then let CI decide the
  full-suite verdict.
- The schema is hand-maintained JSON. Do not round-trip it through a JSON serializer — that would
  reformat the whole file and destroy the "no other question altered" evidence. Edit the two object
  literals in place.

## C. Milestones

Ordered by decision-reversibility: the surface with the durable consequence first, mechanical
propagation last.

### M1 — Delete the two question objects in the template copy

Target: `internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json`.

Remove the question object whose `field` is `git_strategy.personal.auto_branch` (batch 3.3, the last
of that batch's four questions, at local-copy lines 516-536) and the one whose `field` is
`git_strategy.team.auto_branch` (batch 3.6, likewise the last of four, at local-copy lines 726-746).

Both are the **final** element of their `questions` array, so the deletion must also remove the
trailing comma on the now-last preceding element (`git_strategy.{personal,team}.github_integration`)
to keep the array valid. This is the one place the edit can silently produce invalid JSON; AC-TSA-006
is the guard.

Do not touch batch 3.10.

### M2 — Regenerate embedded assets

`make build`. This recompiles the binary with the edited template (`//go:embed all:templates`).

This milestone is measured, not asserted: AC-TSA-005b records the `make build` exit code AND scans
the produced `bin/moai` for the dead-path strings, because AC-TSA-005 compares only the two source
files and would pass with a stale embedded asset. Do NOT substitute a `go test` that reads the
embedded FS — that comparison recompiles both sides from the same tree and is a tautology
(AC-TSA-005b states the exclusion in full).

### M3 — Sync the local mirror

Bring `.claude/skills/moai-workflow-project/schemas/tab_schema.json` to byte-identity with the
template copy. Verify with `diff -q` (AC-TSA-005), not by eye.

### M4 — Run the acceptance batch

Execute the checks in `acceptance.md` — AC-TSA-001 through AC-TSA-008 plus the paired sub-criteria
AC-TSA-005b and AC-TSA-007b — as a single parallel batch and record their verbatim output in
`progress.md` §E.2. AC-TSA-005b depends on M2 having run, so it is the one criterion that is not
independent of milestone ordering.

Pin the baseline, do not re-derive it: AC-TSA-007b compares against `7ed6edb3e` explicitly. Writing
`HEAD` there instead is silently vacuous once M1 has landed, because `HEAD` is then the post-change
commit and the comparison reduces to comparing the file with itself.

## D. Risks

| Risk | Mitigation |
|---|---|
| Deleting one object too many (e.g. also taking `github_integration`) | AC-TSA-001 counts to an exact integer; AC-TSA-007 asserts the total question-count delta is exactly `-2` |
| Trailing-comma breakage from removing a final array element | AC-TSA-006 parses both copies |
| Editing the local copy first, then having `moai update` revert it | M1 targets the template; M3 is a sync, not an independent edit |
| Silent divergence between the two copies | AC-TSA-005 is a `diff -q` on the pair |
| Reformatting the file wholesale | AC-TSA-007 total-count delta plus AC-TSA-003 (batch 3.10 sites unchanged) |
| Silently altering a question object this card does not own | AC-TSA-007b compares the parsed baseline-minus-the-two-objects against the post-change file for deep equality; every count-based criterion passes this mutant |
| Editing the template and skipping `make build`, leaving the shipped embedded asset stale | AC-TSA-005b scans the built `bin/moai`, not the source pair |

## E. Anti-Patterns

- Rebinding instead of deleting — rejected in `spec.md` §3.
- "Fixing" the manual-mode gap while in the file — explicitly out of scope (`spec.md` §4).
- Reading `grep -c auto_branch` as sufficient evidence. It is necessary but not sufficient: it cannot
  distinguish deleting the right two objects from deleting a wrong one. AC-TSA-001 is the criterion
  that decides correctness; AC-TSA-004 is a cheap corroborator.

## F. Cross-References

- `spec.md` §3 — the decision and its rationale
- `spec.md` §6 — the REQ → AC traceability table
- `acceptance.md` — the eight criteria (plus the AC-TSA-005b / AC-TSA-007b paired sub-criteria) and
  their RED-now measurements
- `progress.md` §E.1 — the plan-phase Gaps and Residual-risk record
- `internal/config/types.go` — the struct tags establishing the canonical path
- `CLAUDE.local.md` §2 — the Template-First rule
