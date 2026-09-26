---
id: SPEC-AGENT-MODEL-INHERIT-001
title: "Design — subagent model/effort inheritance"
created: 2026-09-26
---

# Design — SPEC-AGENT-MODEL-INHERIT-001

## §A Target state

One rule replaces every channel: **MoAI assigns no model and no effort to any subagent.** A
subagent runs on whatever the main session runs on. MoAI keeps exactly two model/effort
surfaces, both main-session scoped: the launcher's main-session model/effort (policy persisted in
the preference profile, research.md §J) and the GLM backend mapping.

## §B Decisions

| ID | Decision | Alternatives rejected |
|---|---|---|
| D1 | Delete the frontmatter keys rather than write `model: inherit`. | `model: inherit` is equivalent at resolution step 2, but leaves a field the linter, the web panel, and the emitter keep modelling; the card asks for removal. |
| D2 | Codex manifest: `fields.model_reasoning_effort.emit: false`, rationale row rewritten to "omit — inherit the parent". | Emitting an empty value fails closed by design; keeping emission requires an effort source that no longer exists. |
| D3 | Retire LR-03 and LR-12 in the agent linter; keep LR-13 as a value check on an optional key. | Keeping LR-03 as a warning would flag every MoAI agent; inverting it into "must not declare effort" would reject user agents (REQ-AMI-006). |
| D4 | Move the retained-agent roster to one model-free list and re-point rosterguard and `config.retainedAgentNames` at it; dispose of every rosterguard registry site bound to a removed or rewritten surface in the milestone that removes that surface (table §F). | Deleting rosterguard's roster check would silence a guard unrelated to model policy (research.md §D); leaving the sites in place fails rosterguard (`check.go:77` anchor-not-found, `:168` CountPattern 0 matches). |
| D5 | Codex/GLM tools: resolution becomes pin > backend default. | Keeping a hidden per-agent cell for the codex/GLM path would re-create the matrix under another name. |
| D6 | Retire `internal/harness/cellguard` with the matrix. | Its only comparison target is the deleted matrix; re-pointing it has nothing to bind. docs-site tables that restated cells are rewritten or deleted in M8 (REQ-AMI-025). |
| D7 | Remove the whole agent-settings tab (`agentfm`) — UI and handler/API — and `internal/settings/agentfm`; delete `v4manifest` badge/tier-suggestion helpers once a grep shows no remaining consumer. Operator Q1. | A read-only agent list was offered and declined. |
| D8 | Migration: template drops the keys; `moai update` runs an explicit post-merge strip step on the user's `llm.yaml` and `workflow.yaml` with one report line per removed key; configuration loading ignores the keys in any case. Operator Q2. | "Ignore only" was offered and declined; "hard error" breaks update for users who edited the block. |
| D9 | No new `[HARD]` clause is added. The replacement statement ("subagents inherit; pass no model or effort") lives as one plain sentence in `agent-common-protocol.md` and in `model-policy.md` § Valid Model Field Values. | A new always-loaded HARD clause spends budget t1175 just reclaimed. |
| D10 | `--profile` on `init`/`update` stays as an accepted flag that does nothing but print a deprecation warning. Operator Q4. | Removing the flag breaks scripts that pass it. |
| D11 | Harness v4 manifest `model`/`effort` become optional in `v4manifest` validation; `/moai:harness` generation instructions stop emitting them; the Runner passes neither to `agent()`. Operator Q5. | Keeping them required would force every generated manifest to name a model. |
| D12 | `workflow_agents`, `model_routing`, `model_routing_profiles`, `performance_tier` and the workflow-script `agent()` model/effort literals are removed with their validators (`config/model_routing.go`, agentlint `workflow_lint.go` routing checks and `sentinels.go`, the `lint_haiku_residual.go` surfaces that read them, `ApplyPerformanceTier` and its helpers). Operator Q3. | These assign subagent model/effort under another name. |
| D13 | `init --model-policy` / `--high` / `--medium-alias` / `--low` follow D10 (accept, no-op, deprecation warning naming `moai profile setup`), and the init/update wizards drop their agent model-policy question. Measured basis (research.md §J): `resolveModelPolicy` and the wizard `ModelPolicy` answer reach only `ApplyPerformanceTier` (llm.yaml `performance_tier`), `ApplyProfile` (llm.yaml `profile`) and the update wizard's system.yaml `model_policy` write, which has no Go reader; `TemplateContext.WithModelPolicy` has no non-test caller. The main-session policy is `ProfilePreferences.ModelPolicy` in `~/.moai/claude-profiles/<name>/preferences.yaml`, read by `resolveLaunchEffort` and written by `moai profile setup` / `moai web` — unchanged. `MapModelPolicyToEffort`, `IsValidModelPolicy`, `ValidModelPolicies` stay. | Keeping the flags as preference-profile writers would add a new write path the card did not ask for; removing them breaks scripts (same reason as D10). |
| D14 | Three hosts, all after a backup exists (research.md §J): (a) normal path — inside the "Restore Settings" step of `update_template_sync.go`, after `backup.RestoreMoaiConfigRetained` (:620) and before `renderRetainedKeyAdvisory` (:638); the backup was taken by the "Backup" step (`BackupMoaiConfig`, :539) and the strip's key set is filtered out of the retained-key list so the report names those keys once, as removed; (b) version-matched path — in `update.go` inside the `syncSkipped` branch (:524-533), only when the skip reason is the version match (`runTemplateSyncWithProgress` :773-776), taking its own `BackupMoaiConfig` first; not when the user cancelled the merge (:795-799, which returns the same `true`, so the host distinguishes the two reasons); (c) clean-install path — `update_clean_install.go` after `RestoreMoaiConfig` (:511), at the existing `stripRetiredV2DenyEntries` site (:554). | `update.go:388` (the deny-rule strip host) runs before backup and merge — measured, comment :377-381 — so it would strip keys the backup never saw. Relying on the merge is wrong: it retains old-only keys (§C). Folding into `ApplyProfile` would vanish with `ApplyProfile`. Stripping after a user cancel would override the user's refusal. |

