---
name: moai-core-todowrite-pattern
version: 4.0.0
updated: 2025-11-20
status: stable
tier: core
description: TodoWrite pattern for incremental implementation and placeholder management
allowed-tools: [Read, Write, TodoWrite]
---

# TodoWrite Pattern Expert

**Incremental Development with Placeholders**

> **Pattern**: TodoWrite → Implement → Replace  
> **Tools**: TodoWrite tool, TODO comments

---

## Overview

Strategy for breaking complex tasks into manageable pieces using placeholders.

### Core Concept

Instead of implementing everything at once:

1.  **Plan**: Identify components needed
2.  **Scaffold**: Create structure with TodoWrite placeholders
3.  **Implement**: Fill in placeholders incrementally
4.  **Validate**: Ensure completeness

---

## Usage Pattern

### 1. Initial Scaffolding

**Example: Building an API**:

```python
# Step 1: Create structure with placeholders
class UserAPI:
    def __init__(self):
        TodoWrite("Initialize database connection")

    def create_user(self, email: str, password: str):
        TodoWrite("Validate input")
        TodoWrite("Hash password")
        TodoWrite("Save to database")
        TodoWrite("Return user object")

    def get_user(self, user_id: int):
        TodoWrite("Query database")
        TodoWrite("Handle user not found")
```

### 2. Incremental Implementation

**Round 1: Critical Path**:

```python
def create_user(self, email: str, password: str):
    # ✅ Implemented
    if not email or not password:
        raise ValueError("Email and password required")

    TodoWrite("Hash password")  # Still placeholder
    TodoWrite("Save to database")
    TodoWrite("Return user object")
```

**Round 2: Security**:

```python
def create_user(self, email: str, password: str):
    if not email or not password:
        raise ValueError("Email and password required")

    # ✅ Implemented
    import bcrypt
    hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt())

    TodoWrite("Save to database")
    TodoWrite("Return user object")
```

### 3. Completion Validation

**Check for remaining TODOs**:

```bash
# Find all TodoWrite calls
grep -r "TodoWrite" src/

# Or use tool
rg "TodoWrite" --type python
```

---

## Best Practices

1.  **Descriptive Placeholders**: "TodoWrite('Validate email format')" not "TodoWrite('TODO')".
2.  **Logical Grouping**: Group related TODOs.
3.  **Priority Order**: Implement critical paths first.
4.  **Track Progress**: Maintain list of remaining TODOs.

### When to Use TodoWrite

- Complex feature with clear structure
- Uncertain implementation details
- Team collaboration (assign TODO owners)
- Incremental delivery (MVP → full feature)

### When NOT to Use

- Simple functions (<20 lines)
- Critical security code (implement immediately)
- One-time scripts

---

## Example: Complete Flow

```python
# Initial (90% TodoWrite)
class PaymentProcessor:
    def process_payment(self, amount, card):
        TodoWrite("Validate amount")
        TodoWrite("Validate card")
        TodoWrite("Call payment gateway")
        TodoWrite("Handle success")
        TodoWrite("Handle failure")
        TodoWrite("Log transaction")

# After Round 1 (50% done)
class PaymentProcessor:
    def process_payment(self, amount, card):
        # ✅ Done
        if amount <= 0:
            raise ValueError("Invalid amount")

        # ✅ Done
        if not card.validate():
            raise ValueError("Invalid card")

        TodoWrite("Call payment gateway")
        TodoWrite("Handle success")
        TodoWrite("Handle failure")
        TodoWrite("Log transaction")

# Final (100% done)
class PaymentProcessor:
    def process_payment(self, amount, card):
        if amount <=  0:
            raise ValueError("Invalid amount")

        if not card.validate():
            raise ValueError("Invalid card")

        # ✅ All implemented
        result = self.gateway.charge(amount, card)

        if result.success:
            self.logger.info(f"Payment success: {result.id}")
            return result
        else:
            self.logger.error(f"Payment failed: {result.error}")
            raise PaymentError(result.error)
```

---

## Validation Checklist

- [ ] **Placeholders**: Descriptive TodoWrite messages?
- [ ] **Priority**: Critical paths implemented first?
- [ ] **Tracking**: Remaining TODOs documented?
- [ ] **Completion**: All TODOs resolved before merge?

---

## Related Skills

- `moai-core-rules`: Development workflow
- `moai-foundation-specs`: Requirement breakdown

---

**Last Updated**: 2025-11-20
