#!/usr/bin/env python3
"""
MoAI-ADK Test Coverage Checker v0.1.12 - Modularized
Test coverage measurement and threshold validation
"""

from .main import main
from .models import CoverageResult, FileCoverage
from .checker import CoverageChecker

__all__ = ["main", "CoverageResult", "FileCoverage", "CoverageChecker"]


if __name__ == "__main__":
    main()