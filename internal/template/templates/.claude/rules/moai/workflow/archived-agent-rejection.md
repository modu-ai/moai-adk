---
description: "ARCHIVED_AGENT_REJECTED orchestrator behavior + 12-agent migration table. Loaded when an Agent() spawn references an archived subagent_type."
paths: ".claude/skills/moai/workflows/run.md,.claude/skills/moai/workflows/plan.md,.claude/skills/moai/workflows/sync.md,.claude/rules/moai/workflow/orchestration-mode-selection.md"
metadata:
  version: "1.0.0"
  status: "active"
  updated: "2026-05-25"
  tags: "archived-agent, migration, anthropic-2026-alignment, error-spec, orchestrator-discipline"
---

# Archived Agent Rejection

Canonical specification of the `ARCHIVED_AGENT_REJECTED` error and the 12-agent migration table introduced by the agent catalog consolidation policy. When the MoAI orchestrator detects an `Agent()` spawn attempt — or a paste-ready resume reference — naming one of the 12 archived agents, the orchestrator MUST reject the spawn, emit a structured error referencing this migration table, and (when proceeding) consult the user via `AskUserQuestion` on the replacement-pattern decision.

> Cross-reference: the agent catalog consolidation policy's retain-vs-archive matrix is the design-time SSOT for the retention decision; this rule is the canonical runtime SSOT for the orchestrator's rejection behavior. The archived agent files are preserved offline in a migration backup location (original `core/`, `meta/`, `expert/`, `agency/` substructure with a per-agent README).

---

## §A — Error Specification

### §A.1 Error token

`ARCHIVED_AGENT_REJECTED` is the canonical error token emitted by the orchestrator (and, on a future enforcement milestone, by `internal/agents/` Go-side enforcement) when a spawn attempt names an archived agent. The token is grep-stable and CI-verifiable.

### §A.2 Trigger conditions

The orchestrator's `ARCHIVED_AGENT_REJECTED` error fires when ANY of the following hold:

- A direct `Agent(subagent_type: "<archived-name>", ...)` call is issued from the main session
- A paste-ready resume Block 5 contains `/moai <subcommand>` whose handler would delegate to an archived agent
- A workflow router skill (`.claude/skills/moai/workflows/{plan,run,sync}.md` body) references an archived agent by name in a spawn instruction
- A SPEC artifact's plan.md milestone owner column names an archived agent

### §A.3 Exit code semantics

The orchestrator's behavior on detection:

| Condition | Orchestrator action | User-facing prompt |
|-----------|---------------------|--------------------|
| Direct spawn attempt | Halt spawn, emit `ARCHIVED_AGENT_REJECTED` with violator + canonical replacement | `AskUserQuestion` (after `ToolSearch(query: "select:AskUserQuestion")` preload) presenting replacement options |
| Paste-ready resume reference | Emit informational notice; auto-substitute the replacement-pattern owner (when unambiguous) OR `AskUserQuestion` (when ambiguous) | Depends on ambiguity — single-owner replacements (e.g., `manager-quality` → Stop hook) can substitute silently with an info log; multi-option replacements (e.g., `expert-backend` → per-spawn `Agent(general-purpose)` with backend whitelist) MUST go through `AskUserQuestion` |
| Workflow router skill reference (during M2 modification only) | Reject the modification (rule violation) | n/a — caught at rule-authoring time, not runtime |

When the orchestrator emits `ARCHIVED_AGENT_REJECTED`, the structured payload MUST include:

```text
ARCHIVED_AGENT_REJECTED
violator: <agent-name>
context: <spawn-call | paste-ready-resume | workflow-skill | other>
canonical_replacement: <see migration table in §C below>
migration_doc: .claude/rules/moai/workflow/archived-agent-rejection.md §C
retention_matrix: the agent catalog consolidation policy's retain-vs-archive matrix
```