## §C Migration behaviour under `moai update` (D8, D14)

Measured merge behaviour (research.md §J): `internal/cli/update/backup/node_merge.go:374`
"Second pass: old-only keys (absent from the new template). REQ-UYP-006 retains ALL of them";
`internal/cli/update/backup/merge.go:78` "key only in old → absent from new template → retain +
report". The merge therefore never removes these keys; the strip step does, for every row.

| User file state | 3-way merge result | Strip step result |
|---|---|---|
| Keys absent | nothing to do | no-op, no report line |
| Keys present, untouched since deploy | retained (old-only); these keys are filtered out of the retained-key advisory by host (a) | removed; report lists each key once, as removed |
| Keys present, user-edited (e.g. `agent_overrides` pins) | retained | removed; backup keeps the original; report lists each key and notes it carried user values |
| `llm.yaml` `performance_tier` | retained | removed (Q3) |
| `workflow.yaml` `workflow_agents` / `model_routing` / `model_routing_profiles` | retained | removed (Q2/Q3) |
| `workflow.yaml` `agent_model_guard` | retained | removed; ignored on load in any case |
| Any of the above on a version-matched update (no sync, no merge) | not run | host (b): own backup, then removed and reported |
| User cancels the merge prompt | not run | not run; files byte-identical |

The strip is line/indent-based like the existing `stripRetiredLLMKeys` (which already drops
`plan_type` and `claude_models`), relocated out of `profile_matrix.go` and extended to the key set
above.

## §D HARD-clause and doctrine relocation table

Removal is operator-decided (card t1246). Each row records what leaves and why. HARD inventory
command and output: research.md §F.

