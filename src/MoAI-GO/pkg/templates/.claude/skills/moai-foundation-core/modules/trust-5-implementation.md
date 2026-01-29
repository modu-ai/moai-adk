# TRUST 5 Implementation Patterns

Purpose: Detailed implementation patterns for TRUST 5 principles with working code examples.

Version: 1.0.0
Last Updated: 2026-01-06
Parent: [trust-5-framework.md](trust-5-framework.md)

---

## Test-First Implementation

### RED-GREEN-REFACTOR Example

```python
# RED: Write failing test first
def test_calculate_total_price_with_tax():
    item = ShoppingItem(name="Widget", price=10.00)
    total = calculate_total_with_tax(item, tax_rate=0.10)
    assert total == 11.00  # Fails - function doesn't exist

# GREEN: Minimal implementation
def calculate_total_with_tax(item, tax_rate):
    return item.price * (1 + tax_rate)

# REFACTOR: Improve code quality
def calculate_total_with_tax(item: ShoppingItem, tax_rate: float) -> float:
    """Calculate total price including tax.

    Args:
        item: Shopping item with price
        tax_rate: Tax rate as decimal (0.10 = 10%)

    Returns:
        Total price including tax

    Raises:
        ValueError: If tax_rate not between 0 and 1

    Example:
        >>> item = ShoppingItem("Widget", 10.00)
        >>> calculate_total_with_tax(item, 0.10)
        11.0
    """
    if not 0 <= tax_rate <= 1:
        raise ValueError("Tax rate must be between 0 and 1")

    return item.price * (1 + tax_rate)
```

---

## Readable Code Patterns

### Bad vs Good Examples

```python
# BAD: Unreadable, no types, magic numbers
def calc(x, y):
    if x > 0:
        if y > 0:
            if x + y < 100:
                return x * 1.1 + y * 0.9
    return 0

# GOOD: Readable, typed, constants
TAX_RATE = 0.10
DISCOUNT_RATE = 0.10
MAX_TOTAL = 100.00

def calculate_order_total(
    base_amount: float,
    discount_amount: float
) -> float:
    """Calculate order total with tax and discount.

    Args:
        base_amount: Base order amount before tax
        discount_amount: Discount amount to apply

    Returns:
        Final order total with tax applied

    Raises:
        ValueError: If amounts are negative or exceed max
    """
    if base_amount < 0 or discount_amount < 0:
        raise ValueError("Amounts must be non-negative")

    subtotal = base_amount - discount_amount

    if subtotal > MAX_TOTAL:
        raise ValueError(f"Total exceeds maximum {MAX_TOTAL}")

    return subtotal * (1 + TAX_RATE)
```

---

## Unified Pattern Examples

### Standard Error Handling

```python
class DomainError(Exception):
    """Base error for domain-specific errors."""
    pass

class ValidationError(DomainError):
    """Validation failed."""
    pass

def process_data(data: dict) -> Result:
    """Standard processing pattern."""
    try:
        validated = validate_input(data)
        result = perform_processing(validated)
        return Result(success=True, data=result)

    except ValidationError as e:
        logger.error(f"Validation failed: {e}")
        raise
    except Exception as e:
        logger.exception(f"Processing failed: {e}")
        raise DomainError(f"Processing failed: {e}") from e
```

---

## Secured Code Patterns

### Access Control (RBAC)

```python
from functools import wraps

def require_permission(permission: str):
    """Decorator to enforce permission checks."""
    def decorator(func):
        @wraps(func)
        def wrapper(user: User, *args, **kwargs):
            if not user.has_permission(permission):
                raise UnauthorizedError(
                    f"User lacks permission: {permission}"
                )
            return func(user, *args, **kwargs)
        return wrapper
    return decorator

@require_permission("user:update")
def update_user_profile(user: User, profile_data: dict) -> UserProfile:
    """Update user profile (requires permission)."""
    return user.update_profile(profile_data)
```

### Password Hashing

```python
from bcrypt import hashpw, gensalt, checkpw

def hash_password(plaintext: str) -> str:
    """Hash password securely with bcrypt."""
    salt = gensalt(rounds=12)
    return hashpw(plaintext.encode('utf-8'), salt).decode('utf-8')

def verify_password(plaintext: str, hashed: str) -> bool:
    """Verify password against hash."""
    return checkpw(plaintext.encode('utf-8'), hashed.encode('utf-8'))
```

### Injection Prevention

```python
from sqlalchemy import text

def safe_user_query(username: str) -> List[User]:
    """Query users safely with parameterized query."""
    query = text("SELECT * FROM users WHERE username = :username")
    return db.session.execute(query, {"username": username}).fetchall()
```

### Secure Configuration

```python
import os

def load_secure_config() -> dict:
    """Load configuration from environment variables."""
    config = {
        'DEBUG': os.getenv('DEBUG', 'false').lower() == 'true',
        'DATABASE_URL': os.getenv('DATABASE_URL'),
        'SECRET_KEY': os.getenv('SECRET_KEY'),
        'ALLOWED_HOSTS': os.getenv('ALLOWED_HOSTS', '').split(',')
    }

    required = ['DATABASE_URL', 'SECRET_KEY']
    for key in required:
        if not config.get(key):
            raise ValueError(f"Required config missing: {key}")

    logger.info(f"Config loaded (DEBUG={config['DEBUG']})")
    return config
```

---

## Trackable Patterns

### Traceability Matrix

```yaml
# .moai/specs/traceability.yaml
requirements:
  SPEC-001-REQ-01:
    description: "User registration with email/password"
    implementation:
      - src/auth/registration.py::register_user
      - src/models/user.py::User
    tests:
      - tests/auth/test_registration.py::test_register_user_success
    coverage: 95%
    status: Implemented

  SPEC-001-REQ-02:
    description: "OAuth2 authentication"
    implementation:
      - src/auth/oauth2.py::OAuth2Handler
    tests:
      - tests/auth/test_oauth2.py::test_oauth2_flow
    coverage: 92%
    status: Implemented
```

---

## Works Well With

- [trust-5-framework.md](trust-5-framework.md) - Overview and principles
- [trust-5-validation.md](trust-5-validation.md) - CI/CD integration

---

Version: 1.0.0
Last Updated: 2026-01-06
