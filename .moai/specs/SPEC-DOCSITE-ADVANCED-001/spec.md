---
id: SPEC-DOCSITE-ADVANCED-001
title: "docs-site v3.0 Advanced Guides — 6-page × 4-locale content expansion + advanced/_meta.yaml parity debt fix"
version: "0.1.0"
status: draft
created: 2026-07-13
updated: 2026-07-13
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "docs-site (content/{en,ko,ja,zh}/advanced/ + data/menu/main.yaml)"
lifecycle: spec-anchored
tags: "docs-site, i18n, 4-locale, v3.0, tokenomics, harness, autonomous-loops, hugo, content-expansion"
era: V3R6
tier: L
related_specs: [SPEC-DOCSITE-E2E-001, SPEC-MODEL-TIER-PLANTYPE-001, SPEC-TOKEN-BUDGET-STOP-001, SPEC-HARNESS-EVOLVE-001, SPEC-HARNESS-EVOLVE-002, SPEC-HARNESS-EVOLVE-003, SPEC-GOAL-ENGINE-001]
---

# SPEC-DOCSITE-ADVANCED-001 — docs-site v3.0 Advanced Guides content expansion (6 pages × 4 locales) + _meta.yaml parity debt fix

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-13 | manager-spec | Plan-phase artifact set authored (Tier L, 6 artifacts: spec.md + plan.md + acceptance.md + progress.md + design.md + research.md). Drives the v3.0 3-pillar narrative (Tokenomics · Agentic Loop · Agentic Harness) into the docs-site Advanced section as 6 new pages × 4 locales. Resolves a pre-existing `advanced/_meta.yaml` 4-locale parity debt discovered during plan-phase recon — the gap is wider than the brief described (KO missing 3 entries; EN/JA/ZH missing 6 — see research.md §A). |

---

## §A Context & Problem

### A.1 Problem

The docs-site Advanced section (`/advanced`) currently hosts 14 component-reference pages per locale (skill-guide, agent-guide, builder-agents, hooks-guide, hooks-reference, settings-json, security-notes, statusline, claude-md-guide, harness-profiles, catalog-system, decision-memory, harness-v4-builder, ultracode-workflows). It does NOT yet document MoAI-ADK v3.0's three product-differentiating pillars:

1. **Tokenomics** — the 4-layer token economy (metering · routing · verify-diet · budget-defense) plus the plan_type × tier × agent 60-cell model profile.
2. **Agentic Loop** — autonomous continuation primitives (`/goal`, `/moai goal`, `/moai loop`) and their architectural rationale.
3. **Agentic Harness** — the 3-tier no-Haiku agent architecture, the self-evolving harness (ACE 3-Loop), and how plan_type profiles route work.

This SPEC adds 6 new Advanced pages, each authored in canonical KO and derived across en/ja/zh in the same PR (per `.claude/skills/hns-oss-docs-i18n-rules` canonical-locale chain `ko → en → ja/zh`).

### A.2 Pre-existing parity debt (verified 2026-07-13, this tree)

Recon measured a wider `advanced/_meta.yaml` parity gap than the brief described:

| Locale | Current entries | Content files present | _meta delta |
|--------|----------------|----------------------|-------------|
| ko     | 11             | 14                   | missing 3: decision-memory, harness-v4-builder, ultracode-workflows |
| en     | 8              | 14                   | missing 6: catalog-system, decision-memory, harness-profiles, harness-v4-builder, hooks-reference, ultracode-workflows |
| ja     | 8              | 14                   | missing 6: (same set as en) |
| zh     | 8              | 14                   | missing 6: (same set as en) |

**All 14 content files exist in every locale** (verified via `find` — see research.md §A.1 for the verbatim output). The debt is purely `_meta.yaml` registration, NOT a content gap. The sidebar menu (`main.yaml`) already carries all 14 entries with correct 4-locale name fields — so the visible navigation is correct; only the geekdoc page-title resolver (`_meta.yaml`) is broken per-locale.

