---
id: SPEC-GLM-EFFORT-REBALANCE-001
title: "Plan/sync effort rebalance + GLM session reasoning-state step-down"
version: "0.1.0"
status: in-progress
created: 2026-08-14
updated: 2026-08-14
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: "internal/template/profile_matrix.go + internal/template/glm_effort_overlay.go"
lifecycle: spec-anchored
tags: "glm, effort, profile-matrix, reasoning-effort, cost, template-mirror, agentlint"
tier: M
---

# SPEC-GLM-EFFORT-REBALANCE-001 — Plan/sync effort rebalance + GLM session reasoning-state step-down

## HISTORY

| Version | Date | Change | Author |
|---|---|---|---|
| 0.1.0 | 2026-08-14 | Initial plan-phase authoring. Scope settled with the operator before authoring; no clarification markers remain. | manager-spec |

---

## §1 Background

### 1.1 What the operator asked for

Two coupled changes, decided before this SPEC was authored and not re-opened here:

**Change (1) — profile matrix.** In BOTH the `PerformanceTierHigh` and `PerformanceTierMedium` columns of `defaultProfileMatrix` (`internal/template/profile_matrix.go`):

| Agent | Phase | Current | Target |
|---|---|---|---|
| `manager-spec` | plan | `max` | `high` |
| `plan-auditor` | plan | `max` | `high` |
| `manager-docs` | sync | `high` | `low` |

**Change (2) — GLM session-global reasoning state.** `SessionGLMReasoningState()` (`internal/template/glm_effort_overlay.go`) returns `glmReasoningMax` unconditionally. It becomes `glmReasoningHigh`, and its doc comment — which currently justifies the max return by citing `manager-develop` as the representative coding spawn — is rewritten to match.

### 1.2 Why

Under a GLM backend, `CollapseClaudeEffortToGLM` maps Claude's five effort levels onto z.ai's three reasoning states:

```
low            -> thinking-off
medium | high  -> reasoning-high
xhigh | max    -> reasoning-max
```

The operator wants plan-phase work to stop paying the reasoning-max ceiling, and sync-phase document generation to skip the reasoning phase entirely, so GLM spend falls.

### 1.3 What actually moves the GLM wire — and what does not

This distinction is load-bearing and is stated plainly so no downstream reader over-claims.

The per-agent overlay `ResolveGLMReasoning` is defined and unit-tested but is **not on the delivery wire**. Environment variables and the `settings.local.json` env block are session-global, so `glmReasoningEnvVars()` (`internal/cli/glm.go`) writes exactly one session value, derived from `SessionGLMReasoningState()`. The source calls this "the documented delivery-granularity limitation".

Consequently:

- **Change (2) is what moves the GLM wire today**, through **two** consumer paths, not one:
  - `glmReasoningEnvVars()` (`internal/cli/glm.go`) -> `SessionGLMReasoningState()` -> `ANTHROPIC_REASONING_EFFORT`. This is the sub-agent / `settings.local.json` parity wire point.
  - `glmReasoningEnvVarsForEffort(effort)` (`internal/cli/glm.go`), called from `internal/cli/launcher.go` on the **main-session launch path**. On a non-empty effort it collapses that effort directly; on an **empty** effort it falls back to `SessionGLMReasoningState()` (`glm_effort_overlay.go`, the `SessionGLMReasoningStateForEffort` empty branch). Change (2) therefore also moves the main session's no-preference-set fallback from `reasoning-max` to `reasoning-high`.
- **Change (1) records per-agent intent and takes effect on Claude-backed sessions**, where the agent frontmatter `effort:` is the load-bearing channel. It changes what `ResolveGLMReasoning` *would* return per agent, but nothing consumes that per-agent result on the wire, so it does **not** alter delivered GLM behavior.

A second caveat, carried verbatim from the source: whether z.ai actually consumes `ANTHROPIC_REASONING_EFFORT` through the Anthropic-compat shim, versus requiring `reasoning_effort` in the request body (which would make the env var inert), is marked **UNVERIFIED** in `internal/cli/glm.go` and requires a live z.ai session to settle. This SPEC does not settle it and does not assert a measured cost reduction.

### 1.4 Accepted trade-off — the change is global, not GLM-only

There is no kanban (`-k`) axis and no *effective* per-backend axis for effort. `internal/cli/kanban.go` seeds only the autonomy tier. `llm.glm` does carry a per-tier effort field — `GLMSettings` (`internal/config/types.go`) defines `base_url`, `models`, `context_windows`, and `effort` (`GLMTierEffort`: high / medium / low / fable) — but that field is **store-only with no runtime reader**. The struct comment states it plainly: the launcher injects exactly one session-global `ANTHROPIC_REASONING_EFFORT` derived from the overlay, so the four tier fields "record an intent the current single-channel runtime cannot honor per tier", and the console labels them stored-only.

