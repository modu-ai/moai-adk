---
id: SPEC-AGENT-MODEL-INHERIT-001
title: "Subagents inherit the main session's model and effort — remove per-agent model/effort assignment"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template, internal/config, internal/cli, internal/hook, internal/web, .claude/agents, .claude/rules"
lifecycle: spec-anchored
tags: "agent-model, effort, inheritance, profile-matrix-retirement, agentemit, moai-update-migration, web-console"
tier: L
era: V3R6
related_specs: [SPEC-MODEL-PROFILE-MATRIX-001, SPEC-MODEL-PROFILE-MATRIX-002, SPEC-AGENT-MODEL-ENFORCE-001, SPEC-V3R6-AUDIT-MODEL-PIN-001, SPEC-ALWAYS-LOADED-DIET-002, SPEC-ROLE-NAMING-DOCS-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-26 | manager-spec | Initial draft for card t1246 (operator decision 2026-09-26). Inventory measured at develop `d6992e3a0` (research.md). Run is gated on card t1175 merging to develop. |

---

## §A Context

The operator decided on 2026-09-26 that every subagent takes the main session's model and
effort. Today MoAI assigns them per agent through four channels: agent-file frontmatter
(`model:`/`effort:` in 46 files, research.md §B), a spawn-time injection doctrine backed by a
39-cell profile matrix and `moai model profile`, a PreToolUse audit/guard that checks each spawn
against that matrix, and a web console panel that edits both.

Claude Code resolves a subagent's model as: spawn-time `model` → frontmatter `model` →
`CLAUDE_CODE_SUBAGENT_MODEL` → the main conversation's model; an absent `effort` inherits from
the session (research.md §A, measured against the official documentation). Removing the field
and the spawn-time argument is therefore sufficient for inheritance, provided MoAI sets no
`CLAUDE_CODE_SUBAGENT_MODEL` — which it does not.

This SPEC removes the four channels and the configuration that feeds them, keeps the
main-session model/effort and the GLM alias mapping, and migrates existing user projects on
`moai update`.

## §B Requirements (GEARS)

### B.1 Run entry

- **REQ-AMI-001** (Event-driven): **When** the run phase starts, it shall first confirm that card t1175 (`SPEC-ALWAYS-LOADED-DIET-002`) is merged into the local develop branch, absorb develop into this branch, and re-measure every inventory count in research.md §B–§G against the absorbed tree before editing any file.
- **REQ-AMI-002** (Event-driven): **When** the run phase starts, it shall re-measure the file overlap with branch `WT-role-naming-docs` (card t1257) using `git diff --name-only <merge-base>...WT-role-naming-docs` and record the intersection with this SPEC's touch set in progress.md.
- **REQ-AMI-003** (Event-detected): **When** REQ-AMI-001 finds t1175 not merged, or REQ-AMI-002 finds that t1257 has already changed a file in this SPEC's touch set, the run phase shall halt with a blocker report naming the files and shall edit nothing.

### B.2 Agent definitions

- **REQ-AMI-004** (Ubiquitous): Every agent definition file under `internal/template/templates/.claude/agents/` and under `.claude/agents/` (the `moai/` and `harness/` directories) shall carry neither a `model:` nor an `effort:` frontmatter key.
- **REQ-AMI-005** (Ubiquitous): The Codex agent files under `internal/template/templates/.codex/agents/moai/` shall be regenerated only by `make agents-emit` from the template agent definitions, shall carry neither a `model` nor a `model_reasoning_effort` key, and `make agents-emit-check` shall exit 0.
- **REQ-AMI-006** (Ubiquitous): The agent linter shall not report a missing `effort:` key as a finding, and shall not compare an agent's effort against a canonical per-agent matrix.
- **REQ-AMI-007** (Where): **Where** a user-authored agent file declares `model:` or `effort:`, MoAI shall leave that file unchanged and shall not report the declaration as an error.

### B.3 Spawn-time assignment

- **REQ-AMI-008** (Ubiquitous): MoAI doctrine (rules, skills, workflow definitions, output styles) shall not instruct the orchestrator or any agent to pass a `model` or `effort` value when spawning a subagent.
- **REQ-AMI-009** (Ubiquitous): The `moai` binary shall not provide a per-agent model/effort resolver, a `moai model profile` command, or a per-agent profile matrix.
- **REQ-AMI-010** (Ubiquitous): The PreToolUse hook shall not observe, audit, advise on, or deny subagent spawns on the basis of their model, and shall not write `.moai/logs/agent-model-audit.jsonl`.
- **REQ-AMI-011** (Where): **Where** a project's `workflow.yaml` still carries `workflow.agent_model_guard`, loading the configuration shall succeed and the key shall have no effect.

### B.4 Web console

- **REQ-AMI-012** (Ubiquitous): The `moai web` console shall not render a subagent model/effort or profile-selector panel, and shall not accept a form submission that writes agent frontmatter `model`/`effort` or `llm.profile`/`llm.agent_overrides`.
- **REQ-AMI-013** (Ubiquitous): The `moai web` console's user-preference profile routes (`/profile/create`, `/profile/delete`, rename) and the main-session model/effort controls shall keep their current behaviour.

### B.5 Configuration and migration

- **REQ-AMI-014** (Ubiquitous): The template `llm.yaml` shall not carry the `profile`, `profiles`, `harness_agents`, or `agent_overrides` keys or their explanatory comment blocks.
- **REQ-AMI-015** (Event-driven): **When** `moai update` runs on a project whose `llm.yaml` carries any of the keys in REQ-AMI-014, the update shall leave the project in a state where those keys have no effect on any spawn, and shall report the disposition of each such key in its output.
- **REQ-AMI-016** (Event-driven): **When** `moai init` or `moai update` runs with the main-session model policy set, it shall persist the main-session setting and shall not write `llm.profile`.
- **REQ-AMI-017** (Event-detected): **When** a user passes `--profile` to `moai init` or `moai update`, the command shall follow the operator decision recorded for Q4 in progress.md §E.1 (reject with an error naming the removal, or accept as a no-op with a deprecation warning), and shall never write `llm.profile`.

### B.6 Retained behaviour

- **REQ-AMI-018** (Ubiquitous): The main-session model and effort (`moai cc --model`, `model_policy`, `effort_level`, the launcher-injected effort) shall resolve exactly as before this SPEC.
- **REQ-AMI-019** (Ubiquitous): The GLM model alias mapping (`llm.glm.models`) and the session-global GLM reasoning state shall resolve exactly as before this SPEC.
- **REQ-AMI-020** (Ubiquitous): The cross-model audit pins `workflow.audit.{claude,codex,glm}` shall keep precedence over any default; when no pin is set, the codex and GLM tools shall fall back to their backend default model.
- **REQ-AMI-021** (Ubiquitous): The retained-agent roster consumed by `internal/harness/rosterguard` and by configuration validation shall have exactly one source of truth that carries no model or effort value.

### B.7 Doctrine and build discipline

- **REQ-AMI-022** (Ubiquitous): Every removed or rewritten `[HARD]` clause shall be listed in design.md §D with its file, its old head text, its disposition, and the reason, and the `[HARD]` marker count delta of each touched rule file shall equal the count of clauses that table records as removed from it.
- **REQ-AMI-023** (Ubiquitous): Every change to a file with a template mirror shall be made first under `internal/template/templates/`, then in the local copy, with `make build` after the template edits; local-only files (`.claude/agents/harness/*`, `hns-*` skills, dev-only workflows) shall be edited locally only.
- **REQ-AMI-024** (Ubiquitous): Template edits shall pass the template-neutrality guard — no SPEC IDs, card IDs, internal dates, or single-programming-language bias introduced under `internal/template/templates/**`.
- **REQ-AMI-025** (Ubiquitous): `TestAlwaysLoadedTokenBudget` shall pass after the change, and progress.md shall record its before and after headroom.

## §C Dependencies

- Run starts only after card t1175 (`SPEC-ALWAYS-LOADED-DIET-002`, branch `WT-rules-diet`) merges to develop (REQ-AMI-001). 21 files overlap (research.md §H).
- Card t1257 (`SPEC-ROLE-NAMING-DOCS-001`) is plan-only at `ffc83b3b1`; its planned surface overlaps 90 paths. Whichever card runs second absorbs the other.

## §D Out of Scope

### Out of Scope — main-session model and effort
- `moai cc --model`, `moai glm`, `model_policy`, `effort_level`, the statusline effort display, and the launcher's effort injection are not changed.

### Out of Scope — GLM alias mapping and session reasoning
- `llm.glm.models`, `llm.glm.effort`, `SessionGLMReasoningState*`, and `CollapseClaudeEffortToGLM*` stay. Only per-agent GLM helpers left without consumers are removed.

### Out of Scope — cross-model audit backends
- `workflow.audit.{claude,codex,glm}` pins and the `claude_audit` child process's `--model`/`--effort` flags stay: they select the model of a separate process, not of a Claude Code subagent.

### Out of Scope — user-owned agents and harnesses
- Agent files a user authored, including harness specialists generated into a user project before this change, are not rewritten by `moai update`.

### Out of Scope — SPEC history
- Released CHANGELOG entries and other SPECs' records under `.moai/specs/**` are not rewritten; status changes on superseded SPECs belong to the sync phase.

## §E Conditional scope (pending operator answers — progress.md §E.1)

- Q3 decides whether `workflow_agents`, `model_routing`, `model_routing_profiles`, `performance_tier`, and the dynamic-workflow `agent()` model/effort literals are removed (default in this plan: removed, as they assign subagent model/effort).
- Q5 decides whether harness v4 manifest `model`/`effort` fields become optional and ignored.
- Q6 decides whether the 48 docs-site pages are updated in this SPEC or a follow-up card.

## §F Known residual

- A user who exports `CLAUDE_CODE_SUBAGENT_MODEL` in their own environment still pins subagents (Claude Code resolution step 3). MoAI does not set it; the model-policy rule states the residual.
