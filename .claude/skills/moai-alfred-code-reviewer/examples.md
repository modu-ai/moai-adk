# Code Review Examples

_Last updated: 2025-10-22_

## Example 1: SOLID Principles Review

### Scenario
Reviewing a Python authentication service for SOLID principle compliance.

### Code Under Review
```python
class UserService:
    def create_user(self, data):
        # Validate data
        if not data.get('email'):
            raise ValueError("Email required")

        # Hash password
        import hashlib
        hashed = hashlib.sha256(data['password'].encode()).hexdigest()

        # Save to database
        db.execute("INSERT INTO users ...")

        # Send welcome email
        smtp.send_email(data['email'], "Welcome!")

        return user
```

### Review Findings

**Violations**:
- Single Responsibility Principle: Class handles validation, hashing, database, and email
- Dependency Inversion: Hard-coded dependencies (db, smtp, hashlib)

**Recommendations**:
```python
class UserService:
    def __init__(self, validator, hasher, repository, notifier):
        self.validator = validator
        self.hasher = hasher
        self.repository = repository
        self.notifier = notifier

    def create_user(self, data):
        self.validator.validate(data)
        hashed_password = self.hasher.hash(data['password'])
        user = self.repository.create(data['email'], hashed_password)
        self.notifier.send_welcome(user.email)
        return user
```

---

## Example 2: Static Analysis Integration

### Scenario
Setting up automated code review pipeline with static analysis tools.

### Configuration
```yaml
# .github/workflows/code-review.yml
name: Automated Code Review
on: [pull_request]

jobs:
  static-analysis:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # Python: ruff for linting
      - name: Run Ruff
        run: |
          pip install ruff
          ruff check --output-format=github .

      # TypeScript: Biome for linting
      - name: Run Biome
        run: |
          npm install -g @biomejs/biome
          biome check --write .

      # Security: Semgrep for SAST
      - name: Run Semgrep
        run: |
          pip install semgrep
          semgrep scan --config=auto
```

### Review Automation Results
```
✓ Code style: 0 issues
✓ Type safety: 2 warnings (non-blocking)
✗ Security: 1 critical issue found
  - SQL injection vulnerability in user_query.py:42
  - Recommendation: Use parameterized queries
```

---

## Example 3: Clean Code Principles Review

### Scenario
Reviewing JavaScript code for readability and maintainability.

### Code Under Review
```javascript
function p(u) {
  const d = new Date();
  const x = d.getTime() - u.c;
  if (x > 86400000) {
    return "expired";
  }
  return "active";
}
```

### Review Findings

**Issues**:
- Non-descriptive names: `p`, `u`, `d`, `x`
- Magic number: `86400000`
- Missing comments

**Improved Version**:
```javascript
const MILLISECONDS_PER_DAY = 24 * 60 * 60 * 1000;

/**
 * Checks if a user's subscription is expired
 * @param {Object} user - User object with creation timestamp
 * @returns {string} 'expired' or 'active'
 */
function checkSubscriptionStatus(user) {
  const now = new Date();
  const elapsedTime = now.getTime() - user.createdAt;

  if (elapsedTime > MILLISECONDS_PER_DAY) {
    return "expired";
  }

  return "active";
}
```

---

## Example 4: TRUST 5 Principles Review

### Scenario
Comprehensive review against MoAI-ADK TRUST 5 principles.

### Checklist Application

**T - Test First (Target: ≥85% coverage)**
```bash
$ pytest --cov=src --cov-report=term-missing
TOTAL                     87%    ✓ Passes threshold
Missing coverage in:
  - src/error_handler.py lines 42-45 (error edge case)
```

**R - Readable (Linting & formatting)**
```bash
$ ruff check .
All checks passed!                ✓

$ black --check .
All files formatted correctly     ✓
```

**U - Unified (Type safety)**
```bash
$ mypy src/
src/utils.py:23: error: Missing return type
                          ✗ Add type hints
```

**S - Secured (Security analysis)**
```bash
$ semgrep --config=auto
Found: Hardcoded API key in config.py:15
                          ✗ Use environment variables
```

**T - Trackable (@TAG coverage)**
```bash
$ rg '@(CODE|TEST|SPEC):AUTH' -c
@CODE:AUTH - 12 occurrences    ✓
@TEST:AUTH - 12 occurrences    ✓
@SPEC:AUTH - 1 occurrence      ✓
All TAGs properly linked
```

---

_For detailed patterns and best practices, see SKILL.md and reference.md_
