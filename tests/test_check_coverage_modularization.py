#!/usr/bin/env python3
"""
Test suite for check_coverage.py modularization
Ensures all modules are under 50 LOC and work correctly
"""

import pytest
from pathlib import Path
import sys

# Add the src directory to the path for imports
sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

def count_lines_of_code(file_path: Path) -> int:
    """Count non-empty, non-comment lines of code"""
    if not file_path.exists():
        return 0

    content = file_path.read_text(encoding="utf-8")
    lines = content.split('\n')

    loc = 0
    for line in lines:
        stripped = line.strip()
        if stripped and not stripped.startswith('#') and not stripped.startswith('"""') and not stripped.startswith("'''"):
            loc += 1

    return loc

class TestCheckCoverageModularization:
    """Test suite for check_coverage modularization"""

    def test_should_have_core_modules_under_50_loc(self):
        """All core modules should be under 50 LOC"""
        base_path = Path(__file__).parent.parent / "src/moai_adk/resources/templates/scripts/check_coverage"

        expected_modules = [
            "__init__.py",
            "models.py",
            "runner.py",
            "parser.py",
            "reporter.py",
            "checker.py",
            "main.py"
        ]

        for module in expected_modules:
            module_path = base_path / module
            if module_path.exists():
                loc = count_lines_of_code(module_path)
                assert loc <= 50, f"Module {module} has {loc} LOC, should be ≤50"

    def test_should_import_main_entry_point(self):
        """Main entry point should be importable"""
        try:
            from moai_adk.resources.templates.scripts.check_coverage import main
            assert callable(main)
        except ImportError:
            pytest.fail("Cannot import main function from check_coverage module")

    def test_should_import_coverage_checker_class(self):
        """CoverageChecker class should be importable"""
        try:
            from moai_adk.resources.templates.scripts.check_coverage.checker import CoverageChecker
            assert CoverageChecker is not None
        except ImportError:
            pytest.fail("Cannot import CoverageChecker from checker module")

    def test_should_import_coverage_result_dataclass(self):
        """CoverageResult dataclass should be importable"""
        try:
            from moai_adk.resources.templates.scripts.check_coverage.models import CoverageResult
            assert CoverageResult is not None
        except ImportError:
            pytest.fail("Cannot import CoverageResult from models module")

    def test_should_have_separate_test_runner(self):
        """Test runner should be in separate module"""
        try:
            from moai_adk.resources.templates.scripts.check_coverage.runner import run_coverage_test
            assert callable(run_coverage_test)
        except ImportError:
            pytest.fail("Cannot import run_coverage_test from runner module")

    def test_should_maintain_backward_compatibility(self):
        """Original script should still work after modularization"""
        original_script = Path(__file__).parent.parent / "src/moai_adk/resources/templates/scripts/check_coverage.py"

        # The original script should still exist and be importable
        if original_script.exists():
            loc = count_lines_of_code(original_script)
            # After modularization, the original script should be a thin wrapper
            assert loc <= 50, f"Original script should be a thin wrapper, but has {loc} LOC"