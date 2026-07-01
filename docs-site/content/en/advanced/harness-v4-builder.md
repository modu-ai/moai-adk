---
title: Harness v4 Builder Advanced Guide
weight: 45
draft: false
---

Detailed guide to Harness v4 Builder's 4-phase workflow, Manifest schema, and Runner primitives.

{{< callout type="info" >}}
**One-line summary**: Harness v4 Builder uses Socratic interviews to identify required expertise, then operates dynamic teams via manifest-based Runners.
{{< /callout >}}

## 4-Phase Workflow Details

### Phase 1: ANALYZE

Analyzes the current project's technology stack and requirements.

#### Analysis Targets

- **Project structure**: Directory hierarchy, core package identification
- **Languages used**: Go, Python, TypeScript, Java detection
- **Frameworks**: REST API, gRPC, FastAPI, Django recognition
- **Existing agents**: Catalog existing definitions in `.claude/agents/`
- **Project scale**: Estimate from file count and lines of code
- **Dependencies**: Parse `go.mod`, `package.json`, `pyproject.toml`

#### Output

```yaml
analysis_result:
  languages:
    - go (primary)
    - shell (build scripts)
  frameworks:
    - REST API (net/http)
    - PostgreSQL ORM (sqlc)
  scale: "100~300 files, ~50K LOC"
  existing_agents: 0
  expertise_gaps:
    - Database schema design
    - API error handling patterns
    - Test coverage automation
```

### Phase 2: PLAN

Design team composition based on ANALYZE results.

#### Planning Decisions

| Item | Decision Basis | Example |
|------|----------------|---------|
| **Team Size** | Project complexity × required expertise | 3~5 members |
| **Role Profiles** | Anthropic role_profiles (researcher/architect/implementer/tester/designer/reviewer) | architect, implementer, tester |
| **Worktree Isolation** | Parallel teammate conflict potential | L1_optional (selective isolation) |
| **Model Selection** | Reasoning complexity per role | architect: inherit, tester: haiku |
| **Skill Preload** | Role-specific expertise skills | moai-foundation-core, moai-domain-backend |

#### Plan Validation

Confirm with user before generation:

```
Planned team composition:
- Team name: Backend Development Team
- 3 teammates:
  ① architect (model: inherit)
  ② implementer (model: inherit)
  ③ tester (model: haiku)
- Worktree isolation: L1_optional
- Manifest: .moai/harness/manifest.json

Proceed with this configuration?
```

### Phase 3: GENERATE

Generate actual agent files and manifest after PLAN approval.

#### Generation Outputs

**1. Agent Definition Files**

```
.claude/agents/harness/
├── architect.md
├── implementer.md
└── tester.md
```

Each file is defined with YAML frontmatter:

```yaml
---
name: architect
description: API architecture design expert
tools: Read, Write, Edit, Grep, Glob, Bash
model: inherit
---

You are an API architecture expert for this project.
[Role-specific detailed guidance]
```

**2. Manifest File**

```
.moai/harness/manifest.json
```

Contains Phase and Teammate definitions in JSON (see § Manifest Schema for details).

#### Generation Validation

```bash
ls .claude/agents/harness/
# Confirm architect.md, implementer.md, tester.md

ls .moai/harness/
# Confirm manifest.json

grep -c "\"name\": \"architect\"" .moai/harness/manifest.json
# Verify phase definitions are correct
```

### Phase 4: ACTIVATE

Register the generated harness and make it immediately usable.

#### Activation Steps

1. **Agent validation**: Check syntax of each agent file
2. **Manifest validation**: JSON schema and field validation
3. **Command registration**: Activate `/harness:backend-team` command
4. **Runner initialization**: Prepare manifest-based Runner startup
5. **Worktree creation** (optional): Configure L1 isolation conditions

#### Activation Confirmation

```bash
/harness list
# backend-team should be listed

/harness:backend-team status
# Verify 3 teammates, models, state
```

## Manifest Schema

### Top-level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `spec_id` | string | Yes | Format `HARNESS-{DOMAIN}-{NUM}` |
| `name` | string | Yes | Team display name |
| `version` | string | Yes | Semantic versioning `X.Y.Z` |
| `created_at` | string | Yes | ISO 8601 timestamp |
| `worktree_isolation` | enum | Yes | `L1_optional` \| `none` |
| `phases` | array | Yes | Array of Phase objects |

