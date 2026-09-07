---
id: SPEC-WEB-ANCHOR-SCOPE-001
title: "web console tests — page-wide first-occurrence anchor classification under the codex mirror"
version: "0.1.0"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.1.5 target"
module: internal/web
lifecycle: spec-anchored
tags: web-console, test-anchoring, codex-mirror, classification, scope-narrowing
era: V3R6
tier: M
related_specs: [SPEC-WEB-CODEX-PANEL-001, SPEC-MCP-CONSOLE-001]
---

# SPEC-WEB-ANCHOR-SCOPE-001 — page-wide anchor scope classification under the codex mirror

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-09-07 | Initial draft from card t527 (t509-derived sweep, lane-12). First deliverable is the CLASSIFICATION TABLE (research.md), not a repair. Card evidence: raw superset 60 sites / 19 files measured on tree `bf779ecf2`; discriminator derived from the mirror's emitted-text inventory + tab order; classification outcome: 1 site historically (c)=true (mcp_console_test.go:115, already repaired by t509), 0 pending (c)=true, 23 not-exposed, 36 exposed-but-not-divergent. |

## A. Background

SPEC-WEB-CODEX-PANEL-001 introduced the codex tab: a read-only mirror that re-renders
codex-scattered settings (audit pins, codex opt-ins, codex MCP tool toggles) as
`<code class="key">FIELD</code>` chips on a panel at tab position 8 — BEFORE the MCP
tab at position 11. Tests that anchor a window with `strings.Index(fullPageBody, chip)`
therefore resolve to the MIRROR's copy for mirrored needles, landing the window on a
read-only row that carries none of the asserted features. Card t509 hit this on
`TestMCPConsoleWriteCapableTextDistinction` ("MCP surface degraded" when nothing had
changed) and repaired it by scope narrowing: `panelHTML(t, renderConsolePage(t), "mcp")`.

The card's generalization: the not-last rule protects only `panelHTML` consumers; every
existing page-wide first-occurrence anchor carries the same risk. The defect family is
"the judged target exists — but it is a different one" (판정 대상이 존재하되 다른 것이다),
siblings t506 (verdict seat with no deciding fact) and t509 (verdict expression's target
absent in that shape); this sweep is the third variant.

**Unknown stated by the card**: nobody knows how many of the superset sites are
"genuinely page-wide" vs "panel-intent scanned page-wide" — they look identical. The
first deliverable of this SPEC is therefore the classification table, not a repair.

## B. Requirements (GEARS)

### REQ-WAS-001 — Classification table is the gating artifact (Ubiquitous)

The SPEC shall carry, in `research.md`, a classification table with exactly one row per
raw-superset site (`strings.Index(` in `internal/web/*_test.go` at the measurement tree),
each row recording (a) the intended anchor scope, (b) the scope the page-wide
first-occurrence anchor actually resolves to under mirror-present rendering, and
(c) the divergence verdict.

### REQ-WAS-002 — No bulk substitution (Unwanted)

The SPEC shall not permit mechanical bulk replacement of the superset sites: a repair
lands only on a row whose (c) verdict is TRUE, and every repair commit's diff must be
traceable row-by-row to (c)=TRUE rows of the table.

### REQ-WAS-003 — Repair shape is scope narrowing only (Ubiquitous)

The run phase shall repair a (c)=TRUE row only by narrowing the anchor's receiver — the
t509 shape `panelHTML(t, renderConsolePage(t), "<panel>")` or an equivalent
already-scoped slice — and shall not touch production, render, or save code
(`internal/web/*.go` non-test files unchanged).

### REQ-WAS-004 — Regression is measured under mirror presence (Ubiquitous)

The run phase shall measure every regression verdict on a tree where the codex mirror is
present (`git merge-base --is-ancestor 1aaf4951f HEAD` — the SPEC-WEB-CODEX-PANEL-001
close commit — must be an ancestor of the measured HEAD). A mirror-absent tree cannot
separate pre-repair from post-repair behavior: all sites are green there.

### REQ-WAS-005 — Discriminator outlives the card (Ubiquitous)

The SPEC shall record the anchor-scope discriminator (research.md §4) as a durable
artifact: any FUTURE test added to `internal/web` that anchors a window with a
page-wide first-occurrence search shall be classified by the discriminator before
landing, and a mirror-duplicated needle on a full-page receiver is a defect, not a
test to weaken.

### REQ-WAS-006 — Zero-repair is a valid table outcome (Event-driven)

When the classification table yields zero (c)=TRUE rows, the run phase shall close with
the table + mutant falsification evidence (AC-WAS-005) and no production or test-code
change; absence of repairs is a finding, not an incomplete sweep.

### REQ-WAS-007 — Discriminator falsifiability (Event-driven)

When the discriminator classifies a site (c)=FALSE, the mutant check of AC-WAS-005
shall be able to demonstrate the discriminator is non-vacuous: the t509 site
(mcp_console_test.go:115) reverted to a body-wide anchor must fail under the
mirror-present tree, proving the failure mode the discriminator guards exists and is
detectable.

## C. Constraints

1. Corpus is `internal/web/*_test.go` (the web test package), measured on the worktree
   HEAD; figures from other trees (lead's 55/18 @ 0b1e27877, lane-1's 60/19) are context
   only, never restated as this SPEC's measurement.
2. The lesson-file entry (`feedback_confident_verdicts_over_empty_sets.md`) is close-time
   lane work — out of this SPEC's scope.
3. Mirrored-field inventory is DERIVED (`codexMirrorFieldNames`, `codexmirror.go`): a new
   codex field is mirrored without code change, which is why REQ-WAS-005's discriminator
   keys on the predicate families, not a hand list.

## D. Acceptance Criteria

See `acceptance.md` (AC-WAS-001 … AC-WAS-007).

## E. Out of Scope

### Out of Scope — production code

- No change to `internal/web/*.go` non-test files: the mirror, its render, and every
  save path are untouched (the t509 repair was test-only; this sweep inherits that shape).

### Out of Scope — non-mirror first-occurrence ambiguity

- Anchors whose first-occurrence semantics are ambiguous for reasons UNRELATED to the
  mirror (e.g. `"<form "` matching the first of several forms) are recorded as residual
  observations in research.md §7 but are not repair targets: this card's defect family is
  mirror-induced divergence, and repairing unobserved ambiguity would be speculative.

### Out of Scope — lesson-file maintenance

- Adding the third-variant entry to `feedback_confident_verdicts_over_empty_sets.md` is
  close-time lane work per the card, not SPEC artifact authoring.

### Out of Scope — mechanical anchor lint infrastructure

- Building an automated lint/CI guard for anchor scope is not required; the discriminator
  is recorded as prose + table (REQ-WAS-005). Mechanizing it is a possible follow-up SPEC
  if drift recurs.
