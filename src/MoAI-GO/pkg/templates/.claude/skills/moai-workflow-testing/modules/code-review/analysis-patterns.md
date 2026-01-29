# Automated Code Review - Analysis Patterns

> Sub-module: TRUST 5 analysis pattern implementations
> Parent: [Automated Code Review](../automated-code-review.md)

## TRUST 5 Analysis Methods

### Security Pattern Analysis

```python
async def _analyze_security_patterns(
    self, file_path: str, content: str
) -> List[CodeIssue]:
    """Analyze security patterns using Context7."""
    issues = []
    security_patterns = self.analysis_patterns.get('security', {})
    lines = content.split('\n')

    for category, patterns in security_patterns.items():
        if isinstance(patterns, list):
            for pattern in patterns:
                try:
                    regex = re.compile(pattern, re.IGNORECASE)
                    for line_num, line in enumerate(lines, 1):
                        if regex.search(line):
                            issue = CodeIssue(
                                id=f"security_{category}_{line_num}_{len(issues)}",
                                category=TrustCategory.SAFETY,
                                severity=Severity.HIGH,
                                issue_type=IssueType.SECURITY_VULNERABILITY,
                                title=f"Security Issue: {category.replace('_', ' ').title()}",
                                description=f"Potential {category} vulnerability detected",
                                file_path=file_path,
                                line_number=line_num,
                                column_number=1,
                                code_snippet=line.strip(),
                                suggested_fix=self._get_security_fix_suggestion(category, line),
                                confidence=0.7,
                                rule_violated=f"SECURITY_{category.upper()}",
                                external_reference=self._get_security_reference(category)
                            )
                            issues.append(issue)
                except re.error as e:
                    print(f"Invalid security pattern {pattern}: {e}")

    return issues
```

### Performance Pattern Analysis

```python
async def _analyze_performance_patterns(
    self, file_path: str, content: str
) -> List[CodeIssue]:
    """Analyze performance patterns using Context7."""
    issues = []
    performance_patterns = self.analysis_patterns.get('performance', {})
    lines = content.split('\n')

    for category, patterns in performance_patterns.items():
        if isinstance(patterns, list):
            for pattern in patterns:
                try:
                    regex = re.compile(pattern)
                    for line_num, line in enumerate(lines, 1):
                        if regex.search(line):
                            issue = CodeIssue(
                                id=f"perf_{category}_{line_num}_{len(issues)}",
                                category=TrustCategory.TIMELINESS,
                                severity=Severity.MEDIUM,
                                issue_type=IssueType.PERFORMANCE_ISSUE,
                                title=f"Performance Issue: {category.replace('_', ' ').title()}",
                                description=f"Performance anti-pattern detected: {category}",
                                file_path=file_path,
                                line_number=line_num,
                                column_number=1,
                                code_snippet=line.strip(),
                                suggested_fix=self._get_performance_fix_suggestion(category, line),
                                confidence=0.6,
                                rule_violated=f"PERF_{category.upper()}"
                            )
                            issues.append(issue)
                except re.error as e:
                    print(f"Invalid performance pattern {pattern}: {e}")

    return issues
```

### Quality Pattern Analysis

```python
async def _analyze_quality_patterns(
    self, file_path: str, tree: ast.AST
) -> List[CodeIssue]:
    """Analyze code quality patterns."""
    issues = []
    quality_patterns = self.analysis_patterns.get('quality', {})

    # Analyze function length
    if 'long_functions' in quality_patterns:
        max_lines = quality_patterns['long_functions'].get('max_lines', 50)
        function_issues = self._analyze_function_length(file_path, tree, max_lines)
        issues.extend(function_issues)

    # Analyze complexity
    if 'complex_conditionals' in quality_patterns:
        max_complexity = quality_patterns['complex_conditionals'].get('max_complexity', 10)
        complexity_issues = self._analyze_complexity(file_path, tree, max_complexity)
        issues.extend(complexity_issues)

    # Analyze nesting depth
    if 'deep_nesting' in quality_patterns:
        max_depth = quality_patterns['deep_nesting'].get('max_depth', 4)
        nesting_issues = self._analyze_nesting_depth(file_path, tree, max_depth)
        issues.extend(nesting_issues)

    return issues
```

## Truthfulness Analysis

### Unreachable Code Detection

```python
def _check_unreachable_code(
    self, file_path: str, func_node: ast.AST
) -> List[CodeIssue]:
    """Check for unreachable code after return statements."""
    issues = []

    class UnreachableCodeVisitor(ast.NodeVisitor):
        def __init__(self):
            self.found_return = False
            self.issues = []

        def visit_Return(self, node):
            self.found_return = True
            self.generic_visit(node)

        def generic_visit(self, node):
            if self.found_return and hasattr(node, 'lineno'):
                if isinstance(node, (ast.Expr, ast.Assign, ast.AugAssign)):
                    issue = CodeIssue(
                        id=f"unreachable_{node.lineno}",
                        category=TrustCategory.TRUTHFULNESS,
                        severity=Severity.LOW,
                        issue_type=IssueType.CODE_SMELL,
                        title="Unreachable Code",
                        description="Code after return statement is never executed",
                        file_path=file_path,
                        line_number=node.lineno,
                        column_number=1,
                        code_snippet=f"# Unreachable code at line {node.lineno}",
                        suggested_fix="Remove unreachable code or move before return",
                        confidence=0.7,
                        rule_violated="UNREACHABLE_CODE"
                    )
                    self.issues.append(issue)
            super().generic_visit(node)

    visitor = UnreachableCodeVisitor()
    visitor.visit(func_node)
    return visitor.issues
```

