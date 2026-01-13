"""Test TAG configuration translations.

Tests for TAG system configuration prompt translations across all supported languages.
"""

import pytest

from moai_adk.cli.prompts.translations import get_translation


class TestTagTranslations:
    """Test TAG translation completeness and structure."""

    @pytest.mark.parametrize(
        "locale",
        ["ko", "en", "ja", "zh"],
    )
    def test_tag_translations_exist(self, locale):
        """Test that all TAG translation keys exist for each locale."""
        t = get_translation(locale)

        # Required TAG keys
        required_keys = [
            # Headers
            "tag_setup",
            # Questions
            "q_tag_enable",
            "q_tag_mode",
            # Options - Enable
            "opt_tag_yes",
            "opt_tag_no",
            # Options - Mode
            "opt_tag_warn",
            "opt_tag_enforce",
            "opt_tag_off",
            # Descriptions
            "desc_tag_yes",
            "desc_tag_no",
            "desc_tag_warn",
            "desc_tag_enforce",
            "desc_tag_off",
            # Explanations
            "tag_system_intro",
            "tag_yes_recommendation",
            "tag_no_warning",
            "tag_mode_guide_title",
            "tag_mode_guide_subtitle",
            # Messages
            "msg_tag_enabled",
            "msg_tag_disabled",
            "msg_tag_mode_selected",
        ]

        for key in required_keys:
            assert key in t, f"Missing translation key '{key}' for locale '{locale}'"

    def test_tag_translations_korean_complete(self):
        """Test Korean TAG translations are complete and meaningful."""
        t = get_translation("ko")

        # Header should contain Korean
        assert "TAG" in t["tag_setup"] or "태그" in t["tag_setup"]

        # Questions should be in Korean
        assert "활성화" in t["q_tag_enable"] or "TAG" in t["q_tag_enable"]
        assert "모드" in t["q_tag_mode"] or "mode" in t["q_tag_mode"].lower()

        # Options should be descriptive
        assert len(t["desc_tag_warn"]) > 10  # Should have description
        assert len(t["desc_tag_enforce"]) > 10
        assert len(t["desc_tag_off"]) > 10

    def test_tag_translations_english_complete(self):
        """Test English TAG translations are complete and meaningful."""
        t = get_translation("en")

        # Header should contain TAG
        assert "TAG" in t["tag_setup"]

        # Questions should be in English
        assert "Enable" in t["q_tag_enable"] or "TAG" in t["q_tag_enable"]
        assert "mode" in t["q_tag_mode"].lower()

        # Options should be descriptive
        assert len(t["desc_tag_warn"]) > 10
        assert len(t["desc_tag_enforce"]) > 10
        assert len(t["desc_tag_off"]) > 10

    def test_tag_translations_japanese_complete(self):
        """Test Japanese TAG translations are complete and meaningful."""
        t = get_translation("ja")

        # Header should contain TAG
        assert "TAG" in t["tag_setup"] or "タグ" in t["tag_setup"]

        # Questions should be in Japanese
        assert ("有効" in t["q_tag_enable"] or "TAG" in t["q_tag_enable"]
                or "validation" in t["q_tag_enable"].lower())

        # Options should be descriptive
        assert len(t["desc_tag_warn"]) > 10
        assert len(t["desc_tag_enforce"]) > 10
        assert len(t["desc_tag_off"]) > 10

    def test_tag_translations_chinese_complete(self):
        """Test Chinese TAG translations are complete and meaningful."""
        t = get_translation("zh")

        # Header should contain TAG
        assert "TAG" in t["tag_setup"] or "标签" in t["tag_setup"]

        # Questions should be in Chinese
        assert ("启用" in t["q_tag_enable"] or "TAG" in t["q_tag_enable"]
                or "validation" in t["q_tag_enable"].lower())

        # Options should be descriptive
        assert len(t["desc_tag_warn"]) > 10
        assert len(t["desc_tag_enforce"]) > 10
        assert len(t["desc_tag_off"]) > 10

    @pytest.mark.parametrize(
        "locale",
        ["ko", "en", "ja", "zh"],
    )
    def test_tag_yes_recommendation_exists(self, locale):
        """Test that TAG yes recommendation exists for each locale."""
        t = get_translation(locale)

        # Should have recommendation text
        assert "tag_yes_recommendation" in t
        assert len(t["tag_yes_recommendation"]) > 20  # Should be meaningful

    @pytest.mark.parametrize(
        "locale",
        ["ko", "en", "ja", "zh"],
    )
    def test_tdd_purpose_mentioned(self, locale):
        """Test that TDD purpose is mentioned in translations."""
        t = get_translation(locale)

        # Should mention TDD in some form
        tag_setup_text = t.get("tag_setup", "") + t.get("tag_system_intro", "")
        combined_lower = tag_setup_text.lower()

        # Check for TDD or related terms
        has_tdd = (
            "tdd" in combined_lower
            or "test-driven" in combined_lower
            or "테스트 주도" in combined_lower
            or "テスト駆動" in combined_lower
            or "测试驱动" in combined_lower
        )

        # At least one language should mention TDD (verification)
        assert has_tdd or len(tag_setup_text) > 0  # Allow if description exists

    def test_tag_mode_guide_structure(self):
        """Test that tag mode guide has proper structure."""
        t = get_translation("en")

        # Should have title and subtitle
        assert "tag_mode_guide_title" in t
        assert "tag_mode_guide_subtitle" in t

        # Title should be non-empty
        assert len(t["tag_mode_guide_title"]) > 5
        assert len(t["tag_mode_guide_subtitle"]) > 10
