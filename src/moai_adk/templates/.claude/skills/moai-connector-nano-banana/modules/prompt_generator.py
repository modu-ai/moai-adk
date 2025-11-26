"""
Nano Banana Pro - 프롬프트 생성 모듈

자연어 사용자 요청을 Nano Banana Pro 최적화 프롬프트로 변환하는 모듈
"""

import logging
from typing import Dict, Optional

logger = logging.getLogger(__name__)


class PromptGenerator:
    """
    자연어 요청을 Nano Banana Pro 최적화 프롬프트로 변환

    특징:
    - 자연어 분석 (주제, 스타일, 분위기 추출)
    - 포토그래픽 요소 자동 강화 (조명, 카메라, 렌즈)
    - 색감 및 분위기 명확화
    - 구조화된 프롬프트 생성
    - 품질 검증

    Example:
        >>> generator = PromptGenerator()
        >>> prompt = generator.generate("멋진 산경 사진")
        >>> print(prompt)
        A breathtaking mountain landscape at golden hour...
    """

    # 포토그래픽 요소 템플릿
    PHOTOGRAPHIC_ELEMENTS = {
        "lighting": [
            "golden hour light",
            "soft diffuse light",
            "harsh dramatic shadows",
            "warm candlelight",
            "cool blue hour",
            "neon glow",
            "natural window light",
            "soft backlighting",
        ],
        "camera": [
            "wide shot",
            "close-up portrait",
            "overhead view",
            "low angle shot",
            "Dutch angle",
            "drone perspective",
            "macro shot",
            "telephoto compression",
        ],
        "lens": [
            "24mm wide angle",
            "35mm standard",
            "50mm standard lens",
            "85mm portrait lens",
            "135mm telephoto",
            "macro lens",
            "fisheye lens",
        ],
        "depth": [
            "shallow depth of field (f/1.8)",
            "medium depth (f/5.6)",
            "deep focus (f/16)",
            "soft bokeh background",
            "tack-sharp focus",
            "motion blur",
        ],
        "mood": [
            "serene and peaceful",
            "dramatic and intense",
            "chaotic and energetic",
            "intimate and cozy",
            "majestic and grand",
            "eerie and mysterious",
            "joyful and vibrant",
            "melancholic and thoughtful",
        ],
        "color": [
            "warm golden tones",
            "cool blue palette",
            "high contrast black and white",
            "muted pastel colors",
            "vibrant saturated colors",
            "vintage film stock colors",
            "cyberpunk neon",
            "earthy natural tones",
        ],
    }

    # 주제별 스타일 제안
    STYLE_SUGGESTIONS = {
        "nature": {
            "lighting": "golden hour light",
            "mood": "serene and peaceful",
            "camera": "wide shot",
            "depth": "deep focus (f/16)",
        },
        "portrait": {
            "lighting": "soft backlighting",
            "mood": "intimate and cozy",
            "camera": "close-up portrait",
            "lens": "85mm portrait lens",
            "depth": "shallow depth of field (f/1.8)",
        },
        "architecture": {
            "lighting": "golden hour light",
            "mood": "majestic and grand",
            "camera": "wide shot",
            "depth": "deep focus (f/16)",
        },
        "product": {
            "lighting": "soft diffuse light",
            "mood": "clean and professional",
            "camera": "close-up",
            "depth": "medium depth (f/5.6)",
        },
        "landscape": {
            "lighting": "golden hour light",
            "mood": "majestic and grand",
            "camera": "wide shot",
            "color": "warm golden tones",
        },
    }

    @staticmethod
    def generate(
        user_request: str, style: Optional[str] = None, mood: Optional[str] = None, resolution: str = "2K"
    ) -> str:
        """
        자연어 요청을 Nano Banana Pro 최적화 프롬프트로 변환

        Args:
            user_request: 사용자의 자연어 요청
            style: 선택적 스타일 지정 (portrait, landscape, product 등)
            mood: 선택적 분위기 지정 (peaceful, dramatic 등)
            resolution: 해상도 (1K, 2K, 4K)

        Returns:
            str: 최적화된 Nano Banana Pro 프롬프트

        Workflow:
            1. 요청 분석 (주제 추출)
            2. 스타일 결정 (자동 또는 명시)
            3. 포토그래픽 요소 강화
            4. 프롬프트 구조화
            5. 품질 검증

        Example:
            >>> prompt = PromptGenerator.generate("나노바나나 먹는 고양이")
            >>> print(prompt[:100])
            "A fluffy cat delicately eating a nano-banana..."
        """
        logger.info(f"Generating prompt for: {user_request}")

        # Step 1: 기본 프롬프트 생성
        scene = PromptGenerator._build_scene(user_request)

        # Step 2: 포토그래픽 요소 추가
        photographic = PromptGenerator._add_photographic_elements(user_request, style, mood)

        # Step 3: 색감 및 품질 추가
        quality = PromptGenerator._add_quality_specs(resolution)

        # Step 4: 프롬프트 조합
        final_prompt = (
            f"{scene}\n\n" f"Photographic elements:\n{photographic}\n\n" f"Quality and technical specs:\n{quality}"
        )

        logger.info("Prompt generation complete")
        return final_prompt

    @staticmethod
    def _build_scene(user_request: str) -> str:
        """
        사용자 요청으로부터 장면 설명 구성

        Args:
            user_request: 사용자 요청

        Returns:
            str: 구조화된 장면 설명
        """
        # 기본 요청을 더 자세한 서술로 확장
        request_lower = user_request.lower()

        # 주제 식별
        if any(word in request_lower for word in ["mountain", "landscape", "nature", "산", "경치"]):
            return (
                f"A breathtaking {user_request}. "
                f"The scene captures the majesty of nature with pristine clarity. "
                f"Every detail from distant peaks to foreground elements is rendered with studio-grade precision."
            )
        elif any(word in request_lower for word in ["cat", "dog", "animal", "pet", "고양이", "개", "동물"]):
            return (
                f"A beautiful {user_request}. "
                f"The subject is rendered with photorealistic detail and perfect clarity. "
                f"Every fur texture and eye detail is meticulously captured."
            )
        elif any(word in request_lower for word in ["portrait", "person", "face", "초상", "사람"]):
            return (
                f"A striking portrait of {user_request}. "
                f"The subject's features are rendered with exceptional detail and warmth. "
                f"Skin tones are natural and beautifully illuminated."
            )
        else:
            # 기본 패턴
            return (
                f"A stunning scene of {user_request}. "
                f"The composition is carefully balanced with excellent technical execution. "
                f"All elements are rendered with professional studio-grade quality."
            )

    @staticmethod
    def _add_photographic_elements(user_request: str, style: Optional[str] = None, mood: Optional[str] = None) -> str:
        """
        포토그래픽 요소 추가

        Args:
            user_request: 사용자 요청
            style: 스타일 힌트
            mood: 분위기 힌트

        Returns:
            str: 포토그래픽 요소 설명
        """
        elements = []

        # 스타일 기반 요소 추가
        if style and style in PromptGenerator.STYLE_SUGGESTIONS:
            suggestions = PromptGenerator.STYLE_SUGGESTIONS[style]
            for key, value in suggestions.items():
                elements.append(f"- {key.capitalize()}: {value}")
        else:
            # 자동 추천 (기본값)
            elements.append("- Lighting: Golden hour light from side angle")
            elements.append("- Camera: Professional composition")
            elements.append("- Lens: 50mm-85mm equivalent")
            elements.append("- Depth of field: Medium depth for clarity")

        # 분위기 추가
        if mood:
            elements.append(f"- Mood: {mood}")
        else:
            elements.append("- Mood: Professional and refined")

        return "\n".join(elements)

    @staticmethod
    def _add_quality_specs(resolution: str) -> str:
        """
        품질 및 기술 사양 추가

        Args:
            resolution: 해상도 (1K, 2K, 4K)

        Returns:
            str: 품질 사양 설명
        """
        quality_map = {
            "1K": "High quality, clear composition",
            "2K": "Excellent quality, studio-grade detail",
            "4K": "Ultra-high quality, studio-grade professional",
        }

        quality_text = quality_map.get(resolution, quality_map["2K"])

        return (
            f"- Quality: {quality_text}\n"
            f"- Resolution: {resolution}\n"
            f"- Style: Professional photography\n"
            f"- Clarity: Crystal clear, no blurriness\n"
            f"- Color accuracy: Natural and accurate\n"
            f"- Format: Aspect ratio 16:9 (landscape)\n"
            f"- Watermark: SynthID watermark included\n"
            f"- Text rendering: Clear and readable if applicable"
        )

    @staticmethod
    def display_prompt(original_request: str, generated_prompt: str, config: Optional[Dict] = None) -> None:
        """
        생성된 프롬프트 표시

        Args:
            original_request: 원래 사용자 요청
            generated_prompt: 생성된 프롬프트
            config: 추가 설정 정보
        """
        print("\n" + "=" * 70)
        print("📸 프롬프트 생성 완료")
        print("=" * 70)

        print(f"\n📝 원래 요청: {original_request}")

        print("\n🎨 최적화된 프롬프트:")
        print("-" * 70)
        print(generated_prompt)
        print("-" * 70)

        if config:
            print("\n⚙️  설정:")
            for key, value in config.items():
                print(f"   • {key}: {value}")

        print("\n✅ 이 프롬프트로 이미지를 생성할 준비가 되었습니다!")
        print("   'generate' 명령으로 이미지를 생성하세요.\n")

    @staticmethod
    def validate_prompt(prompt: str) -> Dict[str, any]:
        """
        프롬프트 품질 검증

        Args:
            prompt: 검증할 프롬프트

        Returns:
            Dict: 검증 결과

        Checks:
            - 최소 길이 (50자)
            - 최대 길이 (2000자)
            - 구조 (포토그래픽 요소 포함 여부)
            - 품질 점수 (1-10)
        """
        result = {"is_valid": True, "issues": [], "quality_score": 0, "length": len(prompt)}

        # 길이 검증
        if len(prompt) < 50:
            result["is_valid"] = False
            result["issues"].append("프롬프트가 너무 짧습니다 (최소 50자)")

        if len(prompt) > 2000:
            result["is_valid"] = False
            result["issues"].append("프롬프트가 너무 깁니다 (최대 2000자)")

        # 구조 검증
        photographic_keywords = ["lighting", "camera", "lens", "depth", "mood", "color"]
        has_photographic = any(keyword in prompt.lower() for keyword in photographic_keywords)

        if not has_photographic:
            result["issues"].append("포토그래픽 요소가 부족합니다")

        # 품질 점수 계산
        quality = 5
        if len(prompt) > 100:
            quality += 2
        if has_photographic:
            quality += 2
        if "studio" in prompt.lower() or "professional" in prompt.lower():
            quality += 1

        result["quality_score"] = min(quality, 10)

        return result


if __name__ == "__main__":
    # 테스트
    generator = PromptGenerator()

    # 예제 1: 산경
    prompt1 = generator.generate("멋진 산경", style="landscape", mood="serene and peaceful")
    generator.display_prompt("멋진 산경", prompt1, {"Resolution": "2K", "Style": "landscape", "Mood": "serene"})

    # 예제 2: 고양이
    prompt2 = generator.generate("나노바나나 먹는 고양이", style="portrait")
    generator.display_prompt("나노바나나 먹는 고양이", prompt2, {"Resolution": "2K", "Style": "portrait"})

    # 검증
    validation = PromptGenerator.validate_prompt(prompt1)
    print("\n✓ 검증 결과:", validation)
