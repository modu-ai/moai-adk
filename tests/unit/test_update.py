"""Unit tests for update.py command

Tests for update command with various scenarios.
"""

from pathlib import Path
from unittest.mock import Mock, patch

from click.testing import CliRunner

from moai_adk.cli.commands.update import update


class TestUpdateCommand:
    """Test update command"""

    def test_update_help(self):
        """Test update --help"""
        runner = CliRunner()
        result = runner.invoke(update, ["--help"])
        assert result.exit_code == 0
        assert "Update command with 3-stage workflow" in result.output

    def test_update_not_initialized(self, tmp_path):
        """Test update when project is not initialized"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            result = runner.invoke(update)
            assert result.exit_code != 0
            assert "not initialized" in result.output

    def test_update_check_only(self, tmp_path):
        """Test update --check flag"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai directory
            Path(".moai").mkdir()

            result = runner.invoke(update, ["--check"])
            assert result.exit_code == 0
            assert "Checking versions" in result.output
            # Could be any of these messages depending on version comparison
            assert (
                "Already up to date" in result.output
                or "Update available" in result.output
                or "Dev version" in result.output
            )

    def test_update_check_when_update_available(self, tmp_path):
        """Test update --check when new version is available"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai directory
            Path(".moai").mkdir()

            # Mock get_latest_version to return a newer version
            with patch("moai_adk.cli.commands.update.get_latest_version") as mock_get_version:
                with patch("moai_adk.cli.commands.update.__version__", "0.3.0"):
                    mock_get_version.return_value = "0.3.5"  # Newer version available
                    result = runner.invoke(update, ["--check"])
                    assert result.exit_code == 0
                    assert "Checking versions" in result.output
                    assert "Update available" in result.output

    def test_update_with_backup(self, tmp_path):
        """Test update with backup (default behavior) - uses --force to skip version check"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai structure
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            import json

            config_data = {
                "moai": {"version": "0.6.1"},
                "project": {"template_version": "0.6.0", "optimized": True, "mode": "personal"},
            }
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            # Mock TemplateProcessor and version functions
            with (
                patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor,
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
                patch("moai_adk.cli.commands.update._get_project_config_version") as mock_proj_ver,
            ):

                mock_instance = Mock()
                # Return absolute path instead of relative
                mock_instance.create_backup.return_value = Path.cwd() / ".moai-backups/backup-2025-10-15"
                mock_instance.copy_templates.return_value = None  # Mock copy_templates
                mock_processor.return_value = mock_instance

                # Mock version functions for --force
                mock_current.return_value = "0.6.1"
                mock_latest.return_value = "0.6.1"
                mock_pkg_ver.return_value = "0.6.1"
                mock_proj_ver.return_value = "0.6.0"

                # Use --force to skip version check and test backup process
                result = runner.invoke(update, ["--force"])
                # Should show skip backup message with --force, and show syncing templates
                assert "Skipping backup (--force)" in result.output
                assert "Syncing templates" in result.output
                assert result.exit_code == 0

    def test_update_with_force_flag(self, tmp_path):
        """Test update --force flag (skip backup)"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai structure with template_version
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            import json

            config_data = {"moai": {"version": "0.6.1"}, "project": {"template_version": "0.6.0", "mode": "personal"}}
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            # Mock TemplateProcessor and version functions
            with (
                patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor,
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
                patch("moai_adk.cli.commands.update._get_project_config_version") as mock_proj_ver,
            ):

                mock_instance = Mock()
                mock_instance.copy_templates.return_value = None  # Mock copy_templates
                mock_processor.return_value = mock_instance

                mock_current.return_value = "0.6.1"
                mock_latest.return_value = "0.6.1"
                mock_pkg_ver.return_value = "0.6.1"
                mock_proj_ver.return_value = "0.6.0"

                result = runner.invoke(update, ["--force"])
                assert "Skipping backup (--force)" in result.output
                assert "Syncing templates" in result.output
                # create_backup should NOT be called with --force
                mock_instance.create_backup.assert_not_called()
                assert result.exit_code == 0

    def test_update_with_custom_path(self, tmp_path):
        """Test update with custom --path - uses --force to skip version check"""
        runner = CliRunner()

        # Create project directory
        project_dir = tmp_path / "my-project"
        project_dir.mkdir()
        (project_dir / ".moai").mkdir()
        import json

        config_data = {
            "moai": {"version": "0.6.1"},
            "project": {"template_version": "0.6.0", "optimized": True, "mode": "personal"},
        }
        (project_dir / ".moai" / "config.json").write_text(json.dumps(config_data))

        # Mock TemplateProcessor and version functions
        with (
            patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor,
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
            patch("moai_adk.cli.commands.update._get_project_config_version") as mock_proj_ver,
        ):

            mock_instance = Mock()
            mock_instance.create_backup.return_value = project_dir / ".moai-backups/backup"
            mock_processor.return_value = mock_instance

            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.1"
            mock_pkg_ver.return_value = "0.6.1"
            mock_proj_ver.return_value = "0.6.0"

            # Use --force to skip version check
            result = runner.invoke(update, ["--path", str(project_dir), "--force"])
            assert result.exit_code == 0
            assert "Update complete" in result.output

    def test_update_shows_version_info(self, tmp_path):
        """Test that update shows version information"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            Path(".moai").mkdir()

            result = runner.invoke(update, ["--check"])
            assert result.exit_code == 0
            assert "Current version" in result.output
            assert "Latest version" in result.output

    def test_update_template_processor_called(self, tmp_path):
        """Test that TemplateProcessor methods are called correctly - uses --force"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            import json

            config_data = {
                "moai": {"version": "0.6.1"},
                "project": {"template_version": "0.6.0", "optimized": True, "mode": "personal"},
            }
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            with (
                patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor,
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
                patch("moai_adk.cli.commands.update._get_project_config_version") as mock_proj_ver,
            ):

                mock_instance = Mock()
                mock_instance.create_backup.return_value = Path.cwd() / ".moai-backups/backup"
                mock_instance.copy_templates.return_value = None  # Mock copy_templates
                mock_processor.return_value = mock_instance

                mock_current.return_value = "0.6.1"
                mock_latest.return_value = "0.6.1"
                mock_pkg_ver.return_value = "0.6.1"
                mock_proj_ver.return_value = "0.6.0"

                # Use --force to skip version check and backup
                result = runner.invoke(update, ["--force"])

                # Verify methods were called (backup NOT called with --force)
                mock_instance.create_backup.assert_not_called()
                mock_instance.copy_templates.assert_called_once_with(backup=False, silent=True)
                assert result.exit_code == 0

    def test_update_handles_exception(self, tmp_path):
        """Test update handles exceptions - uses --force to skip version check"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            moai_dir = Path(".moai")
            moai_dir.mkdir()

            # Mock TemplateProcessor to raise exception
            with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
                mock_processor.side_effect = RuntimeError("Mock error")

                # Use --force to skip version check
                result = runner.invoke(update, ["--force"])
                assert result.exit_code != 0
                assert "Update failed" in result.output

    def test_update_shows_update_details(self, tmp_path):
        """Test that update shows detailed update information - uses --force"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            import json

            config_data = {
                "moai": {"version": "0.6.1"},
                "project": {"template_version": "0.6.0", "optimized": True, "mode": "personal"},
            }
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            with (
                patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor,
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
                patch("moai_adk.cli.commands.update._get_project_config_version") as mock_proj_ver,
            ):

                mock_instance = Mock()
                mock_instance.create_backup.return_value = Path.cwd() / ".moai-backups/backup"
                mock_instance.copy_templates.return_value = None  # Mock copy_templates
                mock_processor.return_value = mock_instance

                mock_current.return_value = "0.6.1"
                mock_latest.return_value = "0.6.1"
                mock_pkg_ver.return_value = "0.6.1"
                mock_proj_ver.return_value = "0.6.0"

                # Use --force to skip version check and backup
                result = runner.invoke(update, ["--force"])
                assert ".claude/ update complete" in result.output
                assert ".moai/ update complete" in result.output
                assert "CLAUDE.md merge complete" in result.output
                assert "config.json merge complete" in result.output
                assert result.exit_code == 0

    def test_update_skips_when_same_version_and_optimized(self, tmp_path):
        """Test update skips template sync when version is same and template_version matches"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai structure with matching template_version
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            config_data = {
                "moai": {"version": "0.6.1"},
                "project": {
                    "template_version": "0.6.1",  # Same as package
                    "optimized": True,
                    "name": "test",
                    "mode": "personal",
                },
            }
            import json

            (moai_dir / "config.json").write_text(json.dumps(config_data))

            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
                patch("moai_adk.cli.commands.update._get_project_config_version") as mock_proj_ver,
            ):

                mock_current.return_value = "0.6.1"
                mock_latest.return_value = "0.6.1"
                mock_pkg_ver.return_value = "0.6.1"
                mock_proj_ver.return_value = "0.6.1"

                result = runner.invoke(update)
                assert result.exit_code == 0
                # Should skip template sync when versions match
                assert "Checking versions" in result.output
                assert "Comparing config versions" in result.output
                assert "Templates are up to date" in result.output

    def test_update_suggests_alfred_when_same_version_not_optimized(self, tmp_path):
        """Test update syncs templates when template version is outdated (3-stage workflow Stage 3)"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai structure with outdated template_version
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            config_data = {
                "moai": {"version": "0.3.2"},
                "project": {
                    "template_version": "0.3.0",  # Older than current package
                    "optimized": False,
                    "name": "test",
                    "mode": "personal",
                },
            }
            import json

            (moai_dir / "config.json").write_text(json.dumps(config_data))

            # Mock version functions to return same package version but outdated template
            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
                patch("moai_adk.cli.commands.update._get_project_config_version") as mock_proj_ver,
                patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            ):

                mock_current.return_value = "0.3.2"
                mock_latest.return_value = "0.3.2"
                mock_pkg_ver.return_value = "0.3.2"
                mock_proj_ver.return_value = "0.3.0"
                mock_sync.return_value = True

                result = runner.invoke(update)
                assert result.exit_code == 0
                # In 3-stage workflow, outdated template version goes to Stage 3 (template sync)
                assert "Syncing templates" in result.output
                assert "alfred:0-project update" in result.output

    def test_update_proceeds_when_config_missing(self, tmp_path):
        """Test update syncs templates when config.json missing (3-stage workflow, treats as new project)"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai directory without config.json
            moai_dir = Path(".moai")
            moai_dir.mkdir()

            # Mock version functions to return same version
            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
                patch("moai_adk.cli.commands.update._get_project_config_version") as mock_proj_ver,
                patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            ):

                mock_current.return_value = "0.3.2"
                mock_latest.return_value = "0.3.2"
                mock_pkg_ver.return_value = "0.3.2"
                mock_proj_ver.return_value = "0.0.0"  # No config = version 0.0.0
                mock_sync.return_value = True

                result = runner.invoke(update)
                assert result.exit_code == 0
                # In 3-stage workflow, 0.0.0 < package version, so sync templates
                assert "Syncing templates" in result.output

    def test_update_check_when_local_version_newer(self, tmp_path):
        """Test update --check when local version is newer than PyPI (development version)"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai directory
            Path(".moai").mkdir()

            # Mock version functions to return dev version > latest
            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            ):

                mock_current.return_value = "0.4.0"
                mock_latest.return_value = "0.3.3"  # Older version on PyPI

                result = runner.invoke(update, ["--check"])
                assert result.exit_code == 0
                assert "Checking versions" in result.output
                assert "Development version" in result.output or "Dev version" in result.output

    def test_update_skips_when_local_version_newer(self, tmp_path):
        """Test update proceeds to config comparison when local version is newer than PyPI"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai directory
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            config_data = {
                "moai": {"version": "0.4.0"},
                "project": {"template_version": "0.4.0", "optimized": True, "mode": "personal"},
            }
            import json

            (moai_dir / "config.json").write_text(json.dumps(config_data))

            # Mock version functions to return dev version > latest
            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
                patch("moai_adk.cli.commands.update._get_project_config_version") as mock_proj_ver,
                patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            ):

                mock_current.return_value = "0.4.0"
                mock_latest.return_value = "0.3.3"  # Older version on PyPI
                mock_pkg_ver.return_value = "0.4.0"
                mock_proj_ver.return_value = "0.4.0"  # Same template version
                mock_sync.return_value = True

                result = runner.invoke(update)
                assert result.exit_code == 0
                # Should proceed to config comparison, then skip sync since versions match
                assert "Comparing config versions" in result.output
                assert "Templates are up to date" in result.output

    def test_update_handles_pypi_fetch_failure(self, tmp_path):
        """Test update handles PyPI fetch failure gracefully"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai directory
            Path(".moai").mkdir()

            # Mock version functions - _get_latest_version raises RuntimeError
            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            ):

                mock_current.return_value = "0.6.1"
                mock_latest.side_effect = RuntimeError("Failed to fetch latest version from PyPI")

                result = runner.invoke(update, ["--check"])
                # Should handle error gracefully
                assert "Error" in result.output
                assert "Failed to fetch latest version" in result.output

    def test_update_proceeds_with_force_when_pypi_fails(self, tmp_path):
        """Test update proceeds with --force even when PyPI fetch fails"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai directory
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            import json

            config_data = {"moai": {"version": "0.6.1"}, "project": {"template_version": "0.6.0", "optimized": False}}
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            # Mock version functions - _get_latest_version raises RuntimeError but --force proceeds
            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
                patch("moai_adk.cli.commands.update._get_project_config_version") as mock_proj_ver,
                patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            ):

                mock_current.return_value = "0.6.1"
                mock_latest.side_effect = RuntimeError("Failed to fetch latest version from PyPI")
                mock_pkg_ver.return_value = "0.6.1"
                mock_proj_ver.return_value = "0.6.0"
                mock_sync.return_value = True

                result = runner.invoke(update, ["--force"])
                assert result.exit_code == 0
                # With --force, should proceed to Stage 2 and compare config versions
                assert "Syncing templates" in result.output


