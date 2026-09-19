---
id: SPEC-HEADING-REQ-COLLECT-001
title: "Heading-form REQ definition collection, with body-paragraph text extraction"
version: "0.1.0"
status: completed
created: 2026-09-18
updated: 2026-09-18
author: lane
priority: P2
phase: "v3.2.0 target"
module: "internal/spec"
lifecycle: spec-anchored
tags: "t894, spec-lint, req-collection, modality"
tier: M
---

# SPEC-HEADING-REQ-COLLECT-001

## HISTORY

| Date | Version | Change |
|---|---|---|
| 2026-09-18 | 0.1.0 | Plan-phase authoring. Card t894 — axis 1 of card t801, split out by the lead after the reporting-volume measurement. |
| 2026-09-18 | 0.1.0 | In-card rationale correction (no version bump — prose only) (card t894, lead-directed). §A.3 and §D attributed the `8 → 36` / `28 fewer` `ModalityMalformed` figures to the probe's first-line extractor and stated the shipped extractor's measurement (180 → 181, +1, tree `4ff316bd8`). The "finds more real defects" half of the variant-B rationale is not reproduced by the shipped extractor and was removed; the projection figures are kept, not deleted. Requirements, AC, scope, and the variant decision are unchanged. |

Lineage. Card t801 carried two independent collection gaps. The lead split them:

- Axis 2 — bare numeric-tail shorthand in a sibling `acceptance.md` `maps` list —
  is `SPEC-SIBLING-MAPS-SHORTHAND-001` (`status: completed`). It edits the sibling
  **coverage extractor** union and has a measured **zero** live-corpus delta.
- Axis 1 — this SPEC — edits the **definition collectors**
  (`parseREQsWithProvenance` and its merge) and carries a corpus-wide reporting
  delta. The two share no code path; binding them together would have stalled a
  zero-risk repair behind a corpus-wide one.

Earlier siblings on the same collector: t518 (`SPEC-SPEC-LINT-BLIND-AXES-001`)
added table-form definition collection and the `Widened` / `Source` separation;
t385 widened the list-form separator lexicon.

This SPEC covers ONE defect and ONE decision it forces: heading-form REQ
definitions are not collected at all, and collecting them requires deciding what
becomes `REQEntry.Text`.

## §A. Context — the defect, measured

Measuring instrument: `internal/spec/heading_req_volume_probe_test.go`, a
build-tagged probe (`-tags heading_req_probe`, excluded from the default suite),
run against tree `WT-heading-req-coverage` @ **`9dcbc3dbe`**. Every number in this
section is that probe's output at that SHA. The probe is an instrument, not an
implementation: nothing in it is wired into the live collector.

### A.1 The mechanism

The live collector entry point is `internal/spec/lint.go:688`:

```go
doc.REQs = parseREQsWithProvenance(body)
```

`parseREQsWithProvenance` (`internal/spec/lint_req_widen.go`) is the merge of two
sources, joined by line so the result stays in document order:

1. `parseREQsWide` — the **list** form, whose `reqLineWidePattern` is anchored to a
   markdown list bullet (`^\s*[-*]\s+`).
2. `parseREQsTable` (`internal/spec/lint_req_table.go`, card t518) — the **table** form.

A requirement written as a level-3 heading —

```
### REQ-ADV-001 — Event-driven (When) — advisor rung trigger and re-seed
```

— matches neither anchor, so it is structurally uncollectable today. It does not
fail a rule; it is never visited by one. Six finding codes consume `doc.REQs`
through three loops — `EARSModalityRule` (`lint.go:821`), `REQIDUniquenessRule`
(`lint.go:1036`), `CoverageRule` (`lint.go:1115`) — emitting
`ModalityMalformed`, `ModalityUnjudged`, `LegacyEARSKeyword`, `InvalidREQID`,
`DuplicateREQID`, `CoverageIncomplete`. For a heading-form SPEC, all six are
silent, and that silence is indistinguishable from a pass.

### A.2 The decision the defect forces — what becomes `REQEntry.Text`

