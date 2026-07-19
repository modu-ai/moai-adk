---
id: SPEC-MODEL-PROFILE-MATRIX-001
title: "Per-Agent Model+Effort Profile Matrix (replace plan_type axis)"
version: "0.1.1"
status: draft
created: 2026-07-20
updated: 2026-07-20
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "internal/config, internal/template, internal/cli, internal/web"
lifecycle: spec-anchored
tags: "model-routing, config, llm, profile, init, web-console, glm, migration"
related_specs: [SPEC-MODEL-TIER-PLANTYPE-001, SPEC-WEBCONF-SIMPLIFY-001, SPEC-GLM-EFFORT-TUNE-001]
tier: L
---

# SPEC-MODEL-PROFILE-MATRIX-001 — Per-Agent Model+Effort Profile Matrix

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-07-20 | manager-spec | Initial plan-phase authoring. Replace the `plan_type` (api/subscription) billing axis with a single per-agent-group model+effort **profile** matrix (max/medium/low). Retire `llm.plan_type` + `llm.claude_models` + the 66-cell `tierProfiles` frontmatter-mutation, moving consumption to a runtime-arg spawn injection channel consistent with the super-advisor pattern. |
| 0.1.1 | 2026-07-20 | manager-spec | Plan-audit iter-1 revision (FAIL 0.69 → target iter-2 audit-ready). D1: rewrote REQ-MPM-025/026/027 + AC-MPM-016/017 to model-only spawn injection (per-spawn `effort` is NOT accepted for named subagents; Matrix effort is documented intent consumed by frontmatter default + GLM overlay + Workflow/`Agent(general-purpose)` prompt). D2: enumerated all 4 `ApplyTierProfile` production call sites and corrected the wrong "web is display-only" premise (the web settings-save path mutates frontmatter today) — added REQ-MPM-040 + AC-MPM-025. D3: authored design.md + research.md (Tier L 5-file set). D4: default profile = `medium` confirmed; reconciled `DefaultModelPolicy=high(→max)` vs config/template `medium`. Both `[NEEDS CLARIFICATION]` markers resolved into DECISION-001/002. |

---

## §A. Context and Motivation

### §A.1 Problem

The current model-routing design (SPEC-MODEL-TIER-PLANTYPE-001) uses **two** orthogonal axes to pick each agent's model and effort:

1. `llm.plan_type` ∈ {api, subscription} — a billing-context outer dimension.
2. `llm.performance_tier` ∈ {max, medium, low} — a quality/cost tier.

The cross product is a **66-cell** `tierProfiles` map (11 agents × 3 tiers × 2 plan_types) hard-coded in `internal/template/model_policy.go`, applied by `ApplyTierProfile` which **mutates each shipped agent file's `model:` and `effort:` frontmatter** at `moai init` / `moai update` time.

Two problems motivate this SPEC:

- **The billing axis is confusing and low-value.** `plan_type` is no longer an interactive init question (it defaults silently to `subscription`) and its web selector was already removed (SPEC-WEBCONF-SIMPLIFY-001 — `llm.plan_type` is now a read-only display value). The 2× multiplication of the matrix carries maintenance cost for a dimension users rarely set.
- **Frontmatter mutation conflicts with the `[1m]` inherit-by-default convention.** `ApplyTierProfile` rewrites agent `model:` from `inherit` to a concrete alias (`opus`/`sonnet`/`fable`). A concrete frontmatter `model:` pin does NOT inherit the parent session's `[1m]` context entitlement and can fail spawn (`.claude/rules/moai/development/model-policy.md` § Inherit-by-Default Convention, Anthropic issues #45847/#51060). The per-spawn model **runtime arg** is `[1m]`-safe; the frontmatter pin is not.

### §A.2 Goal

Replace the two-axis (plan_type × tier) design with a **single per-agent-group model+effort profile matrix** (3 profiles: max / medium / low), and move consumption from frontmatter mutation to a **runtime-arg spawn injection** channel (the super-advisor pattern), leaving agent frontmatter at its lint-canonical `model: inherit` + doc-canonical effort.

### §A.3 User-confirmed Matrix A (settled design input — MUST NOT be re-derived)

Six agent GROUPS × three profiles. `git` and `docs` rows are profile-invariant.

