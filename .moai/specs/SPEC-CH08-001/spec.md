---
spec_id: SPEC-CH08-001
title: Claude Code Plugins & Migration - Output Styles → Plugins/Hooks/Skills Architecture
version: 1.0.0-dev
status: In Development
owner: GOOS🪿엉아 / MoAI-ADK Team
tags: ["plugin", "claude-code", "migration", "hooks", "skills", "v1.0"]
created: 2025-10-30
modified: 2025-10-30
language: en
---

# SPEC-CH08-001: Claude Code Plugins & Migration

## 📋 Overview

This SPEC defines the foundational knowledge for **Claude Code Plugins** architecture, required for v1.0 of MoAI-ADK. It covers the migration from deprecated Output Styles to modern Plugins/Hooks/Skills patterns.

### Strategic Goals

1. **Architecture Understanding** — Comprehensive guide to Claude Code plugin system
2. **Migration Blueprint** — Clear path from Output Styles (deprecated) to Plugins (v1.0+)
3. **Reference Architecture** — Production-grade plugin patterns for v1.0 plugins
4. **Developer Enablement** — Enable seamless plugin development across teams

### Scope

- ✅ Plugin.json schema (command, agents, hooks, MCP)
- ✅ Command development patterns
- ✅ Agent orchestration within plugins
- ✅ Hook lifecycle & event patterns
- ✅ MCP server integration
- ✅ Skill development for plugins
- ✅ Settings & configuration management
- ✅ Plugin permissions model (allowed/denied tools)
- ❌ Marketplace implementation (covered in separate SPECs)

---

## 🎯 EARS Requirements

### Ubiquitous Behaviors (Core Features)

**1. Plugin Discovery**
- GIVEN a developer initializes a new Claude Code project
- WHEN the developer explores available plugins
- THEN system displays plugin registry (manifest.json) with:
  - Plugin name, version, description
  - Required Claude Code version
  - Permissions (allowed-tools, denied-tools)
  - Installation command

**2. Plugin Installation**
- GIVEN a developer wants to install a plugin
- WHEN `/plugin install moai-alfred-pm` is executed
- THEN system downloads plugin from registry, validates structure, registers in settings.json

**3. Command Execution**
- GIVEN a user invokes a plugin command
- WHEN `/init-pm [args]` is executed
- THEN system routes to plugin agent, executes command, returns output

**4. Hook System**
- GIVEN a plugin wants to react to events (session start, tool execution, etc.)
- WHEN SessionStart hook fires
- THEN registered hook callbacks execute in order

### Event-Driven Behaviors

**Plugin Lifecycle Events**:
- WHEN `/plugin install` completes → SessionStart hook fires
- WHEN tool execution begins → PreToolUse hook fires
- WHEN tool execution completes → PostToolUse hook fires
- WHEN session terminates → SessionEnd hook fires

**Skill Loading**:
- WHEN plugin command is invoked
- THEN associated skills are loaded from `.claude/skills/` directory
- AND skill content is passed to agent for contextual guidance

### State-Driven Behaviors

**Plugin State Machine**:
- State: `installed` → `configured` → `ready` → `active`
- WHEN plugin transitions from `installed` to `configured`
- THEN system validates plugin.json schema, loads commands, agents, hooks
- AND registers permissions in settings.json

**Tool Access Control**:
- GIVEN a plugin specifies `allowed-tools` and `denied-tools`
- WHEN plugin agent attempts to use a tool
- THEN system enforces permission boundaries (deny-by-default)

### Optional Behaviors

**Multiple Plugins Coordination**:
- GIVEN two plugins want to share context
- WHEN both are active
- THEN settings.json provides shared context via `.moai/memory/` files
- AND agents can reference each other via `Skill()` invocations (optional)

---

## 📐 Plugin Architecture

### Directory Structure

```
my-plugin/
├── .claude-plugin/
│   ├── plugin.json             ← Metadata & configuration
│   └── hooks.json              ← Hook lifecycle definitions
├── commands/
│   ├── cmd-1.md                ← Command templates
│   └── cmd-2.md
├── agents/
│   ├── agent-1.md              ← Sub-agent definitions
│   └── agent-2.md
├── skills/
│   ├── SKILL-FEATURE-001.md    ← Reusable knowledge
│   └── SKILL-FEATURE-002.md
├── README.md                    ← Installation & overview
├── USAGE.md                     ← Usage guide & examples
├── CHANGELOG.md                 ← Version history
└── tests/
    └── test_plugin.py           ← Integration tests
```

### plugin.json Schema

