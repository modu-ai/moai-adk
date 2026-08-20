---
name: moai-domain-design-dna
description: >
  Reverse-engineer a reference design — a screenshot, an image set, or a live
  URL — into a portable Design DNA JSON across three dimensions (measurable
  tokens, qualitative style, special-rendering effects), then generate a new
  self-contained artifact from that JSON. Carries the extraction rules
  (dominance-based colour roles, relative radius measurement, multi-reference
  conflict resolution), a performance-tier technology map for Canvas / WebGL /
  shader / scroll effects, and a delivery gate covering contrast, reduced
  motion, and animation-loop hygiene.

when_to_use: >
  Use when someone points at an existing design and wants its look captured or
  reproduced — "build this in the same style as this screenshot", "extract the
  design tokens from this site", "make a page that feels like this reference" —
  or when a Design DNA JSON already exists and a new artifact must be generated
  from it. Not for report rendering (moai-domain-html-report), static diagrams
  (moai-domain-svg-infographic), or charts (dataviz).

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, WebFetch, Bash
user-invocable: true
metadata:
  version: "1.0.0"
  category: domain
  status: active
  related-skills: moai-ref-ui-polish, moai-domain-html-report, moai-domain-svg-infographic
---

# Design DNA

Most "make it look like this" requests fail the same way: the reference is
looked at once, an impression is formed, and the impression is coded from
memory. What survives is a vague resemblance — the palette drifts, the rhythm
flattens, and the one effect that gave the reference its character is missing
entirely.

This skill inserts an artifact between looking and building. The reference is
first deconstructed into a **Design DNA JSON**, and only that JSON is used to
generate. The intermediate step is what makes the result checkable: every
colour, radius, and easing curve in the output traces to a recorded field, so a
mismatch is a diff rather than an argument about taste.

> **Provenance**: the three-dimension taxonomy and several extraction rules are
> adapted from the MIT-licensed `zanwei/design-dna` skill. See
> `.claude/rules/moai/NOTICE.md` for the retained copyright notice.

## The three dimensions

The split matters because the three are extracted differently and fail
differently.

| Dimension | What it holds | How it is obtained |
|---|---|---|
| `design_system` | What can be **measured** — colour hex values, type scale, spacing base unit, radius, elevation, motion timings, component patterns | Sampled and measured from the reference; numeric |
| `design_style` | What can be **felt** — mood, ornamentation level, composition strategy, whitespace philosophy, interaction personality, microcopy tone | Judged holistically; descriptive words, not numbers |
| `visual_effects` | What **cannot be expressed in plain CSS** — Canvas scenes, WebGL / 3D, shaders, particle systems, scroll-driven motion, cursor behaviour, glassmorphism | Read from source where available, described from screenshots where not |

Collapsing them loses information in both directions. A token dump with no
style dimension reproduces the colours and none of the character; a mood board
with no system dimension reproduces the vibe and no two elements align.

Field-by-field schema and enum vocabularies: `references/dna-schema.md`.

## Phase 1 — Structure

