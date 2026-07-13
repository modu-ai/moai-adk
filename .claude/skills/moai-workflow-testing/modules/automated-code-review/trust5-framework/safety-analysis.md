# Safety Analysis - TRUST 5 Framework

> Module: Safety category deep dive with security vulnerabilities and error handling
> Parent: [TRUST 5 Framework](./trust5-framework.md)
> Complexity: Advanced
> Time: 15+ minutes
> Dependencies: source parser (AST), WebSearch/WebFetch

## Overview

Safety (20% weight) validates security and error handling through security vulnerability detection, error handling validation, resource safety checks, and input validation analysis.

## Advanced Security Analysis

### Comprehensive Security Scan

```text
perform_advanced_security_analysis(file_path, content, tree):
    issues = []
    # Load latest security patterns from Documentation
    security_patterns = load_category_patterns()
    safety_patterns   = security_patterns.safety default {}

    issues.extend(detect_race_conditions(tree))
    issues.extend(detect_resource_leaks(tree))
    issues.extend(analyze_error_handling(tree))
    issues.extend(check_input_validation(file_path, content, safety_patterns))
    return issues
```

### Resource Leak Detection

Detect resources (files, connections, handles) acquired without a cleanup idiom. The exact idiom is language-specific (Python `with`, Go `defer`, JS/TS `using`, Java/JS try-with-resources, Rust drop/RAII); flag acquisitions that are not covered by one.

```text
detect_resource_leaks(tree):
    issues = []
    for node in walk(tree, matching=ResourceAcquisition):   # open(), connect(), acquire()
        if not is_within_cleanup_block(node):               # not inside with/defer/using/try-finally
            issues.append(CodeIssue(
                id="resource_leak_" + node.line,
                category=TrustCategory.SAFETY, severity="medium",
                issue_type="code_smell", title="Potential Resource Leak",
                description="Resource acquired without a cleanup idiom (defer / with / using / RAII)",
                file_path=file_path, line_number=node.line, column_number=node.column,
                code_snippet="# resource may not be released on all paths",
                suggested_fix="Wrap the acquisition in the language's cleanup idiom",
                confidence=0.7, rule_violated="RESOURCE_LEAK"))
    return issues
```

## Detection Patterns

### Common Safety Issues

1. **SQL Injection**: Unsanitized input in database queries
2. **XSS Vulnerabilities**: Unescaped output in web contexts
3. **Resource Leaks**: Unclosed files, connections, handles
4. **Race Conditions**: Concurrent access without synchronization
5. **Missing Error Handling**: Uncaught exceptions
6. **Weak Cryptography**: Insecure algorithms or key sizes
7. **Hardcoded Secrets**: Passwords, API keys in source

### Documentation Integration

```text
# Load safety patterns
safety = docs.get_library_docs(
    "<security/owasp>",
    topic="security vulnerability detection",
    tokens=5000)
```

## Error Handling Analysis

```text
analyze_error_handling(tree):
    issues = []
    for node in walk(tree, matching=CatchClause):
        if catches_all_untyped(node):    # bare catch / catch-all without a typed error
            issues.append(CodeIssue(
                id="bare_catch_" + node.line,
                category=TrustCategory.SAFETY, severity="medium",
                issue_type="code_smell", title="Bare / Catch-All Handler",
                description="Catches all errors, including system-exit and unrelated failures",
                file_path=file_path, line_number=node.line,
                suggested_fix="Catch specific error types, or catch the base error type with logging",
                confidence=0.8, rule_violated="BARE_CATCH"))
    return issues
```

## Best Practices

1. Security First: Apply security-by-default principles
2. Cleanup Idioms: Always use the language's cleanup idiom (defer / with / using / RAII) for resources
3. Input Validation: Validate all external input
4. Error Propagation: Let errors propagate to appropriate handlers
5. Regular Updates: Keep security patterns current

---

Version: 1.0.0
Last Updated: 2026-01-06
