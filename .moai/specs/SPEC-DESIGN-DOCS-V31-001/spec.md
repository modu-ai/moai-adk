---
id: SPEC-DESIGN-DOCS-V31-001
title: "docs-site v3.1-rc.1 renewal — IA redesign, design-system overhaul, Korean book-depth content, sequential 4-locale i18n"
version: 0.2.0
status: completed
created: 2026-08-11
updated: 2026-08-13
author: manager-spec
priority: High
phase: "v3.1-rc.1 target"
module: docs-site
lifecycle: spec-anchored
tags: "docs-site, design-system, i18n, korean, content, v3.1, hugo, geekdoc"
tier: L
related_specs: "SPEC-DESIGN-DOCSV2-001, SPEC-FACTORY-MODE-001, SPEC-INFINITE-GOAL-001, SPEC-PROJECT-NAVIGATOR-001, SPEC-HIERARCHICAL-TEAM-001, SPEC-AUDIT-MULTI-MODEL-001, SPEC-AUTONOMY-TIERS-001, SPEC-MODEL-PROFILE-MATRIX-001"
---

# SPEC-DESIGN-DOCS-V31-001 — docs-site v3.1-rc.1 renewal

## HISTORY

- 2026-08-11 — v0.1.0 draft authored (manager-spec). Four-axis scope: (1) IA redefinition with v3.1 NEW badges, (2) Claude Design handoff application, (3) Korean introductory-book-depth content rewrite, (4) ko-first sequential 4-locale translation. User decisions (pipeline, IA, content depth, i18n order, light-only, full design scope) confirmed upstream via AskUserQuestion and are treated as non-re-openable inputs.
- 2026-08-11 — v0.2.0 iteration-2 revision (manager-spec, re-delegation per plan-auditor FAIL 0.76). Defect delta: D1 (3 NEEDS-CLARIFICATION markers → RESOLVED per user decisions), D2 (29 REQ → 25 REQ, 32 AC → 25 AC via consolidation), D3 (book-level-prose stripper corrected to match its rubric description — verified on two ko samples), D4 (REQ-KO-006 friend-explainability tightened to a mechanical Korean-causal-connector + concrete-capability-noun predicate), D5 (mascot filename typo `moai-logo-4-W` → `moai-logo-4-WH`), D6 (handoff README community-platform provenance noted in research.md), D7 (`(Event-detected)` → `(Event-driven)` labels), D8 (Epic-split escape valve trigger quantified — N=40 turns / T=200 cumulative / 2 same-root sub-milestone failures), D9 (M5 gate promoted to MUST; M6-gating sentence added), D10 (3× i18n wall-clock multiplier quantified).

---

## §A. Context / Problem

The public documentation site at `adk.mo.ai.kr` is pinned at `v3.0.2` (hugo.toml `[params].version`) and carries a 4-locale × 124-page corpus (496 pages). Since the v3.0 freeze (SPEC-DESIGN-DOCSV2-001, 2026-07-16), eleven user-facing capabilities have landed or stabilized without being reflected in the public docs: `/moai goal` with infinite-duration support (SPEC-INFINITE-GOAL-001), Factory Mode (SPEC-FACTORY-MODE-001), the BAS Navigator 3-tier codemap sync (SPEC-PROJECT-NAVIGATOR-001/002/003), `manager-lead` hierarchical teams (SPEC-HIERARCHICAL-TEAM-001), multi-model audit (SPEC-AUDIT-MULTI-MODEL-001), `MOAI_AUTONOMY_TIER` mode tiers (SPEC-AUTONOMY-TIERS-001), the `profile` matrix (SPEC-MODEL-PROFILE-MATRIX-001), and per-agent model enforcement. The current IA (13 sections, `_meta.yaml` weight-ordered) has no mechanism for surfacing new capabilities to a returning reader, and the current Korean content depth is reference-grade (terse, table-heavy, callout-saturated) rather than the learning-grade prose a first-time adopter needs.

