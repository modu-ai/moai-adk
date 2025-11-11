---
name: moai-alfred-agent-guide
version: 5.0.0
created: 2025-10-01
updated: 2025-11-12
status: active
tier: specialization
description: "Complete guide to MoAI-ADK's 19-agent team structure, agent selection decision trees, Haiku vs Sonnet model optimization, and 2025 multi-agent orchestration patterns. Includes role definitions, collaboration protocols, handoff sequences, and performance optimization strategies."
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

**The Complete MoAI-ADK 19-Agent Team Handbook**

> **Primary Agent**: alfred (SuperAgent Orchestrator)  
> **Secondary Agents**: plan-agent, tdd-implementer, test-engineer, doc-syncer, git-manager, qa-validator, tag-agent  
> **Version**: 5.0.0  
> **Keywords**: alfred, agents, orchestration, team, decision-tree, model-selection, collaboration

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

## What It Does

This skill provides the **complete operational guide** for MoAI-ADK's 19-agent team structure, including:

- **Agent Roster**: Detailed roles and responsibilities for all 19 agents
- **Decision Trees**: Systematic agent selection logic for any task
- **Model Selection**: Haiku vs Sonnet optimization criteria
- **Orchestration Patterns**: 2025 best practices for multi-agent coordination
- **Collaboration Protocols**: Handoff sequences and cross-agent communication
- **Performance Optimization**: Resource allocation and context management

**Team Structure** (19 Agents):

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
  ├─ tag-agent (@TAG Management)
  ├─ project-manager (Project Coordination)
  ├─ debug-helper (Troubleshooting)
  └─ trust-checker (TRUST 5 Validation)

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

**Core Principle**: **Delegate-first architecture** — Every task is routed to the most specialized agent. Alfred orchestrates but never executes directly.

---

### Level 2: Practical Implementation (Common Patterns)

## Agent Selection Decision Tree

### Decision Flow

```
User Request
    ↓
┌─────────────────────────────────────────┐
│ Step 1: Determine Task Category         │
│ - Infrastructure/Config?                │
│ - Code Implementation?                  │
│ - Documentation?                        │
│ - Quality/Validation?                   │
│ - Domain-Specific?                      │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ Step 2: Select Primary Agent            │
│ Infrastructure → git-manager/devops     │
│ Code → tdd-implementer                  │
│ Docs → doc-syncer                       │
│ Quality → qa-validator/trust-checker    │
│ Domain → backend/frontend/database...   │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ Step 3: Determine Agent Model           │
│ Complex reasoning? → Sonnet             │
│ Pattern execution? → Haiku              │
│ (See "Haiku vs Sonnet" section)         │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ Step 4: Delegate with Context           │
│ Task(subagent_type="agent-name", ...)   │
└─────────────────────────────────────────┘
```

### Example 1: Feature Implementation Request

**Request**: "Add user authentication with email verification"

**Analysis**:
```
Category: Domain-Specific + Code Implementation
Domain: Backend (auth logic) + Frontend (UI forms)
Complexity: HIGH (security implications, multi-step flow)
Testing: REQUIRED (security-critical)

Agent Selection:
  1. Primary: backend-expert (Sonnet) - Security design
  2. Secondary: security-expert (Sonnet) - Auth strategy validation
  3. Implementer: tdd-implementer (Haiku) - Code generation
  4. Validator: test-engineer (Haiku) - Test coverage
  5. Coordinator: frontend-expert (Haiku) - UI integration

Model Rationale:
  - backend-expert: Sonnet (complex security decisions)
  - security-expert: Sonnet (threat modeling required)
  - tdd-implementer: Haiku (pattern-based code generation)
  - test-engineer: Haiku (standardized testing patterns)
```

**Delegation Sequence**:

```python
# Phase 1: Architecture & Security Design
Task(
    subagent_type="backend-expert",
    description="Design authentication system architecture with email verification",
    prompt="""You are the backend-expert agent (Sonnet).
    
Design a secure authentication system with:
- Email/password registration
- Email verification flow (send token, verify link)
- JWT token management
- Password hashing (bcrypt/argon2)
- Rate limiting (prevent brute force)

Coordinate with security-expert for threat modeling.
Output: Architecture diagram, API endpoints, security checklist.
"""
)

# Phase 2: Security Validation
Task(
    subagent_type="security-expert",
    description="Validate authentication design security posture",
    prompt="""You are the security-expert agent (Sonnet).
    
Review backend-expert's authentication design for:
- OWASP Top 10 compliance
- Token expiration strategies
- CSRF/XSS protections
- SQL injection prevention
- Secure session management

Output: Security audit report, required changes.
"""
)

# Phase 3: TDD Implementation
Task(
    subagent_type="tdd-implementer",
    description="Implement authentication backend with TDD",
    prompt="""You are the tdd-implementer agent (Haiku).
    
Implement the validated authentication design using RED-GREEN-REFACTOR:
- RED: Write failing tests for each endpoint
- GREEN: Implement minimal passing code
- REFACTOR: Optimize with security patterns

Follow backend-expert's architecture and security-expert's requirements.
Output: Production code + comprehensive test suite.
"""
)

# Phase 4: Test Coverage Validation
Task(
    subagent_type="test-engineer",
    description="Validate authentication test coverage",
    prompt="""You are the test-engineer agent (Haiku).
    
Verify authentication implementation has:
- Unit tests: 90%+ coverage
- Integration tests: API endpoints
- Security tests: Auth bypass attempts
- Edge cases: Invalid tokens, expired sessions

Output: Test coverage report, missing scenarios.
"""
)

# Phase 5: Frontend Integration
Task(
    subagent_type="frontend-expert",
    description="Create authentication UI components",
    prompt="""You are the frontend-expert agent (Haiku).
    
Build frontend auth components:
- Registration form with validation
- Email verification status page
- Login form with error handling
- Protected route guards

Coordinate with backend-expert for API contract.
Output: React/Vue components + integration tests.
"""
)
```

### Example 2: Documentation Update Request

**Request**: "Update API documentation for new endpoints"

**Analysis**:
```
Category: Documentation
Complexity: MEDIUM (structured content, technical writing)
Dependencies: Backend code must exist

Agent Selection:
  1. Primary: doc-syncer (Haiku) - Content generation
  2. Validator: qa-validator (Haiku) - Consistency checks

Model Rationale:
  - doc-syncer: Haiku (template-based documentation)
  - qa-validator: Haiku (rule-based validation)
```

**Delegation Sequence**:

```python
# Phase 1: Documentation Generation
Task(
    subagent_type="doc-syncer",
    description="Generate API documentation for new endpoints",
    prompt="""You are the doc-syncer agent (Haiku).
    
Update API documentation with:
- Endpoint descriptions (HTTP method, path, purpose)
- Request/response schemas (JSON examples)
- Authentication requirements
- Error codes and messages
- Usage examples (curl, Python, JavaScript)

Scan backend code for changes, maintain @TAG links.
Output: Updated reference.md, API changelog.
"""
)

# Phase 2: Quality Validation
Task(
    subagent_type="qa-validator",
    description="Validate documentation consistency",
    prompt="""You are the qa-validator agent (Haiku).
    
Check documentation for:
- All endpoints documented
- Request/response schemas match code
- Examples are executable
- @TAG links are valid

Output: Validation report, corrections needed.
"""
)
```

