# SPEC-GUARD-STATE-MODEL-001 — Progress (card t347)

## §E.1 Plan-phase Audit-Ready Signal

**Baseline: `52c3fe590`** @ `WT-guard-state-model` (worktree `.claude/worktrees/t347`), re-pinned at **v0.6.0**. Artifacts were first authored on `091966c55` @ `.claude/worktrees/t333`; every measurement was re-run on `52c3fe590` in the v0.6.0 round, and the old pin survives nowhere because no cell measures the historical tree.

**plan-audit iter-2 has run: PASS-WITH-DEBT 0.81** against the Tier M threshold 0.80, up from iter-1's 0.79 — the regression clause did **not** fire. Dimension scores: Clarity 0.78, Completeness 0.90, Testability 0.72, Traceability 0.88; all seven must-pass criteria passed. **Skip-eligibility was withheld despite the score**, on the ground that five blocking findings stood and one (D1) was a correctness defect in the normative artifact the implementation is built from. Report: `.moai/reports/plan-audit/SPEC-GUARD-STATE-MODEL-001-review-2.md` — present in this worktree at the time of writing but **gitignored by class** (`.gitignore:222`), so it does not travel with the branch; the findings are restated below rather than left behind a path.

**v0.8.0 is the iter-2 repair round, and iteration 3 is terminal** under the 3-iteration cap. Closed, in the auditor's prescribed order:

| Defect | Class | Closed by |
|---|---|---|
| **D1** — a declared-quiet subject with an aged run classified `STALE`; the derivation never consulted its excuse | blocking, major | Entry axis restructured (`spec.md` §C.2): the window test asks *is there a qualifying run inside the window?*, and the declared-condition test is a **single node immediately after it**, ahead of the aged-versus-absent distinction. Rows 4, 5 and 6 carry the new conditions and precedence; **AC-GSM-010 absorbs the verification as a third fixture**. No row, value, requirement or criterion added |
| **D2** — `module:` named a package that does not exist and omitted both M3 writes | blocking, major | Set to the measured write set `.moai/guards, internal/guardliveness, internal/hook`; `internal/cli` dropped on evidence (no milestone writes it — `plan.md` §C records the decision). The `internal/hook` write, previously recorded nowhere, is now in `acceptance.md` §D.7 beside the `guardliveness` one |
| **D3** — AC-GSM-014(b)'s instrument still could not falsify its own mutant | blocking, major | Instrument replaced with a content-bearing comparison; `git status --porcelain --ignored` retained as the tracked-path half only. Re-probed on `52c3fe590` in this round — see the measurement note below |
| **D4** — AC-GSM-016(a) was a presence check where REQ-GSM-013 demands equality | blocking, major | Clause (a) restated as an **equality against the `Surface fold` column for all seven values**, which subsumes (c) |
| **D5** — the derivation AC-GSM-016 prescribed counted HISTORY rows | blocking, major | Replaced with a value-set derivation (→ 7), the row derivation named separately (→ 8), and the cell now says which of the two it means |
| **D6** — two cited audit reports do not resolve on this branch | blocking, minor | `plan.md` §H no longer claims "enumerated there in full"; both citations now state the gitignore-by-class cause and redirect to the artifacts that do travel |
| **D7** — AC-GSM-002's heading still read "four fields" | blocking, minor | Heading → five, **and restated as a derivation** against REQ-GSM-002 |
| **D8** — `progress.md` recorded a "7-row table" | blocking, minor | → derived, `grep -c '^  \| [0-9] \|' spec.md` → 8 |
| **D10** — the row 8 / row 1 interaction was unstated | optional | Taken: recorded in `acceptance.md` §D.7 beside the guard's vacuity bound, with the measurement that the ordering already produces the right answer |
| **D9** — AC-GSM-007's heading does not describe its clause (d) | optional | **Declined.** The auditor said it would not route a round for it, and the §D matrix row already carries the full description. Editing the heading to enumerate four clauses trades a cosmetic gain for a heading that no longer names what the criterion is *for*; the mismatch is recorded here instead |

**D3's measurement, run in this round rather than carried from the audit.** On `52c3fe590`: `git status --porcelain --ignored` collapses an ignored directory to one entry — `.moai/logs/` and `.moai/state/context-usage/` are two such directory entries; the count is left underived because the listing moves whenever an ignored file is written, which is this note's own subject (`git status --porcelain --ignored | grep '^!!'`, directory subset `… | grep '/$'`) — so a file created **inside** an already-collapsed ignored directory produced a `diff` of the before/after listings with rc=0 — no delta — and growing that file from 6 to 41 bytes produced rc=0 again. The second probe establishes the **general** form independently of the collapsing: `git status` reports path status, never content, so it cannot decide "byte-identical" for any ignored path under any flags. The replacement was probed against the same escape: `find . -newer <stamp> -type f -not -path './.git/*'` returned exactly the file the porcelain listing missed. All probe files were removed; no residue.

**The D1 re-derivation was re-run after the edit, not just applied.** Walking the new entry axis from the top: row 8's leading split is unchanged and still decides ahead of rows 1-6; row 1 and row 2 are unchanged and mutually exclusive by row 2's own "kind has a reader" clause; rows 3, 4, 5 and 6 are now **pairwise disjoint by construction** — row 3 is *a qualifying run inside the window*, rows 4/5/6 are all *no qualifying run inside the window*, then row 5 is *excused* against rows 4/6 *not excused*, then row 4 is *runs exist* against row 6 *zero runs*. The widening of row 5 from "zero qualifying runs" to "no qualifying run inside the window" was the one edit that could have created a new overlap — with row 3, for a declared-quiet subject that had just fired — and it does not, because the two conditions are complementary halves of the same test. **The only simultaneously-satisfiable pair remaining on the entry axis is row 8 with rows 1-6**, which the leading split decides and which rows 2 and 8 state; the row 8 / row 1 case within it is now recorded in §D.7 (D10). **Downstream rows were checked individually rather than assumed unaffected**, and the check found a real consequence: rows 4, 5 and 6 now sit behind more M1 fields than their `Flipped by` cells declared — row 5 and row 6 behind the window field and measured-quantity value they previously did not name, row 4 behind the conditional field. Three under-declared cells, invisible to AC-GSM-007 (c) precisely because that clause reads the column. All three corrected; **no M1 field was added**, and `spec.md`'s dependency paragraph moved from "three rows" to five accordingly.

**plan-audit iter-3 has run, it is TERMINAL, and v0.9.0 is a pre-run-phase correction carrying no fourth audit round.** Verdict **PASS-WITH-DEBT 0.86** against the Tier M threshold 0.80 — 0.79 → 0.81 → **0.86**, monotonic, the regression clause did **not** fire. Dimensions: Clarity 0.80, Completeness 0.92, Testability 0.82, Traceability 0.92; all seven must-pass criteria passed; the auditor verified every one of iter-2's D1-D8 and D10 closed, and measured three of the closures rather than reading them. **Skip-eligibility was withheld** on a single blocking finding, and the SPEC was judged **not fit to enter run-phase as written**. Report: `.moai/reports/plan-audit/SPEC-GUARD-STATE-MODEL-001-review-3.md` — gitignored by class (`.gitignore:222`), so it does not travel with the branch and its findings are restated here rather than left behind a path. The 3-iteration cap is reached: **there is no iteration 4.** The auditor sanctioned this correction without one, on the stated ground that its correctness is decidable by re-running the arithmetic rather than by an auditor's judgement — which puts the burden on the walk below, not on a verdict.

| Defect | Class | Closed by |
|---|---|---|
| **D1** — a degraded enumeration classified correct entries `ORPHANED`, and `REQ-GSM-010`'s guard was arithmetically incapable of firing on it | blocking, major | **Enumeration-integrity gate evaluated before the entry axis** (`spec.md` §C.2, REQ-GSM-010): for every declared disk-enumerable entry absent from the enumeration, a **direct existence test of that entry's own locator**; a subject found **present** fails integrity, **suppresses row 8 for every entry**, and carries the all-clear refusal. Row 8's cell and the derivation reconciled to one claim; `AC-GSM-012`'s `Given` corrected and clause (c) extended to assert a zero `ORPHANED` count alongside the integrity-check refusal. **No requirement, no criterion, no row, no value added; no M1 field added** — the locator is REQ-GSM-002's second field |
| **D2** — no criterion bound the JSON store round trip `plan.md` M3 names load-bearing | optional | Taken. `AC-GSM-015`'s `Given` widened by one clause to *any emitted result, **and the same result after a save-then-load cycle***, binding every clause of the criterion — and through the shared result, clause (d)'s additive expectation field, the one `plan.md` itself flags as the tag hazard — to the serialized form |
| **D3** — *"six such lines"* was one reading away from a stale figure, in the note whose subject is stale figures | optional | Taken. Replaced in both `acceptance.md` and this file by the two named **directory** entries plus the derivation command, with the count deliberately left underived because the listing moves whenever an ignored file is written. **Measured this round it reads 7, not 6** — the figure had already drifted, which is the argument for dropping it rather than restating it |
| **D4** — `AC-GSM-014(b)`'s `-newer` sweep asserts an empty result, unachievable in a live checkout | optional | Taken. The assertion is now stated over **the fixture tree the evaluator is run against**, not the development checkout. The auditor's own probe returned the harness's `.moai/logs/trace-*.jsonl` alongside the planted file; the fixture-tree reading was always intended and was never written down |

**The post-fix arithmetic walk on `AC-GSM-012` fixture (ii), which is the deliverable that replaces the fourth audit round.** Fixture: declared **18**, enumeration returns **3** (wrong glob), **all 18 subjects present on disk**.

