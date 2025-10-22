---

name: moai-lang-python
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: Python 3.13+ best practices with pytest 8.4.2, mypy 1.8.0, ruff 0.13.1, and uv 0.9.3. Use when writing or reviewing Python code in project workflows.
keywords: [python, testing, pytest, mypy, ruff, uv, async, fastapi, pydantic]
allowed-tools:
  - Read
  - Bash
---

# Python 3.13 Expert Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-python |
| **Version** | 2.0.0 (2025-10-22) |
| **Python Support** | 3.13.1 (latest), 3.12.7 (LTS), 3.11.10 (maintenance) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when language keywords detected |
| **Trigger cues** | `.py` files, Python frameworks, TDD discussions, async patterns |
| **Tier** | Language / 23 (comprehensive coverage) |

---

## What It Does

Provides **Python 3.13+ expertise** for modern TDD development, including:

- ✅ **Testing Framework**: pytest 8.4.2 (fixtures, asyncio, parametrization)
- ✅ **Code Quality**: ruff 0.13.1 (unified linter + formatter, replaces black/pylint)
- ✅ **Type Safety**: mypy 1.8.0 + Pydantic 2.7.0 (static + runtime validation)
- ✅ **Package Management**: uv 0.9.3 (10x faster than pip)
- ✅ **Python 3.13 Features**: PEP 695 (type params), PEP 701 (f-strings), PEP 698 (@override)
- ✅ **Async/Await**: asyncio.TaskGroup, context variables, concurrent patterns
- ✅ **Security**: Secrets module, secure hashing, SQLAlchemy 2.0.28

---

## When to Use

**Automatic triggers**:
- Python code discussions, `.py` files, framework guidance
- "Writing Python tests", "How to use pytest", "Python type hints"
- Python SPEC implementation (`/alfred:2-run`)
- Async pattern requests

**Manual invocation**:
- Review Python code for TRUST 5 compliance
- Design Python microservices (FastAPI 0.115.0 recommended)
- Upgrade from Python 3.12 to 3.13
- Refactor async code to use TaskGroup

---

## How It Works (Best Practices)

### 1. TDD Framework (pytest 8.4.2)

```python
# Test discovery & fixtures
import pytest
from src.calculator import add

def test_add_positive_numbers():
    """Verify addition of positive integers."""
    assert add(2, 3) == 5

@pytest.mark.asyncio
async def test_async_operation():
    """Test async functions with pytest-asyncio."""
    result = await async_fetch_data()
    assert result is not None
```

**Key Points**:
- ✅ Use pytest 8.4.2+ (not unittest)
- ✅ One assertion per test (clarity)
- ✅ Fixtures for setup/teardown
- ✅ `pytest.mark.asyncio` for async tests
- ✅ `pytest-mock` for mocking (not mock.patch)
- ✅ Coverage ≥85% enforced by quality gate

**CLI Commands**:
```bash
pytest                              # Run all tests
pytest -v                           # Verbose output
pytest --cov=src --cov-report=term # Coverage report (≥85% required)
pytest -k "pattern"                 # Run matching tests
pytest -m asyncio                   # Run async tests only
```

### 2. Code Quality (ruff 0.13.1 — NEW STANDARD)

**⚠️ BREAKING CHANGE**: Ruff 0.13.1 replaces black + pylint + isort. Update workflows:

```yaml
# OLD (deprecated)
- black .                           # Format
- pylint src/                       # Lint
- isort .                           # Import sort

# NEW (ruff 0.13.1)
- ruff check . --fix                # Lint + fix
- ruff format .                     # Format (replaces black)
```

**Configuration** (`pyproject.toml`):
```toml
[tool.ruff]
line-length = 100
target-version = "py313"

[tool.ruff.lint]
select = ["E", "F", "W"]  # Errors, formatting, warnings
extend-select = ["I"]     # Import sorting (replaces isort)
```

**CLI Commands**:
```bash
ruff check .                        # Lint all files
ruff format .                       # Format with auto-fix
ruff check --show-fixes .           # Show what would be fixed
ruff check --select E501 .          # Check specific rule (line length)
```

### 3. Type Safety (mypy 1.8.0 + Pydantic 2.7.0)

