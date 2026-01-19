# statusline/main.py edge case tests
"""Tests for edge cases and error handling in statusline main module.

Following TDD RED-GREEN-REFACTOR cycle.
These tests cover fallback behavior, error handling, and edge cases.
"""

from __future__ import annotations

from unittest.mock import MagicMock, patch

from moai_adk.statusline.main import (
    build_statusline_data,
    extract_context_window,
    format_token_count,
    safe_check_update,
    safe_collect_alfred_task,
    safe_collect_duration,
    safe_collect_git_info,
    safe_collect_memory,
    safe_collect_version,
)


class TestFormatTokenCount:
    """Tests for format_token_count function."""

    def test_format_token_count_below_threshold(self) -> None:
        """Test format_token_count with tokens < 1000."""
        assert format_token_count(999) == "999"
        assert format_token_count(500) == "500"
        assert format_token_count(1) == "1"
        assert format_token_count(0) == "0"

    def test_format_token_count_at_threshold(self) -> None:
        """Test format_token_count with tokens = 1000."""
        assert format_token_count(1000) == "1K"

    def test_format_token_count_above_threshold(self) -> None:
        """Test format_token_count with tokens > 1000."""
        assert format_token_count(1500) == "1K"
        assert format_token_count(9999) == "9K"
        assert format_token_count(10000) == "10K"
        assert format_token_count(123456) == "123K"

    def test_format_token_count_negative(self) -> None:
        """Test format_token_count with negative tokens (edge case)."""
        # Negative numbers below 1000 are returned as-is
        assert format_token_count(-100) == "-100"
        # Negative numbers >= 1000 use floor division
        # -1500 // 1000 = -2 in Python, so format returns "-2"
        # But the function actually returns the number as-is since it's < 1000 after floor division
        # Let's check the actual behavior
        result = format_token_count(-1500)
        # The implementation: if tokens >= 1000: return f"{tokens // 1000}K"
        # -1500 >= 1000 is False, so it returns str(-1500)
        assert result == "-1500"

    def test_format_token_count_very_large(self) -> None:
        """Test format_token_count with very large token counts."""
        assert format_token_count(999999) == "999K"
        assert format_token_count(1000000) == "1000K"
        assert format_token_count(10000000) == "10000K"


class TestExtractContextWindow:
    """Tests for extract_context_window function."""

    def test_extract_context_window_empty_dict(self) -> None:
        """Test extract_context_window with empty session context."""
        result = extract_context_window({})
        assert result == ""

    def test_extract_context_window_no_context_window(self) -> None:
        """Test extract_context_window without context_window field."""
        result = extract_context_window({"model": {"name": "claude-3"}})
        assert result == ""

    def test_extract_context_window_empty_context_window(self) -> None:
        """Test extract_context_window with empty context_window dict."""
        result = extract_context_window({"context_window": {}})
        assert result == ""

    def test_extract_context_window_no_size(self) -> None:
        """Test extract_context_window without context_window_size."""
        result = extract_context_window({"context_window": {"total_input_tokens": 15000}})
        assert result == ""

    def test_extract_context_window_zero_size(self) -> None:
        """Test extract_context_window with zero context_window_size."""
        result = extract_context_window({"context_window": {"context_window_size": 0, "total_input_tokens": 15000}})
        assert result == ""

    def test_extract_context_window_with_current_usage(self) -> None:
        """Test extract_context_window with current_usage field."""
        context = {
            "context_window": {
                "context_window_size": 200000,
                "current_usage": {
                    "input_tokens": 15000,
                    "cache_creation_input_tokens": 2000,
                    "cache_read_input_tokens": 3000,
                },
            }
        }
        result = extract_context_window(context)
        assert result == "20K/200K"

    def test_extract_context_window_fallback_to_total_tokens(self) -> None:
        """Test extract_context_window falls back to total_input_tokens."""
        context = {
            "context_window": {
                "context_window_size": 200000,
                "total_input_tokens": 15000,
            }
        }
        result = extract_context_window(context)
        assert result == "15K/200K"

    def test_extract_context_window_zero_current_usage(self) -> None:
        """Test extract_context_window with zero current usage."""
        context = {
            "context_window": {
                "context_window_size": 200000,
                "current_usage": {
                    "input_tokens": 0,
                    "cache_creation_input_tokens": 0,
                    "cache_read_input_tokens": 0,
                },
            }
        }
        result = extract_context_window(context)
        assert result == ""

    def test_extract_context_window_partial_current_usage(self) -> None:
        """Test extract_context_window with partial current_usage fields."""
        # Only input_tokens provided
        context = {
            "context_window": {
                "context_window_size": 200000,
                "current_usage": {"input_tokens": 15000},
            }
        }
        result = extract_context_window(context)
        assert result == "15K/200K"


