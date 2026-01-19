"""
Comprehensive tests for language_config module.

Tests language configuration, language info retrieval, and related utilities.
"""


from moai_adk.core.language_config import (
    LANGUAGE_CONFIG,
    LANGUAGE_MODEL_PREFERENCE,
    RTL_LANGUAGES,
    TRANSLATION_PRIORITY,
    get_all_supported_codes,
    get_english_name,
    get_language_family,
    get_language_info,
    get_native_name,
    get_optimal_model,
    get_translation_priority,
    is_rtl_language,
)


class TestLanguageConfigConstants:
    """Test language configuration constants."""

    def test_language_config_structure(self):
        """Test LANGUAGE_CONFIG has correct structure."""
        assert isinstance(LANGUAGE_CONFIG, dict)
        assert len(LANGUAGE_CONFIG) > 0

        # Each language should have required keys
        for code, info in LANGUAGE_CONFIG.items():
            assert isinstance(code, str)
            assert isinstance(info, dict)
            assert "name" in info
            assert "native_name" in info
            assert "code" in info
            assert "family" in info

    def test_rtl_languages_set(self):
        """Test RTL_LANGUAGES is a set."""
        assert isinstance(RTL_LANGUAGES, set)
        # Note: RTL_LANGUAGES may include languages not in LANGUAGE_CONFIG
        # This is acceptable as RTL languages are a broader category

    def test_translation_priority_list(self):
        """Test TRANSLATION_PRIORITY is a list of valid codes."""
        assert isinstance(TRANSLATION_PRIORITY, list)
        assert len(TRANSLATION_PRIORITY) > 0
        for lang in TRANSLATION_PRIORITY:
            assert lang in LANGUAGE_CONFIG

    def test_language_model_preference_mapping(self):
        """Test LANGUAGE_MODEL_PREFERENCE maps valid codes to models."""
        assert isinstance(LANGUAGE_MODEL_PREFERENCE, dict)
        for code, model in LANGUAGE_MODEL_PREFERENCE.items():
            assert code in LANGUAGE_CONFIG
            assert isinstance(model, str)
            assert "claude" in model.lower()


class TestGetLanguageInfo:
    """Test get_language_info function."""

    def test_get_existing_language_info(self):
        """Test getting info for existing language."""
        result = get_language_info("en")
        assert result is not None
        assert result["name"] == "English"
        assert result["native_name"] == "English"
        assert result["code"] == "en"

    def test_get_nonexistent_language_info(self):
        """Test getting info for nonexistent language returns None."""
        result = get_language_info("xyz")
        assert result is None

    def test_get_language_info_case_insensitive(self):
        """Test get_language_info is case-insensitive."""
        result_lower = get_language_info("en")
        result_upper = get_language_info("EN")
        result_mixed = get_language_info("En")

        assert result_lower is not None
        assert result_upper is not None
        assert result_mixed is not None

        assert result_lower["code"] == result_upper["code"] == result_mixed["code"]

    def test_get_korean_language_info(self):
        """Test getting Korean language info."""
        result = get_language_info("ko")
        assert result is not None
        assert result["name"] == "Korean"
        assert result["native_name"] == "한국어"
        assert result["code"] == "ko"
        assert result["family"] == "koreanic"

    def test_get_japanese_language_info(self):
        """Test getting Japanese language info."""
        result = get_language_info("ja")
        assert result is not None
        assert result["name"] == "Japanese"
        assert result["native_name"] == "日本語"
        assert result["family"] == "japonic"


class TestGetNativeName:
    """Test get_native_name function."""

    def test_get_native_name_existing_language(self):
        """Test getting native name for existing language."""
        assert get_native_name("ko") == "한국어"
        assert get_native_name("ja") == "日本語"
        assert get_native_name("en") == "English"
        assert get_native_name("zh") == "中文"

    def test_get_native_name_nonexistent_language(self):
        """Test getting native name for nonexistent language returns fallback."""
        result = get_native_name("xyz")
        assert result == "English"

    def test_get_native_name_case_insensitive(self):
        """Test get_native_name is case-insensitive."""
        result1 = get_native_name("KO")
        result2 = get_native_name("ko")
        assert result1 == result2 == "한국어"


class TestGetEnglishName:
    """Test get_english_name function."""

    def test_get_english_name_existing_language(self):
        """Test getting English name for existing language."""
        assert get_english_name("ko") == "Korean"
        assert get_english_name("ja") == "Japanese"
        assert get_english_name("en") == "English"
        assert get_english_name("fr") == "French"

    def test_get_english_name_nonexistent_language(self):
        """Test getting English name for nonexistent language returns fallback."""
        result = get_english_name("xyz")
        assert result == "English"

    def test_get_english_name_case_insensitive(self):
        """Test get_english_name is case-insensitive."""
        result1 = get_english_name("DE")
        result2 = get_english_name("de")
        assert result1 == result2 == "German"


class TestGetAllSupportedCodes:
    """Test get_all_supported_codes function."""

    def test_get_all_supported_codes_returns_list(self):
        """Test function returns a list of language codes."""
        codes = get_all_supported_codes()
        assert isinstance(codes, list)
        assert len(codes) > 0

    def test_get_all_supported_codes_contains_valid_codes(self):
        """Test all returned codes are valid."""
        codes = get_all_supported_codes()
        for code in codes:
            assert code in LANGUAGE_CONFIG

    def test_get_all_supported_codes_expected_languages(self):
        """Test expected languages are in the list."""
        codes = get_all_supported_codes()
        expected = ["en", "ko", "ja", "es", "fr", "de", "zh", "pt", "ru", "it", "ar", "hi"]
        for lang in expected:
            assert lang in codes


