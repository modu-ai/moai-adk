#!/usr/bin/env python3
# @TASK:CONSTITUTION-CHECK-011
"""
MoAI-ADK TRUST 5 Principles Verification System

Modularized TRUST principles checker following TRUST principles.
Each module has a single responsibility and ≤ 50 LOC.
"""

from .core import TrustPrinciplesChecker
from .simplicity import SimplicityChecker
from .architecture import ArchitectureChecker
from .testing import TestingChecker
from .observability import ObservabilityChecker
from .versioning import VersioningChecker
from .reporter import ReportGenerator

__all__ = [
    "TrustPrinciplesChecker",
    "SimplicityChecker",
    "ArchitectureChecker",
    "TestingChecker",
    "ObservabilityChecker",
    "VersioningChecker",
    "ReportGenerator",
]