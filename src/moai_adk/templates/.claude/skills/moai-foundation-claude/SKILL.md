---
name: moai-foundation-claude
aliases: [moai-foundation-claude]
category: foundation
description: Canonical Claude Code authoring kit covering Skills, sub-agents, plugins, slash commands, hooks, memory, settings, sandboxing, headless mode, and advanced agent patterns. Use when creating Claude Code extensions or configuring Claude Code features.
version: 5.0.0
modularized: false
user-invocable: false
allowed-tools:
  - Read
  - Write
  - Edit
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
tags: ['foundation', 'claude-code', 'skills', 'sub-agents', 'plugins', 'slash-commands', 'hooks', 'memory', 'settings', 'sandboxing', 'headless', 'agent-patterns']
---

# Claude Code Authoring Kit

Comprehensive reference for Claude Code Skills, sub-agents, plugins, slash commands, hooks, memory, settings, sandboxing, headless mode, and advanced agent patterns.

## Documentation Index

Core Features:
- [Skills Guide](reference/claude-code-skills-official.md) - Agent Skills creation and management
- [Sub-agents Guide](reference/claude-code-sub-agents-official.md) - Sub-agent development and delegation
- [Plugins Guide](reference/claude-code-plugins-official.md) - Plugin architecture and distribution
- [Slash Commands](reference/claude-code-custom-slash-commands-official.md) - Command creation and orchestration

Configuration:
- [Settings](reference/claude-code-settings-official.md) - Configuration hierarchy and management
- [Memory](reference/claude-code-memory-official.md) - Context and knowledge persistence
- [Hooks](reference/claude-code-hooks-official.md) - Event-driven automation
- [IAM & Permissions](reference/claude-code-iam-official.md) - Access control and security

Advanced Features:
- [Sandboxing](reference/claude-code-sandboxing-official.md) - Security isolation
- [Headless Mode](reference/claude-code-headless-official.md) - Programmatic and CI/CD usage
- [Dev Containers](reference/claude-code-devcontainers-official.md) - Containerized environments
- [CLI Reference](reference/claude-code-cli-reference-official.md) - Command-line interface
- [Statusline](reference/claude-code-statusline-official.md) - Custom status display
- [Advanced Patterns](reference/advanced-agent-patterns.md) - Engineering best practices

## Quick Reference (30 seconds)

Skills: Model-invoked extensions in ~/.claude/skills/ (personal) or .claude/skills/ (project). Three-level progressive disclosure. Max 500 lines.

Sub-agents: Specialized assistants via Task(subagent_type="..."). Own 200K context. Cannot spawn sub-agents. Use /agents command.

Plugins: Reusable bundles in .claude-plugin/plugin.json. Include commands, agents, skills, hooks, MCP servers.

Commands: User-invoked via /command. Parameters: $ARGUMENTS, $1, $2. File refs: @file.

Hooks: Events in settings.json. PreToolUse, PostToolUse, SessionStart, SessionEnd, PreCompact, Notification.

Memory: CLAUDE.md files + .claude/rules/*.md. Enterprise to Project to User hierarchy. @import syntax.

Settings: 6-level hierarchy. Managed to file-managed to CLI to local to shared to user.

Sandboxing: OS-level isolation. Filesystem and network restrictions. Auto-allow safe operations.

Headless: -p flag for non-interactive. --allowedTools, --json-schema, --agents for automation.

## Skill Creation (3 minutes)

### Progressive Disclosure Architecture

Level 1 (Metadata): Name and description loaded at startup, approximately 100 tokens per Skill

Level 2 (Instructions): SKILL.md body loaded when triggered, under 5K tokens recommended

Level 3 (Resources): Additional files loaded on demand, effectively unlimited

### Required Format

```yaml
---
name: skill-name
description: What it does AND when to use it. Third person. Max 1024 chars.
---

# Skill Name

## Quick start
Brief instructions here.

## Details
See [REFERENCE.md](REFERENCE.md) for more.
```

### Best Practices

- Third person descriptions (does not I do)
- Include trigger terms users mention
- Keep under 500 lines
- One level deep references
- Test with Haiku, Sonnet, Opus

## Sub-agent Creation (3 minutes)

### Using /agents Command

1. Type /agents
2. Select Create New Agent
3. Define purpose and tools
4. Press e to edit prompt

### File Format

```yaml
---
name: agent-name
description: When to invoke. Use PROACTIVELY for auto-delegation.
tools: Read, Write, Bash
model: sonnet
---

System prompt here.
```

### Critical Rules

- Cannot spawn other sub-agents
- Cannot use AskUserQuestion effectively
- All user interaction before delegation
- Each gets own 200K context

## Plugin Creation (3 minutes)

### Directory Structure

```
my-plugin/
- .claude-plugin/plugin.json
- commands/
- agents/
- skills/
- hooks/hooks.json
- .mcp.json
```

### Manifest (plugin.json)

```json
{
  "name": "my-plugin",
  "description": "Plugin purpose",
  "version": "1.0.0",
  "author": {"name": "Author"}
}
```

### Commands

/plugin install owner/repo
/plugin validate .
/plugin enable plugin-name

## Advanced Agent Patterns

### Two-Agent Pattern for Long Tasks

Initializer agent: Sets up environment, feature registry, progress docs

Executor agent: Works single features, updates registry, maintains progress

See [Advanced Patterns](reference/advanced-agent-patterns.md) for details.

### Orchestrator-Worker Architecture

Lead agent: Decomposes tasks, spawns workers, synthesizes results

Worker agents: Execute focused tasks, return condensed summaries

### Context Engineering Principles

- Smallest set of high-signal tokens
- Just-in-time retrieval over upfront loading
- Context compaction for long sessions
- External memory files persist outside window

### Tool Design Best Practices

- Consolidate related functions into single tools
- Return high-signal context-aware responses
- Clear parameter names (user_id not user)
- Instructive error messages with examples

## Workflow: Explore-Plan-Code-Commit

Phase 1 Explore: Read files, understand structure, map dependencies

Phase 2 Plan: Use think prompts, outline approach, define criteria

Phase 3 Code: Implement iteratively, verify each step, handle edges

Phase 4 Commit: Descriptive messages, logical groupings, clean history

## MoAI-ADK Integration

### Core Skills

- moai-foundation-claude: This authoring kit
- moai-foundation-core: SPEC system and workflows
- moai-foundation-philosopher: Strategic thinking

### Essential Sub-agents

- spec-builder: EARS specifications
- manager-tdd: TDD execution
- expert-security: Security analysis
- expert-backend: API development
- expert-frontend: UI implementation

## Security Features

### Sandboxing

- Filesystem: Write restricted to cwd
- Network: Domain allowlists via proxy
- OS-level: bubblewrap (Linux), Seatbelt (macOS)

### Dev Containers

- Security-hardened with firewall
- Whitelisted outbound only
- --dangerously-skip-permissions for trusted only

### Headless Safety

- Always use --allowedTools in CI/CD
- Validate inputs before passing to Claude
- Handle errors with exit codes

## Version History

v4.0.0 (2026-01-06): Added plugins, sandboxing, headless, statusline, dev containers, CLI reference, advanced patterns from engineering blogs

v3.0.0 (2025-12-06): Added progressive disclosure, sub-agent details, integration patterns

v2.0.0 (2025-11-26): Initial comprehensive release

