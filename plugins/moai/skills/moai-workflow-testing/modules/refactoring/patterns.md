# Refactoring Patterns

> Sub-module: Specific refactoring techniques with implementation details
> Complexity: Intermediate to Advanced
> Time: 15+ minutes per pattern
> Dependencies: source parser / refactoring tool for the host language

## Overview

This module provides detailed implementation patterns for common refactoring operations, complete with before/after sketches, risk assessments, and best practices. Sketches use language-neutral pseudo-code; apply the same transformation with the host language's refactoring tool (e.g. Go `gopls`/`gofmt`, Python Rope, JS TS-LS, Rust rust-analyzer, IntelliJ).

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

```text
# Before: long method with multiple responsibilities
process_order(order):
    # validate
    if order.items is empty: raise ValueError("Empty order")
    for item in order.items:
        if item.quantity <= 0: raise ValueError("Invalid quantity for " + item.name)
    # calculate total
    total = 0
    for item in order.items: total += item.price * item.quantity
    # apply discount
    if order.customer.is_vip: total *= 0.9
    # persist
    db.execute("INSERT INTO orders ...", order)
    for item in order.items: db.execute("INSERT INTO order_items ...", item)
    # notify
    email.send(order.customer.email, "Order confirmed")
    return total

# After: extracted methods with clear responsibilities
process_order(order):
    validate_order(order)
    total = calculate_order_total(order)
    apply_vip_discount(order)
    save_order_to_database(order)
    send_order_notification(order)
    return total

validate_order(order):
    if order.items is empty: raise ValueError("Empty order")
    for item in order.items:
        if item.quantity <= 0: raise ValueError("Invalid quantity for " + item.name)

calculate_order_total(order):
    return sum(item.price * item.quantity for item in order.items)

apply_vip_discount(order):
    if order.customer.is_vip: order.total *= 0.9

save_order_to_database(order):
    db.execute("INSERT INTO orders ...", order)
    for item in order.items: db.execute("INSERT INTO order_items ...", item)

send_order_notification(order):
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

```text
# Before: complex expressions inline
if user.age >= 18 and user.has_valid_id and user.registered_within(30, days=now):
    grant_access(user)
if order.total > 100 and order.customer.is_vip and order.shipping_address.country == "US":
    apply_free_shipping(order)

# After: extracted variables with clear names
is_adult = user.age >= 18
has_valid_identification = user.has_valid_id
registered_recently = user.registered_within(30, days=now)
if is_adult and has_valid_identification and registered_recently:
    grant_access(user)

meets_free_shipping_threshold = order.total > 100
is_vip_customer = order.customer.is_vip
ships_domestically = order.shipping_address.country == "US"
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

```text
# Before: unnecessary intermediate variable
calculate_price(base_price, tax_rate):
    final_price = base_price * (1 + tax_rate)
    return final_price

# After: inline the variable
calculate_price(base_price, tax_rate):
    return base_price * (1 + tax_rate)

# Before: variable used only once
message = "Hello, " + user.name
print(message)

# After: inline directly
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

```text
# Before: disorganized imports
import os
import sys
from datetime import datetime
from myapp.models import User
import json
from myapp.utils import calculate_total
from collections import defaultdict

# After: organized in three groups with blank lines between
# 1. standard-library imports
import json, os, sys
from collections import defaultdict
from datetime import datetime
# 2. third-party imports
import requests
# 3. local application imports
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
3. Use a blank line between groups
4. Remove unused imports
5. Prefer explicit imports over wildcards

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

```text
# Before: non-descriptive names
calc(d): return d * 1.1
proc(u, o):
    if u.v: o.s = true
    return o

# After: descriptive names
calculate_price_with_tax(base_price): return base_price * 1.1
process_order(user, order):
    if user.is_verified: order.status = PROCESSED
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
3. Follow the host language's naming convention (e.g. snake_case, camelCase, PascalCase)
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

```text
# Before: magic numbers
calculate_shipping_cost(weight):
    if weight < 5:  return 10
    elif weight < 20: return 20
    else: return 35
apply_discount(total):
    if total > 100: return total * 0.9
    return total

# After: named constants (per the host language's constant idiom)
FREE_SHIPPING_WEIGHT_THRESHOLD     = 5
STANDARD_SHIPPING_WEIGHT_THRESHOLD = 20
LIGHT_SHIPPING_COST    = 10
STANDARD_SHIPPING_COST = 20
HEAVY_SHIPPING_COST    = 35
DISCOUNT_THRESHOLD  = 100
DISCOUNT_PERCENTAGE = 0.9

