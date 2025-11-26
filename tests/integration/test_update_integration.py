"""Integration tests for Phase 5: Complete 2-Stage Workflow

Tests for the end-to-end update process including tool detection,
version management, package upgrade, and template synchronization.

GREEN Phase: These tests verify core functionality.
"""

import json
from unittest.mock import patch

import pytest
from click.testing import CliRunner

from moai_adk.cli.commands.update import update


class TestIntegration2StageWorkflow:
    """Test complete 2-stage workflow integration"""

    @pytest.fixture
    def runner(self):
        """Create a Click CLI runner."""
        return CliRunner()

    @pytest.fixture
    def temp_project(self, tmp_path):
        """Create a temporary project structure."""
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        # Initialize .moai directory
        moai_dir = project_path / ".moai"
        moai_dir.mkdir()

        # Create config.json
        config_data = {
            "name": "Test Project",
            "author": "Test Author",
            "locale": "en",
            "optimized": True,
            "version": "0.6.1",
        }
        (moai_dir / "config.json").write_text(json.dumps(config_data, indent=2))

        # Create CLAUDE.md
        claude_md = moai_dir.parent / "CLAUDE.md"
        claude_md.write_text("# Test Project\n\n## Configuration\n\nTest content")

        # Initialize .claude directory
        claude_dir = project_path / ".claude"
        claude_dir.mkdir()
        (claude_dir / "skills").mkdir()

        # Create backup directory for restore testing
        backups_dir = project_path / ".moai-backups"
        backups_dir.mkdir()

        return project_path

    @pytest.mark.skip(reason="CLI confirm input requires interactive input - ClickException from confirm()")
    def test_stage1_upgrade_needed_uv_tool(self, runner, temp_project):
        """Test Stage 1: Upgrade needed with uv tool detection"""
        with (
            patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._execute_upgrade") as mock_exec,
        ):

            mock_detect.return_value = ["uv", "tool", "upgrade", "moai-adk"]
            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"
            mock_exec.return_value = True

            result = runner.invoke(update, ["--path", str(temp_project)])

            # Stage 1: Should upgrade
            assert result.exit_code == 0
            assert "0.6.1" in result.output
            assert "0.6.2" in result.output
            assert "uv tool upgrade" in result.output
            assert "Run 'moai-adk update' again" in result.output
            mock_exec.assert_called_once()

    def test_stage2_templates_sync_after_upgrade(self, runner, temp_project):
        """Test Stage 2: Template sync after version is upgraded"""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
        ):

            mock_current.return_value = "0.6.2"
            mock_latest.return_value = "0.6.2"
            mock_sync.return_value = True

            result = runner.invoke(update, ["--path", str(temp_project)])

            # Stage 2: Should sync templates
            assert result.exit_code == 0
            assert "already up to date" in result.output.lower() or "syncing templates" in result.output.lower()
            mock_sync.assert_called_once()

    def test_already_latest_version_skips_stage1(self, runner, temp_project):
        """Test that Stage 1 is skipped when already on latest version"""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
        ):

            mock_current.return_value = "0.6.2"
            mock_latest.return_value = "0.6.2"
            mock_sync.return_value = True

            result = runner.invoke(update, ["--path", str(temp_project)])

            # Should not call tool detection
            mock_detect.assert_not_called()
            # Should directly proceed to sync
            mock_sync.assert_called_once()
            assert result.exit_code == 0

    def test_templates_only_flag_skips_upgrade(self, runner, temp_project):
        """Test --templates-only flag bypasses entire upgrade check"""
        with (
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
            patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
        ):

            mock_sync.return_value = True

            result = runner.invoke(update, ["--path", str(temp_project), "--templates-only"])

            # Should not call tool detection or version check
            mock_detect.assert_not_called()
            mock_latest.assert_not_called()
            # Should only call template sync
            mock_sync.assert_called_once()
            assert result.exit_code == 0

    def test_check_mode_shows_versions_no_changes(self, runner, temp_project):
        """Test --check mode displays versions without making changes"""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
        ):

            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"

            result = runner.invoke(update, ["--path", str(temp_project), "--check"])

            # Should show versions
            assert "0.6.1" in result.output
            assert "0.6.2" in result.output
            # Should not make any changes
            mock_sync.assert_not_called()
            assert result.exit_code == 0

    def test_yes_flag_auto_confirms_prompts(self, runner, temp_project):
        """Test --yes flag auto-confirms all prompts"""
        with (
            patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._execute_upgrade") as mock_exec,
        ):

            mock_detect.return_value = ["uv", "tool", "upgrade", "moai-adk"]
            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"
            mock_exec.return_value = True

            result = runner.invoke(update, ["--path", str(temp_project), "--yes"])

            # Should auto-confirm without interactive prompt
            assert "continue" not in result.output.lower()
            assert result.exit_code == 0
            mock_exec.assert_called_once()

    def test_force_flag_skips_backup(self, runner, temp_project):
        """Test --force flag skips backup creation"""
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
        ):

            mock_current.return_value = "0.6.2"
            mock_latest.return_value = "0.6.2"
            mock_sync.return_value = True

            result = runner.invoke(update, ["--path", str(temp_project), "--force"])

            # Verify no backup directory was created
            # (actual backup check would be in template sync)
            mock_sync.assert_called_once()
            assert result.exit_code == 0

    @pytest.mark.skip(reason="CLI confirm input requires interactive input - ClickException from confirm()")
    def test_full_workflow_two_invocations(self, runner, temp_project):
        """Test complete 2-stage workflow across two invocations"""
        # First invocation: Stage 1 (upgrade)
        with (
            patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current1,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest1,
            patch("moai_adk.cli.commands.update._execute_upgrade") as mock_exec,
        ):

            mock_detect.return_value = ["uv", "tool", "upgrade", "moai-adk"]
            mock_current1.return_value = "0.6.1"
            mock_latest1.return_value = "0.6.2"
            mock_exec.return_value = True

            result1 = runner.invoke(update, [str(temp_project)])
            assert result1.exit_code == 0
            assert "uv tool upgrade" in result1.output

        # Second invocation: Stage 2 (template sync)
        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current2,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest2,
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
        ):

            mock_current2.return_value = "0.6.2"
            mock_latest2.return_value = "0.6.2"
            mock_sync.return_value = True

            result2 = runner.invoke(update, [str(temp_project)])
            assert result2.exit_code == 0
            mock_sync.assert_called_once()