### Example 3: Bug Fix Request

**Request**: "Fix memory leak in background job processor"

**Analysis**:
```
Category: Debugging + Code Fix
Complexity: HIGH (requires investigation, profiling)
Domain: Backend (job processing)

Agent Selection:
  1. Primary: debug-helper (Sonnet) - Root cause analysis
  2. Profiler: performance-engineer (Sonnet) - Memory profiling
  3. Fixer: tdd-implementer (Haiku) - Code repair
  4. Validator: test-engineer (Haiku) - Regression tests

Model Rationale:
  - debug-helper: Sonnet (complex troubleshooting)
  - performance-engineer: Sonnet (profiling analysis)
  - tdd-implementer: Haiku (pattern-based fix)
  - test-engineer: Haiku (standardized testing)
```

**Delegation Sequence**:

```python
# Phase 1: Root Cause Analysis
Task(
    subagent_type="debug-helper",
    description="Diagnose memory leak in job processor",
    prompt="""You are the debug-helper agent (Sonnet).
    
Investigate memory leak:
- Review code for common leak patterns (unclosed connections, circular refs)
- Analyze logs for growth patterns
- Identify suspected code paths

Coordinate with performance-engineer for profiling.
Output: Hypothesis, suspected files, investigation plan.
"""
)

# Phase 2: Performance Profiling
Task(
    subagent_type="performance-engineer",
    description="Profile job processor memory usage",
    prompt="""You are the performance-engineer agent (Sonnet).
    
Profile application with:
- Memory profiler (memory_profiler, tracemalloc)
- Heap dump analysis
- Object retention tracking

Validate debug-helper's hypothesis with data.
Output: Profiling report, memory hotspots, leak confirmation.
"""
)

# Phase 3: Implement Fix with TDD
Task(
    subagent_type="tdd-implementer",
    description="Fix memory leak with TDD approach",
    prompt="""You are the tdd-implementer agent (Haiku).
    
Fix confirmed memory leak:
- RED: Write test demonstrating leak (measure memory growth)
- GREEN: Implement fix (close connections, break circular refs)
- REFACTOR: Optimize resource management

Follow debug-helper and performance-engineer findings.
Output: Fixed code + leak regression test.
"""
)

# Phase 4: Regression Testing
Task(
    subagent_type="test-engineer",
    description="Validate memory leak fix and add monitoring",
    prompt="""You are the test-engineer agent (Haiku).
    
Verify fix completeness:
- Run leak regression test (memory stable over time)
- Stress test job processor (high load, sustained)
- Add monitoring alerts (memory threshold warnings)

Output: Test results, monitoring configuration.
"""
)
```

---

## Haiku vs Sonnet Model Selection

### Decision Matrix

| Scenario | Recommended Model | Rationale |
|----------|------------------|-----------|
| **Strategic Planning** | Sonnet | Complex reasoning, multiple constraints, long-term implications |
| **Architecture Design** | Sonnet | Novel solutions, tradeoff analysis, security implications |
| **Security Review** | Sonnet | Threat modeling, vulnerability analysis, compliance evaluation |
| **Debugging Complex Issues** | Sonnet | Root cause investigation, hypothesis generation, profiling analysis |
| **Research & Analysis** | Sonnet | Information synthesis, comparative evaluation, recommendations |
| **Code Implementation (TDD)** | Haiku | Pattern-based generation, well-defined requirements, fast execution |
| **Testing & QA** | Haiku | Rule-based validation, coverage analysis, regression testing |
| **Documentation** | Haiku | Template-driven content, structured formatting, consistency checks |
| **Git Operations** | Haiku | Standardized workflows (commit, PR, merge), deterministic logic |
| **File Management** | Haiku | CRUD operations, batch processing, directory organization |

### Performance Characteristics

**Sonnet (claude-sonnet-4-5-20250929)**:
- **Strengths**: Deep reasoning, creative solutions, context synthesis, complex decision-making
- **Speed**: ~2-5 seconds per response (longer for complex tasks)
- **Cost**: ~$3 per million input tokens, ~$15 per million output tokens
- **Best For**: Novel problems, strategic decisions, security-critical tasks, ambiguous requirements

**Haiku (claude-haiku-4-20250514)**:
- **Strengths**: Fast execution, pattern recognition, consistency, efficiency
- **Speed**: ~0.5-1 second per response (near-instant for simple tasks)
- **Cost**: ~$0.25 per million input tokens, ~$1.25 per million output tokens
- **Best For**: Well-defined patterns, repetitive tasks, rule-based execution, batch operations

### Cost Optimization Strategy

**Hybrid Orchestration Pattern** (2025 Best Practice):

```python
# Pattern: Sonnet for planning, Haiku for execution

# Phase 1: Strategic Planning (Sonnet)
# Cost: ~$3-5 per analysis
plan_result = Task(
    subagent_type="plan-agent",  # Uses Sonnet
    description="Analyze complex feature requirements",
    prompt="Strategic analysis of user authentication system..."
)

# Phase 2: Execution (Haiku)
# Cost: ~$0.50 per implementation
implementation_result = Task(
    subagent_type="tdd-implementer",  # Uses Haiku
    description="Implement authentication following validated plan",
    prompt="Execute TDD implementation based on plan-agent design..."
)

# Total Cost: ~$3.50-5.50 (vs ~$15-20 all-Sonnet)
# Performance: Same quality, 60-70% cost reduction
```

**When to Override Defaults**:

```python
# Force Sonnet for critical security decisions
Task(
    subagent_type="backend-expert",
    model_override="sonnet",  # Explicit override
    description="Design OAuth2 authorization flow",
    prompt="Security-critical architecture decision requiring deep analysis..."
)

# Force Haiku for simple batch operations
Task(
    subagent_type="doc-syncer",
    model_override="haiku",  # Explicit override
    description="Update API documentation formatting",
    prompt="Standardize markdown formatting across 50 files..."
)
```

---

## The 19-Agent Team Roster

### Orchestration Layer

#### 1. alfred (SuperAgent)

**Role**: Workflow orchestrator and team coordinator

**Responsibilities**:
- Interpret user intent and clarify ambiguities
- Route tasks to appropriate specialist agents
- Manage multi-agent handoffs and dependencies
- Enforce SPEC-first TDD workflow
- Track progress with TodoWrite
- Report outcomes and coordinate completion

**Model**: Sonnet (complex orchestration decisions)

**When to Use**:
- User-facing interactions
- Workflow coordination across multiple agents
- Strategic decision-making
- Quality gate enforcement

**Delegation Pattern**:
```python
# Alfred never executes directly, always delegates
# Example: User requests "Add user registration"

# Alfred analyzes request
# Routes to plan-agent for strategy
# Coordinates backend-expert, tdd-implementer, test-engineer
# Validates with qa-validator
# Commits with git-manager
# Updates docs with doc-syncer
```

**Skills Used**: All 55 Claude Skills (full knowledge base access)

---

### Core Agents (10)

#### 2. plan-agent (Strategic Planning)

**Role**: Strategic analysis and execution planning

**Responsibilities**:
- Decompose user requests into actionable tasks
- Identify dependencies and execution order
- Estimate effort and resource requirements
- Assess risks and propose mitigation strategies
- Create approval-ready execution plans
- Initialize TodoWrite task tracking

**Model**: Sonnet (complex reasoning required)

