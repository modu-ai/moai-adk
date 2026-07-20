# plan.md — SPEC-DESIGN-DOCSV2-001

> Implementation plan for the docs-site v2 migration. 7 milestones (M1–M7), ordered by decision-reversibility (highest-change-likelihood first). Tier L → harness `thorough`. **Run-phase agent: do NOT implement — this is plan-phase only.**

## §A. Context

Epic DESIGN-V2 / SPEC-1 (docs-site). Sibling SPEC-2 (moai web console) authored later. The docs-site currently runs a 3-layer token stack (warm-cream FROZEN + Clay/Cream/Ink + remap) over a geekdoc visual shell. User decisions fixed: v2 정통 정합 / round3 충실 재현 / mascot 정서 표면 한정 / Epic+2 SPECs. The v2 design bundle is at `.moai/state/ai-design-system/project/`.

## §B. Known Issues (carry into run-phase)

- `assets/css/moai-brand.scss` is STALE and uncompiled — delete as cleanup in M1.
- `moai-docs-theme.js` needs an audit — if it only drove the theme remap, remove; if it carries unrelated logic, preserve with remap code stripped (M1).
- The `mascot.html` shortcode exists but is UNUSED — promoting it to the primary vehicle (M5) is a net-new wiring, not a drop-in.
- round3 `styles.css` uses `--bg #f3f3f3` / `--ink #09110f`; canonical v2 `colors_and_type.css` uses `#f4f4f4` / `#060606` — v2 wins (research.md §A.1).

## §C. Pre-flight (before M1)

- [ ] Confirm the 2 [NEEDS CLARIFICATION] markers in §H are resolved (mono-font, mermaid version) — defaults stand absent override.
- [ ] Snapshot current docs-site build: `cd docs-site && hugo --minify --gc` warning count baseline = 0 (preserve).
- [ ] Snapshot current docs-site rendered HTML across 4 locales for post-migration structural diff.
- [ ] Copy 6 mascot PNGs from bundle to a staging location (M5 installs them).

## §D. Constraints (bind the plan)

Per spec.md §C. The load-bearing constraints for run-phase:
- C-1: `moai-brand.css` unfrozen (spec.md §G records the authorization).
- C-3: 4-locale same-PR obligation.
- C-5: LIGHT-ONLY maintained.
- C-8: Do NOT touch `.moai/config/sections/llm.yaml` and `README.ko.md` (unrelated WIP).
- C-9: Preserve i18n / menu / search infrastructure.

---

## §E. Milestones (M1–M7)

> Ordered by decision-reversibility. M1 (token architecture) and M3 (layout shell) are the highest-change-likelihood decisions; M6 (mascot copy) and M7 (mermaid palette tweak) are the most mechanical.

### M1 — Token Unification + `moai-brand.css` Unfreeze (HIGHEST architecture risk)

**Decision**: collapse the 3-layer stack into a single v2-native `moai-brand.css` + rewritten `moai-design.css` (design.md §A).

**Files**:
- `docs-site/static/moai-brand.css` — UNFREEZE + rewrite `:root` to v2 tokens (`--color-bg #f4f4f4`, `--color-ink #060606`, `--color-primary #3d7d5f` + hover/active, achromatic neutral scale, `--gradient-signature` solid `#3d7d5f`, v2 tracking/radius/shadow/motion tokens). Fold in the motion keyframes from `moai-docs-tokens.css`.
- `docs-site/static/moai-design.css` — rewrite to consume v2 tokens; preserve macOS code-card chrome + layout shell geometry.
- `docs-site/static/moai-docs-tokens.css` — DELETE (Clay/Cream/Ink + MaruBuri `@font-face` folded/removed).
- `docs-site/static/moai-docs-theme.css` — DELETE (remap layer obsolete; `md-*` components folded into `moai-design.css`).
- `docs-site/static/moai-docs-theme.js` — audit; remove if remap-only, preserve stripped otherwise.
- `docs-site/layouts/partials/head/custom.html` — drop the 2 deleted-CSS `<link>` lines + their FNV32a hash lines; keep brand + design `<link>` lines (2 hashes, not 4). Keep Pretendard + Goorm CDN links.
- `docs-site/assets/css/moai-brand.scss` — DELETE (stale, uncompiled).
- Grep-sweep `docs-site/` for `#faf9f5`, `#ecefee`, `#211A14`, `#000000`, `linear-gradient(135deg, #3d7d5f` literals and replace with v2 tokens.

