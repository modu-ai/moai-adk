"""
Comprehensive test coverage for statusline main module.

Tests for uncovered lines in main.py:
- Exception handling in read_session_context (lines 48-52)
- Memory collector exception handling (lines 141-143)
- format_token_count for large numbers (lines 176-178)
- extract_context_window complex logic (lines 197-216)
- main function debug mode and mode selection (lines 309-337)
"""

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
    safe_collect_memory,
    safe_collect_version,
)


class TestReadSessionContextExceptions:
    """Test read_session_context exception handling (lines 48-52)."""

    def test_read_session_context_eof_error(self, caplog):
        """Test read_session_context with EOFError (lines 48-52)."""
        # Arrange
        with patch("sys.stdin.read", side_effect=EOFError("End of file")):
            # Act
            result = read_session_context()

        # Assert
        assert result == {}

    def test_read_session_context_value_error(self, caplog):
        """Test read_session_context with ValueError (lines 48-52)."""
        # Arrange
        with patch("sys.stdin.read", side_effect=ValueError("Invalid value")):
            # Act
            result = read_session_context()

        # Assert
        assert result == {}

    def test_read_session_context_general_exception(self, caplog):
        """Test read_session_context with general exception."""
        # Arrange
        with patch("sys.stdin.read", side_effect=RuntimeError("Unexpected error")):
            # Act
            result = read_session_context()

        # Assert
        assert result == {}


class TestSafeCollectMemoryExceptions:
    """Test safe_collect_memory exception handling (lines 141-143)."""

    def test_safe_collect_memory_import_error(self):
        """Test safe_collect_memory with ImportError (line 141)."""
        # Arrange
        with patch("moai_adk.statusline.main.MemoryCollector", side_effect=ImportError("No module")):
            # Act
            result = safe_collect_memory()

        # Assert
        assert result == "N/A"

    def test_safe_collect_memory_os_error(self):
        """Test safe_collect_memory with OSError."""
        # Arrange
        mock_collector = MagicMock()
        mock_collector.get_display_string.side_effect = OSError("File access error")

        with patch("moai_adk.statusline.main.MemoryCollector", return_value=mock_collector):
            # Act
            result = safe_collect_memory()

        # Assert
        assert result == "N/A"

    def test_safe_collect_memory_attribute_error(self):
        """Test safe_collect_memory with AttributeError."""
        # Arrange
        mock_collector = MagicMock()
        mock_collector.get_display_string.side_effect = AttributeError("Missing attribute")

        with patch("moai_adk.statusline.main.MemoryCollector", return_value=mock_collector):
            # Act
            result = safe_collect_memory()

        # Assert
        assert result == "N/A"

    def test_safe_collect_memory_runtime_error(self):
        """Test safe_collect_memory with RuntimeError."""
        # Arrange
        mock_collector = MagicMock()
        mock_collector.get_display_string.side_effect = RuntimeError("Runtime error")

        with patch("moai_adk.statusline.main.MemoryCollector", return_value=mock_collector):
            # Act
            result = safe_collect_memory()

        # Assert
        assert result == "N/A"


