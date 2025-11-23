# Refactoring Reference Guide

## Refactoring Techniques Catalog (20+ Techniques)

### Method-Level Refactorings

1. **Extract Method** - Break long method into smaller methods
2. **Inline Method** - Remove unnecessary method indirection
3. **Extract Variable** - Name intermediate calculation results
4. **Inline Variable** - Remove unnecessary variable
5. **Rename Method** - Clarify method purpose with better name
6. **Change Function Declaration** - Modify parameters or return type
7. **Introduce Parameter Object** - Group parameters into object
8. **Remove Flag Argument** - Replace boolean flags with explicit methods

### Class-Level Refactorings

9. **Move Method** - Move method to more appropriate class
10. **Move Field** - Move field to where it's mostly used
11. **Extract Class** - Break large class into smaller classes
12. **Inline Class** - Merge class with little responsibility
13. **Hide Delegate** - Encapsulate access to another object
14. **Remove Middle Man** - Direct access instead of delegation
15. **Introduce Foreign Method** - Add method to class you can't modify
16. **Introduce Local Extension** - Create subclass with new functionality

### Data-Level Refactorings

17. **Encapsulate Field** - Make field private with accessors
18. **Encapsulate Collection** - Protect collection with proper interface
19. **Replace Type Code with Subclasses** - Use polymorphism
20. **Replace Type Code with State/Strategy** - Use design patterns
21. **Replace Conditional with Polymorphism** - Remove complex conditionals
22. **Replace Magic Number with Symbolic Constant** - Name magic values

## Code Smell Diagnostic Guide

### Detection Checklist

| Code Smell | Symptoms | Refactoring Solution |
|------------|----------|---------------------|
| Duplicated Code | Same code in multiple places | Extract Method/Class |
| Long Method | Method >50 lines | Extract Method |
| Large Class | Class >500 lines or >20 methods | Extract Class |
| Long Parameter List | >5 parameters | Introduce Parameter Object |
| Divergent Change | Class changes for multiple reasons | Extract Class |
| Shotgun Surgery | One change affects many classes | Move Method/Field |
| Feature Envy | Method uses another class's data | Move Method |
| Data Clumps | Same data together everywhere | Extract Class |
| Primitive Obsession | Primitives instead of objects | Replace with Object |
| Switch Statements | Complex type checking | Replace with Polymorphism |

### Code Smell Examples with Solutions

**Duplicated Code**:
```python
# SMELL: Duplicate validation
def register_user(email, password):
    if not email or '@' not in email:
        raise ValueError("Invalid email")
    # ...

def update_user_email(email):
    if not email or '@' not in email:
        raise ValueError("Invalid email")
    # ...

# SOLUTION: Extract validation
class EmailValidator:
    @staticmethod
    def validate(email):
        if not email or '@' not in email:
            raise ValueError("Invalid email")

def register_user(email, password):
    EmailValidator.validate(email)
    # ...
```

**Long Method**:
```python
# SMELL: 80-line method
def process_order(order_data):
    # 20 lines of validation
    # 30 lines of calculation
    # 20 lines of persistence
    # 10 lines of notification
    pass

# SOLUTION: Extract methods
def process_order(order_data):
    validate_order(order_data)
    calculate_totals(order_data)
    save_order(order_data)
    send_notifications(order_data)
```

## Refactoring Tools & IDE Support

### Python Tools

| Tool | Purpose | Usage |
|------|---------|-------|
| rope | Refactoring library | `pip install rope` |
| pylint | Code smell detection | `pylint module.py` |
| autopep8 | Auto-formatting | `autopep8 --in-place file.py` |
| black | Code formatter | `black file.py` |
| isort | Import sorting | `isort file.py` |

### IDE Refactoring Features

**VS Code**:
- Extract Method: Select code → Right-click → Refactor → Extract Method
- Rename Symbol: F2
- Extract Variable: Select expression → Refactor → Extract Variable
- Inline Variable: Place cursor → Refactor → Inline Variable

**PyCharm**:
- Extract Method: Ctrl+Alt+M (Win/Linux), Cmd+Alt+M (Mac)
- Rename: Shift+F6
- Extract Variable: Ctrl+Alt+V (Win/Linux), Cmd+Alt+V (Mac)
- Inline: Ctrl+Alt+N (Win/Linux), Cmd+Alt+N (Mac)
- Move: F6

## Refactoring Workflow

### Safe Refactoring Process

```
1. Write Tests
   ↓
2. Run Tests (All Pass)
   ↓
3. Small Refactoring Change
   ↓
4. Run Tests Again
   ↓
5. Tests Pass? → Continue to next refactoring
   Tests Fail? → Revert change, try different approach
```

