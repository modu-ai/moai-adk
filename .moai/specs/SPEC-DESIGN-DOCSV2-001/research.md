# research.md — SPEC-DESIGN-DOCSV2-001

> Evidence base for the docs-site v2 migration. Token DELTA table (§C) is the load-bearing artifact referenced by spec.md §F and acceptance.md token-parity ACs.

## §A. v2 Token Core (target state — `colors_and_type.css`)

The canonical v2 token set, sourced verbatim from `.moai/state/ai-design-system/project/colors_and_type.css` `:root`. This is the single target the docs-site CSS layer converges on.

| Token group | Token | v2 value | Role |
|---|---|---|---|
| Point | `--color-primary` | `#3d7d5f` | single accent (mascot sweater green) |
| Point | `--color-primary-hover` | `#316750` | hover |
| Point | `--color-primary-active` | `#265240` | press |
| Point | `--color-ink` | `#060606` | foreground ink (mascot outline) |
| Point | `--color-bg` | `#f4f4f4` | page canvas (achromatic) |
| Point | `--color-surface` | `#ffffff` | cards / modals |
| Neutral | `--neutral-50` → `--neutral-950` | `#f4f4f4` → `#060606` (hue 0%) | 11-step achromatic ramp |
| Semantic | `--color-success` | `#2e8a63` | success |
| Semantic | `--color-warning` | `#c47b2a` | warning |
| Semantic | `--color-danger` | `#c44a3a` | danger |
| Semantic | `--color-info` | `#2a8a8c` | info |
| FG | `--fg-1` / `--fg-2` / `--fg-3` | `#060606` / `#565656` / `#757575` | primary / muted / placeholder |
| Border | `--border-1` / `--border-2` / `--border-strong` | `#d1d1d1` / `#e6e6e6` / `#b5b5b5` | achromatic borders |
| Border | `--border-focus-ring` | `rgba(61,125,95,0.16)` | primary-tinted focus |
| Signature | `--gradient-signature` | `#3d7d5f` (SOLID) | de-emphasized from prior 135deg gradient |
| Signature | `--gradient-signature-soft` | `rgba(61,125,95,0.08)` | subtle tint |
| Signature | `--gradient-signature-dark` | `#316750` | hover tint |
| Type | `--font-sans` | `"Pretendard", system-ui, ...` | Korean + Latin body |
| Type | `--font-latin` | `"Inter", system-ui, sans-serif` | Latin auxiliary |
| Type | `--font-mono` | `"JetBrains Mono", ui-monospace, ...` | code |
| Tracking | `--tracking-display-tight` / `--tracking-display` / `--tracking-heading` / `--tracking-body` / `--tracking-caption` | `-0.075em` / `-0.05em` / `-0.05em` / `-0.025em` / `0` | Notion 자간 |
| Radius | `--radius-sm` / `md` / `lg` / `xl` / `pill` / `full` | `4` / `8` / `16` / `24` / `32` / `9999px` | corner radii |
| Shadow | `--shadow-xs` → `--shadow-xl` | `rgba(6,6,6,X)` 0.04 → 0.12 | ink-based outer-only |
| Shadow | `--shadow-signature` | `0 8px 32px rgba(61,125,95,0.20)` | hover-only green glow |
| Motion | `--easing-default` / `--easing-bounce` / `--easing-smooth` | `cubic-bezier(0.4,0,0.2,1)` / `cubic-bezier(0.34,1.56,0.64,1)` / `cubic-bezier(0.16,1,0.3,1)` | fast ease-out / mascot bounce / page |

### §A.1 Canonical-source conflict: `colors_and_type.css` vs `round3/styles.css`

A minor inconsistency exists inside the bundle:

| Token | `colors_and_type.css` (v2 canonical) | `round3/styles.css` (prototype) | Resolution |
|---|---|---|---|
| bg | `#f4f4f4` | `#f3f3f3` | **`#f4f4f4` wins** (canonical token set per task prompt) |
| ink | `#060606` | `#09110f` | **`#060606` wins** (canonical token set) |
| gradient | solid `#3d7d5f` | solid `#3d7d5f` | agree |
| primary | `#3d7d5f` | `#3d7d5f` | agree |

`README.md` §3 also references `#09110f` as the ink substitute for `#000000`, but the CSS token set is the operational SSOT and it pins `--color-ink: #060606`. The docs-site v2 migration adopts `#060606`. This conflict is recorded so a future reader does not treat `#09110f` prose as authoritative over the CSS.

