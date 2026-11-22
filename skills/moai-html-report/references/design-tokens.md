# Design Tokens — html-report

The CSS variable contract of the html-report skill. All mode templates in
`templates/` use the tokens defined here. The warm palette stays by user
decision (see `craft-fundamentals.md`).

---

## Palette

| CSS variable | Hex | Use |
|--------------|-----|-----|
| `--ivory` | `#FAF9F5` | Page background (warm off-white) |
| `--paper` | `#FFFFFF` | Card / panel / table background |
| `--slate` | `#141413` | Body text (warm near-black) |
| `--clay` | `#D97757` | Accent / link / clay accent (terracotta) |
| `--clay-d` | `#B85C3E` | clay hover / high-risk indicator |
| `--oat` | `#E3DACC` | Secondary background / divider / default chart bar |
| `--olive` | `#788C5D` | Positive signal / secondary accent (sage green) |

### Grayscale

| CSS variable | Hex | Use |
|--------------|-----|-----|
| `--g100` | `#F0EEE6` | Table header background / light surface |
| `--g300` | `#D1CFC5` | Border / divider |
| `--g500` | `#87867F` | Secondary text / label |
| `--g700` | `#3D3D3A` | Mid-intensity text / author name |

### Dark values

Each template carries the two dark-mode blocks (`artifact-contract.md` § 3)
with these values; both blocks set `color-scheme: dark`.

| CSS variable | Dark hex |
|--------------|----------|
| `--ivory` | `#17150F` |
| `--paper` | `#201E19` |
| `--slate` | `#F0EEE6` |
| `--clay` | `#E08B6D` |
| `--clay-d` | `#EC9F82` |
| `--oat` | `#2E2B24` |
| `--olive` | `#97A87B` |
| `--g100` | `#26241E` |
| `--g300` | `#3D3B33` |
| `--g500` | `#87867F` |
| `--g700` | `#C4C2B8` |

---

## Font Variable Contract

Every mode uses the same three font tokens. In html files they hold
Pretendard and JetBrains Mono; the artifact version swaps `--sans` and
`--serif` to the Noto Sans family of the report's language (`fonts.md`).

```css
:root {
  --sans:  "Pretendard", system-ui, -apple-system, "Segoe UI", sans-serif;
  --serif: "Pretendard", ui-serif, Georgia, serif;
  --mono:  "JetBrains Mono", ui-monospace, "SF Mono", Menlo, Consolas, monospace;
}
```

---

## Layout Tokens

| CSS variable | Value | Use |
|--------------|-------|-----|
| `--max-width` | per mode, see `SKILL.md` § Design tokens and width | Report container |
| `--content-max` | `680px` | Prose column (explainer) |
| `--radius-panel` | `12px` | Card / chart panel border-radius |
| `--radius-row` | `8px` | Table row border-radius |
| `--border` | `1.5px solid var(--g300)` | Standard border |

Body line-height is `1.6` (explainer `1.65`, pr `1.55`); heading
letter-spacing is `-0.01em`. Both are set in each template's rules, not as
tokens.

---

## Full `:root` Declaration (copy-paste baseline)

```css
:root {
  /* Palette */
  --ivory:  #FAF9F5;
  --paper:  #FFFFFF;
  --slate:  #141413;
  --clay:   #D97757;
  --clay-d: #B85C3E;
  --oat:    #E3DACC;
  --olive:  #788C5D;

  /* Grayscale */
  --g100: #F0EEE6;
  --g300: #D1CFC5;
  --g500: #87867F;
  --g700: #3D3D3A;

  /* Fonts */
  --sans:  "Pretendard", system-ui, -apple-system, "Segoe UI", sans-serif;
  --serif: "Pretendard", ui-serif, Georgia, serif;
  --mono:  "JetBrains Mono", ui-monospace, "SF Mono", Menlo, Consolas, monospace;

  /* Layout */
  --max-width:    860px;
  --radius-panel: 12px;
  --radius-row:   8px;
  --border:       1.5px solid var(--g300);
}
```

---

## Print Tokens (`@media print` Pattern)

