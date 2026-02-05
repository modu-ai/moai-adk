# TRUST 5 Validation Framework

> Module: Complete TRUST 5 validation implementation with category-specific analysis and scoring
> Parent: [Automated Code Review](./automated-code-review.md)
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: Python 3.8+, ast, Context7 MCP

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

```python
from enum import Enum
from dataclasses import dataclass
from typing import Dict, List, Any

class TrustCategory(Enum):
    """TRUST 5 framework categories."""
    TRUTHFULNESS = "truthfulness"
    RELEVANCE = "relevance"
    USABILITY = "usability"
    SAFETY = "safety"
    TIMELINESS = "timeliness"

@dataclass
class CodeIssue:
    """Individual code issue found during review."""
    id: str
    category: TrustCategory
    severity: str  # critical, high, medium, low, info
    issue_type: str
    title: str
    description: str
    file_path: str
    line_number: int
    column_number: int
    code_snippet: str
    suggested_fix: str
    confidence: float  # 0.0 to 1.0
    rule_violated: str = None
    external_reference: str = None
```

---

## Implementation Guide

### Truthfulness Analysis

Truthfulness validation focuses on code correctness and logic accuracy:

```python
def _analyze_truthfulness(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
    """Analyze code for correctness and logic issues."""
    issues = []

    # Check for unreachable code
    for node in ast.walk(tree):
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            unreachable_issues = self._check_unreachable_code(file_path, node)
            issues.extend(unreachable_issues)

    # Check for logic issues
    logic_issues = self._check_logic_issues(file_path, tree)
    issues.extend(logic_issues)

    return issues
```

Unreachable Code Detection:
- Identifies code after return statements
- Detects code after raise statements
- Finds code after break/continue in loops
- Reports dead code with confidence scores

Logic Issue Detection:
- Checks for None comparison patterns (use 'is None' instead of '== None')
- Identifies constant conditions in if statements
- Detects tautological comparisons
- Finds contradictory conditions

### Relevance Analysis

Relevance analysis validates requirements fulfillment and purpose alignment:

```python
def _analyze_relevance(self, file_path: str, content: str) -> List[CodeIssue]:
    """Analyze code for relevance and requirements fulfillment."""
    issues = []

    # Check for TODO/FIXME comments
    lines = content.split('\n')
    for line_num, line in enumerate(lines, 1):
        if 'TODO:' in line or 'FIXME:' in line:
            issue = CodeIssue(
                id=f"todo_{line_num}",
                category=TrustCategory.RELEVANCE,
                severity="low",
                issue_type="documentation_issue",
                title="Unresolved TODO",
                description=f"TODO/FIXME comment found: {line.strip()}",
                file_path=file_path,
                line_number=line_num,
                column_number=line.find('TODO') if 'TODO' in line else line.find('FIXME'),
                code_snippet=line.strip(),
                suggested_fix="Adddess the TODO/FIXME item or remove the comment",
                confidence=0.6,
                rule_violated="UNRESOLVED_TODO"
            )
            issues.append(issue)

    return issues
```

Relevance Checks:
- TODO/FIXME comment tracking
- Dead code identification (unused imports, variables, functions)
- Feature completeness validation
- Documentation alignment with implementation

### Usability Analysis

Usability assessment focuses on maintainability and code quality:

```python
def _analyze_usability(self, file_path: str, content: str, tree: ast.AST) -> List[CodeIssue]:
    """Analyze code for usability and maintainability."""
    issues = []

    # Check for docstring presence
    for node in ast.walk(tree):
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            if not ast.get_docstring(node):
                issue = CodeIssue(
                    id=f"no_docstring_{node.lineno}",
                    category=TrustCategory.USABILITY,
                    severity="low",
                    issue_type="documentation_issue",
                    title="Missing Docstring",
                    description=f"Function '{node.name}' is missing a docstring",
                    file_path=file_path,
                    line_number=node.lineno,
                    column_number=1,
                    code_snippet=f"def {node.name}(...):",
                    suggested_fix=f"Add a docstring to '{node.name}' explaining its purpose, parameters, and return value",
                    confidence=0.7,
                    rule_violated="MISSING_DOCSTRING"
                )
                issues.append(issue)

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

```python
def _analyze_safety(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
    """Analyze code for safety and error handling."""
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
                    description="Bare except clause can hide unexpected errors",
                    file_path=file_path,
                    line_number=node.lineno,
                    column_number=1,
                    code_snippet="except:",
                    suggested_fix="Specify exception types or use 'except Exception:' with logging",
                    confidence=0.8,
                    rule_violated="BARE_EXCEPT"
                )
                issues.append(issue)

    return issues