1. **Integrity check runs first**, over the enumeration as a whole, before the entry axis. Enumerated 3 ≥ 1, so the zero clause does not decide it. For each of the **15** declared entries absent from the enumeration, the point test reads the entry's own locator: all **15 are present**. Integrity **fails** (15 misses > 0).
2. **Row 8 is suppressed for every entry.** The leading split is not evaluated, so the 15 continue down the axis: kind has a reader → query returned → a qualifying run inside the window → **row 3 `OK`**. `ORPHANED` per-value count = **0**, which is what clause (c) now asserts. The `Given`'s stipulation that every entry otherwise classifies `OK` is satisfiable again — under v0.8.0 it was not.
3. **The guard fires.** The all-clear is refused **on the integrity failure**, whose input is the point test's 15 present-but-unenumerated subjects — not on a zero test, and not on any count of the evaluator's own findings. Clause (c)'s *refused on the declared-vs-enumerated comparison, not on a zero test* is now an outcome the requirement produces rather than one it contradicts.

**Why the tolerance term is no longer circular for *any* partial enumeration, not merely for this fixture.** The v0.8.0 term was *(declared − enumerated) > (`UNDECLARED` + `ORPHANED`)*. Under the ungated axis an entry absent from the enumeration is **exactly** what produces an `ORPHANED` finding, so for any partial enumeration missing *k* declared subjects the left side is *k* and the `ORPHANED` term is *k*: the difference is **identically zero for every k**, and a strict inequality over it can never hold. The defect was therefore structural, not fixture-specific — the guard was unfirable by construction on the whole class of case it named. The v0.9.0 term reads **no classification at all**: its input is a per-subject point test of a named locator, a second and different measurement of the same fact. It cannot inherit the enumeration's defect, because it does not run the enumeration — the glob's pattern, traversal and filtering are all bypassed — and the value it returns is not a function of any finding the evaluator emitted.

**Both directions on row 8, checked rather than assumed — the false `ORPHANED` is closed and the true one stays reachable.** Sound enumeration, 18 declared, one workflow genuinely deleted, enumeration returns **17**: the single absent entry's point test finds the subject **absent**, corroborating the enumeration; misses = 0, integrity **passes**, the leading split is evaluated, and the entry classifies **`ORPHANED`** with *reconcile the manifest*. Row 8 is therefore reachable on exactly the case it was added for, and `AC-GSM-011` fixture (ii) — a manifest entry whose named workflow file is absent from disk — still passes unchanged. The mixed case is decided deliberately rather than by omission: one subject truly deleted **and** others missed fails integrity and suppresses row 8 for **all** entries including the true orphan, because an enumeration known to be wrong is not a basis for advising a deletion, and the refused all-clear reports the degradation instead.

**One new residual is recorded rather than papered over** (`acceptance.md` §D.7): the point test is independent of the enumeration's **pattern**, not of its **root**. A wrong root resolves both the glob and the locator against the same wrong tree, so corroboration agrees and integrity would pass. REQ-GSM-010's zero-file clause is stated as an *integrity failure* rather than only an all-clear refusal precisely to close the realistic form — a wrong root enumerates nothing, so row 8 is suppressed there too. What stays uncovered is a root that is wrong **and non-empty**, where the check verifies the enumeration was complete for the tree it read but not that it read the right tree. No criterion exercises it and the AC budget is at its ceiling.

- Artifacts: `spec.md`, `plan.md`, `acceptance.md` (Tier M set) + this file.
- Requirements: 13 (Tier M ceiling 16). Acceptance criteria: 16 (ceiling 16 — at the ceiling). Both derived at v0.7.0, not transcribed: `grep -cE '^\s*-\s+\*\*REQ-GSM-[0-9]{3}\*\*' spec.md` → 13; `grep -c '^### AC-GSM-' acceptance.md` → 16. v0.7.0 added a **clause**, not a criterion.
- **plan-audit iter-1: FAIL 0.79** against threshold 0.80. *(Its report, `.moai/reports/plan-audit/SPEC-GUARD-STATE-MODEL-001-review-1.md`, does **not** resolve on this branch — `.moai/reports/plan-audit/` is gitignored by class, `.gitignore:222`, so a report written in another worktree does not travel with the branch. The path is provenance, not a readable source; iter-1's findings are enumerated individually in `spec.md`'s v0.4.0 HISTORY row, which is where a reader of this branch should go.)* All 7 blocking and all 4 optional closed at v0.4.0. The headline finding — a **totality hole in the SPEC built to prove totality**: no row covered a declared entry whose workflow file is gone from disk, and it produced a wrong value (`STALE`, or a false `OK`) rather than an empty cell. Closed by row 8 / `ORPHANED`, a bidirectional runtime set comparison, and by **stating the derivation method** beside the table so it can be re-run rather than trusted. Counts unchanged at 13 REQ / 16 AC — every fix extends existing text, since the AC budget is at its ceiling.
- **v0.3.0 carries the producing half of the sibling SPEC's D1 seam repair** (REQ-GSM-012 designator + AC-GSM-015(c)), landed in the same commit as its consuming half. Because the AC budget is at its ceiling, the clause extends AC-GSM-015 rather than adding a criterion.
- **One baseline as of v0.6.0 — the two-baseline apparatus is retired, not relabelled.** Up to v0.5.0 the artifacts carried two pins (`091966c55` for RED-now, `origin/develop` at `ec15ec2cd` for t326 citations) on the stated ground that the trees were diverged. Re-measured: `git merge-base --is-ancestor 091966c55 HEAD` → rc=0 and the same for `ec15ec2cd` → rc=0; every t326 surface resolves in this working tree. **plan.md B-7's claim is therefore false on this tree**, and B-7 is retired and replaced by B-9 (the landed consumer).
- **Reuse verdict (v0.2.0):** t326's fold discipline and leniency principle adopted with citation; its **value vocabulary does not fit** and the measurement showing why is recorded — **four of seven** values here have no counterpart there (the count moved with row 8's `ORPHANED` at v0.4.0), one of its four has none here. Leniency bounded at *meaningless* (`UNREADABLE` → `ok`) versus *incomplete* (`UNKNOWN`/`UNRESOLVED` → `warn`).
- Card: **t347** (issued by the lead; the sibling surfacing SPEC is card t333). Dispatch is a separate lead decision — this lane authored the plan-phase artifacts only.
- Every RED-now cell is pinned to `52c3fe590` and its command was re-run on that tree in the v0.6.0 round; no cell was carried across the scope reduction, and none across the re-baseline.
- **v0.6.0 — the consumer landed, so the seam is compiled rather than prose.** `spec.md` §B.1 records the `Producer` interface, the `Entry`/`Designation`/`Result` types, the JSON round trip, and the single wiring site (`internal/hook/session_start_guard_liveness.go:38`) that M3 now replaces. REQ-GSM-012's designator obligation is discharged concretely as `Result.Clean.Values` holding exactly one value. Measured: a value-token grep over `internal/guardliveness/` + the wiring file returns rc=1 (no matches), so the vocabulary is still unnamed there and this SPEC keeps the freedom §B claims.

### F2 — a clause of the pre-split requirement set landed in neither SPEC (surfaced v0.6.0, CLOSED v0.7.0)

**Verdict at v0.6.0: one clause is vacuous. No whole requirement is.** It was surfaced rather than absorbed, because absorbing a dropped clause into the artifact that dropped it is how the predecessor's repairs relocated defects instead of closing them. **Closed at v0.7.0 by the lead's decision**, recorded at the end of this section together with the measurement that decided how.

**Derivation (mechanical, not recalled).** The pre-split set is the 16 requirements of `SPEC-GUARD-LIVENESS-001` v0.6.0, enumerated from git rather than from memory:

```
$ git show 091966c55:.moai/specs/SPEC-GUARD-LIVENESS-001/spec.md \
    | grep -cE '^\s*-\s+\*\*REQ-GDL-[0-9]{3}\*\*'
16
```

Walking that set against the sibling's own `spec.md` §B.2 mapping table and against the current artifacts of both SPECs:

| Pre-split (`v0.6.0`) | Subject | Landed in |
|---|---|---|
| REQ-GDL-001 | manifest, one entry per workflow file, outside `.moai/config/` | GSM — REQ-GSM-001 |
| REQ-GDL-002 | entry fields | GSM — REQ-GSM-002 (extended to five with `kind`) |
| REQ-GDL-003 | measured-quantity vocabulary | GSM — REQ-GSM-003 |
| REQ-GDL-004 | undeclared workflow file → `UNDECLARED` | GSM — REQ-GSM-008 + table row 7 |
| REQ-GDL-005 | release-cycle-quiet declared condition | GSM — REQ-GSM-004 |
| REQ-GDL-006 | per-subject query, never a global listing | GSM — REQ-GSM-005 |
| REQ-GDL-007 | zero runs in the retained window → `UNKNOWN`; **and not reported as "never fired"** | GSM — table rows 5/6 (**negative clause: see below**) |
| REQ-GDL-008 | most recent run older than the window → `STALE`, **and name the expectation it missed** | GSM — table row 4 (**second clause: NEITHER**) |
| REQ-GDL-009 | closed vocabulary + carried counts | GSM — REQ-GSM-007 + REQ-GSM-009 |
| REQ-GDL-010 | no all-clear while the queried count is zero | GSM — REQ-GSM-010 |
| REQ-GDL-011 | pull-based/attended + unconditional invocation | GDL — REQ-GDL-002 + REQ-GDL-003 |
| REQ-GDL-012 | evaluator mutates nothing | GSM — REQ-GSM-011 |
| REQ-GDL-013 | trigger + no-operator-input | GDL — REQ-GDL-004 + REQ-GDL-005 |
| REQ-GDL-014 | measurement age | GDL — REQ-GDL-006 |
| REQ-GDL-015 | change-leading | GDL — REQ-GDL-007 |
| REQ-GDL-016 | doctrine clause | GDL — REQ-GDL-009 |

