# Artifact Page Contract

What the artifact version of a report satisfies before it is published as a
Claude Artifact. Modes, audience tiers and the md record stay owned by
`SKILL.md`; this page covers only the hosted page. Load the host's
`artifact-design` skill before the publish: it holds the full page contract,
and this page is its summary for reports.

## 1. No document skeleton

The publish wraps the page in its own document skeleton. The artifact version
therefore has no `<!doctype html>`, `<html>`, `<head>` or `<body>` tags and no
charset or viewport meta. It starts with the title, the viewport-fit meta, the
font links and the style block, followed by what the html file held inside
`<body>`:

```html
<title>Payment 502 Incident</title>
<meta name="claude:viewport-fit-sides" content="cover">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Noto+Sans+KR:wght@400;700&family=JetBrains+Mono:wght@400;500&display=swap">
<style>
  /* the template's whole style block, with the changes below */
</style>
<div class="page">…</div>
```

Building it from the html file:

1. Delete the doctype, the `<html>`, `<head>` and `<body>` tags with their
   closing tags, and the charset and viewport metas.
2. Put `<title>`, the viewport-fit meta, the Google Fonts links (§ 6) and the
   `<style>` block at the top.
3. Delete the Pretendard stylesheet link and the jsDelivr preconnect, and set
   `--sans` and `--serif` to the Noto Sans family of the report's language
   (§ 6).
4. Delete the mermaid script, the script that hides the fallback, and every
   `.mmd-fallback` block; keep each `pre.mermaid` (§ 7).
5. Keep the two dark-mode blocks and any small inline script (the explainer
   tabs).

## 2. Title (two to four words)

The `<title>` is a name of two to four words, never a "Name: subtitle" or
"Name — subtitle" form. Ids, periods and the explanation go into the
one-sentence `description` of the publish and into the page body.

| Wrong | Right |
|---|---|
| `<title>Q3 Report: an explainer of the caching layer</title>` | `<title>Caching Layer</title>` |
| `<title>INC-2026-0422 — Payment 502</title>` | `<title>Payment 502 Incident</title>` |

## 3. Color tokens and the two dark blocks

Colors are tokens on `:root`. Dark mode comes through both blocks below, a
media block guarded against an explicit light choice and an attribute block,
and each block sets `color-scheme: dark`: the publish skeleton fixes the light
scheme on `:root`, so without it scrollbars, `<details>` markers and form
controls stay light. `body` has an explicit background in both schemes.

```css
:root {
  --ivory: #FAF9F5;  /* page background */
  --slate: #141413;  /* body text */
  /* ...the rest of the mode palette... */
}
body { background: var(--ivory); color: var(--slate); }

@media (prefers-color-scheme: dark) {
  :root:not([data-theme="light"]) {
    color-scheme: dark;
    --ivory: #17150F;
    --slate: #F0EEE6;
    /* ...the dark value of every color token... */
  }
}
:root[data-theme="dark"] {
  color-scheme: dark;
  --ivory: #17150F;
  --slate: #F0EEE6;
}
```

The six templates already carry both blocks; their dark values are listed in
`design-tokens.md`.

## 4. Layout and size

```css
body {
  padding-block: 24px 96px;
  padding-left: max(16px, env(safe-area-inset-left, 0px));
  padding-right: max(16px, env(safe-area-inset-right, 0px));
}
```

- The side padding is these two `max()` values at every width (a wide page may
  use a larger first value, such as 24px); top and bottom go through
  `padding-block`. The `claude:viewport-fit-sides` meta goes with these two
  side paddings and never without them.
- The page never scrolls sideways. A wide table, code block or diagram scrolls
  inside its own box with `overflow-x: auto`.
- The rendered page is 16 MB or smaller, and embedded `data:` URIs count
  toward that limit.

## 5. External resources

| Kind | Allowed hosts |
|---|---|
| Stylesheets | `fonts.googleapis.com` only |
| Font files | `fonts.gstatic.com` only |
| Scripts | `cdnjs.cloudflare.com`, `cdn.jsdelivr.net/npm/`, `unpkg.com`, `cdn.tailwindcss.com`, `code.jquery.com` |

A script from those hosts names an exact version at least two weeks old (never
a range such as `@11`), is a UMD build loaded by a classic `<script src>`, and
comes before the script that uses it. moai reports use none of them: reports
load no external script except the one mermaid script in html files; small
inline scripts are allowed (the explainer tabs); the artifact version carries
no external script.

## 6. Fonts: Noto Sans of the report's language

The artifact version uses the Noto Sans family of the report's language from
Google Fonts, for body text and headings alike, and JetBrains Mono for code. The
html file keeps Pretendard (`fonts.md`).

| Report language | Family | Google Fonts `family` parameter |
|---|---|---|
| Korean | Noto Sans KR | `family=Noto+Sans+KR:wght@400;700` |
| Japanese | Noto Sans JP | `family=Noto+Sans+JP:wght@400;700` |
| Simplified Chinese | Noto Sans SC | `family=Noto+Sans+SC:wght@400;700` |
| Traditional Chinese | Noto Sans TC | `family=Noto+Sans+TC:wght@400;700` |
| Any other language | Noto Sans | `family=Noto+Sans:wght@400;700` |
| Code, every language | JetBrains Mono | `family=JetBrains+Mono:wght@400;500` |

The link is
`https://fonts.googleapis.com/css2?family=<language family>&family=JetBrains+Mono:wght@400;500&display=swap`,
and the font tokens become:

```css
:root {
  --sans:  "Noto Sans KR", system-ui, -apple-system, sans-serif;
  --serif: "Noto Sans KR", system-ui, sans-serif;
  --mono:  "JetBrains Mono", ui-monospace, "SF Mono", Menlo, Consolas, monospace;
}
```

Why not Pretendard: the artifact page takes stylesheets only from
`fonts.googleapis.com` and font files only from `fonts.gstatic.com`, and
Pretendard is not on Google Fonts. Embedding it as a `data:` URI would spend
megabytes of the 16 MB limit on every report.

## 7. Diagrams: `pre.mermaid` without a script

Structural diagrams stay as `<pre class="mermaid">` blocks and the artifact
viewer renders them. The artifact version has no mermaid script and no
`.mmd-fallback` block (with no script to hide it, it would always show).
Inline SVG charts stay as they are.

## 8. Do not use

- A print button: `window.print()` does not work in the viewer.
- `alert`, `confirm` or `prompt`.
- `<iframe>`, `<object>` or `<embed>`.
- Embedded content from other sites: link to them with `<a href>` only.
- `<a download>`: it does not work.
- `mailto:` links: they are unreliable.
- Hash links other than `#name` to an element of the same page.

## 9. Publish arguments

- `icon`: one generic word on the first publish (for example `report` or
  `chart`); leave it out on a republish.
- `description`: one sentence, shown as the subtitle of the gallery card.
- The artifact is private. moai never shares or deletes it; sharing is the
  user's choice.
- A republish in the same conversation uses the same file path; a later
  conversation passes the earlier `url`, which the md record keeps.
- With a PreToolUse hook on Edit or Write in the session, edit the file first
  and publish it as its own call.

The delivery order (md record, html, one publish, the refusal line) is in
`SKILL.md` § Delivery.
