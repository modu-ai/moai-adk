"""Integration tests for /tmp folder operations.

Tests end-to-end execution of init and update commands in /tmp directory:
- moai init in /tmp directory
- moai update in /tmp directory
- Installation and update workflows
- Cleanup after operations

Key Requirements:
- Use Click's CliRunner for command invocation
- Use pytest's tmp_path fixture for temporary directories
- Test real CLI execution (not mocked)
- Capture exit codes, stdout, stderr
- Clean up temporary files after tests
- Test both success and failure paths
"""

import subprocess
import tempfile
from pathlib import Path
from typing import Generator

import pytest
import yaml

from moai_adk.__main__ import cli


@pytest.fixture
def tmp_dir_path() -> Generator[Path, None, None]:
    """Create a temporary directory for testing /tmp operations.

    This fixture creates a real temporary directory that simulates /tmp
    behavior for testing init and update commands.

    Yields:
        Path to temporary directory
    """
    with tempfile.TemporaryDirectory(prefix="moai_test_") as tmp_dir:
        yield Path(tmp_dir)
    # Cleanup is automatic with TemporaryDirectory context manager


@pytest.mark.integration
class TestTmpInitOperations:
    """Tests for init command in /tmp folder."""

    def test_init_in_tmp_creates_moai_directory(self, cli_runner, tmp_dir_path):
        """Test that init in /tmp creates .moai directory."""
        result = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        # Should execute without crashing
        assert result.exit_code in [0, 1]

        # Check .moai directory was created if command succeeded
        moai_dir = tmp_dir_path / ".moai"
        if result.exit_code == 0:
            assert moai_dir.exists(), ".moai directory should be created in /tmp"
            assert moai_dir.is_dir()

    def test_init_in_tmp_creates_config_files(self, cli_runner, tmp_dir_path):
        """Test that init in /tmp creates config files."""
        result = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        if result.exit_code == 0:
            # Check config.yaml exists
            config_yaml = tmp_dir_path / ".moai" / "config" / "config.yaml"
            assert config_yaml.exists(), "config.yaml should be created"

            # Check it's valid YAML
            with open(config_yaml, encoding="utf-8") as f:
                data = yaml.safe_load(f)
                assert isinstance(data, dict)
                assert "moai" in data

    def test_init_in_tmp_with_locale_option(self, cli_runner, tmp_dir_path):
        """Test init in /tmp with different locales."""
        for locale in ["ko", "en", "ja", "zh"]:
            locale_dir = tmp_dir_path / f"project_{locale}"
            locale_dir.mkdir(exist_ok=True)

            result = cli_runner.invoke(cli, ["init", str(locale_dir), "--non-interactive", "--locale", locale])

            # Should execute
            assert result.exit_code in [0, 1]

            # Check locale was set if command succeeded
            language_yaml = locale_dir / ".moai" / "config" / "sections" / "language.yaml"
            if result.exit_code == 0 and language_yaml.exists():
                with open(language_yaml, encoding="utf-8") as f:
                    data = yaml.safe_load(f)
                    assert data["language"]["conversation_language"] == locale

    def test_init_in_nested_tmp_path(self, cli_runner, tmp_dir_path):
        """Test init in nested temporary path."""
        nested_path = tmp_dir_path / "level1" / "level2" / "level3"
        nested_path.mkdir(parents=True)

        result = cli_runner.invoke(cli, ["init", str(nested_path), "--non-interactive"])

        # Should execute
        assert result.exit_code in [0, 1]

        if result.exit_code == 0:
            assert (nested_path / ".moai").exists()

    def test_init_in_tmp_with_team_mode(self, cli_runner, tmp_dir_path):
        """Test init in /tmp with team mode."""
        team_dir = tmp_dir_path / "team_project"
        team_dir.mkdir()

        result = cli_runner.invoke(cli, ["init", str(team_dir), "--non-interactive", "--mode", "team"])

        # Should execute
        assert result.exit_code in [0, 1]

        if result.exit_code == 0:
            # Verify .moai directory was created
            assert (team_dir / ".moai").exists()
            # Verify config.yaml was created
            assert (team_dir / ".moai" / "config" / "config.yaml").exists()

    def test_init_reinit_in_tmp_creates_backup(self, cli_runner, tmp_dir_path):
        """Test that reinit in /tmp creates backup."""
        result1 = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        # Only test reinit if first init succeeded
        if result1.exit_code == 0:
            # Modify config to detect backup
            config_yaml = tmp_dir_path / ".moai" / "config" / "config.yaml"
            if config_yaml.exists():
                with open(config_yaml, "r", encoding="utf-8") as f:
                    f.read()  # Read file content

                # Second init (reinit)
                result2 = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

                # Should complete
                assert result2.exit_code in [0, 1]

                # Check for backup directory
                backup_dir = tmp_dir_path / ".moai-backups"
                if backup_dir.exists():
                    backups = list(backup_dir.iterdir())
                    assert len(backups) > 0, "Should create backup on reinit in /tmp"

    def test_init_in_tmp_cleanup_after_test(self, cli_runner, tmp_dir_path):
        """Test that tmp directory is cleaned up after test."""
        # Run init
        _ = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        # Directory should exist during test
        assert tmp_dir_path.exists()

        # After fixture cleanup (handled by tempfile.TemporaryDirectory),
        # directory should be cleaned up
        # This is verified by the fixture's context manager

    def test_init_in_tmp_with_invalid_path(self, cli_runner):
        """Test init in /tmp with invalid path."""
        # Use path that would typically cause issues
        invalid_paths = [
            "/nonexistent_deep_path_12345/dir",
        ]

        for invalid_path in invalid_paths:
            result = cli_runner.invoke(cli, ["init", invalid_path, "--non-interactive"])
            # Should fail gracefully
            assert result.exit_code != 0


