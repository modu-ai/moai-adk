---
title: Subagents
weight: 10
draft: false
description: "Claude Code subagents are specialized workers for isolated tasks. Learn their constraints, optional fields (v2.1.172+), v2.1.186 background-mode permissions, and when to delegate."
---

Claude Code subagents are delegated workers that handle side-tasks in a separate context window and return only a result summary to the main conversation.

{{< callout type="info" >}}
**TL;DR**: A subagent is a delegated worker that handles side-tasks such as exploration and verification in its own context, returning only a summary — keeping the main conversation clean.
{{< /callout >}}

{{< callout type="tip" >}}
This page is a concept overview at the Claude Code level. How MoAI-ADK organizes its 8-agent catalog and delegates work, and the hands-on approach to building your own agents, are covered in depth in the [Agent Guide](/advanced/agent-guide) and the [Builder Agent Guide](/advanced/builder-agents).
{{< /callout >}}

## What Is a Subagent

A subagent is a specialized AI worker dedicated to a particular kind of task. When a side-task arises that would otherwise flood the main conversation with search results, logs, and file contents, the subagent handles it in its **own context window** and returns only a result summary.

Each subagent independently owns the following.

| Component | Description |
|-----------|-------------|
| System prompt | The body of the subagent file becomes its role instructions verbatim |
| Tool access | The tools it can use can be restricted via allow/deny lists |
| Independent permissions | It inherits the main conversation's permissions but can add further restrictions |
| Model selection | Cost can be lowered by using a fast, inexpensive model such as `haiku` |

Claude decides when to delegate by reading each subagent's `description`. Writing that description clearly is therefore the starting point for good delegation.

Claude Code includes built-in subagents such as `Explore` (read-only codebase exploration with thoroughness options: quick/medium/very-thorough), `Plan` (plan-mode research), and `general-purpose` (combined exploration + modification tasks).

## Core Constraint: Subagent Nesting Depth (v2.1.172+)

The most important structural constraint is **nesting depth**. Subagents can spawn other subagents, but subject to a **hard depth limit of 5 levels**.

### Depth Configuration

| Setting | Behavior | Enabled when |
|---------|----------|--------------|
| With `Agent` tool included | Nested spawning allowed | Frontmatter `tools:` field includes `Agent` |
| Without `Agent` tool | No nested spawning | `Agent` tool omitted (or `disallowedTools:` blocks it) |

This constraint is also the foundation of MoAI-ADK's orchestration design. Only the orchestrator (the main session) directly spawns subagents, and an invoked agent at depth 4 cannot spawn further (depth-5 cap). As a result, instead of a hierarchical agent chain, MoAI-ADK follows a flat structure in which **the orchestrator invokes each step directly**.

```mermaid
flowchart TD
    M[Main conversation<br/>Orchestrator] --> A[Subagent A<br/>depth 1]
    M --> B[Subagent B<br/>depth 1]
    M --> C[Subagent C<br/>depth 1]
    A -.->|Optional (depth ≤ 4)<br/>Requires Agent tool| X["Nested subagent<br/>depth 2"]
    style X fill:#ffd,stroke:#c80
```

The built-in `Plan` subagent exists separately for a reason: to perform research when plan mode needs context, without hitting the depth limit.

## Background Mode Permissions (v2.1.186+)

Subagents can run in the background (`background: true`). When a background subagent needs a permission for a tool like Bash or WebFetch:

- **v2.1.186 and later**: The permission prompt surfaces in the main session (the user can press Esc to deny that one call only)
- **Before v2.1.186**: The call was automatically rejected

To avoid mid-run permission prompts when running long background tasks, pre-add needed tools to the allowlist in `settings.json`.

## When to Use One

Subagents are most effective in situations like these.

| Situation | Benefit |
|-----------|---------|
| Parallel exploration | Investigate multiple files and directories simultaneously, collect only the summaries |
| Independent verification | Check results in a separate context, free of the main conversation's bias |
| Context isolation | Quarantine large logs and search results away from the main conversation |
| Cost control | Route simple tasks to a fast model such as `haiku` |

Conversely, if a task finishes in a single response, or if it spans multiple steps that **require shared context**, it is better to handle it directly in the main conversation without delegation.

## Definition Overview

A subagent is defined as a Markdown file with YAML frontmatter. You can create one interactively with the `/agents` command, or write the file directly.

```markdown
---
name: code-reviewer
description: Reviews code quality and best practices
tools: Read, Glob, Grep
model: sonnet
---

You are a code reviewer. When invoked, you analyze code and provide
concrete, actionable feedback on quality, security, and best practices.
```

### Required Fields

- `name` — The subagent's identifier (used when delegating)
- `description` — When to delegate (Claude reads only this to decide whether to invoke the agent)

### Optional Fields

| Field | Type | Purpose |
|-------|------|---------|
| `tools` | CSV string | Allow-list of tools (e.g., `Read, Glob, Grep`) |
| `disallowedTools` | CSV string | Deny-list of tools (alternative to `tools:`) |
| `model` | string | Model selection: `sonnet`, `opus`, `haiku`, `fable`, or specific model ID; default `inherit` |
| `permissionMode` | enum | Default permissions (default, plan, acceptEdits, bypass) |
| `maxTurns` | integer | Maximum turn limit for this agent |
| `skills` | list | Skills to load by default |
| `mcpServers` | list | MCP servers to connect |
| `hooks` | list | Hook events to invoke |
| `memory` | enum | Memory scope (user, project, local) |
| `background` | bool | Run in the background (true/false) |
| `effort` | enum | Reasoning effort (low, medium, high, xhigh, max) |
| `isolation: worktree` | string | Run in an isolated worktree copy of the repository |
| `color` | string | Color shown in agent view |
| `initialPrompt` | string | Prompt to send when the subagent is first spawned |

Scope varies based on where the file is stored.

| Location | Scope |
|----------|-------|
| `.claude/agents/` | Current project (include in version control to share with the team) |
| `~/.claude/agents/` | All of your projects |
| A plugin's `agents/` | Wherever the plugin is enabled |

### AskUserQuestion Unavailable in Subagents

User-interaction tools such as `AskUserQuestion` cannot be used in a subagent. This is why, in MoAI-ADK, a subagent cannot ask the user directly and instead returns a **blocker report** to the orchestrator, which then asks the user via `AskUserQuestion`.

## /fork — Session Forking

The `/fork <directive>` command lets you fork the current session's context into a new subagent-like context. The forked subagent:

- Inherits the current conversation content
- Leverages the parent session's prompt cache
- Explores in a new direction independently

## For Depth, See the MoAI Agent Guide

That covers the subagent concept at the Claude Code level. How MoAI-ADK operates its 8-agent catalog on top of this mechanism, how it delegates each stage of the Plan-Run-Sync workflow, and how it generates project-specific domain-expert agents are covered in the advanced guides below.

## Related Docs

- [Agent Guide](/advanced/agent-guide)
- [Builder Agent Guide](/advanced/builder-agents)

## References

- [Create custom subagents (Claude Code official docs)](https://code.claude.com/docs/en/sub-agents)

{{< callout type="tip" >}}
When you create a subagent, write the `description` concretely from the perspective of "when delegation should happen." Claude decides whether to delegate based solely on this description, so if it is vague, even a good tool may never be invoked.
{{< /callout >}}