Every template includes the following `@media print` block.

```css
@media print {
  body {
    background: white;
    color: black;
    padding: 0;
    font-size: 12pt;
  }
  .page {
    max-width: none;
  }
  a[href]::after {
    content: " (" attr(href) ")";
    font-size: 10pt;
    color: #555;
  }
  h1, h2, h3 {
    page-break-after: avoid;
  }
  table, figure, .chart-panel {
    page-break-inside: avoid;
  }
  .no-print {
    display: none !important;
  }
  /* Preserve borders / background colors when printing */
  * {
    -webkit-print-color-adjust: exact;
    print-color-adjust: exact;
  }
}
```

---

## Accessibility: Contrast Table (light theme, WCAG AA)

Ratios computed with the WCAG 2 relative-luminance formula. AA asks 4.5:1 for
body text and 3:1 for large text, which is 24px and up, or 18.66px and up in
bold.

| Foreground | Background | Contrast ratio | AA body text |
|------------|------------|----------------|--------------|
| `--slate` `#141413` | `--ivory` `#FAF9F5` | 17.50:1 | ✓ |
| `--slate` `#141413` | `--paper` `#FFFFFF` | 18.43:1 | ✓ |
| `--slate` `#141413` | `--g100` `#F0EEE6` | 15.87:1 | ✓ |
| `--g700` `#3D3D3A` | `--ivory` `#FAF9F5` | 10.34:1 | ✓ |
| `--g700` `#3D3D3A` | `--paper` `#FFFFFF` | 10.90:1 | ✓ |
| `--g500` `#87867F` | `--ivory` `#FAF9F5` | 3.47:1 | ✗ (large text only) |
| `--g500` `#87867F` | `--paper` `#FFFFFF` | 3.65:1 | ✗ (large text only) |
| `--olive` `#788C5D` | `--ivory` `#FAF9F5` | 3.49:1 | ✗ (large text only) |
| `--clay` `#D97757` | `--ivory` `#FAF9F5` | 2.96:1 | ✗ (decorative only) |
| `--paper` `#FFFFFF` | `--clay` `#D97757` (background) | 3.12:1 | ✗ (large text only) |
| `--paper` `#FFFFFF` | `--clay-d` `#B85C3E` (background) | 4.53:1 | ✓ |

**Rules**:
- `--g500` is for secondary labels only; small body text in `--g500` misses AA.
- `--clay` is for accent lines, background accents and link color. White
  text on a `--clay` background misses AA at body size.
- Body text (`font-size: 14–16px`) uses `--slate` or `--g700`.

Several templates still use `--g500` for small text and white text on
`--clay`; the token values above are unchanged until those are fixed.

---

## Per-Component Color Rules

### Metric Card (stat-card)
- Background: `--paper`
- Border: `var(--border)`
- Warning card: `border-left: 4px solid var(--clay)`
- Number: `font-family: var(--serif); color: var(--slate)`
- Label: `color: var(--g500); text-transform: uppercase`
- Positive delta: `color: var(--olive)`
- Neutral delta: `color: var(--g500)`

### Shipped Table
- Table background: `--paper`
- Header background: `--g100`
- Row divider: `--g100`
- PR link: `color: var(--clay)`
- Author: `color: var(--g700)`

### Risk Dot Color
- Low: `background: var(--olive)`
- Med: `background: var(--clay)`
- High: `background: var(--clay-d)` (`#B85C3E`)

### Velocity SVG Bar
- Normal bar: fill `var(--oat)` (`#E3DACC`)
- Peak bar: fill `var(--clay)` (`#D97757`)
- Gridline (major): stroke `var(--g300)` (`#D1CFC5`)
- Gridline (minor): stroke `var(--g100)` (`#F0EEE6`)
- Text: fill `var(--g500)` (`#87867F`), font-family system-ui (the templates
  set SVG colors as hex attributes today)

### Carryover Panel
- Background: `--oat`
- Item divider: `rgba(20, 20, 19, 0.08)` (--slate at 8% alpha)
- Tag background: `--ivory`, text: `--g700`
- Body: `color: var(--g700)`
