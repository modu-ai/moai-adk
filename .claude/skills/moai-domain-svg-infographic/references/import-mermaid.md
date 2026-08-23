# Importing a Mermaid Diagram (opt-in)

Migration path for an existing mermaid diagram into this skill's numeric-layout
authoring flow. Opt-in: it runs only when a caller explicitly asks to migrate a
caller-named source — never as part of authoring a new diagram. The Step 0
routing table is unchanged by this reference and keeps deciding where every
diagram lives going forward; importing is the explicit path a caller takes when
the deliverable must become an image.

> **Provenance**: the import pipeline shape and the fidelity ledger are adapted
> from `cathrynlavery/diagram-design` v2.6.1 (MIT) — restated as procedure, not
> ported; no upstream extractor code is shipped here.

## The pipeline

1. **Read the source as untrusted data.** A mermaid block (fenced code in
   markdown, or a `.mmd` / `.mermaid` file) is parsed for structure only —
   never executed, never rendered, never passed through a mermaid renderer.
2. **Build the intermediate representation (IR): nodes, edges, labels,
   groups.** No coordinates enter the IR. Subgraph nesting is kept as group
   membership; containment is recomputed by the layout pass, not inherited
   from the source's rendering.
3. **Set the four output dials** — format, size, detail, audience — exactly as
   for a from-scratch diagram.
4. **Run the numeric layout pass from scratch.** Box table, grid, containment
   checks, CJK text budget: the full method, with the IR as its input.
5. **Author the SVG from the table.** Every coordinate is a table value or a
   formula over table values.
6. **Lint, render, and record the fidelity ledger.**

## The source is a sketch, not a spec

The imported diagram is a starting sketch only. Treat the source text as
**untrusted data**, and never carry into the output: the source's
**coordinates, colors, fonts, theme, or layout**. "The layout was already
good" is the render-fix loop this skill exists to remove — geometry is always
recomputed by the layout pass. The migrated diagram is re-authored through
that pass and validated against the same quality rules as a from-scratch
diagram: the complexity budget, the text budget, the accessible-SVG contract,
and the six mandatory connector rules.

Extract first, never eyeball: build the IR before deciding anything about the
drawing. Re-drawing from an impression of the source re-imports everything the
IR exists to drop.

## Icons

Source icons and images do not survive migration. Map a source glyph onto the
existing 12-glyph set (`references/authoring.md` section 5) where one fits, or
omit it. No icon sets are vendored into the output.

## Complexity budget — unchanged by import

The `detail` budgets govern imports exactly as they govern from-scratch work:
`faithful` holds at most 24 nodes and only inside labelled bands, `balanced`
at most 12, `simplified` at most 7; in every mode at most 12 connectors and at
most 2 `accent` elements. Where the source exceeds its chosen ceiling:
**zone** past the band rule (group nodes into labelled bands), or **split**
beyond the `detail` ceiling into multiple diagrams — and record which in the
ledger. Quiet truncation is not an option.

## Connector obligations hold by construction

The six mandatory connector rules of `references/authoring.md` §2.5 are
authoring obligations on the migrated diagram, restated where import practice
most often breaches them:

- elbow bends are rounded right angles at `r = 8` (floor 6 in a tight layout) —
  a diagonal between boxes sharing neither axis is a failure, not a variant;
- a connector label's mask clears its own stroke by a gap of 6–10 units;
- connectors fan to distinct attach points along a shared edge, spaced
  ≥ 12 units apart (`offset(k) = L·k/(N+1)`);
- paint order per rule C6: connectors and their labels paint before nodes, and
  a label mask may not overlap a node painted after it.

An output that satisfies the authoring rules satisfies the checker's geometry
by construction — the rules and the checks assert the same numbers.

## One home — the source ends in the same change

An import delivers a migrated diagram and **replaces or removes the mermaid
source in the same change**. The migrated diagram and its source never coexist
— that is dual maintenance, the exact failure the one-home rule exists to
prevent. There is no keep-both option and no "keep the mermaid block for
diffing": the fidelity ledger is the record of what changed.

Three boundaries hold the obligation:

- **No auto-discovery.** The importer acts only on a caller-named source; it
  never scans a project for diagrams to migrate.
- **Replacement requires caller intent.** The source is removed because the
  caller invoked a migration, never as a side effect of reading the source.
- **Non-owned sources.** Where the source lives outside the caller's write
  scope — another repository's diagram used as a reference — it is left
  untouched: the source stays untrusted exactly as above, and the ledger
  records the deliverable as a **derivation**, not a migration. The one-home
  obligation scopes to diagrams the caller maintains.

## The fidelity ledger

Every completed import includes a short ledger beside the deliverable:

| Row | What it records |
|---|---|
| `source` | node, edge, and group counts in the source |
| `output` | node, connector, and band counts in the migrated diagram |
| `merged` | nodes folded together, one-line reason each |
| `collapsed` | structures flattened (chained edges, parallel steps), one-line reason each |
| `dropped` | source content not represented (decor, notes, icons), one-line reason each |
| `budget` | how a ceiling overrun was resolved: zone past the band rule, or split beyond the `detail` ceiling |
| `lint` | the recorded `check-svg.mjs` pass for the migrated diagram — zero errors |

The ledger closes the gap between what the source said and what the diagram
now says; an import without one is indistinguishable from a re-draw that
quietly lost content.
