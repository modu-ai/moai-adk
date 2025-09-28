#!/usr/bin/env python3
# @TASK:CONSTITUTION-REPORTER-001
"""
Report Generator Module

Generates reports for TRUST principles verification results.
Formats output in various formats and provides actionable insights.
"""

from pathlib import Path
from typing import Dict, Any, Optional
from datetime import datetime


class ReportGenerator:
    """Generates verification reports in multiple formats."""

    def generate_report(self, results: Dict[str, Any], output_path: Optional[Path] = None) -> str:
        """Generate verification report."""
        report_lines = []

        # Header
        report_lines.append("=" * 60)
        report_lines.append("TRUST 5 Principles Verification Report")
        report_lines.append(f"Generated: {datetime.now().isoformat()}")
        report_lines.append("=" * 60)

        # Overall summary
        overall = results.get("overall", {})
        report_lines.append(f"\nOverall Score: {overall.get('score', 0):.1f}/100")
        report_lines.append(f"Status: {'✅ PASSED' if overall.get('passed', False) else '❌ FAILED'}")

        # Individual principles
        principles = {
            "test_first": "T - Test First",
            "readable": "R - Readable",
            "unified": "U - Unified",
            "secured": "S - Secured",
            "trackable": "T - Trackable",
        }

        for key, title in principles.items():
            if key in results:
                result = results[key]
                status = "✅" if result.get("passed", False) else "❌"
                score = result.get("score", 0)
                issues = result.get("issues", [])

                report_lines.append(f"\n{title}: {status} ({score:.1f}/100)")

                if issues:
                    report_lines.append("  Issues:")
                    for issue in issues[:5]:  # Limit to 5 issues per principle
                        report_lines.append(f"    - {issue}")
                    if len(issues) > 5:
                        report_lines.append(f"    ... and {len(issues) - 5} more")

        # Recommendations
        report_lines.append("\n" + "=" * 60)
        report_lines.append("RECOMMENDATIONS")
        report_lines.append("=" * 60)

        recommendations = self.generate_recommendations(results)
        for rec in recommendations:
            report_lines.append(f"• {rec}")

        report_content = "\n".join(report_lines)

        # Save to file if path provided
        if output_path:
            output_path.write_text(report_content, encoding='utf-8')

        return report_content

    def generate_recommendations(self, results: Dict[str, Any]) -> list[str]:
        """Generate actionable recommendations."""
        recommendations = []

        for principle, result in results.items():
            if principle == "overall" or not isinstance(result, dict):
                continue

            score = result.get("score", 0)
            if score < 80:
                recommendations.append(f"Improve {principle} compliance (current: {score:.1f}/100)")

        if not recommendations:
            recommendations.append("All TRUST principles are well implemented! 🎉")

        return recommendations