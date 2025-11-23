# Refactoring Modules - Navigation Index

**Parent Skill**: moai-essentials-refactor
**Version**: 1.2.0
**Last Updated**: 2025-11-24
**Module Count**: 5

---

## Module Directory Structure

```
modules/
├── README.md (this file)
├── advanced-patterns.md - Architecture evolution and design pattern introduction
├── optimization.md - Performance optimization through refactoring
├── refactoring-patterns.md - 10+ advanced refactoring patterns with code examples
├── rope-integration.md - Complete Rope API reference and integration guide
└── technical-debt.md - Technical debt quantification and prioritization framework
```

---

## Module Descriptions

### Core Refactoring Modules

#### **refactoring-patterns.md** (Essential - Start Here)
**Purpose**: Comprehensive catalog of 10+ advanced refactoring patterns with production-ready code examples.

**Coverage**:
- **Method Extraction**: Break down long methods into smaller, focused functions
- **Class Extraction**: Split large classes into cohesive units
- **Replace Conditional with Polymorphism**: Eliminate complex if/else chains
- **Introduce Parameter Object**: Group related parameters
- **Extract Interface**: Define contracts for implementations
- **Move Method/Field**: Improve class responsibilities
- **Inline Method/Variable**: Remove unnecessary indirection
- **Replace Magic Numbers with Constants**: Improve code readability
- **Decompose Conditional**: Simplify complex boolean logic
- **Consolidate Duplicate Code**: DRY principle application

**When to Use**: Code smells detected, cyclomatic complexity >10, method length >50 lines

**Prerequisites**: None (beginner-friendly with examples)

**Example Use Cases**:
- Refactor 100-line method into 5 focused methods
- Replace 7-branch conditional with polymorphism
- Extract data class from anemic domain model
- Eliminate duplicate code across 5 modules

**Learning Path**: Start here → rope-integration.md → technical-debt.md

**Estimated Reading Time**: 2-3 hours

**Code Examples**: Python, JavaScript, TypeScript, Java, C#

---

#### **rope-integration.md** (Tool Guide)
**Purpose**: Complete Rope library integration guide with API reference and practical examples.

**Coverage**:
- Rope installation and project setup
- Extract Method refactoring API
- Rename refactoring API
- Move refactoring API
- Inline refactoring API
- Extract Variable/Field API
- Change Signature API
- Programmatic refactoring workflows
- CI/CD integration patterns
- Error handling and rollback strategies

**When to Use**: Automated refactoring, large-scale code transformations, CI/CD integration

**Prerequisites**: refactoring-patterns.md (understanding patterns)

**Example Use Cases**:
- Automate method extraction across codebase
- Rename 50+ occurrences of variable safely
- Move methods between classes programmatically
- Integrate refactoring validation in CI pipeline

**Learning Path**: refactoring-patterns.md → rope-integration.md → optimization.md

**Estimated Reading Time**: 1.5-2 hours

**Tool Version**: Rope 1.13+

---

#### **technical-debt.md** (Strategic Planning)
**Purpose**: AI-driven technical debt quantification and prioritization framework.

**Coverage**:
- Technical debt measurement metrics
  - Code complexity score
  - Duplication percentage
  - Test coverage gaps
  - Documentation completeness
- AI-powered debt analysis
  - Pattern recognition
  - Impact assessment
  - Effort estimation
- Prioritization frameworks
  - Impact vs. Effort matrix
  - Risk-based prioritization
  - Business value alignment
- Debt tracking and reporting
  - Debt trend visualization
  - Team velocity impact
  - Refactoring ROI calculation

**When to Use**: Project planning, sprint planning, technical roadmap creation

**Prerequisites**: refactoring-patterns.md (understanding code smells)

**Example Use Cases**:
- Quantify codebase debt score (0-100)
- Prioritize 15 refactoring opportunities
- Estimate 12.5 days refactoring effort
- Track debt reduction over time

**Learning Path**: refactoring-patterns.md → technical-debt.md → advanced-patterns.md

**Estimated Reading Time**: 1.5-2 hours

**Frameworks**: AI analysis, Context7 patterns, Refactoring.Guru integration

---

### Advanced Modules

#### **advanced-patterns.md** (Architecture)
**Purpose**: Architecture evolution and design pattern introduction through refactoring.

**Coverage**:
- **Microservices Extraction**: Monolith → microservices refactoring
- **Design Pattern Introduction**:
  - Strategy Pattern (replace conditionals)
  - Factory Pattern (object creation)
  - Observer Pattern (event handling)
  - Decorator Pattern (behavior extension)
  - Template Method Pattern (algorithm skeleton)
