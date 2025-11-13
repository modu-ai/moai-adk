# moai-alfred-agent-guide: Examples

**Practical Agent Selection and Orchestration Examples**

## Feature Implementation Examples

### Example 1: User Authentication Feature

**Request**: "Add user authentication with email verification"

**Analysis & Agent Selection**:
```
Category: Domain-Specific + Code Implementation
Domain: Backend (auth logic) + Frontend (UI forms)
Complexity: HIGH (security implications, multi-step flow)
Testing: REQUIRED (security-critical)
```

**Agent Orchestration**:
```python
# Phase 1: Architecture & Security Design
Task(
    subagent_type="backend-expert",
    model_override="sonnet",  # Complex security decisions
    description="Design authentication system architecture",
    prompt="""Design secure authentication with:
- Email/password registration
- Email verification flow
- JWT token management
- Password hashing (bcrypt/argon2)
- Rate limiting (prevent brute force)

Coordinate with security-expert for threat modeling.
Output: Architecture diagram, API endpoints, security checklist."""
)

# Phase 2: Security Validation
Task(
    subagent_type="security-expert", 
    model_override="sonnet",  # Threat modeling required
    description="Validate authentication design security",
    prompt="""Review authentication design for:
- OWASP Top 10 compliance
- Token expiration strategies
- CSRF/XSS protections
- SQL injection prevention
- Secure session management

Output: Security audit report, required changes."""
)

# Phase 3: TDD Implementation
Task(
    subagent_type="tdd-implementer",
    model_override="haiku",  # Pattern-based code generation
    description="Implement authentication with TDD",
    prompt="""Execute RED-GREEN-REFACTOR:
- RED: Write failing tests for each endpoint
- GREEN: Implement minimal passing code  
- REFACTOR: Optimize with security patterns

Follow backend-expert architecture and security-expert requirements."""
)

# Phase 4: Test Coverage Validation
Task(
    subagent_type="test-engineer",
    model_override="haiku",  # Standardized testing
    description="Validate authentication test coverage",
    prompt="""Verify authentication has:
- Unit tests: 90%+ coverage
- Integration tests: API endpoints
- Security tests: Auth bypass attempts
- Edge cases: Invalid tokens, expired sessions"""
)

# Phase 5: Frontend Integration
Task(
    subagent_type="frontend-expert", 
    model_override="haiku",  # UI component patterns
    description="Create authentication UI components",
    prompt="""Build frontend auth components:
- Registration form with validation
- Email verification status page
- Login form with error handling
- Protected route guards"""
)
```

**Cost Analysis**:
- Sonnet agents (2): ~$6 (architecture + security)
- Haiku agents (3): ~$1.50 (implementation + testing + UI)
- **Total**: ~$7.50 (vs ~$15 for all-Sonnet approach)

---

### Example 2: Bug Fix Request

**Request**: "Fix memory leak in background job processor"

**Analysis & Agent Selection**:
```
Category: Debugging + Code Fix
Complexity: HIGH (requires investigation, profiling)
Domain: Backend (job processing)
```

**Agent Orchestration**:
```python
# Phase 1: Root Cause Analysis
Task(
    subagent_type="debug-helper",
    model_override="sonnet",  # Complex troubleshooting
    description="Diagnose memory leak in job processor",
    prompt="""Investigate memory leak:
- Review code for common leak patterns
- Analyze logs for growth patterns  
- Identify suspected code paths

Coordinate with performance-engineer for profiling.
Output: Hypothesis, suspected files, investigation plan."""
)

# Phase 2: Performance Profiling
Task(
    subagent_type="performance-engineer",
    model_override="sonnet",  # Profiling analysis
    description="Profile job processor memory usage",
    prompt="""Profile application with:
- Memory profiler (memory_profiler, tracemalloc)
- Heap dump analysis
- Object retention tracking

Validate debug-helper's hypothesis with data.
Output: Profiling report, memory hotspots, leak confirmation."""
)

# Phase 3: Fix Implementation
Task(
    subagent_type="tdd-implementer",
    model_override="haiku",  # Pattern-based fix
    description="Fix memory leak with TDD approach", 
    prompt="""Fix confirmed memory leak:
- RED: Write test demonstrating leak
- GREEN: Implement fix (close connections, break refs)
- REFACTOR: Optimize resource management

Follow debug-helper and performance-engineer findings."""
)

# Phase 4: Regression Testing
Task(
    subagent_type="test-engineer",
    model_override="haiku",  # Standardized testing
    description="Validate memory leak fix and add monitoring",
    prompt="""Verify fix completeness:
- Run leak regression test
- Stress test job processor
- Add monitoring alerts

Output: Test results, monitoring configuration."""
)
```

