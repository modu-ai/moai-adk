# Usability Analysis - TRUST 5 Framework

> Module: Usability category deep dive with maintainability metrics and code organization
> Parent: [TRUST 5 Framework](./trust5-framework.md)
> Complexity: Advanced
> Time: 15+ minutes
> Dependencies: Python 3.8+, ast, math

## Overview

Usability (25% weight) assesses maintainability and understandability through code organization evaluation, documentation quality measurement, complexity metrics calculation, and naming convention validation.

## Advanced Maintainability Metrics

### Halstead Complexity Metrics

```python
def _calculate_halstead_metrics(
    self, content: str, tree: ast.AST
) -> Dict[str, float]:
    """Calculate Halstead complexity metrics."""

    # Count operators and operands
    operators = set()
    operands = set()
    total_operators = 0
    total_operands = 0

    for node in ast.walk(tree):
        if isinstance(node, ast.BinOp):
            operators.add(type(node.op).__name__)
            total_operators += 1
        elif isinstance(node, ast.BoolOp):
            operators.add(type(node.op).__name__)
            total_operators += 1
        elif isinstance(node, ast.Name):
            operands.add(node.id)
            total_operands += 1

    n1 = len(operators)  # Unique operators
    n2 = len(operands)  # Unique operands
    N1 = total_operators  # Total operators
    N2 = total_operands  # Total operands

    # Calculate Halstead metrics
    program_length = n1 + n2
    vocabulary = n1 * math.log2(n1) + n2 * math.log2(n2) if n1 > 0 and n2 > 0 else 0
    volume = vocabulary * math.log2(vocabulary) if vocabulary > 0 else 0
    difficulty = (n1 / 2) * (N2 / n2) if n2 > 0 else 0
    effort = difficulty * volume
    time_required = effort / 18  # Seconds
    bugs_delivered = effort / 3000  # Estimated bugs

    return {
        'program_length': program_length,
        'vocabulary': vocabulary,
        'volume': volume,
        'difficulty': difficulty,
        'effort': effort,
        'time_required': time_required,
        'bugs_delivered': bugs_delivered
    }
```

### Comprehensive Maintainability Calculation

```python
def calculate_advanced_maintainability(
    self, file_path: str, content: str, tree: ast.AST
) -> Dict[str, Any]:
    """Calculate advanced maintainability metrics."""

    metrics = {}

    # Calculate Halstead metrics
    halstead = self._calculate_halstead_metrics(content, tree)
    metrics['halstead'] = halstead

    # Calculate Maintainability Index
    mi = self._calculate_maintainability_index(halstead, tree)
    metrics['maintainability_index'] = mi

    # Calculate coupling
    coupling = self._calculate_coupling(tree)
    metrics['coupling'] = coupling

    # Calculate cohesion
    cohesion = self._calculate_cohesion(tree)
    metrics['cohesion'] = cohesion

    return metrics
```

## Code Organization Assessment

### Structural Analysis

```python
def assess_code_organization(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
    """Assess code organization and structure."""

    issues = []

    # Check for circular dependencies
    circular_deps = self._detect_circular_dependencies(tree)
    issues.extend(circular_deps)

    # Check for poor separation of concerns
    separation_issues = self._check_separation_of_concerns(tree)
    issues.extend(separation_issues)

    # Check for inconsistent code style
    style_issues = self._check_code_consistency(tree)
    issues.extend(style_issues)

    return issues
```

## Detection Patterns

### Common Usability Issues

1. **High Complexity**: Functions/classes exceeding complexity thresholds
2. **Poor Naming**: Non-descriptive variable/function names
3. **Long Functions**: Functions exceeding length limits
4. **Deep Nesting**: Excessive indentation levels
5. **Magic Numbers**: Unnamed constants in code
6. **Duplicate Code**: Similar code blocks repeated
7. **Poor Documentation**: Missing or unclear docstrings

### Context7 Integration

```python
# Load usability patterns
usability = await self.context7.get_library_docs(
    context7_library_id="/code-quality/sonarqube",
    topic="maintainability metrics code smells 2025",
    tokens=4000
)
```

## Best Practices

1. Metric Thresholds: Establish project-specific complexity thresholds
2. Naming Conventions: Enforce consistent naming patterns
3. Documentation Standards: Require docstrings for public APIs
4. Code Review: Use metrics to guide code review focus
5. Refactoring: Track metric improvements over time

---

Version: 1.0.0
Last Updated: 2026-01-06
