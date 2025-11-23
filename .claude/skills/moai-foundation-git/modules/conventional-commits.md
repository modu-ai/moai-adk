# Conventional Commits 2025

Complete guide to standardized commit messages for automation, documentation, changelog generation, and semantic versioning.

---

## Core Format Specification

**Standard Structure**:
```
<type>(<scope>): <subject>

[optional body]

[optional footer(s)]
```

**Constraints**:
- `<type>`: Lowercase, no spaces (required)
- `<scope>`: Alphanumeric + hyphens, no spaces (optional but recommended)
- `<subject>`: Present tense, lowercase start, no period (required)
- `[body]`: Detailed explanation, WHY not WHAT (optional)
- `[footer]`: Breaking changes, issue references (optional)

---

## Commit Types (8 Standard Categories)

### 1. feat - New Features

**Purpose**: Add new functionality visible to users

**Examples**:
```bash
# Simple feature
git commit -m "feat(auth): add JWT token validation"

# Feature with scope
git commit -m "feat(api): implement user registration endpoint"

# Feature with body
git commit -m "feat(payment): integrate Stripe payment gateway

Implement complete Stripe integration:
- Payment intent creation
- Webhook handling for payment events
- Refund processing
- Subscription management

Tested with Stripe test mode API"
```

### 2. fix - Bug Fixes

**Purpose**: Patch errors in existing code

**Examples**:
```bash
# Critical bug fix
git commit -m "fix(api): prevent null pointer in user lookup

Add null check before accessing user object.
Prevents server crash when user not found.

Fixes #342"

# Security fix
git commit -m "fix(auth): prevent JWT token reuse after logout

Clear token from cache immediately on logout.
Prevents session hijacking vulnerability.

Security issue reported by @security-team"

# Performance fix
git commit -m "fix(db): optimize slow user query

Add index on users.email column.
Reduces query time from 2s to 50ms.

Related: PERF-015"
```

### 3. docs - Documentation Changes

**Purpose**: Update documentation only (no code changes)

**Examples**:
```bash
# README update
git commit -m "docs(readme): add installation instructions for Windows"

# API documentation
git commit -m "docs(api): document authentication endpoints

Add comprehensive API documentation:
- Request/response examples
- Authentication requirements
- Error codes and handling"

# Code comments
git commit -m "docs(auth): add docstrings to JWT validator"
```

### 4. style - Code Formatting

**Purpose**: Changes that don't affect code meaning (whitespace, formatting, missing semicolons)

**Examples**:
```bash
# Formatting only
git commit -m "style(src): format code with black

Run black on entire src/ directory.
No logic changes."

# Linting fixes
git commit -m "style(api): fix linting errors

- Remove unused imports
- Fix line length violations
- Organize imports per PEP8"
```

### 5. refactor - Code Restructuring

**Purpose**: Code changes that neither fix bugs nor add features

**Examples**:
```bash
# Extract method
git commit -m "refactor(auth): extract token validation logic

Extract validate_token() from login() method.
Improves testability and code reuse.
No behavior changes."

# Rename for clarity
git commit -m "refactor(db): rename get_user to get_user_by_id

Clarify function purpose.
All call sites updated.
Tests still pass."

# Improve architecture
git commit -m "refactor(api): separate authentication concerns

Move auth logic from handlers to auth service.
Implements single responsibility principle.
All tests updated and passing."
```

### 6. perf - Performance Improvements

**Purpose**: Code changes that improve performance

**Examples**:
```bash
# Database optimization
git commit -m "perf(db): add index on frequently queried columns

Add composite index on (user_id, created_at).
Reduces query time by 85% (2s → 300ms).

Benchmarked with 1M records."

# Caching
git commit -m "perf(api): add Redis caching for user lookups

Cache user data with 5-minute TTL.
Reduces database load by 70%.

Measured with load testing."

# Algorithm improvement
git commit -m "perf(search): optimize search algorithm

Replace O(n²) with O(n log n) implementation.
Improves search time for 10K items: 5s → 0.5s."
```

### 7. test - Add/Update Tests

**Purpose**: Add missing tests or correct existing tests

**Examples**:
```bash
# Add missing tests
git commit -m "test(auth): add tests for token expiration

Add comprehensive expiration tests:
- Valid token acceptance
- Expired token rejection
- Edge case: token expires during request

Coverage: 87% → 95%"

# Fix flaky tests
git commit -m "test(api): fix flaky integration test

Replace time.sleep with proper async wait.
Test now runs reliably in CI."

# Update test data
git commit -m "test(db): update test fixtures for new schema"
```

