# SPEC Workflow Reference Guide

## Extended Documentation

This document provides comprehensive reference information for SPEC workflow management, including advanced patterns, integration guides, and troubleshooting procedures.

---

## SPEC Document Templates

### Template 1: Simple CRUD Feature

```markdown
# SPEC-XXX: [Feature Name]

Created: YYYY-MM-DD
Status: Planned
Priority: Medium
Assigned: manager-ddd

## Description
[Brief description of the feature]

## Requirements

### Ubiquitous
- 시스템은 항상 입력을 검증해야 한다
- 시스템은 항상 에러를 로깅해야 한다

### Event-Driven
- WHEN [event] THEN [action]

### State-Driven
- IF [condition] THEN [action]

### Unwanted
- 시스템은 [prohibited action]하지 않아야 한다

## Constraints

Technical:
- Framework: [technology stack]
- Database: [database system]

Business:
- [Business rule 1]
- [Business rule 2]

## Success Criteria

- Test coverage >= 85%
- Response time P95 < 200ms
- Zero security vulnerabilities

## Test Scenarios

| ID | Scenario | Input | Expected | Status |
|---|---|---|---|---|
| TC-1 | Normal case | [input] | [output] | Pending |
```

### Template 2: Complex Workflow Feature

```markdown
# SPEC-XXX: [Complex Feature Name]

Created: YYYY-MM-DD
Status: Planned
Priority: High
Assigned: manager-ddd
Related SPECs: SPEC-YYY, SPEC-ZZZ

## Description
[Detailed description with business context]

### Preconditions
1. [Precondition 1]
2. [Precondition 2]

### Side Effects
1. [Side effect 1]
2. [Side effect 2]

## Requirements

### Ubiquitous
- [System-wide requirements]

### Event-Driven
- WHEN [trigger event] THEN [immediate action]
- WHEN [completion event] THEN [notification action]

### State-Driven
- IF [state condition] THEN [allowed action]
- IF [state condition] THEN [blocked action]

### Unwanted
- 시스템은 [security vulnerability]하지 않아야 한다
- 시스템은 [data integrity issue]하지 않아야 한다

### Optional
- 가능하면 [enhancement feature]을 제공한다

## Constraints

Technical:
- Architecture: [microservices/monolith]
- Transaction: [ACID requirements]
- Performance: [latency targets]

Business:
- Compliance: [regulatory requirements]
- SLA: [service level agreement]

## Success Criteria

Functional:
- All preconditions validated
- All side effects executed in order
- Rollback mechanism for failures

Performance:
- P50 < [Xms]
- P95 < [Yms]
- P99 < [Zms]

Security:
- [Security requirement 1]
- [Security requirement 2]

## Test Scenarios

| ID | Category | Scenario | Input | Expected | Status |
|---|---|---|---|---|---|
| TC-1 | Normal | [happy path] | [input] | [output] | Pending |
| TC-2 | Error | [failure case] | [input] | [error] | Pending |
| TC-3 | Edge | [boundary case] | [input] | [output] | Pending |
| TC-4 | Security | [attack vector] | [input] | [blocked] | Pending |
```

### Template 3: API Endpoint SPEC

```markdown
# SPEC-XXX: [API Endpoint Name]

Created: YYYY-MM-DD
Status: Planned
Priority: Medium
Assigned: expert-backend

## API Definition

```
METHOD /api/v1/resource/{id}

Headers:
  Authorization: Bearer {token}
  Content-Type: application/json

Path Parameters:
  id: integer (1-999999)

Query Parameters:
  filter: string (optional)
  sort: string (optional)

Request Body:
{
  "field1": "value",
  "field2": 123
}
```

## Requirements

### Ubiquitous
- 시스템은 항상 인증을 검증해야 한다
- 시스템은 항상 요청을 로깅해야 한다

### Event-Driven
- WHEN 요청이 수신되면 THEN 스키마를 검증한다
- WHEN 검증이 통과하면 THEN 비즈니스 로직을 실행한다

### State-Driven
- IF 사용자가 인증되었으면 THEN 리소스 접근을 허용한다
- IF 권한이 있으면 THEN 변경 작업을 허용한다

### Unwanted
- 시스템은 SQL injection을 허용하지 않아야 한다
- 시스템은 민감 정보를 응답에 포함하지 않아야 한다

## Response Schemas

Success (200 OK):
```json
{
  "data": {
    "id": 123,
    "field1": "value"
  },
  "meta": {
    "timestamp": "2025-12-07T10:00:00Z"
  }
}
```

Error (400 Bad Request):
```json
{
  "error": "VALIDATION_ERROR",
  "message": "Field 'field1' is required",
  "details": [
    {"field": "field1", "issue": "required"}
  ]
}
```

## Constraints

Technical:
- Rate Limit: 100 requests/minute per user
- Timeout: 30 seconds
- Max Payload: 1MB

## Success Criteria

- OpenAPI 3.0 schema compliance
- Response time P95 < 100ms
- Request validation coverage 100%

## Test Scenarios

| ID | Scenario | Request | Expected | Status |
|---|---|---|---|---|
| TC-1 | Valid request | [full request] | 200 with data | Pending |
| TC-2 | Missing auth | no header | 401 error | Pending |
| TC-3 | Invalid schema | wrong type | 400 error | Pending |
| TC-4 | Rate limit | 101 requests | 429 error | Pending |
```

