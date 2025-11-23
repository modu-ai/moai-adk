# Refactoring Patterns & Techniques

## 5 Core Refactoring Patterns

### Pattern 1: Extract Method

**Purpose**: Break down long methods into smaller, focused functions.

```python
class RefactoringExtractor:
    """Extract Method refactoring pattern."""

    # BEFORE: Long method doing multiple things (Violation of SRP)
    def process_order_before(self, order_data: dict):
        """Process order with mixed concerns."""

        # Validation
        if not order_data.get('customer_id'):
            raise ValueError("Customer ID required")
        if not order_data.get('items'):
            raise ValueError("Order items required")
        if order_data['total'] < 0:
            raise ValueError("Invalid total")

        # Calculate tax
        subtotal = sum(item['price'] * item['quantity'] for item in order_data['items'])
        tax_rate = 0.08
        tax = subtotal * tax_rate
        order_data['tax'] = tax

        # Calculate total
        order_data['total'] = subtotal + tax + order_data.get('shipping', 0)

        # Save to database
        db.orders.insert(order_data)

        # Send email
        email_body = f"Order {order_data['id']} confirmed. Total: ${order_data['total']}"
        send_email(order_data['customer_email'], "Order Confirmation", email_body)

        return order_data

    # AFTER: Extracted into focused methods
    def process_order_after(self, order_data: dict):
        """Process order with extracted methods."""

        self._validate_order_data(order_data)
        self._calculate_order_totals(order_data)
        self._save_order_to_database(order_data)
        self._send_order_confirmation_email(order_data)

        return order_data

    def _validate_order_data(self, order_data: dict) -> None:
        """Validate order data."""
        if not order_data.get('customer_id'):
            raise ValueError("Customer ID required")
        if not order_data.get('items'):
            raise ValueError("Order items required")
        if order_data['total'] < 0:
            raise ValueError("Invalid total")

    def _calculate_order_totals(self, order_data: dict) -> None:
        """Calculate tax and total."""
        subtotal = sum(item['price'] * item['quantity'] for item in order_data['items'])
        tax = self._calculate_tax(subtotal)
        order_data['tax'] = tax
        order_data['total'] = subtotal + tax + order_data.get('shipping', 0)

    def _calculate_tax(self, subtotal: float) -> float:
        """Calculate tax amount."""
        TAX_RATE = 0.08
        return subtotal * TAX_RATE

    def _save_order_to_database(self, order_data: dict) -> None:
        """Save order to database."""
        db.orders.insert(order_data)

    def _send_order_confirmation_email(self, order_data: dict) -> None:
        """Send order confirmation email."""
        email_body = f"Order {order_data['id']} confirmed. Total: ${order_data['total']}"
        send_email(order_data['customer_email'], "Order Confirmation", email_body)
```

### Pattern 2: Inline Method

**Purpose**: Remove unnecessary method indirection.

```python
class InlineRefactoring:
    """Inline Method pattern."""

    # BEFORE: Unnecessary indirection
    def get_rating_before(self, driver):
        return 1 if self._more_than_five_late_deliveries(driver) else 2

    def _more_than_five_late_deliveries(self, driver):
        return driver.late_deliveries > 5

    # AFTER: Inlined for clarity
    def get_rating_after(self, driver):
        return 1 if driver.late_deliveries > 5 else 2
```

### Pattern 3: Move Method

**Purpose**: Move methods to classes where they naturally belong.

```python
# BEFORE: Method in wrong class
class Order:
    def __init__(self, items):
        self.items = items

class OrderProcessor:
    def calculate_total_price(self, order):
        """Calculate total - should be in Order class."""
        return sum(item.price * item.quantity for item in order.items)

# AFTER: Method moved to appropriate class
class Order:
    def __init__(self, items):
        self.items = items

    def calculate_total_price(self):
        """Calculate total price of order."""
        return sum(item.price * item.quantity for item in self.items)

class OrderProcessor:
    def process_order(self, order):
        """Process order using Order's method."""
        total = order.calculate_total_price()
        return total
```

