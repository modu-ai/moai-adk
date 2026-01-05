# Plugin Validation Guide

## Validation Commands

### CLI Validation

Basic Validation:
```bash
claude plugin validate ./my-plugin
```

Verbose Validation:
```bash
claude plugin validate ./my-plugin --verbose
```

Strict Mode:
```bash
claude plugin validate ./my-plugin --strict
```

### Validation Output

Success Output:
```
Plugin validation passed: my-plugin
  - plugin.json: valid
  - commands: 3 found
  - agents: 2 found
  - skills: 1 found
  - hooks: valid
```

Error Output:
```
Plugin validation failed: my-plugin
  ERROR: plugin.json missing required field 'name'
  ERROR: commands/invalid.md missing frontmatter
  WARNING: agent tools reference unknown tool 'CustomTool'
```

## Validation Rules by Component

### plugin.json Validation

Required Checks:
- File exists at .claude-plugin/plugin.json
- Valid JSON syntax
- Contains "name" field
- Name is kebab-case format

Name Validation:
- Lowercase letters, numbers, hyphens only
- Must start with letter
- No consecutive hyphens
- Maximum 64 characters

Optional Field Checks:
- version: Valid semver (X.Y.Z)
- author.email: Valid email format
- author.url: Valid URL
- homepage: Valid URL
- repository: Valid URL
- license: Valid SPDX identifier
- keywords: Array of strings

Path Reference Checks:
- All paths start with "./"
- Referenced paths exist
- No path traversal (..)
- Paths are relative to plugin root

### Command Validation

File Requirements:
- Extension is .md
- File is readable
- File is not empty

Frontmatter Validation:
- YAML frontmatter present (between --- delimiters)
- Contains description field
- Description is non-empty string

Content Validation:
- Body content exists after frontmatter
- No syntax errors in markdown
- Argument references are valid ($1, $2, $ARGUMENTS)

### Agent Validation

File Requirements:
- Extension is .md
- File is readable

Frontmatter Validation:
- Contains name field
- Name is valid identifier
- Description is present (recommended)

Tool Validation:
- All listed tools are valid tool names
- Valid tools: Read, Write, Edit, Grep, Glob, Bash, WebFetch, WebSearch, Task, TodoWrite

Model Validation:
- If specified, must be: sonnet, opus, haiku, inherit
- Default is inherit

Permission Mode Validation:
- If specified, must be: default, bypassPermissions, plan, passthrough
- Default is default

Skill References:
- Referenced skills exist in plugin or project
- Skill names are valid identifiers

### Skill Validation

Directory Requirements:
- Skill directory exists
- SKILL.md file present in directory

Frontmatter Validation:
- Contains name field
- Contains description field
- allowed-tools contains valid tool names

Content Validation:
- File is under 500 lines
- Contains Quick Reference section
- Contains Implementation Guide section

### Hook Validation

File Requirements:
- hooks.json is valid JSON
- hooks.json contains "hooks" object

Event Validation:
- Event names are valid hook events
- Valid events: PreToolUse, PostToolUse, PermissionRequest, UserPromptSubmit, SessionStart, SessionEnd, Stop, SubagentStop, Notification, PreCompact

Hook Entry Validation:
- Each entry has valid structure
- Matcher is string or array
- hooks array is present

Hook Definition Validation:
- type is "command" or "prompt"
- If command: command field is string
- If prompt: prompt field is string

Command Path Validation:
- Script paths are valid
- Referenced scripts exist
- Scripts are executable

### MCP Server Validation

File Requirements:
- .mcp.json is valid JSON
- Contains mcpServers object

Server Entry Validation:
- Each server has unique name
- Contains command field
- Command is string

Args Validation:
- If present, is array
- All elements are strings

Env Validation:
- If present, is object
- All values are strings
- Environment variables properly formatted

Transport Validation:
- If type specified, is: stdio, http, sse
- If http/sse: url field required
- URL is valid format

### LSP Server Validation

File Requirements:
- .lsp.json is valid JSON
- Contains lspServers object

Server Entry Validation:
- Each server has unique name
- Contains command field
- Contains extensionToLanguage mapping

