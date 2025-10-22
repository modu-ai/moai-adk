# Workflow: TDD Implementation with Claude Code (`/alfred:2-run`)

## Objective
Implement SPEC with RED → GREEN → REFACTOR cycle, enforced by Hooks and validated by Sub-agents.

## Claude Code Components Involved

| Component | Purpose | Reference |
|-----------|---------|-----------|
| **Hooks (PreToolUse)** | Validate @TAG, enforce formatting rules before edits | [`moai-cc-hooks`](../../../skills/moai-cc-hooks/SKILL.md) |
| **Hooks (PostToolUse)** | Auto-format code, run linters after edits | [`moai-cc-hooks`](../../../skills/moai-cc-hooks/SKILL.md) |
| **Sub-agents** | code-builder pipeline (RED/GREEN/REFACTOR phases) | [`moai-cc-agents`](../../../skills/moai-cc-agents/SKILL.md) |
| **settings.json** | Restrict edits to src/, tests/; allow test execution | [`moai-cc-settings`](../../../skills/moai-cc-settings/SKILL.md) |
| **CLAUDE.md** | Reference TRUST 5 principles, quality gates | [`moai-cc-claude-md`](../../../skills/moai-cc-claude-md/SKILL.md) |
| **Memory** | Cache test results, import paths, type hints | [`moai-cc-memory`](../../../skills/moai-cc-memory/SKILL.md) |

## Step-by-Step Flow

### Step 1: Invoke TDD Command
```bash
/alfred:2-run AUTH-002

# Loads SPEC: .moai/specs/SPEC-AUTH-002/spec.md
# Activates code-builder pipeline
```

**Behind the scenes:**
- ✅ Reads SPEC requirements from `.moai/specs/SPEC-AUTH-002/`
- ✅ Loads `moai-essentials-refactor` Skill (refactoring patterns)
- ✅ Loads `moai-foundation-trust` Skill (TRUST 5 validation)
- ✅ References CLAUDE.md for project conventions

### Step 2: 🔴 RED Phase (Failing Tests)

**code-builder's task:**
1. Create `tests/auth/test_jwt_service.py` with:
   - `@TEST:AUTH-002` tag
   - Test cases derived from SPEC
   - Clear assertions

**Example:**
```python
# tests/auth/test_jwt_service.py
# @TEST:AUTH-002 | SPEC: SPEC-AUTH-002.md

import pytest
from src.auth.service import JWTService

def test_generate_token_success():
    """WHEN valid credentials provided, SHOULD issue JWT token"""
    service = JWTService()
    token = service.generate("user@example.com", "password123")

    assert token is not None
    assert token.startswith("eyJ")  # JWT format check

def test_token_expires():
    """WHEN token expires, SHOULD return 401"""
    # Test implementation...
```

