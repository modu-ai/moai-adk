---
name: test-lr02-agent
description: Agent with Agent in tools CSV list
tools: Read, Write, Agent, Grep
effort: high
---

# LR-02 Violation Agent

This agent incorrectly includes the Agent tool in its tools list.
Subagents cannot spawn sub-subagents — the Agent tool is only for the orchestrator.

## Instructions

Perform analysis and report results.
