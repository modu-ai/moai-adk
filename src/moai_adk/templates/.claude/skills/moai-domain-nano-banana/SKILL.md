---
name: moai-domain-nano-banana
description: Moai Domain Nano Banana - Professional implementation guide
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: moai, nano, domain, banana  


# 🍌 moai-domain-nano-banana Skill

**Google Nano Banana Pro 이미지 생성 및 편집 전문가**

Enterprise-grade image generation with Nano Banana Pro (Gemini 3 Pro Image Preview), featuring professional asset production, advanced text rendering, real-world grounding, and multi-modal composition.

---

## 🎯 Core Purpose

Provide comprehensive support for professional image generation and editing using Google's Nano Banana Pro model, enabling:

- **Text-to-Image Generation**: Convert detailed prompts to 1K/2K/4K resolution images
- **Image-to-Image Editing**: Style transfer, object manipulation, element addition/removal
- **Real-time Grounding**: Google Search integration for factual, up-to-date content
- **Advanced Composition**: Multi-image reference support, sophisticated text rendering
- **Enterprise Integration**: Google Cloud, Vertex AI, production-ready error handling

---

## 📋 Skill Capabilities

### 1. Image Generation Models

```typescript
// Nano Banana Pro (Recommended for professional work)
Model: "gemini-3-pro-image-preview"
Features:
  - 1K, 2K, 4K resolution support
  - Built-in "Thinking" process for auto-optimization
  - Google Search real-time grounding
  - Up to 14 reference images (6 objects + 5 humans)
  - Sophisticated text rendering in images
  - Aspect ratios: 1:1, 16:9, 21:9, 2:3, 3:2, 3:4, 4:3, 4:5, 5:4, 9:16
  - Processing time: 10-60 seconds (quality priority)
  - Use case: Commercial design, professional content, high-quality assets

// Gemini 2.5 Flash Image (Speed-optimized alternative)
Model: "gemini-2.5-flash-image"
Features:
  - 1024px fixed resolution
  - Low-latency generation
  - Basic text rendering
  - Processing time: 5-15 seconds
  - Use case: Quick iterations, prototypes, high-volume generation
```

### 2. Generation Capabilities

#### Text-to-Image
```python
# Basic text prompt to image
generate_image(
    prompt="A photorealistic portrait of...",
    resolution="2K",          # 1K, 2K, or 4K
    aspect_ratio="16:9",      # Various ratios supported
    enable_google_search=True, # Real-time information
    thinking_process=True,    # Auto-optimize composition
    seed=42                   # Reproducibility (optional)
)
```

**Returns**: Base64-encoded PNG image + metadata + SynthID watermark

#### Image-to-Image Editing
```python
# Style transfer, object manipulation, element adjustment
edit_image(
    original_image_path="./original.png",
    instruction="Transform into Van Gogh's Starry Night style...",
    preserve_composition=True,
    resolution="2K",
    reference_images=[]  # Optional: up to 14 reference images
)
```

**Returns**: Edited image + composition analysis

#### Multi-turn Refinement
```python
# Interactive conversation for iterative improvement
refine_image(
    conversation=[
        {"role": "user", "content": "Initial prompt or image"},
        {"role": "assistant", "content": "Generated image"},
        {"role": "user", "content": "Make the sky more dramatic..."}
    ],
    max_turns=5
)
```

### 3. Prompt Engineering Support

#### Structured Prompt Template
```
[Scene Description]
A [adjective] [subject] doing [action].
The setting is [location] with [environmental details].

[Photographic Elements]
Lighting: [lighting_type], creating [mood].
Camera: [angle] shot with [lens] lens (mm).
Composition: [framing_details].

[Color & Style]
Color palette: [colors]. Style: [art_style].
Mood: [emotional_tone].

[Technical Specs]
Quality: studio-grade, high-resolution, professional photography.
Format: [orientation/ratio].
```

**Bad**: "dog eating banana"
**Good**: "A golden retriever with soft, shiny fur enjoying a peeled banana on a sunny beach. Golden hour lighting illuminates the scene. Shot with 85mm portrait lens, shallow depth of field (bokeh). Warm, joyful mood. Professional photography quality, 16:9 aspect ratio."

#### Photographic Elements Guide

| Element | Examples |
|---------|----------|
| **Lighting** | Golden hour, harsh shadows, soft diffuse, neon glow, candlelight |
| **Camera Angle** | Wide shot, close-up, overhead, low angle, Dutch angle |
| **Lens** | 35mm, 50mm, 85mm (portrait), 24mm (landscape), macro |
| **Depth** | Shallow DoF (f/1.8), Deep DoF (f/16), bokeh, tack-sharp |
| **Mood** | Serene, dramatic, chaotic, intimate, majestic, eerie |
| **Color** | Warm tones, cool palette, high-contrast, muted, vibrant |

