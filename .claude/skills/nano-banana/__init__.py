"""
Nano Banana Pro Skill

Google Nano Banana Pro (Gemini 3 Pro Image Preview)를 사용한
텍스트-이미지 생성, 이미지 편집, 고급 프롬프트 처리를 제공합니다.

주요 기능:
- 텍스트-이미지 생성 (1K/2K/4K 해상도)
- 이미지-이미지 변환 (이미지 편집)
- 멀티턴 대화형 이미지 개선
- 자동 SynthID 워터마크 적용
- 한/영/일/중 다언어 프롬프트 최적화
- 포괄적인 에러 처리 및 재시도 전략

Usage:
    from modules.env_key_manager import EnvKeyManager
    from modules.image_generator import ImageGenerator
    from modules.prompt_generator import PromptGenerator

    # API 키 설정
    api_key = EnvKeyManager.get_api_key()

    # 프롬프트 최적화
    optimized_prompt = PromptGenerator.optimize("아름다운 산 풍경")

    # 이미지 생성
    generator = ImageGenerator(api_key)
    result = generator.generate_image(optimized_prompt, resolution="2048x2048")

Version: 1.0.0
Author: MoAI-ADK
License: MIT
"""

__version__ = "1.0.0"
__author__ = "MoAI-ADK"
__all__ = [
    "EnvKeyManager",
    "PromptGenerator",
    "ImageGenerator",
    "ErrorHandler",
    "FinishReasonHandler",
]

try:
    from modules.env_key_manager import EnvKeyManager
    from modules.prompt_generator import PromptGenerator
    from modules.image_generator import ImageGenerator
    from modules.error_handler import ErrorHandler, FinishReasonHandler
except ImportError as e:
    print(f"Warning: Failed to import modules: {e}")