```

Safety Checks:
- Bare except clause detection
- Exception handling validation
- Resource leak detection (file handles, database connections)
- Input validation review
- Context manager usage validation

### Timeliness Analysis

Timeliness assessment identifies performance and modernization opportunities:

```python
def _analyze_timeliness(self, file_path: str, content: str) -> List[CodeIssue]:
    """Analyze code for timeliness and performance."""
    issues = []

    # Check for deprecated imports
    deprecated_imports = {
        'StringIO': 'io.StringIO',
        'cStringIO': 'io.StringIO'
    }

    lines = content.split('\n')
    for line_num, line in enumerate(lines, 1):
        for old_import, new_import in deprecated_imports.items():
            if f"import {old_import}" in line or f"from {old_import}" in line:
                issue = CodeIssue(
                    id=f"deprecated_import_{line_num}",
                    category=TrustCategory.TIMELINESS,
                    severity="low",
                    issue_type="import_issue",
                    title="Deprecated Import",
                    description=f"Using deprecated import '{old_import}', should use '{new_import}'",
                    file_path=file_path,
                    line_number=line_num,
                    column_number=line.find(old_import),
                    code_snippet=line.strip(),
                    suggested_fix=f"Replace '{old_import}' with '{new_import}'",
                    confidence=0.9,
                    rule_violated="DEPRECATED_IMPORT",
                    auto_fixable=True
                )
                issues.append(issue)

    return issues
```

Timeliness Indicators:
- Deprecated import detection
- Performance anti-pattern identification
- Modern Python features adoption
- Standards compliance checking
- Technology currency validation

---

## Score Calculation

### Category Score Algorithm

```python
def _calculate_trust_scores(self, issues: List[CodeIssue], metrics: Dict[str, Any]) -> Dict[str, Any]:
    """Calculate TRUST 5 scores."""
    category_scores = {}
    category_weights = {
        TrustCategory.TRUTHFULNESS: 0.25,
        TrustCategory.RELEVANCE: 0.20,
        TrustCategory.USABILITY: 0.25,
        TrustCategory.SAFETY: 0.20,
        TrustCategory.TIMELINESS: 0.10
    }

    # Group issues by category
    issues_by_category = {category: [] for category in TrustCategory}
    for issue in issues:
        issues_by_category[issue.category].append(issue)

    # Calculate scores for each category
    for category in TrustCategory:
        category_issues = issues_by_category[category]

        # Calculate penalty based on severity and number of issues
        penalty = 0.0
        for issue in category_issues:
            severity_penalty = {
                'critical': 0.5,
                'high': 0.3,
                'medium': 0.1,
                'low': 0.05,
                'info': 0.01
            }
            penalty += severity_penalty.get(issue.severity, 0.1) * issue.confidence

        # Apply penalties (max penalty of 1.0)
        score = max(0.0, 1.0 - min(penalty, 1.0))
        category_scores[category] = score

    # Calculate overall score
    overall_score = sum(
        category_scores[cat] * category_weights[cat]
        for cat in TrustCategory
    )

    return {
        'overall': overall_score,
        'categories': category_scores
    }
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

```python
reviewer.category_weights = {
    TrustCategory.TRUTHFULNESS: 0.30,  # Increase emphasis on correctness
    TrustCategory.RELEVANCE: 0.15,
    TrustCategory.USABILITY: 0.20,
    TrustCategory.SAFETY: 0.30,        # Increase emphasis on security
    TrustCategory.TIMELINESS: 0.05
}
```

### Custom Severity Penalties

Modify penalty values for severity levels:

```python
reviewer.severity_penalties = {
    'critical': 0.7,  # Stricter penalties
    'high': 0.4,
    'medium': 0.15,
    'low': 0.05,
    'info': 0.0
}
```

### Custom Rule Configuration

Add custom validation rules:

```python
class CustomTruthfulnessAnalyzer:
    """Custom truthfulness validation rules."""

    def analyze_custom_patterns(self, file_path: str, tree: ast.AST) -> List[CodeIssue]:
        """Add project-specific truthfulness checks."""
        issues = []

        # Add custom logic validation
        # Add project-specific correctness checks
        # Add domain-specific validation rules

        return issues

# Integrate custom analyzer
reviewer.custom_analyzers[TrustCategory.TRUTHFULNESS] = CustomTruthfulnessAnalyzer()
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