```json
{
  "id": "moai-alfred-pm",
  "name": "MoAI PM Plugin",
  "version": "1.0.0",
  "description": "Project Management kickoff automation",
  "author": "GOOS🪿",
  "repository": "https://github.com/moai-adk/moai-alfred-marketplace/tree/main/plugins/moai-alfred-pm",
  "minClaudeCodeVersion": "1.0.0",
  "commands": [
    {
      "name": "init-pm",
      "path": "commands/init-pm.md",
      "description": "Initialize project management templates"
    }
  ],
  "agents": [
    {
      "name": "pm-agent",
      "path": "agents/pm-agent.md",
      "type": "specialist"
    }
  ],
  "hooks": {
    "sessionStart": "hooks.json#onSessionStart",
    "preToolUse": "hooks.json#onPreToolUse",
    "postToolUse": "hooks.json#onPostToolUse"
  },
  "mcpServers": [
    {
      "name": "github",
      "type": "optional",
      "configPath": ".mcp.json"
    }
  ],
  "permissions": {
    "allowedTools": ["Read", "Write", "Bash"],
    "deniedTools": ["DeleteFile", "KillProcess"]
  },
  "skills": [
    "moai-plugin-scaffolding",
    "moai-plugin-testing-patterns"
  ],
  "settings": {
    "apiKey": {
      "type": "secret",
      "description": "API key for external service (optional)"
    }
  }
}
```

### hooks.json Schema

```json
{
  "sessionStart": {
    "name": "onSessionStart",
    "description": "Fires when Claude Code session starts",
    "priority": 100,
    "timeout": 5000,
    "conditions": {
      "minClaudeCodeVersion": "1.0.0"
    }
  },
  "preToolUse": {
    "name": "onPreToolUse",
    "description": "Fires before tool execution",
    "priority": 50,
    "timeout": 1000
  },
  "postToolUse": {
    "name": "onPostToolUse",
    "description": "Fires after tool execution",
    "priority": 50,
    "timeout": 2000
  }
}
```

---

## 🔄 Migration Path: Output Styles → Plugins

### Before (v0.x, Deprecated)

Output Styles provided basic customization:
```
.claude/output-style.json
```

❌ **Limitations**:
- Static styling rules only
- No command extensions
- No agent coordination
- No hook system
- No skill integration

### After (v1.0+, Plugins)

Plugins provide full extensibility:
```
.claude/
├── plugins/
│   ├── moai-alfred-pm/
│   └── moai-alfred-uiux/
├── settings.json (plugin registry)
└── skills/ (plugin skills)
```

✅ **Capabilities**:
- Command extensions (`/init-pm`, `/setup-shadcn-ui`)
- Agent orchestration
- Hook-based reactions
- Skill-driven guidance
- Permission-based security
- Multi-plugin coordination

### Migration Checklist

For v0.x users upgrading to v1.0:

- [ ] Remove `.claude/output-style.json`
- [ ] Create plugin directories under `.claude/plugins/`
- [ ] Migrate styling to plugin skills (optional)
- [ ] Define plugin commands & agents
- [ ] Update settings.json with plugin registry
- [ ] Test plugin installation & command execution

---

## 🛠️ Command Development Pattern

### Template Structure

**Example: `commands/init-pm.md`**

```markdown
# /init-pm

Initialize MoAI project management templates.

## Syntax

```bash
/init-pm [project-name]
```

## Arguments

- **project-name** (required): Name of the project (e.g., `my-awesome-project`)

## Options

- `--skip-charter`: Skip project charter generation
- `--template`: Use alternative template (default: `moai-spec`)

## Examples

```bash
/init-pm my-awesome-project
/init-pm ecommerce-platform --template=enterprise
```

## What it does

1. Creates `.moai/specs/SPEC-{PROJECT}-001/` directory
2. Generates EARS SPEC template (spec.md, plan.md, acceptance.md)
3. Creates project charter (charter.md)
4. Builds risk matrix (risk-matrix.json)

## Output

```
.moai/specs/SPEC-MY-AWESOME-001/
├── spec.md (EARS requirement specification)
├── plan.md (implementation plan)
├── acceptance.md (acceptance criteria)
├── charter.md (project charter)
└── risk-matrix.json (risk assessment)
```

## Related Skills

- `moai-plugin-scaffolding` (for plugin generation patterns)
- `moai-foundation-ears` (EARS syntax reference)
```

### Agent Invocation

**Example: `agents/pm-agent.md`**

```markdown
# PM Agent

Specialist agent for project management automation.

## Responsibilities

1. Parse `/init-pm` command arguments
2. Invoke `moai-plugin-scaffolding` skill for SPEC template
3. Generate project charter from user inputs
4. Create risk matrix assessment

## Tools

- Read (template files)
- Write (SPEC documents)
- Edit (risk matrix)

## Interaction

User executes: `/init-pm my-project`
↓
PM Agent receives: command + args
↓
Agent invokes: `Skill("moai-plugin-scaffolding")`
↓
Agent generates: .moai/specs/SPEC-MY-PROJECT-001/
↓
User receives: Generated files + summary report
```

---

## 🎣 Hook System Patterns

### Hook Types & Timing

| Hook | When | Use Case | Timeout |
|------|------|----------|---------|
| **SessionStart** | Claude Code session begins | Initialize plugin state, load configs | 5s |
| **PreToolUse** | Before tool execution | Validate tool calls, enforce permissions | 1s |
| **PostToolUse** | After tool execution | Log execution, cache results | 2s |
| **SessionEnd** | Session terminates | Cleanup, save state | 3s |

