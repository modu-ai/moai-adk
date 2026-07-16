# acceptance.md — SPEC-DESIGN-DOCSV2-001

> Acceptance criteria for the docs-site v2 migration. 30 ACs spanning token parity, typography, layout, mermaid, mascot, components, 4-locale parity, light-only, build gate. Each AC is observable via grep / rendered-DOM assertion / build exit code.

## §A. AC Matrix

| AC ID | REQ IDs | Severity | Summary |
|---|---|---|---|
| AC-TOK-001 | REQ-TOK-001/003/008 | MUST | `--color-bg #f4f4f4` is the single bg token; no `#faf9f5` / `#ecefee` literal |
| AC-TOK-002 | REQ-TOK-002 | MUST | `--color-primary`/`-hover`/`-active` = `#3d7d5f`/`#316750`/`#265240` |
| AC-TOK-003 | REQ-TOK-003 | MUST | `--color-ink #060606` present; no `#211A14` literal |
| AC-TOK-004 | REQ-TOK-004/009 | MUST | `--gradient-signature` = solid `#3d7d5f`; no `linear-gradient(135deg, #3d7d5f` literal |
| AC-TOK-005 | REQ-TOK-005 | MUST | No `#000000` literal in any CSS foreground/background role |
| AC-TOK-006 | REQ-TOK-006 | MUST | No element has both gradient background and box-shadow |
| AC-TOK-007 | REQ-TOK-007 | MUST | Neutral scale is achromatic (hue 0%) `#f4f4f4` → `#060606` |
| AC-TYP-001 | REQ-TYP-001/002 | MUST | Pretendard sole Korean title/body; no MaruBuri `@font-face` |
| AC-TYP-002 | REQ-TYP-003 | SHOULD | Mono-font decision applied (two-token split OR single JetBrains) |
| AC-TYP-003 | REQ-TYP-004 | MUST | No Inter/Roboto/Arial for Korean body (`--font-sans` resolves to Pretendard) |
| AC-TYP-004 | REQ-TYP-005 | SHOULD | Letter-spacing tokens present (display -0.075em, heading -0.05em, body -0.025em) |
| AC-LAY-001 | REQ-LAY-001 | MUST | Home/listing renders round3 docs-index slots (hero + sticky pills + featured + 4:5 grid) |
| AC-LAY-002 | REQ-LAY-002 | MUST | Doc-detail renders round3 docs-detail slots (3-col + sticky TOC + read-progress + next-CTA) |
| AC-LAY-003 | REQ-LAY-003 | MUST | i18n/menu/search infra preserved (4-locale nav, ⌘K search, lang-switch all render) |
| AC-LAY-004 | REQ-LAY-004 | SHOULD | Doc body adopts DocPage article-shell (var-token driven) |
| AC-LAY-005 | REQ-LAY-005 | MUST | geekdoc visual shell replaced; `gdoc-nav`/`gdoc-page` block hooks still drive render |
| AC-MER-001 | REQ-MER-001 | MUST | mermaid themeVariables = v2 palette (`primaryColor #eef4f0`, `lineColor #9fa0a0`, etc.) |
| AC-MER-002 | REQ-MER-002/003 | SHOULD | mermaid version decision applied; no warm-cream `#d6ebde`/`#faf9f5` in mermaid init |
| AC-MAS-001 | REQ-MAS-001 | MUST | 6 MoAI-Mascot PNGs exist in `static/mascots/` |
| AC-MAS-002 | REQ-MAS-002 | MUST | Mascot present on home hero AND 404 AND empty-section (≥3 emotional surfaces) |
| AC-MAS-003 | REQ-MAS-003/004 | MUST | No mascot on data/form/checkout; mascots rendered via shortcode/partial |
| AC-CMP-001 | REQ-CMP-001/002 | MUST | v2 button + card recipes applied (computed styles match) |
| AC-CMP-002 | REQ-CMP-003/004 | SHOULD | v2 shadow + motion tokens present and consumed |
| AC-CMP-003 | REQ-CMP-005 | MUST | No rounded-border + left-color-accent card pattern |
| AC-CMP-004 | REQ-CMP-006/007 | MUST | No body emoji in new/modified components; no full-bleed image bg |
| AC-I18N-001 | REQ-I18N-001/002 | MUST | 4-locale same-PR parity (structural DOM identical modulo translations) |
| AC-I18N-002 | REQ-I18N-003 | MUST | No net-new external URL beyond adk.mo.ai.kr whitelist |
| AC-LIT-001 | REQ-LIT-001/002 | MUST | LIGHT-ONLY maintained (MutationObserver present; dark-toggle display:none; no dark-mode render) |
| AC-BLD-001 | REQ-BLD-001 | MUST | `hugo --minify --gc` exits 0 with 0 warnings |
| AC-BLD-002 | REQ-BLD-002/003 | MUST | `vercel.json` redirects + `/api/i18n-detect` + FNV32a cache-busting intact |

