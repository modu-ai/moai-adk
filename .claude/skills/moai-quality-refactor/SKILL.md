---
name: moai-quality-refactor
description: "Refactoring consolidated: patterns, SOLID improvements, code smell elimination, design patterns"
version: 1.0.0
modularized: true
last_updated: 2025-11-24
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
  - Read
  - Write
  - Edit
compliance_score: 85
modules:
  - refactoring-patterns
  - code-smells
  - design-patterns
dependencies:
  - moai-foundation-trust
  - moai-quality-review
deprecated: false
successor: null
category_tier: 3
auto_trigger_keywords:
  - refactor
  - refactoring
  - code-smell
  - design-pattern
  - cleanup
  - improve
  - extract
  - dry
  - duplication
  - simplify
  - maintainability
agent_coverage:
  - code-refactorer
  - quality-gate
context7_references: []
invocation_api_version: "1.0"
---

## Quick Reference (30 seconds)

**Enterprise Refactoring Consolidated**

Unified refactoring framework consolidating essentials-refactor and mcp-integration with systematic refactoring patterns, code smell elimination, design pattern application, and SOLID principles improvement.

**Core Capabilities**:
- ✅ Systematic refactoring (10+ patterns)
- ✅ Code smell detection and elimination
- ✅ Design pattern application
- ✅ SOLID principles improvement
- ✅ Function extraction and consolidation
- ✅ Automated refactoring with IDE support
- ✅ Refactoring-safe testing

**When to Use**:
- Improving code readability
- Eliminating code duplication
- Extracting complex functions
- Improving SOLID compliance
- Applying design patterns
- Reducing code smell
- Maintaining code quality

**Core Framework**: IDENTIFY → PLAN → EXECUTE → VALIDATE
```
1. Code Smell Detection
   ↓
2. Refactoring Pattern Selection
   ↓
3. Systematic Refactoring
   ↓
4. Test Verification
   ↓
5. Code Review Approval
```

---

## Core Patterns (5-10 minutes each)

### Pattern 1: Extract Method Refactoring

**Concept**: Break large functions into smaller, focused methods.

```python
# ❌ BEFORE: Long, complex function
def process_user_registration(user_data):
    """Process user registration (hard to test, many concerns)."""
    # Validation logic (mixed concern)
    if not user_data.get('email'):
        raise ValueError("Email required")
    if not user_data.get('password'):
        raise ValueError("Password required")
    if len(user_data['password']) < 8:
        raise ValueError("Password too short")

    # Database logic (mixed concern)
    existing = User.query.filter_by(email=user_data['email']).first()
    if existing:
        raise ValueError("Email already registered")

    # Hashing logic (mixed concern)
    import hashlib
    hashed = hashlib.sha256(user_data['password'].encode()).hexdigest()

    # Save logic (mixed concern)
    user = User(email=user_data['email'], password_hash=hashed)
    db.session.add(user)
    db.session.commit()

    # Email logic (mixed concern)
    import smtplib
    send_email(user.email, "Welcome!")

    return user

# ✅ AFTER: Extracted methods (single responsibility)
def process_user_registration(user_data):
    """Orchestrate user registration (single concern)."""
    validate_user_input(user_data)
    check_email_not_registered(user_data['email'])
    user = create_user(user_data)
    send_welcome_email(user.email)
    return user

def validate_user_input(user_data):
    """Validate user input."""
    if not user_data.get('email'):
        raise ValueError("Email required")
    if not user_data.get('password'):
        raise ValueError("Password required")
    if len(user_data['password']) < 8:
        raise ValueError("Password too short")

def check_email_not_registered(email: str):
    """Check email not already registered."""
    existing = User.query.filter_by(email=email).first()
    if existing:
        raise ValueError("Email already registered")

def create_user(user_data):
    """Create user in database."""
    hashed = hash_password(user_data['password'])
    user = User(email=user_data['email'], password_hash=hashed)
    db.session.add(user)
    db.session.commit()
    return user

def hash_password(password: str) -> str:
    """Hash password securely."""
    import hashlib
    return hashlib.sha256(password.encode()).hexdigest()

def send_welcome_email(email: str):
    """Send welcome email."""
    import smtplib
    # Implementation...
```

**Use Case**: Improve testability and readability.

---

### Pattern 2: Replace Magic Numbers with Constants

**Concept**: Use named constants instead of magic numbers.

