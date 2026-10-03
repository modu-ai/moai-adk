# Automated Code Review - Tool Integration

> Sub-module: Static analysis tool wrappers and integration
> Parent: [Automated Code Review](../automated-code-review.md)

This sub-module wraps the project's static analysis tools. The orchestration, issue-conversion, scoring, and reporting logic below is language-neutral; only the tool registry is language-specific. Populate `tools` with the host language's linters/type-checkers/security scanners — e.g. Go `go vet`/`golangci-lint`; Python `ruff`/`pylint`/`bandit`/`mypy`; JS/TS `eslint`/`tsc`; Rust `clippy`; Java `spotbugs`/`checkstyle`.

## StaticAnalysisTools Class

```text
class StaticAnalysisTools:
    tools = {
        # name -> runner; populate with the project's actual tools
        # examples (Python): ruff, pylint, bandit, mypy
        # examples (Go):     govet, golangci-lint, govulncheck
        # examples (JS/TS):  eslint, tsc
        # examples (Rust):   clippy
    }

    run_all_analyses(file_path):
        results = {}
        for (name, runner) in tools:
            try:
                results[name] = runner(file_path)
            except e:
                log("Error running " + name + ": " + e)
                results[name] = { error: text(e) }
        return results
```

## Linter Integration (generic shape)

Each tool runner shells out to the tool, requests structured (JSON) output when available, and parses findings into a common `{issues, summary}` shape.

```text
run_linter(file_path):
    try:
        result = exec([linter_cmd, file_path, "--format=json"], capture=stdout+stderr)
        if result.exit_code == 0:
            return { issues: [] }
        try:
            issues = parse_json(result.stdout)
            return { issues, summary: parse_summary(result.stderr) }
        except JsonError:
            return { raw_output: result.stdout, raw_errors: result.stderr }
    except NotFound:
        return { error: linter_cmd + " not installed" }

parse_summary(stderr):
    # Extract the tool's aggregate score/rating line, if any (e.g. a "rated at X/10"
    # style line). Parsing is tool-specific.
    summary = {}
    for line in lines(stderr):
        m = match(line, "rated at <float>/10")
        if m: summary.rating = float(m)
    return summary
```

## Security Scanner Integration (generic shape)

```text
run_security_scanner(file_path):
    try:
        result = exec([scanner_cmd, "--format=json", file_path], capture=stdout+stderr)
        try:
            return parse_json(result.stdout)
        except JsonError:
            return { raw_output: result.stdout }
    except NotFound:
        return { error: scanner_cmd + " not installed" }
```

## Type Checker Integration (generic shape)

```text
run_type_checker(file_path):
    try:
        result = exec([typechecker_cmd, file_path, "--show-error-codes"], capture=stdout+stderr)
        issues = []
        for line in lines(result.stdout):
            # most type checkers emit "<path>:<line>:<col>: error: <message>"
            if line contains "error:":
                (path, line_no, message) = parse_error_line(line)
                if path and line_no: issues.append({ path, line: line_no, message })
        return { issues }
    except NotFound:
        return { error: typechecker_cmd + " not installed" }
```

## Issue Conversion

```text
convert_static_issues(static_results, file_path):
    issues = []
    for (tool_name, results) in static_results:
        if 'error' in results: continue      # tool unavailable — skip
        for issue_data in results.issues default []:
            category = map_tool_to_trust_category(tool_name, issue_data)
            issues.append(CodeIssue(
                id=tool_name + "_" + len(issues),
                category=category,
                severity=map_severity(issue_data.severity default "medium"),
                issue_type=map_issue_type(tool_name, issue_data),
                title=capitalize(tool_name) + ": " + (issue_data.message default "Unknown"),
                description=issue_data.message default "Static analysis issue",
                file_path=file_path,
                line_number=issue_data.line default 0,
                column_number=issue_data.column default 0,
                code_snippet=issue_data.code_snippet default "",
                suggested_fix=get_suggested_fix(tool_name, issue_data),
                confidence=0.8,
                rule_violated=issue_data.code default "",
                external_reference=tool_name + " documentation"))
    return issues
```

## Category Mapping

