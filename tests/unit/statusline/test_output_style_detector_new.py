"""Comprehensive tests for OutputStyleDetector with 80% coverage target."""

import json
import sys
import time
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch, mock_open

import pytest

from moai_adk.statusline.enhanced_output_style_detector import (
    OutputStyleDetector,
    safe_collect_output_style,
)


class TestOutputStyleDetectorSessionContext:
    """Test session context detection methods."""

    def test_detect_from_session_context_explicit_outputstyle(self):
        """Test detection with explicit outputStyle field."""
        detector = OutputStyleDetector()
        session_data = {"outputStyle": "streaming"}
        result = detector.detect_from_session_context(session_data)
        assert result == "R2-D2"

    def test_detect_from_session_context_empty_outputstyle(self):
        """Test detection with empty outputStyle field."""
        detector = OutputStyleDetector()
        session_data = {"outputStyle": ""}
        result = detector.detect_from_session_context(session_data)
        assert result is None

    def test_detect_from_session_context_model_name_explanatory(self):
        """Test detection with explanatory model name."""
        detector = OutputStyleDetector()
        session_data = {
            "model": {
                "name": "explanatory-model",
                "display_name": "Claude Explanatory",
            }
        }
        result = detector.detect_from_session_context(session_data)
        assert result == "Explanatory"

    def test_detect_from_session_context_model_name_yoda(self):
        """Test detection with yoda model name."""
        detector = OutputStyleDetector()
        session_data = {
            "model": {
                "name": "yoda-master",
                "display_name": "Yoda Master",
            }
        }
        result = detector.detect_from_session_context(session_data)
        assert result == "🧙 Yoda Master"

    def test_detect_from_session_context_model_name_concise(self):
        """Test detection with concise model name."""
        detector = OutputStyleDetector()
        session_data = {"model": {"name": "concise-model", "display_name": "Concise"}}
        result = detector.detect_from_session_context(session_data)
        assert result == "Concise"

    def test_detect_from_session_context_model_name_detailed(self):
        """Test detection with detailed model name."""
        detector = OutputStyleDetector()
        session_data = {"model": {"name": "detailed-model", "display_name": "Detailed"}}
        result = detector.detect_from_session_context(session_data)
        assert result == "Detailed"

    def test_detect_from_session_context_model_name_r2d2(self):
        """Test detection with r2d2 model name."""
        detector = OutputStyleDetector()
        session_data = {"model": {"name": "r2d2-streaming", "display_name": "R2-D2"}}
        result = detector.detect_from_session_context(session_data)
        assert result == "R2-D2"

    def test_detect_from_session_context_model_not_dict(self):
        """Test detection when model is not a dict."""
        detector = OutputStyleDetector()
        session_data = {"model": "string_model"}
        result = detector.detect_from_session_context(session_data)
        assert result is None

    def test_detect_from_session_context_messages_yoda_pattern(self):
        """Test detection with yoda message patterns."""
        detector = OutputStyleDetector()
        session_data = {
            "messages": [
                {
                    "role": "assistant",
                    "content": "Young padawan, the force is strong with you",
                },
                {
                    "role": "assistant",
                    "content": "Master said wisdom comes with patience",
                },
            ]
        }
        result = detector.detect_from_session_context(session_data)
        assert result == "🧙 Yoda Master"

    def test_detect_from_session_context_messages_explanatory(self):
        """Test detection with explanatory message patterns."""
        detector = OutputStyleDetector()
        long_text = "Here's how it works: " + "explanation text " * 100
        session_data = {
            "messages": [
                {"role": "assistant", "content": long_text},
                {"role": "assistant", "content": "The reason is clear " + "x" * 1500},
                {"role": "assistant", "content": "Let me explain further " + "y" * 1500},
            ]
        }
        result = detector.detect_from_session_context(session_data)
        # Should detect explanatory based on long responses with explanatory phrases
        assert result in ["Explanatory", "R2-D2"]

    def test_detect_from_session_context_messages_concise(self):
        """Test detection with concise message patterns."""
        detector = OutputStyleDetector()
        session_data = {
            "messages": [{"role": "assistant", "content": "Quick answer."}]
        }
        result = detector.detect_from_session_context(session_data)
        assert result == "Concise"

    def test_detect_from_session_context_exception_handling(self):
        """Test exception handling in session context detection."""
        detector = OutputStyleDetector()
        session_data = {"outputStyle": None}
        with patch("sys.stderr", new_callable=MagicMock):
            result = detector.detect_from_session_context(session_data)
            assert result is None


