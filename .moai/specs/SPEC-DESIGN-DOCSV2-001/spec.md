---
id: SPEC-DESIGN-DOCSV2-001
title: "docs-site Design System v2 Migration"
version: "0.1.0"
status: draft
created: 2026-07-16
updated: 2026-07-16
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "docs-site"
lifecycle: spec-anchored
tags: "design-system, docs-site, v2-migration, i18n, hugo, geekdoc"
era: V3R6
tier: L
---

# SPEC-DESIGN-DOCSV2-001 — docs-site Design System v2 Migration

> Epic: **DESIGN-V2** (모두의AI Design System v2 전면 적용). This is SPEC-1 of 2 (docs-site). SPEC-2 (moai web console) is authored later.

## HISTORY

| Date | Version | Author | Note |
|---|---|---|---|
| 2026-07-16 | 0.1.0 | manager-spec | Plan-phase artifact set created (spec.md / plan.md / acceptance.md / research.md / design.md). User-confirmed decisions (v2 정통 정합 / round3 충실 재현 / mascot 정서 표면 한정 / Epic+2 SPECs) fixed as inputs. |

---

## §A. User Story

**As a** docs-site reader across ko/en/ja/zh,
**I want** the docs-site to faithfully reflect the 모두의AI Design System v2 (achromatic neutral scale, single green `#3d7d5f`, solid de-emphasized signature, Pretendard-only typography, round3 docs-index/docs-detail layouts, mascot on emotional surfaces),
**so that** the documentation surface is visually consistent with the broader 모두의AI brand platform (marketing site, web console) and the docs reading experience matches the validated round3 prototypes.

**Authoritative design source**: `.moai/state/ai-design-system/project/` Claude Design handoff bundle — `README.md` (SSOT), `colors_and_type.css` (canonical token set), `SKILL.md` (FROZEN rules), `round3/05-docs-index.html` + `round3/06-docs-detail.html` + `round3/styles.css` (layout prototypes), `assets/components/DocPage.jsx` + `MermaidDiagram.jsx` (component contracts), `assets/characters/MoAI-Mascot-*.png` (6 mascot poses).

---

## §B. Functional Requirements (GEARS)

### §B.1 Token Unification

- **REQ-TOK-001** (Ubiquitous): The docs-site CSS layer shall collapse the 3 live token vocabularies (`moai-brand.css` warm-cream FROZEN, `moai-docs-tokens.css` Clay/Cream/Ink, `moai-docs-theme.css` remap) into a single v2 token system sourced from `colors_and_type.css`.
- **REQ-TOK-002** (Ubiquitous): The docs-site `:root` shall define `--color-primary #3d7d5f`, `--color-primary-hover #316750`, and `--color-primary-active #265240` as the single accent color family.
- **REQ-TOK-003** (Ubiquitous): The docs-site `:root` shall define `--color-bg #f4f4f4` and `--color-ink #060606` as the achromatic page canvas and foreground ink (hue 0%).
- **REQ-TOK-004** (Ubiquitous): The docs-site `--gradient-signature` token shall resolve to solid `#3d7d5f` (de-emphasized from the prior `linear-gradient(135deg, #3d7d5f 0%, #181715 100%)`).
- **REQ-TOK-005** (Unwanted): The docs-site CSS shall not contain the literal `#000000` in any background or foreground role; the ink token `#060606` is the mandatory substitute.
- **REQ-TOK-006** (Unwanted): The docs-site CSS shall not apply `background: <gradient>` and `box-shadow` simultaneously on the same element (FROZEN rule — visual noise).
- **REQ-TOK-007** (Ubiquitous): The docs-site neutral scale shall be achromatic (hue 0%) spanning `--neutral-50 #f4f4f4` through `--neutral-950 #060606`, with no warm-cream tint.
- **REQ-TOK-008** (Unwanted): The docs-site CSS shall not retain the warm-cream `--color-bg #faf9f5` value nor the Clay/Cream/Ink `--bg-page #ecefee` value after migration; a single `--color-bg #f4f4f4` replaces both.
- **REQ-TOK-009** (Unwanted): The docs-site CSS shall not retain the prior 135deg signature gradient literal (`linear-gradient(135deg, #3d7d5f 0%, #181715 100%)`) after migration.

