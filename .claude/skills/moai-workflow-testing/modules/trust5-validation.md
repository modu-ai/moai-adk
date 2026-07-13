# TRUST 5 Validation Framework

> Module: Complete TRUST 5 validation implementation with category-specific analysis and scoring
> Parent: [Automated Code Review](./automated-code-review.md)
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: source parser, WebSearch/WebFetch

## Quick Reference

### TRUST 5 Categories

Truthfulness (25% weight):
- Code correctness validation
- Logic error detection
- Unreachable code identification
- Comparison issue checking
- Data flow analysis

Relevance (20% weight):
- Requirements fulfillment
- TODO/FIXME comment tracking
- Dead code detection
- Feature completeness validation
- Purpose alignment checking

Usability (25% weight):
- Maintainability assessment
- Code complexity analysis
- Documentation completeness
- Naming convention validation
- Code organization review

Safety (20% weight):
- Security vulnerability detection
- Error handling validation
- Exception safety checking
- Resource leak detection
- Input validation review

Timeliness (10% weight):
- Performance optimization opportunities
- Deprecated code identification
- Modern practices adoption
- Technology currency validation
- Standards compliance checking

### Core Implementation

The category enum and issue record are language-neutral. The analysis passes below operate on a parsed syntax tree (AST) — use the host language's parser (e.g. Go `go/ast`, Python `ast`, TypeScript compiler API, Rust `syn`, Java JavaParser); the checks themselves generalize across trees.

```text
enum TrustCategory:
    TRUTHFULNESS
    RELEVANCE
    USABILITY
    SAFETY
    TIMELINESS

record CodeIssue:
    id:                text
    category:          TrustCategory
    severity:          text   # critical, high, medium, low, info
    issue_type:        text
    title:             text
    description:       text
    file_path:         text
    line_number:       int
    column_number:     int
    code_snippet:      text
    suggested_fix:     text
    confidence:        float  # 0.0 to 1.0
    rule_violated:     text?  # default none
    external_reference:text?  # default none
```

---

## Implementation Guide

### Truthfulness Analysis

Truthfulness validation focuses on code correctness and logic accuracy:

```text
analyze_truthfulness(file_path, tree):
    issues = []
    # Walk every function/method declaration in the syntax tree
    for node in walk(tree, matching=FunctionDecl):
        issues.extend(check_unreachable_code(file_path, node))
    issues.extend(check_logic_issues(file_path, tree))
    return issues
```

Unreachable Code Detection:
- Identifies code after return statements
- Detects code after throw/raise statements
- Finds code after break/continue in loops
- Reports dead code with confidence scores

