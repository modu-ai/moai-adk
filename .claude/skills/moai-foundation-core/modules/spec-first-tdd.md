# SPEC-First TDD - Specification-Driven Development

Purpose: Specification-driven test-driven development workflow ensuring clear requirements before implementation through EARS format and RED-GREEN-REFACTOR cycles.

Version: 2.0.0 (Modular Split)
Last Updated: 2026-01-06

---

## Quick Reference (30 seconds)

SPEC-First TDD is MoAI-ADK's development methodology combining:

1. SPEC Generation - EARS format requirements (/moai:1-plan)
2. Test-Driven Development - RED-GREEN-REFACTOR (/moai:2-run)
3. Documentation Sync - Auto-generated docs (/moai:3-sync)

Three-Phase Workflow:
```
Phase 1: SPEC → spec-builder → .moai/specs/SPEC-XXX/spec.md
Phase 2: TDD  → tdd-implementer → Code + Tests (≥85% coverage)
Phase 3: Docs → docs-manager → API docs + diagrams
```

Token Budget: SPEC 30K | TDD 180K | Docs 40K | Total 250K

Key Practice: Execute `/clear` after Phase 1 to save 45-50K tokens.

EARS Patterns:
- Ubiquitous: System SHALL always...
- Event-driven: WHEN <event>, system SHALL...
- State-driven: WHILE <state>, system SHALL...
- Unwanted: System SHALL NOT...
- Optional: WHERE possible, system SHOULD...

Extended Documentation:
- [EARS Format Reference](spec-ears-format.md) - Detailed EARS patterns and examples
- [TDD Implementation](spec-tdd-implementation.md) - RED-GREEN-REFACTOR workflows

---

## Implementation Guide (5 minutes)

### Phase 1: SPEC Generation

Purpose: Define clear, testable requirements in EARS format before coding.

Workflow:
```bash
# 1. Generate SPEC
/moai:1-plan "Implement user authentication with JWT tokens"

# 2. spec-builder creates:
.moai/specs/SPEC-001/
    spec.md           # EARS format requirements
    acceptance.md     # Acceptance criteria
    complexity.yaml   # Complexity analysis

# 3. Execute /clear (mandatory)
/clear    # Saves 45-50K tokens, prepares clean context
```

EARS Format Structure:

```markdown
---
spec_id: SPEC-001
title: User Authentication System
version: 1.0.0
complexity: Medium
estimated_effort: 8-12 hours
---

## Requirements

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

### Phase 2: Test-Driven Development

RED-GREEN-REFACTOR Cycle:

```python
# RED: Write failing test first
def test_register_user():
    result = register_user("user@example.com", "SecureP@ssw0rd")
    assert result.success is True  # Fails - function doesn't exist

# GREEN: Minimal implementation
def register_user(email, password):
    return RegistrationResult(success=True, user=User())

# REFACTOR: Improve quality
def register_user(email: str, password: str) -> RegistrationResult:
    """Register new user with email and password.

    Implements SPEC-001-REQ-01
    """
    # Validation, hashing, database operations
    return RegistrationResult(success=True, user=user)
```

Coverage Validation:
```bash
# Run tests with coverage
pytest tests/auth/test_registration.py --cov=src/auth/registration --cov-report=html
```

---

### Phase 3: Documentation Synchronization

Workflow:
```bash
# 1. Generate documentation
/moai:3-sync SPEC-001

# 2. docs-manager creates:
.moai/specs/SPEC-001/
    docs/
        api.md           # API reference
        architecture.md  # Architecture diagram
        testing.md       # Test report
        report.md        # Implementation summary
```

---

## Advanced Patterns

For comprehensive implementation patterns including MFA examples, iterative SPEC refinement, and CI/CD integration, see:

- [EARS Format Reference](spec-ears-format.md) - All EARS patterns with examples
- [TDD Implementation](spec-tdd-implementation.md) - Advanced TDD workflows

---

## Works Well With

Agents:
- spec-builder - EARS format SPEC generation
- tdd-implementer - RED-GREEN-REFACTOR execution
- quality-gate - TRUST 5 validation
- docs-manager - Documentation generation

Skills:
- moai-workflow-testing - Test frameworks

Commands:
- /moai:1-plan - SPEC generation (Phase 1)
- /moai:2-run - TDD implementation (Phase 2)
- /moai:3-sync - Documentation sync (Phase 3)
- /clear - Token optimization between phases

---

Version: 2.0.0
Last Updated: 2026-01-06
Status: Production Ready
