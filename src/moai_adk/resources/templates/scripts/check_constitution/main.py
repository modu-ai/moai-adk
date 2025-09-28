#!/usr/bin/env python3
# @TASK:CONSTITUTION-MAIN-001
"""
Constitution Check Main Entry Point

Command-line interface for the modularized TRUST principles checker.
Provides backward compatibility with the original script.
"""

import argparse
from pathlib import Path
from .core import TrustPrinciplesChecker


def main():
    """Main entry point for TRUST principles verification."""
    parser = argparse.ArgumentParser(description="MoAI-ADK TRUST 5 Principles Verification")
    parser.add_argument(
        "--project-root",
        type=Path,
        default=Path.cwd(),
        help="Project root directory (default: current directory)"
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Enable verbose output"
    )
    parser.add_argument(
        "--report",
        type=Path,
        help="Generate report to specified file"
    )
    parser.add_argument(
        "--fix",
        action="store_true",
        help="Auto-fix issues where possible (placeholder)"
    )

    args = parser.parse_args()

    # Initialize checker
    checker = TrustPrinciplesChecker(args.project_root, args.verbose)

    # Run verification
    try:
        results = checker.run_full_check()

        # Generate and display report
        report = checker.generate_report(args.report)
        print(report)

        # Return appropriate exit code
        return 0 if results["overall"]["passed"] else 1

    except Exception as e:
        print(f"❌ Error during verification: {e}")
        return 1


if __name__ == "__main__":
    exit(main())