| Profile | spec/auditors (`manager-spec`, `plan-auditor`, `sync-auditor`) | develop (`manager-develop`) | advisor (`super-advisor`) | design/harness/e2e (`manager-design`, `builder-harness`, `e2e-tester`) | docs (`manager-docs`) | git (`manager-git`) |
|---|---|---|---|---|---|---|
| **max** | fable + medium | fable + low | fable + medium | opus + high | sonnet + medium | sonnet + low |
| **medium** | opus + high | opus + high | fable + low | opus + medium | sonnet + medium | sonnet + low |
| **low** | opus + low | opus + medium | opus + high | opus + low | sonnet + medium | sonnet + low |

`Explore` (11th agent, Anthropic built-in) has no group row — it stays `model: inherit` and is never injected.

Effort tokens above are already canonical enum values (`low`/`medium`/`high`) — no remapping from a user vocabulary is required. Matrix A intentionally does NOT use `xhigh`/`max` effort for any cell; §B.5 governs how this coexists with the doc-canonical frontmatter effort.

### §A.4 Verification note (investigation findings that shape scope)

> All facts below were verified against the working tree during plan-audit iter-1 revision (commands + observed output recorded in `research.md`). The two `[NEEDS CLARIFICATION]` markers that previously lived in plan.md §E are now settled as **DECISION-001** (Agent-tool per-spawn capability) and **DECISION-002** (default profile) — see plan.md §E.

- The task brief referenced an `agentlint LR-12` rule expecting `xhigh` frontmatter for reasoning agents. Verified against the tree: there is **no `internal/agentlint/` package and no effort-frontmatter lint** (`ls internal/agentlint/` → empty; `grep effort internal/agentlint` → no match). The effort-frontmatter canonical is the **documentation** at `.claude/rules/moai/development/agent-authoring.md § Effort-Level Calibration Matrix` (manager-spec: `xhigh`, etc.), NOT a mechanical gate. The only mechanical agent-frontmatter lint touching model routing is `HaikuResidualRule` (`internal/spec/lint_haiku_residual.go`), which enforces zero `haiku` references — Matrix A uses no haiku, so it remains satisfied.
- `internal/statusline/` does **not** read `plan_type` / `performance_tier` / `claude_models` (grep → no match); the statusline is NOT an affected reader.
- **Agent-tool per-spawn injection capability (DECISION-001).** The Claude Code Agent/Task tool accepts a per-spawn **`model`** runtime arg — proven by the super-advisor pattern (`.claude/rules/moai/development/model-policy.md § Inherit-by-Default Convention`: a per-spawn `Agent(model: "opus")` is a runtime arg *distinct* from the frontmatter `model:` pin, and is `[1m]`-safe because it inherits the parent entitlement). It does **NOT** accept a per-spawn **`effort`** override for a NAMED retained subagent — a named subagent's effort is fixed by its own frontmatter (session-scoped), the `ultrathink` keyword, or the environment, never a per-spawn tool arg. Consequence: the profile can inject **model** at spawn, but the profile's **effort** value is *documented intent* for named subagents (consumed by the frontmatter default, the GLM overlay, and Workflow-script / `Agent(general-purpose)` prompt-level steering — §B.5, REQ-MPM-027), NOT a spawn-time override.
- **The web is NOT display-only for frontmatter (corrects the iter-1 premise).** The web `plan_type` *UI write selector* was removed (SPEC-WEBCONF-SIMPLIFY-001) and `plan_type` survives in the UI only as the read-only `ActivePlanType` display field (`internal/web/schemaform.go`, `internal/web/handlers.go`). **However**, the web *settings-save* path `applyPerfTierEdits` (`internal/web/agentfm.go:84-112`) still calls `template.ApplyTierProfile` at `agentfm.go:108`, which **mutates the shipped agent `.md` frontmatter** whenever `performance_tier` changes on save. The plan-audit iter-1 assumption that "the web change is display-only" was wrong; this save-path mutation is a fourth `ApplyTierProfile` call site that MUST be retired (REQ-MPM-040, AC-MPM-025).
- **`ApplyTierProfile` has exactly FOUR production call sites** (verified: `grep -rn ApplyTierProfile internal/ --include='*.go' | grep -v _test.go`; definition at `internal/template/model_policy.go:540`): (1) `internal/core/project/initializer.go:195` (`moai init`), (2) `internal/cli/update.go:486` (`moai update` primary), (3) `internal/cli/update.go:1447` (`moai update` secondary path), (4) `internal/web/agentfm.go:108` (web settings-save). All four unconditionally mutate agent frontmatter `model:` + `effort:` and MUST be retired (REQ-MPM-024 covers init/update; REQ-MPM-040 covers the web save path).
- **Default-value discrepancy (DECISION-002).** Three "default" values coexist in the tree: the config-layer `DefaultPerformanceTier = "medium"` (`internal/config/defaults.go:77`), the template `performance_tier: "medium"` (`internal/template/templates/.moai/config/sections/llm.yaml`), and the template-context init-selection constant `DefaultModelPolicy = ModelPolicyHigh` (`internal/template/model_policy.go:27`, projecting `high → max` via `MapModelPolicyToTier`). The new `llm.profile` default is **`medium`** (aligned with the config + template defaults, user-confirmed); the legacy `DefaultModelPolicy = "high"` constant is a *separate* init-selection default whose `high → max` projection is preserved unchanged **only** for the legacy read-time alias path (REQ-MPM-002).

