# Research — SPEC-MODEL-PROFILE-MATRIX-001

Codebase investigation grounding the plan-phase design. Every finding below cites the command run and the observed output (per `verification-claim-integrity.md` §3). Verified against the working tree during plan-audit iter-1 revision (2026-07-20). Line numbers are content-token anchors — they drift; the symbol names are authoritative.

## §A. `ApplyTierProfile` production call-site enumeration (D2)

**Command:** `grep -rn "ApplyTierProfile" internal/ --include='*.go' | grep -v "_test.go"`

**Observed (production call sites only):**

| # | Call site | Path:line | Surface |
|---|-----------|-----------|---------|
| def | function definition | `internal/template/model_policy.go:540` | `func ApplyTierProfile(projectRoot, planType, tier string, mgr manifest.Manager) error` |
| 1 | `moai init` | `internal/core/project/initializer.go:195` | init-time frontmatter apply |
| 2 | `moai update` (primary) | `internal/cli/update.go:486` | update-time apply |
| 3 | `moai update` (secondary) | `internal/cli/update.go:1447` | second update path |
| 4 | `moai web` settings-save | `internal/web/agentfm.go:108` | inside `applyPerfTierEdits` (lines 84-112) |

**Finding:** exactly 4 production call sites. The plan-audit iter-1 plan.md §B listed only 2 (`initializer.go` + `update.go:484`), missing `update.go:1447` and the web save path. All four call `ApplyTierProfile`, which mutates agent `.md` frontmatter `model:` + `effort:` (see `model_policy.go:519` doc comment: "ApplyTierProfile patches each shipped agent file's model: AND effort:").

### §A.1 The web save path is NOT display-only (corrects iter-1 premise)

**Read:** `internal/web/agentfm.go:74-112` (`applyPerfTierEdits`).

**Observed:** the helper's own doc comment (lines 74-83) states it "re-applies the {model, effort} tier profile to the shipped agent .md frontmatter" and, when `performance_tier` changes on save, calls `template.ApplyTierProfile(projectRoot, resolvedPlan, resolvedTier, mfMgr)` at line 108. The `plan_type` UI *write selector* was removed (SPEC-WEBCONF-SIMPLIFY-001) — but the save path still resolves `plan_type` from config (`template.ResolveProjectPlanType`, line 106) and mutates frontmatter. The iter-1 assumption "web change = display-only" was factually wrong for the save path. → Grounds REQ-MPM-040 + AC-MPM-025.

## §B. Agent-tool per-spawn injection capability (D1 / DECISION-001)

**Question:** does the Claude Code Agent/Task tool accept a per-spawn `model` AND a per-spawn `effort` for a NAMED retained subagent?

**Evidence (model — supported):** `grep -n "per-spawn" .claude/rules/moai/development/model-policy.md` →
> "deep-reasoning exceptions use per-spawn `Agent(model: "opus")` only for the 5-10% of tasks where Opus wins … even those inherit the parent `[1m]` entitlement because they are spawned without a frontmatter `model:` pin (the per-spawn `model` parameter is a runtime arg, distinct from the frontmatter field that triggers the bug)." (`model-policy.md` ~L97)

This confirms the super-advisor pattern: a per-spawn `model` is a runtime arg, `[1m]`-safe, distinct from the frontmatter pin.

**Evidence (effort — NOT supported for named subagents):** no rule, agent frontmatter mechanism, or Agent-tool schema in the tree exposes a per-spawn `effort` arg for a named subagent. `agent-authoring.md` documents `effort` **only** as a frontmatter field (session-scoped override): `grep -n "effort" .claude/rules/moai/development/agent-authoring.md` →
> "| effort | No | inherit | Session effort override: low, medium, high, xhigh, max …" and "**effort**: Overrides session effort level for this agent." (agent-authoring.md L61, L83)

Effort for a named subagent is therefore set by (a) its frontmatter, (b) the `ultrathink` keyword, or (c) the environment — never a per-spawn tool arg.

**Orchestrator/user confirmation:** the resolution for this SPEC's iter-1 (recorded in the revision brief) confirms this as final: "the Agent tool supports per-spawn `model` injection ONLY — NO per-spawn effort for named subagents."

