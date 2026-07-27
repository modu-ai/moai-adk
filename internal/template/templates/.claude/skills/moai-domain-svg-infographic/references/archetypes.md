# Archetype Skeletons

Four layout archetypes cover nearly every technical infographic request. Each
entry gives a canvas preset, the grid parameters, the derived box table, and the
containment rules specific to that shape. Pick one, instantiate the table, then
return to the skill body's authoring step.

All four share one convention: `W` and `H` are the `viewBox` extents, `M` is the
outer margin, `G` is the gutter between siblings, and `pad` is the inner padding
of a card. Nothing below is a pixel measurement of a screen — the `viewBox` is a
unit space, and the renderer scales it.

---

## A1 — Architecture stack

Layers drawn top to bottom, each layer holding one or more components. Reads as
"what sits on top of what".

**Canvas preset**

```
W = 1200
H = 220 + layers * (layerH + G)     layerH = 150, G = 28, M = 60
```

**Grid, per layer**

```
laneW    = (W - 2*M - (cols-1)*G) / cols
laneX(i) = M + i * (laneW + G)
layerY(j)= M + 60 + j * (layerH + G)      60 reserves the title band
```

**Box table shape**

| id | x | y | w | h |
|----|---|---|---|---|
| `layer-j` | `M` | `layerY(j)` | `W - 2*M` | `layerH` |
| `layer-j.card-i` | `laneX(i)` | `layerY(j) + pad` | `laneW` | `layerH - 2*pad` |

**Archetype-specific containment**

- Every card of layer `j` shares one `y` and one `h`. A card that is taller than
  its siblings means the layer height is wrong, not that one card is special.
- Minimum `laneW` is 180 at `cols <= 4`, 150 at `cols = 5`. Below that, split
  the layer into two rows rather than shrinking further.
- The layer band is a container: its label sits at `(M + pad, layerY(j) - 10)`,
  outside the band, so it never competes with card text.

**Connectors** — layer-to-layer arrows are vertical and centered on the shared
lane: `(laneX(i) + laneW/2, layerY(j) + layerH)` down to
`(laneX(i) + laneW/2, layerY(j+1) - markerLen)`.

---

## A2 — Left-to-right flow

A pipeline of stages with arrows between them. Reads as "what happens in what
order".

**Canvas preset**

```
W = 200 + stages * (stageW + arrowGap)    stageW = 220, arrowGap = 72
H = 420
M = 60
```

Cap `stages` at 6. A seventh stage is a signal to collapse two into a labelled
sub-flow, not to shrink the stage width.

**Grid**

```
stageX(i) = M + i * (stageW + arrowGap)
stageY    = (H - stageH) / 2              stageH = 170, vertically centered
```

**Box table shape**

| id | x | y | w | h |
|----|---|---|---|---|
| `stage-i` | `stageX(i)` | `stageY` | `stageW` | `stageH` |
| `stage-i.badge` | `stageX(i) + pad` | `stageY + pad` | `28` | `28` |
| `stage-i.caption` | `stageX(i)` | `stageY + stageH + 18` | `stageW` | `44` |

**Archetype-specific containment**

- Captions sit below the stage box and may extend to `stageW`, but the last
  caption must still satisfy `stageX(last) + stageW <= W - M`.
- Branch labels ride above the arrow at `(arrowMidX, stageY - 14)` and are
  budgeted against `arrowGap`, not against `stageW` — this is the line that
  overflows most often in CJK.

**Connectors** — horizontal, at the shared vertical center:
`(stageX(i) + stageW, stageY + stageH/2)` to
`(stageX(i+1) - markerLen, stageY + stageH/2)`.

---

## A3 — Side-by-side comparison

Two or three columns compared across a shared set of rows. Reads as "how these
options differ".

**Canvas preset**

```
W = 1100
H = 180 + rows * rowH                     rowH = 76
M = 60
```

**Grid**

```
labelW   = 260                            the leftmost row-label column
colW     = (W - 2*M - labelW - cols*G) / cols
colX(i)  = M + labelW + G + i * (colW + G)
rowY(r)  = M + 110 + r * rowH             110 reserves the column headers
```

**Box table shape**

| id | x | y | w | h |
|----|---|---|---|---|
| `head-i` | `colX(i)` | `M + 30` | `colW` | `64` |
| `row-r.label` | `M` | `rowY(r)` | `labelW` | `rowH` |
| `row-r.cell-i` | `colX(i)` | `rowY(r)` | `colW` | `rowH` |

**Archetype-specific containment**

- Cell content is single-line by contract. A cell needing two lines means `rowH`
  is too small for the whole table — raise `rowH` globally, never for one row.
- Zebra banding is drawn as a full-width rect from `M` to `W - M` behind the
  row, and must be emitted before the cells so it never covers text.

**Connectors** — none. A comparison with arrows is really a flow (A2).

---

## A4 — Hierarchy tree

A root with descending levels. Reads as "what belongs to what".

**Canvas preset**

```
W = 1200
H = 160 + levels * (nodeH + vGap)         nodeH = 92, vGap = 64
M = 60
```

**Grid** — each level is centered as a whole, so a level with fewer nodes stays
visually balanced under its parent:

```
levelW(L)   = count(L) * nodeW + (count(L)-1) * G      nodeW = 200
levelX0(L)  = (W - levelW(L)) / 2
nodeX(L, i) = levelX0(L) + i * (nodeW + G)
nodeY(L)    = M + 60 + L * (nodeH + vGap)
```

**Box table shape**

| id | x | y | w | h |
|----|---|---|---|---|
| `n-L-i` | `nodeX(L, i)` | `nodeY(L)` | `nodeW` | `nodeH` |

**Archetype-specific containment**

- `levelX0(L) >= M` must hold for every level. A negative value means the widest
  level exceeds the canvas: reduce `nodeW`, reduce `G`, or split that level
  across two rows with a shared parent stub.
- A child does not need to sit under its parent horizontally, but the elbow
  connector must not cross a sibling node. Check every elbow's `midY` band
  against the bounding boxes on that band.

**Connectors** — elbow, parent bottom to child top:

```
midY = nodeY(L) + nodeH + vGap/2
path: M parent.cx  nodeY(L)+nodeH
      V midY
      H child.cx
      V nodeY(L+1) - markerLen
```

---

## Choosing among the four

| The message is | Archetype |
|----------------|-----------|
| "these components are layered" | A1 architecture stack |
| "this happens, then that" | A2 left-to-right flow |
| "option A versus option B" | A3 side-by-side comparison |
| "this contains these" | A4 hierarchy tree |

A request that resists all four is usually two diagrams. Split it rather than
inventing a fifth shape: a composite layout has no containment rules, and
without containment rules the numeric layout pass cannot be checked.
