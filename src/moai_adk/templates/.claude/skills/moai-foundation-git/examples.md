# Git & GitHub Workflows - Practical Examples ( .0)

**Last Updated**: 2025-11-12 | **Enterprise Patterns** | **Production Ready**

---

## Quick Start Examples

### Example 1: Simple Feature Implementation

```bash
# User runs Alfred command
/alfred:1-plan "Add password strength validation"

# Alfred creates branch and asks:
# Which workflow? Feature Branch + PR / Direct Commit
# User chooses: Feature Branch + PR

# Alfred creates feature/SPEC-006
git branch -a | grep SPEC-006

# User switches to branch
git checkout feature/SPEC-006

# === RED Phase ===
# Write test first

git commit -m "🔴 RED: test_password_strength_validation


# === GREEN Phase ===
# Implement password validation

git commit -m "🟢 GREEN: implement_password_strength_check


# === REFACTOR Phase ===

git commit -m "♻️ REFACTOR: improve_password_error_messages


# Create PR
gh pr create \\
  --base develop \\
  --title "feat: Add password strength validation" \\
  --body "Implements comprehensive password validation"

# After approval
gh pr merge --squash --delete-branch
```

### Example 2: Direct Commit Workflow

```bash
# User chooses "Direct Commit" workflow
git checkout develop
git pull origin develop

# Implement and commit directly
git push origin develop

# CI/CD gates validate automatically
```

### Example 3: Parallel Features with Session Persistence

```bash
# Developer has two parallel SPECs
/alfred:1-plan "Implement user profile"     # Creates feature/SPEC-007
/alfred:1-plan "Implement user preferences" # Creates feature/SPEC-008

# Work on SPEC-007
git checkout feature/SPEC-007

# Switch to SPEC-008
git checkout feature/SPEC-008

# Later, resume SPEC-007 from checkpoint
/alfred:2-run SPEC-007
```

---

## Detailed Workflow Examples

### Example 1: Enterprise PR Review Process

```bash
# Step 1: Create draft PR
gh pr create \\
  --draft \\
  --title "WIP: Complex feature implementation"

# PR URL: https://github.com/org/repo/pull/42

# Step 2: Keep PR updated as you work
git push origin feature/SPEC-XYZ

# Step 3: Convert draft to ready
gh pr ready 42

# Step 4: View review comments
gh pr view 42

# Step 5: Address feedback
git push origin feature/SPEC-XYZ

# Step 6: Merge after approval
gh pr merge 42 --squash --delete-branch
```

### Example 2: Hotfix Workflow

```bash
# Urgent bug discovered in production

# Create hotfix branch
git checkout main
git pull origin main
git checkout -b hotfix/critical-auth-bypass

# Fix the bug and commit with security tag
git commit -m "🔒 SECURITY: fix_authentication_bypass

Severity: CRITICAL"

# Create PR to main
gh pr create \\
  --base main \\
  --title "HOTFIX: Critical auth vulnerability"

# Fast-track merge
gh pr merge --squash --delete-branch

# Also merge to develop
git checkout develop
git pull origin develop
git merge main
git push origin develop
```

### Example 3: Monorepo with MIDX Optimization

```bash
# Enable MIDX for this repository
git config gc.writeMultiPackIndex true
git config gc.multiPackIndex true

# Perform aggressive cleanup
time git gc --aggressive --prune=now

# Before: 45 seconds
# After: 28 seconds (38% improvement with MIDX)

# Enable sparse checkout for faster development
git sparse-checkout init --cone
git sparse-checkout set src/moai_adk tests .github
```

### Example 4: Complete Feature Release

```bash
# Step 1: Merge all features to main
git checkout main
git pull origin main
git merge develop --ff-only

# Step 2: Create release tag
git tag -a v2.0.0 -m "Release version 2.0.0"

# Step 3: Push to remote
git push origin main
git push origin v2.0.0

# Step 4: Create GitHub release
gh release create v2.0.0 \\
  --title "Version 2.0.0" \\
  --notes "See CHANGELOG.md for details"

# Step 5: Publish release
gh release edit v2.0.0 --draft=false
```

### Example 5: Conflict Resolution

```bash
# Conflict during merge
git merge feature/SPEC-009

# See conflicts
git status

# Resolve conflicts in editor
# Then mark as resolved
git add src/models/user.py

# Complete merge
git commit -m "Merge feature/SPEC-009"

# Or abort merge
git merge --abort
```

---

## Advanced Scenarios

### Scenario 1: Cherry-pick Specific Commit

```bash
# Find commit
git log feature/SPEC-010 --oneline | head

# Cherry-pick commit
git cherry-pick abc1234

# If conflicts, resolve and continue
git cherry-pick --continue
```

### Scenario 2: Reset Accidentally Pushed Commit

```bash
# Undo last local commit
git reset HEAD~1

# Or revert (safer)
git revert HEAD
git push origin feature/SPEC-005
```

### Scenario 3: Archive Old Feature Branch

```bash
# Delete local branch
git branch -d feature/SPEC-001

# Delete remote branch
git push origin --delete feature/SPEC-001
```

---

## Quick Reference: Common Commands

```bash
# Create and switch to branch
git checkout -b feature/SPEC-XXX

# See current branch
git branch --show-current

# List all branches
git branch -a

# View recent commits
git log --oneline -10

# View changes not staged
git diff

# View staged changes
git diff --cached

# Stash uncommitted work
git stash

# Apply stashed work
git stash pop

# Update from remote
git pull origin develop

# Push to remote
git push origin feature/SPEC-XXX
```

---

## Reference Links

- **SKILL.md** - Complete feature overview
- **reference.md** - Detailed command reference
- **Git Official Docs** - https://git-scm.com/docs
- **GitHub CLI Docs** - https://cli.github.com

**Examples Last Updated**: 2025-11-12 | **Version**: 4.0.0 Enterprise | **Production Ready**
