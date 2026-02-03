# Claude Code Sub-agents - Official Documentation Reference

Source: https://code.claude.com/docs/ko/sub-agents
Updated: 2026-01-06

## What are Sub-agents?

Sub-agents are specialized AI assistants that Claude Code can delegate tasks to. Each sub-agent has:

- A specific purpose and domain expertise
- Its own separate context window
- Configurable tools with granular access control
- A custom system prompt that guides behavior

When Claude encounters a task matching a sub-agent's specialty, it can delegate work to that specialized assistant while the main conversation remains focused on high-level goals.

## Key Benefits

Context Preservation: Each sub-agent operates in isolation, preventing main conversation pollution

Specialized Expertise: Fine-tuned with detailed domain instructions for higher success rates

Reusability: Created once, used across projects and shareable with teams

Flexible Permissions: Each can have different tool access levels for security

## Creating Sub-agents

### Quick Start Using /agents Command (Recommended)

Step 1: Open the agents interface by typing /agents

Step 2: Select "Create New Agent" (project or user level)

Step 3: Define the sub-agent:
- Describe its purpose and when to use it
- Select tools (or leave blank to inherit all)
- Press `e` to edit the system prompt in your editor
- Recommended: Have Claude generate it first, then customize

### Direct File Creation

Create markdown files with YAML frontmatter in the appropriate location:

Project Sub-agents: .claude/agents/agent-name.md
Personal Sub-agents: ~/.claude/agents/agent-name.md

## Configuration

### File Format

```yaml
---
name: your-sub-agent-name
description: Description of when this subagent should be invoked
tools: tool1, tool2, tool3
model: sonnet
---

Your subagent's system prompt goes here. This can be multiple paragraphs
and should clearly define the subagent's role, capabilities, and approach
to solving problems.
```

### Configuration Fields

Required Fields:

- name: Unique identifier using lowercase letters and hyphens

- description: Natural language explanation of purpose. Include phrases like "use PROACTIVELY" or "MUST BE USED" to encourage automatic invocation.

Optional Fields:

- tools: Comma-separated tool list. If omitted, inherits all available tools.

- model: Model alias (sonnet, opus, haiku) or 'inherit' to use same model as main conversation. If omitted, uses configured default (usually sonnet).

- permissionMode: Controls permission handling. Valid values: `default`, `acceptEdits`, `dontAsk`, `bypassPermissions`, `plan`, `ignore`.

- skills: Comma-separated list of skill names to auto-load when agent is invoked. Skills are NOT inherited from parent.

- hooks: Define lifecycle hooks scoped to this agent. Supports PreToolUse, PostToolUse, Stop events. Note: `once` field is NOT supported in agent hooks.

### Hooks Configuration (2026-01)

Agents can define hooks in their frontmatter that only run when the agent is active:

```yaml
---
name: code-reviewer
description: Review code changes with quality checks
tools: Read, Grep, Glob, Bash
model: inherit
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

Hook Fields:
- matcher: Regex pattern to match tool names (e.g., "Edit", "Write|Edit", "Bash")
- hooks: Array of hook definitions
  - type: "command" (shell) or "prompt" (LLM)
  - command: Shell command to execute
  - timeout: Timeout in seconds (default: 60)

IMPORTANT: The `once` field is NOT supported in agent hooks. Use skill hooks if you need one-time execution.

### Storage Locations and Priority

Sub-agents are stored as markdown files with YAML frontmatter:

1. Project Level: .claude/agents/ (highest priority)
2. User Level: ~/.claude/agents/ (lower priority)

Project-level definitions take precedence over user-level definitions with the same name.

## Using Sub-agents

### Automatic Delegation

Claude proactively delegates tasks based on:

- Request description matching sub-agent descriptions
- Sub-agent's description field content
- Current context and available tools

Tip: Include phrases like "use PROACTIVELY" or "MUST BE USED" in descriptions to encourage automatic invocation.

### Explicit Invocation

Request specific sub-agents directly:

- "Use the code-reviewer subagent to check my recent changes"
- "Have the debugger subagent investigate this error"

### Sub-agent Chaining

Chain multiple sub-agents for complex workflows:

"First use the code-analyzer subagent to find performance issues, then use the optimizer subagent to fix them"

## Model Selection

Available model options:

- sonnet: Balanced performance and quality (default)
- opus: Highest quality, higher cost
- haiku: Fastest, most cost-effective
- inherit: Use same model as main conversation

If model field is omitted, uses the configured default (usually sonnet).

## Built-in Sub-agents

### Plan Sub-agent

Purpose: Used during plan mode to research codebases
Model: Sonnet (for stronger analysis)
Tools: Read, Glob, Grep, Bash
Auto-invoked: When in plan mode and codebase investigation is needed
Behavior: Prevents infinite nesting of sub-agents while enabling context gathering

## Resumable Agents

Each sub-agent execution gets a unique agentId. Conversations are stored in agent-{agentId}.jsonl format. You can resume previous agent context with full context preserved:

"Resume agent abc123 and now analyze the authorization logic"

Use Cases for Resumable Agents:

- Long-running research tasks
- Iterative improvements
- Multi-step workflows spanning multiple sessions

## CLI-based Configuration

Define sub-agents dynamically via --agents flag:

```bash
claude --agents '{
  "code-reviewer": {
    "description": "Expert code reviewer. Use proactively after code changes.",
    "prompt": "You are a senior code reviewer. Focus on code quality, security, and best practices.",
    "tools": ["Read", "Grep", "Glob", "Bash"],
    "model": "sonnet"
  }
}'
```

Priority Order: CLI definitions have lowest priority, followed by User-level, then Project-level (highest).

## Managing Sub-agents with /agents Command

The /agents command provides an interactive menu to:

- View all available sub-agents (built-in, user, project)
- Create new sub-agents with guided setup
- Edit existing custom sub-agents and tool access
- Delete custom sub-agents
- Manage tool permissions with full available tools list

## Practical Examples

### Code Reviewer

```yaml
---
name: code-reviewer
description: Expert code review specialist. Proactively reviews code for quality, security, and maintainability. Use immediately after writing or modifying code.
tools: Read, Grep, Glob, Bash
model: inherit
---

