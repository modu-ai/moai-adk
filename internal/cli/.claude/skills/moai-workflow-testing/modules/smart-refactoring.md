# Smart Refactoring with Technical Debt Management

> Module: AI-powered code refactoring with technical debt analysis and safe transformation
> Complexity: Advanced
> Time: 25+ minutes
> Dependencies: Python 3.8+, Rope, AST, Context7 MCP, asyncio, dataclasses

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
- Integration with Rope for safe code transformations

Intelligent Analysis:
- Project-specific convention detection
- API boundary identification
- Architectural pattern recognition
- Cross-file dependency analysis
- Impact estimation and effort calculation

### Key Components

TechnicalDebtAnalyzer:
- Analyzes Python codebases for technical debt patterns
- Calculates complexity metrics using AST analysis
- Detects code duplication using similarity algorithms
- Generates prioritized debt items with suggested fixes

AIRefactorer:
- Integrates technical debt analysis with refactoring opportunities
- Creates safe execution plans with risk assessment
- Leverages Context7 MCP for latest refactoring patterns
- Uses Rope library for safe code transformations

RefactorPlan:
- Comprehensive refactoring roadmap with execution strategy
- Time estimation and risk assessment
- Prerequisites and rollback strategies
- Technical debt impact tracking

---

## Quick Reference

### Installation

```bash
# Install required dependencies
pip install rope ast-visitor

# For Context7 integration (optional but recommended)
pip install context7-client
```

### Basic Usage

```python
import asyncio
from smart_refactoring import AIRefactorer

async def main():
    # Initialize refactoring system
    refactorer = AIRefactorer(context7_client=None)
    
    # Analyze and create refactoring plan
    refactor_plan = await refactorer.refactor_with_intelligence(
        codebase_path="/project/src",
        refactor_options={
            'max_risk_level': 'medium',
            'include_tests': True,
            'focus_on': ['complexity', 'duplication']
        }
    )
    
    print(f"Found {len(refactor_plan.opportunities)} opportunities")
    print(f"Estimated time: {refactor_plan.estimated_time}")
    print(f"Risk assessment: {refactor_plan.risk_assessment}")

asyncio.run(main())
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
```python
from smart_refactoring import TechnicalDebtAnalyzer

analyzer = TechnicalDebtAnalyzer()
debt_items = await analyzer.analyze("/project/src")

for item in debt_items[:5]:  # Top 5 priority items
    print(f"[{item.severity.upper()}] {item.description}")
    print(f"  File: {item.file_path}:{item.line_number}")
    print(f"  Impact: {item.impact}")
    print(f"  Estimated effort: {item.estimated_effort}")
    print(f"  Suggested: {item.suggested_fix}")
```

Step 2 - Identify Refactoring Opportunities:
```python
# AIRefactorer automatically analyzes opportunities
opportunities = refactor_plan.opportunities

for opp in opportunities[:3]:
    print(f"\n{opp.type.value}")
    print(f"  Description: {opp.description}")
    print(f"  Confidence: {opp.confidence:.0%}")
    print(f"  Risk: {opp.risk_level}")
    print(f"  Complexity reduction: {opp.complexity_reduction:.0%}")
```

Step 3 - Execute Safe Refactoring:
```python
# Execute refactoring plan in optimal order
for i, opp_index in enumerate(refactor_plan.execution_order):
    opportunity = refactor_plan.opportunities[opp_index]
    
    print(f"\nStep {i+1}: {opportunity.description}")
    print(f"Type: {opportunity.type.value}")
    print(f"Risk: {opportunity.risk_level}")
    
    # Create git commit before each operation
    # git commit -m "Before refactoring: {opportunity.description}"
    
    # Execute refactoring using Rope
    # (Implementation depends on refactoring type)
    
    # Run tests to verify
    # if tests_pass:
    #     git commit -m "After refactoring: {opportunity.description}"
    # else:
    #     git revert HEAD
```

### Configuration Options

Refactor Options:
```python
refactor_options = {
    'max_risk_level': 'medium',  # low, medium, high
    'include_tests': True,
    'focus_on': ['complexity', 'duplication', 'naming'],
    'exclude_patterns': ['*_test.py', 'test_*.py'],
    'min_confidence': 0.6,
    'complexity_threshold': 10,
    'duplication_threshold': 0.8
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

- Naming conventions (snake_case, camelCase)
- Architectural patterns (MVC, microservices)
- API boundaries (public, internal)
- Code organization preferences

See [refactoring/context-aware.md](refactoring/context-aware.md) for advanced context-aware patterns.

### Technical Debt Quantification

Track technical debt reduction over time:

```python
# Before refactoring
initial_debt = await analyzer.analyze("/project/src")
initial_score = calculate_technical_debt_score(initial_debt)

# After refactoring
final_debt = await analyzer.analyze("/project/src")
final_score = calculate_technical_debt_score(final_debt)

improvement = initial_score - final_score
print(f"Technical debt reduced by {improvement:.1%}")
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

- Rope: Python refactoring library
- AST: Python built-in AST module
- Context7 MCP: Latest refactoring patterns (optional)

### Related Modules

- [AI Debugging](./ai-debugging.md) - Debugging with AI assistance
- [Performance Optimization](./performance-optimization.md) - Performance improvement techniques
- [Code Review](./code-review/) - Automated code review patterns

### External References

- Refactoring Guru: https://refactoring.guru/
- Python AST Documentation: https://docs.python.org/3/library/ast.html
- Rope Documentation: https://github.com/python-rope/rope

---

Module: `modules/smart-refactoring.md`
Related: [refactoring/patterns.md](refactoring/patterns.md) | [refactoring/ai-workflows.md](refactoring/ai-workflows.md) | [AI Debugging](./ai-debugging.md)
