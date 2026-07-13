---
id: SPEC-WEBCONF-SIMPLIFY-001
title: "moai web Configuration UI Simplification + Sub-Agent 4-Color Tier Redesign"
version: "0.2.0"
status: in-progress
created: 2026-07-13
updated: 2026-07-13
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "internal/web, internal/settings, internal/template/templates"
lifecycle: spec-anchored
tags: "web-ui, config, sub-agent, tier-model, template-defaults"
tier: L
---

## HISTORY

| Version | Date       | Author       | Change                                                                                                                        |
|---------|------------|--------------|-------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-07-13 | manager-spec | Initial plan-phase authoring (4 locked inputs).                                                                               |
| 0.2.0   | 2026-07-13 | manager-spec | iter-1 audit-fix (FAIL 0.69 → revise). D3 Option A name-keyed tier table (display-only, effort untouched). Tier M → L. Drop `front-launch` phantom. §E preserve-all-other-keys + project/handoff/quality_extras. Deliberate default-change flags. design.md + research.md authored (6-artifact Tier L set). |

---

## §A. Vision and Scope

### §A.1 Problem

The `moai web` configuration console (default port 3041) currently surfaces 17 tabs covering every internal config section. The tab sprawl overwhelms new users, exposes runtime-only / internal-developer state through a user-facing surface, and offers no token-cost reasoning model for the 20-agent sub-agent catalog. The sub-agent frontmatter (agentfm) tab lets editors pick `model` and `effort` per agent but provides no tier signal — an editor cannot see at a glance which agents are reasoning-heavy (expensive) versus narrow-specialist (cheap).

### §A.2 Vision

A simplified web console that surfaces ONLY the configuration a typical user needs to touch, with every removed key preserved as a shipped template default. A new 4-color tier model (🔴🟠🔵🩵) makes per-agent token cost visible at a glance. Per Decision 3 Option A, the tier is a **name-keyed lookup table** (display-only): the badge color comes from a static `agent-name → color` table chosen by each agent's reasoning role, INDEPENDENT of the agent's current `effort` frontmatter. The `model` + `effort` selectors remain the only per-agent write levers, and per-agent override writes `effort` via the existing `agentfm.Patch` mechanism — no agent frontmatter file is rewritten merely to display a tier badge.

### §A.3 In Scope

- Reducing the `moai web` tab set from 17 to exactly 6 surviving tabs.
- Baking the current values of all 11 removed tabs into `internal/template/templates/.moai/config/sections/*.yaml` as shipped defaults.
- Introducing a 4-color tier concept (🔴xhigh / 🟠high / 🔵medium / 🩵low) for the 20-agent catalog, implemented as a **name-keyed lookup table** (Option A), NOT derived from the `effort` frontmatter.
- Redesigning the agentfm UI to render a color badge (from the name table) + model selector + effort selector per agent.
- Removing only web-config user-facing guidance documentation; deep doctrine is preserved.

> **Schema-space siblings (out-of-scope, D12)**: `SectionStatusline`, `SectionQuality` (the main quality section backed by `quality.yaml.tmpl` — DISTINCT from the removed `quality_extras` tab), and `SectionGitConvention` share the schema space in `internal/settings/schema.go:26-36`. They are UNAFFECTED by this SPEC — the 6 surviving tabs and 11 removed tabs do not touch these three section IDs. Listed here only to prevent a run-phase agent from mistakenly sweeping them.

### §A.4 Out of Scope — high-level

See §D for the structured Out-of-Scope sub-headings.

---

## §B. Requirements (GEARS)

The four LOCKED user decisions are encoded as REQ-WC-001 through REQ-WC-010. Each requirement cites its originating decision in `[D<N>]`. iter-1 audit fixes are marked `[D1]`..`[D13]`.

### Decision 1 — Surviving tabs (exactly 6)

**REQ-WC-001** [D1] The `moai web` console shall render exactly six tabs in the following order: `identity`, `language`, `launch`, `git_strategy`, `llm`, and `agentfm`.

**REQ-WC-002** [D1, D2] The `moai web` console shall not render the eleven removed tabs — `project`, `quality_extras` (subject to the REQ-WC-004 toggle exception), `workflow`, `harness`, `ralph`, `feedback`, `observability`, `security`, `mx`, `handoff`, and `cache` — in the tab navigation. (6 surviving + 11 removed = 17 total. The previously listed `front-launch` identifier is a phantom — no such code identifier exists — and is dropped from the removed-tab list; its "removal" is a no-op.)

