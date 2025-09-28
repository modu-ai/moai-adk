#!/usr/bin/env python3
"""
License reporting for check_licenses module
"""

import json
from datetime import datetime
from pathlib import Path
from typing import List, Dict, Any
from .models import PackageLicense


def generate_license_report(
    packages: List[PackageLicense],
    violations: List[str],
    warnings: List[str]
) -> Dict[str, Any]:
    """Generate comprehensive license report"""
    status = "PASS" if not violations else "FAIL"

    # Count by license type
    license_counts = {}
    risk_counts = {"low": 0, "medium": 0, "high": 0, "critical": 0}

    for package in packages:
        license_counts[package.license] = license_counts.get(package.license, 0) + 1
        if package.license_info:
            risk = package.license_info.risk_level
            risk_counts[risk] = risk_counts.get(risk, 0) + 1

    report = {
        "timestamp": datetime.now().isoformat(),
        "status": status,
        "summary": {
            "total_packages": len(packages),
            "violations": len(violations),
            "warnings": len(warnings),
            "license_counts": license_counts,
            "risk_distribution": risk_counts
        },
        "violations": violations,
        "warnings": warnings,
        "packages": [{
            "name": pkg.package, "version": pkg.version, "license": pkg.license,
            "status": pkg.status, "risk_level": pkg.license_info.risk_level if pkg.license_info else "unknown"
        } for pkg in packages]
    }

    return report


# Moved to formatter.py


def save_license_report(report: Dict[str, Any], project_root: Path) -> None:
    """Save license report to file"""
    try:
        reports_dir = project_root / ".moai" / "reports"
        reports_dir.mkdir(parents=True, exist_ok=True)

        report_file = reports_dir / "license_report.json"
        report_file.write_text(json.dumps(report, indent=2))
        print(f"\n📄 Report saved: {report_file}")

    except Exception as e:
        print(f"⚠️  Failed to save report: {e}")