# Repository Guidelines

## Project Structure & Module Organization
- `src/moai_adk/` contains the CLI entrypoint plus modular subpackages; commands live in `cli/commands`, shared services in `core/` and `utils/`.
- Tests split between `tests/unit/` for fast specs and `tests/e2e/` for full Alfred workflows; keep generated artifacts in `.moai/` (notably `specs/`, `memory/`, and `reports/`).
- Published docs and images sit in `docs/`; helper scripts such as `security-scan.sh` stay under `scripts/`; build outputs land in `dist/` when packaging locally.

## Build, Test, and Development Commands
- `uv pip install -e ".[dev]"` sets up an editable install with pytest, ruff, and mypy aligned to `pyproject.toml`.
- `uv run moai --help` sanity-checks the CLI wiring; `uv run moai doctor` validates host prerequisites.
- `uv run pytest -n auto` drives the suite with coverage enabled; pair it with `uv run ruff check` and `uv run mypy src`.

## Coding Style & Naming Conventions
- Python files use 4-space indentation, 120-character lines, and descriptive snake_case; align new modules with related SPEC IDs to aid traceability.
- Preserve the existing `@SPEC/@TEST/@CODE/@DOC` breadcrumbs in docstrings and comments; introduce new identifiers when extending pipelines.
- Ruff governs linting and formatting—limit `ruff format` to files you touched and accept import re-ordering from `ruff check --fix` when necessary.

## Testing Guidelines
- Place fresh unit cases in `tests/unit/test_<feature>.py`; tag slower journeys with `@pytest.mark.e2e` within `tests/e2e/` files.
- The pytest profile already injects `--cov=src/moai_adk` and enforces an 85% minimum; add targeted assertions to protect coverage budgets.
- Share fixtures via `tests/conftest.py` and prefer reusable factory helpers over ad-hoc temporary directories.

## Commit & Pull Request Guidelines
- Follow the prevailing log format `{emoji} CATEGORY: summary` (e.g., `🟢 GREEN: tighten validator paths`) and call out relevant SPEC/TEST IDs in the body.
- Summarize problem, solution, and test evidence in PRs; attach CLI screenshots when terminal output changes.
- Link issues with `Closes #123` syntax, request at least one maintainer review, and note any `.moai/` artifact refresh that reviewers should pull locally.

## Security & Maintenance
- `bash scripts/security-scan.sh` wraps `pip-audit` and `bandit`; run it before tagging a release or merging dependency updates.
- Redact secrets from `.moai/memory` or generated reports prior to committing them.
- Regenerate `uv.lock` only when dependencies shift and highlight the change in commit and PR notes.

## Communication Guidelines
- 모든 이슈, PR, 리뷰, 문서화 코멘트는 한국어로 작성하고 대화하세요; 국제 협업이 필요한 경우에도 우선 한국어 기록을 남깁니다.