**REQ-WC-003** [D1] Where a tab is removed from the web UI, the configuration keys that tab previously edited shall persist in the baked template YAML (`internal/template/templates/.moai/config/sections/*.yaml`) for runtime consumption — UI-hidden, NOT deleted from the config schema.

**REQ-WC-004** [D1, D1-resolved, D7] Where the `quality_extras` tab is removed, the console shall surface a single enable/disable toggle controlling the quality-extras feature on the **`launch`** tab (the resolved placement, OQ-1).

### Decision 2 — Default scope = TEMPLATE defaults

**REQ-WC-005** [D2] The baked default values for every removed tab shall ship from `internal/template/templates/.moai/config/sections/*.yaml` so that `moai init` and `moai update` distribute them identically to all new and existing user projects. Where a baked value intentionally differs from the current template default (a deliberate default-change), §D.Δ enumerates each delta.

### Decision 3 — Sub-agent 4-color tier model (Option A: name-keyed lookup table)

**REQ-WC-006** [D3] The sub-agent tier model shall classify each of the 20 agents in the current catalog into exactly one of four color tiers — 🔴 (`xhigh`), 🟠 (`high`), 🔵 (`medium`), 🩵 (`low`) — via a **name-keyed tier lookup table** keyed by agent name. The tier assignment is INDEPENDENT of the agent's current `effort` frontmatter value: the table is chosen by each agent's reasoning role and is the display-time source of truth for the color badge. No agent frontmatter file is rewritten merely to display a tier badge.

**REQ-WC-007** [D3] The tier model shall expose the following per-tier default suggestion pairs (used when a user clicks a tier color to auto-suggest): 🔴 → `opus` / `xhigh`; 🟠 → `opus` / `high`; 🔵 → `sonnet` / `medium`; 🩵 → `haiku` / `low`. Each suggestion is applied only on explicit user action; the table itself does not write to any agent file.

**REQ-WC-008** [D3] The agentfm UI shall render, per agent, a color tier badge (from the name-keyed lookup table) alongside a model `<select>` and an effort `<select>`, where selecting a tier (color) auto-suggests the tier's default model+effort pair and each selector remains individually overridable. Per-agent override writes the `effort` (and/or `model`) frontmatter via the existing `agentfm.Patch` mechanism — NO new frontmatter key is introduced.

**REQ-WC-009** [D3] The model and effort selectors shall be closed-set validated against the existing `V4EffortValues` (`{low, medium, high, xhigh, max}`) and `V4ModelValues` (`{inherit, haiku, sonnet, opus}`) enumerations; an out-of-set value shall be rejected at validation time. The tier-color map itself is a closed 4-entry set {🔴, 🟠, 🔵, 🩵}; `max` effort and `inherit` model are valid selector values but receive NO tier-color mapping (they denote manual override / session-inherit).

### Decision 4 — Doc removal scope

**REQ-WC-010** [D4] The documentation cleanup shall remove only user-facing web-config guidance; the deep doctrine files — `mx-tag-protocol.md`, `context-window-management.md`, `cache-aware-execution.md` — shall be preserved verbatim because the runtime depends on them.

### Cross-cutting behavioral requirements

**REQ-WC-011** When a user saves any of the six surviving tabs, the `handleSave` handler (`internal/web/handlers.go`) shall persist the values without invoking validators for fields belonging to removed tabs, so the atomic save contract is preserved.

**REQ-WC-012** Where the active backend is GLM, the prompt-cache injection (`internal/runtime/cache_control.go` `InjectCacheControl`) shall remain a no-op (omitted), preserving the existing GLM carve-out independent of the cache tab removal.

**REQ-WC-013** [D10] When a test in `internal/settings/*_test.go` or `internal/web/*_test.go` references a removed tab or a removed section field, the test shall be updated to reflect the new six-tab set and the new tier model — not deleted.

**REQ-WC-014** The `moai web` console shall render its user-visible strings in all four shipped locales (en, ko, ja, zh) via the `internal/web/assets/i18n.js` dictionary; no new locale is added by this SPEC.

---

## §C. Constraints

