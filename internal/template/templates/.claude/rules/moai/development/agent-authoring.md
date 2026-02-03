# Agent Authoring

Guidelines for creating custom agents in MoAI-ADK.

## Agent Definition Location

Custom agents are defined in `.claude/agents/*.md`

## Required Elements

Every agent definition must include:

- Name: Unique identifier (lowercase with hyphens)
- Description: Clear purpose statement
- Domain: Area of expertise
- Tools: Permitted tool list
- Triggers: Keywords that activate the agent

## Agent Categories

### Manager Agents

Coordinate workflows and multi-step processes:

- manager-spec: SPEC document creation
- manager-ddd: DDD implementation cycle
- manager-docs: Documentation generation

### Expert Agents

Domain-specific implementation:

- expert-backend: API and server development
- expert-frontend: UI and client development
- expert-security: Security analysis

### Builder Agents

Create new MoAI components:

- builder-agent: New agent definitions
- builder-skill: New skill creation
- builder-command: Slash command creation

## Rules

- Write agent definitions in English
- Define expertise domain clearly
- Minimize tool permissions (least privilege)
- Include relevant trigger keywords

## Tool Permissions

Recommended tool sets by category:

Manager agents: Read, Write, Edit, Grep, Glob, Bash, Task, TaskCreate, TaskUpdate

Expert agents: Read, Write, Edit, Grep, Glob, Bash

Builder agents: Read, Write, Edit, Grep, Glob

## Agent Invocation

Invoke agents via Task tool:

- "Use the expert-backend subagent to implement the API"
- Task tool with subagent_type parameter

## MoAI Integration

- Use builder-agent subagent for creation
- Skill("moai-foundation-claude") for patterns
- Follow skill-authoring.md for YAML schema
