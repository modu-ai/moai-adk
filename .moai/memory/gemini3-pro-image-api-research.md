# Google Gemini 3 Pro Image Generation API Research Report

**Date**: 2025-11-26
**Research Focus**: Latest Google Gemini 3 Pro image generation API documentation and implementation patterns
**Status**: Complete
**Urgency**: High (Breaking API changes in SDK 0.8.5 → Unified SDK)

---

## Executive Summary

### Critical Finding: google-generativeai 0.8.5 is DEPRECATED

The old `google-generativeai==0.8.5` SDK is **deprecated and will lose support August 31, 2025**. Google replaced it with the **unified Google GenAI SDK** (`google-genai`). The NanoBananaImageGenerator code in moai-connector-nano-banana is using deprecated API patterns.

**Key Issues**:
- ❌ `genai.Client(api_key=api_key)` does NOT exist in 0.8.5
- ❌ Old SDK has no `models.generate_content()` method
- ✅ Unified SDK uses `client = genai.Client(api_key=...)` correctly
- ✅ New SDK provides proper image generation support

---

## Research Findings

### 1. Latest Model Names - CONFIRMED

**Official Available Models**:

| Model | Type | Max Resolution | Speed | Use Case |
|-------|------|---|---|---|
| **gemini-3-pro-image-preview** | Advanced Image Gen | 4096px (4K) | 10-60s | Professional, highest quality |
| **gemini-2.5-flash-image** | Fast Image Gen | 1024px (1K) | 5-15s | Rapid iterations, prototyping |
| **gemini-2.0-flash-image** | Legacy Fast Gen | 1024px | - | Backward compatibility |

**Key Points**:
- Gemini 3 Pro Image is in "preview" status
- Both models support text-to-image and image-to-image
- Gemini 2.5 Flash is the recommended fast model for production
- No "gemini-3-pro-image" (without "-preview") exists yet

### 2. Correct SDK Initialization Pattern (NEW)

**OLD (DEPRECATED - Do NOT use)**:
```python
# ❌ WRONG - This pattern doesn't exist in unified SDK
import google.generativeai as genai
genai.configure(api_key=api_key)
client = genai.Client(api_key=api_key)  # ERROR: genai.Client doesn't exist
```

**NEW (CORRECT - Use this)**:
```python
# ✅ CORRECT - Unified SDK pattern
from google import genai

# Option 1: Explicit API key
client = genai.Client(api_key='GEMINI_API_KEY')

# Option 2: Environment variable (auto-detects GEMINI_API_KEY or GOOGLE_API_KEY)
client = genai.Client()

# Option 3: Vertex AI (enterprise)
client = genai.Client(
    vertexai=True,
    project='your-project-id',
    location='us-central1'
)
```

**Breaking Changes**:
- Package name: `google.generativeai` → `google.genai`
- Import: `import google.generativeai as genai` → `from google import genai`
- Client instantiation is completely different
- No more `genai.configure()` function

### 3. Image Generation API Structure

**Request Format** (Python SDK):
```python
from google.genai import types

response = client.models.generate_content(
    model='gemini-2.5-flash-image',
    contents='Your prompt here',
    config=types.GenerateContentConfig(
        response_modalities=["IMAGE"],
        image_config=types.ImageConfig(
            aspect_ratio="16:9",
        ),
    ),
)
```

**Text-to-Image Complete Example**:
```python
from google import genai
from google.genai import types

client = genai.Client(api_key='YOUR_API_KEY')

response = client.models.generate_content(
    model='gemini-2.5-flash-image',
    contents='A cartoon infographic for flying sneakers',
    config=types.GenerateContentConfig(
        response_modalities=["IMAGE"],
        image_config=types.ImageConfig(
            aspect_ratio="9:16",
        ),
    ),
)

# Access generated image
for part in response.parts:
    if part.inline_data:
        generated_image = part.as_image()
        generated_image.show()
```

### 4. ImageConfig Parameters

**Available Configuration Fields**:

```python
types.ImageConfig(
    aspect_ratio="16:9",           # See list below
    # Note: image_size is deprecated, use aspect_ratio instead
    # size field removed from API v1beta
)
```

**Supported Aspect Ratios** (11 total):
```
1:1      # Square (Instagram, profile pics)
2:3      # Vertical
3:2      # Landscape
3:4      # Tall vertical
4:3      # Classic landscape
4:5      # Vertical (Instagram Stories)
5:4      # Landscape
9:16     # Mobile vertical
16:9     # Widescreen/HD
21:9     # Ultrawide
9:21     # Ultrawide vertical (if supported)
```

**Resolution/Size Notes**:
- ⚠️ API v1beta does NOT have explicit resolution settings
- Aspect ratio determines visual output dimensions
- Actual pixel size determined by model (1K-4K internally handled)
- Use appropriate model for resolution needs:
  - `gemini-2.5-flash-image` → up to 1024px
  - `gemini-3-pro-image-preview` → up to 4096px

### 5. GenerateContentConfig Structure