### §B.2 Typography

- **REQ-TYP-001** (Ubiquitous): The docs-site shall use Pretendard (Variable, CDN jsdelivr v1.3.9) as the sole Korean title and body typeface.
- **REQ-TYP-002** (Unwanted): The docs-site shall not load the MaruBuri `@font-face` from `moai-docs-tokens.css` (the serif title face is retired per v2 정통 정합).
- **REQ-TYP-003** (Capability gate): **Where** the mono-font decision resolves to JetBrains Mono (per `design.md` recommendation), the docs-site shall load JetBrains Mono for code surfaces; **where** it resolves to Goorm Sans Code, the docs-site shall retain the Goorm CDN link. The decision is recorded as a [NEEDS CLARIFICATION] marker in `plan.md` until settled.
- **REQ-TYP-004** (Unwanted): The docs-site shall not use Inter, Roboto, or Arial for Korean body text (Pretendard is mandatory for Korean).
- **REQ-TYP-005** (Ubiquitous): The docs-site letter-spacing shall follow the v2 tracking tokens: display `-0.075em`, heading `-0.05em`, body `-0.025em`, caption `0`.
- **REQ-TYP-006** (Unwanted): The docs-site shall not use all-caps Korean (full-width uppercase Hangul); English abbreviations (`BETA`, `AI`, `CX`) remain permitted.

### §B.3 Layout — round3 Recreation

- **REQ-LAY-001** (Ubiquitous): The docs-site home/listing surface shall recreate the round3 `05-docs-index` layout: docs-hero (eyebrow + h1 + sub + stats) + sticky category pills (with count badges) + featured card (solid green `--gradient-signature` surface) + 4:5 aspect card grid.
- **REQ-LAY-002** (Ubiquitous): The docs-site doc-detail surface shall recreate the round3 `06-docs-detail` layout: 3-column grid (220px TOC + 720px body + 280px rail) with sticky TOC + reading-progress bar + next-CTA + breadcrumb hero.
- **REQ-LAY-003** (Ubiquitous): The docs-site layout replacement shall preserve the i18n / menu / search infrastructure: `content/<locale>/` + `_meta.yaml` + `data/menu/main.yaml` + `menu.html` / `menu/name.html` / `menu/href.html` partials + lang-switch + ⌘K search modal (`search.json`).
- **REQ-LAY-004** (Ubiquitous): The docs-site doc-content body shall adopt the DocPage article-shell contract (var-token driven: `--color-surface`, `--fg-1`, `--font-sans`, `--lh-relaxed`, `--tracking-body`, `--border-1`, `--color-primary`).
- **REQ-LAY-005** (Ubiquitous): The docs-site shall replace the geekdoc visual shell (`baseof.html` chrome, `index.html`, `single.html`-equivalent, header/footer/TOC chrome) while preserving the 4-locale navigation data flow and the Hugo section/page render pipeline.
- **REQ-LAY-006** (Capability gate): **Where** a layout slot requires a category icon, the docs-site shall resolve it via the existing `menu.html` SVG-switch case ladder (cases MUST match `data/menu/main.yaml` `icon:` values).

### §B.4 Mermaid

- **REQ-MER-001** (Ubiquitous): The docs-site mermaid init (`foot.html`) shall apply the v2 themeVariables: `primaryColor #eef4f0`, `primaryBorderColor #3d7d5f`, `primaryTextColor #060606`, `lineColor #9fa0a0`, `actorBkg #eef4f0`, `actorBorder #3d7d5f`, `noteBkgColor #e6e6e6`, `noteBorderColor #b5b5b5`, `clusterBkg #f4f4f4`, `clusterBorder #d1d1d1`, `titleColor #060606`.
- **REQ-MER-002** (Capability gate): **Where** the mermaid version decision resolves to v11 (per `design.md`), the docs-site shall bump the CDN URL from `mermaid@10` to `mermaid@11`; **where** it resolves to stay on v10, the docs-site shall apply only the themeVariables update. The decision is recorded in `plan.md` until settled.
- **REQ-MER-003** (Unwanted): The docs-site mermaid init shall not retain the prior warm-cream palette (`primaryColor #d6ebde`, `lineColor #3d7d5f`, `clusterBkg #faf9f5`) after migration.