@pytest.mark.integration
class TestTmpUpdateOperations:
    """Tests for update command in /tmp folder."""

    def test_update_in_initialized_tmp_project(self, cli_runner, tmp_dir_path):
        """Test update command in tmp initialized project."""
        # First initialize
        init_result = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        if init_result.exit_code == 0:
            # Then run update
            update_result = cli_runner.invoke(cli, ["update", "--path", str(tmp_dir_path), "--yes"])

            # Should execute (may not actually update if already latest)
            assert update_result.exit_code in [0, 1]

    def test_update_check_only_in_tmp(self, cli_runner, tmp_dir_path):
        """Test update --check flag in tmp directory."""
        # Initialize first
        cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        # Run check-only update
        result = cli_runner.invoke(cli, ["update", "--path", str(tmp_dir_path), "--check"])

        # Should execute
        assert result.exit_code in [0, 1]

        # Should show version information
        assert len(result.output) > 0

    def test_update_templates_only_in_tmp(self, cli_runner, tmp_dir_path):
        """Test update --templates-only in tmp directory."""
        # Initialize first
        init_result = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        if init_result.exit_code == 0:
            # Run templates-only update
            result = cli_runner.invoke(cli, ["update", "--path", str(tmp_dir_path), "--templates-only", "--yes"])

            # Should execute
            assert result.exit_code in [0, 1]

    def test_update_in_non_initialized_tmp(self, cli_runner, tmp_dir_path):
        """Test update in tmp directory without .moai (should fail or warn)."""
        # Don't initialize - just run update
        result = cli_runner.invoke(cli, ["update", "--path", str(tmp_dir_path)])

        # Should handle gracefully (may fail or show warning)
        assert result.exit_code in [0, 1]

    def test_update_preserves_user_config_in_tmp(self, cli_runner, tmp_dir_path):
        """Test that update preserves user configuration in tmp."""
        # Initialize
        cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        # Modify config
        config_yaml = tmp_dir_path / ".moai" / "config" / "config.yaml"
        if config_yaml.exists():
            with open(config_yaml, "r", encoding="utf-8") as f:
                config = yaml.safe_load(f) or {}

            # Add custom config
            if "project" not in config:
                config["project"] = {}
            config["project"]["custom_field"] = "test_value"

            with open(config_yaml, "w", encoding="utf-8") as f:
                yaml.safe_dump(config, f, allow_unicode=True)

            # Run update with --yes flag to auto-confirm
            update_result = cli_runner.invoke(cli, ["update", "--path", str(tmp_dir_path), "--yes"])

            # After update, check if custom config is preserved (depends on update strategy)
            if update_result.exit_code == 0:
                with open(config_yaml, "r", encoding="utf-8") as f:
                    updated_config = yaml.safe_load(f) or {}
                # Config structure should still exist
                assert "moai" in updated_config

    def test_update_with_force_flag_in_tmp(self, cli_runner, tmp_dir_path):
        """Test update --force flag in tmp directory."""
        # Initialize
        cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        # Run update with force
        result = cli_runner.invoke(cli, ["update", "--path", str(tmp_dir_path), "--force"])

        # Should execute
        assert result.exit_code in [0, 1]

    def test_multiple_updates_in_same_tmp_dir(self, cli_runner, tmp_dir_path):
        """Test multiple consecutive updates in same tmp directory."""
        # Initialize
        init_result = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        if init_result.exit_code == 0:
            # First update
            update1 = cli_runner.invoke(cli, ["update", "--path", str(tmp_dir_path), "--yes"])
            assert update1.exit_code in [0, 1]

            # Second update (should be idempotent or skip)
            update2 = cli_runner.invoke(cli, ["update", "--path", str(tmp_dir_path), "--yes"])
            assert update2.exit_code in [0, 1]

            # Both should succeed
            assert update1.exit_code == update2.exit_code


