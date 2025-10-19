---
name: moai-essentials-review
description: Automated code review with SOLID principles, code smells, and language-specific best practices
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
tier: 1
auto-load: "true"
---

# Alfred Code Reviewer

## What it does

Automated code review with language-specific best practices, SOLID principles verification, and code smell detection.

## When to use

- "코드 리뷰해줘", "이 코드 개선점은?", "코드 품질 확인", "리뷰 부탁해", "문제점 찾아줘"
- "SOLID 원칙", "베스트 프랙티스", "코드 스멜", "안티패턴", "보안 취약점"
- "Code review", "Quality check", "Best practices", "Security audit"
- Optionally invoked after `/alfred:3-sync`
- Before merging PR or releasing
- During peer code review

## How it works

**Code Constraints Check**:
- File ≤300 LOC
- Function ≤50 LOC
- Parameters ≤5
- Cyclomatic complexity ≤10

**SOLID Principles**:
- Single Responsibility
- Open/Closed
- Liskov Substitution
- Interface Segregation
- Dependency Inversion

**Code Smell Detection**:
- Long Method
- Large Class
- Duplicate Code
- Dead Code
- Magic Numbers

**Language-specific Best Practices**:
- Python: List comprehension, type hints, PEP 8
- TypeScript: Strict typing, async/await, error handling
- Java: Streams API, Optional, Design patterns

## Review Checklist

### Code Constraints
```bash
# File LOC check (≤300)
find src/ -name "*.py" | xargs wc -l | awk '$1 > 300'

# Function LOC check (≤50)
radon cc src/ -s -n D  # Show functions >50 LOC

# Complexity check (≤10)
radon cc src/ -s -n C  # Show complexity ≥10

# Parameter count (≤5)
rg "def \w+\([^)]*," src/ | awk -F, 'NF > 5'
```

### SOLID Principles

**S - Single Responsibility Principle**:
```python
# ❌ Violation: Class does too much
class User:
    def save_to_db(self):
        # Database logic
    def send_email(self):
        # Email logic
    def validate(self):
        # Validation logic

# ✅ Correct: Each class has one responsibility
class User:
    def validate(self):
        # Validation only

class UserRepository:
    def save(self, user):
        # Database only

class EmailService:
    def send(self, user):
        # Email only
```

**O - Open/Closed Principle**:
```python
# ❌ Violation: Must modify code to add new type
def calculate_area(shape):
    if shape.type == "circle":
        return 3.14 * shape.radius ** 2
    elif shape.type == "square":
        return shape.side ** 2

# ✅ Correct: Extend without modifying
class Shape:
    def area(self):
        raise NotImplementedError

class Circle(Shape):
    def area(self):
        return 3.14 * self.radius ** 2

class Square(Shape):
    def area(self):
        return self.side ** 2
```

**L - Liskov Substitution Principle**:
```python
# ❌ Violation: Derived class breaks base contract
class Bird:
    def fly(self):
        pass

class Penguin(Bird):  # Penguins can't fly!
    def fly(self):
        raise Exception("Can't fly")

# ✅ Correct: Correct abstraction
class Bird:
    pass

class FlyingBird(Bird):
    def fly(self):
        pass

class Penguin(Bird):  # No fly method
    pass
```

**I - Interface Segregation Principle**:
```python
# ❌ Violation: Fat interface
class Worker:
    def work(self):
        pass
    def eat(self):
        pass

class Robot(Worker):  # Robots don't eat!
    def eat(self):
        raise NotImplementedError

# ✅ Correct: Segregated interfaces
class Workable:
    def work(self):
        pass

class Eatable:
    def eat(self):
        pass

class Human(Workable, Eatable):
    pass

class Robot(Workable):  # Only what it needs
    pass
```

**D - Dependency Inversion Principle**:
```python
# ❌ Violation: High-level depends on low-level
class EmailService:
    def send(self, message):
        # Concrete implementation

class NotificationService:
    def __init__(self):
        self.email = EmailService()  # Tight coupling

# ✅ Correct: Depend on abstractions
class MessageSender:  # Abstract
    def send(self, message):
        raise NotImplementedError

class EmailService(MessageSender):
    def send(self, message):
        # Implementation

class NotificationService:
    def __init__(self, sender: MessageSender):
        self.sender = sender  # Dependency injection
```

