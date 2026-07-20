# Code Review Workflows and CI/CD Integration

> Module: Automated review workflows for CI/CD pipelines and team collaboration
> Parent: [Automated Code Review](../automated-code-review.md)
> Complexity: Intermediate
> Time: 20+ minutes
> Dependencies: git, CI/CD platforms

## Quick Reference

### CI/CD Integration Platforms

GitHub Actions:
- Automated reviews on pull requests
- Status checks for quality gates
- Comment generation with findings
- Matrix builds for multiple runtime versions

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

```text
automated_review_workflow(project_path, pr_number = none, fail_on_quality_gate = true):
    reviewer = AutomatedCodeReviewer(docs_client=docs)

    # Review the codebase; use the host language's source globs
    report = reviewer.review_codebase(
        project_path=project_path,
        include_patterns=["<source-globs>"],   # e.g. **/*.go, **/*.py, **/*.ts
        exclude_patterns=["/tests/", "/migrations/", "vendor/"])

    quality_gate_passed = check_quality_gates(report)
    if fail_on_quality_gate and not quality_gate_passed:
        raise QualityGateError("Code review quality gates failed")

    generate_review_report(report, pr_number)
    return report
```

---

## Quality Gates

### Quality Gate Configuration

```text
class QualityGateConfig:
    gates = {
        overall_trust_score: 0.70,      # minimum overall score
        truthfulness_score:  0.75,      # minimum truthfulness
        safety_score:        0.80,      # minimum safety
        critical_issues:     0,         # no critical issues allowed
        high_issues:         5,         # max high-severity issues
        medium_issues:       20,        # max medium issues
        new_critical_issues: 0,         # no new critical issues
        coverage_percentage: 80.0       # minimum test coverage
    }

check_quality_gates(report, config):
    gates_passed = true
    failures = []

    if report.overall_trust_score < config.gates.overall_trust_score:
        gates_passed = false
        failures.append("Overall TRUST score " + report.overall_trust_score +
                        " below threshold " + config.gates.overall_trust_score)

    safety_score = report.overall_category_scores.get(SAFETY, 0.0)
    if safety_score < config.gates.safety_score:
        gates_passed = false
        failures.append("Safety score " + safety_score + " below threshold " + config.gates.safety_score)

    critical_count = len(report.critical_issues)
    if critical_count > config.gates.critical_issues:
        gates_passed = false
        failures.append("Found " + critical_count + " critical issues (max: " + config.gates.critical_issues + ")")

    high_count = report.summary_metrics.issues_by_severity.get("high", 0)
    if high_count > config.gates.high_issues:
        gates_passed = false
        failures.append("Found " + high_count + " high severity issues (max: " + config.gates.high_issues + ")")

    return { gates_passed, failures }
```

---

## GitHub Actions Integration

### Workflow Configuration

The workflow runs the MoAI review subcommand on PRs and pushes. Use the host language's setup action in place of the Python setup shown; the `moai review` invocation is language-neutral.

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
    # Optionally matrix across runtime versions for the host language.

    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Full history for better analysis

      - name: Set up runtime
        # Use the setup action for the host language, if review needs it.

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
          name: code-review-report
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
  image: <runtime-image>   # the host language's CI image
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
  image: <runtime-image>
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

```text
generate_markdown_report(report):
    md = "# Code Review Report\n\n## Executive Summary\n\n"
    md += "**Overall TRUST Score:** " + report.overall_trust_score + "\n"
    md += "**Files Reviewed:** " + report.summary_metrics.files_reviewed + "\n"
    md += "**Total Issues:** " + report.summary_metrics.total_issues + "\n"
    md += "**Critical Issues:** " + report.summary_metrics.critical_issues + "\n\n"
    md += "## TRUST 5 Category Scores\n\n"
    for (category, score) in report.overall_category_scores:
        md += "- **" + capitalize(category) + ":** " + score + "\n"
    md += "\n## Issues by Severity\n\n"
    for severity in ["critical", "high", "medium", "low"]:
        count = report.summary_metrics.issues_by_severity.get(severity, 0)
        md += "- **" + capitalize(severity) + ":** " + count + "\n"
    if report.critical_issues:
        md += "\n## Critical Issues\n\n"
        for issue in report.critical_issues[:10]:
            md += "### " + issue.title + "\n"
            md += "- **Location:** `" + issue.file_path + ":" + issue.line_number + "`\n"
            md += "- **Description:** " + issue.description + "\n"
            md += "- **Suggested Fix:** " + issue.suggested_fix + "\n"
            md += "- **Rule:** " + issue.rule_violated + "\n\n"
    md += "\n## Recommendations\n\n"
    for (i, rec) in enumerate(report.recommendations[:5], from=1):
        md += i + ". " + rec + "\n"
    return md
```

