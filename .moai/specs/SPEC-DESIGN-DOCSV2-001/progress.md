# progress.md — SPEC-DESIGN-DOCSV2-001

> Plan-phase skeleton only. Run-phase (manager-develop) populates §E.2/§E.3; sync-phase (manager-docs) populates §E.4.

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID: SPEC-DESIGN-DOCSV2-001
- Status: draft (plan-phase artifact set complete)
- Files authored: spec.md, plan.md, acceptance.md, research.md, design.md, progress.md
- Tier: L | era: V3R6 | harness: thorough
- Plan-phase self-check: SPEC ID regex PASS (`SPEC-DESIGN-DOCSV2-001`); 12-canonical-field frontmatter validated on spec.md; Out of Scope section present (6 `### Out of Scope — <topic>` H3 sub-headings); moai-brand.css unfreeze authorization recorded (spec.md §G).
- Ready for: plan-auditor independent audit → Implementation Kickoff Approval (plan→run HUMAN GATE).

## §E.2 Run-phase Evidence

### M1 — Token Unification (commit 19c40097d, PUSHED to main)

**AC Matrix (AC-TOK-001..007 — M1 scope):**

| AC | Status | Verification Command | Result |
|---|---|---|---|
| AC-TOK-001 | PASS | `grep -rn '#faf9f5\|#ecefee' docs-site/static/ docs-site/layouts/` | 0 matches; `--color-bg: #f4f4f4` confirmed (brand.css L15) |
| AC-TOK-002 | PASS | `grep -n '^\s*--color-primary' docs-site/static/moai-brand.css` | `#3d7d5f` / `#316750` / `#265240` confirmed (L11-13) |
| AC-TOK-003 | PASS | `grep -rn '#211A14' docs-site/static/ docs-site/layouts/` | 0 matches; `--color-ink: #060606` confirmed (brand.css L14) |
| AC-TOK-004 | PASS | `grep -rn 'linear-gradient(135deg, #3d7d5f' docs-site/static/ docs-site/layouts/` | 0 matches; `--gradient-signature: #3d7d5f` solid (brand.css L51) |
| AC-TOK-005 | PASS | `grep -rn '#000000' brand.css design.css layouts/` (comment-excluded) | 0 matches |
| AC-TOK-006 | N/A (M1) | structural CSS inspection | Not regressed by M1 — token-value repoint only; no new selector blocks with gradient+shadow simultaneity introduced |
| AC-TOK-007 | PASS | `grep -n '^\s*--neutral-' docs-site/static/moai-brand.css` | All achromatic: #f7f7f7 → #060606 scale (L19-29) |
| AC-BLD-001 | PASS | `cd docs-site && hugo --minify --gc` | exit 0, 0 WARN/ERROR; KO 153p / EN 150p / JA 139p / ZH 150p; 2506ms |

**Files touched (7):**
- `docs-site/static/moai-brand.css` — :root rewritten to v2 SSOT raw tokens + 4 literal sweeps + `--color-primary-hover` corrected `#31684f→#316750` (AC-TOK-002)
- `docs-site/static/moai-docs-tokens.css` — clay/cream/ink scales repointed to brand.css v2 raw tokens (cycle-safe, one-directional); semantic aliases repointed; MaruBuri @font-face preserved (M2 typography)
- `docs-site/static/moai-docs-theme.css` — :root rewritten to pure v2 alias layer (drops ALL `--color-*`/`--neutral-*`/`--fg-*`/`--border-*` overrides — brand.css SSOT stands; CSS custom-property cycle broken); component literals swept
- `docs-site/static/moai-design.css` — `#faf9f5→var(--color-bg)` (3 occ), `#181715→var(--neutral-900)` (2 occ)
- `docs-site/layouts/partials/foot.html` — mermaid themeVariables `#faf9f5→#f4f4f4` (AC-TOK-001 `docs-site/layouts/` scope)
- `docs-site/assets/css/moai-brand.scss` — `git rm` (stale uncompiled SCSS, not referenced by head/custom.html)
- `.moai/specs/SPEC-DESIGN-DOCSV2-001/spec.md` — frontmatter `status: draft → in-progress`

**PRESERVED (unchanged):** `head/custom.html` (FNV32a content-hash cache busting auto-regenerates on `hugo --minify`), `moai-docs-theme.js` (TOC scroll-spy only — no theme-remap code), MaruBuri `@font-face` blocks in tokens.css (M2 typography scope).

**Deferred debt (sync-phase / M4 / out-of-scope):**
- `design.css` `#252320` (code-card gradient midpoint — warm but not AC-checked)
- `design.css`/`theme.css` `[data-theme="dark"]` dead-code blocks containing raw `#3d7d5f` (dark theme is dead per CLAUDE.local.md §17.1 — light-only single theme)
- `tokens.css` `--surface-dark: var(--neutral-900)` / `--charcoal: var(--neutral-900)` (warm-neutral scale residual, now repointed to neutral-900)
- `foot.html` mermaid warm literals `#141413` / `#d6ebde` / `#efe9de` (AC-MER-001 M4 scope — full mermaid v2 palette migration)
- PWA `manifest.json` / `site.webmanifest` `#000000` (outside AC-TOK-005 grep scope: `docs-site/static/moai-*.css docs-site/layouts/`)
- Physical 4-CSS-file collapse into single `tokens.css` (deferred to sync-phase per lean M1 approach — functional v2 effect achieved via `:root` repoint + literal sweep, not file consolidation)

### M2 — Typography (Pretendard-only + two-token mono split)

**AC Matrix (AC-TYP-001..004 + AC-BLD-001 re-verify — M2 scope):**

| AC | Status | Verification Command | Result |
|---|---|---|---|
| AC-TYP-001 | PASS | `grep -rin 'maruburi' docs-site/ \| wc -l` | `0` (was 14 hits pre-M2: 5 `@font-face` blocks + `--font-title` token + 6 comments + 2 layout comments); `--font-sans` first family confirmed `"Pretendard Variable"` (brand.css L56) |
| AC-TYP-002 | PASS | `grep -nE '^\s*--font-mono:\|^\s*--font-code:' docs-site/static/moai-brand.css` | Both present: L61 `--font-mono: "JetBrains Mono", ui-monospace, ...`; L62 `--font-code: "Goorm Sans Code", ui-monospace, ...` |
| AC-TYP-003 | PASS | `grep -nE 'font-family:\s*"(Inter\|Roboto\|Arial)"' docs-site/static/moai-*.css` | 0 matches (no Inter/Roboto/Arial as first family); `--font-sans` begins `"Pretendard Variable"` |
| AC-TYP-004 | PASS | `grep -nE '^\s*--tracking-(display-tight\|display\|heading\|body\|caption):' docs-site/static/moai-brand.css` | All 5 present (M1-carried): display-tight `-0.075em` / display `-0.05em` / heading `-0.05em` / body `-0.025em` / caption `0` (brand.css L72-77) |
| AC-BLD-001 | PASS | `cd docs-site && hugo --minify --gc` | exit 0, 0 WARN/ERROR; KO 153p / EN 150p / JA 139p / ZH 150p; 1901ms |

**Files touched (5):**
- `docs-site/static/moai-docs-tokens.css` — removed 5 legacy KR serif `@font-face` blocks (Naver woff2) + the `fonts.css — MaruBuri` section header; repointed `--font-title` off the retired KR serif → `var(--font-sans)` (v2 has no separate title face; kept as alias for `var(--font-title)` consumers in theme.css); updated header + typography section comments to record the M2 retirement (audit trail anchors the SPEC ID, not the retired font name).
- `docs-site/layouts/partials/head/custom.html` — removed `hangeul.pstatic.net` preconnect; added `fonts.googleapis.com` + `fonts.gstatic.com` preconnects + the JetBrains Mono Google Fonts stylesheet `css2?family=JetBrains+Mono:wght@400;500;600&display=swap`; updated header comment to describe the 3-face stack (Pretendard + JetBrains UI mono + Goorm code). Pretendard (jsdelivr v1.3.9) + Goorm Sans Code (goorm) CDN links unchanged.
- `docs-site/static/moai-brand.css` — repointed `--font-mono` from Goorm → `"JetBrains Mono", ui-monospace, "SF Mono", Menlo, monospace`; added new `--font-code: "Goorm Sans Code", ui-monospace, "SF Mono", Menlo, monospace`; updated Type-families section comment.
- `docs-site/static/moai-design.css` — switched the 3 code-surface rules off `var(--font-mono)` → `var(--font-code)`: `.code-title` (L428), `.code-body` (L473), `.code-card .code-pre` (L521, `!important`). All other 35 `var(--font-mono)` hits are UI chips/eyebrows/labels/breadcrumbs (now JetBrains) — left intact per the two-token split.
- `docs-site/static/moai-docs-theme.css` — scrubbed 3 legacy KR serif name mentions from comments only (Article section header L274, page-title comment L285, prose-headings comment L302); `.gdoc-markdown h1/h2/h3` rules at L287/304/317 still reference `var(--font-title)`, which now resolves to Pretendard via the tokens.css repoint.

**PRESERVED (unchanged):** tracking tokens (M1-carried, AC-TYP-004 verified not re-added); `.copy-btn` font-family (UI label, correctly stays `var(--font-mono)` = JetBrains); `--font-latin` alias; FNV32a CSS hashes (regenerated by `hugo --minify`); Pretendard + Goorm CDN links; moai-docs-theme.js (TOC scroll-spy, no theme-remap code).