**When to Use**:
- New feature planning
- Refactoring strategies
- Migration planning
- Architecture decisions
- Risk assessment

**Example Delegation**:
```python
Task(
    subagent_type="plan-agent",
    description="Create execution plan for user authentication feature",
    prompt="""Analyze and decompose authentication feature into:
- Required components (backend API, database, frontend UI)
- Dependencies (libraries, services, infrastructure)
- Execution phases (design, implement, test, deploy)
- Risk factors (security, performance, compatibility)
- Resource estimates (time, complexity, specialist agents)

Output: Structured plan with task DAG, agent assignments, milestones.
"""
)
```

**Skills Used**: moai-alfred-workflow, moai-alfred-todowrite-pattern, moai-foundation-tags

---

#### 3. tdd-implementer (TDD Development)

**Role**: Test-driven development execution

**Responsibilities**:
- Execute RED-GREEN-REFACTOR cycle
- Write failing tests first (RED)
- Implement minimal passing code (GREEN)
- Refactor for quality and maintainability
- Maintain @TAG traceability (SPEC→TEST→CODE)
- Follow TRUST 5 principles

**Model**: Haiku (pattern-based execution)

**When to Use**:
- Feature implementation
- Bug fixes with regression tests
- Code refactoring
- API development

**Example Delegation**:
```python
Task(
    subagent_type="tdd-implementer",
    description="Implement user registration endpoint with TDD",
    prompt="""Follow RED-GREEN-REFACTOR for user registration:

RED Phase:
- Test: POST /api/auth/register with valid data → 201 Created
- Test: POST /api/auth/register with duplicate email → 400 Bad Request
- Test: POST /api/auth/register with weak password → 400 Bad Request

GREEN Phase:
- Implement registration logic (minimal passing code)
- Hash passwords with bcrypt
- Store user in database
- Return JWT token

REFACTOR Phase:
- Extract validation logic
- Optimize database queries
- Add error handling

Maintain @TAG links: @SPEC-AUTH-001 → @TEST-AUTH-001 → @CODE-AUTH-001
"""
)
```

**Skills Used**: moai-lang-python, moai-domain-backend, moai-foundation-tags

---

#### 4. test-engineer (Testing & QA)

**Role**: Comprehensive testing and quality validation

**Responsibilities**:
- Design test strategies (unit, integration, e2e)
- Validate test coverage (90%+ target)
- Create regression test suites
- Perform edge case testing
- Generate test reports
- Maintain test infrastructure

**Model**: Haiku (rule-based validation)

**When to Use**:
- Test coverage validation
- Regression testing
- QA gate enforcement
- Test infrastructure setup

**Example Delegation**:
```python
Task(
    subagent_type="test-engineer",
    description="Validate authentication system test coverage",
    prompt="""Analyze authentication implementation for test completeness:

Unit Tests (Target: 95%):
- Password hashing functions
- Token generation/validation
- Email verification logic
- Database models

Integration Tests (Target: 90%):
- API endpoints (register, login, verify)
- Database transactions
- Email delivery

Security Tests:
- SQL injection attempts
- XSS attack vectors
- Brute force protection
- Token expiration

Generate coverage report with missing scenarios.
"""
)
```

**Skills Used**: moai-lang-python, moai-domain-testing, moai-alfred-best-practices

---

#### 5. doc-syncer (Documentation)

**Role**: Documentation generation and synchronization

**Responsibilities**:
- Generate API documentation from code
- Synchronize SPEC → CODE → DOC
- Maintain @TAG references
- Update README, CHANGELOG, reference.md
- Create usage examples
- Validate documentation accuracy

**Model**: Haiku (template-driven content)

**When to Use**:
- Documentation updates
- API reference generation
- Changelog maintenance
- README updates

**Example Delegation**:
```python
Task(
    subagent_type="doc-syncer",
    description="Update API documentation for authentication endpoints",
    prompt="""Generate documentation for authentication API:

Endpoints:
- POST /api/auth/register (User Registration)
- POST /api/auth/login (User Login)
- POST /api/auth/verify-email (Email Verification)
- POST /api/auth/refresh-token (Token Refresh)

For each endpoint document:
- HTTP method and path
- Request body schema (JSON)
- Response schema (success + error cases)
- Authentication requirements
- Usage examples (curl, Python, JavaScript)
- Error codes and meanings

Update reference.md with @TAG links to implementation.
"""
)
```

**Skills Used**: moai-alfred-document-management, moai-foundation-tags, moai-alfred-reporting

---

#### 6. git-manager (Git Operations)

**Role**: Git workflow and version control

**Responsibilities**:
- Create feature branches (feature/SPEC-XXX)
- Commit with TDD cycle (RED, GREEN, REFACTOR)
- Generate conventional commit messages
- Create pull requests targeting develop
- Validate GitFlow compliance
- Tag releases

**Model**: Haiku (deterministic workflows)

**When to Use**:
- Branch creation
- Commit operations
- Pull request creation
- Release tagging

**Example Delegation**:
```python
Task(
    subagent_type="git-manager",
    description="Commit authentication implementation (GREEN phase)",
    prompt="""Create GREEN phase commit for authentication feature:

Commit Message Format:
feat(auth): Implement user registration endpoint

- Add POST /api/auth/register endpoint
- Implement password hashing with bcrypt
- Add email verification token generation
- Create user database model
- Pass all @TEST-AUTH-001 tests

@TAG: @SPEC-AUTH-001 @TEST-AUTH-001 @CODE-AUTH-001

Files Changed:
- src/api/auth/register.py (new)
- src/models/user.py (modified)
- tests/test_auth_register.py (updated)

Execute git commit with conventional format.
"""
)
```

**Skills Used**: moai-alfred-agent-guide, moai-foundation-tags

---

#### 7. qa-validator (Quality Assurance)

**Role**: Quality gate enforcement

**Responsibilities**:
- Enforce TRUST 5 principles (Test First, Readable, Unified, Secured, Trackable)
- Validate code quality metrics
- Check test coverage thresholds
- Verify security best practices
- Validate @TAG integrity
- Generate quality reports

**Model**: Haiku (rule-based validation)

**When to Use**:
- Quality gate checks
- Pre-merge validation
- TRUST 5 enforcement
- Code review automation

**Example Delegation**:
```python
Task(
    subagent_type="qa-validator",
    description="Validate authentication feature quality",
    prompt="""Validate TRUST 5 compliance for authentication:

Test First:
- All code has tests written first ✓/✗
- Test coverage ≥90% ✓/✗

Readable:
- Code follows style guide (Black, flake8) ✓/✗
- Docstrings present for all functions ✓/✗
- Variable names descriptive ✓/✗

Unified:
- Architecture patterns consistent ✓/✗
- Error handling standardized ✓/✗

Secured:
- Passwords hashed (bcrypt/argon2) ✓/✗
- SQL injection prevented (parameterized) ✓/✗
- CSRF/XSS protections enabled ✓/✗

Trackable:
- @TAG links valid (SPEC→TEST→CODE) ✓/✗
- Git commits conventional format ✓/✗

Generate quality gate pass/fail report.
"""
)
```

**Skills Used**: moai-alfred-best-practices, moai-foundation-tags, moai-alfred-rules

---

#### 8. tag-agent (@TAG Management)

**Role**: Traceability and @TAG lifecycle management

