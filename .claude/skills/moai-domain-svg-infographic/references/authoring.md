# Authoring Reference

Everything the skill body points at: the full geometry and connector formula
set including the six mandatory connector rules, the icon set, the semantic-role
palette and type scale, the accessible-SVG contract, the CJK text-budget worked
example, and the manual checklist used when Node is unavailable.

---

## 1. Geometry formulas

### 1.1 Box-derived anchors

Given a box `(x, y, w, h)` and its inner padding `pad`:

```
cx            = x + w/2
cy            = y + h/2
left          = x
right         = x + w
top           = y
bottom        = y + h
innerLeft     = x + pad
innerRight    = x + w - pad
innerWidth    = w - 2*pad
innerTop      = y + pad
innerBottom   = y + h - pad
innerHeight   = h - 2*pad
```

Every other anchor in the file is a composition of these. If a needed anchor
cannot be written as such a composition, the box table is missing a box.

### 1.2 Vertical text rhythm

SVG `<text>` positions by baseline, not by box top, which is why hand-placed
labels drift between font sizes and between scripts. Derive instead:

```
titleBaseline    = innerTop + titleSize                     first line
lineHeight       = round(bodySize * 1.45)                   1.45 is the CJK-safe factor
titleGap         = round(titleSize * 0.55)
lineBaseline(k)  = titleBaseline + titleGap + k*lineHeight  k starts at 0
```

For a block of `n` lines vertically centered in the box:

```
blockH          = (n-1) * lineHeight
firstBaseline   = cy - blockH/2 + bodySize*0.36             0.36 approximates the cap-height offset
```

The `1.45` line-height factor is not the Latin-typical `1.2`: Hangul and Kana
glyphs occupy more of the em box, and `1.2` produces visually touching lines in
mixed content. Use one factor for all scripts so a translated diagram keeps the
same rhythm.

### 1.3 Icon and badge centering

```
iconCenter  = (innerLeft + iconR, cy)                   left-aligned icon
badgeCenter = (innerLeft + badgeR, innerTop + badgeR)   corner badge
textStart   = innerLeft + 2*iconR + iconGap             text after a left icon
```

A left-aligned icon and its adjacent text share `cy`; they never receive
separate vertical offsets.

### 1.4 Rounded rectangles and pills

```
card:  rx = 12
chip:  rx = 8
pill:  rx = h/2                                         fully rounded caps
```

A pill's usable inner width is smaller than a rectangle's, because the round
caps eat horizontal room where the text sits:

```
pillInnerWidth = w - 2*pad - h*0.30
```

Budget pill labels against `pillInnerWidth`, not `innerWidth`. Ignoring the cap
inset is the most common cause of a label that touches its own border.

---

## 2. Connector formulas

### 2.1 Straight

```
horizontal A->B: (A.right, A.cy)  ->  (B.left - markerLen, B.cy)
vertical   A->B: (A.cx, A.bottom) ->  (B.cx, B.top - markerLen)
```

`markerLen` is the arrowhead length in user units — subtract it so the tip lands
on the target border rather than overlapping it. With the marker definition in
section 4, `markerLen = 10`.

### 2.2 Elbow (orthogonal)

```
horizontal-first: midX = (A.right + B.left) / 2
  M A.right A.cy  H midX  V B.cy  H (B.left - markerLen)

vertical-first:   midY = (A.bottom + B.top) / 2
  M A.cx A.bottom  V midY  H B.cx  V (B.top - markerLen)
```

Rounded elbow corners, radius `r` (use `r = 8`), replace each corner with a
quadratic segment:

```
... H (midX - r)  Q midX A.cy midX (A.cy + sign*r)  V (B.cy - sign*r) ...
   where sign = B.cy > A.cy ? 1 : -1
```

### 2.3 Fan-out (one source, many targets)

Route through a single shared trunk so the lines read as one branch rather than
as several unrelated arrows:

```
trunkX = A.right + fanGap                    fanGap = 40
for each target T:
  M A.right A.cy  H trunkX  V T.cy  H (T.left - markerLen)
```

All targets share `trunkX`. Sort targets by `cy` before emitting so the trunk
segments nest instead of crossing.

