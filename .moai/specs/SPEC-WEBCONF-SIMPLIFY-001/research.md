---
id: SPEC-WEBCONF-SIMPLIFY-001
title: "moai web Configuration UI Simplification + Sub-Agent 4-Color Tier Redesign — Research"
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

| Version | Date       | Change                                                          |
|---------|------------|-----------------------------------------------------------------|
| 0.2.0   | 2026-07-13 | Initial research authoring (Tier L artifact, iter-1 audit-fix). |
| 0.3.0   | 2026-07-13 | In-progress amendment note: §A schema.go row gains `FieldDef.Description` (Refinement 2 / REQ-WC-015 per-option descriptions — design.md §H.1 option (a) hybrid). No other codebase-map content changed. |

---

## §A. Verified codebase map (confirmed by `ls` / `grep` at plan-phase)

| Concern                          | Path                                                                                                        | Confirmed | Notes                                                                                                              |
|----------------------------------|-------------------------------------------------------------------------------------------------------------|-----------|--------------------------------------------------------------------------------------------------------------------|
| Web handler entry                | `internal/cli/web.go` → `internal/web/server.go`                                                            | ✓         | Port default 3041.                                                                                                 |
| SSOT schema                      | `internal/settings/schema.go` (`FieldDef`, `allFields()`, `SectionID` constants ~L25-53)                   | ✓         | `SectionStatusline`/`SectionQuality`/`SectionGitConvention` ~L26-36 UNAFFECTED (spec.md §D D12). v0.3.0: `FieldDef` gains a `Description string` field (i18n key) for REQ-WC-015 per-option descriptions — design.md §H.1 option (a) hybrid. |
| Extension-section fields         | `internal/settings/schema_sections.go`                                                                      | ✓         | `V4EffortValues` / `V4ModelValues` accessors ~L46-47.                                                              |
| Tab nav                          | `internal/web/schemaform.go:31-56` `consoleTabs()` + `:71-91` `schemaSectionMetas()`                        | ✓         | Removing a tab = delete entry + reclassify.                                                                        |
| Persistence routing              | `internal/settings/sectionroute.go` (`RouteTypedSave` / `RouteSeam` / `RouteStatusline` / `RouteExcluded`)  | ✓         | Removed tabs → `RouteExcluded`.                                                                                    |
| Save handler                     | `internal/web/handlers.go` (~L270-400 `handleSave`)                                                         | ✓         | Atomic save contract.                                                                                              |
| Sub-agent FM (backend)           | `internal/settings/agentfm/agentfm.go` (`List` / `Patch`)                                                   | ✓         | Stale comment `:49` "8 retained agents" (actual 10 moai-custom; D13).                                              |
| Sub-agent FM (UI)                | `internal/web/fieldsets.templ:361` `fieldsetAgentFrontmatter` + `:385` `agentFMRow`                         | ✓         | Stale comment `:357` "7 sub-agent" (actual 20; D13).                                                               |
| Sub-agent FM (parse/validate)    | `internal/web/agentfm.go:79-117`                                                                             | ✓         | Hooks into `V4EffortValues` / `V4ModelValues`.                                                                     |
| Closed sets                      | `internal/harness/v4manifest/schema.go:43-74`                                                                | ✓         | `effort ∈ {low, medium, high, xhigh, max}`; `model ∈ {inherit, haiku, sonnet, opus}`.                              |
| Prompt cache                     | `internal/runtime/cache_control.go` `InjectCacheControl`                                                    | ✓         | GLM-omit carve-out (C-4).                                                                                          |
| Cache config                     | `internal/config/cache_control.go` + `.moai/config/sections/cache.yaml`                                     | ✓         | Root key `cacheStrategy`.                                                                                          |
| Frontend assets                  | `internal/web/assets/{app.js, console.css, i18n.js}`                                                        | ✓         | 4-locale i18n dict.                                                                                                |
| Templ compile                    | `*_templ.go` via `make build`                                                                               | ✓         | Required after any `fieldsets.templ` edit.                                                                         |
| Template source                  | `internal/template/templates/.moai/config/sections/*.yaml`                                                  | ✓         | 28 files at plan-phase (see §B).                                                                                   |
| `front-launch` identifier        | `grep -rn 'front-launch\|front_launch\|FrontLaunch' internal/`                                              | ✓         | **Phantom — no matches.** Removal is a no-op (D2).                                                                 |

