"""Smoke tests for MoAI-ADK core functionality.

These tests MUST pass before any deployment. They verify that the
package can be imported and basic functionality is intact.

Run with: pytest -m smoke or pytest tests/test_smoke.py
"""

import pytest
from pathlib import Path


class TestSmokeCore:
    """Smoke tests for core MoAI-ADK functionality."""

    @pytest.mark.smoke
    def test_moai_adk_import(self):
        """Test that moai_adk can be imported."""
        import moai_adk
        assert moai_adk.__version__ is not None
        assert isinstance(moai_adk.__version__, str)

    @pytest.mark.smoke
    def test_cli_entry_point_exists(self):
        """Test that CLI entry point exists and is callable."""
        from moai_adk.__main__ import main
        assert callable(main)

    @pytest.mark.smoke
    def test_project_root_structure(self):
        """Test that essential project directories exist."""
        project_root = Path(__file__).parent.parent
        assert (project_root / "src" / "moai_adk").exists()
        assert (project_root / ".claude").exists()
        assert (project_root / ".moai").exists()

    @pytest.mark.critical
    def test_core_modules_importable(self):
        """Test that core modules can be imported."""
        from moai_adk.foundation import core
        from moai_adk.project import configuration
        from moai_adk.cli import commands
        # If we get here without ImportError, the test passes

    @pytest.mark.critical
    def test_template_directory_exists(self):
        """Test that template directory structure is intact."""
        import moai_adk
        moai_root = Path(moai_adk.__file__).parent.parent.parent / "templates"
        assert moai_root.exists()
        assert (moai_root / ".claude").exists()
        assert (moai_root / ".moai").exists()

    @pytest.mark.smoke
    def test_python_version_compatibility(self):
        """Test that Python version meets minimum requirements."""
        import sys
        from packaging import specifiers

        # MoAI-ADK requires Python >= 3.11
        spec = specifiers.SpecifierSet(">=3.11")
        assert sys.version_info.major >= 3
        assert f"{sys.version_info.major}.{sys.version_info.minor}" in spec
