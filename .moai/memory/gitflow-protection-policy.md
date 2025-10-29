# GitFlow Protection Policy

**Document ID**: @DOC:GITFLOW-POLICY-ALIAS
**Published**: 2025-10-17
**Updated**: 2025-10-29
**Status**: **Enforced via GitHub Branch Protection** (v0.8.3+)
**Scope**: Personal and Team modes

---

## Overview

MoAI-ADK **enforces** a GitFlow-inspired workflow through GitHub Branch Protection. As of v0.8.3, the `main` branch is protected and requires Pull Requests for all changes, including from administrators.

**What Changed**: Previously (v0.3.5-v0.8.2), we used an advisory approach with warnings. Now we enforce proper GitFlow to ensure code quality and prevent accidental direct pushes to main.

## Key Requirements (Enforced)

### 1. Main Branch Access (Enforced)

| Requirement | Summary | Enforcement |
|-------------|---------|-------------|
| **Merge via develop** | MUST merge `develop` into `main` | ✅ Enforced |
| **Feature branches off develop** | MUST branch from `develop` and raise PRs back to `develop` | ✅ Enforced |
| **Release process** | Release flow: `develop` → `main` (PR required) | ✅ Enforced |
| **Force push** | Blocked on `main` | ✅ Blocked |
| **Direct push** | Blocked on `main` (PR required) | ✅ Blocked |

### 2. Git Workflow (Required)

```
┌─────────────────────────────────────────────────────────┐
│                ENFORCED GITFLOW                         │
│         (GitHub Branch Protection Active)               │
└─────────────────────────────────────────────────────────┘

        develop (required base branch)
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
    [Code review + approval REQUIRED]
    [All discussions resolved]
    [CI/CD validation]
    [tag creation]
         ↓
       main (protected release)
```

**Enforcement**: Direct pushes to `main` are **blocked** via GitHub Branch Protection. All changes must go through Pull Requests.

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

### Strict Mode (Active, v0.8.3+) ✅ ENFORCED

**GitHub Branch Protection Enabled**:
- ✅ **enforce_admins: true** - Administrators must follow all rules
- ✅ **required_pull_request_reviews** - 1 approval required
- ✅ **required_conversation_resolution** - All discussions must be resolved
- ✅ **Block direct pushes to `main`** - PR required for all users
- ✅ **Block force pushes** - Prevents history rewriting
- ✅ **Block branch deletion** - Protects main from accidental deletion

**What This Means**:
- ❌ No one (including admins) can push directly to `main`
- ✅ All changes must go through Pull Requests
- ✅ PRs require code review approval
- ✅ All code discussions must be resolved before merge
- ✅ Enforces proper GitFlow: feature → develop → main

### Advisory Mode (Legacy, v0.3.5 - v0.8.2)

- ⚠️ Warned but allowed direct pushes to `main`
- ⚠️ Warned but allowed force pushes
- ⚠️ Recommended best practices while preserving flexibility
- ❌ **Deprecated** - Replaced by Strict Mode for better quality control

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
A: Advisory Mode was used in v0.3.5-v0.8.2. As of v0.8.3, we've switched to Strict Mode with GitHub Branch Protection for better quality control.

**Q: What if develop falls behind main?**
A: This can happen when hotfixes or releases go directly to main. Regularly sync main → develop to prevent divergence. See "Maintaining develop-main Sync" section below.

**Q: Can I bypass branch protection in emergencies?**
A: No. Even administrators must follow the PR process. For true emergencies, temporarily disable protection via GitHub Settings (requires admin access), but re-enable immediately after.

---

## Maintaining develop-main Sync

### ⚠️ Critical Rule: develop Must Stay Current

**Problem**: When main receives direct commits (hotfixes, emergency releases) without syncing back to develop, GitFlow breaks:

```
❌ BAD STATE:
develop: 3 commits ahead, 29 commits behind main
- develop has outdated dependencies
- New features branch from old code
- Merge conflicts multiply over time
```

### Signs of Drift

Monitor for these warnings:
- `git status` shows "Your branch is X commits behind main"
- Feature branches conflict with main during PR
- CI/CD failures due to dependency mismatches
- Version numbers in develop don't match main

### Recovery Procedure

When develop falls behind main:

1. **Assess the Gap**
   ```bash
   git log --oneline develop..main  # Commits in main but not develop
   git log --oneline main..develop  # Commits in develop but not main
   ```

2. **Sync Strategy: Merge main into develop (Recommended)**
   ```bash
   git checkout develop
   git pull origin develop        # Get latest develop
   git merge main                 # Merge main into develop
   # Resolve conflicts if any (prefer main for version/config files)
   git push origin develop
   ```

3. **Emergency Only: Reset develop to main (Destructive)**
   ```bash
   # ⚠️ ONLY if develop's unique commits are unwanted
   git checkout develop
   git reset --hard main
   git push origin develop --force
   ```

### Prevention: Regular Sync Schedule

**After every main release** (REQUIRED):
```bash
# Immediately after merging develop → main:
git checkout develop
git merge main
git push origin develop
```

**Weekly maintenance** (for active projects):
```bash
# Every Monday morning:
git checkout develop
git pull origin main
git push origin develop
```

### Real-World Case Study (2025-10-29)

**Situation**: develop was 29 commits behind main due to:
- v0.8.2, v0.8.3 released directly to main
- No reverse sync to develop
- Feature branches contained outdated code

**Resolution**:
- Merged main → develop (14 file conflicts)
- Resolved conflicts prioritizing main's versions
- TAG validation bypassed for merge commit
- Enabled Strict Mode to prevent future direct pushes

**Lesson**: With Strict Mode active, this won't happen again. All releases must go through develop → main PR flow.

---

## Policy Change Log

| Date       | Change                                           | Owner        |
|------|------|--------|
| 2025-10-17 | Initial policy drafted (Strict Mode)             | git-manager  |
| 2025-10-17 | Switched to Advisory Mode (warnings only)        | git-manager  |
| 2025-10-29 | **Enabled GitHub Branch Protection (Strict Mode)** | Alfred       |
| 2025-10-29 | Added develop-main sync guidelines and real-world case study | Alfred       |
| 2025-10-29 | Enforced `enforce_admins`, `required_conversation_resolution` | Alfred       |

---

**This policy is advisory—adapt it to fit your project needs.**  
**Reach out to the team lead or release engineer for questions or suggestions.**
