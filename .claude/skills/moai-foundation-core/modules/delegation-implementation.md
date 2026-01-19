# Delegation Implementation Patterns

Purpose: Detailed implementation patterns for agent delegation including context optimization and workflow examples.

Version: 1.0.0
Last Updated: 2026-01-06
Parent: [delegation-patterns.md](delegation-patterns.md)

---

## Context Passing Optimization

### Efficient Context Structure

```python
class ContextManager:
    """Optimize context passing between agents."""

    def __init__(self):
        self.max_context_size = 50_000  # 50K tokens
        self.optimal_size = 30_000       # 30K tokens

    def prepare_context(self, full_data: dict, agent_type: str) -> dict:
        """Prepare optimized context for specific agent."""

        context_requirements = {
            "code-backend": ["spec_id", "api_design", "database_schema"],
            "code-frontend": ["spec_id", "api_endpoints", "ui_requirements"],
            "security-expert": ["spec_id", "threat_model", "security_requirements"],
            "core-quality": ["spec_id", "code_summary", "test_strategy"],
            "workflow-docs": ["spec_id", "api_spec", "code_summary"]
        }

        required_fields = context_requirements.get(agent_type, [])
        optimized_context = {
            field: full_data.get(field)
            for field in required_fields
            if field in full_data
        }

        if "code_summary" in required_fields:
            optimized_context["code_summary"] = self._compress_code_summary(
                full_data.get("full_code", "")
            )

        estimated_tokens = self._estimate_tokens(optimized_context)
        if estimated_tokens > self.optimal_size:
            optimized_context = self._further_compress(optimized_context)

        return optimized_context

    def _compress_code_summary(self, full_code: str) -> dict:
        """Compress code to summary (functions, classes, key logic)."""
        return {
            "functions": extract_function_signatures(full_code),
            "classes": extract_class_definitions(full_code),
            "key_logic": extract_main_flow(full_code)
        }

    def _estimate_tokens(self, context: dict) -> int:
        """Estimate token count of context."""
        import json
        json_str = json.dumps(context)
        return len(json_str) // 4  # Rough estimate
```

Usage:
```python
context_manager = ContextManager()

full_data = {
    "spec_id": "SPEC-001",
    "full_code": "... 50KB of code ...",
    "api_design": {...},
    "database_schema": {...}
}

backend_context = context_manager.prepare_context(full_data, "code-backend")
# Result: Only spec_id, api_design, database_schema (~25K tokens)

result = await Task(
    subagent_type="code-backend",
    prompt="Implement backend",
    context=backend_context
)
```

---

## Full Sequential Workflow

```python
async def implement_feature_sequential(feature_description: str):
    """Complete sequential workflow with context passing."""

    # Phase 1: SPEC Generation
    spec_result = await Task(
        subagent_type="workflow-spec",
        prompt=f"Generate SPEC for: {feature_description}",
        context={
            "feature": feature_description,
            "requirements": ["TRUST 5 compliance", "≥85% coverage"]
        }
    )

    execute_clear()

    # Phase 2: API Design
    api_result = await Task(
        subagent_type="api-designer",
        prompt="Design REST API for feature",
        context={
            "spec_id": spec_result.spec_id,
            "requirements": spec_result.requirements,
            "constraints": ["RESTful", "JSON", "OpenAPI 3.1"]
        }
    )

    # Phase 3: Backend Implementation
    backend_result = await Task(
        subagent_type="code-backend",
        prompt="Implement backend with DDD",
        context={
            "spec_id": spec_result.spec_id,
            "api_design": api_result.openapi_spec,
            "database_schema": api_result.database_schema
        }
    )

    # Phase 4: Frontend Implementation
    frontend_result = await Task(
        subagent_type="code-frontend",
        prompt="Implement UI components",
        context={
            "spec_id": spec_result.spec_id,
            "api_endpoints": api_result.endpoints,
            "ui_requirements": spec_result.ui_requirements
        }
    )

    # Phase 5: Integration Testing
    integration_result = await Task(
        subagent_type="core-quality",
        prompt="Run integration tests",
        context={
            "spec_id": spec_result.spec_id,
            "backend_endpoints": backend_result.endpoints,
            "frontend_components": frontend_result.components
        }
    )

    # Phase 6: Documentation
    docs_result = await Task(
        subagent_type="workflow-docs",
        prompt="Generate comprehensive documentation",
        context={
            "spec_id": spec_result.spec_id,
            "api_spec": api_result.openapi_spec,
            "backend_code": backend_result.code_summary,
            "frontend_code": frontend_result.code_summary,
            "test_results": integration_result.test_report
        }
    )

    return {
        "spec": spec_result,
        "api": api_result,
        "backend": backend_result,
        "frontend": frontend_result,
        "tests": integration_result,
        "docs": docs_result
    }
```

---

## Token Management in Sequential Flow

```python
def sequential_with_token_management():
    """Sequential flow with strategic /clear execution."""

    # Phase 1: Heavy context (SPEC generation)
    spec = Task(subagent_type="workflow-spec", ...)  # ~30K tokens
    execute_clear()  # Save 45-50K tokens

    # Phase 2: Fresh context (implementation)
    impl = Task(
        subagent_type="workflow-ddd",
        context={"spec_id": spec.id}  # Minimal context
    )  # ~80K tokens

    # Phase 3: Final phase
    docs = Task(
        subagent_type="workflow-docs",
        context={"spec_id": spec.id, "summary": impl.summary}
    )  # ~25K tokens

    # Total: ~135K (within 200K budget)
```

---

## Complex Conditional Logic

```python
async def advanced_conditional_routing(request: dict):
    """Multi-criteria conditional routing."""

    analysis = await Task(
        subagent_type="plan",
        prompt="Analyze request complexity",
        context=request
    )

    if analysis.complexity == "high" and analysis.security_critical:
        return await sequential_secure_workflow(analysis)

    elif analysis.complexity == "high":
        return await parallel_workflow(analysis)

    elif analysis.complexity == "low":
        return await single_agent_workflow(analysis)

    elif analysis.performance_critical:
        return await performance_optimized_workflow(analysis)

    else:
        return await standard_workflow(analysis)

async def sequential_secure_workflow(analysis):
    """High-complexity security workflow."""
    security_review = await Task(
        subagent_type="security-expert",
        prompt="Security architecture review"
    )

    implementation = await Task(
        subagent_type="code-backend",
        prompt="Implement with security controls",
        context={"security_requirements": security_review}
    )

    penetration_test = await Task(
        subagent_type="security-expert",
        prompt="Penetration testing",
        context={"implementation": implementation}
    )

    return {
        "security_review": security_review,
        "implementation": implementation,
        "penetration_test": penetration_test
    }
```

---

## Works Well With

- [delegation-patterns.md](delegation-patterns.md) - Overview
- [delegation-advanced.md](delegation-advanced.md) - Error handling

---

Version: 1.0.0
Last Updated: 2026-01-06
