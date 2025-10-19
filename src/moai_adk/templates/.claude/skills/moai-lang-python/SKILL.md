---
name: moai-lang-python
description: Python best practices with pytest, mypy, ruff, black, and uv package
  management
allowed-tools:
- Read
- Bash
---

# Python Expert

## What it does

Provides Python-specific expertise for TDD development, including pytest testing, mypy type checking, ruff linting, black formatting, and modern uv package management.

## When to use

- "Python 테스트 작성", "pytest 사용법", "Python 타입 힌트"
- Automatically invoked when working with Python projects
- Python SPEC implementation (`/alfred:2-build`)

## How it works

**TDD Framework**:
- **pytest**: Test discovery, fixtures, parametrize, markers
- **coverage.py**: Test coverage ≥85% enforcement
- **pytest-mock**: Mocking and patching

**Type Safety**:
- **mypy**: Static type checking with strict mode
- Type hints for function signatures, return types
- Generic types, Protocols, TypedDict

**Code Quality**:
- **ruff**: Fast Python linter (replaces flake8, isort, pylint)
- **black**: Opinionated code formatter
- Complexity checks (≤10), line length (≤88)

**Package Management**:
- **uv**: Modern, fast package installer
- `pyproject.toml` for project configuration
- Virtual environment management

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Meaningful variable names (no single letters except loops)
- Guard clauses over nested conditions
- Docstrings for public APIs

## Examples

### Example 1: TDD with pytest
User: "/alfred:2-build AUTH-001"
Claude: (creates RED test with pytest, GREEN implementation, REFACTOR with type hints)

### Example 2: Type checking validation
User: "mypy 타입 체크 실행"
Claude: (runs mypy --strict and reports type errors)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Python-specific review)
- alfred-debugger-pro (Python debugging)
