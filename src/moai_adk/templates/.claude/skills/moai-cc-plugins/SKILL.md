---
name: moai-cc-plugins
version: 1.0.0
created: 2025-10-31
updated: 2025-10-31
status: active
description: Claude Code plugin manifest (plugin.json) schema, validation, and best practices based on official documentation
keywords: ['claude-code', 'plugin', 'plugin.json', 'schema', 'validation', 'manifest']
allowed-tools:
  - Read
  - Write
  - Bash
---

# Claude Code Plugins Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-cc-plugins |
| **Version** | 1.0.0 (2025-10-31) |
| **Allowed tools** | Read, Write, Bash |
| **Auto-load** | On demand when plugin development detected |
| **Tier** | Claude Code Configuration |

---

## What It Does

Official Claude Code plugin manifest (plugin.json) schema guide with validation rules, best practices, and executable examples based on Claude Code v2.0 specification.

**Key capabilities**:
- ✅ Official plugin.json schema (Claude Code v2.0)
- ✅ Required vs optional field specifications
- ✅ Directory structure conventions
- ✅ Validation checklist and common pitfalls
- ✅ 5+ working examples with detailed annotations
- ✅ Integration with commands, agents, skills, hooks, and MCP servers

---

## When to Use

**Automatic triggers**:
- Creating new Claude Code plugins
- Debugging plugin manifest errors
- Validating plugin.json structure
- Setting up plugin marketplace

**Manual invocation**:
- Review plugin configuration
- Migrate to official schema
- Design plugin architecture
- Troubleshoot plugin registration issues

---

## Official Plugin.json Schema

### Required Fields

Every plugin.json MUST include these fields:

```json
{
  "name": "my-plugin",           // kebab-case, no spaces
  "description": "Brief description of plugin functionality",
  "version": "1.0.0",             // Semantic versioning (major.minor.patch)
  "author": {                     // Object format (NOT string)
    "name": "Your Name"
  }
}
```

### Optional Fields

```json
{
  // Skills (array of relative paths)
  "skills": [
    "./skills/my-skill/SKILL.md",
    "./skills/another-skill.md"
  ],

  // Commands (array of command metadata - NOT paths)
  "commands": [
    {
      "name": "my-command",
      "description": "Command description"
    }
  ],

  // Agents (array of agent metadata - NOT paths)
  "agents": [
    {
      "name": "my-agent",
      "description": "Agent description"
    }
  ],

  // MCP Servers integration
  "mcpServers": [
    {
      "name": "server-name",
      "type": "optional",          // or "required"
      "configPath": ".mcp.json"
    }
  ],

  // Additional metadata
  "category": "backend",
  "tags": ["fastapi", "python", "backend"],
  "repository": "https://github.com/user/repo",
  "documentation": "https://docs.example.com",
  "permissions": {
    "filesystem": ["read", "write"],
    "network": ["http", "https"]
  },
  "dependencies": ["other-plugin-name"]
}
```

---

## Critical Schema Rules

### ❌ COMMON MISTAKES

1. **Commands/Agents as Arrays with Paths** (WRONG)
   ```json
   // ❌ WRONG - Do NOT include "path" field
   "commands": [
     {
       "name": "init-fastapi",
       "path": "commands/init-fastapi.md",  // ❌ Invalid
       "description": "..."
     }
   ]
   ```

   **✅ CORRECT** - Commands metadata only (actual files in `commands/` directory):
   ```json
   "commands": [
     {
       "name": "init-fastapi",
       "description": "Initialize FastAPI project"
     }
   ]
   ```

2. **Author as String** (WRONG)
   ```json
   // ❌ WRONG
   "author": "John Doe"

   // ✅ CORRECT
   "author": {
     "name": "John Doe"
   }
   ```

3. **Skills with Absolute Paths** (WRONG)
   ```json
   // ❌ WRONG
   "skills": ["moai-framework-fastapi"]

   // ✅ CORRECT
   "skills": ["./skills/moai-framework-fastapi.md"]
   ```