**At requirement granularity, none is vacuous — all 16 have a destination.** At clause granularity, two were checked and they resolve differently:

**Grep scope, stated because this file breaks it.** The two greps below run over the **deliverable-defining artifacts** — this SPEC's `spec.md`, `plan.md`, `acceptance.md`, plus the whole sibling SPEC directory — and deliberately **exclude this `progress.md`**, which quotes both phrases in order to report them. Run unscoped, each grep now matches its own finding.

- **REQ-GDL-007's negative clause (*"shall not report it as 'never fired'"*) is structurally discharged, not dropped.** Zero matches in the scope above (rc=1), but the closed 7-value vocabulary (REQ-GSM-007) contains no "never fired" value, so no evaluator conforming to the table can emit one. The prohibition has nothing left to prohibit — the wording is gone because the state it forbade is unrepresentable.
- **REQ-GDL-008's second clause (*"and name the expectation it missed"*) landed in NEITHER SPEC, and nothing discharges it.** Zero matches in the scope above (rc=1); `grep -rn 'missed'` over this SPEC's three deliverable artifacts returns one hit, an unrelated `plan.md` sentence. Table row 4 classifies `STALE` and gives the implied action *"Investigate why the subject stopped firing"* — it does not require the result to name the window the subject missed. REQ-GSM-009's carried counts are per-value totals and carry no per-entry expectation either. **An operator reading a `STALE` entry is told to investigate and is not told what it missed.** *(Closed at v0.7.0 — see the repair below. The narrower residue that survives it: the entry now **carries** what it missed, and the sibling's render still does not **show** it.)*

**Repaired at v0.7.0, by the row-4 cell extension.** Two dispositions were available — a new `REQ-GSM-014`, or extending the row-4 cell of the normative table — because the AC budget is at its ceiling (16/16, derived: `grep -c '^### AC-GSM-' acceptance.md` → 16) while the requirement budget has headroom (13/16, derived: `grep -cE '^\s*-\s+\*\*REQ-GSM-[0-9]{3}\*\*' spec.md` → 13).

**The lead chose the cell extension, and the reason is the more useful half of the decision.** The headroom made a new requirement look like the cheaper option, but a fourteenth requirement would have had **no acceptance criterion to verify it** — the AC budget could not fund one. An unverified requirement is the same shape as the clause that went missing: present in the text, absent from verification. Inside the table, the obligation is exercised by a fixture that already runs.

**The caveat attached to that choice was live, and it was measured rather than assumed.** The lead's condition was that the extension be verified by an existing criterion — "a cell that grew while the AC still reads only its front half puts the same defect one layer down". Walking the mutant *a classifier that emits `STALE` and names nothing* across all 16 criteria on `52c3fe590`, it survived **16 of 16**: `AC-GSM-007`(a) inspects which classification an entry receives, `AC-GSM-008` inspects which values are reachable, and `AC-GSM-013` — the only criterion reading carried content — reads **result-level counts**, not per-entry data. The remaining thirteen read the manifest, the query path, the set comparison, the mutation surface, the designation, or the fold. **Extending the cell alone would have relocated the defect, not closed it.**

**So the repair is two edits, not one.** `spec.md` row 4's implied-action cell now requires the entry to carry the declared window and the declared measured quantity, copied from the entry's own fields; and `AC-GSM-007` gains **clause (d)** asserting the row-4 fixture case carries them, compared **field-against-field against the manifest entry** rather than checked for non-emptiness (a presence check is satisfied by a constant string). Post-extension the same mutant survives 15 criteria and is killed by (d) alone. Clause (d) flips at **M3** — the classification is M2's, the **carrier** is the emitted result shape, which is M3's — making `AC-GSM-007` the first criterion split across three milestones. Row 4's `Flipped by` cell says so, and its M1 dependencies are **unchanged**: the extension carries the same window and measured-quantity fields row 4 already depended on, so `AC-GSM-007` (c) is unaffected. Counts unchanged at 13 REQ / 16 AC.

**Two bounds recorded rather than closed** (`acceptance.md` §D.7). The published `Entry` (`internal/guardliveness/contract.go`) declares three fields — `Subject`, `Classifications`, `Surface` — and none can hold an expectation, so M3 discharges this with an **additive field on a type the sibling owns**. And the sibling's two render sites print `Subject` alone (`advisory.go:94`, `contract.go:160` `Render`), so the value is **carried and not shown**: the pre-split clause's operator-facing intent is half-discharged here and half unowned by either SPEC. That residue is narrower than the gap F2 found and it is real — it belongs against the sibling, not inside this cell.

**The general form, because the shape is what travels.** A split-mapping table whose row groups several source ids under one wildcard destination (`REQ-GDL-001..010, 012 → REQ-GSM-*`) cannot show the loss of any individual member: the row reads identically whether all eleven arrived or ten did, and nothing in it is false. Detection costs a per-item walk — which is what the sixteen-row table above now is, a reverse map on the receiving side.

### F3 — the re-derivation found a precedence defect (v0.6.0)

**Re-run, not reviewed** — per REQ-GSM-006's own instruction, and it paid for the second time.

**Finding.** A declared entry whose subject is gone from disk *and* whose query errored satisfies row 8's condition and row 2's simultaneously. The entry axis as written put the disk check **last**, so the entry decided row 2 — `UNRESOLVED`, implied action *"Retry; check credentials"* — for a subject no retry can bring back. This is the same permanently-non-retryable-under-a-retry-action defect that row 8 was added at v0.4.0 to close; the v0.4.0 carve-out scoped itself to *not-found* queries and left the error branch behind.

**Shape.** A **wrong value**, not an empty cell — the same signature as row 8's own discovery, and again invisible to a totality proof over rows: the table was complete and the reading order was wrong. Totality over rows and reachability of values say nothing about which row an entry reaches when two conditions hold at once.

**Closed by** leading the entry axis with the disk-absence test, widening the row-2 carve-out from *not-found* to any query outcome, and carrying the precedence in rows 2 and 8. **No row added, no value added, no criterion added** — counts unchanged at 13 REQ / 16 AC.

**Third input: asked and bounded, not assumed away.** The store's persisted-result round trip and the one-activation-stale property were the candidates. Measured on `52c3fe590`: `guardliveness.Activation` has exactly one field (`Root`) and `Produce` receives no prior result, so the persisted result reaches the **render**, not the classifier. It is a third input to the surfacing layer and the seam keeps it out of this one. Recorded as a bound in `acceptance.md` §D.7; the table was not extended.

**Residual, recorded in §D.7:** no criterion asserts the split ordering, because a fixture entry satisfying two rows at once would need a criterion and the AC budget is full.

### Origin

Created by the **scope reduction of `SPEC-GUARD-LIVENESS-001`** after its plan-audit iter-3 returned FAIL 0.667 with a STOP signal (0.800 → 0.800 → 0.667). The operator chose scope reduction — the audit's own recommendation — over a fourth repair round, which the regression clause forbids without an override.

The split ran along the seam the defects kept landing on: the surfacing model converged (both prior mutants re-run at iter-3, neither revivable), the state model did not (D2, D5, N2, T2, T4 are one family). This SPEC receives the state model.

### Inherited findings, carried as starting material

| Finding | Status |
|---|---|
| **T2** — a failed forge query has no admissible classification | Resolved in the state table, row 2 → `UNRESOLVED`; verified by AC-GSM-009 |
| **N2's unresolved half** — the fifth value repaired the no-reader hole and left the more common one open | Resolved structurally: totality is now demonstrated by construction over the state table (AC-GSM-007 + AC-GSM-008) rather than asserted in prose. **Row count derived, not transcribed** — `grep -c '^  \| [0-9] \|' spec.md` → **8**. *(Read "7-row" through v0.4.0, v0.5.0, v0.6.0 and v0.7.0: the pre-`ORPHANED` figure, and the third twin of the transcribed-count failure `plan.md` B-3 is about — after B-8 caught one and `plan.md` §F M2 caught a second. Stating the derivation instead of the number is what stops a fourth.)* |
| **T4** — the plan's vocabulary and milestone map went stale under a repair | Mitigated: the table is the single artifact both layers cite; criteria that genuinely split across milestones are clause-split in the flip lists |
| **T1** — the disk side of the set comparison had no integrity guard | Resolved: REQ-GSM-010 extends the all-clear refusal to a zero-file enumeration; verified by AC-GSM-012 |
| **T8** — two carried counts were consumed by nothing | Resolved: REQ-GSM-009 clause (b) requires every count to have a consumer; verified by AC-GSM-013 |

### Authoring method

Authored **as a state table, not as prose** — the auditor's stated reason the split helps. Prose is what let a five-value vocabulary cover at least six states without anyone noticing: a sentence defining a classification reads complete in isolation, while a table makes an uncovered condition visible as an empty cell. The table (`spec.md` §C.2 REQ-GSM-006) is normative; the prose describes it and does not extend it.

## §E.2 Run-phase Evidence

**M1 and M2.** M1 was measured on the tree entered at `52c3fe590`; **M2 is measured on `8ce725366`**, the M1 commit, at `WT-guard-state-model`, worktree `.claude/worktrees/t347`. M3 is unstarted; every criterion it flips is still RED for absence.

