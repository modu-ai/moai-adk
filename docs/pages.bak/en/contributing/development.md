---
title: Development Environment Setup
description: MoAI-ADK local development environment configuration and contribution guide
status: stable
---

# Development Environment Setup

This guide explains how to set up a local development environment for contributing to MoAI-ADK.

## Prerequisites

- Python 3.13+
- Git
- UV (Python package manager)
- Docker (optional)

## Development Environment Setup

### Step 1: Clone Repository

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
```

### Step 2: Install Development Dependencies

```bash
# Installation using UV (recommended)
uv sync --all-extras

# Or using pip
pip install -e ".[dev,test,docs]"
```

### Step 3: Setup Pre-commit Hooks

```bash
# Install pre-commit hooks
uv run pre-commit install

# Run pre-checks on all files
uv run pre-commit run --all-files
```

## Running Tests

### Full Test Suite

```bash
# Run all tests
uv run pytest

# With coverage report
uv run pytest --cov=src/moai_adk --cov-report=html
```

### Run Specific Tests

```bash
# Test specific file
uv run pytest tests/test_core.py

# Test specific function
uv run pytest tests/test_core.py::test_function_name

# Run by marker
uv run pytest -m integration
```

## Code Style Checks

### Linting

```bash
# Lint with Ruff
uv run ruff check src/ tests/

# Format with Black
uv run black src/ tests/

# Type check with mypy
uv run mypy src/moai_adk
```

### Auto-fix

```bash
# Auto-fix with Ruff
uv run ruff check --fix src/ tests/

# Auto-format with Black
uv run black src/ tests/
```

## Building Documentation

### Local Documentation Server

```bash
cd docs

# Start development server
uv run mkdocs serve

# Visit http://localhost:8000 in browser
```

### Production Build

```bash
# Generate static site
uv run mkdocs build

# Output: site/ directory
```

## Development Workflow

### Create Feature Branch

```bash
# Sync latest develop branch
git checkout develop
git pull origin develop

# Create feature branch
git checkout -b feature/SPEC-XXX

# Or use Alfred
/alfred:1-plan "Feature title"
```

### Local Development and Testing

```bash
# Write code
# ... make changes ...

# Run tests
uv run pytest

# Code style checks
uv run ruff check --fix src/
uv run black src/

# Type check
uv run mypy src/moai_adk
```

### Commit and Push

```bash
# Add changes
git add .

# Commit using Alfred (recommended)
/alfred:2-run SPEC-XXX

# Or manual commit
git commit -m "feat: feature description"
git push origin feature/SPEC-XXX
```

## Pull Request Process

1. **Create PR**: Create PR from feature branch to develop
2. **Auto Checks**: GitHub Actions runs automatic tests and linting
3. **Code Review**: Wait for maintainer review
4. **Merge**: Merge to develop branch after approval

## Debugging

### Log Level Setting

```bash
# Enable debug mode
export MOAI_DEBUG=true
uv run moai-adk init my-project
```

### VS Code Debugging

`.vscode/launch.json` example:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Python: Current File",
      "type": "python",
      "request": "launch",
      "program": "${file}",
      "console": "integratedTerminal"
    }
  ]
}
```

## Reference Documentation

- [Code Style Guide](style.md)
- [Release Process](releases.md)
- [Contributor Code of Conduct](index.md)

## Troubleshooting

### Dependency Errors

```bash
# Clear cache and reinstall
uv cache clean
uv sync --all-extras
```

### Test Failures

```bash
# Run tests with verbose output
uv run pytest -vv

# Run specific test only
uv run pytest tests/test_xxx.py::test_name -vv
```

### Documentation Build Errors

```bash
# Clear cache
rm -rf docs/site docs/.cache

# Rebuild
cd docs
uv run mkdocs build --strict
```

---

**Questions?** Ask questions on GitHub Issues or participate in Discussions!




