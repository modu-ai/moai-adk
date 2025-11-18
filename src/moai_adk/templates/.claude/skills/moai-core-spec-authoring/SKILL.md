---
name: moai-core-spec-authoring
version: 4.0.0
created: 2025-10-23
updated: '2025-11-18'
status: stable
tier: Alfred
description: Complete SPEC document authoring guide with YAML metadata structure (7
  required + 9 optional fields), EARS requirement syntax (5 patterns including Unwanted
  Behaviors), version lifecycle management, TAG integration, pre-submission validation
  checklist, and real-world SPEC examples.
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
stability: stable
---


# SPEC Authoring Skill (Enterprise )

## Level 1: Quick Reference

### Core Capabilities
- **YAML Metadata Structure**: 7 required + 9 optional fields for complete SPEC management
- **EARS Requirement Syntax**: 5 patterns (Universal, Conditional, Unwanted Behavior, Stakeholder, Boundary)
- **Version Lifecycle**: draft → active → deprecated → archived with state management
- **TAG Integration**: SPEC→TEST→CODE→DOC traceability
- **Validation Checklist**: 25+ pre-submission validation criteria

### Quick SPEC Template

```yaml
---
code: SPEC-001
title: Add User Authentication with JWT
status: stable
created_at: 2025-11-13
updated_at: 2025-11-13
priority: high
effort: 8
version: "4.0.0"
epic: AUTH-01
domains:
  - backend
  - security
  - database
---

# SPEC-001: Add User Authentication with JWT

## Overview
Implement secure JWT-based user authentication with password hashing and token validation.

## Requirements

### REQ-001 (Universal)
SPEC: The authentication service SHALL validate JWT tokens using RS256 algorithm against the public key.

### REQ-002 (Conditional)
SPEC: If a JWT token has expired, then the authentication service SHALL reject the request and return HTTP 401.

### REQ-003 (Unwanted Behavior)
SPEC: The authentication service SHALL NOT accept JWT tokens signed with symmetric algorithms (HS256, HS384, HS512).

### REQ-004 (Stakeholder)
As an API consumer, I want to pass JWT tokens in the Authorization header so that my requests are authenticated.

### REQ-005 (Boundary Condition)
SPEC: The authentication service SHALL return HTTP 429 when a single IP attempts more than 10 failed authentications in 5 minutes.

## Unwanted Behaviors

### Security Constraints
- The system SHALL NOT store JWT secrets in source code
- The system SHALL NOT log JWT tokens or sensitive claims

### Performance Constraints
- The system SHALL NOT block on token validation (async pattern)
- The system SHALL NOT cache tokens indefinitely

## Acceptance Criteria
- [ ] All 5 REQ patterns implemented and tested
- [ ] Code coverage ≥85%
- [ ] Security scan passed (OWASP Top 10)
- [ ] Performance: JWT validation p99 latency ≤200ms
```

### EARS Pattern Summary

| Pattern | Syntax | Use Case |
|---------|--------|----------|
| Universal | `The [System] SHALL [Action]` | Core system behavior |
| Conditional | `If [Condition], then [System] SHALL [Action]` | Conditional logic |
| Unwanted Behavior | `The [System] SHALL NOT [Action]` | Security constraints |
| Stakeholder | `As a [Role], I want [Feature] so that [Benefit]` | User stories |
| Boundary Condition | `[System] SHALL [Action] when [Condition]` | Edge cases |

## Level 2: Core Implementation

### YAML Metadata Structure

#### 7 Required Fields

```yaml
---
# SPEC identifier (auto-generated)
code: SPEC-001

# SPEC title (descriptive, 50-80 chars)
title: Add User Authentication with JWT

# SPEC status (draft | active | deprecated | archived)
status: stable

# Creation timestamp (ISO 8601: YYYY-MM-DD)
created_at: 2025-11-13

# Last updated timestamp (ISO 8601: YYYY-MM-DD)
updated_at: 2025-11-13

# Business priority (critical | high | medium | low)
priority: high

# Estimated effort in story points (1-13 scale)
effort: 8
---
```

#### 9 Optional Fields

```yaml
# Version tracking (semantic versioning: major.minor.patch)
version: "4.0.0"

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
```

### EARS Requirement Syntax

