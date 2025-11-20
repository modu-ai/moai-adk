---
name: moai-core-rules
version: 4.0.0
updated: 2025-11-20
status: stable
tier: foundation
description: Core rules for Alfred SuperAgent - 3-Layer Architecture, 4-Step Workflow, Agent-First
allowed-tools: [Read, Bash, AskUserQuestion]
---

# Alfred SuperAgent Core Rules

**Foundation for Agent-Based Development**

> **Architecture**: Commands → Agents → Skills  
> **Workflow**: ADAP (Analyze, Design, Assure, Produce)  
> **Principles**: Agent-First, Delegation, Quality Gates

---

## Overview

Defines essential rules for Alfred's decision-making and execution.

### Key Responsibilities

1.  **3-Layer Architecture**: Commands orchestrate, Agents execute, Skills provide knowledge.
2.  **4-Step Workflow**: Analyze (research) → Design → Assure (quality) → Produce.
3.  **Agent-First**: Delegate all execution to specialized agents.
4.  **TRUST 4**: Enforce Test, Readable, Unified, Secured principles.

---

## Rule 1: 3-Layer Architecture

```
┌──────────────────────┐
|  generate Commands (Orchestration) │  ← User entry points
│  /alfred:1-plan      │    No direct execution
└───────┬──────────────┘
        │ Task(subagent_type=...)
        ↓
┌──────────────────────┐
│  Agents (Execution)  │  ← Spec-builder, tdd-implementer
│  Deep reasoning      │    Domain expertise
└───────┬──────────────┘
        │ Skill("skill-name")
        ↓
┌──────────────────────┐
│  Skills (Knowledge)  │  ← Reusable patterns
│  Stateless playbooks │    <1000 lines
└──────────────────────┘
```

### Command Layer (Orchestration Only)

**Do**: Delegate tasks, ask user questions.  
**Don't**: Execute bash, write files, run tests directly.

```python
# ❌ Wrong: Direct execution
bash("git commit -m 'message'")

# ✅ Correct: Delegate to agent
Task(subagent_type="git-manager", description="Commit changes")
```

### Agent Layer (Execution)

**Do**: Complex analysis, planning, skill invocation.  
**Don't**: Call other agents directly (use Commands for multi-agent orchestration).

### Success Skill Layer (Knowledge Capsules)

**Do**: Provide patterns, best practices.  
**Don't**: Invoke other skills, execute tasks, maintain state.

---

## Rule 2: 4-Step Workflow (ADAP)

1.  **Analyze**: Research latest docs via WebSearch/WebFetch.
2.  **Design**: Architecture based on November 2025 info.
3.  **Assure**: Validate quality (TRUST 4).
4.  **Produce**: Generate files, commit.

---

## Rule 3: Agent-First Paradigm

**Forbidden**: Commands MUST NOT directly:

- Execute bash commands
- Read/write files
- Run git operations
- Execute tests

**Mandatory Delegation**:
| Task | Delegate To |
|------|-------------|
| Planning | plan-agent |
| Code | tdd-implementer |
| Tests | test-engineer |
| Docs | doc-syncer |
| Git | git-manager |

---

## Rule 4: TRUST 4 Quality Gates

Before commit:

- [ ] **Test**: ≥85% coverage
- [ ] **Readable**: Pylint ≥8.0, clear naming
- [ ] **Unified**: Consistent patterns, no duplication
- [ ] **Secured**: OWASP scan passed (bandit)

---

## Rule 5: AskUserQuestion Usage

**When to use**:

- Ambiguous requirements ("Create a module" → "What kind?")
- Technology choice ("Which framework?")
- Architecture decisions ("SQL vs NoSQL?")

**Example**:

```python
AskUserQuestion({
  "question": "Which database?",
  "options": [
    {"label": "PostgreSQL", "description": "Relational"},
    {"label": "MongoDB", "description": "Document"}
  ]
})
```

---

## Rule 6: Commit Message Standards

**TDD Cycle Commits**:

- **RED**: `test(@TAG-001): Add tests` (tests fail)
- **GREEN**: `feat(@TAG-001): Implement feature` (tests pass)
- **REFACTOR**: `refactor(@TAG-001): Optimize code` (tests still pass)

---

## Validation Checklist

- [ ] **Command**: Only orchestration, no direct execution?
- [ ] **Agent**: Proper delegation to subagents?
- [ ] **Skill**: Invoked correctly via `Skill("name")`?
- [ ] **TRUST 4**: All quality gates passed?
- [ ] **Commit**: Follows conventional commits?

---

## Related Skills

- `moai-foundation-trust`: TRUST 4 principles
- `moai-foundation-tags`: TAG traceability
- `moai-foundation-git`: Git workflows

---

**Last Updated**: 2025-11-20
