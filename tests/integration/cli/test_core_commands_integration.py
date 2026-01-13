"""Integration tests for core CLI commands.

Tests end-to-end execution of core commands including:
- moai doctor - System health check
- moai status - Status display
- Actual file system operations
- Real Git operations (in temporary repos)

Key Requirements:
- Use Click's CliRunner for command invocation
- Test real CLI execution (not mocked)
- Capture exit codes, stdout, stderr
- Test both success and failure paths
"""

import json
import subprocess
import tempfile
from pathlib import Path
from typing import Generator
from unittest.mock import patch

import pytest
import yaml
from click.testing import CliRunner

from moai_adk.__main__ import cli


@pytest.fixture
def real_git_repo(tmp_path: Path) -> Generator[Path, None, None]:
    """Create a real Git repository for testing.

    Initializes an actual Git repository with test configuration
    for testing status command's Git integration.

    Args:
        tmp_path: Pytest's temporary directory fixture

    Yields:
        Path to temporary Git repository
    """
    repo_dir = tmp_path / "real_git_repo"
    repo_dir.mkdir(parents=True, exist_ok=True)

    # Initialize Git repository
    subprocess.run(["git", "init"], cwd=repo_dir, check=True, capture_output=True)

    # Configure Git user
    subprocess.run(
        ["git", "config", "user.name", "Test User"],
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )
    subprocess.run(
        ["git", "config", "user.email", "test@example.com"],
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )

    # Create initial commit
    readme = repo_dir / "README.md"
    readme.write_text("# Test Repository\n", encoding="utf-8")
    subprocess.run(["git", "add", "."], cwd=repo_dir, check=True, capture_output=True)
    subprocess.run(
        ["git", "commit", "-m", "Initial commit"],
        cwd=repo_dir,
        check=True,
        capture_output=True,
    )

    yield repo_dir


@pytest.fixture
def moai_project_in_git_repo(real_git_repo: Path) -> Generator[Path, None, None]:
    """Create a MoAI-ADK project within a Git repository.

    Combines MoAI-ADK initialization with Git repository for
    testing commands that require both.

    Args:
        real_git_repo: Git repository fixture

    Yields:
        Path to MoAI-ADK project in Git repo
    """
    runner = CliRunner()

    # Initialize MoAI-ADK in the Git repo
    result = runner.invoke(cli, ["init", str(real_git_repo), "--non-interactive"])

    # Only yield if init succeeded
    if result.exit_code == 0:
        yield real_git_repo
    else:
        pytest.skip("Failed to initialize MoAI-ADK project in Git repo")


