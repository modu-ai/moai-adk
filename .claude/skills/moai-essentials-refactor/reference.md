# Refactoring Reference Guide

> **Version**: 2.0.0 (2025-10-22)
> **Based on**: Martin Fowler's Refactoring (2nd Edition), refactoring.com catalog

Complete reference for refactoring patterns, code smells, and best practices.

---

## Table of Contents

1. [Martin Fowler's Refactoring Catalog](#martin-fowlers-refactoring-catalog)
2. [Code Smells Detection Guide](#code-smells-detection-guide)
3. [SOLID Principles for Refactoring](#solid-principles-for-refactoring)
4. [Refactoring Safety Techniques](#refactoring-safety-techniques)
5. [Language-Specific Refactoring Tools](#language-specific-refactoring-tools)
6. [Metrics and Thresholds](#metrics-and-thresholds)
7. [Common Refactoring Workflows](#common-refactoring-workflows)

---

## Martin Fowler's Refactoring Catalog

### Encapsulation Refactorings

| Pattern | Purpose | When to use |
|---------|---------|-------------|
| **Encapsulate Variable** | Replace direct access with getter/setter | Exposing internal state |
| **Encapsulate Collection** | Return copy, use add/remove methods | Direct collection manipulation |
| **Encapsulate Record** | Convert data structure to class | Multiple related fields |
| **Replace Primitive with Object** | Wrap primitive in domain class | Primitive obsession |
| **Replace Temp with Query** | Convert variable to method call | Temporary variables everywhere |
| **Hide Delegate** | Encapsulate dependency relationships | Excessive coupling |
| **Remove Middle Man** | Expose delegate directly | Unnecessary indirection |

### Moving Features Between Objects

| Pattern | Purpose | When to use |
|---------|---------|-------------|
| **Move Function** | Relocate function to better class | Feature envy |
| **Move Field** | Relocate field to appropriate class | Data clumps |
| **Move Statements to Callers** | Push logic up call chain | Abstraction too general |
| **Move Statements to Function** | Pull logic into called function | Duplicate setup code |
| **Inline Function** | Remove unnecessary indirection | Overly simple methods |
| **Inline Class** | Merge class with another | Class does too little |
| **Extract Class** | Split responsibility into new class | Large class |
| **Extract Function** | Break down complex method | Long method |

### Organizing Data

| Pattern | Purpose | When to use |
|---------|---------|-------------|
| **Split Variable** | Use separate variable per assignment | Variable reused for different purposes |
| **Rename Variable/Field** | Improve clarity | Cryptic names |
| **Introduce Parameter Object** | Group parameters into object | Long parameter list |
| **Combine Functions into Class** | Group related functions | Functions sharing data |
| **Combine Functions into Transform** | Derive values via function | Calculated data |
| **Change Reference to Value** | Make immutable | Treat as value object |
| **Change Value to Reference** | Share single instance | Need identity |

### Simplifying Conditional Logic

| Pattern | Purpose | When to use |
|---------|---------|-------------|
| **Decompose Conditional** | Extract condition parts to methods | Complex boolean logic |
| **Consolidate Conditional Expression** | Combine into single check | Multiple checks, same action |
| **Replace Nested Conditional with Guard Clauses** | Early returns for edge cases | Deep nesting |
| **Replace Conditional with Polymorphism** | Use inheritance instead of switch | Type codes |
| **Introduce Special Case** | Use null object pattern | Repeated null checks |
| **Introduce Assertion** | Make assumption explicit | Hidden assumptions |

### Refactoring APIs

| Pattern | Purpose | When to use |
|---------|---------|-------------|
| **Separate Query from Modifier** | Split read/write methods | Side effects in queries |
| **Parameterize Function** | Use parameter instead of literals | Similar functions |
| **Remove Flag Argument** | Use explicit method names | Boolean parameters |
| **Preserve Whole Object** | Pass object, not fields | Extracting many fields |
| **Replace Parameter with Query** | Derive instead of passing | Redundant parameters |
| **Replace Query with Parameter** | Remove dependency | Function depends on state |
| **Remove Setting Method** | Make field immutable | Field set once |
| **Replace Constructor with Factory Function** | Flexible object creation | Complex construction |
| **Replace Function with Command** | Turn function into object | Need to undo/queue |

### Dealing with Inheritance

| Pattern | Purpose | When to use |
|---------|---------|-------------|
| **Pull Up Method** | Move to superclass | Duplicate in subclasses |
| **Pull Up Field** | Move field to superclass | Same field in subclasses |
| **Pull Up Constructor Body** | Move to superclass constructor | Duplicate initialization |
| **Push Down Method** | Move to subclass | Only used by some subclasses |
| **Push Down Field** | Move to subclass | Only needed by some subclasses |
| **Replace Type Code with Subclasses** | Use inheritance for types | Type-dependent behavior |
| **Remove Subclass** | Convert to configuration | Minimal difference |
| **Extract Superclass** | Create parent class | Common behavior |
| **Collapse Hierarchy** | Merge parent/child | Minimal difference |
| **Replace Subclass with Delegate** | Composition over inheritance | Multiple variations |
| **Replace Superclass with Delegate** | Remove inheritance | Misusing inheritance |

---

## Code Smells Detection Guide

### Bloaters

Code that has grown too large to work with effectively.

#### Long Method
- **Symptom**: Method > 50 lines
- **Solution**: Extract Function, Replace Temp with Query
- **Detection**: `radon cc` (Python), `sonar:CognitiveComplexity` (SonarQube)

#### Large Class
- **Symptom**: Class > 300 lines or many responsibilities
- **Solution**: Extract Class, Extract Subclass
- **Detection**: Line count, number of methods (> 20)

#### Primitive Obsession
- **Symptom**: Using primitives instead of domain objects
- **Solution**: Replace Primitive with Object, Introduce Parameter Object
- **Example**: Using `string` for email instead of `Email` class

#### Long Parameter List
- **Symptom**: Function with > 5 parameters
- **Solution**: Introduce Parameter Object, Preserve Whole Object
- **Detection**: Function signature analysis

#### Data Clumps
- **Symptom**: Same group of variables used together
- **Solution**: Extract Class, Introduce Parameter Object
- **Example**: `(x, y, z)` coordinates repeated everywhere

### Object-Orientation Abusers

Incomplete or incorrect application of OOP principles.

#### Switch Statements
- **Symptom**: Type code with case/switch
- **Solution**: Replace Conditional with Polymorphism
- **Note**: Simple switches are OK, problematic when duplicated

#### Temporary Field
- **Symptom**: Field used only in specific circumstances
- **Solution**: Extract Class, Introduce Special Case
- **Detection**: Fields initialized conditionally

#### Refused Bequest
- **Symptom**: Subclass doesn't use inherited methods
- **Solution**: Replace Inheritance with Delegation
- **Detection**: Unused parent methods

#### Alternative Classes with Different Interfaces
- **Symptom**: Classes do same thing, different names
- **Solution**: Rename Method, Move Method
- **Detection**: Code duplication across classes

### Change Preventers

Make changes difficult in multiple places.

#### Divergent Change
- **Symptom**: One class changed for many reasons
- **Solution**: Split Phase, Extract Class
- **Violation**: Single Responsibility Principle

#### Shotgun Surgery
- **Symptom**: One change requires many small changes everywhere
- **Solution**: Move Method, Move Field, Inline Class
- **Detection**: Change impact analysis

#### Parallel Inheritance Hierarchies
- **Symptom**: Creating subclass requires creating another
- **Solution**: Move Method, Move Field
- **Detection**: Parallel hierarchy patterns

### Dispensables

Unnecessary code that should be removed.

#### Comments
- **Symptom**: Comments explaining what code does
- **Solution**: Extract Function, Rename Method
- **Note**: Comments should explain *why*, not *what*

#### Duplicate Code
- **Symptom**: Same code in multiple places
- **Solution**: Extract Function, Pull Up Method
- **Detection**: `jscpd`, SonarQube duplication rules

#### Lazy Class
- **Symptom**: Class that doesn't do enough
- **Solution**: Inline Class, Collapse Hierarchy
- **Detection**: Classes with < 3 methods

#### Dead Code
- **Symptom**: Unused code
- **Solution**: Delete it
- **Detection**: Coverage reports, `vulture` (Python), `UCDetector` (Java)

#### Speculative Generality
- **Symptom**: "Might need it someday" abstractions
- **Solution**: Collapse Hierarchy, Inline Function
- **Principle**: YAGNI (You Aren't Gonna Need It)

### Couplers

Excessive coupling between classes.

#### Feature Envy
- **Symptom**: Method uses another class more than its own
- **Solution**: Move Method
- **Detection**: Method calls to other classes

#### Inappropriate Intimacy
- **Symptom**: Classes know too much about each other
- **Solution**: Move Method, Extract Class, Hide Delegate
- **Detection**: High coupling metrics

#### Message Chains
- **Symptom**: `a.b().c().d()`
- **Solution**: Hide Delegate
- **Violation**: Law of Demeter

#### Middle Man
- **Symptom**: Class delegates everything
- **Solution**: Remove Middle Man, Inline Function
- **Detection**: Classes with only delegation methods

---

## SOLID Principles for Refactoring

### Single Responsibility Principle (SRP)

**Definition**: A class should have only one reason to change.

**Code Smells**:
- Large Class
- Divergent Change
- Long Method

**Refactorings**:
- Extract Class
- Extract Function
- Split Phase

**Example Check**:
```python
# ❌ Violates SRP
class UserAccount:
    def __init__(self, user):
        self.user = user

    def save_to_database(self):
        # Database logic
        pass

    def send_welcome_email(self):
        # Email logic
        pass

    def calculate_loyalty_points(self):
        # Business logic
        pass

# ✅ Follows SRP
class UserAccount:
    def __init__(self, user):
        self.user = user

    def calculate_loyalty_points(self):
        # Only business logic
        pass

class UserRepository:
    def save(self, user):
        # Only database logic
        pass

class EmailService:
    def send_welcome(self, user):
        # Only email logic
        pass
```

### Open/Closed Principle (OCP)

**Definition**: Software entities should be open for extension, closed for modification.

**Code Smells**:
- Switch Statements on type codes
- Conditional logic that grows with new types

**Refactorings**:
- Replace Conditional with Polymorphism
- Extract Subclass
- Introduce Strategy Pattern

**Example Check**:
```typescript
// ❌ Violates OCP (must modify for new types)
function calculateDiscount(customer: Customer): number {
  if (customer.type === 'regular') {
    return 0;
  } else if (customer.type === 'premium') {
    return 0.1;
  } else if (customer.type === 'vip') {
    return 0.2;
  }
}

// ✅ Follows OCP (extend via new classes)
interface Customer {
  getDiscount(): number;
}

class RegularCustomer implements Customer {
  getDiscount(): number { return 0; }
}

class PremiumCustomer implements Customer {
  getDiscount(): number { return 0.1; }
}

class VipCustomer implements Customer {
  getDiscount(): number { return 0.2; }
}
```

### Liskov Substitution Principle (LSP)

**Definition**: Subtypes must be substitutable for their base types.

**Code Smells**:
- Refused Bequest
- Subtypes that throw on parent methods
- Subtypes that break parent contracts

**Refactorings**:
- Replace Inheritance with Delegation
- Remove Subclass
- Extract Interface

**Example Check**:
```java
// ❌ Violates LSP
class Rectangle {
    protected int width;
    protected int height;

    public void setWidth(int w) { width = w; }
    public void setHeight(int h) { height = h; }
    public int getArea() { return width * height; }
}

class Square extends Rectangle {
    @Override
    public void setWidth(int w) {
        width = w;
        height = w; // Breaks LSP!
    }
}

// ✅ Follows LSP
interface Shape {
    int getArea();
}

class Rectangle implements Shape {
    private int width;
    private int height;

    public Rectangle(int w, int h) {
        width = w;
        height = h;
    }

    public int getArea() { return width * height; }
}

class Square implements Shape {
    private int side;

    public Square(int s) { side = s; }

    public int getArea() { return side * side; }
}
```

### Interface Segregation Principle (ISP)

**Definition**: Clients should not be forced to depend on methods they don't use.

**Code Smells**:
- Fat interfaces
- Classes implementing interface with empty methods

**Refactorings**:
- Extract Interface
- Split Interface

**Example Check**:
```csharp
// ❌ Violates ISP
public interface IWorker
{
    void Work();
    void Eat();
    void Sleep();
}

public class Robot : IWorker
{
    public void Work() { /* ... */ }
    public void Eat() { throw new NotImplementedException(); } // Forced to implement!
    public void Sleep() { throw new NotImplementedException(); }
}

// ✅ Follows ISP
public interface IWorkable
{
    void Work();
}

public interface IFeedable
{
    void Eat();
}

public interface IRestable
{
    void Sleep();
}

public class Robot : IWorkable
{
    public void Work() { /* ... */ }
}

public class Human : IWorkable, IFeedable, IRestable
{
    public void Work() { /* ... */ }
    public void Eat() { /* ... */ }
    public void Sleep() { /* ... */ }
}
```

### Dependency Inversion Principle (DIP)

**Definition**: High-level modules should not depend on low-level modules. Both should depend on abstractions.

**Code Smells**:
- Direct instantiation of dependencies
- Tight coupling to concrete classes
- Hard to test due to dependencies

**Refactorings**:
- Extract Interface
- Introduce Parameter (Dependency Injection)
- Replace Constructor with Factory

**Example Check**:
```python
# ❌ Violates DIP
class EmailService:
    def send(self, message):
        # Send via SMTP
        pass

class UserNotifier:
    def __init__(self):
        self.email_service = EmailService()  # Tight coupling!

    def notify(self, user):
        self.email_service.send(f"Hello {user.name}")

# ✅ Follows DIP
from abc import ABC, abstractmethod

class MessageService(ABC):
    @abstractmethod
    def send(self, message):
        pass

class EmailService(MessageService):
    def send(self, message):
        # Send via SMTP
        pass

class SMSService(MessageService):
    def send(self, message):
        # Send via SMS
        pass

class UserNotifier:
    def __init__(self, message_service: MessageService):
        self.message_service = message_service  # Dependency injection

    def notify(self, user):
        self.message_service.send(f"Hello {user.name}")

# Usage
notifier = UserNotifier(EmailService())
```

---

## Refactoring Safety Techniques

### 1. Test-Driven Refactoring (TDR)

**Workflow**:
1. **GREEN**: Ensure all tests pass before refactoring
2. **REFACTOR**: Apply refactoring pattern
3. **GREEN**: Run tests after each small change
4. **COMMIT**: Commit after successful refactoring

**Key Rules**:
- Never refactor without GREEN tests
- Make small, incremental changes
- Run tests frequently (every 2-5 minutes)
- Revert if tests fail

### 2. Parallel Change (Expand-Contract Pattern)

**Use when**: Changing public APIs

**Steps**:
1. **Expand**: Add new interface alongside old
2. **Migrate**: Update callers to new interface
3. **Contract**: Remove old interface

**Example**:
```typescript
// Step 1: Expand (add new method)
class Calculator {
  // Old method (deprecated)
  calculate(x: number, y: number): number {
    return this.add(x, y);  // Delegate to new method
  }

  // New method
  add(x: number, y: number): number {
    return x + y;
  }
}

// Step 2: Migrate callers
// calc.calculate(5, 3) → calc.add(5, 3)

// Step 3: Contract (remove old method)
class Calculator {
  add(x: number, y: number): number {
    return x + y;
  }
}
```

### 3. Branch by Abstraction

**Use when**: Large-scale refactoring across many files

**Steps**:
1. Create abstraction layer
2. Migrate one module at a time
3. Remove abstraction when complete

### 4. Strangler Fig Pattern

**Use when**: Replacing legacy system

**Steps**:
1. Build new system alongside old
2. Gradually route traffic to new system
3. Remove old system when fully replaced

---

## Language-Specific Refactoring Tools

### Python
- **IDE**: PyCharm (best refactoring support), VS Code + Pylance
- **Complexity**: `radon cc` (cyclomatic complexity), `mccabe`
- **Duplication**: `jscpd`, `pylint` (R0801)
- **Dead Code**: `vulture`
- **All-in-one**: `ruff` (includes many refactoring hints)

### TypeScript/JavaScript
- **IDE**: VS Code, WebStorm
- **Linter**: ESLint + `eslint-plugin-sonarjs`
- **Complexity**: `eslint-plugin-complexity`
- **Duplication**: `jscpd`
- **All-in-one**: Biome (format, lint, complexity)

### Go
- **IDE**: GoLand, VS Code + Go extension
- **Refactoring**: `gofmt`, `goimports`, `go fix`
- **Linter**: `golangci-lint` (includes `gocyclo`, `dupl`)
- **Complexity**: `gocyclo`
- **Simplify**: `gosimple`

### Rust
- **IDE**: RustRover, VS Code + rust-analyzer
- **Linter**: `clippy` (excellent refactoring suggestions)
- **Format**: `rustfmt`
- **Complexity**: Built into `clippy`

### Java
- **IDE**: IntelliJ IDEA (best refactoring), Eclipse
- **Static Analysis**: SonarQube, SpotBugs, PMD
- **Duplication**: CPD (Copy-Paste Detector)
- **Complexity**: CheckStyle, PMD

### C#
- **IDE**: Visual Studio, Rider
- **Analyzer**: Roslyn Analyzers, FxCop
- **Duplication**: ReSharper, SonarQube

---

## Metrics and Thresholds

### Cyclomatic Complexity
- **Acceptable**: ≤ 10
- **Needs refactoring**: > 10
- **Critical**: > 20

### Lines of Code (LOC)
- **Function**: ≤ 50 lines
- **Class**: ≤ 300 lines
- **File**: ≤ 500 lines (Python/TypeScript), ≤ 1000 lines (Java/C#)

### Parameters
- **Maximum**: 5 parameters
- **Ideal**: ≤ 3 parameters

### Nesting Depth
- **Maximum**: 4 levels
- **Ideal**: ≤ 3 levels

### Test Coverage
- **Minimum**: 85%
- **Ideal**: > 90%
- **Critical paths**: 100%

### Duplication
- **Maximum**: 3% duplicated lines
- **Ideal**: < 1%

---

## Common Refactoring Workflows

### Workflow 1: Refactoring Legacy Method

```
1. Add characterization tests (cover current behavior)
2. Ensure tests pass (GREEN)
3. Extract Function for each responsibility
4. Rename variables for clarity
5. Run tests after each extraction
6. Replace conditional logic with polymorphism
7. Final test run
8. Update documentation
```

### Workflow 2: Breaking Down God Class

```
1. Identify responsibilities via method analysis
2. Group related methods
3. Extract Class for each group
4. Move Fields to appropriate classes
5. Update callers to use new classes
6. Run full test suite
7. Remove original class if hollow
```

### Workflow 3: Simplifying Deep Nesting

```
1. Identify guard clauses (edge cases)
2. Extract guard clauses to top
3. Use early returns
4. Extract remaining logic to methods
5. Flatten remaining conditionals
6. Run tests
```

---

## Quick Reference: Smell → Refactoring

| Code Smell | Primary Refactoring | Secondary Options |
|-------------|-------------------|------------------|
| Long Method | Extract Function | Replace Temp with Query |
| Large Class | Extract Class | Extract Subclass, Extract Interface |
| Long Parameter List | Introduce Parameter Object | Preserve Whole Object |
| Divergent Change | Split Phase | Extract Class |
| Shotgun Surgery | Move Method | Inline Class |
| Feature Envy | Move Method | Extract Function |
| Data Clumps | Extract Class | Introduce Parameter Object |
| Primitive Obsession | Replace Primitive with Object | Introduce Parameter Object |
| Switch Statements | Replace Conditional with Polymorphism | Replace Type Code with Subclasses |
| Parallel Inheritance | Move Method | Move Field |
| Lazy Class | Inline Class | Collapse Hierarchy |
| Speculative Generality | Collapse Hierarchy | Inline Function |
| Temporary Field | Extract Class | Introduce Special Case |
| Message Chains | Hide Delegate | Extract Function |
| Middle Man | Remove Middle Man | Inline Function |
| Inappropriate Intimacy | Move Method | Extract Class |
| Alternative Classes with Different Interfaces | Rename Method | Move Method |
| Incomplete Library Class | Introduce Foreign Method | Introduce Local Extension |
| Data Class | Move Method | Encapsulate Field |
| Refused Bequest | Replace Inheritance with Delegation | Push Down Method |
| Comments | Extract Function | Rename Method |
| Duplicate Code | Extract Function | Pull Up Method |

---

## References

- [Martin Fowler's Refactoring Catalog](https://refactoring.com/catalog/) (Official 2nd Edition catalog)
- Martin Fowler, *Refactoring: Improving the Design of Existing Code* (2nd Edition, 2018)
- Robert C. Martin, *Clean Code: A Handbook of Agile Software Craftsmanship* (2008)
- Michael Feathers, *Working Effectively with Legacy Code* (2004)
- [SourceMaking Refactoring Guide](https://sourcemaking.com/refactoring)
- [Refactoring.guru](https://refactoring.guru/) (Visual patterns and examples)

---

**Version**: 2.0.0
**Last Updated**: 2025-10-22
**Part of**: MoAI-ADK Essentials Skills
**Companion**: See examples.md for concrete scenarios
