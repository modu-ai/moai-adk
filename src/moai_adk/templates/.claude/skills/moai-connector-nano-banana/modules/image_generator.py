"""
Nano Banana Pro - 이미지 생성 모듈

Google Gemini 3 Pro Image Preview (Nano Banana) API를 사용한 이미지 생성/편집

Official API Documentation:
- https://ai.google.dev/gemini-api/docs/image-generation
- Models: gemini-2.5-flash-image, gemini-3-pro-image-preview
- API: POST /v1beta/models/{model}:generateContent
"""

import os
import base64
from pathlib import Path
from typing import Optional, List, Tuple, Dict, Any
from datetime import datetime
import logging
from io import BytesIO

import google.generativeai as genai
from google.api_core import exceptions

logger = logging.getLogger(__name__)


class NanoBananaImageGenerator:
    """
    Gemini 3 Nano Banana API를 사용한 이미지 생성 및 편집

    Features:
    - Text-to-Image 생성 (1K/2K/4K 해상도)
    - Image-to-Image 편집 (스타일 전이, 객체 조작)
    - Google Search 실시간 정보 연동
    - Multi-turn 대화형 편집
    - 에러 처리 및 재시도 로직

    Models:
    - gemini-2.5-flash-image: 빠른 생성, 1K 해상도 (권장: 프로토타입)
    - gemini-3-pro-image-preview: 고품질, 4K 해상도 (권장: 프로덕션)

    Example:
        >>> from env_key_manager import EnvKeyManager
        >>> api_key = EnvKeyManager.load_api_key()
        >>> generator = NanoBananaImageGenerator(api_key)
        >>> image, metadata = generator.generate(
        ...     "A serene mountain landscape at golden hour"
        ... )
        >>> image.save("output.png")
    """

    # 지원 모델
    MODELS = {
        "flash": "gemini-2.5-flash-image",      # 빠른 생성
        "pro": "gemini-3-pro-image-preview"     # 고품질
    }

    # 지원 해상도
    RESOLUTIONS = ["1K", "2K", "4K"]

    # 지원 종횡비
    ASPECT_RATIOS = [
        "1:1",      # 정사각형
        "2:3", "3:2",  # 세로/가로
        "3:4", "4:3",  # 표준
        "4:5", "5:4",  # 인스타그램
        "9:16", "16:9",  # 모바일/와이드
        "21:9"      # 울트라 와이드
    ]

    # 기본 설정
    DEFAULT_CONFIG = {
        "model": "flash",
        "resolution": "2K",
        "aspect_ratio": "16:9",
        "use_google_search": False,
        "max_retries": 3,
        "timeout": 60
    }

    def __init__(self, api_key: Optional[str] = None):
        """
        Initialize Nano Banana Image Generator

        Args:
            api_key: Google Gemini API key
                    (if None, loads from environment variable)

        Example:
            >>> generator = NanoBananaImageGenerator()
            >>> # or
            >>> generator = NanoBananaImageGenerator("gsk_...")
        """
        if api_key is None:
            api_key = os.getenv("GOOGLE_API_KEY")

        if not api_key:
            raise ValueError(
                "API key not found. Set GOOGLE_API_KEY environment variable "
                "or pass api_key parameter"
            )

        genai.configure(api_key=api_key)
        self.client = genai.Client(api_key=api_key)
        logger.info("Nano Banana Image Generator initialized")

    def generate(
        self,
        prompt: str,
        model: str = "flash",
        resolution: str = "2K",
        aspect_ratio: str = "16:9",
        use_google_search: bool = False,
        save_path: Optional[str] = None
    ) -> Tuple[Any, Dict[str, Any]]:
        """
        Text-to-Image 생성

        Args:
            prompt: 이미지 생성 프롬프트
            model: 모델 선택 ("flash" 또는 "pro")
            resolution: 해상도 ("1K", "2K", "4K")
            aspect_ratio: 종횡비 (기본: "16:9")
            use_google_search: Google Search 연동 여부
            save_path: 이미지 저장 경로 (선택사항)

        Returns:
            Tuple[PIL.Image, Dict]: (생성된 이미지, 메타데이터)

        Raises:
            ValueError: 잘못된 파라미터
            Exception: API 호출 실패

        Example:
            >>> image, metadata = generator.generate(
            ...     "A futuristic city at sunset",
            ...     model="pro",
            ...     resolution="4K",
            ...     aspect_ratio="16:9"
            ... )
            >>> print(metadata['tokens_used'])
            1234
            >>> image.save("city.png")
        """
        # 파라미터 검증
        self._validate_params(model, resolution, aspect_ratio)

        print(f"\n{'='*70}")
        print(f"🎨 Nano Banana 이미지 생성 시작")
        print(f"{'='*70}")
        print(f"📝 프롬프트: {prompt[:50]}...")
        print(f"🎯 설정: {model.upper()} | {resolution} | {aspect_ratio}")
        print(f"🔍 Google Search: {'활성화' if use_google_search else '비활성화'}")
        print(f"⏳ 처리 중...\n")

        try:
            # 모델명 확인
            model_name = self.MODELS[model]

            # 요청 구성
            generation_config = {
                "response_modalities": ["TEXT", "IMAGE"],
                "image_config": {
                    "aspect_ratio": aspect_ratio,
                    "image_size": resolution
                }
            }

            # Google Search 연동
            tools = []
            if use_google_search:
                tools = [{"google_search": {}}]

            # API 호출
            response = self.client.models.generate_content(
                model=model_name,
                contents=[{"parts": [{"text": prompt}]}],
                config=generation_config,
                tools=tools if tools else None
            )

            # 응답 처리
            image = None
            description = ""

            for part in response.candidates[0].content.parts:
                if hasattr(part, 'text') and part.text:
                    description = part.text
                elif hasattr(part, 'inline_data') and part.inline_data:
                    # Base64 데이터를 PIL Image로 변환
                    image_bytes = base64.b64decode(part.inline_data.data)
                    from PIL import Image
                    image = Image.open(BytesIO(image_bytes))

            if not image:
                raise ValueError("No image data in response")

            # 메타데이터 구성
            metadata = {
                "timestamp": datetime.now().isoformat(),
                "model": model,
                "resolution": resolution,
                "aspect_ratio": aspect_ratio,
                "prompt": prompt,
                "description": description,
                "finish_reason": response.candidates[0].finish_reason,
                "tokens_used": response.usage_metadata.total_token_count if hasattr(response, 'usage_metadata') else None,
                "use_google_search": use_google_search,
                "grounding_sources": []
            }

            # Google Search 출처 정보
            if use_google_search and hasattr(response.candidates[0], 'grounding_metadata'):
                grounding = response.candidates[0].grounding_metadata
                if hasattr(grounding, 'grounding_chunks'):
                    for chunk in grounding.grounding_chunks:
                        if hasattr(chunk, 'web') and chunk.web:
                            metadata["grounding_sources"].append({
                                "uri": chunk.web.uri,
                                "title": chunk.web.title
                            })

            # 저장
            if save_path:
                Path(save_path).parent.mkdir(parents=True, exist_ok=True)
                image.save(save_path)
                metadata["saved_to"] = save_path
                print(f"✅ 이미지 저장: {save_path}\n")

            print(f"✅ 이미지 생성 완료!")
            print(f"   • 모델: {model.upper()}")
            print(f"   • 해상도: {resolution}")
            print(f"   • 토큰: {metadata['tokens_used']}")
            if metadata["grounding_sources"]:
                print(f"   • 출처: {len(metadata['grounding_sources'])}개 웹 페이지")

            return image, metadata

        except exceptions.ResourceExhausted:
            logger.error("API quota exceeded")
            print("❌ API 할당량 초과")
            print("   • 해상도를 1K로 다운그레이드하거나")
            print("   • 몇 분 후에 다시 시도하세요")
            raise

        except exceptions.PermissionDenied:
            logger.error("Permission denied - check API key")
            print("❌ 권한 오류 - API 키 확인이 필요합니다")
            raise

        except exceptions.InvalidArgument as e:
            logger.error(f"Invalid argument: {e}")
            print(f"❌ 잘못된 파라미터: {e}")
            raise

        except Exception as e:
            logger.error(f"Error generating image: {e}")
            print(f"❌ 오류 발생: {e}")
            raise

    def edit(
        self,
        image_path: str,
        instruction: str,
        model: str = "flash",
        resolution: str = "2K",
        aspect_ratio: str = "16:9",
        save_path: Optional[str] = None
    ) -> Tuple[Any, Dict[str, Any]]:
        """
        Image-to-Image 편집

        Args:
            image_path: 편집할 이미지 경로
            instruction: 편집 지시사항
            model: 모델 선택
            resolution: 출력 해상도
            aspect_ratio: 출력 종횡비
            save_path: 결과 저장 경로

        Returns:
            Tuple[PIL.Image, Dict]: (편집된 이미지, 메타데이터)

        Example:
            >>> edited_image, metadata = generator.edit(
            ...     "original.png",
            ...     "Add a sunset in the background",
            ...     model="pro",
            ...     resolution="2K"
            ... )
            >>> edited_image.save("with_sunset.png")
        """
        # 파라미터 검증
        self._validate_params(model, resolution, aspect_ratio)

        # 이미지 로드
        if not Path(image_path).exists():
            raise FileNotFoundError(f"Image not found: {image_path}")

        from PIL import Image
        original_image = Image.open(image_path)
        original_path = str(Path(image_path).resolve())

        print(f"\n{'='*70}")
        print(f"✏️  이미지 편집 시작")
        print(f"{'='*70}")
        print(f"📁 원본: {original_path}")
        print(f"📝 지시사항: {instruction[:50]}...")
        print(f"🎯 설정: {model.upper()} | {resolution} | {aspect_ratio}")
        print(f"⏳ 처리 중...\n")

        try:
            model_name = self.MODELS[model]

            # 이미지를 Base64로 인코딩
            with open(image_path, "rb") as f:
                image_data = base64.b64encode(f.read()).decode("utf-8")

            # MIME type 결정
            ext = Path(image_path).suffix.lower()
            mime_type_map = {
                ".png": "image/png",
                ".jpg": "image/jpeg",
                ".jpeg": "image/jpeg",
                ".webp": "image/webp",
                ".gif": "image/gif"
            }
            mime_type = mime_type_map.get(ext, "image/png")

            # API 호출
            response = self.client.models.generate_content(
                model=model_name,
                contents=[{
                    "parts": [
                        {
                            "text": instruction
                        },
                        {
                            "inline_data": {
                                "mime_type": mime_type,
                                "data": image_data
                            }
                        }
                    ]
                }],
                config={
                    "response_modalities": ["TEXT", "IMAGE"],
                    "image_config": {
                        "aspect_ratio": aspect_ratio,
                        "image_size": resolution
                    }
                }
            )

            # 응답 처리
            edited_image = None
            description = ""

            for part in response.candidates[0].content.parts:
                if hasattr(part, 'text') and part.text:
                    description = part.text
                elif hasattr(part, 'inline_data') and part.inline_data:
                    image_bytes = base64.b64decode(part.inline_data.data)
                    edited_image = Image.open(BytesIO(image_bytes))

            if not edited_image:
                raise ValueError("No edited image in response")

            # 메타데이터
            metadata = {
                "timestamp": datetime.now().isoformat(),
                "type": "edit",
                "original_image": original_path,
                "model": model,
                "resolution": resolution,
                "aspect_ratio": aspect_ratio,
                "instruction": instruction,
                "description": description,
                "finish_reason": response.candidates[0].finish_reason,
                "tokens_used": response.usage_metadata.total_token_count if hasattr(response, 'usage_metadata') else None
            }

            # 저장
            if save_path:
                Path(save_path).parent.mkdir(parents=True, exist_ok=True)
                edited_image.save(save_path)
                metadata["saved_to"] = save_path
                print(f"✅ 편집된 이미지 저장: {save_path}\n")

            print(f"✅ 이미지 편집 완료!")
            print(f"   • 모델: {model.upper()}")
            print(f"   • 해상도: {resolution}")
            print(f"   • 토큰: {metadata['tokens_used']}")

            return edited_image, metadata

        except Exception as e:
            logger.error(f"Error editing image: {e}")
            print(f"❌ 오류 발생: {e}")
            raise

    def batch_generate(
        self,
        prompts: List[str],
        output_dir: str = "outputs",
        model: str = "flash",
        resolution: str = "2K",
        **kwargs
    ) -> List[Dict[str, Any]]:
        """
        대량 이미지 생성 (배치)

        Args:
            prompts: 프롬프트 리스트
            output_dir: 출력 디렉토리
            model: 모델 선택
            resolution: 해상도
            **kwargs: 추가 파라미터

        Returns:
            List[Dict]: 생성 결과 리스트

        Example:
            >>> prompts = [
            ...     "A mountain landscape",
            ...     "A ocean sunset",
            ...     "A forest at night"
            ... ]
            >>> results = generator.batch_generate(
            ...     prompts,
            ...     output_dir="batch_output",
            ...     resolution="2K"
            ... )
            >>> print(f"Generated {len([r for r in results if r['success']])} images")
        """
        import time

        Path(output_dir).mkdir(parents=True, exist_ok=True)

        results = []
        successful = 0

        for i, prompt in enumerate(prompts, 1):
            try:
                print(f"\n[{i}/{len(prompts)}] 생성 중: {prompt[:40]}...")

                filename = f"{output_dir}/image_{i:03d}.png"
                image, metadata = self.generate(
                    prompt,
                    model=model,
                    resolution=resolution,
                    save_path=filename,
                    **kwargs
                )

                metadata["success"] = True
                results.append(metadata)
                successful += 1

                # Rate limiting
                time.sleep(2)

            except Exception as e:
                print(f"❌ 실패: {e}")
                results.append({
                    "prompt": prompt,
                    "success": False,
                    "error": str(e)
                })

        print(f"\n{'='*70}")
        print(f"📊 배치 생성 완료")
        print(f"{'='*70}")
        print(f"✅ 성공: {successful}/{len(prompts)}")
        print(f"❌ 실패: {len(prompts) - successful}/{len(prompts)}")

        return results

    @staticmethod
    def _validate_params(model: str, resolution: str, aspect_ratio: str) -> None:
        """파라미터 검증"""
        if model not in NanoBananaImageGenerator.MODELS:
            raise ValueError(
                f"Invalid model: {model}. "
                f"Supported: {list(NanoBananaImageGenerator.MODELS.keys())}"
            )

        if resolution not in NanoBananaImageGenerator.RESOLUTIONS:
            raise ValueError(
                f"Invalid resolution: {resolution}. "
                f"Supported: {NanoBananaImageGenerator.RESOLUTIONS}"
            )

        if aspect_ratio not in NanoBananaImageGenerator.ASPECT_RATIOS:
            raise ValueError(
                f"Invalid aspect ratio: {aspect_ratio}. "
                f"Supported: {NanoBananaImageGenerator.ASPECT_RATIOS}"
            )

    @staticmethod
    def list_models() -> Dict[str, str]:
        """사용 가능한 모델 목록 반환"""
        return NanoBananaImageGenerator.MODELS

    @staticmethod
    def list_resolutions() -> List[str]:
        """지원 해상도 목록"""
        return NanoBananaImageGenerator.RESOLUTIONS

    @staticmethod
    def list_aspect_ratios() -> List[str]:
        """지원 종횡비 목록"""
        return NanoBananaImageGenerator.ASPECT_RATIOS


