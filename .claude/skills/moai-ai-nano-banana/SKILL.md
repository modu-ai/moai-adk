---
name: moai-ai-nano-banana
description: Nano-Banana AI service integration for content generation, image creation, and AI-powered workflows. Use when integrating AI services for content creation.
version: 1.1.0
category: integration
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
tags:
  - ai
  - content-generation
  - image-generation
  - nano-banana
  - ai-service
  - gemini-3-pro
related-skills:
  - moai-docs-generation
  - moai-domain-uiux
updated: 2025-12-23
status: active
author: MoAI-ADK Team
---

# Nano-Banana AI Service Integration

## Quick Reference (30 seconds)

Nano Banana Pro (gemini-3-pro-image-preview) integration for high-quality image generation with 1K/2K/4K resolution support, style prefixes, and batch processing.

Core Capabilities:
- Image Generation: 1K, 2K, 4K resolution with 10 aspect ratios
- Style Control: Style prefix support for consistent aesthetics
- Batch Processing: JSON/YAML config with concurrent generation
- Smart Retry: Exponential backoff for API quota handling
- Google Search: Grounded generation for factual content

CLI Quick Start:
```bash
# Single image generation
python scripts/generate_image.py -p "Dashboard UI design" -o ui.png -r 4K

# With style prefix
python scripts/generate_image.py -p "Logo" -o logo.png --style "minimalist, vector"

# Batch generation
python scripts/batch_generate.py -c prompts.json -d output/ --concurrency 2
```

When to Use:
- Generating high-quality images for documentation
- Creating visual assets with consistent styling
- Batch processing multiple image prompts
- Building automated image generation pipelines

---

## Implementation Guide (5 minutes)

### CLI Scripts (Recommended)

Single Image Generation (scripts/generate_image.py):
```bash
# Basic usage
python scripts/generate_image.py -p "A fluffy cat" -o cat.png

# High resolution with aspect ratio
python scripts/generate_image.py -p "Dashboard UI" -o ui.png -r 4K -a 16:9

# With style prefix for consistent aesthetics
python scripts/generate_image.py -p "Mountain landscape" -o landscape.png \
    --style "photorealistic, dramatic lighting"

# With Google Search grounding
python scripts/generate_image.py -p "Mount Fuji at sunset" -o fuji.png -g

# Verbose mode
python scripts/generate_image.py -p "Tech logo" -o logo.png -r 4K -a 1:1 \
    --style "minimalist, vector art" -v
```

Batch Generation (scripts/batch_generate.py):
```bash
# From JSON config
python scripts/batch_generate.py -c prompts.json -d output/

# From YAML config with concurrency
python scripts/batch_generate.py -c prompts.yaml -d output/ --concurrency 3

# From command line prompts
python scripts/batch_generate.py --prompts "Cat" "Dog" "Bird" -d output/ \
    --style "watercolor painting"

# With report generation
python scripts/batch_generate.py -c prompts.json -d output/ --report report.json
```

Batch Config File Example (prompts.json):
```json
{
  "defaults": {
    "style": "photorealistic, high detail",
    "resolution": "4K",
    "aspect_ratio": "16:9"
  },
  "images": [
    "A serene mountain landscape",
    {
      "prompt": "Modern tech company logo",
      "filename": "logo.png",
      "resolution": "4K",
      "aspect_ratio": "1:1",
      "style": "minimalist, vector art"
    }
  ]
}
```

### Python API Usage

Direct API Integration:
```python
from google import genai
from google.genai import types
import os

client = genai.Client(api_key=os.getenv('GOOGLE_API_KEY'))

response = client.models.generate_content(
    model="gemini-3-pro-image-preview",
    contents="Modern dashboard UI with dark theme",
    config=types.GenerateContentConfig(
        response_modalities=['TEXT', 'IMAGE'],
        image_config=types.ImageConfig(
            aspect_ratio="16:9",
            image_size="4K"  # 1K, 2K, 4K (uppercase K required)
        )
    )
)

for part in response.candidates[0].content.parts:
    if part.inline_data is not None:
        with open("output.png", "wb") as f:
            f.write(part.inline_data.data)
```

