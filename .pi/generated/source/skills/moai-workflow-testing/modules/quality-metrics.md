# Quality Metrics and Complexity Analysis

> Module: Code quality assessment, complexity calculation, and maintainability metrics
> Parent: [Automated Code Review](./automated-code-review.md)
> Complexity: Intermediate
> Time: 15+ minutes
> Dependencies: Python 3.8+, ast, radon, mccabe

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
- Test Coverage: Test completeness (requires pytest-cov)

Code Smell Detection:
- Long Methods: Functions exceeding length thresholds
- God Classes: Classes with too many responsibilities
- Feature Envy: Methods that use other classes more
- Data Clumps: Group of data items that appear together
- Primitive Obsession: Overuse of primitive types

### Core Implementation

```python
import ast
from typing import Dict, List, Any

class QualityMetricsAnalyzer:
    """Code quality and complexity analyzer."""

    def __init__(self):
        self.complexity_thresholds = {
            'cyclomatic': 10,
            'cognitive': 15,
            'nesting_depth': 4,
            'function_length': 50,
            'parameter_count': 7
        }

    def analyze_file_quality(
        self, file_path: str, content: str, tree: ast.AST
    ) -> Dict[str, Any]:
        """Analyze comprehensive file quality metrics."""

        metrics = {
            'complexity': self._calculate_complexity_metrics(tree),
            'maintainability': self._calculate_maintainability_metrics(content, tree),
            'code_smells': self._detect_code_smells(file_path, content, tree),
            'documentation': self._analyze_documentation(tree),
            'statistics': self._calculate_file_statistics(content, tree)
        }

        return metrics
```

---

## Implementation Guide

### Cyclomatic Complexity

```python
def calculate_cyclomatic_complexity(self, node: ast.AST) -> int:
    """Calculate cyclomatic complexity for an AST node."""

    complexity = 1  # Base complexity

    for child in ast.walk(node):
        if isinstance(child, (
            ast.If, ast.While, ast.For, ast.AsyncFor,
            ast.ExceptHandler, ast.With, ast.AsyncWith
        )):
            complexity += 1
        elif isinstance(child, ast.BoolOp):
            complexity += len(child.values) - 1
        elif isinstance(child, ast.Match):  # Python 3.10+
            complexity += len(child.cases)

    return complexity

def analyze_function_complexity(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
    """Analyze function complexity violations."""

    issues = []

    for node in ast.walk(tree):
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            complexity = self.calculate_cyclomatic_complexity(node)

            if complexity > self.complexity_thresholds['cyclomatic']:
                severity = "high" if complexity > 20 else "medium"

                issue = CodeIssue(
                    id=f"complexity_{node.lineno}",
                    category=TrustCategory.USABILITY,
                    severity=severity,
                    issue_type="code_smell",
                    title="High Cyclomatic Complexity",
                    description=f"Function '{node.name}' has cyclomatic complexity {complexity}",
                    file_path=file_path,
                    line_number=node.lineno,
                    column_number=1,
                    code_snippet=f"def {node.name}(...): # complexity: {complexity}",
                    suggested_fix=f"Consider refactoring '{node.name}' to reduce complexity",
                    confidence=0.9,
                    rule_violated="COMPLEXITY"
                )
                issues.append(issue)

    return issues
```

Cyclomatic Complexity Interpretation:
1-10: Simple, low risk
11-20: Moderate complexity, medium risk
21-50: High complexity, high risk
51+: Very high complexity, very high risk

### Nesting Depth Analysis

```python
def calculate_nesting_depth(self, node: ast.AST, current_depth: int = 0) -> int:
    """Calculate maximum nesting depth for an AST node."""

    max_depth = current_depth

    for child in ast.walk(node):
        if isinstance(child, (
            ast.If, ast.While, ast.For, ast.AsyncFor,
            ast.With, ast.AsyncWith, ast.Try
        )):
            if hasattr(child, 'lineno') and hasattr(node, 'lineno'):
                if child.lineno > node.lineno:
                    child_depth = self.calculate_nesting_depth(child, current_depth + 1)
                    max_depth = max(max_depth, child_depth)

    return max_depth

def analyze_nesting_depth(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
    """Analyze nesting depth violations."""

    issues = []

    for node in ast.walk(tree):
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            max_depth = self.calculate_nesting_depth(node)

            if max_depth > self.complexity_thresholds['nesting_depth']:
                issue = CodeIssue(
                    id=f"nesting_{node.lineno}",
                    category=TrustCategory.USABILITY,
                    severity="medium",
                    issue_type="code_smell",
                    title="Deep Nesting",
                    description=f"Function '{node.name}' has nesting depth {max_depth}",
                    file_path=file_path,
                    line_number=node.lineno,
                    column_number=1,
                    code_snippet=f"def {node.name}(...): # nesting depth: {max_depth}",
                    suggested_fix="Use early returns or extract methods to reduce nesting",
                    confidence=0.8,
                    rule_violated="NESTING_DEPTH"
                )
                issues.append(issue)

    return issues
```

### Function Length Analysis

