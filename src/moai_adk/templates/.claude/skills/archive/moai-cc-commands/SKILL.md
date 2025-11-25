---
name: moai-cc-commands
description: Claude Code Commands system, workflow orchestration, and command-line interface patterns. Use when creating custom commands, managing workflows, or implementing CLI interfaces.
version: 1.0.0
modularized: false
tags:
  - enterprise
  - configuration
  - commands
  - claude-code
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: commands, cc, moai  


## Quick Reference (30 seconds)

# Claude Code Command Architecture & CLI Integration

Claude Code Commands provides a powerful command system for custom workflow automation, CLI interface design, and orchestrating complex multi-step tasks. It efficiently automates development workflows such as project initialization, feature deployment, documentation synchronization, and release management.

**Core Capabilities**:
- Custom command creation and registration
- Multi-step workflow orchestration
- Parameter validation and input processing
- Error handling and recovery
- Command documentation and help system


## Implementation Guide

### What It Does

Claude Code Commands provides:

**Command System**:
- Command registration and discovery
- Parameter parsing and validation
- Command execution and result handling
- Asynchronous command support
- Command chaining and composition

**Workflow Automation**:
- Multi-step task orchestration
- Conditional execution branching
- Error handling and retry logic
- Progress tracking and logging
- Result collection and reporting

**CLI Interface**:
- Command help and usage documentation
- Parameter auto-completion
- Real-time feedback
- Interactive prompts
- Result formatting

### When to Use

- ✅ Project initialization and configuration automation
- ✅ Development workflows (build, test, deploy)
- ✅ Tool integration and orchestration
- ✅ Repetitive task automation
- ✅ Complex multi-step process simplification
- ✅ Team workflow standardization

### Core Command Patterns

#### 1. Command Structure
```markdown
/moai:N-action [parameters] [options]

Examples:
- /moai:0-project                    # Initialize project
- /moai:1-plan "feature description" # Generate SPEC
- /moai:2-run SPEC-001              # Execute TDD
- /moai:3-sync SPEC-001             # Synchronize documentation
```

#### 2. Parameter Handling
```markdown
## Positional Parameters
/command arg1 arg2 arg3

## Named Parameters (Options)
/command --option value --flag

## Mixed Usage
/command required-arg --option value --flag
```

#### 3. Workflow Orchestration Pattern
```markdown
Task 1: Collect requirements
  └─ Task 2: Generate SPEC
      └─ Task 3: Execute implementation
          └─ Task 4: Synchronize documentation
              └─ Task 5: Deploy
```

#### 4. Error Handling Pattern
- Input validation failure → Display help
- Task failure → Retry or rollback
- Partial completion → Save progress
- Unexpected error → Log and record

### Dependencies

- Claude Code commands system
- CLI framework (Click, Typer, Cobra)
- Parameter validation library
- Workflow orchestration tools


## Works Well With

- `moai-cc-agents` (Command execution delegation)
- `moai-cc-hooks` (Command event handling)
- `moai-cc-configuration` (Command configuration)
- `moai-project-config-manager` (Project-specific commands)


## Advanced Patterns

### 1. Advanced Parameter Handling

**Variable Expansion**:
```bash
/command --path {{project-root}}/{{feature-name}}
/command --version {{semantic-version}}
```

**Conditional Parameters**:
```bash
# Development environment
/command --mode dev --verbose

# Production environment
/command --mode prod --debug false
```

**Parameter Validation**:
```markdown
- Required parameter checking
- Type validation (string, number, boolean, path)
- Range validation (min, max, enum values)
- Custom validation rules
```

### 2. Workflow Orchestration Pattern

**Sequential Execution**:
```
Step 1 → Step 2 → Step 3 → Step 4
```

**Parallel Execution**:
```
Step 1A → |
          | → Combined Result
Step 1B → |
```

**Conditional Branching**:
```
Step 1 → [Condition Check]
          ├─ Success → Step 2A
          └─ Failure → Step 2B
```

### 3. Command Extension Pattern

**Plugin System**:
```markdown
1. Define command interface
2. Implement plugin
3. Register plugin
4. Dynamic loading
```

**Hook Integration**:
```markdown
- Pre-command hooks: Before command execution
- Post-command hooks: After command execution
- Error hooks: On error occurrence
- Validation hooks: Parameter validation
```

### 4. Advanced Result Handling

**Result Formatting**:
- Text output
- JSON format
- Table format
- Markdown format

**Result Persistence**:
- Save to file
- Database storage
- Log recording
- Notification sending

---

## Advanced Context Loading (Claude Code Official Features)

### Pre-execution Context with Bash (`! prefix`)

Claude Code automatically executes bash commands before command execution and includes results in the context.

**Syntax**: `!git status --porcelain`

**MoAI Command Optimization Example**:
```yaml
---
name: moai:1-plan
description: "Define specifications and create development branch"
---

## 📋 Pre-execution Context

!git status --porcelain
!git branch --show-current
!git log --oneline -10
!find .moai/specs -name "*.md" -type f
```