### M1 deliverables

| Artifact | Path | What it discharges |
|---|---|---|
| Manifest | `.moai/guards/liveness.yaml` | the census (18 entries) and the declared expectations |
| Schema + reader | `internal/guardstate/manifest.go` | the five fields, the closed measured-quantity vocabulary and its conclusion sets, the release-cycle-conditional form |
| Criteria | `internal/guardstate/{manifest,census,delivery}_test.go` | the M1 clauses below |

### AC matrix — M1 clauses only

| AC | Clause | Status | Verification command | Actual output |
|---|---|---|---|---|
| AC-GSM-001 | census, set-difference both directions | PASS | `go test ./internal/guardstate/... -run Census` | `ok … 0.176s` — `TestCensus_SetDifferenceEmptyBothDirections` passes; 18 declared against 18 enumerated, both differences empty |
| AC-GSM-002 | (a) five fields on every census entry | PASS | same | `TestCensus_EveryEntryCarriesFiveFields` passes over 18 entries |
| AC-GSM-002 | (b) missing measure / missing kind rejected **by name** | PASS | `go test ./internal/guardstate/... -run EntryValidate` | `TestEntryValidate_RejectsMissingFieldsByName` passes; each of the five absences returns its own named error |
| AC-GSM-003 | (a) three named values accepted | PASS | `go test ./internal/guardstate/... -run MeasureVocabulary` | `TestMeasureVocabulary_ClosedAtThreeValues` passes |
| AC-GSM-003 | (b) a fourth value rejected by name | PASS | same | same test; four distinct fourth values each return `ErrUnknownMeasure` |
| AC-GSM-003 | (c) three different qualifying sets | **M2** — not claimed here. The *data* (`Measure.Admits`) ships at M1 and is asserted distinct; the behavioural check over a recorded run fixture is M2's | `go test ./internal/guardstate/... -run MeasureAdmits` | `TestMeasureAdmits_ThreeDistinctConclusionSets` passes |
| AC-GSM-004 | both release-only subjects declare their condition | PASS | `go test ./internal/guardstate/... -run ReleaseOnly` | `TestCensus_ReleaseOnlySubjectsDeclareTheirCondition` passes; `release.yml` and `spec-status-auto-sync.yml` each carry `expected_when: release-cycle` |
| AC-GSM-005 | (b) second-kind entry accepted with no schema change | PASS | `go test ./internal/guardstate/... -run SecondKind` | `TestSecondKindEntry_AcceptedWithoutSchemaChange` passes; a `policy-rule` entry parses through the same struct and the same parser |
| AC-GSM-007 | (c) every row's M1 dependency present in the schema | PASS | `go test ./internal/guardstate/... -run StateTable` | `TestStateTable_EveryRowsM1DependencyIsInTheSchema` passes; 8 rows read from `spec.md`, all four named dependencies matched and carried |

### Invariants

| Invariant | Status | Evidence |
|---|---|---|
| Manifest outside `.moai/config/` (REQ-GSM-001, B-5) | HELD | `ManifestPath = ".moai/guards/liveness.yaml"` |
| Manifest not mirrored into `internal/template/templates/` (B-6) | HELD | no file added under that root; the commit's pathspec is `.moai/guards/` and `internal/guardstate/` only |
| No forge mutation, no push, no PR (REQ-GSM-011 at M1) | HELD | M1 adds no network call; `internal/guardstate` imports `os`, `regexp`, `time`, `errors`, `fmt`, `gopkg.in/yaml.v3` only |
| Subject-agnostic shape (REQ-GSM-001) | HELD | `Kind` and `Locator` are free strings; the second-kind fixture required no schema change — probed by hardcoding `Kind` to a workflow-only enum, which `AC-GSM-005 (b)` killed |

### Mutant probes (the two-cell RED half)

Each M1 criterion was observed killing the mutant it names. The naive implementation was written first and each failure is the criterion's own assertion, not a compile error.

| Probe | Mutant | Killed by |
|---|---|---|
| P0 | reader validates the locator only and defaults a missing measure to `fired-at-all`; `Admits` returns true for every value; `expected_when` declared but tagged `yaml:"-"`; `LoadManifest` unimplemented | AC-GSM-002 (b), AC-GSM-003 (b), AC-GSM-004, AC-GSM-007 (c), and the two census criteria — each on its own assertion |
| P1 | `Kind` hardcoded to a closed `github-workflow` enum (the hardcoded-subject mutant) | AC-GSM-005 (b) |
| P2 | `ExpectedWhen` declared on the struct but not read by the decoder | AC-GSM-004 and AC-GSM-007 (c) — three rows named the dependency and the schema no longer carried it |
| P3 | census of 18 entries **all naming `ci.yml`** (the count-equality mutant: 18 == 18) | AC-GSM-001, on the duplicate-locator assertion |

### The derivation, re-run rather than reviewed

Per REQ-GSM-006's stated method and `plan.md` §F M1, the entry axis was walked top-down against the fields M1 actually ships, not read off the table. Every split is decidable with what M1 delivers:

| Split | Reads | Shipped |
|---|---|---|
| disk-enumerable kind, subject absent from the enumeration, corroborated by a point test of the locator → row 8 | `kind`, `locator` | yes |
| kind has no reader → row 1 | `kind` | yes |
| query returned no result → row 2 | `locator` (to address the query); no M1 field beyond it | yes |
| qualifying run **inside the window** → row 3 | `window`, `measure` (the qualifying predicate) | yes — `Measure.Admits` is the predicate, which is why the conclusion sets ship at M1 rather than at M2 |
| declared condition excuses the absence → row 5 | `expected_when` | yes |
| qualifying runs exist at all → row 4, else row 6 | `measure`; row 4 additionally carries `window` + `measure` copied from the entry | yes — both are entry fields, so M3's carry is a field copy and adds no M1 field |
| disk axis: no entry names the file → row 7 | the declared locator set | yes |

No row was found unreachable and no M1 field was found missing. The walk did surface one thing the flip list alone would not have: **row 3's window test is only decidable if `measure` carries its conclusion set as data**, so shipping `Measure` as an inert string would have left rows 3, 4 and 6 undecidable while every field-completeness check still passed.

## §E.2 (M2) — the classifier

**Measured on `8ce725366`** (the M1 commit) at `WT-guard-state-model`. Every command below was run in this tree, in this run.

### M2 deliverables

| Artifact | Path | What it discharges |
|---|---|---|
| Vocabulary, fold, run history, the two seams, the entry axis | `internal/guardstate/classify.go` | REQ-GSM-003 (qualifying sets), 005 (the per-subject seam), 006 (the axis and its order), 007 (the closed vocabulary), 013 (the fold) |
| The integrity gate, the disk axis, the counts, the all-clear | `internal/guardstate/evaluate.go` | REQ-GSM-006 (gate before axis), 008 (both directions), 009 (the counts), 010 (the refusal), 011 (reads and reports) |
| Criteria | `internal/guardstate/{axis,evaluate,fold}_test.go` | the M2 clauses below |

### AC matrix — M2 clauses

| AC | Clause | Status | Verification command | Actual output |
|---|---|---|---|---|
| AC-GSM-003 | (c) three different qualifying sets over one fixture | PASS | `go test -count=1 ./internal/guardstate/ -run TestMeasure_` | `ok … 0.15s`; `TestMeasure_ThreeValuesYieldDifferentQualifyingSets` and `TestMeasure_FiredWithEffectIsDistinctFromVerdictRendered` pass — see the finding below: the stipulated fixture yields TWO distinct sets, and the second test supplies the separating run |
| AC-GSM-005 | (a) second-kind entry accepted, counted, `UNREADABLE` | PASS | `go test -count=1 ./internal/guardstate/ -run TestAxis_SecondKindEntryIsCountedAndUnreadable` | passes; `declared = 1`, classification `UNREADABLE`, and **zero** queries issued (row 1 precedes the query) |
| AC-GSM-006 | (a) N per-subject queries, (b) zero global listings | PASS | `go test -count=1 ./internal/guardstate/ -run TestEvaluate_PerSubjectQueriesOnly` | passes; 5 entries → 5 per-subject calls, one each, `globalCalls = 0` measured on a **callable** `AllRuns` |
| AC-GSM-007 | (a) one case per row, exactly one classification each | PASS | `go test -count=1 ./internal/guardstate/ -run TestStateTable_EveryRowClassifiesExactlyOnce` | passes; 8 cases, 8 single classifications, row count read from `spec.md` rather than transcribed |
| AC-GSM-007 | (b) table-driven, each value traces to its row | PASS | same | passes; each case asserts `Decision.Row` equals the row it was built for |
| AC-GSM-008 | every one of the seven values reachable | PASS | `go test -count=1 ./internal/guardstate/ -run TestStateTable_EveryValueIsReachable` | passes; all seven observed across the 8 cases (rows 3 and 5 both `OK`) |
| AC-GSM-009 | errored query is `UNRESOLVED`, not `UNKNOWN`, counted | PASS | `go test -count=1 ./internal/guardstate/ -run TestAxis_FailedQueryIsUnresolvedNotUnknown` | passes for both error shapes (auth failure, rate limit); `Counts[UNRESOLVED] = 1`, `Queried = 0` |
| AC-GSM-010 | (i) `OK`, (ii) `UNKNOWN`, (iii) aged declared-quiet is `OK` not `STALE` | PASS | `go test -count=1 ./internal/guardstate/ -run TestAxis_ExcusedAbsenceIsOKAgedOrAbsent` | passes; fixture (iii) classifies `OK` **and traces to row 5**, which is the assertion order mutant OM-2 breaks |
| AC-GSM-011 | (i) `UNDECLARED`, (ii) `ORPHANED`, both counted, neither all-clear | PASS | `go test -count=1 ./internal/guardstate/ -run TestEvaluate_SetComparisonBothDirections` | both subtests pass; fixture (ii) carries a FRESH run, so the false-`OK` direction is the one actually under test |
| AC-GSM-012 | (a) no all-clear, (b) enumerated count visible, (c) refused on integrity, zero `ORPHANED` | PASS | `go test -count=1 ./internal/guardstate/ -run TestEvaluate_DegradedEnumerationCannotReportAllClear` | all four subtests pass; `Counts[ORPHANED] = 0` on both degraded runs, and the healthy-enumeration subtest proves row 8 stays REACHABLE |
| AC-GSM-016 | (a) fold equals the column, (b) never `fail`, (c) the boundary | PASS | `go test -count=1 ./internal/guardstate/ -run TestFold_` | both tests pass; the fold is compared value-against-column **read from `spec.md`**, 7 distinct values over 8 rows |

