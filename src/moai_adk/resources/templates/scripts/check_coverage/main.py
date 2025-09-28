#!/usr/bin/env python3
"""
Main entry point for check_coverage module
"""

import sys
from pathlib import Path
from .checker import CoverageChecker
from .reporter import generate_coverage_report, print_coverage_summary, save_coverage_report


def main():
    """Main entry point for coverage checking"""
    try:
        # Get project root
        if len(sys.argv) > 1:
            project_root = Path(sys.argv[1]).resolve()
        else:
            project_root = Path.cwd()

        if not project_root.exists():
            print(f"❌ Project directory not found: {project_root}")
            sys.exit(1)

        print(f"🔍 Checking test coverage in: {project_root}")

        # Initialize checker
        checker = CoverageChecker(project_root)

        # Run coverage check
        result = checker.run_coverage_check()

        # Generate report
        report = generate_coverage_report(result, checker.violations)

        # Print summary
        print_coverage_summary(result, checker.violations)

        # Save report
        save_coverage_report(report, project_root)

        # Exit with appropriate code
        sys.exit(0 if not checker.has_violations() else 1)

    except Exception as error:
        print(f"❌ Coverage check failed: {error}")
        sys.exit(1)