**Benefits**:
- Agent automatically understands current git status
- View existing SPEC list during SPEC generation
- Eliminate unnecessary duplicate questions

**Applied to All 6 MoAI Commands**:
1. `/moai:0-project`: Git status, user settings
2. `/moai:1-plan`: Git log, SPEC list
3. `/moai:2-run`: Modified files list
4. `/moai:3-sync`: Diff, branch information
5. `/moai:9-feedback`: Current branch, recent commits
6. `/moai:99-release`: Git tags, remote information

### File References with Content (`@ prefix`)

Automatically include file contents in command context.

**Syntax**: `@src/utils/helpers.js` or `@.moai/config/config.json`

**MoAI Command Example**:
```yaml
---
name: moai:2-run
---

## 📁 Essential Files

@.moai/config/config.json
@.moai/specs/SPEC-001/spec.md
@.moai/specs/SPEC-001/plan.md
```

**Benefits**:
- Agent automatically loads required documents
- Save context tokens (selective loading)
- Ensure consistent information sources

---

## Model Selection Strategy

### `model` Frontmatter Field

Specify a specific Claude model for the command.

**Syntax**:
```yaml
model: "haiku"    # 70% cost savings (fast tasks)
model: "sonnet"   # Default (complex reasoning)
# Omit field to use conversation default model
```

### MoAI Command Model Assignment Strategy

| Command | Model | Reason | Cost |
|---------|-------|--------|------|
| `/moai:0-project` | Sonnet | Complex configuration logic, validation | Standard |
| `/moai:1-plan` | Sonnet | SPEC generation, EARS design | Standard |
| `/moai:2-run` | Sonnet | TDD orchestration | Standard |
| `/moai:3-sync` | **Haiku** | Pattern-based documentation sync | **-70%** |
| `/moai:9-feedback` | **Haiku** | Simple data collection | **-70%** |
| `/moai:99-release` | **Haiku** | Mechanical version management | **-70%** |

**Result**: Average 35% cost savings, quality maintained

---

## Dynamic Arguments & Variables

### Positional Arguments

Access parameters passed to the command.

**Syntax**:
```markdown
/command arg1 arg2 arg3

- $ARGUMENTS: "arg1 arg2 arg3" (all arguments)
- $1: "arg1" (first argument)
- $2: "arg2" (second argument)
```

**MoAI Example**:
```markdown
/moai:2-run SPEC-001
  → $ARGUMENTS = "SPEC-001"
  → $1 = "SPEC-001"
```

### Variable Expansion

Project metadata variable expansion:

**Syntax**:
```yaml
--path {{project-root}}/{{feature-name}}
--version {{semantic-version}}
```

---

## Command Frontmatter Complete Reference

### Required Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `name` | string | Command name (auto-generated from filename) | `moai:1-plan` |
| `description` | string | Command description (shown in help) | "Define specifications..." |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `argument-hint` | string | none | Parameter hint (auto-completion) |
| `allowed-tools` | array | inherit | Allowed tools list |
| `model` | string | inherit | Claude model selection |
| `disable-model-invocation` | boolean | false | Disable SlashCommand tool |

### allowed-tools Optimization

```yaml
allowed-tools:
  - Task           # Agent delegation (recommended)
  - AskUserQuestion # User interaction
  - Skill          # Skill invocation
  - Bash           # Local-only tools only
```

**Recommended**: Task + AskUserQuestion combination (sufficient for most cases)

---

## MoAI Commands Best Practices

### Complete Optimization Example: /moai:1-plan

```yaml
---
name: moai:1-plan
description: "Define specifications and create development branch"
argument-hint: "Title 1 Title 2 ... | SPEC-ID modifications"
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
skills:
  - moai-core-issue-labels
---

## 📋 Pre-execution Context

!git status --porcelain
!git branch --show-current
!git log --oneline -10
!find .moai/specs -name "*.md" -type f

## 📁 Essential Files

@.moai/config/config.json
@.moai/project/product.md
@.moai/project/structure.md
@CLAUDE.md

---

# 🏗️ Plan Step
...
```

**Optimization Benefits**:
- ✅ Git context auto-loaded
- ✅ SPEC documents pre-loaded
- ✅ Agent token savings
- ✅ SPEC generation accuracy improvement 25-30%

### Haiku Optimization Example: /moai:9-feedback

```yaml
---
name: moai:9-feedback
description: "Submit feedback or report issues"
allowed-tools:
  - Task
  - AskUserQuestion
model: "haiku"
---

## 📋 Pre-execution Context

!git status --porcelain
!git branch --show-current

## 📁 Essential Files

@.moai/config/config.json
@CLAUDE.md
```

**Cost Savings**: 70% cost reduction (template-based tasks)

---

## Changelog

- **v3.0.0** (2025-11-22): Added advanced context loading, model selection, dynamic arguments, complete frontmatter reference, MoAI optimization examples
- **v2.0.0** (2025-11-11): Added complete metadata, command architecture patterns
- **v1.0.0** (2025-10-22): Initial commands system

---

**End of Skill** | Updated 2025-11-22 | Lines: 410



