---
name: "moai-alfred-spec-authoring"
version: "4.0.0"
created: 2025-10-23
updated: 2025-11-12
status: stable
tier: Alfred
description: >-
  Complete SPEC document authoring guide with YAML metadata structure (7 required + 9 optional fields),
  EARS requirement syntax (5 patterns including Unwanted Behaviors), version lifecycle management,
  TAG integration, pre-submission validation checklist, and real-world SPEC examples.
keywords: 
  - spec
  - authoring
  - ears
  - metadata
  - requirements
  - tdd
  - planning
  - yaml-metadata
  - requirement-syntax
  - validation-checklist
allowed-tools: 
  - Read
  - Bash
  - Glob
---

# SPEC Authoring Skill (Enterprise v4.0.0)

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-spec-authoring |
| **Version** | 4.0.0 (2025-11-12 Enterprise Release) |
| **Tier** | Alfred (Planning & Specification) |
| **Allowed tools** | Read, Bash, Glob |
| **Auto-load** | `/alfred:1-plan`, SPEC authoring, requirements discussions |
| **Coverage** | YAML metadata, EARS syntax, validation, lifecycle management |

---

## What It Does

Provides comprehensive guidance for authoring professional SPEC documents in MoAI-ADK using YAML frontmatter metadata, EARS (Event, Actor, System, Response) requirement syntax with Unwanted Behaviors support, version management lifecycle, TAG integration for traceability, and pre-submission validation checklists.

**Key Capabilities**:
- Complete YAML metadata structure (7 required + 9 optional fields)
- EARS syntax with 5 requirement patterns
- Unwanted Behaviors definition and enforcement
- Version lifecycle management (draft → active → deprecated → archived)
- TAG integration for SPEC→TEST→CODE→DOC traceability
- Pre-submission validation checklist
- Common pitfalls and anti-patterns
- Real-world SPEC examples with annotations

---

## When to Use

**Automatic Triggers**:
- `/alfred:1-plan` command execution
- SPEC document creation requests
- Requirements clarification discussions
- Feature planning sessions
- Change request handling

**Manual Invocation**:
- SPEC template guidance
- Metadata field clarification
- EARS syntax validation
- Version management questions
- TAG traceability setup

---

## YAML Metadata Structure (16 Fields)

### 7 Required Fields

```yaml
---
# SPEC identifier (auto-generated)
code: SPEC-001

# SPEC title (descriptive, 50-80 chars)
title: Add User Authentication with JWT

# SPEC status (draft | active | deprecated | archived)
status: stable

# Creation timestamp (ISO 8601: YYYY-MM-DD)
created_at: 2025-11-12

# Last updated timestamp (ISO 8601: YYYY-MM-DD)
updated_at: 2025-11-12

# Business priority (critical | high | medium | low)
priority: high

# Estimated effort in story points (1-13 scale)
effort: 8

---
```

### 9 Optional Fields

```yaml
# Version tracking (semantic versioning: major.minor.patch)
version: 1.0.0

# Deadline target date (ISO 8601: YYYY-MM-DD)
deadline: 2025-12-15

# Epic this SPEC belongs to (e.g., AUTH-01, ONBOARDING-02)
epic: AUTH-01

# Related SPEC codes (dependencies or conflicts)
depends_on:
  - SPEC-002
  - SPEC-003

# Affected domains (for routing to specialists)
domains:
  - backend
  - security
  - database

# Acceptance criteria complexity rating
acceptance_difficulty: high

# Rollback complexity rating (critical | high | medium | low)
rollback_risk: medium

# Risk assessment notes
risks: |
  - Security: JWT key rotation must be tested
  - Performance: Token validation on every request

# Custom tags for filtering/searching
tags:
  - authentication
  - security
  - jwt
  - users

---
```

---

## EARS Requirement Syntax (5 Patterns)

### Pattern 1: Universal (Always True)

**Syntax**:
```
SPEC: The [System] SHALL [Action]
```

**Example**:
```
SPEC-001-REQ-001: The authentication service SHALL validate
all JWT tokens using RS256 algorithm against the public key.

Related TEST:
- test_valid_jwt_with_rs256_signature
- test_invalid_jwt_with_wrong_algorithm
```

