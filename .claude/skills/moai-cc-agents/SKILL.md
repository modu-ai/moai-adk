---
name: "Creating and Managing Sub-agents in Claude Code"
description: "Design agent personas, define proactive triggers, set tool permissions, structure agent files. Use when building specialized agents for code review, debugging, architecture, or domain-specific tasks."
allowed-tools: "Read, Write, Edit, Glob, Bash"
---

# Creating and Managing Sub-agents

Sub-agents are specialized Claude instances with independent context, custom prompts, and restricted tool access. They handle deep analysis, parallel work, and autonomous tasks.

## Agent File Structure

**Location**: `.claude/agents/`

```yaml
---
name: agent-name
description: Use PROACTIVELY for [specific trigger conditions]
tools: Read, Write, Edit, Glob, Grep, Bash(git:*)
model: sonnet
---

# Agent Name — Specialist Role

Brief description of agent expertise.

## Core Mission

- Primary responsibility
- Scope boundaries
- Success criteria

## Proactive Triggers

- When to activate automatically
- Specific conditions for invocation
- Integration with workflow

## Workflow Steps

1. Input validation
2. Task execution
3. Output verification
4. Handoff to next agent (if applicable)

## Constraints

- What NOT to do
- Delegation rules
- Quality gates
```

## Agent Persona Design Pattern

```markdown
---
name: code-reviewer
description: Use PROACTIVELY for code review requests, PR analysis, or quality checks
tools: Read, Glob, Grep, Bash(git:*)
model: sonnet
---

# Code Reviewer — Quality Expert

Specialized in identifying code quality issues, security risks, and architecture concerns.

## Core Mission

- Review code for SOLID principles
- Identify code smells and anti-patterns
- Verify security best practices
- Suggest improvements with rationale

## Proactive Triggers

- When user mentions "review", "quality", "audit"
- After significant code changes
- Before PR merge

## Workflow Steps

1. **Analyze**: Read changed files, understand context
2. **Evaluate**: Check against TRUST 5 principles
3. **Report**: List findings with severity & fix suggestions
4. **Recommend**: Suggest refactoring or architectural improvements

## Constraints

- No direct edits (only suggestions)
- Focus on maintainability, not style
- Respect existing architecture decisions
```

## High-Freedom: Agent Principles

### Autonomy & Expertise
- Each agent owns 1 specialization (not 3+)
- Define clear "when to activate" triggers
- Agents should make decisions independently

### Tool Access Minimization
- Grant only necessary tools per role
- Restrict Bash to specific commands: `Bash(git:*)`, `Bash(python:*)`
- Never grant `Bash(*)`; always specify pattern

### Handoff & Collaboration
- Agent A completes → hands off to Agent B
- Use clear "next agent" instructions
- No circular dependencies

## Medium-Freedom: Common Agent Patterns

### Pattern 1: Debugger Agent
```yaml
name: debugger
description: Use PROACTIVELY for error diagnosis, test failures, exception analysis
tools: Read, Grep, Glob, Bash(pytest:*), Bash(git:*)
model: sonnet
```

**Mission**: Diagnose errors, provide fix-forward guidance
**Triggers**: `error`, `failed`, `exception`, `debug`
**Output**: Root cause + suggested fixes

### Pattern 2: Architect Agent
```yaml
name: architect
description: Use PROACTIVELY for system design, refactoring, scalability concerns
tools: Read, Glob, Grep, Bash(ls:*), Bash(find:*)
model: sonnet
```

**Mission**: Design systems, propose architectures
**Triggers**: `architecture`, `refactor`, `design`, `scalability`
**Output**: Design document + implementation roadmap

### Pattern 3: Security Agent
```yaml
name: security-auditor
description: Use PROACTIVELY for vulnerability assessment, OWASP checks, secrets detection
tools: Read, Glob, Grep
model: sonnet
```

**Mission**: Find security issues, verify OWASP compliance
**Triggers**: `security`, `audit`, `vulnerability`, `secrets`
**Output**: Risk assessment + remediation steps

## Low-Freedom: Tool Permission Patterns

### Principle of Least Privilege

