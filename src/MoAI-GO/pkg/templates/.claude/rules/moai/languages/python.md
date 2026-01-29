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
- Formatting: black, isort
- Type checking: mypy
- Testing: pytest with coverage >= 85%
- Package management: uv or Poetry

## Preferred Patterns

- Use async/await over callbacks
- Use Pydantic v2 for validation
- Use SQLAlchemy 2.0 async patterns
- Use pytest-asyncio for async tests

## MoAI Integration

- Use Skill("moai-lang-python") for detailed patterns
- Follow TRUST 5 quality gates
- Configure ruff in pyproject.toml
