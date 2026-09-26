---
id: SPEC-AGENT-MODEL-INHERIT-001
title: "Design — subagent model/effort inheritance"
created: 2026-09-26
---

# Design — SPEC-AGENT-MODEL-INHERIT-001

## §A Target state

One rule replaces four channels: **MoAI assigns no model and no effort to any subagent.** A
subagent runs on whatever the main session runs on. MoAI keeps exactly two model/effort
surfaces, both main-session scoped: the launcher's main-session model/effort and the GLM backend
mapping.

## §B Decisions

| ID | Decision | Alternatives rejected |
|---|---|---|
| D1 | Delete the frontmatter keys rather than write `model: inherit`. | `model: inherit` is equivalent at resolution step 2, but leaves a field the linter, the web panel, and the emitter keep modelling; the card asks for removal. |
| D2 | Codex manifest: `fields.model_reasoning_effort.emit: false`, rationale row rewritten to "omit — inherit the parent". | Emitting an empty value fails closed by design; keeping emission requires an effort source that no longer exists. |
| D3 | Retire LR-03 and LR-12 in the agent linter; keep LR-13 as a value check on an optional key. | Keeping LR-03 as a warning would flag every MoAI agent; inverting it into "must not declare effort" would reject user agents (REQ-AMI-007). |
| D4 | Move the retained-agent roster to one model-free list and re-point rosterguard and `config.retainedAgentNames` at it. | Deleting rosterguard's roster check would silence a guard unrelated to model policy (research.md §D). |
| D5 | Codex/GLM tools: resolution becomes pin > backend default. | Keeping a hidden per-agent cell for the codex/GLM path would re-create the matrix under another name. |
| D6 | Retire `internal/harness/cellguard` with the matrix. | Its only comparison target is the deleted matrix; re-pointing it has nothing to bind. docs-site tables that restated cells are rewritten or deleted (Q6). |
| D7 | Remove the whole agent-settings panel (`agentfm` tab) and `internal/settings/agentfm`; delete `v4manifest` badge/tier-suggestion helpers only if a grep shows no remaining consumer. | Keeping a read-only panel still renders model/effort per agent — the thing being removed. Extent confirmed by Q1. |
| D8 | Migration: template drops the keys; `moai update` removes them from a user's `llm.yaml` with a line in the update report, and configuration loading ignores them in any case. | "Ignore only" leaves dead, misleading config in every user project; "hard error" breaks update for users who edited the block. Final choice is Q2. |
| D9 | No new `[HARD]` clause is added. The replacement statement ("subagents inherit; pass no model or effort") lives as one plain sentence in `agent-common-protocol.md` and in `model-policy.md` § Valid Model Field Values. | A new always-loaded HARD clause spends budget t1175 just reclaimed. |

## §C Migration behaviour under `moai update` (D8 default)

| User `llm.yaml` state | Result |
|---|---|
| Keys absent | No change. |
| Keys present, untouched since deploy | The 3-way merge already drops them (template base had them, new template does not); update report names them as removed. |
| Keys present, user-edited (e.g. `agent_overrides` pins) | Removed by an explicit strip step; the pre-update backup keeps the original; update report lists each removed key and its value. |
| `performance_tier` | Per Q3. |
| `workflow.yaml` `agent_model_guard` | Ignored (no YAML ships it). |

The strip reuses the existing write-time retired-key precedent (`stripRetiredLLMKeys`, which
already drops `plan_type` and `claude_models`), relocated so it no longer depends on
`ApplyProfile`.

## §D HARD-clause and doctrine relocation table

Removal is operator-decided (card t1246). Each row records what leaves and why.

