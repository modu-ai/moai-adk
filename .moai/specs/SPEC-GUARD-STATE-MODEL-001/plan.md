# SPEC-GUARD-STATE-MODEL-001 — Implementation Plan

Baseline tree for every measurement: **`091966c55`** @ `WT-guard-liveness` (worktree `.claude/worktrees/t333`).

Milestones are ordered by **decision reversibility**. The manifest schema and the state table lead: every later milestone reads them, and changing the table after the census is populated means re-deciding every entry.

## §A Context

This SPEC was created by the scope reduction of `SPEC-GUARD-LIVENESS-001` after its iter-3 FAIL + STOP. It receives the state model — the family in which that SPEC's hardest defects all landed — and three unresolved findings as **starting material**: T2, N2's unresolved half, and T4 (`spec.md` §A.3).

### A.1 This SPEC's card id

> **`<CARD-PENDING>`** — this SPEC's card id has been requested from the lead and is not yet assigned. **This line is the single place the placeholder sits in this artifact set.** When the id arrives, replace it here and in `SPEC-GUARD-LIVENESS-001/plan.md` §A.1 (the corresponding single placeholder there); nothing else in either artifact set references it.

The sibling SPEC does **not** `depends_on` this one. The seam is a published contract (`spec.md` §B), so the two are independently implementable.

## §B Known issues and constraints carried in

- **B-0 — prose is how the predecessor's state model failed.** A five-value vocabulary covered at least six states and every requirement read complete in isolation. The state table (`spec.md` §C.2 REQ-GSM-006) is the normative artifact for that reason; the prose describes it and does not extend it. **Do not add a classification rule in a sentence — add a row.**
- **B-1 — T2, inherited unresolved.** A failed forge query had no admissible classification: one requirement routed every non-retention failure into the no-reader value, another defined that value as no-reader-only, and the criteria assumed a branch the requirements forbade. Row 2 of the table and AC-GSM-009 resolve it. It is the most likely degraded run in production.
- **B-2 — T1, the disk side of the set comparison.** The comparison has two inputs and only the query side was guarded. An enumeration that silently returns empty reports all-clear honestly about a set it never learned. REQ-GSM-010 and AC-GSM-012 guard the second input.
- **B-3 — T4, plan/requirement drift under repair.** The predecessor's five-value set landed in the requirement layer and never reached the plan's milestone body, and a criterion was scheduled at a milestone that could not deliver it. Mitigation here: the table is the single artifact both layers cite, and criteria that genuinely split across milestones are clause-split in the flip lists below (AC-GSM-003, AC-GSM-005).
- **B-4 — field-level inertness.** The predecessor carried two counts that were required, rendered, and consumed by nothing. REQ-GSM-009 clause (b) and AC-GSM-013 make every count trace to a consumer.
- **B-5 — `moai update` deletes `.moai/config/` wholesale.** The manifest lives outside that root (REQ-GSM-001), or it is a liveness record that itself silently stops.
- **B-6 — the manifest is repository-specific, not template content.** It describes this repository's `.github/workflows/`. It must not be mirrored into `internal/template/templates/`.

## §C Pre-flight (measured on `091966c55`)

| Check | Command | Result |
|---|---|---|
| Manifest absent | `ls .moai/guards` | `No such file or directory`, rc=1 |
| No evaluator verb | `grep -n '"guard"' internal/cli/*.go \| grep -v _test.go` | `internal/cli/constitution.go:49` only — the unrelated `moai constitution guard` verb |
| Workflow census | `ls -1 .github/workflows/*.yml .github/workflows/*.yaml \| wc -l` | `18` |
| File-to-name bijection | `grep -h '^name:' .github/workflows/*.yml .github/workflows/*.yaml \| sort -u \| wc -l` | `18` — the census subtraction's premise holds on this tree |
| Release-only subjects exist | `.github/workflows/spec-status-auto-sync.yml`, `release.yml` present | both present — the entries AC-GSM-004 requires, and what makes table row 5 reachable |

## §D Constraints

- Read-only against the forge (REQ-GSM-011). No issue creation, no dispatch, no re-run, no commit.
- No working-tree write; any persistence lives outside the tree (AC-GSM-014(b)).
- The manifest lives outside `.moai/config/` (B-5) and is not mirrored into the template tree (B-6).
- Exactly one classification value is designated clean (REQ-GSM-012) — the sibling SPEC's contract depends on it and cannot check it.

## §E Self-verification

Every criterion in `acceptance.md` carries a RED-now cell pinned to `091966c55` with its stated reason, and a green-path cell naming its flipping milestone. Every cited command was run on this tree during authoring; **no cell was carried across the scope reduction without re-measurement**, per the predecessor's D7 finding.

## §F Milestones

