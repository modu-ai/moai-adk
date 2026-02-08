---
paths:
  - "**/.moai/config/**"
  - "**/.mcp.json"
  - "**/.claude/settings.json"
  - "**/.claude/settings.local.json"
---

# Settings Management

Claude Code and MoAI configuration management rules.

## Configuration Files

### Claude Code Settings

`.claude/settings.json` - Project-level settings:

- allowedTools: Permitted tool list
- hooks: Hook script definitions
- permissions: Access control
- statusLine: Statusline configuration

### MCP Configuration

`.mcp.json` - MCP server definitions:

- mcpServers: Server command and arguments
- Environment variables for servers

### MoAI Configuration

`.moai/config/` - MoAI-specific settings:

- config.yaml: Main configuration
- sections/quality.yaml: Quality gates, coverage targets
- sections/language.yaml: Language preferences
- sections/user.yaml: User information

## Hooks Configuration

Hooks support environment variables and must be quoted to handle spaces:

```json
{
  "hooks": {
    "SessionStart": [{
      "type": "command",
      "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-session-start.sh\"",
      "timeout": 5
    }],
    "PreToolUse": [{
      "matcher": "Write|Edit|Bash",
      "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-pre-tool.sh\"",
      "timeout": 5
    }]
  }
}
```

**Important**: Quote the entire path: `"\"$CLAUDE_PROJECT_DIR/path\""` not `"$CLAUDE_PROJECT_DIR/path"`

## StatusLine Configuration

StatusLine does NOT support environment variables. Use relative paths from project root:

```json
{
  "statusLine": {
    "type": "command",
    "command": ".moai/status_line.sh"
  }
}
```

Reference: GitHub Issue #7925 - statusline does not expand environment variables.

## Permission Management

Tool permissions in settings.json:

- Read, Write, Edit: File operations
- Bash: Shell command execution
- Task: Agent delegation
- AskUserQuestion: User interaction

## Quality Configuration

Quality gates in quality.yaml:

- development_mode: ddd, tdd, or hybrid
- test_coverage_target: Minimum coverage percentage
- lsp_quality_gates: LSP-based validation

## Language Settings

Language preferences in language.yaml:

- conversation_language: User response language
- agent_prompt_language: Internal communication
- code_comments: Code comment language

## Agent Teams Settings

Agent Teams require both an environment variable and workflow configuration.

### Environment Variable

Enable in `.claude/settings.json`:

```json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
```

This env var must be set for Claude Code to expose the Teams API.

### Workflow Configuration

Team behavior is controlled by the `workflow.team` section in `.moai/config/sections/workflow.yaml`:

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| team.enabled | boolean | false | Master switch for team mode |
| team.max_teammates | integer | 10 | Maximum teammates per team (2-10 recommended) |
| team.default_model | string | inherit | Default model for teammates (inherit/haiku/sonnet/opus) |
| team.require_plan_approval | boolean | true | Require plan approval before implementing |
| team.delegate_mode | boolean | true | Team lead coordination-only mode (no direct implementation) |
| team.teammate_display | string | auto | Display mode: auto, in-process, or tmux |

### Auto-Selection Thresholds

When `workflow.execution_mode` is `auto`, these thresholds determine when team mode activates:

| Setting | Default | Description |
|---------|---------|-------------|
| team.auto_selection.min_domains_for_team | 3 | Minimum distinct domains to trigger team mode |
| team.auto_selection.min_files_for_team | 10 | Minimum affected files to trigger team mode |
| team.auto_selection.min_complexity_score | 7 | Minimum complexity score (1-10) to trigger team mode |

## Rules

- Never commit secrets to settings files
- Use environment variables for sensitive data
- Keep settings minimal and focused
- Hook paths must be quoted when using environment variables
- StatusLine uses relative paths only (no env var expansion)
- Template sources (.tmpl files) belong in `internal/template/templates/` only
- Local projects should contain rendered results, not template sources

## MoAI Integration

- Skill("moai-workflow-project") for project setup
- Skill("moai-foundation-core") for quality framework
- See hooks-system.md for detailed hook configuration patterns
