---
name: moai-workflow-spec
description: SPEC workflow orchestration with EARS format, requirement clarification, and Plan-Run-Sync integration for MoAI-ADK development methodology
version: 1.0.0
category: workflow
tags:
 - workflow
 - spec
 - ears
 - requirements
 - moai-adk
 - planning
updated: 2025-12-07
status: active
author: MoAI-ADK Team
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
---

# SPEC Workflow Management

## Quick Reference (30 seconds)

SPEC Workflow Orchestration - Comprehensive specification management using EARS format for systematic requirement definition and Plan-Run-Sync workflow integration.

Core Capabilities:
- EARS Format Specifications: Five requirement patterns for clarity
- Requirement Clarification: Four-step systematic process
- SPEC Document Templates: Standardized structure for consistency
- Plan-Run-Sync Integration: Seamless workflow connection
- Parallel Development: Git Worktree-based SPEC isolation
- Quality Gates: TRUST 5 framework validation

EARS Five Patterns:
```
Ubiquitous:    "시스템은 항상 [동작]해야 한다" - Always perform
Event-Driven:  "WHEN [이벤트] THEN [동작]" - Trigger-response
State-Driven:  "IF [조건] THEN [동작]" - Conditional behavior
Unwanted:      "시스템은 [동작]하지 않아야 한다" - Prohibition
Optional:      "가능하면 [동작]을 제공한다" - Nice-to-have
```

When to Use:
- Feature planning and requirement definition
- SPEC document creation and maintenance
- Parallel feature development coordination
- Quality assurance and validation planning

Quick Commands:
```bash
# Create new SPEC
/moai:1-plan "user authentication system"

# Create parallel SPECs with Worktrees
/moai:1-plan "login feature" "signup feature" --worktree

# Create SPEC with new branch
/moai:1-plan "payment processing" --branch

# Update existing SPEC
/moai:1-plan SPEC-001 "add OAuth support"
```

---

## Implementation Guide (5 minutes)

### Core Concepts

SPEC-First Development Philosophy:
- EARS format ensures unambiguous requirements
- Requirement clarification prevents scope creep
- Systematic validation through test scenarios
- Integration with TDD workflow for implementation
- Quality gates enforce completion criteria
- Constitution reference ensures project-wide consistency

### Constitution Reference (SDD 2025 Standard)

Constitution defines the project DNA that all SPECs must respect. Before creating any SPEC, verify alignment with project constitution defined in `.moai/project/tech.md` (extended with Constitution section).

Constitution Components:
- Technology Stack: Required versions and frameworks (e.g., Python 3.13+, FastAPI 0.118+)
- Naming Conventions: Variable, function, and file naming standards
- Forbidden Libraries: Libraries explicitly prohibited (with alternatives)
- Architectural Patterns: Layering rules, dependency directions
- Security Standards: Authentication patterns, encryption requirements
- Logging Standards: Log format, structured logging requirements

Constitution Verification Checklist:
- All SPEC technology choices align with Constitution stack versions
- No SPEC introduces forbidden libraries or patterns
- SPEC follows naming conventions defined in Constitution
- SPEC respects architectural boundaries and layering

WHY: Constitution prevents architectural drift and ensures maintainability
IMPACT: SPECs aligned with Constitution reduce integration conflicts by 60%

SPEC Workflow Stages:
1. User Input Analysis: Parse natural language feature description
2. Requirement Clarification: Four-step systematic process
3. EARS Pattern Application: Structure requirements using five patterns
4. Success Criteria Definition: Establish completion metrics
5. Test Scenario Generation: Create verification test cases
6. SPEC Document Generation: Produce standardized markdown output

### EARS Format Deep Dive

**Ubiquitous Requirements** - Always Active:
- Pattern: "시스템은 항상 [동작]해야 한다"
- Use Case: System-wide quality attributes
- Examples:
  - "시스템은 항상 로그를 기록해야 한다" (logging)
  - "시스템은 항상 사용자 입력을 검증해야 한다" (input validation)
  - "시스템은 항상 에러 메시지를 표시해야 한다" (error handling)
- Test Strategy: Include in all feature test suites as common verification

**Event-Driven Requirements** - Trigger-Response:
- Pattern: "WHEN [이벤트]가 발생하면 THEN [동작]한다"
- Use Case: User interactions and inter-system communication
- Examples:
  - "WHEN 사용자가 로그인 버튼을 클릭하면 THEN 인증을 시도한다"
  - "WHEN 파일이 업로드되면 THEN 바이러스 스캔을 실행한다"
  - "WHEN 결제가 완료되면 THEN 영수증을 발송한다"
- Test Strategy: Event simulation with expected response verification

