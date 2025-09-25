"""
Constants and configuration classes for guideline checking.

@FEATURE:QUALITY-CONSTANTS Core constants for TRUST 5 principles
@DESIGN:SEPARATED-CONSTANTS-001 Extracted from oversized guideline_checker.py (924 LOC)
"""

import ast
from dataclasses import dataclass


# @DESIGN:CONSTANTS-001 TRUST 5 principles guideline limits
@dataclass(frozen=True)
class GuidelineLimits:
    """TRUST 5 principles guideline limits as immutable configuration."""
    MAX_FUNCTION_LINES: int = 50
    MAX_FILE_LINES: int = 300
    MAX_PARAMETERS: int = 5
    MAX_COMPLEXITY: int = 10

    # Additional quality thresholds
    MIN_DOCSTRING_LENGTH: int = 10
    MAX_NESTING_DEPTH: int = 4


# @DESIGN:CONSTANTS-002 File patterns for project scanning
class ProjectPatterns:
    """File patterns and exclusions for project scanning."""
    PYTHON_EXTENSION = '.py'
    EXCLUDED_DIRECTORIES = {'__pycache__', '.git', '.pytest_cache', 'venv', '.venv', 'node_modules'}
    EXCLUDED_FILES = {'__init__.py'}  # Optional: exclude minimal init files


# @DESIGN:CONSTANTS-003 AST node types for complexity calculation
COMPLEXITY_NODES = {
    # Control flow nodes that add complexity
    ast.If, ast.While, ast.For, ast.AsyncFor, ast.Try, ast.ExceptHandler,
    # Boolean operations and comparisons
    ast.BoolOp, ast.Compare
}
