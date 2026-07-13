---
id: SPEC-WEBCONF-SIMPLIFY-001
title: "moai web Configuration UI Simplification + Sub-Agent 4-Color Tier Redesign — Design"
version: "0.3.0"
status: completed
created: 2026-07-13
updated: 2026-07-14
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "internal/web, internal/settings, internal/template/templates"
lifecycle: spec-anchored
tags: "web-ui, config, sub-agent, tier-model, template-defaults"
tier: L
---

## HISTORY

| Version | Date       | Change                                                       |
|---------|------------|--------------------------------------------------------------|
| 0.2.0   | 2026-07-13 | Initial design authoring (Tier L artifact, iter-1 audit-fix). |
| 0.3.0   | 2026-07-13 | In-progress amendment. Add §H description-source mechanism (REQ-WC-015, Refinement 2): option (a) hybrid — `FieldDef.Description` i18n key + `i18n.js` owns locale text. Former §H Cross-References → §I. |

---

## §A. Purpose

This design formalizes the **tier data-model** for the sub-agent 4-color concept (Decision 3) and the **§E baking approach** (Decision 2). It exists to make the single most change-likely decision in the SPEC — the tier mechanism — reviewable in one place, separate from the requirements (spec.md) and the milestone scheduling (plan.md).

The central design choice, resolved by the user at iter-1, is **Option A: a name-keyed lookup table** (display-only). The rejected alternative — effort-derivation — is documented in §F so the reasoning is auditable.

---

## §B. The Distinction: tier (display-only, chosen) vs effort (live frontmatter, spawn-time)

| Concept        | Source                            | Mutability                                | Used at                         |
|----------------|-----------------------------------|-------------------------------------------|---------------------------------|
| **Tier color** | Name-keyed lookup table (chosen)  | Static map; changes only via SPEC edit    | Display-time (badge render)     |
| **Effort**     | Each agent's `effort:` frontmatter | Per-agent, via `agentfm.Patch` on override | Spawn-time (runtime effort)     |
| **Model**      | Each agent's `model:` frontmatter  | Per-agent, via `agentfm.Patch` on override | Spawn-time (runtime model)      |

The badge shows the **tier color**. The selectors show and write **effort / model**. These are two DIFFERENT data sources by design:

- The tier answers "what reasoning role does this agent play?" — a stable classification chosen once.
- The effort answers "what effort did the user last set for this agent?" — a mutable spawn-time value.

Conflating them (the rejected effort-derivation alternative) causes three problems documented in §F.

---

## §C. The 20-agent name→tier table (chosen mapping)

Hand-curated by reasoning role. This is the source of truth for badge color. It is a Go `map[string]Tier` constant (or equivalent), NOT derived from any frontmatter.

| Agent                                | Tier | Reasoning-role basis                                                                |
|--------------------------------------|------|-------------------------------------------------------------------------------------|
| `manager-spec`                       | 🔴   | Plan-phase SPEC authoring; deep requirement + GEARS reasoning.                      |
| `plan-auditor`                       | 🔴   | Independent audit; bias-prevention judgment; GEARS compliance evaluation.           |
| `super-advisor`                      | 🔴   | E1-E4 high-reasoning consultation; non-binding prescriptions require deepest reasoning. |
| `sync-auditor`                       | 🔴   | Independent skeptical 4-dimension scoring; harmonic-mean verdict.                   |
| `manager-develop`                    | 🟠   | Run-phase DDD/TDD implementation; heavy reasoning bounded by the SPEC.              |
| `manager-design`                     | 🟠   | D1-D5 design pipeline; creative + structural synthesis.                             |
| `builder-harness`                    | 🟠   | Dynamic harness / specialist generation; structural reasoning.                      |
| `e2e-specialist`                     | 🟠   | Web/mobile/desktop journey scripting; cross-platform complexity.                    |
| `manager-docs`                       | 🔵   | Sync-phase documentation; template-driven, moderate reasoning.                      |
| `manager-git`                        | 🔵   | PR creation per Tier routing; procedural git ops.                                   |
| `quality-specialist`                 | 🔵   | TRUST 5 validation; checklist-driven quality checks.                               |
| `cli-template-specialist`            | 🔵   | Template editing; pattern-application.                                              |
| `workflow-specialist`                | 🔵   | Workflow pattern application; procedural.                                           |
| `hns-github-specialist`              | 🩵   | Narrow GitHub issue/PR scope via `gh` CLI.                                          |
| `hns-oss-docs-content-author-specialist` | 🩵 | Narrow README/docs content authoring scope.                                         |
| `hns-oss-docs-locale-translator-specialist` | 🩵 | Narrow locale-translation scope.                                               |
| `hns-oss-docs-structure-curator-specialist` | 🩵 | Narrow docs-site structure/navigation scope.                                   |
| `hns-release-specialist`             | 🩵   | Narrow release-publishing scope.                                                    |
| `hns-release-update-specialist`      | 🩵   | Narrow CC upstream-change tracking scope.                                           |
| `hook-ci-specialist`                 | 🩵   | Narrow CI/hooks scope.                                                              |

