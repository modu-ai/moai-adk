---
paths:
  - "**/*.py"
  - "**/pyproject.toml"
  - "**/requirements*.txt"
---

# Python Rules

Version: Python 3.13+

## Tooling

- Linting: ruff (not flake8)
- Formatting: black, isort (or ruff format)
- Type checking: mypy or pyright
- Testing: pytest with coverage >= 85%
- Package management: uv or Poetry

## MUST

- Use type hints for all function signatures
- Use async/await for I/O-bound operations
- Validate inputs with Pydantic v2
- Configure ruff in pyproject.toml
- Use context managers for resource management
- Document public APIs with docstrings

## MUST NOT

- Use bare except clauses
- Mutate default arguments (mutable defaults)
- Use wildcard imports (from x import *)
- Ignore type checker errors with # type: ignore without reason
- Store secrets in code or config files
- Use print() for logging (use logging module)

## File Conventions

- test_*.py or *_test.py for test files
- __init__.py for package initialization
- conftest.py for pytest fixtures
- Use snake_case for modules and functions
- Use PascalCase for classes

## Testing

- Use pytest fixtures for setup/teardown
- Use pytest-asyncio for async tests
- Use parametrize for test variations
- Mock external services with pytest-mock

## MoAI Integration

- Use Skill("moai-lang-python") for detailed patterns
- Follow TRUST 5 quality gates