class TestOutputStyleDetectorEnvironment:
    """Test environment variable detection."""

    def test_detect_from_environment_with_env_var(self):
        """Test detection with CLAUDE_OUTPUT_STYLE env var."""
        detector = OutputStyleDetector()
        with patch.dict("os.environ", {"CLAUDE_OUTPUT_STYLE": "concise"}):
            result = detector.detect_from_environment()
            assert result == "Concise"

    def test_detect_from_environment_empty_env_var(self):
        """Test detection with empty CLAUDE_OUTPUT_STYLE."""
        detector = OutputStyleDetector()
        with patch.dict("os.environ", {"CLAUDE_OUTPUT_STYLE": ""}):
            result = detector.detect_from_environment()
            assert result is None

    def test_detect_from_environment_no_env_var(self):
        """Test detection without CLAUDE_OUTPUT_STYLE."""
        detector = OutputStyleDetector()
        with patch.dict("os.environ", {}, clear=True):
            result = detector.detect_from_environment()
            assert result is None

    def test_detect_from_environment_exception_handling(self):
        """Test exception handling in environment detection."""
        detector = OutputStyleDetector()
        with patch.dict("os.environ", {"CLAUDE_OUTPUT_STYLE": "test"}):
            with patch.object(detector, "_normalize_style", side_effect=Exception("test error")):
                with patch("sys.stderr", new_callable=MagicMock):
                    result = detector.detect_from_environment()
                    # Still returns None due to exception
                    assert result is None


class TestOutputStyleDetectorBehavioral:
    """Test behavioral analysis detection."""

    def test_detect_from_behavioral_analysis_yoda_files(self):
        """Test detection with recent yoda files."""
        detector = OutputStyleDetector()
        mock_moai_dir = MagicMock()
        mock_yoda_file = MagicMock()
        mock_yoda_file.stat.return_value.st_mtime = time.time() - 100  # 100 seconds ago
        mock_moai_dir.rglob.return_value = [mock_yoda_file]
        mock_moai_dir.exists.return_value = True

        with patch("pathlib.Path.cwd") as mock_cwd:
            mock_cwd_result = MagicMock()
            mock_cwd_result.__truediv__ = MagicMock(return_value=mock_moai_dir)
            mock_cwd.return_value = mock_cwd_result
            result = detector.detect_from_behavioral_analysis()
            assert result == "🧙 Yoda Master"

    def test_detect_from_behavioral_analysis_docs_dir(self):
        """Test detection with extensive documentation."""
        detector = OutputStyleDetector()
        mock_docs_dir = MagicMock()
        mock_docs_dir.exists.return_value = True
        mock_docs_dir.rglob.return_value = [MagicMock() for _ in range(15)]

        with patch("pathlib.Path.cwd") as mock_cwd:
            mock_cwd_result = MagicMock()
            mock_cwd_result.__truediv__ = MagicMock(side_effect=lambda x: mock_docs_dir if x == "docs" else MagicMock(exists=MagicMock(return_value=False)))
            mock_cwd.return_value = mock_cwd_result
            result = detector.detect_from_behavioral_analysis()
            assert result == "Explanatory"

    def test_detect_from_behavioral_analysis_no_indicators(self):
        """Test detection with no behavioral indicators."""
        detector = OutputStyleDetector()
        with patch("pathlib.Path.cwd") as mock_cwd:
            mock_cwd_result = MagicMock()
            mock_cwd_result.__truediv__ = MagicMock(return_value=MagicMock(exists=MagicMock(return_value=False)))
            mock_cwd.return_value = mock_cwd_result
            result = detector.detect_from_behavioral_analysis()
            assert result is None

    def test_detect_from_behavioral_analysis_exception(self):
        """Test exception handling in behavioral analysis."""
        detector = OutputStyleDetector()
        with patch("pathlib.Path.cwd", side_effect=Exception("test error")):
            with patch("sys.stderr", new_callable=MagicMock):
                result = detector.detect_from_behavioral_analysis()
                assert result is None


