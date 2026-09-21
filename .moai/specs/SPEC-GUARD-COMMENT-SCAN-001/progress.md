# Progress — SPEC-GUARD-COMMENT-SCAN-001 (card t1056)

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **S** — one preprocessing step in one existing file plus one new test file; 8 REQ / 6 AC,
  both inside the 8/8 Tier S ceiling.
- Artifact set: `spec.md` + `plan.md` + `acceptance.md` + this `progress.md`. The separate
  `acceptance.md` departs from the Tier S convention (which inlines AC in `spec.md`) at explicit
  dispatch instruction; the deviation is recorded in `plan.md` §H rather than silently taken.
- Base measured against: HEAD `3dfae918a`, branch `WT-guard-prose`, worktree
  `.claude/worktrees/t1056`.
- Defect class: the third axis of "data is not a command" in the BranchGuard scan-preprocessing
  pipeline. Quoted arguments (`substituteQuotedArguments`) and heredoc bodies
  (`substituteHeredocBodies`) are already collapsed; shell comments are not.
- Narrowing change, so **both** mutation arms are mandatory acceptance criteria — Arm A
  (comment prose no longer matches) and Arm B (the guard is not blunted). Neither alone
  distinguishes a correct narrowing from a disabled guard.
- Discriminant recorded in `spec.md` §A: a refusal without the `BRANCH_GUARD_VIOLATION:` sentinel
  was not produced by this repository's guard. This reversed the card's original premise (E3), and
  E3 is excluded in `spec.md` §E with its measurement.
- Known Gap carried forward, not papered over: the over-match is reproduced at the matcher level
  only; hook-path reachability is unmeasured, and `plan.md` §D decides to keep it a Gap and names
  the file that would close it.
- Evidence file `.moai/reports/t1056/reproduction.md` is gitignored (`.gitignore:227`) and
  untracked — this worktree holds the only copy (`plan.md` §E).
- Plan-phase produced artifacts only: no implementation code, no branch creation, no push, no CI,
  no verification load.

### Plan repair, iteration 1 (plan-audit FAIL 0.625 → repaired)

The first plan-audit returned FAIL at 0.625 against the Tier S threshold 0.75
(`.moai/reports/t1056/verdict.md`, untracked — this worktree holds the only copy). Repaired at HEAD
`5bc42a304`, in repair order D4 → D1 → D2 → D3 → D5, because D4's decision determines D1's wording:

- **D4 (major)** — the manufactured-`#` residual is **ACCEPTED**, with its direction corrected from
  "under-match / fail-open" to **blinding**, and its reach corrected from one token to end-of-line.
  Recorded on the requirement surface (`spec.md` §F) as well as in `plan.md` §A, with both rejected
  repairs and a follow-up coordinate in `substituteQuotedArguments`.
- **D1 (critical)** — AC-GCS-002 gains rows 3-4, placing a comment and a must-survive branch-state
  command on the **same line**. Nothing in the pre-repair set fixed the *direction* of the elision.
  REQ-GCS-004 now also defines "comment run" and states that preceding text survives.
- **D2 (critical)** — AC-GCS-004 gains rows 3-4 with **non-alphanumeric** preceding characters
  (`/`, `.`), row 3 carrying a real command later on the same line. `spec.md` §B.1 carries the bash
  measurement plus a positive control.
- **D3 (major)** — the continuation-line case is recorded as a **blinding** residual in `spec.md`
  §F, and `plan.md` §C's inverted rationale is corrected (handling continuations would *narrow*,
  not widen, the elision). Not re-measured here: every probe form was refused by the Claude Code
  runtime worktree-isolation guard, so the figure is cited to the verdict with that status inline.
- **D5 (major)** — the third §A positive control is re-measured: **112**, not 3. The block's
  conclusion survives; only the attribution was broken.
- **D6 / D7 / D8 (optional)** — D6 recorded as an over-match residual (cited, not re-measured, same
  guard refusal); D7 partially taken (REQ-GCS-007 restated as an outcome, REQ-GCS-004 de-identified
  — REQ-GCS-008 left in place to avoid renumbering); D8 taken ("measurably" → "demonstrably").

The requirement ↔ criterion map in `acceptance.md` was **re-derived from the criteria as they now
read**, not edited incrementally: REQ-GCS-002 now maps to AC-GCS-001 + AC-GCS-004 (positive and
negative halves) and REQ-GCS-004 to AC-GCS-005 + AC-GCS-002, and the map carries an explicit
exclusion table naming the two pass-but-wrong implementations the set now rejects.

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-21
plan_repaired_at: 2026-09-21
plan_repair_iteration: 1
tier: S
req_count: 8
ac_count: 6
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
