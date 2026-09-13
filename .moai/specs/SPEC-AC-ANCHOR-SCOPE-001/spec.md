---
id: SPEC-AC-ANCHOR-SCOPE-001
title: "findACSectionStart anchor scope repair — narrow-miss and empty-anchor axes"
version: "0.1.0"
status: draft
created: 2026-09-14
updated: 2026-09-14
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/spec"
lifecycle: spec-anchored
tags: "spec-lint, acceptance-criteria, parser-anchor, corpus-verification, t747"
tier: M
related_specs: [SPEC-AC-COLLECTOR-ANCHOR-001]
---

# SPEC-AC-ANCHOR-SCOPE-001 — findACSectionStart anchor scope repair

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-14 | manager-spec | Initial plan-phase draft (card t747, branch `WT-ac-anchor-scope`) |

## §A Problem Statement

`findACSectionStart` (`internal/spec/parser.go:123`) selects the inline-AC anchor region of a
`spec.md`. Measured on the live corpus (860 spec.md files, denominator frozen in
`.moai/reports/t747/probe/filelist.txt` — evidence doc
`.moai/reports/t747/anchor-scope-measurement.md`, probe source
`.moai/reports/t747/probe/anchor_scope_probe_test.go`, re-derivation
`T747_PROBE_OUT=<abs> go test ./internal/spec/ -run TestT747AnchorScope -v -count=1`):

| Metric | Value |
|---|---|
| Denominator (spec.md read) | 860 |
| Files with ≥1 AC declaration (whole file) | 152 (1405 declarations) |
| Files anchored by findACSectionStart | 462 |
| Declarations in-section under the current anchor | 1240 |
| Declarations out-section | 165 |
| **Narrow axis** — declarations exist, NO anchor | **14 files** (`probe/anchor-missed.txt`) |
| **Loose axis** — anchored but 0 in-section while declarations exist elsewhere | **9 files** (`probe/empty-anchor.txt`), all 9 via the empty-first-section fallback |

Root causes (both measured, not inferred):

1. **Narrow axis (anchor absence)** — the anchor requires a heading that matches
   `acSectionVocabulary` (`parser.go:69`) at level ≥2. The 14 files carry real AC-declaration
   lists under headings the fixed vocabulary does not name (incl. SPEC-AC-COLLECTOR-ANCHOR-001
   itself with 4 declarations, SPEC-CC297-001 with 19, SPEC-STATUS-AUTO-001 with 25), so the
   parser never sees their AC sections.
2. **Loose axis (empty-anchor preemption)** — when several vocabulary headings exist and ALL
   their sections are empty of AC lines, the fallback (`parser.go:137`, `return first`) anchors
   the FIRST, empty one. The real declarations (1–28 per file) sit under later headings, so the
   file is "anchored" at nothing.

Comparability baseline: t528 (`.moai/reports/t528/probe/merged-tree-remeasure.md` — denominator
815, in-section 1167, frozen-anchor accepted 216), probe committed as
`internal/spec/zz_t528_anchor_probe_test.go` with the frozen-baseline two-column methodology.

## §B Requirements (GEARS)

Mixed-document bound (guard rationale, `internal/spec/lint_coverage_sibling.go:29-32`, preserved
verbatim as a design invariant): spec.md is a MIXED document in which prose must not be read as
AC. The inline path is deliberately scoped twice — the anchor must name the acceptance section
AND the line must carry the `AC-…:` colon form. Of the 165 out-section declarations, the
in=10/out=1 shape (e.g. SPEC-CLAUDEMD-DIET-V2-001, SPEC-DB-SYNC-HARDEN-001) is likely
correctly-excluded prose. This SPEC repairs the 23 defect files WITHOUT absorbing that prose
layer (REQ-ACAS-004).

- **REQ-ACAS-001 (narrow axis)** — **When** a `spec.md` document carries at least one
  AC-declaration line (the discriminator-B colon form) but no heading matched by the
  acceptance-section vocabulary, **While** the document's non-AC prose remains outside the
  anchor region, the inline AC parser shall select an anchor that includes at least one of the
  document's AC-declaration lines rather than no anchor at all.