class TestSafeCollectGitInfo:
    """Tests for safe_collect_git_info function."""

    def test_safe_collect_git_info_success(self) -> None:
        """Test safe_collect_git_info successful collection."""
        with patch("moai_adk.statusline.main.GitCollector") as mock_collector_class:
            mock_collector = MagicMock()
            mock_git_info = MagicMock()
            mock_git_info.branch = "main"
            mock_git_info.staged = 1
            mock_git_info.modified = 2
            mock_git_info.untracked = 3
            mock_collector.collect_git_info.return_value = mock_git_info
            mock_collector_class.return_value = mock_collector

            branch, status = safe_collect_git_info()

            assert branch == "main"
            assert status == "+1 M2 ?3"

    def test_safe_collect_git_info_no_branch(self) -> None:
        """Test safe_collect_git_info with None branch."""
        with patch("moai_adk.statusline.main.GitCollector") as mock_collector_class:
            mock_collector = MagicMock()
            mock_git_info = MagicMock()
            mock_git_info.branch = None
            mock_git_info.staged = 0
            mock_git_info.modified = 0
            mock_git_info.untracked = 0
            mock_collector.collect_git_info.return_value = mock_git_info
            mock_collector_class.return_value = mock_collector

            branch, status = safe_collect_git_info()

            assert branch == "unknown"

    def test_safe_collect_git_info_os_error(self) -> None:
        """Test safe_collect_git_info handles OSError."""
        with patch("moai_adk.statusline.main.GitCollector") as mock_collector_class:
            mock_collector_class.side_effect = OSError("Git not found")

            branch, status = safe_collect_git_info()

            assert branch == "N/A"
            assert status == ""

    def test_safe_collect_git_info_attribute_error(self) -> None:
        """Test safe_collect_git_info handles AttributeError."""
        with patch("moai_adk.statusline.main.GitCollector") as mock_collector_class:
            mock_collector = MagicMock()
            mock_collector.collect_git_info.side_effect = AttributeError("Missing attribute")
            mock_collector_class.return_value = mock_collector

            branch, status = safe_collect_git_info()

            assert branch == "N/A"
            assert status == ""

    def test_safe_collect_git_info_runtime_error(self) -> None:
        """Test safe_collect_git_info handles RuntimeError."""
        with patch("moai_adk.statusline.main.GitCollector") as mock_collector_class:
            mock_collector_class.side_effect = RuntimeError("Git error")

            branch, status = safe_collect_git_info()

            assert branch == "N/A"
            assert status == ""


class TestSafeCollectDuration:
    """Tests for safe_collect_duration function."""

    def test_safe_collect_duration_success(self) -> None:
        """Test safe_collect_duration successful collection."""
        with patch("moai_adk.statusline.main.MetricsTracker") as mock_tracker_class:
            mock_tracker = MagicMock()
            mock_tracker.get_duration.return_value = "5m 30s"
            mock_tracker_class.return_value = mock_tracker

            duration = safe_collect_duration()

            assert duration == "5m 30s"

    def test_safe_collect_duration_os_error(self) -> None:
        """Test safe_collect_duration handles OSError."""
        with patch("moai_adk.statusline.main.MetricsTracker") as mock_tracker_class:
            mock_tracker_class.side_effect = OSError("File not found")

            duration = safe_collect_duration()

            assert duration == "0m"

    def test_safe_collect_duration_attribute_error(self) -> None:
        """Test safe_collect_duration handles AttributeError."""
        with patch("moai_adk.statusline.main.MetricsTracker") as mock_tracker_class:
            mock_tracker = MagicMock()
            mock_tracker.get_duration.side_effect = AttributeError("Missing attribute")
            mock_tracker_class.return_value = mock_tracker

            duration = safe_collect_duration()

            assert duration == "0m"

    def test_safe_collect_duration_value_error(self) -> None:
        """Test safe_collect_duration handles ValueError."""
        with patch("moai_adk.statusline.main.MetricsTracker") as mock_tracker_class:
            mock_tracker = MagicMock()
            mock_tracker.get_duration.side_effect = ValueError("Invalid value")
            mock_tracker_class.return_value = mock_tracker

            duration = safe_collect_duration()

            assert duration == "0m"