Note: the orchestrator generates the prose summary above and surfaces it via the response body and (when applicable) via the `AskUserQuestion` prompt body. Subagents themselves MUST NOT invoke `AskUserQuestion` per the orchestrator-subagent boundary in `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary; the orchestrator translates the rejection into the user-facing prompt.

---

## §B — Detection Triggers (orchestrator-side)

The orchestrator employs the following detection patterns:

| # | Pattern | Detection mechanism |
|---|---------|---------------------|
| 1 | `subagent_type: "manager-strategy"` (or any other archived name) in an `Agent()` parameter list | Pre-spawn parameter inspection — orchestrator parses the planned `Agent()` call before issuing it |
| 2 | `manager-strategy` (or any archived name) as a bare token in a paste-ready resume Block 5 `/moai <subcommand>` | Paste-ready resume parser scans Block 5 for `/moai <subcommand>` and inspects the resolved handler's owner |
| 3 | Workflow router skill body string `subagent_type.*"<archived-name>"` | Static rule scan at M2 modification time (runtime detection is best-effort) |
| 4 | plan.md milestone owner column literal match against the 12-agent archive list | SPEC artifact authoring lint (manager-spec self-check + plan-auditor verification) |

The orchestrator's primary detection scope is patterns #1 and #2 (runtime); patterns #3 and #4 are caught at authoring time.

---

## §C — Migration Table (12 archived agents → replacement pattern)

The following table is the canonical migration reference. Each row documents the archived agent, its replacement pattern, and an example invocation that the orchestrator (or the user, after `AskUserQuestion` consultation) can adopt.

| # | Archived agent | Archive type | Replacement pattern | Example invocation |
|---|----------------|--------------|---------------------|--------------------|
| 1 | `manager-strategy` | Critical archive — architecturally impossible chain | Absorb planning role into `manager-spec` (planning IS strategy per Anthropic 4-phase Step 2) | `Agent(subagent_type: "manager-spec", prompt: "... plan-phase strategic analysis: enumerate trade-offs, propose architecture, identify risks ...")` |
| 2 | `manager-quality` | Phantom archive (0 invocations) | Stop hook enforcement (lint + test + coverage delta) per the sync-phase quality gate policy; OR `/moai gate` skill for ad-hoc invocation | Hook auto-fires on sync-phase commit; OR user runs `Skill("moai", arguments: "gate")` |
| 3 | `manager-brain` | Phantom archive (0 invocations) | Sequential chain at the orchestrator (L1) level: `Explore` → `manager-spec` | First `Agent(subagent_type: "Explore", prompt: "... ideation domain exploration ...")`, then `Agent(subagent_type: "manager-spec", prompt: "... proposal synthesis from Explore findings ...")` |
| 4 | `manager-project` | Phantom archive (0 invocations) | Absorb into `manager-docs` (project docs ARE docs) | `Agent(subagent_type: "manager-docs", prompt: "... project doc rotation: product.md, structure.md, tech.md ...")` |
| 5 | `claude-code-guide` | Phantom archive (0 invocations) | Use Anthropic built-in `Explore` for upstream Claude Code investigation | `Agent(subagent_type: "Explore", prompt: "... fetch and analyze Claude Code v2.X.YY release notes, classify by impact tier ...")` |
| 6 | `researcher` | Phantom archive (0 invocations) | Use Anthropic built-in `Explore` for auto-research | `Agent(subagent_type: "Explore", prompt: "... auto-research the X library API surface, identify integration touchpoints ...")` |
| 7 | `expert-backend` | Domain-expert archive (per-spawn-prompt replacement per the per-spawn specialization policy) | `Agent(general-purpose, model: opus, tools: <backend whitelist>, prompt: <domain instructions>)` injected at delegation time, not embedded in an agent file | `Agent(subagent_type: "general-purpose", model: "opus", tools: "Read, Write, Edit, Bash, Grep, Glob", prompt: "... You are a backend specialist. Apply these backend conventions: [project-specific REST patterns, error wrapping, structured logging] ...")` |
| 8 | `expert-frontend` | Domain-expert archive | Same per-spawn pattern with frontend whitelist + frontend instructions | `Agent(subagent_type: "general-purpose", model: "opus", tools: "Read, Write, Edit, Bash, Grep, Glob", prompt: "... You are a frontend specialist. Apply React 19 / Next.js 16 patterns: [component composition, server actions, accessibility] ...")` |
| 9 | `expert-security` | Domain-expert archive | Stop hook dependency-manifest audit on `go.mod` / `package-lock.json` / `Pipfile.lock` / `Cargo.lock` changes (per the dependency-manifest audit policy); OR per-spawn for code-review tasks | Hook auto-fires on dependency manifest change with `govulncheck` / `npm audit` / equivalent; OR `Agent(subagent_type: "general-purpose", model: "opus", tools: "Read, Grep, Bash", prompt: "... You are a security reviewer. Apply OWASP Top 10 checks: [input validation, authentication, secrets management] ...")` |
| 10 | `expert-devops` | Domain-expert archive | Same per-spawn pattern with devops whitelist + devops instructions | `Agent(subagent_type: "general-purpose", model: "opus", tools: "Read, Write, Edit, Bash, Grep, Glob", prompt: "... You are a devops specialist. Apply CI/CD patterns: [GitHub Actions, deployment pipelines, infrastructure-as-code] ...")` |
| 11 | `expert-performance` | Domain-expert archive | Same per-spawn pattern with performance whitelist + perf instructions | `Agent(subagent_type: "general-purpose", model: "opus", tools: "Read, Bash, Grep, Glob", prompt: "... You are a performance specialist. Profile hot paths, propose optimizations, document benchmark deltas ...")` |
| 12 | `expert-refactoring` | Domain-expert archive | Same per-spawn pattern with refactoring whitelist + refactoring instructions | `Agent(subagent_type: "general-purpose", model: "opus", tools: "Read, Write, Edit, Grep, Glob, Bash", prompt: "... You are a refactoring specialist. Apply ANALYZE-PRESERVE-IMPROVE per spec-workflow.md DDD Mode ...")` |

### §C.1 Why per-spawn prompt injection

Anthropic's recommended pattern for domain expertise is per-spawn parameter injection rather than a static agent file. The 6 `expert-*` archives (#7-#12) all share the same root cause: 0 invocations across the recent 4-SPEC Epic, indicating that the "define when keep spawning the same kind of worker with the same instructions" criterion failed. Domain knowledge is therefore better surfaced in the active conversation context (the orchestrator composes the spawn prompt with domain-specific instructions tailored to the current task) than trapped inside individual agent definition files that are loaded only when explicitly invoked.

### §C.2 Why hook-based enforcement

`manager-quality` (#2) and `expert-security` (#9) overlap with Anthropic's published hook patterns (Stop, PostToolUse, SubagentStop, TaskCompleted). The Stop hook (`sync-phase-quality-gate.sh` per the canonical hook placement policy) mechanically enforces lint + test + coverage-delta + dependency-manifest audit at every sync-phase commit, which is more reliable than orchestrator-discipline phantom-agent spawn calls that historically went unfulfilled.

### §C.3 `claude-code-guide` — MoAI-archived custom file vs official built-in (disambiguation)

Row 5 above archives the MoAI-custom `claude-code-guide` agent **file** (a former `.claude/agents/` definition with 0 invocations). This is a DIFFERENT entity from the official Claude Code built-in helper agent that is also named `claude-code-guide` and ships with the Claude Code runtime. The built-in is a valid, current Anthropic-provided agent — it is NOT archived and invoking it does NOT trigger `ARCHIVED_AGENT_REJECTED`. The rejection rule binds only the MoAI-custom file by that name; the same-named runtime built-in is a separate, valid agent the orchestrator may use. (The archived-history row for the custom `claude-code-guide` file is retained above as a historical record and MUST NOT be removed.)

---

## §D — Orchestrator Recovery Flow

When the orchestrator detects an `ARCHIVED_AGENT_REJECTED` condition, it follows this recovery flow:

```
[Detection]
  ├── Direct Agent() spawn attempt    → Halt spawn
  ├── Paste-ready resume reference    → Halt resume execution at the affected step
  └── Other                            → Emit info notice; continue