class TestUpdateVersionFunctions:
    """Test version detection functions (Phase 1 of v0.6.3)"""

    def test_get_package_config_version_returns_current_version(self):
        """Test _get_package_config_version returns current installed package version"""
        from moai_adk.cli.commands.update import _get_package_config_version

        # Should return __version__ (current installed package version)
        with patch("moai_adk.cli.commands.update.__version__", "0.6.1"):
            result = _get_package_config_version()
            # Package template version = current installed package version
            assert result == "0.6.1"

    def test_get_project_config_version_missing_config(self, tmp_path):
        """Test _get_project_config_version returns 0.0.0 when config missing"""
        from moai_adk.cli.commands.update import _get_project_config_version

        project_path = tmp_path / "test-project"
        project_path.mkdir()

        # No .moai/config.json exists
        result = _get_project_config_version(project_path)
        assert result == "0.0.0"

    def test_get_project_config_version_from_template_version(self, tmp_path):
        """Test _get_project_config_version reads template_version field"""
        import json

        from moai_adk.cli.commands.update import _get_project_config_version

        project_path = tmp_path / "test-project"
        moai_dir = project_path / ".moai" / "config"
        moai_dir.mkdir(parents=True)

        # Create config with template_version
        config_data = {"moai": {"version": "0.6.0"}, "project": {"template_version": "0.6.1", "optimized": False}}
        (moai_dir / "config.json").write_text(json.dumps(config_data))

        result = _get_project_config_version(project_path)
        assert result == "0.6.1"

    def test_get_project_config_version_fallback_to_moai_version(self, tmp_path):
        """Test _get_project_config_version falls back to moai.version"""
        import json

        from moai_adk.cli.commands.update import _get_project_config_version

        project_path = tmp_path / "test-project"
        moai_dir = project_path / ".moai" / "config"
        moai_dir.mkdir(parents=True)

        # Create config without template_version but with moai.version
        config_data = {"moai": {"version": "0.6.0"}, "project": {"optimized": False}}
        (moai_dir / "config.json").write_text(json.dumps(config_data))

        result = _get_project_config_version(project_path)
        assert result == "0.6.0"

    def test_get_project_config_version_invalid_json(self, tmp_path):
        """Test _get_project_config_version raises ValueError on invalid JSON"""
        from moai_adk.cli.commands.update import _get_project_config_version

        project_path = tmp_path / "test-project"
        moai_dir = project_path / ".moai" / "config"
        moai_dir.mkdir(parents=True)

        # Create invalid JSON
        (moai_dir / "config.json").write_text("{invalid json")

        try:
            _get_project_config_version(project_path)
            assert False, "Should raise ValueError"
        except ValueError as e:
            assert "Failed to parse project config.json" in str(e)


