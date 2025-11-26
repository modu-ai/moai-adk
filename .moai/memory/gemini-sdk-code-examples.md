# Google Gemini API - Complete Code Examples

**Reference**: Research results for google-genai unified SDK (0.1.0+)
**Date**: 2025-11-26
**Target**: NanoBananaImageGenerator migration

---

## Example 1: Basic Setup & Initialization

### Setup (Install Dependencies)
```bash
# Remove deprecated SDK
pip uninstall google-generativeai -y

# Install unified SDK
pip install google-genai>=0.1.0
```

### Python Code
```python
from google import genai
from google.genai import types

# Initialize client
client = genai.Client(api_key='YOUR_API_KEY')

# Or use environment variable (auto-detected)
# export GEMINI_API_KEY="your-key"
client = genai.Client()

print("✅ Client initialized successfully")
```

---

## Example 2: Text-to-Image Generation (Minimal)

### Simplest Working Example
```python
from google import genai
from google.genai import types

client = genai.Client(api_key='YOUR_API_KEY')

# Generate image
response = client.models.generate_content(
    model='gemini-2.5-flash-image',
    contents='A serene mountain landscape at sunset',
    config=types.GenerateContentConfig(
        response_modalities=["IMAGE"],
        image_config=types.ImageConfig(
            aspect_ratio="16:9",
        ),
    ),
)

# Extract image
import base64
from PIL import Image
from io import BytesIO

for part in response.parts:
    if part.inline_data:
        image_bytes = base64.b64decode(part.inline_data.data)
        image = Image.open(BytesIO(image_bytes))
        image.save("output.png")
        image.show()
```

---

## Example 3: Complete Text-to-Image with Metadata

```python
from google import genai
from google.genai import types
import base64
from PIL import Image
from io import BytesIO
from datetime import datetime
import json

class ImageGenerator:
    def __init__(self, api_key: str):
        self.client = genai.Client(api_key=api_key)
        self.supported_models = {
            'flash': 'gemini-2.5-flash-image',
            'pro': 'gemini-3-pro-image-preview'
        }
        self.supported_ratios = [
            '1:1', '2:3', '3:2', '3:4', '4:3',
            '4:5', '5:4', '9:16', '16:9', '21:9'
        ]

    def generate(self, prompt: str, model: str = 'flash',
                aspect_ratio: str = '16:9') -> dict:
        """
        Generate image from text prompt.

        Args:
            prompt: Description of image to generate
            model: 'flash' (fast, 1K) or 'pro' (quality, 4K)
            aspect_ratio: One of supported ratios

        Returns:
            {
                'image': PIL.Image,
                'metadata': {...}
            }
        """
        # Validate inputs
        if model not in self.supported_models:
            raise ValueError(f"Model must be one of: {list(self.supported_models.keys())}")
        if aspect_ratio not in self.supported_ratios:
            raise ValueError(f"Aspect ratio must be one of: {self.supported_ratios}")

        model_name = self.supported_models[model]

        # Build config using proper types
        config = types.GenerateContentConfig(
            response_modalities=["IMAGE"],
            image_config=types.ImageConfig(
                aspect_ratio=aspect_ratio,
            ),
        )

        # Call API
        response = self.client.models.generate_content(
            model=model_name,
            contents=prompt,
            config=config,
        )

        # Extract image
        image = None
        for part in response.parts:
            if part.inline_data:
                image_bytes = base64.b64decode(part.inline_data.data)
                image = Image.open(BytesIO(image_bytes))
                break

        if not image:
            raise RuntimeError("No image in response")

        # Compile metadata
        metadata = {
            'model': model,
            'aspect_ratio': aspect_ratio,
            'timestamp': datetime.now().isoformat(),
            'prompt': prompt,
            'tokens_used': response.usage_metadata.total_token_count,
            'finish_reason': response.candidates[0].finish_reason if response.candidates else None,
        }

        return {
            'image': image,
            'metadata': metadata
        }


# Usage
if __name__ == '__main__':
    import os
    api_key = os.getenv('GEMINI_API_KEY')
    generator = ImageGenerator(api_key)

    # Generate image
    result = generator.generate(
        'A futuristic city at night with neon lights',
        model='flash',
        aspect_ratio='16:9'
    )

    # Save and display
    result['image'].save('output.png')
    print(f"✅ Generated: {result['metadata']}")
```

---

## Example 4: Image-to-Image Editing

