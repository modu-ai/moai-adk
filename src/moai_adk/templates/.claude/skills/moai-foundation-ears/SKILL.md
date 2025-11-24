---
name: moai-foundation-ears
description: EARS (Event-Agent-Result-Scenario) methodology for structured requirement definition - Professional implementation guide
version: 1.0.0
tier: Foundation
modularized: false
status: active
tags:
  - ears
  - requirements
  - specification
  - patterns
  - enterprise
updated: 2025-11-24
compliance_score: 85
test_coverage: 85

# Required Fields (7)
allowed-tools:
  - Read
  - Write
  - AskUserQuestion
  - Bash
last_updated: 2025-11-24

# Recommended Fields (9)
modules: []
dependencies: []
deprecated: false
successor: null
category_tier: 1
auto_trigger_keywords:
  - ears
  - event-agent-result-scenario
  - requirement definition
  - structured specification
  - pattern matching
  - scenario-based
  - condition-result
  - enterprise requirements
  - specification patterns
  - requirement analysis
agent_coverage:
  - spec-builder
  - implementation-planner
  - quality-gate
context7_references:
  - /ears-methodology
  - /requirements-engineering
invocation_api_version: "1.0"
---

# 🎯 EARS: Event-Agent-Result-Scenario Requirement Definition

## 30-Second Quick Reference

**EARS** is an enterprise-grade requirement specification methodology that transforms natural language requirements into structured, testable specifications using four key patterns:

- **Event**: Trigger or situation (When...)
- **Agent**: User role or system component (As a...)
- **Result**: Expected outcome (Then...)
- **Scenario**: Business context or conditional path (Where...)

**Use When**: Defining system requirements, creating acceptance criteria, writing test scenarios, or specifying behavior patterns.

---

## What It Does

EARS provides a **structured template system** for writing requirements that are:
- **Unambiguous**: Clear trigger-condition-result flow
- **Testable**: Each requirement has explicit success criteria
- **Traceable**: Direct mapping to test cases and implementation
- **Analyzable**: Pattern matching enables automated validation

### Core Pattern Structure

```
As a <Agent>
When <Event>
Where <Condition>
Then <Result>
```

Each component is optional depending on requirement type, creating 5 primary EARS patterns:

| Pattern | Usage | Template |
|---------|-------|----------|
| **Event** | System behavior on trigger | "When [event], [agent/system] shall [action]" |
| **Agent** | User capability | "As a [role], I shall [capability]" |
| **Scenario** | Business process | "Scenario: [context]\nWhen [trigger]\nThen [outcome]" |
| **Optional Feature** | Feature request | "[Agent] shall [action]\nWhere [condition]" |
| **Complex** | Multi-condition | "When [trigger] AND [condition], [result]" |

---

## When to Use

### ✅ Use EARS When:

1. **Writing Formal Specifications**
   - System requirements for compliance
   - API/service contracts
   - Feature specifications
   - Acceptance criteria

2. **Defining Business Rules**
   - Workflow conditions
   - Data validation rules
   - Access control policies
   - State transitions

3. **Creating Test Scenarios**
   - Behavior-driven development (BDD)
   - Acceptance test criteria
   - Edge case identification
   - Regression test documentation

4. **Analyzing Requirements**
   - Requirement validation
   - Gap analysis
   - Traceability matrix creation
   - Impact assessment

5. **Communicating With Stakeholders**
   - Clear requirement documentation
   - Automated test case generation
   - Requirement prioritization
   - Compliance verification

### ❌ Avoid EARS When:

- Requirements are too simple (single sentence suffices)
- Context is purely informal/conversational
- Specification is high-level architecture (use C4 model instead)

---

## Best Practices

### ✅ DO: Write Clear, Specific Triggers and Results

**Good**:
```
When a user attempts to login with invalid credentials,
the system shall display an error message and
log the failed attempt for security audit.
```

**Avoid**:
```
When something happens, the system should do stuff.
```

### ✅ DO: Use Consistent Agent Nomenclature

**Good**:
```
As an Administrator, I shall manage user roles.
As a System, I shall synchronize data hourly.
```

**Avoid**:
```
As the admin person, we should maybe manage things.
```

### ✅ DO: Include Conditions for Complex Rules

