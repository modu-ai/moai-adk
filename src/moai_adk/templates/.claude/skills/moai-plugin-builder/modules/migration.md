# Plugin Migration Guide

## Converting Standalone Configurations to Plugin Format

### Overview

This guide covers migrating existing `.claude/` configurations to distributable plugin format.

### Before Migration

Existing Standalone Structure:
```
project/
  .claude/
    commands/
      my-command.md
    agents/
      my-agent.md
    skills/
      my-skill/
        SKILL.md
    settings.json
    settings.local.json
```

### After Migration

Plugin Structure:
```
my-plugin/
  .claude-plugin/
    plugin.json
  commands/
    my-command.md
  agents/
    my-agent.md
  skills/
    my-skill/
      SKILL.md
  hooks/
    hooks.json
  README.md
```

### Migration Steps

Step 1: Create Plugin Directory

Create new directory for the plugin at desired location:
```bash
mkdir -p my-plugin/.claude-plugin
mkdir -p my-plugin/commands
mkdir -p my-plugin/agents
mkdir -p my-plugin/skills
mkdir -p my-plugin/hooks
```

Step 2: Create plugin.json

Create the plugin manifest file:
```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "Migrated plugin from standalone configuration",
  "commands": ["./commands"],
  "agents": ["./agents"],
  "skills": ["./skills"],
  "hooks": ["./hooks/hooks.json"]
}
```

Step 3: Copy Components

Copy existing components to plugin:
```bash
# Copy commands
cp .claude/commands/*.md my-plugin/commands/

# Copy agents
cp .claude/agents/*.md my-plugin/agents/

# Copy skills
cp -r .claude/skills/* my-plugin/skills/
```

Step 4: Extract Hooks from settings.json

If settings.json contains hooks, extract to hooks.json:

Original settings.json hooks:
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "validate-write.sh"
          }
        ]
      }
    ]
  }
}
```

New hooks/hooks.json:
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "bash ${CLAUDE_PLUGIN_ROOT}/hooks/validate-write.sh"
          }
        ]
      }
    ]
  }
}
```

Step 5: Update Path References

Update all hardcoded paths to use environment variables:

Before:
```bash
command: "/path/to/project/.claude/scripts/validate.sh"
```

After:
```bash
command: "bash ${CLAUDE_PLUGIN_ROOT}/scripts/validate.sh"
```

Step 6: Copy Supporting Scripts

Copy any script files referenced by hooks or commands:
```bash
mkdir -p my-plugin/scripts
cp .claude/scripts/*.sh my-plugin/scripts/
chmod +x my-plugin/scripts/*.sh
```

Step 7: Create Documentation

Create README.md for the plugin:
```markdown
# My Plugin

## Description

Plugin description and purpose.

## Installation

claude plugin install ./my-plugin

## Commands

- /my-plugin:command-name - Command description

## Agents

- my-agent - Agent description

## Skills

- my-skill - Skill description

## Configuration

Required environment variables or configuration.
```

### Path Reference Updates

Command Files:

Before migration:
```markdown
Reference: @.claude/config.yaml
Script: /project/.claude/scripts/run.sh
```

After migration:
```markdown
Reference: @${CLAUDE_PLUGIN_ROOT}/config.yaml
Script: ${CLAUDE_PLUGIN_ROOT}/scripts/run.sh
```

Agent Files:

Before migration:
```yaml
---
skills:
  - /project/.claude/skills/my-skill
---
```

After migration:
```yaml
---
skills:
  - my-skill
---
```

Hook Commands:

Before migration:
```json
{
  "command": "/project/.claude/hooks/validate.sh"
}
```

After migration:
```json
{
  "command": "bash ${CLAUDE_PLUGIN_ROOT}/hooks/validate.sh"
}
```

### Preserving Existing Functionality

Environment Variables:

If original scripts used project-specific variables:
```bash
# Original
API_URL="$PROJECT_API_URL"

# Migrated - use plugin-safe defaults
API_URL="${PROJECT_API_URL:-https://api.example.com}"
```

File References:

If commands referenced project files:
```markdown
# Original
Load config: @project-config.yaml

# Migrated - use CLAUDE_PROJECT_DIR
Load config: @${CLAUDE_PROJECT_DIR}/project-config.yaml
```

### Testing Migration

Validate Plugin Structure:
```bash
claude plugin validate ./my-plugin
```

Test Plugin Loading:
```bash
claude --plugin-dir ./my-plugin
```

Test Commands:
```bash
# Inside claude session
/my-plugin:command-name
```

Test Agents:
```bash
# Inside claude session
/agents
# Then invoke agent by name
```

### Common Migration Issues

Issue: Commands Not Found

Symptom: Plugin loads but commands unavailable

Resolution:
- Verify commands path in plugin.json: `"commands": ["./commands"]`
- Confirm .md extension on command files
- Check YAML frontmatter has description

Issue: Hooks Not Triggering

Symptom: Hook events not executing scripts

Resolution:
- Use full command with bash: `"bash ${CLAUDE_PLUGIN_ROOT}/hooks/script.sh"`
- Verify script is executable: `chmod +x script.sh`
- Check hooks.json syntax

Issue: Skills Not Loading

Symptom: Agent cannot find skills

Resolution:
- Verify skills path: `"skills": ["./skills"]`
- Confirm SKILL.md exists in skill directory
- Check skill name in agent matches directory name

Issue: Path Resolution Failures

Symptom: File not found errors

Resolution:
- Replace hardcoded paths with `${CLAUDE_PLUGIN_ROOT}`
- Verify referenced files exist in plugin
- Use absolute paths in hook commands

### Rollback Plan

Keep original configuration until migration verified:
```bash
# Backup original
cp -r .claude .claude.backup

# Test plugin
claude --plugin-dir ./my-plugin

# If issues, restore original
cp -r .claude.backup .claude
```

### Post-Migration Cleanup

After successful migration:
1. Remove migrated files from .claude/ (if not needed)
2. Update project documentation
3. Add plugin to version control
4. Create CHANGELOG.md for plugin
5. Consider publishing to marketplace
