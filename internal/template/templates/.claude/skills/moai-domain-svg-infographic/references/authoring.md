# Authoring Reference

Everything the skill body points at: the full geometry and connector formula
set, the icon set, the palette and type scale, the CJK text-budget worked
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

---

## 3. Palette and type scale

This palette is aligned to the project design system (`moai-domain-html-report`
tokens): warm ivory paper, clay terracotta accent, slate ink. The skill keeps
its own copy of the values (it does not read a runtime token file), so the
diagram renders offline; substitute a project's own tokens freely, but
substitute the whole set so contrast relationships survive.

| Role | Value | Use |
|------|-------|-----|
| ink | `#141413` | Primary text (warm black) |
| ink-muted | `#6B6359` | Captions, secondary text |
| surface | `#FAF9F5` | Canvas (ivory) |
| surface-alt | `#F3EFE6` | Band fill, warm tone |
| border | `#D9CDBE` | Hairline card and divider strokes |
| accent | `#D97757` | Focal / active path (clay terracotta) |
| accent-soft | `#FBE9DF` | Focal fill (clay-tint) |
| accent-strong | `#B85C3E` | Clay hover state, eyebrow text |
| positive | `#1a7f37` | Success, allowed |
| caution | `#9a6700` | Warning, degraded |
| negative | `#cf222e` | Failure, forbidden |

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

**Text**

- [ ] Font stack is CJK-first.
- [ ] Every line is within its computed capacity for its script.
- [ ] No label was truncated; no per-language font size appears.

Record which items were checked and which could not be determined by reading
alone. An unchecked item is a gap, not a pass.