---

## §B. Requirements (GEARS)

### §B.1 Config schema evolution + migration

**REQ-MPM-001** — The `llm` configuration section **shall** carry a `profile` field whose value is one of the closed set {`max`, `medium`, `low`}, selecting the active per-agent model+effort profile column.

**REQ-MPM-002** — **Where** `llm.profile` is absent or empty, the system **shall** resolve the effective profile from the legacy `llm.performance_tier` value when present (`max`/`medium`/`low` pass through; legacy `high` maps to `max`), and **shall** default to `medium` when neither field is present. The new-default `medium` (DECISION-002) reconciles with the pre-existing tree state as follows: it matches `DefaultPerformanceTier = "medium"` (`internal/config/defaults.go:77`) and the template `performance_tier: "medium"`; the divergent legacy constant `DefaultModelPolicy = ModelPolicyHigh` (`internal/template/model_policy.go:27`) is an init-selection default **distinct** from the profile default — its `high → max` projection (`MapModelPolicyToTier`) is retained **only** for the legacy `performance_tier: high → max` read-time alias above and **shall not** set the new profile default.

**REQ-MPM-003** — **When** the config loader reads an `llm.plan_type` key, the loader **shall** ignore its value without error (silent strip), preserving backward-compatible load of any existing config that still carries `plan_type`.

**REQ-MPM-004** — **When** the config loader reads an `llm.claude_models` block, the loader **shall** ignore it without error, because the per-group profile matrix supersedes the tier→model map.

**REQ-MPM-005** — **When** a persistence write updates the `llm` section (init/update/web save), the writer **shall** remove the `plan_type` and `claude_models` keys from the written `llm.yaml` (write-time removal of retired fields).

**REQ-MPM-006** — The `llm` section **shall** carry an optional `agent_overrides` map keyed by canonical agent name, each value a `{model, effort}` pair, applied on top of the active profile's group assignment for that agent.

**REQ-MPM-007** — **When** an `agent_overrides` entry names an agent not in the retained catalog, or carries a model/effort value outside the valid enums, the validator **shall** report a validation error naming the offending agent and field.

**REQ-MPM-008** — The `llm.profile` value **shall** be closed-set validated; **when** a non-empty out-of-set value is read, the validator **shall** return a validation error naming the offending value and the closed set {max, medium, low}.

### §B.2 Profile matrix (Matrix A) representation

**REQ-MPM-009** — The system **shall** hold a canonical default profile matrix (Matrix A of §A.3) as a Go-code single source of truth, so an absent or partial config still resolves every agent's `{model, effort}` deterministically.

**REQ-MPM-010** — The shipped template `llm.yaml` **shall** mirror the Matrix A default profile matrix under an `llm.profiles` block (six agent-group keys per profile) for transparency and user editability, and the Go default **shall** be the authoritative fallback for any cell absent from config.

**REQ-MPM-011** — The system **shall** hold the agent-GROUP → agent-name membership (spec_auditors, develop, advisor, design_harness_e2e, docs, git) as a Go-code single source of truth, so a group's model+effort resolves for every member agent.

**REQ-MPM-012** — **When** resolving an agent's effective `{model, effort}`, the resolver **shall** apply precedence: `agent_overrides[agent]` (if present) wins; else the active profile's group cell; else the Go-default group cell.

**REQ-MPM-013** — **When** the resolver is queried for `Explore` or any agent with no group membership, it **shall** return the `inherit` sentinel and **shall not** emit a concrete model.

