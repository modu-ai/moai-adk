---
id: SPEC-SIBLING-MAPS-SHORTHAND-001
title: "Sibling acceptance.md maps-list numeric-tail shorthand expansion"
version: "0.1.0"
status: in-progress
created: 2026-09-18
updated: 2026-09-18
author: lane
priority: P2
phase: "v3.2.0 target"
module: "internal/spec"
lifecycle: spec-anchored
tags: "t801, spec-lint, coverage, sibling-acceptance"
tier: M
---

# SPEC-SIBLING-MAPS-SHORTHAND-001

## HISTORY

| Date | Version | Change |
|---|---|---|
| 2026-09-18 | 0.1.0 | Plan-phase authoring. Card t801, narrowed to the sibling `maps`-list axis after the lead split the heading-collection axis out to card t894. |

Lineage. This SPEC is one of two axes card t524's tool-axis residual left behind:

- t518 (SPEC-SPEC-LINT-BLIND-AXES-001, `status: completed`) closed **table-form**
  REQ definitions in `spec.md`.
- t561 closed **table-form** REQ mappings in the sibling `acceptance.md`, and made
  the deliberate decision that `ExtractRequirementMappings` stays IMMUTABLE,
  adding the sibling-only extractor `lint_coverage_sibling_table.go` instead —
  whose `cellREQIDs` already implements same-cell numeric-tail expansion.
- Card **t894** carries the other residual axis — heading-form REQ **definition**
  collection in `spec.md`. It shares no code path with this SPEC: that axis edits
  the definition collectors (`parseREQsWithProvenance` and its merge), this one
  edits the sibling **coverage** extractor union. The two were split because
  t894 moves the live corpus and this one does not (§A.3), so binding them
  together would stall a zero-risk repair behind a corpus-wide one.

This SPEC covers ONE defect: bare numeric-tail shorthand in a sibling
`acceptance.md` `maps` list is not expanded, producing a FALSE
`CoverageIncomplete` warning.

## §A. Context — the defect, measured

Measuring instrument: `go build -o <scratchpad>/t801/moai ./cmd/moai` from tree
HEAD `881aa4bb8` (build exit 0), invoked BY PATH. A PATH-resolved `moai` does not
satisfy any measurement in this SPEC — a stale installed build reports a clean
pass byte-identically to a fresh one
(`verification-claim-integrity.md` §2.2).

### A.1 The mechanism

`CoverageRule` takes its covered-REQ set as the union of the inline AC section in
`spec.md` and the sibling `acceptance.md`. The sibling half is
`siblingAcceptanceCoveredREQIDs` (`internal/spec/lint_coverage_sibling.go`),
which unions two extractors:

1. `ExtractRequirementMappings` (`internal/spec/ears.go`) — the `maps REQ-…` list form.
2. `siblingTableREQIDs` (`internal/spec/lint_coverage_sibling_table.go`, card t561) — the table form.

`ExtractRequirementMappings` locates each mapping with

```
reqSectionPattern = (?i)maps\s+(REQ-[A-Z0-9-]+(?:\s*,\s*REQ-[A-Z0-9-]+)*)
```

Every comma-separated element must carry the `REQ-` prefix. So on

```
- AC-FIXA-001 (maps REQ-FIXA-001, 002): Given one, When two, Then three.
```

the captured section is `REQ-FIXA-001` and `, 002` is dropped silently. The
enumerator never sees the tail, `REQ-FIXA-002` is reported uncovered, and the
author DID map it. The warning is FALSE.

The same shorthand inside a **table cell** is already expanded — `cellREQIDs`
does it. Only the `maps`-list path is blind.

### A.2 Measured today (baseline, tree `881aa4bb8`)

| Fixture | Sibling mapping | Observed | Raw |
|---|---|---|---|
| A (`SPEC-FIXA-001`) | `maps REQ-FIXA-001, 002` and `maps REQ-FIXA-003`; nothing genuinely uncovered | `WARNING CoverageIncomplete … REQ REQ-FIXA-002 is not referenced by any AC`, exit 0 | `.moai/reports/t801/repro/lint-fixA-before.txt` |
| B (`SPEC-FIXB-001`, control) | `maps REQ-FIXB-001, REQ-FIXB-002` — full ids throughout | 0 traceability findings, exit 0 | `.moai/reports/t801/repro/lint-fixB-before.txt` |

Both runs also emit an unrelated `MissingExclusions` warning. That warning is the
**non-empty-sweep witness**: a fixture lint that emitted nothing at all could mean
the rule set never ran, and a zero read from an empty sweep asserts nothing
(`verification-completeness.md` §1.1). Every fixture criterion in this SPEC
therefore asserts the `MissingExclusions` line is present alongside its own
expectation.

Fixtures live at `.moai/reports/t801/repro/fixtures/`.

### A.3 Corpus census — the delta is expected to be ZERO, and that is the point

