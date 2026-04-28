---
name: moai-reference-anti-patterns
description: "8 anti-patterns with wrong/right code examples mapped to Agent Core Behaviors — loaded during run and review phases"
triggers:
  keywords:
    - "anti-pattern"
    - "over-engineering"
    - "refactor"
    - "simplify"
    - "scope"
    - "style drift"
  agents:
    - "expert-backend"
    - "expert-frontend"
    - "manager-quality"
    - "evaluator-active"
  phases:
    - "run"
    - "review"
user-invocable: false
---

# Anti-Patterns Reference — Karpathy Coding Principles

Concrete wrong/right code examples mapping 8 categories to the 6 Agent Core Behaviors.

## Category 1: Premature Abstraction

**Mapped to Behavior 4** (Enforce Simplicity)

### Wrong — Interface + Factory for Single Implementation

```go
// WRONG — Premature abstraction without multiple implementations
type EmailService interface {
    Send(to, subject, body string) error
}

type emailServiceFactory struct{}

func (f *emailServiceFactory) Create() EmailService {
    return &smtpEmailService{}
}

type smtpEmailService struct{}

func (s *smtpEmailService) Send(to, subject, body string) error {
    // SMTP implementation
    return nil
}
```

### Right — Direct Struct, Add Interface Only When Needed

```go
// RIGHT — Start with concrete struct, add interface when second implementation arrives
type EmailService struct {
    host     string
    port     int
    username string
    password string
}

func (s *EmailService) Send(to, subject, body string) error {
    // SMTP implementation
    return nil
}
```

**Principle**: Add abstractions only when you have at least two implementations that need to be swapped.

---

## Category 2: Over-Engineering / God Object

**Mapped to Behavior 4** (Enforce Simplicity)

### Wrong — Config Struct with 20 Fields, 15 Never Used

```go
// WRONG — God object config with unused fields
type AppConfig struct {
    DatabaseHost      string
    DatabasePort      int
    DatabaseUser      string
    DatabasePassword  string
    DatabaseName      string
    RedisHost         string
    RedisPort         int
    RedisPassword     string
    RedisMaxConns     int
    CacheTTL          int
    LogLevel          string
    LogFormat         string
    LogOutput         string
    MetricsEnabled    bool
    MetricsPort       int
    TraceEnabled      bool
    TraceSampleRate   float64
    FeatureFlags      map[string]bool
    SessionSecret     string
    OAuthProviders    map[string]string
    WebhookURLs       []string
}
```

### Right — Config Struct with Only Required Fields

```go
// RIGHT — Config with only the 5 fields actually used
type AppConfig struct {
    DatabaseURL string
    RedisURL    string
    LogLevel    string
    SessionKey  string
    FeatureFlag bool
}
```

**Principle**: Every field must earn its keep. Remove config options that are never read in production code.

---

## Category 3: Drive-By Refactoring

**Mapped to Behavior 5** (Maintain Scope Discipline)

### Wrong — Renaming Variables While Fixing Unrelated Bug

```go
// WRONG — Touching adjacent code outside the task scope
func fixUserAuth(userID int) error {
    // Bug fix: Check nil pointer
    if user == nil {
        return ErrUserNotFound
    }

    // UNRELATED: Renaming variables in the same function
    // old: userData := user.GetData()
    // new: userInfo := user.RetrieveInfo()
    userInfo := user.RetrieveInfo()

    return authenticate(userInfo)
}
```

### Right — Fix Only the Bug, Note Issues Separately

```go
// RIGHT — Fix only the reported bug, leave naming for a dedicated refactor task
func fixUserAuth(userID int) error {
    // Bug fix: Check nil pointer
    if user == nil {
        return ErrUserNotFound
    }

    userData := user.GetData()
    return authenticate(userData)
}

// TODO: Consider renaming user.GetData() to user.RetrieveInfo() for consistency
```

**Principle**: Touch only what you were asked to touch. Drive-by refactors create noise and risk regressions.

---

## Category 4: Style Drift

**Mapped to Behavior 5** (Maintain Scope Discipline)

### Wrong — Adding camelCase Handler in snake_case File

```go
// WRONG — Breaking existing file convention
package user_handler

import "github.com/gin-gonic/gin"

// File uses snake_case everywhere else
func get_user_profile(c *gin.Context) {
    userID := c.Param("id")
    // ...
}

// NEW: camelCase function breaks convention
func getUserSettings(c *gin.Context) {
    userID := c.Param("id")
    // ...
}
```

### Right — Match Existing snake_case Convention

```go
// RIGHT — Follow the established pattern in the file
package user_handler

import "github.com/gin-gonic/gin"

func get_user_profile(c *gin.Context) {
    userID := c.Param("id")
    // ...
}

// NEW: Matches snake_case convention
func get_user_settings(c *gin.Context) {
    userID := c.Param("id")
    // ...
}
```

**Principle**: Match the code style that already exists in the file. Consistency beats personal preference.

---

## Category 5: Silent Assumption

**Mapped to Behavior 1** (Surface Assumptions)

### Wrong — Assuming Input is Always Non-Nil

```go
// WRONG — No nil check, assumes caller always provides valid input
func ProcessOrder(order *Order) error {
    // Directly accesses order fields without validation
    total := calculateTotal(order.Items)
    return chargePayment(order.PaymentMethod, total)
}
```

### Right — Explicit Nil Check or Document Preconditions