**4-locale impact**: none (CSS is shared).

**Verification**: `hugo --minify --gc` warning-free; grep for forbidden literals returns 0; rendered home page has `--color-bg: #f4f4f4` in computed `:root`.

**Rollback**: `git revert` the M1 commit; the FROZEN `moai-brand.css` is restored from `HEAD~1`.

### M2 — Typography (MaruBuri drop + mono-font decision)

**Decision**: Pretendard-only for title/body; mono font per design.md §D (default = two-token split).

**Files**:
- `docs-site/static/moai-brand.css` — confirm MaruBuri `@font-face` is gone (folded in M1; verify); define `--font-mono` (JetBrains) + `--font-code` (Goorm-aware) tokens; add `@import url('https://fonts.googleapis.com/css2?...JetBrains+Mono...')` if not already present.
- `docs-site/layouts/partials/head/custom.html` — confirm Pretendard CDN link; keep Goorm CDN link (used by `--font-code`); ensure JetBrains Mono CDN is wired.
- Grep-sweep for `MaruBuri` / `maruburi` references across `docs-site/` (CSS, layouts, content) — remove or replace with Pretendard.

**[NEEDS CLARIFICATION: mono-font]** — resolved pre-run; default = two-token split.

**4-locale impact**: none (typography is shared).

**Verification**: grep `@font-face` in `moai-brand.css` returns only Pretendard self-host entries (if any) + zero MaruBuri; rendered page `font-family` computed value resolves to Pretendard for `--font-sans`.

### M3 — Layout Recreation: docs-index + docs-detail (HIGHEST overall risk)

**Decision**: faithfully recreate round3 `05-docs-index` + `06-docs-detail` in the Hugo/geekdoc layout system, preserving i18n/menu/search infra (design.md §B).

**Files**:
- `docs-site/layouts/_default/baseof.html` — rewrite chrome to round3 `.nav` + `.page` + `.footer` geometry; preserve MutationObserver (REQ-LIT-001); preserve menu partial hook + main block hook; re-style TOC aside to round3 `.toc` (left rail).
- `docs-site/layouts/index.html` — rewrite to round3 docs-hero + docs-stats + featured-card + docs-grid. Preserve i18n eyebrow/title/lead via `hugo.Sites`; preserve `.md-stat-row` (or move to params).
- `docs-site/layouts/_default/single.html` (or the section-render equivalent) — rewrite to round3 doc-hero + 3-col doc-layout (TOC left + body center + rail right) + read-progress bar + next-CTA. Adopt DocPage article-shell styling for `.doc-body`.
- `docs-site/layouts/_default/list.html` — rewrite to round3 docs-index variant for section landing (hero + filters + grid).
- `docs-site/layouts/partials/site-header.html` — rewrite visual to round3 `.nav`; PRESERVE ⌘K search + lang-switch + GitHub Star + version pill (infra).
- `docs-site/layouts/partials/site-footer.html` — rewrite visual; preserve copyleft + attributions.
- `docs-site/layouts/partials/doc-rail.html` — NEW. Right-rail partial (actions + related + CTA) per round3 `.rail`.
- `docs-site/layouts/partials/doc-empty.html` — NEW. Empty-state partial (mascot + wit copy + CTA) for sections with no pages.
- `docs-site/layouts/_default/_markup/render-codeblock.html` — restyle to v2 code-card.
- `docs-site/static/js/doc-progress.js` (or inline) — NEW. Reading-progress scroll handler + IntersectionObserver for TOC active state.
- `docs-site/layouts/404.html` — rewrite to round3 geometry + Thinking mascot placement.