---

## SPEC Metadata Schema

### Core Fields

**SPEC ID Format**:
- Pattern: SPEC-XXX where XXX is zero-padded sequential number
- Examples: SPEC-001, SPEC-002, SPEC-042
- Range: SPEC-001 to SPEC-999
- Auto-increment: Managed by manager-spec agent

**Title Format**:
- Language: English
- Capitalization: Title Case
- Format: Noun Phrase describing feature
- Examples: "User Authentication System", "Payment Processing API"

**Status Values**:
- Planned: SPEC created, not yet started
- In Progress: Implementation in RUN phase
- Completed: All success criteria met
- Blocked: Waiting for dependency or decision
- Deprecated: Replaced by newer SPEC

**Priority Levels**:
- High: Critical for MVP, blocking dependencies
- Medium: Important but not blocking
- Low: Enhancement or optional feature

**Assigned Agents**:
- manager-ddd: DDD-based implementation
- manager-spec: SPEC refinement and updates
- expert-backend: Backend-specific features
- expert-frontend: Frontend-specific features
- expert-database: Database schema changes

### Extended Fields

**Related SPECs**:
- Format: Comma-separated SPEC IDs
- Types:
  - Depends On: Required prerequisite SPECs
  - Blocks: SPECs waiting for this SPEC
  - Related: Conceptually connected SPECs
- Example: "Depends On: SPEC-001, Blocks: SPEC-005, SPEC-006"

**Epic**:
- Parent feature group identifier
- Format: EPIC-XXX or feature name
- Use Case: Grouping related SPECs for large features
- Example: "EPIC-AUTH" for authentication-related SPECs

**Estimated Effort**:
- Units: Hours, Story Points, or T-Shirt Sizes
- Format: Numeric value with unit
- Examples: "8 hours", "5 story points", "Large"

**Labels**:
- Format: Comma-separated tags
- Categories: domain, technology, priority, type
- Examples: "backend, security, high-priority, api"

**Version**:
- Format: Semantic versioning (MAJOR.MINOR.PATCH)
- Initial: 1.0.0
- Increment Rules:
  - MAJOR: Breaking changes to requirements
  - MINOR: New requirements added
  - PATCH: Clarifications or corrections

---

## EARS Pattern Selection Guide

### Ubiquitous Pattern Selection

**Use When**:
- Requirement applies to all system operations
- Quality attribute must be system-wide
- No exceptions or conditions exist

**Common Use Cases**:
- Logging and monitoring
- Security measures (authentication, authorization)
- Error handling and recovery
- Data validation
- Audit trails

**Anti-Patterns to Avoid**:
- Don't use for feature-specific requirements
- Don't use when conditions or triggers exist
- Don't use for optional features

### Event-Driven Pattern Selection

**Use When**:
- User action triggers system response
- External event requires system reaction
- Asynchronous processing needed

**Common Use Cases**:
- Button clicks and user interactions
- File uploads and processing
- Webhook callbacks
- Message queue processing
- Real-time notifications

**Anti-Patterns to Avoid**:
- Don't use for state-based conditions
- Don't use for continuous monitoring
- Don't confuse with state-driven patterns

### State-Driven Pattern Selection

**Use When**:
- System behavior depends on current state
- Access control based on user role or status
- Conditional business logic exists

**Common Use Cases**:
- Permission checks (role, status, subscription)
- Order processing (pending, paid, shipped)
- Account states (active, suspended, deleted)
- Feature flags and A/B testing

**Anti-Patterns to Avoid**:
- Don't use for simple event responses
- Don't use for system-wide requirements
- Don't confuse with event-driven patterns

### Unwanted Pattern Selection

**Use When**:
- Security vulnerability must be prevented
- Data integrity must be protected
- Compliance violation must be blocked

**Common Use Cases**:
- Password storage (no plaintext)
- SQL injection prevention
- XSS attack blocking
- PII exposure in logs
- Unauthorized access

