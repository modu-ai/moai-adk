---
id: SPEC-V3R6-DOCS-V3-REBUILD-001
title: "Documentation Site (adk.mo.ai.kr) Full Rebuild to v3.0 Accuracy"
version: "0.1.1"
status: completed
created: 2026-07-01
updated: 2026-07-02
author: manager-docs
priority: P1
phase: "v3.0.0"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs, docs-site, i18n, rebuild"
era: V3R6
---

# SPEC-V3R6-DOCS-V3-REBUILD-001 — Documentation Site Full Rebuild to v3.0 Accuracy

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-07-01 | 0.1.0 | manager-spec | Initial plan-phase authoring. Tier L full rebuild of adk.mo.ai.kr to v3.0.0-rc4 accuracy: IA redesign + validate-then-rewrite of all 380 content files + CC-mirror refresh to latest Claude Code + 4 new v3.0 feature pages + drift correction. Ground truth extracted from live codebase (13 commands, 8 agents, 27 template skills, 3-phase lifecycle). |
| 2026-07-01 | 0.1.1 | manager-spec | plan-auditor PASS-WITH-DEBT (0.83) fix pass. D1: REQ-DVR-013 extended from 2 to all 4 README files (README.ja.md/README.zh.md carry worse drift — retired `/moai design` Design System section ~L926-1061, `/agency` migration refs L57/L63, `/moai coverage` in active workflow chains L623/L633 — verified by direct grep). D2: version-SSOT ownership bullet (M0.6, hugo.toml L55/L56) added to plan.md. D3: AC-DVR-012a given a grep anchor. D4: AC-DVR-019a labelled manual-observation. D5: AC-DVR-014a whitelist reconciled with REQ-DVR-013 (4 README files). |

## Overview (WHY)

The MoAI-ADK documentation site at `adk.mo.ai.kr` has accumulated significant factual drift against the actual v3.0.0-rc4 codebase. Concrete observed drift:

- **Version**: `hugo.toml params.version` says `v3.0.0-rc2`; homepage says "릴리스 후보 2"; `installation.md` shows rc2 examples. Actual binary is **v3.0.0-rc4** (built 2026-06-23, commit `3319defdf`).
- **Command count**: docs and README claim "12 total" while listing 13; retired subcommands (`coverage`, `e2e`, `design`, `brain`, `security` — removed by SPEC-SUBCOMMAND-RETIRE-001) are still documented in README.
- **Skill count**: README says "30 `moai-*` skills", `introduction.md` says "32개 스킬" — both are drift; the template source ships **27 `moai-*` skills**.
- **Agent count**: `advanced/agent-guide.md` has a garbled line (`manager-develop` repeated 6×) and stray archived `expert-*` references.
- **Architecture**: `advanced/builder-agents.md` describes an obsolete 3-builder model (`builder-skill`/`builder-agent`/`builder-plugin`); reality is the Harness v4 Builder.
- **CC mirror**: the `claude-code/` mirror pages exist but predate the latest Claude Code feature set (goal, workflows, scheduled-tasks, agent-teams, agent-view, plugins, MCP, hooks refresh).
- **Missing pages**: four v3.0 features have no dedicated pages — Decision Memory (`moai preference`), `moai inventory`, Harness v4 Builder, `/effort ultracode` dynamic workflows.

Drift in user-facing documentation erodes trust and misleads adopters. A full rebuild to v3.0 accuracy restores the docs as an authoritative reference and establishes a clean baseline before GA.

## Scope Depth (user-approved)

**FULL REBUILD** — redesign the Information Architecture (IA) and rewrite all 380 content files (95 per locale × 4 locales). Sections already accurate (`core-concepts`, `workflow-commands`, `utility-commands`, `quality-commands`, `multi-llm`) are rewritten via **validate-then-rewrite** (reuse the accurate content as a draft baseline, validate against ground truth, then rewrite) — NOT blank-page — to avoid wasted rework. Real effort concentrates on IA redesign, new pages, CC-latest refresh, and drift pages.

## Ground Truth (authoritative = live codebase, NOT README)

| Fact | Value | Source of truth |
|------|-------|-----------------|
| Version | `v3.0.0-rc4` (built 2026-06-23, commit `3319defdf`) | binary / release process |
| `/moai` commands | **13**: clean, codemaps, feedback, fix, gate, harness, loop, mx, plan, project, review, run, sync | `.claude/commands/moai/` |
| Retained agents | **8** = 7 MoAI-custom (manager-spec, manager-develop, manager-docs, manager-git, plan-auditor, sync-auditor, builder-harness) + Explore (built-in) | `.claude/agents/moai/` |
| Retired subcommands (GONE) | coverage, e2e, design, brain, security | SPEC-SUBCOMMAND-RETIRE-001 |
| Skills (doc-facing) | **27 `moai-*`** template-managed (+ 1 `moai` base skill; 2 `harness-*` are user-owned, local-only, NOT template) | `internal/template/templates/.claude/skills/` |
| Lifecycle | **3-phase** (plan → run → sync; Mx folded into sync) | SPEC-V3R6-LIFECYCLE-REDESIGN-001 |
| Go code scale | 100K+ lines across 100+ packages, 85-100% coverage | codebase |
| Content files | 380 total (95/locale × 4) | `docs-site/content/{ko,en,ja,zh}` |
| CC mirror | `claude-code/` = 28 pages/locale × 4 = 112 files | `docs-site/content/*/claude-code/` |

