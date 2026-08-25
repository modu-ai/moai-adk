# verdict.md — SPEC-SKILL-GALLERY-BENCH-001 M1 benchmark verdict table

REQ-SGB-004 verdict table: one row per TypePack form (spec.md §D.2), carrying
verdict, archetype used, gate results G1-G3, settled dial values
(REQ-SGB-002), and the REQ-SGB-006 deviation sentence. All paths are relative
to the worktree root and live under `.moai/reports/t272/` (REQ-SGB-007).

Environment (applies to every row): node v22.14.0; browser
`/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`,
`Google Chrome 151.0.7922.174` (found via well-known install location,
`--headless=new`) — disclosed per G3 in every render log. Preflight
(plan.md §C): node ≥ 18 ✓, fixture render probe exit 0 with verified IHDR ✓,
evidence dir absent before run ✓, worktree on `WT-skillstead-gallery` ✓.

## Verdict table

| # | TypePack form | Verdict | Archetype used | G1 | G2 | G3 | Settled dials (format / size / detail / audience) |
|---|---------------|---------|----------------|----|----|----|------------------|
| 1 | `approval-gate` | PRODUCIBLE | A2 left-to-right flow + accent gate node | PASS | PASS (exit=0, 0 errors, 0 warnings) | PASS (exit=0, IHDR 3320x940 = 2x target) | svg+png / **fit** (W 1660; recorded size deviation — 5 preset stages exceed doc-inline 1200) / balanced / mixed |
| 2 | `before-after` | PRODUCIBLE | A3 side-by-side comparison (2 cols × 3 concern rows) | PASS | PASS (exit=0, 0 errors, 0 warnings) | PASS (exit=0, IHDR 2400x980 = 2x target) | svg+png / doc-inline (W 1200) / balanced / mixed |
| 3 | `cards-kpi-grid` | PRODUCIBLE | A1 single-layer card grid (4 stat tiles) | PASS | PASS (exit=0, 0 errors, 0 warnings) | PASS (exit=0, IHDR 2400x796 = 2x target) | svg+png / doc-inline (W 1200) / balanced / mixed |
| 4 | `decision-matrix` | PRODUCIBLE | A3 side-by-side comparison (3 cols × 4 criteria rows, zebra-banded) | PASS | PASS (exit=0, 0 errors, 0 warnings) | PASS (exit=0, IHDR 2400x1128 = 2x target) | svg+png / doc-inline (W 1200) / **faithful, banded** (recorded detail deviation — 19 nodes > balanced ceiling 12) / mixed |
| 5 | `layer-stack` | PRODUCIBLE | A1 architecture stack (direct match) | PASS | PASS (exit=0, 0 errors, 0 warnings) | PASS (exit=0, IHDR 2400x1924 = 2x target) | svg+png / doc-inline (W 1200) / balanced / mixed |
| 6 | `nested-scope` | PRODUCIBLE | A1 nested containers (containment chain) | PASS | PASS (exit=0, 0 errors, 0 warnings) | PASS (exit=0, IHDR 2400x1660 = 2x target) | svg+png / doc-inline (W 1200) / balanced / mixed |
| 7 | `process-flow` | PRODUCIBLE | A2 left-to-right flow (direct match) | PASS | PASS (exit=0, 0 errors, 0 warnings) | PASS (exit=0, IHDR 3320x840 = 2x target) | svg+png / **fit** (W 1660; recorded size deviation as form 1) / balanced / mixed |
| 8 | `roadmap-timeline` | PRODUCIBLE | A2 stage cards + derived time axis | PASS | PASS (exit=0, 0 errors, 0 warnings) | PASS (exit=0, IHDR 2400x900 = 2x target) | svg+png / doc-inline (W 1200) / balanced / mixed |
| 9 | `topology-component` | PRODUCIBLE | A1 architecture stack + bottom two-lane fan-out | PASS | PASS (exit=0, 0 errors, 0 warnings) | PASS (exit=0, IHDR 2400x1924 = 2x target) | svg+png / doc-inline (W 1200) / balanced / mixed |

Totals: **9 PRODUCIBLE / 0 PARTIAL / 0 NOT-PRODUCIBLE.**

Warning triage (G2 contract): every lint log reports `0 errors, 0 warnings`,
so there are no warnings to triage — recorded here so the empty triage is an
observed result, not an omission.

## Evidence paths (REQ-SGB-003 / REQ-SGB-007)

Every log follows the spec.md §D.2 log-format contract: command line +
verbatim tool output + explicit `exit=N` line.

| Form | G1 artifact | G3 PNG | G2 lint log | G3 render log |
|------|-------------|--------|-------------|---------------|
| approval-gate | .moai/reports/t272/artifacts/approval-gate.svg | .moai/reports/t272/artifacts/approval-gate.png | .moai/reports/t272/logs/approval-gate-lint.txt | .moai/reports/t272/logs/approval-gate-render.txt |
| before-after | .moai/reports/t272/artifacts/before-after.svg | .moai/reports/t272/artifacts/before-after.png | .moai/reports/t272/logs/before-after-lint.txt | .moai/reports/t272/logs/before-after-render.txt |
| cards-kpi-grid | .moai/reports/t272/artifacts/cards-kpi-grid.svg | .moai/reports/t272/artifacts/cards-kpi-grid.png | .moai/reports/t272/logs/cards-kpi-grid-lint.txt | .moai/reports/t272/logs/cards-kpi-grid-render.txt |
| decision-matrix | .moai/reports/t272/artifacts/decision-matrix.svg | .moai/reports/t272/artifacts/decision-matrix.png | .moai/reports/t272/logs/decision-matrix-lint.txt | .moai/reports/t272/logs/decision-matrix-render.txt |
| layer-stack | .moai/reports/t272/artifacts/layer-stack.svg | .moai/reports/t272/artifacts/layer-stack.png | .moai/reports/t272/logs/layer-stack-lint.txt | .moai/reports/t272/logs/layer-stack-render.txt |
| nested-scope | .moai/reports/t272/artifacts/nested-scope.svg | .moai/reports/t272/artifacts/nested-scope.png | .moai/reports/t272/logs/nested-scope-lint.txt | .moai/reports/t272/logs/nested-scope-render.txt |
| process-flow | .moai/reports/t272/artifacts/process-flow.svg | .moai/reports/t272/artifacts/process-flow.png | .moai/reports/t272/logs/process-flow-lint.txt | .moai/reports/t272/logs/process-flow-render.txt |
| roadmap-timeline | .moai/reports/t272/artifacts/roadmap-timeline.svg | .moai/reports/t272/artifacts/roadmap-timeline.png | .moai/reports/t272/logs/roadmap-timeline-lint.txt | .moai/reports/t272/logs/roadmap-timeline-render.txt |
| topology-component | .moai/reports/t272/artifacts/topology-component.svg | .moai/reports/t272/artifacts/topology-component.png | .moai/reports/t272/logs/topology-component-lint.txt | .moai/reports/t272/logs/topology-component-render.txt |

Numeric layout passes (box tables, grid derivations, containment checks) for
all nine forms: `.moai/reports/t272/layout-notes.md`.

## Deviation naming (REQ-SGB-006)

One sentence per produced artifact that is equivalent-but-not-identical to
the TypePack form (operator decision #1: information-structure equivalence;
visual mimicry not required):

1. **approval-gate** — the gate is rendered as an accent-focused stage card
   (★ GATE eyebrow, accent border) with a dashed "changes requested" return
   loop, not as a diamond decision glyph.
2. **before-after** — the two states are rendered as an A3 comparison grid
   over shared concern rows, not as two mirrored freeform picture panels.
3. **cards-kpi-grid** — KPI tiles are rendered as a single-layer A1 card grid
   with one accent focal card and serif numerals, not as dashboard widgets.
4. **decision-matrix** — the matrix is rendered as a zebra-banded A3 grid at
   the faithful detail dial (recorded dial deviation), not at balanced.
5. **layer-stack** — none (direct archetype match; band-and-card stack).
6. **nested-scope** — nesting is rendered as inset rectangles per the
   method's containment inequalities, not as concentric rings (the equivalent
   form spec.md §A anticipated).
7. **process-flow** — the lifecycle is rendered as numbered A2 stages with
   the sync gate marked by accent + ★ SYNC GATE eyebrow, no dedicated gate
   glyph.
8. **roadmap-timeline** — phases are rendered as stage cards bound to a
   derived time axis by milestone ticks, not as contiguous proportional
   time-band segments (the pinned brief carries no dates or durations).
9. **topology-component** — hub-and-spoke connections are rendered as a
   vertical component stack with a two-lane fan-out at the backend layer.

## Failure taxonomy (REQ-SGB-005)

No entries — every form's gates were observed passing without dropping any
structural element of its brief, so no `preset-gap` or `structural-limit`
classification applies. Acceptance.md §D.2 pre-authorizes this outcome
("All nine forms come out PRODUCIBLE → … the taxonomy simply has no entries
(valid outcome)"). The spec.md §B known-issue hypothesis (before-after,
cards-kpi-grid, roadmap-timeline, nested-scope resisting the archetypes) was
tested and did not hold under the operator's information-structure
equivalence criterion; see Follow-up observations for what a stricter
visual-equivalence criterion would change.

## Follow-up observations (non-binding, NOT taxonomy entries)

Recorded for follow-up cards only; none of these degraded a gate or dropped
a structural element, so none affects a verdict:

- A dedicated **timeline-band preset** (contiguous proportional segments
  with dated boundaries) would produce a visually closer roadmap-timeline
  than the A2-stages-plus-axis equivalent form authored here.
- A **two-panel composite preset** (mirrored panels with a shared concern
  rail) would produce a visually closer before-after than the A3 grid.
- A **stat-tile preset** (numeral-forward card with optional icon) would
  formalize what cards-kpi-grid improvised from A1 single-layer cards.
- The lint's connector-geometry tier only recognizes connectors whose
  `<path>` carries `fill="none"` itself, and reads a `<rect>` followed
  immediately by a `<text>` as a label mask — authoring idioms documented in
  `layout-notes.md` Appendix. A follow-up could surface these idioms in the
  skill's authoring reference (out of scope here per REQ-SGB-008).

## Dial-deviation register (REQ-SGB-002)

| Form | Dial | Pinned | As-run | Reason |
|------|------|--------|--------|--------|
| approval-gate | size | doc-inline (1200) | fit (W 1660) | 5 preset-width stages (220 + 72 gutter) exceed 1200; shrinking stages would breach "never shrink the boxes" |
| process-flow | size | doc-inline (1200) | fit (W 1660) | same 5-stage geometry as approval-gate |
| decision-matrix | detail | balanced (≤ 12 nodes) | faithful, banded (≤ 24, labelled bands) | 3 heads + 4 row labels + 12 cells = 19 nodes; the brief's 12-cell matrix exceeds the balanced ceiling without dropping structure |

All other forms ran the pinned defaults unchanged. Every settled value
appears in the verdict table's settled-dials column so a later regeneration
reproduces the artifact.

## Skill immutability checkpoint (REQ-SGB-008 / AC-007, M1)

Command: `git diff --stat origin/main -- .claude/skills/moai-domain-svg-infographic/ internal/template/templates/.claude/skills/moai-domain-svg-infographic/`
Result: empty output (no diff) at the M1 evidence checkpoint on
`WT-skillstead-gallery`.
