---
name: moai-foundation-git
version: 4.0.0
updated: 2025-11-20
status: stable
tier: foundation
description: Git workflows, branching strategies, and commit conventions
allowed-tools: [Read, Bash]
---

# Git Workflow Expert

**Version Control Best Practices**

> **Focus**: Branching Strategy, Commit Conventions, PR Workflow  
> **Standards**: Conventional Commits, Semantic Versioning

---

## Overview

Production-grade Git workflows for team collaboration.

### Core Concepts

1.  **Branching Strategy**: Feature branches, trunk-based development
2.  **Commit Messages**: Conventional commits format
3.  **PR Process**: Code review, CI/CD integration

---

## Branching Strategies

### Git Flow (Enterprise)

```
main (production)
  ├── develop (integration)
  │   ├── feature/user-auth
  │   ├── feature/payment
  │   └── release/v1.2.0
  └── hotfix/critical-bug
```

### Trunk-Based Development (Modern)

```
main (always deployable)
  ├── feature/short-lived-branch-1
  └── feature/short-lived-branch-2
```

**Recommendation**: Trunk-based for small teams (<10), Git Flow for larger teams.

---

## Commit Conventions

### Conventional Commits Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation only
- **style**: Formatting (no code change)
- **refactor**: Code restructuring
- **test**: Adding tests
- **chore**: Build/config changes

### Examples

```bash
# Feature
git commit -m "feat(auth): add JWT authentication

Implements OAuth 2.0 flow with refresh tokens.
Closes #123"

# Bug fix
git commit -m "fix(api): handle null response

Prevents crashes when API returns empty data.
Fixes #456"

# Breaking change
git commit -m "feat(api)!: change response format

BREAKING CHANGE: API now returns {data, meta} instead of raw array"
```

---

## PR Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/user-profile
```

### 2. Make Changes & Commit

```bash
git add .
git commit -m "feat(profile): add user profile page"
```

### 3. Push & Create PR

```bash
git push -u origin feature/user-profile
# Create PR on GitHub/GitLab
```

### 4. Code Review

- ✅ CI/CD passes
- ✅ Tests added
- ✅ Documentation updated
- ✅ 1+ approvals

### 5. Merge

```bash
# Squash merge (clean history)
git merge --squash feature/user-profile

# Or rebase (preserve commits)
git rebase main
```

---

## Best Practices

1.  **Small Commits**: One logical change per commit
2.  **Descriptive Messages**: Explain _why_, not just _what_
3.  **Atomic Commits**: Each commit should leave code in working state
4.  **Linear History**: Use rebase to avoid merge commits
5.  **Protected Branches**: Require PR reviews for `main`

---

## Validation Checklist

- [ ] **Branch**: Feature branch created from `main`?
- [ ] **Commits**: Follow conventional format?
- [ ] **Tests**: All tests passing?
- [ ] **Review**: PR approved by team?
- [ ] **CI/CD**: All checks passed?

---

## Related Skills

- `moai-core-rules`: Development workflow
- `moai-devops-ci`: CI/CD integration

---

**Last Updated**: 2025-11-20
