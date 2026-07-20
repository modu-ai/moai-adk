# Design — SPEC-MODEL-PROFILE-MATRIX-001

Architecture for replacing the two-axis (`plan_type` × `performance_tier`) model-routing design with a single per-agent-group **profile** matrix, moving consumption from agent-frontmatter mutation to a runtime-arg **model** injection channel. Grounded in `research.md` (verified investigation); no fabricated references. Sections ordered by decision-reversibility (data model → resolver → injection → web → migration → removal).

## §A. Config schema (data-model decision — highest reversibility)

### §A.1 New `LLMConfig` fields

`internal/config/types.go` `LLMConfig` gains:

```yaml
llm:
  profile: medium            # NEW — closed set {max, medium, low}; default medium (DECISION-002)
  profiles:                  # NEW — Matrix A mirror (transparency + editability); Go default is authoritative fallback
    max:    { spec_auditors: {model: fable,  effort: medium}, develop: {model: fable, effort: low},  advisor: {model: fable,  effort: medium}, design_harness_e2e: {model: opus, effort: high},   docs: {model: sonnet, effort: medium}, git: {model: sonnet, effort: low} }
    medium: { spec_auditors: {model: opus,   effort: high},   develop: {model: opus,  effort: high}, advisor: {model: fable,  effort: low},    design_harness_e2e: {model: opus, effort: medium}, docs: {model: sonnet, effort: medium}, git: {model: sonnet, effort: low} }
    low:    { spec_auditors: {model: opus,   effort: low},    develop: {model: opus,  effort: medium}, advisor: {model: opus,  effort: high},   design_harness_e2e: {model: opus, effort: low},    docs: {model: sonnet, effort: medium}, git: {model: sonnet, effort: low} }
  agent_overrides:           # NEW — optional, keyed by canonical agent name; {model, effort} pair
    manager-develop: { model: opus, effort: xhigh }
  glm:                       # UNCHANGED — GLM model-id map + overlay input
    models: { fable: glm-5.2, opus: glm-5.2, sonnet: glm-5.2 }
  # RETIRED (read-tolerated, write-stripped): plan_type, claude_models
  # LEGACY (read-time alias only): performance_tier
```

Struct fields (canonical yaml tags): `Profile string yaml:"profile"`, `Profiles map[string]map[string]ModelEffort yaml:"profiles"`, `AgentOverrides map[string]ModelEffort yaml:"agent_overrides"`. A `ModelEffort struct { Model string; Effort string }` type carries the pair. The struct↔YAML symmetry guard (`CONFIG_STRUCT_YAML_MISMATCH`) must stay green for the new fields (AC-MPM-024).

### §A.2 Retired keys (migration-tolerant)

- `plan_type` — stripped on read (no error), removed on write (REQ-MPM-003/005).
- `claude_models` — ignored on read, removed on write + from template (REQ-MPM-004/005).
- `performance_tier` — kept as a read-time legacy alias only: `EffectiveProfile()` returns `profile` if present, else maps `performance_tier` (`high → max`, `max/medium/low` pass-through), else `medium` (REQ-MPM-002).

## §B. Matrix A representation (data-model SSOT)

### §B.1 Go-code SSOT

`internal/template/model_policy.go` replaces the 66-cell `tierProfiles` (2 plan_type blocks) with:

- `defaultProfileMatrix()` — the Matrix A base values (3 profiles × 6 agent-group cells), Go-code authoritative fallback for any absent config cell (REQ-MPM-009).
- Group→agent membership SSOT (REQ-MPM-011): `spec_auditors = {manager-spec, plan-auditor, sync-auditor}`, `develop = {manager-develop}`, `advisor = {super-advisor}`, `design_harness_e2e = {manager-design, builder-harness, e2e-tester}`, `docs = {manager-docs}`, `git = {manager-git}`. `Explore` (and any agent with no membership) → `inherit` (REQ-MPM-013, reuses `modelInherit` sentinel at `model_policy.go:254`).

### §B.2 Resolution precedence (REQ-MPM-012)

`ResolveAgentModelEffort(profile, agent, overrides)`:
1. `agent_overrides[agent]` if present → wins.
2. else the active profile's group cell (from config `profiles`, falling back to `defaultProfileMatrix()`).
3. else the Go-default group cell.
4. no group membership → `{inherit, <none>}`.

## §C. Resolver + injection path (interface + runtime behavior)

### §C.1 Read-only resolver surface

New `moai model profile [--json]` accessor (research §C.5: no such command exists yet — this is a NEW surface). It resolves the active profile + overrides → per-agent `{model, effort}`, applies the GLM overlay when `IsGLMBackend(cfg)` is true, and emits the table (human or `--json`). This is the concrete "runtime-arg channel" surface the orchestrator reads (REQ-MPM-025).

