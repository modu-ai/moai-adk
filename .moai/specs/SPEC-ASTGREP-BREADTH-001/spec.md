---
id: SPEC-ASTGREP-BREADTH-001
title: "ast-grep rule breadth across the ten uncovered languages, and corpus extension"
version: "0.1.0"
status: draft
created: 2026-08-25
updated: 2026-08-25
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/template/templates/.moai/config/astgrep-rules, internal/hook/security/testdata/scan-corpus"
lifecycle: spec-anchored
tags: "ast-grep, ruleset, security, multi-language, differential-corpus, breadth, successor"
tier: L
depends_on: [SPEC-ASTGREP-LANG16-001]
---

# SPEC-ASTGREP-BREADTH-001 — ast-grep rule breadth and corpus extension

> **SKELETON — plan-phase not yet run.** This document carries the ID, scope statement, milestone
> list, and dependency contract only. Its full GEARS requirement set and acceptance criteria are
> authored in its own plan phase, once `SPEC-ASTGREP-LANG16-001` has landed the contract this SPEC
> consumes. §C states why they are deliberately absent rather than pending.

## HISTORY

| Date | Version | Change |
|---|---|---|
| 2026-08-25 | 0.1.0 | Skeleton authored as part of `SPEC-ASTGREP-LANG16-001` v0.4.0's split (plan-audit iteration 1, finding D8). Scope, milestones, and the inherited contract recorded; requirements deferred to this SPEC's own plan phase. |

---

## §A. Context

### A.1 Why this SPEC exists

Card t228 asks for ast-grep coverage across all sixteen supported languages. The work divides
into a **contract** and a **breadth**, and plan-audit forced them into separate SPECs: the
predecessor's requirement and acceptance budgets were saturated at 25/25, so its own audit
findings could not be absorbed without breaching the Tier L ceiling.

`SPEC-ASTGREP-LANG16-001` builds the contract — the `sg test` harness, the coverage matrix and its
four-class checker, the severity predicate, and the `metadata.cwe` anchor. **This SPEC writes the
rules against that contract**, and extends the differential corpus to observe them.

**The split is sequencing, not scope reduction.** Card t228 requirement (2) — per-language
security rules plus idiom rules — and requirement (4) — differential corpus expansion — are
discharged **here, in full**. Neither is deferred beyond this SPEC, and no external gate stands
between the two SPECs: the predecessor is committed work in the same card's lane.

### A.2 Why the contract comes first

Two of the predecessor's defenses were falsified during its plan-audit: R2's anti-contrivance
mitigation (a rule matching `zzzNeverRealApi($X)` passed `sg test` at exit 0 and qualified for
`severity: error`), and the differential corpus's covered-language gate (it calls `t.Skip`, not a
failure, and the skip disables all twelve assertions at once).

Each correction cost roughly one requirement. Had the 80 rules already been written against those
defenses, the same correction would have invalidated 80 rules. That is the concrete argument for
this ordering, and it is measured rather than stylistic.

### A.3 The scale this SPEC carries

| Driver | Volume |
|---|---:|
| Security rules (8 families × 10 languages) | up to 80 |
| Idiom rules (one per newly-covered language) | 10 |
| Matrix cells to resolve | 98 of 112 |
| Corpus fixture pairs | up to 20, plus 2 promotions |
| `sg test` case pairs | one per implemented rule |

The implement-vs-exempt split across the 98 cells is **not known at the predecessor's plan time**
and is not estimated here: the feasibility probe covered 6 patterns of 80 and says so explicitly.
Sizing this SPEC honestly is itself a task for its plan phase, informed by the matrix the
predecessor delivers.

---

## §B. Scope

**In scope**: security rules for rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, and
swift across the eight families; one idiom rule per those ten languages; resolution of the 98
open matrix cells to a rule id or an evidenced rationale; differential-corpus fixture pairs for
every newly-covered language; extension of `coveredCorpusLanguages`; and promotion of the
`java_uncovered.java` / `rs_uncovered.rs` placeholders.

**Not in scope**: everything the predecessor owns (§D), and `r` / `flutter` rules — ast-grep
0.40.5 has no parser for either.

---

## §C. Requirements

**Deliberately absent at this revision.** Writing them now would repeat the mistake this split
corrects: the predecessor's requirements were authored against defenses that had not been
executed, and two of them proved false under test. This SPEC's requirements should be authored
**after** the contract exists and has been exercised over the 26 existing rules, so they can cite
a working harness rather than an intended one.

What this SPEC's plan phase inherits as fixed contract, not open questions:

| Inherited | From |
|---|---|
| Coverage-matrix schema, axes, and four-class checker | REQ-A16-003 … REQ-A16-007 |
| Every implemented rule carries a `valid`/`invalid` pair | REQ-A16-008, REQ-A16-010 |
| Every security rule carries a `metadata.cwe` anchor, an idiomatic `invalid` case, **and** a citation or probe for the symbol its pattern matches at the head | REQ-A16-011 |
| The two-clause `error` promotion predicate | REQ-A16-012 … REQ-A16-015 |
| Enforcement claims must match measured behaviour | REQ-A16-016 |
| **Every corpus-evidence criterion must assert the run did not skip** | REQ-A16-017 |
| Rule-test assets live outside the distributed template tree | REQ-A16-019 |
| Template-First: the ruleset tree is the source of truth, `make build` is run, and it leaves `internal/template/catalog.yaml` **unchanged** — no embed artifact accompanies a ruleset edit | REQ-A16-018 |
| Neutrality — no SPEC ID, internal date, or commit SHA, and all human-language text in English — scoped to the ruleset directory and the repo-side rule-test root, **not** to `internal/template/templates/**` | REQ-A16-021 |
| The rule-id keying decision (id-alone vs id+language) | Predecessor M1 |

REQ-A16-017 is singled out because it is the one inheritance that closes a hazard this SPEC would
otherwise walk into blind: `coveredCorpusLanguages` is an escape hatch. Adding a language to it
without a denying fixture turns the entire differential test green-by-skip, and `go test` prints
`ok`. Eight of this SPEC's ten languages have no accidental forcing function, so the explicit
no-skip assertion is the only thing standing between them and a green run that observes nothing.

---

## §D. Exclusions

### Out of Scope — the contract itself (predecessor: SPEC-ASTGREP-LANG16-001)

- The `sg test` harness, the coverage-matrix document and checker, the severity reclassification
  of the existing 26 rules, and the `metadata.cwe` anchor convention.
- This SPEC **consumes** those as fixed contract and does not renegotiate them. A defect found in
  the contract is an amendment to the predecessor, not a local override here.

### Out of Scope — `r` and `flutter` rules

- Authoring any rule for `r` or `flutter`. ast-grep 0.40.5 has no parser for either, so a rule for
  them cannot be written, let alone tested. The exclusion is version-scoped and recorded in the
  predecessor's matrix, not a ranking of the language.

### Out of Scope — fixing the scanner's deny capability

- Repairing `astGrepScanner.Scan`'s `CombinedOutput` JSON corruption, named by the skip message at
  `internal/hook/pre_tool_scan_differential_test.go:242` as the cause of the corpus gate's skip
  condition. It is a live defect in the pre-write gate and needs its own card.
- This SPEC works within that constraint rather than around it, which is exactly why REQ-A16-017's
  no-skip assertion is inherited as binding.

### Out of Scope — restructuring the corpus harness

- Redesigning `pre_tool_scan_differential_test.go` beyond adding `scanCorpus` rows and extending
  `coveredCorpusLanguages`. Its differential invariant predates both SPECs.
- Modifying the twelve fixtures committed by `a9eb896ce`, or changing any recorded `wantDeny`
  verdict on a pre-existing row, apart from the two sanctioned `_uncovered` → `_deny_*` promotions.
  The corpus is a recorded baseline: making a gate pass by weakening a recorded verdict defeats
  its purpose.

### Out of Scope — new security families

- Inventing families beyond the eight the predecessor measured. Breadth here is across the
  language axis only; the family axis stays fixed so the matrix keeps a stable shape.

---

## §E. Milestones (provisional)

Sizing is provisional pending this SPEC's own plan phase; the shape is not.

| M | Scope |
|---|---|
| M1 | JVM group — java, kotlin, scala. Up to 24 cells. |
| M2 | .NET and systems — csharp, rust, cpp. Up to 24 cells. Expect a higher exemption rate (cpp has no in-scope web-framework surface). |
| M3 | Dynamic web — ruby, php. Up to 16 cells. Expect the opposite skew: both have mainstream web frameworks, so the web families are more likely implementable, and the severity predicate does real work. |
| M4 | elixir, swift. Up to 16 cells. |
| M5 | Idiom rules — one per newly-covered language, severity `warning`. |
| M6 | Corpus harness: promote `java_uncovered.java` / `rs_uncovered.rs`, extend `coveredCorpusLanguages`, add the clean halves. |
| M7 | Corpus fixture pairs for the remaining eight languages. |
| M8 | Close: matrix at 112/112 resolved, neutrality, `make build` + `catalog.yaml`, `sg scan --config`. |

M1-M4 touch disjoint language cells and are mutually independent. M6 depends specifically on the
milestones delivering java and rust rules. A language whose cells all resolve to EXPLICITLY EMPTY
has no `error`-severity rule and therefore MUST NOT enter `coveredCorpusLanguages` — the gate
would demand a denying fixture that cannot exist, and the only escapes are a contrived rule or a
weakened gate.

---

## §F. Cross-references

- `SPEC-ASTGREP-LANG16-001` — the predecessor whose contract this SPEC consumes; see its §A.7
  (corpus enforcement table) and §A.8 (the two falsified defenses).
- `SPEC-ASTGREP-MULTILANG-001` (completed) — landed the original 26-rule curated baseline.
- `.moai/reports/t228/plan-audit-iter1.md` — the audit whose D8 finding produced this split.
- `internal/hook/pre_tool_scan_differential_test.go` — `scanCorpus`, `coveredCorpusLanguages`, and
  the `t.Skip` at line 242.
- `a9eb896ce` (PR #1637, card t227) — landed the corpus this SPEC extends.