---

## §B. FROZEN Rules (v2 — bind the migration)

From `SKILL.md` §3 — every rule below is a hard constraint on the docs-site CSS:

1. `#000000` forbidden → use `#060606` (`--color-ink`).
2. Magenta / purple / orange gradients forbidden → single solid green `--gradient-signature` only.
3. Gradient + shadow simultaneous forbidden (visual noise).
4. Full-bleed image background forbidden → page bg is solid `--color-bg`.
5. All-caps Korean forbidden (English abbreviations OK).
6. Mascot forbidden on data tables / forms / checkout — emotional surfaces only.
7. Rounded-border + left-color-accent card pattern forbidden (AI-slop).
8. Inter / Roboto / Arial forbidden for Korean body → Pretendard mandatory.

These map to spec.md REQ-TOK-005/006/009, REQ-TYP-004/006, REQ-MAS-003, REQ-CMP-005/007.

---

## §C. Token DELTA Table (current docs-site → v2 target)

The load-bearing evidence. Every row is a grep-assertable delta. acceptance.md token-parity ACs reference this table.

### §C.1 Background / ink

| Token / literal | Current value | Current file | v2 target | REQ |
|---|---|---|---|---|
| `--color-bg` | `#faf9f5` (warm-cream) | `moai-brand.css` `:root` | `#f4f4f4` (achromatic) | REQ-TOK-003/008 |
| `--bg-page` | `#ecefee` (neutral-cream) | `moai-docs-tokens.css` | removed (use `--color-bg`) | REQ-TOK-008 |
| `--color-ink` | `#060606` (already) | `moai-brand.css` | `#060606` (no change) | REQ-TOK-003 |
| `--ink-900` | `#211A14` | `moai-docs-tokens.css` | removed (use `--color-ink`) | REQ-TOK-008 |
| `#000000` literals | present in misc CSS | various | `#060606` | REQ-TOK-005 |

### §C.2 Primary / signature

| Token | Current value | Current file | v2 target | REQ |
|---|---|---|---|---|
| `--color-primary` | `#3d7d5f` | `moai-brand.css` | `#3d7d5f` (no change) | REQ-TOK-002 |
| `--color-primary-hover` | (n/a — add) | — | `#316750` | REQ-TOK-002 |
| `--color-primary-active` | (n/a — add) | — | `#265240` | REQ-TOK-002 |
| `--clay-500` | `#3d7d5f` | `moai-docs-tokens.css` | removed (use `--color-primary`) | REQ-TOK-001/008 |
| `--gradient-signature` | `linear-gradient(135deg, #3d7d5f 0%, #181715 100%)` | `moai-brand.css` | `#3d7d5f` (solid) | REQ-TOK-004/009 |

### §C.3 Neutral scale

