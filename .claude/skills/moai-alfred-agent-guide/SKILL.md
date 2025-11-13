---
name: "moai-alfred-agent-guide"
version: "5.0.0"
created: 2025-10-01
updated: 2025-11-13
status: stable
tier: specialization
description: "Essential guide to MoAI-ADK's 19-agent team structure, agent selection decision trees, and 2025 multi-agent orchestration patterns. Core principles for agent delegation, collaboration protocols, and performance optimization."
allowed-tools: "Read, Glob, Grep, TodoWrite, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "alfred"
secondary-agents: [plan-agent, tdd-implementer, test-engineer, doc-syncer, git-manager, qa-validator, tag-agent]
keywords: [alfred, agents, orchestration, team, decision-tree, model-selection, collaboration]
tags: [agent-guide, team-structure, orchestration, haiku-vs-sonnet, delegation, handoff]
orchestration: 
can_resume: true
typical_chain_position: "initial"
depends_on: []
---

# moai-alfred-agent-guide

**Complete Guide to 19-Agent Team Orchestration**

> **Primary Agent**: alfred (SuperAgent Orchestrator)  
> **Team Size**: 19 specialized agents  
> **Core Principle**: Delegate-first architecture  
> **Optimization**: Hybrid Sonnet/Haiku model selection

---

## Quick Start

### What It Does

This skill provides the **operational framework** for MoAI-ADK's 19-agent team:

- **Agent Selection**: Systematic decision trees for any task type
- **Orchestration Patterns**: 2025 best practices for multi-agent coordination
- **Model Optimization**: Sonnet vs Haiku selection criteria (60-70% cost savings)
- **Collaboration Protocols**: Handoff sequences and cross-agent communication

### When to Use

**Always** when:
- Delegating executable tasks to specialized agents
- Coordinating multi-agent workflows
- Selecting appropriate Claude models (Sonnet vs Haiku)
- Planning complex feature implementation

**Basic Usage**:

```bash
Skill("moai-alfred-agent-guide")
```

### Team Structure (19 Agents)

```
Orchestration Layer (1):
  └─ alfred (SuperAgent)

Core Agents (10):
  ├─ plan-agent (Strategic Planning)
  ├─ tdd-implementer (TDD Development)
  ├─ test-engineer (Testing & QA)
  ├─ doc-syncer (Documentation)
  ├─ git-manager (Git Operations)
  ├─ qa-validator (Quality Assurance)
  ├─ project-manager (Project Coordination)
  ├─ debug-helper (Troubleshooting)
  ├─ trust-checker (TRUST 5 Validation)
  └─ tag-agent (TAG System Management)

Domain Specialists (6):
  ├─ backend-expert (Backend Development)
  ├─ frontend-expert (Frontend Development)
  ├─ database-expert (Database Design)
  ├─ security-expert (Security & Auth)
  ├─ devops-expert (Infrastructure)
  └─ api-designer (API Architecture)

Built-in Agents (2):
  ├─ Plan (Strategy & Analysis)
  └─ WebSearch (Research & Validation)
```

**Core Principle**: **Delegate-first architecture** — Every task routes to the most specialized agent. Alfred orchestrates but never executes directly.

---

## Implementation

### Agent Selection Decision Tree

**Step 1: Determine Task Category**
- Infrastructure/Config? → git-manager/devops-expert
- Code Implementation? → tdd-implementer + domain-expert
- Documentation? → doc-syncer
- Quality/Validation? → qa-validator/trust-checker
- Domain-Specific? → backend/frontend/database/security/expert
- Planning/Strategy? → plan-agent
- Debugging? → debug-helper
- Research? → WebSearch

**Step 2: Select Claude Model**
```
Complex reasoning/strategic decisions? → Sonnet
Pattern execution/well-defined tasks? → Haiku
```

**Step 3: Delegate with Context**
```python
Task(
    subagent_type="agent-name",
    description="Clear task description",
    prompt="Detailed context + requirements + constraints"
)
```

### Core Delegation Patterns

**Pattern 1: Orchestrator-Worker (Most Common)**
```python
# User request: "Add user authentication"
alfred → plan-agent → tdd-implementer → test-engineer → doc-syncer
```

**Pattern 2: Handoff Orchestration**
```python
# Sequential context transfer
plan-agent → tdd-implementer → test-engineer → qa-validator → git-manager
```

**Pattern 3: Hierarchical Delegation**
```python
# Domain expert coordinates specialists
alfred → backend-expert → [database-expert, security-expert, tdd-implementer]
```

### Haiku vs Sonnet Model Selection

| Scenario | Model | Rationale |
|----------|-------|-----------|
| **Strategic Planning** | Sonnet | Complex reasoning, multiple constraints |
| **Architecture Design** | Sonnet | Novel solutions, security implications |
| **Security Review** | Sonnet | Threat modeling, vulnerability analysis |
| **Debugging Complex Issues** | Sonnet | Root cause investigation, hypothesis generation |
| **Code Implementation (TDD)** | Haiku | Pattern-based generation, fast execution |
| **Testing & QA** | Haiku | Rule-based validation, coverage analysis |
| **Documentation** | Haiku | Template-driven content, structured formatting |
| **Git Operations** | Haiku | Standardized workflows, deterministic logic |

