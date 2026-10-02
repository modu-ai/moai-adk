# Artifact Page Contract

The publication contract a report MUST satisfy before it is delivered as a
Claude Artifact (`report.format=artifact`). Content, modes, audience tiers, and
the markdown twin remain owned by this skill — this page governs only the
**publication contract** of the hosted page (the artifact-design skill owns the
broader design guidance; a report publish step referencing it for the contract
is correct routing, not a miss).

## 1. Standalone document skeleton

The rendered document is a **complete standalone document** — it starts with its
own `<!doctype html>` and carries its own `<meta charset>` and viewport metas
and inline base styles. It does NOT rely on any outer document skeleton,
wrapper, or skeleton inheritance: nothing is assumed to be provided around it.
"Skeleton removal" in the artifact contract means removing wrapper/skeleton
*inheritance assumptions*, never the document's own skeleton tags — the artifact
must render correctly when opened cold, with nothing else on the page.

## 2. Title rule (2-4 words)

The document `<title>` is a **name of two to four words** — never a
"Name: explainer" style title (no `Title: subtitle` / `X: a report about Y`
formats). The title names the thing; explanation goes elsewhere. A one-sentence
`description` accompanies the publication (the subtitle line on the artifact
card) and is separate from the title.

| Wrong | Right |
|---|---|
| `<title>Q3 Report: an explainer of the caching layer for the platform team</title>` | `<title>Caching Layer</title>` |
| `<title>Report: Incident 502</title>` | `<title>Payment 502 Incident</title>` |

## 3. Color tokens on `:root` + dual dark-mode blocks

Define color tokens on `:root`. Provide dark mode through **BOTH** blocks — a
media block guarded against an explicit light override, and an explicit
attribute block:

```css
:root {
  --ivory: #FAF9F5;  /* page background */
  --slate: #141413;  /* body text */
  /* ...rest of the mode palette... */
}

/* Dark scheme A — follow the OS preference unless the page opts out. */
@media (prefers-color-scheme: dark) {
  :root:not([data-theme="light"]) {
    --ivory: #17150F;
    --slate: #F0EEE6;
    /* ...dark counterparts of every color token... */
  }
  :root:not([data-theme="light"]) body {
    background: var(--ivory);
    color: var(--slate);
  }
}

/* Dark scheme B — explicit toggle wins over the media query. */
:root[data-theme="dark"] {
  --ivory: #17150F;
  --slate: #F0EEE6;
}
:root[data-theme="dark"] body {
  background: var(--ivory);
  color: var(--slate);
}
```

`body` carries an **explicit background in both schemes** — an inherited or
default body background is a contract miss (dark-mode pages that flash white
come from exactly this).

## 4. Layout — phone width, 16px gutter, no horizontal scroll

The layout renders correctly at **phone width**: a side gutter of at least
**16px** and **no horizontal page scroll** at any width. Concretely:

```css
@media (max-width: 640px) {
  body { padding: 24px 16px 96px; }        /* 16px side gutter */
  pre, table { max-width: 100%; }          /* blocks shrink; pre may scroll within itself */
  img, svg { max-width: 100%; height: auto; }
}
```

Page-level horizontal scrolling is the failure; a wide `pre` or table may
scroll *within its own box*, but the page itself never gains a horizontal
scrollbar.

## 5. External resources — host limits

| Kind | Allowed hosts |
|---|---|
| Scripts | `cdnjs.cloudflare.com` or `cdn.jsdelivr.net/npm/` only — and for report artifacts, **no scripts at all** is the default (diagrams pre-render; see § 7) |
| Stylesheets | **Google Fonts only** (`fonts.googleapis.com` / `fonts.gstatic.com`) |

Size ceiling: the rendered page stays at **16MB or smaller**, and embedded
`data:` URIs count toward that limit.

## 6. Fonts — Google Fonts Noto families

Artifact-format reports load Korean type **exclusively from Google Fonts**:

| Role | Family |
|---|---|
| body (sans) | Noto Sans KR |
| serif modes (headings) | Noto Serif KR |
| code | JetBrains Mono |

```html
<link rel="preconnect" href="https://fonts.googleapis.com" crossorigin>
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Noto+Sans+KR:wght@400;700&family=JetBrains+Mono:wght@400;500&display=swap">
```

Decision rationale: (1) the contract restricts stylesheet hosts to Google
Fonts, and Pretendard is not in the Google Fonts catalog — the jsdelivr
Pretendard load used by the html-file format would violate the contract here;
(2) an `@font-face` inline with `data:` URIs would consume the 16MB ceiling
(a full Korean family is multi-megabyte); (3) Noto Sans KR / Noto Serif KR /
JetBrains Mono are already families this skill maps, so no new CDN
relationship is introduced. The html-file format's Pretendard jsdelivr load is
unchanged — this rule scopes to `format=artifact` only.

## 7. Diagrams — `pre.mermaid` pre-render

Artifact documents carry **no external mermaid `<script>`**. Structural
diagrams are emitted as `pre.mermaid` code blocks; the artifact viewer
pre-renders them. The html-file format's tier-gated mermaid CDN and
`<noscript>` fallback do not apply to artifact output.

## 8. Delivery semantics (summary)

- `format=artifact` + Artifact tool available → publish the HTML, present the
  **artifact link**, skip the browser auto-open.
- `format=artifact` + Artifact tool unavailable (Codex / GLM / API-key
  sessions) → deliver through the existing html+md path and say so in the
  summary. Never an error.
- Regenerating a report at the same output path and republishing the same file
  path **preserves the artifact URL**; the `.moai/reports/` pair files are
  always written regardless of delivery format.
