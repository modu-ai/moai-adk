---
name: "Managing GitFlow Branch Strategy"
description: "Enforces MoAI-ADK GitFlow workflow for team and personal modes. Covers branch protection, PR creation rules, release process, and conflict resolution. Essential for /alfred:1-plan (feature branch creation), /alfred:3-sync (PR targeting), and git-manager agent validation. Prevents direct main pushes, enforces develop-based workflow, and ensures code quality."
allowed-tools: "Bash(git:*), Read, Edit"
---

# GitFlow Protection Policy Skill

## Enforced Workflow

```
feature/SPEC-XXX → develop → main
   (your work)   (integration) (release)
       ↓              ↓           ↓
   [commit]      [merge PR]   [auto-deploy]
   [PR to dev]   [code review] [tag release]
```

## Core Rules (Enforced)

### 1. Main Branch Protection
- ❌ **Direct push blocked** – All changes require PR
- ❌ **Force push blocked** – Ensures linear history
- ✅ **PR required** – Code review + approval needed
- ✅ **CI/CD validation** – All checks must pass

### 2. Feature Branch Creation
```bash
# CORRECT: Start from develop
git checkout develop
git pull origin develop
git checkout -b feature/SPEC-{ID}

# WRONG: Starting from main (blocked by rules)
git checkout -b feature/something main  # ❌ Will fail on PR
```

### 3. Pull Request Rules
```
Feature Branch → develop (✅ CORRECT)
Example: feature/SPEC-AUTH-001 → develop

Feature Branch → main (❌ BLOCKED)
Git-manager validates and rejects this pattern
```

### 4. Release Process
```
1. Work on feature/SPEC-{ID}
2. Create PR → develop
3. Code review + merge
4. When develop stable: PR develop → main
5. Release manager merges + creates tag
6. CI/CD auto-deploys package
```

## Common Patterns

| Scenario | Action |
|----------|--------|
| Start feature | Create feature/SPEC-{ID} from develop |
| Push changes | git push origin feature/SPEC-{ID} |
| Create PR | gh pr create --base develop |
| Merge to develop | gh pr merge --squash --delete-branch |
| Release to main | Merge develop → main (when ready) |
| Fix merge conflict | Resolve + rebase, keep clean history |

## Troubleshooting

| Error | Cause | Fix |
|-------|-------|-----|
| "PR to main blocked" | Base branch = main | Close PR, create new with base=develop |
| "Force push rejected" | Attempted force push | Use normal commit + push |
| "Branch out of sync" | Develop changed | git pull origin develop, rebase |