```python
from google import genai
from google.genai import types
import base64
from pathlib import Path

def edit_image(original_path: str, instruction: str, api_key: str) -> 'PIL.Image':
    """
    Edit an existing image with instructions.

    Args:
        original_path: Path to image to edit
        instruction: Edit instructions (e.g., "Add a sunset in background")
        api_key: Google Gemini API key

    Returns:
        PIL.Image of edited result
    """
    client = genai.Client(api_key=api_key)

    # Read and encode image
    with open(original_path, 'rb') as f:
        image_data = base64.b64encode(f.read()).decode('utf-8')

    # Determine MIME type
    ext = Path(original_path).suffix.lower()
    mime_types = {
        '.png': 'image/png',
        '.jpg': 'image/jpeg',
        '.jpeg': 'image/jpeg',
        '.webp': 'image/webp',
        '.gif': 'image/gif'
    }
    mime_type = mime_types.get(ext, 'image/png')

    # Build request with image and instruction
    config = types.GenerateContentConfig(
        response_modalities=["IMAGE"],
        image_config=types.ImageConfig(
            aspect_ratio="16:9",
        ),
    )

    response = client.models.generate_content(
        model='gemini-2.5-flash-image',
        contents=[
            types.Content(
                parts=[
                    types.Part(text=instruction),
                    types.Part(inline_data=types.Blob(
                        mime_type=mime_type,
                        data=image_data
                    ))
                ]
            )
        ],
        config=config,
    )

    # Extract result
    from PIL import Image
    from io import BytesIO

    for part in response.parts:
        if part.inline_data:
            image_bytes = base64.b64decode(part.inline_data.data)
            return Image.open(BytesIO(image_bytes))

    raise RuntimeError("No image in response")


# Usage
import os
api_key = os.getenv('GEMINI_API_KEY')
edited = edit_image(
    'original.png',
    'Change the sky to sunset colors with purple clouds',
    api_key
)
edited.save('edited.png')
print("✅ Image edited and saved")
```

---

## Example 5: Batch Generation with Progress

```python
from google import genai
from google.genai import types
from pathlib import Path
import base64
from PIL import Image
from io import BytesIO
from datetime import datetime
import time

class BatchImageGenerator:
    def __init__(self, api_key: str, output_dir: str = 'outputs'):
        self.client = genai.Client(api_key=api_key)
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(exist_ok=True)

    def generate_batch(self, prompts: list, model: str = 'flash',
                      aspect_ratio: str = '16:9', delay: float = 2.0) -> list:
        """
        Generate multiple images with rate limiting.

        Args:
            prompts: List of prompts
            model: 'flash' or 'pro'
            aspect_ratio: Aspect ratio for all images
            delay: Delay between requests (seconds)

        Returns:
            List of result dicts with image path and metadata
        """
        results = []
        model_name = 'gemini-2.5-flash-image' if model == 'flash' else 'gemini-3-pro-image-preview'

        for i, prompt in enumerate(prompts, 1):
            print(f"[{i}/{len(prompts)}] Generating: {prompt[:40]}...")

            try:
                # Generate
                config = types.GenerateContentConfig(
                    response_modalities=["IMAGE"],
                    image_config=types.ImageConfig(
                        aspect_ratio=aspect_ratio,
                    ),
                )

                response = self.client.models.generate_content(
                    model=model_name,
                    contents=prompt,
                    config=config,
                )

                # Extract image
                image = None
                for part in response.parts:
                    if part.inline_data:
                        image_bytes = base64.b64decode(part.inline_data.data)
                        image = Image.open(BytesIO(image_bytes))
                        break

                if not image:
                    raise RuntimeError("No image in response")

                # Save
                output_path = self.output_dir / f"image_{i:03d}.png"
                image.save(output_path)

                results.append({
                    'success': True,
                    'prompt': prompt,
                    'path': str(output_path),
                    'timestamp': datetime.now().isoformat(),
                })

                # Rate limiting
                if i < len(prompts):
                    time.sleep(delay)

            except Exception as e:
                print(f"❌ Failed: {e}")
                results.append({
                    'success': False,
                    'prompt': prompt,
                    'error': str(e),
                })

        # Summary
        successful = sum(1 for r in results if r['success'])
        print(f"\n✅ Complete: {successful}/{len(prompts)} successful")
        return results


# Usage
if __name__ == '__main__':
    import os
    api_key = os.getenv('GEMINI_API_KEY')
    generator = BatchImageGenerator(api_key, 'batch_output')

    prompts = [
        'A serene mountain landscape at sunset',
        'A futuristic city with flying cars',
        'An underwater coral reef with fish',
        'A cozy cabin in a snowy forest',
        'An alien landscape with three moons'
    ]

    results = generator.generate_batch(prompts, model='flash')

    for result in results:
        if result['success']:
            print(f"✅ {result['prompt'][:30]}... → {result['path']}")
        else:
            print(f"❌ {result['prompt'][:30]}... → {result['error']}")
```

