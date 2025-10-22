# Code Review Reference Guide

> **Version**: 2.0.0 (2025-10-22)
> **Based on**: GitLab/GitHub best practices, Google Engineering Practices, automated review tools (2025)

Complete reference for code review standards, automated tools, and quality gates.

---

## Table of Contents

1. [Code Review Standards](#code-review-standards)
2. [Conventional Comments Format](#conventional-comments-format)
3. [SOLID Principles Checklist](#solid-principles-checklist)
4. [Security Review Checklist (OWASP)](#security-review-checklist-owasp)
5. [Automated Review Tools Matrix](#automated-review-tools-matrix)
6. [Performance Review Patterns](#performance-review-patterns)
7. [API Design Standards](#api-design-standards)
8. [Test Review Criteria](#test-review-criteria)
9. [Documentation Standards](#documentation-standards)
10. [Review Decision Matrix](#review-decision-matrix)

---

## Code Review Standards

### Google's Code Review Philosophy

> **The Standard**: The primary purpose of code review is to make sure that the overall code health of the codebase is improving over time.

**Key Principles**:

1. **There is no "perfect" code** - only better code
2. **Reviewers should favor approving** a CL (change list) once it definitely improves the overall code health, even if it isn't perfect
3. **Continuous improvement** over perfection
4. **Balance** between progress and quality

### GitLab Code Review Guidelines

**Review Workflow**:
```
Author → Reviewer → Maintainer → Merge
```

**Reviewer Responsibilities**:
- ✅ Solution appropriateness
- ✅ Implementation quality
- ✅ Bug and logic issues
- ✅ Edge cases handling
- ✅ Compliance with standards

**Maintainer Responsibilities**:
- ✅ Overall architecture
- ✅ Code organization
- ✅ Consistency across codebase
- ✅ Quality standards
- ✅ Final approval

### Review Size Guidelines

| PR Size | Lines of Code | Review Time | Success Rate |
|---------|--------------|-------------|--------------|
| **Ideal** | < 200 lines | 15-30 min | 95% |
| **Acceptable** | 200-400 lines | 30-60 min | 85% |
| **Large** | 400-1000 lines | 1-2 hours | 60% |
| **Too Large** | > 1000 lines | > 2 hours | 30% |

**Recommendation**: PRs > 400 lines should be split into smaller, logical chunks.

### Review Response Time

**GitLab Guidelines**:
- **Initial Response**: Within 1 business day
- **Blocking Review**: Within 2 hours (if urgent)
- **Non-blocking Comments**: Address within 1-2 days

**Priority Levels**:
1. **P0 - Critical**: Security, production bugs (immediate)
2. **P1 - High**: Feature releases, critical fixes (< 4 hours)
3. **P2 - Normal**: Regular features (< 1 day)
4. **P3 - Low**: Refactoring, docs (< 3 days)

---

## Conventional Comments Format

### Structure

```
<label> [(<decoration>)]: <subject>

[<discussion>]
```

### Labels

| Label | Description | Typically Blocking |
|-------|-------------|-------------------|
| **praise** | Highlight something positive | No |
| **nitpick** | Trivial preference-based request | No |
| **suggestion** | Propose improvement or alternative | Depends |
| **issue** | Highlight specific problem | Usually Yes |
| **todo** | Small, necessary change | Usually Yes |
| **question** | Seek clarification or investigation | No |
| **thought** | Non-blocking observation | No |
| **chore** | Simple task (formatting, typo) | No |

### Decorations

| Decoration | Meaning |
|-----------|---------|
| **(blocking)** | Must be resolved before merge |
| **(non-blocking)** | Optional, can be addressed later |
| **(if-minor)** | Can be ignored if truly trivial |

### Examples

```markdown
**praise:** Great use of the Factory pattern here!

**nitpick (non-blocking):** Consider renaming `data` to `userData` for clarity.

**suggestion (blocking):** This function should validate input before processing.

**issue (blocking):** SQL injection vulnerability - use parameterized queries.

**question:** Have you tested this with negative values?

**thought:** We might want to cache this result in the future.

**chore:** Fix typo in comment (line 42: "recieve" → "receive").
```

---

## SOLID Principles Checklist

### Single Responsibility Principle (SRP)

**Review Questions**:
- [ ] Does this class/function have only one reason to change?
- [ ] Can you describe its purpose in a single sentence?
- [ ] Are there multiple unrelated responsibilities?
- [ ] Would splitting improve testability?

**Red Flags**:
- Class names with "And", "Or", "Manager", "Util"
- Functions > 50 lines
- Classes > 300 lines
- Methods doing I/O, logic, and formatting

**Refactoring**:
- Extract Class
- Extract Function
- Split Phase

### Open/Closed Principle (OCP)

**Review Questions**:
- [ ] Can new functionality be added without modifying existing code?
- [ ] Are there switch/case statements on type codes?
- [ ] Would adding a new type require changes in multiple places?
- [ ] Are abstractions used where appropriate?

**Red Flags**:
- Large switch/case or if-else chains on types
- Direct type checking (`instanceof`, `typeof`)
- Hard-coded type lists

**Refactoring**:
- Replace Conditional with Polymorphism
- Strategy Pattern
- Introduce Interface

### Liskov Substitution Principle (LSP)

**Review Questions**:
- [ ] Can subclasses replace base classes without breaking functionality?
- [ ] Do subclasses throw exceptions on inherited methods?
- [ ] Are preconditions strengthened in subclasses?
- [ ] Are postconditions weakened in subclasses?

**Red Flags**:
- Subclasses throwing `NotImplementedException`
- Overridden methods that do nothing
- Subclasses violating parent contracts

**Refactoring**:
- Replace Inheritance with Delegation
- Extract Interface
- Remove Subclass

### Interface Segregation Principle (ISP)

**Review Questions**:
- [ ] Are interfaces focused and cohesive?
- [ ] Do clients depend only on methods they use?
- [ ] Are there "fat" interfaces with many methods?
- [ ] Do implementers have to stub methods?

**Red Flags**:
- Interfaces with > 10 methods
- Empty method implementations
- Implementers throwing "not supported"

**Refactoring**:
- Split Interface
- Extract Interface
- Compose smaller interfaces

### Dependency Inversion Principle (DIP)

**Review Questions**:
- [ ] Do high-level modules depend on abstractions?
- [ ] Are dependencies injected rather than instantiated?
- [ ] Is the code testable with mocks?
- [ ] Are concrete classes used directly?

**Red Flags**:
- `new` keyword in business logic
- Direct database/API calls in services
- Hard-coded dependencies
- Untestable code

**Refactoring**:
- Introduce Dependency Injection
- Extract Interface
- Use Factory Pattern

---

## Security Review Checklist (OWASP)

### A01:2021 – Broken Access Control

**Check for**:
- [ ] Authentication required for protected resources
- [ ] Authorization checks on all endpoints
- [ ] User can only access own data
- [ ] No privilege escalation possible
- [ ] CORS configured correctly

**Common Vulnerabilities**:
- Missing authorization checks
- Insecure direct object references (IDOR)
- Path traversal attacks
- Forced browsing to admin pages

**Example Check**:
```python
# ❌ Missing authorization
@app.route('/api/user/<user_id>/orders')
def get_orders(user_id):
    return Order.query.filter_by(user_id=user_id).all()

# ✅ With authorization
@app.route('/api/user/<user_id>/orders')
@login_required
def get_orders(user_id):
    if current_user.id != user_id and not current_user.is_admin:
        abort(403)
    return Order.query.filter_by(user_id=user_id).all()
```

### A02:2021 – Cryptographic Failures

**Check for**:
- [ ] Sensitive data encrypted at rest
- [ ] Sensitive data encrypted in transit (HTTPS)
- [ ] Strong encryption algorithms (AES-256, RSA-2048+)
- [ ] No hardcoded secrets or keys
- [ ] Proper key management

**Common Vulnerabilities**:
- Hardcoded API keys, passwords
- Weak encryption (DES, MD5, SHA1)
- Sensitive data in logs
- Unencrypted storage

### A03:2021 – Injection

**Check for**:
- [ ] SQL queries use parameterized statements
- [ ] User input sanitized before rendering
- [ ] NoSQL queries parameterized
- [ ] OS commands avoided or sanitized
- [ ] LDAP queries parameterized

**Common Vulnerabilities**:
- SQL injection
- NoSQL injection
- OS command injection
- LDAP injection
- XSS (Cross-Site Scripting)

**Example Check**:
```python
# ❌ SQL Injection
query = f"SELECT * FROM users WHERE email = '{email}'"

# ✅ Parameterized
query = "SELECT * FROM users WHERE email = %s"
cursor.execute(query, (email,))
```

### A04:2021 – Insecure Design

**Check for**:
- [ ] Threat modeling performed
- [ ] Security requirements defined
- [ ] Rate limiting on APIs
- [ ] Input validation on all inputs
- [ ] Secure defaults used

### A05:2021 – Security Misconfiguration

**Check for**:
- [ ] Default credentials changed
- [ ] Unnecessary features disabled
- [ ] Error messages don't leak information
- [ ] Security headers configured (CSP, HSTS)
- [ ] Latest security patches applied

### A06:2021 – Vulnerable and Outdated Components

**Check for**:
- [ ] All dependencies up to date
- [ ] No known vulnerabilities in dependencies
- [ ] Dependency scanning in CI/CD
- [ ] Unused dependencies removed

**Tools**:
- Python: `pip-audit`, `safety`
- JavaScript: `npm audit`, `snyk`
- Java: `OWASP Dependency-Check`
- Go: `govulncheck`

### A07:2021 – Identification and Authentication Failures

**Check for**:
- [ ] Strong password policy
- [ ] Multi-factor authentication available
- [ ] Session management secure
- [ ] Password reset process secure
- [ ] Credential stuffing protection

### A08:2021 – Software and Data Integrity Failures

**Check for**:
- [ ] Code signing used
- [ ] Integrity checks on updates
- [ ] CI/CD pipeline secured
- [ ] Dependencies verified

### A09:2021 – Security Logging and Monitoring Failures

**Check for**:
- [ ] All authentication events logged
- [ ] All authorization failures logged
- [ ] Logs monitored for anomalies
- [ ] Alerts configured for suspicious activity

### A10:2021 – Server-Side Request Forgery (SSRF)

**Check for**:
- [ ] User-supplied URLs validated
- [ ] Whitelist of allowed domains
- [ ] Internal network access restricted
- [ ] URL parsing libraries used correctly

---

## Automated Review Tools Matrix

### Language-Specific Tools

#### Python

| Tool | Purpose | Installation | Config |
|------|---------|--------------|--------|
| **ruff** | Linter + Formatter (fastest) | `pip install ruff` | `ruff.toml` |
| **pylint** | Comprehensive linter | `pip install pylint` | `.pylintrc` |
| **mypy** | Static type checking | `pip install mypy` | `mypy.ini` |
| **black** | Code formatter | `pip install black` | `pyproject.toml` |
| **pytest** | Test runner + coverage | `pip install pytest pytest-cov` | `pytest.ini` |
| **bandit** | Security linter | `pip install bandit` | `.bandit` |
| **safety** | Dependency vulnerability scan | `pip install safety` | - |

**Recommended Command**:
```bash
ruff check . && mypy . && pytest --cov=src --cov-report=term-missing
```

#### TypeScript/JavaScript

| Tool | Purpose | Installation | Config |
|------|---------|--------------|--------|
| **ESLint** | Linter | `npm i -D eslint` | `.eslintrc.js` |
| **eslint-plugin-sonarjs** | Code smell detection | `npm i -D eslint-plugin-sonarjs` | In `.eslintrc.js` |
| **Prettier** | Formatter | `npm i -D prettier` | `.prettierrc` |
| **Biome** | All-in-one (fast) | `npm i -D @biomejs/biome` | `biome.json` |
| **TypeScript** | Type checking | `npm i -D typescript` | `tsconfig.json` |
| **Vitest** | Test runner | `npm i -D vitest` | `vitest.config.ts` |

**Recommended Command**:
```bash
biome check . && tsc --noEmit && vitest run --coverage
```

#### Go

| Tool | Purpose | Installation | Config |
|------|---------|--------------|--------|
| **golangci-lint** | Meta-linter (20+ linters) | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` | `.golangci.yml` |
| **gofmt** | Formatter | Built-in | - |
| **go vet** | Bug detector | Built-in | - |
| **staticcheck** | Static analysis | `go install honnef.co/go/tools/cmd/staticcheck@latest` | - |
| **govulncheck** | Vulnerability scanner | `go install golang.org/x/vuln/cmd/govulncheck@latest` | - |

**Recommended Command**:
```bash
golangci-lint run && go test -race -coverprofile=coverage.out ./...
```

#### Rust

| Tool | Purpose | Installation | Config |
|------|---------|--------------|--------|
| **clippy** | Linter | `rustup component add clippy` | - |
| **rustfmt** | Formatter | `rustup component add rustfmt` | `rustfmt.toml` |
| **cargo-audit** | Dependency audit | `cargo install cargo-audit` | - |
| **cargo-tarpaulin** | Coverage | `cargo install cargo-tarpaulin` | - |

**Recommended Command**:
```bash
cargo clippy -- -D warnings && cargo test && cargo tarpaulin --out Html
```

### Universal Tools

| Tool | Purpose | Languages | Installation |
|------|---------|-----------|--------------|
| **SonarQube** | Static analysis, coverage, duplicates | 25+ | Docker/Cloud |
| **CodeClimate** | Quality metrics, test coverage | 15+ | Cloud |
| **Codacy** | Automated code review | 40+ | Cloud |
| **jscpd** | Duplicate code detection | All | `npm i -g jscpd` |
| **gitleaks** | Secret detection | All | `brew install gitleaks` |
| **trivy** | Vulnerability scanner | All | `brew install trivy` |

---

## Performance Review Patterns

### Database Performance

#### N+1 Query Problem

**Detection**:
```python
# Profile with SQL logging
import django.db
from django.conf import settings

settings.DEBUG = True  # Enable query logging

# Execute code
result = some_function()

# Check query count
print(f"Queries executed: {len(django.db.connection.queries)}")
```

**Common Patterns**:

| Pattern | Problem | Solution |
|---------|---------|----------|
| **Loop + Query** | `for x in items: query(x.id)` | Use `select_related()` or `prefetch_related()` |
| **Lazy Loading** | Accessing relations in template | Eager load with JOINs |
| **Repeated Queries** | Same query in loop | Cache results |

#### Slow Queries

**Review Checklist**:
- [ ] Indexes on foreign keys
- [ ] Indexes on WHERE clause columns
- [ ] Avoid `SELECT *`, select only needed columns
- [ ] Use EXPLAIN to analyze query plan
- [ ] Pagination for large result sets
- [ ] Consider materialized views for complex queries

### Algorithm Complexity

| Pattern | Bad | Good | Complexity |
|---------|-----|------|------------|
| **Search** | Linear search | Binary search (if sorted) | O(n) → O(log n) |
| **Duplicates** | Nested loops | Set/Map | O(n²) → O(n) |
| **Sorting** | Bubble sort | Built-in sort (Timsort) | O(n²) → O(n log n) |
| **Unique values** | Array.filter + includes | Set | O(n²) → O(n) |

**Review Questions**:
- [ ] What is the Big-O complexity?
- [ ] Does it scale to expected input sizes?
- [ ] Are there built-in methods that are faster?
- [ ] Can caching reduce redundant work?

### Caching Strategies

**Review Checklist**:
- [ ] Cache expensive computations
- [ ] Use appropriate TTL (Time To Live)
- [ ] Implement cache invalidation strategy
- [ ] Consider memory vs. compute trade-off
- [ ] Use CDN for static assets
- [ ] Implement HTTP caching headers

**Caching Layers**:
1. **In-Memory** (Redis, Memcached) - µs latency
2. **Database Query Cache** - ms latency
3. **Application Cache** - Local memory
4. **CDN** - Edge locations
5. **Browser Cache** - Client-side

---

## API Design Standards

### RESTful API Checklist

**URL Structure**:
- [ ] Use nouns, not verbs (`/users`, not `/getUsers`)
- [ ] Use plural for collections (`/users`, not `/user`)
- [ ] Use hierarchy for relationships (`/users/123/orders`)
- [ ] Version your API (`/api/v1/users`)
- [ ] Use kebab-case or snake_case consistently

**HTTP Methods**:

| Method | Purpose | Idempotent | Safe | Example |
|--------|---------|------------|------|---------|
| **GET** | Retrieve resource | Yes | Yes | `GET /users/123` |
| **POST** | Create resource | No | No | `POST /users` |
| **PUT** | Replace resource | Yes | No | `PUT /users/123` |
| **PATCH** | Update resource | No | No | `PATCH /users/123` |
| **DELETE** | Delete resource | Yes | No | `DELETE /users/123` |

**HTTP Status Codes**:

| Code | Meaning | Use Case |
|------|---------|----------|
| **200** | OK | Successful GET, PUT, PATCH, DELETE |
| **201** | Created | Successful POST |
| **204** | No Content | Successful DELETE (no body) |
| **400** | Bad Request | Invalid input |
| **401** | Unauthorized | Missing/invalid authentication |
| **403** | Forbidden | Authenticated but not authorized |
| **404** | Not Found | Resource doesn't exist |
| **409** | Conflict | Duplicate resource |
| **422** | Unprocessable Entity | Validation error |
| **429** | Too Many Requests | Rate limit exceeded |
| **500** | Internal Server Error | Unexpected server error |
| **503** | Service Unavailable | Server overloaded |

**Response Format**:

```json
{
  "data": {
    "id": 123,
    "name": "John Doe"
  },
  "meta": {
    "version": "v1",
    "timestamp": "2025-10-22T10:30:00Z"
  }
}
```

**Error Format**:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid email format",
    "details": [
      {
        "field": "email",
        "issue": "Must be a valid email address"
      }
    ]
  }
}
```

### GraphQL API Checklist

**Schema Design**:
- [ ] Use clear, descriptive type names
- [ ] Implement pagination (Relay-style or offset)
- [ ] Use enums for fixed value sets
- [ ] Document all fields with descriptions
- [ ] Implement proper error handling

**Query Performance**:
- [ ] Implement DataLoader for batching
- [ ] Set query depth limits
- [ ] Set query complexity limits
- [ ] Implement rate limiting
- [ ] Use persisted queries for production

---

## Test Review Criteria

### Test Coverage Standards

**MoAI-ADK Target**: ≥ 85%

**Coverage Types**:
- **Line Coverage**: % of lines executed
- **Branch Coverage**: % of if/else branches taken
- **Function Coverage**: % of functions called
- **Statement Coverage**: % of statements executed

**Critical Areas** (100% coverage required):
- Authentication/authorization logic
- Payment processing
- Data validation
- Security-sensitive code

### Test Quality Checklist

**Structure (AAA Pattern)**:
- [ ] **Arrange**: Setup test data and dependencies
- [ ] **Act**: Execute the code under test
- [ ] **Assert**: Verify the expected outcome

**Test Characteristics**:
- [ ] **Fast**: < 1 second per test
- [ ] **Isolated**: No dependencies between tests
- [ ] **Repeatable**: Same result every time
- [ ] **Self-validating**: Pass/fail without manual inspection
- [ ] **Timely**: Written before/with production code (TDD)

**Edge Cases**:
- [ ] Null/undefined inputs
- [ ] Empty collections
- [ ] Boundary values (0, -1, MAX_INT)
- [ ] Invalid inputs
- [ ] Error conditions

**Example**:
```python
def test_divide():
    # Arrange
    calculator = Calculator()

    # Act
    result = calculator.divide(10, 2)

    # Assert
    assert result == 5

def test_divide_by_zero():
    calculator = Calculator()

    with pytest.raises(ZeroDivisionError):
        calculator.divide(10, 0)
```

---

## Documentation Standards

### Code Documentation

**Function/Method Documentation**:

**Python (Docstring)**:
```python
def calculate_discount(price: float, discount_percent: float) -> float:
    """
    Calculate discounted price.

    Args:
        price: Original price in dollars
        discount_percent: Discount percentage (0-100)

    Returns:
        Discounted price in dollars

    Raises:
        ValueError: If discount_percent < 0 or > 100

    Example:
        >>> calculate_discount(100.0, 20.0)
        80.0
    """
```

**TypeScript (JSDoc)**:
```typescript
/**
 * Calculate discounted price.
 *
 * @param price - Original price in dollars
 * @param discountPercent - Discount percentage (0-100)
 * @returns Discounted price in dollars
 * @throws {RangeError} If discountPercent < 0 or > 100
 *
 * @example
 * ```typescript
 * calculateDiscount(100, 20) // Returns 80
 * ```
 */
function calculateDiscount(price: number, discountPercent: number): number {
  // ...
}
```

### README Standards

**Required Sections**:
1. **Project Title** + Short description
2. **Installation** instructions
3. **Usage** examples
4. **API** documentation (if applicable)
5. **Contributing** guidelines
6. **License**

**Optional Sections**:
- Screenshots
- Features list
- Changelog
- Roadmap
- FAQ

---

## Review Decision Matrix

### Approval Criteria

| Criteria | Weight | Pass Threshold |
|----------|--------|----------------|
| Tests Pass | Critical | 100% |
| Coverage | High | ≥ 85% |
| No Security Issues | Critical | 100% |
| Linting | High | 100% |
| Complexity | Medium | ≤ 10 per function |
| Documentation | Medium | Complete |
| SOLID Compliance | Medium | No major violations |
| Performance | Low | No regressions |

### Decision Tree

```
1. Are there CRITICAL security issues?
   YES → ❌ Request Changes
   NO → Continue

2. Do all tests pass?
   NO → ❌ Request Changes
   YES → Continue

3. Is coverage ≥ 85%?
   NO → ⚠️ Conditional (discuss)
   YES → Continue

4. Are there blocking code smells?
   YES → ❌ Request Changes
   NO → Continue

5. Does it improve code health?
   NO → ❌ Request Changes
   YES → ✅ Approve
```

### Common Review Outcomes

| Status | Icon | Meaning | Next Step |
|--------|------|---------|-----------|
| **Approve** | ✅ | Meets all standards | Merge when ready |
| **Approve with Comments** | ⚠️✅ | Meets standards, non-blocking suggestions | Merge, address later |
| **Request Changes** | ❌ | Blocking issues found | Fix and re-request review |
| **Comment** | 💬 | Questions or observations | Author responds |

---

## Quick Reference Cards

### Pre-Review Checklist

- [ ] Read PR description and linked issues
- [ ] Understand the context and requirements
- [ ] Check CI/CD pipeline status
- [ ] Review automated tool reports (SonarQube, ESLint)
- [ ] Pull branch and test locally (if complex)

### During Review Checklist

- [ ] Code implements stated requirements
- [ ] SOLID principles followed
- [ ] No security vulnerabilities
- [ ] Tests cover happy path and edge cases
- [ ] Performance is acceptable
- [ ] API design is consistent
- [ ] Documentation is complete
- [ ] Error handling is appropriate
- [ ] Code is readable and maintainable

### Post-Review Checklist

- [ ] All comments are clear and actionable
- [ ] Blocking vs. non-blocking clearly marked
- [ ] Tone is constructive and respectful
- [ ] Examples provided where helpful
- [ ] Decision (approve/request changes) justified

---

## References

- [Google Engineering Practices: Code Review](https://google.github.io/eng-practices/review/)
- [GitLab Code Review Guidelines](https://docs.gitlab.com/ee/development/code_review.html)
- [Conventional Comments](https://conventionalcomments.org/)
- [OWASP Top 10 (2021)](https://owasp.org/Top10/)
- [REST API Design Best Practices](https://restfulapi.net/)
- [SonarQube Documentation](https://docs.sonarqube.org/)

---

**Version**: 2.0.0
**Last Updated**: 2025-10-22
**Part of**: MoAI-ADK Essentials Skills
**Companion**: See examples.md for practical review scenarios
