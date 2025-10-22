# Refactoring Examples

_Last updated: 2025-10-22_

## Example 1: Extract Method

### Code Smell: Long Method
Method doing too many things, hard to understand.

### Before
```python
def process_order(order_data):
    # Validate
    if not order_data.get('customer_id'):
        raise ValueError("Customer ID required")
    if not order_data.get('items'):
        raise ValueError("Items required")
    if len(order_data['items']) == 0:
        raise ValueError("Order must have items")

    # Calculate
    total = 0
    for item in order_data['items']:
        price = item['price']
        quantity = item['quantity']
        discount = item.get('discount', 0)
        total += (price * quantity) * (1 - discount)

    # Apply tax
    tax_rate = 0.08
    total_with_tax = total * (1 + tax_rate)

    # Save
    db.execute("INSERT INTO orders ...")
    return total_with_tax
```

### After (Extract Method)
```python
def process_order(order_data):
    validate_order(order_data)
    subtotal = calculate_subtotal(order_data['items'])
    total = apply_tax(subtotal)
    save_order(order_data, total)
    return total

def validate_order(order_data):
    if not order_data.get('customer_id'):
        raise ValueError("Customer ID required")
    if not order_data.get('items') or len(order_data['items']) == 0:
        raise ValueError("Order must have items")

def calculate_subtotal(items):
    return sum(
        item['price'] * item['quantity'] * (1 - item.get('discount', 0))
        for item in items
    )

def apply_tax(amount, rate=0.08):
    return amount * (1 + rate)

def save_order(order_data, total):
    db.execute("INSERT INTO orders ...", {**order_data, 'total': total})
```

---

## Example 2: Replace Conditional with Polymorphism

### Code Smell: Type Code
Switch statement based on type field.

### Before
```python
class Shape:
    def __init__(self, type, width, height=None):
        self.type = type
        self.width = width
        self.height = height

    def area(self):
        if self.type == 'rectangle':
            return self.width * self.height
        elif self.type == 'circle':
            return 3.14 * (self.width / 2) ** 2
        elif self.type == 'square':
            return self.width ** 2
        else:
            raise ValueError(f"Unknown shape: {self.type}")
```

### After (Polymorphism)
```python
from abc import ABC, abstractmethod

class Shape(ABC):
    @abstractmethod
    def area(self):
        pass

class Rectangle(Shape):
    def __init__(self, width, height):
        self.width = width
        self.height = height

    def area(self):
        return self.width * self.height

class Circle(Shape):
    def __init__(self, radius):
        self.radius = radius

    def area(self):
        return 3.14 * self.radius ** 2

class Square(Shape):
    def __init__(self, side):
        self.side = side

    def area(self):
        return self.side ** 2

# Usage
shapes = [Rectangle(5, 10), Circle(7), Square(4)]
total_area = sum(shape.area() for shape in shapes)
```

---

## Example 3: Introduce Parameter Object

### Code Smell: Long Parameter List
Too many parameters, hard to remember order.

### Before
```python
def create_user(
    username,
    email,
    password,
    first_name,
    last_name,
    date_of_birth,
    phone,
    address_line1,
    address_line2,
    city,
    state,
    zip_code,
    country
):
    # Long parameter list is hard to use
    pass

# Calling code (error-prone)
create_user(
    "john_doe",
    "john@example.com",
    "password123",
    "John",
    "Doe",
    "1990-01-01",
    "555-1234",
    "123 Main St",
    "Apt 4",
    "Springfield",
    "IL",
    "62701",
    "USA"
)
```

### After (Parameter Object)
```python
from dataclasses import dataclass

@dataclass
class Address:
    line1: str
    line2: str = ""
    city: str = ""
    state: str = ""
    zip_code: str = ""
    country: str = ""

@dataclass
class UserProfile:
    username: str
    email: str
    password: str
    first_name: str
    last_name: str
    date_of_birth: str
    phone: str
    address: Address

def create_user(profile: UserProfile):
    # Clean, type-safe interface
    pass

# Calling code (clear and maintainable)
address = Address(
    line1="123 Main St",
    line2="Apt 4",
    city="Springfield",
    state="IL",
    zip_code="62701",
    country="USA"
)

profile = UserProfile(
    username="john_doe",
    email="john@example.com",
    password="password123",
    first_name="John",
    last_name="Doe",
    date_of_birth="1990-01-01",
    phone="555-1234",
    address=address
)

create_user(profile)
```

---

## Example 4: Remove Duplication

### Code Smell: Duplicated Code
Same logic repeated in multiple places.

### Before
```javascript
function calculateMonthlyPayment(principal, annualRate, years) {
  const monthlyRate = annualRate / 12 / 100;
  const months = years * 12;
  const payment = principal * 
    (monthlyRate * Math.pow(1 + monthlyRate, months)) /
    (Math.pow(1 + monthlyRate, months) - 1);
  return payment;
}

function calculateTotalPayment(principal, annualRate, years) {
  const monthlyRate = annualRate / 12 / 100;
  const months = years * 12;
  const monthlyPayment = principal * 
    (monthlyRate * Math.pow(1 + monthlyRate, months)) /
    (Math.pow(1 + monthlyRate, months) - 1);
  return monthlyPayment * months;
}
```

### After (DRY - Don't Repeat Yourself)
```javascript
function calculateMonthlyPayment(principal, annualRate, years) {
  const monthlyRate = annualRate / 12 / 100;
  const months = years * 12;
  return principal * 
    (monthlyRate * Math.pow(1 + monthlyRate, months)) /
    (Math.pow(1 + monthlyRate, months) - 1);
}

function calculateTotalPayment(principal, annualRate, years) {
  const monthlyPayment = calculateMonthlyPayment(principal, annualRate, years);
  const months = years * 12;
  return monthlyPayment * months;
}
```

---

_For refactoring patterns and code smells, see reference.md_
