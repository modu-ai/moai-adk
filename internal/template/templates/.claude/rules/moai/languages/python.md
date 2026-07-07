---
paths: "**/*.py,**/pyproject.toml,**/requirements*.txt"
---

# Python Development Guide

Version: Python 3.14+

Core Capabilities:

- Linting: ruff
- Formatting: ruff format, black
- Testing: pytest with coverage >= 85%
- Package management: uv, poetry, pip with pyproject.toml
- Type checking: mypy, pyright

### Quick Patterns

File Conventions:

- Use snake_case for file names and functions
- Use PascalCase for classes
- tests/ directory for test files with test_ prefix
- src/ layout for packages
- pyproject.toml as primary config file

Recommended Practices:

- Use type hints for all function signatures and return types
- Use async/await for I/O-bound operations
- Use Pydantic v2 for data validation with model_validate
- Handle all exceptions with proper error context
- Run ruff check before commit
- Use virtual environments (venv or uv)

Anti-Patterns to Avoid:

- Use bare except clauses
- Use mutable default arguments in function signatures
- Import star (from module import *)
- Use global variables for state management
- Embed credentials or secrets in code
- Mix sync and async code without proper bridges

### Python 3.14 Features

- JIT Compiler (PEP 744): Experimental, enable with PYTHON_JIT=1
- GIL-Free Mode (PEP 703): python3.13t for true parallel threads, experimental (free-threaded build)
- Pattern Matching: match/case with guards and nested patterns
- Improved error messages and deferred evaluation of annotations (PEP 649)

### FastAPI 0.115+ Patterns

Async endpoints with Depends for dependency injection. Lifespan context manager (asynccontextmanager) for startup/shutdown. Class-based dependencies for reusable patterns (Paginator, etc.). Use Pydantic models for request/response validation.

### Django 5.2 LTS

Composite Primary Keys via CompositePrimaryKey in Meta. URL reverse with query parameters and fragments. Automatic model imports in shell. GeneratedField for computed columns.

### Pydantic v2.9

- model_validate (replaces parse_obj), model_validate_json (replaces parse_raw)
- ConfigDict: from_attributes=True, populate_by_name=True, extra="forbid", str_strip_whitespace=True
- Annotated validators with AfterValidator for reusable validation
- model_validator(mode="after") for cross-field validation

### SQLAlchemy 2.0 Async

- create_async_engine with pool_pre_ping=True
- async_sessionmaker with expire_on_commit=False
- Repository pattern: async methods for CRUD with session management
- Streaming: db.stream() with async for for large result sets

### Testing: pytest

- pytest-asyncio fixtures with @pytest_asyncio.fixture
- @pytest.mark.parametrize with ids for readable test names
- Fixture factories for flexible test data generation
- Coverage: pytest --cov=app --cov-report=html, target 85%+

### Type Hints

- Protocol with @runtime_checkable for structural typing
- ParamSpec + TypeVar for typed decorators
- Generic types with constraints

### Package Management

- pyproject.toml with Poetry or uv
- uv: curl -LsSf astral.sh/uv/install.sh | sh, uv venv, uv pip install, uv add

---

## Advanced Patterns

For comprehensive coverage including advanced async patterns, large-scale service architecture, performance profiling, and production deployment configurations, apply the Quick Patterns and framework sections above together with the Context7 Integration mappings; this guide is self-contained.

---

## Context7 Integration

- tiangolo/fastapi - FastAPI framework
- django/django - Django web framework
- pydantic/pydantic - Pydantic validation
- sqlalchemy/sqlalchemy - SQLAlchemy ORM
- pytest-dev/pytest - pytest testing
- numpy/numpy - NumPy
- pandas-dev/pandas - pandas
- pola-rs/polars - Polars DataFrame

---

- `moai-domain-backend` - Backend service architecture
- `moai-domain-database` - Database patterns and ORM optimization
- `moai-workflow-testing` - DDD and testing strategies
- `moai-foundation-quality` - TRUST 5 quality principles

---

## Troubleshooting

- Version check: python --version, python -c "import sys; print(sys.version_info)"
- Async session detached: Set expire_on_commit=False or await session.refresh()
- pytest asyncio warning: Set asyncio_mode="auto" in pyproject.toml
- Pydantic v2 migration: parse_obj -> model_validate, parse_raw -> model_validate_json

---
