# Truthfulness Analysis - TRUST 5 Framework

> Module: Truthfulness category deep dive with logic correctness and data flow analysis
> Parent: [TRUST 5 Framework](./trust5-framework.md)
> Complexity: Advanced
> Time: 10+ minutes
> Dependencies: source parser (AST), WebSearch/WebFetch

## Overview

Truthfulness (25% weight) validates code correctness and logic accuracy through comprehensive analysis of algorithmic correctness, logic error detection, data flow integrity, and contract compliance.

## Logic Correctness Validation

### Tautology Detection

```text
detect_tautologies(tree):
    issues = []
    for node in walk(tree, matching=Comparison):
        # Flag comparisons that are always true, e.g. x > -1 when x is a length.
        if is_always_true_comparison(node):
            issues.append(CodeIssue(
                id="tautology_" + node.line,
                category=TrustCategory.TRUTHFULNESS, severity="low",
                issue_type="code_smell", title="Tautological Comparison",
                description="Comparison is always true",
                file_path=file_path, line_number=node.line, column_number=node.column,
                code_snippet="# tautological comparison detected",
                suggested_fix="Remove the unnecessary comparison or simplify the logic",
                confidence=0.7, rule_violated="TAUTOLOGICAL_COMPARISON"))
    return issues
```

### Comprehensive Logic Validation

```text
validate_logic_correctness(file_path, tree):
    issues = []
    issues.extend(detect_tautologies(tree))
    issues.extend(detect_contradictions(tree))
    issues.extend(detect_constant_conditions(tree))
    issues.extend(detect_type_confusion(tree))
    return issues
```

## Data Flow Analysis

### Variable Usage Validation

```text
analyze_data_flow(file_path, tree):
    issues = []
    issues.extend(check_undefined_variables(tree))
    issues.extend(check_unused_variables(tree))
    issues.extend(check_variable_shadowing(tree))
    return issues
```

## Detection Patterns

### Common Truthfulness Issues

1. **Tautological Comparisons**: Conditions that always evaluate to True/False
2. **Contradictory Conditions**: Mutually exclusive conditions in same logic path
3. **Constant Conditions**: Conditions using only constant values
4. **Type Confusion**: Operations between incompatible types
5. **Undefined Variables**: References to variables before definition
6. **Unused Variables**: Variables defined but never read
7. **Variable Shadowing**: Inner scope variables hiding outer scope

### Documentation Integration

```text
# Load truthfulness patterns (use the host language's correctness library id)
truthfulness = docs.get_library_docs(
    "<code-correctness/<language>>",
    topic="logic error detection patterns",
    tokens=3000)
```

## Best Practices

1. Pattern Matching: Use AST patterns to detect logic errors
2. Type Inference: Apply type inference to catch type confusion
3. Flow Analysis: Track variable usage across scopes
4. Context Awareness: Consider project-specific logic patterns
5. False Positive Reduction: Use confidence scoring to reduce noise

---

Version: 1.0.0
Last Updated: 2026-01-06
