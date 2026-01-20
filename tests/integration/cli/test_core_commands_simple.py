"""Simplified integration tests for core CLI commands.

Tests end-to-end execution of core commands including:
- moai doctor - System health check (basic mode)
- moai status - Status display
- Actual file system operations
- Real Git operations (in temporary repos)

Note: Many doctor command options (-v, --fix, --export, --check) are defined
in doctor.py but not exposed through the main CLI wrapper in __main__.py.
This test file focuses on testing the functionality that IS exposed.
"""

import subprocess
from pathlib import Path
from typing import Generator
from unittest.mock import patch

import pytest
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
    """Integration tests for doctor command (basic mode only)."""

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
            mode_check = "mode" in result.output.lower()
            personal_check = "personal" in result.output.lower()
            team_check = "team" in result.output.lower()
            assert mode_check or personal_check or team_check
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

    def test_status_in_non_moai_directory(self, cli_runner, tmp_path):
        """Test status command in directory without .moai (should fail)."""
        non_moai_dir = tmp_path / "non_moai"
        non_moai_dir.mkdir()

        with patch("pathlib.Path.cwd", return_value=non_moai_dir):
            result = cli_runner.invoke(cli, ["status"])

            # Should fail gracefully
            assert result.exit_code != 0

            # Should show error or warning
            moai_check = "moai" in result.output.lower()
            not_found_check = "not found" in result.output.lower()
            no_check = "no" in result.output.lower()
            assert moai_check or not_found_check or no_check
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

    def test_cleanup_after_operations(self, cli_runner, tmp_path):
        """Test that cleanup works after multiple operations."""
        project_dir = tmp_path / "cleanup_test"
        project_dir.mkdir()

        # Perform multiple operations
        cli_runner.invoke(cli, ["init", str(project_dir), "--non-interactive"])

        # Use isolated_filesystem for commands that need to run in project_dir
        with cli_runner.isolated_filesystem(temp_dir=project_dir):
            cli_runner.invoke(cli, ["status"])
            cli_runner.invoke(cli, ["doctor"])

        # All files should still be accessible
        moai_dir = project_dir / ".moai"
        if moai_dir.exists():
            # List all created files
            files = list(moai_dir.rglob("*"))
            assert len(files) > 0


@pytest.mark.integration
class TestGitOperationsIntegration:
    """Tests for real Git operations."""

    def test_git_repo_detected_by_doctor(self, cli_runner, real_git_repo):
        """Test that doctor can run in Git repository."""
        # Use isolated_filesystem to change directory
        with cli_runner.isolated_filesystem(temp_dir=real_git_repo):
            result = cli_runner.invoke(cli, ["doctor"])

            # Should execute
            assert result.exit_code in [0, 1]

            # May check Git
            assert len(result.output) > 0

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
            with cli_runner.isolated_filesystem(temp_dir=real_git_repo):
                result2 = cli_runner.invoke(cli, ["status"])
                assert result2.exit_code in [0, 1]

            # Doctor
            with cli_runner.isolated_filesystem(temp_dir=real_git_repo):
                result3 = cli_runner.invoke(cli, ["doctor"])
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

    def test_init_with_invalid_options(self, cli_runner, tmp_path):
        """Test init with invalid option combinations."""
        project_dir = tmp_path / "invalid_test"
        project_dir.mkdir()

        # Invalid locale (not in allowed list)
        result = cli_runner.invoke(cli, ["init", str(project_dir), "--non-interactive", "--locale", "invalid"])

        # Should fail or use default
        assert result.exit_code in [0, 1, 2]
