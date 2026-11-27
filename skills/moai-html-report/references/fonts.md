# Font Policy — html-report

The webfont CDN URLs, licenses and preconnect patterns of the html-report
skill. This page is the source for font loading in all six mode templates.

The html files use Pretendard (jsDelivr) and JetBrains Mono (Google Fonts) in
every mode, headings included; the artifact version uses the Noto Sans family of
the report's language from Google Fonts (`artifact-contract.md` § 6).

---

## Font Policy Rationale

Font links are the one kind of external stylesheet the html file loads:

1. **No guessed APIs** — a font `<link>` is a static URL; there are no class
   names or script APIs to get wrong.
2. **No runtime differences** — fonts are visual assets, not the result of
   running a script.
3. **Offline fallback** — when a font CDN cannot be reached, the system-font
   stack in each token still renders the page.
4. **Korean consistency** — system fonts alone differ by OS (macOS Apple SD
   Gothic Neo, Windows Malgun Gothic, iOS), so the html file loads one family.

---

## Mode Font Mapping Table (html files)

| Mode | sans (body) | serif (headings) | mono (code/tag) | CDN source |
|------|-------------|------------------|-----------------|------------|
| `status` | Pretendard | Pretendard 600–700 | JetBrains Mono | jsDelivr + Google |
| `incident` | Pretendard | Pretendard 700 | JetBrains Mono | jsDelivr + Google |
| `plan` | Pretendard | Pretendard 700 | JetBrains Mono | jsDelivr + Google |
| `explainer` | Pretendard | Pretendard 700 | JetBrains Mono | jsDelivr + Google |
| `financial` | Pretendard | Pretendard 700 | JetBrains Mono | jsDelivr + Google |
| `pr` | Pretendard | Pretendard 700 | JetBrains Mono | jsDelivr + Google |

The `--serif` token keeps its name for the heading role; in html files it
holds Pretendard like `--sans`.

---

## CDN URLs and Licenses

### Pretendard
- **License**: OFL-1.1 (SIL Open Font License)
- **CDN**: jsDelivr (GitHub mirror, pinned v1.3.9)
- **URL**: `https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/static/pretendard.css`
- **Build**: the static build — one `@font-face` and one font file per weight
  (100–900), each holding the whole character set; a browser downloads only
  the weights the page uses. It is not the variable font.
- **preconnect host**: `https://cdn.jsdelivr.net`

### JetBrains Mono (Google Fonts)
- **License**: OFL-1.1
- **CDN**: Google Fonts API
- **URL**: `https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&display=swap`
- **preconnect hosts**: `https://fonts.googleapis.com`, `https://fonts.gstatic.com`

---

## preconnect Pattern (all six modes)

```html
<link rel="preconnect" href="https://cdn.jsdelivr.net" crossorigin>
<link rel="preconnect" href="https://fonts.googleapis.com" crossorigin>
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/static/pretendard.css">
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&display=swap">
```

---

## CSS Variables (all six modes)

```css
:root {
  --sans:  "Pretendard", system-ui, -apple-system, "Segoe UI", sans-serif;
  --serif: "Pretendard", ui-serif, Georgia, serif;
  --mono:  "JetBrains Mono", ui-monospace, "SF Mono", Menlo, Consolas, monospace;
}
```

---

## Consistent font-display: swap

Always include `&display=swap` in Google Fonts URLs. The Pretendard
stylesheet sets `font-display: swap` in its own `@font-face` rules.

---

## Artifact Version Fonts

The artifact page takes stylesheets only from `fonts.googleapis.com` and font
files only from `fonts.gstatic.com`, and Pretendard is not on Google Fonts, so
the Pretendard link above never ships in an artifact. A `data:` URI font is
not used either: a full Korean family is megabytes against the artifact's
16 MB limit, which `data:` URIs count toward.

| Role | Family | Source |
|------|--------|--------|
| sans and serif, every mode | Noto Sans of the report's language: Noto Sans KR, Noto Sans JP, Noto Sans SC, Noto Sans TC, or Noto Sans for any other language | Google Fonts |
| mono, every mode | JetBrains Mono | Google Fonts |

Korean example link set:

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Noto+Sans+KR:wght@400;700&family=JetBrains+Mono:wght@400;500&display=swap">
```

The family parameter for each language, the token block and how the artifact
version is built from the html file are in
[`artifact-contract.md`](artifact-contract.md) § 1 and § 6.
