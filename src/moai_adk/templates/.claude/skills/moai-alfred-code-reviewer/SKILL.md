---
name: "moai-alfred-code-reviewer"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-12
status: stable
description: Enterprise systematic code review orchestrator with TRUST 5 principles, multi-language support, Context7 integration, AI-powered quality checks, SOLID principle validation, security vulnerability detection, and maintainability analysis across 25+ programming languages; activates for code reviews, quality standard validation, TRUST 5 enforcement, architectural audits, and automated review automation
keywords: ['code-review', 'TRUST-5', 'quality-assurance', 'SOLID-principles', 'security-scanning', 'maintainability-analysis', 'ai-code-review', 'context7-integration', 'automated-review', 'enterprise-quality']
allowed-tools: 
  - Read
  - Write
  - Edit
  - Glob
  - Bash
  - AskUserQuestion
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
  - WebFetch
---

# Enterprise Systematic Code Review Orchestrator v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-code-reviewer |
| **Version** | 4.0.0 Enterprise (2025-11-12) |
| **Core Framework** | TRUST 5 principles, Context7 integration |
| **AI Integration** | ✅ Context7 MCP, AI quality checks, pattern matching |
| **Auto-load** | When conducting code reviews or quality checks |
| **Languages** | 25+ (Python, JavaScript, Go, Rust, Java, etc.) |
| **Lines of Content** | 950+ with 15+ production examples |
| **Progressive Disclosure** | 3-level (framework, patterns, advanced) |

---

## What It Does

Provides systematic guidance for enterprise-grade code review processes applying **TRUST 5 principles**, validating SOLID principles, identifying security issues, ensuring maintainability, and automating quality gates across all programming languages.

---

## The TRUST 5 Review Framework

### T - Test First
**Focus**: Test coverage, quality, comprehensiveness

**Key Questions**:
- Are tests comprehensive? Do they cover happy path + edge cases?
- Test coverage ≥ 85%?
- Tests verify behavior, not just implementation?
- Edge cases handled: null, empty, boundary values?
- Async/concurrent scenarios tested?

**Tools**: pytest coverage, jest --coverage, go test -cover, cargo test

**Examples**:
```python
# ❌ Bad: Tests implementation, not behavior
def test_add():
    assert add(2, 2) == 4

# ✅ Good: Tests behavior with edge cases
def test_add_positive_numbers():
    assert add(2, 2) == 4
    assert add(-1, 1) == 0
    assert add(0, 0) == 0
    
def test_add_boundary():
    assert add(int.max, 1) == overflow_behavior()
```

### R - Readable
**Focus**: Code clarity, self-documentation, maintainability

**Key Questions**:
- Are function/variable names meaningful and clear?
- Can a new team member understand the intent?
- Comments explain WHY, not WHAT (code shows what)?
- Cyclomatic complexity reasonable (<10)?
- Functions single responsibility?
- Magic numbers extracted as constants?

**Tools**: linters, code formatters, readability checkers

**Examples**:
```python
# ❌ Bad: Unclear intent
def calc(x, y, z):
    return x * (1 + y / 100) - z * 0.05

# ✅ Good: Clear intent with constants
DISCOUNT_RATE = 0.05
TAX_RATE = 0.05

def calculate_final_price(base_price: float, tax_percent: float, discount: float) -> float:
    """Calculate final price after tax and discount.
    
    Args:
        base_price: Original product price
        tax_percent: Tax percentage (0-100)
        discount: Discount amount to subtract
    """
    with_tax = base_price * (1 + tax_percent / 100)
    return with_tax - (discount * DISCOUNT_RATE)
```

### U - Unified
**Focus**: Consistency, patterns, architectural cohesion

**Key Questions**:
- Does code follow team patterns and conventions?
- Consistent with codebase style?
- Uses established error handling patterns?
- Logging strategy aligned?
- Database access follows repository pattern?
- API design consistent with existing endpoints?

**Tools**: style guides, architectural patterns, linters

**Examples**:
```python
# ❌ Bad: Inconsistent error handling
def get_user(user_id):
    try:
        return fetch_from_db(user_id)
    except Exception as e:
        return None  # Inconsistent with rest of codebase

# ✅ Good: Consistent error handling
def get_user(user_id: int) -> User:
    """Get user by ID.
    
    Raises:
        UserNotFoundError: If user doesn't exist
        DatabaseError: If database connection fails
    """
    try:
        return self.user_repository.find_by_id(user_id)
    except DatabaseConnectionError as e:
        logger.error(f"Database error: {e}")
        raise DatabaseError(str(e)) from e
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        raise
```

### S - Secured
**Focus**: Security vulnerabilities, input validation, secret handling

**Key Questions**:
- Are inputs validated before use?
- No hardcoded credentials, API keys, or secrets?
- SQL injection prevention (parameterized queries)?
- XSS prevention (output escaping)?
- CSRF tokens used for state-changing operations?
- Authentication required for sensitive operations?
- Rate limiting on public endpoints?
- Dependency vulnerabilities scanned?

**Tools**: bandit, safety, npm audit, go vet, security scanners