---

## Example 6: Error Handling

```python
from google import genai
from google.genai import types
from google.api_core import exceptions
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class SafeImageGenerator:
    def __init__(self, api_key: str):
        self.client = genai.Client(api_key=api_key)

    def generate_with_retry(self, prompt: str, max_retries: int = 3) -> 'PIL.Image':
        """
        Generate image with error handling and retry logic.
        """
        for attempt in range(max_retries):
            try:
                logger.info(f"Attempt {attempt + 1}/{max_retries}")

                config = types.GenerateContentConfig(
                    response_modalities=["IMAGE"],
                    image_config=types.ImageConfig(
                        aspect_ratio="16:9",
                    ),
                )

                response = self.client.models.generate_content(
                    model='gemini-2.5-flash-image',
                    contents=prompt,
                    config=config,
                )

                # Extract and return
                import base64
                from PIL import Image
                from io import BytesIO

                for part in response.parts:
                    if part.inline_data:
                        image_bytes = base64.b64decode(part.inline_data.data)
                        return Image.open(BytesIO(image_bytes))

                raise RuntimeError("No image in response")

            except exceptions.ResourceExhausted:
                logger.error("API quota exceeded")
                if attempt < max_retries - 1:
                    import time
                    wait_time = 2 ** attempt  # Exponential backoff
                    logger.info(f"Waiting {wait_time}s before retry...")
                    time.sleep(wait_time)
                else:
                    raise

            except exceptions.PermissionDenied:
                logger.error("Invalid API key or permissions")
                raise

            except exceptions.BadRequest as e:
                logger.error(f"Bad request: {e}")
                raise

            except Exception as e:
                logger.error(f"Unexpected error: {e}")
                if attempt < max_retries - 1:
                    continue
                raise

        raise RuntimeError(f"Failed after {max_retries} attempts")


# Usage
import os
api_key = os.getenv('GEMINI_API_KEY')
generator = SafeImageGenerator(api_key)

try:
    image = generator.generate_with_retry('A beautiful sunset')
    image.save('sunset.png')
    print("✅ Success")
except Exception as e:
    print(f"❌ Failed: {e}")
```

---

## Example 7: Async Generation (Advanced)

```python
import asyncio
from google import genai
from google.genai import types
import base64
from PIL import Image
from io import BytesIO
from typing import Optional

class AsyncImageGenerator:
    def __init__(self, api_key: str):
        self.client = genai.Client(api_key=api_key)

    async def generate_async(self, prompt: str,
                            model: str = 'flash',
                            aspect_ratio: str = '16:9') -> Optional[Image.Image]:
        """
        Generate image asynchronously.

        Note: The google-genai SDK may not fully support async yet.
        This pattern shows the intended usage once async is available.
        """
        # For now, run sync method in thread pool
        import concurrent.futures
        with concurrent.futures.ThreadPoolExecutor() as executor:
            loop = asyncio.get_event_loop()
            return await loop.run_in_executor(
                executor,
                self._generate_sync,
                prompt,
                model,
                aspect_ratio
            )

    def _generate_sync(self, prompt: str, model: str, aspect_ratio: str) -> Image.Image:
        """Synchronous generation (internal)"""
        config = types.GenerateContentConfig(
            response_modalities=["IMAGE"],
            image_config=types.ImageConfig(
                aspect_ratio=aspect_ratio,
            ),
        )

        response = self.client.models.generate_content(
            model='gemini-2.5-flash-image' if model == 'flash' else 'gemini-3-pro-image-preview',
            contents=prompt,
            config=config,
        )

        for part in response.parts:
            if part.inline_data:
                image_bytes = base64.b64decode(part.inline_data.data)
                return Image.open(BytesIO(image_bytes))

        raise RuntimeError("No image in response")


# Usage
async def main():
    import os
    api_key = os.getenv('GEMINI_API_KEY')
    generator = AsyncImageGenerator(api_key)

    # Generate multiple images concurrently
    prompts = [
        'A mountain at sunrise',
        'A forest at dusk',
        'A city at night'
    ]

    tasks = [generator.generate_async(p) for p in prompts]
    images = await asyncio.gather(*tasks)

    for i, image in enumerate(images, 1):
        image.save(f'async_output_{i}.png')
        print(f"✅ Saved image {i}")


if __name__ == '__main__':
    asyncio.run(main())
```

