---
id: SPEC-AGENT-MODEL-INHERIT-001
title: "Plan — subagent model/effort inheritance"
created: 2026-09-26
---

# Plan — SPEC-AGENT-MODEL-INHERIT-001

## §A Context

Tier **L**: ≥3 milestones and well over 10 files (touch set 252 paths at `d6992e3a0`,
research.md §I). The decisions most likely to change — retention seams, the migration strip
step, the main-session persistence boundary — are fixed in M1. Code removal then runs
**consumer-first**: every milestone deletes the last consumers of a symbol before, or in the same
milestone as, the symbol itself, so each boundary compiles and its affected-package tests can run
(§G). Mechanical doctrine and docs edits come last; they also share 21 files with card t1175,
which lands before this run starts.

## §B Known issues

- `make agents-emit` fails closed on an empty `effort` until the Codex manifest changes (research.md §B) — M6 changes both in one commit.
- LR-03 turns every MoAI agent into an error once `effort:` is stripped — it is retired in M4, before M6 strips.
- `internal/cli/testdata/codex-rollouts-t1171/**` embeds `DefaultProfileMatrix` text in captured rollouts; these are fixtures and are regenerated only if a test that reads them fails.
- `IsValidProfile` is a homonym: `internal/profile/profile.go` (user preference profiles) and `config.IsValidProfile` (llm.profile). Only the config one is removed; web `profile_crud.go` / `screens.go` use the profile package and stay.
- The local `.moai/config/sections/llm.yaml` is untracked in the primary checkout; this SPEC never commits it.

## §C Pre-flight (run entry)

1. `git merge-base --is-ancestor WT-rules-diet develop` exits 0; absorb develop inside this worktree — REQ-AMI-001.
2. `sh .moai/reports/t1246/touch-set.sh` (C-collated) + t1257 diff from `git merge-base HEAD WT-role-naming-docs` piped through `LC_ALL=C sort` + `LC_ALL=C comm -12`; halt on a non-empty intersection — REQ-AMI-002.
3. Re-run every research.md inventory command on the absorbed tree and record counts in progress.md §E.2.
4. Baseline: `TestAlwaysLoadedTokenBudget` headroom, `make agents-emit-check`, affected-package `go test`.

## §D Constraints

- Template-First; `make build` after template edits; `make agents-emit` after C2 agent edits; never hand-edit `.codex/agents/moai/*.toml`.
- Tests scoped to affected packages (`internal/config`, `internal/template/...`, `internal/cli/...`, `internal/hook`, `internal/web`, `internal/settings/...`, `internal/harness/...`, `internal/spec`); the full suite is CI's job.
- One writer per tree; no push (lead pushes develop).

## §E Self-verification (per milestone)

Each milestone closes with `go build ./...`, `go vet` and `go test` on the affected packages
(non-zero test counts), `golangci-lint run` on changed packages, and for text milestones a grep
proving the removed patterns are gone from both trees.

## §F Milestones

### M0 — Run entry and baseline (Priority High)
REQ-AMI-001, REQ-AMI-002. No edits. Output: progress.md §E.2 with predicate output, touch set,
overlap intersection, re-measured counts, budget headroom.

### M1 — Retention seams, characterisation, migration step (Priority High)
REQ-AMI-012, REQ-AMI-014 (step only), REQ-AMI-017..020. Characterisation tests written before
any removal: main-session `resolveLaunchEffort` over preference-profile inputs; GLM alias
mapping and session reasoning; audit-pin precedence; web preference-profile
create/rename/delete and main-session effort save (status codes and written files). Additive
changes: the model-free roster SSOT with rosterguard and `config.retainedAgentNames` re-pointed
(D4), adding the roster SSOT registry row and re-pointing `profile-matrix-order` and
`config-retained-agent-names` (design §F); codex/GLM task and audit resolution re-pointed to
pin > backend default, removing those `ResolveAgentModelEffort` consumers and their test
references (D5); the strip step at its three hosts (D14 a/b/c) with the retained-key filter and
the update-report lines, tested per design §C row including the version-matched and
user-cancel rows. Nothing is deleted yet except the re-pointed call sites.

### M2 — Web console removal (Priority High)
REQ-AMI-011. Delete the agentfm tab, handlers, app seams, templ blocks (regenerate
`*_templ.go`), `internal/settings/agentfm`, the settings schema entries for the removed keys,
and the `v4manifest` display helpers once orphaned (D7). Consumers removed here: web uses of
`ResolveAgentModelEffort`, `ProfileMatrixAgents`, `AgentGroup`, `ApplyProfile`,
`ApplyPerformanceTier`, `EffectiveProfile`, `AgentOverrides`, `IsValidPerformanceTier`,
`ValidPerformanceTiers`, `WorkflowAgents`, and the web test files that reference them. The
rosterguard sites `v4manifest-agent-tiers`, `v4manifest-tier-test`, `web-agentfm-display-rank`,
`web-agentfm-display-rank-test` and the `web-agentfm-subset-count` exemption leave in the same
milestone (design §F). M1 characterisation tests stay green.