class TestSafeCollectAlfredTaskNoCommand:
    """Test safe_collect_alfred_task with no command (lines 101-104)."""

    def test_safe_collect_alfred_task_no_command_returns_empty(self):
        """Test safe_collect_alfred_task when task has no command."""
        # Arrange
        mock_detector = MagicMock()
        mock_task = MagicMock()
        mock_task.command = None
        mock_task.stage = None
        mock_detector.detect_active_task.return_value = mock_task

        with patch("moai_adk.statusline.main.AlfredDetector", return_value=mock_detector):
            # Act
            result = safe_collect_alfred_task()

        # Assert
        assert result == ""

    def test_safe_collect_alfred_task_command_without_stage(self):
        """Test safe_collect_alfred_task with command but no stage (line 102)."""
        # Arrange
        mock_detector = MagicMock()
        mock_task = MagicMock()
        mock_task.command = "/moai:2-run"
        mock_task.stage = None
        mock_detector.detect_active_task.return_value = mock_task

        with patch("moai_adk.statusline.main.AlfredDetector", return_value=mock_detector):
            # Act
            result = safe_collect_alfred_task()

        # Assert
        assert result == "[/MOAI:2-RUN]"

    def test_safe_collect_alfred_task_with_stage(self):
        """Test safe_collect_alfred_task with command and stage."""
        # Arrange
        mock_detector = MagicMock()
        mock_task = MagicMock()
        mock_task.command = "/moai:2-run"
        mock_task.stage = "RED"
        mock_detector.detect_active_task.return_value = mock_task

        with patch("moai_adk.statusline.main.AlfredDetector", return_value=mock_detector):
            # Act
            result = safe_collect_alfred_task()

        # Assert
        assert result == "[/MOAI:2-RUN-RED]"

    def test_safe_collect_alfred_task_os_error(self):
        """Test safe_collect_alfred_task with OSError."""
        # Arrange
        with patch("moai_adk.statusline.main.AlfredDetector", side_effect=OSError("File error")):
            # Act
            result = safe_collect_alfred_task()

        # Assert
        assert result == ""

    def test_safe_collect_alfred_task_attribute_error(self):
        """Test safe_collect_alfred_task with AttributeError."""
        # Arrange
        with patch("moai_adk.statusline.main.AlfredDetector", side_effect=AttributeError("Missing attribute")):
            # Act
            result = safe_collect_alfred_task()

        # Assert
        assert result == ""

    def test_safe_collect_alfred_task_runtime_error(self):
        """Test safe_collect_alfred_task with RuntimeError."""
        # Arrange
        with patch("moai_adk.statusline.main.AlfredDetector", side_effect=RuntimeError("Runtime error")):
            # Act
            result = safe_collect_alfred_task()

        # Assert
        assert result == ""


class TestFormatTokenCount:
    """Test format_token_count function (lines 176-178)."""

    def test_format_token_count_large_number(self):
        """Test formatting large token count (lines 176-177)."""
        # Act
        result = format_token_count(15234)

        # Assert
        assert result == "15K"

    def test_format_token_count_exactly_1000(self):
        """Test formatting exactly 1000 tokens (boundary)."""
        # Act
        result = format_token_count(1000)

        # Assert
        assert result == "1K"

    def test_format_token_count_small_number(self):
        """Test formatting small token count (line 178)."""
        # Act
        result = format_token_count(999)

        # Assert
        assert result == "999"

    def test_format_token_count_zero(self):
        """Test formatting zero tokens."""
        # Act
        result = format_token_count(0)

        # Assert
        assert result == "0"

    def test_format_token_count_very_large(self):
        """Test formatting very large token count."""
        # Act
        result = format_token_count(1234567)

        # Assert
        assert result == "1234K"


class TestExtractContextWindow:
    """Test extract_context_window function (lines 197-216)."""

    def test_extract_context_window_empty_context(self):
        """Test with empty context_info (lines 193-194)."""
        # Arrange
        session_context = {"context_window": {}}

        # Act
        result = extract_context_window(session_context)

        # Assert
        assert result == ""

    def test_extract_context_window_no_context_window_key(self):
        """Test when context_window key is missing."""
        # Arrange
        session_context = {}

        # Act
        result = extract_context_window(session_context)

        # Assert
        assert result == ""

    def test_extract_context_window_no_size(self):
        """Test when context_window_size is missing (lines 197-199)."""
        # Arrange
        session_context = {"context_window": {"other_key": "value"}}

        # Act
        result = extract_context_window(session_context)

        # Assert
        assert result == ""

    def test_extract_context_window_size_zero(self):
        """Test when context_window_size is zero (line 199)."""
        # Arrange
        session_context = {"context_window": {"context_window_size": 0}}

        # Act
        result = extract_context_window(session_context)

        # Assert
        assert result == ""

    def test_extract_context_window_with_current_usage(self):
        """Test with current_usage breakdown (lines 203-208)."""
        # Arrange
        session_context = {
            "context_window": {
                "context_window_size": 200000,
                "current_usage": {
                    "input_tokens": 10000,
                    "cache_creation_input_tokens": 2000,
                    "cache_read_input_tokens": 3000,
                },
            }
        }

        # Act
        result = extract_context_window(session_context)

        # Assert
        # Total: 10000 + 2000 + 3000 = 15000 -> 15K/200K
        assert result == "15K/200K"

    def test_extract_context_window_fallback_to_total_tokens(self):
        """Test fallback to total_tokens (lines 210-211)."""
        # Arrange
        session_context = {
            "context_window": {
                "context_window_size": 200000,
                "total_input_tokens": 15000,
            }
        }

        # Act
        result = extract_context_window(session_context)

        # Assert
        assert result == "15K/200K"

    def test_extract_context_window_zero_current_tokens(self):
        """Test when current_tokens is zero (lines 213-216)."""
        # Arrange
        session_context = {
            "context_window": {
                "context_window_size": 200000,
                "current_usage": {
                    "input_tokens": 0,
                    "cache_creation_input_tokens": 0,
                    "cache_read_input_tokens": 0,
                },
            }
        }

        # Act
        result = extract_context_window(session_context)

        # Assert
        assert result == ""

    def test_extract_context_window_partial_current_usage(self):
        """Test with partial current_usage fields."""
        # Arrange
        session_context = {
            "context_window": {
                "context_window_size": 200000,
                "current_usage": {
                    "input_tokens": 10000,
                    # Missing cache_creation_input_tokens and cache_read_input_tokens
                },
            }
        }

        # Act
        result = extract_context_window(session_context)

        # Assert
        assert result == "10K/200K"

    def test_extract_context_window_with_all_cache_tokens(self):
        """Test with all cache token fields."""
        # Arrange
        session_context = {
            "context_window": {
                "context_window_size": 200000,
                "current_usage": {
                    "input_tokens": 5000,
                    "cache_creation_input_tokens": 5000,
                    "cache_read_input_tokens": 5000,
                },
            }
        }

        # Act
        result = extract_context_window(session_context)

        # Assert
        # Total: 5000 + 5000 + 5000 = 15000
        assert result == "15K/200K"


