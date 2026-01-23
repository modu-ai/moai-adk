"""Tests for statusline.main module."""

import json
import os
from unittest.mock import MagicMock, patch

from moai_adk.statusline.main import (
    build_statusline_data,
    extract_context_window,
    format_token_count,
    main,
    read_session_context,
    safe_check_update,
    safe_collect_alfred_task,
    safe_collect_duration,
    safe_collect_git_info,
    safe_collect_memory,
    safe_collect_version,
)


class TestReadSessionContext:
    """Test read_session_context function."""

    def test_read_empty_stdin(self):
        """Test reading empty stdin."""
        with patch("sys.stdin.isatty", return_value=True):
            result = read_session_context()
            assert result == {}

    def test_read_valid_json(self):
        """Test reading valid JSON from stdin."""
        test_data = {"model": "claude-opus-4", "cwd": "/path/to/project"}
        with patch("sys.stdin.isatty", return_value=False):
            with patch("sys.stdin.read", return_value=json.dumps(test_data)):
                result = read_session_context()
                assert result == test_data

    def test_read_invalid_json(self):
        """Test reading invalid JSON from stdin."""
        with patch("sys.stdin.isatty", return_value=False):
            with patch("sys.stdin.read", return_value="invalid json {"):
                result = read_session_context()
                assert result == {}

    def test_read_eof_error(self):
        """Test EOF error when reading stdin."""
        with patch("sys.stdin.isatty", return_value=False):
            with patch("sys.stdin.read", side_effect=EOFError()):
                result = read_session_context()
                assert result == {}


class TestFormatTokenCount:
    """Test format_token_count function."""

    def test_format_tokens_below_1000(self):
        """Test formatting tokens below 1000."""
        assert format_token_count(0) == "0"
        assert format_token_count(999) == "999"
        assert format_token_count(500) == "500"

    def test_format_tokens_above_1000(self):
        """Test formatting tokens above 1000."""
        assert format_token_count(1000) == "1K"
        assert format_token_count(15234) == "15K"
        assert format_token_count(999999) == "999K"


class TestExtractContextWindow:
    """Test extract_context_window function.

    The function now returns a dictionary with:
    - used_percentage: Percentage of context window used (0.0-100.0)
    - remaining_percentage: Percentage of context window remaining (0.0-100.0)

    Following Claude Code documentation "Advanced Approach":
    - If current_usage is available, calculate from tokens
    - If current_usage is null/empty, return 0% (session start state)
    """

    def test_empty_context(self):
        """Test extracting from empty context returns 0% used."""
        result = extract_context_window({})
        assert result == {"used_percentage": 0.0, "remaining_percentage": 100.0, "tokens_used": 0, "tokens_max": 200000}

    def test_no_context_window_info(self):
        """Test when context_window info is missing returns 0% used."""
        session_context = {"model": "claude-opus"}
        result = extract_context_window(session_context)
        assert result == {"used_percentage": 0.0, "remaining_percentage": 100.0, "tokens_used": 0, "tokens_max": 200000}

    def test_no_context_window_size(self):
        """Test when context_window_size is missing uses default 200K."""
        session_context = {"context_window": {"other_key": "value"}}
        result = extract_context_window(session_context)
        # No current_usage means 0% used (session start state)
        assert result == {"used_percentage": 0.0, "remaining_percentage": 100.0, "tokens_used": 0, "tokens_max": 200000}

    def test_no_current_usage(self):
        """Test when current_usage is missing returns 0% (session start)."""
        session_context = {"context_window": {"context_window_size": 200000, "total_input_tokens": 15000}}
        result = extract_context_window(session_context)
        # Per Claude Code docs: If current_usage is null, return 0%
        assert result == {"used_percentage": 0.0, "remaining_percentage": 100.0, "tokens_used": 0, "tokens_max": 200000}

    def test_with_current_usage(self):
        """Test with complete current_usage data calculates percentage correctly."""
        session_context = {
            "context_window": {
                "context_window_size": 200000,
                "current_usage": {
                    "input_tokens": 10000,
                    "cache_creation_input_tokens": 3000,
                    "cache_read_input_tokens": 2000,
                },
            }
        }
        result = extract_context_window(session_context)
        # 15000 / 200000 = 7.5%
        assert result["used_percentage"] == 7.5
        assert result["remaining_percentage"] == 92.5
        assert result["tokens_used"] == 15000
        assert result["tokens_max"] == 200000

    def test_zero_current_tokens(self):
        """Test when current_usage is empty dict returns 0% used."""
        session_context = {"context_window": {"context_window_size": 200000, "current_usage": {}}}
        result = extract_context_window(session_context)
        # Empty dict means 0 tokens, so 0% used
        assert result == {"used_percentage": 0.0, "remaining_percentage": 100.0, "tokens_used": 0, "tokens_max": 200000}

    def test_with_context_window_alt_key(self):
        """Test with context_window_info alternative key."""
        session_context = {
            "context_window_info": {
                "context_window_size": 100000,
                "current_usage": {
                    "input_tokens": 50000,
                },
            }
        }
        result = extract_context_window(session_context)
        # 50000 / 100000 = 50%
        assert result["used_percentage"] == 50.0
        assert result["remaining_percentage"] == 50.0
        assert result["tokens_used"] == 50000
        assert result["tokens_max"] == 100000


