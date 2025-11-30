# Delegation Patterns - Agent Orchestration

Purpose: Task delegation strategies to specialized agents, ensuring efficient workflow orchestration without direct execution.

Version: 1.0.0
Last Updated: 2025-11-25

---

## Quick Reference (30 seconds)

Core Principle: Alfred NEVER executes directly. All work via `Task()` delegation to specialized agents.

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

Implementation Example:

```python
# Full sequential workflow
async def implement_feature_sequential(feature_description: str):
 """Sequential workflow with context passing."""
 
 # Phase 1: SPEC Generation
 spec_result = await Task(
 subagent_type="workflow-spec",
 prompt=f"Generate SPEC for: {feature_description}",
 context={
 "feature": feature_description,
 "requirements": ["TRUST 5 compliance", "≥85% coverage"]
 }
 )
 
 # Execute /clear to save tokens
 execute_clear()
 
 # Phase 2: API Design (depends on SPEC)
 api_result = await Task(
 subagent_type="api-designer",
 prompt="Design REST API for feature",
 context={
 "spec_id": spec_result.spec_id,
 "requirements": spec_result.requirements,
 "constraints": ["RESTful", "JSON", "OpenAPI 3.1"]
 }
 )
 
 # Phase 3: Backend Implementation (depends on API design + SPEC)
 backend_result = await Task(
 subagent_type="code-backend",
 prompt="Implement backend with TDD",
 context={
 "spec_id": spec_result.spec_id,
 "api_design": api_result.openapi_spec,
 "database_schema": api_result.database_schema
 }
 )
 
 # Phase 4: Frontend Implementation (depends on API design)
 frontend_result = await Task(
 subagent_type="code-frontend",
 prompt="Implement UI components",
 context={
 "spec_id": spec_result.spec_id,
 "api_endpoints": api_result.endpoints,
 "ui_requirements": spec_result.ui_requirements
 }
 )
 
 # Phase 5: Integration Testing (depends on all implementations)
 integration_result = await Task(
 subagent_type="core-quality",
 prompt="Run integration tests",
 context={
 "spec_id": spec_result.spec_id,
 "backend_endpoints": backend_result.endpoints,
 "frontend_components": frontend_result.components
 }
 )
 
 # Phase 6: Documentation (depends on everything)
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
 
 # Phase 7: Quality Gate (validates everything)
 quality_result = await Task(
 subagent_type="core-quality",
 prompt="Validate TRUST 5 compliance",
 context={
 "spec_id": spec_result.spec_id,
 "implementation": {
 "backend": backend_result,
 "frontend": frontend_result
 },
 "tests": integration_result,
 "documentation": docs_result
 }
 )
 
 return {
 "spec": spec_result,
 "api": api_result,
 "backend": backend_result,
 "frontend": frontend_result,
 "tests": integration_result,
 "docs": docs_result,
 "quality": quality_result
 }
```

Token Management in Sequential Flow:

```python
def sequential_with_token_management():
 """Sequential flow with strategic /clear execution."""
 
 # Phase 1: Heavy context (SPEC generation)
 spec = Task(subagent_type="workflow-spec", ...) # ~30K tokens
 execute_clear() # Save 45-50K tokens
 
 # Phase 2: Fresh context (implementation)
 impl = Task(
 subagent_type="workflow-tdd",
 context={"spec_id": spec.id} # Minimal context
 ) # ~80K tokens
 
 # Phase 3: Final phase
 docs = Task(
 subagent_type="workflow-docs",
 context={"spec_id": spec.id, "summary": impl.summary}
 ) # ~25K tokens
 
 # Total: ~135K (within 200K budget)
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

Implementation Example:

```python
async def implement_feature_parallel(spec_id: str):
 """Parallel workflow for independent tasks."""
 
 # Execute all independent tasks simultaneously
 results = await Promise.all([
 # Backend implementation
 Task(
 subagent_type="code-backend",
 prompt=f"Implement backend for {spec_id}",
 context={"spec_id": spec_id, "focus": "API endpoints"}
 ),
 
 # Frontend implementation
 Task(
 subagent_type="code-frontend",
 prompt=f"Implement UI for {spec_id}",
 context={"spec_id": spec_id, "focus": "Components"}
 ),
 
 # Database schema
 Task(
 subagent_type="data-database",
 prompt=f"Design database for {spec_id}",
 context={"spec_id": spec_id, "focus": "Schema"}
 ),
 
 # Documentation
 Task(
 subagent_type="workflow-docs",
 prompt=f"Generate docs for {spec_id}",
 context={"spec_id": spec_id, "focus": "API docs"}
 )
 ])
 
 backend, frontend, database, docs = results
 
 # Integration step (sequential, depends on parallel results)
 integration = await Task(
 subagent_type="core-quality",
 prompt="Run integration tests",
 context={
 "spec_id": spec_id,
 "backend": backend.summary,
 "frontend": frontend.summary,
 "database": database.schema,
 "docs": docs.summary
 }
 )
 
 return {
 "backend": backend,
 "frontend": frontend,
 "database": database,
 "docs": docs,
 "integration": integration
 }
