# Progress — SPEC-FEEDBACK-PARTICIPATION-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: pending-audit (iteration 3)
- plan_complete_at: 2026-10-07
- Artifacts reflect version 0.5.0 of the spec. Version 0.3.0 was the revision after plan-audit iteration 1 (FAIL 0.62; MP-7 clarification gate) and version 0.4.0 the revision after the operator decisions and the identifier rename; the iteration reports are local gitignored files under `.moai/reports/t1498/` and are not committed. Version 0.5.0 repairs the iteration-2 blocking set (FAIL 0.73, D24-D30): attribution moved to capture with the verdict persisted in the spool line; the console description carries the full disclosure; exec-typed attribution rows are call-site-asserted; AC-014 rescoped with the create-path identity at AC-020; the `summary_requested` marker bounds the crash window with AC-018 covering it in place; the mirror byte-equality guard retired for the participation-absence invariant; the DoD PASS-line rule scoped to named-test lines. 25 requirements and 25 criteria unchanged; DEC-7, DEC-8, and the identifier rename are intact.
- Open clarifications: zero. All former markers are recorded decisions in plan.md section B (DEC-1 to DEC-8).
- SPEC ID rename: DONE (v0.4.0). The directory, the frontmatter `id`, and every cross-reference carry `SPEC-FEEDBACK-PARTICIPATION-001`; the former identifier `SPEC-FEEDBACK-ANON-PARTICIPATION-001` (audit finding D22) is retired and appears only in this line and in the spec HISTORY 0.4.0 row. The branch name stays `WT-feedback-optin-anon`.
- Branch state: rebased onto `origin/main` `5a9d34fbb`; the SPEC commits are `a8bf27c49` (plan artifacts, v0.2.0), `c34e24cab` (v0.3.0), `647b5e789` (v0.4.0), and `28a4a16bd` (branch-base line fix). The v0.5.0 revision lands as the next commit on `WT-feedback-optin-anon`.
- Rebase re-pin (finding D29): the rebase invalidated the `bb54f2903`-era tree pins. E2, E2p, and E21 were re-measured on the current tree and re-pinned to `28a4a16bd` (E2 still green: two `:0` lines, exit 1; E2p 21, exit 0; E21 flipped to exit 1 — the local `feedback.yaml` carries the maintainer's preserved `auto_submit: true` from commit `82677fd27`, so the byte-equality mirror guard is retired and AC-023 asserts the participation-absence invariant instead). Every other ledger entry was re-executed on the post-rebase tree by the iteration-2 audit (MP-8, all reproduce).
- Optional findings D31-D35: D31 carve-out clause added to REQ-ANON-002; D32 walk-authoritative inventory note added and the two post-rebase recover sites named in design section 2 and plan M3; D33 abbreviation rule noted in the AC matrix header; D34 left as self-disclosed (the E11 note carries its own caveat); D35 (the `REQ-ANON-` requirement token family) is presented for an operator decision and is unchanged by this revision.
- gh title-token search feasibility (finding D23) is recorded as the M5 first-test item in plan.md section C.

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
