# Craft Fundamentals — html-report

Production rules for every mode, adapted from the `design-artifact` skill
(plannotator/effective-html). The **design tokens are fixed** — see § Design
Tokens in `SKILL.md` (ivory / clay / Pretendard). This document is *craft*, not
a new look: spacing, numerals, neutrals, copy, structure, and avoiding the
generic-AI layout. Nothing here changes a token.

---

## Spacing belongs to the layout, not the elements

Sibling groups use flex/grid plus `gap`; per-element margins scatter and collapse
unpredictably. Broad content — tables, code blocks, diagrams — sits in its own
container with `overflow-x: auto` so horizontal scroll never leaks to the page
body. This is why the mode skeletons reach for `gap`, not per-card `margin`.

## Numerals stack in tabular columns

Wherever numbers line up — KPI cards, financial tables, velocity bars — switch on
`font-variant-numeric: tabular-nums`. Proportional digits make a financial
statement twitch; tabular digits align by place value.

## Neutrals are choices

A dead-center mid-grey (`#808080`) announces that nobody thought about it. Tint
the neutral faintly toward the accent instead. This skill's greyscale
(`--g100` / `--g300` / `--g500` / `--g700`) is already warm-tinted to sit with
`--clay`; use it deliberately rather than reaching for a raw grey.

## Copy is a material, not garnish

Stand on the reader's side: name things as people know them, not as the backend
does (someone manages *notifications*, never *webhook config*). Verbs stay
active; every control declares its exact effect. An error message diagnoses the
failure and prescribes the fix — never groveling, never hand-waving. Precision
outperforms wit.

## Make structure mean something

Numbering (01 / 02 / 03), eyebrows, dividers, and labels earn their place by
asserting something true about the content. Numbered markers are honest only
when order is real information — a procedure, a dated timeline the reader
follows in sequence. Do not deploy them as decoration.

## Dodge the telltale AI layout

The token palette (warm ivory ground, clay accent) sits inside the territory
`design-artifact` flags as the generic-AI look; this skill keeps it by policy
("Claude style"). What you control is the *layout*: avoid the remaining clichés
— universal center alignment, `rounded-lg` sprayed on every surface, rounded
cards each wearing an accent bar, decorative emoji as section markers (prefer
inline SVG or typographic markers), a gradient hero floating on white. A
pinned-down direction is executed faithfully; absent one, spend the freedom on
composition, not on the cliché.

## Engineer it soundly

Non-void elements closed, attributes double-quoted, keyboard focus visible,
`prefers-reduced-motion` respected. When the rendering question is
structural-diagram vs quantitative-chart, the answer is in `SKILL.md`
§ Diagram Policy — this file does not override it.

---

## Cross-references

- `SKILL.md` § Design Tokens — the fixed palette (the authority; this file never changes a token)
- `SKILL.md` § Diagram Policy — model selection + the zero-JS rendering rule
- `SKILL.md` § Audience Tiers — what depth each tier adds (craft is orthogonal to tier)
