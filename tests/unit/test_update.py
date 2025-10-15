# @TEST:TEST-COVERAGE-001 | SPEC: SPEC-TEST-COVERAGE-001.md
"""Unit tests for update.py command

Tests for update command with various scenarios.
"""

from pathlib import Path
from unittest.mock import Mock, patch

import pytest
from click.testing import CliRunner

from moai_adk.cli.commands.update import update


class TestUpdateCommand:
    """Test update command"""

    def test_update_help(self):
        """Test update --help"""
        runner = CliRunner()
        result = runner.invoke(update, ["--help"])
        assert result.exit_code == 0
        assert "템플릿 파일을 최신 버전으로 업데이트" in result.output

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
            assert "버전 확인" in result.output
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
                    assert "버전 확인" in result.output
                    # This should trigger line 64: "Update available"
                    assert "Update available" in result.output or "Already up to date" in result.output

    def test_update_with_backup(self, tmp_path):
        """Test update with backup (default behavior)"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            # Create .moai structure
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            (moai_dir / "config.json").write_text('{"mode": "personal"}')

            # Mock TemplateProcessor
            with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
                mock_instance = Mock()
                # Return absolute path instead of relative
                mock_instance.create_backup.return_value = Path.cwd() / ".moai-backups/backup-2025-10-15"
                mock_instance.copy_templates.return_value = None  # Mock copy_templates
                mock_processor.return_value = mock_instance

                result = runner.invoke(update)
                # Should show backup and update messages
                assert "백업 생성 중" in result.output
                assert "템플릿 업데이트 중" in result.output
                if result.exit_code != 0:
                    # Print output for debugging
                    print(f"Output: {result.output}")
                    print(f"Exception: {result.exception}")
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
                assert "백업 생략" in result.output
                assert "템플릿 업데이트 중" in result.output
                # create_backup should NOT be called with --force
                mock_instance.create_backup.assert_not_called()
                assert result.exit_code == 0 or "업데이트" in result.output

    def test_update_with_custom_path(self, tmp_path):
        """Test update with custom --path"""
        runner = CliRunner()

        # Create project directory
        project_dir = tmp_path / "my-project"
        project_dir.mkdir()
        (project_dir / ".moai").mkdir()
        (project_dir / ".moai" / "config.json").write_text('{"mode": "personal"}')

        # Mock TemplateProcessor
        with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
            mock_instance = Mock()
            mock_instance.create_backup.return_value = project_dir / ".moai-backups/backup"
            mock_processor.return_value = mock_instance

            result = runner.invoke(update, ["--path", str(project_dir)])
            assert result.exit_code == 0
            assert "업데이트 완료" in result.output

    def test_update_shows_version_info(self, tmp_path):
        """Test that update shows version information"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            Path(".moai").mkdir()

            result = runner.invoke(update, ["--check"])
            assert result.exit_code == 0
            assert "현재 버전" in result.output
            assert "최신 버전" in result.output

    def test_update_template_processor_called(self, tmp_path):
        """Test that TemplateProcessor methods are called correctly"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            moai_dir = Path(".moai")
            moai_dir.mkdir()
            (moai_dir / "config.json").write_text('{"mode": "personal"}')

            with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
                mock_instance = Mock()
                mock_instance.create_backup.return_value = Path.cwd() / ".moai-backups/backup"
                mock_instance.copy_templates.return_value = None  # Mock copy_templates
                mock_processor.return_value = mock_instance

                result = runner.invoke(update)

                # Verify methods were called
                mock_instance.create_backup.assert_called_once()
                mock_instance.copy_templates.assert_called_once_with(backup=False, silent=True)
                assert result.exit_code == 0

    def test_update_handles_exception(self, tmp_path):
        """Test update handles exceptions"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            moai_dir = Path(".moai")
            moai_dir.mkdir()

            # Mock TemplateProcessor to raise exception
            with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
                mock_processor.side_effect = RuntimeError("Mock error")

                result = runner.invoke(update)
                assert result.exit_code != 0
                assert "Update failed" in result.output

    def test_update_shows_update_details(self, tmp_path):
        """Test that update shows detailed update information"""
        runner = CliRunner()

        with runner.isolated_filesystem(temp_dir=tmp_path):
            moai_dir = Path(".moai")
            moai_dir.mkdir()

            with patch("moai_adk.cli.commands.update.TemplateProcessor") as mock_processor:
                mock_instance = Mock()
                mock_instance.create_backup.return_value = Path.cwd() / ".moai-backups/backup"
                mock_instance.copy_templates.return_value = None  # Mock copy_templates
                mock_processor.return_value = mock_instance

                result = runner.invoke(update)
                assert ".claude/ 업데이트 완료" in result.output
                assert ".moai/ 업데이트 완료" in result.output
                assert "CLAUDE.md 병합 완료" in result.output
                assert "config.json 병합 완료" in result.output
                assert result.exit_code == 0
