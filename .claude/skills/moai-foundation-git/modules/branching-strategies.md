# Git Branching Strategies

Comprehensive guide to selecting and implementing the optimal branching workflow based on team size, project complexity, and collaboration requirements.

---

## Strategic Decision Framework

Choose branching strategy based on systematic evaluation:

| Factor | Feature Branch | Direct Commit | Per-SPEC Choice |
|--------|---------------|---------------|-----------------|
| **Team Size** | 2+ developers | 1 developer | 1-3 developers |
| **Code Review** | Required | Optional | Conditional |
| **Release Cadence** | Scheduled | Continuous | Variable |
| **Risk Tolerance** | Low-Medium | High | Medium |
| **CI/CD Integration** | Full pipeline | Minimal | Selective |

**Decision Tree**:
```
Is this a team project?
├─ YES → Feature Branch Workflow (Strategy 1)
└─ NO → Is change high-risk?
    ├─ YES → Feature Branch for review
    └─ NO → Direct Commit for speed (Strategy 2)
```

---

## Strategy 1: Feature Branch + PR (Team Collaboration)

**Best For**: Teams of 2+ developers, projects requiring code review, scheduled releases

### Implementation Pattern

**Branch Creation**:
```bash
# Create feature branch from develop
git switch develop
git pull origin develop
git switch -c feature/SPEC-001-user-authentication

# Naming conventions
feature/SPEC-{id}-{brief-description}
feature/SPEC-042-jwt-token-validation
feature/SPEC-103-payment-gateway-integration
```

**TDD Development Cycle**:
```bash
# RED Phase: Write failing tests
git add tests/test_auth.py
git commit -m "test(auth): add failing JWT validation test

Add comprehensive test for JWT token validation:
- Valid token acceptance
- Expired token rejection
- Malformed token handling
- Missing signature detection

Related: SPEC-001"

# GREEN Phase: Minimal implementation
git add src/auth/jwt_validator.py
git commit -m "feat(auth): implement JWT token validation

Implement basic JWT validation:
- Token signature verification
- Expiration checking
- Claim extraction

Passes all tests in test_auth.py
Related: SPEC-001"

# REFACTOR Phase: Code quality improvement
git add src/auth/jwt_validator.py src/auth/exceptions.py
git commit -m "refactor(auth): improve error handling and type hints

Improvements:
- Add comprehensive type hints
- Extract validation logic to separate methods
- Improve error messages for debugging
- Add docstrings for all public methods

No behavior changes (all tests still pass)
Related: SPEC-001"
```

**Pull Request Creation**:
```bash
# Push feature branch
git push -u origin feature/SPEC-001-user-authentication

# Create PR with AI-generated description
gh pr create \
  --base develop \
  --title "feat(auth): implement JWT authentication (SPEC-001)" \
  --body "$(cat <<'EOF'
## Summary
Implements JWT token-based authentication as specified in SPEC-001.

### Changes
- JWT token validation with signature verification
- Token expiration checking
- Comprehensive error handling
- 95% test coverage (target: 85%)

### Testing
- [x] Unit tests for all validation scenarios
- [x] Integration tests with API endpoints
- [x] Security tests for token manipulation
- [x] Performance tests (< 5ms validation time)

### SPEC Compliance
- [x] REQ-001: Token signature validation
- [x] REQ-002: Expiration checking
- [x] REQ-003: Secure error messages
- [x] REQ-004: Performance requirements met

### Breaking Changes
None - new feature only

### Related Issues
Closes #42 (SPEC-001 implementation)
EOF
)"

# Alternative: Auto-generated description
gh pr create \
  --base develop \
  --title "feat(auth): implement JWT authentication" \
  --generate-description
```

**Code Review Process**:
```bash
# Request specific reviewers
gh pr create --reviewer alice,bob \
  --label "ready-for-review" \
  --milestone "Sprint 5"

# Address review feedback
git add src/auth/jwt_validator.py
git commit -m "fix(auth): address code review feedback

Changes based on @alice review:
- Add constant for token expiry tolerance
- Improve exception hierarchy
- Add logging for validation failures"

git push origin feature/SPEC-001-user-authentication
```