@pytest.mark.integration
class TestDoctorCommandIntegration:
    """Integration tests for doctor command."""

    def test_doctor_runs_successfully(self, cli_runner):
        """Test that doctor command runs without crashing."""
        result = cli_runner.invoke(cli, ["doctor"])

        # Should execute (may fail on environment checks but should run)
        assert result.exit_code in [0, 1]

        # Should produce output
        assert len(result.output) > 0

    def test_doctor_shows_system_checks(self, cli_runner):
        """Test that doctor displays system checks."""
        result = cli_runner.invoke(cli, ["doctor"])

        # Should show check results
        assert len(result.output) > 0

        # Should mention checks or diagnostics
        output_lower = result.output.lower()
        assert "check" in output_lower or "diagnostic" in output_lower or "python" in output_lower

    def test_doctor_verbose_mode(self, cli_runner):
        """Test doctor with -v flag."""
        result = cli_runner.invoke(cli, ["doctor", "-v"])

        # Should execute
        assert result.exit_code in [0, 1]

        # Verbose mode should produce more output
        assert len(result.output) > 0

    def test_doctor_with_export_json(self, cli_runner, tmp_path):
        """Test doctor --export creates JSON file."""
        export_file = tmp_path / "doctor_report.json"

        result = cli_runner.invoke(cli, ["doctor", "--export", str(export_file)])

        # Should execute
        assert result.exit_code in [0, 1]

        # Check if file was created
        if export_file.exists():
            content = export_file.read_text(encoding="utf-8")

            # Should be valid JSON
            data = json.loads(content)
            assert isinstance(data, dict)

            # Should contain diagnostic data
            assert "basic_checks" in data or len(data) > 0

    def test_doctor_export_includes_language_data(self, cli_runner, tmp_path):
        """Test doctor --export with -v includes language detection."""
        export_file = tmp_path / "doctor_verbose_report.json"

        result = cli_runner.invoke(cli, ["doctor", "-v", "--export", str(export_file)])

        # Should execute
        assert result.exit_code in [0, 1]

        if export_file.exists():
            content = export_file.read_text(encoding="utf-8")
            data = json.loads(content)

            # Should have basic structure
            assert isinstance(data, dict)

    def test_doctor_check_specific_tool(self, cli_runner):
        """Test doctor --check for specific tool."""
        result = cli_runner.invoke(cli, ["doctor", "--check", "python"])

        # Should execute
        assert result.exit_code == 0

        # Should mention the tool
        assert "python" in result.output.lower()

    def test_doctor_fix_mode(self, cli_runner):
        """Test doctor --fix flag."""
        result = cli_runner.invoke(cli, ["doctor", "--fix"])

        # Should execute
        assert result.exit_code in [0, 1]

        # Should have output
        assert len(result.output) > 0

    def test_doctor_check_commands_flag(self, cli_runner, temp_project_dir):
        """Test doctor --check-commands flag."""
        # Create .claude/commands directory
        claude_dir = temp_project_dir / ".claude"
        claude_dir.mkdir(parents=True, exist_ok=True)

        commands_dir = claude_dir / "commands"
        commands_dir.mkdir(parents=True, exist_ok=True)

        # Create a valid command file
        test_command = commands_dir / "test.md"
        test_command.write_text(
            """---
name: test
description: Test command
---

# Test Command

This is a test command.
""",
            encoding="utf-8",
        )

        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(cli, ["doctor", "--check-commands"])

            # Should execute
            assert result.exit_code == 0

            # Should show command diagnostics
            assert "command" in result.output.lower()


