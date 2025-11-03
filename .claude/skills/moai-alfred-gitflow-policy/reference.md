# GitFlow Protection Policy - Reference

## GitHub Branch Protection Configuration

### Main Branch Rules (Enforced)

```
Settings → Branches → Branch protection rules
├── Require a pull request before merging
│   └── Require approvals: 1
├── Require status checks to pass
│   └── Require branches to be up to date before merging
├── Require conversation resolution before merging
├── Require signed commits
└── Restrict who can push to matching branches
    └── Allows: [maintainers only]
```

### Develop Branch Rules (Advisory)

```
Settings → Branches → Branch protection rules
├── Require a pull request before merging
│   └── Require approvals: 0 (flexible)
├── Require status checks to pass
│   └── Both: Linting + Tests
└── Dismiss stale pull request approvals when new commits are pushed
```

## Feature Branch Workflow

### 1. Create Feature Branch

```bash
# Always start from develop
git checkout develop
git pull origin develop
git checkout -b feature/SPEC-AUTH-001

# Naming convention: feature/SPEC-{DOMAIN}-{NUMBER}
# Examples: feature/SPEC-AUTH-001, feature/SPEC-CACHE-FIX-002
```

### 2. Make Commits

```bash
# Follow conventional commit format
git commit -m "feat(auth): Implement email verification

- Add email validation logic
- Create verification token system
- Update user model with verification status

@CODE:AUTH-001"
```

### 3. Push and Create PR

```bash
# Push to remote
git push origin feature/SPEC-AUTH-001

# Create PR (gh cli)
gh pr create \
  --base develop \
  --head feature/SPEC-AUTH-001 \
  --title "[FEAT] Implement Email Verification" \
  --body "$(cat <<'EOF'
## Summary
Implements email verification system as per SPEC-AUTH-001

## Test Plan
- [ ] Unit tests pass (npm run test)
- [ ] Email sending verified in dev
- [ ] Verification flow tested end-to-end

@TEST:AUTH-001
@CODE:AUTH-001
EOF
)"
```

### 4. Code Review and Merge

```bash
# After approval
gh pr merge <PR_NUMBER> \
  --squash \
  --delete-branch \
  -m "Merge: Email verification system (SPEC-AUTH-001)"

# Auto-deletes feature branch after merge
```

## Release Process (develop → main)

### When Develop is Stable

```bash
# 1. Ensure develop is fully tested
git checkout develop
git pull origin develop

# 2. Run complete test suite
npm run test
npm run lint
npm run type-check

# 3. Create release PR to main
gh pr create \
  --base main \
  --head develop \
  --title "Release: v0.2.0" \
  --body "$(cat <<'EOF'
## Release v0.2.0

### Features
- EMAIL: Email verification system (SPEC-AUTH-001)
- CACHE: Redis caching (SPEC-CACHE-FIX-001)

### Bug Fixes
- AUTH: Session expiration logic
- API: Rate limiting edge cases

### Breaking Changes
None

## Checklist
- [x] All tests passing
- [x] Documentation updated
- [x] Changelog generated
- [x] Version bumped
- [x] No security vulnerabilities
EOF
)"

# 4. After approval, merge to main
gh pr merge <RELEASE_PR_NUMBER> \
  --merge \
  --delete-branch \
  -m "Release: v0.2.0 - Email verification and caching"

# 5. Create git tag
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0

# 6. CI/CD auto-deploys to PyPI
```

## Conflict Resolution

### When Develop Changed After Feature Branch Created

```bash
# 1. Fetch latest develop
git fetch origin develop

# 2. Rebase (preferred) or merge
git rebase origin/develop

# 3. If conflicts exist, resolve manually
git status  # See conflicted files

# Edit conflicted files, keep desired changes

# 4. Continue rebase
git add .
git rebase --continue

# 5. Force push (safe, rebasing your own branch)
git push origin feature/SPEC-AUTH-001 --force-with-lease
```

### When PR Has Merge Conflicts

```bash
# If CI shows merge conflicts:
# 1. Update local develop
git checkout develop
git pull origin develop

# 2. Rebase feature branch
git checkout feature/SPEC-AUTH-001
git rebase develop

# 3. Resolve conflicts manually
# (same as above)

# 4. Push resolved version
git push origin feature/SPEC-AUTH-001 --force-with-lease

# GitHub PR automatically updates
```

## Error Handling

### PR to Main Blocked

**Error Message**:
```
❌ GitFlow Violation Detected
Base branch: main (forbidden)
Expected: develop

Feature branches must target develop, not main.
```

**Fix**:
```bash
# 1. Close the incorrect PR
gh pr close <PR_NUMBER>

# 2. Create correct PR
gh pr create --base develop --head feature/SPEC-AUTH-001
```

### Force Push to Main Rejected

**Error Message**:
```
❌ [remote rejected] main (protected branch hook declined)
fatal: Could not read from remote repository.
```

**Fix**:
```bash
# Cannot force push to main (by design)
# Create PR instead
gh pr create --base main --head develop
```

### Branch Out of Sync (Behind Develop)

**Error Message**:
```
⚠️ This branch is out-of-date with the base branch
Consider updating your branch before merging.
```

**Fix**:
```bash
# Update from latest develop
git fetch origin
git rebase origin/develop
git push origin feature/SPEC-AUTH-001 --force-with-lease
```

## Checklist: Before Merging to Develop

- [ ] Feature branch created from latest develop
- [ ] All commits follow conventional commit format
- [ ] Code passes linting and type checking
- [ ] Tests pass locally and in CI
- [ ] PR description includes @SPEC:ID and @TEST:ID
- [ ] Code review approved by at least 1 maintainer
- [ ] All conversations resolved
- [ ] Branch is up-to-date with develop
- [ ] Commits are squashed (optional but recommended)

## Checklist: Before Merging to Main

- [ ] All features merged and tested in develop
- [ ] Complete test suite passes
- [ ] Documentation is current
- [ ] Changelog is generated
- [ ] Version is bumped (semver)
- [ ] Security review completed
- [ ] Release notes drafted
- [ ] No regressions detected
