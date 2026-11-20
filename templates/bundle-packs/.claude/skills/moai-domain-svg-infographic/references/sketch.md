# Sketch Preset (opt-in)

A hand-drawn surface treatment for an infographic that should read as a
whiteboard sketch rather than a produced diagram. Useful for early design
discussion, workshop material, and "this is a proposal, not a decision" framing.

**The preset changes appearance only.** The numeric layout pass runs exactly as
it does for the default surface: the same box table, the same containment
checks, the same derived centers and connector endpoints, the same text budget.
Sketchiness is applied on top of the computed geometry, never in place of it. A
sketch diagram whose boxes were placed by eye is not a sketch preset — it is an
unplanned diagram wearing a rough stroke.

## When to use it

| Situation | Preset |
|-----------|--------|
| Proposal, RFC illustration, workshop handout | sketch |
| Early architecture exploration still open to change | sketch |
| Shipped documentation, release material, slide for an external audience | default |
| Anything where the roughness could read as unfinished work | default |

Do not mix the two surfaces in one diagram. A sketch box next to a crisp box
reads as a rendering bug.

## 1. Roughened outlines

Replace each `<rect>` with a closed `<path>` whose corners are offset by a small
deterministic jitter. Determinism matters: a random jitter regenerates
differently on every edit, so the diff becomes unreadable and a re-render is not
reproducible. Derive the offset from the box's own coordinates instead.

```
jitter(n) = ((n * 2654435761) % 1000) / 1000 * 2*A - A      A = 2.5 units
```

`2654435761` is a fixed odd multiplier; any fixed odd constant works. The point
is that the same box always produces the same offsets.

For a box `(x, y, w, h)`, offset the four corners and draw with a quadratic
through each edge midpoint so the sides bow slightly:

```
p0 = (x   + jitter(x+y),       y   + jitter(x-y))
p1 = (x+w + jitter(x+w+y),     y   + jitter(x+w-y))
p2 = (x+w + jitter(x+w+y+h),   y+h + jitter(x+w-y-h))
p3 = (x   + jitter(x+y+h),     y+h + jitter(x-y-h))

M p0  Q mid(p0,p1)+bow p1  Q mid(p1,p2)+bow p2
      Q mid(p2,p3)+bow p3  Q mid(p3,p0)+bow p0  Z

bow = perpendicular offset of 1.5 units, sign from jitter of the edge midpoint
```

Draw each outline twice with slightly different jitter seeds and
`stroke-opacity="0.55"` to imitate a pen passing over the same line.

## 2. Stroke and fill

| Property | Value |
|----------|-------|
| `stroke-width` | `2.2` |
| `stroke-linecap` | `round` |
| `stroke-linejoin` | `round` |
| `fill` | `none`, or a hatch pattern for emphasis |
| `stroke-opacity` | `0.55` per pass, two passes |

Solid fills defeat the effect. Where a fill is needed, use a hatch pattern in
`<defs>` — and give it an explicit `patternUnits`, for the same reason markers
need an explicit `markerUnits`:

```xml
<pattern id="hatch" patternUnits="userSpaceOnUse" width="6" height="6"
         patternTransform="rotate(45)">
  <line x1="0" y1="0" x2="0" y2="6" stroke="#57606a" stroke-width="1"
        stroke-opacity="0.35"/>
</pattern>
```

## 3. Connectors

Apply the same treatment to the derived endpoints from the authoring reference.
Straight segments become single quadratics bowed by 3 units perpendicular to
their direction; elbows keep their corner positions and gain a rounded corner
radius of 10 rather than 8.

The arrowhead marker changes shape but keeps its explicit `markerUnits` and its
`refX = markerLen` relationship, so the endpoint subtraction still holds:

```xml
<marker id="arrow-sketch" markerUnits="userSpaceOnUse"
        markerWidth="12" markerHeight="10" refX="12" refY="5" orient="auto">
  <path d="M 0 1 L 12 5 L 0 9" fill="none" stroke="#57606a"
        stroke-width="1.6" stroke-linecap="round"/>
</marker>
```

Note `markerLen` becomes 12 under this preset. Recompute the endpoint
subtraction with the preset's value rather than carrying 10 over.

## 4. Typography

A handwritten font is optional and is the one part of this preset that can fail
on a viewer's machine. Keep the CJK-first stack in front so Hangul, Kana, and
Han glyphs never fall through to an arbitrary fallback:

```
font-family="'Nanum Pen Script', 'Yomogi', 'Pretendard', 'Noto Sans KR',
             'Noto Sans JP', 'Noto Sans SC', system-ui, sans-serif"
```

Handwritten faces are usually narrower than the 0.60em Latin average the text
budget assumes, so a label that fit under the default surface still fits here.
The reverse is not guaranteed: do not re-budget upward on the assumption that a
narrow face buys room, because the fallback face may be wider than the
handwritten one.

Nudge each text element's baseline by `jitter(x+y)` for a hand-set feel. This is
a rendering offset applied after the derived baseline, not a replacement for it
— the derived value stays in the table.

## 5. Verification differences

The lint script treats a sketch diagram exactly like any other SVG, and every
error class still applies: the doubled outlines mean roughly twice as many
elements, and duplicate `id` values are the failure that appears most often when
outlines are copied for the second pass. Suffix the second pass (`-a` / `-b`)
so the ids stay unique.

Text-fit warnings are noisier under this preset because the handwritten face is
narrower than the estimator assumes. Confirm each one against the rendered PNG
before acting; the estimator has no knowledge of the loaded face.
