# Safety Analysis - TRUST 5 Framework

> Module: Safety category deep dive with security vulnerabilities and error handling
> Parent: [TRUST 5 Framework](./trust5-framework.md)
> Complexity: Advanced
> Time: 15+ minutes
> Dependencies: Python 3.8+, ast, Context7 MCP

## Overview

Safety (20% weight) validates security and error handling through security vulnerability detection, error handling validation, resource safety checks, and input validation analysis.

## Advanced Security Analysis

### Comprehensive Security Scan

```python
async def perform_advanced_security_analysis(
    self, file_path: str, content: str, tree: ast.AST
) -> List[CodeIssue]:
    """Perform advanced security analysis."""

    issues = []

    # Load latest security patterns from Context7
    security_patterns = await self.load_category_patterns()
    safety_patterns = security_patterns.get('safety', {})

    # Check for race conditions
    race_conditions = self._detect_race_conditions(tree)
    issues.extend(race_conditions)

    # Check for resource leaks
    resource_leaks = self._detect_resource_leaks(tree)
    issues.extend(resource_leaks)

    # Check for improper error handling
    error_handling = self._analyze_error_handling(tree)
    issues.extend(error_handling)

    # Check for input validation issues
    validation_issues = await self._check_input_validation(
        file_path, content, safety_patterns
    )
    issues.extend(validation_issues)

    return issues
```

### Resource Leak Detection

```python
def _detect_resource_leaks(self, tree: ast.AST) -> List[CodeIssue]:
    """Detect potential resource leaks."""

    issues = []

    # Check for file handles without context managers
    for node in ast.walk(tree):
        if isinstance(node, ast.Call):
            if isinstance(node.func, ast.Name) and node.func.id == 'open':
                # Check if wrapped in with statement
                if not self._is_in_with_statement(node):
                    issue = CodeIssue(
                        id=f"resource_leak_{node.lineno}",
                        category=TrustCategory.SAFETY,
                        severity="medium",
                        issue_type="code_smell",
                        title="Potential Resource Leak",
                        description="File opened without context manager",
                        file_path=file_path,
                        line_number=node.lineno,
                        column_number=node.col_offset,
                        code_snippet="# File handle may not be properly closed",
                        suggested_fix="Use 'with open(...)' to ensure proper resource cleanup",
                        confidence=0.7,
                        rule_violated="RESOURCE_LEAK"
                    )
                    issues.append(issue)

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

### Context7 Integration

```python
# Load safety patterns
safety = await self.context7.get_library_docs(
    context7_library_id="/security/owasp",
    topic="security vulnerability detection 2025",
    tokens=5000
)
```

## Error Handling Analysis

```python
def _analyze_error_handling(self, tree: ast.AST) -> List[CodeIssue]:
    """Analyze error handling patterns."""

    issues = []

    # Check for bare except clauses
    for node in ast.walk(tree):
        if isinstance(node, ast.ExceptHandler):
            if node.type is None:
                issue = CodeIssue(
                    id=f"bare_except_{node.lineno}",
                    category=TrustCategory.SAFETY,
                    severity="medium",
                    issue_type="code_smell",
                    title="Bare Except Clause",
                    description="Catches all exceptions, including system exits",
                    file_path=file_path,
                    line_number=node.lineno,
                    suggested_fix="Specify exception types to catch",
                    confidence=0.8,
                    rule_violated="BARE_EXCEPT"
                )
                issues.append(issue)

    return issues
```

## Best Practices

1. Security First: Apply security-by-default principles
2. Context Managers: Always use context managers for resources
3. Input Validation: Validate all external input
4. Error Propagation: Let exceptions propagate to appropriate handlers
5. Regular Updates: Keep security patterns current

---

Version: 1.0.0
Last Updated: 2026-01-06
