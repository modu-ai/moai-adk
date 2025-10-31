# Claude Code Plugin Schema Reference

## Complete Field Specifications

### Required Fields

#### name
- **Type**: `string`
- **Format**: kebab-case (lowercase, hyphen-separated)
- **Pattern**: `^[a-z][a-z0-9-]*[a-z0-9]$`
- **Examples**:
  - ✅ `"my-plugin"`
  - ✅ `"backend-plugin"`
  - ✅ `"fastapi-toolkit"`
  - ❌ `"My Plugin"` (contains spaces)
  - ❌ `"my_plugin"` (uses underscores)
  - ❌ `"MyPlugin"` (PascalCase)

**Validation Rules**:
- Must start with lowercase letter
- Only lowercase letters, numbers, and hyphens
- Cannot start or end with hyphen
- No consecutive hyphens
- Minimum 2 characters

#### description
- **Type**: `string`
- **Length**: 10-200 characters recommended
- **Format**: Plain text, single sentence preferred
- **Purpose**: Appears in plugin listings and search results

**Best Practices**:
- Start with action verb or clear functionality statement
- Include key technologies/frameworks
- Mention primary use case
- Keep under 150 characters for best display

**Examples**:
- ✅ `"FastAPI scaffolding with SQLAlchemy 2.0 and Alembic migrations"`
- ✅ `"E2E testing automation with Playwright MCP integration"`
- ✅ `"UI/UX design automation with Figma MCP and shadcn/ui"`
- ❌ `"Plugin"` (too vague)
- ❌ `"This is a really amazing plugin that does lots of things..."` (too long, unclear)

#### version
- **Type**: `string`
- **Format**: Semantic versioning (SemVer 2.0)
- **Pattern**: `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`

**Semantic Versioning Rules**:
- **MAJOR.MINOR.PATCH** (e.g., `1.2.3`)
- **MAJOR**: Breaking changes (incompatible API changes)
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

**Pre-release identifiers** (optional):
- `-alpha`, `-beta`, `-rc.1`, etc.
- Examples: `1.0.0-alpha`, `2.1.0-beta.2`, `3.0.0-rc.1`

**Build metadata** (optional):
- `+build.123`, `+sha.5114f85`, etc.
- Examples: `1.0.0+20130313144700`, `1.0.0-beta+exp.sha.5114f85`

**Examples**:
- ✅ `"1.0.0"` (stable release)
- ✅ `"0.1.0"` (initial development)
- ✅ `"2.3.1"` (patch release)
- ✅ `"1.0.0-beta"` (pre-release)
- ✅ `"1.0.0-rc.1"` (release candidate)
- ❌ `"1.0"` (missing patch version)
- ❌ `"v1.0.0"` (should not include 'v' prefix)
- ❌ `"latest"` (not a version number)

#### author
- **Type**: `object`
- **Required fields**: `name` (string)
- **Optional fields**: `email` (string), `url` (string)

**Schema**:
```json
{
  "name": "string (required)",
  "email": "string (optional)",
  "url": "string (optional)"
}
```

**Examples**:
```json
// Minimal (required only)
"author": {
  "name": "John Doe"
}

// With email
"author": {
  "name": "John Doe",
  "email": "john@example.com"
}

// Full metadata
"author": {
  "name": "John Doe",
  "email": "john@example.com",
  "url": "https://johndoe.dev"
}
```

**Common Mistakes**:
```json
// ❌ String format (INVALID)
"author": "John Doe"

// ❌ Array format (INVALID)
"author": ["John Doe"]

// ✅ Correct object format
"author": {
  "name": "John Doe"
}
```

---

### Optional Fields

#### skills
- **Type**: `array of strings`
- **Format**: Relative paths starting with `./`
- **Pattern**: `^\.\/.*\.md$`
- **File types**: Markdown files (`.md`)

**Path Formats**:
```json
// ✅ Single-file skill
"skills": [
  "./skills/my-skill.md"
]

// ✅ Package-style skill (SKILL.md)
"skills": [
  "./skills/my-skill/SKILL.md"
]

// ✅ Multiple skills
"skills": [
  "./skills/skill-1/SKILL.md",
  "./skills/skill-2.md",
  "./skills/skill-3/SKILL.md"
]

// ❌ Absolute path (INVALID)
"skills": [
  "my-skill"
]

// ❌ Missing ./ prefix (INVALID)
"skills": [
  "skills/my-skill.md"
]

// ❌ Non-markdown file (INVALID)
"skills": [
  "./skills/my-skill.txt"
]
```