```python
# ❌ BEFORE: Magic numbers
def calculate_discount(price, customer_type):
    if customer_type == 1:  # What is 1?
        return price * 0.1  # What does 0.1 mean?
    elif customer_type == 2:
        return price * 0.2
    elif customer_type == 3:
        return price * 0.3
    return 0

# ✅ AFTER: Named constants
class CustomerType:
    REGULAR = 1
    SILVER = 2
    GOLD = 3

CUSTOMER_DISCOUNTS = {
    CustomerType.REGULAR: 0.10,  # 10% discount
    CustomerType.SILVER: 0.20,   # 20% discount
    CustomerType.GOLD: 0.30,     # 30% discount
}

def calculate_discount(price: float, customer_type: int) -> float:
    """Calculate discount based on customer type."""
    return price * CUSTOMER_DISCOUNTS.get(customer_type, 0)
```

**Use Case**: Improve code readability and maintainability.

---

### Pattern 3: Replace Conditionals with Polymorphism

**Concept**: Use polymorphism instead of conditional logic.

```python
# ❌ BEFORE: Conditional logic
def calculate_shipping(order_type):
    if order_type == "standard":
        return order_type * 10
    elif order_type == "express":
        return order_type * 20
    elif order_type == "overnight":
        return order_type * 50

# ✅ AFTER: Polymorphism
from abc import ABC, abstractmethod

class ShippingStrategy(ABC):
    @abstractmethod
    def calculate(self, order_weight):
        pass

class StandardShipping(ShippingStrategy):
    def calculate(self, order_weight):
        return order_weight * 10

class ExpressShipping(ShippingStrategy):
    def calculate(self, order_weight):
        return order_weight * 20

class OvernightShipping(ShippingStrategy):
    def calculate(self, order_weight):
        return order_weight * 50

# Usage
shipping_strategy = ExpressShipping()
cost = shipping_strategy.calculate(5)  # 5 * 20 = 100
```

**Use Case**: Reduce conditional complexity, improve extensibility.

---

### Pattern 4: Rename for Clarity

**Concept**: Rename unclear variables/functions to clarify intent.

```python
# ❌ BEFORE: Unclear names
def process(d):
    """Process data."""
    r = {}
    for i in d:
        r[i['id']] = i['val'] * 2
    return r

# ✅ AFTER: Clear names
def calculate_doubled_values(items):
    """Calculate doubled values by item ID."""
    result = {}
    for item in items:
        item_id = item['id']
        value = item['value']
        result[item_id] = value * 2
    return result

# Or with comprehension
def calculate_doubled_values(items):
    """Calculate doubled values by item ID."""
    return {item['id']: item['value'] * 2 for item in items}
```

**Use Case**: Improve code readability.

---

### Pattern 5: DRY Principle - Eliminate Duplication

**Concept**: Consolidate duplicated code.

```python
# ❌ BEFORE: Duplicated validation
def create_user(username, email, phone):
    if not username or len(username) < 3:
        raise ValueError("Invalid username")
    if not email or '@' not in email:
        raise ValueError("Invalid email")
    if not phone or len(phone) < 10:
        raise ValueError("Invalid phone")
    # ... create user ...

def update_user(username, email, phone):
    if not username or len(username) < 3:
        raise ValueError("Invalid username")
    if not email or '@' not in email:
        raise ValueError("Invalid email")
    if not phone or len(phone) < 10:
        raise ValueError("Invalid phone")
    # ... update user ...

# ✅ AFTER: Shared validation
def validate_user_fields(username, email, phone):
    """Validate user input fields."""
    if not username or len(username) < 3:
        raise ValueError("Invalid username")
    if not email or '@' not in email:
        raise ValueError("Invalid email")
    if not phone or len(phone) < 10:
        raise ValueError("Invalid phone")

def create_user(username, email, phone):
    validate_user_fields(username, email, phone)
    # ... create user ...

def update_user(username, email, phone):
    validate_user_fields(username, email, phone)
    # ... update user ...
```

**Use Case**: Reduce maintenance burden and bugs.

---

## Advanced Documentation

For detailed refactoring patterns:

- **[modules/refactoring-patterns.md](modules/refactoring-patterns.md)** - 15+ refactoring patterns
- **[modules/code-smells.md](modules/code-smells.md)** - Code smell detection and fixes
- **[modules/design-patterns.md](modules/design-patterns.md)** - Design pattern application

---

## Best Practices

### ✅ DO
- Refactor in small, testable steps
- Keep tests passing during refactoring
- Use IDE refactoring tools (safer)
- Review refactoring changes
- Test before/after performance
- Document why refactoring was needed
- Combine with code review

### ❌ DON'T
- Refactor and add features simultaneously
- Skip tests during refactoring
- Refactor code you don't understand
- Make large changes at once
- Ignore performance impacts
- Refactor unless code needs it
- Commit refactoring with features

---

**Status**: Production Ready
**Generated with**: MoAI-ADK Skill Factory