| ID    | Constraint                                                                                                                                                       | Source                                              |
|-------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| C-1   | **Template-First** — every new or changed file under `.claude/` or `.moai/` MUST also be mirrored in `internal/template/templates/`, then `make build`.           | `CLAUDE.local.md` §2                                |
| C-2   | **Template-neutrality §25** — baked defaults MUST NOT contain SPEC IDs, REQ/AC tokens, commit SHAs, internal dates, or archive paths. Operational config values are allowed. | `CLAUDE.local.md` §25                               |
| C-3   | **16-language neutrality §15** — no Go-language bias in any template file; the web console's config sections are language-agnostic.                               | `CLAUDE.local.md` §15                               |
| C-4   | **GLM carve-out** — prompt-cache injection MUST remain a no-op on GLM backends; the existing `cache_control.go` behavior is preserved.                            | `internal/runtime/cache_control.go`                 |
| C-5   | **Atomic save contract** — removed tabs MUST NOT break `handleSave` (`internal/web/handlers.go`); validators for removed fields MUST NOT fire on save of a surviving tab. | `internal/web/handlers.go`                          |
| C-6   | **Test fallout** — schema-section and route-coverage tests in `internal/settings/*_test.go` and `internal/web/*_test.go` WILL break and MUST be updated to reflect the new six-tab set, not blindly deleted. | `internal/settings/`, `internal/web/`               |
| C-7   | [D3] **No agent frontmatter rewrite for tier DISPLAY** — the tier axis is a name-keyed lookup table; the 20 agent effort files are UNTOUCHED by the tier-display feature. Per-agent model/effort override remains via the EXISTING `agentfm.Patch` mechanism (writes `effort` and/or `model` only; NO new frontmatter key). | Decision 3 Option A + design.md §B                  |
| C-8   | **Closed-set validation** — the model and effort selectors are validated against `V4EffortValues` / `V4ModelValues`; `max` effort and `inherit` model are valid selector values but receive no tier-color mapping (manual override sentinels, NOT a 5th/6th tier). | `internal/harness/v4manifest/schema.go`             |

---

## §D. Out of Scope

### Out of Scope — Implementation code

- This SPEC authoring task produces the 6 plan-phase artifacts (spec.md + plan.md + acceptance.md + progress.md + design.md + research.md) ONLY. No Go source, templ, CSS, JS, or YAML is modified by this plan-phase task.
- Run-phase implementation (`/moai run SPEC-WEBCONF-SIMPLIFY-001`) is owned by `manager-develop` and is explicitly out of scope here.

### Out of Scope — Deep doctrine removal

- `mx-tag-protocol.md`, `context-window-management.md`, `cache-aware-execution.md`, and all other `.claude/rules/moai/` doctrine files are PRESERVED. Only user-facing web-config guidance documentation is in scope for removal (Decision 4).
- The `session-handoff.md` cut-line markers, the `goal-directive.md` `/goal` flow, and the `verification-claim-integrity.md` invariant are untouched.

### Out of Scope — Agent catalog membership changes

- The 20-agent catalog (10 `moai/` + 10 `harness/`) is taken as-is. This SPEC does not add, remove, rename, or retire any agent.
- The tier mapping assigns a color to each of the 20 existing agents; it does NOT change which agents exist.

### Out of Scope — 4-color concept productization beyond `moai web`

- The 4-color tier badge is a `moai web` agentfm UI concept only. Docs-site (`adk.mo.ai.kr`) tokenomics pages, README rendering, CLI `--tier` flags, and `moai spec` frontmatter `tier:` field (S/M/L) are NOT modified by this SPEC — they are separate concepts that happen to share the word "tier".
- A future SPEC may productize the 4-color model beyond the web console; that future work is explicitly out of scope here.

### Out of Scope — Runtime behavior changes

- Removed-tab config keys continue to be consumed at runtime exactly as before. Their VALUES are baked as template defaults; their CONSUMPTION code paths are unchanged.
- The `cacheStrategy` config is still read by `internal/config/cache_control.go` at runtime; only the cache TAB is removed.

### Out of Scope — Schema-space siblings (D12)

- `SectionStatusline`, `SectionQuality` (the main quality section, distinct from the removed `quality_extras` tab), and `SectionGitConvention` (`internal/settings/schema.go:26-36`) share the schema space but are UNAFFECTED by this SPEC. The 6 surviving tabs and 11 removed tabs do not touch these three.

### Out of Scope — New locales

- No new `conversation_language` is added. The four shipped locales (en, ko, ja, zh) cover the new tier-label strings.

---

## §D.Δ Deliberate Default-Change Flags (Decision 2 transparency)