```yaml
# ❌ Too permissive
tools: Read, Write, Edit, Bash(*)

# ✅ Appropriate
tools: Read, Glob, Grep  # Read-only analysis

# ✅ Appropriate
tools: Read, Edit, Bash(black:*), Bash(pytest:*)  # Formatter + tests only
```

### Bash Command Restrictions

```yaml
# Allowed patterns:
Bash(git:*)           # All git commands
Bash(npm run:*)       # Only npm run scripts
Bash(python:*)        # Python interpreter
Bash(pytest:*)        # Pytest runner

# Denied patterns:
Bash(rm:*)            # Dangerous deletion
Bash(sudo:*)          # Privilege escalation
Bash(curl:*)          # Arbitrary downloads
```

## Agent Execution Modes

### Mode 1: Inline (main context)
```bash
/role security
"Check this project for vulnerabilities"
# Runs in main session context
```

### Mode 2: Sub-agent (isolated context)
```bash
/role security --agent
"Perform comprehensive security audit"
# Runs in independent context, can work in parallel
```

### Mode 3: Parallel Multi-role
```bash
/multi-role security,performance,qa --agent
"Analyze security, performance, and quality"
# All roles run in parallel, results merged
```

## Agent Registration & Discovery

```bash
# List available agents
/agents

# Create new agent interactively
/agents create

# View specific agent
/agents view security-auditor

# Edit agent
/agents edit security-auditor

# Delete agent
/agents delete security-auditor
```

## Agent Validation Checklist

- [ ] `name` is kebab-case (e.g., `code-reviewer`)
- [ ] `description` includes "Use PROACTIVELY for"
- [ ] `tools` list only necessary tools
- [ ] No `Bash(*)` wildcard; specific patterns only
- [ ] `model` is `haiku` or `sonnet`
- [ ] `Proactive Triggers` section clearly defined
- [ ] No overlapping responsibilities with other agents
- [ ] YAML frontmatter is valid

## Best Practices

✅ **DO**:
- Design agents around specific expertise
- Use `--agent` flag for large analyses
- Combine multiple agents for complex tasks
- Name agents descriptively

❌ **DON'T**:
- Create overpowered agents with all tools
- Allow direct file modifications without approval
- Overlap agent responsibilities
- Use `Bash(*)` without specific patterns

---

## 🤝 Works Well With

**Complementary Skills:**
- **moai-cc-hooks** - Validate agent inputs/outputs with Pre/PostToolUse Hooks
- **moai-cc-commands** - Invoke agents from slash commands (`/alfred:2-run`)
- **moai-cc-settings** - Restrict agent tool access (principle of least privilege)
- **moai-cc-memory** - Cache agent results to avoid re-computation

**MoAI-ADK Workflows:**
- **`/alfred:2-run`** - code-builder pipeline (implementation-planner → tdd-implementer)
- **`/alfred:3-sync`** - tag-agent (verify TAG chains), doc-syncer (generate docs)
- **`/alfred:2-run` failures** - debug-helper agent auto-invoked for error diagnosis

**Example Integration (MoAI-ADK):**
```bash
# 1. Create code-review agent for /alfred:3-sync
@agent-cc-manager "Create code-reviewer agent for TRUST 5 validation"

# 2. Define tool restrictions in YAML
tools: Read, Glob, Grep, Bash(git:*)

# 3. Use in workflow
/alfred:3-sync  # Invokes code-reviewer agent before PR merge
```

**Common MoAI Patterns:**
- ✅ spec-builder (Plan) + moai-foundation-specs = SPEC authoring
- ✅ code-builder (Run) + moai-essentials-refactor = TDD implementation
- ✅ tag-agent (Sync) + moai-foundation-tags = TAG verification
- ✅ debug-helper (Run errors) + error logs = Fix-forward guidance

**General Claude Code Patterns:**
- ✅ Custom agents for code review, security audit, architecture
- ✅ Parallel agents for simultaneous analysis
- ✅ Agent orchestration in commands

**See Also:**
- 📖 **Orchestrator Guide:** `Skill("moai-cc-guide")` → SKILL.md
- 📖 **MoAI Workflows:** `Skill("moai-cc-guide")` → workflows/alfred-*
- 📖 **Orchestration:** `Skill("moai-cc-commands")` → Agent Orchestration

---

**Reference**: Claude Code Sub-agents documentation
**Version**: 1.0.0
