#!/usr/bin/env python3
# @TASK:CONSTITUTION-TESTING-001
"""
Testing Checker Module

Checks TRUST Test First principle - testing practices and coverage.
Validates test existence, coverage, and TDD compliance.
"""

from pathlib import Path
from typing import Dict, Any, List
import re


class TestingChecker:
    """Checks testing practices and TDD compliance."""

    def __init__(self, project_root: Path):
        self.project_root = project_root

    def check_testing_principle(self) -> Dict[str, Any]:
        """Check Test First principle: testing practices."""
        issues = []
        score = 100

        # Check test file existence
        test_coverage = self.check_test_coverage()
        if test_coverage < 80:
            issues.append(f"Low test coverage: {test_coverage}%")
            score -= 30

        # Check test patterns
        test_pattern_issues = self.check_test_patterns()
        if test_pattern_issues:
            issues.extend(test_pattern_issues)
            score -= min(len(test_pattern_issues) * 5, 20)

        # Check TDD evidence
        tdd_issues = self.check_tdd_evidence()
        if tdd_issues:
            issues.extend(tdd_issues)
            score -= min(len(tdd_issues) * 3, 15)

        return {
            "passed": score >= 80,
            "score": max(score, 0),
            "issues": issues,
        }

    def check_test_coverage(self) -> float:
        """Calculate test file coverage ratio."""
        src_files = list(self.project_root.rglob("*.py"))
        src_files = [f for f in src_files if "__pycache__" not in str(f) and "test_" not in f.name]

        test_files = list(self.project_root.rglob("test_*.py"))
        test_files.extend(list(self.project_root.rglob("*_test.py")))
        test_files.extend(list(self.project_root.rglob("tests/*.py")))

        if not src_files:
            return 100.0

        # Simplified coverage calculation
        coverage_ratio = min(len(test_files) / len(src_files), 1.0)
        return coverage_ratio * 100

    def check_test_patterns(self) -> List[str]:
        """Check test naming and structure patterns."""
        issues = []

        for test_file in self.project_root.rglob("test_*.py"):
            try:
                content = test_file.read_text(encoding='utf-8', errors='ignore')

                # Check for test functions
                test_functions = re.findall(r'def test_\w+', content)
                if len(test_functions) < 1:
                    issues.append(f"No test functions in {test_file.name}")

                # Check for assertions
                if not re.search(r'assert\s+', content):
                    issues.append(f"No assertions in {test_file.name}")

            except (UnicodeDecodeError, PermissionError):
                continue

        return issues

    def check_tdd_evidence(self) -> List[str]:
        """Check for TDD evidence in codebase."""
        issues = []

        # Check for test-first evidence (simplified)
        total_files = len(list(self.project_root.rglob("*.py")))
        test_files = len(list(self.project_root.rglob("test_*.py")))

        if total_files > 0 and test_files / total_files < 0.3:
            issues.append("Insufficient test files for TDD practice")

        return issues