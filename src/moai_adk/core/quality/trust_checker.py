# # REMOVED_ORPHAN_CODE:TRUST-002 | SPEC: SPEC-TRUST-001/spec.md | TEST: tests/unit/core/quality/test_trust_checker.py
# type: ignore
"""
Integrated TRUST principle validation system

TRUST 4 principles:
- T: Test First (test coverage ≥85%)
- R: Readable (file ≤300 LOC, function ≤50 LOC, parameters ≤5)
- U: Unified (type safety)
- S: Secured (vulnerability scanning)
"""

import ast
import json
from pathlib import Path
from typing import Any

from moai_adk.core.quality.validators.base_validator import ValidationResult

# ========================================
# Constants (descriptive names)
# ========================================
MIN_TEST_COVERAGE_PERCENT = 85
MAX_FILE_LINES_OF_CODE = 300
MAX_FUNCTION_LINES_OF_CODE = 50
MAX_FUNCTION_PARAMETERS = 5
MAX_CYCLOMATIC_COMPLEXITY = 10

# File encoding
DEFAULT_FILE_ENCODING = "utf-8"

# Constants for validation


class TrustChecker:
    """Integrated TRUST principle validator"""

    def __init__(self):
        """Initialize TrustChecker"""
        self.results: dict[str, ValidationResult] = {}

    # ========================================
    # T: Test First - Coverage Validation
    # ========================================

    def validate_coverage(self, project_path: Path, coverage_data: dict[str, Any]) -> ValidationResult:
        """
        Validate test coverage (≥85%)

        Args:
            project_path: Project path
            coverage_data: Coverage data (total_coverage, low_coverage_files)

        Returns:
            ValidationResult: Validation result
        """
        total_coverage = coverage_data.get("total_coverage", 0)

        if total_coverage >= MIN_TEST_COVERAGE_PERCENT:
            return ValidationResult(
                passed=True,
                message=f"Test coverage: {total_coverage}% (Target: {MIN_TEST_COVERAGE_PERCENT}%)",
            )

        # Generate detailed information on failure
        low_files = coverage_data.get("low_coverage_files", [])
        details = f"Current coverage: {total_coverage}% (Target: {MIN_TEST_COVERAGE_PERCENT}%)\n"
        details += "Low coverage files:\n"
        for file_info in low_files:
            details += f"  - {file_info['file']}: {file_info['coverage']}%\n"
        details += "\nRecommended: Add more test cases to increase coverage."

        return ValidationResult(
            passed=False,
            message=f"Test coverage: {total_coverage}% (Target: {MIN_TEST_COVERAGE_PERCENT}%)",
            details=details,
        )

    # ========================================
    # R: Readable - Code Constraints
    # ========================================

    def validate_file_size(self, src_path: Path) -> ValidationResult:
        """
        Validate file size (≤300 LOC)

        Args:
            src_path: Source code directory path

        Returns:
            ValidationResult: Validation result
        """
        # Input validation (security)
        if not src_path.exists():
            return ValidationResult(
                passed=False,
                message=f"Source path does not exist: {src_path}",
                details="",
            )

        if not src_path.is_dir():
            return ValidationResult(
                passed=False,
                message=f"Source path is not a directory: {src_path}",
                details="",
            )

        violations = []

        for py_file in src_path.rglob("*.py"):
            # Apply guard clause (improves readability)
            if py_file.name.startswith("test_"):
                continue

            try:
                lines = py_file.read_text(encoding="utf-8").splitlines()
                loc = len(lines)

                if loc > MAX_FILE_LINES_OF_CODE:
                    violations.append(f"{py_file.name}: {loc} LOC (Limit: {MAX_FILE_LINES_OF_CODE})")
            except (UnicodeDecodeError, PermissionError):
                # Security: handle file access errors
                continue

        if not violations:
            return ValidationResult(passed=True, message="All files within 300 LOC")

        details = "Files exceeding 300 LOC:\n" + "\n".join(f"  - {v}" for v in violations)
        details += "\n\nRecommended: Refactor large files into smaller modules."

        return ValidationResult(
            passed=False,
            message=f"{len(violations)} files exceed 300 LOC",
            details=details,
        )

    def validate_function_size(self, src_path: Path) -> ValidationResult:
        """
        Validate function size (≤50 LOC)

        Args:
            src_path: Source code directory path

        Returns:
            ValidationResult: Validation result
        """
        violations = []

        for py_file in src_path.rglob("*.py"):
            if py_file.name.startswith("test_"):
                continue

            try:
                content = py_file.read_text()
                tree = ast.parse(content)
                lines = content.splitlines()

                for node in ast.walk(tree):
                    if isinstance(node, ast.FunctionDef):
                        # AST line numbers are 1-based
                        start_line = node.lineno
                        end_line = node.end_lineno if node.end_lineno else start_line  # type: ignore

                        # Compute actual function lines of code (decorators excluded)
                        func_lines = lines[start_line - 1 : end_line]
                        func_loc = len(func_lines)

                        if func_loc > MAX_FUNCTION_LINES_OF_CODE:
                            violations.append(
                                f"{py_file.name}::{node.name}(): {func_loc} LOC (Limit: {MAX_FUNCTION_LINES_OF_CODE})"
                            )
            except SyntaxError:
                continue

        if not violations:
            return ValidationResult(passed=True, message="All functions within 50 LOC")

        details = "Functions exceeding 50 LOC:\n" + "\n".join(f"  - {v}" for v in violations)
        details += "\n\nRecommended: Extract complex functions into smaller ones."

        return ValidationResult(
            passed=False,
            message=f"{len(violations)} functions exceed 50 LOC",
            details=details,
        )

    def validate_param_count(self, src_path: Path) -> ValidationResult:
        """
        Validate parameter count (≤5)

        Args:
            src_path: Source code directory path

        Returns:
            ValidationResult: Validation result
        """
        violations = []

        for py_file in src_path.rglob("*.py"):
            if py_file.name.startswith("test_"):
                continue

            try:
                tree = ast.parse(py_file.read_text())
                for node in ast.walk(tree):
                    if isinstance(node, ast.FunctionDef):
                        param_count = len(node.args.args)
                        if param_count > MAX_FUNCTION_PARAMETERS:
                            violations.append(
                                f"{py_file.name}::{node.name}(): {param_count} parameters "
                                f"(Limit: {MAX_FUNCTION_PARAMETERS})"
                            )
            except SyntaxError:
                continue

        if not violations:
            return ValidationResult(passed=True, message="All functions within 5 parameters")

        details = "Functions exceeding 5 parameters:\n" + "\n".join(f"  - {v}" for v in violations)
        details += "\n\nRecommended: Use data classes or parameter objects."

        return ValidationResult(
            passed=False,
            message=f"{len(violations)} functions exceed 5 parameters",
            details=details,
        )

    def validate_complexity(self, src_path: Path) -> ValidationResult:
        """
        Validate cyclomatic complexity (≤10)

        Args:
            src_path: Source code directory path

        Returns:
            ValidationResult: Validation result
        """
        violations = []

        for py_file in src_path.rglob("*.py"):
            if py_file.name.startswith("test_"):
                continue

            try:
                tree = ast.parse(py_file.read_text())
                for node in ast.walk(tree):
                    if isinstance(node, ast.FunctionDef):
                        complexity = self._calculate_complexity(node)
                        if complexity > MAX_CYCLOMATIC_COMPLEXITY:
                            violations.append(
                                f"{py_file.name}::{node.name}(): complexity {complexity} "
                                f"(Limit: {MAX_CYCLOMATIC_COMPLEXITY})"
                            )
            except SyntaxError:
                continue

        if not violations:
            return ValidationResult(passed=True, message="All functions within complexity 10")

        details = "Functions exceeding complexity 10:\n" + "\n".join(f"  - {v}" for v in violations)
        details += "\n\nRecommended: Simplify complex logic using guard clauses."

        return ValidationResult(
            passed=False,
            message=f"{len(violations)} functions exceed complexity 10",
            details=details,
        )

    def _calculate_complexity(self, node: ast.FunctionDef) -> int:
        """
        Calculate cyclomatic complexity (McCabe complexity)

        Args:
            node: Function AST node

        Returns:
            int: Cyclomatic complexity
        """
        complexity = 1
        for child in ast.walk(node):
            # Add 1 for each branching statement
            if isinstance(child, (ast.If, ast.While, ast.For, ast.ExceptHandler, ast.With)):
                complexity += 1
            # Add 1 for each and/or operator
            elif isinstance(child, ast.BoolOp):
                complexity += len(child.values) - 1
            # elif is already counted as ast.If, no extra handling needed
        return complexity

    # ========================================
    # T: Trackable - Code Traceability
    # ========================================
    # Tracking now handled through SPEC references

    # ========================================
    # Report Generation
    # ========================================

    def generate_report(self, results: dict[str, Any], format: str = "markdown") -> str:
        """
        Generate validation report

        Args:
            results: Validation result dictionary
            format: Report format ("markdown" or "json")

        Returns:
            str: Report string
        """
        if format == "json":
            return json.dumps(results, indent=2)

        # Markdown format
        report = "# TRUST Validation Report\n\n"

        for category, result in results.items():
            status = "✅ PASS" if result.get("passed", False) else "❌ FAIL"
            value = result.get("value", "N/A")
            # Add % suffix when the value is numeric
            if isinstance(value, (int, float)):
                value_str = f"{value}%"
            else:
                value_str = str(value)

            report += f"## {category.upper()}\n"
            report += f"**Status**: {status}\n"
            report += f"**Value**: {value_str}\n\n"

        return report

    # ========================================
    # Tool Selection
    # ========================================

    def select_tools(self, project_path: Path) -> dict[str, str]:
        """
        Automatically select tools by language

        Args:
            project_path: Project path

        Returns:
            dict[str, str]: Selected tool dictionary
        """
        config_path = project_path / ".moai" / "config" / "config.json"
        if not config_path.exists():
            return {
                "test_framework": "pytest",
                "coverage_tool": "coverage.py",
                "linter": "ruff",
                "type_checker": "mypy",
            }

        config = json.loads(config_path.read_text())
        language = config.get("project", {}).get("language", "python")

        if language == "python":
            return {
                "test_framework": "pytest",
                "coverage_tool": "coverage.py",
                "linter": "ruff",
                "type_checker": "mypy",
            }
        elif language == "typescript":
            return {
                "test_framework": "vitest",
                "linter": "biome",
                "type_checker": "tsc",
            }

        # Default (Python)
        return {
            "test_framework": "pytest",
            "coverage_tool": "coverage.py",
            "linter": "ruff",
            "type_checker": "mypy",
        }