### §B.5 Mascot Expansion

- **REQ-MAS-001** (Ubiquitous): The docs-site shall copy the 6 MoAI-Mascot pose PNGs (Thinking, Pointing, Searching, Teaching, Explaining, Coffee) from the bundle into `docs-site/static/mascots/`.
- **REQ-MAS-002** (Event-driven): **When** a reader lands on an emotional surface (home hero, section empty state, 404, loading indicator), the docs-site shall display a context-appropriate mascot pose.
- **REQ-MAS-003** (Unwanted): The docs-site shall not place a mascot on data tables, forms, or checkout surfaces (the docs-site currently has none of these, but the rule binds any future addition).
- **REQ-MAS-004** (Capability gate): **Where** a mascot is placed, the docs-site shall render it via the existing `mascot.html` shortcode (currently unused) or a new placement partial — never as a raw `<img>` in content markdown.

### §B.6 Component Adoption

- **REQ-CMP-001** (Ubiquitous): The docs-site shall adopt the v2 button recipes (primary pill + signature surface, secondary outline, ghost) mapped onto the existing `cw-*` / `gdoc-*` class overrides.
- **REQ-CMP-002** (Ubiquitous): The docs-site shall adopt the v2 card recipes (surface / outline / elevated / gradient) with `radius 16px`, `border 1px #d1d1d1`, `shadow-sm` baseline, `translateY(-2px)` + `shadow-md` hover.
- **REQ-CMP-003** (Ubiquitous): The docs-site shall adopt the v2 shadow system (`xs` / `sm` / `md` / `lg` / `xl` + `signature`) with `rgba(6,6,6,X)` alpha (ink-based, not pure black).
- **REQ-CMP-004** (Ubiquitous): The docs-site shall adopt the v2 motion defaults: `150–250ms` `cubic-bezier(0.4,0,0.2,1)` default; mascot-only bounce `cubic-bezier(0.34,1.56,0.64,1)`; page transition `cubic-bezier(0.16,1,0.3,1)` 600ms.
- **REQ-CMP-005** (Unwanted): The docs-site shall not use the rounded-border + left-color-accent card pattern (the AI-slop anti-pattern explicitly forbidden in SKILL.md).
- **REQ-CMP-006** (Unwanted): The docs-site shall not introduce body emoji in any new or modified component; the `icon.html` shortcode / Lucide SVG / typographic arrows (`→ ← ↓ ✓ ✗`) are the only permitted visual markers.
- **REQ-CMP-007** (Unwanted): The docs-site shall not use full-bleed image backgrounds; the page background is solid `--color-bg #f4f4f4` only.

### §B.7 4-Locale Parity

- **REQ-I18N-001** (Ubiquitous): Every visual change (token, layout, component, mascot, mermaid) shall land identically across all 4 locales (ko / en / ja / zh) in the same PR (same-PR obligation per oss-docs i18n rules).
- **REQ-I18N-002** (Ubiquitous): The docs-site shall preserve the canonical-locale chain (ko canonical → en → ja/zh derivation).
- **REQ-I18N-003** (Unwanted): The docs-site shall not add net-new external URLs beyond the adk.mo.ai.kr whitelist; the bundle CDN deps already in use (jsdelivr Pretendard, jsdelivr mermaid, goorm mono) are permitted, and a JetBrains Mono adoption must reuse an in-use CDN or self-host.

### §B.8 Light-Only Preservation

