---
name: moai-ref-git-workflow
description: >
  Git workflow patterns, branch strategies, conventional commits, and PR templates
  reference for git operations. Agent-extending skill that amplifies manager-git
  expertise with production-grade git workflow patterns.
  NOT for: code implementation, testing, architecture design, documentation content.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-03-30"
  tags: "git, branch, commit, pr, workflow, reference"
  agent: "manager-git"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000

# MoAI Extension: Triggers
triggers:
  keywords: ["git", "commit", "branch", "pr", "merge", "rebase"]
  agents: ["manager-git"]
  phases: ["run", "sync"]
---

# Git Workflow Reference

## Target Agent

`manager-git` - Applies these patterns directly to git operations, branch management, and PR creation.

## Branch Strategy Patterns

### GitHub Flow (Default for Most Projects)

```
main ─────────────────────────────────────────
  └── feat/SPEC-XXX-description ──── PR ──→ merge
```

Rules:
- `main` is always deployable
- Feature branches from `main`
- PR required for all merges
- Delete branch after merge

### GitFlow (Complex Release Cycles)

```
main ──────────────────────────────────────────
  └── develop ─────────────────────────────────
        ├── feature/SPEC-XXX ──── PR ──→ develop
        └── release/v1.2.0 ────── PR ──→ main + develop
```

### Trunk-Based (CI/CD Heavy)

```
main ──────────────────────────────────────────
  └── short-lived branch (< 1 day) ──→ merge
```

## Branch Naming Convention

| Pattern | Example | Use Case |
|---------|---------|----------|
| `feat/SPEC-{ID}-{slug}` | `feat/SPEC-AUTH-001-jwt-auth` | New feature |
| `fix/SPEC-{ID}-{slug}` | `fix/SPEC-BUG-042-null-check` | Bug fix |
| `refactor/{slug}` | `refactor/extract-auth-middleware` | Refactoring |
| `docs/{slug}` | `docs/api-reference-update` | Documentation |
| `chore/{slug}` | `chore/upgrade-dependencies` | Maintenance |

## Conventional Commits Reference

| Type | When | Example |
|------|------|---------|
| `feat` | New feature | `feat(auth): add JWT refresh token flow` |
| `fix` | Bug fix | `fix(api): handle null user in profile endpoint` |
| `refactor` | Code restructure | `refactor(db): extract query builder` |
| `test` | Test changes | `test(auth): add login edge case tests` |
| `docs` | Documentation | `docs(api): update endpoint descriptions` |
| `chore` | Maintenance | `chore(deps): upgrade Go to 1.23` |
| `perf` | Performance | `perf(query): add index for user lookup` |
| `style` | Formatting | `style: apply gofmt formatting` |
| `ci` | CI/CD changes | `ci: add GitHub Actions workflow` |
| `revert` | Revert commit | `revert: undo feat(auth) commit abc123` |

### Commit Message Structure

```
<type>(<scope>): <description>    # max 72 chars

[optional body]                    # what and why, not how

[optional footer]                  # Breaking changes, issue refs
BREAKING CHANGE: <description>
Refs: #123, SPEC-AUTH-001
```

## Pull Request Template

```markdown
## Summary
- [1-3 bullet points describing what this PR does]

## Changes
- [ ] File 1: description of change
- [ ] File 2: description of change

## Test Plan
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manual testing completed

## SPEC Reference
- SPEC-{ID}: {title}

## Checklist
- [ ] Tests pass (`go test ./...`)
- [ ] Linting pass (`golangci-lint run`)
- [ ] No secrets committed
- [ ] Documentation updated if needed
```

## Merge Strategy Selection

| Strategy | When | Command |
|----------|------|---------|
| Squash merge | Feature branches (clean history) | `gh pr merge --squash` |
| Merge commit | Release branches (preserve history) | `gh pr merge --merge` |
| Rebase | Small, clean commits | `gh pr merge --rebase` |

## Git Safety Rules

| Action | Risk | Rule |
|--------|------|------|
| `git push --force` | Overwrites remote | NEVER on main/master, ask user first |
| `git reset --hard` | Loses local changes | Confirm with user first |
| `git checkout .` | Discards changes | Confirm with user first |
| `git branch -D` | Deletes branch | Only after merge confirmed |
| `--no-verify` | Skips hooks | NEVER unless user explicitly requests |
| `git rebase -i` | Interactive (not supported) | NEVER use (requires interactive input) |

## Context Memory in Commits

Embed decision context in commit messages for future session continuity:

```
feat(auth): implement JWT refresh token rotation

Decision: Chose rotation over sliding window for security
Pattern: Middleware chain: RateLimit -> Auth -> Authz -> Handler
Gotcha: Token blacklist requires Redis, not just in-memory cache

Refs: SPEC-AUTH-001
```

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I will clean up the commit messages before merging" | Interactive rebase is error-prone under pressure. Write clean commits from the start. |
| "Force push is fine on my feature branch" | Collaborators or CI may have fetched the branch. Force push destroys their reference. Use --force-with-lease. |
| "This commit is too small to need a conventional format" | Changelog generators, bisect, and blame all depend on consistent commit formats. Every commit matters. |
| "I will push directly to main, it is a small fix" | Direct pushes bypass code review and CI. Even small fixes can break production. |
| "Merge commits are messy, I always squash" | Squash loses individual commit context. Merge commits preserve the development narrative for future debugging. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Commit message does not follow conventional format (type(scope): description)
- Force push to main or shared release branch
- PR merged without CI passing
- Branch name does not indicate feature, fix, or SPEC reference
- Merge conflict markers found in committed files

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] All commit messages follow conventional format (show git log --oneline)
- [ ] Branch name follows convention (feature/, fix/, chore/ prefix)
- [ ] No force pushes to main or protected branches (check reflog or CI)
- [ ] PR has passing CI checks before merge
- [ ] No merge conflict markers in committed files (grep for <<<<<<<)
- [ ] SPEC-ID referenced in commit message or PR description when applicable

<!-- moai:evolvable-end -->