Per Decision 2, baked values are the user's CURRENT values (sourced from screenshots), which in several cases INTENTIONALLY DIFFER from the current `internal/template/templates/` defaults. Each delta is a deliberate default-change — the baked value becomes the NEW shipped template default. Run-phase MUST apply these exactly as listed (not silently "fix" them to match the old template). [D6]

| Section.key                               | Current template value                                   | Baked (NEW default)                            | Rationale                                       |
|-------------------------------------------|----------------------------------------------------------|------------------------------------------------|-------------------------------------------------|
| `workflow.execution_mode`                 | `team`                                                   | `auto`                                         | User screenshot; deliberate auto-selection default. |
| `security.permission.strict_mode`         | `false`                                                  | `true`                                         | User screenshot; stricter default.              |
| `cache.cacheStrategy.enabled`             | `false` (ships disabled)                                 | `true`                                         | User screenshot; opt-in to cache injection.    |
| `git-strategy.*.merge_method`             | `squash` (lowercase)                                     | `Squash` (capitalized)                         | User screenshot value preserved verbatim.      |
| `git-strategy.personal.hooks.pre_push`    | `warn`                                                   | `enforce`                                      | User screenshot; stricter personal-mode default.|
| `git-strategy.team.hooks.pre_push`        | `warn`                                                   | `enforce`                                      | User screenshot; stricter team-mode default.   |
| `harness.effort_mapping.minimal`          | `medium`                                                 | `low`                                          | User screenshot; cost-reduction default.       |
| `harness.effort_mapping.standard`         | `high`                                                   | `medium`                                       | User screenshot; cost-reduction default.       |
| `harness.effort_mapping.thorough`         | `xhigh`                                                  | `high`                                         | User screenshot; cost-reduction default.       |

> **Note on `harness.effort_mapping`**: the baked triple `{low, medium, high}` downgrades reasoning depth across all three harness levels vs the current `{medium, high, xhigh}`. This aligns with the v3.0 tokenomics direction but is a meaningful behavioral shift. It is surfaced here (not as an unresolved clarification marker) per the iter-1 D1 directive — the value is what the user specified; run-phase applies it verbatim and the change is visible in the §E.1 block below + this delta table.

---

## §E. Baked Default Values (keys-of-interest + preserve-all-other-keys)

> **§E scope contract [D4]**: each §E.N block below shows ONLY the keys-of-interest — the headline keys the user decision names. The live template files (`internal/template/templates/.moai/config/sections/<section>.yaml`) carry MANY additional keys (e.g. `harness.yaml` ≈ 8165B includes `plan_audit_global`, `levels.{minimal,standard,thorough}`, `model_upgrade_review.checklist[].{question,affects}`, `learning.rate_limit`, etc.; `llm.yaml` nests `glm.{base_url, models.{high,medium,low,fable}}` + `claude_models` + `performance_tier` + `plan_type`; `workflow.yaml` carries `workflow_agents` 7-purpose taxonomy + `model_routing` legacy 12-cell + `model_routing_profiles` 36-cell). Run-phase MUST `cat` each live file, preserve ALL existing keys verbatim, and overwrite ONLY the keys-of-interest listed here with the baked values. The §D.Δ table above enumerates the keys where the baked value intentionally differs from the current template value.

### §E.1 harness (keys-of-interest; preserve plan_audit_global, levels, model_upgrade_review, learning.rate_limit, etc.)

```yaml
harness:
  default_profile: default
  evaluator:
    memory_scope: per_iteration
  mode_defaults:
    cg: thorough
    solo: auto
    team: auto
  auto_detection:
    enabled: true
  escalation:
    enabled: true
    max_escalations: 2
  effort_mapping:        # DELIBERATE default-change (§D.Δ): {medium,high,xhigh} → {low,medium,high}
    minimal: low
    standard: medium
    thorough: high
learning:
  enabled: true
  log_retention_days: 90
  auto_apply: false  # FROZEN
```

### §E.2 security (keys-of-interest; preserve extra_*_patterns, sandbox.network_allowlist, etc.)

```yaml
security:
  permission:
    strict_mode: true  # DELIBERATE default-change (§D.Δ): false → true
  sandbox:
    required: false
    docker_image: alpine:latest
```

### §E.3 workflow (keys-of-interest; preserve workflow_agents, model_routing, model_routing_profiles, etc.)

