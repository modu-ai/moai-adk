# Advanced Patterns - TRUST 4 Quality Framework

**Deep dive into TRUST 4 principles implementation and validation**

---

## TRUST 4 Framework Deep Dive

### Principle 1: Test First (T)

**The Red-Green-Refactor Cycle**:

```
RED Phase: Write failing test
├─ Test defines requirement
├─ Code doesn't exist yet
└─ Test fails as expected ✓

GREEN Phase: Write minimal code
├─ Simplest code to pass test
├─ Focus on making test pass
└─ Test now passes ✓

REFACTOR Phase: Improve quality
├─ Extract functions/classes
├─ Optimize performance
├─ Add documentation
├─ Keep tests passing ✓
```

**Implementation Example**:

```python
# RED: Write failing test first
def test_calculate_total_price_with_tax():
    item = ShoppingItem(name="Widget", price=10.00)
    total = calculate_total_with_tax(item, tax_rate=0.10)
    assert total == 11.00  # Fails - function doesn't exist

# GREEN: Minimal implementation
def calculate_total_with_tax(item, tax_rate):
    return item.price * (1 + tax_rate)

# REFACTOR: Improve code
def calculate_total_with_tax(item: ShoppingItem, tax_rate: float) -> float:
    """Calculate total price including tax.
    
    Args:
        item: Shopping item with price
        tax_rate: Tax rate as decimal (0.10 = 10%)
    
    Returns:
        Total price including tax
    
    Example:
        >>> item = ShoppingItem("Widget", 10.00)
        >>> calculate_total_with_tax(item, 0.10)
        11.0
    """
    if not 0 <= tax_rate <= 1:
        raise ValueError("Tax rate must be between 0 and 1")
    
    return item.price * (1 + tax_rate)
```

### Principle 2: Readable (R)

**Readability Metrics**:

| Metric | Target | Tool | Check |
|--------|--------|------|-------|
| Cyclomatic Complexity | ≤ 10 | pylint | Max 15 |
| Function Length | ≤ 50 lines | custom | Max 100 |
| Nesting Depth | ≤ 3 levels | pylint | Max 5 |
| Comment Ratio | 15-20% | custom | Min 10% |

**Readability Checklist**:

```
✓ Clear function/variable names (noun_verb pattern)
✓ Single responsibility principle
✓ Type hints on all parameters
✓ Docstrings with examples
✓ No magic numbers
✓ DRY principle applied
✓ SOLID principles followed
```

### Principle 3: Unified (U)

**Consistency Requirements**:

```
Architecture:
  ├─ Same pattern across all modules
  ├─ Same error handling approach
  ├─ Same logging strategy
  └─ Same naming conventions

Testing:
  ├─ Same test structure
  ├─ Same fixtures/factories
  ├─ Same assertion patterns
  └─ Same mock strategies

Documentation:
  ├─ Same docstring format
  ├─ Same README structure
  ├─ Same API documentation
  └─ Same changelog format
```

### Principle 4: Secured (S)

**OWASP Top 10 (2024) Compliance**:

```python
# 1. Broken Access Control
def get_user_profile(user_id: int, current_user: User) -> UserProfile:
    if user_id != current_user.id and not current_user.is_admin:
        raise UnauthorizedError("Cannot access other user profiles")
    return fetch_profile(user_id)

# 2. Cryptographic Failures
from bcrypt import hashpw, gensalt

def hash_password(plaintext: str) -> str:
    salt = gensalt(rounds=12)
    return hashpw(plaintext.encode('utf-8'), salt).decode('utf-8')

# 3. Injection Prevention
from sqlalchemy import text

def safe_query(user_input: str) -> List[User]:
    # Use parameterized queries
    query = text("SELECT * FROM users WHERE name = :name")
    return db.session.execute(query, {"name": user_input}).fetchall()

# 4. Insecure Design
# Implement threat modeling during design phase
class ThreatModel:
    threats = [
        "SQL Injection",
        "Authentication Bypass",
        "Privilege Escalation"
    ]
    mitigations = {
        "SQL Injection": "Use parameterized queries",
        "Authentication Bypass": "Implement RBAC",
        "Privilege Escalation": "Validate permissions"
    }

# 5. Security Misconfiguration
import os
from pathlib import Path

def load_config():
    config = {
        'DEBUG': os.getenv('DEBUG', 'false').lower() == 'true',
        'DATABASE_URL': os.getenv('DATABASE_URL'),  # Never hardcode
        'SECRET_KEY': os.getenv('SECRET_KEY'),      # Load from env
        'ALLOWED_HOSTS': os.getenv('ALLOWED_HOSTS', '').split(',')
    }
    
    # Validate required configs
    if not config['DATABASE_URL']:
        raise ValueError("DATABASE_URL required")
    
    return config
```

---

## TRUST 4 CI/CD Integration

### Quality Gate Pipeline

