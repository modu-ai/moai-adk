---
id: SPEC-WEBCONF-SIMPLIFY-001
title: "moai web Configuration UI Simplification + Sub-Agent 4-Color Tier Redesign — Plan"
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

| Version | Date       | Change                                                                                                                                                                  |
|---------|------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-07-13 | Initial plan-phase authoring.                                                                                                                                           |
| 0.2.0   | 2026-07-13 | iter-1 audit-fix. D3 Option A name-keyed tier table (display-only). M1.1/M1.2 rewritten. OQ-1/OQ-2 resolved (all 3 NEEDS CLARIFICATION removed, D1). C-7 rewritten. M5 stale-count-comment sweep (D13). Tier M → L. |
| 0.3.0   | 2026-07-13 | In-progress amendment (M1 merged; M3–M6 unimplemented → zero rework). Refinement 1: M4 git_strategy surface expanded (`mode` → `mode` + `merge_method` + `hooks.pre_push`). Refinement 2: REQ-WC-015 per-option descriptions + C-9; M4/M5 field-description rendering; M6 i18n description strings (heaviest burden). design.md §H description-source mechanism. |

---

## §A. Context

### §A.1 Problem statement

See `spec.md` §A. The `moai web` console tab sprawl (17 tabs) and the absence of any tier signal in the agentfm UI are the two pain points. The four LOCKED user decisions (spec.md §B REQ-WC-001..010) define the target state precisely.

### §A.2 Verified codebase map (cited paths — confirmed to exist; full key inventory in research.md §A)

| Concern                          | Path                                                                                                        | Notes                                                                                                              |
|----------------------------------|-------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------|
| Web handler entry                | `internal/cli/web.go` → `internal/web/server.go`                                                            | Port default 3041.                                                                                                 |
| SSOT schema                      | `internal/settings/schema.go` (`FieldDef`, `allFields()`, `SectionID` constants ~L25-53)                   | The canonical field/section registry. `SectionStatusline`/`SectionQuality`/`SectionGitConvention` ~L26-36 are UNAFFECTED (spec.md §D D12). |
| Extension-section fields         | `internal/settings/schema_sections.go`                                                                      | `FieldDef`s for extension sections; `V4EffortValues` / `V4ModelValues` accessors ~L46-47.                          |
| Tab nav                          | `internal/web/schemaform.go:31-56` `consoleTabs()` + `:71-91` `schemaSectionMetas()`                        | Removing a tab = delete its entry in `consoleTabs()` and reclassify in `schemaSectionMetas()`.                     |
| Persistence routing              | `internal/settings/sectionroute.go` (`RouteTypedSave` / `RouteSeam` / `RouteStatusline` / `RouteExcluded`)  | Removed tabs → reclassify to `RouteExcluded` (config persists, write path removed).                                |
| Save handler                     | `internal/web/handlers.go` (~L270-400 `handleSave`)                                                         | Atomic save contract; validators for removed fields must not fire.                                                 |
| Sub-agent FM (backend)           | `internal/settings/agentfm/agentfm.go` (`List` / `Patch`, frontmatter-only surgery)                         | `agentDirsFor()` ~L49 (stale comment "8 retained agents" → actual 10 moai-custom; D13 fix).                       |
| Sub-agent FM (UI)                | `internal/web/fieldsets.templ:361` `fieldsetAgentFrontmatter` + `:385` `agentFMRow`                         | `:357` stale comment "7 sub-agent" → actual 20 (D13 fix). Redesign target for the color badge + model/effort selectors. |
| Sub-agent FM (parse/validate)    | `internal/web/agentfm.go:79-117`                                                                             | Validation hooks into `V4EffortValues` / `V4ModelValues`.                                                           |
| Closed sets                      | `internal/harness/v4manifest/schema.go:43-74`                                                                | `effort ∈ {low, medium, high, xhigh, max}`; `model ∈ {inherit, haiku, sonnet, opus}`.                              |
| Prompt cache                     | `internal/runtime/cache_control.go` `InjectCacheControl`                                                    | OMITS on GLM — preserve this carve-out (Constraint C-4).                                                           |
| Cache config                     | `internal/config/cache_control.go` + `.moai/config/sections/cache.yaml` (root key `cacheStrategy`)           | Value baked (spec.md §E.6); tab removed.                                                                           |
| Frontend assets                  | `internal/web/assets/{app.js, console.css, i18n.js}`                                                        | 4-locale i18n dict; tier-label strings added here.                                                                 |
| Templ compile                    | `*_templ.go` via `make build`                                                                               | Required after any `fieldsets.templ` edit.                                                                         |
| Template source (defaults SSOT)  | `internal/template/templates/.moai/config/sections/*.yaml`                                                  | Ships to all users via `moai init` / `moai update`. Live key inventory in research.md §B.                          |