**Cost Optimization**: Hybrid pattern reduces costs 60-70% while maintaining quality.

### Key Agents and Their Roles

#### alfred (SuperAgent Orchestrator)
- **Model**: Sonnet
- **Role**: Workflow coordination, user interaction, task routing
- **Never**: Executes tasks directly - always delegates
- **Skills**: Access to all 55 Claude Skills

#### plan-agent (Strategic Planning)
- **Model**: Sonnet
- **Role**: Task decomposition, dependency analysis, risk assessment
- **Use**: New features, refactoring strategies, migration planning
- **Output**: Structured plans with task DAG and agent assignments

#### tdd-implementer (TDD Development)
- **Model**: Haiku
- **Role**: RED-GREEN-REFACTOR cycle execution
- **Use**: Feature implementation, bug fixes, API development
- **Process**: Tests first → minimal implementation → refactor

#### test-engineer (Testing & QA)
- **Model**: Haiku
- **Role**: Test coverage validation (90%+ target), regression testing
- **Use**: QA gate enforcement, test infrastructure setup
- **Output**: Coverage reports, missing scenarios

#### Domain Specialists (backend/frontend/database/security/expert)
- **Model**: Sonnet for architecture, Haiku for implementation
- **Role**: Domain-specific expertise and best practices
- **Use**: Specialized technical decisions, domain validation

---

## Advanced

### Performance Optimization Strategies

**Context Budget Management**
```python
# ✅ GOOD: Load only relevant skills
Task(
    subagent_type="doc-syncer",
    context_includes=["moai-alfred-document-management"]  # ~20K tokens
)

# ❌ BAD: Load all skills for simple task
Task(
    subagent_type="doc-syncer",
    context_includes="all_skills"  # 200K+ tokens
)
```

**Parallel Execution**
```python
# Independent agents can run in parallel
import asyncio
async def parallel_design():
    backend_task = asyncio.create_task(Task(subagent_type="backend-expert", ...))
    frontend_task = asyncio.create_task(Task(subagent_type="frontend-expert", ...))
    return await asyncio.gather(backend_task, frontend_task)
```

**Model Selection Guidelines**
- **Sonnet (20% of tasks)**: Strategic decisions, security analysis, complex debugging
- **Haiku (80% of tasks)**: Pattern execution, testing, documentation, git operations

### Collaboration Protocols

**Handoff Best Practices**
1. Always include context from previous agent
2. Document assumptions and constraints
3. Validate handoff completeness before proceeding

**Cross-Agent Communication**
```python
# Sequential delegation (alfred orchestrates)
backend_design = Task(subagent_type="backend-expert", ...)
frontend_contract = Task(
    subagent_type="frontend-expert",
    prompt=f"Build UI for API:\n{backend_design.endpoints}"
)
```

**Conflict Resolution**
- Use Plan agent for strategic disagreements
- Alfred makes final decisions based on project constraints
- Document resolution rationale

### Anti-Patterns to Avoid

❌ **Agent Bypassing**: Commands executing tasks directly
❌ **Over-Delegation**: 5 agents for simple tasks
❌ **Context Leakage**: Irrelevant context wastes tokens
❌ **Silent Failures**: No error handling or validation

---

## Security & Compliance

### Agent Security Considerations

**Access Control**
- Validate agent permissions (file access, API calls)
- Sanitize agent inputs (prevent prompt injection)
- Implement agent rate limiting (prevent runaway loops)

**TRUST 5 Enforcement**
- **Test First**: qa-validator validates test coverage
- **Readable**: Code style and documentation checks
- **Unified**: Consistency across architecture patterns
- **Secured**: security-expert validates posture
- **Trackable**: git-manager ensures traceable commits

### Compliance Notes

**Enterprise v4.0 Standards**
- Progressive Disclosure structure (Quick/Implementation/Advanced)
- Quality gate validation before handoffs
- Comprehensive error handling and escalation
- Agent performance metrics tracking

---

## Related Skills

**Prerequisites**:
- `Skill("moai-alfred-workflow")` - 4-step workflow understanding
- `Skill("moai-foundation-trust")` - TRUST 5 principles

**Complementary**:
- `Skill("moai-alfred-context-budget")` - Agent context optimization
- `Skill("moai-alfred-todowrite-pattern")` - Multi-agent task tracking
- `Skill("moai-domain-backend")` - Backend-specific guidance

**Reference Materials**:
- `reference.md` - Complete 19-agent team roster with detailed responsibilities
- `examples.md` - Practical agent selection examples and decision trees

---

*Enterprise v4.0 Compliant*  
*Optimized: 2226 lines → 250 lines (89% reduction)*  
*Maintained by: Primary Agent (alfred)*