### Phase Object

```json
{
  "name": "run",
  "description": "Implementation phase",
  "teammates": [...]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | `plan` \| `run` \| `sync` |
| `description` | string | Phase objective description |
| `teammates` | array | Array of Teammate objects |

### Teammate Object

```json
{
  "name": "api-developer",
  "role": "REST API endpoint development",
  "model": "inherit",
  "mode": "acceptEdits",
  "skills": ["moai-foundation-core"],
  "isolation": "worktree_optional"
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `name` | Required | Teammate ID (hyphens, no spaces) |
| `role` | Required | Role description (free text) |
| `model` | `inherit` | `inherit`, `haiku`, `sonnet`, `opus` |
| `mode` | `acceptEdits` | Permission mode (`acceptEdits`, `default`, `bypassPermissions`) |
| `skills` | `[]` | Array of preloaded skills (e.g. `["moai-foundation-core"]`) |
| `isolation` | None | `worktree_optional` (conditional worktree isolation) |

### Complete Example

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
      "description": "Architecture design and SPEC authoring",
      "teammates": [
        {
          "name": "architect",
          "role": "API architecture expert",
          "model": "inherit",
          "mode": "acceptEdits",
          "skills": ["moai-foundation-core"]
        }
      ]
    },
    {
      "name": "run",
      "description": "Implementation",
      "teammates": [
        {
          "name": "db-engineer",
          "role": "Database design and migration",
          "model": "inherit",
          "mode": "acceptEdits",
          "isolation": "worktree_optional"
        },
        {
          "name": "api-developer",
          "role": "REST API endpoint implementation",
          "model": "inherit",
          "mode": "acceptEdits",
          "isolation": "worktree_optional"
        },
        {
          "name": "test-engineer",
          "role": "Unit and integration testing",
          "model": "haiku",
          "mode": "acceptEdits"
        }
      ]
    }
  ]
}
```

## Runner Primitive

Manifest-based Runner executes the generated team.

### Runner Lifecycle

```
Team Spawn
  ↓
[Phase 1: plan]
  → Create and delegate Teammate(architect)
  → Collect results
  ↓
[Phase 2: run]
  → Create Teammate(db-engineer) in parallel
  → Create Teammate(api-developer) in parallel
  → Create Teammate(test-engineer) sequentially
  → Collect and integrate results
  ↓
[Phase 3: sync]
  → Execute default manager-docs
  ↓
Team Teardown
```

### Runner Configuration

Runner behavior is controlled by manifest fields:

| Setting | Meaning |
|---------|---------|
| `worktree_isolation: "L1_optional"` | Apply auto isolation on conflict detection |
| `worktree_isolation: "none"` | Isolation disabled |
| `model: "inherit"` | Inherit parent session model |
| `model: "haiku"` | Force Haiku model (cost optimization) |
| `skills: ["..."]` | Preload skills |

## Worktree Isolation Rules

### L1_optional Behavior

```
On Runner creation:
├── Teammate 1: main project root
├── Teammate 2: main project root
└── On conflict detection
    ├── Teammate 2 → L1 worktree
    └── Teammate 1 stays main (or also switches)

Result:
└── File conflicts avoided ✓
```

### Isolation Conditions

Activate isolation if any of these is true:

1. **Same file parallel edit**: Two teammates modifying the same file simultaneously
2. **Recursive directory write**: Teammates creating multiple files in the same directory
3. **Dependency conflict**: Teammate A's output feeds into Teammate B (order matters)

### Non-isolation (none) Choice

```
All teammates work on main project
Advantage: Minimal memory, fast parallelism
Disadvantage: Conflict possibility
```

## Related Documentation

- [Harness v4 Builder Usage Guide](/workflow-commands/moai-harness) - Command reference
- [Agent Guide](/advanced/agent-guide) - Agent definition format
- [SPEC-Based Development](/workflow-commands/moai-plan) - Harness and SPEC integration

{{< callout type="info" >}}
**Tip**: After generation, edit the Manifest anytime using `/harness:team-name edit`. Add teammates, change skills, adjust isolation policy — all possible.
{{< /callout >}}
