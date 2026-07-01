# Design — SPEC-V3R6-DOCS-V3-REBUILD-001

Design blueprint for the docs-site v3.0 full rebuild. Covers (A) the new IA blueprint, (B) the content model, (C) the validate-then-rewrite strategy, and (D) the 4-locale propagation design.

> Design intent, not implementation. Concrete page prose is produced in run-phase. This document fixes the structure the rewrite fills in.

## §A New IA Blueprint

### A.1 Current-state IA (13 content sections + menu)

Current per-locale content sections (95 files/locale): `getting-started` (10), `core-concepts` (7), `claude-code` (28, CC mirror), `workflow-commands` (6), `utility-commands` (7), `quality-commands` (4), `db` (5), `multi-llm` (3), `guides` (3), `worktree` (4), `advanced` (15), `cost-optimization` (1), `contributing` (1), + root `_index.md`.

Observed IA drift: `data/menu/main.yaml` lists `cost-optimization` nowhere and carries an empty `contributing` sub-tree; the menu and content have diverged.

### A.2 Target IA (menu ↔ content reconciliation)

The rebuild adopts a menu that is a superset-consistent mirror of the content tree. Design decision per section:

| Section | Target IA decision |
|---------|--------------------|
| Getting Started | Keep. Add `CLI Reference` coverage of `moai inventory` (REQ-DVR-009). |
| Core Concepts | Keep (validate-then-rewrite). |
| Claude Code Guide (CC mirror) | Keep 4-cluster structure (foundations / context-memory / extensibility / agentic); refresh content (REQ-DVR-008). |
| Workflow Commands | Keep; reconcile `moai-harness.md` (REQ-DVR-012). |
| Utility Commands | Keep. |
| Quality Commands | Keep. |
| Database Schema Management | Keep. |
| Multi-LLM | Keep; ensure `/effort ultracode` cross-linked (new page lives under a home TBD in A.3). |
| Guides | Keep. |
| Git Worktree | Keep. |
| Advanced | Keep; rewrite `builder-agents.md` (REQ-DVR-010), fix `agent-guide.md` (REQ-DVR-011); host new Harness v4 Builder + Decision Memory pages (A.3). |
| Cost Optimization | **Reconcile** — surface in the menu OR fold into `multi-llm`. Decision: surface as a menu entry (least-surprise; content already exists standalone). |
| Contributing | Populate the empty `contributing` sub OR remove from menu. Decision: keep 1 content page; remove the empty sub-tree stub. |

### A.3 New-page placement (4 topics × 4 locales = 16 files)

| New topic | Section home | Proposed path (per locale) |
|-----------|--------------|----------------------------|
| Decision Memory (`moai preference`) | Advanced | `advanced/decision-memory.md` |
| `moai inventory` command | Getting Started (CLI area) | `getting-started/inventory.md` (+ CLI-reference cross-link) |
| Harness v4 Builder | Advanced | `advanced/harness-v4-builder.md` |
| `/effort ultracode` dynamic workflows | Advanced (with Multi-LLM + Claude Code cross-links) | `advanced/ultracode-workflows.md` |

> Final path/slug naming is confirmed at M0.4 against the geekdoc `weight`-based sidebar ordering. The acceptance grep (`AC-DVR-009a`) keys on the topic tokens (`preference`/`inventory`/`harness-v4`/`ultracode`) so path choices remain flexible within these tokens.

### A.4 Sidebar generation

Per `hugo.toml`, the sidebar is content-tree `weight`-driven (`geekdocMenuBundle = false`) AND a curated `data/menu/main.yaml` exists. The rebuild keeps both consistent: new pages get a `weight` in front matter AND a `main.yaml` entry (REQ-DVR-016). No theme change (REQ-DVR-014).

## §B Content Model

### B.1 Page front-matter contract

Every page carries the front matter the i18n check enforces (§17.3):

```yaml
---
title: <localized, non-empty>
weight: <int, controls sidebar order>
draft: false
# optional: description, images, translation_status
---
```

Plus: exactly one H1 per page; MoAI-ADK glossary terms untranslated (NFR-DVR-005); emphasis-paren spacing (NFR-DVR-003); Mermaid TD only (NFR-DVR-002); no decorative emoji (NFR-DVR-004).

### B.2 Fact-binding model (single-source numbers)

