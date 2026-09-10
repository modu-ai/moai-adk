# SPEC-ASTGREP-BREADTH-001 — Implementation Plan

> **SKELETON.** Milestone shape and inherited constraints only. Detailed per-milestone work
> breakdown, stop conditions, and AC mapping are authored in this SPEC's own plan phase, once
> `SPEC-ASTGREP-LANG16-001` has landed and its harness has been exercised.

## §A. Entry condition

This SPEC does not start until the predecessor's contract is merged and green:

| Gate | Check |
|---|---|
| Harness exists and works | `sg test` passes over the 26 existing rules, and both mutants (`[Missing]`, `[Noisy]`) were observed failing |
| Matrix and checker exist | 112 keys, four failure classes each proven to fire, wired into the Go test suite |
| Severity settled | All 26 existing rules carry a severity justified by their own test cases; every security rule carries `metadata.cwe` |
| Rule-id keying answered | `design.md` records id-alone vs id+language, and any rename has already landed |
| Asset placement settled | Rule-test fixtures live outside the distributed template tree |

Starting before these are green would repeat the failure the split corrects: writing rules against
a contract that has not been executed.

## §B. Milestones

Per `spec.md` §E. Slicing is by language group so each milestone is independently verifiable by
`sg test` plus the matrix checker, and so M1-M4 can run in any order or in parallel lanes.

```
M1 (jvm)      ─┐
M2 (.net/sys) ─┤
M3 (ruby/php) ─┼─> M5 (idioms) ─┐
M4 (elixir/swift) ─┘             ├─> M7 (corpus pairs) ─> M8 (close)
        └─ java+rust land ─> M6 (corpus promotion) ─┘
```

## §C. Constraints inherited as contract

From `SPEC-ASTGREP-LANG16-001` — not renegotiable here:

- Every implemented rule carries a `valid` and an `invalid` case, and `sg test` passes.
- Every security rule carries `metadata.cwe`, with an `invalid` case instantiating that weakness
  class in idiomatic code for the language.
- `error` severity requires both clauses of the promotion predicate; family membership alone is
  not sufficient.
- **Every corpus-evidence criterion asserts the run did not skip** — output contains `--- PASS`
  and neither `--- SKIP` nor `corpus rejected:`.
- Rule-test assets stay outside the distributed template tree.
- `make build` and `internal/template/catalog.yaml` ride the same commit as any template edit.
- Neutrality binds all human-language text in the ruleset, fixtures included.

## §D. Known hazards carried in

- **The corpus escape hatch.** Adding a language to `coveredCorpusLanguages` without a denying
  fixture skips the entire differential test and prints `ok`. Eight of the ten languages have no
  accidental forcing function; the inherited no-skip assertion is the only guard.
- **Contrived rules.** The `metadata.cwe` anchor raises the floor but is necessary-not-sufficient.
  Under 80 rules across languages the implementer may not work in daily, this stays the
  highest-probability failure mode, and per-milestone review should target it directly.
- **Exempt-only languages.** A language whose cells all resolve to EXPLICITLY EMPTY must not enter
  `coveredCorpusLanguages`.
- **Unknown implement-vs-exempt split.** Not estimated. The plan phase sizes it from the delivered
  matrix rather than from the 6-of-80 feasibility probe.

## §E. Cross-references

- `spec.md` — scope, inherited contract table, exclusions
- `SPEC-ASTGREP-LANG16-001` — the predecessor; §A.7 and §A.8 in particular
- `.moai/reports/t228/plan-audit-iter1.md` — finding D8, which produced this split
