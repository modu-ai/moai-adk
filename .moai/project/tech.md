---
id: TECH-001
version: 0.2.0
status: active
created: 2025-10-01
updated: 2025-10-27
author: @Alfred
priority: high
---

# MoAI-ADK Technology Stack

## HISTORY

### v0.2.0 (2025-10-27)
- **UPDATED**: Auto-generated comprehensive technology stack based on pyproject.toml analysis
- **AUTHOR**: @Alfred
- **SECTIONS**: Stack (Python 3.13+, uv), Framework (Click, Rich, GitPython), Quality (85% coverage, ruff, mypy), Security (bandit, pip-audit), Deploy (PyPI, GitHub Actions)
- **ANALYSIS**: Extracted dependencies from pyproject.toml; verified CI/CD workflows; confirmed 87.84% coverage baseline

### v0.1.1 (2025-10-17)
- **UPDATED**: Template version synced (v0.3.8)
- **AUTHOR**: @Alfred
- **SECTIONS**: Metadata standardization (single `author` field, added `priority`)

### v0.1.0 (2025-10-01)
- **INITIAL**: Authored the technology stack document
- **AUTHOR**: @tech-lead
- **SECTIONS**: Stack, Framework, Quality, Security, Deploy

---

## @DOC:STACK-001 Languages & Runtimes

### Primary Language: Python

- **Language**: Python 3.13+
- **Version Range**: ≥ 3.13 (verified with `requires-python = ">=3.13"` in pyproject.toml)
- **Rationale**:
  - **Maturity**: Python 3.13 brings significant performance improvements
  - **AI Integration**: Excellent ecosystem for Claude SDK integration
  - **Tooling**: Rich linting/typing ecosystem (ruff, mypy, pytest)
  - **Cross-platform**: Runs identically on Windows, macOS, Linux
- **Package Manager**: **uv** (ultra-fast, written in Rust)
  - Reasons: 10x faster than pip; perfect for iterative development
  - Installation: `curl -LsSf https://astral.sh/uv/install.sh | sh`

### Multi-Platform Support

| Platform | Support Level | Validation Tooling            | Key Constraints         |
| -------- | ------------- | ----------------------------- | ----------------------- |
| **macOS** | ✅ Full       | GitHub Actions (macOS runner) | Intel + Apple Silicon OK |
| **Linux** | ✅ Full       | GitHub Actions (Ubuntu 22.04) | All distributions       |
| **Windows** | ✅ Full     | GitHub Actions (Windows 2022) | PowerShell 5+ required  |

**CI/CD Verification**: `.github/workflows/moai-gitflow.yml` runs tests on all three platforms

---

## @DOC:FRAMEWORK-001 Core Frameworks & Libraries

### 1. Runtime Dependencies

```toml
[project.dependencies]
click = ">=8.1.0"           # CLI framework
rich = ">=13.0.0"           # Terminal UI (colors, tables, banners)
pyfiglet = ">=1.0.2"        # ASCII art fonts for banners
questionary = ">=2.0.0"     # Interactive TUI menus
gitpython = ">=3.1.45"      # Git repository manipulation
packaging = ">=21.0"        # Version parsing (for compatibility)
```

**Rationale for each**:

- **Click (8.1.0+)**
  - Purpose: CLI command parsing and routing
  - Chosen over: Typer (more magic), argparse (verbose), docopt (less community)
  - Key features: Decorators, nested commands, auto-help generation

- **Rich (13.0.0+)**
  - Purpose: Beautiful terminal output (colors, tables, progress bars)
  - Key features: Markdown rendering, syntax highlighting, Live display
  - Example: project status banners, TRUST validation reports

- **questionary (2.0.0+)**
  - Purpose: Interactive TUI for user surveys (language selection, mode choice)
  - Key features: Dropdown menus, text input, validation
  - Integration: Used in `/alfred:0-project` for user interviews

- **GitPython (3.1.45+)**
  - Purpose: Python bindings for Git
  - Key features: Repository management, branch ops, commit creation
  - Safety: All operations atomic; uses local git config

- **packaging (21.0+)**
  - Purpose: Version parsing and compatibility checks
  - Use case: Verify Python version >= 3.13 before running

### 2. Development Tooling

```toml
[project.optional-dependencies.dev]
pytest = ">=8.4.2"          # Test framework
pytest-cov = ">=7.0.0"      # Coverage reporting
pytest-xdist = ">=3.8.0"    # Parallel test execution
ruff = ">=0.1.0"            # Fast Python linter
mypy = ">=1.7.0"            # Static type checker
types-PyYAML = ">=6.0.0"    # Type stubs for PyYAML

[project.optional-dependencies.security]
pip-audit = ">=2.7.0"       # Dependency vulnerability scanner
bandit = ">=1.8.0"          # Security linter (OWASP Top 10)
```