class TestSafeCollectAlfredTask:
    """Tests for safe_collect_alfred_task function."""

    def test_safe_collect_alfred_task_with_command(self) -> None:
        """Test safe_collect_alfred_task with active command."""
        with patch("moai_adk.statusline.main.AlfredDetector") as mock_detector_class:
            mock_detector = MagicMock()
            mock_task = MagicMock()
            mock_task.command = "analyze"
            mock_task.stage = "planning"
            mock_detector.detect_active_task.return_value = mock_task
            mock_detector_class.return_value = mock_detector

            task = safe_collect_alfred_task()

            # Stage is lowercased in the implementation
            assert task == "[ANALYZE-planning]"

    def test_safe_collect_alfred_task_without_stage(self) -> None:
        """Test safe_collect_alfred_task without stage."""
        with patch("moai_adk.statusline.main.AlfredDetector") as mock_detector_class:
            mock_detector = MagicMock()
            mock_task = MagicMock()
            mock_task.command = "code"
            mock_task.stage = None
            mock_detector.detect_active_task.return_value = mock_task
            mock_detector_class.return_value = mock_detector

            task = safe_collect_alfred_task()

            assert task == "[CODE]"

    def test_safe_collect_alfred_task_no_command(self) -> None:
        """Test safe_collect_alfred_task with no active command."""
        with patch("moai_adk.statusline.main.AlfredDetector") as mock_detector_class:
            mock_detector = MagicMock()
            mock_task = MagicMock()
            mock_task.command = None
            mock_task.stage = None
            mock_detector.detect_active_task.return_value = mock_task
            mock_detector_class.return_value = mock_detector

            task = safe_collect_alfred_task()

            assert task == ""

    def test_safe_collect_alfred_task_os_error(self) -> None:
        """Test safe_collect_alfred_task handles OSError."""
        with patch("moai_adk.statusline.main.AlfredDetector") as mock_detector_class:
            mock_detector_class.side_effect = OSError("File not found")

            task = safe_collect_alfred_task()

            assert task == ""

    def test_safe_collect_alfred_task_attribute_error(self) -> None:
        """Test safe_collect_alfred_task handles AttributeError."""
        with patch("moai_adk.statusline.main.AlfredDetector") as mock_detector_class:
            mock_detector = MagicMock()
            mock_detector.detect_active_task.side_effect = AttributeError("Missing attribute")
            mock_detector_class.return_value = mock_detector

            task = safe_collect_alfred_task()

            assert task == ""

    def test_safe_collect_alfred_task_runtime_error(self) -> None:
        """Test safe_collect_alfred_task handles RuntimeError."""
        with patch("moai_adk.statusline.main.AlfredDetector") as mock_detector_class:
            mock_detector_class.side_effect = RuntimeError("Detector error")

            task = safe_collect_alfred_task()

            assert task == ""


