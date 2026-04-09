# Claude Code Hooks - Official Documentation Reference

Source: https://code.claude.com/docs/en/hooks

## Key Concepts

### What are Claude Code Hooks?

Hooks are powerful automation tools that extend Claude Code functionality by executing commands or prompts in response to specific events. They provide deterministic control over Claude Code's behavior through event-driven automation.

Security Warning: Hooks execute arbitrary shell commands with system credentials. Use with extreme caution.

### Hook System Architecture

Event Flow:
```
User Action → Event Trigger → Hook Execution → Result Processing
```

Hook Types:

- Command Hooks: Execute shell commands
- Prompt Hooks: Generate and execute prompts
- Validation Hooks: Validate inputs and outputs
- Notification Hooks: Send notifications or logs

## Core Hook Events

### Tool-Related Events

PreToolUse: Before tool execution
- Can block tool execution
- Perfect for validation and security checks
- Receives tool name and parameters

PostToolUse: After successful tool use
- Cannot block (post-execution)
- Ideal for logging and cleanup
- Receives execution results

PermissionRequest: When permission dialogs appear
- Can auto-approve or deny
- Useful for automation workflows
- Receives permission details

### Session-Related Events

SessionStart: When new session begins
- Initialize session state
- Set up environment variables
- Configure session-specific settings

SessionEnd: When session terminates
- Cleanup temporary files
- Save session state
- Generate session reports

SubagentStop: When sub-agent tasks complete
- Process sub-agent results
- Trigger follow-up actions
- Log completion status

Stop: When main agent finishes
- Final cleanup operations
- Generate completion reports
- Prepare for next session

### User Interaction Events

UserPromptSubmit: When user submits prompts
- Validate user input
- Modify prompts programmatically
- Add contextual information

## Hook Configuration Locations

Hooks can be configured in three locations with different capabilities:

### 1. Settings Files (Global/Project)

- Location: `~/.claude/settings.json` (user) or `.claude/settings.json` (project)
- Scope: All sessions in scope
- Features: Full hook types, matchers, timeouts
- Limitation: `once` field NOT supported

### 2. Skill/Slash Command Frontmatter (Component-scoped)

- Location: SKILL.md or command .md frontmatter
- Scope: Only when the skill/command is active
- Features: Full hook types, matchers, timeouts, `once` field
- Special: `once: true` runs hook only once per session

### 3. Agent Frontmatter (Agent-scoped)

- Location: Agent .md frontmatter
- Scope: Only when the agent is running
- Features: PreToolUse, PostToolUse, Stop hooks
- Limitation: `once` field NOT supported (agents only)

## Skill/Command Frontmatter Hooks (2026-01)

Skills and slash commands can define hooks directly in their YAML frontmatter. This is the ONLY location where the `once` field is supported.

### Basic Skill Hook Example

```yaml
---
name: secure-file-operations
description: File operations with security checks
hooks:
  PreToolUse:
    - matcher: "Write|Edit"
      hooks:
        - type: command
          command: "./scripts/security-check.sh $TOOL_INPUT"
          timeout: 30
  PostToolUse:
    - matcher: "Write"
      hooks:
        - type: command
          command: "./scripts/verify-write.sh"
---
```

### Using once: true (Skills Only)

The `once` field ensures a hook runs only once per session, regardless of how many times the tool is used:

```yaml
---
name: setup-skill
description: Skill with one-time initialization
hooks:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: "./init.sh"
          once: true
---
```

IMPORTANT: The `once` field is ONLY supported in skill/slash command frontmatter hooks. It is NOT supported in settings.json or agent frontmatter.

### Slash Command Hook Example

```yaml
---
name: deploy
description: Deploy application with pre-checks
hooks:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: "./scripts/deployment-check.sh"
          timeout: 60
          once: true
---
```

### Agent Frontmatter Hooks

Agents can also define hooks, but `once` is NOT supported:

```yaml
---
name: code-reviewer
description: Review code changes
hooks:
  PreToolUse:
    - matcher: "Edit"
      hooks:
        - type: command
          command: "./scripts/pre-edit-check.sh"
  PostToolUse:
    - matcher: "Edit|Write"
      hooks:
        - type: command
          command: "./scripts/run-linter.sh"
          timeout: 45
---
```

## Hook Configuration Structure

### Basic Configuration

Configure hooks in `settings.json`:

```json
{
 "hooks": {
 "PreToolUse": [
 {
 "matcher": "Bash",
 "hooks": [
 {
 "type": "command",
 "command": "echo 'Executing bash command:' >> ~/.claude/hooks.log"
 }
 ]
 }
 ]
 }
}
```

### Advanced Configuration