| # | File (local + template) | Old clause head | Marker | Disposition | Reason |
|---|---|---|---|---|---|
| H1 | `rules/moai/core/agent-common-protocol.md` § Per-Spawn Model Injection | "When spawning a subagent, pass the model the active profile resolves for that agent as an explicit `model` argument on the spawn." + 4 bullets + audit-hook paragraph | `[ZONE:Evolvable] [HARD]` | Removed. Section replaced by one plain sentence: subagents inherit the main session's model and effort; pass neither on spawn. | The resolver, `moai model profile`, and the audit hook it names are deleted (REQ-AMI-009/010). |
| H2 | `rules/moai/core/agent-common-protocol-reference.md` § Per-Spawn Model Injection rationale | rationale body | none | Removed. | Rationale for H1. |
| H3 | `rules/moai/development/model-policy.md:34` § Inherit-by-Default Convention | "All MoAI agents SHOULD declare `model: inherit` unless…" | `[HARD]` | Rewritten in place: agents declare no `model:`/`effort:`; an absent field resolves to the main session (Claude Code order), with the `CLAUDE_CODE_SUBAGENT_MODEL` residual. Marker kept. | Same intent, new mechanism (D1). |
| H4 | `model-policy.md:245` § Rules | "New agent definitions SHOULD use `model: inherit` (default)…" | `[HARD]` | Rewritten: new MoAI agent definitions omit `model:` and `effort:`. Marker kept. | D1. |
| H5 | `model-policy.md` § Model Policy Tiers, § Per-Agent Profile Resolver, § Harness-Agent Model Policy, § Effort Levels (per-agent part) | tables and prose | no `[HARD]` | Removed; § Model Policy Tiers reduced to the main-session tier only. | Matrix, resolver, `harness_agents` removed. |
| H6 | `model-policy.md:93`, `:110` | Default-model cost lever; GLM reconciliation | `[HARD]` | Kept; wording that contrasts with "per-agent pins" is trimmed. | Main-session / GLM scope (REQ-AMI-018/019). |
| H7 | `rules/moai/development/agent-authoring.md` § Effort-Level Calibration Matrix | per-agent effort table | no `[HARD]` | Removed; frontmatter reference drops `model`/`effort` from the MoAI template. | REQ-AMI-004. |
| H8 | `rules/moai/core/moai-constitution.md:64`, `moai-constitution-detail.md:36` | "Per-agent effort calibration: `agent-authoring.md` § Effort-Level Calibration Matrix." | none | Removed. The Opus effort-defaults paragraph stays (main session). | Pointer to H7. |
| H9 | `rules/moai/workflow/cache-aware-execution.md` directive 5 | "Inherit the session model on spawns… Omit model overrides unless…" | `[ZONE:Evolvable]` | Reworded to state the rule without the exception clause; directive 10 (`[HARD]`, mid-session switch) kept verbatim. | The exception path no longer exists. |
| H10 | `cache-aware-execution-reference.md` directive 10 note | "apparent conflict with `agent-common-protocol` § Per-Spawn Model Injection" | none | Removed. | The conflict disappears with H1. |
| H11 | `rules/moai/development/agent-patterns.md:300` § Static Model Pin | rejected-approach record | none | Kept as history; one line added that the replacement is "no field". | Decision record, not live doctrine. |
| H12 | `skills/moai/workflows/{harness-builder,harness-build-entry}.md` | `Agent(model: "opus", effort: "xhigh")`, `Agent(agentType: "Explore", effort: "low")`, "Omitting effort is a cost leak" | none | Arguments removed; the cost-leak sentence removed. | REQ-AMI-008. |
| H13 | `skills/moai-foundation-core/modules/token-optimization.md` | calibration-matrix pointers | none | Removed. | H7. |
| H14 | `.claude/skills/moai-foundation-cc/reference/claude-code-sub-agents-official.md:79,164` | "If omitted, uses configured default (usually sonnet)." | none | Corrected to the measured resolution order. | Stale versus the official doc (research.md §A). |
| H15 | agent bodies § "Model/effort escalation" and builder-harness frontmatter guidance | pointers to model-policy defaults | none | Reworded to the inheritance rule. | D1. |
| H16 | `.moai/docs/agent-lint.md` (local + template) rows LR-03, LR-12 | rule rows | none | Rows marked retired. | D3. |

Zone registry: `zone-registry.md` carries no entry for H1, H3, H4, or H9 (measured), so no
registry row is deleted.

## §E Removal map (code)

| Area | Remove | Keep / re-point |
|---|---|---|
| config | `LLMConfig.{Profile,Profiles,HarnessAgents,AgentOverrides}`, `profile.go` (profile enum, overrides validation), `AgentModelGuard*` | `retainedAgentNames` → roster SSOT (D4); `PerformanceTier` per Q3 |
| template | `profile_matrix.go` (matrix, groups, resolvers, `ApplyProfile`, harness classes); per-agent GLM helpers if orphaned | `ApplyHarness`, `stripRetiredLLMKeys` (relocated), session GLM helpers |
| cli | `model.go` + root registration; `--profile` handling (Q4); `update_wizard` `ApplyProfile` call; nil-map normalisation for removed maps | main-session model policy path; codex/GLM default resolution (D5) |
| hook | `agent_model_guard.go`, its `pre_tool.go` wiring, `prune_logs` audit-file entry | comment references in sibling guards reworded |
| web | agentfm panel, handlers, app seams, settings tab, templ blocks, `settings/agentfm` | preference-profile CRUD, main-session controls |
| harness | `cellguard` package (D6) | `rosterguard` (re-pointed), `v4manifest` validator per Q5 |
| emitter | manifest `model_reasoning_effort` emission (D2) | everything else |
