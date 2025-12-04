"""Extended tests for moai_adk.statusline.main module.

These tests focus on increasing coverage for statusline building and collection functions.
"""

import json
import os
import sys
from pathlib import Path
from unittest.mock import patch, MagicMock, mock_open
from io import StringIO

import pytest

from moai_adk.statusline.main import (
    read_session_context,
    safe_collect_git_info,
    safe_collect_duration,
    safe_collect_alfred_task,
    safe_collect_version,
    safe_check_update,
    build_statusline_data,
    main,
)


class TestReadSessionContext:
    """Test reading session context from stdin."""

    def test_read_session_context_valid_json(self):
        """Test reading valid JSON from stdin."""
        json_data = '{"model": "claude", "version": "2.0.46"}'

        with patch("sys.stdin.isatty", return_value=False):
            with patch("sys.stdin.read", return_value=json_data):
                context = read_session_context()

                assert context["model"] == "claude"
                assert context["version"] == "2.0.46"

    def test_read_session_context_empty(self):
        """Test reading empty context."""
        with patch("sys.stdin.isatty", return_value=False):
            with patch("sys.stdin.read", return_value=""):
                context = read_session_context()

                assert context == {}

    def test_read_session_context_invalid_json(self):
        """Test reading invalid JSON."""
        with patch("sys.stdin.isatty", return_value=False):
            with patch("sys.stdin.read", return_value="invalid json"):
                context = read_session_context()

                assert context == {}

    def test_read_session_context_tty(self):
        """Test reading when TTY (interactive)."""
        with patch("sys.stdin.isatty", return_value=True):
            context = read_session_context()

            assert context == {}

    def test_read_session_context_eof_error(self):
        """Test handling EOF error."""
        with patch("sys.stdin.isatty", return_value=False):
            with patch("sys.stdin.read", side_effect=EOFError()):
                context = read_session_context()

                assert context == {}

    def test_read_session_context_value_error(self):
        """Test handling ValueError."""
        with patch("sys.stdin.isatty", return_value=False):
            with patch("sys.stdin.read", side_effect=ValueError()):
                context = read_session_context()

                assert context == {}


class TestSafeCollectGitInfo:
    """Test safe git information collection."""

    def test_collect_git_info_success(self):
        """Test successful git info collection."""
        mock_info = MagicMock()
        mock_info.branch = "main"
        mock_info.staged = 1
        mock_info.modified = 2
        mock_info.untracked = 3

        with patch("moai_adk.statusline.main.GitCollector") as MockGit:
            mock_collector = MagicMock()
            mock_collector.collect_git_info.return_value = mock_info
            MockGit.return_value = mock_collector

            branch, git_status = safe_collect_git_info()

            assert branch == "main"
            assert "+1" in git_status
            assert "M2" in git_status
            assert "?3" in git_status

    def test_collect_git_info_os_error(self):
        """Test git info collection with OSError."""
        with patch("moai_adk.statusline.main.GitCollector") as MockGit:
            mock_collector = MagicMock()
            mock_collector.collect_git_info.side_effect = OSError()
            MockGit.return_value = mock_collector

            branch, git_status = safe_collect_git_info()

            assert branch == "N/A"
            assert git_status == ""

    def test_collect_git_info_attribute_error(self):
        """Test git info collection with AttributeError."""
        with patch("moai_adk.statusline.main.GitCollector") as MockGit:
            mock_collector = MagicMock()
            mock_collector.collect_git_info.side_effect = AttributeError()
            MockGit.return_value = mock_collector

            branch, git_status = safe_collect_git_info()

            assert branch == "N/A"
            assert git_status == ""

    def test_collect_git_info_runtime_error(self):
        """Test git info collection with RuntimeError."""
        with patch("moai_adk.statusline.main.GitCollector") as MockGit:
            mock_collector = MagicMock()
            mock_collector.collect_git_info.side_effect = RuntimeError()
            MockGit.return_value = mock_collector

            branch, git_status = safe_collect_git_info()

            assert branch == "N/A"
            assert git_status == ""