Logic Issue Detection:
- Checks for null/none comparison anti-patterns (prefer the language's identity-check idiom over equality with a null sentinel)
- Identifies constant conditions in if statements
- Detects tautological comparisons
- Finds contradictory conditions

### Relevance Analysis

Relevance analysis validates requirements fulfillment and purpose alignment:

```text
analyze_relevance(file_path, content):
    issues = []
    for (line_num, line) in enumerate(lines(content), from=1):
        if "TODO:" in line or "FIXME:" in line:
            issues.append(CodeIssue(
                id="todo_" + line_num,
                category=TrustCategory.RELEVANCE, severity="low",
                issue_type="documentation_issue",
                title="Unresolved TODO",
                description="TODO/FIXME comment found: " + trim(line),
                file_path=file_path, line_number=line_num,
                column_number=index_of(line, "TODO", "FIXME"),
                code_snippet=trim(line),
                suggested_fix="Address the TODO/FIXME item or remove the comment",
                confidence=0.6, rule_violated="UNRESOLVED_TODO"))
    return issues
```

Relevance Checks:
- TODO/FIXME comment tracking
- Dead code identification (unused imports, variables, functions)
- Feature completeness validation
- Documentation alignment with implementation

### Usability Analysis

Usability assessment focuses on maintainability and code quality:

```text
analyze_usability(file_path, content, tree):
    issues = []
    for node in walk(tree, matching=FunctionDecl):
        if not has_doc_comment(node):       # docstring / godoc / JSDoc / rustdoc
            issues.append(CodeIssue(
                id="no_doc_" + node.line,
                category=TrustCategory.USABILITY, severity="low",
                issue_type="documentation_issue",
                title="Missing Documentation",
                description="Function '" + node.name + "' has no doc comment",
                file_path=file_path, line_number=node.line, column_number=1,
                code_snippet=signature(node),
                suggested_fix="Add a doc comment to '" + node.name +
                              "' explaining its purpose, parameters, and return value",
                confidence=0.7, rule_violated="MISSING_DOC"))
    return issues
```

Usability Metrics:
- Function length analysis (default max: 50 lines)
- Cyclomatic complexity calculation (default max: 10)
- Nesting depth assessment (default max: 4 levels)
- Documentation completeness
- Naming convention validation

### Safety Analysis

Safety validation detects security vulnerabilities and error handling issues:

```text
analyze_safety(file_path, tree):
    issues = []
    for node in walk(tree, matching=CatchClause):
        if catches_all_untyped(node):    # bare catch / catch-all without a typed error
            issues.append(CodeIssue(
                id="bare_catch_" + node.line,
                category=TrustCategory.SAFETY, severity="medium",
                issue_type="code_smell",
                title="Bare / Catch-All Handler",
                description="A catch-all handler can swallow unexpected errors",
                file_path=file_path, line_number=node.line, column_number=1,
                code_snippet="catch (untyped)",
                suggested_fix="Catch specific error types, or catch the base error type with logging",
                confidence=0.8, rule_violated="BARE_CATCH"))
    return issues
```

Safety Checks:
- Bare / catch-all handler detection
- Exception handling validation
- Resource leak detection (file handles, database connections)
- Input validation review
- Resource-cleanup idiom validation (defer / using / try-with-resources / RAII)

### Timeliness Analysis

Timeliness assessment identifies performance and modernization opportunities:

```text
analyze_timeliness(file_path, content):
    issues = []
    # Map deprecated APIs/symbols to their modern replacements for the host language
    deprecated = { "oldSymbol": "modernSymbol", ... }    # populated per language/runtime
    for (line_num, line) in enumerate(lines(content), from=1):
        for (old, modern) in deprecated:
            if references(line, old):
                issues.append(CodeIssue(
                    id="deprecated_" + line_num,
                    category=TrustCategory.TIMELINESS, severity="low",
                    issue_type="deprecated_api",
                    title="Deprecated API",
                    description="Using deprecated '" + old + "', should use '" + modern + "'",
                    file_path=file_path, line_number=line_num,
                    column_number=index_of(line, old),
                    code_snippet=trim(line),
                    suggested_fix="Replace '" + old + "' with '" + modern + "'",
                    confidence=0.9, rule_violated="DEPRECATED_API", auto_fixable=true))
    return issues
```

Timeliness Indicators:
- Deprecated API/import detection
- Performance anti-pattern identification
- Modern language-feature adoption
- Standards compliance checking
- Technology currency validation

---

## Score Calculation

### Category Score Algorithm

```text
calculate_trust_scores(issues, metrics):
    category_scores = {}
    category_weights = {
        TRUTHFULNESS: 0.25, RELEVANCE: 0.20, USABILITY: 0.25,
        SAFETY: 0.20, TIMELINESS: 0.10
    }
    issues_by_category = group issues by category

    for category in TrustCategory:
        penalty = 0.0
        for issue in issues_by_category[category]:
            severity_penalty = {
                critical: 0.5, high: 0.3, medium: 0.1, low: 0.05, info: 0.01
            }
            penalty += severity_penalty[issue.severity] default 0.1) * issue.confidence
        category_scores[category] = max(0.0, 1.0 - min(penalty, 1.0))

    overall = sum(category_scores[c] * category_weights[c] for c in TrustCategory)
    return { overall, categories: category_scores }
```

### Score Interpretation

0.9 - 1.0: Excellent quality, minimal issues
0.8 - 0.9: Good quality, some minor issues
0.7 - 0.8: Acceptable quality, moderate issues
0.6 - 0.7: Needs improvement, significant issues
0.0 - 0.6: Poor quality, critical issues present

---

## Advanced Customization

### Custom Category Weights

Adjust category weights to match project priorities:

```text
reviewer.category_weights = {
    TRUTHFULNESS: 0.30,   # increase emphasis on correctness
    RELEVANCE:    0.15,
    USABILITY:    0.20,
    SAFETY:       0.30,   # increase emphasis on security
    TIMELINESS:   0.05
}
```

### Custom Severity Penalties

Modify penalty values for severity levels:

```text
reviewer.severity_penalties = {
    critical: 0.7,   # stricter penalties
    high:     0.4,
    medium:   0.15,
    low:      0.05,
    info:     0.0
}
```

### Custom Rule Configuration

Add custom validation rules:

```text
class CustomTruthfulnessAnalyzer:
    analyze_custom_patterns(file_path, tree):
        issues = []
        # Add custom logic validation
        # Add project-specific correctness checks
        # Add domain-specific validation rules
        return issues

# Integrate custom analyzer
reviewer.custom_analyzers[TRUTHFULNESS] = CustomTruthfulnessAnalyzer()
```

---

## Best Practices

1. Category Balance: Maintain balanced category weights appropriate for project context
2. Severity Calibration: Adjust severity penalties to match team quality standards
3. Custom Rules: Add project-specific validation rules for domain-specific concerns
4. Regular Updates: Update validation patterns to reflect evolving best practices
5. Team Alignment: Ensure category weights align with team priorities and project goals
6. Consistent Application: Apply TRUST 5 validation consistently across entire codebase
7. Actionable Feedback: Provide clear, implementable suggestions for each issue detected
8. Progressive Enhancement: Start with basic validation, progressively add advanced rules

---

## Related Modules

- [Security Analysis](./security-analysis.md): Detailed security vulnerability detection
- [Quality Metrics](./quality-metrics.md): Code quality and complexity analysis
- [automated-code-review/trust5-framework.md](./automated-code-review/trust5-framework.md): Deep dive into TRUST 5 methodology

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: `modules/trust5-validation.md`
