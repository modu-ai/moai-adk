"""Tests for moai_adk.statusline.enhanced_output_style_detector module."""

import json
from unittest.mock import MagicMock, patch

from moai_adk.statusline.enhanced_output_style_detector import (
    OutputStyleDetector,
    safe_collect_output_style,
)


class TestOutputStyleDetector:
    """Test OutputStyleDetector class."""

    def test_init(self):
        """Test OutputStyleDetector initialization."""
        detector = OutputStyleDetector()
        assert detector.cache == {}
        assert detector.cache_ttl == 5

    def test_style_mapping_contains_expected_styles(self):
        """Test that STYLE_MAPPING contains all expected styles."""
        detector = OutputStyleDetector()
        assert "streaming" in detector.STYLE_MAPPING
        assert "explanatory" in detector.STYLE_MAPPING
        assert "concise" in detector.STYLE_MAPPING
        assert "detailed" in detector.STYLE_MAPPING
        assert "yoda" in detector.STYLE_MAPPING

    def test_normalize_style_direct_mapping(self):
        """Test _normalize_style with direct mapping."""
        detector = OutputStyleDetector()
        assert detector._normalize_style("streaming") == "R2-D2"
        assert detector._normalize_style("explanatory") == "Explanatory"
        assert detector._normalize_style("concise") == "Concise"
        assert detector._normalize_style("detailed") == "Detailed"

    def test_normalize_style_case_insensitive(self):
        """Test _normalize_style with case variations."""
        detector = OutputStyleDetector()
        assert detector._normalize_style("STREAMING") == "R2-D2"
        assert detector._normalize_style("Explanatory") == "Explanatory"
        assert detector._normalize_style("CONCISE") == "Concise"

    def test_normalize_style_empty_string(self):
        """Test _normalize_style with empty string."""
        detector = OutputStyleDetector()
        assert detector._normalize_style("") == "Unknown"

    def test_normalize_style_pattern_matching(self):
        """Test _normalize_style with pattern-based matching."""
        detector = OutputStyleDetector()
        assert detector._normalize_style("r2d2") == "R2-D2"
        assert detector._normalize_style("R2D2") == "R2-D2"
        assert detector._normalize_style("streaming mode") == "R2-D2"
        assert detector._normalize_style("yoda mode") == "🧙 Yoda Master"
        assert detector._normalize_style("master tutorial") == "🧙 Yoda Master"

    def test_normalize_style_emoji_extraction(self):
        """Test _normalize_style emoji extraction."""
        detector = OutputStyleDetector()
        assert detector._normalize_style("🤖 R2-D2") == "R2-D2"

    def test_normalize_style_fallback(self):
        """Test _normalize_style fallback to title case."""
        detector = OutputStyleDetector()
        result = detector._normalize_style("CustomStyle")
        assert isinstance(result, str)

    def test_detect_from_session_context_with_outputstyle(self):
        """Test detect_from_session_context with explicit outputStyle."""
        detector = OutputStyleDetector()
        session_data = {"outputStyle": "streaming"}
        result = detector.detect_from_session_context(session_data)
        assert result == "R2-D2"

    def test_detect_from_session_context_with_model_name(self):
        """Test detect_from_session_context with model name indicators."""
        detector = OutputStyleDetector()
        session_data = {
            "model": {
                "name": "explanatory-model",
                "display_name": "Explanatory Claude",
            }
        }
        result = detector.detect_from_session_context(session_data)
        assert result == "Explanatory"

    def test_detect_from_session_context_with_yoda(self):
        """Test detect_from_session_context with yoda indicator."""
        detector = OutputStyleDetector()
        session_data = {
            "model": {
                "name": "yoda-master",
                "display_name": "Yoda Tutorial",
            }
        }
        result = detector.detect_from_session_context(session_data)
        assert result == "🧙 Yoda Master"

    def test_detect_from_session_context_with_messages(self):
        """Test detect_from_session_context with message patterns."""
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

    def test_detect_from_environment_with_env_var(self):
        """Test detect_from_environment with CLAUDE_OUTPUT_STYLE env var."""
        detector = OutputStyleDetector()
        with patch.dict("os.environ", {"CLAUDE_OUTPUT_STYLE": "concise"}):
            result = detector.detect_from_environment()
            assert result == "Concise"

    def test_detect_from_environment_no_env_var(self):
        """Test detect_from_environment without env var."""
        detector = OutputStyleDetector()
        with patch.dict("os.environ", {}, clear=True):
            result = detector.detect_from_environment()
            assert result is None

    def test_get_output_style_default_fallback(self):
        """Test get_output_style returns R2-D2 as default."""
        detector = OutputStyleDetector()
        with patch.dict("os.environ", {}, clear=True):
            with patch("pathlib.Path.cwd") as mock_cwd:
                mock_cwd.return_value = MagicMock()
                result = detector.get_output_style(None)
                assert result == "R2-D2"