**Headline AC count: 30** (20 MUST, 10 SHOULD — SHOULD ACs bind the mono-font / mermaid-version / DocPage / shadow-motion decisions; a SHOULD FAIL does not block merge but must be noted in sync-phase).

---

## §B. Detailed Scenarios (Given-When-Then)

### §B.1 Token Parity

**Scenario AC-TOK-001**: bg token unification
- **Given** the docs-site CSS layer is migrated,
- **When** an auditor runs `grep -rn '#faf9f5\|#ecefee' docs-site/static/ docs-site/layouts/`,
- **Then** the command returns 0 matches AND `grep -n '^\s*--color-bg' docs-site/static/moai-brand.css` returns `--color-bg: #f4f4f4`.

**Scenario AC-TOK-004**: signature gradient de-emphasis
- **Given** the docs-site CSS layer is migrated,
- **When** an auditor runs `grep -n 'linear-gradient(135deg, #3d7d5f' docs-site/static/ docs-site/layouts/ -r`,
- **Then** the command returns 0 matches AND `grep -n '^\s*--gradient-signature' docs-site/static/moai-brand.css` returns a solid `#3d7d5f` (no `linear-gradient`).

**Scenario AC-TOK-005**: no pure-black
- **Given** the docs-site CSS layer is migrated,
- **When** an auditor runs `grep -rn '#000000' docs-site/static/moai-brand.css docs-site/static/moai-design.css docs-site/layouts/ | grep -v '^[^:]*:[0-9]*:[ \t]*\(/\*\|//\)'`,
- **Then** the command returns 0 matches (comment lines excluded).

**Scenario AC-TOK-006**: no gradient+shadow simultaneity
- **Given** the docs-site CSS is migrated,
- **When** an auditor scans each CSS rule block,
- **Then** no single selector's declaration block contains BOTH a `background:.*gradient` (or `background: var(--gradient-signature)` expanding to gradient) AND a `box-shadow` declaration.

### §B.2 Typography

**Scenario AC-TYP-001**: Pretendard-only
- **Given** the docs-site is migrated,
- **When** an auditor runs `grep -rin 'maruburi' docs-site/`,
- **Then** the command returns 0 matches AND `grep -c '@font-face' docs-site/static/moai-brand.css` returns only Pretendard self-host entries (or 0 if Pretendard is CDN-only).

**Scenario AC-TYP-003**: no forbidden Korean-body fonts
- **Given** the docs-site `--font-sans` token,
- **When** an auditor resolves the token,
- **Then** its value begins with `"Pretendard"` and does NOT list Inter, Roboto, or Arial as the first (Korean-body) family.

### §B.3 Layout

**Scenario AC-LAY-001**: docs-index structure
- **Given** the migrated home page at `/ko/`,
- **When** an auditor renders the page and inspects the DOM,
- **Then** the DOM contains (in order): a `.nav`-equivalent header, a docs-hero block (eyebrow + h1 + sub + stats), a sticky filter bar with category pills, a featured card, and a 4:5-aspect card grid.

**Scenario AC-LAY-002**: docs-detail structure
- **Given** a migrated doc detail page,
- **When** an auditor renders the page,
- **Then** the DOM contains a 3-column layout (TOC left + body center + rail right), a reading-progress bar, a doc-hero (crumb + cat + h1 + deck + byline), and a next-CTA.

**Scenario AC-LAY-003**: infra preserved
- **Given** the migrated docs-site,
- **When** an auditor exercises the ⌘K search, the lang-switch (KO/EN/JA/CN), and the sidebar nav across 4 locales,
- **Then** all three function and render locale-appropriate content; `data/menu/main.yaml`-driven sidebar items render their icons (no empty SVGs).

### §B.4 Mermaid

**Scenario AC-MER-001**: v2 mermaid palette
- **Given** the migrated `foot.html`,
- **When** an auditor runs `grep -n "primaryColor.*#eef4f0" docs-site/layouts/partials/foot.html` and `grep -n "lineColor.*#9fa0a0" docs-site/layouts/partials/foot.html`,
- **Then** both commands return 1+ match AND `grep -n '#d6ebde\|#faf9f5' docs-site/layouts/partials/foot.html` returns 0 matches.

### §B.5 Mascot

**Scenario AC-MAS-001**: mascot assets installed
- **Given** M5 is complete,
- **When** an auditor runs `ls docs-site/static/mascots/MoAI-Mascot-*.png`,
- **Then** 6 PNGs are listed (Thinking, Pointing, Searching, Teaching, Explaining, Coffee).

**Scenario AC-MAS-002**: mascot on emotional surfaces
- **Given** the migrated docs-site,
- **When** an auditor renders the home page, the 404 page, and an empty section list page,
- **Then** each rendered DOM contains exactly one `img.mascot` element with a canonical `MoAI-Mascot-<Pose>.png` `src`.

