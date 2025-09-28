#!/usr/bin/env python3
"""
Coverage reporting for check_coverage module
"""

from datetime import datetime
from pathlib import Path
from typing import Dict, Any
from .models import CoverageResult


def generate_coverage_report(result: CoverageResult, violations: list) -> Dict[str, Any]:
    """Generate comprehensive coverage report"""
    status = "PASS" if not violations else "FAIL"

    report = {
        "timestamp": datetime.now().isoformat(),
        "status": status,
        "summary": {
            "total_statements": result.total_statements,
            "covered_statements": result.covered_statements,
            "coverage_percentage": round(result.coverage_percentage, 2),
            "branch_coverage": round(result.branch_coverage, 2) if result.branch_coverage else None
        },
        "violations": violations,
        "missing_coverage": {
            file_path: lines for file_path, lines in result.missing_lines.items()
        } if result.missing_lines else {}
    }

    return report


def print_coverage_summary(result: CoverageResult, violations: list) -> None:
    """Print coverage summary to console"""
    print(f"\n{'='*60}")
    print("🧪 TEST COVERAGE REPORT")
    print(f"{'='*60}")

    print(f"Total Statements: {result.total_statements}")
    print(f"Covered Statements: {result.covered_statements}")
    print(f"Coverage: {result.coverage_percentage:.1f}%")

    if result.branch_coverage:
        print(f"Branch Coverage: {result.branch_coverage:.1f}%")

    # Status determination
    if not violations:
        status = "✅ PASS"
    else:
        status = "❌ FAIL"

    print(f"Status: {status}")

    # Print violations
    if violations:
        print(f"\n⚠️  Violations ({len(violations)}):")
        for violation in violations:
            print(f"  • {violation}")


def save_coverage_report(report: Dict[str, Any], project_root: Path) -> None:
    """Save coverage report to file"""
    try:
        reports_dir = project_root / ".moai" / "reports"
        reports_dir.mkdir(parents=True, exist_ok=True)

        report_file = reports_dir / "coverage_report.json"

        import json
        report_file.write_text(json.dumps(report, indent=2))
        print(f"\n📄 Report saved: {report_file}")

    except Exception as e:
        print(f"⚠️  Failed to save report: {e}")