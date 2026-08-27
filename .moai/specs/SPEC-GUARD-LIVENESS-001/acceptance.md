# SPEC-GUARD-LIVENESS-001 — Acceptance Criteria (card t333, surfacing model)

Baseline tree for every RED-now cell: **`091966c55`** @ `WT-guard-liveness` (worktree `.claude/worktrees/t333`). Every cited command was re-run on this tree during the scope reduction; **no cell was carried across the split without re-running its command**, per the audit's D7 finding that an unreproducible measurement inside a RED-now cell is the evidence-integrity failure the two-cell discipline exists to prevent.

Two-cell discipline per `.claude/rules/moai/development/verification-completeness.md` §2: each criterion carries a RED-now cell stating **why** it is red, and a green-path cell naming the flipping milestone. Each was put through that rule's §2 mutant probe; where a single-clause criterion admitted a mutant, a second clause was added and the mutant is recorded.

Budget: Tier M ≤ 16 acceptance criteria. **Count: 9** — comfortably under, which is the point of the reduction.

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

## §D.1 Criteria (Given-When-Then, two cells each)

### AC-GDL-001 — the advisory consumes the contract and nothing more
**Given** two evaluation results whose classification vocabularies **differ in size and in value names**, each designating one value as clean; **When** the advisory is rendered from each; **Then** BOTH hold — (a) both render correctly, firing on exactly the non-clean entries, and (b) the deliverable's own source contains **no** classification value name, verified by a grep for the state SPEC's value tokens returning rc=1 over this deliverable's files.
- **RED-now:** no advisory path exists on `091966c55`. Measured:

  ```
  $ grep -rln "guard-liveness\|guardLiveness" .claude/hooks/ internal/
  (no output; rc=1)
  ```

  Red for absence — there is nothing to render from either vocabulary.
- **Green path:** M1. Passing output is both renders correct and the value-name grep returning rc=1.
- **Mutant:** clause (a) alone is satisfied by an implementation that hardcodes the state SPEC's five-or-six value names and happens to handle both fixtures because it enumerates all of them. That mutant works today and breaks the moment the state SPEC adds a value — which is precisely the coupling the seam exists to prevent. Clause (b) is what makes the seam falsifiable rather than merely asserted, and it is stated as a grep over source rather than as a design claim.

### AC-GDL-002 — the trigger is the partition, not a list
**Given** an evaluation result carrying at least one non-clean entry, and separately a result in which **every** entry is non-clean but **no two share a classification**; **When** the harness evaluates whether to render; **Then** BOTH hold — (a) an advisory renders in both cases, and (b) the trigger's implementation branches on clean-vs-non-clean only, verified by the same value-name grep as AC-GDL-001(b) returning rc=1 over the trigger's own source.
- **RED-now:** no harness trigger exists (AC-GDL-001's measurement). Red for absence.
- **Green path:** M2. Passing output is both renders plus rc=1 on the grep.
- **Mutant:** this criterion exists because the same mutant killed this SPEC three times. A trigger enumerating classifications passes any fixture whose values happen to be enumerated, and silently fails on the first value the author did not think of — which is what D1 (two named classifications), N1 (those plus a coverage proxy) and T3 (two criteria still encoding the superseded list) each were. Clause (b) makes enumeration mechanically detectable instead of relying on a reviewer noticing it a fourth time.
- **Note on scope:** the second fixture deliberately contains no clean entry and no repeated value, so an implementation that special-cases "all entries share one classification" is not accidentally satisfied.

### AC-GDL-003 — the evaluator is invoked on every host activation (T9)
**Given** a fixture of N ≥ 5 host-surface activations **whose diffs touch no workflow file and no guard-related path**; **When** the host activates each time; **Then** BOTH hold — (a) exactly N evaluator invocations are observed, and (b) the invocation site contains no path filter, changed-file test, or subject-matter condition, verified by reading the wiring.
- **RED-now:** no evaluator and no invocation site exist on `091966c55` (AC-GDL-001's measurement). Red for absence — there is no invocation to count.
- **Green path:** M1. Passing output is `N` invocations from `N` activations.
- **Mutant — the one the audit named the most consequential in the SPEC:** an evaluator gated on whether the session's diff touched `.github/workflows/`. It satisfies "reachable from an already-attended surface", satisfies every criterion covering the advisory, and **is `docs-i18n-check.yml`** — §A.3's own guard, rebuilt inside the deliverable. The N-activations-with-no-matching-diff fixture is what discriminates it: a filtered evaluator produces 0 invocations where an unconditional one produces N.
- **Why this criterion is the load-bearing one:** §D.1's entire answer to constraint 1 is that the evaluator's firing is *entailed* by someone working. That entailment is false if invocation is conditional, and until this criterion the clause asserting it was verified by nothing.

### AC-GDL-004 — no scheduled watcher is introduced
**Given** the merged change; **When** `.github/workflows/` is inspected; **Then** the count of workflow files carrying a `schedule:` trigger is unchanged from its measured baseline.
- **RED-now:** this is a **preserved invariant**, red only against a mutant, so its baseline is measured to make the comparison possible. On `091966c55`:

  ```
  $ grep -l '^  schedule:' .github/workflows/* | wc -l
         3          # codeql.yml, community.yml, release-drafter-cleanup.yml
  ```
- **Green path:** M1; the invariant holds throughout and is asserted at merge.
- **Mutant:** every other criterion in this set is satisfied by implementing the evaluator as a cron workflow that opens an issue — functionally an advisory, structurally the regress §D.1 rejects. This criterion is the negative assertion that excludes it, stated as a count comparison against a measured baseline rather than as "no new schedule was added", which is unfalsifiable without one.

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

## §D.7 Residual risk (recorded, not claimed as closed)

- **The host surface can be removed, and the evaluator stops with it** — the same defect class one layer up. Full closure needs an unattended watcher, which reintroduces the regress `spec.md` §D.1 avoids.
- **The host surface can stop invoking without being removed.** REQ-GDL-003 forecloses the direct form and AC-GDL-003 verifies it on a fixture, but that binds this deliverable's wiring at merge time: it cannot stop a later edit from reintroducing a filter, and **no criterion measures invocation frequency over the deliverable's lifetime**. This is the same shape as §A.3 and it is not closed.
- **The single trigger fires on every sweep once any entry is permanently non-clean.** The failure direction the trigger's own repair created. REQ-GDL-007 mitigates by collapsing standing entries to a count; the mitigation is partial.
- **Change-leading makes a long-standing non-clean entry quiet by design.** Announced once, thereafter only counted. Re-announcing every session is the noise that produces the filter, so the trade is deliberate — but quiet is this card's subject, and this is the sharpest unresolved tension.
- **The advisory reaches one observer, and §A.4's subject is two.** An advisory in observer 1's session leaves observer 2 where §A.4 found the lead. The design narrows *who must already know the question* from "someone" to "whoever attends a session" — an improvement, not a closure.
- **The seam is verified by a grep, which is a proxy for a design property.** AC-GDL-001(b) and AC-GDL-002(b) detect a hardcoded value name in source. They do not detect an implementation that infers the vocabulary structurally — for instance by assuming a fixed value count. The contract is stated in three clauses and only two of them are mechanically checked.
- **The contract's third clause is consumed, not verified.** REQ-GDL-001 asserts that exactly one value means *nothing to report*. This SPEC checks that its own trigger uses that partition; it cannot check that the producing SPEC actually designates exactly one such value. If the state model ever designated two clean values, this SPEC's criteria would all still pass while the advisory under-fired. That is the cost of the contract seam, and it is the state SPEC's REQ-GSM-012 that carries the other half.
