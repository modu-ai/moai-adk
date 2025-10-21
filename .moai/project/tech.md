---
id: TECH-001
version: 0.1.2
status: active
created: 2025-10-01
updated: 2025-10-22
author: @tech-lead
priority: medium
---

# MoAI-ADK Technology Stack

## HISTORY

### v0.1.2 (2025-10-22)
- **UPDATED**: Template optimization complete (v0.4.1)
- **AUTHOR**: @Alfred (@project-manager)
- **SECTIONS**: Expanded with real MoAI-ADK stack (Python, uv, pytest, ruff, mypy)
- **CHANGES**: Added multi-platform support details, quality gates, security policies, deployment strategy

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

- **Language**: Python
- **Version Range**: ≥3.11, <3.14 (tested on 3.11, 3.12, 3.13)
- **Rationale**:
  - Excellent ecosystem for CLI tools (click, rich, typer)
  - Claude Code's Python hooks API (`alfred_hooks.py`)
  - Strong async/await support for future agent parallelization
  - Native support for JSON, YAML, TOML parsing (metadata-heavy workflows)
- **Package Manager**: **uv** (Astral's ultra-fast pip replacement)
  - 10–100× faster than pip/poetry
  - Built-in virtual environment management
  - Lock file support (`uv.lock`) for reproducible builds
  - Cross-platform binary (Windows/macOS/Linux)

### Multi-Platform Support

| Platform    | Support Level | Validation Tooling               | Key Constraints                                  |
| ----------- | ------------- | -------------------------------- | ------------------------------------------------ |
| **Windows** | Full          | GitHub Actions (windows-latest)  | Path separators (pathlib), Git line endings      |
| **macOS**   | Full          | GitHub Actions (macos-latest)    | Case-insensitive filesystem, Xcode dependencies  |
| **Linux**   | Full          | GitHub Actions (ubuntu-latest)   | Primary development environment                  |

**Cross-platform strategy**:
- Use `pathlib.Path` exclusively (no raw string paths)
- Git auto-converts line endings (`.gitattributes` configured)
- Shell commands via `shutil.which()` + subprocess (not direct bash calls)

## @DOC:FRAMEWORK-001 Core Frameworks & Libraries

### 1. Runtime Dependencies

```toml
[project.dependencies]
click = "^8.1.7"              # CLI framework (commands, options, help)
rich = "^13.9.4"              # Terminal formatting (tables, progress bars, syntax highlighting)
pyyaml = "^6.0.2"             # YAML parsing (SPEC front matter, config files)
jinja2 = "^3.1.4"             # Template rendering (SPEC/test/code generation)
gitpython = "^3.1.43"         # Git automation (branch creation, commits, tags)
anthropic = "^0.39.0"         # Claude API client (future: agent API integration)
```

**Dependency philosophy**:
- Minimize transitive dependencies (reduces supply chain risk)
- Prefer stdlib when performance/features are comparable
- Pin major versions, allow minor/patch updates (`^X.Y.Z` = `>=X.Y.Z, <X+1.0.0`)

### 2. Development Tooling

```toml
[project.optional-dependencies]
dev = [
  "pytest ^8.3.4",            # Test runner (TDD cycles)
  "pytest-cov ^6.0.0",        # Coverage measurement (≥85% target)
  "pytest-mock ^3.14.0",      # Mocking utilities (external API tests)
  "ruff ^0.8.4",              # Fast linter + formatter (replaces flake8, black, isort)
  "mypy ^1.13.0",             # Static type checker (TRUST Unified principle)
  "pre-commit ^4.0.1",        # Git hook automation (TRUST gates before commit)
]
```

### 3. Build System

- **Build Tool**: `uv` (handles both dependency resolution and packaging)
- **Package Format**: Python wheel (`.whl`) + source distribution (`.tar.gz`)
- **Distribution Target**: PyPI (primary), GitHub Releases (backup)
- **Performance Goals**:
  - `uv sync` (install deps): <5s (cold), <1s (cached)
  - `uv pip install -e .` (editable install): <3s
  - Full test suite: <30s (unit tests), <2min (integration tests)

## @DOC:QUALITY-001 Quality Gates & Policies

### Test Coverage (TRUST: **T**est First)

- **Target**: ≥85% line coverage, ≥80% branch coverage
- **Measurement Tool**: `pytest-cov` (generates HTML reports + terminal summary)
- **Failure Response**:
  - Block PR merge if coverage drops below 85%
  - Allow temporary waivers with explicit DEBT TAG in code comments
  - Weekly coverage trend report (flag regressions)

```bash
# Run tests with coverage
pytest --cov=src/moai_adk --cov-report=html --cov-report=term-missing --cov-fail-under=85
```

### Static Analysis (TRUST: **R**eadable + **U**nified)

| Tool       | Role                            | Config File        | Failure Handling                         |
| ---------- | ------------------------------- | ------------------ | ---------------------------------------- |
| **ruff**   | Linter + Formatter (all-in-one) | `pyproject.toml`   | Block commit if unfixed (pre-commit)     |
| **mypy**   | Type checker                    | `pyproject.toml`   | Warn on PRs, block on `main` merge       |
| **bandit** | Security linter (future)        | `.bandit`          | Warn on medium, block on high severity   |

**Ruff configuration** (`pyproject.toml`):
```toml
[tool.ruff]
line-length = 100
target-version = "py311"
select = ["E", "F", "I", "N", "W", "UP", "B", "C4", "SIM"]  # Pyflakes, imports, naming, etc.
ignore = ["E501"]  # Line too long (handled by formatter)
```

**Mypy configuration** (`pyproject.toml`):
```toml
[tool.mypy]
python_version = "3.11"
strict = true
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
```

### Automation Scripts (TRUST Gate Pipeline)

```bash
# Local quality gate (pre-commit hook)
ruff check . --fix                     # Auto-fix linting issues
ruff format .                          # Format code (black-compatible)
mypy src/moai_adk                      # Type check
pytest --cov=src/moai_adk --cov-fail-under=85  # Test + coverage

# CI/CD quality gate (GitHub Actions)
uv sync --dev                          # Install deps
ruff check . --no-fix                  # Fail on unfixed lint issues
mypy src/moai_adk                      # Type check
pytest --cov=src/moai_adk --cov-report=xml  # Generate coverage for Codecov
```

## @DOC:SECURITY-001 Security Policy & Operations (TRUST: **S**ecured)

### Secret Management

- **Policy**: Never commit secrets to version control (enforced via pre-commit hooks)
- **Tooling**:
  - Environment variables for API keys (`ANTHROPIC_API_KEY`, `GITHUB_TOKEN`)
  - `.env` files (local development only, `.gitignore`d)
  - Future: Integration with 1Password CLI, AWS Secrets Manager
- **Verification**:
  - `detect-secrets` pre-commit hook (scans diffs for leaked secrets)
  - GitHub secret scanning (alerts on accidental commits)

```bash
# Pre-commit hook example
repos:
  - repo: https://github.com/Yelp/detect-secrets
    hooks:
      - id: detect-secrets
        args: ['--baseline', '.secrets.baseline']
```

### Dependency Security

```toml
[tool.uv]
security_policy = "strict"  # Reject packages with known high/critical CVEs
```

- **Audit Tool**: `pip-audit` (scans `uv.lock` for known vulnerabilities)
- **Update Policy**: Patch/minor updates weekly, major updates quarterly (with SPEC approval)
- **Vulnerability Threshold**:
  - HIGH/CRITICAL: Block release immediately
  - MEDIUM: Fix within 7 days
  - LOW: Fix within 30 days or accept risk with documented DEBT TAG

```bash
# Run security audit
pip-audit --requirement uv.lock --desc
```

### Logging Policy

- **Log Levels**:
  - `DEBUG`: Development only (verbose agent traces, API requests/responses)
  - `INFO`: Production (command execution, agent handoffs, SPEC creation)
  - `WARNING`: Recoverable errors (missing optional config, deprecated features)
  - `ERROR`: Failures requiring user intervention (API auth failures, git conflicts)
- **Sensitive Data Masking**:
  - API keys: Show only first 8 chars (`sk-ant-...****`)
  - File paths: Redact user home directory (`~/...` instead of `/Users/username/...`)
  - Git commit messages: Preserve (public information)
- **Retention Policy**:
  - Local logs: 7 days (rotate via `logging.handlers.RotatingFileHandler`)
  - CI/CD logs: 90 days (GitHub Actions default)
  - Production telemetry (future): 1 year (compliance requirement)

## @DOC:DEPLOY-001 Release Channels & Strategy

### 1. Distribution Channels

- **Primary Channel**: PyPI (`pip install moai-adk`)
- **Secondary Channels**:
  - GitHub Releases (binaries for Windows/macOS/Linux via PyInstaller - future)
  - Direct installation from GitHub (`pip install git+https://github.com/mherod/MoAI-ADK.git`)
- **Release Procedure** (Semantic Versioning):
  1. Update `pyproject.toml` version (SSOT - Single Source of Truth)
  2. Run `/awesome:release-new {patch|minor|major}` (auto-generates CHANGELOG, git tag)
  3. Push to `main` → GitHub Actions builds + publishes to PyPI
  4. Create GitHub Release (Draft) with auto-generated notes
- **Versioning Policy**: SemVer (`MAJOR.MINOR.PATCH`)
  - `PATCH`: Bug fixes, documentation updates, internal refactoring
  - `MINOR`: New features, backward-compatible API changes, new Skills/agents
  - `MAJOR`: Breaking changes (CLI interface, config.json schema, SPEC format)
- **Rollback Strategy**:
  - PyPI: Publish new version with fixes (cannot delete/replace versions)
  - GitHub Releases: Mark as "Pre-release" if issues found
  - User-side: `pip install moai-adk==<previous-version>` (pin to last known good)

### 2. Developer Setup

```bash
# Clone repository
git clone https://github.com/mherod/MoAI-ADK.git
cd MoAI-ADK

# Install uv (if not already installed)
curl -LsSf https://astral.sh/uv/install.sh | sh  # macOS/Linux
# or: pip install uv  # Windows/fallback

# Create virtual environment and install dependencies
uv venv
source .venv/bin/activate  # macOS/Linux
# or: .venv\Scripts\activate  # Windows

# Install in editable mode with dev dependencies
uv pip install -e ".[dev]"

# Verify installation
moai-adk --version
pytest --version
```

### 3. CI/CD Pipeline (GitHub Actions: `.github/workflows/moai-gitflow.yml`)

| Stage          | Objective                       | Tooling                         | Success Criteria                    |
| -------------- | ------------------------------- | ------------------------------- | ----------------------------------- |
| **Lint**       | Code quality enforcement        | ruff check, ruff format --check | No unfixed lint issues              |
| **Type Check** | Static type validation          | mypy                            | No type errors                      |
| **Test**       | Unit + integration tests        | pytest --cov                    | All tests pass, coverage ≥85%       |
| **Build**      | Package verification            | uv build                        | Wheel + sdist build successfully    |
| **Publish**    | Deploy to PyPI (main branch)    | twine upload                    | Package published without conflicts |
| **Release**    | Create GitHub Release (tags)    | gh release create               | Draft release created with notes    |

**Trigger conditions**:
- `push` to `main`: Full pipeline (lint → test → build → publish)
- `pull_request`: Lint + test only (no publish)
- Manual `workflow_dispatch`: Full pipeline with custom version

## Environment Profiles

### Development (`dev`)

```bash
export MOAI_ENV=development
export LOG_LEVEL=DEBUG
export ANTHROPIC_API_KEY=<your-key>  # Required for agent interactions
moai-adk init  # Bootstrap new project
```

**Features**:
- Verbose logging (agent traces, API calls)
- Hot-reload enabled (watch mode for template changes)
- Relaxed TRUST gates (warn instead of block)

### Test (`test`)

```bash
export MOAI_ENV=test
export LOG_LEVEL=INFO
pytest tests/ --cov=src/moai_adk --cov-report=term-missing
```

**Features**:
- Mock Claude API responses (no real API calls)
- Ephemeral git repositories (in-memory testing)
- Strict TRUST gates (enforce all rules)

### Production (`production`)

```bash
export MOAI_ENV=production
export LOG_LEVEL=WARNING
pip install moai-adk
moai-adk --version
```

**Features**:
- Minimal logging (errors + warnings only)
- Strict TRUST gates (block on violations)
- Performance-optimized (caching enabled)

## @CODE:TECH-DEBT-001 Technical Debt Management

### Current Debt (as of v0.4.1)

1. **Agent API integration** – Priority: HIGH
   - Current: Agents invoked via natural language in commands
   - Target: Structured API with typed payloads (`.claude/agents/*/api.json`)
   - Blocker: Claude Code agent API still in beta

2. **Async agent execution** – Priority: MEDIUM
   - Current: Sequential agent handoffs (blocking)
   - Target: Parallel execution for independent tasks (e.g., lint + test simultaneously)
   - Benefit: 30–50% faster `/alfred:2-run` cycles

3. **Windows native Git** – Priority: MEDIUM
   - Current: Requires WSL or Git Bash on Windows
   - Target: Native Windows support via GitPython's Windows backend
   - Benefit: Eliminates WSL dependency for Windows users

4. **Skill usage telemetry** – Priority: LOW
   - Current: No tracking of which Skills are loaded/used
   - Target: Anonymous usage analytics (opt-in) to prioritize Skill improvements
   - Benefit: Data-driven Skill optimization

### Remediation Plan

- **Short term (1 month)**:
  - Document agent API schema (prepare for beta → GA)
  - Windows Git integration tests (CI/CD on windows-latest)

- **Mid term (3 months)**:
  - Implement async agent orchestration (asyncio-based)
  - Add Skill telemetry hooks (localStorage-based counters)

- **Long term (6+ months)**:
  - Agent marketplace (community-contributed agents)
  - Cloud-hosted Skill CDN (faster loading, versioned Skill packs)

## EARS Technical Requirements Guide

### Using EARS for the Stack

Apply EARS patterns when documenting technical decisions and quality gates:

#### Technology Stack EARS Example
```markdown
### Ubiquitous Requirements (Baseline)
- The system shall guarantee TypeScript type safety.
- The system shall provide cross-platform compatibility.

### Event-driven Requirements
- WHEN code is committed, the system shall run tests automatically.
- WHEN a build fails, the system shall notify developers immediately.

### State-driven Requirements
- WHILE in development mode, the system shall offer hot reloading.
- WHILE in production mode, the system shall produce optimized builds.

### Optional Features
- WHERE Docker is available, the system may support container-based deployment.
- WHERE CI/CD is configured, the system may execute automated deployments.

### Constraints
- IF a dependency vulnerability is detected, the system shall halt the build.
- Test coverage shall remain at or above 85%.
- Build time shall not exceed 5 minutes.
```

---

_This technology stack guides tool selection and quality gates when `/alfred:2-run` runs._