**PRESERVED (infra — do NOT break)**:
- `docs-site/layouts/partials/menu.html` + `menu/name.html` + `menu/href.html` — restyle only; icon→SVG case ladder MUST stay in sync with `data/menu/main.yaml` `icon:` values.
- `docs-site/data/menu/main.yaml` — untouched (4-locale name maps).
- `docs-site/content/<locale>/` + `_meta.yaml` — untouched.
- `docs-site/layouts/partials/search.json` + ⌘K modal — untouched.
- `docs-site/layouts/shortcodes/icon.html` — untouched.

**4-locale impact**: structural copy (hero deck, empty-state wit, rail CTAs) must come from `i18n/<locale>.yaml` — ko authored first, en/ja/zh translated same-PR.

**Verification**: rendered `/ko/`, `/en/`, `/ja/`, `/zh/` home + a section list + a detail page each match the round3 slot structure (DOM class assertions in acceptance.md); `hugo --minify --gc` warning-free; ⌘K search still works; lang-switch still works.

**Rollback**: `git revert` the M3 commit(s). Because M3 touches the layout shell, a partial revert is risky — revert the whole M3 commit range. This is why M3 is a single logical commit (or a small tightly-coupled range).

### M4 — Mermaid v2 Palette (+ version decision)

**Decision**: apply v2 themeVariables to `foot.html`; stay on mermaid@10 (design.md §E default).

**Files**:
- `docs-site/layouts/partials/foot.html` — replace the `lightTheme` object's values with v2: `primaryColor #eef4f0`, `primaryTextColor #060606`, `primaryBorderColor #3d7d5f`, `lineColor #9fa0a0` (GRAY, was green), `actorBkg #eef4f0`, `actorBorder #3d7d5f`, `noteBkgColor #e6e6e6`, `noteBorderColor #b5b5b5`, `clusterBkg #f4f4f4`, `clusterBorder #d1d1d1`, `titleColor #060606`, `edgeLabelBackground #ffffff`. Drop the warm-cream `#d6ebde` / `#faf9f5` values.

**[NEEDS CLARIFICATION: mermaid version]** — resolved pre-run; default = stay v10.

**4-locale impact**: none (mermaid theme is global).

**Verification**: render a page with a known mermaid diagram (e.g. a TD flowchart) across 4 locales; confirm lineColor renders gray `#9fa0a0`, primaryColor renders `#eef4f0`. Grep `foot.html` for `#d6ebde` / `#faf9f5` returns 0.

### M5 — Mascot Expansion (6 poses, emotional surfaces)

**Decision**: copy 6 mascot PNGs; wire via `mascot.html` shortcode + new `partials/mascot.html` (design.md §F).

**Files**:
- `docs-site/static/mascots/` — COPY 6 PNGs from bundle: `MoAI-Mascot-Thinking.png`, `MoAI-Mascot-Pointing.png`, `MoAI-Mascot-Searching.png`, `MoAI-Mascot-Teaching.png`, `MoAI-Mascot-Explaining.png`, `MoAI-Mascot-Coffee.png`. (Existing `mascot-coding.png` etc. stay — they are in-use on header/home.)
- `docs-site/layouts/shortcodes/mascot.html` — EXTEND. Accept `pose` param (`thinking|pointing|searching|teaching|explaining|coffee`); emit `<img class="mascot mascot-<pose>" src="/mascots/MoAI-Mascot-<Pose>.png" alt="" loading="lazy" />`.
- `docs-site/layouts/partials/mascot.html` — NEW. Same contract as the shortcode, callable from layouts.
- `docs-site/layouts/index.html` — place Explaining mascot on home hero (or Pointing as alt).
- `docs-site/layouts/404.html` — place Thinking mascot.
- `docs-site/layouts/partials/doc-empty.html` — place Searching mascot on empty section state.
- `docs-site/layouts/_default/_markup/render-codeblock.html` — place Coffee mascot on copy-success state (ephemeral).

