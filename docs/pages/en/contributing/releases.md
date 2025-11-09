---
title: Release Process
description: MoAI-ADK version management and release automation guide
status: stable
---

# Release Process

This guide explains the version management and release procedures for MoAI-ADK.

## Version Management Strategy

MoAI-ADK follows [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH

Example: 0.20.1
    │  │   │
    │  │   └─ PATCH: Bug fixes (maintains compatibility)
    │  └────── MINOR: Feature additions (maintains backward compatibility)
    └───────── MAJOR: Major changes (breaks compatibility)
```

## Release Cycle

### Development Phase (develop branch)

```
1. Develop in feature branch
   feature/SPEC-XXX

2. Create and merge PR to develop
   Review → CI/CD checks → Merge

3. Features accumulate in develop branch
   Multiple features and bug fixes included
```

### Release Preparation (release/ branch)

```
1. Create release branch from develop
   git checkout -b release/v0.20.0

2. Update version
   - src/moai_adk/__init__.py: __version__
   - pyproject.toml: version
   - CHANGELOG.md: Release notes

3. Final testing and bug fixes
   Fix only in release branch

4. Create PR to main
```

### Release Deployment (main branch)

```
1. Approve and merge PR (main)
   git merge release/v0.20.0

2. Create tag
   git tag -a v0.20.0 -m "Release v0.20.0"

3. PyPI deployment automation
   GitHub Actions runs automatically

4. Back-merge to develop
   Sync main → develop
```

## Using Alfred for Releases

MoAI-ADK provides release automation:

```bash
# Patch release (0.20.0 → 0.20.1)
/alfred:release-new patch

# Minor release (0.20.0 → 0.21.0)
/alfred:release-new minor

# Major release (0.20.0 → 1.0.0)
/alfred:release-new major

# Test mode (no actual deployment)
/alfred:release-new patch --dry-run

# Deploy to TestPyPI (testing)
/alfred:release-new patch --testpypi
```

## CHANGELOG Writing

`CHANGELOG.md` format:

```markdown
## [0.20.1] - 2025-11-07

### Added
- New feature 1
- New feature 2

### Fixed
- Bug fix 1
- Bug fix 2

### Changed
- Change 1
- Change 2

### Deprecated
- Deprecated feature

### Security
- Security-related fixes
```

## Version Management Files

### src/moai_adk/__init__.py

```python
"""
MoAI-ADK: Agentic Development Kit
"""

__version__ = "0.20.1"
__author__ = "GoosLab"
__license__ = "MIT"
```

### pyproject.toml

```toml
[project]
name = "moai-adk"
version = "0.20.1"
description = "MoAI-Agentic Development Kit"
```

## Release Checklist

Make sure to check before release:

- [ ] All features merged to develop branch
- [ ] All tests pass (pytest 100% ✓)
- [ ] Code linting passes (ruff, black, mypy ✓)
- [ ] CHANGELOG.md updated
- [ ] Version number consistency checked
  - `__version__` in `__init__.py`
  - `version` in `pyproject.toml`
- [ ] README and documentation up to date
- [ ] Release notes prepared

## Automated Release (GitHub Actions)

`.github/workflows/release.yml` example:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build
        run: uv build

      - name: Publish to PyPI
        run: uv publish
        env:
          UV_PUBLISH_TOKEN: ${{ secrets.PYPI_TOKEN }}
```

## Deployment Targets

### PyPI (Production)

```bash
# Install latest release
pip install moai-adk
```

### TestPyPI (Testing)

```bash
# Install test deployment
pip install -i https://test.pypi.org/simple/ moai-adk
```

### GitHub Releases

- Automatic release creation based on tags
- Include release notes
- Downloadable artifacts

## Emergency Hotfixes

When urgent bug fixes are needed:

```bash
# Create hotfix branch from main
git checkout main
git checkout -b hotfix/v0.20.2

# Fix bug and commit
# ... fixes ...

# Create PR to both main and develop
# main: For urgent deployment
# develop: For integration
```

## Release Maintainers

Releases are performed by:

- **Maintainer**: @goos
- **Co-Maintainer**: Community (optional)

## Reference Materials

- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Python Packaging Guide](https://packaging.python.org/)

---

**Questions?** Ask questions or discuss on GitHub Issues!