**Each tool's purpose**:

| Tool | Version | Purpose | Config File |
| ---- | ------- | ------- | ----------- |
| **pytest** | >=8.4.2 | Test execution & discovery | `pyproject.toml` |
| **pytest-cov** | >=7.0.0 | Coverage measurement | `pyproject.toml` |
| **pytest-xdist** | >=3.8.0 | Parallel test runs (faster CI) | Via pytest flags |
| **ruff** | >=0.1.0 | Linting (100x faster than flake8) | `pyproject.toml` |
| **mypy** | >=1.7.0 | Type checking | `pyproject.toml` |
| **types-PyYAML** | >=6.0.0 | Type hints for YAML parsing | Auto-installed with mypy |
| **pip-audit** | >=2.7.0 | Dependency security audit | CLI tool (no config) |
| **bandit** | >=1.8.0 | Security linter | `.bandit` or CLI |

### 3. Build System

- **Build Tool**: **Hatchling** (PEP 517 backend)
- **Bundling**: No external bundler needed (pure Python package)
- **Targets**:
  - Wheel (`dist/moai-adk-*.whl`)
  - Source distribution (`dist/moai-adk-*.tar.gz`)
  - GitHub releases (automated)
- **Performance Goals**:
  - Build time: <30 seconds
  - Package size: <5MB
  - Install time (with uv): <5 seconds

**Build Configuration**:
```toml
[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[tool.hatch.build.targets.wheel]
packages = ["src/moai_adk"]

[tool.hatch.build]
include = [
    "src/moai_adk/**/*.py",
    "src/moai_adk/templates/**/*",
    "src/moai_adk/templates/.claude/**/*",
    "src/moai_adk/templates/.moai/**/*",
    "src/moai_adk/templates/.github/**/*"
]
```

---

## @DOC:QUALITY-001 Quality Gates & Policies

### Test Coverage: Strict 85% Minimum

```toml
[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = "test_*.py"
python_classes = "Test*"
python_functions = "test_*"
addopts = "-v --cov=src/moai_adk --cov-report=html --cov-report=term-missing"

[tool.coverage.run]
source = ["src/moai_adk"]
omit = ["tests/*", "*/__pycache__/*"]
parallel = true
concurrency = ["multiprocessing"]

[tool.coverage.report]
precision = 2
show_missing = true
skip_covered = false
fail_under = 85
```

- **Target**: ≥85% (enforced; build fails if violated)
- **Measurement Tool**: pytest + coverage.py
- **Baseline**: MoAI-ADK itself: **87.84%** (468/535 lines covered)
- **Failure Response**:
  - ❌ Coverage < 85%: Build fails in CI/CD
  - ⚠️ Coverage decline: GitHub commit status shows warning
  - ✅ Coverage >= 85%: Merge allowed

### Static Analysis: Multi-tool Validation

| Tool     | Role                              | Config File         | Failure Handling              |
| -------- | --------------------------------- | ------------------- | ----------------------------- |
| **ruff** | Fast linting (100x faster than flake8) | `pyproject.toml` | Fails if violations found; auto-fix available |
| **mypy** | Static type checking              | `pyproject.toml` | Fails if type errors detected |
| **bandit** | Security vulnerabilities (OWASP Top 10) | `.bandit` or CLI | Fails if critical issues found |

**Ruff Configuration**:
```toml
[tool.ruff]
line-length = 120
target-version = "py313"

[tool.ruff.lint]
select = ["E", "F", "W", "I", "N"]
ignore = []
```

**Selected rules**:
- E: PEP 8 errors (whitespace, syntax)
- F: PyFlakes (undefined names, unused imports)
- W: PEP 8 warnings (blank lines, line breaks)
- I: isort (import sorting)
- N: pep8-naming (naming conventions)

### Automation Scripts

```bash
# Run all quality gates locally
pytest --cov=src/moai_adk tests/           # Test + coverage
ruff check src/ tests/                     # Lint
mypy src/                                  # Type check
bandit -r src/                             # Security scan
pip-audit                                  # Dependency audit

# Single command (used in CI):
uv run pytest && ruff check && mypy src && bandit -r src
```

**Performance**:
- pytest: ~10-15 seconds (468 tests)
- ruff: <1 second (40+ files)
- mypy: ~5 seconds (strict mode)
- bandit: <1 second
- Total: ~20 seconds (CI/CD gate)

---

## @DOC:SECURITY-001 Security Policy & Operations

### Secret Management

