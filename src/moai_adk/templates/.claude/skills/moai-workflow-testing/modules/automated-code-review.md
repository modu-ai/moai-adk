# Automated Code Review with TRUST 5 Validation

> Module: AI-powered automated code review with TRUST 5 validation framework
> Complexity: Advanced
> Time: 35+ minutes
> Dependencies: Python 3.8+, Context7 MCP, ast, pylint, flake8, bandit, mypy

## Overview

This module provides comprehensive automated code review capabilities using the TRUST 5 framework:
- Truthfulness: Code correctness and logic accuracy
- Relevance: Code meets requirements and purpose
- Usability: Code is maintainable and understandable
- Safety: Code is secure and handles errors properly
- Timeliness: Code meets performance and delivery standards

## Quick Reference

### Core Components

AutomatedCodeReviewer: Main class orchestrating code review workflow
- Integrates multiple static analysis tools
- Uses Context7 for pattern detection
- Generates TRUST 5 scores

Context7CodeAnalyzer: Integration with Context7 for analysis patterns
- Security vulnerability detection
- Performance anti-pattern identification
- Code quality assessment

StaticAnalysisTools: Wrapper for external analyzers
- pylint for code quality
- flake8 for style checking
- bandit for security scanning
- mypy for type checking

### Data Structures

CodeIssue: Individual code issue with metadata
- category, severity, issue_type
- file_path, line_number, code_snippet
- suggested_fix, confidence, auto_fixable

FileReviewResult: Review results for a single file
- issues list, metrics, trust_score
- category_scores, complexity_metrics

CodeReviewReport: Comprehensive project review
- files_reviewed, overall_trust_score
- recommendations, critical_issues

### Basic Usage

```python
# Initialize automated code reviewer
reviewer = AutomatedCodeReviewer(context7_client=context7)

# Review entire codebase
report = await reviewer.review_codebase(
    project_path="/path/to/project",
    include_patterns=["**/*.py"],
    exclude_patterns=["/tests/", "/__pycache__/"]
)

print(f"Overall TRUST Score: {report.overall_trust_score:.2f}")
print(f"Critical Issues: {report.summary_metrics['critical_issues']}")
```

## Implementation Guide

For detailed implementation patterns, see:
- [Core Classes](./code-review/core-classes.md) - Main class implementations
- [Analysis Patterns](./code-review/analysis-patterns.md) - TRUST 5 analysis methods
- [Tool Integration](./code-review/tool-integration.md) - Static analysis tool wrappers

### TRUST 5 Category Analysis

Each category contributes to the overall code quality score:

Truthfulness Analysis:
- Unreachable code detection
- Logic issue identification
- Comparison pattern validation

Relevance Analysis:
- TODO/FIXME comment tracking
- Requirements fulfillment verification

Usability Analysis:
- Docstring presence checking
- Function length validation
- Complexity metrics calculation

Safety Analysis:
- Bare except clause detection
- Security vulnerability scanning
- Error handling verification

Timeliness Analysis:
- Deprecated import detection
- Performance anti-pattern identification

### Score Calculation

Category weights for overall TRUST score:
- Truthfulness: 25%
- Relevance: 20%
- Usability: 25%
- Safety: 20%
- Timeliness: 10%

Severity penalties applied per issue:
- Critical: -0.5 per issue
- High: -0.3 per issue
- Medium: -0.1 per issue
- Low: -0.05 per issue
- Info: -0.01 per issue

## Advanced Features

### Context7-Enhanced Security Analysis

Load latest security patterns from Context7:
```python
security_patterns = await context7.get_library_docs(
    context7_library_id="/security/semgrep",
    topic="security vulnerability detection patterns 2025",
    tokens=4000
)
```

### Custom Pattern Detection

Define custom patterns for project-specific issues:
```python
custom_patterns = {
    'custom_security': [r"eval\(", r"exec\("],
    'custom_performance': [r"for.*in.*range\(len\("]
}
```

### Report Generation

Generate actionable recommendations:
```python
recommendations = report.recommendations
for rec in recommendations:
    print(f"- {rec}")
```

## Best Practices

1. Comprehensive Coverage: Analyze code across all TRUST 5 dimensions
2. Context Integration: Leverage Context7 for up-to-date security patterns
3. Actionable Feedback: Provide specific, implementable suggestions
4. Severity Prioritization: Focus on critical and high-severity issues first
5. Continuous Integration: Integrate into CI/CD pipeline for automated reviews

## Related Modules

- [Smart Refactoring](./smart-refactoring.md) - Refactoring recommendations
- [Performance Optimization](./performance-optimization.md) - Performance analysis
- [AI Debugging](./ai-debugging.md) - Error analysis integration

---

Module: `modules/automated-code-review.md`
Version: 2.0.0 (Modular Architecture)