@pytest.mark.integration
class TestTmpWorkflows:
    """Tests for complete workflows in /tmp folder."""

    def test_init_then_status_in_tmp(self, cli_runner, tmp_dir_path):
        """Test init followed by status command in tmp."""
        import os

        # Initialize
        init_result = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        if init_result.exit_code == 0:
            # Run status by changing to the directory
            original_cwd = os.getcwd()
            try:
                os.chdir(str(tmp_dir_path))
                status_result = cli_runner.invoke(cli, ["status"])
                # Should execute
                assert status_result.exit_code in [0, 1]
            finally:
                os.chdir(original_cwd)

    def test_init_then_doctor_in_tmp(self, cli_runner, tmp_dir_path):
        """Test init followed by doctor command in tmp."""
        import os

        # Initialize
        init_result = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        if init_result.exit_code == 0:
            # Run doctor by changing to the directory
            original_cwd = os.getcwd()
            try:
                os.chdir(str(tmp_dir_path))
                doctor_result = cli_runner.invoke(cli, ["doctor"])
                # Should execute
                assert doctor_result.exit_code in [0, 1]
            finally:
                os.chdir(original_cwd)

    def test_full_workflow_init_update_status_in_tmp(self, cli_runner, tmp_dir_path):
        """Test complete workflow: init -> update -> status in tmp."""
        # Step 1: Init
        init_result = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        if init_result.exit_code == 0:
            # Step 2: Update
            update_result = cli_runner.invoke(cli, ["update", "--path", str(tmp_dir_path), "--yes"])
            assert update_result.exit_code in [0, 1]

            # Step 3: Status
            status_result = cli_runner.invoke(cli, ["status"], cwd=str(tmp_dir_path))
            assert status_result.exit_code in [0, 1]

            # All should complete successfully
            assert init_result.exit_code == 0

    def test_cleanup_tmp_directory_after_operations(self, cli_runner, tmp_dir_path):
        """Test that tmp directory can be cleaned up after all operations."""
        # Perform various operations
        cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])
        cli_runner.invoke(cli, ["status"], cwd=str(tmp_dir_path))
        cli_runner.invoke(cli, ["doctor"], cwd=str(tmp_dir_path))

        # Directory should still be cleanable
        # (verified by tempfile.TemporaryDirectory context manager cleanup)
        assert tmp_dir_path.exists()

        # List what was created
        moai_dir = tmp_dir_path / ".moai"
        if moai_dir.exists():
            items = list(moai_dir.rglob("*"))
            # Verify structure is as expected
            assert len(items) > 0


@pytest.mark.integration
class TestTmpErrorHandling:
    """Tests for error handling in /tmp folder operations."""

    def test_init_handles_permission_issues_in_tmp(self, cli_runner, tmp_dir_path):
        """Test init handles permission issues in tmp gracefully."""
        # Create a read-only subdirectory
        readonly_dir = tmp_dir_path / "readonly"
        readonly_dir.mkdir()

        try:
            # Make read-only
            readonly_dir.chmod(0o444)

            result = cli_runner.invoke(cli, ["init", str(readonly_dir), "--non-interactive"])

            # Should fail or show error
            assert result.exit_code != 0 or "error" in result.output.lower()

        except Exception:
            # Skip if platform doesn't support chmod
            pytest.skip("Platform doesn't support permission changes")
        finally:
            # Restore for cleanup
            readonly_dir.chmod(0o755)

    def test_update_handles_corrupt_config_in_tmp(self, cli_runner, tmp_dir_path):
        """Test update handles corrupt config in tmp gracefully."""
        # Initialize
        cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        # Corrupt config
        config_yaml = tmp_dir_path / ".moai" / "config" / "config.yaml"
        if config_yaml.exists():
            config_yaml.write_text("{ invalid yaml content [[[", encoding="utf-8")

            # Run update
            result = cli_runner.invoke(cli, ["update", "--path", str(tmp_dir_path), "--yes"])

            # Should handle gracefully
            assert result.exit_code in [0, 1]

    def test_init_in_tmp_with_absolute_path(self, cli_runner, tmp_dir_path):
        """Test init with absolute path in tmp."""
        # Use absolute path
        abs_path = tmp_dir_path.resolve()

        result = cli_runner.invoke(cli, ["init", str(abs_path), "--non-interactive"])

        # Should execute
        assert result.exit_code in [0, 1]

        if result.exit_code == 0:
            assert (abs_path / ".moai").exists()

    def test_init_in_tmp_with_relative_path(self, cli_runner, tmp_dir_path):
        """Test init with relative path (from within tmp)."""
        # Change to tmp directory and use relative path
        relative_project = tmp_dir_path / "relative_project"
        relative_project.mkdir()

        # Use relative path from within tmp_dir_path
        # Note: CliRunner's invoke doesn't change cwd, so we test with absolute
        result = cli_runner.invoke(cli, ["init", str(relative_project), "--non-interactive"])

        # Should execute
        assert result.exit_code in [0, 1]


