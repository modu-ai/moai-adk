# Git Workflow Reference Documentation

_Last updated: 2025-10-22_

## GitFlow Branching Model

### Branch Types

| Branch | Purpose | Lifetime | Naming |
|--------|---------|----------|--------|
| main | Production code | Permanent | `main` |
| develop | Integration branch | Permanent | `develop` (optional) |
| feature | New features | Temporary | `feature/<name>` |
| hotfix | Production fixes | Temporary | `hotfix/<name>` |
| release | Release preparation | Temporary | `release/<version>` |

---

## TDD Commit Pattern

### RED-GREEN-REFACTOR

**RED**: Write failing test
```bash
git commit -m "test: add failing test for <feature>

- Describe what the test expects
- Reference SPEC if applicable

@TEST:<TAG-ID>"
```

**GREEN**: Make test pass
```bash
git commit -m "feat: implement <feature> to pass tests

- Minimal implementation to pass test
- Include TAG references

@CODE:<TAG-ID> | TEST: <test-file>"
```

**REFACTOR**: Improve code quality
```bash
git commit -m "refactor: clean up <component> without changing behavior

- Extract methods
- Improve naming
- Add documentation

@CODE:<TAG-ID>"
```

---

## Commit Message Format

### Structure
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `test`: Test changes
- `refactor`: Code restructuring
- `docs`: Documentation
- `chore`: Maintenance
- `perf`: Performance improvement
- `ci`: CI/CD changes

### Example
```
feat(auth): implement JWT token generation

- Add generate_token() function
- Set expiration to 24 hours
- Include user ID and email in payload

@CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md
```

---

## GitHub CLI (gh) Commands

### Pull Request Management
```bash
# Create draft PR
gh pr create --draft --title "Title" --body "Description"

# Mark as ready
gh pr ready

# Merge PR
gh pr merge --squash  # Squash and merge
gh pr merge --merge   # Create merge commit
gh pr merge --rebase  # Rebase and merge

# PR status
gh pr status
gh pr view 123
gh pr checks
```

### Issue Management
```bash
# Create issue
gh issue create --title "Bug" --body "Description"

# Link PR to issue
gh pr create --body "Fixes #123"

# Close issue
gh issue close 123
```

---

## Git Tool Versions (2025-10-22)

| Tool | Version | Status |
|------|---------|--------|
| Git | 2.47.0 | ✅ Current |
| GitHub CLI | 2.63.0 | ✅ Current |

### Installation
```bash
# macOS
brew install git gh

# Linux (Debian/Ubuntu)
sudo apt install git gh

# Windows
winget install Git.Git GitHub.cli
```

---

## Automation Scripts

### Pre-commit Hook (Quality Checks)
```bash
#!/bin/bash
# .git/hooks/pre-commit

set -e

echo "Running pre-commit checks..."

# Linting
ruff check . || exit 1

# Type checking
mypy src/ || exit 1

# Tests
pytest --quiet || exit 1

echo "✓ All checks passed"
```

### Post-merge Hook (Cleanup)
```bash
#!/bin/bash
# .git/hooks/post-merge

# Delete merged branches
git branch --merged | grep -v "\*" | grep -v "main" | xargs -r git branch -d

echo "✓ Cleaned up merged branches"
```

---

## CI/CD Integration

### GitHub Actions Workflow
```yaml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Tests
        run: pytest --cov=src --cov-report=xml

      - name: Quality Checks
        run: |
          ruff check .
          mypy src/

      - name: Security Scan
        run: semgrep scan --config=auto
```

---

## References

- [GitFlow Workflow - Atlassian](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow)
- [GitHub CLI Manual](https://cli.github.com/manual/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Git Documentation](https://git-scm.com/doc)

---

_For practical workflows, see examples.md_