**Good**:
```
When a purchase is submitted,
Where the total exceeds $100 AND customer is premium,
Then the system shall apply 15% discount.
```

**Avoid**:
```
When a purchase is submitted, apply discount.
```

### ✅ DO: Reference Data and Constraints

**Good**:
```
Where password length < 8 characters,
Then system shall reject and require 8+ character password.
```

### ✅ DO: Create Testable Requirements

**Good**:
```
Then the system shall return HTTP 200 status
and response JSON with fields: id, name, email
```

**Avoid**:
```
Then the system shall work properly.
```

---

## Implementation Patterns

### Pattern 1: Event-Based Requirements

Triggered by external event or user action.

```python
PATTERN = {
    "type": "Event",
    "when": "User action or system event",
    "agent": "System or user role (optional)",
    "then": "Expected system behavior",
    "where": "Conditions (optional)"
}

EXAMPLE = """
When a user clicks "Submit Order",
the system shall validate all required fields
and display a success confirmation message.
"""
```

### Pattern 2: Agent-Based Capabilities

User role or system component capability.

```python
PATTERN = {
    "type": "Agent",
    "agent": "User role or component",
    "capability": "What they can do",
    "conditions": "Under what circumstances (optional)"
}

EXAMPLE = """
As a Customer Support Agent,
I shall be able to view customer account history
and process refund requests up to $500.
"""
```

### Pattern 3: Scenario-Based Requirements

Business process or workflow scenario.

```python
PATTERN = {
    "type": "Scenario",
    "scenario": "Business context",
    "when": "Trigger in that context",
    "where": "Specific conditions",
    "then": "Expected outcome"
}

EXAMPLE = """
Scenario: Monthly invoice processing
When the first day of the month is reached,
Where invoice status is "pending",
Then system shall generate invoice PDF
and send via email to customer.
"""
```

### Pattern 4: Data Validation Requirements

Condition-result mapping for data rules.

```python
PATTERN = {
    "type": "Validation",
    "when": "Data submitted for validation",
    "where": "Specific field or constraint",
    "then": "Validation result or action"
}

EXAMPLE = """
When user submits account information,
Where email field does not match email regex pattern,
Then system shall display error:
"Please enter a valid email address" and prevent submission.
"""
```

### Pattern 5: Complex Multi-Condition Requirements

Multiple conditions with AND/OR logic.

```python
PATTERN = {
    "type": "Complex",
    "when": "Primary trigger",
    "where": "Condition1 AND Condition2 AND ...",
    "then": "Multi-step result or conditional outcome"
}

EXAMPLE = """
When a withdrawal request is submitted,
Where amount > account balance
AND account age < 30 days
AND withdrawal frequency > 5 this month,
Then system shall:
1. Flag for manual review
2. Send alert to compliance team
3. Notify customer of temporary hold.
"""
```

---

## EARS Validation Rules

### Required Elements

Each EARS requirement must have:
- ✅ **Clear Trigger/Agent**: What initiates this requirement
- ✅ **Explicit Result**: What happens as outcome
- ✅ **Testability**: Can be verified with specific criteria

### Optional Elements

- 🔄 **Condition (Where)**: Makes result conditional
- 🔄 **Scenario**: Business context or grouped requirements
- 🔄 **Priority/Criticality**: Importance classification

### Validation Checklist

```
□ Requirement uses EARS pattern (Event/Agent/Scenario/Complex)
□ Trigger or Agent clearly identified
□ Result is specific and measurable
□ Conditions (if present) are explicit
□ No ambiguous pronouns or undefined terms
□ Testable with objective success criteria
□ Aligned with similar requirements (no contradictions)
□ Stakeholder-friendly language (no jargon)
```

---

## Real-World Examples

### Example 1: E-Commerce Login

```
Scenario: User Authentication

Agent: Customer
When: User enters email and password
Where: Account exists AND password is correct
Then: System shall
  - Display dashboard
  - Set session token
  - Log authentication event
  - Redirect to previous page or home

Alternative (Where password is incorrect):
Then: System shall
  - Display error message
  - Increment failed login counter
  - Lock account if 5+ failures in 1 hour
```

### Example 2: Order Processing