### §B.3 Selection surface — `moai init`

**REQ-MPM-014** — **When** `moai init` runs interactively, the init wizard **shall** present exactly ONE model-routing question — the profile selection {max, medium, low} — and **shall not** present a `plan_type` (api/subscription) question.

**REQ-MPM-015** — The `moai init` CLI **shall** accept a `--profile <max|medium|low>` flag that selects the profile non-interactively and takes precedence over the wizard answer; the flag value **shall** be closed-set validated with an error naming the valid set.

**REQ-MPM-016** — **When** `moai init` completes, the resolved profile **shall** be persisted to `llm.profile` in the project's `llm.yaml`.

**REQ-MPM-017** — The `moai init` and `moai update` CLIs **shall not** expose a `--plan-type` flag; **when** a retired `--plan-type` flag is passed, the CLI **shall** either reject it with an unknown-flag error or ignore it, and **shall not** write a `plan_type` key.

### §B.4 Selection surface — `moai web` console

**REQ-MPM-018** — The `moai web` Model Policy console **shall** present a profile selector {max, medium, low} whose selection persists to `llm.profile`.

**REQ-MPM-019** — The `moai web` Model Policy console **shall not** display or offer a `plan_type` (api/subscription) selector or read-only `plan_type` value.

**REQ-MPM-020** — The `moai web` Model Policy console **shall** render, per retained agent, the resolved `{model, effort}` under the active profile (derived from the single profile-matrix source, not a re-declared literal).

**REQ-MPM-021** — The `moai web` Model Policy console **shall** offer optional per-agent override editing whose saved values persist to `llm.agent_overrides` in `llm.yaml`.

**REQ-MPM-022** — **When** a per-agent override submitted to the web console carries a model or effort value outside the valid enums, the handler **shall** reject the submission with a client error and **shall not** persist the invalid value.

**REQ-MPM-040** — The `moai web` settings-save path **shall not** mutate agent-file `model:` or `effort:` frontmatter. The `applyPerfTierEdits` helper (`internal/web/agentfm.go:84-112`) currently calls `template.ApplyTierProfile` at `agentfm.go:108` on every `performance_tier` change, mutating the shipped agent `.md` frontmatter — this call (the fourth `ApplyTierProfile` production call site, §A.4) **shall** be retired. **When** a web settings save updates the model-routing selection, the writer **shall** persist the resolved `llm.profile` (and any `llm.agent_overrides`) to `llm.yaml` only, leaving agent frontmatter at `model: inherit` + its doc-canonical effort. This corrects the plan-audit iter-1 "web is display-only" premise, which was factually wrong for the save path.

**REQ-MPM-023** — The shipped agent files under `.claude/agents/moai/` **shall** carry lint-canonical frontmatter (`model: inherit` + the per-agent default effort of `agent-authoring.md § Effort-Level Calibration Matrix`), and this frontmatter **shall** be the static default used when no profile injection occurs.

**REQ-MPM-024** — The system **shall not** mutate agent-file `model:` or `effort:` frontmatter during `moai init` (`initializer.go:195`) or `moai update` (`update.go:486` + `update.go:1447`); the former `ApplyTierProfile` frontmatter-mutation pass **shall** be retired at these three call sites (the fourth, the web settings-save path, is REQ-MPM-040).

**REQ-MPM-025** — The active profile's resolved per-agent `{model, effort}` **shall** be readable by the orchestrator through a read-only resolver surface (a `moai model profile [--json]` accessor and/or the `model-policy.md` spawn-guidance rule). The **model** value **shall** be the runtime arg the orchestrator injects at spawn (per-spawn `model`, DECISION-001); the **effort** value **shall** be emitted for display, GLM-overlay input, and documented-intent purposes (§B.5 REQ-MPM-027) — it is NOT a per-spawn override for a named subagent. In all cases the resolver reads the profile matrix, never a mutated frontmatter pin.

**REQ-MPM-026** — The **model** runtime injection channel **shall** follow the super-advisor per-spawn pattern (the model alias is supplied as a per-spawn runtime arg, `[1m]`-safe, distinct from the frontmatter `model:` field), so a profile change never re-introduces the concrete-frontmatter-`model:` spawn-failure risk (`model-policy.md § Inherit-by-Default Convention`). Per DECISION-001, the Agent/Task tool accepts a per-spawn `model` arg but does **not** accept a per-spawn `effort` arg for a named retained subagent; the injection channel therefore carries model only.

