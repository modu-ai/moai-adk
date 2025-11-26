"""
Nano Banana Pro - Image Generation Module

Image generation and editing using Google Gemini 3 Pro Image Preview API

Official API Documentation:
- https://ai.google.dev/gemini-api/docs/image-generation
- Models: gemini-2-5-flash-image, gemini-3-pro-image-preview
- SDK: google-genai>=1.0.0
"""

import base64
import logging
import os
from datetime import datetime
from io import BytesIO
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

from google import genai
from google.api_core import exceptions
from google.genai import types

logger = logging.getLogger(__name__)


class NanoBananaImageGenerator:
    """
    Image generation and editing using Gemini 3 Nano Banana API

    Features:
    - Text-to-Image generation (1K/4K resolution)
    - Image-to-Image editing (style transfer, object manipulation)
    - Multi-turn conversational editing
    - Error handling and retry logic

    Models:
    - gemini-3-pro-image-preview: High quality, 4K resolution (Nano Banana Pro)

    Example:
        >>> generator = NanoBananaImageGenerator()
        >>> image, metadata = generator.generate(
        ...     "A serene mountain landscape at golden hour"
        ... )
        >>> image.save("output.png")
    """

    # Supported models (gemini-3-pro-image-preview only)
    MODELS = {"pro": "gemini-3-pro-image-preview"}  # High quality 4K (Nano Banana Pro)

    # Supported aspect ratios (11 options)
    ASPECT_RATIOS = [
        "1:1",  # Square
        "2:3",
        "3:2",  # Portrait/Landscape
        "3:4",
        "4:3",  # Standard
        "4:5",
        "5:4",  # Instagram
        "9:16",
        "16:9",  # Mobile/Wide
        "21:9",
        "9:21",  # Ultra wide
    ]

    # Default configuration
    DEFAULT_CONFIG = {
        "model": "pro",  # gemini-3-pro-image-preview
        "aspect_ratio": "16:9",
        "max_retries": 3,
        "timeout": 60,
    }

    def __init__(self, api_key: Optional[str] = None):
        """
        Initialize Nano Banana Image Generator

        Args:
            api_key: Google Gemini API key
                    (if None, loads from GEMINI_API_KEY or GOOGLE_API_KEY environment variable)

        Example:
            >>> generator = NanoBananaImageGenerator()
            >>> # or
            >>> generator = NanoBananaImageGenerator("your-api-key")
        """
        if api_key is None:
            api_key = os.getenv("GEMINI_API_KEY") or os.getenv("GOOGLE_API_KEY")

        if not api_key:
            raise ValueError(
                "API key not found. Set GEMINI_API_KEY or GOOGLE_API_KEY environment variable "
                "or pass api_key parameter"
            )

        self.client = genai.Client(api_key=api_key)
        logger.info("Nano Banana Image Generator initialized")

    def generate(
        self, prompt: str, model: str = "pro", aspect_ratio: str = "16:9", save_path: Optional[str] = None
    ) -> Tuple[Any, Dict[str, Any]]:
        """
        Generate Text-to-Image

        Args:
            prompt: Image generation prompt
            model: Model selection ("pro": gemini-3-pro-image-preview)
            aspect_ratio: Aspect ratio (default: "16:9")
            save_path: Image save path (optional)

        Returns:
            Tuple[PIL.Image, Dict]: (Generated image, metadata)

        Raises:
            ValueError: Invalid parameter
            Exception: API call failed

        Example:
            >>> image, metadata = generator.generate(
            ...     "A futuristic city at sunset",
            ...     model="pro",
            ...     aspect_ratio="16:9"
            ... )
            >>> print(metadata['tokens_used'])
            1234
            >>> image.save("city.png")
        """
        # Validate parameters
        self._validate_params(model, aspect_ratio)

        print(f"\n{'='*70}")
        print("🎨 Nano Banana image generation started")
        print(f"{'='*70}")
        print(f"📝 Prompt: {prompt[:50]}...")
        print(f"🎯 Settings: {model.upper()} | {aspect_ratio}")
        print("⏳ Processing...\n")

        try:
            # Get model name
            model_name = self.MODELS[model]

            # Configure request (latest google-genai SDK)
            config = types.GenerateContentConfig(
                response_modalities=["IMAGE"],
                image_config=types.ImageConfig(
                    aspect_ratio=aspect_ratio,
                ),
            )

            # API call
            response = self.client.models.generate_content(
                model=model_name,
                contents=prompt,
                config=config,
            )

            # Process response
            image = None
            description = ""

            for part in response.parts:
                if hasattr(part, "text") and part.text:
                    description = part.text
                elif hasattr(part, "inline_data") and part.inline_data:
                    # inline_data.data is already bytes type (no base64 decoding needed)
                    from PIL import Image

                    image_data = part.inline_data.data
                    if isinstance(image_data, str):
                        # If string, decode base64
                        image_bytes = base64.b64decode(image_data)
                    else:
                        # If already bytes, use directly
                        image_bytes = image_data
                    image = Image.open(BytesIO(image_bytes))

            if not image:
                raise ValueError("No image data in response")

            # Build metadata
            tokens_used = 0
            if hasattr(response, "usage_metadata") and response.usage_metadata:
                tokens_used = getattr(response.usage_metadata, "total_token_count", 0)

            metadata = {
                "timestamp": datetime.now().isoformat(),
                "model": model,
                "model_name": model_name,
                "aspect_ratio": aspect_ratio,
                "prompt": prompt,
                "description": description,
                "tokens_used": tokens_used,
            }

            # Save
            if save_path:
                Path(save_path).parent.mkdir(parents=True, exist_ok=True)
                image.save(save_path)
                metadata["saved_to"] = save_path
                print(f"✅ Image saved: {save_path}\n")

            print("✅ Image generation completed!")
            print(f"   • Model: {model.upper()}")
            print(f"   • Aspect ratio: {aspect_ratio}")
            print(f"   • Tokens: {metadata['tokens_used']}")

            return image, metadata

        except exceptions.ResourceExhausted:
            logger.error("API quota exceeded")
            print("❌ API quota exceeded")
            print("   • Please try again in a few minutes")
            raise

        except exceptions.PermissionDenied:
            logger.error("Permission denied - check API key")
            print("❌ Permission error - Please check API key")
            raise

        except exceptions.InvalidArgument as e:
            logger.error(f"Invalid argument: {e}")
            print(f"❌ Invalid parameter: {e}")
            raise

        except Exception as e:
            logger.error(f"Error generating image: {e}")
            print(f"❌ Error occurred: {e}")
            raise

    def edit(
        self,
        image_path: str,
        instruction: str,
        model: str = "pro",
        aspect_ratio: str = "16:9",
        save_path: Optional[str] = None,
    ) -> Tuple[Any, Dict[str, Any]]:
        """
        Image-to-Image editing

        Args:
            image_path: Path to image to edit
            instruction: Edit instruction
            model: Model selection
            aspect_ratio: Output aspect ratio
            save_path: Result save path

        Returns:
            Tuple[PIL.Image, Dict]: (Edited image, metadata)

        Example:
            >>> edited_image, metadata = generator.edit(
            ...     "original.png",
            ...     "Add a sunset in the background",
            ...     model="pro"
            ... )
            >>> edited_image.save("with_sunset.png")
        """
        # Validate parameters
        self._validate_params(model, aspect_ratio)

        # Load image
        if not Path(image_path).exists():
            raise FileNotFoundError(f"Image not found: {image_path}")

        from PIL import Image

        Image.open(image_path)
        original_path = str(Path(image_path).resolve())

        print(f"\n{'='*70}")
        print("✏️  Image editing started")
        print(f"{'='*70}")
        print(f"📁 Original: {original_path}")
        print(f"📝 Instruction: {instruction[:50]}...")
        print(f"🎯 Settings: {model.upper()} | {aspect_ratio}")
        print("⏳ Processing...\n")

        try:
            model_name = self.MODELS[model]

            # Encode image to Base64
            with open(image_path, "rb") as f:
                image_data = base64.b64encode(f.read()).decode("utf-8")

            # Determine MIME type
            ext = Path(image_path).suffix.lower()
            mime_type_map = {
                ".png": "image/png",
                ".jpg": "image/jpeg",
                ".jpeg": "image/jpeg",
                ".webp": "image/webp",
                ".gif": "image/gif",
            }
            mime_type = mime_type_map.get(ext, "image/png")

            # Configure request
            config = types.GenerateContentConfig(
                response_modalities=["IMAGE"],
                image_config=types.ImageConfig(
                    aspect_ratio=aspect_ratio,
                ),
            )

            # API call (multimodal input)
            response = self.client.models.generate_content(
                model=model_name,
                contents=[
                    types.Part.from_text(instruction),
                    types.Part.from_bytes(data=base64.b64decode(image_data), mime_type=mime_type),
                ],
                config=config,
            )

            # Process response
            edited_image = None
            description = ""

            for part in response.parts:
                if hasattr(part, "text") and part.text:
                    description = part.text
                elif hasattr(part, "inline_data") and part.inline_data:
                    # inline_data.data is already bytes type
                    image_data = part.inline_data.data
                    if isinstance(image_data, str):
                        image_bytes = base64.b64decode(image_data)
                    else:
                        image_bytes = image_data
                    edited_image = Image.open(BytesIO(image_bytes))

            if not edited_image:
                raise ValueError("No edited image in response")

            # Metadata
            tokens_used = 0
            if hasattr(response, "usage_metadata") and response.usage_metadata:
                tokens_used = getattr(response.usage_metadata, "total_token_count", 0)

            metadata = {
                "timestamp": datetime.now().isoformat(),
                "type": "edit",
                "original_image": original_path,
                "model": model,
                "model_name": model_name,
                "aspect_ratio": aspect_ratio,
                "instruction": instruction,
                "description": description,
                "tokens_used": tokens_used,
            }

            # Save
            if save_path:
                Path(save_path).parent.mkdir(parents=True, exist_ok=True)
                edited_image.save(save_path)
                metadata["saved_to"] = save_path
                print(f"✅ Edited image saved: {save_path}\n")

            print("✅ Image editing completed!")
            print(f"   • Model: {model.upper()}")
            print(f"   • Aspect ratio: {aspect_ratio}")
            print(f"   • Tokens: {metadata['tokens_used']}")

            return edited_image, metadata

        except Exception as e:
            logger.error(f"Error editing image: {e}")
            print(f"❌ Error occurred: {e}")
            raise

    def batch_generate(
        self, prompts: List[str], output_dir: str = "outputs", model: str = "pro", **kwargs
    ) -> List[Dict[str, Any]]:
        """
        Batch image generation

        Args:
            prompts: List of prompts
            output_dir: Output directory
            model: Model selection
            **kwargs: Additional parameters

        Returns:
            List[Dict]: List of generation results

        Example:
            >>> prompts = [
            ...     "A mountain landscape",
            ...     "A ocean sunset",
            ...     "A forest at night"
            ... ]
            >>> results = generator.batch_generate(
            ...     prompts,
            ...     output_dir="batch_output"
            ... )
            >>> print(f"Generated {len([r for r in results if r['success']])} images")
        """
        import time

        Path(output_dir).mkdir(parents=True, exist_ok=True)

        results = []
        successful = 0

        for i, prompt in enumerate(prompts, 1):
            try:
                print(f"\n[{i}/{len(prompts)}] Generating: {prompt[:40]}...")

                filename = f"{output_dir}/image_{i:03d}.png"
                image, metadata = self.generate(prompt, model=model, save_path=filename, **kwargs)

                metadata["success"] = True
                results.append(metadata)
                successful += 1

                # Rate limiting
                time.sleep(2)

            except Exception as e:
                print(f"❌ Failed: {e}")
                results.append({"prompt": prompt, "success": False, "error": str(e)})

        print(f"\n{'='*70}")
        print("📊 Batch generation completed")
        print(f"{'='*70}")
        print(f"✅ Success: {successful}/{len(prompts)}")
        print(f"❌ Failed: {len(prompts) - successful}/{len(prompts)}")

        return results

    @staticmethod
    def _validate_params(model: str, aspect_ratio: str) -> None:
        """Validate parameters"""
        if model not in NanoBananaImageGenerator.MODELS:
            raise ValueError(f"Invalid model: {model}. " f"Supported: {list(NanoBananaImageGenerator.MODELS.keys())}")

        if aspect_ratio not in NanoBananaImageGenerator.ASPECT_RATIOS:
            raise ValueError(
                f"Invalid aspect ratio: {aspect_ratio}. " f"Supported: {NanoBananaImageGenerator.ASPECT_RATIOS}"
            )

    @staticmethod
    def list_models() -> Dict[str, str]:
        """Return list of available models"""
        return NanoBananaImageGenerator.MODELS

    @staticmethod
    def list_aspect_ratios() -> List[str]:
        """Return list of supported aspect ratios"""
        return NanoBananaImageGenerator.ASPECT_RATIOS


if __name__ == "__main__":
    # Test
    from env_key_manager import EnvKeyManager

    # Check API key
    if not EnvKeyManager.is_configured():
        print("❌ API key not configured")
        print("Please configure with:")
        print("  EnvKeyManager.setup_api_key()")
        exit(1)

    api_key = EnvKeyManager.load_api_key()
    generator = NanoBananaImageGenerator(api_key)

    # Example 1: Basic generation
    print("\n🔹 Example 1: Basic image generation")
    image, metadata = generator.generate(
        "A serene mountain landscape at golden hour with snow-capped peaks",
        aspect_ratio="16:9",
        save_path="test_output/example_1.png",
    )

    # Example 2: Image editing
    print("\n🔹 Example 2: Image editing")
    # First generate base image
    image2, _ = generator.generate("A cat sitting on a chair", save_path="test_output/cat_original.png")

    # Edit that image
    edited, metadata2 = generator.edit(
        "test_output/cat_original.png",
        "Make the cat wear a wizard hat with magical sparkles",
        save_path="test_output/cat_wizard.png",
    )

    print("\n✅ All examples completed!")
