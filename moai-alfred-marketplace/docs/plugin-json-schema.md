# plugin.json Schema Reference

**Complete guide to plugin.json manifest structure for Alfred Framework plugins**

## Overview

Every plugin requires a `plugin.json` file in `.claude-plugin/` directory. This manifest defines:
- Plugin metadata (name, version, author)
- Available commands
- Agent specifications
- Hook definitions
- Permission model
- MCP server configurations
- Skill dependencies
- Settings schema

## Schema Structure

### Root Object

```json
{
  "id": "string (required, unique identifier)",
  "name": "string (required, display name)",
  "version": "string (required, semver format)",
  "status": "string (optional, development|beta|stable)",
  "description": "string (required, brief description)",
  "author": "string (required, author name/email)",
  "category": "string (required, pm|frontend|backend|uiux|devops)",
  "tags": ["string[]"],
  "repository": "string (optional, GitHub URL)",
  "license": "string (optional, default MIT)",
  "minClaudeCodeVersion": "string (required, minimum version)",
  "commands": [{...}],
  "agents": [{...}],
  "hooks": {object},
  "mcpServers": [{...}],
  "permissions": {object},
  "skills": ["string[]"],
  "settings": {object},
  "dependencies": ["string[]"],
  "installCommand": "string",
  "releaseNotes": "string"
}
```

## Field Definitions

### Metadata Fields

#### `id` (required)
- Type: `string`
- Pattern: `moai-alfred-{plugin-name}`
- Must be globally unique
- Use lowercase with hyphens (no underscores)

```json
{
  "id": "moai-alfred-myPlugin"
}
```

#### `name` (required)
- Type: `string`
- Display name for plugin
- Human-readable (can contain spaces, uppercase)

```json
{
  "name": "My Awesome Plugin"
}
```

#### `version` (required)
- Type: `string`
- Semantic versioning: `MAJOR.MINOR.PATCH[-PRERELEASE]`
- Examples: `1.0.0`, `1.0.0-dev`, `1.0.0-rc1`, `0.1.0-beta.1`

```json
{
  "version": "1.0.0-dev"
}
```

#### `status` (optional)
- Type: `string`
- Enum: `development`, `beta`, `stable`
- Default: `stable`

```json
{
  "status": "development"
}
```

#### `description` (required)
- Type: `string`
- Brief plugin description (1-2 sentences)
- Explains what plugin does

```json
{
  "description": "FastAPI scaffolding with SQLAlchemy and Alembic support"
}
```

#### `author` (required)
- Type: `string`
- Author name or email format
- Format: `Name` or `Name <email@example.com>`

```json
{
  "author": "GOOS🪿",
  "author": "John Doe <john@example.com>"
}
```

#### `category` (required)
- Type: `string`
- Enum: `pm`, `frontend`, `backend`, `uiux`, `devops`
- Categorizes plugin type

```json
{
  "category": "backend"
}
```

#### `tags` (optional)
- Type: `string[]`
- Array of search tags
- Max 5-8 tags

```json
{
  "tags": ["fastapi", "python", "sqlalchemy", "database", "async"]
}
```

#### `repository` (optional)
- Type: `string`
- GitHub repository URL

```json
{
  "repository": "https://github.com/moai-adk/moai-alfred-myPlugin"
}
```

#### `license` (optional)
- Type: `string`
- Default: `MIT`
- SPDX identifier

```json
{
  "license": "MIT"
}
```

#### `minClaudeCodeVersion` (required)
- Type: `string`
- Minimum Claude Code version required
- Semver format: `1.0.0`, `1.1.0+`

```json
{
  "minClaudeCodeVersion": "1.0.0"
}
```

### Commands

#### Commands Array

```json
{
  "commands": [
    {
      "name": "command-name",
      "path": "commands/command-name.md",
      "description": "What this command does"
    }
  ]
}
```

#### Command Object Fields

- **`name`** (required): Command identifier (lowercase, hyphens)
  - User invokes as: `/command-name`
- **`path`** (required): Relative path to command template (`.md` file)
- **`description`** (required): One-line description of command

#### Example: Multiple Commands