- **Domain-Driven Design (DDD) Refactoring**:
  - Bounded contexts
  - Aggregate roots
  - Value objects
  - Domain events
- **Clean Architecture Transformation**:
  - Dependency inversion
  - Use case extraction
  - Interface adapters
- **Legacy Code Modernization**:
  - Characterization tests
  - Seam identification
  - Strangler Fig pattern

**When to Use**: System architecture evolution, legacy modernization, pattern adoption

**Prerequisites**: refactoring-patterns.md, rope-integration.md, architecture experience

**Example Use Cases**:
- Extract microservice from monolith
- Introduce Strategy Pattern to replace 5-branch conditional
- Refactor to Clean Architecture
- Modernize 10-year-old legacy codebase

**Learning Path**: technical-debt.md → advanced-patterns.md → optimization.md

**Estimated Reading Time**: 2-3 hours

**Complexity**: High (expert-level)

---

#### **optimization.md** (Performance)
**Purpose**: Performance optimization through strategic refactoring.

**Coverage**:
- **Algorithmic Refactoring**:
  - Complexity reduction (O(n²) → O(n log n))
  - Data structure replacement (list → set, dict → lru_cache)
  - Lazy evaluation introduction
- **Memory Optimization Refactoring**:
  - Generator pattern introduction
  - Object pooling
  - Memory-mapped file usage
  - Copy-on-write optimization
- **Database Refactoring**:
  - N+1 query elimination
  - Eager loading introduction
  - Index optimization
  - Query batching
- **Caching Refactoring**:
  - Memoization introduction
  - Cache layer extraction
  - CDN integration
- **Parallel Processing Refactoring**:
  - Async/await introduction
  - Multiprocessing extraction
  - Thread pool introduction

**When to Use**: Performance bottlenecks, scalability issues, resource optimization

**Prerequisites**: refactoring-patterns.md, basic performance concepts

**Example Use Cases**:
- Refactor search from O(n²) to O(n log n)
- Reduce memory from 2GB to 500MB
- Eliminate N+1 queries (50 queries → 2 queries)
- Introduce caching (2s response → 50ms)

**Learning Path**: rope-integration.md → optimization.md → advanced-patterns.md

**Estimated Reading Time**: 1.5-2 hours

**Performance Targets**: 50-80% latency reduction, 40-60% memory reduction

---

## Learning Paths

### Beginner Path (Code Quality)
**Goal**: Improve code quality through basic refactoring

**Sequence**:
1. **Start**: refactoring-patterns.md (2.5 hours)
   - Focus: Extract Method, Replace Conditional, Consolidate Duplicate Code
2. **Next**: rope-integration.md (2 hours)
   - Focus: Automated refactoring with Rope
3. **Practice**: Apply patterns to 3-5 code smells
4. **Finally**: reference.md (in parent SKILL.md)
   - Focus: Best practices, code smell catalog

**Estimated Time**: 6-8 hours
**Outcome**: Refactor code with 85% quality improvement

---

### Intermediate Path (Technical Debt Management)
**Goal**: Quantify and reduce technical debt systematically

**Sequence**:
1. **Start**: refactoring-patterns.md (2.5 hours)
2. **Next**: technical-debt.md (2 hours)
   - Focus: Debt quantification, prioritization
3. **Then**: rope-integration.md (2 hours)
   - Focus: Automated refactoring workflows
4. **Then**: optimization.md (2 hours)
   - Focus: Performance refactoring
5. **Practice**: Reduce debt score from 68 to 30

**Estimated Time**: 10-12 hours
**Outcome**: 70% technical debt reduction

---

### Expert Path (Architecture Evolution)
**Goal**: Transform system architecture through refactoring

**Sequence**:
1. **Start**: advanced-patterns.md (3 hours)
   - Focus: Microservices extraction, DDD, Clean Architecture
2. **Next**: optimization.md (2 hours)
   - Focus: Performance optimization patterns
3. **Then**: technical-debt.md (2 hours)
   - Focus: Strategic debt management
4. **Practice**: Extract 3 microservices from monolith

**Estimated Time**: 10-15 hours
**Outcome**: Successfully evolve architecture, lead refactoring initiatives

---

### Performance-Focused Path (Optimization)
**Goal**: Optimize performance through refactoring

**Sequence**:
1. **Start**: refactoring-patterns.md (2.5 hours)
   - Focus: Replace conditionals, extract methods
2. **Next**: optimization.md (2 hours)
   - Focus: All performance patterns
3. **Then**: rope-integration.md (2 hours)
   - Focus: Automated optimization workflows
4. **Practice**: Achieve 50-80% latency reduction

**Estimated Time**: 8-10 hours
**Outcome**: High-performance code, optimized algorithms

---

## Cross-References