### §A.3 Agent catalog (verified 20 agents)

Confirmed by `ls .claude/agents/{moai,harness}/` — exactly 20 `.md` files:

- **moai/ (10)**: builder-harness, e2e-tester, manager-design, manager-develop, manager-docs, manager-git, manager-spec, plan-auditor, super-advisor, sync-auditor.
- **harness/ (10)**: cli-template-specialist, hns-github-specialist, hns-oss-docs-content-author-specialist, hns-oss-docs-locale-translator-specialist, hns-oss-docs-structure-curator-specialist, hns-release-specialist, hns-release-update-specialist, hook-ci-specialist, quality-specialist, workflow-specialist.

`Explore` (Anthropic built-in) is NOT in `.claude/agents/` and is NOT editable via agentfm — excluded from the 20.

---

## §B. Known Issues / Risks

| ID   | Issue                                                                                                                                                          | Mitigation                                                                                                                       |
|------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------|
| KI-1 | Removing a tab without reclassifying its section to `RouteExcluded` leaves a dangling write path → `handleSave` may 404 or write a partial section.            | M3 deletes the `consoleTabs()` entry AND reclassifies the section in `schemaSectionMetas()` + `sectionroute.go` in the same change. |
| KI-2 | Baked template defaults may diverge from the local `.moai/config/sections/*.yaml` if a developer hand-edits the local copy after `moai update`.                | Document that `moai update` is the re-sync path; the SPEC does not auto-overwrite local edits (preserves user customizations).   |
| KI-3 | The 4-color tier concept is net-new; a future docs-site / README tokenomics page may want to consume the same table.                                              | Scope-limit to `moai web` agentfm in this SPEC (spec.md §D). A follow-up SPEC can productize externally.                         |
| KI-4 | `max` effort and `inherit` model are valid closed-set values but have NO tier-color mapping (they denote manual override / session-inherit).                   | C-8 codifies this; the UI shows a neutral badge (e.g. "custom") when effort is `max` or model is `inherit`.                      |
| KI-5 | Test fallout in `internal/settings/*_test.go` (schema section enumeration) and `internal/web/*_test.go` (tab render, route coverage) is guaranteed.            | M7 updates each broken test to reflect the 6-tab set + tier model; tests are NOT deleted (C-6).                                   |
| KI-6 | The `agentFMRow` templ edit requires `make build` to recompile `*_templ.go`; forgetting the rebuild leaves the UI showing the old row layout.                  | M9 verification gate includes `make build` + a render smoke check.                                                              |
| KI-7 | Under Option A, the name→tier table may drift from a future agent's actual reasoning role if the agent is repurposed.                                            | The name table is a chosen mapping, not derived; a repurposed agent needs a manual table edit. M1.2 documents the rationale per agent so a future editor can re-evaluate. |
| KI-8 | The `harness.effort_mapping` baked triple `{low,medium,high}` is a meaningful downgrade from the current `{medium,high,xhigh}` — a behavioral shift some downstream tests may assert against. | M9 runs the full `go test ./...`; any assertion on the old triple is updated in M7. The delta is logged in spec.md §D.Δ.        |

