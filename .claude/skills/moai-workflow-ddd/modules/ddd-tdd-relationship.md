# DDD and TDD Relationship in MoAI-ADK

Version: 1.0.0
Last Updated: 2026-01-21

---

## Overview

This document clarifies the relationship between Domain-Driven Development (DDD) and Test-Driven Development (TDD) within the MoAI-ADK framework, explaining when to use each approach and how they complement each other.

---

## Clear Distinction: DDD vs TDD

### DDD (Domain-Driven Development)

**Full Name**: Domain-Driven Development
**Core Cycle**: ANALYZE-PRESERVE-IMPROVE
**Primary Purpose**: Behavior-preserving refactoring of existing code
**Starting Point**: Existing code with defined behavior

DDD in MoAI-ADK focuses on:

- **ANALYZE**: Understanding domain boundaries, coupling metrics, and AST structure
- **PRESERVE**: Creating characterization tests to capture current behavior
- **IMPROVE**: Applying structural changes while maintaining identical behavior

Key Characteristics:

- Works with **legacy code** that already exists
- Preserves **existing behavior** without changes
- Uses **characterization tests** that document what code actually does
- Goal is **structural improvement**, not feature addition

### TDD (Test-Driven Development)

**Full Name**: Test-Driven Development
**Core Cycle**: RED-GREEN-REFACTOR
**Primary Purpose**: Creating new functionality through tests
**Starting Point**: No code exists (blank slate)

TDD focuses on:

- **RED**: Write failing test that defines expected behavior
- **GREEN**: Write minimal code to make test pass
- **REFACTOR**: Improve structure while tests stay green

Key Characteristics:

- Works with **greenfield projects** or new features
- Creates **new behavior** through specification
- Uses **specification tests** that define what code should do
- Goal is **feature creation** with test coverage

---

## DDD as a Superset of TDD

DDD encompasses both refactoring scenarios and greenfield development, making it a superset of TDD principles.

### For Legacy Code (DDD Refactoring Mode)

When code already exists with defined behavior:

```
ANALYZE: Understand current structure and domain boundaries
    ↓
PRESERVE: Create characterization tests (capture actual behavior)
    ↓
IMPROVE: Apply structural changes with continuous validation
```

**Characterization Tests**: Document what the code currently does, not what it should do. This creates a safety net for refactoring.

### For Greenfield Projects (DDD New Feature Mode)

When creating new functionality from scratch:

```
ANALYZE: Understand requirements and specifications
    ↓
PRESERVE: Define intended behavior through specification tests (test-first like TDD)
    ↓
IMPROVE: Implement code to satisfy the defined tests
```

**Specification Tests**: Define what the code should do, following TDD's test-first approach.

### Visual Relationship

```
                    DDD (Domain-Driven Development)
                    ┌─────────────────────────────────┐
                    │                                 │
    ┌───────────────┴───────────────┐ ┌─────────────┴─────────────┐
    │                               │ │                           │
    │   Legacy Code Refactoring     │ │   Greenfield Development  │
    │   (ANALYZE-PRESERVE-IMPROVE)  │ │   (includes TDD cycle)     │
    │                               │ │                           │
    │   Characterization Tests      │ │   RED-GREEN-REFACTOR      │
    │   (Document actual behavior)  │ │   (Specification tests)    │
    │                               │ │                           │
    └───────────────────────────────┘ └───────────────────────────┘

                    TDD is contained within DDD
                    for greenfield scenarios
```

---

## When to Use Each Approach

### Use DDD for Refactoring Existing Code

Apply standard DDD cycle when:

- Code already exists with defined behavior
- Goal is improving structure without changing functionality
- Working with legacy code or technical debt
- API contracts must remain identical
- Existing tests should pass unchanged

**Command**: `/moai:2-run SPEC-XXX` with manager-ddd
**Tests**: Characterization tests (capture actual behavior)

Example Scenarios:

- Reducing coupling between modules
- Extracting classes or methods
- Renaming for clarity
- Modernizing code patterns
- Database migration without behavior change

### Use DDD's Greenfield Adaptation for New Features

Apply DDD with test-first approach when:

- Creating new functionality from scratch
- No existing code to preserve
- Behavior specification drives development
- Starting new projects or modules

**Command**: `/moai:2-run SPEC-XXX` with manager-ddd
**Tests**: Specification tests (define expected behavior)

Example Scenarios:

- New API endpoints
- New feature modules
- Greenfield projects
- New business logic

### Never Use Pure TDD Without DDD Context

In MoAI-ADK, always use DDD methodology:

- **For existing code**: DDD's characterization test approach
- **For new features**: DDD's greenfield adaptation (which includes TDD principles)

**Why DDD instead of pure TDD?**

- DDD provides explicit structure through ANALYZE-PRESERVE-IMPROVE
- DDD handles both legacy and greenfield scenarios
- DDD integrates with MoAI-ADK's SPEC system
- DDD includes behavior preservation guarantees
- DDD supports MoAI-ADK's quality gates (TRUST 5)

