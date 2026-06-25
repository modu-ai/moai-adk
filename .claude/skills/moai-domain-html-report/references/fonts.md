# Font Policy — html-report

The webfont CDN URLs, licenses, and preconnect patterns for the html-report skill.
This document is the single source of truth for font loading across all mode templates.

---

## Font Policy Rationale

The Thariq philosophy is to block hallucination / runtime-indeterminism risks from JS/CSS CDNs.
Font CDNs do NOT fall under this concern for the following reasons:

1. **No LLM hallucination risk** — font `<link>` tags are static URLs; no class names or APIs
2. **No runtime indeterminism** — fonts are visual assets, not the result of JS execution
3. **Offline fallback** — when a font CDN is unavailable, the system-font fallback stack still renders correctly
4. **Korean necessity** — system fonts alone cause OS inconsistency (macOS / Windows / iOS)

---

## Mode Font Mapping Table

| Mode | sans (body) | serif (heading/emphasis) | mono (code/tag) | CDN source |
|------|-------------|--------------------------|-----------------|------------|
| `status` | Pretendard | Pretendard 700 | JetBrains Mono | jsdelivr + Google |
| `financial` | Pretendard | Pretendard 700 | JetBrains Mono | jsdelivr + Google |
| `pr` | Pretendard | Pretendard 700 | JetBrains Mono | jsdelivr + Google |
| `incident` | Pretendard | Pretendard 700 | JetBrains Mono | jsdelivr + Google |
| `plan` | Pretendard | Noto Serif KR | JetBrains Mono | jsdelivr + Google |
| `explainer` | Noto Sans KR | Noto Serif KR | JetBrains Mono | Google |
| `editorial` | Pretendard | Chosunilbo Myungjo | JetBrains Mono | jsdelivr + noonfonts |
| `legal` | KoPubWorld Batang | KoPubWorld Batang Bold | JetBrains Mono | jsdelivr (noonfonts) |

> The bundled templates cover the first six modes (`status`, `financial`, `pr`, `incident`, `plan`, `explainer`). The `editorial` and `legal` modes are documented for reference and future extension.

---

## CDN URLs and Licenses

### Pretendard
- **License**: OFL-1.1 (SIL Open Font License)
- **CDN**: jsDelivr (GitHub mirror, pinned v1.3.9)
- **URL**: `https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/static/pretendard.css`
- **Included weights**: 100-900 (Variable Font)
- **preconnect host**: `https://cdn.jsdelivr.net`

### Noto Serif KR + Noto Sans KR (Google Fonts)
- **License**: OFL-1.1
- **CDN**: Google Fonts API
- **URL (combined)**: `https://fonts.googleapis.com/css2?family=Noto+Serif+KR:wght@400;700&family=Noto+Sans+KR:wght@400;700&display=swap`
- **preconnect host**: `https://fonts.googleapis.com`, `https://fonts.gstatic.com`

### JetBrains Mono (Google Fonts)
- **License**: Apache 2.0 (compatible with the skill's Apache-2.0 license)
- **CDN**: Google Fonts API
- **URL**: `https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&display=swap`
- **preconnect host**: `https://fonts.googleapis.com`, `https://fonts.gstatic.com`

### Chosunilbo Myungjo
- **License**: Free (personal/commercial use permitted — operated by Chosun Ilbo; for redistribution, refer to the original files directly)
- **CDN (reference)**: `https://cdn.jsdelivr.net/gh/projectnoonnu/noonfonts_2105_2@1.0/Chosunilbomyungjo.woff`
- **Note**: verify current license status before use. Self-hosting is an alternative.
- **preconnect host**: `https://cdn.jsdelivr.net`

### KoPubWorld Batang (Korea Publishing Industry Association)
- **License**: OFL-style (free distribution, personal/commercial use permitted — Korea Publishing Industry Association)
- **CDN (reference)**: `https://cdn.jsdelivr.net/gh/Project-Noonnu/noonfonts_two@1.0/KoPubWorldBatangLight.woff`
- **Weights**: Light (400), Medium (500) — a separate URL is required for Bold
- **Note**: self-hosting directly from the official distributor (https://www.kopus.org/biz-electronic-font2/) is recommended
- **preconnect host**: `https://cdn.jsdelivr.net`

---

## preconnect Patterns

### status / financial / pr / incident modes (Pretendard + JetBrains Mono)

```html
<link rel="preconnect" href="https://cdn.jsdelivr.net" crossorigin>
<link rel="preconnect" href="https://fonts.googleapis.com" crossorigin>
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/static/pretendard.css">
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&display=swap">
```

### plan mode (Pretendard + Noto Serif KR + JetBrains Mono)

```html
<link rel="preconnect" href="https://cdn.jsdelivr.net" crossorigin>
<link rel="preconnect" href="https://fonts.googleapis.com" crossorigin>
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/static/pretendard.css">
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Noto+Serif+KR:wght@400;700&family=JetBrains+Mono:wght@400;500&display=swap">
```

### explainer mode (Noto Sans KR + Noto Serif KR + JetBrains Mono)

```html
<link rel="preconnect" href="https://fonts.googleapis.com" crossorigin>
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Noto+Sans+KR:wght@400;700&family=Noto+Serif+KR:wght@400;700&family=JetBrains+Mono:wght@400;500&display=swap">
```

### editorial mode (Pretendard + Chosunilbo Myungjo + JetBrains Mono)

```html
<link rel="preconnect" href="https://cdn.jsdelivr.net" crossorigin>
<link rel="preconnect" href="https://fonts.googleapis.com" crossorigin>
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/static/pretendard.css">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/projectnoonnu/noonfonts_2105_2@1.0/Chosunilbomyungjo.woff" as="font" type="font/woff" crossorigin>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&display=swap">
```

Chosunilbo Myungjo references the woff file directly. An inline `@font-face` declaration is an alternative:

```css
@font-face {
  font-family: "Chosunilbo Myungjo";
  src: url("https://cdn.jsdelivr.net/gh/projectnoonnu/noonfonts_2105_2@1.0/Chosunilbomyungjo.woff") format("woff");
  font-weight: 400;
  font-style: normal;
  font-display: swap;
}
```

---

## CSS Variable Override Patterns

Per-mode font application is handled by overriding `--sans`, `--serif`, `--mono` in `:root`.

```css
/* status / financial / pr / incident default */
:root {
  --sans:  "Pretendard", system-ui, -apple-system, sans-serif;
  --serif: "Pretendard", ui-serif, Georgia, serif;
  --mono:  "JetBrains Mono", ui-monospace, "SF Mono", monospace;
}

/* plan mode override */
:root {
  --serif: "Noto Serif KR", ui-serif, Georgia, serif;
}

/* explainer mode override */
:root {
  --sans:  "Noto Sans KR", system-ui, sans-serif;
  --serif: "Noto Serif KR", ui-serif, Georgia, serif;
}

/* editorial mode override */
:root {
  --serif: "Chosunilbo Myungjo", "Chosunilbomyungjo", ui-serif, serif;
}

/* legal mode override */
:root {
  --sans:  "KoPubWorld Batang", "KoPubWorld Batang Light", ui-serif, serif;
  --serif: "KoPubWorld Batang", ui-serif, serif;
}
```

---

## Consistent font-display: swap

Always include `&display=swap` in Google Fonts URLs.
Pretendard (jsdelivr CSS) and KoPubWorld (direct woff) use `font-display: swap` via `@font-face`.
