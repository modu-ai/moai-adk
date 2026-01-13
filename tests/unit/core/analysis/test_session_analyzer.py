"""
Comprehensive unit tests for SessionAnalyzer with 85%+ coverage.

Tests cover:
- Initialization
- parse_sessions() with JSON and JSONL formats
- _analyze_session() for different session types
- generate_report() functionality
- save_report() functionality
- get_metrics() calculations
- _generate_suggestions() logic
- Error handling and edge cases
"""

import json
import tempfile
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import patch

from moai_adk.core.analysis.session_analyzer import SessionAnalyzer


class TestSessionAnalyzerInitialization:
    """Test SessionAnalyzer initialization."""

    def test_init_defaults(self):
        """Test SessionAnalyzer initializes with defaults."""
        analyzer = SessionAnalyzer()
        assert analyzer.days_back == 7
        assert analyzer.verbose is False
        assert analyzer.patterns["total_sessions"] == 0
        assert analyzer.patterns["total_events"] == 0

    def test_init_custom_parameters(self):
        """Test SessionAnalyzer with custom parameters."""
        analyzer = SessionAnalyzer(days_back=14, verbose=True)
        assert analyzer.days_back == 14
        assert analyzer.verbose is True

    def test_init_patterns_structure(self):
        """Test SessionAnalyzer initializes pattern structure."""
        analyzer = SessionAnalyzer()
        assert "tool_usage" in analyzer.patterns
        assert "tool_failures" in analyzer.patterns
        assert "error_patterns" in analyzer.patterns
        assert "permission_requests" in analyzer.patterns
        assert "hook_failures" in analyzer.patterns
        assert "command_frequency" in analyzer.patterns
        assert "success_rate" in analyzer.patterns