```python
def analyze_function_length(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
    """Analyze function length violations."""

    issues = []
    lines = None
    max_lines = self.complexity_thresholds['function_length']

    for node in ast.walk(tree):
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            if lines is None:
                with open(file_path, 'r') as f:
                    lines = f.readlines()

            # Calculate function length (excluding docstring and blanks)
            start_line = node.lineno - 1
            end_line = node.end_lineno - 1 if node.end_lineno else start_line
            func_lines = lines[start_line:end_line + 1]

            # Remove docstring and blank lines
            code_lines = []
            in_docstring = False
            for line in func_lines:
                stripped = line.strip()
                if not in_docstring and ('"""' in line or "'''" in line):
                    in_docstring = True
                    continue
                if in_docstring and ('"""' in line or "'''" in line):
                    in_docstring = False
                    continue
                if not in_docstring and stripped and not stripped.startswith('#'):
                    code_lines.append(line)

            if len(code_lines) > max_lines:
                issue = CodeIssue(
                    id=f"func_length_{node.lineno}",
                    category=TrustCategory.USABILITY,
                    severity="medium",
                    issue_type="code_smell",
                    title="Long Function",
                    description=f"Function '{node.name}' is {len(code_lines)} lines long",
                    file_path=file_path,
                    line_number=node.lineno,
                    column_number=1,
                    code_snippet=f"def {node.name}(...): # {len(code_lines)} lines",
                    suggested_fix=f"Break '{node.name}' into smaller, focused functions",
                    confidence=0.8,
                    rule_violated="FUNC_LENGTH"
                )
                issues.append(issue)

    return issues
```

### Parameter Count Analysis

```python
def analyze_parameter_count(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
    """Analyze parameter count violations."""

    issues = []

    for node in ast.walk(tree):
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            param_count = len(node.args.args)
            max_params = self.complexity_thresholds['parameter_count']

            if param_count > max_params:
                issue = CodeIssue(
                    id=f"param_count_{node.lineno}",
                    category=TrustCategory.USABILITY,
                    severity="medium",
                    issue_type="code_smell",
                    title="Too Many Parameters",
                    description=f"Function '{node.name}' has {param_count} parameters (max: {max_params})",
                    file_path=file_path,
                    line_number=node.lineno,
                    column_number=1,
                    code_snippet=f"def {node.name}({', '.join([arg.arg for arg in node.args.args[:3]])}, ...):",
                    suggested_fix=f"Consider using a data class or configuration object for '{node.name}'",
                    confidence=0.7,
                    rule_violated="PARAMETER_COUNT"
                )
                issues.append(issue)

    return issues
```

### Maintainability Index

```python
def calculate_maintainability_index(
    self, content: str, tree: ast.AST
) -> Dict[str, float]:
    """Calculate maintainability index (MI)."""

    # Calculate Halstead volume
    volume = self._calculate_halstead_volume(content, tree)

    # Calculate cyclomatic complexity
    complexity = self._calculate_total_complexity(tree)

    # Calculate lines of code
    lines_of_code = len([line for line in content.split('\n') if line.strip()])

    # Calculate comment ratio
    comment_lines = len([line for line in content.split('\n') if line.strip().startswith('#')])
    comment_ratio = comment_lines / max(lines_of_code, 1)

    # Maintainability Index formula (MI)
    # MI = 171 - 5.2 * ln(Halstead Volume) - 0.23 * Cyclomatic Complexity - 16.2 * ln(Lines of Code)
    # MI = MI * (100 / (171 - 5.2 * ln(126) - 0.23 * 50 - 16.2 * ln(540)))  # Normalize

    import math

    mi = 171 - (5.2 * math.log(max(volume, 1))) - (0.23 * complexity) - (16.2 * math.log(max(lines_of_code, 1)))

    # Adjust for comment ratio
    mi = mi * (1 + comment_ratio)

    # Normalize to 0-100 scale
    mi_normalized = max(0, min(100, mi))

    return {
        'mi_score': mi_normalized,
        'mi_rating': self._get_mi_rating(mi_normalized),
        'halstead_volume': volume,
        'cyclomatic_complexity': complexity,
        'lines_of_code': lines_of_code,
        'comment_ratio': comment_ratio
    }

def _get_mi_rating(self, mi_score: float) -> str:
    """Get maintainability rating from MI score."""
    if mi_score >= 85:
        return "Excellent"
    elif mi_score >= 70:
        return "Good"
    elif mi_score >= 55:
        return "Moderate"
    elif mi_score >= 40:
        return "Poor"
    else:
        return "Very Poor"
```

### Code Smell Detection

```python
def detect_code_smells(
    self, file_path: str, content: str, tree: ast.AST
) -> List[CodeIssue]:
    """Detect various code smells."""

    smells = []

    # Long methods
    smells.extend(self.analyze_function_length(file_path, tree))

    # High complexity
    smells.extend(self.analyze_function_complexity(file_path, tree))

    # Deep nesting
    smells.extend(self.analyze_nesting_depth(file_path, tree))

    # Too many parameters
    smells.extend(self.analyze_parameter_count(file_path, tree))

    # God classes (too many methods)
    for node in ast.walk(tree):
        if isinstance(node, ast.ClassDef):
            methods = [n for n in node.body if isinstance(n, (ast.FunctionDef, ast.AsyncFunctionDef))]
            if len(methods) > 20:
                issue = CodeIssue(
                    id=f"god_class_{node.lineno}",
                    category=TrustCategory.USABILITY,
                    severity="medium",
                    issue_type="code_smell",
                    title="God Class",
                    description=f"Class '{node.name}' has {len(methods)} methods (max: 20)",
                    file_path=file_path,
                    line_number=node.lineno,
                    column_number=1,
                    code_snippet=f"class {node.name}:  # {len(methods)} methods",
                    suggested_fix=f"Consider splitting '{node.name}' into smaller, focused classes",
                    confidence=0.7,
                    rule_violated="GOD_CLASS"
                )
                smells.append(issue)

    return smells
```

---

## Custom Thresholds

```python
# Customize complexity thresholds
analyzer.complexity_thresholds = {
    'cyclomatic': 15,        # More lenient
    'cognitive': 20,
    'nesting_depth': 5,
    'function_length': 75,   # Allow longer functions
    'parameter_count': 10    # Allow more parameters
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
