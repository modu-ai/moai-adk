---
title: /moai harness
weight: 55
draft: false
---

Create and manage dynamic project-specific expert teams with Harness v4 Builder.

{{< callout type="info" >}}
**Slash command**: Type `/moai harness <natural-language request>` in Claude Code to invoke this command directly.
{{< /callout >}}

## Overview

`/moai harness` invokes MoAI-ADK's **Harness v4 Builder** to automatically generate project-specific expert teams.

### What is Harness v4 Builder?

Harness v4 Builder uses a Socratic interview-based 4-phase workflow (ANALYZE → PLAN → GENERATE → ACTIVATE) to compose your team.

| Phase | Description |
|-------|-------------|
| ANALYZE | Analyzes project structure, languages, existing agents inventory |
| PLAN | Decides team size (3-5 members), role definitions, worktree isolation |
| GENERATE | Creates `.claude/agents/harness/` agent files, `.moai/harness/manifest.json` |
| ACTIVATE | Registers team and enables `/harness:<name>` command |

## Usage

### Step 1: Request Team Generation

```bash
> /moai harness <natural-language request>
```

**Example:**
```
Create an expert team for our Go backend project.
We need specialists for DB migration, REST API endpoints, and unit testing.
```

### Step 2: Builder Auto-Processes 4-Phase

Builder executes phases automatically:

1. **ANALYZE**: Detects Go, PostgreSQL, REST API tech stack
2. **PLAN**: Decides 3-person team (DB Engineer, API Developer, Test Engineer)
3. **GENERATE**:
   - `.claude/agents/harness/db-engineer.md`
   - `.claude/agents/harness/api-developer.md`
   - `.claude/agents/harness/test-engineer.md`
   - `.moai/harness/manifest.json`
4. **ACTIVATE**: Registers `/harness:backend-team` command

### Step 3: Use Generated Team

Team is automatically used in all subsequent work:

```bash
/moai run SPEC-BACKEND-001
/moai run --team SPEC-BACKEND-001    # Force team mode
```

MoAI analyzes SPEC complexity and auto-delegates teammates in manifest phase order.

## Harness Management Commands

### harness list

View all generated harnesses:

```bash
/harness list
```

### harness:<name> status

Check specific harness details:

```bash
/harness:backend-team status
```

Output includes:
- Team member list and roles
- Model selection (inherit, haiku, sonnet, opus)
- Optional worktree isolation settings
- Manifest version and creation date

### harness:<name> edit

Modify manifest.json and agent definitions:

```bash
/harness:backend-team edit
```

Editable items:
- Add/remove team members
- Preload skill list
- Worktree isolation policy
- Role-specific prompts

### harness:<name> remove

Delete harness and related files:

```bash
/harness:backend-team remove
```

Deleted items:
- `.claude/agents/harness/` agent definitions
- `.moai/harness/manifest.json`
- Registered `/harness:<name>` command
- Worktree isolation policies

## Manifest Structure

Harness v4 defines teams with **manifest.json**.

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
          "role": "API Architecture Specialist",
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
          "role": "DB Design and Migration",
          "model": "inherit"
        },
        {
          "name": "api-developer",
          "role": "REST API Endpoints",
          "model": "inherit"
        },
        {
          "name": "test-engineer",
          "role": "Unit Testing",
          "model": "haiku"
        }
      ]
    }
  ]
}
```

### Phase Fields

| Field | Description |
|-------|-------------|
| `name` | Phase name (`plan`, `run`, `sync`) |
| `teammates` | Array of team members for this phase |

### Teammate Fields

| Field | Default | Description |
|-------|---------|-------------|
| `name` | Required | Team member unique identifier |
| `role` | Required | Role description |
| `model` | `inherit` | Model choice (`inherit`, `haiku`, `sonnet`, `opus`) |
| `skills` | `[]` | Skills to preload |

## Worktree Isolation

Harness v4 supports optional worktree isolation.

### L1_optional (Default)

```json
"worktree_isolation": "L1_optional"
```

Claude Code automatically creates L1 worktrees when parallel team members conflict.

- **Optional**: Isolation applied only on conflict detection
- **Automatic**: Runtime detects conflicts and creates worktrees
- **Cost**: Increased memory during isolation

### none

```json
"worktree_isolation": "none"
```

All team members work in project root (minimum memory usage).

## Team Delegation Workflow

Once a harness is active, MoAI automatically uses the team.

### Team Delegation on SPEC Execution

```bash
> /moai run SPEC-BACKEND-001
```

**MoAI's automatic decision:**
1. Estimates SPEC complexity (file count, code lines)
2. Selects appropriate harness
3. Delegates teammates in manifest phase order sequentially/parallel

### Phase-Based Delegation Example

```
PLAN Phase:
  → architect teammate designs architecture

RUN Phase:
  → db-engineer and api-developer in parallel
  → test-engineer sequentially for testing

SYNC Phase:
  → Documentation generation and PR (default manager-docs)
```

## Power of Natural Language Requests

Harness v4 Builder understands Socratic interviews.

### Effective Request Examples

```
Our team is developing a Python FastAPI backend.
We need a team that excels at API endpoints, data validation, and error handling.
```

Builder automatically:
- Detects Python, FastAPI, asyncio tech stack
- Decides on 3-5 person team size
- Specializes each teammate's focus areas
- Preloads necessary skills

### Builder Asks on Unclear Requests

```
We need a team.

→ Builder: What's your project's primary tech? (languages, frameworks)
→ Builder: What areas should the team focus on? (backend, frontend, full-stack)
→ Builder: Any special expertise needed?
```

## Related Documentation

- [Harness v4 Builder Guide](/advanced/builder-agents) - 4-phase details
- [Agent Guide](/advanced/agent-guide) - 8 core agents
- [SPEC-based Development](/workflow-commands/moai-plan) - SPEC workflow overview

{{< callout type="info" >}}
**Tip**: Once you create a harness, it's automatically reused in all subsequent work. Use `/harness:team-name` to manage it anytime.
{{< /callout >}}