### Example: Permission Enforcement Hook

```javascript
// hooks.json
{
  "preToolUse": {
    "name": "enforceDeniedTools",
    "priority": 100,
    "handler": "src/hooks/preToolUse.ts"
  }
}
```

```typescript
// src/hooks/preToolUse.ts
export async function enforceDeniedTools(context: HookContext): Promise<void> {
  const toolName = context.toolCall.name;
  const deniedTools = context.plugin.permissions.deniedTools;

  if (deniedTools.includes(toolName)) {
    throw new Error(`Tool '${toolName}' denied by plugin permissions`);
  }
}
```

---

## 📚 Skill Development for Plugins

### Skill Structure

```
.claude/skills/
├── moai-plugin-scaffolding/
│   ├── SKILL.md (main content, <500 words)
│   ├── examples.md (code examples)
│   └── reference.md (API reference)
└── moai-plugin-testing-patterns/
    ├── SKILL.md
    ├── examples.md
    └── reference.md
```

### SKILL.md Template

```markdown
# Plugin Scaffolding Skill

## Overview
Guide for rapid plugin generation from templates.

## Key Concepts
- Plugin directory structure conventions
- plugin.json schema & validation
- Command template patterns
- Agent architecture guidelines

## Patterns

### Pattern 1: Command-Only Plugin
For simple CLI commands without agents.

[Details...]

### Pattern 2: Agent-Based Plugin
For complex workflows requiring agents.

[Details...]

## Checklist

- [ ] plugin.json created
- [ ] Commands defined
- [ ] Agents (if needed) defined
- [ ] Hooks (if needed) registered
- [ ] Skills linked
- [ ] README.md written
- [ ] Tests written

## See Also
- `moai-foundation-specs` (SPEC authoring)
- `moai-alfred-commands` (Command patterns)
```

---

## 🔐 Permission Model

### Tool Access Control

```json
{
  "permissions": {
    "allowedTools": [
      "Read",           // File read
      "Write",          // File write
      "Edit",           // File editing
      "Bash",           // Command execution
      "Task"            // Sub-agent invocation
    ],
    "deniedTools": [
      "DeleteFile",     // Destructive
      "KillProcess",    // System control
      "Bash(rm -rf)"    // Specific bash patterns
    ]
  }
}
```

### Deny-by-Default Strategy

1. Plugin specifies ONLY `allowedTools`
2. System denies everything else
3. Hook validates at PreToolUse time
4. Agent respects boundary (cannot override)

---

## ✅ Acceptance Criteria

### ch08 Completion Requirements

**1. Documentation** (9 sections)
- [ ] 8-1: Plugin Architecture Overview
- [ ] 8-2: Plugin.json Schema Deep Dive
- [ ] 8-3: Command Development Patterns
- [ ] 8-4: Agent Orchestration
- [ ] 8-5: Hook Lifecycle & Events
- [ ] 8-6: Skill Integration for Plugins
- [ ] 8-7: Permission Model & Security
- [ ] 8-8: Migration Path (Output Styles → Plugins)
- [ ] 8-9: FAQ & Troubleshooting

**2. Hands-on Labs** (4 labs)
- [ ] Lab 8A: Create a simple command-only plugin
- [ ] Lab 8B: Build a plugin with agents
- [ ] Lab 8C: Register hooks and test lifecycle
- [ ] Lab 8D: Implement permission enforcement

**3. Examples & Screenshots**
- [ ] 5+ code examples (JSON, JavaScript, Markdown)
- [ ] 4+ architecture diagrams
- [ ] 3+ plugin execution screenshots
- [ ] CLI command reference chart

**4. FAQ Section**
- [ ] 10+ Q&A entries covering common questions

**5. Quality**
- [ ] 0 broken links
- [ ] 100% code examples executable
- [ ] Consistent tone and terminology
- [ ] All references to moai-alfred (not moai-cc)

---

## 📌 Notes

- This SPEC provides **foundational knowledge only**
- v1.0 plugins (PM, UI/UX, Frontend, Backend, DevOps) will implement this architecture
- Each v1.0 plugin will have its own SPEC (SPEC-CH09-PM-001, etc.)
- Marketplace integration covered in separate SPEC

---

## 🔗 Related SPECs

- **SPEC-CH09-PM-001**: PM Plugin Implementation
- **SPEC-CH09-UIUX-001**: UI/UX Plugin Implementation
- **SPEC-CH09-FE-001**: Frontend Plugin Implementation
- **SPEC-CH09-BE-001**: Backend Plugin Implementation
- **SPEC-CH09-DEVOPS-001**: DevOps Plugin Implementation
- **SPEC-CH10-001**: Full Blog Platform Project

---

**Status**: In Development
**Next Steps**: Begin writing ch08 sections (Day 2-3)
