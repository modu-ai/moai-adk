# SPEC-GUARD-LIVENESS-001 — Acceptance Criteria (card t333, surfacing model)

Baseline tree for every RED-now cell: **`091966c55`** @ `WT-guard-liveness` (worktree `.claude/worktrees/t333`). Every cited command was re-run on this tree during the scope reduction; **no cell was carried across the split without re-running its command**, per the audit's D7 finding that an unreproducible measurement inside a RED-now cell is the evidence-integrity failure the two-cell discipline exists to prevent.

Two-cell discipline per `.claude/rules/moai/development/verification-completeness.md` §2: each criterion carries a RED-now cell stating **why** it is red, and a green-path cell naming the flipping milestone. Each was put through that rule's §2 mutant probe; where a single-clause criterion admitted a mutant, a second clause was added and the mutant is recorded.

Budget: Tier M ≤ 16 acceptance criteria. **Count: 13** — still under, which is the point of the reduction.

**Two baselines, labelled at every use.** RED-now cells measure *this deliverable's absence* and are pinned to `091966c55`. Citations of card t326's landed surfaces are pinned to **`origin/develop` at `ec15ec2cd`**, a diverged tree (diverged, `merge-base --is-ancestor` false — `spec.md` §A.10). Mixing the two pins silently would be an unattributed claim, so each t326 citation names its tree inline.

## §D AC Matrix

| AC | Requirement | Kind | Flipping milestone |
|---|---|---|---|
| AC-GDL-001 | REQ-GDL-001 | Contract conformance pair | M1 |
| AC-GDL-002 | REQ-GDL-004 | Partition trigger + anti-enumeration pair | M2 |
| AC-GDL-003 | REQ-GDL-003 | Unconditional invocation (T9) | M1 |
| AC-GDL-004 | REQ-GDL-002 | Negative assertion (no scheduled watcher) | M1 |
| AC-GDL-005 | REQ-GDL-005 | Zero-input surfacing pair | M2 |
| AC-GDL-006 | REQ-GDL-006 | Derived-age pair (two persisted results) | M2 |
| AC-GDL-007 | REQ-GDL-007 | Change-leading pair | M3 |
| AC-GDL-008 | REQ-GDL-008 | Negative assertion (no mutation) | M3 |
| AC-GDL-009 | REQ-GDL-009 | Additive-only diff assertion | M4 |
| AC-GDL-010 | REQ-GDL-010 | Composition + no-second-surface pair | M2 |
| AC-GDL-011 | REQ-GDL-011 | No-inline-query + declared-bound pair | M2 |
| AC-GDL-012 | REQ-GDL-012 | Refresh initiated, never awaited | M2 |
| AC-GDL-013 | REQ-GDL-013 | Contract-violation guard (absent / null / multi) | M2 |

## §D.1 Criteria (Given-When-Then, two cells each)

### AC-GDL-001 — the advisory consumes the contract and nothing more
**Given** two evaluation results whose classification vocabularies **differ in size and in value names**, each carrying a machine-readable designation of its own clean value, and where **each result also contains a non-clean value that folds to the clean surface value**; **When** the advisory is rendered from each; **Then** ALL THREE hold — (a) both render correctly, firing on exactly the non-clean entries, (b) the deliverable's own source contains **no** classification value name in code, verified by the instrument specified in §D.2 below returning rc=1, and (c) the clean/non-clean partition is derived from the carried designation — an entry that is non-clean but folds to the clean surface value **is** fired on.
- **RED-now:** no advisory path exists on `091966c55`. Measured:

  ```
  $ grep -rln "guard-liveness\|guardLiveness" .claude/hooks/ internal/
  (no output; rc=1)
  ```

  Red for absence — there is nothing to render from either vocabulary.