A Claude Design handoff package is staged at `/tmp/moai-design-handoff/moai-adk/project/` containing six screen prototypes (`01 Docs Home` …`06 Not Found`), a `_ds/` design-system directory with a v2-renewal `colors_and_type.css` token SSOT and a 14KB README codifying tone/copy/color/type/animation, six mascot pose PNGs, and three logo variants. The handoff's token palette is achromatic-neutral-plus-mascot-green (`#3d7d5f`) — the same accent the frozen `moai-brand.css` already carries — but the neutral ramp, type scale, motion system, shadow ladder, and component vocabulary are a genuine delta over the live site. The handoff prototypes are `.dc.html` (a viewer format); the production target is Hugo + hugo-geekdoc, NOT the prototype viewer.

This SPEC governs the simultaneous renewal of all four surfaces: information architecture, visual design system, Korean content quality, and locale derivation order.

---

## §B. Goals & Non-Goals

### §B.1 Goals

1. **Information Architecture renewal.** Define a NEW integrated 12-section section-list TOC (per the Docs Home handoff grid — user decision: the live 13th section `changelog` moves to the site footer, NOT a sidebar section; `cost-optimization` retains its own card per the handoff home grid) that absorbs every v3.1 capability into its natural home and surfaces each as a first-class navigational citizen.
2. **v3.1 discoverability.** Every page that documents a v3.1-originated feature carries a visible "NEW" badge in the sidebar and in the page header, driven by a single declarative source so it can be revoked in v3.2 without a sweep.
3. **Design system application.** Port the handoff's design system to the live Hugo site — tokens, type scale, 6 mascot illustrations, home hero, screen layouts, and the component vocabulary — as a light-only theme (the handoff's dark-mode tokens are intentionally ignored per CLAUDE.local.md §17.1).
4. **Korean content depth lift.** Rewrite every Korean page from reference-grade to introductory-book-grade: friendly prose, concept-first exposition, one or more infographics per concept, runnable code, step-by-step structure, and a volume floor per page class.
5. **Sequential 4-locale translation.** Korean completed and verified FIRST; `en`, `ja`, `zh` derived sequentially (not parallel) from the verified Korean SSOT, each locale fully verified before the next begins.
6. **Self-contained build.** The Hugo site MUST remain free of the handoff's runtime viewer bundle (`_ds_bundle.js`, 132KB); only the extracted token CSS and static mascot/logo PNG assets ship.

### §B.2 Non-Goals (Out of Scope)

### Out of Scope — dark theme

- The handoff defines `[data-theme="dark"]` tokens in `colors_and_type.css`. Implementing a dark mode for docs-site is OUT OF SCOPE. Per CLAUDE.local.md §17.1, docs-site is light-only by HARD rule; the handoff's dark tokens are documented in `research.md` for traceability but MUST NOT be shipped. Any pre-existing `[data-theme="dark"]` selectors in the CSS layer are treated as dead code and are not augmented.

### Out of Scope — handoff runtime JS

- `_ds_bundle.js` (132KB), `_ds_manifest.json`, `support.js`, and the `.dc.html` viewer's `<script type="text/x-dc">` component logic are NOT shipped. They are prototype-viewer runtime. The Hugo site consumes ONLY `colors_and_type.css` (as token source) and the static image assets. No external-runtime JS dependency may be introduced by this SPEC.

### Out of Scope — moai web console documentation

