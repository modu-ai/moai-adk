# Claude Code Reference Guide

> **Version**: 2.0.0 (2025-10-22)
> **Based on**: Claude Code Official Documentation (docs.claude.com), v0.6.0+

Complete technical reference for Claude Code agents, skills, plugins, hooks, and MCP servers.

---

## Table of Contents

1. [Claude Code CLI Commands](#claude-code-cli-commands)
2. [Skill Specification](#skill-specification)
3. [Agent Configuration](#agent-configuration)
4. [Hook Types and Events](#hook-types-and-events)
5. [Plugin Structure](#plugin-structure)
6. [MCP Server Integration](#mcp-server-integration)
7. [Settings Configuration](#settings-configuration)
8. [Progressive Disclosure Model](#progressive-disclosure-model)
9. [Tool Permissions](#tool-permissions)
10. [Troubleshooting Guide](#troubleshooting-guide)

---

## Claude Code CLI Commands

### Essential Commands

| Command | Description | Example |
|---------|-------------|---------|
| `claude` | Start interactive session | `claude` |
| `claude <prompt>` | Execute single prompt | `claude "Run tests"` |
| `/help` | Show available commands | `/help` |
| `/hooks` | Configure hooks | `/hooks` |
| `/plugin` | Manage plugins | `/plugin install <url>` |
| `/clear` | Clear conversation history | `/clear` |
| `/exit` | Exit Claude Code | `/exit` |

### Slash Commands

**Built-in**:
- `/help` - Show help information
- `/clear` - Clear conversation history
- `/hooks` - Manage hooks
- `/plugin` - Plugin management
- `/bashes` - List background shells
- `/compact` - Trigger context compaction

**Custom Commands**:
Custom slash commands are stored in `.claude/commands/` and auto-load when Claude Code starts.

**Creating Custom Commands**:
```bash
# Create command file
cat > .claude/commands/my-command.md << 'EOF'
# My Command

This command does [specific task].

## Usage

When user types `/my-command`:
1. Step 1
2. Step 2
3. Step 3
EOF
```

---

## Skill Specification

### YAML Frontmatter (Required)

```yaml
---
name: Skill Name
description: What it does and when to use it (include trigger keywords)
allowed-tools: Tool1, Tool2, Tool3  # Optional
---
```

### Frontmatter Fields

#### name (Required)

**Format**: String, ≤64 characters
**Examples**:
- ✅ "PDF Processing"
- ✅ "Kubernetes Deployment"
- ❌ "Helper" (too vague)

#### description (Required)

**Format**: String, ≤1024 characters
**Must Include**:
1. What the skill does (capabilities)
2. When to use it (trigger scenarios)
3. Keywords for discovery

**Template**:
```
[Capability 1], [Capability 2]. Use when [trigger 1], [trigger 2], or [trigger 3].
```

**Example**:
```yaml
description: Extract text and tables from PDF files, fill forms, merge documents. Use when working with PDF files or when the user mentions PDFs, document extraction, form filling, or PDF manipulation.
```

**Keywords to Include**:
- Problem domain: PDF, API, database, k8s
- Operations: extract, deploy, analyze, validate
- Tech stack: Python, TypeScript, Docker, AWS
- File types: .pdf, .yaml, .json

#### allowed-tools (Optional)

**Format**: Comma-separated tool names
**Purpose**: Restrict which tools Claude can use within this skill

**Available Tools**:
- `Read` - Read files
- `Write` - Create new files
- `Edit` - Modify existing files
- `Bash` - Execute shell commands
- `Bash(git:*)` - Only git commands
- `Bash(npm:*)` - Only npm commands
- `Bash(python:*)` - Only Python commands
- `Grep` - Search files
- `Glob` - Find files by pattern

**Examples**:
```yaml
# Read-only skill
allowed-tools: Read, Grep, Glob

# Git operations only
allowed-tools: Read, Bash(git:*)

# Full access
allowed-tools: Read, Write, Edit, Bash
```

### Skill File Structure

```markdown
---
[YAML frontmatter]
---

# Skill Name

## What It Does

Brief description of capabilities.

## When to Use

Specific trigger scenarios.

## Quick Start

Minimal example to get started.

## Common Patterns

Pattern 1, Pattern 2, etc.

## References

Links to supporting files.
```

### Supporting Files

**Recommended Structure**:
```
skill-name/
├── SKILL.md           # Main skill instructions (≤500 lines)
├── reference.md       # Detailed API reference
├── examples.md        # Real-world examples
├── scripts/           # Utility scripts
│   └── helper.sh
└── templates/         # Reusable templates
    └── config.yaml
```

**File Linking**:
```markdown
See [reference.md](reference.md) for details.
See [examples.md](examples.md) for scenarios.
Run [scripts/deploy.sh](scripts/deploy.sh) to deploy.
```

---

## Agent Configuration

### Agent File (Markdown)

**Location**: `.claude/agents/agent-name.md` or `~/.claude/agents/agent-name.md`

**Structure**:
```markdown
# Agent Name

You are a [role] agent specialized in [domain].

## Responsibilities

- Task 1
- Task 2

## Guidelines

### When to Act

- Condition 1
- Condition 2

### How to Respond

1. Step 1
2. Step 2

## Example Interaction

**User**: [question]

**Agent**: [response]

## Quality Standards

- Standard 1
- Standard 2
```

### Agent Config JSON (Optional)

**Location**: `.claude/agents/agent-name-config.json`

```json
{
  "name": "Agent Display Name",
  "description": "What this agent does",
  "model": "claude-sonnet-4.5",
  "temperature": 0.7,
  "custom_instructions": "Additional context"
}
```

**Fields**:
- `name`: Display name
- `description`: Purpose (for selection menu)
- `model`: `claude-haiku-4`, `claude-sonnet-4.5`, `claude-opus-4`
- `temperature`: 0.0 (deterministic) to 1.0 (creative)
- `custom_instructions`: Additional context loaded with agent

### Model Selection Guide

| Model | Best For | Speed | Cost |
|-------|----------|-------|------|
| **Haiku 4** | Quick tasks, formatting, docs | Fastest | Lowest |
| **Sonnet 4.5** | Planning, implementation, reasoning | Medium | Medium |
| **Opus 4** | Complex analysis, creative work | Slowest | Highest |

**MoAI-ADK Defaults**:
- Alfred SuperAgent: Sonnet 4.5
- doc-syncer, tag-agent, git-manager: Haiku 4
- spec-builder, code-builder: Sonnet 4.5
- debug-helper: Sonnet 4.5

### Invoking Agents

**Syntax**:
```bash
claude
> "@agent-name Do something"
```

**Examples**:
```bash
> "@db-migrator Add email column to users table"
> "@test-analyzer Review test coverage"
> "@feature-builder Add password reset"
```

---

## Hook Types and Events

### Hook Events

| Event | Timing | Can Block | Use Case |
|-------|--------|-----------|----------|
| **SessionStart** | Session begins | No | Load project context |
| **SessionEnd** | Session ends | No | Cleanup tasks |
| **PreToolUse** | Before tool execution | Yes | Validation, security |
| **PostToolUse** | After tool execution | No | Formatting, linting |
| **UserPromptSubmit** | Before prompt processing | Yes | Input validation |
| **Notification** | On notification | No | Logging |
| **Stop** | After Claude responds | No | Cleanup |
| **SubagentStop** | After subagent task | No | Logging |
| **PreCompact** | Before compaction | No | State preservation |

### Hook Configuration Format

**Location**: `~/.claude/settings.json`

```json
{
  "hooks": {
    "EventName": [
      {
        "matcher": "ToolName or *",
        "hooks": [
          {
            "type": "command",
            "command": "bash /path/to/script.sh"
          }
        ]
      }
    ]
  }
}
```

### PreToolUse Hook

**Purpose**: Run checks before tool execution, optionally block

**Example**: Block destructive git operations
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/git-safety-check.sh \"{{tool_input.command}}\""
          }
        ]
      }
    ]
  }
}
```

**git-safety-check.sh**:
```bash
#!/bin/bash

COMMAND="$1"

# Block dangerous git commands
if [[ "$COMMAND" =~ git[[:space:]]+(push[[:space:]]+--force|reset[[:space:]]+--hard|clean[[:space:]]+-fd) ]]; then
    echo "❌ Dangerous git command blocked: $COMMAND"
    echo "Use --no-verify to override if you're sure."
    exit 1
fi

exit 0  # Allow
```

### PostToolUse Hook

**Purpose**: Run actions after successful tool execution

**Example**: Auto-format after file write
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/format-file.sh \"{{tool_input.file_path}}\""
          }
        ]
      }
    ]
  }
}
```

### SessionStart Hook

**Purpose**: Initialize session context

**Example**: Load project information
```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/load-context.sh"
          }
        ]
      }
    ]
  }
}
```

### Hook Variables

Available template variables in hook commands:

| Variable | Description | Example |
|----------|-------------|---------|
| `{{tool_input.command}}` | Bash command | `git commit` |
| `{{tool_input.file_path}}` | File path | `/src/app.py` |
| `{{tool_input.old_string}}` | Edit old text | `foo` |
| `{{tool_input.new_string}}` | Edit new text | `bar` |

---

## Plugin Structure

### plugin.json Specification

```json
{
  "name": "Plugin Name",
  "version": "1.0.0",
  "description": "What this plugin provides",
  "author": "Author Name",
  "repository": "https://github.com/user/repo",
  "commands": [
    {
      "name": "command-name",
      "file": "commands/command-name.md",
      "description": "What this command does"
    }
  ],
  "agents": [
    {
      "name": "agent-name",
      "file": "agents/agent-name.md",
      "description": "What this agent does"
    }
  ],
  "skills": [
    {
      "name": "skill-name",
      "directory": "skills/skill-name"
    }
  ],
  "mcpServers": {
    "server-name": {
      "command": "node",
      "args": ["servers/server.js"],
      "env": {
        "VAR": "value"
      }
    }
  }
}
```

### Plugin Directory Structure

```
my-plugin/
├── plugin.json                 # Plugin manifest
├── README.md                   # Installation & usage
├── LICENSE                     # License file
├── commands/                   # Slash commands
│   ├── command1.md
│   └── command2.md
├── agents/                     # Sub-agents
│   ├── agent1.md
│   └── agent1-config.json
├── skills/                     # Skills
│   ├── skill1/
│   │   └── SKILL.md
│   └── skill2/
│       └── SKILL.md
└── servers/                    # MCP servers (optional)
    └── custom-server.js
```

### Installing Plugins

**From GitHub**:
```bash
claude /plugin install https://github.com/user/plugin-name
```

**From Local Directory**:
```bash
claude /plugin install /path/to/plugin
```

**Listing Plugins**:
```bash
claude /plugin list
```

**Removing Plugins**:
```bash
claude /plugin remove plugin-name
```

---

## MCP Server Integration

### What are MCP Servers?

**MCP (Model Context Protocol) Servers** extend Claude Code with custom tools, resources, and prompts via a standardized protocol.

**Use Cases**:
- Custom database queries
- API integrations
- Project-specific build tools
- External service access

### MCP Server Configuration

**Location**: `~/.claude/settings.json`

```json
{
  "mcpServers": {
    "server-name": {
      "command": "node",
      "args": ["path/to/server.js"],
      "env": {
        "API_KEY": "key_value",
        "PROJECT_PATH": "/path/to/project"
      }
    }
  }
}
```

### Creating an MCP Server (Node.js)

```javascript
#!/usr/bin/env node

const { MCPServer } = require('@anthropic/mcp-sdk');

const server = new MCPServer({
  name: 'Custom Tools',
  version: '1.0.0'
});

// Register a tool
server.registerTool({
  name: 'custom_action',
  description: 'Performs custom action',
  parameters: {
    type: 'object',
    properties: {
      param1: {
        type: 'string',
        description: 'First parameter'
      }
    },
    required: ['param1']
  },
  handler: async (params) => {
    // Tool implementation
    return {
      result: `Processed: ${params.param1}`
    };
  }
});

// Register a resource
server.registerResource({
  uri: 'project://config',
  name: 'Project Configuration',
  description: 'Access project config',
  mimeType: 'application/json',
  handler: async () => {
    return {
      data: { /* config data */ }
    };
  }
});

server.start();
```

### MCP Tool Example

**Database Query Tool**:
```javascript
server.registerTool({
  name: 'query_database',
  description: 'Query project database',
  parameters: {
    type: 'object',
    properties: {
      sql: {
        type: 'string',
        description: 'SQL query to execute'
      }
    }
  },
  handler: async ({ sql }) => {
    const db = require('better-sqlite3')(process.env.DB_PATH);
    try {
      const results = db.prepare(sql).all();
      return { results };
    } catch (error) {
      return { error: error.message };
    }
  }
});
```

---

## Settings Configuration

### Location

- **User Settings**: `~/.claude/settings.json`
- **Project Settings**: `.claude/settings.json`

### Complete Settings Example

```json
{
  "model": "claude-sonnet-4.5",
  "temperature": 0.7,
  "outputStyle": "concise",
  "hooks": {
    "SessionStart": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/session-start.sh"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/pre-commit.sh \"{{tool_input.command}}\""
          }
        ]
      }
    ]
  },
  "mcpServers": {
    "project-tools": {
      "command": "node",
      "args": ["mcp-servers/tools.js"],
      "env": {
        "PROJECT_ROOT": "/path/to/project"
      }
    }
  },
  "customInstructions": "Additional global context for Claude"
}
```

### Settings Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model` | string | `claude-sonnet-4.5` | Default model |
| `temperature` | number | 0.7 | Response randomness (0-1) |
| `outputStyle` | string | `normal` | `concise`, `normal`, `detailed` |
| `hooks` | object | `{}` | Hook configurations |
| `mcpServers` | object | `{}` | MCP server configurations |
| `customInstructions` | string | - | Additional context |

---

## Progressive Disclosure Model

Claude Code Skills use a three-level loading model to minimize context overhead:

### Level 1: Metadata (Always Active)

**Loaded**: YAML frontmatter only
**Size**: ~100 tokens per skill
**Purpose**: Discovery and activation

```yaml
---
name: PDF Processing
description: Extract text and tables from PDF files...
allowed-tools: Read, Bash
---
```

Claude sees this for all installed skills at session start.

### Level 2: Instructions (On Demand)

**Loaded**: Full SKILL.md content
**Size**: ~2000-5000 tokens
**Trigger**: When Claude determines skill is relevant

**Example**:
```
User: "Extract text from report.pdf"
→ Claude recognizes "PDF" in description
→ Loads full PDF Processing skill
→ Uses instructions to complete task
```

### Level 3: Resources (As Needed)

**Loaded**: Supporting files (reference.md, examples.md, scripts)
**Size**: Variable
**Trigger**: When explicitly referenced or accessed

**Example**:
```
SKILL.md references: "See reference.md for API details"
→ Claude reads reference.md when needed
→ Or user asks: "Show me PDF API reference"
```

### Benefits

**Without Progressive Disclosure**:
- 55 skills × 5000 tokens = 275,000 tokens at startup
- Context window consumed before user interaction

**With Progressive Disclosure**:
- 55 skills × 100 tokens = 5,500 tokens at startup
- 270,000 tokens available for conversation
- Skills load only when relevant

---

## Tool Permissions

### Read-Only Skills

For analysis, documentation, or search tasks:

```yaml
allowed-tools: Read, Grep, Glob
```

**Example**: Code analyzer, documentation generator

### Safe Modification Skills

For code changes without shell access:

```yaml
allowed-tools: Read, Write, Edit
```

**Example**: Refactoring tool, formatter

### Git-Only Skills

For version control operations:

```yaml
allowed-tools: Read, Bash(git:*)
```

**Example**: Branch manager, commit helper

### Full Access Skills

For deployment, builds, or system operations:

```yaml
allowed-tools: Read, Write, Edit, Bash
```

**Example**: Deployment tool, build manager

### Tool Permission Patterns

| Pattern | Tools | Security Level | Use Case |
|---------|-------|----------------|----------|
| **Read-Only** | Read, Grep, Glob | High | Analysis, search |
| **Edit-Only** | Read, Edit | Medium-High | Refactoring |
| **Git-Safe** | Read, Bash(git:*) | Medium | Version control |
| **Command-Specific** | Read, Bash(npm:*) | Medium | Package management |
| **Full Access** | Read, Write, Edit, Bash | Low | Deployment |

---

## Troubleshooting Guide

### Skill Not Activating

**Symptoms**: Claude doesn't use your skill when expected

**Diagnosis**:
1. Check YAML syntax: `---` opening/closing required
2. Verify description includes trigger keywords
3. Confirm file path: `~/.claude/skills/skill-name/SKILL.md`
4. Check indentation (YAML is space-sensitive, no tabs)

**Fix**:
```yaml
# ❌ Invalid (missing closing ---)
---
name: My Skill
description: Does things

# ✅ Valid
---
name: My Skill
description: Does things
---
```

**Test**:
```bash
# Debug mode
claude --debug
> "Use my skill to do X"
# Watch for skill loading messages
```

### Hook Not Executing

**Symptoms**: Hook script doesn't run

**Diagnosis**:
1. Check hook is executable: `chmod +x script.sh`
2. Verify JSON syntax in settings.json
3. Check matcher matches tool name exactly
4. Review hook script exit codes

**Fix**:
```bash
# Make executable
chmod +x ~/.claude/hooks/my-hook.sh

# Test hook directly
bash ~/.claude/hooks/my-hook.sh "test input"
```

### Agent Not Available

**Symptoms**: `@agent-name` not recognized

**Diagnosis**:
1. Check file location: `.claude/agents/agent-name.md`
2. Verify file name matches invocation (case-sensitive)
3. Confirm agent file has markdown format

**Fix**:
```bash
# List available agents
ls .claude/agents/

# Check agent file
cat .claude/agents/agent-name.md
```

### MCP Server Not Starting

**Symptoms**: MCP server tools not available

**Diagnosis**:
1. Check server command is valid
2. Verify dependencies installed
3. Review server logs
4. Confirm environment variables set

**Fix**:
```bash
# Test server directly
node mcp-servers/server.js

# Check logs
claude --debug
# Look for MCP server initialization messages
```

### Context Overflow

**Symptoms**: "Context window full" errors

**Solutions**:
1. **Manual Compaction**: `/compact`
2. **Start Fresh**: `/clear`
3. **Reduce Skill Size**: Split SKILL.md into reference.md
4. **Optimize Prompts**: Use shorter, focused prompts

### Performance Issues

**Symptoms**: Slow responses, high memory

**Diagnosis**:
1. Too many skills loaded
2. Large hook scripts
3. Complex MCP servers
4. Deep skill dependencies

**Solutions**:
- Remove unused skills
- Optimize hook scripts
- Use caching in MCP servers
- Simplify skill instructions

---

## Quick Reference Cards

### Skill Creation Checklist

- [ ] YAML frontmatter complete (name, description)
- [ ] Description includes 3+ trigger keywords
- [ ] File ≤500 lines (or split to reference.md)
- [ ] Relative paths used (not absolute)
- [ ] Examples provided
- [ ] Tested locally

### Hook Creation Checklist

- [ ] Script is executable (`chmod +x`)
- [ ] Exit code 0 = success, non-zero = block/error
- [ ] Error messages clear
- [ ] Template variables used correctly
- [ ] Tested with sample inputs

### Plugin Creation Checklist

- [ ] plugin.json valid JSON
- [ ] README with installation instructions
- [ ] All referenced files exist
- [ ] Version number follows semver
- [ ] License file included
- [ ] Tested installation locally

---

## Version Information

**Claude Code Version**: 0.6.0+
**Documented Features**:
- ✅ Skills (Progressive Disclosure)
- ✅ Sub-agents
- ✅ Hooks (9 event types)
- ✅ Plugins
- ✅ MCP Servers
- ✅ Slash Commands

**Breaking Changes from v0.5.x**:
- Hooks configuration moved to settings.json
- Progressive Disclosure introduced for skills
- Plugin format updated

---

## Official Documentation Links

- [Claude Code CLI](https://docs.claude.com/en/docs/claude-code)
- [Skills Guide](https://docs.claude.com/en/docs/claude-code/skills)
- [Hooks Guide](https://docs.claude.com/en/docs/claude-code/hooks-guide)
- [Plugins Announcement](https://www.anthropic.com/news/claude-code-plugins)

---

**Version**: 2.0.0
**Last Updated**: 2025-10-22
**Part of**: MoAI-ADK Claude Code Operations
**Companion**: See examples.md for practical scenarios
