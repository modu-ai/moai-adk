"""
🔴 RED: Tests for CLI init helpers with focus on SimplifiedInstaller config-based initialization.

@TEST:CONFIG-INIT-001 Test SimplifiedInstaller initialization with Config object
@TASK:INSTALLER-FIX-001 Fix parameter mismatch in finalize_installation function
"""

import pytest
from pathlib import Path
from unittest.mock import Mock, patch, MagicMock

from moai_adk.config import Config
from moai_adk.cli.init_helpers import finalize_installation
from moai_adk.install.installer import SimplifiedInstaller


class TestFinalizationWithConfig:
    """Test finalize_installation function with Config-based SimplifiedInstaller."""

    def test_finalize_installation_creates_config_object(self, tmp_path):
        """
        🔴 RED: Test that finalize_installation creates proper Config object for SimplifiedInstaller.

        This test should FAIL initially because finalize_installation currently passes
        individual parameters instead of Config object to SimplifiedInstaller.
        """
        project_dir = tmp_path / "test-project"
        project_dir.mkdir()

        # Mock SimplifiedInstaller to capture how it's instantiated
        with patch('moai_adk.cli.init_helpers.SimplifiedInstaller') as mock_installer_class:
            # Setup mock installer instance
            mock_installer = Mock()
            mock_result = Mock()
            mock_result.success = True
            mock_result.next_steps = []
            mock_installer.install.return_value = mock_result
            mock_installer_class.return_value = mock_installer

            # Mock create_mode_configuration to prevent file operations
            with patch('moai_adk.cli.init_helpers.create_mode_configuration') as mock_create_mode:
                # This should work after fix: SimplifiedInstaller should be called with Config object
                finalize_installation(
                    project_dir=project_dir,
                    project_mode="personal",
                    force_copy=True,
                    quiet=False
                )

        # Verify SimplifiedInstaller was called exactly once
        mock_installer_class.assert_called_once()

        # Get the arguments passed to SimplifiedInstaller.__init__
        args, kwargs = mock_installer_class.call_args

        # Should be called with Config object as first argument
        assert len(args) == 1, "SimplifiedInstaller should be called with one positional argument (Config)"

        config_arg = args[0]
        assert isinstance(config_arg, Config), f"Expected Config object, got {type(config_arg)}"

        # Verify Config object has correct properties
        assert config_arg.name == project_dir.name
        assert str(config_arg.project_path) == str(project_dir)
        assert config_arg.force_copy == True
        assert config_arg.silent == False  # quiet=False maps to silent=False

    def test_finalize_installation_maps_quiet_to_silent(self, tmp_path):
        """
        🔴 RED: Test that quiet parameter is correctly mapped to Config.silent.
        """
        project_dir = tmp_path / "test-project"
        project_dir.mkdir()

        with patch('moai_adk.cli.init_helpers.SimplifiedInstaller') as mock_installer_class:
            mock_installer = Mock()
            mock_result = Mock()
            mock_result.success = True
            mock_result.next_steps = []
            mock_installer.install.return_value = mock_result
            mock_installer_class.return_value = mock_installer

            with patch('moai_adk.cli.init_helpers.create_mode_configuration'):
                # Test quiet=True maps to silent=True
                finalize_installation(
                    project_dir=project_dir,
                    project_mode="team",
                    force_copy=False,
                    quiet=True
                )

        args, kwargs = mock_installer_class.call_args
        config_arg = args[0]

        assert isinstance(config_arg, Config)
        assert config_arg.silent == True, "quiet=True should map to Config.silent=True"
        assert config_arg.force_copy == False, "force_copy should be preserved"

    def test_finalize_installation_handles_installer_failure(self, tmp_path):
        """
        🔴 RED: Test that finalize_installation properly handles installer failures.
        """
        project_dir = tmp_path / "test-project"
        project_dir.mkdir()

        with patch('moai_adk.cli.init_helpers.SimplifiedInstaller') as mock_installer_class:
            mock_installer = Mock()
            mock_result = Mock()
            mock_result.success = False
            mock_result.errors = ["Installation failed", "Permission denied"]
            mock_installer.install.return_value = mock_result
            mock_installer_class.return_value = mock_installer

            with patch('moai_adk.cli.init_helpers.create_mode_configuration'):
                # Should raise SystemExit on failure
                with pytest.raises(SystemExit):
                    finalize_installation(
                        project_dir=project_dir,
                        project_mode="personal",
                        force_copy=False,
                        quiet=True  # quiet=True to suppress error output during test
                    )

    def test_config_backwards_compatibility_with_project_path(self):
        """
        🔴 RED: Test that Config handles project_path parameter for backward compatibility.
        """
        # This tests the existing Config class behavior
        test_path = "/tmp/test-project"

        # Test with project_path parameter (should be mapped to path)
        config = Config(name="test", project_path=test_path, force_copy=True)

        assert config.path == test_path
        assert str(config.project_path) == test_path
        assert config.force_copy == True


class TestConfigParameter:
    """Test Config object creation and parameter mapping."""

    def test_config_creation_with_all_parameters(self):
        """Test that Config object can be created with all required parameters."""
        config = Config(
            name="test-project",
            path="/tmp/test",
            force_copy=True,
            silent=True
        )

        assert config.name == "test-project"
        assert config.path == "/tmp/test"
        assert config.force_copy == True
        assert config.silent == True
        assert isinstance(config.project_path, Path)

    def test_simplified_installer_accepts_config(self):
        """Test that SimplifiedInstaller can be instantiated with Config object."""
        config = Config(name="test", path="/tmp/test")

        # This should not raise an exception
        try:
            # Just test instantiation - don't call install() to avoid side effects
            installer = SimplifiedInstaller(config)
            assert installer.config == config
        except Exception as e:
            pytest.fail(f"SimplifiedInstaller should accept Config object: {e}")