### 4. Google Search Grounding

```python
# Enable real-time information integration
client.set_google_search_grounding(enabled=True)

# This allows generation based on:
# - Recent events and news
# - Current trends and fashion
# - Latest technology products
# - Real-world locations and conditions
# - Factual information from web
```

**Note**: Image-based search results are excluded for privacy/copyright

### 5. Quality Configuration

```python
class GenerationConfig:
    resolution: Literal["1K", "2K", "4K"] = "2K"
    aspect_ratio: str = "16:9"
    enable_thinking: bool = True        # Compose optimization
    enable_grounding: bool = True       # Google Search integration
    add_watermark: bool = True          # SynthID watermark
    seed: int = None                    # For reproducibility
    max_retries: int = 3
    timeout_seconds: int = 60
```

---

## 🛠️ Implementation Guide

### Installation

```bash
# Python SDK
pip install google-generativeai

# JavaScript SDK
npm install @google/genai

# Go SDK
go get github.com/googleapis/go-genai
```

### Python Implementation

```python
from google import genai
from pathlib import Path
import base64
from datetime import datetime

class NanoBananaPro:
    def __init__(self, api_key: str, model: str = "gemini-3-pro-image-preview"):
        self.client = genai.Client(api_key=api_key)
        self.model = model
        self.generation_history = []

    def generate_image(
        self,
        prompt: str,
        resolution: str = "2K",
        aspect_ratio: str = "16:9",
        enable_google_search: bool = True,
        enable_thinking: bool = True,
        save_path: str = None
    ) -> dict:
        """
        Generate image from text prompt

        Args:
            prompt: Detailed narrative prompt (not keywords)
            resolution: "1K", "2K", or "4K"
            aspect_ratio: "1:1", "16:9", "21:9", "2:3", "3:2", "3:4", "4:3", "4:5", "5:4", "9:16"
            enable_google_search: Use real-time web information
            enable_thinking: Auto-optimize composition
            save_path: Optional file path to save image

        Returns:
            {
                "image_data": base64_string,
                "mime_type": "image/png",
                "resolution": "2K",
                "has_watermark": True,
                "timestamp": "2025-11-22T...",
                "prompt_used": prompt,
                "metadata": {...}
            }
        """
        try:
            # Validate resolution
            valid_resolutions = ["1K", "2K", "4K"]
            if resolution not in valid_resolutions:
                raise ValueError(f"Resolution must be one of {valid_resolutions}")

            # Build request
            generation_config = {
                "imageConfig": {
                    "aspectRatio": aspect_ratio,
                    "imageSize": resolution
                }
            }

            if enable_thinking:
                generation_config["thinking"] = {"budgetTokens": 3000}

            if enable_google_search:
                # Add tools for grounding
                generation_config["tools"] = [{"google_search": {}}]

            # Generate
            response = self.client.models.generate_content(
                model=self.model,
                contents=[{"parts": [{"text": prompt}]}],
                config=generation_config
            )

            # Extract image data
            image_data = None
            for part in response.candidates[0].content.parts:
                if part.inline_data:
                    image_data = part.inline_data.data
                    mime_type = part.inline_data.mime_type
                    break

            if not image_data:
                raise RuntimeError("No image data in response")

            result = {
                "image_data": image_data,
                "mime_type": mime_type,
                "resolution": resolution,
                "has_watermark": True,  # SynthID always included
                "timestamp": datetime.now().isoformat(),
                "prompt_used": prompt,
                "finish_reason": response.candidates[0].finish_reason
            }

            # Save if requested
            if save_path:
                self._save_image(image_data, save_path)
                result["saved_to"] = save_path

            # Store in history
            self.generation_history.append(result)

            return result

        except Exception as e:
            raise RuntimeError(f"Image generation failed: {str(e)}")

    def edit_image(
        self,
        image_path: str,
        instruction: str,
        resolution: str = "2K",
        preserve_composition: bool = True,
        save_path: str = None
    ) -> dict:
        """
        Edit existing image with text instruction

        Args:
            image_path: Path to input image
            instruction: Detailed editing instruction
            resolution: Output resolution
            preserve_composition: Maintain original structure
            save_path: Optional file path to save edited image

        Returns:
            {
                "original_image_path": str,
                "edited_image_data": base64_string,
                "instruction": str,
                "timestamp": str,
                ...
            }
        """
        try:
            # Read and encode image
            with open(image_path, "rb") as f:
                image_bytes = f.read()

            image_base64 = base64.standard_b64encode(image_bytes).decode("utf-8")

            # Determine MIME type
            mime_type = self._get_mime_type(image_path)

            # Build instruction with composition guidance
            full_instruction = instruction
            if preserve_composition:
                full_instruction += "\n\nPreserve the original composition and layout."

            # Generate edited image
            response = self.client.models.generate_content(
                model=self.model,
                contents=[{
                    "parts": [
                        {
                            "inline_data": {
                                "mime_type": mime_type,
                                "data": image_base64
                            }
                        },
                        {"text": full_instruction}
                    ]
                }],
                config={"imageConfig": {"imageSize": resolution}}
            )

            # Extract result
            edited_image_data = None
            for part in response.candidates[0].content.parts:
                if part.inline_data:
                    edited_image_data = part.inline_data.data
                    break

            if not edited_image_data:
                raise RuntimeError("No edited image in response")

            result = {
                "original_image_path": str(image_path),
                "edited_image_data": edited_image_data,
                "mime_type": mime_type,
                "instruction": instruction,
                "timestamp": datetime.now().isoformat(),
                "resolution": resolution
            }

            if save_path:
                self._save_image(edited_image_data, save_path)
                result["saved_to"] = save_path

            return result

        except Exception as e:
            raise RuntimeError(f"Image editing failed: {str(e)}")

    def _save_image(self, image_data: str, save_path: str) -> None:
        """Save base64 image data to file"""
        Path(save_path).parent.mkdir(parents=True, exist_ok=True)
        with open(save_path, "wb") as f:
            f.write(base64.b64decode(image_data))

    def _get_mime_type(self, file_path: str) -> str:
        """Determine MIME type from file extension"""
        ext = Path(file_path).suffix.lower()
        mime_types = {
            ".png": "image/png",
            ".jpg": "image/jpeg",
            ".jpeg": "image/jpeg",
            ".gif": "image/gif",
            ".webp": "image/webp"
        }
        return mime_types.get(ext, "image/png")
```

