---
name: moai-cc-commands
description: "Designing Slash Commands for Claude Code. Create and optimize slash commands with proper argument parsing, tool permissions, and agent orchestration. Use when building workflow entry points, automation commands, or user-facing shortcuts."
allowed-tools: "Read, Write, Edit, Glob, Bash"
---

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | Ops |
| Auto-load | When designing slash commands |

## What It Does

Slash command 설계 및 최적화 방법을 제공합니다. Argument parsing, tool permissions, agent orchestration을 포함한 command 구조를 다룹니다.

## When to Use

- 새로운 slash command를 생성할 때
- 기존 command를 최적화하거나 리팩토링할 때
- Workflow entry point를 설계할 때
- User-facing automation을 구축할 때


# Designing Slash Commands

Slash commands are user-facing entry points that orchestrate sub-agents, manage approvals, and coordinate multi-step workflows. They follow the Plan → Execute → Sync cadence.

## Command File Structure

**Location**: `.claude/commands/`

```yaml
---
name: command-name
description: Brief description of what the command does
argument-hint: "[param1] [param2] [optional-param]"
tools: Read, Write, Task, Bash(git:*)
model: sonnet
---

# Command Title

Brief description of functionality.

## Usage

- `/command-name param1 param2` — Basic usage
- Parameter descriptions
- Expected behavior

## Agent Orchestration

1. Call specific agent for task
2. Handle results
3. Provide user feedback
```

## Command Design Patterns

### Pattern 1: Planning Command
```yaml
---
name: /alfred:1-plan
description: Write SPEC requirements in EARS syntax
argument-hint: "[title]"
tools: Read, Write, Task
model: sonnet
---

# SPEC Planning Command

Initiates SPEC authoring via spec-builder sub-agent.

## Usage

`/alfred:1-plan "User authentication system"`

## Agent Orchestration

1. Invoke spec-builder agent
2. Gather requirements via EARS patterns
3. Create SPEC file in `.moai/specs/`
4. Suggest next step: `/alfred:2-run`
```

### Pattern 2: Code Review Command
```yaml
---
name: /review-code
description: Trigger automated code review with quality analysis
argument-hint: "[file-pattern] [--strict]"
tools: Read, Glob, Grep, Task
model: sonnet
---

# Code Review Command

Analyzes code quality against TRUST 5 principles.

## Usage

- `/review-code src/**/*.ts` — Review TypeScript files
- `/review-code . --strict` — Strict mode (fail on warnings)

## Agent Orchestration

1. Scan files matching pattern
2. Invoke code-reviewer agent
3. Generate report with findings
4. Suggest fixes with severity levels
```

### Pattern 3: Deployment Command
```yaml
---
name: /deploy
description: Deploy application with safety gates
argument-hint: "[env] [--force]"
tools: Read, Write, Task, Bash(git:*)
model: haiku
---

# Deployment Command

Orchestrates multi-step deployment with approval gates.

## Usage

- `/deploy staging` — Deploy to staging
- `/deploy production --force` — Force production deploy

## Agent Orchestration

1. Validate deployment readiness
2. Run pre-deployment checks
3. Ask for user approval
4. Execute deployment
5. Monitor post-deployment
```

## Argument Parsing Pattern

```bash
# Inside command execution
$ARGUMENTS              # Entire argument string
$1, $2, $3            # Individual arguments
$@                    # All arguments as array

# Example: /my-command arg1 arg2 --flag
# $ARGUMENTS = "arg1 arg2 --flag"
# $1 = "arg1"
# $2 = "arg2"
# $3 = "--flag"
```

## High-Freedom: Orchestration Strategies

### Sequential Execution
```
Command: /plan-and-implement
├─ Phase 1: spec-builder (SPEC creation)
└─ Phase 2: code-builder (TDD implementation)
```

### Parallel Execution
```
Command: /analyze-project
├─ Agent 1: security-auditor (vulnerability scan)
├─ Agent 2: performance-analyzer (bottleneck detection)
└─ Agent 3: architecture-reviewer (design review)
```

### Conditional Branching
```
Command: /fix-errors
├─ Check: Are tests failing?
│  ├─ YES → debug-helper agent
│  └─ NO → suggest next steps
```

## Medium-Freedom: Command Templates

### Status Check Command
```yaml
name: /status
description: Show project status summary
argument-hint: "[--verbose]"
tools: Read, Bash(git:*)
model: haiku
---

# Status Check

Displays project health: SPEC status, test results, Git state.

## Checks

1. SPEC completeness
2. Test coverage
3. Git branch status
4. Recent commits
5. TODO items
```

### Bulk Operation Command
```yaml
name: /migrate
description: Run migration across multiple files
argument-hint: "[from] [to] [--preview]"
tools: Read, Glob, Bash
model: sonnet
---

# Migration Command

Migrates code patterns across project.

## Usage

- `/migrate v1-api v2-api --preview` — Show changes without applying
- `/migrate v1-api v2-api` — Apply migration
```

## Low-Freedom: Safety & Approval Patterns

### Approval Gate Pattern
```bash
# Inside command execution

echo "🔴 This action is destructive. Review carefully:"
echo "  • Will delete 10 files"
echo "  • Cannot be undone"
echo ""
read -p "Type 'yes' to confirm: " confirm

if [[ "$confirm" != "yes" ]]; then
  echo "❌ Cancelled"
  exit 1
fi

# Execute dangerous operation
```

### Dry-Run Pattern
```bash
if [[ "${3:-}" == "--preview" ]]; then
  echo "🔍 Preview mode (no changes)"
  # Show what would happen
  exit 0
fi

# Execute actual changes
```

## Command Registry

```bash
# List all commands
/commands

# View command details
/commands view /deploy

# Create new command
/commands create

# Edit command
/commands edit /deploy

# Delete command
/commands delete /deploy
```

## Command Validation Checklist

- [ ] `name` is kebab-case (e.g., `/review-code`)
- [ ] `description` clearly explains purpose
- [ ] `argument-hint` shows expected parameters
- [ ] `tools` list is minimal and justified
- [ ] `model` is `haiku` or `sonnet`
- [ ] Agent orchestration is clearly defined
- [ ] Arguments are properly parsed
- [ ] Safety gates are in place for dangerous operations
- [ ] Feedback to user is clear and actionable

## Best Practices

✅ **DO**:
- Design commands around workflows, not tools
- Use agents for complex logic
- Include preview/dry-run modes for risky operations
- Provide clear feedback at each step
- Link to next command in suggestions

❌ **DON'T**:
- Make commands do too much (limit to 1 coherent workflow)
- Require multiple parameters without defaults
- Skip approval gates for destructive operations
- Leave users guessing what happened

---

**Reference**: Claude Code Slash Commands documentation
**Version**: 1.0.0
