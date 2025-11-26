"""
Tests for version field presence in .moai/config.json after initialization

"""

import json
import tempfile
from pathlib import Path
from unittest.mock import patch

from moai_adk.core.project.initializer import ProjectInitializer


class TestVersionFieldInitialization:
    """Test version field presence in .moai/config.json after moai-adk init"""

    def test_version_field_present_after_fresh_initialization(self, tmp_path: Path) -> None:
        """
        GIVEN: A new project directory with no existing .moai structure
        WHEN: moai-adk init is executed (fresh initialization)
        THEN: .moai/config.json should contain moai.version field with actual version
        """
        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize(backup_enabled=False)

        # Should succeed
        assert result.success is True, f"Initialization failed with errors: {result.errors}"

        # Check that config.json exists
        config_path = tmp_path / ".moai" / "config.json"
        assert config_path.exists(), "Config file should exist after initialization"

        # Load and verify config
        with open(config_path, "r", encoding="utf-8") as f:
            config = json.load(f)

        # RED: These assertions will fail because the version field is not being set correctly
        assert "moai" in config, "Config should have 'moai' section"
        assert "version" in config["moai"], "moai section should have 'version' field"

        # Should not be the placeholder value
        version = config["moai"]["version"]
        assert version != "{{MOAI_VERSION}}", f"Version should not be placeholder, got {version}"
        assert len(version) > 0, "Version should not be empty"
        assert version != "unknown", "Version should not be 'unknown'"

    def test_version_field_present_after_reinitialization(self, tmp_path: Path) -> None:
        """
        GIVEN: An already initialized project with existing .moai/config.json
        WHEN: moai-adk init --reinit is executed (reinitialization)
        THEN: .moai/config.json should preserve moai.version field
        """
        # First initialization
        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize(backup_enabled=False)
        assert result.success is True, f"First initialization failed: {result.errors}"

        # Modify the config to add a version field
        config_path = tmp_path / ".moai" / "config.json"
        with open(config_path, "r", encoding="utf-8") as f:
            config = json.load(f)

        # Add a custom version to test preservation
        custom_version = "1.2.3-custom"
        config["moai"]["version"] = custom_version
        with open(config_path, "w", encoding="utf-8") as f:
            json.dump(config, f, indent=2, ensure_ascii=False)

        # Verify version was set
        with open(config_path, "r", encoding="utf-8") as f:
            modified_config = json.load(f)
        assert modified_config["moai"]["version"] == custom_version

        # Reinitialize
        result = initializer.initialize(reinit=True, backup_enabled=False)
        assert result.success is True, f"Reinitialization failed: {result.errors}"

        # Check that version is preserved
        with open(config_path, "r", encoding="utf-8") as f:
            final_config = json.load(f)

        # RED: This assertion will fail because reinitialization doesn't preserve the version field
        assert "moai" in final_config, "Config should have 'moai' section"
        assert "version" in final_config["moai"], "moai section should have 'version' field"
        assert (
            final_config["moai"]["version"] == custom_version
        ), f"Version should be preserved during reinitialization, expected {custom_version}, got {final_config['moai']['version']}"

    def test_version_field_correct_in_personal_mode(self, tmp_path: Path) -> None:
        """
        GIVEN: A new project initialized in personal mode
        WHEN: moai-adk init is executed with mode=personal
        THEN: .moai/config.json should contain correct version in moai.version
        """
        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize(mode="personal", backup_enabled=False)

        assert result.success is True, f"Initialization failed: {result.errors}"

        config_path = tmp_path / ".moai" / "config.json"
        with open(config_path, "r", encoding="utf-8") as f:
            config = json.load(f)

        # RED: This assertion will fail because the version field is not being set correctly
        assert "moai" in config, "Config should have 'moai' section"
        assert "version" in config["moai"], "moai section should have 'version' field"

        version = config["moai"]["version"]
        assert version != "{{MOAI_VERSION}}", "Version should not be placeholder"
        assert version != "unknown", "Version should not be 'unknown'"

    def test_version_field_correct_in_team_mode(self, tmp_path: Path) -> None:
        """
        GIVEN: A new project initialized in team mode
        WHEN: moai-adk init is executed with mode=team
        THEN: .moai/config.json should contain correct version in moai.version
        """
        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize(mode="team", backup_enabled=False)

        assert result.success is True, f"Initialization failed: {result.errors}"

        config_path = tmp_path / ".moai" / "config.json"
        with open(config_path, "r", encoding="utf-8") as f:
            config = json.load(f)

        # RED: This assertion will fail because the version field is not being set correctly
        assert "moai" in config, "Config should have 'moai' section"
        assert "version" in config["moai"], "moai section should have 'version' field"

        version = config["moai"]["version"]
        assert version != "{{MOAI_VERSION}}", "Version should not be placeholder"
        assert version != "unknown", "Version should not be 'unknown'"

    def test_version_field_missing_error_handling(self, tmp_path: Path) -> None:
        """
        GIVEN: A corrupted config.json without version field
        WHEN: VersionReader tries to read version
        THEN: Should handle gracefully and return 'unknown'
        """
        from moai_adk.statusline.version_reader import VersionReader

        # Create a config without version field
        config_path = tmp_path / ".moai" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        # Config without version field
        config_data = {"moai": {"update_check_frequency": "daily"}, "project": {"name": "TestProject"}}
        config_path.write_text(json.dumps(config_data, indent=2, ensure_ascii=False))

        # This should pass but return 'unknown' (existing behavior)
        reader = VersionReader()
        reader._config_path = config_path

        version = reader.get_version()

        # RED: This assertion will fail because the current implementation returns 'unknown'
        # when version is missing, but the requirement is to have the version field present
        assert version != "unknown", f"Version should be set, got '{version}'"

    def test_version_field_matches_package_version(self, tmp_path: Path) -> None:
        """
        GIVEN: A new project
        WHEN: moai-adk init is executed
        THEN: version in config should match package version
        """
        from moai_adk import __version__ as package_version

        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize(backup_enabled=False)

        assert result.success is True, f"Initialization failed: {result.errors}"

        config_path = tmp_path / ".moai" / "config.json"
        with open(config_path, "r", encoding="utf-8") as f:
            config = json.load(f)

        # RED: This assertion will fail because the version field is not being set to the package version
        assert "moai" in config, "Config should have 'moai' section"
        assert "version" in config["moai"], "moai section should have 'version' field"

        config_version = config["moai"]["version"]
        assert (
            config_version == package_version
        ), f"Config version {config_version} should match package version {package_version}"


