# Nano Banana Pro API Reference

## Authentication Setup

```python
import google.generativeai as genai

# Initialize with API key
genai.configure(api_key="YOUR_GOOGLE_API_KEY")

# Or use environment variable
import os
api_key = os.getenv("GOOGLE_API_KEY")
genai.configure(api_key=api_key)
```

---

## Core Models

| Model | Resolution | Speed | Quality | Use Case |
|-------|-----------|-------|---------|----------|
| **gemini-3-pro-image-preview** | Up to 4K | Fast | Professional | Production images |
| **gemini-2-flash** | Up to 2K | Very Fast | High | Quick iterations |
| **gemini-pro** | Up to 1K | Standard | Good | General purpose |

---

## API Methods

### Generate Images

```python
response = client.models.generate_images(
    model="gemini-3-pro-image-preview",
    prompt="Your prompt here",
    number_of_images=1,
    width=1024,
    height=1024,
    quality="high",
    safety_filter_level="block_medium_and_above"
)

# Access generated image
image_data = response.images[0]
```

### Edit Image

```python
response = client.models.generate_images(
    model="gemini-3-pro-image-preview",
    prompt="Edit prompt",
    input_image=image_bytes,
    image_operation="edit",  # edit, inpaint, style_transfer
    width=1024,
    height=1024
)
```

### Compose Images

```python
response = client.models.generate_images(
    model="gemini-3-pro-image-preview",
    prompt="Composition prompt",
    input_images=[image1_bytes, image2_bytes],
    image_operation="compose"
)
```

---

## Configuration Parameters

| Parameter | Values | Default | Notes |
|-----------|--------|---------|-------|
| **width** | 512-4096 | 1024 | Must match height aspect ratio |
| **height** | 512-4096 | 1024 | Must match width aspect ratio |
| **quality** | standard, high | standard | Higher quality = slower |
| **number_of_images** | 1-4 | 1 | Max 4 per request |
| **enable_search** | true, false | false | Use real-time Google Search |

---

## Safety Filters

```
block_none           # No filtering
block_low           # Block low probability unsafe content
block_medium        # Block medium probability (recommended)
block_medium_and_above  # Strict filtering (default)
block_high_and_above    # Very strict filtering
```

---

## Image Operations

```
generate      # Create image from prompt
edit          # Modify existing image
inpaint       # Remove and fill objects
style_transfer    # Apply artistic style
upscale       # Increase resolution
compose       # Combine multiple images
```

---

## Error Handling

```python
class ImageGenerationError(Exception):
    pass

try:
    response = client.models.generate_images(...)
except genai.types.ClientError as e:
    if "rate limit" in str(e):
        # Handle rate limiting
        await asyncio.sleep(2)
    else:
        raise ImageGenerationError(f"Generation failed: {e}")
```

---

## Common Response Structure

```python
{
    "images": [
        <PIL.Image.Image>,  # Generated image
        ...
    ],
    "metadata": {
        "generation_time_ms": 2340,
        "model": "gemini-3-pro-image-preview",
        "safety_rating": "safe"
    }
}
```

---

**Version**: 4.0.0 | **Last Updated**: 2025-11-22
