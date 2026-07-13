# Delegation Patterns - Agent Orchestration

Purpose: Task delegation strategies to specialized agents, ensuring efficient workflow orchestration without direct execution.

Version: 2.0.0 (Modular Split)
Last Updated: 2026-01-06

---

## Quick Reference (30 seconds)

Core Principle: MoAI NEVER executes directly. All work via `Agent()` delegation to specialized agents.

Three Primary Patterns:
1. Sequential - Dependencies between agents (Phase 1 → Phase 2 → Phase 3)
2. Parallel - Independent agents (Backend + Frontend + Docs simultaneously)
3. Conditional - Analysis-driven routing (security issue → per-spawn general-purpose with security instructions)

Base Syntax (illustrative pseudocode):
```python
result = await Agent(
    subagent_type="general-purpose",   # per-spawn delegation (domain carried in prompt)
    prompt="specific task description",
    context={"necessary": "data"}
)
```

> Note: in practice, retained agents (`manager-spec`, `manager-develop`, `manager-docs`, etc.) are invoked through natural-language delegation ("Use the {agent} subagent to {task}"), never a `subagent_type` code literal — see [agents-reference.md](agents-reference.md). The pseudocode below names agents illustratively to show flow; per-spawn `general-purpose` (with domain instructions in the prompt) is the canonical mechanism for cross-cutting domain work (security, performance, database). Legacy tiered names (`code-backend`, `security-expert`, `core-quality`, etc.) are archived and rejected at spawn.

Agent Selection:
- Simple (1 file): 1-2 agents sequential
- Medium (3-5 files): 2-3 agents sequential
- Complex (10+ files): 5+ agents parallel/sequential mix

Context Size: 20-30K tokens target, 50K maximum

Extended Documentation:
- [Delegation Implementation](delegation-implementation.md) - Detailed patterns and code
- [Delegation Advanced](delegation-advanced.md) - Error handling and hybrid patterns

---

## Implementation Guide (5 minutes)

### Pattern 1: Sequential Delegation

Use Case: When agents have dependencies on each other.

Flow Diagram:
```
Phase 1: Design
    ↓ (design results)
Phase 2: Implementation
    ↓ (implementation + design)
Phase 3: Documentation
    ↓ (all results)
Phase 4: Quality Gate
```

Example:
```python
async def implement_feature_sequential(feature_description: str):
    """Sequential workflow with context passing."""

    # Phase 1: SPEC Generation
    spec_result = await Agent(
        subagent_type="manager-spec",
        prompt=f"Generate SPEC for: {feature_description}",
        context={"requirements": ["TRUST 5 compliance", "≥85% coverage"]}
    )

    execute_clear()  # Save tokens

    # Phase 2: API Design (depends on SPEC) — run-phase implementation
    api_result = await Agent(
        subagent_type="manager-develop",
        prompt="Design REST API for feature (cycle_type=ddd, backend/API domain)",
        context={"spec_id": spec_result.spec_id}
    )

    # Phase 3: Implementation (depends on API design) — run-phase
    backend_result = await Agent(
        subagent_type="manager-develop",
        prompt="Implement backend with DDD (cycle_type=ddd, backend domain)",
        context={"spec_id": spec_result.spec_id, "api_design": api_result}
    )

    return {"spec": spec_result, "api": api_result, "backend": backend_result}
```

---

### Pattern 2: Parallel Delegation

Use Case: When agents work on independent tasks simultaneously.

Flow Diagram:
```
Start
    → Backend Agent → Result 1
    → Frontend Agent → Result 2
    → Database Agent → Result 3
    → Docs Agent → Result 4
        ↓
    All Complete → Integration
```

Example:
```python
async def implement_feature_parallel(spec_id: str):
    """Parallel workflow for independent tasks."""

    results = await Promise.all([
        Agent(
            subagent_type="general-purpose",
            prompt=f"Implement backend for {spec_id} (backend domain)",
            context={"spec_id": spec_id, "focus": "API endpoints"}
        ),
        Agent(
            subagent_type="general-purpose",
            prompt=f"Implement UI for {spec_id} (frontend domain)",
            context={"spec_id": spec_id, "focus": "Components"}
        ),
        Agent(
            subagent_type="general-purpose",
            prompt=f"Design database for {spec_id} (database domain)",
            context={"spec_id": spec_id, "focus": "Schema"}
        )
    ])

    backend, frontend, database = results

    # Integration step (sequential, depends on parallel results)
    integration = await Agent(
        subagent_type="manager-develop",
        prompt="Run integration tests (cycle_type=tdd)",
        context={"backend": backend.summary, "frontend": frontend.summary}
    )

    return {"backend": backend, "frontend": frontend, "database": database}
```

Benefits:

- Execution Time: 3x faster than sequential
- Token Sessions: 4x 200K each (independent contexts)
- Context Isolation: Cleaner, no interference
- Error Impact: Isolated failures

---

### Pattern 3: Conditional Delegation

Use Case: Route to different agents based on analysis results.

Flow Diagram:
```
Analysis Agent → Determines issue type
    → Security issue → general-purpose (security instructions)
    → Performance issue → general-purpose (performance instructions)
    → Quality issue → sync-auditor
    → Bug → manager-develop (autofix)
```

Example:
```python
async def handle_issue_conditional(issue_description: str):
    """Conditional routing based on issue analysis."""

    analysis = await Agent(
        subagent_type="general-purpose",
        prompt=f"Analyze issue (diagnostics): {issue_description}",
        context={"focus": "classification"}
    )

    if analysis.category == "security":
        return await Agent(
            subagent_type="general-purpose",
            prompt="Analyze and fix security issue (security domain)",
            context={"issue": issue_description, "analysis": analysis.details}
        )

    elif analysis.category == "performance":
        return await Agent(
            subagent_type="general-purpose",
            prompt="Optimize performance issue (performance domain)",
            context={"issue": issue_description, "bottleneck": analysis.bottleneck}
        )

    else:
        return await Agent(
            subagent_type="manager-develop",
            prompt="Debug and fix issue (cycle_type=autofix)",
            context={"issue": issue_description, "analysis": analysis.details}
        )
```

---

## Advanced Patterns

For comprehensive implementation patterns including context optimization, error handling, and hybrid workflows, see:

- [Delegation Implementation](delegation-implementation.md) - Detailed code patterns
- [Delegation Advanced](delegation-advanced.md) - Error handling and hybrid patterns

---

## Works Well With

Agents (Delegation Targets — see [agents-reference.md](agents-reference.md) for the full retained catalog):
- manager-spec - SPEC / plan-phase authoring
- manager-develop - Run-phase implementation (DDD/TDD/autofix); per-spawn general-purpose for backend/frontend domain work
- manager-docs - Sync-phase documentation
- sync-auditor - Quality validation (TRUST 5)
- general-purpose (per-spawn, with domain instructions) - Security, performance, database, and other cross-cutting specialist work

Skills:
- moai-foundation-core - Context management (see [token-optimization.md](token-optimization.md))

Foundation Modules:
- [Token Optimization](token-optimization.md) - Context passing strategies
- [Execution Rules](execution-rules.md) - Security constraints for delegation

---

Version: 2.0.0
Last Updated: 2026-01-06
Status: Production Ready