class TestSafeCollectGitInfo:
    """Test safe_collect_git_info function."""

    def test_collect_git_info_success(self):
        """Test successful git info collection."""
        mock_git_info = MagicMock()
        mock_git_info.branch = "main"
        mock_git_info.staged = 1
        mock_git_info.modified = 2
        mock_git_info.untracked = 3

        with patch("moai_adk.statusline.main.GitCollector") as mock_collector:
            mock_instance = MagicMock()
            mock_instance.collect_git_info.return_value = mock_git_info
            mock_collector.return_value = mock_instance

            branch, status = safe_collect_git_info()
            assert branch == "main"
            assert status == "+1 M2 ?3"

    def test_collect_git_info_no_branch(self):
        """Test git info collection when branch is None."""
        mock_git_info = MagicMock()
        mock_git_info.branch = None
        mock_git_info.staged = 0
        mock_git_info.modified = 0
        mock_git_info.untracked = 0

        with patch("moai_adk.statusline.main.GitCollector") as mock_collector:
            mock_instance = MagicMock()
            mock_instance.collect_git_info.return_value = mock_git_info
            mock_collector.return_value = mock_instance

            branch, status = safe_collect_git_info()
            assert branch == "unknown"

    def test_collect_git_info_os_error(self):
        """Test git info collection with OSError."""
        with patch("moai_adk.statusline.main.GitCollector", side_effect=OSError("Git error")):
            branch, status = safe_collect_git_info()
            assert branch == "N/A"
            assert status == ""

    def test_collect_git_info_attribute_error(self):
        """Test git info collection with AttributeError."""
        with patch("moai_adk.statusline.main.GitCollector", side_effect=AttributeError("No attribute")):
            branch, status = safe_collect_git_info()
            assert branch == "N/A"
            assert status == ""


class TestSafeCollectDuration:
    """Test safe_collect_duration function."""

    def test_collect_duration_success(self):
        """Test successful duration collection."""
        with patch("moai_adk.statusline.main.MetricsTracker") as mock_tracker:
            mock_instance = MagicMock()
            mock_instance.get_duration.return_value = "10m"
            mock_tracker.return_value = mock_instance

            result = safe_collect_duration()
            assert result == "10m"

    def test_collect_duration_os_error(self):
        """Test duration collection with OSError."""
        with patch("moai_adk.statusline.main.MetricsTracker", side_effect=OSError("Error")):
            result = safe_collect_duration()
            assert result == "0m"

    def test_collect_duration_value_error(self):
        """Test duration collection with ValueError."""
        with patch("moai_adk.statusline.main.MetricsTracker", side_effect=ValueError("Invalid")):
            result = safe_collect_duration()
            assert result == "0m"