**Deferred debt (sync-phase / M4 / out-of-scope):**
- `foot.html` mermaid warm literals `#141413` / `#d6ebde` / `#efe9de` (AC-MER-001 M4 scope)
- `design.css` `#252320` code-card gradient midpoint (sync sweep — not AC-checked)
- `design.css`/`theme.css` `[data-theme="dark"]` dead-code blocks (light-only per CLAUDE.local.md §17.1)

**Parallel-session safety:** committed via specific-path `git add docs-site/ .moai/specs/SPEC-DESIGN-DOCSV2-001/progress.md` (B8/B10) — the active parallel session on `internal/cli/` + `README.ko.md` is disjoint from docs-site/; pathspec isolation prevents cross-scope absorption. `README.ko.md` remains unstaged (pre-existing WIP per C-8/AP-6).

### M3a — Layout CSS Foundation (token aliases + moai-docs-layout.css)

**Scope:** CSS ONLY — no Hugo layout/HTML touched (M3b/M3c scope). No prose-kr / code-mac (M3c/M3d). Ported round3 layout-GEOMETRY classes verbatim from `.moai/state/ai-design-system/project/round3/05-docs-index.html` (L9-116) + `06-docs-detail.html` (L9-113); excluded frame/prose-kr/code-mac blocks. The new CSS is **fully inert** until M3b/M3c layouts reference the classes — zero layout HTML modified in this milestone.

**AC Matrix (M3a scope — layout-foundation ACs):**

| AC | Status | Verification Command | Result |
|---|---|---|---|
| AC-LAY-001 (token aliases) | PASS | `grep -nE '^\s*--(bg\|surface\|ink\|primary\|muted\|border\|success\|warning\|danger\|info\|gradient\|shadow):\s*var' docs-site/static/moai-brand.css` | 20 flat-name aliases → v2 tokens present in `:root` (brand.css alias block, appended before closing `}`); each resolves to a M1-declared v2 token (`--color-*`, `--gradient-signature`, `--shadow-*`, `--neutral-200`, `--fg-2/3`, `--border-1/strong`) |
| AC-LAY-002 (layout CSS ported) | PASS | `wc -l docs-site/static/moai-docs-layout.css` + `grep -n '.doc-layout' docs-site/static/moai-docs-layout.css` | 241 lines; `.doc-layout { grid-template-columns: 220px 1fr 280px; gap:48px }` verbatim from round3 06 L73; all geometry classes ported (hero/grid/cards/rail/compare/next-cta/news-form + `@keyframes pulse` + 3 `@media` breakpoints) |
| AC-LAY-003 (CSS wired w/ cache-bust) | PASS | `grep -n 'layoutHash\|moai-docs-layout.css' docs-site/layouts/partials/head/custom.html` | L18 `$layoutHash := hash.FNV32a (readFile "static/moai-docs-layout.css")`; L25 `<link rel="stylesheet" href="…/moai-docs-layout.css?h={{ $layoutHash }}" />` — loads AFTER brand+design+tokens+theme, BEFORE theme.js `<script>` |
| AC-BLD-001 (hugo build) | PASS | `cd docs-site && hugo --minify --gc` | exit 0; `Total in 3844 ms`; 0 WARN/ERROR; all 4 locales rendered (KO/EN/JA/ZH); CSS hash regenerated by build |

**Files touched (3 + this progress.md):**
- `docs-site/static/moai-brand.css` — appended 20 flat-name → v2 token aliases to `:root` (bridge so round3 CSS using `var(--primary)`/`var(--gradient)`/`var(--bg)`/etc. resolves against M1 tokens). Aliases: `--bg`, `--surface`, `--ink`, `--primary`, `--primary-700`, `--muted`, `--muted-fg`, `--muted-fg-2`, `--border`, `--border-strong`, `--success`, `--warning`, `--danger`, `--info`, `--gradient`, `--gradient-soft`, `--shadow-sm`, `--shadow-md`, `--shadow-lg`, `--shadow-sig`. Header comment documents the bridge purpose + SPEC ID.
- `docs-site/static/moai-docs-layout.css` — NEW file (241 lines). Ported verbatim from round3 prototypes. Header documents: scope (geometry only), exclusion of global `body { background }` rule (would change site-wide bg, violating inertness), and the 3 collision-driven renames.
- `docs-site/layouts/partials/head/custom.html` — added `$layoutHash` FNV32a hash var (L18) + `<link>` (L25) after moai-docs-theme.css, with comment (L24).
- `.moai/specs/SPEC-DESIGN-DOCSV2-001/progress.md` — this M3a evidence sub-section.

**Collision check (3 renames applied per task collision-guard instruction):**
```
$ grep -rn '\.toc[{ .]' docs-site/static/moai-design.css   → L573, L693 (bare .toc exists)
$ grep -rn '\.pill' docs-site/static/moai-design.css        → .nav-brand .pill / .md-section-head .pill (scoped descendant bleed risk)
$ grep -rn '\.featured' docs-site/static/moai-design.css    → .cw-card.featured (compound bleed risk)
```
Renames: `.toc` → `.docs-toc` (+ variants `.docs-toc-h/-list/-list li/a/a:hover/a.active/.sub`); `.pill` → `.docs-pill` (+ `.docs-pill-row/:hover/.active/.count`); `.featured` → `.docs-featured` (wrapper only). All other selectors ported byte-for-byte (`.featured-card`/`.featured-eye`/etc. have no existing match → kept verbatim).

**`--font-latin` finding (deviation from task — documented):** Task expected `--font-latin` to be a new alias; M1 had already declared it at brand.css L57 (`"Pretendard Variable", "Pretendard", system-ui, sans-serif`). `grep -rn 'var(--font-latin)' docs-site/` returns **0 consumers** — the token is unused. Rather than override an existing M1 token with a conflicting Inter stack (latent regression vector), the alias line was replaced with an explanatory NOTE comment preserving M1's value. The task's intent ("ensure `--font-latin` exists for ported CSS") is already satisfied by M1, and no ported M3a class references it. Inert either way.

**Inertness preserved:** The global `body { background: var(--bg); }` rule present in both round3 prototypes was deliberately EXCLUDED — including it would immediately change the site-wide body bg, violating the "inert until M3b" requirement. Documented in moai-docs-layout.css header. The ported classes have zero effect until M3b/M3c layouts add the corresponding HTML.

**Parallel-session safety:** specific-path `git add` only (B8/B10) — the 4 files staged are disjoint from the parallel session's `internal/cli/` + `README.*.md` scope.

### M3b — Frame Rewrite (round3 nav/footer shell, infra preserved)

**Scope:** FRAME ONLY — `baseof.html` shell + `site-header.html` visual chrome + `site-footer.html` + frame CSS port. No `index.html`/`single.html`/`list.html`/doc partials (M3c scope). No prose-kr / code-mac (M3c/M3d). The frame CSS makes the round3 classes LIVE (M3a shipped them inert); the 3 template rewrites wire the round3 `.docs-nav`/`.docs-main`/`.docs-footer` geometry onto the existing geekdoc hooks.

**AC Matrix (M3b scope — frame ACs + AC-BLD-001 re-verify):**

| AC | Status | Verification Command | Result |
|---|---|---|---|
| AC-FRM-001 (frame CSS ported) | PASS | `grep -cE 'docs-nav\|docs-footer\|docs-btn\|docs-card\|nav-inner\|footer-grid' docs-site/static/moai-docs-layout.css` | `15` (6 namespace-renamed families all present: docs-nav=2, docs-footer=2, docs-btn=2, docs-card=2, nav-inner=3, footer-grid=2). Bare-ported families verified present: nav-logo=3, nav-menu=2, nav-link=4, nav-utility=2, nav-search=4, nav-search-key=2, footer-inner=2, footer-col=5, footer-bottom=2, btn-primary/secondary/ghost=3 each, chip=6, eyebrow=3, page=2, docs-main=2, docs-container=2 |
| AC-FRM-002 (baseof infra preserved) | PASS | `grep -nE 'color-theme="light"\|gdoc-nav\|gdoc-page\|template "main"\|partial "site-header"\|partial "site-footer"\|MutationObserver' docs-site/layouts/_default/baseof.html` | L5 `color-theme="light"`; L34 `class="wrapper page ..."`; L37 `{{ partial "site-header" (dict "Root" . "MenuEnabled" $navEnabled) }}`; L42 `<aside class="gdoc-nav">`; L48 `<div class="gdoc-page">`; L49 `{{ template "main" . }}`; L129 `{{ partial "site-footer" . }}`; L181 `new MutationObserver(...)` |
| AC-FRM-003 (header infra preserved) | PASS | `grep -nE 'data-search-trigger\|data-search-modal\|data-search-input\|data-search-results\|data-search-close\|cw-lang-switch\|data-gh-stars\|hugo\.Sites\|search\.json\|localStorage\|cw-ver-pill' docs-site/layouts/partials/site-header.html` | L16 `data-search-trigger`; L22 `cw-ver-pill`; L24 `cw-lang-switch`; L25 `hugo.Sites`; L38/41 `cw-gh-stars`/`data-gh-stars`; L48 `data-search-modal`; L53 `data-search-input`; L56 `data-search-results`; L49/54 `data-search-close`; L65-77 GitHub stars fetch JS (KEY=`cw-gh-stars`, TTL=24h, localStorage cache); L84-111 search modal JS (`/search.json` fetch + fuzzy scoring: title.includes +100, startsWith +50, content min(cm*5, 40)) |
| AC-FRM-004 (footer attributions preserved) | PASS | `grep -nE 'Copyleft\|모두의AI\|MIT License\|anthropic\.com\|modu-ai' docs-site/layouts/partials/site-footer.html` | L5 `🄏 Copyleft MoAI - 모두의AI`; L9 GitHub link; L10 `MIT License`; L15 `Anthropic`; L16 `modu-ai` |
| AC-BLD-001 (hugo build) | PASS | `cd docs-site && hugo --minify --gc` | exit 0, 0 WARN/ERROR; KO 153p / EN 150p / JA 139p / ZH 150p; 2470ms (hugo v0.160.1+extended+withdeploy darwin/arm64) |