---

### Example 3: Documentation Update

**Request**: "Update API documentation for new endpoints"

**Analysis & Agent Selection**:
```
Category: Documentation
Complexity: MEDIUM (structured content, technical writing)
Dependencies: Backend code must exist
```

**Agent Orchestration**:
```python
# Phase 1: Documentation Generation
Task(
    subagent_type="doc-syncer",
    model_override="haiku",  # Template-based documentation
    description="Generate API documentation for new endpoints",
    prompt="""Update API documentation with:
- Endpoint descriptions (HTTP method, path, purpose)
- Request/response schemas (JSON examples)
- Authentication requirements
- Error codes and messages
- Usage examples (curl, Python, JavaScript)

Output: Updated reference.md, API changelog."""
)

# Phase 2: Quality Validation
Task(
    subagent_type="qa-validator", 
    model_override="haiku",  # Rule-based validation
    description="Validate documentation consistency",
    prompt="""Check documentation for:
- All endpoints documented
- Request/response schemas match code
- Examples are executable

Output: Validation report, corrections needed."""
)
```

---

## Decision Tree Examples

### Web Feature Development

```
User Request: "Add real-time chat to the application"

Decision Tree:
┌─ Feature Type: New feature development
├─ Domain: Frontend + Backend + Database
├─ Complexity: HIGH (WebSocket, real-time, scaling)
├─ Security: MEDIUM (user authentication, data privacy)

Agent Selection:
1. plan-agent (Sonnet) - Strategic planning, task breakdown
2. backend-expert (Sonnet) - WebSocket architecture, message queuing
3. frontend-expert (Sonnet) - Real-time UI components, state management  
4. database-expert (Sonnet) - Message persistence, indexing strategy
5. security-expert (Sonnet) - Real-time security considerations
6. tdd-implementer (Haiku) - WebSocket server implementation
7. tdd-implementer (Haiku) - React chat components
8. test-engineer (Haiku) - End-to-end chat flow tests
9. devops-expert (Haiku) - WebSocket infrastructure deployment

Total: 9 agents, ~45 minutes execution time
```

### Performance Optimization

```
User Request: "Optimize slow database queries"

Decision Tree:
┌─ Task Type: Performance issue  
├─ Domain: Database + Backend
├─ Complexity: HIGH (profiling, optimization)
├─ Analysis Required: YES (identify bottlenecks)

Agent Selection:
1. debug-helper (Sonnet) - Initial problem analysis
2. performance-engineer (Sonnet) - Database profiling
3. database-expert (Sonnet) - Query optimization strategy
4. tdd-implementer (Haiku) - Implement optimizations
5. test-engineer (Haiku) - Performance regression tests

Total: 5 agents, ~25 minutes execution time
```

---

## Model Selection Examples

### When to Use Sonnet (Complex Reasoning)

**Architecture Decisions**:
```python
Task(
    subagent_type="backend-expert",
    model_override="sonnet",  # Required for complex trade-offs
    description="Choose between microservices vs monolith",
    prompt="""Analyze microservices vs monolith for:
- Team size and expertise (5 developers, mixed experience)
- Traffic patterns (10K daily active users, growth to 100K)
- Deployment complexity (DevOps capabilities: intermediate)
- Maintenance burden (long-term sustainability)
- Performance requirements (sub-second response times)

Provide recommendation with detailed trade-off analysis."""
)
```

**Security Analysis**:
```python
Task(
    subagent_type="security-expert", 
    model_override="sonnet",  # Critical for threat modeling
    description="OAuth 2.0 implementation security review",
    prompt="""Perform comprehensive OAuth 2.0 security audit:
- Analyze authorization code flow security
- Review token storage and refresh mechanisms
- Assess CSRF and XSS vulnerabilities  
- Validate state parameter implementation
- Check PKCE (Proof Key for Code Exchange) usage

Identify security gaps and provide remediation strategies."""
)
```