So a per-backend effort axis is not available to this change even though a field of that shape exists. The effort matrix remains a single global surface: a Claude-backed session sees the same stepped-down plan and sync effort as a GLM-backed one. The operator accepts this.

### 1.5 The config-shadow hazard — editing the Go matrix alone changes nothing

`ResolveAgentModelEffort` (`internal/template/profile_matrix.go`) resolves in this order:

1. `llm.agent_overrides[agent]`
2. **the active profile's per-agent cell from config `llm.profiles`**
3. the Go default `defaultProfileMatrix`
4. the `inherit` sentinel

Step 2 hits before step 3. `config.LLMConfig` carries `Profiles map[string]map[string]ModelEffort` (yaml `profiles`) and `HarnessAgents` (yaml `harness_agents`), both read from `.moai/config/sections/llm.yaml`. That file is **gitignored per-project runtime state** — it exists on any machine that has run `moai init`, carrying a full `profiles:` block alongside the user's `profile`, `team_mode`, and GLM settings — so a fresh worktree has no copy of it while a real install does.

The consequence is decisive: **on any machine whose `llm.yaml` already carries a populated `profiles:` block, editing the Go matrix alone changes nothing.** The Go default is consulted only when the config cell is absent. The same mechanism shadows the harness rows — `ResolveHarnessAgentModelEffort` reads `cfg.HarnessAgents[profile][class]` first and falls through to the matrix row only when that cell is absent.

This makes the config surface a required part of the change, not an optional mirror. REQ-GER-012 covers how existing installs pick up the new cells.

### 1.6 Agents deliberately excluded

- **`sync-auditor`** — the code's own phase mapping (`internal/template/profile_matrix.go`, the `defaultProfileMatrix` doc comment) assigns it to the **review** phase, not sync. Excluded.
- **`manager-develop`** (run phase) and **`super-advisor`** (advisor role) — excluded by the operator.
- **`PerformanceTierLow`** — already `low` for all three agents; untouched.

---

## §2 Requirements (GEARS)

### REQ-GER-001 — plan-row effort step-down

The profile matrix shall assign `manager-spec` and `plan-auditor` the effort level `high` in both the `PerformanceTierHigh` column and the `PerformanceTierMedium` column.

### REQ-GER-002 — sync-row effort step-down

The profile matrix shall assign `manager-docs` the effort level `low` in both the `PerformanceTierHigh` column and the `PerformanceTierMedium` column.

### REQ-GER-003 — untouched rows and column

The profile matrix shall preserve, byte-for-byte, the `PerformanceTierLow` column and the `sync-auditor`, `manager-develop`, `super-advisor`, `manager-design`, `builder-harness`, `e2e-tester`, `manager-git`, and `Explore` rows.

### REQ-GER-004 — GLM session-global reasoning state

The session-global GLM reasoning-state deriver shall return the `reasoning-high` state (thinking enabled, `reasoning_effort: high`).

### REQ-GER-005 — harness-class config cells track their named row

**Where** an `llm.yaml` — the shipped template copy or a per-project runtime copy — carries a `harness_agents[profile][class].effort` cell for a class whose `harnessClassRow` row is changed by REQ-GER-001 or REQ-GER-002, the cell shall carry that row's post-change effort value. The affected classes are `synthesize` (`manager-docs` row → `low`) and `research` (`plan-auditor` row → `high`), in **both** the `high` and `medium` columns.

The binding is on the class-to-row mapping in `harnessClassRow`, **not** on the presence of an explanatory comment. The template file annotates its `high` column cells (`# <- manager-docs row`) but leaves the `medium` column cells bare, and a comment-gated requirement would silently exempt the medium column — the exact half-change this SPEC exists to prevent.

Rationale: `ResolveHarnessAgentModelEffort` reads the config cell **first** and falls through to the matrix row only when the cell is absent. A stale cell silently splits the two paths — a project carrying the file gets the old value while a project without it gets the new one. The file's own comment states this hazard.

### REQ-GER-014 — the shipped template config carries the new cells

The shipped `internal/template/templates/.moai/config/sections/llm.yaml` shall carry the post-change values in its `profiles.high` and `profiles.medium` blocks for `manager-spec`, `plan-auditor`, and `manager-docs`, and in its `harness_agents.high` and `harness_agents.medium` blocks per REQ-GER-005.

This is stated separately from REQ-GER-001/002/003 because those bind the Go matrix (`defaultProfileMatrix`) and REQ-GER-012 binds the per-project runtime file. Without this requirement the shipped template's `profiles:` cells would be edited by the plan with no requirement owning them — and, because no test compares the template against `DefaultProfileMatrix()`, nothing else would catch their omission. Every new install would then start life with a config that shadows the Go matrix back to the pre-change values.