### 2.4 Radial (hub and spokes)

For `n` spokes around a hub of radius `R`, with spoke boxes at radius `D`:

```
angle(i)   = -90 + i * (360 / n)             degrees, -90 puts spoke 0 at the top
rad(i)     = angle(i) * PI / 180
spokeCx(i) = hub.cx + D * cos(rad(i))
spokeCy(i) = hub.cy + D * sin(rad(i))
lineStart  = (hub.cx + R*cos(rad(i)), hub.cy + R*sin(rad(i)))
lineEnd    = (spokeCx(i) - (spokeR+markerLen)*cos(rad(i)),
              spokeCy(i) - (spokeR+markerLen)*sin(rad(i)))
```

Spoke boxes are positioned by their center, so convert back:
`x = spokeCx(i) - w/2`, `y = spokeCy(i) - h/2`. Then re-run the containment
check — radial layouts overflow the canvas more often than grid layouts do.

### 2.5 The six mandatory connector rules

The formulas above say where a connector goes. These six say when a connector is
wrong, and they are not preferences — a diagram breaching any of them is a
failure, not a stylistic variant. Each is stated with the number or tolerance it
turns on, so a checker can assert it rather than a reader having to judge it.

**C1 — Orthogonal only; every bend is a rounded right angle at `r = 8`.**
A plain straight segment is allowed only when the two endpoints share an `x` or a
`y`. Otherwise use the elbow of §2.2, with each corner replaced by the quadratic
of that section at `r = 8` (`r = 6` is the floor for a tight layout, where the
shorter of the two segments meeting at the corner is under `4*r`). A diagonal
segment between boxes that share neither axis is an automatic fail.

**C2 — A label's mask clears its own stroke by 6–10 units.**
Every connector label sits on an opaque mask rect, or the stroke shows through
the glyphs. The mask must not touch the stroke either:

```
horizontal segment at y = Sy, label height lh, gap g in [6, 10]:
  maskY = Sy - g - lh          maskBottom = Sy - g
vertical segment at x = Sx, mask width mw:
  maskX = Sx + g               (or Sx - g - mw on the other side)
```

A label whose mask edge lands within 6 units of the stroke hides the connection
it names. If 6 feels cramped for the label size, take 8 or 10 — never less.

Machine-checked as `SVG070` (clearance under 6 units, a mask that touches or
crosses the stroke being the zero case) and `SVG073` (clearance over 10, a
warning). The check associates a mask to a connector only within **16 units**, so
a label placed further from its stroke than that — archetype A2's branch labels,
which sit about 99 units off their connector, among them — associates to none and
is not checked at all.

**C3 — No two connectors share a path; separation is ≥ 12 units.**
Two connectors may not run co-linear for any segment, and two parallel runs stay
`≥ 12` apart along their whole length, not merely at their endpoints. Where two
orthogonal connectors must cross, apply the bridge (hop) at the crossing point
`(X, Y)` with `rHop = 5`, on the horizontal connector by convention:

```
... H (X - rHop)  A rHop rHop 0 0 1 (X + rHop) Y  H ...
```

Stacked connectors are a layout symptom, not a drawing problem: two nodes are
too close, or the diagram is over its node budget. Fix the layout.

**C4 — Connectors on a shared edge fan to distinct attach points.**
No two connectors may meet a box at the same point. For `N` connectors on an
edge of length `L`, connector `k` (counting `1..N` from the edge's leading
corner) attaches at

```
offset(k) = L * k / (N + 1)
```

with `offset(k+1) - offset(k) >= 12` (`>= 8` for a box under 120 units on that
edge). If the even spread cannot hold the floor, the edge carries too many
connectors — re-route some to another edge or split the node. Connectors leaving
a fanned edge route orthogonally from their own attach point; they never merge
into a shared stroke near the box.

Machine-checked as `SVG072`, on **arrival** points only — an endpoint at which a
`marker-end` or `marker-start` resolves — with this rule's own floor: 12 units, 8
on an edge under 120. Two connectors *departing* one point are the deliberate
fan-out of §2.3 and are not reported, so genuine crowding on the departure side
goes uncaught; a connector carrying neither marker has no arrival point and is
never compared.