Collection alone does not define the repair. Modality judgment is defined over a
requirement **statement**, and a heading line carries a section **title**. The
corpus shape is consistently:

```
### REQ-ADV-001 — Event-driven (When) — advisor rung trigger and re-seed

**When** `/moai loop` … the workflow SHALL instruct the orchestrator to …
```

So the collector must choose between two variants:

- **A** — `REQEntry.Text` is the heading's own trailing text (the title).
- **B** — `REQEntry.Text` is the first non-empty paragraph below the heading (the statement).

**Deciding what becomes `REQEntry.Text` is the other half of collection, not
scope creep.** Collecting a heading while judging its modality on that heading is
an incomplete collection, not a minimal one.

### A.3 Reporting volume, measured before the design decision

Corpus swept: **1634** documents; **218** carry at least one domain-qualified
heading-form REQ. Newly collected REQ entries: **1029** — identical under both
variants, because the variants differ only in `Text`, never in what is collected.
Files gaining at least one entry: **124**.

| Finding code | A: heading title as `Text` | B: first body paragraph as `Text` |
|---|---:|---:|
| `ModalityUnjudged` | 1021 | 144 |
| `CoverageIncomplete` | 471 | 471 |
| `ModalityMalformed` | 8 | 36 |
| `LegacyEARSKeyword` | 0 | 0 |
| `InvalidREQID` | 0 | 0 |
| `DuplicateREQID` | 0 | 0 |
| **TOTAL** | **1500** | **651** |
| top-15 file concentration | 28% | 44% |

Read the two columns against each other. Under A, `ModalityUnjudged` fires on
**1021 of 1029** collected entries — 99.2%. A signal that fires on essentially
every member of its population is a constant, not a signal: it reports that the
collector fed the judge a title, and nothing about the requirement. Under B the
total falls **57%** and `ModalityMalformed` rises **8 → 36**.