### Comparison Issue Detection

```python
def _check_comparison_issues(
    self, file_path: str, compare_node: ast.Compare
) -> List[CodeIssue]:
    """Check for comparison logic issues."""
    issues = []

    for op in compare_node.ops:
        if isinstance(op, (ast.Eq, ast.NotEq)):
            for comparator in compare_node.comparators:
                if isinstance(comparator, ast.Constant) and comparator.value is None:
                    issue = CodeIssue(
                        id=f"none_comparison_{compare_node.lineno}",
                        category=TrustCategory.TRUTHFULNESS,
                        severity=Severity.LOW,
                        issue_type=IssueType.CODE_SMELL,
                        title="None Comparison",
                        description="Use 'is' or 'is not' for None comparison",
                        file_path=file_path,
                        line_number=compare_node.lineno,
                        column_number=1,
                        code_snippet="# Use 'is None' instead of '== None'",
                        suggested_fix="Replace '== None' with 'is None'",
                        confidence=0.8,
                        rule_violated="NONE_COMPARISON",
                        auto_fixable=True
                    )
                    issues.append(issue)

    return issues
```

## Usability Analysis

### Docstring Presence Check

```python
def _analyze_usability(
    self, file_path: str, content: str, tree: ast.AST
) -> List[CodeIssue]:
    """Analyze code for usability and maintainability."""
    issues = []

    for node in ast.walk(tree):
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            if not ast.get_docstring(node):
                issue = CodeIssue(
                    id=f"no_docstring_{node.lineno}",
                    category=TrustCategory.USABILITY,
                    severity=Severity.LOW,
                    issue_type=IssueType.DOCUMENTATION_ISSUE,
                    title="Missing Docstring",
                    description=f"Function '{node.name}' is missing a docstring",
                    file_path=file_path,
                    line_number=node.lineno,
                    column_number=1,
                    code_snippet=f"def {node.name}(...):",
                    suggested_fix=f"Add docstring explaining purpose and parameters",
                    confidence=0.7,
                    rule_violated="MISSING_DOCSTRING"
                )
                issues.append(issue)

    return issues
```

## Safety Analysis

### Bare Except Detection

```python
def _analyze_safety(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
    """Analyze code for safety and error handling."""
    issues = []

    for node in ast.walk(tree):
        if isinstance(node, ast.ExceptHandler):
            if node.type is None:
                issue = CodeIssue(
                    id=f"bare_except_{node.lineno}",
                    category=TrustCategory.SAFETY,
                    severity=Severity.MEDIUM,
                    issue_type=IssueType.CODE_SMELL,
                    title="Bare Except Clause",
                    description="Bare except clause can hide unexpected errors",
                    file_path=file_path,
                    line_number=node.lineno,
                    column_number=1,
                    code_snippet="except:",
                    suggested_fix="Specify exception types or use 'except Exception:'",
                    confidence=0.8,
                    rule_violated="BARE_EXCEPT"
                )
                issues.append(issue)

    return issues
```

## Complexity Metrics

### Cyclomatic Complexity

```python
def _calculate_cyclomatic_complexity(self, node: ast.AST) -> int:
    """Calculate cyclomatic complexity for an AST node."""
    complexity = 1  # Base complexity

    for child in ast.walk(node):
        if isinstance(child, (ast.If, ast.While, ast.For, ast.AsyncFor,
                              ast.ExceptHandler, ast.With, ast.AsyncWith)):
            complexity += 1
        elif isinstance(child, ast.BoolOp):
            complexity += len(child.values) - 1

    return complexity
```

### Nesting Depth

```python
def _calculate_nesting_depth(self, node: ast.AST, current_depth: int = 0) -> int:
    """Calculate maximum nesting depth for an AST node."""
    max_depth = current_depth

    for child in ast.walk(node):
        if isinstance(child, (ast.If, ast.While, ast.For, ast.AsyncFor,
                              ast.With, ast.AsyncWith, ast.Try)):
            if hasattr(child, 'lineno') and hasattr(node, 'lineno'):
                if child.lineno > node.lineno:
                    child_depth = self._calculate_nesting_depth(child, current_depth + 1)
                    max_depth = max(max_depth, child_depth)

    return max_depth
```

## Fix Suggestion Helpers

```python
def _get_security_fix_suggestion(self, category: str, line: str) -> str:
    """Get security fix suggestion."""
    suggestions = {
        'sql_injection': "Use parameterized queries or ORM",
        'command_injection': "Use subprocess.run with argument lists",
        'path_traversal': "Validate and sanitize file paths"
    }
    return suggestions.get(category, "Review and fix security vulnerability")

def _get_performance_fix_suggestion(self, category: str, line: str) -> str:
    """Get performance fix suggestion."""
    suggestions = {
        'inefficient_loops': "Use list comprehensions or generators",
        'memory_leaks': "Review memory usage and ensure cleanup"
    }
    return suggestions.get(category, "Optimize for better performance")

def _get_security_reference(self, category: str) -> str:
    """Get external security reference."""
    references = {
        'sql_injection': "OWASP SQL Injection Prevention",
        'command_injection': "OWASP Command Injection Prevention",
        'path_traversal': "OWASP Path Traversal Prevention"
    }
    return references.get(category, "OWASP Top 10 Security Risks")
```

## Related Sub-modules

- [Core Classes](./core-classes.md) - Data structures and main classes
- [Tool Integration](./tool-integration.md) - Static analysis tools

---

Sub-module: `modules/code-review/analysis-patterns.md`