- The `moai web` console redesign (PR #1410, b0d3b61f8) is an internal maintainer surface and is NOT documented in the public docs-site. It is explicitly excluded from the v3.1 NEW-feature catalog.

### Out of Scope — content generation tooling

- This SPEC does not mandate an authoring tool (template-driven, AI-assisted, or manual). The content rewrite is governed purely by the acceptance rubric in `acceptance.md` §A; the tool choice is a run-phase decision owned by `manager-develop`.

### Out of Scope — URL restructure

- `baseURL` remains `https://adk.mo.ai.kr/`. The 4-locale URL prefix scheme (`/ko/`, `/en/`, `/ja/`, `/zh/`) is unchanged. Section slug changes are permitted only where a section is renamed or merged (tracked in `plan.md` §F M1), and every renamed slug MUST be paired with a `vercel.json` redirect entry per CLAUDE.local.md §17.

### Out of Scope — analytics, search backend, comments

- No new analytics, no search-backend swap (geekdoc built-in search stays), and no comments system. The handoff's search screen prototype (`05 Search.dc.html`) informs the visual styling of the existing search affordance only.

---

## §C. Requirements (GEARS notation)

### §C.1 Information Architecture

**REQ-IA-001** (Ubiquitous) — The docs-site information architecture SHALL be expressed as a single integrated section list, derived from the handoff `01 Docs Home` 12-card grid, reconciled against the live 13-section tree, and published as the canonical `content/<locale>/_meta.yaml` weight-ordered manifest for every locale.

**REQ-IA-002** (Ubiquitous) — Every v3.1-originated user-facing capability listed in §F.1 (v3.1 feature catalog) SHALL have a documented home in exactly one section of the new IA; no v3.1 capability MAY be left undocumented, and no capability MAY be documented in more than one section (cross-references via inline links, not duplicate pages).

**REQ-IA-003** (Event-driven) — **When** the IA introduces a section rename or slug change, the docs-site SHALL emit a matching `vercel.json` redirect entry from the old slug to the new slug for every locale, so that any pre-existing external link resolves without a 404.

**REQ-IA-004** (State-driven) — **While** the site is pinned at a `v3.1-rc.*` or `v3.1.*` release, the version pill in the site header (currently `v3.0.0` in the handoff prototypes, `v3.0.2` in the live hugo.toml) SHALL reflect the exact release identifier from `hugo.toml [params].version`, rendered in the monospace pill style defined by the handoff header.

### §C.2 NEW-badge mechanism

**REQ-NB-001** (Ubiquitous) — The "NEW" badge indicator SHALL be driven by a dual source: (a) a Hugo frontmatter flag (`new: true` OR `added_in: "v3.1"` — the version-string form preferred, it enables a future mechanical sunset sweep) on the page file, OR (b) an entry in a section-level `new_items:` list in `_meta.yaml`; both mechanisms are supported and union-combined (either source triggers the badge). The docs-site SHALL additionally ship a `new-badge` Hugo shortcode at `layouts/shortcodes/new-badge.html` that renders the visual badge inline (for use in page bodies, section indexes, callouts) with the token-driven styling (green pill, `--color-primary` background, white text, monospace caption "NEW"). *(Iteration-2 consolidation: merges v0.1.0 REQ-NB-001 mechanism + REQ-NB-002 shortcode into a single mechanism REQ — rationale: the shortcode and the flag are the same indicator surface, tested together.)*

**REQ-NB-002** (State-driven) — **While** a page's frontmatter carries the NEW flag (or a section's `_meta.yaml` carries `new_items:` / `new: true`), the sidebar menu partial (`layouts/partials/menu.html`) SHALL render the badge beside the page title (and, **Where** a section index page is itself v3.1-new, beside the section heading via the section-level flag), and `layouts/_default/single.html` SHALL render it beside the `<h1>`. The badge MUST disappear from both surfaces automatically when the flag is removed or after the documented sunset cycle (default: next minor release). *(Iteration-2 consolidation: merges v0.1.0 REQ-NB-003 sidebar+header rendering + REQ-NB-004 section-level propagation into a single rendering REQ — rationale: both bind the flag to a render surface, tested by one AC.)*

### §C.3 Design system application

**REQ-DS-001** (Ubiquitous) — The docs-site SHALL adopt the handoff's v2-renewal design tokens as the authoritative token layer, ported into `docs-site/static/moai-brand.css` (replacing the current `:root` block) and `docs-site/static/moai-docs-tokens.css`. The token vocabulary — `--color-primary #3d7d5f`, `--color-ink #060606`, `--color-bg #f4f4f4`, the mascot-derived neutral ramp (`--neutral-100 #e6e6e6`, `--neutral-400 #9fa0a0`, `--neutral-900 #141414`), the type scale (`--text-display` clamp, 9-weight Pretendard), the 7-step radius ladder, the 6-step shadow ladder including `--shadow-signature`, and the motion easings (`--easing-default`, `--easing-bounce`, `--easing-smooth`) — SHALL be emitted verbatim from `colors_and_type.css`.

**REQ-DS-002** (State-driven) — **While** `prefers-reduced-motion: reduce` is active at the client, every transition and animation on the docs-site SHALL degrade to 1ms duration, matching the handoff's reduced-motion block; mascot bounce easings (`--easing-bounce`) are NOT exempted.

