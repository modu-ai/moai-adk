---
name: moai-foundation-ears
description: EARS (Easy Approach to Requirements Syntax) methodology for clear, testable, and unambiguous specifications
version: 1.0.0
modularized: false
last_updated: 2025-11-24
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
  - Read
  - Write
  - Edit
compliance_score: 85
modules: []
dependencies: []
deprecated: false
successor: null
category_tier: 1
auto_trigger_keywords:
  - EARS
  - requirements
  - specification
  - testable
  - unambiguous
  - complete
  - scenario
  - conditions
agent_coverage:
  - spec-builder
  - implementation-planner
  - tdd-implementer
context7_references: []
invocation_api_version: "1.0"
---

## Quick Reference (30 seconds)

**EARS** (Easy Approach to Requirements Syntax) is a lightweight, structured methodology for writing clear, testable requirements. It focuses on five core patterns: **Event**, **Agent**, **Result**, **Scenario**, and **Benefit**. EARS helps teams write requirements that are unambiguous, complete, and directly testable.

**Core Value**: Eliminates ambiguous requirements that lead to misinterpretation, scope creep, and failed implementations.

---

## Implementation Guide (5-10 minutes)

### Core EARS Patterns

EARS defines five fundamental requirement patterns:

#### 1. **Basic** Pattern
```
The [system name] shall [action]
```
Example: "The authentication system shall validate user credentials against the database."

#### 2. **Scenario** Pattern
```
GIVEN [precondition]
WHEN [event/trigger]
THEN [system response]
```
Example:
```
GIVEN user has entered correct credentials
WHEN user clicks Login button
THEN system displays dashboard
```

#### 3. **Unwanted Behavior** Pattern
```
The [system name] shall NOT [unwanted action] WHEN [condition]
```
Example: "The API shall NOT accept requests WHEN authentication token is expired."

#### 4. **Conditional** Pattern
```
WHERE [condition]
The [system name] SHALL [action]
```
Example:
```
WHERE user role is 'admin'
The system SHALL display user management panel
```

#### 5. **Complex** Pattern
```
IF [condition]
THEN [action]
ELSE [alternative action]
WHEN [trigger]
```

### Guidelines for Writing Clear Requirements

1. **Use Active Voice**: "System calculates total" not "Total is calculated"
2. **Be Specific**: Avoid "soon", "quickly", "easily" - use measurable metrics
3. **Single Assertion**: One requirement per statement
4. **Testability**: Every requirement must be verifiable
5. **Complete Context**: Include preconditions, triggers, and post-conditions

### Metadata for Each Requirement

Each requirement should include:
- **ID**: Unique identifier (REQ-001, REQ-002)
- **Status**: Draft, Active, Approved, Deprecated
- **Priority**: Critical, High, Medium, Low
- **Effort**: Story points or time estimate
- **Dependencies**: Other requirements it depends on
- **Success Criteria**: How to verify it's met
- **Acceptance Tests**: Specific test cases

---

## Advanced Patterns (10+ minutes)

### Handling Complex Requirements

**Orchestration Pattern**: Requirements that depend on sequencing
```
GIVEN system is in state A
WHEN event X occurs
AND event Y occurs
AND event Z occurs
THEN system transitions to state B
AND triggers action C
```

**Concurrency Pattern**: Parallel requirements
```
The system SHALL:
  - Process request A within 100ms
  - Process request B within 100ms
  - Handle 1000 concurrent requests
```

**Uncertainty Pattern**: When exact behavior is still being defined
```
The system SHALL support [feature] in one of these ways:
  Option A: [approach 1]
  Option B: [approach 2]
Decision pending: Product team to decide Q4 2025
```

### Writing Testable Requirements

Good requirements pass the "Fit Criterion" test:

**Vague**: "The system shall be fast"
**Better**: "The system shall respond to search queries within 500ms"
**Best**: "The system shall respond to search queries within 500ms for 95% of requests measured over 1 hour"