**Distribution**: 🔴×4 · 🟠×4 · 🔵×5 · 🩵×7 = 20.

> **Why this is NOT derived from effort**: 10 of these agents have a current `effort` frontmatter value that DIFFERS from the tier's suggested effort (e.g. `manager-develop` carries `effort: xhigh` today but its tier is 🟠 which suggests `high`). 3 agents have NO effort frontmatter at all (`hns-oss-docs-*`). Under effort-derivation, the badge for these 13 agents would either be "wrong" (showing a color that mismatches the agent's role) or absent. Under the name-keyed table, the badge is always correct-by-construction because it is chosen, not derived. See research.md §C for the full effort landscape.

---

## §D. Tier → suggested-(model, effort) table

When the user clicks a tier color on an agent row, the UI auto-fills the model and effort selectors with these suggested values. The suggestion is applied only on explicit user action (click + confirm). Applying it writes `effort` and/or `model` via the existing `agentfm.Patch` — it does NOT write a `tier:` field, and it does NOT change the name→tier table.

| Tier color | Suggested `model` | Suggested `effort` |
|------------|-------------------|--------------------|
| 🔴          | `opus`            | `xhigh`            |
| 🟠          | `opus`            | `high`             |
| 🔵          | `sonnet`          | `medium`           |
| 🩵          | `haiku`           | `low`              |

`max` effort and `inherit` model are NOT tiers — they are override sentinels. When the user manually sets `effort: max` or `model: inherit`, the badge renders a neutral "custom" state (not a 5th color); the name-table tier is still known internally but the badge surfaces the override.

---

## §E. UI Architecture: `agentFMRow` redesign

The redesigned `internal/web/fieldsets.templ:385` `agentFMRow` renders three elements per agent:

```
┌──────────────────────────────────────────────────────────────────┐
│ [🔴 badge]  manager-spec          model: [opus ▾]   effort: [xhigh ▾] │
└──────────────────────────────────────────────────────────────────┘
```

1. **Color tier badge** — from the name-keyed lookup table (§C) via the new accessor. Static per agent name; does not change when the user edits effort.
2. **Model `<select>`** — populated from `V4ModelValues` (`{inherit, haiku, sonnet, opus}`). Reflects the agent's current `model:` frontmatter.
3. **Effort `<select>`** — populated from `V4EffortValues` (`{low, medium, high, xhigh, max}`). Reflects the agent's current `effort:` frontmatter.

**Tier-click interaction**: clicking the badge (or a "apply tier defaults" affordance) opens a 4-color picker; selecting a color auto-fills the model + effort selectors with that tier's suggested pair (§D). The user can then accept (writes `effort`/`model` via `agentfm.Patch`) or further override either selector independently.

**No new frontmatter key**: the persist path uses ONLY `agentfm.Patch` writing `effort` and/or `model`. The `tier:` key is NEVER written to any agent file (AC-WC-017). The name→tier table lives in Go source, not in agent frontmatter.

