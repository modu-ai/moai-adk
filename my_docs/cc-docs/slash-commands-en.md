# Slash Commands

Extend Claude Code with built-in commands and create custom workflow automation through slash commands.

## Built-in Slash Commands

Claude Code provides a comprehensive set of built-in slash commands for session management, configuration, and workflow automation:

### Session Management

| Command | Purpose | Example |
|---------|---------|---------|
| `/clear` (reset) | Clear conversation history and free up context | `/clear` |
| `/compact [instructions]` | Clear conversation history but keep a summary in context | `/compact "keep focus on authentication"` |
| `/resume` | Resume a conversation from session list | `/resume` |
| `/exit` (quit) | Exit the REPL | `/exit` |
| `/export` | Export the current conversation to a file or clipboard | `/export` |

### Configuration and Setup

| Command | Purpose | Example |
|---------|---------|---------|
| `/config` (theme) | Open config panel | `/config` |
| `/init` | Initialize a new CLAUDE.md file with codebase documentation | `/init` |
| `/login` | Sign in with your Anthropic account | `/login` |
| `/logout` | Sign out from your Anthropic account | `/logout` |
| `/model` | Set the AI model for Claude Code | `/model` |
| `/output-style` | Set the output style directly or from a selection menu | `/output-style` |
| `/output-style:new` | Create a custom output style | `/output-style:new` |
| `/permissions` (allowed-tools) | Manage allow & deny tool permission rules | `/permissions` |
| `/statusline` | Set up Claude Code's status line UI | `/statusline` |
| `/vim` | Toggle between Vim and Normal editing modes | `/vim` |

### Development Workflow

| Command | Purpose | Example |
|---------|---------|---------|
| `/add-dir` | Add a new working directory | `/add-dir ../docs` |
| `/agents` | Manage agent configurations | `/agents` |
| `/hooks` | Manage hook configurations for tool events | `/hooks` |
| `/memory` | Edit Claude memory files | `/memory` |
| `/mcp` | Manage MCP servers | `/mcp` |
| `/review` | Review a pull request | `/review` |
| `/security-review` | Complete a security review of the pending changes on the current branch | `/security-review` |
| `/pr-comments` | Get comments from a GitHub pull request | `/pr-comments` |

### System and Diagnostics

| Command | Purpose | Example |
|---------|---------|---------|
| `/status` | Show Claude Code status including version, model, account, API connectivity, and tool statuses | `/status` |
| `/cost` | Show the total cost and duration of the current session | `/cost` |
| `/doctor` | Diagnose and verify your Claude Code installation and settings | `/doctor` |
| `/help` | Show help and available commands | `/help` |
| `/bashes` | List and manage background bash shells | `/bashes` |

### Integration and Extensions

| Command | Purpose | Example |
|---------|---------|---------|
| `/ide` | Manage IDE integrations and show status | `/ide` |
| `/install-github-app` | Set up Claude GitHub Actions for a repository | `/install-github-app` |
| `/terminal-setup` | Install Shift+Enter key binding for newlines | `/terminal-setup` |
| `/migrate-installer` | Migrate from global npm installation to local installation | `/migrate-installer` |

### Account and Updates

| Command | Purpose | Example |
|---------|---------|---------|
| `/upgrade` | Upgrade to Max for higher rate limits and more Opus | `/upgrade` |
| `/release-notes` | View release notes | `/release-notes` |
| `/bug` | Submit feedback about Claude Code | `/bug` |

## Custom Slash Commands

### Overview

Custom slash commands allow you to define frequently-used prompts as reusable Markdown files. They provide a powerful way to standardize workflows, share best practices with teams, and automate common development tasks.

### Command Syntax

```bash
/<command-name> [arguments]
```

### Command Locations

#### Project Commands

- **Location**: `.claude/commands/`
- **Scope**: Shared with team via source control
- **Display**: Shown with "(project)" label in `/help`
- **Use Case**: Team workflows, project-specific tasks

#### Personal Commands

- **Location**: `~/.claude/commands/`
- **Scope**: Available across all projects for individual user
- **Display**: Shown with "(user)" label in `/help`
- **Use Case**: Personal workflows, individual preferences

### Creating Custom Commands

#### Basic Command Structure

Create a Markdown file in the appropriate commands directory:

```bash
# Project-specific command
mkdir -p .claude/commands
echo "Analyze this code for performance issues and suggest optimizations:" > .claude/commands/optimize.md

# Personal command
mkdir -p ~/.claude/commands
echo "Review this code for security vulnerabilities:" > ~/.claude/commands/security-review.md
```

#### Command with Frontmatter

Use YAML frontmatter to define metadata:

```markdown
---
allowed-tools: Bash(git add:*), Bash(git status:*), Bash(git commit:*)
argument-hint: [message]
description: Create a git commit
model: claude-3-5-haiku-20241022
---

Create a git commit with message: $ARGUMENTS
```

### Frontmatter Options

| Option | Description | Example |
|--------|-------------|---------|
| `allowed-tools` | Specify permitted tools for command execution | `Bash(git:*), Read, Write` |
| `argument-hint` | Display hint for expected arguments | `[pr-number] [priority] [assignee]` |
| `description` | Brief description shown in help | `Create a git commit` |
| `model` | Specific model to use for this command | `claude-3-5-sonnet-20241022` |

### Argument Handling

#### Using $ARGUMENTS

Capture all arguments passed to the command:

```bash
# Command definition
echo 'Fix issue #$ARGUMENTS following our coding standards' > .claude/commands/fix-issue.md

# Usage
> /fix-issue 123 high-priority
# $ARGUMENTS becomes: "123 high-priority"
```

#### Using Positional Arguments

Access specific arguments individually:

```bash
# Command definition
echo 'Review PR #$1 with priority $2 and assign to $3' > .claude/commands/review-pr.md

# Usage
> /review-pr 456 high alice
# $1 = "456", $2 = "high", $3 = "alice"
```

### Advanced Features

#### Executing Bash Commands

Use the `!` prefix to execute bash commands and include their output:

```markdown
---
allowed-tools: Bash(git add:*), Bash(git status:*), Bash(git commit:*)
description: Create a git commit
---

## Context

- Current git status: !`git status`
- Current git diff (staged and unstaged changes): !`git diff HEAD`
- Current branch: !`git branch --show-current`
- Recent commits: !`git log --oneline -10`

## Your task

Based on the above changes, create a single git commit.
```

#### File References

Reference files using the `@` prefix:

```markdown
Review the code in @src/auth.js and suggest improvements based on @docs/security-guidelines.md
```

#### Environment Variables

Access environment variables in commands:

```markdown
Deploy to $ENV_NAME environment with configuration from @config/$ENV_NAME.json
```

### Complex Command Examples

#### Git Workflow Command

```markdown
---
allowed-tools: Bash(git:*), Read, Write
argument-hint: [branch-name] [feature-description]
description: Create feature branch and initial commit
---

Create a new feature branch named $1 for: $2

1. Create and checkout the branch
2. Make initial changes
3. Commit with descriptive message
4. Push to remote with tracking
```

#### Code Review Command

```markdown
---
allowed-tools: Read, Grep, Bash(git diff:*)
argument-hint: [pr-number] [priority] [assignee]
description: Review pull request
---

Review PR #$1 with priority $2 and assign to $3.

Focus on:
- Security vulnerabilities
- Performance implications
- Code style and maintainability
- Test coverage
```

#### Issue Resolution Command

```markdown
---
allowed-tools: Bash, Read, Write, Edit
argument-hint: [issue-number]
description: Find and fix issue
---

Find and fix issue #$ARGUMENTS. Follow these steps:

1. Understand the issue described in the ticket
2. Locate the relevant code in our codebase
3. Implement a solution that addresses the root cause
4. Add appropriate tests
5. Prepare a concise PR description
```

## MCP Slash Commands

### Overview

MCP (Model Context Protocol) slash commands are dynamically discovered from connected MCP servers, providing seamless integration with external tools and services.

### Command Format

```bash
/mcp__<server-name>__<prompt-name> [arguments]
```

### Discovering MCP Commands

Type `/` to see all available commands, including dynamically discovered MCP commands.

### MCP Command Examples

#### Without Arguments

```bash
# List GitHub pull requests
> /mcp__github__list_prs

# Check Linear issues
> /mcp__linear__list_issues
```

#### With Arguments

```bash
# Review specific GitHub PR
> /mcp__github__pr_review 456

# Create Jira issue with priority
> /mcp__jira__create_issue "Bug in login flow" high

# Update Notion page
> /mcp__notion__update_page "Meeting Notes" "Updated content"
```

### MCP Server Integration

MCP commands are automatically available when you have configured MCP servers:

```bash
# Configure MCP servers
claude mcp add github --env GITHUB_TOKEN=your_token -- npx -y github-mcp-server
claude mcp add linear --transport sse https://mcp.linear.app/sse
claude mcp add notion --transport http https://mcp.notion.com/mcp
```

## Command Organization and Namespacing

### Directory Structure

Organize commands using subdirectories for better management:

```
.claude/commands/
├── git/
│   ├── commit.md
│   ├── review.md
│   └── deploy.md
├── testing/
│   ├── unit.md
│   ├── integration.md
│   └── e2e.md
└── docs/
    ├── generate.md
    └── update.md
```

### Namespace Usage

```bash
# Access namespaced commands
> /git/commit "Fix authentication bug"
> /testing/unit
> /docs/generate
```

## Best Practices

### Command Design

1. **Single Responsibility**: Each command should have a clear, focused purpose
2. **Descriptive Names**: Use action-oriented names that clearly indicate what the command does
3. **Consistent Arguments**: Use predictable argument patterns across similar commands
4. **Comprehensive Frontmatter**: Always include description and argument hints

### Security and Permissions

1. **Minimal Tool Access**: Only grant necessary tools in `allowed-tools`
2. **Validate Inputs**: Design commands to handle unexpected arguments gracefully
3. **Avoid Hardcoded Secrets**: Use environment variables for sensitive data
4. **Review Shared Commands**: Audit project commands before committing

### Team Collaboration

1. **Version Control**: Include project commands in source control
2. **Documentation**: Document complex commands with clear examples
3. **Consistent Style**: Establish team conventions for command structure
4. **Regular Review**: Periodically review and update command effectiveness

### Performance Optimization

1. **Efficient Tools**: Choose the most appropriate tools for each task
2. **Cache Results**: Use bash command output caching where beneficial
3. **Limit Scope**: Keep commands focused to avoid unnecessary work
4. **Model Selection**: Use appropriate models for command complexity

## Command Development Workflow

### Creating New Commands

1. **Identify Need**: Recognize repetitive tasks or common workflows
2. **Design Interface**: Plan arguments, tools, and expected behavior
3. **Implement**: Create the command file with appropriate frontmatter
4. **Test**: Verify command works as expected with various inputs
5. **Document**: Add clear description and argument hints
6. **Share**: Commit project commands or document personal commands

### Testing Commands

```bash
# Test basic functionality
> /your-command test-arg

# Test with various argument patterns
> /your-command
> /your-command single-arg
> /your-command multiple args here

# Verify tool permissions
> /your-command --verbose
```

### Debugging Commands

1. **Check Command Discovery**: Ensure commands appear in `/help`
2. **Verify Permissions**: Confirm required tools are allowed
3. **Test Arguments**: Validate argument substitution works correctly
4. **Review Output**: Check that bash commands execute as expected

## Integration with Other Features

### Hooks Integration

Commands can trigger hooks for additional automation:

```markdown
---
allowed-tools: Bash(git:*), Write
description: Deploy with notification
---

Deploy the application and notify team:

1. Run deployment scripts
2. Update deployment log
3. Send team notification via webhook
```

### Subagent Coordination

Commands can work with subagents for complex workflows:

```markdown
Use the deployment-specialist subagent to deploy $1 environment.
Then use the testing-agent to run smoke tests.
Finally, use the documentation-agent to update the deployment log.
```

### Memory System Integration

Commands can reference and update memory files:

```markdown
Update @CLAUDE.md with new deployment procedures based on $ARGUMENTS.
Include lessons learned and optimization suggestions.
```

## Advanced Command Patterns

### Conditional Logic

```markdown
If tests pass, deploy to staging.
If staging tests pass, create production deployment PR.
Otherwise, report failures and suggested fixes.
```

### Command Chaining

```markdown
First run /test to ensure code quality.
If tests pass, run /security-review.
Finally, if both pass, run /deploy staging.
```

### Dynamic Content

```markdown
---
allowed-tools: Bash(git:*), Read
---

Analyze changes since last release:

Recent commits: !`git log --oneline --since="1 week ago"`
Changed files: !`git diff --name-only HEAD~10..HEAD`

Based on these changes, suggest release notes and version bump.
```

Slash commands provide a powerful way to standardize and automate development workflows, making Claude Code an integral part of your development process. They enable teams to codify best practices, reduce repetitive tasks, and maintain consistency across projects.