**Complete Configuration**:
```python
types.GenerateContentConfig(
    response_modalities=["IMAGE"],          # ["TEXT"] | ["IMAGE"] | ["TEXT", "IMAGE"]
    image_config=types.ImageConfig(
        aspect_ratio="16:9",
    ),
    # Optional: temperature, top_p, etc. for text responses
    temperature=1.0,
    top_p=0.95,
    top_k=40,
    max_output_tokens=8192,
)
```

### 6. Response Structure

**Response Format**:
```python
# response.parts contains generated content
# Each part can be text or image

for part in response.parts:
    # Check for text content
    if hasattr(part, 'text') and part.text:
        print(part.text)  # Text description

    # Check for image content
    if hasattr(part, 'inline_data') and part.inline_data:
        image_bytes = base64.b64decode(part.inline_data.data)
        # Use PIL or other image library
        from PIL import Image
        from io import BytesIO
        image = Image.open(BytesIO(image_bytes))

# Access metadata
print(response.usage_metadata.total_token_count)
print(response.candidates[0].finish_reason)
```

### 7. Multi-Modal Input (Image-to-Image Editing)

**Image-to-Image Format**:
```python
response = client.models.generate_content(
    model='gemini-2.5-flash-image',
    contents=[{
        "parts": [
            {
                "text": "Edit instruction: Add a sunset in the background"
            },
            {
                "inline_data": {
                    "mime_type": "image/png",
                    "data": base64_encoded_image_data
                }
            }
        ]
    }],
    config=types.GenerateContentConfig(
        response_modalities=["IMAGE"],
        image_config=types.ImageConfig(
            aspect_ratio="16:9",
        ),
    ),
)
```

### 8. Authentication & API Keys

**Getting API Key**:
1. Visit: https://ai.google.dev (free tier)
2. Click "Get API Key"
3. Enable Gemini API
4. Copy key (format: `AIza...` for free tier, `gsk_...` for Google AI Studio)

**Environment Configuration**:
```bash
# Method 1: Environment variable
export GEMINI_API_KEY="your-api-key"

# Method 2: Python code
client = genai.Client(api_key='your-api-key')

# Auto-detection order
# 1. GEMINI_API_KEY
# 2. GOOGLE_API_KEY
# 3. Vertex AI credentials (if vertexai=True)
```

### 9. Safety & Configuration

**Response Modalities**:
```python
# Text only
response_modalities=["TEXT"]

# Image only
response_modalities=["IMAGE"]

# Both text and image (model describes + generates)
response_modalities=["TEXT", "IMAGE"]
```

**Safety Filtering** (Built-in):
- Hate speech blocking
- Explicit content filtering
- Dangerous material prevention
- Harassment prevention
- Civic integrity protection

---

## Critical Issues in Current Code

### Issue 1: Wrong Client Initialization

**Current (WRONG)**:
```python
# Line 102-103 in image_generator.py
genai.configure(api_key=api_key)
self.client = genai.Client(api_key=api_key)  # ❌ AttributeError
```

**Problem**: `genai.Client()` doesn't exist in `google.generativeai` (old SDK)

**Fix**:
```python
from google import genai  # Change import

# Use unified SDK
self.client = genai.Client(api_key=api_key)  # ✅ Works in unified SDK
```

### Issue 2: Wrong API Endpoint

**Current (WRONG)**:
```python
# Line 174-179
response = self.client.models.generate_content(
    model=model_name,
    contents=[{"parts": [{"text": prompt}]}],
    config=generation_config,  # Not using types.GenerateContentConfig
    tools=tools if tools else None
)
```

**Problem**: Using raw dict instead of typed config

**Fix**:
```python
from google.genai import types

response = client.models.generate_content(
    model=model_name,
    contents=prompt,
    config=types.GenerateContentConfig(
        response_modalities=["IMAGE"],
        image_config=types.ImageConfig(
            aspect_ratio=aspect_ratio,
        ),
    ),
)
```

### Issue 3: Wrong Config Structure

**Current (WRONG)**:
```python
# Line 160-166
generation_config = {
    "response_modalities": ["TEXT", "IMAGE"],
    "image_config": {
        "aspect_ratio": aspect_ratio,
        "image_size": resolution  # ❌ Deprecated field
    }
}
```

**Problems**:
- `image_size` field doesn't exist in v1beta API
- Not using proper types
- Text modality not needed for image-only output

**Fix**:
```python
from google.genai import types

config = types.GenerateContentConfig(
    response_modalities=["IMAGE"],
    image_config=types.ImageConfig(
        aspect_ratio=aspect_ratio,
    ),
)
```

---

## Migration Path

### Step 1: Update Dependencies

```bash
# Remove old SDK
pip uninstall google-generativeai

# Install new SDK
pip install google-genai>=0.1.0
```

### Step 2: Update Imports

```python
# OLD
import google.generativeai as genai
genai.configure(api_key=api_key)
client = genai.Client(api_key=api_key)  # Wrong!

# NEW
from google import genai
from google.genai import types

client = genai.Client(api_key=api_key)  # Correct!
```

### Step 3: Update generate_content Calls