**C5 — A connector does not pass behind a box that is neither its source nor its
destination — except where that box is geometrically unavoidable.**
Rerouting around an intervening box is the default and covers nearly every case.
The exception is narrow: a cross-cutting node — a full-width layer band, a footer
service — physically sits on the only direct orthogonal path between source and
destination, so no reroute exists. In that case, and only then:

- the stroke is **dashed** (`stroke-dasharray="4,3"`), which is what tells the
  reader the intervening box is transit rather than an endpoint;
- the label sits at the connector's **visible end**, typically near the source,
  so it does not fall behind the intervening box;
- **no arrowhead lands on the intervening box's edge** — the marker resolves at
  the true destination only.

An undashed connector crossing behind a non-endpoint box reads as an unlabelled
interaction with that box, which is a factual error about the system, not an
aesthetic one. When in doubt, reroute.

**C6 — A label mask may not overlap a node painted after it.**
C2 keeps the label off its own stroke; this keeps it off the boxes. Connectors
and their labels are painted before nodes, so a mask landing partly inside a node
is covered by that node's fill and the text renders as a fragment on the border.
Place the label on a segment running through open canvas:

```
label leaving A's right edge: maskX >= A.right + 4
label entering B's left edge: maskX + maskW <= B.left - 4
```

A mask fully inside a node is a badge chip and is fine. A mask over a band or
zone container is fine too, because containers are painted first. Only the
partial overlap with a later-painted node fails.

Machine-checked as `SVG071`, deliberately wider than this rule as written: the
check counts any later-painted `<rect>` that is not itself a label mask,
decorative dividers and rules included, because a label clipped by one renders as
a fragment whatever that rect means. Fully inside and fully outside stay silent.

**Paint order, since three of the six depend on it:** containers and bands
first, then connectors, then connector labels, then nodes, then node text.

**`SVG074` is a coverage note, not a rule breach.** The three checks above read
coordinates straight from the attributes, so an element carrying a `transform` is
excluded rather than guessed at. `SVG074` warns once per file to say how much of
the candidate population that was — transitively, so a single wrapping
`<g transform>` takes everything inside it out of the count. It reports what went
unverified; it says nothing about whether the geometry is wrong.

---

## 3. Palette and type scale

Colour is addressed by **semantic role**, never by a raw hex value. Everything
else in this skill — the archetypes, the focal discipline, the connector
rules — says `accent` or `surface`, and looks the value up here. Swapping the
skin therefore means editing this one table; nothing downstream changes.

The default skin is aligned to the project design system
(`moai-domain-html-report` tokens): warm ivory paper, clay terracotta accent,
warm near-black ink. The skill keeps its own copy of the values — it does not
read a runtime token file — so the diagram renders offline. Substitute a
project's own tokens freely, but substitute the whole set, so the contrast
relationships survive the swap.

| Role | Light | Dark | Use |
|------|-------|------|-----|
| `ink` | `#141413` | `#FAF9F5` | Primary text, primary stroke |
| `ink-muted` | `#6B6359` | `#A79C8E` | Captions, secondary text, default arrow stroke |
| `surface` | `#FAF9F5` | `#141413` | Canvas, default node fill |
| `surface-alt` | `#F3EFE6` | `#232220` | Band fill, secondary container |
| `border` | `#D9CDBE` | `#3A3733` | Hairline card and divider strokes |
| `accent` | `#D97757` | `#E08D6F` | Focal node and the active path into it — 1 per diagram, 2 at the absolute most |
| `accent-soft` | `#FBE9DF` | `#3A2A22` | Fill behind an `accent` border |
| `accent-strong` | `#B85C3E` | `#F0A98D` | Eyebrow text, hover state |
| `positive` | `#1a7f37` | `#3FB950` | Success, allowed |
| `caution` | `#9a6700` | `#D29922` | Warning, degraded |
| `negative` | `#cf222e` | `#F85149` | Failure, forbidden |

### 3.0 The inversion rule

The dark column is **derived, not independently maintained**. One rule produces
it, and a role added later inherits the rule instead of needing a judgement:

> Dark mode swaps the two anchors — `ink` and `surface` exchange values — and
> every other role keeps its **distance from its anchor**, not its hex. A role
> written as `ink @ α` keeps `α` and picks up the new `ink`. `accent` is the
> single exception: it shifts one step lighter (`#D97757` → `#E08D6F`), because
> the light value's contrast against ivory does not survive against near-black.

Two consequences worth stating, because both are easy to lose:

- **Opacities never change.** A hairline at `ink @ 0.12` is `ink @ 0.12` in both
  modes; only the resolved colour moves. Re-tuning an alpha for dark mode is how
  the two modes start to drift apart.
- **`ink` on `surface` must clear WCAG AA in both modes**, and `ink-muted` on
  `surface` must clear AA for text at 12 units and above. A substituted skin that
  fails either check is not a skin choice, it is a legibility defect.

### 3.1 Typography — language-aware font stacks

Headings use a serif (the editorial register); body uses a CJK-first sans that
resolves Hangul/Kana/Han before any Latin fallback. Map the heading serif per
locale; the body stack is shared across locales and is already CJK-first.

| Locale | Heading serif | Body sans |
|--------|---------------|-----------|
| `ko` | MaruBuri | Pretendard |
| `en` | Noto Serif | Noto Sans |
| `ja` | Noto Serif JP | Noto Sans JP |
| `zh` | Noto Serif SC | Noto Sans SC |

Heading `font-family`, pick one by locale:

- `ko`: `MaruBuri, 'Noto Serif KR', Georgia, serif`
- `en`: `'Noto Serif', Georgia, serif`
- `ja`: `'Noto Serif JP', serif`
- `zh`: `'Noto Serif SC', serif`

Body `font-family` (CJK-first, same for every locale — CJK glyphs resolve
before the Latin fallback, so a translated diagram keeps the same measured
widths):

```
font-family="Pretendard, 'Noto Sans KR', 'Noto Sans JP', 'Noto Sans SC',
             'Apple SD Gothic Neo', 'Hiragino Sans', 'Microsoft YaHei',
             system-ui, sans-serif"
```

Mono (eyebrows, port numbers, field types, sublabels):
`'JetBrains Mono', ui-monospace, monospace`.

Type scale, in user units:

| Role | Size | Weight |
|------|------|--------|
| Diagram title (serif) | 34 | 600 |
| Section or layer label | 20 | 600 |
| Card title | 17 | 600 |
| Body | 14 | 400 |
| Caption | 12 | 400 |

Keep body text at or above 14 so the 2x PNG stays legible when a slide is
projected. If a label only fits below 14, the label is too long.

### 3.2 Focal discipline — single accent, signalled by colour

One diagram carries **one focal element** (two at the absolute most). Reserve
the `accent` token for that focal node and the active path that leads into it;
every other node stays `ink` / `muted` on `paper`. The focal is signalled by
colour, never by a floating callout:

- **focal node** — `accent-soft` (`#FBE9DF`) fill + `accent` (`#D97757`) border
  at 1.6px + a small uppercase `★ FOCAL` eyebrow in `accent-strong`
- **active-path arrow** entering the focal node — `accent` stroke
- **non-focal nodes** — `paper` fill + `border` hairline (1px)

Scattering `accent` across several "important" nodes erases the signal. A
floating editorial callout (dashed leader + italic serif) belongs to the sparse
small-node layouts of editorial diagrams; in this skill's dense banded layouts
a leader will cross the orthogonal connectors, so do not add one — let the
colour and the `★ FOCAL` eyebrow carry the emphasis.

---

## 4. Marker definition

Define arrowheads once in `<defs>` with an explicit `markerUnits`. Omitting
`markerUnits` selects the `strokeWidth` default, which rescales every arrowhead
with its line's stroke width — the same marker then renders at different sizes
across the diagram:

```xml
<defs>
  <marker id="arrow" markerUnits="userSpaceOnUse"
          markerWidth="10" markerHeight="8"
          refX="10" refY="4" orient="auto">
    <path d="M 0 0 L 10 4 L 0 8 z" fill="#57606a"/>
  </marker>
</defs>
```

`refX` equal to `markerWidth` places the tip at the path endpoint, which is what
the `- markerLen` subtraction in section 2 assumes. Changing one without the
other reintroduces the overlap.

