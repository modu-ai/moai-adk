# Advanced Scoring Algorithms - TRUST 5 Framework

> Module: Weighted scoring algorithms with complexity adjustment and trend analysis
> Parent: [TRUST 5 Framework](./trust5-framework.md)
> Complexity: Advanced
> Time: 10+ minutes
> Dependencies: none (conceptual scoring reference)

## Overview

TRUST 5 scoring uses weighted category calculations (25%, 20%, 25%, 20%, 10%) with advanced factors including severity weighting, confidence scoring, complexity adjustment, and trend analysis for accurate quality assessment.

## Weighted Category Scoring

### Comprehensive Score Calculation

```text
calculate_advanced_trust_scores(issues, metrics):
    category_scores = {}
    category_weights = {
        TRUTHFULNESS: 0.25, RELEVANCE: 0.20, USABILITY: 0.25,
        SAFETY: 0.20, TIMELINESS: 0.10
    }

    for category in TrustCategory:
        category_issues = [i for i in issues if i.category == category]

        # Base penalty: sum of severity_weight * confidence * impact per issue
        penalty = 0.0
        for issue in category_issues:
            penalty += severity_weight(issue.severity) * issue.confidence * impact_factor(issue, metrics)

        # Adjustments
        complexity_penalty = calculate_complexity_penalty(category, metrics)
        trend_factor       = calculate_trend_factor(category, category_issues)

        total_penalty    = penalty + complexity_penalty
        trended_penalty  = total_penalty * trend_factor
        category_scores[category] = max(0.0, 1.0 - min(trended_penalty, 1.0))

    overall = sum(category_scores[c] * category_weights[c] for c in TrustCategory)
    return {
        overall, categories: category_scores,
        trend_factors: { c: calculate_trend_factor(c, [i for i in issues if i.category == c]) for c in TrustCategory }
    }
```

## Scoring Factors

### Severity Weighting

```text
severity_weight(severity):
    weights = { critical: 0.8, high: 0.6, medium: 0.4, low: 0.2 }
    return weights[lowercase(severity)] default 0.3
```

### Confidence Factor

Confidence factor (0.0-1.0) from detection algorithm:
- High confidence (0.8+): Certain detections
- Medium confidence (0.5-0.8): Probable issues
- Low confidence (<0.5): Possible issues

### Impact Factor

```text
impact_factor(issue, metrics):
    impact = 1.0
    # Higher-complexity code amplifies the impact
    if metrics.cyclomatic_complexity default 0 > 20: impact *= 1.2
    # Larger files amplify the impact
    if metrics.line_count default 0 > 500: impact *= 1.1
    return min(impact, 2.0)   # cap at 2x
```

## Advanced Adjustments

### Complexity Penalty

```text
calculate_complexity_penalty(category, metrics):
    base_penalty = 0.0
    # High complexity increases the USABILITY penalty
    if category == USABILITY:
        complexity = metrics.cyclomatic_complexity default 0
        if complexity > 15: base_penalty += 0.1
        if complexity > 25: base_penalty += 0.1
    return base_penalty
```

### Trend Factor

```text
calculate_trend_factor(category, issues):
    historical = get_historical_issues(category)
    if historical is empty: return 1.0      # no trend data
    current_count    = len(issues)
    historical_count = len(historical)
    if current_count < historical_count: return 0.95   # improving — boost
    if current_count > historical_count: return 1.05   # declining — penalty
    return 1.0                                          # stable
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