**Anti-Patterns to Avoid**:
- Don't use for positive requirements
- Don't use for optional restrictions
- Don't duplicate state-driven conditions

### Optional Pattern Selection

**Use When**:
- Feature enhances UX but isn't required
- MVP scope needs clear boundaries
- Future enhancement is planned

**Common Use Cases**:
- Social login (OAuth, SAML)
- Advanced UI features (dark mode, animations)
- Performance optimizations
- Additional export formats

**Anti-Patterns to Avoid**:
- Don't use for core functionality
- Don't use to avoid decision-making
- Don't confuse with unwanted requirements

---

## Quality Validation Checklist

### SPEC Quality Criteria

**Clarity (Score: 0-100)**:
- All requirements use EARS patterns correctly
- No ambiguous language ("should", "might", "usually")
- Technical terms defined or referenced
- Success criteria quantifiable and measurable

**Completeness (Score: 0-100)**:
- All EARS patterns considered (even if empty)
- All error cases documented
- Performance targets specified
- Security requirements defined
- Test scenarios cover all requirements

**Testability (Score: 0-100)**:
- Every requirement has test scenario
- Test inputs and outputs specified
- Edge cases identified
- Negative test cases included

**Consistency (Score: 0-100)**:
- No conflicting requirements
- Terminology used consistently
- EARS pattern usage appropriate
- Constraint alignment with requirements

### Automated Validation Rules

**Rule 1: EARS Pattern Coverage**
- At least 3 of 5 EARS patterns used
- Ubiquitous or Event-Driven pattern present
- No mixing of patterns in single requirement

**Rule 2: Test Scenario Coverage**
- Minimum 5 test scenarios per SPEC
- At least 1 normal case
- At least 2 error cases
- At least 1 edge case

**Rule 3: Success Criteria Quantification**
- Performance targets include metrics (ms, %, count)
- Test coverage target >= 85%
- All criteria measurable and verifiable

**Rule 4: Constraint Specification**
- Technical constraints defined
- Business constraints documented
- No contradictions between constraints and requirements

---

## Troubleshooting Guide

### Common SPEC Issues

**Issue 1: Ambiguous Requirements**
- Symptom: Implementation varies between developers
- Cause: Unclear language or missing details
- Solution: Apply EARS patterns strictly, add examples

**Issue 2: Missing Error Cases**
- Symptom: Production bugs not caught by tests
- Cause: Incomplete error scenario analysis
- Solution: Systematic error case brainstorming, security review

**Issue 3: Untestable Success Criteria**
- Symptom: Cannot determine when feature is complete
- Cause: Vague or qualitative criteria
- Solution: Quantify all metrics, define measurement methods

**Issue 4: Conflicting Requirements**
- Symptom: Cannot satisfy all requirements simultaneously
- Cause: Requirements defined without holistic view
- Solution: Conflict resolution session, priority clarification

**Issue 5: Scope Creep**
- Symptom: SPEC grows during implementation
- Cause: New requirements added without SPEC update
- Solution: Strict change control, new SPEC for additions

### SPEC Update Process

**When to Update SPEC**:
- Requirement change requested by stakeholder
- Implementation reveals missing requirements
- Performance targets need adjustment
- Security vulnerabilities discovered

**Update Procedure**:
1. Create SPEC update request with justification
2. Analyze impact on existing implementation
3. Update SPEC with version increment
4. Notify affected agents and teams
5. Re-run affected test scenarios
6. Update documentation in SYNC phase

**Version Control**:
- Commit SPEC changes with descriptive message
- Tag SPEC versions for major updates
- Maintain CHANGELOG in SPEC directory
- Link commits to SPEC IDs

---

## Integration Patterns

### Sequential Integration (Single Feature)

```
User Request
    ↓
/moai:1-plan "feature description"
    ↓
manager-spec creates SPEC-001
    ↓
/clear (token optimization)
    ↓
/moai:2-run SPEC-001
    ↓
manager-ddd implements with ANALYZE-PRESERVE-IMPROVE
    ↓
/moai:3-sync SPEC-001
    ↓
manager-docs updates documentation
    ↓
Feature Complete
```

### Parallel Integration (Multiple Features)

```
User Request
    ↓
/moai:1-plan "feature1" "feature2" "feature3" --worktree
    ↓
manager-spec creates SPEC-001, SPEC-002, SPEC-003
    ↓
Git Worktree setup for parallel development
    ↓
/clear (token optimization)
    ↓
┌─────────────┬─────────────┬─────────────┐
│ Session 1   │ Session 2   │ Session 3   │
│ SPEC-001    │ SPEC-002    │ SPEC-003    │
│ /moai:2-run │ /moai:2-run │ /moai:2-run │
└─────────────┴─────────────┴─────────────┘
    ↓
Worktree merge to main branch
    ↓
/moai:3-sync SPEC-001 SPEC-002 SPEC-003
    ↓
All Features Complete
```