---

## §C. Pre-flight (Run-Phase Entry Conditions)

Before `/moai run SPEC-WEBCONF-SIMPLIFY-001`, the run-phase agent MUST verify:

1. `cat internal/template/templates/.moai/config/sections/{harness,llm,workflow,security,git-strategy.tmpl,cache,ralph,feedback,observability,mx,handoff,project.yaml.tmpl,quality.yaml.tmpl}.yaml` and capture verbatim — these feed M2 baking (spec.md §E preserve-all-other-keys contract).
2. `grep -n 'RouteExcluded\|RouteTypedSave\|RouteSeam\|RouteStatusline' internal/settings/sectionroute.go` — confirm the route enum is intact before reclassification.
3. `grep -n 'consoleTabs\|schemaSectionMetas' internal/web/schemaform.go` — confirm line numbers (may have drifted from the :31-56 / :71-91 snapshot).
4. `grep -rn 'V4EffortValues\|V4ModelValues' internal/` — confirm accessor locations before adding the name-keyed tier table accessors.
5. `make build` baseline green + `go test ./internal/web/... ./internal/settings/...` baseline captured (pre-change pass/fail counts) — required to quantify M7 fallout.
6. Confirm `front-launch` is a phantom: `grep -rn 'front-launch\|front_launch\|FrontLaunch' internal/` returns no matches (already verified at plan-phase; re-confirm at run-phase start in case of drift).

---

## §D. Constraints (mirrored from spec.md §C)

C-1 Template-First · C-2 Template-neutrality §25 · C-3 16-language neutrality §15 · C-4 GLM carve-out · C-5 Atomic save contract · C-6 Test fallout (update not delete) · C-7 No agent frontmatter rewrite for tier DISPLAY (Option A — effort files untouched; per-agent override via existing `agentfm.Patch`) · C-8 Closed-set validation against `V4EffortValues` / `V4ModelValues` · C-9 Description i18n 4-locale parity (Refinement 2).

---

## §E. Self-Verification (Plan-Phase)

- **GEARS compliance**: every REQ-WC-### in `spec.md` §B uses `shall` + (ubiquitous | `When` | `While` | `Where`); no legacy `IF/THEN`.
- **Out-of-Scope lint**: `spec.md` §D carries six `### Out of Scope — <topic>` H3 sub-headings, each with `-` bullets.
- **Frontmatter schema**: 12 canonical fields + `tier: L` present in all 6 files; `id` matches `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`; `status: draft`.
- **AC traceability**: every REQ-WC-### has at least one AC in `acceptance.md` §D (traceability matrix).
- **Tier mapping decidability**: plan.md §F M1.2 + design.md §C cover all 20 enumerated agents — no agent unassigned, no ambiguity.
- **Clarification-marker zero-count [D1]**: plan.md §G carries NO unresolved clarification markers (OQ-1/OQ-2 resolved; OQ-3/OQ-4 retained as design defaults — non-blocking).
- **Option A coherence [D3]**: REQ-WC-006, C-7, EC-1, M1.1, M1.2 all describe the SAME single model (name-keyed lookup table, display-only, effort untouched). Cross-checked for internal consistency.
- **§E preserve-all-other-keys [D4]**: every §E.N block states the contract; spec.md §D.Δ enumerates the deliberate default-changes.

---

## §F. Milestones (ordered by decision-reversibility — most likely to change FIRST)

### M1 — Tier data-model decision + 20-agent name→tier table (DECISION)

**Highest change-likelihood: introduces a net-new concept and new accessor signatures. Lead with this so human review focuses here first.**

**M1.1 Data-model mechanism (RESOLVED — Option A: name-keyed lookup table, display-only) [D3]**

