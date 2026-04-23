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

## CRITICAL: Three Distinct Reasoning Modes

MoAI has THREE independent deep analysis modes. They are NOT the same thing:

| Mode | Trigger | Mechanism | MCP Tool? | GLM Compatible? | Model |
|------|---------|-----------|-----------|-----------------|-------|
| `--deepthink` | Explicit `--deepthink` flag | Sequential Thinking MCP tool | YES — `mcp__sequential-thinking__sequentialthinking` | NO — generates server_tool_use content type | Any |
| `ultrathink` | Keyword or auto-detection | Claude native extended reasoning (high effort) | NO — native to Claude | YES — no special content type | Any |
| Adaptive Thinking | Automatic on Opus 4.7 | Opus 4.7's only supported thinking mode | NO — built-in | YES | Opus 4.7 only |

**Rules:**
- `--deepthink` → ALWAYS invoke Sequential Thinking MCP. NEVER use for native reasoning.
- `ultrathink` → ALWAYS use Claude's native extended reasoning. NEVER invoke Sequential Thinking MCP.
- They can coexist: `ultrathink --deepthink` activates BOTH modes independently.
- Adaptive Thinking → Opus 4.7's built-in reasoning. Let the model adapt depth automatically based on task complexity.
- On Opus 4.7: do not hardcode a fixed reasoning budget. Adaptive Thinking overrides fixed-budget instructions.

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

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I can think through this without sequential thinking, it is simple" | Simple problems often hide second-order effects. Structured thinking forces you to enumerate them. |
| "The thinking steps are just internal, I do not need to record them" | Unrecorded reasoning cannot be reviewed. The chain of thoughts is evidence for the conclusion. |
| "I already know the answer, the thinking framework is overhead" | Confirmation bias skips disconfirming evidence. The framework forces you to consider alternatives. |
| "Branching is overkill for this decision" | Decisions with more than one viable path benefit from explicit branch comparison, even briefly. |
| "I will use UltraThink instead, it is the same thing" | UltraThink is native extended reasoning. Sequential Thinking is MCP-based structured analysis. They serve different purposes and have different API compatibility. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Architecture decision made without documented reasoning chain
- Sequential thinking session ended with nextThoughtNeeded still true
- Thought chain has no revision steps despite encountering contradictions
- Branching not used when two or more viable alternatives were identified
- --deepthink flag confused with UltraThink (wrong tool for the job)

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Thinking chain has a clear totalThoughts estimate that was met or revised
- [ ] Final thought sets nextThoughtNeeded to false with a conclusion
- [ ] At least one revision step exists if contradictions were encountered
- [ ] Branching used when multiple alternatives were identified (show branch IDs)
- [ ] Conclusion references specific thought numbers as supporting evidence

<!-- moai:evolvable-end -->