- **Green path:** M1. Passing output is both renders correct and the value-name grep returning rc=1.
- **Mutant (a)/(b):** clause (a) alone is satisfied by an implementation that hardcodes the state SPEC's value names and happens to handle both fixtures because it enumerates all of them. That mutant works today and breaks the moment the state SPEC adds a value — precisely the coupling the seam exists to prevent. Clause (b) makes the seam falsifiable rather than asserted, and is stated as a grep over source rather than as a design claim.
- **Mutant (c) — the under-firing mutant, and it is why the contract needed a fourth clause:** with only (a) and (b), the one route left open to a conforming implementation is to trigger on the **surface fold**, since that is machine-readable and names no classification. It under-fires: the producing SPEC folds more than one classification to its clean surface value while only one classification is clean, so every entry in that gap is silently treated as nothing-to-report. That mutant passes (a) on any fixture lacking such an entry and passes (b) outright. The fixture therefore *requires* such an entry, and clause (c) requires the partition to come from the carried designation — which is the only remaining route, and the one REQ-GDL-001 (iii) exists to supply.

### AC-GDL-002 — the trigger is the partition, not a list
**Given** an evaluation result carrying at least one non-clean entry, and separately a result in which **every** entry is non-clean but **no two share a classification**; **When** the harness evaluates whether to render; **Then** BOTH hold — (a) an advisory renders in both cases, and (b) the trigger's implementation branches on clean-vs-non-clean only, verified by the §D.2 instrument returning rc=1 over the trigger's own source.
- **RED-now:** no harness trigger exists (AC-GDL-001's measurement). Red for absence.
- **Green path:** M2. Passing output is both renders plus rc=1 on the grep.
- **Mutant:** this criterion exists because the same mutant killed this SPEC three times. A trigger enumerating classifications passes any fixture whose values happen to be enumerated, and silently fails on the first value the author did not think of — which is what D1 (two named classifications), N1 (those plus a coverage proxy) and T3 (two criteria still encoding the superseded list) each were. Clause (b) makes enumeration mechanically detectable instead of relying on a reviewer noticing it a fourth time.
- **Note on scope:** the second fixture deliberately contains no clean entry and no repeated value, so an implementation that special-cases "all entries share one classification" is not accidentally satisfied.

### AC-GDL-003 — the evaluator is invoked on every host activation (T9)
**Given** a fixture of N ≥ 5 host-surface activations **whose diffs touch no workflow file and no guard-related path**; **When** the host activates each time; **Then** ALL THREE hold — (a) exactly N **refresh initiations** are observed, (b) exactly N **refreshes reach the point of issuing subject queries** — measured at the query layer, not at the call site, so the count is unchanged by where a filter might sit, and (c) the count in (b) is independent of the fixture's diff content, demonstrated by a second fixture of N activations whose diffs **do** touch workflow files yielding the same count.
- **RED-now:** no evaluator, no invocation site, and no query layer exist on `091966c55` (AC-GDL-001's measurement). Red for absence — there is nothing to count at either layer.
- **Green path:** M1. Passing output is `N` initiations, `N` refreshes reaching the query layer, and the same count from both fixtures.
- **Mutant 1 — the call-site filter:** an evaluator gated at its invocation on whether the session's diff touched `.github/workflows/`. It satisfies "reachable from an already-attended surface" and every criterion covering the advisory, and **is `docs-i18n-check.yml`** — §A.3's own guard, rebuilt inside the deliverable. Clause (a) kills it: a filtered evaluator yields 0 initiations where an unconditional one yields N.
- **Mutant 2 — "invoked unconditionally, evaluates conditionally", which is why (b) and (c) exist:** an evaluator called on every activation that returns immediately when the session diff touches no workflow path. It passes a clause counting *invocations* and a clause *reading the call site*, because the filter has moved one frame inward — the same relocation this SPEC has now paid for three times. Clause (b) measures at the **query layer**, past any early return; clause (c) makes the independence explicit by requiring two fixtures with opposite diff content to yield the same count. Nothing else in the set reaches this path: AC-GDL-002, AC-GDL-005, AC-GDL-006 and AC-GDL-007 all supply a result as a `Given`, so none exercises the code that produces one.
- **Why the clauses are counts rather than a reading:** the earlier clause (b) said "verified by reading the wiring" — a human read, inside the criterion the SPEC itself calls load-bearing, while every other negative assertion here is a count or a grep against a measured baseline. It is a count now for the same reason those are.
- **Why this criterion is the load-bearing one:** §D.1's answer to constraint 1 is that the **refresh** is *entailed* by someone working. That entailment is false if the refresh is conditional at any layer, and a criterion stopping at the call site cannot tell the difference.

### AC-GDL-004 — no scheduled watcher is introduced
**Given** the merged change; **When** `.github/workflows/` is inspected; **Then** BOTH hold — (a) the count of workflow files carrying a `schedule:` trigger is unchanged from its measured baseline, and (b) **no job reachable from a `schedule:` trigger references the evaluator** — neither a new job in a scheduled workflow nor a new step in an existing scheduled job.
- **RED-now:** this is a **preserved invariant**, red only against a mutant, so its baseline is measured to make the comparison possible. On `091966c55`:

  ```
  $ grep -l '^  schedule:' .github/workflows/* | wc -l
         3          # codeql.yml, community.yml, release-drafter-cleanup.yml
  ```
- **Green path:** M1; the invariant holds throughout and is asserted at merge.
- **Mutant (a):** every other criterion in this set is satisfied by implementing the evaluator as a cron workflow that opens an issue — functionally an advisory, structurally the regress §D.1 rejects. Clause (a) excludes it, stated as a count against a measured baseline rather than as "no new schedule was added", which is unfalsifiable without one.
- **Mutant (b) — the file count is the wrong granularity on its own:** an implementation adding the evaluator as a **job inside one of the three existing scheduled workflows** leaves the file count at 3, satisfies clause (a), and is a scheduled watcher — exactly what REQ-GDL-002 forbids. Counting the files that *carry* a schedule cannot see a job added *under* one. Clause (b) asserts on schedule-reachable jobs, which is the granularity the requirement actually binds.

### AC-GDL-005 — the advisory arrives with no operator input
**Given** an evaluation result carrying at least one non-clean entry; **When** the operator begins a session and issues no command naming a guard, a workflow file, or the liveness feature itself; **Then** BOTH hold — (a) the advisory is rendered, and (b) the rendering path consumed no operator-supplied guard identifier or query string.
- **RED-now:** on `091966c55` there is no advisory, so the only way to learn a guard's last-fired time is a hand-written targeted query. Red because the answer is reachable **only** by someone who already knows which question to ask — the defect §A.4 records, not merely a missing feature.
- **Green path:** M2. Passing output is the advisory present in a session transcript containing no such command.
- **Mutant:** clause (a) alone is satisfied by shipping a `moai guard liveness` verb plus documentation telling the operator to run it. That renders an advisory, satisfies "the harness surfaces it", and leaves the operator needing to know the mechanism exists — the exact relocation §D.2 rejects. Clause (b) excludes it, and is stated as *inputs consumed* rather than as how the operator felt, so it is decidable.
- **Why the negative clause is load-bearing:** the lead session in §A.4 could run the correct query the moment it was handed the workflow's name. What it could not do was know a query was owed.

### AC-GDL-006 — the advisory's age is derived from the persisted result
**Given** **two** results persisted at distinct known times T₁ and T₂ (T₁ ≠ T₂, both non-zero ages at render time); **When** the advisory is rendered from each; **Then** BOTH hold — (a) each rendered age is non-zero, and (b) the two rendered ages **differ**, and each corresponds to its own persisted timestamp rather than to the render moment.
- **RED-now:** no advisory and no persisted result exist (AC-GDL-001's measurement). Red for absence.
- **Green path:** M2. Passing output is two distinct ages matching T₁ and T₂.
- **Mutant (single-fixture):** a renderer computing age from the moment of rendering always prints `0s`, satisfying a bare "states the age" criterion on every observation while producing exactly the failure the criterion prevents.
- **Mutant (constant-offset), which is why the fixture is two results:** a renderer printing a fixed non-zero age (`1h`) satisfies "age is non-zero" on a single-point fixture while deriving nothing from the persisted timestamp. The two-result fixture discriminates it — a constant-offset renderer produces the same age twice where a correct one produces two.

### AC-GDL-007 — the advisory leads with changes and survives a standing non-clean neighbour
**Given** two consecutive evaluations where entry set S is non-clean in both and entry T newly became non-clean in the second; **When** the second advisory is rendered; **Then** BOTH hold — (a) T appears in the advisory's leading position, and (b) the members of S are represented as a count and are not re-rendered as individual entries.
- **RED-now:** no advisory exists (AC-GDL-001's measurement). Red for absence.
- **Green path:** M3. Passing output shows T named and S collapsed to a count.
- **Mutant:** clause (a) alone is satisfied by a renderer printing the full standing list with T at the top — the block a reader learns to skip after the third session, and how a new advisory inherits the filter an always-red neighbour has already trained (`spec.md` §A.8). Clause (b) forces the collapse.
- **Bounded claim:** this asserts change-leading, not that the advisory gets read. Whether a compact standing count also becomes skippable is measured by nothing here and is recorded in §D.7.

### AC-GDL-008 — the advisory path mutates nothing
**Given** the advisory path and a fixture whose result carries at least one non-clean entry — the case most likely to tempt an action; **When** one render is executed; **Then** BOTH hold — (a) the run issues **zero** mutating forge calls (no issue creation, no comment, no dispatch, no workflow re-run), counted at the call layer, and (b) the working tree and the repository are byte-identical before and after, verified by `git status --porcelain` across the run.
- **RED-now:** no advisory path exists (AC-GDL-001's measurement). Red for absence — there is no call path to count.
- **Green path:** M3. Passing output is a zero mutating-call count and an unchanged tree.
- **Mutant:** clause (a) alone is satisfied by a renderer that writes its result cache into the repository working tree — no forge mutation, but a git write that shows up as drift for the next reader. Clause (b) catches it. Conversely (b) alone is satisfied by a renderer that opens an issue and touches no file, which is the shape AC-GDL-004's mutant note anticipates and closes only the `schedule:` half of.
- **Consequence, deliberate:** the persistence REQ-GDL-007 needs must live outside the working tree, or clause (b) fails. That constraint is an output of this criterion, not an oversight.

### AC-GDL-009 — the doctrine addition is additive only
**Given** the M4 commit touching `.claude/rules/moai/development/verification-completeness.md`; **When** `git diff --numstat 091966c55 -- <that file>` is read; **Then** BOTH hold — (a) the deleted-line count is `0`, and (b) the added text contains a continued-firing clause, matched by a grep that returns rc=1 on the baseline.
- **RED-now (b):** measured on `091966c55`:

  ```
  $ grep -nE "last.fired|continued.firing|stopped firing|liveness|stale guard" \
      .claude/rules/moai/development/verification-completeness.md
  (no output; rc=1)
  ```

  Red because the clause is absent — and this same command is the criterion's own discriminator, so its baseline rc=1 is what makes a later rc=0 meaningful.
- **RED-now (a):** trivially satisfied on the baseline (no diff exists yet); asserted at M4 as a preserved invariant, not as a starting observation.
- **Green path:** M4. Passing output is `<added> 0 <file>` from `--numstat`, with the grep returning rc=0.
- **Mutant:** the grep clause alone is satisfied by a commit that adds the clause and rewrites the surrounding section. The zero-deletions clause enforces "extend, never re-author".

### AC-GDL-010 — the advisory joins the existing block, and opens no second surface
**Given** the merged deliverable and an evaluation result carrying a non-clean entry; **When** the session-start path and the hook package are inspected; **Then** BOTH hold — (a) the advisory's text is emitted through the **existing** contributor helper on that surface, called from within an **already-registered** session-start handler — verified by a call-site check on the deliverable's own code, and (b) the count of **session-start handlers** is **unchanged** from its measured baseline, where a session-start handler is defined as *a type whose `EventType()` method returns `EventSessionStart`*.
- **RED-now:** no advisory exists on `091966c55` (AC-GDL-001's measurement) — no call site to check. Red for absence.
- **Baseline, re-derived from the stated definition on `origin/develop` at `ec15ec2cd`:**

  ```
  $ git grep -A2 "func.*EventType() EventType" origin/develop -- internal/hook \
      | grep -v _test | grep "EventSessionStart"
  internal/hook/auto_update.go-           return EventSessionStart
  internal/hook/handoff_inject.go:func (h *handoffInjectHandler) EventType() EventType { return EventSessionStart }
  internal/hook/session_start.go-         return EventSessionStart
  internal/hook/session_start_compact.go:func (h *sessionStartCompactHandler) EventType() EventType { return EventSessionStart }
  ```

  **Baseline = 4 handlers**, in `auto_update.go`, `handoff_inject.go`, `session_start.go`, `session_start_compact.go`. The two declaration forms — single-line and multi-line — are both captured, which is why the command reads two lines of context rather than matching on one shape.

  An earlier draft cited **7**, which reproduces under no definition. It came from `grep -ln EventSessionStart`, a count of files *mentioning* the token — which includes `registry.go` and `types.go` (the dispatch and the enum) and `session_start_binary_lag.go`, **which registers no handler at all**. The number was carried without re-deriving it, in a criterion whose whole purpose is to replace a judgment with a count. The definition was not adjusted to preserve the figure; the figure was replaced.
- **The composition target, measured on the same tree, and it is the pattern this criterion mandates:** `internal/hook/session_start.go:479` calls `appendAdditionalContext(out, binaryLagAdvisory(ctx, lagRoot))` — the binary-lag advisory is emitted **from inside the existing `sessionStartHandler`**, through the shared helper, registering nothing new. Clause (a) asks this deliverable to do the same thing.
- **Green path:** M2. Passing output is a call to the existing helper from within an existing handler, and a handler count still reading 4.
- **Why the earlier wording did not bite, and this is the point:** it asserted that the advisory's text appears "within that same block, separated from the earlier contributor". The **host runtime concatenates every hook's additional context unconditionally** — measured in `internal/hook/registry.go` on `origin/develop` — so *any* implementation satisfies that, including the second-surface mutant the criterion existed to exclude. The clause was measuring the platform, not the deliverable. Both clauses are now stated against artefacts this deliverable controls.
- **Mutant:** clause (a) alone is satisfied by an implementation that calls the helper from a **newly registered** handler — helper reused, surface still doubled. Clause (b)'s count catches it, in the manner of AC-GDL-004's schedule-count invariant: a countable assertion against a measured baseline rather than a judgment about structure.
- **Why a second surface is the wrong answer even though it is easier:** §A.8's mechanism is that a channel carrying no information gets filtered. Two channels for one concern is the same mechanism applied twice — each carries less than the single channel would, and a reader who learns to skip one has no reason to treat the other differently.

### AC-GDL-011 — the render performs no forge query
**Given** the advisory rendering at the host surface with the forge unreachable (network denied); **When** the session starts; **Then** ALL FOUR hold — (a) the advisory still renders from the persisted result, (b) **zero** forge calls are issued by the render path, counted the same way AC-GDL-001 counts value names, (c) the deliverable **declares its own render join bound** as a named constant, and that declared value is **≤ 250 ms**, and (d) the render completes within that declared bound.
- **RED-now:** no advisory and no persisted result exist (AC-GDL-001's measurement). Red for absence.
- **Green path:** M2. Passing output is a rendered advisory, zero forge calls, and a render inside the bound.
- **Why this criterion exists, and it was not in the design until t326 was read:** the comparable landed path bounds itself at 250 ms — `const binaryLagJoinBound = 250 * time.Millisecond`, measured on `origin/develop` at `ec15ec2cd`. t326 can afford an inline comparison because it is two short local `git` invocations; this SPEC's evaluator issues **one forge query per subject**, 18 on this repository, and no sequence of network round-trips fits that bound.
- **Mutant (a)/(b):** clause (a) alone is satisfied by an implementation that queries the forge inline and succeeds on a fast network. It passes every test on a healthy connection and stalls session start on a slow one — a latency defect appearing only where it hurts. The network-denied fixture makes it fail at test time rather than at a user's session start.
- **Mutant (c) — an undeclared bound is an unbounded one:** a clause asserting only "completes within the host surface's join bound" has no referent, because **no such shared bound exists**. Each contributor on that surface carries its own; the 250 ms figure is the comparable landed contributor's own constant, cited as precedent rather than as a platform limit. An implementer who sets their bound to 30 s satisfies such a clause trivially while session start blocks for 30 s on a slow network — the exact failure the clause exists to prevent. Clause (c) therefore requires a *declared* constant **and** caps it, and clause (d) measures against the declared value.

### AC-GDL-012 — the refresh is initiated and never awaited
**Given** a refresh whose subject queries are made to take **longer than the declared render join bound** (a stalled-forge fixture); **When** the host surface activates; **Then** ALL THREE hold — (a) the refresh is initiated, (b) the render completes within the declared bound **without** waiting for it, rendering from the previously persisted result, and (c) when the refresh later completes, its result is persisted and is the one a **subsequent** activation reads.
- **RED-now:** no refresh, no render and no persistence exist on `091966c55` (AC-GDL-001's measurement). Red for absence.
- **Green path:** M2. Passing output is a render inside the bound during a stalled refresh, followed by the stalled refresh's result appearing at the next activation.
- **Mutant:** clauses (a) and (b) alone are satisfied by an implementation that initiates the refresh, abandons it at the bound, and **discards** the result — the render is fast, the entailment looks intact, and the persisted result never advances, so every activation renders the same increasingly stale verdict while REQ-GDL-006's age disclosure quietly grows. Clause (c) is what requires the abandoned-for-this-turn work to still land.
- **Why this criterion follows from D2's resolution:** binding the refresh to activation (REQ-GDL-003) is only compatible with a latency-bounded host if the refresh is never awaited (REQ-GDL-012). This criterion is where that compatibility is actually measured, on the fixture where the two obligations conflict — a refresh slower than the bound.

### AC-GDL-013 — a contract-violating result is reported, never rendered green
**Given** **five** results that each violate REQ-GDL-001's contract — designation **absent** (i), **null** (ii), **multi-valued** (iii), and an entry carrying **zero** classifications (iv) or **more than one** (v) — each otherwise well-formed and carrying entries; **When** the harness processes each; **Then** BOTH hold — (a) each run renders an advisory naming the contract violation, and (b) **no run reports an all-clear**, including the cases where every entry would classify clean under some guess.
- **RED-now:** no harness and no contract handling exist on `091966c55` (AC-GDL-001's measurement). Red for absence — there is nothing to feed a malformed result to.
- **Green path:** M2. Passing output is three advisories naming the violation, and zero all-clears.
- **Mutant, and it is this card's own subject at the consumer's layer:** an implementation that partitions entries by comparing each classification against the designation. On an absent or null designation **no comparison succeeds**, so nothing is identified as non-clean, so REQ-GDL-004's trigger never fires and **nothing renders** — silently, on a well-formed-looking input. Every other criterion in this set supplies a contract-conforming result as its `Given`, so none exercises this path. The multi-valued case is included because it fails differently: a comparison against the *first* designated value succeeds and under-fires rather than not firing at all.
- **Why this is not a re-opening of the seam defect:** that defect was *the consumer cannot identify the clean value*. This is *the consumer has no defined behaviour when it cannot* — the requirement was added, and the failure mode it opened was left undefined. The producing SPEC already carries the symmetric guard (it refuses an all-clear while its queried count is zero or its enumeration returned nothing); this is its mirror, and the seam stays symmetric because both sides now refuse to be green about an input they could not read.
- **Fixtures (iv) and (v) exist because a partial mirror fails silently.** The producing side guarantees **three** things — one classification per entry, one clean value, and a carried designation — and an earlier draft of this criterion mirrored only the designation clause. An entry carrying two classifications violates the first guarantee, matches the clean value on comparison, and is treated as nothing-to-report: it **under-fires**, which is the same failure the surface-fold route produced before the designator existed. Covering the easiest clause of a three-clause contract is how the fold route under-fired in the first place, one clause inward.
- **Why the enumeration is exhaustive, not illustrative:** REQ-GDL-013's preamble is broad ("*violates the contract in any respect*") while its list is specific, and an implementer reads the list as definitional. The list is therefore stated as *exactly the negations of the contract's clauses*, so widening the contract obliges widening this criterion rather than silently leaving a clause unmirrored.

## §D.2 The seam instrument (specified, because an unrunnable instrument is judgment wearing a count's clothes)

AC-GDL-001(b) and AC-GDL-002(b) are the seam contract's only mechanical enforcement. The instrument is therefore specified here in full rather than described, and it is **scoped, anchored, and case-sensitive** — an unanchored token search is not merely noisy, it is unsatisfiable.

```
grep -rnE '\b(OK|STALE|UNKNOWN|UNDECLARED|UNREADABLE|UNRESOLVED|ORPHANED)\b' <deliverable's own non-test source files> \
  | grep -v '^[^:]*:[0-9]*:[[:space:]]*//'
```

Passing is **rc=1** (no match). Three properties are load-bearing:

- **Case-sensitive and word-anchored.** Measured on this tree, the bare substring `OK` matches **79 files** under `internal/hook/` and 11 times in `session_start.go` alone, because it is a substring of `StatusOK`, `CheckOK`, and the English word `ok`. Word-anchored and case-sensitive, `\bOK\b` matches **6** lines in that whole package — and matches neither `StatusOK` nor `CheckOK`, since no word boundary exists inside them. Without the anchor the criterion returns rc=0 on virtually any Go source and would be *interpreted* at verification time rather than run, which is the substitution of judgment for a count this SPEC has paid for at three separate defects.
- **Scoped to the deliverable's own files**, not to a package. Measured: even the six unambiguous tokens match **5** lines across `internal/hook/`, every one of them prose in a comment or a test name. A package-wide sweep produces false positives that force an interpretive step back into the check.
- **Comment lines excluded.** The seam's substance is *code coupling*: a comment naming a classification does not break when the vocabulary changes, whereas a code reference does. Permitting prose while forbidding code use is the honest line, and the exclusion is a stated pattern rather than a reader's judgment.

**This section is itself the one place in this SPEC that names the vocabulary, and that is unavoidable rather than an oversight** — an instrument cannot search for tokens without naming them. It is not coupling: the instrument runs *against* the deliverable and ships as no part of it, so a vocabulary change edits the pattern here and touches no requirement, no design section, and no delivered line. The consequence for verification is stated so nobody has to rediscover it: **the artifact-level falsification sweep — a value-token grep over this SPEC's three files — must exclude this section and the HISTORY row describing it**, and returns 0 elsewhere. Running that sweep after the §D.2 addition is what surfaced this, which is the difference between running a check and declaring one.

## §D.7 Residual risk (recorded, not claimed as closed)

- **The host surface can be removed, and the evaluator stops with it** — the same defect class one layer up. Full closure needs an unattended watcher, which reintroduces the regress `spec.md` §D.1 avoids.
- **The host surface can stop invoking without being removed.** REQ-GDL-003 forecloses the direct form and AC-GDL-003 verifies it on a fixture, but that binds this deliverable's wiring at merge time: it cannot stop a later edit from reintroducing a filter, and **no criterion measures invocation frequency over the deliverable's lifetime**. This is the same shape as §A.3 and it is not closed.
- **The single trigger fires on every sweep once any entry is permanently non-clean.** The failure direction the trigger's own repair created. REQ-GDL-007 mitigates by collapsing standing entries to a count; the mitigation is partial.
- **Change-leading makes a long-standing non-clean entry quiet by design.** Announced once, thereafter only counted. Re-announcing every session is the noise that produces the filter, so the trade is deliberate — but quiet is this card's subject, and this is the sharpest unresolved tension.
- **The advisory reaches one observer, and §A.4's subject is two.** An advisory in observer 1's session leaves observer 2 where §A.4 found the lead. The design narrows *who must already know the question* from "someone" to "whoever attends a session" — an improvement, not a closure.
- **The seam is verified by a grep, which is a proxy for a design property.** AC-GDL-001(b) and AC-GDL-002(b) detect a hardcoded value name in source. They do not detect an implementation that infers the vocabulary structurally — for instance by assuming a fixed value count. The contract is stated in three clauses and only two of them are mechanically checked.
- **Composing into the landed session-start block changes that block's noise profile.** t326's advisory speaks on one outcome and is silent on every other; REQ-GDL-004 speaks on any non-clean entry. Joining a speak-rarely channel with a speak-often contributor has a cost, and §A.8 is about exactly what a noisier channel does to its readers. Change-leading mitigates it partially. A second surface was rejected as worse, not as costless.
- **The advisory can be one activation stale by construction.** The refresh trigger *is* decided — host activation, unconditionally (REQ-GDL-003) — but its completion is never awaited (REQ-GDL-012), so a reader sees the most recent *completed* refresh rather than one taken this instant. AC-GDL-012 covers the mechanism. What remains open is the **magnitude**: on a surface visited rarely, or where the refresh consistently overruns, the persisted result can be arbitrarily old while the advisory renders normally. REQ-GDL-006's age disclosure makes that legible without bounding it, and bounding it would require awaiting the refresh, which the latency bound forbids.

  *This bullet replaced one asserting the refresh trigger was undecided and uncovered by any criterion — every clause of which the D2 repair had falsified. The repair updated `spec.md` §D.3 and left its twin here, in the register a reader consults precisely to learn what is still open; the stale text told them the design had a hole its requirements had closed. Recorded rather than silently swapped, because this is the second time on this card that a fix moved one surface and left its mirror behind.*
- **The composition target is owned by a SPEC that has not closed.** Card t326's implementation surfaces are landed on `origin/develop` at `ec15ec2cd`, but its SPEC reads `status: in-progress`. REQ-GDL-010 binds this deliverable to compose onto a surface whose owning card can still change it — the contributor helper, the merge semantics, or the bound convention could move before t326 closes, and AC-GDL-010's baseline of 4 handlers is measured against that moving tree. Nothing here would detect such a change until the composition broke. *(Present in `spec.md` §D.3 since v1.3.0 and missing here until v1.4.0 — found by sweeping the twin registers rather than by an audit.)*
- **The contract's third clause is consumed, not verified.** REQ-GDL-001 asserts that exactly one value means *nothing to report*. This SPEC checks that its own trigger uses that partition; it cannot check that the producing SPEC actually designates exactly one such value. If the state model ever designated two clean values, this SPEC's criteria would all still pass while the advisory under-fired. That is the cost of the contract seam, and it is the state SPEC's REQ-GSM-012 that carries the other half.
