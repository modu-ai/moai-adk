# Agents System Reference

Understand MoAI-ADK's 19-member AI agent team structure.

## Overview

Alfred is a super-agent that orchestrates a **team of 19 experts**. Each agent is optimized for a specific domain or task and is automatically activated as needed.

## 19-Member Team Structure

### Organization Chart

```
┌─────────────────────────────────────────┐
│          Alfred (Super-Agent)            │
│         SPEC → TDD → Sync Orchestration  │
└──────────────────┬──────────────────────┘
                   │
        ┌──────────┼──────────┐
        │          │          │
   ┌────▼─────┐  ┌▼────────┐  ┌▼────────────┐
   │  Core    │  │ Expert  │  │ Built-in   │
   │ Agents   │  │Agents   │  │ Agents     │
   │  (10)    │  │ (6)     │  │ (2)        │
   └──────────┘  └────────┘  └────────────┘
```

## Agent Classification

### 1️⃣ Core Sub-agents (10)

Manage the entire project lifecycle:

| Agent                   | Role                     | Activation Condition          |
| ----------------------- | ------------------------ | ----------------------------- |
| **project-manager**     | Project initialization and settings | `/alfred:0-project`  |
| **spec-builder**        | SPEC writing (EARS syntax) | `/alfred:1-plan`     |
| **implementation-planner** | Architecture and implementation planning | `/alfred:2-run` start |
| **tdd-implementer**     | RED→GREEN→REFACTOR execution | `/alfred:2-run` during |
| **doc-syncer**          | Auto document generation and sync | `/alfred:3-sync`     |
| **tag-agent**           | TAG validation and traceability management | `/alfred:3-sync`     |
| **git-manager**         | Git workflow automation | All stages            |
| **trust-checker**      | TRUST 5 principle validation | `/alfred:2-run` completion |
| **quality-gate**        | Release readiness check | `/alfred:3-sync`     |
| **debug-helper**        | Error analysis and resolution | Auto-activated when needed |

### 2️⃣ Expert Agents (6)

Support domain-specific tasks:

| Agent            | Domain                        | Activation Condition              |
| ---------------- | ----------------------------- | --------------------------------- |
| **backend-expert** | API, Server, DB Architecture | Server/API keywords in SPEC   |
| **frontend-expert** | UI, State Management, Performance | Frontend keywords in SPEC |
| **devops-expert** | Deployment, CI/CD, Infrastructure | Deployment keywords in SPEC |
| **ui-ux-expert** | Design System, Accessibility | Design keywords in SPEC     |
| **security-expert** | Security Analysis, Vulnerability Diagnosis | Security keywords in SPEC |
| **database-expert** | DB Design, Optimization, Migration | DB keywords in SPEC         |

### 3️⃣ Built-in Claude Agents (2)

For complex reasoning needs:

- **Claude Opus/Sonnet**: Complex reasoning, deep analysis
- **Claude Haiku**: Lightweight tasks, fast processing

## Hybrid Patterns

### Lead-Specialist Pattern

Specialized domain experts support the lead agent:

```
User Request
    ↓
Alfred (Lead)
    ├─→ Frontend keyword detected
    │   └─→ frontend-expert activated
    ├─→ Database keyword detected
    │   └─→ database-expert activated
    └─→ Security keyword detected
        └─→ security-expert activated
```

**Use Cases**:

- UI component design needed → UI/UX Expert
- DB performance optimization → Database Expert
- Security review → Security Expert

### Master-Clone Pattern

Large-scale tasks are handled in parallel by Alfred clones:

```
Large-scale task (100+ files, 5+ steps)
    ↓
Master Alfred (Orchestration)
    ├─→ Clone-1: Module A refactoring
    ├─→ Clone-2: Module B refactoring
    └─→ Clone-3: Module C refactoring
    ↓
Result merge and integration
```

**Use Cases**:

- Large-scale migration (v1.0 → v2.0)
- Full architecture refactoring
- Multi-domain concurrent work

## Agent Collaboration Methods

### Sequential Collaboration

```
Alfred → spec-builder → implementation-planner → tdd-implementer → doc-syncer
```

### Parallel Collaboration

```
Alfred
├─→ backend-expert (API design)
├─→ frontend-expert (UI design)
└─→ database-expert (DB schema)
    ↓
All complete, then tdd-implementer executes
```

### Conditional Collaboration

```
Alfred
└─→ SPEC analysis
    ├─ Security keyword? → security-expert activated
    ├─ Deployment needed? → devops-expert activated
    └─ Performance optimization? → debug-helper activated
```

## Agent Activation Diagram

```
User Request
    ↓
┌─────────────────────────────────┐
│   Alfred (Intent Detection)      │
├─────────────────────────────────┤
│ 1. Request classification        │
│ 2. Domain detection              │
│ 3. Required agent determination   │
└──────────────┬──────────────────┘
               │
        ┌──────┴──────┐
        │             │
    Core Agent?    Domain Expert?
        │             │
        ▼             ▼
    ┌────────┐    ┌──────────────┐
    │project │    │backend-expert│
    │manager │    │frontend-expert
    └────────┘    └──────────────┘
        ↓             ↓
    ┌────────────────────┐
    │  TDD Execution     │
    │ (tdd-implementer)  │
    └────────────────────┘
        ↓
    ┌────────────────────┐
    │  Verification      │
    │ (trust-checker)    │
    └────────────────────┘
        ↓
    ┌────────────────────┐
    │  Document Sync     │
    │ (doc-syncer)       │
    └────────────────────┘
        ↓
    Complete
```

## Agent Selection Algorithm

How Alfred decides which agents to activate:

```python
# Decision tree
def select_agents(user_request):
    # 1. Domain detection
    if "database" or "schema" or "query" in request:
        activate(database_expert)

    if "frontend" or "ui" or "component" in request:
        activate(frontend_expert)

    if "security" or "auth" or "encryption" in request:
        activate(security_expert)

    if "deploy" or "ci/cd" or "docker" in request:
        activate(devops_expert)

    # 2. Scale judgment
    if file_count > 100 or steps > 5:
        use_master_clone_pattern()
    else:
        use_lead_specialist_pattern()

    # 3. Core agents always activated
    activate(core_agents)
```

## <span class="material-icons">library_books</span> Detailed Guides

- **[Core Sub-agents](core.md)** - Detailed 10 agents
- **[Expert Agents](experts.md)** - Detailed 6 experts

## Related Documents

- [Alfred Super-Agent](guides/alfred/index.md) - Alfred concepts and workflow
- [Skills System](skills/index.md) - 55+ Claude Skills
- [Architecture Explanation](advanced/architecture.md) - 4-layer stack structure

______________________________________________________________________

**Next**: [Core Sub-agents](core.md) or [Expert Agents](experts.md)
