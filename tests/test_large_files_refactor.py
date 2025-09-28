#!/usr/bin/env python3
"""
Tests for large files refactoring to ensure TRUST principles compliance.
Following TDD Red-Green-Refactor cycle.
"""

import os
from pathlib import Path
import pytest


class TestLargeFilesRefactor:
    """Test suite for large files refactoring to comply with TRUST principles."""

    # TRUST principle: Files should be ≤ 50 LOC for optimal readability
    MAX_LOC_PER_FILE = 50
    MAX_FUNCTIONS_PER_FILE = 5
    MAX_PARAMETERS_PER_FUNCTION = 5

    def count_lines_of_code(self, file_path: Path) -> int:
        """Count actual lines of code (excluding comments and empty lines)."""
        if not file_path.exists():
            return 0

        with open(file_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        code_lines = 0
        for line in lines:
            stripped = line.strip()
            # Skip empty lines and comments
            if stripped and not stripped.startswith('#') and not stripped.startswith('"""') and not stripped.startswith("'''"):
                code_lines += 1

        return code_lines

    def test_repair_tags_script_is_modularized(self):
        """Test that repair_tags.py is properly modularized."""
        original_file = Path("src/moai_adk/resources/templates/scripts/repair_tags.py")

        if original_file.exists():
            loc = self.count_lines_of_code(original_file)
            # This should fail initially, forcing us to modularize
            assert loc <= self.MAX_LOC_PER_FILE, f"repair_tags.py has {loc} LOC, should be ≤ {self.MAX_LOC_PER_FILE}"

    def test_check_constitution_script_is_modularized(self):
        """Test that check_constitution.py is properly modularized."""
        original_file = Path("src/moai_adk/resources/templates/scripts/check_constitution.py")

        if original_file.exists():
            loc = self.count_lines_of_code(original_file)
            assert loc <= self.MAX_LOC_PER_FILE, f"check_constitution.py has {loc} LOC, should be ≤ {self.MAX_LOC_PER_FILE}"

    def test_validator_module_is_modularized(self):
        """Test that core validator.py is properly modularized."""
        original_file = Path("src/moai_adk/core/validator.py")

        if original_file.exists():
            loc = self.count_lines_of_code(original_file)
            assert loc <= self.MAX_LOC_PER_FILE, f"validator.py has {loc} LOC, should be ≤ {self.MAX_LOC_PER_FILE}"

    def test_validate_tags_script_is_modularized(self):
        """Test that validate_tags.py is properly modularized."""
        original_file = Path("src/moai_adk/resources/templates/scripts/validate_tags.py")

        if original_file.exists():
            loc = self.count_lines_of_code(original_file)
            assert loc <= self.MAX_LOC_PER_FILE, f"validate_tags.py has {loc} LOC, should be ≤ {self.MAX_LOC_PER_FILE}"

    def test_check_coverage_script_is_modularized(self):
        """Test that check_coverage.py is properly modularized."""
        original_file = Path("src/moai_adk/resources/templates/scripts/check_coverage.py")

        if original_file.exists():
            loc = self.count_lines_of_code(original_file)
            assert loc <= self.MAX_LOC_PER_FILE, f"check_coverage.py has {loc} LOC, should be ≤ {self.MAX_LOC_PER_FILE}"

    def test_git_strategy_module_is_modularized(self):
        """Test that git_strategy.py is properly modularized."""
        original_file = Path("src/moai_adk/core/git_strategy.py")

        if original_file.exists():
            loc = self.count_lines_of_code(original_file)
            assert loc <= self.MAX_LOC_PER_FILE, f"git_strategy.py has {loc} LOC, should be ≤ {self.MAX_LOC_PER_FILE}"

    def test_check_licenses_script_is_modularized(self):
        """Test that check_licenses.py is properly modularized."""
        original_file = Path("src/moai_adk/resources/templates/scripts/check_licenses.py")

        if original_file.exists():
            loc = self.count_lines_of_code(original_file)
            assert loc <= self.MAX_LOC_PER_FILE, f"check_licenses.py has {loc} LOC, should be ≤ {self.MAX_LOC_PER_FILE}"

    def test_no_files_exceed_trust_limits(self):
        """Test that all Python files comply with TRUST principles."""
        src_dir = Path("src/moai_adk")
        violations = []

        for py_file in src_dir.rglob("*.py"):
            # Skip __pycache__ and test files
            if "__pycache__" in str(py_file) or "test_" in py_file.name:
                continue

            loc = self.count_lines_of_code(py_file)
            if loc > self.MAX_LOC_PER_FILE:
                violations.append(f"{py_file}: {loc} LOC")

        if violations:
            violation_list = "\n".join(violations)
            pytest.fail(f"Files violating TRUST principles (≤{self.MAX_LOC_PER_FILE} LOC):\n{violation_list}")

    def test_modularized_files_have_clear_interfaces(self):
        """Test that modularized files have clear public interfaces."""
        # This test will be expanded as we create modules
        # For now, just check that __init__.py files exist for major modules

        expected_modules = [
            "src/moai_adk/core/validator",
            "src/moai_adk/resources/templates/scripts/repair_tags",
            "src/moai_adk/resources/templates/scripts/check_constitution",
        ]

        for module_path in expected_modules:
            module_dir = Path(module_path)
            if module_dir.exists() and module_dir.is_dir():
                init_file = module_dir / "__init__.py"
                assert init_file.exists(), f"Module {module_path} should have __init__.py"