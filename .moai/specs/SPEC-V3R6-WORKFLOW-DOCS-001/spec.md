---
id: SPEC-V3R6-WORKFLOW-DOCS-001
title: "Document the SPEC lifecycle and the Kanban/Factory operating models on docs-site and README"
version: "0.1.0"
status: draft
created: 2026-08-25
updated: 2026-08-25
author: manager-spec
priority: P1
phase: "v3.1.3 target"
module: docs-site
lifecycle: spec-anchored
tags: "docs-site, i18n, kanban, factory-mode, spec-lifecycle, readme"
tier: M
---

## HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-08-25 | manager-spec | Initial creation — plan-phase artifacts for card t273 (Class C, Tier M). Scope fixed by `.moai/reports/t273/gap-map.md` GAP-1..GAP-4; that gap map's Out-of-Scope list is binding on this SPEC. |
| 2026-08-25 | manager-spec | Plan-audit iter-1 revision (review-1 D1-D10): GREEN anchors strengthened (AC-001/007/009/011); run-entry pre-flight decoupled from nav approval (D2); gap-map committed to branch (D3); REQ-WFD-001 Class A canon completion (D4); AC-009 grep `-E` fix (D5); E7 count 33 (D6); REQ-WFD-004 relabel (D7); M1 qualifier move (D8); denominator footnote (D9); REQ-WFD-007 icon wording (D10). Team-lead nav approval recorded 2026-08-25 with three binding conditions (§C.3). |

## §A Context and Problem

MoAI-ADK's public documentation (docs-site `adk.mo.ai.kr`, 4 locales ko/en/ja/zh; README 4-file set, ko canonical) does not cover three operating-model topics that exist only in internal canonical rules:

1. **Kanban card classes A/B/C** (gap-map GAP-1) — defined in `.claude/rules/moai/workflow/kanban-dispatch.md` § Card classes, absent from every public surface (measured: 0 matches across 8 target files).
2. **Factory Mode as a dedicated page** (GAP-2) — the factory section inside `advanced/kanban-mode.md` is a discoverability gap given the first-class `-f` launcher token; the operating numbers (per-lane concurrent sub-agent cap 10, `workers.json` slot registry) are grounded in `README.ko.md:80`, `internal/kanban/factory_slots.go:48`, and `advanced/kanban-mode.md:239`.
3. **The plan>run>sync lifecycle as one integrated flow** (GAP-3) — command-level pages (`workflow-commands/moai-{plan,run,sync}.md`) and the SPEC-document page (`core-concepts/spec-based-dev.md`) exist, but no page presents the three phases, their input→output artifacts, the three gates, and the Tier S/M/L budgets as a single flow.

GAP-4 adds the card-class table to the README kanban section (4 locales) and link verification for the two new pages.

Canonical sources the documentation must not drift from:

- `.claude/rules/moai/workflow/spec-workflow.md` — § Phase Overview, § SPEC Complexity Tier, § Plan/Run/Sync Phase, § Phase Transitions, § Phase 1 Plan Audit Gate
- `.claude/rules/moai/workflow/kanban-dispatch.md` — § The board, § Card classes, § The dispatch cycle, § Factory Mode — the card travels whole
- `.claude/rules/moai/workflow/worktree-integration.md` — § Terminology Glossary (L1/L2 only; deep worktree internals out of scope)

## §B Requirements (GEARS)

REQ/AC budget: Tier M — 12 requirements, 11 acceptance criteria (ceilings 16/16 respected independently).

