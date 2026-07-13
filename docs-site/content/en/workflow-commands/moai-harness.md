---
title: /moai harness
weight: 55
draft: false
---

Creates a project-specific dynamic expert team (harness) and manages the harness learning lifecycle.

{{< callout type="info" >}}
**Slash command**: In Claude Code, type `/moai:harness <natural-language request>` to run this command directly.
{{< /callout >}}

## Overview

`/moai:harness` runs MoAI-ADK's **Harness v4 Builder** to automatically generate a dynamic expert team tailored to your project's requirements.

This command lets you experience v3's third pillar, the **Agentic Harness**, first-hand — a recursive structure in which a harness builds a harness. When your project has a domain the general-purpose agent catalog cannot cover (e.g. a specific DB migration procedure, or in-house API conventions), you can scaffold an expert team for that domain from a single natural-language sentence. The generated harness connects to the **recursive self-learning** subsystem — as usage observations accumulate, the harness produces its own improvement proposals, and the guidance evolves through a user-approval gate.

### What Is the Harness v4 Builder?

The Harness v4 Builder composes teams through a Socratic-interview-based 4-phase workflow (ANALYZE → PLAN → GENERATE → ACTIVATE).

| Phase | Description |
|------|------|
| ANALYZE | Analyzes the project structure, languages used, and existing agent inventory |
| PLAN | Decides the required team size (3-5 members), each member's role, and whether to isolate via worktree |
| GENERATE | Generates `.claude/agents/harness/` agent files and `.moai/harness/manifest.json` |
| ACTIVATE | Registers the team and activates the `/harness:<name>` command |

## How to Use

### Step 1: Request a team in natural language

```bash
> /moai:harness <natural-language request>
```

**Example:**
```
Build an expert team for our Go backend project.
We need a team covering DB migrations, REST API endpoints, and unit tests.
```

### Step 2: The Builder handles it automatically

The Builder runs the 4 phases automatically:

1. **ANALYZE**: detects the Go, PostgreSQL, and REST API tech stack
2. **PLAN**: decides on a 3-member team of DB Engineer, API Developer, Test Engineer
3. **GENERATE**:
   - `.claude/agents/harness/db-engineer.md`
   - `.claude/agents/harness/api-developer.md`
   - `.claude/agents/harness/test-engineer.md`
   - generates `.moai/harness/manifest.json`
4. **ACTIVATE**: registers the `/harness:backend-team` command

### Step 3: Use the generated team

After creation, the team is used automatically in all work:

```bash
/moai run SPEC-BACKEND-001
/moai run --team SPEC-BACKEND-001    # force team mode
```

MoAI analyzes the SPEC's complexity and automatically delegates to team members in the manifest's phase order.

## Harness Management Commands

### harness list

List all created harnesses:

```bash
/harness list
```

### harness:<name> status

Detailed information for a specific harness:

```bash
/harness:backend-team status
```

Output information:
- Team member list and roles
- Models in use (inherit, haiku, sonnet, opus)
- Optional worktree isolation setting
- Manifest version and creation date

### harness:<name> edit

Edit the manifest.json and agent definitions:

```bash
/harness:backend-team edit
```

Editable items:
- Add/remove team members
- Skill preload list
- Worktree isolation policy
- Per-role prompts

### harness:<name> remove

Delete a harness and its associated files:

```bash
/harness:backend-team remove
```

Deleted items:
- `.claude/agents/harness/` agent definitions
- `.moai/harness/manifest.json`
- The registered `/harness:<name>` command
- The worktree isolation policy

## The Harness Learning Lifecycle — Recursive Self-Learning

A harness is not a static artifact that ends at creation. The `/moai harness` subcommands manage the lifecycle of the **learning subsystem**.

| Command | Description |
|--------|------|
| `moai harness status` | Check learning status (observation count, patterns, proposals) |
| `moai harness apply` | Apply proposals (must pass the user-approval gate) |
| `moai harness rollback` | Roll back the most recent apply |
| `moai harness disable` | Disable learning |
| `moai harness list` (v4) | List all learned rules |
| `moai harness edit` (v4) | Edit rules directly |
| `moai harness remove` (v4) | Delete rules |
| `moai harness doctor` (v4) | Diagnose the learning system |

**The 4-tier learning ladder** — as observations accumulate, the learning stage climbs:

| Tier | Observations | Behavior |
|------|---------|------|
| TierObservation | ≥1 | Simple recording |
| TierHeuristic | ≥3 | Pattern recognition |
| TierRule | ≥5 | Rule formation |
| TierAutoUpdate | ≥10 | Automatic update (user approval required) |

**Artifacts**: the `.moai/harness/` directory (usage-log.jsonl, learned-rules.yaml)

{{< callout type="warning" >}}
Automatic evolution is only ever applied under the **user-approval gate**. The evaluator and approval authority sit outside the evolution loop, and you can restore at any time with `moai harness rollback`.
{{< /callout >}}

## Manifest Structure

Harness v4 defines team composition via **manifest.json**.

### manifest.json Example

```json
{
  "spec_id": "HARNESS-BACKEND-001",
  "name": "Backend Development Team",
  "version": "1.0.0",
  "created_at": "2026-07-01T10:00:00Z",
  "worktree_isolation": "L1_optional",
  
  "phases": [
    {
      "name": "plan",
      "teammates": [
        {
          "name": "architect",
          "role": "API architecture specialist",
          "model": "inherit",
          "skills": ["moai-foundation-core"]
        }
      ]
    },
    {
      "name": "run",
      "teammates": [
        {
          "name": "db-engineer",
          "role": "DB design and migrations",
          "model": "inherit"
        },
        {
          "name": "api-developer",
          "role": "REST API endpoints",
          "model": "inherit"
        },
        {
          "name": "test-engineer",
          "role": "Unit tests",
          "model": "haiku"
        }
      ]
    }
  ]
}
```

### Phase Fields

| Field | Description |
|------|------|
| `name` | Phase name (`plan`, `run`, `sync`) |
| `teammates` | Array of team members participating in this phase |

### Teammate Fields

| Field | Default | Description |
|------|--------|------|
| `name` | required | Unique identifier for the team member |
| `role` | required | Description of the member's role |
| `model` | `inherit` | Model choice (`inherit`, `haiku`, `sonnet`, `opus`) |
| `skills` | `[]` | List of skills to preload |

Being able to assign a different model per team member (the `model` field) is an extension of the tokenomics design — there is no reason to use the same model for a reasoning-heavy role like architecture decisions and a lightweight role like repetitive test writing.

## Worktree Isolation

Harness v4 supports optional worktree isolation.

### L1_optional (default)

```json
"worktree_isolation": "L1_optional"
```

Claude Code automatically creates an L1 worktree when it detects conflicts between parallel team members.

- **Optional**: isolation applies only on conflict
- **Automatic**: the runtime creates it automatically after conflict detection
- **Cost**: memory increases with worktree isolation

### none

```json
"worktree_isolation": "none"
```

All team members work in the project root (minimum memory use).

## The Team Delegation Workflow

Once a harness is active, MoAI uses that team automatically.

### Team Delegation During SPEC Execution

```bash
> /moai run SPEC-BACKEND-001
```

**MoAI's automatic judgment:**
1. Estimate SPEC complexity (file count, lines of code)
2. Select the appropriate harness
3. Delegate to team members sequentially/in parallel following the manifest's phase order

### Phase-Based Delegation Example

```
PLAN Phase:
  → the architect member handles architecture design

RUN Phase:
  → db-engineer and api-developer delegated in parallel
  → test-engineer delegated sequentially (tests)

SYNC Phase:
  → doc generation and PR authoring (default manager-docs)
```

## The Power of Natural-Language Requests

The Harness v4 Builder discovers requirements through a Socratic interview.

### Effective Request Example

```
Our team is developing a Python FastAPI backend.
We need a team that is strong at API endpoints, data validation, and error handling.
```

The Builder automatically:
- Detects the Python, FastAPI, asyncio tech stack
- Decides on a team size of 3-5
- Sets each member's area of specialization
- Preloads the necessary skills

### The Builder Asks About Unclear Requests

```
I need a team.

→ Builder: What are the project's main technologies? (language, framework)
→ Builder: What area should the team focus on? (backend, frontend, everything)
→ Builder: Any particular expertise you need?
```

## Related Documents

- [Harness v4 Builder Guide](/advanced/builder-agents) - Builder 4-phase details
- [Agent Guide](/advanced/agent-guide) - Understanding the 10-agent catalog
- [SPEC-Based Development](/workflow-commands/moai-plan) - SPEC workflow overview

{{< callout type="info" >}}
**Tip**: Once you create a harness, that team is used automatically in all subsequent work. You can reuse it at any time with the `/harness:team-name` command.
{{< /callout >}}
