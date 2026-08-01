---
name: moai-domain-svg-infographic
description: >
  Author editable SVG technical infographics — architecture, flow, comparison,
  hierarchy — by computing the layout numerically before writing markup, then
  rendering a 2x PNG via headless Chromium. Carries a CJK-first font stack, a
  deterministic source lint, and mermaid-vs-SVG selection rules.

when_to_use: >
  Use for a static diagram image bound for slides, email, social, or offline
  use, or a freeform architecture infographic needing pixel control or precise
  Korean line wrapping. Markdown-embedded diagrams that change often or stay
  locale-synced remain mermaid.

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
user-invocable: true
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-07-24"
  modularized: "true"
  tags: "svg, infographic, diagram, architecture, flow, png, chromium, cjk, layout"
  related-skills: "moai-domain-html-report"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
---

# SVG Technical Infographic

Produce a hand-editable SVG diagram whose geometry was decided by arithmetic
rather than by eye, plus a 2x PNG raster of it. The output is one static image:
no animation, no scripting, no external asset at view time.

## Step 0 — Decide whether this is an SVG job

This skill is **additive to the mermaid pipeline, never a replacement for it**.
Nothing here migrates, rewrites, or deprecates an existing mermaid diagram, and
no diagram should ever exist in both forms — that is dual maintenance, and it is
the one failure this section exists to prevent.

Route the request before drawing anything:

| Signal | Route to |
|--------|----------|
| The diagram lives inside a markdown document | mermaid |
| It changes often, alongside the prose around it | mermaid |
| It is a standard type: flow, sequence, ER, state, class, gantt | mermaid |
| Its text labels are kept in sync across locales | mermaid |
| The deliverable is an image file for slides, email, social, or offline reading | this skill |
| It is a freeform architecture or concept infographic with no standard shape | this skill |
| Pixel-level control of position, spacing, or layering is required | this skill |
| Korean or other CJK labels must wrap at exact, verified widths | this skill |

When several signals point both ways, mermaid wins: a mermaid block is cheaper
to keep correct than an image. Choose this skill only when the routing table
gives it an unopposed reason.

**One diagram, one home.** If a mermaid version already exists, either replace it
outright (and delete the mermaid block in the same change) or leave it alone.
Never ship both.

## Runtime prerequisites and degradation

Node 18 or later and a headless Chromium-family browser are needed **only to
lint and to render**. Neither is needed to install this skill, to discover it, or
to author the editable SVG — authoring is always available.

| Node 18+ | Headless Chromium | What is delivered |
|----------|-------------------|-------------------|
| present | present | Editable SVG, machine lint report, 2x PNG with the browser executable and version disclosed and PNG header dimensions verified |
| present | absent | Editable SVG plus machine lint report. State plainly that no headless browser was found and no PNG was produced |
| absent | either | Editable SVG plus the manual checklist result from `references/authoring.md`. Do **not** attach a machine-lint label, and do not claim a render |

Never fabricate a PNG, a pixel dimension, or a lint verdict for a tool that did
not run. Say which step was skipped and why.

## The workflow

Six steps, in order. Steps 1 through 3 finish before a single SVG element is
written; that ordering is the whole method.

1. **Frame.** Write down the message the diagram must land, the target medium
   (slide, README hero, email, print), the canvas size, and the label language.
2. **Pick an archetype.** Architecture stack, left-to-right flow, side-by-side
   comparison, or hierarchy tree. Skeletons are in `references/archetypes.md`.
3. **Run the numeric layout pass.** Produce the box table and pass every
   containment and text-budget check below. Do not proceed on a failing row.
4. **Author the SVG** from the table. Every coordinate is either a table value
   or a formula over table values.
5. **Lint the source** with `scripts/check-svg.mjs`. Clear every error; triage
   every warning.
6. **Render and verify** with `scripts/render.mjs`. Confirm the reported PNG
   dimensions match the requested 2x target, then look at the PNG.

## The numeric layout pass

Build one table before authoring. Five owned columns per box — `id`, `x`, `y`,
`w`, `h` — and nothing else is typed by hand. Every other number in the file is
derived from those.

**Grid.** For canvas width `W`, `n` columns, outer margin `M`, gutter `G`:

```
colW    = (W - 2*M - (n-1)*G) / n
colX(i) = M + i * (colW + G)
```