### Pattern 4: Rename Method/Variable

**Purpose**: Improve code readability with clear names.

```python
class RenameRefactoring:
    """Rename for clarity."""

    # BEFORE: Unclear names
    def calc(self, d1, d2):
        t = (d2 - d1).days
        if t > 30:
            return t * 0.1
        return t * 0.15

    # AFTER: Clear, descriptive names
    def calculate_rental_discount(self, start_date, end_date):
        """Calculate rental discount based on duration."""
        rental_days = (end_date - start_date).days
        LONG_TERM_DISCOUNT_RATE = 0.1
        SHORT_TERM_DISCOUNT_RATE = 0.15
        LONG_TERM_THRESHOLD_DAYS = 30

        if rental_days > LONG_TERM_THRESHOLD_DAYS:
            return rental_days * LONG_TERM_DISCOUNT_RATE

        return rental_days * SHORT_TERM_DISCOUNT_RATE
```

### Pattern 5: Replace Conditional with Polymorphism

**Purpose**: Replace complex conditionals with polymorphism.

```python
# BEFORE: Complex conditional logic
class EmployeePayroll:
    def calculate_pay(self, employee):
        if employee.type == "manager":
            return employee.salary + employee.bonus
        elif employee.type == "engineer":
            return employee.salary + (employee.overtime_hours * employee.hourly_rate)
        elif employee.type == "salesperson":
            return employee.salary + (employee.sales * 0.05)
        else:
            return employee.salary

# AFTER: Polymorphism
from abc import ABC, abstractmethod

class Employee(ABC):
    """Base employee class."""

    def __init__(self, salary):
        self.salary = salary

    @abstractmethod
    def calculate_pay(self):
        """Calculate employee pay."""
        pass

class Manager(Employee):
    def __init__(self, salary, bonus):
        super().__init__(salary)
        self.bonus = bonus

    def calculate_pay(self):
        return self.salary + self.bonus

class Engineer(Employee):
    def __init__(self, salary, overtime_hours, hourly_rate):
        super().__init__(salary)
        self.overtime_hours = overtime_hours
        self.hourly_rate = hourly_rate

    def calculate_pay(self):
        return self.salary + (self.overtime_hours * self.hourly_rate)

class Salesperson(Employee):
    def __init__(self, salary, sales):
        super().__init__(salary)
        self.sales = sales

    def calculate_pay(self):
        COMMISSION_RATE = 0.05
        return self.salary + (self.sales * COMMISSION_RATE)
```

## Code Smell Identification (10 Types)

### Smell 1: Duplicated Code

```python
# BAD: Duplicated validation logic
def process_user_registration(user_data):
    if not user_data.get('email'):
        raise ValueError("Email required")
    if not user_data.get('password'):
        raise ValueError("Password required")
    if len(user_data['password']) < 8:
        raise ValueError("Password must be at least 8 characters")
    # ... registration logic

def process_user_login(user_data):
    if not user_data.get('email'):
        raise ValueError("Email required")
    if not user_data.get('password'):
        raise ValueError("Password required")
    # ... login logic

# GOOD: Extract common validation
class UserValidator:
    @staticmethod
    def validate_credentials(user_data):
        if not user_data.get('email'):
            raise ValueError("Email required")
        if not user_data.get('password'):
            raise ValueError("Password required")

    @staticmethod
    def validate_password_strength(password):
        if len(password) < 8:
            raise ValueError("Password must be at least 8 characters")

def process_user_registration(user_data):
    UserValidator.validate_credentials(user_data)
    UserValidator.validate_password_strength(user_data['password'])
    # ... registration logic

def process_user_login(user_data):
    UserValidator.validate_credentials(user_data)
    # ... login logic
```

### Smell 2: Long Method