**State-Driven Requirements** - Conditional Behavior:
- Pattern: "IF [조건]이면 THEN [동작]한다"
- Use Case: Access control, state machines, conditional business logic
- Examples:
  - "IF 계정이 활성 상태이면 THEN 로그인을 허용한다"
  - "IF 재고가 있으면 THEN 주문을 처리한다"
  - "IF 관리자 권한이면 THEN 삭제를 허용한다"
- Test Strategy: State setup with conditional behavior verification

**Unwanted Requirements** - Prohibited Actions:
- Pattern: "시스템은 [동작]하지 않아야 한다"
- Use Case: Security vulnerabilities, data integrity protection
- Examples:
  - "시스템은 평문 비밀번호를 저장하지 않아야 한다"
  - "시스템은 인증되지 않은 접근을 허용하지 않아야 한다"
  - "시스템은 민감 정보를 로그에 기록하지 않아야 한다"
- Test Strategy: Negative test cases with prohibited behavior verification

**Optional Requirements** - Enhancement Features:
- Pattern: "가능하면 [동작]을 제공한다"
- Use Case: MVP scope definition, feature prioritization
- Examples:
  - "가능하면 OAuth 로그인을 제공한다"
  - "가능하면 다크 모드를 지원한다"
  - "가능하면 오프라인 모드를 제공한다"
- Test Strategy: Conditional test execution based on implementation status

### Requirement Clarification Process

**Step 0: Assumption Analysis (Philosopher Framework)** [NEW]

Before defining scope, surface and validate underlying assumptions using AskUserQuestion:

Assumption Categories to Examine:
- Technical Assumptions: Technology capabilities, API availability, performance characteristics
- Business Assumptions: User behavior, market requirements, timeline feasibility
- Team Assumptions: Skill availability, resource allocation, knowledge gaps
- Integration Assumptions: Third-party service reliability, compatibility expectations

Assumption Documentation Format:
- Assumption Statement: Clear description of what is assumed
- Confidence Level: High, Medium, or Low based on evidence
- Evidence Basis: What supports this assumption
- Risk if Wrong: Consequence if assumption proves false
- Validation Method: How to verify before committing significant effort

Example Assumption Analysis:
```markdown
## Assumptions Declared

| # | Assumption | Confidence | Risk if Wrong |
|---|-----------|------------|---------------|
| 1 | Users have stable internet | Medium | Need offline mode |
| 2 | OAuth provider maintains API compatibility | High | Migration needed |
| 3 | Team familiar with JWT patterns | Low | Training required |
```

Use AskUserQuestion to verify critical assumptions before proceeding to scope definition.

**Step 0.5: Root Cause Analysis** [NEW]

For feature requests or problem-driven SPECs, apply Five Whys:
- Surface Problem: What is the user observing or requesting?
- First Why: What immediate need drives this request?
- Second Why: What underlying problem creates that need?
- Third Why: What systemic factor contributes?
- Root Cause: What fundamental issue must the solution address?

Alternative Approaches Section (SPEC Document):
```markdown
## Approaches Considered

| Approach | Pros | Cons | Selected |
|----------|------|------|----------|
| Option A | ... | ... | Yes |
| Option B | ... | ... | No - higher complexity |
| Option C | ... | ... | No - vendor lock-in |
```

**Step 1: Scope Definition**
- Identify supported authentication methods (email/password, OAuth, SSO)
- Define password complexity rules and validation
- Determine login failure handling strategy
- Establish session management approach

**Step 2: Constraint Extraction**
- Performance Requirements: Response time targets (e.g., 500ms P95)
- Security Requirements: OWASP compliance, encryption standards
- Compatibility Requirements: Supported browsers, mobile devices
- Scalability Requirements: Concurrent user targets

**Step 3: Success Criteria Definition**
- Test Coverage: Minimum 85% code coverage target
- Response Time: P50 < 50ms, P95 < 200ms, P99 < 500ms
- Functional Completion: All normal scenarios pass verification
- Quality Gates: Zero linter warnings, zero security vulnerabilities

**Step 4: Test Scenario Creation**
- Normal Cases: Valid inputs with expected outputs
- Error Cases: Invalid inputs with error handling
- Edge Cases: Boundary conditions and corner cases
- Security Cases: Injection attacks, privilege escalation attempts

### SPEC Document Structure

**Header Section**:
```markdown
# SPEC-001: User Authentication System

Created: 2025-12-07
Status: Planned
Priority: High
Assigned: manager-tdd
Related SPECs: SPEC-002 (User Registration)
```