**REQ-MPM-027** — For a NAMED retained subagent, the profile's effort value **shall** be treated as **documented intent**, NOT a spawn-time override (DECISION-001). The effort value **shall** be consumed only through channels that exist: (a) the agent-frontmatter effort default (the session-scoped effort, owned by `agent-authoring.md § Effort-Level Calibration Matrix`, unchanged by this SPEC); (b) the GLM effort overlay (`CollapseClaudeEffortToGLM` — the one runtime effort consumer today); and (c) Workflow-script `agent()` calls and `Agent(general-purpose)` per-spawn prompts, where effort steering is prompt-level. **Where** the profile effort differs from the frontmatter doc-canonical effort (e.g. profile `fable + medium` vs frontmatter `xhigh` for `manager-spec`), the divergence **shall** be recorded as intentional and **shall not** be treated as a config or lint error; the frontmatter effort remains the effective effort for the named-subagent spawn until per-spawn effort injection for named subagents is supported by the runtime.

**REQ-MPM-028** — The retired 66-cell `tierProfiles` map and its `plan_type`-keyed accessors **shall** be removed from all four production call sites (§A.4: `initializer.go`, `update.go` ×2, `web/agentfm.go`), and no reader **shall** resolve model/effort through `plan_type` after this SPEC.

### §B.6 GLM backend interaction

**REQ-MPM-029** — **Where** the session backend is GLM (`IsGLMBackend` true), the profile's model alias **shall** map through `llm.glm.models` (fable/opus/sonnet → the configured GLM ids) and the profile's effort **shall** collapse through `CollapseClaudeEffortToGLM` (+ the existing `manager-develop` coding-max override), so a GLM session honors the selected profile without a separate GLM profile axis.

**REQ-MPM-030** — The profile change **shall not** alter the GLM environment activation flow (`moai glm` / `moai cg` env writes, `team_mode` detection) — the overlay remaps only the effort representation and the model alias lookup, exactly as the existing overlay does.

**REQ-MPM-031** — The system **shall not** introduce a GLM-specific profile (max/medium/low remain backend-neutral); GLM interaction is an overlay on the same profile, not a fourth column.

### §B.7 Template mirror, build, tests, docs

**REQ-MPM-032** — Every config/schema/template change **shall** be made in `internal/template/templates/.moai/config/sections/llm.yaml` FIRST, and `make build` **shall** be run so the embedded template reflects the new schema (Template-First rule).

**REQ-MPM-033** — The local `.moai/config/sections/llm.yaml` and the template `llm.yaml` **shall** both carry the new schema (profile + profiles + no plan_type/claude_models), with the local copy remaining a rendered result of the template default.

**REQ-MPM-034** — The system **shall** carry tests covering: config load of a new-schema `llm.yaml`; migration load of a legacy `llm.yaml` (plan_type + claude_models + performance_tier present) without error; profile→per-agent resolution including override precedence; and a round-trip write that strips the retired keys.

**REQ-MPM-035** — The system **shall** carry tests covering the init profile selection (flag + wizard precedence) and the web console profile selector + per-agent override persist/validate paths.

