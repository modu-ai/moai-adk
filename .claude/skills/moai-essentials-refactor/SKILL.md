---
name: moai-essentials-refactor
description: Refactoring guidance with design patterns and code improvement strategies
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
tier: 1
auto-load: "true"
---

# Alfred Refactoring Coach

## What it does

Refactoring guidance with design pattern recommendations, code smell detection, and step-by-step improvement plans.

## When to use

- "리팩토링 도와줘", "이 코드 개선 방법은?", "디자인 패턴 적용", "코드 정리", "구조 개선"
- "중복 제거", "함수 분리", "클래스 분리", "코드 스멜 제거", "복잡도 낮추기"
- "Refactoring", "Design patterns", "Code cleanup", "Extract method", "DRY principle"
- When code becomes hard to maintain
- Before adding new features to legacy code

## How it works

**Refactoring Techniques**:
- **Extract Method**: 긴 메서드 분리
- **Replace Conditional with Polymorphism**: 조건문 제거
- **Introduce Parameter Object**: 매개변수 그룹화
- **Extract Class**: 거대한 클래스 분리

**Design Pattern Recommendations**:
- Complex object creation → **Builder Pattern**
- Type-specific behavior → **Strategy Pattern**
- Global state → **Singleton Pattern**
- Incompatible interfaces → **Adapter Pattern**
- Delayed object creation → **Factory Pattern**

**3-Strike Rule**:
```
1st occurrence: Just implement
2nd occurrence: Notice similarity (leave as-is)
3rd occurrence: Pattern confirmed → Refactor! 🔧
```

**Refactoring Checklist**:
- [ ] All tests passing before refactoring
- [ ] Code smells identified
- [ ] Refactoring goal clear
- [ ] Change one thing at a time
- [ ] Run tests after each change
- [ ] Commit frequently

## Code Smells & Refactoring Techniques

### Long Method → Extract Method
```python
# ❌ Before (85 LOC)
def process_order(order):
    # Validate order (20 lines)
    if not order.items:
        raise ValueError("Empty order")
    for item in order.items:
        if item.quantity <= 0:
            raise ValueError("Invalid quantity")
    # Calculate total (30 lines)
    subtotal = sum(item.price * item.quantity for item in order.items)
    tax = subtotal * 0.1
    shipping = calculate_shipping(order)
    total = subtotal + tax + shipping
    # Process payment (35 lines)
    # ...

# ✅ After (15 LOC)
def process_order(order):
    validate_order(order)  # Extracted
    total = calculate_total(order)  # Extracted
    process_payment(order, total)  # Extracted
```

### Duplicate Code → Extract Function
```python
# ❌ Before (duplicated logic)
if user.age >= 18:
    user.adult = True
    user.can_vote = True
    user.can_drive = True

if customer.age >= 18:
    customer.adult = True
    customer.can_vote = True
    customer.can_drive = True

# ✅ After (DRY)
def mark_as_adult(person):
    if person.age >= 18:
        person.adult = True
        person.can_vote = True
        person.can_drive = True

mark_as_adult(user)
mark_as_adult(customer)
```

### Long Parameter List → Parameter Object
```python
# ❌ Before (too many parameters)
def create_user(name, email, age, address, city, country, postal_code):
    # ...

# ✅ After (parameter object)
@dataclass
class UserInfo:
    name: str
    email: str
    age: int
    address: str
    city: str
    country: str
    postal_code: str

def create_user(info: UserInfo):
    # ...
```

### Complex Conditional → Guard Clauses
```python
# ❌ Before (nested if-else)
def process_payment(payment):
    if payment:
        if payment.amount > 0:
            if payment.method == "credit_card":
                return charge_credit_card(payment)
            else:
                return "Invalid method"
        else:
            return "Invalid amount"
    else:
        return "No payment"

# ✅ After (guard clauses)
def process_payment(payment):
    if not payment:
        return "No payment"
    if payment.amount <= 0:
        return "Invalid amount"
    if payment.method != "credit_card":
        return "Invalid method"

    return charge_credit_card(payment)
```