### HTML Report

```text
generate_html_report(report):
    # Build a self-contained HTML page: a conic-gradient score circle colored
    # by the overall score band (green/orange/red), category-score cards, and
    # a critical-issues list with severity-colored left borders.
    html = "<!DOCTYPE html><html><head><title>Code Review Report</title><style>"
    html += "body{font-family:sans-serif;margin:20px;}"
    html += ".score-circle{width:150px;height:150px;border-radius:50%;" +
            "background:conic-gradient(" + score_color(report.overall_trust_score) + " " +
            (report.overall_trust_score*360) + "deg,#f0f0f0 0deg);" +
            "display:flex;align-items:center;justify-content:center;font-size:32px;font-weight:bold;}"
    html += ".category-score{margin:10px;padding:10px;border:1px solid #ddd;}"
    html += ".issue{margin:10px 0;padding:10px;border-left:4px solid <severity-color>;}"
    html += "</style></head><body>"
    html += "<h1>Code Review Report</h1>"
    html += "<div class='score-circle'>" + report.overall_trust_score + "</div>"
    html += "<h2>Category Scores</h2><div>"
    for (category, score) in report.overall_category_scores:
        html += "<div class='category-score'><strong>" + capitalize(category) + ":</strong> " + score + "</div>"
    html += "</div>"
    if report.critical_issues:
        html += "<h2>Critical Issues</h2>"
        for issue in report.critical_issues[:10]:
            html += "<div class='issue'><strong>" + issue.title + "</strong><br>"
            html += "Location: " + issue.file_path + ":" + issue.line_number + "<br>"
            html += issue.description + "<br>"
            html += "<strong>Fix:</strong> " + issue.suggested_fix + "</div>"
    html += "</body></html>"
    return html
```

---

## Team Collaboration

### Pull Request Comments

```text
create_pr_review_comments(report, pr_number, github_client):
    # Group issues by file
    issues_by_file = {}
    for file_result in report.files_reviewed:
        for issue in file_result.issues:
            issues_by_file.setdefault(issue.file_path, []).append(issue)

    # Create review comments (limit per file)
    for (file_path, issues) in issues_by_file:
        for issue in issues[:5]:
            body = "**" + issue.title + "**\n"
            body += "**Severity:** " + issue.severity + "\n"
            body += "**Description:** " + issue.description + "\n"
            body += "**Suggested Fix:** " + issue.suggested_fix + "\n"
            if issue.external_reference: body += "[View Rule](" + issue.external_reference + ")\n"
            github_client.create_review_comment(pr_number, body, file_path, issue.line_number)
```

---

## Continuous Monitoring

### Trend Analysis

```text
class ReviewTrendAnalyzer(storage_path):
    save_review_report(report, timestamp = now()):
        report_data = {
            timestamp: timestamp,
            overall_score: report.overall_trust_score,
            category_scores: { cat: score for (cat, score) in report.overall_category_scores },
            issue_counts: report.summary_metrics.issues_by_severity
        }
        append_line(storage_path, to_json(report_data))    # JSONL history

    analyze_trends(days = 30):
        cutoff = now() - days*24*3600
        reports = [parse_json(line) for line in lines(storage_path) if parse_json(line).timestamp >= cutoff]
        if reports is empty: return { error: "No historical data available" }
        scores = [r.overall_score for r in reports]
        avg_score = mean(scores)
        score_change = scores[-1] - scores[0]
        trend = "improving" if score_change > 0.01
                else "declining" if score_change < -0.01
                else "stable"
        return {
            period_days: days, total_reviews: len(reports),
            average_score: avg_score, score_change: score_change, trend: trend,
            category_trends: calculate_category_trends(reports)
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