4. **Plugin Name with Spaces** (WRONG)
   ```json
   // ❌ WRONG
   "name": "My Cool Plugin"

   // ✅ CORRECT
   "name": "my-cool-plugin"
   ```

### ✅ KEY PRINCIPLES

1. **Separation of Concerns**:
   - `plugin.json` = Metadata registry
   - `commands/` directory = Actual command markdown files
   - `agents/` directory = Actual agent markdown files
   - `skills/` directory = Actual skill packages

2. **Relative Paths Only**:
   - All paths in plugin.json MUST start with `./`
   - Example: `./skills/my-skill/SKILL.md`

3. **Kebab-case Naming**:
   - Plugin names: `my-plugin-name`
   - Command names: `init-server`
   - Agent names: `api-designer`

4. **Semantic Versioning**:
   - Format: `major.minor.patch`
   - Example: `1.0.0`, `2.3.1`, `0.1.0-beta`

---

## Directory Structure

### Standard Plugin Layout

```
my-plugin/
├── .claude-plugin/
│   └── plugin.json              # Plugin manifest (metadata only)
├── commands/
│   ├── init-server.md           # Command implementation
│   └── deploy.md
├── agents/
│   ├── backend-agent.md         # Agent implementation
│   └── api-designer.md
├── skills/
│   ├── my-skill/
│   │   ├── SKILL.md             # Skill definition
│   │   ├── reference.md         # Detailed documentation
│   │   └── examples.md          # Usage examples
│   └── another-skill.md         # Single-file skill
├── .mcp.json                    # MCP server configuration (optional)
└── README.md                    # Plugin documentation
```

### File Naming Conventions

| Component | Filename Format | Example |
|-----------|----------------|---------|
| **Plugin manifest** | `.claude-plugin/plugin.json` | Fixed location |
| **Commands** | `commands/{kebab-case}.md` | `commands/init-fastapi.md` |
| **Agents** | `agents/{kebab-case}.md` | `agents/backend-agent.md` |
| **Skills (package)** | `skills/{name}/SKILL.md` | `skills/my-skill/SKILL.md` |
| **Skills (single)** | `skills/{name}.md` | `skills/helper-skill.md` |
| **MCP config** | `.mcp.json` | Root level |

---

## Validation Checklist

Use this checklist before registering your plugin:

### Metadata Validation
- [ ] `name` is kebab-case (no spaces, lowercase)
- [ ] `description` is concise (1-2 sentences)
- [ ] `version` follows semantic versioning
- [ ] `author` is an object with `name` field

### Structure Validation
- [ ] All `skills` paths start with `./`
- [ ] All skill files exist at specified paths
- [ ] Command/agent objects have `name` and `description` only
- [ ] No `path` fields in commands/agents arrays
- [ ] All referenced MCP config files exist

### Directory Validation
- [ ] Commands in `commands/` directory match plugin.json
- [ ] Agents in `agents/` directory match plugin.json
- [ ] Skills in `skills/` directory match plugin.json
- [ ] `.claude-plugin/plugin.json` exists

### Content Validation
- [ ] No hardcoded secrets or credentials
- [ ] All URLs are valid and accessible
- [ ] Dependencies are available
- [ ] Permissions are minimal (least privilege)

### JSON Syntax
- [ ] Valid JSON (no trailing commas, proper quotes)
- [ ] All required fields present
- [ ] No duplicate keys
- [ ] Proper escaping of special characters

---

## Common Patterns

### Pattern 1: Minimal Plugin (Metadata Only)

**Use case**: Simple plugin with just configuration, no custom components.

```json
{
  "name": "my-config-plugin",
  "description": "Configuration plugin for project settings",
  "version": "1.0.0",
  "author": {
    "name": "Your Name"
  }
}
```

### Pattern 2: Skills-Only Plugin

**Use case**: Provide reusable knowledge without commands/agents.

```json
{
  "name": "knowledge-base",
  "description": "Curated knowledge for domain-specific tasks",
  "version": "1.0.0",
  "author": {
    "name": "Your Name"
  },
  "skills": [
    "./skills/domain-patterns/SKILL.md",
    "./skills/best-practices.md"
  ]
}
```

