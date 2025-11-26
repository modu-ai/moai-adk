"""
Nano Banana - Image Generation Module

Professional image generation using Google Gemini Image Generation API

Official API Documentation:
- https://ai.google.dev/gemini-api/docs/image-generation
- Models: gemini-2.5-flash-image, gemini-3-pro-image-preview
- SDK: google-genai>=1.0.0

Features:
- Text-to-Image generation (1K/2K/4K resolution)
- Image-to-Image editing and style transfer
- Multi-turn conversational refinement
- Google Search grounding for real-time data
- Multi-image reference support (up to 14 images)
- Advanced text rendering in images
- Interactive model selection (ask_model_selection)
- Prompt writing guidance (get_prompt_guidance)

Usage Example:
    >>> from modules.image_generator import NanoBananaImageGenerator, ask_model_selection, get_prompt_guidance
    >>>
    >>> # Get prompt guidance
    >>> guidance = get_prompt_guidance("generation")
    >>> print(guidance)
    >>>
    >>> # Ask user for model selection
    >>> model = ask_model_selection("product photography")
    >>>
    >>> # Generate image with selected model
    >>> generator = NanoBananaImageGenerator()
    >>> image, metadata = generator.generate(
    ...     prompt="Professional product photo...",
    ...     model=model,
    ...     resolution="4K" if model == "pro" else None
    ... )
"""

import os
import base64
from pathlib import Path
from typing import Optional, List, Tuple, Dict, Any
from datetime import datetime
import logging
from io import BytesIO

from google import genai
from google.genai import types
from google.api_core import exceptions

logger = logging.getLogger(__name__)


def ask_model_selection(task_description: str = "image generation") -> str:
    """
    Ask user which model they want to use for image generation.

    Args:
        task_description: Description of the task (default: "image generation")

    Returns:
        str: Selected model ("flash" or "pro")

    Example:
        >>> model = ask_model_selection("product photography")
        >>> print(f"Selected model: {model}")
        Selected model: pro
    """
    print(f"\n{'='*70}")
    print(f"Model Selection for {task_description}")
    print(f"{'='*70}\n")

    print("Available models:")
    print("\n1. Gemini 3 Pro Image (Nano Banana Pro)")
    print("   - Professional quality, up to 4K resolution")
    print("   - Google Search grounding, 'thinking' process")
    print("   - Best for: Professional assets, advertising, high-end work")
    print("   - Speed: Slower (10-60s), Cost: Higher")

    print("\n2. Gemini 2.5 Flash Image (Nano Banana)")
    print("   - Fast generation, 1K resolution")
    print("   - Optimized for batch work, low latency")
    print("   - Best for: Quick prototyping, batch processing, variations")
    print("   - Speed: Fast (5-15s), Cost: Lower")

    print("\n" + "="*70)

    while True:
        choice = input("\nSelect model (1=Pro, 2=Flash, or enter 'pro'/'flash'): ").strip().lower()

        if choice in ["1", "pro"]:
            print("\n✓ Selected: Gemini 3 Pro Image (Professional quality)")
            return "pro"
        elif choice in ["2", "flash"]:
            print("\n✓ Selected: Gemini 2.5 Flash Image (Fast generation)")
            return "flash"
        else:
            print("Invalid choice. Please enter 1, 2, 'pro', or 'flash'.")


def get_prompt_guidance(task_type: str = "generation") -> str:
    """
    Provide prompt writing guidance based on task type.

    Args:
        task_type: Type of task ("generation" or "editing")

    Returns:
        str: Prompt guidance text

    Example:
        >>> guidance = get_prompt_guidance("generation")
        >>> print(guidance)
        [Prompt guidance for image generation...]
    """
    if task_type == "generation":
        return """
Prompt Writing Tips for Image Generation:
==========================================

Essential Elements:
- Clear Subject: "Female CEO with blue glasses"
- Style: "Matte acrylic painting style"
- Composition: "Standing on left, hands on table"
- Lighting: "Warm golden hour lighting"
- Details: "Sharp focus, blurred background, cinematic quality"
- Colors: "Warm tones, photorealistic"

Example Professional Prompt:
"A serene Japanese garden in spring, with cherry blossoms reflected
in a still pond. A single wooden bridge spans the water.
Soft golden hour lighting, photorealistic, cinematic composition,
shallow depth of field."

Template:
[Subject & Action] + [Setting] + [Lighting] + [Camera/Lens] +
[Depth of Field] + [Color Palette] + [Style] + [Quality]
"""
    elif task_type == "editing":
        return """
Prompt Writing Tips for Image Editing:
=======================================

Essential Elements:
- Clear Changes: "Change background to autumn forest"
- Preserve Elements: "Keep person as is"
- Style Consistency: "Same photographic style"
- Lighting Consistency: "Same lighting conditions"

Example Editing Prompt:
"Change the background from office to a modern coffee shop.
Keep the person in the same pose and lighting.
Maintain photorealistic style, warm tones."

Template:
[What to Change] + [What to Preserve] + [How to Change] +
[Style Consistency]
"""
    else:
        return "Unknown task type. Please specify 'generation' or 'editing'."