### Invariants

| Invariant | Status | Evidence |
|---|---|---|
| The integrity gate runs BEFORE the entry axis (REQ-GSM-006, REQ-GSM-010) | HELD | `Evaluate` stage 1 precedes stage 2; order mutant **OM-1** moves it after and goes RED with 15 `ORPHANED` on fixture (ii) |
| The declared condition is consulted BEFORE the aged-versus-absent distinction | HELD | the `entryAxis()` slice order is 8, 1, 2, 3, **5**, 4, 6; order mutant **OM-2** swaps 5 and 4 and goes RED on AC-GSM-010 (iii) |
| Row 8 is suppressed on a failed integrity check, row 7 is not (B-2) | HELD | stage 2 reads `ev.IntegrityOK`; stage 3 does not. `Counts[ORPHANED] = 0` on both degraded fixtures while the disk axis still reports |
| The corroboration is a SECOND observation, never a term over the evaluator's own findings | HELD | the gate calls `enum.Exists` per absent locator; no classification count enters the decision. The forbidden term is identically zero for every partial enumeration and can never fire |
| No forge call in any test (B-7) | HELD | the query sits behind `RunQuerier` and every test uses `fakeQuerier`; the package imports no network package — `os`, `regexp`, `time`, `errors`, `fmt`, `context`, `gopkg.in/yaml.v3` only |
| No forge mutation, no working-tree write (REQ-GSM-011) | HELD | `Evaluate` takes two read-only seams and returns a value; nothing in the package opens a file for writing |
| Scope: `internal/guardliveness/` and the wiring site untouched | HELD | the working-tree status names no path under `internal/guardliveness/` or `internal/hook/` |

### Mutant probes (the two-cell RED half)

Each mutant was WRITTEN, its RED OBSERVED, and then reverted; both sources are byte-identical to their pre-mutant state (`cmp` against a pristine copy, after the last revert).

| Probe | Mutant | Observed RED |
|---|---|---|
| M2-P0 | the naive implementation: `Qualifying` returns every run, `Fold` returns `ok` for everything, `IsClean` returns true for everything, `Evaluate` returns an empty always-all-clear result | every M2 criterion failed on **its own assertion**, not on a compile error |
| M2-P1 | errored query classified `UNKNOWN` (row 2 → `ClassUnknown`) | AC-GSM-009 both shapes, **and** AC-GSM-008 — `UNRESOLVED` became unreachable |
| M2-P2 | catch-all default collapsing rows 1, 2 and 6 into `UNKNOWN` | AC-GSM-007 (a) and (b) on rows 1 and 2, **and** AC-GSM-008 on two unreachable values |
| M2-P3 | a repository-global listing called as a fast path | AC-GSM-006 (b): `1 repository-global run listings issued, want 0` |
| M2-P4 | an entry whose kind has no reader is DROPPED rather than counted | AC-GSM-005 (a): the evaluator decided 1 entry and none was the second-kind one |
| M2-P5 | all-clear computed without the non-clean term (findings listed, still green) | AC-GSM-011, both directions |
| M2-P6 | rows 3 and 5 classify `STALE` — the never-emit-`OK` mutant | AC-GSM-008: `value "OK" is produced by no case` |
| M2-P7 | `STALE` folds to `fail` | AC-GSM-016 (a) and (b) |
| M2-P8 | the borrowed leniency adopted WHOLESALE — `UNKNOWN`/`UNRESOLVED` fold to `ok`, and cleanliness read from the fold | AC-GSM-016 (a) and (c), **and** `TestFold_OnlyOKIsClean` on four clean values |
| M2-P9 | the integrity check reduced to a ZERO TEST, corroboration dropped | AC-GSM-012 on fixture (ii) only — 3 of 18 passed unremarked — **and** the healthy-enumeration subtest, because a zero test also cannot admit a true orphan |
| **OM-1** | **the integrity gate moved AFTER the entry axis** | `(c) 15 entries classified ORPHANED on a degraded enumeration` on fixture (ii), and 18 on fixture (i) — the predicted deletion advice for correct entries, at the predicted count |
| **OM-2** | **the declared-condition test moved AFTER the aged-versus-absent distinction** | `(iii) declared-quiet entry with an AGED run classified STALE: the declared condition is being consulted after the aged-versus-absent distinction, so retention selects the label` |

**What the two order mutants establish that no field or value mutant can.** Both leave every row present, every value reachable, and every field shipped; what they change is which row an entry REACHES when more than one row's condition holds. `acceptance.md` §D.7 records that no criterion asserts the ordering — that gap is now partially closed by mechanism rather than by criterion: OM-1 is killed by AC-GSM-012 (c)'s per-value `ORPHANED` count, and OM-2 by AC-GSM-010 (iii)'s row assertion.

**One measurement worth recording, because it confirms a claim rather than assuming it.** Under OM-2 the per-row fixture (`TestStateTable_EveryRowClassifiesExactlyOnce`) **did not fail** — only AC-GSM-010 (iii) did. Its row-5 case is a zero-run entry by construction, so it never reaches the cross. That is direct evidence for `acceptance.md`'s claim that fixture (iii) "was in no other criterion", and it means the totality proof and the reachability proof are **jointly blind** to this ordering.

### Findings raised, not closed here

- **AC-GSM-003 (c) is unsatisfiable as literally worded, on the fixture its own `Given` stipulates.** The `Given` names a fixture holding a `skipped`, a `cancelled`, and a `success` run; the `Then` requires "three different qualifying sets". Over that fixture `fired-with-effect` admits `{success}` and `verdict-rendered` admits `{success}` — **two** distinct sets, not three, and a mutant collapsing those two values survives the criterion. The separating case is a run that reached NO conclusion: `fired-with-effect` admits it (neither skipped nor cancelled), `verdict-rendered` does not. Implemented as two tests — the first asserts exactly the criterion's enumerated behaviour on exactly its stipulated fixture; the second adds the in-progress run and asserts the three set sizes genuinely differ. `acceptance.md` was NOT edited (M2 owns no SPEC body); the repair is a one-run extension of the `Given`, for whoever owns the next authoring round.

- **A separate-card candidate, deliberately NOT taken here.** §D.7 records that the integrity check verifies the enumeration was complete for the tree it READ, not that it read the right tree — a wrong-but-non-empty root is indistinguishable from mass deletion, because the point test resolves against the same root. Closing it needs an observation OUTSIDE both the enumeration and the point test (an anchor the root must contain, or a root identity the caller asserts), which is a **third input** and would reopen REQ-GSM-006's bounded totality claim. The residual is left open, as instructed.

## §E.2 (M3) — the result, its contract, and the wiring

**Measured on `45262e7a2`** (the M2 commit) at `WT-guard-state-model`. Every command below was run in this tree, in this run.

M3 is the milestone that writes a package this SPEC does not own. `internal/guardliveness/` belongs to `SPEC-GUARD-LIVENESS-001` (card t333, `status: completed`, with a landed test suite), and M1 and M2 deliberately did not touch it.

### The cross-SPEC write, and the three questions it actually had to answer

The compile question is the easy half and it was measured first: **every `Entry` literal in the sibling is keyed — 22 of 22, with 0 positional** — so an additive field compiles without touching a single site. Reproduced this run: `grep -rn 'Entry{' internal/guardliveness/` → 22; `grep -rn 'guardliveness\.Entry' internal/ cmd/ pkg/` → 2, both in `internal/hook/session_start_guard_liveness_render_test.go`; both keyed. That establishes nothing about behaviour, so:

| Question | Answer | Why it is not the other option |
|---|---|---|
| **Does the zero value mean something?** | A **nil pointer**: `Expectation *Expectation`. nil is *"this entry missed nothing"* | A value struct's zero is `{Window:"", Measure:""}`, which is indistinguishable from a carried-but-blank expectation. The two states have different meanings and a value struct has one representation for both — the same two-states-one-label defect this SPEC exists to close, one layer down. The pointer is also the representation the sibling **already** uses one type over, for the same distinction: `Result.Clean` is a pointer because *"a nil pointer is an absent designation — distinct from a present-but-null one"* |
| **What happens on the paths that never fill it?** | Three paths, each measured, none assumed | `Unwired()` returns `Result{}` with no entries — nothing to fill. The sibling's own 22 literals leave it nil and `Partition`/`Render`/`Advisory` never read it — the whole existing suite is the measurement (`go test ./internal/guardliveness/...` → ok, 56 PASS). A result persisted **before the field existed** decodes with the key absent: `TestEntryExpectationAbsentFromAnOlderPersistedResult` plants a hand-written pre-field snapshot — hand-written deliberately, since marshalling the current type would silently acquire the current tags and test nothing — and asserts it decodes, leaves the field nil, and still partitions |
| **Is the JSON tag right?** | `json:"expectation"`, verified by **save-then-load through the real store**, not by inspecting the in-memory struct | This is AC-GSM-015's widened `Given`. See the guard-count note under M3-P8 below: the round trip alone does **not** catch a symmetrically renamed tag, and that is measured rather than assumed |

**The sibling's suite stays green**, which is the clause that matters most on a cross-SPEC write: `go test -count=1 ./internal/guardliveness/... ./internal/hook/...` → all ok, `internal/guardliveness` 88.2% coverage, `internal/hook` 85.0%.

### M3 deliverables

| Artifact | Path | What it discharges |
|---|---|---|
| The producer, the disk enumerator, the forge query, the narrowing to the contract | `internal/guardstate/produce.go` (new) | REQ-GSM-005, 009, 011, 012; the M3 clauses below |
| Row 4's carried expectation, as a **column of the table** | `internal/guardstate/classify.go` | REQ-GSM-006 row 4 |
| The seven-value count seeding, and the refusal set that reads the carried counts | `internal/guardstate/evaluate.go` | REQ-GSM-009, 010 |
| The additive `Entry.Expectation` field and the `Expectation` type (**cross-SPEC write**) | `internal/guardliveness/contract.go` | the carrier AC-GSM-007 (d) reads |
| The stub replacement (**cross-SPEC write**) | `internal/hook/session_start_guard_liveness.go:38` | `plan.md` §F M3, "replace the stub at its single wiring site" |
| Criteria | `internal/guardstate/{produce,counts,partition_blind,partition_blind_audit,forge_query}_test.go`, `internal/guardliveness/contract_expectation_test.go`, `internal/hook/session_start_guard_liveness_wiring_test.go` | the M3 clauses below |

### AC matrix — M3 clauses