### Dependency Chain Integration

```
/moai:1-plan "database schema" --branch
    ↓
SPEC-001 created (foundation)
    ↓
/moai:2-run SPEC-001
    ↓
Database schema implemented
    ↓
/moai:1-plan "backend API" --branch
    ↓
SPEC-002 created (depends on SPEC-001)
    ↓
/moai:2-run SPEC-002
    ↓
Backend API implemented
    ↓
/moai:1-plan "frontend UI" --branch
    ↓
SPEC-003 created (depends on SPEC-002)
    ↓
/moai:2-run SPEC-003
    ↓
Frontend UI implemented
    ↓
/moai:3-sync SPEC-001 SPEC-002 SPEC-003
    ↓
Full Stack Feature Complete
```

---

## Performance Optimization

### Token Budget Management

**PLAN Phase Token Usage** (~30% of 200K):
- User input analysis: 5K tokens
- Requirement clarification dialogue: 15K tokens
- EARS pattern generation: 10K tokens
- SPEC document writing: 10K tokens
- Git operations: 5K tokens
- Buffer: 15K tokens

**Strategy**: Execute /clear after SPEC document saved to disk

**RUN Phase Token Usage** (~60% of 200K):
- SPEC document loading: 5K tokens
- DDD cycle execution: 100K tokens
- Code generation: 20K tokens
- Test execution and debugging: 15K tokens
- Quality validation: 10K tokens
- Buffer: 30K tokens

**SYNC Phase Token Usage** (~10% of 200K):
- Documentation generation: 10K tokens
- API spec updates: 5K tokens
- Commit message generation: 2K tokens
- Buffer: 3K tokens

### Session Management Strategy

**Single-Session Approach** (Simple Features):
- Complete PLAN-RUN-SYNC in one session
- No /clear needed if total < 150K tokens
- Best for small features (< 500 LOC)

**Multi-Session Approach** (Complex Features):
- Session 1: PLAN phase, /clear after SPEC saved
- Session 2: RUN phase, /clear after implementation
- Session 3: SYNC phase for documentation
- Best for large features (> 500 LOC)

**Parallel-Session Approach** (Multiple Features):
- Create all SPECs in Session 0, /clear
- Session 1-N: Each SPEC in separate session
- Final Session: Consolidated SYNC for all SPECs
- Best for > 3 independent features

---

## Best Practices

### SPEC Writing Best Practices

1. **Start with User Story**: Convert user story to EARS requirements
2. **One Requirement, One Sentence**: Keep each requirement atomic
3. **Use Concrete Examples**: Include example inputs and outputs
4. **Define Error Cases First**: Security and error handling upfront
5. **Quantify Everything**: Numbers over adjectives ("fast" → "< 200ms")

### Requirement Clarification Best Practices

1. **Ask Open Questions**: "What authentication methods?" not "Email/password only?"
2. **Validate Assumptions**: Confirm implicit requirements explicitly
3. **Use Domain Language**: Align terminology with user's context
4. **Document Decisions**: Record why certain approaches were chosen
5. **Iterate Incrementally**: Build SPEC through dialogue, not single pass

### Test Scenario Best Practices

1. **Normal-First Approach**: Start with happy path scenarios
2. **Error Enumeration**: Systematically list all error conditions
3. **Boundary Testing**: Test limits and edge values
4. **Security Testing**: Include attack vectors and vulnerability checks
5. **Performance Testing**: Add load and stress test scenarios

---

## External Resources

### EARS Format References

- Original EARS Paper: Mavin, A., et al. "Easy Approach to Requirements Syntax (EARS)"
- NASA Systems Engineering Handbook: Requirements definition using EARS
- IEEE Guide for Software Requirements Specifications (IEEE 830)

### Related Documentation

- MoAI-ADK Plan-Run-Sync Workflow: Core methodology documentation
- TRUST 5 Framework: Quality assurance and validation framework
- Git Worktree Documentation: Parallel development environment setup
- DDD Best Practices: Domain-Driven Development implementation guide

### Tool Integration

- SPEC Linters: Automated validation of EARS pattern usage
- Test Coverage Tools: pytest-cov, Istanbul, JaCoCo integration
- API Documentation: OpenAPI/Swagger generation from SPEC
- Continuous Integration: GitHub Actions, Jenkins, CircleCI integration

---

Version: 1.0.0
Last Updated: 2025-12-07