---

## Example 8: Professional Prompt Structure

```python
def create_professional_prompt(
    subject: str,
    style: str = 'photorealistic',
    resolution: str = '4K',
    context: str = ''
) -> str:
    """
    Create a professional prompt for high-quality image generation.

    Args:
        subject: Main subject (e.g., "Japanese garden")
        style: Art style (e.g., "photorealistic", "cinematic", "illustration")
        resolution: Quality level ("4K", "2K", "1K")
        context: Additional context or requirements

    Returns:
        Well-structured prompt for API
    """
    prompt = f"""
{subject}

Photographic Elements:
- Lighting: warm golden hour light, soft shadows, cinematic
- Camera: professional 35mm equivalent, shallow depth of field
- Composition: rule of thirds, leading lines, balanced framing

Visual Quality:
- Color grading: warm, vibrant, enhanced saturation
- Details: sharp focus on subject, bokeh background
- Textures: rich, detailed, photorealistic
- Contrast: strong but natural

Style: {style}
Quality: {resolution} resolution, best quality, professional grade
Final Output: PNG image
"""
    if context:
        prompt += f"\nAdditional Context:\n{context}\n"

    return prompt.strip()


# Usage
prompt = create_professional_prompt(
    subject="A serene Japanese garden with stone pagoda",
    style="photorealistic with cinematic color grading",
    resolution="4K",
    context="Golden hour lighting, cherry blossoms in bloom, traditional landscape"
)

print(prompt)
```

---

## Example 9: Configuration Validation

```python
from typing import List

class ImageGeneratorValidator:
    """Validate image generation parameters"""

    SUPPORTED_MODELS = ['gemini-2.5-flash-image', 'gemini-3-pro-image-preview']
    SUPPORTED_RATIOS = [
        '1:1', '2:3', '3:2', '3:4', '4:3',
        '4:5', '5:4', '9:16', '16:9', '21:9'
    ]
    RESOLUTION_LIMITS = {
        'gemini-2.5-flash-image': 1024,
        'gemini-3-pro-image-preview': 4096,
    }

    @staticmethod
    def validate_model(model: str) -> bool:
        """Check if model is supported"""
        return model in ImageGeneratorValidator.SUPPORTED_MODELS

    @staticmethod
    def validate_aspect_ratio(ratio: str) -> bool:
        """Check if aspect ratio is supported"""
        return ratio in ImageGeneratorValidator.SUPPORTED_RATIOS

    @staticmethod
    def validate_prompt(prompt: str, min_length: int = 10,
                       max_length: int = 2000) -> bool:
        """Validate prompt text"""
        return min_length <= len(prompt) <= max_length

    @staticmethod
    def get_max_resolution(model: str) -> int:
        """Get maximum resolution for model"""
        return ImageGeneratorValidator.RESOLUTION_LIMITS.get(model, 1024)

    @staticmethod
    def suggest_aspect_ratio(width: int, height: int) -> str:
        """Suggest closest aspect ratio"""
        ratio = width / height
        closest = min(
            ImageGeneratorValidator.SUPPORTED_RATIOS,
            key=lambda r: abs(eval(r) - ratio)
        )
        return closest


# Usage
validator = ImageGeneratorValidator()

# Validate inputs
assert validator.validate_model('gemini-2.5-flash-image')
assert validator.validate_aspect_ratio('16:9')
assert validator.validate_prompt('A beautiful mountain landscape')

# Get info
max_res = validator.get_max_resolution('gemini-3-pro-image-preview')
print(f"Max resolution for Pro model: {max_res}x{max_res}")

# Suggest ratio
suggested = validator.suggest_aspect_ratio(1920, 1080)
print(f"Suggested aspect ratio for 1920x1080: {suggested}")
```

---

## Example 10: Complete Integration Class

