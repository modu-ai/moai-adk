---
title: Harness v4 Builder Advanced Guide
weight: 45
draft: false
---

If the [Builder Agents Guide](/en/advanced/builder-agents) was the overview of the Harness v4 Builder, this document is the blueprint — it covers the deliverables of each stage of the 4-phase workflow, the full Manifest schema, and the operating rules of the Runner primitive.

{{< callout type="info" >}}
**One-line summary**: The Harness v4 Builder identifies the expertise you need through a Socratic interview and operates a dynamic team through a manifest-based Runner. Which teammate works with which model is decided by manifest declaration, not by code.
{{< /callout >}}

## 4-Phase Workflow in Detail

### Phase 1: ANALYZE

Analyzes the current project's tech stack and requirements. The goal of this phase is to answer "what expertise is this project missing" with data.

#### What Is Analyzed

- **Project structure**: directory hierarchy, identification of core packages
- **Languages**: detection of Go, Python, TypeScript, Java, etc.
- **Frameworks**: recognition of REST API, gRPC, FastAPI, Django, etc.
- **Existing agents**: catalog of existing definitions in `.claude/agents/`
- **Project scale**: estimation based on file count and lines of code
- **Dependencies**: analysis of `go.mod`, `package.json`, `pyproject.toml`

#### Deliverable

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

Designs the team composition based on the ANALYZE results. Every decision that affects cost — from team size to per-role model assignment — is made in this phase.

#### Planning Decisions

| Item | How decided | Example |
|------|---------|------|
| **Number of specialists** | Project complexity × required expertise (HARD cap 3-7) | 3 specialists |
| **Execution primitive** | Execution shape per specialist | sub-agent, adversarial-fan-out |
| **Isolation** | Likelihood of parallel-specialist conflicts | none \| worktree |
| **Model & effort assignment** | Reasoning complexity per specialist (purpose-based) | content-author: opus/high, translator: sonnet/medium |
| **Companion skills** | Skills needed for specialist expertise | hns-oss-docs-i18n-rules |

Per-specialist model and effort selection is the heart of tokenomics — authoring that needs deep reasoning goes to a higher-tier model with high effort, while repetitive derivation work is assigned to a cheaper model with medium effort. The user approval gate happens at the PLAN→GENERATE boundary via `AskUserQuestion`.

#### Plan Confirmation

The plan is confirmed with the user before generation. No files are ever created without an approval gate.

```
Planned harness composition:
- Name: backend-team
- 3 specialists:
  ① architect (primitive: sub-agent, model: opus, effort: high)
  ② implementer (primitive: sub-agent, model: inherit, effort: high)
  ③ tester (primitive: sub-agent, model: sonnet, effort: medium)
- entry command: /harness:backend-team

Proceed with this composition?
```

### Phase 3: GENERATE

After PLAN approval, the actual agent files and manifest are generated.

#### Generated Artifacts

**1. Agent definition files**

```
.claude/agents/harness/
├── architect.md
├── implementer.md
└── tester.md
```

Each file is defined with a YAML prompt.

```yaml
---
name: architect
description: API architecture design specialist
tools: Read, Write, Edit, Grep, Glob, Bash
model: inherit
---

You are this project's API architecture specialist.
[Detailed per-role instructions]
```

**2. Manifest file**

```
.moai/harness/manifest.json
```

A JSON containing the Phase and Teammate definitions (see § Manifest Schema for the schema).

#### Generation Verification

Right after generation, you can directly verify file existence and definition correctness.

```bash
ls .claude/agents/harness/
# Verify architect.md, implementer.md, tester.md

ls .moai/harness/
# Verify manifest.json

grep -c "\"name\": \"architect\"" .moai/harness/manifest.json
# Verify the phase definitions are correct
```

### Phase 4: ACTIVATE

Registers the generated harness and makes it immediately usable.

#### Activation Steps

1. **Agent validation**: syntax check on each agent file
2. **Manifest validation**: JSON schema and field validation
3. **Command registration**: the `/harness:backend-team` entry command is activated
4. **Runner initialization**: the manifest-based Runner is prepared to start
5. **Worktree creation** (optional): specialist isolation activation conditions configured

#### Activation Check

```bash
moai harness list
# Shows backend-team (name + domain + entry command)

moai harness doctor
# Reference-integrity smoke gate (validates specialist, skill, and workflow references)
```

## Manifest Schema

### Top-Level Fields

