---
name: "Designing Slash Commands for Claude Code"
description: "Create and optimize slash commands with proper argument parsing, tool permissions, and agent orchestration. Use when building workflow entry points, automation commands, or user-facing shortcuts."
allowed-tools: "Read, Write, Edit, Glob, Bash"
---

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

## 🤝 Works Well With

**Complementary Skills:**
- **moai-cc-agents** - Commands orchestrate and invoke agents
- **moai-cc-skills** - Commands reference Skills (load via Skill() function)
- **moai-cc-hooks** - Commands trigger Hook events (SessionStart, PreToolUse, etc.)
- **moai-foundation-git** - Commands integrate with GitFlow workflow

**MoAI-ADK Workflows:**
- **`/alfred:0-project`** - Project initialization command
- **`/alfred:1-plan`** - SPEC authoring command (invokes spec-builder agent)
- **`/alfred:2-run`** - TDD implementation command (invokes code-builder pipeline)
- **`/alfred:3-sync`** - Document sync command (invokes doc-syncer + tag-agent)

**Example Integration (MoAI-ADK):**
```bash
# 1. /alfred:1-plan orchestrates spec-builder agent
/alfred:1-plan "JWT authentication system"

# 2. Behind scenes: /alfred:1-plan loads Skill("moai-foundation-specs"),
#    Skill("moai-foundation-ears"), then invokes spec-builder agent

# 3. Result: SPEC file created + @SPEC:ID assigned
```

**Common MoAI Patterns:**
- ✅ `/alfred:0-project` = Bootstrap + load moai-foundation-langs
- ✅ `/alfred:1-plan` = Invoke spec-builder + load moai-foundation-specs/ears
- ✅ `/alfred:2-run` = Execute code-builder pipeline + load moai-essentials-refactor
- ✅ `/alfred:3-sync` = Invoke tag-agent + doc-syncer + GitHub MCP

**General Claude Code Patterns:**
- ✅ Custom `/my-command` for project workflows
- ✅ `$ARGUMENTS` for flexible parameter passing
- ✅ Command orchestration of multiple agents

**See Also:**
- 📖 **Orchestrator Guide:** `Skill("moai-cc-guide")` → SKILL.md
- 📖 **Plan Phase:** `Skill("moai-cc-guide")` → workflows/alfred-1-plan-flow.md
- 📖 **Run Phase:** `Skill("moai-cc-guide")` → workflows/alfred-2-run-flow.md

---

**Reference**: Claude Code Slash Commands documentation
**Version**: 1.0.0
