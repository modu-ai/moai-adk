# SPEC-SVG-GEOMETRY-CHECKS-001 — Implementation Plan

> Ordered by decision-reversibility: the decisions most likely to change on review come first
> (§A the inference model, §B the diagnostic semantics), the mechanical steps last (§F M3).

---

## §A Context

Card t166. Extend `.claude/skills/moai-domain-svg-infographic/scripts/check-svg.mjs` (609 lines,
Node 18 stdlib, text tokenizer + light element tree) with three geometric checks derived from the
already-settled connector rules C2 / C4 / C6 of `references/authoring.md` §2.5, plus a bipolar
self-test runner. Single measurement path: everything lands inside the existing `lint()`.

---

## §B The inference model (the decision most likely to be challenged)

No authoring convention is added, so the checker has to infer what a connector and a connector
label are from geometry and document shape alone. Three inference steps, each with an explicit
bail-out:

**B0 — Rendered geometry only (applies before B1, B2, and B3).** Any element inside a non-rendered
subtree — `<defs>`, `<marker>`, `<symbol>`, `<clipPath>`, `<mask>`, `<pattern>` — is excluded from
every candidate set. These are definitions, not painted geometry, and document order does not
describe their paint order either, which makes them wrong inputs to the `SVG071`
later-in-document-order test as well as to `SVG072`.