**Accessor surface**: add `TierForAgent(name string) Tier` and `TierSuggestedModelEffort(t Tier) (model, effort string)` accessors in `internal/settings/schema_sections.go` (or a sibling), backed by the table constants in `internal/harness/v4manifest/schema.go` (or a sibling). The `fieldsetAgentFrontmatter` (`fieldsets.templ:361`) threads these through to `agentFMRow`.

---

## §F. The §E baking approach (Decision 2)

Each removed-tab section's keys-of-interest (spec.md §E.1–§E.10) are baked into `internal/template/templates/.moai/config/sections/<section>.yaml`. The approach is **cat-verbatim + overwrite-keys-of-interest + preserve-all-other-keys**:

1. `cat` the live template file (run-phase preflight step 1).
2. Identify the keys-of-interest listed in the spec.md §E.N block.
3. Overwrite ONLY those keys with the baked values.
4. Preserve ALL other keys verbatim — including the ones the spec.md §E.N block does not mention (e.g. `harness.plan_audit_global`, `harness.levels.*`, `harness.model_upgrade_review.checklist[].{question,affects}`, `llm.glm.base_url`, `llm.claude_models`, `llm.performance_tier`, `llm.plan_type`, `workflow.workflow_agents`, `workflow.model_routing`, `workflow.model_routing_profiles`).
5. Where a baked value intentionally differs from the current template value, the spec.md §D.Δ table enumerates the delta — apply it exactly, do NOT silently "fix" it back.

**Why cat-verbatim instead of full-file rewrite**: the live template files are large (`harness.yaml` ≈ 8165B, `workflow.yaml` similar) and carry doctrinal comments, `@MX` annotations, and nested structures that must be preserved byte-for-byte. A full-file rewrite risks dropping keys or mangling comments. The cat-verbatim + overwrite-keys-of-interest approach is the minimum-blast-radius edit.

---

## §G. Rejected Alternative — effort-derivation (why Option A won)

The rejected alternative was: derive the tier color from each agent's live `effort` frontmatter via an `effort → color` map (`xhigh→🔴, high→🟠, medium→🔵, low→🩵`). This was the iter-0 / v0.1.0 design.