| Field | Type | Required | Description |
|------|------|------|------|
| `name` | string | Yes | Harness name (used in the entry command) |
| `domain` | string | Yes | Harness domain description |
| `patterns` | array | Yes | Execution patterns (`Pipeline`, `Fan-out/Fan-in`, `Producer-Reviewer`, etc.) |
| `specialists` | array | Yes | Array of Specialist objects (3-7 HARD cap) |
| `sprint_contract` | object | Yes | Quality dimensions, thresholds, and must_pass gates |
| `companion_skills` | array | — | List of harness-specific companion skills |
| `entry_command` | string | Yes | `/harness:<name>` entry command |
| `runner_workflow` | string | Yes | Runner workflow script file |
| `schedule` | object | — | (Optional) recurring-execution schedule — `mode: discovery-only`, etc. |

### Specialist Object

```json
{
  "role": "content-author",
  "description": "canonical-locale source authoring",
  "agent_file": ".claude/agents/harness/hns-oss-docs-content-author-specialist.md",
  "primitive": "sub-agent",
  "isolation": "none",
  "effort": "high",
  "model": "opus"
}
```

| Field | Description |
|------|------|
| `role` | Specialist role (hyphenated/English) |
| `description` | Role description (free text) |
| `agent_file` | Specialist agent file path (`.claude/agents/harness/`) |
| `primitive` | Execution primitive (`sub-agent`, `adversarial-fan-out`, etc.) |
| `isolation` | Isolation level (`none`, `worktree`) |
| `effort` | Reasoning intensity (`low`, `medium`, `high`, `xhigh`) — purpose-based |
| `model` | Model tier (`opus`, `sonnet`, `haiku`, `inherit`) — purpose-based |

### Sprint Contract

```json
{
  "dimensions": ["locale-parity", "build-clean", "style-compliance", "content-fidelity"],
  "thresholds": { "locale-parity": 1.0, "build-clean": 1.0, "style-compliance": 0.95 },
  "must_pass": ["locale-parity", "build-clean"]
}
```

`dimensions` are the scoring dimensions, `thresholds` are the per-dimension pass thresholds, and `must_pass` defines the gates that must pass.

## Runner Primitive

The manifest-based Runner executes the generated team.

### Runner Lifecycle

```
Team Spawn
  ↓
[Phase 1: plan]
  → Create and delegate to Teammate(architect)
  → Collect results
  ↓
[Phase 2: run]
  → Create Teammate(db-engineer) in parallel
  → Create Teammate(api-developer) in parallel
  → Create Teammate(test-engineer) sequentially
  → Collect and consolidate results
  ↓
[Phase 3: sync]
  → Run the default manager-docs
  ↓
Team Teardown
```

### Runner Configuration

The Runner's behavior is controlled by manifest fields.

| Setting | Meaning |
|------|------|
| `isolation: "worktree"` | Apply worktree isolation to the specialist |
| `isolation: "none"` | Isolation disabled |
| `model: "inherit"` | Inherit the parent session's model |
| `model: "sonnet"` | Low-cost tier for derivation/repetitive work |
| `effort: "high"` \| `"medium"` | Per-specialist reasoning intensity (purpose-based) |
| `companion_skills: ["..."]` | Harness-specific companion skills |

## Worktree Isolation Rules

### L1_optional Behavior

```
On Runner creation:
├── Teammate 1: main project root
├── Teammate 2: main project root
└── On conflict detection
    ├── Teammate 2 → switches to an L1 worktree
    └── Teammate 1 stays on main (or Teammate 1 switches too)

Result:
└── File conflict avoided ✓
```

### Isolation Conditions

Isolation activates when any of the following is true.

1. **Parallel edits to the same file**: two teammates modify the same file simultaneously
2. **Recursive directory writes**: teammates create multiple files in the same directory
3. **Dependency contention**: teammate A's output is teammate B's input (ordering matters)

### When Choosing No Isolation (none)

```
All teammates work in the main project
Pros: minimal memory, fast parallelism
Cons: possibility of conflicts
```

## Related Documents

- [Harness v4 Builder Usage Guide](/en/workflow-commands/moai-harness) - command reference
- [Agent Guide](/en/advanced/agent-guide) - agent definition format
- [SPEC-Based Development](/en/workflow-commands/moai-plan) - Harness and SPEC integration

{{< callout type="info" >}}
**Tip**: After creation, the manifest can be edited anytime — check the edit path with `moai harness edit <name>`. Adding specialists, changing skills, and adjusting the isolation policy are all possible.
{{< /callout >}}
