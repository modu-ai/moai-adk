# SPEC-RED-NOW-THRESHOLD-001 — Progress

Card: t343 · Branch: `WT-red-now-threshold` · Tier: M
Plan-phase measurement tree: `a6bbbf82b` (every ledger entry E-01..E-17) · Current tree: `15453140a` (origin/develop absorbed, fast-forward, zero conflicts)

## §E.1 Plan-phase Audit-Ready Signal

- `plan_complete_at`: 2026-08-29
- `plan_status`: audit-ready (iteration cap reached; a third audit is the operator's call)
- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set)
- Requirements: **15** (REQ-RNT-001..015) · Acceptance criteria: **16** (13 release-blocking,
  3 regression-guard). Tier M ceiling 16/16 — the AC axis sits **exactly at** the ceiling; a
  further criterion requires tiering up or splitting.
- Measured: `grep -c "^\*\*REQ-RNT-" spec.md` → 15; ledger E-12 → 16; E-13 → 13; E-14 → 3.
- Open decisions: none.
- Known residuals, carried openly: mutant M-3 (survives inside a span), mutant M-6 (a command
  whose premise is false — not caught by L2), and the Go test's own non-execution (tool boundary).

## Audit history

| Iteration | Score | Threshold | Verdict |
|---|---|---|---|
| 1 | 0.75 | 0.80 | FAIL — 8 blocking defects (D1..D8) |
| 2 | **0.800** | 0.80 | FAIL — 1 critical blocking defect (N1); score axis cleared |

**0.75 → 0.800 per-dimension delta** (iteration 2 vs 1):

| Dimension | Iter 1 | Iter 2 | Driver |
|---|---|---|---|
| Clarity | 0.75 | 0.80 | "the command" defined; line numbers removed from REQ-RNT-011 |
| Completeness | 0.75 | 0.80 | continued-firing split into two axes; REQ-RNT-013/014 added |
| Testability | 0.50 | 0.60 | five green paths span-scoped; commands moved to the fenced ledger |
| Traceability | 1.00 | 1.00 | unchanged |

Iteration 2 recorded **zero false completion claims** (every asserted closure reproduced under
re-execution), **zero monotonicity violations**, and MP-8 self-applied reproducing 12/12.
Closure: 7 closed (D1, D2, D3, D6, D8, D9, D10), 3 partially closed (D4, D5, D7).
Reports: `.moai/reports/t343/plan-audit-iter1.md`, `.moai/reports/t343/plan-audit-iter2.md`.

## Iteration-2 defects (N1..N7) — what each closed

| Defect | Sev | Closed by |
|---|---|---|
| **N1** two correctives cancelled; scope predicate reached zero commands | critical | REQ-RNT-001 admits the **ledger as a named carrier** in the requirement layer; REQ-RNT-008's predicate made **carrier-independent** (cell / ledger entry / any fenced block); AC-RNT-008 gains `TestCommandScopeIsCarrierIndependent` + fixture `testdata/red_now/ledger/` + relocation mutant; **M-5** recorded in §D.2; `acceptance.md` §D.5 restated — the previous "regression-guard cells are inside the gate" claim was false and is replaced by a claim that holds |
| **N2** three/four element mismatch | major | REQ-RNT-002 and `plan.md` M1 both now say four (command, stdout, exit code, tree SHA) |
| **N3** AC-RNT-001 overclaim | major | Weakened to what holds: span-scoping defeats a token outside the span, **not** inside it (~41 lines); residual recorded as M-3 |
| **N5** expensive-command hole | major | REQ-RNT-013 gains a timeout disposition reusing the auditor's **existing** Bash bound — refusal + demotion, no new regime |
| **N6** broken cross-reference | minor | Boundary stated in §D.3 where it holds, with the reason no `spec.md` exclusion carries it |
| **N7** self-matching count command | major | Counts moved to ledger E-12/E-13/E-14, anchored `^\| \*\*AC-RNT-`; measured 16 / 13 / 3 |

