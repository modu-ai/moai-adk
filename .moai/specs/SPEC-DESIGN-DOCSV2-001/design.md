# design.md — SPEC-DESIGN-DOCSV2-001

> System-design decisions for the docs-site v2 migration. Each section answers a "how" question raised by spec.md requirements. Decisions most likely to change are placed first.

## §A. Token-Unification Architecture

### §A.1 The question

The docs-site currently stacks 3 token vocabularies (moai-brand.css FROZEN warm-cream + moai-docs-tokens.css Clay/Cream/Ink + moai-docs-theme.css remap). v2 정통 정합 collapses them into one v2 system. **Do we (option 1) rewrite `moai-brand.css` in-place as the single v2 SSOT, or (option 2) keep the layered override architecture and add a v2 override layer?**

### §A.2 Decision: Option 1 — single v2-native token file, layering collapsed

**Chosen**: rewrite `moai-brand.css` `:root` block to the v2 token vocabulary (REQ-TOK-001..009), and fold `moai-docs-tokens.css` + `moai-docs-theme.css` contents into it. The 4-file CSS stack becomes a 2-file stack:

| New file | Owns | Lines (est.) |
|---|---|---|
| `static/moai-brand.css` (rewritten, v2-native, re-FROZEN at sync close) | `:root` v2 tokens + ALL `cw-*`/`gdoc-*` component overrides (formerly split across 4 files) + motion keyframes (formerly in tokens.css) | ~2400 |
| `static/moai-design.css` (rewritten) | macOS code-card chrome + layout shell (`.shell/.nav`/sidebar/hero/rail/cards) consuming v2 tokens | ~1100 |

Removed:
- `static/moai-docs-tokens.css` — Clay/Cream/Ink block deleted; MaruBuri `@font-face` deleted (REQ-TYP-002); motion keyframes folded into `moai-brand.css`.
- `static/moai-docs-theme.css` — remap layer obsolete once `moai-brand.css` is v2-native; `md-*` landing/404 components folded into `moai-design.css`.
- `static/moai-docs-theme.js` — audited; if it only drove the theme remap, removed; if it carries unrelated logic, preserved with the remap code stripped.
- `assets/css/moai-brand.scss` — already stale/uncompiled; delete as cleanup.

### §A.3 Rationale

- The layered override architecture existed ONLY to work around the FROZEN status of `moai-brand.css`. Once unfrozen (spec.md §G), the override layer has no purpose.
- 3 vocabularies = 3 places to change a token = drift risk (already observed: warm-cream vs Clay/Cream/Ink diverge on `--bg-page` and `--ink-900`).
- A single v2-native SSOT makes grep-based token-parity ACs (acceptance.md §B) trivially enforceable.
- Round3 충실 재현 (user decision 2) is cleaner against a single token vocabulary.

### §A.4 Cache-busting impact

`head/custom.html` computes 4 FNV32a hashes today. After collapse, it computes 2 (brand + design). The `?h=` fingerprinting mechanism is unchanged (REQ-BLD-003); only the hash inputs shrink. The custom.html edit is part of M1.

### §A.5 Re-FROZE policy

At sync-phase close, the rewritten `moai-brand.css` is re-stamped FROZEN with the v2 token vocabulary as the new frozen baseline. The FROZEN comment header is restored.

---

## §B. Layout Architecture (round3 → Hugo/geekdoc)

### §B.1 The question

round3 is a static prototype. How do we map the round3 visual slots into Hugo's layout system while preserving the i18n / menu / search infrastructure?

### §B.2 Decision: visual shell replaced, data flow preserved

**Replaced** (visual shell — rewritten to v2 + round3 geometry):

