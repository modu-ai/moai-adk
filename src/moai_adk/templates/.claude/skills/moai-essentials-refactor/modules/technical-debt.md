# Technical Debt Quantification

AI-powered technical debt analysis, quantification, and prioritization framework.

## Overview

Technical debt represents the implied cost of additional rework caused by choosing an easy solution now instead of using a better approach that would take longer. This module provides systematic approaches to measure, quantify, and prioritize technical debt.

---

## Debt Indicators

### Code Smell Types

| Smell Type | Weight | Detection Method | Priority |
|------------|--------|------------------|----------|
| Duplicated Code | 0.25 | AST similarity analysis | High |
| Complex Methods | 0.20 | Cyclomatic complexity | High |
| Large Classes | 0.15 | Line count + method count | Medium |
| Long Parameter Lists | 0.10 | Parameter count | Medium |
| Feature Envy | 0.10 | Coupling analysis | Medium |
| Shotgun Surgery | 0.10 | Change impact analysis | Medium |
| Primitive Obsession | 0.10 | Type analysis | Low |

---

## Debt Score Calculation

### Scoring Formula

```python
class DebtScoreCalculator:
    """Calculate technical debt score."""

    def calculate_debt_score(self, indicators: List[DebtIndicator]) -> float:
        """
        Calculate overall debt score (0-100).

        Formula: Σ(severity × frequency × weight)
        """
        weights = {
            "duplicated_code": 0.25,
            "complex_methods": 0.20,
            "large_classes": 0.15,
            "long_parameter_lists": 0.10,
            "feature_envy": 0.10,
            "shotgun_surgery": 0.10,
            "primitive_obsession": 0.10
        }

        total_score = 0.0

        for indicator in indicators:
            # Score = severity (1-10) × frequency (count)
            score = indicator.severity * indicator.frequency

            # Weight by importance
            weighted_score = score * weights.get(indicator.type, 0.05)

            total_score += weighted_score

        # Normalize to 0-100
        return min(total_score, 100.0)
```

### Severity Levels

```python
SEVERITY_LEVELS = {
    1: "Trivial",     # Minor code style issue
    2: "Low",         # Small quality issue
    3: "Low",         # Minor maintainability issue
    4: "Medium",      # Noticeable quality issue
    5: "Medium",      # Significant maintainability issue
    6: "Medium",      # Major quality issue
    7: "High",        # Critical maintainability issue
    8: "High",        # Severe quality issue
    9: "Critical",    # Immediate attention required
    10: "Critical"    # Blocking production deployment
}
```

---

## Effort Estimation

### Estimation Model

```python
class RefactoringEffortEstimator:
    """Estimate refactoring effort in person-days."""

    EFFORT_MAP = {
        "duplicated_code": 0.5,      # 0.5 days per instance
        "complex_methods": 0.3,      # 0.3 days per method
        "large_classes": 1.0,        # 1 day per class
        "long_parameter_lists": 0.2, # 0.2 days per method
        "feature_envy": 0.4,         # 0.4 days per method
        "shotgun_surgery": 0.6,      # 0.6 days per change
        "primitive_obsession": 0.3   # 0.3 days per type
    }

    def estimate_effort(self, indicators: List[DebtIndicator]) -> float:
        """Estimate total refactoring effort."""
        total_days = 0.0

        for indicator in indicators:
            days_per_instance = self.EFFORT_MAP.get(indicator.type, 0.5)
            total_days += days_per_instance * indicator.frequency

        return total_days

    def categorize_effort(self, total_days: float) -> str:
        """Categorize effort level."""
        if total_days < 5:
            return f"{total_days:.1f} days (Low)"
        elif total_days < 15:
            return f"{total_days:.1f} days (Medium)"
        else:
            return f"{total_days:.1f} days (High)"
```

---

## Priority Calculation

### Prioritization Formula