### Switch Statement → Polymorphism
```python
# ❌ Before (type-based branching)
def calculate_price(product):
    if product.type == "book":
        return product.price * 0.9  # 10% discount
    elif product.type == "electronics":
        return product.price * 0.85  # 15% discount
    elif product.type == "clothing":
        return product.price * 0.8  # 20% discount

# ✅ After (polymorphism)
class Product:
    def calculate_price(self):
        raise NotImplementedError

class Book(Product):
    def calculate_price(self):
        return self.price * 0.9

class Electronics(Product):
    def calculate_price(self):
        return self.price * 0.85

class Clothing(Product):
    def calculate_price(self):
        return self.price * 0.8
```

## Design Pattern Examples

### Builder Pattern
```python
# ❌ Before (complex constructor)
pizza = Pizza(size="large", cheese=True, pepperoni=True,
              mushrooms=False, olives=True, ...)

# ✅ After (builder)
pizza = Pizza.Builder() \
    .size("large") \
    .add_cheese() \
    .add_pepperoni() \
    .add_olives() \
    .build()
```

### Strategy Pattern
```python
# ❌ Before (type checking)
def sort_items(items, sort_type):
    if sort_type == "price":
        return sorted(items, key=lambda x: x.price)
    elif sort_type == "name":
        return sorted(items, key=lambda x: x.name)

# ✅ After (strategy)
class SortStrategy:
    def sort(self, items):
        raise NotImplementedError

class PriceSort(SortStrategy):
    def sort(self, items):
        return sorted(items, key=lambda x: x.price)

class NameSort(SortStrategy):
    def sort(self, items):
        return sorted(items, key=lambda x: x.name)
```

### Factory Pattern
```python
# ❌ Before (direct instantiation)
if config.db_type == "postgres":
    db = PostgresDB(config)
elif config.db_type == "mysql":
    db = MySQLDB(config)

# ✅ After (factory)
class DBFactory:
    @staticmethod
    def create(db_type, config):
        if db_type == "postgres":
            return PostgresDB(config)
        elif db_type == "mysql":
            return MySQLDB(config)
        raise ValueError(f"Unknown DB type: {db_type}")

db = DBFactory.create(config.db_type, config)
```

## Examples

### Example 1: Extract Method Refactoring
User: "중복 코드 제거해줘"

Alfred identifies:
```python
# Duplicate pattern detected (3 occurrences)
user_service.validate_email(email)
user_service.hash_password(password)
user_service.create_user(email, password)

# Extracted method
def register_user(email, password):
    user_service.validate_email(email)
    user_service.hash_password(password)
    user_service.create_user(email, password)
```

### Example 2: Reduce Complexity
User: "복잡도 낮춰줘"

Alfred refactors:
```python
# Before: Complexity 15 (too high)
def process_data(data):
    for item in data:
        if item.type == "A":
            if item.value > 10:
                # ... 20 lines
        elif item.type == "B":
            # ... 30 lines

# After: Complexity 5 (each function <10)
def process_data(data):
    for item in data:
        if item.type == "A":
            process_type_a(item)
        elif item.type == "B":
            process_type_b(item)
```

### Example 3: Apply Design Pattern
User: "디자인 패턴 적용해줘"

Alfred recommends:
```python
# Complex object creation detected → Builder Pattern

class ReportBuilder:
    def __init__(self):
        self.report = Report()

    def add_header(self, header):
        self.report.header = header
        return self

    def add_data(self, data):
        self.report.data = data
        return self

    def build(self):
        return self.report

# Usage
report = ReportBuilder() \
    .add_header("Sales Report") \
    .add_data(sales_data) \
    .build()
```

## Works well with

- moai-essentials-review (Review before refactoring)
- moai-foundation-trust (Ensure TRUST R: Readable)