---

## 5. Icon set

Twelve single-path glyphs on a 24x24 grid, drawn with `stroke` and no `fill` so
they inherit color from their group. Scale by wrapping in
`<g transform="translate(cx-12*k, cy-12*k) scale(k)">`.

| Name | Path `d` |
|------|----------|
| server | `M3 5h18v6H3z M3 13h18v6H3z M7 8h.01 M7 16h.01` |
| database | `M12 3c5 0 9 1.3 9 3s-4 3-9 3-9-1.3-9-3 4-3 9-3z M3 6v12c0 1.7 4 3 9 3s9-1.3 9-3V6` |
| cloud | `M6 18a4 4 0 0 1 .8-7.9 6 6 0 0 1 11.5 1.6A3.5 3.5 0 0 1 17.5 18z` |
| user | `M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8z M4 21a8 8 0 0 1 16 0` |
| gear | `M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M12 2v3 M12 19v3 M2 12h3 M19 12h3 M5 5l2 2 M17 17l2 2 M19 5l-2 2 M7 17l-2 2` |
| shield | `M12 3l8 3v6c0 5-3.4 8.3-8 9-4.6-.7-8-4-8-9V6z` |
| box | `M12 3l8 4.5v9L12 21l-8-4.5v-9z M4 7.5l8 4.5 8-4.5 M12 12v9` |
| file | `M6 3h8l4 4v14H6z M14 3v4h4` |
| clock | `M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18z M12 7v5l3 2` |
| bolt | `M13 2L4 14h6l-1 8 9-12h-6z` |
| link | `M9 15l6-6 M10 6l2-2a4 4 0 0 1 6 6l-2 2 M14 18l-2 2a4 4 0 0 1-6-6l2-2` |
| check | `M4 12l5 5L20 6` |

Recommended attributes: `fill="none" stroke="currentColor" stroke-width="1.8"
stroke-linecap="round" stroke-linejoin="round"`.

---

## 6. CJK text budget — worked example

A card is `w = 220`, `pad = 16`, body size `s = 14`.

```
u        = 220 - 2*16                    = 188
Latin    = 188 / (0.60 * 14) = 188/8.4   = 22 characters per line
CJK      = 188 / (1.00 * 14) = 188/14    = 13 characters per line
ratio    = 13 / 22                       = 0.59, the ~60% rule
```

An English label of 21 characters fits. Its Korean translation must therefore be
rewritten to 13 characters or fewer per line, not merely translated:

- Over budget at 19 characters: `사용자 인증 토큰 발급 서비스`
- Within budget at 9 characters: `인증 토큰 발급`
- Within budget as two lines of 8 and 7: `인증 토큰` / `발급 서비스`

The same arithmetic applies to Japanese and Chinese. A mixed line such as
`OAuth 토큰 발급` is budgeted at the CJK rate for its whole length, because the
full-width glyphs dominate the measured width.

Two-line labels are preferable to a smaller font: `lineHeight` is already
derived, so a second line costs a known amount of vertical space that the
containment check will catch, whereas a per-language font size silently breaks
the type scale everywhere the diagram is reused.

---

## 7. Manual checklist (no Node available)

Walk this by reading the SVG source. Report the outcome as a **manual check** —
never as a lint result, and never with a diagnostic code from the script.

**Structure**

- [ ] Root element is `<svg>` and carries a four-number `viewBox`.
- [ ] If `width`/`height` are present, `width/height` equals the `viewBox`
      aspect ratio.
- [ ] Every opened tag is closed or self-closed; the nesting is balanced.

**References**

- [ ] Every `id` value appears exactly once.
- [ ] Every `url(#name)` and `href="#name"` resolves to a declared `id`.
- [ ] Every `marker-start` / `marker-mid` / `marker-end` target exists.

**Markers**

- [ ] Each `<marker>` declares `markerWidth`, `markerHeight`, `refX`, `refY`.
- [ ] Each `<marker>` declares `markerUnits` explicitly.
- [ ] `refX` matches `markerWidth` where the tip should land on the endpoint.

**Layout**