- **REQ-LIT-001** (State-driven): **While** the docs-site is served, the runtime shall remain LIGHT-ONLY (`baseof.html` MutationObserver that kills dark mode + dark-toggle `display:none` + mermaid light theme).
- **REQ-LIT-002** (Unwanted): The docs-site shall not re-enable dark mode even though the v2 token set (`colors_and_type.css`) carries `[data-theme="dark"]` tokens — those dark tokens are inert dead code on the docs-site surface.

### §B.9 Build Gate

- **REQ-BLD-001** (Ubiquitous): The docs-site `hugo --minify --gc` build shall complete warning-free (zero Hugo build warnings).
- **REQ-BLD-002** (Ubiquitous): The docs-site `vercel.json` redirects and the `/api/i18n-detect` edge function shall continue to operate unchanged after the migration.
- **REQ-BLD-003** (Ubiquitous): The docs-site CSS cache-busting mechanism (`head/custom.html` FNV32a `?h=` fingerprint) shall continue to fingerprint every changed CSS file.
- **REQ-BLD-004** (Capability gate): **Where** a layout/section page move occurs, the docs-site `vercel.json` shall add the corresponding locale-aware + non-locale redirects per the oss-docs i18n redirect rule.

---

## §C. Constraints (Non-Functional)

| ID | Constraint | Source |
|---|---|---|
| C-1 | `moai-brand.css` FROZEN status is lifted for this SPEC by explicit user authorization (v2 정통 정합 decision). The unfreeze rationale is recorded in §G below. | User decision 1 + SKILL.md §3 |
| C-2 | Hybrid Trunk 1-person OSS branch strategy — main-direct push is permitted for all tiers (S/M/L). The SPEC may proceed main-direct; no PR is mandated. | CLAUDE.local.md §23 |
| C-3 | 4-locale same-PR obligation — every canonical visual edit MUST land in all 4 locales simultaneously (locale-parity threshold 1.0, must_pass). | hns-oss-docs-i18n-rules §2 |
| C-4 | adk.mo.ai.kr URL whitelist — no net-new external URLs. | hns-oss-docs-i18n-rules §6 |
| C-5 | LIGHT-ONLY maintained — do NOT re-enable dark mode. | User scope decision + ground truth |
| C-6 | No body emoji in any new/modified component (icon shortcode / Lucide / typographic arrows only). | hns-oss-docs-i18n-rules §4 + SKILL.md §3 |
| C-7 | Mermaid TD-only (`flowchart TD` / `graph TB`); no `LR` / `RL`. | hns-oss-docs-i18n-rules §3 |
| C-8 | Do NOT touch the pre-existing uncommitted files `.moai/config/sections/llm.yaml` and `README.ko.md` (unrelated WIP; preserve as-is). | User constraint |
| C-9 | Preserve the i18n / menu / search infrastructure even as the visual shell is replaced. | User constraint + ground truth |
| C-10 | `era: V3R6`, `tier: L`, harness level = `thorough` (full sync-auditor + TRUST 5). plan-auditor will audit after plan-phase. | User directive |
| C-11 | Emphasis-marker spacing rule: `**바이브코딩** (Vibe Coding)` — parenthetical OUTSIDE the markers. | hns-oss-docs-i18n-rules §5 |
| C-12 | Version SSOT: `docs-site/hugo.toml` `params.version` / `params.releaseDate` is the single version surface; do not hardcode divergent version strings. | hns-oss-docs-i18n-rules §7 |

---

## §D. Acceptance Criteria Summary

The full AC matrix (Given-When-Then scenarios, severity, traceability) lives in `acceptance.md`. Headline count: **30 ACs** spanning token parity (grep assertions), font-load assertions, layout slot presence per round3, mascot presence on N emotional surfaces, 4-locale parity, warning-free Hugo build, and mermaid v2 palette.

---

## §E. Out of Scope

### Out of Scope — moai web console

- The moai web console design-system v2 migration is **SPEC-2** of Epic DESIGN-V2, authored later. This SPEC touches `docs-site/` only; `internal/webconsole/` (or equivalent) is untouched.

### Out of Scope — mo.ai.kr marketing site

