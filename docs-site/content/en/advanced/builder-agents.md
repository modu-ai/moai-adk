---
title: Builder Agents and Harness v4
weight: 40
draft: false
---

Detailed guide to Harness v4 Builder for extending MoAI-ADK with dynamic project-specific teams.

{{< callout type="info" >}}
**One-line summary**: Harness v4 Builder dynamically generates project-specific expert teams from natural language requests. It uses a 4-phase workflow (ANALYZE → PLAN → GENERATE → ACTIVATE) and a manifest-based Runner.
{{< /callout >}}

## What is Harness v4 Builder?

Harness v4 Builder uses `/moai harness <natural-language request>` to **dynamically generate project-specific expert teams**.

### Differences from Previous Versions

| Aspect | Previous (v3/Static) | Current (v4 Builder) |
|--------|-----|-----------|
| Generation method | 3 builder agents (builder-skill, builder-agent, builder-plugin) | Single Harness v4 Builder (dynamic) |
| Workflow | User-defined structure | 4-phase ANALYZE → PLAN → GENERATE → ACTIVATE |
| Execution method | Each builder independent | Manifest-based Runner (optional worktree isolation) |
| Scalability | Limited | Auto-detects project context |

## Harness v4 Builder 4-Phase Workflow

### 1. ANALYZE (Analysis Phase)

Analyzes the current project to identify required expertise.

- Analyzes source code structure
- Detects languages and frameworks used
- Surveys existing agents/skills inventory
- Estimates project scope

### 2. PLAN (Planning Phase)

Defines the team composition and roles needed.

- Determines team size (3-5 members)
- Defines each team member's role profile
- Determines worktree isolation needs
- Designs manifest schema

### 3. GENERATE (Generation Phase)

Creates actual agent definitions and configuration.

- Generates agent files under `.claude/agents/harness/`
- Generates `.moai/harness/manifest.json` (Runner configuration)
- Writes role-specific system prompts
- Defines preload skill list

### 4. ACTIVATE (Activation Phase)

Activates the generated team for immediate use.

- Registers agents and validates them
- Initializes manifest Runner
- Creates optional worktrees and isolation settings
- Activates automatic team delegation rules

## Manifest-Based Runner

Harness v4 uses a **Manifest-based Runner** to operate the generated team.

### manifest.json Structure

```json
{
  "spec_id": "HARNESS-PROJECT-001",
  "name": "My Project Custom Team",
  "version": "1.0.0",
  "created_at": "2026-07-01T10:00:00Z",
  "phases": [
    {
      "name": "plan",
      "teammates": [
        {
          "name": "researcher",
          "model": "haiku",
          "mode": "plan",
          "skills": ["moai-foundation-core"]
        }
      ]
    },
    {
      "name": "run",
      "teammates": [
        {
          "name": "implementer",
          "model": "inherit",
          "mode": "acceptEdits",
          "isolation": "worktree_optional"
        }
      ]
    }
  ],
  "worktree_isolation": "L1_optional"
}
```

### Runner Behavior

1. **Phase entry**: Follows manifest phase sequence
2. **Teammate spawn**: Dynamically creates teammates for each phase
3. **Isolation application**: Applies conditional worktree isolation
4. **Result aggregation**: Integrates each teammate's results

## Harness Lifecycle Commands

Generated harnesses are managed with `/harness:<name>` commands.

### Available Commands

```bash
# List all generated harnesses
/harness list

# Check specific harness status
/harness:my-project-team status

# Edit harness configuration
/harness:my-project-team edit

# Delete a harness
/harness:my-project-team remove

# Create new harness with Harness v4 Builder
/moai harness <natural-language request>
```

## Creating Harnesses with Natural Language

### Basic Usage

```bash
> Create an expert team for our backend project.
> We need specialists for API design, DB schema, and testing.
```

### Builder's Workflow

1. ANALYZE: Analyzes project structure (Go, PostgreSQL, REST API)
2. PLAN: Decides on 3-person team (API Designer, DB Specialist, Test Engineer)
3. GENERATE: Creates agent definitions and manifest.json
4. ACTIVATE: Registers team and enables `/harness:backend-team` command

### Generated Output Location

- Agent definitions: `.claude/agents/harness/api-designer.md`, `db-specialist.md`, ...
- Manifest: `.moai/harness/manifest.json`
- Optional worktrees: `~/.moai/worktrees/<project>/` (user opt-in)

## Worktree Isolation (Optional)

Harness v4 supports optional worktree isolation.

### L1 Isolation (Optional)

Claude Code runtime creates L1 worktrees for each agent.

- **When to use**: When parallel team members edit the same files
- **Isolation scope**: Each team member's file writes occur in independent worktree
- **Cost**: Additional memory + reduced parallelism benefit

### Disable Isolation

Set `"worktree_isolation": "none"` in manifest to skip L1 isolation.

## Related Documentation

- [Advanced Harness v4 Guide](/advanced/builder-agents) - 4-phase details and manifest schema
- [Agent Guide](/advanced/agent-guide) - 8 core agents catalog
- [Dynamic Workflows](/advanced/ultracode-workflows) - `/effort ultracode` parallel execution

{{< callout type="info" >}}
**Tip**: Once you generate a custom harness, it's automatically reused in all subsequent work. You can access it anytime with `/harness:team-name`.
{{< /callout >}}
