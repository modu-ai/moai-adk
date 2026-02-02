# Delegation Patterns - Agent Orchestration

Purpose: Task delegation strategies to specialized agents, ensuring efficient workflow orchestration without direct execution.

Version: 2.0.0 (Modular Split)
Last Updated: 2026-01-06

---

## Quick Reference (30 seconds)

Core Principle: MoAI NEVER executes directly. All work via `Task()` delegation to specialized agents.

Three Primary Patterns:
1. Sequential - Dependencies between agents (Phase 1 → Phase 2 → Phase 3)
2. Parallel - Independent agents (Backend + Frontend + Docs simultaneously)
3. Conditional - Analysis-driven routing (Security issue → security-expert)

Base Syntax:
```python
result = await Task(
    subagent_type="agent-name",
    prompt="specific task description",
    context={"necessary": "data"}
)
```

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
    spec_result = await Task(
        subagent_type="workflow-spec",
        prompt=f"Generate SPEC for: {feature_description}",
        context={"requirements": ["TRUST 5 compliance", "≥85% coverage"]}
    )

    execute_clear()  # Save tokens

    # Phase 2: API Design (depends on SPEC)
    api_result = await Task(
        subagent_type="api-designer",
        prompt="Design REST API for feature",
        context={"spec_id": spec_result.spec_id}
    )

    # Phase 3: Implementation (depends on API design)
    backend_result = await Task(
        subagent_type="code-backend",
        prompt="Implement backend with DDD",
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
        Task(
            subagent_type="code-backend",
            prompt=f"Implement backend for {spec_id}",
            context={"spec_id": spec_id, "focus": "API endpoints"}
        ),
        Task(
            subagent_type="code-frontend",
            prompt=f"Implement UI for {spec_id}",
            context={"spec_id": spec_id, "focus": "Components"}
        ),
        Task(
            subagent_type="data-database",
            prompt=f"Design database for {spec_id}",
            context={"spec_id": spec_id, "focus": "Schema"}
        )
    ])

    backend, frontend, database = results

    # Integration step (sequential, depends on parallel results)
    integration = await Task(
        subagent_type="core-quality",
        prompt="Run integration tests",
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
    → Security issue → security-expert
    → Performance issue → performance-engineer
    → Quality issue → core-quality
    → Bug → support-debug
```

Example:
```python
async def handle_issue_conditional(issue_description: str):
    """Conditional routing based on issue analysis."""

    analysis = await Task(
        subagent_type="support-debug",
        prompt=f"Analyze issue: {issue_description}",
        context={"focus": "classification"}
    )

    if analysis.category == "security":
        return await Task(
            subagent_type="security-expert",
            prompt="Analyze and fix security issue",
            context={"issue": issue_description, "analysis": analysis.details}
        )

    elif analysis.category == "performance":
        return await Task(
            subagent_type="performance-engineer",
            prompt="Optimize performance issue",
            context={"issue": issue_description, "bottleneck": analysis.bottleneck}
        )

    else:
        return await Task(
            subagent_type="support-debug",
            prompt="Debug and fix issue",
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

Agents (Delegation Targets):
- workflow-spec - SPEC generation
- workflow-ddd - DDD implementation
- code-backend - Backend development
- code-frontend - Frontend development
- security-expert - Security analysis
- core-quality - Quality validation

Skills:
- moai-foundation-token-optimization - Context management

Foundation Modules:
- [Token Optimization](token-optimization.md) - Context passing strategies
- [Execution Rules](execution-rules.md) - Security constraints for delegation

---

Version: 2.0.0
Last Updated: 2026-01-06
Status: Production Ready
