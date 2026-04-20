# Automated Code Review - Tool Integration

> Sub-module: Static analysis tool wrappers and integration
> Parent: [Automated Code Review](../automated-code-review.md)

## StaticAnalysisTools Class

```python
class StaticAnalysisTools:
    """Wrapper for various static analysis tools."""

    def __init__(self):
        self.tools = {
            'pylint': self._run_pylint,
            'flake8': self._run_flake8,
            'bandit': self._run_bandit,
            'mypy': self._run_mypy
        }

    async def run_all_analyses(self, file_path: str) -> Dict[str, Any]:
        """Run all available static analysis tools."""
        results = {}

        for tool_name, tool_func in self.tools.items():
            try:
                result = await tool_func(file_path)
                results[tool_name] = result
            except Exception as e:
                print(f"Error running {tool_name}: {e}")
                results[tool_name] = {'error': str(e)}

        return results
```

## Pylint Integration

```python
async def _run_pylint(self, file_path: str) -> Dict[str, Any]:
    """Run pylint analysis."""
    try:
        result = subprocess.run(
            ['pylint', file_path, '--output-format=json'],
            capture_output=True,
            text=True
        )

        if result.returncode == 0:
            return {'issues': []}

        try:
            issues = json.loads(result.stdout)
            return {
                'issues': issues,
                'summary': self._parse_pylint_summary(result.stderr)
            }
        except json.JSONDecodeError:
            return {
                'raw_output': result.stdout,
                'raw_errors': result.stderr
            }
    except FileNotFoundError:
        return {'error': 'pylint not installed'}

def _parse_pylint_summary(self, stderr: str) -> Dict[str, Any]:
    """Parse pylint summary from stderr."""
    summary = {}
    for line in stderr.split('\n'):
        if 'rated at' in line:
            match = re.search(r'rated at ([\d.]+)/10', line)
            if match:
                summary['rating'] = float(match.group(1))
    return summary
```

## Flake8 Integration

```python
async def _run_flake8(self, file_path: str) -> Dict[str, Any]:
    """Run flake8 analysis."""
    try:
        result = subprocess.run(
            ['flake8', file_path, '--format=json'],
            capture_output=True,
            text=True
        )

        if result.returncode == 0:
            return {'issues': []}

        issues = []
        for line in result.stdout.split('\n'):
            if line.strip():
                parts = line.split(':')
                if len(parts) >= 4:
                    issues.append({
                        'path': parts[0],
                        'line': int(parts[1]),
                        'column': int(parts[2]),
                        'code': parts[3].strip(),
                        'message': ':'.join(parts[4:]).strip()
                    })

        return {'issues': issues}
    except FileNotFoundError:
        return {'error': 'flake8 not installed'}
```

## Bandit Security Integration

```python
async def _run_bandit(self, file_path: str) -> Dict[str, Any]:
    """Run bandit security analysis."""
    try:
        result = subprocess.run(
            ['bandit', '-f', 'json', file_path],
            capture_output=True,
            text=True
        )

        try:
            bandit_results = json.loads(result.stdout)
            return bandit_results
        except json.JSONDecodeError:
            return {'raw_output': result.stdout}
    except FileNotFoundError:
        return {'error': 'bandit not installed'}
```

## Mypy Type Checking Integration

```python
async def _run_mypy(self, file_path: str) -> Dict[str, Any]:
    """Run mypy type analysis."""
    try:
        result = subprocess.run(
            ['mypy', file_path, '--show-error-codes'],
            capture_output=True,
            text=True
        )

        issues = []
        for line in result.stdout.split('\n'):
            if ':' in line and 'error:' in line:
                parts = line.split(':', 3)
                if len(parts) >= 4:
                    issues.append({
                        'path': parts[0],
                        'line': int(parts[1]),
                        'message': parts[3].strip()
                    })

        return {'issues': issues}
    except FileNotFoundError:
        return {'error': 'mypy not installed'}
```

## Issue Conversion

```python
def _convert_static_issues(
    self, static_results: Dict[str, Any], file_path: str
) -> List[CodeIssue]:
    """Convert static analysis results to CodeIssue objects."""
    issues = []

    for tool_name, results in static_results.items():
        if 'error' in results:
            continue

        tool_issues = results.get('issues', [])
        for issue_data in tool_issues:
            category = self._map_tool_to_trust_category(tool_name, issue_data)

            issue = CodeIssue(
                id=f"{tool_name}_{len(issues)}",
                category=category,
                severity=self._map_severity(issue_data.get('severity', 'medium')),
                issue_type=self._map_issue_type(tool_name, issue_data),
                title=f"{tool_name.title()}: {issue_data.get('message', 'Unknown')}",
                description=issue_data.get('message', 'Static analysis issue'),
                file_path=file_path,
                line_number=issue_data.get('line', 0),
                column_number=issue_data.get('column', 0),
                code_snippet=issue_data.get('code_snippet', ''),
                suggested_fix=self._get_suggested_fix(tool_name, issue_data),
                confidence=0.8,
                rule_violated=issue_data.get('code', ''),
                external_reference=f"{tool_name} documentation"
            )
            issues.append(issue)

    return issues
```

## Category Mapping