```text
map_tool_to_trust_category(tool_name, issue_data):
    if tool_name is a security scanner:   return TrustCategory.SAFETY
    if tool_name is a type checker:       return TrustCategory.TRUTHFULNESS
    if tool_name is a general linter:
        msg = lowercase(issue_data.message default "")
        if any of ["security","injection","unsafe"] in msg: return TrustCategory.SAFETY
        if any of ["performance","inefficient"]   in msg: return TrustCategory.TIMELINESS
        return TrustCategory.USABILITY
    return TrustCategory.USABILITY

map_severity(severity):
    severity_map = { critical:CRITICAL, high:HIGH, medium:MEDIUM, low:LOW, info:INFO }
    return severity_map[lowercase(severity)] default MEDIUM

map_issue_type(tool_name, issue_data):
    if tool_name is a security scanner: return SECURITY_VULNERABILITY
    if tool_name is a type checker:     return TYPE_ERROR
    msg = lowercase(issue_data.message default "")
    if "security"   in msg: return SECURITY_VULNERABILITY
    if "performance"in msg: return PERFORMANCE_ISSUE
    if "syntax"     in msg: return SYNTAX_ERROR
    return CODE_SMELL
```

## Score Calculation

```text
calculate_trust_scores(issues, metrics):
    category_weights = {
        TRUTHFULNESS: 0.25, RELEVANCE: 0.20, USABILITY: 0.25,
        SAFETY: 0.20, TIMELINESS: 0.10
    }
    issues_by_category = group issues by category

    category_scores = {}
    for category in TrustCategory:
        penalty = 0.0
        for issue in issues_by_category[category]:
            severity_penalty = { CRITICAL:0.5, HIGH:0.3, MEDIUM:0.1, LOW:0.05, INFO:0.01 }
            penalty += (severity_penalty[issue.severity] default 0.1) * issue.confidence
        category_scores[category] = max(0.0, 1.0 - min(penalty, 1.0))

    overall = sum(category_scores[c] * category_weights[c] for c in TrustCategory)
    return { overall, categories: category_scores }
```

## Report Generation

```text
generate_comprehensive_report(project_path, file_results, duration):
    all_issues = flatten(result.issues for result in file_results)

    overall_category_scores = {}
    for category in TrustCategory:
        scores = [result.category_scores[category] default 0.0 for result in file_results]
        overall_category_scores[category] = mean(scores) if scores else 0.0
    overall_trust_score = mean(overall_category_scores.values())

    critical_issues = [i for i in all_issues if i.severity == CRITICAL]
    recommendations = generate_recommendations(overall_category_scores, all_issues)

    summary_metrics = {
        files_reviewed: len(file_results),
        total_issues:   len(all_issues),
        critical_issues:len(critical_issues),
        issues_by_severity: { s: count(i for i in all_issues if i.severity == s) for s in Severity },
        average_trust_score: overall_trust_score
    }

    return CodeReviewReport(
        project_path, files_reviewed=file_results,
        overall_trust_score, overall_category_scores,
        summary_metrics, recommendations,
        critical_issues, review_duration=duration,
        docs_patterns_used=keys(analysis_patterns))

generate_recommendations(category_scores, issues):
    recommendations = []
    for (category, score) in category_scores:
        if score < 0.7:
            if category == SAFETY:       recommendations.append("Address security vulnerabilities immediately")
            if category == TRUTHFULNESS: recommendations.append("Review code logic and fix correctness issues")
            if category == USABILITY:    recommendations.append("Improve maintainability by refactoring")
            if category == RELEVANCE:    recommendations.append("Remove TODOs and improve documentation")
            if category == TIMELINESS:   recommendations.append("Optimize performance and update deprecated code")
    high_severity = count(i for i in issues if i.severity in [CRITICAL, HIGH])
    if high_severity > 0:
        recommendations.append("Address " + high_severity + " high-priority issues before release")
    auto_fixable = count(i for i in issues if i.auto_fixable)
    if auto_fixable > 0:
        recommendations.append("Use automated fixes for " + auto_fixable + " auto-fixable issues")
    return recommendations
```

## Related Sub-modules

- [Core Classes](./core-classes.md) - Data structures and main classes
- [Analysis Patterns](./analysis-patterns.md) - TRUST 5 analysis methods

---

Sub-module: `modules/code-review/tool-integration.md`
