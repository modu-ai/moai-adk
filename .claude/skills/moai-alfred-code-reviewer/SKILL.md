---
name: moai-alfred-code-reviewer
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: Automated code review with language-specific best practices, SOLID principles, and actionable improvements.
keywords: ['code-review', 'solid', 'best-practices', 'refactor']
allowed-tools:
  - Read
  - Bash
---

# Alfred Code Reviewer v2.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-code-reviewer |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | `/alfred:3-sync` (doc-syncer), on-demand code review |
| **Tier** | Alfred |
| **Language Coverage** | All 23 MoAI-ADK languages |

---

## What It Does

Automated code review skill that enforces SOLID principles, language-specific best practices, and MoAI-ADK TRUST 5 principles across all supported languages.

**Key capabilities**:
- ✅ Multi-language code review (23 languages)
- ✅ SOLID principles validation
- ✅ TRUST 5 quality gates enforcement
- ✅ Automated static analysis integration
- ✅ Security vulnerability detection
- ✅ Code smell detection and remediation
- ✅ Test coverage analysis
- ✅ @TAG traceability verification
- ✅ Living documentation sync validation

---

## When to Use

**Automatic triggers**:
- `/alfred:3-sync` documentation sync phase
- Pre-PR quality gate validation
- Post-implementation review checkpoints
- Continuous integration code quality pipelines

**Manual invocation**:
- Code quality assessment requests
- Refactoring candidate identification
- Technical debt analysis
- Architecture review sessions
- Security audit preparation

---

## Core Review Principles

### 1. SOLID Principles

#### S - Single Responsibility Principle
**Definition**: A class should have one, and only one, reason to change.

**Detection patterns**:
- Classes with multiple unrelated methods
- Mixed concerns (business logic + persistence + presentation)
- "God objects" with excessive dependencies

**Example violation (Python)**:
```python
class UserManager:
    def create_user(self, data): pass
    def send_welcome_email(self, user): pass
    def generate_pdf_report(self, user): pass
    def log_analytics(self, event): pass
```

**Corrected**:
```python
class UserService:
    def create_user(self, data): pass

class EmailService:
    def send_welcome(self, user): pass

class ReportGenerator:
    def generate_pdf(self, user): pass

class AnalyticsLogger:
    def log(self, event): pass
```

#### O - Open/Closed Principle
**Definition**: Software entities should be open for extension, closed for modification.

**Detection patterns**:
- Large switch/if-else chains for type handling
- Direct modifications to existing classes for new features
- Hardcoded behavior instead of strategy patterns

**Example violation (TypeScript)**:
```typescript
class PaymentProcessor {
  process(type: string, amount: number) {
    if (type === 'credit') {
      // Credit card logic
    } else if (type === 'paypal') {
      // PayPal logic
    } else if (type === 'crypto') {
      // Crypto logic
    }
  }
}
```

**Corrected (Strategy Pattern)**:
```typescript
interface PaymentStrategy {
  process(amount: number): Promise<void>;
}

class CreditCardStrategy implements PaymentStrategy {
  async process(amount: number) { /* ... */ }
}

class PaymentProcessor {
  constructor(private strategy: PaymentStrategy) {}
  async process(amount: number) {
    return this.strategy.process(amount);
  }
}
```

#### L - Liskov Substitution Principle
**Definition**: Subtypes must be substitutable for their base types.

**Detection patterns**:
- Subclass throws errors for inherited methods
- Preconditions strengthened in subclasses
- Postconditions weakened in subclasses
- "Empty implementation" smell

**Example violation (Java)**:
```java
class Bird {
    void fly() { /* ... */ }
}

class Penguin extends Bird {
    @Override
    void fly() {
        throw new UnsupportedOperationException("Penguins can't fly");
    }
}
```

**Corrected (Interface Segregation)**:
```java
interface Bird {
    void eat();
}

interface FlyingBird extends Bird {
    void fly();
}

class Sparrow implements FlyingBird {
    public void eat() { /* ... */ }
    public void fly() { /* ... */ }
}

class Penguin implements Bird {
    public void eat() { /* ... */ }
}
```

#### I - Interface Segregation Principle
**Definition**: Clients should not depend on interfaces they don't use.

**Detection patterns**:
- Fat interfaces with many methods
- Implementations with empty/stub methods
- "NotImplementedException" in interface methods

