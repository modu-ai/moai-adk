# Advanced Scoring Algorithms - TRUST 5 Framework

> Module: Weighted scoring algorithms with complexity adjustment and trend analysis
> Parent: [TRUST 5 Framework](./trust5-framework.md)
> Complexity: Advanced
> Time: 10+ minutes
> Dependencies: Python 3.8+

## Overview

TRUST 5 scoring uses weighted category calculations (25%, 20%, 25%, 20%, 10%) with advanced factors including severity weighting, confidence scoring, complexity adjustment, and trend analysis for accurate quality assessment.

## Weighted Category Scoring

### Comprehensive Score Calculation

```python
def calculate_advanced_trust_scores(
    self, issues: List[CodeIssue], metrics: Dict[str, Any]
) -> Dict[str, Any]:
    """Calculate advanced TRUST 5 scores with weighted factors."""

    category_scores = {}
    category_weights = {
        TrustCategory.TRUTHFULNESS: 0.25,
        TrustCategory.RELEVANCE: 0.20,
        TrustCategory.USABILITY: 0.25,
        TrustCategory.SAFETY: 0.20,
        TrustCategory.TIMELINESS: 0.10
    }

    # Calculate scores for each category
    for category in TrustCategory:
        category_issues = [i for i in issues if i.category == category]

        # Base penalty calculation
        penalty = 0.0
        for issue in category_issues:
            severity_weight = self._get_severity_weight(issue.severity)
            confidence_factor = issue.confidence
            impact_factor = self._get_impact_factor(issue, metrics)
            penalty += severity_weight * confidence_factor * impact_factor

        # Apply complexity adjustment
        complexity_penalty = self._calculate_complexity_penalty(category, metrics)

        # Apply trend factor (historical data)
        trend_factor = self._calculate_trend_factor(category, category_issues)

        # Calculate final score
        total_penalty = penalty + complexity_penalty
        trended_penalty = total_penalty * trend_factor
        score = max(0.0, 1.0 - min(trended_penalty, 1.0))
        category_scores[category] = score

    # Calculate overall score
    overall_score = sum(
        category_scores[cat] * category_weights[cat]
        for cat in TrustCategory
    )

    return {
        'overall': overall_score,
        'categories': category_scores,
        'trend_factors': {cat: self._calculate_trend_factor(cat, [i for i in issues if i.category == cat]) for cat in TrustCategory}
    }
```

## Scoring Factors

### Severity Weighting

```python
def _get_severity_weight(self, severity: str) -> float:
    """Get severity weight for penalty calculation."""

    weights = {
        'critical': 0.8,
        'high': 0.6,
        'medium': 0.4,
        'low': 0.2
    }
    return weights.get(severity.lower(), 0.3)
```

### Confidence Factor

Confidence factor (0.0-1.0) from detection algorithm:
- High confidence (0.8+): Certain detections
- Medium confidence (0.5-0.8): Probable issues
- Low confidence (<0.5): Possible issues

### Impact Factor

```python
def _get_impact_factor(
    self, issue: CodeIssue, metrics: Dict[str, Any]
) -> float:
    """Calculate impact factor based on issue context."""

    # Base impact
    impact = 1.0

    # Adjust for code complexity
    if metrics.get('cyclomatic_complexity', 0) > 20:
        impact *= 1.2

    # Adjust for file size
    if metrics.get('line_count', 0) > 500:
        impact *= 1.1

    return min(impact, 2.0)  # Cap at 2x
```

## Advanced Adjustments

### Complexity Penalty

```python
def _calculate_complexity_penalty(
    self, category: TrustCategory, metrics: Dict[str, Any]
) -> float:
    """Calculate complexity adjustment for category."""

    base_penalty = 0.0

    # High complexity increases penalty
    if category == TrustCategory.USABILITY:
        complexity = metrics.get('cyclomatic_complexity', 0)
        if complexity > 15:
            base_penalty += 0.1
        if complexity > 25:
            base_penalty += 0.1

    return base_penalty
```

### Trend Factor

```python
def _calculate_trend_factor(
    self, category: TrustCategory, issues: List[CodeIssue]
) -> float:
    """Calculate trend factor based on historical data."""

    # Get historical issues for this category
    historical = self._get_historical_issues(category)

    if not historical:
        return 1.0  # No trend data

    current_count = len(issues)
    historical_count = len(historical)

    # Improving trend (fewer issues)
    if current_count < historical_count:
        return 0.95  # Boost score

    # Declining trend (more issues)
    elif current_count > historical_count:
        return 1.05  # Penalty

    return 1.0  # Stable
```

## Score Interpretation

### Overall Score Ranges

- **0.90-1.00**: Excellent quality
- **0.75-0.89**: Good quality
- **0.60-0.74**: Acceptable quality
- **0.40-0.59**: Needs improvement
- **0.00-0.39**: Critical issues

### Category Score Analysis

Individual category scores identify specific improvement areas:
- **Truthfulness < 0.7**: Logic correctness issues
- **Relevance < 0.7**: Requirements alignment problems
- **Usability < 0.7**: Maintainability concerns
- **Safety < 0.7**: Security vulnerabilities
- **Timeliness < 0.7**: Performance optimization needed

## Best Practices

1. Baseline Establishment: Establish baseline scores for project
2. Trend Monitoring: Track scores over time
3. Threshold Setting: Set minimum score thresholds for CI/CD
4. Weight Customization: Adjust weights based on project priorities
5. Historical Analysis: Use trend data to identify patterns

---

Version: 1.0.0
Last Updated: 2026-01-06