| # | File (local + template) | Old clause head | Marker | Disposition | Reason |
|---|---|---|---|---|---|
| H1 | `rules/moai/core/agent-common-protocol.md` § Per-Spawn Model Injection | "When spawning a subagent, pass the model the active profile resolves for that agent as an explicit `model` argument on the spawn." + 4 bullets + audit-hook paragraph | `[ZONE:Evolvable] [HARD]` | **Removed** (−1). Section replaced by one plain sentence: subagents inherit the main session's model and effort; pass neither on spawn. | The resolver, `moai model profile`, and the audit hook it names are deleted (REQ-AMI-008/009). |
| H2 | `rules/moai/core/agent-common-protocol-reference.md` § Per-Spawn Model Injection rationale | rationale body | none | Removed. | Rationale for H1. |
| H3 | `rules/moai/development/model-policy.md:34` § Inherit-by-Default Convention | "All MoAI agents SHOULD declare `model: inherit` unless…" | `[HARD]` | **Rewritten** (0): agents declare no `model:`/`effort:`; an absent field resolves to the main session (Claude Code order), with the `CLAUDE_CODE_SUBAGENT_MODEL` residual. | Same intent, new mechanism (D1). |
| H4 | `model-policy.md:245` § Rules | "New agent definitions SHOULD use `model: inherit` (default)…" | `[HARD]` | **Rewritten** (0): new MoAI agent definitions omit `model:` and `effort:`. | D1. |
| H5 | `model-policy.md` § Model Policy Tiers, § Per-Agent Profile Resolver (incl. the `:187` per-spawn injection sentence and `:194` Workflow `opts.effort` row), § Harness-Agent Model Policy, § Effort Levels (per-agent part) | tables and prose | none | Removed; § Model Policy Tiers reduced to the main-session tier (preference profile). | Matrix, resolver, `harness_agents`, `workflow_agents` removed. |
| H6 | `model-policy.md:93`, `:110` | Default-model cost lever; GLM reconciliation | `[HARD]` | **Rewritten** (0): wording that contrasts with "per-agent pins" trimmed. | Main-session / GLM scope (REQ-AMI-017/018). |
| H7 | `rules/moai/development/agent-authoring.md` § Effort-Level Calibration Matrix | per-agent effort table | none | Removed; frontmatter reference drops `model`/`effort` from the MoAI template. | REQ-AMI-003. |
| H8 | `rules/moai/core/moai-constitution.md:64`, `moai-constitution-detail.md:36` | "Per-agent effort calibration: `agent-authoring.md` § Effort-Level Calibration Matrix." | none | Removed. The Opus effort-defaults paragraph stays (main session). | Pointer to H7. |
| H9 | `rules/moai/workflow/cache-aware-execution.md` directive 5 and directive 10 (`:27`) | d5: "Inherit the session model on spawns… Omit model overrides unless…"; d10: "…`agent-common-protocol.md` § Per-Spawn Model Injection governs which model a subagent runs on — different axes, not a contradiction." | d5 `[ZONE:Evolvable]`; d10 `[ZONE:Evolvable] [HARD]` | d5 **Rewritten**: the exception clause removed. d10 **Rewritten** (0): its last sentence removed — the cross-reference target is H1; the main-session cache rule stays. | Both reference the removed per-spawn path. |
| H10 | `cache-aware-execution-reference.md` directive 10 note | "apparent conflict with `agent-common-protocol` § Per-Spawn Model Injection" | none | Removed. | The conflict disappears with H1. |
| H11 | `rules/moai/development/agent-patterns.md:300` § Static Model Pin | rejected-approach record | none | Kept as history; one line added that the replacement is "no field". | Decision record, not live doctrine. |
| H12 | `skills/moai/workflows/{harness-builder,harness-build-entry}.md` | `Agent(model: "opus", effort: "xhigh")`, `Agent(agentType: "Explore", effort: "low")`, "Omitting effort is a cost leak" | none | Arguments removed; the cost-leak sentence removed; specialist-generation text stops naming `model`/`effort`. | REQ-AMI-003/007. |
| H13 | `skills/moai-foundation-core/modules/token-optimization.md` | calibration-matrix pointers | none | Removed. | H7. |
| H14 | `skills/moai-foundation-cc/reference/claude-code-sub-agents-official.md:79,164` | "If omitted, uses configured default (usually sonnet)." | none | Corrected to the measured resolution order. | Stale versus the official doc (research.md §A). |
| H15 | agent bodies § "Model/effort escalation" and builder-harness frontmatter guidance | pointers to model-policy defaults | none | Reworded to the inheritance rule. | D1. |
| H16 | `.moai/docs/agent-lint.md` (local + template) rows LR-03, LR-12 | rule rows | none | Rows marked retired. | D3. |
| H17 | `rules/moai/workflow/dynamic-workflows.md:124` § Purpose-driven model+effort selection (+ template mirror) | "When a `.claude/workflows/*.js` script invokes `agent()`, the script author SHALL set `effort` explicitly per the purpose taxonomy below…" | `[ZONE:Evolvable] [HARD]` | **Removed** (−1), together with the section's Config-surface paragraph (`:128`, names `workflow_agents` as SSOT), the purpose taxonomy table (`:130-:138`), the reading-order paragraph, and the codemaps-extract worked example. One plain sentence stays: `agent()` takes no `model`/`effort`, so workflow agents inherit the main loop. | Contradicts REQ-AMI-007; names a key REQ-AMI-013 removes. |
| H18 | `rules/moai/core/settings-management.md:21` (both trees) | "…MoAI resolves per-spawn models through its own model pipeline (model-policy.md)" | none | Rewritten: MoAI assigns no per-spawn model. | Stale after H1/H5. |
| H19 | `rules/moai/development/agent-authoring.md:165` (both trees) | example `Agent(subagent_type: "general-purpose", name: "researcher", model: "haiku")` | none | `model:` argument removed from the example. | REQ-AMI-007. |
| H20 | `rules/moai/workflow/archived-agent-rejection.md:86-91` (both trees) | migration rows naming `Agent(general-purpose, model: …)` | none | `model:` arguments removed from the replacement patterns. | REQ-AMI-007. |
| H21 | `.claude/workflows/*.js` (template + local) and the harness Runners | `agent()` `model`/`effort` option literals | n/a | Removed. | D12. |
| H22 | local-only verify-judge effort channel: `.claude/rules/moai/workflow/verify-judge-effort-contract.md:8-9` ("The orchestrator resolves the `verify-judge` model profile and passes its `judge_effort` to `sync-audit-4dim.js`"), `.claude/workflows/sync-audit-4dim.js:197-208` (`args.judge_effort` → `effort: JUDGE_EFFORT` on four `agent()` calls), `.claude/hooks/tests/test-judge-effort-contract.sh` | rule body | none | Rule file removed; the script's `judge_effort` argument, `JUDGE_EFFORT` constant and `effort:` options removed; the hook test removed (it asserts the removed contract and is wired into neither make nor CI). The template `sync-audit-4dim.js` carries no `judge_effort` (measured) and follows H21. | Its source `workflow_agents.verify-judge` is removed (REQ-AMI-013); the rule instructs an effort pass (REQ-AMI-007). |
| H23 | `rules/moai/development/agent-authoring.md:219` and `:333` (both trees) | :219 "Control reasoning depth with `effort` (xhigh for coding/agentic …)"; :333 cross-reference to `dynamic-workflows.md` § Purpose-driven model+effort selection | none | :219 rewritten to say MoAI agents omit `effort`; :333 cross-reference removed (target removed by H17). | D1, H17. |
| H24 | `internal/cli/wizard/translations.go:444/456/468/480`, `internal/cli/profile_setup_translations.go:165/260` | "Agent model policy — Controls token consumption by assigning optimal models to each agent" (four locales) | n/a | Rewritten to main-session wording (the policy only feeds the main-session effort fallback via `MapModelPolicyToEffort`); native wording per locale. | The `moai profile setup` screen the deprecation warning points to must describe what it now does (REQ-AMI-016). |