**PR Merge Workflow**:
```bash
# Option 1: Squash merge (clean history)
gh pr merge 123 --squash \
  --subject "feat(auth): implement JWT authentication (SPEC-001)" \
  --body "Complete JWT implementation with 95% test coverage" \
  --delete-branch

# Option 2: Rebase merge (preserve commit history)
gh pr merge 123 --rebase --delete-branch

# Option 3: Auto-merge when CI passes
gh pr merge 123 --auto --squash --delete-branch
```

### Advantages

- **Code Quality**: Mandatory peer review catches bugs early
- **Knowledge Sharing**: Team learns from each other's code
- **Audit Trail**: Complete history of who reviewed what
- **CI/CD Integration**: Automated tests before merge
- **Risk Mitigation**: Breaking changes caught before main branch

### Workflow Optimization

**Draft PRs for Early Feedback**:
```bash
# Create draft PR for work-in-progress
gh pr create --draft \
  --title "WIP: JWT authentication implementation" \
  --body "Early draft for architecture review"

# Convert to ready when complete
gh pr ready 123
```

**Parallel Feature Development**:
```bash
# Developer 1: Authentication
git switch -c feature/SPEC-001-auth

# Developer 2: Payment (independent)
git switch -c feature/SPEC-002-payment

# Both PRs merge independently
```

---

## Strategy 2: Direct Commit (Individual/Fast Development)

**Best For**: Solo developers, rapid prototyping, low-risk changes, high iteration speed

### Implementation Pattern

**Direct Development**:
```bash
# Work directly on develop branch
git switch develop
git pull origin develop

# TDD cycle with direct commits
# RED: Failing test
git add tests/test_database.py
git commit -m "test(db): add connection pool test"

# GREEN: Minimal implementation
git add src/database/pool.py
git commit -m "feat(db): add connection pool with size limits"

# REFACTOR: Improve code
git add src/database/pool.py src/database/config.py
git commit -m "refactor(db): extract pool configuration to separate module"

# Push directly to develop
git push origin develop
```

**Hotfix Pattern**:
```bash
# Critical bug fix on production
git switch main
git pull origin main

git add src/api/auth.py
git commit -m "fix(api): prevent null pointer in authentication

Critical fix for production issue #456:
- Add null check before user lookup
- Add defensive logging
- Deploy immediately

Tested manually - no time for full test suite"

git push origin main

# Backport to develop
git switch develop
git cherry-pick <commit-hash>
git push origin develop
```

### Advantages

- **Speed**: No PR review delay (instant merge)
- **Simplicity**: Minimal overhead for solo work
- **Rapid Iteration**: Quick experiment cycles
- **Low Ceremony**: No PR templates or review process

### Risk Mitigation

**Pre-commit Validation**:
```bash
#!/bin/bash
# .git/hooks/pre-commit

# Automatic code quality checks
black src/ tests/ --check || exit 1
ruff check src/ tests/ || exit 1
mypy src/ || exit 1

# Run tests before commit
pytest tests/ --quiet || exit 1

echo "✅ All checks passed - commit allowed"
```

**Automated CI on Develop**:
```yaml
# .github/workflows/develop-ci.yml
name: Develop Branch CI

on:
  push:
    branches: [develop]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run full test suite
        run: pytest tests/ --cov --cov-fail-under=85

      - name: Security scan
        run: bandit -r src/

      - name: Rollback on failure
        if: failure()
        run: git revert HEAD && git push
```

---

## Strategy 3: Per-SPEC Choice (Flexible Workflow)

**Best For**: Teams with variable requirements, projects with mixed complexity

### Decision Logic

**Automated Decision During `/moai:1-plan`**:
```python
def select_branching_strategy(spec: Specification) -> BranchingStrategy:
    """
    Determine optimal branching strategy based on SPEC characteristics.

    Returns:
        BranchingStrategy.FEATURE_BRANCH or BranchingStrategy.DIRECT_COMMIT
    """
    risk_factors = [
        spec.affects_critical_paths,
        spec.modifies_database_schema,
        spec.changes_api_contracts,
        spec.requires_security_review,
        spec.estimated_lines_changed > 500
    ]

    risk_score = sum(risk_factors)

    if risk_score >= 3:
        return BranchingStrategy.FEATURE_BRANCH
    elif risk_score == 2:
        # Ask user for decision
        return ask_user_branch_preference(spec)
    else:
        return BranchingStrategy.DIRECT_COMMIT
```