Multiple Event Handlers:
```json
{
 "hooks": {
 "PreToolUse": [
 {
 "matcher": "Bash",
 "hooks": [
 {
 "type": "command",
 "command": "validate-bash-command \"$COMMAND\"",
 "blocking": true
 },
 {
 "type": "prompt",
 "prompt": "Review bash command for security: $COMMAND"
 }
 ]
 },
 {
 "matcher": "Write",
 "hooks": [
 {
 "type": "command",
 "command": "backup-file \"$TARGET_PATH\""
 }
 ]
 }
 ]
 }
}
```

### Complex Hook Patterns

Conditional Execution:
```json
{
 "hooks": {
 "PreToolUse": [
 {
 "matcher": "Bash",
 "hooks": [
 {
 "type": "command",
 "command": "if [[ \"$COMMAND\" == *\"rm -rf\"* ]]; then exit 1; fi",
 "blocking": true
 }
 ]
 }
 ]
 }
}
```

## Hook Types and Usage

### Command Hooks

Shell Command Execution:
```json
{
 "type": "command",
 "command": "echo \"Tool: $TOOL_NAME, Args: $ARGUMENTS\" >> ~/claude-hooks.log",
 "env": {
 "HOOK_LOG_LEVEL": "debug"
 }
}
```

Available Variables:
- `$TOOL_NAME`: Name of the tool being executed
- `$ARGUMENTS`: Tool arguments as JSON string
- `$SESSION_ID`: Current session identifier
- `$USER_INPUT`: User's original input

### Prompt Hooks

Prompt Generation and Execution:
```json
{
 "type": "prompt",
 "prompt": "Review this command for security risks: $COMMAND\n\nProvide a risk assessment and recommendations.",
 "model": "claude-3-5-sonnet-20241022",
 "max_tokens": 500
}
```

Prompt Variables:
- All command hook variables available
- `$HOOK_CONTEXT`: Current hook execution context
- `$PREVIOUS_RESULTS`: Results from previous hooks

### Validation Hooks

Input/Output Validation:
```json
{
 "type": "validation",
 "pattern": "^[a-zA-Z0-9_\\-\\.]+$",
 "message": "File name contains invalid characters",
 "blocking": true
}
```

## Security Considerations

### Security Best Practices

Principle of Least Privilege:
```json
{
 "hooks": {
 "PreToolUse": [
 {
 "matcher": "Bash",
 "hooks": [
 {
 "type": "command",
 "command": "allowed_commands=(npm python git make)",
 "command": "if [[ ! \" ${allowed_commands[@]} \" =~ \" ${COMMAND%% *} \" ]]; then exit 1; fi",
 "blocking": true
 }
 ]
 }
 ]
 }
}
```

Input Sanitization:
```json
{
 "hooks": {
 "UserPromptSubmit": [
 {
 "hooks": [
 {
 "type": "command",
 "command": "echo \"$USER_INPUT\" | sanitize-input",
 "blocking": true
 }
 ]
 }
 ]
 }
}
```

### Dangerous Pattern Detection

Prevent Dangerous Commands:
```json
{
 "hooks": {
 "PreToolUse": [
 {
 "matcher": "Bash",
 "hooks": [
 {
 "type": "command",
 "command": "dangerous_patterns=(\"rm -rf\" \"sudo\" \"chmod 777\" \"dd\" \"mkfs\")",
 "command": "for pattern in \"${dangerous_patterns[@]}\"; do if [[ \"$COMMAND\" == *\"$pattern\"* ]]; then echo \"Dangerous command detected: $pattern\" >&2; exit 1; fi; done",
 "blocking": true
 }
 ]
 }
 ]
 }
}
```

## Hook Management

### Configuration Management

Using /hooks Command:
```bash
# Open hooks configuration editor
/hooks

# View current hooks configuration
/hooks --list

# Test hook functionality
/hooks --test
```

Settings File Locations:
- Global: `~/.claude/settings.json` (user-wide hooks)
- Project: `.claude/settings.json` (project-specific hooks)
- Local: `.claude/settings.local.json` (local overrides)

### Hook Lifecycle Management

Installation:
```bash
# Add hook to configuration
claude config set hooks.PreToolUse[0].matcher "Bash"
claude config set hooks.PreToolUse[0].hooks[0].type "command"
claude config set hooks.PreToolUse[0].hooks[0].command "echo 'Bash executed' >> hooks.log"

# Validate configuration
claude config validate
```

Testing and Debugging:
```bash
# Test individual hook
claude hooks test --event PreToolUse --tool Bash

# Debug hook execution
claude hooks debug --verbose

# View hook logs
claude hooks logs
```

## Common Hook Patterns

### Pre-Commit Validation

Code Quality Checks:
```json
{
 "hooks": {
 "PreToolUse": [
 {
 "matcher": "Bash",
 "hooks": [
 {
 "type": "command",
 "command": "if [[ \"$COMMAND\" == \"git commit\"* ]]; then npm run lint && npm test; fi",
 "blocking": true
 }
 ]
 }
 ]
 }
}
```