**When to Use**: Core system behavior, non-negotiable requirements

---

### Pattern 2: Conditional (If-Then)

**Syntax**:
```
SPEC: If [Condition], then the [System] SHALL [Action]
```

**Example**:
```
SPEC-001-REQ-002: If a JWT token has expired,
then the authentication service SHALL reject the request
and return HTTP 401 Unauthorized with error code TOKEN_EXPIRED.

Related TEST:
- test_expired_token_returns_401
- test_expired_token_error_message
```

**When to Use**: Behavior dependent on specific conditions

---

### Pattern 3: Unwanted Behavior (Negative Requirement)

**Syntax**:
```
SPEC: The [System] SHALL NOT [Action]
```

**Example**:
```
SPEC-001-REQ-003: The authentication service SHALL NOT
accept JWT tokens signed with symmetric algorithms (HS256, HS384, HS512)
in a production environment.

Related TEST:
- test_reject_hs256_signed_token
- test_reject_hs384_signed_token
- test_reject_hs512_signed_token
```

**When to Use**: Security constraints, forbidden operations, anti-patterns

---

### Pattern 4: Stakeholder (User Role-Specific)

**Syntax**:
```
SPEC: As a [User Role], I want [Feature] so that [Benefit]
```

**Example**:
```
SPEC-001-REQ-004: As an API consumer,
I want to pass JWT tokens in the Authorization header
so that my requests are authenticated without exposing tokens.

Related TEST:
- test_jwt_from_authorization_header
- test_jwt_in_query_param_rejected
- test_malformed_authorization_header
```

**When to Use**: User stories, feature requirements, stakeholder concerns

---

### Pattern 5: Boundary Condition (Edge Cases)

**Syntax**:
```
SPEC: [System] SHALL [Action] when [Boundary Condition]
```

**Example**:
```
SPEC-001-REQ-005: The authentication service SHALL return
HTTP 429 Too Many Requests when a single IP address
attempts more than 10 failed authentication attempts within 5 minutes.

Related TEST:
- test_rate_limit_after_10_failures
- test_rate_limit_window_5_minutes
- test_rate_limit_by_ip_address
```

**When to Use**: Edge cases, performance limits, resource constraints

---

## Unwanted Behaviors Section

**Critical** security and quality constraints that MUST be tested:

```yaml
unwanted_behaviors:
  security:
    - The system SHALL NOT store JWT secrets in source code
    - The system SHALL NOT log JWT tokens or sensitive claims
    - The system SHALL NOT accept mixed algorithm tokens

  performance:
    - The system SHALL NOT block on token validation (async pattern)
    - The system SHALL NOT cache tokens indefinitely

  reliability:
    - The system SHALL NOT fail authentication if secondary cache is down
    - The system SHALL NOT accept malformed JSON Web Tokens

  data_integrity:
    - The system SHALL NOT modify token claims during validation
    - The system SHALL NOT accept tokens from untrusted issuers
```

**Each Unwanted Behavior** requires:
1. Test case verifying non-occurrence
2. Security scanning (where applicable)
3. Performance profiling (where applicable)

---

## SPEC Document Structure

