# Nano Banana Pro Image Generation Examples

## Example 1: Basic Text-to-Image Generation

```python
import google.generativeai as genai
from PIL import Image
import io

def generate_product_image():
    """Generate professional product image."""
    client = genai.Client(api_key="YOUR_API_KEY")

    prompt = """
    Professional product photo of a sleek black wireless headphone.
    Studio lighting, white background, 3/4 view angle.
    High quality, commercial photography style.
    4K resolution, sharp focus on product details.
    """

    response = client.models.generate_images(
        model="gemini-3-pro-image-preview",
        prompt=prompt,
        number_of_images=1,
        height=2048,
        width=2048,
        safety_filter_level="block_medium_and_above"
    )

    # Save generated image
    image_data = response.images[0]
    img = Image.open(io.BytesIO(image_data))
    img.save("product_image.png")
    print("Generated product image: product_image.png")

# Usage
generate_product_image()
```

---

## Example 2: Image-to-Image Style Transfer

```python
def apply_art_style():
    """Apply artistic style to existing image."""
    client = genai.Client()

    # Load reference image
    with open("reference_image.jpg", "rb") as f:
        image_data = f.read()

    prompt = """
    Transform this image to look like an oil painting in Van Gogh style.
    Keep all original elements and composition.
    Use thick brushstrokes and vibrant colors.
    Maintain the subject but add artistic texture throughout.
    """

    response = client.models.generate_images(
        model="gemini-3-pro-image-preview",
        prompt=prompt,
        input_image=image_data,
        image_operation="style_transfer"
    )

    # Save styled image
    styled = Image.open(io.BytesIO(response.images[0]))
    styled.save("styled_image.png")
    print("Applied art style: styled_image.png")

# Usage
apply_art_style()
```

---

## Example 3: Text Overlay on Image

```python
def add_text_to_image():
    """Add professional text overlay."""
    client = genai.Client()

    prompt = """
    Create a social media post image with:
    - Background: Modern gradient (blue to purple)
    - Main text: "Launch Your Career" (bold, white, 48pt)
    - Subtitle: "Start Your Journey Today" (light gray, 24pt)
    - Include subtle decorative elements
    - Professional modern design, clean layout
    - 1200x630px size for social media
    """

    response = client.models.generate_images(
        model="gemini-3-pro-image-preview",
        prompt=prompt,
        width=1200,
        height=630
    )

    img = Image.open(io.BytesIO(response.images[0]))
    img.save("social_post.png")
    print("Generated social post: social_post.png")

# Usage
add_text_to_image()
```

---

## Example 4: Object Removal and Inpainting

```python
def remove_object_from_image():
    """Remove unwanted object from image."""
    client = genai.Client()

    with open("landscape.jpg", "rb") as f:
        image_data = f.read()

    prompt = """
    Remove the telephone pole from this landscape image.
    Keep everything else exactly as is.
    Fill the background naturally with sky and scenery.
    Make the removal invisible and seamless.
    """

    response = client.models.generate_images(
        model="gemini-3-pro-image-preview",
        prompt=prompt,
        input_image=image_data,
        image_operation="inpaint",
        safety_filter_level="block_low_and_above"
    )

    edited = Image.open(io.BytesIO(response.images[0]))
    edited.save("landscape_edited.png")
    print("Object removed: landscape_edited.png")

# Usage
remove_object_from_image()
```

---

## Example 5: Real-Time Grounding with Google Search

```python
def generate_current_event_image():
    """Generate image with real-time information."""
    client = genai.Client()

    prompt = """
    Create an infographic about the latest AI breakthroughs (real-time search).
    Include key statistics and innovations.
    Professional modern design with tech aesthetic.
    Use data from current sources for accuracy.
    """

    response = client.models.generate_images(
        model="gemini-3-pro-image-preview",
        prompt=prompt,
        enable_search=True,  # Enable real-time Google Search
        width=1920,
        height=1080
    )

    img = Image.open(io.BytesIO(response.images[0]))
    img.save("infographic.png")
    print("Generated infographic: infographic.png")

# Usage
generate_current_event_image()
```

---

## Example 6: Multi-Image Reference Composition

```python
def create_composite_image():
    """Create image using multiple references."""
    client = genai.Client()

    # Load reference images
    with open("reference1.jpg", "rb") as f:
        ref1 = f.read()
    with open("reference2.jpg", "rb") as f:
        ref2 = f.read()

    prompt = """
    Combine elements from both reference images into one harmonious composition.
    Use color palette from reference 1 and composition style from reference 2.
    Create a seamless, professional result.
    4K resolution with high quality.
    """

    response = client.models.generate_images(
        model="gemini-3-pro-image-preview",
        prompt=prompt,
        input_images=[ref1, ref2],
        image_operation="compose"
    )

    composite = Image.open(io.BytesIO(response.images[0]))
    composite.save("composite.png")
    print("Created composite: composite.png")

# Usage
create_composite_image()
```

