# SPEC-GUARD-LIVENESS-001 — Acceptance Criteria (card t333)

Baseline tree for every RED-now cell: **`d34a789a4`** @ `WT-guard-liveness` (worktree `.claude/worktrees/t333`), equal to `origin/develop` at authoring time. Every RED-now cell below was measured on that tree and states **why** it is red, per `.claude/rules/moai/development/verification-completeness.md` §2 — RED-now alone does not distinguish a criterion this work can flip from one that is red forever.

Each criterion was put through that rule's §2 mutant probe. Where a single-clause criterion admitted a mutant satisfying it while violating its requirement, a second clause was added and the mutant that forced it is recorded in the criterion.

Budget: Tier M ≤ 16 acceptance criteria. **Count: 16 — at the ceiling.**

## §D AC Matrix

| AC | Requirement | Kind | Flipping milestone |
|---|---|---|---|
| AC-GDL-001 | REQ-GDL-001 | RED-first (artifact absence) | M1 |
| AC-GDL-002 | REQ-GDL-002 | RED-first + mutant clause | M1 |
| AC-GDL-003 | REQ-GDL-003 | Closed-vocabulary pair | M1 |
| AC-GDL-004 | REQ-GDL-005 | RED-first (declared quiet) | M1 |
| AC-GDL-005 | REQ-GDL-006 | Positive + negative pair | M2 |
| AC-GDL-006 | REQ-GDL-007 | RED-first (UNKNOWN state) | M2 |
| AC-GDL-007 | REQ-GDL-003, REQ-GDL-008 | RED-first (skipped rejection) | M2 |
| AC-GDL-008 | REQ-GDL-004 | RED-first (UNDECLARED state) | M2 |
| AC-GDL-009 | REQ-GDL-009, REQ-GDL-010 | Mutant-guard pair | M2 |
| AC-GDL-010 | REQ-GDL-011, REQ-GDL-013 | Positive + negative assertion | M3 |
| AC-GDL-011 | REQ-GDL-014 | RED-first (measurement age) | M3 |
| AC-GDL-012 | REQ-GDL-016 | Additive-only diff assertion | M4 |
| AC-GDL-013 | REQ-GDL-013 | Zero-input surfacing pair | M3 |
| AC-GDL-014 | REQ-GDL-015 | Change-leading pair | M3 |
| AC-GDL-015 | REQ-GDL-001 | Subject-agnostic shape pair | M1 |
| AC-GDL-016 | REQ-GDL-012 | Negative assertion (no mutation) | M2 |

## §D.1 Criteria (Given-When-Then, two cells each)

### AC-GDL-001 — the manifest exists and its census is complete
**Given** the repository tree; **When** the manifest is read and its entry count compared against `ls -1 .github/workflows/*.yml .github/workflows/*.yaml | wc -l`; **Then** an entry exists for every workflow file, and the two counts are equal (18 = 18 on this tree).
- **RED-now:** `ls .moai/guards` returns `No such file or directory` (rc=1) on `d34a789a4`. Red because the artifact does not exist, not because a comparison failed.
- **Green path:** M1 populates the census. Passing output is equal counts and an empty set-difference in both directions.
- **Mutant note:** the count-equality clause is stated as a *set* comparison, not a bare number. A mutant that writes 18 entries all naming `ci.yml` satisfies a count check while declaring nothing about the other 17 files.

