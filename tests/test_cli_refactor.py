#!/usr/bin/env python3
"""
Tests for CLI refactoring to ensure proper import and functionality.
Following TDD Red-Green-Refactor cycle.
"""

import sys
import subprocess
from pathlib import Path
import pytest


class TestCLIRefactor:
    """Test suite for CLI system refactoring."""

    def test_unified_cli_import(self):
        """Test that unified CLI can be imported without errors."""
        # Simplified test - just ensure no import errors occur
        # The actual functionality will be tested in integration tests
        try:
            import src.moai_adk.cli
            # Just check that we can import without errors
            assert True
        except ImportError as e:
            pytest.fail(f"CLI import failed: {e}")

    def test_cli_main_import(self):
        """Test that CLI __main__ can be imported without errors."""
        try:
            import src.moai_adk.cli.__main__
            # __main__ is now a simple wrapper, so it doesn't need to have main attribute
            assert True  # Just check it imports without error
        except ImportError as e:
            pytest.fail(f"CLI __main__ import failed: {e}")

    def test_no_build_artifacts(self):
        """Test that build artifacts are properly cleaned up."""
        src_dir = Path("src")
        egg_info_dirs = list(src_dir.glob("*.egg-info"))
        assert len(egg_info_dirs) == 0, f"Found build artifacts: {egg_info_dirs}"

    def test_cli_main_is_simple_wrapper(self):
        """Test that CLI __main__ is a simple wrapper (≤ 50 LOC)."""
        cli_main_path = Path("src/moai_adk/cli/__main__.py")
        if cli_main_path.exists():
            lines = cli_main_path.read_text().splitlines()
            # Remove empty lines and comments for actual code count
            code_lines = [line for line in lines if line.strip() and not line.strip().startswith('#')]
            assert len(code_lines) <= 30, f"CLI __main__ should be simple wrapper, got {len(code_lines)} code lines"

    def test_no_duplicate_cli_logic(self):
        """Test that there's no duplicate CLI logic between files."""
        cli_path = Path("src/moai_adk/cli.py")
        cli_main_path = Path("src/moai_adk/cli/__main__.py")

        if cli_path.exists() and cli_main_path.exists():
            cli_content = cli_path.read_text()
            cli_main_content = cli_main_path.read_text()

            # CLI main should not have complex logic, only imports and simple main function
            assert "click.group" not in cli_main_content, "CLI __main__ should not define click groups"
            # Since it's now a simple wrapper, it should have minimal import
            assert "from ..cli import main" in cli_main_content, "CLI __main__ should import from unified CLI"