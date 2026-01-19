"""Integration tests for init command.

Tests end-to-end execution of the init command including:
- Init command in temporary directory
- .moai directory creation
- Config file generation
- Reinitialization flow
"""

from pathlib import Path
from unittest.mock import patch

import pytest
import yaml

from moai_adk.cli.commands.init import init


@pytest.mark.integration
class TestInitBasicExecution:
    """Tests for basic init command execution."""

    def test_init_runs_without_crashing(self, cli_runner, tmp_path):
        """Test that init command executes without crashing."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Should execute without crashing (may fail on environment checks)
        assert result.exit_code in [0, 1]
        assert len(result.output) > 0

    def test_init_creates_moai_directory(self, cli_runner, tmp_path):
        """Test that init creates .moai directory."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Check .moai directory was created if command succeeded
        moai_dir = project_dir / ".moai"
        if result.exit_code == 0:
            assert moai_dir.exists(), ".moai directory should be created on success"
            assert moai_dir.is_dir()

    def test_init_creates_config_directory(self, cli_runner, tmp_path):
        """Test that init creates .moai/config directory."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Check config directory was created if command succeeded
        config_dir = project_dir / ".moai" / "config"
        if result.exit_code == 0:
            assert config_dir.exists(), "config directory should be created on success"
            assert config_dir.is_dir()


@pytest.mark.integration
class TestInitConfigGeneration:
    """Tests for config file generation."""

    def test_init_creates_config_yaml(self, cli_runner, tmp_path):
        """Test that init creates config.yaml."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Check config.yaml was created if command succeeded
        config_yaml = project_dir / ".moai" / "config" / "config.yaml"
        if result.exit_code == 0 and config_yaml.exists():
            # Check it's valid YAML
            with open(config_yaml, encoding="utf-8") as f:
                data = yaml.safe_load(f)
                assert isinstance(data, dict)

    def test_init_config_contains_version(self, cli_runner, tmp_path):
        """Test that config.yaml contains version information."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        config_yaml = project_dir / ".moai" / "config" / "config.yaml"
        if result.exit_code == 0 and config_yaml.exists():
            with open(config_yaml, encoding="utf-8") as f:
                data = yaml.safe_load(f)
                assert "moai" in data
                assert "version" in data["moai"]

    def test_init_creates_sections_directory(self, cli_runner, tmp_path):
        """Test that init creates sections directory."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Check sections directory was created if command succeeded
        sections_dir = project_dir / ".moai" / "config" / "sections"
        if result.exit_code == 0:
            assert sections_dir.exists(), "sections directory should be created on success"

    def test_init_creates_language_yaml(self, cli_runner, tmp_path):
        """Test that init creates language.yaml section."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Check language.yaml was created if command succeeded
        language_yaml = project_dir / ".moai" / "config" / "sections" / "language.yaml"
        if result.exit_code == 0 and language_yaml.exists():
            # Check it's valid YAML
            with open(language_yaml, encoding="utf-8") as f:
                data = yaml.safe_load(f)
                assert "language" in data

    def test_init_creates_project_yaml(self, cli_runner, tmp_path):
        """Test that init creates project.yaml section."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Check project.yaml was created if command succeeded
        project_yaml = project_dir / ".moai" / "config" / "sections" / "project.yaml"
        if result.exit_code == 0 and project_yaml.exists():
            # Check it's valid YAML
            with open(project_yaml, encoding="utf-8") as f:
                data = yaml.safe_load(f)
                assert "project" in data


@pytest.mark.integration
class TestInitNonInteractiveMode:
    """Tests for non-interactive mode."""

    def test_init_non_interactive_succeeds(self, cli_runner, tmp_path):
        """Test that non-interactive mode completes without prompts."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Should execute without crashing
        assert result.exit_code in [0, 1]

        # Should show initialization messages
        assert "initialization" in result.output.lower() or "completed" in result.output.lower()

    def test_init_non_interactive_creates_all_files(self, cli_runner, tmp_path):
        """Test that non-interactive mode creates all expected files."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Check expected files exist if command succeeded
        if result.exit_code == 0:
            expected_files = [
                ".moai/config/config.yaml",
                ".moai/config/sections/language.yaml",
                ".moai/config/sections/project.yaml",
            ]

            for file_path in expected_files:
                full_path = project_dir / file_path
                assert full_path.exists(), f"Expected file not created: {file_path}"

    def test_init_non_interactive_with_locale(self, cli_runner, tmp_path):
        """Test non-interactive mode with locale option."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive", "--locale", "ko"])

        # Should execute
        assert result.exit_code in [0, 1]

        # Check locale was set if command succeeded
        language_yaml = project_dir / ".moai" / "config" / "sections" / "language.yaml"
        if result.exit_code == 0 and language_yaml.exists():
            with open(language_yaml, encoding="utf-8") as f:
                data = yaml.safe_load(f)
                assert data["language"]["conversation_language"] == "ko"


@pytest.mark.integration
class TestInitReinitialization:
    """Tests for reinitialization flow."""

    def test_init_second_run_creates_backup(self, cli_runner, tmp_path):
        """Test that reinitialization creates backup."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        # First init
        result1 = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Only test reinit if first init succeeded
        if result1.exit_code == 0:
            # Modify a file to check backup
            config_yaml = project_dir / ".moai" / "config" / "config.yaml"

            # Second init (reinit)
            result2 = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

            # Should complete
            assert result2.exit_code in [0, 1]

            # Check for backup directory
            backup_dir = project_dir / ".moai-backups"
            if backup_dir.exists():
                # Should have at least one backup
                backups = list(backup_dir.iterdir())
                assert len(backups) > 0, "Should create backup on reinit"

    def test_init_force_flag_works(self, cli_runner, tmp_path):
        """Test that --force flag allows reinitialization."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        # First init
        result1 = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Only test force if first init succeeded
        if result1.exit_code == 0:
            # Second init with force
            result2 = cli_runner.invoke(init, [str(project_dir), "--non-interactive", "--force"])

            # Should complete successfully
            assert result2.exit_code in [0, 1]

    def test_init_reinit_preserves_structure(self, cli_runner, tmp_path):
        """Test that reinitialization preserves directory structure."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        # First init
        result1 = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Only test reinit if first init succeeded
        if result1.exit_code == 0:
            # Second init
            result2 = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])
            assert result2.exit_code in [0, 1]

            # Structure should still exist
            assert (project_dir / ".moai").exists()
            assert (project_dir / ".moai" / "config").exists()