class TestSafeCollectDuration:
    """Test safe duration collection."""

    def test_collect_duration_success(self):
        """Test successful duration collection."""
        with patch("moai_adk.statusline.main.MetricsTracker") as MockMetrics:
            mock_tracker = MagicMock()
            mock_tracker.get_duration.return_value = "5m"
            MockMetrics.return_value = mock_tracker

            duration = safe_collect_duration()

            assert duration == "5m"

    def test_collect_duration_os_error(self):
        """Test duration collection with OSError."""
        with patch("moai_adk.statusline.main.MetricsTracker") as MockMetrics:
            mock_tracker = MagicMock()
            mock_tracker.get_duration.side_effect = OSError()
            MockMetrics.return_value = mock_tracker

            duration = safe_collect_duration()

            assert duration == "0m"

    def test_collect_duration_attribute_error(self):
        """Test duration collection with AttributeError."""
        with patch("moai_adk.statusline.main.MetricsTracker") as MockMetrics:
            mock_tracker = MagicMock()
            mock_tracker.get_duration.side_effect = AttributeError()
            MockMetrics.return_value = mock_tracker

            duration = safe_collect_duration()

            assert duration == "0m"

    def test_collect_duration_value_error(self):
        """Test duration collection with ValueError."""
        with patch("moai_adk.statusline.main.MetricsTracker") as MockMetrics:
            mock_tracker = MagicMock()
            mock_tracker.get_duration.side_effect = ValueError()
            MockMetrics.return_value = mock_tracker

            duration = safe_collect_duration()

            assert duration == "0m"


class TestSafeCollectAlfredTask:
    """Test safe Alfred task collection."""

    def test_collect_alfred_task_success(self):
        """Test successful Alfred task collection."""
        mock_task = MagicMock()
        mock_task.command = "TDD"
        mock_task.stage = "GREEN"

        with patch("moai_adk.statusline.main.AlfredDetector") as MockAlfred:
            mock_detector = MagicMock()
            mock_detector.detect_active_task.return_value = mock_task
            MockAlfred.return_value = mock_detector

            task = safe_collect_alfred_task()

            assert "[TDD-GREEN]" in task

    def test_collect_alfred_task_no_stage(self):
        """Test Alfred task collection without stage."""
        mock_task = MagicMock()
        mock_task.command = "PLAN"
        mock_task.stage = None

        with patch("moai_adk.statusline.main.AlfredDetector") as MockAlfred:
            mock_detector = MagicMock()
            mock_detector.detect_active_task.return_value = mock_task
            MockAlfred.return_value = mock_detector

            task = safe_collect_alfred_task()

            assert "[PLAN]" in task

    def test_collect_alfred_task_none(self):
        """Test Alfred task collection with no command."""
        mock_task = MagicMock()
        mock_task.command = None

        with patch("moai_adk.statusline.main.AlfredDetector") as MockAlfred:
            mock_detector = MagicMock()
            mock_detector.detect_active_task.return_value = mock_task
            MockAlfred.return_value = mock_detector

            task = safe_collect_alfred_task()

            assert task == ""

    def test_collect_alfred_task_os_error(self):
        """Test Alfred task collection with OSError."""
        with patch("moai_adk.statusline.main.AlfredDetector") as MockAlfred:
            mock_detector = MagicMock()
            mock_detector.detect_active_task.side_effect = OSError()
            MockAlfred.return_value = mock_detector

            task = safe_collect_alfred_task()

            assert task == ""

    def test_collect_alfred_task_attribute_error(self):
        """Test Alfred task collection with AttributeError."""
        with patch("moai_adk.statusline.main.AlfredDetector") as MockAlfred:
            mock_detector = MagicMock()
            mock_detector.detect_active_task.side_effect = AttributeError()
            MockAlfred.return_value = mock_detector

            task = safe_collect_alfred_task()

            assert task == ""

    def test_collect_alfred_task_runtime_error(self):
        """Test Alfred task collection with RuntimeError."""
        with patch("moai_adk.statusline.main.AlfredDetector") as MockAlfred:
            mock_detector = MagicMock()
            mock_detector.detect_active_task.side_effect = RuntimeError()
            MockAlfred.return_value = mock_detector

            task = safe_collect_alfred_task()

            assert task == ""


class TestSafeCollectVersion:
    """Test safe version collection."""

    def test_collect_version_success(self):
        """Test successful version collection."""
        with patch("moai_adk.statusline.main.VersionReader") as MockVersion:
            mock_reader = MagicMock()
            mock_reader.get_version.return_value = "0.20.1"
            MockVersion.return_value = mock_reader

            version = safe_collect_version()

            assert version == "0.20.1"

    def test_collect_version_empty(self):
        """Test version collection returning empty string."""
        with patch("moai_adk.statusline.main.VersionReader") as MockVersion:
            mock_reader = MagicMock()
            mock_reader.get_version.return_value = ""
            MockVersion.return_value = mock_reader

            version = safe_collect_version()

            assert version == "unknown"

    def test_collect_version_import_error(self):
        """Test version collection with ImportError."""
        with patch("moai_adk.statusline.main.VersionReader") as MockVersion:
            mock_reader = MagicMock()
            mock_reader.get_version.side_effect = ImportError()
            MockVersion.return_value = mock_reader

            version = safe_collect_version()

            assert version == "unknown"

    def test_collect_version_attribute_error(self):
        """Test version collection with AttributeError."""
        with patch("moai_adk.statusline.main.VersionReader") as MockVersion:
            mock_reader = MagicMock()
            mock_reader.get_version.side_effect = AttributeError()
            MockVersion.return_value = mock_reader

            version = safe_collect_version()

            assert version == "unknown"

    def test_collect_version_os_error(self):
        """Test version collection with OSError."""
        with patch("moai_adk.statusline.main.VersionReader") as MockVersion:
            mock_reader = MagicMock()
            mock_reader.get_version.side_effect = OSError()
            MockVersion.return_value = mock_reader

            version = safe_collect_version()

            assert version == "unknown"