**Forbidden placements** (REQ-MAS-003): data tables, forms, checkout. Docs-site has none; the rule binds any future addition.

**4-locale impact**: none (mascots are visual; pose→surface map is locale-invariant).

**Verification**: 6 PNGs exist in `static/mascots/`; home / 404 / empty-section each render exactly 1 mascot `<img>` with the canonical filename.

### M6 — Component Adoption (v2 recipes mapped to cw-*/gdoc-*)

**Decision**: apply v2 button / card / chip / callout / shadow / motion recipes to the existing component overrides (design.md §C).

**Files**:
- `docs-site/static/moai-brand.css` — rewrite the `cw-*` / `gdoc-*` component blocks to v2 recipes. Primary CTA pill, surface card, chip, callout, shadow system, motion defaults.
- `docs-site/static/moai-design.css` — rewrite `.md-btn`, `.md-doc-card`, `.md-stat-row`, `.code-card`, `.code-chrome` to v2.

**Forbidden-pattern sweep**: grep for `border-left:.*#3d7d5f` adjacent to `border-radius` (the AI-slop card pattern); remove if found.

**4-locale impact**: none (components are shared CSS).

**Verification**: rendered buttons / cards / callouts match v2 recipes (computed style assertions in acceptance.md).

### M7 — 4-Locale Parity Sweep + Build Gate

**Decision**: verify 4-locale parity across every visual change; lock the warning-free Hugo build.

**Files**:
- `docs-site/i18n/{ko,en,ja,zh-cn}.yaml` — add any new structural copy keys introduced in M3 (hero deck, empty-state wit, rail CTAs). ko canonical, en/ja/zh translated same-PR.
- `docs-site/content/<locale>/` — ONLY if a structural slot requires a new page (out of scope per spec.md §E; default: no content edits).

**Verification** (the M7 gate, run before sync-phase):
- `cd docs-site && hugo --minify --gc` → 0 warnings, 0 errors.
- `diff -r public/ko/ public/en/ public/ja/ public/zh/` structural parity (modulo translated strings).
- 4-locale file-existence parity: every page in ko has en/ja/zh counterparts.
- URL-blacklist grep: `docs.moai-ai.dev` / `adk.moai.com` / `adk.moai.kr` → 0 occurrences.
- Mermaid LR/RL direction grep → 0 occurrences (TD-only).
- Body-emoji scan across new/modified layouts → 0 occurrences.
- Token-parity grep: `#faf9f5` / `#ecefee` / `#211A14` / `linear-gradient(135deg, #3d7d5f` → 0 occurrences across `docs-site/`.
- Mascot placement count: home + 404 + empty-section each have 1 mascot; forbidden surfaces have 0.

**4-locale impact**: this IS the 4-locale gate.

---

## §F. Anti-Patterns (run-phase MUST avoid)

- **AP-1**: editing `data/menu/main.yaml` icon values without adding the matching SVG case in `menu.html` lines 23–41 → empty SVG renders. Always edit both atomically.
- **AP-2**: introducing hardcoded Korean in a shared layout template → 4-locale parity FAIL. Always route through `i18n/<locale>.yaml`.
- **AP-3**: leaving a `#faf9f5` / `linear-gradient(135deg,` literal in a component override buried in `moai-brand.css` → token-parity AC FAIL. Grep-sweep is mandatory.
- **AP-4**: re-enabling dark mode "because v2 has dark tokens" → REQ-LIT-001/002 violation. Dark block stays dead.
- **AP-5**: using body emoji in a new component → REQ-CMP-006 violation. Use `icon.html` shortcode.
- **AP-6**: touching `.moai/config/sections/llm.yaml` or `README.ko.md` (the pre-existing uncommitted WIP) → C-8 violation. Leave as-is.
- **AP-7**: replacing the geekdoc shell without preserving the `gdoc-nav→menu` / `gdoc-page→main` block hooks → Hugo render pipeline breaks. Preserve hooks.
- **AP-8**: applying gradient + shadow simultaneously on a single element → REQ-TOK-006 violation (FROZEN rule).

