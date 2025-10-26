# GitFlow Advisory Policy

**Document ID**: @DOC:GITFLOW-POLICY-001  
**Published**: 2025-10-17  
**Status**: Advisory (recommended, not enforced)  
**Scope**: Personal and Team modes

---

## Overview

MoAI-ADK **recommends** a GitFlow-inspired workflow. This policy shares best practices while letting teams adapt them as needed.

## Key Recommendations

### 1. Main Branch Access (Recommended)

| Recommendation | Summary | Enforcement |
|----------------|---------|-------------|
| **Merge via develop** | Prefer merging `develop` into `main` | Advisory ⚠️ |
| **Feature branches off develop** | Branch from `develop` and raise PRs back to `develop` | Advisory ⚠️ |
| **Release process** | Release flow: `develop` → `main` (release engineer encouraged) | Advisory ⚠️ |
| **Force push** | Warn when force-pushing, but allow it | Warning ⚠️ |
| **Direct push** | Warn on direct pushes to `main`, but allow them | Warning ⚠️ |

### 2. Git Workflow (Recommended)

```
┌─────────────────────────────────────────────────────────┐
│                RECOMMENDED GITFLOW                      │
└─────────────────────────────────────────────────────────┘

        develop (recommended base branch)
          ↑     ↓
    ┌─────────────────┐
    │                 │
    │   developer work │
    │                 │
    ↓                 ↑
feature/SPEC-{ID}   [PR: feature -> develop]
                     [code review + approval]
                     [Merge to develop]

    develop (stable)
         ↓
         │   (release manager prepares)
         ↓
    [PR: develop -> main]
    [CI/CD validation]
    [tag creation]
         ↓
       main (release)
```

**Flexibility**: Direct pushes to `main` are still possible, but the workflow above is preferred.

## Technical Implementation

### Pre-push Hook (Advisory Mode)

**Location**: `.git/hooks/pre-push`  
**Purpose**: Warn on `main` branch pushes without blocking them

```bash
# When attempting to push to main:
⚠️  ADVISORY: Non-standard GitFlow detected

Current branch: feature/SPEC-123
Target branch: main

Recommended GitFlow workflow:
  1. Work on feature/SPEC-{ID} branch (created from develop)
  2. Push to feature/SPEC-{ID} and create PR to develop
  3. Merge into develop after code review
  4. When develop is stable, create PR from develop to main
  5. Release manager merges develop -> main with tag

✓ Push will proceed (flexibility mode enabled)
```

### Force Push Advisory

```bash
⚠️  ADVISORY: Force-push to main branch detected

Recommended approach:
  - Use GitHub PR with proper code review
  - Ensure changes are merged via fast-forward

✓ Push will proceed (flexibility mode enabled)
```

---

## Workflow Examples

### Scenario 1: Standard Feature Development (Recommended)

```bash
# 1. Sync latest code from develop
git checkout develop
git pull origin develop

# 2. Create a feature branch (from develop)
git checkout -b feature/SPEC-001-new-feature

# 3. Implement the change
# ... write code and tests ...

# 4. Commit
git add .
git commit -m "..."

# 5. Push
git push origin feature/SPEC-001-new-feature

# 6. Open a PR: feature/SPEC-001-new-feature -> develop

# 7. Merge into develop after review and approval
```

### Scenario 2: Fast Hotfix (Flexible)

```bash
# When an urgent fix is required:

# Option 1: Recommended (via develop)
git checkout develop
git checkout -b hotfix/critical-bug
# ... apply fix ...
git push origin hotfix/critical-bug
# Open PRs: hotfix -> develop -> main

# Option 2: Direct fix on main (allowed, not recommended)
git checkout main
# ... apply fix ...
git commit -m "Fix critical bug"
git push origin main  # ⚠️ Advisory warning appears but push continues
```

### Scenario 3: Release (Standard or Flexible)

```bash
# Standard approach (recommended):
git checkout develop
gh pr create --base main --head develop --title "Release v1.0.0"

# Direct push (allowed):
git checkout develop
git push origin main  # ⚠️ Advisory warning appears but push continues
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

---

## Policy Modes

### Strict Mode (Legacy, Currently Disabled)

- ❌ Block direct pushes to `main`
- ❌ Block force pushes
- ❌ Block merges into `main` from any branch other than `develop`

### Advisory Mode (Active, v0.3.5+)

- ⚠️ Warn but allow direct pushes to `main`
- ⚠️ Warn but allow force pushes
- ⚠️ Recommend best practices while preserving flexibility
- ✅ Respect user judgment

---

## Recommended Checklist

Every contributor should ensure:

- [ ] `.git/hooks/pre-push` exists and is executable (755)
- [ ] Feature branches fork from `develop`
- [ ] Pull requests target `develop`
- [ ] Releases merge `develop` → `main`

**Verification Commands**:
```bash
ls -la .git/hooks/pre-push
git branch -vv
```

---

## FAQ

**Q: Can we merge into `main` from branches other than `develop`?**  
A: Yes. You will see an advisory warning, but the merge proceeds. The recommended path remains `develop` → `main`.

**Q: Are force pushes allowed?**  
A: Yes. You receive a warning, but the push succeeds. Use with caution.

**Q: Can we commit/push directly to `main`?**  
A: Yes. Expect an advisory warning, yet the push continues.

**Q: Can I disable the hook entirely?**  
A: Yes. Remove `.git/hooks/pre-push` or strip its execute permission.

**Q: Why switch to Advisory Mode?**  
A: To promote best practices while respecting contributor flexibility and judgment.

---

## Policy Change Log

| Date       | Change                                           | Owner        |
|------|------|--------|
| 2025-10-17 | Initial policy drafted (Strict Mode)             | git-manager  |
| 2025-10-17 | Switched to Advisory Mode (warnings only)        | git-manager  |

---

**This policy is advisory—adapt it to fit your project needs.**  
**Reach out to the team lead or release engineer for questions or suggestions.**