### AC-GDL-002 — every entry carries all four fields
**Given** the populated manifest; **When** every entry is checked for `workflow`, `expect_events`, `window`, and `measures`; **Then** BOTH hold — (a) no entry is missing any of the four, and (b) an entry deliberately constructed without `measures` is rejected by the manifest reader with a named error rather than defaulted.
- **RED-now:** no manifest and no reader exist on `d34a789a4` (see AC-GDL-001's measurement). Red for absence.
- **RED-now (b):** the rejection path cannot be exercised at all on the baseline tree — there is nothing to reject with.
- **Green path:** M1 defines the schema and the reader; (a) passes on the populated census, (b) passes when the malformed entry produces the named error.
- **Mutant:** clause (a) alone is satisfied by a reader that silently defaults a missing `measures` to `fired-at-all`. That mutant satisfies the criterion while violating REQ-GDL-002's carrying obligation and, worse, quietly converts the §A.5 all-`skipped` case into a green — which is the defect this SPEC exists to catch. Clause (b) is what excludes it.

### AC-GDL-003 — the measured-quantity vocabulary is closed
**Given** the manifest reader, and a single recorded fixture whose run set holds one `skipped`, one `cancelled`, and one `success` run; **When** an entry declares `measures: fired-at-all`, `fired-with-effect`, or `verdict-rendered`, and separately when one declares any fourth value; **Then** ALL THREE hold — (a) the three named values are accepted, (b) the fourth is rejected with a named error, and (c) **the three values yield three different qualifying sets over that one fixture** — `fired-at-all` admits all three runs, `fired-with-effect` excludes the skipped and cancelled, and `verdict-rendered` admits only the `success`.
- **RED-now:** no vocabulary and no reader exist (AC-GDL-005's measurement). Red for absence, and the criterion is written as a multi-direction set for the reasons below.
- **Green path:** M1 flips (a) and (b); M2 flips (c), since it needs the classifier. Passing output for (c) is three distinct qualifying sets from one fixture.
- **Mutant (a)/(b):** an accept-only criterion passes an all-accepting reader; a reject-only criterion passes a nothing-accepting one. This is the landed rule's own rule-pairing corollary, and it is why both directions sit in one criterion rather than one direction in two.
- **Mutant (c) — a vocabulary value with no consequence:** clauses (a) and (b) together are satisfied by a reader that accepts all three values and then treats them identically — every entry evaluated as `fired-at-all` regardless of what it declared. That mutant passes the whole criteria set while making REQ-GDL-003's one-number-two-events prohibition (`spec.md` §C.1.1) unenforceable: an author could declare `verdict-rendered` and receive `fired-at-all` semantics, which is the conflation the requirement exists to forbid. Only AC-GDL-007 exercised a `measures` value behaviourally before this clause, and it exercised exactly one of the three.
- **Provenance:** this clause exists because the plan-audit's residual-risk section predicted a third instance of the shape behind D1 and D5 — *a state declared in the requirement set with no consequence attached to it* — and warned that closing those two separately without looking would produce it. `fired-at-all` and `verdict-rendered` were that third instance.

### AC-GDL-004 — a legitimately quiet guard declares its condition
**Given** the populated manifest; **When** `spec-status-auto-sync.yml` and `release.yml` are read; **Then** each carries an explicit release-cycle-conditional expectation, and neither is absent from the manifest.
- **RED-now:** both files exist under `.github/workflows/` on `d34a789a4` and neither has any expectation record anywhere in the tree, because no such record exists (AC-GDL-001). Red for absence — and this criterion exists because omission is the tempting shortcut: a release-only guard is the entry an author is most likely to leave out on the reasoning that it is "supposed to be quiet".
- **Green path:** M1. Passing output is both entries present and both carrying the conditional form.

### AC-GDL-005 — per-workflow query only
**Given** the evaluator and a manifest of N entries; **When** one evaluation run is executed against a recorded fixture; **Then** BOTH hold — (a) the evaluator issued exactly N per-workflow queries, one per entry, and (b) it issued zero repository-global run listings.
- **RED-now:** the evaluator does not exist. Measured on `c30f761dd` (and unchanged from `d34a789a4` — no file under `internal/cli/` differs between the two):

  ```
  $ grep -n '"guard"' internal/cli/*.go | grep -v _test.go
  internal/cli/constitution.go:49:		Use:   "guard",
  ```

  The unnarrowed glob returns a second match, `internal/cli/constitution_guard_test.go:93`, which `internal/cli/*.go` does not exclude. Both belong to the unrelated `moai constitution guard` verb; the command is narrowed so the quoted output is what the quoted command produces. Red because there is no query path at all, not because a query was miscounted.
- **Green path:** M2. Passing output is `N` per-workflow queries and `0` global listings.
- **Mutant:** clause (a) alone is satisfied by an evaluator that issues N targeted queries *and also* one global listing it uses as a fast path — which reintroduces §A.4's exact failure for whichever guard the global listing hides. Clause (b) is the negative assertion that excludes it, and it is asserted as a measured call count rather than as a source-text grep, because a grep for a call site is satisfied by a mutant that builds the same request from string fragments.

### AC-GDL-006 — an empty result is UNKNOWN, never "never fired"
**Given** a manifest entry whose per-workflow query returns zero runs; **When** the evaluator classifies it; **Then** the classification is `UNKNOWN`, and the output states that the window is retention-bounded rather than asserting the guard never fired.
- **RED-now:** no classifier exists on `d34a789a4` (AC-GDL-005's measurement). Red for absence.
- **Green path:** M2. Passing output is `UNKNOWN` for that entry, and that entry counted in the `unknown` coverage figure rather than in either the OK or the STALE figure.
- **Why this criterion is not cosmetic:** the two states carry opposite actions. `UNKNOWN` says *go look with a longer window*; "never fired" says *the guard is broken*. Collapsing them manufactures false alarms on release-only guards and would make the advisory unreadable within a week.

### AC-GDL-007 — `fired-with-effect` rejects a skipped run
**Given** an entry declaring `measures: fired-with-effect` and a recorded fixture holding the three real `spec-status-auto-sync.yml` runs of 2026-08-27 (all `conclusion: skipped`), and separately an entry whose fixture holds one `success` run inside its declared window; **When** the evaluator classifies each; **Then** ALL THREE hold — (a) the skipped-run entry does not classify `OK`, (b) its reported last-qualifying-run time is not any of those three timestamps, and (c) **the recent-success entry classifies `OK`**.
- **RED-now:** no evaluator exists (AC-GDL-005). Separately, the fixture's underlying fact is measured and not hypothetical: `gh run list --workflow spec-status-auto-sync.yml --limit 3` on `d34a789a4` returns three `pull_request` runs dated 08-27 09:35 / 08:45 / 08:38, every one `skipped`.
- **Green path:** M2. Passing output classifies the skipped-run entry `STALE` (or `UNKNOWN` if no qualifying run is retained), with the three skipped runs excluded from the qualifying set, and classifies the recent-success entry `OK`.
- **Mutant (a)/(b):** a criterion asserting only "classification is not OK" is satisfied by an evaluator that returns `UNKNOWN` for everything. Clause (b) — that the reported last-qualifying time is not one of the three skipped timestamps — is what forces the exclusion to actually happen rather than the verdict merely to be pessimistic.
- **Mutant (c) — the always-red evaluator:** without an `OK`-path clause, **an evaluator that never classifies anything `OK` passes this entire criteria set.** AC-GDL-006 wants `UNKNOWN`, this criterion's (a) wants "not `OK`", AC-GDL-008 wants `UNDECLARED`, AC-GDL-014 wants `STALE` transitions — and nothing required a healthy guard to read `OK`. Such an evaluator renders an advisory on every run with every entry in it, which is `spec.md` §A.8's always-red mechanism: the second route to the same end state the SPEC names and says it must not inherit. Clause (c) is what closes it, paired with REQ-GDL-009's closed vocabulary.

### AC-GDL-008 — an undeclared workflow file is reported
**Given** a workflow file present under `.github/workflows/` with no manifest entry; **When** the evaluator runs; **Then** that file is classified `UNDECLARED`, appears in the output, and is counted in the `undeclared` coverage figure — and the run does not report an all-clear.
- **RED-now:** no evaluator and no manifest exist on `d34a789a4`, so an undeclared file is the state of all 18 files and nothing reports any of them. Red for absence.
- **Green path:** M2. Passing output names the file and returns a non-all-clear result.
- **Mutant:** a criterion asserting only "appears in the output" is satisfied by an evaluator that lists it under a heading and still returns all-clear. The non-all-clear clause is what makes the classification consequential.
- **Why this is the card's own defect turned on the evaluator:** every instance in `spec.md` §A is a mechanism reporting accurately about a set that had silently become the wrong set (§A.0). An evaluator that reads outcomes only for entries the manifest already lists is that same mechanism — correct about what it examined, silent about what it never examined. This criterion is what forces it to compare the set it examined against the set on disk, and it is the reason a green from this evaluator carries information rather than merely being true.

### AC-GDL-009 — the evaluator declares its own coverage and refuses an empty sweep
**Given** the evaluator and the harness; **When** it emits a result, and separately when it is run against a manifest whose every query fails; **Then** ALL THREE hold — (a) every result carries a measurement timestamp and all four coverage counts (declared, queried, unknown, undeclared), (b) the all-queries-failed run does not report an all-clear and its `queried` count reads `0`, and (c) **that same run renders an advisory to the operator**, naming the degraded coverage, even though no entry is `STALE` and none is `UNDECLARED`.
- **RED-now:** no evaluator and no harness rendering path exist (AC-GDL-005). Red for absence.
- **Green path:** M2 flips (a) and (b); M3 flips (c). Passing output for (c) is a rendered advisory in a run whose every entry classified `UNKNOWN`.
- **Mutant (a)/(b):** this is the vacuous-green shape the landed rule's own evidence footnote names — a sweep that swept nothing still printing ok. Clause (a) alone is satisfied by an evaluator that prints `queried: 0` next to a green verdict; clause (b) is what makes the count decide rather than merely display.
- **Mutant (c) — the one that matters, and it is this SPEC's own subject inside the deliverable:** an evaluator whose every forge query fails classifies every entry `UNKNOWN`, so nothing is `STALE`, nothing is `UNDECLARED`, and a harness triggered only on those two states **renders nothing**. Clauses (a) and (b) both pass, AC-GDL-010(a) and AC-GDL-013 pass vacuously on unsatisfied `Given`s, and the operator sees exactly the silence this card exists to abolish — a mechanism reporting honestly about a set it never successfully examined (`spec.md` §A.0). The gap the clause closes is structural: (a) and (b) bind the **result object**, REQ-GDL-013 binds the **advisory**, and without (c) nothing carries degraded coverage across that boundary.

### AC-GDL-010 — attended invocation, and no scheduled watcher
**Given** the merged change; **When** the invocation surface and `.github/workflows/` are inspected; **Then** BOTH hold — (a) the evaluator is reachable from an already-attended surface and produces an advisory when at least one entry is `STALE` or `UNDECLARED`, and (b) the diff introduces no workflow file carrying a `schedule:` trigger, and the count of `schedule:`-carrying workflow files is unchanged from its baseline value.
- **RED-now (a):** no evaluator and no advisory exist on `d34a789a4`. Red for absence.
- **RED-now (b):** this clause is a *preserved invariant*, not a defect — it is red only against a mutant. Its baseline is measured so the comparison is possible: on `d34a789a4`, `grep -l '^  schedule:' .github/workflows/*` matches `codeql.yml`, `community.yml`, and `release-drafter-cleanup.yml` — 3 files.
- **Green path:** M3 flips (a); (b) holds throughout and is asserted at merge.
- **Mutant:** clause (a) alone is satisfied by implementing the evaluator as a cron workflow that opens an issue — which is functionally an advisory and structurally the regress `spec.md` §D rejects. Clause (b) is the negative assertion that excludes it, and it is stated as a count comparison rather than "no new schedule was added", because the latter is unfalsifiable without a baseline.

### AC-GDL-011 — the advisory carries its own measurement age
**Given** a result **persisted** at a known earlier time T; **When** the advisory is rendered from it at T + Δ with Δ deliberately non-zero; **Then** BOTH hold — (a) the rendered advisory states the age Δ (or the absolute T), and (b) **the stated age is derived from the persisted result's own recorded timestamp**, observed as a rendered Δ > 0 rather than `0s`.
- **RED-now:** no advisory and no persisted result exist (AC-GDL-005). Red for absence.
- **Green path:** M3. Passing output renders a Δ matching the persisted T, not the render moment.
- **Mutant:** clause (a) alone is satisfied by a renderer that computes the displayed age from the moment of rendering. It prints `age: 0s` on every observation, satisfies the criterion textually every time, and produces precisely the failure the criterion exists to prevent — a stale advisory reading as a current all-clear. Clause (b) is what binds the age to the persisted timestamp, and the fixture's deliberately non-zero Δ is what makes the mutant observable rather than merely arguable.
- **Why:** this is the clause that keeps `spec.md` §D honest. The design's answer to self-application is that the evaluator's firing is entailed by attendance — but an advisory rendered from a cached result breaks that entailment silently. Showing the age is what makes the break visible instead of inferable.

### AC-GDL-012 — the doctrine addition is additive only
**Given** the M4 commit touching `.claude/rules/moai/development/verification-completeness.md`; **When** `git diff --numstat d34a789a4 -- <that file>` is read; **Then** BOTH hold — (a) the deleted-line count is `0`, and (b) the added text contains a continued-firing clause, matched by a grep that returns rc=1 on the baseline.
- **RED-now (b):** `grep -nE "last.fired|continued.firing|stopped firing|liveness|stale guard" .claude/rules/moai/development/verification-completeness.md` returns rc=1 with no match on `d34a789a4`. Red because the clause is absent — and this same command is the AC's own discriminator, so its baseline rc=1 is what makes a later rc=0 meaningful.
- **RED-now (a):** trivially satisfied on the baseline (no diff exists yet); it is asserted at M4 as a preserved invariant, not as a starting observation.
- **Green path:** M4. Passing output is `<added> 0 <file>` from `--numstat`, with the grep returning rc=0.
- **Mutant:** the grep clause alone is satisfied by a commit that adds the clause and rewrites the surrounding section. The zero-deletions clause is what enforces "extend, never re-author".

### AC-GDL-013 — the advisory arrives with no operator input
**Given** an evaluation whose result carries at least one `STALE` or `UNDECLARED` entry; **When** the operator begins a session and issues no command naming a guard, a workflow file, or the liveness feature itself; **Then** BOTH hold — (a) the advisory is rendered, and (b) the rendering path consumed no operator-supplied guard identifier or query string.
- **RED-now:** on `d34a789a4` there is no advisory and no evaluator (AC-GDL-005's measurement), so the only way to learn a guard's last-fired time is a hand-written targeted query. Red because the answer is reachable *only* by someone who already knows which question to ask — which is the defect §A.4 records, not merely a missing feature.
- **Green path:** M3. Passing output is the advisory present in a session transcript containing no such command.
- **Mutant:** clause (a) alone is satisfied by shipping a `moai guard liveness` verb plus documentation telling the operator to run it. That mutant renders an advisory, satisfies "the harness surfaces it", and leaves the operator needing to know the mechanism exists — the exact relocation §D.2 rejects. Clause (b) is what excludes it, and it is stated as *inputs consumed* rather than *how the operator felt*, so it is decidable.
- **Why the negative clause is the load-bearing one:** the lead session in §A.4 could run the correct query the moment it was handed the workflow's name. What it could not do was know that a query was owed. A criterion that only checks "an advisory exists when asked for" would have passed against the state that produced this card.

### AC-GDL-014 — the advisory leads with changes and survives a standing red
**Given** two consecutive evaluations where entry set S is `STALE` in both and entry T newly became `STALE` in the second; **When** the second advisory is rendered; **Then** BOTH hold — (a) T appears in the advisory's leading position, and (b) the members of S are represented as a count and are not re-rendered as individual entries.
- **RED-now:** no advisory exists on `d34a789a4`, so there is no rendering order to assert. Red for absence.
- **Green path:** M3. Passing output shows T named and S collapsed to a count.
- **Mutant:** clause (a) alone is satisfied by a renderer that prints the full standing list with T at the top — which is the block a reader learns to skip after the third session, and is how a new advisory inherits the filter that an always-red neighbour has already trained (`spec.md` §A.8). Clause (b) is what forces the collapse.
- **Bounded claim:** this criterion asserts change-leading, not that the advisory gets read. Whether a compact standing count also becomes skippable is not measured by any criterion here and is recorded in §D.7.

### AC-GDL-015 — the manifest holds its watched set as data
**Given** the M1 manifest schema; **When** an entry of a second, non-workflow kind is added — a watched subject carrying a different `kind` and locator, with no GitHub workflow behind it; **Then** BOTH hold — (a) the manifest reader accepts the entry's shape, counts it in `declared`, and includes it in the reported entry list, and (b) accepting it required no change to the manifest schema, the classification vocabulary, or the `measures` vocabulary.
- **RED-now:** no manifest and no schema exist on `d34a789a4` (AC-GDL-001's measurement). Red for absence.
- **Green path:** M1. Passing output is the second-kind entry parsed and reported, with the M1 schema unmodified afterwards.
- **Mutant:** clause (a) alone is satisfied by a reader with a hardcoded `workflow:` field that happens to tolerate an unknown value in it. That mutant accepts the entry while leaving the schema shaped so only a workflow can occupy it — the hardcoded-subject smell §D.4 exists to catch. Clause (b) is what excludes it, and it is stated as *no schema change was required* rather than *the schema looks general*, so it is decidable by diff.
- **Scope note:** this criterion asserts a **shape**, not a capability. It does not require this card to watch anything other than CI guards, and no part of C5 (`spec.md` §E) is in the deliverable. It exists because the test is nearly free at M1 and expensive once 18 entries are written against the schema.
- **Why clause (a) names no classification value:** an earlier draft required the entry to report "as `UNKNOWN` at minimum". That reused `UNKNOWN` for a second, incompatible meaning — REQ-GDL-007 defines it as retention-bounded absence whose implied action is *look again with a longer window*, which is meaningless for an entry whose kind simply has no reader. Two states with different implied actions under one label is the exact defect AC-GDL-006's own rationale argues against, and no requirement mandated the classification anyway. The clause now asserts only what REQ-GDL-001 actually requires — that the entry is accepted, counted, and reported — and names no state.

### AC-GDL-016 — the evaluator mutates nothing
**Given** the evaluator and the harness advisory path; **When** one full evaluation is executed against a recorded fixture whose manifest contains at least one `STALE` and one `UNDECLARED` entry — the classifications most likely to tempt an action; **Then** BOTH hold — (a) the run issues **zero** mutating forge calls (no issue creation, no comment, no dispatch, no workflow re-run, no git write), counted the same way AC-GDL-005 counts queries, and (b) the working tree and the manifest are byte-identical before and after, verified by comparing a checksum of the manifest and `git status --porcelain` across the run.
- **RED-now:** no evaluator exists (AC-GDL-005). Red for absence — there is no call path to count and nothing to run against the fixture.
- **Green path:** M2. Passing output is a zero mutating-call count and an unchanged tree.
- **Mutant:** clause (a) alone is satisfied by an evaluator that writes its result cache into the repository working tree — no forge mutation, but a git write, and one that would then show up as drift for the next reader. Clause (b) catches it. Conversely (b) alone is satisfied by an evaluator that opens a GitHub issue and touches no file, which is precisely the shape AC-GDL-010's own mutant note anticipated ("a cron workflow that opens an issue") and closed only the `schedule:` half of.
- **Why this criterion exists:** REQ-GDL-012 is the SPEC's only safety requirement against the deliverable mutating the forge, and until this criterion it was the one requirement no criterion measured — `plan.md` §D listed it as a constraint, which is a statement of intent, not a verification. The persistence REQ-GDL-015 needs (`plan.md` §M3) must therefore live outside the working tree, or clause (b) fails; that constraint is a deliberate output of this criterion, not an oversight.

## §D.7 Residual risk (recorded, not claimed as closed)

- **The regress is relocated, not eliminated.** The evaluator's liveness is entailed by the liveness of its host surface. If that surface is itself removed or silently stops, the evaluator stops with it — the same defect class, one layer up. Full closure would need an unattended watcher, which reintroduces exactly what `spec.md` §D rejects. Partial mitigation: AC-GDL-011's measurement age makes a stale advisory legible; it does not make an *absent* advisory legible, and nothing in this SPEC does.
- **`gh run list` completeness.** REQ-GDL-007 converts retention loss into an explicit `UNKNOWN`, which prevents a false alarm. It does not recover the lost history: a guard that stopped firing longer ago than the retention window reads `UNKNOWN` forever, indistinguishable from a release-only guard between releases. The expectation window is therefore bounded above by retention, and an entry whose window exceeds it is not answerable by this design.
- **The procedural correlation layer has one executor.** Out of scope per `spec.md` §E, and recorded here because the contrast is the point: correlating scattered observations into a card is done today only by the lead session, which disappears when the lead dies or is cleared. Card t326's deployment check runs without a lead; this card's advisory does too, but the step *after* the advisory — someone deciding a `STALE` classification is worth a card — does not.
- **The host surface can stop invoking the evaluator without being removed.** `spec.md` §D.3 records both directions now, and this is the weaker, likelier one: the surface still exists and still runs, but the condition under which it invokes the evaluator stops matching. It is §A.3 verbatim — `docs-i18n-check.yml` was never removed, its `paths:` filter stopped matching, and the absence read identically to a broken trigger. REQ-GDL-011 forecloses the direct form by requiring invocation to be unconditional on the host's activation, but it binds this deliverable's own wiring only: it cannot stop a later edit from reintroducing a filter, and **no criterion here measures invocation frequency over time**. That is the gap, stated rather than papered over.
- **The advisory reaches one observer, and §A.4's subject is two.** §D.2 answers the discoverability constraint for the operator whose session renders the advisory. §A.4's load-bearing sentence is "there is no signal for 'your view of which guards are alive is incomplete'" — and an advisory rendered in observer 1's session leaves observer 2 exactly where §A.4 found the lead. Nothing in this card closes the multi-observer case; the design narrows *who must already know the question* from "someone" to "whoever attends a session", which is an improvement rather than a closure.
- **Change-leading raises the cost of habituation without abolishing it.** REQ-GDL-015 stops the advisory reprinting an identical block every session, which is what trains a reader's filter. A compact standing count is still something a reader can learn to skip, and no criterion here measures whether the advisory is actually read — AC-GDL-014 measures only that it leads with changes. The always-red neighbour (`spec.md` §A.8) is real and is not repaired by this card, so the new advisory renders next to a channel whose filter is already trained.
- **An entry that stays STALE goes quiet by design.** It is announced once and thereafter only counted. Re-announcing it every session is the noise that produces the filter, so the trade is deliberate — but a long-standing `STALE` is quiet, and quiet is the subject of this card. This is the sharpest unresolved tension in the design and it is recorded rather than argued away.
- **The manifest is hand-maintained.** AC-GDL-001 catches a workflow file added without an entry only when the evaluator next runs, and AC-GDL-008 reports it as `UNDECLARED` rather than blocking. A new guard therefore sits undeclared until someone reads an advisory. This is deliberate — a blocking gate on manifest completeness would be a new always-green risk of its own — but it is a gap, not a solved problem.