This debt MUST be resolved BEFORE or ALONGSIDE the 6-page addition so the 6 new entries land on a clean parity baseline.

### A.3 Content-source readiness (per-page assessment)

Per `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3, the agent verified each page's content source rather than accepting the brief's "may need source consolidation" caveat at face value:

| Page | Source | Status | Honesty caveat required |
|------|--------|--------|-------------------------|
| tokenomics-overview | `project_v3_tokenomics_docs_plan` memory + `.moai/reports/readme-docs-redesign-20260713.{html,md}` + `.moai/reports/readme-draft-v3-rc11.md` (3-pillar narrative) + Token-Economy Epic A-D | READY | None — describes product positioning |
| token-budget | `.moai/specs/SPEC-TOKEN-BUDGET-STOP-001/` + `.claude/rules/moai/workflow/{context-window-management,session-handoff}.md` + verify-diet doctrine | READY | None — behavior is implemented |
| no-haiku-3tier | `.moai/reports/agent-architecture-redesign-v2-20260709.html` + `project_model_tier_plantype_001_completed` (ApplyTierProfile live) | READY | Disclose design-vs-runtime distinction (see REQ-DA-051) |
| plan-type-profiles | `.moai/specs/SPEC-MODEL-TIER-PLANTYPE-001/` (CLOSED) + `.moai/reports/model-tier-redesign-20260712.{html,md}` | READY (authoritative) | GLM overlay wire-effectiveness UNVERIFIED (see REQ-DA-050) |
| self-evolving | `.moai/reports/harness-self-evolving-redesign-final-20260712.html` (v5.1 FINAL SSOT) + SPEC-HARNESS-EVOLVE-{001,002,003} (3/5 closed) + Lilian Weng reference | READY | Disclose EVOLVE-004/005 pending (console + Recall wiring) |
| autonomous-loops | `.moai/specs/SPEC-GOAL-ENGINE-001/` (CLOSED) + AGENTIC-CORE epic (in progress) + `.claude/rules/moai/workflow/goal-directive.md` (canonical /goal reference) | READY | Distinguish native `/goal` (HUMAN-ONLY) from `/moai goal` (PROGRAMMATIC) (see REQ-DA-052) |

**All 6 pages have substantially-ready sources — NO blockers, NO pages deferred.** Full evidence in research.md §B.

### A.4 Tier L vs Epic decision

This SPEC is classified **Tier L** (single SPEC, NOT an Epic of 2-3 SPECs). Rationale (full version in plan.md §F.0):

- The 6 pages share **one cross-cutting structural dependency**: the `_meta.yaml` parity debt fix must land in the same PR as the 6 new entries (otherwise 6 new entries stack on a broken baseline).
- The 3-pillar narrative is a **shared spine** — pages cross-reference each other; coherence is best maintained in one plan-phase cycle.
- The 4-locale derivation discipline + design regime (Claude Warm Editorial) is identical across all 6 pages.
- An Epic would triplicate plan-phase overhead without removing any per-page work.

---

## §B Requirements (GEARS notation)

The `<subject>` field uses the generalized GEARS form (any noun). The compound clause `[Where ...][While ...][When ...]` is applied where multiple modifiers chain.

### Group A — 4-Locale Content Parity (the 24 new files)

**REQ-DA-001** (Ubiquitous): The docs-site advanced section **shall** contain exactly 6 new page slugs — `tokenomics-overview`, `token-budget`, `no-haiku-3tier`, `plan-type-profiles`, `self-evolving`, `autonomous-loops` — each present in all 4 locales (ko, en, ja, zh), yielding exactly 24 new markdown files.

**REQ-DA-002** (Ubiquitous): Each of the 24 new files **shall** carry the canonical Hugo frontmatter (`---\ntitle: ...\n---`) consistent with the existing advanced pages, and the ko file **shall** be the canonical authoring source per the i18n canonical-locale chain.

**REQ-DA-003** (State-driven): **While** all 4 locales of a given slug exist, the page content **shall** preserve cross-locale semantic parity (the same concepts, code snippets, Mermaid diagrams, and structural sections, locale-translated — NOT structurally divergent pages).

### Group B — Canonical KO Authoring (the 6 pages)

**REQ-DA-010** (Ubiquitous): The `ko/advanced/tokenomics-overview.md` page **shall** present the 3-pillar product narrative (🪙 Tokenomics · 🔁 Agentic Loop · 🤖 Agentic Harness) with Tokenomics as the docs-site v3.0 product-differentiation pillar, sourced from `.moai/reports/readme-docs-redesign-20260713.md` + `project_v3_tokenomics_docs_plan` memory.

**REQ-DA-011** (Ubiquitous): The `ko/advanced/token-budget.md` page **shall** document the 4-layer token economy (metering · routing · verify-diet · budget-defense), the model-specific `/clear` thresholds (1M = 50%, 200K/256K = 90%), the paste-ready resume pattern, and the `verify-diet` file-redirect contract.

**REQ-DA-012** (Ubiquitous): The `ko/advanced/no-haiku-3tier.md` page **shall** document the 3-tier agent architecture (Sonnet low / Opus execution / Fable reasoning, with Haiku excluded from the routing model set), the DeepSWE-leaderboard rationale, and the ApplyTierProfile implementation hook.

**REQ-DA-013** (Ubiquitous): The `ko/advanced/plan-type-profiles.md` page **shall** document the `plan_type ∈ {api, subscription}` axis and the 60-cell profile (10 agents × 3 tiers × 2 plan_types), sourced from SPEC-MODEL-TIER-PLANTYPE-001 (CLOSED) + `.moai/reports/model-tier-redesign-20260712.md`.

**REQ-DA-014** (Ubiquitous): The `ko/advanced/self-evolving.md` page **shall** document the ACE 3-Loop harness self-evolution architecture (Loop 0 observation → Loop 1 reflection → Loop 2 promotion), sourced from `.moai/reports/harness-self-evolving-redesign-final-20260712.html` (v5.1 FINAL SSOT) + the closed EVOLVE-001/002/003 SPECs.

**REQ-DA-015** (Ubiquitous): The `ko/advanced/autonomous-loops.md` page **shall** document the autonomous-continuation primitives (`/goal`, `/moai goal`, `/moai loop`) and their distinct trigger semantics (user-TUI vs programmatic vs diagnostic-driven), sourced from `.claude/rules/moai/workflow/goal-directive.md` + SPEC-GOAL-ENGINE-001.

### Group C — 4-Locale Derivation

**REQ-DA-020** (Capability gate): **Where** the canonical-ko page exists, the en/ja/zh derivations **shall** follow the chain `ko → en → ja/zh` per `.claude/skills/hns-oss-docs-i18n-rules` §1 (canonical-locale chains).

**REQ-DA-021** (Ubiquitous): Each derived-locale page **shall** translate the prose, preserve code blocks verbatim (commands, file paths, identifiers stay English), preserve Mermaid diagrams verbatim, and use the locale-specific name field from `main.yaml` when referencing navigation entries.

### Group D — Navigation Registration

**REQ-DA-030** (Ubiquitous): The `docs-site/data/menu/main.yaml` Advanced section **shall** register 6 new sub-entries — one per slug — each carrying the 4-locale name map (ko/en/ja/zh) and the `ref: /advanced/<slug>` field, in the pillar order specified in design.md §C.

**REQ-DA-031** (Ubiquitous): Each per-locale `docs-site/content/<locale>/advanced/_meta.yaml` **shall** carry an entry for each of the 6 new slugs, with the locale-specific title matching the `main.yaml` name field for that locale.

**REQ-DA-032** (Capability gate): **Where** a new slug is registered in `main.yaml`, the slug **shall** ALSO be registered in all 4 per-locale `_meta.yaml` files in the same PR — partial registration (slug in `main.yaml` but missing from `_meta.yaml` for some locale) is the exact parity-debt pattern this SPEC resolves and is prohibited.

### Group E — Parity Debt Pre-Fix (M1)

**REQ-DA-040** (Ubiquitous): Prior to or alongside the 6-page addition, each per-locale `_meta.yaml` **shall** carry an entry for ALL 14 existing advanced content files (closing the parity debt documented in §A.2): ko adds decision-memory, harness-v4-builder, ultracode-workflows; en/ja/zh add catalog-system, decision-memory, harness-profiles, harness-v4-builder, hooks-reference, ultracode-workflows.

**REQ-DA-041** (State-driven): **While** the parity debt fix is in flight, the existing `main.yaml` Advanced section (which already correctly registers all 14 existing entries with 4-locale names) **shall not** be modified — the debt is in `_meta.yaml` only, and `main.yaml` is the reference structure.

### Group F — Content Integrity (Claude Warm Editorial design regime)

**REQ-DA-050** (Unwanted behavior): The 24 new page bodies **shall not** contain body-emoji decoration (`📖 💡 🚀 ✨ 🎉 🔥 📌` etc.) — semantic markers MUST use the `{{</* icon <name> [variant] */>}}` shortcode (per `.claude/skills/hns-oss-docs-i18n-rules` §4 + CLAUDE.local.md §17.1).

**REQ-DA-051** (State-driven): **While** the page is rendered, typography arrows (`→ ← ↓ ✓ ✗`) and MoAI orchestrator-banner reproduction emoji inside fenced code blocks **shall** be preserved verbatim — these are NOT body-emoji and MUST NOT be stripped or icon-substituted.

**REQ-DA-052** (Ubiquitous): All Mermaid diagrams in the 24 new pages **shall** use only `flowchart TD` or `graph TB` direction — `flowchart LR`, `graph LR`, `flowchart RL`, `graph RL` are prohibited per i18n rule §3.

**REQ-DA-053** (Ubiquitous): All hyperlinks in the 24 new pages **shall** use only the `adk.mo.ai.kr` domain for internal docs-site links — `docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr` are blacklisted per i18n rule §6.

**REQ-DA-054** (Ubiquitous): Emphasis-marker spacing **shall** keep parenthetical English glosses OUTSIDE the markers — `**한글 용어** (English Gloss)` is correct; `**한글 용어(English Gloss)**` is prohibited per i18n rule §5.

**REQ-DA-055** (State-driven): **While** the page is rendered, the light single-theme Claude Warm Editorial design **shall** be respected — no `[data-theme="dark"]` branching in any new content; coral `#cc785c` accent usage follows `static/moai-brand.css` (FROZEN) + `static/moai-design.css`.

