# Code Review Workflows and CI/CD Integration

> Module: Automated review workflows for CI/CD pipelines and team collaboration
> Parent: [Automated Code Review](../automated-code-review.md)
> Complexity: Intermediate
> Time: 20+ minutes
> Dependencies: Python 3.8+, git, CI/CD platforms

## Quick Reference

### CI/CD Integration Platforms

GitHub Actions:
- Automated reviews on pull requests
- Status checks for quality gates
- Comment generation with findings
- Matrix builds for multiple Python versions

GitLab CI/CD:
- Pipeline integration with code quality stages
- Merge request automation
- Quality gate enforcement
- Code quality reports

Jenkins:
- Pipeline as code integration
- Build failure on quality gate violations
- Trend analysis and reporting
- Multi-branch pipeline support

### Core Workflow Pattern

```python
async def automated_review_workflow(
    project_path: str,
    pr_number: int = None,
    fail_on_quality_gate: bool = True
) -> CodeReviewReport:
    """Execute automated code review workflow."""

    # Initialize reviewer
    reviewer = AutomatedCodeReviewer(context7_client=context7)

    # Run review
    report = await reviewer.review_codebase(
        project_path=project_path,
        include_patterns=["/*.py"],
        exclude_patterns=["/tests/", "/migrations/"]
    )

    # Check quality gates
    quality_gate_passed = check_quality_gates(report)

    if fail_on_quality_gate and not quality_gate_passed:
        raise QualityGateError("Code review quality gates failed")

    # Generate report
    generate_review_report(report, pr_number)

    return report
```

---

## Quality Gates

### Quality Gate Configuration

```python
class QualityGateConfig:
    """Quality gate configuration."""

    def __init__(self):
        self.gates = {
            'overall_trust_score': 0.70,      # Minimum overall score
            'truthfulness_score': 0.75,       # Minimum truthfulness
            'safety_score': 0.80,             # Minimum safety
            'critical_issues': 0,              # No critical issues allowed
            'high_issues': 5,                  # Maximum high severity issues
            'medium_issues': 20,               # Maximum medium issues
            'new_critical_issues': 0,          # No new critical issues
            'coverage_percentage': 80.0        # Minimum test coverage
        }

def check_quality_gates(report: CodeReviewReport, config: QualityGateConfig) -> bool:
    """Check if code review passes quality gates."""

    gates_passed = True
    failures = []

    # Check overall TRUST score
    if report.overall_trust_score < config.gates['overall_trust_score']:
        gates_passed = False
        failures.append(
            f"Overall TRUST score {report.overall_trust_score:.2f} "
            f"below threshold {config.gates['overall_trust_score']}"
        )

    # Check safety score
    safety_score = report.overall_category_scores.get(TrustCategory.SAFETY, 0.0)
    if safety_score < config.gates['safety_score']:
        gates_passed = False
        failures.append(
            f"Safety score {safety_score:.2f} "
            f"below threshold {config.gates['safety_score']}"
        )

    # Check critical issues
    critical_count = len(report.critical_issues)
    if critical_count > config.gates['critical_issues']:
        gates_passed = False
        failures.append(
            f"Found {critical_count} critical issues "
            f"(max: {config.gates['critical_issues']})"
        )

    # Check high severity issues
    high_count = report.summary_metrics['issues_by_severity'].get('high', 0)
    if high_count > config.gates['high_issues']:
        gates_passed = False
        failures.append(
            f"Found {high_count} high severity issues "
            f"(max: {config.gates['high_issues']})"
        )

    return gates_passed, failures
```

---

## GitHub Actions Integration

### Workflow Configuration

