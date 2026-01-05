# Truthfulness Analysis - TRUST 5 Framework

> Module: Truthfulness category deep dive with logic correctness and data flow analysis
> Parent: [TRUST 5 Framework](./trust5-framework.md)
> Complexity: Advanced
> Time: 10+ minutes
> Dependencies: Python 3.8+, ast, Context7 MCP

## Overview

Truthfulness (25% weight) validates code correctness and logic accuracy through comprehensive analysis of algorithmic correctness, logic error detection, data flow integrity, and contract compliance.

## Logic Correctness Validation

### Tautology Detection

```python
def _detect_tautologies(self, tree: ast.AST) -> List[CodeIssue]:
    """Detect tautological comparisons (always True)."""

    issues = []

    for node in ast.walk(tree):
        if isinstance(node, ast.Compare):
            # Check for comparisons that are always True
            # Example: x > -1 (where x is len(something))
            if self._is_always_true_comparison(node):
                issue = CodeIssue(
                    id=f"tautology_{node.lineno}",
                    category=TrustCategory.TRUTHFULNESS,
                    severity="low",
                    issue_type="code_smell",
                    title="Tautological Comparison",
                    description="Comparison is always True",
                    file_path=file_path,
                    line_number=node.lineno,
                    column_number=node.col_offset,
                    code_snippet="# Tautological comparison detected",
                    suggested_fix="Remove unnecessary comparison or simplify logic",
                    confidence=0.7,
                    rule_violated="TAUTOLOGICAL_COMPARISON"
                )
                issues.append(issue)

    return issues
```

### Comprehensive Logic Validation

```python
async def validate_logic_correctness(
    self, file_path: str, tree: ast.AST
) -> List[CodeIssue]:
    """Comprehensive logic correctness validation."""

    issues = []

    # Check for tautological comparisons
    tautologies = self._detect_tautologies(tree)
    issues.extend(tautologies)

    # Check for contradictory conditions
    contradictions = self._detect_contradictions(tree)
    issues.extend(contradictions)

    # Check for constant conditions
    constant_conditions = self._detect_constant_conditions(tree)
    issues.extend(constant_conditions)

    # Check for type confusion
    type_issues = self._detect_type_confusion(tree)
    issues.extend(type_issues)

    return issues
```

## Data Flow Analysis

### Variable Usage Validation

```python
def analyze_data_flow(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
    """Analyze data flow for correctness issues."""

    issues = []

    # Check for undefined variables
    undefined_vars = self._check_undefined_variables(tree)
    issues.extend(undefined_vars)

    # Check for unused variables
    unused_vars = self._check_unused_variables(tree)
    issues.extend(unused_vars)

    # Check for variable shadowing
    shadowing = self._check_variable_shadowing(tree)
    issues.extend(shadowing)

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

### Context7 Integration

```python
# Load truthfulness patterns
truthfulness = await self.context7.get_library_docs(
    context7_library_id="/code-correctness/python",
    topic="logic error detection patterns 2025",
    tokens=3000
)
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