### Organizing Requirements Hierarchically

```
Epic-001: User Authentication System
  Feature-001: Login
    REQ-001: Accept username/password
    REQ-002: Validate credentials
    REQ-003: Create session
    REQ-004: Redirect to dashboard
  Feature-002: Logout
    REQ-005: Destroy session
    REQ-006: Clear cookies
```

---

## Best Practices

### ✅ DO

1. **Use Consistent Structure**: Apply same pattern throughout document
2. **Define Acronyms**: First mention should define abbreviations
3. **Include Context**: Business reason WHY each requirement exists
4. **Get Stakeholder Sign-off**: Ensure requirements are approved before implementation
5. **Make Requirements Atomic**: One requirement = one testable behavior
6. **Use Clear Language**: Avoid jargon unless stakeholders understand it
7. **Track Changes**: Maintain version history of requirements

### ❌ DON'T

1. **Mix Levels of Abstraction**: Don't mix system requirements with implementation details
2. **Write Implementation Details**: EARS captures WHAT not HOW
3. **Use Weak Modal Verbs**: Avoid "should", "may", "might" - use "shall", "must"
4. **Create Ambiguous Conditions**: IF conditions must be testable and clear
5. **Ignore Non-Functional Requirements**: Include performance, security, scalability
6. **Skip Edge Cases**: Document unwanted behavior patterns explicitly
7. **Write Vague Success Criteria**: Must be objectively measurable

---

## Common Pitfalls and Solutions

| Issue | Problem | Solution |
|-------|---------|----------|
| **Ambiguous Conditions** | "When user is ready" | "When user clicks Submit button" |
| **Vague Actions** | "System handles the request" | "System validates input, processes payment, and sends confirmation" |
| **Missing Context** | Requirement doesn't explain why | Add "Benefit:" or "Business Value:" section |
| **Weak Modal Verbs** | "Should support..." | Use "shall support" or "must support" |
| **Mixed Abstraction** | Technical detail in business requirement | Separate into business req + technical design doc |

---

## Real-World Example: E-Commerce Checkout

```
Feature: Complete Purchase

Scenario-1: Basic Purchase
GIVEN user has items in shopping cart
AND user is logged in
WHEN user proceeds to checkout
AND user reviews order summary
AND user confirms payment method
AND user clicks "Complete Purchase"
THEN system processes payment
AND system creates order record
AND system sends confirmation email
AND user sees order confirmation page

Scenario-2: Invalid Payment
GIVEN user has items in cart
WHEN user submits payment with invalid card
THEN system displays error message
AND order is NOT created
AND user remains on payment page

Scenario-3: Inventory Check
WHERE product inventory < requested quantity
The system SHALL show "Limited Availability" message
AND allow user to adjust quantity OR proceed with available stock
```

---

## Integration with TDD

EARS requirements map directly to test cases:

```
Requirement: "WHEN user login fails, system SHALL display error message"
Test Cases:
  1. test_login_fails_with_invalid_password()
  2. test_login_fails_with_wrong_username()
  3. test_error_message_displayed()
  4. test_form_remains_on_screen()
```

---

## Context7 Integration

For latest requirement methodology patterns and industry best practices, refer to:
- Requirement engineering guides and SMART criteria
- Behavior-driven development (BDD) frameworks
- Acceptance test-driven development (ATDD) patterns

---

## Workflow Integration

1. **Product Definition**: Write epics in user story format
2. **Requirements Gathering**: Translate user stories to EARS format
3. **Specification**: Document detailed requirements using EARS patterns
4. **Implementation**: Map requirements to feature branches
5. **Testing**: Create tests from requirement success criteria
6. **Validation**: Verify implementation against requirements
7. **Maintenance**: Update requirements as features evolve

---

**Status**: Production Ready
**Version**: 1.0.0
**Last Updated**: 2025-11-24