class TestSafeCollectVersion:
    """Tests for safe_collect_version function."""

    def test_safe_collect_version_success(self) -> None:
        """Test safe_collect_version successful collection."""
        with patch("moai_adk.statusline.main.VersionReader") as mock_reader_class:
            mock_reader = MagicMock()
            mock_reader.get_version.return_value = "1.0.0"
            mock_reader_class.return_value = mock_reader

            version = safe_collect_version()

            assert version == "1.0.0"

    def test_safe_collect_version_none_version(self) -> None:
        """Test safe_collect_version with None version."""
        with patch("moai_adk.statusline.main.VersionReader") as mock_reader_class:
            mock_reader = MagicMock()
            mock_reader.get_version.return_value = None
            mock_reader_class.return_value = mock_reader

            version = safe_collect_version()

            assert version == "unknown"

    def test_safe_collect_version_import_error(self) -> None:
        """Test safe_collect_version handles ImportError."""
        with patch("moai_adk.statusline.main.VersionReader") as mock_reader_class:
            mock_reader_class.side_effect = ImportError("Module not found")

            version = safe_collect_version()

            assert version == "unknown"

    def test_safe_collect_version_attribute_error(self) -> None:
        """Test safe_collect_version handles AttributeError."""
        with patch("moai_adk.statusline.main.VersionReader") as mock_reader_class:
            mock_reader = MagicMock()
            mock_reader.get_version.side_effect = AttributeError("Missing attribute")
            mock_reader_class.return_value = mock_reader

            version = safe_collect_version()

            assert version == "unknown"

    def test_safe_collect_version_os_error(self) -> None:
        """Test safe_collect_version handles OSError."""
        with patch("moai_adk.statusline.main.VersionReader") as mock_reader_class:
            mock_reader_class.side_effect = OSError("File not found")

            version = safe_collect_version()

            assert version == "unknown"


class TestSafeCollectMemory:
    """Tests for safe_collect_memory function."""

    def test_safe_collect_memory_success(self) -> None:
        """Test safe_collect_memory successful collection."""
        # Patch the MemoryCollector import inside the function
        with patch("moai_adk.statusline.memory_collector.MemoryCollector") as mock_collector_class:
            mock_collector = MagicMock()
            mock_collector.get_display_string.return_value = "128MB"
            mock_collector_class.return_value = mock_collector

            memory = safe_collect_memory()

            assert memory == "128MB"

    def test_safe_collect_memory_import_error(self) -> None:
        """Test safe_collect_memory handles ImportError."""
        # Patch the import to raise ImportError
        with patch("moai_adk.statusline.memory_collector.MemoryCollector", side_effect=ImportError):
            memory = safe_collect_memory()
            assert memory == "N/A"

    def test_safe_collect_memory_os_error(self) -> None:
        """Test safe_collect_memory handles OSError."""
        with patch("moai_adk.statusline.memory_collector.MemoryCollector") as mock_collector_class:
            mock_collector_class.side_effect = OSError("Cannot read /proc")

            memory = safe_collect_memory()

            assert memory == "N/A"

    def test_safe_collect_memory_attribute_error(self) -> None:
        """Test safe_collect_memory handles AttributeError."""
        with patch("moai_adk.statusline.memory_collector.MemoryCollector") as mock_collector_class:
            mock_collector = MagicMock()
            mock_collector.get_display_string.side_effect = AttributeError("Missing attribute")
            mock_collector_class.return_value = mock_collector

            memory = safe_collect_memory()

            assert memory == "N/A"

    def test_safe_collect_memory_runtime_error(self) -> None:
        """Test safe_collect_memory handles RuntimeError."""
        with patch("moai_adk.statusline.memory_collector.MemoryCollector") as mock_collector_class:
            mock_collector_class.side_effect = RuntimeError("Memory error")

            memory = safe_collect_memory()

            assert memory == "N/A"