You are a senior code reviewer ensuring high standards of code quality and security.

When invoked:
1. Run git diff to see recent changes
2. Focus on modified files
3. Begin review immediately

Review checklist:
- Code is simple and readable
- Functions and variables are well-named
- No duplicated code
- Proper error handling
- No exposed secrets or API keys
- Input validation implemented
- Good test coverage
- Performance considerations adddessed
```

### Debugger

```yaml
---
name: debugger
description: Debugging specialist for errors, test failures, and unexpected behavior. Use proactively when encountering any issues.
tools: Read, Edit, Bash, Grep, Glob
---

You are an expert debugger specializing in root cause analysis.

Debugging process:
- Analyze error messages and logs
- Check recent code changes
- Form and test hypotheses
- Add strategic debug logging
- Inspect variable states

For each issue, provide:
- Root cause explanation
- Evidence supporting the diagnosis
- Specific code fix
- Testing approach
- Prevention recommendations
```

### Data Scientist

```yaml
---
name: data-scientist
description: Data analysis expert for SQL queries and data insights. Use proactively for data analysis tasks.
tools: Bash, Read, Write
model: sonnet
---

You are a data scientist specializing in SQL and BigQuery analysis.

Key practices:
- Write optimized SQL queries with proper filters
- Use appropriate aggregations and joins
- Include comments explaining complex logic
- Format results for readability
- Provide data-driven recommendations
```

## Integration Patterns

### Sequential Delegation

Execute tasks in order, passing results between agents:

Phase 1 Analysis: Invoke spec-builder subagent to analyze requirements
Phase 2 Implementation: Invoke backend-expert subagent with analysis results
Phase 3 Validation: Invoke quality-gate subagent to validate implementation

### Parallel Delegation

Execute independent tasks simultaneously:

Invoke backend-expert, frontend-expert, and test-engineer subagents in parallel for independent implementation tasks

### Conditional Delegation

Route based on analysis results:

Based on analysis findings, route to database-expert for database issues or backend-expert for API issues

## Context Management

### Efficient Data Passing

- Pass only essential information between agents
- Use structured data formats for complex information
- Minimize context size for performance optimization
- Include validation metadata when appropriate

### Context Size Guidelines

- Each Task() creates independent context window
- Each sub-agent operates in its own 200K token session
- Recommended context size: 20K-50K tokens maximum for passed data
- Large datasets should be referenced rather than embedded

## Tool Permissions

Security Principle: Apply least privilege by only granting tools necessary for the agent's domain.

Common Tool Categories:

Read Tools: Read, Grep, Glob (file system access)
Write Tools: Write, Edit, MultiEdit (file modification)
System Tools: Bash (command execution)
Communication Tools: AskUserQuestion, WebFetch (interaction)

Available tools include Claude Code's internal tool set plus any connected MCP server tools.

## Critical Limitations

Sub-agents Cannot Spawn Other Sub-agents: This is a fundamental limitation to prevent infinite recursion. All delegation must flow from the main conversation or command.

Sub-agents Cannot Use AskUserQuestion Effectively: Sub-agents operate in isolated, stateless contexts and cannot interact with users directly. All user interaction must happen in the main conversation before delegating to sub-agents.

Required Pattern: All sub-agent delegation must use the Task() function.

## Best Practices

### 1. Start with Claude

Have Claude generate initial sub-agents, then customize based on your needs.

### 2. Single Responsibility

Design focused sub-agents with clear, single purposes. Each agent should excel at one domain.

### 3. Detailed Prompts

Include specific instructions, examples, and constraints in the system prompt.

### 4. Limit Tool Access

Grant only necessary tools for the sub-agent's role following least privilege principle.

### 5. Version Control

Check in project sub-agents to enable team collaboration through git.

### 6. Clear Descriptions

Make description specific and action-oriented. Include trigger scenarios.

## Testing and Validation

Test Categories:

1. Functionality Testing: Agent performs expected tasks correctly
2. Integration Testing: Agent works properly with other agents
3. Security Testing: Agent respects security boundaries
4. Performance Testing: Agent operates efficiently within token limits

Validation Steps:

1. Test agent behavior with various inputs
2. Verify tool usage respects permissions
3. Validate error handling and recovery
4. Check integration with other agents or skills

## Error Handling

Common Error Types:

- Agent Not Found: Incorrect agent name or file not found
- Permission Denied: Insufficient tool permissions
- Context Overflow: Too much context passed between agents
- Infinite Recursion Attempt: Agent tries to spawn another sub-agent

Recovery Strategies:

- Fallback to basic functionality
- User notification with clear error messages
- Graceful degradation of complex features
- Context optimization for retry attempts

## Security Considerations

Access Control:

- Apply principle of least privilege
- Validate all external inputs
- Restrict file system access where appropriate
- Audit tool usage regularly

Data Protection:

- Never pass sensitive credentials between agents
- Sanitize inputs before processing
- Use secure communication channels
- Log agent activities appropriately