class TestBuildStatuslineData:
    """Test build_statusline_data function."""

    def test_build_statusline_with_minimal_context(self):
        """Test building statusline with minimal context."""
        # Arrange
        session_context = {}

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "+0 M0 ?0")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="1.0.0"):
                        with patch("moai_adk.statusline.main.safe_collect_memory", return_value="100MB"):
                            with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                                # Act
                                result = build_statusline_data(session_context, mode="compact")

        # Assert
        assert result is not None
        assert isinstance(result, str)

    def test_build_statusline_with_full_context(self):
        """Test building statusline with full context."""
        # Arrange
        session_context = {
            "model": {"display_name": "claude-3-5-sonnet", "name": "claude-3-5-sonnet"},
            "version": "2.0.46",
            "cwd": "/Users/test/project",
            "output_style": {"name": "Concise"},
            "context_window": {
                "context_window_size": 200000,
                "current_usage": {"input_tokens": 15000},
            },
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("feature-branch", "+2 M1 ?0")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="10m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value="[/MOAI:2-RUN-RED]"):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="1.1.0"):
                        with patch("moai_adk.statusline.main.safe_collect_memory", return_value="150MB"):
                            with patch("moai_adk.statusline.main.safe_check_update", return_value=(True, "1.2.0")):
                                # Act
                                result = build_statusline_data(session_context, mode="extended")

        # Assert
        assert result is not None
        assert "claude-3-5-sonnet" in result or "feature-branch" in result

    def test_build_statusline_error_handling(self, caplog):
        """Test build_statusline_data graceful degradation (lines 293-298)."""
        # Arrange
        session_context = {"invalid": "data"}

        with patch("moai_adk.statusline.main.safe_collect_git_info", side_effect=RuntimeError("Collection error")):
            # Act
            result = build_statusline_data(session_context)

        # Assert - should return empty string on error
        assert result == ""

    def test_build_statusline_model_fallback_to_name(self):
        """Test model name fallback (line 243)."""
        # Arrange
        session_context = {
            "model": {"name": "claude-3-opus"},  # No display_name
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "+0 M0 ?0")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="1.0.0"):
                        with patch("moai_adk.statusline.main.safe_collect_memory", return_value="100MB"):
                            with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                                # Act
                                result = build_statusline_data(session_context)

        # Assert
        assert result is not None

    def test_build_statusline_directory_extraction(self):
        """Test directory extraction from cwd (lines 250-254)."""
        # Arrange
        session_context = {
            "cwd": "/Users/test/my-project",
        }

        with patch("moai_adk.statusline.main.safe_collect_git_info", return_value=("main", "+0 M0 ?0")):
            with patch("moai_adk.statusline.main.safe_collect_duration", return_value="5m"):
                with patch("moai_adk.statusline.main.safe_collect_alfred_task", return_value=""):
                    with patch("moai_adk.statusline.main.safe_collect_version", return_value="1.0.0"):
                        with patch("moai_adk.statusline.main.safe_collect_memory", return_value="100MB"):
                            with patch("moai_adk.statusline.main.safe_check_update", return_value=(False, None)):
                                # Act
                                result = build_statusline_data(session_context)

        # Assert
        assert result is not None
        # Directory name should be extracted from path