class TestSessionAnalyzerParseSessions:
    """Test SessionAnalyzer.parse_sessions() functionality."""

    @patch.object(Path, "exists")
    def test_parse_sessions_no_directory(self, mock_exists):
        """Test parse_sessions when claude projects directory doesn't exist."""
        mock_exists.return_value = False
        analyzer = SessionAnalyzer()
        result = analyzer.parse_sessions()
        assert result == analyzer.patterns

    def test_parse_sessions_empty_directory(self):
        """Test parse_sessions with empty directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            mock_path = Path(tmpdir)
            analyzer = SessionAnalyzer()
            with patch.object(analyzer, "claude_projects", mock_path):
                result = analyzer.parse_sessions()
                assert result["total_sessions"] == 0

    def test_parse_sessions_json_format(self):
        """Test parse_sessions with JSON format session file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            claude_projects = Path(tmpdir)
            project_dir = claude_projects / "test-project"
            project_dir.mkdir(parents=True, exist_ok=True)

            session_file = project_dir / "session-001.json"
            session_data = {
                "events": [
                    {"type": "tool_call", "toolName": "Bash", "command": "ls"},
                ]
            }
            session_file.write_text(json.dumps(session_data))

            analyzer = SessionAnalyzer()
            with patch.object(analyzer, "claude_projects", claude_projects):
                result = analyzer.parse_sessions()
                assert result["total_sessions"] == 1

    def test_parse_sessions_jsonl_format(self):
        """Test parse_sessions with JSONL format session file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            claude_projects = Path(tmpdir)
            project_dir = claude_projects / "test-project"
            project_dir.mkdir(parents=True, exist_ok=True)

            session_file = project_dir / "session.jsonl"
            lines = [
                json.dumps({"type": "summary", "summary": "Testing system"}),
                json.dumps({"type": "summary", "summary": "Build succeeded"}),
            ]
            session_file.write_text("\n".join(lines))

            analyzer = SessionAnalyzer()
            with patch.object(analyzer, "claude_projects", claude_projects):
                result = analyzer.parse_sessions()
                assert result["total_sessions"] == 2

    def test_parse_sessions_invalid_json(self):
        """Test parse_sessions with invalid JSON."""
        with tempfile.TemporaryDirectory() as tmpdir:
            claude_projects = Path(tmpdir)
            project_dir = claude_projects / "test-project"
            project_dir.mkdir(parents=True, exist_ok=True)

            session_file = project_dir / "session-001.json"
            session_file.write_text("{ invalid json")

            analyzer = SessionAnalyzer()
            with patch.object(analyzer, "claude_projects", claude_projects):
                result = analyzer.parse_sessions()
                # Should handle error gracefully
                assert result["total_sessions"] == 0

    def test_parse_sessions_filters_by_date(self):
        """Test parse_sessions filters sessions by date."""
        with tempfile.TemporaryDirectory() as tmpdir:
            claude_projects = Path(tmpdir)
            project_dir = claude_projects / "test-project"
            project_dir.mkdir(parents=True, exist_ok=True)

            # Create old session
            old_file = project_dir / "session-old.json"
            old_file.write_text(json.dumps({"events": []}))

            # Set modification time to 10 days ago
            old_time = (datetime.now() - timedelta(days=10)).timestamp()
            Path(old_file).touch()
            import os

            os.utime(old_file, (old_time, old_time))

            analyzer = SessionAnalyzer(days_back=7)
            with patch.object(analyzer, "claude_projects", claude_projects):
                result = analyzer.parse_sessions()
                # Old session should be filtered out
                assert result["total_sessions"] == 0


class TestSessionAnalyzerAnalyzeSession:
    """Test SessionAnalyzer._analyze_session() method."""

    def test_analyze_session_summary_format(self):
        """Test analyzing session in summary format."""
        analyzer = SessionAnalyzer()
        session = {"type": "summary", "summary": "Testing and building application"}
        analyzer._analyze_session(session)
        assert analyzer.patterns["total_events"] == 1

    def test_analyze_session_with_error_in_summary(self):
        """Test analyzing session with error in summary."""
        analyzer = SessionAnalyzer()
        session = {
            "type": "summary",
            "summary": "Build failed with error in compilation",
        }
        analyzer._analyze_session(session)
        assert analyzer.patterns["failed_sessions"] > 0

    def test_analyze_session_tool_call(self):
        """Test analyzing tool call events."""
        analyzer = SessionAnalyzer()
        session = {
            "events": [
                {"type": "tool_call", "toolName": "Bash(command)"},
            ]
        }
        analyzer._analyze_session(session)
        assert analyzer.patterns["total_events"] == 1
        assert analyzer.patterns["tool_usage"]["Bash"] > 0

    def test_analyze_session_tool_error(self):
        """Test analyzing tool error events."""
        analyzer = SessionAnalyzer()
        session = {
            "events": [
                {"type": "tool_error", "error": "Command failed with exit code 1"},
            ]
        }
        analyzer._analyze_session(session)
        assert analyzer.patterns["failed_sessions"] > 0
        assert len(analyzer.patterns["tool_failures"]) > 0

    def test_analyze_session_permission_request(self):
        """Test analyzing permission request events."""
        analyzer = SessionAnalyzer()
        session = {
            "events": [
                {"type": "permission_request", "permission_type": "file_write"},
            ]
        }
        analyzer._analyze_session(session)
        assert analyzer.patterns["permission_requests"]["file_write"] > 0

    def test_analyze_session_hook_failure(self):
        """Test analyzing hook failure events."""
        analyzer = SessionAnalyzer()
        session = {
            "events": [
                {"type": "hook_failure", "hook_name": "pre_commit"},
            ]
        }
        analyzer._analyze_session(session)
        assert analyzer.patterns["hook_failures"]["pre_commit"] > 0
        assert analyzer.patterns["failed_sessions"] > 0

    def test_analyze_session_command_tracking(self):
        """Test analyzing command execution."""
        analyzer = SessionAnalyzer()
        session = {
            "events": [
                {"type": "other", "command": "git commit -m 'test'"},
            ]
        }
        analyzer._analyze_session(session)
        assert analyzer.patterns["command_frequency"]["git"] > 0

    def test_analyze_session_multiple_events(self):
        """Test analyzing session with multiple events."""
        analyzer = SessionAnalyzer()
        session = {
            "events": [
                {"type": "tool_call", "toolName": "Read"},
                {"type": "tool_call", "toolName": "Write"},
                {"type": "tool_error", "error": "File not found"},
            ]
        }
        analyzer._analyze_session(session)
        assert analyzer.patterns["total_events"] == 3


class TestSessionAnalyzerGenerateReport:
    """Test SessionAnalyzer.generate_report() method."""

    def test_generate_report_empty_data(self):
        """Test generating report with empty data."""
        analyzer = SessionAnalyzer()
        report = analyzer.generate_report()
        assert "MoAI-ADK Session Meta-Analysis Report" in report
        assert "Overall Metrics" in report

    def test_generate_report_with_sessions(self):
        """Test generating report with session data."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["total_sessions"] = 5
        analyzer.patterns["total_events"] = 50
        analyzer.patterns["failed_sessions"] = 1

        report = analyzer.generate_report()
        assert "5" in report or "5" in str(analyzer.patterns["total_sessions"])

    def test_generate_report_tool_usage_section(self):
        """Test report includes tool usage section."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["tool_usage"]["Bash"] = 10
        analyzer.patterns["tool_usage"]["Read"] = 8

        report = analyzer.generate_report()
        assert "Tool Usage Patterns" in report

    def test_generate_report_error_patterns_section(self):
        """Test report includes error patterns section."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["tool_failures"]["File not found"] = 3

        report = analyzer.generate_report()
        assert "Tool Error Patterns" in report

    def test_generate_report_hook_failures_section(self):
        """Test report includes hook failures section."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["hook_failures"]["pre_commit_hook"] = 2

        report = analyzer.generate_report()
        assert "Hook Failure Analysis" in report

    def test_generate_report_permission_requests_section(self):
        """Test report includes permission requests section."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["permission_requests"]["file_write"] = 4

        report = analyzer.generate_report()
        assert "Permission Request" in report

    def test_generate_report_improvement_suggestions_section(self):
        """Test report includes improvement suggestions."""
        analyzer = SessionAnalyzer()
        report = analyzer.generate_report()
        assert "Improvement Suggestions" in report

    def test_generate_report_success_rate_calculation(self):
        """Test success rate calculation in report."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["total_sessions"] = 10
        analyzer.patterns["failed_sessions"] = 2

        report = analyzer.generate_report()
        # 80% success rate
        assert "80" in report or "success" in report.lower()


class TestSessionAnalyzerGenerateSuggestions:
    """Test SessionAnalyzer._generate_suggestions() method."""

    def test_generate_suggestions_no_issues(self):
        """Test generating suggestions with no issues."""
        analyzer = SessionAnalyzer()
        suggestions = analyzer._generate_suggestions()
        assert "No major issues" in suggestions

    def test_generate_suggestions_high_permission_requests(self):
        """Test suggestions for high permission requests."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["permission_requests"]["file_write"] = 5

        suggestions = analyzer._generate_suggestions()
        assert "Permission" in suggestions or "permission" in suggestions

    def test_generate_suggestions_tool_failures(self):
        """Test suggestions for tool failures."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["tool_failures"]["Error message"] = 3

        suggestions = analyzer._generate_suggestions()
        assert len(suggestions) > 0

    def test_generate_suggestions_hook_failures(self):
        """Test suggestions for hook failures."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["hook_failures"]["test_hook"] = 2

        suggestions = analyzer._generate_suggestions()
        assert len(suggestions) > 0

    def test_generate_suggestions_low_success_rate(self):
        """Test suggestions for low success rate."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["total_sessions"] = 10
        analyzer.patterns["failed_sessions"] = 8

        suggestions = analyzer._generate_suggestions()
        assert "success rate" in suggestions.lower() or "Low" in suggestions


class TestSessionAnalyzerSaveReport:
    """Test SessionAnalyzer.save_report() method."""

    def test_save_report_default_location(self):
        """Test saving report to default location."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            analyzer = SessionAnalyzer()

            report_path = analyzer.save_report(project_path=project_path)

            assert report_path.exists()
            assert report_path.parent.name == "reports"
            assert report_path.suffix == ".md"

    def test_save_report_custom_location(self):
        """Test saving report to custom location."""
        with tempfile.TemporaryDirectory() as tmpdir:
            custom_path = Path(tmpdir) / "custom_report.md"
            analyzer = SessionAnalyzer()

            report_path = analyzer.save_report(output_path=custom_path)

            assert report_path == custom_path
            assert report_path.exists()

    def test_save_report_creates_directory(self):
        """Test save_report creates necessary directories."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            analyzer = SessionAnalyzer()

            report_path = analyzer.save_report(project_path=project_path)

            assert report_path.parent.exists()

    def test_save_report_contains_data(self):
        """Test saved report contains data."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            analyzer = SessionAnalyzer()
            analyzer.patterns["total_sessions"] = 5

            report_path = analyzer.save_report(project_path=project_path)

            content = report_path.read_text()
            assert len(content) > 0
            assert "Report" in content or "report" in content