### JavaScript Implementation

```typescript
import { GoogleGenAI } from "@google/genai";
import * as fs from "fs";
import * as path from "path";

interface GenerationResult {
  imageData: string;
  mimeType: string;
  resolution: string;
  timestamp: string;
  promptUsed: string;
}

class NanoBananaPro {
  private client: GoogleGenAI;
  private model: string = "gemini-3-pro-image-preview";
  private generationHistory: GenerationResult[] = [];

  constructor(apiKey: string) {
    this.client = new GoogleGenAI({ apiKey });
  }

  async generateImage(
    prompt: string,
    resolution: "1K" | "2K" | "4K" = "2K",
    aspectRatio: string = "16:9",
    options: {
      enableGoogleSearch?: boolean;
      enableThinking?: boolean;
      savePath?: string;
    } = {}
  ): Promise<GenerationResult> {
    try {
      const {
        enableGoogleSearch = true,
        enableThinking = true,
        savePath
      } = options;

      // Build generation config
      const generationConfig: any = {
        imageConfig: {
          aspectRatio,
          imageSize: resolution
        }
      };

      if (enableThinking) {
        generationConfig.thinking = { budgetTokens: 3000 };
      }

      if (enableGoogleSearch) {
        generationConfig.tools = [{ google_search: {} }];
      }

      // Generate image
      const response = await this.client.models.generateContent({
        model: this.model,
        contents: [
          {
            parts: [{ text: prompt }]
          }
        ],
        config: generationConfig
      });

      // Extract image
      let imageData: string | null = null;
      let mimeType = "image/png";

      for (const part of response.candidates[0].content.parts) {
        if (part.inlineData) {
          imageData = part.inlineData.data;
          mimeType = part.inlineData.mimeType;
          break;
        }
      }

      if (!imageData) {
        throw new Error("No image data in response");
      }

      const result: GenerationResult = {
        imageData,
        mimeType,
        resolution,
        timestamp: new Date().toISOString(),
        promptUsed: prompt
      };

      // Save if requested
      if (savePath) {
        this.saveImage(imageData, savePath);
        (result as any).savedTo = savePath;
      }

      this.generationHistory.push(result);
      return result;
    } catch (error) {
      throw new Error(`Image generation failed: ${error}`);
    }
  }

  private saveImage(imageData: string, savePath: string): void {
    const dir = path.dirname(savePath);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }
    const buffer = Buffer.from(imageData, "base64");
    fs.writeFileSync(savePath, buffer);
  }
}
```

