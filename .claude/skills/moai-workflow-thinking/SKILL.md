---
name: moai-workflow-thinking
description: >
  Sequential Thinking MCP and UltraThink mode for deep analysis, complex
  problem decomposition, and structured reasoning workflows.
  Use when performing multi-step analysis, architecture decisions, technology selection
  trade-offs, breaking change assessment, or when --ultrathink flag is specified.
  Do NOT use for simple decisions or straightforward implementation tasks.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__sequential-thinking__sequentialthinking
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  modularized: "false"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level_1_tokens: 100
  level_2_tokens: 3000

# MoAI Extension: Triggers
triggers:
  keywords:
    - sequential thinking
    - ultrathink
    - deep analysis
    - complex problem
    - architecture decision
    - technology selection
    - trade-off
    - breaking change
  phases:
    - plan
  agents:
    - manager-strategy
    - manager-spec
---

# Sequential Thinking & UltraThink

Structured reasoning system for complex problem analysis and decision-making.

## Activation Triggers

Use Sequential Thinking MCP when:

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

## UltraThink Mode

Enhanced analysis mode activated by `--ultrathink` flag.

**Activation:**
```
"Implement authentication system --ultrathink"
"Refactor the API layer --ultrathink"
```

**Process:**
1. Request Analysis: Identify core task, detect keywords, recognize complexity
2. Sequential Thinking: Begin structured reasoning
3. Execution Planning: Map subtasks to agents, identify parallel opportunities
4. Execution: Launch agents, integrate results

**UltraThink Parameters:**

Initial Analysis:
```
thought: "Analyzing user request: [content]"
nextThoughtNeeded: true
thoughtNumber: 1
totalThoughts: [estimate]
```

Subtask Decomposition:
```
thought: "Breaking down: 1) [task1] 2) [task2] 3) [task3]"
nextThoughtNeeded: true
thoughtNumber: 2
```

Agent Mapping:
```
thought: "Mapping: [task1] → expert-backend, [task2] → expert-frontend"
nextThoughtNeeded: true
thoughtNumber: 3
```

Execution Strategy:
```
thought: "Strategy: [tasks1,2] parallel, [task3] depends on [task1]"
nextThoughtNeeded: true
thoughtNumber: 4
```

Final Plan:
```
thought: "Plan: Launch [agents] in parallel, then [agent]"
nextThoughtNeeded: false
```

## When to Use

**UltraThink is ideal for:**
- Complex multi-domain tasks (backend + frontend + testing)
- Architecture decisions affecting multiple files
- Performance optimization requiring analysis
- Security review needs
- Refactoring with behavior preservation

**Benefits:**
- Structured decomposition of complex problems
- Explicit agent-task mapping with justification
- Identification of parallel execution opportunities
- Context maintenance throughout reasoning
- Revision capability when approaches need adjustment

## Guidelines

1. Start with reasonable totalThoughts estimate
2. Use isRevision when correcting previous thoughts
3. Maintain thoughtNumber sequence
4. Set nextThoughtNeeded to false only when complete
5. Use branching for exploring alternatives
