# SPEC-TABSCHEMA-AUTOBRANCH-001 — Progress

Card: t316
Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t316`
Branch: `WT-tabschema-autobranch`
Plan-phase baseline HEAD: `7ed6edb3e`

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`.

- SPEC ID regex self-check executed as Bash — output `PASS`.
- SPEC ID uniqueness confirmed: `ls -d .moai/specs/SPEC-TABSCHEMA-AUTOBRANCH-001` → exit 1,
  `No such file or directory`.
- All defect coordinates in `spec.md` §1 re-verified on this tree, not carried over.
- All eight acceptance criteria carry a measurement taken on `7ed6edb3e`.
- `tab_schema.json` itself was NOT edited in plan phase.

Amended at v0.1.1 (plan-audit iteration-1 defect closure): `tier: S` added to frontmatter; batch
3.10 provenance disclosed in `spec.md` §1; REQ → AC traceability table added as `spec.md` §6 with
covering-REQ citations on every AC heading; two paired sub-criteria added (`AC-TSA-007b` diff-shape,
`AC-TSA-005b` embedded asset); the schema's pre-existing counter drift recorded out of scope. No
existing REQ or AC was renumbered, reworded, or deleted. `tab_schema.json` remains unedited.

### §E.1 Gaps — what plan phase did NOT observe

- **`make build` has not been run**, by design (it would dirty the tree during plan phase). No claim
  is made here about the embedded asset's state, and `bin/` does not exist on this tree
  (`ls: bin/: No such file or directory`). AC-TSA-005b is therefore adopted with a stated-but-not-
  observed RED cell: its red-in-principle basis is the source-side count that a build made today
  would embed, not an observed failing run.
- **The manual-mode `automation` block was confirmed structurally, not from a rendered config.** The
  claim that the manual profile carries an `automation` block like personal and team rests on the
  Go struct definition (`ModeProfile` is one type used for all three modes), not on a generated
  `git-strategy.yaml` for a manual-mode project.
- **Template neutrality is asserted as a delta, not as an absolute.** AC-TSA-008 asserts that the
  scan output is unchanged from a non-zero baseline; it does not assert that the template copy is
  free of SPEC IDs and dates, because it is not, for reasons predating this card.
- **`moai spec lint` has not been run against this SPEC by the plan-phase author.** It is a
  Definition-of-Done item for run phase.
- **No runtime consumer was exercised.** No code reads `tab_schema.json` (`spec.md` §4), so the
  interview semantics AC-TSA-001 encodes were reconstructed from the schema's own condition data
  and never executed.
- **Now closed, previously open:** the trailing-comma deletion boundary. It was derived by reading
  at v0.1.0 and has since been verified by execution — the two objects span lines `516-536` and
  `726-746`, each is the final element of its `questions` array, and the preceding element closes
  with `},` at lines `515` and `725`.

### §E.1 Residual risk — what could still be wrong despite the above

- **AC-TSA-001's `mode_admits` predicate is a reconstruction no code enforces.** With no runtime
  consumer there is no executable oracle for it. The risk is bounded by the change being purely
  subtractive: it removes questions bound to a path no struct field matches, which holds under any
  interview semantics.
- **AC-TSA-007b compares parsed JSON**, so a pure reformat of untouched regions would still read
  `True`. `plan.md` §B forbids serializer round-tripping and the numstat corroborator would show it,
  but no criterion mechanically rejects it.
- **AC-TSA-005b's binary scan is attributable only while the dead-path strings stay confined to the
  schema.** Measured now: zero occurrences anywhere in `internal/`, `pkg/`, `cmd/` outside the two
  `tab_schema.json` copies. Should a future change introduce those strings elsewhere in compiled
  code, the criterion's `0` stops being attributable and needs re-scoping.
- **`plan.md` §C line coordinates drift** the moment anything above line 516 changes. They are
  correct on `7ed6edb3e` and are anchored by `field` value as well as by line number, so the drift
  is recoverable — but a stale re-read against a moved tree would mis-target.
- **The schema's self-declared counters stay wrong after this card**, and the `total_settings` drift
  widens from `60 vs 48` to `60 vs 46`. Recorded out of scope in `spec.md` §4 so the widened number
  is not later attributed to this deletion.

_Status: awaiting plan audit and Implementation Kickoff Approval._

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