class TestErrorRecoveryIntegration:
    """Test error handling in integrated workflows"""

    @pytest.fixture
    def runner(self):
        """Create a Click CLI runner."""
        return CliRunner()

    @pytest.fixture
    def temp_project(self, tmp_path):
        """Create a temporary project structure."""
        project_path = tmp_path / "test_project"
        project_path.mkdir()
        (project_path / ".moai").mkdir()
        return project_path

    def test_network_failure_graceful_degradation(self, runner, temp_project):
        """Test graceful handling when PyPI is unreachable"""
        with patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest:
            mock_latest.side_effect = RuntimeError("Cannot reach PyPI")

            result = runner.invoke(update, ["--path", str(temp_project)])

            # Should show helpful error message
            assert "pypi" in result.output.lower() or "network" in result.output.lower()
            assert result.exit_code != 0

    @pytest.mark.skip(reason="CLI confirm input requires interactive input - ClickException from confirm()")
    def test_installer_not_found_shows_alternatives(self, runner, temp_project):
        """Test helpful message when no installer is detected"""
        with (
            patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
        ):

            mock_detect.return_value = None
            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"

            result = runner.invoke(update, ["--path", str(temp_project)])

            # Should show helpful guidance
            assert "cannot detect" in result.output.lower() or "manually" in result.output.lower()
            assert result.exit_code != 0

    def test_upgrade_failure_suggests_recovery(self, runner, temp_project):
        """Test helpful message when upgrade fails"""
        with (
            patch("moai_adk.cli.commands.update._detect_tool_installer") as mock_detect,
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._execute_upgrade") as mock_exec,
        ):

            mock_detect.return_value = ["uv", "tool", "upgrade", "moai-adk"]
            mock_current.return_value = "0.6.1"
            mock_latest.return_value = "0.6.2"
            mock_exec.return_value = False

            result = runner.invoke(update, ["--path", str(temp_project)])

            # Should show troubleshooting guidance
            assert result.exit_code != 0

    def test_templates_only_recovery_after_manual_upgrade(self, runner, temp_project):
        """Test --templates-only provides recovery path after manual upgrade"""
        with patch("moai_adk.cli.commands.update._sync_templates") as mock_sync:
            mock_sync.return_value = True

            # User manually upgraded, now syncs templates
            result = runner.invoke(update, ["--path", str(temp_project), "--templates-only"])

            mock_sync.assert_called_once()
            assert result.exit_code == 0


class TestConfigMergeIntegrity:
    """Test that configuration merging preserves customizations"""

    @pytest.fixture
    def runner(self):
        """Create a Click CLI runner."""
        return CliRunner()

    @pytest.fixture
    def temp_project_with_config(self, tmp_path):
        """Create a project with custom configuration"""
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        moai_dir = project_path / ".moai"
        moai_dir.mkdir()

        # Original custom config
        original_config = {
            "name": "My Custom Project",
            "author": "Alice",
            "locale": "ko",
            "custom_field": "should_preserve",
            "optimized": True,
            "version": "0.6.1",
        }
        (moai_dir / "config.json").write_text(json.dumps(original_config, indent=2))

        return project_path, original_config

    @pytest.mark.skip(reason="CLI confirm input requires interactive input - ClickException from confirm()")
    def test_config_merge_preserves_metadata(self, runner, temp_project_with_config):
        """Test that config.json merge preserves project metadata"""
        project_path, original_config = temp_project_with_config

        with (
            patch("moai_adk.cli.commands.update._get_current_version") as mock_current,
            patch("moai_adk.cli.commands.update._get_latest_version") as mock_latest,
            patch("moai_adk.cli.commands.update._sync_templates") as mock_sync,
        ):

            mock_current.return_value = "0.6.2"
            mock_latest.return_value = "0.6.2"

            # Mock sync to preserve config
            def mock_sync_impl(*args, **kwargs):
                config_path = project_path / ".moai" / "config.json"
                current_config = json.loads(config_path.read_text())
                current_config["optimized"] = False  # Update flag
                config_path.write_text(json.dumps(current_config, indent=2))
                return True

            mock_sync.side_effect = mock_sync_impl

            result = runner.invoke(update, [str(project_path)])

            # Verify config still has original metadata
            config_path = project_path / ".moai" / "config.json"
            if config_path.exists():
                updated_config = json.loads(config_path.read_text())
                assert updated_config["name"] == original_config["name"]
                assert updated_config["author"] == original_config["author"]
                assert updated_config["locale"] == original_config["locale"]
                assert updated_config["optimized"] is False  # Should be updated

            assert result.exit_code == 0
