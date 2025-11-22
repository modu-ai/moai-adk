# 리팩토링 실전 예제

## Example 1: 메서드 추출 (Method Extraction)

**리팩토링 전**:
```python
def calculate_invoice_total(items, tax_rate, discount):
    """복잡한 계산 로직."""
    subtotal = sum(item['price'] * item['quantity'] for item in items)

    # 세금 계산 (복잡함)
    if subtotal > 1000:
        tax = subtotal * tax_rate * 0.95
    else:
        tax = subtotal * tax_rate

    # 할인 계산 (복잡함)
    if discount > 0:
        final_amount = (subtotal + tax) * (1 - discount / 100)
    else:
        final_amount = subtotal + tax

    return final_amount
```

**리팩토링 후** (메서드 추출):
```python
def calculate_invoice_total(items, tax_rate, discount):
    """송장 총액 계산."""
    subtotal = self._calculate_subtotal(items)
    tax = self._calculate_tax(subtotal, tax_rate)
    final_amount = self._apply_discount(subtotal + tax, discount)
    return final_amount

def _calculate_subtotal(self, items):
    """소계 계산."""
    return sum(item['price'] * item['quantity'] for item in items)

def _calculate_tax(self, subtotal, tax_rate):
    """세금 계산 (대량 구매 할인 포함)."""
    base_tax = subtotal * tax_rate
    return base_tax * 0.95 if subtotal > 1000 else base_tax

def _apply_discount(self, amount, discount):
    """할인 적용."""
    return amount * (1 - discount / 100) if discount > 0 else amount
```

## Example 2: 클래스 추출 (Class Extraction)

**리팩토링 전**:
```python
class User:
    def __init__(self, name, email, password):
        self.name = name
        self.email = email
        self.password = password

    def validate_password(self):
        if len(self.password) < 8:
            return False
        if not any(c.isupper() for c in self.password):
            return False
        if not any(c.isdigit() for c in self.password):
            return False
        return True

    def hash_password(self):
        return bcrypt.hashpw(self.password.encode(), bcrypt.gensalt())
```

**리팩토링 후** (클래스 추출):
```python
class PasswordValidator:
    """비밀번호 검증."""
    MIN_LENGTH = 8

    @staticmethod
    def validate(password: str) -> bool:
        """비밀번호 검증."""
        return (
            len(password) >= PasswordValidator.MIN_LENGTH and
            any(c.isupper() for c in password) and
            any(c.isdigit() for c in password)
        )

class PasswordHasher:
    """비밀번호 해싱."""
    @staticmethod
    def hash(password: str) -> str:
        """비밀번호 해싱."""
        return bcrypt.hashpw(password.encode(), bcrypt.gensalt())

class User:
    def __init__(self, name: str, email: str, password: str):
        self.name = name
        self.email = email
        self.password = password

    def validate_password(self) -> bool:
        """비밀번호 검증."""
        return PasswordValidator.validate(self.password)

    def hash_password(self) -> str:
        """비밀번호 해싱."""
        return PasswordHasher.hash(self.password)
```

## Example 3: 함수형 리팩토링 (Functional Refactoring)

**리팩토링 전** (명령형):
```python
def process_orders(orders):
    """주문 처리."""
    valid_orders = []
    for order in orders:
        if order['status'] == 'pending' and order['amount'] > 0:
            valid_orders.append(order)

    processed_orders = []
    for order in valid_orders:
        order['total'] = order['amount'] * order['quantity']
        order['status'] = 'processed'
        processed_orders.append(order)

    return processed_orders
```

**리팩토링 후** (함수형):
```python
def process_orders(orders):
    """주문 처리."""
    return (
        orders
        | filter(lambda o: o['status'] == 'pending' and o['amount'] > 0)
        | map(lambda o: {**o, 'total': o['amount'] * o['quantity'], 'status': 'processed'})
        | list()
    )

# 또는 더 명확하게
from functools import reduce

def process_orders(orders):
    """주문 처리 (명확한 함수형)."""

    def is_valid_order(order):
        return order['status'] == 'pending' and order['amount'] > 0

    def calculate_total(order):
        return {
            **order,
            'total': order['amount'] * order['quantity'],
            'status': 'processed'
        }

    return list(map(calculate_total, filter(is_valid_order, orders)))
```

## Example 4: 디자인 패턴 도입 (Strategy Pattern)

**리팩토링 전**:
```python
class PaymentProcessor:
    def process_payment(self, order, method):
        """결제 처리 (조건문 지옥)."""
        if method == 'credit_card':
            # 신용카드 처리 로직
            validate_card(order.card)
            charge_card(order.card, order.amount)
            return 'processed'
        elif method == 'paypal':
            # PayPal 처리 로직
            validate_paypal(order.paypal_email)
            charge_paypal(order.paypal_email, order.amount)
            return 'processed'
        elif method == 'bank_transfer':
            # 계좌이체 처리 로직
            validate_bank(order.bank_account)
            transfer(order.bank_account, order.amount)
            return 'processed'
        else:
            raise ValueError(f"Unknown method: {method}")
```

**리팩토링 후** (Strategy Pattern):
```python
class PaymentStrategy:
    """결제 전략 인터페이스."""
    def validate(self, order): pass
    def charge(self, order): pass

class CreditCardStrategy(PaymentStrategy):
    """신용카드 결제 전략."""
    def validate(self, order):
        return validate_card(order.card)

    def charge(self, order):
        return charge_card(order.card, order.amount)

class PayPalStrategy(PaymentStrategy):
    """PayPal 결제 전략."""
    def validate(self, order):
        return validate_paypal(order.paypal_email)

    def charge(self, order):
        return charge_paypal(order.paypal_email, order.amount)

class PaymentProcessor:
    def __init__(self, strategy: PaymentStrategy):
        self.strategy = strategy

    def process_payment(self, order):
        """결제 처리."""
        self.strategy.validate(order)
        self.strategy.charge(order)
        return 'processed'
```