**B1 — Connectors.** A `<path>` with a `d` attribute, no `transform`, outside every non-rendered
subtree, whose `d` parses cleanly under the reader of §C, **and whose `fill` is `none`**. A filled
path is a shape, not a connector: the documented idiom carries `fill="none"` plus a stroke.
`<line>` is deliberately **not** admitted: it carries no `d` for the reader and normally no `fill`
either, so the guard above has no defined outcome on it, and the one `<line>` in the skill tree
(`sketch.md` §3's hatch pattern) sits inside a `<pattern>` and is excluded by B0 regardless.
Admitting it would need its own requirement, AC, and fixture for no observed case.

**Arrival vs departure.** Each admitted connector yields a polyline; its **arrival point** is the
endpoint at which a marker resolves — final point under `marker-end`, first under `marker-start`,
**both** under both markers, none when it carries neither. Only arrival points feed `SVG072` (spec.md §B D7). Both
endpoints still count as attach targets for B3's mask-candidate exclusion, since a node whose edge a
connector merely departs from is no more a label mask than one it arrives at.

**B2 — Boxes.** Every untransformed `<rect>` with numeric `x`/`y`/`width`/`height`, outside every
non-rendered subtree. Its four edges are the attach surfaces for `SVG072`.

**B3 — Label masks.** A `<rect>` whose immediately-following sibling is a `<text>`, and which is
not the attach target of any connector endpoint. This is D2 of spec.md §B. It is deliberately
narrow: a mask wrapped in a `<g>` away from its text, or with an intervening element, is not a
candidate, and the check stays silent (D3 — false negatives accepted, false positives are defects).

**Association of a mask to "its own" stroke (needed by `SVG070` / `SVG073`).** The mask's own
connector is the nearest connector polyline by perpendicular distance from the mask rect, among
connectors within **16 units** of it. When none qualifies, the mask is not checked.

The window is bounded by the band it feeds, and that is the whole point of the number. `SVG073`
fires above 10 units, so a window wider than 10 + tolerance turns every association in the gap into
an automatic warning: at the 24-unit window this plan first carried, every non-label
`<rect>`+`<text>` pair — legend swatch, badge chip, KPI tile, axis label — sitting 11-24 units from
a connector would have been warned as "label mask too far from its stroke", on clean geometry, and
would have failed under `--strict`. The tolerance is 6, C2's own lower bound and so the smallest
distance the rule treats as meaningful, giving 10 + 6 = 16. Beyond 16 the mask associates to
nothing and is silent — an accepted false negative in the posture of D3, and the same posture that
covers the A2 label placement recorded in `spec.md` §A.

**The residual, stated rather than claimed away.** Narrowing 24 → 16 does not eliminate the channel:
a non-label `<rect>`+`<text>` pair sitting 11-16 units from a connector still associates and still
warns. The exposure shrinks from a 14-unit band to a 6-unit one, and D2's attach-target clause
already removes the common case (a chip on a node edge), but the remainder is real and is accepted
here rather than asserted away. Removing it entirely needs a positive discriminator for "is this a
connector label", which D1 rules out for this SPEC. The two chip fixtures are therefore placed at
20 units — outside the window, where the rule's own boundary is what makes them clean — and the
11-16 case is recorded as an `acceptance.md` § D.9 row so the residual is visible rather than
implied.

### B4 — Leak re-check against a `<defs>`-bearing document

B0 exists because the model without it produced a concrete false positive on the pre-existing
`$F/a11y-present.svg`: its arrowhead `<path d="M 0 0 L 10 4 L 0 8 z" fill="#6B6359"/>` inside
`<defs><marker>` parses cleanly under §C, its first and last points are both `(0,0)` after `z`, and
`(0,0)` sits exactly on the top and left edges of the `<rect x="0" y="0" width="400" height="200">`
background — two endpoints binding to one edge at separation 0, which `SVG072` reported under the
model as it then stood.

Three independent guards now stop it, and the honest statement is that the last of them alone would
have sufficed for this particular shape: B0 removes it from the connector set; B1's `fill="none"`
clause removes it again; and the arrival-only ruling (`spec.md` §B D7) means this path — closed, and
carrying no marker of its own — contributes no arrival point at all, so no pair exists to compare.
That third guard is narrower than it first appears: a closed path carrying *both* markers has two
coincident arrivals and is reported, which is why `c4-closed-both-markers` is a must-flag fixture.
B0 is
still load-bearing — it is what keeps `<defs>` geometry out of the box set and out of `SVG071`'s
document-order scan, neither of which the arrival ruling touches — and it is the **only** guard for
the `fill="none"` sketch-mode arrowhead of `sketch.md` §3, which B1 does not catch (REQ-SGC-009).

Every inference step was then re-walked over that same document, and the result recorded here so a
reviewer can check the walk rather than trust it:

| Step | Element in `a11y-present.svg` | Outcome |
|---|---|---|
| B1 | `<path d="M 0 0 L 10 4 L 0 8 z" fill="…">` in `<defs><marker>` | excluded twice — B0 subtree, and filled |
| B1 | `<path d="M 160 100 H 230" fill="none" stroke="…">` | admitted; the one real connector |
| B2 | background `<rect x=0 y=0 400x200>` | admitted as a box; no connector endpoint binds it once B0 holds |
| B2 | `<marker>` geometry attributes | excluded by B0 |
| B3 | `<rect x=40 y=70 120x60>` followed by `<text>Client</text>` | **not** a mask candidate — the connector's start point `(160,100)` sits on its right edge, so it is an attach target |
| B3 | `<rect x=240 y=70 120x60>` followed by `<text>API</text>` | **not** a mask candidate — the connector's end point `(230,100)` is 10 units off its left edge, inside the `markerLen` tolerance, so it is an attach target |
| B3 | background `<rect>` | not a mask candidate — its next sibling is the connector `<path>`, not a `<text>` |
| SVG071 | no mask candidate exists in this document | vacuously silent |
| SVG072 | one connector, endpoints on two different boxes and edges | no pair shares an edge |

The two node rects are the load-bearing case for D2's "not an attach target" clause: without it,
every labelled node in every diagram this skill emits would be a mask candidate.

---

## §C The path-geometry reader

A small `d`-attribute scanner returning `{ points: [[x,y], …] } | null`:

| Command | Handling |
|---|---|
| `M` / `m` | move; starts a subpath, emits the point |
| `L` / `l` | line; emits the point |
| `H` / `h` | horizontal line; emits `[x', yPrev]` |
| `V` / `v` | vertical line; emits `[xPrev, y']` |
| `Q` / `q` | rounded corner of §2.2 — **subdivide**: emit `B(t)` at t = 0.25, 0.5, 0.75 plus the endpoint; the control point is NOT emitted |
| `A` / `a` | crossing hop of §2.3 — pass-through: emit the endpoint only |
| `Z` / `z` | close; emit the subpath start |
| anything else (`C`, `S`, `T`, arcs with unparsable flag runs, malformed numbers) | return `null` |

`null` means "skip this connector silently" (REQ-SGC-011). Both absolute and relative forms are
handled; implicit repeated coordinate pairs after a command letter are handled the way SVG defines
them, and an odd trailing coordinate count returns `null`.

**Why `Q` is subdivided rather than reduced to its control point.** §2.2 replaces each elbow corner
with `Q midX A.cy midX (A.cy + sign*r)`, whose control point is the *un-rounded* corner — a point
the stroke never reaches. Emitting it puts the polyline outside the true stroke on the convex side
of the bend, by up to `0.25·r·√2 ≈ 2.83` units at the documented `r = 8`. The sign is what makes
this a defect rather than a rounding nuisance: on the convex side — open canvas, and a plausible
label position — the measured clearance is *shorter* than the truth, so a mask authored at C2's
documented minimum of 6 measures ≈3.2 and is reported `SVG070` on a compliant diagram. Three
interior samples bound the residual error to well under a tenth of a unit at `r = 8`, cost four
multiplications per corner, and need no library. The `c2-mask-outer-elbow-corner-at-6` fixture is
declared clean and is what keeps this class guarded.

---

## §D Diagnostic semantics

**`SVG070` / `SVG073` (C2).** Compute the minimum distance from the mask rect to the associated
connector polyline — associated meaning within the 16-unit window of §B, never wider. Overlap or
touching counts as distance 0. `d < 6` → `SVG070` (error); `6 <= d <= 10` → silent;
`10 < d <= 16` → `SVG073` (warning); no association → silent. Position: the mask rect's offset.

**`SVG071` (C6, deliberately wider — §B D8).** For each label mask, scan `<rect>` elements appearing
later in document order, **excluding every other label-mask candidate** — admitting masks would
report two overlapping connector labels, which C6 does not make a failure (REQ-SGC-003). Every other
later rect stays in scope, decorative dividers and rules included: separating those from nodes
geometrically would need the marking convention D1 declines, and the fault the check catches — a
label rendering as a fragment on a later rect's border — is real whether or not C6 names that rect.
Classify each remaining pair: **disjoint** (no shared area, or edge-touching only) → silent;
**contained** (mask fully inside the later rect) → silent, that is the badge chip; **partial**
(non-empty intersection that is neither) → `SVG071`. Containment of the later rect inside the mask
is also partial-in-effect and is reported, because the mask is then clipped on all sides by a later
paint.

**`SVG072` (C4).** Gather every connector's **arrival** point only — the final polyline point under
`marker-end`, the first under `marker-start`, both under both markers, none when the connector
carries neither (§B, spec.md §B D7). Departure points are not gathered, which is what keeps the §2.3 fan trunk and the A4 tree
stem clean. Bind each arrival to a box edge: candidate edges are those whose perpendicular distance
to it is `<= 10` (markerLen tolerance, REQ-SGC-012) and whose span contains its projection
(REQ-SGC-012); take the single nearest, ties broken by edge order top/right/bottom/left for
determinism. Group by `(box, edge)`, sort by projected offset, and compare consecutive pairs
against the floor: 12 when the edge length is `>= 120`, 8 when it is `< 120` (REQ-SGC-005).
Separation 0 is the coincident case and reports under the same code (REQ-SGC-004). Position: the
later of the two offending path elements.

**Transform aggregate — `SVG074` (REQ-SGC-006).** `hasTransform` walks ancestors
(`check-svg.mjs:297-303`), so a single `<g transform>` wrapping a diagram excludes everything in it.
The count reported is therefore the **transitively excluded** population, not the number of elements
literally carrying the attribute, and the message states it against the total candidate count — "N
of M candidate elements skipped (transform)" — so the wrapper case discloses that nothing was
verified instead of reporting "1 element skipped". Emit exactly one such warning per file,
positioned at the root `<svg>` offset so its `file:line:column` is stable. Consequence, accepted and
recorded as constraint K6: a warning trips `--strict`, so a diagram with any transformed element now
fails under `--strict` where it previously passed. The default (non-strict) exit code is unchanged.

---

## §E Risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | The association window blames a connector a non-label `<rect>`+`<text>` pair merely sits near → false positive | The window is bounded at 16 by the band it feeds (§B); `legend-chip-near-connector` and `badge-chip-near-connector` are clean fixtures placing exactly that pair at 20 units, and `chip-in-window` pins the 11-16 residual as an accepted `{SVG073}`. **Realised** at the 24-unit window this plan first carried; narrowed, not eliminated (§B) |
| R1b | Non-rendered geometry (`<defs>` markers, symbols, clip paths) leaks into a candidate set → false positive | B0 excludes the subtrees; the arrival-only rule and B1's `fill` clause are further guards, though neither is universal — B1 misses the `fill="none"` sketch arrowhead, which is why AC-SGC-006 pins that exact shape. **Realised** on `a11y-present.svg` under the pre-amendment model (§B4) |
| R1c | A documented multi-connector idiom is reported as an error → false positive on an archetype | Arrival-only binding (spec.md §B D7) with `c4-fanout-shared-origin` and `c4-tree-stem` as clean fixtures. **Realised**: `authoring.md` §2.3 and `archetypes.md` A4 both emitted `SVG072` under the literal reading of C4 |
| R1d | The `Q` control point biases C2 distance inward → a compliant mask reported at the documented minimum | Subdivision at t = 0.25/0.5/0.75 (§C), with `c2-mask-outer-elbow-corner-at-6` declared clean. **Realised** in the arithmetic: ≈2.83 units of inward bias at `r = 8` against a 6-unit threshold |
| R2 | The `d` reader silently swallows a real connector, so a genuine breach goes unreported | Accepted by D3, but bounded: three must-flag fixtures (`Q` elbow, `A` hop, relative form) emit their code only if the reader parsed the command, so a reader that returned `null` for everything fails the suite rather than passing silently |
| R3 | Label masks in real diagrams sit inside a `<g>`, or outside the 16-unit window as `archetypes.md` A2's branch labels do (≈99 units) | Accepted false negatives, both disclosed — B3's narrowness here, A2's placement in `spec.md` §A. A follow-up card may widen either once real-world shapes are sampled |
| R4 | New checks change the diagnostic count on existing fixtures | `a11y-present.svg` clean / `a11y-missing.svg` exact-set assertions guard this |
| R5 | Template mirror drifts, or leaks internal markers | `diff -rq` parity plus a grep for SPEC/card/date/SHA markers, both ACs |

---

## §F Milestones

Priority order; no time estimates.

### M1 — Fixture harness first (lands RED)

- Write `scripts/test-check-svg.mjs`: enumerate `scripts/fixtures/*.svg`, read each file's leading
  `<!-- expect: … -->` header, spawn `node check-svg.mjs <file> --json`, parse the JSON, compare
  the exact set of `code` values, print `PASS`/`FAIL` per fixture, exit 0/1.
- Backfill expectation headers onto `a11y-present.svg` (empty set) and `a11y-missing.svg` (its
  current exact code set, read from the tool, not assumed).
- Add the new fixtures: `c2-mask-too-close` (SVG070), `c2-mask-clear` (clean), `c2-mask-too-far`
  (SVG073), `c6-mask-partial` (SVG071), `c6-mask-inside` (clean, badge chip), `c6-mask-outside`
  (clean), `c4-attach-crowded` (SVG072, long edge, gap < 12), `c4-attach-coincident` (SVG072,
  gap 0), `c4-attach-short-edge-ok` (clean, edge < 120 with gap between 8 and 12),
  `c4-attach-short-edge-bad` (SVG072, edge < 120 with gap < 8), `c4-attach-spread` (clean),
  `defs-marker-noise` (clean — a `<defs><marker>` arrowhead whose geometry would bind to a box edge
  if it were admitted, plus a `<symbol>` and a `<clipPath>` carrying rects; asserts B0 and B1's
  `fill="none"` clause), `path-cubic-unreadable` (clean — a connector whose `d` uses a `C` cubic,
  asserting the reader's silent skip), `transform-skipped` (exactly `{SVG074}` — three transformed
  elements).
- Add the boundary fixtures that realise the remaining `acceptance.md` § D.9 rows, since that table
  is now the register the numbered ACs delegate to and every row of it must be a fixture the runner
  asserts: `c2-mask-touching` (distance 0), `c2-mask-at-6` and `c2-mask-at-10` (clean, the inclusive
  band edges), `c2-mask-at-11` (SVG073, the first warning value), `c4-attach-at-12` (clean, long
  edge) and `c4-attach-at-8` (clean, short edge), `c4-edge-exactly-120` (12-unit floor applies),
  `c6-mask-over-earlier-rect` (clean — container / band painted first), `c6-mask-contains-later`
  (SVG071), `attach-endpoint-11-off` and `attach-projection-outside-span` (clean — endpoint binds to
  no edge), `mask-no-adjacent-text` (clean — the accepted false negative), `no-transform` (clean —
  zero `SVG074`).
- Add the fixtures the audit's false-positive channels require, each one pinning a rule that would
  otherwise be unguarded: `c4-fanout-shared-origin` and `c4-tree-stem` (clean — the §2.3 and A4
  idioms, arrival-only binding); `c2-mask-outer-elbow-corner-at-6` (clean — the `Q` subdivision);
  `c2-mask-too-close-hop` (`{SVG070}` — an `A` hop, the reader's positive assertion) and
  `path-relative-form` (clean — relative commands and implicit repeated pairs);
  `legend-chip-near-connector` and `badge-chip-near-connector` (clean — non-label `<rect>`+`<text>`
  pairs at 20 units, outside the association window, plus one at 13 units pinning the residual
  in-window warning as an accepted case — fixture `chip-in-window`, `{SVG073}` — per plan.md §B); `c6-mask-over-later-mask`
  (clean — mask-over-mask is outside C6); `transform-wrapper` (`{SVG074}` — one `<g transform>`
  wrapping the diagram; the fixture pins the code set only, since the runner compares codes, and the
  **count semantics are asserted by AC-SGC-011's message criterion**, read by a human).
- Add the arrival-semantics fixtures, which exist because a clean result alone cannot distinguish
  "the marker rule held" from "the check was never wired": `c4-closed-both-markers`
  (`{SVG072}` — a closed `fill="none"` path carrying both markers, coincident arrivals on one edge);
  `c4-markerless-pair` (clean) paired with `c4-marker-pair-emits` (`{SVG072}` — the same geometry
  with `marker-end` on both). Each pair differs in exactly the property under test.
- Each fixture satisfies the accessible contract and structural checks (REQ-SGC-015).
- Expected state at the end of M1: the runner exits 1, because the must-flag fixtures produce no
  SVG07x code yet. That failure is the RED signal and is recorded as evidence.

### M2 — Geometry engine and the five diagnostics

- Add the `d` reader (§C) including the `Q` subdivision, the box/edge model, the arrival-point rule (§B), the mask-candidate rule (§B3), and the 16-unit association
  rule, all as helpers inside `check-svg.mjs`.
- Add the five checks (§D) into `lint()`, after `SVG040`, using `report()`.
- Runner goes GREEN: every fixture's exact code set matches.
- Re-confirm `a11y-present.svg` still exits 0 and `a11y-missing.svg` still exits 1.

### M3 — Documentation, mirror, build

- `check-svg.mjs` header comment: add the SVG07x tier line.
- `SKILL.md`: runner row in the script table, updated fixtures row, slop-symptom row 12 annotated
  with the codes now checking it.
- `authoring.md`: §2.5 annotates C2/C4/C6 with their codes; §8.3 points at the runner.
- Mirror every added/changed file into
  `internal/template/templates/.claude/skills/moai-domain-svg-infographic/`, run `make build`,
  verify `diff -rq` parity and neutrality.
- Go-side verification limited to the template-embed packages `make build` touches; the full
  verdict is CI's.

---

## §G Anti-patterns

- A second script, a second parser, or a second geometry model (breaks D4/REQ-SGC-008).
- Reporting a diagnostic derived from a `d` the reader could not fully interpret.
- One transform-skip note per element instead of one aggregate per file.
- Asserting only exit codes in the runner instead of the exact code set.
- A new `class` or attribute convention to make classification easier (breaks D1).
- Any SPEC ID, card id, date, or SHA inside `internal/template/templates/**`.
- `go test ./...` locally.

---

## §H Cross-references

- `spec.md` — requirements and out-of-scope
- `acceptance.md` — the AC matrix and the named verification commands
- `references/authoring.md` §2.5 — the rule text being mechanised