class NanoBananaImageGenerator:
    """
    Professional image generation using Gemini Image Generation API

    Supports both Gemini 2.5 Flash (fast) and Gemini 3 Pro (professional quality).

    Features:
    - Text-to-Image: Generate images from detailed prompts
    - Image-to-Image: Edit existing images with instructions
    - Multi-turn Chat: Iterative refinement through conversation
    - Google Search: Real-time grounding for factual content
    - Multi-reference: Up to 14 reference images (6 objects + 5 persons)
    - High Resolution: 1K, 2K, 4K output options (Gemini 3 Pro)
    - Advanced Text: Sophisticated text rendering (Gemini 3 Pro)

    Models:
    - gemini-2.5-flash-image: Fast, general-purpose (~5-15s)
    - gemini-3-pro-image-preview: Professional quality, advanced features (~10-60s)

    Example:
        >>> generator = NanoBananaImageGenerator()
        >>>
        >>> # Basic text-to-image
        >>> image, metadata = generator.generate(
        ...     prompt="A serene mountain landscape at golden hour",
        ...     model="flash",
        ...     aspect_ratio="16:9"
        ... )
        >>> image.save("landscape.png")
        >>>
        >>> # High-resolution professional output
        >>> image, metadata = generator.generate(
        ...     prompt="Professional product photo of wireless headphones",
        ...     model="pro",
        ...     resolution="4K",
        ...     aspect_ratio="1:1"
        ... )
        >>>
        >>> # Image editing
        >>> edited, metadata = generator.edit(
        ...     image_path="photo.png",
        ...     instruction="Add a sunset in the background",
        ...     model="pro"
        ... )
        >>>
        >>> # Multi-turn refinement with chat
        >>> chat = generator.create_chat(model="pro")
        >>> response1 = chat.send_message("Create a modern office illustration")
        >>> response2 = chat.send_message("Make it more colorful")
        >>> response3 = chat.send_message("Add people working")
    """

    # Supported models
    MODELS = {
        "flash": "gemini-2.5-flash-image",      # Fast, general-purpose
        "pro": "gemini-3-pro-image-preview"      # Professional, advanced features
    }

    # Supported aspect ratios (11 options)
    ASPECT_RATIOS = [
        "1:1",         # Square
        "2:3", "3:2",  # Portrait/Landscape
        "3:4", "4:3",  # Standard
        "4:5", "5:4",  # Instagram
        "9:16", "16:9", # Mobile/Wide
        "21:9", "9:21" # Ultra wide
    ]

    # Resolution options (Gemini 3 Pro only)
    RESOLUTIONS = ["1K", "2K", "4K"]

    def __init__(self, api_key: Optional[str] = None):
        """
        Initialize Nano Banana Image Generator

        Args:
            api_key: Google Gemini API key
                    (if None, loads from GEMINI_API_KEY or GOOGLE_API_KEY environment variable)

        Raises:
            ValueError: If API key not found

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
        self,
        prompt: str,
        model: str = "flash",
        aspect_ratio: str = "16:9",
        resolution: Optional[str] = None,
        enable_google_search: bool = False,
        save_path: Optional[str] = None
    ) -> Tuple[Any, Dict[str, Any]]:
        """
        Generate image from text prompt (Text-to-Image)

        Args:
            prompt: Detailed image generation prompt
            model: Model selection
                  - "flash": Fast, general-purpose (gemini-2.5-flash-image)
                  - "pro": Professional quality (gemini-3-pro-image-preview)
            aspect_ratio: Output aspect ratio (default: "16:9")
                         Options: 1:1, 2:3, 3:2, 3:4, 4:3, 4:5, 5:4, 9:16, 16:9, 21:9, 9:21
            resolution: Output resolution (Gemini 3 Pro only)
                       Options: "1K", "2K", "4K" (default: "1K")
                       Note: Use uppercase 'K'
            enable_google_search: Enable real-time Google Search grounding
            save_path: Image save path (optional)

        Returns:
            Tuple[PIL.Image, Dict]: (Generated image, metadata)

        Raises:
            ValueError: Invalid parameters
            exceptions.ResourceExhausted: API quota exceeded
            exceptions.PermissionDenied: Permission denied
            exceptions.InvalidArgument: Invalid argument

        Example:
            >>> # Basic generation
            >>> image, metadata = generator.generate(
            ...     prompt="A futuristic city at sunset",
            ...     model="flash",
            ...     aspect_ratio="16:9"
            ... )
            >>>
            >>> # High-resolution professional output
            >>> image, metadata = generator.generate(
            ...     prompt="Professional product photo of luxury watch",
            ...     model="pro",
            ...     resolution="4K",
            ...     aspect_ratio="1:1"
            ... )
            >>>
            >>> # Real-time data integration
            >>> image, metadata = generator.generate(
            ...     prompt="Infographic about current AI breakthroughs",
            ...     model="pro",
            ...     enable_google_search=True
            ... )
        """
        # Validate parameters
        self._validate_params(model, aspect_ratio, resolution)

        print(f"\n{'='*70}")
        print(f"🎨 Nano Banana image generation started")
        print(f"{'='*70}")
        print(f"📝 Prompt: {prompt[:50]}...")
        print(f"🎯 Model: {model.upper()}")
        print(f"📐 Aspect ratio: {aspect_ratio}")
        if resolution and model == "pro":
            print(f"🖼️  Resolution: {resolution}")
        if enable_google_search:
            print(f"🔍 Google Search: Enabled")
        print(f"⏳ Processing...\n")

        try:
            # Get model name
            model_name = self.MODELS[model]

            # Build config
            config_params = {
                "response_modalities": ["TEXT", "IMAGE"],
                "image_config": types.ImageConfig(
                    aspect_ratio=aspect_ratio,
                ),
            }

            # Add resolution for Gemini 3 Pro
            if model == "pro" and resolution:
                config_params["image_config"] = types.ImageConfig(
                    aspect_ratio=aspect_ratio,
                    image_size=resolution
                )

            # Add Google Search tool if enabled
            if enable_google_search:
                config_params["tools"] = [{"google_search": {}}]

            config = types.GenerateContentConfig(**config_params)

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
                if hasattr(part, 'text') and part.text:
                    description = part.text
                elif hasattr(part, 'inline_data') and part.inline_data:
                    # inline_data.data is already bytes
                    image_data = part.inline_data.data
                    if isinstance(image_data, str):
                        image_bytes = base64.b64decode(image_data)
                    else:
                        image_bytes = image_data

                    from PIL import Image
                    image = Image.open(BytesIO(image_bytes))

            if not image:
                raise ValueError("No image data in response")

            # Build metadata
            tokens_used = 0
            if hasattr(response, 'usage_metadata') and response.usage_metadata:
                tokens_used = getattr(response.usage_metadata, 'total_token_count', 0)

            metadata = {
                "timestamp": datetime.now().isoformat(),
                "model": model,
                "model_name": model_name,
                "aspect_ratio": aspect_ratio,
                "resolution": resolution if model == "pro" else None,
                "google_search": enable_google_search,
                "prompt": prompt,
                "description": description,
                "tokens_used": tokens_used
            }

            # Save
            if save_path:
                Path(save_path).parent.mkdir(parents=True, exist_ok=True)
                image.save(save_path)
                metadata["saved_to"] = save_path
                print(f"✅ Image saved: {save_path}\n")

            print(f"✅ Image generation completed!")
            print(f"   • Model: {model.upper()}")
            print(f"   • Aspect ratio: {aspect_ratio}")
            if resolution and model == "pro":
                print(f"   • Resolution: {resolution}")
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
        model: str = "flash",
        aspect_ratio: str = "16:9",
        resolution: Optional[str] = None,
        save_path: Optional[str] = None
    ) -> Tuple[Any, Dict[str, Any]]:
        """
        Edit existing image with instructions (Image-to-Image)

        Args:
            image_path: Path to image to edit
            instruction: Edit instruction (detailed description)
            model: Model selection ("flash" or "pro")
            aspect_ratio: Output aspect ratio
            resolution: Output resolution (Gemini 3 Pro only)
            save_path: Result save path

        Returns:
            Tuple[PIL.Image, Dict]: (Edited image, metadata)

        Example:
            >>> # Style transfer
            >>> edited, metadata = generator.edit(
            ...     image_path="photo.png",
            ...     instruction="Transform to Van Gogh oil painting style",
            ...     model="pro"
            ... )
            >>>
            >>> # Object manipulation
            >>> edited, metadata = generator.edit(
            ...     image_path="landscape.png",
            ...     instruction="Remove the telephone pole and fill background naturally",
            ...     model="pro"
            ... )
            >>>
            >>> # Add elements
            >>> edited, metadata = generator.edit(
            ...     image_path="room.png",
            ...     instruction="Add a sunset view through the window",
            ...     model="flash"
            ... )
        """
        # Validate parameters
        self._validate_params(model, aspect_ratio, resolution)

        # Load image
        if not Path(image_path).exists():
            raise FileNotFoundError(f"Image not found: {image_path}")

        from PIL import Image
        original_image = Image.open(image_path)
        original_path = str(Path(image_path).resolve())

        print(f"\n{'='*70}")
        print(f"✏️  Image editing started")
        print(f"{'='*70}")
        print(f"📁 Original: {original_path}")
        print(f"📝 Instruction: {instruction[:50]}...")
        print(f"🎯 Model: {model.upper()}")
        print(f"⏳ Processing...\n")

        try:
            model_name = self.MODELS[model]

            # Determine MIME type
            ext = Path(image_path).suffix.lower()
            mime_type_map = {
                ".png": "image/png",
                ".jpg": "image/jpeg",
                ".jpeg": "image/jpeg",
                ".webp": "image/webp",
                ".gif": "image/gif"
            }
            mime_type = mime_type_map.get(ext, "image/png")

            # Read image bytes
            with open(image_path, "rb") as f:
                image_bytes = f.read()

            # Build config
            config_params = {
                "response_modalities": ["TEXT", "IMAGE"],
                "image_config": types.ImageConfig(
                    aspect_ratio=aspect_ratio,
                ),
            }

            if model == "pro" and resolution:
                config_params["image_config"] = types.ImageConfig(
                    aspect_ratio=aspect_ratio,
                    image_size=resolution
                )

            config = types.GenerateContentConfig(**config_params)

            # API call (multimodal input)
            response = self.client.models.generate_content(
                model=model_name,
                contents=[
                    types.Part.from_text(instruction),
                    types.Part.from_bytes(
                        data=image_bytes,
                        mime_type=mime_type
                    )
                ],
                config=config,
            )

            # Process response
            edited_image = None
            description = ""

            for part in response.parts:
                if hasattr(part, 'text') and part.text:
                    description = part.text
                elif hasattr(part, 'inline_data') and part.inline_data:
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
            if hasattr(response, 'usage_metadata') and response.usage_metadata:
                tokens_used = getattr(response.usage_metadata, 'total_token_count', 0)

            metadata = {
                "timestamp": datetime.now().isoformat(),
                "type": "edit",
                "original_image": original_path,
                "model": model,
                "model_name": model_name,
                "aspect_ratio": aspect_ratio,
                "resolution": resolution if model == "pro" else None,
                "instruction": instruction,
                "description": description,
                "tokens_used": tokens_used
            }

            # Save
            if save_path:
                Path(save_path).parent.mkdir(parents=True, exist_ok=True)
                edited_image.save(save_path)
                metadata["saved_to"] = save_path
                print(f"✅ Edited image saved: {save_path}\n")

            print(f"✅ Image editing completed!")
            print(f"   • Model: {model.upper()}")
            print(f"   • Tokens: {metadata['tokens_used']}")

            return edited_image, metadata

        except Exception as e:
            logger.error(f"Error editing image: {e}")
            print(f"❌ Error occurred: {e}")
            raise

    def create_chat(
        self,
        model: str = "pro",
        enable_google_search: bool = False
    ):
        """
        Create multi-turn chat for iterative image refinement

        Args:
            model: Model selection ("flash" or "pro")
            enable_google_search: Enable Google Search grounding

        Returns:
            Chat: Chat session object

        Example:
            >>> chat = generator.create_chat(model="pro")
            >>>
            >>> # Turn 1: Generate initial image
            >>> response1 = chat.send_message(
            ...     "Create a vibrant infographic explaining photosynthesis"
            ... )
            >>> image1 = response1.get_image()
            >>> image1.save("turn1.png")
            >>>
            >>> # Turn 2: Refine
            >>> response2 = chat.send_message(
            ...     "Make it more colorful and add a title"
            ... )
            >>> image2 = response2.get_image()
            >>> image2.save("turn2.png")
            >>>
            >>> # Turn 3: Final touches
            >>> response3 = chat.send_message(
            ...     "Translate all text to Spanish"
            ... )
            >>> image3 = response3.get_image()
            >>> image3.save("final.png")
        """
        model_name = self.MODELS[model]

        config_params = {
            "response_modalities": ["TEXT", "IMAGE"]
        }

        if enable_google_search:
            config_params["tools"] = [{"google_search": {}}]

        config = types.GenerateContentConfig(**config_params)

        return self.client.chats.create(
            model=model_name,
            config=config
        )

    def generate_with_references(
        self,
        prompt: str,
        reference_images: List[str],
        model: str = "pro",
        aspect_ratio: str = "16:9",
        resolution: Optional[str] = None,
        save_path: Optional[str] = None
    ) -> Tuple[Any, Dict[str, Any]]:
        """
        Generate image using multiple reference images (Gemini 3 Pro only)

        Supports up to 14 reference images:
        - Up to 6 high-fidelity object images
        - Up to 5 person images for character consistency

        Args:
            prompt: Generation prompt
            reference_images: List of reference image paths (max 14)
            model: Model selection (must be "pro")
            aspect_ratio: Output aspect ratio
            resolution: Output resolution ("1K", "2K", or "4K")
            save_path: Save path

        Returns:
            Tuple[PIL.Image, Dict]: (Generated image, metadata)

        Example:
            >>> # Style transfer from multiple references
            >>> image, metadata = generator.generate_with_references(
            ...     prompt="An office group photo of these people making funny faces",
            ...     reference_images=[
            ...         "person1.png",
            ...         "person2.png",
            ...         "person3.png",
            ...         "person4.png",
            ...         "person5.png"
            ...     ],
            ...     model="pro",
            ...     aspect_ratio="5:4",
            ...     resolution="2K"
            ... )
        """
        if model != "pro":
            raise ValueError("Multi-image reference only supported with Gemini 3 Pro (model='pro')")

        if len(reference_images) > 14:
            raise ValueError("Maximum 14 reference images supported")

        print(f"\n{'='*70}")
        print(f"🎨 Multi-reference image generation started")
        print(f"{'='*70}")
        print(f"📝 Prompt: {prompt[:50]}...")
        print(f"📚 References: {len(reference_images)} images")
        print(f"⏳ Processing...\n")

        try:
            model_name = self.MODELS[model]

            # Build contents with text and images
            contents = [types.Part.from_text(prompt)]

            for img_path in reference_images:
                if not Path(img_path).exists():
                    raise FileNotFoundError(f"Reference image not found: {img_path}")

                ext = Path(img_path).suffix.lower()
                mime_type_map = {
                    ".png": "image/png",
                    ".jpg": "image/jpeg",
                    ".jpeg": "image/jpeg",
                    ".webp": "image/webp"
                }
                mime_type = mime_type_map.get(ext, "image/png")

                with open(img_path, "rb") as f:
                    image_bytes = f.read()

                contents.append(
                    types.Part.from_bytes(
                        data=image_bytes,
                        mime_type=mime_type
                    )
                )

            # Build config
            config_params = {
                "response_modalities": ["TEXT", "IMAGE"],
                "image_config": types.ImageConfig(
                    aspect_ratio=aspect_ratio,
                    image_size=resolution or "1K"
                ),
            }

            config = types.GenerateContentConfig(**config_params)

            # API call
            response = self.client.models.generate_content(
                model=model_name,
                contents=contents,
                config=config,
            )

            # Process response
            image = None
            description = ""

            for part in response.parts:
                if hasattr(part, 'text') and part.text:
                    description = part.text
                elif hasattr(part, 'inline_data') and part.inline_data:
                    from PIL import Image
                    image_data = part.inline_data.data
                    if isinstance(image_data, str):
                        image_bytes = base64.b64decode(image_data)
                    else:
                        image_bytes = image_data
                    image = Image.open(BytesIO(image_bytes))

            if not image:
                raise ValueError("No image in response")

            # Metadata
            tokens_used = 0
            if hasattr(response, 'usage_metadata') and response.usage_metadata:
                tokens_used = getattr(response.usage_metadata, 'total_token_count', 0)

            metadata = {
                "timestamp": datetime.now().isoformat(),
                "type": "multi_reference",
                "model": model,
                "reference_count": len(reference_images),
                "aspect_ratio": aspect_ratio,
                "resolution": resolution,
                "prompt": prompt,
                "description": description,
                "tokens_used": tokens_used
            }

            if save_path:
                Path(save_path).parent.mkdir(parents=True, exist_ok=True)
                image.save(save_path)
                metadata["saved_to"] = save_path
                print(f"✅ Image saved: {save_path}\n")

            print(f"✅ Multi-reference generation completed!")
            print(f"   • References: {len(reference_images)}")
            print(f"   • Tokens: {metadata['tokens_used']}")

            return image, metadata

        except Exception as e:
            logger.error(f"Error in multi-reference generation: {e}")
            print(f"❌ Error occurred: {e}")
            raise

    def batch_generate(
        self,
        prompts: List[str],
        output_dir: str = "outputs",
        model: str = "flash",
        **kwargs
    ) -> List[Dict[str, Any]]:
        """
        Batch image generation with error handling

        Args:
            prompts: List of prompts
            output_dir: Output directory
            model: Model selection
            **kwargs: Additional parameters (aspect_ratio, resolution, etc.)

        Returns:
            List[Dict]: List of generation results

        Example:
            >>> prompts = [
            ...     "A mountain landscape at sunset",
            ...     "A futuristic city with flying cars",
            ...     "An underwater coral reef scene"
            ... ]
            >>> results = generator.batch_generate(
            ...     prompts,
            ...     output_dir="batch_output",
            ...     model="flash",
            ...     aspect_ratio="16:9"
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
                image, metadata = self.generate(
                    prompt,
                    model=model,
                    save_path=filename,
                    **kwargs
                )

                metadata["success"] = True
                results.append(metadata)
                successful += 1

                # Rate limiting
                time.sleep(2)

            except Exception as e:
                print(f"❌ Failed: {e}")
                results.append({
                    "prompt": prompt,
                    "success": False,
                    "error": str(e)
                })

        print(f"\n{'='*70}")
        print(f"📊 Batch generation completed")
        print(f"{'='*70}")
        print(f"✅ Success: {successful}/{len(prompts)}")
        print(f"❌ Failed: {len(prompts) - successful}/{len(prompts)}")

        return results

    @staticmethod
    def _validate_params(
        model: str,
        aspect_ratio: str,
        resolution: Optional[str] = None
    ) -> None:
        """Validate parameters"""
        if model not in NanoBananaImageGenerator.MODELS:
            raise ValueError(
                f"Invalid model: {model}. "
                f"Supported: {list(NanoBananaImageGenerator.MODELS.keys())}"
            )

        if aspect_ratio not in NanoBananaImageGenerator.ASPECT_RATIOS:
            raise ValueError(
                f"Invalid aspect ratio: {aspect_ratio}. "
                f"Supported: {NanoBananaImageGenerator.ASPECT_RATIOS}"
            )

        if resolution and resolution not in NanoBananaImageGenerator.RESOLUTIONS:
            raise ValueError(
                f"Invalid resolution: {resolution}. "
                f"Supported: {NanoBananaImageGenerator.RESOLUTIONS}"
            )

        if resolution and model != "pro":
            raise ValueError(
                f"Resolution setting only supported with Gemini 3 Pro (model='pro'). "
                f"Current model: {model}"
            )

    @staticmethod
    def list_models() -> Dict[str, str]:
        """Return list of available models"""
        return NanoBananaImageGenerator.MODELS

    @staticmethod
    def list_aspect_ratios() -> List[str]:
        """Return list of supported aspect ratios"""
        return NanoBananaImageGenerator.ASPECT_RATIOS

    @staticmethod
    def list_resolutions() -> List[str]:
        """Return list of supported resolutions"""
        return NanoBananaImageGenerator.RESOLUTIONS


if __name__ == "__main__":
    # Test with model selection and prompt guidance
    try:
        from env_key_manager import EnvKeyManager

        # Check API key
        if not EnvKeyManager.is_configured():
            print("❌ API key not configured")
            print("Please configure with:")
            print("  EnvKeyManager.setup_api_key()")
            exit(1)

        api_key = EnvKeyManager.load_api_key()
    except ImportError:
        # Fallback if env_key_manager not available
        api_key = os.getenv("GEMINI_API_KEY") or os.getenv("GOOGLE_API_KEY")
        if not api_key:
            print("❌ API key not found")
            print("Set GEMINI_API_KEY or GOOGLE_API_KEY environment variable")
            exit(1)

    generator = NanoBananaImageGenerator(api_key)

    # Example 1: With model selection and prompt guidance
    print("\n🔹 Example 1: Interactive model selection")
    print(get_prompt_guidance("generation"))

    model = ask_model_selection("landscape photography")

    image, metadata = generator.generate(
        "A serene mountain landscape at golden hour with snow-capped peaks",
        model=model,
        aspect_ratio="16:9",
        resolution="4K" if model == "pro" else None,
        save_path="test_output/example_1_selected_model.png"
    )

    # Example 2: High-resolution professional output (pre-selected)
    print("\n🔹 Example 2: High-resolution 4K image (Pro model)")
    image2, metadata2 = generator.generate(
        "Professional product photo of luxury watch, studio lighting, 4K detail",
        model="pro",
        aspect_ratio="1:1",
        resolution="4K",
        save_path="test_output/example_2_4k.png"
    )

    # Example 3: Image editing with guidance
    print("\n🔹 Example 3: Image editing")
    print(get_prompt_guidance("editing"))

    # First generate base image
    image3, _ = generator.generate(
        "A cat sitting on a chair",
        model="flash",
        save_path="test_output/example_3_original.png"
    )

    # Edit that image
    edited, metadata3 = generator.edit(
        "test_output/example_3_original.png",
        "Make the cat wear a wizard hat with magical sparkles",
        model="flash",
        save_path="test_output/example_3_edited.png"
    )

    # Example 4: Multi-turn chat
    print("\n🔹 Example 4: Multi-turn refinement")
    chat = generator.create_chat(model="pro")

    response1 = chat.send_message("Create a modern office illustration")
    for part in response1.parts:
        if hasattr(part, 'inline_data') and part.inline_data:
            from PIL import Image
            img_data = part.inline_data.data
            if isinstance(img_data, str):
                img_bytes = base64.b64decode(img_data)
            else:
                img_bytes = img_data
            img = Image.open(BytesIO(img_bytes))
            img.save("test_output/example_4_turn1.png")

    response2 = chat.send_message("Make it more colorful and add people")
    for part in response2.parts:
        if hasattr(part, 'inline_data') and part.inline_data:
            from PIL import Image
            img_data = part.inline_data.data
            if isinstance(img_data, str):
                img_bytes = base64.b64decode(img_data)
            else:
                img_bytes = img_data
            img = Image.open(BytesIO(img_bytes))
            img.save("test_output/example_4_turn2_refined.png")

    print("\n✅ All examples completed!")
    print(f"   • Check test_output/ directory for generated images")
    print(f"\n💡 Tips:")
    print(f"   • Use ask_model_selection() to let users choose model interactively")
    print(f"   • Use get_prompt_guidance() to show prompt writing tips")
    print(f"   • Refer to examples.md for comprehensive prompt examples")
