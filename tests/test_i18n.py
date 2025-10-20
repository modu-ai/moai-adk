# @TEST:I18N-001 | SPEC: SPEC-I18N-001.md
"""
Unit tests for i18n module.

i18n 모듈 단위 테스트.
"""

import pytest
from moai_adk.i18n import (
    load_messages,
    t,
    get_supported_locales,
    get_locale_name,
    validate_locale,
    SUPPORTED_LOCALES,
    DEFAULT_LOCALE,
    FALLBACK_LOCALE,
)


class TestLoadMessages:
    """Test load_messages function."""

    def test_load_messages_ko(self):
        """한국어 메시지 로드 테스트"""
        messages = load_messages("ko")
        assert isinstance(messages, dict)
        assert messages["session_start"] == "🚀 MoAI-ADK 세션 시작"
        assert messages["language"] == "개발 언어"

    def test_load_messages_en(self):
        """영어 메시지 로드 테스트"""
        messages = load_messages("en")
        assert isinstance(messages, dict)
        assert messages["session_start"] == "🚀 MoAI-ADK Session Started"
        assert messages["language"] == "Language"

    def test_load_messages_ja(self):
        """일본어 메시지 로드 테스트"""
        messages = load_messages("ja")
        assert isinstance(messages, dict)
        assert messages["session_start"] == "🚀 MoAI-ADK セッション開始"
        assert messages["language"] == "開発言語"

    def test_load_messages_zh(self):
        """중국어 메시지 로드 테스트"""
        messages = load_messages("zh")
        assert isinstance(messages, dict)
        assert messages["session_start"] == "🚀 MoAI-ADK 会话开始"
        assert messages["language"] == "开发语言"

    def test_load_messages_th(self):
        """태국어 메시지 로드 테스트"""
        messages = load_messages("th")
        assert isinstance(messages, dict)
        assert messages["session_start"] == "🚀 MoAI-ADK เซสชันเริ่มต้น"
        assert messages["language"] == "ภาษาการพัฒนา"

    def test_load_messages_unsupported_locale(self):
        """지원하지 않는 locale은 영어로 대체"""
        messages = load_messages("fr")  # French (not supported)
        assert messages["session_start"] == "🚀 MoAI-ADK Session Started"

    def test_load_messages_caching(self):
        """LRU 캐시 동작 확인"""
        # First call
        messages1 = load_messages("ko")
        # Second call (should be cached)
        messages2 = load_messages("ko")
        # Should be the same object (cached)
        assert messages1 is messages2


class TestTranslate:
    """Test t() translation function."""

    def test_translate_simple_key(self):
        """단순 키 번역 테스트"""
        assert t("session_start", "ko") == "🚀 MoAI-ADK 세션 시작"
        assert t("session_start", "en") == "🚀 MoAI-ADK Session Started"

    def test_translate_nested_key(self):
        """중첩 키 번역 테스트"""
        assert t("error.no_git", "ko") == "❌ Git 리포지토리가 아닙니다"
        assert t("error.no_git", "en") == "❌ Not a Git repository"

    def test_translate_with_format_single_variable(self):
        """단일 변수 포맷팅 테스트"""
        result_ko = t("checkpoint_created", "ko", name="before-merge")
        assert "before-merge" in result_ko
        assert "체크포인트" in result_ko

        result_en = t("checkpoint_created", "en", name="before-merge")
        assert "before-merge" in result_en
        assert "Checkpoint" in result_en

    def test_translate_with_format_multiple_variables(self):
        """복수 변수 포맷팅 테스트"""
        result_ko = t("context_loaded", "ko", count=3)
        assert "3" in result_ko

        result_en = t("context_loaded", "en", count=5)
        assert "5" in result_en

    def test_translate_missing_key(self):
        """존재하지 않는 키는 키 자체를 반환"""
        assert t("nonexistent_key", "ko") == "nonexistent_key"
        assert t("error.nonexistent", "en") == "error.nonexistent"

    def test_translate_invalid_format_variables(self):
        """잘못된 포맷 변수는 무시됨"""
        # Missing variable (should return unformatted message)
        result = t("checkpoint_created", "ko")
        assert "{name}" in result

    def test_translate_default_locale(self):
        """기본 locale (ko) 테스트"""
        result = t("session_start")  # No locale specified
        assert "세션" in result


class TestGetSupportedLocales:
    """Test get_supported_locales function."""

    def test_returns_list(self):
        """리스트 반환 확인"""
        locales = get_supported_locales()
        assert isinstance(locales, list)

    def test_contains_five_locales(self):
        """5개 언어 포함 확인"""
        locales = get_supported_locales()
        assert len(locales) == 5

    def test_contains_expected_locales(self):
        """예상 언어 포함 확인"""
        locales = get_supported_locales()
        assert set(locales) == {"ko", "en", "ja", "zh", "th"}

    def test_returns_copy(self):
        """복사본 반환 확인 (원본 수정 방지)"""
        locales1 = get_supported_locales()
        locales2 = get_supported_locales()
        assert locales1 is not locales2  # Different objects


class TestGetLocaleName:
    """Test get_locale_name function."""

    def test_get_locale_name_ko(self):
        """한국어 이름 확인"""
        assert get_locale_name("ko") == "Korean (한국어)"

    def test_get_locale_name_en(self):
        """영어 이름 확인"""
        assert get_locale_name("en") == "English"

    def test_get_locale_name_ja(self):
        """일본어 이름 확인"""
        assert get_locale_name("ja") == "Japanese (日本語)"

    def test_get_locale_name_zh(self):
        """중국어 이름 확인"""
        assert get_locale_name("zh") == "Chinese (中文)"

    def test_get_locale_name_th(self):
        """태국어 이름 확인"""
        assert get_locale_name("th") == "Thai (ไทย)"

    def test_get_locale_name_unknown(self):
        """알 수 없는 locale은 그대로 반환"""
        assert get_locale_name("fr") == "fr"


class TestValidateLocale:
    """Test validate_locale function."""

    def test_validate_supported_locale(self):
        """지원되는 locale 검증"""
        assert validate_locale("ko") == "ko"
        assert validate_locale("en") == "en"
        assert validate_locale("ja") == "ja"

    def test_validate_uppercase_locale(self):
        """대문자 locale은 소문자로 변환"""
        assert validate_locale("EN") == "en"
        assert validate_locale("Ko") == "ko"

    def test_validate_whitespace_locale(self):
        """공백 제거"""
        assert validate_locale(" ko ") == "ko"
        assert validate_locale("en  ") == "en"

    def test_validate_unsupported_locale(self):
        """지원하지 않는 locale은 영어로 대체"""
        assert validate_locale("fr") == "en"
        assert validate_locale("de") == "en"

    def test_validate_empty_locale(self):
        """빈 locale은 영어로 대체"""
        assert validate_locale("") == "en"


class TestConstants:
    """Test module constants."""

    def test_supported_locales_constant(self):
        """SUPPORTED_LOCALES 상수 확인"""
        assert SUPPORTED_LOCALES == ["ko", "en", "ja", "zh", "th"]

    def test_default_locale_constant(self):
        """DEFAULT_LOCALE 상수 확인"""
        assert DEFAULT_LOCALE == "ko"

    def test_fallback_locale_constant(self):
        """FALLBACK_LOCALE 상수 확인"""
        assert FALLBACK_LOCALE == "en"