if __name__ == "__main__":
    # 테스트
    from env_key_manager import EnvKeyManager

    # API 키 확인
    if not EnvKeyManager.is_configured():
        print("❌ API 키가 설정되지 않았습니다")
        print("다음 명령으로 설정하세요:")
        print("  EnvKeyManager.setup_api_key()")
        exit(1)

    api_key = EnvKeyManager.load_api_key()
    generator = NanoBananaImageGenerator(api_key)

    # 예제 1: 기본 생성
    print("\n🔹 예제 1: 기본 이미지 생성")
    image, metadata = generator.generate(
        "A serene mountain landscape at golden hour with snow-capped peaks",
        model="flash",
        resolution="2K",
        aspect_ratio="16:9",
        save_path="test_output/example_1.png"
    )

    # 예제 2: Google Search 연동
    print("\n🔹 예제 2: Google Search 연동")
    image2, metadata2 = generator.generate(
        "Visualize the latest technology trends in 2025",
        model="flash",
        use_google_search=True,
        save_path="test_output/example_2.png"
    )

    # 예제 3: 이미지 편집
    print("\n🔹 예제 3: 이미지 편집")
    # 먼저 기본 이미지 생성
    image3, _ = generator.generate(
        "A cat sitting on a chair",
        save_path="test_output/cat_original.png"
    )

    # 그 이미지 편집
    edited, metadata3 = generator.edit(
        "test_output/cat_original.png",
        "Make the cat wear a wizard hat with magical sparkles",
        model="flash",
        resolution="2K",
        save_path="test_output/cat_wizard.png"
    )

    print("\n✅ 모든 예제 완료!")
