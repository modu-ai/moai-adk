#!/usr/bin/env python3
"""
License report formatting for check_licenses module
"""

from typing import List
from .models import PackageLicense


def print_license_summary(
    packages: List[PackageLicense],
    violations: List[str],
    warnings: List[str]
) -> None:
    """Print license summary to console"""
    print(f"\n{'='*60}")
    print("🔒 LICENSE COMPLIANCE REPORT")
    print(f"{'='*60}")

    print(f"Total Packages: {len(packages)}")
    print(f"Violations: {len(violations)}")
    print(f"Warnings: {len(warnings)}")

    # Status determination
    if not violations:
        status = "✅ PASS" if not warnings else "🟡 PASS (with warnings)"
    else:
        status = "❌ FAIL"

    print(f"Status: {status}")

    # Print violations
    if violations:
        print(f"\n⚠️  Violations ({len(violations)}):")
        for violation in violations[:5]:  # Show first 5
            print(f"  • {violation}")
        if len(violations) > 5:
            print(f"  ... and {len(violations) - 5} more")

    # Print warnings
    if warnings:
        print(f"\n🔍 Warnings ({len(warnings)}):")
        for warning in warnings[:3]:  # Show first 3
            print(f"  • {warning}")
        if len(warnings) > 3:
            print(f"  ... and {len(warnings) - 3} more")