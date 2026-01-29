# EARS Format Reference

Purpose: Comprehensive guide to Easy Approach to Requirements Syntax (EARS) patterns for SPEC generation.

Version: 1.0.0
Last Updated: 2026-01-06
Parent: [spec-first-ddd.md](spec-first-ddd.md)

---

## EARS Pattern Overview

EARS provides five requirement patterns for clear, unambiguous specifications:

1. **Ubiquitous** - Always active requirements
2. **Event-driven** - Triggered by specific events
3. **State-driven** - Active during specific states
4. **Unwanted** - Negative requirements (prevention)
5. **Optional** - Nice-to-have features

---

## Pattern 1: Ubiquitous Requirements

Syntax: `The system SHALL [action].`

Example:
```markdown
### SPEC-001-REQ-01: User Registration (Ubiquitous)
Pattern: Ubiquitous
Statement: The system SHALL register users with email and password validation.

Acceptance Criteria:
- Email format validated (RFC 5322)
- Password strength: ≥8 chars, mixed case, numbers, symbols
- Duplicate email rejected with clear error
- Success returns user ID and confirmation email sent

Test Coverage Target: ≥90%
```

---

## Pattern 2: Event-Driven Requirements

Syntax: `WHEN [event], the system SHALL [action].`

Example:
```markdown
### SPEC-001-REQ-02: JWT Token Generation (Event-driven)
Pattern: Event-driven
Statement: WHEN a user successfully authenticates, the system SHALL generate a JWT token with 1-hour expiry.

Acceptance Criteria:
- Token includes user ID, email, role claims
- Token signed with RS256 algorithm
- Expiry set to 1 hour from generation
- Refresh token generated with 7-day expiry

Test Coverage Target: ≥95%
```

---

## Pattern 3: State-Driven Requirements

Syntax: `WHILE [state], the system SHALL [action].`

Example:
```markdown
### SPEC-001-REQ-03: Token Validation (State-driven)
Pattern: State-driven
Statement: WHILE a request includes Authorization header, the system SHALL validate JWT token before processing.

Acceptance Criteria:
- Expired tokens rejected with 401 Unauthorized
- Invalid signature rejected with 401 Unauthorized
- Valid token extracts user claims successfully
- Token blacklist checked (revoked tokens)

Test Coverage Target: ≥95%
```

---

## Pattern 4: Unwanted Requirements

Syntax: `The system SHALL NOT [action].`

Example:
```markdown
### SPEC-001-REQ-04: Weak Password Prevention (Unwanted)
Pattern: Unwanted
Statement: The system SHALL NOT allow passwords from common password lists (top 10K).

Acceptance Criteria:
- Common passwords rejected (e.g., "password123")
- Sequential patterns rejected (e.g., "abc123")
- User-specific patterns rejected (e.g., email prefix)
- Clear error message with improvement suggestions

Test Coverage Target: ≥85%
```

---

## Pattern 5: Optional Requirements

Syntax: `WHERE [condition], the system SHOULD [action].`

Example:
```markdown
### SPEC-001-REQ-05: OAuth2 Integration (Optional)
Pattern: Optional
Statement: WHERE user chooses, the system SHOULD support OAuth2 authentication via Google and GitHub.

Acceptance Criteria:
- OAuth2 providers configurable
- User can link multiple providers to one account
- Provider-specific profile data merged
- Graceful fallback if provider unavailable

Test Coverage Target: ≥80%
```

---

## Complex Requirement Example

Multi-pattern requirements for complex scenarios:

```markdown
### SPEC-002-REQ-03: Multi-Factor Authentication (Event-driven + State-driven)
Pattern: Event-driven + State-driven
Statement:
- WHEN a user attempts login with MFA enabled (Event)
- WHILE the MFA verification is pending (State)
- The system SHALL send TOTP code and require verification within 5 minutes

Acceptance Criteria:
1. Event trigger: Login attempt detected
2. State check: User has MFA enabled
3. Action: Generate TOTP code (6 digits, 30s validity)
4. Notification: Send code via SMS or email
5. Verification: User submits code within 5 minutes
6. Expiry: Code expires after 5 minutes
7. Rate limiting: Max 3 failed attempts, then 15-minute lockout

Test Scenarios:
- Happy path: User submits valid code within time
- Expired code: User submits code after 5 minutes
- Invalid code: User submits incorrect code
- Rate limit: User exceeds 3 failed attempts
- Disabled MFA: User without MFA enabled
```

---

## Complexity Analysis Template

```yaml
# .moai/specs/SPEC-001/complexity.yaml
complexity_metrics:
  total_requirements: 5
  critical_requirements: 3

  complexity_breakdown:
    SPEC-001-REQ-01: Medium   # Standard CRUD + validation
    SPEC-001-REQ-02: Medium   # JWT library integration
    SPEC-001-REQ-03: High     # Security validation logic
    SPEC-001-REQ-04: Low      # Lookup validation
    SPEC-001-REQ-05: High     # External API integration

  estimated_effort:
    development: 8 hours
    testing: 4 hours
    total: 12 hours

  risk_factors:
    - Security-critical functionality
    - External OAuth2 provider dependencies
    - Token expiry edge cases

  dependencies:
    - PyJWT library
    - bcrypt library
    - OAuth2 client libraries
```

---

## Works Well With

- [spec-first-ddd.md](spec-first-ddd.md) - Main workflow overview
- [spec-ddd-implementation.md](spec-ddd-implementation.md) - DDD patterns

---

Version: 1.0.0
Last Updated: 2026-01-06