```

Parallel Execution Benefits:

| Aspect | Sequential | Parallel | Improvement |
|--------|-----------|----------|-------------|
| Execution Time | 45 min | 15 min | 3x faster |
| Token Sessions | Shared 200K | 4x 200K each | 4x capacity |
| Context Isolation | Shared context | Independent | Cleaner |
| Error Impact | Blocks all | Isolated | Resilient |

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
 → Documentation → workflow-docs
```

Implementation Example:

```python
async def handle_issue_conditional(issue_description: str):
 """Conditional routing based on issue analysis."""
 
 # Phase 1: Analyze issue type
 analysis = await Task(
 subagent_type="support-debug",
 prompt=f"Analyze issue: {issue_description}",
 context={"focus": "classification"}
 )
 
 # Phase 2: Route to appropriate specialist
 if analysis.category == "security":
 return await Task(
 subagent_type="security-expert",
 prompt="Analyze and fix security issue",
 context={
 "issue": issue_description,
 "analysis": analysis.details,
 "owasp_category": analysis.owasp_match
 }
 )
 
 elif analysis.category == "performance":
 return await Task(
 subagent_type="performance-engineer",
 prompt="Optimize performance issue",
 context={
 "issue": issue_description,
 "analysis": analysis.details,
 "bottleneck": analysis.bottleneck_location
 }
 )
 
 elif analysis.category == "quality":
 return await Task(
 subagent_type="core-quality",
 prompt="Validate and improve quality",
 context={
 "issue": issue_description,
 "analysis": analysis.details,
 "trust5_violations": analysis.violations
 }
 )
 
 elif analysis.category == "bug":
 return await Task(
 subagent_type="support-debug",
 prompt="Debug and fix issue",
 context={
 "issue": issue_description,
 "analysis": analysis.details,
 "stack_trace": analysis.stack_trace
 }
 )
 
 else: # documentation, feature request, etc.
 return await Task(
 subagent_type="workflow-docs",
 prompt="Handle documentation request",
 context={
 "issue": issue_description,
 "analysis": analysis.details
 }
 )
```

Complex Conditional Logic:

```python
async def advanced_conditional_routing(request: dict):
 """Multi-criteria conditional routing."""
 
 # Phase 1: Multi-dimensional analysis
 analysis = await Task(
 subagent_type="plan",
 prompt="Analyze request complexity",
 context=request
 )
 
 # Phase 2: Route based on multiple factors
 
 # High complexity + Security-critical
 if analysis.complexity == "high" and analysis.security_critical:
 return await sequential_secure_workflow(analysis)
 
 # High complexity + Low security
 elif analysis.complexity == "high":
 return await parallel_workflow(analysis)
 
 # Low complexity
 elif analysis.complexity == "low":
 return await single_agent_workflow(analysis)
 
 # Medium complexity + Performance-critical
 elif analysis.performance_critical:
 return await performance_optimized_workflow(analysis)
 
 # Default: Standard workflow
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

## Advanced Implementation (10+ minutes)

### Context Passing Optimization

Efficient Context Structure:

```python
class ContextManager:
 """Optimize context passing between agents."""
 
 def __init__(self):
 self.max_context_size = 50_000 # 50K tokens
 self.optimal_size = 30_000 # 30K tokens
 
 def prepare_context(self, full_data: dict, agent_type: str) -> dict:
 """Prepare optimized context for specific agent."""
 
 # Agent-specific context requirements
 context_requirements = {
 "code-backend": ["spec_id", "api_design", "database_schema"],
 "code-frontend": ["spec_id", "api_endpoints", "ui_requirements"],
 "security-expert": ["spec_id", "threat_model", "security_requirements"],
 "core-quality": ["spec_id", "code_summary", "test_strategy"],
 "workflow-docs": ["spec_id", "api_spec", "code_summary"]
 }
 
 # Extract only required fields
 required_fields = context_requirements.get(agent_type, [])
 optimized_context = {
 field: full_data.get(field)
 for field in required_fields
 if field in full_data
 }
 
 # Add compressed summaries for large data
 if "code_summary" in required_fields:
 optimized_context["code_summary"] = self._compress_code_summary(
 full_data.get("full_code", "")
 )
 
 # Validate size
 estimated_tokens = self._estimate_tokens(optimized_context)
 if estimated_tokens > self.optimal_size:
 optimized_context = self._further_compress(optimized_context)
 
 return optimized_context
 
 def _compress_code_summary(self, full_code: str) -> str:
 """Compress code to summary (functions, classes, key logic)."""
 summary = {
 "functions": extract_function_signatures(full_code),
 "classes": extract_class_definitions(full_code),
 "key_logic": extract_main_flow(full_code)
 }
 return summary
 
 def _estimate_tokens(self, context: dict) -> int:
 """Estimate token count of context."""
 import json
 json_str = json.dumps(context)
 # Rough estimate: 1 token ≈ 4 characters
 return len(json_str) // 4

# Usage
context_manager = ContextManager()

full_data = {
 "spec_id": "SPEC-001",
 "full_code": "... 50KB of code ...",
 "api_design": {...},
 "database_schema": {...}
}

# Prepare optimized context for code-backend
backend_context = context_manager.prepare_context(full_data, "code-backend")
# Result: Only spec_id, api_design, database_schema (no full_code)

backend_result = await Task(
 subagent_type="code-backend",
 prompt="Implement backend",
 context=backend_context # Optimized to ~25K tokens
)
```

### Error Handling and Recovery

Resilient Delegation Pattern:

```python
from typing import Optional
import asyncio

class ResilientDelegation:
 """Handle delegation failures with retry and fallback."""
 
 def __init__(self):
 self.max_retries = 3
 self.retry_delay = 5 # seconds
 
 async def delegate_with_retry(
 self,
 agent_type: str,
 prompt: str,
 context: dict,
 fallback_agent: Optional[str] = None
 ):
 """Delegate with automatic retry and fallback."""
 
 for attempt in range(self.max_retries):
 try:
 result = await Task(
 subagent_type=agent_type,
 prompt=prompt,
 context=context
 )
 return result
 
 except AgentExecutionError as e:
 if attempt < self.max_retries - 1:
 # Retry with exponential backoff
 delay = self.retry_delay * (2 attempt)
 await asyncio.sleep(delay)
 continue
 
 # Final attempt failed
 if fallback_agent:
 # Try fallback agent
 return await self._fallback_delegation(
 fallback_agent,
 prompt,
 context,
 original_error=e
 )
 
 # No fallback, raise error
 raise
 
 except ContextTooLargeError as e:
 # Context exceeds limit, compress and retry
 compressed_context = self._compress_context(context)
 return await Task(
 subagent_type=agent_type,
 prompt=prompt,
 context=compressed_context
 )
 
 async def _fallback_delegation(
 self,
 fallback_agent: str,
 prompt: str,
 context: dict,
 original_error: Exception
 ):
 """Execute fallback delegation."""
 
 # Add error context
 fallback_context = {
 context,
 "original_error": str(original_error),
 "fallback_mode": True
 }
 
 return await Task(
 subagent_type=fallback_agent,
 prompt=f"[FALLBACK] {prompt}",
 context=fallback_context
 )

