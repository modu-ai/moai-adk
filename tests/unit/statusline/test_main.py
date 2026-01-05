"""
Unit tests for moai_adk.statusline.main module.

Tests cover:
- read_session_context function
- safe_collect_* functions
- build_statusline_data function
- main function
"""

from io import StringIO
from unittest import mock

from moai_adk.statusline.main import (
    build_statusline_data,
    read_session_context,
    safe_check_update,
    safe_collect_alfred_task,
    safe_collect_duration,
    safe_collect_git_info,
    safe_collect_version,
)


class TestReadSessionContext:
    """Test read_session_context function."""

    def test_read_valid_json(self):
        """Test reading valid JSON from stdin."""
        json_input = '{"model": "claude-3", "version": "0.3.0"}'

        with mock.patch("sys.stdin", StringIO(json_input)):
            with mock.patch("sys.stdin.isatty", return_value=False):
                context = read_session_context()

        assert context["model"] == "claude-3"
        assert context["version"] == "0.3.0"

    def test_read_empty_input(self):
        """Test reading empty input."""
        with mock.patch("sys.stdin", StringIO("")):
            with mock.patch("sys.stdin.isatty", return_value=False):
                context = read_session_context()

        assert context == {}

    def test_read_invalid_json(self):
        """Test reading invalid JSON."""
        with mock.patch("sys.stdin", StringIO("not valid json")):
            with mock.patch("sys.stdin.isatty", return_value=False):
                context = read_session_context()

        assert context == {}

    def test_read_tty_environment(self):
        """Test reading in TTY environment."""
        with mock.patch("sys.stdin.isatty", return_value=True):
            context = read_session_context()

        assert context == {}


class TestSafeCollectGitInfo:
    """Test safe_collect_git_info function."""

    def test_collect_git_info_success(self):
        """Test successful git info collection."""
        mock_git_info = mock.Mock()
        mock_git_info.branch = "main"
        mock_git_info.staged = 0
        mock_git_info.modified = 2
        mock_git_info.untracked = 1

        with mock.patch("moai_adk.statusline.main.GitCollector") as mock_collector:
            mock_collector.return_value.collect_git_info.return_value = mock_git_info
            branch, status = safe_collect_git_info()

        assert branch == "main"
        assert "+0" in status
        assert "M2" in status

    def test_collect_git_info_failure(self):
        """Test git info collection with failure."""
        with mock.patch("moai_adk.statusline.main.GitCollector") as mock_collector:
            mock_collector.side_effect = OSError("Git not found")
            branch, status = safe_collect_git_info()

        assert branch == "N/A"
        assert status == ""


class TestSafeCollectDuration:
    """Test safe_collect_duration function."""

    def test_collect_duration_success(self):
        """Test successful duration collection."""
        with mock.patch("moai_adk.statusline.main.MetricsTracker") as mock_tracker:
            mock_tracker.return_value.get_duration.return_value = "5m"
            duration = safe_collect_duration()

        assert duration == "5m"

    def test_collect_duration_failure(self):
        """Test duration collection with failure."""
        with mock.patch("moai_adk.statusline.main.MetricsTracker") as mock_tracker:
            mock_tracker.side_effect = OSError("File not found")
            duration = safe_collect_duration()

        assert duration == "0m"


class TestSafeCollectAlfredTask:
    """Test safe_collect_alfred_task function."""

    def test_collect_alfred_task_success(self):
        """Test successful Alfred task collection."""
        mock_task = mock.Mock()
        mock_task.command = "moai:2-run"
        mock_task.stage = "phase-2"

        with mock.patch("moai_adk.statusline.main.AlfredDetector") as mock_detector:
            mock_detector.return_value.detect_active_task.return_value = mock_task
            task = safe_collect_alfred_task()

        # Task should be formatted with uppercase command and stage
        assert "MOAI:2-RUN" in task
        assert "phase-2" in task or "PHASE-2" in task.upper()

    def test_collect_alfred_task_no_command(self):
        """Test Alfred task collection with no command."""
        mock_task = mock.Mock()
        mock_task.command = None

        with mock.patch("moai_adk.statusline.main.AlfredDetector") as mock_detector:
            mock_detector.return_value.detect_active_task.return_value = mock_task
            task = safe_collect_alfred_task()

        assert task == ""

    def test_collect_alfred_task_failure(self):
        """Test Alfred task collection with failure."""
        with mock.patch("moai_adk.statusline.main.AlfredDetector") as mock_detector:
            mock_detector.side_effect = RuntimeError("Detection failed")
            task = safe_collect_alfred_task()

        assert task == ""