### Group G — Source Truthfulness (verification-claim-integrity binding)

**REQ-DA-060** (Ubiquitous): The `plan-type-profiles.md` page **shall** disclose that the GLM backend effort overlay's wire-effectiveness is **unverified pending live GLM outbound observation** — the page MUST describe it as "implemented + wired, wire validity pending live verification", NOT as "works guaranteed" (per `project_model_tier_plantype_001_completed` §HARD honesty constraint).

**REQ-DA-061** (Ubiquitous): The `no-haiku-3tier.md` page **shall** distinguish design-stage intent (the v2 architecture report) from implemented behavior (the live `ApplyTierProfile` 60-cell profile per SPEC-MODEL-TIER-PLANTYPE-001) — readers MUST be able to tell what is design vs what ships today.

**REQ-DA-062** (Ubiquitous): The `autonomous-loops.md` page **shall** distinguish the native Claude Code `/goal` command (HUMAN-ONLY TUI command the model cannot invoke) from the MoAI-owned `/moai goal` programmatic counterpart (Axis B per `.claude/rules/moai/workflow/native-invocation-model.md`), and from `/moai loop` (Ralph Engine diagnostic preset).

**REQ-DA-063** (Ubiquitous): The `self-evolving.md` page **shall** disclose which Loop 2 surfaces are PRODUCTION-wired (EVOLVE-001/002/003 closed) vs which remain in flight (EVOLVE-004 console verbs + EVOLVE-005 Recall wiring + typed parser) — readers MUST be able to tell what is live today vs roadmap.