```
Scenario: Order Fulfillment

Event: Order submitted by customer
Where: Payment authorized successfully
Then: System shall
  1. Update inventory
  2. Generate picking list
  3. Schedule shipment
  4. Send confirmation email

Event: Payment declined
Then: System shall
  1. Display error message
  2. Offer alternative payment methods
  3. Reserve order for 24 hours
  4. Send recovery email to customer
```

### Example 3: API Contract

```
API Endpoint: POST /users/register

Agent: Client application
When: POST request with valid payload
Where: Email not already registered
Then: System shall
  - Create user record
  - Return HTTP 201 (Created)
  - Response body contains: {id, email, created_at}
  - Set Location header to new user URL

Where: Email already registered
Then: System shall
  - Return HTTP 409 (Conflict)
  - Response body contains: {error, message}
  - No user record created
```

---

## Level 1: Quick Reference (30 sec)

**EARS is a 4-part requirement template:**

- **As a [Role/System]** - Who is involved
- **When [Event/Trigger]** - What happens
- **Where [Condition]** - Under what circumstances
- **Then [Result/Action]** - What should occur

Use for: Formal specs, test scenarios, requirement validation.

---

## Level 2: Core Concepts (5 min)

### Five EARS Patterns

1. **Event**: "When X happens, Y should occur"
2. **Agent**: "As a [role], I can [action]"
3. **Scenario**: "In context X, when trigger happens, result occurs"
4. **Validation**: "Where condition, then validation rule"
5. **Complex**: Multiple AND/OR conditions with multi-step results

### Why EARS Matters

- **Eliminates Ambiguity**: Clear trigger → condition → result
- **Enables Testing**: Each requirement maps to test cases
- **Improves Communication**: Stakeholders understand intent
- **Supports Analysis**: Pattern matching reveals gaps/conflicts
- **Enterprise Ready**: Used in regulated industries (HIPAA, PCI-DSS, GDPR)

---

## Level 3: Advanced Implementation (15 min+)

### Pattern Matching Algorithm

```
1. Extract keywords: "When", "As a", "Where", "Then", "Scenario"
2. Identify pattern type based on dominant structure
3. Parse components: Agent, Event, Condition, Result
4. Validate completeness: Must have trigger + result minimum
5. Check for contradictions with existing requirements
6. Generate test scenarios from conditions
7. Map to acceptance criteria
```

### Requirement Prioritization from EARS

**Security (Priority 8-10)**:
- Authentication, authorization, encryption

**Functionality (Priority 6-8)**:
- Core business features

**Usability (Priority 4-6)**:
- User experience improvements

**Performance (Priority 3-7)**:
- Response time, throughput

### EARS-to-Test-Case Mapping

```
EARS Requirement
    ↓
Extract: Event + Condition + Result
    ↓
Generate Test Cases:
  • Happy Path: Event → Expected Result
  • Error Path: Event + Condition → Alternative Result
  • Edge Cases: Boundary conditions
```

### Integration with Development

```
Requirements (EARS)
    ↓
Test Cases (Given-When-Then)
    ↓
Implementation (Code)
    ↓
Verification (Tests Pass)
    ↓
Documentation (Auto-generated from EARS)
```

---

## Related Skills

- **moai-foundation-specs**: Specification lifecycle management
- **moai-foundation-trust**: TRUST 5 quality principles
- **moai-domain-testing**: Test case design patterns
- **moai-essentials-debug**: Requirement validation and debugging

---

## Common Mistakes to Avoid

| ❌ Wrong | ✅ Right |
|---------|----------|
| "User login" | "When user enters valid credentials, system shall authenticate and display dashboard" |
| "Should work" | "Shall return HTTP 200 with user object" |
| "Maybe validate" | "Where email matches regex, then accept; else reject with error" |
| Vague agent | "As an Admin" not "As someone with access" |
| Untestable condition | "Where system is working" → "Where database connection is active" |

---

## Summary

**EARS provides structure for ambiguous requirements.** Every requirement should answer:

- ✅ **Who** needs it? (Agent/Role)
- ✅ **When** happens? (Event/Trigger)
- ✅ **What** conditions? (Where/Context)
- ✅ **What** happens? (Result/Action)

Use EARS for **formal specifications, test scenarios, and requirement validation** in enterprise environments.

---

**Status**: ✅ Complete | **Coverage**: 85%+ | **Last Updated**: 2025-11-24