```json
{
  "commands": [
    {
      "name": "init-fastapi",
      "path": "commands/init-fastapi.md",
      "description": "Initialize FastAPI project with uv"
    },
    {
      "name": "db-setup",
      "path": "commands/db-setup.md",
      "description": "Setup database with Alembic migrations"
    },
    {
      "name": "resource-crud",
      "path": "commands/resource-crud.md",
      "description": "Generate CRUD endpoints from SPEC"
    }
  ]
}
```

### Agents

#### Agents Array

```json
{
  "agents": [
    {
      "name": "agent-name",
      "path": "agents/agent-name.md",
      "type": "specialist",
      "description": "Agent purpose"
    }
  ]
}
```

#### Agent Object Fields

- **`name`** (required): Agent identifier (lowercase, hyphens)
- **`path`** (required): Relative path to agent template (`.md` file)
- **`type`** (required): Agent type (specialist, coordinator)
  - `specialist`: Handles single responsibility (most common)
  - `coordinator`: Orchestrates multiple agents
- **`description`** (required): What agent does

#### Example

```json
{
  "agents": [
    {
      "name": "fastapi-agent",
      "path": "agents/fastapi-agent.md",
      "type": "specialist",
      "description": "Scaffolds FastAPI projects with SQLAlchemy models"
    },
    {
      "name": "migration-agent",
      "path": "agents/migration-agent.md",
      "type": "specialist",
      "description": "Manages Alembic migrations"
    }
  ]
}
```

### Hooks

#### Hooks Object

```json
{
  "hooks": {
    "sessionStart": ".claude-plugin/hooks.json#onSessionStart",
    "preToolUse": ".claude-plugin/hooks.json#onPreToolUse",
    "postToolUse": ".claude-plugin/hooks.json#onPostToolUse",
    "sessionEnd": ".claude-plugin/hooks.json#onSessionEnd"
  }
}
```

#### Hook References

Hooks point to handlers defined in `hooks.json`:

- **`sessionStart`** (optional): Fires when session begins
- **`preToolUse`** (optional): Fires before tool execution
- **`postToolUse`** (optional): Fires after tool execution
- **`sessionEnd`** (optional): Fires when session ends

#### Example

```json
{
  "hooks": {
    "sessionStart": ".claude-plugin/hooks.json#onSessionStart",
    "preToolUse": ".claude-plugin/hooks.json#onPreToolUse"
  }
}
```

See [hooks.json Schema](./hooks-json-schema.md) for details.

### MCP Servers

#### MCP Array

```json
{
  "mcpServers": [
    {
      "name": "server-name",
      "type": "required|optional",
      "configPath": ".mcp.json"
    }
  ]
}
```

#### MCP Object Fields

- **`name`** (required): MCP server name (e.g., `vercel`, `supabase`)
- **`type`** (required): Enum `required` or `optional`
  - `required`: Plugin cannot function without this MCP server
  - `optional`: Plugin can function, but MCP enhances functionality
- **`configPath`** (required): Path to MCP configuration file

#### Example

```json
{
  "mcpServers": [
    {
      "name": "vercel",
      "type": "optional",
      "configPath": ".mcp.json"
    },
    {
      "name": "supabase",
      "type": "optional",
      "configPath": ".mcp.json"
    }
  ]
}
```

### Permissions

#### Permissions Object

```json
{
  "permissions": {
    "allowedTools": ["Read", "Write", "Edit", "Bash"],
    "deniedTools": ["DeleteFile", "KillProcess"]
  }
}
```

#### Permission Fields

- **`allowedTools`** (required): Array of explicitly allowed tools
  - Only these tools can be invoked by plugin
  - Default: empty array (deny-by-default)
- **`deniedTools`** (optional): Array of explicitly denied tools
  - Takes precedence over allowedTools
  - For blacklist exceptions

#### Available Tools

**Read-Only**:
- `Read` - File reading
- `Glob` - File pattern matching
- `Grep` - Content search

**Modification**:
- `Write` - File creation/overwrite
- `Edit` - File modification

**System**:
- `Bash` - Shell command execution
- `Bash(npm:*)` - npm commands only
- `Bash(python3:*)` - python3 execution
- `Bash(git:*)` - git operations