### REQ-GER-012 — existing installs pick up the new cells

**Where** a project already carries a populated `.moai/config/sections/llm.yaml`, the affected `profiles` and `harness_agents` cells shall be brought to their post-change values by an **in-place refresh of those cells**, leaving every other key in the file — `profile`, `performance_tier`, `team_mode`, `glm.*`, `agent_overrides` — unmodified.

Mechanism and evidence for why an in-place cell refresh is the mechanism, rather than a whole-file regeneration:

- **No existing code path refreshes these blocks.** `ApplyProfile` rewrites only the `profile:` line; `ApplyPerformanceTier` only `performance_tier:`; `stripRetiredLLMKeys` removes only the retired `plan_type` line and `claude_models` block. None of them touch `profiles:` or `harness_agents:`. A migration therefore has to be an explicit edit — nothing performs it as a side effect.
- **Whole-file regeneration would destroy runtime state.** `saveLLMSection` (`internal/cli/glm.go`) marshals the whole loaded `LLMConfig` back to the file, so it re-serializes whatever `profiles` block was loaded — it preserves stale cells rather than refreshing them, and any path that instead overwrote the file from the template would discard the user's `profile`, `team_mode`, and GLM settings.
- **Absent cells fall through to the Go default.** Because step 3 of the precedence is reached only when the config cell is absent, deleting a `profiles:` block is a valid recovery for an install whose config has drifted — the file then tracks the Go matrix permanently. This is the documented recovery path, not the mechanism this SPEC applies, because it also discards any deliberate per-project cell customization.

### REQ-GER-013 — the change is observable on a checkout that already has a config

**When** `moai model profile --json` is run on a checkout whose `.moai/config/sections/llm.yaml` is present and populated, the tool shall report the post-change cells for the three agents. A change verifiable only against a pristine or absent config does not satisfy this requirement.

### REQ-GER-006 — frontmatter tracks the medium column

The shipped agent frontmatter `effort:` field for `manager-spec`, `plan-auditor`, and `manager-docs` shall equal that agent's `PerformanceTierMedium` cell, in both the working tree and the template tree.

### REQ-GER-007 — resolver reports the new cells

**When** an operator runs `moai model profile --json` under the default (`medium`) profile, the tool shall report `{opus, high}` for `manager-spec`, `{opus, high}` for `plan-auditor`, and `{opus, low}` for `manager-docs`.

### REQ-GER-008 — prose describing the matrix is re-stated

**While** the repository carries prose that enumerates which agents take which effort level, that prose shall state the post-change assignment rather than the pre-change one.

### REQ-GER-009 — structural invariants preserved

The profile matrix shall not change its cell count (33), shall not change the model value of any row, shall not introduce an `xhigh` cell, and shall not violate the per-row monotonicity invariant `high >= medium >= low`.

### REQ-GER-010 — embedded template matches source

The template filesystem embedded in the built binary shall carry the same effort values as the template source tree, because templates reach the binary through `//go:embed all:templates` and a source edit is invisible until the binary is rebuilt.

Verification note: this requirement is **not** observable through `moai model profile`. `runModelProfile` (`internal/cli/model.go`) loads the on-disk project config via `config.NewLoader().Load(...)` and never reads the embedded FS, so that command's output is identical with or without a rebuild. Deciding REQ-GER-010 requires reading the embedded FS directly (`template.EmbeddedTemplates()`); AC-GER-010 does that.

### REQ-GER-011 — no unearned behavioral claim

The delivered artifacts shall not assert that change (1) alters delivered GLM-backend behavior, and shall not assert a measured GLM cost reduction, because the per-agent overlay is off the delivery wire and z.ai's consumption of `ANTHROPIC_REASONING_EFFORT` is unverified.

---

## §3 Constraints

### 3.1 A partial change fails lint

`internal/cli/agentlint/agent_lint.go` rule **LR-12** derives `canonicalEffortMatrix` from `template.DefaultProfileMatrix()[medium]` and compares it against each agent's frontmatter `effort:`. The Go matrix and the six agent files therefore have to move in one unit — changing either side alone produces an LR-12 drift finding.

### 3.2 Template-First

Every edit under `internal/template/templates/` is made in the template source, then `make build` recompiles the binary. The working-tree copies under `.claude/` are mirrored in the same change.

### 3.3 Byte-parity CI

`.claude/rules/moai/development/model-policy.md` is enrolled in the byte-parity mirror allowlist (`internal/template/rule_template_mirror_test.go`). Both trees must receive identical edits.

### 3.4 Template neutrality

No SPEC ID, REQ token, or internal date may leak into `internal/template/templates/**`. The CI guard is `.github/workflows/template-neutrality-check.yaml`.

### 3.5 Cross-platform test-layer verification