### Code Smell Detection

**Dead Code**:
```bash
# Find unused imports
rg "^import \w+" src/ | sort | uniq -d

# Find unused variables
ruff check src/ --select F841
```

**Magic Numbers**:
```bash
# Find hardcoded numbers (excluding 0, 1)
rg "\b[2-9][0-9]*\b" src/ -n
```

**Duplicate Code**:
```bash
# Find similar code blocks
jscpd src/ --min-lines 5
```

## Review Report Example

```markdown
╔══════════════════════════════════════════════════════════╗
║              Code Review Report                          ║
╠══════════════════════════════════════════════════════════╣
║ 🔴 Critical Issues (3)                                   ║
╠══════════════════════════════════════════════════════════╣
║ 1. src/auth/service.py:45                                ║
║    Function too long: 85 LOC (limit: 50)                 ║
║    → Extract methods: validate(), hash_password()        ║
╠══════════════════════════════════════════════════════════╣
║ 2. src/api/handler.ts:120                                ║
║    Missing error handling for async operation            ║
║    → Add try-catch block                                 ║
╠══════════════════════════════════════════════════════════╣
║ 3. src/db/repository.java:200                            ║
║    Magic number: 86400                                   ║
║    → Extract constant: SECONDS_IN_DAY                    ║
╠══════════════════════════════════════════════════════════╣
║ ⚠️  Warnings (5)                                         ║
╠══════════════════════════════════════════════════════════╣
║ 1. src/utils/helper.py:30 - Unused import               ║
║ 2. src/models/user.ts:15 - Weak type (any)              ║
║ 3. src/services/auth.go:88 - Deep nesting (6 levels)    ║
║ 4. src/core/processor.rs:142 - Unwrap on Result         ║
║ 5. src/lib/validator.java:55 - Empty catch block        ║
╠══════════════════════════════════════════════════════════╣
║ ✅ Good Practices Found                                  ║
╠══════════════════════════════════════════════════════════╣
║ • Test coverage: 92% (>85%)                              ║
║ • Consistent naming convention                           ║
║ • Type hints used throughout                             ║
║ • Guard clauses for early returns                        ║
║ • No hardcoded secrets                                   ║
╠══════════════════════════════════════════════════════════╣
║ Overall Score: 7.5/10 (Good)                             ║
║ Recommendation: Fix 3 critical issues before merge       ║
╚══════════════════════════════════════════════════════════╝
```

## Examples

### Example 1: SOLID Violation Detection
User: "이 코드 리뷰해줘"

Alfred detects:
```python
# SRP Violation: Class does too much
class UserManager:
    def create_user(self):
        # User creation
    def send_welcome_email(self):
        # Email sending
    def log_activity(self):
        # Logging

Recommendation:
→ Split into: User, EmailService, ActivityLogger
```

### Example 2: Security Issue Detection
User: "보안 취약점 찾아줘"

Alfred finds:
```python
# ❌ SQL Injection vulnerability
query = f"SELECT * FROM users WHERE id = {user_id}"

# ❌ Hardcoded secret
API_KEY = "sk-1234567890abcdef"

# ❌ No input validation
def transfer_money(amount):
    # No check if amount > 0

Fix:
1. Use parameterized queries
2. Use environment variables
3. Add input validation
```

### Example 3: Language-specific Best Practices
User: "TypeScript 베스트 프랙티스 확인"

Alfred reviews:
```typescript
// ⚠️ Using 'any' (weak typing)
function process(data: any) { }

// ⚠️ No error handling
async function fetchData() {
    const res = await fetch(url);
    return res.json();
}

Recommendations:
1. Replace 'any' with specific type
2. Add try-catch for async operations
3. Use strict TypeScript config
```

## Works well with

- moai-foundation-specs (Review against SPEC requirements)
- moai-essentials-refactor (Apply refactoring suggestions)
- moai-foundation-trust (Validate TRUST principles)
