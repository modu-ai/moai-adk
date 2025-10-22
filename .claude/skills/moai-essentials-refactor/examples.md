# Refactoring Examples: Real-World Scenarios

> **Version**: 2.0.0 (2025-10-22)
> **Based on**: Martin Fowler's Refactoring (2nd Edition), SOLID principles, modern best practices

This document provides concrete refactoring examples across multiple languages and scenarios, demonstrating how to apply refactoring patterns safely and effectively.

---

## Table of Contents

1. [Extract Function: Breaking Down Complex Methods](#1-extract-function-breaking-down-complex-methods)
2. [Replace Conditional with Polymorphism](#2-replace-conditional-with-polymorphism)
3. [Introduce Parameter Object](#3-introduce-parameter-object)
4. [Replace Loop with Pipeline](#4-replace-loop-with-pipeline)
5. [Encapsulate Collection](#5-encapsulate-collection)
6. [Replace Nested Conditional with Guard Clauses](#6-replace-nested-conditional-with-guard-clauses)
7. [Pull Up Method: Extracting Common Behavior](#7-pull-up-method-extracting-common-behavior)
8. [Split Phase: Separating Concerns](#8-split-phase-separating-concerns)
9. [Introduce Special Case (Null Object Pattern)](#9-introduce-special-case-null-object-pattern)
10. [Decompose Conditional: Simplifying Complex Logic](#10-decompose-conditional-simplifying-complex-logic)

---

## 1. Extract Function: Breaking Down Complex Methods

**Problem**: Long method doing multiple things, violating Single Responsibility Principle.

### Before (Python)

```python
def process_order(order_data):
    """Process customer order with validation and calculation."""
    # Validation
    if not order_data.get('customer_id'):
        raise ValueError("Customer ID required")
    if not order_data.get('items'):
        raise ValueError("No items in order")

    # Calculate totals
    subtotal = 0
    for item in order_data['items']:
        subtotal += item['price'] * item['quantity']

    # Apply discounts
    discount = 0
    if order_data.get('coupon_code'):
        if order_data['coupon_code'] == 'SAVE10':
            discount = subtotal * 0.1
        elif order_data['coupon_code'] == 'SAVE20':
            discount = subtotal * 0.2

    # Calculate tax
    tax_rate = 0.08
    tax = (subtotal - discount) * tax_rate

    # Calculate final total
    total = subtotal - discount + tax

    # Save to database
    db.orders.insert({
        'customer_id': order_data['customer_id'],
        'subtotal': subtotal,
        'discount': discount,
        'tax': tax,
        'total': total,
        'status': 'pending'
    })

    return total
```

### After (Python)

```python
# @CODE:REFACTOR-001 | SPEC: SPEC-REFACTOR-001.md | TEST: tests/refactor/test_order.py

def process_order(order_data):
    """Process customer order with validation and calculation."""
    validate_order(order_data)

    subtotal = calculate_subtotal(order_data['items'])
    discount = calculate_discount(subtotal, order_data.get('coupon_code'))
    tax = calculate_tax(subtotal - discount)
    total = subtotal - discount + tax

    save_order(order_data['customer_id'], subtotal, discount, tax, total)

    return total


def validate_order(order_data):
    """Validate order has required fields."""
    if not order_data.get('customer_id'):
        raise ValueError("Customer ID required")
    if not order_data.get('items'):
        raise ValueError("No items in order")


def calculate_subtotal(items):
    """Calculate subtotal from line items."""
    return sum(item['price'] * item['quantity'] for item in items)


def calculate_discount(subtotal, coupon_code):
    """Apply coupon discount to subtotal."""
    discount_rates = {
        'SAVE10': 0.1,
        'SAVE20': 0.2
    }
    rate = discount_rates.get(coupon_code, 0)
    return subtotal * rate


def calculate_tax(taxable_amount, tax_rate=0.08):
    """Calculate tax on taxable amount."""
    return taxable_amount * tax_rate


def save_order(customer_id, subtotal, discount, tax, total):
    """Persist order to database."""
    db.orders.insert({
        'customer_id': customer_id,
        'subtotal': subtotal,
        'discount': discount,
        'tax': tax,
        'total': total,
        'status': 'pending'
    })
```

**Benefits**:
- Each function has single responsibility
- Functions ≤ 10 lines each
- Easy to test in isolation
- Clear intent from function names

---

## 2. Replace Conditional with Polymorphism

**Problem**: Type code with switch/case that's hard to extend (violates Open/Closed Principle).

### Before (TypeScript)

```typescript
interface Employee {
  type: 'engineer' | 'manager' | 'salesperson';
  monthlySalary: number;
  commission: number;
  bonus: number;
}

function getPayAmount(employee: Employee): number {
  switch (employee.type) {
    case 'engineer':
      return employee.monthlySalary;
    case 'manager':
      return employee.monthlySalary + employee.bonus;
    case 'salesperson':
      return employee.monthlySalary + employee.commission;
    default:
      throw new Error('Unknown employee type');
  }
}
```

### After (TypeScript)

```typescript
// @CODE:REFACTOR-002:DOMAIN | SPEC: SPEC-REFACTOR-002.md

abstract class Employee {
  protected monthlySalary: number;

  constructor(monthlySalary: number) {
    this.monthlySalary = monthlySalary;
  }

  abstract getPayAmount(): number;
}

class Engineer extends Employee {
  getPayAmount(): number {
    return this.monthlySalary;
  }
}

class Manager extends Employee {
  private bonus: number;

  constructor(monthlySalary: number, bonus: number) {
    super(monthlySalary);
    this.bonus = bonus;
  }

  getPayAmount(): number {
    return this.monthlySalary + this.bonus;
  }
}

class Salesperson extends Employee {
  private commission: number;

  constructor(monthlySalary: number, commission: number) {
    super(monthlySalary);
    this.commission = commission;
  }

  getPayAmount(): number {
    return this.monthlySalary + this.commission;
  }
}

// Usage
const engineer = new Engineer(5000);
const manager = new Manager(6000, 1000);
const salesperson = new Salesperson(4000, 2000);

console.log(engineer.getPayAmount());      // 5000
console.log(manager.getPayAmount());       // 7000
console.log(salesperson.getPayAmount());   // 6000
```

**Benefits**:
- Open for extension (add new employee types)
- Closed for modification (no switch/case changes)
- Each class handles its own logic
- Type-safe polymorphism

---

## 3. Introduce Parameter Object

**Problem**: Functions with too many parameters (violates parameter limit ≤5).

### Before (Go)

```go
func CreateInvoice(
    customerName string,
    customerEmail string,
    customerAddress string,
    invoiceNumber string,
    invoiceDate time.Time,
    items []LineItem,
    taxRate float64,
    discountPercent float64,
) *Invoice {
    // Implementation...
}
```

### After (Go)

```go
// @CODE:REFACTOR-003:DATA | SPEC: SPEC-REFACTOR-003.md | TEST: tests/refactor/invoice_test.go

type Customer struct {
    Name    string
    Email   string
    Address string
}

type InvoiceParams struct {
    Customer        Customer
    InvoiceNumber   string
    InvoiceDate     time.Time
    Items           []LineItem
    TaxRate         float64
    DiscountPercent float64
}

func CreateInvoice(params InvoiceParams) *Invoice {
    invoice := &Invoice{
        Customer:      params.Customer,
        Number:        params.InvoiceNumber,
        Date:          params.InvoiceDate,
        Items:         params.Items,
    }

    invoice.CalculateTotal(params.TaxRate, params.DiscountPercent)

    return invoice
}

// Usage
invoice := CreateInvoice(InvoiceParams{
    Customer: Customer{
        Name:    "John Doe",
        Email:   "john@example.com",
        Address: "123 Main St",
    },
    InvoiceNumber:   "INV-001",
    InvoiceDate:     time.Now(),
    Items:           lineItems,
    TaxRate:         0.08,
    DiscountPercent: 10,
})
```

**Benefits**:
- Single parameter (meets ≤5 constraint)
- Related data grouped logically
- Easy to extend with new fields
- Better IDE autocomplete support

---

## 4. Replace Loop with Pipeline

**Problem**: Imperative loops that are hard to read (modern languages offer better alternatives).

### Before (JavaScript)

```javascript
function processOrders(orders) {
  const activeOrders = [];

  for (let i = 0; i < orders.length; i++) {
    if (orders[i].status === 'active') {
      activeOrders.push(orders[i]);
    }
  }

  const sortedOrders = activeOrders.sort((a, b) => {
    return b.total - a.total;
  });

  const topOrders = [];
  for (let i = 0; i < Math.min(5, sortedOrders.length); i++) {
    topOrders.push(sortedOrders[i]);
  }

  const orderTotals = [];
  for (let i = 0; i < topOrders.length; i++) {
    orderTotals.push(topOrders[i].total);
  }

  return orderTotals;
}
```

### After (JavaScript)

```javascript
// @CODE:REFACTOR-004 | SPEC: SPEC-REFACTOR-004.md | TEST: tests/refactor/order-pipeline.test.js

function processOrders(orders) {
  return orders
    .filter(order => order.status === 'active')
    .sort((a, b) => b.total - a.total)
    .slice(0, 5)
    .map(order => order.total);
}
```

**Benefits**:
- 80% fewer lines of code
- Declarative intent (what, not how)
- No intermediate variables
- Chainable transformations

---

## 5. Encapsulate Collection

**Problem**: Direct access to internal collections breaks encapsulation.

### Before (Java)

```java
public class Team {
    private List<Player> players;

    public Team() {
        this.players = new ArrayList<>();
    }

    public List<Player> getPlayers() {
        return players; // DANGER: Caller can modify directly!
    }
}

// Client code can break invariants
Team team = new Team();
team.getPlayers().clear(); // Oops! Direct modification
```

### After (Java)

```java
// @CODE:REFACTOR-005:DOMAIN | SPEC: SPEC-REFACTOR-005.md | TEST: tests/refactor/TeamTest.java

public class Team {
    private List<Player> players;

    public Team() {
        this.players = new ArrayList<>();
    }

    // Return defensive copy
    public List<Player> getPlayers() {
        return Collections.unmodifiableList(players);
    }

    // Controlled mutation methods
    public void addPlayer(Player player) {
        if (player == null) {
            throw new IllegalArgumentException("Player cannot be null");
        }
        if (players.size() >= 15) {
            throw new IllegalStateException("Team is full");
        }
        players.add(player);
    }

    public boolean removePlayer(Player player) {
        return players.remove(player);
    }

    public int getPlayerCount() {
        return players.size();
    }
}
```

**Benefits**:
- Collection invariants protected
- Clear API for modifications
- Easy to add validation rules
- Prevents accidental corruption

---

## 6. Replace Nested Conditional with Guard Clauses

**Problem**: Deep nesting makes code hard to follow (cyclomatic complexity > 10).

### Before (Rust)

```rust
fn calculate_discount(customer: &Customer, order: &Order) -> f64 {
    if customer.is_active {
        if customer.loyalty_points > 1000 {
            if order.total > 100.0 {
                if order.items.len() > 5 {
                    return order.total * 0.15;
                } else {
                    return order.total * 0.10;
                }
            } else {
                return order.total * 0.05;
            }
        } else {
            if order.total > 50.0 {
                return order.total * 0.05;
            } else {
                return 0.0;
            }
        }
    } else {
        return 0.0;
    }
}
```

### After (Rust)

```rust
// @CODE:REFACTOR-006 | SPEC: SPEC-REFACTOR-006.md | TEST: tests/refactor/discount_test.rs

fn calculate_discount(customer: &Customer, order: &Order) -> f64 {
    // Guard clauses handle edge cases early
    if !customer.is_active {
        return 0.0;
    }

    if customer.loyalty_points <= 1000 {
        return calculate_basic_discount(order);
    }

    calculate_premium_discount(order)
}

fn calculate_basic_discount(order: &Order) -> f64 {
    if order.total > 50.0 {
        order.total * 0.05
    } else {
        0.0
    }
}

fn calculate_premium_discount(order: &Order) -> f64 {
    if order.total <= 100.0 {
        return order.total * 0.05;
    }

    if order.items.len() > 5 {
        order.total * 0.15
    } else {
        order.total * 0.10
    }
}
```

**Benefits**:
- Maximum nesting depth: 2 (was 5)
- Cyclomatic complexity: 4 per function (was 12)
- Early returns reduce cognitive load
- Clear separation of logic paths

---

## 7. Pull Up Method: Extracting Common Behavior

**Problem**: Duplicate code across subclasses (DRY violation).

### Before (C#)

```csharp
public class CreditCardPayment : Payment
{
    public override decimal CalculateFee()
    {
        decimal baseFee = Amount * 0.029m + 0.30m;
        if (Amount > 1000)
            baseFee *= 0.95m; // 5% discount for large transactions
        return baseFee;
    }
}

public class DebitCardPayment : Payment
{
    public override decimal CalculateFee()
    {
        decimal baseFee = Amount * 0.015m + 0.25m;
        if (Amount > 1000)
            baseFee *= 0.95m; // Same discount logic (duplicate!)
        return baseFee;
    }
}
```

### After (C#)

```csharp
// @CODE:REFACTOR-007:DOMAIN | SPEC: SPEC-REFACTOR-007.md | TEST: tests/Refactor/PaymentTests.cs

public abstract class Payment
{
    public decimal Amount { get; set; }

    public decimal CalculateFee()
    {
        decimal baseFee = CalculateBaseFee();
        return ApplyLargeTransactionDiscount(baseFee);
    }

    // Template method pattern
    protected abstract decimal CalculateBaseFee();

    // Common behavior pulled up
    private decimal ApplyLargeTransactionDiscount(decimal fee)
    {
        const decimal LARGE_TRANSACTION_THRESHOLD = 1000m;
        const decimal DISCOUNT_MULTIPLIER = 0.95m;

        if (Amount > LARGE_TRANSACTION_THRESHOLD)
        {
            return fee * DISCOUNT_MULTIPLIER;
        }
        return fee;
    }
}

public class CreditCardPayment : Payment
{
    protected override decimal CalculateBaseFee()
    {
        return Amount * 0.029m + 0.30m;
    }
}

public class DebitCardPayment : Payment
{
    protected override decimal CalculateBaseFee()
    {
        return Amount * 0.015m + 0.25m;
    }
}
```

**Benefits**:
- Discount logic in one place
- Easy to modify discount rules
- Template method pattern applied
- Subclasses focus on their unique logic

---

## 8. Split Phase: Separating Concerns

**Problem**: Function does multiple unrelated tasks (violates Single Responsibility).

### Before (Python)

```python
def generate_report(user_id, report_type):
    """Generate and send report to user."""
    # Phase 1: Fetch data
    user = db.users.find_one({'id': user_id})
    if report_type == 'sales':
        data = db.sales.find({'user_id': user_id})
    elif report_type == 'inventory':
        data = db.inventory.find({'user_id': user_id})

    # Phase 2: Format data
    if report_type == 'sales':
        report = format_sales_report(data)
    elif report_type == 'inventory':
        report = format_inventory_report(data)

    # Phase 3: Send email
    subject = f"{report_type.title()} Report"
    send_email(user['email'], subject, report)

    return report
```

### After (Python)

```python
# @CODE:REFACTOR-008 | SPEC: SPEC-REFACTOR-008.md | TEST: tests/refactor/test_report.py

from dataclasses import dataclass
from typing import List, Dict

@dataclass
class ReportData:
    """Intermediate data structure between phases."""
    user_email: str
    report_type: str
    raw_data: List[Dict]


def generate_report(user_id: str, report_type: str) -> str:
    """Orchestrate report generation (split into clear phases)."""
    report_data = fetch_report_data(user_id, report_type)
    formatted_report = format_report(report_data)
    send_report(report_data.user_email, report_data.report_type, formatted_report)
    return formatted_report


def fetch_report_data(user_id: str, report_type: str) -> ReportData:
    """Phase 1: Fetch data from database."""
    user = db.users.find_one({'id': user_id})

    if report_type == 'sales':
        raw_data = list(db.sales.find({'user_id': user_id}))
    elif report_type == 'inventory':
        raw_data = list(db.inventory.find({'user_id': user_id}))
    else:
        raise ValueError(f"Unknown report type: {report_type}")

    return ReportData(
        user_email=user['email'],
        report_type=report_type,
        raw_data=raw_data
    )


def format_report(report_data: ReportData) -> str:
    """Phase 2: Format data into report."""
    if report_data.report_type == 'sales':
        return format_sales_report(report_data.raw_data)
    elif report_data.report_type == 'inventory':
        return format_inventory_report(report_data.raw_data)
    else:
        raise ValueError(f"Unknown report type: {report_data.report_type}")


def send_report(user_email: str, report_type: str, content: str) -> None:
    """Phase 3: Send report via email."""
    subject = f"{report_type.title()} Report"
    send_email(user_email, subject, content)
```

**Benefits**:
- Each phase tested independently
- Clear data flow via ReportData
- Easy to replace phases (e.g., send to Slack instead of email)
- Follows transformation pipeline pattern

---

## 9. Introduce Special Case (Null Object Pattern)

**Problem**: Repeated null/undefined checks clutter code.

### Before (TypeScript)

```typescript
function displayUserProfile(user: User | null) {
  const name = user ? user.name : 'Guest';
  const email = user ? user.email : 'No email';
  const role = user ? user.role : 'visitor';
  const loginCount = user ? user.loginCount : 0;

  console.log(`Name: ${name}`);
  console.log(`Email: ${email}`);
  console.log(`Role: ${role}`);
  console.log(`Logins: ${loginCount}`);
}
```

### After (TypeScript)

```typescript
// @CODE:REFACTOR-009:DOMAIN | SPEC: SPEC-REFACTOR-009.md | TEST: tests/refactor/user.test.ts

interface User {
  name: string;
  email: string;
  role: string;
  loginCount: number;
  isGuest(): boolean;
}

class RegisteredUser implements User {
  constructor(
    public name: string,
    public email: string,
    public role: string,
    public loginCount: number
  ) {}

  isGuest(): boolean {
    return false;
  }
}

// Special case object for null/missing user
class GuestUser implements User {
  name = 'Guest';
  email = 'No email';
  role = 'visitor';
  loginCount = 0;

  isGuest(): boolean {
    return true;
  }
}

function displayUserProfile(user: User) {
  // No null checks needed!
  console.log(`Name: ${user.name}`);
  console.log(`Email: ${user.email}`);
  console.log(`Role: ${user.role}`);
  console.log(`Logins: ${user.loginCount}`);
}

// Usage
const user = fetchUser(userId) ?? new GuestUser();
displayUserProfile(user);
```

**Benefits**:
- Eliminates null checks
- Polymorphic behavior for edge cases
- Type-safe (no null/undefined)
- Follows Null Object Pattern

---

## 10. Decompose Conditional: Simplifying Complex Logic

**Problem**: Complex boolean expressions that are hard to understand.

### Before (Ruby)

```ruby
def eligible_for_premium_shipping?(order, customer)
  if (order.total > 50 && customer.membership_level == 'gold') ||
     (order.total > 100 && customer.membership_level == 'silver') ||
     (order.total > 200 && customer.loyalty_points > 500) ||
     (customer.recent_orders_count > 10 && order.weight < 5)
    return true
  else
    return false
  end
end
```

### After (Ruby)

```ruby
# @CODE:REFACTOR-010 | SPEC: SPEC-REFACTOR-010.md | TEST: tests/refactor/shipping_spec.rb

def eligible_for_premium_shipping?(order, customer)
  gold_member_eligible?(order, customer) ||
    silver_member_eligible?(order, customer) ||
    loyal_customer_eligible?(order, customer) ||
    frequent_buyer_eligible?(order, customer)
end

def gold_member_eligible?(order, customer)
  customer.membership_level == 'gold' && order.total > 50
end

def silver_member_eligible?(order, customer)
  customer.membership_level == 'silver' && order.total > 100
end

def loyal_customer_eligible?(order, customer)
  order.total > 200 && customer.loyalty_points > 500
end

def frequent_buyer_eligible?(order, customer)
  customer.recent_orders_count > 10 && order.weight < 5
end
```

**Benefits**:
- Each condition has clear intent
- Easy to modify individual rules
- Testable in isolation
- Self-documenting code

---

## Integration with MoAI-ADK

### TDD Workflow for Refactoring

```bash
# RED: Write test for desired behavior
# Create .moai/specs/SPEC-REFACTOR-011/spec.md
/alfred:1-plan "Refactor payment calculation module"

# GREEN: Apply refactoring pattern
/alfred:2-run SPEC-REFACTOR-011

# REFACTOR: Clean up and verify
/alfred:3-sync auto
```

### Quality Gates

Before committing refactored code:

1. **Test Coverage**: Maintain ≥85%
2. **Complexity**: Functions ≤10 cyclomatic complexity
3. **Length**: Functions ≤50 LOC, files ≤300 LOC
4. **Linting**: Pass all language-specific linters
5. **TAG Integrity**: Update @TAG references

---

## Common Refactoring Smells to Watch For

| Code Smell | Refactoring Pattern | Section |
|------------|-------------------|---------|
| Long Method | Extract Function | #1 |
| Large Class | Extract Class | Similar to #1 |
| Switch Statements | Replace Conditional with Polymorphism | #2 |
| Long Parameter List | Introduce Parameter Object | #3 |
| Loops | Replace Loop with Pipeline | #4 |
| Primitive Obsession | Replace Primitive with Object | #3 |
| Data Clumps | Introduce Parameter Object | #3 |
| Duplicate Code | Pull Up Method, Extract Function | #7, #1 |
| Divergent Change | Split Phase, Extract Class | #8 |
| Shotgun Surgery | Move Method, Inline Class | - |
| Feature Envy | Move Method | - |
| Deep Nesting | Replace Nested Conditional with Guard Clauses | #6 |
| Null Checks Everywhere | Introduce Special Case | #9 |
| Complex Conditionals | Decompose Conditional | #10 |

---

## Refactoring Safety Checklist

Before refactoring:

- [ ] All tests passing (GREEN state)
- [ ] Version control committed
- [ ] Behavior change scope documented
- [ ] Backup created (automatic via MoAI-ADK)

During refactoring:

- [ ] Run tests after each small change
- [ ] Maintain backward compatibility
- [ ] Update documentation in parallel
- [ ] Log decisions in @TAG HISTORY

After refactoring:

- [ ] Full test suite passes
- [ ] Code coverage unchanged or increased
- [ ] Performance benchmarks comparable
- [ ] Code review completed
- [ ] Living docs updated via `/alfred:3-sync`

---

## References

- Martin Fowler, *Refactoring: Improving the Design of Existing Code* (2nd Edition, 2018)
- [Refactoring Catalog](https://refactoring.com/catalog/) (Official online catalog)
- Gang of Four, *Design Patterns: Elements of Reusable Object-Oriented Software*
- Robert C. Martin, *Clean Code*

---

**Version**: 2.0.0
**Last Updated**: 2025-10-22
**Part of**: MoAI-ADK Essentials Skills
**Related Skills**: `moai-essentials-review`, `moai-foundation-trust`