| Current file | Round3 slot mapped | Action |
|---|---|---|
| `layouts/_default/baseof.html` | nav shell + page wrapper + footer shell + dark-mode MutationObserver | Rewrite chrome to round3 `.nav` / `.page` / `.footer` geometry; preserve MutationObserver (REQ-LIT-001); preserve `gdoc-nav→menu partial` + `gdoc-page→main` block hooks; preserve right-rail TOC `aside.cw-toc` (re-styled to round3 `.toc`) |
| `layouts/index.html` | docs-hero + docs-stats + featured + docs-grid (home) | Rewrite to round3 `05-docs-index` slot structure; preserve `hugo.Sites` i18n eyebrow/title/lead; preserve `.md-stat-row` data (may stay hardcoded or move to params) |
| `layouts/_default/single.html` (or list.html) | doc-hero + doc-layout (TOC + body + rail) + read-progress | Rewrite to round3 `06-docs-detail` 3-col grid; adopt DocPage article-shell styling (REQ-LAY-004); add read-progress JS + IntersectionObserver TOC active |
| `layouts/partials/site-header.html` | round3 `.nav` | Rewrite visual to round3 nav geometry; PRESERVE brand mascot img + "MoAI-ADK"/"docs" tag + ⌘K search input + version pill + lang-switch (KO/EN/JA/CN over `hugo.Sites`) + GitHub Star button + ⌘K modal + 24h localStorage star cache + fuzzy `/search.json` search. These are infra (REQ-LAY-003). |
| `layouts/partials/site-footer.html` | round3 `.footer` | Rewrite visual; preserve copyleft + GitHub/MIT + Anthropic/modu-ai attribution |
| `layouts/_default/_markup/render-codeblock.html` | round3 `.code-mac` | Restyle to v2 code-card (macOS chrome, green lang pill, copy button); existing hook preserved |

**Preserved** (infra — NOT touched):

| File | Why preserved |
|---|---|
| `layouts/partials/menu.html` + `menu/name.html` + `menu/href.html` | Sidebar reads `data/menu/main.yaml`; icon→SVG case ladder at lines 23–41 MUST match main.yaml `icon:` values (REQ-LAY-006). Restyle only. |
| `content/<locale>/` + `_meta.yaml` | 4-locale content SSOT |
| `data/menu/main.yaml` | 4-locale nav SSOT |
| `layouts/partials/search.json` + ⌘K modal | Search infra |
| `api/i18n-detect.ts` + `vercel.json` redirects | Edge function + redirects (REQ-BLD-002/004) |
| `layouts/shortcodes/icon.html` + `mascot.html` | icon used 297× across 55 files; mascot shortcode currently unused but is the REQ-MAS-004 vehicle |

### §B.3 TOC + reading-progress wiring

geekdoc already renders a right-rail TOC (`aside.cw-toc` in baseof.html). The round3 `06-docs-detail` places TOC in the LEFT rail (220px) and keeps the right rail (280px) for actions/related/CTA.

**Decision**: adopt round3's 3-col layout (TOC left, body center, rail right). The existing `aside.cw-toc` moves to the left column and re-styles to round3 `.toc-list` (active state = primary border-left). The right rail is a NEW partial `layouts/partials/doc-rail.html` rendering actions + related + CTA. Reading-progress bar + IntersectionObserver are new JS (inline in single.html or a small `static/js/doc-progress.js`).

### §B.4 Featured + docs-grid data source

round3 hardcodes the 4:5 card grid. The docs-site docs-index is Hugo-template-driven:

- **Featured**: a single page pinned by `weight: 0` in `_meta.yaml` OR a new params field `featured_page`. Rendered as the round3 `featured-card` (solid green, 2-col).
- **Card grid**: Hugo section list (`range .Site.RegularPages` filtered by section + locale), rendered as round3 `doc-card` with `doc-thumb` (themed bg per category) + `doc-body`. Category pill counts come from a Hugo taxonomy count.

---

## §C. Component Mapping (v2 recipe → existing class)

