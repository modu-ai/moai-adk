# Subagents in Claude Code

## Overview

Custom subagents in Claude Code are specialized AI assistants that can be invoked to handle specific types of tasks. They enable more efficient problem-solving by providing task-specific configurations with customized system prompts, tools and a separate context window.

## What are Subagents?

Subagents are specialized AI assistants within Claude Code designed to handle specific tasks with unique characteristics:

- Have a specific purpose and expertise area
- Operate in a separate context window
- Can be configured with specific tool access
- Include a custom system prompt to guide behavior
- Proactively delegate tasks based on their defined purpose

## Key Features

### Architecture and Isolation

Subagents enable developers to create independent, task-specific AI agents with their own context, tools, and prompts. Designed for modular development, subagents can be orchestrated automatically or invoked manually, allowing teams to delegate work such as debugging, documentation, or test generation without overloading a single-context window.

Subagents operate in isolation from each other and the main agent, reducing the risk of context spillover and enabling more predictable task execution.

### Benefits

1. **Context Preservation**: Keeps main conversation focused and prevents context pollution
2. **Specialized Expertise**: Fine-tuned for specific domains and tasks
3. **Reusability**: Can be used across different projects and shared with teams
4. **Flexible Permissions**: Controllable tool access for security and focus
5. **Modular Workflows**: Enable sophisticated multi-agent coordination

## Creating and Managing Subagents

### File Structure

Each subagent is defined in a Markdown file with YAML frontmatter and stored either in a project-specific directory or a global user directory. Project-specific agents take precedence, supporting customization per project.

**File Locations:**
- Project subagents: `.claude/agents/`
- User subagents: `~/.claude/agents/`

When subagent names conflict, project-level subagents take precedence over user-level subagents.

### Subagent Structure

Each subagent is defined in a Markdown file with this structure:

```markdown
---
name: your-sub-agent-name
description: Description of when this subagent should be invoked
tools: tool1, tool2, tool3 # Optional - inherits all tools if omitted
---
Your subagent's system prompt goes here.
```

### Creation Methods

#### Interactive Management

The `/agents` command provides a comprehensive interface for subagent management:

1. Run `/agents` command
2. Select "Create New Agent"
3. Choose project-level or user-level subagent
4. Define the subagent:
   - Name
   - Description
   - Optional tool restrictions
   - Detailed system prompt

The Claude Code CLI provides an interactive workflow for managing subagents. Developers can scaffold a new agent through guided prompts, then edit the generated file in their preferred text editor. This hybrid approach combines automation with full developer control, fitting into existing development practices without requiring a new IDE or workflow.

#### Manual Creation Example

```bash
# Create a project subagent
mkdir -p .claude/agents
echo '---
name: test-runner
description: Use proactively to run tests and fix failures
---
You are a test automation expert.
When you see code changes, proactively run the appropriate tests.
If tests fail, analyze the failures and fix them while preserving the original test intent.' > .claude/agents/test-runner.md
```

### Example Subagent Configuration

```markdown
---
name: code-reviewer
description: Expert code review specialist. Proactively reviews code for quality, security, and maintainability.
tools: Read, Grep, Glob, Bash
model: sonnet
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
- Performance considerations addressed

Provide feedback organized by priority:

- Critical issues (must fix)
- Warnings (should fix)
- Suggestions (consider improving)

Include specific examples of how to fix issues.
```

## Using Subagents

### Invocation Methods

#### Manual Invocation
Request a specific subagent by mentioning it in your command:
- "Use the test-runner subagent to fix failing tests"
- "Have the code-reviewer subagent look at my recent changes"
- "Ask the debugger subagent to investigate this error"

#### Automatic Orchestration
Developers can trigger subagents manually for direct control, or rely on Claude Code's orchestration engine to match tasks with the most suitable subagents automatically. Claude Code automatically delegates tasks to appropriate subagents based on task description and context.

### Proactive Usage

To encourage more proactive subagent use, include phrases like "use PROACTIVELY" or "MUST BE USED" in your description field.

## Security and Permissions

Security and permission management are built into the subagent architecture. Each subagent's configuration explicitly lists the tools it's allowed to access, such as running shell commands or accessing external resources.

Anthropic's documentation recommends granting only the minimum set of permissions required for each subagent's role, limiting the blast radius in sensitive environments.

Subagents can be granted access to any of Claude Code's internal tools. See the tools documentation for a complete list of available tools.

## Best Practices

### Design Principles

1. **Start with Claude-generated agents**: We highly recommend generating your initial subagent with Claude and then iterating on it to make it personally yours. This approach gives you the best results - a solid foundation that you can customize to your specific needs.

2. **Design focused subagents**: Create subagents with single, clear responsibilities rather than trying to make one subagent do everything. This improves performance and makes subagents more predictable.

### Implementation Guidelines

3. **Write detailed prompts**: Include specific instructions, examples, and constraints in your system prompts. The more guidance you provide, the better the subagent will perform.

4. **Limit tool access**: Only grant tools that are necessary for the subagent's purpose. This improves security and helps the subagent focus on relevant actions.

5. **Version control**: Check project subagents into version control so your team can benefit from and improve them collaboratively.

## Advanced Patterns

### Multi-Agent Coordination

Subagents coordinate automatically for complex tasks. The system intelligently sequences multiple specialists based on task requirements.

### Subagent Chaining

Chain multiple subagents for complex workflows:

```
First use code-analyzer to find issues, then optimizer to fix them
```

### Parallel Execution

Beyond standalone usage, some of the most powerful applications involve running multiple Claude instances in parallel:

- A simple but effective approach is to have one Claude write code while another reviews or tests it
- You can do something similar with tests: have one Claude write tests, then have another Claude write code to make the tests pass
- You can even have your Claude instances communicate with each other by giving them separate working scratchpads and telling them which one to write to and which one to read from

This separation often yields better results than having a single Claude handle everything.

### Dynamic Selection

Claude Code intelligently selects subagents based on context and task requirements.

## Community Resources

A large collection of community-created subagents has emerged and is available on the internet for Claude Code users to leverage and learn from. This includes GitHub repositories containing over 60 specialized subagents organized into various domains, including:

- Development & Architecture
- Language Specialist
- Infrastructure & Operations
- Business & Marketing
- And more

## Performance Considerations

- **Context efficiency**: Preserves main context for longer sessions
- **Latency**: Subagents start fresh each invocation
- **Isolation**: Each subagent operates independently, preventing context spillover
- **Modular execution**: Enables complex workflows without overwhelming single context

The subagent system represents a significant advancement in modular AI development, allowing teams to build sophisticated multi-agent workflows while maintaining clear separation of concerns and security boundaries.