---

## MoAI-ADK Workflow Integration

### Command Layer

**Primary Command**: `/moai:2-run SPEC-XXX`

This command invokes manager-ddd to execute DDD methodology for both refactoring and new feature development.

### Agent Layer

**Primary Agent**: manager-ddd

The manager-ddd agent:
- Loads moai-workflow-ddd skill automatically
- Executes ANALYZE-PRESERVE-IMPROVE cycle
- Creates appropriate tests (characterization or specification)
- Validates behavior preservation or implementation
- Integrates with moai-workflow-testing for test patterns

### Skill Layer

**Primary Skills**:

1. **moai-workflow-ddd**: DDD methodology and cycle execution
2. **moai-workflow-testing**: Testing patterns and validation
3. **moai-tool-ast-grep**: Structural analysis and transformation
4. **moai-foundation-quality**: TRUST 5 quality gates

### Workflow Integration

```
/moai:1-plan "description"
         ↓
    manager-spec
         ↓
    SPEC Document (EARS format)
         ↓
/moai:2-run SPEC-XXX
         ↓
    manager-ddd
    ├── moai-workflow-ddd (DDD methodology)
    ├── moai-workflow-testing (testing patterns)
    ├── moai-tool-ast-grep (structural analysis)
    └── moai-foundation-quality (quality gates)
         ↓
    Implementation
    ├── Legacy code: ANALYZE-PRESERVE-IMPROVE
    └── New features: ANALYZE-PRESERVE-IMPROVE (test-first)
         ↓
/moai:3-sync SPEC-XXX
         ↓
    Documentation
```

---

## Test Types Comparison

### Characterization Tests (DDD Legacy Mode)

**Purpose**: Document what code currently does
**Creation**: After code exists, during PRESERVE phase
**Basis**: Actual observed behavior
**Change**: Refactoring must keep tests passing

Example:
```python
def test_characterize_user_authentication():
    """Characterization test: Documents actual behavior"""
    result = authenticate_user("existing_user", "password")
    # Test captures what actually happens, even if surprising
    assert result["status"] == "success"
    assert result["token"] is not None
```

### Specification Tests (DDD Greenfield Mode / TDD)

**Purpose**: Define what code should do
**Creation**: Before code exists (test-first)
**Basis**: Requirements and specifications
**Change**: Tests drive implementation

Example:
```python
def test_specify_user_authentication_success():
    """Specification test: Defines expected behavior"""
    result = authenticate_user("valid_user", "valid_password")
    # Test defines what should happen
    assert result["status"] == "success"
    assert result["token"] is not None
    assert len(result["token"]) == 32
```

---

## Decision Flowchart

```
Does the code I'm changing already exist?
         │
         ├─ YES → Use DDD Refactoring Mode
         │         - ANALYZE structure
         │         - PRESERVE with characterization tests
         │         - IMPROVE incrementally
         │
         └─ NO  → Use DDD Greenfield Mode (includes TDD)
                   - ANALYZE requirements
                   - PRESERVE with specification tests (test-first)
                   - IMPROVE to satisfy tests
```

---

## Key Takeaways

1. **DDD = Domain-Driven Development**: MoAI-ADK's methodology using ANALYZE-PRESERVE-IMPROVE cycle
2. **TDD = Test-Driven Development**: Traditional RED-GREEN-REFACTOR cycle for new features
3. **DDD is a superset**: Includes TDD's test-first approach for greenfield scenarios
4. **Legacy code**: Use DDD with characterization tests
5. **New features**: Use DDD's greenfield adaptation (test-first like TDD)
6. **Always use DDD in MoAI-ADK**: Never use pure TDD without DDD context
7. **Command**: `/moai:2-run SPEC-XXX` always uses manager-ddd for DDD methodology

---

## Related Documentation

- **moai-workflow-ddd/SKILL.md**: Core DDD methodology and cycle details
- **manager-ddd agent**: DDD implementation specialist
- **moai-workflow-testing**: Testing patterns within DDD context
- **moai-foundation-quality**: TRUST 5 quality validation framework

---

## Quick Reference Table

| Aspect | DDD (Legacy) | DDD (Greenfield/TDD) |
|--------|--------------|----------------------|
| **Cycle** | ANALYZE-PRESERVE-IMPROVE | RED-GREEN-REFACTOR (within DDD) |
| **Starting Point** | Existing code | No code exists |
| **Tests** | Characterization (actual) | Specification (expected) |
| **Goal** | Structural improvement | Feature creation |
| **Behavior Change** | None (preserve) | Create new behavior |
| **When to Use** | Refactoring | New features |

---

**Status**: Active
**Maintained By**: MoAI-ADK Team
**Related Skills**: moai-workflow-ddd, moai-workflow-testing, moai-foundation-quality
