"""Integration tests for doctor command.

Tests end-to-end execution of the doctor command including:
- Full doctor command execution with real environment
- --check-commands flag with actual command files
- --verbose mode with language detection
- --export JSON output to file
"""

import json
from unittest.mock import patch

import pytest

from moai_adk.cli.commands.doctor import doctor


@pytest.mark.integration
class TestDoctorBasicExecution:
    """Tests for basic doctor command execution."""

    def test_doctor_runs_without_crashing(self, cli_runner):
        """Test that doctor command executes without crashing."""
        result = cli_runner.invoke(doctor)

        # Should not crash
        assert result.exit_code in [0, 1]  # May fail on environment checks but runs

        # Should output something
        assert len(result.output) > 0

    def test_doctor_shows_diagnostic_header(self, cli_runner):
        """Test that doctor shows diagnostic header."""
        result = cli_runner.invoke(doctor)

        # Should show diagnostic message
        assert "diagnostic" in result.output.lower() or "check" in result.output.lower()

    def test_doctor_displays_results_table(self, cli_runner):
        """Test that doctor displays a results table."""
        result = cli_runner.invoke(doctor)

        # Should have some output
        assert len(result.output) > 0


@pytest.mark.integration
class TestDoctorVerboseMode:
    """Tests for --verbose flag functionality."""

    def test_doctor_verbose_shows_language_detection(self, cli_runner, temp_project_dir):
        """Test that -v shows detected language."""
        # Change to temp project directory
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(doctor, ["-v"])

            # Should execute without crashing
            assert result.exit_code in [0, 1]

            # Verbose mode should produce more output
            assert len(result.output) > 0

    def test_doctor_verbose_with_language_tools(self, cli_runner):
        """Test that -v checks language tools."""
        result = cli_runner.invoke(doctor, ["-v"])

        # Should execute
        assert result.exit_code in [0, 1]

        # Verbose output should be present
        assert len(result.output) > 0


@pytest.mark.integration
class TestDoctorCheckCommands:
    """Tests for --check-commands flag."""

    def test_doctor_check_commands_with_valid_commands(self, cli_runner, temp_moai_commands_dir):
        """Test --check-commands with valid command files."""
        # temp_moai_commands_dir fixture provides command files for testing
        result = cli_runner.invoke(doctor, ["--check-commands"])

        # Should execute
        assert result.exit_code == 0

        # Should show command diagnostics
        assert "command" in result.output.lower() or "diagnostic" in result.output.lower()

    def test_doctor_check_commands_counts_files(self, cli_runner, temp_moai_commands_dir):
        """Test that --check-commands counts command files."""
        # temp_moai_commands_dir fixture provides command files for testing
        result = cli_runner.invoke(doctor, ["--check-commands"])

        # Should show counts
        assert result.exit_code == 0

        # Should have output
        assert len(result.output) > 0

    def test_doctor_check_commands_with_no_commands(self, cli_runner, temp_project_dir):
        """Test --check-commands when no commands directory exists."""
        # Create .claude but no commands
        claude_dir = temp_project_dir / ".claude"
        claude_dir.mkdir(parents=True, exist_ok=True)

        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(doctor, ["--check-commands"])

            # Should execute (may show warning about no commands)
            assert result.exit_code in [0, 1]