```yaml
# .github/workflows/code-review.yml
name: Automated Code Review

on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [main, develop]

jobs:
  code-review:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: [3.8, 3.9, 3.10, 3.11]

    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Full history for better analysis

      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: ${{ matrix.python-version }}

      - name: Install dependencies
        run: |
          # Go binary - no package installation needed

      - name: Run automated code review
        run: |
          moai review \
            --path . \
            --output review-report.json \
            --format json \
            --fail-on-gate

      - name: Upload review report
        uses: actions/upload-artifact@v3
        with:
          name: code-review-report-${{ matrix.python-version }}
          path: review-report.json

      - name: Comment PR with results
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const report = JSON.parse(fs.readFileSync('review-report.json', 'utf8'));

            const comment = `## Code Review Results
            **Overall TRUST Score:** ${report.overall_trust_score.toFixed(2)}

            ### Category Scores
            ${Object.entries(report.overall_category_scores).map(([cat, score]) =>
              `- **${cat}:** ${score.toFixed(2)}`
            ).join('\n')}

            ### Issues Summary
            - **Critical:** ${report.summary_metrics.critical_issues}
            - **High:** ${report.summary_metrics.issues_by_severity.high}
            - **Medium:** ${report.summary_metrics.issues_by_severity.medium}
            - **Low:** ${report.summary_metrics.issues_by_severity.low}

            ${report.critical_issues.length > 0 ? `
            ### Critical Issues
            ${report.critical_issues.map(issue =>
              `- **${issue.title}** in \`${issue.file_path}:${issue.line_number}\`
                ${issue.description}`
            ).join('\n')}
            ` : ''}
            `;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: comment
            });
```

---

## GitLab CI/CD Integration

### Pipeline Configuration

```yaml
# .gitlab-ci.yml
stages:
  - test
  - review
  - report

code_review:
  stage: review
  image: python:3.10
  script:
    - moai review --path . --output review-report.json --format json
  artifacts:
    paths:
      - review-report.json
    reports:
      codequality: review-report.json
    expire_in: 1 week
  only:
    - merge_requests
    - main
    - develop

quality_gate:
  stage: report
  image: python:3.10
  script:
    - moai quality-gate --report review-report.json --fail-on-violation
  dependencies:
    - code_review
  allow_failure: false
  only:
    - merge_requests
```

---

## Report Generation

### Markdown Report

```python
def generate_markdown_report(report: CodeReviewReport) -> str:
    """Generate comprehensive Markdown report."""

    md = f"""# Code Review Report

## Executive Summary

**Overall TRUST Score:** {report.overall_trust_score:.2f}
**Files Reviewed:** {report.summary_metrics['files_reviewed']}
**Total Issues:** {report.summary_metrics['total_issues']}
**Critical Issues:** {report.summary_metrics['critical_issues']}

## TRUST 5 Category Scores

"""

    for category, score in report.overall_category_scores.items():
        md += f"- **{category.value.title()}:** {score:.2f}\n"

    md += "\n## Issues by Severity\n\n"

    for severity in ['critical', 'high', 'medium', 'low']:
        count = report.summary_metrics['issues_by_severity'].get(severity, 0)
        md += f"- **{severity.title()}:** {count}\n"

    if report.critical_issues:
        md += "\n## Critical Issues\n\n"
        for issue in report.critical_issues[:10]:
            md += f"### {issue.title}\n"
            md += f"- **Location:** `{issue.file_path}:{issue.line_number}`\n"
            md += f"- **Description:** {issue.description}\n"
            md += f"- **Suggested Fix:** {issue.suggested_fix}\n"
            md += f"- **Rule:** {issue.rule_violated}\n\n"

    md += "\n## Recommendations\n\n"
    for i, rec in enumerate(report.recommendations[:5], 1):
        md += f"{i}. {rec}\n"

    return md
```

### HTML Report

```python
def generate_html_report(report: CodeReviewReport) -> str:
    """Generate HTML report with interactive elements."""

    html = f"""
<!DOCTYPE html>
<html>
<head>
    <title>Code Review Report</title>
    <style>
        body {{ font-family: Arial, sans-serif; margin: 20px; }}
        .score-circle {{
            width: 150px;
            height: 150px;
            border-radius: 50%;
            background: conic-gradient(
                {'green' if report.overall_trust_score > 0.8 else 'orange' if report.overall_trust_score > 0.6 else 'red'}
                {report.overall_trust_score * 360}deg,
                #f0f0f0 0deg
            );
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 32px;
            font-weight: bold;
        }}
        .category-score {{ margin: 10px; padding: 10px; border: 1px solid #ddd; }}
        .issue {{ margin: 10px 0; padding: 10px; border-left: 4px solid {'red' if issue.severity == 'critical' else 'orange' if issue.severity == 'high' else 'blue'}; }}
    </style>
</head>
<body>
    <h1>Code Review Report</h1>
    <div class="score-circle">{report.overall_trust_score:.2f}</div>

    <h2>Category Scores</h2>
    <div>
"""

    for category, score in report.overall_category_scores.items():
        html += f'<div class="category-score"><strong>{category.value.title()}:</strong> {score:.2f}</div>'

    html += "</div>"

    if report.critical_issues:
        html += "<h2>Critical Issues</h2>"
        for issue in report.critical_issues[:10]:
            html += f'<div class="issue"><strong>{issue.title}</strong><br>'
            html += f'Location: {issue.file_path}:{issue.line_number}<br>'
            html += f'{issue.description}<br>'
            html += f'<strong>Fix:</strong> {issue.suggested_fix}</div>'

    html += "</body></html>"

    return html