### M1 — the manifest schema and the state table (least reversible)

The two decisions worth the most review attention. Every later milestone reads them, and changing the table after 18 entries exist means re-deciding all of them.

- Define the manifest schema: locator, expected events, window, measured quantity (REQ-GSM-002), plus the release-cycle-conditional form (REQ-GSM-004).
- Fix the measured-quantity vocabulary at exactly three values with the conclusion set each admits (REQ-GSM-003).
- **Author the state table as the normative artifact** (REQ-GSM-006) — 7 rows, 6 values, each value with a distinct implied action. Check totality in both directions **before** populating the census: every row maps to exactly one value, and every value is reachable from at least one row (REQ-GSM-007). An uncovered condition must appear as a missing row, never as a sentence.
- Apply the subject-agnostic shape test: each entry carries kind, locator and cadence as data. If accepting a second-kind entry requires changing the schema or either vocabulary, the schema has hardcoded its subject — reshape it now (B-0's sibling failure mode).
- Choose the manifest path outside `.moai/config/` (B-5).
- Populate the census: one entry per workflow file, 18 of 18, including the two release-only subjects.

Flips: AC-GSM-001, AC-GSM-002, AC-GSM-003 (a)(b), AC-GSM-004, AC-GSM-005 (b).

### M2 — the classifier

- Query per subject; never a repository-global listing (REQ-GSM-005).
- Decide every entry by the table (REQ-GSM-006). The classifier is table-driven, so the 7-row fixture in AC-GSM-007 is a direct test of the artifact rather than of a paraphrase of it.
- Row 2 — a query that did not return classifies `UNRESOLVED`, never `UNKNOWN` (B-1).
- Rows 5 and 6 — the declared condition is what separates excused absence from unexcused; identical observable input, two classifications.
- Row 7 — the set comparison against disk enumeration (REQ-GSM-008).
- Refuse an all-clear when the queried count is zero **or** the enumeration returned zero files (REQ-GSM-010, B-2).

Flips: AC-GSM-003 (c), AC-GSM-005 (a), AC-GSM-006, AC-GSM-007, AC-GSM-008, AC-GSM-009, AC-GSM-010, AC-GSM-011, AC-GSM-012.

### M3 — the result and its contract

- Emit the measurement timestamp and the count set, with **each count traced to a consumer** (REQ-GSM-009, B-4).
- No forge mutation, no working-tree write (REQ-GSM-011).
- Publish the contract: exactly one designated clean value (REQ-GSM-012).

Flips: AC-GSM-013, AC-GSM-014, AC-GSM-015.

**Milestone map check.** All 15 criteria appear across the flip lists, and each listed milestone can actually deliver what it is assigned — a union count answers "is every criterion listed?" and is structurally blind to "can that milestone deliver it?", which is how T4 escaped the predecessor's repair. AC-GSM-003 and AC-GSM-005 are clause-split for that reason: their schema halves land at M1, their classifier halves at M2.

## §G Anti-patterns to avoid

- **Adding a classification rule as a sentence instead of a row.** B-0. The table is normative; prose that extends it silently recreates the uncovered-condition failure.
- **Reusing `UNKNOWN` as a general "could not determine" bucket.** Its implied action is *look again with a longer window*, which is wrong advice for an auth failure and meaningless for a kind with no reader. Two states with different implied actions under one label is the defect that produced this SPEC.
- **Guarding the query side and not the disk side.** B-2. `UNDECLARED: 0` is indistinguishable between a complete census and an enumeration that found nothing.
- **Building the classifier to read outcomes only.** "For each manifest entry, fetch its runs and judge them" is accurate about the entries it holds and silent about the files it does not. The set comparison is not an add-on to that loop — it is what makes the loop's green informative.
- **Carrying a count nothing consumes.** B-4.
- **Writing an expectation whose one number measures two events.** "Fires every N days" awards full marks to a guard that runs faithfully and catches nothing.
- **Designating a second clean value.** It would break the sibling SPEC's contract silently: every criterion there would still pass while its advisory under-fired (AC-GSM-015).
- **Mirroring the manifest into the template tree.** B-6.

## §H Cross-references

- `.moai/specs/SPEC-GUARD-LIVENESS-001/` — the sibling SPEC consuming this one's published contract.
- `.moai/reports/plan-audit/SPEC-GUARD-LIVENESS-001-review-3.md` — the FAIL + STOP and the scope-reduction recommendation this SPEC implements; T1, T2, T4 are enumerated there in full.
- `.moai/reports/t333/trigger-axis-observation.md` — the empirical grounding for the firing-expectation vocabulary.
- `.claude/rules/moai/development/verification-completeness.md` — the two-cell adoption discipline and the mutant probe this SPEC's criteria follow.
