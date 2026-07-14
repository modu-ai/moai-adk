---
title: Subagents
weight: 10
draft: false
description: "An overview of Claude Code subagents — the concept, isolated-context delegation, and how to define them."
---

# Subagents

A Claude Code subagent is a delegated worker that handles side tasks in a separate context window and returns only a summary of the results to the main conversation.

{{< callout type="info" >}}
**One-line summary**: A subagent is a delegated worker that handles side jobs like exploration and verification in its own context and returns only a summary, keeping the main conversation clean.
{{< /callout >}}

{{< callout type="tip" >}}
This page is a Claude Code-level conceptual overview. How MoAI-ADK composes and delegates its 11-agent catalog (10 MoAI-custom + 1 Anthropic built-in `Explore`), and the hands-on way to build your own agents, are covered in depth in the [Agent Guide](/advanced/agent-guide) and the [Builder Agents Guide](/advanced/builder-agents).
{{< /callout >}}

## What Is a Subagent

A subagent is a specialized AI worker dedicated to a particular kind of task. When a side task arises that would flood the main conversation with search results, logs, and file contents, the subagent handles it in **its own context window** and returns only a summary.

Each subagent independently has:

| Component | Description |
|-----------|------|
| System prompt | The subagent file's body becomes the role instructions verbatim |
| Tool access | Usable tools can be restricted via allow/deny lists |
| Independent permissions | Inherits the main conversation's permissions but can add restrictions |
| Model choice | Cost can be lowered with a fast, cheap model like `haiku` |

Claude looks at each subagent's `description` to decide when to delegate. Writing a clear description is therefore the starting point of good delegation.

Claude Code includes these built-in subagents:

| Agent | Characteristics |
|---------|------|
| **Explore** | Read-only codebase exploration (Haiku, fast); the thoroughness option offers quick/medium/very-thorough |
| **Plan** | Plan-mode research (read-only) |
| **general-purpose** | Access to all tools; can both explore and modify |

Explore and Plan skip the main session's CLAUDE.md and git status, running faster and lighter.

## The Core Constraint: Subagents Cannot Spawn Subagents

This is the most important structural constraint. **Subagents cannot spawn other subagents.** Delegation descends only one level from the main conversation, and infinite nesting cannot occur.

### After v2.1.172: Limited Nesting (Depth-5 Cap)

Starting with Claude Code v2.1.172, **conditional subagent nesting** is possible. There is a configuration knob.

| Setting | Behavior | Usage |
|------|------|------|
| Include `Agent` in the subagent definition (frontmatter `tools:` list) | Nesting allowed | Up to depth 5 (hard cap) |
| Omit the `Agent` tool | Nesting prohibited | Flat orchestration only |

This constraint is also the bedrock of MoAI-ADK's orchestration design. **Only the orchestrator (the main session) invokes subagents**, and an invoked agent may delegate again only if it does not hit the depth limit. So instead of hierarchical agent chains, the design follows a flat structure where **the orchestrator directly calls each stage** (MoAI's base principle).

```mermaid
flowchart TD
    M[Main conversation<br/>Orchestrator] --> A[Subagent A<br/>Exploration]
    M --> B[Subagent B<br/>Verification]
    M --> C[Subagent C<br/>Implementation]
    A -.->|Condition: with the Agent tool<br/>only up to depth 5| X["Nested subagent<br/>(limited)"]
    style X fill:#ffd,stroke:#c80
```

This is also why the built-in `Plan` subagent exists separately: to perform research in plan mode when context is needed, without circumventing this constraint.

## Background Permission Prompts (v2.1.186)

When running a subagent in the background (`background: true`) and it encounters a tool requiring permission (e.g., Bash, WebFetch):

- **Before v2.1.186**: auto-denied (no permission prompt)
- **From v2.1.186**: **the prompt appears in the main session** (Esc denies just that call)

So before starting long background work, it is best to pre-add the needed tools to the allow list in `settings.json`.

## When to Use Them