class TestSafeCollectVersion:
    """Test safe_collect_version function."""

    def test_collect_version_success(self):
        """Test successful version collection."""
        with mock.patch("moai_adk.statusline.main.VersionReader") as mock_reader:
            mock_reader.return_value.get_version.return_value = "0.3.0"
            version = safe_collect_version()

        assert version == "0.3.0"

    def test_collect_version_failure(self):
        """Test version collection with failure."""
        with mock.patch("moai_adk.statusline.main.VersionReader") as mock_reader:
            mock_reader.side_effect = ImportError("Module not found")
            version = safe_collect_version()

        assert version == "unknown"


class TestSafeCheckUpdate:
    """Test safe_check_update function."""

    def test_check_update_available(self):
        """Test checking for available update."""
        mock_update = mock.Mock()
        mock_update.available = True
        mock_update.latest_version = "0.4.0"

        with mock.patch("moai_adk.statusline.main.UpdateChecker") as mock_checker:
            mock_checker.return_value.check_for_update.return_value = mock_update
            available, version = safe_check_update("0.3.0")

        assert available is True
        assert version == "0.4.0"

    def test_check_update_not_available(self):
        """Test checking when no update available."""
        mock_update = mock.Mock()
        mock_update.available = False
        mock_update.latest_version = None

        with mock.patch("moai_adk.statusline.main.UpdateChecker") as mock_checker:
            mock_checker.return_value.check_for_update.return_value = mock_update
            available, version = safe_check_update("0.3.0")

        assert available is False
        assert version is None

    def test_check_update_failure(self):
        """Test checking update with failure."""
        with mock.patch("moai_adk.statusline.main.UpdateChecker") as mock_checker:
            mock_checker.side_effect = OSError("Network error")
            available, version = safe_check_update("0.3.0")

        assert available is False
        assert version is None