All version/count facts resolve to the M0 fact-sheet (research.md §1), not to prose memory. The version renders via the `{{< version >}}` shortcode bound to `hugo.toml params.version` (SSOT) — pages must not hardcode `v3.0.0-rc4` inline where the shortcode applies (avoids future version-drift and the double-v hazard EC-2). Counts (13 commands / 8 agents / 27 skills) are written as literals but every occurrence is cross-checked at M4 by grep-based ACs.

### B.3 CC-mirror traceability model

Each `claude-code/` page carries a hidden HTML-comment provenance marker (or a "based on" front-matter field) naming the canonical Claude Code doc surface it mirrors (research.md §3 mapping). This satisfies AC-DVR-008a traceability and enforces verification-claim-integrity (no fabricated feature claims).

## §C Validate-then-Rewrite Strategy

### C.1 Two-track rewrite

- **Track A (validate-then-rewrite)** — sections flagged accurate (core-concepts, workflow-commands, utility-commands, quality-commands, multi-llm): the existing content is loaded as the draft baseline, each fact is validated against the M0 fact-sheet, then the prose is rewritten for v3.0 clarity. NOT blank-page (AP-DVR-001). A per-page validation log line records "validated against fact-sheet vN" (AC-DVR-006b).
- **Track B (full rewrite)** — drift pages (builder-agents, agent-guide, moai-harness), the CC mirror, and the 4 new pages: authored fresh against ground truth + research.

### C.2 Per-locale independent validation

EC-1: accuracy in ko does NOT imply accuracy in en/ja/zh. Each locale is validated independently before its rewrite. The propagation chain (§D) governs translation, but the fact-validation step runs per locale.

### C.3 Deprecation handling

EC-3: a feature deprecated between rc2 and rc4 is REMOVED, not merely refreshed. The CC research delta note (research.md §3) explicitly flags removals so the rewrite deletes stale pages/sections rather than paraphrasing them forward.

## §D 4-Locale Propagation Design

### D.1 Chain (per §17.3)

`ko` (canonical authoring) → `en` (fact + language parity) → `ja` / `zh` (parallel). ko-only merge is prohibited; every content change lands in all 4 locales within the same milestone (REQ-DVR-015).

### D.2 Milestone-to-locale mapping

- M1 (ko core) + M2 (ko CC) author ko first.
- M3 propagates M1+M2 to en, then ja/zh in parallel, then rebuilds `data/menu/main.yaml` 4-locale entries.
- M4 verifies parity mechanically (`scripts/docs-i18n-check.sh`).

### D.3 Parity invariants

1. Identical file count per locale (AC-DVR-015a).
2. Identical relative paths per locale (AC-DVR-015b).
3. Non-empty localized `title` per page.
4. Menu `ref` resolves in every locale (AC-DVR-016a).
5. Glossary terms untranslated (NFR-DVR-005).

### D.4 Humanize pass (translation quality)

en/ja/zh receive an idiomatic humanize pass (not literal word-by-word) consistent with prior docs-v3 practice; CJK locales use native idiom, never invented transliteration. Version/count facts are copied verbatim (numbers do not translate).

## §E Design Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| CC research incomplete/outdated at M2 | M0.3 research is a hard precondition (plan.md §C); M2 blocked until the delta note exists. |
| Double-v version render (EC-2) | Bind to `{{< version >}}` shortcode; M4 double-v grep (AC-DVR-017b). |
| Skill-count drift reintroduced | Single M0.2 reconciliation + M4 grep across content + README (AC-DVR-004b). |
| Menu/content divergence recurs | IA reconciliation table (A.2) + AC-DVR-007a/-016a mechanical check. |
| Fabricated CC feature claims | Traceability marker (B.3) + verification-claim-integrity binding. |
| 4-locale volume causes ko-only drift | Same-milestone parity rule (D.1) + mechanical i18n check gate (M4). |

## §F Cross-References

- `spec.md` — requirements. `plan.md` — milestones + anti-patterns. `acceptance.md` — ACs. `research.md` — fact-sheet + CC research PLAN.
- `.moai/docs/docs-site-i18n-rules.md` §17 — parity/URL/Mermaid/build doctrine.
- `docs-site/hugo.toml` — version SSOT, theme, i18n config.
- `docs-site/data/menu/main.yaml` — IA navigation source.
