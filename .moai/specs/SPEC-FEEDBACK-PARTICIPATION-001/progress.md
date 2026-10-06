# Progress — SPEC-FEEDBACK-PARTICIPATION-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: pending-audit (iteration 4, delta round)
- plan_complete_at: 2026-10-07
- Artifacts reflect version 0.5.1 of the spec. Versions 0.3.0-0.5.0 carried the plan through plan-audit iterations 1-3 (0.62 → 0.73 → 0.79; the iteration reports are local gitignored files under `.moai/reports/t1498/` and are not committed). Version 0.5.1 is the iteration-3 delta repair: D36 — every bugreport store (spool, queue, ledger, outbox log) moved to `<moai home>/state/bugreport/` so a repository-forged `verdict: moai` spool cannot drive publication, with the forged-project-spool arm folded into AC-003; D37 — the queue lock gains an owner-verified stale-lock break so the marker→unlock crash window cannot wedge every later mutation, with the kill-mid-Mutate fault case folded into AC-018; D39 — E2p's record refreshed to the committed-tree count (22). D38 skipped by scope decision. REQ/AC id sets unchanged (25/25); DEC-7, DEC-8, and the identifier rename are intact; D35 remains the recorded open operator decision.
- Open clarifications: zero. All former markers are recorded decisions in plan.md section B (DEC-1 to DEC-8).
- SPEC ID rename: DONE (v0.4.0). The directory, the frontmatter `id`, and every cross-reference carry `SPEC-FEEDBACK-PARTICIPATION-001`; the former identifier `SPEC-FEEDBACK-ANON-PARTICIPATION-001` (audit finding D22) is retired and appears only in this line and in the spec HISTORY 0.4.0 row. The branch name stays `WT-feedback-optin-anon`.
- Branch state: rebased onto `origin/main` `5a9d34fbb`; the SPEC commits are `a8bf27c49` (v0.2.0), `c34e24cab` (v0.3.0), `647b5e789` (v0.4.0), `28a4a16bd` (branch-base line fix), and `c957ecc9d` (v0.5.0). The v0.5.1 delta lands as the next commit on `WT-feedback-optin-anon`.
- Rebase re-pin (finding D29): E2, E2p, and E21 carry their re-measurement pins; E2p was refreshed again in this delta (D39: 22 on the committed tree at `c957ecc9d`, historical series recorded in the entry).
- Optional findings D31-D35: D31 carve-out clause added to REQ-ANON-002; D32 walk-authoritative inventory note added and the two post-rebase recover sites named in design section 2 and plan M3; D33 abbreviation rule noted in the AC matrix header; D34 left as self-disclosed (the E11 note carries its own caveat); D35 (the `REQ-ANON-` requirement token family) is presented for an operator decision and is unchanged by this revision.
- gh title-token search feasibility (finding D23) is recorded as the M5 first-test item in plan.md section C.

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
