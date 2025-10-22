# Refactoring Reference Documentation

_Last updated: 2025-10-22_

## Martin Fowler's Refactoring Catalog

### Common Refactorings

| Refactoring | Problem | Solution |
|-------------|---------|----------|
| Extract Method | Long method | Break into smaller methods |
| Inline Method | Unnecessary indirection | Merge method into caller |
| Extract Variable | Complex expression | Assign to named variable |
| Inline Variable | Trivial variable | Use expression directly |
| Rename Variable | Unclear name | Use descriptive name |
| Extract Class | Class doing too much | Split into multiple classes |
| Inline Class | Class doing too little | Merge into another class |
| Move Method | Method in wrong class | Move to appropriate class |

---

## Code Smells (Martin Fowler)

### Bloaters
Issues making code large and unwieldy.

**Long Method**
- **Problem**: Method too long (>50 lines)
- **Refactoring**: Extract Method, Replace Temp with Query

**Large Class**
- **Problem**: Class has too many responsibilities
- **Refactoring**: Extract Class, Extract Subclass

**Long Parameter List**
- **Problem**: Too many parameters (>3)
- **Refactoring**: Introduce Parameter Object, Preserve Whole Object

**Data Clumps**
- **Problem**: Same group of data appearing together
- **Refactoring**: Extract Class, Introduce Parameter Object

### Object-Orientation Abusers
Violations of OOP principles.

**Switch Statements**
- **Problem**: Type checking with switch/if-else chains
- **Refactoring**: Replace Conditional with Polymorphism

**Temporary Field**
- **Problem**: Field only set in certain circumstances
- **Refactoring**: Extract Class, Introduce Null Object

**Refused Bequest**
- **Problem**: Subclass doesn't use inherited methods
- **Refactoring**: Replace Inheritance with Delegation

### Change Preventers
Changes require many modifications.

**Divergent Change**
- **Problem**: One class changed for multiple reasons
- **Refactoring**: Extract Class, Move Method

**Shotgun Surgery**
- **Problem**: One change requires many small edits
- **Refactoring**: Move Method, Inline Class

**Parallel Inheritance Hierarchies**
- **Problem**: Creating subclass requires creating another
- **Refactoring**: Move Method, Move Field

### Dispensables
Unnecessary code.

**Comments** (explaining what code does)
- **Problem**: Comments compensating for bad code
- **Refactoring**: Extract Method, Rename Method

**Duplicate Code**
- **Problem**: Same code in multiple places
- **Refactoring**: Extract Method, Pull Up Method

**Dead Code**
- **Problem**: Unused code
- **Refactoring**: Delete it

---

## SOLID Principles Review

### Single Responsibility Principle (SRP)
**Guideline**: A class should have only one reason to change.

**Bad**:
```python
class User:
    def save_to_database(self):  # Database concern
        pass
    def send_email(self):        # Email concern
        pass
```

**Good**:
```python
class User:
    pass

class UserRepository:
    def save(self, user): pass

class EmailService:
    def send(self, user): pass
```

### Open/Closed Principle (OCP)
**Guideline**: Open for extension, closed for modification.

**Use**: Strategy pattern, polymorphism instead of conditionals.

### Liskov Substitution Principle (LSP)
**Guideline**: Subtypes must be substitutable for base types.

**Violation**: Subclass throws exceptions parent doesn't.

### Interface Segregation Principle (ISP)
**Guideline**: Many client-specific interfaces > one general interface.

### Dependency Inversion Principle (DIP)
**Guideline**: Depend on abstractions, not concretions.

---

## Refactoring Workflow

### 1. Identify Code Smell
Run static analysis, code review, complexity metrics.

### 2. Choose Refactoring
Match smell to appropriate refactoring technique.

### 3. Write Tests
Ensure existing functionality is covered (>85% coverage).

### 4. Apply Refactoring
Make small, incremental changes.

### 5. Run Tests
Verify behavior unchanged after each step.

### 6. Commit
Create atomic commits for each refactoring step.

---

## Design Patterns for Refactoring

| Pattern | Use When | Refactoring |
|---------|----------|-------------|
| Strategy | Conditional logic | Replace Conditional with Polymorphism |
| Factory | Object creation scattered | Extract Factory Method |
| Template Method | Similar algorithms | Form Template Method |
| Decorator | Extending behavior | Replace Inheritance with Delegation |
| Observer | Notification logic | Introduce Observer |

---

## Tools

### Static Analysis
- **Python**: ruff, pylint, radon (complexity)
- **JavaScript**: ESLint, SonarJS
- **Go**: golangci-lint
- **Rust**: clippy

### Complexity Metrics
```bash
# Python (radon)
radon cc src/ -a  # Cyclomatic complexity

# JavaScript (complexity-report)
complexity-report src/

# Go (gocyclo)
gocyclo -avg .
```

---

## References

- [Refactoring: Improving the Design of Existing Code - Martin Fowler](https://refactoring.com/)
- [Refactoring Catalog](https://refactoring.com/catalog/)
- [Code Smells](https://refactoring.guru/refactoring/smells)
- [Design Patterns](https://refactoring.guru/design-patterns)

---

_For refactoring examples, see examples.md_
