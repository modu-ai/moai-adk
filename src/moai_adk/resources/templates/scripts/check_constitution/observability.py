#!/usr/bin/env python3
# @TASK:CONSTITUTION-OBSERVABILITY-001
"""
Observability Checker Module

Checks TRUST Secured principle - logging, monitoring, and security.
Validates logging practices, security patterns, and error handling.
"""

from pathlib import Path
from typing import Dict, Any, List
import re


class ObservabilityChecker:
    """Checks observability and security practices."""

    def __init__(self, project_root: Path):
        self.project_root = project_root

    def check_observability_principle(self) -> Dict[str, Any]:
        """Check Secured principle: observability and security."""
        issues = []
        score = 100

        # Check logging practices
        logging_issues = self.check_logging_practices()
        if logging_issues:
            issues.extend(logging_issues)
            score -= min(len(logging_issues) * 10, 30)

        # Check error handling
        error_handling_issues = self.check_error_handling()
        if error_handling_issues:
            issues.extend(error_handling_issues)
            score -= min(len(error_handling_issues) * 5, 20)

        # Check security patterns
        security_issues = self.check_security_patterns()
        if security_issues:
            issues.extend(security_issues)
            score -= min(len(security_issues) * 8, 25)

        return {
            "passed": score >= 80,
            "score": max(score, 0),
            "issues": issues,
        }

    def check_logging_practices(self) -> List[str]:
        """Check logging implementation and practices."""
        issues = []

        # Check for logging usage
        has_logging = False
        for py_file in self.project_root.rglob("*.py"):
            if "__pycache__" in str(py_file):
                continue
            try:
                content = py_file.read_text(encoding='utf-8', errors='ignore')
                if re.search(r'import logging|from.*logging', content):
                    has_logging = True
                    break
            except (UnicodeDecodeError, PermissionError):
                continue

        if not has_logging:
            issues.append("No logging implementation found")

        return issues

    def check_error_handling(self) -> List[str]:
        """Check error handling patterns."""
        issues = []

        files_with_exceptions = 0
        total_py_files = 0

        for py_file in self.project_root.rglob("*.py"):
            if "__pycache__" in str(py_file):
                continue
            total_py_files += 1
            try:
                content = py_file.read_text(encoding='utf-8', errors='ignore')
                if re.search(r'try:|except:|raise', content):
                    files_with_exceptions += 1
            except (UnicodeDecodeError, PermissionError):
                continue

        if total_py_files > 0 and files_with_exceptions / total_py_files < 0.5:
            issues.append("Insufficient error handling coverage")

        return issues

    def check_security_patterns(self) -> List[str]:
        """Check security implementation patterns."""
        issues = []

        # Check for potential security issues (simplified)
        for py_file in self.project_root.rglob("*.py"):
            if "__pycache__" in str(py_file):
                continue
            try:
                content = py_file.read_text(encoding='utf-8', errors='ignore')

                # Check for hardcoded secrets
                if re.search(r'password\s*=\s*["\'][^"\']+["\']', content, re.IGNORECASE):
                    issues.append(f"Potential hardcoded password in {py_file.name}")

                # Check for SQL injection patterns
                if re.search(r'execute\s*\([^)]*%[^)]*\)', content):
                    issues.append(f"Potential SQL injection in {py_file.name}")

            except (UnicodeDecodeError, PermissionError):
                continue

        return issues