**Validation Rules**:
- Must start with `./`
- Must end with `.md`
- File must exist at specified path
- Path is relative to plugin root

#### commands
- **Type**: `array of objects`
- **Object schema**: `{name: string, description: string}`
- **Purpose**: Metadata registry (NOT implementation paths)

**Schema**:
```json
{
  "name": "string (kebab-case, required)",
  "description": "string (required)"
}
```

**CRITICAL**: Commands array contains metadata only. Actual command implementations live in `commands/` directory as markdown files.

**Examples**:
```json
// ✅ Correct format
"commands": [
  {
    "name": "init-server",
    "description": "Initialize server with FastAPI"
  },
  {
    "name": "run-tests",
    "description": "Execute test suite with pytest"
  }
]

// ❌ WRONG - Do NOT include path field
"commands": [
  {
    "name": "init-server",
    "path": "commands/init-server.md",  // ❌ NOT ALLOWED
    "description": "..."
  }
]

// ❌ WRONG - String format
"commands": [
  "init-server",
  "run-tests"
]
```

**Discovery Mechanism**:
Claude Code automatically discovers command implementations by:
1. Reading `commands` array from plugin.json
2. Looking for `commands/{name}.md` files
3. Matching by command name

**Directory Structure**:
```
my-plugin/
├── .claude-plugin/
│   └── plugin.json          # Contains: {"commands": [{"name": "init-server", ...}]}
└── commands/
    └── init-server.md       # Actual implementation
```

#### agents
- **Type**: `array of objects`
- **Object schema**: `{name: string, description: string}`
- **Purpose**: Metadata registry (NOT implementation paths)

**Schema**:
```json
{
  "name": "string (kebab-case, required)",
  "description": "string (required)"
}
```

**Same principles as commands**: Agent array is metadata only. Implementations in `agents/` directory.

**Examples**:
```json
// ✅ Correct format
"agents": [
  {
    "name": "backend-agent",
    "description": "FastAPI and SQLAlchemy specialist"
  },
  {
    "name": "api-designer",
    "description": "REST API design expert"
  }
]

// ❌ WRONG - Do NOT include path/type fields
"agents": [
  {
    "name": "backend-agent",
    "path": "agents/backend-agent.md",  // ❌ NOT ALLOWED
    "type": "specialist",                // ❌ NOT ALLOWED
    "description": "..."
  }
]
```

**Directory Structure**:
```
my-plugin/
├── .claude-plugin/
│   └── plugin.json          # Contains: {"agents": [{"name": "backend-agent", ...}]}
└── agents/
    └── backend-agent.md     # Actual implementation
```

#### mcpServers
- **Type**: `array of objects`
- **Object schema**: `{name: string, type: string, configPath: string}`
- **Purpose**: External tool integration via Model Context Protocol

**Schema**:
```json
{
  "name": "string (required)",
  "type": "required | optional (required)",
  "configPath": "string (required)"
}
```

**Fields**:
- **name**: Unique identifier for MCP server
- **type**: `"required"` (must be installed) or `"optional"` (user choice)
- **configPath**: Relative path to MCP config file (usually `.mcp.json`)

**Examples**:
```json
// Single optional MCP server
"mcpServers": [
  {
    "name": "playwright-mcp",
    "type": "optional",
    "configPath": ".mcp.json"
  }
]

// Multiple MCP servers
"mcpServers": [
  {
    "name": "playwright-mcp",
    "type": "optional",
    "configPath": ".mcp-playwright.json"
  },
  {
    "name": "figma-mcp",
    "type": "required",
    "configPath": ".mcp-figma.json"
  }
]
```

**Best Practices**:
- Prefer `"type": "optional"` when possible (better UX)
- Use separate config files for each MCP server
- Document MCP server setup in README.md

#### category
- **Type**: `string`
- **Purpose**: Plugin classification for discovery/filtering
- **Examples**: `"backend"`, `"frontend"`, `"devops"`, `"ui-ux"`, `"testing"`, `"database"`, `"deployment"`

