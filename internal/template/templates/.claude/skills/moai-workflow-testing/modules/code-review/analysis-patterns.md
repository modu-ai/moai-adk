# Automated Code Review - Analysis Patterns

> Sub-module: TRUST 5 analysis pattern implementations
> Parent: [Automated Code Review](../automated-code-review.md)

The patterns below operate on a parsed syntax tree (AST) plus source text. Use the host language's parser (e.g. Go go/ast, Python ast, TypeScript compiler API, Rust syn, Java JavaParser); the checks themselves generalize across trees.

## TRUST 5 Analysis Methods

### Security Pattern Analysis

```text
analyze_security_patterns(file_path, content):
    issues = []
    security_patterns = analysis_patterns['security']     # category -> list of regexes
    for (category, patterns) in security_patterns:
        for pattern in patterns:
            regex = compile(pattern, case_insensitive)
            for (line_num, line) in enumerate(lines(content), from=1):
                if regex.search(line):
                    issues.append(CodeIssue(
                        id="security_" + category + "_" + line_num,
                        category=TrustCategory.SAFETY, severity=HIGH,
                        issue_type=SECURITY_VULNERABILITY,
                        title="Security Issue: " + humanize(category),
                        description="Potential " + category + " vulnerability detected",
                        file_path=file_path, line_number=line_num, column_number=1,
                        code_snippet=trim(line),
                        suggested_fix=security_fix_suggestion(category, line),
                        confidence=0.7,
                        rule_violated="SECURITY_" + uppercase(category),
                        external_reference=security_reference(category)))
    return issues
```

### Performance Pattern Analysis

```text
analyze_performance_patterns(file_path, content):
    issues = []
    performance_patterns = analysis_patterns['performance']
    for (category, patterns) in performance_patterns:
        for pattern in patterns:
            regex = compile(pattern)
            for (line_num, line) in enumerate(lines(content), from=1):
                if regex.search(line):
                    issues.append(CodeIssue(
                        id="perf_" + category + "_" + line_num,
                        category=TrustCategory.TIMELINESS, severity=MEDIUM,
                        issue_type=PERFORMANCE_ISSUE,
                        title="Performance Issue: " + humanize(category),
                        description="Performance anti-pattern detected: " + category,
                        file_path=file_path, line_number=line_num, column_number=1,
                        code_snippet=trim(line),
                        suggested_fix=performance_fix_suggestion(category, line),
                        confidence=0.6,
                        rule_violated="PERF_" + uppercase(category)))
    return issues
```

### Quality Pattern Analysis

```text
analyze_quality_patterns(file_path, tree):
    issues = []
    q = analysis_patterns['quality']
    if 'long_functions'      in q: issues.extend(analyze_function_length(file_path, tree, q.long_functions.max_lines      default 50))
    if 'complex_conditionals'in q: issues.extend(analyze_complexity      (file_path, tree, q.complex_conditionals.max_complexity default 10))
    if 'deep_nesting'        in q: issues.extend(analyze_nesting_depth   (file_path, tree, q.deep_nesting.max_depth          default 4))
    return issues
```

## Truthfulness Analysis

### Unreachable Code Detection

```text
check_unreachable_code(file_path, func_node):
    issues = []
    # Walk the function body; once a terminator (return/throw/break/continue)
    # is seen, any subsequent statement is unreachable.
    found_terminator = false
    for node in walk(func_node.body, in_source_order):
        if node is a terminator: found_terminator = true
        if found_terminator and node is an executable statement:
            issues.append(CodeIssue(
                id="unreachable_" + node.line,
                category=TrustCategory.TRUTHFULNESS, severity=LOW,
                issue_type=CODE_SMELL, title="Unreachable Code",
                description="Code after a terminator statement is never executed",
                file_path=file_path, line_number=node.line, column_number=1,
                code_snippet="# Unreachable code at line " + node.line,
                suggested_fix="Remove unreachable code or move it before the terminator",
                confidence=0.7, rule_violated="UNREACHABLE_CODE"))
    return issues
```

### Comparison Issue Detection