## Requirements (GEARS)

### Accuracy foundation (Ubiquitous)

- **REQ-DVR-001** (Version accuracy): The docs-site **shall** reflect `v3.0.0-rc4` as the current version across every version-bearing surface (`hugo.toml` `params.version` + `releaseDate`, homepage `_index.md`, `getting-started/introduction.md`, `getting-started/installation.md`) in all 4 locales.
- **REQ-DVR-002** (Command count accuracy): The docs-site **shall** document exactly the 13 `/moai` commands (clean, codemaps, feedback, fix, gate, harness, loop, mx, plan, project, review, run, sync) and **shall not** reference any retired subcommand (coverage, e2e, design, brain, security).
- **REQ-DVR-003** (Agent count accuracy): The docs-site **shall** document exactly the 8 retained agents (7 MoAI-custom + Explore) and **shall not** present any archived `expert-*` agent as an active agent.
- **REQ-DVR-004** (Skill count accuracy): The docs-site skill count **shall** be derived from the template source `internal/template/templates/.claude/skills/` (27 `moai-*` skills), NOT the local dev project, and **shall** be consistent across all pages and all 4 locales.
- **REQ-DVR-005** (Lifecycle accuracy): The docs-site **shall** describe the 3-phase lifecycle (plan → run → sync) and **shall not** reintroduce a 4th "Mx phase" as a separate lifecycle phase.

### Rebuild depth and IA

- **REQ-DVR-006** (Full rebuild scope): The rebuild **shall** redesign the IA and rewrite all 380 content files. **Where** a section is already accurate (`core-concepts`, `workflow-commands`, `utility-commands`, `quality-commands`, `multi-llm`), the rewrite **shall** use validate-then-rewrite (accurate content reused as draft baseline), NOT blank-page.
- **REQ-DVR-007** (New IA blueprint): The docs-site **shall** adopt a new IA blueprint / sitemap that reconciles the navigation menu (`data/menu/main.yaml`) with the actual content sections, including sections currently present in content but absent from the menu (`cost-optimization`).
- **REQ-DVR-008** (CC mirror refresh): The `claude-code/` mirror (28 pages/locale × 4 = 112 files) **shall** be refreshed to reflect the latest Claude Code feature set (goal, workflows, scheduled-tasks, agent-teams, agent-view, plugins, MCP, hooks), backed by the run-phase research (REQ-DVR-020).
- **REQ-DVR-009** (New feature pages): The docs-site **shall** add 4 new v3.0 feature topics — Decision Memory (`moai preference`: 3-tier memory, adaptive recommendation, decay), `moai inventory` command, Harness v4 Builder, `/effort ultracode` dynamic workflows — each present in all 4 locales (16 pages).

### Drift-specific rewrites

- **REQ-DVR-010** (builder-agents rewrite): `advanced/builder-agents.md` **shall** be rewritten to describe the Harness v4 Builder (4-phase ANALYZE → PLAN → GENERATE → ACTIVATE + manifest-driven Runner + conditional worktree isolation), replacing the obsolete 3-builder model.
- **REQ-DVR-011** (agent-guide correction): `advanced/agent-guide.md` **shall** be corrected — the garbled repeated line removed, stray archived-agent references removed, and dynamic-team spawning (`Agent(general-purpose)` with domain context) added.
- **REQ-DVR-012** (harness reconcile): `workflow-commands/moai-harness.md` **shall** reconcile the v3 learning-system description against the v4 Builder so the page presents one coherent, current harness model.
- **REQ-DVR-013** (README drift correction): All 4 repo-root README files (`README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`) **shall** be corrected — remove retired subcommand rows (`coverage`, `e2e`, `design`), correct the command count to 13, and correct the skill count to 27. **Where** a locale carries additional drift — `README.ja.md` / `README.zh.md` retain a retired `/moai design` Design System section (~L926-1061), `/agency → /moai design` v2.12.0 migration references (L57/L63), `/moai coverage` inside active workflow chains (L623/L633), and stale `v2.6.0`-as-current version references — that locale **shall** additionally have those surfaces corrected. This aligns all 4 README files with the same 4-locale parity discipline as REQ-DVR-015; the ja/zh under-scoping MUST NOT be deferred silently.

### Constraints (design + parity)

- **REQ-DVR-014** (Theme preservation): The rebuild **shall** keep the `hugo-geekdoc` theme with NO visual, layout, or CSS change; the work is content-only.
- **REQ-DVR-015** (4-locale parity): **When** a content change lands for the canonical locale (ko), the docs-site **shall** reflect the same change in en, ja, and zh within the same milestone; file counts and paths **shall** match across all 4 locales.
- **REQ-DVR-016** (Menu 4-locale rebuild): `data/menu/main.yaml` **shall** carry 4-locale (ko/en/ja/zh) entries for every new or moved page introduced by the IA rebuild.

