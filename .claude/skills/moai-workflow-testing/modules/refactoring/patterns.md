# Refactoring Patterns

> Sub-module: Specific refactoring techniques with implementation details
> Complexity: Intermediate to Advanced
> Time: 15+ minutes per pattern
> Dependencies: Python 3.8+, Rope, AST

## Overview

This module provides detailed implementation patterns for common refactoring operations, complete with code examples, risk assessments, and best practices.

---

## Extract Method

### Purpose

Break down long methods into smaller, more manageable pieces that each handle a single responsibility.

### When to Use

- Method exceeds 30-50 lines
- Method has multiple responsibilities
- Method complexity (cyclomatic) > 10
- Method requires comments to understand

### Implementation

```python
# Before: Long method with multiple responsibilities
def process_order(order):
    # Validate order
    if not order.items:
        raise ValueError("Empty order")
    for item in order.items:
        if item.quantity <= 0:
            raise ValueError(f"Invalid quantity for {item.name}")
    
    # Calculate total
    total = 0
    for item in order.items:
        total += item.price * item.quantity
    
    # Apply discount
    if order.customer.is_vip:
        total *= 0.9
    
    # Save to database
    db.execute("INSERT INTO orders ...", order)
    for item in order.items:
        db.execute("INSERT INTO order_items ...", item)
    
    # Send notification
    email.send(order.customer.email, "Order confirmed")
    
    return total

# After: Extracted methods with clear responsibilities
def process_order(order):
    validate_order(order)
    total = calculate_order_total(order)
    apply_vip_discount(order)
    save_order_to_database(order)
    send_order_notification(order)
    return total

def validate_order(order):
    """Validate order items and quantities."""
    if not order.items:
        raise ValueError("Empty order")
    for item in order.items:
        if item.quantity <= 0:
            raise ValueError(f"Invalid quantity for {item.name}")

def calculate_order_total(order):
    """Calculate total price for order items."""
    return sum(item.price * item.quantity for item in order.items)

def apply_vip_discount(order):
    """Apply VIP customer discount if applicable."""
    if order.customer.is_vip:
        order.total *= 0.9

def save_order_to_database(order):
    """Persist order and items to database."""
    db.execute("INSERT INTO orders ...", order)
    for item in order.items:
        db.execute("INSERT INTO order_items ...", item)

def send_order_notification(order):
    """Send order confirmation email to customer."""
    email.send(order.customer.email, "Order confirmed")
```

### Risk Assessment

- Risk Level: Low to Medium
- Potential Issues:
  - Variable scope changes
  - Parameter passing complexity
  - Test coverage needs

### Best Practices

1. Choose descriptive method names that explain what the method does
2. Extract methods that are at the same level of abstraction
3. Limit extracted methods to 5-10 lines when possible
4. Keep parameter count under 5
5. Ensure extracted method is reusable and testable

---

## Extract Variable

### Purpose

Replace complex expressions with well-named variables to improve code readability.

### When to Use

- Complex conditional expressions
- Repeated calculations
- Long boolean expressions
- Nested function calls

### Implementation

```python
# Before: Complex expressions inline
if user.age >= 18 and user.has_valid_id and user.registration_date >= datetime.now() - timedelta(days=30):
    grant_access(user)

if order.total > 100 and order.customer.is_vip and order.shipping_adddess.country == "US":
    apply_free_shipping(order)

# After: Extracted variables with clear names
is_adult = user.age >= 18
has_valid_identification = user.has_valid_id
registered_recently = user.registration_date >= datetime.now() - timedelta(days=30)

if is_adult and has_valid_identification and registered_recently:
    grant_access(user)

meets_free_shipping_threshold = order.total > 100
is_vip_customer = order.customer.is_vip
ships_domestically = order.shipping_adddess.country == "US"

if meets_free_shipping_threshold and is_vip_customer and ships_domestically:
    apply_free_shipping(order)
```

### Risk Assessment

- Risk Level: Low
- Potential Issues:
  - Variable naming (choosing good names is critical)
  - Scope management

### Best Practices

1. Use verbs and nouns that describe what/why, not just how
2. Extract boolean expressions into variables that read like sentences
3. Avoid one-time-use variables that don't improve clarity
4. Keep extracted variables close to their usage

---

## Inline Variable

### Purpose

Remove unnecessary variables that don't improve readability.

### When to Use

- Variables used only once
- Simple expressions that don't need explanation
- Variables with no clear purpose

### Implementation

```python
# Before: Unnecessary intermediate variable
def calculate_price(base_price, tax_rate):
    final_price = base_price * (1 + tax_rate)
    return final_price

# After: Inline the variable
def calculate_price(base_price, tax_rate):
    return base_price * (1 + tax_rate)

# Before: Variable used only once
message = "Hello, " + user.name
print(message)

# After: Inline directly
print("Hello, " + user.name)
```

### Risk Assessment

- Risk Level: Low
- Potential Issues:
  - Reduced debugging capabilities
  - Less descriptive code if overused

### Best Practices

1. Keep variables that add semantic meaning
2. Inline only when expression is simple and clear
3. Consider debugging needs before inlining
4. Don't inline if it reduces readability