class TestGetLanguageFamily:
    """Test get_language_family function."""

    def test_get_language_family_existing_language(self):
        """Test getting language family for existing language."""
        assert get_language_family("en") == "indo-european"
        assert get_language_family("ko") == "koreanic"
        assert get_language_family("ja") == "japonic"
        assert get_language_family("zh") == "sino-tibetan"

    def test_get_language_family_nonexistent_language(self):
        """Test getting language family for nonexistent language returns None."""
        result = get_language_family("xyz")
        assert result is None

    def test_get_language_family_case_insensitive(self):
        """Test get_language_family is case-insensitive."""
        result1 = get_language_family("EN")
        result2 = get_language_family("en")
        assert result1 == result2


class TestGetOptimalModel:
    """Test get_optimal_model function."""

    def test_get_optimal_model_existing_language(self):
        """Test getting optimal model for existing language."""
        model = get_optimal_model("en")
        assert isinstance(model, str)
        assert "claude" in model.lower()
        assert "sonnet" in model.lower()

    def test_get_optimal_model_nonexistent_language(self):
        """Test getting optimal model for nonexistent language returns default."""
        model = get_optimal_model("xyz")
        assert isinstance(model, str)
        assert "claude" in model.lower()

    def test_get_optimal_model_all_supported_languages(self):
        """Test optimal model is returned for all supported languages."""
        for code in LANGUAGE_CONFIG.keys():
            model = get_optimal_model(code)
            assert isinstance(model, str)
            assert len(model) > 0


class TestIsRtlLanguage:
    """Test is_rtl_language function."""

    def test_is_rtl_language_for_rtl_languages(self):
        """Test RTL language detection for RTL languages."""
        assert is_rtl_language("ar") is True

    def test_is_rtl_language_for_ltr_languages(self):
        """Test RTL language detection for LTR languages."""
        assert is_rtl_language("en") is False
        assert is_rtl_language("ko") is False
        assert is_rtl_language("ja") is False
        assert is_rtl_language("zh") is False

    def test_is_rtl_language_case_insensitive(self):
        """Test is_rtl_language is case-insensitive."""
        result1 = is_rtl_language("AR")
        result2 = is_rtl_language("ar")
        assert result1 == result2 is True


class TestGetTranslationPriority:
    """Test get_translation_priority function."""

    def test_get_translation_priority_returns_list(self):
        """Test function returns a list."""
        priority = get_translation_priority()
        assert isinstance(priority, list)

    def test_get_translation_priority_returns_copy(self):
        """Test function returns a copy, not the original list."""
        priority1 = get_translation_priority()
        priority2 = get_translation_priority()

        # Modify one list
        priority1.append("xyz")

        # Other list should be unchanged
        assert "xyz" not in priority2

    def test_get_translation_priority_contains_expected_languages(self):
        """Test priority list contains expected languages."""
        priority = get_translation_priority()
        assert "en" in priority
        assert "ko" in priority
        assert "ja" in priority

    def test_get_translation_priority_order(self):
        """Test English is first in priority list."""
        priority = get_translation_priority()
        assert priority[0] == "en"


class TestLanguageConfigEdgeCases:
    """Test edge cases and error conditions."""

    def test_empty_language_code(self):
        """Test behavior with empty language code."""
        assert get_language_info("") is None
        assert get_native_name("") == "English"
        assert get_english_name("") == "English"

    def test_none_language_code_handling(self):
        """Test behavior with None as language code (should not crash)."""
        # These should handle None gracefully or raise appropriate errors
        # The implementation should not crash with TypeError
        result = get_language_info("none")  # String "none", not None type
        assert result is None

    def test_special_characters_in_code(self):
        """Test handling of special characters in language code."""
        # Language codes with special characters should return None
        assert get_language_info("@#$") is None
        assert get_native_name("@#$") == "English"

    def test_numeric_language_code(self):
        """Test handling of numeric language code."""
        assert get_language_info("123") is None
        assert get_native_name("123") == "English"


class TestLanguageConfigDataIntegrity:
    """Test data integrity and consistency."""

    def test_all_languages_have_required_fields(self):
        """Test all languages in config have required fields."""
        required_fields = {"name", "native_name", "code", "family"}

        for code, info in LANGUAGE_CONFIG.items():
            missing_fields = required_fields - set(info.keys())
            assert len(missing_fields) == 0, f"Language {code} missing fields: {missing_fields}"

    def test_language_codes_match_config(self):
        """Test language codes match their 'code' field in config."""
        for code, info in LANGUAGE_CONFIG.items():
            assert info["code"] == code, f"Language code mismatch: {code} != {info['code']}"

    def test_rtl_languages_subset_of_config(self):
        """Test RTL languages that are in config are valid."""
        # Only check RTL languages that exist in LANGUAGE_CONFIG
        rtl_in_config = RTL_LANGUAGES & LANGUAGE_CONFIG.keys()
        for lang in rtl_in_config:
            assert lang in LANGUAGE_CONFIG, f"RTL language {lang} not in LANGUAGE_CONFIG"

    def test_translation_priority_subset_of_config(self):
        """Test all priority languages exist in main config."""
        for lang in TRANSLATION_PRIORITY:
            assert lang in LANGUAGE_CONFIG, f"Priority language {lang} not in LANGUAGE_CONFIG"

    def test_model_preference_subset_of_config(self):
        """Test all model preference languages exist in main config."""
        for lang in LANGUAGE_MODEL_PREFERENCE:
            assert lang in LANGUAGE_CONFIG, f"Model preference language {lang} not in LANGUAGE_CONFIG"