**Example violation (C#)**:
```csharp
interface IRepository<T> {
    T Get(int id);
    List<T> GetAll();
    void Create(T entity);
    void Update(T entity);
    void Delete(int id);
    void BulkInsert(List<T> entities);
    void ExecuteRawSql(string sql);
}
```

**Corrected (Segregated Interfaces)**:
```csharp
interface IReadRepository<T> {
    T Get(int id);
    List<T> GetAll();
}

interface IWriteRepository<T> {
    void Create(T entity);
    void Update(T entity);
    void Delete(int id);
}

interface IBulkRepository<T> {
    void BulkInsert(List<T> entities);
}
```

#### D - Dependency Inversion Principle
**Definition**: Depend on abstractions, not concretions.

**Detection patterns**:
- Direct instantiation of dependencies via `new`
- Hardcoded database connections
- Tight coupling to specific implementations
- Missing constructor injection

**Example violation (Go)**:
```go
type UserService struct {
    db *sql.DB  // Concrete dependency
}

func NewUserService() *UserService {
    db, _ := sql.Open("postgres", "...") // Hardcoded
    return &UserService{db: db}
}
```

**Corrected (Dependency Injection)**:
```go
type UserRepository interface {
    Create(user User) error
    FindByID(id int) (User, error)
}

type UserService struct {
    repo UserRepository  // Abstract dependency
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}
```

---

### 2. Clean Code Principles

#### Meaningful Names
**Rules**:
- Use intention-revealing names
- Avoid disinformation (e.g., `accountList` when it's not a List)
- Make meaningful distinctions (`Product` vs `ProductInfo` vs `ProductData` is unclear)
- Use pronounceable names (`genymdhms` → `generationTimestamp`)
- Use searchable names (single-letter variables only in short scopes)

**Bad examples**:
```python
# ❌ Non-descriptive
def p(u):
    d = new Date()
    x = d.getTime() - u.c
    return x > 86400000

# ❌ Hungarian notation (outdated)
strUserName = "Alice"
intUserAge = 30
```

**Good examples**:
```python
# ✅ Intention-revealing
MILLISECONDS_PER_DAY = 86_400_000

def is_subscription_expired(user: User) -> bool:
    current_time = datetime.now()
    elapsed_time = current_time - user.created_at
    return elapsed_time.total_seconds() * 1000 > MILLISECONDS_PER_DAY
```

#### Functions
**Target metrics**:
- Length: ≤50 lines
- Parameters: ≤3 (use objects for more)
- Cyclomatic complexity: ≤10
- Nesting depth: ≤3 levels

**Function rules**:
1. **Do one thing** (single level of abstraction)
2. **Descriptive names** (verb phrases: `calculateTotalPrice`, `sendEmail`)
3. **No side effects** (pure functions where possible)
4. **Command-Query Separation** (either do something OR return something, not both)

**Bad example (multiple responsibilities)**:
```javascript
// ❌ Does too much
function processOrder(order) {
  // Validation
  if (!order.items.length) throw new Error("Empty order");

  // Calculation
  let total = 0;
  for (let item of order.items) {
    total += item.price * item.quantity;
  }

  // Persistence
  db.save(order);

  // Notification
  email.send(order.customer, `Order total: $${total}`);

  return total;
}
```

**Good example (separated concerns)**:
```javascript
// ✅ Single responsibility per function
function validateOrder(order) {
  if (!order.items.length) throw new Error("Empty order");
}

function calculateTotal(items) {
  return items.reduce((sum, item) => sum + item.price * item.quantity, 0);
}

function saveOrder(order) {
  return db.save(order);
}

function notifyCustomer(customer, total) {
  return email.send(customer, `Order total: $${total}`);
}

function processOrder(order) {
  validateOrder(order);
  const total = calculateTotal(order.items);
  saveOrder(order);
  notifyCustomer(order.customer, total);
  return total;
}
```

#### Comments
**When to use**:
- ✅ Explain WHY (intent, rationale, decisions)
- ✅ Legal/licensing information
- ✅ Clarify complex algorithms (with references)
- ✅ Warning of consequences
- ✅ TODO/FIXME with owner and ticket ID

**When NOT to use**:
- ❌ Explain WHAT (code should be self-documenting)
- ❌ Journal comments (use git history)
- ❌ Commented-out code (delete it)
- ❌ Redundant comments (repeat what code says)

**Bad comments**:
```python
# ❌ Redundant
# Increment i by 1
i = i + 1

# ❌ Journal
# 2023-01-15: Added user validation
# 2023-02-20: Fixed null pointer bug
# 2023-03-10: Refactored validation logic
```

**Good comments**:
```python
# ✅ Explain WHY
# We use SHA-256 instead of bcrypt here because this hash
# is for cache keys, not password storage. Performance > security.
cache_key = hashlib.sha256(data).hexdigest()

# ✅ Warning
# WARNING: Changing this constant will invalidate all existing
# user sessions and force re-authentication. Coordinate with ops.
SESSION_TIMEOUT_HOURS = 24

# ✅ Clarify complex algorithm
# Implements Dijkstra's shortest path algorithm
# See: https://en.wikipedia.org/wiki/Dijkstra%27s_algorithm
def find_shortest_path(graph, start, end):
    # ...
```

#### Error Handling
**Principles**:
- Use exceptions, not error codes
- Don't return `null` (use `Optional`/`Result`/`Maybe`)
- Don't pass `null`
- Fail fast, fail loud
- Context-rich error messages
- Wrap third-party exceptions

**Bad example (error codes)**:
```python
# ❌ Error codes
def get_user(user_id):
    if user_id < 0:
        return -1  # Error: invalid ID
    user = db.find(user_id)
    if user is None:
        return -2  # Error: not found
    return user

# Caller must check:
result = get_user(123)
if result == -1:
    print("Invalid ID")
elif result == -2:
    print("User not found")
else:
    print(f"User: {result.name}")
```

**Good example (exceptions)**:
```python
# ✅ Exceptions with context
class InvalidUserIdError(ValueError):
    def __init__(self, user_id):
        super().__init__(f"User ID must be positive, got: {user_id}")

class UserNotFoundError(LookupError):
    def __init__(self, user_id):
        super().__init__(f"User not found: {user_id}")

def get_user(user_id: int) -> User:
    if user_id < 0:
        raise InvalidUserIdError(user_id)

    user = db.find(user_id)
    if user is None:
        raise UserNotFoundError(user_id)

    return user

# Caller handles exceptions
try:
    user = get_user(123)
    print(f"User: {user.name}")
except InvalidUserIdError as e:
    logger.error(e)
except UserNotFoundError as e:
    logger.warning(e)
```

---

### 3. Code Smells Detection

#### Bloaters
**Long Method** (>50 lines)
- **Detection**: Count lines excluding comments/whitespace
- **Fix**: Extract sub-functions, use guard clauses

**Large Class** (>300 lines)
- **Detection**: Line count, method count (>20)
- **Fix**: Split by responsibility, extract helper classes

**Primitive Obsession**
- **Detection**: Many primitive parameters, no value objects
- **Example**: `createUser(string email, string name, int age, string address, string phone)`
- **Fix**: Create domain objects: `createUser(UserProfile profile)`

**Long Parameter List** (>3 params)
- **Detection**: Function signature inspection
- **Fix**: Parameter object, builder pattern

#### Object-Orientation Abusers
**Switch Statements** (for type checking)
- **Detection**: Large switch/if-else chains on type
- **Fix**: Polymorphism, strategy pattern

**Temporary Field**
- **Detection**: Fields used only in certain circumstances
- **Fix**: Extract class for those scenarios

**Refused Bequest**
- **Detection**: Subclass doesn't use inherited methods
- **Fix**: Replace inheritance with delegation

#### Change Preventers
**Divergent Change** (one class changed for multiple reasons)
- **Detection**: Git history shows changes for unrelated features
- **Fix**: Split class by change reason (SRP)

**Shotgun Surgery** (one change requires many class edits)
- **Detection**: Feature changes touch many files
- **Fix**: Move related behavior together

#### Dispensables
**Duplicate Code**
- **Detection**: Similar logic in multiple places
- **Fix**: Extract function/class, use inheritance/composition

**Dead Code**
- **Detection**: Uncalled functions, unreachable branches
- **Tools**: Coverage tools, static analysis
- **Fix**: Delete it

**Speculative Generality**
- **Detection**: Abstract interfaces with single implementation
- **Fix**: YAGNI - remove unnecessary abstraction

#### Couplers
**Feature Envy** (method uses more data from another class)
- **Detection**: Many calls to another object's getters
- **Fix**: Move method to the envied class

**Message Chains** (`a.getB().getC().getD()`)
- **Detection**: Chained getter calls
- **Fix**: Hide delegate, introduce intermediary

**Middle Man** (class delegates everything)
- **Detection**: Most methods just call another object
- **Fix**: Remove middle man, inline the delegation

---

## Language-Specific Review Checklists

### Python (Ruff + Mypy + Black)

**Linting**: Ruff (replaces Flake8, isort, pyupgrade)
```bash
ruff check .                    # Check all rules
ruff check --fix .              # Auto-fix
ruff format .                   # Format (replaces Black)
```

**Type checking**: Mypy
```bash
mypy src/ --strict              # Strict mode
mypy src/ --disallow-untyped-defs  # Require all type hints
```

**Python-specific checks**:
- [ ] Type hints on all public functions (PEP 484)
- [ ] Docstrings (Google/NumPy style)
- [ ] No mutable default arguments (`def func(items=[]):`)
- [ ] Use `pathlib.Path` not string paths
- [ ] Context managers for resources (`with open(...)`)
- [ ] List comprehensions over `map`/`filter` (more Pythonic)
- [ ] `enumerate()` over manual indexing
- [ ] `dataclasses` or `pydantic` for DTOs

**Common pitfalls**:
```python
# ❌ Mutable default argument
def add_item(item, items=[]):
    items.append(item)
    return items

# ✅ Use None and create new list
def add_item(item, items=None):
    if items is None:
        items = []
    items.append(item)
    return items

# ❌ String path manipulation
path = "/home/user/" + filename + ".txt"

# ✅ pathlib
path = Path("/home/user") / filename
path = path.with_suffix(".txt")
```

---

### TypeScript/JavaScript (Biome)

**Linting + Formatting**: Biome (replaces ESLint + Prettier)
```bash
biome check --write .           # Lint + format
biome check --unsafe .          # Apply unsafe fixes
biome ci .                      # CI mode (no fixes)
```

**Type checking**: Built-in TSC
```bash
tsc --noEmit                    # Type check only
tsc --strict                    # Strict mode
```

**TypeScript-specific checks**:
- [ ] Enable `strict` mode in `tsconfig.json`
- [ ] No `any` types (use `unknown` if needed)
- [ ] Prefer `interface` over `type` for objects
- [ ] Use `readonly` for immutable data
- [ ] Discriminated unions for type narrowing
- [ ] Avoid `!` non-null assertion (indicates design flaw)
- [ ] Use `as const` for literal types

**Common pitfalls**:
```typescript
// ❌ Implicit any
function process(data) {
    return data.map(x => x * 2);
}

// ✅ Explicit types
function process(data: number[]): number[] {
    return data.map(x => x * 2);
}

// ❌ Non-null assertion
const user = users.find(u => u.id === id)!;

// ✅ Explicit check
const user = users.find(u => u.id === id);
if (!user) throw new Error(`User not found: ${id}`);
```

---

### Go (golangci-lint)

**Linting**: golangci-lint (aggregates 50+ linters)
```bash
golangci-lint run                # Run all enabled linters
golangci-lint run --fix          # Auto-fix
```

**Formatting**: gofmt (built-in)
```bash
gofmt -w .                       # Format in place
goimports -w .                   # Format + organize imports
```

**Go-specific checks**:
- [ ] Error handling (never ignore errors)
- [ ] Defer cleanup (`defer file.Close()`)
- [ ] Context propagation in APIs
- [ ] Goroutine leak prevention (always have exit path)
- [ ] Use `errgroup` for coordinated goroutines
- [ ] Interface segregation (accept interfaces, return structs)
- [ ] Avoid global state (pass dependencies)

**Common pitfalls**:
```go
// ❌ Ignored error
data, _ := os.ReadFile("config.json")

// ✅ Handle error
data, err := os.ReadFile("config.json")
if err != nil {
    return fmt.Errorf("read config: %w", err)
}

// ❌ Goroutine leak
go func() {
    for {
        // Infinite loop with no exit
        work()
    }
}()

// ✅ Context-based cancellation
func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            work()
        }
    }
}
```

---

### Rust (Clippy + Rustfmt)

**Linting**: Clippy (official linter)
```bash
cargo clippy -- -D warnings      # Treat warnings as errors
cargo clippy --fix               # Auto-fix
```

**Formatting**: Rustfmt (official formatter)
```bash
cargo fmt                        # Format
cargo fmt --check                # CI mode
```

**Rust-specific checks**:
- [ ] No `unwrap()` in production code (use `?` or `expect()` with context)
- [ ] Minimize `clone()` (use references, `Cow`, or `Rc`/`Arc`)
- [ ] Lifetime annotations clear and minimal
- [ ] Use `thiserror` for error types
- [ ] Prefer iterators over manual loops
- [ ] Use `#[must_use]` for error types
- [ ] Document unsafe blocks with safety invariants

**Common pitfalls**:
```rust
// ❌ Panic on None
let value = map.get(&key).unwrap();

// ✅ Explicit handling
let value = map.get(&key)
    .ok_or_else(|| anyhow!("Key not found: {}", key))?;

// ❌ Unnecessary clones
fn process(data: Vec<String>) {
    let copy = data.clone();
    // ... only reads data
}

// ✅ Borrow instead
fn process(data: &[String]) {
    // ... read-only access
}
```

---

### Java (SpotBugs + PMD + Checkstyle)

**Static analysis**: SpotBugs, PMD
```bash
mvn spotbugs:check               # Bug detection
mvn pmd:check                    # Code quality
mvn checkstyle:check             # Style enforcement
```

**Formatting**: Google Java Format
```bash
java -jar google-java-format.jar --replace src/**/*.java
```

**Java-specific checks**:
- [ ] Use `Optional` instead of null returns
- [ ] Close resources (try-with-resources)
- [ ] Immutable objects where possible (`final` fields)
- [ ] Avoid checked exceptions for control flow
- [ ] Use `StringBuilder` in loops (not `+`)
- [ ] Override `equals()` and `hashCode()` together
- [ ] Use `@Override` annotation

**Common pitfalls**:
```java
// ❌ Null return
public User findUser(int id) {
    return users.get(id); // May return null
}

// ✅ Optional
public Optional<User> findUser(int id) {
    return Optional.ofNullable(users.get(id));
}

// ❌ Resource leak
FileReader reader = new FileReader("file.txt");
// ... exception before close()

// ✅ Try-with-resources
try (FileReader reader = new FileReader("file.txt")) {
    // ... automatic close
}
```

---

## TRUST 5 Principles Review

### T - Test First (≥85% coverage)

**Coverage tools by language**:
```bash
# Python
pytest --cov=src --cov-report=html --cov-fail-under=85

# TypeScript/JavaScript
vitest --coverage --coverage.threshold.lines=85

# Go
go test -cover ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Rust
cargo tarpaulin --out Html --output-dir coverage

# Java
mvn jacoco:report jacoco:check
```

**Review checklist**:
- [ ] Overall coverage ≥85%
- [ ] No critical paths uncovered (auth, payment, data loss scenarios)
- [ ] Edge cases tested (null, empty, boundary values)
- [ ] Error paths tested (what happens when things fail)
- [ ] Integration tests for system boundaries
- [ ] Tests are deterministic (no flaky tests)

---

### R - Readable (Linting & Formatting)

**Enforce with CI**:
```yaml
# .github/workflows/lint.yml
name: Lint
on: [pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Python
        run: ruff check . && mypy src/

      - name: TypeScript
        run: biome ci .

      - name: Go
        run: golangci-lint run

      - name: Rust
        run: cargo clippy -- -D warnings
```

**Review checklist**:
- [ ] No linting errors
- [ ] Consistent formatting (automated)
- [ ] Naming conventions followed
- [ ] Comments explain WHY, not WHAT
- [ ] No "TODO" without owner/ticket

---

### U - Unified (Type Safety)

**Type checking tools**:
```bash
# Python: mypy
mypy src/ --strict

# TypeScript: tsc
tsc --noEmit --strict

# Go: built-in
go build ./...

# Rust: built-in
cargo check

# Java: built-in
mvn compile
```

**Review checklist**:
- [ ] Type checker passes with no errors
- [ ] No `any` / `unknown` without justification
- [ ] Generic types properly constrained
- [ ] Return types explicit (not inferred)
- [ ] Schema validation for external data

---

### S - Secured (Security Analysis)

**Security tools** (language-agnostic):
```bash
# SAST: Semgrep
semgrep scan --config=auto

# Dependency scanning
npm audit                        # Node.js
pip-audit                        # Python
cargo audit                      # Rust
mvn dependency-check:check       # Java

# Secret scanning
gitleaks detect --source .
trufflehog filesystem .
```

**OWASP Top 10 checklist**:
- [ ] A01: Broken Access Control
  - [ ] Authorization checks on all protected endpoints
  - [ ] No direct object references without validation
- [ ] A02: Cryptographic Failures
  - [ ] No hardcoded secrets (use env vars)
  - [ ] TLS for all external communication
  - [ ] Use bcrypt/argon2 for passwords, never MD5/SHA1
- [ ] A03: Injection
  - [ ] Parameterized queries (no string concatenation)
  - [ ] Input validation on all user data
  - [ ] Output encoding for HTML/XML/JSON
- [ ] A04: Insecure Design
  - [ ] Threat modeling for sensitive features
  - [ ] Rate limiting on APIs
  - [ ] Secure defaults
- [ ] A05: Security Misconfiguration
  - [ ] No default credentials
  - [ ] Minimal dependencies (reduce attack surface)
  - [ ] Security headers (CSP, HSTS, etc.)

---

### T - Trackable (@TAG Coverage)

**TAG verification**:
```bash
# Scan all TAGs
rg '@(SPEC|CODE|TEST|DOC):' -n .moai/specs/ src/ tests/ docs/

# Check for orphans (CODE without SPEC)
rg '@CODE:AUTH-001' -l src/      # CODE exists
rg '@SPEC:AUTH-001' -l .moai/specs/  # SPEC missing → orphan

# Verify TAG links
rg '@CODE:AUTH-001.*SPEC:' src/  # Should reference SPEC file
```

**Review checklist**:
- [ ] Every implementation has @CODE tag
- [ ] Every test has @TEST tag linked to @CODE
- [ ] Every @CODE references its @SPEC
- [ ] No duplicate TAG IDs
- [ ] No orphan TAGs (CODE without SPEC)
- [ ] HISTORY section updated in SPECs

---

## Code Review Workflow

### Phase 1: Automated Pre-Review
**Before human review, run automated checks**:

```bash
# 1. Tests
pytest --cov=src --cov-fail-under=85

# 2. Linting
ruff check . && mypy src/

# 3. Security
semgrep scan --config=auto
pip-audit

# 4. TAG verification
rg '@(SPEC|CODE|TEST):' -n src/ tests/

# 5. Complexity check
radon cc src/ -a -nb  # Cyclomatic complexity
```

**Automation gate**: If any automated check fails, block PR until fixed.

---

### Phase 2: Human Review Checklist

**Architecture**:
- [ ] Design aligns with SPEC requirements
- [ ] SOLID principles followed
- [ ] No unnecessary abstraction (YAGNI)
- [ ] Appropriate design patterns used

**Code Quality**:
- [ ] Functions ≤50 lines
- [ ] Cyclomatic complexity ≤10
- [ ] No code duplication
- [ ] Clear, descriptive naming
- [ ] Error handling comprehensive

**Testing**:
- [ ] Happy path covered
- [ ] Edge cases covered
- [ ] Error scenarios covered
- [ ] Tests are isolated (no shared state)
- [ ] No flaky tests

**Documentation**:
- [ ] Public APIs documented
- [ ] Complex algorithms explained
- [ ] SPEC updated if behavior changed
- [ ] Living docs synced

**Security**:
- [ ] Input validation present
- [ ] No hardcoded secrets
- [ ] Authorization checks present
- [ ] Sensitive data encrypted/hashed

---

### Phase 3: Review Feedback Template

**Use this format for actionable feedback**:

```markdown
## Summary
[Overall assessment: Approve / Request Changes / Comment]

## Critical Issues (Must fix before merge)
- [ ] **File**: `src/auth.py:42`
  - **Issue**: SQL injection vulnerability
  - **Fix**: Use parameterized query: `cursor.execute("SELECT * FROM users WHERE id = ?", (user_id,))`

## Major Issues (Should fix)
- [ ] **File**: `src/user_service.py:15-45`
  - **Issue**: Function violates SRP (validation + persistence + email)
  - **Fix**: Extract `EmailService` and `UserValidator`

## Minor Issues (Consider fixing)
- [ ] **File**: `src/utils.py:23`
  - **Issue**: Non-descriptive variable name `x`
  - **Fix**: Rename to `elapsed_time`

## Positive Feedback
- ✅ Excellent test coverage (92%)
- ✅ Clear documentation
- ✅ Good use of type hints

## Estimated Effort
- Critical: 1 hour
- Major: 2 hours
- Minor: 30 minutes
- **Total**: ~3.5 hours
```

---

### Phase 4: Post-Review Validation

**After changes applied, verify**:
```bash
# 1. All review comments addressed
gh pr review-comments list --state=unresolved

# 2. Re-run automated checks
./scripts/quality-gate.sh

# 3. Verify TAG updates
rg '@CODE:.*HISTORY' src/  # HISTORY updated?

# 4. Living docs synced
/alfred:3-sync
```

---

## Integration with Alfred Workflow

### Auto-trigger from doc-syncer

When `/alfred:3-sync` runs, `doc-syncer` sub-agent invokes this skill:

```python
# Pseudo-code for doc-syncer integration
def sync_documentation():
    # Step 1: Run code review
    review_results = run_code_review()

    # Step 2: Check TRUST 5
    trust_validation = validate_trust_principles()

    # Step 3: TAG verification
    tag_integrity = verify_tag_chains()

    # Step 4: Block if critical issues
    if review_results.has_critical_issues():
        raise QualityGateError("Code review failed - see logs")

    # Step 5: Update Living Docs
    update_living_documentation()

    # Step 6: Mark PR as Ready (if Draft)
    if all_gates_passed():
        gh_pr_ready()
```

---

### Manual Review Command

```bash
# Invoke Alfred code review
@agent-cc-manager invoke moai-alfred-code-reviewer

# Review specific files
@agent-cc-manager invoke moai-alfred-code-reviewer src/auth.py src/user.py

# Review with focus
@agent-cc-manager invoke moai-alfred-code-reviewer --focus=security
@agent-cc-manager invoke moai-alfred-code-reviewer --focus=solid
@agent-cc-manager invoke moai-alfred-code-reviewer --focus=trust
```

---

## Advanced Topics

### Complexity Metrics

**Cyclomatic Complexity** (McCabe):
- Measures number of linearly independent paths
- Formula: `M = E - N + 2P` (E=edges, N=nodes, P=connected components)
- Target: ≤10 per function

**Example**:
```python
def calculate_discount(user, total):  # Complexity = 4
    discount = 0
    if user.is_premium:              # +1
        discount += 0.1
    if total > 100:                  # +1
        discount += 0.05
    if user.referral_code:           # +1
        discount += 0.02
    return total * (1 - discount)
```

**Tools**:
```bash
# Python
radon cc src/ -a -nb

# JavaScript
npx complexity-report src/

# Go
gocyclo -over 10 .

# Java
mvn pmd:check  # Includes cyclomatic complexity
```

---

### Code Duplication Detection

**Tools**:
```bash
# PMD Copy/Paste Detector (multi-language)
pmd cpd --minimum-tokens 50 --files src/

# Language-specific
jscpd src/              # JavaScript/TypeScript
duplo src/**/*.py       # Python
```

**Acceptable duplication**:
- Test setup code (but consider fixtures/helpers)
- Boilerplate (consider code generation)
- Different domains with similar structure (NOT duplication, just similarity)

---

### Static Analysis Tool Matrix

| Language | Linter | Formatter | Type Checker | Security | Complexity |
|----------|--------|-----------|--------------|----------|------------|
| **Python** | ruff | ruff format | mypy | semgrep, bandit | radon |
| **TypeScript** | Biome | Biome | tsc | semgrep | complexity-report |
| **JavaScript** | Biome | Biome | - | semgrep | complexity-report |
| **Go** | golangci-lint | gofmt | built-in | gosec | gocyclo |
| **Rust** | clippy | rustfmt | built-in | cargo-audit | - |
| **Java** | SpotBugs, PMD | google-java-format | built-in | SpotBugs | PMD |
| **Kotlin** | ktlint | ktlint | built-in | detekt | detekt |
| **C/C++** | clang-tidy | clang-format | - | cppcheck | lizard |
| **C#** | Roslyn | dotnet format | built-in | Security Code Scan | - |
| **Swift** | SwiftLint | swift-format | built-in | - | - |
| **Ruby** | RuboCop | RuboCop | Sorbet | brakeman | flog |
| **PHP** | PHPCS | PHP-CS-Fixer | PHPStan | Psalm | phpmetrics |

---

## Common Anti-Patterns by Language

### Python
- **Mutable default arguments**: `def func(items=[])`
- **Wildcard imports**: `from module import *`
- **Bare except**: `except:` (catch all)
- **Global variables**: Use dependency injection
- **String manipulation of paths**: Use `pathlib`

### JavaScript/TypeScript
- **Callback hell**: Use async/await
- **Global variables**: Use modules, IIFE
- **== instead of ===**: Use strict equality
- **Ignoring Promise rejections**: Always `.catch()` or `try/catch`
- **Mutating parameters**: Clone or use immutable patterns

### Go
- **Ignoring errors**: `data, _ := func()`
- **Goroutine leaks**: No cancellation mechanism
- **Global state**: Pass dependencies explicitly
- **Naked returns**: Makes code unclear
- **Pointer to loop variable**: `&item` in loop (creates shared reference)

### Rust
- **Excessive `unwrap()`**: Use `?` operator or `expect()` with context
- **Unnecessary `clone()`**: Use references, `Cow`, or smart pointers
- **String vs &str confusion**: Prefer `&str` for parameters
- **Manual memory management**: Use RAII pattern
- **`unsafe` without invariants**: Document safety requirements

### Java
- **Null returns**: Use `Optional`
- **Checked exceptions for control flow**: Use unchecked for unrecoverable
- **String concatenation in loops**: Use `StringBuilder`
- **Ignoring equals/hashCode contract**: Override both or neither
- **Finalizers**: Use try-with-resources, avoid `finalize()`

---

## Performance Review Guidelines

**When to optimize**:
- ✅ Measured bottleneck (profiler data)
- ✅ Failing SLA/performance requirements
- ✅ Resource exhaustion (OOM, CPU spikes)

**When NOT to optimize**:
- ❌ Premature optimization (no metrics)
- ❌ Micro-optimizations (negligible impact)
- ❌ Sacrificing readability for <10% gain

**Profiling tools**:
```bash
# Python
python -m cProfile -o output.prof script.py
snakeviz output.prof

# JavaScript/Node
node --prof script.js
node --prof-process isolate-*.log

# Go
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Rust
cargo flamegraph

# Java
java -agentlib:hprof=cpu=samples MyApp
jvisualvm
```

---

## Changelog

- **v2.0.0** (2025-10-22): Comprehensive expansion to 1,200+ lines with multi-language support, SOLID principles, Clean Code guidelines, TRUST 5 integration, and automated review workflows
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (TRUST 5 principles validation)
- `moai-alfred-refactoring-coach` (refactoring guidance for identified issues)
- `moai-alfred-tag-scanning` (TAG integrity verification)
- `moai-essentials-review` (lightweight review for quick checks)
- `moai-alfred-trust-validation` (quality gate enforcement)

---

## Best Practices

✅ **DO**:
- Run automated checks before human review
- Provide actionable feedback with file/line references
- Focus on high-impact issues (SOLID, security, correctness)
- Approve code that meets quality gates even if minor issues remain
- Use tools to enforce consistency (linters, formatters)
- Document decisions in code comments
- Verify TAG traceability

❌ **DON'T**:
- Block PRs for style preferences (let tools decide)
- Request changes without providing examples
- Ignore automated tool warnings without justification
- Mix refactoring with feature changes
- Skip security review for "small" changes
- Allow code with <85% test coverage
- Merge code with unresolved critical issues

---

## References

### Books
- [Clean Code - Robert C. Martin](https://www.oreilly.com/library/view/clean-code-a/9780136083238/)
- [Refactoring - Martin Fowler](https://martinfowler.com/books/refactoring.html)
- [Design Patterns - Gang of Four](https://en.wikipedia.org/wiki/Design_Patterns)
- [The Pragmatic Programmer - Hunt & Thomas](https://pragprog.com/titles/tpp20/)

### Online Resources
- [Google Engineering Practices - Code Review](https://google.github.io/eng-practices/review/)
- [SOLID Principles - Uncle Bob](https://blog.cleancoder.com/uncle-bob/2020/10/18/Solid-Relevance.html)
- [Refactoring Catalog - Martin Fowler](https://refactoring.com/catalog/)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Static Analysis Tools 2025](https://www.aikido.dev/blog/static-code-analysis-tools)

### Tool Documentation (2025)
- [Ruff (Python)](https://docs.astral.sh/ruff/)
- [Biome (JS/TS)](https://biomejs.dev/)
- [golangci-lint (Go)](https://golangci-lint.run/)
- [Clippy (Rust)](https://doc.rust-lang.org/clippy/)
- [Semgrep (Security)](https://semgrep.dev/docs/)

---

_For practical examples, see `examples.md`. For tool references, see `reference.md`._
