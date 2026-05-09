---
name: moai-workflow-thinking
description: >
  Sequential Thinking MCP for structured step-by-step analysis via --deepthink flag.
  Separate from UltraThink which is Claude's native extended reasoning mode.
  Use for multi-step analysis or architecture decisions.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob, mcp__sequential-thinking__sequentialthinking
effort: high
user-invocable: false
metadata:
  version: "2.0.0"
  category: "workflow"
  status: "active"
  modularized: "false"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000

# MoAI Extension: Triggers
triggers:
  keywords: ["sequential thinking", "deepthink", "deep analysis", "complex problem", "architecture decision", "technology selection", "trade-off", "breaking change"]
  phases:
    - plan
  agents:
    - manager-strategy
    - manager-spec
---

# Sequential Thinking MCP (--deepthink)

Structured step-by-step reasoning via `mcp__sequential-thinking__sequentialthinking` MCP tool.

## CRITICAL: Two Distinct Modes

MoAI has TWO independent deep analysis modes. They are NOT the same thing:

| Mode | Trigger | Mechanism | MCP Tool? | GLM Compatible? |
|------|---------|-----------|-----------|-----------------|
| `--deepthink` | Explicit `--deepthink` flag | Sequential Thinking MCP tool | YES — `mcp__sequential-thinking__sequentialthinking` | NO — generates server_tool_use content type |
| `ultrathink` | Keyword or auto-detection | Claude native extended reasoning (high effort) | NO — native to Claude | YES — no special content type |

**Rules:**
- `--deepthink` → ALWAYS invoke Sequential Thinking MCP. NEVER use for native reasoning.
- `ultrathink` → ALWAYS use Claude's native extended reasoning. NEVER invoke Sequential Thinking MCP.
- They can coexist: `ultrathink --deepthink` activates BOTH modes independently.

## Activation Triggers (--deepthink only)

Use Sequential Thinking MCP when `--deepthink` flag is explicitly present:

- Breaking down complex problems into steps
- Planning and design with room for revision
- Architecture decisions affect 3+ files
- Technology selection between multiple options
- Performance vs maintainability trade-offs
- Breaking changes under consideration
- Multiple approaches exist to solve the same problem
- Repetitive errors occur

## Tool Parameters

**Required Parameters:**
- `thought` (string): Current thinking step content
- `nextThoughtNeeded` (boolean): Whether another step is needed
- `thoughtNumber` (integer): Current thought number (starts from 1)
- `totalThoughts` (integer): Estimated total thoughts needed

**Optional Parameters:**
- `isRevision` (boolean): Whether this revises previous thinking
- `revisesThought` (integer): Which thought is being reconsidered
- `branchFromThought` (integer): Branching point for alternatives
- `branchId` (string): Branch identifier
- `needsMoreThoughts` (boolean): If more thoughts needed beyond estimate

## Usage Pattern

**Step 1 - Initial Analysis:**
```
thought: "Analyzing the problem: [describe problem]"
nextThoughtNeeded: true
thoughtNumber: 1
totalThoughts: 5
```

**Step 2 - Decomposition:**
```
thought: "Breaking down: [sub-problems]"
nextThoughtNeeded: true
thoughtNumber: 2
totalThoughts: 5
```

**Step 3 - Revision (if needed):**
```
thought: "Revising thought 2: [correction]"
isRevision: true
revisesThought: 2
thoughtNumber: 3
totalThoughts: 5
nextThoughtNeeded: true
```

**Final Step - Conclusion:**
```
thought: "Conclusion: [final answer]"
thoughtNumber: 5
totalThoughts: 5
nextThoughtNeeded: false
```

## Guidelines

1. Start with reasonable totalThoughts estimate
2. Use isRevision when correcting previous thoughts
3. Maintain thoughtNumber sequence
4. Set nextThoughtNeeded to false only when complete
5. Use branching for exploring alternatives