[Recovery]
  Step 1: ToolSearch(query: "select:AskUserQuestion")
  Step 2: Compose AskUserQuestion round:
          - question: "<violator> is archived. Select replacement:"
          - options (max 4, first option "(권장)" / "(Recommended)"):
              (a) <canonical replacement from migration table §C> (권장)
              (b) <alternative pattern if applicable>
              (c) Continue with manual handling
              (d) Abort
  Step 3: AskUserQuestion — collect user response
  Step 4: Construct fresh Agent() spawn (or revised paste-ready execution) with the user's choice
  Step 5: Re-attempt the spawn
```

When a single-owner unambiguous replacement exists (rows #2, #5, #6 — Stop hook / Explore / Explore), Step 2-4 MAY be replaced with an informational substitution log:

```
[archived-agent-substitution] manager-quality → Stop hook (sync-phase-quality-gate.sh)
[archived-agent-substitution] claude-code-guide → Explore
[archived-agent-substitution] researcher → Explore
```

For all other rows (multi-option replacements), `AskUserQuestion` is mandatory.

---

## §E — Anti-Patterns

The following patterns violate the archived-agent rejection contract:

- **Silent substitution for multi-option replacements** — rows #1, #3, #4, #7-#12 require `AskUserQuestion`; silently switching `manager-strategy` → `manager-spec` without user confirmation can mis-frame the task scope (planning vs strategic decomposition)
- **Re-introducing archived agents in template** — `internal/template/templates/.claude/agents/moai/<archived>.md` MUST NOT contain any of the 12 archived files; re-introduction triggers a template-leak audit failure
- **Paste-ready resume → autonomous execution without Implementation Kickoff Approval** — when a paste-ready resume references an archived agent and the orchestrator silently substitutes, the user loses the Implementation Kickoff Approval plan-to-implement gate signal (per the Implementation Kickoff Approval mandatory restoration policy)
- **Skipping the `ToolSearch` preload before `AskUserQuestion`** — `AskUserQuestion` is a deferred tool; the orchestrator MUST `ToolSearch(query: "select:AskUserQuestion")` per `.claude/rules/moai/core/askuser-protocol.md` § ToolSearch Preload Procedure
- **Returning an `AskUserQuestion` invocation from a subagent body** — subagents MUST NOT prompt the user; this is the orchestrator's responsibility (CLAUDE.md §8 + agent-common-protocol.md § User Interaction Boundary). If a subagent encounters an archived-agent reference in its scope, it returns a structured blocker report and the orchestrator runs the recovery flow
- **Updating an archived agent's body content** — archived agents in the offline migration backup location are read-only historical preservation; future revival requires a dedicated revival SPEC, not a body modification

---

## §F — Cross-References

- The agent catalog consolidation policy's retain-vs-archive matrix — design-time SSOT for the 8-retain (at decision time; now 10 per CLAUDE.md §4) / 12-archive decision with per-agent rationale and Anthropic citations
- The agent catalog consolidation policy's design-level migration table (this rule's §C is the canonical runtime SSOT)
- The offline migration backup location's README — archive backup directory with per-agent README entries
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` — sibling rule documenting the 5-mode autonomous selection at Phase 4 (independent of archived-agent rejection)
- `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary — orchestrator-subagent boundary (subagents return blocker reports; orchestrator runs `AskUserQuestion`)
- `.claude/rules/moai/core/askuser-protocol.md` § ToolSearch Preload Procedure — deferred tool preload contract
- `.claude/rules/moai/development/agent-patterns.md` § Per-Spawn Domain Specialization — canonical per-spawn-prompt pattern documentation
- `.claude/rules/moai/development/agent-patterns.md` § Read-only Investigation — `Explore` canonical reference (read-only investigation pattern for rows #5, #6)
- The agent catalog consolidation policy's acceptance criteria — verification commands for this rule's §A error token + §C migration table coverage
- Anthropic Sub-agents documentation — *"Subagents cannot spawn other subagents. If your workflow requires nested delegation, use Skills or chain subagents from the main conversation."*
- Anthropic best-practices documentation — *"Define a custom subagent when you keep spawning the same kind of worker with the same instructions."*

---

Version: 1.0.0
Origin: the canonical agent catalog consolidation policy
Status: Active — applies to all `Agent()` spawn attempts and paste-ready resume executions