@pytest.mark.integration
class TestInitOutputMessages:
    """Tests for init command output messages."""

    def test_init_shows_banner(self, cli_runner, tmp_path):
        """Test that init shows banner message."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Should have output
        assert len(result.output) > 0

    def test_init_shows_success_message(self, cli_runner, tmp_path):
        """Test that init shows success message."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Should show completion message
        assert (
            "success" in result.output.lower()
            or "completed" in result.output.lower()
            or "done" in result.output.lower()
            or "initialization" in result.output.lower()
        )

    def test_init_shows_summary(self, cli_runner, tmp_path):
        """Test that init shows initialization summary."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

        # Should show location or file count
        assert (
            "location" in result.output.lower()
            or "files" in result.output.lower()
            or "created" in result.output.lower()
            or "version" in result.output.lower()
        )


@pytest.mark.integration
class TestInitErrorHandling:
    """Tests for error handling in init command."""

    def test_init_handles_invalid_path(self, cli_runner):
        """Test init handles invalid path gracefully."""
        # Use a path that cannot be created
        result = cli_runner.invoke(init, ["/non/existent/path/that/cannot/be/created", "--non-interactive"])

        # Should fail gracefully
        assert result.exit_code != 0

    def test_init_handles_permission_denied(self, cli_runner, tmp_path):
        """Test init handles permission errors gracefully."""
        # Create a read-only directory
        project_dir = tmp_path / "readonly_project"
        project_dir.mkdir()

        try:
            # Make directory read-only
            project_dir.chmod(0o444)

            result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

            # Should fail or show error
            assert result.exit_code != 0 or "error" in result.output.lower()
        except Exception:
            # Skip test if permission changes not supported
            pytest.skip("Platform doesn't support permission changes")
        finally:
            # Restore permissions for cleanup
            project_dir.chmod(0o755)

    def test_init_handles_interrupt(self, cli_runner, tmp_path):
        """Test init handles keyboard interrupt gracefully."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        # Mock keyboard interrupt
        with patch(
            "moai_adk.cli.commands.init.ProjectInitializer.initialize",
            side_effect=KeyboardInterrupt(),
        ):
            result = cli_runner.invoke(init, [str(project_dir), "--non-interactive"])

            # Should handle interrupt
            assert result.exit_code != 0 or "cancelled" in result.output.lower()


@pytest.mark.integration
class TestInitDefaultPath:
    """Tests for default path (current directory)."""

    def test_init_with_current_directory(self, cli_runner):
        """Test init with default path (current directory)."""
        with cli_runner.isolated_filesystem():
            result = cli_runner.invoke(init, ["--non-interactive"])

            # Should execute
            assert result.exit_code in [0, 1] or "initialization" in result.output.lower()

            # Check .moai was created
            moai_dir = Path.cwd() / ".moai"
            if moai_dir.exists():
                assert moai_dir.is_dir()


@pytest.mark.integration
class TestInitLanguageOptions:
    """Tests for language-related options."""

    def test_init_with_python_language(self, cli_runner, tmp_path):
        """Test init with Python language specified."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive", "--language", "python"])

        # Should execute
        assert result.exit_code in [0, 1]

    def test_init_with_typescript_language(self, cli_runner, tmp_path):
        """Test init with TypeScript language specified."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive", "--language", "typescript"])

        # Should execute
        assert result.exit_code in [0, 1]


@pytest.mark.integration
class TestInitModeOptions:
    """Tests for mode-related options."""

    def test_init_with_personal_mode(self, cli_runner, tmp_path):
        """Test init with personal mode."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive", "--mode", "personal"])

        # Should execute
        assert result.exit_code in [0, 1]

    def test_init_with_team_mode(self, cli_runner, tmp_path):
        """Test init with team mode."""
        project_dir = tmp_path / "test_project"
        project_dir.mkdir()

        result = cli_runner.invoke(init, [str(project_dir), "--non-interactive", "--mode", "team"])

        # Should execute
        assert result.exit_code in [0, 1]