> **[HARD] Instrument attribution — corrected at run-phase close; do not restore the
> deleted claim.** The `8 → 36` above (and §D's `28 fewer`) is the PROBE's
> **first-line extractor**: it takes only the first non-empty line below the heading.
> The extractor that SHIPPED takes the whole run of consecutive non-empty lines
> (plan.md §B), and its measured linter delta is `ModalityMalformed` **180 → 181,
> delta +1** — measured against the real linter at tree **`4ff316bd8`**, not projected.
> The two differ because **35 of the probe's 36 are artifacts of the probe's own
> truncation**: they are requirements whose `SHALL` sits on a later line, so cutting at
> the first line hid the token and the judge read the fragment as malformed. Joining
> the whole paragraph restores the `SHALL` and the requirement judges as conforming
> (the M2 evidence phase measured all 35 as having a joined tail that supplies `SHALL`,
> with zero non-prose tails).
>
> The consequence is stated plainly: the **"finds more real defects" half of the
> original reading is NOT reproduced by the extractor that shipped.** The clause "B
> finds more real defects while producing less noise" was false of the shipped
> collector and has been removed rather than softened.
>
> This is **not suppression**. Those 35 are not defects the shipped collector declines
> to report — they are defects the PROBE invented; not reporting them is the accurate
> outcome. B's lower count comes from removing an instrument artifact, never from
> withholding a finding the collector reaches.
>
> **Variant B still ships, and the shipped measurement supports it MORE strongly than
> the projection did**, on the axis that actually decides — the noise ratio:
> **0.99 under A against 0.043 under B** (`ModalityUnjudged` delta **+44** over **1031**
> newly collected entries, same tree). A signal firing on 99% of its population is a
> constant, not a signal. The `8 → 36` and `28 fewer` figures are KEPT, not deleted:
> they were the actual inputs to the plan-phase decision and must stay recoverable
> from this SPEC — they are simply labelled as the probe's projection.

`CoverageIncomplete` is **471 under both** because it keys on the ID, never on the
text. It is the part of the delta that is inherent to collecting at all.

### A.4 Why the volume measurement came first, and what it forbids

The card states it as a [HARD] constraint: measure the reporting volume BEFORE
deciding how to introduce the change, and do not make mass suppression a rational
choice. A rule whose output is unbearable gets turned off or ignored, and then
the repair has killed the gate it was meant to restore.

The measurement is what let variant B be chosen on evidence rather than on taste,
and it is why **no staging mechanism appears anywhere in this SPEC** (§D). All six
codes' newly-reachable findings activate in one change.

### A.5 The gap this SPEC must carry into run-phase

The probe **reimplements** the six findings' conditions rather than invoking the
real linter. Every number in §A.3 therefore rests on that transcription being
faithful, and the corpus itself moves — 218 / 1634 was true of the working tree at
`9dcbc3dbe` and of no other tree.

Both are closed at implementation time, not here: the projected per-code counts
are reconciled against actual `moai spec lint` output at a named tree SHA
(AC-HRC-009), and the corpus census is re-measured rather than quoted from this
document.

## §B. Requirements

### B.1 Collection

- **REQ-HRC-001** — The live REQ definition collector shall collect requirement definitions written as a level-3 markdown heading (`### REQ-…`) into `doc.REQs`.
- **REQ-HRC-002** — The heading-form pattern shall mirror `reqLineWidePattern` in ID shape, optional bold markers, optional parenthesised classifier, and `—`/`:` separator, differing from it ONLY in the anchor. A second, independently-derived ID lexicon shall not be introduced.
- **REQ-HRC-003** — Heading entries shall be merged into `doc.REQs` in document order, and every entry produced today by the list and table sources shall keep its `ID`, `Text`, `Line`, `Widened` and `Source` values unchanged.

### B.2 Text extraction — the decision

- **REQ-HRC-004** — For a heading-form entry, `REQEntry.Text` shall be the first non-empty paragraph following the heading (variant B), not the heading's own trailing text.
- **REQ-HRC-005** — Where no such paragraph exists, `REQEntry.Text` shall fall back to the heading's own trailing text, so no collected entry carries empty `Text`.
- **REQ-HRC-006** — The paragraph search shall stop at the next markdown heading of any level and at end of document. It shall not read a paragraph belonging to a following section.

### B.3 Provenance and severity

- **REQ-HRC-007** — Every heading-form entry shall carry `Widened: true`, so `reqFindingSeverity` resolves its error-severity findings to advisory `warning` exactly as it does for the other newly-reachable shapes.
- **REQ-HRC-008** — A new `REQSource` value shall record the heading shape for ATTRIBUTION only. `Source` shall not reach `reqFindingSeverity` or any other severity decision, preserving the single-severity-axis prohibition already stated on `REQEntry.Source`.

### B.4 Activation

- **REQ-HRC-009** — All newly-reachable findings across the six consuming codes shall become reportable in one change. The implementation shall not introduce a staged rollout, a per-code opt-in, a per-file or per-corpus-size gate, an allowlist, or any other mechanism whose effect is to withhold a finding the collector now reaches.

### B.5 Evidence obligations

- **REQ-HRC-010** — The implementation shall be accompanied by a reconciliation of the probe's projected per-code counts against actual `moai spec lint` output at a named tree SHA, and by a corpus census re-measured at implementation time. A count carried over from §A.3 shall not be presented as an implementation-time measurement.
- **REQ-HRC-011** — The two collector comments that state `doc.REQs` feeds **four** findings (`internal/spec/lint_req_widen.go`, `internal/spec/lint_req_table.go`) shall be corrected to name the six codes that consume it today. Both predate axis 2, which added `ModalityUnjudged` and `LegacyEARSKeyword`.

## §C. Scope

In scope:

- A heading-form definition collector, added as a third source to `parseREQsWithProvenance`'s merge.
- The variant-B text extractor, its bounded paragraph search, and its fallback.
- One new `REQSource` value, attribution-only.
- The two stale comment corrections (REQ-HRC-011).
- The probe reconciliation, the corpus re-measurement, and one mutant probe.

Files expected to change: a new `internal/spec/lint_req_heading.go` plus its test;
`internal/spec/lint_req_widen.go` (the merge, plus the stale comment);
`internal/spec/lint.go` (the new `REQSource` value and its `String()` case);
`internal/spec/lint_req_table.go` (stale comment only — the merge helper is reused,
not rewritten, unless a third input requires it).
`internal/spec/heading_req_volume_probe_test.go` is retained as the reconciliation
instrument.

## §D. Out of Scope

The following are deliberately out of scope for this SPEC.

### Out of Scope — staged rollout and any other suppression device

- A staged rollout, a per-code opt-in flag, a finding-count cap, a per-file allowlist, or a "new SPECs only" cutoff are all rejected (REQ-HRC-009). Staging is itself a suppression device: it converts an activated finding into a deferred one while reporting the work as done. Every affected code is ALREADY advisory — `ModalityUnjudged` and `CoverageIncomplete` set `Advisory: true` at their emission sites regardless of provenance, and the remaining four are demoted for these entries by `reqFindingSeverity` via `Widened` (REQ-HRC-007) — so **nothing gates**, and there is no gate-breakage that staging would be protecting against.
- Variant A (heading title as `Text`) is rejected on the measurement, not on preference: it reports 1500 findings of which 1021 are a constant. The accompanying "28 fewer `ModalityMalformed` defects than variant B" is the PROBE's first-line-extractor projection and is NOT reproduced by the extractor that shipped — whose measured delta is `ModalityMalformed` 180 → 181 (+1) at tree `4ff316bd8`, because 35 of the probe's 36 were artifacts of its own truncation (§A.3, instrument-attribution note). The figure is kept as the plan-phase input it was; the deciding axis is the noise ratio — 0.99 under A against 0.043 under B — on which variant A is rejected more strongly than the projection suggested.

### Out of Scope — corpus remediation

- Rewriting any live SPEC's requirements so the newly-reported findings go away. This SPEC restores a blind collector; it does not clean the corpus. The 651 findings are the corpus's existing state becoming visible.

### Out of Scope — heading levels other than `###`

- `##`, `####`, and deeper heading anchors are not collected. The probe measured the `###` population; no other level has a measured population, and widening an anchor for an unmeasured one is the shape of change §A.4 forbids.

### Out of Scope — the sibling coverage extractor

- The sibling `acceptance.md` `maps`-list numeric-tail expansion is `SPEC-SIBLING-MAPS-SHORTHAND-001` (card t801), already completed. This SPEC does not touch `lint_coverage_sibling*.go` or `internal/spec/ears.go`.

### Out of Scope — severity promotion

- Promoting any of the six codes from advisory to gating is unchanged and untouched. This SPEC makes findings visible; it does not make them block.

## §E. Constraints

| # | Constraint | Enforced by |
|---|---|---|
| C1 | Existing list and table entries are byte-identical in `ID`/`Text`/`Line`/`Widened`/`Source` | REQ-HRC-003, AC-HRC-007 |
| C2 | Heading entries are advisory, so the change adds no gating finding | REQ-HRC-007, AC-HRC-005 |
| C3 | `Source` never reaches a severity decision | REQ-HRC-008, AC-HRC-006 |
| C4 | The paragraph search never crosses into the next section | REQ-HRC-006, AC-HRC-004 |
| C5 | No suppression mechanism exists in the shipped diff | REQ-HRC-009, AC-HRC-008 |
| C6 | Every implementation-time count is re-measured, never quoted from §A.3 | REQ-HRC-010, AC-HRC-009 |
| C7 | The variant actually shipped is B, decided mechanically rather than by inspection | AC-HRC-002, AC-HRC-008's ratio discriminator |

C7 deserves its own row because variant A and variant B collect the **same 1029
entries** and differ only in `Text`. A collection-count check therefore cannot
tell them apart, and neither can a total-findings check alone once the corpus has
moved. The discriminator that survives corpus movement is the **ratio**:
`ModalityUnjudged` newly reported divided by entries newly collected is ~0.99
under A and ~0.14 under B (§A.3). AC-HRC-008 decides on that ratio, and AC-HRC-002
decides the same property directly at the unit level.
