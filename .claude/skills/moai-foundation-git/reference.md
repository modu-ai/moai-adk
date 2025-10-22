# Git & GitHub CLI Reference

> Git 2.47.0 | GitHub CLI (gh) 2.63.0 | 2025-10-22

---

## Git Best Practices (2024-2025 Standards)

### Commit Message Format

**Standard Structure**:
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

**Rules**:
- Use imperative mood: "Add feature" not "Added feature"
- Subject line ≤50 characters
- Body line length ≤72 characters
- Separate subject from body with blank line
- Explain WHAT and WHY, not HOW

**Examples**:
```bash
✅ feat(auth): add JWT token validation

Implement JWT verification middleware to validate tokens
on protected API routes. Uses HS256 algorithm with 24h expiration.

Refs: @TAG:AUTH-001

❌ Fixed stuff  # Too vague
❌ Added new authentication with JWT tokens and validation middleware plus error handling  # Too long
```

### Branch Strategy (GitFlow)

**Branch Types**:
- `main` / `master`: Production-ready code
- `develop`: Integration branch for next release
- `feature/<name>`: New features
- `bugfix/<name>`: Bug fixes
- `hotfix/<name>`: Critical production fixes
- `release/<version>`: Release preparation

**Naming Conventions**:
```bash
✅ feature/user-authentication
✅ bugfix/login-validation
✅ hotfix/security-patch-v1.2.3

❌ my-branch  # Too generic
❌ feature_user_auth  # Use hyphens, not underscores
```

### Workflow Best Practices

1. **Never commit to main directly** — always use feature branches
2. **Pull before push** — `git pull --rebase origin main`
3. **Commit frequently locally** — push when ready for review
4. **Use atomic commits** — one logical change per commit
5. **Write meaningful messages** — future you will thank you

---

## GitHub CLI (gh) Commands

### Pull Request Management

**Create PR**:
```bash
# Interactive (recommended)
gh pr create

# With options
gh pr create --base main --head feature/auth \
  --title "Add JWT authentication" \
  --body "Implements JWT-based auth with HS256 signing"

# Auto-fill from commits
gh pr create --fill

# Draft PR
gh pr create --draft
```

**List PRs**:
```bash
gh pr list
gh pr list --state open
gh pr list --author "@me"
gh pr list --label "bug"
```

**Review PR**:
```bash
gh pr view 123
gh pr diff 123
gh pr checks 123
gh pr review 123 --approve
gh pr review 123 --request-changes --body "Please fix X"
```

**Merge PR**:
```bash
gh pr merge 123
gh pr merge 123 --squash
gh pr merge 123 --merge
gh pr merge 123 --rebase
gh pr merge 123 --auto  # Auto-merge when checks pass
```

### Issue Management

```bash
gh issue create --title "Bug: Login fails" --body "Description"
gh issue list
gh issue view 456
gh issue close 456
```

### Repository Operations

```bash
gh repo view
gh repo clone owner/repo
gh repo fork
gh repo create my-new-repo --public
```

### GitHub Actions

```bash
gh workflow list
gh workflow view
gh run list
gh run view 789
gh run watch 789  # Watch live
```

---

## MoAI-ADK GitFlow Automation

### TDD Commit Pattern

**RED Phase**:
```bash
git checkout -b feature/auth-validation
git add tests/auth/test_jwt.py
git commit -m "test: add failing test for JWT validation

RED phase: Test expects HS256 token validation.

Refs: @TAG:AUTH-001"
```

**GREEN Phase**:
```bash
git add src/auth/jwt.py
git commit -m "feat: implement JWT token validation

GREEN phase: Validates HS256 tokens with expiration check.
Passes all authentication tests.

Refs: @TAG:AUTH-001"
```

**REFACTOR Phase**:
```bash
git add src/auth/jwt.py
git commit -m "refactor: extract token validation logic

REFACTOR phase: Improves code readability without changing behavior.
Extracts validation into separate functions.

Refs: @TAG:AUTH-001"
```

### Automated PR Workflow

```bash
# 1. Create feature branch
git checkout -b feature/auth-validation

# 2. Implement with TDD commits (RED → GREEN → REFACTOR)

# 3. Push and create Draft PR
git push -u origin feature/auth-validation
gh pr create --draft --title "Add JWT validation" --body "$(cat <<EOF
## Summary
- Implements JWT token validation with HS256
- Adds comprehensive test coverage

## Test Plan
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Coverage ≥85%

Refs: @TAG:AUTH-001
EOF
)"

# 4. Mark PR as Ready when complete
gh pr ready
```

---

## Common Git Operations

### Undoing Changes

```bash
# Undo last commit (keep changes)
git reset --soft HEAD~1

# Undo last commit (discard changes)
git reset --hard HEAD~1

# Revert a commit (create new commit)
git revert <commit-hash>

# Discard local changes
git checkout -- <file>
git restore <file>  # Git 2.23+
```

### Branch Operations

```bash
# List branches
git branch
git branch -r  # Remote branches
git branch -a  # All branches

# Delete branch
git branch -d feature/old-feature  # Safe delete
git branch -D feature/old-feature  # Force delete

# Rename branch
git branch -m old-name new-name
```

### Stashing

```bash
# Stash changes
git stash
git stash save "Work in progress"

# List stashes
git stash list

# Apply stash
git stash apply
git stash pop  # Apply and remove

# Clear all stashes
git stash clear
```

---

## Security Best Practices

### Pre-commit Checks

```bash
# Never commit secrets
# Use .gitignore for:
.env
*.key
credentials.json
secrets/

# Scan for secrets before push
git diff --cached | grep -E "(API_KEY|PASSWORD|SECRET)"
```

### Signed Commits (GPG)

```bash
# Configure GPG signing
git config --global user.signingkey <key-id>
git config --global commit.gpgsign true

# Sign individual commit
git commit -S -m "feat: add feature"

# Verify signatures
git log --show-signature
```

---

## GitHub CLI Configuration

### Aliases

```bash
# ~/.config/gh/config.yml

aliases:
  prc: pr create --fill
  prv: pr view --web
  prl: pr list --state open
  co: pr checkout
```

Usage:
```bash
gh prc  # Create PR with auto-fill
gh prv 123  # Open PR in browser
```

### Authentication

```bash
# Login
gh auth login

# Check status
gh auth status

# Refresh token
gh auth refresh
```

---

## Resources

**Official Documentation**:
- Git: https://git-scm.com/doc
- GitHub CLI: https://cli.github.com/manual/
- Git Book (Pro Git): https://git-scm.com/book/en/v2

**GitFlow Reference**:
- Original blog post: https://nvie.com/posts/a-successful-git-branching-model/
- Atlassian GitFlow tutorial: https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow

**Commit Message Conventions**:
- Conventional Commits: https://www.conventionalcommits.org/
- Angular convention: https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit

---

**Last Updated**: 2025-10-22
**Tools**: Git 2.47.0, GitHub CLI 2.63.0
**Maintained by**: MoAI-ADK Foundation Team