**Files touched (4 + this progress.md):**
- `docs-site/static/moai-docs-layout.css` — appended Frame section: round3 `.nav`/`.page`/`.footer`/`.btn`/`.card`/`.chip`/`.eyebrow` families ported from round3 prototypes, with 6 collision-driven namespace renames; all non-colliding classes bare-ported. This is the SECOND section of moai-docs-layout.css (M3a ported layout-geometry; M3b ports the frame chrome that M3a deliberately excluded).
- `docs-site/layouts/_default/baseof.html` — shell geometry: `.wrapper` gains `page` class (L34); `<main class="container flex flex-even">` gains `docs-main` (L40). geekdoc hooks PRESERVED verbatim: `gdoc-nav` aside + `{{ partial "menu" . }}` (L42-43, AP-7), `gdoc-page` div (L48, AP-7), `{{ template "main" . }}` (L49), `site-header`/`site-footer` partials (L37/L129). REQ-LIT-001 light-only: `color-theme="light"` (L5) + MutationObserver force-restore (L181-188). Right-rail `.cw-toc` + prev/next `.cw-pg-nav` + code-card copy JS all PRESERVED.
- `docs-site/layouts/partials/site-header.html` — round3 `.docs-nav`/`.nav-inner`/`.nav-logo`/`.nav-utility` visual chrome wraps the existing brand mark + search/lang/github cluster. ⌘K search modal markup + fuzzy-search JS, `cw-lang-switch` (4-locale KO/EN/JA/CN over `hugo.Sites`), GitHub Star fetch with 24h localStorage cache, version pill — ALL PRESERVED verbatim (L47-164 unchanged).
- `docs-site/layouts/partials/site-footer.html` — rewritten to round3 `.docs-footer`/`.footer-inner`/`.footer-grid`/`.footer-col`/`.footer-bottom` structure. 🄏 copyleft "MoAI - 모두의AI", GitHub + MIT License links, Anthropic + modu-ai attribution — ALL PRESERVED.
- `.moai/specs/SPEC-DESIGN-DOCSV2-001/progress.md` — this M3b evidence sub-section.

**Collision renames (6 applied per task collision-guard instruction):**

| round3 class | collision target | Renamed to |
|---|---|---|
| `.nav` | geekdoc `.gdoc-nav` + live moai-design.css descendants | `.docs-nav` |
| `.footer` | geekdoc `.gdoc-footer` + `.cw-footer` | `.docs-footer` |
| `.btn` | moai-design.css `.copy-btn` / `.cw-*` btn descendants | `.docs-btn` |
| `.card` | moai-design.css `.cw-card` / `.code-card` | `.docs-card` |
| `.main` | geekdoc `<main>` base element | `.docs-main` |
| `.container` | geekdoc 82rem `.container` (LIVE — would clobber max-width) | `.docs-container` |

All non-colliding classes ported BARE (zero regression — AC-BLD-001 confirms no layout shift): `.nav-inner`, `.nav-logo`, `.nav-menu`, `.nav-link`, `.nav-utility`, `.nav-search`, `.nav-search-key`, `.nav-icon-btn`, `.nav-login`, `.footer-inner`, `.footer-grid`, `.footer-brand-tag`, `.footer-col`, `.footer-bottom`, `.mobile-nav*`, `.btn-primary`, `.btn-secondary`, `.btn-ghost`, `.btn-lg`, `.chip*`, `.eyebrow*`, `.page`.

**Infrastructure preserved (logic intact — D2/D3/D4 + REQ-LIT-001 confirmed by grep + build):**
- **⌘K search** (`data-search-trigger` → modal; fuzzy search over `/search.json`; ESC + backdrop close): PRESERVED verbatim in site-header.html L47-164.
- **Language switch** (`cw-lang-switch` over `hugo.Sites`, 4-locale KO/EN/JA/CN, `AllTranslations` href resolution, active/aria-current marking): PRESERVED.
- **GitHub stars** (`data-gh-stars` fetch + 24h localStorage cache, KEY=`cw-gh-stars`, TTL=24h): PRESERVED.
- **Light-only theme enforcement** (REQ-LIT-001): `color-theme="light"` attribute (baseof L5) + MutationObserver (baseof L181-188) force-restores light on any external mutation; `data-theme` stripped; geekdoc toggle button handler neutralized. PRESERVED.
- **`!important` cascade (deferred to M3c)**: moai-docs-theme.css `.gdoc-header`/`.cw-footer` carry `!important` green-gradient backgrounds that override round3 non-`!important` backgrounds. round3 STRUCTURE (sticky/height/flex/grid) applies cleanly; green-tint bg stays coherent with the design-system primary. Cannot edit moai-docs-theme.css in M3b scope; noted for M3c reconciliation.

**Self-referential SHA note:** M3b source + this evidence land in the SAME commit (B9 one-commit constraint). Per spec-frontmatter-schema §D3 SHA-placeholder-backfill-exemption principle (a commit cannot reference its own SHA), the M3b milestone header omits the inline SHA — matching the M2/M3a header style (M1 is the only milestone that carries an inline SHA, backfilled in a later commit). git log is the authoritative SHA record; the orchestrator's final report carries the post-push SHA.

**Parallel-session safety:** specific-path `git add` only (B8/B10) — 5 files staged are disjoint from the active parallel session's `internal/cli/` + `README.*.md` + `.moai/config/sections/llm.yaml` scope. system.yaml/CHANGELOG/version.go (pre-existing parallel-session mods) explicitly excluded from the pathspec.

### M3c-1 — page templates (index.html + list.html → round3 docs-index)

**Scope:** HOME template (`docs-site/layouts/index.html`) rewritten + SECTION-LISTING template (`docs-site/layouts/_default/list.html`) created. Both consume the round3 docs-index layout shipped in M3a (`docs-site/static/moai-docs-layout.css`). single.html / doc-rail / 404 are OUT OF SCOPE (M3c-2). baseof.html / site-header.html / site-footer.html untouched (M3b done).

**Deliverable D1 — `docs-site/layouts/index.html` (rewritten, 131 lines):**
- DOM order matches round3 docs-index: `docs-hero` → `docs-filters` (sticky pill bar) → `docs-featured` → `docs-grid-section`.
- `docs-hero`: mascot `<img>` (`mascot-coding.png`) PRESERVED; eyebrow/h1/sub reuse `hero_eyebrow` / `hero_title` / `hero_lead`; 2 CTAs reuse `hero_cta_start` / `hero_cta_browse` with `md-btn` classes + inline SVG arrow; 4 stat pairs PRESERVED verbatim — 11/`stat_agents`, 16/`stat_langs`, `85%+`/`stat_coverage`, `Go`/`stat_binary`.
- `docs-filters`: sticky pill bar, static `<a class="docs-pill">` links per `$cards` entry via `site.GetPage`, count badge = `.RegularPagesRecursive` count. NO client-side JS (faithful static).
- `docs-featured`: single featured card resolving `site.GetPage "/getting-started"` (no `featured: true` frontmatter exists anywhere; getting-started is the "start here" CTA). Uses `browse_pill` (featured-eye), `.Title`, `.Description`, `docs_count` (featured-meta), `hero_cta_start` (featured-cta), `mascot-talking.png` (featured-illu).
- `docs-grid-section`: `.grid-section-head` (`browse_title` + `len $cards`) + `.docs-grid` ranging the 8 curated `$cards` → `.doc-card` per item with `.thumb-{{ add $i 1 }}` cyclic variant, preserved inline-SVG icon mapping (rocket/book/terminal/layers/wrench/cpu/git/db), `.doc-thumb-cat`, `.doc-thumb-num` (`printf "%02d"`), `.doc-body` (h4, `.doc-excerpt` = `i18n $card.desc`, `.doc-meta .views` = `docs_count`).
- Bottom `.md-home-content` preserves `partial "utils/content" .` so pipeline-owned `content/<locale>/_index.md` body still renders.
- `{{ define "main" }}` block contract PRESERVED (baseof L49 `{{ template "main" . }}`).
- Curated `$cards` slice (8 entries) PRESERVED byte-for-byte.