- The mo.ai.kr marketing site lives in a separate repo and is covered by the bundle's own `ui_kits/website/`. Not touched here.

### Out of Scope — Content rewrites

- No documentation copy changes unless structurally required by a new layout slot (e.g. a round3 detail hero requires a deck/subtitle field — that structural copy is in scope; rewriting prose body content is out).

### Out of Scope — Dark mode

- Dark mode is explicitly excluded. The v2 `colors_and_type.css` `[data-theme="dark"]` block is imported as dead code; the docs-site stays LIGHT-ONLY (REQ-LIT-001/002). Do not wire a dark toggle.

### Out of Scope — Net-new documentation pages

- Scope is design-system application to the existing page set, not content expansion. New doc pages, new sections, new _meta.yaml entries are out of scope.

### Out of Scope — v2 token authoring / bundle edits

- The bundle at `.moai/state/ai-design-system/project/` is a read-only handoff. This SPEC consumes it; it does not edit `colors_and_type.css`, `README.md`, or any bundle asset. Bundle edits are a separate brand-SSOT workflow.

---

## §F. Token Delta Reference

The authoritative current → v2 token DELTA table (per-token: current docs-site value → v2 target value, with file-of-origin) lives in `research.md` §C. spec.md requirements REQ-TOK-001 through REQ-TOK-009 reference that table as evidence.

---

## §G. moai-brand.css Unfreeze Authorization

**Authorization**: The `moai-brand.css` file in `docs-site/static/` is currently marked FROZEN (per CLAUDE.local.md §17.1 and the oss-docs structure-map). The user-confirmed decision **"v2 정통 정합" (full faithful adoption of the bundle's v2 design system)** explicitly authorizes unfreezing and rewriting `moai-brand.css` for this migration.

**Rationale**: The FROZEN warm-cream token vocabulary (`--color-bg #faf9f5`, `--gradient-signature linear-gradient(135deg, #3d7d5f 0%, #181715 100%)`, warm neutral ramp) is fundamentally incompatible with the v2 achromatic system (`--color-bg #f4f4f4`, solid `--gradient-signature #3d7d5f`, hue-0% neutral scale). A faithful v2 adoption cannot be achieved by additive override layers alone — the FROZEN `:root` block must be rewritten. The token-unification architecture in `design.md` §A defines the collapse strategy (single v2 token file vs. layered override).

**Scope of unfreeze**: Limited to this SPEC's run-phase. The new `moai-brand.css` (or its successor token file) is re-stamped FROZEN at sync-phase close, with the v2 token vocabulary as the new frozen baseline.

---

## §H. Cross-References

- **Epic**: DESIGN-V2 (모두의AI Design System v2 전면 적용). Sibling: SPEC-2 (moai web console, TBD).
- **Design bundle SSOT**: `.moai/state/ai-design-system/project/` (`README.md`, `SKILL.md`, `colors_and_type.css`, `round3/`, `assets/components/`, `assets/characters/`).
- **i18n rules SSOT**: `.moai/docs/docs-site-i18n-rules.md` + Skill `hns-oss-docs-i18n-rules`.
- **Design-icon regime**: CLAUDE.local.md §17.1 (Claude Warm Editorial → superseded by v2 정통 정합 per user decision 1).
- **Branch strategy**: CLAUDE.local.md §23 (Hybrid Trunk 1-person OSS).
- **plan.md**: `.moai/specs/SPEC-DESIGN-DOCSV2-001/plan.md` — 7 milestones (M1–M7).
- **acceptance.md**: `.moai/specs/SPEC-DESIGN-DOCSV2-001/acceptance.md` — 30 ACs.
- **research.md**: `.moai/specs/SPEC-DESIGN-DOCSV2-001/research.md` — v2 token core, FROZEN rules, round3 breakdown, 3-layer divergence, mono-font decision, mermaid version decision, token DELTA table.
- **design.md**: `.moai/specs/SPEC-DESIGN-DOCSV2-001/design.md` — token-unification architecture, layout architecture, component mapping, mascot placement, 4-locale parity, light-only preservation.
