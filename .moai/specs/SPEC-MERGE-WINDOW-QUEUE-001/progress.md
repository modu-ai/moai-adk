# SPEC-MERGE-WINDOW-QUEUE-001 — Progress

> Card t1479 · created 2026-10-03 by manager-spec (plan phase)

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: spec.md + plan.md + acceptance.md + design.md + research.md (Tier L set) +
  decision-index.md (`interview.decision_gate: on`) + progress.md.
- Branch `WT-merge-window-queue`, renamed in place; `git merge develop` (local) → "Already up to
  date" at `d7112d005` (tree `632f65b47aa5`).
- SPEC id regex check (Bash) → `PASS`; uniqueness: no `MERGE-WINDOW` entry in this tree's or the
  develop worktree's `.moai/specs/`.
- RED-now baseline cells E1-E10 measured on that tree (research.md §R1); E11 (no-`--wait`
  acquire fixture) captured on a build of this branch (Go code = `d7112d005`) and committed ahead
  of any code in `3bc274dac` (`.moai/reports/t1479/baseline-acquire-nowait/`).
- Decisions: Q1 = recorded operator approval; Q2-Q6 (v0.2.0) and Q8-Q13 (v0.3.0, plan-audit
  iteration 1) = leader decisions (mission contract 07d28c4b) in the verdict lines. Target v3.2.0.
- v0.4.0: Q14 (heartbeat 15 s / window 60 s / re-entry grace 120 s), Q15 (starvation counting +
  three-requeue bound), Q16 (non-test-command residual risk) recorded as leader decisions. No open
  decision blocks run entry.
- Plan audit iteration 1: FAIL 0.75 (`.moai/reports/t1479/plan-audit-iter1.md`, verbatim copy).
  v0.3.0 revision: 25 REQ / 25 AC, contiguous 001-025; push verb moved to SPEC-CANDIDATE-CI-001.
- Plan audit iteration 2: FAIL 0.69, regressed — STOP (`.moai/reports/t1479/plan-audit-iter2.md`,
  verbatim copy). v0.5.0 scope reduction per leader decision Q17: reserved tickets, `--slice` /
  between-slices, front-once, requeue counter and three-requeue rule removed; plain FIFO with the
  owner pid on tickets stamped onto the holder at promotion (N1). Q18 lane merge verb
  `moai integration merge --card <id>` folded into REQ-MWQ-017. N5 hand-off item (e) in research
  §R6 (for the leader to forward to t1478). 23 REQ / 23 AC. No open decision blocks run entry.
- Plan audit iteration 3: FAIL 0.75 (`.moai/reports/t1479/plan-audit-iter3.md`, verbatim copy;
  first ceiling hit). v0.6.0 delta per leader decision Q19 inside the auditor's fix_scope: one merge
  path (complete calls the merge step, adopts a prior landing), SHA pinning and per-cause failure
  exits with abort + clean check (else `hold`), integration target on tickets and copied at
  promotion; O1/O2 one-liners, O3/O4 noted in research §R5. 23 REQ / 23 AC.
- Plan audit iteration 4: FAIL 0.75, claude + codex agree (`.moai/reports/t1479/plan-audit-iter4.md`,
  verbatim copy). Operator decision Q20 (AskUserQuestion 2026-10-03): one more narrow delta round, no
  new REQ. v0.7.0: complete's card gates before develop moves; adoption tied to the branch's current
  tip and a valid record, with the clause order in REQ-MWQ-019; ancestry precondition plus a defined
  post-merge outcome (commit left, `hold`); holder check reads first and refuses before any queue
  mutation. 23 REQ / 23 AC.

- Final re-read (`.moai/reports/t1479/plan-audit-final.md`, verbatim copy): one critical (data loss
  on an added path colliding with an ignored/untracked file) fixed as REQ-MWQ-018 cause 13
  (thirteen codes, AC-MWQ-018 rows 13a-13d as of v0.9.0). Per the Q22 convergence rule the remaining non-critical
  findings are run-phase obligations:
  - **O1** — separate the adoption path's failure handling from the merge step's: a
    post-adoption transition failure must not release or alter a window held by another session
    (release only when the caller is the holder); add a test that a foreign holder's window record
    bytes are unchanged.
  - **O2** — run the pre- and post-merge clean checks as
    `git status --porcelain --untracked-files=all` (overrides `status.showUntrackedFiles`), with a
    regression case under `status.showUntrackedFiles=no`.
  - **O3** — implement AC-MWQ-018 row 8b by injecting the dirty state / autostash residue through a
    seam after the cause-12 pre-check passes (a pre-existing local edit would trip cause 12 first),
    and assert `hold` is written before release.
  - **O4** — pin the merge verb card gate's "version as read" to the same read REQ-MWQ-019 step 1
    uses.

- Delta re-read of `d468ff19c` (`.moai/reports/t1479/plan-audit-delta-d468ff19c.md`, FAIL 0.80,
  D1 critical + D2 major). v0.8.0, leader decision Q24 (last repair round, no new REQ): D1 — cause 13
  widened to added path, ancestor of an added path, and path beneath an added path (AC-MWQ-018 rows
  13a-13c; row 13b is the auditor's reproduction and the RED fixture the run phase writes first);
  D2 — plan.md M5 corrected to thirteen causes and the REQ-MWQ-017 pre-merge order. 23 REQ / 23 AC,
  cause count thirteen. Remaining findings are run-phase obligations:
  - **O2 (extended, D3/D6)** — by the REQ-MWQ-017 order the clean-worktree check (cause 12, run with
    `--untracked-files=all` per base O2) precedes the collision check, so a non-ignored untracked
    collision is refused as cause 12 by design and is shadowed there under every
    `status.showUntrackedFiles` setting. The regression expects cause 12 for that case. Cause 13's
    untracked branch is exercised at check level (a unit test or a seam that presents the untracked
    state after the cause-12 pre-check has passed, as in O3), together with the O2 clean-check
    regression run under `status.showUntrackedFiles=no`.
  - **O5 (D4)** — symlink collisions and case-insensitive filesystems (macOS default) are unspecified
    for cause 13's existence test; the run phase decides the handling (for example an `lstat` or
    `git ls-files`-based test rather than case-sensitive path equality), records the decision and its
    evidence in §E.2, and tests the check against a varied collision fixture set rather than only
    rows 13a-13d. Not a plan blocker; no REQ is created for it.

- Delta re-read of `9d9d5fffa` (`.moai/reports/t1479/plan-audit-delta-9d9d5fffa.md`, FAIL 0.88, D5
  major + D6/D7 optional). v0.9.0, operator-decided final repair round (decision-index Q25), no new
  REQ: D5 — REQ-MWQ-017 reads "path" as a leaf entry and counts a directory-to-leaf change as an
  added path; AC-MWQ-018 row 13d added (second RED fixture beside 13b). D6 — O2 extension reworded
  above to the REQ-MWQ-017 order. D7 — row 13b's RED-now cell carries a re-executable command
  sequence. 23 REQ / 23 AC, cause count thirteen.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
