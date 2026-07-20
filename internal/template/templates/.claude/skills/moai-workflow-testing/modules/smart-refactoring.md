# Smart Refactoring with Technical Debt Management

> Module: AI-powered code refactoring with technical debt analysis and safe transformation
> Complexity: Advanced
> Time: 25+ minutes
> Dependencies: source parser (AST), the project's refactoring tool, WebSearch/WebFetch (optional)

## Overview

Smart refactoring combines AI analysis with traditional refactoring tools to identify, prioritize, and safely execute code transformations while quantifying and reducing technical debt.

### Core Capabilities

Technical Debt Analysis:
- Automated code complexity detection (cyclomatic, cognitive, nesting depth)
- Code duplication identification across files
- Method length and parameter count analysis
- Naming convention violation detection
- Severity-based prioritization (critical, high, medium, low)

AI-Powered Refactoring:
- Context-aware refactoring opportunity identification
- Safe transformation planning with risk assessment
- Execution order optimization (low-risk first, then high-impact)
- Rollback strategy generation
- Integration with the host language's refactoring tool for safe code transformations

Intelligent Analysis:
- Project-specific convention detection
- API boundary identification
- Architectural pattern recognition
- Cross-file dependency analysis
- Impact estimation and effort calculation

### Key Components

TechnicalDebtAnalyzer:
- Analyzes a codebase for technical debt patterns
- Calculates complexity metrics via AST analysis
- Detects code duplication using similarity algorithms
- Generates prioritized debt items with suggested fixes

AIRefactorer:
- Integrates technical debt analysis with refactoring opportunities
- Creates safe execution plans with risk assessment
- Leverages WebSearch/WebFetch for latest refactoring patterns
- Uses the host language's refactoring tooling for safe code transformations

RefactorPlan:
- Comprehensive refactoring roadmap with execution strategy
- Effort estimate and risk assessment
- Prerequisites and rollback strategies
- Technical debt impact tracking

---

## Quick Reference

### Setup

Use the host language's own refactoring tooling — no extra packages are required for the workflow itself. Examples per language: Python Rope / pyrefly; Go `gopls` refactor; TS/JS language-server rename/extract; Rust rust-analyzer; IntelliJ family. WebSearch/WebFetch integration is optional.

### Basic Usage

```text
# Initialize the refactoring system (Documentation client is optional)
refactorer = AIRefactorer(docs_client=none)

# Analyze and create a refactoring plan
plan = refactorer.refactor_with_intelligence(
    codebase_path="/project/src",
    refactor_options={
        max_risk_level: "medium",
        include_tests:  true,
        focus_on:       ["complexity", "duplication"]
    })

print("Found " + len(plan.opportunities) + " opportunities")
print("Estimated effort: " + plan.estimated_effort)
print("Risk assessment: " + plan.risk_assessment)
```

### Technical Debt Categories

Code Complexity:
- Cyclomatic complexity > 10 (medium), > 15 (high), > 20 (critical)
- Cognitive complexity > 15
- Nesting depth > 4 levels

Duplication:
- Similar code blocks across files
- Repeated patterns with > 80% similarity
- Copied and pasted code segments

Method Issues:
- Methods > 50 lines
- Functions with > 7 parameters
- Multiple responsibilities in single method

Naming:
- Single-letter variables (except loop counters)
- Temp/tmp prefix variables
- Non-descriptive abbreviations

### Refactoring Types

Extract Method:
- Trigger: Method > 30 lines or complexity > 8
- Risk: Low to medium
- Impact: Reduces complexity, improves readability

Extract Variable:
- Trigger: Complex expressions with multiple operations
- Risk: Low
- Impact: Improves code comprehension

Reorganize Imports:
- Trigger: > 10 imports in single file
- Risk: Low
- Impact: Better dependency management

Inline Variable:
- Trigger: Variables used once with simple values
- Risk: Low
- Impact: Reduces unnecessary indirection

Move Module:
- Trigger: Logical grouping opportunities
- Risk: Medium to high
- Impact: Better architecture, reduced coupling

---

## Implementation Guide

### Workflow Overview

Step 1 - Analyze Technical Debt:
```text
analyzer = TechnicalDebtAnalyzer()
debt_items = analyzer.analyze("/project/src")

for item in debt_items[:5]:   # top 5 priority items
    print("[" + uppercase(item.severity) + "] " + item.description)
    print("  File: " + item.file_path + ":" + item.line_number)
    print("  Impact: " + item.impact + "  Effort: " + item.estimated_effort)
    print("  Suggested: " + item.suggested_fix)
```