Recommended shape: a new `moai model` parent command with a `profile` subcommand (distinct from the existing `moai profile` settings command, research §C.5). Exact command naming is a run-phase detail; the read-only-resolver contract is the binding requirement.

### §C.2 Injection path — MODEL only (DECISION-001)

```
active llm.profile ──► resolver ──► per-agent {model, effort}
                                        │
                    ┌───────────────────┴───────────────────┐
                 MODEL                                    EFFORT
                    │                                        │
        per-spawn Agent(model: <alias>)         documented intent — consumed by:
        runtime arg ([1m]-safe,                 (a) agent frontmatter default (unchanged)
        super-advisor pattern)                  (b) GLM overlay (CollapseClaudeEffortToGLM)
                                                (c) Workflow-script agent() /
                                                    Agent(general-purpose) prompt steering
```

The Agent/Task tool accepts a per-spawn `model` but NOT a per-spawn `effort` for a named subagent (research §B). So the resolver's **model** is the spawn-injected runtime arg; the **effort** is emitted for the web display, GLM-overlay input, and documented intent (REQ-MPM-025/026/027). For a named-subagent Claude spawn, the effective effort remains the frontmatter doc-canonical value; the profile-vs-frontmatter divergence (e.g. profile `medium` vs frontmatter `xhigh`) is intentional, not an error (AC-MPM-017).

### §C.3 Frontmatter left unmutated

Agent `.md` frontmatter stays at the template-source lint-canonical `model: inherit` + doc-canonical effort (research §C.2). No init/update/web pass rewrites it (§E below). This restores inherit-by-default and removes the `[1m]` spawn-fail risk (REQ-MPM-023/024/040).

## §D. Web UI change (user-facing UX)

`internal/web` (`schemaform.go`, `agentfm.go`, `handlers.go`, `fieldsets.templ`):

1. **Remove** the read-only `ActivePlanType` display (REQ-MPM-019).
2. **Add** a profile selector {max, medium, low} → persists `llm.profile` (REQ-MPM-018).
3. **Render** per-agent resolved `{model, effort}` from the single profile-matrix source (no re-declared literal) (REQ-MPM-020).
4. **Add** optional per-agent override editing → `llm.agent_overrides` with enum validation; out-of-enum → client error, persists nothing (REQ-MPM-021/022).
5. **Retire the save-path frontmatter mutation** (REQ-MPM-040, D2): `applyPerfTierEdits` (`agentfm.go:84-112`) currently calls `template.ApplyTierProfile` at `agentfm.go:108` on every perf-tier save — this call is removed; the web save persists selection to `llm.yaml` only, leaving agent frontmatter unmutated.

## §E. Migration flow

```
read legacy llm.yaml (plan_type + claude_models + performance_tier)
      │
      ├─ plan_type      → ignored (no error)                       [REQ-MPM-003]
      ├─ claude_models   → ignored (no error)                       [REQ-MPM-004]
      └─ performance_tier → EffectiveProfile() alias (high→max)     [REQ-MPM-002]
      │
   resolve active profile (profile ?? perf_tier-alias ?? medium)
      │
   any write (init / update / web save)
      │
      └─ emit profile:, strip plan_type + claude_models             [REQ-MPM-005]
```

Legacy `llm.yaml` with `plan_type: subscription` + `performance_tier: max` + `claude_models` loads without error and resolves to profile `max` (AC-MPM-002). Round-trip write strips the retired keys (AC-MPM-003).

## §F. Removal (mechanical — lowest reversibility, do last)

Delete `tierProfiles` (`model_policy.go:297-315`), `ApplyTierProfile` (`model_policy.go:540`) + its 4 call sites (research §A), `ApplyPlanType`, `ResolveProjectPlanType`, `EffectivePlanType`, `validatePlanType`, `PlanType*` constants. Replace with the profile resolver (REQ-MPM-024/028/037/038/040). `MapModelPolicyToTier` (`high→max`) is retained only for the `performance_tier` read-time alias, not for the profile default (DECISION-002).

## §G. Invariants preserved

- `HaikuResidualRule` green — Matrix A has zero haiku (AC-MPM-024).
- `model_routing_profiles` (Tier×Phase, `RouteModelFor`) unmodified — separate axis (REQ-MPM-038).
- GLM env activation flow (`moai glm`/`moai cg`, `team_mode`, `IsGLMBackend`) unmodified — overlay-only (REQ-MPM-030).
- No GLM-specific profile column — backend-neutral (REQ-MPM-031).
- No claim of GLM wire-effectiveness — "implemented + wired, live-validation pending" carried forward (REQ-MPM-039, AC-MPM-023).
