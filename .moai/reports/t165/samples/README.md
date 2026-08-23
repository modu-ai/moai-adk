# Mermaid-bypass evidence (REQ-7)

The exception list in `SKILL.md` § Step 0 admits a diagram type only when a
rendered pair exists for it: the mermaid form, and the form the absorbed rules
produce. This directory holds those pairs. A type with no pair here may not
appear on the list, whatever the argument for it.

## Candidate set

The SPEC seeded four candidates. Their status after this run:

| Candidate | Pair produced | Outcome |
|---|---|---|
| `journey` | yes — `journey-mermaid.png` / `journey-absorbed.png` | comparison below; decision is the reader's |
| `timeline` | no | maps onto archetype A2 like `journey` does, but no pair was rendered, so it cannot be listed |
| `quadrant` | no | dropped on scope: a two-axis scatter matches none of the four archetypes, and adding one is the type-catalogue path this SPEC rejects |
| `ER-schema` | no | dropped for the same reason, and the routing table already sends ER to mermaid explicitly |

## How the pair was produced

```
mermaid   mmdc -i journey.mmd -o journey-mermaid.png -s 2 -b white
          @mermaid-js/mermaid-cli, rendered through Google Chrome 151.0.7922.170
absorbed  node scripts/render.mjs journey-absorbed.svg --out journey-absorbed.png
          same browser; 2400x1140 verified against the 2x target from the viewBox
```

`journey-absorbed.svg` lints clean under `check-svg.mjs` (0 errors). It reports
two `SVG030` warnings, both on the legend text: the heuristic measures a label
against the nearest preceding `<rect>`, which for a legend entry is its 10-unit
colour swatch rather than a container. Triaged against the render as noise.

## What the two renders differ on

Observations from the two PNGs, not claims about mermaid in general:

| | mermaid | absorbed rules |
|---|---|---|
| Canvas use | roughly the lower 60% is empty; the emoji satisfaction marks float in open space | every band is occupied; the canvas ends where the content does |
| Focal signal | three pastel section fills (blue, yellow, pink) carry no meaning ranking; nothing marks the 2-point drop | one accent on the drop — card, bar, and step number — and nothing else |
| Satisfaction encoding | five discrete emoji faces, read one at a time | bar heights on a shared baseline, so the shape of the journey is visible at a glance |
| Longest CJK label | `두 번째 세션 복귀` set on one line by widening its box, so the boxes are unequal | budgeted against the 8-character CJK capacity and wrapped to two lines, boxes stay equal |
| Legend | a two-item key floating at the top-left, inside the diagram area | its own strip below a hairline, outside the diagram area |
| Type roles | one sans family throughout | serif title, CJK-first sans for names, mono for step numbers and the axis |

## What this comparison does not establish

- It is one sample of one journey. A different journey — more stages, Latin
  labels, no satisfaction track — could narrow the gap.
- "Better" is a design judgement the SPEC leaves to a human. What is recorded
  here is that two artifacts exist and what separates them.
- The cost side is real and does not appear in either image: a carve-out makes
  every future diagram of this type an image, with an image's maintenance cost,
  and takes it off the locale-sync path. That cost is why the list is capped by
  evidence rather than by argument.