Non-HARD residue beyond H1–H24 is enumerated at run time by the AC-AMI-007 greps; any hit they
print becomes an added row here before sync. Zone registry: `zone-registry.md` carries no entry
for H1, H3, H4, H9, or H17 (measured), so no registry row is deleted.

## §E Removal map (code)

| Area | Remove | Keep / re-point |
|---|---|---|
| config | `LLMConfig.{Profile,Profiles,HarnessAgents,AgentOverrides,PerformanceTier}`, `profile.go` (profile enum, overrides validation), `AgentModelGuard*`, `model_routing.go` + `WorkflowAgents`/`ModelRouting*` fields (D12) | `retainedAgentNames` → roster SSOT (D4) |
| template | `profile_matrix.go` (matrix, groups, resolvers, `ApplyProfile`, harness classes); `model_policy.go` perf-tier helpers (`ApplyPerformanceTier`, `IsValidPerformanceTier`, `ValidPerformanceTiers`, `ResolveProjectPerformanceTier`, `MapModelPolicyToTier`, `NormalizeToTier`) once their callers are gone; per-agent GLM helpers if orphaned | `ApplyHarness`, strip step (relocated, D14), `MapModelPolicyToEffort`, `IsValidModelPolicy`, session GLM helpers |
| cli | `model.go` + root registration; flag handling reduced to deprecation warnings (D10/D13); wizard `model_policy` question in init/update; `update_wizard` `ApplyProfile` + system.yaml `model_policy` write; agentlint LR-03/LR-12 and routing checks; nil-map normalisation for removed maps; MCP tool descriptions citing the resolver | main-session model policy path; codex/GLM default resolution (D5); `moai profile setup` wizard |
| hook | `agent_model_guard.go`, its `pre_tool.go` wiring, `prune_logs` audit-file entry | comment references in sibling guards reworded |
| web | agentfm panel, handlers, app seams, settings tab, templ blocks, `settings/agentfm`, schema entries for removed keys | preference-profile CRUD, main-session controls |
| harness | `cellguard` package (D6); required-field checks for specialist `model`/`effort` (D11) | `rosterguard` (re-pointed), `v4manifest` validator with optional fields |
| workflows / docs | `agent()` model/effort literals in `.claude/workflows/*.js` (D12); docs-site pages (M8) | everything else |
| emitter | manifest `model_reasoning_effort` emission (D2) | everything else |