### Pattern 3: Command-Driven Plugin

**Use case**: Workflow automation with slash commands.

```json
{
  "name": "workflow-plugin",
  "description": "Automate common development workflows",
  "version": "1.0.0",
  "author": {
    "name": "Your Name"
  },
  "commands": [
    {
      "name": "init-project",
      "description": "Initialize new project structure"
    },
    {
      "name": "run-tests",
      "description": "Execute full test suite"
    }
  ]
}
```

### Pattern 4: Agent-Powered Plugin

**Use case**: Specialized AI agents for specific domains.

```json
{
  "name": "backend-plugin",
  "description": "Backend development specialists",
  "version": "1.0.0",
  "author": {
    "name": "Your Name"
  },
  "agents": [
    {
      "name": "api-designer",
      "description": "REST API design specialist"
    },
    {
      "name": "db-optimizer",
      "description": "Database query optimization expert"
    }
  ]
}
```

### Pattern 5: Full-Featured Plugin

**Use case**: Complete development environment with all components.

```json
{
  "name": "fullstack-plugin",
  "description": "Full-stack development toolkit",
  "version": "2.0.0",
  "author": {
    "name": "Your Name"
  },
  "commands": [
    {
      "name": "init-stack",
      "description": "Initialize full-stack project"
    }
  ],
  "agents": [
    {
      "name": "frontend-agent",
      "description": "Frontend specialist"
    },
    {
      "name": "backend-agent",
      "description": "Backend specialist"
    }
  ],
  "skills": [
    "./skills/frontend-patterns/SKILL.md",
    "./skills/backend-patterns/SKILL.md"
  ],
  "category": "fullstack",
  "tags": ["react", "fastapi", "typescript", "python"]
}
```

### Pattern 6: MCP Integration Plugin

**Use case**: External tool integration via Model Context Protocol.

```json
{
  "name": "playwright-plugin",
  "description": "E2E testing automation with Playwright MCP",
  "version": "1.0.0",
  "author": {
    "name": "Your Name"
  },
  "commands": [
    {
      "name": "playwright-setup",
      "description": "Initialize Playwright-MCP for E2E testing"
    }
  ],
  "skills": [
    "./skills/moai-testing-playwright-mcp.md"
  ],
  "mcpServers": [
    {
      "name": "playwright-mcp",
      "type": "optional",
      "configPath": ".mcp.json"
    }
  ]
}
```

---

## Best Practices

### 1. Minimal Permissions Principle
Request only the permissions your plugin actually needs:

```json
// ❌ AVOID
"permissions": {
  "filesystem": ["read", "write", "delete"],
  "network": ["*"]
}

// ✅ PREFER
"permissions": {
  "filesystem": ["read"],
  "network": ["https"]
}
```

### 2. Progressive Disclosure
Start simple, add complexity as needed:

1. **v1.0.0**: Minimal plugin (metadata + 1-2 core features)
2. **v1.1.0**: Add skills (knowledge capsules)
3. **v1.2.0**: Add commands (workflow automation)
4. **v2.0.0**: Add agents (specialized reasoning)

### 3. Clear Documentation
Every plugin should have:

- Concise `description` in plugin.json (1-2 sentences)
- Detailed README.md with usage examples
- Inline comments in command/agent markdown files
- Skill documentation (SKILL.md, reference.md, examples.md)

### 4. Version Management
Follow semantic versioning strictly:

- **Patch (x.y.Z)**: Bug fixes, documentation updates
- **Minor (x.Y.0)**: New features, backward compatible
- **Major (X.0.0)**: Breaking changes, schema updates

### 5. Dependency Management
Keep dependencies minimal and explicit:

```json
"dependencies": ["base-plugin"]  // Only if truly required
```

Avoid circular dependencies:
- Plugin A depends on Plugin B
- Plugin B depends on Plugin A  // ❌ NOT ALLOWED

### 6. MCP Server Integration
Use MCP servers for:
- External tool integration (Playwright, Figma, etc.)
- Database connections
- API integrations
- File system operations beyond basic read/write

Mark as `"type": "optional"` when possible:
```json
"mcpServers": [
  {
    "name": "playwright-mcp",
    "type": "optional",  // Users can choose not to install
    "configPath": ".mcp.json"
  }
]
```

---

## Troubleshooting Guide

### Error: "name: Required"
**Cause**: Missing top-level `name` field
**Fix**: Add `"name": "my-plugin-name"` to plugin.json

### Error: "author: Expected object, received string"
**Cause**: Author specified as string instead of object
**Fix**: Change `"author": "John"` to `"author": {"name": "John"}`

### Error: "Plugin name cannot contain spaces"
**Cause**: Plugin name not in kebab-case
**Fix**: Change `"My Plugin"` to `"my-plugin"`

### Error: "skills.0: Invalid input: must start with './'"
**Cause**: Skill path is absolute or missing `./` prefix
**Fix**: Change `"my-skill"` to `"./skills/my-skill.md"`

### Error: "commands.0.path: Unrecognized key"
**Cause**: Command object includes `path` field
**Fix**: Remove `path` field - commands are discovered from `commands/` directory

### Error: "Invalid JSON syntax"
**Cause**: Trailing commas, missing quotes, etc.
**Fix**: Validate JSON with linter (e.g., `jq . plugin.json`)

### Plugin not loading
**Checklist**:
1. Verify `.claude-plugin/plugin.json` exists
2. Check all referenced files exist
3. Validate JSON syntax
4. Ensure plugin name is unique
5. Check Claude Code logs for detailed errors

---

## Integration with Marketplace

### Marketplace.json Structure

```json
{
  "name": "my-marketplace",
  "owner": {
    "name": "marketplace-owner"
  },
  "plugins": [
    {
      "name": "my-plugin",
      "source": "./plugins/my-plugin",
      "description": "Brief plugin description"
    }
  ]
}
```

### Plugin Registration Flow

```
1. Create plugin directory structure
2. Write plugin.json with metadata
3. Implement commands/agents/skills
4. Validate with checklist
5. Add to marketplace.json
6. Run: /plugin marketplace add /path/to/marketplace
7. Install: /plugin install my-plugin@my-marketplace
```

---

## Inputs

- Existing plugin directories
- Plugin configuration requirements
- Component specifications (commands, agents, skills)
- Integration requirements (MCP servers, dependencies)

## Outputs

- Valid plugin.json manifests
- Validated directory structures
- Integration configurations
- Troubleshooting reports

## Failure Modes

- Invalid JSON syntax
- Missing required fields
- Incorrect path formats
- Circular dependencies
- Permission conflicts

## Dependencies

- Access to plugin directories via Read/Write tools
- JSON validation capabilities
- File system operations via Bash

---

## References (Latest Documentation)

- [Claude Code Plugin Documentation](https://docs.claude.com/en/docs/claude-code/plugins)
- [Plugin Marketplace Schema Report](../../.moai/reports/plugin-marketplace-schema-fix-report.md)
- [Claude Code Official Templates](https://github.com/anthropics/claude-code-templates)

_Documentation links verified 2025-10-31_

---

## Changelog

- **v1.0.0** (2025-10-31): Initial release with Claude Code v2.0 official schema, validation rules, 6 patterns, troubleshooting guide

---

## Works Well With

- `moai-alfred-cc-manager` (Claude Code plugin management)
- `moai-foundation-trust` (quality validation)
- `moai-essentials-debug` (debugging support)

---

**Quick Reference Card**:
```
REQUIRED FIELDS:
  name, description, version, author{name}

OPTIONAL FIELDS:
  skills, commands, agents, mcpServers,
  category, tags, repository, documentation,
  permissions, dependencies

KEY RULES:
  ✅ kebab-case names
  ✅ Relative paths (./...)
  ✅ Semantic versioning
  ✅ Author as object
  ❌ No spaces in names
  ❌ No path field in commands/agents
  ❌ No absolute skill paths
```
