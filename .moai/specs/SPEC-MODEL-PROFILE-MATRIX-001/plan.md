# Implementation Plan — SPEC-MODEL-PROFILE-MATRIX-001

## §A. Context

Replace the two-axis (`plan_type` × `performance_tier`) model-routing design with a single per-agent-group **profile** matrix (max/medium/low), and move consumption from agent-frontmatter mutation (`ApplyTierProfile`) to a runtime-arg spawn injection channel (super-advisor pattern). See spec.md §A for the full motivation and Matrix A.

Milestones are ordered by **decision-reversibility** — the highest-change-likelihood decisions (data model, resolver interface, user-facing selection flows) lead; mechanical removal + docs trail. Human review should focus on M1–M3.

## §B. Known Issues / Investigation Findings (verified against tree)

| Finding | Evidence | Impact on plan |
|---|---|---|
| No `agentlint LR-12` / effort lint exists | `ls internal/agentlint/` empty; `grep effort internal/agentlint` no match | Frontmatter-vs-profile effort divergence is NOT a lint failure (D3 is clean); reconciliation is a doc concern |
| `plan_type` no longer an interactive init question | `internal/cli/wizard/questions.go:23-25` | REQ-MPM-014 ADDS a profile question; there is no plan_type question to "remove" from the wizard, only the `--plan-type` flag + silent default to strip |
| Web `plan_type` **UI** write selector removed, but web save-path is NOT display-only (D2 correction) | UI selector gone (`internal/web/handlers.go:340-341`, `schemaform.go:82 ActivePlanType`) BUT `applyPerfTierEdits` (`internal/web/agentfm.go:84-112`) calls `ApplyTierProfile` at `agentfm.go:108` on every perf-tier save → **mutates agent frontmatter today** | Web change = remove `ActivePlanType` display + add profile selector + overrides **+ retire the `agentfm.go:108` save-path mutation** (REQ-MPM-040). The iter-1 "display-only" premise was wrong. |
| `statusline` does NOT read these fields | `grep PerformanceTier\|ClaudeModels\|PlanType internal/statusline` no match | statusline is NOT an affected reader |
| `ApplyTierProfile` has **4** production call sites (verified `grep -rn ApplyTierProfile internal/ --include='*.go' \| grep -v _test.go`) | (1) `internal/core/project/initializer.go:195` (init) · (2) `internal/cli/update.go:486` (update primary) · (3) `internal/cli/update.go:1447` (update secondary) · (4) `internal/web/agentfm.go:108` (web save) — def at `internal/template/model_policy.go:540` | Retiring all 4 restores `model: inherit` and removes the `[1m]` spawn-fail risk (REQ-MPM-024 = 1-3; REQ-MPM-040 = 4) |
| Agent-tool per-spawn accepts `model` only, NOT `effort`, for named subagents (D1) | super-advisor `Agent(model: "opus")` per-spawn is a runtime arg (`model-policy.md § Inherit-by-Default Convention` ~L97); no per-spawn `effort` arg for a named subagent | Profile injects **model** at spawn; profile **effort** = documented intent (frontmatter default + GLM overlay + Workflow/`Agent(general-purpose)` prompt) — DECISION-001 |
| Default discrepancy: `DefaultModelPolicy=high(→max)` vs config/template `medium` (D2/D4) | `internal/template/model_policy.go:27` (`ModelPolicyHigh`, projects high→max via `MapModelPolicyToTier`) vs `internal/config/defaults.go:77` (`DefaultPerformanceTier="medium"`) vs template `performance_tier: "medium"` | New `llm.profile` default = `medium`; legacy `high→max` alias retained only on read path — DECISION-002 |
| `RouteModelFor` perfTier axis is separate | `internal/config/model_routing.go`, keyed by perfTier not plan_type | model_routing_profiles is OUT OF SCOPE; profile token is compatible |

## §C. Pre-flight (before run-phase)

1. Both former `[NEEDS CLARIFICATION]` markers are already resolved as DECISION-001 / DECISION-002 (§E) — no clarification round is pending.
2. Confirm `development_mode` (quality.yaml) for run-phase cycle_type (ddd vs tdd).
3. `git fetch origin main` + divergence check (shared-checkout race mitigation).

## §D. Key Design Decisions

