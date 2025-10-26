# Python Version Matrix — Latest Stable Versions (2025-10-22)

**Reference Guide for moai-lang-python Skill Creation & Updates**

---

## Core Python Versions

| Version | Status | Released | EOL | Recommended | Use Case |
|---------|--------|----------|-----|-------------|----------|
| **3.13.1** | Latest | 2024-10 | 2029-10 | ✅ New projects | Modern features, best performance |
| **3.12.7** | LTS | 2023-10 | 2028-10 | ✅ Production | Enterprise stability, wide library support |
| **3.11.10** | Maintenance | 2022-10 | 2027-10 | ⚠️ Legacy | Gradual migration from 3.10 |
| **3.10.x** | EOL | 2021-10 | 2026-10 | ❌ Deprecated | No new code; migrate to 3.12+ |
| **3.9.x** | EOL | 2020-10 | 2025-10 | ❌ Deprecated | CRITICAL: Stop using |

**Recommended Selection for New Skills**:
- **Default**: Python 3.13.1 (unless compatibility required)
- **LTS Alternative**: Python 3.12.7 (if enterprise stability preferred)
- **Never use**: 3.10 and below (EOL, security risks)

---

## Python 3.13 New Features

### PEP 695 — Type Parameter Syntax

```python
# OLD (3.11)
from typing import TypeVar, Generic

T = TypeVar('T')

class Stack(Generic[T]):
    def push(self, item: T) -> None: ...

# NEW (3.13)
class Stack[T]:
    def push(self, item: T) -> None: ...
```

**Skill Impact**: Update type hints examples to use new syntax

### PEP 701 — Syntactic Formalization of f-strings

```python
# OLD (3.11)
name = "world"
print(f"Hello, {name}!")

# NEW (3.13) — Nested f-strings, multiline expressions
user = {"name": "Alice", "age": 30}
print(f"User: {user['name']}, Age: {user['age']}")

# Nested f-strings now work!
print(f"Formatted: {f'{x:>10}'}")
```

**Skill Impact**: Add f-string best practices section

### PEP 698 — Override Decorator for Static Typing

```python
from typing import override

class Parent:
    def method(self) -> None: ...

class Child(Parent):
    @override
    def method(self) -> None: ...  # Mypy will check this is actually overriding
```

**Skill Impact**: Add override decorator to inheritance examples

### PEP 709 — Improved Error Messages for Syntax Errors

- Better error locations and suggestions
- Clear hints for common mistakes

**Skill Impact**: Document new error messages in debugging section

---

## Python Ecosystem Tools (Latest Stable)

### Testing Frameworks

| Tool | Latest | LTS | Recommended | Notes |
|------|--------|-----|-------------|-------|
| **pytest** | 8.4.2 | 8.4.x | ✅ | Industry standard, 99% of Python projects |
| **unittest** | 3.13 | stdlib | ✅ | Built-in, minimal learning curve |
| **tox** | 4.15.1 | 4.x | ⚠️ | Test across multiple environments |
| **nose2** | 0.9.3 | 0.9.x | ❌ | Legacy, use pytest instead |

**Skill Recommendation**: Default to pytest 8.4.2 for all examples

### Code Quality Tools

| Tool | Latest | Purpose | Recommended |
|------|--------|---------|-------------|
| **ruff** | 0.13.1 | Linter + formatter (REPLACES black + pylint) | ✅ New standard |
| **black** | 24.10.0 | Formatter only | ⚠️ Use ruff instead |
| **pylint** | 3.2.6 | Linter only | ⚠️ Use ruff instead |
| **flake8** | 7.1.2 | Linter | ⚠️ Slower than ruff |
| **isort** | 5.13.2 | Import sorter | ⚠️ Integrated into ruff |

**Skill Recommendation**: Lead with ruff 0.13.1, mention legacy tools as "deprecated pattern"

### Type Checking

| Tool | Latest | Purpose | Recommended |
|------|--------|---------|-------------|
| **mypy** | 1.8.0 | Static type checker | ✅ Industry standard |
| **pyright** | 1.1.380 | VSCode native type checker | ✅ Faster alternative |
| **pydantic** | 2.7.0 | Runtime validation | ✅ Essential for APIs |