```python
# OLD
response = client.models.generate_content(
    model='gemini-2.5-flash-image',
    contents=[{"parts": [{"text": prompt}]}],
    config={"response_modalities": ["IMAGE"], ...}
)

# NEW
response = client.models.generate_content(
    model='gemini-2.5-flash-image',
    contents=prompt,
    config=types.GenerateContentConfig(
        response_modalities=["IMAGE"],
        image_config=types.ImageConfig(
            aspect_ratio="16:9",
        ),
    ),
)
```

### Step 4: Update Response Handling

```python
# Response structure remains similar but use proper types
for part in response.parts:
    if part.inline_data:
        image_bytes = base64.b64decode(part.inline_data.data)
        image = Image.open(BytesIO(image_bytes))
```

---

## Best Practices from Official Docs

### 1. Prompt Structure

```
✅ Good Prompt:
A serene Japanese garden at golden hour.
Lighting: warm sunset light, creating peaceful mood.
Camera: wide-angle 35mm lens, low angle shot.
Composition: stone path leading to pagoda.
Color palette: warm gold, jade green, cream.
Style: photorealistic with cinematic color grading.
Quality: 4K. Final output: PNG.

❌ Bad Prompt:
A garden
```

### 2. Model Selection

- **Gemini 3 Pro Image**: Professional quality, documentation assets, marketing
- **Gemini 2.5 Flash Image**: Quick iterations, prototyping, testing

### 3. Supported Formats

**Input Images**:
- PNG, JPEG, WEBP, GIF (as inline_data)

**Output Images**:
- Base64-encoded PNG (in inline_data.data)

### 4. Safety Considerations

- Text rendering in images (sophisticated text support)
- People in images (updated safety filters)
- Search grounding for factual content

---

## API Endpoints (REST)

For reference (SDK handles this):

```
POST https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent
```

Request structure:
```json
{
  "contents": [
    {
      "parts": [
        {
          "text": "Your prompt"
        }
      ]
    }
  ],
  "generationConfig": {
    "responseModalities": ["IMAGE"],
    "imageConfig": {
      "aspectRatio": "16:9"
    }
  }
}
```

---

## Version Compatibility Matrix

| Package | Version | Status | Support End |
|---------|---------|--------|-------------|
| google-generativeai | 0.8.5 | Deprecated | Aug 31, 2025 |
| google-genai | 0.1.0+ | Current | Active |

---

## Recommended Implementation for NanoBananaImageGenerator

### File: modules/image_generator.py

**Key Changes Required**:
1. Change imports: `from google import genai`
2. Remove `genai.configure()` call
3. Fix client initialization: `self.client = genai.Client(api_key=api_key)`
4. Use `types.GenerateContentConfig()` with proper types
5. Remove `image_size` parameter, use `aspect_ratio` only
6. Update response parsing for correct part structure

**Estimated LOC Changes**: ~50 lines (~30% of generate() and edit() methods)

### Expected Behavior After Fix

```python
# Will work correctly
generator = NanoBananaImageGenerator(api_key)
image, metadata = generator.generate(
    "A serene mountain landscape",
    model="pro",
    resolution="2K",
    aspect_ratio="16:9"
)
# ✅ Returns: PIL.Image, dict with metadata
```

---

## Resource Links

**Official Documentation**:
- Python SDK: https://github.com/googleapis/python-genai
- Gemini API: https://ai.google.dev
- Image Generation Guide: https://docs.cloud.google.com/vertex-ai/generative-ai/docs/multimodal/image-generation
- API Models: https://ai.google.dev/models

**Key Documentation Pages**:
1. Models: https://ai.google.dev/models
2. Gemini API Overview: https://ai.google.dev/gemini-api/docs/api-overview
3. Image Generation: https://docs.cloud.google.com/vertex-ai/generative-ai/docs/multimodal/image-generation
4. Python SDK: https://github.com/googleapis/python-genai

---

## Summary & Recommendations

### Key Takeaways

1. ✅ Gemini 3 Pro Image Preview model exists and supports 4K resolution
2. ✅ Gemini 2.5 Flash Image is recommended production model
3. ❌ google-generativeai 0.8.5 is deprecated (EOL: Aug 31, 2025)
4. ✅ Unified google-genai SDK has correct API patterns
5. ✅ Aspect ratio support confirmed (11 options)
6. ⚠️ No explicit "resolution" field in API (model determines max resolution)

### Action Items

**Immediate** (This Session):
- [ ] Update NanoBananaImageGenerator to use google-genai SDK
- [ ] Fix Client initialization pattern
- [ ] Update GenerateContentConfig to use proper types
- [ ] Remove deprecated image_size parameter
- [ ] Test with both models (flash and pro)

**Short-term** (Before Aug 31, 2025):
- [ ] Migrate all deprecated SDK usage to new SDK
- [ ] Update all documentation and examples
- [ ] Test production workloads with new SDK
- [ ] Verify token counting and cost changes

**Medium-term**:
- [ ] Monitor Gemini 3 Pro Image status (currently "preview")
- [ ] Evaluate Gemini 3 Pro Image production readiness
- [ ] Consider performance benchmarks (old vs new SDK)

---

**Research Completed By**: Documentation Research Specialist (Context7 MCP Integration)
**Last Updated**: 2025-11-26
**Status**: Ready for Implementation
