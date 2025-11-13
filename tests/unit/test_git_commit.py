"""Unit tests for git/commit.py module

Tests for commit message formatting utilities.
"""

import pytest

from moai_adk.core.git.commit import format_commit_message


class TestFormatCommitMessage:
    """Test format_commit_message function"""

    # Korean (ko) locale tests
    def test_format_commit_message_red_ko(self):
        """Should format RED stage commit in Korean"""
        result = format_commit_message("red", "사용자 인증 테스트 작성", "ko")
        assert result == "🔴 RED: 사용자 인증 테스트 작성"
        assert result.startswith("🔴 RED:")

    def test_format_commit_message_green_ko(self):
        """Should format GREEN stage commit in Korean"""
        result = format_commit_message("green", "인증 로직 구현", "ko")
        assert result == "🟢 GREEN: 인증 로직 구현"
        assert result.startswith("🟢 GREEN:")

    def test_format_commit_message_refactor_ko(self):
        """Should format REFACTOR stage commit in Korean"""
        result = format_commit_message("refactor", "코드 구조 개선", "ko")
        assert result == "♻️ REFACTOR: 코드 구조 개선"
        assert result.startswith("♻️ REFACTOR:")

    def test_format_commit_message_docs_ko(self):
        """Should format DOCS stage commit in Korean"""
        result = format_commit_message("docs", "문서 업데이트", "ko")
        assert result == "📝 DOCS: 문서 업데이트"
        assert result.startswith("📝 DOCS:")

    # English (en) locale tests
    def test_format_commit_message_red_en(self):
        """Should format RED stage commit in English"""
        result = format_commit_message("red", "Write authentication tests", "en")
        assert result == "🔴 RED: Write authentication tests"

    def test_format_commit_message_green_en(self):
        """Should format GREEN stage commit in English"""
        result = format_commit_message("green", "Implement authentication", "en")
        assert result == "🟢 GREEN: Implement authentication"

    def test_format_commit_message_refactor_en(self):
        """Should format REFACTOR stage commit in English"""
        result = format_commit_message("refactor", "Improve code structure", "en")
        assert result == "♻️ REFACTOR: Improve code structure"

    def test_format_commit_message_docs_en(self):
        """Should format DOCS stage commit in English"""
        result = format_commit_message("docs", "Update documentation", "en")
        assert result == "📝 DOCS: Update documentation"

    # Default locale test
    def test_format_commit_message_default_locale(self):
        """Should default to Korean locale when not specified"""
        result = format_commit_message("red", "테스트 작성")
        assert result.startswith("🔴 RED:")
        assert "테스트 작성" in result

    def test_format_commit_message_unknown_locale_defaults_to_en(self):
        """Should default to English for unknown locale"""
        result = format_commit_message("red", "Test message", "unknown")
        assert result == "🔴 RED: Test message"

    # Invalid stage test
    def test_format_commit_message_invalid_stage_raises_error(self):
        """Should raise ValueError for invalid stage"""
        with pytest.raises(ValueError, match="Invalid stage"):
            format_commit_message("invalid_stage", "description", "ko")

    # Case insensitivity test
    def test_format_commit_message_case_insensitive_stage(self):
        """Should handle uppercase stage names"""
        result = format_commit_message("RED", "테스트 작성", "ko")
        assert result.startswith("🔴 RED:")

    # Japanese (ja) locale test
    def test_format_commit_message_ja(self):
        """Should format commit in Japanese locale"""
        result = format_commit_message("green", "認証ロジック実装", "ja")
        assert result.startswith("🟢 GREEN:")
        assert "認証ロジック実装" in result

    # Chinese (zh) locale test
    def test_format_commit_message_zh(self):
        """Should format commit in Chinese locale"""
        result = format_commit_message("refactor", "改进代码结构", "zh")
        assert result.startswith("♻️ REFACTOR:")
        assert "改进代码结构" in result

    # Emoji presence tests
    def test_format_commit_message_contains_emoji(self):
        """All commit messages should contain appropriate emoji"""
        stages = ["red", "green", "refactor", "docs"]
        emojis = {"red": "🔴", "green": "🟢", "refactor": "♻️", "docs": "📝"}

        for stage in stages:
            result = format_commit_message(stage, "test", "en")
            assert emojis[stage] in result
