# SPEC-AUTONOMY-PRECONDITION-001 — Progress

Card t1245, track A2b. Worktree `.claude/worktrees/t1245`, branch `WT-push-serialize-sign`,
base `develop` at `553e224f3`.

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `progress.md`.
- Tier **M**: two components in one package plus one projection function; **9 live requirements /
  12 criteria** at v0.1.1; below the Tier L threshold of ≥3 milestones **and** ≥10 files.
  `design.md` is added beyond the Tier M set because the card is Class C and carries two genuine
  design decisions (design.md §A).
- **v0.1.1 revisions** (two lead corrections): all A1 citations pinned to commit `67a2f55cb`
  (v0.5.0) with every coordinate re-measured at that SHA rather than carried over; `push_requires_lease`
  read from `moai contract show --json` per A1 `:382`; the A-Q2 ownership clause and the required
  ordering `A1 → A2 + A2b → A3` quoted verbatim; `decide` added to the deny matcher; the
  receipt-path allowance withdrawn (REQ-AP-006, AC-AP-011, AC-AP-012 retired and not re-used) with
  `AC-AE-025` recorded as excluded to A3, reason N5.
- SPEC ID regex check executed as Bash against the `internal/spec/lint.go` pattern: `PASS`.
- Requirements in GEARS notation; every requirement traces to ≥1 criterion and every criterion maps
  one requirement (acceptance.md §G).
- Exclusions section present with five `### Out of Scope — …` sub-headings (spec.md §G).
- ID-reuse and ID-retirement records written in A1's own format (spec.md §H.1-H.2); no retired id
  re-used.
- Premises measured against `553e224f3` before any requirement was written; two dispatch premises
  did not hold and are recorded as absent rather than assumed (spec.md §C.2, §C.3, §C.6).
- Transferred text read from `WT-escalation-detector~1` through committed refs only; the criterion
  id re-use hazard is recorded with a provenance table (spec.md §H).
- Open items after the corrections: O1 resolved (`decide` in scope; its owner unidentified), O2
  partially resolved (outright deny adopted; a role distinction still unnamed), plus body register
  (plan.md §I O3) and the audit sink (spec.md §F O3). None blocks a criterion's evaluability.
- Dependency picture after v0.1.1: **M2 depends on nothing from A1** (the guard reads a command line
  and answers), and only M1 and M3 touch A1 surfaces. Plan phase claims no implementation.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
