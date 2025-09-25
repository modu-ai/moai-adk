"""
Quality management modules for MoAI-ADK.

This package contains quality assurance components:
- coverage_manager: Test coverage analysis and validation
- guideline_checker: TRUST 5 guideline compliance validation
- tdd_manager: TDD Red-Green-Refactor automation
- constitution_checker: Development guide compliance validation
- quality_gates: Automated quality gate enforcement

@FEATURE:QUALITY-001 Quality management system
"""

from .constitution_checker import ConstitutionChecker, ConstitutionError
from .coverage_manager import CoverageError, CoverageManager
from .guideline_checker import GuidelineChecker, GuidelineError
from .quality_gates import QualityGateError, QualityGates
from .tdd_manager import TDDError, TDDManager

__all__ = [
    "ConstitutionChecker",
    "ConstitutionError",
    "CoverageError",
    "CoverageManager",
    "GuidelineChecker",
    "GuidelineError",
    "QualityGateError",
    "QualityGates",
    "TDDError",
    "TDDManager",
]
