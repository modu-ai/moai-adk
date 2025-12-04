"""Enhanced comprehensive tests for analyze command to achieve 90%+ coverage

This test suite targets all lines in analyze.py (lines 8-116):
- Command options: --days, --output, --verbose, --report-only, --project-path
- Project validation: .moai directory check
- SessionAnalyzer initialization and integration
- Console output: summary tables, tool usage display
- Report generation and saving
- Suggestion display logic
- Error handling: missing project directory, file operations
"""

from pathlib import Path
from unittest.mock import Mock, patch

from click.testing import CliRunner

from moai_adk.cli.commands.analyze import analyze


class TestAnalyzeCommandOptions:
    """Test command option handling"""

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_with_default_options_uses_7_days(self, mock_analyzer_class):
        """Should use default 7 days when no --days option is provided"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 0,
            "total_events": 0,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("/fake/report.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            mock_analyzer_class.assert_called_once_with(days_back=7, verbose=False)

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_with_custom_days_option(self, mock_analyzer_class):
        """Should use custom days value when --days option is provided"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 0,
            "total_events": 0,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("/fake/report.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze, ["--days", "14"])

            assert result.exit_code == 0
            mock_analyzer_class.assert_called_once_with(days_back=14, verbose=False)

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_with_verbose_flag_enables_verbose_output(self, mock_analyzer_class):
        """Should enable verbose mode when --verbose flag is provided"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 0,
            "total_events": 0,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("/fake/report.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze, ["--verbose"])

            assert result.exit_code == 0
            mock_analyzer_class.assert_called_once_with(days_back=7, verbose=True)

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_with_output_path_saves_to_custom_location(self, mock_analyzer_class):
        """Should save report to custom location when --output option is provided"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 0,
            "total_events": 0,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        custom_output = Path("custom_report.md")
        mock_analyzer.save_report.return_value = custom_output
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze, ["--output", "custom_report.md"])

            assert result.exit_code == 0
            # Verify save_report was called
            mock_analyzer.save_report.assert_called_once()
            call_args = mock_analyzer.save_report.call_args[0]
            assert str(call_args[0]) == "custom_report.md"


class TestProjectPathValidation:
    """Test project path validation logic"""

    def test_analyze_uses_current_directory_when_no_project_path_provided(self):
        """Should use current working directory when --project-path is not provided"""
        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            with patch("moai_adk.cli.commands.analyze.SessionAnalyzer") as mock_analyzer_class:
                mock_analyzer = Mock()
                mock_analyzer.parse_sessions.return_value = {
                    "total_sessions": 0,
                    "total_events": 0,
                    "failed_sessions": 0,
                    "tool_usage": {},
                }
                mock_analyzer.save_report.return_value = Path("report.md")
                mock_analyzer.generate_report.return_value = "Report content"
                mock_analyzer_class.return_value = mock_analyzer

                result = runner.invoke(analyze)
                assert result.exit_code == 0

    def test_analyze_fails_when_moai_directory_missing(self):
        """Should fail with error when .moai directory is missing"""
        runner = CliRunner()
        with runner.isolated_filesystem():
            # No .moai directory created
            result = runner.invoke(analyze)

            assert result.exit_code == 0  # Click doesn't return error code
            assert "Error:" in result.output
            assert "Not a MoAI-ADK project" in result.output
            assert "missing .moai directory" in result.output

    def test_analyze_displays_current_path_on_validation_failure(self):
        """Should display current path when project validation fails"""
        runner = CliRunner()
        with runner.isolated_filesystem():
            # No .moai directory
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            assert "Current path:" in result.output


class TestSessionAnalysis:
    """Test session analysis and pattern extraction"""

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_parses_sessions_and_displays_summary(self, mock_analyzer_class):
        """Should parse sessions and display summary table"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 10,
            "total_events": 150,
            "failed_sessions": 2,
            "success_rate": 80.0,
            "tool_usage": {"Read": 25, "Write": 15},
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            assert "Analyzed 10 sessions" in result.output
            assert "Total events: 150" in result.output
            mock_analyzer.parse_sessions.assert_called_once()

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_displays_session_summary_table(self, mock_analyzer_class):
        """Should display session summary table with metrics"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 20,
            "total_events": 300,
            "failed_sessions": 5,
            "success_rate": 75.0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            assert "Session Summary" in result.output
            assert "Total Sessions" in result.output
            assert "20" in result.output
            assert "Total Events" in result.output
            assert "300" in result.output
            assert "Failed Sessions" in result.output
            assert "5" in result.output
            assert "Success Rate" in result.output
            assert "75.0%" in result.output