Subagents are most effective in situations like these:

| Situation | Effect |
|------|------|
| Parallel exploration | Investigate multiple files/directories simultaneously and collect only summaries |
| Independent verification | Check results in a separate context, free of the main conversation's bias |
| Context isolation | Quarantine large logs and search results away from the main conversation |
| Cost control | Route simple work to a fast model like `haiku` |

Conversely, for work that finishes in a single response, or multi-stage work that **needs shared context**, handling it directly in the main conversation without delegation is better.

## Defining Subagents — Overview

A subagent is defined as a markdown file with YAML frontmatter. You can create one interactively with the `/agents` command or write the file directly.

```markdown
---
name: code-reviewer
description: Reviews code quality and best practices
tools: Read, Glob, Grep
model: sonnet
---

You are a code reviewer. When invoked, analyze the code and
provide specific, actionable feedback on quality, security, and best practices.
```

### Required Fields

- `name` — the subagent's name (referenced when delegating)
- `description` — explains when to delegate (Claude judges from this alone)

### Optional Fields

| Field | Function |
|------|------|
| `tools` | Allowed tools (comma-separated list) |
| `disallowedTools` | Blocked tools (usable instead of an allowlist) |
| `model` | Model choice: `sonnet`, `opus`, `haiku`, `fable`, or a specific model ID; default `inherit` (the main session model) |
| `permissionMode` | Tool permission default (default, plan, acceptEdits, bypass) |
| `maxTurns` | Maximum turn limit |
| `skills` | Default skills to load |
| `mcpServers` | MCP servers to connect |
| `hooks` | Hook events to invoke |
| `memory` | Memory scope (user, project, local) |
| `background` | If `true`, runs in the background |
| `effort` | Reasoning intensity (low, medium, high, xhigh, max) |
| `isolation: worktree` | Works in an isolated copy of the repository |
| `color` | Color shown in the agent view |
| `initialPrompt` | The prompt used when the subagent is first spawned |

Where the file lives determines its scope.

| Location | Scope |
|------|------|
| `.claude/agents/` | The current project (put under version control to share with the team) |
| `~/.claude/agents/` | All my projects |
| A plugin's `agents/` | Wherever the plugin is enabled |

### AskUserQuestion Unavailable

User-interaction tools like `AskUserQuestion` cannot be used inside subagents (an asymmetric boundary). This is why, in MoAI-ADK, subagents cannot question the user directly and instead return a blocker report to the orchestrator.

## `/fork` — Session Fork

The `/fork <directive>` command forks the current session. The forked subagent:

- Inherits the current conversation contents
- Leverages the parent's prompt cache
- Explores in a new direction

## Going Deeper via the MoAI Agent Guide

That covers the Claude Code-level subagent concept. MoAI-ADK operates an **11-agent catalog** on top of this mechanism — the Manager family (manager-spec / manager-develop / manager-docs / manager-git / manager-design) handles the plan→run→sync lifecycle, the Evaluator family (plan-auditor / sync-auditor) handles independent audits, builder-harness handles harness scaffold generation, super-advisor handles high-reasoning consultation, e2e-tester handles E2E test execution across web/mobile/desktop, and the Anthropic built-in `Explore` handles read-only exploration. The separation of planning and auditing — the agent that built something never checks its own work — is the core design of this catalog. Declaratively assigning each agent the model and reasoning depth (effort) matched to its task is the tokenomics principle of "plan deeply, implement cheaply, verify independently." Details are covered in the advanced guides below.

## Related Documents

- [Agent Guide](/advanced/agent-guide)
- [Builder Agents Guide](/advanced/builder-agents)

## References

- [Create custom subagents (Claude Code official docs)](https://code.claude.com/docs/en/sub-agents)

{{< callout type="tip" >}}
When creating a subagent, write the `description` concretely from the perspective of "when should this be delegated to?" Claude decides delegation from this description alone — if it is vague, a good tool goes uncalled.
{{< /callout >}}