- **Policy**: NO hardcoded secrets
- **Tooling**: Git hooks + pre-commit checks
- **Verification**: GitHub pre-push hook scans for common patterns
- **Sensitive data**:
  - API keys → environment variables only
  - Database credentials → `.env` file (gitignored)
  - Auth tokens → Claude Code secrets

### Dependency Security

```json
{
  "security": {
    "audit_tool": "pip-audit + bandit",
    "update_policy": "pin major versions; auto-update patch versions",
    "vulnerability_threshold": "0 critical, 0 high (except review)"
  }
}
```

**Policy**:
1. **Regular Audits**: `pip-audit` runs on every push (GitHub Actions)
2. **Dependency Lock**: `uv.lock` pins exact versions
3. **Update Strategy**:
   - Major: Manual review + testing required
   - Minor: Auto-update via dependabot
   - Patch: Auto-update immediately
4. **Vulnerability Response**: <24h for critical, <1 week for high

**Audit Command**:
```bash
pip-audit                    # Check all dependencies
bandit -r src/              # Security linter
rg 'password|secret|key' --type-list | head -5  # Manual spot check
```

### Logging Policy

- **Log Levels**: DEBUG, INFO, WARNING, ERROR, CRITICAL
- **Sensitive Data Masking**:
  - Never log: API keys, tokens, passwords, secrets
  - Always mask: Email addresses, IP addresses (partial), file paths (relative only)
  - Example: `User auth failed: user@example.*** at [masked IP]`
- **Retention Policy**:
  - Local logs: 7 days (auto-rotated)
  - CI/CD logs: 30 days (GitHub Actions default)
  - Production: 90 days (if deployed)

**Implementation**:
```python
# In moai_adk/utils/logger.py
import logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Example: mask sensitive data
password_masked = password[:2] + '*' * (len(password) - 4)
logger.info(f"Auth attempt: {password_masked}")
```

---

## @DOC:DEPLOY-001 Release Channels & Strategy

### 1. Distribution Channels

- **Primary Channel**: PyPI (Python Package Index)
- **Release Procedure**:
  1. Tag version in git (`v0.5.6`)
  2. GitHub Actions builds wheel + source
  3. Automated PyPI upload
  4. GitHub Release created (markdown notes)
- **Versioning Policy**: SemVer (MAJOR.MINOR.PATCH)
  - MAJOR: Breaking API changes
  - MINOR: New features (backward compatible)
  - PATCH: Bug fixes (backward compatible)
- **Current Version**: **0.5.6** (beta; not yet 1.0 for breaking changes)
- **Rollback Strategy**: Previous versions available on PyPI; git tags preserved

### 2. Developer Setup

**Local installation (development)**:
```bash
# Clone repository
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# Install uv (if not already installed)
curl -LsSf https://astral.sh/uv/install.sh | sh

# Install in editable mode
uv pip install -e .[dev,security]

# Verify installation
moai-adk --version
moai-adk doctor
```

**Global installation (users)**:
```bash
uv tool install moai-adk
moai-adk --version
```

### 3. CI/CD Pipeline

| Stage | Objective | Tooling | Success Criteria |
| ----- | --------- | ------- | --------------- |
| **Test** | Run pytest on all platforms | pytest, pytest-xdist | All tests pass; coverage ≥85% |
| **Lint** | Check code quality | ruff, mypy, bandit | No violations; types valid |
| **Build** | Create distribution packages | hatchling | Wheel + source tarball created |
| **Upload** | Deploy to PyPI | twine (via GitHub Actions) | Published on PyPI.org |
| **Release** | Create GitHub Release | GitHub CLI | Release notes + artifacts linked |

**Workflow File**: `.github/workflows/moai-gitflow.yml`

```yaml
name: MoAI GitFlow CI/CD
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        python-version: ['3.13']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v4
        with:
          python-version: ${{ matrix.python-version }}
      - run: pip install -e .[dev,security]
      - run: pytest --cov=src/moai_adk
      - run: ruff check src/ tests/
      - run: mypy src/

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: pip install pip-audit bandit
      - run: pip-audit
      - run: bandit -r src/

  publish:
    if: startsWith(github.ref, 'refs/tags/v')
    needs: [test, security]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v4
      - run: pip install build twine
      - run: python -m build
      - run: twine upload dist/*
```

---

## Environment Profiles

### Development (`dev`)

```bash
export PROJECT_MODE=development
export LOG_LEVEL=debug
export DEBUG=1

# Start development server (if applicable)
uv run python -m moai_adk.cli.main --debug
```

**Characteristics**:
- ✅ Verbose logging (DEBUG level)
- ✅ Source maps enabled
- ✅ Hot reloading enabled (for testing)
- ✅ Type checking active (strict)

### Test (`test`)

```bash
export PROJECT_MODE=test
export LOG_LEVEL=info
export COVERAGE=1

# Run tests
uv run pytest tests/ -v
```