**REQ-MPM-036** — The system **shall** produce a documentation-impact list (README profile-count references, docs-site model-routing pages, `model-policy.md`) enumerating the surfaces that describe the retired plan_type axis, flagged as a follow-up sync-phase scope (NOT authored in this SPEC's run phase).

### §B.8 Boundary requirements (unwanted behavior)

**REQ-MPM-037** — The system **shall not** modify agent-frontmatter effort *values* as a design deliverable — the per-agent doc-canonical effort stays owned by `agent-authoring.md § Effort-Level Calibration Matrix`; this SPEC only stops the mutation pass and injects at runtime.

**REQ-MPM-038** — The system **shall not** modify `delegation.yaml`, the GLM env-writing code paths, or the `model_routing_profiles` (Tier×Phase) map in `workflow.yaml` — the profile axis is per-agent, distinct from the Tier×Phase spawn-routing axis.

**REQ-MPM-039** — The system **shall not** claim live GLM wire-effectiveness for profile-injected effort; the existing "implemented+wired, live-validation pending" honesty constraint (inherited from SPEC-MODEL-TIER-PLANTYPE-001) **shall** carry forward unchanged.

---

## §C. Success Criteria

- A fresh `moai init --profile medium` produces an `llm.yaml` with `profile: medium`, no `plan_type` key, no `claude_models` block, and unmutated agent frontmatter (`model: inherit`).
- A legacy `llm.yaml` (carrying `plan_type: subscription` + `performance_tier: max` + `claude_models`) loads without error and resolves to profile `max`.
- The web Model Policy console shows a profile selector + per-agent resolved model+effort + optional override editing, with zero `plan_type` surface, and a web settings save persists only to `llm.yaml` — leaving agent `.md` frontmatter unmutated (`model: inherit`).
- The orchestrator can obtain the resolved per-agent `{model, effort}` for the active profile from a read-only resolver surface without reading agent frontmatter; the resolved **model** is the per-spawn runtime arg, while the resolved **effort** is documented intent (display / GLM overlay / Workflow prompt) for named subagents.
- `HaikuResidualRule` and all existing config CI guards remain green.

---

## §D. Exclusions

### Out of Scope — agent frontmatter effort values
- Changing the per-agent default effort *values* in `.claude/agents/moai/*.md` frontmatter. Those stay owned by `agent-authoring.md § Effort-Level Calibration Matrix`. This SPEC retires the *mutation pass* and adds *runtime injection*; it does not re-author the canonical frontmatter effort table.

### Out of Scope — Tier×Phase spawn-routing map
- `model_routing_profiles` in `workflow.yaml` and the `RouteModelFor(specTier, phase, perfTier)` accessor. That is a separate per-(Tier×Phase) cost axis keyed by the same `max/medium/low` token; the profile rename is compatible and requires no change to it.

### Out of Scope — GLM environment flow
- The `moai glm` / `moai cg` environment activation, `team_mode` persistence, and `IsGLMBackend` detection predicate. The profile is consumed by the existing GLM effort overlay unchanged; no new GLM env behavior.

### Out of Scope — live GLM wire-effectiveness validation
- Confirming that z.ai actually consumes the collapsed `reasoning_effort` for profile-injected effort. This inherits the "implemented+wired, live-validation pending" debt from SPEC-MODEL-TIER-PLANTYPE-001 and is not resolved here.

### Out of Scope — delegation map
- `.moai/config/sections/delegation.yaml` (per-subcommand agent/skill designation) is unaffected by the model-routing profile change.

### Out of Scope — documentation authoring
- Rewriting README / docs-site model-routing prose. This SPEC produces the impact list (REQ-MPM-036); the actual doc edits are follow-up sync-phase scope.

---

## §E. References

- SPEC-MODEL-TIER-PLANTYPE-001 (superseded axis: plan_type × tier; source of `tierProfiles`, `ApplyTierProfile`, GLM effort overlay).
- SPEC-WEBCONF-SIMPLIFY-001 (removed the web plan_type write selector; name-keyed tier UI).
- SPEC-GLM-EFFORT-TUNE-001 (GLM coding-max override set = {manager-develop}).
- `.claude/rules/moai/development/model-policy.md` (inherit-by-default, `[1m]` bug, 3-tier max/medium/low).
- `.claude/rules/moai/development/agent-authoring.md § Effort-Level Calibration Matrix` (doc-canonical per-agent effort).
- `internal/template/model_policy.go` (`tierProfiles` lines 297-315, `ApplyTierProfile` line 540, `MapModelPolicyToTier` line 385, `DefaultModelPolicy = ModelPolicyHigh` line 27, `modelInherit` line 254).
- `internal/template/glm_effort_overlay.go` (`CollapseClaudeEffortToGLM` line 88, `IsGLMBackend` line 207, coding-max override set `{manager-develop}` lines 102-121).
- `internal/config/{types.go,plan_type.go,validation.go,defaults.go}` (LLMConfig schema, plan_type enum, validation, `DefaultPerformanceTier = "medium"` line 77).
- `ApplyTierProfile` production call sites (4): `internal/core/project/initializer.go:195`, `internal/cli/update.go:486`, `internal/cli/update.go:1447`, `internal/web/agentfm.go:108` (`applyPerfTierEdits`).
- `.claude/rules/moai/development/model-policy.md § Inherit-by-Default Convention` (per-spawn `model` runtime arg is `[1m]`-safe and distinct from the frontmatter pin — DECISION-001 evidence).