**Static Type Checking** (mypy):

```python
from typing import override

class Parent:
    def method(self, x: int) -> str: ...

class Child(Parent):
    @override  # NEW in Python 3.13 — mypy validates this
    def method(self, x: int) -> str:
        return str(x)
```

**Runtime Validation** (Pydantic 2.7.0):

```python
from pydantic import BaseModel, Field

class User(BaseModel):
    id: int = Field(gt=0)           # Must be > 0
    name: str = Field(min_length=1)
    email: str                      # Auto-validated as email
```

**CLI Commands**:
```bash
mypy .                              # Run type checker
mypy --strict .                     # Strict mode (recommended)
mypy --show-column-numbers .        # Precise error locations
```

### 4. Package Management (uv 0.9.3)

**Why uv?**: 10x faster than pip, integrated with ruff + pytest

```bash
# Create virtual environment
uv venv                             # Create .venv/
source .venv/bin/activate           # Activate

# Install dependencies
uv add pytest ruff mypy             # Add to pyproject.toml
uv add --dev pytest-asyncio         # Add as dev dependency
uv sync                             # Install all (from lock file)

# Publish
uv publish                          # Push to PyPI
```

**pyproject.toml** (uv config):
```toml
[project]
name = "my-project"
version = "2.0.0"
requires-python = ">=3.13"

[project.optional-dependencies]
dev = ["pytest>=8.4.2", "ruff>=0.13.1", "mypy>=1.8.0"]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```

### 5. Async Patterns (Python 3.13)

**TaskGroup** (cleaner than `asyncio.gather`):

```python
import asyncio

async def main():
    async with asyncio.TaskGroup() as tg:
        task1 = tg.create_task(fetch_user(1))
        task2 = tg.create_task(fetch_posts(1))
        # Tasks run concurrently
        # Exceptions propagate automatically (no need for gather)

asyncio.run(main())
```

**Context Variables** (thread-safe in async):

```python
from contextvars import ContextVar

request_id = ContextVar('request_id', default=None)

async def handle_request(req_id):
    token = request_id.set(req_id)
    # All spawned tasks inherit this context var
    await process_async()
```

### 6. Security Best Practices

**Secrets Module** (token generation):
```python
import secrets

api_key = secrets.token_urlsafe(32)  # Safe random tokens
nonce = secrets.token_bytes(16)       # Cryptographic nonce
```

**Secure Hashing** (sha256, not md5):
```python
import hashlib

# ✅ SECURE
hash_obj = hashlib.sha256(b"password")

# ❌ INSECURE (removed in Python 3.13)
hash_obj = hashlib.md5(b"password")   # ValueError!
```

---

## Python 3.13 New Features

### PEP 695 — Type Parameter Syntax

```python
# OLD (3.12)
from typing import TypeVar, Generic
T = TypeVar('T')
class Stack(Generic[T]):
    def push(self, item: T) -> None: ...

# NEW (3.13)
class Stack[T]:
    def push(self, item: T) -> None: ...
```

### PEP 701 — Improved F-Strings

```python
# NEW: Nested f-strings, arbitrary expressions
user = {"name": "Alice", "age": 30}
print(f"User: {user['name']}, Age: {user['age']}")  # Works!

# Nested f-strings
x = 10
print(f"Result: {f'{x:>10}'}")  # Works in 3.13!
```

### PEP 698 — Override Decorator

```python
from typing import override

class Parent:
    def method(self) -> None: ...

class Child(Parent):
    @override  # mypy checks this is actually overriding
    def method(self) -> None: ...
```

---

## Example Workflow

**Setup** (uv + Python 3.13):
```bash
uv venv --python 3.13               # Create venv with Python 3.13
source .venv/bin/activate
uv add pytest ruff mypy fastapi pydantic
```

**TDD Loop** (pytest):
```bash
pytest                              # RED: Watch tests fail
# [implement code]
pytest                              # GREEN: Watch tests pass
ruff check --fix .                  # REFACTOR: Fix code quality
```

**Quality Gate** (before commit):
```bash
pytest --cov=src --cov-report=term # Coverage ≥85%?
ruff check .                        # Lint pass?
mypy --strict .                     # Type check pass?
```

---

