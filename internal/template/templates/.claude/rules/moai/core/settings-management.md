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
