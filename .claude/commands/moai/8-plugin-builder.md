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

### PHASE 1: Requirements Gathering

Analyze user's request from $ARGUMENTS and gather additional requirements.

Step 1: Analyze User Request

Parse the natural language request to identify:

- Core purpose and problem being solved
- Domain area (e.g., documentation, testing, deployment, code review)
- Implicit requirements and constraints
- Suggested plugin name (convert to kebab-case)

Step 2: Confirm Analysis Results

Use AskUserQuestion to present your analysis and confirm:

- Proposed plugin name
- Identified purpose and domain
- Target audience (personal, team, public)

Step 3: Determine Component Configuration

Based on the analyzed requirements, recommend components and confirm with user:

- Skills: Recommend if domain knowledge is needed
- Agents: Recommend if automated workflows are needed
- Commands: Recommend if user-triggered actions are needed

Step 4: Independence Confirmation

Explain independence requirements to user and obtain confirmation.

---

### PHASE 2: Skill Generation (Conditional)

Execute only if user selected skill inclusion.

Use Task to invoke builder-skill agent:

Prompt Content:
- Plugin name and domain information
- Skill purpose and knowledge area description
- Independence requirements (see section above)
- Output location: plugins/{plugin-name}/skills/{skill-name}/

Generated Artifacts:
- SKILL.md file (500 lines or less)
- modules/ directory (optional)

---

### PHASE 3: Agent Generation (Conditional)

Execute only if user selected agent inclusion.

Use Task to invoke builder-agent agent:

Prompt Content:
- Plugin name and domain information
- Agent role and responsibility description
- Independence requirements (see section above)
- Output location: plugins/{plugin-name}/agents/

Generated Artifacts:
- {agent-name}.md file

---

### PHASE 4: Command Generation (Conditional)

Execute only if user selected command inclusion.

Use Task to invoke builder-command agent:

Prompt Content:
- Plugin name and domain information
- Command purpose and parameter description
- Independence requirements (see section above)
- Output location: plugins/{plugin-name}/commands/

Generated Artifacts:
- {command-name}.md file

---

### PHASE 5: Plugin Integration

Use Task to invoke builder-plugin agent:

Prompt Content:
- Plugin name, domain, purpose
- All component paths generated in PHASE 2-4
- Independence requirements (see section above)
- Output location: plugins/{plugin-name}/

Generated Artifacts:
- .claude-plugin/plugin.json (plugin manifest)
- README.md (usage guide)
- LICENSE (license file)
- CHANGELOG.md (change history)

---

### PHASE 6: Independence Verification

Inspect all files in the generated plugin to verify independence.

Verification Items:
- No skill references with moai- prefix
- No context7 related references
- No Alfred references
- No @CLAUDE.md references
- No /moai: command references

Report verification results to user.

---

## Final Report

Provide the following information after plugin generation is complete:

Report Content:
- Plugin location
- List of generated components
- Independence verification results
- Next steps guidance (testing, GitHub deployment)

Next Step Options:
- Plugin testing: Verify generated plugin functionality
- GitHub deployment: Create repository and deploy
- Additional components: Add more skills/agents/commands
- Complete: End work

---

## Plugin Directory Structure

Standard structure for generated plugins:

plugins/{plugin-name}/
- .claude-plugin/
  - plugin.json (required manifest)
- commands/ (optional)
  - {command-name}.md
- agents/ (optional)
  - {agent-name}.md
- skills/ (optional)
  - {skill-name}/
    - SKILL.md
- README.md
- LICENSE
- CHANGELOG.md

Note: Component directories (commands/, agents/, skills/) must be located at plugin root, not inside .claude-plugin/.

---

## Execution Instructions

To execute this command:

1. Analyze the user's natural language request from $ARGUMENTS.
2. Present analysis results and confirm via AskUserQuestion.
3. Recommend and confirm component configuration.
4. Execute PHASE 2-4 based on selected components.
5. Integrate all components in PHASE 5.
6. Verify independence in PHASE 6.
7. Provide final report.

Start execution immediately.