**Conclusion:** profile injects **model** at spawn; profile **effort** is documented intent for named subagents (consumed by frontmatter default + GLM overlay + Workflow/`Agent(general-purpose)` prompt). → Grounds DECISION-001, REQ-MPM-025/026/027, AC-MPM-016/017.

## §C. Current schema + frontmatter state (D2 investigation)

### §C.1 `tierProfiles` (the 66-cell map being retired)

**Read:** `internal/template/model_policy.go:297-315`. **Observed:** two blocks keyed by `plan_type` (subscription vs api), each mapping 11 agents × 3 tiers → `{model, effort}`. 11 × 3 × 2 = 66 cells. Sample rows (subscription block): `manager-spec: {{"fable","high"},{"fable","high"},{"opus","high"}}`, `super-advisor: {{"fable","xhigh"},{"fable","high"},{"opus","high"}}`. → confirms the "66-cell" claim in spec.md §A.1.

### §C.2 Template SOURCE vs local (mutated) agent frontmatter

**Command:** `grep -n "^model:\|^effort:" internal/template/templates/.claude/agents/moai/manager-{spec,develop}.md`

**Observed (template source):** `manager-spec` → `model: inherit`, `effort: xhigh`; `manager-develop` → `model: inherit`, `effort: xhigh`.

**Command:** `grep -n "^model:\|^effort:" .claude/agents/moai/manager-{spec,develop}.md`

**Observed (local, deployed):** `manager-spec` → `model: opus`, `effort: xhigh`; `manager-develop` → `model: opus`, `effort: high`.

**Finding:** the **template source** ships `model: inherit` (lint-canonical), but the **deployed local copies** carry concrete `model: opus` (and a mutated `effort` for manager-develop) because `ApplyTierProfile` ran at init/update. This is the exact `[1m]`-inheritance hazard the SPEC targets: a concrete frontmatter `model:` pin does not inherit the parent session's `[1m]` entitlement (`model-policy.md § Inherit-by-Default Convention`, Anthropic #45847/#51060). Retiring the mutation restores `model: inherit` on deploy. → Grounds REQ-MPM-023/024/040 + AC-MPM-015/025.

### §C.3 Default-value discrepancy (D4 / DECISION-002)

**Command:** `grep -rn "DefaultModelPolicy\|DefaultPerformanceTier" internal/ --include='*.go' | grep -v _test.go`

**Observed:**
- `internal/template/model_policy.go:27` → `const DefaultModelPolicy = ModelPolicyHigh` (= `"high"`), projected `high → max` by `MapModelPolicyToTier` (line 385, comment line 402: "projects high→max on the TIER axis").
- `internal/config/defaults.go:77` → `DefaultPerformanceTier = "medium"`; used at `defaults.go:436` (`PerformanceTier: DefaultPerformanceTier`).
- Template `internal/template/templates/.moai/config/sections/llm.yaml:9` → `performance_tier: "medium"`.

**Finding:** three "defaults" coexist. The new `llm.profile` default is `medium` (matches the config + template defaults). The legacy `DefaultModelPolicy = "high"` constant is a separate init-selection default; its `high → max` projection is retained only for the legacy read-time alias. → Grounds DECISION-002 + REQ-MPM-002.

### §C.4 Template `llm.yaml` retired keys

**Command:** `grep -n "performance_tier\|plan_type\|claude_models\|profile" internal/template/templates/.moai/config/sections/llm.yaml`

**Observed:** `performance_tier: "medium"` (line 9), `plan_type: "subscription"` (line 15), `claude_models:` block (line 22). No `profile:` / `profiles:` / `agent_overrides:` keys yet. → the template rewrite (REQ-MPM-010/032/033) adds `profile` + `profiles` + `agent_overrides` and removes `plan_type` + `claude_models`.

### §C.5 `moai model profile` accessor does not exist yet

**Command:** `grep -rn '"profile"\|model profile\|"model"' internal/cli/*.go | grep -iv test`

**Observed:** an unrelated `moai profile` command (`internal/cli/profile.go:15`, settings profile) and a `--profile` flag on `moai doctor sandbox` (`doctor_sandbox.go:45`). There is **no** `moai model profile` accessor. → REQ-MPM-025's resolver is a NEW CLI surface to build in run-phase (design.md §C), not an existing one.

