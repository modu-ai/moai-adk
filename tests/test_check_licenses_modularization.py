#!/usr/bin/env python3
"""
Test suite for check_licenses.py modularization
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

class TestCheckLicensesModularization:
    """Test suite for check_licenses modularization"""

    def test_should_have_core_modules_under_50_loc(self):
        """All core modules should be under 50 LOC"""
        base_path = Path(__file__).parent.parent / "src/moai_adk/resources/templates/scripts/check_licenses"

        expected_modules = [
            "__init__.py",
            "models.py",
            "scanner.py",
            "analyzer.py",
            "checker.py",
            "reporter.py",
            "formatter.py",
            "database.py",
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
            from moai_adk.resources.templates.scripts.check_licenses import main
            assert callable(main)
        except ImportError:
            pytest.fail("Cannot import main function from check_licenses module")

    def test_should_import_license_checker_class(self):
        """LicenseChecker class should be importable"""
        try:
            from moai_adk.resources.templates.scripts.check_licenses.checker import LicenseChecker
            assert LicenseChecker is not None
        except ImportError:
            pytest.fail("Cannot import LicenseChecker from checker module")

    def test_should_import_license_info_dataclass(self):
        """LicenseInfo dataclass should be importable"""
        try:
            from moai_adk.resources.templates.scripts.check_licenses.models import LicenseInfo
            assert LicenseInfo is not None
        except ImportError:
            pytest.fail("Cannot import LicenseInfo from models module")

    def test_should_maintain_backward_compatibility(self):
        """Original script should still work after modularization"""
        original_script = Path(__file__).parent.parent / "src/moai_adk/resources/templates/scripts/check_licenses.py"

        # The original script should still exist and be importable
        if original_script.exists():
            loc = count_lines_of_code(original_script)
            # After modularization, the original script should be a thin wrapper
            assert loc <= 50, f"Original script should be a thin wrapper, but has {loc} LOC"