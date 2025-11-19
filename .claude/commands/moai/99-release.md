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

## EXCEPTION: Local-Only Development Tool

This command is exempt from "Zero Direct Tool Usage" principle because:

1. **Local development only** - Not distributed with package distributions
2. **Maintainer-only usage** - GoosLab (project owner) exclusive access
3. **Direct system access required** - PyPI release automation requires direct shell commands for:
   - Version bumping in multiple files
   - Package building and publishing
   - GitHub release tag management
4. **Local-only tool exception pattern** - Similar to `/moai:9-feedback` command

**Production commands** (`/moai:0-project`, `/moai:1-plan`, `/moai:2-run`, `/moai:3-sync`) must strictly adhere to agent delegation principle.

---

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
   - Update version in src/moai_adk/**init**.py
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

## 🎯 Release Strategy: Personal vs Team Mode

### Personal Mode (GitHub Flow)

```
feature/SPEC-XXX (local)
    ↓
main (fast-forward merge, ~10 min)
    ↓
Auto-tag vX.X.X
    ↓
CI/CD → PyPI (GitHub Actions)
```

**Commands**:

```bash
# 1. Switch to main (if needed)
git checkout main
git merge feature/SPEC-XXX

# 2. Bump version (patch/minor/major)
/moai:release version patch      # 0.25.11 → 0.25.12

# 3. Generate changelog
/moai:release changelog

# 4. Commit and push
git add CHANGELOG.md
git commit -m "chore: Release v0.25.12"
git push origin main

# 5. CI/CD auto-handles PyPI deployment
# No need for manual /moai:release release!
```

### Team Mode (Git-Flow)

```
feature/SPEC-XXX
    ↓
develop (Pull Request)
    ↓
main (Release PR, ~30 min)
    ↓
Auto-tag vX.X.X
    ↓
CI/CD → PyPI (GitHub Actions)
```

**Commands**:

```bash
# 1. Create feature branch (from develop)
git checkout -b feature/SPEC-XXX

# 2. After implementation, create PR
gh pr create --base develop

# 3. After merge to develop, prepare release
git checkout main
git merge develop

# 4. Bump version
/moai:release version patch

# 5. Generate changelog
/moai:release changelog

# 6. Commit and push
git add CHANGELOG.md
git commit -m "chore: Release v0.25.12"
git push origin main

# 7. CI/CD auto-handles PyPI deployment
```

**Current Mode**: Team Mode (develop-based)

- Auto-detection: 3+ contributors → Git-Flow
- Manual override: Edit config.json git_strategy.team.enabled

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

### 4. Release to PyPI (AUTOMATED via CI/CD)

**✅ DO NOT run `/moai:release release` manually!**

Release is **completely automated** via GitHub Actions:

```bash
# Just push to main with proper version + changelog
git push origin main

# GitHub Actions automatically:
# 1. Detects version bump in pyproject.toml
# 2. Runs quality checks (tests, linting)
# 3. Builds package
# 4. Publishes to PyPI
# 5. Creates GitHub Release with changelog
```

**CI/CD Pipeline**: `.github/workflows/release.yml`

- **Trigger**: Push to main branch (with tag v*.*.\*)
- **Condition**: All tests must pass
- **Action**: Auto build → PyPI publish → GitHub Release
- **Secrets**: PYPI_API_TOKEN (configured in GitHub)

**Requirements**:

- PYPI_API_TOKEN secret configured in GitHub Settings
- Version tag matches format: `v*.*.*`
- All tests pass
- Code quality standards met

**Manual Override (Local Testing Only)**:

```bash
# Only for testing locally (do NOT use in production)
/moai:release release       # Test locally first
# ⚠️ This is for development testing only
# Always use CI/CD for actual PyPI releases
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

### GitHub Release 문서 작성 규칙

```
📝 GitHub Release Notes 포맷:

## 🎯 한글 섹션 (한국어)
- 기능 설명
- 버그 수정
- 개선사항
- 주의사항

---

## 🎯 English Section
- Feature descriptions (English)
- Bug fixes (English)
- Improvements (English)
- Notes (English)

🤖 Generated with Claude Code

Co-Authored-By: 🎩 Alfred@MoAI
```

**규칙**:

1. 항상 한글 섹션 먼저 작성
2. `---` 구분선으로 구분
3. 그 다음 영문 섹션 작성
4. Footer: 🤖 Generated with Claude Code + Co-Authored-By

### See Also

- `.moai/release/RELEASE_SETUP.md` - Detailed setup instructions
- `.moai/release/ROLLBACK_GUIDE.md` - Emergency procedures
- `.moai/release/CHECKLIST.md` - 6-phase release validation
- `.github/workflows/release.yml` - Automated PyPI deployment (CI/CD)

---

## Implementation Notes

- **Script Location**: `.moai/release/` (not deployed with package)
- **Execution**: All scripts via `uv run` for consistent environment
- **Rollback Strategy**: Git history-aware with PyPI API integration
- **Safety Checks**: Pre-flight validation before each operation
- **Error Recovery**: Comprehensive error handling with rollback support

## 🔄 Automated CI/CD Deployment

### Main Branch → PyPI Automatic Deployment

**Trigger**: Push to main with tag `v*.*.*`

**Workflow**: `.github/workflows/release.yml`

**Steps** (자동 실행):

1. 코드 품질 검증 (Quality checks)
2. 패키지 빌드 (Build package)
3. PyPI 배포 (Publish to PyPI)
4. GitHub Release 생성 (Create GitHub Release)
5. 배포 완료 알림 (Post deployment comment)

**Requirements**:

- PYPI_API_TOKEN secret configured in GitHub
- Version tag must match `v*.*.*` format
- All tests must pass
- Code quality standards must be met

**Manual Override**:

```bash
# Local testing before main push
/moai:release release       # Test locally
git add .                   # Stage changes
git commit -m "Release v0.X.X"
git push origin main        # Triggers CI/CD
# → CI/CD automatically handles PyPI deployment
```

**Monitoring**:

- GitHub Actions: `.github/workflows/release.yml`
- PyPI Package: https://pypi.org/project/moai-adk/
- GitHub Releases: Releases page

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

## ⚡️ EXECUTION DIRECTIVE

**You must NOW execute the requested subcommand immediately.**

1. Analyze the arguments (check, version, changelog, release, rollback).
2. Execute the corresponding bash command or script using `Bash` tool.
3. Do NOT just describe what you will do. DO IT.