```

---

## Team Collaboration

### Pull Request Comments

```python
async def create_pr_review_comments(
    report: CodeReviewReport,
    pr_number: int,
    github_client
):
    """Create review comments on pull request."""

    # Group issues by file
    issues_by_file = {}
    for file_result in report.files_reviewed:
        for issue in file_result.issues:
            if issue.file_path not in issues_by_file:
                issues_by_file[issue.file_path] = []
            issues_by_file[issue.file_path].append(issue)

    # Create review comments
    for file_path, issues in issues_by_file.items():
        for issue in issues[:5]:  # Limit comments per file
            comment_body = f"""
**{issue.title}**
**Severity:** {issue.severity}
**Description:** {issue.description}
**Suggested Fix:** {issue.suggested_fix}

[View Rule]({issue.external_reference}) if available
"""

            await github_client.create_review_comment(
                pr_number=pr_number,
                body=comment_body,
                path=file_path,
                line=issue.line_number
            )
```

---

## Continuous Monitoring

### Trend Analysis

```python
class ReviewTrendAnalyzer:
    """Analyze code review trends over time."""

    def __init__(self, storage_path: str):
        self.storage_path = storage_path

    def save_review_report(self, report: CodeReviewReport, timestamp: float = None):
        """Save review report for trend analysis."""

        if timestamp is None:
            timestamp = time.time()

        report_data = {
            'timestamp': timestamp,
            'overall_score': report.overall_trust_score,
            'category_scores': {cat.value: score for cat, score in report.overall_category_scores.items()},
            'issue_counts': report.summary_metrics['issues_by_severity']
        }

        # Append to history file
        with open(self.storage_path, 'a') as f:
            f.write(json.dumps(report_data) + '\n')

    def analyze_trends(self, days: int = 30) -> Dict[str, Any]:
        """Analyze review trends over specified period."""

        cutoff_time = time.time() - (days * 24 * 3600)
        reports = []

        # Load historical reports
        with open(self.storage_path, 'r') as f:
            for line in f:
                report_data = json.loads(line.strip())
                if report_data['timestamp'] >= cutoff_time:
                    reports.append(report_data)

        if not reports:
            return {'error': 'No historical data available'}

        # Calculate trends
        scores = [r['overall_score'] for r in reports]
        avg_score = sum(scores) / len(scores)

        # Compare first and last
        score_change = scores[-1] - scores[0]
        trend = 'improving' if score_change > 0.01 else 'declining' if score_change < -0.01 else 'stable'

        return {
            'period_days': days,
            'total_reviews': len(reports),
            'average_score': avg_score,
            'score_change': score_change,
            'trend': trend,
            'category_trends': self._calculate_category_trends(reports)
        }
```

---

## Best Practices

1. Quality Gates: Set appropriate thresholds for project quality standards
2. Incremental Rollout: Start with warning-only gates, gradually enforce
3. Team Training: Educate team on review feedback and best practices
4. Custom Rules: Customize rules for project-specific requirements
5. Regular Updates: Keep security patterns and quality rules current
6. Performance: Cache results to avoid redundant analysis
7. Feedback Loop: Use review insights for continuous improvement
8. Integration: Seamlessly integrate with existing CI/CD workflows

---

## Related Modules

- [Automated Code Review](../automated-code-review.md): Main review system
- [trust5-validation.md](../trust5-validation.md): Quality gate configuration

---

Version: 1.0.0
Last Updated: 2026-01-06
Module: `modules/automated-code-review/review-workflows.md`
