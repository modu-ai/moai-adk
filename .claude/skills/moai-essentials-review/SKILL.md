---
name: moai-essentials-review
version: 4.0.0
created: 2025-10-22
updated: 2025-11-11
status: active
description: "Automated code review and quality analysis with Context7. Perform static analysis, security audits, style checks, and best practices validation. Use when reviewing code changes, ensuring quality standards, or implementing automated review pipelines."
keywords: ['code-review', 'quality-analysis', 'static-analysis', 'security-audit', 'automated-review', 'context7', 'mcp-integration', 'quality-gates']
allowed-tools: "Read, Write, Edit, Glob, Bash, AskUserQuestion, mcp__context7__resolve-library-id, mcp__context7__get-library-docs, WebFetch"
---

# Automated Code Review and Quality Analysis

## Quick Start

1. **Configure review tools**: Set up static analyzers and linters
2. **Define quality gates**: Establish review criteria and thresholds
3. **Run automated reviews**: Integrate into CI/CD pipeline with Context7

## Core Review Patterns

### Pattern 1: Static Code Analysis

```python
# Automated static analysis configuration
# .pylintrc
[MAIN]
disable = C0114, C0115, C0116  # Disable docstring requirements for prototypes
load-plugins = pylint_django, pylint_flask

[DESIGN]
max-complexity = 10
max-args = 7
max-locals = 15

[FORMAT]
max-line-length = 88
```

### Pattern 2: Security Analysis

```python
# Security vulnerability scanning
import bandit
import subprocess

def run_security_scan(file_path):
    """Run Bandit security scanner"""
    try:
        result = subprocess.run([
            'bandit', '-r', file_path, '-f', 'json'
        ], capture_output=True, text=True)

        return {
            'issues': json.loads(result.stdout),
            'severity_counts': count_severity(result.stdout)
        }
    except Exception as e:
        return {'error': str(e)}
```

### Pattern 3: Code Quality Metrics

```python
# Quality metrics calculation
def calculate_quality_metrics(code_content):
    """Calculate various code quality metrics"""
    return {
        'cyclomatic_complexity': calculate_complexity(code_content),
        'maintainability_index': calculate_maintainability(code_content),
        'code_duplication': detect_duplicates(code_content),
        'test_coverage': get_coverage_percentage(),
        'documentation_coverage': calculate_doc_coverage(code_content)
    }
```

## Context7 Integration Examples

### Code Analysis Libraries with Context7

```python
# Get latest code review standards from Context7
def get_context7_review_standards():
    """
    Context7-backed code review standards:
    - Latest Python best practices
    - Security vulnerability patterns
    - Performance anti-patterns
    - Code style guidelines
    """
    pass
```

## Review Categories and Checkpoints

### 1. Code Style and Formatting

| Check | Tool | Description |
|-------|------|-------------|
| PEP 8 compliance | Black, Flake8 | Python style guide enforcement |
| Import sorting | isort | Organized import statements |
| Type hints | mypy | Static type checking |
| Docstring quality | pydocstyle | Documentation standards |

### 2. Security Analysis

```python
# Security review checklist
SECURITY_CHECKS = {
    'sql_injection': [
        'cursor.execute() with string formatting',
        'f-strings in SQL queries',
        '% formatting in SQL'
    ],
    'command_injection': [
        'os.system() with user input',
        'subprocess.call() with shell=True',
        'eval() on user data'
    ],
    'hardcoded_secrets': [
        'API keys in source code',
        'password strings',
        'database connection strings'
    ]
}
```

### 3. Performance Analysis

```python
# Performance review patterns
def review_performance_issues(code_ast):
    """Analyze code for performance anti-patterns"""
    issues = []

    # Check for inefficient loops
    for node in ast.walk(code_ast):
        if isinstance(node, ast.For):
            if has_nested_loops(node):
                issues.append({
                    'type': 'nested_loop',
                    'severity': 'medium',
                    'suggestion': 'Consider using list comprehensions or itertools'
                })

    return issues
```

### 4. Architecture and Design

```python
# Design pattern validation
def review_design_patterns(code_structure):
    """Review adherence to SOLID principles"""
    return {
        'single_responsibility': check_srp_violations(code_structure),
        'open_closed': check_ocp_violations(code_structure),
        'liskov_substitution': check_lsp_violations(code_structure),
        'interface_segregation': check_isp_violations(code_structure),
        'dependency_inversion': check_dip_violations(code_structure)
    }
```

## Automated Review Pipeline

### CI/CD Integration

```yaml
# .github/workflows/code-review.yml
name: Automated Code Review
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run Static Analysis
        run: |
          black --check .
          flake8 .
          mypy .

      - name: Run Security Scan
        run: |
          bandit -r . -f json -o security-report.json

      - name: Run Quality Metrics
        run: |
          radon cc . --json > complexity-report.json
          radon mi . --json > maintainability-report.json
```

### Review Report Generation

```python
# Generate comprehensive review report
def generate_review_report(project_path):
    """Generate comprehensive code review report"""
    tools = {
        'style': run_style_checks(project_path),
        'security': run_security_checks(project_path),
        'complexity': run_complexity_analysis(project_path),
        'coverage': run_coverage_analysis(project_path),
        'dependencies': check_dependencies(project_path)
    }

    return {
        'summary': generate_summary(tools),
        'issues': collect_all_issues(tools),
        'metrics': calculate_metrics(tools),
        'recommendations': generate_recommendations(tools)
    }
```

## Context7 Research Integration

