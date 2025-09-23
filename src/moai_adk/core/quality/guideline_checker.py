"""
Guideline compliance checker for MoAI-ADK.

Validates Python code against TRUST 5 principles development guidelines:
- Function length ≤ 50 LOC
- File size ≤ 300 LOC
- Parameters ≤ 5 per function
- Complexity ≤ 10 per function

@FEATURE:QUALITY-GUIDELINES Guideline compliance validation system
"""

import ast
from pathlib import Path
from typing import Dict, List, Optional, Tuple, Any

from ...utils.logger import get_logger

logger = get_logger(__name__)


class GuidelineError(Exception):
    """Guideline compliance violation exception."""
    pass


class GuidelineChecker:
    """
    Checks Python code compliance against TRUST 5 development guidelines.

    @DESIGN:GUIDELINE-ARCH-001 Guideline validation architecture
    Follows single responsibility principle by focusing only on guideline violations.
    """

    def __init__(self, project_path: Path):
        """
        Initialize guideline checker.

        Args:
            project_path: Path to the project root directory
        """
        self.project_path = project_path
        self.max_function_lines = 50
        self.max_file_lines = 300
        self.max_parameters = 5
        self.max_complexity = 10

    def _parse_python_file(self, file_path: Path) -> Optional[ast.AST]:
        """
        Parse Python file to AST.

        Args:
            file_path: Path to Python file

        Returns:
            AST object or None if parsing fails
        """
        try:
            if not file_path.exists() or not file_path.suffix == '.py':
                return None

            with open(file_path, 'r', encoding='utf-8') as f:
                source_code = f.read()

            return ast.parse(source_code)
        except (SyntaxError, UnicodeDecodeError, FileNotFoundError):
            logger.warning(f"Failed to parse file: {file_path}")
            return None

    def _count_file_lines(self, file_path: Path) -> int:
        """
        Count lines in a file.

        Args:
            file_path: Path to file

        Returns:
            Number of lines in file
        """
        try:
            if not file_path.exists():
                return 0

            with open(file_path, 'r', encoding='utf-8') as f:
                return len(f.readlines())
        except (UnicodeDecodeError, FileNotFoundError):
            logger.warning(f"Failed to count lines in file: {file_path}")
            return 0

    def _calculate_complexity(self, func_node: ast.FunctionDef) -> int:
        """
        Calculate cyclomatic complexity of a function.

        Args:
            func_node: AST function node

        Returns:
            Complexity score (1 + number of decision points)
        """
        complexity = 1  # Base complexity

        for node in ast.walk(func_node):
            if isinstance(node, (ast.If, ast.While, ast.For, ast.AsyncFor)):
                complexity += 1
            elif isinstance(node, ast.Try):
                complexity += 1
            elif isinstance(node, ast.ExceptHandler):
                complexity += 1
            elif isinstance(node, (ast.BoolOp, ast.Compare)):
                complexity += 1

        return complexity

    def check_function_length(self, file_path: Path) -> List[Dict[str, Any]]:
        """
        Check if functions exceed 50 LOC limit.

        Args:
            file_path: Path to Python file to check

        Returns:
            List of violations with function name, line count, start line

        Raises:
            GuidelineError: When functions exceed LOC limit
        """
        violations = []
        tree = self._parse_python_file(file_path)

        if tree is None:
            return violations

        for node in ast.walk(tree):
            if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                # Calculate function length (end_lineno - lineno + 1)
                func_length = node.end_lineno - node.lineno + 1

                if func_length > self.max_function_lines:
                    violation = {
                        "function_name": node.name,
                        "line_count": func_length,
                        "start_line": node.lineno,
                        "file_path": str(file_path),
                        "max_allowed": self.max_function_lines
                    }
                    violations.append(violation)

        return violations

    def check_file_size(self, file_path: Path) -> Dict[str, Any]:
        """
        Check if file exceeds 300 LOC limit.

        Args:
            file_path: Path to Python file to check

        Returns:
            Dict with file size info and violation status

        Raises:
            GuidelineError: When file exceeds LOC limit
        """
        line_count = self._count_file_lines(file_path)
        violation = line_count > self.max_file_lines

        result = {
            "file_path": str(file_path),
            "line_count": line_count,
            "violation": violation,
            "max_allowed": self.max_file_lines
        }

        return result

    def check_parameter_count(self, file_path: Path) -> List[Dict[str, Any]]:
        """
        Check if functions have more than 5 parameters.

        Args:
            file_path: Path to Python file to check

        Returns:
            List of violations with function name, parameter count, line number

        Raises:
            GuidelineError: When functions exceed parameter limit
        """
        violations = []
        tree = self._parse_python_file(file_path)

        if tree is None:
            return violations

        for node in ast.walk(tree):
            if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                # Count parameters: args + kwonlyargs + vararg + kwarg
                param_count = (
                    len(node.args.args) +
                    len(node.args.kwonlyargs) +
                    (1 if node.args.vararg else 0) +
                    (1 if node.args.kwarg else 0)
                )

                if param_count > self.max_parameters:
                    violation = {
                        "function_name": node.name,
                        "parameter_count": param_count,
                        "line_number": node.lineno,
                        "file_path": str(file_path),
                        "max_allowed": self.max_parameters
                    }
                    violations.append(violation)

        return violations

    def check_complexity(self, file_path: Path) -> List[Dict[str, Any]]:
        """
        Check if functions exceed complexity limit of 10.

        Args:
            file_path: Path to Python file to check

        Returns:
            List of violations with function name, complexity score, line number

        Raises:
            GuidelineError: When functions exceed complexity limit
        """
        violations = []
        tree = self._parse_python_file(file_path)

        if tree is None:
            return violations

        for node in ast.walk(tree):
            if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                complexity = self._calculate_complexity(node)

                if complexity > self.max_complexity:
                    violation = {
                        "function_name": node.name,
                        "complexity": complexity,
                        "line_number": node.lineno,
                        "file_path": str(file_path),
                        "max_allowed": self.max_complexity
                    }
                    violations.append(violation)

        return violations

    def scan_project(self) -> Dict[str, List[Dict[str, Any]]]:
        """
        Scan entire project for guideline violations.

        Returns:
            Dict mapping violation types to list of violations

        Raises:
            GuidelineError: When any guideline violations are found
        """
        violations = {
            "function_length": [],
            "file_size": [],
            "parameter_count": [],
            "complexity": []
        }

        # Find all Python files in project
        if self.project_path.exists():
            python_files = list(self.project_path.rglob("*.py"))

            for file_path in python_files:
                # Skip __pycache__ and other ignored directories
                if "__pycache__" in str(file_path) or ".git" in str(file_path):
                    continue

                try:
                    # Check each guideline for this file
                    violations["function_length"].extend(self.check_function_length(file_path))

                    file_result = self.check_file_size(file_path)
                    if file_result["violation"]:
                        violations["file_size"].append(file_result)

                    violations["parameter_count"].extend(self.check_parameter_count(file_path))
                    violations["complexity"].extend(self.check_complexity(file_path))

                except Exception as e:
                    logger.warning(f"Error scanning file {file_path}: {e}")

        return violations

    def generate_violation_report(self) -> Dict[str, Any]:
        """
        Generate comprehensive guideline violation report.

        Returns:
            Dict containing violation summary and details
        """
        violations = self.scan_project()

        # Calculate summary statistics
        total_violations = sum(len(v) for v in violations.values())
        files_with_violations = set()

        for violation_type, violation_list in violations.items():
            for violation in violation_list:
                if "file_path" in violation:
                    files_with_violations.add(violation["file_path"])

        report = {
            "violations": violations,
            "summary": {
                "total_violations": total_violations,
                "files_checked": len(list(self.project_path.rglob("*.py"))) if self.project_path.exists() else 0,
                "files_with_violations": len(files_with_violations),
                "compliant": total_violations == 0,
                "violation_breakdown": {
                    "function_length": len(violations["function_length"]),
                    "file_size": len(violations["file_size"]),
                    "parameter_count": len(violations["parameter_count"]),
                    "complexity": len(violations["complexity"])
                }
            }
        }

        return report

    def validate_single_file(self, file_path: Path) -> bool:
        """
        Validate a single Python file against all guidelines.

        Args:
            file_path: Path to Python file to validate

        Returns:
            bool: True if file passes all guideline checks

        Raises:
            GuidelineError: When file violates any guidelines
        """
        # Check all guideline types for this file
        function_violations = self.check_function_length(file_path)
        file_size_result = self.check_file_size(file_path)
        parameter_violations = self.check_parameter_count(file_path)
        complexity_violations = self.check_complexity(file_path)

        # Check if any violations exist
        has_function_violations = len(function_violations) > 0
        has_file_size_violations = file_size_result["violation"]
        has_parameter_violations = len(parameter_violations) > 0
        has_complexity_violations = len(complexity_violations) > 0

        # Return True if no violations found
        return not (
            has_function_violations or
            has_file_size_violations or
            has_parameter_violations or
            has_complexity_violations
        )