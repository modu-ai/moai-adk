---
name: test-lr04-agent
description: Agent with dead hook entry referencing absent tool
tools: Read, Grep
effort: medium
hooks:
  PostToolUse:
    - matcher: Write|Edit
      hooks:
        - type: command
          command: echo "post-write hook"
---

# LR-04 Dead Hook Agent

This agent has a hook that references Write and Edit tools,
but neither Write nor Edit is in the tools list.
This creates a dead hook entry that can never trigger.

## Instructions

Read files and grep for patterns.
