---
name: hns-oss-docs-structure-map
description: >
  docs-site structure map for the oss-docs harness structure-curator: exact
  paths and schemas for hugo.toml, per-locale content/<locale>/_meta.yaml,
  data/menu/main.yaml (4-locale name maps + icon values), the icon-to-SVG-case
  coupling in layouts/partials/menu.html, shortcodes, the FROZEN
  moai-brand.css, vercel.json redirect examples, and the known
  design-vs-guides divergence. Loaded by the structure-curator before any
  navigation or config edit.
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.0.0"
  category: "harness"
  status: "active"
  updated: "2026-07-13"
  tags: "oss-docs,docs-site,hugo,geekdoc,menu,redirects,structure"
---

# docs-site Structure Map

Hugo **geekdoc** site at `docs-site/`, deployed to **adk.mo.ai.kr** via
Vercel (auto-deploy on push). All paths below are relative to `docs-site/`.

## Path map

| Surface | Path | Notes |
|---------|------|-------|
| Site config | `hugo.toml` | NOT hugo.yaml. `defaultContentLanguage = "ko"`; version SSOT `params.version` / `params.releaseDate` |
| Content | `content/{ko,en,ja,zh}/` | ko canonical; 4 locale trees mirror each other |
| Section order | `content/<locale>/_meta.yaml` | per-locale; a section change lands in ALL FOUR files |
| Sidebar menu | `data/menu/main.yaml` | 4-locale name maps + `icon:` per entry |
| Menu icons | `layouts/partials/menu.html` | SVG `switch`/case per `icon:` value — coupling below |
| Shortcodes | `layouts/shortcodes/` | `icon.html` (variants ok/warn/danger/primary/muted), etc. |
| CSS | `static/moai-brand.css` (**FROZEN** — never edit), `static/moai-design.css` | Claude Warm Editorial, light-only theme |
| Redirects | `vercel.json` | `redirects` array — the only Vercel surface this harness touches |

## main.yaml entry schema

Each sidebar entry carries a 4-locale name map and an icon:

```yaml
- name:
    ko: 시작하기
    en: Getting Started
    ja: はじめに
    zh: 快速开始
  ref: /getting-started
  icon: rocket
```

## icon ↔ menu.html SVG-case coupling [HARD]

Every `icon:` value in `main.yaml` MUST have a matching case in the SVG
switch inside `layouts/partials/menu.html`. An unmatched value renders an
empty `<svg>` — a silent visual defect (no build warning). After any icon
edit:

```bash
grep -n '"<icon-value>"' docs-site/layouts/partials/menu.html
```

If absent, add the SVG path case in the same change.

## vercel.json redirect pattern

Moved/renamed pages require BOTH entries:

```json
{
  "redirects": [
    { "source": "/:locale(ko|en|ja|zh)/old-path", "destination": "/:locale/new-path" },
    { "source": "/old-path", "destination": "/ko/new-path" }
  ]
}
```

The Vercel project binding itself is immutable — this harness edits only the
`redirects` array.

## Known divergence to reconcile

`content/<locale>/_meta.yaml` carries a **`design`** section while
`data/menu/main.yaml` carries **`guides`**. When touching either file,
reconcile toward the SSOT design report's 12→11 section restructure
(`.moai/reports/readme-docs-redesign-20260713.md`) and record the resolution
direction in your report.

## Tooling reality

- `gen_menu.py` (referenced by the legacy i18n rules doc) DOES NOT exist —
  menu edits are manual; use the coupling grep above.
- Build check: `cd docs-site && hugo --minify --gc` must complete
  warning-free (a malformed `_meta.yaml` or menu entry surfaces here).
  Full recipe: Skill("hns-oss-docs-verify").