```go
// RIGHT — Explicit nil check with clear error
func ProcessOrder(order *Order) error {
    if order == nil {
        return ErrInvalidOrder
    }

    total := calculateTotal(order.Items)
    return chargePayment(order.PaymentMethod, total)
}
```

**Python equivalent**:

```python
# Python equivalent
def process_order(order):
    if order is None:
        raise ValueError("Invalid order: None")

    total = calculate_total(order.items)
    return charge_payment(order.payment_method, total)
```

**Principle**: Never assume inputs are valid. Check explicitly or document preconditions in the function comment.

---

## Category 6: Guessing Over Clarifying

**Mapped to Behavior 2** (Manage Confusion Actively)

### Wrong — Implementing Ambiguous Requirement Without Asking

```go
// WRONG — Guessing at the meaning of "optimize the query"
func optimizeQuery(q *Query) *Query {
    // Assumption: "optimize" means add index hint
    q.Hint = "USE INDEX (primary)"
    return q
}
```

### Right — Stop and Surface the Ambiguity

```go
// RIGHT — Return error documenting the ambiguity
func optimizeQuery(q *Query) (*Query, error) {
    // Specification unclear: "optimize the query" could mean:
    // - Add index hint
    // - Rewrite JOIN order
    // - Add LIMIT clause
    // - Cache results
    return nil, ErrAmbiguousRequirement{
        Context: "optimizeQuery",
        Options: []string{"index_hint", "rewrite_join", "add_limit", "cache"},
    }
}
```

**Principle**: When requirements are ambiguous, STOP and surface the confusion. Do not guess.

---

## Category 7: Sycophantic Agreement

**Mapped to Behavior 3** (Push Back When Warranted)

### Wrong — Implementing Known-Bad Approach Without Pushback

```go
// WRONG — Proceeding with known security flaw because "user asked for it"
func StorePassword(password string) error {
    // User requested: "store passwords in plain text for debugging"
    // This is implemented without objection
    query := "INSERT INTO users (password) VALUES (?)"
    _, err := db.Exec(query, password)
    return err
}
```

### Right — State Concrete Downside, Propose Alternative, Accept Informed Override

```go
// RIGHT — Document the risk, propose alternative
func StorePassword(password string) error {
    // SECURITY WARNING: Plain text storage exposes passwords in:
    // - Database dumps
    // - Server logs
    // - Insider access
    //
    // Alternative: Use bcrypt hashed passwords (cost = 12)
    //
    // If you still want plain text for debugging, set env var:
    // INSECURE_MODE=true
    if os.Getenv("INSECURE_MODE") != "true" {
        hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
        password = string(hashed)
    }

    query := "INSERT INTO users (password) VALUES (?)"
    _, err := db.Exec(query, password)
    return err
}
```

**Principle**: Point out issues directly. Accept user override only when they proceed with full information.

---

## Category 8: Claiming Without Evidence

**Mapped to Behavior 6** (Verify, Don't Assume)

### Wrong — Marking Task Complete Without Verification

```go
// WRONG — No verification, assumes implementation works
func fixPaymentCalculation() {
    // Changed the calculation logic
    total = items * price * 1.1 // Added tax

    // Returns without testing
    log.Println("Payment calculation fixed")
}
```

### Right — Run Tests, Show Output, Verify Behavior

```go
// RIGHT — Verify with tests and show evidence
func fixPaymentCalculation() error {
    total = items * price * 1.1 // Added tax

    // Verification: Run unit tests
    if err := testPaymentCalculation(); err != nil {
        return fmt.Errorf("verification failed: %w", err)
    }

    log.Printf("Payment calculation fixed. Total: %.2f (items=%d, price=%.2f)",
        total, items, price)
    return nil
}

func testPaymentCalculation() error {
    // Test case 1: Basic calculation
    if got := calculateTotal(2, 100.0); got != 220.0 {
        return fmt.Errorf("expected 220.0, got %.2f", got)
    }

    // Test case 2: Zero items
    if got := calculateTotal(0, 100.0); got != 0.0 {
        return fmt.Errorf("expected 0.0, got %.2f", got)
    }

    return nil
}
```

**Python equivalent**:

```python
# Python equivalent
def fix_payment_calculation():
    total = items * price * 1.1  # Added tax

    # Verification: Run assertions
    assert calculate_total(2, 100.0) == 220.0, "Basic calculation failed"
    assert calculate_total(0, 100.0) == 0.0, "Zero items failed"

    print(f"Payment calculation fixed. Total: {total:.2f}")
```

**Principle**: Every task requires evidence of completion. "Seems right" is never sufficient.

---

## Mapping Summary

| Category | Core Behavior | Key Insight |
|----------|---------------|-------------|
| Premature Abstraction | Enforce Simplicity | Add interfaces when second implementation arrives |
| Over-Engineering | Enforce Simplicity | Every config field must earn its keep |
| Drive-By Refactoring | Maintain Scope Discipline | Touch only what was asked to touch |
| Style Drift | Maintain Scope Discipline | Match existing file conventions |
| Silent Assumption | Surface Assumptions | Check nil explicitly or document preconditions |
| Guessing Over Clarifying | Manage Confusion | Stop when requirements are ambiguous |
| Sycophantic Agreement | Push Back When Warranted | State downside, accept informed override |
| Claiming Without Evidence | Verify, Don't Assume | Run tests and show output |

---

**Version**: 1.0.0  
**Source**: SPEC-KARPATHY-001 M1  
**Last Updated**: 2026-04-28
