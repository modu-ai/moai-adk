# TAG System Reference

> CODE-FIRST traceability with @TAG markers

_Last updated: 2025-10-22_

---

## TAG Structure

### TAG Format

```
@<TYPE>:<DOMAIN>-<###> | SPEC: <spec-file> | TEST: <test-file>
```

**TAG Types**:
- `@SPEC`: Specification documents (`.moai/specs/`)
- `@CODE`: Implementation code (`src/`, `lib/`)
- `@TEST`: Test files (`tests/`, `__tests__/`)
- `@DOC`: Documentation (`docs/`, `README.md`)

**Example**:
```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/test_jwt.py

def validate_jwt_token(token: str) -> bool:
    """Validates JWT token with HS256 algorithm."""
    # Implementation
    pass
```

---

## TAG Lifecycle

### 1. SPEC Creation (TAG Birth)

```markdown
<!-- .moai/specs/SPEC-AUTH-001/spec.md -->
---
id: SPEC-AUTH-001
---

# @SPEC:AUTH-001 | JWT Token Validation

## Requirements
...
```

### 2. Test Creation (RED Phase)

```python
# tests/auth/test_jwt.py
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

def test_jwt_validation():
    """EARS: When token is valid, system shall authenticate user."""
    assert validate_jwt_token(VALID_TOKEN) == True
```

### 3. Implementation (GREEN Phase)

```python
# src/auth/jwt.py
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/test_jwt.py

def validate_jwt_token(token: str) -> bool:
    try:
        jwt.decode(token, SECRET, algorithms=["HS256"])
        return True
    except jwt.InvalidTokenError:
        return False
```

---

## TAG Validation Commands

### Find All TAGs

```bash
# Scan entire codebase
rg '@(SPEC|CODE|TEST|DOC):' -n .moai/ src/ tests/ docs/

# Output:
# .moai/specs/SPEC-AUTH-001/spec.md:5:# @SPEC:AUTH-001
# src/auth/jwt.py:1:# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/test_jwt.py
# tests/auth/test_jwt.py:2:# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md
```

### Detect Orphan TAGs

```bash
# Find CODE TAGs without corresponding SPEC
for code_tag in $(rg '@CODE:([A-Z]+-[0-9]+)' src/ -o -r '$1' | sort -u); do
  if ! rg "@SPEC:$code_tag" .moai/specs/ -q; then
    echo "❌ Orphan CODE TAG: $code_tag (no SPEC found)"
  fi
done

# Find SPEC TAGs without implementation
for spec_tag in $(rg '@SPEC:([A-Z]+-[0-9]+)' .moai/specs/ -o -r '$1' | sort -u); do
  if ! rg "@CODE:$spec_tag" src/ -q; then
    echo "⚠️  SPEC without implementation: $spec_tag"
  fi
done
```

### TAG Chain Verification

```bash
#!/bin/bash
# Verify complete TAG chain (SPEC ↔ TEST ↔ CODE)

verify_tag_chain() {
  local tag_id="$1"

  echo "Verifying TAG chain for: $tag_id"

  # Check SPEC exists
  if rg "@SPEC:$tag_id" .moai/specs/ -q; then
    echo "✅ SPEC found"
  else
    echo "❌ SPEC missing"
    return 1
  fi

  # Check TEST exists
  if rg "@TEST:$tag_id" tests/ -q; then
    echo "✅ TEST found"
  else
    echo "❌ TEST missing"
    return 1
  fi

  # Check CODE exists
  if rg "@CODE:$tag_id" src/ -q; then
    echo "✅ CODE found"
  else
    echo "❌ CODE missing"
    return 1
  fi

  echo "✅ Complete TAG chain verified"
  return 0
}

# Usage
verify_tag_chain "AUTH-001"
```

---

## TAG Inventory Report

### Generate TAG Matrix

```bash
#!/bin/bash
# Generate TAG inventory matrix

echo "TAG ID | SPEC | TEST | CODE | Status"
echo "-------|------|------|------|-------"

for tag_id in $(rg '@(SPEC|TEST|CODE):([A-Z]+-[0-9]+)' -o -r '$2' .moai/ src/ tests/ | sort -u); do
  has_spec=$(rg "@SPEC:$tag_id" .moai/specs/ -q && echo "✓" || echo "✗")
  has_test=$(rg "@TEST:$tag_id" tests/ -q && echo "✓" || echo "✗")
  has_code=$(rg "@CODE:$tag_id" src/ -q && echo "✓" || echo "✗")

  if [[ "$has_spec" == "✓" && "$has_test" == "✓" && "$has_code" == "✓" ]]; then
    status="🟢 Complete"
  elif [[ "$has_spec" == "✓" ]]; then
    status="🟡 In Progress"
  else
    status="🔴 Orphan"
  fi

  echo "$tag_id | $has_spec | $has_test | $has_code | $status"
done
```

**Example Output**:
```
TAG ID     | SPEC | TEST | CODE | Status
-----------|------|------|------|------------
AUTH-001   | ✓    | ✓    | ✓    | 🟢 Complete
AUTH-002   | ✓    | ✓    | ✗    | 🟡 In Progress
PAYMENT-003| ✗    | ✗    | ✓    | 🔴 Orphan
```

---

## CODE-FIRST Principle

### Rule: Source Code is Truth

**Priority Order**:
1. **Code TAGs** — reality (what exists)
2. **Test TAGs** — verification (what's tested)
3. **SPEC TAGs** — intent (what's planned)

**Synchronization Flow**:
```
Code Changes → Update Tests → Update SPEC → Regenerate Docs
```

**Anti-Pattern** (DON'T):
```
Update SPEC → Hope code matches
```

**Correct Pattern** (DO):
```
Update Code → Verify Tests Pass → Sync SPEC → Run /alfred:3-sync
```

---

## TAG Best Practices

### DO:
- ✅ Add TAG immediately when creating file
- ✅ Include cross-references (SPEC → TEST → CODE)
- ✅ Run TAG validation before commits
- ✅ Update HISTORY when TAG content changes
- ✅ Use domain prefixes consistently (AUTH, PAYMENT, USER)

### DON'T:
- ❌ Reuse TAG IDs for different features
- ❌ Skip TAG validation checks
- ❌ Create CODE without corresponding TEST
- ❌ Leave orphan TAGs in codebase
- ❌ Use inconsistent TAG formats

---

## Integration with /alfred:3-sync

The `doc-syncer` sub-agent automatically:
1. Scans all TAGs in codebase
2. Detects orphans and missing links
3. Generates TAG inventory report
4. Updates Living Documentation with TAG references
5. Flags quality gate violations (orphan TAGs)

**Trigger**: Runs automatically during `/alfred:3-sync`

---

## Resources

**Related Skills**:
- `moai-foundation-specs` — SPEC metadata standards
- `moai-foundation-trust` — TRUST traceability principle
- `moai-alfred-tag-scanning` — Automated TAG inventory

**Tools**:
- `ripgrep (rg)` — Fast TAG scanning
- `jq` — JSON processing for TAG reports

---

**Last Updated**: 2025-10-22
**Maintained by**: MoAI-ADK Foundation Team