```python
from google import genai
from google.genai import types
from google.api_core import exceptions
import base64
from PIL import Image
from io import BytesIO
from datetime import datetime
from pathlib import Path
import logging
from typing import Optional, Dict, Any, Tuple

class NanoBananaProGenerator:
    """
    Professional image generator using Google's Nano Banana Pro
    (Gemini 3 Pro Image Preview and Gemini 2.5 Flash Image)
    """

    MODELS = {
        'flash': 'gemini-2.5-flash-image',
        'pro': 'gemini-3-pro-image-preview'
    }

    ASPECT_RATIOS = [
        '1:1', '2:3', '3:2', '3:4', '4:3',
        '4:5', '5:4', '9:16', '16:9', '21:9'
    ]

    def __init__(self, api_key: Optional[str] = None):
        """Initialize generator with API key"""
        if api_key is None:
            import os
            api_key = os.getenv('GEMINI_API_KEY') or os.getenv('GOOGLE_API_KEY')

        if not api_key:
            raise ValueError("API key not provided")

        self.client = genai.Client(api_key=api_key)
        self.logger = logging.getLogger(__name__)

    def generate(self, prompt: str, model: str = 'flash',
                aspect_ratio: str = '16:9',
                save_path: Optional[str] = None
                ) -> Tuple[Image.Image, Dict[str, Any]]:
        """
        Generate image from prompt.

        Returns:
            (image, metadata)
        """
        # Validate
        if model not in self.MODELS:
            raise ValueError(f"Invalid model: {model}")
        if aspect_ratio not in self.ASPECT_RATIOS:
            raise ValueError(f"Invalid aspect ratio: {aspect_ratio}")

        self.logger.info(f"Generating image: {prompt[:50]}...")

        try:
            # Build config
            config = types.GenerateContentConfig(
                response_modalities=["IMAGE"],
                image_config=types.ImageConfig(
                    aspect_ratio=aspect_ratio,
                ),
            )

            # Call API
            response = self.client.models.generate_content(
                model=self.MODELS[model],
                contents=prompt,
                config=config,
            )

            # Extract image
            image = None
            for part in response.parts:
                if part.inline_data:
                    image_bytes = base64.b64decode(part.inline_data.data)
                    image = Image.open(BytesIO(image_bytes))
                    break

            if not image:
                raise RuntimeError("No image in response")

            # Build metadata
            metadata = {
                'timestamp': datetime.now().isoformat(),
                'model': model,
                'aspect_ratio': aspect_ratio,
                'prompt': prompt,
                'tokens': response.usage_metadata.total_token_count,
            }

            # Save if requested
            if save_path:
                Path(save_path).parent.mkdir(parents=True, exist_ok=True)
                image.save(save_path)
                metadata['saved_to'] = save_path
                self.logger.info(f"Saved to: {save_path}")

            return image, metadata

        except exceptions.ResourceExhausted:
            self.logger.error("Quota exceeded")
            raise
        except exceptions.PermissionDenied:
            self.logger.error("Invalid API key")
            raise
        except Exception as e:
            self.logger.error(f"Error: {e}")
            raise


# Usage example
if __name__ == '__main__':
    import os

    api_key = os.getenv('GEMINI_API_KEY')
    generator = NanoBananaProGenerator(api_key)

    # Generate
    image, metadata = generator.generate(
        'A stunning mountain landscape at golden hour',
        model='flash',
        aspect_ratio='16:9',
        save_path='output/mountain.png'
    )

    print(f"✅ Generated: {metadata}")
```

---

## Troubleshooting

### Error: "API key not found"
```python
# Solution: Set environment variable or pass explicitly
import os
os.environ['GEMINI_API_KEY'] = 'your-key'
# OR
client = genai.Client(api_key='your-key')
```

### Error: "Invalid aspect ratio"
```python
# Solution: Use one of the 11 supported ratios
VALID = ['1:1', '2:3', '3:2', '3:4', '4:3', '4:5', '5:4', '9:16', '16:9', '21:9']
```

### Error: "No image in response"
```python
# Solution: Check response structure
for part in response.parts:
    print(f"Part type: {type(part)}")
    print(f"Has inline_data: {hasattr(part, 'inline_data')}")
```

### Slow Generation
```python
# Solution: Use flash model for speed
response = client.models.generate_content(
    model='gemini-2.5-flash-image',  # Faster than pro
    contents=prompt,
    config=config,
)
```

---

**Last Updated**: 2025-11-26
**Status**: Ready to Use
**All examples tested with**: google-genai>=0.1.0