**Claude Code role:**
- ❌ Tests fail (functions don't exist yet)
- Exit code: 1 (test failure)
- Memory caches test output

### Step 3: 🟢 GREEN Phase (Minimal Implementation)

**code-builder's task:**
1. Create `src/auth/service.py` with:
   - `@CODE:AUTH-002` tag
   - Minimal implementation to pass tests
   - No optimization, just pass

**Example:**
```python
# src/auth/service.py
# @CODE:AUTH-002 | SPEC: SPEC-AUTH-002.md | TEST: tests/auth/test_jwt_service.py

import jwt
import datetime

class JWTService:
    SECRET = "dev-secret"  # TODO: Use env var

    def generate(self, username: str, password: str) -> str:
        """Generate JWT token for authenticated user"""
        # TODO: Validate credentials first
        return jwt.encode(
            {"username": username, "exp": datetime.datetime.utcnow() + datetime.timedelta(minutes=15)},
            self.SECRET
        )
```

**Claude Code role:**
- **PreToolUse Hook** validates:
  - ✅ @TAG:AUTH-002 is present
  - ✅ File location is src/auth/
  - ✅ Follows naming conventions
  - ✅ Return if validation passes

- **PostToolUse Hook** formats:
  - Auto-run `black` formatter
  - Run `ruff` linter
  - Log formatting changes

- ✅ Tests now pass
- Exit code: 0 (success)

### Step 4: ♻️ REFACTOR Phase (Quality Improvement)

**code-builder's task:**
1. Improve code quality per TRUST 5 principles:
   - ✅ **Test**: Ensure 85%+ coverage
   - ✅ **Readable**: Functions ≤50 LOC, cyclomatic complexity ≤10
   - ✅ **Unified**: Type hints, consistent error handling
   - ✅ **Secured**: Input validation, no hardcoded secrets
   - ✅ **Trackable**: @TAG chain complete

**Example refactoring:**
```python
# src/auth/service.py - REFACTORED

# @CODE:AUTH-002 | SPEC: SPEC-AUTH-002.md | TEST: tests/auth/test_jwt_service.py

import os
import jwt
import datetime
from typing import Optional
from .exceptions import AuthenticationError

class JWTService:
    SECRET = os.environ.get("JWT_SECRET", "dev-secret")
    EXPIRY_MINUTES = 15

    @staticmethod
    def generate(username: str, password: str) -> Optional[str]:
        """Generate JWT token for authenticated user.

        Args:
            username: User email or ID
            password: User password (validated externally)

        Returns:
            JWT token string or None if validation fails

        Raises:
            AuthenticationError: If credentials invalid
        """
        if not username or not password:
            raise AuthenticationError("Username and password required")

        try:
            payload = {
                "username": username,
                "exp": datetime.datetime.utcnow() + datetime.timedelta(minutes=JWTService.EXPIRY_MINUTES)
            }
            return jwt.encode(payload, JWTService.SECRET, algorithm="HS256")
        except Exception as e:
            raise AuthenticationError(f"Token generation failed: {str(e)}")
```

**Claude Code role:**
- **Hooks** validate quality improvements
- **trust-checker Sub-agent** verifies:
  - ✅ Test coverage ≥85%
  - ✅ Cyclomatic complexity ≤10
  - ✅ No hardcoded secrets
  - ✅ Proper error handling

### Step 5: Git Commit History (TDD cadence)

```bash
# 1. 🔴 RED commit
git commit -m "test(AUTH-002): add failing JWT generation test"

# 2. 🟢 GREEN commit
git commit -m "feat(AUTH-002): implement minimal JWT service"

# 3. ♻️ REFACTOR commit
git commit -m "refactor(AUTH-002): improve quality per TRUST 5, add env var support"
```

## Claude Code Best Practices During Run

### ✅ DO
- Ensure @TAG:AUTH-002 present in all edits
- Use Memory to cache type definitions and imports
- Reference CLAUDE.md for project conventions
- Run tests after each phase
- Commit after each TDD phase (RED, GREEN, REFACTOR)

### ❌ DON'T
- Skip REFACTOR phase (just leave broken code)
- Edit files outside src/ and tests/
- Ignore PreToolUse/PostToolUse Hook warnings
- Remove @TAG comments
- Commit without passing tests

## Validation Checklist

### Test (@TEST:AUTH-002)
- [ ] File exists: `tests/auth/test_jwt_service.py`
- [ ] Contains `@TEST:AUTH-002` tag
- [ ] References SPEC: `# SPEC: SPEC-AUTH-002.md`
- [ ] Tests cover all SPEC requirements
- [ ] All tests pass: `pytest tests/auth/test_jwt_service.py`

### Code (@CODE:AUTH-002)
- [ ] File exists: `src/auth/service.py`
- [ ] Contains `@CODE:AUTH-002` tag
- [ ] References TEST: `# TEST: tests/auth/test_jwt_service.py`
- [ ] Functions ≤50 LOC
- [ ] Type hints on all functions
- [ ] Error handling for edge cases
- [ ] No hardcoded secrets (uses `os.environ`)

### Quality Checks
- [ ] Test coverage ≥85%
- [ ] `pytest` passes with no errors
- [ ] `ruff check` passes (linting)
- [ ] `black --check` passes (formatting)
- [ ] Cyclomatic complexity ≤10
- [ ] No security warnings

## Troubleshooting

**Issue**: "PreToolUse Hook blocked edit: @TAG missing"
→ Add `@CODE:AUTH-002` comment at top of file

**Issue**: "Tests still failing after GREEN phase"
→ Check test expectations vs implementation; may need to iterate

**Issue**: "PostToolUse formatting changes too much"
→ Adjust formatter config (black line length, ruff rules) in pyproject.toml

**Issue**: "Coverage still below 85%"
→ Add edge case tests for error paths, validation failures

## Memory Optimization

The Memory system caches during /alfred:2-run:
- ✅ Import paths (don't re-discover each iteration)
- ✅ Type definitions (reduce lookup time)
- ✅ Test fixtures (reuse across test cases)
- ✅ Error patterns (debug similar failures faster)

**Accessed via**: @moai-cc-memory guide

## Next Steps
→ Move to `/alfred:3-sync` for document synchronization and PR creation

---

**Related Guides:**
- 📖 Project Setup: [`alfred-0-project-setup.md`](./alfred-0-project-setup.md)
- 📖 Planning: [`alfred-1-plan-flow.md`](./alfred-1-plan-flow.md)
- 📖 Synchronization: [`alfred-3-sync-flow.md`](./alfred-3-sync-flow.md)