| Measurement | Value | Command |
|---|---|---|
| `CoverageIncomplete` findings, whole corpus, at `881aa4bb8` | 2018 | path-invoked `moai spec lint`; recorded at `.moai/reports/t801/repro/corpus-before-counts.txt` |
| Corpus totals at `881aa4bb8` | `3 error(s), 3151 warning(s)` | same |
| `maps REQ-…(, NNN)+` occurrences across `spec.md` + `acceptance.md` in `.moai/specs`, **excluding this SPEC's own directory** | **0** | `grep -rlE 'maps[[:space:]]+REQ-[A-Z0-9-]+([[:space:]]*,[[:space:]]*[0-9]+)+' --include=spec.md --include=acceptance.md .moai/specs \| grep -v 'SPEC-SIBLING-MAPS-SHORTHAND-001' \| wc -l` |
| `maps REQ-` occurrences corpus-wide (the population the repair reads), **excluding this SPEC's own directory** | **1099** — see the attribution note below | `grep -rnoE 'maps[[:space:]]+REQ-' --include=spec.md --include=acceptance.md .moai/specs \| grep -v 'SPEC-SIBLING-MAPS-SHORTHAND-001' \| wc -l` |

Attribution note for the fourth row. Unlike the other three, this figure is **not** determined by a
commit SHA. The command sweeps `.moai/specs`, which contains **untracked** SPEC directories, so the
count moves as untracked directories are added or deleted without any commit occurring. It is
therefore attributed to **the working tree of `881aa4bb8` as of 2026-09-18**, not to `881aa4bb8`
alone (`verification-claim-integrity.md` §2).

**The self-exclusion in the command is what makes the figure stable, and it is there for a measured
reason, not for tidiness.** `.moai/specs` contains this SPEC, so a self-including sweep counts this
document's own prose — and every `maps REQ-…` this SPEC writes about the defect moves the number
this SPEC records. That is not a hypothetical: an earlier draft recorded `1128` from a self-including
sweep; the correction to `1120` was falsified in the same round by one added line of this SPEC's own
prose (the tree then read `1121`). Self-inclusion — not untracked-directory churn — was the operative
mechanism both times. Row 3 already excludes this directory for the same reason; row 4 now uses the
same exclusion, so editing this SPEC no longer moves the figure. The residual untracked-set movement
above remains real: a later reader re-running the command should expect a difference from *other*
directories rather than treat one as a regression.

The zero in the third row is load-bearing twice over. It means the repair has a
**zero live-corpus delta**, so its evidence is the fixtures and the mutant — not a
corpus diff. And it means the zero must be **measured after the change**, not
assumed: a repair that silently moved 2018 would be over-reaching, and the only
way to know is to re-run the same command against the same corpus and compare
(AC-SMS-009).

The 1099 figure bounds the blast radius honestly: 1099 `maps` sections are re-read
by the new locator, and every one of them must produce the same id set it produces
today unless it carries a numeric tail.

This SPEC is repaired because fixture A proves the false-warning mechanism is live
and will fire the moment an author writes the shorthand — not because it is
currently firing.

## §B. Requirements

### B.1 The expansion

- **REQ-SMS-001** — The sibling `acceptance.md` coverage extractor shall expand a bare comma-separated numeric tail that follows a full REQ id inside the same `maps` section, yielding the full id formed from the preceding id's prefix and that tail.
- **REQ-SMS-002** — The expansion shall reuse the shape already implemented by `cellREQIDs` (`lint_coverage_sibling_table.go`) — a prefix carried from the nearest preceding full id, applied comma-by-comma within one textual unit — by calling it or by calling a helper extracted from it. A second, independently-written expansion rule shall not exist in the package.
- **REQ-SMS-003** — `ExtractRequirementMappings` and `reqSectionPattern` (`internal/spec/ears.go`) shall not be modified. The repair shall live on the sibling path only, so the inline `spec.md` AC path that also calls that function is untouched.

### B.2 The textual unit — what bounds the expansion

- **REQ-SMS-004** — The textual unit the expansion operates on shall be the capture of a **sibling-local widened `maps`-section locator**: a locator anchored at a `maps` token that admits, after the first full REQ id, comma-separated elements that are either a full REQ id or a bare numeric tail. Enumeration and expansion shall read that capture alone and shall never read the line, the paragraph, or the file that contains it.
- **REQ-SMS-005** — The locator's inter-element separator shall admit horizontal whitespace only, so a `maps` section ends at the end of its line. The expansion shall not cross a line boundary, and shall not cross from one `maps` section into another.
- **REQ-SMS-006** — A REQ id that appears on the same line as a `maps` list but outside the locator's capture shall not be counted as covered. A bare numeric tail with no full REQ id preceding it inside the same capture shall not be expanded, so a prefix is never inferred from a neighbouring section, cell, or line.

### B.3 Evidence obligations

- **REQ-SMS-007** — The implementation shall be accompanied by: the fixture-A false warning measured absent, the fixture-B control measured unchanged, an over-reach control and a line-boundary control measured, one mutant probe measured RED, and a whole-corpus before/after `CoverageIncomplete` comparison re-measured with the same command against the recorded `2018` baseline at `881aa4bb8`. A delta asserted rather than measured does not discharge this requirement.

