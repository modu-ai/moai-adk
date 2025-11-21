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

---

## Skill-Enhanced Delegation Patterns

### Pattern 1: Basic Agent Delegation with Skill Loading

**Scenario**: Delegate a task to an agent that will load relevant skills

**Pattern**:
```python
Task(
  subagent_type="backend-expert",
  description="Implement user authentication API",
  prompt="""
  Implement a secure user authentication REST API endpoint.

  Load relevant skills:
  - moai-domain-backend: Backend architecture patterns
  - moai-security-auth: Authentication best practices
  - moai-security-api: API security patterns
  - moai-lang-python: FastAPI implementation

  Requirements:
  - JWT token generation
  - Password hashing with bcrypt
  - OWASP compliance
  - 85%+ test coverage
  """
)
```

**When to Use**: Standard agent delegation with multiple skill requirements

---

### Pattern 2: Multi-Domain Task with Cross-Domain Skills

**Scenario**: Task requires skills from multiple domains

**Pattern**:
```python
Task(
  subagent_type="backend-expert",
  description="Implement secure backend with performance optimization",
  prompt="""
  Implement backend system with security and performance requirements.

  Load domain skills:
  - moai-domain-backend: Core backend patterns
  - moai-security-api: API security patterns
  - moai-essentials-perf: Performance optimization
  - moai-domain-database: Database patterns
  - moai-lang-python: Python implementation

  Cross-domain combination for comprehensive solution.
  Target: <100ms p95 latency, OWASP compliant, 90%+ test coverage
  """
)
```

**When to Use**: Complex tasks spanning multiple technical domains

---

### Pattern 3: Quality-First Implementation with TRUST 5

**Scenario**: Task requires TRUST 5 quality gate compliance

**Pattern**:
```python
Task(
  subagent_type="tdd-implementer",
  description="Implement feature with TRUST 5 compliance",
  prompt="""
  Implement feature using RED-GREEN-REFACTOR TDD cycle.

  Load TRUST 5 skills:
  - moai-foundation-trust: TRUST 5 principles framework
  - moai-essentials-review: Code quality review patterns
  - moai-core-code-reviewer: Review orchestration
  - moai-domain-testing: Testing strategies
  - moai-core-dev-guide: TDD workflow patterns

  Ensure all 5 TRUST principles are satisfied:
  1. Test-first (test coverage 85%+)
  2. Readable (clear code, good naming)
  3. Unified (consistent patterns)
  4. Secured (OWASP compliance)
  5. Trackable (git history, test tracking)
  """
)
```

**When to Use**: High-quality implementations requiring TRUST 5 validation

---

### Pattern 4: Design & Implementation Chain with Skill Progression

**Scenario**: Design phase followed by implementation phase with different skills

**Pattern**:
```python
# Phase 1: Design with api-designer
design = await Task(
  subagent_type="api-designer",
  description="Design REST API architecture",
  prompt="""
  Design secure REST API with OpenAPI specification.

  Load design skills:
  - moai-domain-web-api: REST API design patterns
  - moai-security-api: Security patterns
  - moai-foundation-ears: Requirements analysis

  Generate OpenAPI 3.0 specification with all endpoints and security schemes.
  """
)

# Phase 2: Implementation with backend-expert (using design output)
implementation = await Task(
  subagent_type="backend-expert",
  description="Implement REST API from design",
  prompt=f"""
  Implement REST API based on the following design:
  {design.openapi_spec}

  Load implementation skills:
  - moai-domain-backend: Backend implementation patterns
  - moai-security-auth: Authentication implementation
  - moai-essentials-perf: Performance optimization
  - moai-domain-testing: Testing strategies
  - moai-lang-python: FastAPI implementation

  Follow design specification exactly, implement all endpoints with security.
  """
)
```

**When to Use**: Large features requiring design then implementation phases

---

### Pattern 5: Security-Enhanced Development with Parallel Validation

**Scenario**: Implementation with concurrent security validation

**Pattern**:
```python
# Main implementation
implementation = await Task(
  subagent_type="backend-expert",
  description="Implement backend service",
  prompt="""
  Implement authentication backend service.

  Load skills:
  - moai-domain-backend
  - moai-security-api
  - moai-lang-python
  """
)

# Parallel: Security validation
security_review = await Task(
  subagent_type="security-expert",
  description="Validate implementation for security",
  prompt=f"""
  Review implementation for security compliance:
  {implementation.code}

  Load security skills:
  - moai-domain-security: Security patterns validation
  - moai-security-owasp: OWASP top 10 compliance
  - moai-security-auth: Authentication security
  - moai-security-encryption: Encryption standards

  Generate security report with risk ratings and remediation priorities.
  """
)
```

**When to Use**: High-security requirements or production deployments

---

## Agent Skill Loading Considerations

### Token Efficiency

**Challenge**: Loading skills adds context cost
**Solution**:
- Load only skills needed for specific task
- Combine multiple related skills efficiently
- Agent execution token budget: 180K max for implementation

### Skill Combination Guidelines

**Recommended Order for Loading Multiple Skills**:
1. **Foundation skill** (moai-foundation-[domain]) - Core principles
2. **Domain skill** (moai-domain-[domain]) - Domain expertise
3. **Enhancement skills** (moai-essentials-*) - Quality, performance, debugging
4. **Language/Tech skills** (moai-lang-*, moai-[tool]-*) - Implementation
5. **Context7 references** (if needed for latest documentation)

**Example**:
```python
# Optimal skill loading order
prompt = """
Implement authentication system.

Load skills in order:
1. moai-foundation-trust (TRUST principles)
2. moai-domain-backend (backend architecture)
3. moai-essentials-review (code quality)
4. moai-security-auth (authentication patterns)
5. moai-lang-python (Python implementation)

Context7: /tiangolo/fastapi (latest FastAPI documentation)
"""
```

### Multi-Skill Loading Pattern

```python
# Complete skill stack for backend implementation
Task(
  subagent_type="backend-expert",
  prompt="""
  Implement authentication API with complete skill coverage.

  Foundation Layer:
  - moai-foundation-trust (quality principles)
  - moai-foundation-specs (SPEC traceability)

  Domain Layer:
  - moai-domain-backend (backend patterns)
  - moai-domain-database (database integration)

  Security Layer:
  - moai-security-api (API security)
  - moai-security-auth (authentication)

  Implementation Layer:
  - moai-lang-python (Python/FastAPI)
  - moai-essentials-perf (performance optimization)

  Quality Layer:
  - moai-essentials-review (code review)
  - moai-domain-testing (testing strategies)

  Comprehensive implementation with full skill coverage.
  """
)
```

---

These patterns ensure efficient, reliable, and maintainable agent delegation throughout the MoAI-ADK system.