| Token | Current | v2 target | REQ |
|---|---|---|---|
| warm neutral ramp | warm-tinted (#faf9f5 family) | achromatic hue-0% (#f4f4f4 → #060606) | REQ-TOK-007 |

### §C.4 Typography

| Item | Current | v2 target | REQ |
|---|---|---|---|
| Korean title face | MaruBuri (serif) via `@font-face` in `moai-docs-tokens.css` | Pretendard-only (drop MaruBuri `@font-face`) | REQ-TYP-001/002 |
| Korean body face | Pretendard Variable (jsdelivr v1.3.9) | Pretendard Variable (unchanged) | REQ-TYP-001 |
| Mono face | Goorm Sans Code (goorm CDN) | **JetBrains Mono** (recommended, see §D) OR Goorm Sans Code retained | REQ-TYP-003 |
| Latin auxiliary | Inter (Google Fonts via bundle) | Inter (unchanged, in-token) | — |

### §C.5 Mermaid palette

| themeVariable | Current (`foot.html`) | v2 target (`MermaidDiagram.jsx`) | REQ |
|---|---|---|---|
| `primaryColor` | `#d6ebde` | `#eef4f0` | REQ-MER-001/003 |
| `primaryBorderColor` | `#3d7d5f` | `#3d7d5f` (agree) | REQ-MER-001 |
| `primaryTextColor` | `#141413` | `#060606` | REQ-MER-001 |
| `lineColor` | `#3d7d5f` (green) | `#9fa0a0` (gray) | REQ-MER-001/003 |
| `actorBkg` | `#d6ebde` | `#eef4f0` | REQ-MER-001 |
| `noteBkgColor` | `#d6ebde` | `#e6e6e6` | REQ-MER-001/003 |
| `noteBorderColor` | `#3d7d5f` | `#b5b5b5` | REQ-MER-001 |
| `clusterBkg` | `#faf9f5` | `#f4f4f4` | REQ-MER-001/003 |
| `clusterBorder` | `#3d7d5f` | `#d1d1d1` | REQ-MER-001 |
| `titleColor` | `#141413` | `#060606` | REQ-MER-001 |
| `edgeLabelBackground` | `#faf9f5` | `#ffffff` | REQ-MER-001 |
| mermaid version | `mermaid@10` (CDN) | `mermaid@11` (bundle) OR stay v10 | REQ-MER-002 |

---

## §D. Mono-Font Decision — JetBrains Mono vs Goorm Sans Code

### §D.1 The tension

- **v2 정통 정합** (user decision 1) points to **JetBrains Mono** — it is the canonical `--font-mono` in `colors_and_type.css` line 77 and the bundle's `DocPage.jsx` / `MermaidDiagram.jsx` assume it. The round3 prototype loads it via Google Fonts.
- **Current docs-site** uses **Goorm Sans Code** (goorm CDN `statics.goorm.io/fonts/GoormSansCode/v1.0.1/GoormSansCode.min.css`), loaded in `head/custom.html` line 8. The round3 `styles.css` ALSO defines a `--font-code: "CloudSansCode"` (Goorm Sans Code via projectnoonnu CDN) as a separate token from `--font-mono` (JetBrains Mono) — round3 uses Goorm for code-card bodies and JetBrains for `.latin` / `.mono` UI chips.

### §D.2 Korean-aware consideration

Goorm Sans Code is Korean-aware (designed by goorm for Korean code comments / mixed CJK+Latin code). JetBrains Mono is Latin-only. The docs-site code blocks routinely contain Korean comments and Korean string literals. A pure JetBrains Mono render falls back to the system Korean face for Hangul glyphs inside code, which can look inconsistent.

### §D.3 Recommendation

**Recommended**: adopt the round3 two-token pattern —
- `--font-mono: "JetBrains Mono"` for UI chips, eyebrows, meta strips, code-card titles (Latin-only surfaces).
- `--font-code: "Goorm Sans Code", "JetBrains Mono"` for code-card BODIES (Korean-aware, fallback JetBrains).

This is the round3-충실-재현 faithful choice (user decision 2) AND keeps Korean readability inside code blocks. It does NOT add a net-new CDN dep (Goorm is already in use).

**[NEEDS CLARIFICATION: mono-font strategy]** — confirm the two-token split vs. a single JetBrains Mono (v2 정통 정합 purist). Resolved in `design.md` §D with a default (two-token split) and a one-line override path.

---

## §E. Mermaid Version Decision — v10 vs v11

### §E.1 The tension

- **Current docs-site**: `mermaid@10` (foot.html line 87). Stable on the docs-site for months.
- **Bundle**: `mermaid@11` (MermaidDiagram.jsx line 46). v11 has API changes (notably `mermaid.run` vs `mermaid.init` deprecation tweaks, and `securityLevel` defaults).

### §E.2 Risk

The docs-site `foot.html` mermaid loader is hand-rolled (lazy-load on `pre.gdoc-mermaid` detection, `theme:'base'` init, `mermaid.run({querySelector})`). A v10→v11 bump risks:
- `mermaid.run` signature changes (v11 may require explicit container).
- `themeVariables` key renames (none observed in the delta, but mermaid minor versions occasionally rename sequence-diagram keys).
- The geekdoc `render-codeblock-mermaid.html` hook output shape.

### §E.3 Recommendation

**Recommended**: stay on **mermaid@10** for this SPEC (apply only the themeVariables update REQ-MER-001). Rationale: the v2 palette shift is entirely a themeVariables change — no v11 feature is referenced. Bumping to v11 adds load-test risk for zero visual gain. A separate mermaid@11 bump SPEC can follow once the v2 palette is stable.

**[NEEDS CLARIFICATION: mermaid version]** — confirm stay-on-v10 default vs. v11 bump. Resolved in `design.md` §E.

---

## §F. round3 Prototype Breakdown

### §F.1 docs-index (`round3/05-docs-index.html`)

Layout slots (top-to-bottom):

1. **nav** (sticky, 64px, glassy `color-mix(bg 85%)` + `blur(12px)`) — logo-mark + nav-menu + nav-search (⌘K) + theme-toggle + login.
2. **docs-hero** — `docs-eyebrow` (mono uppercase) + `docs-h1` (clamp 36–56px, weight 900, tracking-display) + `docs-sub` + `docs-stats` row (4 stats, mono numbers, top-border separator).
3. **docs-filters** (sticky, top:64px) — `pill-row` (category pills with count badges, active = ink bg) + `filter-divider` + `sort-select`.
4. **featured** — `featured-card` (solid `--gradient` bg, 24px radius, 2-col 1.5fr/1fr) with `featured-eye` (pulse dot) + h2 + p + meta + `featured-cta` (white pill on green) + `featured-illu` (aspect 1:1, glassy).
5. **docs-grid-section** — `grid-section-head` (h3 + count) + `docs-grid` (4-col, `repeat(4,1fr)`, 16px gap) of `doc-card`s.
6. **doc-card** — `doc-thumb` (aspect 4:5, themed bg `.thumb-1..8`, cat badge + icon + big ghost number) + `doc-body` (tags + h4 + excerpt + meta strip with read-time/views).
7. **footer** — 4-col grid (brand tag + 3 link cols) + bottom strip.
8. **mobile-nav** — 5-col bottom sticky (hidden on desktop).

Responsive: ≤1023px → 2-col grid + 1-col featured; ≤640px → 1-col grid.

### §F.2 docs-detail (`round3/06-docs-detail.html`)

Layout slots:

1. **nav** (same shell).
2. **read-progress** (fixed, top:64px, 3px, gradient fill on scroll).
3. **doc-hero** (solid `--gradient` bg, green) — `crumb` (breadcrumb) + `doc-cat` pill + `doc-h1` (clamp 32–48px) + `doc-deck` + `doc-byline` (author avatar + meta strip).
4. **doc-layout** (3-col grid: `220px 1fr 280px`, 48px gap) —
   - **toc** (sticky, top:96px) — `toc-h` + `toc-list` (active = primary border-left + bold).
   - **doc-body** (max 720px, `prose-kr`) — h2/h3, callout (soft gradient bg + left border), code-mac (macOS code card), code-mac-preview, pattern card, compare grid (bad/good), next-cta (green gradient).
   - **rail** (sticky, top:96px) — rail-cards (actions, related guides, paid-course CTA with gradient bg).
5. **footer** + **mobile-nav**.

JS: reading-progress scroll handler + IntersectionObserver for TOC active state.

Responsive: ≤1199px → 2-col (rail hidden); ≤1023px → 1-col (TOC hidden).

### §F.3 round3 → docs-site mapping consideration

round3 is a static prototype (no i18n, no Hugo, no geekdoc). The docs-site migration MUST recreate the visual slots inside Hugo's layout system:
- `nav` → existing `site-header.html` partial (preserve ⌘K search + lang-switch + GitHub Star — these are NOT in round3 but are load-bearing docs-site infra).
- `docs-hero` / `doc-hero` → new hero partial or section template.
- `docs-filters` pills → Hugo taxonomy-driven (categories from `_meta.yaml`).
- `docs-grid` → Hugo section list template.
- `toc` + `read-progress` → geekdoc already has TOC infra (`aside.cw-toc` in baseof.html); adapt styling.
- `rail` → new right-rail partial.
- `footer` → existing `site-footer.html` partial (preserve copyleft + attributions).

The i18n / menu / search infrastructure is round3-agnostic and MUST be carried over unchanged (REQ-LAY-003).

---

## §G. Current docs-site 3-Layer Token Divergence

The current docs-site CSS stack (load order is load-bearing, wired in `head/custom.html` with FNV32a `?h=` cache-busting):

| # | File | Lines | Vocabulary | Status |
|---|---|---|---|---|
| 1 | `static/moai-brand.css` | 2139 | **FROZEN SSOT**. `:root` warm-cream (`--color-bg #faf9f5`, `--color-primary #3d7d5f`, warm neutral ramp, `--gradient-signature linear-gradient(135deg,#3d7d5f 0%,#181715 100%)`) + ALL `cw-*`/`gdoc-*` overrides. | **UNFROZEN by this SPEC** (§G of spec.md) |
| 2 | `static/moai-design.css` | 1243 | macOS code-card chrome + layout shell `.shell/.nav` + sidebar + callouts + hero + plugin grid + stat tiles + track cards + release timeline + filter chips. | Rewritten to v2 tokens |
| 3 | `static/moai-docs-tokens.css` | 332 | Clay/Cream/Ink (`--clay-500 #3d7d5f`, `--bg-page #ecefee`, `--ink-900 #211A14`) + MaruBuri `@font-face` + motion keyframes. | **Collapsed into #1** (MaruBuri dropped, Clay/Cream/Ink removed) |
| 4 | `static/moai-docs-theme.css` | 897 | Remaps FROZEN `--color-*` onto Clay/Cream/Ink (so existing components re-tint without touching brand.css) + `md-*` landing/404 components. | **Collapsed into #1/#2** (remap layer obsolete once brand.css is v2-native) |
| 5 | `assets/css/moai-brand.scss` | 797 | **STALE, NOT compiled** (`#144a46` / `#f3f3f3`). | Ignored (not a source of truth); may be deleted as cleanup |

The v2 migration collapses layers 1+3+4 into a single v2-native token + component file (strategy in `design.md` §A). Layer 2 is rewritten to consume the v2 tokens. Layer 5 is stale and ignored.

---

## §H. Component Contract Evidence

### §H.1 DocPage.jsx

Article shell, var-token driven. Tokens consumed (all present in v2 `colors_and_type.css`): `--color-surface`, `--fg-1`, `--font-sans`, `--lh-relaxed`, `--tracking-body`, `--shadow-md`, `--border-1`, `--color-primary`, `--font-mono`, `--tracking-display`, `--fg-2`. Sizes: a4 (210×297mm), letter (216×279mm), legal (216×356mm). The docs-site doc-body adopts this contract (REQ-LAY-004).

### §H.2 MermaidDiagram.jsx

Already covered in §C.5 / §E. Loads `mermaid@11` from jsdelivr CDN; uses `theme:'base'` + the v2 themeVariables. The docs-site `foot.html` applies the same themeVariables but via `mermaid@10` (the version delta in §E).

---

## §I. Mascot Pose Inventory

The 6 poses at `.moai/state/ai-design-system/project/assets/characters/`:

| Pose | File | Emotional surface (round3 / SKILL.md mapping) |
|---|---|---|
| Thinking | `MoAI-Mascot-Thinking.png` | loading indicator, "고민 상태" |
| Pointing | `MoAI-Mascot-Pointing.png` | CTA, 안내, home hero |
| Searching | `MoAI-Mascot-Searching.png` | search empty result, "빈 결과" |
| Teaching | `MoAI-Mascot-Teaching.png` | tutorial, onboarding |
| Explaining | `MoAI-Mascot-Explaining.png` | welcome, home hero, explanation |
| Coffee | `MoAI-Mascot-Coffee.png` | success state, "여유" |

Current docs-site `static/mascots/` has: `mascot-coding.png` (header + home hero), `mascot-talking.png` (home divider), `mascot-bubble.png`, `mascot-coding-alt.png` (unused). The 6 bundle poses are NET-NEW additions (REQ-MAS-001). The placement map is in `design.md` §F.

---

## §J. References

- Design bundle: `.moai/state/ai-design-system/project/README.md`, `SKILL.md`, `colors_and_type.css`, `round3/05-docs-index.html`, `round3/06-docs-detail.html`, `round3/styles.css`, `assets/components/DocPage.jsx`, `assets/components/MermaidDiagram.jsx`, `assets/characters/MoAI-Mascot-*.png`.
- docs-site ground truth: `docs-site/hugo.toml`, `docs-site/vercel.json`, `docs-site/layouts/partials/head/custom.html`, `docs-site/layouts/partials/foot.html`, `docs-site/layouts/_default/baseof.html`, `docs-site/layouts/index.html`, `docs-site/static/moai-brand.css`, `docs-site/static/moai-design.css`, `docs-site/static/moai-docs-tokens.css`, `docs-site/static/moai-docs-theme.css`, `docs-site/data/menu/main.yaml`.
- i18n rules: `.moai/docs/docs-site-i18n-rules.md` + Skill `hns-oss-docs-i18n-rules`.
- CLAUDE.local.md §17.1 (prior Claude Warm Editorial regime — superseded by v2 정통 정합 for this SPEC), §23 (Hybrid Trunk).