**Skill Recommendation**: Include mypy 1.8.0 + Pydantic 2.7.0 examples

### Package Management

| Tool | Latest | Purpose | Recommended |
|------|--------|---------|-------------|
| **uv** | 0.9.3 | Fast Python package manager | ✅ NEW STANDARD (2025) |
| **pip** | 24.3.1 | Built-in package manager | ⚠️ 10x slower than uv |
| **poetry** | 1.8.3 | Dependency management | ⚠️ Still used in legacy projects |
| **pipenv** | 2025.1.13 | Virtual env + deps | ❌ Deprecated, use uv |

**Skill Recommendation**: Feature uv 0.9.3 as primary, mention pip for compatibility

### Virtual Environment

| Tool | Latest | Purpose | Recommended |
|------|--------|---------|-------------|
| **venv** | 3.13 | Built-in (stdlib) | ✅ For isolated environments |
| **virtualenv** | 20.26.3 | Enhanced venv | ⚠️ Use venv for modern Python |
| **uv venv** | 0.9.3 | uv's venv command | ✅ Preferred with uv |

**Skill Recommendation**: Show venv setup, mention uv venv as faster alternative

---

## Web Frameworks (Python Ecosystem)

| Framework | Latest | Type | Recommended | Maturity |
|-----------|--------|------|-------------|----------|
| **FastAPI** | 0.115.0 | Async/REST | ✅ New projects | Stable (4+ years) |
| **Django** | 5.1.3 | Full-stack | ✅ Large projects | Very stable (19+ years) |
| **Flask** | 3.1.3 | Lightweight | ✅ Microservices | Stable (15+ years) |
| **Starlette** | 0.40.1 | Async base | ✅ Foundation for FastAPI | Mature |
| **Quart** | 0.20.0 | Async Flask | ⚠️ Use FastAPI instead | Less popular |

**Skill Impact**: Create web framework comparison table with 3.13 async features

---

## Key Tools Version Matrix (Build Reference)

```yaml
# .python-version (pyproject.toml)
python = "^3.12"

# tools/dev-dependencies
pytest = "^8.4.2"
mypy = "^1.8.0"
ruff = "^0.13.1"
black = "^24.10.0"  # deprecated, use ruff
black = "^24.10.0"  # use for formatter fallback
uv = "^0.9.3"

# tools/runtime
pydantic = "^2.7.0"
fastapi = "^0.115.0"
sqlalchemy = "^2.0.28"
alembic = "^1.13.4"

# optional/monitoring
opentelemetry-api = "^1.24.0"
prometheus-client = "^0.21.0"
```

**Migration Path** (if upgrading from 3.12):

```bash
# Old workflow (3.12 + black + pylint)
python -m pip install black pylint pytest

# New workflow (3.13 + ruff + pytest + uv)
uv venv
source .venv/bin/activate
uv add --dev ruff pytest mypy uv
```

---

## Async/Await Patterns (Python 3.13)

### asyncio Core

```python
# Python 3.13: TaskGroup (cleaner than gather)
import asyncio

async def main():
    async with asyncio.TaskGroup() as tg:
        task1 = tg.create_task(coroutine1())
        task2 = tg.create_task(coroutine2())
        # Tasks run concurrently, exceptions propagate
```

### Context Variables (3.13 improvements)

```python
from contextvars import ContextVar

request_id = ContextVar('request_id', default=None)

async def handle_request(id):
    token = request_id.set(id)
    # All spawned tasks inherit this context var
```

**Skill Impact**: Add async patterns section with TaskGroup examples

---

## Security Best Practices (Python 3.13)

### Hashlib Security

Python 3.13 removes deprecated hash algorithms. Update to:

```python
import hashlib

# ✅ SECURE (Python 3.13+)
hash_obj = hashlib.sha256(b"data")

# ❌ INSECURE (removed in 3.13)
hash_obj = hashlib.md5(b"data")  # Will raise ValueError
```

**Skill Impact**: Add security warnings section

### Secrets Module

```python
import secrets

# Generate secure random tokens
token = secrets.token_urlsafe(32)  # For API keys
nonce = secrets.token_bytes(16)    # For crypto
```