---

## §G. Self-Verification (run-phase §E.1)

The run-phase agent (manager-develop) populates `progress.md` §E.2/§E.3 with verbatim command output. At plan-phase, the expectations are fixed here:

- **Token parity**: `grep -rn '#faf9f5\|#ecefee\|#211A14\|#000000\|linear-gradient(135deg, #3d7d5f' docs-site/static/ docs-site/layouts/ | grep -v '_test' | wc -l` → 0.
- **MaruBuri purge**: `grep -rin 'maruburi' docs-site/ | wc -l` → 0.
- **Build**: `cd docs-site && hugo --minify --gc 2>&1 | grep -ic 'warn\|error'` → 0.
- **4-locale parity**: section count across `content/ko/`, `content/en/`, `content/ja/`, `content/zh/` is equal (or the documented ko-only delta).
- **Mascot presence**: rendered home / 404 / empty-section each have exactly 1 `img.mascot`.
- **Mermaid palette**: `grep -n 'primaryColor.*#eef4f0\|lineColor.*#9fa0a0' docs-site/layouts/partials/foot.html` → 2 hits (both tokens present).

---

## §H. [NEEDS CLARIFICATION] Markers (resolve before Implementation Kickoff Approval)

- **[NEEDS CLARIFICATION: mono-font]** (M2) — default = two-token split (`--font-mono` JetBrains + `--font-code` Goorm). Override = single JetBrains (v2 purist). design.md §D.
- **[NEEDS CLARIFICATION: mermaid version]** (M4) — default = stay v10. Override = bump v11 + load-test `mermaid.run`. design.md §E.

Both have documented defaults and override paths; neither blocks plan-audit. The orchestrator resolves them at the Implementation Kickoff Approval gate (plan→run HUMAN GATE) via AskUserQuestion if the user wants to override; otherwise the defaults stand.

---

## §I. Rollback Path

- **M1/M2/M4/M5/M6 rollback**: `git revert <commit>` — each milestone is a self-contained logical commit. The FROZEN `moai-brand.css` is recoverable from `HEAD~<N>`.
- **M3 rollback** (layout shell — highest risk): M3 is committed as a single logical commit (or a tightly-coupled range). Revert the whole range. The geekdoc shell returns to its pre-M3 state. A partial M3 revert is NOT supported (the layout files are tightly coupled).
- **Full rollback**: revert the entire SPEC commit range. The docs-site returns to its pre-SPEC 3-layer state. The FROZEN warm-cream `moai-brand.css` is restored.

Because the SPEC uses Hybrid Trunk main-direct push (C-2), each milestone push to `main` is a deploy to adk.mo.ai.kr via Vercel. The rollback path is `git revert` + push, which Vercel auto-deploys.

---

## §J. Cross-References

- spec.md: `.moai/specs/SPEC-DESIGN-DOCSV2-001/spec.md` — GEARS requirements, constraints, Out of Scope, unfreeze authorization.
- acceptance.md: `.moai/specs/SPEC-DESIGN-DOCSV2-001/acceptance.md` — 30 ACs.
- research.md: `.moai/specs/SPEC-DESIGN-DOCSV2-001/research.md` — token DELTA table, round3 breakdown, 3-layer divergence, mono-font + mermaid version analysis.
- design.md: `.moai/specs/SPEC-DESIGN-DOCSV2-001/design.md` — token-unification architecture, layout architecture, component mapping, mascot placement, 4-locale parity, light-only preservation.
- Epic: DESIGN-V2. Sibling SPEC-2 (moai web console) TBD.