# Usage
delegator = ResilientDelegation()

result = await delegator.delegate_with_retry(
 agent_type="code-backend",
 prompt="Implement complex feature",
 context=large_context,
 fallback_agent="support-debug" # Fallback if code-backend fails
)
```

### Hybrid Delegation Patterns

Sequential + Parallel Combination:

```python
async def hybrid_workflow(spec_id: str):
 """Combine sequential and parallel patterns."""
 
 # Phase 1: Sequential (SPEC → Design)
 spec = await Task(
 subagent_type="workflow-spec",
 prompt=f"Generate SPEC {spec_id}"
 )
 
 design = await Task(
 subagent_type="api-designer",
 prompt="Design API",
 context={"spec_id": spec.id}
 )
 
 execute_clear()
 
 # Phase 2: Parallel (Implementation)
 impl_results = await Promise.all([
 Task(
 subagent_type="code-backend",
 prompt="Backend",
 context={"spec_id": spec.id, "api": design}
 ),
 Task(
 subagent_type="code-frontend",
 prompt="Frontend",
 context={"spec_id": spec.id, "api": design}
 ),
 Task(
 subagent_type="data-database",
 prompt="Database",
 context={"spec_id": spec.id, "api": design}
 )
 ])
 
 backend, frontend, database = impl_results
 
 # Phase 3: Sequential (Testing → QA)
 tests = await Task(
 subagent_type="core-quality",
 prompt="Integration tests",
 context={
 "spec_id": spec.id,
 "backend": backend.summary,
 "frontend": frontend.summary,
 "database": database.summary
 }
 )
 
 qa = await Task(
 subagent_type="core-quality",
 prompt="Quality validation",
 context={
 "spec_id": spec.id,
 "tests": tests.results,
 "coverage": tests.coverage
 }
 )
 
 return {
 "spec": spec,
 "design": design,
 "backend": backend,
 "frontend": frontend,
 "database": database,
 "tests": tests,
 "qa": qa
 }
```

Conditional + Parallel Combination:

```python
async def conditional_parallel_workflow(requests: list):
 """Route multiple requests in parallel based on analysis."""
 
 # Phase 1: Parallel analysis
 analyses = await Promise.all([
 Task(
 subagent_type="plan",
 prompt=f"Analyze request",
 context={"request": req}
 )
 for req in requests
 ])
 
 # Phase 2: Conditional routing (grouped by type)
 security_tasks = []
 feature_tasks = []
 bug_tasks = []
 
 for analysis in analyses:
 if analysis.category == "security":
 security_tasks.append(
 Task(
 subagent_type="security-expert",
 prompt="Handle security issue",
 context={"analysis": analysis}
 )
 )
 elif analysis.category == "feature":
 feature_tasks.append(
 Task(
 subagent_type="code-backend",
 prompt="Implement feature",
 context={"analysis": analysis}
 )
 )
 elif analysis.category == "bug":
 bug_tasks.append(
 Task(
 subagent_type="support-debug",
 prompt="Debug issue",
 context={"analysis": analysis}
 )
 )
 
 # Phase 3: Parallel execution by category
 results = await Promise.all([
 *security_tasks,
 *feature_tasks,
 *bug_tasks
 ])
 
 return results
```

---

## Works Well With

Agents (Delegation Targets):
- workflow-spec - SPEC generation
- workflow-tdd - TDD implementation
- code-backend - Backend development
- code-frontend - Frontend development
- security-expert - Security analysis
- core-quality - Quality validation
- workflow-docs - Documentation
- support-debug - Issue analysis

Skills:
- moai-foundation-token-optimization - Context management
- moai-core-agent-factory - Agent creation
- moai-cc-configuration - Config-based routing

Foundation Modules:
- [Agents Reference](agents-reference.md) - 26-agent catalog with delegation guidelines
- [Token Optimization](token-optimization.md) - Context passing strategies
- [Execution Rules](execution-rules.md) - Security constraints for delegation

---

Version: 1.0.0
Last Updated: 2025-11-25
Status: Production Ready