### §C.6 `modelInherit` sentinel exists

**Read:** `internal/template/model_policy.go:254` → `const modelInherit = "inherit"`; `internal/config/model_routing.go:71` → `Model: "inherit"`. → the resolver's `inherit` return for `Explore`/no-group agents (REQ-MPM-013, AC-MPM-007) reuses an existing sentinel, not a new one.

## §D. GLM effort overlay mechanics (D5 / REQ-MPM-029..031)

**Read:** `internal/template/glm_effort_overlay.go`.

**Observed:**
- `func CollapseClaudeEffortToGLM(effort string) GLMReasoningState` (line 88) — collapses Claude effort {low, medium, high, xhigh, max} → z.ai reasoning state {thinking-off, reasoning-high, reasoning-max}; unrecognized → `reasoning-max` (omit-default, totality clause, line 97).
- `glmCodingMaxOverrideAgents` (lines 102-121) — the coding-max override set is exactly `{manager-develop}` (singleton post SPEC-GLM-EFFORT-TUNE-001 P1). `IsGLMCodingMaxOverrideAgent` forces `reasoning-max` for members (line 131+).
- `func IsGLMBackend(cfg config.LLMConfig) bool` (line 207) — backend detection predicate.
- The overlay comment (line 10) states it is "an OVERLAY, not a third plan_type: plan_type stays" — the profile matrix is backend-neutral; GLM is an overlay on the same profile, not a fourth column.

**Finding:** the profile's model alias maps through `llm.glm.models` and effort collapses via `CollapseClaudeEffortToGLM` (+ the `manager-develop` coding-max override → `reasoning-max`). No new GLM axis is needed. → Grounds REQ-MPM-029/030/031 + AC-MPM-019. NOTE: the overlay is a legitimate **runtime effort consumer** (DECISION-001 channel (b)) — under GLM, the profile effort IS consumed at the overlay layer, unlike the named-subagent Claude spawn where effort stays frontmatter-default.

## §E. Non-affected readers (scope boundary)

- **statusline:** `grep -rn "PerformanceTier\|ClaudeModels\|PlanType" internal/statusline` → no match. The statusline does NOT read these fields; not an affected reader.
- **`agentlint LR-12`:** `ls internal/agentlint/` → empty; no effort-frontmatter lint exists. The effort-frontmatter canonical is documentation (`agent-authoring.md § Effort-Level Calibration Matrix`), not a mechanical gate. → the profile-vs-frontmatter effort divergence (REQ-MPM-027) is not a lint failure.
- **`HaikuResidualRule`** (`internal/spec/lint_haiku_residual.go`): enforces zero `haiku` references. Matrix A uses no haiku → stays green (AC-MPM-024).
- **`model_routing_profiles`** (Tier×Phase axis, `RouteModelFor`, `internal/config/model_routing.go`): keyed by `perfTier`, a separate cost axis. OUT OF SCOPE (REQ-MPM-038).

## §F. Evidence summary table (Claim → Command → Grounds)

| Claim | Command | Grounds |
|-------|---------|---------|
| 4 ApplyTierProfile call sites | `grep -rn ApplyTierProfile internal/ --include='*.go' \| grep -v _test.go` | REQ-MPM-024/028/040, AC-MPM-025 |
| web save-path mutates frontmatter | Read `agentfm.go:74-112` | REQ-MPM-040, AC-MPM-025 |
| per-spawn model-only for named subagents | `grep -n per-spawn model-policy.md` + orchestrator confirm | DECISION-001, REQ-MPM-025/026/027 |
| template source = model: inherit; local = model: opus | `grep -n '^model:' templates/... vs .claude/...` | REQ-MPM-023/024, AC-MPM-015 |
| DefaultModelPolicy=high vs config medium | `grep -rn DefaultModelPolicy\|DefaultPerformanceTier internal/` | DECISION-002, REQ-MPM-002 |
| tierProfiles = 66 cells | Read `model_policy.go:297-315` | spec §A.1 |
| GLM overlay = CollapseClaudeEffortToGLM + {manager-develop} coding-max | Read `glm_effort_overlay.go` | REQ-MPM-029..031, AC-MPM-019 |
| no `moai model profile` yet | `grep -rn model profile internal/cli/` | REQ-MPM-025 (new surface) |
