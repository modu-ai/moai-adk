# MoAI-ADK Deployment Guide

> Last Updated: 2026-01-13
> Target: 1-Person Open Source Development

---

## Deployment Philosophy

### Core Principles

1. **Deploy First, Test Later**: Testing failures should not block releases
2. **Rapid Iteration**: Enable quick fixes and hotfixes without complex workflows
3. **Pragmatic Quality**: Tests inform quality, not enforce it
4. **Developer Autonomy**: Single developer can ship independently

---

## Deployment Workflows

### Workflow 1: Simple Deploy (Recommended)

**File**: `.github/workflows/deploy-simple.yml`

**When to Use**:
- Regular releases
- Hotfixes
- Quick iterations

**Trigger**:
```bash
git tag v1.1.0
git push origin v1.1.0
```

**Features**:
- Build and publish to PyPI
- Create GitHub Release automatically
- Non-blocking smoke tests
- 15-minute timeout
- Single job, simple debugging

**Manual Trigger**:
```bash
gh workflow run deploy-simple.yml -f version=1.1.0
```

### Workflow 2: Unified Release Pipeline

**File**: `.github/workflows/release.yml`

**When to Use**:
- Staged releases (TestPyPI → PyPI)
- Manual control over environment
- Comprehensive validation needed

**Trigger**:
```bash
# Tag push (auto deploy to PyPI)
git tag v1.1.0
git push origin v1.1.0

# Or manual with staging
gh workflow run release.yml -f version=1.1.0 -f target_environment=test
```

**Features**:
- Environment selection (test/production)
- Comprehensive version validation
- Config file consistency checks
- Bilingual release notes
- Multi-stage deployment

### Workflow 3: CI Pipeline

**File**: `.github/workflows/ci.yml`

**When to Use**:
- Pull request validation
- Development branch testing
- Coverage tracking

**Trigger**:
- Push to `main` or `develop`
- Pull requests to `main` or `develop`

**Features**:
- Multi-version Python testing (3.11-3.14)
- Non-blocking test execution
- Coverage reporting (non-blocking)
- Security scanning (informational)
- Benchmark tracking

---

## Step-by-Step Release Process

### Standard Release

1. **Update version**:
```bash
# Update pyproject.toml (single source of truth)
vim pyproject.toml  # version = "1.1.0"

# Sync all versions
.github/scripts/sync-versions.sh 1.1.0
```

2. **Commit changes**:
```bash
git add .
git commit -m "chore: bump version to 1.1.0"
git push origin main
```

3. **Create tag**:
```bash
git tag v1.1.0
git push origin v1.1.0
```

4. **Monitor deployment**:
```bash
gh run list --workflow=deploy-simple.yml
gh run view --watch
```

### Hotfix Release

1. **Fix on main**:
```bash
git checkout main
# Make fix
git commit -m "fix: critical bug"
```

2. **Quick release**:
```bash
# Bump patch version
vim pyproject.toml  # 1.1.0 → 1.1.1

# Tag and push
git tag v1.1.1
git push origin main v1.1.1
```

---

## Testing Strategy

### Non-Blocking Tests

**CI Pipeline**:
```yaml
- name: Test
  run: uv run pytest tests/
  continue-on-error: true  # Always deploy
```

**Rationale**:
- Test failures inform quality, don't enforce it
- Deployed packages can be yanked if critical issues found
- Open source users are early testers
- Faster release cycle enables quicker fixes

### Quality Gates

**Recommended Workflow**:
1. Run tests locally before committing
2. Review CI results after deployment
3. Monitor user feedback
4. Ship fixes for critical issues

---

## Version Synchronization

### Single Source of Truth

**Authoritative**: `pyproject.toml`
```toml
[project]
version = "1.1.0"
```

### Sync Script

```bash
.github/scripts/sync-versions.sh 1.1.0
```

**Updates**:
- `src/moai_adk/version.py` - Fallback version
- `.moai/config/sections/system.yaml` - System version
- `.moai/config/config.yaml` - Config version
- Template files - Use `{{MOAI_VERSION}}` placeholder

### Verification

Before release, verify consistency:
```bash
grep -r "1.1.0" pyproject.toml src/moai_adk/version.py .moai/config/
```

---

## Troubleshooting

### Release Workflow Fails

**Issue**: Version mismatch error

**Solution**:
```bash
# Check current versions
grep version pyproject.toml
grep version src/moai_adk/version.py

# Run sync script
.github/scripts/sync-versions.sh 1.1.0

# Commit and retry
git add .
git commit -m "chore: sync versions"
git push
```

### PyPI Upload Fails

**Issue**: Version already exists

**Solution**:
```bash
# Check if version exists
pip index versions moai-adk

# Bump version
vim pyproject.toml  # 1.1.0 → 1.1.1

# Delete old tag (local + remote)
git tag -d v1.1.0
git push origin :refs/tags/v1.1.0

# Create new tag
git tag v1.1.1
git push origin v1.1.1
```

### Tests Failing in CI

**Issue**: Tests fail but deployment continues

**Solution**:
- Review test results in CI logs
- Fix critical issues locally
- Ship hotfix if needed
- Non-blocking by design

---

## Comparison Matrix

| Feature | Simple Deploy | Unified Release | CI Pipeline |
|---------|--------------|-----------------|-------------|
| **Complexity** | Low | High | Medium |
| **Setup Time** | 2 min | 10 min | 5 min |
| **Use Case** | Regular releases | Staged releases | PR validation |
| **Test Blocking** | No | No | No |
| **Environments** | PyPI only | TestPyPI + PyPI | None |
| **Manual Control** | Yes | Yes | No |
| **Best For** | 1-person dev | Team projects | All projects |

---

## Best Practices

### For 1-Person Open Source

1. **Use Simple Deploy** for most releases
2. **Run tests locally** before pushing
3. **Monitor PyPI downloads** after release
4. **Respond to issues quickly**
5. **Keep changelog updated**

### Version Numbering

- **Patch** (1.1.0 → 1.1.1): Bug fixes
- **Minor** (1.1.0 → 1.2.0): New features
- **Major** (1.1.0 → 2.0.0): Breaking changes

### Pre-Release Checklist

- [ ] Version updated in `pyproject.toml`
- [ ] All versions synced with script
- [ ] CHANGELOG.md updated
- [ ] Tests pass locally
- [ ] No sensitive data in commit

---

## References

- PyPI Documentation: https://pypi.org/
- GitHub Actions: https://docs.github.com/actions
- UV Publishing: https://github.com/astral-sh/uv

---

**Version**: 1.0.0
**Author**: DevOps Team
**License**: COPYLEFT-3.0