```markdown
---
# YAML Metadata (7 required + 9 optional fields)
---

# SPEC-XXX: [Title]

## Overview
[2-3 sentence summary of what this SPEC delivers]

## Requirements

### REQ-001 (Universal Pattern)
SPEC: The [System] SHALL...

### REQ-002 (Conditional Pattern)
SPEC: If [Condition], then...

### REQ-003 (Unwanted Behavior Pattern)
SPEC: The [System] SHALL NOT...

### REQ-004 (Stakeholder Pattern)
As a [User], I want...

### REQ-005 (Boundary Condition Pattern)
SPEC: [System] SHALL ... when [Boundary]

## Unwanted Behaviors

### Security Constraints
- Description of what MUST NOT happen
- Rationale / Risk
- Testing approach

### Performance Constraints
- Description of what MUST NOT happen
- Rationale / Impact
- Testing approach

## Acceptance Criteria

- [ ] All 5 REQ patterns implemented
- [ ] All Unwanted Behaviors tested
- [ ] Code coverage ≥85%
- [ ] Security scan passed (OWASP Top 10)
- [ ] Performance baseline met
- [ ] Documentation updated

## Implementation Notes

### Architecture Impact
[How this SPEC affects system design]

### Database Changes
[Any schema migrations required]

### Configuration
[New configuration parameters needed]

## Testing Strategy

### Unit Tests
[Framework-specific test approach]

### Integration Tests
[Cross-service testing]

### Security Tests
[OWASP Top 10 coverage]

## References

### Official Documentation
- [Link to official spec/standard]
- [Link to API docs]

### Related SPECs
- SPEC-XXX: [Related feature]
- SPEC-YYY: [Conflicting requirement]

### TAGs
- @SPEC: SPEC-001 (this document)
- @TEST: SPEC-001-TEST-001, SPEC-001-TEST-002, ...
- @CODE: SPEC-001-CODE-001, SPEC-001-CODE-002, ...

---
```

---

## Version Lifecycle Management

### Lifecycle States

```
DRAFT → ACTIVE → DEPRECATED → ARCHIVED
  ↓                  ↓
Under Review    Stable, In Use
  ↓                  ↓
Changes Expected    Changes Require
Feedback Pending    Major Version Bump
```

### Field Updates by State

| Field | Draft | Active | Deprecated | Archived |
|-------|-------|--------|------------|----------|
| `status` | draft | active | deprecated | archived |
| `updated_at` | Updated daily | Updated on changes | Updated when deprecated | Immutable |
| `version` | 0.1.0 | 1.0.0+ | 1.x (unchanged) | 1.x (unchanged) |
| `depends_on` | Flexible | Fixed | Marked deprecated | Immutable |

### State Transitions

**DRAFT → ACTIVE**:
- All acceptance criteria defined
- At least 2 reviewers approved
- No critical open issues
- `version` bumped to 1.0.0
- Status change commit created

**ACTIVE → DEPRECATED**:
- Marked with deprecation reason
- Migration path documented
- Replacement SPEC linked
- 30-day notice period
- Status change commit created

**DEPRECATED → ARCHIVED**:
- No active code references
- All dependent SPECs archived
- Historical record maintained
- No further changes allowed
- Archive commit created

---

## TAG Integration for Traceability

### TAG Structure (SPEC→TEST→CODE→DOC)

```
@SPEC: SPEC-001             # Main specification
  ↓
@TEST: SPEC-001-TEST-001    # Test case
@TEST: SPEC-001-TEST-002
  ↓
@CODE: SPEC-001-CODE-001    # Implementation
@CODE: SPEC-001-CODE-002
  ↓
@DOC: SPEC-001-DOC-001      # Documentation
```

### TAG Placement Rules

**SPEC Document**:
```markdown
---
# SPEC-001: Feature Name
# @SPEC: SPEC-001
---
```

**Test File**:
```python
# @TEST: SPEC-001-TEST-001
def test_requirement_001():
    """Test SPEC-001 REQ-001 universal pattern."""
    pass
```

**Implementation**:
```python
# @CODE: SPEC-001-CODE-001
def authenticate_user(token: str) -> bool:
    """Validate JWT token per SPEC-001."""
    pass
```

**Documentation**:
```markdown
<!-- @DOC: SPEC-001-DOC-001 -->

## Authentication Flow

Per SPEC-001, the system SHALL validate all JWT tokens...
```

---

## Pre-Submission Validation Checklist

### Metadata Validation

- [ ] `code` field filled (SPEC-XXX format)
- [ ] `title` is descriptive (50-80 characters)
- [ ] `status` is one of: draft | active | deprecated | archived
- [ ] `created_at` is ISO 8601 format (YYYY-MM-DD)
- [ ] `updated_at` matches actual update date
- [ ] `priority` is one of: critical | high | medium | low
- [ ] `effort` is between 1-13 (story points)

### Requirement Syntax

