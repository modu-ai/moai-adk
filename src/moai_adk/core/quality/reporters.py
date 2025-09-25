"""
Report generation utilities for guideline checking.

@FEATURE:QUALITY-REPORTERS Report generation for TRUST 5 principles violations
@DESIGN:SEPARATED-REPORTERS-001 Extracted from oversized guideline_checker.py (761 LOC)
"""

from typing import Any

from ...utils.logger import get_logger
from .constants import GuidelineLimits

logger = get_logger(__name__)


class ViolationReporter:
    """Generates reports for guideline violations."""

    def __init__(self, limits: GuidelineLimits):
        """Initialize reporter with guideline limits."""
        self.limits = limits

    def generate_violation_report(
        self, violations: dict[str, list[dict[str, Any]]]
    ) -> dict[str, Any]:
        """
        Generate comprehensive violation report.

        Args:
            violations: Dictionary of violations by type

        Returns:
            Comprehensive report dictionary
        """
        total_files_scanned = self._count_scanned_files(violations)
        total_violations = self._count_total_violations(violations)

        report = {
            "summary": {
                "total_files_scanned": total_files_scanned,
                "total_violations": total_violations,
                "files_with_violations": self._count_files_with_violations(violations),
                "compliance_rate": self._calculate_compliance_rate(
                    total_files_scanned, violations
                ),
            },
            "violations_by_type": {
                "file_size": len(violations.get("file_size", [])),
                "function_length": len(violations.get("function_length", [])),
                "parameter_count": len(violations.get("parameter_count", [])),
                "complexity": len(violations.get("complexity", [])),
            },
            "severity_breakdown": self._get_severity_breakdown(violations),
            "worst_violations": self._get_worst_violations(violations),
            "guidelines": {
                "max_file_lines": self.limits.MAX_FILE_LINES,
                "max_function_lines": self.limits.MAX_FUNCTION_LINES,
                "max_parameters": self.limits.MAX_PARAMETERS,
                "max_complexity": self.limits.MAX_COMPLEXITY,
            },
            "recommendations": self._generate_recommendations(violations),
        }

        return report

    def _count_scanned_files(self, violations: dict[str, list[dict[str, Any]]]) -> int:
        """Count total number of files scanned."""
        files = set()
        for violation_list in violations.values():
            for violation in violation_list:
                if "file_path" in violation:
                    files.add(violation["file_path"])
        return len(files)

    def _count_total_violations(
        self, violations: dict[str, list[dict[str, Any]]]
    ) -> int:
        """Count total number of violations."""
        return sum(len(v_list) for v_list in violations.values())

    def _count_files_with_violations(
        self, violations: dict[str, list[dict[str, Any]]]
    ) -> int:
        """Count files that have at least one violation."""
        files_with_violations = set()
        for violation_list in violations.values():
            for violation in violation_list:
                if violation.get(
                    "violation", True
                ):  # Default True for actual violations
                    if "file_path" in violation:
                        files_with_violations.add(violation["file_path"])
        return len(files_with_violations)

    def _calculate_compliance_rate(
        self, total_files: int, violations: dict[str, list[dict[str, Any]]]
    ) -> float:
        """Calculate compliance rate as percentage."""
        if total_files == 0:
            return 100.0

        files_with_violations = self._count_files_with_violations(violations)
        compliant_files = total_files - files_with_violations
        return round((compliant_files / total_files) * 100, 2)

    def _get_severity_breakdown(
        self, violations: dict[str, list[dict[str, Any]]]
    ) -> dict[str, int]:
        """Get breakdown by severity level."""
        severity_count = {"critical": 0, "high": 0, "medium": 0, "low": 0}

        for violation_list in violations.values():
            for violation in violation_list:
                severity = violation.get("severity", "medium")
                if severity in severity_count:
                    severity_count[severity] += 1

        return severity_count

    def _get_worst_violations(
        self, violations: dict[str, list[dict[str, Any]]]
    ) -> dict[str, Any]:
        """Get the worst violations for each type."""
        worst = {}

        # Worst file size violation
        file_violations = violations.get("file_size", [])
        if file_violations:
            worst_file = max(file_violations, key=lambda x: x.get("line_count", 0))
            worst["largest_file"] = {
                "file_path": worst_file.get("file_path"),
                "line_count": worst_file.get("line_count"),
                "times_over_limit": round(
                    worst_file.get("line_count", 0) / self.limits.MAX_FILE_LINES, 2
                ),
            }

        # Worst function length violation
        func_violations = violations.get("function_length", [])
        if func_violations:
            worst_func = max(func_violations, key=lambda x: x.get("line_count", 0))
            worst["longest_function"] = {
                "function_name": worst_func.get("function_name"),
                "file_path": worst_func.get("file_path"),
                "line_count": worst_func.get("line_count"),
                "times_over_limit": round(
                    worst_func.get("line_count", 0) / self.limits.MAX_FUNCTION_LINES, 2
                ),
            }

        # Worst complexity violation
        complexity_violations = violations.get("complexity", [])
        if complexity_violations:
            worst_complex = max(
                complexity_violations, key=lambda x: x.get("complexity", 0)
            )
            worst["most_complex_function"] = {
                "function_name": worst_complex.get("function_name"),
                "file_path": worst_complex.get("file_path"),
                "complexity": worst_complex.get("complexity"),
                "times_over_limit": round(
                    worst_complex.get("complexity", 0) / self.limits.MAX_COMPLEXITY, 2
                ),
            }

        return worst

    def _generate_recommendations(
        self, violations: dict[str, list[dict[str, Any]]]
    ) -> list[str]:
        """Generate actionable recommendations based on violations."""
        recommendations = []

        file_violations = violations.get("file_size", [])
        if file_violations:
            recommendations.append(
                f"Split {len(file_violations)} large files into smaller modules. "
                "Consider extracting classes or related functions into separate files."
            )

        func_violations = violations.get("function_length", [])
        if func_violations:
            recommendations.append(
                f"Refactor {len(func_violations)} long functions. "
                "Break them into smaller, single-purpose functions."
            )

        param_violations = violations.get("parameter_count", [])
        if param_violations:
            recommendations.append(
                f"Reduce parameters in {len(param_violations)} functions. "
                "Consider using parameter objects or configuration classes."
            )

        complexity_violations = violations.get("complexity", [])
        if complexity_violations:
            recommendations.append(
                f"Simplify {len(complexity_violations)} complex functions. "
                "Reduce nested conditions and extract helper methods."
            )

        if not recommendations:
            recommendations.append(
                "All code complies with TRUST 5 principles guidelines!"
            )

        return recommendations
