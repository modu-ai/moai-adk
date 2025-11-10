# TAG System Complete Reference

The TAG system, core of MoAI-ADK's traceability system.

## Purpose

Ensures **complete traceability** by connecting SPEC, TEST, CODE, and DOC all together following CODE-FIRST principles.

```
SPEC-001 (Requirements)
    ↓
@TEST:SPEC-001 (Tests)
    ↓
@CODE:SPEC-001 (Implementation)
    ↓
@DOC:SPEC-001 (Documentation)
    ↓
Cross-reference (Complete traceability)
```

## TAG Types

| TAG         | Location         | Purpose        | Example              |
| ----------- | ---------------- | -------------- | -------------------- |
| **SPEC-ID** | .moai/specs/     | Requirements   | SPEC-001             |
| **@TEST**   | tests/           | Test code      | @TEST:SPEC-001:\*    |
| **@CODE**   | src/             | Implementation code | @CODE:SPEC-001:\* |
| **@DOC**    | docs/            | Documentation  | @DOC:SPEC-001:\*     |

## TAG Writing Rules

### SPEC TAG

```
SPEC-001: First spec
SPEC-002: Second spec
SPEC-N: Nth spec
```

### @TEST TAG

```python
# @TEST:SPEC-001:login_success
def test_login_success():
    pass

# @TEST:SPEC-001:login_failure
def test_login_failure():
    pass
```

### @CODE TAG

```python
# @CODE:SPEC-001:register_user
def register_user(email, password):
    pass

# @CODE:SPEC-001:validate_email
def validate_email(email):
    pass
```

### @DOC TAG

```markdown
# API Documentation @DOC:SPEC-001:api

This is the API documentation for SPEC-001.
```

## TAG Validation Rules

| Rule       | Description                              | On Violation |
| ---------- | ---------------------------------------- | ------------ |
| **Uniqueness** | Same TAG must not be duplicated         | Error        |
| **Completeness** | SPEC→TEST→CODE→DOC all must exist       | Warning      |
| **Consistency** | TAG format consistency                   | Error        |
| **Traceability** | Cross-reference possible                 | Warning      |

## TAG Scanning and Validation

```bash
# Query TAG status
moai-adk status

# Detailed query for specific SPEC TAG
moai-adk status --spec SPEC-001

# Execute TAG validation
/alfred:3-sync auto SPEC-001

# Remove duplicate TAGs
/alfred:tag-dedup --dry-run
/alfred:tag-dedup --apply --backup
```

## <span class="material-icons">library_books</span> Detailed Guides

- **[TAG Types](types.md)** - Detailed explanation of each TAG type
- **[Traceability System](traceability.md)** - TAG chain and completeness validation

______________________________________________________________________

**Next**: [TAG Types](types.md) or [Traceability System](traceability.md)
