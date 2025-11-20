# Agent Delegation Patterns

## Overview

This document defines standardized patterns for delegating tasks to MoAI-ADK agents, ensuring optimal workflow execution and context management.

## Delegation Principles

### Core Rules

1. **Never Execute Directly**: Always use `Task()` for delegation
2. **Pass Relevant Context**: Include necessary information for agent success
3. **Choose Appropriate Agents**: Match task complexity to agent specialization
4. **Handle Results Properly**: Process and validate agent outputs
5. **Manage Context Efficiently**: Use context passing between related tasks

### Delegation Syntax

```python
# Standard delegation pattern
result = await Task(
    subagent_type="agent_name",
    prompt="specific task description",
    context={"key": "value"},
    debug=false/true
)
```

## Agent Selection Decision Tree

### Task Type Analysis

**Requirements and Planning**:
```
User request contains "plan", "spec", "requirements", "design"
→ DELEGATE to: spec-builder
```

**Implementation Tasks**:
```
User request contains "implement", "build", "create", "develop"
→ DELEGATE to: tdd-implementer
```

**API/Backend Tasks**:
```
User request contains "API", "backend", "server", "database"
→ DELEGATE to: api-designer OR backend-expert
```

**Frontend/UI Tasks**:
```
User request contains "UI", "frontend", "component", "interface"
→ DELEGATE to: frontend-expert OR component-designer
```

**Security Tasks**:
```
User request contains "security", "auth", "vulnerability", "OWASP"
→ DELEGATE to: security-expert
```

**Testing Tasks**:
```
User request contains "test", "validate", "verify", "quality"
→ DELEGATE to: test-engineer OR quality-gate
```

**Documentation Tasks**:
```
User request contains "document", "guide", "manual", "README"
→ DELEGATE to: docs-manager
```

## Context Passing Patterns

### Simple Context Transfer

```python
# Basic context passing
design_result = await Task(
    subagent_type="api-designer",
    prompt="Design REST API for user authentication"
)

# Pass result to next agent
implementation = await Task(
    subagent_type="backend-expert",
    prompt=f"Implement the designed API: {design_result.api_spec}",
    context={"api_design": design_result.api_spec}
)
```

### Rich Context Transfer

```python
# Comprehensive context passing
spec_result = await Task(
    subagent_type="spec-builder",
    prompt="Create specification for e-commerce platform"
)

# Rich context for implementation
implementation = await Task(
    subagent_type="tdd-implementer",
    prompt="Implement according to specification",
    context={
        "specification": spec_result.spec_content,
        "requirements": spec_result.requirements,
        "test_cases": spec_result.test_cases,
        "constraints": spec_result.constraints
    }
)
```

### Chained Context Transfer

```python
# Multi-agent context chain
# 1. Design phase
design = await Task(
    subagent_type="ui-ux-expert",
    prompt="Design user authentication interface"
)

# 2. Component design
components = await Task(
    subagent_type="component-designer",
    prompt="Create components based on design",
    context={"ux_design": design.design_specs}
)

# 3. Implementation
implementation = await Task(
    subagent_type="frontend-expert",
    prompt="Implement designed components",
    context={
        "ux_design": design.design_specs,
        "component_specs": components.component_library
    }
)
```

## Workflow Patterns

### Simple Task Delegation

```python
# Single agent delegation
async def implement_simple_feature(feature_description):
    result = await Task(
        subagent_type="tdd-implementer",
        prompt=f"Implement: {feature_description}"
    )
    return result
```

### Complex Multi-Agent Workflow