---

## Example 7: Batch Image Generation

```python
async def batch_generate_variations():
    """Generate multiple variations efficiently."""
    client = genai.Client()

    base_prompt = "Professional portrait of a {role} in modern office setting"
    roles = ["engineer", "designer", "product manager", "data scientist"]

    for i, role in enumerate(roles):
        prompt = base_prompt.format(role=role)

        response = client.models.generate_images(
            model="gemini-3-pro-image-preview",
            prompt=prompt,
            number_of_images=1,
            width=1024,
            height=1024
        )

        img = Image.open(io.BytesIO(response.images[0]))
        img.save(f"portrait_{role}.png")
        print(f"Generated {i+1}/4: {role} portrait")

# Usage
import asyncio
asyncio.run(batch_generate_variations())
```

---

## Example 8: Image Quality Optimization

```python
def generate_hires_image():
    """Generate high-quality image with optimization."""
    client = genai.Client()

    prompt = """
    Ultra high-quality product photograph of luxury watch.
    4K resolution, perfect lighting, sharp focus.
    Professional jewelry photography style.
    Show intricate details and craftsmanship.
    """

    # Generate at 2K first
    response = client.models.generate_images(
        model="gemini-3-pro-image-preview",
        prompt=prompt,
        width=2048,
        height=2048,
        quality="high"  # High quality setting
    )

    # Save original
    img = Image.open(io.BytesIO(response.images[0]))
    img.save("watch_2k.png")

    # Upscale using image enhancement
    response_upscale = client.models.generate_images(
        model="gemini-3-pro-image-preview",
        prompt="Upscale and enhance this image to 4K with improved details",
        input_image=response.images[0],
        image_operation="upscale",
        width=4096,
        height=4096
    )

    img_4k = Image.open(io.BytesIO(response_upscale.images[0]))
    img_4k.save("watch_4k.png")
    print("Generated: watch_4k.png (4K resolution)")

# Usage
generate_hires_image()
```

---

## Example 9: Batch Processing with Error Handling

```python
async def batch_process_with_errors():
    """Process batch with comprehensive error handling."""
    client = genai.Client()

    prompts = [
        "Modern office building architecture",
        "Serene forest landscape",
        "Abstract digital art",
        "Professional team photo"
    ]

    results = {
        "success": [],
        "failed": []
    }

    for i, prompt in enumerate(prompts, 1):
        try:
            response = client.models.generate_images(
                model="gemini-3-pro-image-preview",
                prompt=prompt,
                width=1024,
                height=1024,
                timeout=60  # 60 second timeout
            )

            img = Image.open(io.BytesIO(response.images[0]))
            filename = f"image_{i}.png"
            img.save(filename)

            results["success"].append({
                "prompt": prompt,
                "file": filename
            })
            print(f"✓ Generated {i}/{len(prompts)}: {filename}")

        except Exception as e:
            results["failed"].append({
                "prompt": prompt,
                "error": str(e)
            })
            print(f"✗ Failed {i}/{len(prompts)}: {e}")

    return results

# Usage
import asyncio
results = asyncio.run(batch_process_with_errors())
print(f"\nSummary: {len(results['success'])} success, {len(results['failed'])} failed")
```

---

## Example 10: Production Workflow Integration

```python
class NanoBananaImagePipeline:
    """Production-ready image generation pipeline."""

    def __init__(self, api_key: str):
        self.client = genai.Client(api_key=api_key)

    async def generate_marketing_assets(self, campaign_name: str):
        """Generate complete marketing asset set."""
        assets = {}

        # Social media posts
        social_prompts = {
            "instagram": "Instagram post design for {campaign}",
            "twitter": "Twitter header image for {campaign}",
            "linkedin": "LinkedIn article image for {campaign}"
        }

        for platform, prompt_template in social_prompts.items():
            prompt = prompt_template.format(campaign=campaign_name)
            response = await self._generate_with_retry(prompt)
            assets[platform] = self._save_asset(response, f"asset_{platform}.png")

        return assets

    async def _generate_with_retry(self, prompt: str, max_retries: int = 3):
        """Generate with exponential backoff retry."""
        for attempt in range(max_retries):
            try:
                return self.client.models.generate_images(
                    model="gemini-3-pro-image-preview",
                    prompt=prompt,
                    width=1024,
                    height=1024
                )
            except Exception as e:
                if attempt == max_retries - 1:
                    raise
                await asyncio.sleep(2 ** attempt)

    def _save_asset(self, response, filename: str) -> str:
        """Save generated image."""
        img = Image.open(io.BytesIO(response.images[0]))
        img.save(filename)
        return filename

# Usage
pipeline = NanoBananaImagePipeline("YOUR_API_KEY")
assets = asyncio.run(pipeline.generate_marketing_assets("Q4 Launch"))
print(f"Generated assets: {assets}")
```

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready
