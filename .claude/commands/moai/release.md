---
name: moai:release
description: "Automated release and version management for MoAI-ADK packages"
argument-hint: "[check|version|changelog|release|rollback] [major|minor|patch]"
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
---

# 🚀 MoAI-ADK Release Automation

> **Local-Only Tool**: This command is for local development only, similar to the Yoda system.
> Never deployed with package distributions.
> Use for PyPI releases, version bumping, changelog generation, and emergency rollbacks.

---

## Command Purpose

Automated release workflow for MoAI-ADK package:

1. **Check Quality** (`/moai:release check`)
   - Run pytest, mypy, ruff, black, bandit validation
   - Verify all tests pass and code quality standards met

2. **Bump Version** (`/moai:release version [major|minor|patch]`)
   - Update version in pyproject.toml
   - Update version in src/moai_adk/__init__.py
   - Update .moai/config/config.json

3. **Generate Changelog** (`/moai:release changelog`)
   - Analyze git history since last release
   - Generate CHANGELOG.md automatically
   - Include commit messages and metadata

4. **Release to PyPI** (`/moai:release release`)
   - Execute quality checks
   - Build package (uv build)
   - Publish to PyPI (requires PYPI_TOKEN)
   - Tag commit and create GitHub release

5. **Rollback Release** (`/moai:release rollback`)
   - Revert version changes
   - Delete PyPI release
   - Delete GitHub release tags
   - Restore previous commit state

---

## File Structure

```
.moai/release/
├── quality-check.sh          # Integrated quality validation
├── bump-version.py           # Semantic version management
├── generate-changelog.py     # Automated CHANGELOG generation
├── release-helper.sh         # Reusable utility functions
├── release-rollback.sh       # Emergency rollback automation
├── CHECKLIST.md             # Pre-release validation checklist
├── RELEASE_SETUP.md         # PyPI token and secrets setup
└── ROLLBACK_GUIDE.md        # Emergency procedures and recovery
```

---

## Quick Start

### 1. Pre-Release Checklist
```bash
/moai:release check           # Validate quality metrics
cat .moai/release/CHECKLIST.md  # Review 6-phase checklist
```

### 2. Bump Version
```bash
/moai:release version patch   # e.g., 0.25.4 → 0.25.5
/moai:release version minor   # e.g., 0.25.4 → 0.26.0
/moai:release version major   # e.g., 0.25.4 → 1.0.0
```

### 3. Generate Changelog
```bash
/moai:release changelog       # Creates CHANGELOG.md entry
git add CHANGELOG.md
git commit -m "docs: Update CHANGELOG for vX.X.X"
```

### 4. Release to PyPI
```bash
/moai:release release         # Complete release automation
# Requires: PYPI_TOKEN environment variable
# Steps:
#  1. Run quality checks
#  2. Build package
#  3. Publish to PyPI
#  4. Create GitHub release
#  5. Tag commit
```

### 5. Emergency Rollback
```bash
/moai:release rollback        # Revert everything
# Restores:
#  1. Version files
#  2. Removes PyPI release
#  3. Deletes GitHub release
#  4. Reverts git commits
```

---

## Configuration

### PyPI Token Setup
```bash
# 1. Generate token at https://pypi.org/manage/account/tokens/
# 2. Set environment variable:
export PYPI_TOKEN="pypi-AgEIcHlwaS5vcmc..."

# 3. Or add to .env (never commit):
echo "PYPI_TOKEN=..." >> .env
```

### GitHub Secrets
```bash
# For automatic GitHub releases:
# 1. GitHub Settings → Developer Settings → Personal Access Tokens
# 2. Create token with: repo, workflow, write:packages
# 3. Store in: ~/MoAI/MoAI-ADK/.github/workflows/secrets.json
```

### See Also
- `.moai/release/RELEASE_SETUP.md` - Detailed setup instructions
- `.moai/release/ROLLBACK_GUIDE.md` - Emergency procedures
- `.moai/release/CHECKLIST.md` - 6-phase release validation

---

## Implementation Notes

- **Script Location**: `.moai/release/` (not deployed with package)
- **Execution**: All scripts via `uv run` for consistent environment
- **Rollback Strategy**: Git history-aware with PyPI API integration
- **Safety Checks**: Pre-flight validation before each operation
- **Error Recovery**: Comprehensive error handling with rollback support

---

## See Also

- `/alfred:0-project` - Project initialization
- `/alfred:1-plan` - SPEC planning
- `/alfred:2-run` - Implementation
- `/alfred:3-sync` - Synchronization and validation

---

**Status**: Local-Only Development Tool
**Version**: 0.25.4+
**Deployment**: Excluded from PyPI distributions