## Tool Version Matrix (2025-10-22)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **Python** | 3.13.1 | Runtime | ✅ Latest |
| **pytest** | 8.4.2 | Testing | ✅ Current |
| **ruff** | 0.13.1 | Lint/Format | ✅ New standard |
| **mypy** | 1.8.0 | Type checking | ✅ Current |
| **uv** | 0.9.3 | Package manager | ✅ Recommended |
| **FastAPI** | 0.115.0 | Web framework | ✅ Latest |
| **Pydantic** | 2.7.0 | Validation | ✅ Latest |
| **SQLAlchemy** | 2.0.28 | ORM | ✅ Latest |

---

## Inputs

- Python source directories (e.g., `src/`, `app/`)
- Configuration files (`pyproject.toml`, `pytest.ini`)
- Test suites and sample data
- Existing CI/CD workflows

## Outputs

- Test/lint execution plan for Python 3.13
- Code review checklist (TRUST 5 principles)
- Migration guide (3.12 → 3.13)
- Performance optimization recommendations

## Failure Modes

- ❌ Python 3.13 not installed → Recommend `uv venv --python 3.13`
- ❌ Dependencies missing → Run `uv sync`
- ❌ Tests fail → Use `debug-helper` agent for triage

---

## Inputs

- Language-specific source directories (e.g. `src/`, `app/`).
- Language-specific build/test configuration files (e.g. `pyproject.toml`).
- Relevant test suites and sample data.

## Outputs

- Test/lint execution plan tailored to Python 3.13+.
- TRUST 5 review checkpoints (coverage, linting, types, security, tags).
- Migration path from older Python versions.

## Failure Modes

- When Python 3.13 runtime is not installed.
- When project dependencies are not in pyproject.toml.
- When test coverage falls below 85%.

## Dependencies

- Access to the project file is required using the Read/Bash tools.
- Integration with `moai-foundation-langs` for language detection.
- Integration with `moai-foundation-trust` for quality gate enforcement.

---

## References (Latest Documentation)

- **Python 3.13**: https://docs.python.org/3.13/ (accessed 2025-10-22)
- **pytest 8.4.2**: https://docs.pytest.org/en/stable/ (accessed 2025-10-22)
- **ruff 0.13.1**: https://docs.astral.sh/ruff/ (accessed 2025-10-22)
- **mypy 1.8.0**: https://mypy.readthedocs.io/ (accessed 2025-10-22)
- **uv 0.9.3**: https://docs.astral.sh/uv/ (accessed 2025-10-22)
- **FastAPI 0.115.0**: https://fastapi.tiangolo.com/ (accessed 2025-10-22)
- **Pydantic 2.7.0**: https://docs.pydantic.dev/ (accessed 2025-10-22)

---

## Changelog

- **v2.0.0** (2025-10-22): Major update with Python 3.13 support, ruff 0.13.1 standard, uv 0.9.3 package manager, async/await TaskGroup patterns, security best practices, PEP 695/701/698 features
- **v1.0.0** (2025-03-29): Initial Skill with pytest, mypy, ruff, black, uv coverage

---

## Works Well With

- `moai-foundation-trust` (coverage verification, quality gates)
- `moai-alfred-code-reviewer` (Python-specific code review)
- `moai-alfred-debugger-pro` (Python debugging with pdb, pudb)
- `moai-lang-sql` (SQLAlchemy integration)
- `moai-domain-backend` (FastAPI microservices)

---

## Best Practices

✅ **DO**:
- Use ruff 0.13.1 as linter + formatter (not black + pylint)
- Specify exact Python version: `requires-python = "^3.13"`
- Use pytest 8.4.2 for all tests
- Enable type checking: mypy --strict
- Run quality gate before each commit
- Use uv for package management (10x faster)
- Add docstrings to public APIs
- Use f-strings (PEP 701 supports nested expressions)

❌ **DON'T**:
- Use black + pylint (deprecated, use ruff instead)
- Use md5 hashing (removed in Python 3.13)
- Mix pytest with unittest
- Ignore coverage requirements (<85% fails)
- Use old type hint syntax (use PEP 695 `class Foo[T]:`)
- Use `asyncio.gather` without error handling (use TaskGroup instead)
