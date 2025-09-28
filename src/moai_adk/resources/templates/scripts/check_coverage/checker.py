#!/usr/bin/env python3
"""
Coverage checker class for check_coverage module
"""

from pathlib import Path
from typing import Dict, Any, List
from .models import CoverageResult
from .config import load_coverage_config, detect_test_framework
from .runner import run_coverage_test


class CoverageChecker:
    """Test coverage checker"""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_dir = project_root / ".moai"

        # Load configuration
        self.config = load_coverage_config(self.moai_dir)

        # Results storage
        self.coverage_data = {}
        self.violations: List[str] = []

    def run_coverage_check(self) -> CoverageResult:
        """Run comprehensive coverage check"""
        print("🧪 Starting test coverage analysis...")

        # Detect test framework
        framework = detect_test_framework(self.project_root)
        print(f"  Detected test framework: {framework}")

        # Run coverage test
        if framework == "pytest":
            result = run_coverage_test(self.project_root, self.config)
        else:
            raise RuntimeError(f"Unsupported test framework: {framework}")

        # Validate coverage
        self.validate_coverage(result)

        return result

    def validate_coverage(self, result: CoverageResult) -> None:
        """Validate coverage against thresholds"""
        min_coverage = self.config["min_coverage"]

        if result.coverage_percentage < min_coverage:
            self.violations.append(
                f"Coverage {result.coverage_percentage:.1f}% below minimum {min_coverage}%"
            )

        if result.branch_coverage and result.branch_coverage < self.config["min_branch_coverage"]:
            self.violations.append(
                f"Branch coverage {result.branch_coverage:.1f}% below minimum {self.config['min_branch_coverage']}%"
            )

    def has_violations(self) -> bool:
        """Check if there are any violations"""
        return len(self.violations) > 0