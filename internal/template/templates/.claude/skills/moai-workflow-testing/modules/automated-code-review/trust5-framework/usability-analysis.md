# Usability Analysis - TRUST 5 Framework

> Module: Usability category deep dive with maintainability metrics and code organization
> Parent: [TRUST 5 Framework](./trust5-framework.md)
> Complexity: Advanced
> Time: 15+ minutes
> Dependencies: source parser (AST)

## Overview

Usability (25% weight) assesses maintainability and understandability through code organization evaluation, documentation quality measurement, complexity metrics calculation, and naming convention validation.

## Advanced Maintainability Metrics

### Halstead Complexity Metrics

Halstead metrics count distinct and total operators/operands in the syntax tree. Walk the host language's AST to collect them (binary operators, boolean operators as operators; identifiers/literals as operands).

```text
calculate_halstead_metrics(content, tree):
    operators = set()        # unique operator kinds
    operands  = set()        # unique operand identifiers
    total_operators = 0
    total_operands  = 0

    for node in walk(tree):
        if node is a BinaryOp:  operators.add(kind(node.op)); total_operators += 1
        if node is a BoolOp:    operators.add(kind(node.op)); total_operators += 1
        if node is an Identifier: operands.add(node.name);    total_operands  += 1

    n1 = len(operators)   # unique operators
    n2 = len(operands)    # unique operands
    N1 = total_operators  # total operators
    N2 = total_operands   # total operands

    # Halstead formulas
    program_length = n1 + n2
    vocabulary     = n1*log2(n1) + n2*log2(n2)   if n1 > 0 and n2 > 0 else 0
    volume         = vocabulary * log2(vocabulary) if vocabulary > 0 else 0
    difficulty     = (n1/2) * (N2/n2)             if n2 > 0 else 0
    effort         = difficulty * volume
    time_required  = effort / 18        # seconds
    bugs_delivered = effort / 3000      # estimated bugs

    return { program_length, vocabulary, volume, difficulty, effort, time_required, bugs_delivered }
```

### Comprehensive Maintainability Calculation

```text
calculate_advanced_maintainability(file_path, content, tree):
    metrics = {}
    metrics.halstead              = calculate_halstead_metrics(content, tree)
    metrics.maintainability_index = calculate_maintainability_index(metrics.halstead, tree)
    metrics.coupling              = calculate_coupling(tree)
    metrics.cohesion              = calculate_cohesion(tree)
    return metrics
```

## Code Organization Assessment

### Structural Analysis

```text
assess_code_organization(file_path, tree):
    issues = []
    issues.extend(detect_circular_dependencies(tree))
    issues.extend(check_separation_of_concerns(tree))
    issues.extend(check_code_consistency(tree))
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
7. **Poor Documentation**: Missing or unclear doc comments

### Documentation Integration

```text
# Load usability patterns
usability = docs.get_library_docs(
    "<code-quality/sonarqube>",
    topic="maintainability metrics code smells",
    tokens=4000)
```

## Best Practices

1. Metric Thresholds: Establish project-specific complexity thresholds
2. Naming Conventions: Enforce consistent naming patterns
3. Documentation Standards: Require doc comments for public APIs
4. Code Review: Use metrics to guide code review focus
5. Refactoring: Track metric improvements over time

---

Version: 1.0.0
Last Updated: 2026-01-06
