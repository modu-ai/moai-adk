---
title: Advanced
weight: 10
draft: false
---
# Advanced

Covers MoAI-ADK's internal structure and advanced features in depth.

{{< callout type="info" >}}
This section is a guide for developers who understand MoAI-ADK's basic concepts and want to grasp internal operating principles.
{{< /callout >}}

## Learning Structure

MoAI-ADK consists of 7 core components:

```mermaid
flowchart TD
    CLAUDE["CLAUDE.md<br/>Project Constitution"] --> SETTINGS["settings.json<br/>Permissions & Environment"]
    CLAUDE --> RULES[".claude/rules/<br/>Conditional Rules"]

    SETTINGS --> HOOKS["Hooks<br/>Event Automation"]
    SETTINGS --> MCP["MCP Servers<br/>External Tool Connections"]

    RULES --> SKILLS["Skills<br/>Expert Knowledge Modules"]
    SKILLS --> AGENTS["Agents<br/>Expert Agents"]

    AGENTS --> BUILDERS["Builder Agents<br/>Extension Creators"]
```

## Table of Contents

| Topic | Description |
|-------|-------------|
| [Skill Guide](/advanced/skill-guide) | Skill system that grants expert knowledge to AI |
| [Agent Guide](/advanced/agent-guide) | Specialized AI task executor system |
| [Builder Agent Guide](/advanced/builder-agents) | Creating skills, agents, commands, plugins |
| [Hooks Guide](/advanced/hooks-guide) | Event-based automation scripts |
| [settings.json Guide](/advanced/settings-json) | Claude Code global settings management |
| [CLAUDE.md Guide](/advanced/claude-md-guide) | Project guideline file system |
| [MCP Servers](/advanced/mcp-servers) | External tool connection protocol |
| [Google Stitch Guide](/advanced/stitch-guide) | AI-based UI/UX design generation tool |

{{< callout type="info" >}}
Each document can be read independently, but reading sequentially starting from **Skill Guide** allows you to systematically understand the entire architecture.
{{< /callout >}}
