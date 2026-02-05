# Relevance Analysis - TRUST 5 Framework

> Module: Relevance category deep dive with requirements traceability and dead code detection
> Parent: [TRUST 5 Framework](./trust5-framework.md)
> Complexity: Advanced
> Time: 10+ minutes
> Dependencies: Python 3.8+, ast, requirements mapping

## Overview

Relevance (20% weight) validates requirements fulfillment and purpose alignment through feature completeness checks, requirements traceability, dead code identification, and purpose alignment validation.

## Requirements Traceability

### Missing Requirement Detection

```python
def validate_requirements_traceability(
    self, file_path: str, content: str, requirements: List[str]
) -> List[CodeIssue]:
    """Validate requirements traceability in code."""

    issues = []

    # Check for unimplemented requirements
    for req_id, requirement in requirements:
        if req_id not in content:
            issue = CodeIssue(
                id=f"missing_requirement_{req_id}",
                category=TrustCategory.RELEVANCE,
                severity="medium",
                issue_type="documentation_issue",
                title="Missing Requirement Implementation",
                description=f"Requirement {req_id} not found in code",
                file_path=file_path,
                line_number=1,
                column_number=1,
                code_snippet=f"# TODO: Implement {req_id}: {requirement}",
                suggested_fix=f"Implement requirement {req_id}: {requirement}",
                confidence=0.8,
                rule_violated="MISSING_REQUIREMENT"
            )
            issues.append(issue)

    return issues
```

## Dead Code Detection

### Unused Function Detection

```python
def detect_dead_code(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
    """Detect dead code (unused functions, classes, imports)."""

    issues = []

    # Find all defined functions
    defined_functions = set()
    for node in ast.walk(tree):
        if isinstance(node, ast.FunctionDef):
            defined_functions.add(node.name)

    # Find all called functions
    called_functions = set()
    for node in ast.walk(tree):
        if isinstance(node, ast.Call):
            if isinstance(node.func, ast.Name):
                called_functions.add(node.func.id)

    # Identify unused functions (excluding main, test functions)
    unused_functions = defined_functions - called_functions
    for func_name in unused_functions:
        if not func_name.startswith('_') and func_name not in ['main', 'test']:
            # Find the function definition node
            for node in ast.walk(tree):
                if isinstance(node, ast.FunctionDef) and node.name == func_name:
                    issue = CodeIssue(
                        id=f"dead_code_{node.lineno}",
                        category=TrustCategory.RELEVANCE,
                        severity="low",
                        issue_type="code_smell",
                        title="Dead Code",
                        description=f"Function '{func_name}' is never called",
                        file_path=file_path,
                        line_number=node.lineno,
                        column_number=1,
                        code_snippet=f"def {func_name}(...):",
                        suggested_fix=f"Remove unused function '{func_name}' or update references",
                        confidence=0.6,
                        rule_violated="DEAD_CODE"
                    )
                    issues.append(issue)
                    break

    return issues
```

## Detection Patterns

### Common Relevance Issues

1. **Missing Requirements**: Specified features not implemented
2. **Unused Functions**: Functions defined but never called
3. **Unused Imports**: Imported modules not referenced
4. **Dead Classes**: Classes defined but never instantiated
5. **Commented Code**: Large blocks of commented-out code
6. **TODO/FIXME**: Unresolved development markers
7. **Feature Creep**: Code beyond original requirements

### Purpose Alignment Analysis

```python
def check_purpose_alignment(
    self, file_path: str, content: str, purpose: str
) -> List[CodeIssue]:
    """Check if code aligns with stated purpose."""

    issues = []

    # Analyze code complexity vs purpose
    complexity = self._calculate_complexity(content)

    # Check for over-engineering
    if complexity > self._get_expected_complexity(purpose):
        issue = CodeIssue(
            id="over_engineered",
            category=TrustCategory.RELEVANCE,
            severity="low",
            issue_type="code_smell",
            title="Over-Engineered Solution",
            description=f"Code complexity exceeds purpose requirements",
            file_path=file_path,
            suggested_fix="Simplify implementation to match purpose",
            confidence=0.7,
            rule_violated="PURPOSE_MISALIGNMENT"
        )
        issues.append(issue)

    return issues
```

## Best Practices

1. Requirements Mapping: Maintain explicit requirement-to-code traceability
2. Regular Cleanup: Remove dead code during refactoring
3. Purpose Documentation: Document module/class/function purposes
4. Complexity Monitoring: Monitor complexity growth vs requirements
5. Impact Analysis: Analyze impact before removing dead code

---

Version: 1.0.0
Last Updated: 2026-01-06