class TestVersionFieldIntegration:
    """Integration tests for version field handling"""

    def test_complete_initialization_with_version_field(self) -> None:
        """
        GIVEN: Complete initialization workflow
        WHEN: All phases execute
        THEN: Final config should have version field properly set
        """
        with tempfile.TemporaryDirectory() as tmpdir:
            tmp_path = Path(tmpdir)
            initializer = ProjectInitializer(tmp_path)

            # Track all phases
            with (
                patch.object(initializer.executor, "execute_preparation_phase"),
                patch.object(initializer.executor, "execute_directory_phase"),
                patch.object(initializer.executor, "execute_resource_phase") as mock_res,
                patch.object(initializer.executor, "execute_configuration_phase") as mock_conf,
                patch.object(initializer.executor, "execute_validation_phase"),
            ):

                # Mock resource phase to return typical files
                mock_res.return_value = [".moai/", ".claude/"]
                mock_conf.return_value = [".moai/config.json"]

                result = initializer.initialize(backup_enabled=False)

                assert result.success is True, f"Initialization failed: {result.errors}"

                # Check that config file was created
                config_path = tmp_path / ".moai" / "config.json"
                assert config_path.exists(), "Config file should exist"

                # Load and verify config
                with open(config_path, "r", encoding="utf-8") as f:
                    config = json.load(f)

                # RED: These assertions will fail because the version field is not being set
                assert "moai" in config, "Config should have 'moai' section"
                assert "version" in config["moai"], "moai section should have 'version' field"

                version = config["moai"]["version"]
                assert version != "{{MOAI_VERSION}}", f"Version should not be placeholder, got {version}"
                assert version != "unknown", "Version should not be 'unknown'"

    def test_version_field_with_existing_config_preservation(self, tmp_path: Path) -> None:
        """
        GIVEN: Existing .moai/config.json with version field
        WHEN: Reinitialization is performed
        THEN: Existing version should be preserved, not overwritten with template
        """
        # Create existing config with custom version
        config_path = tmp_path / ".moai" / "config.json"
        config_path.parent.mkdir(parents=True, exist_ok=True)

        existing_config = {
            "moai": {"version": "2.0.0-custom", "update_check_frequency": "weekly"},
            "project": {"name": "ExistingProject"},
        }
        config_path.write_text(json.dumps(existing_config, indent=2, ensure_ascii=False))

        # Initialize project
        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize(reinit=True, backup_enabled=False)

        assert result.success is True, f"Reinitialization failed: {result.errors}"

        # Check that custom version is preserved
        with open(config_path, "r", encoding="utf-8") as f:
            final_config = json.load(f)

        # RED: This assertion will fail because reinitialization overwrites the version field
        assert (
            final_config["moai"]["version"] == "2.0.0-custom"
        ), f"Custom version should be preserved, got {final_config['moai']['version']}"