- [ ] At least 3 REQ patterns used (Universal, Conditional, Unwanted)
- [ ] Each REQ follows EARS syntax strictly
- [ ] Requirements are specific and testable
- [ ] No ambiguous language ("should", "may", "might")
- [ ] All REQs are actionable (have test cases)

### Unwanted Behaviors

- [ ] Security constraints listed (if applicable)
- [ ] Performance constraints listed (if applicable)
- [ ] Reliability constraints listed (if applicable)
- [ ] Data integrity constraints listed (if applicable)
- [ ] Each Unwanted Behavior has a test approach

### Acceptance Criteria

- [ ] All 5 EARS patterns implemented
- [ ] All Unwanted Behaviors testable
- [ ] Code coverage target ≥85% specified
- [ ] Security scan type specified
- [ ] Performance baseline defined (if applicable)

### TAG Integration

- [ ] @SPEC tag added to document header
- [ ] @TEST tags linked for each requirement
- [ ] @CODE tags reserved (to be filled during implementation)
- [ ] @DOC tags reserved (to be filled during sync)

### Documentation Quality

- [ ] Overview section is 2-3 sentences
- [ ] Architecture impact explained
- [ ] Database changes documented
- [ ] Configuration parameters listed
- [ ] Testing strategy defined
- [ ] Related SPECs referenced

### Final Review

- [ ] No TODO or placeholder text
- [ ] All links are valid (internal and external)
- [ ] Formatting is consistent (markdown syntax)
- [ ] No confidential information exposed
- [ ] Ready for team review

---

## Common Pitfalls & Anti-Patterns

### Anti-Pattern 1: Ambiguous Requirements

**Bad**:
```
SPEC-001-REQ-001: The system should authenticate users quickly.
```

**Good**:
```
SPEC-001-REQ-001: The authentication service SHALL validate JWT tokens
and return a response within 50ms on average,
with p99 latency not exceeding 200ms.
```

---

### Anti-Pattern 2: Vague Acceptance Criteria

**Bad**:
```
- The feature should work
- Tests should pass
- No obvious bugs
```

**Good**:
```
- [ ] All 12 test cases pass (unit + integration)
- [ ] Code coverage ≥85% (src/auth/validate.py)
- [ ] Security scan: OWASP Top 10 coverage complete
- [ ] Performance: JWT validation p99 latency ≤200ms
```

---

### Anti-Pattern 3: Missing Unwanted Behaviors

**Bad**:
```
# (No unwanted_behaviors section)
```

**Good**:
```
unwanted_behaviors:
  security:
    - The system SHALL NOT store plaintext passwords
    - The system SHALL NOT log authentication tokens

  performance:
    - The system SHALL NOT block main thread on token validation
```

---

### Anti-Pattern 4: Incomplete TAG Integration

**Bad**:
```
# SPEC-001: Feature Name
# (No @SPEC, @TEST, @CODE, @DOC tags)
```

**Good**:
```
---
# SPEC-001: Feature Name
# @SPEC: SPEC-001
# Related: @TEST: SPEC-001-TEST-001,002,003
# Implementation: @CODE: SPEC-001-CODE-001,002
# Documentation: @DOC: SPEC-001-DOC-001
---
```

---

### Anti-Pattern 5: Misaligned Effort Estimation

**Bad**:
```
effort: 2  # Simple text change but 10 database migrations needed
```

**Good**:
```
effort: 8  # Includes: 3 code files, 2 test suites, 
           # 1 database migration, 4 hours review/testing
```

---

## SPEC Review Checklist (For Reviewers)

### Requirements Review

- [ ] All requirements follow EARS patterns
- [ ] Requirements are testable (can write automated tests)
- [ ] No conflicting requirements
- [ ] Scope is reasonable for effort estimate

### Implementation Review

- [ ] Architecture decisions are documented
- [ ] Database changes are migration-safe
- [ ] Configuration is production-ready
- [ ] Deprecation path is clear (if replacing existing feature)

### Testing Review

- [ ] Testing strategy covers all requirements
- [ ] Security testing approach is appropriate
- [ ] Performance baselines are realistic
- [ ] Edge cases (boundary conditions) are covered