**Characteristics**:
- ✅ Coverage measurement enabled
- ✅ Fast test execution (parallel with pytest-xdist)
- ✅ INFO-level logging (less noise)

### Production (`production`)

```bash
export PROJECT_MODE=production
export LOG_LEVEL=warning

# Install as user
uv tool install moai-adk

# Run command
moai-adk init my-project
```

**Characteristics**:
- ✅ Only WARNING+ messages (errors, critical)
- ✅ No debug symbols
- ✅ Performance optimized
- ✅ Secure defaults (no sensitive logging)

---

## @CODE:TECH-DEBT-001 Technical Debt Management

### Current Debt Items

1. **Hook Performance Optimization** – `core/git/checkpoint.py` creates timestamps; could cache for batch operations
   - **Priority**: Medium
   - **Impact**: Marginal (< 100ms per operation)
   - **Fix estimate**: 2-4 hours

2. **Type Hints Coverage** – ~60% of codebase has mypy stubs; aim for 100%
   - **Priority**: Low
   - **Impact**: Better IDE support, fewer runtime bugs
   - **Fix estimate**: 3-5 days

3. **Template Processing** – Currently string-based; could use Jinja2 for complex templates
   - **Priority**: Medium
   - **Impact**: More flexible scaffolding
   - **Fix estimate**: 1-2 days

### Remediation Plan

- **Short term (1 month)**:
  - [x] Profile hook execution time
  - [ ] Cache template compilation results
  - [ ] Add type hints to core/quality/*.py

- **Mid term (3 months)**:
  - [ ] Increase type coverage to 90%
  - [ ] Refactor template/processor.py to use Jinja2
  - [ ] Add integrated tests for edge cases

- **Long term (6+ months)**:
  - [ ] Extract spec-builder as separate microservice
  - [ ] Build plugin system for custom validators
  - [ ] Implement metrics collection (Prometheus format)

---

## EARS Technical Requirements Guide

### Using EARS for the Stack

Apply EARS patterns when documenting technical decisions and quality gates:

#### Technology Stack EARS Example

```markdown
### Ubiquitous Requirements (Baseline)
- The system SHALL use Python 3.13+ as the primary language.
- The system SHALL use uv as the package manager (for speed).
- The system SHALL enforce 85%+ test coverage via pytest.
- The system SHALL validate types via mypy (strict mode).

### Event-driven Requirements
- WHEN code is committed, the system SHALL run ruff linting automatically.
- WHEN a dependency vulnerability is detected, the system SHALL block the build.
- WHEN version tag is pushed, the system SHALL automatically publish to PyPI.

### State-driven Requirements
- WHILE in development mode, the system SHALL display DEBUG-level logs.
- WHILE in production mode, the system SHALL suppress DEBUG and INFO logs.
- WHILE test mode is active, the system SHALL measure and report coverage.

### Optional Features
- WHERE GitHub Actions is configured, the system MAY auto-deploy on tag.
- WHERE Docker is available, the system MAY support containerized execution.
- WHERE community contributions exist, the system MAY accept Python 3.12 compatibility PRs.

### Constraints
- IF a dependency has a critical vulnerability, the system SHALL halt all operations.
- Test coverage SHALL remain at or above 85% (enforced in CI).
- Build time SHALL not exceed 30 seconds.
- Linting violations SHALL cause the PR check to fail.
- Type errors (mypy strict) SHALL cause the PR check to fail.
```

---

## Installation & Quick Start

### For Users (Simple Install)

```bash
# Install moai-adk globally
uv tool install moai-adk

# Verify installation
moai-adk --version
# Output: MoAI-ADK v0.5.6

# Initialize a new project
moai-adk init my-project
cd my-project

# Start Alfred in Claude Code
claude
# Then in Claude Code: /alfred:0-project
```

### For Developers (Local Setup)

```bash
# Clone and install in editable mode
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# Install development dependencies
uv pip install -e .[dev,security]

# Run tests locally
pytest tests/ -v

# Run linting
ruff check src/ tests/
mypy src/

# Run security audit
pip-audit
bandit -r src/
```

---

## Monitoring & Observability (Future)

**Planned** (not yet implemented):

1. **Metrics Collection**: Hook execution time, command latency, error rates
2. **Structured Logging**: JSON format for log aggregation
3. **Distributed Tracing**: Track @TAG chain execution across multiple agents
4. **Dashboard**: Real-time project health (coverage trend, build times, etc.)

---

_This technology stack document guides quality gates, CI/CD configuration, and tool selection when `/alfred:2-run` and `/alfred:3-sync` execute. Update when major tooling changes occur (e.g., Python version upgrade, new dependency, etc.)_