class TestSafeCollectAlfredTask:
    """Test safe_collect_alfred_task function."""

    def test_collect_alfred_task_with_command(self):
        """Test collecting Alfred task with command."""
        mock_task = MagicMock()
        mock_task.command = "plan"
        mock_task.stage = "execution"

        with patch("moai_adk.statusline.main.AlfredDetector") as mock_detector:
            mock_instance = MagicMock()
            mock_instance.detect_active_task.return_value = mock_task
            mock_detector.return_value = mock_instance

            result = safe_collect_alfred_task()
            assert result == "[PLAN-execution]"

    def test_collect_alfred_task_no_command(self):
        """Test collecting Alfred task without command."""
        mock_task = MagicMock()
        mock_task.command = ""
        mock_task.stage = None

        with patch("moai_adk.statusline.main.AlfredDetector") as mock_detector:
            mock_instance = MagicMock()
            mock_instance.detect_active_task.return_value = mock_task
            mock_detector.return_value = mock_instance

            result = safe_collect_alfred_task()
            assert result == ""

    def test_collect_alfred_task_runtime_error(self):
        """Test Alfred task collection with RuntimeError."""
        with patch("moai_adk.statusline.main.AlfredDetector", side_effect=RuntimeError("Error")):
            result = safe_collect_alfred_task()
            assert result == ""


class TestSafeCollectVersion:
    """Test safe_collect_version function."""

    def test_collect_version_success(self):
        """Test successful version collection."""
        with patch("moai_adk.statusline.main.VersionReader") as mock_reader:
            mock_instance = MagicMock()
            mock_instance.get_version.return_value = "1.0.0"
            mock_reader.return_value = mock_instance

            result = safe_collect_version()
            assert result == "1.0.0"

    def test_collect_version_none(self):
        """Test version collection when version is None."""
        with patch("moai_adk.statusline.main.VersionReader") as mock_reader:
            mock_instance = MagicMock()
            mock_instance.get_version.return_value = None
            mock_reader.return_value = mock_instance

            result = safe_collect_version()
            assert result == "unknown"

    def test_collect_version_import_error(self):
        """Test version collection with ImportError."""
        with patch("moai_adk.statusline.main.VersionReader", side_effect=ImportError()):
            result = safe_collect_version()
            assert result == "unknown"


class TestSafeCollectMemory:
    """Test safe_collect_memory function."""

    def test_collect_memory_success(self):
        """Test successful memory collection."""
        with patch("moai_adk.statusline.memory_collector.MemoryCollector") as mock_collector:
            mock_instance = MagicMock()
            mock_instance.get_display_string.return_value = "128MB"
            mock_collector.return_value = mock_instance

            result = safe_collect_memory()
            assert result == "128MB"

    def test_collect_memory_import_error(self):
        """Test memory collection with ImportError."""
        with patch(
            "moai_adk.statusline.memory_collector.MemoryCollector",
            side_effect=ImportError("No module"),
        ):
            result = safe_collect_memory()
            assert result == "N/A"

    def test_collect_memory_os_error(self):
        """Test memory collection with OSError."""
        with patch("moai_adk.statusline.memory_collector.MemoryCollector", side_effect=OSError("Error")):
            result = safe_collect_memory()
            assert result == "N/A"


class TestSafeCheckUpdate:
    """Test safe_check_update function."""

    def test_check_update_available(self):
        """Test update check with available update."""
        mock_update_info = MagicMock()
        mock_update_info.available = True
        mock_update_info.latest_version = "2.0.0"

        with patch("moai_adk.statusline.main.UpdateChecker") as mock_checker:
            mock_instance = MagicMock()
            mock_instance.check_for_update.return_value = mock_update_info
            mock_checker.return_value = mock_instance

            available, latest = safe_check_update("1.0.0")
            assert available is True
            assert latest == "2.0.0"

    def test_check_update_not_available(self):
        """Test update check with no available update."""
        mock_update_info = MagicMock()
        mock_update_info.available = False
        mock_update_info.latest_version = None

        with patch("moai_adk.statusline.main.UpdateChecker") as mock_checker:
            mock_instance = MagicMock()
            mock_instance.check_for_update.return_value = mock_update_info
            mock_checker.return_value = mock_instance

            available, latest = safe_check_update("1.0.0")
            assert available is False
            assert latest is None

    def test_check_update_os_error(self):
        """Test update check with OSError."""
        with patch("moai_adk.statusline.main.UpdateChecker", side_effect=OSError("Error")):
            available, latest = safe_check_update("1.0.0")
            assert available is False
            assert latest is None