- **D1 — Config key.** Introduce `llm.profile` (max/medium/low). Keep `performance_tier` as a **read-time legacy alias** only (if `profile` absent, use it; `high`→`max`). Write path emits `profile`, strips `plan_type` + `claude_models`. Default profile = `medium` (matches current template default). *Reversibility: HIGH (schema).*
- **D2 — Matrix storage.** Matrix A base values are a **Go-code SSOT** (`defaultProfileMatrix()` replacing `tierProfiles`), mirrored into the template `llm.yaml` under `llm.profiles:` for transparency. Group→agent membership is a Go-code SSOT. Config `llm.agent_overrides` is the per-agent editable surface. Go default is authoritative fallback for any absent cell — this keeps Matrix A drift-proof while honoring "stored in config". *Reversibility: HIGH (data model).*
- **D3 — Frontmatter reconciliation (the core tension, corrected per DECISION-001).** Agent frontmatter stays lint-canonical: `model: inherit` + the `agent-authoring.md § Effort-Level Calibration Matrix` effort (e.g. `manager-spec: xhigh`). `ApplyTierProfile` frontmatter mutation is **retired** at all 4 call sites. The profile matrix is the **runtime-injection SSOT** for **model**; the orchestrator reads it via a read-only resolver and injects the **model** alias at spawn (super-advisor per-spawn `model` arg — `[1m]`-safe, distinct from the frontmatter pin). The Agent/Task tool does **NOT** accept a per-spawn `effort` arg for a named subagent (DECISION-001), so the profile's **effort** is NOT a spawn-time override — it is **documented intent** consumed by: (a) the frontmatter effort default (session-scoped, unchanged), (b) the GLM effort overlay (`CollapseClaudeEffortToGLM`), and (c) Workflow-script `agent()` / `Agent(general-purpose)` prompt-level effort steering. Where the profile effort (e.g. `medium`) diverges from the frontmatter effort (e.g. `xhigh`), the divergence is intentional documented intent and not an error; the frontmatter effort remains the effective named-subagent spawn effort until the runtime supports per-spawn effort for named subagents. `HaikuResidualRule` stays green (matrix has no haiku). *Reversibility: HIGH (user-facing runtime behavior + interface).*
- **D4 — Resolver interface.** Add a read-only `moai model profile [--json]` accessor that resolves active profile + overrides → per-agent `{model, effort}` (+ GLM overlay applied when `IsGLMBackend`). The orchestrator's spawn guidance (`model-policy.md`) references it. This is the concrete "runtime-arg channel" surface. *Reversibility: MEDIUM (new CLI + doc).*
- **D5 — GLM mapping.** Under `IsGLMBackend`, the resolver maps the group model alias through `llm.glm.models` and collapses effort via `CollapseClaudeEffortToGLM` (+ manager-develop coding-max override). No new GLM axis; overlay unchanged. *Reversibility: MEDIUM.*
- **D6 — Migration.** plan_type: strip on read (no error), remove on write. claude_models: ignore on read, remove on write + from template. performance_tier→profile alias. `--plan-type` flag + wizard silent-default + web read-only display all removed. *Reversibility: MEDIUM (backward-compat).*
- **D7 — Removal.** Delete the 66-cell `tierProfiles`, `ApplyTierProfile`, `ApplyPlanType`, `ResolveProjectPlanType`, `EffectivePlanType`, `validatePlanType`, `PlanType*` constants, and **all 4** `ApplyTierProfile` call sites (`initializer.go:195`, `update.go:486`, `update.go:1447`, `web/agentfm.go:108` — the last inside `applyPerfTierEdits`). Replace with the profile resolver. *Reversibility: LOW (mechanical), do last.*

## §E. Design Decisions (resolved — formerly Open Questions)

Both plan-audit iter-1 `[NEEDS CLARIFICATION]` markers are settled below. No open clarification markers remain; the run-phase may proceed after Implementation Kickoff Approval.

- **DECISION-001 (named-subagent runtime injection capability) — decided 2026-07-20.** *Decision:* the profile injects **model only** at spawn for named retained subagents; the profile **effort** is documented intent (not a per-spawn override). *Rationale:* the Claude Code Agent/Task tool accepts a per-spawn `model` runtime arg (proven `[1m]`-safe by the super-advisor pattern, `model-policy.md § Inherit-by-Default Convention`) but does **not** accept a per-spawn `effort` arg for a NAMED subagent — a named subagent's effort is fixed by its frontmatter (session-scoped), `ultrathink`, or the environment. Therefore option (a) "inject model only + treat frontmatter effort as the effective effort" is adopted; option (b) "re-spawn all retained agents as `Agent(general-purpose)`" is rejected (it would discard the named-agent catalog and its frontmatter defaults for no gain). The resolver still emits effort, consumed by: the frontmatter default, the GLM overlay (`CollapseClaudeEffortToGLM`), and Workflow-script `agent()` / `Agent(general-purpose)` prompt-level steering, plus the web display. The mutation retirement (REQ-MPM-024 + REQ-MPM-040) is therefore complete for **model**; **effort** frontmatter is left unmutated at its doc-canonical value. Reflected in REQ-MPM-025/026/027 and AC-MPM-016/017.
- **DECISION-002 (profile default value) — decided 2026-07-20.** *Decision:* the shipped `llm.profile` default is **`medium`**. *Rationale:* it matches the two config-layer/template defaults already in the tree (`DefaultPerformanceTier = "medium"` at `internal/config/defaults.go:77`; template `performance_tier: "medium"`) and `model-policy.md` names `medium` the default; it is the cost-balanced middle profile. The divergent legacy constant `DefaultModelPolicy = ModelPolicyHigh` (`internal/template/model_policy.go:27`, projecting `high → max`) is a **separate** init-selection default and is NOT used as the profile default — its `high → max` projection is preserved only for the legacy `performance_tier: high → max` read-time alias. Reflected in REQ-MPM-002 and AC-MPM-001.