---

## 📊 Performance & Optimization

### Resolution Selection Guide

| Resolution | Use Case | Processing Time | Token Cost | Output Quality |
|-----------|----------|-----------------|-----------|-----------------|
| **1K** | Web thumbnails, quick previews | 10-20s | ~1-2K | Good |
| **2K** | Web images, social media | 20-35s | ~2-4K | Excellent |
| **4K** | Print materials, high-detail, posters | 40-60s | ~4-8K | Studio-grade |

### Cost Optimization Strategies

1. **Batch similar requests** together to maximize throughput
2. **Use 1K for iterations**, upgrade to 2K/4K for finals
3. **Enable caching** for frequently used prompts
4. **Reuse reference images** across multiple generations

### Error Handling & Retries

```python
# Quota exceeded: suggest resolution downgrade
if "RESOURCE_EXHAUSTED" in error:
    logger.info("Quota exceeded. Try: reduce resolution to 1K")

# Safety filter triggered: restructure prompt
if "SAFETY_RATING" in error:
    logger.info("Safety filter triggered. Avoid: violence, explicit content")

# Timeout: simplify prompt
if "DEADLINE_EXCEEDED" in error:
    logger.info("Timeout. Try: simpler, more focused prompt")
```

---

## 🔐 Security Best Practices

1. **API Key Management**
   ```bash
   export GOOGLE_API_KEY="your-api-key"  # Never commit keys
   ```

2. **Input Validation**
   ```python
   # Verify image ownership before processing
   # Sanitize user prompts for malicious content
   ```

3. **Output Handling**
   ```python
   # All outputs include SynthID watermark
   # Store metadata for audit trails
   ```

4. **Google Cloud Integration**
   ```python
   # Use Vertex AI for production deployments
   # Enable Cloud Audit Logs for compliance
   ```

---

## 📚 Usage Examples

### Example 1: Professional Marketing Asset

```python
from nano_banana import NanoBananaPro

client = NanoBananaPro(api_key="YOUR_API_KEY")

# Generate high-quality product image
result = client.generate_image(
    prompt="""
    A sleek, modern smartphone displaying a vibrant app interface,
    positioned at a 45-degree angle on a minimalist white desk.
    Soft, diffused window light illuminates the scene from the left.
    Sharp focus on the device screen, blurred background with subtle shadows.
    Shot with 50mm lens on a full-frame camera. Clean, professional aesthetic.
    4K studio photography quality. Aspect ratio 16:9.
    """,
    resolution="4K",
    aspect_ratio="16:9",
    enable_google_search=True
)

# Save result
with open("product-hero.png", "wb") as f:
    f.write(base64.b64decode(result["image_data"]))
```

### Example 2: Style Transfer

```python
# Read original image
original_path = "./city-street.jpg"

# Apply artistic style
edited = client.edit_image(
    image_path=original_path,
    instruction="""
    Transform this photograph into the artistic style of
    Vincent van Gogh's 'Starry Night'. Use swirling,
    impasto brushstrokes, deep blues and bright yellows,
    while preserving the original composition.
    """,
    resolution="2K",
    save_path="./city-starry-night.png"
)
```

### Example 3: Multi-turn Refinement

```python
# Start with initial generation
results = []

# Turn 1: Initial concept
img1 = client.generate_image(
    prompt="A serene Japanese garden with cherry blossoms"
)
results.append(img1)

# Turn 2: Refine with editing instruction
img2 = client.edit_image(
    image_path="./temp-garden-v1.png",
    instruction="Add morning mist and golden sunlight through the trees"
)
results.append(img2)

# Turn 3: Final polish
img3 = client.edit_image(
    image_path="./temp-garden-v2.png",
    instruction="Enhance colors, increase saturation slightly, add subtle bokeh effect"
)
results.append(img3)

# Save final
final_path = f"./garden-final-{img3['timestamp']}.png"
```

---

## 🎓 Prompt Engineering Masterclass

### Anatomy of a Great Prompt

```
SCENE LAYER (Foundation)
↓
A [emotional adjective] [subject] [action].
The setting is [specific location] with [environmental details].

PHOTOGRAPHIC LAYER (Technique)
↓
Lighting: [light type] from [direction], creating [mood].
Camera: [camera type/angle], [lens details], [depth of field].
Composition: [framing], [perspective], [balance].

COLOR & STYLE LAYER (Aesthetic)
↓
Color palette: [specific colors].
Art style: [reference or technique].
Mood/Atmosphere: [emotional quality].

QUALITY LAYER (Standards)
↓
Quality: [professional standard].
Aspect ratio: [ratio].
SynthID watermark: [included by default].
```

