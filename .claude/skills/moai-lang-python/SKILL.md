---

name: moai-lang-python
description: Python best practices with pytest, mypy, ruff, black, and uv package management. Use when writing or reviewing Python code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Python Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Python code discussions, framework guidance, or file extensions such as .py. |
| Tier | 3 |

## What it does

Provides Python-specific expertise for TDD development, including pytest testing, mypy type checking, ruff linting, black formatting, and modern uv package management.

## When to use

- Engages when the conversation references Python work, frameworks, or files like .py.
- “Writing Python tests”, “How to use pytest”, “Python type hints”
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

## Examples
```bash
python -m pytest && ruff check . && black --check .
```

## Inputs
- Language-specific source directories (e.g. `src/`, `app/`).
- Language-specific build/test configuration files (e.g. `package.json`, `pyproject.toml`, `go.mod`).
- Relevant test suites and sample data.

## Outputs
- Test/lint execution plan tailored to the selected language.
- List of key language idioms and review checkpoints.

## Failure Modes
- When the language runtime or package manager is not installed.
- When the main language cannot be determined in a multilingual project.

## Dependencies
- Access to the project file is required using the Read/Grep tool.
- When used with `Skill("moai-foundation-langs")`, it is easy to share cross-language conventions.

## References
- Python Software Foundation. "Python Developer's Guide." https://docs.python.org/3/ (accessed 2025-03-29).
- Pytest. "pytest Documentation." https://docs.pytest.org/en/stable/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Python-specific review)
- alfred-debugger-pro (Python debugging)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