- **REQ-ACAS-002 (loose axis)** — **When** every acceptance-vocabulary-matching section in a
  `spec.md` holds zero AC lines while AC-declaration lines exist under other headings of the
  same document, the inline AC parser shall not anchor an empty vocabulary section; it shall
  anchor the region containing the declarations.

- **REQ-ACAS-003 (no-regression control)** — The inline AC parser shall produce parse results
  identical to the pre-repair baseline (byte-identical in-section declaration sets) for every
  declaration-bearing `spec.md` outside the 23-file defect set (`probe/anchor-missed.txt` ∪
  `probe/empty-anchor.txt`), except entries carried on an explicit justified-delta list recorded
  in run-phase evidence.

- **REQ-ACAS-004 (prose-bound)** — The inline AC parser shall not newly move any declaration
  line into the in-section count unless that line lies inside a region newly covered by a
  defect-file repair or an entry on the justified-delta list; the out-section prose layer
  (in=10/out=1 shape) shall remain excluded.

- **REQ-ACAS-005 (verification methodology)** — **When** the anchor probe runs, it shall derive
  all corpus counts in a single run against the frozen denominator
  (`.moai/reports/t747/probe/filelist.txt`) and report the narrow-miss count and the
  empty-anchor count as the two headline before/after columns.

- **REQ-ACAS-006 (frozen baseline discipline)** — The probe promoted into the tree shall follow
  the t528 two-column methodology (`internal/spec/zz_t528_anchor_probe_test.go` precedent):
  before-images frozen and never overwritten, `declRe` (discriminator B) byte-frozen, corpus
  numbers re-derived in-run and never carried across two runs or two trees.

## §C Constraints

- Only `internal/spec` parser behavior changes; no public CLI surface change
  (`moai spec lint` / `spec list` outputs change only insofar as AC parsing feeds them).
- The double scoping (anchor names the section + line carries the colon form) is a design
  invariant; the repair may widen WHAT names the section, not remove the section-naming gate
  entirely without stating its bound against the prose layer (REQ-ACAS-004).
- t528 pinned evidence (`.moai/reports/t528/**`) is read-only.

## §D Out of Scope

### Out of Scope — acceptance.md sibling path

- `lint_coverage_sibling.go` already reads `acceptance.md` via `ExtractRequirementMappings` over
  the full text, by deliberate design (option ii of its owning SPEC). This SPEC touches the
  spec.md inline anchor only; the sibling path's behavior, tests, and rationale stay untouched.

### Out of Scope — out-section prose-layer reclassification

- The ~165 out-section declarations outside the 23 defect files are NOT reclassified by this
  SPEC. Sample-reading them (which are prose mentions vs missed anchors) is run-phase
  verification input, not a deliverable.

### Out of Scope — extractACLines anchor-level break semantics

- The secondary difference between the probe's in-section scan shape (break at any `##` line,
  t528-comparable) and `extractACLines`' anchor-level break (reads `###` sub-sections) is
  recorded as a known divergence and resolved by a separate card, if ever.

### Out of Scope — t528 pinned evidence and probe fixtures

- `.moai/reports/t528/**` is pinned evidence: read-only, never edited or re-measured.
- The t747 frozen artifacts under `.moai/reports/t747/probe/` are the before-column record;
  the committed in-tree probe re-derives numbers in-run rather than rewriting them.

## §E Success Criteria

- Narrow-miss: 14 → 0 repaired-or-justified (per-file disposition, AC-747-001).
- Empty-anchor: 9 → 0 repaired-or-justified (per-file disposition, AC-747-002).
- No-regression control: decl-bearing files outside the defect set byte-identical, or explicit
  justified delta (AC-747-003).
- Prose bound: every newly in-section declaration attributable to a defect repair or justified
  delta (AC-747-004).

Dependency note: the no-regression control set is "declaration-bearing files not in the defect
lists"; its exact membership and count are derived in-run by the committed probe (the plan
figures ~129–142 differ by overlap accounting; the in-run derivation is authoritative).