## §F. Milestones (priority-ordered by reversibility — highest-change-likelihood first)

### M1 — Config schema + Matrix A data model (data-model decision)
- Add `Profile` + `AgentOverrides` fields to `LLMConfig`; retire `PlanType`/`ClaudeModels` from the write path (read-tolerant).
- Replace `tierProfiles` (66-cell) with `defaultProfileMatrix()` (Matrix A) + group-membership SSOT + `ResolveAgentModelEffort(profile, agent, overrides)` with the D2 precedence.
- `EffectiveProfile()` (profile → alias performance_tier → default medium) + `validateProfile` + `validateAgentOverrides`.
- Template `llm.yaml`: add `profile` + `profiles` + `agent_overrides`, remove `plan_type` + `claude_models`; `make build`.
- Covers REQ-MPM-001..013, 032, 033.

### M2 — Runtime resolver + frontmatter reconciliation (interface + runtime behavior)
- Add `moai model profile [--json]` read-only accessor (resolved per-agent `{model, effort}`, GLM overlay applied when backend GLM).
- Update `model-policy.md` spawn-guidance to reference the resolver as the runtime-arg injection source; restore/keep agent frontmatter `model: inherit` + doc-canonical effort.
- Covers REQ-MPM-023, 025, 026, 027, 029, 030, 031.

### M3 — Selection surfaces: init wizard + web console (user-facing UX)
- init: ONE profile question; `--profile` flag; persist `llm.profile`; remove `--plan-type` flag + silent plan_type default.
- web: profile selector (persist `llm.profile`); remove `ActivePlanType` display; per-agent resolved model+effort render (from the single matrix source); optional per-agent override editing → `llm.agent_overrides` with validation; **retire the `applyPerfTierEdits` → `ApplyTierProfile` frontmatter mutation at `agentfm.go:108` (REQ-MPM-040) — the web save persists to `llm.yaml` only**.
- Covers REQ-MPM-014..022, 040.

### M4 — Retire plan_type + ApplyTierProfile (mechanical removal — do last)
- Delete `tierProfiles`, `ApplyTierProfile`, `ApplyPlanType`, `ResolveProjectPlanType`, `EffectivePlanType`, `validatePlanType`, `PlanType*` constants; remove init/update call sites to `ApplyTierProfile`.
- Remove `POST /model-policy/plan-type` remnants if any; remove plan_type test fixtures.
- Covers REQ-MPM-024, 028, 037, 038.

### M5 — Tests + docs impact list
- Config load/migration/round-trip/resolution tests; init flag+wizard tests; web selector+override tests.
- Produce the documentation-impact list (REQ-MPM-036) — README profile-count refs + docs-site model-routing pages + `model-policy.md` — as a run-phase artifact flagged for follow-up sync; do NOT author the doc edits here.
- Covers REQ-MPM-034, 035, 036, 039.

## §G. Anti-Patterns to Avoid

- Re-deriving or "improving" Matrix A cells — they are settled design input (§A.3), copy verbatim.
- Re-introducing a concrete frontmatter `model:` pin (breaks `[1m]` inherit-by-default — the whole point of D3).
- Introducing haiku anywhere in the matrix (HaikuResidualRule fails).
- Treating `model_routing_profiles` (Tier×Phase) as the same axis and editing it — it is separate.
- Claiming GLM wire-effectiveness (REQ-MPM-039 honesty constraint).
- Blind `grep -v` field-name edits — verify each reader (initializer.go, update.go, web/agentfm.go, web/schemaform.go, web/handlers.go) individually before removing plan_type.

## §H. Cross-References

- spec.md §B (REQ-MPM-001..040), acceptance.md (AC-MPM-001..025), progress.md (§E audit signals).
- design.md (architecture: config schema §A, Matrix A SSOT §B, resolver + model-only injection §C, web change §D, migration §E, removal §F).
- research.md (verified investigation: 4 call-site grep §A, Agent-tool capability §B, current schema state §C, GLM overlay §D, non-affected readers §E).
- `internal/config/{types.go,plan_type.go,validation.go,defaults.go}`, `internal/template/{model_policy.go,glm_effort_overlay.go}`, `internal/cli/{init.go,update.go,wizard/}`, `internal/web/{schemaform.go,agentfm.go,handlers.go,fieldsets.templ}`, `internal/core/project/initializer.go`.
- `internal/spec/lint_haiku_residual.go` (must stay green).
- `.claude/rules/moai/development/{model-policy.md,agent-authoring.md}`.
