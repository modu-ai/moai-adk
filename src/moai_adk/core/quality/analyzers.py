"""
Code analysis utilities for guideline checking.

@FEATURE:QUALITY-ANALYZERS Core analysis functions for code validation
@DESIGN:SEPARATED-ANALYZERS-001 Extracted from oversized guideline_checker.py (761 LOC)
"""

import ast
from pathlib import Path
from typing import Dict, List, Optional, Any

from .constants import GuidelineLimits, ProjectPatterns, COMPLEXITY_NODES
from ...utils.logger import get_logger

logger = get_logger(__name__)


class CodeAnalyzer:
    """Analyzes Python code for TRUST principle compliance."""

    def __init__(self, limits: GuidelineLimits):
        """Initialize analyzer with guideline limits."""
        self.limits = limits

    def parse_python_file(self, file_path: Path) -> Optional[ast.AST]:
        """
        Parse Python file into AST.

        Args:
            file_path: Path to Python file

        Returns:
            AST tree or None if parsing failed
        """
        try:
            with open(file_path, 'r', encoding='utf-8') as file:
                content = file.read()

            # Check if file is empty or contains only whitespace
            if not content.strip():
                logger.warning(f"Empty file: {file_path}")
                return None

            return ast.parse(content, filename=str(file_path))

        except SyntaxError as e:
            logger.error(f"Syntax error in {file_path}: {e}")
            return None
        except UnicodeDecodeError as e:
            logger.error(f"Encoding error in {file_path}: {e}")
            return None
        except Exception as e:
            logger.error(f"Unexpected error parsing {file_path}: {e}")
            return None

    def count_file_lines(self, file_path: Path) -> int:
        """
        Count lines in a file, excluding empty lines and comments.

        Args:
            file_path: Path to file

        Returns:
            Number of significant lines
        """
        try:
            with open(file_path, 'r', encoding='utf-8') as file:
                lines = file.readlines()

            significant_lines = 0
            for line in lines:
                stripped = line.strip()
                # Skip empty lines and comment-only lines
                if stripped and not stripped.startswith('#'):
                    significant_lines += 1

            return significant_lines

        except Exception as e:
            logger.error(f"Error counting lines in {file_path}: {e}")
            return 0

    def calculate_complexity(self, func_node: ast.FunctionDef) -> int:
        """
        Calculate cyclomatic complexity of a function.

        Args:
            func_node: AST function node

        Returns:
            Complexity score
        """
        complexity = 1  # Base complexity

        for node in ast.walk(func_node):
            if type(node) in COMPLEXITY_NODES:
                complexity += 1

        return complexity

    def discover_python_files(self, project_path: Path) -> List[Path]:
        """
        Discover all Python files in project.

        Args:
            project_path: Root project path

        Returns:
            List of Python file paths
        """
        python_files = []

        try:
            for file_path in project_path.rglob('*.py'):
                # Skip excluded directories
                if any(excluded in file_path.parts for excluded in ProjectPatterns.EXCLUDED_DIRECTORIES):
                    continue

                # Skip minimal init files if configured
                if file_path.name in ProjectPatterns.EXCLUDED_FILES:
                    if self._is_minimal_init_file(file_path):
                        continue

                python_files.append(file_path)

        except Exception as e:
            logger.error(f"Error discovering Python files: {e}")

        return sorted(python_files)

    def _is_minimal_init_file(self, file_path: Path) -> bool:
        """
        Check if __init__.py file is minimal and should be skipped.

        Args:
            file_path: Path to __init__.py file

        Returns:
            True if file is minimal (< 10 significant lines)
        """
        if file_path.name != '__init__.py':
            return False

        try:
            line_count = self.count_file_lines(file_path)
            return line_count < self.limits.MIN_DOCSTRING_LENGTH

        except Exception as e:
            logger.warning(f"Error checking init file {file_path}: {e}")
            return False