### Internal Skill References
- **Main Skill**: [moai-essentials-refactor SKILL.md](../SKILL.md) - 5 core patterns and quick reference
- **Related Skills**:
  - `moai-essentials-debug` - AI debugging with error analysis
  - `moai-essentials-perf` - AI performance profiling with Scalene
  - `moai-essentials-review` - AI automated code review
  - `moai-foundation-trust` - TRUST 5 quality framework
  - `moai-core-code-reviewer` - Systematic code review orchestration

### External References
- **Context7 Libraries**:
  - `/rope/rope` - Rope refactoring library
  - `/psf/black` - Python code formatter
  - `/pylint-dev/pylint` - Static code analysis
  - `/eslint/eslint` - JavaScript linting with refactoring rules
- **Refactoring Resources**:
  - [Refactoring.Guru](https://refactoring.guru/) - Patterns and code smells
  - [Martin Fowler's Refactoring](https://martinfowler.com/books/refactoring.html)

---

## Search & Index

### Quick Topic Lookup

**Refactoring Patterns**:
- Extract Method → refactoring-patterns.md
- Replace Conditional → refactoring-patterns.md
- Extract Class → refactoring-patterns.md
- Move Method → refactoring-patterns.md, rope-integration.md

**Tools**:
- Rope API → rope-integration.md
- Automated refactoring → rope-integration.md
- CI/CD integration → rope-integration.md

**Technical Debt**:
- Debt quantification → technical-debt.md
- Prioritization → technical-debt.md
- Debt tracking → technical-debt.md

**Architecture**:
- Microservices → advanced-patterns.md
- Design patterns → advanced-patterns.md
- DDD → advanced-patterns.md
- Clean Architecture → advanced-patterns.md

**Performance**:
- Algorithm optimization → optimization.md
- Memory optimization → optimization.md
- Database refactoring → optimization.md
- Caching → optimization.md

---

## Module Statistics

| Module | Lines | Complexity | Est. Reading Time |
|--------|-------|------------|-------------------|
| refactoring-patterns.md | 800+ | Medium | 2-3 hours |
| rope-integration.md | 500+ | Medium | 1.5-2 hours |
| technical-debt.md | 600+ | Medium | 1.5-2 hours |
| advanced-patterns.md | 700+ | High | 2-3 hours |
| optimization.md | 500+ | Medium | 1.5-2 hours |

**Total Estimated Learning Time**: 12-18 hours (all modules)

---

## Success Metrics by Module

| Module | Target Metric | Success Criteria |
|--------|---------------|------------------|
| refactoring-patterns.md | Code quality | 85% improvement |
| rope-integration.md | Automation | 90% pattern application |
| technical-debt.md | Debt reduction | 70% reduction |
| advanced-patterns.md | Architecture transformation | 80% success rate |
| optimization.md | Performance | 50-80% latency reduction |

---

## Refactoring Decision Matrix

### When to Use Each Module

| Scenario | Module | Priority |
|----------|--------|----------|
| Complex method (>50 lines) | refactoring-patterns.md | High |
| Duplicate code (>5 instances) | refactoring-patterns.md | High |
| Large-scale refactoring | rope-integration.md | High |
| Planning sprint | technical-debt.md | Medium |
| Performance issue | optimization.md | High |
| Architecture change | advanced-patterns.md | High |

---

## Code Quality Targets

### Before Refactoring
- Cyclomatic Complexity: avg 12
- Code Duplication: 15%
- Method Length: avg 80 lines
- Test Coverage: 70%

### After Refactoring (Target)
- Cyclomatic Complexity: avg 6 (50% reduction)
- Code Duplication: 3% (80% reduction)
- Method Length: avg 20 lines (75% reduction)
- Test Coverage: ≥85% (maintained or improved)

---

## Contribution Guidelines

When adding new refactoring modules:
1. Include before/after code examples
2. Provide Rope API integration patterns
3. Add Context7 references to Refactoring.Guru
4. Document success metrics
5. Include test-first refactoring examples
6. Add troubleshooting section

---

## Integration with Parent Skill

### Core Patterns in SKILL.md
1. Extract Method with Rope
2. Context7-Enhanced Refactoring
3. Replace Conditional with Polymorphism
4. Technical Debt Quantification
5. Safe Transformation with Rollback

### Module Expansion
- **SKILL.md** provides overview and 5 core patterns
- **Modules** provide detailed implementation, examples, and advanced techniques
- **reference.md** (in parent) provides quick lookup and troubleshooting

---

**Last Updated**: 2025-11-24
**Maintained By**: MoAI-ADK
**Status**: Production Ready
**Module Architecture**: Progressive Disclosure
**Compliance Score**: 95%
