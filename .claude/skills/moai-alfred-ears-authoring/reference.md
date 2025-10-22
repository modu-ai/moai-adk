# EARS Reference Documentation

_Last updated: 2025-10-22_

## EARS (Easy Approach to Requirements Syntax)

### Overview
EARS is a structured approach to writing clear, testable requirements developed by Rolls-Royce PLC in 2009.

### Five Requirement Patterns

#### 1. Ubiquitous Requirements
**Always active, no trigger**

**Syntax**: The <system> shall <response>

**Use when**: Requirement applies continuously

**Examples**:
- The system shall encrypt all data at rest.
- The API shall support JSON and XML formats.

#### 2. Event-Driven Requirements
**Triggered by specific event**

**Syntax**: When <trigger>, the <system> shall <response>

**Use when**: System reacts to discrete events

**Examples**:
- When login fails 3 times, the system shall lock the account.
- When order is placed, the system shall send confirmation email.

#### 3. State-Driven Requirements
**Active while condition persists**

**Syntax**: While <state>, the <system> shall <response>

**Use when**: Behavior depends on ongoing state

**Examples**:
- While offline, the app shall cache all changes.
- While loading, the system shall display progress indicator.

#### 4. Optional Feature Requirements
**Conditional on feature presence**

**Syntax**: Where <feature>, the <system> shall <response>

**Use when**: Feature is optional or configurable

**Examples**:
- Where premium tier is enabled, the system shall remove ads.
- Where GPS is available, the app shall show location-based content.

#### 5. Complex Requirements
**Combination of patterns**

**Syntax**: While <state>, when <trigger>, the <system> shall <response>

**Use when**: Multiple conditions apply

**Examples**:
- While authenticated, when session expires, the system shall redirect to login.
- While processing, when error occurs, the system shall rollback transaction.

---

## EARS Keywords

| Keyword | Pattern | Meaning |
|---------|---------|---------|
| (none) | Ubiquitous | Always applies |
| When | Event-Driven | Triggered by event |
| While | State-Driven | Active during state |
| Where | Optional Feature | Conditional on feature |
| (combined) | Complex | Multiple conditions |

---

## Writing EARS Requirements

### Best Practices

✅ **DO**:
- Use clear, specific language
- One requirement per statement
- Make requirements testable
- Use consistent terminology
- Start with the pattern keyword

❌ **DON'T**:
- Use vague terms ("appropriate", "reasonable")
- Combine multiple requirements
- Use "and/or" constructions
- Mix different patterns inappropriately

### Examples

**Poor (vague)**:
```
The system should handle errors appropriately.
```

**Better (EARS)**:
```
When an API call fails, the system shall retry up to 3 times with exponential backoff.
```

---

## MoAI-ADK SPEC Integration

### SPEC File Structure
```markdown
---
id: AUTH-001
version: 0.1.0
status: draft
created: 2025-10-22
---

# Authentication Service SPEC

## Requirements (EARS Format)

### Ubiquitous
- The system shall hash all passwords using bcrypt with cost factor 12.

### Event-Driven
- When login succeeds, the system shall create a JWT token valid for 24 hours.
- When login fails, the system shall log the attempt with IP address and timestamp.

### State-Driven
- While user session is active, the system shall refresh the token every 15 minutes.

### Optional
- Where 2FA is enabled, the system shall require TOTP code after password verification.
```

---

## Validation Checklist

- [ ] Every requirement starts with EARS keyword (or none for ubiquitous)
- [ ] System name is clearly identified
- [ ] Response is specific and measurable
- [ ] No ambiguous terms ("appropriate", "as needed")
- [ ] Each requirement is independently testable
- [ ] Complex requirements use correct keyword combination

---

## References

- [EARS Official Guide - Alistair Mavin](https://alistairmavin.com/ears/)
- [EARS Tutorial (IEEE)](https://ieeexplore.ieee.org/document/5328509)
- [Jama Software - EARS Notation](https://www.jamasoftware.com/requirements-management-guide/writing-requirements/adopting-the-ears-notation-to-improve-requirements-engineering/)

---

_For practical examples, see examples.md_
