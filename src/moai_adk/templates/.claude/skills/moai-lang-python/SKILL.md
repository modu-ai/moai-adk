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
- 언어별 소스 디렉터리(e.g. `src/`, `app/`).
- 언어별 빌드/테스트 설정 파일(예: `package.json`, `pyproject.toml`, `go.mod`).
- 관련 테스트 스위트 및 샘플 데이터.

## Outputs
- 선택된 언어에 맞춘 테스트/린트 실행 계획.
- 주요 언어 관용구와 리뷰 체크포인트 목록.

## Failure Modes
- 언어 런타임이나 패키지 매니저가 설치되지 않았을 때.
- 다중 언어 프로젝트에서 주 언어를 판별하지 못했을 때.

## Dependencies
- Read/Grep 도구로 프로젝트 파일 접근이 필요합니다.
- `Skill("moai-foundation-langs")`와 함께 사용하면 교차 언어 규약 공유가 용이합니다.

## References
- Python Software Foundation. "Python Developer's Guide." https://docs.python.org/3/ (accessed 2025-03-29).
- Pytest. "pytest Documentation." https://docs.pytest.org/en/stable/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Python-specific review)
- alfred-debugger-pro (Python debugging)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
