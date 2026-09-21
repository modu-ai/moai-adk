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

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-21
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