### Quality Review

- [ ] Documentation is comprehensive
- [ ] Links are valid
- [ ] Metadata is complete and accurate
- [ ] Formatting is professional

---

## Real-World SPEC Example Structure

### Example 1: Feature SPEC with All Patterns

**SPEC-105: Email Notification Service**

```yaml
---
code: SPEC-105
title: Email Notification Service with Template Engine
status: stable
created_at: 2025-11-12
updated_at: 2025-11-12
priority: high
effort: 13
version: 1.0.0
epic: NOTIFICATIONS-01
depends_on:
  - SPEC-104  # User profiles
  - SPEC-102  # Email template schema
rollback_risk: high
---

# SPEC-105: Email Notification Service

## Overview
Implement asynchronous email notification service supporting templated
messages, retry logic, and delivery tracking for user notifications.

## Requirements

### REQ-001 (Universal)
SPEC: The notification service SHALL send emails asynchronously
without blocking the calling request.
@TEST: SPEC-105-TEST-001, SPEC-105-TEST-002

### REQ-002 (Conditional)
SPEC: If email delivery fails, the service SHALL retry up to 3 times
with exponential backoff (1s, 2s, 4s) before marking as failed.
@TEST: SPEC-105-TEST-003, SPEC-105-TEST-004

### REQ-003 (Unwanted Behavior)
SPEC: The service SHALL NOT send duplicate emails to the same recipient
within a 5-minute window.
@TEST: SPEC-105-TEST-005

### REQ-004 (Stakeholder)
As an application developer, I want to send templated emails
so that I don't have to manage HTML email formatting.
@TEST: SPEC-105-TEST-006, SPEC-105-TEST-007

### REQ-005 (Boundary Condition)
SPEC: The notification service SHALL process at least 1,000 emails/second
and SHALL NOT exceed 500MB memory usage under sustained load.
@TEST: SPEC-105-TEST-008 (load test)

## Unwanted Behaviors

### Security Constraints
- The system SHALL NOT expose email addresses in error logs
- The system SHALL NOT store plaintext passwords in configuration
- The system SHALL NOT process unverified sender addresses

### Performance Constraints
- The system SHALL NOT block on SMTP connection establishment
- The system SHALL NOT load entire email templates into memory for large batches
- The system SHALL NOT exceed network bandwidth quota

### Reliability Constraints
- The system SHALL NOT fail if a single mail server is unavailable
- The system SHALL NOT lose email jobs if service restarts

## Acceptance Criteria

- [ ] Async email sending tested (no blocking)
- [ ] Retry logic with exponential backoff implemented
- [ ] Duplicate detection working for 5-minute window
- [ ] Template rendering supports variables and conditionals
- [ ] Delivery tracking database updated
- [ ] Performance: 1,000+ emails/second
- [ ] Memory: ≤500MB under sustained load
- [ ] Code coverage: ≥85%
- [ ] Security scan: OWASP compliance verified
- [ ] Documentation: API guide + examples

---

# Related Skills
- `moai-alfred-best-practices`: TRUST 5 principles for SPEC authoring
- `moai-foundation-tags`: @TAG system and traceability
- `moai-alfred-spec-validation`: Automated SPEC validation
```

---

## Summary

**moai-alfred-spec-authoring** (Enterprise v4.0.0) provides:
- ✓ Complete YAML metadata structure (7 required + 9 optional)
- ✓ EARS requirement syntax with 5 patterns
- ✓ Unwanted Behaviors definition and enforcement
- ✓ Version lifecycle management (draft → active → deprecated → archived)
- ✓ TAG integration for SPEC→TEST→CODE→DOC traceability
- ✓ Pre-submission validation checklist (25+ items)
- ✓ Real-world SPEC examples with annotations
- ✓ Common pitfalls and anti-patterns guide
- ✓ Review checklist for approvers
- ✓ TRUST 5 principles integration

**Use when**: Creating new SPEC documents, reviewing requirements, clarifying acceptance criteria, or establishing SPEC best practices.