---

## §B. Live template-structure findings (the iter-1 auditor's key inventory, formalized)

The iter-1 auditor flagged that the v0.1.0 §E blocks were ~20-line paraphrased subsets that missed many live keys. This section records the live structure of each section file so the run-phase agent can honor the spec.md §E preserve-all-other-keys contract.

### §B.1 harness.yaml (≈8165B, 184 lines)

Live top-level + nested keys (beyond the §E.1 keys-of-interest):

- `harness.default_profile`, `harness.evaluator.memory_scope` (FROZEN per_iteration, `@MX:NOTE`)
- `harness.mode_defaults.{solo, team, cg}`
- `harness.auto_detection.{enabled, rules.{minimal, standard, thorough}.conditions[]}`
- `harness.escalation.{enabled, triggers[], max_escalations}`
- `harness.effort_mapping.{minimal, standard, thorough}` — **live values `{medium, high, xhigh}`**; baked values `{low, medium, high}` (spec.md §D.Δ deliberate downgrade)
- `harness.plan_audit_global.{always_enabled, enforce_gate_on_spec_creation, rationale}` — the auditor's "plan_audit / always_enabled" finding
- `harness.levels.{minimal, standard, thorough}` with `{description, skip_phases[], evaluator, evaluator_mode, sprint_contract, plan_audit.{enabled, max_iterations, require_must_pass}, ...}`
- `harness.model_upgrade_review.{enabled, checklist[].{id, question, action, affects}, trigger, output}` — the auditor's "question / affects" finding (these are `checklist` entry sub-keys, not top-level)
- `learning.{enabled, auto_apply, tier_thresholds[], rate_limit.{max_per_week, cooldown_hours}, log_retention_days}`

**Run-phase obligation**: preserve ALL of the above verbatim; overwrite ONLY `effort_mapping.{minimal,standard,thorough}` (to `{low,medium,high}`) and the §E.1 keys-of-interest.

### §B.2 llm.yaml (52 lines)

Live structure (the v0.1.0 §E.5 had a FLAT `glm.{high,medium,low,fable}` — wrong):

- `llm.mode: ""`, `llm.team_mode: ""`, `llm.glm_env_var: "GLM_API_KEY"`
- `llm.performance_tier: "medium"`, `llm.plan_type: "subscription"`
- `llm.claude_models.{high: opus, medium: sonnet, low: sonnet}`
- `llm.glm.base_url: "https://api.z.ai/api/anthropic"`
- `llm.glm.models.{high: glm-5.2, medium: glm-4.7, low: glm-4.5-air, fable: glm-5.2}`
- (commented) `llm.glm.context_windows` override block

**Run-phase obligation**: preserve ALL keys; overwrite ONLY `glm.models.{high,medium,low,fable}` values if the baked values differ (they do not — baked matches live for these four); `mode`/`team_mode` bake as empty string (matches live).

### §B.3 workflow.yaml (132 lines)

Live `execution_mode: team` at line 20 — **baked value `auto`** (spec.md §D.Δ deliberate default-change).

Additional live keys beyond §E.3 keys-of-interest:

- `workflow.default_mode: ""`
- `workflow.agentic_loop.max_iterations: 10`
- `workflow.auto_clear.{after_plan, after_run, enabled, token_threshold}`
- `workflow.loop_prevention.{failure_pattern_detection, max_iterations, max_retries_per_operation}`
- `workflow.worktree.{auto_create, auto_merge, auto_cleanup, tmux_preferred, session_name_pattern}` (+ commented `sparse_paths`)
- `workflow.token_budget.{plan, run, sync}`
- `workflow.workflow_agents` — 7-purpose taxonomy `{read-only-extract, mechanical-transform, synthesize, research, verify-judge, implement, design-architecture}` each `{model, effort}`
- `workflow.model_routing` — LEGACY flat 12-cell (`{S,M,L}-{plan,run,sync,mx}`)
- `workflow.model_routing_profiles` — 3 perfTiers (`max, medium, low`) × 12 cells = 36 entries

**Run-phase obligation**: preserve ALL keys; overwrite ONLY `execution_mode` (to `auto`) and the §E.3 keys-of-interest whose baked values differ.

### §B.4 security.yaml (50 lines)

Live `permission.strict_mode: false` — **baked value `true`** (spec.md §D.Δ deliberate flip).

Additional live keys: `extra_dangerous_bash_patterns[]`, `extra_deny_patterns[]`, `extra_ask_patterns[]`, `extra_sensitive_content_patterns[]`, `permission.{pre_allowlist, session_rules}`, `sandbox.{required, network_allowlist, env_scrub_extra, docker_image}`.

### §B.5 git-strategy.yaml.tmpl (114 lines)

Template-variable rendered (`{{.GitMode}}`, `{{.GitProvider}}`, etc.). Live `merge_method: squash` (lowercase) and `pre_push: warn` for all three modes — **baked values `Squash` (capitalized) and `enforce` for personal/team** (spec.md §D.Δ).

Additional live keys per mode: `workflow`, `environment`, `github_integration`, `push_to_remote`, `branch_creation.{prompt_always, auto_enabled}`, `automation.{auto_branch, auto_commit, auto_pr, auto_push}`, `hooks.{pre_commit, pre_push, commit_msg}`, `commit_style.{format, scope_required}`, plus team-only `draft_pr, required_reviews, branch_protection, main_branch, branch_prefix`.

### §B.6 cache.yaml (17 lines)

Live `cacheStrategy.enabled: false` — **baked value `true`** (spec.md §D.Δ deliberate flip). `session_ttl: "1h"` matches.

### §B.7 handoff.yaml (15 lines)

`handoff.{mode: manual, guide: false}` — added to §E.8 in v0.2.0 (was missing in v0.1.0).

### §B.8 quality.yaml.tmpl (240 lines)

The `quality_extras` concept has NO separate file — its keys live inside `quality.yaml.tmpl`. The file's top-level key is `constitution:` (NOT `quality:`), plus `report_generation:` and `lsp_state_tracking:`. Sub-blocks include `constitution.{development_mode, session_effort_default, enforce_quality, test_coverage_target, ddd_settings, tdd_settings, coverage_exemptions, test_quality, lsp_quality_gates, ast_grep_gate, principles, lsp_integration, memory_guard}`. The main `SectionQuality` (D12 — UNAFFECTED) reads this file; the removed `quality_extras` tab exposed a subset (likely `coverage_exemptions` + `test_quality` + `report_generation` + `lsp_state_tracking`).

### §B.9 project.yaml.tmpl / ralph.yaml / feedback.yaml / observability.yaml / mx.yaml

`project.yaml.tmpl` exists; `cat` at run-phase and bake verbatim. `ralph.yaml`, `feedback.yaml`, `observability.yaml`, `mx.yaml` likewise — each bakes as-is per spec.md §E.7 / §E.9.

---

## §C. The 20-agent effort landscape (motivated Option A)

Surveyed at plan-phase by `grep -m1 -E '^effort:'` across all 20 agent files:

| Agent                                | Actual `effort` | M1.2 tier | Match?  |
|--------------------------------------|-----------------|-----------|---------|
| `manager-spec`                       | `xhigh`         | 🔴        | ✓       |
| `plan-auditor`                       | `xhigh`         | 🔴        | ✓       |
| `super-advisor`                      | `xhigh`         | 🔴        | ✓       |
| `sync-auditor`                       | `xhigh`         | 🔴        | ✓       |
| `manager-develop`                    | `xhigh`         | 🟠        | ✗ (mismatch) |
| `manager-design`                     | `xhigh`         | 🟠        | ✗ (mismatch) |
| `builder-harness`                    | `high`          | 🟠        | ✓       |
| `e2e-tester`                     | `high`          | 🟠        | ✓       |
| `manager-docs`                       | `medium`        | 🔵        | ✓       |
| `manager-git`                        | `low`           | 🔵        | ✗ (mismatch) |
| `quality-specialist`                 | `high`          | 🔵        | ✗ (mismatch) |
| `cli-template-specialist`            | `high`          | 🔵        | ✗ (mismatch) |
| `workflow-specialist`                | `high`          | 🔵        | ✗ (mismatch) |
| `hns-github-specialist`              | `high`          | 🩵        | ✗ (mismatch) |
| `hns-oss-docs-content-author-specialist` | **ABSENT**   | 🩵        | n/a (absent) |
| `hns-oss-docs-locale-translator-specialist` | **ABSENT** | 🩵      | n/a (absent) |
| `hns-oss-docs-structure-curator-specialist` | **ABSENT** | 🩵      | n/a (absent) |
| `hns-release-specialist`             | `high`          | 🩵        | ✗ (mismatch) |
| `hns-release-update-specialist`      | `high`          | 🩵        | ✗ (mismatch) |
| `hook-ci-specialist`                 | `high`          | 🩵        | ✗ (mismatch) |

**Tally**: 10 effort-value mismatches + 3 absent-effort agents = **13 discrepancies** (the iter-1 auditor's "13 M1.2-vs-actual mismatches").

### §C.1 Why the 13 discrepancies are EXPECTED and IRRELEVANT under Option A

Under Option A (name-keyed lookup table), the tier is chosen by reasoning role and is INDEPENDENT of the live effort. The discrepancies are not defects to fix — they are the natural state of a system where:

- The reasoning role (tier) is a stable classification.
- The current effort is a mutable spawn-time value that may have drifted from the role's suggested effort for any number of reasons (cost-reduction sweep, temporary boost for a hard task, historical accident).

The badge shows the role-based tier (stable, informative). The effort selector shows the live effort (mutable, actionable). The user can, if they want, click the tier color to "snap" effort back to the tier's suggested value — but this is an explicit action, not an automatic one.

This is the core argument that won Option A the iter-1 design fork: the 13 discrepancies are a FEATURE (the system surfaces the drift between role and current setting), not a bug. Under the rejected effort-derivation alternative, the 13 discrepancies would each require a 20-agent-file rewrite to eliminate — a large behavioral side-effect for a display-only feature.

### §C.2 The 3 absent-effort agents (Option A eliminates the circular dependency)

`hns-oss-docs-content-author-specialist`, `hns-oss-docs-locale-translator-specialist`, and `hns-oss-docs-structure-curator-specialist` carry NO `effort:` frontmatter line. Under effort-derivation they would have no badge (or require a special-case fallback). Under Option A they render 🩵 directly from the name table — no special case, no fallback, no circular dependency. This is the second core argument for Option A.

---

## §D. Cross-References

- `spec.md` §B (REQ-WC-006..009) — tier requirements; §D.Δ — deliberate default-changes; §E — keys-of-interest.
- `design.md` §B–§G — Option A data-model design + the rejected effort-derivation alternative (this section's §C.1 + §C.2 are the evidence base for design.md §G).
- `plan.md` §A.2 — codebase map (this file §A is the formalized superset); §F M1.2 — 20-agent table.
- `acceptance.md` §C GWT-5 / GWT-16 / §D AC-WC-005 / AC-WC-016 / AC-WC-021 — the ACs that assert the name-keyed behavior.