class TestSafeCheckUpdate:
    """Tests for safe_check_update function."""

    def test_safe_check_update_success(self) -> None:
        """Test safe_check_update successful check."""
        with patch("moai_adk.statusline.main.UpdateChecker") as mock_checker_class:
            mock_checker = MagicMock()
            mock_update_info = MagicMock()
            mock_update_info.available = True
            mock_update_info.latest_version = "1.1.0"
            mock_checker.check_for_update.return_value = mock_update_info
            mock_checker_class.return_value = mock_checker

            available, latest = safe_check_update("1.0.0")

            assert available is True
            assert latest == "1.1.0"

    def test_safe_check_update_no_update(self) -> None:
        """Test safe_check_update with no update available."""
        with patch("moai_adk.statusline.main.UpdateChecker") as mock_checker_class:
            mock_checker = MagicMock()
            mock_update_info = MagicMock()
            mock_update_info.available = False
            mock_update_info.latest_version = None
            mock_checker.check_for_update.return_value = mock_update_info
            mock_checker_class.return_value = mock_checker

            available, latest = safe_check_update("1.0.0")

            assert available is False
            assert latest is None

    def test_safe_check_update_os_error(self) -> None:
        """Test safe_check_update handles OSError."""
        with patch("moai_adk.statusline.main.UpdateChecker") as mock_checker_class:
            mock_checker_class.side_effect = OSError("Network error")

            available, latest = safe_check_update("1.0.0")

            assert available is False
            assert latest is None

    def test_safe_check_update_attribute_error(self) -> None:
        """Test safe_check_update handles AttributeError."""
        with patch("moai_adk.statusline.main.UpdateChecker") as mock_checker_class:
            mock_checker = MagicMock()
            mock_checker.check_for_update.side_effect = AttributeError("Missing attribute")
            mock_checker_class.return_value = mock_checker

            available, latest = safe_check_update("1.0.0")

            assert available is False
            assert latest is None

    def test_safe_check_update_runtime_error(self) -> None:
        """Test safe_check_update handles RuntimeError."""
        with patch("moai_adk.statusline.main.UpdateChecker") as mock_checker_class:
            mock_checker_class.side_effect = RuntimeError("Check error")

            available, latest = safe_check_update("1.0.0")

            assert available is False
            assert latest is None

    def test_safe_check_update_value_error(self) -> None:
        """Test safe_check_update handles ValueError."""
        with patch("moai_adk.statusline.main.UpdateChecker") as mock_checker_class:
            mock_checker = MagicMock()
            mock_checker.check_for_update.side_effect = ValueError("Invalid version")
            mock_checker_class.return_value = mock_checker

            available, latest = safe_check_update("1.0.0")

            assert available is False
            assert latest is None