## Correction round (post-iteration-2, coordinator-relayed)

| Item | Closed by |
|---|---|
| **C1** §A.2's refuting command named an existing test and proved nothing (survived both audits) | Evidence replaced with `-run TestMigrationParityDoesNotExistXYZ` → `ok … [no tests to run]`, exit 0. **Conclusion unchanged, evidence replaced**, recorded in new §A.2.1 rather than swapped silently |
| **C2** framed as one cell | Reframed to the measured census: **nine** criteria on one false premise; AC-TOSQ-011 excluded and asserted neither defective nor sound |
| **C3** L2 verdict could be read off an `ok` token | **REQ-RNT-015** + **AC-RNT-015**: verdict keys on executed-test count. `internal/hook/evidence_writer.go` `deriveFromOutputText` cited as the reason (`hasPrecisePass` returns before inspecting any count); the file is **not touched** — t341's surface |
| **C4** cross-card coordination | `plan.md` §H extended: nine-cell family, both lane-5 findings credited, reciprocal-citation note |
| **C5** failure chain unrecorded | **§A.2.2**: four actors, rule known and actively applied, defect survived every human-judgment layer → the corrective must be mechanical. Bounded honestly — **MP-8 would not have caught this incident either** (mutant **M-6**) |

## Monotonicity

No criterion was weakened, removed, or reclassified in either round. The same three
regression-guards persist (AC-RNT-002, -011, -012). N3 and C1 weaken *claims*, not criteria: N3
drops an immunity assertion while keeping every predicate; C1 replaces evidence while keeping the
conclusion. Scope grew monotonically: 12 → 14 → **15 REQ**, 13 → 15 → **16 AC**.

## Audit debt — `0.800` does not describe the current text

[HARD] **The `0.800` in the table above is the plan-auditor's score for the iteration-2 text, not a
measurement of the artifacts as they now stand.** SSOT for this record: `spec.md` §E.

- **Audits run since the last revision: none.** Iteration 2 returned **FAIL** at 0.800; the SPEC
  was then revised to close one **critical** defect (N1) and four **major** ones (N2, N3, N5, N7),
  plus a minor (N6), followed by a correction round that replaced the §A.2 evidence, reframed the
  census from one cell to nine, and added REQ-RNT-015 / AC-RNT-015. None of that has been audited.
- **Cap reached, no third iteration.** `.moai/config/sections/harness.yaml:77` sets
  `plan_audit_tier_ceilings: M: 2`. The operator ruled to accept within the cap; the coordinator's
  own recommendation was a third audit, and the ruling went the other way. The cost is carried
  explicitly rather than absorbed.
- **N1's closure was confirmed by the author and the coordinator — not by an independent audit.**
  That distinction is the point of the debt. §A.2.1 is the standing counter-example: a claim that
  survived two independent re-executing audits and was still wrong.
- **Reading `0.800` as this SPEC's current quality is the error the SPEC prohibits** — a value
  detached from what it measured. Attaching it to its subject is the same discipline REQ-RNT-001
  imposes on every RED cell.

**Not debt — design facts.** Mutant **M-6** (MP-8 re-executes a command but does not verify the
command measures its stated premise) is a deliberate, recorded boundary of the three-layer
mechanism, not an unpaid item; it is not scheduled for closure, and REQ-RNT-015 narrows only the
`ok … [no tests to run]` case by design. Mutant **M-3** and the Go test's own non-execution are
carried residuals, named where they hold. Closing any of them would add criteria against a 16/16 AC
ceiling — a tier decision belonging to the lead.

## Cross-SPEC state

`SPEC-TODO-LANDING-STATE-001` (card t331), the source of the release-blocking / regression-guard
precedent, has **landed** — `status: completed`, resolvable at
`.moai/specs/SPEC-TODO-LANDING-STATE-001/` on `15453140a`. Its §C sentence was re-read from the
landed copy and is byte-identical to the pre-landing quote (`acceptance.md:93-95`). It is now
citable by path; it remains a precedent, not a dependency.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
