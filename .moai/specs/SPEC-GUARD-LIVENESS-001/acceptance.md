# SPEC-GUARD-LIVENESS-001 — Acceptance Criteria (card t333)

Baseline tree for every RED-now cell: **`d34a789a4`** @ `WT-guard-liveness` (worktree `.claude/worktrees/t333`), equal to `origin/develop` at authoring time. Every RED-now cell below was measured on that tree and states **why** it is red, per `.claude/rules/moai/development/verification-completeness.md` §2 — RED-now alone does not distinguish a criterion this work can flip from one that is red forever.

Each criterion was put through that rule's §2 mutant probe. Where a single-clause criterion admitted a mutant satisfying it while violating its requirement, a second clause was added and the mutant that forced it is recorded in the criterion.

Budget: Tier M ≤ 16 acceptance criteria. **Count: 15.**

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
| AC-GDL-012 | REQ-GDL-015 | Additive-only diff assertion | M4 |
| AC-GDL-013 | REQ-GDL-013 | Zero-input surfacing pair | M3 |
| AC-GDL-014 | REQ-GDL-016 | Change-leading pair | M3 |
| AC-GDL-015 | REQ-GDL-001 | Subject-agnostic shape pair | M1 |

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
**Given** the manifest reader; **When** an entry declares `measures: fired-at-all`, `fired-with-effect`, or `verdict-rendered`, and separately when one declares any fourth value; **Then** the three named values are accepted and the fourth is rejected with a named error.
- **RED-now:** no vocabulary is defined anywhere on `d34a789a4` — the reader that would accept or reject does not exist. Red for absence, and the criterion is written as a two-direction pair for the reason below.
- **Green path:** M1 fixes the vocabulary; passing output accepts exactly three and rejects the fourth.
- **Mutant:** an accept-only criterion passes an all-accepting reader; a reject-only criterion passes a nothing-accepting one. This is the rule's own rule-pairing corollary, and it is why both directions are asserted in one criterion rather than one direction in two.

### AC-GDL-004 — a legitimately quiet guard declares its condition
**Given** the populated manifest; **When** `spec-status-auto-sync.yml` and `release.yml` are read; **Then** each carries an explicit release-cycle-conditional expectation, and neither is absent from the manifest.
- **RED-now:** both files exist under `.github/workflows/` on `d34a789a4` and neither has any expectation record anywhere in the tree, because no such record exists (AC-GDL-001). Red for absence — and this criterion exists because omission is the tempting shortcut: a release-only guard is the entry an author is most likely to leave out on the reasoning that it is "supposed to be quiet".
- **Green path:** M1. Passing output is both entries present and both carrying the conditional form.

### AC-GDL-005 — per-workflow query only
**Given** the evaluator and a manifest of N entries; **When** one evaluation run is executed against a recorded fixture; **Then** BOTH hold — (a) the evaluator issued exactly N per-workflow queries, one per entry, and (b) it issued zero repository-global run listings.
- **RED-now:** the evaluator does not exist on `d34a789a4` — `grep -n '"guard"' internal/cli/*.go` returns only `internal/cli/constitution.go:49`, an unrelated `moai constitution guard` verb. Red because there is no query path at all, not because a query was miscounted.
- **Green path:** M2. Passing output is `N` per-workflow queries and `0` global listings.
- **Mutant:** clause (a) alone is satisfied by an evaluator that issues N targeted queries *and also* one global listing it uses as a fast path — which reintroduces §A.4's exact failure for whichever guard the global listing hides. Clause (b) is the negative assertion that excludes it, and it is asserted as a measured call count rather than as a source-text grep, because a grep for a call site is satisfied by a mutant that builds the same request from string fragments.

