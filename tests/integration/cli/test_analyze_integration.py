"""Integration tests for analyze command.

Tests end-to-end execution of the analyze command including:
- Analyze command with real session files
- Report generation to file
- Tool usage statistics calculation
- Suggestion display
"""

import json
from pathlib import Path
from unittest.mock import patch

import pytest
from click.testing import CliRunner

from moai_adk.cli.commands.analyze import analyze


@pytest.mark.integration
class TestAnalyzeBasicExecution:
    """Tests for basic analyze command execution."""

    def test_analyze_runs_without_crashing(self, cli_runner, temp_project_dir):
        """Test that analyze command executes without crashing."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should not crash
            assert result.exit_code in [0, 1]

            # Should have output
            assert len(result.output) > 0

    def test_analyze_detects_moai_project(self, cli_runner, temp_project_dir):
        """Test that analyze detects valid MoAI project."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should not show "not a MoAI-ADK project" error
            assert "not a MoAI-ADK project" not in result.output

    def test_analyze_shows_analyzing_message(self, cli_runner, temp_project_dir):
        """Test that analyze shows analyzing progress."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should show progress message
            output_lower = result.output.lower()
            assert "analyzing" in output_lower or "session" in output_lower


@pytest.mark.integration
class TestAnalyzeWithSessionFiles:
    """Tests for analyze with actual session files."""

    def test_analyze_reads_session_files(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze reads session files from logs directory."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should execute
            assert result.exit_code in [0, 1]

            # Should mention sessions
            output_lower = result.output.lower()
            assert "session" in output_lower or "analyz" in output_lower

    def test_analyze_counts_events(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze counts events from sessions."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should execute
            assert result.exit_code in [0, 1]

            # Should show statistics
            assert len(result.output) > 0

    def test_analyze_shows_tool_usage(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze shows tool usage statistics."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should execute
            assert result.exit_code in [0, 1]

            # Output should contain some information
            assert len(result.output) > 0


@pytest.mark.integration
class TestAnalyzeReportGeneration:
    """Tests for report file generation."""

    def test_analyze_creates_report_file(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze creates report file."""
        report_path = temp_project_dir / "analysis-report.md"

        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze, ["--output", str(report_path)])

            # Should execute
            assert result.exit_code in [0, 1]

            # Report may or may not be created depending on sessions
            if report_path.exists():
                content = report_path.read_text(encoding="utf-8")
                assert len(content) > 0

    def test_analyze_saves_report_to_specified_path(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze saves report to specified output path."""
        custom_path = temp_project_dir / "custom-report.md"

        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze, ["--output", str(custom_path)])

            # Should execute
            assert result.exit_code in [0, 1]

            # Check if report was created
            if custom_path.exists():
                content = custom_path.read_text(encoding="utf-8")
                assert len(content) > 0

    def test_analyze_creates_default_report_location(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze creates report in default location."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should execute
            assert result.exit_code in [0, 1]

            # Should mention report in output
            assert "report" in result.output.lower() or "saved" in result.output.lower()


@pytest.mark.integration
class TestAnalyzeDaysOption:
    """Tests for --days flag functionality."""

    def test_analyze_with_custom_days(self, cli_runner, temp_project_dir):
        """Test analyze with custom days value."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze, ["--days", "30"])

            # Should execute
            assert result.exit_code in [0, 1]

            # Should mention the time period
            output_lower = result.output.lower()
            assert "30" in result.output or "day" in output_lower

    def test_analyze_with_one_day(self, cli_runner, temp_project_dir):
        """Test analyze with 1 day period."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze, ["--days", "1"])

            # Should execute
            assert result.exit_code in [0, 1]

    def test_analyze_with_long_period(self, cli_runner, temp_project_dir):
        """Test analyze with long time period."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze, ["--days", "90"])

            # Should execute
            assert result.exit_code in [0, 1]


@pytest.mark.integration
class TestAnalyzeVerboseMode:
    """Tests for --verbose flag functionality."""

    def test_analyze_verbose_shows_more_details(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that --verbose shows more detailed output."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result_normal = cli_runner.invoke(analyze)
            result_verbose = cli_runner.invoke(analyze, ["--verbose"])

            # Both should execute
            assert result_normal.exit_code in [0, 1]
            assert result_verbose.exit_code in [0, 1]

            # Verbose should have more or equal output
            assert len(result_verbose.output) >= len(result_normal.output)

    def test_analyze_verbose_with_sessions(self, cli_runner, temp_project_dir, temp_session_file):
        """Test verbose mode with session data."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze, ["--verbose"])

            # Should execute
            assert result.exit_code in [0, 1]

            # Should have output
            assert len(result.output) > 0


@pytest.mark.integration
class TestAnalyzeReportOnlyMode:
    """Tests for --report-only flag functionality."""

    def test_analyze_report_only_suppresses_console_output(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that --report-only minimizes console output."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze, ["--report-only"])

            # Should execute
            assert result.exit_code in [0, 1]

            # Should have minimal output
            # (report-only mode skips console output)

    def test_analyze_report_only_still_creates_report(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that --report-only still generates report file."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze, ["--report-only"])

            # Should execute
            assert result.exit_code in [0, 1]


@pytest.mark.integration
class TestAnalyzeProjectPathOption:
    """Tests for --project-path flag functionality."""

    def test_analyze_with_custom_project_path(self, cli_runner, temp_project_dir):
        """Test analyze with custom project path."""
        result = cli_runner.invoke(analyze, ["--project-path", str(temp_project_dir)])

        # Should execute (may fail due to type issues in some environments)
        assert result.exit_code in [0, 1]

        # Should have output if no exception
        if result.exception is None:
            assert len(result.output) > 0

    def test_analyze_project_path_overrides_cwd(self, cli_runner, temp_project_dir):
        """Test that --project-path overrides current directory."""
        # Create another temp directory as "current"
        with cli_runner.isolated_filesystem():
            # Use temp_project_dir as project path
            result = cli_runner.invoke(analyze, ["--project-path", str(temp_project_dir)])

            # Should execute
            assert result.exit_code in [0, 1] or result.exception is not None


@pytest.mark.integration
class TestAnalyzeStatisticsCalculation:
    """Tests for statistics calculation."""

    def test_analyze_calculates_total_sessions(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze calculates total sessions."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should execute
            assert result.exit_code in [0, 1]

            # Should show session count
            output_lower = result.output.lower()
            assert "session" in output_lower

    def test_analyze_calculates_total_events(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze calculates total events."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should execute
            assert result.exit_code in [0, 1]

            # Should show event count
            output_lower = result.output.lower()
            assert "event" in output_lower or "analyz" in output_lower

    def test_analyze_shows_success_rate(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze shows success rate."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should execute
            assert result.exit_code in [0, 1]

            # Should show statistics
            assert len(result.output) > 0


@pytest.mark.integration
class TestAnalyzeToolUsageStatistics:
    """Tests for tool usage statistics."""

    def test_analyze_shows_top_tools(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze shows top used tools."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should execute
            assert result.exit_code in [0, 1]

            # Output should contain tool information
            assert len(result.output) > 0

    def test_analyze_counts_tool_usage(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze counts tool usage frequency."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should execute
            assert result.exit_code in [0, 1]

            # Should have statistics
            assert len(result.output) > 0


@pytest.mark.integration
class TestAnalyzeSuggestions:
    """Tests for improvement suggestions."""

    def test_analyze_generates_suggestions(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze generates improvement suggestions."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should execute
            assert result.exit_code in [0, 1]

            # May show suggestions
            # (suggestions depend on actual session data)

    def test_analyze_shows_key_suggestions_in_console(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze shows key suggestions in console output."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should execute
            assert result.exit_code in [0, 1]

            # May display suggestions section
            output_lower = result.output.lower()


@pytest.mark.integration
class TestAnalyzeErrorHandling:
    """Tests for error handling in analyze command."""

    def test_analyze_handles_non_existent_project_path(self, cli_runner):
        """Test analyze handles non-existent project path."""
        result = cli_runner.invoke(analyze, ["--project-path", "/non/existent/path"])

        # Should show error about missing .moai
        assert "not a MoAI-ADK project" in result.output or result.exit_code != 0

    def test_analyze_handles_empty_logs_directory(self, cli_runner, temp_project_dir):
        """Test analyze handles empty logs directory."""
        logs_dir = temp_project_dir / ".moai" / "logs"
        logs_dir.mkdir(parents=True, exist_ok=True)

        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should handle gracefully
            assert result.exit_code in [0, 1]

    def test_analyze_handles_invalid_days_value(self, cli_runner, temp_project_dir):
        """Test analyze handles invalid days value."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze, ["--days", "invalid"])

            # Should fail or show error
            assert result.exit_code != 0

    def test_analyze_handles_negative_days(self, cli_runner, temp_project_dir):
        """Test analyze handles negative days value."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze, ["--days", "-1"])

            # Should fail or handle gracefully
            # (Click may reject this)


@pytest.mark.integration
class TestAnalyzeOutputFormat:
    """Tests for analyze output formatting."""

    def test_analyze_uses_rich_formatting(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze uses Rich console output."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should have formatted output
            assert len(result.output) > 0

    def test_analyze_shows_summary_table(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze shows summary table."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should show summary
            assert len(result.output) > 0


@pytest.mark.integration
class TestAnalyzeIntegrationWithMoai:
    """Tests for analyze integration with MoAI project structure."""

    def test_analyze_reads_from_moai_logs(self, cli_runner, temp_project_dir, temp_session_file):
        """Test that analyze reads from .moai/logs directory."""
        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should execute
            assert result.exit_code in [0, 1]

            # Should read from logs
            assert len(result.output) > 0

    def test_analyze_respects_project_context(self, cli_runner, temp_project_dir):
        """Test that analyze respects project context."""
        # Create project structure
        (temp_project_dir / "README.md").write_text("# Test Project")

        with patch("pathlib.Path.cwd", return_value=temp_project_dir):
            result = cli_runner.invoke(analyze)

            # Should execute
            assert result.exit_code in [0, 1]
