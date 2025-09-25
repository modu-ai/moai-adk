"""
Code validation functions for guideline checking.

@FEATURE:QUALITY-VALIDATORS Validation functions for TRUST 5 principles
@DESIGN:SEPARATED-VALIDATORS-001 Extracted from oversized guideline_checker.py (761 LOC)
"""

import ast
from pathlib import Path
from typing import Any

from ...utils.logger import get_logger
from .analyzers import CodeAnalyzer
from .constants import GuidelineLimits

logger = get_logger(__name__)


class GuidelineValidator:
    """Validates code against TRUST 5 principles guidelines."""

    def __init__(self, limits: GuidelineLimits):
        """Initialize validator with guideline limits."""
        self.limits = limits
        self.analyzer = CodeAnalyzer(limits)

    def check_function_length(self, file_path: Path) -> list[dict[str, Any]]:
        """
        Check if functions exceed maximum line limit.

        Args:
            file_path: Path to Python file

        Returns:
            List of function length violations
        """
        violations = []
        tree = self.analyzer.parse_python_file(file_path)

        if not tree:
            return violations

        try:
            for node in ast.walk(tree):
                if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                    function_lines = node.end_lineno - node.lineno + 1

                    if function_lines > self.limits.MAX_FUNCTION_LINES:
                        violations.append(
                            {
                                "type": "function_length",
                                "function_name": node.name,
                                "line_count": function_lines,
                                "limit": self.limits.MAX_FUNCTION_LINES,
                                "line_number": node.lineno,
                                "severity": "high"
                                if function_lines > self.limits.MAX_FUNCTION_LINES * 1.5
                                else "medium",
                            }
                        )

        except Exception as e:
            logger.error(f"Error checking function length in {file_path}: {e}")

        return violations

    def check_file_size(self, file_path: Path) -> dict[str, Any]:
        """
        Check if file exceeds maximum line limit.

        Args:
            file_path: Path to Python file

        Returns:
            File size violation details
        """
        line_count = self.analyzer.count_file_lines(file_path)

        violation = {
            "type": "file_size",
            "file_path": str(file_path),
            "line_count": line_count,
            "limit": self.limits.MAX_FILE_LINES,
            "violation": line_count > self.limits.MAX_FILE_LINES,
        }

        if violation["violation"]:
            violation["severity"] = (
                "critical" if line_count > self.limits.MAX_FILE_LINES * 2 else "high"
            )

        return violation

    def check_parameter_count(self, file_path: Path) -> list[dict[str, Any]]:
        """
        Check if functions have too many parameters.

        Args:
            file_path: Path to Python file

        Returns:
            List of parameter count violations
        """
        violations = []
        tree = self.analyzer.parse_python_file(file_path)

        if not tree:
            return violations

        try:
            for node in ast.walk(tree):
                if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                    param_count = len(node.args.args)

                    # Add keyword-only arguments
                    param_count += len(node.args.kwonlyargs)

                    # Add *args and **kwargs if present
                    if node.args.vararg:
                        param_count += 1
                    if node.args.kwarg:
                        param_count += 1

                    if param_count > self.limits.MAX_PARAMETERS:
                        violations.append(
                            {
                                "type": "parameter_count",
                                "function_name": node.name,
                                "parameter_count": param_count,
                                "limit": self.limits.MAX_PARAMETERS,
                                "line_number": node.lineno,
                                "severity": "high"
                                if param_count > self.limits.MAX_PARAMETERS * 1.5
                                else "medium",
                            }
                        )

        except Exception as e:
            logger.error(f"Error checking parameter count in {file_path}: {e}")

        return violations

    def check_complexity(self, file_path: Path) -> list[dict[str, Any]]:
        """
        Check if functions exceed maximum complexity.

        Args:
            file_path: Path to Python file

        Returns:
            List of complexity violations
        """
        violations = []
        tree = self.analyzer.parse_python_file(file_path)

        if not tree:
            return violations

        try:
            for node in ast.walk(tree):
                if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                    complexity = self.analyzer.calculate_complexity(node)

                    if complexity > self.limits.MAX_COMPLEXITY:
                        violations.append(
                            {
                                "type": "complexity",
                                "function_name": node.name,
                                "complexity": complexity,
                                "limit": self.limits.MAX_COMPLEXITY,
                                "line_number": node.lineno,
                                "severity": "critical"
                                if complexity > self.limits.MAX_COMPLEXITY * 2
                                else "high",
                            }
                        )

        except Exception as e:
            logger.error(f"Error checking complexity in {file_path}: {e}")

        return violations

    def validate_single_file(self, file_path: Path) -> bool:
        """
        Validate single file against all guidelines.

        Args:
            file_path: Path to Python file

        Returns:
            True if file passes all checks
        """
        try:
            # Check file size
            file_check = self.check_file_size(file_path)
            if file_check["violation"]:
                return False

            # Check function length
            function_violations = self.check_function_length(file_path)
            if function_violations:
                return False

            # Check parameter count
            param_violations = self.check_parameter_count(file_path)
            if param_violations:
                return False

            # Check complexity
            complexity_violations = self.check_complexity(file_path)
            if complexity_violations:
                return False

            return True

        except Exception as e:
            logger.error(f"Error validating file {file_path}: {e}")
            return False