**Responsibilities**:
- Generate @TAG identifiers (SPEC, TEST, CODE, DOC)
- Validate @TAG chains (SPEC→TEST→CODE→DOC)
- Update @TAG references across files
- Detect orphaned @TAGs
- Generate traceability reports
- Maintain @TAG registry

**Model**: Haiku (deterministic identifier management)

**When to Use**:
- @TAG generation
- Traceability validation
- @TAG chain verification
- Orphaned @TAG cleanup

**Example Delegation**:
```python
Task(
    subagent_type="tag-agent",
    description="Generate @TAG chain for authentication feature",
    prompt="""Create complete @TAG chain for authentication:

SPEC TAG:
@SPEC-AUTH-001: User registration with email verification
Location: .moai/specs/SPEC-AUTH-001.md

TEST TAG:
@TEST-AUTH-001: Registration endpoint test suite
Location: tests/test_auth_register.py
Links: @SPEC-AUTH-001

CODE TAG:
@CODE-AUTH-001: Registration endpoint implementation
Location: src/api/auth/register.py
Links: @SPEC-AUTH-001 @TEST-AUTH-001

DOC TAG:
@DOC-AUTH-001: Registration API documentation
Location: .moai/docs/reference.md
Links: @SPEC-AUTH-001 @CODE-AUTH-001

Validate chain integrity and update @TAG registry.
"""
)
```

**Skills Used**: moai-foundation-tags, moai-alfred-workflow

---

#### 9. project-manager (Project Coordination)

**Role**: Project planning and resource coordination

**Responsibilities**:
- Coordinate multi-feature roadmaps
- Manage sprint planning
- Track milestones and deliverables
- Allocate agent resources
- Generate project reports
- Facilitate stakeholder communication

**Model**: Sonnet (strategic coordination)

**When to Use**:
- Sprint planning
- Roadmap coordination
- Resource allocation
- Project reporting

**Example Delegation**:
```python
Task(
    subagent_type="project-manager",
    description="Plan authentication system sprint",
    prompt="""Plan 2-week sprint for authentication system:

Sprint Goal: Complete user authentication with email verification

Features:
1. User registration (SPEC-AUTH-001) - 3 days
2. Email verification (SPEC-AUTH-002) - 2 days
3. User login (SPEC-AUTH-003) - 2 days
4. Token refresh (SPEC-AUTH-004) - 1 day
5. Password reset (SPEC-AUTH-005) - 3 days
6. Integration testing (SPEC-AUTH-006) - 2 days

Agent Allocation:
- backend-expert (design): 2 days
- security-expert (review): 1 day
- tdd-implementer (code): 8 days
- test-engineer (QA): 4 days
- frontend-expert (UI): 5 days

Generate sprint backlog with dependencies and milestones.
"""
)
```

**Skills Used**: moai-alfred-workflow, moai-alfred-todowrite-pattern, moai-foundation-tags

---

#### 10. debug-helper (Troubleshooting)

**Role**: Debugging and root cause analysis

**Responsibilities**:
- Investigate bugs and errors
- Perform root cause analysis
- Profile application performance
- Analyze logs and stack traces
- Propose fix strategies
- Coordinate with domain experts

**Model**: Sonnet (complex investigation)

**When to Use**:
- Bug investigation
- Performance issues
- Production errors
- Unexpected behavior

**Example Delegation**:
```python
Task(
    subagent_type="debug-helper",
    description="Debug authentication token expiration issue",
    prompt="""Investigate reported issue: Users logged out unexpectedly

Symptoms:
- Users report being logged out after 5 minutes
- Expected: Sessions valid for 24 hours
- Token refresh endpoint returns 401 Unauthorized

Investigation Steps:
1. Review token generation code (src/api/auth/tokens.py)
2. Check JWT expiration configuration
3. Analyze token refresh logic
4. Review authentication middleware
5. Inspect database session storage

Hypotheses:
- Token expiration misconfigured (5 min vs 24 hr)
- Refresh token not being stored
- Middleware checking wrong token type

Generate root cause analysis with proposed fix.
"""
)
```

**Skills Used**: moai-lang-python, moai-domain-backend, moai-alfred-dev-guide

---

#### 11. trust-checker (TRUST 5 Validation)

**Role**: TRUST 5 principle enforcement

**Responsibilities**:
- Validate Test First principle (code has tests)
- Check Readability (style guides, documentation)
- Verify Unified patterns (consistency)
- Audit Security posture (OWASP compliance)
- Validate Traceability (@TAG integrity)
- Generate TRUST 5 compliance reports

**Model**: Haiku (rule-based validation)

**When to Use**:
- Pre-merge validation
- Quality gate enforcement
- TRUST 5 audits
- Compliance reporting

**Example Delegation**:
```python
Task(
    subagent_type="trust-checker",
    description="Validate authentication feature TRUST 5 compliance",
    prompt="""Perform TRUST 5 audit for authentication system:

Test First:
✓ Tests written before implementation
✓ Coverage ≥90%
✓ All tests passing

Readable:
✓ Code follows Black style guide
✓ Docstrings for all public functions
✓ Variable names descriptive
✓ Comments explain complex logic

Unified:
✓ Error handling consistent across endpoints
✓ Response format standardized
✓ Architecture patterns followed

Secured:
✓ Passwords hashed with bcrypt (cost=12)
✓ SQL queries parameterized (no injection)
✓ JWT tokens signed with HS256
✓ CSRF tokens validated
✓ Rate limiting enabled (10 req/min)

Trackable:
✓ @TAG chain complete (SPEC→TEST→CODE→DOC)
✓ Git commits conventional format
✓ Changelog updated

Generate compliance report with pass/fail for each principle.
"""
)
```

**Skills Used**: moai-alfred-best-practices, moai-alfred-rules, moai-foundation-tags

---

### Domain Specialists (6)

#### 12. backend-expert (Backend Development)

**Role**: Backend architecture and implementation specialist

**Responsibilities**:
- Design backend APIs and services
- Implement business logic
- Database schema design
- Performance optimization
- Integration with external services
- Backend security patterns

**Model**: Sonnet (architecture design) or Haiku (implementation)

**When to Use**:
- API design
- Backend architecture
- Service implementation
- Database schema design

**Example Delegation**:
```python
Task(
    subagent_type="backend-expert",
    description="Design authentication API architecture",
    prompt="""Design comprehensive authentication API:

Requirements:
- User registration with email verification
- Login with JWT tokens
- Token refresh mechanism
- Password reset flow
- Role-based access control (RBAC)

Design:
1. Endpoint structure (RESTful conventions)
2. Request/response schemas (JSON)
3. Database models (User, Token, Role)
4. Authentication middleware
5. Error handling strategy
6. Performance considerations (caching, rate limiting)

Coordinate with security-expert for security review.
Output: API specification, database schema, sequence diagrams.
"""
)
```

**Skills Used**: moai-domain-backend, moai-domain-web-api, moai-lang-python

---

#### 13. frontend-expert (Frontend Development)

**Role**: Frontend architecture and UI implementation specialist

**Responsibilities**:
- Design frontend architecture
- Implement UI components
- State management patterns
- Frontend performance optimization
- Accessibility compliance
- Frontend testing strategies

**Model**: Sonnet (architecture) or Haiku (component implementation)

**When to Use**:
- UI component design
- Frontend architecture
- State management
- Client-side routing

