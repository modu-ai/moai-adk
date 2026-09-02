# layout-notes.md — numeric layout passes for the 9 TypePack briefs

Working notes per SPEC-SKILL-GALLERY-BENCH-001 M1 (plan.md §F step 3: "run the
numeric layout pass; pick an archetype"). Each section records the settled
dials, the archetype, the derived grid, the box table, and the containment
check result. Shared conventions across all nine:

- Arrival endpoints follow the fixture convention (`c4-attach-spread.svg`):
  path endpoint = target border − markerLen (10), tip lands at the endpoint.
- marker: `markerUnits="userSpaceOnUse"`, 10x8, refX=10, refY=4, fill `#6B6359`.
- Paint order: background → containers/bands → connectors → connector label
  masks + text → nodes → node text (authoring.md §2.5).
- Type ramp (doc-inline): title 34 serif / layer label 20 / card title 17 /
  body 14 / caption 12. lineHeight = 20 (round(14×1.45)), titleGap = 9
  (round(17×0.55)). Card pad = 16, rx = 12.
- All labels English (SPEC §E.5: SPEC-language artifacts are English); the
  root font stack stays CJK-first per the skill contract.
- Roles: ink `#141413`, ink-muted `#6B6359`, surface `#FAF9F5`,
  surface-alt `#F3EFE6`, border `#D9CDBE`, accent `#D97757`,
  accent-soft `#FBE9DF`, accent-strong `#B85C3E`.

## 1. approval-gate — A2 flow, gate = focal

Dials: svg+png / **fit** (W 1660; 5 stages × preset 292 > doc-inline 1200 —
recorded size deviation) / balanced / mixed.

```
W = 200 + 5*(220+72) = 1660   H = 470 (preset 420 + 50: dashed return path
                                  routed at y=400 below captions 313..357,
                                  needs 400 <= H-M)
stageX(i) = 60 + i*292 → 60, 352, 644, 936, 1228     stageY = 125
stageW = 220, stageH = 170, cy = 210
caption y = 313, h = 44 (A2 caption row)
```

| id | x | y | w | h |
|----|---|---|---|---|
| stage-1 PR Draft | 60 | 125 | 220 | 170 |
| stage-2 CI Matrix | 352 | 125 | 220 | 170 |
| stage-3 Code Review | 644 | 125 | 220 | 170 |
| stage-4 Maintainer Approval (GATE, accent) | 936 | 125 | 220 | 170 |
| stage-5 Merge | 1228 | 125 | 220 | 170 |
| caption-i | stageX(i) | 313 | 220 | 44 |

Connectors: 4 horizontal at y=210, `(x+220,210) → (next.x−10,210)`.
Return (dashed 4,3, caution): `M 754 295 V 392 Q 754 400 746 400 H 178 Q 170 400 170 392 V 305`
(arrival 10 below stage-1 bottom border 295). Label "changes requested",
mask 397..527 × 374..392 (gap 8 above stroke 400; captions end 357 — clear).
Containment: stage5 right 1448 ≤ 1600 ✓; captions bottom 357 ≤ 410 ✓.
Nodes 10 (5 stages + 5 captions) ≤ 12 ✓; connectors 5 ≤ 12 ✓; accent 1 ✓.

## 2. before-after — A3 comparison, 2 cols × 3 concern rows

Dials: svg+png / doc-inline (W 1200) / balanced / mixed.

```
W = 1200 (dial overrides A3 preset 1100)
M = 60, labelW = 260, G = 28, cols = 2
colW = (1200-120-260-56)/2 = 382
colX(0) = 348, colX(1) = 758    (colX(i) = M+labelW+G+i*(colW+G))
title baseline 96 (34 serif) → header band pushed below it:
heads y = 120, h = 64 (bottom 184)
rowY(r) = 200 + r*76 → 200, 276, 352     rowH = 76
H = 352 + 76 + M = 490
```

| id | x | y | w | h |
|----|---|---|---|---|
| head-0 BEFORE | 348 | 120 | 382 | 64 |
| head-1 AFTER | 758 | 120 | 382 | 64 |
| row-r.label | 60 | rowY(r) | 260 | 76 |
| row-r.cell-0 | 348 | rowY(r) | 382 | 76 |
| row-r.cell-1 | 758 | rowY(r) | 382 | 76 |

Zebra bands (surface-alt) behind rows 0 and 2, full width 60..1140.
Containment: col1 right 1140 = W−M ✓; row2 bottom 428 ≤ 430 (H−M) ✓.
Note: two containment corrections applied during the numeric pass — (a) the
A3 preset's own H formula (180+rows·rowH = 408) fails its containment check
(bottom 398 > H−M 348); (b) the preset's header position (M+30 = 90) collides
with the 34px diagram title band (glyphs to y≈104). As-built: headers 120,
rows 200/276/352, H 490.
Nodes 11 ≤ 12 ✓; no connectors ✓. Deviation: comparison grid, not mirrored
picture panels.

## 3. cards-kpi-grid — A1 single layer, 4 stat cards

Dials: svg+png / doc-inline (W 1200) / balanced / mixed.

```
W = 1200, H = 220 + 1*(150+28) = 398, M = 60, G = 28, cols = 4
laneW = (1200-120-3*28)/4 = 249   (≥ A1 min 180 at cols ≤ 4 ✓)
laneX(i) = 60 + i*277 → 60, 337, 614, 891
layerY(0) = 120, layerH = 150
cards: (laneX, 136, 249, 118)     pad 16
```

Stat cards (illustrative values, disclosed in `<desc>`):
card-1 Cards completed **128** (accent focal) · card-2 PR merge rate **91%**
· card-3 CI pass rate **96%** · card-4 Review turnaround **4.2 h**.
Value baseline = innerTop + 34 = 186 (34 serif); label baseline 219 (14).
Layer label "FACTORY LANE" mono at (76, 110).
Containment: card-4 right 1140 = W−M ✓; cards bottom 254 ≤ 338 ✓.
Nodes 4 ✓; no connectors ✓. Deviation: stat tiles as single-layer A1 cards.

## 4. decision-matrix — A3 comparison, 3 cols × 4 criteria rows

Dials: svg+png / doc-inline (W 1200) / **faithful, banded** (recorded detail
deviation: 3 heads + 4 labels + 12 cells = 19 nodes > balanced ceiling 12;
zebra row bands are the labelled bands faithful mode requires) / mixed.

```
W = 1200, M = 60, labelW = 260, G = 28, cols = 3
colW = (1200-120-260-84)/3 = 245.33
colX(i) = 348 + i*273.33 → 348, 621.33, 894.67
title band correction (as form 2): heads y = 120, h = 64
rowY(r) = 200 + r*76 → 200, 276, 352, 428    rowH = 76
H = 428 + 76 + M = 564
```

Columns PostgreSQL / MongoDB / Redis; rows Data model / Scaling / Ops cost /
Familiarity. Cell text budget at 14px: u = 245.33−28 = 217.33 → 25 Latin
chars; longest cell "vertical + replicas" (19) ✓.
Containment: col2 right 1140 = W−M ✓; row3 bottom 504 ≤ 504 (H−M) ✓ (exact).
Deviation: banded faithful grid (dial change forced by the 12-cell brief).

## 5. layer-stack — A1 architecture stack (direct match)

Dials: svg+png / doc-inline (W 1200 = A1 preset) / balanced / mixed.

```
W = 1200, M = 60, layerH = 150, G = 28
title-band correction: layer-1 label (layerY−10) collides with the 34px title,
so layers shift +30: layerY(j) = 150 + j*178 → 150, 328, 506, 684
H = 220 + 4*178 + 30 = 962
L1/L2/L4 cols = 1 → laneW = 1080, cards (60, layerY+16, 1080, 118)
L3 cols = 3 → laneW = (1080-56)/3 = 341.33
L3 laneX = 60, 429.33, 798.67   cards (laneX, 522, 341.33, 118)
```

L1 cmd/moai · L2 internal/cli · L3 config / hook / spec · L4 internal/template.
Connectors (band-to-band per the A1 rule, arrival = border − 10):
L1→L2 (600,300)→(600,318); L2→L3 three verticals at x = 230.67, 600, 969.33
(departures spread on the L2 band bottom edge, 369 apart ≥ 12 ✓) arriving at
the L3 band top −10; L3→L4 three verticals in the same lanes (656→674).
Containment: cards inside bands pad 16 ✓; L4 band bottom 834 ≤ 902 (H−M) ✓.
Nodes 6 ✓; connectors 7 ≤ 12 ✓. Deviation: none (direct archetype match).

## 6. nested-scope — A1 nested containers (containment chain)

Dials: svg+png / doc-inline (W 1200) / balanced / mixed.

```
W = 1200, H = 830, M = 60
L1 organization band  (60, 130, 1080, 640)
L2 lead session band  (100, 190, 1000, 540)   inset 40 / 60
L3 lane worktree band (140, 250, 920, 440)    inset 40 / 60
L4 evidence card      (180, 310, 840, 160)    inset 40 / 60
```

Band labels (mono 12, x+16, y+24): organization / kanban lead session /
lane worktree; L4 card carries the evidence path in mono.
Containment: parent.x+pad ≤ child.x and child.x+child.w ≤ parent.x+parent.w−pad
at every level (pad ≥ 16, actual inset 40) ✓; L1 bottom 770 = H−M ✓ (exact).
Nodes 4 ✓; no connectors ✓; accent = L4 evidence card (focal).
Deviation: nested rectangles instead of concentric rings (equivalent form
explicitly anticipated by spec.md §A).

## 7. process-flow — A2 flow with numbered badges, sync gate marked

Dials: svg+png / **fit** (W 1660; same 5-stage size deviation as form 1) /
balanced / mixed.

```
W = 1660, H = 420 (no return path this form)   stageY = 125
stageX(i) = 60 + i*292 → 60, 352, 644, 936, 1228
stageW = 220, stageH = 170
badges: 28x28 at (stageX+16, 141), numbered 1..5
captions (stageX, 313, 220, 44)
```

Stages backlog → plan → run → sync (★ SYNC GATE eyebrow + accent focal) → done.
Connectors: 4 horizontal at y=210.
Containment: stage5 right 1448 ≤ 1600 ✓; captions bottom 357 ≤ 360 ✓ (tight
but inside; caption text baseline 331 keeps glyphs above 344).
Nodes 10 ✓; connectors 4 ✓; accent 1 ✓. Deviation: numbered lifecycle stages;
gate marked by accent + eyebrow, no dedicated diamond glyph.

## 8. roadmap-timeline — A2 stages bound to a time axis

Dials: svg+png / doc-inline (W 1200; 3 stages × 220 + 2 × 72 = 804 content,
centered at stageX0 = (1200−804)/2 = 198) / balanced / mixed.

```
W = 1200, H = 450, M = 60
stageX = 198, 490, 782   stageY = 125, stageW = 220, stageH = 170
axis y = 400, line 60 → 1160 with arrow marker at right end
milestone dots (r=6) at (308, 400), (600, 400), (892, 400)
tick connectors: (cx, 295) → (cx, 394) straight vertical, thin
dot labels (mono 12) baseline 420 under each dot
```

Phases PLAN / RUN / SYNC with two body lines each; dots carry milestone
labels (SPEC-audit PASS / ACs verified / 3-phase close); axis right-end label
"Epic timeline". Containment: stage bodies inside stage boxes; dot labels
baseline 420 ≤ 450 ✓. Nodes 3 ✓; connectors 3 ticks + 1 axis ≤ 12 ✓;
accent = SYNC stage (close) ✓. Deviation: phase cards + milestone ticks on an
axis line; no proportional duration (brief pins none).

## 9. topology-component — A1 stack, fan-out at the backend layer

Dials: svg+png / doc-inline (W 1200) / balanced / mixed.

```
W = 1200, M = 60, layerH = 150, G = 28
title-band correction as form 5: layerY(j) = 150 + j*178 → 150, 328, 506, 684
H = 962
L1/L2/L3 cols = 1 → cards (60, layerY+16, 1080, 118)
L4 cols = 2 → laneW = (1080-28)/2 = 526; laneX = 60, 614
L4 cards (laneX, 700, 526, 118); cx = 323, 873
```

L1 browser · L2 web console + statusline · L3 moai CLI (accent focal) ·
L4 tmux lanes + MCP broker.
Connectors (band-to-band, arrival = border − 10): L1→L2 and L2→L3 vertical
center (600,300→318 / 478→496); L3→L4 two verticals at x = 323, 873
(656→674; departures 250 apart on the L3 band bottom edge ✓).
Containment: L4 cards inside band pad 16 ✓; band bottom 834 ≤ 902 ✓.
Nodes 5 ✓; connectors 4 ✓; accent 1 ✓. Deviation: hub-and-spoke rendered as
a vertical stack with two-lane fan-out at the bottom layer.

---

## Appendix — check-svg.mjs calibration observed during the run

Two authoring idioms determine whether the lint's connector-geometry tier
sees a diagram's structure at all (read from `scripts/check-svg.mjs`
lines 853-919, confirmed by the first approval-gate run which reported the
background rect as a mask):

1. **`fill="none"` must sit on the `<path>` element itself.** A `<g
   fill="none">` wrapper sets the rendered geometry but the checker reads
   `el.attrs.fill` per path, so group-inherited paths are not counted as
   connectors and the C2/C4/C6 machinery goes silent for them.
2. **A `<rect>` whose immediately-following element sibling is a `<text>` is
   read as a label mask** (unless a connector endpoint attaches within 16
   units). A background rect directly followed by the diagram title therefore
   lints as a mask that "clears its connector by 0.0 units". Wrapping the
   title (and container labels) in a `<g>` restores the intended reading.

All nine as-built artifacts apply both idioms; the two diagrams that rely on
rect+text card interiors keep them unattached-to-connectors (harmless: no
connector, or disjoint from later rects).