## §F Rosterguard registry disposition (D4)

Measured from `internal/harness/rosterguard/registry.go` at `d6992e3a0` (research.md §K). Each row is
handled in the same milestone as the surface it anchors on, so `go test ./internal/harness/rosterguard/...`
stays green at every boundary.

| Site ID (registry line) | Anchored surface | Disposition | Milestone |
|---|---|---|---|
| new row (added) | the model-free roster SSOT | Added: the canonical roster site | M1 |
| `profile-matrix-order` (:23) | `internal/template/profile_matrix.go` order literal | Re-pointed to the roster SSOT, then removed with the literal | M1 re-point, removed M5 |
| `profile-matrix-group-membership` (:35) | `profile_matrix.go` group map | Removed | M5 |
| `config-retained-agent-names` (:48) | `internal/config/profile.go` | Re-pointed to the roster SSOT (the name set is derived from it) | M1 |
| `profile-matrix-test-expectations` (:91) | `profile_matrix_test.go` | Removed | M5 |
| `v4manifest-agent-tiers` (:107), `v4manifest-tier-test` (:116) | `v4manifest/schema.go` tier table, `tier_test.go` | Removed with the tier helpers (D7) | M2 |
| `shipped-key-inventory` (:203) | `internal/config/testdata/shipped_key_inventory.yaml` | Fixture regenerated without the removed keys; site kept | M5 |
| `template-llm-yaml` (:210) | template `llm.yaml` per-agent cells | Removed (the cells go) | M5 |
| `product-md-profile-matrix-size` (:239), `tech-md-profile-matrix-size` (:254) | `.moai/project/product.md:149`, `.moai/project/tech.md:17` | Prose rewritten, sites removed | M5 |
| `web-agentfm-display-rank` (:358), `web-agentfm-display-rank-test` (:371), numeral exemption `web-agentfm-subset-count` (:822) | `internal/web/agentfm.go`, `agentfm_ordering_test.go` | Removed | M2 |
| `model-policy-profile-matrix-size` (:432), `-mirror` (:440) | `model-policy.md` Per-Agent Profile Resolver row count (H5) | Removed | M7 |
| `agent-authoring-catalog` (:164), `docs-truth-catalog` (:217), `manager-design-catalog-citation` (:575), exemptions `manager-docs-then-8` / `manager-spec-then-8` (:736-737) | catalog sections not removed by this SPEC | Kept; re-run in M7 to confirm the anchors survive H7/H15/H23 | M7 check |
| numeral exemption `agentlint-section-marker` (:817) | a comment in `agent_lint.go` | Kept if the comment survives LR-03/LR-12 retirement; removed otherwise | M4 |
