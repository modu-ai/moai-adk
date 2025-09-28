#!/usr/bin/env python3
"""
Data models for check_coverage module
"""

from dataclasses import dataclass
from typing import List, Dict, Optional


@dataclass
class CoverageResult:
    """Coverage result structure"""
    total_statements: int
    covered_statements: int
    coverage_percentage: float
    missing_lines: Dict[str, List[int]]  # File-wise uncovered lines
    branch_coverage: Optional[float] = None


@dataclass
class FileCoverage:
    """Per-file coverage information"""
    file_path: str
    statements: int
    missing: int
    coverage: float
    missing_lines: List[int]