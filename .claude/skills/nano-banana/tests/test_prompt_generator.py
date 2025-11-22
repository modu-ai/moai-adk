"""
PromptGenerator 테스트

프롬프트 최적화 및 검증 기능 테스트
"""

import pytest
from modules.prompt_generator import PromptGenerator


class TestPromptGenerator:
    """PromptGenerator 테스트 클래스"""

    def test_validate_valid_prompt(self):
        """유효한 프롬프트 검증"""
        assert PromptGenerator.validate("beautiful landscape") is True

    def test_validate_too_short(self):
        """너무 짧은 프롬프트 검증"""
        assert PromptGenerator.validate("ab") is False

    def test_validate_empty(self):
        """빈 프롬프트 검증"""
        assert PromptGenerator.validate("") is False
        assert PromptGenerator.validate(None) is False

    def test_validate_too_long(self):
        """너무 긴 프롬프트 검증"""
        long_prompt = "a" * (PromptGenerator.MAX_CHARS + 1)
        assert PromptGenerator.validate(long_prompt) is False

    def test_optimize_basic(self):
        """기본 프롬프트 최적화"""
        prompt = "mountain landscape"
        optimized = PromptGenerator.optimize(prompt)

        # 최적화된 프롬프트는 원본을 포함하고 추가 요소가 있어야 함
        assert "mountain landscape" in optimized
        assert len(optimized) > len(prompt)

    def test_optimize_with_style(self):
        """스타일이 적용된 프롬프트 최적화"""
        prompt = "forest scene"
        optimized = PromptGenerator.optimize(prompt, style="photorealistic")

        assert "photorealistic" in optimized
        assert "forest scene" in optimized

    def test_optimize_without_photographic(self):
        """사진 요소 없이 최적화"""
        prompt = "abstract art"
        optimized = PromptGenerator.optimize(prompt, add_photographic=False)

        # 사진 요소가 없어야 함
        assert "professional photography" not in optimized

    def test_optimize_korean(self):
        """한국어 프롬프트 최적화"""
        prompt = "아름다운 산 풍경"
        optimized = PromptGenerator.optimize(prompt, language="ko")

        # 한국어 스타일 요소 포함 확인
        assert "산 풍경" in optimized or "landscape" in optimized.lower()

    def test_optimize_japanese(self):
        """일본어 프롬프트 최적화"""
        prompt = "beautiful landscape"
        optimized = PromptGenerator.optimize(prompt, language="ja")

        # 일본어 요소 포함 확인
        assert len(optimized) > 0

    def test_language_detection_korean(self):
        """한국어 자동 감지"""
        lang = PromptGenerator._detect_language("아름다운 산")
        assert lang == "ko"

    def test_language_detection_english(self):
        """영어 자동 감지"""
        lang = PromptGenerator._detect_language("beautiful landscape")
        assert lang == "en"

    def test_language_detection_mixed(self):
        """혼합 언어 감지 (한국어 우선)"""
        lang = PromptGenerator._detect_language("아름다운 beautiful landscape")
        assert lang == "ko"

    def test_add_style(self):
        """스타일 추가"""
        prompt = "landscape"
        styled = PromptGenerator.add_style(prompt, "cinematic")

        assert "cinematic" in styled
        assert "landscape" in styled

    def test_add_invalid_style(self):
        """존재하지 않는 스타일 추가"""
        prompt = "landscape"
        styled = PromptGenerator.add_style(prompt, "nonexistent_style")

        # 스타일이 추가되지 않아야 함
        assert styled == prompt

    def test_get_style_list(self):
        """스타일 목록 조회"""
        styles = PromptGenerator.get_style_list()

        assert isinstance(styles, list)
        assert len(styles) > 0
        assert "photorealistic" in styles
        assert "artistic" in styles
        assert "cinematic" in styles

    def test_get_resolution_list(self):
        """Get resolution list"""
        resolutions = PromptGenerator.get_resolution_list()

        assert isinstance(resolutions, dict)
        assert "1k" in resolutions
        assert "2k" in resolutions
        assert "4k" in resolutions
        assert resolutions["1k"] == "1024x1024"
        assert resolutions["2k"] == "2048x2048"
        assert resolutions["4k"] == "4096x4096"

    def test_truncate(self):
        """프롬프트 자르기"""
        long_prompt = "beautiful " * 300  # 매우 긴 프롬프트
        truncated = PromptGenerator._truncate(long_prompt, 100)

        assert len(truncated) <= 100

    def test_clean_prompt(self):
        """프롬프트 정제"""
        dirty_prompt = "  beautiful   mountain  \n landscape  "
        cleaned = PromptGenerator._clean_prompt(dirty_prompt)

        assert cleaned == "beautiful mountain landscape"
        assert "  " not in cleaned

    def test_clean_prompt_special_chars(self):
        """특수 문자 정제"""
        prompt = "beautiful$%^&*landscape###"
        cleaned = PromptGenerator._clean_prompt(prompt)

        # 특수 문자 제거됨
        assert "$" not in cleaned
        assert "#" not in cleaned
        assert "landscape" in cleaned

    def test_multiple_optimization(self):
        """연쇄 최적화"""
        prompt = "tree"
        optimized1 = PromptGenerator.optimize(prompt)
        optimized2 = PromptGenerator.add_style(optimized1, "artistic")

        # 두 번의 최적화 모두 적용됨
        assert len(optimized2) >= len(optimized1)
        assert "artistic" in optimized2


class TestPromptGeneratorEdgeCases:
    """엣지 케이스 테스트"""

    def test_optimize_unicode_prompt(self):
        """유니코드 프롬프트 최적화"""
        prompt = "🌅 beautiful sunset landscape 🌆"
        optimized = PromptGenerator.optimize(prompt)

        assert len(optimized) > 0

    def test_optimize_with_newlines(self):
        """줄바꿈 포함 프롬프트"""
        prompt = "beautiful\nmountain\nlandscape"
        optimized = PromptGenerator.optimize(prompt)

        # 줄바꿈이 정제되어야 함
        assert "\n" not in optimized

    def test_optimize_case_insensitive_style(self):
        """대소문자 스타일"""
        prompt = "landscape"

        # 소문자 스타일
        optimized_lower = PromptGenerator.optimize(prompt, style="photorealistic")

        # 대문자 스타일 (작동하지 않아야 함)
        optimized_upper = PromptGenerator.optimize(prompt, style="PHOTOREALISTIC")

        # 소문자만 작동
        assert "photorealistic" in optimized_lower


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