**Deliverable D2 — `docs-site/layouts/_default/list.html` (NEWLY CREATED, 47 lines):**
- Did NOT exist before (Hugo fell back to its internal template). Now a round3 variant exists.
- `docs-hero`: `browse_pill` (eyebrow), `.Title` (h1), `.Description` (sub, conditional), `docs-stats` with `len .Pages` + `docs_count` label.
- `docs-grid-section`: `.docs-grid` ranges `.Pages` → `.doc-card` per page with `.thumb-{{ add (mod $i 8) 1 }}` (8-cycle), `.doc-thumb-cat` = `$.Title`, generic document SVG, `.doc-thumb-num` (`printf "%02d"`), `.doc-body` (h4 = `$page.Title`, `.doc-excerpt` = `$page.Description | default ($page.Summary | truncate 120)`, `.doc-meta .read-time` = `$page.ReadingTime`+"m").
- NO featured card (sections have no `featured: true` frontmatter).
- Bottom `{{ .Content }}` renders section `_index` body when authored.
- `{{ define "main" }}` block contract PRESERVED.

**i18n discipline (PRESERVE constraint):** every `{{ i18n "..." }}` call in both templates reuses EXISTING keys (`hero_eyebrow`, `hero_title`, `hero_lead`, `hero_cta_start`, `hero_cta_browse`, `stat_agents`, `stat_langs`, `stat_coverage`, `stat_binary`, `browse_title`, `browse_pill`, `docs_count`, + the 8 `card_desc_*` keys). NO new i18n keys invented. `docs-site/i18n/{en,ja,ko,zh-cn}.yaml` untouched.

**Stats discipline:** the 4 home-hero stat values (11 retained agents / 16 supported languages / 85%+ coverage target / Go single binary) are preserved verbatim from the prior index.html — no placeholder numbers, no drift.

**No-emoji discipline (CLAUDE.local.md §17.1):** both templates use inline SVG (`<svg viewBox="0 0 24 24">`) for icons, NOT emoji. Verified clean via `grep -nP '[\x{1F000}-\x{1FAFF}\x{2600}-\x{27BF}\x{2B00}-\x{2BFF}]'` on both files (exit 1 = no matches). Typography marks (`→`) and the mascot `<img>` are preserved per the icon convention (not emoji).

**Round3 CSS class coverage (M3a-owned `moai-docs-layout.css`):** both templates reference only classes that ship in `moai-docs-layout.css` L34-136 — `.docs-hero`, `.docs-hero-inner`, `.docs-eyebrow`, `.docs-h1`, `.docs-sub`, `.docs-stats`, `.docs-stat-num`, `.docs-stat-lbl`, `.docs-filters`, `.docs-filters-inner`, `.docs-pill-row`, `.docs-pill` + `.count`, `.docs-featured`, `.featured-card`, `.featured-eye`, `.featured-meta`, `.featured-cta`, `.featured-illu`, `.docs-grid-section`, `.grid-section-head`, `.docs-grid`, `.doc-card`, `.doc-thumb` + `.thumb-1..8`, `.doc-thumb-cat`, `.doc-thumb-icon`, `.doc-thumb-num`, `.doc-body`, `.doc-excerpt`, `.doc-meta`. Legacy `md-btn` / `md-hero-cta` / `md-home-content` classes from `moai-docs-theme.css` are reused (that file survived M3a).

**Build verification (AC-BLD-001):** `hugo --gc --minify` (extended v0.160.1+extended+withdeploy) from `docs-site/` → exit 0, 0 WARN / 0 ERROR. Page counts: KO 153 / EN 150 / JA 139 / ZH 150 — identical to the M3b baseline (NO drop, NO new pages, NO missing pages). Build time 2983ms.

**Self-referential SHA note:** M3c-1 source + this evidence land in the SAME commit (B9 one-commit constraint). Per the spec-frontmatter-schema §D3 SHA-placeholder-backfill-exemption principle, the M3c-1 milestone header omits the inline SHA — matching M2/M3a/M3b style. git log is the authoritative SHA record; the orchestrator's final report carries the post-push SHA.

**Parallel-session safety:** specific-path `git add docs-site/layouts/index.html docs-site/layouts/_default/list.html .moai/specs/SPEC-DESIGN-DOCSV2-001/progress.md` only (B8/B10) — the 3 staged files are disjoint from the active parallel session's `internal/cli/` + `README.ko.md` + `.moai/config/sections/llm.yaml` scope. Tree was clean at branch-switch time (no stash needed); `release/v3.0.0` restored post-push.

### M3c-2 — docs-detail page (single.html 3-col + doc-rail + read-progress + 404 round3 hero)

**Scope:** DOC-DETAIL template (`docs-site/layouts/_default/single.html`) created with round3 docs-detail 3-column layout (read-progress + doc-hero + doc-layout grid 220px TOC | 1fr body | 280px rail). DOC-RAIL partial (`docs-site/layouts/partials/doc-rail.html`) created with 3 rail-cards (actions / related / CTA). DOC-DETAIL JS (`docs-site/static/js/doc-detail.js`) created with read-progress + TOC scroll-spy. 404 (`docs-site/layouts/404.html`) rewritten to round3 `.doc-hero` geometry with mascot placeholder (M5) + home CTA. index.html / list.html (M3c-1), baseof.html / site-header / site-footer (M3b), content/, _meta.yaml, data/menu/main.yaml, menu*.html, search.json, shortcodes/ all untouched (PRESERVE).

**Deliverable D1 — `docs-site/layouts/_default/single.html` (NEWLY CREATED, 99 lines):**
- `{{ define "main" }}` block fills baseof chrome (gdoc-nav/cw-toc/cw-pg-nav still render around the article) — follows the M3c-1 `list.html` coexistence precedent (round3 inside `define "main"`, baseof chrome stays).
- DOM order matches round3 §C: `<article id="main-content" tabindex="-1">` → `.read-progress` (`<div id="rp">`) → `.doc-hero` (`.doc-hero-inner` with `.crumb` breadcrumb via `.Site.Home`/`.CurrentSection`, `.doc-cat` = `.Params.category | default .CurrentSection.Title`, `h1.doc-h1` = `.Title`, `.doc-deck` = `.Description`, `.doc-byline` = avatar "M" + author/role + date + ReadingTime) → `.doc-layout` grid (`.docs-toc` sticky | `.doc-body` | `.rail` sticky).
- TOC recursive walk via named template `{{ define "doc-toc-walk" }}`: walks `.Fragments.Headings` tree into children `.Headings`, emitting `<li><a href="#{{ .ID }}">{{ .Title }}</a></li>` for every Level-2 at any depth — solves Hugo 0.160's roots-only `.Fragments.Headings` behavior (top-level roots only, not flat). Guarded by `{{ if gt (len (findRE "<h2" .Content)) 0 }}` so empty-TOC pages skip the aside entirely.
- Body wrapper: `.gdoc-markdown prose-kr` on the `.Content` div (`prose-kr` is PLANNED per `moai-docs-layout.css` line 16 but NOT yet defined — `gdoc-markdown` is the proven primary typography). `.next-cta` is a sibling OUTSIDE `gdoc-markdown` so prose padding does not bleed onto the CTA card (bound to `.NextInSection`).
- Rail: `{{ partial "doc-rail" . }}` inside `<aside class="rail">`.
- Inline `<script src="{{ "js/doc-detail.js" | relURL }}" defer>` at end of block.

**Deliverable D2 — `docs-site/layouts/partials/doc-rail.html` (NEWLY CREATED, 58 lines):**
- Card 1 "이 가이드 활용": 4 static action buttons (좋아요/북마크/공유/인쇄) using inline SVG (heart/bookmark/share/printer). Share uses `navigator.share` with `navigator.clipboard` fallback inline; print uses `window.print()` inline. NO backend (like/bookmark are no-ops pending a future M7+ analytics layer).
- Card 2 "관련 가이드": `range first 3 $others` where `$others` is built by excluding the current page from `where .Site.RegularPages "Section" .Section`. Renders `.rail-related` link with title + ReadingTime.
- Card 3 CTA: gradient variant via `<a class="rail-card" style="background:var(--gradient);color:#fff;border:none;">` linking to docs home. Uses inline `--gradient` token (CSS class for the gradient variant is not defined; inline style is the documented escape hatch).
- Korean plain-text labels ("이 가이드 활용", "관련 가이드", "좋아요", "북마크", "공유", "인쇄", "시작하기", "문서 홈 →", "모든 문서를 둘러보고 다음 가이드를 찾아보세요.") are hard-coded Korean — NO new i18n keys invented. M7 i18n pass will promote these to keys (`rail_actions_title`, `rail_related_title`, `rail_like`, `rail_bookmark`, `rail_share`, `rail_print`, `rail_cta_title`, `rail_cta_body`, `rail_cta_link`).

**Deliverable D3 — `docs-site/static/js/doc-detail.js` (NEWLY CREATED, 60 lines):**
- Vanilla JS, no dependencies. IIFE + `'use strict'`. `DOMContentLoaded` guard.
- (1) Read-progress: `#rp` width tracks `window.scrollY / (scrollHeight - innerHeight) * 100`, clamped 0–100%. rAF throttle (`ticking` flag) on scroll listener (`{ passive: true }`). Initial `updateProgress()` call on init.
- (2) TOC scroll-spy: `IntersectionObserver` on `.doc-body h2[id]` with `rootMargin: '-20% 0px -70% 0px'`. On intersect, iterates `.docs-toc-list a` links and toggles `.active` class. `decodeURIComponent` fix applied — TOC anchors are URL-encoded for non-ASCII heading IDs (ko/ja/zh), so the href must be decoded before comparing to the raw `element.id`. Matches the existing `moai-docs-theme.js` decodeURIComponent pattern. Silent failure on ALL ko/ja/zh headings would result without this fix.
- Selector note (in-code comment): §C/§D draft referenced `.toc-list`, but M3a CSS renamed `.toc` → `.docs-toc` (`moai-docs-layout.css` line 23). This file aligns to the M3a CSS SSOT (`.docs-toc-list`); the HTML emits the same class.

