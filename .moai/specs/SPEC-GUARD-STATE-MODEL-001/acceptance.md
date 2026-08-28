# SPEC-GUARD-STATE-MODEL-001 — Acceptance Criteria (card t347)

Baseline tree for every RED-now cell: **`091966c55`** @ `WT-guard-liveness` (worktree `.claude/worktrees/t333`). Every cited command was **run on this tree during authoring** — no cell was carried across the scope reduction from the predecessor SPEC without re-measurement.

Two-cell discipline per `.claude/rules/moai/development/verification-completeness.md` §2. Each criterion carries a RED-now cell stating **why** it is red and a green-path cell naming its flipping milestone, and each was put through that rule's §2 mutant probe.

Budget: Tier M ≤ 16 acceptance criteria. **Count: 16 — at the ceiling.**

**Two baselines, labelled at every use.** RED-now cells measure *this deliverable's absence* and are pinned to `091966c55`. Citations of card t326's landed surfaces are pinned to **`origin/develop` at `ec15ec2cd`**, a diverged tree (diverged, `merge-base --is-ancestor` false). Reading the baseline tree for a t326 surface reports a landed feature as absent, so each such citation names its tree inline.

## §D AC Matrix

| AC | Requirement | Kind | Flipping milestone |
|---|---|---|---|
| AC-GSM-001 | REQ-GSM-001 | Artifact + census pair | M1 |
| AC-GSM-002 | REQ-GSM-002 | Field completeness + rejection pair | M1 |
| AC-GSM-003 | REQ-GSM-003 | Closed vocabulary + behavioural distinctness | M1 (a)(b) / M2 (c) |
| AC-GSM-004 | REQ-GSM-004 | Declared-quiet expectation | M1 |
| AC-GSM-005 | REQ-GSM-001 | Subject-agnostic shape pair | M1 |
| AC-GSM-006 | REQ-GSM-005 | Per-subject query + negative pair | M2 |
| AC-GSM-007 | REQ-GSM-006, REQ-GSM-007 | **Table totality — every row classified + delivery check** | **M1 flips (c) · M2 flips (a)(b)** |
| AC-GSM-008 | REQ-GSM-007 | **Value reachability — every value reachable** | M2 |
| AC-GSM-009 | REQ-GSM-006 | Row 2 — failed query (T2) | M2 |
| AC-GSM-010 | REQ-GSM-006 | Rows 5 and 6 — excused vs unexcused absence | M2 |
| AC-GSM-011 | REQ-GSM-008 | Set comparison against disk | M2 |
| AC-GSM-012 | REQ-GSM-010 | Enumeration-degradation guard (T1) | M2 |
| AC-GSM-013 | REQ-GSM-009 | Counts carried + each consumed | M3 |
| AC-GSM-014 | REQ-GSM-011 | Negative assertion (no mutation) | M3 |
| AC-GSM-015 | REQ-GSM-012 | Published-contract conformance | M3 |
| AC-GSM-016 | REQ-GSM-013 | Surface fold + never-`fail` + reuse boundary | M2 |

## §D.1 Criteria (Given-When-Then, two cells each)

### AC-GSM-001 — the manifest exists and its census is complete
**Given** the repository tree; **When** the manifest is read and its entry set compared against the workflow-file set on disk; **Then** an entry exists for every workflow file, with an empty set-difference in **both** directions.
- **RED-now:** measured on `091966c55` — `ls .moai/guards` returns `No such file or directory` (rc=1). Red because the artifact does not exist, not because a comparison failed. The disk side is measured: `ls -1 .github/workflows/*.yml .github/workflows/*.yaml | wc -l` → `18`.
- **Green path:** M1 populates the census. Passing output is an empty set-difference both ways.
- **Mutant:** a count-equality check is satisfied by 18 entries all naming the same file. The criterion is stated as a **set** comparison in both directions for that reason.