**Requirements Section** (EARS Format):
```markdown
## Requirements

### Ubiquitous
- 시스템은 항상 로그인 시도를 로깅해야 한다
- 시스템은 항상 비밀번호를 해싱하여 저장해야 한다

### Event-Driven
- WHEN 사용자가 로그인 버튼을 클릭하면 THEN 자격증명을 검증한다
- WHEN 로그인이 성공하면 THEN JWT 토큰을 발급한다

### State-Driven
- IF 계정이 활성 상태이면 THEN 로그인을 허용한다
- IF 로그인 실패 횟수가 5회 이상이면 THEN 계정을 일시 잠금한다

### Unwanted
- 시스템은 평문 비밀번호를 로그에 기록하지 않아야 한다
- 시스템은 토큰 없이 보호된 리소스 접근을 허용하지 않아야 한다

### Optional
- 가능하면 OAuth 2.0 소셜 로그인을 제공한다
- 가능하면 이중 인증(2FA)을 지원한다
```

**Constraints Section**:
```markdown
## Constraints

Technical Constraints:
- Backend: Node.js 20+, Express.js framework
- Database: PostgreSQL 15+ for user credentials
- Authentication: JWT with RS256 algorithm
- Password Hashing: bcrypt with salt rounds 12

Business Constraints:
- Session timeout: 24 hours for standard users, 1 hour for admin
- Password complexity: Minimum 8 characters, mixed case, numbers, symbols
- Login attempt limit: 5 failures trigger 15-minute account lockout
```

**Success Criteria Section**:
```markdown
## Success Criteria

Functional Criteria:
- All EARS requirements implemented and verified
- Test coverage >= 85% for authentication module
- All test scenarios pass with expected results

Performance Criteria:
- Login response time P95 < 200ms
- Token generation time < 50ms
- Password hashing time < 500ms

Security Criteria:
- OWASP Authentication Cheat Sheet compliance
- No SQL injection vulnerabilities (verified by SQLMap)
- No XSS vulnerabilities (verified by OWASP ZAP)
```

**Test Scenarios Section**:
```markdown
## Test Scenarios

| ID | Category | Scenario | Input | Expected | Status |
|---|---|---|---|---|---|
| TC-1 | Normal | Valid login | email+password | JWT token, 200 | Pending |
| TC-2 | Error | Invalid password | wrong password | 401 error | Pending |
| TC-3 | Error | Nonexistent user | unknown email | 401 error | Pending |
| TC-4 | Edge | Empty password | empty string | 400 error | Pending |
| TC-5 | Security | SQL injection | ' OR '1'='1 | 400 error, blocked | Pending |
| TC-6 | State | Locked account | valid credentials | 403 error | Pending |
| TC-7 | Performance | Concurrent logins | 100 requests/sec | < 200ms P95 | Pending |
```

### Plan-Run-Sync Workflow Integration

**PLAN Phase** (/moai:1-plan):
- manager-spec agent analyzes user input
- EARS format requirements generation
- Requirement clarification with user interaction
- SPEC document creation in .moai/specs/ directory
- Git branch creation (optional --branch flag)
- Git Worktree setup (optional --worktree flag)

**RUN Phase** (/moai:2-run):
- manager-tdd agent loads SPEC document
- RED-GREEN-REFACTOR TDD cycle execution
- moai-workflow-testing skill reference for test patterns
- Domain Expert agent delegation (expert-backend, expert-frontend, etc.)
- Quality validation through manager-quality agent

**SYNC Phase** (/moai:3-sync):
- manager-docs agent synchronizes documentation
- API documentation generation from SPEC
- README and architecture document updates
- CHANGELOG entry creation
- Version control commit with SPEC reference

### Parallel Development with Git Worktree

**Worktree Concept**:
- Independent working directories for multiple branches
- Each SPEC gets isolated development environment
- No branch switching needed for parallel work
- Reduced merge conflicts through feature isolation

**Worktree Creation Process**:
```bash
# Command creates two SPECs with Worktrees
/moai:1-plan "login feature" "signup feature" --worktree

# Result directory structure:
# /project                    (main branch)
# /project-worktrees/SPEC-001 (login feature branch)
# /project-worktrees/SPEC-002 (signup feature branch)
```

**Worktree Benefits**:
- Parallel Development: Multiple features developed simultaneously
- Team Collaboration: Clear ownership boundaries per SPEC
- Dependency Isolation: Different library versions per feature
- Risk Reduction: Unstable code doesn't affect other features

**Worktree Cleanup**:
```bash
# After feature completion and merge
git worktree remove /project-worktrees/SPEC-001
git branch -d SPEC-001-login-feature
```

---

## Advanced Implementation (10+ minutes)

For advanced patterns including SPEC templates, quality validation, and workflow optimization, see:

- [Advanced Patterns](advanced-patterns.md): Custom SPEC templates, validation automation
- [Reference Guide](reference.md): SPEC metadata schema, integration examples
- [Examples](examples.md): Real-world SPEC documents, workflow scenarios

## Resources

### SPEC File Organization

Directory Structure:
```
.moai/
├── specs/
│   ├── SPEC-001-user-authentication.md
│   ├── SPEC-002-user-registration.md
│   └── SPEC-003-password-reset.md
├── memory/
│   └── last-session-state.json
└── docs/
    └── api-documentation.md
```

### SPEC Metadata Schema

Required Fields:
- SPEC ID: Sequential number (SPEC-001, SPEC-002, etc.)
- Title: Feature name in English
- Created: ISO 8601 timestamp
- Status: Planned, In Progress, Completed, Blocked
- Priority: High, Medium, Low
- Assigned: Agent responsible for implementation

Optional Fields:
- Related SPECs: Dependencies and related features
- Epic: Parent feature group
- Estimated Effort: Time estimate in hours or story points
- Labels: Tags for categorization

### SPEC Lifecycle Management (SDD 2025 Standard)

Lifecycle Level Field (NEW):

SPEC maturity follows three levels based on SDD 2025 best practices:

Level 1 - spec-first:
- Description: SPEC written before implementation, discarded after completion
- Use Case: One-time features, prototypes, experiments
- Maintenance Policy: No maintenance required after implementation
- Recommended For: Quick iterations, exploratory development

Level 2 - spec-anchored:
- Description: SPEC maintained alongside implementation for evolution and maintenance
- Use Case: Core features, API contracts, integration points
- Maintenance Policy: Quarterly review, update when implementation changes
- Recommended For: Production features requiring long-term maintenance

Level 3 - spec-as-source:
- Description: SPEC is the single source of truth; only SPEC is edited by humans
- Use Case: Critical systems, regulated environments, code generation workflows
- Maintenance Policy: SPEC changes trigger implementation regeneration
- Recommended For: High-reliability systems, compliance-required features

SPEC Document Header with Lifecycle:
```markdown
---
id: SPEC-AUTH-001
version: "1.0.0"
status: "planned"
created: "2025-12-19"
updated: "2025-12-19"
author: "developer"
priority: "high"
lifecycle_level: "spec-anchored"
maintenance_policy: "quarterly_review"
---
```

Lifecycle Transition Rules:
- spec-first to spec-anchored: When feature becomes production-critical
- spec-anchored to spec-as-source: When compliance or regeneration workflow required
- Downgrade allowed but requires explicit justification in SPEC history

Spec Drift Prevention:
- For spec-anchored and spec-as-source levels, /moai:3-sync validates SPEC-implementation alignment
- Drift detection triggers warning before documentation sync
- Manual override requires explicit acknowledgment

### Quality Metrics

SPEC Quality Indicators:
- Requirement Clarity: All EARS patterns used appropriately
- Test Coverage: All requirements have corresponding test scenarios
- Constraint Completeness: Technical and business constraints defined
- Success Criteria Measurability: Quantifiable completion metrics

Validation Checklist:
- All EARS requirements testable
- No ambiguous language ("should", "might", "usually")
- All error cases documented
- Performance targets quantified
- Security requirements OWASP-compliant

### Works Well With

- moai-foundation-core: SPEC-First TDD methodology and TRUST 5 framework
- moai-workflow-testing: TDD implementation and test automation
- moai-workflow-project: Project initialization and configuration
- moai-worktree: Git Worktree management for parallel development
- manager-spec: SPEC creation and requirement analysis agent
- manager-tdd: TDD implementation based on SPEC requirements
- manager-quality: TRUST 5 quality validation and gate enforcement

### Integration Examples

Sequential Workflow:
```bash
# Step 1: Plan - Create SPEC
/moai:1-plan "user authentication system"

# Step 2: Run - Implement with TDD
/moai:2-run SPEC-001

# Step 3: Sync - Update documentation
/moai:3-sync SPEC-001
```

Parallel Workflow:
```bash
# Create multiple SPECs with Worktrees
/moai:1-plan "backend API" "frontend UI" "database schema" --worktree

# Parallel implementation in separate sessions
# Session 1: /moai:2-run SPEC-001 (backend API)
# Session 2: /moai:2-run SPEC-002 (frontend UI)
# Session 3: /moai:2-run SPEC-003 (database schema)
```

### Token Management

Session Strategy:
- PLAN phase: Requirements analysis and SPEC generation (~30% tokens)
- Clear context: /clear command after SPEC document saved
- RUN phase: Fresh session for TDD implementation (~60% tokens)
- SYNC phase: Documentation update in final session (~10% tokens)

Context Optimization:
- SPEC document persists in .moai/specs/ directory
- Session memory in .moai/memory/ for cross-session context
- Minimal context transfer through SPEC ID reference
- Agent delegation reduces token overhead

---

Version: 1.2.0 (SDD 2025 Standard Integration)
Last Updated: 2025-12-19
Integration Status: Complete - Full Plan-Run-Sync workflow with SDD 2025 features (Constitution, Tasks Decomposition, Lifecycle Management)