### Common Pitfalls & Solutions

| ❌ Pitfall | ✅ Solution |
|-----------|-----------|
| "Cat picture" | "A fluffy orange tabby cat with bright green eyes, sitting on a sunlit windowsill, looking out at a snowy winter landscape" |
| "Nice landscape" | "A dramatic mountain vista at golden hour, with snow-capped peaks reflecting in a pristine alpine lake, stormy clouds parting above" |
| Multiple objects (list) | "A cozy bookshelf scene: worn leather armchair, stack of vintage books, reading lamp with warm glow, fireplace in background" |
| Vague style ("beautiful") | "Shot with 85mm portrait lens, shallow depth of field (f/2.8), film photography aesthetic, warm color grading, 1970s nostalgic feel" |

---

## 🚀 Advanced Topics

### Google Cloud Integration

```python
from google.cloud import aiplatform

# Use Vertex AI for production
aiplatform.init(project="your-project-id")

# Higher availability, SLA guarantees, audit logs
# Compatible with Nano Banana Pro models
```

### Batch Processing

```python
# Generate multiple images efficiently
prompts = [
    "A serene forest at dawn",
    "A bustling market at sunset",
    "An industrial cityscape at night"
]

results = []
for prompt in prompts:
    result = client.generate_image(prompt, resolution="2K")
    results.append(result)
    # Respectful rate limiting
    time.sleep(2)
```

### Custom Thought Process Budgets

```python
# Configure internal reasoning depth
generation_config = {
    "thinking": {
        "budgetTokens": 5000  # 1K-10K range
        # More tokens = deeper composition analysis
        # 1K = quick optimization
        # 10K = deep, nuanced refinement
    }
}
```

---

## 📋 Testing & Validation

### Unit Tests

```python
def test_generate_basic_image():
    client = NanoBananaPro(api_key="test_key")
    result = client.generate_image("A red apple on white background")
    assert result["resolution"] == "2K"
    assert result["mime_type"] == "image/png"
    assert len(result["image_data"]) > 1000

def test_invalid_resolution():
    client = NanoBananaPro(api_key="test_key")
    with pytest.raises(ValueError):
        client.generate_image("test", resolution="8K")
```

### Integration Tests

```python
def test_end_to_end_generation_and_edit():
    client = NanoBananaPro(api_key=os.getenv("GOOGLE_API_KEY"))

    # Generate
    img = client.generate_image("Professional office space")
    assert img["saved_to"]

    # Edit
    edited = client.edit_image(
        img["saved_to"],
        "Add plants and natural lighting"
    )
    assert edited["original_image_path"] == img["saved_to"]
```

---

## 🔗 Related Resources

- **Google Nano Banana Pro Official**: https://blog.google/intl/ko-kr/company-news/technology/nano-banana-pro/
- **Gemini API Docs**: https://ai.google.dev/gemini-api/docs/image-generation
- **Context7 Library**: `/websites/ai_google_dev_gemini-api`
- **Go SDK**: `googleapis/go-genai`
- **JavaScript SDK**: `/googleapis/js-genai`

---

## 📞 Support & Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| "API key not found" | Set `GOOGLE_API_KEY` environment variable |
| "Quota exceeded" | Reduce resolution to 1K or wait for quota reset |
| "Safety rating triggered" | Restructure prompt, avoid explicit content |
| "Timeout error" | Simplify prompt, reduce resolution, increase timeout |
| "Invalid aspect ratio" | Use: 1:1, 16:9, 21:9, 2:3, 3:2, 3:4, 4:3, 4:5, 5:4, 9:16 |

### Debug Mode

```python
import logging

logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger("nano_banana")

# Full request/response logging enabled
client = NanoBananaPro(api_key="YOUR_KEY")
```

---

## 📊 Benchmarks (Expected Performance)

**Hardware**: Standard API endpoint
**Test Prompt Complexity**: Medium (detailed scene description)

| Metric | 1K | 2K | 4K |
|--------|----|----|-----|
| Avg Generation Time | 12s | 25s | 45s |
| Tokens/Request | ~1,500 | ~3,000 | ~6,000 |
| Quality Score (1-10) | 8 | 9.5 | 10 |
| Watermark Applied | ✓ | ✓ | ✓ |

---

**Skill Version**: 1.0.0
**Last Updated**: 2025-11-22
**Compatibility**: Python 3.8+, JavaScript ES2020+, Go 1.18+