```python
# BAD: 50+ line method
def process_invoice(invoice_data):
    # 10 lines of validation
    # 15 lines of calculation
    # 10 lines of database operations
    # 15 lines of notification
    pass  # Total: 50+ lines

# GOOD: Extracted into focused methods
def process_invoice(invoice_data):
    validate_invoice(invoice_data)
    calculate_totals(invoice_data)
    save_invoice(invoice_data)
    send_notifications(invoice_data)
```

### Smell 3: Large Class

```python
# BAD: God object with too many responsibilities
class OrderManager:
    def validate_order(self): pass
    def calculate_tax(self): pass
    def calculate_shipping(self): pass
    def save_to_database(self): pass
    def send_email(self): pass
    def process_payment(self): pass
    def update_inventory(self): pass
    def generate_invoice(self): pass
    # ... 20+ methods

# GOOD: Single responsibility classes
class OrderValidator:
    def validate(self, order): pass

class OrderCalculator:
    def calculate_tax(self, order): pass
    def calculate_shipping(self, order): pass

class OrderRepository:
    def save(self, order): pass

class OrderNotifier:
    def send_confirmation_email(self, order): pass

class PaymentProcessor:
    def process(self, order): pass
```

### Smell 4: Long Parameter List

```python
# BAD: Too many parameters
def create_user(first_name, last_name, email, phone, address, city, state, zip_code, country):
    pass

# GOOD: Use data class or object
from dataclasses import dataclass

@dataclass
class UserRegistrationData:
    first_name: str
    last_name: str
    email: str
    phone: str
    address: str
    city: str
    state: str
    zip_code: str
    country: str

def create_user(user_data: UserRegistrationData):
    pass
```

### Smell 5: Feature Envy

```python
# BAD: Method uses another class's data too much
class Order:
    def __init__(self, customer):
        self.customer = customer

class OrderProcessor:
    def calculate_discount(self, order):
        # Feature envy: using customer data extensively
        if order.customer.membership_level == "gold":
            base = order.customer.total_purchases
            return base * 0.15
        elif order.customer.membership_level == "silver":
            return order.customer.total_purchases * 0.10
        return 0

# GOOD: Move method to Customer class
class Customer:
    def __init__(self, membership_level, total_purchases):
        self.membership_level = membership_level
        self.total_purchases = total_purchases

    def calculate_discount(self):
        DISCOUNT_RATES = {
            "gold": 0.15,
            "silver": 0.10
        }
        rate = DISCOUNT_RATES.get(self.membership_level, 0)
        return self.total_purchases * rate

class Order:
    def __init__(self, customer):
        self.customer = customer

    def get_discount(self):
        return self.customer.calculate_discount()
```

### Smell 6-10: Quick Reference

- **Smell 6: Primitive Obsession** - Use objects instead of primitives
- **Smell 7: Switch Statements** - Replace with polymorphism
- **Smell 8: Speculative Generality** - Remove unused flexibility
- **Smell 9: Temporary Field** - Extract into separate class
- **Smell 10: Message Chains** - Apply Law of Demeter

## Safe Refactoring Strategy

### Step 1: Write Tests First

```python
import pytest

def test_calculate_discount():
    """Test before refactoring."""
    order = Order(items=[Item(price=100, quantity=2)])
    discount = calculate_discount(order)
    assert discount == 20.0  # 10% discount on $200
```

### Step 2: Refactor in Small Steps

```python
# Refactor 1: Extract constant
DISCOUNT_RATE = 0.10

def calculate_discount(order):
    total = sum(item.price * item.quantity for item in order.items)
    return total * DISCOUNT_RATE

# Refactor 2: Extract method
def calculate_order_total(order):
    return sum(item.price * item.quantity for item in order.items)

def calculate_discount(order):
    total = calculate_order_total(order)
    return total * DISCOUNT_RATE
```

### Step 3: Run Tests After Each Change

```python
# Run tests after EVERY refactoring step
pytest test_calculate_discount.py

# If tests pass: ✅ Continue
# If tests fail: ❌ Revert and try different approach
```

---

**Last Updated**: 2025-11-23
**Status**: Production Ready
**Lines**: 270
**Code Examples**: 6+ comprehensive refactoring patterns
