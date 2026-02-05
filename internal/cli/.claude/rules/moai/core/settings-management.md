# Settings Management

Claude Code and MoAI configuration management rules.

## Configuration Files

### Claude Code Settings

`.claude/settings.json` - Project-level settings:

- allowedTools: Permitted tool list
- hooks: Hook script definitions
- permissions: Access control

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

## MoAI Integration

- Skill("moai-workflow-project") for project setup
- Skill("moai-foundation-core") for quality framework
