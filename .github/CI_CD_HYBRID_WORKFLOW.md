# CI/CD Configuration for Hybrid Personal-Pro Workflow

## Overview
MoAI-ADK CI/CD pipeline automatically adapts to Personal and Team modes.

## Current CI/CD Workflows

### 1. Continuous Integration (ci.yml)
**Triggers**: PR or push to main/develop

**Jobs**:
- Code Quality (black, ruff, mypy)
- Unit Tests (pytest with coverage)
- Build Verification

**Branches**:
```yaml
on:
  pull_request:
    branches: [main, develop]  ✅ Both main and develop supported
  push:
    branches: [develop, main]  ✅ Both main and develop supported
```

### 2. Package Verification (package-verify.yml)
**Triggers**: Push/PR affecting source code

**Jobs**:
- Source Distribution Verification
- Wheel Verification
- Installation Verification

**Branches**:
```yaml
on:
  push:
    branches: ['**']           ✅ All branches (for development)
  pull_request:
    branches: [main, develop]  ✅ Main/develop only (for safety)
```

### 3. Release Pipeline (release.yml)
**Triggers**: Tag push (v*.*.*)

**Jobs**:
- Create GitHub Release
- Deploy to PyPI
- Create Release Notes

**Workflow**:
```
Tag (v0.25.11) → GitHub Release → PyPI
```

## Personal Mode CI/CD Flow

```
developer creates feature/SPEC-XXX
          ↓
    pushes to origin
          ↓
    PR created to main
          ↓
   CI/CD Checks (automatic)
   - Code quality (black, ruff, mypy)
   - Unit tests (pytest, ≥85% coverage)
   - Package verification
          ↓
   CI passes? → YES
          ↓
   Merge to main
          ↓
   Create tag (v0.25.XX)
          ↓
   GitHub Release + PyPI Deploy
```

## Team Mode CI/CD Flow (Future)

```
developer creates feature/SPEC-XXX
          ↓
    pushes to origin
          ↓
   PR created to develop
          ↓
   CI/CD Checks (automatic)
          ↓
   Code Review + Approval
          ↓
   Merge to develop
          ↓
   Integration Tests (develop branch)
          ↓
   PR: develop → main
          ↓
   CI/CD Final Checks
          ↓
   Merge to main
          ↓
   Create tag (v0.25.XX)
          ↓
   GitHub Release + PyPI Deploy
```

## Status Checks Required for Merge

### main Branch (Strict)
```yaml
required_status_checks:
  - CI (code-quality)
  - CI (test)
  - CI (build)
  - Package Verify
  - All required checks must pass
```

### develop Branch (Moderate, Team Mode)
```yaml
required_status_checks:
  - CI (code-quality)
  - CI (test)
  - Package Verify
  - Code review approval
```

## Automatic Workflow Selection

**Alfred** automatically selects correct workflow based on:

1. **Branch Detection**
   ```bash
   CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

   if [ "$CURRENT_BRANCH" = "main" ]; then
       # Personal mode: main-based workflow
       echo "Personal mode activated"
   elif [ "$CURRENT_BRANCH" = "develop" ]; then
       # Team mode (or team transition)
       echo "Team mode activated"
   fi
   ```

2. **Contributor Count**
   ```python
   contributors = get_git_contributors()

   if len(contributors) >= 3:
       # Automatically suggest Team mode
       suggest_team_mode = true
   fi
   ```

3. **Config-Based Selection**
   ```json
   {
     "git_strategy": {
       "personal": {
         "enabled": true,
         "workflow": "github-flow"
       },
       "team": {
         "enabled": false,
         "auto_switch_threshold": 3
       }
     }
   }
   ```

## Deployment Process

### PyPI Deployment
```yaml
# release.yml
on:
  push:
    tags:
      - "v*.*.*"  # Production PyPI
      - "v*.*.*-test"  # TestPyPI

jobs:
  deploy:
    - Create GitHub Release
    - Upload to PyPI (or TestPyPI)
    - Update changelog
```

### Vercel Documentation Deployment
```yaml
# docs-deploy.yml
on:
  push:
    branches: [main]  # Only main deploys to production docs
```

## Testing CI/CD Changes

### Local Testing
```bash
# Simulate PR to main
git push origin feature/SPEC-TEST-001 --force

# Check GitHub Actions
gh run list --repo modu-ai/moai-adk

# View specific run
gh run view 12345
```

### Manual Workflow Trigger
```bash
# Trigger CI manually
gh workflow run ci.yml --ref main

# Trigger package verification
gh workflow run package-verify.yml --ref main
```

## Monitoring CI/CD

### View Workflow Status
```bash
# List all workflows
gh workflow list --repo modu-ai/moai-adk

# View recent runs
gh run list --repo modu-ai/moai-adk --limit 10

# Check specific job
gh run view 12345 --json jobs
```

### Troubleshooting Failed Checks

**If Code Quality Fails**
```bash
# Auto-fix formatting
uv run black src/ tests/ .claude/

# Check types
uv run mypy src/

# Lint
uv run ruff check src/
```

**If Tests Fail**
```bash
# Run tests locally
uv run pytest tests/ --cov=src/moai_adk --cov-report=html

# View coverage
open htmlcov/index.html
```

**If Package Verification Fails**
```bash
# Run verification locally
uv run python scripts/verify-package-integrity.py
```

## Summary

| Aspect | Personal Mode | Team Mode | Status |
|--------|---------------|-----------|--------|
| **Base Branch** | main | develop | ✅ Configured |
| **CI Triggers** | main, develop | develop, main | ✅ Active |
| **Test Coverage** | ≥85% | ≥85% | ✅ Enforced |
| **Code Quality** | black, ruff, mypy | Same | ✅ Active |
| **Status Checks** | 3 checks | 4 checks | ✅ Configured |
| **Deployment** | Tag → PyPI | Tag → PyPI | ✅ Active |
| **Auto-switching** | Based on contributors | Threshold: 3 | ✅ Ready |

---

**Last Updated**: 2025-11-18
**Version**: 0.25.11
**Workflow**: Hybrid Personal-Pro with Automatic Adaptation