```python
# Complex feature requiring multiple agents
async def implement_complex_feature(requirements):
    # Phase 1: Specification
    spec = await Task(
        subagent_type="spec-builder",
        prompt=f"Create specification: {requirements}",
        context={"complexity": "high"}
    )

    # Phase 2: Architecture Design
    architecture = await Task(
        subagent_type="api-designer",
        prompt="Design system architecture",
        context={"specification": spec.content}
    )

    # Phase 3: Backend Implementation
    backend = await Task(
        subagent_type="backend-expert",
        prompt="Implement backend services",
        context={
            "specification": spec.content,
            "architecture": architecture.design
        }
    )

    # Phase 4: Security Review
    security = await Task(
        subagent_type="security-expert",
        prompt="Review and secure implementation",
        context={"backend_code": backend.implementation}
    )

    # Phase 5: Quality Validation
    validation = await Task(
        subagent_type="quality-gate",
        prompt="Validate complete implementation",
        context={
            "specification": spec.content,
            "backend_code": backend.implementation,
            "security_review": security.analysis
        }
    )

    return validation
```

### Parallel Execution Pattern

```python
# Parallel agent execution for independent tasks
async def implement_parallel_features(features):
    tasks = []

    for feature in features:
        if feature.type == "backend":
            task = Task(
                subagent_type="backend-expert",
                prompt=f"Implement: {feature.description}",
                context={"feature_id": feature.id}
            )
        elif feature.type == "frontend":
            task = Task(
                subagent_type="frontend-expert",
                prompt=f"Implement: {feature.description}",
                context={"feature_id": feature.id}
            )
        tasks.append(task)

    # Execute all tasks in parallel
    results = await asyncio.gather(*tasks)
    return results
```

### Error Handling Pattern

```python
# Robust delegation with error handling
async def safe_delegation(task_description, agent_type, context=None):
    try:
        result = await Task(
            subagent_type=agent_type,
            prompt=task_description,
            context=context or {}
        )
        return result
    except Exception as e:
        # Analyze error with debug helper
        error_analysis = await Task(
            subagent_type="debug-helper",
            prompt=f"Analyze delegation error: {e}",
            context={
                "failed_task": task_description,
                "target_agent": agent_type,
                "context": context
            },
            debug=True
        )

        # Attempt recovery with fallback agent
        if agent_type != "general-purpose":
            recovery = await Task(
                subagent_type="general-purpose",
                prompt=f"Recover from error and complete: {task_description}",
                context={"error_analysis": error_analysis}
            )
            return recovery

        raise e
```

## Context Management Strategies

### Minimal Context Principle

```python
# Load only essential context
essential_context = {
    "spec_id": "SPEC-001",
    "primary_requirements": reqs[:3],  # Only top 3 requirements
    "technical_constraints": constraints[:2]  # Only major constraints
}

# Avoid including entire files or large datasets
# INSTEAD: Include summaries or references
```

### Context Compression

```python
# Compress large contexts before passing
def compress_context(large_context):
    return {
        "summary": large_context.get("summary", ""),
        "key_points": large_context.get("key_points", [])[:5],
        "requirements": large_context.get("requirements", [])[:3],
        "file_references": large_context.get("files", [])[:10]
    }
```

### Context Caching

```python
# Cache frequently used context
context_cache = {}

async def get_cached_context(cache_key, context_loader):
    if cache_key not in context_cache:
        context_cache[cache_key] = await context_loader()
    return context_cache[cache_key]
```

## Agent Coordination Patterns

### Supervisor Pattern

```python
# Supervisor agent coordinates multiple specialists
async def supervisor_workflow(project_requirements):
    # Supervisor breaks down work
    work_breakdown = await Task(
        subagent_type="plan",
        prompt="Break down project into manageable tasks",
        context={"requirements": project_requirements}
    )

    # Execute tasks in dependency order
    results = {}
    for task in work_breakdown.tasks:
        if task.dependencies_met(results):
            agent = select_agent_for_task(task)
            result = await Task(
                subagent_type=agent,
                prompt=task.description,
                context={"task": task, "previous_results": results}
            )
            results[task.id] = result

    return results
```

### Specialist Collaboration

```python
# Multiple specialists collaborate on complex problem
async def specialist_collaboration(problem_description):
    # Parallel analysis by different specialists
    specialists = ["security-expert", "performance-engineer", "ui-ux-expert"]

    analyses = await asyncio.gather(*[
        Task(
            subagent_type=specialist,
            prompt=f"Analyze from {specialist} perspective: {problem_description}"
        )
        for specialist in specialists
    ])

    # Synthesize recommendations
    synthesis = await Task(
        subagent_type="project-manager",
        prompt="Synthesize specialist recommendations into unified plan",
        context={"specialist_analyses": analyses}
    )

    return synthesis
```

