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

## §E.3 Run-phase Audit-Ready Signal

_(pending run-phase)_

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