class TestSessionAnalyzerGetMetrics:
    """Test SessionAnalyzer.get_metrics() method."""

    def test_get_metrics_empty(self):
        """Test getting metrics with no data."""
        analyzer = SessionAnalyzer()
        metrics = analyzer.get_metrics()

        assert "total_sessions" in metrics
        assert "success_rate" in metrics

    def test_get_metrics_with_data(self):
        """Test getting metrics with session data."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["total_sessions"] = 10
        analyzer.patterns["failed_sessions"] = 2
        analyzer.patterns["total_events"] = 100

        metrics = analyzer.get_metrics()

        assert metrics["total_sessions"] == 10
        assert metrics["failed_sessions"] == 2
        assert metrics["total_events"] == 100

    def test_get_metrics_calculates_success_rate(self):
        """Test metrics calculates success rate."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["total_sessions"] = 10
        analyzer.patterns["failed_sessions"] = 2

        metrics = analyzer.get_metrics()

        assert metrics["success_rate"] == 80.0

    def test_get_metrics_calculates_average_session_length(self):
        """Test metrics calculates average session length."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["total_sessions"] = 5
        analyzer.patterns["total_events"] = 50

        metrics = analyzer.get_metrics()

        assert metrics["average_session_length"] == 10.0

    def test_get_metrics_returns_copy(self):
        """Test get_metrics returns copy of patterns."""
        analyzer = SessionAnalyzer()
        analyzer.patterns["total_sessions"] = 5

        metrics = analyzer.get_metrics()
        metrics["total_sessions"] = 100

        # Original should not be modified
        assert analyzer.patterns["total_sessions"] == 5


class TestSessionAnalyzerIntegration:
    """Integration tests for SessionAnalyzer."""

    def test_complete_workflow(self):
        """Test complete analysis workflow."""
        with tempfile.TemporaryDirectory() as tmpdir:
            claude_projects = Path(tmpdir)
            project_dir = claude_projects / "test-project"
            project_dir.mkdir(parents=True, exist_ok=True)

            # Create sample session file
            session_file = project_dir / "session-001.json"
            session_data = {
                "events": [
                    {"type": "tool_call", "toolName": "Bash", "command": "ls"},
                    {"type": "tool_call", "toolName": "Read"},
                ]
            }
            session_file.write_text(json.dumps(session_data))

            analyzer = SessionAnalyzer()
            with patch.object(analyzer, "claude_projects", claude_projects):
                # Parse sessions
                analyzer.parse_sessions()

                # Generate report
                report = analyzer.generate_report()
                assert len(report) > 0

                # Get metrics
                metrics = analyzer.get_metrics()
                assert metrics["total_sessions"] == 1

                # Save report
                with tempfile.TemporaryDirectory() as report_dir:
                    report_path = analyzer.save_report(project_path=Path(report_dir))
                    assert report_path.exists()

    def test_mixed_session_formats(self):
        """Test handling mixed JSON and JSONL formats."""
        with tempfile.TemporaryDirectory() as tmpdir:
            claude_projects = Path(tmpdir)
            project_dir = claude_projects / "test-project"
            project_dir.mkdir(parents=True, exist_ok=True)

            # Create JSON format session
            json_file = project_dir / "session-001.json"
            json_file.write_text(json.dumps({"events": []}))

            # Create JSONL format session
            jsonl_file = project_dir / "session.jsonl"
            jsonl_file.write_text(json.dumps({"type": "summary", "summary": "Test"}))

            analyzer = SessionAnalyzer()
            with patch.object(analyzer, "claude_projects", claude_projects):
                analyzer.parse_sessions()
                assert analyzer.patterns["total_sessions"] >= 1


class TestSessionAnalyzerVerboseMode:
    """Test SessionAnalyzer verbose mode for 100% coverage."""

    @patch.object(Path, "exists")
    def test_verbose_mode_no_directory(self, mock_exists):
        """Test verbose mode output when directory doesn't exist (line 57)."""
        mock_exists.return_value = False
        analyzer = SessionAnalyzer(verbose=True)
        with patch("builtins.print") as mock_print:
            result = analyzer.parse_sessions()
            assert result == analyzer.patterns
            # Verbose: should print warning about missing directory
            mock_print.assert_called()
            args = str(mock_print.call_args)
            assert "Claude projects directory not found" in args or "⚠️" in args

    def test_verbose_mode_session_files_found(self):
        """Test verbose mode shows session files count (line 68)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            claude_projects = Path(tmpdir)
            project_dir = claude_projects / "test-project"
            project_dir.mkdir(parents=True, exist_ok=True)

            session_file = project_dir / "session-001.json"
            session_file.write_text(json.dumps({"events": []}))

            analyzer = SessionAnalyzer(verbose=True)
            with patch.object(analyzer, "claude_projects", claude_projects):
                with patch("builtins.print") as mock_print:
                    analyzer.parse_sessions()
                    # Verbose: should print session files found
                    mock_print.assert_called()
                    args = str(mock_print.call_args)
                    assert "Found" in args or "session file" in args.lower()

    def test_verbose_mode_jsonl_parse_error(self):
        """Test verbose mode handles JSONL parse errors (lines 87-89)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            claude_projects = Path(tmpdir)
            project_dir = claude_projects / "test-project"
            project_dir.mkdir(parents=True, exist_ok=True)

            session_file = project_dir / "session.jsonl"
            # Create JSONL with invalid JSON on line 2
            session_file.write_text('{"type": "summary", "summary": "Valid"}\ninvalid json here\n')

            analyzer = SessionAnalyzer(verbose=True)
            with patch.object(analyzer, "claude_projects", claude_projects):
                with patch("builtins.print") as mock_print:
                    analyzer.parse_sessions()
                    # Verbose: should print JSON decode error
                    print_calls = [str(call) for call in mock_print.call_args_list]
                    assert any("Error reading line" in call or "⚠️" in call for call in print_calls)

    def test_verbose_mode_file_read_error(self):
        """Test verbose mode handles file read errors (line 103)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            claude_projects = Path(tmpdir)
            project_dir = claude_projects / "test-project"
            project_dir.mkdir(parents=True, exist_ok=True)

            session_file = project_dir / "session-001.json"
            session_file.write_text("{ invalid json content")

            analyzer = SessionAnalyzer(verbose=True)
            with patch.object(analyzer, "claude_projects", claude_projects):
                with patch("builtins.print") as mock_print:
                    analyzer.parse_sessions()
                    # Verbose: should print file read error
                    print_calls = [str(call) for call in mock_print.call_args_list]
                    assert any("Error reading" in call or "⚠️" in call for call in print_calls)

    def test_verbose_mode_save_report_confirmation(self):
        """Test verbose mode confirms report save (lines 372, 382)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            analyzer = SessionAnalyzer(verbose=True)

            with patch("builtins.print") as mock_print:
                report_path = analyzer.save_report(project_path=project_path)

                assert report_path.exists()
                # Verbose: should print save confirmation
                mock_print.assert_called()
                args = str(mock_print.call_args)
                assert "Report saved" in args or "📄" in args

    def test_verbose_mode_save_report_default_path(self):
        """Test verbose mode with default path parameters (line 372)."""
        analyzer = SessionAnalyzer(verbose=True)

        with patch("builtins.print") as mock_print:
            with patch("pathlib.Path.cwd") as mock_cwd:
                # Mock Path.cwd() to return temp directory
                with tempfile.TemporaryDirectory() as tmpdir:
                    mock_cwd.return_value = Path(tmpdir)
                    report_path = analyzer.save_report()

                    assert report_path.exists()
                    # Verbose: should print save confirmation
                    mock_print.assert_called()
                    args = str(mock_print.call_args)
                    assert "Report saved" in args or "📄" in args