### Auto-Backup System

File Modification Backup:
```json
{
 "hooks": {
 "PreToolUse": [
 {
 "matcher": "Write",
 "hooks": [
 {
 "type": "command",
 "command": "cp \"$TARGET_PATH\" \"$TARGET_PATH.backup.$(date +%s)\""
 }
 ]
 },
 {
 "matcher": "Edit",
 "hooks": [
 {
 "type": "command",
 "command": "cp \"$TARGET_PATH\" \"$TARGET_PATH.backup.$(date +%s)\""
 }
 ]
 }
 ]
 }
}
```

### Session Logging

Comprehensive Activity Logging:
```json
{
 "hooks": {
 "PostToolUse": [
 {
 "hooks": [
 {
 "type": "command",
 "command": "echo \"$(date '+%Y-%m-%d %H:%M:%S') - Tool: $TOOL_NAME, Duration: $DURATION_MS ms, Success: $SUCCESS\" >> ~/.claude/session-logs/$SESSION_ID.log"
 }
 ]
 },
 {
 "matcher": "*",
 "hooks": [
 {
 "type": "command",
 "command": "echo \"$(date '+%Y-%m-%d %H:%M:%S') - Session: $SESSION_ID, Event: $EVENT_TYPE\" >> ~/.claude/activity.log"
 }
 ]
 }
 ]
 }
}
```

## Error Handling and Recovery

### Error Handling Strategies

Graceful Degradation:
```json
{
 "hooks": {
 "PreToolUse": [
 {
 "matcher": "Bash",
 "hooks": [
 {
 "type": "command",
 "command": "if ! validate-command \"$COMMAND\"; then echo \"Command validation failed, proceeding with caution\"; exit 0; fi",
 "blocking": false
 }
 ]
 }
 ]
 }
}
```

Fallback Mechanisms:
```json
{
 "hooks": {
 "PreToolUse": [
 {
 "matcher": "*",
 "hooks": [
 {
 "type": "command",
 "command": "primary-command \"$ARGUMENTS\" || fallback-command \"$ARGUMENTS\"",
 "fallback": {
 "type": "command",
 "command": "echo \"Primary hook failed, using fallback\""
 }
 }
 ]
 }
 ]
 }
}
```

## Performance Optimization

### Hook Performance

Asynchronous Execution:
```json
{
 "hooks": {
 "PostToolUse": [
 {
 "hooks": [
 {
 "type": "command",
 "command": "background-process \"$ARGUMENTS\" &",
 "async": true
 }
 ]
 }
 ]
 }
}
```

Conditional Hook Execution:
```json
{
 "hooks": {
 "PreToolUse": [
 {
 "matcher": "Bash",
 "condition": "$COMMAND != 'git status'",
 "hooks": [
 {
 "type": "command",
 "command": "complex-validation \"$COMMAND\""
 }
 ]
 }
 ]
 }
}
```

## Integration with Other Systems

### External Service Integration

Webhook Integration:
```json
{
 "hooks": {
 "SessionEnd": [
 {
 "hooks": [
 {
 "type": "command",
 "command": "curl -X POST https://api.example.com/webhook -d '{\"session_id\": \"$SESSION_ID\", \"events\": \"$EVENT_COUNT\"}'"
 }
 ]
 }
 ]
 }
}
```

Database Logging:
```json
{
 "hooks": {
 "PostToolUse": [
 {
 "hooks": [
 {
 "type": "command",
 "command": "psql -h localhost -u claude -d hooks -c \"INSERT INTO tool_usage (session_id, tool_name, timestamp) VALUES ('$SESSION_ID', '$TOOL_NAME', NOW())\""
 }
 ]
 }
 ]
 }
}
```

## Best Practices

### Development Guidelines

Hook Development Checklist:
- [ ] Test hooks in isolation before deployment
- [ ] Implement proper error handling and logging
- [ ] Use non-blocking hooks for non-critical operations
- [ ] Validate all inputs and sanitize outputs
- [ ] Document hook dependencies and requirements
- [ ] Implement graceful fallbacks for critical operations
- [ ] Monitor hook performance and resource usage
- [ ] Regular security audits and permission reviews

Performance Guidelines:
- Keep hook execution time under 100ms for critical paths
- Use asynchronous execution for non-blocking operations
- Minimize file I/O operations in hot paths
- Cache frequently used data and configuration
- Implement rate limiting for external API calls

Security Guidelines:
- Never expose sensitive credentials in hook commands
- Validate and sanitize all user inputs
- Use principle of least privilege for file system access
- Implement proper access controls for external integrations
- Regular security reviews and penetration testing

This comprehensive reference provides all the information needed to create, configure, and manage Claude Code Hooks effectively and securely.