**Example Delegation**:
```python
Task(
    subagent_type="frontend-expert",
    description="Design authentication UI components",
    prompt="""Design frontend authentication components:

Components Needed:
1. RegistrationForm
   - Email/password inputs with validation
   - Password strength indicator
   - Terms of service checkbox
   - Submit button with loading state

2. LoginForm
   - Email/password inputs
   - "Remember me" checkbox
   - Forgot password link
   - Error message display

3. EmailVerificationPage
   - Verification status display
   - Resend verification email button
   - Redirect to login after success

4. PasswordResetForm
   - Email input for reset request
   - New password input for reset completion
   - Confirmation message

Design Patterns:
- Form validation (client-side + server-side)
- Error handling (display user-friendly messages)
- Loading states (disable buttons during API calls)
- Accessibility (ARIA labels, keyboard navigation)

Output: Component specifications, state management strategy, API integration plan.
"""
)
```

**Skills Used**: moai-domain-frontend, moai-lang-javascript, moai-domain-web-api

---

#### 14. database-expert (Database Design)

**Role**: Database architecture and optimization specialist

**Responsibilities**:
- Design database schemas
- Optimize queries and indexes
- Data migration strategies
- Database performance tuning
- Data integrity and constraints
- Backup and recovery planning

**Model**: Sonnet (schema design) or Haiku (query optimization)

**When to Use**:
- Database schema design
- Query optimization
- Data migrations
- Performance tuning

**Example Delegation**:
```python
Task(
    subagent_type="database-expert",
    description="Design authentication database schema",
    prompt="""Design database schema for authentication system:

Tables:
1. users
   - id (UUID, primary key)
   - email (VARCHAR, unique, indexed)
   - password_hash (VARCHAR, bcrypt)
   - email_verified (BOOLEAN, default false)
   - created_at (TIMESTAMP)
   - updated_at (TIMESTAMP)

2. email_verification_tokens
   - id (UUID, primary key)
   - user_id (UUID, foreign key → users.id)
   - token (VARCHAR, unique, indexed)
   - expires_at (TIMESTAMP)
   - created_at (TIMESTAMP)

3. refresh_tokens
   - id (UUID, primary key)
   - user_id (UUID, foreign key → users.id)
   - token (VARCHAR, unique, indexed)
   - expires_at (TIMESTAMP)
   - revoked (BOOLEAN, default false)
   - created_at (TIMESTAMP)

Design Considerations:
- Indexes for performance (email, token lookups)
- Foreign key constraints (cascade delete)
- Token expiration cleanup (cron job strategy)
- Query optimization (explain analyze critical paths)

Output: SQL schema, migration scripts, index strategy.
"""
)
```

**Skills Used**: moai-domain-database, moai-lang-python, moai-domain-backend

---

#### 15. security-expert (Security & Authentication)

**Role**: Security architecture and threat mitigation specialist

**Responsibilities**:
- Security architecture design
- Threat modeling and risk assessment
- Authentication/authorization strategies
- Vulnerability assessment
- Compliance validation (OWASP, GDPR)
- Security testing and auditing

**Model**: Sonnet (complex security analysis)

**When to Use**:
- Security design
- Threat modeling
- Vulnerability assessment
- Compliance validation

**Example Delegation**:
```python
Task(
    subagent_type="security-expert",
    description="Security audit of authentication system",
    prompt="""Perform comprehensive security audit:

OWASP Top 10 Validation:
1. Injection (SQL, NoSQL, Command)
   - Check: Parameterized queries used
   - Check: Input sanitization present

2. Broken Authentication
   - Check: Password hashing (bcrypt cost ≥12)
   - Check: Token expiration configured
   - Check: Brute force protection (rate limiting)
   - Check: Session management secure

3. Sensitive Data Exposure
   - Check: HTTPS enforced
   - Check: Passwords never logged
   - Check: Tokens stored securely

4. XML External Entities (XXE)
   - Check: Not applicable (JSON API)

5. Broken Access Control
   - Check: RBAC implemented correctly
   - Check: Authorization checks on all endpoints

6. Security Misconfiguration
   - Check: Default credentials changed
   - Check: Error messages don't leak info
   - Check: Security headers present (HSTS, CSP)

7. Cross-Site Scripting (XSS)
   - Check: Input validation on frontend
   - Check: Output encoding

8. Insecure Deserialization
   - Check: JSON parsing safe

9. Using Components with Known Vulnerabilities
   - Check: Dependencies up to date
   - Check: Security patches applied

10. Insufficient Logging & Monitoring
    - Check: Authentication failures logged
    - Check: Anomaly detection configured

Generate security audit report with risk ratings (Critical/High/Medium/Low).
"""
)
```

**Skills Used**: moai-domain-security, moai-alfred-best-practices, moai-domain-backend

---

#### 16. devops-expert (Infrastructure & DevOps)

**Role**: Infrastructure architecture and deployment specialist

**Responsibilities**:
- Infrastructure as Code (IaC) design
- CI/CD pipeline configuration
- Container orchestration (Docker, Kubernetes)
- Cloud platform management (AWS, Azure, GCP)
- Monitoring and observability
- Deployment strategies (blue-green, canary)

**Model**: Sonnet (infrastructure design) or Haiku (configuration)

**When to Use**:
- Infrastructure design
- CI/CD setup
- Deployment automation
- Monitoring configuration

**Example Delegation**:
```python
Task(
    subagent_type="devops-expert",
    description="Design CI/CD pipeline for authentication service",
    prompt="""Design comprehensive CI/CD pipeline:

Pipeline Stages:
1. Build
   - Install dependencies (pip install -r requirements.txt)
   - Lint code (flake8, black --check)
   - Type check (mypy)

2. Test
   - Run unit tests (pytest tests/unit/)
   - Run integration tests (pytest tests/integration/)
   - Generate coverage report (coverage ≥90%)

3. Security Scan
   - Dependency vulnerability scan (safety check)
   - SAST scan (bandit)
   - Container image scan (trivy)

4. Build Docker Image
   - Build image (docker build -t auth-service:latest)
   - Tag with commit SHA (auth-service:${GIT_SHA})
   - Push to registry (ECR/GCR/DockerHub)

5. Deploy to Staging
   - Deploy to staging environment
   - Run smoke tests
   - Health check endpoints

6. Deploy to Production (manual approval)
   - Blue-green deployment strategy
   - Gradual traffic shift (10% → 50% → 100%)
   - Automated rollback on errors

Output: GitHub Actions / GitLab CI YAML configuration, deployment scripts.
"""
)
```

**Skills Used**: moai-domain-devops, moai-lang-shell, moai-alfred-workflow

---

#### 17. api-designer (API Architecture)

**Role**: API design and contract specialist

**Responsibilities**:
- RESTful API design
- GraphQL schema design
- API versioning strategies
- OpenAPI/Swagger documentation
- API contract testing
- Backward compatibility management

**Model**: Sonnet (API architecture) or Haiku (OpenAPI generation)

**When to Use**:
- API design
- Contract definition
- API versioning
- OpenAPI documentation