class TestBuildStatuslineData:
    """Test build_statusline_data function."""

    def test_build_statusline_basic(self):
        """Test building basic statusline."""
        session_context = {
            "model": {"display_name": "Claude Opus 4"},
            "cwd": "/path/to/project",
            "version": "1.0.0",
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "+1 M0")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="1.0.0"):
                        with patch("moai_adk.statusline.main.safe_collect_memory", return_value="128MB"):
                            with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                                with patch("moai_adk.statusline.main.StatuslineRenderer") as mock_renderer:
                                    mock_instance = MagicMock()
                                    mock_instance.render.return_value = "test statusline"
                                    mock_renderer.return_value = mock_instance

                                    result = build_statusline_data(session_context)
                                    assert result == "test statusline"

    def test_build_statusline_with_context_window(self):
        """Test building statusline with context window info."""
        session_context = {
            "model": {"name": "claude-opus"},
            "cwd": "/project",
            "context_window": {
                "context_window_size": 200000,
                "current_usage": {"input_tokens": 15000},
            },
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="0m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="1.0.0"):
                        with patch("moai_adk.statusline.main.safe_collect_memory", return_value="N/A"):
                            with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                                with patch("moai_adk.statusline.main.StatuslineRenderer") as mock_renderer:
                                    mock_instance = MagicMock()
                                    mock_instance.render.return_value = "statusline"
                                    mock_renderer.return_value = mock_instance

                                    result = build_statusline_data(session_context)
                                    assert result == "statusline"

    def test_build_statusline_exception_handling(self):
        """Test exception handling in statusline building."""
        # Invalid session context that might cause errors
        session_context = {"invalid_key": "value"}

        # Even with invalid context, should return empty string instead of crashing
        result = build_statusline_data(session_context)
        # Should handle gracefully - either return statusline or empty string
        assert isinstance(result, str)


class TestMain:
    """Test main function."""

    def test_main_basic_execution(self):
        """Test basic main execution."""
        session_context = {"model": {"name": "claude"}, "cwd": "/project"}

        with patch("moai_adk.statusline.main.read_session_context", return_value=session_context):
            with patch.dict(os.environ, {"MOAI_STATUSLINE_MODE": "compact"}, clear=False):
                with patch("moai_adk.statusline.main.build_statusline_data", return_value="status"):
                    with patch("builtins.print") as mock_print:
                        main()
                        mock_print.assert_called_once_with("status", end="")

    def test_main_debug_mode(self):
        """Test main with debug mode enabled."""
        session_context = {"model": {"name": "claude"}, "cwd": "/project"}

        with patch("moai_adk.statusline.main.read_session_context", return_value=session_context):
            with patch.dict(os.environ, {"MOAI_STATUSLINE_DEBUG": "1"}, clear=False):
                with patch("moai_adk.statusline.main.build_statusline_data", return_value="status"):
                    with patch("builtins.print"):
                        with patch("sys.stderr.write") as mock_stderr:
                            main()
                            # Should write debug info to stderr
                            assert mock_stderr.called

    def test_main_extended_mode(self):
        """Test main with extended mode."""
        session_context = {
            "model": {"name": "claude"},
            "cwd": "/project",
            "statusline": {"mode": "extended"},
        }

        with patch("moai_adk.statusline.main.read_session_context", return_value=session_context):
            with patch("moai_adk.statusline.main.build_statusline_data", return_value="extended"):
                with patch("builtins.print") as mock_print:
                    main()
                    mock_print.assert_called_once_with("extended", end="")

    def test_main_empty_statusline(self):
        """Test main when statusline is empty."""
        with patch("moai_adk.statusline.main.read_session_context", return_value={}):
            with patch("moai_adk.statusline.main.build_statusline_data", return_value=""):
                with patch("builtins.print") as mock_print:
                    main()
                    # Should not print anything if statusline is empty
                    mock_print.assert_not_called()