```yaml
workflow:
  execution_mode: auto  # DELIBERATE default-change (§D.Δ): team → auto
  agentic_loop:
    max_iterations: 10
  auto_clear:
    enabled: true
    after_plan: true
    after_run: false
    token_threshold: 150000
  loop_prevention:
    failure_pattern_detection: true
    max_iterations: 100
    max_retries_per_operation: 3
  token_budget:
    plan: 30000
    run: 180000
    sync: 40000
  worktree:
    auto_cleanup: true
    auto_create: false
    auto_merge: true
    tmux_preferred: true
```

### §E.4 git-strategy (SIMPLIFIED — UI exposes ONLY `mode`; full block baked)

```yaml
git_strategy:
  mode: Manual
  manual:
    hooks:
      pre_push: warn
    merge_method: Squash        # DELIBERATE (§D.Δ): squash → Squash
  personal:
    hooks:
      pre_push: enforce         # DELIBERATE (§D.Δ): warn → enforce
    merge_method: Squash
  team:
    hooks:
      pre_push: enforce         # DELIBERATE (§D.Δ): warn → enforce
    merge_method: Squash
```

The web UI exposes ONLY the `mode` field; the rest bake as template defaults.

### §E.5 llm (SIMPLIFIED — UI exposes `glm.models.{high,medium,low,fable}` tier mapping; hides runtime-only `mode`/`team_mode`)

```yaml
llm:
  mode: ""
  team_mode: ""
  glm:
    base_url: "https://api.z.ai/api/anthropic"   # preserve from live
    models:
      high: glm-5.2
      medium: glm-4.7
      low: glm-4.5-air
      fable: glm-5.2
  # preserve from live: performance_tier, plan_type, claude_models, glm_env_var
```

The web UI exposes ONLY the `glm.models.{high,medium,low,fable}` tier mapping; `mode`, `team_mode`, `performance_tier`, `plan_type`, `claude_models` bake as template defaults.

### §E.6 cache (TAB removed, value baked)

```yaml
cacheStrategy:
  enabled: true        # DELIBERATE default-change (§D.Δ): false → true
  session_ttl: "1h"
```

### §E.7 ralph / feedback / observability / mx

Each section's current `internal/template/templates/.moai/config/sections/{ralph,feedback,observability,mx}.yaml` content is baked as-is. The exact key set is read at run-phase start by `cat`-ing each live file and copying verbatim — do not paraphrase. UI access is removed; values persist.

### §E.8 handoff (TAB removed, value baked) [D5]

```yaml
handoff:
  mode: manual
  guide: false
```

### §E.9 project (TAB removed, value baked) [D5]

The current `internal/template/templates/.moai/config/sections/project.yaml.tmpl` content is baked as-is. Run-phase `cat`s the live template file and copies verbatim.

### §E.10 quality_extras (TAB removed EXCEPT enable/disable toggle; keys live inside quality.yaml.tmpl) [D5]

`quality_extras` has NO separate template file. Its keys live inside `internal/template/templates/.moai/config/sections/quality.yaml.tmpl` — specifically the non-`constitution` blocks (`report_generation`, `lsp_state_tracking`, and the `constitution.coverage_exemptions` / `constitution.test_quality` sub-blocks). The full content bakes as-is; the web UI removes the full tab and surfaces ONLY the enable/disable toggle (REQ-WC-004, on the `launch` tab).

---

## §F. Cross-References

- **Codebase map + live-template key inventory**: see `research.md` §A (verified paths) + §B (live-template-structure findings, including harness.yaml 8165B key inventory, llm.yaml nested structure, workflow.yaml execution_mode=team, the 20-agent effort landscape with 13 M1.2-vs-actual discrepancies and 3 absent-effort hns-oss-docs-* agents).
- **Tier data-model design (Option A name-keyed table)**: see `design.md` §A–§G (table location, name→tier mapping, tier→suggested-model/effort, UI architecture, §E baking approach, the rejected effort-derivation alternative).
- **Tier mapping (20 agents, name→color)**: see `plan.md` §F milestone M1.2 + `design.md` §C.
- **Acceptance criteria**: see `acceptance.md` §D AC matrix.
- **CLAUDE.local.md §2 / §15 / §25**: Template-First, 16-language neutrality, template-internal-isolation doctrine.
- **CLAUDE.md §4**: retained 11-agent catalog (the web `agentfm` tab edits a superset that includes harness specialists — 20 total).
