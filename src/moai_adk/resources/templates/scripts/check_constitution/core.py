#!/usr/bin/env python3
# @TASK:CONSTITUTION-CORE-001
"""
TRUST Principles Checker Core Module

Main orchestration class for TRUST principles verification.
Coordinates all principle checkers and maintains results.
"""

from pathlib import Path
from typing import Dict, Any
from .simplicity import SimplicityChecker
from .architecture import ArchitectureChecker
from .testing import TestingChecker
from .observability import ObservabilityChecker
from .versioning import VersioningChecker
from .reporter import ReportGenerator


class TrustPrinciplesChecker:
    """Main TRUST principles verification orchestrator."""

    def __init__(self, project_root: Path, verbose: bool = False):
        self.project_root = project_root
        self.verbose = verbose
        self.moai_dir = project_root / ".moai"

        # Initialize results structure
        self.results = {
            "test_first": {"passed": False, "score": 0, "issues": []},
            "readable": {"passed": False, "score": 0, "issues": []},
            "unified": {"passed": False, "score": 0, "issues": []},
            "secured": {"passed": False, "score": 0, "issues": []},
            "trackable": {"passed": False, "score": 0, "issues": []},
            "overall": {"passed": False, "score": 0, "compliance_level": ""},
        }

        # Initialize checkers
        self.simplicity = SimplicityChecker(project_root)
        self.architecture = ArchitectureChecker(project_root)
        self.testing = TestingChecker(project_root)
        self.observability = ObservabilityChecker(project_root)
        self.versioning = VersioningChecker(project_root)
        self.reporter = ReportGenerator()

    def run_full_check(self) -> Dict[str, Any]:
        """Run complete TRUST principles verification."""
        # Run all principle checks
        self.results["readable"] = self.simplicity.check_simplicity_principle()
        self.results["unified"] = self.architecture.check_architecture_principle()
        self.results["test_first"] = self.testing.check_testing_principle()
        self.results["secured"] = self.observability.check_observability_principle()
        self.results["trackable"] = self.versioning.check_versioning_principle()

        # Calculate overall score
        total_score = sum(r["score"] for r in self.results.values() if isinstance(r, dict) and "score" in r)
        self.results["overall"]["score"] = total_score / 5
        self.results["overall"]["passed"] = self.results["overall"]["score"] >= 80

        return self.results

    def generate_report(self, output_path: Path = None) -> str:
        """Generate verification report."""
        return self.reporter.generate_report(self.results, output_path)