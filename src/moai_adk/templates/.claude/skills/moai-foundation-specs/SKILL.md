---
name: moai-foundation-specs
version: 4.0.0
updated: 2025-11-20
status: stable
tier: foundation
description: SPEC writing standards with EARS format and TAG traceability
allowed-tools: [Read, Write, WebSearch, WebFetch]
---

# SPEC Writing Expert

**Requirement Specification Standards**

> **Format**: EARS (Easy Approach to Requirements Syntax)  
> **Traceability**: TAG-based (SPEC→TEST→CODE→DOC)

---

## Overview

Standardized approach to writing clear, testable requirements.

### SPEC Components

1.  **Functional Requirements**: What the system does.
2.  **Non-Functional Requirements**: Performance, security, usability.
3.  **Acceptance Criteria**: Testable conditions for completion.
4.  **TAG Assignment**: Unique identifier for traceability.

---

## EARS Format (5 Types)

### 1. Ubiquitous Requirements

Always true, no conditions.

**Template**: The [system] shall [action].

**Example**:

```
The authentication system shall encrypt passwords using bcrypt.
```

### 2. Event-Driven Requirements

Triggered by events.

**Template**: When [trigger], the [system] shall [action].

**Example**:

```
When a user submits login credentials, the system shall validate against the database.
```

### 3. State-Driven Requirements

Depends on system state.

**Template**: While [state], the [system] shall [action].

**Example**:

```
While user is authenticated, the system shall display personalized dashboard.
```

### 4. Optional Requirements

May or may not occur.

**Template**: Where [condition], the [system] shall [action].

**Example**:

```
Where MFA is enabled, the system shall require second factor authentication.
```

### 5. Complex Requirements

Multiple conditions.

**Template**: If [condition], then [action], else [alternative].

**Example**:

```
If password validation fails, then system shall display error message, else proceed to dashboard.
```

---

## SPEC Template

```markdown
# SPEC-001: User Authentication

## Overview

Implement secure user authentication with JWT tokens.

## Functional Requirements

### @AUTH-001: Login Validation

**Type**: Event-Driven
**Requirement**: When user submits credentials, system shall validate email and password format before database lookup.
**Acceptance Criteria**:

- Email format: RFC 5322 compliant
- Password: min 8 characters
- Return validation errors within 100ms

### @AUTH-002: Password Hashing

**Type**: Ubiquitous
**Requirement**: System shall hash passwords using bcrypt with 12 rounds.
**Acceptance Criteria**:

- Bcrypt algorithm (OWASP recommended)
- 12 salt rounds (2025 standard)
- Unique salt per password

## Non-Functional Requirements

### @AUTH-NF-001: Performance

- Login response time: <500ms (p95)
- Concurrent logins: 1000/sec

### @AUTH-NF-002: Security

- OWASP Top 10 compliance
- TLS 1.3 for transmission
- Rate limiting: 5 attempts/min per IP

## TAG Traceability

- SPEC: @AUTH-001, @AUTH-002
- Tests: test_auth.py
- Code: auth/login.py
- Docs: api/authentication.md
```

---

## Best Practices

1.  **One Requirement Per TAG**: Don't combine multiple requirements.
2.  **Testable**: Each requirement must be verifiable.
3.  **Unambiguous**: Use "shall" (mandatory), "should" (recommended), "may" (optional).
4.  **Measurable**: Include quantifiable acceptance criteria.

---

## Validation Checklist

- [ ] **EARS Format**: Correct template used?
- [ ] **TAG Assignment**: Unique @TAG-### assigned?
- [ ] **Testable**: Can be verified with tests?
- [ ] **Complete**: Acceptance criteria defined?
- [ ] **Traceable**: Linked to TEST/CODE/DOC?

---

## Related Skills

- `moai-foundation-ears`: Detailed EARS syntax
- `moai-foundation-tags`: TAG management
- `moai-foundation-trust`: Quality principles

---

**Last Updated**: 2025-11-20