Extension Mapping Validation:
- Keys start with "."
- Values are non-empty strings
- At least one mapping present

## Common Validation Errors

### plugin.json Errors

Error: Invalid plugin name
```
ERROR: Plugin name 'My_Plugin' is invalid
  - Must be kebab-case
  - Use lowercase letters, numbers, hyphens
  Fix: Rename to 'my-plugin'
```

Error: Missing name field
```
ERROR: plugin.json missing required field 'name'
  Fix: Add "name": "plugin-name" to plugin.json
```

Error: Invalid version format
```
ERROR: Version '1.0' is not valid semver
  Fix: Use format X.Y.Z (e.g., "1.0.0")
```

Error: Path does not exist
```
ERROR: Path './commands' does not exist
  Fix: Create commands directory or remove from plugin.json
```

### Command Errors

Error: Missing frontmatter
```
ERROR: commands/test.md missing YAML frontmatter
  Fix: Add frontmatter with description:
  ---
  description: Command description
  ---
```

Error: Missing description
```
ERROR: commands/test.md frontmatter missing 'description'
  Fix: Add description field to frontmatter
```

### Agent Errors

Error: Invalid tool reference
```
ERROR: agents/helper.md references invalid tool 'Execute'
  Valid tools: Read, Write, Edit, Grep, Glob, Bash, WebFetch, WebSearch, Task, TodoWrite
  Fix: Use valid tool name
```

Error: Invalid model
```
ERROR: agents/helper.md has invalid model 'gpt4'
  Valid models: sonnet, opus, haiku, inherit
  Fix: Use valid model name
```

### Hook Errors

Error: Invalid event name
```
ERROR: hooks.json contains invalid event 'BeforeWrite'
  Valid events: PreToolUse, PostToolUse, PermissionRequest, UserPromptSubmit, SessionStart, SessionEnd, Stop, SubagentStop, Notification, PreCompact
  Fix: Use valid event name
```

Error: Invalid hook type
```
ERROR: Hook type 'script' is invalid
  Valid types: command, prompt
  Fix: Use 'command' or 'prompt'
```

Error: Script not found
```
ERROR: Hook script './hooks/validate.sh' not found
  Fix: Create script or update path
```

### MCP Server Errors

Error: Missing command
```
ERROR: MCP server 'my-server' missing 'command' field
  Fix: Add command field with executable path
```

Error: Invalid args type
```
ERROR: MCP server 'my-server' args must be array
  Fix: Change args to array format: ["arg1", "arg2"]
```

### LSP Server Errors

Error: Missing extensionToLanguage
```
ERROR: LSP server 'python' missing 'extensionToLanguage'
  Fix: Add extension mapping: {".py": "python"}
```

Error: Invalid extension format
```
ERROR: Extension 'py' should start with '.'
  Fix: Change to '.py'
```

## Validation Best Practices

Pre-Submission Checklist:
1. Run `claude plugin validate` before distribution
2. Test plugin locally with `--plugin-dir`
3. Verify all paths resolve correctly
4. Test all commands work as expected
5. Confirm hooks trigger on events
6. Validate MCP/LSP servers connect

Automated Validation:

Add to CI/CD pipeline:
```yaml
# GitHub Actions example
- name: Validate Plugin
  run: claude plugin validate ./my-plugin --strict
```

Version Control Hooks:

Add pre-commit validation:
```bash
#!/bin/bash
# .git/hooks/pre-commit
claude plugin validate . --strict
if [ $? -ne 0 ]; then
  echo "Plugin validation failed"
  exit 1
fi
```

## Validation Levels

Level 1 - Syntax:
- JSON/YAML validity
- Required fields present
- Basic type checking

Level 2 - Semantic:
- Path references exist
- Tool names valid
- Event names valid
- Model names valid

Level 3 - Functional:
- Scripts executable
- Commands invocable
- Servers connectable
- Hooks triggerable

Running Different Levels:

Syntax only:
```bash
claude plugin validate ./my-plugin --level syntax
```

Full validation:
```bash
claude plugin validate ./my-plugin --level full
```
