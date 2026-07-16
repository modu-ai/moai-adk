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