## Debugging and Monitoring

### Delegation Monitoring

```python
# Track delegation performance
delegation_metrics = {
    "total_delegations": 0,
    "successful_delegations": 0,
    "failed_delegations": 0,
    "average_response_time": 0
}

async def monitored_delegation(agent_type, prompt, context=None):
    start_time = time.time()
    delegation_metrics["total_delegations"] += 1

    try:
        result = await Task(
            subagent_type=agent_type,
            prompt=prompt,
            context=context
        )

        delegation_metrics["successful_delegations"] += 1
        response_time = time.time() - start_time
        delegation_metrics["average_response_time"] = (
            (delegation_metrics["average_response_time"] * (delegation_metrics["successful_delegations"] - 1) + response_time) /
            delegation_metrics["successful_delegations"]
        )

        return result

    except Exception as e:
        delegation_metrics["failed_delegations"] += 1
        raise e
```

### Debug Helper Integration

```python
# Use debug helper for troubleshooting
async def debug_delegation(agent_type, prompt, context, last_error=None):
    debug_info = await Task(
        subagent_type="debug-helper",
        prompt="Analyze delegation failure and provide solution",
        context={
            "target_agent": agent_type,
            "prompt": prompt,
            "context": context,
            "last_error": last_error,
            "system_state": "get_current_state()"
        },
        debug=True
    )

    return debug_info
```

## Best Practices

### Delegation Guidelines

1. **Specific Prompts**: Provide clear, specific task descriptions
2. **Appropriate Context**: Include only relevant information
3. **Agent Specialization**: Match tasks to agent expertise
4. **Error Handling**: Always include error handling and recovery
5. **Result Validation**: Validate and process agent outputs

### Anti-Patterns

**Avoid These**:
```python
# BAD: Vague delegation
result = await Task(
    subagent_type="backend-expert",
    prompt="fix the backend"
)

# BAD: Too much context
result = await Task(
    subagent_type="frontend-expert",
    prompt="implement this",
    context={"entire_codebase": load_all_files()}  # Too much!
)

# BAD: No error handling
result = await Task(
    subagent_type="api-designer",
    prompt="design API"
    # No error handling!
)
```

**Use These**:
```python
# GOOD: Specific delegation
result = await Task(
    subagent_type="backend-expert",
    prompt="Implement JWT authentication endpoint with refresh token support",
    context={"api_version": "v2", "security_requirements": reqs}
)

# GOOD: Appropriate context
result = await Task(
    subagent_type="frontend-expert",
    prompt="Implement login form component",
    context={"design_spec": login_design, "validation_rules": rules}
)

# GOOD: With error handling
try:
    result = await Task(
        subagent_type="api-designer",
        prompt="Design user authentication API"
    )
except Exception as e:
    debug_result = await debug_delegation("api-designer", prompt, context, e)
```

## Performance Optimization

### Batch Delegation

```python
# Process multiple similar tasks efficiently
async def batch_delegation(tasks, agent_type):
    """Delegate multiple similar tasks to same agent"""
    batch_prompt = "Process the following tasks:\n"
    for i, task in enumerate(tasks):
        batch_prompt += f"{i+1}. {task}\n"

    results = await Task(
        subagent_type=agent_type,
        prompt=batch_prompt,
        context={"task_count": len(tasks)}
    )

    return results
```

### Result Caching

```python
# Cache delegation results for repeated tasks
result_cache = {}

async def cached_delegation(cache_key, agent_type, prompt, context=None):
    if cache_key in result_cache:
        return result_cache[cache_key]

    result = await Task(
        subagent_type=agent_type,
        prompt=prompt,
        context=context
    )

    result_cache[cache_key] = result
    return result
```

These patterns ensure efficient, reliable, and maintainable agent delegation throughout the MoAI-ADK system.