### AC-GSM-002 — every entry carries its four fields
**Given** the populated manifest; **When** each entry is checked for **kind**, locator, expected events, window, and measured quantity; **Then** BOTH hold — (a) no entry is missing any of the **five**, and (b) an entry deliberately constructed without a measured quantity, and separately one without a `kind`, are each **rejected with a named error** rather than defaulted.
- **RED-now:** no manifest and no reader exist on `091966c55` (AC-GSM-001's measurement). The rejection path cannot be exercised at all — there is nothing to reject with.
- **Green path:** M1. (a) passes on the census, (b) passes when the malformed entry produces the named error.
- **Mutant:** clause (a) alone is satisfied by a reader that silently defaults a missing measured quantity to `fired-at-all`. That mutant quietly converts a `skipped`-only subject into a green, which is the defect the vocabulary exists to catch.
- **Why `kind` is checked here and not only in the data-shape requirement:** row 1 of the state table decides on `kind`, so an entry missing it is **undecidable** — and an earlier draft listed four fields here while REQ-GSM-001 mandated `kind` elsewhere, so such an entry passed field-completeness and reached the classifier. A field the table decides on that the completeness check does not see is a gap between two requirements, not a wording preference.

### AC-GSM-003 — the measured-quantity vocabulary is closed and behaviourally distinct
**Given** the manifest reader and one recorded fixture holding a `skipped`, a `cancelled`, and a `success` run; **When** an entry declares each of the three values, and separately any fourth value; **Then** ALL THREE hold — (a) the three named values are accepted, (b) the fourth is rejected with a named error, and (c) **the three yield three different qualifying sets over that one fixture** — `fired-at-all` admits all three runs, `fired-with-effect` excludes the skipped and cancelled, `verdict-rendered` admits only the success.
- **RED-now:** no vocabulary and no reader exist (AC-GSM-001's measurement).
- **Green path:** M1 flips (a) and (b); M2 flips (c), which needs the classifier.
- **Mutant (a)/(b):** an accept-only criterion passes an all-accepting reader; a reject-only criterion passes a nothing-accepting one — the landed rule's own rule-pairing corollary.
- **Mutant (c):** (a) and (b) together are satisfied by a reader that accepts all three and treats them identically. That makes REQ-GSM-003's one-number-two-events prohibition unenforceable: an author could declare `verdict-rendered` and receive `fired-at-all` semantics, which is the conflation the requirement forbids.

### AC-GSM-004 — a legitimately quiet subject declares its condition
**Given** the populated manifest; **When** `spec-status-auto-sync.yml` and `release.yml` are read; **Then** each carries an explicit release-cycle-conditional expectation and neither is absent.
- **RED-now:** both files exist under `.github/workflows/` on `091966c55` and no expectation record exists anywhere in the tree (AC-GSM-001). Red for absence — and this criterion exists because omission is the tempting shortcut for exactly these entries.
- **Green path:** M1. Both entries present, both carrying the conditional form.
- **Consequence:** these entries are what make row 5 of the state table reachable, so omitting them would also make AC-GSM-008 unsatisfiable.

### AC-GSM-005 — the manifest holds its watched set as data
**Given** the M1 schema; **When** an entry of a second, non-workflow kind is added, carrying a different kind and locator with no forge workflow behind it; **Then** BOTH hold — (a) the reader accepts it, counts it in `declared`, and classifies it `UNREADABLE`, and (b) accepting it required **no change** to the manifest schema, the classification vocabulary, or the measured-quantity vocabulary.
- **RED-now:** no manifest and no schema exist (AC-GSM-001's measurement).
- **Green path:** M1 for the schema shape (b); M2 for the classification (a), which needs the classifier — the criterion is clause-split across milestones for that reason.
- **Mutant:** clause (a) alone is satisfied by a reader with a hardcoded workflow-path field that tolerates an unknown value in it. Clause (b) excludes it, stated as *no schema change was required* rather than *the schema looks general*, so it is decidable by diff.
- **Scope note:** this asserts a **shape**, not a capability. No second kind ships here.

### AC-GSM-006 — per-subject query only
**Given** the evaluator and a manifest of N entries; **When** one evaluation runs against a recorded fixture; **Then** BOTH hold — (a) exactly N per-subject queries are issued, one per entry, and (b) **zero** repository-global run listings are issued.
- **RED-now:** the evaluator does not exist. Measured on `091966c55`:

  ```
  $ grep -n '"guard"' internal/cli/*.go | grep -v _test.go
  internal/cli/constitution.go:49:		Use:   "guard",
  ```

  The single match is the unrelated `moai constitution guard` verb. (The unnarrowed glob also returns `constitution_guard_test.go:93`, same unrelated verb — the command is narrowed so the quoted output is what the quoted command produces.) Red because there is no query path at all.
- **Green path:** M2. Passing output is `N` per-subject queries and `0` global listings.
- **Mutant:** clause (a) alone is satisfied by an evaluator that issues N targeted queries **and also** one global listing used as a fast path — which reintroduces the exact failure for whichever subject the global listing hides. Clause (b) is stated as a measured call count rather than a source grep, because a grep for a call site is satisfied by a mutant building the same request from string fragments.

### AC-GSM-007 — the state table is total: every row classifies
Stated as two independently-evaluable halves, because they are satisfiable at different milestones and a compound `When` cannot be met at the earlier one.

**Clauses (a)(b) — requires the classifier. Given** a fixture supplying **one case per row** of the REQ-GSM-006 table (8 rows); **When** the evaluator classifies each case; **Then** BOTH hold — (a) every case receives **exactly one** classification, none unclassified and none receiving more than one, and (b) the classifier is table-driven, so each case's value traces to its row.

**Clause (c) — a schema read, requiring no evaluator. Given** the M1 schema and the table's `Flipped by` column; **When** the column is read against the schema; **Then** (c) **for every row** — not only rows that happen to declare a dependency — the M1 fields that row's decision depends on are present in the M1 schema.
- **RED-now (a)(b):** no classifier exists (AC-GSM-006's measurement). Red for absence — no case can be decided.
- **RED-now (c):** no manifest schema exists (AC-GSM-001's measurement), so no row's dependency can be present in it.
- **Green path:** **M1 flips (c)** — it is a schema read and needs no evaluator; **M2 flips (a)(b)**. Passing output is 8 cases with 8 single classifications, and every row's dependency present in the M1 schema.
- **Why the criterion is split:** an earlier draft carried one compound `Given/When` — "the evaluator classifies each case **and** the column is read against the schema" — while assigning clause (c) to M1, where no evaluator exists. The compound could not be satisfied at the milestone it was assigned to, and the green-path cell independently named M2 alone. **That is the predecessor's own failure shape appearing inside the criterion built to prevent it**: a clause scheduled where it cannot be evaluated. The split follows AC-GSM-003 and AC-GSM-005, which already carry per-clause milestones.
- **Why this criterion is the centre of the SPEC:** the predecessor's state model failed three audits by leaving a condition uncovered while every requirement read complete in isolation. A per-row fixture makes an uncovered condition a **failing case** rather than a sentence nobody wrote. This is the "state table, not prose" instruction made mechanical.
- **Mutant:** a classifier with a catch-all default satisfies "every case receives exactly one classification" while collapsing rows 1, 2 and 6 into one value. AC-GSM-008 is the paired direction that excludes it, and AC-GSM-009 and AC-GSM-010 pin the specific rows that have historically collapsed.
- **Delivery clause (c) is stated over ALL rows, not over rows that declare a dependency.** An earlier draft scoped it to "rows whose cell declares an M1 dependency (rows 1, 3, 4, 5)", which made the check **declaration-driven**: a row that under-declares its dependency is invisible to the very check the column exists to enable. Row 6 was exactly that case — the negative branch of row 5's conditional test, equally dependent on the same M1 field, declaring none. Scoping the clause over all rows removes the escape hatch; the row-6 cell is also now corrected.
- **Why clause (c) is a separate assertion from (a):** the predecessor's milestone map passed a **union count** — every criterion appeared in some flip list — while one criterion sat at a milestone that could not deliver it. A union count answers "is every criterion listed?" and is structurally blind to "can the listed milestone deliver it?". Clause (c) is the second question, and it is checked against the schema because that is the only place the answer lives.
- **Mutant (c):** clauses (a) and (b) together are satisfied by a table whose `Flipped by` column reads `M2` uniformly. That column would be true and useless — every classification is emitted by the classifier — and it would hide exactly the failure it exists to catch: a row whose M1 field was never shipped produces the **wrong value**, not a missing one. Row 5 is the sharp case: without its conditional field it silently becomes row 6, and a correctly-quiet release-only subject is reported as an anomaly on every sweep.

### AC-GSM-008 — every classification value is reachable
**Given** the same 8-row fixture; **When** the classifications are collected; **Then** **each of the seven values** — `OK`, `STALE`, `UNKNOWN`, `UNDECLARED`, `UNREADABLE`, `UNRESOLVED`, `ORPHANED` — is produced by at least one case.
- **RED-now:** no classifier exists (AC-GSM-006's measurement).
- **Green path:** M2. Passing output is all seven values observed across the 8 cases (rows 3 and 5 both yield `OK`).
- **Mutant this closes:** without it, **an evaluator that never emits `OK` passes every other criterion in this set** — AC-GSM-009 wants `UNRESOLVED`, AC-GSM-010 wants `UNKNOWN`, AC-GSM-011 wants `UNDECLARED`, and nothing requires a healthy subject to read clean. Such an evaluator marks every entry non-clean, so the consuming SPEC's advisory fires on every sweep with every entry in it — the always-red mechanism, arriving through the classifier.
- **Pairing:** AC-GSM-007 and AC-GSM-008 are the two directions of REQ-GSM-007's totality claim and are adopted as a pair. Either alone admits a mutant on the other direction.

### AC-GSM-009 — a failed query is `UNRESOLVED`, not silence (T2)
**Given** a manifest entry whose kind has a reader and whose query **returns an error** (simulated auth failure, then rate limit); **When** the evaluator classifies it; **Then** the classification is `UNRESOLVED`, it is **not** `UNKNOWN`, and it is counted in a per-value count of its own.
- **RED-now:** no classifier exists (AC-GSM-006's measurement). The condition itself is the audit's T2: under the predecessor's requirements this entry had **no admissible classification at all** — one requirement routed it into the no-reader value, another defined that value as no-reader-only, and the acceptance criteria assumed a branch the requirements forbade.
- **Green path:** M2. Passing output is `UNRESOLVED` for both error shapes.
- **Why the negative clause:** `UNKNOWN` means retention-bounded absence and its implied action is *look again with a longer window*, which is wrong advice for an auth failure. Two states with different implied actions under one label is the defect that produced this SPEC.
- **Note:** this is the single most likely degraded run in production, which is why it gets its own criterion rather than riding on AC-GSM-007's row coverage.

### AC-GSM-010 — excused absence is `OK`; unexcused absence is `UNKNOWN`
**Given** two entries whose queries each return zero qualifying runs — one whose declared condition says firing is not currently expected, one whose condition does not account for the absence; **When** the evaluator classifies each; **Then** the first is `OK` and the second is `UNKNOWN`.
- **RED-now:** no classifier exists (AC-GSM-006's measurement).
- **Green path:** M2. Passing output is the two distinct classifications from the same observable input.
- **Why both directions in one criterion:** the input is identical — zero runs — and only the declared expectation separates them. A criterion asserting either direction alone is satisfied by a classifier that ignores the declared condition entirely and always returns its one preferred value. This is also what keeps the consuming SPEC's advisory from firing every session on a healthy repository, so a collapse here is not local: it degrades the sibling SPEC's trigger into an always-red channel.

### AC-GSM-011 — an undeclared workflow file is reported
**Given** two fixtures exercising the set comparison in **both** directions — (i) a workflow file present on disk with no manifest entry, and (ii) a manifest entry whose named workflow file is absent from disk; **When** the evaluator runs on each; **Then** ALL hold — (i) classifies `UNDECLARED` and (ii) classifies **`ORPHANED`**, each appears in the output, each is counted, and **neither run reports an all-clear**.
- **RED-now:** no evaluator and no manifest exist (AC-GSM-006's measurement), so every one of the 18 files is undeclared and nothing reports any of them.
- **Green path:** M2. Passing output names the file and returns a non-all-clear result.
- **Mutant:** a criterion asserting only "appears in the output" is satisfied by an evaluator that lists it under a heading and still returns all-clear. The non-all-clear clause makes the classification consequential.
- **Why it is load-bearing:** this is the set comparison. An evaluator that reads outcomes only for entries the manifest already lists is the card's own defect at the evaluator's layer — correct about what it examined, silent about what it never examined.
- **Why fixture (ii) exists, and it is the totality hole this SPEC was built to make impossible:** an earlier draft covered the disk→entry direction only. The mirror — a declared entry whose subject was **deleted** — reached the classifier and was decided by run history, which outlives the deletion: it classified `STALE` (advice to investigate a subject that no longer exists) or, on a recent enough run, `OK` — **a false clean**. It did not present as an empty cell, which is why a totality claim over the table's own rows could not see it. Fixture (ii) is the runtime check that the second direction is actually implemented, not merely asserted at census time by AC-GSM-001.

### AC-GSM-012 — a degraded enumeration cannot report all-clear (T1)
**Given** two degraded enumerations with every manifest entry otherwise classifying `OK` — (i) one returning **zero files** (simulated wrong working directory, then a permissions failure), and (ii) one returning a **non-zero but partial** count (3 of 18, simulated wrong glob); **When** each run completes; **Then** ALL THREE hold — (a) **neither** run reports an all-clear, (b) the enumerated-files count is visible in each output, and (c) run (ii) is refused on the **declared-vs-enumerated** comparison, not on a zero test.
- **RED-now:** no evaluator and no enumeration exist (AC-GSM-006's measurement).
- **Green path:** M2. Passing output is a non-all-clear result with `enumerated: 0` visible.
- **Mutant this closes, and it is the card's own defect one layer in:** the set comparison has two inputs and the predecessor guarded only one. An evaluator whose enumeration silently returns empty classifies every manifest entry `OK`, reports `UNDECLARED: 0`, has a non-zero queried count so the zero-query guard does not bite, and **reports all-clear honestly about a set it never learned**. `UNDECLARED: 0` is indistinguishable between "census complete" and "enumeration found nothing" — clause (b) is what separates them.
- **Why AC-GSM-011 does not catch it:** that criterion's fixtures supply an undeclared file and an orphaned entry, so enumeration works there by construction.
- **Why fixture (ii) exists — a zero-check is the card's own defect at a lower dose:** an enumeration returning 3 of 18 files is non-zero, so a zero-guard passes it unremarked. It yields no `UNDECLARED` finding for the 15 files it never saw and reports all-clear about a set that had silently become the wrong set. The guard therefore has to test that the **right** thing happened, not merely that *something* did — the same proxy failure as the parent card's `queried < declared` trigger. It is nearly free, because REQ-GSM-009 already carries both counts.

### AC-GSM-013 — every carried count has a consumer
**Given** the evaluator's result; **When** it is emitted; **Then** BOTH hold — (a) it carries the measurement timestamp, the enumerated-files count, the declared count, the queried count, and a per-value count for each of the seven classifications, and (b) **each carried count is consumed by a named consumer that reads it** — a decision it changes, or a guard condition it feeds — with no count that is rendered and read by nothing.
- **RED-now:** no evaluator exists (AC-GSM-006's measurement).
- **Green path:** M3. Passing output is the full count set with each count traced to a consumer.
- **Why clause (b):** the parent card's audit found no inert requirement but **two inert fields** — counts required, rendered, and consumed by nothing. Field-level inertness inside a live requirement is the residue this clause prevents, and the enumerated-files count is a count with a consumer by construction (REQ-GSM-010's guard reads it, and now compares it against the declared count).
- **The disjunct that had to go, because it made the clause vacuous:** an earlier draft accepted a count as consumed if it was "named in a decision, a guard condition, **or the published result's own contract**". Every carried count is by definition part of the published result, so the third disjunct is satisfied by any emitted field — and it contradicted REQ-GSM-009, which says a count *rendered* and consumed by nothing is exactly the inertness to avoid. Mutant it admitted: emit all seven per-value counts plus the timestamp, declare each "consumed by the published result's own contract", pass the criterion with every count inert. The clause now requires a consumer that **reads** the count.

### AC-GSM-014 — the evaluator mutates nothing
**Given** the evaluator and a fixture whose manifest contains at least one `STALE` and one `UNDECLARED` entry — the classifications most likely to tempt an action; **When** one full evaluation runs; **Then** BOTH hold — (a) **zero** mutating forge calls (no issue creation, no comment, no dispatch, no re-run), counted the same way AC-GSM-006 counts queries, and (b) the working tree is byte-identical before and after, verified by **`git status --porcelain --ignored`** across the run.
- **RED-now:** no evaluator exists (AC-GSM-006's measurement) — there is no call path to count.
- **Green path:** M3. Passing output is a zero mutating-call count and an unchanged tree.
- **Mutant:** clause (a) alone is satisfied by an evaluator writing its cache into the working tree — no forge mutation, but a git write that becomes drift for the next reader. Clause (b) alone is satisfied by an evaluator that opens an issue and touches no file.
- **Why `--ignored`, and this is the criterion's own mutant escaping its instrument:** plain `git status --porcelain` **does not report ignored paths**. An evaluator persisting its cache to a gitignored path inside the tree — `.moai/state/…` is the natural choice, and `plan.md` §D contemplates exactly this location — produces no porcelain output, passes both clauses, and violates REQ-GSM-011's "shall not write to the repository working tree". The criterion named that mutant and used an instrument blind to it. `--ignored` is what makes the stated mutant actually fail.

### AC-GSM-015 — the published contract holds
**Given** any emitted result; **When** its classifications and its designation are inspected; **Then** ALL THREE hold — (a) every entry carries exactly one classification, (b) **exactly one value in the vocabulary is designated clean**, verified by asserting the designation is single-valued rather than by counting how many entries happen to be `OK`, and (c) **the result carries that designation in machine-readable form**, such that a consumer given only the result — and no knowledge of any value's name — can partition entries into clean and non-clean. Clause (c) is mechanized the way its mirror is: the partitioning test harness contains **no classification value token**, verified by a value-token grep over that harness returning rc=1, so the harness's blindness is measured rather than asserted.
- **RED-now:** no evaluator and no published designation exist (AC-GSM-006's measurement).
- **Green path:** M3. Passing output is a single designated clean value.
- **Why this criterion exists here and not in the consuming SPEC:** `SPEC-GUARD-LIVENESS-001` consumes the contract and can verify that its own trigger uses the clean/non-clean partition — but it **cannot** verify that exactly one clean value exists, because it never names the vocabulary. If this SPEC ever designated two clean values, every criterion in the consuming SPEC would still pass while its advisory under-fired. This criterion is the other half of that seam.
- **Mutant (b):** clause (b) stated as "at least one entry classifies `OK`" is satisfied while a second value is also treated as clean. Asserting the **designation** rather than the observed entries is what closes it.
- **Mutant (c), and it is the seam's own failure mode:** clauses (a) and (b) together are satisfied by a result that designates the clean value **in prose in this SPEC** and emits nothing. The consumer is then forced to hardcode the literal — which its own criteria forbid — or to read the surface fold, which under-reports because `UNREADABLE` also folds to `ok` while only `OK` is clean. Clause (c) is what makes the designation a *carried field* rather than a documented fact, and the test is stated as *a consumer given only the result and no value names can partition it*, so it is decidable without reference to either SPEC's prose.
- **Symmetry note:** this clause is the producing half of a two-sided obligation. Its consuming half is `SPEC-GUARD-LIVENESS-001` REQ-GDL-001 (iii)/(iv) and AC-GDL-001(c). Repairing one side alone re-opens the gap in the other direction, which is why both landed in the same commit.

### AC-GSM-016 — every classification folds, and none folds to `fail`
**Given** the 8-row fixture of AC-GSM-007 plus the table's `Surface fold` column; **When** each case is classified and folded; **Then** ALL THREE hold — (a) every one of the seven classifications has a declared fold into `ok` / `warn` / `fail`, (b) **no classification folds to `fail`**, and (c) the meaningless/incomplete boundary holds — `UNREADABLE` folds to `ok` while `UNKNOWN`, `UNRESOLVED` and `ORPHANED` fold to `warn`.
- **RED-now:** no classifier and no fold exist on this SPEC's baseline (AC-GSM-006's measurement). Red for absence. The surface vocabulary the fold targets is measured on **`origin/develop` at `ec15ec2cd`** — `CheckOK`/`CheckWarn`/`CheckFail` in `internal/cli/uikit/types.go`, three values with no skipped state — a diverged tree from this SPEC's baseline (`spec.md` §C.2.1).
- **Green path:** M2. Passing output is six declared folds, zero `fail`, and the boundary as stated.
- **Mutant (b):** without the never-`fail` clause, an implementation folding `UNDECLARED` or `STALE` to `fail` satisfies every other criterion while promoting a routine sweep to a failing exit status. Card t326 reached the same conclusion for the same reason and recorded it in source: *"never Fail — a Fail here would promote every downstream `moai doctor` run to exit 1"*.
- **Mutant (c), and it is this SPEC's own subject:** an implementation that adopts t326's leniency wholesale folds **all** uncertainty to `ok` — `UNREADABLE`, `UNKNOWN` and `UNRESOLVED` alike. It passes (a) and (b), reads as principled reuse, and reproduces the defect this card exists to catch: a mechanism reporting green about a set it never learned. Clause (c) is where the reuse stops, and it stops at the line between *meaningless* (no reader — the comparison never applied) and *incomplete* (the comparison applied and could not finish).

## §D.7 Residual risk (recorded, not claimed as closed)

- **The state table is total over the two inputs the derivation names, which is not the same as total over reality.** AC-GSM-007 proves every row classifies and AC-GSM-008 proves every value is reachable; neither proves the eight rows exhaust what the mechanism can encounter. **This bound is not theoretical — it has already been hit once.** The v0.2.0 table asserted totality over seven rows and a by-construction re-derivation broke it immediately at the deleted-subject condition, which produced a wrong value rather than an empty cell. The improvement is that the derivation method is now stated beside the table (REQ-GSM-006), so a reader can re-run it instead of trusting the result; the residual is that a *third input* would open conditions this derivation cannot reach, and nothing here would surface that.
- **`UNREADABLE`'s distinctness rests on one axis only.** Its implied action ("None until a reader exists") is operationally identical to `OK`'s ("None"), and both fold to `ok`. What separates them is **cleanliness**: `OK` is the clean value and `UNREADABLE` is not, carried to consumers by the designator (REQ-GSM-012). REQ-GSM-007's distinctness clause is worded over *observable axes* for this reason rather than over implied actions, which the pair would have satisfied only in wording. The value is not inert — it is the sole classification that is non-clean while folding to the clean surface value, which is exactly the gap entry the sibling's own seam criterion requires as a fixture — but its work is carried by one axis, and if the designator were ever dropped the pair would become genuinely indistinguishable.
- **A second-kind entry would be permanently `UNREADABLE`.** No second-kind reader ships (§D Out of Scope), so such an entry is a standing non-clean entry whose implied action is "None" — the permanently-non-clean neighbour the sibling SPEC warns trains a reader's filter. Nothing here prevents one being declared.
- **`gh run list` retention bounds the design from above.** REQ-GSM-006 row 6 converts retention loss into an explicit `UNKNOWN`, preventing a false alarm. It does not recover the lost history: a subject that stopped firing longer ago than the retained window reads `UNKNOWN` indefinitely. An expectation window exceeding retention is therefore not answerable by this design, and the manifest does not currently reject one.
- **Row 5 depends on the declared condition being written correctly.** A release-only entry whose condition is mis-authored classifies `OK` while genuinely stale — a false clean, which is the worst direction. AC-GSM-004 checks the condition is *present*, not that it is *right*, and nothing here can check the latter.
- **The manifest is hand-maintained.** AC-GSM-011 catches a workflow file added without an entry only when the evaluator next runs, and reports it as `UNDECLARED` rather than blocking. A new guard therefore sits undeclared until someone reads the result. Deliberate — a blocking completeness gate would be a new always-green risk — but it is a gap.
- **`UNREADABLE` is asserted by one criterion and exercised by no shipping code path.** No second-kind reader ships (§D Out of Scope), so the value's only exercise is AC-GSM-005's shape fixture. It is reachable per AC-GSM-008, but its implied action has never been acted on.
- **The evaluator's own liveness is out of this SPEC's hands.** Whether it runs at all is `SPEC-GUARD-LIVENESS-001`'s REQ-GDL-003. A correct state model invoked by nothing produces no classifications, and nothing here would detect that — the split places this SPEC's own liveness in its sibling, which is the honest consequence of the seam.