**REQ-DS-003** (Ubiquitous) — The docs-site SHALL NOT ship a dark theme. The `[data-theme="dark"]` selector block from `colors_and_type.css` MUST NOT be ported into the production CSS; any pre-existing dark-mode CSS in the docs-site tree is left untouched (frozen as dead code) but MUST NOT be augmented, extended, or newly wired by this SPEC.

**REQ-DS-004** (Ubiquitous) — The six mascot pose PNGs (`MoAI-Mascot-Explaining`, `-Coffee`, `-Pointing`, `-Searching`, `-Teaching`, `-Thinking`) and the three logo variants (`moai-logo-1`, `moai-logo-4`, `moai-logo-4-WH`) SHALL be copied into `docs-site/static/mascots/` and `docs-site/static/logos/` respectively; the existing `mascot` shortcode (`layouts/shortcodes/mascot.html`) is extended so that its `$poses` slice continues to accept the six canonical pose names. *(v0.2.0 D5 fix: `moai-logo-4-W` was a typo; the staged handoff file is `moai-logo-4-WH.png`, verified by `ls /tmp/moai-design-handoff/moai-adk/project/assets/`.)*

**REQ-DS-005** (Ubiquitous) — The docs-site home page (`layouts/index.html` + `content/<locale>/_index.md`) SHALL be restructured to match the Docs Home handoff (`01 Docs Home.dc.html`) layout: hero block with mascot + value-proposition headline + dual CTA + install-command card, the 3-card value grid (Tokenomics / Self-Learning / Harness), the section grid (the NEW integrated TOC), and the book CTA card. The current home page's flat section list is retired.

**REQ-DS-006** (Ubiquitous) — The docs-site header partial (`layouts/partials/site-header.html`) SHALL carry the sticky transparent header with `backdrop-filter: blur(12px)` (or a fallback solid color on non-supporting browsers), the search affordance with `⌘K` affordance hint, the version pill, the language switcher, and the GitHub link — matching the handoff header across all six screen prototypes (they share the header).

**REQ-DS-007** (Ubiquitous) — Every inline icon used in docs content SHALL continue to use the existing `{{< icon <name> [variant] >}}` shortcode (CLAUDE.local.md §17.1); the handoff's Lucide icon vocabulary is absorbed by extending the `layouts/partials/menu.html` SVG switch and the `icon.html` shortcode's supported name set, NOT by introducing a new icon shortcode.

### §C.4 Korean content rewrite (introductory-book depth)

**REQ-KO-001** (Ubiquitous) — Every Korean page (`content/ko/**/*.md`) SHALL be rewritten to introductory-book-grade prose as defined by the rubric in `acceptance.md` §A. The rubric has five pillars: (1) concept-first exposition, (2) step-by-step structure, (3) infographic density (at least one Mermaid `flowchart TD` diagram, `{{< mascot <pose> >}}` illustration with caption, or inline SVG/PNG figure with descriptive alt text, on every concept page — reference / CLI pages MAY waive this floor with an explicit `infographic_floor_waived: true` frontmatter key plus a one-line reason; *(v0.2.0 D2 consolidation: former REQ-KO-004 infographic rule folded here as pillar 3 — the floor belongs to the rubric, not a standalone REQ)*), (4) runnable examples, and (5) friend-of-a-friend explainability. The friend-explainability pillar (5) is enforced by a MECHANICAL sidecar predicate defined in `acceptance.md` §A.4 — the page's authoring-trail sidecar MUST contain (5a) at least one Korean causal connector (`왜냐하면` / `때문에` / `따라서` / `그래서` / `덕분에`) AND (5b) at least one concrete noun naming a specific moai-adk capability (e.g. `SPEC` / `TRUST 5` / `harness` / `goal` / `factory` / `에이전트` — NOT generic `AI` or `도구`). *(v0.2.0 D4 fix: former REQ-KO-006 demoted from a self-audit-only bar to this mechanical predicate — see acceptance.md §A.4 for the grep-able check.)* The prior reference-grade table-and-callout style is retired for body content (reference tables remain valid where a true enumeration is the clearest expression).

