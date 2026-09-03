# Progress — SPEC-TODO-LANDING-EVIDENCE-001

Card: **t359** · Worktree: `.claude/worktrees/t359` · Branch: `WT-landing-evidence`
Tier: **L** (5 artifacts) · 21 requirements · 21 acceptance criteria · **v0.3.1**

## §E.1 Plan-phase Audit-Ready Signal

**Trajectory**: iter-1 FAIL 0.80 → iter-2 **PASS-WITH-DEBT 0.89** → iter-3 **PASS-WITH-DEBT 0.92**
(Tier L threshold 0.85). All seven must-pass criteria pass in every round; all thirteen iteration-1
defects and all seven iteration-2 defects RESOLVED; no stagnation, no scope reduction, no score
regression. Reports: `.moai/reports/t359/plan-audit.md` @ `2f1c36151`, `plan-audit-iter2.md` @
`8a1ae5b70`, `plan-audit-iter3.md` @ `c0cfb2520`. The audit ceiling of three iterations is spent —
there will be no confirming re-audit, which is why iteration 3's blocking items were closed as a
documentation revision (v0.3.1) rather than carried forward.

**v0.3.1 — iteration-3 closure (3 blocking, all closed here; nothing carries into run-phase):**

| Defect | Disposition |
|---|---|
| F1 (minor) | AC-TLE-020's boundary RED was unconditional in text but fires only when the fixture ref's history mentions **exactly one** of the two card ids. The Given now pins that premise, the RED hangs on it, and both degenerate cases (neither id mentioned / both mentioned) are stated |
| F2 (minor) | AC-TLE-018 clause (c)'s second conjunct (frozen switch's accepted-version set equals the live one) **withdrawn as unrealisable** — the live set is control flow at `backlog_sqlite.go:286-298`, not extractable, and its cheapest runnable reading misses the drift it named. The DDL byte-identity conjunct, which carries the independent RED, is untouched |
| F3 (minor) | `research.md` §R.10.1's causal sentence attributed non-determinism to `grep -m1`, which has no such property (`/usr/bin/grep -m1`: 6/6 argument-order runs at HEAD `c0cfb2520`). The claim is **withdrawn** and replaced with the shell-function / parallel-grep cause; the block is labelled one capture of a non-deterministic ordering, and `spec.md` §A.3b's twin block now carries the same note |

**Run-phase debt from the plan-audit: none.** F1-F3 were the only blocking items and are closed
above. The unchanged residuals below are recorded limits of the SPEC, not audit debt.

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

**Open optional item, held by the operator** — **F4** from iteration 3, classed optional by the
audit and deliberately not fixed at v0.3.1: AC-TLE-020's case (d) groups its two sub-conditions
under one `unrunnable` stderr classification, while REQ-TLE-020 asks the stderr to name which check
failed. The two differ — a missing `git` makes **both** checks unrunnable, whereas an unresolvable
`--ref` leaves the existence check runnable and defeats only reachability. Case (d) stays
distinguishable from (b) and (c); it is only undistinguished internally. Fixing it is one added
clause requiring (d)'s stderr to name the unrunnable check. The operator holds this; it is not
run-phase debt unless they ask for it.

**Not resolved** (recorded, not closed): no criterion runs a genuinely older binary against a
post-change database — the backward divergence between the frozen replica and a released build is
undetected by construction (§G); the `--sha` validation commands are named but were not executed in
this tree (`research.md` §R.10.5); a reachable-but-wrong operator SHA stays machine-undetectable, and
the human-noticing mitigation is recorded but not built.

**Carried into run-phase as an observation duty, not as debt**: F1's fix is a stated premise, and a
stated premise is not a built fixture. With the audit ceiling spent there is no confirming re-audit,
so the only place the premise can be checked is M3's own RED observation — plant the card-token leak
and watch the boundary clause actually red. Until that is observed, treat AC-TLE-020's boundary
clause as asserted rather than demonstrated. The same discipline generalises from the audit's
closing note: across three rounds the recurring defect was a clause that could not fail, so every
newly written assertion should be accompanied by the mutation that reds it before it is accepted.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