**Examples**:
```python
# ❌ Bad: SQL injection vulnerability
def get_user(user_id):
    query = f"SELECT * FROM users WHERE id = {user_id}"  # Vulnerable!
    return db.execute(query)

# ✅ Good: Parameterized query
def get_user(user_id: int) -> User:
    query = "SELECT * FROM users WHERE id = ?"
    return db.execute(query, [user_id])
```

### T - Trackable

**Key Questions**:
- Changelog entry added?
- Git history clear and atomic?
- Breaking changes documented?
- Migration guides for version updates?


**Examples**:
```python
# ❌ Bad: No traceability
def calculate_discount(price, customer_type):
    if customer_type == "vip":
        return price * 0.8
    return price

# ✅ Good: Full traceability
def calculate_discount(price: float, customer_type: str) -> float:
    """Calculate discount based on customer type.
    
    Implements SPEC-042: VIP customers receive 20% discount
    
    Linked to:
    - SPEC-042: VIP pricing requirements
    - TEST-042-001: VIP discount validation
    - PR #1234: Feature implementation
    """
    VIP_DISCOUNT_RATE = 0.20
    
    if customer_type == "vip":
        return price * (1 - VIP_DISCOUNT_RATE)
    return price
```

---

## SOLID Principles Checklist

| Principle | Focus | Review Question |
|-----------|-------|-----------------|
| **S**ingle Responsibility | One reason to change | Does this class/function do one thing? |
| **O**pen/Closed | Open for extension, closed for modification | Can behavior be extended without modifying? |
| **L**iskov Substitution | Substitutable subtypes | Can derived classes replace base without breaking? |
| **I**nterface Segregation | Minimal, specific interfaces | Are clients forced to depend on methods they don't use? |
| **D**ependency Inversion | Depend on abstractions, not concretions | Do high-level modules depend on low-level implementations? |

---

## Code Review Process (4-Step)

### Step 1: Automated Checks (5 min)
```
✓ Linting & formatting
✓ Security scanning (bandit, safety, npm audit)
✓ Dependency vulnerabilities
✓ Test coverage ≥85%
✓ Type checking (mypy, TypeScript, etc.)
```

### Step 2: Architecture Review (15 min)
```
✓ SOLID principles
✓ Design patterns appropriate?
✓ Consistency with codebase
✓ Scalability implications?
✓ Performance implications?
```

### Step 3: Security Audit (10 min)
```
✓ Input validation
✓ No hardcoded secrets
✓ Authentication/authorization correct?
✓ SQL injection prevention
✓ XSS prevention
✓ CSRF tokens present
```

### Step 4: Implementation Review (20 min)
```
✓ TRUST 5 checklist
✓ Edge cases handled?
✓ Error messages helpful?
✓ Documentation complete?
```

---

## Review Depth Matrix

| Change Type | Severity | Automation | Review Time | Focus Areas |
|-------------|----------|-----------|------------|------------|
| **Security fix** | 🔴 Critical | Full scan | 30+ min | Vulnerabilities, test coverage, audit trail |
| **Core architecture** | 🔴 Critical | Partial | 45+ min | Design patterns, scalability, consistency |
| **Feature (new)** | 🟡 Major | Full scan | 30 min | Completeness, TRUST 5, documentation |
| **Bug fix** | 🟢 Minor | Partial | 15 min | Root cause, test coverage, regressions |
| **Documentation** | 🟢 Minor | Basic | 5 min | Accuracy, completeness, examples |
| **Configuration** | 🟡 Medium | Full | 10 min | Security, best practices, side effects |
| **Refactoring** | 🟢 Minor | Full | 15 min | Behavior preservation, performance |

---

## Best Practices

### DO
- **Automate repetitive checks**: Linting, coverage, formatting
- **Focus human review on high-value areas**: Architecture, security, design
- **Be constructive**: Review code, not people
- **Explain WHY**: Help reviewer understand the reasoning
- **Request specific changes**: Not vague "improve this"
- **Provide examples**: Show the better approach
- **Flag trade-offs**: Explain choices made
- **Document decisions**: Comment on why certain patterns chosen

### DON'T
- **Nitpick style**: Let linters handle formatting
- **Reject without alternatives**: Always suggest improvements
- **Make personal comments**: Focus on code quality
- **Review when tired**: Quality suffers
- **Block on minor issues**: Distinguish critical from nice-to-have
- **Skip security review**: Always check authentication, validation, secrets
- **Ignore test coverage**: Enforce ≥85% requirement

---

## Integration with Context7

**Live Security Scanning**: Get latest vulnerability patterns from official databases  
**Best Practice Integration**: Apply latest security recommendations from official docs  
**Version-Aware Checks**: Context7 provides version-specific security guidance  
**Automated Fix Suggestions**: Context7 patterns for common vulnerability fixes

---

## Related Skills

- `moai-alfred-practices` (Code patterns and best practices)
- `moai-essentials-refactor` (Refactoring strategies)

---

**For detailed review checklists**: [reference.md](reference.md)  
**For real-world examples**: [examples.md](examples.md)  
**Last Updated**: 2025-11-12  
**Status**: Production Ready (Enterprise v4.0.0)
