---
name: moai-lang-python
description: Python best practices with pytest, mypy, ruff, black, and uv package management
allowed-tools:
  - Read
  - Bash
tier: 3
auto-load: "true"
---

# Python Expert

## What it does

Provides Python-specific expertise for TDD development, including pytest testing, mypy type checking, ruff linting, black formatting, and modern uv package management.

## When to use

- "Python 테스트 작성", "pytest 사용법", "Python 타입 힌트", "Python 백엔드", "데이터 과학", "머신러닝", "자동화 스크립트"
- "FastAPI", "Django", "Flask", "Django REST Framework"
- "pandas", "NumPy", "scikit-learn", "TensorFlow", "PyTorch"
- Automatically invoked when working with Python projects
- Python SPEC implementation (`/alfred:2-run`)

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

## Modern Python (3.10+)

**Recommended Version**: Python 3.11+ for production, 3.10+ for legacy support

**Modern Features**:
- **Pattern matching** (3.10+): `match/case` statements
- **Type union syntax** (3.10+): `int | str` instead of `Union[int, str]`
- **Structural pattern matching** (3.10+): Complex data unpacking
- **Self type** (3.11+): Self-referential type hints
- **Exception groups** (3.11+): Multiple exception handling
- **Async context managers**: `async with` for resource management

**Version Check**:
```bash
python --version  # Ensure Python 3.10+
python -m py_compile script.py  # Syntax check
```

## Package Management Commands

### Using uv (Recommended - Modern & Fast)
```bash
# Initialize project
uv init my_project
cd my_project

# Install dependencies
uv pip install pytest mypy ruff black

# Create virtual environment
uv venv

# Activate venv
source .venv/bin/activate  # macOS/Linux
.\.venv\Scripts\activate  # Windows

# Install from pyproject.toml
uv pip install -e .

# Add/remove dependencies
uv pip install numpy pandas
uv pip remove numpy

# Freeze requirements
uv pip freeze > requirements.txt

# Run tests
uv run pytest

# Check formatting
uv run ruff check .
uv run black . --check

# Type checking
uv run mypy .
```

### Using pip (Traditional)
```bash
# Virtual environment
python -m venv .venv
source .venv/bin/activate

# Install
pip install pytest mypy ruff black

# Requirements
pip install -r requirements.txt
pip freeze > requirements.txt
```

### Using poetry (Dependency Management)
```bash
poetry init  # Initialize
poetry add pytest mypy ruff black  # Add dependencies
poetry install  # Install all
poetry run pytest  # Run commands
poetry update  # Update dependencies
poetry build  # Build package
poetry publish  # Publish to PyPI
```

## Examples

### Example 1: TDD with pytest
User: "/alfred:2-run AUTH-001"
Claude: (creates RED test with pytest, GREEN implementation, REFACTOR with type hints)

### Example 2: Type checking validation
User: "mypy 타입 체크 실행"
Claude: (runs mypy --strict and reports type errors)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Python-specific review)
- alfred-debugger-pro (Python debugging)