Step 2 - Identify Refactoring Opportunities:
```text
# AIRefactorer automatically analyzes opportunities
for opp in refactor_plan.opportunities[:3]:
    print(opp.type)
    print("  Description: " + opp.description)
    print("  Confidence: " + pct(opp.confidence))
    print("  Risk: " + opp.risk_level)
    print("  Complexity reduction: " + pct(opp.complexity_reduction))
```

Step 3 - Execute Safe Refactoring:
```text
# Execute the refactoring plan in optimal order
for (i, opp_index) in enumerate(refactor_plan.execution_order, from=1):
    opportunity = refactor_plan.opportunities[opp_index]
    print("Step " + i + ": " + opportunity.description)
    print("  Type: " + opportunity.type + "  Risk: " + opportunity.risk_level)

    # Create a git commit before each operation
    # git commit -m "Before refactoring: {description}"
    # Execute the refactoring via the host language's refactoring tool
    # Run tests to verify
    #   if pass: git commit -m "After refactoring: {description}"
    #   else:    git revert HEAD
```

### Configuration Options

Refactor Options:
```text
refactor_options = {
    max_risk_level:        "medium",                  # low, medium, high
    include_tests:         true,
    focus_on:              ["complexity", "duplication", "naming"],
    exclude_patterns:      ["*_test.<ext>", "test_*.<ext>"],   # host language's test naming
    min_confidence:        0.6,
    complexity_threshold:  10,
    duplication_threshold: 0.8
}
```

### Integration with Testing

Pre-Refactoring Checklist:
- Comprehensive test suite exists
- All tests passing
- Test coverage > 80%
- Performance benchmarks recorded

Post-Refactoring Verification:
- Run full test suite
- Verify performance benchmarks
- Check for breaking changes
- Update documentation

### Rollback Strategy

Safe Refactoring Protocol:
1. Create git commit before each operation
2. Run automated tests after each change
3. Maintain detailed change log
4. Use git revert for individual rollbacks
5. Keep backup of original codebase

---

## Advanced Features

### Context-Aware Refactoring

The AIRefactorer can detect and respect project-specific conventions:

- Naming conventions (snake_case, camelCase, kebab-case)
- Architectural patterns (MVC, microservices)
- API boundaries (public, internal)
- Code organization preferences

See [refactoring/context-aware.md](refactoring/context-aware.md) for advanced context-aware patterns.

### Technical Debt Quantification

Track technical debt reduction over time:

```text
# Before refactoring
initial_debt  = analyzer.analyze("/project/src")
initial_score = calculate_technical_debt_score(initial_debt)

# After refactoring
final_debt    = analyzer.analyze("/project/src")
final_score   = calculate_technical_debt_score(final_debt)

improvement = initial_score - final_score
print("Technical debt reduced by " + pct(improvement))
```

### Safe Refactoring Patterns

For detailed refactoring techniques and best practices, see:
- [refactoring/patterns.md](refactoring/patterns.md) - Specific refactoring techniques
- [refactoring/ai-workflows.md](refactoring/ai-workflows.md) - AI-assisted refactoring workflows

---

## Best Practices

1. Incremental Refactoring: Apply changes incrementally with testing at each step
2. Test Coverage: Ensure comprehensive test coverage before major refactoring
3. Version Control: Commit changes before and after each major refactoring step
4. Documentation: Update documentation to reflect refactored code structure
5. Performance Monitoring: Monitor performance impact of refactoring changes

---

## Resources

### Dependencies

- Refactoring tool: the host language's refactoring library/LSP (Rope, gopls, rust-analyzer, etc.)
- Source parser: the host language's AST module
- WebSearch/WebFetch: Latest refactoring patterns (optional)

### Related Modules

- [AI Debugging](./ai-debugging.md) - Debugging with AI assistance
- [Performance Optimization](./performance-optimization.md) - Performance improvement techniques
- [Code Review](./code-review/) - Automated code review patterns

### External References

- Refactoring Guru: https://refactoring.guru/
- Martin Fowler's Refactoring catalog

---

Module: `modules/smart-refactoring.md`
Related: [refactoring/patterns.md](refactoring/patterns.md) | [refactoring/ai-workflows.md](refactoring/ai-workflows.md) | [AI Debugging](./ai-debugging.md)