### 8. chore - Build/Tooling Updates

**Purpose**: Changes to build process, dependencies, or development tools

**Examples**:
```bash
# Dependency update
git commit -m "chore(deps): upgrade pytest to 8.2.0

Update pytest and related plugins.
All tests still pass."

# CI/CD configuration
git commit -m "chore(ci): add automated security scanning

Add Bandit security scanner to CI pipeline.
Runs on every PR."

# Development tools
git commit -m "chore(dev): add pre-commit hooks

Configure black, ruff, and mypy hooks.
Ensures code quality before commit."
```

---

## Breaking Changes

**Format**:
```bash
# Breaking change with ! suffix
git commit -m "feat(api)!: change authentication endpoint

BREAKING CHANGE: /login moved to /auth/login

Migrate all clients to new endpoint.
Old endpoint will be removed in v3.0.0.

Migration guide: docs/migration-v2-to-v3.md"
```

**Footer Format**:
```
BREAKING CHANGE: <description>

<migration instructions>
<deprecation timeline>
```

**Real-World Examples**:
```bash
# API contract change
git commit -m "feat(api)!: require email verification for registration

BREAKING CHANGE: Registration endpoint now requires verified email

Users must verify email before account activation.
Update client apps to handle verification flow.

Deprecation: Unverified registrations rejected after 2025-12-01"

# Configuration change
git commit -m "chore(config)!: move config from .env to config.json

BREAKING CHANGE: Environment variables no longer supported

Migrate to config.json format.
Migration script: scripts/migrate_env_to_json.py

Old format support ends: 2025-11-30"
```

---

## Multiple Scopes

**Format**: Use comma-separated scopes when change affects multiple modules

**Examples**:
```bash
# Multiple related modules
git commit -m "feat(auth,api): implement OAuth2 authentication

Add OAuth2 support to authentication and API layers:
- OAuth2 provider integration
- API endpoints for OAuth flow
- Token exchange implementation"

# Cross-cutting concern
git commit -m "refactor(auth,db,api): standardize error handling

Implement consistent error handling across layers:
- Custom exception hierarchy
- Error logging with context
- User-friendly error messages"
```

---

## Commit Message Best Practices

### ✅ DO

**Use imperative mood**:
```bash
✅ "fix(api): prevent null pointer"
❌ "fixed null pointer"
❌ "fixes null pointer"
❌ "fixing null pointer"
```

**Be specific about scope**:
```bash
✅ "feat(auth): add JWT validation"
❌ "feat: add feature"
```

**Explain WHY in body**:
```bash
✅ "refactor(db): extract connection pooling

Separate connection pool logic for better testability
and to support multiple database backends."

❌ "refactor(db): change code"
```

**Reference issues**:
```bash
✅ "fix(api): handle timeout errors

Fixes #342
Related: #338, #340"
```

**Keep subject under 50 characters**:
```bash
✅ "feat(auth): add 2FA support"
❌ "feat(auth): add two-factor authentication support with SMS and authenticator app integration"
```

### ❌ DON'T

**Don't mix types**:
```bash
❌ "feat,fix(api): add endpoint and fix bug"
✅ Two separate commits:
   "feat(api): add user endpoint"
   "fix(api): handle null user"
```

**Don't use vague subjects**:
```bash
❌ "fix: fix bug"
❌ "chore: update stuff"
❌ "refactor: changes"
```

**Don't commit unrelated changes**:
```bash
❌ Single commit with:
   - New feature
   - Bug fix
   - Documentation update

✅ Three separate commits
```

---

## Automation Integration

### Pre-commit Hook Validation

**Enforce format**:
```bash
#!/bin/bash
# .git/hooks/commit-msg

commit_msg_file="$1"
commit_msg=$(cat "$commit_msg_file")

# Regex pattern for conventional commits
pattern="^(feat|fix|docs|style|refactor|perf|test|chore)(\([a-z0-9\-]+\))?: .+"

if ! echo "$commit_msg" | grep -qE "$pattern"; then
  echo "ERROR: Commit message must follow Conventional Commits format"
  echo ""
  echo "Format: type(scope): subject"
  echo ""
  echo "Types: feat, fix, docs, style, refactor, perf, test, chore"
  echo "Example: feat(auth): add JWT validation"
  exit 1
fi

# Check subject line length
subject_line=$(echo "$commit_msg" | head -n 1)
if [ ${#subject_line} -gt 72 ]; then
  echo "ERROR: Subject line must be ≤72 characters (current: ${#subject_line})"
  exit 1
fi

echo "✅ Commit message format valid"
```