### M3 — Hook guard removal (Priority High)
REQ-AMI-009, REQ-AMI-010. Delete `agent_model_guard.go` and its test, `pre_tool.go` wiring,
`AgentModelGuard` config type and default (its only other consumers), `prune_logs` entry,
gitignore-artifact test row; reword sibling-guard comments. Test that a config carrying
`agent_model_guard` still loads.

### M4 — CLI, lint and guard consumers (Priority Medium)
REQ-AMI-005, REQ-AMI-008 (command), REQ-AMI-015, REQ-AMI-016. Delete `moai model` (`model.go`,
root registration); reduce `--profile` / `--model-policy` / `--high` / `--medium-alias` /
`--low` to deprecation warnings (D10/D13); drop the agent model-policy question from the
init/update wizards and the `update_wizard` `ApplyProfile` + system.yaml `model_policy` write;
rewrite the retained `moai profile setup` `model_policy` label to main-session wording in four
locales (design H24); retire agentlint LR-03/LR-12 and the routing checks in `workflow_lint.go` /
`sentinels.go` (and dispose of the `agentlint-section-marker` exemption if its comment goes); retire
the routing surfaces in `internal/spec/lint_haiku_residual.go`; retire `cellguard` (D6); remove
the nil-map normalisation in `cli/glm.go` and resolver citations in `mcp_server.go` tool
descriptions. After M4 no non-test consumer of the matrix, the config profile fields, or the
routing keys remains outside `internal/config` and `internal/template`.

### M5 — Producers and schema (Priority Medium)
REQ-AMI-008 (resolver/matrix), REQ-AMI-013. Delete `profile_matrix.go` resolvers and matrix, the
`model_policy.go` perf-tier helpers (D12/D13), the config fields, `profile.go` override
validation, `model_routing.go`, per-agent GLM helpers left orphaned (grep-proven); template
`llm.yaml` and `workflow.yaml` drop the keys and comment blocks; regenerate
`internal/config/testdata/shipped_key_inventory.yaml`; rewrite `.moai/project/product.md:149`
and `tech.md:17`; remove the rosterguard sites `profile-matrix-order`,
`profile-matrix-group-membership`, `profile-matrix-test-expectations`, `template-llm-yaml`,
`product-md-profile-matrix-size`, `tech-md-profile-matrix-size` (design §F). Re-run the M1
characterisation tests.

### M6 — Agent frontmatter, Codex emission, harness manifests, workflow scripts (Priority Medium)
REQ-AMI-003, REQ-AMI-004, REQ-AMI-006, REQ-AMI-007 (scripts). In one commit: strip
`model:`/`effort:` from the 12 template agents, set the Codex manifest
`model_reasoning_effort.emit: false` (D2), adapt `haiku_effort_guard_test` and agentemit golden
tests, run `make agents-emit` and `make agents-emit-check`; then strip the 12 local moai agents and
the 10 local harness agents. Make harness v4 manifest `model`/`effort` optional in `v4manifest`
validation and remove them from the local harness manifests and Runners (D11); remove the
`agent()` `model`/`effort` literals from template and local workflow scripts (D12), including the
local `sync-audit-4dim.js` `judge_effort` argument and the `test-judge-effort-contract.sh` hook test
(design H22).

### M7 — Doctrine text (Priority Low — mechanical, overlaps t1175)
REQ-AMI-007 (doctrine), REQ-AMI-021..024. Apply design.md §D H1–H20, H22 (rule file) and H23
template-first, then local; remove the rosterguard sites `model-policy-profile-matrix-size` and
`-mirror` with H5 and re-run rosterguard to confirm the kept catalog anchors survive (design §F);
`make build`; `TestAlwaysLoadedTokenBudget` before/after; template-neutrality guard;
zone-registry check; run the AC-AMI-007 greps and add a §D row for any residual hit.

### M8 — docs-site (Priority Low)
REQ-AMI-025. Rewrite or remove the docs-site pages (60 in the touch set at `d6992e3a0`: 52 by
pattern + 8 listed in research.md §I, re-measured at M0) in
four locales in one change set, add a redirect for every removed page, keep only deprecation
notes for `--profile` / `--model-policy`, and run the oss-docs verify recipe. CHANGELOG and the
§C supersession closures belong to sync.

## §G Ordering rationale — consumer-first, measured

Measured with `grep -rlE "<symbol>" --include='*.go' internal cmd pkg | grep -v _test.go` at
`d6992e3a0` (research.md §C/§J). Each symbol's non-test consumers leave in the milestone shown,
never after the milestone that deletes the symbol.

