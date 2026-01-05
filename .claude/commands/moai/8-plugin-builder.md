---
name: plugin-builder
description: "Claude Code plugin builder for standalone plugins"
argument-hint: "[request] - Natural language description of desired plugin"
allowed-tools: Task, AskUserQuestion, TodoWrite
model: inherit
---

# Standalone Plugin Builder

Creates standalone Claude Code plugins from scratch without MoAI-ADK dependencies.

## Command Purpose

This command runs on the main thread, sequentially invoking multiple builder agents to create a fully standalone Claude Code plugin.

Target: `$ARGUMENTS` (user's natural language request describing the desired plugin)

---

## Request Analysis

Before proceeding, analyze the user's request in $ARGUMENTS to extract:

- Plugin purpose and domain
- Suggested plugin name (derive from purpose if not specified)
- Required components (skills, agents, commands)
- Target audience and use cases

---

## Insufficient Request Handling

If the request is unclear or missing critical information, use AskUserQuestion to gather required details.

Required Information (must ask if missing):

- Plugin purpose: What problem does this plugin solve?
- Primary use case: How will users interact with this plugin?

Optional Information (ask if helpful for better results):

- Target audience: Personal use, team, or public distribution?
- Preferred components: Skills, agents, commands, or combination?
- Integration needs: Any external services or APIs?

AskUserQuestion Strategy:

- Ask maximum 2 questions at a time (avoid overwhelming user)
- Provide concrete examples in option descriptions
- Use multiSelect when multiple choices are valid
- Progress to next questions only after receiving answers

Example Questions:

Question 1 - If purpose is unclear:
- Header: "Plugin Purpose"
- Question: "What is the main purpose of this plugin?"
- Options: Code quality tools, Documentation generation, Deployment automation, Testing utilities

Question 2 - If components are unclear:
- Header: "Components"
- Question: "Which components should this plugin include?"
- Options with multiSelect: Skills (domain knowledge), Agents (automated workflows), Commands (user actions)

---

## Independence Requirements (Mandatory)

All builder agent invocations must include the following independence requirements:

- No MoAI-ADK internal skill references (moai-platform-*, moai-lang-*, moai-foundation-*, etc.)
- No Context7 MCP references (exclude context7-libraries field)
- No Alfred orchestration references
- No @CLAUDE.md file references
- No /moai:* command references
- Use only standard Claude Code tools (Read, Write, Edit, Grep, Glob, Bash, Task, AskUserQuestion)

---

## Execution Workflow

### PHASE 1: Discovery - Plugin Requirements

Goals:
- Understand plugin purpose and domain
- Identify target audience and use cases
- Determine scope and boundaries

Process:

Step 1: Analyze User Request

Parse the natural language request from $ARGUMENTS to identify:

- Core purpose and problem being solved
- Domain area (e.g., documentation, testing, deployment, code review)
- Implicit requirements and constraints
- Suggested plugin name (convert to kebab-case)

Step 2: Identify Use Cases

Determine specific use cases:
- Who will use this plugin? (personal, team, public)
- What problems does it solve?
- What workflows does it enable?
- What are the success criteria?

Step 3: Scope Assessment

Assess plugin complexity:
- Simple: Single-purpose, minimal components
- Medium: Multiple components, moderate complexity
- Complex: Multiple integration points, advanced features

Deliverables:
- Requirements summary
- Use case documentation
- Complexity assessment
- Initial scope boundaries

---

### PHASE 2: Component Planning - Determine Components Needed

Goals:
- Identify required plugin components
- Map component responsibilities
- Define component interactions

Process:

Step 1: Component Analysis

Determine which components are needed:

Skills (domain knowledge):
- Needed for: Domain-specific knowledge, patterns, best practices
- Output: SKILL.md with modular sections
- Example: moai-lang-python for Python development expertise

Agents (automated workflows):
- Needed for: Multi-step operations, orchestration, delegation
- Output: Agent definition with tool permissions
- Example: code-reviewer for automated PR review

Commands (user actions):
- Needed for: Direct user invocation, slash commands
- Output: Command definition with execution logic
- Example: /review-pr for triggering PR review

Hooks (automation triggers):
- Needed for: Event-driven automation, pre/post actions
- Output: Hook scripts in hooks/hooks.json
- Example: Pre-commit validation

Step 2: Component Mapping

Create component map showing:
- Each component's responsibility
- Inter-component dependencies
- Data flow between components
- Execution sequence

Step 3: Confirm Components with User

Use AskUserQuestion to confirm component selection:
- Present recommended components
- Explain rationale for each
- Allow user to add/remove components
- Get final approval

Deliverables:
- Component list with descriptions
- Component dependency map
- User-approved component configuration

---

### PHASE 3: Detailed Design - Component Specifications

Goals:
- Create detailed specifications for each component
- Define interfaces and contracts
- Specify behavior and edge cases

Process:

Step 1: Skill Specifications

For each skill component:

Define structure:
- Quick Reference: Core capabilities and use cases
- Implementation Guide: Detailed workflows and patterns
- Advanced Features: Extended functionality
- Works Well With: Integration points

Specify content:
- Domain knowledge areas
- Best practices and patterns
- Common pitfalls and solutions
- Technology stack references

Step 2: Agent Specifications

For each agent component:

Define behavior:
- Primary Mission (15 words max)
- Core Capabilities (3-7 bullet points)
- Scope Boundaries (IN/OUT scope)
- Delegation Protocol
- Tool Permissions (least-privilege)

Specify execution:
- Input format and validation
- Processing logic
- Output format
- Error handling

Step 3: Command Specifications

For each command component:

Define interface:
- Name and description
- Argument format and hints
- Allowed tools
- Execution model

Specify workflow:
- Pre-execution context
- Execution phases
- Output format
- Error handling

Step 4: Hook Specifications

For each hook component:

Define triggers:
- Event type (pre/post)
- Execution conditions
- Validation requirements

Specify behavior:
- Input parameters
- Processing logic
- Output/return codes
- Error handling

Deliverables:
- Detailed specifications for each component
- Interface definitions
- Behavior specifications
- Error handling specifications

---

### PHASE 4: Plugin Structure Creation - Directory Structure

Goals:
- Create plugin directory structure
- Set up configuration files
- Initialize component placeholders

Process:

Step 1: Create Directory Structure

Create standard plugin structure:

```
plugins/{plugin-name}/
├── .claude-plugin/
│   └── plugin.json          (required manifest)
├── commands/                 (optional)
│   └── {command-name}.md
├── agents/                   (optional)
│   └── {agent-name}.md
├── skills/                   (optional)
│   └── {skill-name}/
│       ├── SKILL.md
│       └── modules/          (optional)
├── hooks/                    (optional)
│   └── hooks.json
├── .mcp.json                 (optional, for MCP tools)
├── README.md                 (required)
├── LICENSE                   (required)
└── CHANGELOG.md              (required)
```

Step 2: Create plugin.json Manifest

Create .claude-plugin/plugin.json with:

```json
{
  "name": "plugin-name",
  "description": "Brief plugin description (max 100 chars)",
  "version": "1.0.0",
  "author": {
    "name": "Author Name"
  },
  "capabilities": {
    "commands": ["command-name"],
    "agents": ["agent-name"],
    "skills": ["skill-name"]
  },
  "claude": {
    "minVersion": "1.0.0"
  }
}
```

Step 3: Create Metadata Files

Create initial metadata files:
- README.md with usage template
- LICENSE with chosen license
- CHANGELOG.md with initial version entry

Step 4: Create Component Placeholders

Create placeholder files for each approved component:
- commands/{name}.md with template structure
- agents/{name}.md with template structure
- skills/{name}/SKILL.md with template structure
- hooks/hooks.json with empty hooks array

Deliverables:
- Complete plugin directory structure
- plugin.json manifest file
- Initial metadata files (README, LICENSE, CHANGELOG)
- Component placeholder files

---

### PHASE 5: Component Implementation - Build Each Component

Goals:
- Implement each component per specifications
- Ensure quality and consistency
- Validate independence requirements

Process:

Step 1: Skill Implementation

For each skill:

Use Task to invoke builder-skill agent with:
- Skill name and domain information
- Purpose and knowledge area description
- Independence requirements (no moai- prefix, no Context7, no Alfred)
- Output location: plugins/{plugin-name}/skills/{skill-name}/

Generated structure:
- SKILL.md file (500 lines or less)
- modules/ directory (optional, for progressive disclosure)

Quality checks:
- No references to moai-* skills
- No Context7 MCP references
- No Alfred orchestration references
- Third-person descriptions
- Clear usage instructions

Step 2: Agent Implementation

For each agent:

Use Task to invoke builder-agent agent with:
- Agent role and responsibility description
- Tool permissions (least-privilege)
- Independence requirements
- Output location: plugins/{plugin-name}/agents/

Generated structure:
- {agent-name}.md file with:
  - YAML frontmatter (name, description, tools, model)
  - Primary Mission (15 words max)
  - Core Capabilities
  - Scope Boundaries
  - Delegation Protocol

Quality checks:
- Cannot spawn other sub-agents
- Clear scope boundaries
- Minimal tool permissions
- No AskUserQuestion dependency

Step 3: Command Implementation

For each command:

Use Task to invoke builder-command agent with:
- Command purpose and parameter description
- Argument format and hints
- Independence requirements
- Output location: plugins/{plugin-name}/commands/

Generated structure:
- {command-name}.md file with:
  - YAML frontmatter
  - Pre-execution context
  - Execution workflow
  - Output format

Quality checks:
- Clear argument hints
- Specified allowed tools
- Proper error handling
- User-friendly output

Step 4: Hook Implementation

For each hook:

Create hooks/hooks.json with:
- Hook definitions
- Event triggers
- Command references
- Execution conditions

Deliverables:
- Fully implemented skills
- Fully implemented agents
- Fully implemented commands
- Fully implemented hooks
- Quality validation reports

---

### PHASE 6: Validation - Quality Check

Goals:
- Verify all components meet quality standards
- Ensure independence requirements are satisfied
- Validate plugin functionality

Process:

Step 1: Independence Verification

Inspect all files to verify:

Forbidden Patterns:
- No skill references with moai- prefix
- No context7 related references (context7-libraries field)
- No Alfred references (CLAUDE.md, Alfred orchestration)
- No @CLAUDE.md file references
- No /moai: command references

Required Patterns:
- Use only standard Claude Code tools
- Standalone operation without external dependencies
- Self-contained documentation

Step 2: Component Validation

Validate each component type:

Skills:
- SKILL.md exists and is under 500 lines
- modules/ directory organized with progressive disclosure
- Third-person descriptions used
- Clear trigger terms documented
- Works Well With section present

Agents:
- YAML frontmatter complete
- Primary Mission under 15 words
- Core Capabilities specific (3-7 bullets)
- Scope Boundaries explicit (IN/OUT)
- Tool permissions minimal (least-privilege)
- Delegation Protocol defined

Commands:
- YAML frontmatter complete
- Argument hints clear
- Allowed tools specified
- Error handling defined
- Output format specified

Step 3: Integration Validation

Verify plugin integration:
- plugin.json valid and complete
- All components registered in capabilities
- Component dependencies satisfied
- No circular references

Step 4: Syntax Validation

Validate file formats:
- YAML syntax valid (plugin.json, hooks.json)
- Markdown syntax valid (.md files)
- No broken links or references

Deliverables:
- Independence verification report
- Component validation report
- Integration validation report
- Syntax validation report

---

### PHASE 7: Testing - Verify Functionality

Goals:
- Test plugin installation
- Verify component functionality
- Validate user workflows

Process:

Step 1: Installation Test

Test plugin installation:
- Use /plugin validate to verify plugin structure
- Check that plugin.json is valid
- Verify all components are discoverable
- Test plugin load without errors

Step 2: Component Functionality Test

Test each component type:

Skills:
- Verify skill loads correctly
- Test skill invocation
- Validate skill content quality

Agents:
- Verify agent creation
- Test agent execution
- Validate agent tool permissions
- Test agent delegation patterns

Commands:
- Verify command registration
- Test command invocation
- Validate argument parsing
- Test command output

Hooks:
- Verify hook registration
- Test hook trigger conditions
- Validate hook execution
- Test hook return values

Step 3: User Workflow Test

Test typical user workflows:
- Install plugin
- Use primary feature
- Verify expected output
- Check error handling

Step 4: Edge Case Testing

Test edge cases:
- Missing required parameters
- Invalid input formats
- Component failure scenarios
- Resource constraints

Deliverables:
- Installation test report
- Component test report
- User workflow test report
- Edge case test report

---

### PHASE 8: Documentation - Finalize and Document

Goals:
- Complete documentation for all components
- Create usage examples
- Prepare for distribution

Process:

Step 1: Complete README.md

Create comprehensive README with:

Required sections:
- Plugin name and short description
- Quick start guide
- Installation instructions
- Usage examples
- Configuration options
- Component reference
- Troubleshooting
- Contributing guidelines
- License information

Step 2: Complete CHANGELOG.md

Create detailed changelog:

Format:
```markdown
# Changelog

## [1.0.0] - YYYY-MM-DD

### Added
- Initial plugin release
- Core feature description

### Components
- Skills: skill-name
- Agents: agent-name
- Commands: command-name
```

Step 3: Component Documentation

Ensure each component has complete documentation:

Skills:
- Quick reference complete
- Implementation guide detailed
- Advanced features documented
- Examples provided

Agents:
- Purpose clear
- Capabilities listed
- Usage examples provided
- Integration patterns documented

Commands:
- Usage syntax clear
- Examples provided
- Options documented
- Edge cases covered

Step 4: Distribution Preparation

Prepare for distribution:
- Verify LICENSE file is appropriate
- Tag release version
- Create release notes
- Prepare GitHub/GitLab repository

Deliverables:
- Complete README.md
- Complete CHANGELOG.md
- Complete component documentation
- Release preparation checklist

---

## Final Report

After all 8 phases complete, provide comprehensive report:

Report Content:
- Plugin location (full path)
- List of all components created
- Independence verification results
- Quality validation summary
- Test results summary
- Installation instructions
- Usage examples
- Next steps guidance

Next Step Options:
- Plugin testing: Manual testing in target environment
- GitHub deployment: Create repository and publish
- Local installation: Install via /plugin install local
- Additional components: Add more skills/agents/commands
- Complete: End work

---

## Execution Instructions

To execute this command:

1. Analyze user request from $ARGUMENTS
2. Execute PHASE 1: Discovery - gather requirements
3. Execute PHASE 2: Component Planning - determine components
4. Execute PHASE 3: Detailed Design - specify components
5. Execute PHASE 4: Plugin Structure - create directories
6. Execute PHASE 5: Component Implementation - build components
7. Execute PHASE 6: Validation - quality checks
8. Execute PHASE 7: Testing - verify functionality
9. Execute PHASE 8: Documentation - finalize docs
10. Provide final report

Start execution immediately.