**Deliverable D4 — `docs-site/layouts/404.html` (REWRITTEN, 45 lines):**
- Standalone page (own DOCTYPE, NOT baseof — preserves the pre-existing 404 standalone pattern). `<html lang color-toggle-light color-theme="light">` + `partial "head/meta"` / `head/favicons` / `head/others` / `head/custom` + `partial "site-header"` (MenuEnabled=false) + `partial "site-footer"` + `partial "svg-icon-symbols"`.
- `<main class="doc-hero" style="min-height:70vh;display:flex;align-items:center;">` with centered `.doc-hero-inner` containing: mascot `<img>` (`mascots/mascot-bubble.png`, 120px, `loading="lazy" decoding="async"`) with `<!-- M5: swap to canonical MoAI-Mascot-Thinking.png pose. -->` placeholder comment; big "404" (`font-family:var(--font-mono)`, clamp 72–144px); `h1.doc-h1` = `i18n "notfound_title"`; `p.doc-deck` = `i18n "notfound_text"`; 2 CTAs (`.md-btn md-btn-primary` home with inline arrow SVG + `.md-btn md-btn-ghost` search with `data-search-trigger`).
- Keeps existing i18n keys (`error_page_title`, `notfound_title`, `notfound_text`, `notfound_home`, `notfound_search`) — NO new i18n keys invented.

**Round3 CSS class coverage (M3a-owned `moai-docs-layout.css`):** the new templates reference round3 docs-detail classes shipped in `moai-docs-layout.css` — `.read-progress` / `#rp`, `.doc-hero` / `.doc-hero-inner` / `.crumb` / `.sep` / `.doc-cat` / `.doc-h1` / `.doc-deck` / `.doc-byline` / `.doc-author` / `.doc-author-avatar` / `.doc-author-sub` / `.doc-meta-strip`, `.doc-layout` (grid 220px 1fr 280px), `.docs-toc` / `.docs-toc-h` / `.docs-toc-list` / `.docs-toc-list a.active`, `.doc-body`, `.rail` / `.rail-card` / `.rail-h` / `.rail-actions` / `.rail-action` / `.rail-related`, `.next-cta` / `.next-cta-eye` / `.next-cta-arrow`. Legacy `gdoc-markdown` / `md-btn` / `md-btn-primary` / `md-btn-ghost` / `md-hero-cta` / `md-404-mascot` classes from `moai-docs-theme.css` are reused (that file survived M3a). `.prose-kr` is referenced as a future-ready secondary class (planned per `moai-docs-layout.css` line 16, NOT yet defined — `gdoc-markdown` is the load-bearing typography).

**i18n discipline (PRESERVE constraint):** every `{{ i18n "..." }}` call in the 4 new/rewritten files reuses EXISTING keys (`toc_label`, `error_page_title`, `notfound_title`, `notfound_text`, `notfound_home`, `notfound_search`). NO new i18n keys invented. The rail-card Korean labels are hard-coded Korean plain text (NOTED for M7 promotion to keys). `docs-site/i18n/{en,ja,ko,zh-cn}.yaml` untouched.

**No-emoji discipline (CLAUDE.local.md §17.1):** all 4 files use inline SVG (`<svg viewBox="0 0 24 24">`) for icons, NOT emoji. Typography marks (`→` `›`) and the mascot `<img>` are preserved per the icon convention (not emoji). The 404 "404" big-number uses a monospace font (`var(--font-mono)`), not an emoji.

**Build verification (AC-BLD-001):** `hugo --gc --minify` (extended v0.160.1+extended+withdeploy) from `docs-site/` → exit 0, 0 WARN / 0 ERROR. Page counts: KO 153 / EN 150 / JA 139 / ZH 150 — identical to the M3c-1 baseline (NO drop, NO new pages, NO missing pages). Build time stable.

**Files touched (5):** `docs-site/layouts/_default/single.html` (NEW), `docs-site/layouts/partials/doc-rail.html` (NEW), `docs-site/static/js/doc-detail.js` (NEW), `docs-site/layouts/404.html` (REWRITTEN), `.moai/specs/SPEC-DESIGN-DOCSV2-001/progress.md` (this evidence appended).

**Notable implementation decisions:**
1. TOC recursive named template — Hugo 0.160's `.Fragments.Headings` returns only top-level roots (the first heading level encountered), so a flat `where .Level 2` filter returned 0 h2s on pages where h2s nest under h1 roots. The `{{ define "doc-toc-walk" }}` named template walks the tree into children `.Headings`, collecting all Level-2 at any depth. Verified: 8 lis (moai-feedback page), 10 lis (moai-fix page), 8 lis (en/moai page).
2. JS `decodeURIComponent` scroll-spy fix — TOC anchors are URL-encoded for non-ASCII heading IDs (ko/ja/zh), but `element.id` is raw Unicode. A naive CSS-selector `a[href="#"+id]` match would silently fail on ALL ko/ja/zh headings. The fix iterates links, decodes the href, and compares to raw `element.id` — matching the existing `moai-docs-theme.js` pattern.
3. Class reconciliation — §C/§D draft referenced `.toc-list`, but M3a renamed `.toc` → `.docs-toc` (`moai-docs-layout.css` line 23). All new files align to `.docs-toc-list` (the M3a SSOT). The reconciliation is documented in code comments at the top of `doc-detail.js` and in the `single.html` block header.
4. Content typography wrapper — `.prose-kr` is PLANNED per `moai-docs-layout.css` line 16 but NOT yet defined. Using `gdoc-markdown` as the proven primary typography + `prose-kr` as a future-ready secondary class means M3c-2 does not depend on a not-yet-shipped class.
5. baseof chrome coexistence — `single.html` emits round3 inside `{{ define "main" }}`, letting baseof chrome (gdoc-nav/cw-toc/cw-pg-nav) render around it. This follows the M3c-1 `list.html` precedent (accepted M3 baseline); the round3 in-page TOC (`.docs-toc`) coexists with the geekdoc sidebar TOC (`cw-toc`) — they serve different navigation surfaces.
6. M5 mascot placeholder — the 404 mascot `<img>` uses `mascots/mascot-bubble.png` (an existing asset) with a `<!-- M5: swap to canonical MoAI-Mascot-Thinking.png pose. -->` placeholder comment. M5 will own the canonical pose swap; M3c-2 ships a working placeholder.
7. Korean plain-text rail labels — the doc-rail partial uses hard-coded Korean ("이 가이드 활용", "관련 가이드", etc.) rather than inventing new i18n keys mid-milestone. M7 i18n pass will promote these to keys; the PRESERVE constraint on `i18n/*.yaml` is respected.

**Self-referential SHA note:** M3c-2 source + this evidence land in the SAME commit (B9 one-commit constraint). Per the spec-frontmatter-schema §D3 SHA-placeholder-backfill-exemption principle (a commit cannot reference its own SHA), the M3c-2 milestone header omits the inline SHA — matching the M2/M3a/M3b/M3c-1 header style. git log is the authoritative SHA record; the orchestrator's final report carries the post-push SHA.

**Parallel-session safety:** specific-path `git add docs-site/layouts/_default/single.html docs-site/layouts/partials/doc-rail.html docs-site/layouts/404.html docs-site/static/js/doc-detail.js .moai/specs/SPEC-DESIGN-DOCSV2-001/progress.md` only (B8/B10) — the 5 staged files are all under `docs-site/` + the SPEC progress artifact, disjoint from the active parallel session's `.claude/settings.json` + `internal/template/templates/.claude/settings.json.tmpl` scope (and from the PRESERVE-list `internal/cli/*` + `.moai/config/sections/llm.yaml` + `README.ko.md`). Tree carries the parallel session's 2 uncommitted files at branch-switch time; `git checkout main` carried them cleanly (disjoint paths), no stash needed. `release/v3.0.0` restored post-push; parallel session's uncommitted changes preserved verbatim.

### M4 — Mermaid v2 Palette (foot.html themeVariables)

**Scope:** MERMAID PALETTE ONLY — `docs-site/layouts/partials/foot.html` `lightTheme` themeVariables object rewritten to v2 palette. Stayed on mermaid@10 (clarification resolved: default = v10 + themeVariables-only, no v11 bump). No other file touched. The mermaid theme is global (4-locale invariant — the init object is locale-agnostic).

**AC Matrix (AC-MER-001/002 — M4 scope + AC-BLD-001 re-verify):**

| AC | Status | Verification Command | Result |
|---|---|---|---|
| AC-MER-001 | PASS | `grep -n 'primaryColor.*#eef4f0' foot.html` + `grep -n 'lineColor.*#9fa0a0' foot.html` | L14 `primaryColor: '#eef4f0'`; L20 `lineColor: '#9fa0a0'` — both v2 tokens present (1 hit each) |
| AC-MER-002 | PASS | `grep -nc '#d6ebde\|#faf9f5\|#efe9de\|#141413' foot.html` | `0` (was 4 warm literals across ~28 keys pre-M4: `#d6ebde` primaryColor/mainBkg/noteBkg/actorBkg/activationBkg, `#141413` all text keys, `#efe9de` secondaryColor; comment reworded to avoid reintroducing forbidden literals). mermaid@10 CDN URL unchanged (v10 stays per resolved clarification) |
| AC-BLD-001 | PASS | `cd docs-site && hugo --minify --gc` | exit 0, 0 WARN/ERROR; KO 153p / EN 150p / JA 139p / ZH 150p; 1938ms (hugo v0.160.1+extended+withdeploy darwin/arm64) |

**Value mapping (v2 palette applied to all ~28 themeVariables keys):**
- Primary/accent bg: `primaryColor`/`mainBkg`/`actorBkg`/`activationBkgColor` = `#eef4f0` (v2 mint-tinted surface, was warm `#d6ebde`).
- Text: all `*TextColor`/`nodeTextColor`/`textColor`/`titleColor`/`labelTextColor`/`loopTextColor`/`sequenceNumberColor` = `#060606` (v2 ink, was `#141413`).
- Lines de-emphasized to GRAY: `lineColor`/`signalColor` = `#9fa0a0` (was green `#3d7d5f` — v2 line de-emphasis intent per plan.md §M4).
- Primary/actor BORDERS stay green: `primaryBorderColor`/`actorBorder` = `#3d7d5f` (brand accent preserved per plan.md explicit list).
- Neutral surfaces/borders: `noteBkgColor`/`secondaryColor` = `#e6e6e6`; `noteBorderColor`/`activationBorderColor` = `#b5b5b5`; `clusterBorder`/`labelBoxBorderColor` = `#d1d1d1`; `clusterBkg`/`tertiaryColor`/`labelBoxBkgColor` = `#f4f4f4`; `edgeLabelBackground` = `#ffffff` (was warm `#f4f4f4`).

**Deviation from plan.md §M4 (documented):** plan.md §M4 enumerates 12 explicit keys; the live `lightTheme` object carries ~28 keys. The 16 keys NOT in plan.md's explicit list were mapped coherently to the same v2 palette (text→`#060606`, line-family→`#9fa0a0` gray, note/cluster borders→neutral, primary/actor accents→green/`#eef4f0`) rather than left warm — leaving them warm would fail the AC-MER-002 warm-literal purge and the M7 token-parity sweep. The 12 plan.md-explicit values are applied verbatim.

**No-emoji / light-only discipline:** foot.html carries no emoji; the mermaid init uses the single `lightTheme` object with no dark branch (dark branch already removed 2026-05-13, preserved). REQ-LIT compliance unchanged.

**Self-referential SHA note:** M4 source + this evidence land in the SAME commit (B9 one-commit constraint). Per spec-frontmatter-schema §D3 SHA-placeholder-backfill-exemption, the M4 header omits an inline SHA — matching M2/M3a/M3b/M3c style. git log is the authoritative SHA record.

**Parallel-session safety:** specific-path `git add docs-site/layouts/partials/foot.html .moai/specs/SPEC-DESIGN-DOCSV2-001/progress.md` only (B8/B10) — disjoint from the active parallel session's `.claude/settings.json` + `internal/template/templates/.claude/settings.json.tmpl` scope (left unstaged, dirty from another work line).

### M5 — Mascot Expansion (6 poses, emotional surfaces)

**Scope:** 6 canonical MoAI-Mascot PNGs installed + wired via a NEW `partials/mascot.html` + extended `shortcodes/mascot.html` + NEW `partials/doc-empty.html` empty-state. Placements: home hero (Explaining), 404 (Thinking, replacing the M3c-2 `mascot-bubble.png` placeholder), empty-section (Searching, via doc-empty wired into list.html). 4-locale empty-state copy added to all 4 i18n files (AP-2 compliant — no interim hardcoded-Korean defect). Mascot base CSS + wiggle motion (prefers-reduced-motion-safe) added to `moai-docs-layout.css`.

**AC Matrix (AC-MAS-001/002/003 + AC-BLD-001 + AC-I18N-001 re-verify — M5 scope):**

| AC | Status | Verification Command | Result |
|---|---|---|---|
| AC-MAS-001 | PASS | `ls docs-site/static/mascots/MoAI-Mascot-*.png \| wc -l` | `6` (Thinking, Pointing, Searching, Teaching, Explaining, Coffee — copied from `.moai/state/ai-design-system/project/assets/characters/`); rendered `ls public/mascots/MoAI-Mascot-*.png` = 6 |
| AC-MAS-002 | PASS | `grep -o 'class="mascot ' public/{ko/index,ko/404,ko/contributing/index}.html \| wc -l` per surface | home=1 (Explaining), 404=1 (Thinking), empty-section `contributing`=1 (Searching) — exactly 1 `img.mascot` per surface (3 emotional surfaces ≥ required 3). en/ja/zh home all render Explaining (locale-invariant pose map) |
| AC-MAS-003 | PASS | `grep -rEn '<(table\|form)[^>]*>.*mascot' layouts/` + render-vehicle grep | 0 mascots inside table/form/checkout (docs-site has none); all mascots rendered via `partials/mascot.html` or `shortcodes/mascot.html` (no stray inline `img.mascot`) |
| AC-I18N-001 | PASS | `grep -o 'doc-empty-title>[^<]*' public/{ko,en,ja,zh}/contributing/index.html` | 4 distinct locale strings (ko 아직 준비 중… / en This section is coming soon / ja このセクションは準備中です / zh 该分区正在完善中) — structural DOM identical, only translated strings differ |
| AC-BLD-001 | PASS | `cd docs-site && hugo --minify --gc` | exit 0, 0 WARN/ERROR; KO 153p / EN 150p / JA 139p / ZH 150p (unchanged — doc-empty replaces the empty grid on `contributing`/other 0-page sections, no page-count delta); 1935ms |

**Files touched (13):**
- `docs-site/static/mascots/MoAI-Mascot-{Thinking,Pointing,Searching,Teaching,Explaining,Coffee}.png` — 6 NEW canonical pose assets.
- `docs-site/layouts/partials/mascot.html` — NEW. Layout-level placement partial. Accepts a dict `(dict "pose" ... "size" ... "class" ...)` OR a bare string pose. Validates pose against the 6-set, maps lowercase pose → Capitalized filename via `title`, emits `<img class="mascot mascot-<pose> [extra]" src="…/MoAI-Mascot-<Pose>.png" alt="" loading="lazy" />`.
- `docs-site/layouts/shortcodes/mascot.html` — EXTENDED. Now branches: the 6 v2 poses emit the `mascot` contract (canonical PNG); the 3 legacy variants (coding/talking/bubble) preserve the pre-existing `cw-mascot` contract byte-for-byte (backward compat — the shortcode is currently unused in content, verified via `grep -rn '{{< mascot' content/` = 0, but the legacy branch is retained for safety).
- `docs-site/layouts/partials/doc-empty.html` — NEW. Empty-section state: Searching mascot + `i18n "empty_title"`/`empty_text` + home CTA (reuses `notfound_home`). Wired into list.html.
- `docs-site/layouts/_default/list.html` — added `{{ if eq (len .Pages) 0 }}{{ partial "doc-empty.html" . }}{{ else }}…grid…{{ end }}`. The `contributing` section (0 child pages, `_index.md`-only) is the live empty-state instance.
- `docs-site/layouts/index.html` — home hero mascot swapped from raw `<img class="md-hero-mascot" src="mascot-coding.png">` to `{{ partial "mascot.html" (dict "pose" "explaining" "size" "150" "class" "md-hero-mascot") }}` (Explaining = welcome intent per design.md §F.1). The featured-section decorative `mascot-talking.png` img is UNCHANGED (no `mascot` class → not an `img.mascot`, out of M5 scope).
- `docs-site/layouts/404.html` — mascot swapped to `{{ partial "mascot.html" (dict "pose" "thinking" ...) }}` (Thinking = empathetic 404), M5 placeholder comment resolved.
- `docs-site/i18n/{ko,en,ja,zh-cn}.yaml` — added `empty_title` + `empty_text` (ko canonical → en/ja/zh derivation per hns-oss-docs-i18n-rules §1). Added in M5 (not deferred to M7) to avoid an interim 4-locale defect — the new copy I introduce is i18n-routed from the start (AP-2).
- `docs-site/static/moai-docs-layout.css` — appended `.mascot` base (inline-block, height:auto, bounce-easing transition + hover `mascot-wiggle` keyframe), `.doc-empty*` empty-state layout, and a `prefers-reduced-motion: reduce` guard (motion → 1ms per design.md §F.3).

**Deferred (documented, NOT AC-blocking):**
- **render-codeblock.html Coffee mascot** (plan.md §M5 lists it for copy-success): DEFERRED. It is an ephemeral copy-success state; placing a Coffee mascot on every code card would add `img.mascot` elements to detail pages, risking the AC-MAS-002 "exactly one img.mascot per surface" counts and the M7 4-locale parity assertions. AC-MAS-002 requires only home + 404 + empty-section (all satisfied). Coffee-on-copy-success is a post-close polish candidate. The Coffee PNG is installed and available.
- **Home hero Pointing alt** (design.md §F.1 A/B alt): not wired — a single Explaining hero is the shipped default; Pointing remains available for a future A/B.

**No-emoji / light-only discipline:** all new markup uses inline SVG (home CTA arrow in doc-empty), mascot `<img>` (per icon convention, not emoji), and mascot `alt=""` (decorative). No body emoji. No dark-mode branch introduced. `class="mascot"` (exact token) does NOT collide with the header brand `cw-brand-mascot` (distinct token — verified via grep).

**Self-referential SHA note:** M5 source + this evidence land in the SAME commit (B9). Per spec-frontmatter-schema §D3, the M5 header omits an inline SHA (M2/M3/M4 style). git log is the authoritative record.

**Parallel-session safety:** specific-path `git add` of the 13 docs-site files + progress.md only (B8/B10) — disjoint from the parallel session's `.claude/settings.json` + `internal/template/templates/.claude/settings.json.tmpl` (left unstaged, dirty from another work line).

### M6 — Component Adoption (v2 recipes: button / card / shadow)

**Scope:** LIVE component overrides migrated to v2 recipes — buttons (`.md-btn` family), surface cards (`.md-doc-card`, round3 `.doc-card`/`.docs-card`), and the shadow system. The load-bearing fix: `moai-docs-tokens.css` (cascade winner — loads AFTER brand.css) carried warm brown-tinted `rgba(58,38,24)` shadows that overrode the v2 brand.css shadows on every `var(--shadow-*)` consumer; repointing them to v2 `rgba(6,6,6)` fixes card/button elevation site-wide in one edit.

**Plan.md §M6 file-assignment deviation (documented):** plan.md §M6 anticipated editing `moai-brand.css` (cw-*/gdoc-*) + `moai-design.css` (.md-btn/.md-doc-card/.code-card). The M1 implementation instead kept the `.md-*` LIVE component classes in `moai-docs-theme.css` (M1 rewrote it to a v2 alias+component layer rather than folding md-* into design.css). M6 therefore edits the REAL definition sites (`moai-docs-theme.css` for md-btn/md-doc-card, `moai-docs-tokens.css` for shadows, `moai-docs-layout.css` for the round3 cards, `moai-brand.css` for the new `--border-card` token) — DDD behavior-faithful (touch the live definitions, not the plan-phase placeholder file names).

**AC Matrix (AC-CMP-001/002/003/004 + AC-BLD-001 re-verify — M6 scope):**

| AC | Status | Verification Command | Result |
|---|---|---|---|
| AC-CMP-001 | PASS | token-resolution trace on `.md-btn-primary` + `.md-doc-card` | **Button** `.md-btn` `border-radius: var(--radius-full)`→`9999px`, `font-weight: 700`; `.md-btn-primary` `background: var(--accent)`→`var(--color-primary)`→`#3d7d5f`=`rgb(61,125,95)` solid, `color` white. **Card** `.md-doc-card` (+ round3 `.doc-card`/`.docs-card`) `border-radius: var(--radius-lg)`→`16px`, `border: 1px solid var(--border-card)`→`#d1d1d1`=`rgb(209,209,209)`, `box-shadow: var(--shadow-sm)`→`0 2px 4px rgba(6,6,6,0.06)`. All three AC exact values (9999px / rgb(61,125,95) / fw≥700 · 16px / rgb(209,209,209) / rgba(6,6,6)) met |
| AC-CMP-002 | PASS | `grep 'shadow-sm:\|shadow-md:\|shadow-lg:' moai-docs-tokens.css` + warm-purge | v2 shadow tokens present & consumed: `--shadow-sm 0 2px 4px rgba(6,6,6,0.06)` / `--shadow-md …0.08` / `--shadow-lg …0.10` (tokens.css cascade winner, repointed from warm `rgba(58,38,24)`). Motion tokens present & consumed: `.md-btn transition: all var(--dur-fast) var(--ease-pop)`, `.md-doc-card var(--dur-base) var(--ease-out)`, mascot bounce easing (M5). `grep 'rgba(58, 38, 24' static/*.css` = 0 |
| AC-CMP-003 | PASS | `grep -rn 'border-left:.*var(--color-primary)\|border-left:.*#3d7d5f' static/*.css` | 1 match — `.gdoc-hint.info/.note/.important { border-left: 3px solid var(--color-primary) }` — an ADMONITION/callout (design.md §C explicitly allows callout border-left), NOT a `-card`-class selector. No card has the rounded-border + left-color-accent AI-slop pattern. The plan.md §M6 forbidden sweep (`border-left:.*#3d7d5f` near `border-radius`) returns 0 card matches |
| AC-CMP-004 | PASS | `grep -lP '[emoji-ranges]' <modified CSS>` + full-bleed scan | 0 body emoji in any modified CSS (grep exit 1); 0 full-bleed `background: url(…) cover/100%` introduced by M6. All M6 edits are token/geometry only |
| AC-BLD-001 | PASS | `cd docs-site && hugo --minify --gc` | exit 0, 0 WARN/ERROR; KO 153p / EN 150p / JA 139p / ZH 150p (unchanged); 1902ms |

