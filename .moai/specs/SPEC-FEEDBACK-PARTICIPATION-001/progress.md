# Progress — SPEC-FEEDBACK-PARTICIPATION-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: pending-audit (iteration 5, delta round)
- plan_complete_at: 2026-10-07
- Artifacts reflect version 0.5.2 of the spec. Versions 0.3.0-0.5.1 carried the plan through plan-audit iterations 1-4 (0.62 → 0.73 → 0.79 → 0.86; the iteration reports are local gitignored files under `.moai/reports/t1498/` and are not committed). Version 0.5.2 is the iteration-4 delta repair, one clause: D40 — the lock-lifecycle age-only break alternative deleted; the stale-lock break fires only on a verified-dead owner and a live owner always blocks (an age-only break could discard a live slow owner's committed mutation); AC-018 pins the invariant in place with the live-owner arm and `TestLiveOwnerOfAgeExceededLockStillBlocks`; the acceptance preface broadens "user-scoped consent store" to the user-scoped stores. No new REQ or AC (25/25); DEC-7, DEC-8, and the identifier rename are intact; D35 remains the recorded open operator decision.
- Open clarifications: zero. All former markers are recorded decisions in plan.md section B (DEC-1 to DEC-8).
- SPEC ID rename: DONE (v0.4.0). The directory, the frontmatter `id`, and every cross-reference carry `SPEC-FEEDBACK-PARTICIPATION-001`; the former identifier `SPEC-FEEDBACK-ANON-PARTICIPATION-001` (audit finding D22) is retired and appears only in this line and in the spec HISTORY 0.4.0 row. The branch name stays `WT-feedback-optin-anon`.
- Branch state: rebased onto `origin/main` `5a9d34fbb`; the SPEC commits are `a8bf27c49` (v0.2.0), `c34e24cab` (v0.3.0), `647b5e789` (v0.4.0), `28a4a16bd` (branch-base line fix), `c957ecc9d` (v0.5.0), and `be75563eb` (v0.5.1). The v0.5.2 delta lands as the next commit on `WT-feedback-optin-anon`.
- Rebase re-pin (finding D29): E2, E2p, and E21 carry their re-measurement pins; E2p was refreshed again in this delta (D39: 22 on the committed tree at `c957ecc9d`, historical series recorded in the entry).
- Optional findings D31-D35: D31 carve-out clause added to REQ-ANON-002; D32 walk-authoritative inventory note added and the two post-rebase recover sites named in design section 2 and plan M3; D33 abbreviation rule noted in the AC matrix header; D34 left as self-disclosed (the E11 note carries its own caveat); D35 (the `REQ-ANON-` requirement token family) is presented for an operator decision and is unchanged by this revision.
- gh title-token search feasibility (finding D23) is recorded as the M5 first-test item in plan.md section C.

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
