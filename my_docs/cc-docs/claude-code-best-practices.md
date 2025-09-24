# Claude Code: Best Practices for Agentic Coding

## Overview

Claude Code is a command line tool for agentic coding, designed to provide flexible and powerful integration of AI into coding workflows. The tool is intentionally low-level and unopinionated, allowing engineers to customize their experience.

## Key Best Practices

### 1. Customize Your Setup

#### Create `CLAUDE.md` Files

- Use `CLAUDE.md` to document:
  - Common bash commands
  - Core files and utility functions
  - Code style guidelines
  - Testing instructions
  - Repository etiquette
  - Developer environment setup
  - Unexpected behaviors or warnings

Example `CLAUDE.md`:

```markdown
# Bash commands

- npm run build: Build the project
- npm run typecheck: Run the typechecker

# Code style

- Use ES modules (import/export) syntax
- Destructure imports when possible
```

#### Manage Permissions

- Customize tool allowlist through:
  - Selecting "Always allow" during sessions
  - Using `/permissions` command
  - Manually editing `.claude/settings.json`
  - Using `--allowedTools` CLI flag

### 2. Expand Claude's Capabilities

#### Use Bash Tools

- Provide instructions for custom bash tools
- Tell Claude about tool usage and help documentation
- Document frequently used tools in `CLAUDE.md`

#### Leverage MCP (Model Context Protocol)

- Connect to multiple MCP servers
- Configure through project, global, or `.mcp.json` settings

#### Create Custom Slash Commands

- Store prompt templates in `.claude/commands`
- Use `$ARGUMENTS` for dynamic command parameters

### 3. Recommended Workflows

#### Explore, Plan, Code, Commit

1. Read relevant files
2. Create a plan (use "think" modes)
3. Implement solution
4. Commit and create pull request

#### Test-Driven Development

1. Write tests
2. Confirm tests fail
3. Implement code to pass tests
4. Commit changes

#### Visual Design Iteration

1. Provide screenshots or visual mocks
2. Let Claude implement the design
3. Iterate based on feedback

### 4. Effective Prompting Techniques

#### Be Specific

- Provide clear, explicit instructions
- Include context about your goals
- Share examples of desired output

#### Use Think Modes

- For complex tasks, instruct Claude to think through the problem
- Use prompts like "Think step by step about this problem"

#### Batch Operations

- Request multiple related tasks in a single prompt
- Use parallel tool calls for efficiency

### 5. Team Collaboration

#### Share Project Memory

- Check in `CLAUDE.md` to version control
- Document team conventions and standards
- Include onboarding instructions

#### Create Custom Agents

- Design specialized agents for common tasks
- Share agent configurations with team
- Use consistent naming conventions

## Common Patterns

### Debugging Workflow

```
1. Describe the issue
2. Share error messages
3. Let Claude investigate with tools
4. Review proposed fixes
5. Test and iterate
```

### Feature Implementation

```
1. Share requirements
2. Let Claude explore codebase
3. Review implementation plan
4. Execute implementation
5. Run tests
6. Commit changes
```

### Code Review

```
1. Share diff or files
2. Request specific review focus
3. Address feedback
4. Update code
```

## Performance Tips

- Use parallel tool calls for faster execution
- Limit context to relevant files
- Clear conversation when switching tasks
- Use agents for specialized workflows

## Security Considerations

- Review permissions regularly
- Limit tool access appropriately
- Never share sensitive credentials
- Use environment variables for secrets