### When to Use Haiku (Pattern Execution)

**Standard Testing**:
```python
Task(
    subagent_type="test-engineer",
    model_override="haiku",  # Efficient for standard testing patterns
    description="Validate API test coverage",
    prompt="""Check authentication API has:
- Unit tests for all endpoints (target: 95% coverage)
- Integration tests for request/response flows
- Error handling tests (400, 401, 403, 500 responses)
- Edge case tests (invalid input, timeouts)

Generate coverage report with missing scenarios."""
)
```

**Documentation Updates**:
```python
Task(
    subagent_type="doc-syncer",
    model_override="haiku",  # Template-based content generation
    description="Update API documentation",
    prompt="""Document new endpoints:
- GET /api/v1/users/profile (user profile retrieval)
- PUT /api/v1/users/profile (profile update)
- POST /api/v1/users/avatar (avatar upload)

For each endpoint: description, request schema, response schema, authentication, examples.
Follow existing documentation format and style."""
)
```

---

## Error Handling Examples

### Agent Failure Recovery

```python
try:
    # Primary agent delegation
    result = Task(subagent_type="tdd-implementer", description="Implement feature X")
    validate_agent_output(result)
    
except AgentError as e:
    # Escalate to debug-helper for complex failures
    debug_result = Task(
        subagent_type="debug-helper",
        model_override="sonnet",
        description=f"Investigate tdd-implementer failure: {e}",
        prompt="Analyze why tdd-implementer failed and propose fix strategy."
    )
    
    # Retry with debugging insights
    result = Task(
        subagent_type="tdd-implementer", 
        description="Retry implementation with debug insights",
        prompt=f"Debug insights:\n{debug_result.analysis}\n\nImplement feature X."
    )
```

### Context Validation

```python
def validate_handoff_context(previous_result, required_fields):
    """Validate agent handoff completeness"""
    missing = []
    
    for field in required_fields:
        if field not in previous_result.output:
            missing.append(field)
    
    if missing:
        raise HandoffError(f"Incomplete handoff: missing {missing}")
    
    return True

# Usage
validate_handoff_context(plan_result, ["tasks", "dependencies", "agent_assignments"])
Task(subagent_type="tdd-implementer", prompt=f"Execute plan:\n{plan_result.output}")
```

---

## Performance Optimization Examples

### Parallel Agent Execution

```python
import asyncio

async def parallel_feature_design():
    """Execute independent agents in parallel for 3x speedup"""
    
    # Parallel design phase
    backend_task = asyncio.create_task(
        Task(subagent_type="backend-expert", description="Design API endpoints")
    )
    frontend_task = asyncio.create_task(
        Task(subagent_type="frontend-expert", description="Design UI components") 
    )
    database_task = asyncio.create_task(
        Task(subagent_type="database-expert", description="Design database schema")
    )
    
    # Wait for all to complete (30 seconds vs 90 seconds sequential)
    backend_design, frontend_design, database_schema = await asyncio.gather(
        backend_task, frontend_task, database_task
    )
    
    return {
        "backend": backend_design,
        "frontend": frontend_design, 
        "database": database_schema
    }
```

### Context Budget Optimization

```python
# Efficient context loading for specialized agents
def get_agent_context(agent_type, task_complexity):
    """Load only relevant context based on agent and task"""
    
    base_context = ["moai-foundation-trust", "moai-alfred-workflow"]
    
    if agent_type == "backend-expert":
        return base_context + [
            "moai-domain-backend", 
            "moai-domain-web-api",
            "moai-lang-python"
        ]
    elif agent_type == "doc-syncer":
        return base_context + [
            "moai-alfred-document-management",
            "moai-alfred-reporting"
        ]
    # ... other agents
    
# Usage
Task(
    subagent_type="backend-expert",
    context_includes=get_agent_context("backend-expert", "medium")
)
```

---

*Generated with MoAI-ADK Skill Factory v5.0*  
*Examples extracted from original 2226-line comprehensive guide*