- The tier is a **name-keyed lookup table** keyed by agent name. It is the display-time source of truth for the color badge. It is INDEPENDENT of each agent's `effort` frontmatter — the 20 effort files are UNTOUCHED by the tier-display feature (C-7).
- Location: a static `agentName → tier` map (Go `map[string]Tier` or equivalent). Preferred location: `internal/harness/v4manifest/schema.go` (alongside the existing closed-set definitions) OR a sibling file, surfaced via accessors in `internal/settings/schema_sections.go` next to `V4EffortValues` / `V4ModelValues`.
- The REJECTED alternative (effort-derivation: `effort → color` map keyed off each agent's live effort frontmatter) was rejected because it (a) would require rewriting 20 agent effort files to align tier-with-effort, (b) creates a circular dependency for the 3 absent-effort agents (hns-oss-docs-*), and (c) makes the badge color flapping-dependent on whichever effort the user last set. design.md §G documents the rejection in full.
- The tier→suggested-(model,effort) table (M1.1 sub-table) IS used — but only as a SUGGESTION when the user clicks a tier color to auto-fill the model/effort selectors. Applying the suggestion writes `effort`/`model` via the existing `agentfm.Patch`; it does NOT change the name→tier table.

| Tier color | Suggested `model` | Suggested `effort` |
|------------|-------------------|--------------------|
| 🔴          | `opus`            | `xhigh`            |
| 🟠          | `opus`            | `high`             |
| 🔵          | `sonnet`          | `medium`           |
| 🩵          | `haiku`           | `low`              |
| (neutral)   | (no suggestion)   | `max` / `inherit` are override sentinels, NOT a 5th tier |

**M1.2 — 20-agent name→tier mapping (RESOLVED — display-only, chosen by reasoning role)**

The tier for each agent is CHOSEN by the agent's reasoning role. It is a static, hand-curated mapping — NOT derived from the agent's current effort. The "13 M1.2-vs-actual discrepancies" the iter-1 auditor counted (10 effort-value mismatches + 3 absent-effort agents) are EXPECTED and IRRELEVANT under Option A: the badge shows the chosen tier regardless of what the effort frontmatter currently says. (research.md §C documents the actual effort landscape for traceability.)

| Agent                                | Tier | Rationale (reasoning-role basis)                                                                            |
|--------------------------------------|------|-------------------------------------------------------------------------------------------------------------|
| `manager-spec`                       | 🔴   | Plan-phase SPEC authoring; deep requirement + GEARS reasoning.                                              |
| `plan-auditor`                       | 🔴   | Independent audit; bias-prevention judgment; GEARS compliance evaluation.                                   |
| `super-advisor`                      | 🔴   | E1-E4 high-reasoning consultation; non-binding prescriptions require deepest reasoning.                     |
| `sync-auditor`                       | 🔴   | Independent skeptical 4-dimension scoring; harmonic-mean verdict; deep judgment parallel to plan-auditor.   |
| `manager-develop`                    | 🟠   | Run-phase DDD/TDD implementation; heavy reasoning but bounded by the SPEC.                                  |
| `manager-design`                     | 🟠   | D1-D5 design pipeline; creative + structural synthesis.                                                     |
| `builder-harness`                    | 🟠   | Dynamic harness / specialist generation; structural reasoning about agent/skill architecture.               |
| `e2e-tester`                     | 🟠   | Web/mobile/desktop journey scripting; cross-platform complexity.                                            |
| `manager-docs`                       | 🔵   | Sync-phase documentation; template-driven, moderate reasoning.                                              |
| `manager-git`                        | 🔵   | PR creation per Tier routing; procedural git ops.                                                           |
| `quality-specialist`                 | 🔵   | TRUST 5 validation; checklist-driven quality checks.                                                        |
| `cli-template-specialist`            | 🔵   | Template editing; pattern-application, moderate reasoning.                                                  |
| `workflow-specialist`                | 🔵   | Workflow pattern application; procedural.                                                                   |
| `hns-github-specialist`              | 🩵   | Narrow GitHub issue/PR scope via `gh` CLI.                                                                  |
| `hns-oss-docs-content-author-specialist` | 🩵 | Narrow README/docs content authoring scope.                                                                 |
| `hns-oss-docs-locale-translator-specialist` | 🩵 | Narrow locale-translation scope.                                                                       |
| `hns-oss-docs-structure-curator-specialist` | 🩵 | Narrow docs-site structure/navigation scope.                                                            |
| `hns-release-specialist`             | 🩵   | Narrow release-publishing scope (Enhanced GitHub Flow scripts).                                             |
| `hns-release-update-specialist`      | 🩵   | Narrow CC upstream-change tracking scope.                                                                   |
| `hook-ci-specialist`                 | 🩵   | Narrow CI/hooks scope.                                                                                      |

**Distribution**: 🔴×4 · 🟠×4 · 🔵×5 · 🩵×7 = 20. Supersedes the stale 10-agent memory `project_agent_token_cost_color_tiers.md`.

### M2 — Template defaults baked (data shipping; preserve-all-other-keys)

- For each section file in `internal/template/templates/.moai/config/sections/`, `cat` the live file, preserve ALL existing keys verbatim, and overwrite ONLY the keys-of-interest listed in spec.md §E.1–§E.10 with the baked values.
- Honor the spec.md §D.Δ deliberate default-change table — apply each delta exactly (do NOT silently "fix" a baked value back to the old template value).
- Sections touched: `harness.yaml`, `llm.yaml`, `workflow.yaml`, `security.yaml`, `git-strategy.yaml.tmpl`, `cache.yaml`, `ralph.yaml`, `feedback.yaml`, `observability.yaml`, `mx.yaml`, `handoff.yaml`, `project.yaml.tmpl`, `quality.yaml.tmpl`.
- Run `make build` after the template edits (constraint C-1).
- Template-neutrality §25 sweep (C-2): confirm no SPEC IDs / REQ tokens / SHAs leak into the baked YAML.

### M3 — Tab set reduction (schema reclassification mechanics)

- Delete the 11 removed-tab entries from `internal/web/schemaform.go` `consoleTabs()` (lines ~31-56; reconfirm at run-phase start per §C preflight step 3).
- Reclassify each removed section in `schemaSectionMetas()` (lines ~71-91) so it is no longer rendered.
- Reclassify each removed section's route in `internal/settings/sectionroute.go` to `RouteExcluded` — config persists, write path removed (KI-1 mitigation). Exception: `quality_extras` retains a toggle on the `launch` tab (M4).
- Update `internal/settings/schema.go` `allFields()` if it carries render-hints that would re-surface removed tabs.
- The `front-launch` identifier is a phantom (D2) — no removal action; if a stale reference appears anywhere, delete it as dead code.

### M4 — Simplified surviving-tab surfaces + field descriptions [Refinement 1 + Refinement 2]

- **git_strategy**: surface the core fields `mode` + `merge_method` + `hooks.pre_push` for the active mode as top-level selectors (REQ-WC-016, Refinement 1); the per-provider nesting (`branch_creation`, `automation`, `commit_style`, `github_integration`, `push_to_remote`, `draft_pr`, `required_reviews`, `branch_protection`) bakes as template defaults (M2) and remains UI-hidden.
- **llm**: surface ONLY `glm.models.{high, medium, low, fable}` tier mapping; hide `mode`, `team_mode`, `performance_tier`, `plan_type`, `claude_models` (baked).
- **quality_extras**: replace the full tab with a single enable/disable toggle on the **`launch`** tab (OQ-1 resolved — D1).
- **Field descriptions (Refinement 2, REQ-WC-015)**: each selectable field rendered across the `identity`, `language`, `launch`, `git_strategy`, and `llm` tabs SHALL render a localized description below the field label (per the design.md §H mechanism: `FieldDef.Description` i18n key resolved via `i18n.js`). Each `<option>` within a select carries a per-option description via the native `title` attribute. The description strings themselves are authored in M6 (4-locale).

### M5 — Sub-agent FM color-tier UI redesign + stale-comment sweep [D13]

- Redesign `internal/web/fieldsets.templ:385` `agentFMRow` to render:
  1. A color tier badge from the name-keyed lookup table (M1.2) — via the new accessor.
  2. A model `<select>` populated from `V4ModelValues`.
  3. An effort `<select>` populated from `V4EffortValues`.
- Selecting a tier (color) auto-suggests the tier's suggested model+effort pair (M1.1 sub-table); each selector remains individually overridable (REQ-WC-008). Applying the suggestion writes `effort`/`model` via the existing `agentfm.Patch` (C-7 — no new FM key).
- Update `fieldsetAgentFrontmatter` (`fieldsets.templ:361`) to thread the name-table accessor through.
- Update `internal/web/agentfm.go:79-117` parse/validate to closed-set-validate against `V4EffortValues` / `V4ModelValues` (REQ-WC-009, C-8).
- **Stale-comment sweep [D13]**: update `fieldsets.templ:357` "7 sub-agent" → "20 sub-agent" (actual catalog count); update `agentfm.go:49` "8 retained agents" → "10 moai-custom agents" (actual moai/ count; harness/ adds 10 more). Comments only — no behavior change.
- **Tier option descriptions (Refinement 2, REQ-WC-015)**: the agentfm tab's tier-color picker options (🔴🟠🔵🩵 + the neutral "custom") and the model/effort selectors each carry a per-option description (via `title` attribute / i18n key per design.md §H). Description strings authored in M6.
- Run `make build` (KI-6).

### M6 — Frontend assets + 4-locale i18n (tier labels + per-field/option descriptions) [Refinement 2 makes this the heaviest milestone]

- Add tier-label strings (color names, "custom" neutral label, tooltip text for `max`/`inherit`) to `internal/web/assets/i18n.js` in all four locales (en, ko, ja, zh).
- **Per-field/option description strings (Refinement 2, REQ-WC-015, C-9)**: author `fieldDesc.<sectionID>.<fieldID>` entries (and `fieldDesc.<sectionID>.<fieldID>.option.<value>` for per-option) for every selectable field across the 6 surviving tabs, in all 4 locales. This is the single largest authoring burden in the SPEC — estimate ~200–400 strings per locale × 4 locales = ~800–1600 strings total. Recommended run-phase approach: draft `en` first, then translate to `ko`/`ja`/`zh`; consider mechanical aid where sensible. Budget accordingly.
- Add the color-badge CSS + the `.field-description` muted-helper-text CSS class to `internal/web/assets/console.css`.
- Wire the tier-badge render AND the field-description-below-label render in `internal/web/assets/app.js`.

### M7 — Test fallout updates

- Update `internal/settings/*_test.go`: schema-section enumeration tests to reflect the 6-tab set + `RouteExcluded` reclassification + any assertion referencing the old `harness.effort_mapping` triple (KI-8).
- Update `internal/web/*_test.go`: tab-render tests (assert exactly 6 tabs), route-coverage tests, and add agentfm tier-badge render tests (assert color comes from the name table for each of the 20 agents — NOT from effort).
- Tests are UPDATED, not deleted (C-6).

### M8 — Web-config documentation cleanup

- Remove ONLY user-facing web-config guidance documentation (the tabs-explained pages that reference the 11 removed tabs).
- PRESERVE verbatim: `mx-tag-protocol.md`, `context-window-management.md`, `cache-aware-execution.md`, and all other `.claude/rules/moai/` doctrine (Decision 4, REQ-WC-010).

### M9 — Final verification

- `make build` succeeds (KI-6).
- `go test ./internal/web/... ./internal/settings/...` passes (C-6).
- `go test ./...` full-suite green (no cascade; KI-8 covers the effort_mapping triple).
- Template-neutrality CI guard (`.github/workflows/template-neutrality-check.yaml`) passes (C-2).
- GLM carve-out verified: `grep -n 'InjectCacheControl' internal/runtime/cache_control.go` confirms the GLM-omit branch is intact (C-4).
- 6-tab render smoke check via `moai web` (or a render test) confirms exactly 6 tabs.
- Name-table coherence: a unit test asserts all 20 catalog agents have a tier entry (no agent unassigned; no orphan entry for a non-existent agent).

---

## §G. Open Questions (resolved at iter-1 — ZERO unresolved clarification markers remain [D1])

- **OQ-1 (RESOLVED → `launch` tab)**: the `quality_extras` enable/disable toggle lives on the **`launch`** tab. Encoded in REQ-WC-004 + AC-WC-004 + M4.
- **OQ-2 (RESOLVED → phantom)**: `front-launch` is a phantom (no code identifier); its removal is a no-op. Encoded in REQ-WC-002 + GWT-2 + M3.
- **OQ-3 (design default — non-blocking)**: tier badge carries color + tooltip (lower i18n burden than a visible textual role label). Default recorded; user can upgrade to a visible label later without SPEC change.
- **OQ-4 (design default — non-blocking)**: `max` effort / `inherit` model render a neutral grey "custom" badge with tooltip (NOT a 5th/6th tier). Default recorded in C-8 + M1.1.

No unresolved clarification markers remain in this file or in research.md.

---

## §H. Anti-Patterns

- **AP-1** Blindly deleting failing tests in `internal/settings/*_test.go` / `internal/web/*_test.go` instead of updating them to the new 6-tab + tier-model reality (violates C-6).
- **AP-2** Deriving the tier badge from the agent's live `effort` frontmatter instead of from the name-keyed lookup table (violates C-7 / Option A; causes circular-dependency for absent-effort agents and badge flapping when effort changes).
- **AP-3** Adding a new agent frontmatter key (e.g. `tier:`) instead of using the name-keyed lookup table (violates C-7; causes 20-agent-file churn).
- **AP-4** Removing a `consoleTabs()` entry WITHOUT reclassifying the section route to `RouteExcluded`, leaving a dangling write path that breaks `handleSave` (KI-1, violates C-5).
- **AP-5** Editing local `.moai/config/sections/*.yaml` directly instead of `internal/template/templates/.moai/config/sections/*.yaml` (violates C-1 Template-First; the change would not ship to users).
- **AP-6** Modifying `cache_control.go`'s GLM-omit branch while removing the cache tab (violates C-4; the carve-out is orthogonal to the UI removal).
- **AP-7** Embedding SPEC IDs / REQ tokens / commit SHAs into the baked template YAML (violates C-2 template-neutrality §25).
- **AP-8** "Fixing" a spec.md §D.Δ deliberate default-change back to the old template value (violates Decision 2; the baked value IS the new default).
- **AP-9** Treating `max` effort or `inherit` model as a 5th/6th tier color — they are override sentinels, not tiers (C-8, KI-4).

---

## §I. Cross-References

- `spec.md` §B (REQ-WC-001..016) — the encoded requirements; §C (C-1..C-9) — constraints; §D — Out-of-Scope sub-headings; §D.Δ — deliberate default-change table; §E — baked keys-of-interest.
- `design.md` §A–§H — Option A data-model design, name→tier table, UI architecture, rejected effort-derivation alternative, description-source mechanism (§H).
- `research.md` §A — verified codebase paths; §B — live-template key inventory; §C — 20-agent effort landscape.
- `acceptance.md` §D — AC matrix mapped to REQs.
- `CLAUDE.local.md` §2 (Template-First) / §15 (16-language neutrality) / §25 (template-internal-isolation).
- Memory `project_agent_token_cost_color_tiers.md` — stale 10-agent split, superseded by M1.2.