`GOOS=windows go build` does not compile `_test.go` files. `GOOS=windows go vet ./...` does, and is the check that establishes the Windows test layer still compiles.

---

## §4 Exclusions — what this SPEC does not build

### Out of Scope — per-agent GLM delivery granularity

- Making `ResolveGLMReasoning` reach the wire per agent. The session-global env / `settings.local.json` channel is the documented delivery-granularity limitation; giving each spawn its own reasoning state is a separate design problem.
- Adding a `reasoning_effort` request-body path so the per-agent value survives to z.ai.

### Out of Scope — verifying z.ai shim consumption

- Empirically determining whether z.ai honours `ANTHROPIC_REASONING_EFFORT` through the Anthropic-compat shim. This needs a live z.ai session and is already tracked as an unverified delivery note in `internal/cli/glm.go`.
- Measuring the resulting GLM spend delta.

### Out of Scope — new effort axes

- A kanban (`-k`) effort axis. `internal/cli/kanban.go` seeds only the autonomy tier.
- Wiring the existing but store-only `llm.glm.effort.{high,medium,low,fable}` fields to a runtime reader, which would give GLM its own effort axis. They persist a preference the single-channel `ANTHROPIC_REASONING_EFFORT` wire cannot honour per tier (`internal/config/types.go`).
- Any mechanism that would let the plan/sync step-down apply to GLM sessions but not Claude sessions.

### Out of Scope — agents outside the three named rows

- `sync-auditor`, whose phase mapping is review rather than sync.
- `manager-develop`, `super-advisor`, and every other matrix row.
- The `PerformanceTierLow` column.

### Out of Scope — an automated config migration for other machines

- Building a migration command that finds and refreshes stale `profiles` / `harness_agents` cells in every existing install's gitignored `llm.yaml`. REQ-GER-012 binds this checkout and the shipped template; it does not ship a migrator.
- Whether `moai update`'s template-managed config sync refreshes an already-populated `llm.yaml` in place, overwrites it whole, or prompts. `internal/cli/update_template_sync.go` deploys with `forceUpdate=true` behind a merge-confirmation preview and `detectUserModifiedConfigs` classifies `.moai/config/sections/*.yaml` as template-managed, but the resulting behaviour on a populated `llm.yaml` was NOT executed and observed during planning. Treating it as the migration path would be an unverified claim.
- Consequently, other machines carrying a stale `profiles:` block keep resolving the old cells until their config is refreshed. This is stated as a known limitation rather than silently assumed away.

### Out of Scope — a template-to-matrix parity test

- No test asserts that `internal/template/templates/.moai/config/sections/llm.yaml` agrees with `template.DefaultProfileMatrix()`. The two are kept aligned by hand and by the file's own comment, so the shipped mirror can drift silently.
- Adding that guard would prevent the whole class of drift this SPEC has to correct by hand, but it is a new test surface with its own design (which direction is authoritative, how `harness_agents` maps onto rows) and belongs in its own SPEC.

### Out of Scope — pre-existing docs-site matrix drift

- All **four** locales of `docs-site/content/{en,ko,ja,zh}/advanced/profile-matrix.md` carry a 33-cell table that **already disagrees** with the committed Go matrix (each shows `manager-spec | opus / high` in the `high` column where the code carries `max`). Each also carries a harness-class table repeating the `synthesize` / `research` rows. This drift predates this SPEC and is not caused by it.
- Correcting it means re-deriving all 33 cells plus the harness-class rows across four locales — a larger change with its own review surface. It belongs in its own SPEC.
- This SPEC does not make the drift worse: after the change, three of the affected docs-site cells coincidentally become correct.

---

## §5 Cross-references

- `internal/template/profile_matrix.go` — the 33-cell Go SSOT, the agent-group membership map, `harnessClassRow`, and `ResolveAgentModelEffort`.
- `internal/template/glm_effort_overlay.go` — `CollapseClaudeEffortToGLM`, the `manager-develop` coding-max override set, and `SessionGLMReasoningState`.
- `internal/cli/glm.go` — `glmReasoningEnvVars()` (sub-agent / settings parity wire) and `glmReasoningEnvVarsForEffort()` (main-session launch path, reached from `internal/cli/launcher.go`). Both reach `SessionGLMReasoningState()`, the second only on the empty-effort branch.
- `internal/cli/agentlint/agent_lint.go` — LR-12 and `buildCanonicalEffortMatrix`.
- `.claude/rules/moai/development/model-policy.md` — the tier table, the effort-baseline prose, and the phase-weighting paragraph.
- SPEC-GLM-EFFORT-TUNE-001 (completed) — the sibling SPEC that reduced the coding-max override set to `{manager-develop}`.
- SPEC-MODEL-TIER-PLANTYPE-001 — the SPEC that introduced the GLM effort overlay and the Branch-B env delivery.