---

## Reorganize Imports

### Purpose

Clean up and organize import statements for better maintainability.

### When to Use

- Import statements scattered throughout file
- Unused imports present
- Imports not grouped logically
- Conflicting import aliases

### Implementation

```python
# Before: Disorganized imports
import os
import sys
from datetime import datetime
from myapp.models import User
import json
from myapp.utils import calculate_total
from collections import defaultdict

# After: Organized imports (PEP 8 style)
# Standard library imports
import json
import os
import sys
from collections import defaultdict
from datetime import datetime

# Third-party imports
import requests

# Local application imports
from myapp.models import User
from myapp.utils import calculate_total
```

### Risk Assessment

- Risk Level: Low
- Potential Issues:
  - Circular import issues
  - Breaking changes in import order

### Best Practices

1. Group imports: standard library, third-party, local
2. Sort within each group alphabetically
3. Use separate blank line between groups
4. Remove unused imports
5. Use explicit imports over wildcards

---

## Rename Method/Variable

### Purpose

Improve code clarity by using descriptive names that explain purpose.

### When to Use

- Names don't describe what something does
- Names use abbreviations or jargon
- Names are too generic (data, info, temp)
- Names conflict with domain language

### Implementation

```python
# Before: Non-descriptive names
def calc(d):
    return d * 1.1

def proc(u, o):
    if u.v:
        o.s = True
    return o

# After: Descriptive names
def calculate_price_with_tax(base_price):
    return base_price * 1.1

def process_order(user, order):
    if user.is_verified:
        order.status = OrderStatus.PROCESSED
    return order
```

### Risk Assessment

- Risk Level: Low to Medium
- Potential Issues:
  - Breaking changes in public APIs
  - References in other files/modules
  - Serialization/deserialization issues

### Best Practices

1. Use verbs for methods (calculate, process, validate)
2. Use nouns for variables and classes
3. Follow language naming conventions (snake_case for Python)
4. Rename across entire codebase consistently
5. Update documentation and comments

---

## Replace Magic Numbers with Constants

### Purpose

Replace literal values with named constants for better maintainability.

### When to Use

- Numbers appear directly in code
- Values have specific business meaning
- Numbers repeated in multiple places
- Values need to be changed frequently

### Implementation

```python
# Before: Magic numbers
def calculate_shipping_cost(weight):
    if weight < 5:
        return 10
    elif weight < 20:
        return 20
    else:
        return 35

def apply_discount(total):
    if total > 100:
        return total * 0.9
    return total

# After: Named constants
FREE_SHIPPING_WEIGHT_THRESHOLD = 5
STANDARD_SHIPPING_WEIGHT_THRESHOLD = 20
LIGHT_SHIPPING_COST = 10
STANDARD_SHIPPING_COST = 20
HEAVY_SHIPPING_COST = 35

DISCOUNT_THRESHOLD = 100
DISCOUNT_PERCENTAGE = 0.9

def calculate_shipping_cost(weight):
    if weight < FREE_SHIPPING_WEIGHT_THRESHOLD:
        return LIGHT_SHIPPING_COST
    elif weight < STANDARD_SHIPPING_WEIGHT_THRESHOLD:
        return STANDARD_SHIPPING_COST
    else:
        return HEAVY_SHIPPING_COST

def apply_discount(total):
    if total > DISCOUNT_THRESHOLD:
        return total * DISCOUNT_PERCENTAGE
    return total
```

### Risk Assessment

- Risk Level: Low
- Potential Issues:
  - Global namespace pollution
  - Finding good constant names

### Best Practices

1. Use UPPER_SNAKE_CASE for constants
2. Group related constants together
3. Add comments explaining business logic
4. Consider using enums for related constants
5. Place constants at module level or in config class

---

## Simplify Conditional Expressions

### Purpose

Reduce complexity of conditional logic for better readability.

### When to Use

- Nested if statements
- Complex boolean expressions
- Repeated condition checks
- Guard clauses missing

### Implementation

```python
# Before: Nested conditionals
def calculate_discount(user, order):
    if user:
        if user.is_active:
            if order.total > 100:
                if user.is_vip:
                    return 0.2
                else:
                    return 0.1
            else:
                return 0
        else:
            return 0
    else:
        return 0

# After: Guard clauses and early returns
def calculate_discount(user, order):
    if not user or not user.is_active:
        return 0
    
    if order.total <= 100:
        return 0
    
    return 0.2 if user.is_vip else 0.1

# Before: Complex boolean expression
if (user.age >= 18 and user.has_valid_id and user.country == "US") or \
   (user.age >= 21 and user.country == "EU") or \
   (user.is_vip and user.age >= 16):
    grant_access(user)

# After: Extract to helper method
def is_eligible_for_access(user):
    if user.is_vip and user.age >= 16:
        return True
    
    if user.country == "US":
        return user.age >= 18 and user.has_valid_id
    
    if user.country == "EU":
        return user.age >= 21
    
    return False

if is_eligible_for_access(user):
    grant_access(user)
```

### Risk Assessment