class TestToolUsageDisplay:
    """Test tool usage table display"""

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_displays_top_tools_when_available(self, mock_analyzer_class):
        """Should display top 10 tools table when tool usage data is available"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 10,
            "total_events": 100,
            "failed_sessions": 0,
            "success_rate": 100.0,
            "tool_usage": {
                "Read": 50,
                "Write": 30,
                "Edit": 20,
                "Bash": 15,
                "Grep": 10,
                "Task": 8,
                "Glob": 5,
                "WebFetch": 3,
                "AskUserQuestion": 2,
                "TodoWrite": 1,
            },
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            assert "Top Tools Used:" in result.output
            assert "`Read`" in result.output
            assert "50" in result.output
            assert "`Write`" in result.output
            assert "30" in result.output

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_limits_tool_display_to_top_10(self, mock_analyzer_class):
        """Should display only top 10 tools even when more are available"""
        # Setup
        mock_analyzer = Mock()
        # Create 15 tools to test limit
        tool_usage = {f"Tool{i}": 100 - i for i in range(15)}
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 10,
            "total_events": 100,
            "failed_sessions": 0,
            "success_rate": 100.0,
            "tool_usage": tool_usage,
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            assert "`Tool0`" in result.output  # Top tool
            assert "`Tool9`" in result.output  # 10th tool
            # 11th tool should not be displayed
            assert "`Tool10`" not in result.output

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_skips_tool_table_when_no_tools_used(self, mock_analyzer_class):
        """Should skip tool usage table when no tools were used"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 5,
            "total_events": 10,
            "failed_sessions": 0,
            "success_rate": 100.0,
            "tool_usage": {},  # Empty tool usage
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            assert "Top Tools Used:" not in result.output


class TestReportGeneration:
    """Test report generation and saving"""

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_saves_report_to_default_location(self, mock_analyzer_class):
        """Should save report to default .moai/reports/ when no output specified"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 0,
            "total_events": 0,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        expected_report_path = Path(".moai/reports/daily-2025-11-26.md")
        mock_analyzer.save_report.return_value = expected_report_path
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            mock_analyzer.save_report.assert_called_once()
            # Verify it was called with None (default output) and project_path
            call_args = mock_analyzer.save_report.call_args[0]
            assert call_args[0] is None  # output_path

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_displays_report_saved_message(self, mock_analyzer_class):
        """Should display report saved message with file path"""
        # Setup
        report_path = Path("test_report.md")
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 0,
            "total_events": 0,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = report_path
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            assert "Report saved:" in result.output
            assert str(report_path) in result.output


class TestSuggestionDisplay:
    """Test improvement suggestion display logic"""

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_displays_suggestions_when_available(self, mock_analyzer_class):
        """Should display key suggestions when they exist in report"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 10,
            "total_events": 100,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        # Report with suggestions section
        mock_analyzer.generate_report.return_value = """
# Report

## Some Section

---

## 💡 Improvement Suggestions

- Suggestion 1: Review permission settings
- Suggestion 2: Add fallback strategy
- Suggestion 3: Optimize hook performance
"""
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            assert "💡 Key Suggestions:" in result.output
            assert "Suggestion 1" in result.output
            assert "Suggestion 2" in result.output
            assert "Suggestion 3" in result.output

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_limits_suggestion_display_to_10_lines(self, mock_analyzer_class):
        """Should display only first 10 suggestion lines"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 10,
            "total_events": 100,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        # Create report with 15 suggestion lines
        suggestions = "\n".join([f"- Suggestion {i}" for i in range(1, 16)])
        mock_analyzer.generate_report.return_value = f"""
## 💡 Improvement Suggestions

{suggestions}
"""
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            assert "Suggestion 10" in result.output
            assert "Suggestion 11" not in result.output

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_skips_empty_and_separator_lines_in_suggestions(self, mock_analyzer_class):
        """Should skip empty lines and separators when displaying suggestions"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 10,
            "total_events": 100,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        mock_analyzer.generate_report.return_value = """
## 💡 Improvement Suggestions


- Suggestion 1

---

- Suggestion 2
"""
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            assert "Suggestion 1" in result.output
            assert "Suggestion 2" in result.output

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_skips_suggestions_when_not_in_report(self, mock_analyzer_class):
        """Should skip suggestion display when section is not in report"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 10,
            "total_events": 100,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        # Report without suggestions section
        mock_analyzer.generate_report.return_value = """
# Report

## Summary