## Example 5: 기술 부채 제거 (Technical Debt Removal)

**리팩토링 전** (기술 부채):
```python
# 전역 변수 사용 (나쁜 패턴)
_global_cache = {}
_global_counter = 0

def add_item(item):
    """아이템 추가."""
    global _global_cache, _global_counter
    _global_cache[_global_counter] = item
    _global_counter += 1
    return _global_counter

def get_item(item_id):
    """아이템 조회."""
    return _global_cache.get(item_id)

def clear_cache():
    """캐시 초기화."""
    global _global_cache
    _global_cache = {}
```

**리팩토링 후** (기술 부채 제거):
```python
class ItemCache:
    """아이템 캐시 관리."""
    def __init__(self):
        self._cache = {}
        self._counter = 0

    def add_item(self, item):
        """아이템 추가."""
        self._counter += 1
        self._cache[self._counter] = item
        return self._counter

    def get_item(self, item_id):
        """아이템 조회."""
        return self._cache.get(item_id)

    def clear(self):
        """캐시 초기화."""
        self._cache.clear()
        self._counter = 0

    def get_size(self):
        """캐시 크기."""
        return len(self._cache)
```

## Example 6: 코드 복제 제거 (DRY Principle)

**리팩토링 전** (복제 코드):
```python
def validate_email_field(email):
    """이메일 필드 검증."""
    if not email:
        return False, "Email is required"
    if '@' not in email:
        return False, "Invalid email format"
    if '.' not in email:
        return False, "Invalid email format"
    return True, ""

def validate_phone_field(phone):
    """전화번호 필드 검증."""
    if not phone:
        return False, "Phone is required"
    if not phone.isdigit():
        return False, "Phone must contain only digits"
    if len(phone) < 10:
        return False, "Phone must be at least 10 digits"
    return True, ""

def validate_name_field(name):
    """이름 필드 검증."""
    if not name:
        return False, "Name is required"
    if len(name) < 2:
        return False, "Name must be at least 2 characters"
    return True, ""
```

**리팩토링 후** (DRY):
```python
class FieldValidator:
    """필드 검증 기본 클래스."""

    def validate(self, value):
        """검증 실행."""
        if not value:
            return False, f"{self.field_name} is required"

        errors = self.validate_specific(value)
        if errors:
            return False, errors

        return True, ""

    def validate_specific(self, value):
        """특정 필드 검증 (서브클래스에서 구현)."""
        raise NotImplementedError

class EmailValidator(FieldValidator):
    """이메일 검증."""
    field_name = "Email"

    def validate_specific(self, email):
        if '@' not in email or '.' not in email:
            return "Invalid email format"
        return ""

class PhoneValidator(FieldValidator):
    """전화번호 검증."""
    field_name = "Phone"

    def validate_specific(self, phone):
        if not phone.isdigit():
            return "Phone must contain only digits"
        if len(phone) < 10:
            return "Phone must be at least 10 digits"
        return ""

class NameValidator(FieldValidator):
    """이름 검증."""
    field_name = "Name"

    def validate_specific(self, name):
        if len(name) < 2:
            return "Name must be at least 2 characters"
        return ""
```

## Example 7: 성능 리팩토링 (Performance Refactoring)

**리팩토링 전** (비효율적):
```python
def find_duplicates(numbers):
    """중복 찾기 (O(n²) 복잡도)."""
    duplicates = []
    for i in range(len(numbers)):
        for j in range(i + 1, len(numbers)):
            if numbers[i] == numbers[j]:
                duplicates.append(numbers[i])
    return duplicates
```

**리팩토링 후** (효율적):
```python
def find_duplicates(numbers):
    """중복 찾기 (O(n) 복잡도)."""
    seen = set()
    duplicates = set()

    for num in numbers:
        if num in seen:
            duplicates.add(num)
        else:
            seen.add(num)

    return list(duplicates)
```

## Example 8: 테스트 가능성 개선 (Testability Refactoring)

**리팩토링 전** (테스트 불가능):
```python
class OrderService:
    def __init__(self):
        self.db = DatabaseConnection("prod_db")  # 실제 DB
        self.email = EmailService()  # 실제 이메일

    def create_order(self, order_data):
        """주문 생성."""
        order = Order(order_data)
        self.db.save(order)
        self.email.send_confirmation(order)
        return order
```

**리팩토링 후** (테스트 가능):
```python
class OrderService:
    def __init__(self, db: Database, email: EmailService):
        """주입된 의존성으로 초기화."""
        self.db = db
        self.email = email

    def create_order(self, order_data):
        """주문 생성."""
        order = Order(order_data)
        self.db.save(order)
        self.email.send_confirmation(order)
        return order

# 테스트
def test_create_order():
    mock_db = MockDatabase()
    mock_email = MockEmailService()
    service = OrderService(mock_db, mock_email)

    order = service.create_order({'id': 1, 'amount': 100})

    assert mock_db.saved_orders[-1].id == 1
    assert mock_email.sent_emails[-1].subject == "Order Confirmation"
```

---

**Last Updated**: 2025-11-22
**Total Examples**: 8 practical refactoring scenarios
**Focus**: Readable, Maintainable, SOLID principles