If `colW` falls below the archetype's minimum card width, reduce `n` or widen
`W`. Do not shrink the margin to rescue a column count.

**Containment.** Check every row, and stop if any fails:

```
M <= x            and  x + w <= W - M
M <= y            and  y + h <= H - M
parent.x + pad <= child.x   and  child.x + child.w <= parent.x + parent.w - pad
```

**Derived geometry.** Centers and anchors come from the box, never from a
per-language nudge:

```
cx            = x + w/2
cy            = y + h/2
iconCenter    = (x + pad + iconR, y + h/2)
titleBaseline = y + pad + titleSize
lineBaseline(k) = titleBaseline + titleGap + k*lineHeight
```

If you find yourself moving an icon down three units "because the Korean text
sits low", the formula is wrong. Fix the formula, not the instance. Hand-tuned
per-language offsets are exactly the render-fix loop this method removes.

**Connectors.** Endpoints are derived too; the arrowhead length is subtracted
from the terminal end so the marker tip lands on the border, not inside it:

```
horizontal A->B: (A.x + A.w, A.cy) -> (B.x - markerLen, B.cy)
vertical   A->B: (A.cx, A.y + A.h) -> (B.cx, B.y - markerLen)
elbow      A->B: midX = (A.x + A.w + B.x) / 2
                 path: M A.x+A.w A.cy  H midX  V B.cy  H B.x-markerLen
```

Full formula set, including radial and multi-lane fan-out, is in
`references/authoring.md`.

## Text budget — CJK first

Set a CJK-first font stack on the root so Hangul, Kana, and Han glyphs resolve
before any Latin fallback is consulted. A Latin-first stack makes CJK glyphs fall
through to an arbitrary system font and silently changes every measured width:

```
font-family="Pretendard, 'Noto Sans KR', 'Noto Sans JP', 'Noto Sans SC',
             'Apple SD Gothic Neo', 'Hiragino Sans', 'Microsoft YaHei',
             system-ui, sans-serif"
```

Capacity per line, for usable width `u = w - 2*pad` at font size `s`:

```
Latin: capacity = u / (0.60 * s)     average Latin advance is about 0.60em
CJK:   capacity = u / (1.00 * s)     full-width advance is 1.00em
```

The ratio between them is the working rule: **a Korean, Japanese, or Chinese
line holds roughly 60% of the character count a Latin line holds** in the same
box at the same size. Budget the copy against that number and **edit the wording
to fit before authoring**. A line that mixes scripts is budgeted at the CJK rate
for its whole length.

Two things are forbidden here because both hide the problem instead of solving
it: truncating a label after the fact, and shrinking the font size for one
language only. Rewrite the label.

## Linting the source

```bash
node ${CLAUDE_SKILL_DIR}/scripts/check-svg.mjs diagram.svg          # human-readable diagnostics
node ${CLAUDE_SKILL_DIR}/scripts/check-svg.mjs diagram.svg --json   # machine-readable
node ${CLAUDE_SKILL_DIR}/scripts/check-svg.mjs diagram.svg --strict # warnings also fail
```

Every diagnostic carries `file:line:column`, a stable code, and a message. The
two tiers are not interchangeable:

**Errors — deterministic, always fix.** Unbalanced tags; missing or malformed
`viewBox`; a `width`/`height` pair whose aspect ratio contradicts the `viewBox`;
duplicate `id`; a local reference (`url(#id)`, `href="#id"`) with no matching
`id`; a `<marker>` missing required geometry; a `<marker>` that leans on the
implicit `markerUnits` default, which rescales arrowheads with stroke width and
is the usual cause of arrowheads that look right in one diagram and wrong in the
next.

**Warnings — heuristic, triage individually.** Estimated text overflow of its
container rect; a pill too narrow for its label once the round-cap inset is
applied; an element extending past the `viewBox`. These use character-advance
estimation, so they are advisory: confirm in the rendered PNG rather than
reflowing the layout on the warning alone. A warning that survives visual
inspection is a real defect; one that does not is noise.

Exit status is `0` when no error was found, `1` on any error (or on any warning
under `--strict`), `2` on a usage or read failure.

Without Node, walk the manual checklist in `references/authoring.md` instead and
report it as a manual check — never as a lint result.