**Scenario AC-MAS-003**: no mascot on forbidden surfaces
- **Given** the migrated docs-site (which has no data tables / forms / checkout),
- **When** an auditor greps the layouts for mascot placement,
- **Then** no mascot partial/shortcode is placed inside a `<table>`, `<form>`, or checkout-template partial.

### §B.6 Components

**Scenario AC-CMP-001**: v2 button + card recipes
- **Given** the migrated docs-site,
- **When** an auditor resolves the computed style of a `.btn-primary` (or v2-equivalent) and a `.card`-class element,
- **Then** the button has `border-radius: 9999px`, `background: rgb(61,125,95)` (solid `#3d7d5f`), `font-weight >= 700`, and the card has `border-radius: 16px`, `border: 1px solid rgb(209,209,209)` (`#d1d1d1`), `box-shadow` with `rgba(6,6,6,...)`.

**Scenario AC-CMP-003**: no AI-slop card pattern
- **Given** the migrated CSS,
- **When** an auditor greps for a `border-left` + `border-radius` adjacency in a single selector block,
- **Then** no card-class selector has both a colored left border AND a rounded radius as its decorative pattern.

### §B.7 4-Locale Parity

**Scenario AC-I18N-001**: structural parity
- **Given** the migrated docs-site,
- **When** an auditor runs `diff -r docs-site/public/ko/ docs-site/public/en/` (and ja, zh) on the rendered output,
- **Then** the diff shows ONLY translated-string differences (no structural / class / DOM-shape differences).

**Scenario AC-I18N-002**: URL whitelist
- **Given** the migrated docs-site,
- **When** an auditor runs `grep -rE 'docs\.moai-ai\.dev|adk\.moai\.com|adk\.moai\.kr' docs-site/`,
- **Then** the command returns 0 matches (adk.mo.ai.kr is the only valid docs domain; note `.mo.ai.kr` vs `.moai.kr`).

### §B.8 Light-Only

**Scenario AC-LIT-001**: no dark mode
- **Given** the migrated docs-site,
- **When** an auditor renders any page and inspects the `<html>` element,
- **Then** `data-theme="light"` (or color-theme="light") is forced, a MutationObserver is attached, and no dark-mode toggle is visible (`display:none`).

### §B.9 Build Gate

**Scenario AC-BLD-001**: warning-free build
- **Given** the migrated docs-site,
- **When** an auditor runs `cd docs-site && hugo --minify --gc 2>&1`,
- **Then** the command exits 0 AND the output contains 0 lines matching `WARN`/`ERROR`.

**Scenario AC-BLD-002**: infra intact
- **Given** the migrated docs-site,
- **When** an auditor renders the home page and triggers `/api/i18n-detect` + exercises a `vercel.json` redirect,
- **Then** the edge function returns a locale AND the redirect resolves AND the rendered HTML references the 2 CSS files with `?h=<FNV32a>` cache-bust query params.

---

## §C. Severity Definitions

- **MUST**: blocks merge. A MUST FAIL at sync-phase gates blocks the SPEC from `completed`.
- **SHOULD**: does not block merge but MUST be noted in sync-phase debt log. SHOULD ACs cover decisions with documented defaults (mono-font, mermaid version, DocPage adoption, shadow/motion).

## §D. Indirect Verification (grep-based, no runtime)

The following ACs are verified by static grep over `docs-site/` (no Hugo render needed) — useful for fast iteration:
- AC-TOK-001, AC-TOK-003, AC-TOK-004, AC-TOK-005 (token literal greps)
- AC-TYP-001 (MaruBuri purge)
- AC-MER-001 (mermaid palette greps)
- AC-MAS-001 (mascot PNG `ls`)
- AC-I18N-002 (URL blacklist)
- AC-CMP-004 (body-emoji + full-bleed-image greps)

The remaining ACs require a Hugo render + DOM inspection (AC-LAY-001/002/003, AC-MAS-002, AC-CMP-001, AC-LIT-001, AC-BLD-001/002).

## §E. Closure Gates

A SPEC closure (sync-phase `completed`) requires:
1. All 20 MUST ACs PASS.
2. All 10 SHOULD ACs PASS OR explicitly deferred in the sync-phase debt log with a follow-up SPEC ID.
3. `hugo --minify --gc` exits 0, 0 warnings (AC-BLD-001).
4. 4-locale same-PR parity verified (AC-I18N-001).
5. Token-parity grep returns 0 forbidden literals (AC-TOK-001/003/004/005).
6. `moai-brand.css` re-stamped FROZEN with v2 vocabulary at sync close.

## §F. Forward-Looking Checks (post-merge debt)

- Mermaid@11 bump (if M4 stayed on v10) → follow-up SPEC.
- Mono-font single-token consolidation (if M2 adopted two-token split and the user later wants v2 purism) → follow-up SPEC.
- mascot-coding.png / mascot-talking.png legacy assets (not in the v2 6-pose set) — audit for removal or retention as alternates.
- `moai-docs-theme.js` removal (if M1 preserved it stripped) — confirm no remaining reference, then delete.