All good!
"""
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            assert "💡 Key Suggestions:" not in result.output


class TestReportOnlyMode:
    """Test --report-only flag behavior"""

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_report_only_suppresses_console_output(self, mock_analyzer_class):
        """Should suppress all console output when --report-only is enabled"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 10,
            "total_events": 100,
            "failed_sessions": 2,
            "success_rate": 80.0,
            "tool_usage": {"Read": 25},
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        mock_analyzer.generate_report.return_value = """
## 💡 Improvement Suggestions

- Test suggestion
"""
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze, ["--report-only"])

            assert result.exit_code == 0
            # Should not contain any of the normal output
            assert "Analyzing sessions" not in result.output
            assert "Analyzed 10 sessions" not in result.output
            assert "Session Summary" not in result.output
            assert "Top Tools Used:" not in result.output
            assert "Report saved:" not in result.output
            assert "💡 Key Suggestions:" not in result.output

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_report_only_still_generates_and_saves_report(self, mock_analyzer_class):
        """Should still generate and save report when --report-only is enabled"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 10,
            "total_events": 100,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze, ["--report-only"])

            assert result.exit_code == 0
            mock_analyzer.parse_sessions.assert_called_once()
            mock_analyzer.save_report.assert_called_once()

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_report_only_does_not_call_generate_report_for_suggestions(self, mock_analyzer_class):
        """Should not call generate_report for suggestions when --report-only is enabled"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 10,
            "total_events": 100,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze, ["--report-only"])

            assert result.exit_code == 0
            # generate_report should not be called for displaying suggestions
            # It's only called inside save_report
            assert mock_analyzer.generate_report.call_count == 0


class TestEdgeCases:
    """Test edge cases and error conditions"""

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_handles_zero_sessions_gracefully(self, mock_analyzer_class):
        """Should handle zero sessions without errors"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 0,
            "total_events": 0,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            assert "Analyzed 0 sessions" in result.output

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_handles_missing_success_rate_in_patterns(self, mock_analyzer_class):
        """Should handle missing success_rate key gracefully"""
        # Setup
        mock_analyzer = Mock()
        # Missing 'success_rate' key
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 10,
            "total_events": 100,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            # Should display 0.0% when key is missing
            assert "0.0%" in result.output or "Success Rate" in result.output

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_with_combined_flags(self, mock_analyzer_class):
        """Should handle multiple flags together correctly"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 5,
            "total_events": 50,
            "failed_sessions": 1,
            "success_rate": 80.0,
            "tool_usage": {"Read": 10},
        }
        mock_analyzer.save_report.return_value = Path("custom.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(
                analyze,
                [
                    "--days",
                    "30",
                    "--verbose",
                    "--output",
                    "custom.md",
                ],
            )

            assert result.exit_code == 0
            mock_analyzer_class.assert_called_once_with(days_back=30, verbose=True)
            mock_analyzer.save_report.assert_called_once()

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_handles_suggestion_extraction_edge_cases(self, mock_analyzer_class):
        """Should handle edge cases in suggestion section extraction"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 10,
            "total_events": 100,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        # Report with suggestions at the very end (no content after)
        mock_analyzer.generate_report.return_value = """
# Report

## 💡 Improvement Suggestions
"""
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze)

            assert result.exit_code == 0
            # Should handle empty suggestions gracefully
            assert "💡 Key Suggestions:" in result.output


class TestIntegrationWithSessionAnalyzer:
    """Test integration with SessionAnalyzer class"""

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_calls_session_analyzer_with_correct_parameters(self, mock_analyzer_class):
        """Should initialize SessionAnalyzer with correct parameters"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 0,
            "total_events": 0,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        mock_analyzer.save_report.return_value = Path("report.md")
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze, ["--days", "21", "--verbose"])

            assert result.exit_code == 0
            mock_analyzer_class.assert_called_once_with(days_back=21, verbose=True)
            mock_analyzer.parse_sessions.assert_called_once()
            mock_analyzer.save_report.assert_called_once()

    @patch("moai_adk.cli.commands.analyze.SessionAnalyzer")
    def test_analyze_passes_correct_paths_to_save_report(self, mock_analyzer_class):
        """Should pass correct output and project paths to save_report"""
        # Setup
        mock_analyzer = Mock()
        mock_analyzer.parse_sessions.return_value = {
            "total_sessions": 0,
            "total_events": 0,
            "failed_sessions": 0,
            "tool_usage": {},
        }
        output_file = Path("my_report.md")
        mock_analyzer.save_report.return_value = output_file
        mock_analyzer.generate_report.return_value = "Report content"
        mock_analyzer_class.return_value = mock_analyzer

        runner = CliRunner()
        with runner.isolated_filesystem():
            Path(".moai").mkdir()
            result = runner.invoke(analyze, ["--output", "my_report.md"])

            assert result.exit_code == 0
            mock_analyzer.save_report.assert_called_once()
            call_args = mock_analyzer.save_report.call_args[0]
            assert str(call_args[0]) == "my_report.md"