- [ ] Every box in the table satisfies the containment inequalities.
- [ ] No element extends beyond the `viewBox` on any side.
- [ ] Centers, baselines, and endpoints trace to a formula, not to a literal.

**Accessibility**

- [ ] Root `<svg>` carries `role="img"` (or `aria-hidden="true"` if decorative).
- [ ] `aria-labelledby` names ids that exist in the file.
- [ ] `<title>` is the first child of `<svg>`, before `<defs>`.
- [ ] `<desc>` is present and states the subject, not the shapes.
- [ ] Neither id is the bare `title` / `desc`; both are prefixed per diagram.

**Connectors**

- [ ] All six rules of section 2.5 hold — bends rounded at `r = 8`, label masks
      6-10 units clear of their stroke and clear of later-painted nodes, no two
      connectors sharing a path or an attach point, and any transit behind a
      non-endpoint box dashed and unmarked.

**Text**

- [ ] Font stack is CJK-first.
- [ ] Every line is within its computed capacity for its script.
- [ ] No label was truncated; no per-language font size appears.

Record which items were checked and which could not be determined by reading
alone. An unchecked item is a gap, not a pass.
---

## 8. Accessible SVG contract

An SVG carries no accessible name of its own. Without the four elements below,
assistive technology announces the diagram as an unlabelled graphic — every
label inside it is inert, because screen readers do not read `<text>` nodes out
of a graphic they cannot name. This is a correctness gap, not a polish item, and
it applies to every diagram this skill emits.

### 8.1 The skeleton

Copy this and replace `<slug>` with a token unique to the diagram — the output
filename without its extension is the obvious choice (`auth-flow`,
`auth-flow-dark`):

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1200 720"
     role="img" aria-labelledby="<slug>-title <slug>-desc">
  <title id="<slug>-title">Authentication token flow</title>
  <desc id="<slug>-desc">Three services exchange a short-lived access token: the
    client requests one from the auth service, which validates the credential
    against the user store before the client calls the API with it.</desc>
  <defs>…</defs>
  …
</svg>
```

Four constraints hold it together:

1. **`role="img"` on the root.** Without it the element has no accessible role
   and `aria-labelledby` has nothing to name.
2. **`<title>` is the first child**, before `<defs>` and before any drawn
   element. A title placed later may be ignored.
3. **IDs are prefixed per diagram.** Bare `title` / `desc` IDs are banned: two
   diagrams inlined into one host document would collide, and the second would
   be announced with the first one's name. `aria-labelledby` must name IDs that
   actually exist in the same document.
4. **`<desc>` describes the content, not the geometry.** "Org chart showing a
   command centre routing work to specialist agents" is useful; "a box at the
   top with five boxes below it" is worse than nothing — it costs the reader
   time and conveys no subject.

Keep `<title>` to roughly 60 characters — it is the diagram's name, near enough
to what a heading above it would say. Keep `<desc>` to one sentence stating what
a reader would take away from seeing the image.

### 8.2 Decorative marks

A graphic with no informational content — a specimen glyph sheet, a rule, an
ornament — carries `aria-hidden="true"` on its root instead, and no `<title>`
or `<desc>`. Giving a decorative mark an accessible name adds noise to the
reading order. This is the only exemption; a diagram is never decorative.

### 8.3 Checking it

`scripts/check-svg.mjs` asserts the contract mechanically (`SVG060`–`SVG064`).
Both directions are worth running, because a check that only ever passes proves
nothing:

```bash
node scripts/test-check-svg.mjs   # every fixture, exact code set asserted
```

The runner replaces running the two fixtures by hand: it lints all 42 fixtures —
the accessible-contract pair and the connector-geometry set — and compares each
one's diagnostics against the exact code set that fixture declares, so a check
that stops firing fails just as loudly as one that fires wrongly.

The checker sees structure, not usefulness: a `<desc>` reading "diagram"
satisfies it and tells a reader nothing. Absence is mechanical; vacuity is
yours to catch.

---

## Attribution

The six connector rules (section 2.5), the semantic-role skin and its inversion
rule (section 3), and the accessible-SVG contract (section 8) were adapted from
`cathrynlavery/diagram-design` v2.6.1 (MIT), restated rather than copied. Full
note in `SKILL.md` section Attribution.