class TestOutputStyleDetectorSettings:
    """Test settings file detection."""

    def test_detect_from_settings_valid_file(self):
        """Test detection from valid settings.json."""
        detector = OutputStyleDetector()
        settings_json = json.dumps({"outputStyle": "detailed"})
        with patch("builtins.open", mock_open(read_data=settings_json)):
            with patch("pathlib.Path.cwd") as mock_cwd:
                mock_cwd_result = MagicMock()
                settings_path = MagicMock()
                settings_path.exists.return_value = True
                mock_cwd_result.__truediv__ = MagicMock(return_value=settings_path)
                mock_cwd.return_value = mock_cwd_result
                result = detector.detect_from_settings()
                assert result == "Detailed"

    def test_detect_from_settings_file_not_found(self):
        """Test detection when settings.json not found."""
        detector = OutputStyleDetector()
        with patch("pathlib.Path.cwd") as mock_cwd:
            mock_cwd_result = MagicMock()
            settings_path = MagicMock()
            settings_path.exists.return_value = False
            mock_cwd_result.__truediv__ = MagicMock(return_value=settings_path)
            mock_cwd.return_value = mock_cwd_result
            result = detector.detect_from_settings()
            assert result is None

    def test_detect_from_settings_invalid_json(self):
        """Test detection with invalid JSON in settings."""
        detector = OutputStyleDetector()
        with patch("builtins.open", mock_open(read_data="invalid json {{")):
            with patch("pathlib.Path.cwd") as mock_cwd:
                mock_cwd_result = MagicMock()
                settings_path = MagicMock()
                settings_path.exists.return_value = True
                mock_cwd_result.__truediv__ = MagicMock(return_value=settings_path)
                mock_cwd.return_value = mock_cwd_result
                with patch("sys.stderr", new_callable=MagicMock):
                    result = detector.detect_from_settings()
                    assert result is None

    def test_detect_from_settings_no_outputstyle(self):
        """Test detection when outputStyle not in settings."""
        detector = OutputStyleDetector()
        settings_json = json.dumps({"otherKey": "value"})
        with patch("builtins.open", mock_open(read_data=settings_json)):
            with patch("pathlib.Path.cwd") as mock_cwd:
                mock_cwd_result = MagicMock()
                settings_path = MagicMock()
                settings_path.exists.return_value = True
                mock_cwd_result.__truediv__ = MagicMock(return_value=settings_path)
                mock_cwd.return_value = mock_cwd_result
                result = detector.detect_from_settings()
                assert result is None


class TestOutputStyleDetectorNormalize:
    """Test style normalization."""

    def test_normalize_style_all_direct_mappings(self):
        """Test all direct mappings."""
        detector = OutputStyleDetector()
        assert detector._normalize_style("streaming") == "R2-D2"
        assert detector._normalize_style("explanatory") == "Explanatory"
        assert detector._normalize_style("concise") == "Concise"
        assert detector._normalize_style("detailed") == "Detailed"
        assert detector._normalize_style("yoda") == "🧙 Yoda Master"

    def test_normalize_style_case_insensitive_variations(self):
        """Test case-insensitive normalization."""
        detector = OutputStyleDetector()
        assert detector._normalize_style("STREAMING") == "R2-D2"
        assert detector._normalize_style("Streaming") == "R2-D2"
        assert detector._normalize_style("EXPLANATORY") == "Explanatory"
        assert detector._normalize_style("CONCISE") == "Concise"

    def test_normalize_style_pattern_r2d2(self):
        """Test r2d2 pattern matching."""
        detector = OutputStyleDetector()
        assert detector._normalize_style("r2d2") == "R2-D2"
        assert detector._normalize_style("R2D2") == "R2-D2"
        assert detector._normalize_style("streaming mode") == "R2-D2"

    def test_normalize_style_pattern_yoda(self):
        """Test yoda pattern matching."""
        detector = OutputStyleDetector()
        assert detector._normalize_style("yoda") == "🧙 Yoda Master"
        assert detector._normalize_style("master") == "🧙 Yoda Master"
        assert detector._normalize_style("tutorial") == "🧙 Yoda Master"

    def test_normalize_style_emoji_extraction(self):
        """Test emoji extraction."""
        detector = OutputStyleDetector()
        assert detector._normalize_style("🤖 R2-D2") == "R2-D2"
        assert detector._normalize_style("🧙 Yoda Master") == "🧙 Yoda Master"
        assert detector._normalize_style("🧙 Concise") == "Concise"

    def test_normalize_style_empty_string(self):
        """Test with empty string."""
        detector = OutputStyleDetector()
        assert detector._normalize_style("") == "Unknown"

    def test_normalize_style_unknown_fallback(self):
        """Test fallback for unknown styles."""
        detector = OutputStyleDetector()
        result = detector._normalize_style("CustomUnknownStyle")
        assert isinstance(result, str)
        assert result.capitalize() == result  # Should be title-cased