- **REQ-WFD-001** (Ubiquitous): The docs-site kanban-mode page shall present the three card classes — A (direct close: one file, one line, no design judgement — CI catches the regression; admitted on checked evidence: measured one-file diff plus CI green on the head that will merge; plan column skipped), B (defect, cause unknown: plan column skipped but the sync review gate NOT skipped, with cause-establishing evidence written to the card's progress record), and C (design change: all three working columns) — in all four locales, using the normative heading tokens of §C.4.
- **REQ-WFD-002** (Ubiquitous): The four README files shall carry a compact card-class table in the kanban section, authored in `README.ko.md` (canonical) and derived to `README.md` / `README.ja.md` / `README.zh.md`.
- **REQ-WFD-003** (Ubiquitous): The docs-site shall provide a dedicated Factory Mode page at `advanced/factory-mode.md` in all four locales, covering: the lead + lane-1..N structure; whole-card routing contrasted with Kanban's column-to-column movement; the lane-internal plan>run>sync stages with per-stage sub-agent spawning; the per-lane concurrent sub-agent cap of 10 (launcher-injected `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`); staggered lane activation; and `workers.json` slot ownership at `.moai/state/factory/workers.json`.
- **REQ-WFD-004** (State-driven): While the dedicated Factory Mode page exists, the kanban-mode page shall retain a Factory Mode summary that links to `advanced/factory-mode.md`, in all four locales.
- **REQ-WFD-005** (Ubiquitous): The docs-site core-concepts section shall provide a SPEC lifecycle page at `core-concepts/spec-lifecycle.md` in all four locales, covering: the 3-phase table (command / owning agent / purpose); per-phase input→output artifacts; the three gates — Implementation Kickoff Approval (the plan→run human gate, mandatory and score-independent), plan-audit (independent audit, per-tier PASS thresholds 0.75 / 0.80 / 0.85), and sync-auditor (4-dimension quality scoring); the Tier S/M/L artifact sets (2 / 3 / 5 files) and REQ/AC ceilings (8 / 16 / 25); a one-paragraph Route A/B summary; and the /clear strategy.
- **REQ-WFD-006** (Ubiquitous): The spec-lifecycle page and `core-concepts/spec-based-dev.md` shall cross-link each other with an explicit division of labor — spec-based-dev.md answers "what is a SPEC document"; spec-lifecycle.md answers "how does the lifecycle flow" — minimizing duplication.
- **REQ-WFD-007** (Capability gate): Where the team lead has approved the navigation plan (approval granted 2026-08-25; binding conditions in §C.3), the two new pages shall be registered in the per-locale `_meta.yaml` files (4 locales × 2 pages) and in `docs-site/data/menu/main.yaml` (leaf entries: 4-locale name map + `ref`). No menu icon is introduced; if an icon is ever attached, it must reuse the existing SVG cases (`school` or `flash_on`). Page authoring itself is not gated on this approval.
- **REQ-WFD-008** (Ubiquitous): The canonical-locale chain shall govern all content changes — docs-site authored in ko and derived ko → en → ja/zh; README authored in `README.ko.md` and derived to en/ja/zh — with every canonical change landing in all 4 locales in the same PR.
- **REQ-WFD-009** (Unwanted): The new and modified pages shall not contain Mermaid LR/RL directions, body-text emoji (the `{{</* icon */>}}` shortcode is the mechanism instead), forbidden URL domains (`docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr` — only `adk.mo.ai.kr` is valid), or emphasis markers enclosing parenthetical text.
- **REQ-WFD-010** (Ubiquitous): The 4-locale page inventory shall remain in exact parity after the change — every locale moves from the measured baseline of 150 pages to 152 — and the change shall introduce no NEW per-page section-count divergence against the `.locale-parity-baseline` ratchet.
- **REQ-WFD-011** (Ubiquitous): The documented facts shall match the canonical sources verbatim where they are protocol values: the phase/agent table (`/moai plan` → manager-spec, `/moai run` → manager-develop, `/moai sync` → manager-docs), the Tier budgets (artifact sets 2/3/5; thresholds 0.75/0.80/0.85; ceilings 8/16/25), the card-class semantics, and the factory operating numbers (cap 10; `.moai/state/factory/workers.json`). Any divergence from canon is a defect.
- **REQ-WFD-012** (Event-driven): When run-phase completes, the full `hns-oss-docs-verify` exit gate shall pass — warning-free hugo build, sitemap existence, URL-blacklist zero matches, Mermaid direction zero matches, 4-locale file-existence parity, README 4-file H2 heading parity (baseline 12/12/12/12), and a clean body-emoji scan.

## §C Constraints

### §C.1 i18n HARD obligations (SSOT: `.moai/docs/docs-site-i18n-rules.md`; digest: Skill `hns-oss-docs-i18n-rules`)

- ko canonical for BOTH surfaces (docs-site content and README); derivation chains ko → en → ja/zh (docs-site) and ko → en/ja/zh (README); 4-locale same-PR obligation.
- Mermaid TD-only (`flowchart TD` / `graph TB`); translation preserves diagram direction verbatim.
- No body-text emoji; icon shortcodes `{{</* icon <name> [variant] */>}}`; typographic symbols (→ ← ↓ ✓ ✗) preserved.
- Emphasis-marker spacing: `**단어** (Word)` — parenthetical outside the markers.
- URL whitelist: only `adk.mo.ai.kr`.
- Version SSOT: `docs-site/hugo.toml` `params.version` (measured v3.1.2 on this tree); no divergent hardcoded version displays.
- New pages are not moves — no `vercel.json` redirect entries required. Vercel binding immutable; specialists never commit/push (publishing stays orchestrator/human-gated).

### §C.2 Canonical-fidelity constraint

All facts documented under REQ-WFD-011 trace to the three canonical rule files listed in §A. Where a canonical value also appears in existing public pages (e.g. `README.ko.md:80` cap 10, `advanced/kanban-mode.md:239`), the new pages restate the same value — no second opinion.

### §C.3 Navigation gating (approval granted 2026-08-25)

Nav edits (`_meta.yaml` × 4 locales, `data/menu/main.yaml`) are structure-curator domain and required team-lead approval before run-phase nav edits. **The team lead approved the navigation plan on 2026-08-25**, with three binding conditions that become M4 completion criteria (plan.md §F M4): (1) the new items' order in all 4 locales' `_meta.yaml` must be identical across locales (factory-mode adjacent to kanban-mode); (2) `data/menu/main.yaml` name maps must carry all 4 keys ko/en/ja/zh — a missing key passes the build but breaks rendering; (3) the exit gate is the FULL `hns-oss-docs-verify` recipe observed passing (warning-free hugo build, sitemap, URL blacklist, Mermaid direction, 4-locale file existence, section-count parity). Plan milestone M0 records the approval and reports completion; milestone M4 executes nav edits under the three conditions. Page authoring (M1–M3) never waits on nav approval.

### §C.4 Normative heading/content tokens (AC anchors)

The card-class section heading carries, per locale: ko `카드 클래스`, en `Card Classes`, ja `カードクラス`, zh `卡片类别`. Protocol values stay locale-verbatim: `Class A/B/C`, `Implementation Kickoff`, `0.75` / `0.80` / `0.85`, `workers.json`, `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`, `sync-auditor`, `factory-mode`, `spec-lifecycle`, and the four sync-auditor dimension names `Functionality` / `Security` / `Craft` / `Consistency` (the 4-dimension scoring semantics; AC anchors rely on them).

### §C.5 Touch boundary

This SPEC touches `docs-site/` and the 4 README files only. No files under `internal/`, `cmd/`, `pkg/`, or `internal/template/templates/` are modified.

## §D Acceptance Criteria (summary)

Full binary criteria with RED-NOW evidence and GREEN-PATH per criterion live in `acceptance.md` (11 ACs). Traceability: AC-WFD-001..002 → REQ-001/002; AC-WFD-003..005 → REQ-003/004; AC-WFD-006..008, AC-WFD-011 → REQ-005/006/011; AC-WFD-009 → REQ-007; AC-WFD-010 → REQ-010. REQ-008/009/012 are process obligations and regression gates verified by the plan §E verify recipe (they already pass on the current tree and are therefore disqualified as RED-based ACs — see `acceptance.md` §D.6).

## Out of Scope

Binding per `.moai/reports/t273/gap-map.md` § 명시적 비범위:

### Out of Scope — wholesale rewrite of existing pages
- No full rewrite of existing 4-locale pages; the 524-page locale parity set is preserved (t87 drop rationale). Denominator note: 150 `.md` files per locale (incl. 19 `_index.md`) = 131 content pages per locale; 131 × 4 = **524 content pages**. The 150 × 4 = 600 figure in acceptance.md E5 counts all `.md` files — different denominators, both correct.

### Out of Scope — Origin-Trail Chain internals
- JSONL / WorktreeNode 13-field / two-phase backfill internals are not re-documented; the existing kanban-mode.md deep section stays as-is.

### Out of Scope — session handoff and context-window documentation
- Separate topic, separate card. Not covered here.

### Out of Scope — multi-llm/kanban-mode.md restructure
- Keeps its role as the operating-procedure document; untouched.

### Out of Scope — CLI code and template changes
- No `internal/`, `cmd/`, `pkg/`, or `internal/template/templates/` changes. Docs + README only.

### Out of Scope — vercel.json redirects
- New pages are additions, not moves; no redirect entries.

## §F Dependencies and Cross-References

- Gap map (scope source, binding): `.moai/reports/t273/gap-map.md`
- Canonical rules: `.claude/rules/moai/workflow/spec-workflow.md`, `kanban-dispatch.md`, `worktree-integration.md`
- i18n SSOT: `.moai/docs/docs-site-i18n-rules.md`; verify recipe: Skill `hns-oss-docs-verify`
- Related SPECs: SPEC-V3R6-DOCS-I18N-PARITY-001, SPEC-V3R6-DOCS-I18N-COMPLETION-001 (locale-parity ratchet lineage) — non-blocking references
- Card: t273 (Kanban Class C, Tier M), branch `WT-workflow-docs`, base `db1362739`
