---
name: test-orchestrator-agent
description: Orchestrator-class agent with AskUserQuestion in tools (exempt from LR-01)
tools: Read, Write, Edit, Grep, Glob, Bash, Skill, AskUserQuestion, Agent, TodoWrite
effort: xhigh
---

# Orchestrator Agent (LR-01 Exempt)

This agent is an orchestrator-class agent. Because it declares AskUserQuestion
in its tools list, it is exempt from LR-01 enforcement.

## Instructions

Use AskUserQuestion to ask the user for decisions.
You may call AskUserQuestion with a questions parameter.
This usage is permitted because this agent has AskUserQuestion in its tools.