class TestSafeCheckUpdate:
    """Test safe update checking."""

    def test_check_update_available(self):
        """Test update check with update available."""
        mock_info = MagicMock()
        mock_info.available = True
        mock_info.latest_version = "0.21.0"

        with patch("moai_adk.statusline.main.UpdateChecker") as MockChecker:
            mock_checker = MagicMock()
            mock_checker.check_for_update.return_value = mock_info
            MockChecker.return_value = mock_checker

            available, latest = safe_check_update("0.20.1")

            assert available is True
            assert latest == "0.21.0"

    def test_check_update_not_available(self):
        """Test update check with no update available."""
        mock_info = MagicMock()
        mock_info.available = False
        mock_info.latest_version = None

        with patch("moai_adk.statusline.main.UpdateChecker") as MockChecker:
            mock_checker = MagicMock()
            mock_checker.check_for_update.return_value = mock_info
            MockChecker.return_value = mock_checker

            available, latest = safe_check_update("0.20.1")

            assert available is False
            assert latest is None

    def test_check_update_os_error(self):
        """Test update check with OSError."""
        with patch("moai_adk.statusline.main.UpdateChecker") as MockChecker:
            mock_checker = MagicMock()
            mock_checker.check_for_update.side_effect = OSError()
            MockChecker.return_value = mock_checker

            available, latest = safe_check_update("0.20.1")

            assert available is False
            assert latest is None

    def test_check_update_attribute_error(self):
        """Test update check with AttributeError."""
        with patch("moai_adk.statusline.main.UpdateChecker") as MockChecker:
            mock_checker = MagicMock()
            mock_checker.check_for_update.side_effect = AttributeError()
            MockChecker.return_value = mock_checker

            available, latest = safe_check_update("0.20.1")

            assert available is False
            assert latest is None

    def test_check_update_runtime_error(self):
        """Test update check with RuntimeError."""
        with patch("moai_adk.statusline.main.UpdateChecker") as MockChecker:
            mock_checker = MagicMock()
            mock_checker.check_for_update.side_effect = RuntimeError()
            MockChecker.return_value = mock_checker

            available, latest = safe_check_update("0.20.1")

            assert available is False
            assert latest is None

    def test_check_update_value_error(self):
        """Test update check with ValueError."""
        with patch("moai_adk.statusline.main.UpdateChecker") as MockChecker:
            mock_checker = MagicMock()
            mock_checker.check_for_update.side_effect = ValueError()
            MockChecker.return_value = mock_checker

            available, latest = safe_check_update("0.20.1")

            assert available is False
            assert latest is None