```python
class DebtPrioritizer:
    """Prioritize refactoring actions by impact and effort."""

    def calculate_priority(self, indicator: DebtIndicator) -> float:
        """
        Calculate priority score.

        Formula: impact_score / effort_score
        High impact, low effort = high priority
        """
        impact_score = self._calculate_impact_score(indicator)
        effort_score = self._calculate_effort_score(indicator)

        # Avoid division by zero
        if effort_score == 0:
            effort_score = 0.1

        return impact_score / effort_score

    def _calculate_impact_score(self, indicator: DebtIndicator) -> float:
        """Calculate impact score (0-10)."""
        # Impact factors
        severity_weight = 0.4
        frequency_weight = 0.3
        complexity_weight = 0.3

        # Normalize values
        severity = indicator.severity / 10.0
        frequency = min(indicator.frequency / 10.0, 1.0)
        complexity = indicator.complexity / 10.0

        # Weighted score
        impact = (
            severity * severity_weight +
            frequency * frequency_weight +
            complexity * complexity_weight
        )

        return impact * 10.0

    def _calculate_effort_score(self, indicator: DebtIndicator) -> float:
        """Calculate effort score (0-10)."""
        effort_days = RefactoringEffortEstimator.EFFORT_MAP.get(
            indicator.type, 0.5
        ) * indicator.frequency

        # Normalize to 0-10 scale
        if effort_days < 1:
            return 2.0
        elif effort_days < 3:
            return 4.0
        elif effort_days < 7:
            return 6.0
        elif effort_days < 15:
            return 8.0
        else:
            return 10.0
```

---

## Report Generation

### Technical Debt Report Format

```python
@dataclass
class TechnicalDebtReport:
    """Technical debt report."""
    total_debt_score: float           # 0-100
    priority_actions: List[PriorityAction]
    estimated_effort: str             # "X.X days (Low/Medium/High)"
    debt_indicators: List[DebtIndicator]

    def generate_report(self) -> str:
        """Generate formatted report."""
        report = [
            "Technical Debt Report",
            "=" * 50,
            f"Total Debt Score: {self.total_debt_score:.1f} / 100",
            f"Estimated Refactoring Effort: {self.estimated_effort}",
            "",
            "Priority Actions:",
        ]

        for i, action in enumerate(self.priority_actions[:10], 1):
            priority_label = self._get_priority_label(action.priority)
            report.append(
                f"{i}. [{priority_label}] {action.indicator.type} "
                f"({action.indicator.frequency} instances) - "
                f"Impact: {action.impact_score:.1f}, "
                f"Effort: {action.effort_score:.1f} days"
            )

        return "\n".join(report)

    def _get_priority_label(self, priority: float) -> str:
        """Get priority label."""
        if priority > 2.0:
            return "HIGH"
        elif priority > 1.0:
            return "MEDIUM"
        else:
            return "LOW"
```

---

## Context7 Integration

### Debt Pattern Matching

```python
async def analyze_debt_with_context7(self, project_path: str):
    """Analyze technical debt with Context7 patterns."""
    # Get Context7 debt patterns (code smells and anti-patterns)
    debt_patterns = await self.context7.get_library_docs(
        context7_library_id="/refactoring/code-smells",
        topic="code smells technical debt patterns anti-patterns",
        tokens=3000
    )

    # AI-driven analysis
    ai_analysis = self.ai_analyzer.analyze_codebase(project_path)

    # Match patterns
    debt_indicators = self._correlate_debt_patterns(
        ai_analysis, debt_patterns
    )

    return debt_indicators
```

---

## Example Report

```
Technical Debt Report
==================================================
Total Debt Score: 68.5 / 100 (High)
Estimated Refactoring Effort: 12.5 days (Medium)

Priority Actions:
1. [HIGH] Duplicated Code (15 instances) - Impact: 8.5, Effort: 7.5 days
   - Extract common validation logic into UserValidator class
   - Consolidate duplicate database queries

2. [HIGH] Complex Methods (8 instances) - Impact: 7.2, Effort: 2.4 days
   - Extract method for process_order (82 lines → 4 methods)
   - Simplify calculate_discount (cyclomatic complexity: 12 → 4)

3. [MEDIUM] Large Classes (3 instances) - Impact: 6.8, Effort: 3.0 days
   - Split OrderManager into 4 focused classes

4. [LOW] Long Parameter Lists (5 instances) - Impact: 3.5, Effort: 1.0 days
   - Introduce Parameter Objects for create_user and update_profile
```

---

**Last Updated**: 2025-11-24
**Status**: Production Ready
