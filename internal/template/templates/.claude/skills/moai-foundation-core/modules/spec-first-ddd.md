# SPEC-First DDD - Specification-Driven Development

Purpose: Specification-driven domain-driven development workflow ensuring clear requirements before implementation through GEARS format (current; EARS retained as legacy reference for 6-month backward-compat) and ANALYZE-PRESERVE-IMPROVE cycles.

Version: 3.0.0 (DDD Migration)
Last Updated: 2026-01-17

---

## Quick Reference (30 seconds)

SPEC-First DDD is MoAI-ADK's development methodology combining:

1. SPEC Generation - GEARS format requirements (current; EARS as legacy reference) (/moai:1-plan)
2. Domain-Driven Development - ANALYZE-PRESERVE-IMPROVE (/moai:2-run)
3. Documentation Sync - Auto-generated docs (/moai:3-sync)

Three-Phase Workflow:
```
Phase 1: SPEC → spec-builder → .moai/specs/SPEC-XXX/spec.md
Phase 2: DDD  → ddd-implementer → Code + Tests (≥85% coverage)
Phase 3: Docs → docs-manager → API docs + diagrams
```

Token Budget: SPEC 30K | DDD 180K | Docs 40K | Total 250K

Key Practice: Execute `/clear` after Phase 1 to save 45-50K tokens.

GEARS Patterns (current notation):
- Ubiquitous: The <subject> shall <behavior>
- Event-driven: When <event>, the <subject> shall <behavior>
- State-driven: While <state>, the <subject> shall <behavior>
- Where (capability gate): Where <capability or feature flag>, the <subject> shall <behavior>
- Event-detected (replaces the deprecated conditional modality): When <undesired-condition-detected>, the <subject> shall <response>

Unified compound clause: `[Where ...][While ...][When ...] The <subject> shall <behavior>` — any subset of the three modifiers may chain. The `<subject>` is generalized (any noun: system, component, service, agent, function, artifact).

EARS Patterns (legacy reference, 6-month backward-compat — expires 2026-11-22):
- Ubiquitous: System SHALL always...
- Event-driven: WHEN <event>, system SHALL...
- State-driven: WHILE <state>, system SHALL...
- Unwanted: System SHALL NOT...
- Optional: WHERE possible, system SHOULD...

Extended Documentation:
- Canonical GEARS authoring guide: `.claude/skills/moai-workflow-spec/SKILL.md` § "GEARS Format" (current)
- [EARS Format Reference (legacy reference, deprecated — see GEARS Format guide)](spec-ears-format.md) - Detailed EARS patterns and examples for legacy SPECs
- [DDD Implementation](spec-ddd-implementation.md) - ANALYZE-PRESERVE-IMPROVE workflows

---

## Implementation Guide (5 minutes)

### Phase 1: SPEC Generation

Purpose: Define clear, testable requirements in GEARS format (current; EARS legacy reference supported for the 6-month backward-compat window) before coding.

Workflow:
```bash
# 1. Generate SPEC
/moai:1-plan "Implement user authentication with JWT tokens"

# 2. spec-builder creates:
.moai/specs/SPEC-001/
    spec.md           # GEARS format requirements (current; EARS for legacy SPECs)
    acceptance.md     # Acceptance criteria
    complexity.yaml   # Complexity analysis

# 3. Execute /clear (mandatory)
/clear    # Saves 45-50K tokens, prepares clean context
```

GEARS Format Structure (current; subject "system" shown for readability, but any noun is valid per generalized-subject rule):

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

### Phase 2: Domain-Driven Development

ANALYZE-PRESERVE-IMPROVE Cycle:

```python
# ANALYZE: Understand existing code and behavior
def analyze_existing_registration():
    """Analyze current registration implementation.

    - Identify existing behavior patterns
    - Document current test coverage
    - Map dependencies and side effects
    """
    pass

# PRESERVE: Create characterization tests
def test_register_user_existing_behavior():
    """Characterization test for existing behavior."""
    result = register_user("user@example.com", "SecureP@ssw0rd")
    assert result.success is True  # Documents existing behavior

# IMPROVE: Refactor with behavior preservation
def register_user(email: str, password: str) -> RegistrationResult:
    """Register new user with email and password.

    Implements SPEC-001-REQ-01
    Behavior preserved from existing implementation.
    """
    # Improved validation, hashing, database operations
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

- Canonical GEARS authoring guide: `.claude/skills/moai-workflow-spec/SKILL.md` § "GEARS Format" (current notation)
- [EARS Format Reference (legacy reference, deprecated — see GEARS Format guide)](spec-ears-format.md) - All EARS patterns with examples for the 88 legacy SPECs
- [DDD Implementation](spec-ddd-implementation.md) - Advanced DDD workflows

---

## Works Well With

Agents:
- spec-builder - GEARS format SPEC generation (current; EARS as legacy reference)
- ddd-implementer - ANALYZE-PRESERVE-IMPROVE execution
- quality-gate - TRUST 5 validation
- docs-manager - Documentation generation

Skills:
- moai-workflow-testing - Test frameworks

Commands:
- /moai:1-plan - SPEC generation (Phase 1)
- /moai:2-run - DDD implementation (Phase 2)
- /moai:3-sync - Documentation sync (Phase 3)
- /clear - Token optimization between phases

---

Version: 3.0.0
Last Updated: 2026-01-17
Status: Production Ready