### Review Research Patterns

```python
# Research-backed review techniques
def apply_research_based_reviews():
    """
    Latest findings from code review research:
    - Machine learning for defect prediction
    - Automated patch generation
    - Code smell detection algorithms
    - Review effectiveness optimization
    """
    pass
```

## Quality Gates and Thresholds

### Defining Quality Standards

```python
# Quality gate definitions
QUALITY_GATES = {
    'complexity': {
        'max_cyclomatic': 10,
        'max_cognitive': 15,
        'fail_threshold': 15
    },
    'security': {
        'max_high_issues': 0,
        'max_medium_issues': 5,
        'max_low_issues': 20
    },
    'coverage': {
        'min_statement_coverage': 80,
        'min_branch_coverage': 70,
        'min_function_coverage': 90
    },
    'maintainability': {
        'min_maintainability_index': 70,
        'max_duplication_percentage': 3
    }
}
```

### Gate Enforcement

```python
def enforce_quality_gates(review_results):
    """Enforce quality gate criteria"""
    passed = True
    failures = []

    for gate, config in QUALITY_GATES.items():
        if not check_gate_passed(review_results[gate], config):
            passed = False
            failures.append({
                'gate': gate,
                'actual': review_results[gate],
                'threshold': config
            })

    return {'passed': passed, 'failures': failures}
```

## Review Anti-patterns

### ❌ Common Review Mistakes

```python
# 1. Nitpicking over style
# Bad: Focus on trivial formatting issues
# Good: Focus on logical and architectural problems

# 2. Ignoring context
# Bad: Review without understanding requirements
# Good: Consider business context and constraints

# 3. Over-engineering suggestions
# Bad: Propose complex solutions for simple problems
# Good: Suggest pragmatic, maintainable approaches
```

## Real-World Examples

### Example 1: Web Service Review

```python
# Review checklist for REST APIs
API_REVIEW_CHECKLIST = {
    'endpoint_design': [
        'RESTful naming conventions',
        'Appropriate HTTP methods',
        'Consistent response formats'
    ],
    'security': [
        'Authentication implementation',
        'Rate limiting',
        'Input validation'
    ],
    'documentation': [
        'OpenAPI/Swagger spec',
        'Endpoint documentation',
        'Error response documentation'
    ]
}
```

### Example 2: Data Processing Pipeline Review

```python
# Data pipeline review patterns
def review_data_pipeline(code):
    """Review data processing code for common issues"""
    return {
        'data_quality': check_data_validations(code),
        'performance': check_processing_efficiency(code),
        'error_handling': check_exception_handling(code),
        'scalability': check_scaling_patterns(code),
        'monitoring': check_observability(code)
    }
```

## Review Best Practices

### Effective Review Process

1. **Clear Review Criteria**: Define what constitutes a good review
2. **Consistent Standards**: Apply rules uniformly across all code
3. **Constructive Feedback**: Focus on improvement, not criticism
4. **Context Awareness**: Understand the business and technical context
5. **Continuous Improvement**: Learn from review patterns and outcomes

### Review Communication

```python
# Review comment templates
REVIEW_TEMPLATES = {
    'security_issue': """
    🔒 Security Issue Detected
    **Risk**: {risk_level}
    **Location**: {file}:{line}
    **Issue**: {description}
    **Recommendation**: {fix_suggestion}
    """,

    'performance_concern': """
    ⚡ Performance Concern
    **Impact**: {performance_impact}
    **Current Complexity**: {complexity}
    **Suggestion**: {optimization_recommendation}
    """,

    'style_improvement': """
    🎨 Style Improvement
    **Current**: {current_pattern}
    **Recommended**: {recommended_pattern}
    **Reason**: {justification}
    """
}
```

## Review Automation Tools

### Tool Integration

```python
# Integration with popular review tools
REVIEW_TOOLS = {
    'sonarqube': {
        'url': 'https://sonarqube.example.com',
        'project_key': 'my-project',
        'quality_gate': 'my-quality-gate'
    },
    'codeclimate': {
        'token': os.getenv('CODECLIMATE_TOKEN'),
        'coverage_threshold': 80
    },
    'github': {
        'review_templates': 'templates/',
        'auto_merge_enabled': False,
        'required_reviews': 2
    }
}
```

## Review Metrics and KPIs

### Measuring Review Effectiveness

```python
# Review metrics tracking
def track_review_metrics():
    """Track key review effectiveness metrics"""
    return {
        'review_coverage': 'Percentage of code reviewed',
        'defect_density': 'Defects per thousand lines of code',
        'review_time': 'Average time per review',
        'fix_rate': 'Percentage of issues addressed',
        'recurrence_rate': 'Repeated issue frequency'
    }
```

## Review Checklist

- [ ] Configure static analysis tools
- [ ] Define quality gates and thresholds
- [ ] Set up automated review pipeline
- [ ] Create review templates and standards
- [ ] Establish review process documentation
- [ ] Configure notification and escalation
- [ ] Set up metrics tracking
- [ ] Review and update criteria regularly
- [ ] Train team on review standards
- [ ] Monitor review effectiveness

## Works Well With

- `Skill("moai-essentials-debug")` - Debugging review findings
- `Skill("moai-essentials-perf")` - Performance review patterns
- `Skill("moai-essentials-refactor")` - Refactoring recommendations
- `Skill("moai-foundation-trust")` - TRUST 5 principle enforcement

---

**Reference**: Automated code review best practices with Context7 integration
**Version**: 4.0.0 Enterprise
