# Automated Code Review with TRUST 5 Validation

> Module: AI-powered automated code review with TRUST 5 validation framework and comprehensive quality analysis
> Complexity: Advanced
> Time: 35+ minutes
> Dependencies: static analyzers for your language, WebSearch/WebFetch (optional)

## Quick Reference

### Core Capabilities

The automated code review system provides comprehensive code quality analysis across multiple dimensions:

TRUST 5 Framework:
- Truthfulness: Code correctness and logic accuracy validation
- Relevance: Requirements fulfillment and purpose alignment checking
- Usability: Maintainability and understandability assessment
- Safety: Security vulnerability and error handling detection
- Timeliness: Performance standards and modern practices verification

Static Analysis Integration (per-language; see references/multi-language-support.md):
- Code quality / style linters for your language
- Security vulnerability scanners for your language
- Type checkers where the language is statically typed

Documentation-Enhanced Analysis:
- Up-to-date security patterns from OWASP and Semgrep
- Performance anti-patterns from profiling best practices
- Code quality patterns from SonarQube standards
- TRUST 5 validation framework patterns

### Key Components

This module describes a review workflow, not an importable SDK. Drive the review
with your language's own analyzers and aggregate the findings against the TRUST 5
rubric. Per-language tool inventories (Python pylint/flake8/bandit/mypy,
Go staticcheck/gosec, Rust clippy, JS/TS ESLint, etc.) live in
[../references/multi-language-support.md](../references/multi-language-support.md).

```
# Conceptual review workflow (language-neutral):
#   1. discover source files (include/exclude globs for your project)
#   2. run each language-appropriate analyzer over the file set
#   3. normalize findings into a common issue shape:
#        { category, severity, type, title, description,
#          file_path, line_number, suggested_fix, confidence }
#   4. score each TRUST 5 category (0.0-1.0) with severity-weighted penalties
#   5. compute the weighted overall score and emit a prioritized report
```

### TRUST 5 Scores

Category Score Calculation:
- Scores range from 0.0 to 1.0
- Penalties applied based on issue severity and confidence
- Weighted average for overall score
- Category weights: Truthfulness (25%), Relevance (20%), Usability (25%), Safety (20%), Timeliness (10%)

### Issue Severity Levels

Critical: Security vulnerabilities, syntax errors, data loss risks
High: Complex logic issues, major performance problems, significant safety concerns
Medium: Code smells, maintainability issues, moderate performance problems
Low: Style violations, minor documentation issues, small optimizations
Info: Suggestions and best practice recommendations

---

## Implementation Guide

### Basic Code Review Workflow

Step 1: Initialize the automated code reviewer with optional Documentation client for enhanced pattern detection

Step 2: Review the codebase by specifying:
- Project path to analyze
- Include patterns for files to review (default: ["/*.py"])
- Exclude patterns for directories to skip (default: ["/__pycache__/", "/venv/", "/tests/"])

Step 3: Analyze the generated report which includes:
- Overall TRUST score across all categories
- Per-file review results with individual issues
- Summary metrics with issue counts by severity and category
- Critical issues requiring immediate attention
- Actionable recommendations prioritized by impact

### Single File Review

For reviewing individual files, run the language-appropriate analyzer on one
file and map its findings into the common issue shape above. Per-file trust is
the same severity-weighted score computed over that file's issues only.

### Understanding Code Issues

Each detected issue carries these fields (the normalized shape every analyzer's
output is mapped into): category, severity, type, title, description,
file_path, line_number, code_snippet, suggested_fix, confidence, optional
rule_violated, optional external_reference.

### Customizing Analysis Patterns

Configure thresholds to match project standards. Tune these per project
(common dimensions: long-function max lines, max cyclomatic complexity, max
nesting depth) and re-run the analyzers with the adjusted configuration.
Quality-threshold tuning is project-specific; there is no SDK call to make.

---

## Advanced Modules

For detailed implementation and advanced features, see the specialized modules:

### TRUST 5 Validation Framework

See [trust5-validation.md](./trust5-validation.md) for:
- Complete TRUST 5 category implementations
- Custom validation rules and patterns
- Category-specific analysis methods
- Score calculation algorithms
- Penalty and weight customization

### Static Analysis Integration

See [static-analysis.md](./static-analysis.md) for:
- pylint, flake8, bandit, mypy integration details
- Tool configuration and customization
- Result parsing and normalization
- Tool-to-TRUST category mapping
- Error handling and fallback strategies

### Security Analysis

See [security-analysis.md](./security-analysis.md) for:
- Documentation-enhanced security pattern detection
- OWASP Top 10 vulnerability scanning
- SQL injection, command injection, path traversal detection
- Security fix suggestions with references
- Business logic vulnerability analysis

### Quality Metrics

See [quality-metrics.md](./quality-metrics.md) for:
- Function length and complexity analysis
- Nesting depth detection
- Cyclomatic complexity calculation
- Code metrics and statistics
- Maintainability indices

### Advanced TRUST 5 Framework

See [automated-code-review/trust5-framework.md](./automated-code-review/trust5-framework.md) for:
- Deep dive into TRUST 5 methodology
- Category-specific analysis patterns
- Advanced scoring algorithms
- Custom rule creation
- Integration with external validation tools

### Documentation Integration

When up-to-date framework or library patterns are needed during review, use WebSearch / WebFetch against the official documentation:
- Search for the framework's official docs site
- Fetch the relevant section (security, performance, code-quality) to ground pattern recommendations
- Fall back to established best-practice patterns when a source is unreachable

### Review Workflows

See [automated-code-review/review-workflows.md](./automated-code-review/review-workflows.md) for:
- CI/CD pipeline integration
- Automated review workflows
- Report generation and formatting
- Team collaboration patterns
- Continuous quality monitoring

---

## Best Practices

1. Comprehensive Coverage: Analyze code across all TRUST 5 dimensions for complete quality assessment
2. Context Integration: Leverage Documentation for up-to-date security and quality patterns
3. Actionable Feedback: Provide specific, implementable suggestions with code examples
4. Severity Prioritization: Focus on critical and high-severity issues first for maximum impact
5. Continuous Integration: Integrate into CI/CD pipeline for automated reviews on every commit
6. Custom Thresholds: Adjust analysis thresholds to match project standards and team preferences
7. Regular Updates: Keep Documentation patterns current for latest vulnerability detection
8. Team Consistency: Use consistent review rules across entire codebase for uniform quality

---

## Related Modules

- [Smart Refactoring](./smart-refactoring.md): Automated refactoring with code quality improvements
- [Performance Optimization](./performance-optimization.md): Performance profiling and bottleneck detection
- [AI Debugging](./ai-debugging.md): AI-powered debugging and error resolution

---

## Module Structure

```
automated-code-review.md (this file)
├── trust5-validation.md (TRUST 5 framework implementation)
├── static-analysis.md (pylint, flake8, bandit, mypy integration)
├── security-analysis.md (security vulnerability detection)
├── quality-metrics.md (code quality, complexity, metrics)
└── automated-code-review/
    ├── trust5-framework.md (deep dive into TRUST 5 categories)
    └── review-workflows.md (CI/CD and team workflows)
```

---

Version: 2.0.0 (Modular Structure)
Last Updated: 2026-01-06
Module: `modules/automated-code-review.md`