- Risk Level: Low to Medium
- Potential Issues:
  - Logic changes if not careful
  - Test coverage needs

### Best Practices

1. Use guard clauses to reduce nesting
2. Extract complex conditions to named methods
3. Use early returns to handle edge cases
4. Prefer polymorphism over complex conditionals
5. Keep boolean expressions simple and readable

---

## Decompose Conditional

### Purpose

Extract complex conditional logic into separate methods.

### When to Use

- Complex if/else statements
- Conditionals with business rules
- Repeated conditional logic
- Hard-to-test conditions

### Implementation

```python
# Before: Complex conditional logic
def calculate_shipping_cost(order):
    if order.weight < 5 and order.destination.country == "US":
        return 5.0
    elif order.weight < 5 and order.destination.country != "US":
        return 15.0
    elif order.weight >= 5 and order.weight < 20 and order.destination.country == "US":
        return 10.0
    elif order.weight >= 5 and order.weight < 20 and order.destination.country != "US":
        return 25.0
    else:
        return 50.0

# After: Decomposed into helper methods
def calculate_shipping_cost(order):
    if is_light_weight(order):
        return get_domestic_cost() if is_domestic(order) else get_international_cost(order.weight)
    elif is_medium_weight(order):
        return get_domestic_cost() * 2 if is_domestic(order) else get_international_cost(order.weight)
    else:
        return get_heavy_weight_cost()

def is_light_weight(order):
    return order.weight < 5

def is_medium_weight(order):
    return 5 <= order.weight < 20

def is_domestic(order):
    return order.destination.country == "US"

def get_domestic_cost():
    return 5.0

def get_international_cost(weight):
    return 15.0 if weight < 5 else 25.0

def get_heavy_weight_cost():
    return 50.0
```

### Risk Assessment

- Risk Level: Medium
- Potential Issues:
  - Increased method count
  - Performance considerations (method calls)

### Best Practices

1. Name condition methods to read like sentences
2. Keep helper methods private (leading underscore)
3. Extract repeated logic into reusable methods
4. Test each condition method independently
5. Document business rules clearly

---

## Extract Class

### Purpose

Extract functionality from a large class into separate, focused classes.

### When to Use

- Class has multiple responsibilities
- Class grows too large (> 300 lines)
- Class has low cohesion
- Class can be divided into logical components

### Implementation

```python
# Before: Large class with multiple responsibilities
class OrderProcessor:
    def __init__(self):
        self.db = Database()
        self.email_sender = EmailSender()
        self.payment_gateway = PaymentGateway()
    
    def process_order(self, order):
        # Validate
        self.validate_order(order)
        # Process payment
        self.process_payment(order)
        # Save to database
        self.save_order(order)
        # Send email
        self.send_confirmation(order)
    
    def validate_order(self, order):
        # Validation logic
        pass
    
    def process_payment(self, order):
        # Payment logic
        pass
    
    def save_order(self, order):
        # Database logic
        pass
    
    def send_confirmation(self, order):
        # Email logic
        pass

# After: Separated concerns into different classes
class OrderValidator:
    def validate(self, order):
        # Validation logic
        pass

class OrderRepository:
    def __init__(self, db):
        self.db = db
    
    def save(self, order):
        # Database logic
        pass

class OrderConfirmationService:
    def __init__(self, email_sender):
        self.email_sender = email_sender
    
    def send_confirmation(self, order):
        # Email logic
        pass

class OrderProcessor:
    def __init__(self, validator, repository, confirmation_service, payment_gateway):
        self.validator = validator
        self.repository = repository
        self.confirmation_service = confirmation_service
        self.payment_gateway = payment_gateway
    
    def process_order(self, order):
        self.validator.validate(order)
        self.payment_gateway.process_payment(order)
        self.repository.save(order)
        self.confirmation_service.send_confirmation(order)
```

### Risk Assessment

- Risk Level: High
- Potential Issues:
  - Breaking dependencies
  - Interface changes
  - Testing complexity
  - Refactoring cascades

### Best Practices

1. Identify clear responsibility boundaries
2. Use dependency injection for collaboration
3. Maintain clear interfaces between classes
4. Update all references to extracted functionality
5. Test thoroughly before and after extraction

---

## Best Practices Summary

1. Understand the code before refactoring
2. Write tests first (TDD approach)
3. Make small, incremental changes
4. Run tests after each change
5. Commit frequently for easy rollback
6. Update documentation and comments
7. Consider team conventions and style
8. Profile performance before and after
9. Communicate changes to team
10. Review refactoring with peers

---

## Resources

### Tools

- Rope: Automated Python refactoring
- PyCharm: Built-in refactoring tools
- VS Code: Refactoring extensions
- Black: Code formatting

### References

- Refactoring Guru: https://refactoring.guru/
- Martin Fowler's Refactoring Book
- Clean Code by Robert C. Martin
- Working Effectively with Legacy Code by Michael Feathers

---

Sub-module: `modules/refactoring/patterns.md`
Related: [ai-workflows.md](./ai-workflows.md) | [../smart-refactoring.md](../smart-refactoring.md)