## §C. Scope

In scope:

- A sibling-only widened `maps`-section locator plus numeric-tail expander, added as a third extractor to `siblingAcceptanceCoveredREQIDs`'s union.
- Four fixtures (A reused; over-reach, line-boundary, and no-preceding-id added) with their tests.
- One mutant probe and the whole-corpus before/after measurement.

Files expected to change: one new `internal/spec/lint_coverage_sibling_maps.go`
plus its test, and the union in `internal/spec/lint_coverage_sibling.go`. Possibly
a helper extraction inside `internal/spec/lint_coverage_sibling_table.go` to satisfy
REQ-SMS-002. `internal/spec/ears.go` is byte-unchanged.

## §D. Out of Scope

The following are deliberately out of scope for this SPEC.

### Out of Scope — heading-form REQ definition collection

- Collecting requirement definitions written as `### REQ-…` headings in `spec.md` is card **t894**, not this SPEC. It edits the definition collectors; this SPEC edits the sibling coverage extractor. They share no code path, and t894 carries a live corpus delta this SPEC does not.

### Out of Scope — inline spec.md shorthand expansion

- The inline `spec.md` AC path (`parseSingleACLine` → `ExtractRequirementMappings`) stays unexpanded. `ExtractRequirementMappings` is immutable per t561's decision (REQ-SMS-003), and the measured census (§A.3) records **0** live occurrences of `maps REQ-…(, NNN)+` across `spec.md` and `acceptance.md`, so the residual has no live instance. It is named here rather than left silent.

### Out of Scope — wrapped `maps` lists spanning two lines

- A `maps` list whose element run continues onto the next physical line is NOT expanded across that boundary (REQ-SMS-005). Measured in the working tree of `881aa4bb8` as of 2026-09-18: **0** occurrences of a `maps REQ-…` list ending in a trailing comma at end-of-line (re-confirmed at 0 by the lane, with the same `SPEC-SIBLING-MAPS-SHORTHAND-001` self-exclusion row 4 uses), against **1099** `maps REQ-` occurrences corpus-wide (§A.3 row 4 and its attribution note). Admitting the line crossing would let a `maps` section ending in a comma absorb the following list item, which is an unbounded widening bought for a population of zero.

### Out of Scope — counting any bare REQ token as coverage

- Widening the sibling predicate to "any `REQ-…` token appearing in `acceptance.md` counts as covered" is rejected. `lint_coverage_sibling.go` already records why: it would silence most genuine findings and would count a REQ named in prose *as excluded* as if it were covered. The mapping form is the coverage declaration.
- Declined forms, named rather than left implicit: a tail with a non-numeric body (`, 00b`), a tail separated by anything other than a comma, a tail preceding the first full id in its section, and a full id written without the `REQ-` prefix.

### Out of Scope — severity and corpus remediation

- Promoting `CoverageIncomplete` from advisory `warning` to `error` is unchanged; this SPEC does not meet the standing promotion conditions recorded on that symbol.
- Authoring or amending the acceptance mappings of any live SPEC. This SPEC repairs a false warning; it does not clean the corpus.

## §E. Constraints

| # | Constraint | Enforced by |
|---|---|---|
| C1 | `internal/spec/ears.go` byte-unchanged | REQ-SMS-003, AC-SMS-006 |
| C2 | Exactly one numeric-tail expansion rule in the package | REQ-SMS-002, **AC-SMS-010** (the positive call-site assertion — the load-bearing enforcer), AC-SMS-005 (two name-blind/name-bound grep tripwires, not exhaustive) |
| C3 | Fixture B (control, full ids) stays at 0 traceability findings | REQ-SMS-007, AC-SMS-007 |
| C4 | A REQ token outside the `maps` capture is never counted as covered | REQ-SMS-006, AC-SMS-003 |
| C5 | Expansion never crosses a line boundary | REQ-SMS-005, AC-SMS-004 |
| C6 | Corpus delta measured against the `2018` / `881aa4bb8` baseline, not assumed | REQ-SMS-007, AC-SMS-009 |
| C7 | Every criterion asserting a **zero** count cites its non-empty-sweep witness | §A.2, AC-SMS-001, AC-SMS-007 |

C7 is scoped to zero-asserting criteria deliberately, and the scope is what makes it true of this
document. Only a zero can be satisfied by a sweep that never ran, which is why AC-SMS-001 and
AC-SMS-007 — the two criteria that assert `zero CoverageIncomplete lines` — each require the
`MissingExclusions` line alongside it (AC-SMS-008 cites the witness as well). A criterion asserting
a **non-zero** count is self-witnessing and needs no separate witness clause: AC-SMS-002 expects 2
lines, AC-SMS-003 expects 2, AC-SMS-004 expects 1, and an empty sweep yields 0 and fails each of
them outright.
