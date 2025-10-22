# Code Review Examples: Real-World Scenarios

> **Version**: 2.0.0 (2025-10-22)
> **Based on**: GitLab/GitHub best practices, Google Engineering Practices, SOLID principles, automated review tools (2025)

This document provides concrete code review examples demonstrating how to identify issues, provide constructive feedback, and enforce quality standards.

---

## Table of Contents

1. [Pull Request Review: Authentication Module](#1-pull-request-review-authentication-module)
2. [SOLID Principles Violations](#2-solid-principles-violations)
3. [Security Issues Detection](#3-security-issues-detection)
4. [Performance Review Scenario](#4-performance-review-scenario)
5. [API Design Review](#5-api-design-review)
6. [Test Quality Review](#6-test-quality-review)
7. [Documentation Review](#7-documentation-review)
8. [Automated Review Tool Integration](#8-automated-review-tool-integration)
9. [Pre-commit Hook Review Workflow](#9-pre-commit-hook-review-workflow)
10. [GitLab/GitHub Review Best Practices](#10-gitlabgithub-review-best-practices)

---

## 1. Pull Request Review: Authentication Module

### PR Context

**Title**: Add JWT-based authentication
**Description**: Implements JWT tokens for user authentication
**Files Changed**: 3 files (+250 lines, -10 lines)
**Reviewer**: @reviewer-name
**SPEC**: SPEC-AUTH-001.md

### Code Submitted

```python
# src/auth/jwt_handler.py
import jwt
from datetime import datetime

SECRET_KEY = "my_secret_key_123"  # ⚠️ Issue 1

def create_token(user_id, email):  # ⚠️ Issue 2
    payload = {
        "user_id": user_id,
        "email": email,
        "exp": datetime.utcnow() + timedelta(hours=24)
    }
    token = jwt.encode(payload, SECRET_KEY, algorithm="HS256")
    return token

def verify_token(token):  # ⚠️ Issue 3
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=["HS256"])
        return payload
    except:  # ⚠️ Issue 4
        return None
```

### Review Comments (Conventional Comment Format)

#### Critical Issues

**Comment 1 (Blocking):**
```markdown
**suggestion (blocking):** Hardcoded secret key is a **critical security vulnerability**.

**Problem**: Secret keys must never be committed to version control.

**Solution**: Use environment variables:

```python
import os
SECRET_KEY = os.getenv("JWT_SECRET_KEY")
if not SECRET_KEY:
    raise ValueError("JWT_SECRET_KEY environment variable not set")
```

**Reference**: OWASP A02:2021 – Cryptographic Failures

**Refs**: @TAG:AUTH-001
```

**Comment 2 (Blocking):**
```markdown
**nitpick (blocking):** Missing import for `timedelta`.

```python
from datetime import datetime, timedelta  # Add timedelta
```

This will cause a runtime error. Please add a test case to catch this.
```

**Comment 3 (Non-blocking):**
```markdown
**suggestion (non-blocking):** Add input validation and type hints.

```python
def create_token(user_id: int, email: str) -> str:
    """Create JWT token for authenticated user."""
    if not user_id or not email:
        raise ValueError("user_id and email are required")

    # Implementation...
```

**Why**: Improves type safety and prevents None values from causing issues downstream.
```

**Comment 4 (Blocking):**
```markdown
**issue (blocking):** Bare `except` clause swallows all exceptions.

**Problem**: This catches `KeyboardInterrupt`, `SystemExit`, and masks real errors.

**Solution**:
```python
def verify_token(token: str) -> dict | None:
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=["HS256"])
        return payload
    except jwt.ExpiredSignatureError:
        logger.warning("Token expired")
        return None
    except jwt.InvalidTokenError as e:
        logger.error(f"Invalid token: {e}")
        return None
```

**Reference**: PEP 8, Python Anti-Patterns
```

#### Additional Observations

**Praise:**
```markdown
**praise:** Great start on implementing JWT authentication! The payload structure looks good.
```

**Question:**
```markdown
**question:** Have you considered adding refresh tokens for better UX? Short-lived access tokens + long-lived refresh tokens is a common pattern.

Not blocking this PR, but worth discussing for a follow-up.
```

### Review Decision

**Status**: ❌ **Changes Requested**

**Summary**:
- 2 blocking security issues
- 1 blocking runtime error
- 1 non-blocking improvement suggestion
- Must address all blocking issues before approval

**Next Steps**:
1. Move secret to environment variable
2. Add missing import
3. Fix exception handling
4. Add test coverage for error cases
5. Re-request review after fixes

---

## 2. SOLID Principles Violations

### Single Responsibility Principle (SRP) Violation

**Before** (Review Comment):
```typescript
// ❌ NITPICK: This class has too many responsibilities

class UserService {
  // Database operations
  async createUser(data: UserData): Promise<User> { /* ... */ }
  async updateUser(id: string, data: UserData): Promise<User> { /* ... */ }
  async deleteUser(id: string): Promise<void> { /* ... */ }

  // Email operations
  async sendWelcomeEmail(user: User): Promise<void> { /* ... */ }
  async sendPasswordResetEmail(email: string): Promise<void> { /* ... */ }

  // Validation
  validateEmail(email: string): boolean { /* ... */ }
  validatePassword(password: string): boolean { /* ... */ }

  // Authentication
  async login(email: string, password: string): Promise<string> { /* ... */ }
  async logout(token: string): Promise<void> { /* ... */ }
}
```

**Review Comment**:
```markdown
**suggestion (blocking):** `UserService` violates Single Responsibility Principle.

**Issue**: This class handles database, email, validation, and authentication - 4 different responsibilities.

**Refactoring Plan**:
1. **UserRepository**: Database operations (create, update, delete)
2. **EmailService**: Email sending logic
3. **UserValidator**: Validation rules
4. **AuthService**: Login/logout/token management

**Benefit**: Each class becomes easier to test, modify, and reason about.

**SOLID**: Single Responsibility Principle

**Refs**: @TAG:USER-001, SPEC-USER-001.md
```

### Open/Closed Principle (OCP) Violation

**Before**:
```python
# ❌ ISSUE: Hard to extend, must modify for new payment types

def process_payment(payment_type, amount):
    if payment_type == "credit_card":
        # Process credit card
        fee = amount * 0.029
    elif payment_type == "paypal":
        # Process PayPal
        fee = amount * 0.034
    elif payment_type == "bank_transfer":
        # Process bank transfer
        fee = amount * 0.01
    else:
        raise ValueError("Unknown payment type")

    return amount + fee
```

**Review Comment**:
```markdown
**suggestion (blocking):** Violates Open/Closed Principle. Adding new payment methods requires modifying this function.

**Refactoring** (Strategy Pattern):

```python
from abc import ABC, abstractmethod

class PaymentProcessor(ABC):
    @abstractmethod
    def calculate_fee(self, amount: float) -> float:
        pass

class CreditCardProcessor(PaymentProcessor):
    def calculate_fee(self, amount: float) -> float:
        return amount * 0.029

class PayPalProcessor(PaymentProcessor):
    def calculate_fee(self, amount: float) -> float:
        return amount * 0.034

class BankTransferProcessor(PaymentProcessor):
    def calculate_fee(self, amount: float) -> float:
        return amount * 0.01

def process_payment(processor: PaymentProcessor, amount: float) -> float:
    return amount + processor.calculate_fee(amount)

# Usage
processor = CreditCardProcessor()
total = process_payment(processor, 100.00)
```

**Benefit**: New payment types can be added without modifying existing code.

**SOLID**: Open/Closed Principle

**Refs**: @TAG:PAYMENT-002
```

### Dependency Inversion Principle (DIP) Violation

**Before**:
```java
// ❌ ISSUE: High-level class depends on low-level implementation

public class OrderService {
    private MySQLDatabase database;  // Tight coupling!

    public OrderService() {
        this.database = new MySQLDatabase();  // Direct instantiation
    }

    public void saveOrder(Order order) {
        database.save(order);
    }
}
```

**Review Comment**:
```markdown
**issue (blocking):** Violates Dependency Inversion Principle.

**Problem**:
1. `OrderService` is tightly coupled to `MySQLDatabase`
2. Cannot switch to PostgreSQL/MongoDB without changing `OrderService`
3. Hard to test (cannot mock database)

**Solution** (Dependency Injection):

```java
public interface OrderRepository {
    void save(Order order);
}

public class MySQLOrderRepository implements OrderRepository {
    public void save(Order order) {
        // MySQL implementation
    }
}

public class OrderService {
    private final OrderRepository repository;

    // Dependency injected via constructor
    public OrderService(OrderRepository repository) {
        this.repository = repository;
    }

    public void saveOrder(Order order) {
        repository.save(order);
    }
}

// Usage (Spring/Guice framework or manual injection)
OrderRepository repo = new MySQLOrderRepository();
OrderService service = new OrderService(repo);
```

**Benefits**:
- Easy to swap database implementations
- Testable with mock repositories
- Follows SOLID principles

**SOLID**: Dependency Inversion Principle

**Refs**: @TAG:ORDER-003
```

---

## 3. Security Issues Detection

### SQL Injection Vulnerability

**Code**:
```python
# ❌ CRITICAL SECURITY ISSUE

def get_user_by_email(email):
    query = f"SELECT * FROM users WHERE email = '{email}'"  # Vulnerable!
    return db.execute(query)
```

**Review Comment**:
```markdown
**issue (blocking):** **CRITICAL SECURITY VULNERABILITY** - SQL Injection

**Attack Vector**:
```python
email = "admin@example.com' OR '1'='1"
# Results in: SELECT * FROM users WHERE email = 'admin@example.com' OR '1'='1'
# Returns ALL users!
```

**Fix** (Parameterized Query):
```python
def get_user_by_email(email: str):
    query = "SELECT * FROM users WHERE email = %s"
    return db.execute(query, (email,))  # Parameterized
```

**Reference**: OWASP A03:2021 – Injection

**Security Score**: 🔴 CRITICAL

**Refs**: @TAG:AUTH-005
```

### XSS (Cross-Site Scripting)

**Code**:
```javascript
// ❌ SECURITY: XSS vulnerability

function displayUserComment(comment) {
  document.getElementById('comment').innerHTML = comment;  // Dangerous!
}
```

**Review Comment**:
```markdown
**issue (blocking):** XSS vulnerability - unescaped user input rendered as HTML.

**Attack**:
```javascript
comment = "<script>alert(document.cookie)</script>"
// Executes malicious JavaScript!
```

**Fix**:
```javascript
function displayUserComment(comment) {
  // Option 1: Use textContent (safe)
  document.getElementById('comment').textContent = comment;

  // Option 2: Sanitize HTML (if HTML is needed)
  import DOMPurify from 'dompurify';
  document.getElementById('comment').innerHTML = DOMPurify.sanitize(comment);
}
```

**Reference**: OWASP A03:2021 – Injection (XSS)

**Refs**: @TAG:COMMENT-006
```

### Hardcoded Credentials

**Code**:
```javascript
// ❌ CRITICAL: Credentials in source code

const AWS_CONFIG = {
  accessKeyId: 'AKIAIOSFODNN7EXAMPLE',
  secretAccessKey: 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY',
  region: 'us-east-1'
};
```

**Review Comment**:
```markdown
**issue (blocking):** **CRITICAL** - AWS credentials hardcoded in source code.

**Risk**:
1. Credentials exposed in version control history
2. Anyone with repo access can use these credentials
3. Credentials may be accidentally committed to public repos

**Solution**:
```javascript
const AWS_CONFIG = {
  accessKeyId: process.env.AWS_ACCESS_KEY_ID,
  secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY,
  region: process.env.AWS_REGION || 'us-east-1'
};

// Add to .env (and .env.example without values)
// Add .env to .gitignore
```

**Immediate Action**:
1. **Rotate these credentials immediately** via AWS console
2. Remove from git history: `git filter-branch` or BFG Repo-Cleaner
3. Add `.env` to `.gitignore`

**Reference**: OWASP A07:2021 – Identification and Authentication Failures

**Security Score**: 🔴 CRITICAL
```

---

## 4. Performance Review Scenario

### N+1 Query Problem

**Code**:
```python
# ❌ PERFORMANCE: N+1 queries

def get_orders_with_items():
    orders = Order.objects.all()  # 1 query
    for order in orders:
        items = order.items.all()  # N queries (one per order!)
        order.items_list = list(items)
    return orders
```

**Review Comment**:
```markdown
**issue (blocking):** N+1 query problem will cause severe performance degradation.

**Problem**: For 1000 orders, this executes **1001 database queries** (1 + 1000).

**Performance Impact**:
- 1000 orders × 10ms per query = **10 seconds** (unacceptable!)

**Solution** (Eager Loading):
```python
def get_orders_with_items():
    # Single query with JOIN
    return Order.objects.prefetch_related('items').all()
```

**Result**: **1 query** instead of 1001 (100x faster).

**Verification**: Add `django.db.connection.queries` logging in tests.

**Performance**: 🔴 Critical

**Refs**: @TAG:ORDER-007
```

### Inefficient Algorithm

**Code**:
```javascript
// ❌ PERFORMANCE: O(n²) when O(n) exists

function findDuplicates(arr) {
  const duplicates = [];
  for (let i = 0; i < arr.length; i++) {
    for (let j = i + 1; j < arr.length; j++) {
      if (arr[i] === arr[j] && !duplicates.includes(arr[i])) {
        duplicates.push(arr[i]);
      }
    }
  }
  return duplicates;
}
```

**Review Comment**:
```markdown
**suggestion (non-blocking):** Inefficient O(n²) algorithm. Can be optimized to O(n).

**Performance**:
- Current: O(n²) - For 10,000 items: ~100 million operations
- Optimized: O(n) - For 10,000 items: ~10,000 operations

**Solution**:
```javascript
function findDuplicates(arr) {
  const seen = new Set();
  const duplicates = new Set();

  for (const item of arr) {
    if (seen.has(item)) {
      duplicates.add(item);
    } else {
      seen.add(item);
    }
  }

  return Array.from(duplicates);
}
```

**Complexity**: O(n) time, O(n) space

**Refs**: @TAG:UTIL-008
```

---

## 5. API Design Review

### RESTful API Issues

**Code**:
```python
# ❌ API DESIGN ISSUES

@app.route('/get_user', methods=['POST'])  # Issue 1
def get_user():
    user_id = request.json['id']  # Issue 2
    user = User.query.get(user_id)
    if user:
        return jsonify(user.to_dict())  # Issue 3
    return jsonify({'error': 'Not found'})  # Issue 4
```

**Review Comment**:
```markdown
**suggestion (blocking):** Multiple REST API design issues.

**Issue 1**: Use GET method, not POST for retrieval
**Issue 2**: Missing input validation and error handling
**Issue 3**: Missing HTTP status codes
**Issue 4**: Inconsistent error response format

**Improved Design**:
```python
@app.route('/api/v1/users/<int:user_id>', methods=['GET'])
def get_user(user_id: int):
    """Get user by ID."""
    if user_id <= 0:
        return jsonify({'error': 'Invalid user ID'}), 400

    user = User.query.get(user_id)
    if not user:
        return jsonify({'error': 'User not found'}), 404

    return jsonify({
        'data': user.to_dict(),
        'meta': {'version': 'v1'}
    }), 200
```

**Improvements**:
1. ✅ RESTful URL structure (`/api/v1/users/:id`)
2. ✅ Correct HTTP method (GET)
3. ✅ URL parameter instead of JSON body
4. ✅ Proper HTTP status codes (200, 400, 404)
5. ✅ Consistent response structure
6. ✅ API versioning (`/api/v1/`)
7. ✅ Input validation

**Reference**: RESTful API Design Best Practices

**Refs**: @TAG:API-009
```

---

## 6. Test Quality Review

### Insufficient Test Coverage

**Code**:
```python
# src/calculator.py
def divide(a, b):
    return a / b

# tests/test_calculator.py
def test_divide():
    assert divide(10, 2) == 5  # Only happy path!
```

**Review Comment**:
```markdown
**issue (blocking):** Test coverage insufficient - missing edge cases.

**Missing Test Cases**:
1. Division by zero
2. Negative numbers
3. Floating point precision
4. Type validation (non-numeric inputs)

**Required Tests**:
```python
import pytest

def test_divide_positive_numbers():
    assert divide(10, 2) == 5

def test_divide_negative_numbers():
    assert divide(-10, 2) == -5
    assert divide(10, -2) == -5

def test_divide_by_zero():
    with pytest.raises(ZeroDivisionError):
        divide(10, 0)

def test_divide_floating_point():
    assert abs(divide(10, 3) - 3.333333) < 0.0001

def test_divide_invalid_input():
    with pytest.raises(TypeError):
        divide("10", 2)
```

**Coverage Target**: ≥85% (currently ~20%)

**Refs**: @TAG:CALC-010, TRUST Principle (Test First)
```

---

## 7. Documentation Review

### Missing Documentation

**Code**:
```typescript
// ❌ No documentation

export async function processData(data: any, options: any): Promise<any> {
  // Complex logic here...
}
```

**Review Comment**:
```markdown
**nitpick (non-blocking):** Missing function documentation and type definitions.

**Improved**:
```typescript
/**
 * Process user data according to specified options.
 *
 * @param data - Raw user data from API
 * @param options - Processing options
 * @param options.format - Output format ('json' | 'csv')
 * @param options.validate - Whether to validate data (default: true)
 * @returns Processed data in specified format
 * @throws {ValidationError} If data validation fails
 *
 * @example
 * ```typescript
 * const result = await processData(
 *   { name: 'John', age: 30 },
 *   { format: 'json', validate: true }
 * );
 * ```
 */
export async function processData(
  data: UserData,
  options: ProcessOptions
): Promise<ProcessedData> {
  // Implementation...
}

interface UserData {
  name: string;
  age: number;
}

interface ProcessOptions {
  format: 'json' | 'csv';
  validate?: boolean;
}

interface ProcessedData {
  // ...
}
```

**Benefits**:
- Clear API contract
- IDE autocomplete support
- Example usage
- Error documentation

**Refs**: @TAG:DOC-011
```

---

## 8. Automated Review Tool Integration

### SonarQube Review

**SonarQube Report**:
```
Code Smells: 15
Bugs: 3
Vulnerabilities: 2
Code Coverage: 68% (Target: 85%)
Duplicated Code: 8.2% (Target: <3%)
```

**Review Comment**:
```markdown
**issue (blocking):** SonarQube quality gate failed.

**Critical Issues**:
1. **Security Hotspot**: SQL injection in `UserRepository.findByEmail()` (🔴 Critical)
2. **Bug**: Potential NullPointerException in `OrderService.calculateTotal()` (🟠 Major)
3. **Code Coverage**: 68% < 85% target (🟡 Minor)

**Action Items**:
1. Fix SQL injection with parameterized queries
2. Add null checks or use Optional<T>
3. Add test cases to reach 85% coverage
4. Refactor duplicated code blocks

**SonarQube Report**: [View Report](https://sonarqube.example.com/project)

**Refs**: TRUST Principle (Test Coverage ≥85%)
```

### ESLint Review

**ESLint Report**:
```
src/utils/validator.ts
  15:3   error    'email' is defined but never used           no-unused-vars
  23:5   warning  Unexpected console statement                no-console
  42:10  error    Expected '===' and instead saw '=='         eqeqeq
```

**Review Comment**:
```markdown
**nitpick (non-blocking):** ESLint errors must be resolved.

**Fixes**:
```typescript
// Line 15: Remove unused variable
- const email = user.email;

// Line 23: Replace console with proper logging
- console.log('Validating user:', user);
+ logger.info('Validating user', { userId: user.id });

// Line 42: Use strict equality
- if (user.status == 'active')
+ if (user.status === 'active')
```

**Alternative**: If console.log is intentional (debugging), add:
```typescript
// eslint-disable-next-line no-console
console.log('Debug:', data);
```

**Linting**: All ESLint errors must be resolved before merge.
```

---

## 9. Pre-commit Hook Review Workflow

### Git Hook Example

```bash
# .git/hooks/pre-commit (or use Husky/pre-commit framework)

#!/bin/bash

echo "Running pre-commit checks..."

# 1. Run linter
echo "🔍 Running ESLint..."
npm run lint
if [ $? -ne 0 ]; then
  echo "❌ ESLint failed. Please fix linting errors."
  exit 1
fi

# 2. Run formatter check
echo "🎨 Checking code formatting..."
npm run format:check
if [ $? -ne 0 ]; then
  echo "❌ Code formatting issues. Run: npm run format"
  exit 1
fi

# 3. Run tests
echo "🧪 Running tests..."
npm test
if [ $? -ne 0 ]; then
  echo "❌ Tests failed. Please fix failing tests."
  exit 1
fi

# 4. Check for secrets
echo "🔒 Scanning for secrets..."
gitleaks detect --no-git --verbose
if [ $? -ne 0 ]; then
  echo "❌ Potential secrets detected. Please review."
  exit 1
fi

echo "✅ All pre-commit checks passed!"
exit 0
```

### Review Comment Integration

```markdown
**suggestion (non-blocking):** Add pre-commit hooks to enforce quality standards.

**Setup** (using Husky):
```bash
npm install --save-dev husky lint-staged

# package.json
{
  "husky": {
    "hooks": {
      "pre-commit": "lint-staged"
    }
  },
  "lint-staged": {
    "*.{ts,tsx,js,jsx}": [
      "eslint --fix",
      "prettier --write",
      "jest --findRelatedTests --passWithNoTests"
    ]
  }
}
```

**Benefits**:
- Catches issues before code review
- Reduces review iteration cycles
- Enforces consistent standards

**Refs**: @TAG:INFRA-012
```

---

## 10. GitLab/GitHub Review Best Practices

### Effective Review Comments

#### ✅ Good Examples

**Specific and Actionable**:
```markdown
**suggestion:** Extract magic number to named constant.

```python
- TIMEOUT = 30  # What does 30 represent?
+ REQUEST_TIMEOUT_SECONDS = 30
+ RETRY_ATTEMPTS = 3
```

**Explanation**: Makes code self-documenting and easier to modify.
```

**Constructive Feedback**:
```markdown
**praise:** Great use of dependency injection here! This makes testing much easier.

**suggestion (minor):** Consider adding a docstring to explain the factory pattern for future maintainers.
```

**Ask Questions**:
```markdown
**question:** Have you considered using a connection pool here? With high traffic, creating new connections per request might be a bottleneck.

Not blocking, but worth discussing.
```

#### ❌ Bad Examples

**Vague**:
```markdown
This looks wrong.
```

**Demanding**:
```markdown
Change this immediately. This is terrible code.
```

**Unhelpful**:
```markdown
Doesn't work for me. 🤷
```

### Conventional Comment Format

Use structured comment prefixes:

| Prefix | Meaning | Blocking |
|--------|---------|----------|
| **praise** | Positive feedback | No |
| **nitpick** | Minor style/preference | Usually No |
| **suggestion** | Improvement idea | Maybe |
| **issue** | Actual problem | Usually Yes |
| **question** | Clarification needed | No |
| **thought** | General observation | No |
| **chore** | Maintenance task | Maybe |

**Example**:
```markdown
**suggestion (non-blocking):** Consider using async/await for better readability.

```typescript
// Instead of promises chains
getUser(id).then(user => getUserOrders(user.id)).then(orders => ...)

// Use async/await
const user = await getUser(id);
const orders = await getUserOrders(user.id);
```

**Refs**: @TAG:API-013
```

### Review Checklist

**Before Submitting Review**:

- [ ] Understand the PR context and requirements
- [ ] Check that all tests pass
- [ ] Verify code coverage meets 85% threshold
- [ ] Review SOLID principles compliance
- [ ] Check for security vulnerabilities
- [ ] Assess performance implications
- [ ] Verify API design follows REST standards
- [ ] Confirm documentation is adequate
- [ ] Check that @TAG references are updated
- [ ] Verify SPEC alignment (if applicable)
- [ ] Run automated tools (linters, SonarQube)
- [ ] Test locally if complex changes

**When Approving**:

- [ ] All blocking comments resolved
- [ ] CI/CD pipeline green
- [ ] No merge conflicts
- [ ] TRUST principles satisfied

**When Requesting Changes**:

- [ ] Clear, actionable feedback provided
- [ ] Blocking vs non-blocking clearly marked
- [ ] Constructive tone maintained
- [ ] Examples or alternatives provided

---

## Integration with MoAI-ADK

### Review Workflow with Alfred

```bash
# After implementation complete
/alfred:3-sync auto

# Review automation triggers:
# 1. Automated linting (ESLint/Ruff/Clippy)
# 2. SonarQube scan
# 3. Test coverage report
# 4. @TAG integrity check
# 5. Living docs update

# Manual review checklist enforced:
# - SOLID principles compliance
# - Security scan (OWASP checks)
# - Performance benchmarks
# - API design standards
```

### Quality Gate Enforcement

**MoAI-ADK TRUST Principles**:

1. **T**est Coverage: ≥85% required
2. **R**eadable: Linter passes, complexity ≤10
3. **U**nified: Type-safe or validated
4. **S**ecured: No security warnings
5. **T**rackable: @TAG references current

**Review Status**:
```
✅ Test Coverage: 92% (Target: 85%)
✅ Complexity: Max 8 (Target: ≤10)
✅ Type Safety: TypeScript strict mode
⚠️ Security: 1 medium issue (needs fix)
✅ TAG Coverage: 100%

Overall: ⚠️ Changes Requested
```

---

## References

- [Google Engineering Practices: Code Review](https://google.github.io/eng-practices/review/)
- [GitLab Code Review Guidelines](https://docs.gitlab.com/ee/development/code_review.html)
- [Conventional Comments](https://conventionalcomments.org/)
- [OWASP Top 10 (2021)](https://owasp.org/Top10/)
- [SOLID Principles](https://en.wikipedia.org/wiki/SOLID)
- [The Upsource Blog: SOLID Principles](https://blog.jetbrains.com/upsource/tag/code-review/)

---

**Version**: 2.0.0
**Last Updated**: 2025-10-22
**Part of**: MoAI-ADK Essentials Skills
**Related Skills**: `moai-essentials-refactor`, `moai-foundation-trust`, `moai-alfred-code-reviewer`
