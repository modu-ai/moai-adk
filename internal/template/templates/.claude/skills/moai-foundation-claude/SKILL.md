---
name: moai-foundation-claude
description: >
  Canonical Claude Code authoring kit covering Skills, sub-agents, plugins, slash commands,
  hooks, memory, settings, sandboxing, headless mode, and advanced agent patterns.
  Use when creating Claude Code extensions or configuring Claude Code features.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "5.0.0"
  category: "foundation"
  status: "active"
  updated: "2026-01-11"
  modularized: "false"
  tags: "foundation, claude-code, skills, sub-agents, plugins, slash-commands, hooks, memory, settings, sandboxing, headless, agent-patterns"
  aliases: "moai-foundation-claude"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords:
    - "skill"
    - "agent"
    - "plugin"
    - "slash command"
    - "hook"
    - "sandbox"
    - "headless"
    - "memory"
    - "settings"
    - "claude code"
    - "sub-agent"
    - "agent pattern"
    - "orchestration"
    - "delegation"
  agents:
    - "builder-agent"
    - "builder-skill"
    - "builder-plugin"
  phases:
    - "plan"
    - "run"
    - "sync"
---

# Claude Code Authoring Kit

Comprehensive reference for Claude Code Skills, sub-agents, plugins, slash commands, hooks, memory, settings, sandboxing, headless mode, and advanced agent patterns.

## Documentation Index

Core Features:

- reference/claude-code-skills-official.md - Agent Skills creation and management
- reference/claude-code-sub-agents-official.md - Sub-agent development and delegation
- reference/claude-code-plugins-official.md - Plugin architecture and distribution
- reference/claude-code-custom-slash-commands-official.md - Command creation and orchestration

Configuration:

- reference/claude-code-settings-official.md - Configuration hierarchy and management
- reference/claude-code-memory-official.md - Context and knowledge persistence
- reference/claude-code-hooks-official.md - Event-driven automation
- reference/claude-code-iam-official.md - Access control and security

Advanced Features:

- reference/claude-code-sandboxing-official.md - Security isolation
- reference/claude-code-headless-official.md - Programmatic and CI/CD usage
- reference/claude-code-devcontainers-official.md - Containerized environments
- reference/claude-code-cli-reference-official.md - Command-line interface
- reference/claude-code-statusline-official.md - Custom status display
- reference/advanced-agent-patterns.md - Engineering best practices

## Quick Reference

Skills: Model-invoked extensions in ~/.claude/skills/ (personal) or .claude/skills/ (project). Three-level progressive disclosure. Max 500 lines.

Sub-agents: Specialized assistants via Task(subagent_type="..."). Own 200K context. Cannot spawn sub-agents. Use /agents command.

Plugins: Reusable bundles in .claude-plugin/plugin.json. Include commands, agents, skills, hooks, MCP servers.

Commands: User-invoked via /command. Parameters: $ARGUMENTS, $1, $2. File refs: @file.

Hooks: Events in settings.json. PreToolUse, PostToolUse, SessionStart, SessionEnd, PreCompact, Notification.

Memory: CLAUDE.md files + .claude/rules/*.md. Enterprise to Project to User hierarchy. @import syntax.

Settings: 6-level hierarchy. Managed to file-managed to CLI to local to shared to user.

Sandboxing: OS-level isolation. Filesystem and network restrictions. Auto-allow safe operations.

Headless: -p flag for non-interactive. --allowedTools, --json-schema, --agents for automation.

## Skill Creation

### Progressive Disclosure Architecture

Level 1 (Metadata): Name and description loaded at startup, approximately 100 tokens per Skill

Level 2 (Instructions): SKILL.md body loaded when triggered, under 5K tokens recommended

Level 3 (Resources): Additional files loaded on demand, effectively unlimited

### Required Format

Create a SKILL.md file with YAML frontmatter containing name in kebab-case and description explaining what it does and when to use it in third person. Maximum 1024 characters for description. After the frontmatter, include a heading with the skill name, a Quick Start section with brief instructions, and a Details section referencing REFERENCE.md for more information.

### Best Practices

- Third person descriptions (does not I do)
- Include trigger terms users mention
- Keep under 500 lines
- One level deep references
- Test with Haiku, Sonnet, Opus

## Sub-agent Creation

### Using /agents Command

Type /agents, select Create New Agent, define purpose and tools, press e to edit prompt.

### File Format

Create a markdown file with YAML frontmatter containing name, description explaining when to invoke (use PROACTIVELY for auto-delegation), tools as comma-separated list (Read, Write, Bash), and model specification (sonnet). After frontmatter, include the system prompt.

### Critical Rules

- Cannot spawn other sub-agents
- Cannot use AskUserQuestion effectively
- All user interaction before delegation
- Each gets own 200K context

## Plugin Creation

### Directory Structure

Create my-plugin directory with .claude-plugin/plugin.json, commands directory, agents directory, skills directory, hooks/hooks.json, and .mcp.json file.

### Manifest (plugin.json)

Create a JSON object with name, description explaining plugin purpose, version as 1.0.0, and author object containing name field.

### Commands

Use /plugin install owner/repo to install from GitHub.
Use /plugin validate . to validate current directory.
Use /plugin enable plugin-name to enable a plugin.

## Advanced Agent Patterns

### Two-Agent Pattern for Long Tasks

Initializer agent: Sets up environment, feature registry, progress docs

Executor agent: Works single features, updates registry, maintains progress

See reference/advanced-agent-patterns.md for details.

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

### Explore/Search Performance Optimization

When using Explore agent or direct exploration tools (Grep, Glob, Read), apply these optimizations to prevent performance bottlenecks with GLM models:

**AST-Grep Priority**
- Use structural search (ast-grep) before text-based search (Grep)
- Load moai-tool-ast-grep skill for complex pattern matching
- Example: `sg -p 'class $X extends Service' --lang python` is faster than `grep -r "class.*extends.*Service"`

**Search Scope Limitation**
- Always use `path` parameter to limit search scope
- Example: `Grep(pattern="func ", path="internal/core/")` instead of `Grep(pattern="async def")`

**File Pattern Specificity**
- Use specific Glob patterns instead of wildcards
- Example: `Glob(pattern="internal/core/*.go")` instead of `Glob(pattern="src/**/*.py")`

**Parallel Processing**
- Execute independent searches in parallel (single message, multiple tool calls)
- Maximum 5 parallel searches to prevent context fragmentation

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
- manager-ddd: DDD execution
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

## Resources

For detailed patterns and working examples, see the reference directory.

Version History:

- v5.0.0 (2026-01-11): Converted to narrative format per CLAUDE.md Documentation Standards
- v4.0.0 (2026-01-06): Added plugins, sandboxing, headless, statusline, dev containers, CLI reference, advanced patterns
- v3.0.0 (2025-12-06): Added progressive disclosure, sub-agent details, integration patterns
- v2.0.0 (2025-11-26): Initial comprehensive release