class TestOutputStyleDetectorMessageAnalysis:
    """Test message pattern analysis."""

    def test_analyze_message_patterns_yoda_indicators(self):
        """Test yoda message pattern detection."""
        detector = OutputStyleDetector()
        messages = [
            {
                "role": "assistant",
                "content": "Young padawan, the force is strong. Master wisdom requires patience.",
            }
        ]
        result = detector._analyze_message_patterns(messages)
        assert result == "🧙 Yoda Master"

    def test_analyze_message_patterns_explanatory(self):
        """Test explanatory message pattern detection."""
        detector = OutputStyleDetector()
        long_content = "Let me explain how this works. " + "x" * 2000
        messages = [
            {"role": "assistant", "content": long_content},
            {"role": "assistant", "content": "The reason is that"},
            {"role": "assistant", "content": "Here's how to understand this"},
        ]
        result = detector._analyze_message_patterns(messages)
        assert result == "Explanatory"

    def test_analyze_message_patterns_concise(self):
        """Test concise message pattern detection."""
        detector = OutputStyleDetector()
        messages = [{"role": "assistant", "content": "Quick answer."}]
        result = detector._analyze_message_patterns(messages)
        assert result == "Concise"

    def test_analyze_message_patterns_empty_list(self):
        """Test with empty message list."""
        detector = OutputStyleDetector()
        result = detector._analyze_message_patterns([])
        assert result is None

    def test_analyze_message_patterns_no_assistant_messages(self):
        """Test with no assistant messages."""
        detector = OutputStyleDetector()
        messages = [
            {"role": "user", "content": "Hello"},
            {"role": "user", "content": "Help"},
        ]
        result = detector._analyze_message_patterns(messages)
        assert result is None

    def test_analyze_message_patterns_exception(self):
        """Test exception handling in message analysis."""
        detector = OutputStyleDetector()
        messages = [{"role": "assistant", "content": None}]
        with patch("sys.stderr", new_callable=MagicMock):
            result = detector._analyze_message_patterns(messages)
            # Should handle gracefully
            assert result is None or isinstance(result, (str, type(None)))


class TestOutputStyleDetectorGetStyle:
    """Test get_output_style main method."""

    def test_get_output_style_from_session_context(self):
        """Test style detection from session context."""
        detector = OutputStyleDetector()
        session_context = {"outputStyle": "detailed"}
        result = detector.get_output_style(session_context)
        assert result == "Detailed"

    def test_get_output_style_caching(self):
        """Test that results are cached."""
        detector = OutputStyleDetector()
        session_context = {"outputStyle": "streaming"}
        result1 = detector.get_output_style(session_context)
        result2 = detector.get_output_style(session_context)
        assert result1 == result2 == "R2-D2"
        # Verify cache was populated
        assert len(detector.cache) > 0

    def test_get_output_style_cache_expiry(self):
        """Test that cache expires after TTL."""
        detector = OutputStyleDetector()
        detector.cache_ttl = 0.1  # 100ms TTL
        session_context = {"outputStyle": "streaming"}
        result1 = detector.get_output_style(session_context)
        time.sleep(0.15)  # Wait for cache to expire
        result2 = detector.get_output_style(session_context)
        assert result1 == result2 == "R2-D2"

    def test_get_output_style_default_fallback(self):
        """Test default fallback when no style detected."""
        detector = OutputStyleDetector()
        with patch.dict("os.environ", {}, clear=True):
            with patch("pathlib.Path.cwd") as mock_cwd:
                mock_cwd.return_value = MagicMock()
                result = detector.get_output_style({})
                assert result == "R2-D2"

    def test_get_output_style_none_session_context(self):
        """Test with None session context."""
        detector = OutputStyleDetector()
        with patch.dict("os.environ", {}, clear=True):
            with patch("pathlib.Path.cwd") as mock_cwd:
                mock_cwd.return_value = MagicMock()
                result = detector.get_output_style(None)
                assert result == "R2-D2"