| Symbol / field | Consumers → removed in | Producer deleted in |
|---|---|---|
| `ResolveAgentModelEffort` | cli/{glm_task,mcp_codex,mcp_glm} M1 · web M2 · hook M3 · cli/model.go, mcp_server.go text M4 | M5 |
| `ProfileMatrixAgents` | rosterguard M1 (re-point) · web M2 · cli/model.go M4 | M5 |
| `DefaultProfileMatrix` / `ResolveHarnessAgentModelEffort` | agentlint LR-12, cellguard M4 | M5 |
| `ApplyProfile` / `ApplyPerformanceTier` / `IsValidPerformanceTier` / `NormalizeToTier` | web M2 · cli/{init,update,update_wizard}, wizard/questions.go M4 | M5 |
| `AgentModelGuard` | hook, config defaults — same milestone | M3 |
| `LLMConfig.{Profile,Profiles,HarnessAgents,AgentOverrides,PerformanceTier}` / `EffectiveProfile` | web M2 · cli/glm.go, cli/model.go, agentlint M4 · template M5 | M5 |
| `WorkflowAgents` / `ModelRouting*` | settings schema M2 · agentlint, spec haiku lint M4 | M5 |
| agent frontmatter `effort:` | agentlint LR-03 M4 · emitter manifest M6 (same commit as strip) | M6 |
| rosterguard registry sites (design §F) | sites leave with their surface: M1 re-point, M2 web/v4manifest, M5 matrix/config/llm.yaml/project docs, M7 model-policy | same milestone |

**Test files.** `go vet` compiles `_test.go`, so the rule above binds tests too: **each milestone
removes or adapts every `_test.go` reference, in any package, to a symbol it deletes.** Measured
with the same command without `grep -v _test.go` (`grep -rlE "<symbol>" --include='*_test.go' internal cmd pkg`):

| Symbol set | Test files (package) | Milestone |
|---|---|---|
| `ResolveAgentModelEffort` | cli/{codex_session,glm_task,mcp_audit,mcp_convergence,mcp_glm}_test.go (M1, with D5); web/{g3_profile_matrix,m5_agentfm,mcp_audit_surface}_test.go (M2); template/profile_matrix_test.go (M5) | M1/M2/M5 |
| `ProfileMatrixAgents` | rosterguard/rosterguard_test.go (M1 re-point); web/webux_followup_test.go (M2); cellguard/cellguard_test.go (M4); template/profile_matrix_test.go (M5) | M1/M2/M4/M5 |
| `DefaultProfileMatrix` / `ResolveHarnessAgentModelEffort` | web/webux_followup_test.go (M2); cli/agentlint/agent_lint_test.go, cli/profile_setup_schema_options_test.go, cli/wizard/model_policy_matrix_agreement_test.go, cellguard_test.go (M4); template/{embedded_llm_yaml,profile_matrix}_test.go (M5) | M2/M4/M5 |
| `ApplyProfile` / `ApplyPerformanceTier` / `NormalizeToTier` / `IsValidPerformanceTier` | template/{model_policy,profile_matrix}_test.go | M5 |
| `AgentModelGuard` | config/defaults_test.go, hook/agent_model_guard_test.go | M3 |
| `LLMConfig` profile fields / `EffectiveProfile` | web/{fieldsets_states,g3_profile_matrix,webux_followup,webux_haiku_effort}_test.go (M2); cli/{init_quiet_wizard,mcp_audit,profile_setup_schema_options}_test.go (M4); config/{profile,validation}_test.go, template/{embedded_llm_yaml,namespace_protection_audit,profile_matrix}_test.go (M5) | M2/M4/M5 |
| `WorkflowAgents` / `ModelRouting*` | web/agent_settings_test.go (M2); cli/agentlint/workflow_lint_test.go, spec/lint_haiku_residual_test.go (M4); config/{model_routing,workflow_agents}_test.go (M5) | M2/M4/M5 |
| `agentfm` package / web panel | settings/agentfm/agentfm_test.go and the 19 web test files referencing agentfm, cli/mcp_audit_test.go, rosterguard/numeral_test.go, v4manifest/tier_test.go | M2 |

## §H Anti-patterns

- Deleting a producer before its last consumer — the package stops compiling and the milestone's self-verification cannot run.
- Stripping C2 `effort:` before the manifest change — `make agents-emit` fails closed.
- Hand-editing Codex toml files.
- Removing the roster together with the matrix — rosterguard loses its canonical list.
- Treating a leftover config key as an error — breaks `moai update` for edited configs.
- Running `go test ./...` locally.

## §I Cross-references

research.md (inventory, overlap, touch set, persistence measurements), design.md (decisions,
migration, HARD-clause table), acceptance.md (criteria), progress.md §E.1 (operator answers
Q1–Q6, resolved 2026-09-26).