### Verification gates

- **REQ-DVR-017** (Build gate): **When** `hugo --minify` runs in `docs-site/`, the build **shall** complete with zero warnings, generate `public/sitemap.xml`, and emit the `v3.0.0-rc4` version string.
- **REQ-DVR-018** (i18n check gate): **When** `scripts/docs-i18n-check.sh` runs, it **shall** pass — 4-locale file count/path parity, non-empty `title` front matter, H1 present per page, and MoAI-ADK glossary terms preserved untranslated.
- **REQ-DVR-019** (Deploy verify): The §17.6 build/deploy checklist **shall** pass — Mermaid TD diagrams render, the language switcher navigates locale A → locale B, the Vercel preview serves all 4 locale home pages, and `robots.txt` / `llms.txt` domains match `adk.mo.ai.kr`.

### Research planning (plan-phase deferral)

- **REQ-DVR-020** (Research capture): The M0 research task **shall** capture the CC-latest research PLAN (target URLs, what to fetch, per-page mapping) in `research.md`. The actual Claude Code research **shall** be performed in run-phase (M0/M2), NOT at plan-phase.

### Non-functional constraints (from §17 doctrine)

- **NFR-DVR-001** (URL): The forbidden URLs (`docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr`) **shall not** appear anywhere in `docs-site/`; only `adk.mo.ai.kr` is canonical.
- **NFR-DVR-002** (Mermaid direction): Mermaid diagrams **shall** use `TD`/`TB` direction only; `LR` is prohibited.
- **NFR-DVR-003** (Emphasis spacing): Emphasis-paren spacing **shall** be `**term** (English)`, never `**term(English)**`.
- **NFR-DVR-004** (No emoji): Body content **shall not** contain decorative emoji.
- **NFR-DVR-005** (Glossary): MoAI-ADK glossary terms (e.g. "MoAI-ADK") **shall** remain untranslated across all locales.
- **NFR-DVR-006** (Vercel binding): The Vercel binding (project ID `prj_EZaVdfE3gJeXVbizafBEECpniINP`, root `docs-site/`, framework preset Hugo) **shall** remain unchanged.
- **NFR-DVR-007** (No version snapshot): The rebuild **shall** update the `latest` content only; the §17.4 major version snapshot **shall not** be created (deferred to GA).

## Exclusions

Content that is explicitly NOT part of this SPEC. Items below route to their correct home or are deferred.

### Out of Scope — Theme and frontend

- No change to the `hugo-geekdoc` theme, `static/*.css`, `layouts/`, partials, shortcodes, or any visual/layout element. This is a content-only rebuild (REQ-DVR-014). Any design change belongs to a separate design-system SPEC.

### Out of Scope — Version snapshot

- No `content/{locale}/v3/` frozen snapshot (§17.4). Only `latest` is updated. The major-version snapshot is deferred to the GA release and is owned by the release process (`manager-git` + `scripts/docs-release-snapshot.go`).

### Out of Scope — Claude Code research execution at plan-phase

- The actual fetching of latest Claude Code documentation is NOT performed during plan-phase. Plan-phase only produces the research PLAN in `research.md` (REQ-DVR-020); execution happens in run-phase.

### Out of Scope — Codebase and CLI behavior

- No change to `internal/`, `cmd/`, or `pkg/` Go code, CLI behavior, command implementations, or the template source under `internal/template/templates/`. The rebuild documents existing behavior; it does not modify it. (README.md/README.ko.md at repo root ARE in scope per REQ-DVR-013.)

### Out of Scope — Vercel and infrastructure

- No change to the Vercel project binding, domain configuration, `vercel.json` (beyond domain-string correctness under NFR-DVR-001), or CI pipeline definitions.

### Out of Scope — Book landing page

- The external `/book` landing page (a separate external property linked from the menu) is not rewritten by this SPEC.

## Dependencies and References

- SPEC-SUBCOMMAND-RETIRE-001 — source of retired subcommand list (coverage/e2e/design/brain/security).
- SPEC-V3R6-LIFECYCLE-REDESIGN-001 — 3-phase lifecycle authority.
- SPEC-V3R6-HARNESS-V4-001 — Harness v4 Builder authority (for REQ-DVR-010).
- SPEC-V3R6-AGENT-TEAM-REBUILD-001 — 8-agent catalog + archived expert-* rejection (for REQ-DVR-003/011).
- SPEC-V3R6-DOCS-V3-README-001 — sibling naming-family SPEC; convention anchor for this ID.
- `.moai/docs/docs-site-i18n-rules.md` §17 — 4-locale parity, URL blacklist, Mermaid TD, build/deploy checklist, Vercel binding (SSOT for NFRs).

## Acceptance Criteria

See `acceptance.md` for per-requirement testable/verifiable acceptance criteria and Given-When-Then scenarios.