calculate_shipping_cost(weight):
    if weight < FREE_SHIPPING_WEIGHT_THRESHOLD:  return LIGHT_SHIPPING_COST
    elif weight < STANDARD_SHIPPING_WEIGHT_THRESHOLD: return STANDARD_SHIPPING_COST
    else: return HEAVY_SHIPPING_COST
apply_discount(total):
    if total > DISCOUNT_THRESHOLD: return total * DISCOUNT_PERCENTAGE
    return total
```

### Risk Assessment

- Risk Level: Low
- Potential Issues:
  - Global namespace pollution
  - Finding good constant names

### Best Practices

1. Use the host language's constant-naming convention (often UPPER_SNAKE_CASE)
2. Group related constants together
3. Add comments explaining business logic
4. Consider using enums for related constants
5. Place constants at module level or in a config type

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

```text
# Before: nested conditionals
calculate_discount(user, order):
    if user:
        if user.is_active:
            if order.total > 100:
                if user.is_vip: return 0.2
                else: return 0.1
            else: return 0
        else: return 0
    else: return 0

# After: guard clauses and early returns
calculate_discount(user, order):
    if not user or not user.is_active: return 0
    if order.total <= 100: return 0
    return 0.2 if user.is_vip else 0.1

# Before: complex boolean expression
if (user.age >= 18 and user.has_valid_id and user.country == "US")
 or (user.age >= 21 and user.country == "EU")
 or (user.is_vip and user.age >= 16):
    grant_access(user)

# After: extract to a helper method
is_eligible_for_access(user):
    if user.is_vip and user.age >= 16: return true
    if user.country == "US": return user.age >= 18 and user.has_valid_id
    if user.country == "EU": return user.age >= 21
    return false

if is_eligible_for_access(user): grant_access(user)
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

```text
# Before: complex conditional logic
calculate_shipping_cost(order):
    if order.weight < 5 and order.destination.country == "US":  return 5.0
    elif order.weight < 5 and order.destination.country != "US": return 15.0
    elif 5 <= order.weight < 20 and order.destination.country == "US":  return 10.0
    elif 5 <= order.weight < 20 and order.destination.country != "US": return 25.0
    else: return 50.0

# After: decomposed into helper methods
calculate_shipping_cost(order):
    if is_light_weight(order):
        return get_domestic_cost() if is_domestic(order) else get_international_cost(order.weight)
    elif is_medium_weight(order):
        return get_domestic_cost() * 2 if is_domestic(order) else get_international_cost(order.weight)
    else:
        return get_heavy_weight_cost()

is_light_weight(order):  return order.weight < 5
is_medium_weight(order): return 5 <= order.weight < 20
is_domestic(order):      return order.destination.country == "US"
get_domestic_cost():     return 5.0
get_international_cost(weight): return 15.0 if weight < 5 else 25.0
get_heavy_weight_cost(): return 50.0
```

### Risk Assessment

- Risk Level: Medium
- Potential Issues:
  - Increased method count
  - Performance considerations (method calls)

### Best Practices

1. Name condition methods to read like sentences
2. Keep helper methods private/internal
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

```text
# Before: large class with multiple responsibilities
class OrderProcessor:
    db
    email_sender
    payment_gateway
    process_order(order):
        validate_order(order)        # validation
        process_payment(order)       # payment
        save_order(order)            # persistence
        send_confirmation(order)     # email
    validate_order(order): ...
    process_payment(order): ...
    save_order(order): ...
    send_confirmation(order): ...

# After: separated concerns into focused classes
class OrderValidator: validate(order): ...
class OrderRepository: save(order): ...
class OrderConfirmationService: send_confirmation(order): ...

class OrderProcessor(validator, repository, confirmation_service, payment_gateway):
    process_order(order):
        validator.validate(order)
        payment_gateway.process_payment(order)
        repository.save(order)
        confirmation_service.send_confirmation(order)
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

- Refactoring tools: the host language's IDE/LSP (PyCharm/Rope, VS Code, gopls, rust-analyzer, IntelliJ)
- Formatter: the host language's formatter (gofmt, black, prettier, rustfmt)

### References

- Refactoring Guru: https://refactoring.guru/
- Martin Fowler's Refactoring Book
- Clean Code by Robert C. Martin
- Working Effectively with Legacy Code by Michael Feathers

---

Sub-module: `modules/refactoring/patterns.md`
Related: [ai-workflows.md](./ai-workflows.md) | [../smart-refactoring.md](../smart-refactoring.md)