```text
check_comparison_issues(file_path, compare_node):
    issues = []
    for op in compare_node.ops:
        if op is equality or inequality:
            for comparator in compare_node.comparators:
                if comparator is a null/none literal:
                    # Most languages have an identity-check idiom for null/none;
                    # prefer it over equality (Python: is None, JS: === null,
                    # Go: explicit nil check, etc.)
                    issues.append(CodeIssue(
                        id="null_comparison_" + compare_node.line,
                        category=TrustCategory.TRUTHFULNESS, severity=LOW,
                        issue_type=CODE_SMELL, title="Null Comparison Anti-Pattern",
                        description="Use the language's identity-check idiom for null/none",
                        file_path=file_path, line_number=compare_node.line,
                        code_snippet="# Use identity check, not equality, for null/none",
                        suggested_fix="Replace equality-with-null with the identity-check idiom",
                        confidence=0.8, rule_violated="NULL_COMPARISON", auto_fixable=true))
    return issues
```

## Usability Analysis

### Doc Comment Presence Check

```text
analyze_usability(file_path, content, tree):
    issues = []
    for node in walk(tree, matching=FunctionDecl):
        if not has_doc_comment(node):     # docstring / godoc / JSDoc / rustdoc / javadoc
            issues.append(CodeIssue(
                id="no_doc_" + node.line,
                category=TrustCategory.USABILITY, severity=LOW,
                issue_type=DOCUMENTATION_ISSUE, title="Missing Documentation",
                description="Function '" + node.name + "' has no doc comment",
                file_path=file_path, line_number=node.line, column_number=1,
                code_snippet=signature(node),
                suggested_fix="Add a doc comment explaining purpose and parameters",
                confidence=0.7, rule_violated="MISSING_DOC"))
    return issues
```

## Safety Analysis

### Bare / Catch-All Handler Detection

```text
analyze_safety(file_path, tree):
    issues = []
    for node in walk(tree, matching=CatchClause):
        if catches_all_untyped(node):    # bare catch / catch-all without a typed error
            issues.append(CodeIssue(
                id="bare_catch_" + node.line,
                category=TrustCategory.SAFETY, severity=MEDIUM,
                issue_type=CODE_SMELL, title="Bare / Catch-All Handler",
                description="A catch-all handler can hide unexpected errors",
                file_path=file_path, line_number=node.line, column_number=1,
                code_snippet="catch (untyped)",
                suggested_fix="Catch specific error types, or catch the base error type with logging",
                confidence=0.8, rule_violated="BARE_CATCH"))
    return issues
```

## Complexity Metrics

### Cyclomatic Complexity

```text
calculate_cyclomatic_complexity(node):
    complexity = 1   # base
    for child in walk(node):
        if child is a decision point (if, while, for, catch, with/using, case):
            complexity += 1
        if child is a boolean operator (and/or) with N operands:
            complexity += (N - 1)
    return complexity
```

### Nesting Depth

```text
calculate_nesting_depth(node, current_depth = 0):
    max_depth = current_depth
    for child in nested_blocks_of(node):     # if/while/for/with/try bodies, etc.
        max_depth = max(max_depth, calculate_nesting_depth(child, current_depth + 1))
    return max_depth
```

## Fix Suggestion Helpers

```text
security_fix_suggestion(category, line):
    suggestions = {
        sql_injection:    "Use parameterized queries or an ORM",
        command_injection:"Pass arguments as a list, not a shell string",
        path_traversal:   "Validate and sanitize file paths; constrain to allowed roots"
    }
    return suggestions[category] default "Review and fix the security vulnerability"

performance_fix_suggestion(category, line):
    suggestions = {
        inefficient_loops: "Replace nested loops with a vectorized / map-filter idiom",
        memory_leaks:      "Review allocations and ensure cleanup (defer/using/RAII)"
    }
    return suggestions[category] default "Optimize for better performance"

security_reference(category):
    references = {
        sql_injection:     "OWASP SQL Injection Prevention",
        command_injection: "OWASP Command Injection Prevention",
        path_traversal:    "OWASP Path Traversal Prevention"
    }
    return references[category] default "OWASP Top 10 Security Risks"
```

## Related Sub-modules

- [Core Classes](./core-classes.md) - Data structures and main classes
- [Tool Integration](./tool-integration.md) - Static analysis tools

---

Sub-module: `modules/code-review/analysis-patterns.md`