| v2 recipe (bundle) | Existing docs-site class | Migration action |
|---|---|---|
| Primary CTA pill + signature | `.md-btn` / `.btn-primary` | Restyle to `padding 14px 24px; radius 9999px; background: var(--gradient-signature) solid #3d7d5f; fw 700; tracking -0.025em`; hover adds `shadow-signature` |
| Secondary outline button | `.btn-secondary` | `background: surface; border 1.5px #b5b5b5; color: ink` |
| Ghost button | `.btn-ghost` | `color: ink; hover background: brightness(0.97)` |
| Surface card | `.md-doc-card` / `.card` | `background: surface; radius 16px; padding 24px; border 1px #d1d1d1; shadow-sm; hover: translateY(-2px) + shadow-md` |
| Elevated card | (variant) | `shadow-lg`, no border |
| Outline card | (variant) | `border 1.5px #b5b5b5`, no shadow |
| Gradient card (accent) | `.featured-card` / `.next-cta` | `background: var(--gradient-signature) solid #3d7d5f; color #fff` |
| Chip / tag | `.doc-tag` / `.chip` | `radius 999px / 4px; fw 700; tracking body` |
| Eyebrow | `.docs-eyebrow` | `font-mono; size 11–12px; tracking 0.12em; uppercase; color primary` |
| Callout | `.callout` | `background: gradient-signature-soft; border-left 4px solid primary; radius 8px` |
| Code card (macOS) | `.code-mac` / `.code-card` | Preserve macOS chrome; re-tint to v2 (traffic lights unchanged, title bar #303030, body #1e1e1e — these are component-local dark, NOT a theme conflict per CLAUDE.local.md §17.1) |
| Empty state (mascot + wit + CTA) | (new) `.md-empty` | Mascot pose + witty copy + CTA button; used on 404 / empty search / section-with-no-pages |

Forbidden patterns (skipped, REQ-CMP-005/007): rounded-border + left-color-accent card; full-bleed image bg; gradient + shadow simultaneous.

---

## §D. Mono-Font Decision (resolves REQ-TYP-003)

**Default**: adopt the round3 two-token split —
- `--font-mono: "JetBrains Mono", ui-monospace, ...` → UI chips, eyebrows, meta strips, code-card TITLES.
- `--font-code: "Goorm Sans Code", "JetBrains Mono", ...` → code-card BODIES (Korean-aware: Korean comments/literals inside code render cleanly).

**Rationale**: round3 충실 재현 (decision 2) + keeps Korean readability inside code blocks + no net-new CDN (Goorm already loaded; JetBrains loaded via the bundle's Google Fonts `@import` which the docs-site adds to `moai-brand.css` `@import url('https://fonts.googleapis.com/css2?...JetBrains+Mono...')`).

**Override path** (one line): if the user prefers v2 정통 정합 purism (single JetBrains Mono, accept Hangul fallback), drop `--font-code` and alias it to `--font-mono`. This is a `design.md §D` toggle, not a SPEC re-open.

**[NEEDS CLARIFICATION: mono-font]** marker in `plan.md` M2 resolves this. Default stands absent explicit override.

---

## §E. Mermaid Version Decision (resolves REQ-MER-002)

**Default**: stay on **mermaid@10**, apply only the themeVariables update (REQ-MER-001).

**Rationale**: the v2 palette shift is a pure themeVariables change (§F of research.md). No v11 feature is used. A v10→v11 bump adds loader-API risk for zero visual gain. A separate follow-up SPEC can bump to v11 if a v11-only feature becomes needed.

**Override path**: if the user wants parity with the bundle (`MermaidDiagram.jsx` uses v11), bump the CDN URL to `mermaid@11` and load-test the `mermaid.run` signature. This is a `plan.md` M4 toggle.

**[NEEDS CLARIFICATION: mermaid version]** marker in `plan.md` M4 resolves this.

---

## §F. Mascot Placement Architecture

### §F.1 Pose → surface map

| Surface | Pose | File edited | Rationale |
|---|---|---|---|
| Home hero | Explaining (welcome, two-hands-open) | `layouts/index.html` | round3 hero intent = welcome; Explaining fits |
| Home hero (alt) | Pointing | `layouts/index.html` | CTA-pointer variant for A/B |
| Section empty state (no pages) | Searching | new `layouts/partials/doc-empty.html` | "빈 결과" — empty section/list |
| 404 | Thinking | `layouts/404.html` | "고민" — empathetic 404 |
| Loading indicator | Thinking | `layouts/_default/baseof.html` (inline loader) or omitted if no loader exists | "로딩 중" — cognitive |
| Success state (form submit, copy-to-clipboard) | Coffee | `render-codeblock.html` copy success + any form partial | "여유" — post-success |
| Tutorial / onboarding | Teaching | content shortcode usage (when tutorial pages exist) | "지시봉 설명" |

**Forbidden placements** (REQ-MAS-003): data tables, forms, checkout. The docs-site has no data tables / checkout; the only form-adjacent surface is the ⌘K search input (header) — no mascot on it.

### §F.2 Render vehicle

- **Existing `mascot.html` shortcode** (currently unused): promote it to the primary vehicle. Extend it to accept a `pose` param (`thinking|pointing|searching|teaching|explaining|coffee`) and emit `<img class="mascot mascot-<pose>" src="/mascots/MoAI-Mascot-<Pose>.png" alt="" loading="lazy" />`.
- **Placement partial** `layouts/partials/mascot.html`: for layout-level placement (home hero, 404, empty state) where a shortcode is not reachable, add a partial that takes the same `pose` param.
- Both render the same `<img>` contract → same CSS target `.mascot`.

### §F.3 Motion

Mascot-only bounce easing `cubic-bezier(0.34,1.56,0.64,1)` (REQ-CMP-004). On hover, a gentle `wiggle` keyframe (formerly in `moai-docs-tokens.css`, folded into `moai-brand.css`). Respect `prefers-reduced-motion: reduce` → 1ms.

---

## §G. 4-Locale Parity Mechanism

### §G.1 Same-PR obligation enforcement

Every CSS / layout / partial edit is locale-agnostic by construction (CSS is shared, layouts render via `hugo.Sites` over all 4 locales). The same-PR obligation (REQ-I18N-001) binds only PROSE copy changes — e.g. a new round3 hero deck field requires ko + en + ja + zh strings.

**Mechanism**: any layout slot that introduces new copy MUST pull from `i18n/<locale>.yaml` or `data/menu/main.yaml` name maps. Hardcoded Korean in a layout template is a locale-parity FAIL.

### §G.2 Verification

acceptance.md §B.8 AC: `diff -r` the rendered HTML across locales for a structural page (e.g. `/ko/` vs `/en/` vs `/ja/` vs `/zh/`) — the DOM structure MUST be identical modulo translated strings. A locale-parity grep verifies no hardcoded Korean leaked into a shared layout.

### §G.3 Canonical chain

ko canonical → en → ja/zh derivation (hns-oss-docs-i18n-rules §1). Structural copy authored in ko first; translations follow in the same PR.

---

## §H. Light-Only Preservation

### §H.1 Mechanism (unchanged)

- `baseof.html` forces `color-theme="light"` on `<html>` + a MutationObserver that re-forces light if anything flips it (REQ-LIT-001).
- The dark-toggle button (if present in round3 nav) is `display:none` on docs-site.
- Mermaid init uses the light themeVariables only (no dark branch) — `foot.html` already strips the dark branch (comment "라이트 테마 단일화 — 다크 분기 제거 2026-05-13").

### §H.2 v2 dark tokens handling

`colors_and_type.css` carries a `[data-theme="dark"]` block. When the docs-site imports the v2 tokens, the dark block comes along as dead code (REQ-LIT-002). It is inert because:
- The MutationObserver prevents `data-theme="dark"` from ever applying.
- The dark-toggle never renders.

The dark block is NOT stripped from the imported token set (it is harmless dead code and stripping it would diverge from the canonical `colors_and_type.css`). A comment notes its inert status.

---

## §I. Risks & Mitigations (design-level)

| Risk | Mitigation |
|---|---|
| geekdoc shell replacement breaks the Hugo render pipeline (section list, taxonomy, TOC) | M3 (layout) is the highest-risk milestone; rollback path in plan.md §I; per-milestone `hugo --minify --gc` warning-free gate (REQ-BLD-001) |
| Token collapse misses a `#faf9f5`/`#ecefee`/`#211A14` literal buried in a component override | acceptance.md token-parity AC greps the entire `docs-site/static/` + `layouts/` for the forbidden literals |
| 4-locale parity drift on new hero copy | All new copy via `i18n/*.yaml`; ko authored first, 3 translations same-PR |
| Mermaid themeVariables key typo breaks diagrams | acceptance.md AC greps the `foot.html` themeVariables block for the exact v2 hex values |
| Mascot pose filename casing (`MoAI-Mascot-Thinking.png` capital-T) vs shortcode param (`thinking` lowercase) | Placement map in §F fixes the canonical casing; shortcode `pose` param is lowercase, filename is Capitalized — the partial maps one to the other |

---

## §J. Decision Summary (most-likely-to-change first)

| # | Decision | Default | Override path |
|---|---|---|---|
| 1 | Token architecture | Single v2-native `moai-brand.css`, layering collapsed | Keep layered override only if a hidden dependency on the remap layer surfaces during M1 |
| 2 | Mono font | Two-token split (`--font-mono` JetBrains + `--font-code` Goorm) | Drop `--font-code`, single JetBrains (v2 purist) |
| 3 | Mermaid version | Stay v10, themeVariables-only | Bump v11 + load-test `mermaid.run` |
| 4 | Featured-page selection | `weight: 0` in `_meta.yaml` | New params field `featured_page` |
| 5 | TOC position | Left rail (round3 faithful) | Right rail (geekdoc default) if round3 layout proves cramped on tablet |
| 6 | Mascot render vehicle | Extend `mascot.html` shortcode + new `partials/mascot.html` | Inline `<img>` only if shortcode proves restrictive |

All defaults are round3-충실-재현-faithful (user decision 2) unless an override is explicitly requested at plan-audit or Implementation Kickoff Approval.
