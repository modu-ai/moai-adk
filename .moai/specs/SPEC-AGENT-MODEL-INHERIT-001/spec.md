---
id: SPEC-AGENT-MODEL-INHERIT-001
title: "Subagents inherit the main session's model and effort — remove per-agent model/effort assignment"
version: "0.6.0"
status: completed
created: 2026-09-26
updated: 2026-09-29
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template, internal/config, internal/cli, internal/hook, internal/web, .claude/agents, .claude/rules"
lifecycle: spec-anchored
tags: "agent-model, effort, inheritance, profile-matrix-retirement, agentemit, moai-update-migration, web-console"
tier: L
era: V3R6
related_specs: [SPEC-MODEL-PROFILE-MATRIX-001, SPEC-MODEL-PROFILE-MATRIX-002, SPEC-MODEL-MATRIX-CORE-001, SPEC-MODEL-MATRIX-CONFIG-001, SPEC-MODEL-MATRIX-SURFACES-001, SPEC-MODEL-MATRIX-DOCS-001, SPEC-AGENT-MODEL-ENFORCE-001, SPEC-V3R6-AUDIT-MODEL-PIN-001, SPEC-ALWAYS-LOADED-DIET-002, SPEC-ROLE-NAMING-DOCS-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-26 | manager-spec | Initial draft for card t1246 (operator decision 2026-09-26). Inventory measured at develop `d6992e3a0` (research.md). Run is gated on card t1175 merging to develop. |
| 0.2.0 | 2026-09-26 | manager-spec | Operator answers Q1–Q6 folded in (progress.md §E.1). The former run-entry halt requirement was folded into the first two requirements to hold the Tier L ceiling. |
| 0.3.0 | 2026-09-26 | manager-spec | plan-audit iter-1 (FAIL 0.76, `.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-1.md`) D1–D16 addressed. Requirements renumbered to a contiguous REQ-AMI-001..025 (the ids formerly 004..026 are now 003..025; the folded run-entry halt stays inside 001/002). Merge premise corrected from measured `node_merge.go` behaviour (design §C); main-session persistence surface named and the `--model-policy` family made consistent with Q4 (REQ-AMI-015/016, design D13); run-entry touch set committed as a file (research §I); milestones re-sequenced consumer-first so every boundary compiles (plan §F/§G); the dynamic-workflows `[HARD]` clause and stale non-HARD sentences added to design §D; docs-site page set re-measured with a narrowed pattern plus the Q3 keys (52 pages); four live matrix SPECs named with a supersession sentence (§C). |
| 0.4.0 | 2026-09-26 | manager-spec | plan-audit iter-2 (FAIL 0.82, `.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-2.md`) N1–N9 addressed: strip-step host re-measured and moved after backup+merge on the normal path, with its own backup on the version-matched path (REQ-AMI-014, design D14); rosterguard registry sites bound to removed surfaces assigned per milestone (design D4, plan §G); the local verify-judge effort channel inventoried (design H22) and made visible to AC-AMI-007; test-file consumers added to plan §G; AC-AMI-006 lints the fixture path explicitly; touch set C-collated and extended (research §I); docs residue, agent-authoring residue and profile-setup wording added. Operator decision recorded: the init/update wizard agent model-policy question is deleted (progress.md §E.1 Q7). Counts unchanged: 25 REQ / 25 AC. |
| 0.5.0 | 2026-09-26 | manager-spec | plan-audit iter-3 (FAIL 0.83, `.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-3.md`); one extra round approved by the operator beyond the Tier L cap (progress.md §E.1). R1–R9 addressed: roster premise names its real surviving consumer and `retainedAgentNames` is removed with its last consumer while `ModelEffort` is kept and relocated (design D4/§E); `shipped-key-inventory` and `web-i18n-agent-descriptions` rosterguard sites disposed with the i18n keys (design §F); clean-install path filters the retained-key advisory (D14 c); AC-AMI-006 uses an observable linter signal; Out of Scope reconciled with H24; ja/zh wording and eight `effort_level` strings added; ja/zh multi-llm index pages added; Q7 carried to §E and DoD; update skip-reason return and cancel seam named (D14 b). Counts unchanged: 25 REQ / 25 AC. |
| 0.6.0 | 2026-09-29 | manager-spec | Sync-audit F1 amendment (sync-audit FAIL 77.2, `.moai/reports/t1246/sync-audit.md`; lead dispatch 2026-09-29, repair option (i) — formal amendment, not observer-layer deletion). REQ-AMI-009 redefined to the declaration-only observer design the run landed: the guard exists solely to classify each `Agent` spawn as declared vs inherit — no enforcement, no model/effort resolution, no advisory or deny — and MUST NOT write `.moai/logs/agent-model-audit.jsonl`; the file staying unwritten is the requirement. AC-AMI-009 reworded to match (reword only; count unchanged: 25 REQ / 25 AC). Rationale: the M5 declared/inherit transition proved superior to the planned full deletion (plan.md M3, design §E hook row) — deleting the observer would lose the observation capability, and the auditor assessed the removal work itself as verified; the failure was close-out truthfulness. Prescription: sync-audit F1 option (i) + lead dispatch 2026-09-29. Prune/audit-file fate measured in the same commit (design §G): no prune entry in `internal/cli` (grep: 1 hit, a doctor test fixture); the age-out entry survives at `internal/hook/prune_logs.go:169`. |

---

## §A Context

The operator decided on 2026-09-26 that every subagent takes the main session's model and
effort. Today MoAI assigns them per agent through four channels: agent-file frontmatter
(`model:`/`effort:` in 46 files, research.md §B), a spawn-time injection doctrine backed by a
39-cell profile matrix and `moai model profile`, a PreToolUse audit/guard that checks each spawn
against that matrix, and a web console panel that edits both. Two further channels assign
subagent model/effort under other names: the `workflow.yaml` `workflow_agents` /
`model_routing*` blocks with the dynamic-workflow `agent()` literals, and the harness v4 manifest
specialist fields.

Claude Code resolves a subagent's model as: spawn-time `model` → frontmatter `model` →
`CLAUDE_CODE_SUBAGENT_MODEL` → the main conversation's model; an absent `effort` inherits from
the session (research.md §A, measured against the official documentation). Removing the field
and the spawn-time argument is therefore sufficient for inheritance, provided MoAI sets no
`CLAUDE_CODE_SUBAGENT_MODEL` — which it does not.

This SPEC removes those channels and the configuration that feeds them, keeps the main-session
model/effort (whose policy lives in the preference profile, research.md §J) and the GLM alias
mapping, and migrates existing user projects on `moai update`.

## §B Requirements (GEARS)

### B.1 Run entry

- **REQ-AMI-001** (Event-driven): **When** the run phase starts, it shall first evaluate `git merge-base --is-ancestor WT-rules-diet develop` and, on exit 0, absorb develop into this branch and re-measure every inventory count in research.md §B–§J against the absorbed tree before editing any file; **when** that command exits non-zero (card t1175 not merged), it shall halt with a blocker report and edit nothing.
- **REQ-AMI-002** (Event-driven): **When** the run phase starts, it shall regenerate the touch set with `sh .moai/reports/t1246/touch-set.sh`, compute `git diff --name-only $(git merge-base HEAD WT-role-naming-docs) WT-role-naming-docs`, record both lists and their `LC_ALL=C comm -12` intersection (both lists C-collated) in progress.md; **when** that intersection is non-empty, the run phase shall halt with a blocker report naming the files and shall edit nothing.

### B.2 Agent definitions

- **REQ-AMI-003** (Ubiquitous): Every agent definition file under `internal/template/templates/.claude/agents/` and under `.claude/agents/` (the `moai/` and `harness/` directories) shall carry neither a `model:` nor an `effort:` frontmatter key, and the `/moai:harness` generation instructions (the harness-builder and harness-build-entry workflow skills and the builder-harness agent body, both trees) shall not instruct writing a `model`/`effort` value into a generated specialist file or harness v4 manifest.
- **REQ-AMI-004** (Ubiquitous): The Codex agent files under `internal/template/templates/.codex/agents/moai/` shall be regenerated only by `make agents-emit` from the template agent definitions, shall carry neither a `model` nor a `model_reasoning_effort` key, and `make agents-emit-check` shall exit 0.
- **REQ-AMI-005** (Ubiquitous): The agent linter shall not report a missing `effort:` key as a finding, and shall not compare an agent's effort against a canonical per-agent matrix.
- **REQ-AMI-006** (Where): **Where** a user-authored agent file or an existing harness v4 manifest declares `model`/`effort`, MoAI shall leave that file unchanged, shall parse and validate it without error (the manifest fields become optional), and shall not report the declaration as an error.

### B.3 Spawn-time assignment

- **REQ-AMI-007** (Ubiquitous): MoAI doctrine (rules, skills, output styles) shall not instruct the orchestrator or any agent to pass or set a `model` or `effort` value when spawning a subagent, and the dynamic-workflow scripts under `.claude/workflows/` (template and local) shall pass no `model` or `effort` option to `agent()` and accept no effort-carrying argument such as `judge_effort`.
- **REQ-AMI-008** (Ubiquitous): The `moai` binary shall not provide a per-agent model/effort resolver, a `moai model profile` command, or a per-agent profile matrix.
- **REQ-AMI-009** (Ubiquitous): The agent-model guard (`internal/hook/agent_model_guard.go`) shall exist solely as a declaration-only observation layer — it shall classify every `Agent` spawn as `declared` (a `model` argument was passed) or `inherit` (none was), shall perform no enforcement, no model/effort resolution, and no model-based advisory or deny, and shall not write `.moai/logs/agent-model-audit.jsonl`; the audit file staying unwritten is itself part of the requirement.
  - *Amendment 2026-09-29 (v0.6.0, sync-audit F1 — lead dispatch 2026-09-29, repair option (i): formal manager-spec amendment, not observer-layer deletion).* The original REQ-AMI-009 forbade observation outright and prescribed deleting the guard, its `pre_tool.go` wiring, and the `prune_logs` audit-file entry (plan.md M3, design §E hook row). The run instead landed the guard as the declared/inherit classifier the audit measured as behaviourally correct (sync-audit E-3) but formally in breach. **Rationale:** the M5 declared/inherit transition design proved superior to the planned full deletion — deleting the observer would discard the declaration-observation capability for no requirement gain — and the auditor assessed the removal work itself as verified; the failure was close-out truthfulness, not execution. **Prescription source:** `.moai/reports/t1246/sync-audit.md` F1 option (i) + lead dispatch 2026-09-29. Prune/audit-file fate measured in the same commit (design §G).
- **REQ-AMI-010** (Where): **Where** a project's `workflow.yaml` still carries `workflow.agent_model_guard`, loading the configuration shall succeed and the key shall have no effect.

### B.4 Web console

- **REQ-AMI-011** (Ubiquitous): The `moai web` console shall carry no agent-settings tab — neither its UI (tab, panel, profile selector, per-agent model/effort controls) nor its handler/API path — and shall not write agent frontmatter `model`/`effort` or `llm.profile`/`llm.agent_overrides` on any request.
- **REQ-AMI-012** (Ubiquitous): The `moai web` console's user-preference profile routes (`/profile/create`, `/profile/delete`, rename) and the main-session model/effort controls shall keep the behaviour captured by the M1 characterisation tests.

### B.5 Configuration and migration

- **REQ-AMI-013** (Ubiquitous): The template `llm.yaml` shall not carry the `profile`, `profiles`, `harness_agents`, `agent_overrides`, or `performance_tier` keys, the template `workflow.yaml` shall not carry the `workflow_agents`, `model_routing`, or `model_routing_profiles` keys, and neither shall carry their explanatory comment blocks.
- **REQ-AMI-014** (Event-driven): **When** `moai update` runs on a project whose `llm.yaml` or `workflow.yaml` carries any of the keys in REQ-AMI-013, or whose `workflow.yaml` carries `agent_model_guard`, the update shall remove each such key from the user's file in an explicit strip step that runs after the configuration backup and, where a merge runs, after the configuration merge; the backup shall hold the original keys; every removed key shall appear once in the update report as removed (not as retained); the step shall also run on a version-matched update that skips the template sync, taking its own backup first, and shall not run when the user cancels the merge.
- **REQ-AMI-015** (Event-driven): **When** `moai init` or `moai update` runs, it shall write neither `llm.profile` nor `llm.performance_tier`, shall not ask the agent model-policy wizard question, and shall leave the main-session model policy exactly where it is persisted today — the preference profile `~/.moai/claude-profiles/<name>/preferences.yaml` key `model_policy`, written by `moai profile setup` and `moai web`.
- **REQ-AMI-016** (Event-driven): **When** a user passes `--profile`, `--model-policy`, `--high`, `--medium-alias`, or `--low` to `moai init` or `--profile` to `moai update`, the command shall accept the flag, shall emit a deprecation warning stating that subagents now inherit the main session's model and effort and naming `moai profile setup` for the main-session policy, shall complete otherwise unchanged, and shall write neither `llm.profile` nor `llm.performance_tier`.

### B.6 Retained behaviour

- **REQ-AMI-017** (Ubiquitous): The main-session model and effort (`moai cc --model`, the preference-profile `model_policy` and `effort_level`, the launcher-injected effort from `resolveLaunchEffort`) shall resolve exactly as before this SPEC.
- **REQ-AMI-018** (Ubiquitous): The GLM model alias mapping (`llm.glm.models`) and the session-global GLM reasoning state shall resolve exactly as before this SPEC.
- **REQ-AMI-019** (Ubiquitous): The cross-model audit pins `workflow.audit.{claude,codex,glm}` shall keep precedence over any default; when no pin is set, the codex and GLM tools shall fall back to their backend default model.
- **REQ-AMI-020** (Ubiquitous): The retained-agent roster consumed by `internal/harness/rosterguard` (its only consumer once `config.validateAgentOverrides`, the sole reader of `retainedAgentNames`, is removed) shall have exactly one source of truth that carries no model or effort value, and every rosterguard registry site anchored on a surface this SPEC removes or rewrites shall be removed or re-pointed in the milestone that removes or rewrites that surface.

### B.7 Doctrine and build discipline

- **REQ-AMI-021** (Ubiquitous): Every removed or rewritten `[HARD]` clause shall be listed in design.md §D with its file, its old head text, its disposition, and the reason, and the `[HARD]` marker count delta of each touched rule file shall equal the count of clauses that table records as removed from it (a rewritten clause keeps its marker and counts zero).
- **REQ-AMI-022** (Ubiquitous): Every change to a file with a template mirror shall be made first under `internal/template/templates/`, then in the local copy, with `make build` after the template edits; local-only files (`.claude/agents/harness/*`, `hns-*` skills, dev-only workflows) shall be edited locally only.
- **REQ-AMI-023** (Ubiquitous): Template edits shall pass the template-neutrality guard — no SPEC IDs, card IDs, internal dates, or single-programming-language bias introduced under `internal/template/templates/**`.
- **REQ-AMI-024** (Ubiquitous): `TestAlwaysLoadedTokenBudget` shall pass after the change, and progress.md shall record its before and after headroom.
- **REQ-AMI-025** (Ubiquitous): The docs-site pages in the touch set (62 at `d6992e3a0`: the 52 matched by the research.md §F docs pattern plus the 10 listed explicitly in research.md §F, re-measured at run entry) shall be rewritten or removed in one change set, every removed page's URL shall carry a redirect, and the four locales shall keep section parity.

## §C Dependencies and superseded work

- Run starts only after card t1175 (`SPEC-ALWAYS-LOADED-DIET-002`, branch `WT-rules-diet`) merges to develop (REQ-AMI-001). 21 files overlap (research.md §H).
- Card t1257 (`SPEC-ROLE-NAMING-DOCS-001`, branch `WT-role-naming-docs`) has not changed any touch-set file at `024b95f77` (intersection 0, research.md §I); its planned surface overlaps. REQ-AMI-002 re-measures at run entry.
- **Supersession.** This SPEC reverses the target of SPEC-MODEL-PROFILE-MATRIX-001 (completed), SPEC-MODEL-PROFILE-MATRIX-002 (superseded; its `partially_superseded_by` successors follow) and those successors SPEC-MODEL-MATRIX-CORE-001 (in-progress), SPEC-MODEL-MATRIX-CONFIG-001 (draft), SPEC-MODEL-MATRIX-SURFACES-001 (draft), and SPEC-MODEL-MATRIX-DOCS-001 (in-progress), and removes the guard of SPEC-AGENT-MODEL-ENFORCE-001 (completed). No lane shall start or resume one of the four SPEC-MODEL-MATRIX-* SPECs once this SPEC is in run. Their lifecycle closure — `status: superseded` with `superseded_by: SPEC-AGENT-MODEL-INHERIT-001` on the four non-terminal ones and a `partially_superseded_by` note on the completed ones — is performed in this SPEC's sync phase (manager-spec owns `* → superseded`).

## §D Out of Scope

### Out of Scope — main-session model and effort
- `moai cc --model`, `moai glm`, the preference-profile `model_policy` and `effort_level`, the behaviour of the `moai profile setup` wizard (its `model_policy` and `effort_level` wording is rewritten to main-session terms — design H24), the statusline effort display, and the launcher's effort injection are not changed.

### Out of Scope — GLM alias mapping and session reasoning
- `llm.glm.models`, `llm.glm.effort`, `SessionGLMReasoningState*`, and `CollapseClaudeEffortToGLM*` stay. Only per-agent GLM helpers left without consumers are removed.

### Out of Scope — cross-model audit backends
- `workflow.audit.{claude,codex,glm}` pins and the `claude_audit` child process's `--model`/`--effort` flags stay: they select the model of a separate process, not of a Claude Code subagent.

### Out of Scope — user-owned agents and harnesses
- Agent files a user authored, including harness specialists generated into a user project before this change, are not rewritten by `moai update`.

### Out of Scope — SPEC history
- Released CHANGELOG entries and other SPECs' bodies under `.moai/specs/**` are not rewritten; only the frontmatter lifecycle closure named in §C happens, in sync.

## §E Resolved scope (operator answers 2026-09-26 — progress.md §E.1)

- Q1: the whole agent-settings tab (UI + API) is removed — REQ-AMI-011.
- Q2: `moai update` strips leftover keys and lists them in the update report — REQ-AMI-014.
- Q3: `workflow_agents`, `model_routing`, `model_routing_profiles`, `performance_tier`, and the dynamic-workflow `agent()` model/effort values are removed — REQ-AMI-007, REQ-AMI-013.
- Q4: `--profile` stays as a no-op with a deprecation warning; the `--model-policy` / `--high` / `--medium-alias` / `--low` init flags follow the same rule because measurement shows they reach only the removed keys (research.md §J) — REQ-AMI-016.
- Q5: harness v4 manifest `model`/`effort` become optional; `/moai:harness` stops generating them; existing manifests still parse — REQ-AMI-003, REQ-AMI-006.
- Q6: the docs-site pages are in scope as M8 — REQ-AMI-025.
- Q7: the init/update wizard "agent model policy" question is deleted; the main-session policy stays in `moai profile setup` — REQ-AMI-015, design D13.

## §F Known residual

- A user who exports `CLAUDE_CODE_SUBAGENT_MODEL` in their own environment still pins subagents (Claude Code resolution step 3). MoAI does not set it; the model-policy rule states the residual.
