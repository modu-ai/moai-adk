# TAG System Examples

_Last updated: 2025-10-22_

## Example 1: Complete TAG Chain

### SPEC File
```markdown
<!-- .moai/specs/SPEC-AUTH-001/spec.md -->
# @SPEC:AUTH-001 | JWT Token Validation

## Requirements
- The system shall validate JWT tokens with HS256
```

### Test File
```python
# tests/auth/test_jwt.py
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

def test_jwt_validation():
    assert validate_jwt_token(VALID_TOKEN) == True
```

### Implementation File
```python
# src/auth/jwt.py
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/test_jwt.py

def validate_jwt_token(token: str) -> bool:
    return jwt.decode(token, SECRET, algorithms=["HS256"])
```

---

## Example 2: Detecting Orphan TAGs

```bash
# Find CODE without SPEC
rg '@CODE:AUTH-005' src/  # Found
rg '@SPEC:AUTH-005' .moai/specs/  # Not found → Orphan!
```

---

**For complete TAG reference, see [reference.md](reference.md)**