### CI/CD Validation

**GitHub Actions**:
```yaml
# .github/workflows/commit-validation.yml
name: Validate Commit Messages

on: [pull_request]

jobs:
  validate-commits:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Validate commit messages
        run: |
          commits=$(git log origin/main..HEAD --pretty=format:"%s")

          while IFS= read -r commit; do
            if ! echo "$commit" | grep -qE "^(feat|fix|docs|style|refactor|perf|test|chore)(\([a-z0-9\-]+\))?: .+"; then
              echo "❌ Invalid commit: $commit"
              exit 1
            fi
          done <<< "$commits"

          echo "✅ All commits follow Conventional Commits format"
```

### Automatic Changelog Generation

**Using commitizen**:
```bash
# Install commitizen
pip install commitizen

# Generate changelog
cz changelog

# Output: CHANGELOG.md
## v2.1.0 (2025-11-23)

### Features
- **auth**: add JWT validation (abc123)
- **api**: implement user registration (def456)

### Bug Fixes
- **db**: prevent connection leak (ghi789)
- **api**: handle null user gracefully (jkl012)

### Performance Improvements
- **search**: optimize search algorithm (mno345)
```

---

## Real-World TDD Workflow Example

**Complete feature implementation**:
```bash
# Phase 1: RED - Write failing test
git add tests/test_user_registration.py
git commit -m "test(api): add failing user registration test

Add comprehensive registration tests:
- Valid user registration
- Duplicate email rejection
- Invalid email format handling
- Password strength validation

All tests currently fail (no implementation yet)
Related: SPEC-042"

# Phase 2: GREEN - Minimal implementation
git add src/api/registration.py src/db/models.py
git commit -m "feat(api): implement user registration endpoint

Implement basic registration:
- POST /api/register endpoint
- Email uniqueness validation
- Password hashing with bcrypt
- User creation in database

All tests now pass
Coverage: 89% (target: 85%)
Related: SPEC-042"

# Phase 3: REFACTOR - Code quality
git add src/api/registration.py src/api/validators.py
git commit -m "refactor(api): extract validation logic

Extract email and password validators:
- Separate EmailValidator class
- Separate PasswordValidator class
- Improve error messages

No behavior changes (all tests still pass)
Related: SPEC-042"

# Phase 4: Documentation
git add docs/api/registration.md
git commit -m "docs(api): document registration endpoint

Add comprehensive API documentation:
- Request/response examples
- Error codes and handling
- Rate limiting information
- Security considerations

Related: SPEC-042"
```

---

## Semantic Versioning Integration

**Version Bump Rules**:
```
feat commit → MINOR version bump (1.2.0 → 1.3.0)
fix commit → PATCH version bump (1.2.0 → 1.2.1)
BREAKING CHANGE → MAJOR version bump (1.2.0 → 2.0.0)
```

**Automated versioning**:
```bash
# Using semantic-release
npm install -g semantic-release

# Analyze commits and bump version
semantic-release

# Output:
# - Analyzes commits since last tag
# - Determines version bump
# - Updates package.json
# - Creates git tag
# - Generates changelog
# - Publishes release
```

---

## Troubleshooting Common Issues

**Issue: Typo in commit type**:
```bash
# Wrong type
git commit -m "feat(api): add endpoint"
# Typo: "featt" instead of "feat"

# Fix with amend (if not pushed)
git commit --amend -m "feat(api): add endpoint"
```

**Issue: Forgot scope**:
```bash
# No scope
git commit -m "feat: add validation"

# Fix with amend
git commit --amend -m "feat(auth): add JWT validation"
```

**Issue: Mixed changes in one commit**:
```bash
# Split commit
git reset HEAD~1
git add src/auth.py
git commit -m "feat(auth): add JWT validation"
git add src/api.py
git commit -m "fix(api): handle timeout"
```

---

**Last Updated**: 2025-11-23
**Lines**: 278