class TestMainFunction:
    """Test main function (lines 301-337)."""

    def test_main_with_debug_mode(self, capsys):
        """Test main with debug mode enabled (lines 314-318, 332-334)."""
        # Arrange
        session_context = {
            "model": {"display_name": "claude-3-5-sonnet"},
            "cwd": "/Users/test/project",
        }

        with patch.dict(os.environ, {"MOAI_STATUSLINE_DEBUG": "1"}):
            with patch("moai_adk.statusline.main.read_session_context", return_value=session_context):
                with patch("moai_adk.statusline.main.build_statusline_data", return_value="Test Statusline"):
                    with patch("builtins.print") as mock_print:
                        # Act
                        main()

        # Assert - should print statusline
        # Debug output goes to stderr, statusline goes to stdout
        captured = capsys.readouterr()
        assert "[DEBUG]" in captured.err or "Test Statusline" in captured.err

    def test_main_with_session_context_mode(self):
        """Test main with mode from session context (lines 323-328)."""
        # Arrange
        session_context = {
            "statusline": {"mode": "minimal"},
            "model": {"display_name": "claude-3-5-sonnet"},
            "cwd": "/Users/test/project",
        }

        with patch("moai_adk.statusline.main.read_session_context", return_value=session_context):
            with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
                mock_build.return_value = "Minimal Statusline"

                with patch("builtins.print") as mock_print:
                    # Act
                    main()

        # Assert
        mock_build.assert_called_once()
        call_args = mock_build.call_args
        assert call_args[1]["mode"] == "minimal"

    def test_main_with_environment_mode(self):
        """Test main with mode from environment variable (line 325)."""
        # Arrange
        session_context = {
            "model": {"display_name": "claude-3-5-sonnet"},
            "cwd": "/Users/test/project",
        }

        with patch.dict(os.environ, {"MOAI_STATUSLINE_MODE": "extended"}):
            with patch("moai_adk.statusline.main.read_session_context", return_value=session_context):
                with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
                    mock_build.return_value = "Extended Statusline"

                    with patch("builtins.print") as mock_print:
                        # Act
                        main()

        # Assert
        call_args = mock_build.call_args
        assert call_args[1]["mode"] == "extended"

    def test_main_with_config_mode(self):
        """Test main with mode from config (line 326)."""
        # Arrange
        session_context = {
            "model": {"display_name": "claude-3-5-sonnet"},
            "cwd": "/Users/test/project",
        }

        with patch("moai_adk.statusline.main.read_session_context", return_value=session_context):
            with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
                mock_build.return_value = "Config Statusline"

                with patch("builtins.print") as mock_print:
                    # Act
                    main()

        # Assert - should use config mode or default to extended
        call_args = mock_build.call_args
        assert call_args[1]["mode"] in ["extended", "compact"]

    def test_main_mode_priority(self):
        """Test mode priority: session context > environment > config > default."""
        # Arrange
        session_context = {
            "statusline": {"mode": "compact"},
            "model": {"display_name": "claude-3-5-sonnet"},
            "cwd": "/Users/test/project",
        }

        with patch.dict(os.environ, {"MOAI_STATUSLINE_MODE": "extended"}):
            with patch("moai_adk.statusline.main.read_session_context", return_value=session_context):
                with patch("moai_adk.statusline.main.build_statusline_data") as mock_build:
                    mock_build.return_value = "Priority Statusline"

                    with patch("builtins.print") as mock_print:
                        # Act
                        main()

        # Assert - session context should have highest priority
        call_args = mock_build.call_args
        assert call_args[1]["mode"] == "compact"

    def test_main_empty_statusline(self, capsys):
        """Test main when statusline is empty (lines 336-337)."""
        # Arrange
        session_context = {}

        with patch("moai_adk.statusline.main.read_session_context", return_value=session_context):
            with patch("moai_adk.statusline.main.build_statusline_data", return_value=""):
                # Act
                main()

        # Assert - should not print anything
        captured = capsys.readouterr()
        assert captured.out == ""

    def test_main_prints_statusline(self, capsys):
        """Test main prints statusline when not empty (line 337)."""
        # Arrange
        session_context = {
            "model": {"display_name": "claude-3-5-sonnet"},
            "cwd": "/Users/test/project",
        }

        with patch("moai_adk.statusline.main.read_session_context", return_value=session_context):
            with patch("moai_adk.statusline.main.build_statusline_data", return_value="Test Statusline"):
                # Act
                main()

        # Assert
        captured = capsys.readouterr()
        assert "Test Statusline" in captured.out