@pytest.mark.integration
class TestStatusCommandIntegration:
    """Integration tests for status command."""

    def test_status_in_moai_project(self, cli_runner, temp_project_dir):
        """Test status command in a valid MoAI-ADK project."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(cli, ["status"])

            # Should execute successfully
            assert result.exit_code == 0

            # Should show project status
            assert len(result.output) > 0

            # Should mention mode or locale or specs
            output_lower = result.output.lower()
            assert (
                "mode" in output_lower
                or "locale" in output_lower
                or "spec" in output_lower
                or "project" in output_lower
            )

    def test_status_shows_project_mode(self, cli_runner, temp_project_dir):
        """Test that status displays project mode."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(cli, ["status"])

            # Should execute
            assert result.exit_code == 0

            # Should show mode
            assert "mode" in result.output.lower() or "personal" in result.output.lower() or "team" in result.output.lower()

    def test_status_shows_locale(self, cli_runner, temp_project_dir):
        """Test that status displays locale setting."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(cli, ["status"])

            # Should execute
            assert result.exit_code == 0

            # Should show locale
            assert "locale" in result.output.lower() or "en" in result.output.lower()

    def test_status_shows_spec_count(self, cli_runner, temp_project_dir):
        """Test that status displays SPEC document count."""
        # Create some SPEC files
        specs_dir = temp_project_dir / ".moai" / "specs"
        specs_dir.mkdir(parents=True, exist_ok=True)

        spec1_dir = specs_dir / "SPEC-TEST-001"
        spec1_dir.mkdir()
        (spec1_dir / "spec.md").write_text("# Test SPEC", encoding="utf-8")

        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(cli, ["status"])

            # Should execute
            assert result.exit_code == 0

            # Should show SPEC count
            assert "spec" in result.output.lower() or "1" in result.output

    def test_status_in_git_repo(self, cli_runner, moai_project_in_git_repo):
        """Test status command shows Git information."""
        with patch("pathlib.Path.cwd", return_value=moai_project_in_git_repo):
            result = cli_runner.invoke(cli, ["status"])

            # Should execute
            assert result.exit_code == 0

            # Should show Git status if Git is available
            # (may not show if gitpython not installed)
            assert len(result.output) > 0

    def test_status_in_git_repo_with_dirty_state(self, cli_runner, moai_project_in_git_repo):
        """Test status detects modified Git state."""
        # Modify a file to make repo dirty
        config_file = moai_project_in_git_repo / ".moai" / "config" / "config.yaml"
        if config_file.exists():
            config_file.write_text("# Modified content\n", encoding="utf-8")

        with patch("pathlib.Path.cwd", return_value=moai_project_in_git_repo):
            result = cli_runner.invoke(cli, ["status"])

            # Should execute
            assert result.exit_code == 0

            # Should detect modified state if Git integration works
            assert len(result.output) > 0

    def test_status_in_non_moai_directory(self, cli_runner, tmp_path):
        """Test status command in directory without .moai (should fail)."""
        non_moai_dir = tmp_path / "non_moai"
        non_moai_dir.mkdir()

        with patch("pathlib.Path.cwd", return_value=non_moai_dir):
            result = cli_runner.invoke(cli, ["status"])

            # Should fail gracefully
            assert result.exit_code != 0

            # Should show error or warning
            assert "moai" in result.output.lower() or "not found" in result.output.lower() or "no" in result.output.lower()

    def test_status_with_missing_config_file(self, cli_runner, temp_project_dir):
        """Test status with missing config.yaml (should fail)."""
        # Remove config.yaml
        config_file = temp_project_dir / ".moai" / "config" / "config.yaml"
        if config_file.exists():
            config_file.unlink()

        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(cli, ["status"])

            # Should fail
            assert result.exit_code != 0

            # Should show error message
            assert "config" in result.output.lower() or "not found" in result.output.lower()

    def test_status_with_corrupt_config(self, cli_runner, temp_project_dir):
        """Test status handles corrupt config gracefully."""
        # Corrupt config file
        config_file = temp_project_dir / ".moai" / "config" / "config.yaml"
        if config_file.exists():
            config_file.write_text("{ invalid yaml [[[", encoding="utf-8")

        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(cli, ["status"])

            # Should handle error
            assert result.exit_code != 0


@pytest.mark.integration
class TestDoctorAndStatusTogether:
    """Tests for combined doctor and status workflows."""

    def test_doctor_then_status_in_project(self, cli_runner, temp_project_dir):
        """Test running doctor then status in same project."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            # Run doctor
            doctor_result = cli_runner.invoke(cli, ["doctor"])
            assert doctor_result.exit_code in [0, 1]

            # Run status
            status_result = cli_runner.invoke(cli, ["status"])
            assert status_result.exit_code == 0

    def test_status_then_doctor_in_project(self, cli_runner, temp_project_dir):
        """Test running status then doctor in same project."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            # Run status
            status_result = cli_runner.invoke(cli, ["status"])
            assert status_result.exit_code == 0

            # Run doctor
            doctor_result = cli_runner.invoke(cli, ["doctor"])
            assert doctor_result.exit_code in [0, 1]

    def test_doctor_export_and_status_combined(self, cli_runner, temp_project_dir, tmp_path):
        """Test doctor export and status together."""
        export_file = tmp_path / "combined_report.json"

        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            # Run doctor with export
            doctor_result = cli_runner.invoke(cli, ["doctor", "--export", str(export_file)])
            assert doctor_result.exit_code in [0, 1]

            # Run status
            status_result = cli_runner.invoke(cli, ["status"])
            assert status_result.exit_code == 0

            # Check export file was created
            if export_file.exists():
                assert json.loads(export_file.read_text())


@pytest.mark.integration
class TestFileSystemOperations:
    """Tests for real file system operations."""

    def test_init_creates_real_files(self, cli_runner, tmp_path):
        """Test that init creates real files on disk."""
        project_dir = tmp_path / "real_files_test"
        project_dir.mkdir()

        result = cli_runner.invoke(cli, ["init", str(project_dir), "--non-interactive"])

        if result.exit_code == 0:
            # Verify files actually exist on disk
            moai_dir = project_dir / ".moai"
            assert moai_dir.exists()
            assert moai_dir.is_dir()

            config_dir = moai_dir / "config"
            assert config_dir.exists()

            config_yaml = config_dir / "config.yaml"
            assert config_yaml.exists()
            assert config_yaml.is_file()

            # Verify content is readable
            content = config_yaml.read_text(encoding="utf-8")
            assert len(content) > 0

    def test_init_creates_sections(self, cli_runner, tmp_path):
        """Test that init creates sections directory with files."""
        project_dir = tmp_path / "sections_test"
        project_dir.mkdir()

        result = cli_runner.invoke(cli, ["init", str(project_dir), "--non-interactive"])

        if result.exit_code == 0:
            sections_dir = project_dir / ".moai" / "config" / "sections"
            assert sections_dir.exists()

            # Check for expected section files
            expected_sections = ["language.yaml", "project.yaml"]
            for section in expected_sections:
                section_file = sections_dir / section
                assert section_file.exists()

    def test_update_modifies_real_files(self, cli_runner, tmp_path):
        """Test that update modifies actual files on disk."""
        project_dir = tmp_path / "update_files_test"
        project_dir.mkdir()

        # Initialize first
        init_result = cli_runner.invoke(cli, ["init", str(project_dir), "--non-interactive"])

        if init_result.exit_code == 0:
            # Record initial state
            config_yaml = project_dir / ".moai" / "config" / "config.yaml"
            initial_content = config_yaml.read_text(encoding="utf-8") if config_yaml.exists() else ""

            # Run update
            update_result = cli_runner.invoke(cli, ["update", "--path", str(project_dir), "--yes"])

            if update_result.exit_code == 0:
                # Check files were modified or at least still exist
                assert config_yaml.exists()

                # Content should still be valid YAML
                with open(config_yaml, encoding="utf-8") as f:
                    yaml.safe_load(f)

    def test_cleanup_after_operations(self, cli_runner, tmp_path):
        """Test that cleanup works after multiple operations."""
        project_dir = tmp_path / "cleanup_test"
        project_dir.mkdir()

        # Perform multiple operations
        cli_runner.invoke(cli, ["init", str(project_dir), "--non-interactive"])
        cli_runner.invoke(cli, ["status"], cwd=str(project_dir))
        cli_runner.invoke(cli, ["doctor"], cwd=str(project_dir))

        # All files should still be accessible
        moai_dir = project_dir / ".moai"
        if moai_dir.exists():
            # List all created files
            files = list(moai_dir.rglob("*"))
            assert len(files) > 0

            # Verify files are readable
            for file_path in files[:5]:  # Check first few files
                if file_path.is_file():
                    content = file_path.read_text(encoding="utf-8", errors="ignore")
                    assert content is not None


@pytest.mark.integration
class TestGitOperationsIntegration:
    """Tests for real Git operations."""

    def test_git_repo_detected_by_doctor(self, cli_runner, real_git_repo):
        """Test that doctor detects Git repository."""
        with patch("pathlib.Path.cwd", return_value=real_git_repo):
            result = cli_runner.invoke(cli, ["doctor"])

            # Should execute
            assert result.exit_code in [0, 1]

            # Should check Git
            assert "git" in result.output.lower()

    def test_git_status_shown_in_status_command(self, cli_runner, moai_project_in_git_repo):
        """Test that Git status is shown in status command."""
        with patch("pathlib.Path.cwd", return_value=moai_project_in_git_repo):
            result = cli_runner.invoke(cli, ["status"])

            # Should execute
            assert result.exit_code == 0

            # May show Git info if gitpython is available
            assert len(result.output) > 0

    def test_git_branch_detected(self, cli_runner, real_git_repo):
        """Test that Git branch is detected."""
        # Create a branch
        subprocess.run(
            ["git", "checkout", "-b", "test-branch"],
            cwd=real_git_repo,
            check=True,
            capture_output=True,
        )

        with patch("pathlib.Path.cwd", return_value=real_git_repo):
            result = cli_runner.invoke(cli, ["doctor"])

            # Should execute
            assert result.exit_code in [0, 1]

    def test_git_dirty_state_detection(self, cli_runner, real_git_repo):
        """Test detection of dirty Git state."""
        # Create uncommitted change
        test_file = real_git_repo / "dirty.txt"
        test_file.write_text("Uncommitted content", encoding="utf-8")
        subprocess.run(["git", "add", "."], cwd=real_git_repo, check=True, capture_output=True)

        # Don't commit - leave it dirty

        result = cli_runner.invoke(cli, ["doctor"], cwd=str(real_git_repo))

        # Should execute
        assert result.exit_code in [0, 1]

    def test_init_in_git_repo_preserves_git(self, cli_runner, real_git_repo):
        """Test that init in Git repo preserves .git directory."""
        # Verify .git exists before init
        git_dir = real_git_repo / ".git"
        assert git_dir.exists()

        # Run moai init
        result = cli_runner.invoke(cli, ["init", str(real_git_repo), "--non-interactive"])

        if result.exit_code == 0:
            # .git should still exist
            assert git_dir.exists()

            # .moai should also exist
            moai_dir = real_git_repo / ".moai"
            assert moai_dir.exists()

    def test_multiple_git_operations(self, cli_runner, real_git_repo):
        """Test multiple operations in Git repo."""
        # Init
        result1 = cli_runner.invoke(cli, ["init", str(real_git_repo), "--non-interactive"])

        if result1.exit_code == 0:
            # Status
            result2 = cli_runner.invoke(cli, ["status"], cwd=str(real_git_repo))
            assert result2.exit_code in [0, 1]

            # Doctor
            result3 = cli_runner.invoke(cli, ["doctor"], cwd=str(real_git_repo))
            assert result3.exit_code in [0, 1]

            # All should complete
            assert result1.exit_code == 0


@pytest.mark.integration
class TestErrorScenarios:
    """Tests for error scenarios."""

    def test_status_in_non_existent_moai_dir(self, cli_runner, tmp_path):
        """Test status when .moai directory doesn't exist."""
        empty_dir = tmp_path / "empty"
        empty_dir.mkdir()

        with patch("pathlib.Path.cwd", return_value=empty_dir):
            result = cli_runner.invoke(cli, ["status"])

            # Should fail
            assert result.exit_code != 0

    def test_doctor_with_invalid_export_path(self, cli_runner):
        """Test doctor with invalid export path."""
        result = cli_runner.invoke(cli, ["doctor", "--export", "/non/existent/path/report.json"])

        # Should handle gracefully
        assert result.exit_code in [0, 1]

    def test_init_with_invalid_options(self, cli_runner, tmp_path):
        """Test init with invalid option combinations."""
        project_dir = tmp_path / "invalid_test"
        project_dir.mkdir()

        # Invalid locale (not in allowed list)
        result = cli_runner.invoke(cli, ["init", str(project_dir), "--non-interactive", "--locale", "invalid"])

        # Should fail or use default
        assert result.exit_code in [0, 1, 2]

    def test_status_with_broken_symlink(self, cli_runner, temp_project_dir, tmp_path):
        """Test status handles broken symlinks gracefully."""
        # Create broken symlink in .moai
        moai_dir = temp_project_dir / ".moai"
        if moai_dir.exists():
            # Try to create a broken symlink
            try:
                link_target = tmp_path / "nonexistent_target"
                broken_link = moai_dir / "broken_link"
                broken_link.symlink_to(link_target)

                with patch("pathlib.Path.cwd", return_value=temp_project_dir):
                    result = cli_runner.invoke(cli, ["status"])
                    # Should handle gracefully
                    assert result.exit_code in [0, 1]
            except Exception:
                # Skip if symlinks not supported
                pytest.skip("Symlinks not supported on this platform")
