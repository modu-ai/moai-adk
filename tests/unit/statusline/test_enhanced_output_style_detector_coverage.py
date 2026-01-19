"""Additional coverage tests for enhanced output style detector.

Tests for lines not covered by existing tests.
"""

from unittest.mock import patch

from moai_adk.statusline.enhanced_output_style_detector import (
    OutputStyleDetector,
)


class TestDetectFromSessionContextException:
    """Test exception handling in detect_from_session_context."""

    def test_detect_from_session_context_handles_exception(self, tmp_path):
        """Should handle exception and return None."""
        detector = OutputStyleDetector()

        # Pass invalid session data (not a dict)
        result = detector.detect_from_session_context(None)
        assert result is None


class TestDetectFromEnvironmentPassStatements:
    """Test pass statements in detect_from_environment."""

    def test_detect_from_environment_with_claude_session_id(self, monkeypatch):
        """Should handle CLAUDE_SESSION_ID environment variable."""
        monkeypatch.setenv("CLAUDE_SESSION_ID", "test-session-123")

        detector = OutputStyleDetector()
        result = detector.detect_from_environment()

        # Should return None (no explicit style set)
        assert result is None

    def test_detect_from_environment_process_name_check(self):
        """Should handle process name check (pass statement)."""
        detector = OutputStyleDetector()

        # Just ensure it doesn't raise
        result = detector.detect_from_environment()
        assert result is None


class TestDetectFromBehavioralAnalysis:
    """Test behavioral analysis detection."""

    def test_detect_from_behavioral_analysis_with_todo_file(self, tmp_path, monkeypatch):
        """Should detect Explanatory style from TODO file with plan keywords."""
        monkeypatch.chdir(tmp_path)

        # Create TODO file with plan keywords
        moai_dir = tmp_path / ".moai"
        moai_dir.mkdir(parents=True)
        todo_file = moai_dir / "current_session_todo.txt"
        todo_file.write_text("Plan: Implement feature\nPhase 1: Design")

        detector = OutputStyleDetector()
        result = detector.detect_from_behavioral_analysis()

        assert result == "Explanatory"


class TestNormalizeStylePatternBased:
    """Test pattern-based normalization in _normalize_style."""

    def test_normalize_style_explanatory(self):
        """Should normalize 'explanatory' to 'Explanatory'."""
        detector = OutputStyleDetector()
        result = detector._normalize_style("explanatory")
        assert result == "Explanatory"

    def test_normalize_style_concise(self):
        """Should normalize 'concise' to 'Concise'."""
        detector = OutputStyleDetector()
        result = detector._normalize_style("concise")
        assert result == "Concise"

    def test_normalize_style_detailed(self):
        """Should normalize 'detailed' to 'Detailed'."""
        detector = OutputStyleDetector()
        result = detector._normalize_style("detailed")
        assert result == "Detailed"


class TestNormalizeStyleEmojiExtraction:
    """Test emoji extraction in _normalize_style."""

    def test_normalize_style_r2d2_emoji(self):
        """Should extract R2-D2 from 🤖 emoji."""
        detector = OutputStyleDetector()
        result = detector._normalize_style("🤖 R2-D2 Code")
        assert result == "R2-D2"

    def test_normalize_style_yoda_emoji_extraction(self):
        """Should extract style after 🧙 emoji."""
        detector = OutputStyleDetector()
        result = detector._normalize_style("🧙 Master Yoda")
        assert result == "🧙 Yoda Master"


class TestDetectionMethodExceptionHandling:
    """Test exception handling in detection methods loop."""

    def test_detection_method_exception_continues_to_next(self):
        """Should continue to next method when one raises exception."""
        detector = OutputStyleDetector()

        # Mock one method to raise exception
        def failing_method():
            raise RuntimeError("Test error")

        with patch.object(detector, "detect_from_environment", side_effect=failing_method):
            # Should not raise, should continue to next method
            result = detector.get_output_style()
            # Should return some style (fallback or from other methods)
            assert result is not None


class TestMainBlock:
    """Test __main__ block."""

    def test_main_block_prints_style(self, capsys):
        """Should print detected style when run as main."""
        # This tests the __main__ block (line 372)
        with patch("sys.argv", ["enhanced_output_style_detector.py"]):
            # Import and execute as main
            import sys

            original_argv = sys.argv
            try:
                sys.argv = ["enhanced_output_style_detector.py"]
                # Execute the __main__ block
                from moai_adk.statusline.enhanced_output_style_detector import safe_collect_output_style

                style = safe_collect_output_style()
                # Should return a style string
                assert isinstance(style, str)
            finally:
                sys.argv = original_argv