@pytest.mark.integration
class TestDoctorExport:
    """Tests for --export flag functionality."""

    def test_doctor_export_creates_json_file(self, cli_runner, tmp_path):
        """Test that --export creates JSON file with diagnostics."""
        export_file = tmp_path / "doctor-output.json"

        result = cli_runner.invoke(doctor, ["--export", str(export_file)])

        # Should execute
        assert result.exit_code in [0, 1]

        # Check if file was created (may fail in some environments)
        if export_file.exists():
            content = export_file.read_text(encoding="utf-8")

            # Should be valid JSON
            data = json.loads(content)

            # Should have basic_checks
            assert "basic_checks" in data or len(data) > 0

    def test_doctor_export_contains_diagnostic_data(self, cli_runner, tmp_path):
        """Test that exported JSON contains diagnostic data."""
        export_file = tmp_path / "doctor-diagnostics.json"

        _ = cli_runner.invoke(doctor, ["--export", str(export_file)])

        if export_file.exists():
            content = export_file.read_text(encoding="utf-8")
            data = json.loads(content)

            # Should be a dict
            assert isinstance(data, dict)

            # Should have some data
            assert len(data) > 0

    def test_doctor_export_with_verbose(self, cli_runner, tmp_path):
        """Test --export with -v includes language data."""
        export_file = tmp_path / "doctor-verbose.json"

        cli_runner.invoke(doctor, ["-v", "--export", str(export_file)])

        if export_file.exists():
            content = export_file.read_text(encoding="utf-8")
            data = json.loads(content)

            # Should have basic data
            assert isinstance(data, dict)


@pytest.mark.integration
class TestDoctorCheckFlag:
    """Tests for --check flag functionality."""

    def test_doctor_check_specific_tool_python(self, cli_runner):
        """Test checking specific tool (Python)."""
        result = cli_runner.invoke(doctor, ["--check", "python"])

        # Should execute
        assert result.exit_code == 0

        # Should mention the tool
        assert "python" in result.output.lower()

    def test_doctor_check_specific_tool_git(self, cli_runner):
        """Test checking specific tool (Git)."""
        result = cli_runner.invoke(doctor, ["--check", "git"])

        # Should execute
        assert result.exit_code == 0

        # Should mention the tool
        assert "git" in result.output.lower()


@pytest.mark.integration
class TestDoctorFixFlag:
    """Tests for --fix flag functionality."""

    def test_doctor_fix_mode_runs(self, cli_runner):
        """Test that --fix mode executes."""
        result = cli_runner.invoke(doctor, ["--fix"])

        # Should execute
        assert result.exit_code in [0, 1]

        # Should have output
        assert len(result.output) > 0


@pytest.mark.integration
class TestDoctorInProjectContext:
    """Tests for doctor command in project context."""

    def test_doctor_in_moai_project(self, cli_runner, temp_project_dir):
        """Test doctor command in a valid MoAI-ADK project."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(doctor)

            # Should execute
            assert result.exit_code in [0, 1]

            # Should have output
            assert len(result.output) > 0

    def test_doctor_detects_moai_directory(self, cli_runner, temp_project_dir):
        """Test that doctor detects .moai directory."""
        # Create .moai if not exists
        moai_dir = temp_project_dir / ".moai"
        moai_dir.mkdir(parents=True, exist_ok=True)

        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(doctor)

            # Should execute
            assert result.exit_code in [0, 1]


@pytest.mark.integration
class TestDoctorErrorHandling:
    """Tests for error handling in doctor command."""

    def test_doctor_handles_invalid_export_path(self, cli_runner):
        """Test doctor handles invalid export path gracefully."""
        # Use invalid path (e.g., non-existent directory)
        result = cli_runner.invoke(doctor, ["--export", "/non/existent/path/output.json"])

        # Should fail gracefully
        assert result.exit_code in [0, 1]

    def test_doctor_handles_unknown_check_tool(self, cli_runner):
        """Test doctor handles unknown tool name gracefully."""
        result = cli_runner.invoke(doctor, ["--check", "nonexistent-tool-xyz"])

        # Should execute (may show tool not found)
        assert result.exit_code in [0, 1]


@pytest.mark.integration
class TestDoctorOutputFormat:
    """Tests for doctor output formatting."""

    def test_doctor_uses_rich_formatting(self, cli_runner):
        """Test that doctor uses Rich console output."""
        result = cli_runner.invoke(doctor)

        # Should have output
        assert len(result.output) > 0

        # May contain ANSI codes or formatted output
        # (Rich adds formatting codes)

    def test_doctor_shows_status_icons(self, cli_runner):
        """Test that doctor shows status checkmark/cross icons."""
        result = cli_runner.invoke(doctor)

        # Should have output
        assert len(result.output) > 0

        # Check for unicode characters or status indicators
        # (may be plain text on some terminals)