**Usage**:
```json
"category": "backend"
```

**Suggested Categories**:
- `backend`: Server-side development
- `frontend`: Client-side development
- `fullstack`: Full-stack development
- `devops`: Infrastructure and deployment
- `testing`: Quality assurance and testing
- `ui-ux`: Design and user experience
- `database`: Data management
- `security`: Security and authentication
- `ai-ml`: AI/ML development
- `documentation`: Documentation tooling

#### tags
- **Type**: `array of strings`
- **Purpose**: Searchable keywords for plugin discovery
- **Format**: Lowercase, single words or hyphenated phrases

**Examples**:
```json
// Backend plugin
"tags": ["fastapi", "python", "sqlalchemy", "backend", "rest-api"]

// Frontend plugin
"tags": ["react", "nextjs", "typescript", "frontend", "shadcn-ui"]

// DevOps plugin
"tags": ["docker", "kubernetes", "terraform", "devops", "ci-cd"]
```

**Best Practices**:
- Use 3-10 tags
- Include technology names (frameworks, languages)
- Add domain keywords (backend, frontend, etc.)
- Use lowercase
- Avoid redundancy with category/name

#### repository
- **Type**: `string (URL)`
- **Format**: Valid HTTP/HTTPS URL
- **Purpose**: Link to source code repository

**Examples**:
```json
"repository": "https://github.com/user/my-plugin"
```

**Supported Platforms**:
- GitHub: `https://github.com/user/repo`
- GitLab: `https://gitlab.com/user/repo`
- Bitbucket: `https://bitbucket.org/user/repo`

#### documentation
- **Type**: `string (URL)`
- **Format**: Valid HTTP/HTTPS URL
- **Purpose**: Link to external documentation

**Examples**:
```json
"documentation": "https://my-plugin-docs.example.com"
```

**Common Patterns**:
- GitHub Pages: `https://user.github.io/my-plugin`
- ReadTheDocs: `https://my-plugin.readthedocs.io`
- Custom domain: `https://docs.my-plugin.dev`

#### permissions
- **Type**: `object`
- **Purpose**: Declare required system permissions
- **Fields**: `filesystem`, `network`, `process`, etc.

**Schema**:
```json
{
  "filesystem": ["read", "write", "delete"],
  "network": ["http", "https"],
  "process": ["spawn"]
}
```

**Permission Types**:

**filesystem**:
- `"read"`: Read file contents
- `"write"`: Create/modify files
- `"delete"`: Delete files/directories

**network**:
- `"http"`: HTTP requests
- `"https"`: HTTPS requests
- `"*"`: All network access (avoid if possible)

**process**:
- `"spawn"`: Execute external commands
- `"kill"`: Terminate processes

**Best Practices**:
```json
// ❌ AVOID - Too permissive
"permissions": {
  "filesystem": ["read", "write", "delete"],
  "network": ["*"],
  "process": ["spawn", "kill"]
}

// ✅ PREFER - Minimal permissions
"permissions": {
  "filesystem": ["read"],
  "network": ["https"]
}
```

#### dependencies
- **Type**: `array of strings`
- **Format**: Plugin names (kebab-case)
- **Purpose**: Declare plugin dependencies

**Examples**:
```json
// Single dependency
"dependencies": ["base-plugin"]

// Multiple dependencies
"dependencies": ["base-plugin", "ui-toolkit"]
```

**Validation Rules**:
- Dependency plugin must be installed
- No circular dependencies
- Dependency versions are not specified (always uses latest compatible)

**Circular Dependency Detection**:
```
❌ NOT ALLOWED:
  Plugin A → depends on → Plugin B
  Plugin B → depends on → Plugin A

✅ ALLOWED:
  Plugin A → depends on → Plugin B
  Plugin C → depends on → Plugin B
  (Plugin B has no dependencies)
```

---

## Schema Version History

### v2.0 (Current - 2025-10-31)
- **Breaking Changes**:
  - `commands` and `agents` arrays no longer accept `path` field
  - Discovery mechanism changed to directory-based
  - `author` must be object format (string no longer accepted)
  - Skills must use relative paths with `./` prefix

