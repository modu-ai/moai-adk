# Agent Delegation Patterns

Core patterns used by Alfred when delegating to agents via Task().

## Base Delegation Syntax

```
result = await Task(
    subagent_type="agent_name",
    prompt="specific and clear task description",
    context={"required": "context"}
)
```

---

## Pattern 1: Sequential Delegation

**Use Case**: When agents have dependencies on each other

**Flow**:
```
1. design-phase complete
   ↓
2. Pass results as context to next agent
   ↓
3. implementation-phase begins
```

**Example**:
```
# Phase 1: Design
design = Task(api-designer, "API design")

# Phase 2: Implementation (pass design results)
implementation = Task(
    backend-expert,
    "Implement API",
    context={"api_design": design}
)
```

---

## Pattern 2: Parallel Delegation

**Use Case**: When agents are independent

**Flow**:
```
Task1 (backend)     →  Wait for completion
Task2 (frontend)    →  Wait for completion
Task3 (documentation) →  Wait for completion
          All complete →  Integrate
```

**Example**:
```
results = await Promise.all([
    Task(backend-expert, "Implement backend"),
    Task(frontend-expert, "Implement frontend"),
    Task(docs-manager, "Generate documentation")
])
```

---

## Pattern 3: Conditional Delegation

**Use Case**: Route to different agents based on analysis results

**Flow**:
```
Analysis complete
  ↓
Determine issue type
  ├→ Security issue → security-expert
  ├→ Performance issue → performance-engineer
  └→ Quality issue → quality-gate
```

---

## Context Passing Guide

### Required Context Fields

```
context={
    "spec_id": "SPEC-001",           # Task ID
    "requirements": [list],          # Requirements
    "constraints": [constraints]     # Constraints
}
```

### Exclude Unnecessary Data

❌ Full codebase
❌ All file contents
❌ Historical conversation logs
❌ Large binary data

### Optimal Context Size

- Minimum: Only information agent needs for the task
- Maximum: Under 50K tokens

---

## Error Handling

**On Failure**:
1. Request error analysis from `debug-helper`
2. Identify root cause
3. Retry with different agent or implement recovery

---

## Delegation Checklist

- [ ] Agent name correct (lowercase, hyphens)
- [ ] Prompt is specific and clear
- [ ] Context contains only necessary information
- [ ] Dependencies considered (sequential vs parallel)
- [ ] Error handling plan in place

---

Refer to @.moai/memory/agents.md for detailed agent descriptions.