**Interactive User Decision**:
```bash
# MoAI-ADK prompts during /moai:1-plan
Creating SPEC-001: User authentication with JWT

Branching Strategy Decision:
  Risk Assessment: MEDIUM
  - Modifies security-critical code: YES
  - Affects API contracts: YES
  - Database changes: NO
  - Estimated LOC: 350

Recommendation: Feature branch (code review suggested)

Choose workflow:
  1. Feature branch + PR (recommended)
  2. Direct commit to develop

Your choice: 1

✅ Creating feature/SPEC-001-user-auth branch
```

### Configuration

**Project-Level Default**:
```json
{
  "git_strategy": {
    "mode": "per_spec",
    "decision_criteria": {
      "auto_feature_branch_if": {
        "critical_paths": true,
        "database_changes": true,
        "api_changes": true,
        "security_sensitive": true,
        "lines_changed_threshold": 500
      },
      "auto_direct_commit_if": {
        "documentation_only": true,
        "test_only": true,
        "config_changes": false,
        "lines_changed_threshold": 100
      }
    }
  }
}
```

---

## Real-World Scenarios

### Scenario 1: Startup (2-5 developers, fast iteration)

**Configuration**:
- Main strategy: Feature branch for major features
- Exception: Direct commit for bug fixes <50 lines
- Review: 1 approver required
- CI: Full test suite on PR

**Example Workflow**:
```bash
# New feature: Feature branch
/moai:1-plan "User profile dashboard"
# → Creates feature/SPEC-010 branch

# Bug fix: Direct commit
git switch develop
git commit -m "fix(ui): correct button alignment"
git push origin develop
```

### Scenario 2: Enterprise (20+ developers, compliance requirements)

**Configuration**:
- Main strategy: Feature branch mandatory
- Exception: None (all changes require PR)
- Review: 2 approvers required
- CI: Full test suite + security scan + compliance checks

**Example Workflow**:
```bash
# All changes go through PR
/moai:1-plan "Payment processing enhancement"
# → Creates feature/SPEC-042 branch
# → Creates draft PR immediately
# → Requires security team approval
```

### Scenario 3: Solo Developer (personal project)

**Configuration**:
- Main strategy: Direct commit
- Exception: Feature branch for experimental features
- Review: None (self-review)
- CI: Basic tests on develop

**Example Workflow**:
```bash
# Regular development: Direct commit
git commit -m "feat(blog): add markdown support"

# Experimental feature: Feature branch
git switch -c experiment/real-time-editing
# Test extensively before merge
```

---

## Best Practices Checklist

### ✅ DO

- Use descriptive branch names: `feature/SPEC-{id}-{description}`
- Follow Conventional Commits 2025 format
- Keep feature branches short-lived (<3 days active development)
- Delete merged branches immediately
- Sync with base branch daily: `git pull origin develop`
- Write clear PR descriptions with SPEC references
- Use draft PRs for early feedback
- Squash commits for clean history (feature branches only)

### ❌ DON'T

- Create long-running branches (>1 week without merge)
- Mix multiple features in one branch
- Force push to shared branches without team agreement
- Skip CI checks with `--no-verify`
- Leave stale branches unmerged for >2 weeks
- Merge without resolving all review comments
- Use generic branch names: `fix`, `update`, `test`

---

## Performance Metrics

**Feature Branch Workflow**:
- Average PR review time: 4-8 hours
- Merge cycle: 1-2 days
- Defect rate: 15% lower than direct commit
- Knowledge sharing: High

**Direct Commit Workflow**:
- Average commit-to-deploy: <15 minutes
- Merge cycle: Immediate
- Defect rate: Baseline
- Knowledge sharing: Low

**Per-SPEC Choice**:
- Critical features: Feature branch (review required)
- Minor fixes: Direct commit (faster)
- Hybrid defect rate: 8% lower than pure direct commit
- Flexibility: High

---

**Last Updated**: 2025-11-23
**Lines**: 289