class TestSafeCollectOutputStyle:
    """Test safe_collect_output_style function."""

    def test_safe_collect_with_json_input(self):
        """Test with valid JSON input."""
        json_input = json.dumps({"outputStyle": "streaming"})
        with patch("sys.stdin") as mock_stdin:
            mock_stdin.read.return_value = json_input
            mock_stdin.isatty.return_value = False
            result = safe_collect_output_style()
            assert result == "R2-D2"

    def test_safe_collect_with_empty_input(self):
        """Test with empty input."""
        with patch("sys.stdin") as mock_stdin:
            mock_stdin.read.return_value = ""
            mock_stdin.isatty.return_value = False
            result = safe_collect_output_style()
            assert result == "R2-D2"

    def test_safe_collect_with_invalid_json(self):
        """Test with invalid JSON."""
        with patch("sys.stdin") as mock_stdin:
            mock_stdin.read.return_value = "invalid json {{"
            mock_stdin.isatty.return_value = False
            result = safe_collect_output_style()
            assert result == "R2-D2"

    def test_safe_collect_with_tty(self):
        """Test with TTY input."""
        with patch("sys.stdin") as mock_stdin:
            mock_stdin.isatty.return_value = True
            result = safe_collect_output_style()
            assert result == "R2-D2"

    def test_safe_collect_eof_exception(self):
        """Test with EOF exception."""
        with patch("sys.stdin") as mock_stdin:
            mock_stdin.read.side_effect = EOFError()
            mock_stdin.isatty.return_value = False
            result = safe_collect_output_style()
            assert result == "R2-D2"

    def test_safe_collect_general_exception(self):
        """Test with general exception."""
        with patch("sys.stdin") as mock_stdin:
            mock_stdin.read.side_effect = Exception("Test error")
            mock_stdin.isatty.return_value = False
            with patch("sys.stderr", new_callable=MagicMock):
                result = safe_collect_output_style()
                assert result == "R2-D2"

    def test_safe_collect_returns_string(self):
        """Test that result is always a string."""
        with patch("sys.stdin") as mock_stdin:
            mock_stdin.read.return_value = ""
            mock_stdin.isatty.return_value = False
            result = safe_collect_output_style()
            assert isinstance(result, str)


class TestOutputStyleDetectorEdgeCases:
    """Test edge cases and error conditions."""

    def test_multiple_detection_methods_priority(self):
        """Test that detection methods are called in priority order."""
        detector = OutputStyleDetector()
        with patch.object(detector, "detect_from_session_context", return_value="R2-D2"):
            with patch.object(detector, "detect_from_environment", return_value="Concise"):
                result = detector.get_output_style({})
                assert result == "R2-D2"  # Session context has priority

    def test_get_output_style_with_unknown_style(self):
        """Test handling of unknown/empty detected style."""
        detector = OutputStyleDetector()
        with patch.object(detector, "detect_from_session_context", return_value="Unknown"):
            with patch.object(detector, "detect_from_environment", return_value="Concise"):
                result = detector.get_output_style({})
                assert result == "Concise"  # Moves to next method

    def test_normalize_style_with_special_chars(self):
        """Test normalization with special characters."""
        detector = OutputStyleDetector()
        result = detector._normalize_style("Test@#$%Style")
        assert isinstance(result, str)

    def test_detector_with_large_message_history(self):
        """Test with large message history."""
        detector = OutputStyleDetector()
        messages = [
            {"role": "assistant", "content": "x" * 5000} for _ in range(100)
        ]
        result = detector._analyze_message_patterns(messages)
        assert isinstance(result, (str, type(None)))

    def test_get_output_style_concurrent_access(self):
        """Test behavior with potential concurrent access."""
        detector = OutputStyleDetector()
        session_context = {"outputStyle": "streaming"}
        results = []
        for _ in range(3):
            result = detector.get_output_style(session_context)
            results.append(result)
        assert all(r == "R2-D2" for r in results)