**Example Delegation**:
```python
Task(
    subagent_type="api-designer",
    description="Design RESTful authentication API contract",
    prompt="""Design comprehensive API contract:

Endpoint: POST /api/v1/auth/register

Request:
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "accept_terms": true
}

Response (201 Created):
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "email_verified": false,
    "created_at": "2025-11-12T10:30:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "550e8400-e29b-41d4-a716-446655440001",
  "expires_in": 3600
}

Response (400 Bad Request):
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Email already registered",
    "details": [
      {
        "field": "email",
        "message": "Email address is already in use"
      }
    ]
  }
}

Design all authentication endpoints with:
- Request/response schemas
- Validation rules
- Error codes
- Rate limiting (10 requests/minute)
- OpenAPI 3.0 specification

Output: OpenAPI YAML, API contract document, usage examples.
"""
)
```

**Skills Used**: moai-domain-web-api, moai-domain-backend, moai-alfred-document-management

---

### Built-in Agents (2)

#### 18. Plan (Built-in Strategic Analysis Agent)

**Role**: Deep strategic analysis and planning (Claude Code built-in)

**Responsibilities**:
- Complex strategic analysis
- Multi-constraint optimization
- Comparative evaluation
- Recommendation generation
- Risk assessment

**Model**: Sonnet (Claude Code default)

**When to Use**:
- Strategic decisions
- Technology selection
- Architecture comparisons
- Risk analysis

**Example Delegation**:
```python
Task(
    subagent_type="Plan",
    description="Analyze authentication technology options",
    prompt="""Compare authentication technologies for production system:

Options:
1. JWT (JSON Web Tokens)
   - Pros: Stateless, scalable, widely supported
   - Cons: Token revocation complexity, larger payload

2. Session-based (Server-side storage)
   - Pros: Easy revocation, smaller payload
   - Cons: Requires centralized session store, less scalable

3. OAuth 2.0 (Delegated authorization)
   - Pros: Industry standard, third-party login
   - Cons: Complex implementation, external dependencies

Evaluation Criteria:
- Scalability (10K → 1M users)
- Security posture
- Implementation complexity
- Operational cost
- Maintenance burden

Generate comparative analysis with recommendation.
"""
)
```

**Skills Used**: All Claude Code built-in capabilities

---

#### 19. WebSearch (Built-in Research Agent)

**Role**: Web research and information validation (Claude Code built-in)

**Responsibilities**:
- Latest documentation lookup
- Best practice research
- Technology comparison
- Version compatibility checks
- Security advisory lookup

**Model**: Sonnet (Claude Code default)

**When to Use**:
- Latest framework versions
- Security advisories
- Best practice research
- Documentation lookup

**Example Usage**:
```python
# WebSearch is used directly via WebSearch tool, not Task()
WebSearch(
    query="Python bcrypt library latest version 2025 security best practices"
)

# Results inform other agent decisions
# Example: security-expert uses findings to validate password hashing strategy
```

---

## Multi-Agent Orchestration Patterns (2025)

### Pattern 1: Orchestrator-Worker (Most Common)

**Use Case**: Single orchestrator delegates to specialized workers

**Structure**:
```
alfred (Orchestrator)
  ├─ plan-agent (Strategic Planning)
  ├─ tdd-implementer (Code Implementation)
  ├─ test-engineer (Testing)
  └─ doc-syncer (Documentation)
```

**Example**:
```python
# User request: "Add user registration"

# Phase 1: Alfred orchestrates
# Phase 2: plan-agent creates strategy
# Phase 3: tdd-implementer executes code
# Phase 4: test-engineer validates tests
# Phase 5: doc-syncer updates docs
# Phase 6: Alfred reports completion
```

**Advantages**:
- Clear ownership and accountability
- Isolated contexts (agent-specific knowledge)
- Parallel execution (independent workers)
- Simple coordination logic

---

### Pattern 2: Hierarchical Agent Pattern

**Use Case**: Multi-level delegation with sub-orchestrators

**Structure**:
```
alfred (SuperAgent)
  └─ backend-expert (Sub-orchestrator)
      ├─ tdd-implementer (Worker)
      ├─ database-expert (Worker)
      └─ security-expert (Worker)
```

**Example**:
```python
# User request: "Build e-commerce checkout"

# Level 1: alfred delegates to backend-expert
Task(subagent_type="backend-expert", description="Design checkout system")

# Level 2: backend-expert delegates to specialists
#   - database-expert: Design order/payment tables
#   - security-expert: Validate PCI DSS compliance
#   - tdd-implementer: Implement checkout logic

# Results bubble up: backend-expert → alfred → user
```

**Advantages**:
- Encapsulates domain complexity
- Reduces alfred's coordination overhead
- Enables domain-specific workflows

---

### Pattern 3: Handoff Orchestration

**Use Case**: Sequential agent handoffs with context transfer

**Structure**:
```
plan-agent → tdd-implementer → test-engineer → qa-validator → git-manager
     ↓            ↓               ↓                ↓              ↓
   Plan      Code + Tests    Test Report    Quality Gate    Git Commit
```

**Example**:
```python
# Phase 1: plan-agent creates strategy
plan_result = Task(
    subagent_type="plan-agent",
    description="Plan authentication feature",
    prompt="Decompose user authentication into tasks..."
)

# Phase 2: tdd-implementer receives plan
code_result = Task(
    subagent_type="tdd-implementer",
    description="Implement authentication per plan",
    prompt=f"Execute plan:\n{plan_result.output}\n\nFollow TDD cycle..."
)

# Phase 3: test-engineer validates
test_result = Task(
    subagent_type="test-engineer",
    description="Validate test coverage",
    prompt=f"Validate implementation:\n{code_result.files}\n\nCheck coverage ≥90%..."
)

# Phase 4: qa-validator enforces quality
qa_result = Task(
    subagent_type="qa-validator",
    description="TRUST 5 validation",
    prompt=f"Validate quality:\n{test_result.report}\n\nEnforce TRUST 5..."
)

# Phase 5: git-manager commits
Task(
    subagent_type="git-manager",
    description="Commit GREEN phase",
    prompt=f"Commit changes:\n{code_result.files}\n\nQuality: {qa_result.status}..."
)
```

**Advantages**:
- Clear pipeline stages
- Context transfer between agents
- Quality gates enforced at each step
- Audit trail of decisions

---

### Pattern 4: Magentic Pattern (Context-Aware Selection)

**Use Case**: Dynamic agent selection based on context analysis

**Structure**:
```
alfred analyzes request context
  ├─ If security-critical → security-expert
  ├─ If performance issue → performance-engineer
  ├─ If bug report → debug-helper
  └─ If feature request → plan-agent
```

**Example**:
```python
# User request: "Fix slow database queries"

# Alfred analyzes context:
# - Keywords: "slow", "database", "queries"
# - Category: Performance issue
# - Domain: Database

# Agent selection:
primary_agent = "performance-engineer"  # Performance focus
secondary_agent = "database-expert"     # Database domain

# Delegation:
Task(
    subagent_type="performance-engineer",
    description="Diagnose slow database queries",
    prompt="Profile database queries and identify bottlenecks. Coordinate with database-expert for optimization."
)
```

**Advantages**:
- Adapts to request context
- Optimizes for task type
- Reduces trial-and-error

---

### Pattern 5: Group Chat Orchestration

**Use Case**: Multiple agents collaborate simultaneously with shared context

**Structure**:
```
Shared Context: Authentication Security Review

Participants:
├─ backend-expert (API design)
├─ security-expert (Threat modeling)
├─ database-expert (Schema design)
└─ devops-expert (Deployment security)
```