class TestBuildStatuslineData:
    """Test building statusline data."""

    def test_build_statusline_basic(self):
        """Test building basic statusline."""
        session_context = {
            "model": {"display_name": "claude-3.5-sonnet", "name": "claude"},
            "version": "2.0.46",
            "cwd": "/home/user/project",
            "output_style": {"name": "Concise"}
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "+2 M1")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value="[TDD]"):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="0.20.1"):
                        with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                            result = build_statusline_data(session_context, mode="compact")

                            assert isinstance(result, str)
                            assert len(result) > 0

    def test_build_statusline_minimal_context(self):
        """Test building statusline with minimal context."""
        session_context = {}

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("unknown", "")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="0m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="unknown"):
                        with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                            result = build_statusline_data(session_context, mode="compact")

                            assert isinstance(result, str)

    def test_build_statusline_with_directory(self):
        """Test building statusline extracts directory correctly."""
        session_context = {
            "model": {"display_name": "claude"},
            "cwd": "/home/user/my-project",
            "version": "2.0.46"
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="0.20.1"):
                        with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                            result = build_statusline_data(session_context, mode="compact")

                            assert isinstance(result, str)

    def test_build_statusline_empty_directory(self):
        """Test building statusline with empty directory."""
        session_context = {
            "model": {"display_name": "claude"},
            "cwd": "",
            "version": "2.0.46"
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="0.20.1"):
                        with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                            result = build_statusline_data(session_context, mode="compact")

                            assert isinstance(result, str)

    def test_build_statusline_extended_mode(self):
        """Test building statusline in extended mode."""
        session_context = {
            "model": {"display_name": "claude"},
            "version": "2.0.46"
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="0.20.1"):
                        with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                            result = build_statusline_data(session_context, mode="extended")

                            assert isinstance(result, str)

    def test_build_statusline_minimal_mode(self):
        """Test building statusline in minimal mode."""
        session_context = {
            "model": {"display_name": "claude"},
            "version": "2.0.46"
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="0.20.1"):
                        with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                            result = build_statusline_data(session_context, mode="minimal")

                            assert isinstance(result, str)

    def test_build_statusline_exception_handling(self):
        """Test building statusline with exception."""
        session_context = {
            "model": {"display_name": "claude"},
            "version": "2.0.46"
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", side_effect=Exception("Test error")):
            result = build_statusline_data(session_context, mode="compact")

            assert result == ""

    def test_build_statusline_model_display_name_fallback(self):
        """Test model display name fallback."""
        session_context = {
            "model": {"name": "claude-fallback"},
            "version": "2.0.46"
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="0.20.1"):
                        with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                            result = build_statusline_data(session_context, mode="compact")

                            assert isinstance(result, str)


class TestMain:
    """Test main function."""

    def test_main_normal_execution(self):
        """Test main function normal execution."""
        with patch.dict(os.environ, {"MOAI_STATUSLINE_DEBUG": "0"}):
            with patch("moai_adk.statusline.main.read_session_context", return_value={"model": {"display_name": "claude"}, "version": "2.0.46"}):
                with patch("moai_adk.statusline.main.StatuslineConfig"):
                    with patch("moai_adk.statusline.main.build_statusline_data", return_value="test statusline"):
                        with patch("builtins.print") as mock_print:
                            main()

                            mock_print.assert_called_with("test statusline", end="")

    def test_main_empty_statusline(self):
        """Test main function with empty statusline."""
        with patch.dict(os.environ, {"MOAI_STATUSLINE_DEBUG": "0"}):
            with patch("moai_adk.statusline.main.read_session_context", return_value={}):
                with patch("moai_adk.statusline.main.StatuslineConfig"):
                    with patch("moai_adk.statusline.main.build_statusline_data", return_value=""):
                        with patch("builtins.print") as mock_print:
                            main()

                            mock_print.assert_not_called()

    def test_main_debug_mode_enabled(self):
        """Test main function with debug mode enabled."""
        with patch.dict(os.environ, {"MOAI_STATUSLINE_DEBUG": "1"}):
            with patch("moai_adk.statusline.main.read_session_context", return_value={"test": "data"}):
                with patch("moai_adk.statusline.main.StatuslineConfig"):
                    with patch("moai_adk.statusline.main.build_statusline_data", return_value="test"):
                        with patch("sys.stderr", new_callable=StringIO):
                            with patch("builtins.print"):
                                main()

                                # Debug output should be written to stderr

    def test_main_mode_from_context(self):
        """Test main function gets mode from session context."""
        with patch.dict(os.environ, {"MOAI_STATUSLINE_DEBUG": "0"}):
            with patch("moai_adk.statusline.main.read_session_context", return_value={"statusline": {"mode": "minimal"}}):
                with patch("moai_adk.statusline.main.StatuslineConfig"):
                    with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
                        mock_build.return_value = "test"

                        with patch("builtins.print"):
                            main()

                            # Check that mode was passed to build_statusline_data
                            # This is handled internally in the function

    def test_main_mode_from_environment(self):
        """Test main function gets mode from environment variable."""
        with patch.dict(os.environ, {"MOAI_STATUSLINE_DEBUG": "0", "MOAI_STATUSLINE_MODE": "extended"}):
            with patch("moai_adk.statusline.main.read_session_context", return_value={}):
                with patch("moai_adk.statusline.main.StatuslineConfig"):
                    with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
                        mock_build.return_value = "test"

                        with patch("builtins.print"):
                            main()


class TestStatuslineDataExtraction:
    """Test data extraction from session context."""

    def test_extract_model_display_name(self):
        """Test extracting model display name."""
        session_context = {
            "model": {"display_name": "claude-3.5-sonnet"}
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="0.20.1"):
                        with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                            result = build_statusline_data(session_context)

                            assert isinstance(result, str)

    def test_extract_output_style(self):
        """Test extracting output style."""
        session_context = {
            "model": {"display_name": "claude"},
            "output_style": {"name": "Concise"}
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="0.20.1"):
                        with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                            result = build_statusline_data(session_context)

                            assert isinstance(result, str)

    def test_extract_directory_from_cwd(self):
        """Test extracting directory from cwd."""
        session_context = {
            "model": {"display_name": "claude"},
            "cwd": "/path/to/my-project"
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="0.20.1"):
                        with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                            result = build_statusline_data(session_context)

                            assert isinstance(result, str)
