---
paths:
  - "**/*.py"
  - "**/pyproject.toml"
  - "**/requirements*.txt"
---

# Python Rules

Version: Python 3.13+

## Tooling

- Linting: ruff (700+ rules, replaces flake8+isort+pyupgrade)
- Formatting: ruff format or black
- Type checking: mypy with strict mode
- Security: bandit for vulnerability scanning
- Testing: pytest with coverage >= 85%
- Package management: uv (recommended) or Poetry

## Ruff Configuration

```toml
# pyproject.toml
[tool.ruff]
line-length = 88
select = ["E", "F", "I", "B", "ANN", "S"]  # errors, pyflakes, isort, bugbear, annotations, security

[tool.ruff.lint.per-file-ignores]
"tests/*" = ["S101"]  # allow assert in tests
```

## Best Practices (2026)

- Integrate mypy into CI/CD pipeline for type error prevention
- Use type hints for all function signatures (PEP 673/674 supported)
- Use Pydantic v2 with model_validator for cross-field validation
- Use SQLAlchemy 2.0 async patterns with create_async_engine
- Use pytest-asyncio with asyncio_mode="auto"

## Security

- Run `ruff check --select S` for security linting
- Run `bandit -r src/` for vulnerability scanning
- Validate all user inputs with Pydantic

## MoAI Integration

- Use Skill("moai-lang-python") for detailed patterns
- Follow TRUST 5 quality gates