**Example**:
```python
# Simultaneous collaboration on shared goal

# All agents receive same context
shared_context = """
Goal: Design secure authentication system for financial application

Constraints:
- PCI DSS Level 1 compliance required
- 99.99% uptime SLA
- < 200ms latency for auth requests
- Support 100K concurrent users

Each agent provides domain-specific input.
"""

# Parallel agent invocations
backend_design = Task(subagent_type="backend-expert", prompt=shared_context + "Design API architecture")
security_review = Task(subagent_type="security-expert", prompt=shared_context + "Threat model and mitigations")
database_schema = Task(subagent_type="database-expert", prompt=shared_context + "Design schema for scale")
deployment_plan = Task(subagent_type="devops-expert", prompt=shared_context + "Plan secure deployment")

# Alfred synthesizes results
# Resolves conflicts, generates unified design
```

**Advantages**:
- Parallel expert input
- Holistic system design
- Cross-domain validation

---

## Agent Collaboration Protocols

### Handoff Sequence Best Practices

**Rule 1**: Always include context from previous agent

```python
# ❌ BAD: No context transfer
Task(subagent_type="test-engineer", description="Write tests")

# ✅ GOOD: Context from previous agent
Task(
    subagent_type="test-engineer",
    description="Write tests for authentication implementation",
    prompt=f"""Validate implementation from tdd-implementer:

Files Modified:
{tdd_implementer_result.files}

Implementation Summary:
{tdd_implementer_result.summary}

Write tests covering all code paths.
"""
)
```

**Rule 2**: Document assumptions and constraints

```python
Task(
    subagent_type="backend-expert",
    description="Design payment processing API",
    prompt="""Design payment API with constraints:

Assumptions:
- Using Stripe as payment provider
- Subscription-based billing model
- Monthly recurring payments

Constraints:
- PCI DSS compliance required (Level 2)
- Webhook reliability critical (retries, idempotency)
- Support refunds and proration

Document design decisions for downstream agents.
"""
)
```

**Rule 3**: Validate handoff completeness

```python
# After agent completes, verify outputs before next handoff
if not validate_agent_output(plan_result):
    raise HandoffError("plan-agent output incomplete, missing task DAG")

# Only proceed when validated
Task(subagent_type="tdd-implementer", prompt=f"Execute plan:\n{plan_result.output}")
```

---

### Cross-Agent Communication

**Pattern**: Agent A needs input from Agent B

**Implementation**:
```python
# Option 1: Sequential delegation (alfred orchestrates)
backend_design = Task(subagent_type="backend-expert", ...)
frontend_contract = Task(
    subagent_type="frontend-expert",
    prompt=f"Build UI for API:\n{backend_design.endpoints}"
)

# Option 2: Embedded coordination (agent-to-agent)
Task(
    subagent_type="backend-expert",
    description="Design API for frontend consumption",
    prompt="""Design API endpoints.
    
After design, coordinate with frontend-expert to validate:
- Endpoint naming conventions
- Request/response formats
- Error handling expectations

Adjust design based on frontend feedback.
"""
)
```

---

### Conflict Resolution

**Scenario**: Two agents provide conflicting recommendations

**Example**:
```python
# backend-expert recommends: REST API
# api-designer recommends: GraphQL API

# Alfred resolves conflict
conflict_resolution = Task(
    subagent_type="Plan",
    description="Resolve API design conflict: REST vs GraphQL",
    prompt="""Two agents disagree on API architecture:

backend-expert recommendation: REST API
Rationale: Simpler implementation, well-understood, easier caching

api-designer recommendation: GraphQL API
Rationale: Flexible queries, reduced overfetching, better for mobile

Project constraints:
- Team has REST experience (no GraphQL)
- Mobile app is primary client (data efficiency critical)
- Timeline: 4 weeks

Provide decision with justification.
"""
)

# Plan agent provides strategic recommendation
# Alfred implements chosen approach
```

---

## Performance Optimization

### Context Budget Management

**Problem**: Large agent contexts consume tokens and slow responses

**Solution**: Scope agent contexts to domain-specific knowledge

```python
# ❌ BAD: Load all 55 skills for simple task
Task(
    subagent_type="doc-syncer",
    description="Update README",
    context_includes="all_skills"  # 200K+ tokens
)

# ✅ GOOD: Load only relevant skills
Task(
    subagent_type="doc-syncer",
    description="Update README",
    context_includes=["moai-alfred-document-management", "moai-alfred-reporting"]  # ~20K tokens
)
```

**Best Practice**: Use `Skill("skill-name")` for on-demand loading within agents.

---

### Parallel Agent Execution

**Pattern**: Execute independent agents in parallel

```python
# ❌ BAD: Sequential execution (slow)
backend_result = Task(subagent_type="backend-expert", ...)  # 30 seconds
frontend_result = Task(subagent_type="frontend-expert", ...) # 30 seconds
database_result = Task(subagent_type="database-expert", ...)  # 30 seconds
# Total: 90 seconds

# ✅ GOOD: Parallel execution (fast)
import asyncio

async def parallel_design():
    backend_task = asyncio.create_task(Task(subagent_type="backend-expert", ...))
    frontend_task = asyncio.create_task(Task(subagent_type="frontend-expert", ...))
    database_task = asyncio.create_task(Task(subagent_type="database-expert", ...))
    
    results = await asyncio.gather(backend_task, frontend_task, database_task)
    return results

# Total: 30 seconds (3x speedup)
```

---

### Model Selection Optimization

**Strategy**: Use Haiku for 80% of tasks, Sonnet for 20%

```python
# Typical feature implementation breakdown:
# - Planning: Sonnet (1 agent, 5 min)
# - Implementation: Haiku (3 agents, 10 min)
# - Testing: Haiku (2 agents, 5 min)
# - Documentation: Haiku (1 agent, 2 min)
# - Validation: Haiku (2 agents, 3 min)

# Cost Analysis:
# All Sonnet: $50 per feature
# Hybrid (Sonnet planning, Haiku execution): $12 per feature
# Savings: 76% cost reduction, same quality
```

---

### Level 3: Advanced Patterns (Expert Reference)

## Advanced Orchestration Scenarios

### Scenario 1: Complex Multi-Domain Feature

**Request**: "Build real-time chat with encryption and moderation"

**Agent Orchestration**:

```python
# Phase 1: Strategic Planning (Sonnet)
plan = Task(subagent_type="plan-agent", description="Plan real-time chat system")

# Phase 2: Domain Design (Parallel, Sonnet)
backend_design = Task(subagent_type="backend-expert", description="WebSocket server + message queue")
security_design = Task(subagent_type="security-expert", description="End-to-end encryption strategy")
database_design = Task(subagent_type="database-expert", description="Message persistence + indexing")
frontend_design = Task(subagent_type="frontend-expert", description="React chat UI + state management")

# Phase 3: Implementation (Parallel, Haiku)
backend_impl = Task(subagent_type="tdd-implementer", description="WebSocket server (Python/asyncio)")
frontend_impl = Task(subagent_type="tdd-implementer", description="React chat components")
moderation_impl = Task(subagent_type="backend-expert", description="AI-based content moderation")

# Phase 4: Integration Testing (Haiku)
integration_tests = Task(subagent_type="test-engineer", description="End-to-end chat flow tests")

# Phase 5: Security Audit (Sonnet)
security_audit = Task(subagent_type="security-expert", description="Encryption + moderation security review")

# Phase 6: Deployment (Haiku)
deployment = Task(subagent_type="devops-expert", description="WebSocket infrastructure + scaling")

# Total Agents: 11 (1 plan, 4 design, 3 implement, 1 test, 1 audit, 1 deploy)
# Execution Time: ~45 minutes (with parallelization)
```