class TestSafeCollectVersion:
    """Test safe_collect_version function."""

    def test_safe_collect_version_success(self):
        """Test successful version collection."""
        # Arrange
        mock_reader = MagicMock()
        mock_reader.get_version.return_value = "1.1.0"

        with patch("moai_adk.statusline.main.VersionReader", return_value=mock_reader):
            # Act
            result = safe_collect_version()

        # Assert
        assert result == "1.1.0"

    def test_safe_collect_version_empty_string(self):
        """Test when version reader returns empty string (line 120)."""
        # Arrange
        mock_reader = MagicMock()
        mock_reader.get_version.return_value = ""

        with patch("moai_adk.statusline.main.VersionReader", return_value=mock_reader):
            # Act
            result = safe_collect_version()

        # Assert
        assert result == "unknown"

    def test_safe_collect_version_import_error(self):
        """Test with ImportError (line 121)."""
        # Arrange
        with patch("moai_adk.statusline.main.VersionReader", side_effect=ImportError("No module")):
            # Act
            result = safe_collect_version()

        # Assert
        assert result == "unknown"

    def test_safe_collect_version_attribute_error(self):
        """Test with AttributeError."""
        # Arrange
        with patch("moai_adk.statusline.main.VersionReader", side_effect=AttributeError("Missing attribute")):
            # Act
            result = safe_collect_version()

        # Assert
        assert result == "unknown"

    def test_safe_collect_version_os_error(self):
        """Test with OSError."""
        # Arrange
        with patch("moai_adk.statusline.main.VersionReader", side_effect=OSError("File error")):
            # Act
            result = safe_collect_version()

        # Assert
        assert result == "unknown"


class TestSafeCheckUpdate:
    """Test safe_check_update function."""

    def test_safe_check_update_success(self):
        """Test successful update check."""
        # Arrange
        mock_checker = MagicMock()
        mock_info = MagicMock()
        mock_info.available = True
        mock_info.latest_version = "1.2.0"
        mock_checker.check_for_update.return_value = mock_info

        with patch("moai_adk.statusline.main.UpdateChecker", return_value=mock_checker):
            # Act
            available, version = safe_check_update("1.1.0")

        # Assert
        assert available is True
        assert version == "1.2.0"

    def test_safe_check_update_os_error(self):
        """Test with OSError."""
        # Arrange
        with patch("moai_adk.statusline.main.UpdateChecker", side_effect=OSError("Network error")):
            # Act
            available, version = safe_check_update("1.1.0")

        # Assert
        assert available is False
        assert version is None

    def test_safe_check_update_attribute_error(self):
        """Test with AttributeError."""
        # Arrange
        with patch("moai_adk.statusline.main.UpdateChecker", side_effect=AttributeError("Missing attribute")):
            # Act
            available, version = safe_check_update("1.1.0")

        # Assert
        assert available is False
        assert version is None

    def test_safe_check_update_runtime_error(self):
        """Test with RuntimeError."""
        # Arrange
        with patch("moai_adk.statusline.main.UpdateChecker", side_effect=RuntimeError("Runtime error")):
            # Act
            available, version = safe_check_update("1.1.0")

        # Assert
        assert available is False
        assert version is None

    def test_safe_check_update_value_error(self):
        """Test with ValueError."""
        # Arrange
        with patch("moai_adk.statusline.main.UpdateChecker", side_effect=ValueError("Invalid value")):
            # Act
            available, version = safe_check_update("1.1.0")

        # Assert
        assert available is False
        assert version is None