class TestUpdateThreeStageWorkflow:
    """Test 3-stage workflow (Phase 2 of v0.6.3)"""

    def test_update_skips_sync_when_template_version_up_to_date(self, tmp_path):
        """Test update skips template sync when versions are equal"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai structure with template_version
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            import json

            config_data = {
                "moai": {"version": "0.6.1"},
                "project": {"template_version": "0.6.1", "optimized": False, "name": "test"},  # Same as package
            }
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            # Mock version functions
            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
                patch("moai_adk.cli.commands.update._get_project_config_version") as mock_proj_ver,
            ):

                mock_current.return_value = "0.6.1"
                mock_latest.return_value = "0.6.1"
                mock_pkg_ver.return_value = "0.6.1"
                mock_proj_ver.return_value = "0.6.1"

                result = runner.invoke(update)
                assert result.exit_code == 0
                # Should show config version comparison
                assert "Comparing config versions" in result.output
                # Should NOT sync templates
                assert "Project already has latest template version" in result.output
                assert "Templates are up to date" in result.output

    def test_update_syncs_when_template_version_outdated(self, tmp_path):
        """Test update syncs templates when package version > project version"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai structure with older template_version
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            import json

            config_data = {
                "moai": {"version": "0.6.1"},
                "project": {"template_version": "0.6.0", "optimized": False, "name": "test"},  # Older than package
            }
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            # Mock version functions and template sync
            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
                patch("moai_adk.cli.commands.update._get_project_config_version") as mock_proj_ver,
                patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            ):

                mock_current.return_value = "0.6.1"
                mock_latest.return_value = "0.6.1"
                mock_pkg_ver.return_value = "0.6.1"
                mock_proj_ver.return_value = "0.6.0"
                mock_sync.return_value = True

                result = runner.invoke(update)
                assert result.exit_code == 0
                # Should show config version comparison
                assert "Comparing config versions" in result.output
                # Should sync templates
                assert "Syncing templates" in result.output
                assert "0.6.0 → 0.6.1" in result.output
                mock_sync.assert_called_once()

    def test_update_handles_version_detection_error(self, tmp_path):
        """Test update proceeds with safe defaults when version detection fails"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai structure
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            import json

            config_data = {"moai": {"version": "0.6.1"}, "project": {"optimized": False, "name": "test"}}
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            # Mock version functions - config version detection fails
            with (
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
                patch("moai_adk.cli.commands.update._get_project_config_version") as mock_proj_ver,
                patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
                patch("moai_adk.cli.commands.update.__version__", "0.6.1"),
            ):

                mock_current.return_value = "0.6.1"
                mock_latest.return_value = "0.6.1"
                # Simulate version detection error
                mock_pkg_ver.side_effect = ValueError("Config parse error")
                mock_proj_ver.side_effect = ValueError("Config parse error")
                mock_sync.return_value = True

                result = runner.invoke(update)
                assert result.exit_code == 0
                # Should show warning but still proceed
                assert "Warning" in result.output
                # Should sync templates (safe choice on error)
                assert "Syncing templates" in result.output

    def test_update_preserves_template_version_in_config(self, tmp_path):
        """Test that _preserve_project_metadata updates template_version"""
        from moai_adk.cli.commands.update import _preserve_project_metadata

        project_path = tmp_path / "test-project"
        moai_dir = project_path / ".moai" / "config"
        moai_dir.mkdir(parents=True)

        # Create initial config
        import json

        config_data = {"moai": {"version": "0.6.0"}, "project": {"name": "test", "optimized": False}}
        (moai_dir / "config.json").write_text(json.dumps(config_data))

        # Call _preserve_project_metadata to update version
        context = {
            "PROJECT_NAME": "test",
            "PROJECT_MODE": "personal",
            "PROJECT_DESCRIPTION": "Test project",
            "CREATION_TIMESTAMP": "2025-10-28T00:00:00",
        }
        _preserve_project_metadata(project_path, context, {}, "0.6.1")

        # Verify template_version was set
        updated_config = json.loads((moai_dir / "config.json").read_text())
        assert updated_config["project"]["template_version"] == "0.6.1"
        assert updated_config["moai"]["version"] == "0.6.1"

    def test_get_project_config_version_with_placeholder(self, tmp_path):
        """Test that _get_project_config_version detects and handles placeholder values"""
        import json

        from moai_adk.cli.commands.update import _get_project_config_version

        project_path = tmp_path / "test-project"
        moai_dir = project_path / ".moai" / "config"
        moai_dir.mkdir(parents=True)

        # Test 1: Placeholder in moai.version (no template_version)
        config_data = {"moai": {"version": "{{MOAI_VERSION}}"}}
        (moai_dir / "config.json").write_text(json.dumps(config_data))

        result = _get_project_config_version(project_path)
        assert result == "0.0.0", "Should return 0.0.0 when moai.version is a placeholder"

        # Test 2: Placeholder in project.template_version - falls back to moai.version
        config_data = {"moai": {"version": "0.8.0"}, "project": {"template_version": "{{TEMPLATE_VERSION}}"}}
        (moai_dir / "config.json").write_text(json.dumps(config_data))

        result = _get_project_config_version(project_path)
        assert result == "0.8.0", "Should fall back to moai.version when template_version is a placeholder"

        # Test 3: Both have placeholders
        config_data = {"moai": {"version": "{{MOAI_VERSION}}"}, "project": {"template_version": "{{TEMPLATE_VERSION}}"}}
        (moai_dir / "config.json").write_text(json.dumps(config_data))

        result = _get_project_config_version(project_path)
        assert result == "0.0.0", "Should return 0.0.0 when both versions are placeholders"

        # Test 4: Valid versions (no placeholders)
        config_data = {"moai": {"version": "0.8.1"}, "project": {"template_version": "0.8.1"}}
        (moai_dir / "config.json").write_text(json.dumps(config_data))

        result = _get_project_config_version(project_path)
        assert result == "0.8.1", "Should return template_version when it's valid"

        # Test 5: template_version preferred over moai.version
        config_data = {"moai": {"version": "0.8.0"}, "project": {"template_version": "0.8.1"}}
        (moai_dir / "config.json").write_text(json.dumps(config_data))

        result = _get_project_config_version(project_path)
        assert result == "0.8.1", "Should prefer template_version over moai.version"

    def test_compare_versions_with_invalid_version(self):
        """Test that _compare_versions handles InvalidVersion exceptions gracefully"""
        import pytest
        from packaging.version import InvalidVersion

        from moai_adk.cli.commands.update import _compare_versions

        # Valid versions should work fine
        result = _compare_versions("0.8.0", "0.8.1")
        assert result == -1, "0.8.0 should be less than 0.8.1"

        # Invalid versions should raise InvalidVersion
        with pytest.raises(InvalidVersion):
            _compare_versions("{{MOAI_VERSION}}", "0.8.1")

        with pytest.raises(InvalidVersion):
            _compare_versions("0.8.1", "{{MOAI_VERSION}}")

    def test_update_command_with_unsubstituted_config(self, tmp_path):
        """Test that update command handles unsubstituted placeholders gracefully"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai structure with placeholder version
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            import json

            config_data = {
                "moai": {"version": "{{MOAI_VERSION}}"},  # Unsubstituted placeholder
                "project": {"template_version": "0.8.0", "mode": "personal"},
            }
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            # Mock the TemplateProcessor to avoid actual file operations
            with (
                patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor,
                patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
                patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
                patch("moai_adk.cli.commands.update._get_package_config_version") as mock_pkg_ver,
                patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_installer,
            ):

                mock_instance = Mock()
                mock_instance.copy_templates.return_value = None
                mock_processor.return_value = mock_instance

                mock_current.return_value = "0.8.1"
                mock_latest.return_value = "0.8.1"
                mock_pkg_ver.return_value = "0.8.1"
                mock_installer.return_value = ["echo", "upgraded"]

                # Run update - should handle the invalid version gracefully
                result = runner.invoke(update, ["--force"])

                # Should not fail, should show warning about invalid version
                assert result.exit_code == 0, f"Update should succeed but got: {result.output}"
                assert (
                    "Comparing config versions" in result.output
                    or "Invalid version" in result.output
                    or "Forcing template sync" in result.output
                ), f"Should handle invalid version gracefully: {result.output}"