| AC | Clause | Status | Verification command | Actual output |
|---|---|---|---|---|
| AC-GSM-007 | (d) the row-4 entry carries the declared window and measured quantity, **equal field-against-field to the manifest entry** | PASS | `go test -count=1 -run TestProducedRow4EntryCarriesTheDeclaredExpectation ./internal/guardstate/` | passes. The comparison target is the manifest **read back through `LoadManifest`**, not a string literal in the test — comparing against literals would test the literals. The paired negative (`TestProducedNonRow4EntriesCarryNoExpectation`) asserts a row-3 and a row-7 entry carry **nil**, so the field cannot degrade into decoration present on every entry |
| AC-GSM-013 | (a) timestamp + enumerated + declared + queried + a per-value count for **each of the seven** | PASS | `go test -count=1 -run TestCarriedCountSetIsComplete ./internal/guardstate/` | passes; `len(Counts) == len(Classifications()) == 7`, asserted over the whole closed vocabulary rather than over the values the fixture produced |
| AC-GSM-013 | (b) **each** carried count is consumed by a named consumer that reads it | PASS | `go test -count=1 -run 'TestEveryCarriedCountFeedsAGuard\|TestMeasurementTimestampFeedsTheWindowComparison\|TestDroppedDeclaredEntryRefusesTheAllClear' ./internal/guardstate/` | passes, 4 subtests + 2. Measured per count by **changing it and requiring a decision to move** — see the consumer table below |
| AC-GSM-014 | (a) zero mutating forge calls | PASS | `go test -count=1 -run 'TestProducerMutatesNothing\|TestForgeQueryIsPerSubjectAndReadOnly' ./internal/guardstate/` | passes; `globalCalls == 0` on a **callable** `AllRuns`, and the real query's argv is measured token-exact: `run list --workflow <base> --limit 100 --json conclusion,createdAt` |
| AC-GSM-014 | (b) the working tree is byte-identical, by a **content-bearing** comparison | PASS | same | passes; a path+sha256 manifest over the fixture tree excluding `.git`, taken immediately before and after, compared. Asserted over the **fixture tree** (v0.9.0's correction), never the development checkout |
| AC-GSM-015 | (a) exactly one classification per entry, (b) exactly one designated clean value, on the in-memory result **and its round-tripped twin** | PASS | `go test -count=1 -run TestProducedResultHoldsTheContractBothInMemoryAndRoundTripped ./internal/guardstate/` | passes for both forms; (b) asserted on the **designation** — `len(Clean.Values) == 1` — not by counting how many entries happen to be clean |
| AC-GSM-015 | (c) a consumer given only the result, and no value names, can partition it — the harness's blindness **measured** | PASS | `go test -count=1 -run 'TestAValueBlindConsumerCanPartitionTheResult\|TestThisHarnessNamesNoValue' ./internal/guardstate/` | both pass. The harness is `partition_blind_test.go`; its auditor greps it for `\bOK\b`, `(?i)\b(stale\|unknown\|undeclared\|unreadable\|unresolved\|orphaned)\b`, `\bClass[A-Z]`, `\bSurface[A-Z]` and requires no match |

**Why the blind harness and its auditor are two files.** The first version put both in one file and it **failed on its own audit**: `the value-blind harness names the vocabulary: matched 6 time(s), first "stale"` — the matches were the auditor's own regex literal. An auditor has to name the vocabulary in order to forbid it, so an auditor living inside the harness is the very match it searches for. Split into `partition_blind_test.go` (the harness, greps clean) and `partition_blind_audit_test.go` (the auditor, which may name them).

### The count-consumer table (AC-GSM-013 clause (b))

Clause (b) exists because the parent card's audit found **two inert fields** — counts required, rendered, and read by nothing — and the disjunct that made the clause vacuous (*"or the published result's own contract"*, satisfied by any emitted field) was removed at plan time. So each count is traced to a consumer that **reads the carried field**, and the trace is measured by changing the count and requiring an outcome to move.

| Count | Consumer that reads it | Measured by |
|---|---|---|
| `MeasuredAt` | the window comparison — `Evaluate` feeds the **carried** field into every entry observation, so moving it moves which row an entry reaches | `TestMeasurementTimestampFeedsTheWindowComparison`: the same history classifies `OK` at +1d and `STALE` at +30d |
| `Enumerated` | the integrity gate's zero-file clause, and a refusal term | `TestEveryCarriedCountFeedsAGuard/the_enumerated-files_count` |
| `Declared` | a refusal term comparing it against the entries actually **decided** | `TestDroppedDeclaredEntryRefusesTheAllClear` + `.../the_declared_count` |
| `Queried` | a refusal term; counting only queries that SUCCEEDED is what makes it a guard | `.../the_queried_count` |
| the seven per-value counts | the non-clean term is summed **from the count map**, which is what gives all seven a reading consumer rather than only the values a run produced | `.../a_per-value_count` |

`AllClear` is now `len(Refusals()) == 0` rather than a separate expression, so the boolean and its stated reasons cannot drift apart (`TestAllClearMatchesTheRefusalSet`).

### Invariants

| Invariant | Status | Evidence |
|---|---|---|
| The wiring is UNCONDITIONAL — no branch short of the query layer | HELD | `session_start_guard_liveness.go:38` reads `= guardstate.NewProducer()`, a naming and not an `if p != nil`. `TestWiringIsUnconditional` counts the producer reached exactly once per refresh |
| The sibling SPEC's landed suite still passes | HELD | `go test -count=1 ./internal/guardliveness/... ./internal/hook/...` → all ok. 56 PASS in `guardliveness`, coverage 88.2%; `internal/hook` 85.0% |
| A result persisted before the field existed still decodes and still partitions | HELD | `TestEntryExpectationAbsentFromAnOlderPersistedResult`, on a hand-written pre-field snapshot |
| No forge MUTATION anywhere on the path | HELD | every argv the querier builds is `run list …`, asserted token-exact against a mutating-verb set. The type has two methods and both are listings |
| No write to the evaluated tree | HELD | the content-bearing digest is unchanged across a full evaluation; the persistence the consumer does lives under `~/.moai`, outside every evaluated tree |
| `Entry.Surface` is carried and never consulted | HELD | `grep -rn '\.Surface' internal/guardliveness/ internal/hook/ \| grep -v _test.go` → rc=1, no matches, unchanged from the pre-flight measurement |
| The clean designation is DERIVED from the vocabulary, not written down twice | HELD | `cleanValue()` scans `Classifications()` for `IsClean()`. A second literal would be a second place to be wrong, and designating a value `IsClean` disagreed with would break the seam silently |
| Scope: no SPEC body edited | HELD | `git status --short` shows `spec.md`, `plan.md`, `acceptance.md` carrying only their pre-existing v0.9.0 authoring; this milestone's diff to them is empty |

### Mutant probes

Each mutant was WRITTEN, its RED OBSERVED, and reverted; **every revert was verified byte-identical by sha256 against the pre-mutant source**, in the same step that applied it.

| Probe | Mutant | Observed RED | Guards that fired |
|---|---|---|---|
| **M3-P0** | **the wiring reverted to the stub** — `= guardliveness.Unwired()` | `the session-start site still names the stub producer: every activation reaches the seam and asks nothing, which is the absent-execution shape this family is about` | 1 — `TestWiredProducerIsNotTheStub`. Observed against the **genuine pre-wiring state**: the test was written and run before the wiring changed, so this RED is the real thing rather than a re-introduced mutant |
| M3-P1 | row 4 classifies and **names nothing** (`CarriesExpectation` dropped) | `the row-4 entry carries no expectation: it names the classification and not what was missed` | 1 |
| M3-P2 | the expectation is a **constant string** (`"expectation missed"`) | `carried window "expectation missed" != declared window "7d"` | 1 — this is why clause (d) is an equality and not a presence check |
| M3-P3 | the declared count made inert (`if false && …`) | `a declared entry was dropped and the all-clear stood`; `the declared count changed and no decision moved: the count is inert` | **2** |
| M3-P4 | the seven-value count seeding removed | `the count set holds 2 values, want one per classification (7)` | 1 |
| M3-P5 | the evaluator caches into `.moai/state/` **inside the evaluated tree** | `the evaluator wrote to the tree it evaluated: [created: .moai/state/guard-cache.json]` | 1 — and this is the exact escape AC-GSM-014's own history is about: the write is gitignored, produces no porcelain delta inside an already-collapsed ignored directory, and the content-bearing digest catches it anyway |
| M3-P6 | **two** clean values designated | `the designation names 2 values; a partition has no single referent`; `in memory: the designation names 2 values, want exactly 1` | **2** |
| M3-P7 | the designation not carried (`Clean: nil`, *"documented in spec.md"*) | `the result carries no designation, so a consumer must hardcode a literal to partition it` | **2** |
| M3-P8a | the JSON tag dropped (`json:"-"`) | `the expectation key is absent from the wire shape`; `the carried expectation did not survive the round trip: nil after load` | **2** |
| M3-P8b | the JSON tag renamed **symmetrically** (`json:"missed"`) | `the expectation key is absent from the wire shape: {…,"missed":null}` | **1 — and see the note below** |
| M3-P9 | an absent expectation emitted as a **blank struct** instead of nil | `.github/workflows/healthy-one.yml missed nothing yet carries an expectation {Window: Measure:}` | 1 |

**The guard count is reported per probe because a silence read without it inverts a conclusion, and M3-P8b is the case where it mattered.** Two guards sit at the JSON-tag point: the store round trip (which AC-GSM-015's widened `Given` mandates) and the wire-shape assertion. Under a **symmetric** rename the round trip stays **green** — marshal and unmarshal agree on the new key, so save-then-load is lossless — and only the wire-shape test fires. Had the round trip been the only guard, as the criterion's own wording would have left it, a renamed tag would have passed while every consumer reading the persisted file by key found nothing. Reported as a finding below rather than absorbed.

**Every M3 probe went RED, so no probe here required the "is this silence a missing guard, or another guard catching it?" disambiguation.** The counts are reported so the question is answerable by a later reader without re-running them.

### Findings raised, not closed here

- **AC-GSM-015's round-trip `Given` is necessary and not sufficient for the tag hazard it was added for.** v0.9.0 widened the `Given` to bind clause (d)'s additive field to the serialized form, on the stated ground that *"a field with a wrong or missing tag arrives at the consumer as a contract violation that never happened"*. Measured (M3-P8b): a save-then-load cycle **cannot** detect a symmetrically renamed tag, because both directions use the same tag. The hazard has two shapes — an **asymmetric** tag defect, which the round trip catches, and a **symmetric** rename, which only a key-name assertion catches. This milestone ships both guards; the criterion's wording still names only the first. A one-clause extension for whoever owns the next authoring round. No SPEC body was edited (M3 owns none).

- **The `gh` invocation itself is not covered, and the boundary is stated rather than implied.** `runGH` is two lines — that the binary is `gh`, and that it runs with the evaluated tree as its working directory — and it is reached only by an actual subprocess. Everything `list` does *with* an answer or a failure is covered through a seam (`forge_query_test.go`: a failed listing is an error and not an empty history, an undecodable one likewise, the conclusion/timestamp mapping, the empty-listing case, the working directory, and the deliberately-reachable global listing). What is **not** verified anywhere is that a real `gh run list --workflow` on a real repository returns the shape assumed here. That is an end-to-end observation this milestone did not take.

- **The `.moai/guards/liveness.yaml` census matches this tree, measured through the production enumerator.** A throwaway probe ran `Evaluate` against the worktree root with the real manifest and the real `diskEnumerator`: `declared=18 enumerated=18 queried=18 integrityOK=true allClear=true`, refusals empty, zero `ORPHANED` and zero `UNDECLARED`. So M1's 18-entry census and this tree's `.github/workflows/` agree exactly, and the locator form the manifest declares is the form the enumerator returns. The probe was removed; **it is not shipped as a test**, deliberately — it would turn any future workflow addition into a red build in an unrelated package, which is a policy decision this milestone does not own.

- **Two files were rewritten by the test run and restored.** `go test ./internal/hook/` rewrites `.moai/specs/SPEC-HOOK-PRETOOL-PERF-001/{baseline,postchange}.md` — a known behaviour of that package's perf fixtures, unrelated to this SPEC. Both were restored with an explicit-pathspec `git restore`; `git status --short` no longer names them.

- **Sixteen files under `internal/hook/` are gofmt-unclean and none of them is this milestone's.** Attributed mechanically rather than asserted: the intersection of `gofmt -l` and `git diff --name-only HEAD` is **empty**, so every unclean file is unmodified against `45262e7a2` and the condition pre-dates this milestone.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_milestone: M3
run_status: M1+M2+M3 complete — run-phase closed
run_baseline_head: 45262e7a2        # the M2 commit M3 was measured on
run_commit_sha: <backfill>
ac_pass_count: 23         # M1: 001, 002(a), 002(b), 003(a), 003(b), 004, 005(b), 007(c)
                          # M2: 003(c), 005(a), 006, 007(a), 007(b), 008, 009, 010, 011, 012, 016
                          # M3: 007(d), 013, 014, 015
ac_fail_count: 0
ac_deferred_count: 0
coverage_internal_guardstate: "92.2% of statements"      # was 92.1% at M2
coverage_internal_guardliveness: "88.2% of statements"   # the sibling's package, unchanged in shape
coverage_internal_hook: "85.0% of statements"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host: "go build ./... → rc=0"
  windows: "GOOS=windows GOARCH=amd64 go build ./... → rc=0"
  windows_test_compile: "GOOS=windows GOARCH=amd64 go vet ./internal/guardstate/ ./internal/guardliveness/ ./internal/hook/ → rc=0"
race: "go test -race -count=1 ./internal/guardstate/... ./internal/guardliveness/... → ok (1.721s, 1.974s); ./internal/hook/ → ok (37.138s)"
lint: "golangci-lint run --timeout=2m ./internal/guardstate/... ./internal/guardliveness/... ./internal/hook/ → 0 issues"
gofmt: "gofmt -l over the three packages → 16 files, ALL under internal/hook/ and ALL unmodified vs 45262e7a2 (intersection with `git diff --name-only HEAD` is empty). None is this milestone's; the condition pre-dates it"
subagent_boundary: "grep -rn 'AskUserQuestion' internal/guardstate internal/guardliveness | grep -v _test.go → rc=1, no matches"
sibling_regression: "go test -count=1 ./internal/guardliveness/... ./internal/hook/... → all ok"
tests_passing: 128        # 72 guardstate + 56 guardliveness (`--- PASS` lines, go test -v)
mutants_observed_red: 11  # M3-P0 (wiring) .. M3-P9; every one RED
mutant_residue: none      # sha256 before/after each apply-and-revert → identical, verified per probe
cross_spec_writes: 2      # internal/guardliveness/contract.go (additive field), internal/hook/session_start_guard_liveness.go:38 (the stub replacement)
total_run_phase_files: 16 # cumulative: 1 manifest + 4 sources + 11 tests (M3 adds 1 source + 7 tests, and modifies 4)
status_transition_performed: false   # see the note below
```

**The `draft → in-progress` transition was NOT performed, deliberately.** All four plan-phase artifacts carry substantial uncommitted v0.9.0 authoring in the working tree. Setting `status:` on them would require staging files whose uncommitted content this milestone did not author, which the sweep prohibition forbids. The transition is left to whoever lands the plan-phase artifacts.

**M2 does not perform it either, for the same unchanged reason.** The working tree at the M2 commit still shows all four plan-phase artifacts modified and unstaged; nothing about that changed between M1 and M2, so the decision is carried rather than revisited.

**M3 does not perform it either, and the reason is still unchanged rather than merely carried.** Re-read at the M3 commit: `git status --short` shows `spec.md`, `plan.md`, `acceptance.md` and this file all modified and unstaged, carrying the v0.9.0 authoring this run-phase did not write. This milestone stages `progress.md` by explicit pathspec because §E.2/§E.3 are its own artifact; setting `status:` on the other three would mean staging content it did not author. The `draft → in-progress` transition remains with whoever lands the plan-phase artifacts, and since run-phase is now closed it will land alongside them rather than ahead of them.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-31
sync_commit_sha: <backfill>            # a commit cannot know its own hash; backfilled in a follow-up commit
sync_baseline_head: 282dbfc38          # the M3 commit this phase was measured on
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-GUARD-STATE-MODEL-001' CHANGELOG.md → 1 BEFORE emission. NOT a duplicate of this SPEC's entry: the single hit is inside card t333's entry, which names this SPEC as the producer that was out of scope there. Emission proceeded; see the F2 record below"
b12_self_test_b: "grep -c '^### AC-GSM-' .moai/specs/SPEC-GUARD-STATE-MODEL-001/acceptance.md → 16. Non-zero, and it matches acceptance.md's own declared Tier M ceiling ('Count: 16 — at the ceiling'). The token-anchored form `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' | sort -u | wc -l` returns 17; the seventeenth is AC-GDL-001, a cross-reference to the predecessor SPEC and not a criterion of this one"
b12_self_test_c: "every path named in the CHANGELOG entry verified present with ls before commit: .moai/guards/liveness.yaml, internal/guardstate/{manifest,classify,evaluate,produce}.go, internal/guardliveness/contract.go, internal/hook/session_start_guard_liveness.go"
changelog_entry_position: "[Unreleased] › ### Added, first bullet"
ac_criteria_count: 16                  # criteria, the acceptance.md SSOT
ac_pass_count: 16
ac_fail_count: 0
tests: "go test -count=1 -cover ./internal/guardstate/... ./internal/guardliveness/... ./internal/hook/... → rc=0, every package ok"
coverage_internal_guardstate: "91.1% of statements"    # re-measured this phase; see the discrepancy note below
coverage_internal_guardliveness: "88.2% of statements"
coverage_internal_hook: "85.0% of statements"
vet: "go vet ./internal/guardstate/... ./internal/guardliveness/... → rc=0"
gofmt: "gofmt -l over internal/guardstate/, internal/guardliveness/ and this card's two internal/hook/ files → empty"
spec_lint_before: "moai spec lint <spec.md> → rc=0, no findings"
spec_lint_after: "moai spec lint <spec.md> → rc=0, no findings; full-tree scan raises exactly one OwnershipTransitionInvalid and it names SPEC-LSPMCP-001, not this SPEC — see the vacuous-silence note below"
frontmatter_status_transitions:
  spec_md: "draft → completed (single sync commit). in-progress and implemented never existed in history"
  plan_md: "n/a — carries no frontmatter"
  acceptance_md: "n/a — carries no frontmatter"
  progress_md: "n/a — carries no frontmatter"
  updated_field: "unchanged at 2026-08-31 — the v0.9.0 plan-phase authoring already set it to this date, which is the sync date"
canary_compliance_check: "n/a — this SPEC defines no forward-looking policy that its own sync tests"
graph_freshness_contribution: "6 described-source files (4 new: internal/guardstate/{manifest,classify,evaluate,produce}.go; 2 modified: internal/guardliveness/contract.go, internal/hook/session_start_guard_liveness.go) + 1 new package. NOT a delta claim — see the note below"
codemaps_regenerated: false
push: false
pr: false
```

### F1 — the `draft → in-progress` transition never happened, and it was declared rather than skipped

**Measured, not inferred.** `git log --oneline --all -S'status: in-progress' -- .moai/specs/SPEC-GUARD-STATE-MODEL-001/spec.md` returns empty: no commit ever set it. The finding as delegated read as an omission by `manager-develop`. It is not one, and the correction matters because the two have different remedies.

**§E.3 declares the refusal and states its reason**, and each of M1, M2 and M3 re-states it rather than carrying it silently: all four plan-phase artifacts held substantial uncommitted v0.9.0 authoring in the working tree, so setting `status:` on `spec.md` would have meant staging a file whose content that milestone did not author. That is the sweep prohibition applied correctly, not a step forgotten.

**One sentence in §E.3 is nevertheless contradicted by the commit it describes.** It reads *"This milestone stages `progress.md` by explicit pathspec because §E.2/§E.3 are its own artifact"*. `git show --stat 282dbfc38` lists twelve files and `progress.md` is not among them; neither is it in `8ce725366` or `45262e7a2`. §E.3 is `manager-develop`'s section and is not edited here — the contradiction is recorded rather than repaired.

**The consequence, which is larger than the missing transition.** **No run-phase commit touched any of this SPEC's four artifacts.** This sync commit is therefore the first commit to carry any of them, and it lands the plan-phase v0.9.0 authoring, the run-phase §E.1/§E.2/§E.3 evidence, and this section together. The alternative — refusing to stage them and returning a blocker — was rejected because it is mechanically impossible to land the `spec.md` frontmatter transition without staging that file whole, so a blocker here would have blocked the close itself while changing nothing about who authored what.

**How the SPEC is closed, stated rather than assumed.** `draft → completed` on this single commit. The intermediate transitions are not backdated and no artifact is edited to read as though they occurred; this record is the history.

**The ownership lint is silent about that transition, and the silence is vacuous.** `expectedOwnerForTransition` (`internal/spec/lint_ownership.go`) has cells for `draft → in-progress`, `draft → implemented`, `in-progress → implemented` and `implemented → completed`. It has **no cell for `draft → completed`**, so it returns `ownerNone` and the rule cannot produce a finding — the transition passes by being unmatched, not by being judged sound. Worth noting in the other direction: had this commit chosen `draft → implemented`, the matrix *would* have expected `manager-develop` and flagged this `docs(` commit. A rule whose stricter reading fires and whose looser reading is silent is the wrong way round, and it is left as a finding rather than worked around.

**A divergence from the sibling card in the same batch, surfaced rather than settled here.** Card t333 (`SPEC-GUARD-LIVENESS-001`) deliberately withheld `completed` at its own sync on the ground that its branch was unpushed and no CI had judged any of its commits. That reasoning applies verbatim to this card. `completed` is taken here because the dispatch directed it and because the canonical 3-phase close merges the transition onto the sync commit — but the batch is now inconsistent between two adjacent cards, and reversing this one to `implemented` is a one-line change if the lead prefers parity.

### F2 — a CHANGELOG statement became false when M3 landed, and the B12 grep is what surfaced it

The pre-emission `grep -c 'SPEC-GUARD-STATE-MODEL-001' CHANGELOG.md` returned **1**. The hit is not a duplicate entry for this SPEC: it sits inside card t333's entry, in the paragraph headed *"What this does NOT yet do, stated because the gap is the point"*, whose opening sentence states that `guardliveness.Unwired()` is what every production activation reaches, so on a real tree nothing is persisted and the advisory renders nothing.

**M3 closed exactly that gap**, so the sentence now tells a reader a shipped feature is inert when it is live — the larger error by some distance, and the reason it is corrected rather than left.

**Corrected by appending, never by rewriting.** The paragraph is preserved verbatim and a bracketed sentence is appended to it: `**[Corrected 2026-08-31 by card t347, appended rather than rewritten.]**` followed by what changed and by which card. Two properties are deliberate. It **preserves the record that the gap existed**, which is what motivated this SPEC and would be lost by deletion. And it **corrects only the sentence measured** — the same paragraph's claim about the unexercised async render branch was not re-measured here, so the appended text says so explicitly rather than implying the whole paragraph now stands.

### The two paragraphs this card was asked to carry as a durable finding

**(a) A compile error is not a mutant observation.** Reverting the wiring to `Unwired()` first produced an unused-import compile error, not a test failure. That is not an observation: anyone actually reverting the wiring removes the now-unused import too, and then it compiles. Only after removing the import did the faithful mutant run — and `TestWiredProducerIsNotTheStub` went red. Stopping at the compile error would have credited the compiler for what a test must guard.

**(b) `Unwired()` is not nil, which is why the guard tests behaviour rather than presence.** A nil check would pass with the stub reinstated. That is precisely the shape card t333 shipped — seam present, tests green, real tree rendering nothing, because no producer was ever wired. This guard calls `Produce` and asserts the result is not the stub's.

### Measurements that disagree with what was carried, recorded rather than reconciled

- **Coverage.** §E.3 records `coverage_internal_guardstate: "92.2% of statements"`. Re-measured in this phase on `282dbfc38` — `go test -count=1 -cover ./internal/guardstate/...` → `coverage: 91.1% of statements`. The dispatch that opened this phase also carried 91.1%. §E.3 is not edited; the sync-phase figure above is the one attributed to a command run against this tree in this phase.
- **Acceptance-criteria count, three figures in circulation.** The `acceptance.md` SSOT is **16 criteria**, derived (`grep -c '^### AC-GSM-'` → 16) and matching the file's own declared Tier M ceiling. §E.3 carries `ac_pass_count: 23`, which counts **clauses**, not criteria — several criteria are split (AC-GSM-007 alone carries four). The dispatch stated "AC 20/20", which matches neither figure. The count reported for this close is 16 criteria, 16 PASS, 0 FAIL.
- **Graph Freshness is reported as a contribution, not as a delta, because the delta is not attributable.** `moai graph check` in this worktree reads `codemaps metric=described-source-diff value=145 threshold=40 verdict=stale`; the develop head carries an inherited `value=45`. The difference is **not** this card's: the worktree base `52c3fe590` already carries other cards' drift, and that base was not measured. What is attributable is the file set — `git diff --name-only 52c3fe590..HEAD` filtered to non-test Go sources gives the **6 files** named in the YAML block above, four of them a new package. Codemaps were **not** regenerated (the batch regenerates once at the end, in one lane).

### Residual risk

- **Nothing here has been judged by CI.** The branch is unpushed and no commit on it has been built or tested anywhere but this machine. The `completed` status asserts a close that no integration has confirmed — the exact objection card t333 raised against its own `completed` and acted on differently.
- **The `gh` invocation itself remains uncovered**, as §E.2 (M3) records: no test observes that a real `gh run list --workflow` on a real repository returns the shape the querier assumes.
- **The census agrees with this tree and with no other.** `.moai/guards/liveness.yaml`'s 18 entries were verified against this repository's `.github/workflows/` through the production enumerator (§E.2 M3), by a probe that was deliberately not shipped as a test. A workflow added or removed after this commit puts the manifest and the disk out of agreement with nothing failing to say so — which the classifier reports as `UNDECLARED` or `ORPHANED` at the next session start rather than at CI.