```python
def _map_tool_to_trust_category(
    self, tool_name: str, issue_data: Dict
) -> TrustCategory:
    """Map static analysis tool to TRUST category."""
    if tool_name == 'bandit':
        return TrustCategory.SAFETY
    elif tool_name == 'mypy':
        return TrustCategory.TRUTHFULNESS
    elif tool_name == 'pylint':
        message = issue_data.get('message', '').lower()
        if any(kw in message for kw in ['security', 'injection', 'unsafe']):
            return TrustCategory.SAFETY
        elif any(kw in message for kw in ['performance', 'inefficient']):
            return TrustCategory.TIMELINESS
        else:
            return TrustCategory.USABILITY
    return TrustCategory.USABILITY

def _map_severity(self, severity: str) -> Severity:
    """Map severity string to Severity enum."""
    severity_map = {
        'critical': Severity.CRITICAL,
        'high': Severity.HIGH,
        'medium': Severity.MEDIUM,
        'low': Severity.LOW,
        'info': Severity.INFO
    }
    return severity_map.get(severity.lower(), Severity.MEDIUM)

def _map_issue_type(self, tool_name: str, issue_data: Dict) -> IssueType:
    """Map tool issue to IssueType enum."""
    if tool_name == 'bandit':
        return IssueType.SECURITY_VULNERABILITY
    elif tool_name == 'mypy':
        return IssueType.TYPE_ERROR

    message = issue_data.get('message', '').lower()
    if 'security' in message:
        return IssueType.SECURITY_VULNERABILITY
    elif 'performance' in message:
        return IssueType.PERFORMANCE_ISSUE
    elif 'syntax' in message:
        return IssueType.SYNTAX_ERROR
    return IssueType.CODE_SMELL
```

## Score Calculation

```python
def _calculate_trust_scores(
    self, issues: List[CodeIssue], metrics: Dict[str, Any]
) -> Dict[str, Any]:
    """Calculate TRUST 5 scores."""
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
    category_scores = {}
    for category in TrustCategory:
        category_issues = issues_by_category[category]

        penalty = 0.0
        for issue in category_issues:
            severity_penalty = {
                Severity.CRITICAL: 0.5,
                Severity.HIGH: 0.3,
                Severity.MEDIUM: 0.1,
                Severity.LOW: 0.05,
                Severity.INFO: 0.01
            }
            penalty += severity_penalty.get(issue.severity, 0.1) * issue.confidence

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

## Report Generation

```python
def _generate_comprehensive_report(
    self, project_path: str, file_results: List[FileReviewResult], duration: float
) -> CodeReviewReport:
    """Generate comprehensive code review report."""
    all_issues = []
    for result in file_results:
        all_issues.extend(result.issues)

    # Calculate overall scores
    overall_category_scores = {}
    for category in TrustCategory:
        scores = [result.category_scores.get(category, 0.0) for result in file_results]
        overall_category_scores[category] = sum(scores) / len(scores) if scores else 0.0

    overall_trust_score = sum(overall_category_scores.values()) / len(overall_category_scores)

    # Get critical issues
    critical_issues = [i for i in all_issues if i.severity == Severity.CRITICAL]

    # Generate recommendations
    recommendations = self._generate_recommendations(overall_category_scores, all_issues)

    summary_metrics = {
        'files_reviewed': len(file_results),
        'total_issues': len(all_issues),
        'critical_issues': len(critical_issues),
        'issues_by_severity': {
            s.value: len([i for i in all_issues if i.severity == s])
            for s in Severity
        },
        'average_trust_score': overall_trust_score
    }

    return CodeReviewReport(
        project_path=project_path,
        files_reviewed=file_results,
        overall_trust_score=overall_trust_score,
        overall_category_scores=overall_category_scores,
        summary_metrics=summary_metrics,
        recommendations=recommendations,
        critical_issues=critical_issues,
        review_duration=duration,
        context7_patterns_used=list(self.analysis_patterns.keys())
    )

def _generate_recommendations(
    self, category_scores: Dict[TrustCategory, float], issues: List[CodeIssue]
) -> List[str]:
    """Generate actionable recommendations."""
    recommendations = []

    for category, score in category_scores.items():
        if score < 0.7:
            if category == TrustCategory.SAFETY:
                recommendations.append("Adddess security vulnerabilities immediately")
            elif category == TrustCategory.TRUTHFULNESS:
                recommendations.append("Review code logic and fix correctness issues")
            elif category == TrustCategory.USABILITY:
                recommendations.append("Improve maintainability by refactoring")
            elif category == TrustCategory.RELEVANCE:
                recommendations.append("Remove TODOs and improve documentation")
            elif category == TrustCategory.TIMELINESS:
                recommendations.append("Optimize performance and update deprecated code")

    high_severity = len([i for i in issues if i.severity in [Severity.CRITICAL, Severity.HIGH]])
    if high_severity > 0:
        recommendations.append(f"Adddess {high_severity} high-priority issues before release")

    auto_fixable = len([i for i in issues if i.auto_fixable])
    if auto_fixable > 0:
        recommendations.append(f"Use automated fixes for {auto_fixable} auto-fixable issues")

    return recommendations
```

## Related Sub-modules

- [Core Classes](./core-classes.md) - Data structures and main classes
- [Analysis Patterns](./analysis-patterns.md) - TRUST 5 analysis methods

---

Sub-module: `modules/code-review/tool-integration.md`
