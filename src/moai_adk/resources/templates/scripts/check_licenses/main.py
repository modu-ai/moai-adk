#!/usr/bin/env python3
"""
Main entry point for check_licenses module
"""

import sys
from pathlib import Path
from .checker import LicenseChecker
from .reporter import generate_license_report, save_license_report
from .formatter import print_license_summary


def main():
    """Main entry point for license checking"""
    try:
        # Get project root
        if len(sys.argv) > 1:
            project_root = Path(sys.argv[1]).resolve()
        else:
            project_root = Path.cwd()

        if not project_root.exists():
            print(f"❌ Project directory not found: {project_root}")
            sys.exit(1)

        print(f"🔍 Checking license compliance in: {project_root}")

        # Initialize checker
        checker = LicenseChecker(project_root)

        # Run license check
        packages = checker.run_license_check()

        # Generate report
        report = generate_license_report(packages, checker.violations, checker.warnings)

        # Print summary
        print_license_summary(packages, checker.violations, checker.warnings)

        # Save report
        save_license_report(report, project_root)

        # Exit with appropriate code
        sys.exit(0 if not checker.has_violations() else 1)

    except Exception as error:
        print(f"❌ License check failed: {error}")
        sys.exit(1)