### Group H — Hugo Build & Site Integrity

**REQ-DA-070** (Event-driven): **When** a maintainer runs `cd docs-site && hugo --minify --gc`, the build **shall** exit 0 with ZERO warnings — malformed `_meta.yaml` or `main.yaml` entries surface here.

**REQ-DA-071** (Ubiquitous): The build **shall** generate the 24 new pages in the sitemap (`docs-site/public/sitemap.xml`) at the paths `/{locale}/advanced/<slug>/index.html` for each of the 4 locales × 6 slugs.

**REQ-DA-072** (Ubiquitous): The hugo build **shall** complete without "page not found" warnings for any of the 6 new slugs in any locale, confirming `_meta.yaml` and `main.yaml` registrations are consistent.

---

## §C Constraints

### C.1 Design regime (Claude Warm Editorial — frozen)

- Light single theme (no dark-theme branching in new content).
- `static/moai-brand.css` is FROZEN — never edited in this SPEC.
- Code blocks render via the existing `layouts/_default/_markup/render-codeblock.html` hook (macOS dark card); no new render hook added.
- Mermaid via existing `foot.html` CDN UMD (`mermaid@10`, theme `'base'`, coral themeVariables); no new theme JS.
- Icon shortcodes via existing `layouts/shortcodes/icon.html` (variants: `ok|warn|danger|primary|muted`); no new shortcode added.