```yaml
name: TRUST 4 Quality Gates

on: [push, pull_request]

jobs:
  test-first:
    name: "T: Test Coverage"
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests with coverage
        run: pytest --cov=src --cov-fail-under=85
  
  readable:
    name: "R: Code Quality"
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Pylint
        run: pylint src/ --fail-under=8.0
      - name: Black format check
        run: black --check src/
  
  unified:
    name: "U: Consistency Check"
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Architecture validation
        run: python .moai/scripts/validate_architecture.py
  
  secured:
    name: "S: Security Check"
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Bandit security scan
        run: bandit -r src/ -ll
      - name: Dependency audit
        run: pip audit
```

---

## TRUST 4 Validation Framework

### 4-Step Validation

```python
class TRUSTValidator:
    def validate(self, code: str) -> ValidationResult:
        """Validate code against TRUST 4 principles."""
        
        result = ValidationResult()
        
        # T: Test First
        result.test_coverage = self._measure_coverage(code)
        result.test_passed = result.test_coverage >= 85
        
        # R: Readable
        result.complexity = self._measure_complexity(code)
        result.readable_passed = result.complexity <= 10
        
        # U: Unified
        result.consistency = self._check_consistency(code)
        result.unified_passed = result.consistency > 90
        
        # S: Secured
        result.security_issues = self._security_scan(code)
        result.secured_passed = len(result.security_issues) == 0
        
        result.overall_passed = all([
            result.test_passed,
            result.readable_passed,
            result.unified_passed,
            result.secured_passed
        ])
        
        return result
```

---

## TRUST 4 Metrics

### Dashboard Metrics

```python
class TRUSTMetrics:
    def __init__(self):
        self.test_coverage = 0  # Target: >= 85%
        self.code_quality = 0   # Target: >= 8.0
        self.consistency_score = 0  # Target: >= 90%
        self.security_score = 100   # Target: 100
    
    def get_overall_score(self) -> int:
        weights = {'test': 0.25, 'quality': 0.25, 'consistency': 0.25, 'security': 0.25}
        return (
            self.test_coverage * weights['test'] +
            self.code_quality * weights['quality'] +
            self.consistency_score * weights['consistency'] +
            self.security_score * weights['security']
        )
    
    def is_production_ready(self) -> bool:
        return (
            self.test_coverage >= 85 and
            self.code_quality >= 8.0 and
            self.consistency_score >= 90 and
            self.security_score == 100
        )
```

---

## TRUST 4 in Different Domains

### Backend Implementation

```python
# T: Test first - Unit tests for business logic
def test_process_payment():
    order = create_order(items=[...])
    result = process_payment(order, payment_method="credit_card")
    assert result.status == "completed"

# R: Readable - Clear function names
def calculate_total_with_tax(items: List[Item], tax_rate: float) -> float:
    subtotal = sum(item.price for item in items)
    return subtotal * (1 + tax_rate)

# U: Unified - Same error handling
class PaymentError(Exception):
    pass

def process_payment(order: Order) -> PaymentResult:
    try:
        # Process payment
        pass
    except Exception as e:
        raise PaymentError(f"Payment failed: {e}") from e

# S: Secured - Input validation
def validate_amount(amount: float) -> bool:
    if amount <= 0:
        raise ValueError("Amount must be positive")
    if amount > 1000000:
        raise ValueError("Amount exceeds maximum")
    return True
```

### Frontend Implementation

```typescript
// T: Test first - Component tests
test("Button click handler called correctly", () => {
  const mockHandler = jest.fn();
  render(<Button onClick={mockHandler}>Click</Button>);
  fireEvent.click(screen.getByText("Click"));
  expect(mockHandler).toHaveBeenCalledOnce();
});

// R: Readable - Clear component structure
function UserProfile({ userId }: { userId: string }) {
  const [user, setUser] = useState<User | null>(null);
  
  useEffect(() => {
    fetchUser(userId).then(setUser);
  }, [userId]);
  
  if (!user) return <Loading />;
  return <div>{user.name}</div>;
}

// U: Unified - Same patterns across components
const useUserData = (userId: string) => {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  useEffect(() => {
    fetchData().then(setData).catch(setError).finally(() => setLoading(false));
  }, [userId]);
  
  return { data, loading, error };
};

// S: Secured - XSS prevention
function SafeDisplay({ content }: { content: string }) {
  return <div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(content) }} />;
}
```

---

## TRUST 4 Checklists

### Pre-Commit Checklist

- [ ] Tests written before code
- [ ] All tests passing
- [ ] Test coverage >= 85%
- [ ] Code is readable (pylint >= 8.0)
- [ ] No magic numbers or hardcoded values
- [ ] Docstrings added
- [ ] Following code style
- [ ] No security issues
- [ ] No hardcoded secrets
- [ ] Error handling implemented

### Code Review Checklist

- [ ] Requirement clarity (TRUST T)
- [ ] Test coverage adequate (TRUST T)
- [ ] Code readability (TRUST R)
- [ ] Consistency with codebase (TRUST U)
- [ ] Security implications considered (TRUST S)
- [ ] Performance impact assessed
- [ ] Documentation updated
- [ ] No breaking changes

---

**Last Updated**: 2025-11-22
**Status**: Production Ready
**Framework**: TRUST 4 Quality Assurance