#### Pattern 1: Universal (Always True)

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

#### Pattern 2: Conditional (If-Then)

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

#### Pattern 3: Unwanted Behavior (Negative Requirement)

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

#### Pattern 4: Stakeholder (User Role-Specific)

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

#### Pattern 5: Boundary Condition (Edge Cases)

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

## Level 3: Advanced Features

### Version Lifecycle Management

#### Lifecycle States

```
DRAFT → ACTIVE → DEPRECATED → ARCHIVED
  ↓                  ↓
Under Review    Stable, In Use
  ↓                  ↓
Changes Expected    Changes Require
Feedback Pending    Major Version Bump
```

#### State Transitions

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

### Unwanted Behaviors Section

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

### TAG Integration for Traceability

#### TAG Structure (SPEC→TEST→CODE→DOC)

```
      ↓
      ↓
      ↓
```

#### TAG Placement Rules

**SPEC Document**:
```markdown
---
# SPEC-001: Feature Name
---
```

**Test File**:
```python
def test_requirement_001():
    """Test SPEC-001 REQ-001 universal pattern."""
    pass
```

**Implementation**:
```python
def authenticate_user(token: str) -> bool:
    """Validate JWT token per SPEC-001."""
    pass
```

**Documentation**:
```markdown
## Authentication Flow
Per SPEC-001, the system SHALL validate all JWT tokens...
```

## Level 4: Reference & Integration

### Pre-Submission Validation Checklist

#### Metadata Validation
- [ ] `code` field filled (SPEC-XXX format)
- [ ] `title` is descriptive (50-80 characters)
- [ ] `status` is one of: draft | active | deprecated | archived
- [ ] `created_at` is ISO 8601 format (YYYY-MM-DD)
- [ ] `updated_at` matches actual update date
- [ ] `priority` is one of: critical | high | medium | low
- [ ] `effort` is between 1-13 (story points)

#### Requirement Syntax
- [ ] At least 3 REQ patterns used (Universal, Conditional, Unwanted)
- [ ] Each REQ follows EARS syntax strictly
- [ ] Requirements are specific and testable
- [ ] No ambiguous language ("should", "may", "might")
- [ ] All REQs are actionable (have test cases)

#### Unwanted Behaviors
- [ ] Security constraints listed (if applicable)
- [ ] Performance constraints listed (if applicable)
- [ ] Reliability constraints listed (if applicable)
- [ ] Data integrity constraints listed (if applicable)
- [ ] Each Unwanted Behavior has a test approach

#### Acceptance Criteria
- [ ] All 5 EARS patterns implemented
- [ ] All Unwanted Behaviors testable
- [ ] Code coverage target ≥85% specified
- [ ] Security scan type specified
- [ ] Performance baseline defined (if applicable)

#### Final Review
- [ ] No TODO or placeholder text
- [ ] All links are valid (internal and external)
- [ ] Formatting is consistent (markdown syntax)
- [ ] No confidential information exposed
- [ ] Ready for team review

### Common Pitfalls & Anti-Patterns

#### Anti-Pattern 1: Ambiguous Requirements

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

#### Anti-Pattern 2: Vague Acceptance Criteria

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

#### Anti-Pattern 3: Missing Unwanted Behaviors

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
```

### When to Use

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

### Related Skills

- `moai-core-best-practices` - TRUST 5 principles for SPEC authoring
- `moai-core-spec-validation` - Automated SPEC validation
- `moai-foundation-specs` - SPEC lifecycle management
- `moai-foundation-trust` - Security and compliance principles
- `moai-core-workflow` - SPEC creation workflows

### TRUST Principles Applied

**T**est First: All requirements include test cases and validation criteria
**R**eadable: Clear EARS syntax with unambiguous language
**U**nified: Consistent structure across all SPEC documents
**S**ecured: Unwanted Behaviors section with security constraints
**T**raceable: TAG integration for complete requirement traceability

---

**Enterprise   Compliance**: Progressive disclosure with comprehensive EARS syntax, validation checklists, and lifecycle management.
**Last Updated**: 2025-11-18  
**Dependencies**: YAML metadata format, EARS specification, TAG system
**See Also**: [examples.md](./examples.md) for detailed SPEC examples