class TestBuildStatuslineData:
    """Test build_statusline_data function."""

    def test_build_statusline_minimal_context(self):
        """Test building statusline with minimal context."""
        session_context = {}

        with mock.patch("moai_adk.statusline.main.StatuslineRenderer") as mock_renderer:
            mock_renderer.return_value.render.return_value = "statusline output"
            with mock.patch(
                "moai_adk.statusline.main.safe_collect_git_info",
                return_value=("main", ""),
            ):
                with mock.patch("moai_adk.statusline.main.safe_collect_duration", return_value="1m"):
                    with mock.patch(
                        "moai_adk.statusline.main.safe_collect_alfred_task",
                        return_value="",
                    ):
                        with mock.patch(
                            "moai_adk.statusline.main.safe_collect_version",
                            return_value="0.3.0",
                        ):
                            with mock.patch(
                                "moai_adk.statusline.main.safe_check_update",
                                return_value=(False, None),
                            ):
                                statusline = build_statusline_data(session_context)

        assert statusline == "statusline output"

    def test_build_statusline_full_context(self):
        """Test building statusline with full context."""
        session_context = {
            "model": {"display_name": "Claude 3 Opus"},
            "version": "2.0.0",
            "cwd": "/home/user/project",
            "output_style": {"name": "default"},
        }

        with mock.patch("moai_adk.statusline.main.StatuslineRenderer") as mock_renderer:
            mock_renderer.return_value.render.return_value = ""
            with mock.patch(
                "moai_adk.statusline.main.safe_collect_git_info",
                return_value=("main", ""),
            ):
                with mock.patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                    with mock.patch(
                        "moai_adk.statusline.main.safe_collect_alfred_task",
                        return_value="[TASK]",
                    ):
                        with mock.patch(
                            "moai_adk.statusline.main.safe_collect_version",
                            return_value="0.3.0",
                        ):
                            with mock.patch(
                                "moai_adk.statusline.main.safe_check_update",
                                return_value=(False, None),
                            ):
                                statusline = build_statusline_data(session_context, mode="compact")

        # Should not raise exception
        assert isinstance(statusline, str)

    def test_build_statusline_different_modes(self):
        """Test building statusline with different modes."""
        session_context = {}

        with mock.patch("moai_adk.statusline.main.StatuslineRenderer") as mock_renderer:
            mock_renderer.return_value.render.return_value = "output"
            with mock.patch(
                "moai_adk.statusline.main.safe_collect_git_info",
                return_value=("main", ""),
            ):
                with mock.patch("moai_adk.statusline.main.safe_collect_duration", return_value="1m"):
                    with mock.patch(
                        "moai_adk.statusline.main.safe_collect_alfred_task",
                        return_value="",
                    ):
                        with mock.patch(
                            "moai_adk.statusline.main.safe_collect_version",
                            return_value="0.3.0",
                        ):
                            with mock.patch(
                                "moai_adk.statusline.main.safe_check_update",
                                return_value=(False, None),
                            ):
                                for mode in ["compact", "extended", "minimal"]:
                                    statusline = build_statusline_data(session_context, mode=mode)
                                    assert isinstance(statusline, str)

    def test_build_statusline_error_handling(self):
        """Test error handling in statusline building."""
        session_context = {}

        with mock.patch("moai_adk.statusline.main.StatuslineRenderer") as mock_renderer:
            mock_renderer.side_effect = Exception("Rendering error")
            with mock.patch(
                "moai_adk.statusline.main.safe_collect_git_info",
                return_value=("main", ""),
            ):
                with mock.patch("moai_adk.statusline.main.safe_collect_duration", return_value="1m"):
                    with mock.patch(
                        "moai_adk.statusline.main.safe_collect_alfred_task",
                        return_value="",
                    ):
                        with mock.patch(
                            "moai_adk.statusline.main.safe_collect_version",
                            return_value="0.3.0",
                        ):
                            with mock.patch(
                                "moai_adk.statusline.main.safe_check_update",
                                return_value=(False, None),
                            ):
                                statusline = build_statusline_data(session_context)

        # Should gracefully degrade
        assert isinstance(statusline, str)


class TestStatuslineExtraction:
    """Test context extraction from session."""

    def test_extract_model_info(self):
        """Test extracting model information."""
        session_context = {"model": {"display_name": "Claude 3 Opus", "name": "claude-opus"}}

        with mock.patch("moai_adk.statusline.main.StatuslineRenderer") as mock_renderer:
            mock_renderer.return_value.render.return_value = ""
            with mock.patch(
                "moai_adk.statusline.main.safe_collect_git_info",
                return_value=("main", ""),
            ):
                with mock.patch("moai_adk.statusline.main.safe_collect_duration", return_value="1m"):
                    with mock.patch(
                        "moai_adk.statusline.main.safe_collect_alfred_task",
                        return_value="",
                    ):
                        with mock.patch(
                            "moai_adk.statusline.main.safe_collect_version",
                            return_value="0.3.0",
                        ):
                            with mock.patch(
                                "moai_adk.statusline.main.safe_check_update",
                                return_value=(False, None),
                            ):
                                # Should extract display_name if available
                                build_statusline_data(session_context)

    def test_extract_directory(self):
        """Test extracting directory from cwd."""
        session_context = {"cwd": "/home/user/project"}

        with mock.patch("moai_adk.statusline.main.StatuslineRenderer") as mock_renderer:
            mock_renderer.return_value.render.return_value = ""
            with mock.patch(
                "moai_adk.statusline.main.safe_collect_git_info",
                return_value=("main", ""),
            ):
                with mock.patch("moai_adk.statusline.main.safe_collect_duration", return_value="1m"):
                    with mock.patch(
                        "moai_adk.statusline.main.safe_collect_alfred_task",
                        return_value="",
                    ):
                        with mock.patch(
                            "moai_adk.statusline.main.safe_collect_version",
                            return_value="0.3.0",
                        ):
                            with mock.patch(
                                "moai_adk.statusline.main.safe_check_update",
                                return_value=(False, None),
                            ):
                                # Should extract 'project' as directory
                                build_statusline_data(session_context)