Three fatal problems (confirmed by the iter-1 auditor's effort-landscape survey — research.md §C):

1. **20-agent-file rewrite required**. To make the badge match the agent's reasoning role, every agent's `effort` frontmatter would need to be rewritten to align with the role-based tier. That is 20 file edits with behavioral side-effects (the effort value is consumed at spawn-time), violating the "display-only" intent and Constraint C-7.
2. **Circular dependency for absent-effort agents**. 3 agents (`hns-oss-docs-content-author-specialist`, `hns-oss-docs-locale-translator-specialist`, `hns-oss-docs-structure-curator-specialist`) have NO `effort` frontmatter. Under effort-derivation, their badge would be absent or require a special-case fallback — a circular dependency (the tier display depends on effort, but effort is absent). The v0.1.0 EC-1 hand-waved this as "fall back to the catalog-default tier" — which is just the name-keyed table in disguise.
3. **Badge flapping**. The effort value is mutable (the user can override it per-agent). Under effort-derivation, the badge color would change whenever the user changes effort — even though the agent's reasoning role has not changed. This is confusing: the badge should communicate a stable property (role), not a mutable one (current effort setting).

Option A (name-keyed lookup table) eliminates all three: no file rewrite, no circular dependency (the 3 absent-effort agents get 🩵 from the table directly), no flapping (the table is static). The cost is one Go source map that must be maintained when the catalog changes — a small, explicit, reviewable cost.

---

## §H. Description-source mechanism (REQ-WC-015, Refinement 2)

REQ-WC-015 requires every selectable field/option across the 6 surviving tabs to display a localized description. This section decides the data mechanism, the render location, and surfaces the i18n burden.

### §H.1 Chosen mechanism — option (a) hybrid: `FieldDef.Description` i18n key + `i18n.js` owns the locale text

Add a `Description string` field to `FieldDef` (`internal/settings/schema.go`). The field carries an **i18n KEY** (NOT inline text). The `internal/web/assets/i18n.js` dictionary owns the resolved per-locale text. Empty `Description` = the field has no description (render nothing).

**Key convention**:
- Field-level: `fieldDesc.<sectionID>.<fieldID>`
- Per-option: `fieldDesc.<sectionID>.<fieldID>.option.<optionValue>`

### §H.2 Why option (a) over (b) and (c)

| Option | Verdict | Reason |
|--------|---------|--------|
| (a) `FieldDef.Description` = i18n key + `i18n.js` text | **CHOSEN** | Gives a testable anchor: a unit test enumerates `FieldDef` entries for the 6 surviving tabs, asserts each has a non-empty `Description` key, asserts that key exists in all 4 locale dicts. Schema declares presence + key; i18n owns text. Forward-compat (a field without a description has empty `Description`, no render). |
| (b) Pure i18n-convention (`<section>.<field>` keys, no schema change) | Rejected | Cannot enumerate "which fields exist" without re-reading the schema, making the 4-locale parity test brittle (the test would have to mirror the field list in two places — schema + test fixture — inviting drift). |
| (c) Inline text in `FieldDef.Description` | Rejected | Requires Go recompilation for any copy fix; does not natively support 4-locale parity (one inline string cannot serve 4 locales without a re-resolution layer, which collapses back to option (a)). |

### §H.3 Render location — below the field label, as muted helper text

The description renders **below the field label**, styled via a new `.field-description` CSS class in `internal/web/assets/console.css` (muted color, smaller font). This is cleaner than:

- **Tooltip on hover** — hidden on touch/mobile, requires a hover affordance, delays the information.
- **Info icon (ⓘ) with click** — extra click to read; adds a JS interaction for static text.

For per-option descriptions within a `<select>`, each `<option>` carries a native `title` attribute (browser-native tooltip on hover, zero custom JS). The field-level description below the label explains the field; the per-option `title` gives on-hover detail per choice. This split matches the existing `fieldsets.templ` + `console.css` patterns (label + control + helper-text rows) without introducing a new widget.

### §H.4 i18n burden (the heaviest part of M6)

4 locales × (~20–30 selectable fields across the 6 surviving tabs) × (1 field-level description + 2–5 per-option descriptions where applicable). Estimate: **~200–400 strings per locale × 4 locales = ~800–1600 strings total**.

This is the single largest authoring burden in the SPEC. M6 calls it out explicitly. Recommended run-phase approach: draft the `en` entries first (one pass over the schema's surviving-tab `FieldDef` set), then translate to `ko`/`ja`/`zh` (ideally with the `moai-domain-humanize` specialist for naturalization). The `FieldDef.Description` key convention keeps the string set grep-able and CI-validatable (AC-WC-022 + AC-WC-023).

---

## §I. Cross-References

- `spec.md` §B REQ-WC-006..009 (tier) + REQ-WC-015 (descriptions) + REQ-WC-016 (git_strategy core surface); §C C-7 (no agent-file rewrite for display) + C-8 (closed-set validation) + C-9 (description i18n 4-locale parity); §D.Δ (deliberate default-changes); §E (baked keys-of-interest).
- `plan.md` §F M1 (data-model decision + 20-agent table) / M4 (surviving-tab surfaces + field descriptions) / M5 (UI redesign + tier option descriptions + stale-comment sweep) / M6 (i18n description strings — heaviest).
- `acceptance.md` §D AC-WC-005..008 / AC-WC-016..018 / AC-WC-021 (tier) + AC-WC-022..024 (descriptions + git_strategy surface).
- `research.md` §C — the 20-agent effort landscape that motivated Option A.
- `internal/harness/v4manifest/schema.go:43-74` — closed-set definitions the table respects.
- `internal/settings/agentfm/agentfm.go` — the existing `Patch` mechanism reused for per-agent override (NO new frontmatter key).
- `internal/settings/schema.go` `FieldDef` — gains the `Description string` field (option (a), §H.1).