**REQ-KO-002** (State-driven) — **While** a page belongs to one of the page classes defined in `acceptance.md` §A.2 (concept page, tutorial page, reference page, onboarding page), the page SHALL meet the per-class floor metrics: minimum prose word count, minimum runnable code block count, and minimum step-by-step `## Step N` heading count. (The infographic floor and the friend-explainability predicate are owned by REQ-KO-001's pillars 3 and 5; this REQ owns the quantitative per-class floors only.)

**REQ-KO-003** (Event-driven) — **When** a Korean page introduces a technical term for the first time (e.g. "에이전트(스스로 일하는 AI)", "SPEC(Software Engineering Process Catalog)"), the page SHALL expand the term once inline using the parenthetical-gloss pattern from the handoff README §2 ("Tech jargon: 첫 등장 시 한 번 풀어쓰기"); subsequent occurrences MAY use the bare term. *(v0.2.0 D7 fix: GEARS label `(Event-detected)` → `(Event-driven)` — canonical pattern name.)*

**REQ-KO-004** (Ubiquitous) — The Korean voice SHALL follow the handoff README §2 register and vocabulary rules: `~합니다` body formality with `~해보세요 / 할까요?` CTA register permitted; the word "모두의" appears naturally at least once on the home page and the introduction page; the banned-vocabulary list (혁신적인, leverage, 솔루션, Game-changing, 절대로, 유일한, 최고의, "지금 안 하면 평생 후회", etc.) is NEVER used. *(v0.2.0 D2 renumber: former REQ-KO-005, renumbered after the REQ-KO-004→REQ-KO-001 fold and the REQ-KO-006→REQ-KO-001 fold.)*

### §C.5 Sequential 4-locale translation

**REQ-I18N-001** (Ubiquitous) — The docs-site i18n pipeline SHALL execute in the fixed sequential order: `ko` (canonical, fully complete and verified) → `en` → `ja` → `zh`. Each locale's verification gate (`hugo build` exit 0, 4-locale page-count parity, 4-locale H2-section-count parity, URL blacklist clean, TD-only Mermaid, no body emoji) MUST pass before the next locale begins.

**REQ-I18N-002** (Event-driven) — **When** a non-canonical locale translation is initiated, it SHALL derive from the verified Korean SSOT at the commit the Korean gate passed; re-translation against a later Korean revision is permitted only when the Korean revision has re-passed its gate. *(v0.2.0 D7 fix: GEARS label `(Event-detected)` → `(Event-driven)` — canonical pattern name.)*

**REQ-I18N-003** (State-driven) — **While** the Korean locale is in progress (rewrite authored but gate not yet passed), the `en`, `ja`, `zh` locales MUST remain frozen at their pre-SPEC content; parallel translation is explicitly prohibited by the user decision (encoded here as a HARD ordering constraint, not a preference).

**REQ-I18N-004** (Ubiquitous) — Every derived-locale page SHALL preserve the Korean page's Mermaid diagrams, code blocks, mascot placements, shortcode usage, and frontmatter `weight` / `new` / `added_in` flags; only natural-language prose and heading text are translated. Shortcode argument values that are semantic identifiers (icon names, mascot pose names) are NEVER translated.

**REQ-I18N-005** (Ubiquitous) — The `data/menu/main.yaml` per-locale name maps SHALL be updated for every locale so that the sidebar heading labels are localized; the icon value is identical across locales (it is a semantic identifier consumed by the menu partial's SVG switch).

### §C.6 Build & neutrality

**REQ-BL-001** (Ubiquitous) — The docs-site Hugo build (`cd docs-site && hugo --gc --minify`) SHALL exit 0 with zero warnings across all four locales at every milestone boundary.

**REQ-BL-002** (Capability gate) — **Where** any artifact touched by this SPEC is mirrored into `internal/template/templates/` (CLAUDE.local.md §2 Template-First Rule), the mirrored artifact SHALL NOT carry any internal SPEC ID, REQ/AC token, audit citation, internal date, or commit SHA (CLAUDE.local.md §25). The SPEC files under `.moai/specs/SPEC-DESIGN-DOCS-V31-001/` are repo-local and are NOT mirrored; the neutrality constraint does not restrict SPEC content.

**REQ-BL-003** (Ubiquitous) — The docs-site SHALL remain a self-contained static site: no runtime fetch of `_ds_bundle.js`, no external CDN dependency for the design-system CSS (Pretendard self-host via `docs-site/static/fonts/` is permitted; Google Fonts `Inter` + `JetBrains Mono` via the existing `layouts/partials/head/custom.html` CDN link is permitted as the established pattern), and no JS-originated runtime dependency introduced by the handoff port.

---

## §D. Constraints

- **CLAUDE.local.md §17** — docs-site 4-locale sync obligation; Mermaid TD-only; `{{< icon >}}` shortcode (no body emoji); `adk.mo.ai.kr` URL whitelist.
- **CLAUDE.local.md §17.1** — Light-only theme (HARD); design-system doc carries the v2-renewal tokens; icon-to-SVG-case coupling in `menu.html`.
- **CLAUDE.local.md §25** — Template-neutrality (applies only to mirrored artifacts, not to SPEC files).
- **CLAUDE.local.md §2** — Template-First Rule (if it ships to users, it lives in `internal/template/templates/` first; docs-site itself is NOT a distributed template — it is the live `adk.mo.ai.kr` site — so the Template-First cycle binds only to assets that are ALSO distributed, e.g. any settings/hook/CSS that a user project would receive).
- **hugo-geekdoc theme contract** — layouts override the theme via `docs-site/layouts/`; the theme's own files are NOT edited.
- **FROZEN design baseline (SPEC-DESIGN-DOCSV2-001)** — the current `moai-brand.css` is FROZEN at the v2 token vocabulary (2026-07-16). This SPEC's REQ-DS-001 performs an AUTHORIZED unfreeze: the v2-renewal token port replaces the `:root` block under a new FROZEN stamp (re-stamped at this SPEC's sync-phase close, citing SPEC-DESIGN-DOCS-V31-001).
- **Korean canonical SSOT** — Korean is the canonical locale for this SPEC's content; `en/ja/zh` derive from `ko`, never the reverse.

---

## §E. Risks

1. **M4 scale risk (Korean content rewrite, 124 pages).** This is by far the largest milestone. Mitigation: decompose by section into sub-milestones (M4.1 … M4.N) in `plan.md` §F; each sub-milestone is independently committable; the Epic-split recommendation in §G.1 provides a governance escape valve if M4 exceeds the SPEC's manageable scope.
2. **Token-port regression risk.** Replacing the `:root` token block may cascade into existing layouts that hard-coded intermediate hexes. Mitigation: `moai-docs-layout.css`, `moai-docs-theme.css`, `moai-design.css` are audited for hardcoded hexes (not `var(--…)`) before the port; all hardcoded values are converted to token references or verified-compatible.
3. **NEW-badge sunset debt.** A declarative flag without a sunset mechanism rots into permanent "NEW" badges. Mitigation: the `added_in: "v3.1"` form is preferred over `new: true` precisely because it carries the version that enables a future mechanical sweep (remove badges where `added_in < current_minor − 1`).
4. **Translation drift.** Sequential translation means `zh` is authored weeks after `ko`; if `ko` is revised in the interim, `zh` risks deriving from a stale SSOT. Mitigation: REQ-I18N-002 pins the derivation commit; any `ko` revision after the pin requires a documented re-derivation delta.
5. **Handoff prototype fidelity.** The `.dc.html` prototypes are a viewer format with inline styles; a literal port produces unmaintainable inline-styled Hugo templates. Mitigation: the prototypes are a visual target, NOT a code template; the port extracts structure (section ordering, grid shape, component placement) and restyles using the token vocabulary.
6. **Light-only divergence from handoff.** The handoff prototypes show a dark-mode toggle in the header. Removing it is a user-visible divergence. Mitigation: documented in `research.md` §A; the toggle is omitted from `site-header.html`; the `aria-label="다크 모드"` button in the prototype is NOT ported.

---

## §F. Dependencies & Cross-references

### §F.1 v3.1 feature catalog (verified — user-facing classification)

| Feature | Source SPEC | Status | User-facing? | New IA home |
|---|---|---|---|---|
| `/moai goal` (infinite-duration, REAL bounds) | SPEC-INFINITE-GOAL-001 | completed | YES | workflow-commands (new page) |
| Factory Mode (one-session plan→run→verify→sync) | SPEC-FACTORY-MODE-001 | completed | YES | advanced (new page) + workflow-commands (mention) |
| BAS Navigator 3-tier codemap sync | SPEC-PROJECT-NAVIGATOR-001/002/003 | completed | YES | advanced (new page) + utility-commands (command ref) |
| `manager-lead` hierarchical team | SPEC-HIERARCHICAL-TEAM-001 | completed | YES | advanced (new page) |
| Multi-model audit convergence | SPEC-AUDIT-MULTI-MODEL-001 | completed | YES | advanced (new page) |
| `MOAI_AUTONOMY_TIER` mode tiers | SPEC-AUTONOMY-TIERS-001 | completed | YES | advanced (new page) + cost-optimization (mention) |
| `profile` matrix (`max/medium/low`) | SPEC-MODEL-PROFILE-MATRIX-001 | completed | YES | multi-llm (rewrite) + cli-reference (command ref) |
| Per-agent model enforcement | SPEC-AGENT-MODEL-ENFORCE-001 | in-progress | YES (partial) | multi-llm (mention, no dedicated page until SPEC completes) |
| Dynamic Workflows / `ultracode` | (accumulated, CC2219-ALIGN) | completed | YES | advanced (existing ultracode-workflows page, rewrite) |
| Stop-chain / per-edit hook consolidation | SPEC-STOPCHAIN-TRIM-001 | completed | YES (behavioral) | advanced (autonomy-tier page mentions) |
| Agent body diet + parallel batching | SPEC-AGENT-PARALLEL-OPT-001 | completed | YES (indirect) | claude-code (mention) |
| CC 2.1.219 upstream alignment | SPEC-CC2219-UPSTREAM-ALIGN-001 | completed | YES | claude-code (rewrite) |
| Harness learning surface (LSEL pipeline internals excluded) | SPEC-HARNESS-LEARNING-EVO 001/002 | completed | YES (surface only) | advanced (new page: harness-learning.md) |

**Excluded from catalog** (internal, not user-facing): `moai web` console redesign (PR #1410), harness learning LSEL pipeline internals (the user-facing surface is now cataloged above; only the internal Tier-Ladder pipeline mechanics remain excluded), any in-flight SPEC not yet at `completed` by this SPEC's authoring date (except SPEC-AGENT-MODEL-ENFORCE-001, which is listed because it is user-visible and near completion).

### §F.2 Predecessor SPECs

- **SPEC-DESIGN-DOCSV2-001** — established the v2 token baseline that this SPEC unfreezes and renews. Its `moai-brand.css` FROZEN stamp is the authorization baseline; this SPEC performs an authorized re-stamp.
- **SPEC-DOCSITE-ADVANCED-001** — last docs-site content expansion (6 Advanced pages × 4-locale). Its i18n patterns (canonical KO, sequential derivation, TD-only Mermaid, icon-shortcode usage) are the established pattern and are inherited unchanged.

### §F.3 Rule references

- CLAUDE.local.md §17 (docs-site i18n rules), §17.1 (design component + icon conventions), §25 (template neutrality), §2 (Template-First).
- `.moai/docs/docs-site-i18n-rules.md` — canonical docs-site i18n SSOT.
- `hns-oss-docs-i18n-rules`, `hns-oss-docs-structure-map`, `hns-oss-docs-verify` — the oss-docs harness skills that bind at run-phase.

---

## §G. Design decisions

### §G.1 Single Tier L SPEC — user-decided; Epic-split as documented fallback

**Decision (user-confirmed at plan-phase): implement as a single Tier L SPEC, with M4 decomposed into section-scoped sub-milestones M4.1–M4.8 (one per section) per `plan.md` §F M4.** The Epic-split path into a child SPEC `SPEC-DESIGN-DOCS-V31-CONTENT-KO-001` remains a **documented fallback**, not the committed path — it fires only on a concrete trigger (see plan.md §F M4 for the N=40 / T=200 / 2-same-root-failure thresholds).

**Rationale.** Three of the four axes (IA, design system, i18n ordering) are tightly coupled — the IA defines where the NEW badges live, the design system defines how the badges render, and the i18n ordering gates when translation begins. Splitting these into separate SPECs creates cross-SPEC coordination overhead larger than the per-SPEC savings. The content rewrite (M4) is the one axis that is genuinely decomposable: a page in `cli-reference` has no authoring dependency on a page in `core-concepts`. Keeping M4 inside this SPEC preserves a single audit trail and a single "v3.1 docs ship" sync event. The per-section sub-milestone decomposition (M4.1–M4.8) carries the decomposability directly within the single-SPEC run-phase, and the Epic-split fallback exists in case that decomposition still overflows a single run-phase envelope. *(v0.2.0 D8 fix: the escape valve is now a quantified fallback, not an undefined mid-run judgment call — triggers live in plan.md §F M4.)*

### §G.2 `_ds_bundle.js` handling — extract CSS only, do NOT ship runtime JS

**Decision: the Hugo site consumes ONLY `colors_and_type.css` (as token source, ported into `moai-brand.css` and `moai-docs-tokens.css`) and the static image assets (`assets/characters/*.png`, `assets/moai-logo-*.png`). `_ds_bundle.js` (132KB), `_ds_manifest.json`, `support.js`, `_adherence.oxlintrc.json`, and the `.dc.html` viewer's `<script type="text/x-dc">` component logic are NOT shipped, NOT vendored, NOT transcribed into Hugo templates.**

**Rationale.** (1) CSP friendliness — the docs-site is served from `adk.mo.ai.kr` via Vercel and carries no external-runtime JS dependency today; introducing one would expand the CSP surface and add a runtime failure mode (a 132KB bundle loading before first paint). (2) The prototype viewer is not the production target — its runtime exists to render `.dc.html` files inside a design-tool preview, not to power a Hugo site. (3) Extracting the tokens captures the load-bearing design system; the runtime is a viewer convenience. REQ-BL-003 codifies this as a HARD constraint.

### §G.3 NEW-badge mechanism — frontmatter flag + shortcode, dual-source

**Decision: dual mechanism.**

1. **Page-level flag.** A page MAY carry `new: true` OR `added_in: "v3.1"` (preferred — the version string enables a future mechanical sunset sweep) in its YAML frontmatter. The presence of either flag marks the page as NEW.
2. **Section-level list.** A section's `_meta.yaml` MAY carry a `new_items:` list of slugs (or a section-level `new: true` for an entirely new section). This covers the case where the index page is not itself new but several child pages are, and the section should be badged in the sidebar.
3. **Shortcode.** `layouts/shortcodes/new-badge.html` renders the visual badge inline (for page bodies, section indexes, callouts). The shortcode accepts an optional version argument: `{{< new-badge v3.1 >}}`.
4. **Sidebar rendering.** `layouts/partials/menu.html` is extended: when iterating pages, it reads the page's `.Params.new` OR `.Params.added_in`, and when iterating sections, it reads the section `_meta.yaml`'s `new_items:` list; any match renders the badge beside the title. The badge vanishes automatically when the flag is removed.
5. **Page-header rendering.** `layouts/_default/single.html` renders the badge beside the `<h1>` when the same flag is present.

**Rationale for dual source.** Page-level frontmatter is the natural home for "this page is new". But sections that aggregate many new child pages (e.g. a new `advanced/factory-mode` page inside the `advanced` section) benefit from a section-level declaration so the section heading itself is badged, without requiring every child to repeat the flag. The two mechanisms are union-combined (either source triggers the badge). The preferred `added_in: "v3.1"` form carries the version, enabling the sunset sweep (a future SPEC MAY mechanically strip badges where `added_in` is older than `current_minor − 1`).

---

## §H. Cross-references

- `.moai/specs/SPEC-DESIGN-DOCSV2-001/` — predecessor design SPEC (v2 token freeze).
- `.moai/specs/SPEC-DOCSITE-ADVANCED-001/` — predecessor content-expansion SPEC (i18n pattern source).
- `.moai/docs/docs-site-i18n-rules.md` — docs-site i18n SSOT.
- `/tmp/moai-design-handoff/moai-adk/project/` — Claude Design handoff staging (NOT in repo; consumed at run-phase).
- `docs-site/hugo.toml`, `docs-site/static/moai-brand.css`, `docs-site/layouts/`, `docs-site/content/ko/_meta.yaml`, `docs-site/data/menu/main.yaml` — production targets.
