# Design Tokens — html-report

The CSS variable contract (single source of truth) for the html-report skill.
All mode templates in `templates/` MUST consume the tokens defined in this document verbatim.

---

## Palette

Anthropic-inspired palette. Common extracted values across the reference sample set.

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

---

## Font Variable Contract

```css
:root {
  --sans:  "Pretendard", system-ui, -apple-system, "Segoe UI", sans-serif;
  --serif: "Pretendard", ui-serif, Georgia, "Times New Roman", serif;
  --mono:  "JetBrains Mono", ui-monospace, "SF Mono", Menlo, Consolas, monospace;
}
```

Mode-specific override example (plan mode):
```css
:root {
  --serif: "Noto Serif KR", ui-serif, Georgia, serif;
}
```

---

## Layout Tokens

| CSS variable | Value | Use |
|--------------|-------|-----|
| `--max-width` | `860px` | Report container (report / status / financial) |
| `--max-width-wide` | `1120px` | Index / dashboard (index reference only) |
| `--radius-panel` | `12px` | Card / chart panel border-radius |
| `--radius-row` | `8px` | Table row border-radius |
| `--border` | `1.5px solid var(--g300)` | Standard border |
| `--body-lh` | `1.6` | Body line-height |
| `--heading-ls` | `-0.01em` | Heading letter-spacing |

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
  --sans:  "Pretendard", system-ui, -apple-system, sans-serif;
  --serif: "Pretendard", ui-serif, Georgia, serif;
  --mono:  "JetBrains Mono", ui-monospace, "SF Mono", monospace;

  /* Layout */
  --max-width:    860px;
  --radius-panel: 12px;
  --radius-row:   8px;
  --border:       1.5px solid var(--g300);
}
```

---

## Print Tokens (`@media print` Pattern)

Every template MUST include the following `@media print` block.

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

## Accessibility: Contrast Verification Table (WCAG AA ≥ 4.5:1)

| Foreground | Background | Contrast ratio | AA pass |
|------------|------------|----------------|---------|
| `--slate` `#141413` | `--ivory` `#FAF9F5` | 17.9:1 | ✓ |
| `--slate` `#141413` | `--paper` `#FFFFFF` | 18.1:1 | ✓ |
| `--slate` `#141413` | `--g100` `#F0EEE6` | 16.2:1 | ✓ |
| `--g700` `#3D3D3A` | `--ivory` `#FAF9F5` | 8.6:1 | ✓ |
| `--g700` `#3D3D3A` | `--paper` `#FFFFFF` | 8.8:1 | ✓ |
| `--g500` `#87867F` | `--ivory` `#FAF9F5` | 3.7:1 | ✗ (secondary text only, passes AA for large text ≥14px) |
| `--clay` `#D97757` | `--ivory` `#FAF9F5` | 3.2:1 | ✗ (decorative / background use only) |
| `--paper` `#FFFFFF` | `--clay` `#D97757` (background) | 3.2:1 | ✗ (not for text — background accent only) |

**Rules**:
- `--g500` is allowed only for secondary labels (`font-size: 11–12px`) (passes AA for large text)
- `--clay` is used only for accent lines / background accents / link color. Never put white text on a clay background
- Body text (`font-size: 14–16px`) MUST use `--slate` or `--g700`

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
- Text: fill `var(--g500)` (`#87867F`), font-family system-ui (CSS variables not supported inside SVG)

### Carryover Panel
- Background: `--oat`
- Item divider: `rgba(20, 20, 19, 0.08)` (--slate at 8% alpha)
- Tag background: `--ivory`, text: `--g700`
- Body: `color: var(--g700)`