**Skill Impact**: Include security patterns section

---

## Migration Guide: 3.12 → 3.13

### Breaking Changes

| Change | 3.12 | 3.13 | Action |
|--------|------|------|--------|
| `black` tool | Standard | Deprecated | Migrate to `ruff` |
| Old f-strings | Works | Still works | Migrate to PEP 701 (optional) |
| `@override` | N/A | New | Add to inheritance hierarchies |
| Type hints | `from typing import` | `class Foo[T]:` | Update to PEP 695 (optional) |
| `md5` hash | Supported | ValueError | Use `sha256` |

### Automatic Migration Tools

```bash
# Run 2to3 for syntax updates (if needed)
2to3 -w /path/to/code

# Then use ruff format to apply new standards
ruff format .
```

**Skill Impact**: Create "Upgrading to Python 3.13" section

---

## SKILL.md Content Recommendations

### Sections to Create/Update

1. **What is Python 3.13?**
   - New features (PEP 695, 701, 698, 709)
   - When to upgrade
   - Migration path from 3.12

2. **Core Testing Pattern** (pytest 8.4.2)
   - Basic test structure
   - Fixtures (asyncio-aware)
   - Mocking strategies

3. **Code Quality Tools** (ruff 0.13.1)
   - Linting (replaces black + pylint)
   - Formatting
   - Integration with editors (VSCode, PyCharm)

4. **Type Checking** (mypy 1.8.0 + Pydantic 2.7.0)
   - Static type hints (PEP 695 syntax)
   - Runtime validation with Pydantic
   - Common pitfalls

5. **Async Patterns** (asyncio + TaskGroup)
   - TaskGroup for concurrent tasks
   - Context variables in async code
   - Testing async code with pytest

6. **Security Best Practices**
   - Secrets module for token generation
   - Secure hashing (sha256)
   - Avoiding deprecated patterns

7. **Package Management** (uv 0.9.3)
   - Creating virtual environments
   - Installing dependencies
   - Publishing to PyPI

8. **Examples**
   - RESTful API with FastAPI 0.115.0
   - Database integration (SQLAlchemy 2.0.28)
   - Testing with pytest 8.4.2 + mocking

---

## Reference.md CLI Commands

```bash
# Python version management
python --version                    # Check current version
python3.13 --version               # Check specific version
pyenv versions                      # List installed versions (if using pyenv)

# Virtual environment
python -m venv .venv                # Create venv
source .venv/bin/activate           # Activate (macOS/Linux)
.venv\Scripts\activate              # Activate (Windows)

# Package management (uv)
uv venv                             # Create with uv
uv add pytest ruff mypy             # Add dependencies
uv sync                             # Install from lock file
uv publish                          # Publish to PyPI

# Code quality (ruff)
ruff check .                        # Lint code
ruff format .                       # Format code
ruff check --select E501 .         # Check specific rule (line length)

# Type checking
mypy .                              # Run type checker
mypy --strict .                     # Strict mode

# Testing
pytest                              # Run all tests
pytest -v                           # Verbose output
pytest -s                           # Show print statements
pytest tests/test_file.py -k "pattern"  # Run specific tests

# Performance profiling
python -m cProfile script.py        # Profile script
python -m timeit "code"             # Time code snippet
```

---

## Skill Update Checklist for moai-lang-python v2.0

- [ ] Update SKILL.md frontmatter: version 2.0, created 2025-10-22
- [ ] Add Python 3.13 features section (PEP 695, 701, 698)
- [ ] Replace all tool examples with current versions (ruff, uv, pytest 8.4.2)
- [ ] Create reference.md with CLI command matrix
- [ ] Create examples.md with 5+ code samples (async, FastAPI, testing)
- [ ] Add migration guide (3.12 → 3.13)
- [ ] Add security best practices
- [ ] Update all version numbers in examples
- [ ] Verify all links point to current docs
- [ ] Test all code examples against Python 3.13.1
- [ ] Deploy to both locations (project root + templates)

---

**Version**: 0.1.0
**Created**: 2025-10-22
**Framework**: MoAI-ADK + Python 3.13.1 + Ecosystem Tools
**Next Review**: 2025-11-22 (1 month for fast-moving ecosystem)