- **New Features**:
  - MCP server integration support
  - Enhanced metadata (category, tags)
  - Permissions system
  - Dependency management

- **Deprecations**:
  - Custom schema URLs (`$schema` field) no longer used
  - Legacy `id` field removed
  - Legacy `status` field removed

### v1.0 (Legacy)
- Initial plugin.json schema
- Commands/agents with `path` fields
- Author as string format
- Custom schema support

---

## Migration Guide: v1.0 → v2.0

### Step 1: Update Author Format
```json
// Before (v1.0)
"author": "John Doe"

// After (v2.0)
"author": {
  "name": "John Doe"
}
```

### Step 2: Remove Path Fields from Commands/Agents
```json
// Before (v1.0)
"commands": [
  {
    "name": "init-server",
    "path": "commands/init-server.md",
    "description": "..."
  }
]

// After (v2.0)
"commands": [
  {
    "name": "init-server",
    "description": "..."
  }
]
```

### Step 3: Convert Skills to Relative Paths
```json
// Before (v1.0)
"skills": ["my-skill", "another-skill"]

// After (v2.0)
"skills": [
  "./skills/my-skill/SKILL.md",
  "./skills/another-skill.md"
]
```

### Step 4: Remove Legacy Fields
```json
// Remove these fields (no longer supported):
"id": "moai-plugin-backend",        // ❌ Remove
"status": "development",             // ❌ Remove
"$schema": "https://..."             // ❌ Remove
```

### Step 5: Validate with Official Schema
```bash
# Use Claude Code validation
/plugin validate ./my-plugin

# Or manual JSON validation
jq . .claude-plugin/plugin.json
```

---

## Validation Tools

### JSON Syntax Validation
```bash
# Using jq
jq . .claude-plugin/plugin.json

# Using python
python -m json.tool .claude-plugin/plugin.json

# Using node
node -e "JSON.parse(require('fs').readFileSync('.claude-plugin/plugin.json'))"
```

### Schema Validation Script
```python
import json
import re
from pathlib import Path

def validate_plugin(plugin_path):
    """Validate plugin.json against Claude Code v2.0 schema"""
    manifest = json.loads(Path(plugin_path).read_text())
    
    errors = []
    
    # Required fields
    if 'name' not in manifest:
        errors.append("Missing required field: name")
    elif not re.match(r'^[a-z][a-z0-9-]*[a-z0-9]$', manifest['name']):
        errors.append("Invalid name format (must be kebab-case)")
    
    if 'description' not in manifest:
        errors.append("Missing required field: description")
    
    if 'version' not in manifest:
        errors.append("Missing required field: version")
    elif not re.match(r'^\d+\.\d+\.\d+', manifest['version']):
        errors.append("Invalid version format (must be semantic versioning)")
    
    if 'author' not in manifest:
        errors.append("Missing required field: author")
    elif not isinstance(manifest['author'], dict):
        errors.append("Author must be object (not string)")
    elif 'name' not in manifest['author']:
        errors.append("Author object missing 'name' field")
    
    # Skills validation
    if 'skills' in manifest:
        for skill in manifest['skills']:
            if not skill.startswith('./'):
                errors.append(f"Skill path must start with './': {skill}")
            if not skill.endswith('.md'):
                errors.append(f"Skill must be markdown file: {skill}")
    
    # Commands/Agents validation
    for component in ['commands', 'agents']:
        if component in manifest:
            for item in manifest[component]:
                if 'path' in item:
                    errors.append(f"{component} should NOT have 'path' field: {item['name']}")
                if 'name' not in item:
                    errors.append(f"{component} item missing 'name' field")
                if 'description' not in item:
                    errors.append(f"{component} item missing 'description' field")
    
    return errors

# Usage
errors = validate_plugin('.claude-plugin/plugin.json')
if errors:
    print("Validation errors:")
    for error in errors:
        print(f"  - {error}")
else:
    print("✅ Plugin manifest is valid")
```

---

## Official Documentation Links

- [Claude Code Plugin Guide](https://docs.claude.com/en/docs/claude-code/plugins)
- [Model Context Protocol (MCP) Specification](https://modelcontextprotocol.io)
- [Semantic Versioning 2.0](https://semver.org)
- [JSON Schema Validation](https://json-schema.org)

_Last updated: 2025-10-31_