class TestSafeCollectOutputStyle:
    """Test safe_collect_output_style function."""

    def test_safe_collect_output_style_with_json_input(self):
        """Test safe_collect_output_style with JSON input."""
        json_input = json.dumps({"outputStyle": "streaming"})
        with patch("sys.stdin") as mock_stdin:
            mock_stdin.read.return_value = json_input
            mock_stdin.isatty.return_value = False
            result = safe_collect_output_style()
            assert result == "R2-D2"

    def test_safe_collect_output_style_with_empty_input(self):
        """Test safe_collect_output_style with empty input."""
        with patch("sys.stdin") as mock_stdin:
            mock_stdin.read.return_value = ""
            mock_stdin.isatty.return_value = False
            result = safe_collect_output_style()
            assert result == "R2-D2"

    def test_safe_collect_output_style_with_invalid_json(self):
        """Test safe_collect_output_style with invalid JSON."""
        with patch("sys.stdin") as mock_stdin:
            mock_stdin.read.return_value = "invalid json {{"
            mock_stdin.isatty.return_value = False
            result = safe_collect_output_style()
            assert result == "R2-D2"

    def test_safe_collect_output_style_returns_string(self):
        """Test safe_collect_output_style always returns string."""
        with patch("sys.stdin") as mock_stdin:
            mock_stdin.read.return_value = ""
            mock_stdin.isatty.return_value = False
            result = safe_collect_output_style()
            assert isinstance(result, str)


class TestOutputStyleDetectorEdgeCases:
    """Test edge cases and error conditions."""

    def test_normalize_style_with_special_characters(self):
        """Test _normalize_style with special characters."""
        detector = OutputStyleDetector()
        result = detector._normalize_style("Test@#$%")
        assert isinstance(result, str)

    def test_analyze_message_patterns_empty_messages(self):
        """Test _analyze_message_patterns with empty messages."""
        detector = OutputStyleDetector()
        result = detector._analyze_message_patterns([])
        assert result is None

    def test_analyze_message_patterns_yoda(self):
        """Test _analyze_message_patterns with yoda indicators."""
        detector = OutputStyleDetector()
        messages = [
            {
                "role": "assistant",
                "content": "young padawan, the force is strong. Master wisdom requires patience.",
            }
        ]
        result = detector._analyze_message_patterns(messages)
        assert result == "🧙 Yoda Master"

    def test_analyze_message_patterns_concise(self):
        """Test _analyze_message_patterns with concise response."""
        detector = OutputStyleDetector()
        messages = [{"role": "assistant", "content": "Short answer here."}]
        result = detector._analyze_message_patterns(messages)
        assert result == "Concise"

    def test_multiple_normalization_variations(self):
        """Test multiple style name variations."""
        detector = OutputStyleDetector()
        assert detector._normalize_style("concise") == "Concise"
        assert detector._normalize_style("CONCISE") == "Concise"
        assert detector._normalize_style("Concise") == "Concise"