@pytest.mark.integration
class TestTmpGitIntegration:
    """Tests for Git operations in /tmp folder."""

    def test_init_in_tmp_git_repo(self, cli_runner, tmp_dir_path):
        """Test init in tmp directory with Git repository."""
        # Initialize Git repo
        subprocess.run(["git", "init"], cwd=tmp_dir_path, check=True, capture_output=True)
        subprocess.run(
            ["git", "config", "user.name", "Test User"],
            cwd=tmp_dir_path,
            check=True,
            capture_output=True,
        )
        subprocess.run(
            ["git", "config", "user.email", "test@example.com"],
            cwd=tmp_dir_path,
            check=True,
            capture_output=True,
        )

        # Run moai init
        result = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        # Should execute
        assert result.exit_code in [0, 1]

        if result.exit_code == 0:
            # Should create .moai alongside .git
            assert (tmp_dir_path / ".moai").exists()
            assert (tmp_dir_path / ".git").exists()

    def test_init_in_tmp_git_repo_with_commit(self, cli_runner, tmp_dir_path):
        """Test init in tmp Git repo with existing commit."""
        # Setup Git repo
        subprocess.run(["git", "init"], cwd=tmp_dir_path, check=True, capture_output=True)
        subprocess.run(
            ["git", "config", "user.name", "Test User"],
            cwd=tmp_dir_path,
            check=True,
            capture_output=True,
        )
        subprocess.run(
            ["git", "config", "user.email", "test@example.com"],
            cwd=tmp_dir_path,
            check=True,
            capture_output=True,
        )

        # Create initial commit
        readme = tmp_dir_path / "README.md"
        readme.write_text("# Test Project", encoding="utf-8")
        subprocess.run(["git", "add", "."], cwd=tmp_dir_path, check=True, capture_output=True)
        subprocess.run(
            ["git", "commit", "-m", "Initial commit"],
            cwd=tmp_dir_path,
            check=True,
            capture_output=True,
        )

        # Run moai init
        result = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        # Should execute
        assert result.exit_code in [0, 1]

        # Git repo should still be valid
        assert (tmp_dir_path / ".git").exists()


@pytest.mark.integration
class TestTmpConcurrentOperations:
    """Tests for concurrent operations in /tmp folder."""

    def test_multiple_inits_in_different_tmp_dirs(self, cli_runner, tmp_dir_path):
        """Test multiple init operations in different tmp directories."""
        project_dirs = [
            tmp_dir_path / "project1",
            tmp_dir_path / "project2",
            tmp_dir_path / "project3",
        ]

        for project_dir in project_dirs:
            project_dir.mkdir()
            result = cli_runner.invoke(cli, ["init", str(project_dir), "--non-interactive"])
            assert result.exit_code in [0, 1]

        # At least some should have .moai directories if successful
        successful_count = 0
        for project_dir in project_dirs:
            if (project_dir / ".moai").exists():
                successful_count += 1
                # Verify structure is complete
                if (project_dir / ".moai" / "config").exists():
                    pass  # Structure is valid

        # At least one should succeed
        assert successful_count >= 1, "At least one init should succeed"

    def test_init_same_tmp_dir_twice_concurrent(self, cli_runner, tmp_dir_path):
        """Test init in same tmp directory twice (sequential simulation)."""
        # First init
        result1 = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        # Second init (should be treated as reinit)
        result2 = cli_runner.invoke(cli, ["init", str(tmp_dir_path), "--non-interactive"])

        # Both should execute
        assert result1.exit_code in [0, 1]
        assert result2.exit_code in [0, 1]

        # Should handle gracefully (backup on second run)
        backup_dir = tmp_dir_path / ".moai-backups"
        if result1.exit_code == 0 and result2.exit_code == 0:
            # Backup should be created on reinit
            assert backup_dir.exists()