---

### Scenario 2: Emergency Production Bug

**Request**: "Critical: Payment processing failing for 10% of transactions"

**Agent Orchestration**:

```python
# Phase 1: Immediate Triage (Sonnet, High Priority)
triage = Task(
    subagent_type="debug-helper",
    description="Emergency triage: Payment failures",
    priority="CRITICAL",
    prompt="""Diagnose payment processing failures:
    
Symptoms:
- 10% of payments failing with 500 errors
- Started 2 hours ago
- Error logs: "stripe_api_timeout"

Immediate actions:
1. Check Stripe API status
2. Review recent code deployments
3. Analyze error logs
4. Identify affected user segment

Generate incident report with immediate recommendations.
"""
)

# Phase 2: Root Cause Analysis (Sonnet)
root_cause = Task(
    subagent_type="performance-engineer",
    description="Profile payment processing bottleneck",
    prompt=f"""Investigate payment failures:

Triage findings:
{triage.report}

Deep analysis:
- Profile payment endpoint latency
- Check database query performance
- Review Stripe API call patterns
- Analyze network timeouts

Identify root cause and propose fix.
"""
)

# Phase 3: Hotfix Implementation (Haiku, Expedited)
hotfix = Task(
    subagent_type="tdd-implementer",
    description="Implement payment processing fix",
    prompt=f"""Implement hotfix:

Root cause:
{root_cause.analysis}

Fix strategy:
- Add retry logic with exponential backoff
- Increase Stripe API timeout (5s → 15s)
- Add circuit breaker (fail fast after 3 retries)

Deploy to production immediately after validation.
"""
)

# Phase 4: Validation (Haiku, Expedited)
validation = Task(
    subagent_type="test-engineer",
    description="Validate hotfix in staging",
    prompt="Test payment processing fix with 1000 simulated transactions. Verify success rate >99%."
)

# Phase 5: Emergency Deployment (Haiku)
deployment = Task(
    subagent_type="devops-expert",
    description="Deploy payment hotfix to production",
    prompt="Blue-green deployment with immediate rollback capability. Monitor error rates for 10 minutes."
)

# Total Time: ~20 minutes (emergency priority)
# Cost: ~$5 (Sonnet for analysis, Haiku for execution)
```

---

## Anti-Patterns to Avoid

### ❌ Anti-Pattern 1: Agent Bypassing

**Problem**: Commands executing tasks directly instead of delegating

```python
# ❌ BAD: Command implements feature
/alfred:2-run SPEC-001
  → alfred reads files, writes code, runs tests
  → NO agent delegation

# ✅ GOOD: Command orchestrates agents
/alfred:2-run SPEC-001
  → alfred delegates to plan-agent
  → plan-agent creates strategy
  → alfred delegates to tdd-implementer
  → tdd-implementer writes code
  → alfred delegates to test-engineer
  → test-engineer validates tests
```

---

### ❌ Anti-Pattern 2: Over-Delegation

**Problem**: Excessive agent handoffs for simple tasks

```python
# ❌ BAD: 5 agents for simple README update
alfred → plan-agent → doc-syncer → qa-validator → git-manager
(10 minutes for 2-line change)

# ✅ GOOD: Direct delegation for simple tasks
alfred → doc-syncer (with simple validation)
(30 seconds for 2-line change)
```

**Rule**: Use single agent for simple, well-defined tasks.

---

### ❌ Anti-Pattern 3: Context Leakage

**Problem**: Agents receiving irrelevant context

```python
# ❌ BAD: Frontend agent gets backend implementation details
Task(
    subagent_type="frontend-expert",
    description="Build user dashboard",
    context_includes=["backend_database_schema", "backend_orm_models", "backend_api_internals"]
    # Irrelevant context wastes tokens
)

# ✅ GOOD: Frontend agent gets only API contract
Task(
    subagent_type="frontend-expert",
    description="Build user dashboard",
    context_includes=["api_contract", "component_patterns", "state_management"]
    # Relevant context only
)
```

---

### ❌ Anti-Pattern 4: Silent Failures

**Problem**: Agent errors not reported or handled

```python
# ❌ BAD: No error handling
result = Task(subagent_type="tdd-implementer", ...)
# If agent fails, workflow continues with invalid state

# ✅ GOOD: Explicit error handling
try:
    result = Task(subagent_type="tdd-implementer", ...)
    validate_agent_output(result)
except AgentError as e:
    # Escalate to debug-helper
    Task(subagent_type="debug-helper", description=f"Investigate failure: {e}")
```

---

## 🎯 Best Practices Checklist

**Must-Have (Agent Selection)**:
- ✅ Always delegate executable tasks to specialized agents
- ✅ Use decision tree for systematic agent selection
- ✅ Optimize model selection (Sonnet for reasoning, Haiku for execution)
- ✅ Validate agent outputs before handoffs
- ✅ Document assumptions and constraints

**Recommended (Orchestration)**:
- ✅ Use parallel execution for independent tasks
- ✅ Scope agent contexts to domain knowledge
- ✅ Implement error handling and escalation
- ✅ Track agent performance metrics
- ✅ Maintain agent collaboration protocols

**Security (Agent Operations)**:
- 🔒 Validate agent permissions (file access, API calls)
- 🔒 Sanitize agent inputs (prevent prompt injection)
- 🔒 Audit agent actions (traceability with @TAGs)
- 🔒 Implement agent rate limiting (prevent runaway loops)

---

## 🔗 Integration with Other Skills

**Prerequisite Skills**:
- Skill("moai-alfred-workflow") – Understand 4-step workflow before agent selection
- Skill("moai-foundation-tags") – @TAG system for traceability

**Complementary Skills**:
- Skill("moai-alfred-context-budget") – Optimize agent context loading
- Skill("moai-alfred-todowrite-pattern") – Track multi-agent task progress
- Skill("moai-alfred-best-practices") – TRUST 5 principles for agent work

**Next Steps**:
- Skill("moai-domain-backend") – Backend-specific agent guidance
- Skill("moai-domain-frontend") – Frontend-specific agent guidance
- Skill("moai-domain-testing") – Testing agent best practices

---

## 📈 Version History

**v5.0.0** (2025-11-12)
- ✨ Complete 19-agent team roster with detailed role descriptions
- ✨ Agent selection decision tree and flowcharts
- ✨ Haiku vs Sonnet model optimization guide
- ✨ 2025 multi-agent orchestration patterns (5 patterns)
- ✨ Agent collaboration protocols and handoff sequences
- ✨ Performance optimization strategies
- ✨ 15+ comprehensive code examples
- ✨ Advanced orchestration scenarios
- ✨ Anti-pattern catalog with solutions

**v4.0.0** (2025-11-12)
- ✨ Context7 MCP integration
- ✨ Progressive Disclosure structure
- ✨ Basic agent delegation patterns

---

**Generated with**: MoAI-ADK Skill Factory v5.0  
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (alfred)