### C.2 Tooling reality

- `gen_menu.py` referenced in the comment header of `main.yaml` DOES NOT exist — menu edits are manual (per `hns-oss-docs-structure-map` skill "Tooling reality").
- `docs-i18n-check.sh` also DOES NOT exist — verification uses the inline recipe in `.claude/skills/hns-oss-docs-verify` (the canonical verify skill, loaded at run phase).

### C.3 Scope boundaries

- This SPEC is **docs-only**: zero Go code, zero template-tree (`internal/template/templates/`) writes, zero doctrine-tree (`.claude/rules/`) writes, zero README writes (the v3.0 README redesign is a separate work surface tracked by `project_readme_v3_rc11_redesign_draft` memory and is NOT in this SPEC's scope).
- The docs-site is moai-adk-go local product documentation — NOT a distributed template. No template-tree mirror is required (CLAUDE.local.md §2 Template-First Rule does not bind this SPEC).

---

## §D Acceptance Criteria Matrix (summary)

The full Given-When-Then AC enumeration lives in `acceptance.md`. Summary by group:

| Group | AC count | must_pass | Severity profile |
|-------|----------|-----------|------------------|
| A — 4-locale file parity | 4 | all | Critical |
| B — 6 pages canonical KO authored | 6 | all | Critical |
| C — Derivation chain | 3 | all | Critical |
| D — Navigation registration | 4 | all | Critical |
| E — Parity debt pre-fix | 3 | all | Critical |
| F — Design regime | 6 | all | High |
| G — Source truthfulness | 4 | all | High |
| H — Hugo build | 3 | all | Critical |
| **Total** | **33** | **all** | — |

---

## §E Exclusions

### Out of Scope — v3.0 README redesign

- The 3-pillar README rewrite (`README.md` / `README.ko.md` / `README.ja.md` / `README.zh.md`) tracked by `project_readme_v3_rc11_redesign_draft` and `project_v3_tokenomics_docs_plan` is NOT in this SPEC's scope. The docs-site pages MAY cross-reference the README, but the README files themselves are out of scope.

### Out of Scope — template-tree mirror

- `internal/template/templates/` is NOT touched. The docs-site is moai-adk-go local product documentation, not a user-distributed template. CLAUDE.local.md §2 Template-First Rule and §15 Language Neutrality do not bind this SPEC.

### Out of Scope — new menu icons / menu.html SVG cases

- No new top-level `icon:` values are introduced in `main.yaml`. All 6 new Advanced sub-entries inherit the existing `icon: school` from the parent Advanced section. `layouts/partials/menu.html` SVG switch is NOT modified.

### Out of Scope — vercel.json redirects

- `vercel.json` `redirects` array is NOT modified. Redirects are required only for moved/renamed pages; this SPEC adds NEW pages, so no redirect entries.

### Out of Scope — FROZEN CSS / theme JS

- `static/moai-brand.css` (FROZEN) is not edited. `static/moai-design.css` is not edited. `layouts/partials/head/custom.html` (font loader) is not edited. `layouts/partials/foot.html` (Mermaid CDN UMD) is not edited.

### Out of Scope — Implementation of in-flight source SPECs

- EVOLVE-004 (console verbs) and EVOLVE-005 (Recall wiring + typed parser) are NOT implemented by this SPEC. The `self-evolving.md` page documents them as roadmap items per REQ-DA-063.
- AGENTIC-CORE SPEC-2 (autonomous/semi-autonomous kickoff REQ) is NOT implemented. The `autonomous-loops.md` page documents the implemented subset (`/moai goal` per SPEC-GOAL-ENGINE-001) and references AGENTIC-CORE as roadmap.

### Out of Scope — automated translation tooling

- No new automation for ko→en→ja/zh derivation is introduced. The 4-locale derivation is performed by the existing oss-docs harness specialists (content-author + locale-translator) per `.claude/agents/harness/` and the `hns-oss-docs-*` skill family.

---

## §F Cross-References

- **Skills loaded at run-phase start**: `hns-oss-docs-i18n-rules` (HARD i18n rules), `hns-oss-docs-structure-map` (path/schema map), `hns-oss-docs-verify` (verify recipe), `hns-oss-docs-readme-sync` (NOT used — README out of scope).
- **i18n SSOT**: `.moai/docs/docs-site-i18n-rules.md` + CLAUDE.local.md §17 / §17.1 (design regime).
- **Content-source SPECs** (related_specs in frontmatter): SPEC-MODEL-TIER-PLANTYPE-001, SPEC-TOKEN-BUDGET-STOP-001, SPEC-HARNESS-EVOLVE-001/002/003, SPEC-GOAL-ENGINE-001, SPEC-DOCSITE-E2E-001 (predecessor pattern for 4-locale docs work).
- **Memory topics** (content sources): `project_v3_tokenomics_docs_plan`, `project_model_tier_plantype_001_completed`, `project_harness_evolve_epic`, `project_goal_engine_cli_gap_handoff`, `project_agentic_core_epic_progress`, `project_readme_v3_rc11_redesign_draft`, `project_agent_token_cost_color_tiers`.
- **Design reports** (canonical SSOTs per page): `.moai/reports/{readme-docs-redesign-20260713,model-tier-redesign-20260712,harness-self-evolving-redesign-final-20260712,agent-architecture-redesign-v2-20260709}.md` (the .md twin is the agent-context canonical per `project_v3_tokenomics_docs_plan`).
- **Plan-phase doctrine**: `.claude/rules/moai/development/spec-frontmatter-schema.md` (frontmatter schema), `.claude/rules/moai/workflow/sprint-round-naming.md` (Epic/SPEC/Milestone taxonomy), `.claude/skills/moai-workflow-spec/SKILL.md` (GEARS notation).

---

Version: 0.1.0 | Tier: L | Era: V3R6 | Status: draft | Author: manager-spec