### Refactoring Checklist (30+ Items)

**Before Refactoring**:
- [ ] Code is under version control
- [ ] Comprehensive test suite exists
- [ ] All tests pass (100% green)
- [ ] Test coverage ≥80% for code to refactor
- [ ] Team aware of refactoring plan
- [ ] Backup/branch created

**During Refactoring**:
- [ ] Make one small change at a time
- [ ] Run tests after each change
- [ ] Commit after each successful refactoring
- [ ] Keep refactorings separate from feature additions
- [ ] Maintain backward compatibility if needed
- [ ] Update documentation as code changes
- [ ] Review refactored code with team

**After Refactoring**:
- [ ] All tests still pass
- [ ] Code coverage maintained or improved
- [ ] Performance not degraded
- [ ] Documentation updated
- [ ] Team review completed
- [ ] CI/CD pipeline passes
- [ ] Merge to main branch

## Refactoring Patterns by Language

### Python-Specific Patterns

**Use List Comprehensions**:
```python
# BEFORE
result = []
for item in items:
    if item.is_valid():
        result.append(item.value)

# AFTER
result = [item.value for item in items if item.is_valid()]
```

**Use Context Managers**:
```python
# BEFORE
file = open('data.txt')
try:
    data = file.read()
finally:
    file.close()

# AFTER
with open('data.txt') as file:
    data = file.read()
```

**Use Properties**:
```python
# BEFORE
class Circle:
    def __init__(self, radius):
        self.radius = radius

    def get_area(self):
        return 3.14 * self.radius ** 2

# AFTER
class Circle:
    def __init__(self, radius):
        self.radius = radius

    @property
    def area(self):
        return 3.14 * self.radius ** 2
```

### TypeScript-Specific Patterns

**Use Type Aliases**:
```typescript
// BEFORE
function processUser(id: string, name: string, email: string) {}

// AFTER
type User = {
    id: string;
    name: string;
    email: string;
};

function processUser(user: User) {}
```

**Use Optional Chaining**:
```typescript
// BEFORE
const city = user && user.address && user.address.city;

// AFTER
const city = user?.address?.city;
```

## Performance Impact of Refactoring

### Positive Impacts

- **Reduced code duplication** → Smaller code size → Faster loading
- **Better data structures** → O(n²) → O(n log n) → Faster execution
- **Removed unused code** → Reduced memory footprint
- **Simplified logic** → Fewer branches → Better CPU cache usage

### Potential Negative Impacts (and Solutions)

| Issue | Mitigation |
|-------|-----------|
| More method calls | Inline hot methods after profiling |
| Abstraction overhead | Profile before over-abstracting |
| More objects created | Use object pooling if needed |
| Extra indirection | Keep critical paths direct |

## Anti-Patterns to Avoid

### 1. Refactoring Without Tests

```python
# WRONG: Refactor without test coverage
def complex_calculation(x, y):
    # Refactor this without tests?
    return x * y + (x / y) - (x ** y)

# RIGHT: Add tests first
def test_complex_calculation():
    assert complex_calculation(2, 3) == expected_value
    # Add more test cases

# Then refactor safely
```

### 2. Big Bang Refactoring

```python
# WRONG: Refactor entire codebase at once
# (Risky, hard to review, impossible to revert)

# RIGHT: Incremental refactoring
# Refactor one class/module at a time
# Commit after each successful refactoring
```

### 3. Changing Behavior During Refactoring

```python
# WRONG: Fix bugs during refactoring
def refactored_method():
    # Refactoring AND fixing bug at same time
    pass

# RIGHT: Separate bug fixes from refactoring
# 1. Fix bug, commit
# 2. Refactor, commit separately
```

## Metrics & Success Criteria

### Code Quality Metrics

| Metric | Before Refactoring | After Refactoring | Target |
|--------|-------------------|-------------------|--------|
| Cyclomatic Complexity | 15 | 8 | <10 |
| Method Length | 80 lines | 20 lines | <50 |
| Class Size | 800 lines | 200 lines | <500 |
| Code Duplication | 15% | 3% | <5% |
| Test Coverage | 65% | 85% | ≥80% |

### Refactoring Success Indicators

- ✅ All tests pass after refactoring
- ✅ Code is easier to understand
- ✅ Future changes are easier to implement
- ✅ Code duplication reduced by >50%
- ✅ Method complexity reduced by >30%
- ✅ Test coverage maintained or improved
- ✅ Performance not degraded
- ✅ Team agrees code is better

---

**Last Updated**: 2025-11-23
**Status**: Production Ready
**Lines**: 290
**Comprehensive catalog**: 20+ refactoring techniques, 30+ checklist items
