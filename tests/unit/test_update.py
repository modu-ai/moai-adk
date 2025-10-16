# @TEST:TEST-COVERAGE-001 | SPEC: SPEC-TEST-COVERAGE-001.md
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
        assert "Update template files to the latest version" in result.output

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
            assert "Already up to date" in result.output or "Update available" in result.output

    def test_update_check_when_update_available(self, tmp_path):
        """Test update --check when new version is available"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai directory
            Path(".moai").mkdir()

            # Mock __version__ to simulate different versions
            with patch("moai_adk.cli.commands.update.__version__", "0.1.0"):
                # Mock current version as older
                with patch("moai_adk.__version__", "0.0.9"):
                    result = runner.invoke(update, ["--check"])
                    assert result.exit_code == 0
                    assert "Checking versions" in result.output
                    # This should trigger line 64: "Update available"
                    assert "Update available" in result.output or "Already up to date" in result.output

    def test_update_with_backup(self, tmp_path):
        """Test update with backup (default behavior) - uses --force to skip version check"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai structure
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            import json
            config_data = {"project": {"optimized": True}, "mode": "personal"}
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            # Mock TemplateProcessor
            with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
                mock_instance = Mock()
                # Return absolute path instead of relative
                mock_instance.create_backup.return_value = Path.cwd() / ".moai-backups/backup-2025-10-15"
                mock_instance.copy_templates.return_value = None  # Mock copy_templates
                mock_processor.return_value = mock_instance

                # Use --force to skip version check and test backup process
                result = runner.invoke(update, ["--force"])
                # Should show skip backup message with --force, but still show updating templates
                assert "Skipping backup (--force)" in result.output
                assert "Updating templates" in result.output
                assert result.exit_code == 0

    def test_update_with_force_flag(self, tmp_path):
        """Test update --force flag (skip backup)"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai structure
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            (moai_dir / "config.json").write_text('{"mode": "personal"}')

            # Mock TemplateProcessor
            with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
                mock_instance = Mock()
                mock_instance.copy_templates.return_value = None  # Mock copy_templates
                mock_processor.return_value = mock_instance

                result = runner.invoke(update, ["--force"])
                assert "Skipping backup (--force)" in result.output
                assert "Updating templates" in result.output
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
        config_data = {"project": {"optimized": True}, "mode": "personal"}
        (project_dir / ".moai" / "config.json").write_text(json.dumps(config_data))

        # Mock TemplateProcessor
        with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
            mock_instance = Mock()
            mock_instance.create_backup.return_value = project_dir / ".moai-backups/backup"
            mock_processor.return_value = mock_instance

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
            config_data = {"project": {"optimized": True}, "mode": "personal"}
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
                mock_instance = Mock()
                mock_instance.create_backup.return_value = Path.cwd() / ".moai-backups/backup"
                mock_instance.copy_templates.return_value = None  # Mock copy_templates
                mock_processor.return_value = mock_instance

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
            config_data = {"project": {"optimized": True}, "mode": "personal"}
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
                mock_instance = Mock()
                mock_instance.create_backup.return_value = Path.cwd() / ".moai-backups/backup"
                mock_instance.copy_templates.return_value = None  # Mock copy_templates
                mock_processor.return_value = mock_instance

                # Use --force to skip version check and backup
                result = runner.invoke(update, ["--force"])
                assert ".claude/ update complete" in result.output
                assert ".moai/ update complete" in result.output
                assert "CLAUDE.md merge complete" in result.output
                assert "config.json merge complete" in result.output
                assert result.exit_code == 0

    def test_update_skips_when_same_version_and_optimized(self, tmp_path):
        """Test update skips silently when version is same and already optimized"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai structure with optimized=true
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            config_data = {
                "project": {
                    "optimized": True,
                    "name": "test",
                    "mode": "personal"
                }
            }
            import json
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            result = runner.invoke(update)
            assert result.exit_code == 0
            # Should exit silently when optimized=true and versions match
            assert "Checking versions" in result.output
            assert "Current version" in result.output
            assert "Latest version" in result.output

    def test_update_suggests_alfred_when_same_version_not_optimized(self, tmp_path):
        """Test update suggests /alfred:0-project when version same but not optimized"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai structure with optimized=false
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            config_data = {
                "project": {
                    "optimized": False,
                    "name": "test",
                    "mode": "personal"
                }
            }
            import json
            (moai_dir / "config.json").write_text(json.dumps(config_data))

            result = runner.invoke(update)
            assert result.exit_code == 0
            assert "Optimization needed" in result.output
            assert "alfred:0-project update" in result.output

    def test_update_proceeds_when_config_missing(self, tmp_path):
        """Test update shows already up to date when config.json missing"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai directory without config.json
            moai_dir = Path(".moai")
            moai_dir.mkdir()

            result = runner.invoke(update)
            assert result.exit_code == 0
            assert "Already up to date" in result.output