## Rendering and verifying the PNG

```bash
node ${CLAUDE_SKILL_DIR}/scripts/render.mjs diagram.svg --out diagram.png            # 2x default
node ${CLAUDE_SKILL_DIR}/scripts/render.mjs diagram.svg --out diagram.png --scale 3
```

The renderer resolves a Chromium-family executable from `CHROME_PATH`, then from
the well-known install locations for the platform, then from `PATH`. It reports
**the exact executable it used and that browser's version string** — always
include both in the deliverable, because a diagram rendered by a different
browser build is a different artifact.

It computes the target as `round(viewBox_w * scale) x round(viewBox_h * scale)`,
screenshots at that window size, then reads the PNG's own `IHDR` header and
compares the stored dimensions against the target. A mismatch is a failure, not
a rounding note.

Exit status: `0` verified, `1` render or verification failed, `2` no headless
browser found, `3` usage error. **Exit 2 is the degradation signal** — deliver
the SVG alone and state the limitation.

## Bundled references

| File | Contents |
|------|----------|
| `references/archetypes.md` | The four archetype skeletons with their canvas presets, grid parameters, and per-archetype containment rules |
| `references/authoring.md` | Full geometry and connector formula set, the icon set, palette and type scale, and the manual no-Node checklist |
| `references/sketch.md` | Opt-in hand-drawn preset layered over the same computed layout |

| Script | Purpose |
|--------|---------|
| `scripts/check-svg.mjs` | Deterministic source lint, errors and warnings, `file:line:column` diagnostics |
| `scripts/render.mjs` | Headless-Chromium 2x PNG render with browser disclosure and PNG header verification |

Both scripts run on the Node 18 standard library alone. There is no package to
install and no browser bundled.

## Relationship to the report renderer

`moai-domain-html-report` renders a markdown report into one self-contained HTML
file and may embed mermaid inside it. That skill owns reports; this one owns
standalone diagram images. They compose — a report may link or embed a PNG this
skill produced — and neither replaces the other.

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I will sketch the SVG first and fix the coordinates once I see it" | That is the render-fix loop. Each visual fix invalidates a neighbour and the diagram never converges. Compute the table first. |
| "The Korean label is only slightly too long, it will fit" | It will not: CJK glyphs are full-width, so the line holds about 60% of the Latin count. Rewrite the label before authoring. |
| "I nudged the icon down 3 units and it looks right now" | A per-instance nudge means the center formula is wrong. Derive from box geometry and the nudge disappears everywhere. |
| "No browser here, but the PNG would have been 2400x1600" | An unrendered size is a guess. Deliver the SVG and state that no PNG was produced. |
| "The lint only reported warnings, so the file is clean" | Warnings are heuristic, not absent. Triage each against the rendered PNG before dismissing it. |
| "This flowchart would look nicer as an SVG" | A markdown-embedded, frequently-changing standard diagram stays mermaid. Nicer is not a routing reason. |
| "I will keep the mermaid block and add the SVG for slides" | Two sources for one diagram drift apart. Pick one home. |
<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- SVG elements were written before the box table existed.
- A coordinate in the file cannot be traced to a table value or a formula.
- The same diagram exists as both a mermaid block and an SVG.
- A font stack lists a Latin family before any CJK family.
- A label was truncated, or a font size was reduced for one language only.
- A PNG dimension, browser version, or lint verdict is reported for a command
  that was never run.
- Lint errors were downgraded to warnings to get to a render.
- A `<marker>` has no explicit `markerUnits`.
<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Routing table consulted and the SVG choice has an unopposed reason.
- [ ] No mermaid version of this diagram remains alongside the SVG.
- [ ] Box table complete before authoring; every coordinate traces to it.
- [ ] All containment checks pass for boxes and for children inside boxes.
- [ ] Centers, baselines, and connector endpoints are derived, not hand-tuned.
- [ ] Font stack is CJK-first; every line fits its computed capacity.
- [ ] `check-svg.mjs` reports zero errors; each warning triaged and recorded.
- [ ] `render.mjs` verified the PNG header against the 2x target.
- [ ] Browser executable and version disclosed with the PNG.
- [ ] Any skipped step named explicitly, with no substitute claim.
<!-- moai:evolvable-end -->