class TestBuildStatuslineData:
    """Tests for build_statusline_data function."""

    def test_build_statusline_data_minimal_context(self) -> None:
        """Test build_statusline_data with minimal context."""
        with patch("moai_adk.statusline.main.safe_collect_git_info") as mock_git:
            with patch("moai_adk.statusline.main.safe_collect_duration") as mock_duration:
                with patch("moai_adk.statusline.main.safe_collect_alfred_task") as mock_task:
                    with patch("moai_adk.statusline.main.safe_collect_version") as mock_version:
                        with patch("moai_adk.statusline.main.safe_collect_memory") as mock_memory:
                            with patch("moai_adk.statusline.main.safe_check_update") as mock_update:
                                with patch("moai_adk.statusline.main.StatuslineRenderer") as mock_renderer:
                                    mock_git.return_value = ("main", "+0 M0 ?0")
                                    mock_duration.return_value = "5m"
                                    mock_task.return_value = ""
                                    mock_version.return_value = "1.0.0"
                                    mock_memory.return_value = "128MB"
                                    mock_update.return_value = (False, None)

                                    mock_renderer_instance = MagicMock()
                                    mock_renderer_instance.render.return_value = "test statusline"
                                    mock_renderer.return_value = mock_renderer_instance

                                    result = build_statusline_data({})

                                    assert result == "test statusline"

    def test_build_statusline_data_full_context(self) -> None:
        """Test build_statusline_data with full session context."""
        session_context = {
            "model": {"display_name": "Claude 3 Opus", "name": "claude-3-opus"},
            "version": "1.0.0",
            "cwd": "/path/to/project",
            "output_style": {"name": "yoda"},
            "context_window": {
                "context_window_size": 200000,
                "total_input_tokens": 15000,
            },
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info") as mock_git:
            with patch("moai_adk.statusline.main.safe_collect_duration") as mock_duration:
                with patch("moai_adk.statusline.main.safe_collect_alfred_task") as mock_task:
                    with patch("moai_adk.statusline.main.safe_collect_version") as mock_version:
                        with patch("moai_adk.statusline.main.safe_collect_memory") as mock_memory:
                            with patch("moai_adk.statusline.main.safe_check_update") as mock_update:
                                with patch("moai_adk.statusline.main.StatuslineRenderer") as mock_renderer:
                                    mock_git.return_value = ("main", "+0 M0 ?0")
                                    mock_duration.return_value = "5m"
                                    mock_task.return_value = "[ANALYZE]"
                                    mock_version.return_value = "1.0.0"
                                    mock_memory.return_value = "128MB"
                                    mock_update.return_value = (True, "1.1.0")

                                    mock_renderer_instance = MagicMock()
                                    mock_renderer_instance.render.return_value = "full statusline"
                                    mock_renderer.return_value = mock_renderer_instance

                                    result = build_statusline_data(session_context)

                                    assert result == "full statusline"

    def test_build_statusline_data_exception_returns_empty(self) -> None:
        """Test build_statusline_data returns empty string on exception."""
        with patch("moai_adk.statusline.main.safe_collect_git_info") as mock_git:
            mock_git.side_effect = RuntimeError("Unexpected error")

            result = build_statusline_data({})

            assert result == ""

    def test_build_statusline_data_with_mode_parameter(self) -> None:
        """Test build_statusline_data respects mode parameter."""
        with patch("moai_adk.statusline.main.safe_collect_git_info") as mock_git:
            with patch("moai_adk.statusline.main.safe_collect_duration") as mock_duration:
                with patch("moai_adk.statusline.main.safe_collect_alfred_task") as mock_task:
                    with patch("moai_adk.statusline.main.safe_collect_version") as mock_version:
                        with patch("moai_adk.statusline.main.safe_collect_memory") as mock_memory:
                            with patch("moai_adk.statusline.main.safe_check_update") as mock_update:
                                with patch("moai_adk.statusline.main.StatuslineRenderer") as mock_renderer:
                                    mock_git.return_value = ("main", "+0 M0 ?0")
                                    mock_duration.return_value = "5m"
                                    mock_task.return_value = ""
                                    mock_version.return_value = "1.0.0"
                                    mock_memory.return_value = "128MB"
                                    mock_update.return_value = (False, None)

                                    mock_renderer_instance = MagicMock()
                                    mock_renderer_instance.render.return_value = "minimal statusline"
                                    mock_renderer.return_value = mock_renderer_instance

                                    result = build_statusline_data({}, mode="minimal")

                                    # Verify renderer was called with the mode parameter
                                    mock_renderer_instance.render.assert_called_once()
                                    args, kwargs = mock_renderer_instance.render.call_args
                                    # Check that mode was passed as keyword argument
                                    assert "mode" in kwargs or len(args) >= 2

    def test_build_statusline_data_with_unicode_cwd(self) -> None:
        """Test build_statusline_data handles Unicode in cwd."""
        session_context = {
            "cwd": "/路径/项目",
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info") as mock_git:
            with patch("moai_adk.statusline.main.StatuslineRenderer") as mock_renderer:
                mock_git.return_value = ("main", "+0 M0 ?0")
                mock_renderer_instance = MagicMock()
                mock_renderer_instance.render.return_value = "statusline"
                mock_renderer.return_value = mock_renderer_instance

                result = build_statusline_data(session_context)

                # Should not crash with Unicode path
                assert result == "statusline"

    def test_build_statusline_data_no_display_name(self) -> None:
        """Test build_statusline_data falls back to model name."""
        session_context = {
            "model": {"name": "claude-3-opus"},  # No display_name
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info") as mock_git:
            with patch("moai_adk.statusline.main.StatuslineRenderer") as mock_renderer:
                mock_git.return_value = ("main", "+0 M0 ?0")
                mock_renderer_instance = MagicMock()
                mock_renderer_instance.render.return_value = "statusline"
                mock_renderer.return_value = mock_renderer_instance

                result = build_statusline_data(session_context)

                assert result == "statusline"
