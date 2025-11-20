---
name: moai-foundation-ears
version: 4.0.0
updated: 2025-11-20
status: stable
tier: foundation
description: EARS (Easy Approach to Requirements Syntax) format for clear requirements
allowed-tools: [Read, Write]
---

# EARS Format Expert

**Easy Approach to Requirements Syntax**

> **Purpose**: Write clear, testable requirements  
> **Format**: 5 requirement types

---

## Overview

EARS provides templates for writing unambiguous requirements.

### 5 Requirement Types

1.  **Ubiquitous**: Always true
2.  **Event-Driven**: Triggered by events
3.  **State-Driven**: Depends on system state
4.  **Optional**: May occur
5.  **Complex**: Multiple conditions

---

## Templates

### 1. Ubiquitous Requirements

**When**: No conditions, always true.  
**Template**: The [system] shall [action].

**Examples**:

- The system shall encrypt passwords using bcrypt.
- The API shall return JSON responses.

### 2. Event-Driven Requirements

**When**: Triggered by specific events.  
**Template**: When [trigger], the [system] shall [action].

**Examples**:

- When user clicks "Submit", the form shall validate all fields.
- When payment succeeds, the system shall send confirmation email.

### 3. State-Driven Requirements

**When**: Depends on system state.  
**Template**: While [state], the [system] shall [action].

**Examples**:

- While user is authenticated, the dashboard shall display personalized content.
- While battery is below 20%, the app shall enable power-saving mode.

### 4. Optional Requirements

**When**: Feature may or may not be used.  
**Template**: Where [condition], the [system] shall [action].

**Examples**:

- Where dark mode is enabled, the UI shall use dark theme colors.
- Where GPS is available, the app shall show user location.

### 5. Complex Requirements

**When**: Multiple conditions or alternatives.  
**Template**: If [condition], then [action], else [alternative].

**Examples**:

- If login succeeds, then redirect to dashboard, else display error message.
- If file size > 10MB, then compress before upload, else upload directly.

---

## Acceptance Criteria

Each requirement should include measurable acceptance criteria:

**Example**:

```
Requirement: When user submits login form, the system shall validate credentials.

Acceptance Criteria:
✓ Email format verified (RFC 5322)
✓ Password length ≥8 characters
✓ Validation completes in <100ms
✓ Error messages are user-friendly
```

---

## Anti-Patterns

### ❌ Avoid Ambiguity

**Bad**: The system should work fast.  
**Good**: The system shall respond to API requests in <500ms (p95).

### ❌ Avoid Mixing Concerns

**Bad**: When user logs in, the system shall validate credentials and send welcome email.  
**Good**: Split into two requirements (login validation + email notification).

---

## Validation Checklist

- [ ] **Template**: Correct EARS template used?
- [ ] **Testable**: Can be verified with tests?
- [ ] **Unambiguous**: No vague terms (e.g., "fast", "user-friendly")?
- [ ] **Measurable**: Quantifiable criteria provided?

---

## Related Skills

- `moai-foundation-specs`: SPEC writing
- `moai-foundation-tags`: Traceability

---

**Last Updated**: 2025-11-20
