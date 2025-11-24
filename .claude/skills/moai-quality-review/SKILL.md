---
name: moai-quality-review
description: Code review consolidating automated analysis, SOLID validation, style checks, and quality gates
version: 1.0.0
modularized: true
last_updated: 2025-11-24
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
  - Read
  - Write
  - Edit
compliance_score: 85
modules:
  - review-checklists
  - quality-gates
  - best-practices
dependencies:
  - moai-foundation-trust
deprecated: false
successor: null
category_tier: 3
auto_trigger_keywords:
  - review
  - code-review
  - pr
  - pull-request
  - quality
  - checklist
  - standards
  - consistency
  - solid
  - linting
  - formatting
agent_coverage:
  - code-reviewer
  - quality-gate
context7_references: []
invocation_api_version: "1.0"
---

## Quick Reference (30 seconds)

**Enterprise Code Review Consolidated**

Unified code review framework consolidating code-reviewer and essentials-review with automated analysis, peer review checklists, SOLID validation, style consistency, and quality gates.

**Core Capabilities**:
- ✅ Automated code analysis (linting, type checking)
- ✅ SOLID principles validation
- ✅ Code style consistency (Black, Prettier)
- ✅ Review checklists for all languages
- ✅ Quality gate enforcement
- ✅ Performance impact assessment
- ✅ Security review patterns

**When to Use**:
- Code review before merge
- Quality gate enforcement
- Style consistency checks
- SOLID principles validation
- Documentation completeness
- Test coverage verification

**Core Framework**: CHECK → ANALYZE → VALIDATE → APPROVE
```
1. Automated Checks (linting, types)
   ↓
2. Code Analysis (SOLID, patterns)
   ↓
3. Quality Gates (coverage, size)
   ↓
4. Peer Review (checklists)
   ↓
5. Approval & Merge
```

---

## Core Patterns (5-10 minutes each)

### Pattern 1: Automated Code Review with Linting

**Concept**: Use linters to catch style issues before human review.

```bash
# Python: ruff
pip install ruff
ruff check src/
ruff format src/

# TypeScript: eslint + prettier
npm install --save-dev eslint prettier
npx eslint src/
npx prettier --write src/
```

**Use Case**: Enforce style consistency automatically.

---

### Pattern 2: Quality Gate Enforcement

**Concept**: Block merge if quality metrics fall below threshold.

```yaml
# .github/workflows/quality-gate.yml
name: Quality Gate

on: [pull_request]

jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      # Test Coverage
      - name: Test Coverage
        run: |
          pytest tests/ --cov=src --cov-fail-under=85
          mypy src/ --strict

      # Code Quality
      - name: Code Quality
        run: |
          ruff check src/
          ruff format --check src/

      # Security
      - name: Security Scan
        run: |
          bandit -r src/
          pip install safety && safety check
```

**Use Case**: Enforce 85%+ coverage before merge.

---

### Pattern 3: SOLID Principles Review Checklist

**Concept**: Validate SOLID principles in code review.

```markdown
# SOLID Principles Checklist

## Single Responsibility Principle
- [ ] Class has one reason to change
- [ ] Methods do one thing
- [ ] No mixed concerns (UI + business logic)

## Open-Closed Principle
- [ ] Open for extension (inheritance, composition)
- [ ] Closed for modification (stable interfaces)
- [ ] Uses abstraction (interfaces, base classes)

## Liskov Substitution Principle
- [ ] Subclasses can replace parent classes
- [ ] No unexpected behavior changes
- [ ] Contracts are honored

## Interface Segregation Principle
- [ ] Interfaces are focused and small
- [ ] Clients only depend on used methods
- [ ] No "fat" interfaces

## Dependency Inversion Principle
- [ ] Depends on abstractions, not concretions
- [ ] Uses dependency injection
- [ ] High-level doesn't depend on low-level
```

**Use Case**: Ensure SOLID compliance in architecture reviews.

---

### Pattern 4: Code Review Checklist Template

**Concept**: Standardized review checklist for PRs.

```markdown
# Code Review Checklist

## Functionality
- [ ] Feature matches requirements
- [ ] No breaking changes
- [ ] Error handling is comprehensive
- [ ] Edge cases are covered

## Code Quality
- [ ] Follows style guide (ruff/prettier)
- [ ] No duplicate code
- [ ] Naming is clear and consistent
- [ ] Complexity is acceptable

## Testing
- [ ] Unit tests added/updated
- [ ] Coverage remains ≥85%
- [ ] Integration tests pass
- [ ] E2E tests pass (if applicable)

## Documentation
- [ ] Code comments explain "why", not "what"
- [ ] API documentation updated
- [ ] Type hints present (Python/TypeScript)
- [ ] Changelog updated

## Security
- [ ] No hardcoded secrets
- [ ] Input validation implemented
- [ ] SQL injection prevented
- [ ] OWASP guidelines followed

## Performance
- [ ] No obvious performance regressions
- [ ] Benchmarks included (if relevant)
- [ ] Bundle size impact assessed
- [ ] Memory usage acceptable

## Approved!
- [ ] All comments resolved
- [ ] All checks passing
- [ ] Ready to merge
```

**Use Case**: Standardize code review across team.

---

### Pattern 5: Performance Impact Assessment

**Concept**: Analyze performance impact before merge.

```python
# Before merge: Compare metrics
def review_performance_impact(base_branch, pr_branch):
    """Compare performance between branches."""
    metrics = {
        "bundle_size": compare_bundle_size(base_branch, pr_branch),
        "test_speed": compare_test_speed(base_branch, pr_branch),
        "memory_usage": compare_memory(base_branch, pr_branch),
    }

    # Alert if regression detected
    if metrics["bundle_size"] > 5:  # >5KB increase
        print("⚠️ Bundle size increased by", metrics["bundle_size"], "KB")
    if metrics["test_speed"] < 0.8:  # >20% slower
        print("⚠️ Tests 20% slower")

    return metrics
```

**Use Case**: Detect performance regressions before merge.

---

## Advanced Documentation

For detailed review patterns:

- **[modules/review-checklists.md](modules/review-checklists.md)** - Language-specific checklists
- **[modules/quality-gates.md](modules/quality-gates.md)** - Quality gate configuration
- **[modules/best-practices.md](modules/best-practices.md)** - Code review best practices

---

## Best Practices

### ✅ DO
- Use automated tools (linting, formatting)
- Review for logic and design, not style
- Suggest improvements, not demands
- Approve and merge quickly (< 4 hours)
- Include positive feedback
- Explain reasoning in comments
- Use code review as teaching opportunity

### ❌ DON'T
- Request style changes (use autoformat)
- Block on bike-shedding
- Approve without reading code
- Leave reviews in draft for days
- Be dismissive of suggestions
- Focus only on problems
- Skip security review

---

**Status**: Production Ready
**Generated with**: MoAI-ADK Skill Factory