When the request is for the schema itself ("what does a design profile
contain?"), read `references/dna-schema.md`, present the three dimensions and
their field groups, and ask whether any dimension should be extended or
dropped before extraction begins.

## Phase 2 — Analyze (reference → DNA JSON)

Read `references/dna-schema.md` first, then work reference by reference.

- **Image or screenshot** → read it directly and analyse its visual properties.
- **URL** → fetch the page. Prefer the source: `<canvas>` elements, WebGL
  contexts, animation-library imports, custom shaders, `IntersectionObserver`
  scroll triggers, and SVG `<animate>` are *stated* facts, where a screenshot
  only supports inference.
- **Video or interaction capture** → the only reliable source for motion
  timings, scroll choreography, and transition personality.

### Extraction rules

1. **Colour roles are assigned by dominance, not by hue.** Primary is the
   colour holding the most area, secondary the supporting surface, accent
   whatever carries the calls to action. Map the neutral scale as an ordered
   ramp from the lightest background to the darkest text.
2. **Measure radius relatively.** Record it against the element height that
   carries it — an absolute `12px` is meaningless once the button is a
   different size, while "half the control height" survives rescaling. The
   concentric-radius rule for nested surfaces is `moai-ref-ui-polish`'s; do not
   restate it here, apply it there.
3. **Infer type scale from ratios**, not from absolute sizes. Heading-to-body
   ratio and line-height rhythm transfer across viewports; pixel sizes do not.
4. **Density is proximity.** Judge spacing from element gaps relative to the
   base unit, and section rhythm from whether those gaps stay constant or
   escalate between sections.
5. **When references conflict, record the dominant pattern and name the
   variant.** Averaging two references produces a design neither of them has.
6. **Fill every field.** An empty string is indistinguishable from "not looked
   at". Where the reference genuinely does not exercise a field, say so in the
   value — an explicit "not observed" is data; a blank is a gap.
7. **Effects absent from the reference get `enabled: false`.** This is the
   non-invention rule, and it is load-bearing: an unset flag invites the
   generator to add a particle field nobody asked for.
8. **What cannot be identified goes in `composite_notes`.** A screenshot shows
   that a surface glows without showing how. Describing the observation beats
   guessing the implementation and recording the guess as fact.

Output the completed JSON, then ask whether any value should be adjusted before
generation.

## Phase 3 — Generate (DNA JSON + content → artifact)

Read `references/effects-implementation.md` before implementing any
`visual_effects` entry.

**Order matters, because early decisions constrain later ones.** Colour and
typography together carry most of a design's identity, so they are settled
first; effects are layered onto a design that already works without them.

1. Colour and typography
2. Spacing and layout
3. Shape and elevation
4. `design_style` qualitative fields — these guide the judgement calls the
   token values do not determine
5. `visual_effects`
6. Motion and interaction, last: an interface whose static layout is wrong is
   not rescued by animating it

Emit `design_system` as CSS custom properties in a single `:root` block, so
every downstream value has one definition and a token swap is one edit.

**Fetch real assets rather than approximating them.** When the reference is a
URL and the design needs its logo, image, or font, retrieve the actual asset
from that source. A recreated approximation is the single most visible
difference between a copy and a clone.

**Default output is a self-contained file** — inline CSS and JS, no build
step — unless a framework was specified. The self-contained convention itself
is documented once in `moai-domain-html-report` § output; the difference here
is scope: that skill renders a *report* from markdown, this one generates a
*designed artifact* from a DNA profile.

## Delivery gate

Verify before handing over. Each item below has failed in practice.

- Every colour in the output traces to a DNA palette entry.
- Font families, spacing rhythm, and radius match their DNA tokens.
- Body text meets 4.5:1 contrast, large text 3:1.
- `prefers-reduced-motion: reduce` is honoured — not merely detected.
- Nothing renders for an effect whose `enabled` flag is `false`.
- Animation loops use `requestAnimationFrame`; `setInterval` is not an
  animation primitive and drifts against the display refresh.
- Canvas and WebGL contexts are sized to their container and handle resize.
- The declared `fallback_strategy` is actually implemented, not just recorded.

## Polish iteration

A first pass that reads thin is usually not a token problem — it is an
attention problem, and re-reading the reference beats re-reading the output.
Re-attach the same references and audit against them along six axes:
hierarchy, ornamentation, typographic rhythm, motion, materiality, and overall
interface finish. Merge the conclusions back into the current implementation
rather than regenerating from scratch, which discards the parts that were
already right.

## Boundaries

| Request | Owner |
|---|---|
| Markdown report → HTML | `moai-domain-html-report` |
| Architecture / flow diagram | `moai-domain-svg-infographic` |
| Chart, dashboard, or categorical palette | `dataviz` |
| Component-level finish: concentric radius, optical alignment, hit areas, easing craft | `moai-ref-ui-polish` |
| Hosted claude.ai visual-identity page | `artifact-design` |
| Design-system sync with the Claude Design product | `manager-design` |

This skill owns one thing the others do not: turning an **existing reference**
into a structured profile, and generating from that profile. Where the design
is authored from first principles rather than deconstructed, those skills own
it.

## References

- `references/dna-schema.md` — the three-dimension field list and enum vocabularies
- `references/effects-implementation.md` — performance tiers, technology selection, and per-effect implementation patterns
