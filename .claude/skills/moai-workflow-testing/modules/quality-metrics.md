# Quality Metrics and Complexity Analysis

> Module: Code quality assessment, complexity calculation, and maintainability metrics
> Parent: [Automated Code Review](./automated-code-review.md)
> Complexity: Intermediate
> Time: 15+ minutes
> Dependencies: source parser (AST), complexity tool for the host language

## Quick Reference

### Quality Metrics Categories

Code Complexity Metrics:
- Cyclomatic Complexity: Decision complexity measurement
- Cognitive Complexity: Human cognitive load estimation
- Nesting Depth: Control flow nesting levels
- Function Length: Lines of code per function
- Parameter Count: Number of function parameters

Maintainability Indices:
- Maintainability Index: Overall maintainability score
- Technical Debt: Effort required to fix issues
- Code Duplication: Repeated code patterns
- Comment Ratio: Documentation coverage
- Test Coverage: Test completeness (requires the project's coverage tool)

Code Smell Detection:
- Long Methods: Functions exceeding length thresholds
- God Classes: Classes with too many responsibilities
- Feature Envy: Methods that use other classes more
- Data Clumps: Group of data items that appear together
- Primitive Obsession: Overuse of primitive types

### Core Implementation

The analyzer walks a parsed syntax tree (AST) and computes metrics. Use the host language's parser (e.g. Go `go/ast`, Python `ast`, TS compiler API, Rust `syn`, Java JavaParser); many languages also ship a dedicated complexity tool (Go `gocyclo`, Python `radon`/`mccabe`, JS `escomplex`).

```text
class QualityMetricsAnalyzer:
    complexity_thresholds = {
        cyclomatic: 10, cognitive: 15, nesting_depth: 4,
        function_length: 50, parameter_count: 7
    }

    analyze_file_quality(file_path, content, tree):
        return {
            complexity:      complexity_metrics(tree),
            maintainability: maintainability_metrics(content, tree),
            code_smells:     detect_code_smells(file_path, content, tree),
            documentation:   analyze_documentation(tree),
            statistics:      file_statistics(content, tree)
        }
```

---

## Implementation Guide

### Cyclomatic Complexity

```text
calculate_cyclomatic_complexity(node):
    complexity = 1   # base
    for child in walk(node):
        if child is a decision point (if, while, for, catch, with/using, case):
            complexity += 1
        if child is a boolean operator (and/or) with N operands:
            complexity += (N - 1)
        if child is a pattern-match with N cases:        # languages with match/switch-expr
            complexity += N
    return complexity

analyze_function_complexity(file_path, tree):
    issues = []
    for node in walk(tree, matching=FunctionDecl):
        c = calculate_cyclomatic_complexity(node)
        if c > complexity_thresholds.cyclomatic:
            severity = "high" if c > 20 else "medium"
            issues.append(CodeIssue(
                id="complexity_" + node.line,
                category=TrustCategory.USABILITY, severity=severity,
                issue_type="code_smell", title="High Cyclomatic Complexity",
                description="Function '" + node.name + "' has cyclomatic complexity " + c,
                file_path=file_path, line_number=node.line, column_number=1,
                code_snippet=signature(node) + "  # complexity: " + c,
                suggested_fix="Consider refactoring '" + node.name + "' to reduce complexity",
                confidence=0.9, rule_violated="COMPLEXITY"))
    return issues
```

Cyclomatic Complexity Interpretation:
1-10: Simple, low risk
11-20: Moderate complexity, medium risk
21-50: High complexity, high risk
51+: Very high complexity, very high risk

### Nesting Depth Analysis

```text
calculate_nesting_depth(node, current_depth = 0):
    max_depth = current_depth
    for child in nested_blocks_of(node):     # if/while/for/with/try bodies
        max_depth = max(max_depth, calculate_nesting_depth(child, current_depth + 1))
    return max_depth

analyze_nesting_depth(file_path, tree):
    issues = []
    for node in walk(tree, matching=FunctionDecl):
        d = calculate_nesting_depth(node)
        if d > complexity_thresholds.nesting_depth:
            issues.append(CodeIssue(
                id="nesting_" + node.line,
                category=TrustCategory.USABILITY, severity="medium",
                issue_type="code_smell", title="Deep Nesting",
                description="Function '" + node.name + "' has nesting depth " + d,
                file_path=file_path, line_number=node.line, column_number=1,
                code_snippet=signature(node) + "  # nesting depth: " + d,
                suggested_fix="Use early returns or extract methods to reduce nesting",
                confidence=0.8, rule_violated="NESTING_DEPTH"))
    return issues
```

### Function Length Analysis

```text
analyze_function_length(file_path, tree):
    issues = []
    max_lines = complexity_thresholds.function_length
    for node in walk(tree, matching=FunctionDecl):
        # Count source lines of the function body, excluding the doc comment
        # and blank lines.
        code_lines = count_code_lines(node, excluding=[doc_comment, blank_lines])
        if code_lines > max_lines:
            issues.append(CodeIssue(
                id="func_length_" + node.line,
                category=TrustCategory.USABILITY, severity="medium",
                issue_type="code_smell", title="Long Function",
                description="Function '" + node.name + "' is " + code_lines + " lines long",
                file_path=file_path, line_number=node.line, column_number=1,
                code_snippet=signature(node) + "  # " + code_lines + " lines",
                suggested_fix="Break '" + node.name + "' into smaller, focused functions",
                confidence=0.8, rule_violated="FUNC_LENGTH"))
    return issues
```

### Parameter Count Analysis

```text
analyze_parameter_count(file_path, tree):
    issues = []
    max_params = complexity_thresholds.parameter_count
    for node in walk(tree, matching=FunctionDecl):
        param_count = len(node.parameters)
        if param_count > max_params:
            issues.append(CodeIssue(
                id="param_count_" + node.line,
                category=TrustCategory.USABILITY, severity="medium",
                issue_type="code_smell", title="Too Many Parameters",
                description="Function '" + node.name + "' has " + param_count +
                            " parameters (max: " + max_params + ")",
                file_path=file_path, line_number=node.line, column_number=1,
                code_snippet=signature(node),
                suggested_fix="Consider grouping parameters into a data/option record for '" +
                              node.name + "'",
                confidence=0.7, rule_violated="PARAMETER_COUNT"))
    return issues
```

### Maintainability Index

```text
calculate_maintainability_index(content, tree):
    volume        = halstead_volume(content, tree)     # token-based vocabulary metric
    complexity    = total_complexity(tree)
    lines_of_code = count non-blank lines of content
    comment_lines = count lines whose first non-blank token is a comment
    comment_ratio = comment_lines / max(lines_of_code, 1)

    # Maintainability Index (MI), the standard formula:
    #   MI = 171 - 5.2*ln(volume) - 0.23*complexity - 16.2*ln(loc)
    mi = 171 - 5.2*ln(max(volume,1)) - 0.23*complexity - 16.2*ln(max(lines_of_code,1))
    mi = mi * (1 + comment_ratio)        # adjust for comment ratio
    mi = clamp(mi, 0, 100)               # normalize to 0-100

    return {
        mi_score: mi, mi_rating: mi_rating(mi),
        halstead_volume: volume, cyclomatic_complexity: complexity,
        lines_of_code: lines_of_code, comment_ratio: comment_ratio
    }

mi_rating(mi_score):
    if mi_score >= 85: return "Excellent"
    if mi_score >= 70: return "Good"
    if mi_score >= 55: return "Moderate"
    if mi_score >= 40: return "Poor"
    return "Very Poor"
```

### Code Smell Detection

```text
detect_code_smells(file_path, content, tree):
    smells = []
    smells.extend(analyze_function_length(file_path, tree))
    smells.extend(analyze_function_complexity(file_path, tree))
    smells.extend(analyze_nesting_depth(file_path, tree))
    smells.extend(analyze_parameter_count(file_path, tree))

    # God classes (too many methods)
    for node in walk(tree, matching=ClassDecl):
        methods = count of methods in node
        if methods > 20:
            smells.append(CodeIssue(
                id="god_class_" + node.line,
                category=TrustCategory.USABILITY, severity="medium",
                issue_type="code_smell", title="God Class",
                description="Class '" + node.name + "' has " + methods + " methods (max: 20)",
                file_path=file_path, line_number=node.line, column_number=1,
                code_snippet="class " + node.name + "  # " + methods + " methods",
                suggested_fix="Consider splitting '" + node.name + "' into smaller, focused classes",
                confidence=0.7, rule_violated="GOD_CLASS"))
    return smells
```

---

## Custom Thresholds

```text
# Customize complexity thresholds to match project standards
analyzer.complexity_thresholds = {
    cyclomatic: 15,       # more lenient
    cognitive: 20,
    nesting_depth: 5,
    function_length: 75,  # allow longer functions
    parameter_count: 10   # allow more parameters
}
```

---

## Best Practices

1. Threshold Customization: Adjust thresholds to match project standards
2. Progressive Improvement: Set realistic targets and improve gradually
3. Team Consistency: Use consistent thresholds across entire codebase
4. Regular Review: Monitor metrics trends over time
5. Refactoring Priority: Focus on high-complexity, high-risk code first
6. Documentation Balance: Balance comment ratio with self-documenting code
7. Test Coverage: Combine quality metrics with test coverage analysis
8. Technical Debt: Track and prioritize technical debt reduction

---

## Related Modules

- [TRUST 5 Validation](./trust5-validation.md): Usability category scoring
- [automated-code-review/trust5-framework.md](./automated-code-review/trust5-framework.md): Advanced quality patterns

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: `modules/quality-metrics.md`