### Model Specifications

Model: gemini-3-pro-image-preview (Nano Banana Pro)

Supported Resolutions:
- 1K: Standard quality, fast generation
- 2K: Balanced quality and speed (default)
- 4K: Highest quality, detailed output

Supported Aspect Ratios:
- Square: 1:1
- Portrait: 2:3, 3:4, 4:5, 9:16
- Landscape: 3:2, 4:3, 5:4, 16:9, 21:9

Special Features:
- Thinking mode enabled by default
- Up to 14 reference images for mixing
- Advanced text rendering for logos
- Google Search grounding option

### Style Prefix Examples

Common Style Prefixes:
- "photorealistic, 8K, detailed" - Photo-like quality
- "minimalist, vector art, clean" - Simple graphics
- "watercolor painting, soft colors" - Artistic style
- "3D render, cinematic lighting" - 3D appearance
- "anime style, vibrant colors" - Animated look
- "professional photography, studio lighting" - Product shots
- "digital art, fantasy, ethereal" - Creative illustrations
- "infographic, data visualization" - Charts and diagrams

### Error Handling

The scripts include automatic retry with exponential backoff:
- Maximum 5 retry attempts (configurable)
- Base delay: 2 seconds, doubling each attempt
- Maximum delay: 60-120 seconds
- Jitter added for quota errors to avoid thundering herd

Non-retryable errors (immediate failure):
- Invalid API key
- Invalid argument (wrong aspect ratio/resolution)

---

## Advanced Patterns (10+ minutes)

### Programmatic Usage

Using generate_image.py as a module:
```python
from scripts.generate_image import generate_image

# Single image with all options
result = generate_image(
    prompt="Modern tech startup office",
    output_path="office.png",
    aspect_ratio="16:9",
    resolution="4K",
    style="photorealistic, natural lighting",
    enable_grounding=False,
    max_retries=5,
    verbose=True
)

if result["success"]:
    print(f"Generated: {result['output_path']}")
    print(f"Time: {result['generation_time_seconds']}s")
```

Using batch_generate.py programmatically:
```python
import asyncio
from scripts.batch_generate import (
    ImageTask, run_batch_generation, save_batch_report
)

# Create tasks
tasks = [
    ImageTask(
        prompt="Dashboard UI",
        output_path="output/dashboard.png",
        resolution="4K",
        aspect_ratio="16:9",
        style="modern, clean",
        task_id=1
    ),
    ImageTask(
        prompt="Mobile app icon",
        output_path="output/icon.png",
        resolution="4K",
        aspect_ratio="1:1",
        style="minimalist",
        task_id=2
    )
]

# Run batch
result = asyncio.run(run_batch_generation(
    tasks,
    concurrency=2,
    max_retries=5,
    verbose=True
))

# Save report
save_batch_report(result, "report.json")

---

## Configuration

Environment Variables:
```bash
# Required
GOOGLE_API_KEY=your_google_ai_api_key

# Optional (for .env file support)
pip install python-dotenv
```

Get your API key from: https://aistudio.google.com/apikey

Dependencies:
```bash
pip install google-genai python-dotenv pyyaml
```

---

## Works Well With

Complementary Skills:
- `moai-docs-generation` - Automated documentation with images
- `moai-domain-uiux` - UI/UX design asset generation
- `moai-domain-frontend` - Frontend visual components
- `moai-workflow-templates` - Template-based image workflows

Integration Points:
- Documentation illustration pipelines
- Design system asset generation
- Marketing material creation
- Automated visual content workflows

---

## Resources

Skill Files:
- `SKILL.md` - Main documentation (this file)
- `scripts/generate_image.py` - Single image generation CLI
- `scripts/batch_generate.py` - Batch generation CLI
- `examples.md` - Additional workflow examples

External Documentation:
- Google AI Studio: https://aistudio.google.com/
- Gemini API Reference: https://ai.google.dev/gemini-api/docs

---

Version: 1.1.0
Last Updated: 2025-12-23
Status: Active