**Advanced**:
- `Task` - Sub-agent invocation
- `Skill` - Skill invocation
- `NotebookEdit` - Jupyter notebook editing

**Deny-Only**:
- `DeleteFile` - File deletion
- `KillProcess` - Process termination

#### Permission Examples

**Example 1: Read-Only Plugin**
```json
{
  "permissions": {
    "allowedTools": ["Read", "Glob", "Grep"],
    "deniedTools": []
  }
}
```

**Example 2: Frontend with Specific Bash**
```json
{
  "permissions": {
    "allowedTools": ["Read", "Write", "Edit", "Bash(npm:*)", "Bash(git:*)"],
    "deniedTools": []
  }
}
```

**Example 3: Backend with Python**
```json
{
  "permissions": {
    "allowedTools": ["Read", "Write", "Edit", "Bash(python3:*)", "Bash(pip:*)", "Task"],
    "deniedTools": ["DeleteFile"]
  }
}
```

### Skills

#### Skills Array

```json
{
  "skills": [
    "moai-foundation-ears",
    "moai-spec-authoring",
    "moai-plugin-scaffolding"
  ]
}
```

#### Skill References

- Type: `string[]`
- References existing Claude Skills
- Skills automatically loaded when plugin is invoked
- Max 5-10 skills recommended

#### Example

```json
{
  "skills": [
    "moai-lang-fastapi-patterns",
    "moai-lang-python",
    "moai-domain-backend",
    "moai-domain-database",
    "moai-plugin-scaffolding"
  ]
}
```

### Settings

#### Settings Object

```json
{
  "settings": {
    "apiKey": {
      "type": "secret",
      "description": "API key for external service"
    },
    "databaseUrl": {
      "type": "string",
      "description": "Database connection URL",
      "default": "postgresql://localhost/mydb"
    },
    "enableAnalytics": {
      "type": "boolean",
      "description": "Enable usage analytics",
      "default": false
    }
  }
}
```

#### Setting Types

- **`secret`**: Sensitive data (passwords, API keys)
  - Stored encrypted in `.claude/settings.json`
  - Never shown in plaintext
- **`string`**: Text value
- **`boolean`**: True/false flag
- **`number`**: Numeric value
- **`array`**: Array of values

#### Setting Fields

- **`type`** (required): Data type
- **`description`** (required): Human-readable description
- **`default`** (optional): Default value

#### Example

```json
{
  "settings": {
    "vercelToken": {
      "type": "secret",
      "description": "Vercel API token for deployment"
    },
    "supabaseUrl": {
      "type": "string",
      "description": "Supabase project URL"
    },
    "supabaseKey": {
      "type": "secret",
      "description": "Supabase anonymous key"
    }
  }
}
```

### Dependencies

#### Dependencies Array

```json
{
  "dependencies": ["moai-alfred-pm", "moai-alfred-uiux"]
}
```

#### Dependency Fields

- Type: `string[]`
- References other plugin IDs that must be installed
- Automatically installed as prerequisites

#### Example

```json
{
  "dependencies": [
    "moai-alfred-uiux",
    "moai-alfred-pm"
  ]
}
```

### Install Command

#### Install Command

```json
{
  "installCommand": "/plugin install moai-alfred-myPlugin"
}
```

- Type: `string`
- Standard format: `/plugin install {plugin-id}`
- Shown in documentation

### Release Notes

#### Release Notes

```json
{
  "releaseNotes": "Initial v1.0.0-dev release with 3 commands"
}
```

- Type: `string`
- Brief summary of latest release
- Appears in marketplace

## Complete Example: Backend Plugin

```json
{
  "id": "moai-alfred-backend",
  "name": "Backend Plugin",
  "version": "1.0.0-dev",
  "status": "development",
  "description": "FastAPI 0.120.2 + uv scaffolding - SQLAlchemy 2.0, Alembic migrations",
  "author": "GOOS🪿",
  "category": "backend",
  "tags": ["fastapi", "python", "sqlalchemy", "database", "async"],
  "repository": "https://github.com/moai-adk/moai-alfred-marketplace/tree/main/plugins/moai-alfred-backend",
  "license": "MIT",
  "minClaudeCodeVersion": "1.0.0",

  "commands": [
    {
      "name": "init-fastapi",
      "path": "commands/init-fastapi.md",
      "description": "Initialize FastAPI project with uv"
    },
    {
      "name": "db-setup",
      "path": "commands/db-setup.md",
      "description": "Setup database with Alembic migrations"
    },
    {
      "name": "resource-crud",
      "path": "commands/resource-crud.md",
      "description": "Generate CRUD endpoints from SPEC"
    }
  ],

  "agents": [
    {
      "name": "backend-agent",
      "path": "agents/backend-agent.md",
      "type": "specialist",
      "description": "FastAPI scaffolding and database management"
    }
  ],

  "hooks": {
    "sessionStart": ".claude-plugin/hooks.json#onSessionStart",
    "preToolUse": ".claude-plugin/hooks.json#onPreToolUse"
  },

  "mcpServers": [],

  "permissions": {
    "allowedTools": [
      "Read",
      "Write",
      "Edit",
      "Bash(python3:*)",
      "Bash(pip:*)",
      "Task"
    ],
    "deniedTools": []
  },

  "skills": [
    "moai-lang-fastapi-patterns",
    "moai-lang-python",
    "moai-domain-backend",
    "moai-domain-database",
    "moai-plugin-scaffolding"
  ],

  "settings": {
    "databaseUrl": {
      "type": "string",
      "description": "Database connection URL",
      "default": "postgresql://localhost/mydb"
    },
    "pythonVersion": {
      "type": "string",
      "description": "Python version to use",
      "default": "3.11"
    }
  },

  "dependencies": [],

  "installCommand": "/plugin install moai-alfred-backend",
  "releaseNotes": "Initial v1.0.0-dev with FastAPI 0.120.2 and full async support"
}
```

## Validation Rules

### Required Fields

- ✅ `id` - Must be unique globally
- ✅ `name` - Human-readable name
- ✅ `version` - Semantic versioning
- ✅ `description` - Plugin purpose
- ✅ `author` - Author name/email
- ✅ `category` - One of 5 categories
- ✅ `minClaudeCodeVersion` - Minimum version support
- ✅ `commands` - At least one command
- ✅ `permissions` - Explicit permission definition

### Optional Fields

- ⚪ `status` - development|beta|stable
- ⚪ `tags` - Search keywords
- ⚪ `repository` - GitHub URL
- ⚪ `license` - License type
- ⚪ `agents` - Agent definitions
- ⚪ `hooks` - Hook bindings
- ⚪ `mcpServers` - MCP integrations
- ⚪ `skills` - Skill dependencies
- ⚪ `settings` - Plugin configuration
- ⚪ `dependencies` - Plugin dependencies

### Field Constraints

| Field | Type | Min/Max | Constraints |
|-------|------|---------|-------------|
| `id` | string | 3-50 chars | `moai-alfred-*` pattern |
| `name` | string | 3-100 chars | Human-readable |
| `version` | string | semver | MAJOR.MINOR.PATCH format |
| `description` | string | 10-500 chars | 1-2 sentences |
| `tags` | array | 0-8 items | Lowercase, hyphens |
| `commands` | array | 1+ items | At least one command |
| `agents` | array | 0+ items | Optional |
| `skills` | array | 0-10 items | Recommended max 5-10 |

## Migration from v0.x

If upgrading from older plugins:

1. **Update version**: Use semver (not arbitrary strings)
2. **Add category**: Pick one of 5 new categories
3. **Rename id**: Use `moai-alfred-*` pattern (not `moai-cc-*`)
4. **Define permissions**: Explicitly declare allowedTools

## See Also

- [hooks.json Schema](./hooks-json-schema.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [Security Policy](../SECURITY.md)
- [SPEC-CH08-001](../../.moai/specs/SPEC-CH08-001/spec.md)

---

**Version**: 1.0.0
**Last Updated**: 2025-10-30

🔗 Generated with [Claude Code](https://claude.com/claude-code)
