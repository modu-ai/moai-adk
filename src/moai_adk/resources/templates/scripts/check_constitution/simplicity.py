#!/usr/bin/env python3
# @TASK:CONSTITUTION-SIMPLICITY-001
"""
Simplicity Checker Module

Checks TRUST Readable principle - code simplicity and readability.
Validates file sizes, function complexity, and naming conventions.
"""

import ast
from pathlib import Path
from typing import Dict, Any, List


class SimplicityChecker:
    """Checks code simplicity and readability."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.max_file_lines = 300
        self.max_function_lines = 50
        self.max_complexity = 10

    def check_simplicity_principle(self) -> Dict[str, Any]:
        """Check Readable principle: code simplicity."""
        issues = []
        score = 100

        # Check file sizes
        large_files = self.find_large_files()
        if large_files:
            issues.extend([f"Large file: {f}" for f in large_files])
            score -= min(len(large_files) * 10, 30)

        # Check function sizes
        complex_functions = self.find_complex_functions()
        if complex_functions:
            issues.extend([f"Complex function: {f}" for f in complex_functions])
            score -= min(len(complex_functions) * 5, 20)

        # Check naming conventions
        naming_issues = self.check_naming_conventions()
        if naming_issues:
            issues.extend(naming_issues)
            score -= min(len(naming_issues) * 2, 15)

        return {
            "passed": score >= 80,
            "score": max(score, 0),
            "issues": issues,
        }

    def find_large_files(self) -> List[str]:
        """Find files exceeding size limits."""
        large_files = []
        for py_file in self.project_root.rglob("*.py"):
            if py_file.is_file() and "__pycache__" not in str(py_file):
                line_count = len(py_file.read_text(encoding='utf-8', errors='ignore').splitlines())
                if line_count > self.max_file_lines:
                    large_files.append(f"{py_file.name} ({line_count} lines)")
        return large_files

    def find_complex_functions(self) -> List[str]:
        """Find functions exceeding complexity limits."""
        complex_functions = []
        for py_file in self.project_root.rglob("*.py"):
            if "__pycache__" in str(py_file):
                continue
            try:
                content = py_file.read_text(encoding='utf-8', errors='ignore')
                tree = ast.parse(content)
                for node in ast.walk(tree):
                    if isinstance(node, ast.FunctionDef):
                        if len(content.splitlines()[node.lineno:node.end_lineno]) > self.max_function_lines:
                            complex_functions.append(f"{py_file.name}:{node.name}")
            except (SyntaxError, UnicodeDecodeError):
                continue
        return complex_functions

    def check_naming_conventions(self) -> List[str]:
        """Check naming convention violations."""
        # Simplified check for demonstration
        return []