# Progress — SPEC-TODO-LANDING-EVIDENCE-001

Card: **t359** · Worktree: `.claude/worktrees/t359` · Branch: `WT-landing-evidence`
Tier: **L** (5 artifacts) · 21 requirements · 21 acceptance criteria · **v0.3.0**

## §E.1 Plan-phase Audit-Ready Signal

**Trajectory**: iter-1 FAIL 0.80 → iter-2 **PASS-WITH-DEBT 0.89** (Tier L threshold 0.85). All seven
must-pass criteria pass in both rounds; all thirteen iteration-1 defects RESOLVED; no stagnation, no
scope reduction. Reports: `.moai/reports/t359/plan-audit.md` @ `2f1c36151`,
`plan-audit-iter2.md` @ `8a1ae5b70`.

**v0.3.0 — iteration-2 debt closure (3 blocking + 4 non-blocking, all addressed):**

| Defect | Disposition |
|---|---|
| E1 (major) | AC-TLE-020's boundary clause could not fail — it varied card *text* while the predicate keys on card *id*. Replaced with one `--sha` against **two card ids**, asserted on an accepting and a refusing branch |
| E2 (major) | AC-TLE-018 (b) was a coverage assertion, not a detector; the "fail independently" claim **withdrawn**. New clause (c) asserts the frozen replica equals the live `backlogDDL` + accepted-version set, supplying (b)'s independent RED. §G gains forward drift |
| E3 (minor) | AC-TLE-020 case (d) added — the cannot-be-run branch REQ-TLE-020 binds, with the write-refuses vs read-degrades contrast made explicit |
| E4 (non-blocking) | status block re-pasted with full paths; the non-argument output ordering disclosed |
| E5 (non-blocking) | `plan.md` §B renumbered 1-5 |
| E6 (non-blocking) | AC-TLE-021 re-renders and counts its own row; `7` demoted to a note |
| E7 (non-blocking) | §G's "nothing flags it" **withdrawn**; the commit-subject mitigation recorded as considered-and-not-built, with record-time capture named as the only §D-compatible shape |

**Counts**: unchanged at 21 requirements / 21 criteria (Tier L ceiling 25/25). This delta strengthened
two criteria and added one case; it added no requirement.

**Preserved from v0.2.0** (iteration-2-confirmed, deliberately not touched): D5's three plants,
reproduced by the auditor and measured independent (`ITEMS-ONLY GUARD TRIPS=false`,
`ARCHIVED GUARD TRIPS=true`; the name-set assertion does not separate 019c while the tuple assertion
does) — so D13 is genuinely closed rather than recorded; the "record checks itself" argument for
reachability at the requirement level; and the guard-ownership decay hand-off sitting inside
AC-TLE-019 where run-phase reads it.

**Open for the Implementation Kickoff Approval gate** — three decisions, none a defect:
1. The stored shape (§B.1 — one JSON-bearing column versus four scalar columns).
2. The seventh-column contract change (§G, inherited from half A).
3. **New at v0.3.0**: whether to capture the delivering commit's subject at record time (§G, E7).
   It would add a seventh fact to the record and change REQ-TLE-016's cell format, so it is a design
   decision rather than a debt fix — surfaced here rather than folded in.

**Not resolved** (recorded, not closed): no criterion runs a genuinely older binary against a
post-change database — the backward divergence between the frozen replica and a released build is
undetected by construction (§G); the `--sha` validation commands are named but were not executed in
this tree (`research.md` §R.10.5); a reachable-but-wrong operator SHA stays machine-undetectable, and
the human-noticing mitigation is recorded but not built.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