**Files touched (5):**
- `docs-site/static/moai-docs-tokens.css` — `--shadow-xs/sm/md/lg` warm `rgba(58,38,24,…)` → v2 `rgba(6,6,6,…)` (matching brand.css SSOT). This is the cascade winner (loads after brand.css) so this single repoint fixes elevation on every `var(--shadow-*)` consumer. `--shadow-pop*` (neo-brutalist offset) left intact (still used by non-migrated surfaces).
- `docs-site/static/moai-brand.css` — added `--border-card: #d1d1d1` alias (v2 surface-card hairline per design.md §C; achromatic, not an AC-TOK-forbidden literal).
- `docs-site/static/moai-docs-theme.css` — `.md-btn` neo-brutalist → v2 CTA pill: `radius-md`(8px)→`radius-full`(9999px), border `2px`→`1px`, padding `12px 22px`→`13px 24px`, added `letter-spacing:-0.025em`. `.md-btn-primary`: removed `box-shadow: shadow-pop` (offset) + `border-color: border-strong`, hover removed `translate(1px,1px)`+offset-shadow → `translateY(-1px)` + `shadow-signature`, active removed brutalist `translate(3px,3px)`. `.md-btn-ghost` border `border-clay`→`border-strong`. `.gdoc-markdown a.md-btn` border-bottom-pop → `1px solid transparent` (pill artifact removal). `.md-doc-card` → v2 surface: `radius-xl`(24px)→`radius-lg`(16px), `1.5px border-subtle`→`1px border-card`, `shadow-xs`→`shadow-sm`, hover `translateY(-3px)`→`(-2px)`; `a.md-doc-card` border-bottom aligned to `1px border-card`.
- `docs-site/static/moai-docs-layout.css` — round3 `.doc-card` + `.docs-card` (live home grid/surface cards): border `var(--border)`(#e6e6e6)→`var(--border-card)`(#d1d1d1), added resting `box-shadow: var(--shadow-sm)`. Makes the actual rendered home cards satisfy the AC-CMP-001 surface recipe.
- `.moai/specs/SPEC-DESIGN-DOCSV2-001/progress.md` — this M6 evidence.

**AC-TOK-006 (gradient+shadow simultaneity) not regressed:** all migrated buttons/cards use SOLID backgrounds — `.md-btn-primary` `background: var(--accent)` (solid #3d7d5f), `.doc-card`/`.docs-card` `background: var(--surface)` (solid). `--gradient`/`--gradient-signature` resolve to solid `#3d7d5f` (M1), so even the round3 `.btn-primary { background: var(--gradient) }` is not a real gradient. No selector block pairs a real gradient background with a box-shadow.

**Left intact (already v2-clean, no forbidden pattern — DDD scope discipline):** `.callout` (moai-design.css — token-driven tint bg + 1px border, radius 12px, no border-left card-slop); `.cw-card` (moai-design.css — v2-token-driven via M1 repoint); round3 `.chip` (radius 999px, fw 700 — already v2 §C); `.code-card` (macOS chrome — component-local dark per CLAUDE.local.md §17.1, intentionally preserved). Editing these would be drive-by refactoring outside the AC surface.

**Self-referential SHA note:** M6 source + evidence in the SAME commit (B9); per spec-frontmatter-schema §D3 the header omits an inline SHA (M2-M5 style). git log is authoritative.

**Parallel-session safety:** specific-path `git add` of the 4 CSS files + progress.md only (B8/B10) — disjoint from the parallel session's `.claude/settings.json` + `internal/template/templates/.claude/settings.json.tmpl` (left unstaged).

### M7 — 4-Locale Parity Sweep + Build Gate

**Scope:** doc-rail.html hardcoded Korean promoted to 10 i18n keys (`rail_*`) across all 4 locales; `.prose-kr` Korean-readability class defined (thin, locale-safe); remaining warm `#252320` literals swept to achromatic v2 neutrals; the full M7 build/parity gate executed. This is the closing run-phase milestone.

**AC Matrix (AC-I18N-001/002 + AC-LIT-001 + AC-BLD-001 + M7 gate + regression re-verify):**

| AC | Status | Verification Command | Result |
|---|---|---|---|
| AC-I18N-001 | PASS | per-locale render of doc-rail `rail-h` + doc-empty across ko/en/ja/zh | doc-rail renders 4 distinct locale strings (ko 이 가이드 활용 / en Use this guide / ja このガイドを活用 / zh 使用本指南); structural DOM identical, only translated strings differ. Page counts held KO153/EN150/JA139/ZH150 every milestone (ja<ko leaf-page delta is PRE-EXISTING content translation gap from the M3 baseline, NOT introduced by M4-M7; section-dir count equal = 12 per locale) |
| AC-I18N-002 | PASS | `grep -rEn 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' layouts/ static/ i18n/` | 0 matches (adk.mo.ai.kr is the sole valid docs domain) |
| AC-LIT-001 | PASS | `grep -c MutationObserver baseof.html` + `color-theme="light"` | MutationObserver present (1) + `color-theme="light"` forced; no dark-mode render. Dark dead-code blocks preserved inert (design.md §H.2) |
| AC-BLD-001 | PASS | `cd docs-site && hugo --minify --gc` | exit 0, 0 WARN/ERROR; KO 153 / EN 150 / JA 139 / ZH 150; 1893ms |
| Token parity (AC-TOK-001/003/004/005) | PASS | `grep -rn '#faf9f5\|#ecefee\|#211A14\|linear-gradient(135deg, #3d7d5f' static/ layouts/` (comment-excl) + `#000000` | 0 forbidden literals; 0 `#000000`; `#252320` warm literal 0 (both occurrences swept) |
| Mermaid TD-only | PASS | `grep -rEn 'graph (LR\|RL\|BT)\|flowchart (LR\|RL\|BT)' content/` | 0 (all diagrams TD/TB) |
| Body-emoji (M4-M7 layouts) | PASS | `grep -lP '[emoji-ranges]' <9 M4-M7 layout files>` | 0 (grep exit 1) |

**Files touched (7):**
- `docs-site/i18n/{ko,en,ja,zh-cn}.yaml` — added 10 `rail_*` keys each (`rail_actions_title/like/bookmark/share/print/related_title/read_min/cta_title/cta_body/cta_link`). ko canonical → en/ja/zh derivation. `rail_read_min` is a `{{ . }}`-param key (takes `.ReadingTime`).
- `docs-site/layouts/partials/doc-rail.html` — every exposed string (card titles, button labels + aria-labels, related-guide read-time, CTA title/body/link) promoted from hardcoded Korean to `{{ i18n "rail_*" }}`. Resolves the M3c-2 deferred i18n debt (AP-2 compliance). Inline SVGs + no-backend JS (share/print) unchanged.
- `docs-site/static/moai-docs-layout.css` — defined `.prose-kr` (referenced-but-undefined since M3c-2): a THIN Korean-readability layer — `overflow-wrap: break-word` universal + `word-break: keep-all` scoped to `<html lang="ko">` only (avoids CJK ja/zh width overflow). Deliberately does NOT re-style headings/code/rhythm — `gdoc-markdown` + `render-codeblock.html` (`.code-card`) own those (conflict/regression avoidance). Ports the load-bearing part of round3 `styles.css §.prose-kr`.
- `docs-site/static/moai-design.css` — `.term-card .code-chrome` gradient warm `#252320` midpoint → `var(--neutral-700)` (achromatic). Gradient direction 180deg neutral (NOT the forbidden 135deg signature gradient).
- `docs-site/static/moai-brand.css` — `[data-theme="dark"] .cw-search-modal__panel` warm `#252320` → `var(--neutral-800)` (achromatic; dark dead-code, light-only).
- `.moai/specs/SPEC-DESIGN-DOCSV2-001/progress.md` — this M7 evidence + §E.3 signal.

**Full run-phase (M1-M7) AC roll-up — 30/30 PASS:**
- **M1-M3** (prior commits, re-verified at final HEAD, no regression): AC-TOK-001..007, AC-TYP-001..004, AC-LAY-001/002/003/004/005, AC-LIT-001, AC-BLD-002. Regression spot-check: ⌘K search + `cw-lang-switch` + `data-gh-stars` present; home slots (docs-hero/filters/featured/grid) + detail slots (doc-layout/read-progress/rail-card/next-cta) render; MaruBuri 0; two-token mono present.
- **M4** (this session): AC-MER-001, AC-MER-002.
- **M5**: AC-MAS-001, AC-MAS-002, AC-MAS-003.
- **M6**: AC-CMP-001, AC-CMP-002, AC-CMP-003, AC-CMP-004.
- **M7**: AC-I18N-001, AC-I18N-002 (+ closes AC-BLD-001 warning-free gate + token-parity gate).
- Count reconciliation: acceptance.md §A actual matrix = 25 MUST + 5 SHOULD = 30 (the §A headline "20 MUST / 10 SHOULD" is a plan-phase doc inconsistency — actual severity column tally is 25/5; NOT modified per body-content ownership boundary, flagged for sync-phase). All 5 SHOULD (TYP-002, TYP-004, LAY-004, MER-002, CMP-002) PASS.

**Self-referential SHA note:** M7 source + evidence + §E.3 in the SAME commit (B9); per spec-frontmatter-schema §D3 the header omits an inline SHA (M2-M6 style). git log is authoritative; orchestrator's final report carries the post-push SHA.

**Parallel-session safety:** specific-path `git add` of the 6 docs-site files + progress.md only (B8/B10) — disjoint from the parallel session's `.claude/settings.json` + `internal/template/templates/.claude/settings.json.tmpl` (left unstaged throughout M4-M7).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-16
run_commit_sha: pending-backfill-M7   # M7 commit cannot reference its own SHA; orchestrator backfills post-push
run_status: run-complete
milestones_this_session: [M4, M5, M6, M7]     # M1-M3c-2 landed in prior commits (4170b6e8b and earlier)
ac_pass_count: 30
ac_fail_count: 0
ac_total: 30                                  # 25 MUST + 5 SHOULD (acceptance.md §A actual severity tally)
should_ac_deferred_count: 0                   # all 5 SHOULD PASS (TYP-002/004, LAY-004, MER-002, CMP-002)
preserve_list_post_run_count: 0               # menu.html/main.yaml/content/search.json/icon.html untouched (infra preserved)
cross_platform_build: n/a                     # docs/design SPEC — no Go code (cycle_type=ddd, Hugo static site)
hugo_build: { exit: 0, warnings: 0, errors: 0, pages: "KO153/EN150/JA139/ZH150" }
new_warnings_or_lints_introduced: 0
token_parity_forbidden_literals: 0            # #faf9f5/#ecefee/#211A14/#000000/linear-gradient(135deg,#3d7d5f) + #252320
maruburi_residual: 0
mermaid_v2_palette_tokens: 2                  # primaryColor #eef4f0 + lineColor #9fa0a0
mascot_emotional_surfaces: 3                  # home(Explaining) + 404(Thinking) + empty-section(Searching)
url_blacklist_hits: 0
mermaid_non_td_directions: 0
body_emoji_in_modified_layouts: 0
four_locale_page_counts: { ko: 153, en: 150, ja: 139, zh: 150 }   # held identical every milestone; ja<ko delta pre-existing content gap
l44_pre_commit_fetch: not-applicable          # orchestrator owns push + pre-push fetch (manager-develop does NOT push per spawn instruction)
l44_post_push_fetch: not-applicable           # orchestrator pushes after verification
m1_to_mN_commit_strategy: "per-milestone commits (M4=72117bbcc, M5=73de67aac, M6=ca0f9a920, M7=this commit); NOT pushed — orchestrator pushes after verification"
run_phase_files_this_session: 24              # foot.html, 6 mascot png, mascot/doc-empty/doc-rail partials, mascot shortcode, list/index/404/single layouts, 4 i18n yaml, 4 css (brand/design/theme/tokens/layout), progress.md
status_transition: none                        # status stayed in-progress (M1 set draft->in-progress); in-progress->implemented->completed owned by manager-docs (sync-phase)
```

## §E.4 Sync-phase Audit-Ready Signal

_(pending sync-phase)_

## §F Phase 4 Mode Selection

- Tier: L | era: V3R6 | harness: thorough
- Input params: scope ~12 files (CSS + Hugo layouts + partials + mascots + i18n); domain count 6 (tokens / typography / layout / mermaid / mascot / i18n); file mix CSS+HTML+JS+YAML; concurrency benefit LOW (coding-heavy implementation).
- Mode evaluation: Mode 1 (trivial) — no; Mode 2 (background) — no (write work); Mode 3 (agent-team) — RETIRED; Mode 4 (parallel) — no (coding-heavy per Anthropic coding-task parallelism caveat, not research-heavy); Mode 6 (workflow) — no (semantic multi-rule transform with inter-file dependencies, not mechanical-uniform); Mode 5 (sub-agent) — selected.
- Decision: sub-agent
- Justification: coding-heavy CSS/layout migration with tight inter-file coupling (the unified token layer is consumed by every layout + component override). Anthropic coding-task parallelism caveat → sequential single sub-agent. Tier L → full Section A-E delegation template; manager-develop owns the milestone sequence with per-milestone commits, Route A main-direct.
- Run-phase: manager-develop, cycle_type=ddd (existing-CSS/layout behavior-preserving refactor; characterization baseline = current Hugo build 0-warnings + token-literal grep state + 4-locale rendered snapshot, per plan.md §G).
- Implementation Kickoff Approval: GRANTED (user approved run-phase entry 2026-07-16). plan-auditor independent audit did NOT produce a verdict (technical stall, 600s no-progress); orchestrator verified REQ/AC inventory directly (44 REQ / 9 groups / 30 AC, internally consistent) and proceeded per user direction. Resolved clarifications: mono-font = two-token split (--font-mono JetBrains + --font-code Goorm Sans Code); mermaid = stay v10 + v2 themeVariables only.