### AC-GDL-006 — an empty result is UNKNOWN, never "never fired"
**Given** a manifest entry whose per-workflow query returns zero runs; **When** the evaluator classifies it; **Then** the classification is `UNKNOWN`, and the output states that the window is retention-bounded rather than asserting the guard never fired.
- **RED-now:** no classifier exists on `d34a789a4` (AC-GDL-005's measurement). Red for absence.
- **Green path:** M2. Passing output is `UNKNOWN` for that entry, and that entry counted in the `unknown` coverage figure rather than in either the OK or the STALE figure.
- **Why this criterion is not cosmetic:** the two states carry opposite actions. `UNKNOWN` says *go look with a longer window*; "never fired" says *the guard is broken*. Collapsing them manufactures false alarms on release-only guards and would make the advisory unreadable within a week.

### AC-GDL-007 — `fired-with-effect` rejects a skipped run
**Given** an entry declaring `measures: fired-with-effect` and a recorded fixture holding the three real `spec-status-auto-sync.yml` runs of 2026-08-27 (all `conclusion: skipped`); **When** the evaluator classifies it; **Then** the classification is not `OK`, and the reported last-qualifying-run time is not any of those three timestamps.
- **RED-now:** no evaluator exists (AC-GDL-005). Separately, the fixture's underlying fact is measured and not hypothetical: `gh run list --workflow spec-status-auto-sync.yml --limit 3` on `d34a789a4` returns three `pull_request` runs dated 08-27 09:35 / 08:45 / 08:38, every one `skipped`.
- **Green path:** M2. Passing output classifies the entry `STALE` (or `UNKNOWN` if no qualifying run is retained), with the three skipped runs excluded from the qualifying set.
- **Mutant:** a criterion asserting only "classification is not OK" is satisfied by an evaluator that returns `UNKNOWN` for everything. The second clause — that the reported last-qualifying time is not one of the three skipped timestamps — is what forces the exclusion to actually happen rather than the verdict merely to be pessimistic.

### AC-GDL-008 — an undeclared workflow file is reported
**Given** a workflow file present under `.github/workflows/` with no manifest entry; **When** the evaluator runs; **Then** that file is classified `UNDECLARED`, appears in the output, and is counted in the `undeclared` coverage figure — and the run does not report an all-clear.
- **RED-now:** no evaluator and no manifest exist on `d34a789a4`, so an undeclared file is the state of all 18 files and nothing reports any of them. Red for absence.
- **Green path:** M2. Passing output names the file and returns a non-all-clear result.
- **Mutant:** a criterion asserting only "appears in the output" is satisfied by an evaluator that lists it under a heading and still returns all-clear. The non-all-clear clause is what makes the classification consequential.

### AC-GDL-009 — the evaluator declares its own coverage and refuses an empty sweep
**Given** the evaluator; **When** it emits a result, and separately when it is run against a manifest whose every query fails; **Then** BOTH hold — (a) every result carries a measurement timestamp and all four coverage counts (declared, queried, unknown, undeclared), and (b) the all-queries-failed run does not report an all-clear and its `queried` count reads `0`.
- **RED-now:** no evaluator exists on `d34a789a4` (AC-GDL-005). Red for absence.
- **Green path:** M2. Passing output for (b) is a non-all-clear result whose `queried: 0` is visible in the same output.
- **Mutant:** this is the vacuous-green shape the landed rule's own evidence footnote names — a sweep that swept nothing still printing ok. Clause (a) alone is satisfied by an evaluator that prints `queried: 0` next to a green verdict; clause (b) is what makes the count decide rather than merely display.

### AC-GDL-010 — attended invocation, and no scheduled watcher
**Given** the merged change; **When** the invocation surface and `.github/workflows/` are inspected; **Then** BOTH hold — (a) the evaluator is reachable from an already-attended surface and produces an advisory when at least one entry is `STALE` or `UNDECLARED`, and (b) the diff introduces no workflow file carrying a `schedule:` trigger, and the count of `schedule:`-carrying workflow files is unchanged from its baseline value.
- **RED-now (a):** no evaluator and no advisory exist on `d34a789a4`. Red for absence.
- **RED-now (b):** this clause is a *preserved invariant*, not a defect — it is red only against a mutant. Its baseline is measured so the comparison is possible: on `d34a789a4`, `grep -l '^  schedule:' .github/workflows/*` matches `codeql.yml`, `community.yml`, and `release-drafter-cleanup.yml` — 3 files.
- **Green path:** M3 flips (a); (b) holds throughout and is asserted at merge.
- **Mutant:** clause (a) alone is satisfied by implementing the evaluator as a cron workflow that opens an issue — which is functionally an advisory and structurally the regress `spec.md` §D rejects. Clause (b) is the negative assertion that excludes it, and it is stated as a count comparison rather than "no new schedule was added", because the latter is unfalsifiable without a baseline.

### AC-GDL-011 — the advisory carries its own measurement age
**Given** an advisory rendered from a result measured at time T; **When** it is displayed at time T + Δ; **Then** the rendered advisory states the age Δ (or the absolute T), so a reader can tell a current all-clear from a stale one.
- **RED-now:** no advisory exists on `d34a789a4`. Red for absence.
- **Green path:** M3. Passing output shows the age alongside the verdict.
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
**Given** the M1 manifest schema; **When** an entry of a second, non-workflow kind is added — a watched subject carrying a different `kind` and locator, with no GitHub workflow behind it; **Then** BOTH hold — (a) the manifest reader accepts the entry's shape and reports it (as `UNKNOWN` at minimum, since no reader for that kind exists yet), and (b) accepting it required no change to the manifest schema, the classification vocabulary, or the `measures` vocabulary.
- **RED-now:** no manifest and no schema exist on `d34a789a4` (AC-GDL-001's measurement). Red for absence.
- **Green path:** M1. Passing output is the second-kind entry parsed and reported, with the M1 schema unmodified afterwards.
- **Mutant:** clause (a) alone is satisfied by a reader with a hardcoded `workflow:` field that happens to tolerate an unknown value in it. That mutant accepts the entry while leaving the schema shaped so only a workflow can occupy it — the hardcoded-subject smell §D.4 exists to catch. Clause (b) is what excludes it, and it is stated as *no schema change was required* rather than *the schema looks general*, so it is decidable by diff.
- **Scope note:** this criterion asserts a **shape**, not a capability. It does not require this card to watch anything other than CI guards, and no part of C5 (`spec.md` §E) is in the deliverable. It exists because the test is nearly free at M1 and expensive once 18 entries are written against the schema.

## §D.7 Residual risk (recorded, not claimed as closed)

- **The regress is relocated, not eliminated.** The evaluator's liveness is entailed by the liveness of its host surface. If that surface is itself removed or silently stops, the evaluator stops with it — the same defect class, one layer up. Full closure would need an unattended watcher, which reintroduces exactly what `spec.md` §D rejects. Partial mitigation: AC-GDL-011's measurement age makes a stale advisory legible; it does not make an *absent* advisory legible, and nothing in this SPEC does.
- **`gh run list` completeness.** REQ-GDL-007 converts retention loss into an explicit `UNKNOWN`, which prevents a false alarm. It does not recover the lost history: a guard that stopped firing longer ago than the retention window reads `UNKNOWN` forever, indistinguishable from a release-only guard between releases. The expectation window is therefore bounded above by retention, and an entry whose window exceeds it is not answerable by this design.
- **The procedural correlation layer has one executor.** Out of scope per `spec.md` §E, and recorded here because the contrast is the point: correlating scattered observations into a card is done today only by the lead session, which disappears when the lead dies or is cleared. Card t326's deployment check runs without a lead; this card's advisory does too, but the step *after* the advisory — someone deciding a `STALE` classification is worth a card — does not.
- **Change-leading raises the cost of habituation without abolishing it.** REQ-GDL-016 stops the advisory reprinting an identical block every session, which is what trains a reader's filter. A compact standing count is still something a reader can learn to skip, and no criterion here measures whether the advisory is actually read — AC-GDL-014 measures only that it leads with changes. The always-red neighbour (`spec.md` §A.8) is real and is not repaired by this card, so the new advisory renders next to a channel whose filter is already trained.
- **An entry that stays STALE goes quiet by design.** It is announced once and thereafter only counted. Re-announcing it every session is the noise that produces the filter, so the trade is deliberate — but a long-standing `STALE` is quiet, and quiet is the subject of this card. This is the sharpest unresolved tension in the design and it is recorded rather than argued away.
- **The manifest is hand-maintained.** AC-GDL-001 catches a workflow file added without an entry only when the evaluator next runs, and AC-GDL-008 reports it as `UNDECLARED` rather than blocking. A new guard therefore sits undeclared until someone reads an advisory. This is deliberate — a blocking gate on manifest completeness would be a new always-green risk of its own — but it is a gap, not a solved problem.
